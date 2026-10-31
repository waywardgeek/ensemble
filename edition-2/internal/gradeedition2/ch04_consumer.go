package gradeedition2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// ConsumerCase describes application behavior, without prescribing its Go API.
// The separate consumer runs its demonstration with no arguments and reads the
// existing LLM_* provider settings. These fields are grader input, not a new
// protocol the application must implement.
type ConsumerCase struct {
	// Name identifies the application's registered tool.
	Name string `json:"name"`
	// Arguments supplies the JSON object delivered to the real handler.
	Arguments json.RawMessage `json:"arguments"`
	// Result is the complete successful tool text expected at the provider.
	Result string `json:"result"`
	// Diagnostic identifies the timestamped line emitted by that tool.
	Diagnostic string `json:"diagnostic"`
}

// ReadConsumerCase loads a caller's example, or the accepted shout demonstration
// when path is empty. It rejects incomplete fixtures before running a submission.
func ReadConsumerCase(path string) (ConsumerCase, error) {
	example := ConsumerCase{"shout", json.RawMessage(`{"text":"hello framework"}`), "HELLO FRAMEWORK", "shout called"}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return example, err
		}
		example = ConsumerCase{}
		if err := json.Unmarshal(data, &example); err != nil {
			return example, err
		}
	}
	var args map[string]any
	if err := json.Unmarshal(example.Arguments, &args); err != nil || args == nil || example.Name == "" || example.Result == "" || example.Diagnostic == "" {
		return example, fmt.Errorf("consumer case needs name, object arguments, nonempty result and diagnostic")
	}
	return example, nil
}

func exerciseConsumer(dir string, example ConsumerCase) []grade.Check {
	tool := grade.Check{ID: "external-consumer", Title: "Separate consumer builds, declares and executes its custom tool", Points: 25, Earned: 25, Passed: true}
	log := grade.Check{ID: "tool-diagnostic", Title: "Consumer emits the expected timestamped tool diagnostic", Points: 10, Earned: 10, Passed: true}
	failed := func(err error) []grade.Check {
		fail(&tool, "%v", err)
		fail(&log, "consumer did not complete: %v", err)
		return []grade.Check{tool, log}
	}
	work, err := os.MkdirTemp("", "edition2-consumer-*")
	if err != nil {
		return failed(err)
	}
	defer os.RemoveAll(work)
	// Build the same module-root program whose dependency graph is inspected.
	// Reusing the CLI builder's optional cmd/ preference could inspect one
	// program while executing another.
	bin := filepath.Join(work, "consumer")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = dir
	if output, err := build.CombinedOutput(); err != nil {
		return failed(fmt.Errorf("consumer build: %w: %s", err, output))
	}
	const callID = "external-consumer-call"
	fake := fakevendor.New([]fakevendor.Reply{
		{ToolName: example.Name, ToolArgs: string(example.Arguments), ToolID: callID},
		{Text: "Consumer exercise complete."},
	})
	defer fake.Close()
	// The demonstration is one finite call, not an interactive session that
	// can wait indefinitely for input. Its real HTTP and job paths still run.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL="+fakevendor.Models["anthropic"], "LLM_RESOLVED_MODEL=", "LLM_API_KEY=fake-key", "LLM_BASE_URL="+fake.URL())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return failed(fmt.Errorf("consumer run: %w: %s", err, output))
	}
	requests := fake.Requests()
	if len(requests) < 2 {
		fail(&tool, "consumer did not send a subsequent request containing the result")
	} else {
		var first map[string]any
		if err := json.Unmarshal(requests[0].Body, &first); err != nil {
			fail(&tool, "decode declaration request: %v", err)
		}
		declared := false
		// A fake can name an undeclared tool. Successful dispatch alone therefore
		// cannot show that a real provider was told the tool exists.
		for _, raw := range list(first["tools"]) {
			decl := object(raw)
			description, _ := decl["description"].(string)
			if decl["name"] == example.Name && strings.TrimSpace(description) != "" && object(decl["input_schema"])["type"] == "object" {
				declared = true
			}
		}
		if !declared {
			fail(&tool, "custom tool %q lacks its description/object schema in the initial request", example.Name)
		}
		// The fake's final sentence is deliberately unrelated to the expected
		// result. Only the real, correlated tool_result earns credit: the old
		// grader mistook an unknown-tool error for successful custom dispatch.
		text, isError, err := anthropicToolText(requests[1].Body, callID)
		if err != nil || isError || text != example.Result {
			fail(&tool, "custom result: error=%v is_error=%t text=%q; expected %q", err, isError, text, example.Result)
		}
	}
	if !timestampedDiagnostic(string(output), example.Diagnostic) {
		fail(&log, "no timestamped line contains diagnostic %q", example.Diagnostic)
	}
	// Result and diagnostic earn independent scores: neither substitutes for
	// the other, and source review establishes who actually owns the writer.
	log.Details = append(log.Details, "output observation only; source review must verify actual parent-interface access to the root logger")
	return []grade.Check{tool, log}
}

// Accept ordinary Go logger dates and ISO/RFC3339 dates without choosing a
// logging API for students. Validate the time text as well as its shape.
func timestampedDiagnostic(output, diagnostic string) bool {
	stamp := regexp.MustCompile(`\d{4}[-/]\d{2}[-/]\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?`)
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, diagnostic) {
			continue
		}
		for _, value := range stamp.FindAllString(line, -1) {
			for _, layout := range []string{time.RFC3339Nano, "2006/01/02 15:04:05", "2006-01-02 15:04:05"} {
				if _, err := time.Parse(layout, value); err == nil {
					return true
				}
			}
		}
	}
	return false
}
