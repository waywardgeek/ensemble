// Package mcp provides a JSON-RPC 2.0 client for the Model Context Protocol.
//
// MCP is the tool seam. External tools join the agent's registry at runtime
// via a standard protocol — same internal tool interface, different origin.
// The chapter builds three transports (pipe, stdio, websocket), a client that
// drives the MCP handshake and tool calls, and a bridge that converts MCP
// tools into the agent's own tool registry. Every tool call goes through the
// existing job infrastructure: MCP tools are processes, not functions.
package mcp

import (
	"encoding/json"
	"fmt"
	"sync"
)

// JSON-RPC 2.0 message types.

// Request is a JSON-RPC 2.0 request or notification.
// When ID is nil, the message is a notification (no response expected).
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is the error object inside a JSON-RPC 2.0 response.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// Codec handles outgoing request correlation and incoming response routing.
// A background goroutine reads from the transport and dispatches responses
// to waiting callers via a map of channels.
type Codec struct {
	transport Transport

	mu      sync.Mutex
	nextID  int
	pending map[int]chan Response
	closed  bool

	// onRequest is called for incoming requests (server → client).
	// Set by the Client to handle reverse tool calls.
	onRequest func(Request)

	done chan struct{} // closed when readLoop exits
	err  error        // first error from readLoop
}

// NewCodec wraps a transport with correlation tracking.
// Call Run to start the read loop.
func NewCodec(t Transport) *Codec {
	return &Codec{
		transport: t,
		nextID:    1,
		pending:   make(map[int]chan Response),
		done:      make(chan struct{}),
	}
}

// Run starts the background read loop. Call this once after creating the codec.
func (c *Codec) Run() {
	go c.readLoop()
}

// Call sends a request and waits for the correlated response.
func (c *Codec) Call(method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("codec closed")
	}
	id := c.nextID
	c.nextID++
	ch := make(chan Response, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	var raw json.RawMessage
	if params != nil {
		var err error
		raw, err = json.Marshal(params)
		if err != nil {
			c.mu.Lock()
			delete(c.pending, id)
			c.mu.Unlock()
			return nil, fmt.Errorf("marshal params: %w", err)
		}
	}

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  raw,
	}
	data, err := json.Marshal(req)
	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}
	if err := c.transport.Send(data); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}

	// Wait for the response or codec shutdown.
	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Result, nil
	case <-c.done:
		return nil, fmt.Errorf("codec closed while waiting for response to %s (id=%d)", method, id)
	}
}

// Notify sends a notification (no ID, no response expected).
func (c *Codec) Notify(method string, params any) error {
	var raw json.RawMessage
	if params != nil {
		var err error
		raw, err = json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
	}
	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  raw,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.transport.Send(data)
}

// Respond sends a JSON-RPC response (for incoming requests in reverse mode).
func (c *Codec) Respond(id any, result any, rpcErr *RPCError) error {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   rpcErr,
	}
	if rpcErr == nil && result != nil {
		raw, err := json.Marshal(result)
		if err != nil {
			return err
		}
		resp.Result = raw
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.transport.Send(data)
}

// Close shuts down the codec and its transport.
func (c *Codec) Close() error {
	c.mu.Lock()
	c.closed = true
	// Unblock all waiting callers.
	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}
	c.mu.Unlock()
	return c.transport.Close()
}

// readLoop reads messages from the transport and routes them.
func (c *Codec) readLoop() {
	defer close(c.done)
	for {
		raw, err := c.transport.Recv()
		if err != nil {
			c.mu.Lock()
			c.err = err
			c.closed = true
			for id, ch := range c.pending {
				close(ch)
				delete(c.pending, id)
			}
			c.mu.Unlock()
			return
		}
		c.dispatch(raw)
	}
}

// dispatch routes an incoming message as either a response or a request.
func (c *Codec) dispatch(raw json.RawMessage) {
	// Peek at the message to determine its type.
	var peek struct {
		ID     any    `json:"id"`
		Method string `json:"method"`
		Result any    `json:"result"`
		Error  any    `json:"error"`
	}
	if json.Unmarshal(raw, &peek) != nil {
		return
	}

	// A message with a method is a request or notification.
	if peek.Method != "" {
		// If it also has an ID, it is a request (expects a response).
		// If no ID, it is a notification (no response needed).
		var req Request
		if json.Unmarshal(raw, &req) != nil {
			return
		}
		if c.onRequest != nil {
			c.onRequest(req)
		}
		return
	}

	// Otherwise it is a response — route to the waiting caller.
	var resp Response
	if json.Unmarshal(raw, &resp) != nil {
		return
	}
	// The ID comes back as a float64 from JSON unmarshalling.
	id := toInt(resp.ID)
	if id == 0 {
		return
	}
	c.mu.Lock()
	ch, ok := c.pending[id]
	if ok {
		delete(c.pending, id)
	}
	c.mu.Unlock()
	if ok {
		ch <- resp
	}
}

// toInt converts a JSON number (float64) or int to int.
func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}
