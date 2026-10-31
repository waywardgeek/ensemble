package agent

import "sync"

// The signal coalesces wakeups; the slice retains every message. Closing a
// producer-facing channel would race with late workers, so this inbox stays open.
type box struct {
	mu    sync.Mutex
	queue []any
	// Wake coalesces notifications; the queue, not this channel, retains messages.
	Wake chan struct{}
}

func newBox() *box { return &box{Wake: make(chan struct{}, 1)} }

// Post accepts live input without waiting for execution or dropping its message.
func (b *box) Post(message any) {
	b.mu.Lock()
	b.queue = append(b.queue, message)
	b.mu.Unlock()
	select {
	case b.Wake <- struct{}{}:
	default:
	}
}

// Drain removes all queued messages in arrival order for the sole actor consumer.
func (b *box) Drain() []any {
	b.mu.Lock()
	defer b.mu.Unlock()
	q := b.queue
	b.queue = nil
	return q
}
