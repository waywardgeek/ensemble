package eventlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/jsonvalue"
	"example.com/ensemble/internal/persistence"
	"strings"
	"testing"
)

type owner struct{ writes int }

func (o *owner) Ensemble() common.Ensemble    { return o }
func (o *owner) Config() common.Config        { return common.Config{DisableStreaming: true} }
func (o *owner) Workspace() string            { return "" }
func (o *owner) Logf(string, ...any)          {}
func (o *owner) Publish(string, common.Event) {}
func (*owner) AllocateHandle() uint64         { return 0 }
func (o *owner) Write(b []byte) (int, error) {
	o.writes++
	return len(b) / 2, errors.New("simulated partial write")
}
func (o *owner) Close() error { return nil }
func TestPartialWritePermanentlyFaultsLog(t *testing.T) {
	parent := &owner{}
	log := &Log{parent: parent, writer: parent}
	if err := log.Append(common.Event{}); err == nil {
		t.Fatal("partial write accepted")
	}
	if err := log.Append(common.Event{}); err == nil || parent.writes != 1 {
		t.Fatal("appended after partial record")
	}
}

func (*owner) Observe(common.Observation)                       {}
func (*owner) Collect([]common.RequestHandle) common.Collection { return nil }

func (owner) ModelReady(common.ModelOperation) {}

func (owner) ClaimSettingsPath(path string) (string, error) { return path, nil }
func (owner) ReleaseSettingsPath(string)                    {}

// Test fixture is a composition root for the owner interface.
func (a *owner) Codec() common.SessionCodec { return persistence.NewCodec(a) }

func (*owner) ReleaseSession(string) {}

func TestPreparedSessionRecordExactAcceptedRaw(t *testing.T) {
	o := &owner{}
	l := &Log{parent: o, session: true}
	from := common.Provenance{Vendor: "anthropic", Model: "fixture", Surface: "messages"}
	e := common.Event{Seq: 1, Type: "response_ended", Time: "2026-10-08T00:00:00Z", Response: &common.Response{From: from, Parts: []common.Part{{Type: "tool_call", CallID: "c", Name: "local", From: &from, Args: json.RawMessage("{\n \"z\": 1.0, \"a\": 9007199254740993 }"), Opaque: json.RawMessage("{ \"signature\": \"x\\n y\", \"n\": 1e3 }")}}, Usage: &common.Usage{}, RawUsage: json.RawMessage("{\r\n \"tokens\": 1.0 }")}}
	p, err := l.Prepare(e)
	if err != nil {
		t.Fatal(err)
	}
	accepted := p.Event()
	raw := p.(*preparedEvent).bytes
	if bytes.Count(raw, []byte{'\n'}) != 1 {
		t.Fatal("prepared record has physical embedded LF")
	}
	if string(accepted.Response.RawUsage) != `{"tokens":1.0}` || string(accepted.Response.Parts[0].Args) != `{"z":1.0,"a":9007199254740993}` || string(accepted.Response.Parts[0].Opaque) != `{"signature":"x\n y","n":1e3}` {
		t.Fatal("preparation changed number/order/signature semantics")
	}
	if string(e.Response.RawUsage) != "{\r\n \"tokens\": 1.0 }" {
		t.Fatal("preparation mutated caller")
	}
	e.Response.RawUsage = json.RawMessage(`{"invalid":"\ud800"}`)
	if _, err = l.Prepare(e); err == nil {
		t.Fatal("unpaired surrogate accepted")
	}
}

func TestPreparedEmptyToolResultPreservesPresence(t *testing.T) {
	o := &owner{}
	l := &Log{parent: o, session: true}
	p, err := l.Prepare(common.Event{Seq: 2, Type: "tool_returned", Time: "2026-10-08T00:00:00Z", Tool: &common.ToolEvent{CallID: "c", Parts: []common.Part{}}})
	if err != nil {
		t.Fatal(err)
	}
	if p.Event().Tool.Parts == nil {
		t.Fatal("empty result became absent at acceptance")
	}
}

func TestPreparedSessionWriteExactRecordBound(t *testing.T) {
	o := &owner{}
	l := &Log{parent: o, session: true}
	empty := ""
	e := common.Event{Seq: 1, Type: "message_received", Time: "2026-10-08T00:00:00Z", Message: &common.Entry{Actor: "system", Purpose: "instruction", Parts: []common.Part{{Type: "text", Text: &empty}}}}
	p, err := l.Prepare(e)
	if err != nil {
		t.Fatal(err)
	}
	overhead := len(p.(*preparedEvent).bytes)
	body := strings.Repeat("x", common.SkillRecordLimit-overhead)
	e.Message.Parts[0].Text = &body
	p, err = l.Prepare(e)
	if err != nil {
		t.Fatal("exact64MiB refused", err)
	}
	if len(p.(*preparedEvent).bytes) != common.SkillRecordLimit {
		t.Fatal("wrong physical size")
	}
	body += "x"
	if _, err = l.Prepare(e); err == nil {
		t.Fatal("one-over accepted")
	}
}

func TestPreparedSessionLogAndCountAdmission(t *testing.T) {
	for _, kind := range []string{"bytes", "count"} {
		t.Run(kind, func(t *testing.T) {
			o := &owner{}
			w := &retainedWriter{}
			l := &Log{parent: o, session: true, writer: w}
			e := common.Event{Seq: 1, Type: "error_occurred", Time: "2026-10-08T00:00:00Z", Error: &common.EventError{Code: "fixture", Message: "fixture"}}
			p, err := l.Prepare(e)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "bytes" {
				l.bytes = (1 << 30) - int64(len(p.(*preparedEvent).bytes))
			} else {
				l.count = 999999
			}
			if err = l.AppendPrepared(p); err != nil {
				t.Fatal("exact bound", err)
			}
			before := w.Len()
			if err = l.AppendPrepared(p); err == nil {
				t.Fatal("one-over admitted")
			}
			if w.Len() != before || !l.faulted {
				t.Fatal("overflow wrote or failed to fault")
			}
		})
	}
}

type retainedWriter struct{ bytes.Buffer }

func (*retainedWriter) Close() error { return nil }

func (r *owner) JSON() common.JSONService      { return jsonvalue.New(r) }
func (r *owner) MCP() common.MCPService        { return nil }
func (r *owner) MCPRuntime() common.MCPRuntime { return nil }
