package common

// PauseGate lets a WebSocket client hold the tool dispatch loop without
// touching the mailbox. The gate is checked before each tool dispatch; the
// WS handler calls Pause/Unpause. Neither imports the other — they share
// only this type through internal/common.

import (
	"context"
	"sync"
)

// PauseGate blocks tool dispatch while a client is paused. It is shared
// between the actor (which checks it) and the WebSocket hub (which sets it).
type PauseGate struct {
	mu     sync.Mutex
	cond   *sync.Cond
	paused bool
}

// NewPauseGate creates an unpaused gate.
func NewPauseGate() *PauseGate {
	g := &PauseGate{}
	g.cond = sync.NewCond(&g.mu)
	return g
}

// Pause sets the gate to paused. The next WaitIfPaused call will block.
func (g *PauseGate) Pause() {
	g.mu.Lock()
	g.paused = true
	g.mu.Unlock()
}

// Unpause clears the pause flag and wakes any goroutine blocked in
// WaitIfPaused.
func (g *PauseGate) Unpause() {
	g.mu.Lock()
	g.paused = false
	g.cond.Broadcast()
	g.mu.Unlock()
}

// WaitIfPaused blocks until unpaused or ctx is done. Returns true if
// unpaused normally, false if the context was cancelled (the actor was
// interrupted or shut down).
func (g *PauseGate) WaitIfPaused(ctx context.Context) bool {
	g.mu.Lock()
	if !g.paused {
		g.mu.Unlock()
		return true
	}

	// Start a goroutine that broadcasts when the context is done, so the
	// cond.Wait loop can check ctx.Err() and bail.
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			g.cond.Broadcast()
		case <-done:
		}
	}()

	for g.paused {
		if ctx.Err() != nil {
			g.mu.Unlock()
			close(done)
			return false
		}
		g.cond.Wait()
	}
	g.mu.Unlock()
	close(done)
	return true
}
