package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"time"

	"example.com/ensemble/internal/common"
)

// Tool names and wire validation belong here. Jobs sees validated durations,
// compiled expressions and report requests, never tool schemas or JSON fields.
func (r *Registry) ResolveLimits(call common.Part) (common.Limits, string, error) {
	return r.resolve(call, nil, false)
}
func (r *Registry) ResolveConsumed(call common.Part, pending *common.LimitValues) (common.Limits, string, error) {
	return r.resolve(call, pending, true)
}
func (r *Registry) resolve(call common.Part, pending *common.LimitValues, recorded bool) (common.Limits, string, error) {
	var supplied common.LimitOverrides
	var err error
	switch call.Name {
	case "run_command", "wait_for_job", "send_input", "tool_limits":
		supplied, err = r.limitOverrides(call.Args)
	}
	if err != nil {
		supplied = common.LimitOverrides{}
	}
	var limits common.Limits
	var consumed bool
	if recorded {
		limits, consumed = r.parent.Jobs().ResolveConsumed(supplied, pending)
	} else {
		limits, consumed = r.parent.Jobs().Resolve(supplied)
	}
	note := ""
	if consumed {
		if err != nil {
			note = fmt.Sprintf("tool_limits consumed by %s: validation refused: %v\n", call.Name, err)
		} else {
			pattern := ""
			if limits.Pattern != nil {
				pattern = limits.Pattern.String()
			}
			note = fmt.Sprintf("tool_limits consumed by %s: ai_callback_delay=%g ai_callback_pattern=%q max_output_bytes=%d\n", call.Name, limits.Delay.Seconds(), pattern, limits.MaxBytes)
		}
	}
	return limits, note, err
}
func (r *Registry) limitOverrides(raw json.RawMessage) (common.LimitOverrides, error) {
	var fields map[string]json.RawMessage
	out := common.LimitOverrides{}
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return out, r.failure("arguments must be an object")
	}
	for key, value := range fields {
		switch key {
		case "ai_callback_delay":
			var seconds float64
			if string(value) == "null" || json.Unmarshal(value, &seconds) != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds < 0 || seconds >= float64(math.MaxInt64)/float64(time.Second) {
				return out, r.failure("ai_callback_delay must be finite nonnegative seconds within duration range")
			}
			duration := time.Duration(seconds * float64(time.Second))
			out.Delay = &duration
		case "ai_callback_pattern":
			var pattern string
			if string(value) == "null" || json.Unmarshal(value, &pattern) != nil {
				return out, r.failure("ai_callback_pattern must be a Go regular expression string")
			}
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				return out, r.failure("ai_callback_pattern: invalid Go regular expression")
			}
			out.PatternSet = true
			if pattern != "" {
				out.Pattern = compiled
			}
		case "max_output_bytes":
			var budget int
			if json.Unmarshal(value, &budget) != nil || budget <= 0 {
				return out, r.failure("max_output_bytes must be a positive integer")
			}
			out.MaxBytes = &budget
		}
	}
	return out, nil
}
func (r *Registry) Supervise(call common.Part, limits common.Limits, note string) error {
	event, task := r.BeginSupervision(call, limits, note)
	if event != nil {
		return r.parent.RecordTool(common.Event{Type: "tool_returned", Tool: event})
	}
	if task.Ready != nil {
		<-task.Ready
	}
	if task.Input != nil {
		var err error
		task.Request.MatchStart, _, err = r.parent.Jobs().SendContext(context.Background(), task.Job, *task.Input)
		if err != nil {
			return err
		}
	}
	return r.parent.Jobs().Report(task.Job, task.Request)
}
func (r *Registry) BeginSupervision(call common.Part, limits common.Limits, note string) (*common.ToolEvent, *common.ReportTask) {
	finish := func(text string, err error) (*common.ToolEvent, *common.ReportTask) {
		if err != nil {
			text = call.Name + " failed: " + err.Error()
		}
		text = note + text
		return &common.ToolEvent{CallID: call.CallID, IsError: err != nil, Parts: []common.Part{{Type: "text", Text: &text}}}, nil
	}
	entry, ok := r.entries[call.Name]
	if !ok {
		return finish("", r.failure("tool unavailable"))
	}
	args, err := r.decode(call.Name, entry, call.Args)
	if err != nil {
		return finish("", err)
	}
	manager := r.parent.Jobs()
	if call.Name == "tool_limits" {
		supplied, err := r.limitOverrides(call.Args)
		if err != nil {
			return finish("", err)
		}
		if supplied.Delay == nil && supplied.MaxBytes == nil && !supplied.PatternSet {
			return finish("", r.failure("supply at least one limit field"))
		}
		if r.parent.Config().DataDir == "" {
			manager.SetLimits(supplied)
		}
		return finish("tool_limits set: supplied overrides apply to the next attempted call only, including invalid calls or another setter.", nil)
	}
	handle := args["handle"].(int)
	if handle <= 0 {
		return finish("", r.failure("handle must be a positive integer"))
	}
	job, err := manager.Lookup(uint64(handle))
	if err != nil {
		return finish("", err)
	}
	var ready <-chan struct{}
	var input *string
	request := common.JobReport{CallID: call.CallID, Limits: limits, Note: note, MatchStart: -1}
	switch call.Name {
	case "kill_job":
		ready, err = manager.RequestKill(job, "kill_job")
	case "send_input":
		text := args["input"].(string)
		if args["append_newline"].(bool) {
			text += "\n"
		}
		input = &text
	case "wait_for_job":
	default:
		return finish("", r.failure("not a supervision operation"))
	}
	if err != nil {
		return finish("", err)
	}
	return nil, &common.ReportTask{Job: job, Request: request, Ready: ready, Input: input}
}

// LimitCandidate decodes a setter without committing Jobs' pending setting.
func (r *Registry) LimitCandidate(call common.Part) (common.LimitValues, error) {
	value := common.LimitValues{}
	if call.Name != "tool_limits" {
		return value, r.failure("not a limit setter")
	}
	// Historical validation has an inert Registry with no installed handlers.
	// Decoding belongs here, while live availability is checked by dispatch.
	entry := entry{fields: r.limitFields()}
	if _, err := r.decode(call.Name, entry, call.Args); err != nil {
		return value, err
	}
	if _, err := r.limitOverrides(call.Args); err != nil {
		return value, err
	}
	if err := json.Unmarshal(call.Args, &value); err != nil {
		return value, r.failure("invalid typed limits")
	}
	if value.Delay == nil && value.Pattern == nil && value.MaxBytes == nil {
		return value, r.failure("supply at least one limit field")
	}
	return value, nil
}
