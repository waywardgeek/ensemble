package ws

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// ServeMCP exposes the browser's MCP server on a TCP listener, so an agent in
// another process can see and drive the GUI the same way the in-process agent
// does through the gui-debug skill.
//
// The wire is the stdio wire: one JSON-RPC message per line. That choice makes
// the client side trivial. Any MCP client that can launch a subprocess reaches
// this port through a stdio-to-TCP pipe (cmd/mcp-connect), with no transport
// code of its own.
//
// Each connection becomes one more source tag in the hub's existing routing:
// requests are broadcast to the browsers tagged "tcp-N", and mcp.js echoes the
// tag on its reply, which lands back here. Nothing in the browser changes.
//
// ServeMCP returns when the listener is closed. Bind it to loopback: anyone who
// can reach this port can click buttons and type into the GUI.
func (h *Hub) ServeMCP(ln net.Listener) error {
	n := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		n++
		go h.serveMCPConn(conn, fmt.Sprintf("tcp-%d", n))
	}
}

// mcpReplyTimeout bounds how long a browser reply waits for a wedged TCP
// client. The receiver runs on the browser's read loop, so it must never block
// forever; a reply dropped after this long belongs to a client that has stopped
// reading anyway.
const mcpReplyTimeout = 5 * time.Second

// mcpLineLimit is the largest single message accepted from a TCP client.
const mcpLineLimit = 16 << 20

// rpcHead is the part of a JSON-RPC message the relay needs to route it.
type rpcHead struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
}

func (h *Hub) serveMCPConn(conn net.Conn, source string) {
	out := make(chan []byte, 64)
	done := make(chan struct{})

	// Every live tab answers a broadcast request; the router lets only the
	// first reply per id through, so two open tabs do not send the client two
	// responses with one id.
	replies := newReplyRouter()

	reply := func(b json.RawMessage) {
		select {
		case out <- b:
		case <-done:
		case <-time.After(mcpReplyTimeout):
		}
	}

	h.SetMCPReceiver(source, func(payload json.RawMessage) {
		var head rpcHead
		if json.Unmarshal(payload, &head) != nil {
			return
		}
		if head.ID != nil && head.Method == "" {
			replies.route(head.ID, payload) // false: a duplicate, dropped
			return
		}
		reply(payload)
	})

	go func() {
		w := bufio.NewWriter(conn)
		for {
			select {
			case b := <-out:
				w.Write(b)
				w.WriteByte('\n')
				if w.Flush() != nil {
					conn.Close()
					return
				}
			case <-done:
				return
			}
		}
	}()

	sc := bufio.NewScanner(conn)
	sc.Buffer(make([]byte, 64<<10), mcpLineLimit)
	for sc.Scan() {
		line := append([]byte(nil), sc.Bytes()...)
		var head rpcHead
		if json.Unmarshal(line, &head) != nil {
			continue
		}
		isRequest := head.ID != nil && head.Method != ""
		if isRequest {
			replies.expect(head.ID, reply)
		}
		if h.BroadcastJSONRPC(line, source) == 0 && isRequest {
			// No browser is open, so nothing will ever answer. Say so now
			// rather than leaving the client to wait out its timeout.
			replies.forget(head.ID)
			reply(noGUIError(head.ID))
		}
	}

	h.RemoveMCPReceiver(source)
	close(done)
	conn.Close()
}

// noGUIError is the JSON-RPC error returned for a request that reached no
// browser.
func noGUIError(id json.RawMessage) []byte {
	b, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    -32000,
			"message": string(ErrNoGUI),
		},
	})
	return b
}
