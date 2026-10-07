package ws

import (
	"sync"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// A departing client must never crash the process.
//
// Thirteen call sites send on Client.send, and the teardown at the end of
// ServeWS used to close that channel - while not being one of the senders.
// Both sides mutate shared state under h.mu and then act after releasing it:
// Observe snapshots the live clients and unlocks before sending, teardown
// deletes the client and unlocks before closing. So a client could be
// snapshotted, then deleted and closed, then sent to:
//
//	panic: send on closed channel
//	ws.(*Hub).Observe -> llm.(*Actor).notify -> llm.(*Actor).setState
//
// which killed the agent mid-turn whenever a browser tab closed at the wrong
// moment. It reproduced in roughly one grader run in six.
//
// Note that the select/default guarding most of those sends is no defense:
// default saves a sender from a FULL channel, never from a closed one.
//
// The fix is ownership, not a tighter window: nothing closes c.send at all,
// and writePump stops on c.done instead. This test pins that down. The send
// buffers are deliberately tiny so the buffers fill and the send paths are
// genuinely contended rather than always succeeding immediately.
func TestObserveDuringTeardownDoesNotPanic(t *testing.T) {
	// The window must be widened structurally, not hoped for. Observe
	// snapshots the client set under the lock, releases it, and only then
	// sends. With one client that gap is a few nanoseconds and teardown can
	// essentially never land inside it - an earlier version ran 5000 rounds
	// of one client and never once reproduced the bug.
	//
	// With a large snapshot the send LOOP becomes the window: Observe is
	// still working through the first clients while teardown is already
	// closing the later ones, so the loop reliably reaches a client that has
	// been torn down since the snapshot was taken.
	const (
		rounds      = 50
		clientCount = 2000
		observers   = 2
	)

	// Only log needs a value: Observe reads h.log.Events, while guiLog.Log
	// has a nil-receiver guard.
	h := &Hub{clients: make(map[*Client]bool), log: &common.Log{}}

	for i := 0; i < rounds; i++ {
		all := make([]*Client, 0, clientCount)
		h.mu.Lock()
		for j := 0; j < clientCount; j++ {
			c := &Client{
				hub:  h,
				send: make(chan []byte, 1),
				done: make(chan struct{}),
				live: true,
			}
			all = append(all, c)
			h.clients[c] = true
		}
		h.mu.Unlock()

		var wg sync.WaitGroup
		wg.Add(observers + 1)
		for j := 0; j < observers; j++ {
			go func() {
				defer wg.Done()
				h.Observe(common.StateChanged{})
			}()
		}
		go func() {
			defer wg.Done()
			for _, c := range all {
				h.removeClient(c)
			}
		}()

		// A panic here happens on another goroutine and takes the test
		// binary down with it, which is the failure signal we want: it
		// cannot be swallowed and reported as a pass.
		wg.Wait()
	}
}
