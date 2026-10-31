package gui_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// The real executable must select the file before attempting to parse invalid
// inherited environment settings. Start from another cwd to exercise paths too.
func TestFileConfiguredEntryIgnoresEnvironment(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "gui")
	if output, err := exec.Command("go", "build", "-o", executable, "./cmd").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	requests := make(chan map[string]any, 1)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		if r.Header.Get("Authorization") != "Bearer fixture-key" {
			t.Error("wrong credential source")
		}
		requests <- body
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"File config works."}]}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer provider.Close()
	path := applicationFixture(t)
	data, _ := os.ReadFile(path)
	var config map[string]any
	json.Unmarshal(data, &config)
	config["provider"].(map[string]any)["base_url"] = provider.URL + "/v1"
	config["provider"].(map[string]any)["reasoning_summary"] = "detailed"
	config["disable_streaming"] = true
	config["recall"] = map[string]any{"disabled": true}
	data, _ = json.Marshal(config)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	// Flag help/errors must never print a loaded API credential as a default.
	for _, option := range []string{"--help", "--not-a-real-flag"} {
		output, _ := exec.Command(executable, "--config", path, option).CombinedOutput()
		if len(output) == 0 {
			t.Fatal("missing flag diagnostic")
		}
		if strings.Contains(string(output), "fixture-key") {
			t.Fatal("credential leaked into flag output")
		}
	}
	socket, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := socket.Addr().(*net.TCPAddr).Port
	socket.Close()
	cmd := exec.Command(executable, "--config", path, "--port", strconv.Itoa(port))
	cmd.Dir = t.TempDir()
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	poison := map[string]string{"LLM_VENDOR": "invalid-vendor", "LLM_MODEL": "wrong-model", "OPENAI_BILLING_MODE": "wrong-route", "LLM_BASE_URL": "http://127.0.0.1:1", "OPENAI_REASONING_SUMMARY": "wrong-summary", "EN_PRIMARY_SKILL": "does-not-exist", "EN_SKILLS_DIR": "/missing", "EN_HOME": "/missing-auth-home", "OPENAI_OIDC_ISSUER": "invalid-issuer", "EN_CODEX_AUTH_FILE": "/missing-token", "LLM_API_KEY": "wrong-key"}
	for _, entry := range os.Environ() {
		name := strings.SplitN(entry, "=", 2)[0]
		if _, drop := poison[name]; !drop {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	for name, value := range poison {
		cmd.Env = append(cmd.Env, name+"="+value)
	}
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	var once sync.Once
	stop := func() {
		once.Do(func() {
			cmd.Process.Signal(os.Interrupt)
			select {
			case <-exited:
			case <-time.After(5 * time.Second):
				cmd.Process.Kill()
				<-exited
			}
		})
	}
	t.Cleanup(stop)
	address := "localhost:" + strconv.Itoa(port)
	deadline := time.Now().Add(10 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			ready = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !ready {
		stop()
		t.Fatalf("configured GUI did not listen:\n%s", output.String())
	}
	peer := dial(t, "http://"+address)
	events := until(t, peer, kind("current_settings"))
	settings := events[len(events)-1]["settings"].(map[string]any)
	app, ok := settings["application"].(map[string]any)
	if !ok || app["provider"].(map[string]any)["reasoning_summary"] != "detailed" {
		t.Fatalf("wrong config: %v", settings)
	}
	send(t, peer, "prompt", "Check configuration.")
	events = until(t, peer, kind("turn_ended"))
	if events[len(events)-1]["text"] != "File config works." {
		t.Fatalf("unexpected response: %v", events[len(events)-1])
	}
	select {
	case body := <-requests:
		reasoning, ok := body["reasoning"].(map[string]any)
		if !ok || reasoning["summary"] != "detailed" || reasoning["effort"] != "high" || body["model"] != "gpt-6-astra" {
			t.Fatalf("wrong request settings: %v", body)
		}
	case <-time.After(time.Second):
		t.Fatal("no provider request")
	}
	stop()
	if !strings.Contains(output.String(), "provider environment ignored") {
		t.Fatal("launch source not reported")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "save.json")); err != nil {
		t.Fatal("save path did not resolve against config file:", err)
	}
}
