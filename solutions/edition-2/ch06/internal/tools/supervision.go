package tools

import (
	"encoding/json"
	"ensemble/internal/common"
)

// Keep the three shared wait arguments identical on each waiting declaration.
// These are schema fragments, not a second settings or dispatch framework.
func waitSchema(properties, required string) json.RawMessage {
	if properties != "" {
		properties += ","
	}
	return json.RawMessage(`{"type":"object","properties":{` + properties + `"ai_callback_delay":{"type":"number","description":"Seconds before looking again, default 3; never a job deadline"},"ai_callback_pattern":{"type":"string","description":"Regex on unseen output"},"max_output_bytes":{"type":"integer","description":"Inline output byte cap, default 16384"}},"required":[` + required + `]}`)
}
func supervision() []common.ToolDefinition {
	return []common.ToolDefinition{
		{ToolDeclaration: common.ToolDeclaration{Name: "wait_for_job", Description: "Wait for a job, including a finished job. Return only unseen output; waiting does not stop execution.", Schema: waitSchema(`"handle":{"type":"integer"}`, `"handle"`)}, Supervision: true, Run: waitJob},
		{ToolDeclaration: common.ToolDeclaration{Name: "send_input", Description: "Send a line of input to a running PTY job, then wait. A missing trailing newline is added.", Schema: waitSchema(`"handle":{"type":"integer"},"input":{"type":"string"}`, `"handle","input"`)}, Supervision: true, Run: sendInput},
		{ToolDeclaration: common.ToolDeclaration{Name: "kill_job", Description: "Kill a job's process group. Go-only work is marked killed but its goroutine cannot be stopped.", Schema: json.RawMessage(`{"type":"object","properties":{"handle":{"type":"integer"}},"required":["handle"]}`)}, Supervision: true, Run: killJob},
		{ToolDeclaration: common.ToolDeclaration{Name: "tool_limits", Description: "Set one-shot wait/output limits for the very next call, whichever tool it is. That call reports consumption; explicit call arguments override these limits.", Schema: waitSchema("", "")}, Supervision: true, Run: toolLimits},
	}
}
func findJob(owner common.ToolContext, raw json.RawMessage) (common.Job, error) {
	var args struct {
		// Handle identifies an existing job within the owning Agent.
		Handle int `json:"handle"`
	}
	if err := decode(raw, &args, "handle"); err != nil {
		return nil, err
	}
	return owner.Engine().Agent().Jobs().Find(args.Handle)
}
func waitJob(owner common.ToolContext, raw json.RawMessage) (string, error) {
	job, err := findJob(owner, raw)
	if err != nil {
		return "", err
	}
	return job.Wait(owner.Limits())
}
func sendInput(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// Input is the line sent to the selected interactive job.
		Input string `json:"input"`
	}
	if err := decode(raw, &args, "input"); err != nil {
		return "", err
	}
	job, err := findJob(owner, raw)
	if err != nil {
		return "", err
	}
	if err := job.Send(args.Input); err != nil {
		return "", err
	}
	return job.Wait(owner.Limits())
}
func killJob(owner common.ToolContext, raw json.RawMessage) (string, error) {
	job, err := findJob(owner, raw)
	if err != nil {
		return "", err
	}
	if err := job.Kill("kill_job"); err != nil {
		return "", err
	}
	return job.Wait(common.Limits{MaxBytes: 16384})
}
func toolLimits(owner common.ToolContext, raw json.RawMessage) (string, error) {
	if err := owner.Engine().Agent().Jobs().SetLimits(raw); err != nil {
		return "", err
	}
	return "Limits set for the next call, whichever tool that is; its result will report consumption.", nil
}
