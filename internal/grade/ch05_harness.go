package grade

// Chapter 5 harness: verify the refactoring produced a reusable framework.

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/coding-agents-course/internal/fakevendor"
)

// Ch5Result holds evidence for the ch5 checks.
type Ch5Result struct {
	AgentBuildOK  bool
	AgentBuildErr string
	ExBuildOK     bool
	ExBuildErr    string
	ImportGraph   map[string][]string
	ExToolCalled  bool
	ExToolOutput  string
	Base          string
	Ch4Result     *Ch4Result
	HelpersErr    string
}

func Ch5Run(dir string) (*Ch5Result, error) {
	r := &Ch5Result{ImportGraph: make(map[string][]string)}
	agentDir := filepath.Join(dir, "agent")
	exDir := filepath.Join(dir, "ch05")

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

	// Build agent binary from cmd/ (use absolute paths for Ch4Run).
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
		r.ExBuildErr = "ch05/ directory not found"
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

	// Drive exercise binary.
	if r.ExBuildOK {
		exBin := filepath.Join(exDir, "bin")
		r.ExToolCalled, r.ExToolOutput = driveCh5Ex(exBin, exDir)
	}

	// ch4 parity: run ch4's harness on the built agent binary.
	if r.AgentBuildOK {
		agentBin := filepath.Join(agentDir, "bin")
		ch4r, err := Ch4Run(agentBin)
		if err != nil {
			r.HelpersErr = err.Error()
		} else {
			r.Ch4Result = ch4r
		}
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

func driveCh5Ex(bin, workDir string) (toolCalled bool, toolOutput string) {
	// Two-turn conversation: first the fake asks the model to call calculate,
	// then it replies with the final answer.
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

	// The fake received at least two requests if the tool was called: the
	// first is the user prompt, the second includes the tool result.
	reqs := fake.Requests()
	if len(reqs) >= 2 {
		// The second request should contain the tool result.
		s := string(reqs[1].Body)
		if strings.Contains(s, "tool_result") || strings.Contains(s, "calculate") {
			toolCalled = true
			// Extract tool output from the request body.
			if i := strings.Index(s, "Result:"); i >= 0 {
				end := i + 40
				if end > len(s) {
					end = len(s)
				}
				toolOutput = s[i:end]
				// Trim to the next quote.
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
