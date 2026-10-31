package ensemble_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"ensemble/ensemble"
)

// TestMCPSlowEphemera waits beyond the real three-second callback boundary.
// Both modes must supply completed context before any provider request exists.
func TestMCPSlowEphemera(t *testing.T) {
	for _, name := range []string{"startup", "snapshot"} {
		t.Run(name, func(t *testing.T) {
			requests := make(chan string, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				requests <- string(body)
				io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
			}))
			defer server.Close()
			a := mcpAgent(t, server.URL)
			p := peer(t, a, "test", name)
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			answer := a.Submit(ctx, "use completed automatic context")
			call := receiveRPC(t, p.requests)
			body := ""
			select {
			case body = <-requests:
				t.Error("provider request preceded automatic completion")
			case <-time.After(3400 * time.Millisecond):
			}
			// Releasing the actual peer, rather than a fake job, lets the handler,
			// job output, collector, history and renderer finish in their real order.
			p.reply(call["id"], map[string]any{"content": []map[string]string{{"type": "text", "text": "COMPLETED_AUTO_CONTEXT"}}})
			select {
			case result := <-answer:
				if result.Err != nil || result.Text != "done" {
					t.Fatalf("turn result: %+v", result)
				}
			case <-ctx.Done():
				t.Fatal("completed ephemera did not release request")
			}
			if body == "" {
				body = <-requests
			}
			if !strings.Contains(body, "COMPLETED_AUTO_CONTEXT") || strings.Contains(body, "Use wait_for_job") {
				t.Errorf("completed automatic context absent; running notice delivered: %s", body)
			}
		})
	}
}

// TestMCPEphemeraInterrupt preserves execution while abandoning the old turn's
// collection. A new turn must work before the old peer response is released.
func TestMCPEphemeraInterrupt(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests <- string(body)
		io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	a := mcpAgent(t, server.URL)
	p := peer(t, a, "test", "startup")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	first := a.Submit(ctx, "old turn")
	oldCall := receiveRPC(t, p.requests)
	var handle int
	for _, event := range a.History().EventsAfter(0) {
		if event.Type == "tool_called" && event.Tool.Name == "startup" {
			handle = event.Tool.Job.Handle
		}
	}
	job, err := a.Jobs().Find(handle)
	if err != nil {
		t.Fatal(err)
	}
	a.Post(ensemble.Interrupt{})
	select {
	case result := <-first:
		if !errors.Is(result.Err, ensemble.ErrInterrupted) {
			t.Fatalf("interrupt result: %+v", result)
		}
	case <-ctx.Done():
		t.Fatal("interrupt remained blocked on ephemera")
	}
	second := a.Submit(ctx, "new turn")
	newCall := receiveRPC(t, p.requests)
	if string(newCall["method"]) != `"tools/call"` {
		t.Fatalf("turn interrupt canceled continuing RPC: %s", newCall)
	}
	p.reply(newCall["id"], map[string]any{"content": []map[string]string{{"type": "text", "text": "NEW_CONTEXT"}}})
	select {
	case result := <-second:
		if result.Err != nil {
			t.Fatal(result.Err)
		}
	case <-ctx.Done():
		t.Fatal("old ephemera blocked the new turn")
	}
	if job.Context().Err() != nil || job.Data().Status != "running" {
		t.Fatal("turn interrupt stopped the automatic job")
	}
	body := <-requests
	if !strings.Contains(body, "NEW_CONTEXT") || strings.Contains(body, "OLD_CONTEXT") {
		t.Fatal("new turn did not use its own automatic context")
	}
	p.reply(oldCall["id"], map[string]any{"content": []map[string]string{{"type": "text", "text": "OLD_CONTEXT"}}})
	select {
	case <-job.Context().Done():
	case <-ctx.Done():
		t.Fatal("late automatic job did not finish")
	}
	output, err := os.ReadFile(job.Data().Output.Locator)
	if err != nil || string(output) != "OLD_CONTEXT" {
		t.Fatalf("late output was lost: %q %v", output, err)
	}
	projection, _ := json.Marshal(a.History().Context())
	if strings.Contains(string(projection), "OLD_CONTEXT") {
		t.Fatal("late automatic result entered a newer conversation")
	}
}
