package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"ensemble/internal/common"
)

func responsesEndpoint(base string, mode common.AuthMode) string {
	if base == "" {
		return "https://api.openai.com/v1/responses"
	}
	// Caller-provided bases and full URLs stay on the caller's host/path. A full
	// Responses endpoint is never doubled and no credentials select another host.
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/responses") {
		return base
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/responses"
	}
	return base + "/v1/responses"
}
func renderResponses(c common.Context, cfg common.Config, defs []common.ToolDefinition) ([]byte, error) {
	instructions, err := system(c, cfg)
	if err != nil {
		return nil, err
	}
	input := []any{}
	results := map[string]common.Part{}
	calls := map[string]bool{}
	for _, entry := range c.Dialogue {
		for _, p := range entry.Parts {
			switch p.Type {
			case "tool_call":
				if p.CallID == "" || calls[p.CallID] {
					return nil, fmt.Errorf("ambiguous Responses replay call identity")
				}
				calls[p.CallID] = true
			case "tool_result":
				if !calls[p.CallID] {
					return nil, fmt.Errorf("orphan Responses replay result")
				}
				if _, exists := results[p.CallID]; exists {
					return nil, fmt.Errorf("ambiguous Responses replay result")
				}
				results[p.CallID] = p
			}
		}
	}
	message := func(role, text string) {
		kind := "input_text"
		if role == "assistant" {
			kind = "output_text"
		}
		input = append(input, map[string]any{"role": role, "content": []any{map[string]any{"type": kind, "text": text}}})
	}
	for _, entry := range c.Dialogue {
		if entry.Actor == common.System {
			continue
		}
		role := "user"
		if entry.Actor == common.Assistant {
			role = "assistant"
		}
		replayReasoning := false
		for _, p := range entry.Parts {
			if p.Type == "opaque" && matches(p.From, cfg) {
				replayReasoning = true
			}
		}
		for _, p := range entry.Parts {
			switch p.Type {
			case "text", "thinking":
				message(role, p.Text)
			case "reasoning_summary":
				if !replayReasoning {
					message(role, p.Text)
				}
			case "redacted":
				message(role, p.Stub)
			case "image", "blob":
				input = append(input, map[string]any{"role": role, "content": []any{map[string]any{"type": "input_image", "image_url": "data:" + p.MIME + ";base64," + mediaData(p)}}})
			case "opaque":
				if matches(p.From, cfg) {
					var item responseItem
					if json.Unmarshal(p.Data, &item) != nil || item.Type != "reasoning" || item.Encrypted == "" {
						return nil, fmt.Errorf("invalid Responses continuation")
					}
					input = append(input, json.RawMessage(append([]byte(nil), p.Data...)))
				}
			case "tool_call":
				if role != "assistant" {
					return nil, fmt.Errorf("tool call must be agent-authored")
				}
				args, err := canonical(p.Args)
				if err != nil {
					return nil, err
				}
				call := map[string]any{"type": "function_call", "call_id": p.CallID, "name": p.Name, "arguments": string(args)}
				if p.Namespace != "" {
					call["namespace"] = p.Namespace
				}
				input = append(input, call)
				output := "No result was recorded; execution outcome unknown."
				if result, ok := results[p.CallID]; ok {
					output, err = textParts(result.Parts)
					if err != nil {
						return nil, err
					}
					if result.IsError {
						output = "Tool error: " + output
					}
				}
				input = append(input, map[string]any{"type": "function_call_output", "call_id": p.CallID, "output": output})
			case "tool_delta": // The complete current inventory is sent below.
			case "tool_result": // Adjacent to its actual call, not its arrival position.
			default:
				return nil, fmt.Errorf("unsupported Responses part %q", p.Type)
			}
		}
	}
	if len(c.Ephemera) > 0 {
		handoff, tail, err := projectedTail(c)
		if err != nil {
			return nil, err
		}
		message("user", handoff+tail)
	}
	body := map[string]any{"model": cfg.Model, "instructions": instructions, "input": input, "store": false, "stream": cfg.Streaming(), "include": []string{"reasoning.encrypted_content"}}
	tools := []any{}
	for _, d := range defs {
		schema, err := canonical(d.Parameters)
		if err != nil {
			return nil, err
		}
		tools = append(tools, map[string]any{"type": "function", "name": d.Name, "description": d.Description, "parameters": schema})
	}
	if len(tools) > 0 {
		if cfg.AuthMode == common.AuthChatGPT {
			body["tools"] = []any{map[string]any{"type": "namespace", "name": "ensemble", "description": "Ensemble local tools", "tools": tools}}
		} else {
			body["tools"] = tools
		}
	}
	effort, summary := common.ResponsesCapabilities(cfg)
	reasoning := map[string]string{}
	if effort && cfg.ReasoningEffort != "" {
		reasoning["effort"] = cfg.ReasoningEffort
	}
	if summary && cfg.ReasoningSummary != "" && cfg.ReasoningSummary != "none" {
		reasoning["summary"] = cfg.ReasoningSummary
	}
	if len(reasoning) > 0 {
		body["reasoning"] = reasoning
	}
	if cfg.AuthMode == common.AuthChatGPT {
		body["stream"] = true
	} else {
		decorateResponses(body, c, cfg)
	}
	return json.Marshal(body)
}

type responseContent struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Refusal string `json:"refusal,omitempty"`
}
type responseItem struct {
	Type      string            `json:"type"`
	ID        string            `json:"id,omitempty"`
	Role      string            `json:"role,omitempty"`
	Status    string            `json:"status,omitempty"`
	CallID    string            `json:"call_id,omitempty"`
	Name      string            `json:"name,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Arguments string            `json:"arguments,omitempty"`
	Content   []responseContent `json:"content,omitempty"`
	Summary   []responseContent `json:"summary,omitempty"`
	Encrypted string            `json:"encrypted_content,omitempty"`
}

func responseKey(output int, kind string, part int) string {
	return fmt.Sprintf("%d/%s/%d", output, kind, part)
}
func responsesParts(b []byte, from common.Provenance, outputIndices []int) (common.PartList, []string, error) {
	var envelope struct {
		Status string            `json:"status"`
		Output []json.RawMessage `json:"output"`
		Error  json.RawMessage   `json:"error"`
	}
	if json.Unmarshal(b, &envelope) != nil {
		return nil, nil, fmt.Errorf("decoding Responses output failed")
	}
	if envelope.Status != "completed" || (len(envelope.Error) > 0 && string(envelope.Error) != "null") {
		return nil, nil, fmt.Errorf("Responses request did not complete")
	}
	parts := common.PartList{}
	keys := []string{}
	add := func(p common.Part, key string) { p.From = &from; parts = append(parts, p); keys = append(keys, key) }
	calls := map[string]bool{}
	for i, itemRaw := range envelope.Output {
		if outputIndices != nil {
			i = outputIndices[i]
		}
		var item responseItem
		if err := json.Unmarshal(itemRaw, &item); err != nil {
			return nil, nil, fmt.Errorf("invalid Responses output item")
		}
		switch item.Type {
		case "message":
			if item.Role != "assistant" || (item.Status != "" && item.Status != "completed") {
				return nil, nil, fmt.Errorf("invalid Responses message")
			}
			for j, content := range item.Content {
				text := content.Text
				if content.Type == "refusal" {
					text = content.Refusal
				} else if content.Type != "output_text" {
					return nil, nil, fmt.Errorf("unsupported Responses message content")
				}
				add(common.Text(text), responseKey(i, "text", j))
			}
		case "function_call":
			args := json.RawMessage(item.Arguments)
			if (item.Namespace != "" && item.Namespace != "ensemble") || item.CallID == "" || calls[item.CallID] || item.Name == "" || !json.Valid(args) || (item.Status != "" && item.Status != "completed") {
				return nil, nil, fmt.Errorf("invalid Responses function call")
			}
			calls[item.CallID] = true
			add(common.Part{Type: "tool_call", CallID: item.CallID, Name: item.Name, Namespace: item.Namespace, Args: args}, responseKey(i, "tool", 0))
		case "reasoning":
			for j, summary := range item.Summary {
				if summary.Type != "summary_text" {
					return nil, nil, fmt.Errorf("invalid Responses reasoning summary")
				}
				add(common.Part{Type: "reasoning_summary", Text: summary.Text}, responseKey(i, "summary", j))
			}
			if item.Encrypted != "" {
				if item.ID == "" {
					return nil, nil, fmt.Errorf("Responses continuation missing identity")
				}
				add(common.Part{Type: "opaque", Data: append(json.RawMessage(nil), itemRaw...), From: &from}, responseKey(i, "opaque", 0))
			}
		default:
			return nil, nil, fmt.Errorf("unsupported Responses output item")
		}
	}
	return parts, keys, nil
}
func parseResponses(b []byte, from common.Provenance) (common.PartList, error) {
	p, _, err := responsesParts(b, from, nil)
	return p, err
}

// The retained recovery part may contain encrypted continuation; diagnostic
// copies may not. This projection never changes the actual provider request.
func responsesDiagnostic(body []byte) []byte {
	var object map[string]json.RawMessage
	if json.Unmarshal(body, &object) != nil {
		return []byte(`{"diagnostic":"unavailable"}`)
	}
	var items []map[string]json.RawMessage
	if json.Unmarshal(object["input"], &items) == nil {
		for _, item := range items {
			if _, ok := item["encrypted_content"]; ok {
				item["encrypted_content"] = json.RawMessage(`"[private continuation retained]"`)
			}
		}
		object["input"], _ = json.Marshal(items)
	}
	b, err := json.Marshal(object)
	if err != nil {
		return []byte(`{"diagnostic":"unavailable"}`)
	}
	return b
}
