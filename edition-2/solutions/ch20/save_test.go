package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// The subprocess owns startup, EOF and replacement. Fake only the provider;
// the public composition root, actor, history and client all run unchanged.
func TestCLISaveLifecycle(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "ensemble")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	var mu sync.Mutex
	var requests [][]byte
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		requests = append(requests, body)
		mu.Unlock()
		io.WriteString(w, `{"model":"fake","content":[{"type":"tool_use","id":"submit","name":"submit","input":{"result":"remembered"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer provider.Close()
	for _, override := range []bool{false, true} {
		name := "default"
		if override {
			name = "override"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			args := []string{"chat"}
			save := filepath.Join(dir, "save.json")
			if override {
				save = filepath.Join(dir, "chosen.json")
				args = append(args, "--save", save)
			}
			run := func(prompt string) ([]byte, error) {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, binary, args...)
				cmd.Dir = dir
				cmd.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL=fake", "LLM_API_KEY=private-test-key", "LLM_BASE_URL="+provider.URL, "EN_DISABLE_STREAMING=1", "EN_PRIMARY_SKILL=", "EN_SKILLS_DIR=")
				cmd.Stdin = strings.NewReader(prompt + "\n")
				return cmd.CombinedOutput()
			}
			if out, err := run("remember cobalt"); err != nil || !bytes.Contains(out, []byte("remembered")) {
				t.Fatalf("first chat: %v %s", err, out)
			}
			data, err := os.ReadFile(save)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(data, &fields); err != nil || len(fields) != 4 {
				t.Fatalf("four-field save: %s %v", data, err)
			}
			var config struct {
				// Model identifies the request model recorded at save time.
				Model string `json:"model"`
				// Vendor records the translator using its public disk spelling.
				Vendor string `json:"vendor"`
				// SystemPrompt distinguishes a present empty prompt from an omitted field.
				SystemPrompt *string `json:"system_prompt"`
				// Tools exposes exact declaration field names for the portable save.
				Tools []map[string]json.RawMessage `json:"tools"`
			}
			if err := json.Unmarshal(fields["config"], &config); err != nil {
				t.Fatal(err)
			}
			if config.Model != "fake" || config.Vendor != "anthropic" || config.SystemPrompt == nil || len(config.Tools) == 0 || config.Tools[0]["name"] == nil || config.Tools[0]["description"] == nil || config.Tools[0]["schema"] == nil {
				t.Fatalf("save configuration shape: %s", fields["config"])
			}
			mu.Lock()
			var request struct {
				// System is the actual Anthropic role instruction received at the boundary.
				System []struct {
					// Text is the actual instruction content, excluding cache metadata.
					Text string `json:"text"`
				} `json:"system"`
			}
			decodeErr := json.Unmarshal(requests[len(requests)-1], &request)
			mu.Unlock()
			if decodeErr != nil || len(request.System) != 1 || *config.SystemPrompt != request.System[0].Text {
				t.Fatalf("saved prompt differs from actual default wire instruction: saved=%q wire=%q err=%v", *config.SystemPrompt, request.System, decodeErr)
			}
			if bytes.Contains(data, []byte("private-test-key")) || bytes.Contains(data, []byte(provider.URL)) {
				t.Fatal("save leaked transport configuration")
			}
			if out, err := run("what word"); err != nil {
				t.Fatalf("second chat: %v %s", err, out)
			}
			mu.Lock()
			latest := append([]byte(nil), requests[len(requests)-1]...)
			mu.Unlock()
			if !bytes.Contains(latest, []byte("remember cobalt")) || !bytes.Contains(latest, []byte("what word")) {
				t.Fatalf("second process forgot history: %s", latest)
			}
			if override {
				if _, err := os.Stat(filepath.Join(dir, "save.json")); !os.IsNotExist(err) {
					t.Fatal("override also wrote default path")
				}
			}
			bad := []byte(`{"config":`)
			if err := os.WriteFile(save, bad, 0600); err != nil {
				t.Fatal(err)
			}
			mu.Lock()
			before := len(requests)
			mu.Unlock()
			out, err := run("must not send")
			if err == nil || !bytes.Contains(out, []byte(save)) {
				t.Fatalf("bad save not refused by name: %v %s", err, out)
			}
			after, _ := os.ReadFile(save)
			mu.Lock()
			count := len(requests)
			mu.Unlock()
			if !bytes.Equal(bad, after) || count != before {
				t.Fatal("bad save overwritten or provider called after refusal")
			}
		})
	}
}
