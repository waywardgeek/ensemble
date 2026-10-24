package engine

import (
	"bytes"
	"encoding/json"
	"ensemble/internal/common"
	"reflect"
	"testing"
)

func TestOpenAIToolResultProjection(t *testing.T) {
	cfg := common.Config{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "route", ResolvedModel: "producer"}
	source := cfg.Target()
	call := func(id string) common.Part {
		return common.Part{Type: "tool_call", CallID: id, Name: "job", Args: json.RawMessage(`{"n":1}`)}
	}
	result := func(id, text string) common.Part {
		return common.Part{Type: "tool_result", CallID: id, Name: "job", Parts: common.PartList{common.Text(text)}, IsError: id == "a"}
	}
	entry := func(actor common.Actor, parts ...common.Part) common.Entry {
		return common.Entry{Actor: actor, Parts: parts}
	}
	c := common.Context{Dialogue: []common.Entry{
		entry(common.Human, common.Text("before")),
		entry(common.Assistant, common.Part{Type: "text", Text: "group one", From: &source}, common.Part{Type: "opaque", From: &source, Data: json.RawMessage(`{"reasoning_content":"private"}`)}, call("a"), call("b")),
		entry(common.Human, common.Text("hint one")),
		entry(common.Tool, result("b", "B running")),
		entry(common.Human, common.Part{Type: "redacted", Stub: "hint two"}),
		entry(common.Tool, result("a", "A failed")),
		entry(common.Assistant, common.Text("between groups")),
		entry(common.Human, common.Text("next question")),
		entry(common.Assistant, call("c")),
		entry(common.Human, common.Text("hint three")),
		entry(common.Tool, result("c", "C complete")),
		entry(common.Human, common.Text("after")),
	}, Ephemera: common.PartList{common.Text("ephemeral")}}
	want := []struct{ role, id, text string }{
		{"user", "", "before"}, {"assistant", "", "group one"},
		{"tool", "b", "B running"}, {"tool", "a", "A failed"},
		{"user", "", "hint one"}, {"user", "", "hint two"},
		{"assistant", "", "between groups"}, {"user", "", "next question"},
		{"assistant", "", ""}, {"tool", "c", "C complete"}, {"user", "", "hint three"},
		{"user", "", "after"}, {"system", "", "ephemeral"},
	}
	before, _ := json.Marshal(c)
	for _, target := range []string{"producer", "other-producer"} {
		t.Run(target, func(t *testing.T) {
			cfg.ResolvedModel = target
			data, err := render(c, cfg, nil)
			if err != nil {
				t.Fatal(err)
			}
			var body struct {
				Messages []oMessage `json:"messages"`
			}
			if err = json.Unmarshal(data, &body); err != nil {
				t.Fatal(err)
			}
			if len(body.Messages) != len(want)+1 {
				t.Fatalf("messages=%s", data)
			}
			for i, w := range want {
				m := body.Messages[i+1]
				text, _ := m.Content.(string)
				if m.Role != w.role || m.ToolCallID != w.id || text != w.text {
					t.Errorf("message %d=%+v; want %+v", i+1, m, w)
				}
			}
			first := body.Messages[2]
			if len(first.ToolCalls) != 2 || first.ToolCalls[0].ID != "a" || first.ToolCalls[1].ID != "b" || first.ToolCalls[0].Function.Arguments != `{"n":1}` {
				t.Errorf("call group changed: %+v", first)
			}
			if len(body.Messages[9].ToolCalls) != 1 || body.Messages[9].ToolCalls[0].ID != "c" {
				t.Errorf("second group changed: %+v", body.Messages[9])
			}
			if target == "producer" {
				if first.ReasoningContent == nil || *first.ReasoningContent != "private" {
					t.Errorf("compatible opaque lost: %+v", first)
				}
			} else if first.ReasoningContent != nil {
				t.Errorf("incompatible opaque replay: %+v", first)
			}
			again, err := render(c, cfg, nil)
			if err != nil || !bytes.Equal(data, again) {
				t.Fatalf("projection not repeatable: %v", err)
			}
			after, _ := json.Marshal(c)
			if !bytes.Equal(before, after) {
				t.Fatal("source Context mutated")
			}
		})
	}
	// Incomplete groups keep their established unknown-result behavior. Do not
	// invent a result, partially hoist a batch, or cross unrelated assistant text.
	for _, tc := range []struct {
		name     string
		dialogue []common.Entry
		roles    []string
	}{
		{"incomplete", []common.Entry{entry(common.Assistant, call("a"), call("b")), entry(common.Human, common.Text("hint")), entry(common.Tool, result("a", "A running"))}, []string{"system", "assistant", "user", "tool"}},
		{"unrelated assistant", []common.Entry{entry(common.Assistant, call("a")), entry(common.Human, common.Text("hint")), entry(common.Assistant, common.Text("unrelated")), entry(common.Tool, result("a", "A complete"))}, []string{"system", "assistant", "user", "assistant", "tool"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := render(common.Context{Dialogue: tc.dialogue}, cfg, nil)
			if err != nil {
				t.Fatal(err)
			}
			var body struct {
				Messages []oMessage `json:"messages"`
			}
			if err = json.Unmarshal(data, &body); err != nil {
				t.Fatal(err)
			}
			var roles []string
			for _, m := range body.Messages {
				roles = append(roles, m.Role)
			}
			if !reflect.DeepEqual(roles, tc.roles) {
				t.Fatalf("unknown/unrelated behavior changed: %s", data)
			}
			count := 0
			for _, m := range body.Messages {
				if m.Role == "tool" {
					count++
					if m.ToolCallID != "a" {
						t.Fatalf("fabricated result: %+v", m)
					}
				}
			}
			if count != 1 {
				t.Fatalf("result count=%d", count)
			}
		})
	}
}
