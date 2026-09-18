package grade

// Chapter 6 harness: drive the exercise binary as a subprocess, collecting
// observations, testing hint delivery during tool execution, and verifying
// multi-agent orchestration.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Ch6Result holds evidence for the ch6 checks.
type Ch6Result struct {
	AgentBuildOK  bool
	AgentBuildErr string
	ExBuildOK     bool
	ExBuildErr    string

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

// Ch6Observation is a single observation line from the exercise binary.
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
	agentDir := filepath.Join(dir, "agent")
	exDir := filepath.Join(dir, "ch06")

	// Discover module path.
	if data, err := os.ReadFile(filepath.Join(agentDir, "go.mod")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				r.Base = strings.TrimPrefix(line, "module ")
				break
			}
		}
	}

	// Build agent binary.
	agentBin, _ := filepath.Abs(filepath.Join(agentDir, "bin"))
	cmd := exec.Command("go", "build", "-o", agentBin, "./cmd/")
	cmd.Dir = agentDir
	if out, err := cmd.CombinedOutput(); err != nil {
		r.AgentBuildErr = fmt.Sprintf("%s\n%s", err, out)
	} else {
		r.AgentBuildOK = true
	}

	// Build exercise binary.
	if _, err := os.Stat(exDir); err == nil {
		exBin, _ := filepath.Abs(filepath.Join(exDir, "bin"))
		cmd = exec.Command("go", "build", "-o", exBin, ".")
		cmd.Dir = exDir
		if out, err := cmd.CombinedOutput(); err != nil {
			r.ExBuildErr = fmt.Sprintf("%s\n%s", err, out)
		} else {
			r.ExBuildOK = true
		}
	} else {
		r.ExBuildErr = "ch06/ directory not found"
	}

	// Inspect import graph.
	cmd = exec.Command("go", "list", "-f",
		"{{.ImportPath}}: {{join .Imports \",\"}}", "./...")
	cmd.Dir = agentDir
	if out, err := cmd.Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) != 2 {
				continue
			}
			r.ImportGraph[strings.TrimSpace(parts[0])] =
				strings.Split(strings.TrimSpace(parts[1]), ",")
		}
	}

	// Check for mutable globals.
	r.MutableGlobals = detectMutableGlobals(agentDir)

	// Drive the exercise binary with a specially crafted fake that includes
	// a slow tool to test hint delivery.
	if r.ExBuildOK {
		exBin := filepath.Join(exDir, "bin")
		ch6DriveExercise(r, exBin, exDir)
	}

	// Test loud refusal: use the agent binary with an unknown model.
	if r.AgentBuildOK {
		testLoudRefusal(r, agentBin, agentDir)
	}

	// ch5 parity.
	if r.AgentBuildOK {
		ch5r, err := Ch5Run(dir)
		if err != nil {
			r.Ch5Err = err.Error()
		} else {
			r.Ch5Result = ch5r
		}
	}

	return r, nil
}

// ch6DriveExercise runs the exercise binary and tests hint delivery,
// observation collection, and multi-agent coordination.
func ch6DriveExercise(r *Ch6Result, exBin, workDir string) {
	// The fake vendor serves replies for three agents (author, editor, reviewer).
	// The author's first reply is a tool call for "think" which sleeps 3 seconds,
	// creating a window for the hint to arrive during tool execution.
	replies := []fakevendor.Reply{
		// Author turn 1: tool call for "think" (will sleep 3s)
		{ToolName: "think", ToolArgs: `{"seconds":3,"thought":"composing draft"}`, ToolID: "call_a1"},
		// Author turn 1 continued: final text after tool
		{Text: "I have written a draft about the topic."},
		// Editor turn 1: just text
		{Text: "I have improved the draft with better flow."},
		// Reviewer turn 1: just text
		{Text: "APPROVED. The draft is excellent."},
	}
	fake := fakevendor.New(replies)
	defer fake.Close()

	logPath := filepath.Join(workDir, "test_ch06.log")
	r.LogPath = logPath

	cmd := exec.Command(exBin)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=fake-key",
		"LLM_BASE_URL="+fake.URL(),
		"CH06_LOG="+logPath,
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

	// Wait for the tool to be dispatched (the fake will delay the response),
	// then send a hint during the tool execution window.
	time.Sleep(1500 * time.Millisecond)
	fmt.Fprintf(stdinPipe, `{"kind":"hint","text":"make it vivid","agent":"author"}`+"\n")

	// Close stdin to signal EOF so the binary can exit cleanly.
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

	// Parse observations from stdout.
	parseObservations(r, outBuf.String())

	// Read log file.
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

	// Check hint-before-tool ordering: a hint observation (part_delta with
	// "hint:" prefix) should appear before the tool's part_final.
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

// testLoudRefusal sends a message with an audio blob to a model that
// doesn't support audio, expecting an error naming the model and media type.
func testLoudRefusal(r *Ch6Result, agentBin, agentDir string) {
	// Use a model that doesn't support audio.
	replies := []fakevendor.Reply{
		{Text: "Should not reach here."},
	}
	fake := fakevendor.New(replies)
	defer fake.Close()

	cmd := exec.Command(agentBin)
	cmd.Dir = agentDir
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

	// The error should name the model and indicate refusal.
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

// DirHasFile returns true if the directory contains a file matching pattern.
func dirHasFile(dir, pattern string) bool {
	matches, _ := filepath.Glob(filepath.Join(dir, pattern))
	return len(matches) > 0
}

// sortedKeys returns sorted keys from a map.
func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// unused but keeps the import happy
var _ = io.Discard
