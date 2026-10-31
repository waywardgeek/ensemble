package engine

import (
	"encoding/json"
	"errors"

	"ensemble/internal/common"
)

type openAICall struct {
	// ID is the provider-issued call identity, retained for its matching result.
	ID string `json:"id"`
	// Type selects the wire block variant before its other fields are interpreted.
	Type string `json:"type"`
	// Function wraps the function name and encoded argument string.
	Function struct {
		// Name is the declared function identifier used to correlate calls and results.
		Name string `json:"name"`
		// Arguments contains JSON encoded as a string on Chat Completions.
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// Pointer content distinguishes present-but-empty text from absent content.
// Tool-only assistant messages can omit it; a text reply of "" cannot be
// normalized into absence and later replayed as null.
type openAIMessage struct {
	// Role is the provider spelling chosen from neutral authorship.
	Role string `json:"role"`
	// Content preserves the provider's ordered blocks or message text.
	Content any `json:"content,omitempty"`
	// ToolCalls retains all calls in their provider response order.
	ToolCalls []openAICall `json:"tool_calls,omitempty"`
	// ToolCallID quotes the exact call identity on an OpenAI tool message.
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// Chat Completions uses an explicit tool role and keeps results separate from
// the next human message. Its system instruction lives inside the ordered
// message list, unlike Anthropic's top-level system field.
func (e *engine) renderOpenAI(c *common.Context) ([]byte, error) {
	system := e.systemPrompt()
	messages := []openAIMessage{{Role: "system", Content: &system}}
	for _, entry := range entries(c) {
		role := "user"
		if entry.Actor == common.AgentActor {
			role = "assistant"
		}
		content := ""
		message := openAIMessage{Role: role}
		hasText := false
		var media []map[string]any
		for _, p := range entry.Parts {
			if p.Type == "blob" {
				media = []map[string]any{}
				break
			}
		}
		for i, p := range entry.Parts {
			if err := e.refuseBlob(p); err != nil {
				return nil, err
			}
			switch p.Type {
			case "blob":
				block, err := e.openAIBlob(p)
				if err != nil {
					return nil, err
				}
				media = append(media, block)
			case "text":
				if media != nil {
					media = append(media, map[string]any{"type": "text", "text": p.Text})
				}
				content += p.Text
				hasText = true
			case "redacted":
				if media != nil {
					media = append(media, map[string]any{"type": "text", "text": p.Stub})
				}
				content += p.Stub
				hasText = true
			// Arguments are already canonical JSON in the context. Only this
			// surface requires encoding those bytes as a string field.
			case "tool_call":
				call := openAICall{ID: callID(entry, i, p), Type: "function"}
				call.Function.Name, call.Function.Arguments = p.Name, string(p.Args)
				message.ToolCalls = append(message.ToolCalls, call)
			case "tool_result":
				callEntry, j, call, err := callFor(c, p.CallID)
				if err != nil {
					return nil, err
				}
				for _, part := range p.Parts {
					if part.Type == "blob" {
						return nil, errors.New("OpenAI tool results do not support media on this renderer")
					}
					if err := e.refuseBlob(part); err != nil {
						return nil, err
					}
				}
				value := text(p.Parts)
				messages = append(messages, openAIMessage{Role: "tool", Content: &value, ToolCallID: callID(callEntry, j, call)})
			case "opaque": // Chat Completions has no opaque replay field on this surface.
			default:
				return nil, errors.New("unsupported OpenAI part")
			}
		}
		// A legitimate empty string stays present; omitted content is reserved for
		// tool-only turns. Neither case is serialized as a bare assistant null.
		if hasText {
			message.Content = &content
		}
		if media != nil {
			message.Content = media
		}
		if hasText || len(message.ToolCalls) > 0 || media != nil {
			messages = append(messages, message)
		}
	}
	var tools []map[string]any
	for _, d := range e.parent.Declarations() {
		tools = append(tools, map[string]any{"type": "function", "function": map[string]any{"name": d.Name, "description": d.Description, "parameters": d.Schema}})
	}
	return json.Marshal(struct {
		// Tools contains only the declarations visible to this Agent.
		Tools []map[string]any `json:"tools,omitempty"`
		// Model selects or reports the provider model identity for this exchange.
		Model string `json:"model"`
		// MaxTokens bounds generated response tokens on this provider surface.
		MaxTokens int `json:"max_completion_tokens"`
		// Messages retains provider message order, including tool-result boundaries.
		Messages []openAIMessage `json:"messages"`
	}{tools, e.parent.Config().Model, 1024, messages})
}

// This parser consumes the first response choice, matching one conversational
// turn. The exercise requests one reply and does not introduce alternative
// choices, server-side threads or response continuation IDs.
func (e *engine) parseOpenAI(data []byte) (common.ResponseData, error) {
	var wire struct {
		// Model selects or reports the provider model identity for this exchange.
		Model string `json:"model"`
		// Choices holds provider alternatives; this client consumes the first.
		Choices []struct {
			// Message contains the first choice's text and function calls.
			Message struct {
				// Content preserves the provider's ordered blocks or message text.
				Content *string `json:"content"`
				// ToolCalls retains all calls in their provider response order.
				ToolCalls []openAICall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		// Usage must be present to distinguish measured counts from missing accounting.
		Usage *struct {
			// Input retains the provider's input count before cache normalization.
			Input *int `json:"prompt_tokens"`
			// Output retains the provider's generated-token count before normalization.
			Output *int `json:"completion_tokens"`
			// Details separates cache subsets from the inclusive prompt-token count.
			Details struct {
				// Read records the provider's cached-input token category.
				Read int `json:"cached_tokens"`
				// Write records the provider's cache-creation token category.
				Write int `json:"cache_write_tokens"`
			} `json:"prompt_tokens_details"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return common.ResponseData{}, errors.New("invalid OpenAI response")
	}
	if wire.Usage == nil || wire.Usage.Input == nil || wire.Usage.Output == nil {
		return common.ResponseData{}, errors.New("OpenAI response missing usage")
	}
	if len(wire.Choices) == 0 {
		return common.ResponseData{}, errors.New("OpenAI response missing choice")
	}
	from := e.Target()
	if wire.Model != "" {
		from.Model = wire.Model
	}
	// Cached/read and cache-write tokens are subsets of prompt_tokens here.
	// Subtract both to obtain ordinary input; completion_tokens already
	// includes reasoning, so no second output addition belongs here.
	u := wire.Usage
	result := common.ResponseData{From: from, Usage: common.Usage{Input: *u.Input - u.Details.Read - u.Details.Write, CacheRead: u.Details.Read, CacheWrite: u.Details.Write, Output: *u.Output}}
	// Pointer presence preserves "" separately from null/absent tool-only content.
	message := wire.Choices[0].Message
	if message.Content != nil {
		result.Parts = append(result.Parts, common.Part{Type: "text", Text: *message.Content})
	}
	for _, call := range message.ToolCalls {
		args, err := canonical(json.RawMessage(call.Function.Arguments))
		if err != nil {
			return result, err
		}
		result.Parts = append(result.Parts, common.Part{Type: "tool_call", CallID: call.ID, Name: call.Function.Name, Args: args, From: from})
	}
	if len(result.Parts) == 0 {
		return result, errors.New("OpenAI response has no content")
	}
	return result, nil
}
