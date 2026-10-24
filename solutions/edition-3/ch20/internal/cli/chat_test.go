package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

// Retain the first completion after the real actor has gone idle. The adapter
// must classify the next line at the actor, not from its still-pending reply.
type bufferedChatAgent struct {
	ensemble.Agent
	ctx      context.Context
	finished chan struct{}
	release  chan struct{}
	once     sync.Once
	calls    atomic.Int32
}

func (a *bufferedChatAgent) gate(reply <-chan ensemble.Result) <-chan ensemble.Result {
	if a.calls.Add(1) != 1 {
		a.once.Do(func() { close(a.release) })
		return reply
	}
	held := make(chan ensemble.Result, 1)
	go func() {
		r := <-reply
		close(a.finished)
		select {
		case <-a.release:
		case <-a.ctx.Done():
		}
		held <- r
	}()
	return held
}
func (a *bufferedChatAgent) Submit(ctx context.Context, s string) <-chan ensemble.Result {
	return a.gate(a.Agent.Submit(ctx, s))
}
func (a *bufferedChatAgent) SubmitLive(ctx context.Context, s string) <-chan ensemble.Result {
	return a.gate(a.Agent.SubmitLive(ctx, s))
}
func TestChatBufferedReplyDoesNotStrandInput(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := requests.Add(1)
		b, _ := io.ReadAll(r.Body)
		if n == 2 && !bytes.Contains(b, []byte("second prompt")) {
			t.Errorf("second prompt missing: %s", b)
		}
		text := "first answer"
		if n == 2 {
			text = "second answer"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"model": "offline", "content": []any{map[string]string{"type": "text", "text": text}}, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	dir := t.TempDir()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "offline", APIKey: "dummy", BaseURL: server.URL, BaseDir: dir, DataDir: filepath.Join(dir, "jobs")})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wrapped := &bufferedChatAgent{Agent: a, ctx: ctx, finished: make(chan struct{}), release: make(chan struct{})}
	in, writer := io.Pipe()
	defer in.Close()
	defer writer.Close()
	go func() {
		_, _ = io.WriteString(writer, "first prompt\n")
		select {
		case <-wrapped.finished:
		case <-ctx.Done():
			return
		}
		_, _ = io.WriteString(writer, "second prompt\n")
		_ = writer.Close()
	}()
	var out, diagnostics bytes.Buffer
	if err := runChat(ctx, wrapped, in, &out, &diagnostics); err != nil {
		t.Fatalf("%v; output=%s diagnostics=%s", err, out.String(), diagnostics.String())
	}
	if requests.Load() != 2 || strings.Count(out.String(), "first answer") != 1 || strings.Count(out.String(), "second answer") != 1 {
		t.Fatalf("EOF did not drain both private turns: requests=%d output=%s", requests.Load(), out.String())
	}
}
