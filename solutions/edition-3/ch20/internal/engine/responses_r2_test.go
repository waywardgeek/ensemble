package engine

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"ensemble/internal/common"
)

func TestResponsesR2CachePreservesContent(t *testing.T) {
	cfg := common.Config{Vendor: common.OpenAI, Surface: common.Responses, Model: "gpt-ch19-course", ResolvedModel: "gpt-ch19-course", AuthMode: common.AuthAPI, StreamCapability: common.StreamSSE}
	from := common.Provenance{Vendor: cfg.Vendor, Surface: cfg.Surface, Model: cfg.Model}
	c := common.Context{System: "Exact constitution\nWith its own whitespace.", Dialogue: []common.Entry{
		{Actor: common.Human, Parts: common.PartList{common.Text("original user input")}},
		{Actor: common.Assistant, Parts: common.PartList{{Type: "opaque", From: &from, Data: json.RawMessage(`{"type":"reasoning","id":"reason","encrypted_content":"PRIVATE","provider_extra":{"prompt_cache_options":"opaque","prompt_cache_breakpoint":"not a marker"}}`)}, {Type: "tool_call", CallID: "original-id", Name: "local", Args: json.RawMessage(`{"x":"exact arg"}`)}}},
		{Actor: common.Tool, Parts: common.PartList{{Type: "tool_result", CallID: "original-id", Parts: common.PartList{common.Text("actual result")}}}},
		{Actor: common.Human, Parts: common.PartList{common.Text("last retained text")}},
	}, Ephemera: common.PartList{common.Text("temporary tail")}}
	before, _ := json.Marshal(c)
	marked, err := render(c, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	plainCfg := cfg
	plainCfg.AuthMode = common.AuthChatGPT
	plain, err := render(c, plainCfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(c)
	if !bytes.Equal(before, after) {
		t.Fatal("canonical context changed")
	}
	clean := stripCacheMarkers(cfg, marked)
	var got, want map[string]any
	if err = json.Unmarshal(clean, &got); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(plain, &want); err != nil {
		t.Fatal(err)
	}
	if got["prompt_cache_options"] != nil {
		t.Fatal("Responses lens retained top-level cache mode")
	}
	input := got["input"].([]any)
	developer := input[0].(map[string]any)
	content := developer["content"].([]any)
	block := content[0].(map[string]any)
	if developer["role"] != "developer" || block["type"] != "input_text" || block["text"] != want["instructions"] {
		t.Fatalf("constitution changed: %s", marked)
	}
	got["instructions"] = block["text"]
	got["input"] = input[1:]
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("nonmarker data changed:\nmarked %s\nplain %s", clean, plain)
	}
	root, ok := wireRoot(marked)
	if !ok {
		t.Fatal("invalid wire")
	}
	_, markers := markerMembers(cfg, root, marked)
	var boundaries []string
	for _, item := range root.field("input").items {
		for _, part := range item.field("content").items {
			if markers[part.start] {
				boundaries = append(boundaries, wireText(marked, part.field("text")))
			}
		}
	}
	if len(boundaries) != 2 || boundaries[0] != want["instructions"] || boundaries[1] != "last retained text" {
		t.Fatalf("wrong boundaries: %q", boundaries)
	}
	var lens cache
	lens.Observe(cfg, marked)
	if lens.Status().MarkerCount != 2 {
		t.Fatalf("lens marker count: %+v", lens.Status())
	}
}
