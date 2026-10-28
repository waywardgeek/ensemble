package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

func appendHandoffFact(t *testing.T, h common.History, e common.Event) {
	t.Helper()
	if err := h.Append(e); err != nil {
		t.Fatal(err)
	}
}

func assertLatestHandoff(t *testing.T, h common.History, text string, seq uint64) {
	t.Helper()
	var notes []common.Entry
	for _, entry := range h.Context().Dialogue {
		if entry.Kind == common.KindHandoff {
			notes = append(notes, entry)
		}
	}
	if len(notes) != 1 || notes[0].Seq != seq || len(notes[0].Parts) != 1 || notes[0].Parts[0].Text != text {
		t.Fatalf("active handoffs = %+v, want only %q at %d", notes, text, seq)
	}
}

// Repeated checkpoints replace only the active note. Original events, including
// each checkpoint's sequence boundary, survive replay and both save formats.
func TestMicroHandoffReplacesOlderNotes(t *testing.T) {
	a := loaded(t, ensemble.OpenAI, conversation+firstResult+lastResult)
	defer a.Shutdown()
	h := a.History()
	path := filepath.Join(a.Config().DataDir, "handoff.json")
	appendHandoffFact(t, h, common.Event{Type: "skill_loaded", Skill: &common.SkillData{Name: "kept-skill", Body: "kept instructions"}})
	appendHandoffFact(t, h, common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "old note"}})
	if err := h.Checkpoint(path); err != nil {
		t.Fatal(err)
	}
	appendHandoffFact(t, h, common.Event{Type: "recall_attached", Recall: &common.RecallData{Parts: []common.Part{{Type: "text", Text: "old recall"}}}})
	appendHandoffFact(t, h, common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: "current dialogue"}}}})
	appendHandoffFact(t, h, common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "middle note"}})
	appendHandoffFact(t, h, common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "newest note"}})
	assertLatestHandoff(t, h, "newest note", 11)

	wantContext, wantEvents := h.Context(), h.EventsAfter(0)
	var boundaries []uint64
	var notes []string
	for _, event := range wantEvents {
		if event.Type == "micro_handoff" {
			boundaries = append(boundaries, event.Seq)
			notes = append(notes, event.Handoff.Text)
		}
	}
	if !reflect.DeepEqual(boundaries, []uint64{7, 10, 11}) || !reflect.DeepEqual(notes, []string{"old note", "middle note", "newest note"}) {
		t.Fatalf("archive lost checkpoint records: %v %v", boundaries, notes)
	}
	var log, save bytes.Buffer
	if err := h.Dump(&log); err != nil {
		t.Fatal(err)
	}
	if err := h.Save(&save); err != nil {
		t.Fatal(err)
	}
	for _, variant := range []string{"live", "replay", "snapshot", "journal-tail"} {
		t.Run(variant, func(t *testing.T) {
			var err error
			switch variant {
			case "replay":
				err = h.Load(bytes.NewReader(log.Bytes()))
			case "snapshot":
				err = h.Restore(bytes.NewReader(save.Bytes()))
			case "journal-tail":
				err = h.Recover(path)
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(h.Context(), wantContext) || !reflect.DeepEqual(h.EventsAfter(0), wantEvents) {
				t.Fatal("recovery changed context or original events")
			}
			raw, err := a.Engine().Render(h.Context())
			if err != nil {
				t.Fatal(err)
			}
			for _, wanted := range []string{"hello", "decision", "current dialogue", "kept instructions", "newest note"} {
				if !bytes.Contains(raw, []byte(wanted)) {
					t.Errorf("request lost %q", wanted)
				}
			}
			for _, unwanted := range []string{"old note", "middle note", "old recall", "secret", "tool_call", "tool_result"} {
				if bytes.Contains(raw, []byte(unwanted)) {
					t.Errorf("request retained %q", unwanted)
				}
			}
			if bytes.Count(raw, []byte("micro_handoff completed")) != 1 {
				t.Fatal("request must render exactly one completed checkpoint")
			}
		})
	}
}

// A sibling tool holds both new notes out of Dialogue. Landing them in capture
// order must replace the old note and then the first held note, not the skill.
func TestMicroHandoffReplacesHeldNotes(t *testing.T) {
	a := loaded(t, ensemble.OpenAI, "")
	defer a.Shutdown()
	h := a.History()
	appendHandoffFact(t, h, common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "prior checkpoint"}})
	from := common.Provenance{Vendor: common.OpenAI, Surface: common.Responses, Model: "model"}
	appendHandoffFact(t, h, common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: "hello"}}}})
	appendHandoffFact(t, h, common.Event{Type: "request_sent", Request: &common.RequestData{To: from}})
	appendHandoffFact(t, h, common.Event{Type: "response_ended", Response: &common.ResponseData{From: from, Parts: []common.Part{
		{Type: "text", Text: "decision"},
		{Type: "tool_call", CallID: "one", Name: "micro_handoff", Args: []byte(`{"text":"first held note"}`), From: from},
		{Type: "tool_call", CallID: "two", Name: "micro_handoff", Args: []byte(`{"text":"last held note"}`), From: from},
		{Type: "tool_call", CallID: "three", Name: "sibling", Args: []byte(`{}`), From: from},
	}}})
	appendHandoffFact(t, h, common.Event{Type: "tool_returned", Tool: &common.ToolData{CallID: "one", Parts: []common.Part{{Type: "text", Text: "first acknowledgment"}}}})
	appendHandoffFact(t, h, common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "first held note"}})
	appendHandoffFact(t, h, common.Event{Type: "skill_loaded", Skill: &common.SkillData{Name: "held-skill", Body: "held instructions"}})
	appendHandoffFact(t, h, common.Event{Type: "tool_returned", Tool: &common.ToolData{CallID: "two", Parts: []common.Part{{Type: "text", Text: "last acknowledgment"}}}})
	appendHandoffFact(t, h, common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "last held note"}})
	assertLatestHandoff(t, h, "prior checkpoint", 1)
	if len(h.Context().Held) != 3 || h.Context().Turn != "tools_pending" {
		t.Fatal("notes and skill must wait for the remaining sibling")
	}
	var log, save bytes.Buffer
	if err := h.Dump(&log); err != nil {
		t.Fatal(err)
	}
	if err := h.Save(&save); err != nil {
		t.Fatal(err)
	}
	for _, variant := range []string{"live", "replay", "snapshot"} {
		t.Run(variant, func(t *testing.T) {
			var err error
			if variant == "replay" {
				err = h.Load(bytes.NewReader(log.Bytes()))
			} else if variant == "snapshot" {
				err = h.Restore(bytes.NewReader(save.Bytes()))
			}
			if err != nil {
				t.Fatal(err)
			}
			assertLatestHandoff(t, h, "prior checkpoint", 1)
			appendHandoffFact(t, h, common.Event{Type: "tool_returned", Tool: &common.ToolData{CallID: "three", Parts: []common.Part{{Type: "text", Text: "last result"}}}})
			assertLatestHandoff(t, h, "last held note", 9)
			c := h.Context()
			if len(c.Held) != 0 || c.Turn != "input_pending" || countKind(c, common.KindSkill) != 1 || len(h.EventsAfter(0)) != 10 {
				t.Fatal("landing lost skill, archive or batch state")
			}
			for _, entry := range c.Dialogue {
				for _, part := range entry.Parts {
					if part.Type != "text" {
						t.Fatalf("tool traffic survived checkpoint: %+v", part)
					}
				}
			}
		})
	}
}

// Failure to append the durable event cannot replace the committed note or
// consume its successor's sequence number. No separate handoff transaction is needed.
func TestMicroHandoffJournalFailureKeepsPriorNote(t *testing.T) {
	a := loaded(t, ensemble.OpenAI, "")
	defer a.Shutdown()
	h := a.History()
	path := filepath.Join(a.Config().DataDir, "handoff.json")
	appendHandoffFact(t, h, common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "committed note"}})
	if err := h.Checkpoint(path); err != nil {
		t.Fatal(err)
	}
	before, events := h.Context(), h.EventsAfter(0)
	// A directory reliably refuses journal writes, including for privileged tests.
	if err := os.Remove(path + ".journal"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path+".journal", 0700); err != nil {
		t.Fatal(err)
	}
	replacement := common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "replacement note"}}
	if err := h.Append(replacement); err == nil {
		t.Fatal("handoff succeeded without a writable journal")
	}
	if !reflect.DeepEqual(h.Context(), before) || !reflect.DeepEqual(h.EventsAfter(0), events) {
		t.Fatal("failed journal append changed committed state")
	}
	if err := os.Remove(path + ".journal"); err != nil {
		t.Fatal(err)
	}
	if err := h.Recover(path); err != nil {
		t.Fatal(err)
	}
	assertLatestHandoff(t, h, "committed note", 1)
	appendHandoffFact(t, h, replacement)
	if err := h.Recover(path); err != nil {
		t.Fatal(err)
	}
	assertLatestHandoff(t, h, "replacement note", 2)
	if len(h.EventsAfter(0)) != 2 {
		t.Fatal("retry lost the old event or archived a failed append")
	}
}

// Exercise the actual tool/actor/HTTP path, not only reduction or Render. The
// legacy snapshot deliberately has multiple active notes, as a pre-fix process
// could save. No migration is assumed: the next successful handoff replaces them.
func TestMicroHandoffReplacementOnWire(t *testing.T) {
	for _, responses := range []bool{false, true} {
		for _, legacy := range []bool{false, true} {
			t.Run(fmt.Sprintf("responses=%t/legacy=%t", responses, legacy), func(t *testing.T) {
				const old1, old2 = "OBSOLETE_CHECKPOINT_ALPHA", "OBSOLETE_CHECKPOINT_BETA"
				const first, latest = "CURRENT_CHECKPOINT_ONE", "CURRENT_CHECKPOINT_TWO"
				const dialogue = "ORDINARY_DIALOGUE_MUST_SURVIVE"
				var calls atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					raw, err := io.ReadAll(r.Body)
					if err != nil {
						t.Error(err)
						http.Error(w, "read", 500)
						return
					}
					n := calls.Add(1)
					wire := string(raw)
					if !strings.Contains(wire, dialogue) {
						t.Errorf("request %d lost ordinary dialogue", n)
					}
					if n == 1 && legacy {
						for _, note := range []string{old1, old2} {
							if !strings.Contains(wire, note) {
								t.Errorf("fixture did not restore legacy note %s", note)
							}
						}
					}
					if n >= 2 {
						want := first
						forbidden := []string{old1, old2}
						if n >= 3 {
							want = latest
							forbidden = append(forbidden, first)
						}
						if strings.Count(wire, want) != 1 || strings.Count(wire, "micro_handoff completed. Your checkpoint note:") != 1 {
							t.Errorf("request %d must contain exactly one latest handoff", n)
						}
						for _, note := range forbidden {
							if strings.Contains(wire, note) {
								t.Errorf("request %d retained older handoff %s", n, note)
							}
						}
					}
					if n > 4 {
						t.Errorf("unexpected extra request %d", n)
						http.Error(w, "extra request", 500)
						return
					}
					note := first
					if n == 2 {
						note = latest
					}
					if responses {
						if r.URL.Path != "/responses" {
							t.Errorf("wrong Responses route %s", r.URL.Path)
						}
						output := []any{map[string]any{"type": "message", "content": []any{map[string]string{"type": "output_text", "text": "done"}}}}
						if n <= 2 {
							args, _ := json.Marshal(map[string]string{"text": note})
							output[0] = map[string]any{"type": "message", "content": []any{map[string]string{"type": "output_text", "text": "I will save a checkpoint."}}}
							output = append(output, map[string]any{"type": "function_call", "call_id": fmt.Sprintf("checkpoint-%d", n), "name": "micro_handoff", "arguments": string(args)})
						}
						data, _ := json.Marshal(map[string]any{"response": map[string]any{"status": "completed", "model": "gpt-6-astra", "output": output, "usage": map[string]int{"input_tokens": 10, "output_tokens": 2}}})
						w.Header().Set("Content-Type", "text/event-stream")
						fmt.Fprintf(w, "event: response.completed\ndata: %s\n\n", data)
					} else {
						content := []map[string]any{{"type": "text", "text": "done"}}
						if n <= 2 {
							content = []map[string]any{contextCall(fmt.Sprintf("checkpoint-%d", n), "micro_handoff", map[string]any{"text": note})}
						}
						json.NewEncoder(w).Encode(map[string]any{"content": content, "usage": map[string]int{"input_tokens": 10, "output_tokens": 2}})
					}
				}))
				defer server.Close()
				dir := t.TempDir()
				cfg := ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", BaseURL: server.URL, DataDir: dir, SavePath: filepath.Join(dir, "session.json"), BuiltinTools: true, DisableStreaming: true, Recall: ensemble.RecallSettings{Disabled: true}}
				if responses {
					cfg.Vendor, cfg.Model, cfg.BillingMode = ensemble.OpenAI, "gpt-6-astra", ensemble.CredentialCodex
					cfg.CodexAuthFile = filepath.Join(dir, "dummy-auth.json")
					writeCodex(t, cfg.CodexAuthFile, "dummy-token", "dummy-account")
				}
				if legacy {
					// Construct old-format active state explicitly: replaying old events
					// through today's reducer would already remove the older note.
					sf := common.SaveFile{Config: common.SaveConfig{Vendor: cfg.Vendor, Model: cfg.Model}, AsOf: 2, Context: &common.Context{Turn: "idle"}}
					for i, note := range []string{old1, old2} {
						seq := uint64(i + 1)
						sf.Context.Dialogue = append(sf.Context.Dialogue, common.Entry{Seq: seq, Kind: common.KindHandoff, Actor: common.System, Parts: []common.Part{{Type: "text", Text: note}}})
						sf.Log = append(sf.Log, common.Event{Seq: seq, Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: note}})
					}
					data, err := json.Marshal(sf)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(cfg.SavePath, data, 0600); err != nil {
						t.Fatal(err)
					}
				}
				a, err := ensemble.New(io.Discard).NewAgent(cfg)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { a.Shutdown() })
				if err := a.Load(); err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if answer, err := a.Ask(ctx, dialogue); err != nil || answer != "done" {
					t.Fatalf("tool loop: answer=%q err=%v", answer, err)
				}
				if calls.Load() != 3 {
					t.Fatalf("tool loop requests=%d, want 3", calls.Load())
				}
				if err := a.Save(); err != nil {
					t.Fatal(err)
				}
				if err := a.Shutdown(); err != nil {
					t.Fatal(err)
				}
				restored, err := ensemble.New(io.Discard).NewAgent(cfg)
				if err != nil {
					t.Fatal(err)
				}
				defer restored.Shutdown()
				if err := restored.Load(); err != nil {
					t.Fatal(err)
				}
				if answer, err := restored.Ask(ctx, "Continue after restart."); err != nil || answer != "done" {
					t.Fatalf("restart: answer=%q err=%v", answer, err)
				}
				if calls.Load() != 4 {
					t.Fatalf("total requests=%d, want 4", calls.Load())
				}
				var archived []string
				for _, event := range restored.History().EventsAfter(0) {
					if event.Type == "micro_handoff" {
						archived = append(archived, event.Handoff.Text)
					}
				}
				want := []string{first, latest}
				if legacy {
					want = append([]string{old1, old2}, want...)
				}
				if !reflect.DeepEqual(archived, want) {
					t.Fatalf("archived handoffs=%v, want %v", archived, want)
				}
			})
		}
	}
}
