package tool

import (
	"context"
	"encoding/json"
	"ensemble/internal/common"
	"io"
)

func (r *registry) contextTools() []entry {
	var out []entry
	for _, name := range []string{"compact_context", "micro_handoff", "keep_tool_results"} {
		name := name
		schema := json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)
		description := "Compact older tool results in this Agent's context without affecting durable history or other Agents."
		if name == "micro_handoff" {
			schema = json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`)
			description = "Record a concise checkpoint of goal, decisions, changes, tests, blockers, live handles, and next steps; compact after this tool batch."
		} else if name == "keep_tool_results" {
			description = "Protect this Agent's current tool results for one following tool batch, delaying automatic stubbing."
		}
		out = append(out, entry{definition: common.ToolDefinition{Name: name, Description: description, Parameters: schema}, handler: func(ctx context.Context, owner common.ToolCall, raw json.RawMessage) error {
			args, err := decode(raw)
			if err != nil {
				return err
			}
			text := ""
			if name == "micro_handoff" {
				if err = args.only("text"); err != nil {
					return err
				}
				text, err = args.text("text", true)
			} else {
				err = args.only()
			}
			if err != nil {
				return err
			}
			result, err := r.parent.CompactContext(ctx, name, text)
			if err != nil {
				return err
			}
			_, err = io.WriteString(owner, result)
			return err
		}})
	}
	return out
}
