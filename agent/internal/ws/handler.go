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
	"github.com/waywardgeek/ensemble/agent/internal/common"
)

func newUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
}

// Hub fans observations out to WebSocket clients. It implements
// common.Observer; its Observe method MUST NOT BLOCK.
//
// Reconnection uses the event log, not an observation buffer. The event log
// is append-only so readers take no lock (Log.Events elements are stable once
// written). The only mutable shared state is the in-flight partial map and
// the client set, which are guarded by mu.
type Hub struct {
	mu        sync.Mutex
	clients   map[*Client]bool
	inflight  map[uint64][]byte // part_id → accumulated partial JSON
	logLen    int                // last-observed len(log.Events), updated under mu
	gate      *common.PauseGate
	send      func(common.Inbound) // forwards prompt/hint/interrupt to the actor
	guiLog    *GuiLogger
	log       *common.Log // event log — read-only access for reconnection
}

// NewHub creates a hub. send is called for every prompt/hint/interrupt
// received from a browser; gate controls tool-dispatch pausing; eventLog
// provides read access to the append-only event log for reconnection.
func NewHub(gate *common.PauseGate, send func(common.Inbound), guiLogPath string, eventLog *common.Log) *Hub {
	h := &Hub{
		clients:  make(map[*Client]bool),
		inflight: make(map[uint64][]byte),
		gate:     gate,
		send:     send,
		guiLog:   NewGuiLogger(guiLogPath),
		log:      eventLog,
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
	// Update in-flight state and build the wire message for this observation.
	data := marshalObservation(obs)
	if data == nil {
		return
	}

	h.mu.Lock()
	switch v := obs.(type) {
	case common.PartDelta:
		h.inflight[v.PartID] = append(h.inflight[v.PartID], v.Chunk...)
	case common.PartFinal:
		delete(h.inflight, v.PartID)
	}
	// Record the event log length. Observe runs on the engine goroutine
	// which has already appended to Log.Events, so len() sees the latest
	// state. Storing under mu creates a happens-before for subscribe on
	// another goroutine. The event log itself needs no lock — elements
	// are immutable once written, and a slow WebSocket send only touches
	// elements that existed before this snapshot.
	h.logLen = len(h.log.Events)
	logLen := h.logLen
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		if c.live {
			clients = append(clients, c)
		}
	}
	h.mu.Unlock()

	h.guiLog.Log("<", data)

	// Live clients receive all observations directly — deltas, finals,
	// tool events, state changes, and turn boundaries. catchUp is NOT
	// called for live clients (it's reconnection-only, called from
	// subscribe). Instead, lastSeq advances so reconnection starts from
	// the right place. The user's prompt is delivered separately via the
	// echo in handleClientMessage above.
	var latestSeq common.Seq
	if logLen > 0 {
		latestSeq = h.log.Events[logLen-1].Seq
	}

	for _, c := range clients {
		select {
		case c.send <- data:
		default:
			// Slow client. Drop rather than block the actor.
		}
		if latestSeq > c.lastSeq {
			c.lastSeq = latestSeq
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

// subscribe reports the available event-log range and marks the client
// live so it receives new observations going forward. The client then
// calls fetch to pull the events it needs.
func (h *Hub) subscribe(c *Client) {
	h.mu.Lock()
	logLen := h.logLen
	h.mu.Unlock()

	events := h.log.Events
	if logLen > len(events) {
		logLen = len(events)
	}

	// Report the available range so the client can decide what to fetch.
	var firstSeq, lastSeq common.Seq
	if logLen > 0 {
		firstSeq = events[0].Seq
		lastSeq = events[logLen-1].Seq
	}

	rangeMsg, _ := json.Marshal(map[string]any{
		"type":  "event_range",
		"first": firstSeq,
		"last":  lastSeq,
	})
	select {
	case c.send <- rangeMsg:
	default:
	}
	h.guiLog.Log("<", rangeMsg)

	h.mu.Lock()
	c.lastSeq = lastSeq
	c.live = true
	h.mu.Unlock()
}

// fetch sends event-log entries in the requested seq range, followed by
// any in-flight partials. The client calls this after receiving event_range.
func (h *Hub) fetch(c *Client, fromSeq, toSeq common.Seq) {
	h.mu.Lock()
	logLen := h.logLen
	h.mu.Unlock()

	events := h.log.Events
	if logLen > len(events) {
		logLen = len(events)
	}

	// Send renderable events in [fromSeq, toSeq].
	for i := 0; i < logLen; i++ {
		e := events[i]
		if e.Seq < fromSeq {
			continue
		}
		if e.Seq > toSeq {
			break
		}
		if msgs := renderEvent(e); len(msgs) > 0 {
			for _, data := range msgs {
				select {
				case c.send <- data:
				default:
				}
			}
		}
	}

	// Send in-flight partials so the client catches up to the live stream.
	h.mu.Lock()
	for partID, accumulated := range h.inflight {
		data := marshalPartialState(partID, accumulated)
		if data != nil {
			select {
			case c.send <- data:
			default:
			}
		}
	}
	// Advance lastSeq so the client doesn't re-fetch on the next subscribe.
	if toSeq > c.lastSeq {
		c.lastSeq = toSeq
	}
	h.mu.Unlock()
}

// handleClientMessage dispatches an incoming client message.
func (h *Hub) handleClientMessage(c *Client, raw []byte) {
	h.guiLog.Log(">", raw)

	var msg struct {
		Type string     `json:"type"`
		Text string     `json:"text"`
		From common.Seq `json:"from"`
		To   common.Seq `json:"to"`
	}
	if json.Unmarshal(raw, &msg) != nil {
		return
	}

	switch msg.Type {
	case "subscribe":
		h.subscribe(c)
	case "fetch":
		h.fetch(c, msg.From, msg.To)
	case "prompt":
		if msg.Text != "" {
			// Echo the prompt back to all connected clients as a message
			// event immediately, before forwarding to the actor. The event
			// log will also record a MessageReceived event, but live clients
			// need the prompt now — catchUp would deliver it too late.
			echo, _ := json.Marshal(map[string]any{
				"type":  "message",
				"actor": "user",
				"text":  msg.Text,
			})
			h.mu.Lock()
			for cl := range h.clients {
				if cl.live {
					select {
					case cl.send <- echo:
					default:
					}
				}
			}
			h.mu.Unlock()
			h.guiLog.Log("<", echo)
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
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	live    bool       // receives live messages only after subscribe
	lastSeq common.Seq // last event-log Seq delivered to this client
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
// Event log → wire messages
// ----------------------------------------------------------------

// isRenderable returns true if the event should be sent to the GUI.
func isRenderable(e common.Event) bool {
	switch e.Type {
	case common.ResponseEnded, common.ToolCalled, common.ToolReturned,
		common.MessageReceived, common.ErrorOccurred:
		return true
	}
	return false
}

// renderEvent converts an event-log entry into zero or more wire messages.
// Each returned []byte is a complete JSON object ready for the WebSocket.
func renderEvent(e common.Event) [][]byte {
	var msgs [][]byte

	switch e.Type {
	case common.ResponseEnded:
		if e.Response == nil {
			break
		}
		for _, p := range e.Response.Parts {
			m := partToWireMsg(e.Seq, p)
			if m != nil {
				data, _ := json.Marshal(m)
				msgs = append(msgs, data)
			}
		}

	case common.ToolCalled:
		if e.Tool == nil {
			break
		}
		m := map[string]any{
			"type":    "tool_dispatched",
			"seq":     e.Seq,
			"call_id": e.Tool.CallID,
			"name":    e.Tool.Name,
		}
		if e.Tool.Args != nil {
			m["input"] = json.RawMessage(e.Tool.Args)
		}
		data, _ := json.Marshal(m)
		msgs = append(msgs, data)

	case common.ToolReturned:
		if e.Tool == nil {
			break
		}
		m := map[string]any{
			"type":    "tool_finished",
			"seq":     e.Seq,
			"call_id": e.Tool.CallID,
		}
		isError := e.Tool.IsError
		m["is_error"] = isError
		// Send the result text from the tool result parts.
		if len(e.Tool.Parts) > 0 {
			var result string
			for _, p := range e.Tool.Parts {
				if tp, ok := p.(common.TextPart); ok {
					result += tp.Text
				}
			}
			m["result"] = result
		}
		data, _ := json.Marshal(m)
		msgs = append(msgs, data)

	case common.MessageReceived:
		if e.Message == nil {
			break
		}
		m := map[string]any{
			"type":  "message",
			"seq":   e.Seq,
			"actor": e.Message.Actor.String(),
		}
		var text string
		for _, p := range e.Message.Parts {
			if tp, ok := p.(common.TextPart); ok {
				text += tp.Text
			}
		}
		m["text"] = text
		data, _ := json.Marshal(m)
		msgs = append(msgs, data)

	case common.ErrorOccurred:
		if e.Error == nil {
			break
		}
		m := map[string]any{
			"type":    "error",
			"seq":     e.Seq,
			"message": e.Error.Message,
		}
		data, _ := json.Marshal(m)
		msgs = append(msgs, data)
	}

	return msgs
}

// partToWireMsg converts a finalized Part into a wire message.
func partToWireMsg(seq common.Seq, p common.Part) map[string]any {
	switch v := p.(type) {
	case common.TextPart:
		return map[string]any{
			"type": "part_final",
			"seq":  seq,
			"kind": "text",
			"text": v.Text,
		}
	case common.ToolCallPart:
		return map[string]any{
			"type":    "part_final",
			"seq":     seq,
			"kind":    "tool_call",
			"tool":    v.Name,
			"call_id": v.CallID,
			"args":    string(v.Args),
		}
	}
	return nil
}

// ----------------------------------------------------------------
// Observation → wire JSON (for live streaming)
// ----------------------------------------------------------------

// marshalObservation converts a live observation to a wire message.
// These are ephemeral — streaming deltas and state changes.
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
		// PartFinal observations are delivered via the event log on
		// reconnect. For live clients they still arrive as observations
		// so the GUI can switch from streaming to finalized rendering.
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

// marshalPartialState creates a wire message for in-flight partial content.
// Sent on subscribe so a new client can see what's currently streaming.
func marshalPartialState(partID uint64, accumulated []byte) []byte {
	m := map[string]any{
		"type":    "part_partial",
		"part_id": partID,
		"content": string(accumulated),
	}
	data, _ := json.Marshal(m)
	return data
}
