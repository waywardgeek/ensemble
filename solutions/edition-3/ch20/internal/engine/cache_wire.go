package engine

import (
	"bytes"
	"encoding/json"
	"ensemble/internal/common"
	"sort"
)

// Spans refer to the original wire bytes. JSON validation happens once up front;
// this reader only locates members, never re-encodes strings, numbers or objects.
// In particular a marker-looking word inside text/tool arguments is not metadata.
type wireNode struct {
	start, end int
	fields     []wireField
	items      []wireNode
}
type wireField struct {
	name         string
	start, comma int
	value        wireNode
}
type byteSpan struct{ start, end int }

func wireSpace(b []byte, i int) int {
	for i < len(b) && (b[i] == ' ' || b[i] == '\n' || b[i] == '\r' || b[i] == '\t') {
		i++
	}
	return i
}
func wireStringEnd(b []byte, i int) int {
	i++
	for i < len(b) {
		if b[i] == '\\' {
			i += 2
			continue
		}
		if b[i] == '"' {
			return i + 1
		}
		i++
	}
	return i
}
func wireValue(b []byte, i int) wireNode {
	i = wireSpace(b, i)
	n := wireNode{start: i}
	switch b[i] {
	case '{':
		i = wireSpace(b, i+1)
		for b[i] != '}' {
			start := i
			end := wireStringEnd(b, i)
			var key string
			_ = json.Unmarshal(b[i:end], &key)
			i = wireSpace(b, end)
			value := wireValue(b, i+1)
			i = wireSpace(b, value.end)
			f := wireField{name: key, start: start, comma: -1, value: value}
			if b[i] == ',' {
				f.comma = i
				i = wireSpace(b, i+1)
			}
			n.fields = append(n.fields, f)
		}
		n.end = i + 1
	case '[':
		i = wireSpace(b, i+1)
		for b[i] != ']' {
			v := wireValue(b, i)
			n.items = append(n.items, v)
			i = wireSpace(b, v.end)
			if b[i] == ',' {
				i = wireSpace(b, i+1)
			}
		}
		n.end = i + 1
	case '"':
		n.end = wireStringEnd(b, i)
	default:
		for i < len(b) && !bytes.ContainsRune([]byte(" \r\n\t,]}"), rune(b[i])) {
			i++
		}
		n.end = i
	}
	return n
}
func wireRoot(b []byte) (wireNode, bool) {
	if !json.Valid(b) {
		return wireNode{}, false
	}
	return wireValue(b, 0), true
}
func (n wireNode) field(name string) wireNode {
	for i := len(n.fields) - 1; i >= 0; i-- {
		if n.fields[i].name == name {
			return n.fields[i].value
		}
	}
	return wireNode{}
}
func memberSpans(n wireNode, index int) []byteSpan {
	f := n.fields[index]
	r := []byteSpan{{f.start, f.value.end}}
	if f.comma >= 0 {
		r = append(r, byteSpan{f.comma, f.comma + 1})
	} else if index > 0 {
		p := n.fields[index-1].comma
		r = append(r, byteSpan{p, p + 1})
	}
	return r
}
func withoutSpans(b []byte, spans []byteSpan) []byte {
	sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })
	out := make([]byte, 0, len(b))
	at := 0
	for _, s := range spans {
		if s.start > at {
			out = append(out, b[at:s.start]...)
		}
		if s.end > at {
			at = s.end
		}
	}
	return append(out, b[at:]...)
}
func markerMembers(cfg common.Config, root wireNode, b []byte) ([]byteSpan, map[int]bool) {
	var spans []byteSpan
	marked := map[int]bool{}
	add := func(n wireNode, key string) {
		for i, f := range n.fields {
			if f.name == key {
				spans = append(spans, memberSpans(n, i)...)
				marked[n.start] = true
			}
		}
	}
	switch {
	case cfg.Vendor == common.Anthropic && cfg.Surface == common.Messages:
		for _, key := range []string{"tools", "system"} {
			for _, n := range root.field(key).items {
				if key == "tools" || wireText(b, n.field("type")) == "text" {
					add(n, "cache_control")
				}
			}
		}
		for _, m := range root.field("messages").items {
			for _, n := range m.field("content").items {
				switch wireText(b, n.field("type")) {
				case "text", "tool_use", "tool_result":
					add(n, "cache_control")
				}
				if marked[n.start] {
					marked[m.start] = true
				}
			}
		}
	case cfg.Vendor == common.OpenAI && cfg.Surface == common.Responses:
		for _, m := range root.field("input").items {
			for _, n := range m.field("content").items {
				kind := wireText(b, n.field("type"))
				if kind == "input_text" || kind == "output_text" {
					add(n, "prompt_cache_breakpoint")
				}
				if marked[n.start] {
					marked[m.start] = true
				}
			}
		}
		add(root, "prompt_cache_options")
	case cfg.Vendor == common.OpenAI && cfg.Surface == common.ChatCompletions:
		for _, m := range root.field("messages").items {
			for _, n := range m.field("content").items {
				if wireText(b, n.field("type")) == "text" {
					add(n, "prompt_cache_breakpoint")
				}
				if marked[n.start] {
					marked[m.start] = true
				}
			}
		}
		add(root, "prompt_cache_options")
	}
	return spans, marked
}
func wireText(b []byte, n wireNode) string {
	var text string
	if n.end > n.start {
		_ = json.Unmarshal(b[n.start:n.end], &text)
	}
	return text
}
func stripCacheMarkers(cfg common.Config, b []byte) []byte {
	root, ok := wireRoot(b)
	if !ok {
		return append([]byte(nil), b...)
	}
	spans, _ := markerMembers(cfg, root, b)
	return withoutSpans(b, spans)
}

type promptSection struct {
	source string
	offset int
	name   string
	raw    []byte
	node   wireNode
}

// Follow provider cache order, not the JSON object's top-level field order.
// Only the leading OpenAI SYSTEM message is the baseline section. A later
// SYSTEM-role ephemeral tail stays in messages at its actual wire position.
func promptSections(b []byte) []promptSection {
	root, ok := wireRoot(b)
	if !ok {
		return nil
	}
	var out []promptSection
	for _, key := range []string{"tools", "system", "systemInstruction", "instructions", "messages", "contents", "input"} {
		n := root.field(key)
		if n.end <= n.start {
			continue
		}
		if key == "messages" && root.field("system").end == 0 && len(n.items) > 0 && (wireText(b, n.items[0].field("role")) == "system" || wireText(b, n.items[0].field("role")) == "developer") {
			boundary := n.items[0].end
			baseline := n
			baseline.end = boundary
			baseline.items = n.items[:1]
			tail := n
			tail.start = boundary
			tail.items = n.items[1:]
			out = append(out, promptSection{source: key, name: "system", raw: b[n.start:boundary], node: baseline}, promptSection{source: key, offset: 1, name: "messages", raw: b[boundary:n.end], node: tail})
		} else {
			name := key
			switch key {
			case "systemInstruction", "instructions":
				name = "system"
			case "contents", "input":
				name = "messages"
			}
			out = append(out, promptSection{source: key, name: name, raw: b[n.start:n.end], node: n})
		}
	}
	return out
}
func promptBytes(sections []promptSection) []byte {
	var b []byte
	for _, s := range sections {
		b = append(b, s.raw...)
	}
	return b
}
func optionBytes(b []byte) []byte {
	root, ok := wireRoot(b)
	if !ok {
		return b
	}
	var spans []byteSpan
	for i, f := range root.fields {
		switch f.name {
		case "tools", "system", "systemInstruction", "instructions", "messages", "contents", "input":
			spans = append(spans, memberSpans(root, i)...)
		}
	}
	return withoutSpans(b, spans)
}
func bytePrefix(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}
func sectionChange(prior, current promptSection) string {
	if bytes.Equal(prior.raw, current.raw) {
		return "unchanged"
	}
	if len(current.node.items) > 0 && len(prior.node.items) == 0 {
		return "appended"
	}
	if len(current.node.items) > len(prior.node.items) && len(prior.node.items) > 0 {
		end := prior.node.items[len(prior.node.items)-1].end - prior.node.start
		if end <= len(current.raw) && bytes.Equal(prior.raw[:end], current.raw[:end]) {
			return "appended"
		}
	}
	return "changed"
}
