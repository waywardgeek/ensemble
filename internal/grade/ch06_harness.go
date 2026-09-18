package grade

// Chapter 6 harness: drive the student's binary as a subprocess, collecting
// observations, testing hint delivery during tool execution, and verifying
// multi-agent orchestration.
//
// Takes a single directory — the student's library tree. Builds the binary
// from cmd/, runs it, and tests behavior.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Ch6Result holds evidence for the ch6 checks.
type Ch6Result struct {
	BuildOK  bool
	BuildErr string

	// Observations collected from stdout.
	Observations []Ch6Observation

	// Whether a hint was observed before its surrounding tool completion.
	HintBeforeTool bool
	// The ordered list of observation types received.
	ObsOrder []string

	// Import graph for hub-clean check.
	ImportGraph map[string][]string
	Base        string

	// Log file content for replay check.
	LogContent string
	LogPath    string

	// Loud refusal: error message when unsupported media is used.
	LoudRefusal string

	// Multi-agent: observation agent IDs seen.
	AgentsSeen map[string]bool

	// State changes observed.
	StateChanges []Ch6Observation
	// Content observations (part_delta, part_final).
	ContentObs []Ch6Observation

	// ch5 parity result.
	Ch5Result *Ch5Result
	Ch5Err    string

	// Mutable globals.
	MutableGlobals []string

	// Error from helpers.
	HelpersErr string
}

// Ch6Observation is a single observation line from the binary.
type Ch6Observation struct {
	Observation string `json:"observation"`
	Agent       string `json:"agent"`
	Text        string `json:"text,omitempty"`
	Chunk       string `json:"chunk,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	Error       string `json:"error,omitempty"`
	PartID      uint64 `json:"part_id,omitempty"`
	Seq         int    `json:"seq,omitempty"`
	Pipeline    string `json:"pipeline,omitempty"`
	Assistant   string `json:"assistant,omitempty"`
}

func Ch6Run(dir string) (*Ch6Result, error) {
	dir, _ = filepath.Abs(dir)
	r := &Ch6Result{
		ImportGraph: make(map[string][]string),
		AgentsSeen:  make(map[string]bool),
	}

	// Discover module path.
	r.Base = DiscoverBase(dir)

	// Build binary.
	bin, cleanup, err := Build(dir)
	if err != nil {
		r.BuildErr = err.Error()
		return r, nil
	}
	defer cleanup()
	r.BuildOK = true

	// Inspect import graph.
	r.ImportGraph = DiscoverImportGraph(dir)

	// Check for mutable globals.
	r.MutableGlobals = detectMutableGlobals(dir)

	// Drive the binary with a specially crafted fake that includes
	// a slow tool to test hint delivery.
	ch6DriveExercise(r, bin, dir)

	// Test loud refusal: use the binary with an unknown model.
	testLoudRefusal(r, bin, dir)

	// ch5 parity.
	ch5r, err := Ch5Run(dir)
	if err != nil {
		r.Ch5Err = err.Error()
	} else {
		r.Ch5Result = ch5r
	}

	return r, nil
}

// ch6DriveExercise runs the binary and tests hint delivery,
// observation collection, and multi-agent coordination.
func ch6DriveExercise(r *Ch6Result, bin, workDir string) {
	replies := []fakevendor.Reply{
		// Turn 1: tool call for "think" (will sleep 3s)
		{ToolName: "think", ToolArgs: `{"seconds":3,"thought":"composing draft"}`, ToolID: "call_a1"},
		// Turn 1 continued: final text after tool
		{Text: "I have written a draft about the topic."},
		// Turn 2: just text (for multi-agent, the student may wire multiple actors)
		{Text: "I have improved the draft with better flow."},
		// Turn 3: just text
		{Text: "APPROVED. The draft is excellent."},
	}
	fake := fakevendor.New(replies)
	defer fake.Close()

	logPath := filepath.Join(workDir, "test_ch06.log")
	r.LogPath = logPath

	cmd := exec.Command(bin)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=fake-key",
		"LLM_BASE_URL="+fake.URL(),
		"CH02_LOG="+logPath,
	)

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		r.HelpersErr = "stdin pipe: " + err.Error()
		return
	}

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		r.HelpersErr = "start: " + err.Error()
		return
	}

	// Send the initial prompt.
	fmt.Fprintf(stdinPipe, `{"kind":"prompt","text":"Write about the ocean"}`+"\n")

	// Wait for the tool to be dispatched, then send a hint.
	time.Sleep(1500 * time.Millisecond)
	fmt.Fprintf(stdinPipe, `{"kind":"hint","text":"make it vivid","agent":"author"}`+"\n")

	// Close stdin to signal EOF.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(8 * time.Second)
		stdinPipe.Close()
	}()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
	case <-time.After(20 * time.Second):
		cmd.Process.Kill()
		r.HelpersErr = "timeout"
		return
	}
	wg.Wait()

	parseObservations(r, outBuf.String())

	if data, err := os.ReadFile(logPath); err == nil {
		r.LogContent = string(data)
	}
}

func parseObservations(r *Ch6Result, output string) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var obs Ch6Observation
		if err := json.Unmarshal([]byte(line), &obs); err != nil {
			continue
		}
		r.Observations = append(r.Observations, obs)
		r.ObsOrder = append(r.ObsOrder, obs.Observation)

		if obs.Agent != "" {
			r.AgentsSeen[obs.Agent] = true
		}

		switch obs.Observation {
		case "state_changed":
			r.StateChanges = append(r.StateChanges, obs)
		case "part_delta", "part_final":
			r.ContentObs = append(r.ContentObs, obs)
		}
	}

	hintSeen := false
	for _, obs := range r.Observations {
		if obs.Observation == "part_delta" && strings.HasPrefix(obs.Chunk, "hint:") {
			hintSeen = true
		}
		if obs.Observation == "part_final" && hintSeen {
			r.HintBeforeTool = true
			break
		}
	}
}

func testLoudRefusal(r *Ch6Result, bin, workDir string) {
	replies := []fakevendor.Reply{
		{Text: "Should not reach here."},
	}
	fake := fakevendor.New(replies)
	defer fake.Close()

	cmd := exec.Command(bin)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=unknown-model-xyz",
		"LLM_API_KEY=fake-key",
		"LLM_BASE_URL="+fake.URL(),
	)

	var inBuf bytes.Buffer
	inBuf.WriteString(`{"user":"test"}` + "\n")
	cmd.Stdin = &inBuf

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return
	}
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		return
	}

	combined := outBuf.String() + errBuf.String()
	if strings.Contains(combined, "unknown-model-xyz") ||
		strings.Contains(combined, "unsupported") ||
		strings.Contains(combined, "not supported") ||
		strings.Contains(combined, "unknown model") {
		r.LoudRefusal = combined
	}
}

func GradeCh6(dir string) ([]Check, string) {
	r, err := Ch6Run(dir)
	if err != nil {
		return nil, err.Error()
	}
	return Ch6Evaluate(r), r.HelpersErr
}

// unused but keeps the import happy
var _ = io.Discard
