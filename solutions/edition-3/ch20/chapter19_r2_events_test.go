package ensemble_test

import (
	"context"
	"encoding/json"
	"ensemble"
	"ensemble/internal/common"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type ch19RenewAuth struct {
	common.Auth
	old, new   string
	renew      bool
	failReject bool
	rejected   atomic.Bool
}

func (a *ch19RenewAuth) Resolve(context.Context, common.Config) (common.Credential, error) {
	token := a.old
	if a.rejected.Load() {
		token = a.new
	}
	mode := common.AuthAPI
	if a.renew {
		mode = common.AuthChatGPT
	}
	return common.Credential{Mode: mode, Token: token}, nil
}
func (a *ch19RenewAuth) Reject(context.Context, common.Credential) error {
	if a.failReject {
		return fmt.Errorf("fixture renewal refused %s", a.old)
	}
	a.rejected.Store(true)
	return nil
}
func ch19EventText(events []common.Event) string {
	var text strings.Builder
	for _, event := range events {
		if event.Response != nil {
			for _, part := range event.Response.Parts {
				if part.Type == "text" {
					text.WriteString(part.Text)
				}
			}
		}
	}
	return text.String()
}
func TestCh19R2PublicCredentialEvents(t *testing.T) {
	for _, renew := range []bool{false, true} {
		t.Run(fmt.Sprint("renew=", renew), func(t *testing.T) {
			t.Setenv("UNRELATED_TEST_TOKEN", "ordinary")
			auth := &ch19RenewAuth{old: "fixture-current-Bearer-491", new: "fixture-renewed-Bearer-872", renew: renew}
			content := "ordinary prefix " + auth.old + " ordinary suffix"
			want := "ordinary prefix [REDACTED] ordinary suffix"
			if renew {
				content += " " + auth.new
				want += " [REDACTED]"
			}
			const args = `{ "keep" : [1, 2], "text" : "ordinary" }`
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := requests.Add(1)
				if renew && n == 1 {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				wantToken := auth.old
				if renew {
					wantToken = auth.new
				}
				if r.Header.Get("Authorization") != "Bearer "+wantToken {
					t.Error("wrong effective token")
				}
				text, _ := json.Marshal(content)
				arguments, _ := json.Marshal(args)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"choices":[{"message":{"role":"assistant","content":%s,"tool_calls":[{"id":"call-1","type":"function","function":{"name":"local","arguments":%s}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`, text, arguments)
			}))
			defer server.Close()
			app := ensemble.New(nil)
			defer app.Shutdown()
			if err := app.SetAuth(auth); err != nil {
				t.Fatal(err)
			}
			a, err := app.NewAgent(ensemble.Config{Model: "fixture", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, AuthMode: ensemble.AuthAPI, BaseURL: server.URL, BaseDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			var callbacks []common.Event
			var deltas strings.Builder
			events, err := a.Engine().ExchangeRequest(context.Background(), []byte(`{"messages":[]}`), common.ParseCallbacks{OnEvent: func(e common.Event) { callbacks = append(callbacks, e) }, OnDelta: func(d common.PartDelta) {
				if string(d.Kind) == "text" {
					deltas.WriteString(d.Chunk)
				}
			}})
			if err != nil {
				t.Fatal(err)
			}
			for name, got := range map[string]string{"returned": ch19EventText(events), "callback": ch19EventText(callbacks), "delta": deltas.String()} {
				if got != want {
					t.Errorf("%s visible content not protected: %q", name, got)
				}
			}
			for _, event := range events {
				if event.Response != nil {
					for _, part := range event.Response.Parts {
						if part.Type == "tool_call" && string(part.Args) != `{"keep":[1,2],"text":"ordinary"}` {
							t.Errorf("ordinary arguments changed: %s", part.Args)
						}
					}
				}
			}
			raw, _ := json.Marshal(events)
			if strings.Contains(string(raw), auth.old) || renew && strings.Contains(string(raw), auth.new) {
				t.Error("effective token retained in public event")
			}
			if renew && requests.Load() != 2 {
				t.Fatal("renewal retry not exercised")
			}
		})
	}
}
func TestCh19R2StreamingCredentialEvents(t *testing.T) {
	const token = "fixture-stream-Bearer-537"
	first := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		frame := func(text string) {
			encoded, _ := json.Marshal(text)
			fmt.Fprintf(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":%s}}]}\n\n", encoded)
			w.(http.Flusher).Flush()
		}
		frame("ordinary first | ")
		select {
		case <-first:
		case <-time.After(3 * time.Second):
			t.Error("ordinary streaming delivery was buffered until completion")
			return
		}
		frame(token[:9])
		frame(token[9:])
		frame(" | tail f")
		fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n")
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	_ = app.SetAuth(ch19Auth{token: token})
	a, err := app.NewAgent(ensemble.Config{Model: "fixture", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, AuthMode: ensemble.AuthAPI, BaseURL: server.URL, BaseDir: t.TempDir(), StreamCapability: ensemble.StreamSSE})
	if err != nil {
		t.Fatal(err)
	}
	var deltas strings.Builder
	var callbacks []common.Event
	events, err := a.Engine().ExchangeRequest(context.Background(), []byte(`{"messages":[],"stream":true}`), common.ParseCallbacks{OnDelta: func(d common.PartDelta) {
		deltas.WriteString(d.Chunk)
		if strings.Contains(deltas.String(), "ordinary first | ") {
			select {
			case first <- struct{}{}:
			default:
			}
		}
	}, OnEvent: func(e common.Event) { callbacks = append(callbacks, e) }})
	if err != nil {
		t.Fatal(err)
	}
	want := "ordinary first | [REDACTED] | tail f"
	for name, got := range map[string]string{"returned": ch19EventText(events), "callback": ch19EventText(callbacks), "delta": deltas.String()} {
		if got != want {
			t.Errorf("%s stream content not protected/preserved: %q", name, got)
		}
	}
}

func TestCh19R2RenewalErrorEcho(t *testing.T) {
	auth := &ch19RenewAuth{old: "fixture-rejected-current-619", renew: true, failReject: true}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) }))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	_ = app.SetAuth(auth)
	a, err := app.NewAgent(ensemble.Config{Model: "fixture", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, AuthMode: ensemble.AuthAPI, BaseURL: server.URL, BaseDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	events, err := a.Engine().ExchangeRequest(context.Background(), []byte(`{"messages":[]}`))
	raw, _ := json.Marshal(events)
	if err == nil || strings.Contains(err.Error(), auth.old) || strings.Contains(string(raw), auth.old) {
		t.Fatalf("renewal error leaked: %s, %v", raw, err)
	}
	if !strings.Contains(err.Error(), "fixture renewal refused [REDACTED]") {
		t.Fatalf("ordinary error context lost: %v", err)
	}
}
