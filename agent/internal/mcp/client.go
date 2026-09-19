package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolInfo carries enough to build a common.ToolDecl from a discovered
// MCP tool. The Ephemeral field ("round", "turn", or "") determines
// whether the engine auto-calls this tool.
type ToolInfo struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
	Ephemeral   string          `json:"ephemeral,omitempty"`
}

// ToolResult is the content returned by an MCP tools/call response.
type ToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentItem is one piece of content in a tool result.
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Client drives the MCP protocol: handshake, tool discovery, tool calls,
// and reverse tool handling.
type Client struct {
	codec    *Codec
	tools    []ToolInfo
	reverse  func(name string, args json.RawMessage) (string, error)
}

// NewClient creates an MCP client on the given transport.
// Call Initialize before any other method.
func NewClient(t Transport) *Client {
	c := &Client{}
	c.codec = NewCodec(t)
	c.codec.onRequest = c.handleIncoming
	c.codec.Run()
	return c
}

// Initialize performs the MCP handshake: sends initialize, waits for the
// server's response, then sends notifications/initialized.
func (c *Client) Initialize(ctx context.Context) error {
	params := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]string{
			"name":    "ensemble",
			"version": "0.1.0",
		},
	}
	_, err := c.codec.Call("initialize", params)
	if err != nil {
		return fmt.Errorf("mcp initialize: %w", err)
	}
	return c.codec.Notify("notifications/initialized", nil)
}

// ListTools sends tools/list and returns the discovered tools.
func (c *Client) ListTools(ctx context.Context) ([]ToolInfo, error) {
	raw, err := c.codec.Call("tools/list", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("mcp tools/list: %w", err)
	}

	var result struct {
		Tools []ToolInfo `json:"tools"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("mcp tools/list unmarshal: %w", err)
	}
	c.tools = result.Tools
	return result.Tools, nil
}

// CallTool sends tools/call and waits for the result. If the context is
// cancelled (e.g., the job is killed), the client sends $/cancelRequest
// and returns the context error.
func (c *Client) CallTool(ctx context.Context, name string, args json.RawMessage) (*ToolResult, error) {
	params := map[string]any{
		"name": name,
	}
	if len(args) > 0 {
		var parsed any
		if json.Unmarshal(args, &parsed) == nil {
			params["arguments"] = parsed
		}
	}

	// Use a goroutine so we can respect context cancellation.
	type result struct {
		raw json.RawMessage
		err error
	}
	ch := make(chan result, 1)
	go func() {
		raw, err := c.codec.Call("tools/call", params)
		ch <- result{raw, err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			return nil, r.err
		}
		var tr ToolResult
		if err := json.Unmarshal(r.raw, &tr); err != nil {
			return nil, fmt.Errorf("mcp tools/call unmarshal: %w", err)
		}
		return &tr, nil
	case <-ctx.Done():
		// Context cancelled — tell the server to cancel.
		_ = c.codec.Notify("$/cancelRequest", nil)
		return nil, ctx.Err()
	}
}

// SetReverseHandler sets the handler for incoming tool calls from the MCP
// server (bidirectional mode). When the server sends a tools/call request:
// 1. Look up the tool in the agent's registry
// 2. Execute it
// 3. Send the result back as a JSON-RPC response
func (c *Client) SetReverseHandler(handler func(name string, args json.RawMessage) (string, error)) {
	c.reverse = handler
}

// handleIncoming processes incoming requests from the MCP server.
func (c *Client) handleIncoming(req Request) {
	switch req.Method {
	case "tools/call":
		c.handleReverseCall(req)
	default:
		// Unknown method — send error response if it has an ID.
		if req.ID != nil {
			_ = c.codec.Respond(req.ID, nil, &RPCError{
				Code:    -32601,
				Message: fmt.Sprintf("method not found: %s", req.Method),
			})
		}
	}
}

// handleReverseCall handles an incoming tools/call from the MCP server.
func (c *Client) handleReverseCall(req Request) {
	if c.reverse == nil {
		if req.ID != nil {
			_ = c.codec.Respond(req.ID, nil, &RPCError{
				Code:    -32601,
				Message: "no reverse handler registered",
			})
		}
		return
	}

	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		if req.ID != nil {
			_ = c.codec.Respond(req.ID, nil, &RPCError{
				Code:    -32602,
				Message: fmt.Sprintf("invalid params: %v", err),
			})
		}
		return
	}

	result, err := c.reverse(params.Name, params.Arguments)
	if err != nil {
		if req.ID != nil {
			_ = c.codec.Respond(req.ID, ToolResult{
				Content: []ContentItem{{Type: "text", Text: err.Error()}},
				IsError: true,
			}, nil)
		}
		return
	}

	if req.ID != nil {
		_ = c.codec.Respond(req.ID, ToolResult{
			Content: []ContentItem{{Type: "text", Text: result}},
		}, nil)
	}
}

// Close shuts down the MCP client.
func (c *Client) Close() error {
	return c.codec.Close()
}
