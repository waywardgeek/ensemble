package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ensemble/ensemble"
)

// This fake is the external provider: the real root, Agent and Engine all run.
// Reflecting received messages makes cross-Agent history leaks observable.
func TestAgentsKeepTheirOwnersAndState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Error("wrong Anthropic API version")
		}
		var request struct {
			Model    string
			Messages []struct {
				Role    string
				Content []struct{ Type, Text string }
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		var text []string
		for _, m := range request.Messages {
			if m.Role == "user" {
				for _, part := range m.Content {
					text = append(text, part.Text)
				}
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]string{{"type": "text", "text": request.Model + ":" + strings.Join(text, "/")}},
			"usage":   map[string]int{"input_tokens": len(request.Messages), "output_tokens": 2},
		})
	}))
	defer server.Close()
	app := ensemble.New(io.Discard)
	a, err := app.NewAgent(ensemble.Config{DataDir: t.TempDir(), BaseURL: server.URL, APIKey: "fake-key", Model: "first"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := app.NewAgent(ensemble.Config{DataDir: t.TempDir(), BaseURL: server.URL, APIKey: "fake-key", Model: "second"})
	if err != nil {
		t.Fatal(err)
	}
	for _, child := range []ensemble.Agent{a, b} {
		if child.Ensemble() != app || child.Engine().Agent() != child || child.History().Agent() != child {
			t.Fatal("broken parent relationship")
		}
	}
	for _, turn := range []struct {
		agent          ensemble.Agent
		question, want string
	}{
		{a, "red", "first:red"}, {b, "blue", "second:blue"}, {a, "green", "first:red/green"},
	} {
		got, err := turn.agent.Ask(context.Background(), turn.question)
		if err != nil || got != turn.want {
			t.Fatalf("answer = %q, %v; want %q", got, err, turn.want)
		}
	}
	if a.Engine().Usage() != (ensemble.Usage{Input: 4, Output: 4}) || b.Engine().Usage() != (ensemble.Usage{Input: 1, Output: 2}) {
		t.Fatal("usage was lost or shared between Agents")
	}
}

func TestFailedExchangeIsLoggedWithoutAResponse(t *testing.T) {
	// These are ordinary Chapter 1 failures, not tool or streaming fixtures.
	for _, failure := range []struct {
		name   string
		status int
		body   string
	}{
		{"HTTP", 429, `{"content":[{"type":"text","text":"fake-key"}],"usage":{"input_tokens":1,"output_tokens":2}}`},
		{"JSON", 200, `not JSON fake-key`},
		{"missing usage", 200, `{"content":[{"type":"text","text":"answer"}]}`},
		{"empty answer", 200, `{"content":[],"usage":{"input_tokens":0,"output_tokens":0}}`},
	} {
		t.Run(failure.name, func(t *testing.T) {
			// Chapter 2 records failed attempts honestly; only a valid reply becomes
			// agent dialogue. Consecutive human messages merge on the Anthropic wire.
			// The fake then becomes healthy so we exercise actual subsequent use.
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if calls == 1 {
					w.WriteHeader(failure.status)
					_, _ = io.WriteString(w, failure.body)
					return
				}
				var request struct {
					Messages []struct {
						Role    string
						Content []struct{ Type, Text string }
					}
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
				}
				if len(request.Messages) != 1 || len(request.Messages[0].Content) != 2 || request.Messages[0].Content[1].Text != "try again" {
					t.Error("failed attempt was lost or became an assistant response")
				}
				_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":2}}`)
			}))
			defer server.Close()
			// Observe diagnostics at their real owner, without replacing the logger.
			var logs bytes.Buffer
			app := ensemble.New(&logs)
			a, err := app.NewAgent(ensemble.Config{DataDir: t.TempDir(), BaseURL: server.URL, APIKey: "fake-key", Model: "fake"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = a.Ask(context.Background(), "first"); err == nil {
				t.Fatal("failed response accepted")
			}
			if calls != 1 {
				t.Fatalf("failed request retried: %d calls", calls)
			}
			if logs.Len() == 0 {
				t.Fatal("failure did not reach root logger")
			}
			if strings.Contains(logs.String(), "fake-key") || strings.Contains(err.Error(), "fake-key") {
				t.Fatal("credential echoed in diagnostic")
			}
			got, err := a.Ask(context.Background(), "try again")
			if err != nil || got != "ok" {
				t.Fatalf("next turn = %q, %v", got, err)
			}
		})
	}
}
