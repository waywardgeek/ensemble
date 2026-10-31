// Package gui is an optional HTTP/WebSocket client of the public Ensemble API.
package gui

import (
	"context"
	"encoding/json"
	"ensemble/ensemble"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Server owns browser connections; its supplied Agent remains the conversation
// owner. Construct the server before starting the Agent, and close it afterward.
type Server struct {
	agent    ensemble.Agent
	settings Settings
	// settingsMu orders patches and subscriptions without blocking actor observations.
	settingsMu       sync.Mutex
	mu               sync.Mutex
	clients          map[*websocket.Conn]chan [][]byte
	partials         map[uint64]map[string]any
	events           []ensemble.Event
	seq              uint64
	logMu            sync.Mutex
	log              io.Writer
	closed           bool
	connections      map[*websocket.Conn]bool
	tunnels          map[*websocket.Conn]*WSTransport
	workers          sync.WaitGroup
	relays           map[net.Listener]bool
	relayConnections map[net.Conn]bool
}

// New attaches a root observer. Ensemble and Agent are application services,
// not parents of this optional transport; the server owns its connections and
// browser preferences. Invalid saved settings fail construction before attachment.
func New(root *ensemble.Ensemble, agent ensemble.Agent, log io.Writer) (*Server, error) {
	if root == nil || agent == nil || agent.Ensemble() != root || log == nil {
		panic("GUI requires its Ensemble, Agent and log")
	}
	s := &Server{agent: agent, clients: make(map[*websocket.Conn]chan [][]byte), partials: make(map[uint64]map[string]any), log: log}
	s.connections = make(map[*websocket.Conn]bool)
	s.tunnels = make(map[*websocket.Conn]*WSTransport)
	s.relays = make(map[net.Listener]bool)
	s.relayConnections = make(map[net.Conn]bool)
	if err := s.loadSettings(); err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}
	root.AddObserver(s)
	return s, nil
}

// Submit retains the original external client seam over the same Agent.
func Submit(ctx context.Context, agent ensemble.Agent, input string) (string, error) {
	return agent.Ask(ctx, input)
}

// ObserveAgent makes the history frontier and partials one synchronized view.
// Subscribe takes this same lock: a final can never overtake its old partial.
func (s *Server) ObserveAgent(a ensemble.Agent, observation ensemble.Observation) {
	if a != s.agent {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	// Advance durable facts before transient delivery. In particular, a tool
	// start must bring its already-recorded arguments along with it.
	s.capture()
	switch o := observation.(type) {
	case ensemble.BillingSelected:
		s.broadcast(map[string]any{"type": "billing_selected", "source": o.Source})
	case ensemble.PartDelta:
		// A reconnect needs the accumulated prefix; live subscribers need only the
		// new fragment. Sending the accumulated value live would duplicate words.
		if o.Hint {
			return
		} // The durable hint was already published by capture.
		previous := ""
		if p := s.partials[o.PartID]; p != nil {
			previous = p["chunk"].(string)
		}
		s.partials[o.PartID] = map[string]any{"type": "part_delta", "part_id": o.PartID, "kind": o.Kind, "chunk": previous + o.Chunk}
		s.broadcast(map[string]any{"type": "part_delta", "part_id": o.PartID, "kind": o.Kind, "chunk": o.Chunk})
	case ensemble.StateChanged:
		s.broadcast(map[string]any{"type": "state_changed", "from": o.From, "to": o.To})
	case ensemble.TurnEnded:
		s.broadcast(s.usageWire())
		// Cancellation may leave a part without a final. Those abandoned bytes
		// must not reappear as an active stream on the next subscription.
		s.partials = make(map[uint64]map[string]any)
		s.broadcast(map[string]any{"type": "turn_ended", "text": o.Text, "error": o.Error})
	}
}

// New durable data always comes from history. The GUI retains a renderable view
// of that log, never reconstructs facts by trusting transient progress chunks.
func (s *Server) capture() {
	for _, event := range s.agent.History().EventsAfter(s.seq) {
		s.seq = event.Seq
		s.events = append(s.events, event)
		if event.Type == "response_ended" {
			s.broadcast(s.usageWire())
			for _, p := range event.Response.Parts {
				delete(s.partials, p.ArtifactID)
			}
		}
		for _, message := range eventWire(event) {
			s.broadcast(message)
		}
	}
}

// Live delivery and reconnect use the same projection. The browser receives
// display-ready facts and never has to understand provider frames or reducers.
func eventWire(e ensemble.Event) []map[string]any {
	var out []map[string]any
	switch e.Type {
	case "message_received", "hint_received":
		// System ephemera are instructions to the model, not human chat cards.
		// Hints are visible because classification must not hide the user input.
		if e.Message.Actor == ensemble.Human {
			for i, p := range e.Message.Parts {
				out = append(out, map[string]any{"type": "user_message", "part_id": fmt.Sprintf("user-%d-%d", e.Seq, i), "part": p})
			}
		}
	case "response_ended":
		// Artifact IDs were captured by the actor before storage. Replay and live
		// finals therefore address exactly the cards created by streamed deltas.
		for _, p := range e.Response.Parts {
			out = append(out, map[string]any{"type": "part_final", "part_id": p.ArtifactID, "part": p})
		}
	case "tool_called":
		out = append(out, map[string]any{"type": "tool_dispatched", "call_id": e.Tool.CallID, "name": e.Tool.Name, "input": e.Tool.Args})
	case "tool_returned", "job_report":
		// Late real results survive interruption. Displaying them does not restart
		// the ended conversation or turn a continuing job into a failed one.
		var text strings.Builder
		for _, p := range e.Tool.Parts {
			text.WriteString(p.Text)
		}
		out = append(out, map[string]any{"type": "tool_finished", "call_id": e.Tool.CallID, "result": text.String(), "is_error": e.Tool.IsError})
	}
	return out
}

// Only callers holding mu may enumerate receivers. Subscription and teardown
// share that lock, but network writes and transport logging never hold it.
func (s *Server) broadcast(message map[string]any) {
	data, _ := json.Marshal(message)
	for _, queue := range s.clients {
		select {
		case queue <- [][]byte{data}:
		default:
		} // Slow radios cannot stop the station.
	}
}
func (s *Server) trace(direction string, data []byte) {
	s.logMu.Lock()
	defer s.logMu.Unlock()
	// Compact incoming payloads too: every message is exactly one trace line.
	var value any
	if json.Unmarshal(data, &value) == nil {
		data, _ = json.Marshal(value)
	} else {
		data, _ = json.Marshal(string(data))
	}
	fmt.Fprintf(s.log, "%s %s %s\n", time.Now().UTC().Format(time.RFC3339Nano), direction, data)
}

// Handler serves embedded assets by default; a nonempty directory explicitly
// selects editable disk assets. HTTP registration remains local to this Server.
func (s *Server) Handler(directory string) http.Handler {
	mux := http.NewServeMux()
	assets := http.FS(Assets())
	if directory != "" {
		assets = http.Dir(directory)
	}
	mux.Handle("/", http.FileServer(assets))
	mux.HandleFunc("/ws", s.socket)
	return mux
}
func (s *Server) socket(w http.ResponseWriter, r *http.Request) {
	// Gorilla's default browser origin policy requires the same HTTP host.
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		conn.Close()
		return
	}
	s.connections[conn] = true
	s.workers.Add(1)
	s.mu.Unlock()
	// The queue carries batches so replay has one admission point even when
	// its history window exceeds the capacity chosen for live progress.
	queue := make(chan [][]byte, 256)
	tunnel := newWSTransport(s, queue)
	s.mu.Lock()
	s.tunnels[conn] = tunnel
	s.mu.Unlock()
	defer func() { tunnel.Close(); close(tunnel.relayDone) }()
	var mcpStarted bool
	// Separate lifetime signals keep teardown safe without closing a channel
	// that the broadcaster can still reach. The writer is joined before return.
	done, written := make(chan struct{}), make(chan struct{})
	defer func() {
		conn.Close()
		close(done)
		<-written
		s.mu.Lock()
		delete(s.clients, conn)
		delete(s.connections, conn)
		delete(s.tunnels, conn)
		s.mu.Unlock()
		s.workers.Done()
	}()
	// Gorilla permits one reader and one writer per connection. All outbound
	// messages, including replay, follow this single writer in queue order.
	go func() {
		defer close(written)
		for {
			select {
			case batch := <-queue:
				for _, data := range batch {
					if conn.WriteMessage(websocket.TextMessage, data) != nil {
						conn.Close()
						return
					}
					s.trace("<", data)
				}
			case <-done:
				return
			}
		}
	}()
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		s.trace(">", data)
		var input struct {
			// Type selects the client command; Text carries its human input.
			Type, Text string
			// Settings carries a sparse preference patch; absent fields stay unchanged.
			Settings json.RawMessage
			// Payload carries an opaque JSON-RPC object for this connection.
			Payload json.RawMessage
		}
		if json.Unmarshal(data, &input) != nil {
			continue
		}
		switch input.Type {
		case "jsonrpc":
			if tunnel.relayReply(input.Payload) {
				continue
			}
			select {
			case tunnel.incoming <- input.Payload:
			case <-tunnel.done:
			}
		case "subscribe":
			s.settingsMu.Lock()
			s.mu.Lock()
			if s.closed {
				s.mu.Unlock()
				s.settingsMu.Unlock()
				return
			}
			s.capture()
			if _, exists := s.clients[conn]; !exists {
				// One queued batch preserves the entire replay, even when it exceeds the
				// live progress buffer. Socket writes remain outside the observer lock.
				queue <- s.replay()
				s.clients[conn] = queue
			}
			s.mu.Unlock()
			s.settingsMu.Unlock()
			if !mcpStarted {
				mcpStarted = true
				go func() {
					if err := s.agent.MCP().AttachBrowser(context.Background(), tunnel); err != nil {
						s.agent.Ensemble().Logf("browser MCP: %v", err)
					}
				}()
			}
		case "update_settings":
			// Settings retain one order across tabs, but the actor must be free
			// to publish progress while this call waits for its acknowledgment.
			s.settingsMu.Lock()
			settings, err := s.updateSettings(input.Settings)
			s.mu.Lock()
			if err != nil {
				data, _ := json.Marshal(map[string]any{"type": "settings_error", "error": err.Error()})
				select {
				case queue <- [][]byte{data}:
				default:
					conn.Close()
				}
			} else {
				s.broadcast(map[string]any{"type": "settings_changed", "settings": settings})
			}
			s.mu.Unlock()
			s.settingsMu.Unlock()
		case "prompt":
			// Live input uses the same actor door as terminal input. It may become
			// a hint; the browser does not guess whether the Agent is currently busy.
			s.agent.Post(ensemble.UserMessage{Text: input.Text})
		case "hint":
			s.agent.Post(ensemble.Hint{Text: input.Text})
		case "interrupt":
			s.agent.Post(ensemble.Interrupt{})
		case "pause", "unpause":
			// These commands affect the Agent gate directly, ahead of queued tool
			// completions. No connection owns a share or veto of that global state.
			if input.Type == "pause" {
				s.agent.Pause()
			} else {
				s.agent.Unpause()
			}
			s.mu.Lock()
			s.broadcast(map[string]any{"type": "pause_changed", "paused": s.agent.Paused()})
			s.mu.Unlock()
		}
	}
}

// The fixed window retains the last hour OR 100 renderable events, whichever
// is larger. Configuration and restoring an Agent from disk come later.
func (s *Server) replay() [][]byte {
	cutoff, count, start := time.Now().Add(-time.Hour), 0, len(s.events)
	for start > 0 {
		event := s.events[start-1]
		if len(eventWire(event)) > 0 {
			count++
		}
		if count > 100 && event.Time.Before(cutoff) {
			break
		}
		start--
	}
	var out [][]byte
	add := func(message map[string]any) {
		message["replay"] = true
		data, _ := json.Marshal(message)
		out = append(out, data)
	}
	for _, event := range s.events[start:] {
		for _, m := range eventWire(event) {
			add(m)
		}
	}
	// Maps do not preserve stream order. Increasing actor-assigned identities
	// restore creation order when several partial artifacts are active.
	ids := make([]uint64, 0, len(s.partials))
	for id := range s.partials {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		add(s.partials[id])
	}
	// A fresh receiver needs effective state even if no transition occurs
	// after it subscribes. These status snapshots are silent replay messages.
	// Replay reports the current state, not a historical transition. Equal
	// endpoints retain the public envelope without inventing a previous state.
	state := s.agent.History().Context().Turn
	add(map[string]any{"type": "state_changed", "from": state, "to": state})
	add(map[string]any{"type": "pause_changed", "paused": s.agent.Paused()})
	add(map[string]any{"type": "current_settings", "settings": s.snapshotSettings()})
	add(s.usageWire())
	return out
}

// Close makes this observer inert and disconnects receivers. Data channels are
// never closed, so concurrent fan-out cannot panic on a closed channel.
func (s *Server) Close() {
	s.mu.Lock()
	s.closed = true
	for listener := range s.relays {
		listener.Close()
	}
	for conn := range s.relayConnections {
		conn.Close()
	}
	for conn := range s.connections {
		conn.Close()
	}
	s.mu.Unlock()
	// Join writers before the application closes gui.log.
	s.workers.Wait()
}

// Agent reaches the conversation served by this server's transport children.
func (s *Server) Agent() ensemble.Agent { return s.agent }

// usageWire reads Engine's session counter, never the replayed history total.
func (s *Server) usageWire() map[string]any {
	u := s.agent.Engine().Meter()
	return map[string]any{"type": "usage", "input": u.Input, "cache_write": u.CacheWrite, "cache_read": u.CacheRead, "output": u.Output, "total": u.Total, "cost_usd": u.CostUSD, "priced": u.Priced, "hit_rate": u.HitRate}
}
