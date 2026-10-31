package ensemble_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
)

func sandboxAgent(t *testing.T, root string, safe bool) ensemble.Agent {
	t.Helper()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "offline", DataDir: filepath.Join(root, "state"), BuiltinTools: true}, ensemble.AgentSpec{SandboxRoot: root, SafeMode: safe})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := a.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	return a
}
func sandboxCall(t *testing.T, a ensemble.Agent, name string, args map[string]any) (string, error) {
	t.Helper()
	args["ai_callback_delay"] = 5
	raw, _ := json.Marshal(args)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.Invoke(ctx, name, raw)
}

// The real actor/jobs/tools run for both allowed operations and refused escapes.
func TestSandboxFiles(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "outside.txt")
	os.WriteFile(secret, []byte("OUTSIDE_FILE_MARKER"), 0600)
	os.Symlink(outside, filepath.Join(root, "escape"))
	a := sandboxAgent(t, root, false)
	for _, call := range []struct {
		name string
		args map[string]any
		want string
	}{
		{"write_file", map[string]any{"path": "sub/notes.txt", "content": "one apple\ntwo pear\n"}, "wrote"},
		{"read_file", map[string]any{"path": "sub/notes.txt"}, "one apple"},
		{"edit_file", map[string]any{"path": "sub/notes.txt", "old_text": "apple", "new_text": "plum"}, "replaced"},
		{"search_files", map[string]any{"path": "sub", "pattern": "plum"}, "sub/notes.txt:1:one plum"},
		{"list_directory", map[string]any{"path": "sub"}, "notes.txt"},
	} {
		out, err := sandboxCall(t, a, call.name, call.args)
		if err != nil || !strings.Contains(out, call.want) || strings.Contains(out, root) {
			t.Fatalf("allowed %s: %q %v", call.name, out, err)
		}
	}
	rel, _ := filepath.Rel(root, secret)
	for _, path := range []string{secret, rel, "escape/outside.txt", "escape/missing/new.txt"} {
		for _, name := range []string{"read_file", "write_file", "edit_file", "search_files", "list_directory"} {
			out, err := sandboxCall(t, a, name, map[string]any{"path": path, "content": "bad", "overwrite": true, "old_text": "OUTSIDE", "new_text": "bad", "pattern": "."})
			if err == nil {
				t.Errorf("escape accepted %s %s: %q", name, path, out)
			}
			// Absolute input may be repeated; the undisclosed sandbox root may not.
			if strings.Contains(out, root) || strings.Contains(fmt.Sprint(err), root) {
				t.Errorf("host root leaked: %q %v", out, err)
			}
		}
	}
	// Exercise the operation boundary independently: write_file's earlier parent
	// directory check must not hide a broken OpenFile used by state persistence.
	f, openErr := a.Sandbox().OpenFile(secret, os.O_WRONLY|os.O_TRUNC, 0600)
	if openErr == nil {
		f.Close()
		t.Error("direct confined open escaped")
	}
	data, _ := os.ReadFile(secret)
	if string(data) != "OUTSIDE_FILE_MARKER" {
		t.Fatal("outside file changed")
	}
	if _, err := os.Stat(filepath.Join(outside, "missing")); !os.IsNotExist(err) {
		t.Fatal("outside directory created")
	}
	// Canonicalization must accept the same root through a symlink spelling.
	alias := filepath.Join(t.TempDir(), "alias")
	os.Symlink(root, alias)
	if _, err := a.Sandbox().ReadFile(filepath.Join(alias, "sub/notes.txt")); err != nil {
		t.Fatalf("canonical positive: %v", err)
	}
	if _, err := a.Sandbox().ReadFile("sub/absent"); err == nil || strings.Contains(err.Error(), root) {
		t.Fatalf("missing-file diagnostic: %v", err)
	}
}

// Seatbelt must permit ordinary programs and Git while confining descendants.
// A local HTTP server supplies an independent reachable network positive control.
func TestSandboxCommands(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS Seatbelt exercise")
	}
	root, outside := t.TempDir(), t.TempDir()
	secret := filepath.Join(outside, "outside.txt")
	cacheLookalike := filepath.Join(outside, "xcrun_db-private-marker")
	os.WriteFile(cacheLookalike, []byte("OUTSIDE_COMMAND_MARKER"), 0600)
	os.WriteFile(secret, []byte("OUTSIDE_COMMAND_MARKER"), 0600)
	t.Setenv("ODD_CREDENTIAL", "FAKE_ENV_CREDENTIAL_MARKER")
	t.Setenv("SERVICE_TOKEN", "FAKE_TOKEN_MARKER")
	t.Setenv("SSH_AUTH_SOCK", filepath.Join(outside, "socket"))
	t.Setenv("GIT_CONFIG_GLOBAL", secret)
	a := sandboxAgent(t, root, false)
	run := func(script string) string {
		t.Helper()
		out, err := sandboxCall(t, a, "run_command", map[string]any{"command": script})
		if err != nil {
			t.Fatalf("command launch: %v %s", err, out)
		}
		return out
	}
	for _, script := range []string{
		"printf allowed > local.txt && cat local.txt && grep allowed local.txt",
		"python3 -c 'print(open(\"local.txt\").read())'",
		"git init -q repo && cd repo && printf allowed > note && git add note && git commit -qm allowed && git log -1 --format=%s",
	} {
		out := run(script)
		if !strings.Contains(out, "allowed") || !strings.Contains(out, "exit_code: 0") {
			t.Fatalf("positive command: %s", out)
		}
	}
	for _, script := range []string{
		"cat " + secret,
		"cat " + cacheLookalike,
		"printf bad > " + cacheLookalike,
		"sh -c 'sh -c \"cat " + secret + "\"'",
		"printf bad > " + filepath.Join(outside, "written"),
		"cat /etc/hosts",
	} {
		out := run(script)
		if strings.Contains(out, "OUTSIDE_COMMAND_MARKER") || strings.Contains(out, "exit_code: 0") {
			t.Fatalf("kernel escape: %s", out)
		}
	}
	if data, err := os.ReadFile(cacheLookalike); err != nil || string(data) != "OUTSIDE_COMMAND_MARKER" {
		t.Errorf("nested cache lookalike changed: %q %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(outside, "written")); !os.IsNotExist(err) {
		t.Fatal("outside write occurred")
	}
	os.Symlink(outside, filepath.Join(root, "escape"))
	for _, cwd := range []string{"../escape", "escape"} {
		outCwd, errCwd := sandboxCall(t, a, "run_command", map[string]any{"command": "true", "cwd": cwd})
		if errCwd == nil || !strings.Contains(outCwd, cwd) || strings.Contains(outCwd, root) || strings.Contains(outCwd, a.Spec().SandboxRoot) {
			t.Errorf("cwd refusal leaked root: %q %v", outCwd, errCwd)
		}
	}
	outCwd, errCwd := sandboxCall(t, a, "run_command", map[string]any{"command": "cat note", "cwd": "repo"})
	if errCwd != nil || !strings.Contains(outCwd, "allowed") || !strings.Contains(outCwd, "exit_code: 0") {
		t.Fatalf("allowed cwd: %q %v", outCwd, errCwd)
	}

	out := run("env")
	for _, forbidden := range []string{"FAKE_ENV_CREDENTIAL_MARKER", "FAKE_TOKEN_MARKER", "SSH_AUTH_SOCK=", secret} {
		if strings.Contains(out, forbidden) {
			t.Errorf("environment leaked %s", forbidden)
		}
	}
	if !strings.Contains(out, "HOME="+a.Spec().SandboxRoot) {
		t.Fatal("HOME not rooted")
	}
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1); io.WriteString(w, "network-control") }))
	defer server.Close()
	response, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if hits.Load() != 1 {
		t.Fatal("host network control failed")
	}
	out = run("curl -sS --max-time 2 " + server.URL)
	if strings.Contains(out, "network-control") || strings.Contains(out, "exit_code: 0") || hits.Load() != 1 {
		t.Fatalf("command network escape: %s hits=%d", out, hits.Load())
	}
}

// Derivation, unconditional narrowing and the alarm are separate assertions so
// deleting only one mechanism cannot pass by tripping the other first.
func TestSandboxChildSpec(t *testing.T) {
	root := t.TempDir()
	parent := ensemble.AgentSpec{DataDir: "parent", LogPath: "parent.log", SavePath: "parent.save", SandboxRoot: root, SafeMode: true}
	derived := parent.Child("child")
	if derived.DataDir != "child" || derived.LogPath != "" || derived.SavePath != "" || !derived.SafeMode || derived.EnableWebSearch || derived.SandboxRoot != root {
		t.Errorf("derivation: %+v", derived)
	}
	child := ensemble.AgentSpec{EnableWebSearch: true}
	clamped, err := child.Clamp(parent)
	if !clamped.SafeMode || clamped.EnableWebSearch || clamped.SandboxRoot != root {
		t.Errorf("narrowing: %+v", clamped)
	}
	if err == nil {
		t.Error("widening diagnostic absent")
	}
	narrower := parent.Child("child")
	narrower.SandboxRoot = filepath.Join(root, "not-yet-created")
	clamped, err = narrower.Clamp(parent)
	if err != nil || clamped.SandboxRoot != narrower.SandboxRoot {
		t.Errorf("valid narrowing refused: %+v %v", clamped, err)
	}
	alias := filepath.Join(root, "escape")
	os.Symlink(t.TempDir(), alias)
	narrower.SandboxRoot = alias
	clamped, err = narrower.Clamp(parent)
	if clamped.SandboxRoot != root || err == nil {
		t.Error("symlink child widened")
	}
}

// Safe mode removes exec and attendants from the actual registry and dispatch;
// file operations still run with their own confinement.
func TestSandboxSafeMode(t *testing.T) {
	a := sandboxAgent(t, t.TempDir(), true)
	for _, name := range []string{"run_command", "send_input", "wait_for_job", "kill_job", "tool_limits"} {
		for _, d := range a.Declarations() {
			if d.Name == name {
				t.Errorf("safe declaration %s", name)
			}
		}
		if _, err := sandboxCall(t, a, name, map[string]any{"command": "echo BAD"}); err == nil {
			t.Errorf("safe dispatch %s", name)
		}
	}
	out, err := sandboxCall(t, a, "write_file", map[string]any{"path": "safe.txt", "content": "safe-file-control"})
	if err != nil {
		t.Fatal(out, err)
	}
	out, err = sandboxCall(t, a, "read_file", map[string]any{"path": "safe.txt"})
	if err != nil || !strings.Contains(out, "safe-file-control") {
		t.Fatal(out, err)
	}
}

// The real model HTTP client remains outside confinement. Echoed test credentials
// must be sanitized before trace, history, or the model's next context sees them.
func TestSandboxCredentialLogs(t *testing.T) {
	const marker = "FAKE_MODEL_CREDENTIAL_MARKER"
	var authenticated atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated.Store(r.Header.Get("x-api-key") == marker)
		fmt.Fprintf(w, `{"content":[{"type":"text","text":%q}],"usage":{"input_tokens":1,"output_tokens":1}}`, "reply "+marker)
	}))
	defer server.Close()
	var log bytes.Buffer
	root := t.TempDir()
	logFile, err := os.Create(filepath.Join(root, "api.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()
	sink := io.MultiWriter(&log, logFile)
	app := ensemble.New(sink)
	app.SetAPILog(sink)
	a, err := app.NewAgent(ensemble.Config{Model: "offline", APIKey: marker, BaseURL: server.URL, BuiltinTools: true}, ensemble.AgentSpec{SandboxRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	answer, err := a.Ask(context.Background(), "hello")
	if err != nil || !authenticated.Load() {
		t.Fatalf("trusted HTTP: %q %v", answer, err)
	}
	app.Logf("credential diagnostic %s", marker)
	logText, logErr := sandboxCall(t, a, "read_file", map[string]any{"path": "api.log"})
	if logErr != nil || !strings.Contains(logText, "reply") || strings.Contains(logText, marker) {
		t.Fatalf("readable trace credential: %q %v", logText, logErr)
	}
	var history bytes.Buffer
	a.History().Dump(&history)
	if strings.Contains(answer+log.String()+history.String(), marker) {
		t.Fatal("credential reached context or log")
	}
}

// Disabled web access withholds the actual skill and never contacts its peer.
// The enabled control exercises discovery, loading, connection and tool bridging.
func TestSandboxWebGate(t *testing.T) {
	var contacts atomic.Int32
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contacts.Add(1)
		var message struct {
			// ID correlates the external protocol reply.
			ID json.RawMessage
			// Method selects initialization or discovery at the fake peer.
			Method string
		}
		json.NewDecoder(r.Body).Decode(&message)
		if len(message.ID) == 0 {
			w.WriteHeader(202)
			return
		}
		result := `{}`
		if message.Method == "initialize" {
			result = `{"protocolVersion":"2024-11-05"}`
		}
		if message.Method == "tools/list" {
			result = `{"tools":[{"name":"web_probe","inputSchema":{"type":"object"}}]}`
		}
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":%s}`, message.ID, result)
	}))
	defer peer.Close()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "primary"), 0700)
	os.MkdirAll(filepath.Join(dir, "web-search"), 0700)
	os.WriteFile(filepath.Join(dir, "primary", "SKILL.md"), []byte("---\nname: primary\ntype: primary\nloadable-skills: web-search\ntools: read_file\n---\nAvailable: $SKILLS\n"), 0600)
	os.WriteFile(filepath.Join(dir, "web-search", "SKILL.md"), []byte(fmt.Sprintf("---\nname: web-search\ntype: loadable\ntools: web_probe\nmcp_servers:\n  - name: web\n    transport: url\n    url: %s\n---\nUse web_probe.\n", peer.URL)), 0600)
	for _, enabled := range []bool{false, true} {
		a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "offline", BuiltinTools: true}, ensemble.AgentSpec{SandboxRoot: t.TempDir(), SkillsDir: dir, PrimarySkill: "primary", EnableWebSearch: enabled})
		if err != nil {
			t.Fatal(err)
		}
		names := a.Skills().LoadableSkills()
		listed := strings.Contains(strings.Join(names, ","), "web-search")
		_, loadErr := a.Skills().LoadDynamic("web-search", 1)
		if listed != enabled || (loadErr == nil) != enabled {
			t.Errorf("web gate enabled=%v names=%v err=%v", enabled, names, loadErr)
		}
		if !enabled && contacts.Load() != 0 {
			t.Error("disabled skill contacted peer")
		}
		bridged := false
		for _, d := range a.Declarations() {
			if d.Name == "web_probe" {
				bridged = true
			}
		}
		if bridged != enabled {
			t.Errorf("web bridge enabled=%v bridged=%v", enabled, bridged)
		}
		a.Shutdown()
	}
	if contacts.Load() == 0 {
		t.Fatal("enabled peer control never ran")
	}
}

// Model-writable symlinks also target trusted state filenames. Retained job
// output and snapshot backups must use the same confined operations as tools.
func TestSandboxStateFiles(t *testing.T) {
	root := t.TempDir()
	a := sandboxAgent(t, root, false)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	os.WriteFile(outside, []byte("STATE_OUTSIDE_MARKER"), 0600)
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(a.Config().DataDir, "save.json.bak")
	os.Remove(backup)
	os.Symlink(outside, backup)
	if err := a.Save(); err == nil {
		t.Error("snapshot backup followed outside symlink")
	}
	os.MkdirAll(filepath.Join(a.Config().DataDir, "io"), 0700)
	os.Symlink(outside, filepath.Join(a.Config().DataDir, "io", "1"))
	if _, err := sandboxCall(t, a, "read_file", map[string]any{"path": "absent"}); err == nil {
		t.Error("job output symlink accepted")
	}
	data, err := os.ReadFile(outside)
	if err != nil || string(data) != "STATE_OUTSIDE_MARKER" {
		t.Fatalf("state write escaped: %q %v", data, err)
	}
}

// Local media uses the same owner boundary as read_file, before provider encoding.
func TestSandboxMedia(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "media.bin")
	os.WriteFile(outside, []byte("OUTSIDE_MEDIA_MARKER"), 0600)
	os.WriteFile(filepath.Join(root, "media.bin"), []byte("inside media"), 0600)
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o-mini"}, ensemble.AgentSpec{SandboxRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	render := func(path string) ([]byte, error) {
		return a.Engine().Render(&ensemble.Context{Dialogue: []ensemble.Entry{{Actor: ensemble.Human, Parts: []ensemble.Part{{Type: "blob", MIME: "image/png", Ref: ensemble.Ref{Kind: ensemble.RefPath, Locator: path}}}}}})
	}
	body, err := render(filepath.Join(root, "media.bin"))
	if err != nil || !bytes.Contains(body, []byte(base64.StdEncoding.EncodeToString([]byte("inside media")))) {
		t.Fatalf("allowed media missing: %s %v", body, err)
	}
	if body, err := render(outside); err == nil {
		t.Errorf("outside media encoded: %s", body)
	}
}
