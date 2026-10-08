package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/ensemble"
)

// Run the actual command entrypoint in a child test process: this checks both
// terminal JSON and the aggregate process exit without building another CLI.
func TestConsumerProcess(t *testing.T) {
	if os.Getenv("CH09_CONSUMER_PROCESS") != "1" {
		return
	}
	os.Args = []string{"skills-consumer", "--ask"}
	main()
	os.Exit(0)
}

func TestConsumerIndependentTerminalOutcomes(t *testing.T) {
	for _, partial := range []bool{false, true} {
		name := "all-success"
		if partial {
			name = "first-round-limit-second-success"
		}
		t.Run(name, func(t *testing.T) {
			var mu sync.Mutex
			calls := map[string]int{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					System string `json:"system"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					http.Error(w, "bad fixture request", 400)
					return
				}
				agent := "alpha"
				if strings.Contains(body.System, "beta-$TOOLS") {
					agent = "beta"
				} else if !strings.Contains(body.System, "alpha-$TOOLS") {
					t.Error("missing exact project binding")
				}
				mu.Lock()
				calls[agent]++
				n := calls[agent]
				mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				if n > 2 {
					t.Error("per-Agent cap exceeded")
					http.Error(w, "cap", 500)
					return
				}
				content := []map[string]any{}
				if n == 1 || partial && agent == "alpha" {
					content = append(content, map[string]any{"type": "text", "text": agent + " partial observation."})
					content = append(content, map[string]any{"type": "tool_use", "id": fmt.Sprintf("%s-%d", agent, n), "name": "read_file", "input": map[string]any{"path": "notes.txt"}})
				} else {
					content = append(content, map[string]any{"type": "text", "text": agent + " complete report."})
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"model": "fixture", "content": content, "stop_reason": map[bool]string{true: "tool_use", false: "end_turn"}[n == 1 || partial && agent == "alpha"], "usage": map[string]int{"input_tokens": 10, "output_tokens": 2}})
			}))
			defer server.Close()
			workspace := t.TempDir()
			for _, name := range []string{"alpha", "beta"} {
				dir := filepath.Join(workspace, name)
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte(name+" note\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "-test.run=^TestConsumerProcess$")
			cmd.Dir = workspace
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, "LLM_") && !strings.HasPrefix(entry, "EN_DISABLE_STREAMING=") && !strings.HasPrefix(entry, "CH09_CONSUMER_PROCESS=") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, "CH09_CONSUMER_PROCESS=1", "LLM_VENDOR=anthropic", "LLM_MODEL=fixture", "LLM_API_KEY=local-fixture", "LLM_BASE_URL="+server.URL, "LLM_SYSTEM=", "EN_DISABLE_STREAMING=1")
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if ctx.Err() != nil {
				t.Fatal("consumer failed to finish independent bounded attempts")
			}
			if partial {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 1 {
					t.Fatalf("wanted aggregate exit1, got %v; %s", err, stderr.String())
				}
			} else if err != nil {
				t.Fatalf("success failed: %v; %s", err, stderr.String())
			}
			records := []terminalRecord{}
			for _, line := range bytes.Split(bytes.TrimSpace(stdout.Bytes()), []byte{'\n'}) {
				var fields map[string]json.RawMessage
				if err := json.Unmarshal(line, &fields); err != nil {
					t.Fatalf("bad terminal JSON: %v", err)
				}
				if _, ok := fields["Completion"]; !ok {
					continue
				}
				var record terminalRecord
				if err := json.Unmarshal(line, &record); err != nil {
					t.Fatal(err)
				}
				records = append(records, record)
			}
			if len(records) != 2 {
				t.Fatalf("wanted both terminal records, got %d: %s", len(records), stdout.String())
			}
			for i, record := range records {
				agent := []string{"alpha", "beta"}[i]
				c := record.Completion
				if c.AgentID != record.Agent || c.AgentID == "" || c.RequestID == "" {
					t.Fatal("terminal identity missing")
				}
				if i == 1 && record.Agent == records[0].Agent {
					t.Fatal("Agents conflated")
				}
				if c.Usage != (ensemble.Usage{Input: 20, Output: 4}) {
					t.Fatalf("partial/complete usage lost: %+v", c.Usage)
				}
				if len(c.Parts) == 0 || partial && i == 0 && !strings.Contains(c.Text, agent+" partial observation.") {
					t.Fatal("available partial data lost")
				}
				if partial && i == 0 {
					if c.Outcome != "round_limit" || c.Error == nil || c.Error.Code != "round_limit" || record.Error == "" {
						t.Fatalf("partial mislabeled: %+v", record)
					}
				} else if c.Outcome != "success" || c.Error != nil || record.Error != "" || !strings.Contains(c.Text, agent+" complete report.") {
					t.Fatalf("success mislabeled: %+v", record)
				}
				raw, err := os.ReadFile(filepath.Join(workspace, agent, "events.jsonl"))
				if err != nil {
					t.Fatal(err)
				}
				sent := 0
				for _, line := range bytes.Split(raw, []byte{'\n'}) {
					if len(line) == 0 {
						continue
					}
					var e ensemble.Event
					if err := json.Unmarshal(line, &e); err != nil {
						t.Fatal(err)
					}
					if e.Type == "request_sent" {
						sent++
					}
					if e.Type == "turn_started" && (e.Turn.Policy.MaxModelRequests != 2 || e.Turn.Policy.EffectiveMaxModelRequests != 2) {
						t.Fatal("request policy changed")
					}
				}
				if sent != 2 {
					t.Fatalf("expected2 bounded requests, got%d", sent)
				}
			}
			mu.Lock()
			defer mu.Unlock()
			if calls["alpha"] != 2 || calls["beta"] != 2 {
				t.Fatalf("independent attempts missing or retried: %v", calls)
			}
			t.Logf("terminal outcomes: %s/%s; each2 HTTP; aggregate error=%t", records[0].Completion.Outcome, records[1].Completion.Outcome, err != nil)
		})
	}
}
