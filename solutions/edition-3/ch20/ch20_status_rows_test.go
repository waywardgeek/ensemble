package ensemble_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"ensemble"
)

const ch20NestedSchema = `{"type":"object","properties":{"summary":{"type":"string"},"report":{"type":"object","properties":{"total":{"type":"integer","const":9007199254740993},"ratio":{"type":"number","const":0.10000000000000001},"fixed":{"type":"boolean"},"items":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"detail":{"type":["null","string"]}},"required":["name","detail"],"additionalProperties":false}}},"required":["total","ratio","fixed","items"],"additionalProperties":false}},"required":["summary","report"],"additionalProperties":false}`
const ch20NestedPayload = `{"summary":"reviewed","report":{"total":9007199254740993,"ratio":0.10000000000000001,"fixed":false,"items":[{"name":"one","detail":null},{"name":"two","detail":"precise"}]}}`

func exactJSON(t *testing.T, data []byte) any {
	t.Helper()
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var value any
	if err := d.Decode(&value); err != nil {
		t.Fatalf("JSON %s: %v", data, err)
	}
	return value
}

func ch20SameOutcome(a, b ensemble.AgentOutcome) bool {
	return a.AgentID == b.AgentID && a.ParentID == b.ParentID && a.State == b.State && a.Error == b.Error &&
		bytes.Equal(a.Submission, b.Submission)
}

func TestCh20StatusProjectsOutcomes(t *testing.T) {
	started, release := make(chan struct{}, 1), make(chan struct{})
	var once sync.Once
	root, parent := ch20Rig(t, func(w http.ResponseWriter, body []byte) {
		switch {
		case bytes.Contains(body, []byte("custom-status")):
			started <- struct{}{}
			<-release
			ch20Reply(w, ch20NestedPayload, true)
		case bytes.Contains(body, []byte("default-status")):
			ch20Reply(w, `{"result":"normalized default","ignored":{"nested":false}}`, true)
		default:
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, `{"error":{"message":"deliberate child failure"}}`)
		}
	})
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	ctx := ch20Context(t)
	custom, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "custom", Prompt: "custom-status", SubmitSchema: json.RawMessage(ch20NestedSchema)})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("custom turn did not start")
	}
	running, err := root.Status(custom.ID())
	if err != nil || running.State != "running" || running.Outcome != nil || running.ParentID != parent.ID() {
		t.Fatalf("running status %+v %v", running, err)
	}
	once.Do(func() { close(release) })
	customOutcome, err := root.Wait(ctx, custom.ID())
	if err != nil || customOutcome.State != "submitted" || !reflect.DeepEqual(exactJSON(t, customOutcome.Submission), exactJSON(t, []byte(ch20NestedPayload))) {
		t.Fatalf("custom outcome %+v %v", customOutcome, err)
	}
	ordinary, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "default", Prompt: "default-status"})
	if err != nil {
		t.Fatal(err)
	}
	defaultOutcome, err := root.Wait(ctx, ordinary.ID())
	if err != nil || string(defaultOutcome.Submission) != `{"result":"normalized default"}` {
		t.Fatalf("default outcome %+v %v", defaultOutcome, err)
	}
	failed, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "failed", Prompt: "failed-status"})
	if err != nil {
		t.Fatal(err)
	}
	failedOutcome, err := root.Wait(ctx, failed.ID())
	if err != nil || failedOutcome.State != "failed" || failedOutcome.Error == "" || len(failedOutcome.Submission) != 0 {
		t.Fatalf("failed outcome %+v %v", failedOutcome, err)
	}
	outcomes := map[string]ensemble.AgentOutcome{custom.ID(): customOutcome, ordinary.ID(): defaultOutcome, failed.ID(): failedOutcome}
	assertStatus := func(status ensemble.AgentStatus) {
		t.Helper()
		want, found := outcomes[status.AgentID]
		if !found || status.Outcome == nil || status.State != want.State || status.ParentID != parent.ID() || !ch20SameOutcome(*status.Outcome, want) {
			t.Fatalf("status does not project retained outcome: %+v want %+v", status, want)
		}
	}
	for id := range outcomes {
		status, err := root.Status(id)
		if err != nil {
			t.Fatal(err)
		}
		assertStatus(status)
	}
	children := root.Children(parent.ID())
	if len(children) != 3 {
		t.Fatalf("children=%+v", children)
	}
	for _, child := range children {
		assertStatus(child)
		child.Outcome.AgentID = "corrupted-copy"
		if len(child.Outcome.Submission) > 0 {
			child.Outcome.Submission[0] = '!'
		}
	}
	status, err := root.Status(custom.ID())
	if err != nil {
		t.Fatal(err)
	}
	assertStatus(status)
	status.Outcome.Submission[0] = '!'
	status.Outcome.State = "changed-copy"
	status, err = root.Status(custom.ID())
	if err != nil {
		t.Fatal(err)
	}
	assertStatus(status)
	parts, failedCall := parent.Tools().Dispatch(ctx, "agents_status", json.RawMessage(`{}`))
	var text strings.Builder
	for _, part := range parts {
		if part.Type == "text" {
			text.WriteString(part.Text)
		}
	}
	wire := text.String()
	if failedCall {
		t.Fatal(wire)
	}
	var toolStatuses []ensemble.AgentStatus
	for _, line := range strings.Split(wire, "\n") {
		if json.Unmarshal([]byte(line), &toolStatuses) == nil && len(toolStatuses) == 3 {
			break
		}
	}
	if len(toolStatuses) != 3 {
		t.Fatalf("no status projection in tool output: %s", wire)
	}
	for _, child := range toolStatuses {
		assertStatus(child)
	}
	joined, err := root.Join(ctx, parent.ID(), []string{custom.ID(), ordinary.ID(), failed.ID()})
	if err != nil || len(joined.Outcomes) != 3 {
		t.Fatalf("join %+v %v", joined, err)
	}
	for _, outcome := range joined.Outcomes {
		if !ch20SameOutcome(outcome, outcomes[outcome.AgentID]) {
			t.Fatalf("join changed retained outcome: %+v", outcome)
		}
	}
}

func TestCh20CustomDurableRowAndRecovery(t *testing.T) {
	root, parent := ch20Rig(t, func(w http.ResponseWriter, _ []byte) { ch20Reply(w, ch20NestedPayload, true) })
	ctx := ch20Context(t)
	child, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "reviewer", Prompt: "review", SubmitSchema: json.RawMessage(ch20NestedSchema)})
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := root.Wait(ctx, child.ID())
	if err != nil || outcome.State != "submitted" {
		t.Fatalf("submit %+v %v", outcome, err)
	}
	parentConfig, childConfig := parent.Config(), child.Config()
	path := filepath.Join(childConfig.DataDir, "submit.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(raw, []byte("\n")) != 1 {
		t.Fatalf("not one JSONL row: %s", raw)
	}
	row := exactJSON(t, raw).(map[string]any)
	if row["version"] != json.Number("1") || row["agent_id"] != child.ID() || row["parent_id"] != parent.ID() || row["timestamp"] == nil {
		t.Fatalf("row version/owner metadata: %s", raw)
	}
	if _, exists := row["v"]; exists {
		t.Fatal("legacy v field written")
	}
	if _, exists := row["custom"]; exists {
		t.Fatal("custom flag substituted for schema")
	}
	if !reflect.DeepEqual(row["submit_schema"], exactJSON(t, []byte(ch20NestedSchema))) || !reflect.DeepEqual(row["payload"], exactJSON(t, []byte(ch20NestedPayload))) {
		t.Fatalf("row lost immutable schema or exact payload: %s", raw)
	}
	if err := root.Shutdown(); err != nil {
		t.Fatal(err)
	}
	reopen := func(config ensemble.Config) (*ensemble.Ensemble, ensemble.Agent, error) {
		t.Helper()
		e := ensemble.New(log.New(io.Discard, "", 0))
		t.Cleanup(func() { _ = e.Shutdown() })
		if _, err := e.NewAgent(parentConfig); err != nil {
			t.Fatal(err)
		}
		a, err := e.NewAgent(config)
		return e, a, err
	}
	restored, again, err := reopen(childConfig)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := restored.Wait(ctx, again.ID())
	if err != nil || (recovered.AgentID != outcome.AgentID || recovered.ParentID != outcome.ParentID || recovered.State != outcome.State || recovered.Error != outcome.Error || !bytes.Equal(recovered.Submission, outcome.Submission)) {
		t.Fatalf("recovered %+v want %+v error=%v", recovered, outcome, err)
	}
	status, err := restored.Status(again.ID())
	if err != nil || status.Outcome == nil || !ch20SameOutcome(*status.Outcome, recovered) {
		t.Fatalf("recovered status %+v %v", status, err)
	}
	if result := <-again.Submit(ctx, "must remain terminal"); result.Err == nil {
		t.Fatal("terminal recovery admitted work")
	}
	if err := restored.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if now, _ := os.ReadFile(path); !bytes.Equal(now, raw) {
		t.Fatal("terminal recovery appended another row")
	}
	for _, suffix := range []string{"not json\n", `{"version":2,"submit_schema":{},"payload":{}}` + "\n"} {
		if err := os.WriteFile(path, append(append([]byte(nil), raw...), []byte(suffix)...), 0600); err != nil {
			t.Fatal(err)
		}
		if e, _, err := reopen(childConfig); err == nil {
			e.Shutdown()
			t.Fatal("malformed later row silently returned earlier success")
		}
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	loose := childConfig
	loose.SubmitSchema = json.RawMessage(`{"type":"object"}`)
	row["payload"].(map[string]any)["report"].(map[string]any)["total"] = json.Number("9007199254740992")
	bad, _ := json.Marshal(row)
	if err := os.WriteFile(path, append(bad, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if e, _, err := reopen(loose); err == nil {
		e.Shutdown()
		t.Fatal("looser current schema blessed invalid recorded payload")
	}
}

func TestCh20SubmissionAppendRefusesInvalidFile(t *testing.T) {
	arrived, release := make(chan struct{}, 1), make(chan struct{})
	var once sync.Once
	root, parent := ch20Rig(t, func(w http.ResponseWriter, _ []byte) {
		arrived <- struct{}{}
		<-release
		ch20Reply(w, `{"result":"would be valid"}`, true)
	})
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	ctx := ch20Context(t)
	child, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "append-guard", Prompt: "finish"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-arrived:
	case <-ctx.Done():
		t.Fatal("provider did not start")
	}
	path := filepath.Join(child.Config().DataDir, "submit.md")
	invalid := []byte("{\"version\":2}\n{\"version\":1}\n")
	if err := os.WriteFile(path, invalid, 0600); err != nil {
		t.Fatal(err)
	}
	once.Do(func() { close(release) })
	out, err := root.Wait(ctx, child.ID())
	if err != nil || out.State != "failed" || out.Error == "" || len(out.Submission) != 0 {
		t.Fatalf("invalid existing rows became success: %+v %v", out, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, invalid) {
		t.Fatalf("failed append rewrote invalid evidence: %s %v", after, err)
	}
}
