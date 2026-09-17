package common

// The inbound seam: what the agent hears.
//
// ONE queue carries all of it — that is the entire point. A prompt, a hint
// typed mid-turn, a tool finishing, and an interrupt are the same kind of
// event to the loop.

import "sync"

// Inbound is anything the agent can hear. Sealed union.
type Inbound interface{ isInbound() }

// UserMessage starts a new turn.
type UserMessage struct{ Text string }

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
}

// Interrupt tells the actor loop to stop processing the current turn.
type Interrupt struct{}

func (UserMessage) isInbound()   {}
func (Hint) isInbound()          {}
func (ToolCompleted) isInbound() {}
func (Interrupt) isInbound()     {}

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
