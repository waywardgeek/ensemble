package llm

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"mime"
	"strings"
)

type renderer struct {
	parent common.Engine
	target common.Provenance
	model  string
	err    error
}

func (r *renderer) Engine() common.Engine { return r.parent }
func (r *renderer) fail(message string) {
	if r.err == nil {
		r.err = failure(r.parent, "%s", message)
	}
}
func (r *renderer) match(p *common.Provenance) bool { return p != nil && *p == r.target }
func (r *renderer) resultText(parts []common.Part) string {
	lines := []string{}
	for _, p := range parts {
		switch p.Type {
		case "text":
			lines = append(lines, *p.Text)
		case "blob":
			if p.Ref.Kind != 2 {
				r.unsupportedReference(p)
			}
			lines = append(lines, "["+p.MIME+"] "+p.Ref.Locator)
		case "redacted":
			text := p.Stub
			if p.Ref != nil {
				text += " " + p.Ref.Locator
			}
			lines = append(lines, text)
		}
	}
	return strings.Join(lines, "\n")
}
func Render(owner common.Engine, c common.Context, config common.Config) ([]byte, error) {
	if err := ValidateConfig(owner, config, false); err != nil {
		return nil, err
	}
	if unresolved(owner, c) {
		return nil, failure(owner, "unanswered tool calls require results before rendering")
	}
	target := Route(owner, config)
	if config.ResolvedModel != "" {
		target.Model = config.ResolvedModel
	}
	r := renderer{parent: owner, target: target, model: config.Model}
	system := []string{config.System}
	for _, entry := range c.Instructions {
		for _, part := range entry.Parts {
			if part.Text != nil {
				system = append(system, *part.Text)
			}
		}
	}
	entries := append([]common.Entry(nil), c.Entries...)
	if c.Pending != nil {
		entries = append(entries, *c.Pending)
	}
	for _, entry := range c.Ephemera {
		entry.Actor = "human"
		entries = append(entries, entry)
	}
	entries = append(entries, c.Hints...)
	var body any
	switch config.Vendor {
	case "anthropic", "gemini":
		messages := []map[string]any{}
		key := "content"
		if config.Vendor == "gemini" {
			key = "parts"
		}
		for _, entry := range entries {
			role := "user"
			if entry.Actor == "agent" {
				role = "assistant"
				if config.Vendor == "gemini" {
					role = "model"
				}
			}
			blocks := []any{}
			for _, part := range entry.Parts {
				if block := r.block(part, config.Vendor, c); block != nil {
					blocks = append(blocks, block)
				}
			}
			if len(blocks) == 0 {
				continue
			}
			if len(messages) > 0 && messages[len(messages)-1]["role"] == role {
				last := messages[len(messages)-1]
				last[key] = append(last[key].([]any), blocks...)
			} else {
				messages = append(messages, map[string]any{"role": role, key: blocks})
			}
		}
		if config.Vendor == "anthropic" {
			m := map[string]any{"model": config.Model, "max_tokens": config.MaxTokens, "system": strings.Join(system, "\n"), "messages": messages}
			if len(config.Tools) > 0 {
				m["tools"] = config.Tools
			}
			body = m
		} else {
			m := map[string]any{"systemInstruction": map[string]any{"parts": []any{map[string]any{"text": strings.Join(system, "\n")}}}, "contents": messages, "generationConfig": map[string]any{"maxOutputTokens": config.MaxTokens}}
			if len(config.Tools) > 0 {
				tools := []any{}
				for _, tool := range config.Tools {
					// parameters is an OpenAPI subset; preserve the caller's JSON Schema here.
					tools = append(tools, map[string]any{"name": tool.Name, "description": tool.Description, "parametersJsonSchema": tool.Schema})
				}
				m["tools"] = []any{map[string]any{"functionDeclarations": tools}}
			}
			body = m
		}
	case "openai":
		messages := []any{map[string]any{"role": "system", "content": strings.Join(system, "\n")}}
		for _, entry := range entries {
			if entry.Actor == "tool" {
				for _, part := range entry.Parts {
					text := r.resultText(part.Parts)
					if part.IsError {
						text = "Tool failed: " + text
					}
					messages = append(messages, map[string]any{"role": "tool", "tool_call_id": part.CallID, "content": text})
				}
				continue
			}
			role := "user"
			if entry.Actor == "agent" {
				role = "assistant"
			}
			text := []string{}
			calls := []any{}
			for _, part := range entry.Parts {
				switch part.Type {
				case "text":
					text = append(text, *part.Text)
				case "redacted":
					s := part.Stub
					if part.Ref != nil {
						s += " " + part.Ref.Locator
					}
					text = append(text, s)
				case "tool_call":
					if len(part.Opaque) > 0 && !r.match(part.From) {
						r.fail("incompatible call-bound replay material")
					}
					calls = append(calls, map[string]any{"id": part.CallID, "type": "function", "function": map[string]any{"name": part.Name, "arguments": string(part.Args)}})
				case "blob":
					r.unsupportedReference(part)
				case "opaque":
					if r.match(part.From) {
						r.fail("unsupported opaque Chat Completions material")
					}
				}
			}
			if len(text) == 0 && len(calls) == 0 {
				continue
			}
			m := map[string]any{"role": role, "content": strings.Join(text, "")}
			if len(calls) > 0 {
				m["tool_calls"] = calls
			}
			messages = append(messages, m)
		}
		m := map[string]any{"model": config.Model, "max_completion_tokens": config.MaxTokens, "messages": messages}
		if len(config.Tools) > 0 {
			tools := []any{}
			for _, tool := range config.Tools {
				tools = append(tools, map[string]any{"type": "function", "function": map[string]any{"name": tool.Name, "description": tool.Description, "parameters": tool.Schema}})
			}
			m["tools"] = tools
		}
		body = m
	}
	if r.err != nil {
		return nil, r.err
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, failure(owner, "cannot encode request")
	}
	return data, nil
}
func (r *renderer) block(p common.Part, vendor string, c common.Context) any {
	if p.Type == "opaque" {
		if r.match(p.From) {
			return p.Data
		}
		return nil
	}
	if p.Type == "tool_call" && len(p.Opaque) > 0 && !r.match(p.From) {
		r.fail("incompatible call-bound replay material")
		return nil
	}
	if vendor == "anthropic" {
		switch p.Type {
		case "text":
			return map[string]any{"type": "text", "text": *p.Text}
		case "tool_call":
			return map[string]any{"type": "tool_use", "id": p.CallID, "name": p.Name, "input": p.Args}
		case "tool_result":
			return map[string]any{"type": "tool_result", "tool_use_id": p.CallID, "content": r.resultText(p.Parts), "is_error": p.IsError}
		case "redacted":
			s := p.Stub
			if p.Ref != nil {
				s += " " + p.Ref.Locator
			}
			return map[string]any{"type": "text", "text": s}
		case "blob":
			r.unsupportedReference(p)
		}
		return nil
	}
	var out map[string]any
	switch p.Type {
	case "text":
		out = map[string]any{"text": *p.Text}
	case "tool_call":
		out = map[string]any{"functionCall": map[string]any{"id": p.CallID, "name": p.Name, "args": p.Args}}
	case "tool_result":
		call := c.Calls[p.CallID]
		key := "result"
		if p.IsError {
			key = "error"
		}
		out = map[string]any{"functionResponse": map[string]any{"id": p.CallID, "name": call.Part.Name, "response": map[string]any{key: r.resultText(p.Parts)}}}
	case "blob":
		if p.Ref.Kind != 2 {
			r.unsupportedReference(p)
			return nil
		}
		out = map[string]any{"fileData": map[string]any{"mimeType": p.MIME, "fileUri": p.Ref.Locator}}
	case "redacted":
		s := p.Stub
		if p.Ref != nil {
			s += " " + p.Ref.Locator
		}
		out = map[string]any{"text": s}
	}
	if out != nil && len(p.Opaque) > 0 && r.match(p.From) {
		out["thoughtSignature"] = p.Opaque
	}
	return out
}

// Diagnostics identify the missing mapping, never the retained private locator.
func (r *renderer) unsupportedReference(p common.Part) {
	media, _, err := mime.ParseMediaType(p.MIME)
	if err != nil {
		media = "invalid media MIME"
	}
	kind := "reference"
	if p.Ref != nil {
		switch p.Ref.Kind {
		case 1:
			kind = "path"
		case 2:
			kind = "URI"
		case 3:
			kind = "handle"
		}
	}
	r.fail(fmt.Sprintf("unsupported reference: model %q has no %s %s mapping for %s", r.model, r.target.Surface, kind, media))
}
