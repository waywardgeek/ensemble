package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"example.com/ensemble/internal/common"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

func Clone[T any](owner common.Engine, value T) (T, error) {
	// Copying never encodes: capture remains proportional copying on Actor, while
	// validation and bounded serialization belong to their explicit boundaries.
	v := cloneValue(reflect.ValueOf(value), false)
	if !v.IsValid() {
		var zero T
		return zero, nil
	}
	return v.Interface().(T), nil
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
	if p.ArgumentsText != nil && p.Type != "tool_call" {
		return "argument replay string requires a tool call"
	}
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
			if p.ArgumentsText != nil {
				if p.From.Vendor != "openai" {
					return "invalid argument replay provenance"
				}
				original := *p.ArgumentsText
				// Marshal RawMessage only for this correspondence check: it
				// retains member order/duplicates and number tokens while
				// removing outer whitespace and applying JSON string escaping.
				left, le := json.Marshal(json.RawMessage(original))
				right, re := json.Marshal(p.Args)
				if le != nil || re != nil || !bytes.Equal(left, right) {
					return "argument replay string disagrees with call"
				}
			}
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
	if c.Index.Ready {
		return c.Index.Unresolved != 0
	}
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
	if c.Session != nil {
		var parts []common.Part
		result := false
		if e.Message != nil {
			parts = e.Message.Parts
		}
		if e.Response != nil {
			parts = e.Response.Parts
		}
		if e.Tool != nil {
			parts = e.Tool.Parts
			result = true
		}
		for _, p := range parts {
			if reason := semanticPart(owner, p, result); reason != "" {
				return bad(reason)
			}
		}
	}
	n := 0
	for _, present := range []bool{e.Message != nil, e.Request != nil, e.Response != nil, e.Tool != nil, e.Redact != nil, e.Error != nil, e.Job != nil, e.Turn != nil, e.Hint != nil, e.Skills != nil, e.Session != nil, e.Limits != nil} {
		if present {
			n++
		}
	}
	if n != 1 {
		return bad("exactly one event payload is required")
	}
	if c.Session != nil && e.Seq != c.LastSeq+1 {
		return bad("session sequence gap")
	}
	switch e.Type {
	case "session_initialized":
		if c.LastSeq != 0 || e.Seq != 1 || e.Session == nil || e.Session.Identity == nil || e.Session.OriginAsOf != 0 || e.Session.OriginSHA256 != "" || e.Session.HighWatermarks != nil {
			return bad("invalid session initializer")
		}
 if (e.Session.Version==2)!=(len(e.Session.Identity.MCPBindings)>0)||e.Session.Version!=0&&e.Session.Version!=2{return bad("invalid session format")}
	case "session_anchor":
		if c.Session == nil || e.Session == nil || e.Session.Identity != nil || e.Session.Version != c.Session.Version || e.Session.SessionID != c.Session.SessionID || e.Session.OriginAsOf != c.LastSeq || e.Session.HighWatermarks == nil || e.Session.HighWatermarks.Event != c.LastSeq {
			return bad("invalid session anchor")
		}
	case "tool_limits_set", "tool_limits_consumed":
		if c.Session == nil || e.Limits == nil {
			return bad("limits facts require a session")
		}
	case "skills_initialized", "skills_changed":
		if e.Skills == nil || (e.Type == "skills_initialized") != (e.Skills.Action == "initialize") || e.Type == "skills_initialized" && (c.LastSeq != 0 && !(c.Session != nil && c.LastSeq == 1)) {
			return bad("invalid skill event placement")
		}
		// Skills validates the graph through the composition root before this reducer.
	case "turn_started":
		if c.Session != nil && (e.Turn == nil || e.Turn.RequestIndex == 0 || e.Turn.RequestIndex <= c.RequestCursor) {
			return bad("session request index must advance")
		}
		if e.Turn != nil && e.Turn.Policy != nil {
			p := e.Turn.Policy
			effective := p.MaxModelRequests
			if effective == 0 {
				effective = common.DefaultMaxModelRequests
			}
			if p.MaxModelRequests < 0 || p.MaxModelRequests > 256 || p.EffectiveMaxModelRequests != effective {
				return bad("invalid turn policy capture")
			}
		}
		if e.Turn == nil || e.Turn.RequestID == "" || e.Turn.Outcome != "" || c.TurnID != "" || c.Active || c.Pending != nil || unresolved(owner, c) || c.TurnIDs[e.Turn.RequestID] {
			return bad("invalid turn start")
		}
	case "hint_received":
		if e.Hint == nil || e.Hint.RequestID == "" || e.Hint.RequestID != c.TurnID || strings.TrimSpace(e.Hint.Text) == "" {
			return bad("hint requires its active turn")
		}
	case "turn_ended":
		if e.Turn == nil || e.Turn.RequestID == "" || e.Turn.RequestID != c.TurnID || c.Active || unresolved(owner, c) {
			return bad("invalid turn end")
		}
		switch e.Turn.Outcome {
		case "success":
			if !c.FinalResponse {
				return bad("success requires final response")
			}
		case "interrupted", "canceled", "error", "round_limit", "stopped":
		default:
			return bad("unknown turn outcome")
		}
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
			if m.Purpose != "dialogue" || c.Pending != nil || c.Active || (c.ExplicitTurns && c.TurnID == "") {
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
		if c.SkillMode && r != nil && r.Configuration != nil && r.Configuration.System != "" {
			return bad("skill request capture must omit primary bytes")
		}
		if r != nil && r.Delivery != "" && r.Delivery != "stream" && r.Delivery != "plain" {
			return bad("invalid request delivery")
		}
		if r == nil || !validProvenance(owner, &r.To) || c.Active || unresolved(owner, c) || (c.Pending == nil && !c.Continuation) {
			return bad("invalid event or conversation transition")
		}
		if r.Ephemera == nil {
			r.Ephemera = make([]uint64, 0, len(c.Ephemera))
			for _, entry := range c.Ephemera {
				r.Ephemera = append(r.Ephemera, entry.Seq)
			}
		}
		expected := make([]uint64, 0, len(c.Hints))
		for _, hint := range c.Hints {
			expected = append(expected, hint.Seq)
		}
		if r.Hints == nil {
			r.Hints = expected
		}
		if !reflect.DeepEqual(expected, r.Hints) {
			return bad("request must consume exactly pending hints in order")
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
			if call.Dispatched || t.Name != call.Part.Name || !object(owner, t.Args) || !owner.Agent().Codec().EqualArguments(t.Args, call.Part.Args, c.Session != nil) {
				return bad("invalid event or conversation transition")
			}
			if t.Job != nil {
				if _, exists := c.Jobs[t.Job.Handle]; exists || t.Job.Status != "running" || t.Job.Bytes != 0 {
					return bad("job creation must introduce a new running handle")
				}
				if reason := jobProblem(owner, *t.Job, nil); reason != "" {
					return bad(reason)
				}
			}
		} else {
			if call.JobHandle != 0 {
				if t.Job == nil || t.Job.Handle != call.JobHandle {
					return bad("result job does not belong to original call")
				}
				previous := c.Jobs[call.JobHandle]
				if reason := jobProblem(owner, *t.Job, &previous); reason != "" {
					return bad(reason)
				}
				if previous.Status != "running" && !reflect.DeepEqual(previous, *t.Job) {
					return bad("result disagrees with terminal job")
				}
				if previous.Status == "running" && t.Job.Status != "running" {
					return bad("terminal result requires terminal event")
				}
			} else if t.Job != nil {
				return bad("result invents a job")
			}
			if t.Parts == nil {
				return bad("invalid event or conversation transition")
			}
			for i := range t.Parts {
				if reason := partProblem(owner, &t.Parts[i], true); reason != "" {
					return bad(reason)
				}
			}
		}
	case "job_ended", "job_killed":
		if e.Job == nil {
			return bad("job payload required")
		}
		previous, ok := c.Jobs[e.Job.Handle]
		if !ok || previous.Status != "running" {
			return bad("terminal job must refer to an existing running job")
		}
		expected := "done"
		if e.Type == "job_killed" {
			expected = "killed"
		}
		if e.Job.Status != expected {
			return bad("terminal event status mismatch")
		}
		if reason := jobProblem(owner, *e.Job, &previous); reason != "" {
			return bad(reason)
		}
	case "redacted":
		r := e.Redact
		if r == nil {
			return bad("redaction payload is required")
		}
		if r.From == 0 || r.From > r.To || r.To > c.LastSeq {
			return bad("redaction span must name positive ordered sequences already in this log")
		}
		if r.Level != "redact_result" {
			return bad("redaction level must be redact_result")
		}
		if strings.TrimSpace(r.Reason) == "" {
			return bad("redaction reason must be nonempty")
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
			return bad("redaction span contains no tool results; inspect history for tool_returned events")
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
	c.Index = nextIndex(owner, *c, e)
	recordSemantic(owner, c, e)
	c.LastSeq = e.Seq
	if c.Calls == nil {
		c.Calls = map[string]common.CallState{}
	}
	switch e.Type {
	case "skills_initialized", "skills_changed":
		c.SkillMode = true
		for _, a := range e.Skills.Activated {
			if a.Type == "primary" {
				c.SkillPrimary = a.Body
				continue
			}
			m := common.Entry{Seq: e.Seq, Actor: "system", Purpose: "skill", SkillName: a.Name, Activation: a.Activation, Parts: []common.Part{}}
			if a.Body != "" {
				m.Parts = append(m.Parts, Text(fmt.Sprintf("[skill %s activation %d]\n%s\n[/skill]", a.Name, a.Activation, a.Body)))
			}
			if c.SkillBatch != 0 {
				m.Anchor = c.SkillBatch
				c.DeferredSkills = append(c.DeferredSkills, m)
			} else if c.Pending != nil {
				c.PendingSkills = append(c.PendingSkills, m)
			} else {
				c.Entries = append(c.Entries, m)
			}
		}
	case "turn_started":
		c.ExplicitTurns = true
		c.TurnID = e.Turn.RequestID
		c.FinalResponse = false
		if c.TurnIDs == nil {
			c.TurnIDs = map[string]bool{}
		}
		c.TurnIDs[c.TurnID] = true
	case "turn_ended":
		c.TurnID = ""
		c.Continuation = false
		c.Pending = nil
	case "hint_received":
		c.Hints = append(c.Hints, common.Entry{Seq: e.Seq, Actor: "human", Purpose: "hint", Anchor: c.SkillBatch, Parts: []common.Part{Text(e.Hint.Text)}})
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
		if c.SkillMode {
			entries := c.Entries[:0]
			for _, m := range c.Entries {
				if m.Purpose != "hint" {
					entries = append(entries, m)
				}
			}
			c.Entries = entries
		}
		c.Hints = nil
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
		c.Entries = append(c.Entries, c.PendingSkills...)
		c.PendingSkills = nil
		c.Active = false
		c.Continuation = false
		c.Entries = append(c.Entries, common.Entry{Seq: e.Seq, Actor: "agent", Purpose: "dialogue", Parts: e.Response.Parts})
		c.FinalResponse = true
		for _, p := range e.Response.Parts {
			if p.Type == "tool_call" {
				c.FinalResponse = false
				c.Calls[p.CallID] = common.CallState{Part: p}
			}
		}
		if c.SkillMode && !c.FinalResponse {
			c.SkillBatch = e.Seq
			for i := range c.Hints {
				if c.Hints[i].Anchor == 0 {
					c.Hints[i].Anchor = e.Seq
				}
			}
		}
	case "tool_called":
		call := c.Calls[e.Tool.CallID]
		if e.Tool.Job != nil {
			if c.Jobs == nil {
				c.Jobs = map[uint64]common.JobSnapshot{}
			}
			c.Jobs[e.Tool.Job.Handle] = *e.Tool.Job
			call.JobHandle = e.Tool.Job.Handle
		}
		call.Dispatched = true
		call.CalledAt = e.Seq
		c.Calls[e.Tool.CallID] = call
	case "job_ended", "job_killed":
		c.Jobs[e.Job.Handle] = *e.Job
	case "tool_returned":
		if e.Tool.Job != nil {
			c.Jobs[e.Tool.Job.Handle] = *e.Tool.Job
		}
		call := c.Calls[e.Tool.CallID]
		call.Returned = true
		call.ReturnedAt = e.Seq
		c.Calls[e.Tool.CallID] = call
		c.Entries = append(c.Entries, common.Entry{Seq: e.Seq, Actor: "tool", Purpose: "dialogue", Parts: []common.Part{{Type: "tool_result", CallID: e.Tool.CallID, Parts: e.Tool.Parts, IsError: e.Tool.IsError}}})
		c.Continuation = !unresolved(owner, *c)
		if c.Continuation && c.SkillBatch != 0 {
			material := append([]common.Entry{}, c.DeferredSkills...)
			for _, h := range c.Hints {
				if h.Anchor == c.SkillBatch {
					material = append(material, h)
				}
			}
			sort.SliceStable(material, func(i, j int) bool { return material[i].Seq < material[j].Seq })
			c.Entries = append(c.Entries, material...)
			c.DeferredSkills = nil
			c.SkillBatch = 0
		}
	case "error_occurred":
		c.Entries = append(c.Entries, c.PendingSkills...)
		c.PendingSkills = nil
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

func jobProblem(owner common.Engine, job common.JobSnapshot, previous *common.JobSnapshot) string {
	if job.Handle == 0 || job.Bytes < 0 || job.Output.Kind != 3 || job.Output.Locator != fmt.Sprintf("cr/io/%d", job.Handle) {
		return "invalid job identity, locator, or byte count"
	}
	if job.Status != "running" && job.Status != "done" && job.Status != "killed" {
		return "invalid job status"
	}
	if job.ExitCode != nil && job.Status != "done" {
		return "exit code requires normally completed job"
	}
	if job.Cwd != "" && !filepath.IsAbs(job.Cwd) {
		return "job cwd must be absolute"
	}
	if job.Status == "killed" {
		if job.Reason != "kill_job" && job.Reason != "shutdown" {
			return "invalid kill reason"
		}
	} else if job.Reason != "" {
		return "kill reason on non-killed job"
	}
	if previous != nil && (job.Output != previous.Output || job.Bytes < previous.Bytes || (previous.Cwd != "" && job.Cwd != previous.Cwd)) {
		return "job locator, cwd, or bytes changed inconsistently"
	}
	return ""
}

// SnapshotContext gives callers one stable representation for empty structural
// collections, whether reduction followed a checkpoint or the complete log.
// Raw JSON keeps its exact bytes and absence; Clone of events never normalizes.
func SnapshotContext(owner common.Engine, value common.Context) common.Context {
	return cloneValue(reflect.ValueOf(value), true).Interface().(common.Context)
}
