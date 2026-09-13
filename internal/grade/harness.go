package grade

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/coding-agents-course/internal/fakeanthropic"
)

// Timeouts. Generous enough that a correct program on a slow machine never
// trips them, short enough that a wedged submission fails in under a minute.
const (
	RoundTimeout = 20 * time.Second
	ExitTimeout  = 15 * time.Second
)

// graderLine is what the grader writes to the program's stdin.
type graderLine struct {
	User string `json:"user"`
}

// programLine is every shape the program is allowed to write to stdout. A
// legal line sets exactly one of the two fields.
type programLine struct {
	Assistant *string `json:"assistant,omitempty"`
	Usage     *struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"usage,omitempty"`
}

// RunResult is the raw evidence collected from one graded run. Nothing here is
// a judgement; checks.go turns this into verdicts.
type RunResult struct {
	Answers   []string // program's assistant line per round, in order
	GotRounds int      // how many rounds produced a well-formed answer

	UsageSeen   bool
	UsageInput  int
	UsageOutput int

	Stderr    string
	ExitCode  int
	ExitError string

	Protocol []string // protocol violations, in the order discovered
	Extra    []string // unexpected stdout lines

	Records   []fakeanthropic.Record
	FakeUsage fakeanthropic.Usage
}

// Build compiles the submission if given a directory, and returns a path to an
// executable. A path to an existing executable file is returned unchanged.
func Build(path string) (bin string, cleanup func(), err error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", nil, err
	}
	if !info.IsDir() {
		abs, err := filepath.Abs(path)
		return abs, func() {}, err
	}
	tmp, err := os.MkdirTemp("", "course-grade-")
	if err != nil {
		return "", nil, err
	}
	bin = filepath.Join(tmp, "submission")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmp)
		return "", nil, fmt.Errorf("go build failed:\n%s", out)
	}
	return bin, func() { os.RemoveAll(tmp) }, nil
}

// Run starts the fake server, executes the submission against it, and returns
// the evidence. An error is returned only for grader-side failures; a
// misbehaving submission produces a RunResult with violations recorded.
func Run(bin string) (*RunResult, error) {
	fake := &fakeanthropic.Server{Script: Replies()}
	baseURL, err := fake.Start()
	if err != nil {
		return nil, fmt.Errorf("starting fake server: %w", err)
	}
	defer fake.Close()

	res := &RunResult{}

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_BASE_URL="+baseURL,
		"ANTHROPIC_API_KEY="+fakeanthropic.ExpectedAPIKey,
		"ANTHROPIC_MODEL="+fakeanthropic.ExpectedModel,
		// Belt and braces: a student who wired the base URL under a different
		// name still lands on the fake rather than on the real API.
		"ANTHROPIC_API_URL="+baseURL,
		// Same, for the model. Chapter 2's submission prefers LLM_MODEL over
		// ANTHROPIC_MODEL, and this harness also runs as Chapter 2's phase 1.
		// Pinning both names keeps ExpectedModel the single source of truth
		// and stops an ambient LLM_MODEL in the grader's own environment from
		// failing a correct submission.
		"LLM_MODEL="+fakeanthropic.ExpectedModel,
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting submission: %w", err)
	}

	lines := make(chan string, 64)
	readErr := make(chan error, 1)
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
		for sc.Scan() {
			lines <- sc.Text()
		}
		readErr <- sc.Err()
		close(lines)
	}()

	// nextLine returns the next non-blank stdout line, or an error on timeout
	// or clean EOF.
	nextLine := func() (string, error) {
		for {
			select {
			case l, ok := <-lines:
				if !ok {
					return "", io.EOF
				}
				if strings.TrimSpace(l) == "" {
					continue
				}
				return l, nil
			case <-time.After(RoundTimeout):
				return "", errors.New("timeout")
			}
		}
	}

	for i, round := range Script {
		req, _ := json.Marshal(graderLine{User: round.User})
		if _, err := stdin.Write(append(req, '\n')); err != nil {
			res.Protocol = append(res.Protocol,
				fmt.Sprintf("round %d: could not write to the program's stdin (%v) — it exited early?", i+1, err))
			break
		}

		// Read until this round's protocol line arrives. Junk on stdout is
		// recorded as a violation but does not desynchronize the stream: one
		// defect should produce one precise diagnosis, not a cascade.
		var answered bool
		for !answered {
			line, err := nextLine()
			if err != nil {
				res.Protocol = append(res.Protocol,
					fmt.Sprintf("round %d: no reply on stdout (%v)", i+1, err))
				break
			}

			var pl programLine
			if err := json.Unmarshal([]byte(line), &pl); err != nil {
				res.Protocol = append(res.Protocol,
					fmt.Sprintf("round %d: stdout line is not JSON: %q", i+1, truncate(line, 200)))
				res.Extra = append(res.Extra, line)
				continue
			}
			if pl.Assistant == nil {
				res.Protocol = append(res.Protocol,
					fmt.Sprintf("round %d: expected {\"assistant\": ...}, got %q", i+1, truncate(line, 200)))
				res.Extra = append(res.Extra, line)
				continue
			}
			res.Answers = append(res.Answers, *pl.Assistant)
			res.GotRounds++
			answered = true
		}
		if !answered {
			break
		}
	}

	// EOF on stdin is the signal to report usage and exit.
	_ = stdin.Close()

	for {
		line, err := nextLine()
		if err != nil {
			break
		}
		var pl programLine
		if err := json.Unmarshal([]byte(line), &pl); err != nil || pl.Usage == nil {
			res.Extra = append(res.Extra, line)
			continue
		}
		res.UsageSeen = true
		res.UsageInput = pl.Usage.Input
		res.UsageOutput = pl.Usage.Output
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				res.ExitCode = ee.ExitCode()
			} else {
				res.ExitCode = -1
			}
			res.ExitError = err.Error()
		}
	case <-time.After(ExitTimeout):
		_ = cmd.Process.Kill()
		res.ExitCode = -1
		res.ExitError = "program did not exit after stdin was closed"
	}
	<-readErr

	res.Stderr = stderr.String()
	res.Records = fake.Records()
	res.FakeUsage = fake.TotalUsage()
	return res, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
