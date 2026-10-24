package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type lockedOutput struct {
	sync.Mutex
	bytes.Buffer
}

func (b *lockedOutput) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	return b.Buffer.Write(p)
}
func (b *lockedOutput) text() string { b.Lock(); defer b.Unlock(); return b.Buffer.String() }

// Exercise the actual executable path, including registry selection, role effects,
// private completion handoffs, JSON-line output, and each saved conversation.
func TestWorkflowRegisteredPipeline(t *testing.T) {
	var mu sync.Mutex
	counts := map[string]int{}
	var order []string
	call := func(id, name string, input any) any {
		return map[string]any{"type": "tool_use", "id": id, "name": name, "input": input}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var request struct {
			System string `json:"system"`
		}
		if err := json.Unmarshal(body, &request); err != nil {
			t.Error(err)
		}
		role := ""
		for _, candidate := range []string{"author", "editor", "reviewer"} {
			if strings.Contains(request.System, "You are the "+candidate+" ") {
				role = candidate
			}
		}

		mu.Lock()
		counts[role]++
		n := counts[role]
		order = append(order, role)
		mu.Unlock()
		parts := []any{map[string]any{"type": "text", "text": "I will carry out this role's registered tools."}}
		switch role {
		case "author":
			if n == 1 {
				parts = append(parts, call("write", "write_draft", map[string]string{"text": "a small draft"}), call("count", "word_count", map[string]any{}), call("think", "think", map[string]any{"seconds": 0, "thought": "ready"}))
			} else if !bytes.Contains(body, []byte(`3`)) || !bytes.Contains(body, []byte(`ready`)) {
				t.Errorf("author did not consume real tools: %s", body)
			}
		case "editor":
			if !bytes.Contains(body, []byte("a small draft")) {
				t.Errorf("author draft missing from handoff: %s", body)
			}
			if n == 1 {
				parts = append(parts, call("read", "read_draft", map[string]any{}), call("edit", "edit_draft", map[string]string{"text": "an improved concise draft"}))
			} else if !bytes.Contains(body, []byte("an improved concise draft")) {
				t.Errorf("editor effect missing: %s", body)
			}
		case "reviewer":
			if !bytes.Contains(body, []byte("an improved concise draft")) {
				t.Errorf("edited draft missing from handoff: %s", body)
			}
			if n == 1 {
				parts = append(parts, call("review", "review", map[string]string{"decision": "accept", "notes": "Checked the edited manuscript."}))
			} else if !bytes.Contains(body, []byte("Checked the edited manuscript.")) {
				t.Errorf("review effect missing: %s", body)
			}
		default:
			t.Errorf("unknown role: %s", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"model": "offline", "content": parts, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "workflow.jsonl")
	env := map[string]string{"LLM_MODEL": "offline", "LLM_API_KEY": "dummy", "LLM_BASE_URL": server.URL, "CH06_LOG": path}
	var stdout lockedOutput
	if err := runIO(strings.NewReader("{\"kind\":\"prompt\",\"text\":\"local test\"}\n"), &stdout, io.Discard, func(k string) string { return env[k] }); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	got := strings.Join(order, ",")
	mu.Unlock()
	if got != "author,author,editor,editor,reviewer,reviewer" {
		t.Fatalf("unreliable stage order: %s", got)
	}
	for _, data := range []string{stdout.text(), func() string {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}()} {
		last := -1
		for _, role := range []string{"author", "editor", "reviewer"} {
			at := strings.Index(data, `"pipeline":"`+role+`_done"`)
			if at <= last {
				t.Fatalf("missing/out-of-order completion for %s: %s", role, data)
			}
			last = at
		}
	}
	for _, role := range []string{"author", "editor", "reviewer"} {
		b, err := os.ReadFile(path + "." + role + ".jsonl")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(b, []byte(`"type":"tool_returned"`)) {
			t.Fatalf("%s lacks saved actual tool result: %s", role, b)
		}
	}
}
