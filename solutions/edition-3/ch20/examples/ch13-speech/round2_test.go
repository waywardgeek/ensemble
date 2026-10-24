package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestNormalOneShotErrorObservation(t *testing.T) {
	d := driverPage(t, "error")
	path := filepath.Join(t.TempDir(), "speech.jsonl")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runTestSession(ctx, d.page, "question", path, "transcript", "", false, false, t.TempDir()); err != nil {
		t.Fatalf("completed error observation failed: %v", err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	heard, outcome := false, false
	for {
		var entry map[string]any
		err = decoder.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if entry["event"] == "utterance" && entry["source"] == "error" && entry["text"] == "Error: target rejected request" {
			heard = true
		}
		if entry["kind"] == "target_outcome" && entry["status"] == "error" {
			outcome = true
		}
	}
	if !heard || !outcome {
		t.Fatalf("actual page reply/error outcome not captured: heard=%v outcome=%v", heard, outcome)
	}
}

func peerPipe(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() { os.Stdin = previous; reader.Close(); writer.Close() })
	return reader, writer
}
func expectPromptPeerExit(t *testing.T, done <-chan error, cancel context.CancelFunc) {
	t.Helper()
	select {
	case err := <-done:
		if err == nil {
			t.Error("peer closure concealed interrupted activity")
		}
	case <-time.After(400 * time.Millisecond):
		t.Error("EOF did not promptly stop active peer work")
		cancel()
		select {
		case <-done:
		case <-time.After(4 * time.Second):
			t.Fatal("session did not finish after test cancellation")
		}
	}
}
func TestJSONLPeerEOFDuringOperation(t *testing.T) {
	for _, op := range []string{"sleep", "wait"} {
		t.Run(op, func(t *testing.T) {
			_, writer := peerPipe(t)
			outcome := "success"
			command := `{"op":"sleep","ms":30000}`
			if op == "wait" {
				outcome = "pending"
				command = `{"op":"wait","timeout_ms":30000}`
			}
			d := driverPage(t, outcome)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				done <- runTestSession(ctx, d.page, "question", filepath.Join(t.TempDir(), "speech.jsonl"), "transcript", "", true, false, t.TempDir())
			}()
			if _, err := fmt.Fprintln(writer, command); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				t.Fatalf("interactive operation ended while input remained open: %v", err)
			case <-time.After(100 * time.Millisecond):
			}
			writer.Close()
			expectPromptPeerExit(t, done, cancel)
		})
	}
}
func TestListenerPeerEOFDuringAsk(t *testing.T) {
	_, writer := peerPipe(t)
	started := make(chan struct{}, 1)
	stopped := make(chan struct{}, 1)
	// Only a blocked loopback HTTP transport: no response/model/text fixture.
	transport := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Errorf("unexpected listener setup request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "only the pending Ask transport is allowed", http.StatusBadRequest)
			return
		}
		var body struct {
			Tools []struct {
				Function struct {
					Name       string                     `json:"name"`
					Parameters map[string]json.RawMessage `json:"parameters"`
				} `json:"function"`
			} `json:"tools"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		names := []string{}
		for _, tool := range body.Tools {
			names = append(names, tool.Function.Name)
			if tool.Function.Name == "submit" {
				if string(tool.Function.Parameters["properties"]) != "{}" || string(tool.Function.Parameters["additionalProperties"]) != "false" || len(tool.Function.Parameters["required"]) != 0 {
					t.Errorf("listener submit schema was replaced: %+v", tool.Function.Parameters)
				}
			}
		}
		sort.Strings(names)
		if !reflect.DeepEqual(names, []string{"hear", "sleep", "submit", "type", "wait"}) {
			t.Errorf("listener declarations = %v", names)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		started <- struct{}{}
		<-r.Context().Done()
		stopped <- struct{}{}
	}))
	defer transport.Close()
	for key, value := range map[string]string{"ENSEMBLE_LISTENER_MODEL": "local-test", "ENSEMBLE_LISTENER_VENDOR": "openai", "ENSEMBLE_LISTENER_BASE_URL": transport.URL, "ENSEMBLE_LISTENER_API_KEY": "not-real", "ENSEMBLE_LISTENER_GOAL": "wait for speech"} {
		t.Setenv(key, value)
	}
	d := driverPage(t, "success")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- runTestSession(ctx, d.page, "question", filepath.Join(t.TempDir(), "speech.jsonl"), "transcript", "", false, true, t.TempDir())
	}()
	select {
	case <-started:
	case err := <-done:
		t.Fatalf("listener setup failed before pending Ask: %v", err)
	case <-time.After(time.Second):
		t.Fatal("listener did not reach test transport")
	}
	writer.Close()
	expectPromptPeerExit(t, done, cancel)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Error("listener request remained active after peer shutdown")
	}
}

func TestNormalClosedStdinStillObserves(t *testing.T) {
	_, writer := peerPipe(t)
	writer.Close()
	d := driverPage(t, "success")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runTestSession(ctx, d.page, "question", filepath.Join(t.TempDir(), "speech.jsonl"), "transcript", "", false, false, t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestNormalCaptureFailureStillFails(t *testing.T) {
	d := driverPage(t, "capture-fails")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runTestSession(ctx, d.page, "question", filepath.Join(t.TempDir(), "speech.jsonl"), "transcript", "", false, false, t.TempDir()); err == nil {
		t.Fatalf("capture failure: %v", err)
	}
}

func runTestSession(ctx context.Context, p *page, prompt, logPath, mode, typeDuring string, jsonl, listen bool, dir string) (err error) {
	var peer *peerInput
	if jsonl || listen {
		peer, err = watchPeer(ctx, os.Stdin, jsonl)
		if err != nil {
			return err
		}
		ctx = peer.ctx
		defer func() { err = errors.Join(err, peer.close()) }()
	}
	return runSession(ctx, p, prompt, logPath, mode, typeDuring, jsonl, listen, dir, peer)
}

func TestPeerOperationsWhileOpen(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	peer, err := watchPeer(context.Background(), reader, true)
	if err != nil {
		t.Fatal(err)
	}
	defer peer.close()
	d := driverPage(t, "success")
	started := time.Now()
	if _, err = d.act(peer.ctx, operation{Op: "sleep", MS: 30}); err != nil {
		t.Fatal(err)
	}
	if time.Since(started) < 20*time.Millisecond || peer.ctx.Err() != nil {
		t.Fatal("finite sleep or open peer lifetime did not survive")
	}
}
func TestPeerEOFDuringPageCall(t *testing.T) {
	started := make(chan struct{})
	upgrade := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		if _, _, err = conn.ReadMessage(); err != nil {
			return
		}
		close(started)
		conn.ReadMessage()
	}))
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	peer, err := watchPeer(ctx, reader, true)
	if err != nil {
		t.Fatal(err)
	}
	defer peer.close()
	done := make(chan error, 1)
	go func() { done <- (&page{conn: conn}).call(peer.ctx, "Runtime.evaluate", nil, nil) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("page call did not reach blocked transport")
	}
	writer.Close()
	expectPromptPeerExit(t, done, cancel)
}
func TestTargetOutcomeWriteFailure(t *testing.T) {
	d := driverPage(t, "error")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := d.act(ctx, operation{Op: "type", Text: "question"}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.act(ctx, operation{Op: "submit"}); err != nil {
		t.Fatal(err)
	}
	d.speech = json.NewEncoder(failedSpeechWriter{})
	if err := d.observeOneShot(ctx); err == nil {
		t.Fatal("normal observation swallowed outcome capture failure")
	}
}
func TestPeerProcessHelper(t *testing.T) {
	if os.Getenv("CH13_PEER_HELPER") != "close-open-stdin" {
		t.Skip("owned subprocess helper")
	}
	peer, err := watchPeer(context.Background(), os.Stdin, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("reader ready")
	time.Sleep(100 * time.Millisecond)
	if err := peer.close(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("reader closed")
}
func TestPeerStopsReaderWithOpenProcessStdin(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPeerProcessHelper$")
	cmd.Env = append(os.Environ(), "CH13_PEER_HELPER=close-open-stdin")
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	lines := bufio.NewScanner(output)
	ready, closed := false, false
	for lines.Scan() {
		if lines.Text() == "reader ready" {
			ready = true
		}
		if lines.Text() == "reader closed" {
			closed = true
		}
	}
	err = cmd.Wait()
	if err != nil || !ready || !closed {
		t.Fatalf("owned stdin reader shutdown: ready=%v closed=%v err=%v stderr=%s", ready, closed, err, stderr.String())
	}
}
