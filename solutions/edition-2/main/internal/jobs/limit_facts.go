package jobs

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"math"
	"reflect"
	"regexp"
	"sort"
	"time"
)

func validLimitValues(v common.LimitValues) bool {
	if v.Delay == nil && v.Pattern == nil && v.MaxBytes == nil {
		return false
	}
	if v.Delay != nil && (math.IsNaN(*v.Delay) || math.IsInf(*v.Delay, 0) || *v.Delay < 0 || *v.Delay >= float64(math.MaxInt64)/float64(time.Second)) {
		return false
	}
	if v.Pattern != nil {
		if _, err := regexp.Compile(*v.Pattern); err != nil {
			return false
		}
	}
	return v.MaxBytes == nil || *v.MaxBytes > 0
}
func copyLimits(v *common.LimitValues) *common.LimitValues {
	if v == nil {
		return nil
	}
	out := *v
	if v.Delay != nil {
		x := *v.Delay
		out.Delay = &x
	}
	if v.Pattern != nil {
		x := *v.Pattern
		out.Pattern = &x
	}
	if v.MaxBytes != nil {
		x := *v.MaxBytes
		out.MaxBytes = &x
	}
	return &out
}
func overlayValues(s *Service, l *common.Limits, v common.LimitValues) {
	if v.Delay != nil {
		l.Delay = time.Duration(*v.Delay * float64(time.Second))
	}
	if v.MaxBytes != nil {
		l.MaxBytes = *v.MaxBytes
	}
	if v.Pattern != nil {
		l.Pattern = nil
		if *v.Pattern != "" {
			l.Pattern, _ = regexp.Compile(*v.Pattern)
		}
	}
}
func (s *Service) PendingLimits() *common.LimitValues {
	s.mu.Lock()
	defer s.mu.Unlock()
	return copyLimits(s.pending)
}
func (s *Service) ValidateLimitEvent(c common.Context, e common.Event) error {
	if c.Session == nil {
		return nil
	}
	bad := func() error {
		return &common.SessionError{Code: "session_corrupt", Detail: "invalid one-shot limit transition"}
	}
	pending := s.PendingLimits()
	if e.Type == "tool_called" && pending != nil {
		return bad()
	}
	if e.Type == "tool_returned" && e.Tool != nil && !e.Tool.IsError {
		call := c.Calls[e.Tool.CallID]
		if call.Part.Name == "tool_limits" {
			found := false
			for _, f := range c.LimitFacts {
				if f.Kind == "set" && f.CallID == e.Tool.CallID {
					found = true
				}
			}
			if !found {
				return bad()
			}
		}
	}
	if e.Limits == nil {
		return nil
	}
	l := e.Limits
	call, ok := c.Calls[l.CallID]
	if !ok || call.Returned || !validLimitValues(l.Overrides) {
		return bad()
	}
	for _, f := range c.LimitFacts {
		if f.CallID == l.CallID && (f.Kind == "set" && e.Type == "tool_limits_set" || f.Kind == "consumed" && e.Type == "tool_limits_consumed") {
			return bad()
		}
	}
	switch e.Type {
	case "tool_limits_consumed":
		if call.Dispatched || l.Name != call.Part.Name || pending == nil || !reflect.DeepEqual(l.Overrides, *pending) {
			return bad()
		}
	case "tool_limits_set":
		if !call.Dispatched || call.Part.Name != "tool_limits" || l.Name != "" || pending != nil {
			return bad()
		}
		expected, err := s.parent.Registry().LimitCandidate(call.Part)
		if err != nil || !reflect.DeepEqual(expected, l.Overrides) {
			return bad()
		}
	default:
		return bad()
	}
	return nil
}
func (s *Service) ApplyLimitEvent(e common.Event) {
	if e.Limits == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if e.Type == "tool_limits_set" {
		s.pending = copyLimits(&e.Limits.Overrides)
	} else if e.Type == "tool_limits_consumed" {
		s.pending = nil
	}
}
func (s *Service) RestoreLimits(c common.Context, expected *common.LimitValues) error {
	bad := func() error {
		return &common.SessionError{Code: "session_corrupt", Detail: "invalid saved one-shot limits"}
	}
	var pending *common.LimitValues
	// Dispatch/result coordinates let the one-shot owner validate ordering even
	// when those events have left the bounded display window.
	type point struct {
		seq  uint64
		call string
		kind string
		fact *common.LimitFact
	}
	points := []point{}
	callSource := map[string]uint64{}
	resultError := map[string]bool{}
	for _, e := range c.Entries {
		for _, p := range e.Parts {
			if p.Type == "tool_call" {
				callSource[p.CallID] = e.Seq
			}
			if p.Type == "tool_result" {
				resultError[p.CallID] = p.IsError
			}
		}
	}
	for id, call := range c.Calls {
		if call.Dispatched {
			points = append(points, point{seq: call.CalledAt, call: id, kind: "dispatch"})
		}
		if call.Returned {
			points = append(points, point{seq: call.ReturnedAt, call: id, kind: "result"})
		}
	}
	for i := range c.LimitFacts {
		f := &c.LimitFacts[i]
		points = append(points, point{seq: f.Seq, call: f.CallID, kind: f.Kind, fact: f})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].seq < points[j].seq })
	var orderingPending *common.LimitValues
	consuming := ""
	sets := map[string]bool{}
	var previous uint64
	for _, p := range points {
		if p.seq <= previous {
			return bad()
		}
		previous = p.seq
		switch p.kind {
		case "dispatch":
			if orderingPending != nil || consuming != "" && consuming != p.call {
				return bad()
			}
			consuming = ""
		case "result":
			if c.Calls[p.call].Part.Name == "tool_limits" && !resultError[p.call] != sets[p.call] {
				return bad()
			}
		case "set":
			if consuming != "" || orderingPending != nil {
				return bad()
			}
			orderingPending = &p.fact.Overrides
			sets[p.call] = true
		case "consumed":
			if consuming != "" || orderingPending == nil || !reflect.DeepEqual(*orderingPending, p.fact.Overrides) || callSource[p.call] >= p.seq {
				return bad()
			}
			orderingPending = nil
			consuming = p.call
		default:
			return bad()
		}
	}
	var seq uint64
	seen := map[string]bool{}
	for _, f := range c.LimitFacts {
		call, ok := c.Calls[f.CallID]
		if !ok || f.Seq <= seq || f.Seq > c.LastSeq || !validLimitValues(f.Overrides) || seen[f.Kind+"/"+f.CallID] {
			return bad()
		}
		seq = f.Seq
		seen[f.Kind+"/"+f.CallID] = true
		switch f.Kind {
		case "consumed":
			if pending == nil || !reflect.DeepEqual(*pending, f.Overrides) || f.Name != call.Part.Name || call.CalledAt != 0 && f.Seq >= call.CalledAt {
				return bad()
			}
			pending = nil
		case "set":
			if pending != nil || f.Name != "" || call.Part.Name != "tool_limits" || call.CalledAt == 0 || f.Seq <= call.CalledAt || call.ReturnedAt != 0 && f.Seq >= call.ReturnedAt {
				return bad()
			}
			raw, _ := json.Marshal(f.Overrides)
			if !s.parent.Codec().EqualJSON(raw, call.Part.Args) {
				return bad()
			}
			pending = copyLimits(&f.Overrides)
		default:
			return bad()
		}
	}
	if !reflect.DeepEqual(pending, expected) {
		return bad()
	}
	s.mu.Lock()
	s.pending = copyLimits(pending)
	s.mu.Unlock()
	return nil
}

func (s *Service) ResolveConsumed(explicit common.LimitOverrides, pending *common.LimitValues) (common.Limits, bool) {
	limits := common.Limits{Delay: 3 * time.Second, MaxBytes: 16384}
	if pending != nil {
		overlayValues(s, &limits, *pending)
	}
	overlay(s, &limits, explicit)
	return limits, pending != nil
}
