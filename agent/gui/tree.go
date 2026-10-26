package gui

import (
	"encoding/json"
	"ensemble/ensemble"
	"fmt"
	"github.com/gorilla/websocket"
)

// A conversation view belongs to the transport server; the core Agent still
// owns all history. Cursors and partial IDs never cross an Agent boundary.
type agentView struct {
	parent   *Server
	agent    ensemble.Agent
	settings Settings
	partials map[uint64]map[string]any
	events   []ensemble.Event
	seq      uint64
	loaded   bool
}

func (s *Server) view(a ensemble.Agent) *agentView {
	if v := s.views[a]; v != nil {
		return v
	}
	v := &agentView{parent: s, agent: a, partials: make(map[uint64]map[string]any), settings: Settings{Theme: "system", TTSEnabled: true, TTSSpeed: 1, FontSize: 16}}
	s.views[a] = v
	return v
}
func (s *Server) selection(conn *websocket.Conn) *agentView {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v := s.selected[conn]; v != nil {
		return v
	}
	return s.view(s.agent)
}
func wire(message map[string]any) []byte { data, _ := json.Marshal(message); return data }
func (s *Server) sendError(queue chan [][]byte, err error) {
	select {
	case queue <- [][]byte{wire(map[string]any{"type": "agent_error", "error": err.Error()})}:
	default:
	}
}

// Conversation fanout is selected per connection. Topology updates still reach
// every subscriber, including tabs reading another Agent's retained history.
func (v *agentView) broadcast(message map[string]any) {
	message["agent_key"] = ensemble.AgentKey(v.agent)
	data := wire(message)
	for conn, queue := range v.parent.clients {
		if v.parent.selected[conn] != v {
			continue
		}
		select {
		case queue <- [][]byte{data}:
		default:
		}
	}
}

// Tree reads can traverse several managers. Coalesce refresh requests outside
// the synchronous actor observer so registry or child locks never reverse into
// an Agent that is waiting for that observer to return.
func (s *Server) treeUpdates() {
	defer close(s.treeStopped)
	for {
		select {
		case <-s.treeWake:
			// Selection can wait on settings/actor acknowledgment. Keep it away
			// from synchronous observations and retain no dismissed-ID registry.
			s.mu.Lock()
			stale := make(map[*websocket.Conn]*agentView)
			for conn, v := range s.selected {
				if v.agent.ParentAgent() != nil && !s.root.ContainsAgent(v.agent) {
					stale[conn] = v
				}
			}
			s.mu.Unlock()
			for conn, v := range stale {
				s.mu.Lock()
				queue := s.clients[conn]
				s.mu.Unlock()
				if err := s.selectConversation(conn, queue, v.agent.ParentAgent(), true, v); err != nil {
					s.sendError(queue, err)
					conn.Close()
				}
			}
			tree := s.root.TreeSnapshot(s.agent)
			s.mu.Lock()
			if !s.closed {
				s.broadcast(map[string]any{"type": "agent_tree", "root": tree})
			}
			s.mu.Unlock()
		case <-s.treeDone:
			return
		}
	}
}

// Explicit selection and removal fallback share one ordered settings/replay path.
// expected is only for deferred fallback: a user's newer selection wins over it.
func (s *Server) selectConversation(conn *websocket.Conn, queue chan [][]byte, selected ensemble.Agent, force bool, expected *agentView) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	for {
		for selected.ParentAgent() != nil && !s.root.ContainsAgent(selected) {
			selected = selected.ParentAgent()
		}
		s.mu.Lock()
		if s.closed || !s.connections[conn] {
			s.mu.Unlock()
			return fmt.Errorf("connection closed")
		}
		if expected != nil && s.selected[conn] != expected {
			s.mu.Unlock()
			return nil
		}
		// Removal uses this same view lock. Recheck after acquiring it so a
		// selection delayed at the lock cannot recreate an already-deleted view.
		if selected.ParentAgent() != nil && !s.root.ContainsAgent(selected) {
			s.mu.Unlock()
			continue
		}
		v := s.view(selected)
		s.mu.Unlock()
		if !v.loaded {
			if err := v.loadSettings(); err != nil {
				if selected.ParentAgent() != nil && !s.root.ContainsAgent(selected) {
					continue
				}
				return err
			}
			v.loaded = true
		}
		tree := s.root.TreeSnapshot(s.agent)
		s.mu.Lock()
		if selected.ParentAgent() != nil && !s.root.ContainsAgent(selected) {
			delete(s.views, selected)
			s.mu.Unlock()
			continue
		}
		if s.closed || !s.connections[conn] {
			s.mu.Unlock()
			return fmt.Errorf("connection closed")
		}
		v.capture()
		if _, exists := s.clients[conn]; !exists || force {
			batch := [][]byte{wire(map[string]any{"type": "agent_selected", "key": ensemble.AgentKey(selected)}), wire(map[string]any{"type": "agent_tree", "root": tree})}
			batch = append(batch, v.replay()...)
			// Replay is indivisible. A stalled receiver disconnects rather than
			// blocking all actor observers or silently losing half its history.
			select {
			case queue <- batch:
			default:
				s.mu.Unlock()
				conn.Close()
				return fmt.Errorf("replay receiver is full")
			}
			s.clients[conn] = queue
			s.selected[conn] = v
		}
		s.mu.Unlock()
		return nil
	}
}
