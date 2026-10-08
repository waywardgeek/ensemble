package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"example.com/ensemble/internal/common"
)

type pauseRegistration struct {
	parent common.Actor
	id     uint64
}

func (r *pauseRegistration) Actor() common.Actor { return r.parent }
func (r *pauseRegistration) Update(typing, speaking bool) (common.PauseState, error) {
	return r.parent.UpdatePause(r.id, typing, speaking, false)
}
func (r *pauseRegistration) Close() error {
	_, err := r.parent.UpdatePause(r.id, false, false, true)
	return err
}

type watchItem struct {
	record common.WatchRecord
	bytes  int
}
type watch struct {
	id     uint64
	parent common.Actor
	mu     sync.Mutex
	items  []watchItem
	bytes  int
	reason string
	wake   chan struct{}
	done   chan struct{}
}

func (w *watch) Actor() common.Actor   { return w.parent }
func (w *watch) Done() <-chan struct{} { return w.done }
func (w *watch) Status() string        { w.mu.Lock(); defer w.mu.Unlock(); return w.reason }
func (w *watch) Close() error          { w.stop("closed"); w.parent.CloseWatch(w.id); return nil }
func (w *watch) stop(reason string)    { w.mu.Lock(); defer w.mu.Unlock(); w.stopLocked(reason) }
func (w *watch) stopLocked(reason string) {
	if w.reason != "" {
		return
	}
	w.reason = reason
	w.items = nil
	w.bytes = 0
	close(w.done)
}
func (w *watch) push(record common.WatchRecord, size int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.reason != "" {
		return
	}
	if len(w.items) >= common.WatchItems || size > common.WatchBytes-w.bytes {
		w.stopLocked("overflow")
		return
	}
	owned, err := Clone(w.parent.Agent().Engine(), record)
	if err != nil {
		w.stopLocked("encoding failure")
		return
	}
	w.items = append(w.items, watchItem{owned, size})
	w.bytes += size
	select {
	case w.wake <- struct{}{}:
	default:
	}
}
func (w *watch) Next(ctx context.Context) (common.WatchRecord, error) {
	for {
		w.mu.Lock()
		if w.reason != "" {
			reason := w.reason
			w.mu.Unlock()
			return common.WatchRecord{}, fmt.Errorf("watch %s", reason)
		}
		if len(w.items) > 0 {
			item := w.items[0]
			w.items[0] = watchItem{}
			w.items = w.items[1:]
			w.bytes -= item.bytes
			w.mu.Unlock()
			return item.record, nil
		}
		w.mu.Unlock()
		select {
		case <-ctx.Done():
			return common.WatchRecord{}, ctx.Err()
		case <-w.done:
		case <-w.wake:
		}
	}
}
func (a *Actor) ask(m common.ActorMessage) (common.ActorReply, error) {
	m.Reply = make(chan common.ActorReply, 1)
	if err := a.enqueue(m, true); err != nil {
		return common.ActorReply{}, err
	}
	reply := <-m.Reply
	return reply, reply.Error
}
func (a *Actor) CloseWatch(id uint64) {
	_ = a.enqueue(common.ActorMessage{Kind: "close_watch", Registration: id}, true)
}
func (a *Actor) Watch() (common.WatchSnapshot, common.Watch, error) {
	r, e := a.ask(common.ActorMessage{Kind: "watch"})
	return r.Snapshot, r.Watch, e
}
func (a *Actor) RegisterPause() (common.PauseRegistration, error) {
	r, e := a.ask(common.ActorMessage{Kind: "register_pause"})
	return r.Registration, e
}
func (a *Actor) UpdatePause(id uint64, typing, speaking, close bool) (common.PauseState, error) {
	kind := "pause"
	if close {
		kind = "close_pause"
	}
	r, e := a.ask(common.ActorMessage{Kind: kind, Registration: id, Typing: typing, Speaking: speaking})
	if close {
		if _, ok := e.(common.StoppedError); ok {
			return r.Pause, nil
		}
	}
	return r.Pause, e
}
func (a *Actor) pauseState() common.PauseState {
	p := common.PauseState{Revision: a.revision}
	for _, causes := range a.pauses {
		if causes[0] {
			p.TypingClients++
		}
		if causes[1] {
			p.SpeakingClients++
		}
	}
	p.Paused = p.TypingClients+p.SpeakingClients > 0
	return p
}
func (a *Actor) receiveWatch(m common.ActorMessage) bool {
	reply := common.ActorReply{}
	switch m.Kind {
	case "close_watch":
		for i, w := range a.watches {
			if w.id == m.Registration {
				copy(a.watches[i:], a.watches[i+1:])
				a.watches[len(a.watches)-1] = nil
				a.watches = a.watches[:len(a.watches)-1]
				break
			}
		}
		return true
	case "watch":
		if a.projectionInvalid {
			m.Reply <- common.ActorReply{Error: fmt.Errorf("active projection unavailable; retry after model end")}
			return true
		}
		a.nextWatch++
		s := a.parent.WatchSource()
		s.AgentID = a.parent.ID()
		s.Generation = fmt.Sprintf("%s-g%d", s.AgentID, a.nextWatch)
		s.Watermark = a.revision
		s.State.Lifecycle = a.state
		s.State.QueuedRequestIDs = []string{}
		if a.active != nil {
			id := a.active.id
			s.State.ActiveRequestID = &id
		}
		for _, r := range a.pending {
			s.State.QueuedRequestIDs = append(s.State.QueuedRequestIDs, r.id)
		}
		p := a.pauseState()
		s.State.Paused = p.Paused
		s.State.TypingClients = p.TypingClients
		s.State.SpeakingClients = p.SpeakingClients
		if a.model != nil {
			op := common.ActiveOperation{AgentID: s.AgentID, RequestID: a.model.RequestID(), OperationID: a.model.ID(), Delivery: a.model.Delivery()}
			s.State.ActiveOperation = &op
			ids := []int{}
			for id := range a.partials {
				ids = append(ids, id)
			}
			sort.Ints(ids)
			for _, id := range ids {
				part := common.WatchPartial{ActiveOperation: op, PartID: id, Channels: map[string]string{}}
				for ch, b := range a.partials[id] {
					part.Channels[ch] = string(b)
				}
				s.Partials = append(s.Partials, part)
			}
		}
		w := &watch{id: a.nextWatch, parent: a, wake: make(chan struct{}, 1), done: make(chan struct{})}
		a.watches = append(a.watches, w)
		reply.Snapshot = s
		reply.Watch = w
	case "register_pause":
		a.nextPause++
		a.pauses[a.nextPause] = [2]bool{}
		reply.Registration = &pauseRegistration{parent: a, id: a.nextPause}
	case "pause", "close_pause":
		before := a.pauseState()
		_, exists := a.pauses[m.Registration]
		if m.Kind == "close_pause" {
			delete(a.pauses, m.Registration)
		} else if !exists {
			reply.Error = fmt.Errorf("pause registration closed")
		} else {
			a.pauses[m.Registration] = [2]bool{m.Typing, m.Speaking}
		}
		after := a.pauseState()
		if before.TypingClients != after.TypingClients || before.SpeakingClients != after.SpeakingClients {
			a.publish(common.Observation{Kind: "pause_changed", AgentID: a.parent.ID(), Paused: after.Paused, TypingClients: after.TypingClients, SpeakingClients: after.SpeakingClients})
		}
		reply.Pause = a.pauseState()
		// The acknowledgment linearizes before a later pending admission. A separate
		// mailbox item lets already queued controls keep their ordering as well.
		if before.Paused && !after.Paused && a.active != nil && a.state == "tools_pending" && a.active.report == nil {
			_ = a.enqueue(common.ActorMessage{Kind: "resume_tools"}, false)
		}
	default:
		return false
	}
	m.Reply <- reply
	return true
}

// PublishDurable is called only by the owning Agent's actor-admitted append path.
// It is synchronous: enqueueing to our own mailbox would lose the snapshot cut.
func (a *Actor) PublishDurable(e common.Event) {
	a.publish(common.Observation{AgentID: a.parent.ID(), Kind: e.Type, Seq: e.Seq, Event: e})
}
func (a *Actor) publish(o common.Observation) {
	a.revision++
	if o.Kind == "model_begin" {
		a.projectionInvalid = false
		a.partials = map[int]map[string][]byte{}
		a.partialBytes = 0
	}
	if o.Kind == "part_delta" {
		if a.partialBytes+len(o.Text) <= 16*1024*1024 {
			if a.partials[o.PartID] == nil {
				a.partials[o.PartID] = map[string][]byte{}
			}
			a.partials[o.PartID][o.Channel] = append(a.partials[o.PartID][o.Channel], o.Text...)
			a.partialBytes += len(o.Text)
		} else {
			a.projectionInvalid = true
			for _, w := range a.watches {
				w.stop("projection overflow")
			}
		}
	}
	if o.Kind == "model_end" {
		a.projectionInvalid = false
		a.partials = nil
		a.partialBytes = 0
	}
	r := common.WatchRecord{Revision: a.revision, Observation: o}
	encoded, err := json.Marshal(r)
	live := a.watches[:0]
	for _, w := range a.watches {
		if err != nil {
			w.stop("encoding failure")
		} else {
			w.push(r, len(encoded))
		}
		if w.Status() == "" {
			live = append(live, w)
		}
	}
	clear(a.watches[len(live):])
	a.watches = live
	a.parent.Ensemble().Observe(o)
}

func RenderableEvent(owner common.Engine, kind string) bool {
	switch kind {
	case "message_received", "hint_received", "response_ended", "tool_called", "tool_returned", "job_ended", "job_killed", "turn_started", "turn_ended", "error_occurred":
		return true
	}
	return false
}
