package grade

// Chapter 7 harness — build the exercise, drive it, and measure the SHAPE of
// the stream it reports.
//
// The exercise is run TWICE against the same script: once normally, and once
// with CH07_NO_STREAM=1. The second run is what makes "streaming is not a
// mode" checkable rather than assertable. Both runs must produce the same
// final text and the same parts; only the chunk count may differ.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/coding-agents-course/internal/fakevendor"
)

// Ch7Delta is one observed PartDelta.
type Ch7Delta struct {
	PartID uint64
	Kind   string
	Chunk  string
}

// Ch7PartFinal is one observed part_final, capturing the kind for ordering.
type Ch7PartFinal struct {
	PartID uint64
	Kind   string // "text", "tool_call", or "thinking"
}

// Ch7Exec is one execution of the exercise binary.
type Ch7Exec struct {
	Ran       bool
	ExitErr   string
	Stdout    string
	Stderr    string
	Lines     []map[string]any
	Deltas    []Ch7Delta
	FinalText map[uint64]string // part_id -> finalized text, text parts only
	FinalTool map[uint64]string // part_id -> finalized tool name
	Finals    []Ch7PartFinal    // ordered part finals, all kinds
	Assistant string
	Stats     []map[string]any
}

// KindCount counts deltas of one kind.
func (r Ch7Exec) KindCount(kind string) int {
	n := 0
	for _, d := range r.Deltas {
		if d.Kind == kind {
			n++
		}
	}
	return n
}

// TextByPart concatenates text deltas per part, in arrival order.
func (r Ch7Exec) TextByPart() map[uint64]string {
	out := map[uint64]string{}
	for _, d := range r.Deltas {
		if d.Kind == "text" {
			out[d.PartID] += d.Chunk
		}
	}
	return out
}

type Ch7Result struct {
	Base string

	ExBuildOK  bool
	ExBuildErr string

	Stream Ch7Exec
	Plain  Ch7Exec

	Ch6Result *Ch6Result
	Ch6Err    string

	HelpersErr string
}

// ch7Replies is the script both runs see.
//
// One response carrying reasoning, reply text, AND a tool call, so a single
// turn exercises all three delta kinds. The text is deliberately longer than
// a few characters: at three runes per chunk a short reply cannot distinguish
// a real incremental parser from one that emits the whole thing at once.
func ch7Replies() []fakevendor.Reply {
	return []fakevendor.Reply{
		{
			Thinking: "The caller wants the configuration file. I should call the tool before answering.",
			Text:     "Let me take a look at that configuration for you.",
			ToolName: "think",
			ToolArgs: `{"seconds":0,"thought":"inspecting the configuration"}`,
			ToolID:   "call_ch7_1",
			Usage:    fakevendor.Canonical{Input: 120, Output: 64},
		},
		{
			Text:  "The configuration sets the listen address to port 8080.",
			Usage: fakevendor.Canonical{Input: 190, Output: 24},
		},
	}
}

// Ch7Run builds and drives the ch7 exercise.
func Ch7Run(path string) (*Ch7Result, error) {
	base, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	r := &Ch7Result{Base: base}

	tmp, err := os.MkdirTemp("", "ch7grade")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	// Build the exercise. A separate module with its own replace directive,
	// so this also proves the public API is reachable from outside.
	exDir := filepath.Join(base, "ch07")
	if _, err := os.Stat(exDir); err != nil {
		r.ExBuildErr = "ch07/ directory not found"
	} else {
		bin := filepath.Join(tmp, "ch07bin")
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Dir = exDir
		if out, err := cmd.CombinedOutput(); err != nil {
			r.ExBuildErr = strings.TrimSpace(string(out))
		} else {
			r.ExBuildOK = true
			r.Stream = driveCh7(bin, tmp, false)
			r.Plain = driveCh7(bin, tmp, true)
		}
	}

	// ch6 parity, against the same submission.
	ch6, err := Ch6Run(base)
	if err != nil {
		r.Ch6Err = err.Error()
	} else {
		r.Ch6Result = ch6
	}

	return r, nil
}

// driveCh7 runs the exercise once and parses its observation stream.
func driveCh7(bin, tmp string, noStream bool) Ch7Exec {
	run := Ch7Exec{
		FinalText: map[uint64]string{},
		FinalTool: map[uint64]string{},
	}

	srv := fakevendor.New(ch7Replies())
	defer srv.Close()

	name := "stream"
	if noStream {
		name = "plain"
	}
	logPath := filepath.Join(tmp, name+".log")

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=test-key",
		"CH07_LOG="+logPath,
		"CH07_AGENT_LOG="+filepath.Join(tmp, name+"-agent.log"),
	)
	if noStream {
		cmd.Env = append(cmd.Env, "CH07_NO_STREAM=1")
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		run.ExitErr = err.Error()
		return run
	}
	if err := cmd.Start(); err != nil {
		run.ExitErr = err.Error()
		return run
	}

	fmt.Fprintln(stdin, `{"kind":"prompt","text":"what does the configuration say?"}`)
	stdin.Close()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			run.ExitErr = err.Error()
		}
	case <-time.After(30 * time.Second):
		// A generous bound that still distinguishes a hang from slow. A
		// stream parser that waits for a terminator the fake never sends
		// blocks forever, and that must read as a failure, not a timeout of
		// the whole grader.
		_ = cmd.Process.Kill()
		run.ExitErr = "exercise did not exit within 30s (stream parser may be waiting for input that never comes)"
	}

	run.Ran = true
	run.Stdout = stdout.String()
	run.Stderr = stderr.String()

	sc := bufio.NewScanner(strings.NewReader(run.Stdout))
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) != nil {
			continue
		}
		run.Lines = append(run.Lines, m)

		if s, ok := m["assistant"].(string); ok {
			run.Assistant = s
			continue
		}

		switch m["observation"] {
		case "part_delta":
			d := Ch7Delta{
				PartID: uint64(numOf(m["part_id"])),
				Kind:   strOf(m["kind"]),
				Chunk:  strOf(m["chunk"]),
			}
			run.Deltas = append(run.Deltas, d)
		case "part_final":
			id := uint64(numOf(m["part_id"]))
			kind := "thinking" // default: neither text nor tool means thinking/opaque
			if t, ok := m["text"].(string); ok {
				run.FinalText[id] = t
				kind = "text"
			}
			if t, ok := m["tool"].(string); ok {
				run.FinalTool[id] = t
				kind = "tool_call"
			}
			run.Finals = append(run.Finals, Ch7PartFinal{PartID: id, Kind: kind})
		case "stream_stats":
			run.Stats = append(run.Stats, m)
		}
	}
	return run
}

func numOf(v any) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func strOf(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
