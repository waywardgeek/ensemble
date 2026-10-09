// Package agent owns conversation lifetime and each Agent's configuration.
package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ensemble/internal/common"
)

// The root retains this Agent; the Agent retains its children. Neither child
// imports the other implementation package. Their shared interfaces let the
// owner coordinate one synchronous turn without introducing a scheduler.
type agent struct {
	parent  common.Ensemble
	config  common.Config
	engine  common.Engine
	history common.History
	tools   []common.ToolDefinition
	jobs    common.Jobs
}

func New(parent common.Ensemble, cfg common.Config) *agent {
	if parent == nil {
		panic("Agent requires Ensemble")
	}
	return &agent{parent: parent, config: cfg}
}
func (a *agent) Ensemble() common.Ensemble { return a.parent }
func (a *agent) Config() common.Config     { return a.config }
func (a *agent) Engine() common.Engine     { return a.engine }
func (a *agent) History() common.History   { return a.history }

// The composition root installs this Agent's visible registry. Keeping one
// ordered slice makes declaration order stable and dispatch agree with visibility.
func (a *agent) SetTools(definitions []common.ToolDefinition) { a.tools = definitions }
func (a *agent) Declarations() []common.ToolDeclaration {
	var declarations []common.ToolDeclaration
	for _, tool := range a.tools {
		declarations = append(declarations, tool.ToolDeclaration)
	}
	return declarations
}

// A call has its actual Agent as parent; tools reach Engine and job services
// through that owner. Only the dispatch-created job is specific to this call.
type callContext struct {
	parent common.Agent
	job    common.Job
	limits common.Limits
}

func (c *callContext) Agent() common.Agent   { return c.parent }
func (c *callContext) Engine() common.Engine { return c.parent.Engine() }
func (c *callContext) Job() common.Job       { return c.job }
func (c *callContext) Limits() common.Limits { return c.limits }
func (a *agent) Jobs() common.Jobs           { return a.jobs }
func (a *agent) Shutdown() error             { return a.jobs.Shutdown() }
func (a *agent) AttachJobs(j common.Jobs) {
	if j == nil || j.Agent() != a {
		panic("Jobs must belong to this Agent")
	}
	a.jobs = j
}

// There is one dispatch funnel, including unknown tools and local file tools.
// Four explicit supervision definitions are the only calls without a new job.
func (a *agent) dispatch(call common.Part) error {
	var definition *common.ToolDefinition
	for i := range a.tools {
		if a.tools[i].Name == call.Name {
			definition = &a.tools[i]
			break
		}
	}
	// Consume a one-shot setting even if lookup or argument validation fails.
	// Leaving it pending after an error would make the next call surprising.
	limits, notice, limitErr := a.jobs.Consume(call.Args)
	context := &callContext{parent: a, limits: limits}
	event := &common.ToolData{CallID: call.CallID, Name: call.Name, Args: call.Args}
	if definition == nil || !definition.Supervision {
		job, err := a.jobs.Start()
		if err != nil {
			return err
		}
		context.job = job
		data := job.Data()
		event.Job = &data
	}
	// The running handle is a fact before the handler can produce side effects.
	if err := a.history.Append(common.Event{Type: "tool_called", Tool: event}); err != nil {
		return err
	}
	// Unknown calls and malformed arguments still need a result paired with
	// their ID; an unanswered call would poison the next vendor request.
	run := func() (string, error) {
		if limitErr != nil {
			return "", limitErr
		}
		if definition == nil {
			return "", fmt.Errorf("unknown tool %q", call.Name)
		}
		return definition.Run(context, call.Args)
	}
	var result string
	var toolErr error
	if context.job == nil {
		result, toolErr = run()
		if toolErr != nil {
			result = toolErr.Error()
		}
	} else {
		// The closure retains this call after dispatch wakes. A panic remains
		// an invariant failure, never a fabricated ordinary tool result.
		go func() {
			output, err := run()
			context.job.Finish(output, err)
		}()
		result, toolErr = context.job.Wait(limits)
	}
	if notice != "" {
		result = fmt.Sprintf("[tool_limits consumed by this %s call: %s]\n", call.Name, notice) + result
	}
	// Only Agent appends conversational events. Background readers update
	// the job; this returned fact captures what the dispatch observed.
	returned := &common.ToolData{CallID: call.CallID, Parts: []common.Part{{Type: "text", Text: result}}, IsError: toolErr != nil}
	if context.job != nil {
		data := context.job.Data()
		returned.Job = &data
	}
	return a.history.Append(common.Event{Type: "tool_returned", Tool: returned})
}

// The root constructs children; attachment verifies both sides of ownership
// before exposing the assembled Agent to clients.
func (a *agent) AttachEngine(e common.Engine) {
	if e == nil || e.Agent() != a {
		panic("Engine must belong to this Agent")
	}
	a.engine = e
}
func (a *agent) AttachHistory(h common.History) {
	if h == nil || h.Agent() != a {
		panic("History must belong to this Agent")
	}
	a.history = h
}

// Queueing is not a model request. Record the same arrival fact as any other
// input; the reducer classifies it and the next RequestSent consumes it.
// This keeps a pending instruction recoverable from the log alone.
func (a *agent) Ephemeral(value string) error {
	return a.history.Append(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.System, Parts: []common.Part{{Type: "text", Text: value}}}})
}

// Ask remains synchronous. Failed attempts now remain honest facts in the log,
// unlike Chapter 1's completed-pair slice; ErrorOccurred ends an in-flight turn.
func (a *agent) Ask(ctx context.Context, question string) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "", errors.New("user message is empty")
	}
	if err := a.history.Append(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: question}}}}); err != nil {
		return "", err
	}
	var answer strings.Builder
	// A tool call is the middle of a human turn. Record the complete reply
	// before running its calls, then return every result before requesting more.
	// Sixteen bounds model rounds. A tool wait returns without ending its job.
	for round := 0; round < 16; round++ {
		// Intermediate narration remains in the log; the caller receives the
		// final reply rather than a concatenation of every tool-planning sentence.
		answer.Reset()
		events, sendErr := a.engine.Send(ctx, a.history.Context())
		var calls []common.Part
		for _, event := range events {
			if err := a.history.Append(event); err != nil {
				return "", err
			}
			if event.Response != nil {
				for _, p := range event.Response.Parts {
					if p.Type == "text" {
						answer.WriteString(p.Text)
					}
					if p.Type == "tool_call" {
						calls = append(calls, p)
					}
				}
			}
		}
		if sendErr != nil || len(calls) == 0 {
			return answer.String(), sendErr
		}
		for _, call := range calls {
			if err := a.dispatch(call); err != nil {
				return "", err
			}
		}
	}
	return answer.String(), errors.New("tool loop reached 16 rounds")
}
