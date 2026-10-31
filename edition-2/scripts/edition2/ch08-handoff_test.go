package gui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

type ch08HandoffOwner struct {
	*connectorProbeOwner
	entered, release chan struct{}
}

func (o *ch08HandoffOwner) Ch08SnapshotBoundary(ctx context.Context) {
	close(o.entered)
	select {
	case <-o.release:
	case <-ctx.Done():
	}
}
func ch08Handoff(t *testing.T) (*ch08HandoffOwner, *Connector, func() map[string]any, func(map[string]any)) {
	t.Helper()
	base, _ := probeOwner(t)
	owner := &ch08HandoffOwner{base, make(chan struct{}), make(chan struct{})}
	c, peer := probePair(t, owner, true, nil)
	send := func(v map[string]any) {
		t.Helper()
		if e := peer.WriteJSON(v); e != nil {
			t.Fatal(e)
		}
	}
	next := func() map[string]any { return probeJSON(t, peer) }
	send(map[string]any{"type": "subscribe", "id": "s"})
	if x := next(); x["type"] != "preferences_snapshot" {
		t.Fatal("preference snapshot not first", x)
	}
	if x := next(); x["type"] != "snapshot_begin" {
		t.Fatal("Agent snapshot not second", x)
	}
	probeReceive(t, owner.entered)
	return owner, c, next, send
}
func ch08Eventually(t *testing.T, f func() bool, reason string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if f() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal(reason)
}
func TestCh08HandoffControls(t *testing.T) {
	owner, c, next, send := ch08Handoff(t)
	send(map[string]any{"type": "pause", "id": "p", "typing": true, "speaking": false})
	send(map[string]any{"type": "preferences_update", "id": "u", "base_revision": 0, "patch": map[string]any{"font_size": 22}})
	ch08Eventually(t, func() bool {
		c.mu.Lock()
		n := len(c.deferred)
		c.mu.Unlock()
		return n == 1 && owner.Preferences().Snapshot().Revision == 1
	}, "controls did not progress while snapshot held")
	close(owner.release)
	if x := next(); x["type"] != "snapshot_end" {
		t.Fatal("control interleaved initial Agent snapshot", x)
	}
	pause, changed, ack := false, false, false
	for i := 0; i < 8 && !ack; i++ {
		x := next()
		switch x["type"] {
		case "ack":
			pause = x["id"] == "p"
		case "preferences_changed":
			if x["revision"] != float64(1) {
				t.Fatal("new preference revision lost", x)
			}
			changed = true
		case "preferences_ack":
			if !changed {
				t.Fatal("settings ack overtook applied change")
			}
			ack = x["id"] == "u"
		}
	}
	if !pause || !changed || !ack {
		t.Fatal("handoff lost controls or preference change")
	}
}
func TestCh08HandoffCombinedCount(t *testing.T) {
	for _, count := range []int{255, 256, 257} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			owner, c, next, send := ch08Handoff(t)
			for i := 0; i < count; i++ {
				send(map[string]any{"type": "unknown", "id": fmt.Sprint(i)})
			}
			ch08Eventually(t, func() bool { c.mu.Lock(); n := len(c.deferred); c.mu.Unlock(); return n == count || c.ctx.Err() != nil }, "deferred controls did not reach intended capacity")
			close(owner.release)
			if count == 255 {
				if x := next(); x["type"] != "snapshot_end" {
					t.Fatal("exact combined capacity rejected or interleaved", x)
				}
				for i := 0; i < count; i++ {
					x := next()
					if x["type"] != "error" || x["id"] != fmt.Sprint(i) {
						t.Fatal("deferred FIFO/control lost", i, x)
					}
				}
				if c.ctx.Err() != nil {
					t.Fatal("exact combined capacity closed")
				}
			} else {
				select {
				case <-c.Done():
				case <-time.After(2 * time.Second):
					t.Fatal("full deferred queue deadlocked snapshot producer")
				}
				// Native close reason remains required by prior actual-socket overflow tests;
				// here the distinguishing fact is joining the blocked handoff without drops.
				if !strings.Contains(fmt.Sprint(c.ctx.Err()), "canceled") {
					t.Fatal("overflow did not invalidate connection")
				}
			}
		})
	}
}
