package gradeedition2

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter19 retains the seven original sandbox weights. The submitted CLI and
// actual OS enforce the boundaries; the only fake is the external model peer.
// library is the public package path relative to the submitted module root.
func Chapter19(bin, core, library string) (grade.Report, error) {
	return chapter19(bin, core, library, false)
}

// Chapter19CLI uses the restarted root library and isolated CLI fixture while
// preserving all seven checks and the historical entry point above.
func Chapter19CLI(bin, core string) (grade.Report, error) {
	return chapter19(bin, core, ".", true)
}

func chapter19(bin, core, library string, current bool) (grade.Report, error) {
	checks := grade.Ch23Checks(grade.Ch23Result{})
	for i := range checks {
		checks[i].Passed = true
		checks[i].Earned = checks[i].Points
		checks[i].Details = nil
	}
	if runtime.GOOS != "darwin" {
		for i := range checks {
			fail(&checks[i], "this exercise requires macOS Seatbelt; no unconfined substitute")
		}
		return grade.NewTitledReport("Edition 2, Chapter 19 — The Sandbox", checks, ""), nil
	}
	sandboxChecks(bin, checks, current)
	sandboxSafe(bin, &checks[5], current)
	sandboxChild(core, library, &checks[4], current)
	return grade.NewTitledReport("Edition 2, Chapter 19 — The Sandbox", checks, ""), nil
}

type sandboxFixture struct{ base, root, outside, skills, nonce, credential string }

func newSandboxFixture() (f sandboxFixture, err error) {
	f.base, err = os.MkdirTemp("", "sandbox-grade-")
	if err != nil {
		return
	}
	f.root = filepath.Join(f.base, "work")
	f.outside = filepath.Join(f.base, "vault")
	f.nonce = filepath.Base(f.base)
	f.credential = "sk-course-credential-" + f.nonce
	for _, p := range []string{f.root, f.outside, filepath.Join(f.base, "data")} {
		if err = os.MkdirAll(p, 0700); err != nil {
			return
		}
	}
	f.skills = filepath.Join(f.base, "skills")
	err = recallWrite(f.skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Sandbox exercise\ntype: primary\ntools: read_file write_file edit_file list_directory search_files run_command send_input wait_for_job kill_job tool_limits\n---\nPerform the requested local operations.\n"})
	if err != nil {
		return
	}
	for p, text := range map[string]string{filepath.Join(f.root, "notes.txt"): "INSIDE_" + f.nonce, filepath.Join(f.outside, "file.txt"): "FILE_SECRET_" + f.nonce, filepath.Join(f.outside, "command.txt"): "COMMAND_SECRET_" + f.nonce} {
		if err = os.WriteFile(p, []byte(text), 0600); err != nil {
			return
		}
	}
	err = os.Symlink(f.outside, filepath.Join(f.root, "escape"))
	return
}

type sandboxProbe struct {
	id, tool string
	args     map[string]any
}

func (p sandboxProbe) reply() fakevendor.Reply {
	raw, _ := json.Marshal(p.args)
	return skillCall(p.id, p.tool, string(raw))
}

var sandboxExit = regexp.MustCompile(`(?i)exit_code[: ]+(-?\d+)`)

// Command failure is normally a completed job with a nonzero exit, not a failed
// tool invocation. Require a correlated completed observation in either form.
func sandboxCommand(c *grade.Check, observation sandboxObservation, id string, success bool) string {
	text, isError, err := observation.result()
	if err != nil || text == "" {
		fail(c, "command %s has no correlated result: %v", id, err)
		return text
	}
	terminal := text
	if len(observation) != 0 {
		last := observation[len(observation)-1]
		terminal, _, _ = anthropicToolText(last.request.Body, last.id)
	}
	m := sandboxExit.FindStringSubmatch(terminal)
	if len(m) == 2 {
		code, _ := strconv.Atoi(m[1])
		if (code == 0) != success {
			fail(c, "command %s exit %d, success wanted %v", id, code, success)
		}
	} else if success || !isError {
		fail(c, "command %s did not report a completed exit or refusal: %q", id, text)
	}
	if success && isError {
		fail(c, "allowed command %s refused: %s", id, text)
	}
	return text
}

func sandboxChecks(bin string, checks []grade.Check, current bool) {
	sandboxChecksFor(bin, checks, current, false)
}

func sandboxChecksFor(bin string, checks []grade.Check, current, executable bool) {
	path, kernel, network, credentials, leak := &checks[0], &checks[1], &checks[2], &checks[3], &checks[6]
	failed := func(err error) {
		for _, c := range []*grade.Check{path, kernel, network, credentials, leak} {
			fail(c, "sandbox session: %v", err)
		}
	}
	f, err := newSandboxFixture()
	if f.base != "" && !executable {
		defer os.RemoveAll(f.base)
	}
	if err != nil {
		failed(err)
		return
	}
	// A reachable host control uses the very same target kept alive for the child.
	// Absence of an arbitrary curl marker alone would not establish confinement.
	var hits atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1); fmt.Fprint(w, "NETWORK_"+f.nonce) }))
	defer target.Close()
	response, err := http.Get(target.URL)
	if err != nil {
		failed(err)
		return
	}
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil || string(body) != "NETWORK_"+f.nonce || hits.Load() != 1 {
		failed(fmt.Errorf("reachable-target control failed"))
		return
	}
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
	// The API trace starts with the full tool schema. Request an explicit file
	// and supervisor output budget so that a truncated prefix cannot masquerade
	// as evidence about the seeded response later in the readable log.
	probes := []sandboxProbe{
		{"read", "read_file", map[string]any{"path": "notes.txt"}},
		{"write", "write_file", map[string]any{"path": "local.txt", "content": "LOCAL_" + f.nonce}},
		{"edit", "edit_file", map[string]any{"path": "local.txt", "old_text": "LOCAL_", "new_text": "EDITED_"}},
		{"edited", "read_file", map[string]any{"path": "local.txt"}},
		{"search", "search_files", map[string]any{"path": ".", "pattern": "EDITED_" + f.nonce}},
		{"list", "list_directory", map[string]any{"path": "."}},
		{"relative", "read_file", map[string]any{"path": "../vault/file.txt"}},
		{"absolute", "read_file", map[string]any{"path": filepath.Join(f.outside, "file.txt")}},
		{"symlink", "read_file", map[string]any{"path": "escape/file.txt"}},
		{"write-escape", "write_file", map[string]any{"path": "escape/file.txt", "content": "BAD", "overwrite": true}},
		{"command", "run_command", map[string]any{"command": "printf COMMAND_OK > command.txt && cat command.txt", "ai_callback_delay": 5}},
		{"python", "run_command", map[string]any{"command": "/usr/bin/python3 -c 'print(open(\"command.txt\").read())'", "ai_callback_delay": 5}},
		{"git", "run_command", map[string]any{"command": sandboxGitCommand(current), "ai_callback_delay": 5}},
		{"command-escape", "run_command", map[string]any{"command": "cat " + q(filepath.Join(f.outside, "command.txt")), "ai_callback_delay": 5}},
		{"descendant", "run_command", map[string]any{"command": "sh -c " + q("sh -c "+q("cat "+q(filepath.Join(f.outside, "command.txt")))), "ai_callback_delay": 5}},
		{"command-write", "run_command", map[string]any{"command": "printf BAD > " + q(filepath.Join(f.outside, "created.txt")), "ai_callback_delay": 5}},
		{"network", "run_command", map[string]any{"command": "curl -sS --max-time 2 " + q(target.URL), "ai_callback_delay": 5}},
		{"environment", "run_command", map[string]any{"command": "env", "ai_callback_delay": 5}},
		{"cwd", "run_command", map[string]any{"command": "true", "cwd": "escape", "ai_callback_delay": 5}},
	}
	if executable {
		for i := range probes {
			if probes[i].id == "git" {
				probes[i].args["command"] = sandboxExecutableGitCommand
			}
		}
	}
	probes = append(probes, sandboxLogProbes(current)...)
	reqs, result, err := sandboxLaunchFor(bin, f, false, probes, current, executable)
	if err != nil {
		failed(err)
		return
	}
	for id, want := range map[string]string{"read": "INSIDE_" + f.nonce, "edited": "EDITED_" + f.nonce, "search": "EDITED_" + f.nonce, "list": "local.txt"} {
		if !strings.Contains(result[id].skill(path, id, false), want) {
			fail(path, "allowed %s lost actual content", id)
		}
	}
	for _, id := range []string{"write", "edit"} {
		result[id].skill(path, id, false)
	}
	if b, e := os.ReadFile(filepath.Join(f.root, "local.txt")); e != nil || string(b) != "EDITED_"+f.nonce {
		fail(path, "actual allowed edit not present: %q %v", b, e)
	}
	for _, id := range []string{"relative", "absolute", "symlink", "write-escape"} {
		result[id].skill(path, id, true)
	}
	if b, e := os.ReadFile(filepath.Join(f.outside, "file.txt")); e != nil || string(b) != "FILE_SECRET_"+f.nonce {
		fail(path, "outside file changed: %q %v", b, e)
	}
	for id, want := range map[string]string{"command": "COMMAND_OK", "python": "COMMAND_OK", "git": "git-control"} {
		if !strings.Contains(sandboxCommand(kernel, result[id], id, true), want) {
			fail(kernel, "allowed %s lost actual output", id)
		}
	}
	for _, id := range []string{"command-escape", "descendant", "command-write"} {
		sandboxCommand(kernel, result[id], id, false)
	}
	if _, e := os.Stat(filepath.Join(f.outside, "created.txt")); !os.IsNotExist(e) {
		fail(kernel, "outside command write created a file: %v", e)
	}
	sandboxCommand(network, result["network"], "network", false)
	if hits.Load() != 1 {
		fail(network, "child reached host-controlled HTTP target, hits=%d", hits.Load())
	}
	env := sandboxCommand(credentials, result["environment"], "environment", true)
	if !strings.Contains(env, "HOME=") {
		fail(credentials, "environment positive observation absent")
	}
	if current && !executable {
		result["log-limit"].skill(credentials, "log-limit", false)
	}
	log := result["log"].skill(credentials, "log", false)
	if !strings.Contains(log, "CREDENTIAL_ECHO_") {
		fail(credentials, "actual readable API trace did not contain earlier response")
	}
	// Secrets were never put in call arguments. Scan real transmitted context and
	// readable files after success, not just a formatter's claimed refusal.
	for _, r := range reqs {
		for marker, c := range map[string]*grade.Check{"FILE_SECRET_" + f.nonce: path, "COMMAND_SECRET_" + f.nonce: kernel} {
			if strings.Contains(string(r.Body), marker) {
				fail(c, "protected marker reached provider context")
			}
		}
	}
	if strings.Contains(env, f.credential) {
		fail(credentials, "credential reached command environment result")
	}
	if strings.Contains(log, f.credential) {
		fail(credentials, "credential reached successful readable API-log result")
	}
	for id, asked := range map[string]string{"relative": "../vault/file.txt", "symlink": "escape/file.txt", "cwd": "escape"} {
		text := result[id].skill(leak, id, true)
		if !strings.Contains(text, asked) {
			fail(leak, "refusal %s lost caller-supplied path %q", id, asked)
		}
		if strings.Contains(text, f.base) {
			fail(leak, "refusal %s disclosed constructed host root", id)
		}
	}
}

func sandboxSafe(bin string, c *grade.Check, current bool) {
	sandboxSafeFor(bin, c, current, false)
}

func sandboxSafeFor(bin string, c *grade.Check, current, executable bool) {
	f, err := newSandboxFixture()
	if f.base != "" && !executable {
		defer os.RemoveAll(f.base)
	}
	if err != nil {
		fail(c, "%v", err)
		return
	}
	probes := []sandboxProbe{{"write", "write_file", map[string]any{"path": "safe.txt", "content": "SAFE_" + f.nonce}}, {"read", "read_file", map[string]any{"path": "safe.txt"}}}
	// Preserve the original four-tool contract. The accepted implementation also
	// withholds tool_limits, but that extra choice is not a fifth numeric rule.
	banned := []string{"run_command", "send_input", "wait_for_job", "kill_job"}
	for _, name := range banned {
		probes = append(probes, sandboxProbe{name, name, map[string]any{"command": "printf BAD > escaped.txt", "handle": 1}})
	}
	reqs, _, err := sandboxLaunchFor(bin, f, true, probes, current, executable)
	if err != nil {
		fail(c, "safe session: %v", err)
		return
	}
	for _, r := range reqs {
		names := map[string]bool{}
		for _, v := range list(r.JSON()["tools"]) {
			names[fmt.Sprint(object(v)["name"])] = true
		}
		for _, name := range banned {
			if names[name] {
				fail(c, "safe mode advertised %s", name)
			}
		}
		if !names["read_file"] || !names["write_file"] {
			fail(c, "safe mode lost file positive declarations")
		}
	}
	skillResult(c, reqs[1], "write", false)
	if !strings.Contains(skillResult(c, reqs[2], "read", false), "SAFE_"+f.nonce) {
		fail(c, "safe file result absent")
	}
	for i, name := range banned {
		if executable {
			sandboxExecutableRefusal(c, reqs[i+3], name)
		} else {
			sandboxSafeRefusal(c, reqs[i+3], name, current)
		}
	}
	if b, e := os.ReadFile(filepath.Join(f.root, "safe.txt")); e != nil || string(b) != "SAFE_"+f.nonce {
		fail(c, "actual safe write missing")
	}
	if _, e := os.Stat(filepath.Join(f.root, "escaped.txt")); !os.IsNotExist(e) {
		fail(c, "guessed disabled command executed")
	}
}

// The restarted file schema separates file bytes from one-shot observation limits.
// Keep the historical direct-argument fixture unchanged for its original caller.
func sandboxLogProbes(current bool) []sandboxProbe {
	if current {
		return []sandboxProbe{
			{"log-limit", "tool_limits", map[string]any{"max_output_bytes": 65536}},
			{"log", "read_file", map[string]any{"path": "api.log", "max_bytes": 65536}},
		}
	}
	return []sandboxProbe{{"log", "read_file", map[string]any{"path": "api.log", "max_bytes": 65536, "max_output_bytes": 65536}}}
}

func sandboxSafeRefusal(c *grade.Check, r fakevendor.Recorded, id string, current bool) {
	if !current {
		webUnknown(c, r, id)
		return
	}
	// Chapter 19 requires absence, refusal and nonexecution, not Chapter 18's
	// unknown-tool wording. The surrounding check verifies the other two facts.
	if strings.TrimSpace(skillResult(c, r, id, true)) == "" {
		fail(c, "%s has no safe-mode refusal explanation", id)
	}
}
