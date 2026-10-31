package grade

// Chapter 9 harness — settings over WebSocket.
//
// Tests settings roundtrip, broadcast to multiple clients, persistence
// across restart, and delivery on new subscription.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// ----------------------------------------------------------------
// Result types
// ----------------------------------------------------------------

// Ch9Result holds evidence for the ch9 checks.
type Ch9Result struct {
	Base string

	BuildOK  bool
	BuildErr string

	// Roundtrip: send update_settings, expect settings_changed back.
	RoundtripOK       bool
	RoundtripSent     map[string]any
	RoundtripReceived map[string]any
	RoundtripErr      string

	// On-connect: new connection receives current_settings with previously-set values.
	OnConnectOK       bool
	OnConnectReceived map[string]any
	OnConnectErr      string

	// Broadcast: second client receives settings_changed when first updates.
	BroadcastOK       bool
	BroadcastReceived map[string]any
	BroadcastErr      string

	// Persist: settings survive binary restart.
	PersistOK       bool
	PersistReceived map[string]any
	PersistErr      string

	// Ch8 parity.
	Ch8Result *Ch8Result
	Ch8Err    string

	HelpersErr string
}

// ch9WsEnvelope captures the raw settings payload that Ch8WsMsg cannot.
type ch9WsEnvelope struct {
	Type     string         `json:"type"`
	Settings map[string]any `json:"settings"`
}

// ----------------------------------------------------------------
// Fake vendor script (minimal — settings tests need only the server)
// ----------------------------------------------------------------

func ch9Replies() []fakevendor.Reply {
	return []fakevendor.Reply{
		{
			Text:  "Settings test response.",
			Usage: fakevendor.Canonical{Input: 10, Output: 5},
		},
	}
}

// ----------------------------------------------------------------
// Main harness
// ----------------------------------------------------------------

// Ch9Run builds and drives the student's binary for settings tests.
func Ch9Run(dir string) (*Ch9Result, error) {
	dir, _ = filepath.Abs(dir)
	r := &Ch9Result{Base: dir}

	tmp, err := os.MkdirTemp("", "ch9grade")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	bin, cleanup, err := Build(dir)
	if err != nil {
		r.BuildErr = err.Error()
		return r, nil
	}
	defer cleanup()
	r.BuildOK = true

	guiDir := filepath.Join(dir, "web", "gui")

	// 1. Roundtrip test.
	ch9DriveRoundtrip(r, bin, guiDir, tmp)

	// 2. Broadcast + on-connect + persist (share one lifecycle).
	ch9DriveBroadcastAndPersist(r, bin, guiDir, tmp)

	// 3. Ch8 parity.
	ch8, err := Ch8Run(dir)
	if err != nil {
		r.Ch8Err = err.Error()
	} else {
		r.Ch8Result = ch8
	}

	return r, nil
}

// ----------------------------------------------------------------
// Raw WebSocket reader for settings messages
// ----------------------------------------------------------------

// ch9RawReader starts a goroutine that reads WebSocket frames into a channel.
type ch9RawReader struct {
	ch   chan json.RawMessage
	done chan error
}

func newCh9RawReader(conn *websocket.Conn) *ch9RawReader {
	rr := &ch9RawReader{
		ch:   make(chan json.RawMessage, 64),
		done: make(chan error, 1),
	}
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				rr.done <- err
				return
			}
			rr.ch <- json.RawMessage(data)
		}
	}()
	return rr
}

// waitForType waits for a message with the given type and returns the parsed
// settings sub-object. Drains other messages.
func (rr *ch9RawReader) waitForType(typ string, timeout time.Duration) (map[string]any, bool) {
	deadline := time.After(timeout)
	for {
		select {
		case raw := <-rr.ch:
			var env ch9WsEnvelope
			if json.Unmarshal(raw, &env) == nil && env.Type == typ {
				return env.Settings, true
			}
		case <-rr.done:
			return nil, false
		case <-deadline:
			return nil, false
		}
	}
}

// drain reads and discards messages for the given duration.
func (rr *ch9RawReader) drain(d time.Duration) {
	deadline := time.After(d)
	for {
		select {
		case <-rr.ch:
		case <-rr.done:
			return
		case <-deadline:
			return
		}
	}
}

// ----------------------------------------------------------------
// Roundtrip test
// ----------------------------------------------------------------

func ch9DriveRoundtrip(r *Ch9Result, bin, guiDir, tmp string) {
	port := freePort()
	if port == "" {
		r.RoundtripErr = "could not find free port"
		return
	}

	srv := fakevendor.New(ch9Replies())
	defer srv.Close()

	rtTmp := filepath.Join(tmp, "roundtrip")
	os.MkdirAll(rtTmp, 0755)

	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = rtTmp
	cmd.Env = ch9Env(srv.URL(), rtTmp, "rt.log")
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		r.RoundtripErr = "stdin pipe: " + err.Error()
		return
	}
	if err := cmd.Start(); err != nil {
		r.RoundtripErr = "start: " + err.Error()
		return
	}
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		r.RoundtripErr = "HTTP server did not start"
		return
	}

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:"+port+"/ws", nil)
	if err != nil {
		r.RoundtripErr = "ws dial: " + err.Error()
		return
	}
	defer conn.Close()

	rr := newCh9RawReader(conn)

	conn.WriteJSON(map[string]any{"type": "subscribe"})
	rr.drain(2 * time.Second) // Drain subscribe handshake.

	// Send update_settings.
	settings := map[string]any{
		"theme":     "light",
		"tts_speed": float64(1.5),
	}
	r.RoundtripSent = settings
	conn.WriteJSON(map[string]any{
		"type":     "update_settings",
		"settings": settings,
	})

	// Wait for settings_changed.
	received, ok := rr.waitForType("settings_changed", 5*time.Second)
	if !ok {
		r.RoundtripErr = "no settings_changed received"
		return
	}
	r.RoundtripOK = true
	r.RoundtripReceived = received
}

// ----------------------------------------------------------------
// Broadcast + on-connect + persist test
// ----------------------------------------------------------------

func ch9DriveBroadcastAndPersist(r *Ch9Result, bin, guiDir, tmp string) {
	port := freePort()
	if port == "" {
		r.BroadcastErr = "could not find free port"
		return
	}

	srv := fakevendor.New(ch9Replies())
	defer srv.Close()

	persistTmp := filepath.Join(tmp, "persist")
	os.MkdirAll(persistTmp, 0755)

	// --- First server instance ---
	cmd1 := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd1.Dir = persistTmp
	cmd1.Env = ch9Env(srv.URL(), persistTmp, "bc.log")
	cmd1.Stderr = os.Stderr
	stdin1, err := cmd1.StdinPipe()
	if err != nil {
		r.BroadcastErr = "stdin pipe: " + err.Error()
		return
	}
	if err := cmd1.Start(); err != nil {
		r.BroadcastErr = "start: " + err.Error()
		return
	}

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		stdin1.Close()
		cmd1.Process.Kill()
		cmd1.Wait()
		r.BroadcastErr = "HTTP server did not start"
		return
	}

	wsURL := "ws://localhost:" + port + "/ws"

	// Connect client 1 and client 2.
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		stdin1.Close()
		cmd1.Process.Kill()
		cmd1.Wait()
		r.BroadcastErr = "ws dial conn1: " + err.Error()
		return
	}
	rr1 := newCh9RawReader(conn1)
	conn1.WriteJSON(map[string]any{"type": "subscribe"})
	rr1.drain(2 * time.Second)

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		conn1.Close()
		stdin1.Close()
		cmd1.Process.Kill()
		cmd1.Wait()
		r.BroadcastErr = "ws dial conn2: " + err.Error()
		return
	}
	rr2 := newCh9RawReader(conn2)
	conn2.WriteJSON(map[string]any{"type": "subscribe"})
	rr2.drain(2 * time.Second)

	// Client 1 sends update_settings.
	settings := map[string]any{
		"theme":     "dark",
		"tts_speed": float64(2.0),
	}
	conn1.WriteJSON(map[string]any{
		"type":     "update_settings",
		"settings": settings,
	})

	// Drain client 1's echo.
	rr1.drain(2 * time.Second)

	// Client 2 should receive settings_changed (broadcast test).
	received, ok := rr2.waitForType("settings_changed", 5*time.Second)
	if ok {
		r.BroadcastOK = true
		r.BroadcastReceived = received
	} else {
		r.BroadcastErr = "conn2 never got settings_changed"
	}

	conn1.Close()
	conn2.Close()

	// --- Kill first server ---
	stdin1.Close()
	cmd1.Process.Kill()
	cmd1.Wait()

	time.Sleep(500 * time.Millisecond) // Let port free.

	// --- Second server instance (same working directory for persistence) ---
	port2 := freePort()
	if port2 == "" {
		r.PersistErr = "could not find free port for second instance"
		return
	}

	srv2 := fakevendor.New(ch9Replies())
	defer srv2.Close()

	cmd2 := exec.Command(bin, "--port", port2, "--gui-dir", guiDir)
	cmd2.Dir = persistTmp
	cmd2.Env = ch9Env(srv2.URL(), persistTmp, "persist.log")
	cmd2.Stderr = os.Stderr
	stdin2, err := cmd2.StdinPipe()
	if err != nil {
		r.PersistErr = "stdin pipe: " + err.Error()
		return
	}
	if err := cmd2.Start(); err != nil {
		r.PersistErr = "start: " + err.Error()
		return
	}
	defer func() {
		stdin2.Close()
		cmd2.Process.Kill()
		cmd2.Wait()
	}()

	if !waitForHTTP("http://localhost:"+port2+"/", 10*time.Second) {
		r.PersistErr = "second HTTP server did not start"
		return
	}

	// Connect and subscribe — should receive current_settings with persisted values.
	conn3, _, err := websocket.DefaultDialer.Dial("ws://localhost:"+port2+"/ws", nil)
	if err != nil {
		r.PersistErr = "ws dial persist: " + err.Error()
		return
	}
	defer conn3.Close()

	rr3 := newCh9RawReader(conn3)
	conn3.WriteJSON(map[string]any{"type": "subscribe"})

	// On second server, current_settings should arrive on subscribe.
	received, ok = rr3.waitForType("current_settings", 5*time.Second)
	if !ok {
		r.PersistErr = "no current_settings received on reconnect"
		r.OnConnectErr = "current_settings not received"
		return
	}

	r.OnConnectOK = true
	r.OnConnectReceived = received

	// Check that theme survived (set to "dark" in the first instance).
	r.PersistReceived = received
	if v, ok := received["theme"]; ok && v == "dark" {
		r.PersistOK = true
	} else {
		r.PersistErr = fmt.Sprintf("theme missing or wrong after restart: %v", received)
	}
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

func ch9Env(baseURL, tmpDir, logName string) []string {
	return append(os.Environ(),
		"LLM_BASE_URL="+baseURL,
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=test-key",
		"CH02_LOG="+filepath.Join(tmpDir, logName),
	)
}
