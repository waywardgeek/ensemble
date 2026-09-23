package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Bridge converts MCP-discovered tools into the agent's tool registry.
//
// MCP tools are processes, not functions. Each bridged tool's handler is a
// normal tool handler — takes context.Context via the Call, returns result.
// The framework's existing job infrastructure wraps every call, giving it a
// handle, cancellation, and wait/kill.
//
// The handler calls client.CallTool. If the job is killed (context cancelled),
// the client stops waiting and sends $/cancelRequest to the MCP server.
//
// Non-ephemeral MCP tools are registered as normal tools. Ephemeral MCP tools
// are registered with the Ephemeral field set.
func Bridge(client *Client, tools []ToolInfo) []common.Tool {
	var result []common.Tool
	for _, ti := range tools {
		ti := ti // capture
		t := common.Tool{
			Name:        ti.Name,
			Description: ti.Description,
			Schema:      ti.InputSchema,
			Ephemeral:   ti.Ephemeral,
			Run: func(c *common.Call, args json.RawMessage) (string, error) {
				// MCP tool calls use context.Background because jobs do not
				// carry a Context. The job infrastructure handles the kill case
				// by discarding the result — same as for any other tool that
				// outlives its dispatcher.
				tr, err := client.CallTool(context.Background(), ti.Name, args)
				if err != nil {
					return "", fmt.Errorf("mcp tool %s: %w", ti.Name, err)
				}
				if tr.IsError {
					return extractText(tr), fmt.Errorf("mcp tool %s returned error: %s", ti.Name, extractText(tr))
				}
				return extractText(tr), nil
			},
		}
		result = append(result, t)
	}
	return result
}

// extractText concatenates text content items from a ToolResult.
func extractText(tr *ToolResult) string {
	var parts []string
	for _, item := range tr.Content {
		if item.Type == "text" {
			parts = append(parts, item.Text)
		}
	}
	return strings.Join(parts, "\n")
}
