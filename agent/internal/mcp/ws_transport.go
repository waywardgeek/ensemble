package mcp

import (
	"encoding/json"
	"fmt"
)

// WSTransport tunnels JSON-RPC over the existing WebSocket connection.
// The Hub routes messages with "type":"jsonrpc" to/from this transport.
//
// Outgoing: WSTransport calls sendFn, which writes {"type":"jsonrpc","data":{...}}
// to all connected WebSocket clients.
//
// Incoming: The Hub's handleClientMessage calls Deliver when it receives a
// "jsonrpc" message, which puts the data on the receive channel.
type WSTransport struct {
	sendFn func(json.RawMessage) // write to WebSocket clients
	recv   chan json.RawMessage
	done   chan struct{}
}

// NewWSTransport creates a transport that tunnels over WebSocket.
// sendFn is called to broadcast a JSON-RPC message to connected clients.
func NewWSTransport(sendFn func(json.RawMessage)) *WSTransport {
	return &WSTransport{
		sendFn: sendFn,
		recv:   make(chan json.RawMessage, 64),
		done:   make(chan struct{}),
	}
}

func (w *WSTransport) Send(msg json.RawMessage) error {
	select {
	case <-w.done:
		return fmt.Errorf("ws transport closed")
	default:
	}
	w.sendFn(msg)
	return nil
}

func (w *WSTransport) Recv() (json.RawMessage, error) {
	select {
	case msg, ok := <-w.recv:
		if !ok {
			return nil, fmt.Errorf("ws transport closed")
		}
		return msg, nil
	case <-w.done:
		return nil, fmt.Errorf("ws transport closed")
	}
}

func (w *WSTransport) Close() error {
	select {
	case <-w.done:
	default:
		close(w.done)
	}
	return nil
}

// Deliver is called by the Hub when a "jsonrpc" message arrives from a client.
func (w *WSTransport) Deliver(data json.RawMessage) {
	select {
	case w.recv <- data:
	case <-w.done:
	}
}
