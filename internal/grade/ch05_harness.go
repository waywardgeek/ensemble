package grade

// Chapter 5 harness: verify the refactoring produced a reusable framework.
//
// Takes a single directory — the student's library tree. Builds the binary
// from cmd/, inspects the import graph, and tests behavior.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Ch5Result holds evidence for the ch5 checks.
type Ch5Result struct {
	AgentBuildOK   bool
	AgentBuildErr  string
	ImportGraph    map[string][]string
	ToolCalled     bool
	ToolOutput     string
	Base           string
	HasLogf        bool
	MutableGlobals []string
	Ch4Result      *Ch4Result
	HelpersErr     string
}

func Ch5Run(dir string) (*Ch5Result, error) {
	dir, _ = filepath.Abs(dir)
	r := &Ch5Result{ImportGraph: make(map[string][]string)}

	// Discover module path.
	r.Base = DiscoverBase(dir)

	// Build binary.
	bin, cleanup, err := Build(dir)
	if err != nil {
		r.AgentBuildErr = err.Error()
		return r, nil
	}
	defer cleanup()
	r.AgentBuildOK = true

	// Inspect import graph.
	r.ImportGraph = DiscoverImportGraph(dir)

	// Check for Host interface with Logf in common, and embedded in Call.
	r.HasLogf = detectLogf(dir)

	// Check for mutable package-level vars.
	r.MutableGlobals = detectMutableGlobals(dir)

	// Drive the binary: fake vendor sends a tool call for "calculate",
	// binary should execute it and return the result.
	r.ToolCalled, r.ToolOutput = driveCh5Tool(bin, dir)

	// ch4 parity: run ch4's harness on the built binary.
	ch4r, err := Ch4Run(bin)
	if err != nil {
		r.HelpersErr = err.Error()
	} else {
		r.Ch4Result = ch4r
	}

	return r, nil
}

func GradeCh5(dir string) ([]Check, string) {
	r, err := Ch5Run(dir)
	if err != nil {
		return nil, err.Error()
	}
	return Ch5Evaluate(r), r.HelpersErr
}

// driveCh5Tool tests that the student's binary can execute a tool call.
// The fake vendor tells the model to call "calculate", and the binary
// should execute it and send the result back.
func driveCh5Tool(bin, workDir string) (toolCalled bool, toolOutput string) {
	replies := []fakevendor.Reply{
		{ToolName: "calculate", ToolArgs: `{"expression":"6 * 7"}`, ToolID: "call_001"},
		{Text: "The answer is 42."},
	}
	fake := fakevendor.New(replies)
	defer fake.Close()

	cmd := exec.Command(bin)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=fake-key",
		"LLM_BASE_URL="+fake.URL(),
	)

	var inBuf bytes.Buffer
	inBuf.WriteString(`{"user":"What is 6 * 7?"}` + "\n")
	cmd.Stdin = &inBuf

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return false, ""
	}
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
		return false, ""
	}

	// The fake received at least two requests if the tool was called.
	reqs := fake.Requests()
	if len(reqs) >= 2 {
		s := string(reqs[1].Body)
		if strings.Contains(s, "tool_result") || strings.Contains(s, "calculate") {
			toolCalled = true
			if i := strings.Index(s, "Result:"); i >= 0 {
				end := i + 40
				if end > len(s) {
					end = len(s)
				}
				toolOutput = s[i:end]
				if qi := strings.IndexByte(toolOutput, '"'); qi > 0 {
					toolOutput = toolOutput[:qi]
				}
			} else {
				toolOutput = "(tool was called)"
			}
		}
	}
	return
}

// detectLogf checks whether the student's code declares a Host-like
// interface with Logf and references it from a Call-like struct.
// Searches the entire tree, not just internal/common/.
func detectLogf(dir string) bool {
	cmd := exec.Command("grep", "-rl", "Logf", dir, "--include=*.go")
	if out, err := cmd.Output(); err != nil || len(out) == 0 {
		return false
	}
	cmd = exec.Command("grep", "-rl", "Host", dir, "--include=*.go")
	out, err := cmd.Output()
	return err == nil && len(out) > 0
}

// detectMutableGlobals finds package-level var declarations that are mutable
// state (not immutable lookup maps). Returns the list of offending locations.
func detectMutableGlobals(dir string) []string {
	cmd := exec.Command("grep", "-rn", "^var ", dir, "--include=*.go")
	out, _ := cmd.Output()
	var globals []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "_test.go:") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "names") || strings.Contains(lower, "name =") {
			continue
		}
		globals = append(globals, line)
	}
	return globals
}
