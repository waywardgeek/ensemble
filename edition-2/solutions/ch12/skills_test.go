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
	"sync/atomic"
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

// TestGUIDebugStartup exercises the shipped flags and real skill/MCP/job loop.
// The two fakes are external peers: a stateful stdio browser and an HTTP vendor.
func TestGUIDebugStartup(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 unavailable")
	}
	dir := t.TempDir()
	peer := `import json,sys
state='SNAPSHOT_ENABLED'
for line in sys.stdin:
 m=json.loads(line); method=m.get('method')
 if 'id' not in m: continue
 if method=='initialize': result={'protocolVersion':'2024-11-05','capabilities':{'tools':{}}}
 elif method=='tools/list': result={'tools':[{'name':n,'description':n,'inputSchema':{'type':'object'},'ephemeral':mode} for n,mode in [('gui_snapshot','round'),('tts_queue','round'),('gui_click',''),('gui_input','')]]}
 elif method=='tools/call':
  name=m['params']['name']; args=m['params'].get('arguments',{})
  if name=='gui_click': state='SNAPSHOT_DISABLED'; text='clicked '+args['selector']
  elif name=='gui_input': text='set '+args['selector']+' to '+args['text']
  elif name=='gui_snapshot': text=state
  else: text='[{"text":"PENDING_SPEECH","state":"queued"}]'
  result={'content':[{'type':'text','text':text}]}
 else: continue
 print(json.dumps({'jsonrpc':'2.0','id':m['id'],'result':result}),flush=True)
`
	args, _ := json.Marshal([]string{"-u", "-c", peer})
	for name, body := range map[string]string{
		"ensemble":  "---\nname: ensemble\ntype: primary\nloadable-skills: gui-debug\n---\nPrimary manual",
		"gui-debug": "---\nname: gui-debug\ntype: loadable\ndepends: ensemble\ntools: gui_snapshot tts_queue gui_click gui_input\nmcp_servers:\n  - name: browser-debug\n    transport: stdio\n    command: " + python + "\n    args: " + string(args) + "\n---\nBROWSER_STARTUP_MANUAL",
	} {
		if err := os.Mkdir(filepath.Join(dir, name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name, "SKILL.md"), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	var rounds atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			// System must retain the startup skill's manual as well as its tools.
			System string `json:"system"`
			// Messages preserve the actual rendered ephemera and tool results.
			Messages json.RawMessage `json:"messages"`
			// Tools are declarations sent over the external provider boundary.
			Tools []struct {
				// Name identifies callable capabilities visible to the model.
				Name string `json:"name"`
			} `json:"tools"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		round := rounds.Add(1)
		text := string(body.Messages)
		if !strings.Contains(body.System, "BROWSER_STARTUP_MANUAL") {
			t.Error("startup manual missing")
		}
		names := map[string]bool{}
		for _, tool := range body.Tools {
			names[tool.Name] = true
		}
		if names["gui_snapshot"] || names["tts_queue"] {
			t.Error("automatic browser tools became callable")
		}
		if round <= 3 {
			if !names["gui_click"] || !names["gui_input"] || !strings.Contains(text, "PENDING_SPEECH") {
				t.Errorf("browser skill not active at round %d: %s", round, text)
			}
			want := "SNAPSHOT_DISABLED"
			if round == 1 {
				want = "SNAPSHOT_ENABLED"
			}
			if !strings.Contains(text, want) || (round > 1 && strings.Contains(text, "SNAPSHOT_ENABLED")) {
				t.Errorf("stale or missing round %d snapshot: %s", round, text)
			}
		} else if names["gui_click"] || names["gui_input"] || strings.Contains(text, "PENDING_SPEECH") || strings.Contains(text, "SNAPSHOT_") {
			t.Errorf("unloaded browser capability survived: %s", text)
		}
		if round == 3 && !strings.Contains(text, "set #search-box to chapter13") {
			t.Error("gui_input result lost its arguments")
		}
		name, input := "", map[string]string{}
		switch round {
		case 1:
			name, input = "gui_click", map[string]string{"selector": "#button-A"}
		case 2:
			name, input = "gui_input", map[string]string{"selector": "#search-box", "text": "chapter13"}
		case 3:
			name, input = "unload_skill", map[string]string{"name": "gui-debug"}
		}
		content := []any{map[string]any{"type": "text", "text": "browser exercise finished"}}
		if name != "" {
			content = []any{map[string]any{"type": "tool_use", "id": fmt.Sprint(round), "name": name, "input": input}}
		}
		json.NewEncoder(w).Encode(map[string]any{"content": content, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer server.Close()
	binary := filepath.Join(t.TempDir(), "ensemble")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
	command := exec.Command(binary, "--gui-debug", "--skills-dir", dir, "prompt", "exercise the browser")
	command.Dir = t.TempDir()
	// The flag must override the environment directory. Primary defaults to the
	// shipped ensemble skill; callers can still choose another identity explicitly.
	command.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL=fake", "LLM_API_KEY=test", "LLM_BASE_URL="+server.URL, "EN_SKILLS_DIR=/missing", "EN_PRIMARY_SKILL=")
	output, err := command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "browser exercise finished") || rounds.Load() != 4 {
		t.Fatalf("startup: %v rounds=%d output=%s", err, rounds.Load(), output)
	}
}
