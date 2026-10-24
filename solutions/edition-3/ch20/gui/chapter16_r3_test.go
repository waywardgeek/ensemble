package gui

import (
	"bytes"
	"encoding/json"
	"ensemble"
	"reflect"
	"strings"
	"testing"
)

func TestCh16R3SummaryCheckpointCatchup(t *testing.T) {
	const summary = "Checking the local details."
	const prefix = "Checking "
	const answer = "The ordinary answer."
	const private = "private-continuation-fixture"
	for _, partial := range []bool{false, true} {
		name := "entirely unseen"
		if partial {
			name = "prefix already seen"
		}
		t.Run(name, func(t *testing.T) {
			app, a := idleAgent(t, false)
			h := NewHub(app)
			h.unsubscribe()
			defer h.Close()
			// Drive the real Hub with immutable checkpoints, controlling coalescing
			// explicitly rather than depending on observer or provider timing.
			h.mu.Lock()
			defer h.mu.Unlock()
			checkpoint := a.View()
			c := h.connectAt(map[string]ensemble.ViewSnapshot{a.ID(): checkpoint})
			_ = replayData(t, c)
			_ = drain(c)
			v := h.view(a)
			var seen string
			if partial {
				checkpoint.Revision++
				checkpoint.TurnID = 9
				checkpoint.Partials = []ensemble.PartDelta{{TurnID: 9, PartID: 101, Kind: ensemble.DeltaReasoningSummary, Chunk: prefix}}
				h.sync(v, checkpoint)
				early := drain(c)
				if len(early) != 1 || early[0].Type != "part_delta" || early[0].Kind != "reasoning_summary" || early[0].PartID != 101 || early[0].TurnID != 9 || early[0].Chunk != prefix || early[0].Replay {
					t.Fatalf("missing causal summary prefix: %+v", early)
				}
				seen = early[0].Chunk
			}
			final := ch10Event(t, `{"type":"response_ended","turn_id":9,"response":{"from":{"vendor":"openai","surface":"responses","model":"gpt-ch19-course"},"parts":[{"type":"reasoning_summary","display_id":101,"text":"Checking the local details."},{"type":"tool_call","display_id":102,"call_id":"local-call","name":"lookup","args":{"key":"local"}},{"type":"text","display_id":103,"text":"The ordinary answer."},{"type":"opaque","display_id":104,"from":{"vendor":"openai","surface":"responses","model":"gpt-ch19-course"},"data":{"type":"reasoning","encrypted_content":"private-continuation-fixture"}}]}}`)
			if err := a.Log().Append(final); err != nil {
				t.Fatal(err)
			}
			canonicalBefore, _ := json.Marshal(a.Log().Events())
			checkpoint.Revision++
			checkpoint.TurnID = 9
			checkpoint.Through = a.Log().LastSeq()
			checkpoint.Partials = nil
			h.sync(v, checkpoint)
			live := drain(c)
			wantOrder := []string{"part_delta:reasoning_summary", "part_final:reasoning_summary", "part_final:tool_call", "part_delta:text", "part_final:text", "part_final:continuation"}
			var order []string
			finals := map[string]uint64{}
			for _, w := range live {
				raw, _ := json.Marshal(w)
				if strings.Contains(string(raw), private) {
					t.Fatal("private continuation leaked")
				}
				if w.Type != "part_delta" && w.Type != "part_final" {
					continue
				}
				kind := w.Kind
				if w.Part != nil {
					kind = w.Part.Kind
					finals[kind] = w.PartID
				}
				order = append(order, w.Type+":"+kind)
				if w.AgentID != a.ID() || w.TurnID != 9 || w.Replay {
					t.Fatalf("wrong live association: %+v", w)
				}
				switch {
				case w.Type == "part_delta" && kind == "reasoning_summary":
					if w.PartID != 101 || w.Chunk != summary[len(seen):] {
						t.Fatalf("wrong/duplicated suffix: seen=%q wire=%+v", seen, w)
					}
					seen += w.Chunk
				case w.Type == "part_delta" && kind == "text":
					if w.PartID != 103 || w.Chunk != answer {
						t.Fatalf("ordinary text mixed with summary: %+v", w)
					}
				case w.Type == "part_final" && kind == "tool_call":
					if w.PartID != 102 || w.Part.CallPartID != 102 || w.Part.CallID != "local-call" || w.Part.Name != "lookup" {
						t.Fatalf("adjacent call lost identity: %+v", w)
					}
				}
			}
			if !reflect.DeepEqual(order, wantOrder) || seen != summary || finals["reasoning_summary"] != 101 || finals["text"] != 103 {
				t.Fatalf("coalesced final lost/reordered live content: seen=%q order=%q finals=%v", seen, order, finals)
			}
			if len(v.partial) != 0 {
				t.Fatalf("final retained stale partials: %+v", v.partial)
			}
			h.sync(v, checkpoint) // same final checkpoint
			checkpoint.Revision++
			h.sync(v, checkpoint) // later wake, same committed frontier
			if again := drain(c); len(again) != 0 {
				t.Fatalf("duplicate final/suffix: %+v", again)
			}
			reconnect := h.connectAt(map[string]ensemble.ViewSnapshot{a.ID(): checkpoint})
			replay := replayData(t, reconnect)
			if len(replay.Partials) != 0 {
				t.Fatalf("reconnect resurrected partials: %+v", replay.Partials)
			}
			recovered := map[string]uint64{}
			for _, w := range replay.Events {
				raw, _ := json.Marshal(w)
				if strings.Contains(string(raw), private) {
					t.Fatal("private continuation leaked on reconnect")
				}
				if w.Type == "part_delta" {
					t.Fatalf("reconnect replayed provider chunks: %+v", w)
				}
				if w.Type == "part_final" {
					if !w.Replay || recovered[w.Part.Kind] != 0 {
						t.Fatalf("duplicate/non-replay final: %+v", w)
					}
					recovered[w.Part.Kind] = w.PartID
				}
			}
			if !reflect.DeepEqual(recovered, finals) {
				t.Fatalf("reconnect changed final identities: live=%v replay=%v", finals, recovered)
			}
			if again := drain(c); len(again) != 0 {
				t.Fatalf("reconnect duplicated live suffix/final: %+v", again)
			}
			canonicalAfter, _ := json.Marshal(a.Log().Events())
			if !bytes.Equal(canonicalBefore, canonicalAfter) {
				t.Fatal("GUI changed canonical content")
			}
		})
	}
}
