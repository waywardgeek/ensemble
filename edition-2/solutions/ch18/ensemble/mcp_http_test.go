package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
)

// Only the external peer is fake: grants, jobs, client and renderer stay real.
func TestMCPHTTPSkillAndResults(t *testing.T) {
	t.Setenv("ENSEMBLE_TEST_MCP_KEY", "test-bearer")
	var calls, posts atomic.Int32
	marker := "page-token-" + t.TempDir()
	large := strings.Repeat("page content ", 7000) + marker
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts.Add(1)
		if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer test-bearer" || r.Header.Get("Accept") != "application/json, text/event-stream" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing HTTP method, auth or content negotiation")
		}
		var m struct {
			// ID preserves correlation from the real client.
			ID json.RawMessage
			// Method selects the external protocol operation.
			Method string
			// Params contains the actual tool dispatch inputs.
			Params struct {
				// Name is the advertised remote tool identity.
				Name string
				// Arguments preserves query and URL values across dispatch.
				Arguments map[string]string
			}
		}
		json.NewDecoder(r.Body).Decode(&m)
		if m.Method != "initialize" && r.Header.Get("Mcp-Session-Id") != "session-test" {
			t.Error("session header missing")
		}
		result := any(map[string]any{})
		switch m.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "session-test")
			result = map[string]string{"protocolVersion": "2024-11-05"}
		case "notifications/initialized":
			w.WriteHeader(http.StatusAccepted)
			return
		case "tools/list":
			// Several frames, CRLF, comments and multiline data; notifications must not
			// become discovery replies. The correlated result uses the same POST body.
			w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
			fmt.Fprint(w, ": heartbeat\r\nevent: message\r\ndata: {\r\ndata: \"jsonrpc\":\"2.0\",\"method\":\"notifications/progress\"}\r\n\r\n")
			fmt.Fprintf(w, "data: {\"jsonrpc\":\"2.0\",\"id\":%s,\"result\":{\"tools\":[", m.ID)
			for i, name := range []string{"firecrawl_search", "firecrawl_scrape", "account_delete"} {
				if i > 0 {
					fmt.Fprint(w, ",")
				}
				fmt.Fprintf(w, "{\"name\":%q,\"description\":\"test\",\"inputSchema\":{\"type\":\"object\"}}", name)
			}
			fmt.Fprint(w, "]}}\n\n")
			return
		case "tools/call":
			calls.Add(1)
			text, failed := "", false
			switch m.Params.Name {
			case "firecrawl_search":
				text = "search for " + m.Params.Arguments["query"]
			case "firecrawl_scrape":
				switch m.Params.Arguments["url"] {
				case "tool-error":
					text, failed = "remote scrape refused; test-bearer", true
				case "http-error":
					http.Error(w, "quota exhausted; test-bearer", http.StatusTooManyRequests)
					return
				default:
					text = large
				}
			default:
				t.Error("undeclared remote capability executed")
			}
			result = map[string]any{"content": []map[string]string{{"type": "text", "text": text}}, "isError": failed}
			// Large single-line SSE verifies no Scanner token cap truncates a page.
			if text == large {
				raw, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": m.ID, "result": result})
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprintf(w, "data: %s\n\n", raw)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": m.ID, "result": result})
	}))
	defer server.Close()
	dir := skillFiles(t, map[string]string{
		"base":       "---\nname: base\ntype: primary\ntools: account_delete\nloadable-skills: web-search\n---\nBase",
		"web-search": "---\nname: web-search\ntype: loadable\ntools: firecrawl_search firecrawl_scrape\nmcp_servers:\n  - name: web\n    transport: url\n    url: " + server.URL + "\n    auth-env: ENSEMBLE_TEST_MCP_KEY\n---\nWeb evidence",
	})
	var rounds atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if rounds.Add(1) == 1 {
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"page","name":"firecrawl_scrape","input":{"url":"page"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
			return
		}
		// Assert the actual result block, not earlier call arguments or model text.
		var request struct {
			// Messages is the real vendor conversation projection.
			Messages []struct {
				// Content carries message blocks or the bytes of a tool result.
				Content []struct {
					// Type distinguishes tool results from earlier call arguments.
					Type string
					// Content carries message blocks or the bytes of a tool result.
					Content json.RawMessage
				}
			}
		}
		if err := json.Unmarshal(body, &request); err != nil {
			t.Error(err)
		}
		found := false
		for _, m := range request.Messages {
			for _, b := range m.Content {
				if b.Type == "tool_result" && bytes.Contains(b.Content, []byte(marker)) {
					found = true
				}
			}
		}
		if !found {
			t.Error("full fetched page missing from model tool result")
		}
		io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer provider.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "fake", BaseURL: provider.URL, Workspace: t.TempDir(), DataDir: t.TempDir()}, ensemble.AgentSpec{SkillsDir: dir, PrimarySkill: "base"})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if posts.Load() != 0 || slices.Contains(toolNames(a), "firecrawl_search") {
		t.Fatal("HTTP connected before skill load")
	}
	if _, err := a.Invoke(ctx, "firecrawl_search", nil); err == nil || !strings.Contains(err.Error(), "unknown tool") {
		t.Fatal("unloaded tool admitted")
	}
	if _, err := a.Invoke(ctx, "load_skill", json.RawMessage(`{"name":"web-search"}`)); err != nil {
		t.Fatal(err)
	}
	names := toolNames(a)
	if !slices.Contains(names, "firecrawl_search") || !slices.Contains(names, "firecrawl_scrape") || slices.Contains(names, "account_delete") {
		t.Errorf("wrong skill tools: %v", names)
	}
	if _, err := a.Invoke(ctx, "account_delete", nil); err == nil || !strings.Contains(err.Error(), "unknown tool") {
		t.Error("undeclared capability admitted")
	}
	if got, err := a.Invoke(ctx, "firecrawl_search", json.RawMessage(`{"query":"primary source"}`)); err != nil || got != "search for primary source" {
		t.Fatalf("search result %q %v", got, err)
	}
	if got, err := a.Ask(ctx, "Fetch a page"); err != nil || got != "done" {
		t.Fatalf("fetch turn %q %v", got, err)
	}
	for _, tc := range []struct{ url, reason string }{{"tool-error", "remote scrape refused"}, {"http-error", "quota exhausted"}} {
		args, _ := json.Marshal(map[string]string{"url": tc.url})
		_, err := a.Invoke(ctx, "firecrawl_scrape", args)
		if err == nil || !strings.Contains(err.Error(), tc.reason) || strings.Contains(err.Error(), "test-bearer") {
			t.Fatalf("lost or unsafe diagnostic: %v", err)
		}
	}
	if calls.Load() != 4 {
		t.Fatalf("unexpected remote calls: %d", calls.Load())
	}
	done := a.MCP().Done("web")
	if _, err := a.Invoke(ctx, "unload_skill", json.RawMessage(`{"name":"web-search"}`)); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(toolNames(a), "firecrawl_search") {
		t.Error("unload retained tool declaration")
	}
	if _, err := a.Invoke(ctx, "firecrawl_search", nil); err == nil || !strings.Contains(err.Error(), "unknown tool") {
		t.Error("unloaded capability admitted")
	}
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("unload did not close client")
	}
}

// A real callback returns before the POST completes. Kill must cancel that job's
// socket; unloading separately cancels an active call and wakes the idle reader.
func TestMCPHTTPJobCancellationAndClose(t *testing.T) {
	started, canceled := make(chan struct{}, 2), make(chan struct{}, 2)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m struct {
			// ID preserves correlation from the real client.
			ID json.RawMessage
			// Method selects the external protocol operation.
			Method string
		}
		json.NewDecoder(r.Body).Decode(&m)
		if m.Method == "tools/call" {
			started <- struct{}{}
			select {
			case <-r.Context().Done():
				canceled <- struct{}{}
			case <-release:
			}
			return
		}
		if m.ID == nil {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		result := `{"protocolVersion":"2024-11-05"}`
		if m.Method == "tools/list" {
			result = `{"tools":[{"name":"wait_remote","description":"wait","inputSchema":{"type":"object"}}]}`
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":%s}`, m.ID, result)
	}))
	defer server.Close()
	defer close(release)
	t.Setenv("ENSEMBLE_EMPTY_MCP_KEY", "")
	var log bytes.Buffer
	dir := skillFiles(t, map[string]string{
		"base":   "---\nname: base\ntype: primary\nloadable-skills: remote\n---\nBase",
		"remote": "---\nname: remote\ntype: loadable\ntools: wait_remote\nmcp_servers:\n  - name: remote\n    transport: url\n    url: " + server.URL + "\n    auth-env: ENSEMBLE_EMPTY_MCP_KEY\n---\nRemote",
	})
	a, err := ensemble.New(&log).NewAgent(ensemble.Config{Model: "fake", DataDir: t.TempDir()}, ensemble.AgentSpec{SkillsDir: dir, PrimarySkill: "base"})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := a.Invoke(ctx, "load_skill", json.RawMessage(`{"name":"remote"}`)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(log.String(), "connecting without authentication") {
		t.Fatal("missing auth degradation diagnostic")
	}
	for _, stop := range []string{"kill", "unload"} {
		if out, err := a.Invoke(ctx, "wait_remote", json.RawMessage(`{"ai_callback_delay":0}`)); err != nil || !strings.Contains(out, "running") {
			t.Fatalf("callback %q %v", out, err)
		}
		select {
		case <-started:
		case <-ctx.Done():
			t.Fatal("HTTP call never started")
		}
		var handle int
		for _, event := range a.History().EventsAfter(0) {
			if event.Type == "tool_called" && event.Tool.Name == "wait_remote" {
				handle = event.Tool.Job.Handle
			}
		}
		job, err := a.Jobs().Find(handle)
		if err != nil {
			t.Fatal(err)
		}
		if job.Context().Err() != nil {
			t.Fatal("callback expiry canceled HTTP job")
		}
		if stop == "kill" {
			if err := job.Kill("test"); err != nil {
				t.Fatal(err)
			}
		} else {
			if _, err := a.Invoke(ctx, "unload_skill", json.RawMessage(`{"name":"remote"}`)); err != nil {
				t.Fatal(err)
			}
		}
		select {
		case <-canceled:
		case <-ctx.Done():
			t.Fatalf("%s did not cancel blocked POST", stop)
		}
	}
}

// The grader supplies its own primary; verify the shipped primary separately.
func TestShippedWebSkill(t *testing.T) {
	path, err := filepath.Abs("../skills")
	if err != nil {
		t.Fatal(err)
	}
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "fake", BuiltinTools: true, DataDir: t.TempDir()}, ensemble.AgentSpec{SkillsDir: path, PrimarySkill: "ensemble"})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if !slices.Contains(toolNames(a), "wait_for_job") {
		t.Fatal("primary cannot collect slow web jobs")
	}
	if !slices.Contains(a.Skills().LoadableSkills(), "web-search") {
		t.Fatal("primary does not offer web-search")
	}
}
