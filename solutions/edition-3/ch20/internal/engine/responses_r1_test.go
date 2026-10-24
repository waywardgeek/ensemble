package engine

import (
	"bytes"
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func r1Frame(v any) string { b, _ := json.Marshal(v); return "data: " + string(b) + "\n\n" }
func TestResponsesR1CompletedItems(t *testing.T) {
	cfg := common.Config{Vendor: common.OpenAI, Surface: common.Responses, Model: "gpt-6.1-sol"}
	reasoning := json.RawMessage(`{"type":"reasoning","id":"r","summary":[{"type":"summary_text","text":"Public"}],"encrypted_content":"PRIVATE","provider_extension":"keep"}`)
	call := json.RawMessage(`{"type":"function_call","id":"item","call_id":"call","name":"local","arguments":"{\"x\":1}","status":"completed"}`)
	// Arrival order is not output order. Sparse indices must retain delta identity.
	prefix := r1Frame(map[string]any{"type": "response.created", "response": map[string]any{"model": cfg.Model}}) + r1Frame(map[string]any{"type": "response.reasoning_summary_text.delta", "output_index": 2, "summary_index": 0, "delta": "Public"}) + r1Frame(map[string]any{"type": "response.output_item.added", "output_index": 5, "item": map[string]any{"type": "function_call", "name": "local", "call_id": "call", "arguments": ""}}) + r1Frame(map[string]any{"type": "response.function_call_arguments.delta", "output_index": 5, "delta": `{"x":1}`}) + r1Frame(map[string]any{"type": "response.output_item.done", "output_index": 5, "item": call}) + r1Frame(map[string]any{"type": "response.output_item.done", "output_index": 2, "item": reasoning})
	for _, omit := range []bool{false, true} {
		t.Run(fmt.Sprint("omitted=", omit), func(t *testing.T) {
			terminal := map[string]any{"status": "completed", "usage": map[string]int{"input_tokens": 11, "output_tokens": 4}}
			if !omit {
				terminal["output"] = []any{}
			}
			var deltas []common.PartDelta
			events, err := parseResponsesSSE(strings.NewReader(prefix+r1Frame(map[string]any{"type": "response.completed", "response": terminal})), cfg, common.ParseCallbacks{OnDelta: func(d common.PartDelta) { deltas = append(deltas, d) }})
			if err != nil {
				t.Fatal(err)
			}
			r := events[1].Response
			if len(r.Parts) != 3 || r.Parts[0].Type != "reasoning_summary" || r.Parts[1].Type != "opaque" || r.Parts[2].Type != "tool_call" || r.Parts[0].ID != deltas[0].PartID || r.Parts[2].ID != deltas[1].PartID || !bytes.Equal(r.Parts[1].Data, reasoning) || r.Parts[1].From.Model != cfg.Model || r.From.Model != cfg.Model || r.Usage.Input != 11 || r.Usage.Output != 4 {
				t.Fatalf("lost completed items/identity/usage: %+v", r)
			}
		})
	}
	for _, ending := range []string{"", r1Frame(map[string]any{"type": "response.failed"}), r1Frame(map[string]any{"type": "response.incomplete"}), r1Frame(map[string]any{"type": "response.completed", "response": map[string]any{"status": "completed", "output": []any{}}})} {
		events, err := parseResponsesSSE(strings.NewReader(prefix+ending), cfg, common.ParseCallbacks{})
		if err == nil {
			t.Fatal("unfinished/unaccounted output accepted")
		}
		for _, e := range events {
			if e.Response != nil && len(e.Response.Parts) > 0 {
				t.Fatal("failed stream committed completed items")
			}
		}
	}
	// A nonempty terminal body, not an earlier done copy, is authoritative.
	old := json.RawMessage(`{"type":"message","role":"assistant","content":[{"type":"output_text","text":"obsolete"}]}`)
	terminal := json.RawMessage(`{"status":"completed","model":"gpt-6.1-sol","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"authoritative"}]}],"usage":{"input_tokens":5,"output_tokens":2}}`)
	events, err := parseResponsesSSE(strings.NewReader(r1Frame(map[string]any{"type": "response.output_item.done", "output_index": 0, "item": old})+r1Frame(map[string]any{"type": "response.completed", "response": terminal})), cfg, common.ParseCallbacks{})
	if err != nil || len(events[1].Response.Parts) != 1 || events[1].Response.Parts[0].Text != "authoritative" {
		t.Fatal("terminal did not replace done copy", events, err)
	}
}

func TestResponsesR1ModelPolicy(t *testing.T) {
	for _, model := range []string{"gpt-ch19-course", "gpt-6.1-sol", "gpt-5.6-sol"} {
		for _, effort := range []string{"", "none", "high"} {
			for _, summary := range []string{"", "none", "detailed"} {
				t.Run(model+"/"+effort+"/"+summary, func(t *testing.T) {
					c := common.Config{Vendor: common.OpenAI, Model: model, ReasoningEffort: effort, ReasoningSummary: summary, RequireToolExplanation: true}
					if err := common.NormalizeOpenAI(&c); err != nil {
						t.Fatal(err)
					}
					if c.Surface != common.Responses || c.RequireToolExplanation {
						t.Fatalf("wrong selected model policy: %+v", c)
					}
					b, err := render(common.Context{}, c, nil)
					if err != nil {
						t.Fatal(err)
					}
					var body struct {
						Reasoning map[string]string `json:"reasoning"`
					}
					json.Unmarshal(b, &body)
					wantEffort, wantSummary := effort, summary
					if wantEffort == "" {
						wantEffort = "medium"
					}
					if wantSummary == "" {
						wantSummary = "auto"
					}
					if wantSummary == "none" {
						wantSummary = ""
					}
					if body.Reasoning["effort"] != wantEffort || body.Reasoning["summary"] != wantSummary {
						t.Fatalf("wrong reasoning %s", b)
					}
				})
			}
		}
	}
	for _, gate := range []bool{false, true} {
		c := common.Config{Vendor: common.OpenAI, Surface: common.Responses, Model: "unknown", ReasoningEffort: "high", ReasoningSummary: "detailed", RequireToolExplanation: gate}
		if err := common.NormalizeOpenAI(&c); err != nil {
			t.Fatal(err)
		}
		b, err := render(common.Context{}, c, nil)
		if err != nil || bytes.Contains(b, []byte(`"reasoning"`)) || c.RequireToolExplanation != gate {
			t.Fatal("unknown policy guessed or explicit gate lost", string(b), err)
		}
	}
	c := common.Config{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "gpt-6.1-sol", RequireToolExplanation: true}
	if err := common.NormalizeOpenAI(&c); err != nil || c.Surface != common.ChatCompletions || !c.RequireToolExplanation {
		t.Fatal("non-Responses contract changed", c, err)
	}
}

func TestResponsesR1ReplayPairing(t *testing.T) {
	call := func(id string) common.Entry {
		return common.Entry{Actor: common.Assistant, Parts: common.PartList{{Type: "tool_call", CallID: id, Name: "local", Args: json.RawMessage(`{}`)}}}
	}
	result := func(id, text string) common.Entry {
		return common.Entry{Actor: common.Tool, Parts: common.PartList{{Type: "tool_result", CallID: id, Parts: common.PartList{common.Text(text)}}}}
	}
	for _, tc := range []struct {
		name    string
		entries []common.Entry
		bad     bool
		want    []string
	}{
		{"repeated calls", []common.Entry{call("c"), result("c", "first"), call("c"), result("c", "second")}, true, nil},
		{"duplicate result", []common.Entry{call("c"), result("c", "first"), result("c", "second")}, true, nil},
		{"orphan", []common.Entry{result("absent", "orphan")}, true, nil},
		{"late unique", []common.Entry{call("a"), call("b"), result("b", "B"), result("a", "A")}, false, []string{"A", "B"}},
		{"missing", []common.Entry{call("a")}, false, []string{"No result was recorded; execution outcome unknown."}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := common.Context{Dialogue: tc.entries}
			before, _ := json.Marshal(c)
			b, err := render(c, common.Config{Vendor: common.OpenAI, Surface: common.Responses, Model: "gpt-6.1-sol"}, nil)
			after, _ := json.Marshal(c)
			if !bytes.Equal(before, after) {
				t.Fatal("canonical dialogue mutated")
			}
			if tc.bad {
				if err == nil {
					t.Fatalf("ambiguous pairing accepted: %s", b)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var body struct {
				Input []struct {
					Type, Output string
					CallID       string `json:"call_id"`
				}
			}
			json.Unmarshal(b, &body)
			var outputs, callIDs, resultIDs, wantIDs []string
			for _, entry := range tc.entries {
				for _, part := range entry.Parts {
					if part.Type == "tool_call" {
						wantIDs = append(wantIDs, part.CallID)
					}
				}
			}
			for _, i := range body.Input {
				if i.Type == "function_call" {
					callIDs = append(callIDs, i.CallID)
				}
				if i.Type == "function_call_output" {
					outputs = append(outputs, i.Output)
					resultIDs = append(resultIDs, i.CallID)
				}
			}
			if !reflect.DeepEqual(callIDs, wantIDs) || !reflect.DeepEqual(resultIDs, wantIDs) {
				t.Fatalf("call/result IDs rewritten: %s", b)
			}
			if !reflect.DeepEqual(outputs, tc.want) {
				t.Fatalf("wrong actual pairing %s", b)
			}
		})
	}
}
