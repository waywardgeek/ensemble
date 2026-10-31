package gradeedition2

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

func replayRequests(bin string) grade.Check {
	c := grade.Check{ID: "replay-is-live", Title: "Persisted request prefixes render the actual captured HTTP bytes", Points: 15, Earned: 15, Passed: true}
	work, err := os.MkdirTemp("", "edition2-replay-*")
	if err != nil {
		fail(&c, "%v", err)
		return c
	}
	defer os.RemoveAll(work)
	fake := fakevendor.New([]fakevendor.Reply{
		{ToolName: "run_command", ToolArgs: `{"command":"printf REPLAY_TOOL_RESULT"}`, ToolID: "replay-tool"},
		{Text: "First turn complete."},
		{Text: "Second turn complete."},
	})
	defer fake.Close()
	// Use the exercise's existing fake-model fixture for both paths. A fake's
	// response identity is provenance, not a capability-table routing choice.
	const model = "fake-model"
	if _, err := runModel(bin, work, "anthropic", model, fake.URL(), "{\"user\":\"first turn\"}\n{\"user\":\"second turn\"}\n"); err != nil {
		fail(&c, "capture run: %v", err)
		return c
	}
	requests := fake.Requests()
	if len(requests) != 3 {
		fail(&c, "expected three requests across a tool round and two user turns, got %d", len(requests))
		return c
	}
	// A repeated empty log could trivially match an empty request. Establish
	// that the real tool result reached the captured request before replaying it.
	text, isError, err := anthropicToolText(requests[1].Body, "replay-tool")
	if err != nil || isError || !bytes.Contains([]byte(text), []byte("REPLAY_TOOL_RESULT")) {
		fail(&c, "captured tool result missing: text=%q is_error=%t error=%v", text, isError, err)
		return c
	}
	log, err := os.ReadFile(filepath.Join(work, "input.log"))
	if err != nil {
		fail(&c, "read captured event log: %v", err)
		return c
	}
	lines := bytes.Split(bytes.TrimSpace(log), []byte("\n"))
	index := 0
	for i, line := range lines {
		var event map[string]any
		if err := json.Unmarshal(line, &event); err != nil {
			fail(&c, "event log line %d: %v", i+1, err)
			return c
		}
		kind, _ := event["type"].(string)
		if normalize(kind) != "requestsent" {
			continue
		}
		if index >= len(requests) {
			fail(&c, "more request markers than captured HTTP requests")
			return c
		}
		// request_sent consumes pending input. Render the prefix BEFORE it,
		// using the same model/configuration and CLI builtin registry as capture.
		path := filepath.Join(work, "prefix.jsonl")
		if err := os.WriteFile(path, bytes.Join(lines[:i], []byte("\n")), 0600); err != nil {
			fail(&c, "%v", err)
			return c
		}
		body, err := runModel(bin, work, "anthropic", model, fake.URL(), "", "render", path)
		if err != nil || !bytes.Equal(bytes.TrimSpace([]byte(body)), bytes.TrimSpace(requests[index].Body)) {
			fail(&c, "request %d differs from replayed prefix: %v", index+1, err)
		}
		index++
	}
	if index != len(requests) {
		fail(&c, "replayed %d request prefixes for %d HTTP requests", index, len(requests))
	}
	if c.Passed {
		c.Details = []string{"three request bodies match byte for byte; only surrounding transport whitespace is ignored", "mid-HTTP hint consumption is additionally protected by the required source regression; this probe covers two turns and an actual tool result"}
	}
	return c
}
