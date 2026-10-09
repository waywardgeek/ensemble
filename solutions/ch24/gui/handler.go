// Package ws provides the WebSocket transport for the GUI.
//
// The agent does not know this package exists. The hub is an Observer: it
// receives observations from the actor, serialises them, and fans them out
// to every connected browser. No package under internal/llm, internal/tools,
// or internal/jobs imports this one.
package gui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/waywardgeek/ensemble/agent/internal/common"
)

func newUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
}

// Server fans observations out to WebSocket clients. It implements
// common.Observer; its Observe method MUST NOT BLOCK.
//
// Reconnection uses the event log, not an observation buffer. The event log
// is append-only so readers take no lock (Log.Events elements are stable once
// written). The only mutable shared state is the in-flight partial map and
// the client set, which are guarded by mu.
type Server struct {
	mu           sync.Mutex
	clients      map[*Client]bool
	inflight     map[uint64][]byte // part_id → accumulated partial JSON
	logLen       int               // last-observed len(log.Events), updated under mu
	gate         *common.PauseGate
	send         func(common.Inbound) // forwards prompt/hint/interrupt to the actor
	model        func() string        // reports the model currently in force (see AgentHooks.Model)
	guiLog       *GuiLogger
	ttsLog       *TTSLogger                       // speech channel record; nil-safe when unset
	log          *common.Log                      // event log — read-only access for reconnection
	settings     common.SettingsSource            // GUI-editable settings; nil = no settings
	usage        common.UsageSource               // session token tally; nil = no meter
	mcpReceivers map[string]func(json.RawMessage) // source tag → JSON-RPC receiver (in-process agents)
	mcpAgents    map[string]*Client               // source tag → WebSocket client (remote agents like virtual user)
	selfReplies  *replyRouter                     // replies to the agent's own CallGUI requests
	selfSeq      atomic.Int64                     // CallGUI request ids
}

// AgentHooks bundles everything the GUI server needs from the agent side,
// supplied once at construction. One AgentHooks value is one agent. Every
// field is optional except as noted; a nil field disables the feature it
// feeds rather than crashing.
type AgentHooks struct {
	// Gate controls tool-dispatch pausing.
	Gate *common.PauseGate
	// Send is called for every prompt/hint/interrupt received from a browser.
	Send func(common.Inbound)
	// EventLog provides read access to the append-only event log for
	// reconnection replay.
	EventLog *common.Log
	// Settings provides the GUI-editable settings store.
	Settings common.SettingsSource
	// Usage provides the running session token tally for the status meter.
	Usage common.UsageSource
	// Model reports the model currently in force, for pricing. A closure
	// rather than a string because the operator can switch models mid-session,
	// and a value captured at construction would price every later turn at the
	// old rate while looking entirely correct.
	Model func() string
}

// New creates a GUI server bound to one agent's hooks. guiLogPath is the
// path of the GUI's presentation log (a GUI-side concern, so it travels
// outside AgentHooks).
func New(hooks AgentHooks, guiLogPath string) *Server {
	h := &Server{
		clients:      make(map[*Client]bool),
		inflight:     make(map[uint64][]byte),
		gate:         hooks.Gate,
		send:         hooks.Send,
		model:        hooks.Model,
		guiLog:       NewGuiLogger(guiLogPath),
		log:          hooks.EventLog,
		settings:     hooks.Settings,
		usage:        hooks.Usage,
		mcpReceivers: make(map[string]func(json.RawMessage)),
		mcpAgents:    make(map[string]*Client),
		selfReplies:  newReplyRouter(),
	}
	h.mcpReceivers[selfSource] = h.receiveSelf
	// logLen is the hub's notion of how much of the log is renderable. It was
	// previously advanced only by Observe, so on a restored session — where the
	// log is already full but no observation has arrived yet — it stayed at zero
	// and subscribe replayed nothing at all. Seed it from the log we are handed
	// so a connecting client sees the existing conversation immediately.
	if hooks.EventLog != nil {
		h.logLen = len(hooks.EventLog.Events)
	}
	return h
}

// Close closes the gui.log file.
func (h *Server) Close() {
	if h.guiLog != nil {
		h.guiLog.Close()
	}
	h.ttsLog.Close()
}

// SetTTSLog opens a speech log at path. Empty path disables it.
//
// Optional rather than a constructor parameter: most runs do not want one, and
// a nil logger is safe to call.
func (h *Server) SetTTSLog(path string) {
	h.ttsLog = NewTTSLogger(path)
}

// SetMCPReceiver registers a callback for incoming JSON-RPC messages tagged
// with the given source. The coding agent uses source "" (default); the
// virtual user uses source "vu". Called when wiring a WSTransport.
func (h *Server) SetMCPReceiver(source string, recv func(json.RawMessage)) {
	h.mu.Lock()
	h.mcpReceivers[source] = recv
	h.mu.Unlock()
}

// RemoveMCPReceiver unregisters the receiver for the given source tag.
func (h *Server) RemoveMCPReceiver(source string) {
	h.mu.Lock()
	delete(h.mcpReceivers, source)
	h.mu.Unlock()
}

// BroadcastJSONRPC sends a JSON-RPC message to all connected WebSocket
// clients, wrapped as {"type":"jsonrpc","source":"...","payload":{...}}.
// The source tag lets mcp.js echo it back so the hub can route the response.
// It returns how many browsers it reached; zero means nothing will answer.
func (h *Server) BroadcastJSONRPC(data json.RawMessage, source string) int {
	envelope := map[string]any{
		"type":    "jsonrpc",
		"payload": json.RawMessage(data),
	}
	if source != "" {
		envelope["source"] = source
	}
	msg, _ := json.Marshal(envelope)

	h.mu.Lock()
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		if c.live {
			clients = append(clients, c)
		}
	}
	h.mu.Unlock()

	for _, c := range clients {
		select {
		case c.send <- msg:
		default:
		}
	}
	return len(clients)
}

// Observe implements common.Observer. Must not block.
func (h *Server) Observe(obs common.Observation) {
	// A turn just finished, so the session tally has moved: refresh the
	// meter. This runs before the marshal below because that path returns
	// early for observations with no wire representation, and before h.mu
	// is taken because broadcastUsage acquires it itself.
	//
	// Usage is already recorded by the time this fires: the engine records
	// on ResponseEnded, and a turn's last response always ends before the
	// turn does.
	if _, ok := obs.(common.TurnEnded); ok {
		h.broadcastUsage()
	}

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
func (h *Server) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := newUpgrader().Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()

	go c.writePump()
	c.readPump() // blocks until disconnect

	h.removeClient(c)
}

// removeClient detaches a departing client: it leaves the client set, drops
// any MCP agent registrations it owned, and stops its writePump.
//
// Close done, NEVER c.send. Thirteen call sites send on a client's send
// channel - eleven spelled c.send, plus agentClient.send and cl.send - and
// the only one that may close a channel is the sole sender, which this is
// not. Closing it here raced every one of them: a sender snapshots the
// client under the lock, releases it, and sends after this function has
// already closed the channel, crashing the process with "send on closed
// channel" and taking the agent down mid-turn.
//
// The select/default guarding most of those sends looks like protection and
// is not: default saves you from a FULL channel, never from a closed one.
// Leaving c.send open costs nothing - an unreferenced channel is garbage
// collected whether or not it was ever closed - while writePump now stops on
// done instead of on the close.
func (h *Server) removeClient(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	// Clean up any MCP agent registrations for this client.
	for source, agent := range h.mcpAgents {
		if agent == c {
			delete(h.mcpAgents, source)
		}
	}
	h.mu.Unlock()
	close(c.done)
}

// subscribe reports the available event-log range, sends the full
// renderable window, and marks the client live. The client can
// optionally call fetch later for gap repair or older history.
// sendReplay delivers one replayed message, waiting for room in the client's
// buffer rather than dropping it. Live frames may be dropped safely because a
// newer one supersedes them, but a replayed frame is never re-sent: dropping it
// silently removes part of the conversation the user is trying to read. The
// timeout keeps a dead or wedged client from pinning this goroutine forever.
func sendReplay(c *Client, data []byte) bool {
	t := time.NewTimer(5 * time.Second)
	defer t.Stop()
	select {
	case c.send <- data:
		return true
	case <-c.done:
		// The client left. Once writePump has stopped nothing drains c.send,
		// so without this case every remaining replay write would burn the
		// full timeout before giving up on a socket that is already gone.
		return false
	case <-t.C:
		return false
	}
}

func (h *Server) subscribe(c *Client) {
	h.mu.Lock()
	logLen := h.logLen
	h.mu.Unlock()

	events := h.log.Events
	if logLen > len(events) {
		logLen = len(events)
	}

	// Report the available range so the client knows what exists.
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

	// Send the renderable event-log window. Unlike a live frame, a dropped
	// replay message is never re-sent, so it leaves a permanent hole in the
	// history the user sees. A long conversation renders to far more messages
	// than the 256 slots c.send holds, so wait for room instead of dropping.
	// The timeout bounds the wait so a dead client cannot pin this goroutine.
replay:
	for i := 0; i < logLen; i++ {
		e := events[i]
		if msgs := renderEvent(e); len(msgs) > 0 {
			for _, data := range msgs {
				if !sendReplay(c, data) {
					break replay
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
	c.lastSeq = lastSeq
	c.live = true
	h.mu.Unlock()

	// Send current settings and the model catalog so the client has the full state.
	if h.settings != nil {
		s := h.settingsAtLoad()
		data, _ := json.Marshal(map[string]any{
			"type":     "current_settings",
			"settings": s,
			"models":   common.ModelListJSON(),
		})
		select {
		case c.send <- data:
		default:
		}
		h.guiLog.Log("<", data)
	}

	// Send the session meter so a freshly loaded page shows the running
	// total rather than zeros until the next turn happens to end.
	if data := h.usageFrame(); data != nil {
		select {
		case c.send <- data:
		default:
		}
		h.guiLog.Log("<", data)
	}
}

// usageFrame builds the session meter frame, or returns nil if there is no
// usage source to read.
//
// Cost is computed here rather than in the browser so that the price table has
// exactly one home. The event log deliberately stores counts and never money
// (prices change; counts are history), but this frame is a live display and
// dollars are what the reader wants.
//
// priced is carried separately because an unpriced model and a model that has
// spent nothing both cost zero, and the GUI must be able to tell them apart:
// one renders a dash, the other "$0.00".
// effectiveModel names the model whose prices apply to this session.
//
// This reports the model the engine will actually dial, not the one most
// recently picked in the GUI. The two used to differ, and silently: the GUI
// wrote its choice to the settings store, nothing ever applied that to the
// engine, and this readout consulted the store first. So the meter cheerfully
// confirmed a model that every request ignored, which is a worse failure than
// showing nothing at all.
//
// Now that a model switch is applied to the engine for real, Cfg.Model is the
// single source of truth and the readout cannot drift from the wire. The
// settings store remains only as a fallback for a hub wired without an
// engine, as in tests.
// settingsAtLoad returns the settings a freshly connected client should see.
//
// Nothing is pending on a fresh load, so the model shown is the model
// actually in force, not the stored one. Startup never applies a stored model
// to the engine: only ContextTarget, Bands and LogRetention are wired from
// settings. A model left in settings.json by an earlier session would
// otherwise be displayed while the agent runs the startup default, which is
// exactly the drift effectiveModel exists to stop.
func (h *Server) settingsAtLoad() common.Settings {
	s := h.settings.Get()
	s.Model = h.effectiveModel()
	return s
}

// settingsForDisplay returns settings as the operator should SEE them.
//
// An unset Model means "follow the startup default". That is the right thing
// to STORE and the wrong thing to SHOW: a picker handed an empty string falls
// back to its first option and then reports a model the agent is not running.
// That is the failure the usage meter already had, and showing the wrong
// model is worse than showing none.
//
// It fills an EMPTY model only. After the operator picks a model the actor
// applies the switch asynchronously, so the live model can still be the old
// one for a moment; overwriting unconditionally would snap the picker back to
// the previous model right after a change.
func (h *Server) settingsForDisplay(s common.Settings) common.Settings {
	if s.Model == "" {
		s.Model = h.effectiveModel()
	}
	return s
}

func (h *Server) effectiveModel() string {
	if h.model != nil {
		if m := h.model(); m != "" {
			return m
		}
	}
	if h.settings != nil {
		return h.settings.Get().Model
	}
	return ""
}

func (h *Server) usageFrame() []byte {
	if h.usage == nil {
		return nil
	}
	session := h.usage.SessionUsage()
	last := h.usage.LastUsage()

	var (
		priced bool
		cost   float64
	)
	// Price each model's own counts at that model's own rates.
	//
	// This used to be CostUSD(session, priceOfCurrentModel), which billed
	// every token the session had ever spent at whatever model happened to
	// be selected right now. Switching from a cheap model to an expensive
	// one retroactively re-priced the entire history, so the displayed cost
	// jumped without a request having been sent -- a wrong number that looks
	// completely plausible, which is the kind that survives.
	cost, priced = common.CostByModel(h.usage.UsageByModel())

	data, _ := json.Marshal(map[string]any{
		"type": "usage",
		// Per-response: the operator sees how large the current prompt is
		// and how much of it was cached, not a running sum that only grows.
		"input":       last.Input,
		"cache_write": last.CacheWrite,
		"cache_read":  last.CacheRead,
		"output":      last.Output,
		// Session-wide: cost and hit rate accumulate because that is what
		// a developer means by "what has this session spent."
		"hit_rate": common.CacheHitRate(session),
		"cost_usd": cost,
		"priced":   priced,
		// Session totals for the tooltip, so the operator can still see
		// the running sum on hover.
		"session_input":       session.Input,
		"session_cache_write": session.CacheWrite,
		"session_cache_read":  session.CacheRead,
		"session_output":      session.Output,
	})
	return data
}

// broadcastUsage pushes the session meter to every live client.
func (h *Server) broadcastUsage() {
	data := h.usageFrame()
	if data == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if !c.live {
			continue
		}
		select {
		case c.send <- data:
		default:
		}
	}
}

// fetch sends event-log entries in the requested seq range, followed by
// any in-flight partials. This is an explicit range request, not part of
// initial subscription: subscribe already pushes the complete history.
func (h *Server) fetch(c *Client, fromSeq, toSeq common.Seq) {
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
func (h *Server) handleClientMessage(c *Client, raw []byte) {
	h.guiLog.Log(">", raw)

	var msg struct {
		Type     string          `json:"type"`
		Text     string          `json:"text"`
		From     common.Seq      `json:"from"`
		To       common.Seq      `json:"to"`
		Settings json.RawMessage `json:"settings"`
		Payload  json.RawMessage `json:"payload"`  // JSON-RPC payload
		Source   string          `json:"source"`   // MCP source tag ("vu", "", etc.)
		TTS      *TTSEvent       `json:"tts"`      // speech channel record
		Typing   bool            `json:"typing"`   // gate: a person is typing
		Speaking bool            `json:"speaking"` // gate: speech is playing
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
			h.send(common.UserMessage{Text: msg.Text})
		}
	case "hint":
		if msg.Text != "" {
			h.send(common.Hint{Text: msg.Text})
		}
	case "interrupt":
		h.send(common.Interrupt{})

	case "reset":
		h.send(common.Reset{})
	case "pause":
		h.ttsLog.Log(TTSEvent{Kind: "pause", Cause: gateCause(msg.Typing, msg.Speaking)})
		if h.gate != nil {
			h.gate.Pause()
		}
	case "unpause":
		h.ttsLog.Log(TTSEvent{Kind: "resume"})
		if h.gate != nil {
			h.gate.Unpause()
		}
	// The browser reports what it handed to the speech channel. Only the client
	// knows this: the decision about what is speakable is made there, so the
	// server cannot derive these lines, only be told them.
	case "tts":
		if msg.TTS != nil {
			ev := *msg.TTS
			if ev.Kind == "" {
				ev.Kind = "utterance"
			}
			h.ttsLog.Log(ev)
		}
	case "update_settings":
		if h.settings != nil && msg.Settings != nil {
			was := h.effectiveModel()
			updated := h.settings.ApplyRaw(msg.Settings)
			h.broadcastSettings(updated)

			// A model switch is applied by the actor, between turns. The
			// actor owns the config, so changing it from this goroutine
			// could land midway through rendering a request. Compared
			// against the model the engine is actually dialing, not against
			// the previous setting, because the setting starts empty when
			// the model came from the environment.
			if h.send != nil && updated.Model != "" && updated.Model != was {
				h.send(common.SetModel{Model: updated.Model})
			}
		}
	case "jsonrpc":
		// JSON-RPC routing between agents and the browser's MCP server.
		//
		// Three kinds of agents can send/receive jsonrpc:
		// 1. In-process agents (coding agent): registered via SetMCPReceiver
		// 2. Remote agents (virtual user): WebSocket clients, tracked in mcpAgents
		// 3. Browser clients: host the MCP server (gui_snapshot, gui_click, etc.)
		//
		// Routing:
		// - If a receiver exists for the source tag → deliver to it (browser → in-process agent)
		// - If a mcpAgent exists for the source tag → deliver to it (browser → remote agent)
		// - Otherwise this is an agent sending a request → forward to all browser clients
		h.mu.Lock()
		recv := h.mcpReceivers[msg.Source]
		agentClient := h.mcpAgents[msg.Source]
		h.mu.Unlock()

		if msg.Payload == nil {
			return
		}

		if recv != nil {
			recv(msg.Payload)
		} else if agentClient != nil && agentClient != c {
			select {
			case agentClient.send <- raw:
			default:
			}
		} else {
			// Agent request → forward to browser clients.
			if msg.Source != "" {
				h.mu.Lock()
				h.mcpAgents[msg.Source] = c
				h.mu.Unlock()
			}
			h.mu.Lock()
			for cl := range h.clients {
				if cl.live && cl != c {
					select {
					case cl.send <- raw:
					default:
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// broadcastSettings sends a settings_changed message to all live clients.
func (h *Server) broadcastSettings(s common.Settings) {
	s = h.settingsForDisplay(s)
	data, err := json.Marshal(map[string]any{
		"type":     "settings_changed",
		"settings": s,
	})
	if err != nil {
		return
	}
	h.guiLog.Log("<", data)

	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if c.live {
			select {
			case c.send <- data:
			default:
			}
		}
	}
}

// ----------------------------------------------------------------
// Client
// ----------------------------------------------------------------

// Client represents a single WebSocket connection.
type Client struct {
	hub     *Server
	conn    *websocket.Conn
	send    chan []byte
	done    chan struct{} // closed once, by ServeWS teardown, to stop writePump
	live    bool          // receives live messages only after subscribe
	lastSeq common.Seq    // last event-log Seq delivered to this client
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
//
// It stops on done rather than on c.send being closed, because nothing closes
// c.send - see removeClient. Draining whatever is already buffered first
// keeps the old behavior that a client gets the messages queued before it
// left.
func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		select {
		case msg := <-c.send:
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-c.done:
			// Drain anything already queued, then stop.
			for {
				select {
				case msg := <-c.send:
					if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
						return
					}
				default:
					return
				}
			}
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
		common.MessageReceived, common.ErrorOccurred, common.ConversationReset:
		return true
	}
	return false
}

// renderEvent converts an event-log entry into zero or more wire messages.
// Each returned []byte is a complete JSON object ready for the WebSocket.
func renderEvent(e common.Event) [][]byte {
	var msgs [][]byte

	switch e.Type {
	case common.ConversationReset:
		// The client clears its transcript on this and keeps everything else.
		// It arrives through the ordinary event stream rather than as a reply
		// to the button press, so a reconnecting client replaying the log
		// rebuilds the same cleared screen instead of a stale one.
		data, _ := json.Marshal(map[string]any{"type": "conversation_reset", "seq": e.Seq})
		msgs = append(msgs, data)

	case common.ResponseEnded:
		if e.Response == nil {
			break
		}
		for i, p := range e.Response.Parts {
			m := partToWireMsg(e.Seq, i, p)
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
//
// partIndex differentiates parts within one ResponseEnded event.
// Without it, every replayed part_final shares part_id undefined and
// the GUI's artifact Map collapses all agent responses into a single
// DOM element — the root cause of "assistant text vanishes on restart".
func partToWireMsg(seq common.Seq, partIndex int, p common.Part) map[string]any {
	partID := fmt.Sprintf("r%d.%d", seq, partIndex)
	switch v := p.(type) {
	case common.TextPart:
		return map[string]any{
			"type":    "part_final",
			"seq":     seq,
			"part_id": partID,
			"kind":    "text",
			"text":    v.Text,
		}
	case common.ToolCallPart:
		return map[string]any{
			"type":    "part_final",
			"seq":     seq,
			"part_id": partID,
			"kind":    "tool_call",
			"tool":    v.Name,
			"call_id": v.CallID,
			"args":    string(v.Args),
		}
	case common.OpaquePart:
		// Thinking blocks are stored as OpaquePart with a "thinking" type
		// field and the thinking text in the JSON data. Extract the text
		// so the GUI can display it on reconnect — without this, thinking
		// is visible during live streaming but vanishes after restart.
		var opaque struct {
			Type     string `json:"type"`
			Thinking string `json:"thinking"`
		}
		if json.Unmarshal(v.Data, &opaque) == nil && opaque.Type == "thinking" && opaque.Thinking != "" {
			return map[string]any{
				"type":    "part_final",
				"seq":     seq,
				"part_id": partID,
				"kind":    "thinking",
				"text":    opaque.Thinking,
			}
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
	case common.ConversationCleared:
		// Must match the frame renderEvent produces for ConversationReset so
		// that a live clear and a replayed one drive the client down the same
		// path. The client does not track per-event seq, so none is sent.
		m = map[string]any{"type": "conversation_reset"}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
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
