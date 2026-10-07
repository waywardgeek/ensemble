package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/ensemble"
)

func TestIndependentAgentsAndFailureAtomicity(t *testing.T) {
	var requests [][]ensemble.Message
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Messages []ensemble.Message `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		requests = append(requests, request.Messages)
		if request.Messages[len(request.Messages)-1].Content == "reject" {
			fmt.Fprint(w, `{"content":[{"type":"text","text":"unaccounted"}]}`)
			return
		}
		fmt.Fprint(w, `{"content":[{"type":"text","text":"hel"},{"type":"thinking","thinking":"private"},{"type":"text","text":"lo"}],"usage":{"input_tokens":7,"output_tokens":3}}`)
	}))
	defer server.Close()
	var diagnostics bytes.Buffer
	app := ensemble.New(&diagnostics)
	config := ensemble.Config{APIKey: "test", Model: "test", BaseURL: server.URL + "///"}
	a, err := app.NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	b, err := app.NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	for _, turn := range []struct {
		agent    *ensemble.Agent
		question string
	}{{a, "alpha"}, {b, "beta"}, {a, "reject"}, {a, "again"}} {
		answer, err := turn.agent.Ask(context.Background(), turn.question)
		if turn.question == "reject" {
			if err == nil {
				t.Fatal("accepted missing usage")
			}
			continue
		}
		if err != nil || answer != "hello" {
			t.Fatalf("answer %q, error %v", answer, err)
		}
	}
	if len(requests) != 4 || len(requests[1]) != 1 || requests[1][0].Content != "beta" {
		t.Fatalf("agent state leaked: %#v", requests)
	}
	last := requests[3]
	if len(last) != 3 || last[0].Content != "alpha" || last[1].Content != "hello" || last[2].Content != "again" {
		t.Fatalf("failed pair committed: %#v", last)
	}
	if a.Usage() != (ensemble.Usage{Input: 14, Output: 6}) || b.Usage() != (ensemble.Usage{Input: 7, Output: 3}) {
		t.Fatal("usage was shared or invalid response counted")
	}
	if !strings.Contains(diagnostics.String(), "missing or invalid usage") {
		t.Fatal("parser could not reach root logger")
	}
	copy := a.History()
	copy[0].Content = "mutated"
	if a.History()[0].Content != "alpha" {
		t.Fatal("history accessor exposes mutable owned state")
	}
}

func TestCancellationAndDeadlineDiagnostics(t *testing.T) {
	for _, stage := range []string{"headers", "body", "after-json"} {
		for _, kind := range []string{"cancel", "deadline"} {
			t.Run(stage+"/"+kind, func(t *testing.T) {
				started := make(chan struct{})
				release := make(chan struct{})
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if stage != "headers" {
						w.Header().Set("Content-Type", "application/json")
						if stage == "after-json" {
							fmt.Fprint(w, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":2}}`)
						}
						w.(http.Flusher).Flush()
					}
					close(started)
					select {
					case <-r.Context().Done():
					case <-release:
					}
				}))
				defer server.Close()
				// Cleanup must not depend on server-side cancellation detection.
				defer close(release)
				ctx, cancel := context.WithCancel(context.Background())
				want := error(context.Canceled)
				message := "canceled"
				if kind == "deadline" {
					cancel()
					ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
					want, message = context.DeadlineExceeded, "timed out"
				}
				defer cancel()
				if kind == "cancel" {
					go func() { <-started; cancel() }()
				}
				var diagnostics bytes.Buffer
				app := ensemble.New(&diagnostics)
				a, err := app.NewAgent(ensemble.Config{APIKey: "credential-marker", Model: "test", BaseURL: server.URL + "/private-url-marker"})
				if err != nil {
					t.Fatal(err)
				}
				_, err = a.Ask(ctx, "hi")
				if !errors.Is(err, want) {
					t.Fatalf("error %v does not preserve %v", err, want)
				}
				if !strings.Contains(diagnostics.String(), message) || strings.Contains(diagnostics.String(), "malformed") {
					t.Fatalf("incorrect diagnostic: %s", diagnostics.String())
				}
				for _, secret := range []string{"credential-marker", "private-url-marker", server.URL} {
					if strings.Contains(err.Error()+diagnostics.String(), secret) {
						t.Fatal("diagnostic exposed private request material")
					}
				}
				if len(a.History()) != 0 || a.Usage() != (ensemble.Usage{}) {
					t.Fatal("failed request committed state")
				}
			})
		}
	}
}

func TestRejectedResponses(t *testing.T) {
	for _, body := range []string{`broken`, `{}`, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":-1,"output_tokens":1}}`, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1.5,"output_tokens":1}}`, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":null}}`, `{"content":[],"usage":{"input_tokens":1,"output_tokens":1}}`} {
		t.Run(body, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; fmt.Fprint(w, body) }))
			defer server.Close()
			a, err := ensemble.New(nil).NewAgent(ensemble.Config{APIKey: "test", Model: "test", BaseURL: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = a.Ask(context.Background(), "hi"); err == nil {
				t.Fatal("invalid response accepted")
			}
			if calls != 1 || len(a.History()) != 0 || a.Usage() != (ensemble.Usage{}) {
				t.Fatal("failure mutated state or retried")
			}
		})
	}
}
