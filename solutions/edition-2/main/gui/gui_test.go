package gui_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	gui "ensemble-gui-stub"
	"ensemble/ensemble"
)

// This builds outside the core module using only public values. The HTTP fake
// stands in for the provider; the GUI stub and ordinary client share one Agent.
func TestPublicAgentIsShared(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"content":[{"type":"text","text":"hello"}],"usage":{"input_tokens":1,"output_tokens":2}}`)
	}))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "fake", APIKey: "fake", BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if answer, err := gui.Submit(context.Background(), a, "from the GUI"); err != nil || answer != "hello" {
		t.Fatalf("GUI reply %q: %v", answer, err)
	}
	if _, err := a.Ask(context.Background(), "from the CLI"); err != nil {
		t.Fatal(err)
	}
	if len(a.History().Context().Dialogue) != 4 || a.Engine().Usage().Output != 4 {
		t.Fatal("clients did not share Agent state")
	}
}
