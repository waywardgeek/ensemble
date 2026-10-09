package common

// The inbound seam: what the agent hears.
//
// ONE queue carries all of it — that is the entire point. A prompt, a hint
// typed mid-turn, a tool finishing, and an interrupt are the same kind of
// event to the loop.

import (
	"context"
	"sync"
)

// Inbound is anything the agent can hear. Sealed union.
type Inbound interface{ isInbound() }

// UserMessage starts a new turn.
type UserMessage struct{ Text string }

// Request is a blocking caller's distinct turn (or ephemeral attachment).
// Completion is runtime control data, never serialized into the event log.
type Request struct {
	Text      string
	Ephemeral bool
	Context   context.Context
	Reply     chan TurnResult
}
type TurnResult struct {
	Text string
	Err  error
}

// Hint is a message that arrives while a turn is already running. It is not
// a separate channel: classification happens in the reducer, by turn state,
// never at capture.
type Hint struct{ Text string }

// ToolCompleted is DELIVERED INTO the queue rather than awaited in a select.
// This is the whole fix. The tool already ran on its own goroutine; the agent
// was deaf because the loop that would drain events was parked waiting for it.
type ToolCompleted struct {
	CallID  string
	Result  string
	IsError bool
	// Workers return data; only the actor records it. Nil Tool means already recorded.
	Tool *ToolData
}

// ToolEvents carries effects once a tool returns, independently of its job
// report: a callback deadline can report the still-running job earlier.
type ToolEvents struct{ Events []Event }

func (ToolEvents) isInbound() {}

// Interrupt tells the actor loop to stop processing the current turn.
type Interrupt struct{}

// Reset clears the conversation and the memories recalled to serve it.
//
// It travels as an inbound message rather than as a direct call because the
// actor owns the context; anything else would mutate it from another
// goroutine. Like Interrupt, it is a bare struct: the instruction carries no
// argument because there is nothing to choose.
type Reset struct{}

// SetModel switches the model that subsequent requests are rendered for.
//
// It travels as an inbound message for the same reason Reset does: the actor
// owns the config, and applying a switch from the WebSocket goroutine could
// land midway through rendering a request. Delivered into the queue, it is
// applied between turns, which is what makes "switch at any time" safe: ask
// mid-turn and the change takes effect when that turn ends, never partway
// through a render.
type SetModel struct{ Model string }

func (Request) isInbound()       {}
func (UserMessage) isInbound()   {}
func (Hint) isInbound()          {}
func (ToolCompleted) isInbound() {}
func (Interrupt) isInbound()     {}
func (Reset) isInbound()         {}
func (SetModel) isInbound()      {}

// Mailbox is a mutex-guarded FIFO queue with a signal channel. Post never
// blocks or drops.
//
// waywardgeest — the mailbox is where the ghost leaves its letters.
type Mailbox struct {
	mu    sync.Mutex
	queue []Inbound
	sig   chan struct{}
}

// NewMailbox creates a ready-to-use mailbox.
func NewMailbox() *Mailbox {
	return &Mailbox{sig: make(chan struct{}, 1)}
}

// Post enqueues a message. It never blocks and never drops.
func (m *Mailbox) Post(msg Inbound) {
	m.mu.Lock()
	m.queue = append(m.queue, msg)
	m.mu.Unlock()
	// Non-blocking signal: if there is already a pending signal, skip.
	select {
	case m.sig <- struct{}{}:
	default:
	}
}

// Drain returns all queued messages, clearing the queue. Returns nil if empty.
func (m *Mailbox) Drain() []Inbound {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.queue) == 0 {
		return nil
	}
	out := m.queue
	m.queue = nil
	return out
}

// Signal returns the channel to select on for new messages.
func (m *Mailbox) Signal() <-chan struct{} {
	return m.sig
}
