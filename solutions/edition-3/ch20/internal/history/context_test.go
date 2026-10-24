package history

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"ensemble/internal/common"
)

func TestChapter14TotalProjection(t *testing.T) {
	h := New(nil)
	appendEvent := func(e common.Event) {
		t.Helper()
		if err := h.Append(e); err != nil {
			t.Fatal(err)
		}
	}
	appendEvent(common.Event{Type: common.ResponseEnded}) // known malformed payload: no-op
	appendEvent(common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text("visible")}}})
	appendEvent(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{From: common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}, Parts: common.PartList{{Type: "tool_call", From: &common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}, CallID: "c", Name: "read", Args: json.RawMessage(`{}`)}}}})
	appendEvent(common.Event{Type: common.SkillLoaded, Skill: &common.SkillData{Name: "late", Body: "instructions", ActiveSeq: 3}})
	appendEvent(common.Event{Type: common.MicroHandoff, Handoff: &common.HandoffData{Text: "checkpoint"}})
	if c := h.Context(); len(c.Held) != 1 || c.PendingHandoff == "" || len(c.ActiveSkills) != 0 {
		t.Fatalf("premature survivors: %+v", c)
	}
	appendEvent(common.Event{Type: common.ToolReturned, Tool: &common.ToolData{CallID: "c", Name: "read", Parts: common.PartList{common.Text("large")}}})
	c := h.Context()
	if c.Handoff != "checkpoint" || len(c.Dialogue) != 3 || c.Dialogue[0].Parts[0].Text != "visible" || c.Dialogue[1].Seq != 4 || c.Dialogue[2].Seq != 5 {
		t.Fatalf("survivor identity/order: %+v", c)
	}
	appendEvent(common.Event{Type: common.SkillUnloaded, Skill: &common.SkillData{Name: "late"}})
	appendEvent(common.Event{Type: common.MicroHandoff, Handoff: &common.HandoffData{Text: "new checkpoint"}})
	c = h.Context()
	if !c.ActiveSkills[0].PendingUnload || len(c.Dialogue) != 3 {
		t.Fatalf("unload lost manual: %+v", c)
	}
	var dump bytes.Buffer
	if err := h.Dump(&dump); err != nil {
		t.Fatal(err)
	}
	replay := New(nil)
	if err := replay.Load(&dump); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c, replay.Context()) {
		t.Fatalf("replay mismatch: %+v / %+v", c, replay.Context())
	}
	if len(h.Events()) != 8 {
		t.Fatal("projection rewrote canonical log")
	}
}
func TestChapter14PairedAndSparseRedaction(t *testing.T) {
	c := empty()
	apply := func(e common.Event) { e.Seq = c.AsOf + 1; Apply(&c, e) }
	apply(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{From: common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}, Parts: common.PartList{common.Text("visible"), {Type: "tool_call", From: &common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}, CallID: "c", Name: "read", Args: json.RawMessage(`{}`)}}}})
	apply(common.Event{Type: common.ToolReturned, Tool: &common.ToolData{CallID: "c", Name: "read", Parts: common.PartList{common.Text("large")}}})
	apply(common.Event{Type: common.Redacted, Redact: &common.RedactData{From: 2, To: 2, Parts: []int{0, 99}, Level: common.RedactResult}})
	if c.Dialogue[1].Parts[0].Stub != "" {
		t.Fatal("invalid selector partially changed projection")
	}
	apply(common.Event{Type: common.Redacted, Redact: &common.RedactData{From: 2, To: 2, Parts: []int{0}, Level: common.RedactResult}})
	if p := c.Dialogue[1].Parts[0]; p.CallID != "c" || p.Name != "read" || p.Stub == "" {
		t.Fatalf("lost result wrapper: %+v", p)
	}
	apply(common.Event{Type: common.Redacted, Redact: &common.RedactData{From: 2, To: 2, Level: common.RedactTool}})
	if len(c.Dialogue) != 1 || len(c.Dialogue[0].Parts) != 1 || c.Dialogue[0].Parts[0].Text != "visible" {
		t.Fatalf("orphan/missing visible dialogue: %+v", c)
	}
}

func TestChapter14SnapshotBackupJournal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	h := diskHistory(t, path)
	if _, err := h.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	appendText := func(text string) {
		t.Helper()
		if err := h.Append(common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text(text)}}}); err != nil {
			t.Fatal(err)
		}
	}
	appendText("one")
	if err := h.SaveFile(path, archiveConfig(), common.LogDialogueOnly); err != nil {
		t.Fatal(err)
	}
	appendText("two")
	if err := h.SaveFile(path, archiveConfig(), common.LogDialogueOnly); err != nil {
		t.Fatal(err)
	}
	appendText("three")
	journal, err := readJournal(path + ".journal")
	if err != nil {
		t.Fatal(err)
	}
	if len(journal) != 2 || journal[0].Seq != 2 {
		t.Fatalf("backup tail trimmed: %+v", journal)
	}
	if err := os.WriteFile(path, []byte("broken primary"), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path+".journal", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(`{"seq":4`)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	recovered := diskHistory(t, path)
	r, err := recovered.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Backup || recovered.LastSeq() != 3 || len(recovered.Context().Dialogue) != 3 {
		t.Fatalf("fallback+tail: %+v %+v", r, recovered.Context())
	}
	if err = recovered.Append(common.Event{Type: common.TurnEnded}); err != nil {
		t.Fatal(err)
	}
	again := diskHistory(t, path)
	if _, err = again.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if again.LastSeq() != 4 {
		t.Fatal("append after torn-tail repair lost")
	}
	// A missing suffix must not publish a partially replayed recovery candidate.
	if err = os.WriteFile(path+".journal", []byte(`{"seq":3,"at":"0001-01-01T00:00:00Z","type":"turn_ended"}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	prior := again.Context()
	if _, err = again.LoadFile(path); err == nil {
		t.Fatal("journal gap accepted")
	}
	if !reflect.DeepEqual(prior, again.Context()) {
		t.Fatal("failed candidate partially published")
	}
}

type diagnosticParent struct {
	common.Agent
	logger *log.Logger
}

func (p diagnosticParent) Sandbox() common.Sandbox { return p.Agent.Sandbox() }

func (p diagnosticParent) Parent() common.Ensemble { return diagnosticEnsemble{logger: p.logger} }

type diagnosticEnsemble struct {
	common.Ensemble
	logger *log.Logger
}

func (p diagnosticEnsemble) Logger() *log.Logger { return p.logger }
func TestChapter14RecoveredSemanticDiagnostics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.save")
	h := diskHistory(t, path)
	if err := h.Append(common.Event{Type: common.MessageReceived}); err != nil {
		t.Fatal(err)
	}
	if err := h.SaveFile(path, archiveConfig(), common.LogFull); err != nil {
		t.Fatal(err)
	}
	if err := h.Append(common.Event{Type: common.MessageReceived}); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	restored := New(diagnosticParent{Agent: diskHistoryOwner(t, path), logger: log.New(&output, "", 0)})
	if _, err := restored.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), "history: skip message_received") != 2 {
		t.Fatalf("missing recovery diagnostics: %s", output.String())
	}
	if restored.LastSeq() != 2 || len(restored.Context().Dialogue) != 0 || len(restored.Events()) != 2 {
		t.Fatal("diagnostic changed canonical replay")
	}
}

func TestChapter14SnapshotRejectsMissingTailEvent(t *testing.T) {
	h := New(nil)
	appendText(t, h, "one")
	c := h.Context()
	saved := common.SaveFile{Config: archiveConfig(), AsOf: 1, Context: &c, Log: []common.Event{{Seq: 3, Time: time.Now().UTC(), Type: common.MessageReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text("three")}}}}}
	data, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = decodeSave(data); err == nil {
		t.Fatal("snapshot tail skipped an unrecoverable event")
	}
}

func TestChapter14LiveAdmissionProvenanceReplays(t *testing.T) {
	p := common.Provenance{Vendor: common.Anthropic, Model: "offline", Surface: common.Messages}
	response := common.Event{Seq: 1, TurnID: 1, Type: common.ResponseEnded, Response: &common.ResponseData{From: p, Parts: common.PartList{{Type: "tool_call", From: &p, CallID: "load", Name: "load_skill", Args: json.RawMessage(`{}`)}}}}
	admission := common.Event{Seq: 2, TurnID: 1, Type: common.ToolCalled, Tool: &common.ToolData{CallID: "load", Name: "load_skill", Args: json.RawMessage(`{}`)}}
	c := empty()
	Apply(&c, response)
	Apply(&c, admission)
	if c.Dialogue[0].Parts[0].AdmissionSeq != 2 || response.Response.Parts[0].AdmissionSeq != 0 {
		t.Fatal("projection admission missing or canonical response rewritten")
	}
	path := filepath.Join(t.TempDir(), "history.save")
	h := diskHistory(t, path)
	for _, event := range []common.Event{response, admission, {Type: common.ToolReturned, Tool: &common.ToolData{CallID: "load", Name: "load_skill", Parts: common.PartList{common.Text("loaded")}}}, {Type: common.RequestSent, Request: &common.RequestData{To: p}}, {Type: common.RequestSent, Request: &common.RequestData{To: p}}} {
		if err := h.Append(event); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.SaveFile(path, archiveConfig(), common.LogFull); err != nil {
		t.Fatal(err)
	}
	restored := diskHistory(t, path)
	if _, err := restored.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if got := restored.Context(); got.Dialogue[0].Parts[0].AdmissionSeq != 2 || got.PreviousSentThrough != 3 || got.SentThrough != 4 {
		t.Fatalf("lost live provenance/decision boundary: %+v", got)
	}
	if restored.Events()[0].Response.Parts[0].AdmissionSeq != 0 {
		t.Fatal("canonical response was mutated")
	}
}
