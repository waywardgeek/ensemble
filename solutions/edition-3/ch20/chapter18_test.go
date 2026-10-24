package ensemble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"ensemble"
	"ensemble/internal/common"
)

func TestChapter18StatusAndModel(t *testing.T) {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		model := "claude-sonnet-4-6"
		if count == 2 {
			model = "unpriced-model"
		}
		fmt.Fprintf(w, `{"model":%q,"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":10,"output_tokens":3,"cache_read_input_tokens":5,"cache_creation_input_tokens":0}}`, model)
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", APIKey: "dummy", BaseURL: server.URL, DataDir: t.TempDir(), Tools: []string{"agent_status"}})
	if err != nil {
		t.Fatal(err)
	}
	status := func() common.EngineStatus {
		t.Helper()
		text, failed := c3Call(t, a, "agent_status", map[string]any{})
		if failed {
			t.Fatal(text)
		}
		var s common.EngineStatus
		if err := json.Unmarshal([]byte(text), &s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	if s := status(); s.Requests != 0 || s.SelectedModel != "claude-sonnet-4-6" || s.LastModel != "" {
		t.Fatalf("initial: %+v", s)
	}
	send := func() {
		t.Helper()
		wire, err := a.Render()
		if err != nil {
			t.Fatal(err)
		}
		if _, err = a.Engine().ExchangeRequest(context.Background(), wire); err != nil {
			t.Fatal(err)
		}
	}
	send()
	first := status()
	if first.Cost.Status != "known" || first.Cost.USD <= 0 || first.LastModel != "claude-sonnet-4-6" {
		t.Fatalf("first: %+v", first)
	}
	history := a.Log().Events()
	if err := a.SetModel("unpriced-model"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(history, a.Log().Events()) {
		t.Fatal("model switch rewrote history")
	}
	between := status()
	if between.SelectedModel != "unpriced-model" || between.LastModel != first.LastModel || between.Cost != first.Cost {
		t.Fatalf("switch rewrote history/accounting: %+v", between)
	}
	if err := a.SetModel("gpt-5.4"); err == nil {
		t.Fatal("cross-provider switch accepted")
	}
	if err := a.SetModel("  "); err == nil {
		t.Fatal("empty switch accepted")
	}
	if a.Config().Model != "unpriced-model" {
		t.Fatal("failed switch mutated config")
	}
	send()
	mixed := status()
	if mixed.Requests != 2 || mixed.Session.Output != 6 || mixed.LastModel != "unpriced-model" || mixed.Cost.Status != "partial" || mixed.Cost.USD != first.Cost.USD || len(mixed.Buckets) != 2 {
		t.Fatalf("mixed pricing: %+v", mixed)
	}
	if mixed.SessionHitRate == nil || *mixed.SessionHitRate != 1.0/3 {
		t.Fatalf("hit rate: %+v", mixed)
	}
}

func TestChapter18PublicWebSkill(t *testing.T) {
	t.Setenv("CH18_WEB_SECRET", "mcp-only-test-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer mcp-only-test-token" {
			t.Error("wrong independent MCP credential")
		}
		var m struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Error(err)
			return
		}
		if len(m.ID) == 0 {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var result string
		switch m.Method {
		case "initialize":
			result = `{"protocolVersion":"2025-03-26","capabilities":{"tools":{}}}`
		case "tools/list":
			result = `{"tools":[{"name":"firecrawl_search","description":"search fixture","inputSchema":{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}},{"name":"firecrawl_scrape","description":"scrape fixture","inputSchema":{"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}}]}`
		case "tools/call":
			result = `{"content":[{"type":"text","text":"loopback search result https://example.test/source"}],"isError":false}`
		default:
			t.Errorf("unexpected MCP method %s", m.Method)
			result = `{}`
		}
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":%s}`, m.ID, result)
	}))
	defer server.Close()
	skills, err := filepath.Abs("skills")
	if err != nil {
		t.Fatal(err)
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{EnableWebSearch: true, Vendor: ensemble.OpenAI, Model: "gpt-5.4", APIKey: "different-provider-key", BaseURL: "http://127.0.0.1:1", BaseDir: t.TempDir(), DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: skills, Primary: "ensemble"}, MCPServers: []ensemble.MCPServer{{Name: "firecrawl", Transport: "url", URL: server.URL, AuthEnv: "CH18_WEB_SECRET"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, failed := c3Call(t, a, "firecrawl_search", map[string]any{"query": "test"}); !failed {
		t.Fatal("unloaded remote tool was enabled")
	}
	if text, failed := c3Call(t, a, "load_skill", map[string]any{"name": "web-search"}); failed {
		t.Fatal(text)
	}
	text, failed := c3Call(t, a, "firecrawl_search", map[string]any{"query": "test"})
	if failed || !strings.Contains(text, "loopback search result") {
		t.Fatalf("public dispatch: %s", text)
	}
	for _, variable := range common.ProcessEnvironment(a) {
		if strings.HasPrefix(variable, "CH18_WEB_SECRET=") {
			t.Fatal("MCP key inherited by child process")
		}
	}
	snapshot, _ := json.Marshal(a.Skills().Snapshot())
	if strings.Contains(string(snapshot), "mcp-only-test-token") || strings.Contains(string(snapshot), "different-provider-key") {
		t.Fatal("credential in public skill snapshot")
	}
}

func TestChapter18UnavailableToolsAreUnknown(t *testing.T) {
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "local-fixture-model", Tools: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	const name = "ungranted_probe"
	absent, absentFailed := c3Call(t, a, name, map[string]any{})
	var calls atomic.Int32
	err = a.Tools().RegisterTool(common.ToolSpec{
		ToolDefinition: common.ToolDefinition{Name: name, Description: "Never granted fixture"},
		Enabled:        false,
		Handler:        func(context.Context, common.ToolCall, json.RawMessage) error { calls.Add(1); return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	denied, deniedFailed := c3Call(t, a, name, map[string]any{})
	if !absentFailed || !deniedFailed || absent != denied {
		t.Fatalf("unavailable responses differ: absent=%q (%t), denied=%q (%t)", absent, absentFailed, denied, deniedFailed)
	}
	if !strings.Contains(absent, "unknown tool") || strings.Contains(strings.ToLower(absent), "disabled") {
		t.Fatalf("availability disclosed: %q", absent)
	}
	for _, def := range a.Tools().Definitions() {
		if def.Name == name {
			t.Fatal("ungranted tool declared")
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("ungranted handler executed %d times", calls.Load())
	}
}
