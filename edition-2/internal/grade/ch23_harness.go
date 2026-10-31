package grade

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Chapter 23 asks for a sandbox: file tools that cannot escape a root, commands
// confined by the kernel rather than by a Go function, and credentials that
// never reach the conversation.
//
// Why this grader drives the real binary rather than calling the library.
// Chapter 21 shipped a feature that every structural check passed and that was
// dead on arrival, because the skill file never listed its tool. A sandbox has
// the same failure mode and a worse consequence: a student can write a perfect
// path resolver, unit test it in isolation, and never call it from the tool
// that opens the file. The library test passes. The agent reads /etc/passwd.
// Every isolation a test buys is a thing it stops observing, and the thing
// worth observing here is the running program.
//
// Why the assertions are phrased as sentinels. The chapter's thesis is that the
// asset under protection is the context window, because that is the only
// surface prompt injection reaches. So the grader plants a distinctive string
// in a file outside the sandbox and asserts it never appears anywhere in the
// recorded conversation. That is the real property stated directly: not "the
// function returned an error" but "the secret never reached the model". A
// student whose error text differs from ours still passes, which is what the
// chapter's own Yours clause promises.
//
// Every escape probe is paired with a positive control. A sandbox that refuses
// everything is trivial to build and worthless to use. The controls are not
// decoration: while building this chapter's implementation, the first profile
// that appeared to confine perfectly was in fact denying every operation,
// including the legitimate ones, and it looked like a complete success.
const (
	ch23Inside = "CH23-INSIDE-OK" // content of a file inside the root
	// Two secrets, not one. The file tools and the kernel profile are separate
	// boundaries, and a single shared sentinel cannot say which of them leaked:
	// the grader would have to fail both checks on one observation, so a mutant
	// that disabled only the file tools would appear to kill the kernel check
	// as well. Separate sentinels make each check name the mechanism it guards.
	ch23FileSecret = "CH23-FILE-SECRET" // reachable only by a file tool
	ch23CmdSecret  = "CH23-CMD-SECRET"  // reachable only by a command
	ch23CmdOK      = "CH23-CMD-OK"      // a command that should succeed
	ch23EnvClean   = "CH23-ENV-CLEAN"   // printed when the credential is absent
	ch23Cred       = "sk-ch23-7f3a-LEAKED"
	ch23NetOK      = "CH23-NET-REACHED" // printed when the network is reachable
)

// Ch23Run builds the student's agent and interrogates it.
func Ch23Run(dir string) Ch23Result {
	var r Ch23Result

	bin, cleanup, err := Build(dir)
	if err != nil {
		r.Fatal = fmt.Sprintf("the agent does not build: %v", err)
		return r
	}
	defer cleanup()

	ch23DriveSandbox(bin, &r)
	ch23DriveSafeMode(bin, &r)
	ch23ProbeChild(dir, &r)
	return r
}

// ch23Fixture is the world the sandboxed agent runs in: a root it may touch and
// a secret it may not.
type ch23Fixture struct {
	base       string
	root       string
	fileSecret string // absolute path, reachable only by a file tool
	cmdSecret  string // absolute path, reachable only by a command
	dataDir    string
}

func ch23NewFixture() (ch23Fixture, error) {
	base, err := os.MkdirTemp("", "ch23grade")
	if err != nil {
		return ch23Fixture{}, err
	}
	f := ch23Fixture{
		base:       base,
		root:       filepath.Join(base, "work"),
		fileSecret: filepath.Join(base, "vault", "file.txt"),
		cmdSecret:  filepath.Join(base, "vault", "cmd.txt"),
		dataDir:    filepath.Join(base, "data"),
	}
	for _, d := range []string{f.root, filepath.Join(base, "vault"), f.dataDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return f, err
		}
	}
	if err := os.WriteFile(filepath.Join(f.root, "notes.txt"), []byte(ch23Inside+"\n"), 0o644); err != nil {
		return f, err
	}
	if err := os.WriteFile(f.fileSecret, []byte(ch23FileSecret+"\n"), 0o600); err != nil {
		return f, err
	}
	if err := os.WriteFile(f.cmdSecret, []byte(ch23CmdSecret+"\n"), 0o600); err != nil {
		return f, err
	}
	return f, nil
}

// ch23SandboxChecks are the checks the sandboxed session decides. They are
// listed once, because any failure to run that session must fail all of them.
// An earlier version of this harness failed only the first and let the other
// four pass, which meant a session that never launched scored fifty five points
// out of sixty five. A check that cannot be decided must not be reported as
// passed.
var ch23SandboxChecks = []string{
	"path-confinement",
	"kernel-confinement",
	"no-network",
	"no-credentials",
	"no-host-path-leak",
}

func (r *Ch23Result) failAll(ids []string, format string, args ...any) {
	for _, id := range ids {
		r.fail(id, format, args...)
	}
}

// ch23DriveSandbox runs one session under --sandbox and answers checks 1, 2, 3,
// 4 and 7. One session serves five checks because each probe is independent and
// a session costs a process launch; the probes are ordered so that the positive
// controls run first, which means a total failure to launch shows up as a
// control failure rather than as five mysterious passes.
func ch23DriveSandbox(bin string, r *Ch23Result) {
	for _, id := range ch23SandboxChecks {
		r.ran(id)
	}

	f, err := ch23NewFixture()
	if err != nil {
		r.failAll(ch23SandboxChecks, "could not build the fixture: %v", err)
		return
	}
	defer os.RemoveAll(f.base)

	// The network probe aims at a local HTTP server the grader owns. That keeps
	// the check hermetic: no internet access is required to grade, and the
	// agent's own successful traffic to the fake vendor is the positive control
	// proving the binary still has a network while its children do not.
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer target.Close()
	netURL := target.URL

	probes := []struct{ name, tool, args string }{
		{"read inside", "read_file", ch23Args(map[string]string{"path": "notes.txt"})},
		{"command inside", "run_command", ch23Args(map[string]string{"command": "echo " + ch23CmdOK})},
		{"read relative escape", "read_file", ch23Args(map[string]string{"path": "../vault/file.txt"})},
		{"read absolute escape", "read_file", ch23Args(map[string]string{"path": f.fileSecret})},
		{"command escape", "run_command", ch23Args(map[string]string{"command": "cat " + f.cmdSecret})},
		{"grandchild escape", "run_command", ch23Args(map[string]string{"command": `sh -c "cat ` + f.cmdSecret + `"`})},
		{"network", "run_command", ch23Args(map[string]string{"command": "curl -s -o /dev/null -m 5 " + netURL + " && echo " + ch23NetOK})},
		{"credential", "run_command", ch23Args(map[string]string{"command": "printenv OPENAI_API_KEY || echo " + ch23EnvClean})},
	}

	replies := make([]fakevendor.Reply, 0, len(probes)+1)
	for i, p := range probes {
		replies = append(replies, fakevendor.Reply{
			ToolName: p.tool,
			ToolArgs: p.args,
			ToolID:   fmt.Sprintf("ch23_%d", i),
		})
	}
	replies = append(replies, fakevendor.Reply{Text: "done"})

	srv := fakevendor.New(replies)
	defer srv.Close()

	env := append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"OPENAI_API_KEY="+ch23Cred,
		"ANTHROPIC_API_KEY="+ch23Cred,
	)
	if err := ch23Launch(bin, env, []string{"--sandbox", f.root}, f.dataDir); err != nil {
		r.failAll(ch23SandboxChecks, "the agent did not complete a sandboxed session: %v", err)
		return
	}

	results := ch23ToolResults(srv.Requests())
	if len(results) == 0 {
		r.failAll(ch23SandboxChecks, "the agent produced no tool results, so nothing could be observed")
		return
	}

	// Positive controls first. If these fail the escape results are
	// meaningless, because a sandbox that blocks everything blocks these too.
	if !ch23Observed(results, ch23Inside) {
		r.fail("path-confinement", "read_file could not read a file inside the sandbox root: a sandbox that refuses legitimate work is not a sandbox")
	}
	if !ch23Observed(results, ch23CmdOK) {
		r.fail("kernel-confinement", "run_command could not run a command inside the sandbox root: a sandbox that refuses legitimate work is not a sandbox")
	}

	// The escapes. Each sentinel names the boundary that should have stopped it.
	if ch23Observed(results, ch23FileSecret) {
		r.fail("path-confinement", "a file tool read %s, which is outside the sandbox root", f.fileSecret)
	}
	if ch23Observed(results, ch23CmdSecret) {
		r.fail("kernel-confinement", "a command or a descendant of a command read %s, which is outside the sandbox root", f.cmdSecret)
	}

	if ch23Observed(results, ch23NetOK) {
		r.fail("no-network", "a sandboxed command reached the network at %s", netURL)
	}

	if ch23Observed(results, ch23Cred) {
		r.fail("no-credentials", "the credential reached the conversation; a tool result carried it")
	} else if !ch23Observed(results, ch23EnvClean) {
		r.fail("no-credentials", "the credential probe produced neither the credential nor the marker that proves it was stripped, so the check could not be decided")
	}

	// The refusal must name what the model asked for, not where the host keeps
	// it. Scoped to the refusal itself: a system prompt that tells the model its
	// root is a reasonable design and must not fail this check.
	if msg, ok := ch23RefusalFor(srv.Requests(), "../vault/file.txt"); !ok {
		r.fail("no-host-path-leak", "no refusal was recorded for the relative escape, so the message could not be inspected")
	} else if strings.Contains(msg, f.root) || strings.Contains(msg, f.base) {
		r.fail("no-host-path-leak", "the refusal revealed the host path: %q", msg)
	}
}

// ch23DriveSafeMode answers check 6. Safe mode is an absence, and an absence is
// invisible unless something asserts it, which is why the check insists the
// file tools are still present: that half is the positive control. Section 6 of
// the coder review records a real bug found exactly this way, where tool
// removal silently did nothing for every tool whose name contained an
// underscore and safe mode reported itself enabled while run_command stayed
// fully callable.
func ch23DriveSafeMode(bin string, r *Ch23Result) {
	r.ran("safe-mode-absence")

	f, err := ch23NewFixture()
	if err != nil {
		r.fail("safe-mode-absence", "could not build the fixture: %v", err)
		return
	}
	defer os.RemoveAll(f.base)

	srv := fakevendor.New([]fakevendor.Reply{{Text: "done"}})
	defer srv.Close()

	env := append(os.Environ(), "LLM_BASE_URL="+srv.URL(), "OPENAI_API_KEY="+ch23Cred)
	args := []string{"--sandbox", f.root, "--safe-mode"}
	if err := ch23Launch(bin, env, args, f.dataDir); err != nil {
		r.fail("safe-mode-absence", "the agent did not complete a safe mode session: %v", err)
		return
	}

	reqs := srv.Requests()
	if len(reqs) == 0 {
		r.fail("safe-mode-absence", "the agent never called the model, so the advertised tools could not be observed")
		return
	}
	tools := ch23AdvertisedTools(reqs[0].Body)
	if len(tools) == 0 {
		r.fail("safe-mode-absence", "the request advertised no tools at all, so the check could not be decided")
		return
	}
	for _, banned := range []string{"run_command", "send_input", "kill_job", "wait_for_job"} {
		if tools[banned] {
			r.fail("safe-mode-absence", "safe mode still advertised %s; removing a tool from the prompt is not the same as removing it from the registry", banned)
		}
	}
	// The positive control. Without it, deleting every tool would score full
	// marks on this check.
	if !tools["read_file"] {
		r.fail("safe-mode-absence", "safe mode removed read_file as well; the point is to withhold execution, not to disarm the agent")
	}
}

// ch23ProbeChild answers check 5, which has two halves that must fail
// independently. The wall is the clamp, applied unconditionally, so that
// deleting the diagnostic cannot widen a child. The alarm is the error, so that
// an agent which asks for more than it has is told. A grader that tested only
// the error would pass an implementation whose entire security boundary was one
// early return.
//
// This is the only check that compiles against the student's package rather
// than driving the binary, because sub-agent spawning does not exist yet: Child
// and Clamp have no runtime path to observe. Both signatures are printed in the
// chapter, so they are contract rather than implementation detail.
func ch23ProbeChild(dir string, r *Ch23Result) {
	r.ran("child-cannot-widen")

	module, err := ch23ModulePath(dir)
	if err != nil {
		r.fail("child-cannot-widen", "could not read the student go.mod: %v", err)
		return
	}

	probeDir := filepath.Join(dir, "ch23gradeprobe")
	if err := os.MkdirAll(probeDir, 0o755); err != nil {
		r.fail("child-cannot-widen", "could not plant the probe: %v", err)
		return
	}
	defer os.RemoveAll(probeDir)

	src := strings.Replace(ch23ProbeSource, "STUDENT_MODULE", module, 1)
	if err := os.WriteFile(filepath.Join(probeDir, "main.go"), []byte(src), 0o644); err != nil {
		r.fail("child-cannot-widen", "could not write the probe: %v", err)
		return
	}

	cmd := exec.Command("go", "run", "./ch23gradeprobe")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	text := string(out)
	if !strings.Contains(text, "CH23-PROBE-") {
		r.fail("child-cannot-widen", "the probe did not compile or run against AgentSpec.Child and AgentSpec.Clamp: %v\n%s", err, ch23Tail(text))
		return
	}
	if !strings.Contains(text, "CH23-PROBE-CLAMPED") {
		r.fail("child-cannot-widen", "a child spec that asked for wider permissions than its parent was not clamped; the effective permissions must narrow even if the error is never read\n%s", ch23Tail(text))
	}
	if !strings.Contains(text, "CH23-PROBE-REPORTED") {
		r.fail("child-cannot-widen", "widening was clamped but never reported; an agent that asks for more than it has must be told\n%s", ch23Tail(text))
	}
	if !strings.Contains(text, "CH23-PROBE-INHERITS") {
		r.fail("child-cannot-widen", "a child derived from its parent did not inherit the parent's permissions\n%s", ch23Tail(text))
	}
}

// ch23ProbeSource is planted inside the student's module so that its imports
// resolve. It prints markers rather than returning an exit code, so a partial
// failure still says which half broke.
const ch23ProbeSource = `package main

import (
	"errors"
	"fmt"

	"STUDENT_MODULE"
)

func main() {
	parent := agent.AgentSpec{
		DataDir:         "/parent",
		SandboxRoot:     "/parent/work",
		SafeMode:        true,
		EnableWebSearch: false,
	}

	inherited := parent.Child("/parent/work/kid")
	if inherited.SafeMode && !inherited.EnableWebSearch {
		fmt.Println("CH23-PROBE-INHERITS")
	}

	greedy := parent.Child("/parent/work/kid")
	greedy.SafeMode = false
	greedy.EnableWebSearch = true

	got, err := greedy.Clamp(parent)
	if got.SafeMode && !got.EnableWebSearch {
		fmt.Println("CH23-PROBE-CLAMPED")
	}
	if err != nil {
		fmt.Println("CH23-PROBE-REPORTED")
	}
	_ = errors.Is
	fmt.Println("CH23-PROBE-DONE")
}
`

// ch23Launch runs the agent through one prompt and waits for it to finish. The
// agent reads newline delimited JSON on stdin and writes it on stdout, which is
// the interface chapter 14 built for the virtual user. The working directory
// matters: the agent writes its data directory relative to where it runs, so
// each session gets its own.
func ch23Launch(bin string, env, args []string, workdir string) error {
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	cmd.Dir = workdir
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
		for sc.Scan() {
			var msg struct {
				Assistant string `json:"assistant"`
			}
			if json.Unmarshal(sc.Bytes(), &msg) == nil && msg.Assistant != "" {
				done <- nil
				return
			}
		}
		done <- fmt.Errorf("the agent produced no assistant turn; stderr: %s", ch23Tail(stderr.String()))
	}()

	fmt.Fprintln(stdin, `{"kind":"prompt","text":"go"}`)

	var result error
	select {
	case result = <-done:
	case <-time.After(120 * time.Second):
		result = fmt.Errorf("the agent did not answer within 120s; stderr: %s", ch23Tail(stderr.String()))
	}
	stdin.Close()
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	return result
}

// ch23ToolResults collects the content of every tool result in the recorded
// conversation, in order.
//
// Searching tool results rather than whole request bodies is not a detail, it
// is the difference between a grader that observes and one that lies. An
// earlier version of this harness searched the entire body, and three of its
// sentinels appeared in the probe's own command string: the network probe ran
// "curl ... && echo CH23-NET-REACHED", so the marker was present in the tool
// call arguments whether or not curl ever succeeded. The grader reported a
// network escape that had not happened, and two positive controls passed
// without proving anything at all. A tool call is the model's words. Only a
// tool result is an observation.
func ch23ToolResults(reqs []fakevendor.Recorded) []string {
	var out []string
	seen := map[string]bool{}
	for _, q := range reqs {
		var body any
		if json.Unmarshal(q.Body, &body) != nil {
			continue
		}
		var walk func(any)
		walk = func(n any) {
			switch v := n.(type) {
			case map[string]any:
				t, _ := v["type"].(string)
				role, _ := v["role"].(string)
				if t == "tool_result" || role == "tool" {
					if s, ok := v["content"].(string); ok && !seen[s] {
						seen[s] = true
						out = append(out, s)
					}
				}
				if t == "functionResponse" {
					if s, ok := v["response"].(string); ok && !seen[s] {
						seen[s] = true
						out = append(out, s)
					}
				}
				for _, c := range v {
					walk(c)
				}
			case []any:
				for _, c := range v {
					walk(c)
				}
			}
		}
		walk(body)
	}
	return out
}

// ch23Observed reports whether a sentinel appears in anything the agent
// observed, as opposed to anything the model was scripted to say.
func ch23Observed(results []string, sentinel string) bool {
	for _, s := range results {
		if strings.Contains(s, sentinel) {
			return true
		}
	}
	return false
}

// ch23Conversation concatenates every recorded request body. Used only where
// the question really is about the whole conversation.
func ch23Conversation(reqs []fakevendor.Recorded) string {
	var b strings.Builder
	for _, q := range reqs {
		b.Write(q.Body)
		b.WriteByte('\n')
	}
	return b.String()
}

// ch23RefusalFor finds the tool result that answered a call carrying the given
// argument. It walks the JSON rather than pattern matching the text, so a
// student's own error wording does not affect it.
func ch23RefusalFor(reqs []fakevendor.Recorded, arg string) (string, bool) {
	for _, q := range reqs {
		var body any
		if json.Unmarshal(q.Body, &body) != nil {
			continue
		}
		var found string
		var walk func(any)
		walk = func(n any) {
			switch v := n.(type) {
			case map[string]any:
				if t, _ := v["type"].(string); t == "tool_result" {
					if s, ok := v["content"].(string); ok && strings.Contains(s, arg) {
						found = s
					}
				}
				if role, _ := v["role"].(string); role == "tool" {
					if s, ok := v["content"].(string); ok && strings.Contains(s, arg) {
						found = s
					}
				}
				for _, c := range v {
					walk(c)
				}
			case []any:
				for _, c := range v {
					walk(c)
				}
			}
		}
		walk(body)
		if found != "" {
			return found, true
		}
	}
	return "", false
}

// ch23AdvertisedTools reads the tool names a request offered the model. Reading
// the wire is the point: a tool withheld from the prompt but left in the
// registry is still callable, and a tool left in the prompt but removed from the
// registry is an error waiting to happen.
func ch23AdvertisedTools(body []byte) map[string]bool {
	out := map[string]bool{}
	var doc any
	if json.Unmarshal(body, &doc) != nil {
		return out
	}
	var walk func(any)
	walk = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			if name, ok := v["name"].(string); ok {
				if _, hasSchema := v["input_schema"]; hasSchema {
					out[name] = true
				}
				if _, hasParams := v["parameters"]; hasParams {
					out[name] = true
				}
			}
			if fn, ok := v["function"].(map[string]any); ok {
				if name, ok := fn["name"].(string); ok {
					out[name] = true
				}
			}
			for _, c := range v {
				walk(c)
			}
		case []any:
			for _, c := range v {
				walk(c)
			}
		}
	}
	walk(doc)
	return out
}

// ch23ModulePath reads the student's module path so the planted probe imports
// the student's own package rather than a name this grader assumed.
func ch23ModulePath(dir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(b), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(rest), nil
		}
	}
	return "", fmt.Errorf("no module line in go.mod")
}

func ch23Args(m map[string]string) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func ch23Tail(s string) string {
	if len(s) > 1500 {
		return "..." + s[len(s)-1500:]
	}
	return s
}
