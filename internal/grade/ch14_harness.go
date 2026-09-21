package grade

// Chapter 14 is graded through the speech channel itself.
//
// Every other grader in this book reads the student's code. This one does
// not, and the reason is worth stating plainly: the speech channel has no
// single shape. One reader builds it in a browser with speechSynthesis,
// another pipes text to a screen reader, another writes a terminal agent
// whose output is already linear. A grader that loads their modules and
// calls their methods would be grading whether they kept our names.
//
// So the contract is a log and a script. The student declares a way to run
// their system with the speech engine replaced by a recorder, we script the
// model's side of the conversation, and we read what came out. Nothing here
// names a method, a module, a file layout, or a language.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// ch14Entry is one line of the speech log: one thing the system said, or one
// moment it stopped saying things.
type ch14Entry struct {
	MS    int    `json:"ms"`
	Kind  string `json:"kind"`
	Seq   int    `json:"seq"`
	Text  string `json:"text"`
	Raw   string `json:"raw"`
	Cause string `json:"cause"`
}

// ch14Manifest is what the student's harness prints for --describe. Every
// field has a default that matches the reference, so a reader who kept our
// tool names writes no manifest at all.
type ch14Manifest struct {
	ReadFileTool string `json:"read_file_tool"`
	ReadFileArg  string `json:"read_file_arg"`
	SupportsType bool   `json:"supports_type_during_turn"`
	SupportsOff  bool   `json:"supports_streaming_off"`
	LogFormat    string `json:"log_format"`
}

// ch14Scenario is one launch of the student's system.
//
// Scenarios are grouped by mode rather than one per check, because a launch
// costs seconds and a mode costs nothing to share. The streamed-prose run
// feeds four checks at once.
type ch14Scenario struct {
	name      string
	prompt    string
	replies   []fakevendor.Reply
	streaming string            // "off" disables streaming; "" leaves it on
	typeText  string            // typed mid-turn, for the gate
	plant     map[string]string // files written into the workspace first
	brokenLLM bool              // point the system at an endpoint that fails
}

// ch14Discover finds the student's harness script.
func ch14Discover(dir string) (string, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	// The conventional location first, then anywhere in the tree, so a
	// reader who filed it elsewhere is found rather than failed.
	candidates := []string{
		filepath.Join(root, "scripts", "tts-harness.sh"),
		filepath.Join(root, "tts-harness.sh"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	var found string
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || found != "" || d.IsDir() {
			return nil
		}
		if strings.Contains(d.Name(), "tts-harness") {
			found = p
		}
		return nil
	})
	if found != "" {
		return found, nil
	}
	return "", fmt.Errorf("no speech harness found: expected scripts/tts-harness.sh under %s", dir)
}

// ch14Describe asks the harness to describe itself, falling back to defaults
// that match the reference implementation.
func ch14Describe(script string) ch14Manifest {
	man := ch14Manifest{
		ReadFileTool: "read_file",
		ReadFileArg:  "path",
		SupportsType: true,
		SupportsOff:  true,
		LogFormat:    "jsonl",
	}
	cmd := exec.Command(script, "--describe")
	out, err := cmd.Output()
	if err != nil {
		return man
	}
	var got ch14Manifest
	if err := json.Unmarshal(out, &got); err != nil {
		return man
	}
	if got.ReadFileTool != "" {
		man.ReadFileTool = got.ReadFileTool
	}
	if got.ReadFileArg != "" {
		man.ReadFileArg = got.ReadFileArg
	}
	if got.LogFormat != "" {
		man.LogFormat = got.LogFormat
	}
	man.SupportsType = got.SupportsType
	man.SupportsOff = got.SupportsOff
	return man
}

// ch14Run performs one scenario: plant the files, script the model, run the
// student's harness, and read back what was spoken.
func ch14Run(script string, sc ch14Scenario) ([]ch14Entry, string, error) {
	work, err := os.MkdirTemp("", "ch14-work-")
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(work)

	for name, body := range sc.plant {
		if err := os.WriteFile(filepath.Join(work, name), []byte(body), 0o644); err != nil {
			return nil, "", err
		}
	}

	// A scenario can ask for an endpoint that fails, which is how the
	// grader provokes the error a listener would otherwise never hear
	// about. It is a real failure on the real path, not a synthesized
	// message injected behind the system's back.
	var baseURL string
	if sc.brokenLLM {
		bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"error":{"message":"upstream is on fire"}}`, http.StatusInternalServerError)
		}))
		defer bad.Close()
		baseURL = bad.URL
	} else {
		srv := fakevendor.New(sc.replies)
		defer srv.Close()
		baseURL = srv.URL()
	}

	logFile := filepath.Join(work, "tts.log")

	args := []string{sc.prompt}
	if sc.typeText != "" {
		args = append(args, "--type-during-turn", sc.typeText)
	}

	cmd := exec.Command(script, args...)
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+baseURL,
		"TTS_LOG="+logFile,
		"WORKSPACE="+work,
	)
	if sc.streaming != "" {
		cmd.Env = append(cmd.Env, "STREAMING="+sc.streaming)
	}

	var stderr strings.Builder
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr

	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return nil, stderr.String(), err
	}
	go func() { done <- cmd.Wait() }()

	select {
	case err = <-done:
	case <-time.After(90 * time.Second):
		_ = cmd.Process.Kill()
		<-done
		return nil, stderr.String(), fmt.Errorf("harness did not finish within 90s")
	}

	entries, readErr := ch14ReadLog(logFile)
	if readErr != nil {
		if err != nil {
			return nil, stderr.String(), fmt.Errorf("harness exited with %v and wrote no readable log: %v", err, readErr)
		}
		return nil, stderr.String(), readErr
	}
	return entries, stderr.String(), nil
}

// ch14ReadLog parses the speech log. Unparseable lines are skipped rather
// than fatal: a log a human can also read may reasonably carry a comment.
func ch14ReadLog(path string) ([]ch14Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no speech log at %s: %v", path, err)
	}
	defer f.Close()

	var entries []ch14Entry
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var e ch14Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, sc.Err()
}

// ch14Spoken returns just the utterances, in order.
func ch14Spoken(entries []ch14Entry) []string {
	var out []string
	for _, e := range entries {
		if e.Kind == "utterance" {
			out = append(out, e.Text)
		}
	}
	return out
}

// ch14AllSpeech joins every utterance, for checks that care about what was
// said rather than how it was divided.
func ch14AllSpeech(entries []ch14Entry) string {
	return strings.Join(ch14Spoken(entries), " ")
}
