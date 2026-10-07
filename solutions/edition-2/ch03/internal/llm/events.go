package llm

import (
	"encoding/json"
	"errors"
	"example.com/ensemble/internal/common"
	"reflect"
	"strings"
	"time"
)

func Clone[T any](owner common.Engine, value T) (T, error) {
	var out T
	data, err := json.Marshal(value)
	if err == nil {
		err = json.Unmarshal(data, &out)
	}
	if err != nil {
		return out, failure(owner, "cannot copy malformed owned data")
	}
	return out, nil
}
func Text(text string) common.Part { return common.Part{Type: "text", Text: &text} }
func validProvenance(owner common.Engine, p *common.Provenance) bool {
	if p == nil || p.Model == "" {
		return false
	}
	if p.Surface == "generatecontent" {
		p.Surface = "generate_content"
	}
	return (p.Vendor == "anthropic" && p.Surface == "messages") || (p.Vendor == "openai" && p.Surface == "chat_completions") || (p.Vendor == "gemini" && p.Surface == "generate_content")
}
func object(owner common.Engine, raw json.RawMessage) bool {
	var v map[string]json.RawMessage
	return len(raw) > 0 && json.Unmarshal(raw, &v) == nil && v != nil
}

// partProblem returns only static categories, never content supplied by a caller.
func partProblem(owner common.Engine, p *common.Part, result bool) string {
	if p.Ref != nil {
		if p.Ref.Kind < 1 || p.Ref.Kind > 3 {
			return "invalid reference kind"
		}
		if p.Ref.Locator == "" {
			return "missing reference locator"
		}
	}
	if result && p.Type != "text" && p.Type != "blob" && p.Type != "redacted" {
		return "unsupported tool-result part"
	}
	switch p.Type {
	case "text":
		if p.Text != nil && ((p.From == nil && len(p.Opaque) == 0) || (validProvenance(owner, p.From) && len(p.Opaque) > 0 && json.Valid(p.Opaque))) {
			return ""
		}
		return "invalid text part or replay provenance"
	case "tool_call":
		if p.CallID != "" && p.Name != "" && object(owner, p.Args) && validProvenance(owner, p.From) && (len(p.Opaque) == 0 || json.Valid(p.Opaque)) {
			return ""
		}
		return "invalid tool-call fields"
	case "tool_result":
		if p.CallID == "" || p.Parts == nil {
			return "invalid tool-result fields"
		}
		for i := range p.Parts {
			if reason := partProblem(owner, &p.Parts[i], true); reason != "" {
				return reason
			}
		}
		return ""
	case "opaque":
		if validProvenance(owner, p.From) && len(p.Data) > 0 && json.Valid(p.Data) {
			return ""
		}
		return "invalid opaque part or provenance"
	case "blob":
		if p.MIME != "" && p.Ref != nil {
			return ""
		}
		return "missing blob MIME or reference"
	case "redacted":
		if p.Stub != "" {
			return ""
		}
		return "missing redaction stub"
	}
	return "unknown part type"
}
func unresolved(owner common.Engine, c common.Context) bool {
	for _, call := range c.Calls {
		if !call.Returned {
			return true
		}
	}
	return false
}
func CanPrompt(owner common.Engine, c common.Context) error {
	if c.Pending != nil || c.Active {
		return failure(owner, "Agent already has a pending request")
	}
	if unresolved(owner, c) {
		return failure(owner, "unanswered tool calls require results before another prompt")
	}
	return nil
}
func Validate(owner common.Engine, c common.Context, e *common.Event) error {
	bad := func(reason string) error {
		err := &validationError{reason: reason}
		owner.Agent().Ensemble().Logf("%s", err)
		return err
	}
	t, err := time.Parse(time.RFC3339Nano, e.Time)
	_, offset := t.Zone()
	if e.Seq == 0 || e.Seq <= c.LastSeq {
		return bad("sequence must be positive and strictly increasing")
	}
	if err != nil || offset != 0 {
		return bad("invalid UTC timestamp")
	}
	n := 0
	for _, present := range []bool{e.Message != nil, e.Request != nil, e.Response != nil, e.Tool != nil, e.Redact != nil, e.Error != nil} {
		if present {
			n++
		}
	}
	if n != 1 {
		return bad("exactly one event payload is required")
	}
	switch e.Type {
	case "message_received":
		m := e.Message
		if m == nil || m.Parts == nil {
			return bad("invalid event or conversation transition")
		}
		if m.Purpose == "" {
			if m.Actor == "system" {
				m.Purpose = "ephemeral"
			} else {
				m.Purpose = "dialogue"
			}
		}
		if m.Actor == "human" {
			if m.Purpose != "dialogue" || c.Pending != nil || c.Active {
				return bad("invalid event or conversation transition")
			}
		} else if m.Actor != "system" || (m.Purpose != "instruction" && m.Purpose != "ephemeral") {
			return bad("invalid event or conversation transition")
		}
		for i := range m.Parts {
			if reason := partProblem(owner, &m.Parts[i], false); reason != "" {
				return bad(reason)
			}
			if m.Parts[i].Type == "tool_call" || m.Parts[i].Type == "tool_result" {
				return bad("invalid event or conversation transition")
			}
			if m.Actor == "system" && m.Parts[i].Type != "text" {
				return bad("invalid event or conversation transition")
			}
		}
	case "request_sent":
		r := e.Request
		if r == nil || !validProvenance(owner, &r.To) || c.Active || unresolved(owner, c) || (c.Pending == nil && !c.Continuation) {
			return bad("invalid event or conversation transition")
		}
		if r.Ephemera == nil {
			r.Ephemera = make([]uint64, 0, len(c.Ephemera))
			for _, entry := range c.Ephemera {
				r.Ephemera = append(r.Ephemera, entry.Seq)
			}
		}
		seen := map[uint64]bool{}
		for _, seq := range r.Ephemera {
			found := false
			for _, entry := range c.Ephemera {
				if seq == entry.Seq {
					found = true
				}
			}
			if !found || seen[seq] {
				return bad("invalid event or conversation transition")
			}
			seen[seq] = true
		}
	case "response_started", "response_ended":
		r := e.Response
		if r == nil {
			return bad("response payload is required")
		}
		if !validProvenance(owner, &r.From) || (r.Requested != nil && !validProvenance(owner, r.Requested)) {
			return bad("invalid response provenance")
		}
		if unresolved(owner, c) {
			return bad("response arrived before tool results")
		}
		if !c.Active && c.Pending == nil && !c.Continuation {
			return bad("unsolicited response")
		}
		if e.Type == "response_started" {
			break
		}
		if r.Parts == nil || r.Usage == nil || r.Usage.Input < 0 || r.Usage.CacheRead < 0 || r.Usage.CacheWrite < 0 || r.Usage.Output < 0 {
			return bad("invalid event or conversation transition")
		}
		visible := false
		ids := map[string]bool{}
		for i := range r.Parts {
			p := &r.Parts[i]
			if reason := partProblem(owner, p, false); reason != "" {
				return bad(reason)
			}
			if p.Type == "text" {
				visible = true
			}
			if p.Type == "tool_call" {
				visible = true
				if _, ok := c.Calls[p.CallID]; ok || ids[p.CallID] {
					return bad("invalid event or conversation transition")
				}
				ids[p.CallID] = true
			}
			if p.Type == "tool_result" {
				return bad("invalid event or conversation transition")
			}
		}
		if !visible {
			return bad("invalid event or conversation transition")
		}
	case "tool_called", "tool_returned":
		t := e.Tool
		if t == nil {
			return bad("invalid event or conversation transition")
		}
		call, ok := c.Calls[t.CallID]
		if !ok || call.Returned {
			return bad("invalid event or conversation transition")
		}
		if e.Type == "tool_called" {
			var a, b any
			_ = json.Unmarshal(t.Args, &a)
			_ = json.Unmarshal(call.Part.Args, &b)
			if call.Dispatched || t.Name != call.Part.Name || !object(owner, t.Args) || !reflect.DeepEqual(a, b) {
				return bad("invalid event or conversation transition")
			}
		} else {
			if t.Parts == nil {
				return bad("invalid event or conversation transition")
			}
			for i := range t.Parts {
				if reason := partProblem(owner, &t.Parts[i], true); reason != "" {
					return bad(reason)
				}
			}
		}
	case "redacted":
		r := e.Redact
		if r == nil || r.From == 0 || r.From > r.To || r.To > c.LastSeq || r.Level != "redact_result" || r.Reason == "" {
			return bad("invalid event or conversation transition")
		}
		found := false
		for _, entry := range c.Entries {
			if entry.Seq >= r.From && entry.Seq <= r.To {
				for _, p := range entry.Parts {
					if p.Type == "tool_result" {
						found = true
					}
				}
			}
		}
		if !found {
			return bad("invalid event or conversation transition")
		}
	case "error_occurred":
		if e.Error == nil || e.Error.Code == "" || e.Error.Message == "" {
			return bad("invalid event or conversation transition")
		}
	default:
		return bad("invalid event or conversation transition")
	}
	return nil
}

// Apply runs only after validation and durable append; it cannot fail halfway.
// Its event is a separate owned copy so redaction cannot rewrite history bytes.
func Apply(owner common.Engine, c *common.Context, e common.Event) {
	c.LastSeq = e.Seq
	if c.Calls == nil {
		c.Calls = map[string]common.CallState{}
	}
	switch e.Type {
	case "message_received":
		m := *e.Message
		m.Seq = e.Seq
		switch m.Purpose {
		case "instruction":
			c.Instructions = append(c.Instructions, m)
		case "ephemeral":
			c.Ephemera = append(c.Ephemera, m)
		default:
			c.Pending = &m
		}
	case "request_sent":
		c.Active = true
		keep := c.Ephemera[:0]
		for _, entry := range c.Ephemera {
			consume := false
			for _, seq := range e.Request.Ephemera {
				if seq == entry.Seq {
					consume = true
				}
			}
			if !consume {
				keep = append(keep, entry)
			}
		}
		c.Ephemera = keep
	case "response_ended":
		// A validated response commits the pair and accounting through one append path.
		// Failed attempts retain their log facts without poisoning completed dialogue.
		if c.Pending != nil {
			c.Entries = append(c.Entries, *c.Pending)
		}
		c.Pending = nil
		c.Active = false
		c.Continuation = false
		c.Entries = append(c.Entries, common.Entry{Seq: e.Seq, Actor: "agent", Purpose: "dialogue", Parts: e.Response.Parts})
		for _, p := range e.Response.Parts {
			if p.Type == "tool_call" {
				c.Calls[p.CallID] = common.CallState{Part: p}
			}
		}
	case "tool_called":
		call := c.Calls[e.Tool.CallID]
		call.Dispatched = true
		c.Calls[e.Tool.CallID] = call
	case "tool_returned":
		call := c.Calls[e.Tool.CallID]
		call.Returned = true
		c.Calls[e.Tool.CallID] = call
		c.Entries = append(c.Entries, common.Entry{Seq: e.Seq, Actor: "tool", Purpose: "dialogue", Parts: []common.Part{{Type: "tool_result", CallID: e.Tool.CallID, Parts: e.Tool.Parts, IsError: e.Tool.IsError}}})
		c.Continuation = !unresolved(owner, *c)
	case "error_occurred":
		c.Pending = nil
		c.Active = false
		c.Continuation = false
	case "redacted":
		for i := range c.Entries {
			entry := &c.Entries[i]
			if entry.Seq < e.Redact.From || entry.Seq > e.Redact.To {
				continue
			}
			for j := range entry.Parts {
				p := &entry.Parts[j]
				if p.Type != "tool_result" {
					continue
				}
				for k, old := range p.Parts {
					p.Parts[k] = common.Part{Type: "redacted", Stub: "[redacted]", Ref: old.Ref}
				}
			}
		}
	}
}
func TextAnswer(owner common.Engine, parts []common.Part) string {
	var out strings.Builder
	for _, p := range parts {
		if p.Type == "text" && p.Text != nil {
			out.WriteString(*p.Text)
		}
	}
	return out.String()
}
func Route(owner common.Engine, config common.Config) common.Provenance {
	surface := "messages"
	if config.Vendor == "openai" {
		surface = "chat_completions"
	}
	if config.Vendor == "gemini" {
		surface = "generate_content"
	}
	return common.Provenance{Vendor: config.Vendor, Model: config.Model, Surface: surface}
}
func ValidateConfig(owner common.Engine, c common.Config, live bool) error {
	if c.Vendor != "anthropic" && c.Vendor != "openai" && c.Vendor != "gemini" {
		return failure(owner, "unknown LLM_VENDOR")
	}
	if c.Model == "" {
		return failure(owner, "%s_MODEL or LLM_MODEL is required; discover models with GET /v1/models (Gemini: /v1beta/models)", strings.ToUpper(c.Vendor))
	}
	if live && c.APIKey == "" {
		return failure(owner, "%s_API_KEY or LLM_API_KEY is required", strings.ToUpper(c.Vendor))
	}
	for _, tool := range c.Tools {
		if tool.Name == "" || !object(owner, tool.Schema) {
			return failure(owner, "invalid tool declaration")
		}
	}
	return nil
}

// Only our static validation categories may be included with a log line number.
// Other errors are deliberately not interpolated: their text may be external.
type validationError struct{ reason string }

func (e *validationError) Error() string { return e.reason }
func ParseEventError(owner common.Engine, line int, cause error) error {
	reason := "invalid event data"
	var validation *validationError
	if errors.As(cause, &validation) {
		reason = validation.reason
	}
	return failure(owner, "invalid event log at line %d: %s", line, reason)
}
