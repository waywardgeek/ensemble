package mcp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// The transport is the only MCP transport whose replies arrive as the body of
// the request that caused them. These tests pin the two body shapes, the
// no-body case, and the error paths, against a real HTTP server rather than a
// hand-rolled fake, so a wrong belief about SSE framing cannot be shared by
// the code and its test.

func newTestServer(t *testing.T, h http.HandlerFunc) (*httptest.Server, *HTTPTransport) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	tr, err := NewHTTPTransport(srv.URL, nil)
	if err != nil {
		t.Fatalf("NewHTTPTransport: %v", err)
	}
	t.Cleanup(func() { _ = tr.Close() })
	return srv, tr
}

// recvWithin fails rather than hanging when nothing arrives. Recv blocks by
// contract until a message or a close, so a bug that enqueues nothing turns a
// test into a ten-minute timeout instead of a failure. A mutation audit reads
// failures, not hangs.
func recvWithin(t *testing.T, tr *HTTPTransport, d time.Duration) json.RawMessage {
	t.Helper()
	type result struct {
		msg json.RawMessage
		err error
	}
	ch := make(chan result, 1)
	go func() {
		m, err := tr.Recv()
		ch <- result{m, err}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("Recv: %v", r.err)
		}
		return r.msg
	case <-time.After(d):
		t.Fatalf("Recv blocked for %s with nothing queued", d)
		return nil
	}
}

func TestHTTPTransportSSEBody(t *testing.T) {
	_, tr := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"ok\":true}}\n\n")
	})
	if err := tr.Send(json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"ping"}`)); err != nil {
		t.Fatalf("Send: %v", err)
	}
	got := recvWithin(t, tr, 2*time.Second)
	if !strings.Contains(string(got), `"ok":true`) {
		t.Fatalf("Recv returned %s, want the data frame payload", got)
	}
}

func TestHTTPTransportJSONBody(t *testing.T) {
	_, tr := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"jsonrpc":"2.0","id":7,"result":{"shape":"json"}}`)
	})
	if err := tr.Send(json.RawMessage(`{"jsonrpc":"2.0","id":7,"method":"ping"}`)); err != nil {
		t.Fatalf("Send: %v", err)
	}
	got := recvWithin(t, tr, 2*time.Second)
	if !strings.Contains(string(got), `"shape":"json"`) {
		t.Fatalf("Recv returned %s, want the JSON body", got)
	}
}

// A single POST may answer with several frames. All of them must reach Recv,
// in order — dropping the tail would lose a server-initiated request.
func TestHTTPTransportMultipleFrames(t *testing.T) {
	_, tr := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, ": a comment\ndata: {\"n\":1}\n\nevent: message\ndata: {\"n\":2}\n\n")
	})
	if err := tr.Send(json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"ping"}`)); err != nil {
		t.Fatalf("Send: %v", err)
	}
	for _, want := range []string{`{"n":1}`, `{"n":2}`} {
		got := recvWithin(t, tr, 2*time.Second)
		if string(got) != want {
			t.Fatalf("Recv = %s, want %s", got, want)
		}
	}
}

// A notification is answered with 202 and no body. Send must succeed and
// enqueue nothing; a transport that invented a reply would desynchronize the
// codec's id matching.
func TestHTTPTransportNotificationEnqueuesNothing(t *testing.T) {
	_, tr := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	if err := tr.Send(json.RawMessage(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)); err != nil {
		t.Fatalf("Send: %v", err)
	}
	// Close unblocks Recv; without it this would hang, which is the point —
	// nothing was queued.
	if err := tr.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := tr.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("Recv after notification = %v, want io.EOF", err)
	}
}

// A non-2xx reply must name the status AND carry the body. Quota and auth
// failures are reported in the body, and losing it leaves the agent unable to
// say why the web went away.
func TestHTTPTransportErrorStatusIncludesBody(t *testing.T) {
	_, tr := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		io.WriteString(w, `{"error":"daily limit reached"}`)
	})
	err := tr.Send(json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	if err == nil {
		t.Fatal("Send on 429 returned nil, want an error")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error %q does not name the status", err)
	}
	if !strings.Contains(err.Error(), "daily limit reached") {
		t.Errorf("error %q drops the body, which is where the reason lives", err)
	}
}

// A stateful server issues a session on initialize and expects it echoed on
// every later request.
func TestHTTPTransportEchoesSessionID(t *testing.T) {
	var mu sync.Mutex
	var seen []string
	_, tr := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.Header.Get("Mcp-Session-Id"))
		mu.Unlock()
		w.Header().Set("Mcp-Session-Id", "sess-42")
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"jsonrpc":"2.0","id":1,"result":{}}`)
	})
	for i := 0; i < 2; i++ {
		if err := tr.Send(json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"ping"}`)); err != nil {
			t.Fatalf("Send %d: %v", i, err)
		}
		recvWithin(t, tr, 2*time.Second)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 2 {
		t.Fatalf("server saw %d requests, want 2", len(seen))
	}
	if seen[0] != "" {
		t.Errorf("first request sent session %q, want none", seen[0])
	}
	if seen[1] != "sess-42" {
		t.Errorf("second request sent session %q, want %q", seen[1], "sess-42")
	}
}

// Close must wake a Recv that is ALREADY parked. This is the real shutdown
// path: the Codec's read loop sits in Recv for the life of the connection, so
// a Close that does not broadcast leaves that goroutine blocked forever and
// the agent never exits. Closing before calling Recv does not exercise it —
// that path returns through the closed check without ever waiting.
func TestHTTPTransportCloseWakesParkedRecv(t *testing.T) {
	_, tr := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	type result struct {
		msg json.RawMessage
		err error
	}
	ch := make(chan result, 1)
	go func() {
		m, err := tr.Recv()
		ch <- result{m, err}
	}()

	// Let the goroutine reach the cond.Wait before closing.
	waitUntil(t, func() bool {
		tr.mu.Lock()
		defer tr.mu.Unlock()
		return tr.parked > 0
	}, 2*time.Second, "Recv never parked")

	if err := tr.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	select {
	case r := <-ch:
		if !errors.Is(r.err, io.EOF) {
			t.Fatalf("parked Recv returned %v, want io.EOF", r.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not wake the parked Recv")
	}
}

func waitUntil(t *testing.T, cond func() bool, d time.Duration, msg string) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal(msg)
}

func TestHTTPTransportSendAfterCloseFails(t *testing.T) {
	_, tr := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	if err := tr.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := tr.Send(json.RawMessage(`{}`)); err == nil {
		t.Fatal("Send after Close returned nil, want an error")
	}
}

func TestHTTPTransportRejectsNonHTTPURL(t *testing.T) {
	if _, err := NewHTTPTransport("mcp.example.com/v2", nil); err == nil {
		t.Fatal("NewHTTPTransport accepted a schemeless URL, want an error")
	}
}
