package gui_test

import (
	"bytes"
	"context"
	"example.com/ensemble"
	"example.com/ensemble-gui"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPublicClientBoundary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"content":[{"type":"text","text":"hello"}],"usage":{"input_tokens":5,"output_tokens":2}}`)
	}))
	defer server.Close()
	var diagnostics bytes.Buffer
	app := ensemble.New(&diagnostics)
	a, err := app.NewAgent(ensemble.Config{APIKey: "test", Model: "test", BaseURL: server.URL, LogPath: filepath.Join(t.TempDir(), "a.log")})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := app.NewAgent(ensemble.Config{APIKey: "test", Model: "test", BaseURL: server.URL, LogPath: filepath.Join(t.TempDir(), "b.log")})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	client, err := gui.New(app, a.ID())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	prompt := "hi"
	result, err := client.Submit(context.Background(), ensemble.ClientRequest{Prompt: &prompt})
	if err != nil || result.Text != "hello" {
		t.Fatalf("result %v / %v", result, err)
	}
	if _, err = b.Ask(context.Background(), "separate"); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(3 * time.Second)
	var events []ensemble.Observation
	for {
		events = client.Observations()
		ended := false
		for _, o := range events {
			ended = ended || o.Kind == "turn_ended"
		}
		if ended {
			break
		}
		select {
		case <-client.Updates():
		case <-deadline:
			t.Fatal("GUI did not receive durable turn end")
		}
	}
	durable := []ensemble.Observation{}
	for _, event := range events {
		if event.AgentID != a.ID() {
			t.Fatal("wrong Agent")
		}
		if event.Event.Seq != 0 {
			durable = append(durable, event)
		}
	}
	if len(durable) != 5 {
		t.Fatalf("durable events=%d", len(durable))
	}
	for i, event := range durable {
		if event.Seq != uint64(i+1) {
			t.Fatal("wrong durable order")
		}
	}
	for _, event := range events {
		if event.Event.Message != nil {
			*event.Event.Message.Parts[0].Text = "mutated"
		}
	}
	for _, event := range client.Observations() {
		if event.Event.Message != nil && *event.Event.Message.Parts[0].Text != "hi" {
			t.Fatal("snapshot aliased")
		}
	}
	client.Close()
	if _, err = client.Submit(context.Background(), ensemble.ClientRequest{Prompt: &prompt}); err == nil || !strings.Contains(diagnostics.String(), "GUI client closed") {
		t.Fatal("closed-client diagnostic did not reach Ensemble logger")
	}
	if _, err = a.Ask(context.Background(), "after close"); err != nil {
		t.Fatal(err)
	}
	if len(client.Observations()) != len(events) {
		t.Fatal("closed client received later event")
	}
}
