package engine

import (
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"strings"
	"testing"
)

func TestChapter14ProviderContext(t *testing.T) {
	d := common.ToolDeclaration{Name: "fresh", Description: "fresh declaration", Schema: json.RawMessage(`{"type":"object"}`)}
	delta, _ := json.Marshal(common.ToolsData{Added: []common.ToolDeclaration{d}})
	c := common.Context{System: "preamble", ToolsInitialized: true, BaseThrough: 1, Tools: []common.ToolDeclaration{d}, ActiveSkills: []common.SkillData{{Name: "base", Body: "BASE-INSTRUCTIONS"}, {Name: "late", Body: "LATE-INSTRUCTIONS", ActiveSeq: 3}}, Dialogue: []common.Entry{
		{Seq: 2, Actor: common.Human, Parts: common.PartList{common.Text("question")}},
		{Seq: 3, Actor: common.System, Kind: common.KindSkill, Skill: &common.SkillData{Name: "late", Body: "LATE-INSTRUCTIONS", ActiveSeq: 3}},
		{Seq: 4, Actor: common.System, Kind: common.KindTools, Tools: &common.ToolsData{Added: []common.ToolDeclaration{d}}, Parts: common.PartList{{Type: "tool_decls", Data: delta}}},
	}, Ephemera: common.PartList{common.Text("LATEST-HINT")}}
	for _, tc := range []struct {
		vendor  common.Vendor
		surface common.Surface
		model   string
	}{{common.Anthropic, common.Messages, "claude-sonnet-4-6"}, {common.OpenAI, common.ChatCompletions, "gpt-5.4"}, {common.Gemini, common.GenerateContent, "gemini-3.8-flash"}, {common.Anthropic, common.Messages, "claude-sonnet-5-course"}, {common.Anthropic, common.Messages, "claude-opus-5-course"}} {
		t.Run(tc.model+fmt.Sprint(tc.surface), func(t *testing.T) {
			data, err := render(c, common.Config{Vendor: tc.vendor, Surface: tc.surface, Model: tc.model}, []common.ToolDefinition{{Name: d.Name, Description: d.Description, Parameters: d.Schema}})
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			for _, want := range []string{"preamble", "BASE-INSTRUCTIONS", "LATE-INSTRUCTIONS", "fresh", "LATEST-HINT"} {
				if !strings.Contains(text, want) {
					t.Fatalf("missing %s: %s", want, text)
				}
			}
			var wire map[string]json.RawMessage
			if err := json.Unmarshal(data, &wire); err != nil {
				t.Fatal(err)
			}
			var system, dialogue json.RawMessage
			switch tc.vendor {
			case common.Anthropic:
				system, dialogue = wire["system"], wire["messages"]
			case common.Gemini:
				system, dialogue = wire["systemInstruction"], wire["contents"]
			case common.OpenAI:
				var messages []struct {
					Role    string          `json:"role"`
					Content json.RawMessage `json:"content"`
				}
				if err := json.Unmarshal(wire["messages"], &messages); err != nil {
					t.Fatal(err)
				}
				if len(messages) == 0 || messages[0].Role != "system" {
					t.Fatal("missing initial SYSTEM message")
				}
				system = messages[0].Content
				dialogue, _ = json.Marshal(messages[1:])
			}
			for _, body := range []string{"preamble", "BASE-INSTRUCTIONS", "LATE-INSTRUCTIONS"} {
				if strings.Count(string(system), body) != 1 {
					t.Errorf("SYSTEM count for %s is not one: %s", body, system)
				}
				if strings.Contains(string(dialogue), body) {
					t.Errorf("instruction %s leaked into dialogue: %s", body, dialogue)
				}
			}
			if (tc.model == "claude-opus-5-course") != strings.Contains(text, "tool_addition") {
				t.Fatalf("wrong declaration lane: %s", text)
			}
		})
	}
}
