package ensemble_test

import (
	"context"
	"ensemble/ensemble"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestPausedCancellation releases a caller-cancelled gate without an explicit
// unpause. Request completion must stay reliable even with no browser attached.
func TestPausedCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"content":[{"type":"tool_use","id":"paused-call","name":"run_command","input":{"command":"touch forbidden"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	a := actorAgent(t, server, func(cfg *ensemble.Config) { cfg.BuiltinTools = true })
	a.Pause()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reply := a.Submit(ctx, "pause before execution")
	watchCtx, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()
	if _, err := a.Wait(watchCtx, func(o ensemble.Observation) bool { _, ok := o.(ensemble.PartFinal); return ok }); err != nil {
		t.Fatal(err)
	}
	cancel()
	// Cleanup releases a broken waiter too, so mutation failure cannot hang tests.
	defer a.Unpause()
	if result := receive(t, reply); !errors.Is(result.Err, context.Canceled) {
		t.Fatal("cancelled paused request did not complete", result.Err)
	}
}
