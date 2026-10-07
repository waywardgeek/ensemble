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
	events := client.Observations()
	if len(events) != 3 {
		t.Fatalf("events=%d", len(events))
	}
	for i, event := range events {
		if event.AgentID != a.ID() || event.Seq != uint64(i+1) {
			t.Fatal("wrong Agent or event order")
		}
	}
	events[0].Event.Message.Parts[0].Text = new(string)
	if *client.Observations()[0].Event.Message.Parts[0].Text != "hi" {
		t.Fatal("snapshot aliased")
	}
	client.Close()
	if _, err = client.Submit(context.Background(), ensemble.ClientRequest{Prompt: &prompt}); err == nil || !strings.Contains(diagnostics.String(), "GUI client closed") {
		t.Fatal("closed-client diagnostic did not reach Ensemble logger")
	}
	if _, err = a.Ask(context.Background(), "after close"); err != nil {
		t.Fatal(err)
	}
	if len(client.Observations()) != 3 {
		t.Fatal("closed client received later event")
	}
}
