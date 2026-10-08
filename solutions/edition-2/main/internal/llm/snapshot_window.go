package llm

import (
	"bytes"
	"example.com/ensemble/internal/common"
	"reflect"
	"time"
)

// ValidateWindow checks bounded accepted-event witnesses against their owning
// semantic facts. It does not synthesize omitted event history.
func ValidateWindow(owner common.Engine, c common.Context, window []common.WindowEvent) error {
	bad := func() error {
		return &common.SessionError{Code: "session_corrupt", Detail: "saved watch witness disagrees with semantic state"}
	}
	entries := map[uint64]common.Entry{}
	for _, list := range [][]common.Entry{c.Entries, c.Instructions, c.Ephemera, c.Hints} {
		for _, e := range list {
			entries[e.Seq] = e
		}
	}
	responses := map[uint64]common.ResponseFact{}
	for _, r := range c.Responses {
		responses[r.Seq] = r
	}
	for _, item := range window {
		e := item.Event
		at, err := time.Parse(time.RFC3339Nano, e.Time)
		_, offset := at.Zone()
		if err != nil || offset != 0 {
			return bad()
		}
		n := 0
		for _, present := range []bool{e.Message != nil, e.Request != nil, e.Response != nil, e.Tool != nil, e.Redact != nil, e.Error != nil, e.Job != nil, e.Turn != nil, e.Hint != nil, e.Skills != nil, e.Session != nil, e.Limits != nil} {
			if present {
				n++
			}
		}
		if n != 1 {
			return bad()
		}
		switch e.Type {
		case "response_ended":
			if e.Response == nil {
				return bad()
			}
			r := e.Response
			f, ok := responses[e.Seq]
			if !ok || r.Usage == nil || r.From != f.From || !reflect.DeepEqual(r.Requested, f.Requested) || !reflect.DeepEqual(r.ModelReported, f.ModelReported) || *r.Usage != f.Usage || !bytes.Equal(r.RawUsage, f.RawUsage) || r.StopReason != f.StopReason || !reflect.DeepEqual(r.Parts, entries[e.Seq].Parts) {
				return bad()
			}
		case "message_received":
			if e.Message == nil {
				return bad()
			}
			m := *e.Message
			if m.Actor != "human" && m.Actor != "system" || m.Purpose != "dialogue" && m.Purpose != "instruction" && m.Purpose != "ephemeral" || len(m.Parts) == 0 {
				return bad()
			}
			for _, p := range m.Parts {
				if semanticPart(owner, p, false) != "" {
					return bad()
				}
			}
			m.Seq = e.Seq
			if saved, ok := entries[e.Seq]; ok && !reflect.DeepEqual(m, saved) {
				return bad()
			}
		case "tool_called":
			if e.Tool == nil {
				return bad()
			}
			t := e.Tool
			call, ok := c.Calls[t.CallID]
			if !ok || call.CalledAt != e.Seq || t.Name != call.Part.Name || !owner.Agent().Codec().EqualArguments(t.Args, call.Part.Args, true) || t.IsError || len(t.Parts) != 0 {
				return bad()
			}
			if t.Job != nil && (call.JobHandle != t.Job.Handle || jobProblem(owner, *t.Job, nil) != "") {
				return bad()
			}
		case "tool_returned":
			if e.Tool == nil {
				return bad()
			}
			t := e.Tool
			call, ok := c.Calls[t.CallID]
			if !ok || call.ReturnedAt != e.Seq || t.Name != "" || len(t.Args) != 0 {
				return bad()
			}
			parts, err := Clone(owner, t.Parts)
			if err != nil {
				return bad()
			}
			for _, p := range parts {
				if semanticPart(owner, p, true) != "" {
					return bad()
				}
			}
			for _, r := range c.Redactions {
				if r.From <= e.Seq && e.Seq <= r.To {
					for i, p := range parts {
						parts[i] = common.Part{Type: "redacted", Stub: "[redacted]", Ref: p.Ref, Parts: p.Parts[:0]}
					}
				}
			}
			saved := entries[e.Seq]
			if len(saved.Parts) != 1 || saved.Parts[0].IsError != t.IsError || !reflect.DeepEqual(saved.Parts[0].Parts, parts) {
				return bad()
			}
			if t.Job != nil && (call.JobHandle != t.Job.Handle || jobProblem(owner, *t.Job, nil) != "") {
				return bad()
			}
		case "turn_started", "turn_ended":
			if e.Turn == nil {
				return bad()
			}
			t := e.Turn
			f, ok := c.Turns[t.RequestID]
			if !ok {
				return bad()
			}
			if e.Type == "turn_started" {
				if f.Start != e.Seq || f.Index != t.RequestIndex || t.Outcome != "" || !reflect.DeepEqual(f.Policy, t.Policy) {
					return bad()
				}
			} else if f.End != e.Seq || f.Outcome != t.Outcome || t.Policy != nil || t.RequestIndex != 0 {
				return bad()
			}
		case "hint_received":
			if e.Hint == nil || e.Hint.Text == "" {
				return bad()
			}
			found := false
			for _, g := range c.Guidance {
				if g.Seq == e.Seq && g.Kind == "hint" {
					found = true
				}
			}
			if !found {
				return bad()
			}
			if saved, ok := entries[e.Seq]; ok && (len(saved.Parts) != 1 || saved.Parts[0].Text == nil || *saved.Parts[0].Text != e.Hint.Text) {
				return bad()
			}
		case "redacted":
			if e.Redact == nil {
				return bad()
			}
			r := e.Redact
			found := false
			for _, f := range c.Redactions {
				if f.Seq == e.Seq && f.From == r.From && f.To == r.To && f.Level == r.Level && f.Reason == r.Reason {
					found = true
				}
			}
			if !found {
				return bad()
			}
		case "job_ended", "job_killed":
			if e.Job == nil || jobProblem(owner, *e.Job, nil) != "" {
				return bad()
			}
			saved, ok := c.Jobs[e.Job.Handle]
			if !ok || saved != *e.Job { // ExitCode is a value pointer, so compare recursively.
				if !ok || !reflect.DeepEqual(saved, *e.Job) {
					return bad()
				}
			}
		case "error_occurred":
			if e.Error == nil || e.Error.Code == "" || e.Error.Message == "" {
				return bad()
			}
		case "skills_initialized", "skills_changed":
			if e.Skills == nil || !c.SkillMode {
				return bad()
			} // Skills validates its transition witness.
		default:
			return bad()
		}
	}
	return nil
}
