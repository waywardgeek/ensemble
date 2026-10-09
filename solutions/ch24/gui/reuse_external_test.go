// This test lives in package gui_test and uses only exported API: the root
// agent package and gui itself. It is the import path an application outside
// the ensemble binary follows (the reason the package left internal/),
// exercised in-tree so a change that re-privatizes the surface fails loudly.
// The gorilla import below is test-side plumbing for a websocket client; the
// headless-framework property concerns the non-test dependency closure.
package gui_test

import (
	"bufio"
	"io"
	"net"
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
	// A sentinel event planted in a consumer-built log. The vocabulary
	// (Event, MessageData, TextPart) is root-aliased precisely so an
	// application outside the module can do this.
	const sentinel = "replay sentinel 24"
	log := &agent.Log{
		Events: []agent.Event{{
			Seq:  1,
			Type: agent.MessageReceived,
			Time: time.Unix(0, 0).UTC(),
			Message: &agent.MessageData{
				Actor: agent.ActorAgent,
				Parts: agent.PartList{agent.TextPart{Text: sentinel}},
			},
		}},
		Next:  2,
		Clock: time.Now,
	}

	inbound := make(chan agent.Inbound, 1)
	srv := gui.New(gui.AgentHooks{
		Send:     func(m agent.Inbound) { inbound <- m },
		EventLog: log,
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

	// The agent-eyes relay is part of the public surface too: an embedding
	// application can let an agent see and drive this GUI by exposing the
	// same MCP port the ensemble binary offers. This leg runs BEFORE any
	// websocket client connects: with no browser attached the relay
	// answers a well-formed JSON-RPC error immediately, whereas a connected
	// client that ignores MCP frames would leave the request waiting on a
	// browser reply that never comes.
	mln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer mln.Close()
	go srv.ServeMCP(mln)
	mcp, err := net.Dial("tcp", mln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer mcp.Close()
	_ = mcp.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := mcp.Write([]byte(`{"jsonrpc":"2.0","id":7,"method":"tools/list"}` + "\n")); err != nil {
		t.Fatal(err)
	}
	reply, err := bufio.NewReader(mcp).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, `"jsonrpc"`) || !strings.Contains(reply, `"id":7`) {
		t.Fatalf("MCP relay reply malformed: %q", reply)
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

	// The consumer's own event reaches a browser through the standard
	// subscribe-and-replay path: the GUI renders content the ensemble
	// binary never produced.
	if err := conn.WriteJSON(map[string]string{"type": "subscribe"}); err != nil {
		t.Fatal(err)
	}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		var frame struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := conn.ReadJSON(&frame); err != nil {
			t.Fatalf("replay: connection ended before the sentinel arrived: %v", err)
		}
		if frame.Type == "message" && strings.Contains(frame.Text, sentinel) {
			break
		}
	}

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
