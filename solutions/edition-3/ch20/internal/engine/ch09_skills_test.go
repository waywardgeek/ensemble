package engine

import (
	"encoding/json"
	"ensemble/internal/common"
	"strings"
	"testing"
)

func TestSkillToolReferencesAreContentNotInstructions(t *testing.T) {
	defs := []common.ToolDefinition{{Name: "load_skill", Description: "load", Parameters: json.RawMessage(`{"type":"object"}`)}, {Name: "write_file", Description: "write", Parameters: json.RawMessage(`{"type":"object"}`)}}
	catalog := []common.ToolSpec{{ToolDefinition: defs[0], Enabled: true, Origin: "initial"}, {ToolDefinition: defs[1], Enabled: true, Origin: "tool", EventSeq: 7}}
	c := common.Context{Dialogue: []common.Entry{{Actor: common.Assistant, Parts: common.PartList{{Type: "tool_call", Name: "load_skill", CallID: "load", Args: json.RawMessage(`{"name":"writer"}`)}}}, {Actor: common.Tool, Parts: common.PartList{{Type: "tool_result", CallID: "load", Parts: common.PartList{common.Text("ordinary result")}}}}}}
	for _, vendor := range []common.Vendor{common.Anthropic, common.OpenAI, common.Gemini} {
		cfg := common.Config{Vendor: vendor, Model: "fixture", SystemPrompt: "startup\nloaded instructions"}
		raw, err := render(c, cfg, defs, anthropicDisclosure{catalog: catalog, references: map[string][]string{"load": {"write_file"}}})
		if err != nil {
			t.Fatal(err)
		}
		text := string(raw)
		if !strings.Contains(text, "loaded instructions") {
			t.Fatalf("%s missing system instructions", vendor)
		}
		if (vendor == common.Anthropic) != strings.Contains(text, "tool_reference") {
			t.Fatalf("%s reference projection: %s", vendor, text)
		}
		if vendor == common.Anthropic {
			if strings.Contains(text, "ordinary result") {
				t.Error("reference-bearing wire result contains canonical status text")
			}
		} else {
			ordinary, err := render(c, cfg, defs)
			if err != nil {
				t.Fatal(err)
			}
			if string(ordinary) != text || !strings.Contains(text, "ordinary result") {
				t.Fatalf("%s disclosure changed ordinary result: %s", vendor, text)
			}
		}
		if c.Dialogue[1].Parts[0].Parts[0].Text != "ordinary result" {
			t.Fatal("projection rewrote canonical status")
		}
	}
	for _, model := range []string{"claude-opus-4-1", "claude-sonnet-4", "unknown"} {
		if supportsToolReferences(model) {
			t.Fatal(model)
		}
	}
	if !supportsToolReferences("claude-sonnet-4-5-20250929") {
		t.Fatal("documented model unsupported")
	}
}

func TestChapter09Round2OrdinaryResults(t *testing.T) {
	for _, tc := range []struct {
		name, tool                 string
		reference, failed, missing bool
	}{
		{name: "successful load without references", tool: "load_skill"},
		{name: "failed load", tool: "load_skill", reference: true, failed: true},
		{name: "missing result", tool: "load_skill", reference: true, missing: true},
		{name: "ordinary tool", tool: "read_file", reference: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := common.Context{Dialogue: []common.Entry{{Actor: common.Assistant, Parts: common.PartList{{Type: "tool_call", Name: tc.tool, CallID: "call", Args: json.RawMessage(`{}`)}}}}}
			if !tc.missing {
				c.Dialogue = append(c.Dialogue, common.Entry{Actor: common.Tool, Parts: common.PartList{{Type: "tool_result", CallID: "call", IsError: tc.failed, Parts: common.PartList{common.Text("ordinary result")}}}})
			}
			disclosure := anthropicDisclosure{}
			if tc.reference {
				disclosure.references = map[string][]string{"call": {"write_file"}}
			}
			b, err := render(c, common.Config{Vendor: common.Anthropic}, nil, disclosure)
			if err != nil {
				t.Fatal(err)
			}
			var wire struct {
				Messages []aMessage `json:"messages"`
			}
			if err := json.Unmarshal(b, &wire); err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, message := range wire.Messages {
				for _, raw := range message.Content {
					var block aBlock
					if err := json.Unmarshal(raw, &block); err != nil {
						t.Fatal(err)
					}
					if block.Type != "tool_result" {
						continue
					}
					count++
					want := "ordinary result"
					if tc.missing {
						want = "No result was recorded; execution outcome unknown."
					}
					if message.Role != "user" || block.ToolUseID != "call" || block.Content == nil || *block.Content != want || block.IsError != (tc.failed || tc.missing) {
						t.Fatalf("ordinary result changed: %s", b)
					}
				}
			}
			if count != 1 || strings.Contains(string(b), "tool_reference") {
				t.Fatalf("unexpected reference/result projection: %s", b)
			}
		})
	}
}
