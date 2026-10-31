package ensemble_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
)

const lifetimeResponse = `{"status":"completed","model":"gpt-6-astra","output":[{"type":"message","content":[{"type":"output_text","text":"Still here."}]}],"usage":{"input_tokens":1,"output_tokens":1}}`

func lifetimeAgent(t *testing.T, server *httptest.Server, buffered bool) ensemble.Agent {
	t.Helper()
	return actorAgent(t, server, func(cfg *ensemble.Config) {
		cfg.Vendor, cfg.Model, cfg.APIKey, cfg.BillingMode = ensemble.OpenAI, "gpt-6-astra", "fixture", "api"
		cfg.DisableStreaming = buffered
		cfg.Recall.Disabled = true
	})
}

func completeLifetimeResponse(w http.ResponseWriter, buffered bool) {
	if buffered {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, lifetimeResponse)
	} else {
		frame(w, wireFrame{"response.completed", decoded(`{"response":` + lifetimeResponse + `}`)})
	}
}

// This deliberately crosses the old real 60-second deadline: replacing that
// deadline with another test-only client would not catch the production bug.
// Both subtests overlap, so the complete regression takes about one minute.
func TestModelResponseBeyondOneMinute(t *testing.T) {
	if testing.Short() {
		t.Skip("real response-lifetime regression waits 61 seconds")
	}
	for _, buffered := range []bool{false, true} {
		name := "streaming-body"
		if buffered {
			name = "buffered-before-headers"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(io.Discard, r.Body)
				r.Body.Close()
				if !buffered {
					w.Header().Set("Content-Type", "text/event-stream")
					io.WriteString(w, ": waiting for model\n\n")
					w.(http.Flusher).Flush()
				}
				timer := time.NewTimer(61 * time.Second)
				defer timer.Stop()
				select {
				case <-timer.C:
					completeLifetimeResponse(w, buffered)
				case <-r.Context().Done():
				}
			}))
			defer server.Close()
			a := lifetimeAgent(t, server, buffered)
			ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
			defer cancel()
			text, err := a.Ask(ctx, "Take your time.")
			if err != nil || text != "Still here." {
				t.Fatalf("long response: text=%q err=%v", text, err)
			}
		})
	}
}

// No implicit deadline must not mean no cancellation. Exercise real actor
// interruption and shutdown, and observe cancellation at the external server.
func TestBlockedModelResponseCancellation(t *testing.T) {
	for _, mode := range []string{"interrupt", "shutdown"} {
		for _, buffered := range []bool{false, true} {
			name := mode + "/streaming-body"
			if buffered {
				name = mode + "/before-headers"
			}
			t.Run(name, func(t *testing.T) {
				entered, canceled := make(chan struct{}), make(chan struct{})
				var calls atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					io.Copy(io.Discard, r.Body)
					r.Body.Close()
					if calls.Add(1) > 1 {
						if !buffered {
							w.Header().Set("Content-Type", "text/event-stream")
						}
						completeLifetimeResponse(w, buffered)
						return
					}
					if !buffered {
						w.Header().Set("Content-Type", "text/event-stream")
						io.WriteString(w, ": waiting for model\n\n")
						w.(http.Flusher).Flush()
					}
					close(entered)
					<-r.Context().Done()
					close(canceled)
				}))
				defer server.Close()
				a := lifetimeAgent(t, server, buffered)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				result := a.Submit(ctx, "Wait for interrupt.")
				select {
				case <-entered:
				case <-ctx.Done():
					t.Fatal("request did not reach provider")
				}
				if mode == "interrupt" {
					a.Post(ensemble.Interrupt{})
				} else if err := a.Shutdown(); err != nil {
					t.Fatal(err)
				}
				select {
				case r := <-result:
					if r.Err == nil {
						t.Fatal("canceled response incorrectly succeeded")
					}
					if mode == "interrupt" && !errors.Is(r.Err, ensemble.ErrInterrupted) {
						t.Fatalf("interrupt error: %v", r.Err)
					}
				case <-ctx.Done():
					t.Fatal("canceled turn did not return")
				}
				select {
				case <-canceled:
				case <-ctx.Done():
					t.Fatal("provider request was not canceled")
				}
				if mode == "interrupt" {
					text, err := a.Ask(ctx, "Continue after interrupt.")
					if err != nil || text != "Still here." {
						t.Fatalf("reused agent: text=%q err=%v", text, err)
					}
				}
			})
		}
	}
}
