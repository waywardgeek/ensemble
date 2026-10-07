package llm

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
)

type parser struct {
	parent common.Engine
	err    error
}

func (p *parser) Engine() common.Engine { return p.parent }
func (p *parser) invalid() {
	if p.err == nil {
		p.err = failure(p.parent, "malformed model response or usage")
	}
}
func (p *parser) obj(raw json.RawMessage) map[string]json.RawMessage {
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) != nil || m == nil {
		p.invalid()
	}
	return m
}
func (p *parser) str(raw json.RawMessage) string {
	var s string
	if len(raw) == 0 || string(raw) == "null" || json.Unmarshal(raw, &s) != nil {
		p.invalid()
	}
	return s
}
func (p *parser) list(raw json.RawMessage) []json.RawMessage {
	var a []json.RawMessage
	if json.Unmarshal(raw, &a) != nil || a == nil {
		p.invalid()
	}
	return a
}
func (p *parser) count(m map[string]json.RawMessage, key string, required bool) int64 {
	raw, ok := m[key]
	if !ok && !required {
		return 0
	}
	var n int64
	if !ok || string(raw) == "null" || json.Unmarshal(raw, &n) != nil || n < 0 {
		p.invalid()
	}
	return n
}
func (p *parser) truth(raw json.RawMessage) bool {
	var b bool
	if string(raw) == "null" || json.Unmarshal(raw, &b) != nil {
		p.invalid()
	}
	return b
}

// Parse returns facts only. The Agent's durable append path owns all mutation.
func Parse(owner common.Engine, config common.Config, body []byte, responseSeq uint64) (common.Response, error) {
	p := parser{parent: owner}
	root := p.obj(body)
	requested := Route(owner, config)
	from := requested
	reported := false
	modelField := "model"
	if config.Vendor == "gemini" {
		modelField = "modelVersion"
	}
	if raw, ok := root[modelField]; ok {
		from.Model = p.str(raw)
		reported = true
		if from.Model == "" {
			p.invalid()
		}
	}
	out := common.Response{From: from, Requested: &requested, ModelReported: &reported, Parts: []common.Part{}}
	usageField := "usage"
	if config.Vendor == "gemini" {
		usageField = "usageMetadata"
	}
	out.RawUsage = append(json.RawMessage(nil), root[usageField]...)
	u := p.obj(out.RawUsage)
	usage := common.Usage{}
	switch config.Vendor {
	case "anthropic":
		usage = common.Usage{Input: p.count(u, "input_tokens", true), Output: p.count(u, "output_tokens", true), CacheWrite: p.count(u, "cache_creation_input_tokens", false), CacheRead: p.count(u, "cache_read_input_tokens", false)}
		for _, raw := range p.list(root["content"]) {
			b := p.obj(raw)
			kind := p.str(b["type"])
			switch kind {
			case "text":
				s := p.str(b["text"])
				out.Parts = append(out.Parts, Text(s))
			case "tool_use":
				out.Parts = append(out.Parts, common.Part{Type: "tool_call", CallID: p.str(b["id"]), From: &from, Name: p.str(b["name"]), Args: append(json.RawMessage(nil), b["input"]...)})
			default:
				out.Parts = append(out.Parts, common.Part{Type: "opaque", From: &from, Data: append(json.RawMessage(nil), raw...)})
			}
		}
	case "openai":
		details := map[string]json.RawMessage{}
		if raw, ok := u["prompt_tokens_details"]; ok && string(raw) != "null" {
			details = p.obj(raw)
		}
		usage.CacheRead = p.count(details, "cached_tokens", false)
		usage.CacheWrite = p.count(details, "cache_write_tokens", false)
		usage.Input = p.count(u, "prompt_tokens", true) - usage.CacheRead - usage.CacheWrite
		usage.Output = p.count(u, "completion_tokens", true)
		choices := p.list(root["choices"])
		if len(choices) == 0 {
			p.invalid()
			break
		}
		m := p.obj(p.obj(choices[0])["message"])
		if raw, ok := m["content"]; ok && string(raw) != "null" {
			s := p.str(raw)
			out.Parts = append(out.Parts, Text(s))
		}
		if raw, ok := m["tool_calls"]; ok && string(raw) != "null" {
			for _, call := range p.list(raw) {
				c := p.obj(call)
				if p.str(c["type"]) != "function" {
					p.invalid()
				}
				fn := p.obj(c["function"])
				args := p.str(fn["arguments"])
				out.Parts = append(out.Parts, common.Part{Type: "tool_call", CallID: p.str(c["id"]), From: &from, Name: p.str(fn["name"]), Args: json.RawMessage(args)})
			}
		}
	case "gemini":
		usage.CacheRead = p.count(u, "cachedContentTokenCount", false)
		usage.Input = p.count(u, "promptTokenCount", true) - usage.CacheRead
		usage.Output = p.count(u, "candidatesTokenCount", true) + p.count(u, "thoughtsTokenCount", false)
		candidates := p.list(root["candidates"])
		if len(candidates) == 0 {
			p.invalid()
			break
		}
		content := p.obj(p.obj(candidates[0])["content"])
		for index, raw := range p.list(content["parts"]) {
			b := p.obj(raw)
			if thought, ok := b["thought"]; ok && p.truth(thought) {
				out.Parts = append(out.Parts, common.Part{Type: "opaque", From: &from, Data: append(json.RawMessage(nil), raw...)})
				continue
			}
			part := common.Part{}
			if text, ok := b["text"]; ok {
				part = Text(p.str(text))
			} else if call, ok := b["functionCall"]; ok {
				fn := p.obj(call)
				id := fmt.Sprintf("call-%d-%d", responseSeq, index)
				if raw, ok := fn["id"]; ok {
					id = p.str(raw)
				}
				part = common.Part{Type: "tool_call", CallID: id, From: &from, Name: p.str(fn["name"]), Args: append(json.RawMessage(nil), fn["args"]...)}
			} else {
				part = common.Part{Type: "opaque", From: &from, Data: append(json.RawMessage(nil), raw...)}
			}
			if signature, ok := b["thoughtSignature"]; ok && part.Type != "opaque" {
				_ = p.str(signature)
				part.From = &from
				part.Opaque = append(json.RawMessage(nil), signature...)
			}
			out.Parts = append(out.Parts, part)
		}
	default:
		p.invalid()
	}
	if usage.Input < 0 || usage.Output < 0 {
		p.invalid()
	}
	out.Usage = &usage
	visible := false
	for i := range out.Parts {
		if partProblem(owner, &out.Parts[i], false) != "" {
			p.invalid()
		}
		if out.Parts[i].Type == "text" || out.Parts[i].Type == "tool_call" {
			visible = true
		}
	}
	if !visible {
		p.invalid()
	}
	return out, p.err
}
