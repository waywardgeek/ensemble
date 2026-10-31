package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// The server waits for real redirected stdout before sending the rest. A parser
// test alone cannot catch a terminal writer buffering otherwise-correct deltas.
// The first chunk deliberately resembles the human-hint acknowledgement label;
// model prose must stay intact on stdout regardless of its spelling.
func TestChatFlushesChunksAndSeparatesChannels(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "ensemble")
	build := exec.Command("go", "build", "-o", bin, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		send := func(data string) { fmt.Fprintf(w, "data: %s\n\n", data); w.(http.Flusher).Flush() }
		send(`{"type":"message_start","message":{"usage":{"input_tokens":2,"output_tokens":1}}}`)
		send(`{"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":""}}`)
		send(`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"reasoning-only"}}`)
		send(`{"type":"content_block_stop","index":0}`)
		send(`{"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}`)
		send(`{"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"hint:early"}}`)
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		send(`{"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"late"}}`)
		send(`{"type":"content_block_stop","index":1}`)
		send(`{"type":"message_delta","usage":{"output_tokens":4}}`)
		send(`{"type":"message_stop"}`)
	}))
	defer server.Close()
	cmd := exec.Command(bin, "chat")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL=claude-sonnet-4-6", "LLM_API_KEY=test", "LLM_BASE_URL="+server.URL, "EN_DISABLE_STREAMING=0")
	out := &notifyingWriter{seen: make(chan struct{})}
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = out, &stderr
	cmd.Stdin = strings.NewReader("hello\n")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-out.seen:
	case err := <-done:
		t.Fatal("chat ended before first chunk", err, stderr.String())
	case <-time.After(4 * time.Second):
		once.Do(func() { close(release) })
		cmd.Process.Kill()
		<-done
		t.Fatalf("chat withheld assistant chunk while response remained open: stdout=%q stderr=%q", out.buffer.String(), stderr.String())
	}
	once.Do(func() { close(release) })
	if err := <-done; err != nil {
		t.Fatal(err, stderr.String())
	}
	if out.buffer.String() != "hint:earlylate\n" {
		t.Fatal("reply duplicated or contaminated", out.buffer.String())
	}
	if !strings.Contains(stderr.String(), "reasoning-only") {
		t.Fatal("thinking missing from stderr")
	}
	if strings.Contains(out.buffer.String()+stderr.String(), "\x1b[") {
		t.Fatal("ANSI escaped into redirected output")
	}
}

type notifyingWriter struct {
	buffer bytes.Buffer
	seen   chan struct{}
	once   sync.Once
}

// Write acknowledges bytes only when the subprocess pipe actually delivers them.
func (w *notifyingWriter) Write(p []byte) (int, error) {
	n, err := w.buffer.Write(p)
	if strings.Contains(w.buffer.String(), "early") {
		w.once.Do(func() { close(w.seen) })
	}
	return n, err
}

// The delivered CLI must execute the owner's selection operation; merely
// printing a confirmation would leave the next provider request unchanged.
func TestChatModelSelection(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "ensemble")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
	requested := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			// Model records routing at the actual external boundary.
			Model string
		}
		json.NewDecoder(r.Body).Decode(&body)
		requested <- body.Model
		fmt.Fprint(w, `{"content":[{"type":"text","text":"selected"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "chat")
	cmd.Dir = t.TempDir()
	// A minimal child environment prevents unrelated provider credentials from
	// entering a fake-only process or loading a user's private auth profile.
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + t.TempDir(), "EN_HOME=" + t.TempDir(), "LLM_VENDOR=anthropic", "LLM_MODEL=claude-opus-5", "LLM_API_KEY=fake", "LLM_BASE_URL=" + server.URL, "EN_DISABLE_STREAMING=1"}
	cmd.Stdin = strings.NewReader("/model claude-sonnet-5\nhello\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(out))
	}
	select {
	case model := <-requested:
		if model != "claude-sonnet-5" {
			t.Fatal("CLI model control did not reach provider", model)
		}
	default:
		t.Fatal("CLI never requested an answer")
	}
	if !strings.Contains(string(out), "selected") {
		t.Fatal("CLI answer missing", string(out))
	}
}
