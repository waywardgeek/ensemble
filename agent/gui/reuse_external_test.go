// This test lives in package gui_test and uses only exported API: the root
// agent package and gui itself. It is the import path an application outside
// the ensemble binary follows (the reason the package left internal/),
// exercised in-tree so a change that re-privatizes the surface fails loudly.
// The gorilla import below is test-side plumbing for a websocket client; the
// headless-framework property concerns the non-test dependency closure.
package gui_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	agent "github.com/waywardgeek/ensemble/agent"
	"github.com/waywardgeek/ensemble/agent/gui"
)

func TestGUIReusableOutsideEnsemble(t *testing.T) {
	inbound := make(chan agent.Inbound, 1)
	srv := gui.New(gui.AgentHooks{
		Send:     func(m agent.Inbound) { inbound <- m },
		EventLog: &agent.Log{Clock: time.Now},
	}, "")

	mux := http.NewServeMux()
	mux.Handle("/", gui.StaticHandler(""))
	mux.HandleFunc("/ws", srv.ServeWS)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// The embedded assets serve with no ensemble binary and no disk layout.
	resp, err := http.Get(ts.URL + "/artifact-scroll.js")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "ArtifactScroll") {
		t.Fatalf("artifact-scroll.js: status %d, marker found: %v",
			resp.StatusCode, strings.Contains(string(body), "ArtifactScroll"))
	}

	// Single components mount from the exported FS without the shell.
	if _, err := gui.Assets.ReadFile("web/renderers.js"); err != nil {
		t.Fatalf("renderers.js not in exported assets: %v", err)
	}

	// A browser prompt reaches the agent through the one hook the consumer
	// supplied. This is the whole agent-facing contract: no engine, no actor,
	// no ensemble wiring.
	wsURL := strings.Replace(ts.URL, "http", "ws", 1) + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.WriteJSON(map[string]string{"type": "prompt", "text": "hello from outside"}); err != nil {
		t.Fatal(err)
	}
	select {
	case m := <-inbound:
		um, ok := m.(agent.UserMessage)
		if !ok || um.Text != "hello from outside" {
			t.Fatalf("Send hook got %#v, want UserMessage{hello from outside}", m)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("prompt never reached the Send hook")
	}
}
