package gradeedition2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

func exerciseWorkflow(core, dir string) []grade.Check {
	checks := []grade.Check{
		{ID: "not-deaf", Title: "Hint is observed before the matching slow-tool result/report", Points: 25, Earned: 25, Passed: true},
		{ID: "observer-fires", Title: "Observers expose state changes and complete parts", Points: 15, Earned: 15, Passed: true},
		{ID: "multi-agent-observations", Title: "Three named roles emit sequential completion records", Points: 15, Earned: 15, Passed: true},
	}
	setupFailure := func(err error) []grade.Check {
		for i := range checks {
			fail(&checks[i], "%v", err)
		}
		return checks
	}
	work, err := os.MkdirTemp("", "edition2-workflow-*")
	if err != nil {
		return setupFailure(err)
	}
	defer os.RemoveAll(work)
	if grade.DiscoverBase(dir) == "" {
		return setupFailure(fmt.Errorf("workflow needs its own module: %s", dir))
	}
	if err := publicDependency(core, dir); err != nil {
		return setupFailure(fmt.Errorf("workflow dependency: %w", err))
	}
	// Never silently substitute the ordinary CLI for a missing external app.
	bin := filepath.Join(work, "workflow")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = dir
	if output, err := build.CombinedOutput(); err != nil {
		return setupFailure(fmt.Errorf("workflow build: %w: %s", err, output))
	}
	const callID = "call_a1"
	// Role completion must follow the real application tools. Text claiming a
	// draft or review exists is not a substitute for changing shared manuscript
	// state; the documented example schemas keep this fixture implementation-free.
	fake := fakevendor.New([]fakevendor.Reply{
		{ToolName: "think", ToolArgs: `{"seconds":3,"thought":"composing draft"}`, ToolID: callID},
		{ToolName: "write_draft", ToolArgs: `{"text":"The ocean carries waves to shore."}`, ToolID: "call_write"},
		{ToolName: "word_count", ToolArgs: `{}`, ToolID: "call_count"},
		{Text: "I have written a draft about the topic."},
		{ToolName: "read_draft", ToolArgs: `{}`, ToolID: "call_read"},
		{ToolName: "edit_draft", ToolArgs: `{"text":"Blue ocean waves break against the sunlit shore."}`, ToolID: "call_edit"},
		{Text: "I have improved the draft with better flow."},
		{ToolName: "review", ToolArgs: `{"decision":"accept","notes":"The edited draft is vivid and clear."}`, ToolID: "call_review"},
		{Text: "APPROVED. The draft is excellent."},
	})
	defer fake.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	cmd.Dir = work
	logPath := filepath.Join(work, "workflow.jsonl")
	cmd.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL=fake-model", "LLM_RESOLVED_MODEL=", "LLM_API_KEY=fake-key", "LLM_BASE_URL="+fake.URL(), "CH06_LOG="+logPath, "CH02_LOG="+logPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	input, err := cmd.StdinPipe()
	if err != nil {
		return setupFailure(err)
	}
	if err := cmd.Start(); err != nil {
		return setupFailure(err)
	}
	_, promptErr := fmt.Fprintln(input, `{"kind":"prompt","text":"Write about the ocean"}`)
	// Retain the original three-second tool and mid-tool delivery timing. EOF
	// ends input; the application must still await its requested workflow.
	time.Sleep(1500 * time.Millisecond)
	_, hintErr := fmt.Fprintln(input, `{"kind":"hint","text":"make it vivid","agent":"author"}`)
	_ = input.Close()
	if err := cmd.Wait(); err != nil || promptErr != nil || hintErr != nil {
		return setupFailure(fmt.Errorf("workflow run: %v; prompt %v; hint %v: %s", err, promptErr, hintErr, &stderr))
	}
	observations, err := jsonLines(stdout.Bytes())
	if err != nil {
		return setupFailure(fmt.Errorf("workflow stdout: %v", err))
	}
	hint, result := -1, -1
	state, final := false, false
	var ended []string
	seen := map[string]bool{}
	for i, obs := range observations {
		agent, _ := obs["agent"].(string)
		seen[agent] = true
		switch obs["observation"] {
		case "state_changed":
			// The discriminator alone proves no transition. Preserve the actual
			// old/new state payload without imposing an implementation's enum.
			from, _ := obs["from"].(string)
			to, _ := obs["to"].(string)
			state = state || (from != "" && to != "" && from != to)
		case "part_delta":
			chunk, _ := obs["chunk"].(string)
			if agent == "author" && strings.HasPrefix(chunk, "hint:") && strings.TrimSpace(strings.TrimPrefix(chunk, "hint:")) == "make it vivid" {
				hint = i
			}
		case "part_final":
			part := object(obs["part"])
			// A nonempty object (for example only {type:text}) is not completed
			// content. Require one actual final text supplied by the fake.
			text, _ := part["text"].(string)
			kind, _ := field(part, "type", "kind").(string)
			if normalize(kind) == "text" && (text == "I have written a draft about the topic." || text == "I have improved the draft with better flow." || text == "APPROVED. The draft is excellent.") {
				final = true
			}
			if agent == "author" && normalize(kind) == "toolresult" && field(part, "call_id", "callid") == callID {
				// Another completed part can arrive while think is still running.
				// Match the dispatched call, not merely the next finalizer.
				result = i
			}
		case "turn_ended":
			ended = append(ended, agent)
		}
	}
	if hint < 0 || result < 0 || hint >= result {
		fail(&checks[0], "author hint index %d, correlated think result/report index %d", hint, result)
	}
	// A default three-second callback can race the three-second handler. Its
	// result may honestly say running; neither this check nor its label claims
	// terminal completion. An unknown-tool error is not a slow-tool report.
	requests := fake.Requests()
	if len(requests) < 2 {
		fail(&checks[0], "no request delivered the think result")
	} else if _, isError, err := anthropicToolText(requests[1].Body, callID); err != nil || isError {
		fail(&checks[0], "think result: %v, is_error=%t", err, isError)
	}
	if !state || !final {
		fail(&checks[1], "state_changed=%t, complete part_final=%t; deltas alone are insufficient", state, final)
	}
	if strings.Join(ended, ",") != "author,editor,reviewer" || !seen["author"] || !seen["editor"] || !seen["reviewer"] {
		fail(&checks[2], "expected author/editor/reviewer completion order; got %v", ended)
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		fail(&checks[2], "workflow log: %v", err)
	} else if records, err := jsonLines(log); err != nil {
		fail(&checks[2], "workflow log: %v", err)
	} else {
		// Keep the old check's useful application-log coverage, without calling
		// observation counts or pipeline markers proof of core log replay.
		var stages []string
		loggedObservation := false
		for _, record := range records {
			if stage, ok := record["pipeline"].(string); ok {
				stages = append(stages, stage)
			}
			loggedObservation = loggedObservation || record["observation"] != nil
		}
		if strings.Join(stages, ",") != "author_done,editor_done,reviewer_done" || !loggedObservation {
			fail(&checks[2], "workflow log lacks ordered stage markers or observations: %v", stages)
		}
	}
	checks[2].Details = append(checks[2].Details, "records establish named sequential progress; real role tools, stored-draft handoff and reliable private request replies remain required source/integration review")
	return checks
}

func jsonLines(data []byte) ([]map[string]any, error) {
	var records []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
