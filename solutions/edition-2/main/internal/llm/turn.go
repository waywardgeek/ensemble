package llm

import (
	"context"
	"example.com/ensemble/internal/common"
	"fmt"
)

// Turn runs under its Agent's admission lock. Each accepted batch is durable
// before continuation; neither provider failure nor the round bound rolls it back.
func (e *Engine) Turn(ctx context.Context) (common.ClientResult, error) {
	owner, ok := e.parent.(common.TurnAgent)
	if !ok {
		return common.ClientResult{}, failure(e, "Agent does not provide turn services")
	}
	fail := func(code string, err error) (common.ClientResult, error) {
		if writeErr := owner.RecordTurn(common.Event{Type: "error_occurred", Error: &common.EventError{Code: code, Message: err.Error()}}); writeErr != nil {
			return common.ClientResult{}, writeErr
		}
		return common.ClientResult{}, err
	}
	for round := 1; round <= 16; round++ {
		config := owner.Config()
		config.Tools = owner.Registry().Declarations()
		state := owner.TurnSnapshot()
		// Render before request_sent consumes ephemera; a render failure did not send them.
		body, err := Render(e, state, config)
		if err != nil {
			return fail("request_failed", err)
		}
		seqs := make([]uint64, 0, len(state.Ephemera))
		for _, entry := range state.Ephemera {
			seqs = append(seqs, entry.Seq)
		}
		if err = owner.RecordTurn(common.Event{Type: "request_sent", Request: &common.RequestEvent{To: Route(e, config), Ephemera: seqs}}); err != nil {
			return common.ClientResult{}, err
		}
		response, err := e.Exchange(ctx, body, owner.NextSequence())
		if err != nil {
			return fail("request_failed", err)
		}
		if err = owner.RecordTurn(common.Event{Type: "response_ended", Response: &response}); err != nil {
			return common.ClientResult{}, err
		}
		committed := owner.TurnSnapshot()
		response.Parts = committed.Entries[len(committed.Entries)-1].Parts
		calls := 0
		for _, part := range response.Parts {
			if part.Type != "tool_call" {
				continue
			}
			calls++
			call := &Call{parent: e, part: part}
			if err = call.dispatch(); err != nil {
				return common.ClientResult{}, err
			}
		}
		if calls == 0 {
			parts, _ := Clone(e, response.Parts)
			return common.ClientResult{Text: TextAnswer(e, parts), Parts: parts, Usage: e.Usage()}, nil
		}
	}
	return fail("round_limit", failure(e, "round_limit: sixteen model requests completed; final tool batch retained"))
}

// A call has one parent, Engine; registry and job services remain Agent-owned.
type Call struct {
	parent common.Engine
	part   common.Part
}

func (c *Call) Engine() common.Engine { return c.parent }
func (c *Call) dispatch() error {
	owner := c.Engine().Agent().(common.TurnAgent)
	manager := owner.Jobs()
	limits, note, limitErr := owner.Registry().ResolveLimits(c.part)
	available, supervision := owner.Registry().Kind(c.part.Name)
	var job common.Job
	var err error
	if available && !supervision && limitErr == nil {
		job, err = manager.Create()
		if err != nil {
			return err
		}
	}
	called := common.ToolEvent{CallID: c.part.CallID, Name: c.part.Name, Args: c.part.Args}
	if job != nil {
		snapshot := job.Snapshot()
		called.Job = &snapshot
	}
	if err = owner.RecordTurn(common.Event{Type: "tool_called", Tool: &called}); err != nil {
		if job != nil {
			manager.Abort(job)
		}
		return err
	}
	if limitErr != nil || !available {
		text := "tool is unavailable to this Agent"
		if limitErr != nil {
			text = limitErr.Error()
		}
		text = note + c.part.Name + " failed: " + text
		return owner.RecordTurn(common.Event{Type: "tool_returned", Tool: &common.ToolEvent{CallID: c.part.CallID, IsError: true, Parts: []common.Part{Text(text)}}})
	}
	if supervision {
		return owner.Registry().Supervise(c.part, limits, note)
	}
	manager.Start(job, c.part)
	if err := manager.Report(job, common.JobReport{CallID: c.part.CallID, Limits: limits, Note: note, Original: true, MatchStart: -1}); err != nil {
		return fmt.Errorf("tool %s completion could not be recorded; inspect its actual effect before any new attempt: %w", c.part.Name, err)
	}
	return nil
}
