package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// This uses the real DOM, GUI scripts, WebSocket transport and replay renderer.
// No model or speech service is involved. Set CHROME_BIN on other platforms.
func TestGUIReplayAndServerRestart(t *testing.T) {
	chrome := os.Getenv("CHROME_BIN")
	if chrome == "" {
		for _, candidate := range []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "google-chrome", "chromium", "chromium-browser"} {
			if p, err := exec.LookPath(candidate); err == nil {
				chrome = p
				break
			}
		}
	}
	if chrome == "" {
		t.Skip("Chrome not installed; set CHROME_BIN to run the real-browser replay test")
	}
	events := []common.Event{
		msgEvent(1, common.ActorHuman, "first question"),
		responseEvent(2, common.TextPart{Text: "first answer"}),
		msgEvent(3, common.ActorHuman, "last question"),
		responseEvent(4, common.TextPart{Text: "last answer"}),
	}
	newHub := func(events []common.Event) *Hub {
		log := &common.Log{Events: events, Next: common.Seq(len(events) + 1), Clock: time.Now}
		return NewHub(nil, nil, "", log, nil, nil)
	}
	var active atomic.Pointer[Hub]
	active.Store(newHub(events))
	guiDir, err := filepath.Abs("../../web/gui")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { active.Load().ServeWS(w, r) })
	// Keep this test offline. The real renderer already has a no-CDN fallback.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.FileServer(http.Dir(guiDir)).ServeHTTP(w, r)
			return
		}
		data, err := os.ReadFile(filepath.Join(guiDir, "index.html"))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		var lines []string
		for _, line := range strings.Split(string(data), "\n") {
			if !strings.Contains(line, `<script src="https://`) {
				lines = append(lines, line)
			}
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, strings.Join(lines, "\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	browser := startReplayBrowser(t, chrome, srv.URL)
	browser.wait(t, `document.querySelector('#chat-scroll').textContent.includes('last answer')`)
	// A second fetch is asynchronous. A round-trip frame after page load lets
	// both paths settle before asserting order and cardinality.
	time.Sleep(250 * time.Millisecond)
	transcript := `Array.from(document.querySelector('#chat-scroll').children, e => e.textContent)`
	if got := browser.eval(t, transcript); got != `["first question","first answer","last question","last answer"]` {
		t.Errorf("initial replay duplicated or reordered transcript: %s", got)
	}
	first := active.Load()
	first.Observe(common.PartDelta{PartID: 1, Kind: common.DeltaText, Chunk: "before restart"})
	first.Observe(common.PartFinal{PartID: 1, Part: common.TextPart{Text: "before restart"}})
	browser.wait(t, `document.querySelector('#chat-scroll').textContent.includes('before restart')`)

	// Replace the hub just as a fresh process restores the log and resets its
	// ephemeral live-part counter. Leave the browser tab open across disconnect.
	restored := append(append([]common.Event(nil), events...), responseEvent(5, common.TextPart{Text: "before restart"}))
	second := newHub(restored)
	active.Store(second)
	first.mu.Lock()
	for c := range first.clients {
		c.conn.Close()
	}
	first.mu.Unlock()
	deadline := time.Now().Add(10 * time.Second)
	for {
		second.mu.Lock()
		connected := false
		for c := range second.clients {
			connected = connected || c.live
		}
		second.mu.Unlock()
		if connected {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("browser did not reconnect")
		}
		time.Sleep(20 * time.Millisecond)
	}
	second.Observe(common.PartDelta{PartID: 1, Kind: common.DeltaText, Chunk: "after restart"})
	second.Observe(common.PartFinal{PartID: 1, Part: common.TextPart{Text: "after restart"}})
	browser.wait(t, `document.querySelector('#chat-scroll').textContent.includes('after restart')`)
	want := `["first question","first answer","last question","last answer","before restart","after restart"]`
	if got := browser.eval(t, transcript); got != want {
		t.Errorf("reconnect retained stale artifacts or live IDs: %s", got)
	}

	// A final-only response must be visible even without a delta or turn-ended
	// frame following it. Use enough text to overflow the real scroll viewport.
	second.Observe(common.PartFinal{PartID: 2, Part: common.TextPart{Text: strings.Repeat("line\n", 100) + "final-only reply"}})
	browser.wait(t, `document.querySelector('#chat-scroll').textContent.includes('final-only reply')`)
	if got := browser.eval(t, `(() => { const e=document.querySelector('#chat-scroll'); return e.scrollHeight-e.clientHeight-e.scrollTop < 2 })()`); got != "true" {
		t.Error("final-only reply was not scrolled into view")
	}
}

type replayBrowser struct {
	conn    *websocket.Conn
	id      int
	session string
}

func startReplayBrowser(t *testing.T, chrome, url string) *replayBrowser {
	t.Helper()
	profile := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, chrome, "--headless=new", "--no-sandbox", "--disable-gpu", "--no-first-run", "--no-default-browser-check", "--remote-debugging-port=0", "--user-data-dir="+profile, "about:blank")
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() { cancel(); cmd.Wait() })
	var endpoint string
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(filepath.Join(profile, "DevToolsActivePort"))
		parts := strings.Split(strings.TrimSpace(string(data)), "\n")
		if err == nil && len(parts) == 2 {
			endpoint = "ws://127.0.0.1:" + parts[0] + parts[1]
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if endpoint == "" {
		t.Fatal("Chrome did not publish its debugging port")
	}
	conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	b := &replayBrowser{conn: conn}
	result := b.call(t, "Target.createTarget", map[string]any{"url": url})
	var target struct {
		TargetID string `json:"targetId"`
	}
	if err := json.Unmarshal(result, &target); err != nil {
		t.Fatal(err)
	}
	result = b.call(t, "Target.attachToTarget", map[string]any{"targetId": target.TargetID, "flatten": true})
	var attached struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal(result, &attached); err != nil {
		t.Fatal(err)
	}
	b.session = attached.SessionID
	return b
}

func (b *replayBrowser) call(t *testing.T, method string, params any) json.RawMessage {
	t.Helper()
	b.id++
	req := map[string]any{"id": b.id, "method": method, "params": params}
	if b.session != "" {
		req["sessionId"] = b.session
	}
	b.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := b.conn.WriteJSON(req); err != nil {
		t.Fatal(err)
	}
	b.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		var reply struct {
			ID     int             `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  json.RawMessage `json:"error"`
		}
		if err := b.conn.ReadJSON(&reply); err != nil {
			t.Fatal(err)
		}
		if reply.ID != b.id {
			continue
		}
		if len(reply.Error) != 0 {
			t.Fatalf("%s: %s", method, reply.Error)
		}
		return reply.Result
	}
}
func (b *replayBrowser) eval(t *testing.T, expression string) string {
	t.Helper()
	raw := b.call(t, "Runtime.evaluate", map[string]any{"expression": expression, "returnByValue": true})
	var result struct {
		Result struct {
			Value json.RawMessage `json:"value"`
		} `json:"result"`
		Exception json.RawMessage `json:"exceptionDetails"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Exception) != 0 {
		return "exception: " + string(result.Exception)
	}
	return string(result.Result.Value)
}
func (b *replayBrowser) wait(t *testing.T, expression string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if b.eval(t, expression) == "true" {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("browser timed out waiting for %s", expression)
}
