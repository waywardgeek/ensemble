package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// This external peer requests each real application tool. Its later response
// examines the received result, so declarations or final narration cannot pass
// for a working handoff between the three distinct Agents.
func TestWorkflowUsesRoleToolsAndEditedDraft(t *testing.T) {
	var mu sync.Mutex
	counts := map[string]int{}
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			// System carries the role instruction outside ordinary Anthropic messages.
			System string
			// Messages retains provider message order, including tool-result boundaries.
			Messages []struct {
				// Content preserves the provider's ordered blocks or message text.
				Content []struct {
					// Type selects the wire block variant before its other fields are interpreted. Text
					// holds visible provider or user content, including an explicitly empty string. Name
					// is the declared function identifier used to correlate calls and results.
					Type, Text, Name string
					// Content preserves the provider's ordered blocks or message text.
					Content []struct {
						// Text holds visible provider or user content, including an explicitly empty string.
						Text string
					}
				}
			}
			// Tools contains only the declarations visible to this Agent.
			Tools []struct {
				// Name is the declared function identifier used to correlate calls and results.
				Name string
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Error(err)
			return
		}
		role := "reviewer"
		if strings.Contains(in.System, "You are the author.") {
			role = "author"
		}
		if strings.Contains(in.System, "You are the editor.") {
			role = "editor"
		}
		mu.Lock()
		counts[role]++
		round := counts[role]
		mu.Unlock()
		expected := map[string][]string{"author": {"write_draft", "word_count", "think"}, "editor": {"read_draft", "edit_draft"}, "reviewer": {"review"}}[role]
		var names []string
		for _, tool := range in.Tools {
			names = append(names, tool.Name)
		}
		if strings.Join(names, ",") != strings.Join(expected, ",") {
			t.Errorf("%s tools %v", role, names)
		}
		call := ""
		args := map[string]any{}
		switch role {
		case "author":
			if round == 1 {
				call = "write_draft"
				args["text"] = "Ocean draft"
			} else if round == 2 {
				call = "word_count"
			}
		case "editor":
			if round == 1 {
				call = "read_draft"
			} else if round == 2 {
				call = "edit_draft"
				args["text"] = "Ocean draft improved"
			}
		case "reviewer":
			if round == 1 {
				call = "review"
				args = map[string]any{"decision": "accept", "notes": "clear"}
			}
		}
		raw, _ := json.Marshal(in.Messages)
		mu.Lock()
		seen = append(seen, role+":"+string(raw))
		mu.Unlock()
		part := map[string]any{"type": "text", "text": role + " complete"}
		if call != "" {
			part = map[string]any{"type": "tool_use", "id": fmt.Sprintf("%s%d", role, round), "name": call, "input": args}
		}
		json.NewEncoder(w).Encode(map[string]any{"content": []any{part}, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer server.Close()
	dir := t.TempDir()
	binary := filepath.Join(dir, "workflow")
	build := exec.Command("go", "build", "-o", binary, ".")
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %s %v", raw, err)
	}
	command := exec.Command(binary)
	command.Dir = dir
	command.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + dir, "LLM_VENDOR=anthropic", "LLM_MODEL=fake", "LLM_API_KEY=fake", "LLM_BASE_URL=" + server.URL, "CH06_LOG=" + filepath.Join(dir, "observations.jsonl")}
	command.Stdin = strings.NewReader("{\"kind\":\"prompt\",\"text\":\"Ocean\"}\n")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	raw, err := command.Output()
	if err != nil {
		t.Fatalf("workflow: %s %v", stderr.String(), err)
	}
	for _, name := range []string{"author", "editor", "reviewer"} {
		if !bytes.Contains(raw, []byte(`"pipeline":"`+name+`_done"`)) {
			t.Fatalf("missing %s completion", name)
		}
	}
	mu.Lock()
	joined := strings.Join(seen, "\n")
	mu.Unlock()
	for _, want := range []string{`"Text":"2"`, `"Content":[{"Text":"Ocean draft"}]`, `reviewer:[{"Content":[{"Type":"text","Text":"Ocean draft improved"`, "accept: clear"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("handoff/result missing %q: %s", want, joined)
		}
	}
	for _, name := range []string{"author", "editor", "reviewer"} {
		file, err := os.Open(filepath.Join(dir, "observations.jsonl."+name))
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(file)
		file.Close()
		if !bytes.Contains(body, []byte(`"type":"tool_returned"`)) {
			t.Fatal("role did not retain tool result", name)
		}
	}
}
