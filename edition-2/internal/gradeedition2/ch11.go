package gradeedition2

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter11 grades actual MCP protocol, model use and optional GUI transport.
// Native browser effects, ownership and job lifetime retain separate source and
// integration evidence; the external peers here replace only protocol boundaries.
func Chapter11(bin, core, gui, guiModel, consumer, workflow, exercise, streamModel string, example ConsumerCase, media map[string]string) (grade.Report, error) {
	checks := mcpChecks(bin, gui, guiModel)
	// Persistence keeps its direct startup/restart checks and its own narrow
	// independent-session parity launcher. MCP adds no alternate runtime wrapper.
	previous, err := Chapter10(bin, core, gui, guiModel, consumer, workflow, exercise, streamModel, example, media)
	if err != nil {
		return grade.Report{}, err
	}
	parity := grade.Check{ID: "ch10-parity", Title: "All corrected Chapter 10 checks pass", Points: 10, Earned: 10, Passed: true}
	for _, c := range previous.Checks {
		if !c.Passed {
			fail(&parity, "%s: %s", c.ID, strings.Join(c.Details, "; "))
		}
	}
	return grade.NewTitledReport("Edition 2, Chapter 11 — MCP", append(checks, parity), ""), nil
}

// Chapter11Executable keeps the actual CLI protocol and model-use probes.
// The optional GUI is a separate C# host and has its own real browser evidence;
// this report does not award that check or inherited Go topology/consumer credit.
func Chapter11Executable(bin string) grade.Report {
	return grade.NewTitledReport("Edition 3 C#, Chapter 11 — CLI MCP (excludes GUI and inherited Go checks)", mcpCLIchecks(bin), "")
}

func mcpChecks(bin, gui, model string) []grade.Check {
	checks := mcpCLIchecks(bin)
	tunnel := grade.Ch12Checks(grade.Ch12Result{WSTunnelOK: true})[5]
	if err := mcpTunnelCheck(gui, model); err != nil {
		fail(&tunnel, "%v", err)
	}
	return append(checks, tunnel)
}

func mcpCLIchecks(bin string) []grade.Check {
	// Preserve the historical five CLI weights and IDs, but replace the old harness
	// observations: its tool/ephemera booleans followed discovery without model use.
	checks := grade.Ch12Checks(grade.Ch12Result{HandshakeOK: true, DiscoveryOK: true, ToolCallOK: true, EphemeralOK: true, ReverseOK: true})[:5]
	checks[0].Title = "MCP handshake over the actual protocol pipe"
	checks[3].Title = "Fresh round and once-per-turn automatic context"
	// A pipe carries RPC only. Never award engine/declaration credit merely because
	// a peer completed initialize/list; the actual model conversation is separate.
	mcpPipeChecks(bin, &checks[0], &checks[4])
	mcpModelChecks(bin, &checks[1], &checks[2], &checks[3])
	return checks
}

// One scanner owns stdout until EOF and then reaps the process. Cancellation
// releases a blocked reader or sender; teardown never races another cmd.Wait.
type mcpCommand struct {
	process *os.Process
	stdin   io.WriteCloser
	lines   chan string
	done    chan error
	cancel  context.CancelFunc
	stderr  bytes.Buffer
}

func startMCPCommand(bin, work string, args, env []string) (*mcpCommand, error) {
	base := append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL=fake-model", "LLM_API_KEY=fake-mcp-key", "LLM_BASE_URL=http://127.0.0.1:1/unused", "EN_DISABLE_STREAMING=1", "EN_SKILLS_DIR=", "EN_PRIMARY_SKILL=", "CH02_LOG="+filepath.Join(work, "events.jsonl"))
	return startMCPCommandEnv(bin, work, args, append(base, env...))
}

// Explicit environments let newer fixtures isolate settings without changing
// the defaults of historical callers.
func startMCPCommandEnv(bin, work string, args, env []string) (*mcpCommand, error) {
	return startMCPCommandEnvTimeout(bin, work, args, env, 15*time.Second)
}

func startMCPCommandEnvTimeout(bin, work string, args, env []string, lifetime time.Duration) (*mcpCommand, error) {
	ctx, cancel := context.WithTimeout(context.Background(), lifetime)
	p := &mcpCommand{lines: make(chan string, 128), done: make(chan error, 1), cancel: cancel}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = work
	cmd.Env = env
	cmd.Stderr = &p.stderr
	var err error
	p.stdin, err = cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		p.stdin.Close()
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		cancel()
		p.stdin.Close()
		return nil, err
	}
	p.process = cmd.Process
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 4096), 2<<20)
		for scanner.Scan() {
			select {
			case p.lines <- scanner.Text():
			case <-ctx.Done():
			}
		}
		p.done <- cmd.Wait()
		close(p.lines)
	}()
	return p, nil
}
func (p *mcpCommand) close() { p.stdin.Close(); p.cancel(); <-p.done }
func (p *mcpCommand) read() (string, error) {
	return p.readWithin(5 * time.Second)
}

func (p *mcpCommand) readWithin(budget time.Duration) (string, error) {
	select {
	case line, ok := <-p.lines:
		if !ok {
			return "", fmt.Errorf("command stdout closed: %s: %w", strings.TrimSpace(p.stderr.String()), io.EOF)
		}
		return line, nil
	case <-time.After(budget):
		return "", fmt.Errorf("command output deadline")
	}
}
func (p *mcpCommand) rpc() (map[string]any, error) {
	line, err := p.read()
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err = json.Unmarshal([]byte(line), &m); err != nil {
		return nil, fmt.Errorf("non-RPC stdout %q: %w", line, err)
	}
	return m, nil
}
func (p *mcpCommand) send(m any) error { return json.NewEncoder(p.stdin).Encode(m) }

// Keep the peer envelopes visible here: the grader and submitted client do not
// share an MCP codec that could make the same serialization mistake on both ends.
func mcpReply(id any, result any) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "result": result}
}
func mcpRequest(id int, name string, args any) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "method": "tools/call", "params": map[string]any{"name": name, "arguments": args}}
}
func mcpInitResult() map[string]any {
	return map[string]any{"protocolVersion": "2024-11-05", "serverInfo": map[string]string{"name": "course-peer", "version": "1"}, "capabilities": map[string]any{"tools": map[string]any{}}}
}
func mcpTool(name, mode string) map[string]any {
	return map[string]any{"name": name, "description": "Course external " + name, "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"text": map[string]string{"type": "string"}}}, "ephemeral": mode}
}

// This handshake is the published historical exercise version, not a claim of
// current MCP-standard compliance. IDs must be present, distinct and correlated.
func mcpHandshake(read func() (map[string]any, error), send func(any) error) error {
	init, err := read()
	if err != nil {
		return err
	}
	if init["jsonrpc"] != "2.0" || init["method"] != "initialize" || init["id"] == nil {
		return fmt.Errorf("invalid initialize: %v", init)
	}
	if err = send(mcpReply(init["id"], mcpInitResult())); err != nil {
		return err
	}
	ready, err := read()
	if err != nil {
		return err
	}
	if ready["jsonrpc"] != "2.0" || ready["method"] != "notifications/initialized" || ready["id"] != nil {
		return fmt.Errorf("missing initialized notification: %v", ready)
	}
	listed, err := read()
	if err != nil {
		return err
	}
	if listed["jsonrpc"] != "2.0" || listed["method"] != "tools/list" || listed["id"] == nil || listed["id"] == init["id"] {
		return fmt.Errorf("invalid tools/list: %v", listed)
	}
	return send(mcpReply(listed["id"], map[string]any{"tools": []any{mcpTool("auto_round", "round")}}))
}

// --mcp-pipe is an incoming service: the fixture is its external client.
func mcpServiceHandshake(read func() (map[string]any, error), send func(any) error) error {
	if err := send(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{}, "clientInfo": map[string]string{"name": "grader", "version": "1"}}}); err != nil {
		return err
	}
	m, err := read()
	if err != nil {
		return err
	}
	result, _ := m["result"].(map[string]any)
	if m["jsonrpc"] != "2.0" || m["id"] != float64(1) || m["error"] != nil || result["protocolVersion"] != "2024-11-05" {
		return fmt.Errorf("invalid correlated initialize response: %v", m)
	}
	if err := send(map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"}); err != nil {
		return err
	}
	if err := send(map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": map[string]any{}}); err != nil {
		return err
	}
	m, err = read()
	if err != nil {
		return err
	}
	result, _ = m["result"].(map[string]any)
	tools, _ := result["tools"].([]any)
	if m["jsonrpc"] != "2.0" || m["id"] != float64(2) || m["error"] != nil {
		return fmt.Errorf("invalid correlated tools/list response: %v", m)
	}
	for _, item := range tools {
		tool, _ := item.(map[string]any)
		if tool["name"] == "read_file" {
			return nil
		}
	}
	return fmt.Errorf("enabled read_file missing from service discovery")
}

func mcpPipeChecks(bin string, handshake, reverse *grade.Check) {
	work, err := os.MkdirTemp("", "edition2-mcp-pipe-")
	if err != nil {
		fail(handshake, "%v", err)
		fail(reverse, "%v", err)
		return
	}
	defer os.RemoveAll(work)
	// A pre-existing file is a safe readiness control. Its content appears only
	// in the successful result, never in the tools/call request being answered.
	const marker = "REVERSE_FILE_ONLY"
	if err = os.WriteFile(filepath.Join(work, "reverse.txt"), []byte(marker), 0600); err != nil {
		fail(reverse, "%v", err)
		return
	}
	p, err := startMCPCommand(bin, work, []string{"--mcp-pipe"}, nil)
	if err != nil {
		fail(handshake, "%v", err)
		fail(reverse, "%v", err)
		return
	}
	defer p.close()
	if err = mcpServiceHandshake(p.rpc, p.send); err != nil {
		fail(handshake, "%v", err)
		fail(reverse, "handshake prevented reverse exercise: %v", err)
		return
	}
	// tools/list has no subsequent client-ready acknowledgement. Some clients
	// install their reverse handler just after consuming its response. Poll the
	// real, side-effect-free read until it succeeds; a persistent refusal still
	// fails. Never retry malformed/miscorrelated replies or successful empty data.
	id := 900
	deadline := time.Now().Add(time.Second)
	for {
		if err = p.send(mcpRequest(id, "read_file", map[string]string{"path": "reverse.txt"})); err != nil {
			fail(reverse, "%v", err)
			return
		}
		reply, e := p.rpc()
		if e != nil {
			fail(reverse, "read_file: %v", e)
			return
		}
		text, bad, e := mcpResult(reply, id)
		if e == nil && bad && time.Now().Before(deadline) {
			id++
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if e != nil || bad || !strings.Contains(text, marker) {
			fail(reverse, "read_file: correlated=%v error=%t text=%q", e, bad, text)
		}
		break
	}
	// Retain startup observations so a readiness retry does not erase evidence
	// of the older command's registration race. Refusals below are never retried.
	reverse.Details = append(reverse.Details, fmt.Sprintf("enabled reverse read readiness attempts=%d", id-899))
	// Unknown errors alone never established working reverse execution. After
	// the enabled control, require actual errors for unknown/hidden names. A
	// forbidden external call appears on this wire and fails response identity.
	for _, name := range []string{"missing_mcp_tool", "auto_round"} {
		id++
		if err = p.send(mcpRequest(id, name, map[string]any{})); err != nil {
			fail(reverse, "%v", err)
			return
		}
		reply, e := p.rpc()
		if e != nil {
			fail(reverse, "%s: %v", name, e)
			return
		}
		text, bad, e := mcpResult(reply, id)
		if e != nil || !bad {
			fail(reverse, "%s: correlated=%v error=%t text=%q", name, e, bad, text)
		}
	}

}

// Inspect only the correlated result content. The marker deliberately does not
// occur in call arguments, so dropping returned bytes cannot accidentally pass.
func mcpResult(m map[string]any, id int) (string, bool, error) {
	if m["jsonrpc"] != "2.0" || m["id"] != float64(id) {
		return "", false, fmt.Errorf("wrong response identity: %v", m)
	}
	// The exercise permits either a protocol error or an isError tool result.
	// A string merely saying "unknown" inside a success cannot satisfy refusal.
	if e := object(m["error"]); e != nil {
		return fmt.Sprint(e["message"]), true, nil
	}
	result := object(m["result"])
	if result == nil {
		return "", false, fmt.Errorf("no result or error")
	}
	var text []string
	for _, v := range list(result["content"]) {
		c := object(v)
		if c["type"] == "text" {
			text = append(text, fmt.Sprint(c["text"]))
		}
	}
	bad, _ := result["isError"].(bool)
	if len(text) == 0 {
		return "", bad, fmt.Errorf("no textual tool result")
	}
	return strings.Join(text, "\n"), bad, nil
}

// The executable peer is intentionally small: newline JSON-RPC, discovery and
// content replies. It replaces an external tool server, never Agent/job behavior.
const mcpStdioPeer = `import json,sys
counts={}
for line in sys.stdin:
 m=json.loads(line)
 with open(sys.argv[1],'a') as trace: trace.write(json.dumps(m)+'\n')
 method=m.get('method')
 if 'id' not in m: continue
 if method=='initialize': result={'protocolVersion':'2024-11-05','capabilities':{'tools':{}}}
 elif method=='tools/list':
  result={'tools':[{'name':name,'description':'Course external '+name,'inputSchema':{'type':'object','properties':{'text':{'type':'string'}}},'ephemeral':mode} for name,mode in [('remote_echo',''),('remote_error',''),('auto_round','round'),('auto_turn','turn'),('auto_broken','round'),('not_granted','')]]}
 elif method=='tools/call':
  p=m['params'];name=p['name'];counts[name]=counts.get(name,0)+1
  if name=='remote_echo': texts=['REMOTE_RESULT_ONLY:'+str(p.get('arguments',{}).get('text','')),'SECOND_RESULT_BLOCK']
  elif name=='remote_error': texts=['REMOTE_FAILURE_ONLY']
  elif name=='auto_broken': texts=['AUTO_FAILURE_ONLY']
  else: texts=[name.upper()+'_'+str(counts[name])]
  result={'content':[{'type':'text','text':text} for text in texts],'isError':name in ['remote_error','auto_broken']}
 else: continue
 print(json.dumps({'jsonrpc':'2.0','id':m['id'],'result':result}),flush=True)
`

// Discovery may normalize the object envelope and add locally owned job
// controls. The remote fields and required semantics must still survive.
func mcpNormalizedSchema(got, original any) bool {
	decode := func(v any) map[string]any {
		b, _ := json.Marshal(v)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		return m
	}
	a, b := decode(got), decode(original)
	if a == nil || a["type"] != "object" || a["$schema"] != nil {
		return false
	}
	props, remote := object(a["properties"]), object(b["properties"])
	if props == nil {
		return false
	}
	for name, kind := range map[string]string{"ai_callback_delay": "number", "ai_callback_pattern": "string", "max_output_bytes": "integer"} {
		if _, owned := remote[name]; owned {
			continue
		}
		if control, present := props[name]; present {
			if object(control)["type"] != kind {
				return false
			}
			delete(props, name)
		}
	}
	// The fixture's remote schema has no required keys or closed-object flag.
	// Empty required and a local object policy are allowed, not extra remote
	// constraints or controls becoming required. Everything else compares intact.
	for _, m := range []map[string]any{a, b} {
		if required, ok := m["required"].([]any); ok && len(required) == 0 {
			delete(m, "required")
		}
	}
	if _, specified := b["additionalProperties"]; !specified {
		if policy, present := a["additionalProperties"]; present {
			if _, ok := policy.(bool); !ok {
				return false
			}
			delete(a, "additionalProperties")
		}
	}
	delete(b, "$schema")
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return sameJSON(x, y)
}

// A failed observation may be omitted or honestly labeled unavailable/error in
// request-local diagnostics. A bare result marker, or accumulating copies, is
// not valid automatic snapshot data. No particular application's label is used.
func mcpFailureDiagnostic(messages any) bool {
	const marker = "AUTO_FAILURE_ONLY"
	b, _ := json.Marshal(messages)
	n := strings.Count(string(b), marker)
	if n == 0 {
		return true
	}
	if n != 1 {
		return false
	}
	labeled := false
	var visit func(any)
	visit = func(value any) {
		switch v := value.(type) {
		case string:
			for _, line := range strings.Split(v, "\n") {
				if at := strings.Index(line, marker); at >= 0 {
					prefix := strings.ToLower(line[:at])
					labeled = strings.Contains(prefix, "unavailable") || strings.Contains(prefix, "error") || strings.Contains(prefix, "fail")
				}
			}
		case []any:
			for _, item := range v {
				visit(item)
			}
		case map[string]any:
			for _, item := range v {
				visit(item)
			}
		}
	}
	visit(messages)
	return labeled
}

func mcpModelChecks(bin string, discovery, call, ephemera *grade.Check) {
	badSetup := func(err error) {
		for _, c := range []*grade.Check{discovery, call, ephemera} {
			fail(c, "model session: %v", err)
		}
	}
	python, err := exec.LookPath("python3")
	if err != nil {
		badSetup(err)
		return
	}
	work, err := os.MkdirTemp("", "edition2-mcp-model-")
	if err != nil {
		badSetup(err)
		return
	}
	defer os.RemoveAll(work)
	script := filepath.Join(work, "peer.py")
	trace := filepath.Join(work, "peer.jsonl")
	if err = os.WriteFile(script, []byte(mcpStdioPeer), 0600); err != nil {
		badSetup(err)
		return
	}
	skill := filepath.Join(work, "skills", "primary")
	if err = os.MkdirAll(skill, 0700); err != nil {
		badSetup(err)
		return
	}
	args, _ := json.Marshal([]string{"-u", script, trace})
	body := "---\nname: primary\ndescription: MCP course fixture\ntype: primary\ntools: read_file remote_echo remote_error auto_round auto_turn auto_broken\nmcp_servers:\n  - name: course\n    transport: stdio\n    command: " + python + "\n    args: " + string(args) + "\n---\nExercise external tools.\n"
	if err = os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte(body), 0600); err != nil {
		badSetup(err)
		return
	}
	// Distinct call identities bind each observed result to its actual request.
	// The hidden guessed call is deliberate: declaration absence alone cannot
	// establish that ordinary dispatch also rejects automatic-only tools.
	provider := fakevendor.New([]fakevendor.Reply{
		skillCall("remote-echo", "remote_echo", `{"text":"ARGUMENT_ONLY"}`),
		skillCall("remote-error", "remote_error", `{}`),
		skillCall("hidden-auto", "auto_round", `{}`),
		{Text: "MCP_FIRST_TURN_DONE"}, {Text: "MCP_SECOND_TURN_DONE"},
	})
	defer provider.Close()
	p, err := startMCPCommand(bin, work, []string{"chat"}, []string{"LLM_BASE_URL=" + provider.URL(), "EN_SKILLS_DIR=" + filepath.Dir(skill), "EN_PRIMARY_SKILL=primary"})
	if err != nil {
		badSetup(err)
		return
	}
	defer p.close()
	// Final answer text is an existing human interface observation. Waiting for it
	// before the next input prevents the next prompt being correctly taken as a hint.
	for _, marker := range []string{"MCP_FIRST_TURN_DONE", "MCP_SECOND_TURN_DONE"} {
		if _, err = fmt.Fprintln(p.stdin, "Exercise MCP now."); err != nil {
			badSetup(err)
			return
		}
		for {
			line, e := p.read()
			if e != nil {
				badSetup(e)
				return
			}
			if strings.Contains(line, marker) {
				break
			}
		}
	}
	requests := provider.Requests()
	if len(requests) != 5 {
		badSetup(fmt.Errorf("want five actual requests, got %d", len(requests)))
		return
	}
	// Check real provider declarations on every round: a registry containing the
	// right schema would not prove the renderer actually sent it to the model.
	for _, r := range requests {
		skillDeclarations(discovery, r, []string{"read_file", "remote_echo", "remote_error", "load_skill", "unload_skill"})
		for _, v := range list(r.JSON()["tools"]) {
			tool := object(v)
			if tool["name"] == "remote_echo" {
				expected := mcpTool("remote_echo", "")
				if tool["description"] != expected["description"] || !mcpNormalizedSchema(tool["input_schema"], expected["inputSchema"]) {
					fail(discovery, "discovered description/schema changed")
				}
			}
		}
	}
	// The second content block detects lossy single-block bridge implementations;
	// unique returned text prevents earlier call arguments from satisfying this.
	got := skillResult(call, requests[1], "remote-echo", false)
	if !strings.Contains(got, "REMOTE_RESULT_ONLY:ARGUMENT_ONLY") || !strings.Contains(got, "SECOND_RESULT_BLOCK") {
		fail(call, "actual remote result/arguments or content blocks lost: %q", got)
	}
	if !strings.Contains(skillResult(call, requests[2], "remote-error", true), "REMOTE_FAILURE_ONLY") {
		fail(call, "remote isError/content did not reach model as a failed result")
	}
	skillResult(ephemera, requests[3], "hidden-auto", true)
	// Compare only model message content, not schemas or recorded history events.
	// A fresh counter proves execution while old-marker absence proves consumption.
	for i, r := range requests {
		messages, _ := json.Marshal(r.JSON()["messages"])
		text := string(messages)
		for n := 1; n <= 5; n++ {
			want := 0
			if n == i+1 {
				want = 1
			}
			if strings.Count(text, fmt.Sprintf("AUTO_ROUND_%d", n)) != want {
				fail(ephemera, "request%d lacks only its fresh round%d context", i+1, i+1)
				break
			}
		}
		// Computed once at turn start, but reusable on each request of that turn.
		turn := 1
		if i == 4 {
			turn = 2
		}
		for n := 1; n <= 2; n++ {
			want := 0
			if n == turn {
				want = 1
			}
			if strings.Count(text, fmt.Sprintf("AUTO_TURN_%d", n)) != want {
				fail(ephemera, "request%d has wrong once-per-turn context", i+1)
			}
		}
		// Failed automatic output is diagnostic, not useful context. Reaching all
		// five requests also proves this failure did not abort either turn.
		if !mcpFailureDiagnostic(r.JSON()["messages"]) {
			fail(ephemera, "request%d presented failed automatic output as data or accumulated diagnostics", i+1)
		}
	}
	// Read what the independent process actually received. Reply scripts alone do
	// not prove invocation cadence, hidden dispatch refusal, or argument forwarding.
	data, e := os.ReadFile(trace)
	if e != nil {
		fail(discovery, "stdio peer never started: %v", e)
		fail(ephemera, "no external automatic invocations")
		return
	}
	counts := map[string]int{}
	initialized := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) != nil {
			fail(call, "invalid peer trace")
			continue
		}
		if m["method"] == "notifications/initialized" {
			initialized++
		}
		if m["method"] == "tools/call" {
			params := object(m["params"])
			name, _ := params["name"].(string)
			counts[name]++
			if name == "remote_echo" && object(params["arguments"])["text"] != "ARGUMENT_ONLY" {
				fail(call, "outgoing MCP arguments changed")
			}
		}
	}
	if initialized != 1 {
		fail(discovery, "stdio initialized notifications=%d", initialized)
	}
	for name, want := range map[string]int{"auto_round": 5, "auto_turn": 2, "auto_broken": 5} {
		if counts[name] != want {
			fail(ephemera, "external %s calls=%d want%d", name, counts[name], want)
		}
	}
	if counts["remote_echo"] != 1 || counts["remote_error"] != 1 {
		fail(call, "external normal invocation counts: %v", counts)
	}
}

func mcpTunnelCheck(gui, model string) error {
	// Build the independently supplied GUI module through the existing driver.
	// Looking for a particular source filename proved neither startup nor routing.
	bin, cleanup, err := grade.Build(gui)
	if err != nil {
		return err
	}
	defer cleanup()
	work, err := os.MkdirTemp("", "edition2-mcp-ws-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	if err = os.WriteFile(filepath.Join(work, "reverse.txt"), []byte("WS_REVERSE_FILE_ONLY"), 0600); err != nil {
		return err
	}
	static, err := filepath.Abs(filepath.Join(gui, "web", "gui"))
	if err != nil {
		return err
	}
	skillsDir := filepath.Join(work, "skills")
	// Start HTTP with a non-WebSocket primary. The declared browser skill is
	// activated only after a browser peer can answer discovery on this socket.
	for name, body := range map[string]string{
		"primary": "---\nname: primary\ndescription: MCP tunnel fixture\ntype: primary\ntools: read_file\nloadable-skills: browser\n---\nExercise the declared browser tunnel.",
		"browser": "---\nname: browser\ndescription: MCP tunnel peer\ntype: loadable\ntools: auto_round\nmcp_servers:\n  - name: browser-debug\n    transport: websocket\n---\nExercise only this connected peer.",
	} {
		dir := filepath.Join(skillsDir, name)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0600); err != nil {
			return err
		}
	}
	url, stop, err := startGUI(bin, static, model, "http://127.0.0.1:1/unused", work, "EN_SKILLS_DIR="+skillsDir, "EN_PRIMARY_SKILL=primary")
	if err != nil {
		return err
	}
	defer stop()
	p, err := openGUIPeer(url)
	if err != nil {
		return err
	}
	defer p.close()
	if err = p.send("mcp_open", ""); err != nil {
		return err
	}
	// A browser peer supplies only JSON-RPC. No DOM implementation, GUI source
	// filename or public API name stands in for observing both tunnel directions.
	read := func() (map[string]any, error) {
		ms, e := p.until(3*time.Second, messageKind("jsonrpc"))
		if e != nil {
			return nil, e
		}
		return object(ms[len(ms)-1]["payload"]), nil
	}
	send := func(v any) error { return p.conn.WriteJSON(map[string]any{"type": "jsonrpc", "payload": v}) }
	if err = mcpHandshake(read, send); err != nil {
		return fmt.Errorf("tunnel handshake: %w", err)
	}
	if err = p.conn.WriteJSON(map[string]any{"type": "load_skill", "name": "browser"}); err != nil {
		return err
	}
	if _, err = p.until(3*time.Second, func(m map[string]any) bool {
		if m["type"] != "skill_state" {
			return false
		}
		for _, item := range list(m["skills"]) {
			skill := object(item)
			if skill["name"] == "browser" && skill["enabled"] == true && skill["pending_unload"] != true {
				return true
			}
		}
		return false
	}); err != nil {
		return fmt.Errorf("activate declared browser skill: %w", err)
	}
	// One successful file result requires a complete outbound request, inbound
	// peer request and outbound correlated reply; handshake alone misses that path.
	if err = send(mcpRequest(990, "read_file", map[string]string{"path": "reverse.txt"})); err != nil {
		return err
	}
	reply, err := read()
	if err != nil {
		return err
	}
	text, bad, err := mcpResult(reply, 990)
	if err != nil || bad || !strings.Contains(text, "WS_REVERSE_FILE_ONLY") {
		return fmt.Errorf("tunnel reverse result: %v error=%t text=%q", err, bad, text)
	}
	return nil
}
