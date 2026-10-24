package tool

import (
	"context"
	"encoding/json"
	"ensemble/internal/common"
)

func (r *registry) statusTool() entry {
	return entry{definition: common.ToolDefinition{Name: "agent_status", Description: "Read selected and last-producing models, measured response/session tokens, cache hit rate and session cost. Cost status distinguishes known, partial and unknown pricing; known charges remain visible.", Parameters: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)}, handler: func(ctx context.Context, call common.ToolCall, raw json.RawMessage) error {
		args, err := decode(raw)
		if err != nil {
			return err
		}
		if err = args.only(); err != nil {
			return err
		}
		return json.NewEncoder(call).Encode(call.Parent().Engine().Status())
	}}
}
