package engine

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"ensemble/internal/common"
)

// Responses sends complete local history, never provider-owned conversation IDs.
// Cache markers live on content blocks, so even the constitution is an item.
func (e *engine) renderResponses(c *common.Context) ([]byte, error) {
	input := []map[string]any{}
	message := func(role, kind, value string) map[string]any {
		return map[string]any{"type": "message", "role": role, "content": []map[string]any{{"type": kind, "text": value}}}
	}
	input = append(input, message("developer", "input_text", e.SystemPrompt()))
	markable := func() int {
		for i := len(input) - 1; i >= 0; i-- {
			if _, ok := input[i]["content"].([]map[string]any); ok {
				return i
			}
		}
		return -1
	}
	bound, stable, checkpoint := compactionBound(c), -1, -1
	for index, entry := range entries(c) {
		role, kind := "user", "input_text"
		if entry.Kind == common.KindSkill {
			role = "developer"
		}
		if entry.Actor == common.AgentActor {
			role, kind = "assistant", "output_text"
		}
		for i, p := range entry.Parts {
			switch p.Type {
			case "text":
				input = append(input, message(role, kind, p.Text))
			case "redacted":
				input = append(input, message(role, kind, p.Stub))
			case "blob":
				data, err := e.blob(p)
				if err != nil {
					return nil, err
				}
				block := map[string]any{"type": "input_file", "filename": filepath.Base(p.Ref.Locator), "file_data": "data:" + p.MIME + ";base64," + data}
				if strings.HasPrefix(p.MIME, "image/") {
					block = map[string]any{"type": "input_image", "image_url": "data:" + p.MIME + ";base64," + data}
				}
				input = append(input, map[string]any{"type": "message", "role": role, "content": []map[string]any{block}})
			case "tool_call":
				item := map[string]any{"type": "function_call", "call_id": callID(entry, i, p), "name": p.Name, "arguments": string(p.Args)}
				if p.Namespace != "" {
					item["namespace"] = p.Namespace
				}
				input = append(input, item)
			case "tool_result":
				for _, part := range p.Parts {
					if part.Type == "blob" {
						return nil, errors.New("Responses tool results do not support media on this renderer")
					}
				}
				input = append(input, map[string]any{"type": "function_call_output", "call_id": p.CallID, "output": text(p.Parts)})
			case "opaque":
				// Encrypted reasoning is valid only for its original model and surface.
				if p.From == e.Target() {
					var item map[string]any
					if json.Unmarshal(p.Data, &item) != nil {
						return nil, errors.New("invalid Responses replay item")
					}
					input = append(input, item)
				}
			case "tool_decls":
			default:
				return nil, errors.New("unsupported Responses part")
			}
		}
		if index < len(c.Dialogue) {
			stable = markable()
			if index == bound {
				checkpoint = stable
			}
		}
	}
	request := map[string]any{"model": e.parent.Config().Model, "input": input, "stream": true, "store": false, "include": []string{"reasoning.encrypted_content"}}
	reasoning := map[string]string{"summary": "auto"}
	cfg := e.parent.Config()
	if cfg.ReasoningSummary != "" {
		reasoning["summary"] = cfg.ReasoningSummary
	}
	if cfg.ReasoningEffort != "" {
		reasoning["effort"] = cfg.ReasoningEffort
	}
	request["reasoning"] = reasoning
	var tools []map[string]any
	for _, d := range e.declarations(c) {
		tools = append(tools, map[string]any{"type": "function", "name": d.Name, "description": d.Description, "parameters": d.Schema, "strict": false})
	}
	caps := capabilitiesFor(cfg)
	if cfg.Credentials != nil && cfg.Credentials.Kind() == common.CredentialCodex {
		// Codex requires the instruction field. Keep the same instruction-bearing
		// input as public Responses so ordinary conversation replay is unchanged.
		request["instructions"] = e.SystemPrompt()
		request["prompt_cache_key"] = codexCacheKey(cfg)
	}
	if len(tools) > 0 {
		request["tools"] = tools
		if caps.namespace != "" {
			request["tools"] = []any{map[string]any{"type": "namespace", "name": caps.namespace, "description": "Ensemble local tools", "tools": tools}}
		}
	}
	if e.Features().ExplicitCache && !caps.omitExplicitCache {
		landed := false
		for _, i := range []int{0, checkpoint, stable} {
			if i < 0 {
				continue
			}
			blocks, ok := input[i]["content"].([]map[string]any)
			if !ok || len(blocks) == 0 {
				continue
			}
			last := blocks[len(blocks)-1]
			if value, ok := last["text"].(string); ok && value == "" {
				continue
			}
			last["prompt_cache_breakpoint"] = map[string]string{"mode": "explicit"}
			landed = true
		}
		if landed {
			request["prompt_cache_options"] = map[string]string{"mode": "explicit"}
		}
	}
	for _, field := range caps.omit {
		delete(request, field)
	}
	return json.Marshal(request)
}
