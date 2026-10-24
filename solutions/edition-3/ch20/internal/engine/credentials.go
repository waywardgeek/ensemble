package engine

import (
	"bytes"
	"encoding/json"
	"ensemble/internal/common"
	"slices"
	"strings"
)

// Credential values belong to one exchange (including its renewal), not to a
// global registry. These helpers edit exact echoes, never secret-like patterns.
func redactCredentialJSON(raw json.RawMessage, replacer *strings.Replacer) json.RawMessage {
	var out []byte
	last := 0
	for i := 0; i < len(raw); i++ {
		if raw[i] != '"' {
			continue
		}
		start := i
		for i++; i < len(raw); i++ {
			if raw[i] == '\\' {
				i++
				continue
			}
			if raw[i] == '"' {
				break
			}
		}
		if i >= len(raw) {
			break
		}
		var text string
		if json.Unmarshal(raw[start:i+1], &text) != nil {
			continue
		}
		clean := replacer.Replace(text)
		if clean == text {
			continue
		}
		encoded, _ := json.Marshal(clean)
		out = append(out, raw[last:start]...)
		out = append(out, encoded...)
		last = i + 1
	}
	if last == 0 {
		return raw
	}
	return append(out, raw[last:]...)
}
func redactCredentialParts(parts common.PartList, replacer *strings.Replacer) {
	for i := range parts {
		p := &parts[i]
		p.Text = replacer.Replace(p.Text)
		p.Name = replacer.Replace(p.Name)
		p.Namespace = replacer.Replace(p.Namespace)
		p.CallID = replacer.Replace(p.CallID)
		p.MIME = replacer.Replace(p.MIME)
		p.Stub = replacer.Replace(p.Stub)
		if p.Ref != nil {
			p.Ref.Locator = replacer.Replace(p.Ref.Locator)
		}
		if p.From != nil {
			p.From.Model = replacer.Replace(p.From.Model)
		}
		p.Args = redactCredentialJSON(p.Args, replacer)
		p.Data = redactCredentialJSON(p.Data, replacer)
		p.Opaque = redactCredentialJSON(p.Opaque, replacer)
		redactCredentialParts(p.Parts, replacer)
	}
}
func redactCredentialEvents(events []common.Event, replacer, errorReplacer *strings.Replacer) {
	for i := range events {
		e := &events[i]
		if e.Response != nil {
			redactCredentialParts(e.Response.Parts, replacer)
			e.Response.Thinking = replacer.Replace(e.Response.Thinking)
			e.Response.From.Model = replacer.Replace(e.Response.From.Model)
		}
		if e.Error != nil {
			e.Error.Message = errorReplacer.Replace(e.Error.Message)
			e.Error.Code = errorReplacer.Replace(e.Error.Code)
		}
	}
}

type credentialDeltaKey struct {
	part uint64
	kind common.DeltaKind
}

// A fragment keeps its original publication position and metadata even when
// some bytes must wait for the rest of a possible credential in the same part.
type credentialFragment struct {
	delta           common.PartDelta
	raw             string
	ready           strings.Builder
	done, published bool
}
type credentialDeltas struct {
	patterns []string
	replacer *strings.Replacer
	pending  map[credentialDeltaKey][]*credentialFragment
	queue    []*credentialFragment
	emit     func(common.PartDelta)
}

// Include only literal spellings of the known values, including JSON's two
// standard HTML-escaping modes. This is not a provider-key pattern detector.
func credentialSpellings(tokens []string) []string {
	var patterns []string
	for _, token := range tokens {
		if token == "" {
			continue
		}
		patterns = append(patterns, token)
		encoded, _ := json.Marshal(token)
		escaped := string(encoded[1 : len(encoded)-1])
		if escaped != token {
			patterns = append(patterns, escaped)
		}
		var plain bytes.Buffer
		encoder := json.NewEncoder(&plain)
		encoder.SetEscapeHTML(false)
		_ = encoder.Encode(token)
		unescaped := plain.String()[1 : plain.Len()-2]
		if unescaped != token && unescaped != escaped {
			patterns = append(patterns, unescaped)
		}
	}
	// Replacer gives priority to the first matching pair. Normalize once so
	// complete fields and the incremental scanner both choose the longest exact
	// spelling at a position, regardless of current/renewed credential order.
	slices.SortFunc(patterns, func(a, b string) int {
		if len(a) != len(b) {
			return len(b) - len(a)
		}
		return strings.Compare(a, b)
	})
	return slices.Compact(patterns)
}
func credentialReplacer(patterns []string) *strings.Replacer {
	pairs := make([]string, 0, 2*len(patterns))
	for _, pattern := range patterns {
		pairs = append(pairs, pattern, "[REDACTED]")
	}
	// One pass: an inserted marker must never be processed as another token.
	return strings.NewReplacer(pairs...)
}
func newCredentialDeltas(tokens []string, emit func(common.PartDelta)) *credentialDeltas {
	patterns := credentialSpellings(tokens)
	return &credentialDeltas{patterns: patterns, replacer: credentialReplacer(patterns), emit: emit, pending: map[credentialDeltaKey][]*credentialFragment{}}
}

// Resolve one part's undecided bytes into their original fragment slots. Only
// a suffix that might still be an exact credential remains pending. A matched
// credential becomes one marker at its first byte's original position.
func (d *credentialDeltas) resolve(key credentialDeltaKey, final bool) {
	pending := d.pending[key]
	var joined strings.Builder
	for _, fragment := range pending {
		joined.WriteString(fragment.raw)
	}
	text := joined.String()
	consume := func(n int, redact bool) {
		if redact {
			pending[0].ready.WriteString("[REDACTED]")
		}
		for n > 0 {
			fragment := pending[0]
			size := min(n, len(fragment.raw))
			if !redact {
				fragment.ready.WriteString(fragment.raw[:size])
			}
			fragment.raw = fragment.raw[size:]
			n -= size
			if fragment.raw == "" {
				fragment.done = true
				pending[0] = nil
				pending = pending[1:]
			}
		}
	}
	start := 0
	for i := 0; i < len(text); {
		match := 0
		partial := false
		for _, token := range d.patterns {
			if strings.HasPrefix(text[i:], token) && len(token) > match {
				match = len(token)
			}
			if !final && len(text[i:]) < len(token) && strings.HasPrefix(token, text[i:]) {
				partial = true
			}
		}
		if partial {
			consume(i-start, false)
			start = len(text)
			break
		}
		if match > 0 {
			consume(i-start, false)
			consume(match, true)
			i += match
			start = i
		} else {
			i++
		}
	}
	consume(len(text)-start, false)
	if len(pending) == 0 {
		delete(d.pending, key)
	} else {
		d.pending[key] = pending
	}
}

// Never publish a later part ahead of an unresolved earlier fragment: arrival
// order is also the actor's explanation-before-tool policy evidence. Ordinary
// unambiguous streams drain immediately. If an ambiguous part never resumes,
// later payload necessarily waits until parse completion (including errors).
func (d *credentialDeltas) drain() {
	for len(d.queue) > 0 {
		fragment := d.queue[0]
		if fragment.ready.Len() > 0 {
			delta := fragment.delta
			delta.Chunk = fragment.ready.String()
			d.emit(delta)
			fragment.ready.Reset()
			fragment.published = true
		}
		if !fragment.done {
			return
		}
		if !fragment.published {
			d.emit(fragment.delta)
		}
		d.queue[0] = nil
		d.queue = d.queue[1:]
	}
	d.queue = nil
}
func (d *credentialDeltas) push(delta common.PartDelta) {
	if d.emit == nil {
		return
	}
	key := credentialDeltaKey{delta.PartID, delta.Kind}
	fragment := &credentialFragment{delta: delta, raw: delta.Chunk, done: delta.Chunk == ""}
	fragment.delta.Chunk = ""
	fragment.delta.Name = d.replacer.Replace(delta.Name)
	d.queue = append(d.queue, fragment)
	if fragment.raw != "" {
		d.pending[key] = append(d.pending[key], fragment)
		d.resolve(key, false)
	}
	d.drain()
}
func (d *credentialDeltas) flush() {
	if d.emit == nil {
		return
	}
	for key := range d.pending {
		d.resolve(key, true)
	}
	d.drain()
}
