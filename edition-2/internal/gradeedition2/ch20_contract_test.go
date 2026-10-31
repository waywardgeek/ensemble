package gradeedition2

import (
	"encoding/json"
	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestChapter20Environment(t *testing.T) {
	for _, key := range []string{"LLM_BASE_URL", "ANTHROPIC_BASE_URL", "OPENAI_BASE_URL", "GEMINI_BASE_URL", "LLM_API_KEY", "OPENAI_API_KEY", "EN_SANDBOX_ROOT", "EN_SAFE_MODE", "EN_SKILLS_DIR", "EN_PRIMARY_SKILL", "OPENAI_BILLING_MODE", "CH07_AGENT_LOG", "CH02_LOG", "E3_TRACE", "EDITION2_WEB_TOKEN"} {
		t.Setenv(key, "host-poison")
	}
	work := t.TempDir()
	skills := filepath.Join(work, "skills")
	endpoint := "http://127.0.0.1:34567"
	for _, dir := range []string{skills, ""} {
		env := chapter20Env(work, dir, endpoint)
		values := map[string]string{}
		for _, item := range env {
			k, v, _ := strings.Cut(item, "=")
			if _, ok := values[k]; ok {
				t.Fatalf("duplicate %s", k)
			}
			values[k] = v
			if strings.Contains(v, "host-poison") {
				t.Fatalf("inherited %s", k)
			}
		}
		want := map[string]string{"HOME": work, "TMPDIR": work, "LLM_MODEL": "claude-sonnet-4-6", "LLM_BASE_URL": endpoint, "LLM_API_KEY": "fake", "EN_SANDBOX_ROOT": "", "EN_SAFE_MODE": "false", "EN_WEB_SEARCH": "false", "EN_SKILLS_DIR": dir, "EN_PRIMARY_SKILL": "base", "LLM_SPEAKING": "false", "EN_DISABLE_STREAMING": "1"}
		if dir == "" {
			want["EN_PRIMARY_SKILL"] = ""
		}
		for k, v := range want {
			if got, ok := values[k]; !ok || got != v {
				t.Errorf("%s=%q want %q", k, got, v)
			}
		}
	}
}

func TestChapter20CurrentToolLimits(t *testing.T) {
	for _, name := range []string{"join_agents", "send_message_to_parent"} {
		original := &fakevendor.Reply{ToolID: "operation", ToolName: name, ToolArgs: `{"ai_callback_delay":60,"sentinel":"unchanged"}`}
		first := chapter20ToolLimits(nil, original)
		if first == original || len(first.Tools) != 0 || first.ToolID != "operation-limits" || first.ToolName != "tool_limits" {
			t.Fatalf("first response: %+v", first)
		}
		var limits map[string]any
		if err := json.Unmarshal([]byte(first.ToolArgs), &limits); err != nil || len(limits) != 1 || limits["ai_callback_delay"] != float64(60) {
			t.Fatalf("limits: %v %v", limits, err)
		}
		result := []byte(`{"messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"operation-limits","content":"configured"}]}]}`)
		next := chapter20ToolLimits(result, original)
		if len(next.Tools) != 0 || next.ToolID != "operation" || next.ToolName != name {
			t.Fatalf("operation: %+v", next)
		}
		var args map[string]any
		if err := json.Unmarshal([]byte(next.ToolArgs), &args); err != nil || len(args) != 1 || args["sentinel"] != "unchanged" {
			t.Fatalf("operation args: %v %v", args, err)
		}
		failed := []byte(`{"messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"operation-limits","is_error":true,"content":"refused"}]}]}`)
		if got := chapter20ToolLimits(failed, original); got.ToolName != "" || len(got.Tools) != 0 || !strings.Contains(got.Text, "failed") {
			t.Fatalf("failed setter authorized operation or retry: %+v", got)
		}
		wrong := []byte(`{"messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"other-limits","content":"configured"}]}]}`)
		if got := chapter20ToolLimits(wrong, original); got.ToolName != "tool_limits" {
			t.Fatalf("wrong result authorized operation: %+v", got)
		}
		if original.ToolArgs != `{"ai_callback_delay":60,"sentinel":"unchanged"}` || len(original.Tools) != 0 {
			t.Fatal("historical reply mutated")
		}
	}
	for _, original := range []*fakevendor.Reply{nil, {Text: "ordinary"}, {ToolID: "limits", ToolName: "tool_limits", ToolArgs: `{"ai_callback_delay":5}`}, {ToolID: "wait", ToolName: "wait_for_job", ToolArgs: `{"ai_callback_delay":5,"handle":1}`}, {ToolID: "run", ToolName: "run_command", ToolArgs: `{"ai_callback_delay":60,"command":"true"}`}, {ToolID: "call", ToolName: "join_agents", ToolArgs: `{"agent_ids":[]}`}} {
		if chapter20ToolLimits(nil, original) != original {
			t.Fatal("unrelated reply changed")
		}
	}
}

func TestChapter20ReplayVisibleContent(t *testing.T) {
	const marker = "PUBLIC_EVENT_22"
	frame := map[string]any{"type": "part_final", "part": map[string]any{"type": "text", "text": marker}}
	replay := map[string]any{"type": "replay", "data": map[string]any{"events": []any{frame}}}
	if !reuseReplayText(frame, marker, true) || !reuseReplayText(replay, marker, true) {
		t.Fatal("actual visible text lost")
	}
	if reuseReplayText(replay, marker, false) || !reuseReplayText(frame, marker, false) {
		t.Fatal("historical decoder changed")
	}
	for _, m := range []map[string]any{
		{"type": "replay", "text": marker, "data": map[string]any{"events": []any{}}},
		{"type": "replay", "data": map[string]any{"events": []any{map[string]any{"type": "part_final", "part": map[string]any{"type": "tool_call", "name": "probe", "args": map[string]any{"text": marker}}}}}},
		{"type": "replay", "data": map[string]any{"partials": []any{frame}}},
		{"type": "error", "text": marker},
	} {
		if reuseReplayText(m, marker, true) {
			t.Fatalf("non-dialogue marker accepted: %v", m)
		}
	}
}

func TestChapter20CurrentResults(t *testing.T) {
	notice := "[tool_limits consumed by this join_agents call: ai_callback_delay 1m0s, ai_callback_pattern none, max_output_bytes 16384]\n"
	payload := `{"outcomes":[{"agent_id":"middle","state":"submitted","submission":{"result":"MIDDLE_DONE_22"}}]}`
	if currentFleetJSON(notice+payload) != payload || !currentFleetSubmitted(notice+payload, "middle") {
		t.Fatal("valid correlated terminal outcome lost")
	}
	for _, raw := range []string{
		`{"questions":[{"agent_id":"middle","text":"MIDDLE_DONE_22"}],"outcomes":[]}`,
		`{"outcomes":[{"agent_id":"other","state":"submitted","submission":{"result":"MIDDLE_DONE_22"}}]}`,
		`{"outcomes":[{"agent_id":"middle","state":"failed","submission":{"result":"MIDDLE_DONE_22"}}]}`,
		`{"outcomes":[{"agent_id":"middle","state":"submitted"}]}`,
		"arbitrary prefix\n" + payload,
		"[tool_limits consumed by this join_agents call: ai_callback_delay invalid, ai_callback_pattern none, max_output_bytes 16384]\n" + payload,
	} {
		if currentFleetSubmitted(raw, "middle") {
			t.Fatalf("nonterminal/invalid evidence accepted: %s", raw)
		}
	}
	if _, err := fleetStatuses(notice+payload, true, false); err == nil {
		t.Fatal("legacy decoder changed")
	}
	typed := `[{"agent_id":"reviewer","state":"submitted","submission":{"summary":"STRUCTURED_RESULT_22","findings":[{"path":"main.go","fixed":false}],"follow_up":null}}]`
	rows, err := fleetStatuses(typed, false, true)
	if err != nil || len(rows) != 1 {
		t.Fatalf("status: %v %v", rows, err)
	}
	got := rows[0]["submission"].(map[string]any)
	if got["follow_up"] != nil || got["findings"].([]any)[0].(map[string]any)["fixed"] != false {
		t.Fatal("typed payload lost")
	}
}

func TestChapter20BrowserReplyOwnership(t *testing.T) {
	frame := map[string]any{"agent_id": "current-session", "source": "legacy-session"}
	payload := map[string]any{"jsonrpc": "2.0", "id": "outside-22", "result": map[string]any{"marker": "RELAY_RESULT_22"}}
	for _, current := range []bool{false, true} {
		got := reuseBrowserReply(frame, payload, current)
		key, other, want := "source", "agent_id", "legacy-session"
		if current {
			key, other, want = "agent_id", "source", "current-session"
		}
		if len(got) != 3 || got["type"] != "jsonrpc" || got[key] != want || got[other] != nil || got["payload"].(map[string]any)["id"] != "outside-22" {
			t.Fatalf("current=%v: %v", current, got)
		}
	}
}

func TestChapter20AddressedReplyAttribution(t *testing.T) {
	for _, raw := range []string{`middle ANSWER_22`, `{"request_id":"middle-question-1","text":"ANSWER_22"}`, `{"sender":"other","request_id":"middle-question-1","text":"ANSWER_22"}`, `{"sender":"middle","text":"ANSWER_22"`} {
		if _, err := fleetAddressedReply(raw, true); err == nil {
			t.Fatalf("invalid current addressed reply accepted: %s", raw)
		}
	}
	got, err := fleetAddressedReply(`{"sender":"middle","recipient":"leaf","request_id":"middle-question-1","text":"ANSWER_22"}`, true)
	if err != nil {
		t.Fatal(err)
	}
	var message map[string]any
	if err := json.Unmarshal([]byte(got), &message); err != nil {
		t.Fatal(err)
	}
	if message["sender"] != "middle" || message["text"] != "ANSWER_22" || message["request_id"] != nil {
		t.Fatalf("attribution: %v", message)
	}
	if got, err := fleetAddressedReply("middle ANSWER_22", false); err != nil || got != "middle ANSWER_22" {
		t.Fatal("historical fallback changed")
	}
}

func TestChapter20StatusOutcomeWrapper(t *testing.T) {
	const valid = `[{"agent_id":"child","parent_id":"root","state":"submitted","outcome":{"agent_id":"child","parent_id":"root","state":"submitted","submission":{"result":"done","nested":[false,null]}}}]`
	rows, err := fleetStatuses(valid, false, true)
	if err != nil || len(rows) != 1 || !reflect.DeepEqual(rows[0]["submission"], fleetJSON(`{"result":"done","nested":[false,null]}`)) {
		t.Fatalf("wrapped payload: %v %v", rows, err)
	}
	for _, bad := range []string{
		`[{"agent_id":"child","state":"submitted","outcome":"done"}]`,
		`[{"agent_id":"child","state":"submitted","outcome":{"agent_id":"other","state":"submitted"}}]`,
		`[{"agent_id":"child","parent_id":"root","state":"submitted","outcome":{"agent_id":"child","parent_id":"other","state":"submitted"}}]`,
		`[{"agent_id":"child","state":"submitted","outcome":{"agent_id":"child","state":"failed"}}]`,
	} {
		if _, err := fleetStatuses(bad, false, true); err == nil {
			t.Fatalf("accepted conflicting wrapper: %s", bad)
		}
	}
	legacy, err := fleetStatuses(valid, false, false)
	if err != nil || legacy[0]["submission"] != nil {
		t.Fatalf("changed legacy decoding: %v %v", legacy, err)
	}
}

func TestChapter20CurrentReplyAndSavePath(t *testing.T) {
	got, err := fleetAddressedReply(`{"Sender":"middle","Recipient":"leaf","RequestID":"middle-question-1","Text":"ANSWER_22"}`, true)
	if err != nil || strings.Contains(got, "RequestID") || !strings.Contains(got, "ANSWER_22") {
		t.Fatalf("public reply: %s %v", got, err)
	}
	for _, bad := range []string{`{"Sender":"other","RequestID":"middle-question-1","Text":"ANSWER_22"}`, `{"RequestID":"middle-question-1","Text":"ANSWER_22"}`, `{"Sender":"other","sender":"middle","Text":"ANSWER_22"}`, `{"Sender":{},"sender":{},"Text":"ANSWER_22"}`, `{"Sender":[],"sender":[],"Text":"ANSWER_22"}`, `{"Sender":null,"Text":"ANSWER_22"}`} {
		if _, err := fleetAddressedReply(bad, true); err == nil {
			t.Fatalf("accepted sender spoof: %s", bad)
		}
	}
	ids, err := fleetAddressedReply(`{"Sender":"middle","TurnID":"ANSWER_22","turn_id":"ANSWER_22","RequestID":"ANSWER_22","Text":"ordinary"}`, true)
	if err != nil || strings.Contains(ids, "ANSWER_22") {
		t.Fatalf("correlation masquerades as content: %s %v", ids, err)
	}
	row := map[string]any{"agent_id": "reviewer"}
	path := filepath.Join("owned", "submit.md")
	if got, err := fleetSubmissionSavePath(path, row, true); err != nil || got != filepath.Join("owned", "reviewer", "save.json") {
		t.Fatalf("current save: %s %v", got, err)
	}
	if got, err := fleetSubmissionSavePath(path, row, false); err != nil || got != filepath.Join("owned", "save.json") {
		t.Fatalf("legacy save: %s %v", got, err)
	}
	row["agent_id"] = "other"
	if _, err := fleetSubmissionSavePath(path, row, true); err == nil {
		t.Fatal("accepted wrong row owner")
	}
}

func TestChapter20UndispatchedObservation(t *testing.T) {
	const skipped = "interrupted, not executed: live input arrived before dispatch"
	if !fleetUndispatchedJoin(skipped, true) || fleetUndispatchedJoin(skipped, false) || fleetUndispatchedJoin("other failure", true) || fleetUndispatchedJoin("prefix "+skipped, true) {
		t.Fatal("undispatched observation classification")
	}
}
func TestChapter20AddressedMessagingCLI(t *testing.T) {
	bin := os.Getenv("CH20_TEST_CLI")
	if bin == "" {
		t.Skip("set CH20_TEST_CLI to an explicitly pinned current CLI")
	}
	c := grade.Check{ID: "addressed-messaging", Passed: true, Points: 15, Earned: 15}
	fleetMessagingWith(fleetDriver{bin: bin, current: true}, &c)
	if !c.Passed {
		t.Fatalf("addressed messaging: %v", c.Details)
	}
}

func TestChapter20RunningQuestionJob(t *testing.T) {
	for _, tc := range []struct {
		text string
		want int
	}{
		{"job 5 running; 0 bytes total; output: /tmp/out", 5},
		{"job 5 running; 120 bytes total; output: /tmp/out\npartial", 5},
		{"job 5 done; 0 bytes total; output: /tmp/out", 0},
		{"job 5 failed; 0 bytes total; output: /tmp/out", 0},
		{"job 5 killed; 0 bytes total; output: /tmp/out", 0},
		{"job 0 running; 0 bytes total", 0},
		{"job -5 running; 0 bytes total", 0},
		{"job 99999999999999999999999999999 running; 0 bytes total", 0},
		{"quoted job 5 running; 0 bytes total", 0},
		{"handle: 5 running; 0 bytes total", 0},
		{"job 5 running-ish; 0 bytes total", 0},
		{"{\"question\":\"job 5 running;\"}", 0},
		{"", 0},
	} {
		if got := fleetRunningJob(tc.text); got != tc.want {
			t.Errorf("%q: got %d want %d", tc.text, got, tc.want)
		}
	}
}
