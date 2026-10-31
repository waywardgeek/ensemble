package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"ensemble/internal/common"
)

// The client is also the correlation codec. A separate exported codec would add
// an ownership layer without another user; one reader routes every response.
type client struct {
	parent    common.MCP
	transport common.MCPTransport
	mu        sync.Mutex
	next      int
	pending   map[int]chan message
	done      chan struct{}
	once      sync.Once
	tools     []toolInfo
}

type message struct {
	// JSONRPC fixes the envelope version independently of the MCP version.
	JSONRPC string `json:"jsonrpc"`
	// ID is retained as JSON so reverse replies preserve the peer's identity.
	ID json.RawMessage `json:"id,omitempty"`
	// Method is present only on requests and notifications.
	Method string `json:"method,omitempty"`
	// Params preserves method arguments without interpreting tool schemas.
	Params json.RawMessage `json:"params,omitempty"`
	// Result carries a correlated success payload.
	Result json.RawMessage `json:"result,omitempty"`
	// Error carries a protocol failure instead of a result.
	Error *rpcError `json:"error,omitempty"`
}
type rpcError struct {
	// Code is the peer's JSON-RPC error category.
	Code int `json:"code"`
	// Message explains the protocol failure.
	Message string `json:"message"`
}
type toolInfo struct {
	// Name is the advertised dispatch identity, preserved without normalization.
	Name string `json:"name"`
	// Description is the server's model-facing tool explanation.
	Description string `json:"description"`
	// Schema is the server's vendor-neutral inputSchema.
	Schema json.RawMessage `json:"inputSchema"`
	// Ephemeral marks automatic tools that must never be declared to the model.
	Ephemeral string `json:"ephemeral,omitempty"`
}
type content struct {
	// Type distinguishes text blocks from other MCP content.
	Type string `json:"type"`
	// Text carries the text content supported by this chapter.
	Text string `json:"text"`
}
type toolResult struct {
	// Content preserves the server's ordered result blocks.
	Content []content `json:"content"`
	// IsError distinguishes tool execution failure from protocol failure.
	IsError bool `json:"isError,omitempty"`
}

func newClient(parent common.MCP, transport common.MCPTransport) *client {
	if parent == nil || transport == nil {
		panic("MCP client requires its manager and transport")
	}
	c := &client{parent: parent, transport: transport, pending: make(map[int]chan message), done: make(chan struct{})}
	go c.read()
	return c
}

// MCP exposes the actual owner to helpers; no tool or logger is copied here.
func (c *client) MCP() common.MCP { return c.parent }
func (c *client) close() error {
	var err error
	c.once.Do(func() { close(c.done); err = c.transport.Close() })
	return err
}
func (c *client) send(m message) error {
	m.JSONRPC = "2.0"
	raw, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return c.transport.Send(raw)
}
func raw(value any) json.RawMessage { data, _ := json.Marshal(value); return data }

func (c *client) request(ctx context.Context, method string, params any, result any) error {
	c.mu.Lock()
	c.next++
	id := c.next
	reply := make(chan message, 1)
	c.pending[id] = reply
	c.mu.Unlock()
	defer func() { c.mu.Lock(); delete(c.pending, id); c.mu.Unlock() }()
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := c.send(message{ID: raw(id), Method: method, Params: raw(params)}); err != nil {
		return err
	}
	select {
	case m := <-reply:
		if m.Error != nil {
			return fmt.Errorf("MCP %s (%d): %s", method, m.Error.Code, m.Error.Message)
		}
		return json.Unmarshal(m.Result, result)
	case <-ctx.Done():
		// Cancellation follows job lifetime, not the model's callback deadline.
		// The historical exercise explicitly uses this cancellation notification.
		go c.send(message{Method: "$/cancelRequest", Params: raw(map[string]any{"id": id})})
		return ctx.Err()
	case <-c.done:
		return errClosed
	}
}

func (c *client) initialize(ctx context.Context) error {
	var result struct {
		// ProtocolVersion must match the historical exercise peer we implement.
		ProtocolVersion string `json:"protocolVersion"`
	}
	params := map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{}, "clientInfo": map[string]string{"name": "Ensemble", "version": "1"}}
	if err := c.request(ctx, "initialize", params, &result); err != nil {
		return err
	}
	if result.ProtocolVersion != "2024-11-05" {
		return fmt.Errorf("unsupported MCP protocol version %q", result.ProtocolVersion)
	}
	if err := c.send(message{Method: "notifications/initialized"}); err != nil {
		return err
	}
	var listed struct {
		// Tools is the discovery result, still subject to the owning skill's grant.
		Tools []toolInfo `json:"tools"`
	}
	if err := c.request(ctx, "tools/list", map[string]any{}, &listed); err != nil {
		return err
	}
	for _, info := range listed.Tools {
		if info.Name == "" || !json.Valid(info.Schema) {
			return fmt.Errorf("MCP tool requires name and inputSchema")
		}
		if info.Ephemeral != "" && info.Ephemeral != "round" && info.Ephemeral != "turn" {
			return fmt.Errorf("invalid MCP ephemeral mode %q", info.Ephemeral)
		}
	}
	c.tools = listed.Tools
	return nil
}

func (c *client) call(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if len(args) == 0 {
		args = raw(map[string]any{})
	}
	var result toolResult
	if err := c.request(ctx, "tools/call", map[string]any{"name": name, "arguments": args}, &result); err != nil {
		return "", err
	}
	var text []string
	for _, part := range result.Content {
		if part.Type == "text" {
			text = append(text, part.Text)
		}
	}
	joined := strings.Join(text, "\n")
	if result.IsError {
		return "", fmt.Errorf("MCP tool %s: %s", name, joined)
	}
	return joined, nil
}

func (c *client) read() {
	defer c.close()
	for {
		data, err := c.transport.Recv()
		if err != nil {
			return
		}
		var m message
		if err := json.Unmarshal(data, &m); err != nil {
			c.parent.Agent().Ensemble().Logf("invalid MCP message: %v", err)
			continue
		}
		if m.Method != "" {
			// Reverse tools may themselves use this peer. Never occupy its reader
			// while executing a tool, or both sides could wait for the same reader.
			if len(m.ID) != 0 {
				go c.reverse(m)
			}
			continue
		}
		var id int
		if json.Unmarshal(m.ID, &id) != nil {
			continue
		}
		c.mu.Lock()
		reply := c.pending[id]
		c.mu.Unlock()
		if reply != nil {
			select {
			case reply <- m:
			default:
			}
		}
	}
}

func (c *client) reverse(m message) {
	response := message{ID: m.ID}
	if m.Method != "tools/call" {
		response.Error = &rpcError{-32601, "method not found"}
	} else {
		var params struct {
			// Name is the advertised dispatch identity, preserved without normalization.
			Name string `json:"name"`
			// Arguments are passed through real Agent dispatch without schema rewriting.
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(m.Params, &params); err != nil || params.Name == "" {
			response.Error = &rpcError{-32602, "tools/call requires name and arguments"}
		} else {
			if len(params.Arguments) == 0 {
				params.Arguments = raw(map[string]any{})
			}
			text, err := c.parent.Agent().Invoke(context.Background(), params.Name, params.Arguments)
			if err != nil {
				text = err.Error()
			}
			response.Result = raw(toolResult{Content: []content{{Type: "text", Text: text}}, IsError: err != nil})
		}
	}
	if err := c.send(response); err != nil {
		c.parent.Agent().Ensemble().Logf("MCP reverse reply: %v", err)
	}
}
