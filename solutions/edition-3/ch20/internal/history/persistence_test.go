package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ensemble/internal/common"
)

func archiveConfig() common.SavedConfig {
	return common.SavedConfig{Vendor: common.Anthropic, Surface: common.Messages, Model: "fixture", Tools: []common.ToolDeclaration{}}
}
func appendText(t *testing.T, h common.History, text string) {
	t.Helper()
	if err := h.Append(common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text(text)}}}); err != nil {
		t.Fatal(err)
	}
}
func TestArchiveBackupRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	h := diskHistory(t, path)
	appendText(t, h, "one")
	if err := h.SaveFile(path, archiveConfig(), common.LogFull); err != nil {
		t.Fatal(err)
	}
	appendText(t, h, "two")
	if err := h.SaveFile(path, archiveConfig(), common.LogFull); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	recovered := diskHistory(t, path)
	// Chapter 14 retains and replays the tail needed by the validated backup.
	r, err := recovered.LoadFile(path)
	if err != nil || !r.Backup || r.AsOf != 2 || len(recovered.Context().Dialogue) != 2 {
		t.Fatalf("%+v %v", r, err)
	}
	// Saving from fallback must not rotate corrupt primary over valid backup.
	appendText(t, recovered, "replacement")
	if err := recovered.SaveFile(path, archiveConfig(), common.LogFull); err != nil {
		t.Fatal(err)
	}
	backup, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = decodeSave(backup)
	if err != nil {
		t.Fatal(err)
	}
}
func TestArchiveReplayTailAndNullContext(t *testing.T) {
	h := New(nil)
	appendText(t, h, "one")
	snap := h.Context()
	log := h.Events()
	log = append(log, common.Event{Seq: 2, Time: time.Now().UTC(), Type: common.MessageReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text("two")}}})
	s := common.SaveFile{Config: archiveConfig(), AsOf: 1, Context: &snap, Log: log}
	for _, noContext := range []bool{false, true} {
		if noContext {
			s.Context = nil
		}
		raw, _ := json.Marshal(s)
		_, got, err := decodeSave(raw)
		if err != nil || len(got.Dialogue) != 2 {
			t.Fatalf("null=%t: %+v %v", noContext, got, err)
		}
	}
	s.Context = nil
	s.Log = s.Log[1:]
	raw, _ := json.Marshal(s)
	if _, _, err := decodeSave(raw); err == nil {
		t.Fatal("partial log rebuilt missing context")
	}
}
func TestDialogueArchiveRetainsRecoveryContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	h := diskHistory(t, path)
	appendText(t, h, "visible")
	if err := h.SaveFile(path, archiveConfig(), common.LogDialogueOnly); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var s common.SaveFile
	if err = json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	if len(s.Log) != 0 || len(s.DialogueLog) != 1 || s.DialogueLog[0].Role != "user" || s.Context == nil {
		t.Fatalf("%s", raw)
	}
	restored := diskHistory(t, path)
	if _, err = restored.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	appendText(t, restored, "next")
	if restored.LastSeq() != 2 {
		t.Fatal(restored.LastSeq())
	}
	if err = restored.SaveFile(path, archiveConfig(), common.LogDialogueOnly); err != nil {
		t.Fatal(err)
	}
}
func TestBadArchiveDoesNotInstallOrOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	h := diskHistory(t, path)
	appendText(t, h, "kept")
	if err := os.WriteFile(path, []byte("{bad"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := h.LoadFile(path); err == nil {
		t.Fatal("accepted invalid archive")
	}
	if len(h.Context().Dialogue) != 1 {
		t.Fatal("mutated history")
	}
	if err := h.SaveFile(path, archiveConfig(), common.LogFull); err == nil {
		t.Fatal("overwrote invalid archive without recovery candidate")
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "{bad" {
		t.Fatal(string(raw))
	}
	if err := h.Append(common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{{Type: "opaque", From: &common.Provenance{Vendor: common.Anthropic, Surface: common.Messages, Model: "fixture"}, Data: json.RawMessage("bad")}}}}); err == nil {
		if err = h.SaveFile(filepath.Join(filepath.Dir(path), "malformed.json"), archiveConfig(), common.LogFull); err == nil {
			t.Fatal("accepted malformed raw JSON")
		}
	}
	if !strings.Contains(string(raw), "bad") {
		t.Fatal("lost corruption evidence")
	}
}

func TestCh10R1PublicEnvelopeAndHints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	h := diskHistory(t, path)
	appendText(t, h, "prompt")
	for _, actor := range []common.Actor{common.Human, common.System} {
		if err := h.Append(common.Event{Type: common.HintReceived, Message: &common.MessageData{Actor: actor, Parts: common.PartList{common.Text(string(actor) + " hint")}}}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := h.SaveFile(path, archiveConfig(), common.LogDialogueOnly); err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err = json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		for k := range fields {
			switch k {
			case "config", "as_of", "context", "log", "log_mode", "dialogue_log":
			default:
				t.Errorf("unsolicited envelope field %q", k)
			}
		}
		var s common.SaveFile
		if err = json.Unmarshal(raw, &s); err != nil {
			t.Fatal(err)
		}
		if len(s.DialogueLog) != 2 || s.DialogueLog[1].SourceSeq != 2 || s.DialogueLog[1].Role != "user" || s.DialogueLog[1].Text != "human hint" {
			t.Errorf("visible hint archive: %+v", s.DialogueLog)
		}
		restored := diskHistory(t, path)
		if _, err = restored.LoadFile(path); err != nil {
			t.Fatal(err)
		}
		h = restored
	}
}

func TestCh10R2DirectToolArchiveValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	h := diskHistory(t, path)
	if err := h.Append(common.Event{Type: common.ToolCalled, Tool: &common.ToolData{Name: "read_file", Args: json.RawMessage(`{"path":"input"}`)}}); err != nil {
		t.Fatal(err)
	}
	if err := h.SaveFile(path, archiveConfig(), common.LogFull); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		corrupt func(*common.SaveFile)
	}{
		{"model-turn-requires-call-id", func(s *common.SaveFile) { s.Log[0].TurnID = 1 }},
		{"argument-object", func(s *common.SaveFile) { s.Log[0].Tool.Args = json.RawMessage(`[]`) }},
		{"typed-payload", func(s *common.SaveFile) { s.Log[0].Type = common.MessageReceived }},
		{"sequence-anchor", func(s *common.SaveFile) { s.Log[0].Seq = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var s common.SaveFile
			if err := json.Unmarshal(original, &s); err != nil {
				t.Fatal(err)
			}
			tc.corrupt(&s)
			raw, err := json.Marshal(s)
			if err != nil {
				t.Fatal(err)
			}
			bad := filepath.Join(t.TempDir(), "save.json")
			if err = os.WriteFile(bad, raw, 0600); err != nil {
				t.Fatal(err)
			}
			next := diskHistory(t, bad)
			_, err = next.LoadFile(bad)
			if tc.name == "sequence-anchor" {
				if err == nil || next.LastSeq() != 0 || len(next.Events()) != 0 {
					t.Fatal("invalid envelope installed")
				}
			} else {
				// Chapter 14 keeps recognized semantic junk as a canonical no-op.
				if err != nil {
					t.Fatal(err)
				}
				if next.LastSeq() != 1 || len(next.Events()) != 1 || len(next.Context().Dialogue) != 0 {
					t.Fatal("semantic no-op changed projection or lost canonical fact")
				}
			}
		})
	}
}

func TestCh10R3SemanticContextJSON(t *testing.T) {
	a := common.Context{Dialogue: []common.Entry{{Seq: 1, Actor: common.Assistant, Parts: common.PartList{{Type: "tool_call", CallID: "native", Name: "tool", Args: json.RawMessage(`{"nested":{"n":9007199254740992,"s":"x"},"list":[1,2]}`)}}}}, Ephemera: common.PartList{}, Pending: []string{}}
	b := clone(a)
	b.Ephemera = nil
	b.Pending = nil
	b.Dialogue[0].Parts[0].Args = json.RawMessage("{\n \"list\": [1, 2],\n \"nested\": {\"s\": \"x\", \"n\": 9007199254740992}\n}")
	original := string(b.Dialogue[0].Parts[0].Args)
	if !sameContext(a, b) {
		t.Error("JSON whitespace/key order and omitted empties changed context meaning")
	}
	if string(b.Dialogue[0].Parts[0].Args) != original {
		t.Fatal("comparison mutated raw canonical value")
	}
	b.Dialogue[0].Parts[0].Args = json.RawMessage(`{"nested":{"n":9007199254740993,"s":"x"},"list":[1,2]}`)
	if sameContext(a, b) {
		t.Fatal("comparison lost large integer content disagreement")
	}
}
