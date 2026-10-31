package ch10_test

// These controls use public operations and actual local HTTP. The JSON adapter
// below follows persistence-format.md; no private runtime import is used.
import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"example.com/ensemble"
)

func callReply(t *testing.T, x exchange, id, name string, args json.RawMessage) {
	t.Helper()
	// Arguments are a provider string, so even malformed content reaches the
	// provider parser without this fixture silently repairing it.
	data, err := json.Marshal(map[string]any{"model": "gpt-4.1-mini-2025-04-14", "choices": []any{map[string]any{"index": 0, "message": map[string]any{"role": "assistant", "content": "", "tool_calls": []any{map[string]any{"id": id, "type": "function", "function": map[string]any{"name": name, "arguments": string(args)}}}}, "finish_reason": "tool_calls"}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1}})
	if err != nil {
		t.Fatal(err)
	}
	x.reply <- data
}
func attempted(t *testing.T, a *ensemble.Agent, e *endpoint, id, name string, args json.RawMessage) {
	t.Helper()
	h, err := a.Submit("local controlled attempt " + id)
	if err != nil {
		t.Fatal(err)
	}
	callReply(t, nextRequest(t, e), id, name, args)
	answer(t, "openai", nextRequest(t, e))
	finish(t, h)
}
func pendingLimits(t *testing.T, a *ensemble.Agent) json.RawMessage {
	t.Helper()
	x := export(t, a)
	var value struct {
		State struct {
			Limits json.RawMessage `json:"limits"`
		} `json:"state"`
	}
	if err := json.Unmarshal(x.Bytes, &value); err != nil {
		t.Fatal(err)
	}
	return value.State.Limits
}
func expectLimits(t *testing.T, a *ensemble.Agent, expected string) {
	t.Helper()
	var got, want any
	if err := json.Unmarshal(pendingLimits(t, a), &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pending limit projection: got %v want %v", got, want)
	}
}
func limitEvents(t *testing.T, a *ensemble.Agent, after uint64) []map[string]json.RawMessage {
	t.Helper()
	var result []map[string]json.RawMessage
	for _, event := range a.Events() {
		if event.Seq <= after {
			continue
		}
		raw, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		var value map[string]json.RawMessage
		if err = json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		result = append(result, value)
	}
	return result
}
func stringField(t *testing.T, value map[string]json.RawMessage, key string) string {
	t.Helper()
	var s string
	if err := json.Unmarshal(value[key], &s); err != nil {
		t.Fatal(err)
	}
	return s
}
func seqField(t *testing.T, value map[string]json.RawMessage) uint64 {
	t.Helper()
	var n uint64
	if err := json.Unmarshal(value["seq"], &n); err != nil {
		t.Fatal(err)
	}
	return n
}
func consumption(t *testing.T, a *ensemble.Agent, after uint64, id, name string) {
	t.Helper()
	var consumed, called uint64
	count := 0
	for _, event := range limitEvents(t, a, after) {
		kind := stringField(t, event, "type")
		if kind == "tool_limits_consumed" {
			var fact struct {
				CallID, Name string
				Overrides    map[string]any
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(event["limits"], &raw); err != nil {
				t.Fatal(err)
			}
			fact.CallID = stringField(t, raw, "call_id")
			fact.Name = stringField(t, raw, "name")
			if err := json.Unmarshal(raw["overrides"], &fact.Overrides); err != nil {
				t.Fatal(err)
			}
			if fact.CallID != id || fact.Name != name || !reflect.DeepEqual(fact.Overrides, map[string]any{"max_output_bytes": float64(80), "ai_callback_pattern": ""}) {
				t.Fatalf("consumption copied different setting: %+v", fact)
			}
			count++
			consumed = seqField(t, event)
		}
		if kind == "tool_called" {
			var tool map[string]json.RawMessage
			if err := json.Unmarshal(event["tool"], &tool); err != nil {
				t.Fatal(err)
			}
			if stringField(t, tool, "call_id") == id {
				called = seqField(t, event)
			}
		}
	}
	if count != 1 || consumed == 0 || called <= consumed {
		t.Fatalf("next attempt did not consume once before dispatch: count=%d consumed=%d called=%d", count, consumed, called)
	}
}

func TestCh10PublicLimitsRestartAndLiteralAttempt(t *testing.T) {
	e := localEndpoint(t)
	o := wireOptions(t, "openai", e)
	o.Config.Builtins = []string{"tool_limits", "wait_for_job"}
	a := opened(t, application(t), o)
	attempted(t, a, e, "setter", "tool_limits", json.RawMessage(`{"max_output_bytes":80,"ai_callback_pattern":""}`))
	expectLimits(t, a, `{"max_output_bytes":80,"ai_callback_pattern":""}`)
	seed := export(t, a)
	if _, err := a.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	expectLimits(t, a, `{"max_output_bytes":80,"ai_callback_pattern":""}`)
	closed(t, a)
	for _, mode := range []string{"checkpoint", "null-rebuild", "snapshot-import"} {
		for _, attempt := range []struct{ name, tool, args, after string }{
			{"unknown", "not_installed", `{}`, `null`}, {"disabled", "read_file", `{"path":"unused"}`, `null`},
			{"invalid-setter", "tool_limits", `{"max_output_bytes":0}`, `null`}, {"second-setter", "tool_limits", `{"max_output_bytes":128}`, `{"max_output_bytes":128}`},
			{"stale-supervision", "wait_for_job", `{"handle":999999}`, `null`},
		} {
			t.Run(mode+"/"+attempt.name, func(t *testing.T) {
				target := o
				target.Config.DataDir = filepath.Join(o.Config.Workspace, mode+"-"+attempt.name)
				app := application(t)
				var b *ensemble.Agent
				var err error
				if mode == "snapshot-import" {
					b, err = app.ImportSession(seed.Bytes, target)
				} else {
					copies(t, o.Config.DataDir, target.Config.DataDir)
					if mode == "null-rebuild" {
						var outer map[string]json.RawMessage
						if err = json.Unmarshal(seed.Bytes, &outer); err != nil {
							t.Fatal(err)
						}
						outer["state"] = json.RawMessage("null")
						outer["state_sha256"] = json.RawMessage("null")
						raw, _ := json.Marshal(outer)
						if err = os.WriteFile(filepath.Join(target.Config.DataDir, "checkpoint.json"), raw, 0600); err != nil {
							t.Fatal(err)
						}
					}
					b, err = app.OpenSession(target)
				}
				if err != nil {
					t.Fatal(err)
				}
				defer b.Close()
				noStartupHTTP(t, e)
				expectLimits(t, b, `{"max_output_bytes":80,"ai_callback_pattern":""}`)
				boundary := export(t, b).AsOf
				attempted(t, b, e, "next", attempt.tool, json.RawMessage(attempt.args))
				consumption(t, b, boundary, "next", attempt.tool)
				expectLimits(t, b, attempt.after)
			})
		}
	}
}

func TestCh10PublicLimitsMalformedAndPaused(t *testing.T) {
	e := localEndpoint(t)
	o := wireOptions(t, "openai", e)
	o.Config.Builtins = []string{"tool_limits"}
	a := opened(t, application(t), o)
	attempted(t, a, e, "setter", "tool_limits", json.RawMessage(`{"max_output_bytes":80,"ai_callback_pattern":""}`))
	boundary := export(t, a).AsOf
	h, err := a.Submit("malformed response control")
	if err != nil {
		t.Fatal(err)
	}
	callReply(t, nextRequest(t, e), "bad", "tool_limits", json.RawMessage(`[`))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := h.Wait(ctx)
	if err != nil || result.Outcome != "error" {
		t.Fatalf("malformed response not refused: %+v %v", result, err)
	}
	expectLimits(t, a, `{"max_output_bytes":80,"ai_callback_pattern":""}`)
	for _, ev := range limitEvents(t, a, boundary) {
		if stringField(t, ev, "type") == "tool_limits_consumed" {
			t.Fatal("unaccepted malformed call consumed pending setting")
		}
	}
	pause, err := a.RegisterPause()
	if err != nil {
		t.Fatal(err)
	}
	defer pause.Close()
	if _, err = pause.Update(true, false); err != nil {
		t.Fatal(err)
	}
	_, watch, err := a.Watch()
	if err != nil {
		t.Fatal(err)
	}
	defer watch.Close()
	boundary = export(t, a).AsOf
	h, err = a.Submit("paused accepted call")
	if err != nil {
		t.Fatal(err)
	}
	callReply(t, nextRequest(t, e), "paused", "not_installed", json.RawMessage(`{}`))
	for {
		record, err := watch.Next(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if record.Observation.Event.Type == "response_ended" && record.Observation.Event.Seq > boundary {
			break
		}
	}
	for _, ev := range limitEvents(t, a, boundary) {
		kind := stringField(t, ev, "type")
		if kind == "tool_limits_consumed" || kind == "tool_called" {
			t.Fatal("pause admitted attempt or consumed setting")
		}
	}
	_, err = a.Checkpoint()
	code(t, err, "session_busy")
	if _, err = pause.Update(false, false); err != nil {
		t.Fatal(err)
	}
	answer(t, "openai", nextRequest(t, e))
	finish(t, h)
	consumption(t, a, boundary, "paused", "not_installed")
	expectLimits(t, a, `null`)
}

func TestCh10PublicCheckpointPermissionFailureAndRecovery(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Fatal("permission fault requires an unprivileged test process; root is not a passing control")
	}
	o := options(t, "permission-fault")
	a := opened(t, application(t), o)
	first, err := a.Checkpoint()
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(o.Config.DataDir, "checkpoint.json")
	before, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(o.Config.DataDir, 0500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(o.Config.DataDir, 0700)
	_, err = a.Checkpoint()
	code(t, err, "session_io")
	after, err := os.ReadFile(file)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("pre-create failure damaged prior checkpoint")
	}
	current := state(t, a)
	if current.CheckpointSeq == nil || *current.CheckpointSeq != first.AsOf {
		t.Fatal("failed save falsely published new anchor")
	}
	if err = os.Chmod(o.Config.DataDir, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Checkpoint(); err != nil {
		t.Fatalf("healthy Agent failed deliberate retry: %v", err)
	}
	closed(t, a)
	closed(t, a)
}

func TestCh10PublicLimitOrderingWithoutWindowWitness(t *testing.T) {
	e := localEndpoint(t)
	o := wireOptions(t, "openai", e)
	o.Config.Builtins = []string{"tool_limits"}
	app := application(t)
	a := opened(t, app, o)
	if err := a.Ephemeral("consumed guidance witness"); err != nil {
		t.Fatal(err)
	}
	attempted(t, a, e, "setter", "tool_limits", json.RawMessage(`{"max_output_bytes":80,"ai_callback_pattern":""}`))
	attempted(t, a, e, "next", "not_installed", json.RawMessage(`{}`))
	for i := 0; i < 101; i++ {
		if err := a.Ephemeral("retained later renderable guidance"); err != nil {
			t.Fatal(err)
		}
	}
	x := export(t, a)
	if _, err := app.InspectCheckpoint(x.Bytes); err != nil {
		t.Fatalf("genuine settled old-call positive refused: %v", err)
	}
	cmd := exec.Command("python3", os.Getenv("CH10_LIMIT_CASES"))
	cmd.Stdin = bytes.NewReader(x.Bytes)
	var output, diagnostic bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &diagnostic
	if err := cmd.Run(); err != nil {
		t.Fatalf("limit mutation preparation: %v: %s", err, diagnostic.String())
	}
	var generated struct {
		Cases []struct {
			Name         string `json:"name"`
			Bytes        []byte `json:"bytes"`
			ExpectedCode string `json:"expected_code"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(output.Bytes(), &generated); err != nil || len(generated.Cases) != 11 {
		t.Fatalf("missing prepared limit controls: %v", err)
	}
	for _, c := range generated.Cases {
		t.Run(c.Name, func(t *testing.T) { _, err := app.InspectCheckpoint(c.Bytes); code(t, err, c.ExpectedCode) })
	}
	if _, err := app.InspectCheckpoint(x.Bytes); err != nil {
		t.Fatal("refused mutations altered original positive")
	}
	noStartupHTTP(t, e)
}

func TestCh10PublicCloseSaveFailureKeepsErrorAndReleasesLock(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Fatal("permission fault requires an unprivileged test process")
	}
	o := options(t, "close-fault")
	app := application(t)
	a := opened(t, app, o)
	if _, err := a.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(o.Config.DataDir, "checkpoint.json")
	before, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Ephemeral("accepted tail before failed final checkpoint"); err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(o.Config.DataDir, 0500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(o.Config.DataDir, 0700)
	first := a.Close()
	code(t, first, "session_io")
	after, err := os.ReadFile(file)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed final checkpoint changed prior committed bytes")
	}
	if err = os.Chmod(o.Config.DataDir, 0700); err != nil {
		t.Fatal(err)
	}
	second := a.Close()
	code(t, second, "session_io")
	if first.Error() != second.Error() {
		t.Fatal("repeated close lost original failure")
	}
	b, err := app.OpenSession(o)
	if err != nil {
		t.Fatalf("failed-save close retained lock/reservation: %v", err)
	}
	defer b.Close()
	if got := b.Snapshot(); len(got.Ephemera) != 1 {
		t.Fatal("failed final checkpoint lost accepted log tail")
	}
}
