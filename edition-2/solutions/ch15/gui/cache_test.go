package gui_test

import (
	"context"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUsageWireSession exercises the optional GUI through its real WebSocket.
// A restored response remains replayable without charging this process for it.
func TestUsageWireSession(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":20,"cache_read_input_tokens":80,"cache_creation_input_tokens":0,"output_tokens":2}}`)
	}))
	defer provider.Close()
	cfg := ensemble.Config{Model: "claude-sonnet-5", BaseURL: provider.URL, DataDir: t.TempDir(), DisableStreaming: true}
	root := ensemble.New(io.Discard)
	a, err := root.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Ask(context.Background(), "old session"); err != nil {
		t.Fatal(err)
	}
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}
	a.Shutdown()
	root = ensemble.New(io.Discard)
	a, err = root.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if err := a.Load(); err != nil {
		t.Fatal(err)
	}
	hub, err := gui.New(root, a, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Close()
	server := httptest.NewServer(hub.Handler("web/gui"))
	defer server.Close()
	c := dial(t, server.URL)
	frames := until(t, c, kind("usage"))
	initial := frames[len(frames)-1]
	if initial["input"] != float64(0) || initial["cost_usd"] != float64(0) {
		t.Fatalf("replayed lifetime charge: %v", initial)
	}
	for i := 1; i <= 2; i++ {
		send(t, c, "prompt", "new session")
		frames = until(t, c, kind("turn_ended"))
		var meter map[string]any
		for _, f := range frames {
			if f["type"] == "usage" {
				meter = f
			}
		}
		if meter == nil || meter["input"] != float64(20) || meter["cache_read"] != float64(80) || meter["hit_rate"] != .8 || meter["priced"] != true || meter["total"].(map[string]any)["output"] != float64(2*i) {
			t.Fatalf("usage frame: %v", meter)
		}
	}
}
