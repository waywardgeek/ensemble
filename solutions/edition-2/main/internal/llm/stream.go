package llm

import (
	"context"
	"encoding/json"
	"io"
	"sort"
	"strings"

	"example.com/ensemble/internal/common"
)

type streamBlock struct {
	id      int
	kind    string
	raw     map[string]json.RawMessage
	values  map[string]*strings.Builder
	callID  string
	stopped bool
}
type streamParser struct {
	parent                     common.ModelOperation
	p                          parser
	config                     common.Config
	nextID, size               int
	model                      string
	usage                      map[string]json.RawMessage
	blocks                     map[int]*streamBlock
	parts                      []*geminiPart
	ids                        []int
	text, refusal              strings.Builder
	textID, refusalID          int
	sawText, started, finished bool
	stop                       string
}

func rawValue(v any) json.RawMessage             { b, _ := json.Marshal(v); return b }
func (s *streamParser) fail(reason string) error { return failure(s.parent.Engine(), "%s", reason) }
func (s *streamParser) allocate() int            { s.nextID++; return s.nextID }
func (s *streamParser) add(n int) error {
	s.size += n
	if s.size > responseLimit {
		return s.fail("assembled response exceeds 16 MiB")
	}
	return nil
}
func (s *streamParser) emit(ctx context.Context, id int, channel, text string) error {
	if text == "" {
		return nil
	}
	return s.parent.Emit(ctx, common.Fragment{PartID: id, Channel: channel, Text: text})
}
func (s *streamParser) identity(root map[string]json.RawMessage, key string) error {
	if raw, ok := root[key]; ok {
		model := s.p.str(raw)
		if model != "" {
			if s.model != "" && s.model != model {
				return s.fail("conflicting returned model identities")
			}
			s.model = model
		}
	}
	return s.p.err
}
func (s *streamParser) index(raw json.RawMessage) (int, error) {
	var index int
	if len(raw) == 0 || json.Unmarshal(raw, &index) != nil || index < 0 {
		return 0, s.fail("invalid stream part index")
	}
	return index, nil
}
func parseStream(ctx context.Context, owner common.ModelOperation, config common.Config, r io.Reader) (common.ParsedResponse, error) {
	s := streamParser{parent: owner, p: parser{parent: owner.Engine(), operation: owner}, config: config, blocks: map[int]*streamBlock{}}
	reader := newSSE(owner, r)
	for {
		event, err := reader.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return common.ParsedResponse{}, err
		}
		if config.Vendor == "openai" && event.data == "[DONE]" {
			if s.stop == "" || s.usage == nil {
				return common.ParsedResponse{}, s.fail("stream ended without finish reason or usage")
			}
			s.finished = true
			break
		}
		root := s.p.obj([]byte(event.data))
		if s.p.err != nil {
			return common.ParsedResponse{}, s.p.err
		}
		if _, ok := root["error"]; ok {
			return common.ParsedResponse{}, s.fail("provider reported stream error")
		}
		switch config.Vendor {
		case "anthropic":
			err = s.messages(ctx, root)
		case "openai":
			err = s.chat(ctx, root)
		case "gemini":
			err = s.gemini(ctx, root)
		default:
			err = s.fail("unknown streaming API")
		}
		if err == nil {
			err = s.p.err
		}
		if err != nil {
			return common.ParsedResponse{}, err
		}
		if config.Vendor == "anthropic" && s.finished {
			break
		}
	}
	if !s.finished {
		return common.ParsedResponse{}, s.fail("stream ended before API terminal signal")
	}
	body, ids, err := s.envelope()
	if err != nil {
		return common.ParsedResponse{}, err
	}
	parsed, err := parseResponse(owner.Engine(), config, body, 0, owner)
	parsed.PartIDs = ids
	return parsed, err
}
func (s *streamParser) messages(ctx context.Context, root map[string]json.RawMessage) error {
	kind := s.p.str(root["type"])
	switch kind {
	case "message_start":
		if s.started {
			return s.fail("duplicate message_start")
		}
		s.started = true
		m := s.p.obj(root["message"])
		if err := s.identity(m, "model"); err != nil {
			return err
		}
		s.usage = s.p.obj(m["usage"])
	case "content_block_start":
		if !s.started {
			return s.fail("block before message_start")
		}
		index, err := s.index(root["index"])
		if err != nil {
			return err
		}
		if s.blocks[index] != nil {
			return s.fail("duplicate block start")
		}
		b := &streamBlock{id: s.allocate(), raw: s.p.obj(root["content_block"]), values: map[string]*strings.Builder{}}
		b.kind = s.p.str(b.raw["type"])
		s.blocks[index] = b
		size := len(root["content_block"])
		initial := ""
		if b.kind == "text" || b.kind == "thinking" {
			field := b.kind
			initial = s.p.str(b.raw[field])
			b.values[field] = &strings.Builder{}
			b.values[field].WriteString(initial)
			delete(b.raw, field)
			if b.kind == "text" {
				size = len(initial)
			}
			if signature, ok := b.raw["signature"]; ok {
				b.values["signature"] = &strings.Builder{}
				b.values["signature"].WriteString(s.p.str(signature))
				delete(b.raw, "signature")
			}
		}
		if err = s.add(size); err != nil {
			return err
		}
		switch b.kind {
		case "text":
			return s.emit(ctx, b.id, "text", initial)
		case "thinking":
			return s.emit(ctx, b.id, "thinking", initial)
		case "tool_use":
			return s.emit(ctx, b.id, "tool_name", s.p.str(b.raw["name"]))
		}
	case "content_block_delta":
		index, err := s.index(root["index"])
		if err != nil {
			return err
		}
		b := s.blocks[index]
		if b == nil || b.stopped {
			return s.fail("delta requires open block")
		}
		d := s.p.obj(root["delta"])
		kind := s.p.str(d["type"])
		field, channel := "", ""
		switch kind {
		case "text_delta":
			if b.kind == "text" {
				field = "text"
				channel = "text"
			}
		case "thinking_delta":
			if b.kind == "thinking" {
				field = "thinking"
				channel = "thinking"
			}
		case "signature_delta":
			if b.kind == "thinking" {
				field = "signature"
			}
		case "input_json_delta":
			if b.kind == "tool_use" {
				field = "partial_json"
				channel = "tool_args"
			}
		}
		if field == "" {
			return s.fail("unsupported delta for block kind")
		}
		text := s.p.str(d[field])
		if field == "partial_json" {
			field = "input"
			if b.values[field] == nil {
				s.size -= len(b.raw[field])
				delete(b.raw, field)
			}
		}
		if err = s.add(len(text)); err != nil {
			return err
		}
		if b.values[field] == nil {
			b.values[field] = &strings.Builder{}
		}
		b.values[field].WriteString(text)
		if channel != "" {
			return s.emit(ctx, b.id, channel, text)
		}
	case "content_block_stop":
		index, err := s.index(root["index"])
		if err != nil {
			return err
		}
		b := s.blocks[index]
		if b == nil || b.stopped {
			return s.fail("stop requires open block")
		}
		b.stopped = true
		if b.kind == "tool_use" {
			if b.values["input"] == nil {
				return s.emit(ctx, b.id, "tool_args", string(b.raw["input"]))
			}
		}
	case "message_delta":
		if !s.started {
			return s.fail("delta before message_start")
		}
		if u, ok := root["usage"]; ok {
			for k, v := range s.p.obj(u) {
				s.usage[k] = v
			}
		}
		if d, ok := root["delta"]; ok {
			m := s.p.obj(d)
			if v, ok := m["stop_reason"]; ok && string(v) != "null" {
				s.stop = s.p.str(v)
			}
		}
	case "message_stop":
		if !s.started || s.stop == "" {
			return s.fail("message_stop requires stop reason")
		}
		for _, b := range s.blocks {
			if !b.stopped {
				return s.fail("message_stop with open block")
			}
		}
		s.finished = true
	case "error":
		return s.fail("provider reported stream error")
	}
	return s.p.err
}
func (s *streamParser) chat(ctx context.Context, root map[string]json.RawMessage) error {
	if err := s.identity(root, "model"); err != nil {
		return err
	}
	if u, ok := root["usage"]; ok && string(u) != "null" {
		s.usage = s.p.obj(u)
	}
	for _, raw := range s.p.list(root["choices"]) {
		c := s.p.obj(raw)
		index, err := s.index(c["index"])
		if err != nil {
			return err
		}
		if index != 0 {
			continue
		}
		if f, ok := c["finish_reason"]; ok && string(f) != "null" {
			s.stop = s.p.str(f)
		}
		d := s.p.obj(c["delta"])
		if value, ok := d["content"]; ok && string(value) != "null" {
			text := s.p.str(value)
			if err = s.add(len(text)); err != nil {
				return err
			}
			if s.textID == 0 {
				s.textID = s.allocate()
			}
			s.sawText = true
			s.text.WriteString(text)
			if err = s.emit(ctx, s.textID, "text", text); err != nil {
				return err
			}
		}
		if value, ok := d["refusal"]; ok && string(value) != "null" {
			text := s.p.str(value)
			if err = s.add(len(text)); err != nil {
				return err
			}
			if s.refusalID == 0 {
				s.refusalID = s.allocate()
			}
			s.refusal.WriteString(text)
		}
		if value, ok := d["tool_calls"]; ok && string(value) != "null" {
			for _, raw := range s.p.list(value) {
				call := s.p.obj(raw)
				index, err := s.index(call["index"])
				if err != nil {
					return err
				}
				b := s.blocks[index]
				if b == nil {
					b = &streamBlock{id: s.allocate(), raw: map[string]json.RawMessage{}, values: map[string]*strings.Builder{}}
					s.blocks[index] = b
				}
				if v, ok := call["id"]; ok {
					id := s.p.str(v)
					if _, ok := b.raw["id"]; ok {
						if b.callID != id {
							return s.fail("conflicting tool call identity")
						}
					} else if err = s.add(len(id)); err != nil {
						return err
					}
					b.callID = id
					b.raw["id"] = v
				}
				if v, ok := call["type"]; ok && s.p.str(v) != "function" {
					return s.fail("unsupported tool call type")
				}
				if v, ok := call["function"]; ok {
					fn := s.p.obj(v)
					for _, field := range []string{"name", "arguments"} {
						if v, ok := fn[field]; ok {
							text := s.p.str(v)
							if err = s.add(len(text)); err != nil {
								return err
							}
							if b.values[field] == nil {
								b.values[field] = &strings.Builder{}
							}
							b.values[field].WriteString(text)
							channel := "tool_args"
							if field == "name" {
								channel = "tool_name"
							}
							if err = s.emit(ctx, b.id, channel, text); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	return s.p.err
}

// Gemini records retain their first raw part and classification. Appending a
// compatible run touches only its builder and counters, never its prior JSON.
type geminiPart struct {
	raw                 json.RawMessage
	fields              map[string]json.RawMessage
	text                strings.Builder
	channel             string
	mergeable, merged   bool
	cost, encoded, base int
}

func classifyGemini(p *parser, raw json.RawMessage) (*geminiPart, []common.Fragment) {
	b := p.obj(raw)
	part := &geminiPart{raw: raw, fields: b, cost: len(raw)}
	if value, hasText := b["text"]; hasText {
		text := p.str(value)
		thought := false
		if value, ok := b["thought"]; ok {
			thought = p.truth(value)
		}
		signature, signed := b["thoughtSignature"]
		if signed {
			_ = p.str(signature)
		}
		for key := range b {
			if key != "text" && key != "thought" && key != "thoughtSignature" {
				return part, nil
			}
		}
		part.channel = "text"
		if thought {
			part.channel = "thinking"
		}
		part.mergeable = !signed
		if !thought {
			part.cost = len(text) + len(signature)
		}
		if part.mergeable {
			part.text.WriteString(text)
			// Only the incoming fragment is encoded for opaque thought accounting.
			// Canonical structural cost is fixed even when the original had whitespace.
			part.encoded = len(rawValue(text)) - 2
			empty := map[string]json.RawMessage{"text": json.RawMessage(`""`)}
			if value, ok := b["thought"]; ok {
				empty["thought"] = value
			}
			part.base = len(rawValue(empty))
		}
		return part, []common.Fragment{{Channel: part.channel, Text: text}}
	}
	if value, ok := b["functionCall"]; ok {
		call := p.obj(value)
		return part, []common.Fragment{{Channel: "tool_name", Text: p.str(call["name"])}, {Channel: "tool_args", Text: string(call["args"])}}
	}
	return part, nil
}
func appendGemini(p *parser, parts []*geminiPart, next *geminiPart) ([]*geminiPart, bool, int) {
	if next.mergeable && len(parts) > 0 {
		last := parts[len(parts)-1]
		if last.mergeable && last.channel == next.channel {
			before := last.cost
			last.text.WriteString(next.text.String())
			last.encoded += next.encoded
			last.merged = true
			if last.channel == "text" {
				last.cost += next.cost
			} else {
				last.cost = last.base + last.encoded
			}
			return parts, true, last.cost - before
		}
	}
	return append(parts, next), false, next.cost
}
func materializeGemini(p *parser, parts []*geminiPart) []json.RawMessage {
	out := make([]json.RawMessage, 0, len(parts))
	for _, part := range parts {
		raw := part.raw
		if part.merged {
			part.fields["text"] = rawValue(part.text.String())
			raw = rawValue(part.fields)
		}
		out = append(out, raw)
	}
	return out
}
func (s *streamParser) gemini(ctx context.Context, root map[string]json.RawMessage) error {
	if err := s.identity(root, "modelVersion"); err != nil {
		return err
	}
	if u, ok := root["usageMetadata"]; ok {
		s.usage = s.p.obj(u)
	}
	if raw, ok := root["candidates"]; ok {
		candidates := s.p.list(raw)
		for _, raw := range candidates {
			c := s.p.obj(raw)
			index := 0
			var err error
			if v, ok := c["index"]; ok {
				index, err = s.index(v)
				if err != nil {
					return err
				}
			} else if len(candidates) != 1 {
				return s.fail("ambiguous candidate index")
			}
			if index != 0 {
				continue
			}
			if value, ok := c["content"]; ok {
				content := s.p.obj(value)
				if value, ok := content["parts"]; ok {
					for _, raw := range s.p.list(value) {
						next, fragments := classifyGemini(&s.p, raw)
						var merged bool
						var size int
						s.parts, merged, size = appendGemini(&s.p, s.parts, next)
						if err = s.add(size); err != nil {
							return err
						}
						if !merged {
							s.ids = append(s.ids, s.allocate())
						}
						id := s.ids[len(s.ids)-1]
						for _, fragment := range fragments {
							if err = s.emit(ctx, id, fragment.Channel, fragment.Text); err != nil {
								return err
							}
						}

					}
				}
			}
			if f, ok := c["finishReason"]; ok {
				reason := s.p.str(f)
				if reason != "" {
					s.stop = reason
					s.finished = true
				}
			}
		}
	}
	return s.p.err
}
func (s *streamParser) envelope() ([]byte, []int, error) {
	root := map[string]any{}
	indices := make([]int, 0, len(s.blocks))
	for index := range s.blocks {
		indices = append(indices, index)
	}
	sort.Ints(indices)
	ids := []int{}
	switch s.config.Vendor {
	case "anthropic":
		blocks := []any{}
		for _, index := range indices {
			b := s.blocks[index]
			for field, value := range b.values {
				if field == "input" {
					b.raw[field] = json.RawMessage(value.String())
				} else {
					b.raw[field] = rawValue(value.String())
				}
			}
			blocks = append(blocks, b.raw)
			ids = append(ids, b.id)
		}
		root["content"] = blocks
		root["usage"] = s.usage
		root["stop_reason"] = s.stop
	case "openai":
		m := map[string]any{"content": nil}
		if s.sawText {
			m["content"] = s.text.String()
			ids = append(ids, s.textID)
		}
		calls := []any{}
		for _, index := range indices {
			b := s.blocks[index]
			for field, value := range b.values {
				b.raw[field] = rawValue(value.String())
			}
			calls = append(calls, map[string]any{"type": "function", "id": b.raw["id"], "function": map[string]any{"name": b.raw["name"], "arguments": b.raw["arguments"]}})
			ids = append(ids, b.id)
		}
		if len(calls) > 0 {
			m["tool_calls"] = calls
		}
		if s.refusalID != 0 {
			m["refusal"] = s.refusal.String()
			ids = append(ids, s.refusalID)
		}
		root["choices"] = []any{map[string]any{"index": 0, "message": m, "finish_reason": s.stop}}
		root["usage"] = s.usage
	case "gemini":
		root["candidates"] = []any{map[string]any{"index": 0, "content": map[string]any{"parts": materializeGemini(&s.p, s.parts)}, "finishReason": s.stop}}
		root["usageMetadata"] = s.usage
		ids = s.ids
	}
	if s.model != "" {
		key := "model"
		if s.config.Vendor == "gemini" {
			key = "modelVersion"
		}
		root[key] = s.model
	}
	body, err := json.Marshal(root)
	if err != nil {
		return nil, nil, s.fail("invalid assembled response")
	}
	return body, ids, nil
}

// Unknown text-bearing combinations are indivisible replay material. Validate
// recognized fields before classification so opaque retention cannot hide errors.
func knownGeminiText(p *parser, b map[string]json.RawMessage) bool {
	raw, ok := b["text"]
	if !ok {
		return false
	}
	_ = p.str(raw)
	if raw, ok := b["thought"]; ok {
		_ = p.truth(raw)
	}
	if raw, ok := b["thoughtSignature"]; ok {
		_ = p.str(raw)
	}
	for key := range b {
		if key != "text" && key != "thought" && key != "thoughtSignature" {
			return false
		}
	}
	return true
}
