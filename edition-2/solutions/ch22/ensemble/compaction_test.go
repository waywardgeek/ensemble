package ensemble_test

import (
	"ensemble/ensemble"
	"ensemble/internal/common"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The peer is only an external provider; corpus, projection and runtime remain real.
func compactionWindowAgent(t *testing.T) (ensemble.Agent, ensemble.Config) {
	t.Helper()
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"reply","type":"message","role":"assistant","content":[{"type":"text","text":"done"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	t.Cleanup(peer.Close)
	cfg := ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-ch16-course", BaseURL: peer.URL, APIKey: "fake", DataDir: t.TempDir(), Memory: ensemble.MemorySettings{Session: ensemble.BandSettings{Budget: 1, HighWatermark: 1, LowWatermark: 1}}}
	a, err := ensemble.New(io.Discard).NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Shutdown() })
	return a, cfg
}
func appendCompactionFact(t *testing.T, a ensemble.Agent, e common.Event) {
	t.Helper()
	if err := a.History().Append(e); err != nil {
		t.Fatal(err)
	}
}
