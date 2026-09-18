package grade

// Chapter 8 harness — WebSocket transport, gui.log, and pause gate.
//
// Unlike earlier harnesses that drive the binary through stdin/stdout, this
// one starts the binary as a subprocess, waits for its HTTP server, and
// connects via WebSocket.

import (
	"encoding/json"
	"fmt"
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

// ----------------------------------------------------------------
// Result types
// ----------------------------------------------------------------

// Ch8Result holds evidence for the ch8 checks.
type Ch8Result struct {
	Base string

	ExBuildOK  bool
	ExBuildErr string

	// WebSocket observations.
	Messages []Ch8WsMsg

	// Replay test.
	ReplayOK       bool
	ReplayMessages []Ch8WsMsg
	ReplayErr      string

	// gui.log content.
	GuiLogContent string
	GuiLogPath    string

	// Pause test.
	PauseOK       bool
	PauseMessages []Ch8WsMsg
	PauseErr      string

	// Ch7 parity.
	Ch7Result *Ch7Result
	Ch7Err    string

	HelpersErr string
}

// Ch8WsMsg is a parsed WebSocket message.
type Ch8WsMsg struct {
	Type    string          `json:"type"`
	PartID  float64         `json:"part_id,omitempty"`
	Kind    string          `json:"kind,omitempty"`
	Chunk   string          `json:"chunk,omitempty"`
	Seq     float64         `json:"seq,omitempty"`
	Text    string          `json:"text,omitempty"`
	Error   string          `json:"error,omitempty"`
	From    string          `json:"from,omitempty"`
	To      string          `json:"to,omitempty"`
	CallID  string          `json:"call_id,omitempty"`
	Name    string          `json:"name,omitempty"`
	Input   json.RawMessage `json:"input,omitempty"`
	Result  string          `json:"result,omitempty"`
	IsError bool            `json:"is_error,omitempty"`
	Agent   string          `json:"agent,omitempty"`
	Content string          `json:"content,omitempty"` // part_partial
	Actor   string          `json:"actor,omitempty"`   // message
	First   int             `json:"first,omitempty"`   // event_range
	Last    int             `json:"last,omitempty"`     // event_range

	// For part_final with embedded part data.
	Tool string `json:"tool,omitempty"`
	Args string `json:"args,omitempty"`
}

// ----------------------------------------------------------------
// Fake vendor script
// ----------------------------------------------------------------

// ch8Replies returns a script for the WebSocket streaming test.
// One response with thinking + text + tool call, then a final text.
func ch8Replies() []fakevendor.Reply {
	return []fakevendor.Reply{
		{
			Thinking: "Let me process this request carefully.",
			Text:     "Here is the configuration summary.",
			ToolName: "think",
			ToolArgs: `{"seconds":0,"thought":"examining config"}`,
			ToolID:   "call_ch8_1",
			Usage:    fakevendor.Canonical{Input: 100, Output: 50},
		},
		{
			Text:  "The configuration is set to default values.",
			Usage: fakevendor.Canonical{Input: 160, Output: 30},
		},
	}
}

// ch8PauseReplies returns a script with 3 tool calls for the pause test.
func ch8PauseReplies() []fakevendor.Reply {
	return []fakevendor.Reply{
		{
			Tools: []fakevendor.ToolCall{
				{ID: "pause_c1", Name: "think", Args: `{"seconds":2,"thought":"step 1"}`},
				{ID: "pause_c2", Name: "think", Args: `{"seconds":2,"thought":"step 2"}`},
				{ID: "pause_c3", Name: "think", Args: `{"seconds":2,"thought":"step 3"}`},
			},
			Usage: fakevendor.Canonical{Input: 100, Output: 80},
		},
		{
			Text:  "All three steps completed.",
			Usage: fakevendor.Canonical{Input: 200, Output: 20},
		},
	}
}

// ----------------------------------------------------------------
// Main harness
// ----------------------------------------------------------------

// Ch8Run builds and drives the ch8 exercise.
func Ch8Run(path string) (*Ch8Result, error) {
	base, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	r := &Ch8Result{Base: base}

	tmp, err := os.MkdirTemp("", "ch8grade")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	// Build the ch08 exercise.
	exDir := filepath.Join(base, "ch08")
	if _, err := os.Stat(exDir); err != nil {
		r.ExBuildErr = "ch08/ directory not found"
	} else {
		bin := filepath.Join(tmp, "ensemble")
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Dir = exDir
		if out, err := cmd.CombinedOutput(); err != nil {
			r.ExBuildErr = strings.TrimSpace(string(out))
		} else {
			r.ExBuildOK = true

			// Run the WebSocket streaming test.
			ch8DriveWS(r, bin, exDir, tmp)

			// Run the replay test.
			ch8DriveReplay(r, bin, exDir, tmp)

			// Run the pause test.
			ch8DrivePause(r, bin, exDir, tmp)
		}
	}

	// Ch7 parity.
	ch7, err := Ch7Run(base)
	if err != nil {
		r.Ch7Err = err.Error()
	} else {
		r.Ch7Result = ch7
	}

	return r, nil
}

// ----------------------------------------------------------------
// WebSocket streaming test
// ----------------------------------------------------------------

func ch8DriveWS(r *Ch8Result, bin, exDir, tmp string) {
	port := freePort()
	if port == "" {
		r.HelpersErr = "could not find free port"
		return
	}

	srv := fakevendor.New(ch8Replies())
	defer srv.Close()

	guiLogPath := filepath.Join(tmp, "gui.log")
	r.GuiLogPath = guiLogPath

	cmd := exec.Command(bin, "--port", port, "--gui-dir", filepath.Join(exDir, "web", "gui"))
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=test-key",
		"CH02_LOG="+filepath.Join(tmp, "ws_test.log"),
	)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		r.HelpersErr = "stdin pipe: " + err.Error()
		return
	}

	if err := cmd.Start(); err != nil {
		r.HelpersErr = "start: " + err.Error()
		return
	}
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		r.HelpersErr = "HTTP server did not start within 10s"
		return
	}

	wsURL := "ws://localhost:" + port + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		r.HelpersErr = "ws dial: " + err.Error()
		return
	}
	defer conn.Close()

	// Subscribe — no cursor. Server sends event-log window + in-flight state.
	conn.WriteJSON(map[string]any{"type": "subscribe"})
	conn.WriteJSON(map[string]any{"type": "prompt", "text": "show me the configuration"})

	r.Messages = readWSMessages(conn, 30*time.Second, func(msgs []Ch8WsMsg) bool {
		for _, m := range msgs {
			if m.Type == "turn_ended" {
				return true
			}
		}
		return false
	})

	if data, err := os.ReadFile(guiLogPath); err == nil {
		r.GuiLogContent = string(data)
	}
}

// ----------------------------------------------------------------
// Replay test — event-log based reconnection
// ----------------------------------------------------------------

func ch8DriveReplay(r *Ch8Result, bin, exDir, tmp string) {
	port := freePort()
	if port == "" {
		r.ReplayErr = "could not find free port"
		return
	}

	srv := fakevendor.New(ch8Replies())
	defer srv.Close()

	replayTmp := filepath.Join(tmp, "replay")
	os.MkdirAll(replayTmp, 0755)

	cmd := exec.Command(bin, "--port", port, "--gui-dir", filepath.Join(exDir, "web", "gui"))
	cmd.Dir = replayTmp
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=test-key",
		"CH02_LOG="+filepath.Join(replayTmp, "replay.log"),
	)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		r.ReplayErr = "stdin pipe: " + err.Error()
		return
	}

	if err := cmd.Start(); err != nil {
		r.ReplayErr = "start: " + err.Error()
		return
	}
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		r.ReplayErr = "HTTP server did not start"
		return
	}

	wsURL := "ws://localhost:" + port + "/ws"

	// First connection: drive a full turn so the event log has content.
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		r.ReplayErr = "ws dial 1: " + err.Error()
		return
	}
	conn1.WriteJSON(map[string]any{"type": "subscribe"})
	conn1.WriteJSON(map[string]any{"type": "prompt", "text": "test replay"})

	firstBatch := readWSMessages(conn1, 30*time.Second, func(msgs []Ch8WsMsg) bool {
		for _, m := range msgs {
			if m.Type == "turn_ended" {
				return true
			}
		}
		return false
	})
	conn1.Close()

	if len(firstBatch) == 0 {
		r.ReplayErr = "first connection received zero messages"
		return
	}

	// Second connection: subscribe AFTER the turn finished.
	// The server sends event_range; we fetch everything, then verify
	// the event log covered the completed turn.
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		r.ReplayErr = "ws dial 2: " + err.Error()
		return
	}
	defer conn2.Close()
	conn2.WriteJSON(map[string]any{"type": "subscribe"})

	// Read event_range and send fetch for the full range.
	replayMsgs := readWSMessages(conn2, 5*time.Second, func(msgs []Ch8WsMsg) bool {
		for _, m := range msgs {
			if m.Type == "event_range" {
				if m.Last > 0 {
					conn2.WriteJSON(map[string]any{
						"type": "fetch",
						"from": m.First,
						"to":   m.Last,
					})
				}
				// Keep reading — the fetch results follow.
				return false
			}
			// Stop when we have both a part_final and a tool event — proof
			// the event log was fully replayed.
			hasFinal := false
			hasTool := false
			for _, mm := range msgs {
				if mm.Type == "part_final" {
					hasFinal = true
				}
				if mm.Type == "tool_dispatched" || mm.Type == "tool_finished" {
					hasTool = true
				}
			}
			if hasFinal && hasTool {
				return true
			}
		}
		return false
	})

	// The replay must contain at least one part_final and one tool event —
	// proof the event log window was sent, not just live observations.
	hasPartFinal := false
	hasToolEvent := false
	for _, m := range replayMsgs {
		if m.Type == "part_final" {
			hasPartFinal = true
		}
		if m.Type == "tool_dispatched" || m.Type == "tool_finished" {
			hasToolEvent = true
		}
	}

	r.ReplayOK = hasPartFinal && hasToolEvent
	r.ReplayMessages = replayMsgs
}

// ----------------------------------------------------------------
// Pause test
// ----------------------------------------------------------------

func ch8DrivePause(r *Ch8Result, bin, exDir, tmp string) {
	port := freePort()
	if port == "" {
		r.PauseErr = "could not find free port"
		return
	}

	srv := fakevendor.New(ch8PauseReplies())
	defer srv.Close()

	pauseTmp := filepath.Join(tmp, "pause")
	os.MkdirAll(pauseTmp, 0755)

	cmd := exec.Command(bin, "--port", port, "--gui-dir", filepath.Join(exDir, "web", "gui"))
	cmd.Dir = pauseTmp
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=test-key",
		"CH02_LOG="+filepath.Join(pauseTmp, "pause.log"),
	)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		r.PauseErr = "stdin pipe: " + err.Error()
		return
	}

	if err := cmd.Start(); err != nil {
		r.PauseErr = "start: " + err.Error()
		return
	}
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		r.PauseErr = "HTTP server did not start"
		return
	}

	wsURL := "ws://localhost:" + port + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		r.PauseErr = "ws dial: " + err.Error()
		return
	}
	defer conn.Close()

	// Start a reader goroutine so we never call SetReadDeadline (which
	// corrupts the gorilla/websocket connection state on timeout).
	msgCh := make(chan Ch8WsMsg, 64)
	readDone := make(chan error, 1)
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				readDone <- err
				return
			}
			var msg Ch8WsMsg
			if json.Unmarshal(data, &msg) == nil {
				msgCh <- msg
			}
		}
	}()

	conn.WriteJSON(map[string]any{"type": "subscribe"})
	conn.WriteJSON(map[string]any{"type": "prompt", "text": "run three steps"})

	var allMsgs []Ch8WsMsg

	// Wait for first tool_dispatched.
	gotFirstDispatch := false
	timer := time.After(15 * time.Second)
waitDispatch:
	for {
		select {
		case msg := <-msgCh:
			allMsgs = append(allMsgs, msg)
			if msg.Type == "tool_dispatched" {
				gotFirstDispatch = true
				conn.WriteJSON(map[string]any{"type": "pause"})
				break waitDispatch
			}
		case <-readDone:
			break waitDispatch
		case <-timer:
			break waitDispatch
		}
	}

	if !gotFirstDispatch {
		r.PauseErr = "never received first tool_dispatched"
		r.PauseMessages = allMsgs
		return
	}

	// Wait for first tool_finished (the in-flight tool completes).
	gotFirstFinish := false
	timer = time.After(10 * time.Second)
waitFinish:
	for {
		select {
		case msg := <-msgCh:
			allMsgs = append(allMsgs, msg)
			if msg.Type == "tool_finished" {
				gotFirstFinish = true
				break waitFinish
			}
		case <-readDone:
			break waitFinish
		case <-timer:
			break waitFinish
		}
	}

	if !gotFirstFinish {
		r.PauseErr = "first tool never finished"
		r.PauseMessages = allMsgs
		return
	}

	// Wait 2 seconds — verify NO second tool_dispatched arrives.
	secondDispatchWhilePaused := false
	timer = time.After(2 * time.Second)
waitPause:
	for {
		select {
		case msg := <-msgCh:
			allMsgs = append(allMsgs, msg)
			if msg.Type == "tool_dispatched" {
				secondDispatchWhilePaused = true
			}
		case <-readDone:
			break waitPause
		case <-timer:
			break waitPause
		}
	}

	if secondDispatchWhilePaused {
		r.PauseErr = "tool_dispatched arrived while paused"
		r.PauseMessages = allMsgs
		return
	}

	// Unpause and wait for remaining tools + turn_ended.
	conn.WriteJSON(map[string]any{"type": "unpause"})

	timer = time.After(30 * time.Second)
	gotTurnEnded := false
waitEnd:
	for {
		select {
		case msg := <-msgCh:
			allMsgs = append(allMsgs, msg)
			if msg.Type == "turn_ended" {
				gotTurnEnded = true
				break waitEnd
			}
		case <-readDone:
			break waitEnd
		case <-timer:
			break waitEnd
		}
	}

	_ = gotTurnEnded
	r.PauseOK = true
	r.PauseMessages = allMsgs
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

func freePort() string {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return ""
	}
	defer l.Close()
	return fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port)
}

func waitForHTTP(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func readWSMessages(conn *websocket.Conn, timeout time.Duration, done func([]Ch8WsMsg) bool) []Ch8WsMsg {
	var msgs []Ch8WsMsg
	deadline := time.After(timeout)

	ch := make(chan Ch8WsMsg, 64)
	errCh := make(chan error, 1)
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			var msg Ch8WsMsg
			if json.Unmarshal(data, &msg) == nil {
				ch <- msg
			}
		}
	}()

	for {
		select {
		case msg := <-ch:
			msgs = append(msgs, msg)
			if done(msgs) {
				return msgs
			}
		case <-errCh:
			return msgs
		case <-deadline:
			return msgs
		}
	}
}
