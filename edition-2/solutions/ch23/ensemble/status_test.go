package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
)

// Actual provider responses establish the counts. Switching models must neither
// erase them nor conceal unknown cost after work on an unpriced model.
func TestAgentStatusModelAccounting(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			// Model verifies actual routing, not a settings echo.
			Model string `json:"model"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if calls.Add(1) == 1 {
			if body.Model != "claude-sonnet-5" {
				t.Error("first request has wrong model")
			}
			io.WriteString(w, `{"content":[{"type":"text","text":"one"}],"usage":{"input_tokens":100,"output_tokens":10,"cache_read_input_tokens":20,"cache_creation_input_tokens":30}}`)
		} else {
			if body.Model != "claude-opus-5" {
				t.Error("model switch never reached provider")
			}
			io.WriteString(w, `{"content":[{"type":"text","text":"two"}],"usage":{"input_tokens":200,"output_tokens":20,"cache_read_input_tokens":40,"cache_creation_input_tokens":60}}`)
		}
	}))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-5", ResolvedModel: "prior-resolved-id", BaseURL: server.URL, DataDir: t.TempDir(), BuiltinTools: true, DisableStreaming: true})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	status := func(want ...string) {
		t.Helper()
		out, err := a.Invoke(ctx, "agent_status", nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, value := range want {
			if !strings.Contains(out, value) {
				t.Errorf("status missing %q: %s", value, out)
			}
		}
	}
	status("model: claude-sonnet-5", "last response: 0 in, 0 out, 0 cache read, 0 cache write", "cache hit rate: 0.00%", "cost: $0.000000")
	if _, err := a.Ask(ctx, "first"); err != nil {
		t.Fatal(err)
	}
	status("last response: 100 in, 10 out, 20 cache read, 30 cache write", "session total: 100 in, 10 out, 20 cache read, 30 cache write", "cost: $0.000379")
	before := a.Engine().Meter()
	if err := a.SetModel("claude-opus-5"); err != nil {
		t.Fatal(err)
	}
	if a.Engine().Meter() != before {
		t.Fatal("selection changed historical counts or cost")
	}
	if a.Engine().Target().Model != "claude-opus-5" || a.Config().Vendor != ensemble.Anthropic {
		t.Fatal("selection retained stale identity or changed provider")
	}
	status("model: claude-opus-5", "cost: $0.000379")
	if _, err := a.Ask(ctx, "second"); err != nil {
		t.Fatal(err)
	}
	status("last response: 200 in, 20 out, 40 cache read, 60 cache write", "session total: 300 in, 30 out, 60 cache read, 90 cache write", "cache hit rate: 13.33%", "cost: unknown")
	if err := a.SetModel("unpriced-model"); err != nil {
		t.Fatal(err)
	}
	// Once unpriced work occurs, another selection cannot make that total known.
	status("model: unpriced-model", "cost: unknown")
	fresh, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "unpriced-model", DataDir: t.TempDir(), BuiltinTools: true})
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Shutdown()
	out, err := fresh.Invoke(ctx, "agent_status", nil)
	if err != nil || !strings.Contains(out, "cost: unknown") || strings.Contains(out, "cost: $") {
		t.Fatalf("unknown reported as measured dollars: %q %v", out, err)
	}
}

// The tool sees usage already charged for the model response that called it;
// its result returns through normal provider rendering, not an ad hoc status API.
func TestAgentStatusToolLoop(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if calls.Add(1) == 1 {
			if !strings.Contains(string(body), `"name":"agent_status"`) {
				t.Error("status absent from declarations")
			}
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"status","name":"agent_status","input":{}}],"usage":{"input_tokens":70,"output_tokens":3,"cache_read_input_tokens":20,"cache_creation_input_tokens":10}}`)
			return
		}
		var request struct {
			// Messages is the real provider conversation, including typed tool results.
			Messages []struct {
				// Content retains the result discriminator and body separately.
				Content []struct {
					// Type identifies a tool result rather than assistant narration.
					Type string
					// Content carries the tool's report in the vendor envelope.
					Content json.RawMessage
				}
			}
		}
		json.Unmarshal(body, &request)
		found := false
		for _, m := range request.Messages {
			for _, b := range m.Content {
				if b.Type == "tool_result" {
					var blocks []struct {
						// Text is the tool output inside Anthropic's content blocks.
						Text string
					}
					if err := json.Unmarshal(b.Content, &blocks); err != nil || len(blocks) != 1 {
						t.Error("tool result does not contain one text block")
						continue
					}
					s := blocks[0].Text
					found = strings.HasPrefix(s, "model: claude-sonnet-5\n") && strings.Contains(s, "last response: 70 in, 3 out, 20 cache read, 10 cache write") && strings.Contains(s, "cache hit rate: 20.00%") && strings.Contains(s, "cost: $0.000199")
				}
			}
		}
		if !found {
			t.Error("actual status tool result missing or wrong")
		}
		io.WriteString(w, `{"content":[{"type":"text","text":"reported"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	dir := skillFiles(t, map[string]string{"base": "---\nname: base\ntype: primary\ntools: agent_status\n---\nUse status"})
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: server.URL, DataDir: t.TempDir(), BuiltinTools: true, DisableStreaming: true}, ensemble.AgentSpec{SkillsDir: dir, PrimarySkill: "base"})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if out, err := a.Ask(context.Background(), "report"); err != nil || out != "reported" {
		t.Fatalf("status loop %q %v", out, err)
	}
}

// A gated request establishes the busy state before selection is attempted.
// Rejection leaves the current request's model and provider route unchanged.
func TestModelSelectionBusyAndInvalid(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: server.URL, DataDir: t.TempDir(), DisableStreaming: true})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if err := a.SetModel("  "); err == nil {
		t.Fatal("empty model accepted")
	}
	result := a.Submit(context.Background(), "wait")
	<-entered
	err = a.SetModel("claude-opus-5")
	unchanged := a.Config().Model == "claude-sonnet-5"
	close(release)
	if result := <-result; result.Err != nil {
		t.Fatal(result.Err)
	}
	if err == nil || !unchanged {
		t.Fatal("busy selection altered the running request")
	}
	if err := a.SetModel("claude-opus-5"); err != nil {
		t.Fatal("idle positive control:", err)
	}
}
