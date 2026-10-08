package llm

import (
	"context"
	"sync"
	"unicode/utf8"

	"example.com/ensemble/internal/common"
)

const pendingLimit = 1 << 20
const drainLimit = 64 << 10
const responseLimit = 16 << 20

// The producer and actor share only this bounded transfer store. A notice owns
// readiness until Drain transfers it back; HTTP never holds the mailbox lock.
type operation struct {
	parent             common.ModelEngine
	id, requestID      string
	config             common.Config
	mu                 sync.Mutex
	pending            []common.Fragment
	bytes              int
	noticed, discarded bool
	changed            chan struct{}
}

func (e *Engine) NewOperation(id, requestID string, config common.Config) common.ModelOperation {
	return &operation{parent: e, id: id, requestID: requestID, config: config, changed: make(chan struct{})}
}
func (o *operation) Engine() common.ModelEngine { return o.parent }
func (o *operation) ID() string                 { return o.id }
func (o *operation) RequestID() string          { return o.requestID }
func (o *operation) Delivery() string {
	if o.config.DisableStreaming {
		return "plain"
	}
	return "stream"
}
func (o *operation) signal() { close(o.changed); o.changed = make(chan struct{}) }
func (o *operation) Emit(ctx context.Context, f common.Fragment) error {
	if !utf8.ValidString(f.Text) {
		return failure(o.parent, "invalid UTF-8 fragment")
	}
	for len(f.Text) > 0 {
		o.mu.Lock()
		if o.discarded {
			o.mu.Unlock()
			return context.Canceled
		}
		capacity := pendingLimit - o.bytes
		if capacity == 0 {
			changed := o.changed
			o.mu.Unlock()
			select {
			case <-changed:
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		n := len(f.Text)
		if n > capacity {
			n = capacity
		}
		for n > 0 && n < len(f.Text) && !utf8.RuneStart(f.Text[n]) {
			n--
		}
		if n == 0 {
			changed := o.changed
			o.mu.Unlock()
			select {
			case <-changed:
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		piece := f
		piece.Text = f.Text[:n]
		f.Text = f.Text[n:]
		o.pending = append(o.pending, piece)
		o.bytes += n
		notify := !o.noticed
		o.noticed = true
		o.mu.Unlock()
		if notify {
			o.parent.Agent().ModelReady(o)
		}
	}
	return nil
}
func (o *operation) Drain() ([]common.Fragment, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	out := []common.Fragment{}
	budget := drainLimit
	for len(o.pending) > 0 && budget > 0 {
		f := o.pending[0]
		n := len(f.Text)
		if n > budget {
			n = budget
		}
		for n > 0 && n < len(f.Text) && !utf8.RuneStart(f.Text[n]) {
			n--
		}
		if n == 0 {
			break
		}
		piece := f
		piece.Text = f.Text[:n]
		out = append(out, piece)
		budget -= n
		o.bytes -= n
		if n == len(f.Text) {
			o.pending[0] = common.Fragment{}
			o.pending = o.pending[1:]
		} else {
			o.pending[0].Text = f.Text[n:]
		}
	}
	more := len(o.pending) > 0
	o.noticed = more
	o.signal()
	return out, more
}
func (o *operation) Discard() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.discarded = true
	o.pending = nil
	o.bytes = 0
	o.signal()
}
func (o *operation) Exchange(ctx context.Context, body []byte) (common.ParsedResponse, error) {
	parsed, err := o.parent.ExchangeOperation(ctx, o, body, o.config)
	// Even an error follows fragments already decoded. Cancellation releases this
	// wait; the actor has already rejected and discarded that operation.
	for {
		o.mu.Lock()
		empty := o.bytes == 0
		changed := o.changed
		o.mu.Unlock()
		if empty {
			return parsed, err
		}
		select {
		case <-changed:
		case <-ctx.Done():
			return common.ParsedResponse{}, ctx.Err()
		}
	}
}
