package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter14Executable uses the C# executable's public host contract. The eight
// behavioral weights remain; historical inline tools and Go topology are not
// requirements of this path. Fakes replace providers and the external MCP peer.
func Chapter14Executable(bin string) (grade.Report, error) {
	r := contextChecksFor(bin, true)
	checks := grade.Ch15Checks(grade.Ch15Result{Fatal: r.Fatal, Errs: r.Errs})
	checks[4].Title = "C# SYSTEM manuals, dynamic declarations, actual MCP invocation and unload revocation"
	return grade.NewTitledReport("Edition 3, Chapter 14 — C# executable context checks", checks, ""), nil
}

// This host accepts prompts on its public socket, not stdin. A joined SIGTERM
// is its ordinary graceful server shutdown; SIGKILL is reserved for crash tests.
func contextExecutableLaunch(bin, skillsDir, guiDir string, o contextOpts) (out contextOut) {
	srv := fakevendor.New(o.replies)
	defer srv.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return contextOut{fatal: err.Error()}
	}
	port := fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
	listener.Close()
	args := []string{"--port", port, "--gui-dir", guiDir, "--no-stream",
		"--vendor", "anthropic", "--model", "offline-context-fixture", "--base-url", srv.URL(),
		"--skills-dir", skillsDir, "--primary-skill", "base",
		"--event-log", filepath.Join(o.dir, "events.jsonl"), "--settings", filepath.Join(o.dir, "settings.json"),
		"--save", filepath.Join(o.dir, "save.json"), "--save-log", "-1"}
	if o.startupManual {
		args = append(args, "--startup-skills", "manual")
	}
	if o.model == contextStubModel {
		args = append(args, "--stub-tool-results")
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = o.dir
	// The outer grader runs in an isolated HOME/TMP environment. Explicit model
	// flags and dummy credentials keep all provider traffic on this loopback peer.
	cmd.Env = append(os.Environ(), "LLM_API_KEY=fake-key", "EN_STARTUP_SKILLS=", "EN_STUB_TOOL_RESULTS=0",
		"EN_DEFERRED_TOOLS=0", "EN_DISABLE_SAVE=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err = cmd.Start(); err != nil {
		return contextOut{fatal: err.Error()}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	finished := false
	defer func() {
		if !finished {
			cmd.Process.Kill()
			<-done
		}
		out.reqs = srv.Requests()
		if encoded, e := json.Marshal(out.reqs); e == nil {
			os.WriteFile(filepath.Join(o.dir, "provider-requests.json"), encoded, 0600)
		}
		// Preserve useful failure evidence rather than hide it behind a socket timeout.
		os.WriteFile(filepath.Join(o.dir, "host.stdout"), stdout.Bytes(), 0600)
		os.WriteFile(filepath.Join(o.dir, "host.stderr"), stderr.Bytes(), 0600)
		if out.fatal != "" {
			out.fatal += ": " + stderr.String()
		}
	}()
	var peer *guiPeer
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		peer, err = openGUIPeer("ws://127.0.0.1:" + port + "/ws")
		if err == nil {
			break
		}
		select {
		case e := <-done:
			finished = true
			return contextOut{fatal: fmt.Sprintf("GUI exited before ready: %v", e)}
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		return contextOut{fatal: "GUI socket: " + err.Error()}
	}
	defer peer.close()
	send := func(kind, text string) error {
		return peer.conn.WriteJSON(map[string]string{"type": kind, "text": text, "agent_id": "root"})
	}
	if err = send("subscribe", ""); err != nil {
		return contextOut{fatal: err.Error()}
	}
	if _, err = readSettings(peer, "current_settings"); err != nil {
		return contextOut{fatal: err.Error()}
	}
	target := o.target
	if target == 0 {
		target = 400000
	}
	if err = peer.conn.WriteJSON(map[string]any{"type": "update_settings", "agent_id": "root",
		"settings": map[string]int{"context_target_bytes": target}}); err != nil {
		return contextOut{fatal: err.Error()}
	}
	settings, err := readSettings(peer, "settings_updated")
	if err != nil || settings["context_target_bytes"] != float64(target) {
		return contextOut{fatal: fmt.Sprintf("context target not applied: %v (%v)", settings, err)}
	}
	for i, prompt := range o.prompts {
		if err = send("prompt", prompt); err != nil {
			return contextOut{fatal: err.Error()}
		}
		messages, e := peer.until(30*time.Second, messageKind("turn_ended"))
		if e != nil {
			return contextOut{fatal: fmt.Sprintf("prompt %d completion: %v", i+1, e)}
		}
		ending := messages[len(messages)-1]
		if problem, ok := ending["error"].(string); ok && problem != "" {
			return contextOut{fatal: "turn failed: " + problem}
		}
		want := i + 1
		if i < len(o.expect) {
			want = o.expect[i]
		}
		if n := len(srv.Requests()); n != want {
			return contextOut{fatal: fmt.Sprintf("prompt %d made %d requests, want %d", i+1, n, want)}
		}
	}
	if o.kill {
		if err = cmd.Process.Kill(); err != nil {
			return contextOut{fatal: err.Error()}
		}
		err = <-done
		finished = true
		exit, ok := err.(*exec.ExitError)
		if !ok {
			return contextOut{fatal: fmt.Sprintf("SIGKILL exit was not observed: %v", err)}
		}
		status, ok := exit.Sys().(syscall.WaitStatus)
		if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
			return contextOut{fatal: fmt.Sprintf("expected SIGKILL, got %v", err)}
		}
		return out
	}
	if err = cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return contextOut{fatal: err.Error()}
	}
	select {
	case err = <-done:
		finished = true
		if err != nil {
			out.fatal = "nonzero graceful GUI exit: " + err.Error()
		}
	case <-time.After(15 * time.Second):
		out.fatal = "GUI did not save and exit after SIGTERM"
	}
	return out
}

func contextExecutablePrefix(run func(contextOpts) contextOut, dir string, r *contextResult) {
	const name = "frozen-prefix"
	replies := []fakevendor.Reply{
		contextCall("F1", "load_skill", `{"name":"manual"}`),
		contextCall("F2", "load_skill", `{"name":"mcp-tools"}`),
		contextCall("F3", "gui_input", `{"selector":"#search-box","text":"INPUT_ARGUMENT_ONLY"}`),
		contextCall("F4", "unload_skill", `{"name":"mcp-tools"}`),
		contextCall("F5", "gui_input", `{"selector":"#search-box","text":"MUST_NOT_RUN"}`),
		contextText("MCP lifecycle complete."),
	}
	out := run(contextOpts{dir: dir, model: contextStubModel, prompts: []string{"Load the manual, use MCP, then unload it."}, expect: []int{6}, replies: replies})
	if out.fatal != "" || len(out.reqs) != 6 {
		r.fail(name, "session: %s (requests %d)", out.fatal, len(out.reqs))
		return
	}
	for i, q := range out.reqs {
		req, err := contextParse(q.Body)
		if err != nil {
			r.fail(name, "request %d: %v", i+1, err)
			return
		}
		if msg := contextInstructions(req, i > 0); msg != "" {
			r.fail(name, "request %d: %s", i+1, msg)
			return
		}
		var tools []struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Tools, &tools); err != nil {
			r.fail(name, "tools: %v", err)
			return
		}
		names := map[string]bool{}
		for _, tool := range tools {
			if names[tool.Name] {
				r.fail(name, "duplicate declaration %s", tool.Name)
				return
			}
			names[tool.Name] = true
		}
		if names["list_directory"] != (i > 0) {
			r.fail(name, "request %d manual declaration mismatch", i+1)
			return
		}
		for _, tool := range []string{"gui_snapshot", "tts_queue", "gui_click", "gui_input"} {
			if names[tool] != (i >= 2 && i < 4) {
				r.fail(name, "request %d declaration %s mismatch", i+1, tool)
				return
			}
		}
		for _, tool := range []string{"read_file", "load_skill", "unload_skill", "micro_handoff", "keep_tool_results"} {
			if !names[tool] {
				r.fail(name, "request %d lost base control %s", i+1, tool)
				return
			}
		}
		if strings.Contains(guiDebugText(q), "SNAPSHOT_ONLY_") || strings.Contains(guiDebugText(q), "TTS_ONLY_") {
			r.fail(name, "request %d contains unsolicited observations", i+1)
			return
		}
		if i >= 2 && strings.Count(string(req.System), "Tools from the fake MCP server.") != 1 {
			r.fail(name, "request %d lost or duplicated retained MCP manual", i+1)
			return
		}
	}
	text, failed, err := anthropicToolText(out.reqs[3].Body, "F3")
	if err != nil || failed || !strings.Contains(text, "INPUT_RESULT_ONLY:INPUT_ARGUMENT_ONLY") {
		r.fail(name, "correlated input result missing: %v %t %q", err, failed, text)
		return
	}
	_, failed, err = anthropicToolText(out.reqs[5].Body, "F5")
	if err != nil || !failed {
		r.fail(name, "post-unload call was not refused: %v", err)
		return
	}
	r.set(name, "")
}
