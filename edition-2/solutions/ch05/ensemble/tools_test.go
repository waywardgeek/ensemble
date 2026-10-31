package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

type toolCall struct {
	name string
	args map[string]any
}

// This provider fake supplies real wire calls, then echoes the received result
// content. Files, shell, Agent, HTTP, parsing, history and rendering remain real.
func toolSession(t *testing.T, calls ...toolCall) (ensemble.Agent, []common.Part) {
	t.Helper()
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			blocks := []map[string]any{{"type": "text", "text": "working"}}
			for i, call := range calls {
				blocks = append(blocks, map[string]any{"type": "tool_use", "id": string(rune('a' + i)), "name": call.name, "input": call.args})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"content": blocks, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
			return
		}
		var body struct {
			// Messages retains provider message order, including tool-result boundaries.
			Messages []struct {
				// Content preserves the provider's ordered blocks or message text.
				Content []json.RawMessage
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		last := body.Messages[len(body.Messages)-1].Content
		if len(last) != len(calls) {
			t.Errorf("wire results = %d, want %d", len(last), len(calls))
		}
		for i, raw := range last {
			var result struct {
				// Type selects the wire block variant before its other fields are interpreted.
				Type string
				// ID is the provider-issued call identity, retained for its matching result.
				ID string `json:"tool_use_id"`
			}
			_ = json.Unmarshal(raw, &result)
			if result.Type != "tool_result" || result.ID != string(rune('a'+i)) {
				t.Errorf("result correlation/order: %s", raw)
			}
		}
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{DataDir: t.TempDir(), Model: "fake", BaseURL: server.URL, BuiltinTools: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := a.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	answer, err := a.Ask(context.Background(), "exercise tools")
	if err != nil || answer != "done" {
		t.Fatalf("tool turn: %q %v", answer, err)
	}
	var results []common.Part
	for _, entry := range a.History().Context().Dialogue {
		for _, p := range entry.Parts {
			if p.Type == "tool_result" {
				results = append(results, p)
			}
		}
	}
	if requests != 2 || len(results) != len(calls) || a.History().Context().Turn != "idle" {
		t.Fatalf("incomplete turn: requests=%d results=%d state=%s", requests, len(results), a.History().Context().Turn)
	}
	return a, results
}
func resultText(p common.Part) string { return p.Parts[0].Text }

// Real file handlers must enforce their argument and edit contracts through model
// dispatch.
func TestFileContractsThroughLoop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "notes.txt")
	_, results := toolSession(t,
		toolCall{"write_file", map[string]any{"path": path, "content": "one\nneedle\nthree\nfour\nneedle\nsix\nseven\neight\nneedle\nten\n"}},
		toolCall{"read_file", map[string]any{"path": path, "start_line": 2, "end_line": 3, "max_bytes": 7}},
		toolCall{"search_files", map[string]any{"path": path, "pattern": "needle", "context_lines": 1}},
		toolCall{"write_file", map[string]any{"path": path, "content": "tail\n", "append": true}},
		toolCall{"read_file", map[string]any{"path": path, "start_line": 11}},
		toolCall{"edit_file", map[string]any{"path": path, "old_text": "needle", "new_text": "wrong"}},
		toolCall{"edit_file", map[string]any{"path": path, "old_text": "absent", "new_text": "wrong"}},
		toolCall{"write_file", map[string]any{"path": path}},
		toolCall{"write_file", map[string]any{"path": filepath.Join(dir, "overlap.txt"), "content": "aaa"}},
		toolCall{"edit_file", map[string]any{"path": filepath.Join(dir, "overlap.txt"), "old_text": "aa", "new_text": "wrong"}},
	)
	if got := resultText(results[1]); !strings.HasPrefix(got, "needle\n\n[truncated;") || strings.Contains(got, "three") {
		t.Fatalf("read range/cap: %q", got)
	}
	want := path + ":2:needle\n"
	if !strings.Contains(resultText(results[2]), want) {
		t.Fatal("search omitted matching line")
	}
	expected := ""
	for i, line := range []string{"one", "needle", "three", "four", "needle", "six"} {
		sep := "-"
		if line == "needle" {
			sep = ":"
		}
		expected += path + sep + string(rune('1'+i)) + sep + line + "\n"
	}
	expected += "--\n" + path + "-8-eight\n" + path + ":9:needle\n" + path + "-10-ten\n"
	if got := resultText(results[2]); got != expected {
		t.Fatalf("grep windows:\n%s\nwant:\n%s", got, expected)
	}
	if results[3].IsError || resultText(results[4]) != "tail\n" {
		t.Fatal("append did not preserve/add bytes")
	}
	for i, want := range map[int]string{5: "matched 3 times", 6: "matched 0 times", 7: "missing required argument content", 9: "matched 2 times"} {
		if !results[i].IsError || !strings.Contains(resultText(results[i]), want) {
			t.Fatalf("failure %d: %+v", i, results[i])
		}
	}
	data, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(data), "wrong") || !strings.HasSuffix(string(data), "tail\n") {
		t.Fatal("refusal changed file")
	}
}

// The shell must receive filtered credentials and report its actual exit status.
func TestCommandEnvironmentAndExitThroughLoop(t *testing.T) {
	for _, name := range []string{"LLM_API_KEY", "ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY"} {
		t.Setenv(name, "dummy-secret")
	}
	t.Setenv("ENSEMBLE_TOOL_TEST", "ordinary")
	_, results := toolSession(t, toolCall{"run_command", map[string]any{"command": `printf '%s' "$LLM_API_KEY$ANTHROPIC_API_KEY$OPENAI_API_KEY$GEMINI_API_KEY"; printf '%s' "$ENSEMBLE_TOOL_TEST"; printf 'problem' >&2; exit 7`}})
	got := resultText(results[0])
	if results[0].IsError || got != "ordinaryproblem\nexit_code: 7\n" {
		t.Fatalf("command result/credential isolation: %q error=%t", got, results[0].IsError)
	}
}

// Only enabled per-Agent tools may appear in each vendor's declarations.
func TestRegistryDeclarationsAndIsolation(t *testing.T) {
	for _, vendor := range []ensemble.Vendor{ensemble.Anthropic, ensemble.OpenAI, ensemble.Gemini} {
		app := ensemble.New(io.Discard)
		for _, enabled := range []bool{true, false} {
			a, err := app.NewAgent(ensemble.Config{DataDir: t.TempDir(), Vendor: vendor, Model: "fake", BuiltinTools: enabled})
			if err != nil {
				t.Fatal(err)
			}
			raw, err := a.Engine().Render(a.History().Context())
			if err != nil {
				t.Fatal(err)
			}
			var body map[string]json.RawMessage
			_ = json.Unmarshal(raw, &body)
			_, present := body["tools"]
			if present != enabled {
				t.Fatalf("%s registry enabled=%t tools present=%t", vendor, enabled, present)
			}
			if enabled && len(a.Declarations()) != 10 {
				t.Fatal("registry lacks ten tools")
			}
		}
	}
}

// A zero setting must allow the named default batch budget and then refuse the next
// batch.
func TestLoopRoundBound(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = io.WriteString(w, `{"content":[{"type":"tool_use","id":"again","name":"read_file","input":{"path":42}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{DataDir: t.TempDir(), Model: "fake", BaseURL: server.URL, BuiltinTools: true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.Ask(context.Background(), "keep calling")
	if err == nil || !strings.Contains(err.Error(), "tool round limit") || requests != ensemble.DefaultMaxToolRounds+1 {
		t.Fatalf("round bound: requests=%d error=%v", requests, err)
	}
}

// ID-less Gemini calls must remain distinct and pair with their own tool results.
func TestGeminiMissingIDsStillPairDistinctCalls(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			_, _ = io.WriteString(w, `{"candidates":[{"content":{"parts":[{"functionCall":{"name":"read_file","args":{"path":42}}},{"functionCall":{"name":"read_file","args":{}}}]}}],"usageMetadata":{"promptTokenCount":1}}`)
			return
		}
		var body struct {
			// Contents carries Gemini conversation turns using its user/model vocabulary.
			Contents []struct {
				// Parts retains ordered Gemini content blocks within a turn.
				Parts []struct {
					// Response is the object-valued Gemini tool testimony, never a bare string.
					Response *struct {
						// ID is the provider-issued call identity, retained for its matching result.
						ID string `json:"id"`
					} `json:"functionResponse"`
				}
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		var ids []string
		for _, entry := range body.Contents {
			for _, p := range entry.Parts {
				if p.Response != nil {
					ids = append(ids, p.Response.ID)
				}
			}
		}
		if len(ids) != 2 || ids[0] == "" || ids[0] == ids[1] {
			t.Errorf("missing/distinct result IDs: %v", ids)
		}
		_, _ = io.WriteString(w, `{"candidates":[{"content":{"parts":[{"text":"recovered"}]}}],"usageMetadata":{"promptTokenCount":1}}`)
	}))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{DataDir: t.TempDir(), Vendor: ensemble.Gemini, Model: "fake", BaseURL: server.URL, BuiltinTools: true})
	if err != nil {
		t.Fatal(err)
	}
	answer, err := a.Ask(context.Background(), "go")
	if err != nil || answer != "recovered" || a.History().Context().Turn != "idle" {
		t.Fatalf("missing ID loop: %q %v", answer, err)
	}
}
