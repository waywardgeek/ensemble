package gui

import (
	"context"
	"encoding/json"
	"ensemble/ensemble"
	"fmt"
	"io"
	"strings"
	"sync"
)

// WSTransport carries reliable MCP messages through one browser's existing
// writer queue. It is not the lossy progress broadcast and never targets all tabs.
type WSTransport struct {
	parent   ensemble.MCPTransportOwner
	queue    chan [][]byte
	incoming chan json.RawMessage
	done     chan struct{}
	once     sync.Once
	// Relay requests share the browser writer, but not the Agent MCP lifetime.
	relayDone chan struct{}
	relayMu   sync.Mutex
	relayNext uint64
	pending   map[string]chan json.RawMessage
}

func newWSTransport(parent ensemble.MCPTransportOwner, queue chan [][]byte) *WSTransport {
	if parent == nil {
		panic("WebSocket transport requires Server")
	}
	return &WSTransport{
		parent: parent, queue: queue, incoming: make(chan json.RawMessage),
		done: make(chan struct{}), relayDone: make(chan struct{}),
		pending: make(map[string]chan json.RawMessage),
	}
}

// Server exposes the actual owner of this browser connection.
func (t *WSTransport) Server() ensemble.MCPTransportOwner { return t.parent }

// Send frames one RPC onto the same serialized writer used by GUI messages.
func (t *WSTransport) Send(message json.RawMessage) error {
	data, err := json.Marshal(map[string]any{"type": "jsonrpc", "payload": message})
	if err != nil {
		return err
	}
	select {
	case t.queue <- [][]byte{data}:
		return nil
	case <-t.done:
		return io.EOF
	}
}

// Recv receives only this browser's protocol frames, preserving their order.
func (t *WSTransport) Recv() (json.RawMessage, error) {
	select {
	case data := <-t.incoming:
		return data, nil
	case <-t.done:
		return nil, io.EOF
	}
}

// Close ends MCP without disconnecting the user's independent chat interface.
func (t *WSTransport) Close() error { t.once.Do(func() { close(t.done) }); return nil }

// relay reserves string IDs outside the core client's numeric namespace. The
// caller's ID is restored only after a reply arrives from this same browser.
func (t *WSTransport) relay(ctx context.Context, request map[string]json.RawMessage) (json.RawMessage, error) {
	original := request["id"]
	t.relayMu.Lock()
	t.relayNext++
	id := fmt.Sprintf("gui-relay-%d", t.relayNext)
	reply := make(chan json.RawMessage, 1)
	if original != nil {
		t.pending[id] = reply
	}
	t.relayMu.Unlock()
	defer func() { t.relayMu.Lock(); delete(t.pending, id); t.relayMu.Unlock() }()
	if original != nil {
		request["id"], _ = json.Marshal(id)
	}
	data, err := json.Marshal(map[string]any{"type": "jsonrpc", "payload": request})
	if err != nil {
		return nil, err
	}
	select {
	case t.queue <- [][]byte{data}:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-t.relayDone:
		return nil, io.EOF
	}
	if original == nil {
		return nil, nil
	} // Notifications never receive a reply.
	select {
	case data := <-reply:
		var response map[string]json.RawMessage
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, err
		}
		response["id"] = original
		return json.Marshal(response)
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-t.relayDone:
		return nil, io.EOF
	}
}

// relayReply consumes only our private ID namespace. Late replies after a TCP
// disconnect must not fall through to, or block, the independent Agent client.
func (t *WSTransport) relayReply(data json.RawMessage) bool {
	var response struct {
		// ID recognizes replies reserved for this browser relay.
		ID string `json:"id"`
	}
	if json.Unmarshal(data, &response) != nil || !strings.HasPrefix(response.ID, "gui-relay-") {
		return false
	}
	t.relayMu.Lock()
	defer t.relayMu.Unlock()
	if reply := t.pending[response.ID]; reply != nil {
		delete(t.pending, response.ID)
		reply <- data
	}
	return true
}
