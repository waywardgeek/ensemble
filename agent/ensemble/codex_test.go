package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
)

// The fake is the external provider. Real credentials, renderer, SSE parser,
// actor, tool dispatch and accounting run unchanged through their public API.
func TestCodexToolLoopAndCredentialReread(t *testing.T) {
	file := filepath.Join(t.TempDir(), "auth.json")
	writeCodex(t, file, "token-one", "account-one")
	var trace bytes.Buffer
	var calls atomic.Int32
	var key string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		token, account := "token-one", "account-one"
		if n > 1 {
			token, account = "token-two", "account-two"
		}
		if r.URL.Path != "/responses" || r.Header.Get("Authorization") != "Bearer "+token || r.Header.Get("ChatGPT-Account-ID") != account {
			t.Error("request did not use the selected route and current credential pair")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		if body["store"] != false || body["stream"] != true || !strings.HasPrefix(fmt.Sprint(body["instructions"]), "Use the echo tool.\n\n") || !strings.Contains(fmt.Sprint(body["instructions"]), "ordinary user-visible text") {
			t.Error("Codex delivery/instruction contract missing")
		}
		if body["prompt_cache_options"] != nil || strings.Contains(fmt.Sprint(body), "prompt_cache_breakpoint") {
			t.Error("explicit caching leaked to Codex")
		}
		current, _ := body["prompt_cache_key"].(string)
		if current == "" || (key != "" && current != key) {
			t.Error("cache key is missing or unstable")
		}
		key = current
		if r.Header.Get("session-id") != current {
			t.Error("cache affinity header differs from body key")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		if n == 1 {
			// Rotate both fields as Codex does between calls; the running Agent must
			// observe the replacement without restarting or writing to the file.
			writeCodex(t, file, "token-two", "account-two")
			frame(w, wireFrame{"response.reasoning_summary_text.delta", decoded(`{"output_index":0,"delta":"Check echo."}`)})
			frame(w, wireFrame{"response.output_text.delta", decoded(`{"output_index":1,"delta":"I will call echo to verify the tool result and credential rotation."}`)})
			frame(w, wireFrame{"response.completed", decoded(`{"response":{"status":"completed","model":"gpt-6-astra","output":[{"type":"reasoning","id":"r1","encrypted_content":"opaque-replay"},{"type":"message","content":[{"type":"output_text","text":"I will call echo to verify the tool result and credential rotation."}]},{"type":"function_call","call_id":"c1","name":"echo","arguments":"{}"}],"usage":{"input_tokens":5000,"output_tokens":10,"input_tokens_details":{"cached_tokens":4800}}}}`)})
		} else {
			items, _ := json.Marshal(body["input"])
			if !strings.Contains(string(items), `"output":"actual tool output"`) || !strings.Contains(string(items), "opaque-replay") {
				t.Error("actual tool result or encrypted reasoning was lost")
			}
			frame(w, wireFrame{"response.output_text.delta", decoded(`{"output_index":0,"delta":"Done."}`)})
			frame(w, wireFrame{"response.completed", decoded(`{"response":{"status":"completed","model":"gpt-6-astra","output":[{"type":"message","content":[{"type":"output_text","text":"Done."}]}],"usage":{"input_tokens":5100,"output_tokens":5,"input_tokens_details":{"cached_tokens":4900}}}}`)})
		}
	}))
	defer server.Close()
	root := ensemble.New(io.Discard)
	root.SetAPILog(&trace)
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-6-astra", BillingMode: ensemble.CredentialCodex, CodexAuthFile: file, BaseURL: server.URL, DataDir: t.TempDir(), SystemPrompt: "Use the echo tool.", Recall: ensemble.RecallSettings{Disabled: true}})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if err := a.RegisterTool(ensemble.ToolDeclaration{Name: "echo", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) { return "actual tool output", nil }); err != nil {
		t.Fatal(err)
	}
	var summary string
	answer, err := a.AskWatching(context.Background(), "Call echo", ensemble.StreamCallbacks{OnDelta: func(_ uint64, kind ensemble.DeltaKind, text string) {
		if kind == ensemble.DeltaReasoningSummary {
			summary += text
		}
	}})
	if err != nil || answer != "Done." || summary != "Check echo." || calls.Load() != 2 {
		t.Fatalf("loop: answer=%q summary=%q calls=%d err=%v", answer, summary, calls.Load(), err)
	}
	u := a.Engine().Usage()
	if u.Input != 400 || u.CacheRead != 9700 || u.Output != 15 {
		t.Fatalf("usage: %+v", u)
	}
	for _, secret := range []string{"token-one", "account-one", "token-two", "account-two"} {
		if strings.Contains(trace.String(), secret) {
			t.Fatal("credential leaked into API trace")
		}
	}
	data, err := os.ReadFile(file)
	if err != nil || !strings.Contains(string(data), "token-two") {
		t.Fatal("credential file was modified by Ensemble")
	}
}

// Deliberately supplied API keys must never turn a failed Codex request into a
// successful metered request. Even same-host redirects are refused.
func TestCodexRefusesFallbackAndRedirect(t *testing.T) {
	for _, status := range []int{401, 302, 307} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Header.Get("Authorization") != "Bearer codex-secret" {
					t.Error("wrong funding source")
				}
				w.Header().Set("Location", "/redirected")
				w.WriteHeader(status)
			}))
			defer server.Close()
			file := filepath.Join(t.TempDir(), "auth.json")
			writeCodex(t, file, "codex-secret", "codex-account")
			before, _ := os.ReadFile(file)
			a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-6-astra", BillingMode: ensemble.CredentialCodex, CodexAuthFile: file, APIKey: "metered-secret", BaseURL: server.URL, DataDir: t.TempDir(), Recall: ensemble.RecallSettings{Disabled: true}})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Shutdown()
			if _, err := a.Ask(context.Background(), "hello"); err == nil {
				t.Fatal("rejection accepted")
			}
			if calls.Load() != 1 {
				t.Fatalf("request retried or redirected: %d", calls.Load())
			}
			after, _ := os.ReadFile(file)
			if !bytes.Equal(before, after) {
				t.Fatal("external credential file changed")
			}
		})
	}
}

// Invalid profiles and an incompatible model selected between turns must fail
// locally, before a credential-bearing request can reach the external peer.
func TestCodexRejectsInvalidSelection(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(500) }))
	defer server.Close()
	for _, profile := range []string{"missing", "{broken", `{"tokens":{"access_token":"secret"}}`, `{"auth_mode":"apikey","tokens":{"access_token":"secret","account_id":"account"}}`, "valid"} {
		t.Run(profile, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "auth.json")
			if profile == "valid" {
				writeCodex(t, file, "secret", "account")
			} else if profile != "missing" {
				if err := os.WriteFile(file, []byte(profile), 0600); err != nil {
					t.Fatal(err)
				}
			}
			a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-6-astra", BillingMode: ensemble.CredentialCodex, CodexAuthFile: file, APIKey: "metered-secret", BaseURL: server.URL, DataDir: t.TempDir(), Recall: ensemble.RecallSettings{Disabled: true}})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Shutdown()
			if profile == "valid" {
				if err := a.SetModel("gpt-4o-mini"); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := a.Ask(context.Background(), "hello"); err == nil {
				t.Fatal("invalid selection accepted")
			}
		})
	}
	if calls.Load() != 0 {
		t.Fatalf("invalid selection sent %d requests", calls.Load())
	}
}

func writeCodex(t *testing.T, file, token, account string) {
	t.Helper()
	data, err := json.Marshal(map[string]any{"auth_mode": "chatgpt", "tokens": map[string]string{"access_token": token, "account_id": account}})
	if err != nil {
		t.Error(err)
		return
	}
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Error(err)
	}
}

// Both shipped clients use ConfigFromEnv. The opt-in chooses the file without
// reading it, so offline rendering still works with a missing local sign-in.
func TestCodexConfiguration(t *testing.T) {
	t.Setenv("LLM_VENDOR", "openai")
	t.Setenv("LLM_MODEL", "gpt-6-astra")
	t.Setenv("LLM_BASE_URL", "")
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("OPENAI_BILLING_MODE", "codex")
	t.Setenv("LLM_WOL_MAC", "02:00:00:00:00:01")
	t.Setenv("EN_CODEX_AUTH_FILE", "")
	t.Setenv("CODEX_HOME", t.TempDir())
	cfg, err := ensemble.ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CodexAuthFile != filepath.Join(os.Getenv("CODEX_HOME"), "auth.json") {
		t.Fatal("Codex home ignored")
	}
	if cfg.WakeOnLANMAC != "02:00:00:00:00:01" {
		t.Fatal("Wake-on-LAN MAC ignored")
	}
	t.Setenv("EN_CODEX_AUTH_FILE", filepath.Join(t.TempDir(), "selected.json"))
	cfg, err = ensemble.ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CodexAuthFile != os.Getenv("EN_CODEX_AUTH_FILE") {
		t.Fatal("explicit file ignored")
	}
	cfg.DataDir = t.TempDir()
	a, err := ensemble.New(io.Discard).NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if a.Config().BaseURL != "https://chatgpt.com/backend-api/codex" {
		t.Fatal("wrong default endpoint")
	}
	if _, err := a.Engine().Render(&ensemble.Context{}); err != nil {
		t.Fatalf("offline render needs credentials: %v", err)
	}
	t.Setenv("OPENAI_BILLING_MODE", "api")
	cfg, err = ensemble.ConfigFromEnv()
	if err != nil || cfg.CodexAuthFile != "" {
		t.Fatal("Codex path selected without opt-in")
	}
}

// Recall and compression share transport credentials but have different input
// builders. Exercise both real builders so a working chat cannot hide an absent
// required instructions field in the history-free judge request.
func TestCodexAuxiliaryRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		if body["instructions"] == nil || body["instructions"] == "" || body["prompt_cache_key"] != r.Header.Get("session-id") || r.Header.Get("Authorization") != "Bearer auxiliary-token" {
			t.Error("auxiliary request lost its Codex contract")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		output := `[{"type":"message","content":[{"type":"output_text","text":"selected"}]}]`
		if body["tools"] != nil {
			output = `[{"type":"function_call","call_id":"compress","name":"submit","arguments":"{\"memory\":\"I retained the reason.\"}"}]`
		}
		frame(w, wireFrame{"response.completed", decoded(`{"response":{"status":"completed","model":"gpt-6-astra","output":` + output + `,"usage":{"input_tokens":50,"output_tokens":10}}}`)})
	}))
	defer server.Close()
	file := filepath.Join(t.TempDir(), "auth.json")
	writeCodex(t, file, "auxiliary-token", "auxiliary-account")
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-6-astra", BillingMode: ensemble.CredentialCodex, CodexAuthFile: file, BaseURL: server.URL, DataDir: t.TempDir(), Recall: ensemble.RecallSettings{Disabled: true}})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if value, err := a.Engine().Judge(context.Background(), "Select relevance"); err != nil || value != "selected" {
		t.Fatalf("judge: %q %v", value, err)
	}
	if value, err := a.Engine().Compress(context.Background(), "Remember the reason", 2); err != nil || value != "I retained the reason." {
		t.Fatalf("compress: %q %v", value, err)
	}
}

// The launcher configures summaries through the environment. Verify that the
// real Codex request carries that choice, independently of provider verbosity.
func TestCodexDetailedSummaryFromEnvironment(t *testing.T) {
	file := filepath.Join(t.TempDir(), "auth.json")
	writeCodex(t, file, "fixture-token", "fixture-account")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			http.Error(w, "bad JSON", http.StatusBadRequest)
			return
		}
		reasoning, _ := body["reasoning"].(map[string]any)
		if reasoning["summary"] != "detailed" || reasoning["effort"] != "high" {
			t.Errorf("reasoning options did not reach provider: %v", reasoning)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		frame(w, wireFrame{"response.completed", decoded(`{"response":{"status":"completed","model":"gpt-6-astra","output":[{"type":"message","content":[{"type":"output_text","text":"Done."}]}],"usage":{"input_tokens":1,"output_tokens":1}}}`)})
	}))
	defer server.Close()
	for key, value := range map[string]string{
		"LLM_VENDOR": "openai", "LLM_MODEL": "gpt-6-astra", "LLM_BASE_URL": server.URL,
		"OPENAI_BILLING_MODE": "codex", "EN_CODEX_AUTH_FILE": file,
		"OPENAI_REASONING_EFFORT": "high", "OPENAI_REASONING_SUMMARY": "detailed",
	} {
		t.Setenv(key, value)
	}
	cfg, err := ensemble.ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	cfg.DataDir, cfg.Workspace = t.TempDir(), t.TempDir()
	cfg.SystemPrompt = "Be concise."
	cfg.Recall.Disabled = true
	a, err := ensemble.New(io.Discard).NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := a.Ask(ctx, "Check request settings."); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("provider calls: %d", calls.Load())
	}
}
