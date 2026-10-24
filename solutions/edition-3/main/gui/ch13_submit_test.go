package gui

import (
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// The page uses production input/form handlers. The Go side delivers their
// commands through the real WS server and feeds back only real server frames.
// Holding claims until all submit commands are processed makes the old epoch
// invalidation deterministic; no preclaimed speech frame is injected.
func TestCh13SubmitPromptSurvivesOwnRelease(t *testing.T) {
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done() // Isolate prompt readback from model response speech.
	}))
	app := ensemble.New(nil)
	t.Cleanup(func() { _ = app.Shutdown(); peer.Close() })
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "offline", APIKey: "fake", BaseURL: peer.URL, StreamCapability: ensemble.StreamPlain, DisableStreaming: true, Speaking: true, Tools: []string{}, BaseDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServerWithSettings(app, "web/gui", filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	server.DefaultAgent = a.ID()
	web := httptest.NewServer(server.Handler(server.Assets))
	defer web.Close()
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(web.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	node := exec.CommandContext(ctx, "node", "ch13_submit_bridge.cjs")
	var stderr safeBuffer
	node.Stderr = &stderr
	in, err := node.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	out, err := node.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = node.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = in.Close()
		if err := node.Wait(); err != nil {
			t.Errorf("controlled page exit: %v; %s", err, stderr.String())
		}
	}()
	enc, dec := json.NewEncoder(in), json.NewDecoder(out)
	type pageReply struct {
		Error  string    `json:"error"`
		Sent   []command `json:"sent"`
		Epoch  uint64    `json:"epoch"`
		Status struct {
			Ready bool `json:"ready"`
		} `json:"status"`
		Channel struct {
			Events []struct{ Event, Source, Text string } `json:"events"`
		} `json:"channel"`
	}
	step := func(op any) pageReply {
		t.Helper()
		if err := enc.Encode(op); err != nil {
			t.Fatal(err)
		}
		var reply pageReply
		if err := dec.Decode(&reply); err != nil {
			t.Fatalf("controlled page reply: %v; %s", err, stderr.String())
		}
		if reply.Error != "" {
			input, _ := json.Marshal(op)
			t.Fatalf("controlled page: %s; input=%s", reply.Error, input)
		}
		return reply
	}
	var current pageReply
	var generated []command
	var history []string
	readUntil := func(done func(Wire) bool) Wire {
		t.Helper()
		_ = ws.SetReadDeadline(time.Now().Add(3 * time.Second))
		for {
			var raw json.RawMessage
			if err := ws.ReadJSON(&raw); err != nil {
				t.Fatalf("WS read: %v; frames=%v", err, history)
			}
			var frame Wire
			if err := json.Unmarshal(raw, &frame); err != nil {
				t.Fatal(err)
			}
			current = step(map[string]any{"op": "frame", "frame": raw})
			generated = append(generated, current.Sent...)
			history = append(history, fmt.Sprintf("%s epoch=%d", frame.Type, current.Epoch))
			if done(frame) {
				return frame
			}
		}
	}
	write := func(c command) {
		t.Helper()
		if err := ws.WriteJSON(c); err != nil {
			t.Fatal(err)
		}
	}
	readUntil(func(Wire) bool { return current.Status.Ready })
	if len(generated) != 0 {
		t.Fatalf("unexpected setup commands: %+v", generated)
	}
	const prompt = "Birch leaves shade the trail."
	typed := step(map[string]any{"op": "input", "text": prompt})
	if len(typed.Sent) != 1 || typed.Sent[0].Type != "pause" {
		t.Fatalf("typing commands=%+v", typed.Sent)
	}
	write(typed.Sent[0])
	readUntil(func(w Wire) bool { return w.Type == "state_changed" && w.State != nil && w.State.Paused })
	submitted := step(map[string]any{"op": "submit"})
	if len(submitted.Sent) != 2 {
		t.Fatalf("submit commands=%+v", submitted.Sent)
	}
	var request Wire
	for _, c := range submitted.Sent {
		before := current.Epoch
		write(c)
		switch c.Type {
		case "input":
			request = readUntil(func(w Wire) bool {
				data, _ := w.Data.(map[string]any)
				return w.Type == "speech_requested" && data["reason"] == "prompt"
			})
		case "unpause":
			if !c.Automatic {
				t.Fatal("submit unexpectedly used explicit Continue")
			}
			readUntil(func(w Wire) bool { return w.Type == "state_changed" && w.State != nil && w.State.Epoch > before })
		default:
			t.Fatalf("unexpected submit command: %+v", c)
		}
	}
	data, _ := request.Data.(map[string]any)
	id, _ := data["id"].(string)
	var claim command
	for _, c := range generated {
		if c.Type == "claim_speech" && c.SpeechID == id {
			claim = c
		}
	}
	if claim.Type == "" {
		t.Fatalf("page emitted no claim for request %+v; commands=%+v; frames=%v", data, generated, history)
	}
	generated = nil
	write(claim)
	result := readUntil(func(w Wire) bool { return w.Type == "speech_claimed" || w.Type == "error" })
	countPrompt := func() int {
		n := 0
		for _, e := range current.Channel.Events {
			if e.Event == "utterance" && e.Source == "human" && e.Text == prompt {
				n++
			}
		}
		return n
	}
	t.Logf("submit order=%s -> %s; request epoch=%v; claim epoch=%d; result=%s", submitted.Sent[0].Type, submitted.Sent[1].Type, data["epoch"], current.Epoch, result.Type)
	if result.Type != "speech_claimed" || countPrompt() != 1 {
		t.Fatalf("prompt lost before channel admission: count=%d result=%+v frames=%v", countPrompt(), result, history)
	}
	// A real duplicate claim is ignored. The subsequent real pause acknowledgement
	// is a deterministic barrier proving the duplicate has been processed.
	write(claim)
	if len(generated) != 1 || generated[0].Type != "pause" {
		t.Fatalf("speech pause commands=%+v", generated)
	}
	write(generated[0])
	readUntil(func(w Wire) bool { return w.Type == "state_changed" && w.State != nil && w.State.Paused })
	if countPrompt() != 1 {
		t.Fatalf("duplicate prompt admission: %d", countPrompt())
	}
}
