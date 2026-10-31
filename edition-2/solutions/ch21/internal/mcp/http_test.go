package mcp_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
	"ensemble/internal/mcp"
)

// Close must wake a receiver even when no response was ever queued. The actual
// owner exists; only the HTTP peer is replaced. Bound the observable, not just
// the outer test process, so a missed wakeup reports a useful failure.
func TestHTTPCloseWakesReceiver(t *testing.T) {
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "fake", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	transport := mcp.NewHTTP(a.MCP(), common.MCPServer{URL: "http://unused"})
	defer transport.Close()
	received := make(chan error, 1)
	go func() { _, err := transport.Recv(); received <- err }()
	// Done belongs to client.close and closes before transport.Close. Observe
	// the actual receiver parked in Cond.Wait instead of mistaking Done for
	// reader completion. This bound fails with the wait state in the diagnostic.
	deadline := time.Now().Add(time.Second)
	for {
		stack := make([]byte, 64*1024)
		stack = stack[:runtime.Stack(stack, true)]
		parked := false
		for _, goroutine := range strings.Split(string(stack), "\n\n") {
			if strings.Contains(goroutine, "httpTransport).Recv") && strings.Contains(goroutine, "sync.(*Cond).Wait") {
				parked = true
			}
		}
		if parked {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("receiver never parked in Cond.Wait")
		}
		runtime.Gosched()
	}
	transport.Close()
	select {
	case err := <-received:
		if err == nil {
			t.Fatal("close returned a message instead of failure")
		}
	case <-time.After(time.Second):
		t.Fatal("close did not wake receiver")
	}
	if err := transport.Send(json.RawMessage(`{"jsonrpc":"2.0","method":"test"}`)); err == nil {
		t.Fatal("send after close succeeded")
	}
}

// A redirect is outside the configured endpoint contract and must not receive
// the Bearer credential, even on the same host where net/http would forward it.
func TestHTTPRedirectDoesNotForwardCredential(t *testing.T) {
	t.Setenv("ENSEMBLE_REDIRECT_KEY", "test-credential")
	followed := make(chan struct{}, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { followed <- struct{}{}; w.WriteHeader(202) }))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "fake", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	transport := mcp.NewHTTP(a.MCP(), common.MCPServer{URL: source.URL, AuthEnv: "ENSEMBLE_REDIRECT_KEY"})
	defer transport.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := transport.SendContext(ctx, json.RawMessage(`{"jsonrpc":"2.0","method":"test"}`)); err == nil {
		t.Error("redirect accepted")
	}
	select {
	case <-followed:
		t.Fatal("credential endpoint redirect followed")
	default:
	}
}
