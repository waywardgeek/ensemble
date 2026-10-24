package engine

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"ensemble/internal/common"
)

// One reader for all vendors. ReadString has no Scanner token ceiling: a frame
// can span arbitrary network reads or exceed 64 KiB. SSE permits CRLF, comments,
// repeated data fields and a final frame without an empty separator.
type sseReader struct{ r *bufio.Reader }
type sseFrame struct{ event, data string }

func (s *sseReader) next() (sseFrame, error) {
	var f sseFrame
	var data []string
	for {
		line, err := s.r.ReadString('\n')
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if line == "" && len(data) > 0 {
			f.data = strings.Join(data, "\n")
			return f, nil
		}
		if line == "" {
			f = sseFrame{}
		}
		if !strings.HasPrefix(line, ":") {
			key, value, found := strings.Cut(line, ":")
			if found {
				value = strings.TrimPrefix(value, " ")
			}
			switch key {
			case "event":
				f.event = value
			case "data":
				data = append(data, value)
			}
		}
		if err != nil {
			if err == io.EOF && len(data) > 0 {
				f.data = strings.Join(data, "\n")
				return f, nil
			}
			return f, err
		}
	}
}

type streamPart struct {
	part common.Part
	// Wire bytes are retained until the end, including signatures and provider
	// fields needed for same-source replay. They are not dispatched as tool input.
	wire      map[string]json.RawMessage
	arguments strings.Builder
}
type streamAssembly struct {
	config      common.Config
	callbacks   common.ParseCallbacks
	parts       []*streamPart
	indexes     map[string]int
	model, stop string
	usage       json.RawMessage
	done        bool
	openAIDone  bool
	blocks      map[string]bool
	extras      map[string]json.RawMessage
}

func (s *streamAssembly) part(key, kind string) *streamPart {
	if index, ok := s.indexes[key]; ok {
		return s.parts[index]
	}
	p := &streamPart{part: common.Part{ID: uint64(len(s.parts)), Type: kind}, wire: map[string]json.RawMessage{}}
	s.indexes[key] = len(s.parts)
	s.parts = append(s.parts, p)
	return p
}
func (s *streamAssembly) delta(p *streamPart, chunk string) {
	if s.callbacks.OnDelta != nil {
		s.callbacks.OnDelta(common.PartDelta{PartID: p.part.ID, Kind: common.DeltaKind(p.part.Type), Chunk: chunk, Name: p.part.Name})
	}
}
func merge(dst, src map[string]json.RawMessage) {
	for k, v := range src {
		dst[k] = v
	}
}
func rawMap(raw json.RawMessage) map[string]json.RawMessage {
	var m map[string]json.RawMessage
	_ = json.Unmarshal(raw, &m)
	return m
}

// Frames may omit usage while counts are still arriving. Merge valid partial
// objects, but retain supplied null/invalid whole usage for the shared parser.
func (s *streamAssembly) mergeUsage(raw json.RawMessage) {
	if len(raw) == 0 {
		return
	}
	update := rawMap(raw)
	if update == nil {
		s.usage = raw
		return
	}
	prior := rawMap(s.usage)
	if prior == nil {
		prior = map[string]json.RawMessage{}
	}
	merge(prior, update)
	s.usage, _ = json.Marshal(prior)
}
func put(m map[string]json.RawMessage, k string, v any) { m[k], _ = json.Marshal(v) }
func (s *streamAssembly) frame(f sseFrame) error {
	if f.data == "[DONE]" {
		s.openAIDone = true
		return nil
	}
	var m map[string]json.RawMessage
	if json.Unmarshal([]byte(f.data), &m) != nil {
		return errors.New("invalid stream frame")
	}
	if m["error"] != nil || f.event == "error" {
		return errors.New("provider stream error")
	}
	switch s.config.Vendor {
	case common.OpenAI:
		return s.openai(m)
	case common.Anthropic:
		return s.anthropic(m)
	case common.Gemini:
		return s.gemini(m)
	}
	return errors.New("unsupported stream vendor")
}
func (s *streamAssembly) openai(m map[string]json.RawMessage) error {
	if model := str(m["model"]); model != "" {
		s.model = model
	}
	s.mergeUsage(m["usage"])
	var choices []map[string]json.RawMessage
	if raw := m["choices"]; raw != nil && json.Unmarshal(raw, &choices) != nil {
		return errors.New("invalid stream choices")
	}
	for _, c := range choices {
		if number(c, "index") != 0 {
			continue
		}
		if stop := str(c["finish_reason"]); stop != "" {
			s.stop = stop
			s.done = true
		}
		d := rawMap(c["delta"])
		if text, ok := d["content"]; ok && string(text) != "null" {
			p := s.part("text", "text")
			chunk := str(text)
			p.part.Text += chunk
			s.delta(p, chunk)
		}
		for _, key := range []string{"reasoning_content", "reasoning"} {
			if value, present := d[key]; present && string(value) != "null" {
				chunk := str(value)
				p := s.part(key, "thinking")
				p.part.Text += chunk
				put(p.wire, "field", key)
				s.delta(p, chunk)
			}
		}
		for key, raw := range d {
			switch key {
			case "role", "content", "reasoning_content", "reasoning", "tool_calls":
				continue
			}
			if key == "refusal" {
				p := s.part("refusal", "text")
				put(p.wire, "field", "refusal")
				p.part.Text += str(raw)
				s.delta(p, str(raw))
				put(s.extras, key, p.part.Text)
			} else {
				s.extras[key] = raw
			}
		}
		var calls []map[string]json.RawMessage
		if raw := d["tool_calls"]; raw != nil && json.Unmarshal(raw, &calls) != nil {
			return errors.New("invalid stream calls")
		}
		for _, call := range calls {
			key := "tool:" + string(call["index"])
			if call["index"] == nil {
				return errors.New("missing stream tool index")
			}
			p := s.part(key, "tool_call")
			if id := str(call["id"]); id != "" {
				p.part.CallID = id
			}
			fn := rawMap(call["function"])
			for k, v := range call {
				if k != "index" && k != "function" {
					p.wire[k] = v
				}
			}
			wireFn := rawMap(p.wire["function"])
			if wireFn == nil {
				wireFn = map[string]json.RawMessage{}
			}
			merge(wireFn, fn)
			put(p.wire, "function", wireFn)
			if name := str(fn["name"]); name != "" {
				p.part.Name += name
				// Emit the arriving fragment, not the accumulated name.
				s.delta(p, name)
			}
			chunk := str(fn["arguments"])
			p.arguments.WriteString(chunk)
			s.delta(p, chunk)
		}
	}
	return nil
}
func (s *streamAssembly) anthropic(m map[string]json.RawMessage) error {
	switch str(m["type"]) {
	case "message_start":
		msg := rawMap(m["message"])
		s.model = str(msg["model"])
		s.mergeUsage(msg["usage"])
	case "content_block_start":
		key := string(m["index"])
		_, exists := s.blocks[key]
		if key == "" || exists {
			return errors.New("invalid stream block index")
		}
		s.blocks[key] = true
		b := rawMap(m["content_block"])
		kind := str(b["type"])
		p := s.part("block:"+string(m["index"]), map[string]string{"text": "text", "thinking": "thinking", "redacted_thinking": "thinking", "tool_use": "tool_call"}[kind])
		merge(p.wire, b)
		switch kind {
		case "text":
			p.part.Text = str(b["text"])
			s.delta(p, p.part.Text)
		case "thinking":
			p.part.Text = str(b["thinking"])
			s.delta(p, p.part.Text)
		case "tool_use":
			p.part.CallID = str(b["id"])
			p.part.Name = str(b["name"])
			p.part.Args = b["input"]
			s.delta(p, p.part.Name)
		case "redacted_thinking":
		default:
			p.part.Type = "opaque"
		}
	case "content_block_delta":
		i, ok := s.indexes["block:"+string(m["index"])]
		if !ok {
			return errors.New("stream delta without block")
		}
		p := s.parts[i]
		d := rawMap(m["delta"])
		switch str(d["type"]) {
		case "text_delta":
			chunk := str(d["text"])
			p.part.Text += chunk
			s.delta(p, chunk)
		case "thinking_delta":
			chunk := str(d["thinking"])
			p.part.Text += chunk
			s.delta(p, chunk)
		case "input_json_delta":
			chunk := str(d["partial_json"])
			p.arguments.WriteString(chunk)
			s.delta(p, chunk)
		case "signature_delta":
			put(p.wire, "signature", str(p.wire["signature"])+str(d["signature"]))
		default:
			merge(p.wire, d)
		}
	case "content_block_stop":
		key := string(m["index"])
		if !s.blocks[key] {
			return errors.New("stream stop without block")
		}
		s.blocks[key] = false
	case "message_delta":
		d := rawMap(m["delta"])
		s.stop = str(d["stop_reason"])
		s.mergeUsage(m["usage"])
	case "message_stop":
		s.done = true
	}
	return nil
}
func (s *streamAssembly) gemini(m map[string]json.RawMessage) error {
	if model := str(m["modelVersion"]); model != "" {
		s.model = model
	}
	s.mergeUsage(m["usageMetadata"])
	var candidates []map[string]json.RawMessage
	if raw := m["candidates"]; raw != nil && json.Unmarshal(raw, &candidates) != nil {
		return errors.New("invalid stream candidates")
	}
	for _, c := range candidates {
		if number(c, "index") != 0 {
			continue
		}
		if stop := str(c["finishReason"]); stop != "" {
			s.stop = stop
			s.done = true
		}
		content := rawMap(c["content"])
		var parts []map[string]json.RawMessage
		if raw := content["parts"]; raw != nil && json.Unmarshal(raw, &parts) != nil {
			return errors.New("invalid stream parts")
		}
		for i, b := range parts {
			// A trailing signature closes exactly the preceding unsigned native
			// part. A standalone signature (including a second binding) stays
			// an opaque part; it must never overwrite an earlier signature.
			if len(b) == 1 && b["thoughtSignature"] != nil && len(s.parts) > 0 {
				previous := s.parts[len(s.parts)-1]
				if previous.wire["thoughtSignature"] == nil && (previous.wire["text"] != nil || previous.wire["functionCall"] != nil) {
					previous.wire["thoughtSignature"] = b["thoughtSignature"]
					continue
				}
			}
			var p *streamPart
			if i == 0 && len(s.parts) > 0 && geminiTextContinuation(s.parts[len(s.parts)-1].wire, b) {
				p = s.parts[len(s.parts)-1]
			} else {
				kind := "opaque"
				switch {
				case string(b["thought"]) == "true":
					if b["text"] != nil {
						kind = "thinking"
					}
				case b["functionCall"] != nil:
					kind = "tool_call"
				case b["text"] != nil:
					kind = "text"
				}
				p = s.part("gemini:"+strconv.Itoa(len(s.parts)), kind)
			}
			merge(p.wire, b)
			switch p.part.Type {
			case "text", "thinking":
				chunk := str(b["text"])
				p.part.Text += chunk
				s.delta(p, chunk)
			case "tool_call":
				call := rawMap(b["functionCall"])
				p.part.Name = str(call["name"])
				p.part.CallID = str(call["id"])
				p.part.Args = call["args"]
				s.delta(p, p.part.Name+string(p.part.Args))
			}
		}
	}
	return nil
}

// Gemini repeats an unfinished text part only as the first part of a later
// frame. Explicit siblings, signatures, and unfamiliar fields close that part.
func geminiTextContinuation(previous, next map[string]json.RawMessage) bool {
	if previous["text"] == nil || next["text"] == nil {
		return false
	}
	for _, part := range []map[string]json.RawMessage{previous, next} {
		for key := range part {
			if key != "text" && key != "thought" {
				return false
			}
		}
	}
	return (string(previous["thought"]) == "true") == (string(next["thought"]) == "true")
}

// IDs follow the provider's concrete output fields/blocks, not a semantic-kind
// match after parsing (two opaque reasoning fields can arrive in either order).
func (s *streamAssembly) finish() ([]byte, []uint64, error) {
	if s.stop == "" || !s.done || (s.config.Vendor == common.OpenAI && !s.openAIDone) {
		return nil, nil, errors.New("incomplete stream response")
	}
	for _, open := range s.blocks {
		if open {
			return nil, nil, errors.New("SSE response ended with an unfinished content block")
		}
	}
	for _, p := range s.parts {
		if p.part.Type == "tool_call" {
			if p.arguments.Len() > 0 {
				p.part.Args = json.RawMessage(p.arguments.String())
			}
			if p.part.Name == "" || (s.config.Vendor != common.Gemini && p.part.CallID == "") {
				return nil, nil, errors.New("incomplete streamed tool call")
			}
			args, err := canonical(p.part.Args)
			if err != nil {
				return nil, nil, errors.New("incomplete streamed tool arguments")
			}
			p.part.Args = args
		}
	}
	var body any
	var ids []uint64
	switch s.config.Vendor {
	case common.OpenAI:
		msg := map[string]any{}
		for _, field := range []struct{ key, name string }{{"text", "content"}, {"refusal", "refusal"}, {"reasoning_content", "reasoning_content"}, {"reasoning", "reasoning"}} {
			if i, ok := s.indexes[field.key]; ok {
				p := s.parts[i]
				msg[field.name] = p.part.Text
				if field.name != "refusal" || p.part.Text != "" {
					ids = append(ids, p.part.ID)
				}
			}
		}
		var calls []any
		for _, p := range s.parts {
			if p.part.Type == "tool_call" {
				put(p.wire, "id", p.part.CallID)
				put(p.wire, "type", "function")
				fn := rawMap(p.wire["function"])
				if fn == nil {
					fn = map[string]json.RawMessage{}
				}
				put(fn, "name", p.part.Name)
				put(fn, "arguments", string(p.part.Args))
				put(p.wire, "function", fn)
				calls = append(calls, p.wire)
				ids = append(ids, p.part.ID)
			}
		}
		if len(calls) > 0 {
			msg["tool_calls"] = calls
		}
		body = map[string]any{"model": s.model, "choices": []any{map[string]any{"message": msg, "finish_reason": s.stop}}, "usage": s.usage}
	case common.Anthropic:
		var blocks []any
		for _, p := range s.parts {
			switch str(p.wire["type"]) {
			case "text":
				put(p.wire, "text", p.part.Text)
				ids = append(ids, p.part.ID)
			case "thinking":
				put(p.wire, "thinking", p.part.Text)
				ids = append(ids, p.part.ID)
			case "redacted_thinking":
				ids = append(ids, p.part.ID)
			case "tool_use":
				put(p.wire, "input", json.RawMessage(p.part.Args))
				ids = append(ids, p.part.ID)
			}
			blocks = append(blocks, p.wire)
		}
		body = map[string]any{"model": s.model, "content": blocks, "stop_reason": s.stop, "usage": s.usage}
	case common.Gemini:
		var parts []any
		for _, p := range s.parts {
			if p.part.Type == "text" || p.part.Type == "thinking" {
				put(p.wire, "text", p.part.Text)
			}
			parts = append(parts, p.wire)
			ids = append(ids, p.part.ID)
		}
		body = map[string]any{"modelVersion": s.model, "candidates": []any{map[string]any{"content": map[string]any{"parts": parts}, "finishReason": s.stop}}, "usageMetadata": s.usage}
	default:
		return nil, nil, errors.New("unsupported SSE provider")
	}
	encoded, err := json.Marshal(body)
	return encoded, ids, err
}

func parseSSE(r io.Reader, c common.Config, cb common.ParseCallbacks) ([]common.Event, error) {
	reader := sseReader{bufio.NewReader(r)}
	s := streamAssembly{config: c, callbacks: cb, indexes: map[string]int{}, blocks: map[string]bool{}, extras: map[string]json.RawMessage{}}
	for {
		frame, err := reader.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return s.failure("reading stream failed")
		}
		if cb.OnFrame != nil {
			cb.OnFrame(common.Frame{Event: frame.event, Data: frame.data})
		}
		if err = s.frame(frame); err != nil {
			return s.failure(err.Error())
		}
	}
	body, ids, err := s.finish()
	if err != nil {
		return s.failure(err.Error())
	}
	events, err := parse(200, body, c)
	if err != nil {
		return events, err
	}
	// Parse supplies the exact same final semantic/provenance policy as plain
	// JSON. Match final parts to arrival IDs (OpenAI's canonical field order may
	// differ from arrival order), never treating zero as an absent identity.
	for i := range events {
		r := events[i].Response
		if events[i].Type != common.ResponseEnded || r == nil {
			continue
		}
		if len(r.Parts) != len(ids) {
			return s.failure("streamed native part layout changed during final parsing")
		}
		for j := range r.Parts {
			r.Parts[j].ID = ids[j]
		}
	}
	return events, nil
}
func parseFailure(message string) ([]common.Event, error) {
	return []common.Event{errorEvent(message)}, errors.New(message)
}

func str(b json.RawMessage) string { var s string; _ = json.Unmarshal(b, &s); return s }
func number(m map[string]json.RawMessage, key string) int {
	var n int
	_ = json.Unmarshal(m[key], &n)
	return n
}

// Interrupted streams can still carry measured charges. Retain those facts,
// but never turn an incomplete tool buffer into a final executable part.
func (s *streamAssembly) failure(message string) ([]common.Event, error) {
	usageJSON, _ := json.Marshal(s.usage)
	key := "usage"
	if s.config.Vendor == common.Gemini {
		key = "usageMetadata"
	}
	usage, missing, _ := parseUsage(s.config, map[string]json.RawMessage{key: usageJSON})
	from := common.Provenance{Vendor: s.config.Vendor, Surface: s.config.Surface, Model: s.model}
	return []common.Event{
		{Type: common.ResponseStarted, Response: &common.ResponseData{From: from}},
		{Type: common.ResponseEnded, Response: &common.ResponseData{From: from, Usage: usage, UsageMissing: missing}},
		errorEvent(message),
	}, errors.New(message)
}
