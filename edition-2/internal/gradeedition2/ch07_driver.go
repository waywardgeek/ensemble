package gradeedition2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

const guiIntro = "Here is the configuration summary."
const guiAnswer = "The configuration is set to default values."

// The reader has one lifetime, owned by its socket. Returning from an assertion
// closes that lifetime too; a full message queue cannot strand a reader forever.
type guiPeer struct {
	conn     *websocket.Conn
	messages chan map[string]any
	errors   chan error
	pong     chan string
	done     chan struct{}
}

func openGUIPeer(url string) (*guiPeer, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	p := &guiPeer{c, make(chan map[string]any, 256), make(chan error, 1), make(chan string, 1), make(chan struct{})}
	c.SetPongHandler(func(s string) error {
		select {
		case p.pong <- s:
		case <-p.done:
		}
		return nil
	})
	go func() {
		for {
			var m map[string]any
			if err := c.ReadJSON(&m); err != nil {
				p.errors <- err
				return
			}
			select {
			case p.messages <- m:
			case <-p.done:
				return
			}
		}
	}()
	return p, nil
}
func (p *guiPeer) close() { close(p.done); p.conn.Close() }
func (p *guiPeer) send(kind, text string) error {
	return p.conn.WriteJSON(map[string]string{"type": kind, "text": text})
}
func (p *guiPeer) until(timeout time.Duration, done func(map[string]any) bool) ([]map[string]any, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	var out []map[string]any
	for {
		select {
		case m := <-p.messages:
			out = append(out, m)
			if done != nil && done(m) {
				return out, nil
			}
		case err := <-p.errors:
			return out, err
		case <-timer.C:
			if done == nil {
				return out, nil
			}
			return out, fmt.Errorf("WebSocket observation deadline")
		}
	}
}
func messageKind(kind string) func(map[string]any) bool {
	return func(m map[string]any) bool { return m["type"] == kind }
}

// Both chapter implementations dispatch each client command synchronously in
// their WebSocket read loop. Pong is therefore an existing ordered barrier after
// pause, not a new application acknowledgement or a sleep standing in for one.
func (p *guiPeer) pauseBarrier() error {
	if err := p.send("pause", ""); err != nil {
		return err
	}
	if err := p.conn.WriteControl(websocket.PingMessage, []byte("pause-applied"), time.Now().Add(time.Second)); err != nil {
		return err
	}
	select {
	case value := <-p.pong:
		if value == "pause-applied" {
			return nil
		}
		return fmt.Errorf("unexpected pong")
	case err := <-p.errors:
		return err
	case <-time.After(3 * time.Second):
		return fmt.Errorf("pause ordering barrier timed out")
	}
}

// Run the actual GUI executable in a disposable workspace. Stdin stays open for
// the historical combined CLI/GUI, but no protocol translation is performed.
func startGUI(bin, static, model, base, work string, extraEnv ...string) (string, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	port := fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
	listener.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cmd := exec.CommandContext(ctx, bin, "--port", port, "--gui-dir", static)
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL="+model, "LLM_RESOLVED_MODEL=", "LLM_API_KEY=fake-key", "LLM_BASE_URL="+base, "EN_DISABLE_STREAMING=0", "CH02_LOG="+filepath.Join(work, "event.log"))
	cmd.Env = append(cmd.Env, extraEnv...)
	// Discard terminal progress here: WebSocket traffic is the measured client.
	// Drain stderr concurrently through exec and read it only after process exit.
	var stderr bytes.Buffer
	cmd.Stdout = io.Discard
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return "", nil, err
	}
	if err = cmd.Start(); err != nil {
		cancel()
		stdin.Close()
		return "", nil, err
	}
	stop := func() {
		// Release a deliberately blocked command before stopping its owner. The
		// success marker prevents a leftover shell from waiting on removed scratch.
		os.WriteFile(filepath.Join(work, "release"), nil, 0600)
		stdin.Close()
		cmd.Process.Signal(os.Interrupt)
		done := make(chan struct{})
		go func() { cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(time.Second):
			cancel()
			<-done
		}
		cancel()
	}
	client := &http.Client{Timeout: 200 * time.Millisecond}
	url := "http://127.0.0.1:" + port
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, e := client.Get(url + "/")
		if e == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return "ws://127.0.0.1:" + port + "/ws", stop, nil
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	stop()
	return "", nil, fmt.Errorf("GUI did not serve HTTP: %s", strings.TrimSpace(stderr.String()))
}

func driveGUI(bin, static, model string) (live, replay []map[string]any, trace string, err error) {
	work, err := os.MkdirTemp("", "edition2-gui-")
	if err != nil {
		return
	}
	defer os.RemoveAll(work)
	static, err = filepath.Abs(static)
	if err != nil {
		return
	}
	provider := fakevendor.New([]fakevendor.Reply{
		{Thinking: "Let me process this request carefully.", Text: guiIntro, ToolName: "run_command", ToolArgs: `{"command":"printf GUI_TOOL_RESULT"}`, ToolID: "gui_call", Usage: fakevendor.Canonical{Input: 100, Output: 50}},
		{Text: guiAnswer, Usage: fakevendor.Canonical{Input: 160, Output: 30}},
	})
	defer provider.Close()
	url, stop, err := startGUI(bin, static, model, provider.URL(), work)
	if err != nil {
		return
	}
	defer stop()
	first, err := openGUIPeer(url)
	if err != nil {
		return
	}
	defer first.close()
	if err = first.send("subscribe", ""); err != nil {
		return
	}
	if err = first.send("prompt", "show me the configuration"); err != nil {
		return
	}
	live, err = first.until(10*time.Second, messageKind("turn_ended"))
	if err != nil {
		return
	}
	second, e := openGUIPeer(url)
	if e != nil {
		err = e
		return
	}
	defer second.close()
	if err = second.send("subscribe", ""); err != nil {
		return
	}
	// Completed replay has no required end marker. A bounded quiet collection on
	// this otherwise idle Agent avoids inventing one and leaves the socket reader
	// valid. Active-partial handoff is tested causally by the retained source suite.
	replay, err = second.until(500*time.Millisecond, nil)
	if err != nil {
		return
	}
	requests := provider.Requests()
	// A fake's second answer does not establish that the first tool succeeded.
	// Inspect the real request returning its correlated, nonerror output.
	if len(requests) != 2 {
		err = fmt.Errorf("GUI expected two actual provider requests, got %d", len(requests))
		return
	}
	text, bad, e := anthropicToolText(requests[1].Body, "gui_call")
	if e != nil || bad || !strings.Contains(text, "GUI_TOOL_RESULT") {
		err = fmt.Errorf("GUI did not execute actual tool: %v error=%t result=%q", e, bad, text)
		return
	}
	data, e := os.ReadFile(filepath.Join(work, "gui.log"))
	if e != nil {
		err = e
		return
	}
	trace = string(data)
	return
}

func driveGUIPause(bin, static, model string) (messages []map[string]any, err error) {
	work, err := os.MkdirTemp("", "edition2-gui-pause-")
	if err != nil {
		return
	}
	defer os.RemoveAll(work)
	static, err = filepath.Abs(static)
	if err != nil {
		return
	}
	// Real shell completion, not an unknown-tool error, creates the boundary.
	// The first command waits for a file controlled by the WebSocket driver.
	command := "while [ ! -f '" + filepath.Join(work, "release") + "' ]; do sleep .01; done; printf GUI_FIRST"
	commands := []string{command, "printf GUI_SECOND > '" + filepath.Join(work, "second") + "'; printf GUI_SECOND", "printf GUI_THIRD"}
	var calls []fakevendor.ToolCall
	for i, command := range commands {
		args, _ := json.Marshal(map[string]any{"command": command, "ai_callback_delay": 10})
		calls = append(calls, fakevendor.ToolCall{ID: fmt.Sprintf("gui_pause_%d", i+1), Name: "run_command", Args: string(args)})
	}
	provider := fakevendor.New([]fakevendor.Reply{{Tools: calls}, {Text: "All three steps completed."}})
	defer provider.Close()
	url, stop, err := startGUI(bin, static, model, provider.URL(), work)
	if err != nil {
		return
	}
	defer stop()
	first, err := openGUIPeer(url)
	if err != nil {
		return
	}
	defer first.close()
	second, err := openGUIPeer(url)
	if err != nil {
		return
	}
	defer second.close()
	if err = first.send("subscribe", ""); err != nil {
		return
	}
	if err = second.send("subscribe", ""); err != nil {
		return
	}
	if err = first.send("prompt", "run three steps"); err != nil {
		return
	}
	batch, e := first.until(5*time.Second, messageKind("tool_dispatched"))
	messages = append(messages, batch...)
	if e != nil {
		err = e
		return
	}
	if err = first.pauseBarrier(); err != nil {
		return
	}
	if err = os.WriteFile(filepath.Join(work, "release"), nil, 0600); err != nil {
		return
	}
	batch, e = first.until(5*time.Second, messageKind("tool_finished"))
	messages = append(messages, batch...)
	if e != nil {
		err = e
		return
	}
	batch, e = first.until(200*time.Millisecond, nil)
	messages = append(messages, batch...)
	if e != nil {
		err = e
		return
	}
	starts := 0
	for _, m := range messages {
		if m["type"] == "tool_dispatched" {
			starts++
		}
	}
	if starts != 1 {
		err = fmt.Errorf("pause allowed %d tool starts before other-tab unpause", starts)
		return
	}
	if _, e = os.Stat(filepath.Join(work, "second")); e == nil {
		err = fmt.Errorf("second tool actually ran while paused")
		return
	}
	// Explicit unpause from another tab releases this Agent-wide gate. Neither
	// tab has a persistent local veto, and the first result was allowed to finish.
	if err = second.send("unpause", ""); err != nil {
		return
	}
	batch, e = first.until(5*time.Second, messageKind("turn_ended"))
	messages = append(messages, batch...)
	if e != nil {
		err = e
		return
	}
	for i, marker := range []string{"GUI_FIRST", "GUI_SECOND", "GUI_THIRD"} {
		id := fmt.Sprintf("gui_pause_%d", i+1)
		found := false
		for _, m := range messages {
			if m["type"] == "tool_finished" && m["call_id"] == id && m["is_error"] == false && strings.Contains(fmt.Sprint(m["result"]), marker) {
				found = true
			}
		}
		if !found {
			err = fmt.Errorf("pause flow lacks actual nonerror result %s", id)
			return
		}
	}
	data, e := os.ReadFile(filepath.Join(work, "second"))
	if e != nil || string(data) != "GUI_SECOND" {
		err = fmt.Errorf("other-tab unpause did not execute second tool")
		return
	}
	requests := provider.Requests()
	if len(requests) != 2 {
		err = fmt.Errorf("pause flow needs two provider requests, got %d", len(requests))
		return
	}
	for i, marker := range []string{"GUI_FIRST", "GUI_SECOND", "GUI_THIRD"} {
		text, bad, e := anthropicToolText(requests[1].Body, fmt.Sprintf("gui_pause_%d", i+1))
		if e != nil || bad || !strings.Contains(text, marker) {
			err = fmt.Errorf("pause flow provider result %d: %v error=%t result=%q", i+1, e, bad, text)
			return
		}
	}
	return
}
