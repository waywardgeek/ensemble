package mcp

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// ClientWSTransport connects to a WebSocket hub as a client and tunnels
// JSON-RPC messages with a source tag. The virtual user uses this to reach
// the browser's MCP server through the hub.
//
// Outgoing: wraps each message as {"type":"jsonrpc","source":"vu","payload":{...}}
// Incoming: filters for matching source tag and delivers the payload.
type ClientWSTransport struct {
	conn     *websocket.Conn
	source   string             // e.g. "vu"
	incoming chan json.RawMessage
	done     chan struct{}
	closeOnce sync.Once
}

// NewClientWSTransport dials the hub WebSocket and starts a read loop
// that filters for jsonrpc messages matching the given source tag.
func NewClientWSTransport(url, source string) (*ClientWSTransport, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", url, err)
	}

	t := &ClientWSTransport{
		conn:     conn,
		source:   source,
		incoming: make(chan json.RawMessage, 64),
		done:     make(chan struct{}),
	}
	go t.readLoop()
	return t, nil
}

// readLoop reads WebSocket messages, filters for jsonrpc envelopes matching
// our source tag, and delivers payloads to the incoming channel.
func (t *ClientWSTransport) readLoop() {
	defer t.Close()
	for {
		_, data, err := t.conn.ReadMessage()
		if err != nil {
			return
		}
		var envelope struct {
			Type    string          `json:"type"`
			Source  string          `json:"source"`
			Payload json.RawMessage `json:"payload"`
		}
		if json.Unmarshal(data, &envelope) != nil {
			continue
		}
		if envelope.Type != "jsonrpc" || envelope.Source != t.source {
			continue
		}
		if envelope.Payload == nil {
			continue
		}
		select {
		case t.incoming <- envelope.Payload:
		case <-t.done:
			return
		}
	}
}

// Send wraps a JSON-RPC message in a source-tagged envelope and sends it.
func (t *ClientWSTransport) Send(msg json.RawMessage) error {
	select {
	case <-t.done:
		return fmt.Errorf("client ws transport closed")
	default:
	}
	envelope := map[string]any{
		"type":    "jsonrpc",
		"source":  t.source,
		"payload": json.RawMessage(msg),
	}
	return t.conn.WriteJSON(envelope)
}

// Recv blocks until a JSON-RPC response arrives from the hub.
func (t *ClientWSTransport) Recv() (json.RawMessage, error) {
	select {
	case msg, ok := <-t.incoming:
		if !ok {
			return nil, fmt.Errorf("client ws transport closed")
		}
		return msg, nil
	case <-t.done:
		return nil, fmt.Errorf("client ws transport closed")
	}
}

// Close shuts down the transport and the underlying WebSocket connection.
func (t *ClientWSTransport) Close() error {
	t.closeOnce.Do(func() {
		close(t.done)
		t.conn.Close()
	})
	return nil
}
