package engine

import (
	"bytes"
	"encoding/json"
	"ensemble/internal/common"
)

// The newest surviving Chapter14 handoff is the settled context boundary.
// Mark the retained projection; rendering never recreates removed tool traffic.
func compactionBound(c *common.Context) int {
	boundary := -1
	for i, e := range c.Dialogue {
		if e.Kind == common.KindHandoff {
			boundary = i
		}
	}
	return boundary
}

// Mark only provider-supported content blocks, excluding private thinking and
// inline declaration traffic. Raw replay blocks remain otherwise untouched.
func markAnthropic(blocks []json.RawMessage) bool {
	for i := len(blocks) - 1; i >= 0; i-- {
		var block map[string]json.RawMessage
		if json.Unmarshal(blocks[i], &block) != nil {
			continue
		}
		var kind string
		json.Unmarshal(block["type"], &kind)
		switch kind {
		case "text", "image", "document", "tool_use", "tool_result":
			if _, exists := block["cache_control"]; exists {
				return true
			}
			// Add metadata without reordering the preexisting content keys.
			raw := bytes.TrimSpace(blocks[i])
			blocks[i] = append(append([]byte(nil), raw[:len(raw)-1]...), []byte(`,"cache_control":{"type":"ephemeral"}}`)...)
			return true
		}
	}
	return false
}

// OpenAI breakpoints belong to content parts, never the tools array. Returning
// evidence of a landed marker prevents explicit mode from disabling implicit
// caching when no markable content exists.
func markOpenAI(message *openAIMessage) bool {
	var blocks []map[string]any
	switch content := message.Content.(type) {
	case *string:
		if content == nil || *content == "" {
			return false
		}
		blocks = []map[string]any{{"type": "text", "text": *content}}
	case []map[string]any:
		blocks = content
		if len(blocks) > 0 && blocks[len(blocks)-1]["type"] == "text" && blocks[len(blocks)-1]["text"] == "" {
			return false
		}
	default:
		return false
	}
	if len(blocks) == 0 {
		return false
	}
	blocks[len(blocks)-1]["prompt_cache_breakpoint"] = map[string]string{"mode": "explicit"}
	message.Content = blocks
	return true
}
