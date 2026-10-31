package tools

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"ensemble/internal/common"
)

// Management calls are ordinary jobs, so human steering can release their wait
// without killing the operation. All target lookup stays inside the caller's fleet.
func managementTools() []common.ToolDefinition {
	id := `"agent_id":{"type":"string"}`
	content := id + `,"content":{"type":"string"},"cursor":{"type":"integer"}`
	return []common.ToolDefinition{
		{ToolDeclaration: common.ToolDeclaration{Name: "wait_for_agent_change", Description: "Wait for owned children to change, exit, ask a question, match output pattern, or stop progressing. Empty agent_ids means all children.", Schema: waitSchema(`"agent_ids":{"type":"array","items":{"type":"string"}},"pattern":{"type":"string"},"timeout_seconds":{"type":"number"},"stuck_seconds":{"type":"number"}`, "")}, Run: waitAgentChange},
		{ToolDeclaration: common.ToolDeclaration{Name: "send_message", Description: "Send live steering to an owned child (hint when busy, new turn when idle). Return finalized output since cursor.", Schema: waitSchema(content, `"agent_id","content"`)}, Run: sendAgentMessage},
		{ToolDeclaration: common.ToolDeclaration{Name: "wait_for_agent", Description: "Wait for new finalized child output after cursor, or return when idle. Cursor is local to this child.", Schema: waitSchema(id+`,"cursor":{"type":"integer"}`, `"agent_id"`)}, Run: waitAgentOutput},
		{ToolDeclaration: common.ToolDeclaration{Name: "ask_agent", Description: "Queue a separate question to an owned child and wait for that turn's result.", Schema: waitSchema(content, `"agent_id","content"`)}, Run: askAgent},
		{ToolDeclaration: common.ToolDeclaration{Name: "get_submission", Description: "Read an owned child's submitted result, with durable submit.md fallback.", Schema: waitSchema(id, `"agent_id"`)}, Run: getSubmission},
		{ToolDeclaration: common.ToolDeclaration{Name: "kill_agent", Description: "Cancel and remove an owned child immediately, including its jobs and descendants; retain saved results.", Schema: waitSchema(id, `"agent_id"`)}, Run: killAgent},
		{ToolDeclaration: common.ToolDeclaration{Name: "get_sub_agent_status", Description: "Report an owned child's lifecycle, activity and measured session token usage.", Schema: waitSchema(id, `"agent_id"`)}, Run: getAgentStatus},
		{ToolDeclaration: common.ToolDeclaration{Name: "interrupt_agent", Description: "End an owned child's turn while keeping the Agent and other jobs alive. mode now cancels the request; after_tool waits for the next tool report.", Schema: waitSchema(id+`,"mode":{"type":"string","enum":["now","after_tool"]}`, `"agent_id"`)}, Run: interruptAgent},
	}
}

type managementArgs struct {
	// ID resolves within the calling Agent's children.
	ID string `json:"agent_id"`
	// Content is the child's new prompt or steering message.
	Content string `json:"content"`
	// Cursor excludes already-read history events.
	Cursor uint64 `json:"cursor"`
	// Mode chooses immediate or after-tool interruption.
	Mode string `json:"mode"`
}

func target(owner common.ToolContext, raw json.RawMessage) (common.Agent, managementArgs, error) {
	var args managementArgs
	if err := decode(raw, &args, "agent_id"); err != nil {
		return nil, args, err
	}
	a, err := owner.Agent().Fleet().Find(args.ID)
	return a, args, err
}
func sendAgentMessage(owner common.ToolContext, raw json.RawMessage) (string, error) {
	a, args, err := target(owner, raw)
	if err != nil {
		return "", err
	}
	if err = a.Post(common.UserMessage{Text: args.Content}); err != nil {
		return "", err
	}
	output, err := owner.Agent().Fleet().Output(owner.Job().Context(), args.ID, args.Cursor, false)
	if err != nil {
		return "", err
	}
	return fleetJSON(output)
}
func waitAgentOutput(owner common.ToolContext, raw json.RawMessage) (string, error) {
	_, args, err := target(owner, raw)
	if err != nil {
		return "", err
	}
	output, err := owner.Agent().Fleet().Output(owner.Job().Context(), args.ID, args.Cursor, true)
	if err != nil {
		return "", err
	}
	return fleetJSON(output)
}
func askAgent(owner common.ToolContext, raw json.RawMessage) (string, error) {
	a, args, err := target(owner, raw)
	if err != nil {
		return "", err
	}
	return a.Ask(owner.Job().Context(), args.Content)
}
func getSubmission(owner common.ToolContext, raw json.RawMessage) (string, error) {
	_, args, err := target(owner, raw)
	if err != nil {
		return "", err
	}
	result, err := owner.Agent().Fleet().SubmissionData(args.ID)
	if err != nil {
		return "", err
	}
	if len(result.Payload) != 0 {
		return fleetJSON(map[string]any{"agent_id": args.ID, "payload": result.Payload})
	}
	return fleetJSON(map[string]string{"agent_id": args.ID, "result": result.Result})
}
func killAgent(owner common.ToolContext, raw json.RawMessage) (string, error) {
	_, args, err := target(owner, raw)
	if err != nil {
		return "", err
	}
	if err = owner.Agent().Fleet().ShutdownChild(args.ID); err != nil {
		return "", err
	}
	return "agent killed; saved work retained", nil
}
func getAgentStatus(owner common.ToolContext, raw json.RawMessage) (string, error) {
	a, _, err := target(owner, raw)
	if err != nil {
		return "", err
	}
	return fleetJSON(a.Fleet().Status())
}
func interruptAgent(owner common.ToolContext, raw json.RawMessage) (string, error) {
	a, args, err := target(owner, raw)
	if err != nil {
		return "", err
	}
	if args.Mode != "" && args.Mode != "now" && args.Mode != "after_tool" {
		return "", errors.New("mode must be now or after_tool")
	}
	if err = a.Post(common.Interrupt{AfterTool: args.Mode == "after_tool"}); err != nil {
		return "", err
	}
	return "interrupt delivered; agent remains available", nil
}
func waitAgentChange(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// IDs selects children of this caller only.
		IDs []string `json:"agent_ids"`
		// Pattern optionally matches finalized output.
		Pattern string `json:"pattern"`
		// Timeout bounds the observation job independently from its first wait.
		Timeout float64 `json:"timeout_seconds"`
		// Stuck reports a busy child after this many seconds without progress.
		Stuck float64 `json:"stuck_seconds"`
	}
	if err := decode(raw, &args); err != nil {
		return "", err
	}
	if args.Timeout < 0 || args.Stuck < 0 {
		return "", errors.New("timeouts must be nonnegative")
	}
	ctx := owner.Job().Context()
	if args.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(args.Timeout*float64(time.Second)))
		defer cancel()
	}
	return owner.Agent().Fleet().WaitChange(ctx, args.IDs, args.Pattern, time.Duration(args.Stuck*float64(time.Second)))
}
