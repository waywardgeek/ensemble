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
		calls := 0
		for _, part := range response.Parts {
			if part.Type != "tool_call" {
				continue
			}
			calls++
			if err = owner.RecordTurn(common.Event{Type: "tool_called", Tool: &common.ToolEvent{CallID: part.CallID, Name: part.Name, Args: part.Args}}); err != nil {
				return common.ClientResult{}, err
			}
			result := owner.Registry().Execute(part)
			if err = owner.RecordTurn(common.Event{Type: "tool_returned", Tool: &result}); err != nil {
				return common.ClientResult{}, fmt.Errorf("tool %s completion could not be recorded; inspect its actual effect before any new attempt: %w", part.Name, err)
			}
		}
		if calls == 0 {
			parts, _ := Clone(e, response.Parts)
			return common.ClientResult{Text: TextAnswer(e, parts), Parts: parts, Usage: e.Usage()}, nil
		}
	}
	return fail("round_limit", failure(e, "round_limit: sixteen model requests completed; final tool batch retained"))
}
