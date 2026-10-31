package gui_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestSavedBudgetEntry builds the actual two-client application and restarts it
// in one data directory. Only the external provider is fake: HTTP rendering,
// tools, the actor, stdin, WebSocket input and persistence all remain real.
func TestSavedBudgetEntry(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "gui")
	if out, err := exec.Command("go", "build", "-o", binary, "./cmd").CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	workspace := t.TempDir()
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			// Messages are the actual provider-bound conversation from the GUI command.
			Messages []struct {
				// Role identifies whether a retained message is human or assistant.
				Role string `json:"role"`
				// Content retains the real wire shape for the EOF/restart assertion.
				Content any `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		prompt, results := "", 0
		for _, message := range req.Messages {
			if message.Role == "user" {
				prompt = fmt.Sprint(message.Content)
				results = 0
			}
			if message.Role == "tool" {
				results++
			}
		}
		if prompt == "small" && results == 0 {
			close(entered)
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
		}
		message := map[string]any{"role": "assistant", "content": "finished eighteen"}
		finish := "stop"
		if results < 18 {
			finish = "tool_calls"
			message["content"] = "I will run the next budget fixture step to verify the saved execution limit."
			message["tool_calls"] = []any{map[string]any{"id": fmt.Sprintf("call-%d", results), "type": "function", "function": map[string]string{"name": "run_command", "arguments": `{"command":"printf budget","yield_time_ms":1000}`}}}
		}
		json.NewEncoder(w).Encode(map[string]any{"model": "gpt-4o-mini", "choices": []any{map[string]any{"message": message, "finish_reason": finish}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1}})
	}))
	defer provider.Close()
	start := func() (*websocket.Conn, io.WriteCloser, func()) {
		t.Helper()
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		port := listener.Addr().(*net.TCPAddr).Port
		listener.Close()
		command := exec.Command(binary, "--port", fmt.Sprint(port), "--no-stream")
		command.Dir = workspace
		command.Env = append(os.Environ(), "LLM_VENDOR=openai", "LLM_MODEL=gpt-4o-mini", "LLM_API_KEY=fake", "LLM_BASE_URL="+provider.URL)
		stdin, _ := command.StdinPipe()
		stderr, _ := command.StderrPipe()
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		// Read startup status before connecting; subsequent stderr is drained so
		// many real tool results cannot fill a pipe and masquerade as a stuck actor.
		ready := make(chan struct{})
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				if strings.HasPrefix(scanner.Text(), "Open http") {
					close(ready)
				}
			}
		}()
		var stopOnce sync.Once
		stop := func() { stopOnce.Do(func() { stdin.Close(); command.Process.Signal(os.Interrupt); command.Wait() }) }
		t.Cleanup(stop)
		select {
		case <-ready:
		case <-time.After(5 * time.Second):
			t.Fatal("application did not start")
		}
		url := "http://127.0.0.1:" + fmt.Sprint(port)
		// Listener startup races the informational print by a few instructions.
		var peer *websocket.Conn
		for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
			peer, _, err = websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(url, "http")+"/ws", nil)
			if err == nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if err != nil {
			t.Fatal(err)
		}
		// This executable runs away from the source tree with no --gui-dir. A
		// working WebSocket alone would not establish deployable browser assets.
		for _, asset := range []string{"/", "/app.js", "/artifact-scroll.js", "/renderers.js"} {
			response, err := http.Get(url + asset)
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil || response.StatusCode != 200 || len(body) == 0 {
				t.Fatalf("relocated asset %s: status=%d bytes=%d err=%v", asset, response.StatusCode, len(body), err)
			}
		}
		t.Cleanup(func() { peer.Close() })
		send(t, peer, "subscribe", "")
		until(t, peer, kind("current_settings"))
		return peer, stdin, stop
	}
	update := func(peer *websocket.Conn, limit int) {
		t.Helper()
		if err := peer.WriteJSON(map[string]any{"type": "update_settings", "settings": map[string]int{"max_tool_rounds": limit}}); err != nil {
			t.Fatal(err)
		}
		until(t, peer, kind("settings_changed"))
	}
	check := func(peer *websocket.Conn, count int, limited bool) {
		t.Helper()
		events := until(t, peer, kind("turn_ended"))
		calls := 0
		for _, event := range events {
			if event["type"] == "tool_dispatched" {
				calls++
			}
		}
		end := events[len(events)-1]
		if calls != count || strings.Contains(fmt.Sprint(end["error"]), "tool round limit") != limited {
			t.Fatalf("budget: calls=%d want=%d limited=%v end=%v", calls, count, limited, end)
		}
	}
	peer, _, stop := start()
	update(peer, 2)
	stop()
	peer, stdin, stop := start()
	// Save a new budget while the first response is outstanding. The old turn
	// keeps its two batches; the terminal's next turn gets the newly saved 18.
	send(t, peer, "prompt", "small")
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("provider not entered")
	}
	update(peer, 18)
	once.Do(func() { close(release) })
	check(peer, 2, true)
	fmt.Fprintln(stdin, "terminal")
	check(peer, 18, false)
	stop()
	peer, _, stop = start()
	send(t, peer, "prompt", "restored")
	check(peer, 18, false)
	update(peer, 0)
	send(t, peer, "prompt", "default")
	check(peer, 18, false)
	stop()
}
