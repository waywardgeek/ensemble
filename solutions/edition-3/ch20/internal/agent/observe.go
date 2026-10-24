package agent

import (
	"ensemble/internal/common"
	"sync"
)

// Queues are owned by the subscriber, not an actor callback executing terminal
// I/O. There is no fixed buffer that could silently drop text or final notices.
type observationQueue struct {
	mu      sync.Mutex
	pending []common.Observation
	wake    chan struct{}
	owner   *agent
	closed  bool
}

func (a *agent) Observe() common.ObservationQueue {
	q := &observationQueue{owner: a, wake: make(chan struct{}, 1)}
	a.observersMu.Lock()
	if a.queues == nil {
		a.queues = make(map[*observationQueue]struct{})
	}
	a.queues[q] = struct{}{}
	a.observersMu.Unlock()
	return q
}
func (q *observationQueue) Wake() <-chan struct{} { return q.wake }
func (q *observationQueue) Drain() []common.Observation {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := q.pending
	q.pending = nil
	return out
}
func (q *observationQueue) push(o common.Observation) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	q.pending = append(q.pending, o)
	select {
	case q.wake <- struct{}{}:
	default:
	}
}
func (q *observationQueue) Close() {
	q.owner.observersMu.Lock()
	delete(q.owner.queues, q)
	q.owner.observersMu.Unlock()
	q.mu.Lock()
	q.closed = true
	q.pending = nil
	q.mu.Unlock()
}

// View returns detached slice headers from an immutable actor checkpoint; no
// History serialization, view callback, or I/O executes on this path.
func (a *agent) View() common.ViewSnapshot {
	v := a.viewValue.Load().(common.ViewSnapshot)
	v.Partials = append([]common.PartDelta(nil), v.Partials...)
	v.Speech.Parts = append(common.PartList(nil), v.Speech.Parts...)
	return v
}
func (a *agent) updateView(o common.Observation) {
	v := &a.viewState
	switch x := o.(type) {
	case common.PartDelta:
		old := a.viewParts[x.PartID]
		if x.Name == "" {
			x.Name = old.Name
		}
		x.Chunk = old.Chunk + x.Chunk
		a.viewParts[x.PartID] = x
	case common.HistoryCommitted:
		e := x.Event
		if e.TurnID != 0 && e.TurnID != v.TurnID && e.Type == common.MessageReceived {
			v.TurnID = e.TurnID
			v.ControlEpoch++
		}
		if e.Type == common.TurnInterrupted {
			v.ControlEpoch++
		}
		if e.Response != nil && e.Type == common.ResponseEnded {
			for _, p := range e.Response.Parts {
				delete(a.viewParts, p.DisplayID)
			}
		}
		if e.Type == common.TurnEnded || e.Type == common.TurnInterrupted {
			clear(a.viewParts)
		}
	case common.StateChanged:
		v.State = common.TurnState(x.To)
	case common.Paused:
		v.Paused = x.Paused
		v.Reason = x.Reason
		if !x.Paused {
			v.ControlEpoch++
			v.ContinueEpoch++
		}
	case common.SpeakingChanged:
		v.Speaking = x.Speaking
		if !x.Speaking {
			v.ControlEpoch++
		}
	case common.SpeechRequested:
		v.SpeechID++
		v.Speech = x
		v.SpeechEpoch = v.ControlEpoch
	case common.TurnEndedObservation:
		clear(a.viewParts)
	}
	v.Revision++
	v.Partials = make([]common.PartDelta, 0, len(a.viewParts))
	for _, part := range a.viewParts {
		v.Partials = append(v.Partials, part)
	}
	a.viewValue.Store(*v)
}
