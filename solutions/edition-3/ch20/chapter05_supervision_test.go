package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"ensemble"
)

func TestChapter5SupervisionLogSchema(t *testing.T) {
	app, a, _ := c4Agent(t)
	type expectation struct {
		name, args      string
		control, failed bool
	}
	expected := map[string]expectation{}
	invoke := func(id, name, args string, control, failed bool) {
		t.Helper()
		expected[id] = expectation{name, args, control, failed}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := a.Tools().Invoke(ctx, ensemble.Part{Type: "tool_call", CallID: id, Name: name, Args: json.RawMessage(args)}, ""); err != nil {
			t.Fatal(err)
		}
	}
	invoke("limits", "tool_limits", `{"ai_callback_delay":0,"max_output_bytes":1024}`, true, false)
	command := `stty -echo; printf 'armed\n'; IFS= read -r line; printf 'got:%s\n' "$line"; IFS= read -r ignored`
	args, err := json.Marshal(map[string]any{"command": command})
	if err != nil {
		t.Fatal(err)
	}
	invoke("ordinary-process", "run_command", string(args), false, false)
	job, err := a.Jobs().Get(1)
	if err != nil {
		t.Fatal(err)
	}
	ch5Eventually(t, "PTY reached its input read", func() bool { return job.Data().Bytes >= 6 })
	invoke("observe", "wait_for_job", `{"handle":1,"ai_callback_delay":0}`, true, false)
	invoke("input", "send_input", `{"handle":1,"input":"hello","ai_callback_pattern":"got:hello","ai_callback_delay":2}`, true, false)
	invoke("cursor", "wait_for_job", `{"handle":1,"ai_callback_delay":0}`, true, false)
	invoke("kill", "kill_job", `{"handle":1}`, true, false)
	invoke("bad-handle", "wait_for_job", `{"handle":999,"ai_callback_delay":0}`, true, true)
	if len(app.JobTable()) != 1 {
		t.Fatalf("supervision allocated a new job: count=%d", len(app.JobTable()))
	}
	c4Register(t, a, "ordinary_failure", func(_ context.Context, call ensemble.ToolCall, _ json.RawMessage) error {
		if _, err := io.WriteString(call, "partial output"); err != nil {
			return err
		}
		return errors.New("real ordinary failure")
	})
	invoke("ordinary-error", "ordinary_failure", `{}`, false, true)
	invoke("failed-target", "wait_for_job", `{"handle":2,"ai_callback_delay":0}`, true, true)
	if err := app.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if len(app.JobTable()) != 2 {
		t.Fatalf("expected only two ordinary jobs; got %d", len(app.JobTable()))
	}
	var encoded bytes.Buffer
	if err := a.Log().Dump(&encoded); err != nil {
		t.Fatal(err)
	}
	calls, returns := map[string]int{}, map[string]int{}
	reports := map[string]string{}
	killed := 0
	dec := json.NewDecoder(&encoded)
	for {
		var fields map[string]json.RawMessage
		if err := dec.Decode(&fields); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		var e ensemble.Event
		if err := json.Unmarshal(raw, &e); err != nil {
			t.Fatal(err)
		}
		if e.Type == ensemble.JobKilled && e.Job != nil && e.Job.Handle == 1 {
			killed++
			if e.Job.Reason != "kill_job" || e.Job.Status != ensemble.JobKilledStatus || e.Job.ExitCode != nil || e.Job.CallID != "ordinary-process" {
				t.Fatalf("lost real killed metadata: %+v", e.Job)
			}
		}
		if e.Type != ensemble.ToolCalled && e.Type != ensemble.ToolReturned {
			continue
		}
		if e.Tool == nil {
			t.Fatal("missing serialized tool data")
		}
		want, ok := expected[e.Tool.CallID]
		if !ok {
			t.Fatalf("unexpected tool call id %q", e.Tool.CallID)
		}
		if e.Tool.Name != want.name || string(e.Tool.Args) != want.args {
			t.Fatalf("lost control/ordinary call correlation: %+v", e.Tool)
		}
		var toolFields map[string]json.RawMessage
		if err := json.Unmarshal(fields["tool"], &toolFields); err != nil {
			t.Fatal(err)
		}
		_, ownsJob := toolFields["job"]
		if want.control {
			if ownsJob {
				t.Errorf("%s %s serialized tool.job on supervision; target belongs in args/report", e.Type, e.Tool.Name)
			}
		} else {
			d := e.Tool.Job
			if !ownsJob || d == nil || d.Handle < 1 || d.Output.Locator == "" || d.Name != e.Tool.Name || d.CallID != e.Tool.CallID || string(d.Args) != want.args {
				t.Fatalf("ordinary job metadata missing/uncorrelated: %+v", e.Tool)
			}
		}
		if e.Type == ensemble.ToolCalled {
			calls[e.Tool.CallID]++
			continue
		}
		returns[e.Tool.CallID]++
		if e.Tool.IsError != want.failed {
			t.Errorf("%s error=%v, want %v", e.Tool.CallID, e.Tool.IsError, want.failed)
		}
		for _, p := range e.Tool.Parts {
			reports[e.Tool.CallID] += p.Text
		}
	}
	for id := range expected {
		if calls[id] != 1 || returns[id] != 1 {
			t.Errorf("%s called=%d returned=%d, want one exact pair", id, calls[id], returns[id])
		}
	}
	for _, id := range []string{"observe", "input", "cursor"} {
		if !strings.Contains(reports[id], "job 1 running;") || !strings.Contains(reports[id], "output: "+job.Data().Output.Locator) {
			t.Errorf("%s lost actual target report: %s", id, reports[id])
		}
	}
	if !strings.Contains(reports["input"], "got:hello") {
		t.Errorf("send_input lost real output: %s", reports["input"])
	}
	if strings.Contains(reports["cursor"], "got:hello") {
		t.Errorf("observation replayed consumed output: %s", reports["cursor"])
	}
	if !strings.Contains(reports["cursor"], fmt.Sprintf("%d bytes total", job.Data().Bytes)) {
		t.Errorf("cursor report lost retained-byte count: %s", reports["cursor"])
	}
	if !strings.Contains(reports["kill"], "job 1 killed;") || !strings.Contains(reports["kill"], "reason: kill_job") {
		t.Errorf("kill lost target status: %s", reports["kill"])
	}
	if !strings.Contains(reports["bad-handle"], "unknown job handle 999") {
		t.Errorf("lost actual control error: %s", reports["bad-handle"])
	}
	if !strings.Contains(reports["ordinary-error"], "real ordinary failure") || !strings.Contains(reports["failed-target"], "job 2 done;") {
		t.Errorf("lost ordinary failure or observed target: %q / %q", reports["ordinary-error"], reports["failed-target"])
	}
	if killed != 1 {
		t.Errorf("serialized killed facts=%d, want exactly one", killed)
	}
}
