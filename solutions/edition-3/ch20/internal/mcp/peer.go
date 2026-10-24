// Package mcp owns Agent-scoped JSON-RPC connections, not conversation execution.
package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"

	"ensemble/internal/common"
)

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return fmt.Sprintf("MCP %d: %s", e.Code, e.Message) }

type message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}
type handler func(context.Context, string, json.RawMessage) (any, error)
type peer struct {
	transport common.MCPTransport
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	next      uint64
	writer    chan struct{}
	pending   map[string]chan message
	incoming  map[string]context.CancelFunc
	done      chan struct{}
	once      sync.Once
	err       error
	handle    handler
	drain     bool
	requests  sync.WaitGroup
}

func newPeer(t common.MCPTransport, h handler, drain bool) *peer {
	ctx, cancel := context.WithCancel(context.Background())
	p := &peer{transport: t, ctx: ctx, cancel: cancel, pending: map[string]chan message{}, incoming: map[string]context.CancelFunc{}, done: make(chan struct{}), writer: make(chan struct{}, 1), handle: h, drain: drain}
	go p.read()
	return p
}
func (p *peer) close(err error) {
	p.once.Do(func() {
		p.mu.Lock()
		p.err = err
		p.mu.Unlock()
		p.cancel()
		close(p.done)
		_ = p.transport.Close()
	})
}
func (p *peer) send(ctx context.Context, m message) error {
	m.JSONRPC = "2.0"
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	// Queued cancellation relinquishes only this sender, not the live writer.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return io.ErrClosedPipe
	case p.writer <- struct{}{}:
	}
	defer func() { <-p.writer }()
	if err := ctx.Err(); err != nil {
		return err
	}
	// HTTP cancellation belongs to one request, not the whole connection.
	if t, ok := p.transport.(interface {
		SendContext(context.Context, json.RawMessage) error
	}); ok {
		return t.SendContext(ctx, b)
	}
	// Once writing has started, Close is the transport contract's unblock path.
	stop := context.AfterFunc(ctx, func() { p.close(ctx.Err()) })
	defer stop()
	if err = p.transport.Send(b); err != nil {
		p.close(err)
	}
	return err
}

func (p *peer) notify(ctx context.Context, method string, params any) error {
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return p.send(ctx, message{Method: method, Params: b})
}
func (p *peer) call(ctx context.Context, method string, params, out any) error {
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.next++
	id := strconv.FormatUint(p.next, 10)
	ch := make(chan message, 1)
	p.pending[id] = ch
	p.mu.Unlock()
	defer func() { p.mu.Lock(); delete(p.pending, id); p.mu.Unlock() }()
	if err = p.send(ctx, message{ID: json.RawMessage(id), Method: method, Params: b}); err != nil {
		return err
	}
	select {
	case m := <-ch:
		if m.Error != nil {
			return m.Error
		}
		if len(m.Result) == 0 {
			return errors.New("MCP response missing result")
		}
		return json.Unmarshal(m.Result, out)
	case <-ctx.Done():
		go func() { _ = p.notify(p.ctx, "$/cancelRequest", map[string]any{"id": json.RawMessage(id)}) }()
		return ctx.Err()
	case <-p.done:
		p.mu.Lock()
		defer p.mu.Unlock()
		return p.err
	}
}
func validID(id json.RawMessage) bool {
	var v any
	if json.Unmarshal(id, &v) != nil {
		return false
	}
	switch v.(type) {
	case string, float64:
		return true
	}
	return false
}
func (p *peer) read() {
	for {
		raw, err := p.transport.Recv()
		if err != nil {
			if p.drain && errors.Is(err, io.EOF) {
				p.requests.Wait()
			}
			p.close(err)
			return
		}
		var m message
		if json.Unmarshal(raw, &m) != nil || m.JSONRPC != "2.0" {
			_ = p.send(p.ctx, message{ID: json.RawMessage("null"), Error: &rpcError{-32600, "invalid JSON-RPC message"}})
			continue
		}
		if m.Method == "" {
			if !validID(m.ID) {
				continue
			}
			p.mu.Lock()
			ch := p.pending[string(m.ID)]
			p.mu.Unlock()
			if ch != nil {
				select {
				case ch <- m:
				default:
				}
			}
			continue
		}
		if len(m.ID) > 0 && !validID(m.ID) {
			_ = p.send(p.ctx, message{ID: json.RawMessage("null"), Error: &rpcError{-32600, "invalid request id"}})
			continue
		}
		if m.Method == "$/cancelRequest" && len(m.ID) == 0 {
			var params struct {
				ID json.RawMessage `json:"id"`
			}
			if json.Unmarshal(m.Params, &params) == nil {
				p.mu.Lock()
				cancel := p.incoming[string(params.ID)]
				p.mu.Unlock()
				if cancel != nil {
					cancel()
				}
			}
			continue
		}
		ctx, cancel := context.WithCancel(p.ctx)
		if len(m.ID) > 0 {
			p.mu.Lock()
			_, duplicate := p.incoming[string(m.ID)]
			if !duplicate {
				p.incoming[string(m.ID)] = cancel
			}
			p.mu.Unlock()
			if duplicate {
				cancel()
				_ = p.send(p.ctx, message{ID: m.ID, Error: &rpcError{-32600, "duplicate request id"}})
				continue
			}
		}
		// Incoming calls never wait on the reader: reverse calls can run while an
		// outgoing context or tool call is waiting for its correlated response.
		p.requests.Add(1)
		if m.Method == "initialize" || m.Method == "notifications/initialized" {
			p.request(ctx, cancel, m)
		} else {
			go p.request(ctx, cancel, m)
		}
	}
}
func (p *peer) request(ctx context.Context, cancel context.CancelFunc, m message) {
	defer cancel()
	defer func() { p.mu.Lock(); delete(p.incoming, string(m.ID)); p.mu.Unlock() }()
	defer p.requests.Done()
	if p.handle == nil {
		return
	}
	result, err := p.handle(ctx, m.Method, m.Params)
	if len(m.ID) == 0 {
		return
	}
	reply := message{ID: m.ID}
	if err != nil {
		var rpc *rpcError
		if !errors.As(err, &rpc) {
			rpc = &rpcError{-32603, err.Error()}
		}
		reply.Error = rpc
	} else {
		reply.Result, err = json.Marshal(result)
		if err != nil {
			reply.Error = &rpcError{-32603, err.Error()}
		}
	}
	_ = p.send(p.ctx, reply)
}
