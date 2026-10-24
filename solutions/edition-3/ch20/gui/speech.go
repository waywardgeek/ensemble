package gui

import (
	"ensemble"
	"errors"
	"fmt"
	"strings"
	"time"
)

// The hub lock owns speech admission and publication. At most one active task
// and one coalesced text buffer exist per Agent; accepted text is never dropped
// to satisfy a task-count limit. Only the claimed browser speaks a task.
type spokenPart struct {
	Text   string `json:"text"`
	Source string `json:"source"`
	PartID uint64 `json:"part_id,omitempty"`
	TurnID uint64 `json:"turn_id,omitempty"`
}
type speechTask struct {
	segments                       []spokenPart
	id, agent, text, owner, reason string
	epoch                          uint64
	turn                           uint64
	timer                          *time.Timer
}
type speaker struct {
	hub     *Hub
	next    uint64
	tasks   map[string]*speechTask
	active  map[string]string
	pending map[string]*speechTask
	closed  bool
}

func newSpeaker(h *Hub) *speaker {
	return &speaker{hub: h, tasks: map[string]*speechTask{}, active: map[string]string{}, pending: map[string]*speechTask{}}
}
func (s *speaker) request(agent string, o ensemble.SpeechRequested) {
	text := cleanSpeech(readable(o.Parts))
	s.enqueue(agent, spokenPart{Text: text, Source: "system"}, o.Reason, o.TurnID)
}
func (s *speaker) enqueue(agent string, part spokenPart, reason string, turn uint64) {
	if strings.TrimSpace(part.Text) == "" || s.closed {
		return
	}
	part.TurnID = turn
	epoch := s.hub.agents[agent].state.Epoch
	if s.active[agent] != "" {
		if t := s.pending[agent]; t != nil {
			// Coalesce scheduling, not semantics. Each segment keeps its own origin.
			t.text += "\n\n" + part.Text
			t.segments = append(t.segments, part)
			if t.reason != reason {
				t.reason = "automatic"
			}
			return
		}
	}
	s.next++
	t := &speechTask{id: fmt.Sprint(s.next), agent: agent, text: part.Text, segments: []spokenPart{part}, reason: reason, epoch: epoch, turn: turn}
	if s.active[agent] != "" {
		s.pending[agent] = t
		return
	}
	s.activate(t)
}
func (s *speaker) activate(t *speechTask) {
	s.tasks[t.id] = t
	s.active[t.agent] = t.id
	s.hub.broadcast(Wire{Type: "speech_requested", AgentID: t.agent, Data: map[string]any{"id": t.id, "native": false, "reason": t.reason, "epoch": t.epoch, "turn_id": t.turn}})
	s.hub.trace("speech_queued", t.agent, t.id)
	if t.owner != "" {
		s.hub.broadcast(Wire{Type: "speech_claimed", AgentID: t.agent, Data: map[string]any{"id": t.id, "owner": t.owner, "text": t.text, "segments": t.segments, "reason": t.reason, "epoch": t.epoch, "turn_id": t.turn}})
		return
	}
	t.timer = time.AfterFunc(30*time.Second, func() {
		s.hub.mu.Lock()
		defer s.hub.mu.Unlock()
		if s.tasks[t.id] == t && t.owner == "" {
			s.finish(t, "unclaimed")
		}
	})
}
func (s *speaker) claim(c *client, agent, id string) error {
	s.hub.mu.Lock()
	defer s.hub.mu.Unlock()
	if a, ok := s.hub.en.Agent(agent); ok {
		s.hub.sync(s.hub.view(a), a.View())
	}
	t := s.tasks[id]
	if t == nil || t.agent != agent {
		return errors.New("speech request expired")
	}
	if t.owner != "" {
		return nil
	}
	select {
	case <-c.done:
		return errors.New("speaker disconnected")
	default:
	}
	t.owner = c.id
	if t.timer != nil {
		t.timer.Stop()
	}
	// Claim validity and its publication share one critical section: a cancelled
	// request can never emit a later claim behind its cancellation notification.
	s.hub.broadcast(Wire{Type: "speech_claimed", AgentID: agent, Data: map[string]any{"id": id, "owner": c.id, "text": t.text, "segments": t.segments, "reason": t.reason, "epoch": t.epoch, "turn_id": t.turn}})
	return nil
}
func (s *speaker) complete(c *client, agent, id, status string) error {
	s.hub.mu.Lock()
	defer s.hub.mu.Unlock()
	t := s.tasks[id]
	if t == nil {
		return nil
	}
	if t.agent != agent || t.owner != c.id {
		return errors.New("not speech owner")
	}
	s.finish(t, status)
	return nil
}
func (s *speaker) finish(t *speechTask, status string) {
	if s.tasks[t.id] != t {
		return
	}
	delete(s.tasks, t.id)
	delete(s.active, t.agent)
	if t.timer != nil {
		t.timer.Stop()
	}
	s.hub.trace("speech_"+status, t.agent, t.id)
	s.hub.broadcast(Wire{Type: "speech_done", AgentID: t.agent, Data: map[string]any{"id": t.id, "status": status, "epoch": t.epoch, "turn_id": t.turn}})
	if status == "cancelled" {
		// Pending scheduling can span turns. Cancel this task's turns, not a
		// different Agent or text admitted for a later turn.
		turns := map[uint64]bool{t.turn: true}
		for _, part := range t.segments {
			turns[part.TurnID] = true
		}
		if pending := s.pending[t.agent]; pending != nil {
			kept := pending.segments[:0]
			for _, part := range pending.segments {
				if !turns[part.TurnID] {
					kept = append(kept, part)
				}
			}
			if len(kept) == 0 {
				delete(s.pending, t.agent)
			} else {
				pending.segments = kept
				pending.turn = kept[0].TurnID
				text := make([]string, len(kept))
				for i, part := range kept {
					text[i] = part.Text
				}
				pending.text = strings.Join(text, "\n\n")
			}
		}
	}

	if next := s.pending[t.agent]; next != nil {
		delete(s.pending, t.agent)
		next.owner = t.owner
		s.activate(next)
		return
	}
	s.hub.broadcast(Wire{Type: "speech_idle", AgentID: t.agent, Data: map[string]any{"owner": t.owner, "epoch": t.epoch, "turn_id": t.turn}})
	if t.owner == "" && t.turn != 0 {
		if a, ok := s.hub.en.Agent(t.agent); ok {
			_ = a.Post(ensemble.Unpause{TurnID: t.turn})
		}
	}
}
func (s *speaker) cancelAgent(agent string, release bool) {
	delete(s.pending, agent)
	id := s.active[agent]
	t := s.tasks[id]
	if t == nil {
		return
	}
	delete(s.active, agent)
	delete(s.tasks, id)
	if t.timer != nil {
		t.timer.Stop()
	}
	s.hub.broadcast(Wire{Type: "speech_cancelled", AgentID: agent, Data: map[string]any{"id": id, "owner": t.owner, "release": release, "epoch": s.hub.agents[agent].state.Epoch}})
	s.hub.trace("speech_cancelled", agent, id)
}
func (s *speaker) cancelClient(owner string) {
	for agent, id := range s.active {
		if t := s.tasks[id]; t != nil && t.owner == owner {
			s.cancelAgent(agent, false)
		}
	}
}
func (s *speaker) close() {
	s.closed = true
	for agent := range s.active {
		s.cancelAgent(agent, false)
	}
}
func speechParts(text string) ensemble.PartList {
	return ensemble.PartList{{Type: "text", Text: strings.TrimSpace(text)}}
}
