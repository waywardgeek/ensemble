package engine

import (
	"encoding/json"
	"errors"

	"ensemble/internal/common"
)

// Wire structs stay private to this translator. RawMessage preserves opaque
// blocks without turning them into public context fields or relying on map
// iteration when constructing an ordered message.
type anthropicBlock struct {
	// Source encodes local image or PDF bytes as an Anthropic base64 source.
	Source any `json:"source,omitempty"`
	// Type selects the wire block variant before its other fields are interpreted.
	Type string `json:"type"`
	// Text holds visible provider or user content, including an explicitly empty string.
	Text string `json:"text,omitempty"`
	// ID is the provider-issued call identity, retained for its matching result.
	ID string `json:"id,omitempty"`
	// Name is the declared function identifier used to correlate calls and results.
	Name string `json:"name,omitempty"`
	// Input preserves the tool-use arguments as JSON for dispatch and replay.
	Input json.RawMessage `json:"input,omitempty"`
	// ToolUseID quotes the exact tool_use identity receiving this result.
	ToolUseID string `json:"tool_use_id,omitempty"`
	// Content preserves the provider's ordered blocks or message text.
	Content []json.RawMessage `json:"content,omitempty"`
	// IsError distinguishes ordinary tool failure from a successful result.
	IsError bool `json:"is_error,omitempty"`
}
type anthropicMessage struct {
	// Role is the provider spelling chosen from neutral authorship.
	Role string `json:"role"`
	// Content preserves the provider's ordered blocks or message text.
	Content []json.RawMessage `json:"content"`
}

// Anthropic keeps system instructions outside messages and gives tool results
// no role of their own. Those are rendering decisions; the log retains the
// distinct Human, Agent and Tool speakers.
func (e *engine) renderAnthropic(c *common.Context) ([]byte, error) {
	messages := []anthropicMessage{}
	for _, entry := range entries(c) {
		role := "user"
		if entry.Actor == common.AgentActor {
			role = "assistant"
		}
		blocks, err := e.anthropicParts(c, entry)
		if err != nil {
			return nil, err
		}
		if len(blocks) == 0 {
			continue
		}
		// Merge adjacent equal roles without changing the conversation itself.
		// User-block ordering is resolved after the complete merged turn exists.
		if len(messages) > 0 && messages[len(messages)-1].Role == role {
			messages[len(messages)-1].Content = append(messages[len(messages)-1].Content, blocks...)
		} else {
			messages = append(messages, anthropicMessage{role, blocks})
		}
	}
	// Results can arrive on either side of human text. Anthropic requires every
	// tool_result before all other content in the merged user message. Partition
	// only this wire view, preserving order within each group and all raw bytes.
	for i, message := range messages {
		if message.Role != "user" {
			continue
		}
		var results, other []json.RawMessage
		for _, raw := range message.Content {
			var block struct {
				// Type selects the wire block variant before its other fields are interpreted.
				Type string `json:"type"`
			}
			if err := json.Unmarshal(raw, &block); err != nil {
				return nil, errors.New("invalid Anthropic user block")
			}
			if block.Type == "tool_result" {
				results = append(results, raw)
			} else {
				other = append(other, raw)
			}
		}
		messages[i].Content = append(results, other...)
	}
	// Declaration spelling stays here beside the matching call/result spelling.
	var tools []map[string]any
	for _, d := range e.parent.Declarations() {
		tools = append(tools, map[string]any{"name": d.Name, "description": d.Description, "input_schema": d.Schema})
	}
	return json.Marshal(struct {
		// Tools contains only the declarations visible to this Agent.
		Tools []map[string]any `json:"tools,omitempty"`
		// Model selects or reports the provider model identity for this exchange.
		Model string `json:"model"`
		// MaxTokens bounds generated response tokens on this provider surface.
		MaxTokens int `json:"max_tokens"`
		// System carries the role instruction outside ordinary Anthropic messages.
		System string `json:"system"`
		// Messages retains provider message order, including tool-result boundaries.
		Messages []anthropicMessage `json:"messages"`
	}{tools, e.parent.Config().Model, 1024, e.systemPrompt(), messages})
}
func (e *engine) anthropicParts(c *common.Context, entry common.Entry) ([]json.RawMessage, error) {
	var out []json.RawMessage
	for i, p := range entry.Parts {
		if err := e.refuseBlob(p); err != nil {
			return nil, err
		}
		b := anthropicBlock{}
		switch p.Type {
		case "blob":
			data, err := e.blob(p)
			if err != nil {
				return nil, err
			}
			b.Type = "image"
			if p.MIME == "application/pdf" {
				b.Type = "document"
			}
			b.Source = map[string]string{"type": "base64", "media_type": p.MIME, "data": data}
		case "text":
			b.Type, b.Text = "text", p.Text
		case "redacted":
			b.Type, b.Text = "text", p.Stub
		case "tool_call":
			b.Type, b.ID, b.Name, b.Input = "tool_use", callID(entry, i, p), p.Name, p.Args
		// The result quotes exactly the ID sent for the corresponding call.
		// Nested text/stubs use the same block encoder as ordinary dialogue.
		case "tool_result":
			callEntry, j, call, err := callFor(c, p.CallID)
			if err != nil {
				return nil, err
			}
			b.Type, b.ToolUseID, b.IsError = "tool_result", callID(callEntry, j, call), p.IsError
			b.Content, err = e.anthropicParts(c, common.Entry{Parts: p.Parts})
			if err != nil {
				return nil, err
			}
		case "opaque":
			if p.From == e.Target() {
				out = append(out, p.Data)
			}
			continue
		default:
			return nil, errors.New("unsupported Anthropic part")
		}
		// Empty text is still a text block; it must not disappear into null content.
		if b.Type == "text" {
			raw, _ := json.Marshal(struct {
				// Type selects the wire block variant before its other fields are interpreted.
				Type string `json:"type"`
				// Text holds visible provider or user content, including an explicitly empty string.
				Text string `json:"text"`
			}{"text", b.Text})
			out = append(out, raw)
		} else {
			raw, err := json.Marshal(b)
			if err != nil {
				return nil, err
			}
			out = append(out, raw)
		}
	}
	return out, nil
}

// The producer comes from the response, not just the requested alias. Missing
// usage is not measured zero; required pointer counts make that distinction
// visible before a completed-response event can be emitted.
func (e *engine) parseAnthropic(data []byte) (common.ResponseData, error) {
	var wire struct {
		// Model selects or reports the provider model identity for this exchange.
		Model string `json:"model"`
		// Content preserves the provider's ordered blocks or message text.
		Content []json.RawMessage `json:"content"`
		// Usage must be present to distinguish measured counts from missing accounting.
		Usage *struct {
			// Input retains the provider's input count before cache normalization.
			Input *int `json:"input_tokens"`
			// Output retains the provider's generated-token count before normalization.
			Output *int `json:"output_tokens"`
			// Write records the provider's cache-creation token category.
			Write int `json:"cache_creation_input_tokens"`
			// Read records the provider's cached-input token category.
			Read int `json:"cache_read_input_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return common.ResponseData{}, errors.New("invalid Anthropic response")
	}
	if wire.Usage == nil || wire.Usage.Input == nil || wire.Usage.Output == nil {
		return common.ResponseData{}, errors.New("Anthropic response missing usage")
	}
	from := e.Target()
	if wire.Model != "" {
		from.Model = wire.Model
	}
	// Anthropic already excludes both cache categories from input_tokens.
	// Passing these four counts through is correct; subtracting cache again
	// would undercount a warm request.
	result := common.ResponseData{From: from, Usage: common.Usage{Input: *wire.Usage.Input, Output: *wire.Usage.Output, CacheWrite: wire.Usage.Write, CacheRead: wire.Usage.Read}}
	// Walk the blocks in provider order. Extracting text and calls in separate
	// passes would preserve their values but lose how the reply interleaved them.
	for _, raw := range wire.Content {
		var b anthropicBlock
		if err := json.Unmarshal(raw, &b); err != nil {
			return result, errors.New("invalid Anthropic block")
		}
		switch b.Type {
		case "text":
			result.Parts = append(result.Parts, common.Part{Type: "text", Text: b.Text})
		case "tool_use":
			args, err := canonical(b.Input)
			if err != nil {
				return result, err
			}
			result.Parts = append(result.Parts, common.Part{Type: "tool_call", CallID: b.ID, Name: b.Name, Args: args, From: from})
		case "thinking", "redacted_thinking":
			// The opaque wire block can be replayed only to its exact producer. It is
			// never interpreted as prose or displayed as an assistant answer.
			result.Parts = append(result.Parts, common.Part{Type: "opaque", From: from, Data: raw})
		default:
			return result, errors.New("unknown Anthropic response block")
		}
	}
	if len(result.Parts) == 0 {
		return result, errors.New("Anthropic response has no content")
	}
	return result, nil
}
