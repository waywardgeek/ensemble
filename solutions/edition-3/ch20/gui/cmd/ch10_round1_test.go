package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// The subprocess executes the actual GUI main, with real stdin and a loopback
// HTTP fixture. No browser, external provider, or lossy progress signal is used.
func TestCh10R1GUIProcessEOF(t *testing.T) {
	if os.Getenv("E3_TEST_GUI_CHILD") == "1" {
		flag.CommandLine = flag.NewFlagSet("gui", flag.ExitOnError)
		os.Args = []string{"gui", "-port", "0", "-gui-dir", os.Getenv("E3_TEST_ASSETS")}
		main()
		return
	}
	assets, err := filepath.Abs("../web/gui")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	started := make(chan struct{})
	release := make(chan struct{})
	var requests atomic.Int32
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		if requests.Add(1) == 1 {
			close(started)
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
			fmt.Fprint(w, `{"model":"fixture","content":[{"type":"tool_use","id":"effect-call","name":"write_file","input":{"path":"effect.txt","content":"finished-before-exit"}}],"stop_reason":"tool_use","usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			fmt.Fprint(w, `{"model":"fixture","content":[{"type":"text","text":"FINAL-EOF-ANSWER"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer peer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestCh10R1GUIProcessEOF$")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "E3_TEST_GUI_CHILD=1", "E3_TEST_ASSETS="+assets, "LLM_VENDOR=anthropic", "LLM_MODEL=fixture", "ANTHROPIC_BASE_URL="+peer.URL, "LLM_API_KEY=fake-local-only", "LLM_TOOLS=write_file", "EN_DISABLE_STREAMING=true", "EN_SKILLS_DIR=", "EN_PRIMARY_SKILL=", "ENSEMBLE_SAVE=", "LLM_LOG_PATH=", "LLM_PAUSE_ON_START=false")
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(in, "finish my last admitted prompt\n")
	select {
	case <-started:
	case <-ctx.Done():
		_ = in.Close()
		_ = cmd.Wait()
		t.Fatalf("no request: %s", output.String())
	}
	_ = in.Close()
	time.Sleep(100 * time.Millisecond)
	close(release)
	if err = cmd.Wait(); err != nil {
		t.Fatalf("GUI exit: %v\n%s", err, output.String())
	}
	effect, err := os.ReadFile(filepath.Join(dir, "effect.txt"))
	if err != nil || string(effect) != "finished-before-exit" {
		t.Errorf("last prompt effect: %q %v\n%s", effect, err, output.String())
	}
	raw, err := os.ReadFile(filepath.Join(dir, "save.json"))
	if err != nil {
		t.Fatal(err)
	}
	var save map[string]json.RawMessage
	if err = json.Unmarshal(raw, &save); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(save["context"]), "FINAL-EOF-ANSWER") || !strings.Contains(string(save["log"]), "turn_ended") {
		t.Errorf("incomplete save after EOF: %s\n%s", raw, output.String())
	}
}
