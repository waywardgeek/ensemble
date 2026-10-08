package llm

import (
	"context"
	"encoding/json"
	"io"
	"sort"

	"example.com/ensemble/internal/common"
)

type streamBlock struct {
	id      int
	kind    string
	raw     map[string]json.RawMessage
	args    string
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
	parts                      []json.RawMessage
	ids                        []int
	text, refusal              string
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
		b := &streamBlock{id: s.allocate(), raw: s.p.obj(root["content_block"])}
		b.kind = s.p.str(b.raw["type"])
		s.blocks[index] = b
		size := len(root["content_block"])
		if b.kind == "text" {
			size = len(s.p.str(b.raw["text"]))
		}
		if err = s.add(size); err != nil {
			return err
		}
		switch b.kind {
		case "text":
			return s.emit(ctx, b.id, "text", s.p.str(b.raw["text"]))
		case "thinking":
			return s.emit(ctx, b.id, "thinking", s.p.str(b.raw["thinking"]))
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
		if err = s.add(len(text)); err != nil {
			return err
		}
		if field == "partial_json" {
			b.args += text
		} else {
			old := ""
			if value, ok := b.raw[field]; ok {
				old = s.p.str(value)
			}
			b.raw[field] = rawValue(old + text)
		}
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
			if b.args != "" {
				b.raw["input"] = json.RawMessage(b.args)
			} else {
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
			s.text += text
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
			s.refusal += text
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
					b = &streamBlock{id: s.allocate(), raw: map[string]json.RawMessage{}}
					s.blocks[index] = b
				}
				if v, ok := call["id"]; ok {
					id := s.p.str(v)
					if old, ok := b.raw["id"]; ok && s.p.str(old) != id {
						return s.fail("conflicting tool call identity")
					}
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
							old := ""
							if v, ok := b.raw[field]; ok {
								old = s.p.str(v)
							}
							b.raw[field] = rawValue(old + text)
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

// textRun describes only unsigned text-only blocks; opaque fields end a run.
func textRun(p *parser, b map[string]json.RawMessage) (string, string, bool) {
	raw, ok := b["text"]
	if !ok {
		return "", "", false
	}
	channel := "text"
	allowed := 1
	if thought, ok := b["thought"]; ok {
		allowed++
		if p.truth(thought) {
			channel = "thinking"
		}
	}
	if len(b) != allowed {
		return "", "", false
	}
	return channel, p.str(raw), true
}
func appendGemini(p *parser, parts []json.RawMessage, raw json.RawMessage) ([]json.RawMessage, bool) {
	b := p.obj(raw)
	if !knownGeminiText(p, b) {
		return append(parts, append(json.RawMessage(nil), raw...)), false
	}
	channel, text, run := textRun(p, b)
	if run && len(parts) > 0 {
		last := p.obj(parts[len(parts)-1])
		oldChannel, old, oldRun := textRun(p, last)
		if oldRun && oldChannel == channel {
			last["text"] = rawValue(old + text)
			parts[len(parts)-1] = rawValue(last)
			return parts, true
		}
	}
	return append(parts, append(json.RawMessage(nil), raw...)), false
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
						previous := 0
						if len(s.parts) > 0 {
							previous = geminiSize(&s.p, s.parts[len(s.parts)-1])
						}
						var merged bool
						s.parts, merged = appendGemini(&s.p, s.parts, raw)
						size := geminiSize(&s.p, s.parts[len(s.parts)-1])
						if merged {
							size -= previous
						}
						if err = s.add(size); err != nil {
							return err
						}
						if !merged {
							s.ids = append(s.ids, s.allocate())
						}
						id := s.ids[len(s.ids)-1]
						b := s.p.obj(raw)
						if _, hasText := b["text"]; hasText && !knownGeminiText(&s.p, b) {
							continue
						}
						if value, ok := b["text"]; ok && knownGeminiText(&s.p, b) {
							channel := "text"
							if v, ok := b["thought"]; ok && s.p.truth(v) {
								channel = "thinking"
							}
							if err = s.emit(ctx, id, channel, s.p.str(value)); err != nil {
								return err
							}
						}
						if value, ok := b["functionCall"]; ok {
							call := s.p.obj(value)
							if err = s.emit(ctx, id, "tool_name", s.p.str(call["name"])); err != nil {
								return err
							}
							if err = s.emit(ctx, id, "tool_args", string(call["args"])); err != nil {
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
			blocks = append(blocks, b.raw)
			ids = append(ids, b.id)
		}
		root["content"] = blocks
		root["usage"] = s.usage
		root["stop_reason"] = s.stop
	case "openai":
		m := map[string]any{"content": nil}
		if s.sawText {
			m["content"] = s.text
			ids = append(ids, s.textID)
		}
		calls := []any{}
		for _, index := range indices {
			b := s.blocks[index]
			calls = append(calls, map[string]any{"type": "function", "id": b.raw["id"], "function": map[string]any{"name": b.raw["name"], "arguments": b.raw["arguments"]}})
			ids = append(ids, b.id)
		}
		if len(calls) > 0 {
			m["tool_calls"] = calls
		}
		if s.refusalID != 0 {
			m["refusal"] = s.refusal
			ids = append(ids, s.refusalID)
		}
		root["choices"] = []any{map[string]any{"index": 0, "message": m, "finish_reason": s.stop}}
		root["usage"] = s.usage
	case "gemini":
		root["candidates"] = []any{map[string]any{"index": 0, "content": map[string]any{"parts": s.parts}, "finishReason": s.stop}}
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

func geminiSize(p *parser, raw json.RawMessage) int {
	b := p.obj(raw)
	if knownGeminiText(p, b) {
		channel, text, run := textRun(p, b)
		if run && channel == "text" {
			return len(text)
		}
		if thought, ok := b["thought"]; !ok || !p.truth(thought) {
			return len(p.str(b["text"])) + len(b["thoughtSignature"])
		}
	}
	return len(raw)
}
