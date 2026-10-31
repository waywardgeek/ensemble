package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Handler is the owning application's sequential receive boundary. Return from
// OnDeliver accepts delivery and triggers its receipt. Handlers must hand long
// work to their own goroutines and must not synchronously call client RPC methods.
type Handler interface {
	// OnDeliver accepts a message before the client acknowledges it.
	OnDeliver(Envelope)
	// OnConnected reports the authenticated welcome before any deliveries.
	OnConnected(WelcomePayload)
	// OnDisconnected reports the terminal connection outcome once.
	OnDisconnected(string)
	// OnError reports malformed traffic or an asynchronous protocol refusal.
	OnError(error)
}

// NopHandler supplies optional callbacks for embedding in small applications.
type NopHandler struct{}

// OnDeliver accepts an otherwise ignored message.
func (NopHandler) OnDeliver(Envelope) {}

// OnConnected ignores the authenticated welcome.
func (NopHandler) OnConnected(WelcomePayload) {}

// OnDisconnected ignores connection shutdown.
func (NopHandler) OnDisconnected(string) {}

// OnError ignores an asynchronous diagnostic.
func (NopHandler) OnError(error) {}

// ClientConfig selects one peer identity and an explicit reconnect receipt.
type ClientConfig struct {
	// GatewayURL is the WebSocket endpoint ending in /v1/ws.
	GatewayURL string
	// AgentID selects a configured key; it is not itself authentication.
	AgentID string
	// StaticKey is the client's private identity, kept only in this process.
	StaticKey Keypair
	// GatewayPubKey pins the expected gateway independent of any TLS proxy.
	GatewayPubKey []byte
	// LastAckedSeq resumes messages after the last delivery accepted by this handler.
	LastAckedSeq uint64
	// HandshakeTimeout bounds startup; zero uses ten seconds.
	HandshakeTimeout time.Duration
}

// Client owns one socket, its directional ciphers, and request bookkeeping.
// Reconnect is explicit: retain LastAckedSeq and call Connect with the same identity.
// This object does not own the application's accepted work or private-key file.
// A new connection gets fresh ciphers and RPC IDs; recipient delivery sequences
// remain gateway-owned and survive that transport lifetime.
type Client struct {
	owner     Handler
	conn      *websocket.Conn
	noise     *NoiseKK
	writeMu   sync.Mutex
	pendingMu sync.Mutex
	pending   map[uint64]chan RPC
	next      atomic.Uint64
	ack       atomic.Uint64
	done      chan struct{}
	closeOnce sync.Once
}

// Handler exposes the application owner without copying its state into the client.
func (c *Client) Handler() Handler { return c.owner }

// Connect authenticates and reads welcome before returning. The supplied context
// owns connection lifetime as well as startup; cancelling it closes the socket.
func Connect(ctx context.Context, cfg ClientConfig, owner Handler) (*Client, error) {
	if owner == nil {
		return nil, errors.New("gateway handler is required")
	}
	u, err := url.Parse(cfg.GatewayURL)
	if err != nil {
		return nil, err
	}
	query := u.Query()
	query.Set("agent_id", cfg.AgentID)
	u.RawQuery = query.Encode()
	timeout := cfg.HandshakeTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	startup, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, _, err := websocket.DefaultDialer.DialContext(startup, u.String(), nil)
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			conn.Close()
		}
	}()
	// Deadlines bound both handshake frames and the encrypted hello/welcome pair.
	deadline, _ := startup.Deadline()
	conn.SetReadDeadline(deadline)
	conn.SetWriteDeadline(deadline)
	n, err := NewNoiseKK("initiator", cfg.StaticKey, cfg.GatewayPubKey, []byte("ensemble-gateway-v1:"+cfg.AgentID))
	if err != nil {
		return nil, err
	}
	first, err := n.WriteHandshakeMessage(nil)
	if err != nil {
		return nil, err
	}
	if err = conn.WriteMessage(websocket.BinaryMessage, first); err != nil {
		return nil, err
	}
	kind, data, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if kind != websocket.BinaryMessage {
		return nil, errors.New("expected binary handshake")
	}
	if _, err = n.ReadHandshakeMessage(data); err != nil {
		return nil, err
	}
	c := &Client{owner: owner, conn: conn, noise: n, done: make(chan struct{}), pending: make(map[uint64]chan RPC)}
	c.ack.Store(cfg.LastAckedSeq)
	if err = c.write(startup, Frame{Type: "hello", AgentID: cfg.AgentID, Version: 1, Seq: cfg.LastAckedSeq}); err != nil {
		return nil, err
	}
	frame, err := c.read()
	if err != nil {
		return nil, err
	}
	if frame.Type != "welcome" || frame.Welcome == nil || frame.Welcome.AgentID != cfg.AgentID {
		return nil, errors.New("gateway refused welcome: " + frame.Error)
	}
	conn.SetReadDeadline(time.Time{})
	ok = true
	// Only this goroutine invokes callbacks, preserving traffic order. The
	// application may begin receiving immediately, even before Connect returns;
	// handlers must not assume the caller has assigned its Client pointer yet.
	go c.receive(*frame.Welcome)
	go func() {
		select {
		case <-ctx.Done():
			c.Close()
		case <-c.done:
		}
	}()
	return c, nil
}

// Every writer, including handler receipts and concurrent RPC callers, shares
// one lock. Serializing only the WebSocket write would still race the Noise
// nonce and produce a connection that fails after apparently successful sends.
func (c *Client) write(ctx context.Context, frame Frame) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	deadline := time.Now().Add(10 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	c.conn.SetWriteDeadline(deadline)
	data, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	data, err = c.noise.Encrypt(data, nil)
	if err == nil {
		err = c.conn.WriteMessage(websocket.BinaryMessage, data)
	}
	// An encryption nonce cannot be rolled back after a failed write.
	if err != nil {
		c.Close()
	}
	return err
}
func (c *Client) read() (Frame, error) {
	kind, data, err := c.conn.ReadMessage()
	if err != nil {
		return Frame{}, err
	}
	if kind != websocket.BinaryMessage {
		return Frame{}, errors.New("expected encrypted binary frame")
	}
	data, err = c.noise.Decrypt(data, nil)
	if err != nil {
		return Frame{}, err
	}
	var frame Frame
	err = json.Unmarshal(data, &frame)
	return frame, err
}

// Callbacks and receipt bookkeeping have one reader. RPC callers use independent
// buffered replies, so a slow caller cannot make a different request consume its
// answer. The callback contract deliberately leaves accepted-work persistence
// to the application: a network receipt cannot establish what a model executed.
func (c *Client) receive(welcome WelcomePayload) {
	reason := "connection closed"
	defer func() { c.Close(); c.owner.OnDisconnected(reason); close(c.done) }()
	c.owner.OnConnected(welcome)
	for {
		frame, err := c.read()
		if err != nil {
			reason = err.Error()
			return
		}
		switch frame.Type {
		case "deliver":
			if frame.Envelope == nil {
				c.owner.OnError(errors.New("delivery missing envelope"))
				return
			}
			c.owner.OnDeliver(*frame.Envelope)
			// Handler acceptance precedes the receipt. If the write fails, reconnecting
			// without the newly returned receipt safely delivers the message again.
			if err = c.write(context.Background(), Frame{Type: "ack", Seq: frame.Envelope.Seq}); err != nil {
				reason = err.Error()
				return
			}
			c.ack.Store(frame.Envelope.Seq)
		case "mcp_response":
			// Late replies belong to their original caller. Never hand an expired
			// request's result to whichever operation happens to be waiting now.
			if frame.RPC == nil {
				c.owner.OnError(errors.New("missing MCP response"))
				return
			}
			c.pendingMu.Lock()
			reply := c.pending[frame.RPC.ID]
			c.pendingMu.Unlock()
			if reply != nil {
				select {
				case reply <- *frame.RPC:
				default:
				}
			}
		case "error":
			c.owner.OnError(errors.New(frame.Error))
		case "pong":
		default:
			c.owner.OnError(fmt.Errorf("unexpected gateway frame %q", frame.Type))
			return
		}
	}
}

// Request performs one correlated MCP method over the same authenticated socket.
func (c *Client) Request(ctx context.Context, method string, params any) (json.RawMessage, error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	// Install correlation before writing: a fast local gateway may reply before
	// this goroutine starts waiting. A buffered result also tolerates a caller
	// whose deadline expired while the receive goroutine was finding its entry.
	id := c.next.Add(1)
	reply := make(chan RPC, 1)
	c.pendingMu.Lock()
	c.pending[id] = reply
	c.pendingMu.Unlock()
	defer func() { c.pendingMu.Lock(); delete(c.pending, id); c.pendingMu.Unlock() }()
	if err = c.write(ctx, Frame{Type: "mcp_request", RPC: &RPC{JSONRPC: "2.0", ID: id, Method: method, Params: raw}}); err != nil {
		return nil, err
	}
	select {
	case response := <-reply:
		if response.Error != nil {
			return nil, errors.New(response.Error.Message)
		}
		return response.Result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, errors.New("gateway disconnected")
	}
}

// CallTool invokes an administration tool and returns its MCP result object.
func (c *Client) CallTool(ctx context.Context, name string, arguments any) (json.RawMessage, error) {
	data, err := c.Request(ctx, "tools/call", map[string]any{"name": name, "arguments": arguments})
	if err != nil {
		return nil, err
	}
	var result struct {
		// IsError distinguishes an executed tool refusal from a protocol failure.
		IsError bool `json:"isError"`
		// Content carries the gateway's safe tool diagnostic or JSON result text.
		Content []struct {
			// Text is one MCP text content item.
			Text string `json:"text"`
		} `json:"content"`
	}
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	if result.IsError {
		message := "gateway tool failed"
		if len(result.Content) > 0 {
			message = result.Content[0].Text
		}
		return nil, errors.New(message)
	}
	return data, nil
}

// Send returns once gateway authorization and queue admission have succeeded.
func (c *Client) Send(ctx context.Context, message OutboundMessage) error {
	_, err := c.CallTool(ctx, "send-message", message)
	return err
}

// Respond preserves the conversation identity and routes back to the authenticated sender.
func (c *Client) Respond(ctx context.Context, request Envelope, content string) error {
	return c.Send(ctx, OutboundMessage{ToAgent: request.Source.UserID, Content: content, ConversationID: request.ConversationID})
}

// DeclareContactable replaces this agent's accepted sender groups, including an empty set.
func (c *Client) DeclareContactable(ctx context.Context, groups []string) error {
	_, err := c.CallTool(ctx, "declare-contactable", map[string]any{"groups": groups})
	return err
}

// RegisterChild creates a distinct identity with a subset of the caller's groups.
func (c *Client) RegisterChild(ctx context.Context, child ChildRegistration) error {
	_, err := c.CallTool(ctx, "register-child", child)
	return err
}

// LastAckedSeq returns the last receipt successfully written after handler acceptance.
func (c *Client) LastAckedSeq() uint64 { return c.ack.Load() }

// Done closes after the receive goroutine and its final callback have returned.
func (c *Client) Done() <-chan struct{} { return c.done }

// Close interrupts network I/O; it does not wait on a possibly active handler.
// A callback may itself close the client, so joining the reader here would
// deadlock that callback. Applications join through Done after releasing any
// work their handlers are waiting on.
func (c *Client) Close() { c.closeOnce.Do(func() { c.conn.Close() }) }
