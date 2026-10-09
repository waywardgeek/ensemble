package gui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// silentBrowser subscribes, so it counts as an open GUI, and never answers.
func silentBrowser(t *testing.T, url string) {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial hub: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	conn.WriteJSON(map[string]string{"type": "subscribe"})
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

func pendingCount(h *Server) int {
	h.selfReplies.mu.Lock()
	defer h.selfReplies.mu.Unlock()
	return len(h.selfReplies.pending)
}

// view_gui returns the browser's snapshot text. Two tabs are open and both
// answer every call; the duplicates must find nothing waiting, so once they
// have had time to arrive the router holds no entries. A router that kept the
// entry after delivering would grow by one per call, forever.
func TestViewGUIReturnsSnapshot(t *testing.T) {
	h, wsURL, _ := startMCPPort(t)
	fakeBrowser(t, wsURL, "tab-a")
	fakeBrowser(t, wsURL, "tab-b")
	waitLive(t, h, 2)

	tool := h.ViewGUITool()
	for i := 0; i < 3; i++ {
		out, err := tool.Run(nil, nil)
		if err != nil {
			t.Fatalf("view_gui call %d: %v", i, err)
		}
		if !strings.HasPrefix(out, "snapshot from tab-") {
			t.Fatalf("view_gui call %d returned %q, want the browser's snapshot text", i, out)
		}
	}
	time.Sleep(100 * time.Millisecond) // let the second tab's replies land
	if n := pendingCount(h); n != 0 {
		t.Fatalf("router holds %d entries after every call was answered, want 0", n)
	}
}

// With no browser open the call fails at once with ErrNoGUI, instead of
// waiting out the timeout for a reply that cannot come.
func TestViewGUINoBrowserFailsFast(t *testing.T) {
	h, _, _ := startMCPPort(t)
	start := time.Now()
	_, err := h.ViewGUITool().Run(nil, nil)
	if !errors.Is(err, ErrNoGUI) {
		t.Fatalf("got %v, want ErrNoGUI", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("no-GUI failure took %v; it should not wait for a reply", d)
	}
	if n := pendingCount(h); n != 0 {
		t.Fatalf("router holds %d entries for a request that went nowhere", n)
	}
}

// A browser that never answers: the caller gives up when its context ends,
// and the abandoned id is forgotten rather than left in the router.
func TestCallGUIGivesUpAndForgets(t *testing.T) {
	h, wsURL, _ := startMCPPort(t)
	silentBrowser(t, wsURL)
	waitLive(t, h, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := h.CallGUI(ctx, "tools/call", map[string]any{"name": "gui_snapshot"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want the context deadline", err)
	}
	if n := pendingCount(h); n != 0 {
		t.Fatalf("router holds %d entries after the caller gave up", n)
	}
}
