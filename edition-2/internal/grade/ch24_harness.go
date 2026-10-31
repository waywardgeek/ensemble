package grade

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Chapter 24: Freeing the GUI.
//
// The thesis is that the agent framework must not know the GUI exists, and
// that the GUI must be usable by an application that is not the ensemble
// binary. Both halves are graded from OUTSIDE the student's module: a probe
// module with a replace directive is the only honest way to prove that
// internal/ no longer blocks a consumer, and `go list -deps` on that probe
// is the only honest way to prove the framework's dependency closure is
// websocket-free. An in-tree test would compile either way.

const ch24Gorilla = "github.com/gorilla/websocket"

// Ch24Run builds the submission and the external probes, then grades the
// seven published checks.
func Ch24Run(dir string) Ch24Result {
	r := Ch24Result{}

	abs, err := filepath.Abs(dir)
	if err != nil {
		r.failAll("resolving submission dir: %v", err)
		return r
	}
	base := DiscoverBase(abs)
	if base == "" {
		r.failAll("no module path in %s/go.mod", dir)
		return r
	}

	// The submission must build at all before anything is worth grading.
	bin, cleanup, err := Build(abs)
	if err != nil {
		r.failAll("submission does not build: %v", err)
		return r
	}
	defer cleanup()

	probe, probeCleanup, err := ch24WriteProbes(abs, base)
	if err != nil {
		// The probe module failing to assemble is a harness problem for the
		// binary-driven checks, but it IS the verdict for the reuse checks:
		// a consumer outside the module could not be constructed.
		for _, id := range []string{"headless-linkage", "public-reuse",
			"components-without-shell", "mcp-without-gorilla"} {
			r.fail(id, "external probe module could not be assembled: %v", err)
		}
	} else {
		defer probeCleanup()
		ch24LinkageChecks(&r, probe, abs, base)
		ch24ReuseChecks(&r, probe)
	}

	ch24WireCheck(&r, abs)
	ch24AssetsCheck(&r, bin)
	ch24RootCleanCheck(&r, abs, base)
	return r
}

// ch24WriteProbes assembles a Go module OUTSIDE the submission that depends
// on it through a replace directive. Two programs: headless imports only the
// framework root; guiapp imports the gui package and serves it with fake
// hooks. go.sum is copied from the submission so the build resolves from the
// local module cache (GOPROXY=off keeps grading hermetic).
func ch24WriteProbes(abs, base string) (dir string, cleanup func(), err error) {
	dir, err = os.MkdirTemp("", "ch24-probe-")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { os.RemoveAll(dir) }

	gomod := fmt.Sprintf("module ch24probe\n\ngo 1.25\n\nrequire %s v0.0.0-00010101000000-000000000000\n\nreplace %s => %s\n",
		base, base, abs)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		cleanup()
		return "", nil, err
	}
	sum, err := os.ReadFile(filepath.Join(abs, "go.sum"))
	if err == nil {
		err = os.WriteFile(filepath.Join(dir, "go.sum"), sum, 0o644)
	}
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("copying go.sum: %w", err)
	}

	headless := fmt.Sprintf(`package main

import (
	"fmt"

	agent %q
)

func main() {
	_ = agent.AgentSpec{}
	fmt.Println("headless ok")
}
`, base)

	guiapp := fmt.Sprintf(`package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	agent %q
	"%s/gui"
)

func main() {
	sentinel := os.Getenv("CH24_SENTINEL")
	log := &agent.Log{
		Events: []agent.Event{{
			Seq:  1,
			Type: agent.MessageReceived,
			Time: time.Unix(0, 0).UTC(),
			Message: &agent.MessageData{
				Actor: agent.ActorAgent,
				Parts: agent.PartList{agent.TextPart{Text: sentinel}},
			},
		}},
		Next:  2,
		Clock: time.Now,
	}
	srv := gui.New(gui.AgentHooks{
		Send: func(m agent.Inbound) {
			if um, ok := m.(agent.UserMessage); ok {
				fmt.Printf("PROMPT=%%s\n", um.Text)
			}
		},
		EventLog: log,
	}, "")

	js, err := gui.Assets.ReadFile("web/artifact-scroll.js")
	if err != nil {
		fmt.Println("ERR=" + err.Error())
		os.Exit(1)
	}
	if _, err := gui.Assets.ReadFile("web/renderers.js"); err != nil {
		fmt.Println("ERR=" + err.Error())
		os.Exit(1)
	}
	fmt.Println("COMPONENT_OK")

	// The agent-eyes relay: the same consumer exposes the GUI's MCP port.
	// ServeMCP is part of the public reuse surface because an agent that
	// can see this GUI is as much the point as a human who can.
	mln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("ERR=" + err.Error())
		os.Exit(1)
	}
	go srv.ServeMCP(mln)
	fmt.Printf("MCPADDR=%%s\n", mln.Addr().String())

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", srv.ServeWS)
	mux.HandleFunc("/artifact-scroll.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write(js)
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("ERR=" + err.Error())
		os.Exit(1)
	}
	fmt.Printf("ADDR=%%s\n", ln.Addr().String())
	_ = http.Serve(ln, mux)
}
`, base, base)

	for name, src := range map[string]string{"headless": headless, "guiapp": guiapp} {
		sub := filepath.Join(dir, name)
		if err := os.MkdirAll(sub, 0o755); err != nil {
			cleanup()
			return "", nil, err
		}
		if err := os.WriteFile(filepath.Join(sub, "main.go"), []byte(src), 0o644); err != nil {
			cleanup()
			return "", nil, err
		}
	}

	// Both probes must build before their dependency closures mean anything.
	// guiapp is built to a real binary rather than run via `go run`: the
	// grader must be able to kill the server it started, and killing a
	// `go run` wrapper orphans the grandchild, which then holds the
	// grader's output pipe open forever.
	out, err := ch24Go(dir, "build", "-o", "bin/", "./...")
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("probe build failed:\n%s", out)
	}
	return dir, cleanup, nil
}

// ch24Go runs a go command in dir with hermetic module resolution.
func ch24Go(dir string, args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOPROXY=off", "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func ch24Deps(dir, pkg string) (map[string]bool, error) {
	out, err := ch24Go(dir, "list", "-deps", pkg)
	if err != nil {
		return nil, fmt.Errorf("go list -deps %s: %v\n%s", pkg, err, out)
	}
	deps := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		deps[strings.TrimSpace(line)] = true
	}
	return deps, nil
}

// ch24LinkageChecks grades headless-linkage and mcp-without-gorilla: the
// framework closure must carry MCP but neither the GUI nor gorilla, and the
// probe that DOES import the GUI proves the detector can see gorilla at all.
func ch24LinkageChecks(r *Ch24Result, probe, abs, base string) {
	r.ran("headless-linkage")
	r.ran("mcp-without-gorilla")

	headless, err := ch24Deps(probe, "./headless")
	if err != nil {
		r.fail("headless-linkage", "%v", err)
		r.fail("mcp-without-gorilla", "%v", err)
		return
	}
	guiDeps, err := ch24Deps(probe, "./guiapp")
	if err != nil {
		r.fail("headless-linkage", "positive control failed: %v", err)
		return
	}

	// Positive control: a consumer that imports the GUI must visibly link
	// gorilla and the gui package. If it does not, the absence assertions
	// below are blind, not passing.
	if !guiDeps[ch24Gorilla] || !guiDeps[base+"/gui"] {
		r.fail("headless-linkage",
			"positive control failed: the gui-importing probe does not link gorilla (%v) and %s/gui (%v); the absence detector is blind",
			guiDeps[ch24Gorilla], base, guiDeps[base+"/gui"])
		return
	}

	for _, forbidden := range []string{base + "/gui", base + "/mcpws", ch24Gorilla} {
		if headless[forbidden] {
			r.fail("headless-linkage",
				"a consumer importing only the framework root links %s", forbidden)
		}
	}

	if !headless[base+"/internal/mcp"] {
		r.fail("mcp-without-gorilla",
			"%s/internal/mcp is not in the framework closure: MCP support left the framework instead of just the websocket transport", base)
	}
	if headless[ch24Gorilla] {
		r.fail("mcp-without-gorilla",
			"the framework closure still links gorilla: the MCP client websocket transport was not split out")
	}

	// Virtual-user is the positive control for the mcpws split: the dial
	// transport must still exist, still work, and still be the one place
	// that links gorilla by choice.
	vu, err := ch24Deps(abs, "./cmd/virtual-user")
	if err != nil {
		r.fail("mcp-without-gorilla", "listing virtual-user deps: %v", err)
		return
	}
	if !vu[base+"/mcpws"] || !vu[ch24Gorilla] {
		r.fail("mcp-without-gorilla",
			"virtual-user no longer dials through mcpws (links mcpws: %v, gorilla: %v)",
			vu[base+"/mcpws"], vu[ch24Gorilla])
	}
}

// ch24ReuseChecks runs the guiapp probe: a planted event must replay to a
// websocket client, a browser prompt must reach the consumer's Send hook,
// and the Artifact scroll must serve from the exported FS while index.html
// stays absent.
func ch24ReuseChecks(r *Ch24Result, probe string) {
	r.ran("public-reuse")
	r.ran("components-without-shell")

	sentinel := fmt.Sprintf("ch24-replay-%d", time.Now().UnixNano())
	nonce := fmt.Sprintf("ch24-prompt-%d", time.Now().UnixNano())

	cmd := exec.Command(filepath.Join(probe, "bin", "guiapp"))
	cmd.Dir = probe
	cmd.Env = append(os.Environ(), "CH24_SENTINEL="+sentinel)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.fail("public-reuse", "probe stdout: %v", err)
		return
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		r.fail("public-reuse", "starting guiapp probe: %v", err)
		return
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	lines := make(chan string, 64)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	waitLine := func(prefix string, limit time.Duration) (string, bool) {
		deadline := time.After(limit)
		for {
			select {
			case ln, ok := <-lines:
				if !ok {
					return "", false
				}
				if strings.HasPrefix(ln, prefix) {
					return strings.TrimPrefix(ln, prefix), true
				}
			case <-deadline:
				return "", false
			}
		}
	}

	if _, ok := waitLine("COMPONENT_OK", 60*time.Second); !ok {
		r.fail("components-without-shell",
			"probe could not read artifact-scroll.js and renderers.js from the exported Assets FS")
		r.fail("public-reuse", "guiapp probe failed before serving")
		return
	}
	mcpAddr, ok := waitLine("MCPADDR=", 15*time.Second)
	if !ok {
		r.fail("public-reuse", "guiapp probe never reported its MCP relay address")
		return
	}
	addr, ok := waitLine("ADDR=", 15*time.Second)
	if !ok {
		r.fail("public-reuse", "guiapp probe never reported its address")
		return
	}

	// Components without the shell: the component serves, the shell 404s.
	resp, err := http.Get("http://" + addr + "/artifact-scroll.js")
	if err != nil {
		r.fail("components-without-shell", "fetching artifact-scroll.js: %v", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "ArtifactScroll") {
			r.fail("components-without-shell",
				"artifact-scroll.js from the exported FS: status %d, ArtifactScroll marker present: %v",
				resp.StatusCode, strings.Contains(string(body), "ArtifactScroll"))
		}
	}
	if resp, err := http.Get("http://" + addr + "/index.html"); err == nil {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			r.fail("components-without-shell",
				"index.html serves from the component-only consumer: the shell came along")
		}
	}

	// Agent eyes: the MCP relay answers JSON-RPC on a bare TCP listener
	// from the same consumer. This leg runs BEFORE any websocket client
	// connects: with no browser attached the documented answer is an
	// immediate error carrying the request's own id, which proves the
	// relay end to end without a headless browser in the grader. A
	// connected client that ignored MCP frames would instead leave the
	// request waiting forever on a browser reply.
	mcp, err := net.DialTimeout("tcp", mcpAddr, 5*time.Second)
	if err != nil {
		r.fail("public-reuse", "dialing the MCP relay: %v", err)
		return
	}
	defer mcp.Close()
	_ = mcp.SetDeadline(time.Now().Add(10 * time.Second))
	if _, err := mcp.Write([]byte(`{"jsonrpc":"2.0","id":7,"method":"tools/list"}` + "\n")); err != nil {
		r.fail("public-reuse", "writing to the MCP relay: %v", err)
		return
	}
	line, err := bufio.NewReader(mcp).ReadString('\n')
	if err != nil {
		r.fail("public-reuse", "the MCP relay never answered: %v", err)
		return
	}
	var rpc struct {
		Jsonrpc string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Error   json.RawMessage `json:"error"`
		Result  json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal([]byte(line), &rpc); err != nil {
		r.fail("public-reuse", "MCP relay reply is not JSON-RPC: %v in %q", err, line)
		return
	}
	if rpc.Jsonrpc != "2.0" || strings.TrimSpace(string(rpc.ID)) != "7" ||
		(len(rpc.Error) == 0 && len(rpc.Result) == 0) {
		r.fail("public-reuse",
			"MCP relay reply malformed: jsonrpc=%q id=%s error/result present=%v/%v",
			rpc.Jsonrpc, rpc.ID, len(rpc.Error) > 0, len(rpc.Result) > 0)
		return
	}

	// Public reuse, direction one: the planted event replays to a client.
	ws, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		r.fail("public-reuse", "dialing probe websocket: %v", err)
		return
	}
	defer ws.Close()
	if err := ws.WriteJSON(map[string]string{"type": "subscribe"}); err != nil {
		r.fail("public-reuse", "subscribe: %v", err)
		return
	}
	_ = ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		var frame struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := ws.ReadJSON(&frame); err != nil {
			r.fail("public-reuse",
				"the planted event never replayed to the websocket client: %v", err)
			return
		}
		if frame.Type == "message" && strings.Contains(frame.Text, sentinel) {
			break
		}
	}

	// Direction two: a browser prompt reaches the consumer's Send hook.
	if err := ws.WriteJSON(map[string]string{"type": "prompt", "text": nonce}); err != nil {
		r.fail("public-reuse", "sending prompt: %v", err)
		return
	}
	got, ok := waitLine("PROMPT=", 10*time.Second)
	if !ok {
		r.fail("public-reuse", "the browser prompt never reached the consumer's Send hook")
		return
	}
	if got != nonce {
		r.fail("public-reuse", "Send hook received %q, want %q", got, nonce)
	}
}

// ch24WireCheck replays chapter 22's scripted session against the real
// binary: prompts over stdin, a model switch over the GUI websocket, three
// assistant replies. If the wire or the stdin protocol changed, this breaks.
func ch24WireCheck(r *Ch24Result, abs string) {
	r.ran("wire-compat")
	out := ch22Drive(abs)
	if out.fatal != "" {
		r.fail("wire-compat", "%s", out.fatal)
		return
	}
	if len(out.replies) != 3 {
		r.fail("wire-compat", "scripted session produced %d replies, want 3", len(out.replies))
		return
	}
	// The model switch travels over the GUI websocket as an
	// update_settings frame, and the recorded vendor requests are the
	// ground truth that it took effect. The switch lands at a turn
	// boundary rather than instantly, so the assertion is endpoint
	// shaped: the session starts on the dear model and ends on the
	// cheap one. Three replies with no switch is a wire that LOOKS
	// compatible while dropping frames on the floor.
	if len(out.models) < 2 || out.models[0] != ch22ModelDear ||
		out.models[len(out.models)-1] != ch22ModelCheap {
		r.fail("wire-compat",
			"the GUI-socket model switch never reached the vendor: request models %v", out.models)
	}
}

// ch24AssetsCheck starts the submission binary from an empty, unrelated
// working directory and asks for the GUI over HTTP. Before the embed, the
// server could only find its assets relative to a blessed CWD.
func ch24AssetsCheck(r *Ch24Result, bin string) {
	r.ran("assets-embedded")

	srv := fakevendor.New([]fakevendor.Reply{{Text: "unused"}})
	defer srv.Close()

	work, err := os.MkdirTemp("", "ch24-cwd-")
	if err != nil {
		r.fail("assets-embedded", "tempdir: %v", err)
		return
	}
	defer os.RemoveAll(work)

	skills := ch22Skills(work)
	port := freePort()
	cmd := exec.Command(bin, "--port", port)
	cmd.Dir = work
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_MODEL="+ch22ModelDear,
		"LLM_API_KEY=test",
		"EN_SKILL_DIR="+skills,
		"EN_PRIMARY_SKILL=base",
	)
	stdin, _ := cmd.StdinPipe()
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		r.fail("assets-embedded", "starting binary: %v", err)
		return
	}
	defer func() {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	if !ch22WaitPort(port, 10*time.Second) {
		r.fail("assets-embedded", "binary never opened its GUI port")
		return
	}
	for _, probe := range []struct{ path, marker string }{
		{"/", "gui.js"},
		{"/renderers.js", ""},
	} {
		resp, err := http.Get("http://127.0.0.1:" + port + probe.path)
		if err != nil {
			r.fail("assets-embedded", "GET %s: %v", probe.path, err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			r.fail("assets-embedded",
				"GET %s from an unrelated cwd: status %d; the GUI is not embedded",
				probe.path, resp.StatusCode)
			return
		}
		if probe.marker != "" && !strings.Contains(string(body), probe.marker) {
			r.fail("assets-embedded", "GET %s: %q marker missing", probe.path, probe.marker)
			return
		}
	}
}

// ch24RootCleanCheck is the structural half: the framework root's files
// import neither the gui package nor gorilla, and the websocket-era symbols
// are gone. The NewAgent scan is the positive control proving the scanner
// reads the files it claims to.
func ch24RootCleanCheck(r *Ch24Result, abs, base string) {
	r.ran("root-clean")

	entries, err := os.ReadDir(abs)
	if err != nil {
		r.fail("root-clean", "reading submission root: %v", err)
		return
	}
	forbidden := []*regexp.Regexp{
		regexp.MustCompile(`\bWSHub\b`),
		regexp.MustCompile(`\bNewWSHub\b`),
		regexp.MustCompile(`\bNewMCPWSTransport\b`),
	}
	sawNewAgent := false
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(abs, name))
		if err != nil {
			r.fail("root-clean", "reading %s: %v", name, err)
			return
		}
		src := string(data)
		if strings.Contains(src, `"`+base+`/gui"`) {
			r.fail("root-clean", "%s imports %s/gui: the framework root knows the GUI exists", name, base)
		}
		if strings.Contains(src, `"`+ch24Gorilla+`"`) {
			r.fail("root-clean", "%s imports gorilla: the framework root links a websocket library", name)
		}
		for _, re := range forbidden {
			if re.MatchString(src) {
				r.fail("root-clean", "%s still mentions %s: a websocket-era symbol survives at the root", name, re.String())
			}
		}
		if strings.Contains(src, "func NewAgent(") {
			sawNewAgent = true
		}
	}
	if !sawNewAgent {
		r.fail("root-clean",
			"positive control failed: the scanner never saw func NewAgent( in the root package; it is not reading the right files")
	}
}
