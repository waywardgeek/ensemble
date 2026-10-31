package engine

import (
	"encoding/json"
	"errors"

	"ensemble/internal/common"
)

type openAICall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// Pointer content distinguishes present-but-empty text from absent content.
// Tool-only assistant messages can omit it; a text reply of "" cannot be
// normalized into absence and later replayed as null.
type openAIMessage struct {
	Role       string       `json:"role"`
	Content    *string      `json:"content,omitempty"`
	ToolCalls  []openAICall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
}

// Chat Completions uses an explicit tool role and keeps results separate from
// the next human message. Its system instruction lives inside the ordered
// message list, unlike Anthropic's top-level system field.
func (e *engine) renderOpenAI(c *common.Context) ([]byte, error) {
	system := systemPrompt
	messages := []openAIMessage{{Role: "system", Content: &system}}
	for _, entry := range entries(c) {
		role := "user"
		if entry.Actor == common.AgentActor {
			role = "assistant"
		}
		content := ""
		message := openAIMessage{Role: role}
		hasText := false
		for i, p := range entry.Parts {
			if err := e.refuseBlob(p); err != nil {
				return nil, err
			}
			switch p.Type {
			case "text":
				content += p.Text
				hasText = true
			case "redacted":
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
		if hasText || len(message.ToolCalls) > 0 {
			messages = append(messages, message)
		}
	}
	return json.Marshal(struct {
		Model     string          `json:"model"`
		MaxTokens int             `json:"max_completion_tokens"`
		Messages  []openAIMessage `json:"messages"`
	}{e.parent.Config().Model, 1024, messages})
}

// This parser consumes the first response choice, matching one conversational
// turn. The exercise requests one reply and does not introduce alternative
// choices, server-side threads or response continuation IDs.
func (e *engine) parseOpenAI(data []byte) (common.ResponseData, error) {
	var wire struct {
		Model   string `json:"model"`
		Choices []struct {
			Message openAIMessage `json:"message"`
		} `json:"choices"`
		Usage *struct {
			Input   *int `json:"prompt_tokens"`
			Output  *int `json:"completion_tokens"`
			Details struct {
				Read  int `json:"cached_tokens"`
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
	from := e.target()
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
