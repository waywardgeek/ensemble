package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Exercise the published prompt command, including the environment-to-AgentSpec
// boundary. The fake is only the external vendor; the compiled CLI remains real.
func TestPromptSkillEnvironment(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "base"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "base", "SKILL.md"), []byte("---\nname: base\ntype: primary\ntools: read_file\n---\nIdentity $CUSTOM_VAR"), 0644); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["system"] != "Identity cli-value" {
			t.Errorf("CLI primary: %v", body["system"])
		}
		fmt.Fprint(w, `{"content":[{"type":"text","text":"prompt worked"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	binary := filepath.Join(t.TempDir(), "ensemble")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
	cmd := exec.Command(binary, "prompt", "hello")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL=fake", "LLM_API_KEY=test", "LLM_BASE_URL="+server.URL, "EN_SKILLS_DIR="+dir, "EN_PRIMARY_SKILL=base", "EN_CUSTOM_VAR=cli-value")
	output, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "prompt worked" {
		t.Fatalf("prompt command: %q %v", output, err)
	}
}
