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
	Type      string            `json:"type"`
	Text      string            `json:"text,omitempty"`
	ID        string            `json:"id,omitempty"`
	Name      string            `json:"name,omitempty"`
	Input     json.RawMessage   `json:"input,omitempty"`
	ToolUseID string            `json:"tool_use_id,omitempty"`
	Content   []json.RawMessage `json:"content,omitempty"`
	IsError   bool              `json:"is_error,omitempty"`
}
type anthropicMessage struct {
	Role    string            `json:"role"`
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
	return json.Marshal(struct {
		Model     string             `json:"model"`
		MaxTokens int                `json:"max_tokens"`
		System    string             `json:"system"`
		Messages  []anthropicMessage `json:"messages"`
	}{e.parent.Config().Model, 1024, systemPrompt, messages})
}
func (e *engine) anthropicParts(c *common.Context, entry common.Entry) ([]json.RawMessage, error) {
	var out []json.RawMessage
	for i, p := range entry.Parts {
		if err := e.refuseBlob(p); err != nil {
			return nil, err
		}
		b := anthropicBlock{}
		switch p.Type {
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
			if p.From == e.target() {
				out = append(out, p.Data)
			}
			continue
		default:
			return nil, errors.New("unsupported Anthropic part")
		}
		// Empty text is still a text block; it must not disappear into null content.
		if b.Type == "text" {
			raw, _ := json.Marshal(struct {
				Type string `json:"type"`
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
		Model   string            `json:"model"`
		Content []json.RawMessage `json:"content"`
		Usage   *struct {
			Input  *int `json:"input_tokens"`
			Output *int `json:"output_tokens"`
			Write  int  `json:"cache_creation_input_tokens"`
			Read   int  `json:"cache_read_input_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return common.ResponseData{}, errors.New("invalid Anthropic response")
	}
	if wire.Usage == nil || wire.Usage.Input == nil || wire.Usage.Output == nil {
		return common.ResponseData{}, errors.New("Anthropic response missing usage")
	}
	from := e.target()
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
