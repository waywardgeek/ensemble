package gui

import (
	"encoding/json"
	"ensemble/ensemble"
	"io"
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
}

func newWSTransport(parent ensemble.MCPTransportOwner, queue chan [][]byte) *WSTransport {
	if parent == nil {
		panic("WebSocket transport requires Server")
	}
	return &WSTransport{parent: parent, queue: queue, incoming: make(chan json.RawMessage), done: make(chan struct{})}
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
