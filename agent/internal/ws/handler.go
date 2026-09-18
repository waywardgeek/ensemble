// Package ws provides the WebSocket transport for the GUI.
//
// The agent does not know this package exists. The hub is an Observer: it
// receives observations from the actor, serialises them, and fans them out
// to every connected browser. No package under internal/llm, internal/tools,
// or internal/jobs imports this one.
package ws

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/waywardgeek/coding-agents-course/agent/internal/common"
)

func newUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
}

// Hub fans observations out to WebSocket clients. It implements
// common.Observer; its Observe method MUST NOT BLOCK.
type Hub struct {
	mu       sync.Mutex
	messages [][]byte         // append-only buffer of serialised JSON
	clients  map[*Client]bool // all connected clients
	gate     *common.PauseGate
	send     func(common.Inbound) // forwards prompt/hint/interrupt to the actor
	guiLog   *GuiLogger
}

// NewHub creates a hub. send is called for every prompt/hint/interrupt
// received from a browser; gate controls tool-dispatch pausing.
func NewHub(gate *common.PauseGate, send func(common.Inbound), guiLogPath string) *Hub {
	h := &Hub{
		clients: make(map[*Client]bool),
		gate:    gate,
		send:    send,
		guiLog:  NewGuiLogger(guiLogPath),
	}
	return h
}

// Close closes the gui.log file.
func (h *Hub) Close() {
	if h.guiLog != nil {
		h.guiLog.Close()
	}
}

// Observe implements common.Observer. Must not block.
func (h *Hub) Observe(obs common.Observation) {
	data := marshalObservation(obs)
	if data == nil {
		return
	}

	h.mu.Lock()
	h.messages = append(h.messages, data)
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		if c.live {
			clients = append(clients, c)
		}
	}
	h.mu.Unlock()

	h.guiLog.Log("<", data)
	for _, c := range clients {
		select {
		case c.send <- data:
		default:
			// Slow client. Drop rather than block the actor.
		}
	}
}

// ServeWS upgrades an HTTP request to a WebSocket connection.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := newUpgrader().Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()

	go c.writePump()
	c.readPump() // blocks until disconnect

	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	close(c.send)
}

// subscribe replays messages from cursor and marks the client live.
func (h *Hub) subscribe(c *Client, cursor uint64) {
	h.mu.Lock()
	var replay [][]byte
	if cursor < uint64(len(h.messages)) {
		replay = make([][]byte, uint64(len(h.messages))-cursor)
		copy(replay, h.messages[cursor:])
	}
	c.live = true
	h.mu.Unlock()

	for _, m := range replay {
		select {
		case c.send <- m:
		default:
		}
	}
}

// handleClientMessage dispatches an incoming client message.
func (h *Hub) handleClientMessage(c *Client, raw []byte) {
	h.guiLog.Log(">", raw)

	var msg struct {
		Type   string `json:"type"`
		Cursor uint64 `json:"cursor"`
		Text   string `json:"text"`
	}
	if json.Unmarshal(raw, &msg) != nil {
		return
	}

	switch msg.Type {
	case "subscribe":
		h.subscribe(c, msg.Cursor)
	case "prompt":
		if msg.Text != "" {
			h.send(common.UserMessage{Text: msg.Text})
		}
	case "hint":
		if msg.Text != "" {
			h.send(common.Hint{Text: msg.Text})
		}
	case "interrupt":
		h.send(common.Interrupt{})
	case "pause":
		if h.gate != nil {
			h.gate.Pause()
		}
	case "unpause":
		if h.gate != nil {
			h.gate.Unpause()
		}
	}
}

// ----------------------------------------------------------------
// Client
// ----------------------------------------------------------------

// Client represents a single WebSocket connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	live bool // receives live messages only after subscribe
}

// readPump reads messages from the WebSocket and dispatches them.
func (c *Client) readPump() {
	defer c.conn.Close()
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		c.hub.handleClientMessage(c, data)
	}
}

// writePump writes messages from the send channel to the WebSocket.
func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// ----------------------------------------------------------------
// Observation → JSON
// ----------------------------------------------------------------

func marshalObservation(obs common.Observation) []byte {
	var m map[string]any

	switch v := obs.(type) {
	case common.PartDelta:
		m = map[string]any{
			"type":    "part_delta",
			"part_id": v.PartID,
			"kind":    v.Kind.String(),
			"chunk":   v.Chunk,
		}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
	case common.PartFinal:
		m = map[string]any{
			"type":    "part_final",
			"part_id": v.PartID,
			"seq":     v.Seq,
		}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		switch p := v.Part.(type) {
		case common.TextPart:
			m["text"] = p.Text
		case common.ToolCallPart:
			m["tool"] = p.Name
			m["args"] = string(p.Args)
		case common.OpaquePart:
			m["opaque"] = true
		}
	case common.StateChanged:
		m = map[string]any{
			"type": "state_changed",
			"from": v.From.String(),
			"to":   v.To.String(),
		}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
	case common.TurnEnded:
		m = map[string]any{
			"type": "turn_ended",
			"text": v.Text,
		}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		if v.Err != "" {
			m["error"] = v.Err
		}
	case common.ToolDispatched:
		m = map[string]any{
			"type":    "tool_dispatched",
			"call_id": v.CallID,
			"name":    v.Name,
		}
		if v.Input != nil {
			m["input"] = json.RawMessage(v.Input)
		}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
	case common.ToolFinished:
		m = map[string]any{
			"type":     "tool_finished",
			"call_id":  v.CallID,
			"result":   v.Result,
			"is_error": v.IsError,
		}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
	default:
		return nil
	}

	data, _ := json.Marshal(m)
	return data
}
