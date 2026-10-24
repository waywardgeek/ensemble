package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"ensemble/internal/common"
)

const fleetSource = "builtin:fleet"

func (r *registry) fleetTools() []entry {
	fleet, ok := r.parent.Parent().(common.Fleet)
	if !ok {
		return nil
	}
	id := r.parent.ID()
	encode := func(v any, err error) (string, error) {
		if err != nil {
			return "", err
		}
		raw, err := json.Marshal(v)
		return string(raw), err
	}
	owned := func(a arguments) (string, error) {
		target, err := a.text("agent_id", true)
		if err != nil {
			return "", err
		}
		status, err := fleet.Status(target)
		if err != nil {
			return "", err
		}
		if status.ParentID != id {
			return "", fmt.Errorf("%q is not your direct child", target)
		}
		return target, nil
	}
	makeEntry := func(name, description, schema string, fn func(context.Context, arguments) (string, error)) entry {
		return entry{source: fleetSource, definition: common.ToolDefinition{Name: name, Description: description, Parameters: json.RawMessage(schema)}, execute: fn}
	}
	targetSchema := `{"type":"object","properties":{"agent_id":{"type":"string"}},"required":["agent_id"],"additionalProperties":false}`
	entries := []entry{
		makeEntry("spawn_sub_agent", "Start an owned delegated Agent with independent history and jobs, an application-configured state directory, and a terminal submit contract. IDs are unique and retained after completion.", `{"type":"object","properties":{"agent_id":{"type":"string"},"task":{"type":"string"},"model":{"type":"string"},"reasoning_effort":{"type":"string"},"submit_schema":{"type":"object"}},"required":["agent_id","task"],"additionalProperties":false}`, func(ctx context.Context, a arguments) (string, error) {
			if err := a.only("agent_id", "task", "model", "reasoning_effort", "submit_schema"); err != nil {
				return "", err
			}
			childID, err := a.text("agent_id", true)
			if err != nil {
				return "", err
			}
			prompt, err := a.text("task", true)
			if err != nil {
				return "", err
			}
			model, err := a.text("model", false)
			if err != nil {
				return "", err
			}
			reasoning, err := a.text("reasoning_effort", false)
			if err != nil {
				return "", err
			}
			child, err := fleet.Spawn(ctx, id, common.Config{ID: childID, Prompt: prompt, Model: model, ReasoningEffort: reasoning, SubmitSchema: a["submit_schema"]})
			if err != nil {
				return "", err
			}
			status, err := fleet.Status(child.ID())
			return encode(status, err)
		}),
		makeEntry("agents_status", "List managed direct children and their current states.", `{"type":"object","properties":{},"additionalProperties":false}`, func(_ context.Context, a arguments) (string, error) {
			if err := a.only(); err != nil {
				return "", err
			}
			return encode(fleet.Children(id), nil)
		}),
		makeEntry("shutdown_agent", "Recursively stop an owned child, retaining its final outcome.", targetSchema, func(_ context.Context, a arguments) (string, error) {
			if err := a.only("agent_id"); err != nil {
				return "", err
			}
			target, err := owned(a)
			if err != nil {
				return "", err
			}
			if err := fleet.ShutdownAgent(target); err != nil {
				return "", err
			}
			outcome, err := fleet.Outcome(target)
			return encode(outcome, err)
		}),

		makeEntry("join_agents", "Join selected direct children, or all current children when agent_ids is omitted or empty. Failures are returned as terminal outcomes.", `{"type":"object","properties":{"agent_ids":{"type":"array","items":{"type":"string"}}},"additionalProperties":false}`, func(ctx context.Context, a arguments) (string, error) {
			if err := a.only("agent_ids"); err != nil {
				return "", err
			}
			var ids []string
			if raw, ok := a["agent_ids"]; ok {
				if err := json.Unmarshal(raw, &ids); err != nil {
					return "", err
				}
			}
			outcomes, err := fleet.Join(ctx, id, ids)
			return encode(outcomes, err)
		}),
		makeEntry("respond_to_parent_message", "Answer a question addressed to this Agent. Omit message_id only when that child has one pending question.", `{"type":"object","properties":{"agent_id":{"type":"string"},"message_id":{"type":"string"},"response":{"type":"string"}},"required":["agent_id","response"],"additionalProperties":false}`, func(_ context.Context, a arguments) (string, error) {
			if err := a.only("agent_id", "message_id", "response"); err != nil {
				return "", err
			}
			target, err := a.text("agent_id", true)
			if err != nil {
				return "", err
			}
			request, err := a.text("message_id", false)
			if err != nil {
				return "", err
			}
			response, err := a.text("response", true)
			if err != nil {
				return "", err
			}
			return "Answer delivered.", fleet.RespondToQuestion(id, target, request, response)
		}),
	}
	{
		entries = append(entries, makeEntry("send_message_to_parent", "Ask the addressed recipient, defaulting to the parent or configured human address. The correlated reply includes its sender for follow-up. This does not submit.", `{"type":"object","properties":{"content":{"type":"string"},"human":{"type":"boolean"},"recipient":{"type":"string"}},"required":["content"],"additionalProperties":false}`, func(ctx context.Context, a arguments) (string, error) {
			if err := a.only("content", "human", "recipient"); err != nil {
				return "", err
			}
			text, err := a.text("content", true)
			if err != nil {
				return "", err
			}
			var human bool
			if raw, ok := a["human"]; ok {
				if err := json.Unmarshal(raw, &human); err != nil {
					return "", err
				}
			}
			recipient, err := a.text("recipient", false)
			if err != nil {
				return "", err
			}
			if human {
				if recipient != "" && recipient != fleet.HumanAddress() {
					return "", fmt.Errorf("human and recipient disagree")
				}
				recipient = fleet.HumanAddress()
			}
			reply, err := fleet.SendMessage(ctx, id, recipient, text)
			return encode(reply, err)
		}))
		schema := r.parent.Config().SubmitSchema
		if len(schema) == 0 {
			schema = common.DefaultSubmitSchema
		}
		entries = append(entries, makeEntry("submit", "Finish this conversation with a schema-valid result. Join or shut down unfinished children first. Accepted submit ends the task without another provider request.", string(schema), func(context.Context, arguments) (string, error) {
			return "", fmt.Errorf("submit must use its actor control path")
		}))
	}
	return entries
}
