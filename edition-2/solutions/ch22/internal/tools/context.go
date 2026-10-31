package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"ensemble/internal/common"
)

// These handlers acknowledge ordinary calls. The actor records the handoff
// after the result, and the reducer delays clearing until the whole batch ends.
func contextTools() []common.ToolDefinition {
	return []common.ToolDefinition{
		{ToolDeclaration: common.ToolDeclaration{Name: "micro_handoff", Description: "Checkpoint at a completed micro-goal. Write text with what succeeded, what failed, source locations and next steps. Removes all tool traffic and opaque reasoning from context but preserves conversation and this note.", Schema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}`)}, Supervision: true, Run: handoff},
		{ToolDeclaration: common.ToolDeclaration{Name: "keep_tool_results", Description: "Keep the previous batch of tool results instead of automatic stubbing on supported models. Does not keep results of tools called alongside this call.", Schema: json.RawMessage(`{"type":"object","properties":{}}`)}, Supervision: true, Run: keepResults},
	}
}
func handoff(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args common.MicroHandoffData
	if err := decode(raw, &args, "text"); err != nil {
		return "", err
	}
	if strings.TrimSpace(args.Text) == "" {
		return "", fmt.Errorf("handoff text must not be empty")
	}
	return "Checkpoint recorded after this tool batch.", nil
}
func keepResults(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct{}
	if err := decode(raw, &args); err != nil {
		return "", err
	}
	return "Previous tool results kept.", nil
}
