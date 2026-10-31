package gui

import (
	"encoding/json"
	"ensemble/ensemble"
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
	v := &agentView{parent: s, agent: a, partials: make(map[uint64]map[string]any), settings: Settings{Theme: "dark", TTSEnabled: true, TTSSpeed: 1, FontSize: 16}}
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
