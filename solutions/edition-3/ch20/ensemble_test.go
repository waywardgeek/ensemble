package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
	"ensemble/internal/agent"
	"ensemble/internal/engine"
)

const goodReply = `{"content":[{"type":"text","text":"first "},{"type":"ignored","text":"not text"},{"type":"text","text":"second"}],"usage":{"input_tokens":3,"output_tokens":2}}`

func newAgent(t *testing.T, app *ensemble.Ensemble, url, key, model string) ensemble.Agent {
	t.Helper()
	a, err := app.NewAgent(ensemble.Config{BaseURL: url, APIKey: key, Model: model})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// TestConversationAndOwners exercises real wire requests while distinguishing parents, configuration, and retained state.
func TestConversationAndOwners(t *testing.T) {
	type request struct {
		Model     string
		MaxTokens int `json:"max_tokens"`
		System    string
		Messages  []ensemble.Message
		Stream    bool
	}
	received := make(chan request, 4)
	keys := make(chan string, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/messages" || r.Header.Get("Content-Type") != "application/json" || r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("wrong wire method/path/headers")
		}
		var body request
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("request JSON: %v", err)
		}
		received <- body
		keys <- r.Header.Get("x-api-key")
		fmt.Fprint(w, goodReply)
	}))
	defer server.Close()
	app := ensemble.New(nil)
	otherApp := ensemble.New(nil)
	a := newAgent(t, app, server.URL+"/", "key-a", "model-a")
	b := newAgent(t, app, server.URL, "key-b", "model-b")
	c := newAgent(t, otherApp, server.URL, "key-c", "model-c")
	if a.Parent() != app || b.Parent() != app || c.Parent() != otherApp || a.Engine().Parent() != a || b.Engine().Parent() != b || a.Engine() == b.Engine() {
		t.Fatal("ownership graph does not identify the actual parents")
	}
	// Mutating returned configuration must not change future requests.
	config := a.Config()
	config.APIKey, config.Model = "changed", "changed"
	var history []ensemble.Message
	for _, question := range []string{"Keep this fact: Amber Tern.", "What was the fact?"} {
		reply, err := a.Ask(context.Background(), question)
		if err != nil || reply != "first second" {
			t.Fatalf("reply=%q err=%v", reply, err)
		}
		body := <-received
		want := append(history, ensemble.Message{Role: "user", Content: question})
		if !reflect.DeepEqual(body.Messages, want) || body.Model != "model-a" || body.MaxTokens <= 0 || body.System == "" || body.Stream || <-keys != "key-a" {
			t.Fatalf("request did not preserve config and complete history: %+v", body)
		}
		history = append(want, ensemble.Message{Role: "assistant", Content: reply})
	}
	if !reflect.DeepEqual(a.History(), history) || a.Usage() != (ensemble.Usage{Input: 6, Output: 4}) || a.Engine().Usage() != a.Usage() {
		t.Fatalf("history=%v usage=%+v", a.History(), a.Usage())
	}
	copyHistory := a.History()
	copyHistory[0].Content = "corruption"
	if !reflect.DeepEqual(a.History(), history) {
		t.Fatal("History exposed the retained slice")
	}
	for i, isolated := range []ensemble.Agent{b, c} {
		if len(isolated.History()) != 0 || isolated.Usage() != (ensemble.Usage{}) {
			t.Fatal("new Agent inherited another conversation")
		}
		if _, err := isolated.Ask(context.Background(), "independent"); err != nil {
			t.Fatal(err)
		}
		body := <-received
		if len(body.Messages) != 1 || body.Model != []string{"model-b", "model-c"}[i] || <-keys != []string{"key-b", "key-c"}[i] {
			t.Fatal("Agent configuration or conversation leaked across owners")
		}
	}
	if a.Usage() != (ensemble.Usage{Input: 6, Output: 4}) {
		t.Fatal("another Agent changed the first Agent's usage")
	}
}

// TestFailureDoesNotCommitHistory keeps the completed-answer view distinct from
// Chapter 2 causal dialogue, which retains failed attempts and their paid usage.
func TestFailureDoesNotCommitHistory(t *testing.T) {
	var calls atomic.Int32
	requests := make(chan []ensemble.Message, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Messages []ensemble.Message }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("request JSON: %v", err)
		}
		requests <- body.Messages
		if calls.Add(1) == 2 {
			// This response was paid for, but is not a completed text answer.
			fmt.Fprint(w, `{"content":[],"usage":{"input_tokens":7,"output_tokens":1}}`)
			return
		}
		fmt.Fprint(w, goodReply)
	}))
	defer server.Close()
	a := newAgent(t, ensemble.New(nil), server.URL, "fake-key", "fake-model")
	if _, err := a.Ask(context.Background(), "first"); err != nil {
		t.Fatal(err)
	}
	<-requests
	before := a.History()
	if _, err := a.Ask(context.Background(), "failed"); err == nil {
		t.Fatal("empty answer succeeded")
	}
	<-requests
	if !reflect.DeepEqual(before, a.History()) || a.Usage() != (ensemble.Usage{Input: 10, Output: 3}) {
		t.Fatal("failed turn changed history or lost measured usage")
	}
	if _, err := a.Ask(context.Background(), "retry manually"); err != nil {
		t.Fatal(err)
	}
	got := <-requests
	if len(got) != len(before)+1 || !reflect.DeepEqual(got[:len(before)], before) ||
		got[len(before)].Role != "user" || !strings.Contains(got[len(before)].Content, "failed") ||
		!strings.HasSuffix(got[len(before)].Content, "retry manually") {
		t.Fatalf("causal failed input or retry missing from next request: %v", got)
	}
	if calls.Load() != 3 || a.Usage() != (ensemble.Usage{Input: 13, Output: 5}) {
		t.Fatal("unexpected request count or usage")
	}
}

// TestResponseErrorsAndUsage separates measured zero, absent accounting, and malformed or unusable responses.
func TestResponseErrorsAndUsage(t *testing.T) {
	for _, tc := range []struct {
		name, body, wantError string
		status                int
		usage                 ensemble.Usage
	}{
		{"status", "remote-secret fake-key", "HTTP status 429", 429, ensemble.Usage{}},
		{"malformed", "remote-secret fake-key", "decoding response failed", 200, ensemble.Usage{Input: -1, Output: -1, CacheRead: -1, CacheWrite: -1}},
		{"missing usage", `{"content":[{"type":"text","text":"answer"}]}`, "missing usage", 200, ensemble.Usage{Input: -1, Output: -1, CacheRead: -1, CacheWrite: -1}},
		{"null usage", `{"usage":null}`, "missing usage", 200, ensemble.Usage{Input: -1, Output: -1, CacheRead: -1, CacheWrite: -1}},
		{"partial usage", `{"usage":{"input_tokens":1}}`, "missing usage", 200, ensemble.Usage{Input: 1, Output: -1}},
		{"negative usage", `{"usage":{"input_tokens":-1,"output_tokens":2}}`, "negative usage", 200, ensemble.Usage{Input: -1, Output: 2}},
		{"empty text", `{"content":[{"type":"text","text":"  "}],"usage":{"input_tokens":5,"output_tokens":2}}`, "no usable text", 200, ensemble.Usage{Input: 5, Output: 2}},
		{"measured zero", `{"content":[{"type":"text","text":"zero"}],"usage":{"input_tokens":0,"output_tokens":0,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`, "", 200, ensemble.Usage{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			var diagnostics, otherDiagnostics bytes.Buffer
			app := ensemble.New(log.New(&diagnostics, "owner: ", 0))
			other := ensemble.New(log.New(&otherDiagnostics, "other: ", 0))
			_ = newAgent(t, other, server.URL, "other-key", "other-model")
			a := newAgent(t, app, server.URL, "fake-key", "fake-model")
			_, err := a.Ask(context.Background(), "question")
			if tc.wantError == "" {
				if err != nil || len(a.History()) != 2 || diagnostics.Len() != 0 {
					t.Fatalf("zero accounting rejected: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) || len(a.History()) != 0 || !strings.Contains(diagnostics.String(), "owner: "+err.Error()) {
					t.Fatalf("err=%v history=%v log=%q", err, a.History(), diagnostics.String())
				}
				for _, secret := range []string{"remote-secret", "fake-key", server.URL} {
					if strings.Contains(err.Error()+diagnostics.String(), secret) {
						t.Fatal("diagnostic disclosed remote or credential data")
					}
				}
			}
			if calls.Load() != 1 || a.Usage() != tc.usage || otherDiagnostics.Len() != 0 {
				t.Fatalf("calls=%d usage=%+v otherLog=%q", calls.Load(), a.Usage(), otherDiagnostics.String())
			}
		})
	}
}

// TestValidationBeforeHTTP rejects missing owners and configuration before spending an HTTP request.
func TestValidationBeforeHTTP(t *testing.T) {
	var missingRoot *ensemble.Ensemble
	if _, err := missingRoot.NewAgent(ensemble.Config{APIKey: "fake-key", Model: "fake-model"}); err == nil {
		t.Fatal("nil root accepted")
	}
	if _, err := agent.New(nil, ensemble.Config{APIKey: "fake-key", Model: "fake-model"}); err == nil || err.Error() != "Agent requires an Ensemble parent" {
		t.Fatal("nil Ensemble parent accepted")
	}
	if _, err := engine.New(nil); err == nil {
		t.Fatal("nil Agent parent accepted")
	}
	app := ensemble.New(nil)
	for _, config := range []ensemble.Config{
		{APIKey: "fake-key"},
		{APIKey: "fake-key", Model: "fake-model", BaseURL: "https://user:secret@example.invalid"},
		{APIKey: "fake-key", Model: "fake-model", BaseURL: "not-a-url"},
	} {
		if _, err := app.NewAgent(config); err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatalf("invalid config error=%v", err)
		}
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); fmt.Fprint(w, goodReply) }))
	defer server.Close()
	offline := newAgent(t, app, server.URL, "", "fake-model")
	if _, err := offline.Render(); err != nil {
		t.Fatalf("credential-free offline render: %v", err)
	}
	if _, err := offline.Ask(context.Background(), "no paid request"); err == nil || calls.Load() != 0 || offline.Usage() != (ensemble.Usage{}) {
		t.Fatal("missing key reached live HTTP or incurred usage")
	}
	a := newAgent(t, app, server.URL, "fake-key", "fake-model")
	if _, err := a.Ask(context.Background(), " \n\t"); err == nil || calls.Load() != 0 || len(a.History()) != 0 || a.Usage() != (ensemble.Usage{}) {
		t.Fatal("blank question reached HTTP or changed state")
	}
	defaultAgent := newAgent(t, app, "", "fake-key", "fake-model")
	if defaultAgent.Config().BaseURL != "https://api.anthropic.com" {
		t.Fatal("incorrect default URL")
	}
}

// TestCancellationWhileWaiting checks caller cancellation during both header and body waits, without a model deadline.
func TestCancellationWhileWaiting(t *testing.T) {
	for _, bodyPhase := range []bool{false, true} {
		t.Run(fmt.Sprintf("body=%t", bodyPhase), func(t *testing.T) {
			arrived := make(chan struct{})
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(io.Discard, r.Body) // Allow the HTTP server to observe disconnect.
				if bodyPhase {
					w.WriteHeader(200)
					w.(http.Flusher).Flush()
				}
				close(arrived)
				select {
				case <-r.Context().Done():
				case <-release:
				}
			}))
			defer func() { close(release); server.Close() }()
			app := ensemble.New(nil)
			defer app.Shutdown()
			a := newAgent(t, app, server.URL, "fake-key", "fake-model")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := a.Ask(ctx, "cancel me"); done <- err }()
			select {
			case <-arrived:
			case <-time.After(5 * time.Second):
				t.Fatal("request never arrived")
			}
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("caller error=%v, want context.Canceled", err)
				}
				if completed := a.History(); len(completed) != 0 {
					t.Fatalf("completed-text history after cancellation=%+v", completed)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("caller cancellation did not release HTTP wait")
			}
			// Ask interruption can precede Engine completion. Join the actual
			// worker, then use its stale canonical receipt (not the completed-
			// text History view or the actor's own interruption error) to choose
			// the one truthful accounting outcome for this request.
			stopped := make(chan error, 1)
			go func() { stopped <- app.Shutdown() }()
			select {
			case err := <-stopped:
				if err != nil {
					t.Fatalf("shutdown join: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("canceled worker did not settle")
			}
			var receipts []string
			for _, event := range a.Log().Events() {
				if event.Type == ensemble.ErrorOccurred && event.Stale && event.Error != nil {
					receipts = append(receipts, event.Error.Message)
				}
			}
			if len(receipts) != 1 {
				t.Fatalf("settled worker receipts=%q usage=%+v canonical_events=%+v", receipts, a.Usage(), a.Log().Events())
			}
			wantUsage := ensemble.Usage{}
			switch receipts[0] {
			case context.Canceled.Error(): // Do did not return response headers.
			case "reading response failed", "decoding response failed":
				if !bodyPhase {
					t.Fatalf("unexpected HTTP-200 parse receipt before headers: %q", receipts)
				}
				wantUsage = ensemble.Usage{Input: -1, Output: -1, CacheRead: -1, CacheWrite: -1}
			default:
				t.Fatalf("unexpected settled worker receipt=%q usage=%+v", receipts, a.Usage())
			}
			completed, usage := a.History(), a.Usage()
			if len(completed) != 0 || usage != wantUsage {
				t.Fatalf("settled cancellation: completed_text_history=%+v usage=%+v want_usage=%+v worker_receipts=%q", completed, usage, wantUsage, receipts)
			}
		})
	}
}

// TestRedirectIsNotFollowed ensures an unexpected redirect cannot forward the credential-bearing request.
func TestRedirectIsNotFollowed(t *testing.T) {
	var forwarded atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded.Add(1); fmt.Fprint(w, goodReply) }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	a := newAgent(t, ensemble.New(nil), server.URL, "fake-key", "fake-model")
	if _, err := a.Ask(context.Background(), "question"); err == nil || forwarded.Load() != 0 {
		t.Fatal("redirect forwarded a credential-bearing request")
	}
}

// TestTransportErrors checks sanitized errors for an incomplete body and a disconnected local peer.
func TestTransportErrors(t *testing.T) {
	for _, unreachable := range []bool{false, true} {
		t.Run(fmt.Sprintf("unreachable=%t", unreachable), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Closing with fewer bytes than promised must not become a reply.
				w.Header().Set("Content-Length", "1000")
				fmt.Fprint(w, "remote-secret")
			}))
			defer server.Close()
			if unreachable {
				server.Close()
			}
			var diagnostics bytes.Buffer
			a := newAgent(t, ensemble.New(log.New(&diagnostics, "", 0)), server.URL, "fake-key", "fake-model")
			answer, err := a.Ask(context.Background(), "question")
			want := "reading response failed"
			wantUsage := ensemble.Usage{Input: -1, Output: -1, CacheRead: -1, CacheWrite: -1}
			if unreachable {
				want = "HTTP request failed"
				wantUsage = ensemble.Usage{}
			}
			if answer != "" || err == nil || err.Error() != want || diagnostics.String() != want+"\n" || len(a.History()) != 0 || a.Usage() != wantUsage {
				t.Fatalf("err=%v diagnostics=%q history=%v usage=%+v", err, diagnostics.String(), a.History(), a.Usage())
			}
		})
	}
}
