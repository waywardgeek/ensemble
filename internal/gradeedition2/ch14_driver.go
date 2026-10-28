package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Edition 2 Chapter 14 retains original Chapter 15 checks of context management from the vendor's side of the
// wire. Every check reads the requests the fake vendor records, and a
// few also read or plant events in ch11's save.json. None of them read
// the student's own log layout, which rule 9 leaves to the student.

// Models the grader launches under. Rule 6 is gated on a features row,
// so the grader needs one model that stubs per round trip and one that
// does not. Both rows are part of the student's table; the TL;DR names
// them.
const (
	contextStubModel   = "claude-opus-5-course"   // stubs per round trip, inline tools
	contextNoStubModel = "claude-sonnet-5-course" // neither
)

// contextSmallTarget is the smallest target the settings accept. At this
// size the ladder must fire within a dozen file reads, which keeps the
// session short enough to grade in seconds.
const contextSmallTarget = 20000

// Markers. Each is unique enough that a substring count in a request
// body means exactly what the check says it means.
const (
	contextPrimaryMark = "PRIMARY-BODY-context-14"
	contextManualMark  = "MANUAL-BODY-7f3a91"
	contextHandoffText = "HANDOFF-NOTE-c41e: reads done, ladder fired, next step is the summary."
)

// contextFileMark is the marker repeated through planted file n.
func contextFileMark(n int) string { return fmt.Sprintf("FILE%02d-CONTENT-", n) }

// contextWriteFile plants a workspace file of about size bytes whose every
// line carries its marker, so any surviving fragment of the result is
// still recognisable.
func contextWriteFile(dir string, n, size int) error {
	var b strings.Builder
	for b.Len() < size {
		fmt.Fprintf(&b, "%sline %04d of planted file %02d\n", contextFileMark(n), b.Len(), n)
	}
	full := filepath.Join(dir, contextPath(n))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(b.String()), 0o644)
}

// contextPath is planted file n's workspace-relative path. It is long on
// purpose: the calls band measures call bytes, and a dozen short paths
// never fill it, so without long paths the RedactTool watermark would
// never be reached and rule 7's second band would go ungraded.
func contextPath(n int) string {
	return filepath.Join("notes-"+strings.Repeat("d", 190), fmt.Sprintf("f%02d-%s.txt", n, strings.Repeat("x", 180)))
}

// contextCreateSkills writes the fixture skills. "base" is the primary
// skill; "manual" is loadable and adds the list_directory tool; "mcp-tools" is
// loadable and connects the same stateful external peer used by Chapter 12.
func contextCreateSkills(dir string) error {
	python, err := exec.LookPath("python3")
	if err != nil {
		return err
	}
	script := filepath.Join(filepath.Dir(dir), "browser.py")
	if err := os.WriteFile(script, []byte(guiDebugPeer), 0600); err != nil {
		return err
	}
	args, _ := json.Marshal([]string{"-u", script})
	skills := map[string]string{
		"base": `---
name: base
description: Base agent skill for the context-management checks
type: primary
tools: read_file load_skill unload_skill keep_tool_results micro_handoff
loadable-skills: manual mcp-tools
---
` + contextPrimaryMark + `
You are a helpful agent.
`,
		"manual": `---
name: manual
description: A manual that must outlive tool clearing
type: loadable
tools: list_directory
---
` + contextManualMark + `
Always read the manual before editing. This body is the whole manual.
`,
	}
	mcpSkill := fmt.Sprintf(`---
name: mcp-tools
description: Connects a fake MCP server mid-session
tools: gui_snapshot tts_queue gui_click gui_input
type: loadable
mcp_servers:
  - name: fakemcp
    transport: stdio
    command: %s
    args: %s
---
Tools from the fake MCP server.
`, python, args)
	skills["mcp-tools"] = mcpSkill
	for name, body := range skills {
		d := filepath.Join(dir, name)
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// contextOpts describes one launch.
type contextOpts struct {
	// startupManual makes C# runtime configuration explicit across restarts.
	startupManual bool
	dir           string
	target        int
	model         string
	prompts       []string
	// expect[i] is how many requests the vendor must have recorded once
	// prompt i's turn is over. Count is checked after the observed turn end;
	// provider arrival alone does not establish completion.
	expect  []int
	replies []fakevendor.Reply
	// kill ends the run with SIGKILL instead of closing stdin, so the
	// agent gets no chance to snapshot.
	kill bool
}

type contextOut struct {
	reqs  []fakevendor.Recorded
	fatal string
	// meters are fresh public usage snapshots after a memory attempt drains.
	meters []map[string]any
}

// Settings and turn completion are observed on the public socket. A request
// arriving at the vendor is too early to close stdin or crash the process: its
// final response and journal append must have completed first.
func contextLaunch(bin, skillsDir, guiDir string, o contextOpts) (out contextOut) {
	srv := fakevendor.New(o.replies)
	defer srv.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return contextOut{fatal: err.Error()}
	}
	port := fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
	listener.Close()
	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = o.dir
	cmd.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL="+o.model, "LLM_RESOLVED_MODEL=", "LLM_API_KEY=fake-key", "LLM_BASE_URL="+srv.URL(), "EN_DISABLE_STREAMING=1", "EN_SKILLS_DIR="+skillsDir, "EN_PRIMARY_SKILL=base", "CH02_LOG="+filepath.Join(o.dir, "events.jsonl"))
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = io.Discard, &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return contextOut{fatal: err.Error()}
	}
	if err = cmd.Start(); err != nil {
		stdin.Close()
		return contextOut{fatal: err.Error()}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	finished := false
	defer func() {
		stdin.Close()
		if !finished {
			cmd.Process.Kill()
			<-done
		}
		out.reqs = srv.Requests()
		if out.fatal != "" {
			out.fatal += ": " + stderr.String()
		}
		// An unload defect must not strand this grader's external peer.
		if pid := guiDebugPID(o.dir); pid > 0 {
			if p, e := os.FindProcess(pid); e == nil {
				p.Kill()
			}
		}
	}()
	var peer *guiPeer
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		peer, err = openGUIPeer("ws://localhost:" + port + "/ws")
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
		return contextOut{fatal: "GUI socket did not start: " + err.Error()}
	}
	defer peer.close()
	if err = peer.send("subscribe", ""); err != nil {
		return contextOut{fatal: err.Error()}
	}
	if _, err = readSettings(peer, "current_settings"); err != nil {
		return contextOut{fatal: err.Error()}
	}
	target := o.target
	if target == 0 {
		target = 400000
	}
	if err = peer.conn.WriteJSON(map[string]any{"type": "update_settings", "settings": map[string]int{"context_target": target}}); err != nil {
		return contextOut{fatal: err.Error()}
	}
	settings, err := readSettings(peer, "settings_changed")
	if err != nil || settings["context_target"] != float64(target) {
		return contextOut{fatal: fmt.Sprintf("context_target %d not applied in full snapshot: got %v (%v)", target, settings["context_target"], err)}
	}
	for i, prompt := range o.prompts {
		if err = json.NewEncoder(stdin).Encode(map[string]string{"kind": "prompt", "text": prompt}); err != nil {
			return contextOut{fatal: err.Error()}
		}
		messages, e := peer.until(20*time.Second, messageKind("turn_ended"))
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
			return contextOut{fatal: fmt.Sprintf("completed prompt %d made %d requests, want %d", i+1, n, want)}
		}
	}
	if o.kill {
		if err = cmd.Process.Kill(); err != nil {
			return contextOut{fatal: "SIGKILL: " + err.Error()}
		}
		<-done
		finished = true
		return out
	}
	stdin.Close()
	select {
	case err = <-done:
		finished = true
		if err != nil {
			out.fatal = "nonzero GUI exit after EOF: " + err.Error()
		}
	case <-time.After(10 * time.Second):
		out.fatal = "GUI did not save and exit after EOF"
	}
	return out
}

// ----------------------------------------------------------------
// Reading an Anthropic request body
// ----------------------------------------------------------------

// contextReq is the part of an Anthropic Messages request the checks read.
// Anything else in the body is ignored.
type contextReq struct {
	// System retains the exact startup instruction bytes for prefix comparison.
	System json.RawMessage `json:"system"`
	// Tools retains the exact startup declaration bytes for prefix comparison.
	Tools json.RawMessage `json:"tools"`
	// Messages preserves provider order, including inline system deltas.
	Messages []struct {
		// Role identifies the provider message boundary for placement checks.
		Role string `json:"role"`
		// Content accepts the provider string or ordered block representation.
		Content json.RawMessage `json:"content"`
	} `json:"messages"`
}

// contextBlock is one content block. Content may be a string or a list of
// blocks; a string is treated as one text block.
type contextBlock struct {
	// Type selects the public wire or event variant under inspection.
	Type string `json:"type"`
	// ID identifies this provider call.
	ID string `json:"id"`
	// Name identifies the declared public tool.
	Name string `json:"name"`
	// ToolUseID correlates this result to its provider call.
	ToolUseID string `json:"tool_use_id"`
	// Raw retains the complete block for result-content assertions.
	Raw json.RawMessage `json:"-"`
}

func contextParse(body []byte) (contextReq, error) {
	var r contextReq
	err := json.Unmarshal(body, &r)
	return r, err
}

// contextBlocks flattens every message's content into blocks, keeping each
// block's raw bytes so a check can ask what a tool result contains.
func contextBlocks(r contextReq) []contextBlock {
	var out []contextBlock
	for _, m := range r.Messages {
		var s string
		if json.Unmarshal(m.Content, &s) == nil {
			raw, _ := json.Marshal(map[string]string{"type": "text", "text": s})
			out = append(out, contextBlock{Type: "text", Raw: raw})
			continue
		}
		var list []json.RawMessage
		if json.Unmarshal(m.Content, &list) != nil {
			continue
		}
		for _, raw := range list {
			var b contextBlock
			json.Unmarshal(raw, &b)
			b.Raw = raw
			out = append(out, b)
		}
	}
	return out
}

// contextMessagesText is the messages array alone, as bytes, so a marker
// count excludes the system prompt and the tools array.
func contextMessagesText(r contextReq) string {
	b, _ := json.Marshal(r.Messages)
	return string(b)
}

// contextPairing reports every tool_use without a tool_result and every
// tool_result without a tool_use. Empty means well paired.
func contextPairing(blocks []contextBlock) []string {
	uses := map[string]bool{}
	results := map[string]bool{}
	for _, b := range blocks {
		switch b.Type {
		case "tool_use":
			uses[b.ID] = true
		case "tool_result":
			results[b.ToolUseID] = true
		}
	}
	var bad []string
	for id := range uses {
		if !results[id] {
			bad = append(bad, "call "+id+" has no result")
		}
	}
	for id := range results {
		if !uses[id] {
			bad = append(bad, "result "+id+" has no call")
		}
	}
	return bad
}

// contextResultFor returns the raw tool_result block for a call id.
func contextResultFor(blocks []contextBlock, id string) (string, bool) {
	for _, b := range blocks {
		if b.Type == "tool_result" && b.ToolUseID == id {
			return string(b.Raw), true
		}
	}
	return "", false
}

func contextHasUse(blocks []contextBlock, id string) bool {
	for _, b := range blocks {
		if b.Type == "tool_use" && b.ID == id {
			return true
		}
	}
	return false
}

// contextReplies builders.
func contextText(s string) fakevendor.Reply {
	return fakevendor.Reply{Text: s, Usage: fakevendor.Canonical{Input: 100, Output: 30}}
}

func contextCall(id, name, args string) fakevendor.Reply {
	return fakevendor.Reply{ToolID: id, ToolName: name, ToolArgs: args,
		Usage: fakevendor.Canonical{Input: 100, Output: 30}}
}

func contextRead(id string, n int) fakevendor.Reply {
	args, _ := json.Marshal(map[string]string{"path": contextPath(n)})
	return contextCall(id, "read_file", string(args))
}

// contextCopyDir copies a run directory's regular files (save.json, its
// backup, settings, the workspace files, and whatever the student's
// log layout put there) so a later launch starts from the same state.
func contextCopyDir(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}

// Null only the disposable projection: replay still has the complete original log.
func contextNullSnapshot(data []byte) ([]byte, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, err
	}
	top["context"] = json.RawMessage("null")
	return json.Marshal(top)
}

// The same-batch read must actually succeed even though its result is immediately
// cleared from the next request. The durable public event log supplies that proof.
func contextHandoffLog(dir string) string { return contextHandoffLogFor(dir, false) }

func contextHandoffLogFor(dir string, executable bool) string {
	data, err := os.ReadFile(filepath.Join(dir, "save.json"))
	if err != nil {
		return err.Error()
	}
	var save struct {
		// Log preserves the durable public event sequence.
		Log []struct {
			// Seq supplies the recorded order independently of timestamps.
			Seq uint64 `json:"seq"`
			// Type selects the public wire or event variant under inspection.
			Type string `json:"type"`
			// Tool carries the public call or result payload.
			Tool struct {
				// CallID correlates dispatch and completion with the model call.
				CallID string `json:"call_id"`
				// IsError distinguishes a successful operation from an ordinary refusal.
				IsError bool `json:"is_error"`
				// Parts carries actual returned content, not the earlier arguments.
				Parts json.RawMessage `json:"parts"`
			} `json:"tool"`
			// Parts carries the C# marked note; Handoff is the Go payload.
			Parts []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"parts"`
			Handoff struct {
				// Text is the recorded note that must survive tool clearing.
				Text string `json:"text"`
			} `json:"handoff"`
		} `json:"log"`
	}
	if err := json.Unmarshal(data, &save); err != nil {
		return err.Error()
	}
	var call, result, note, sibling uint64
	for _, event := range save.Log {
		if event.Tool.CallID == "H2" {
			if event.Type == "tool_called" {
				call = event.Seq
			}
			if event.Type == "tool_returned" && !event.Tool.IsError {
				result = event.Seq
			}
		}
		if !executable && event.Type == "micro_handoff" && event.Handoff.Text == contextHandoffText {
			note = event.Seq
		}
		if executable && event.Type == "handoff_recorded" && len(event.Parts) == 1 &&
			event.Parts[0].Type == "handoff" && event.Parts[0].Text == contextHandoffText {
			note = event.Seq
		}
		if event.Type == "tool_returned" && event.Tool.CallID == "H3" && !event.Tool.IsError && strings.Contains(string(event.Tool.Parts), contextFileMark(contextLadderFiles)) {
			sibling = event.Seq
		}
	}
	if call == 0 || result <= call || note <= result || sibling == 0 || sibling >= note {
		return fmt.Sprintf("handoff log requires call < successful result < separate note and successful same-batch read: call=%d result=%d note=%d sibling=%d", call, result, note, sibling)
	}
	return ""
}

// Inspect actual inline block kinds, never a tool name echoed in a manual or
// call argument. System deltas must follow the entire user turn (ephemera too)
// and precede its assistant continuation, or end the request.
func contextInline(req contextReq) (map[string]bool, map[string]bool, error) {
	added, removed := map[string]bool{}, map[string]bool{}
	for i, m := range req.Messages {
		if m.Role != "system" {
			continue
		}
		// Consecutive system messages form one legal declaration group; the
		// contract constrains its outer neighbors, not renderer coalescing.
		first := i == 0 || req.Messages[i-1].Role != "system"
		last := i+1 == len(req.Messages) || req.Messages[i+1].Role != "system"
		if (first && (i == 0 || req.Messages[i-1].Role != "user")) || (last && i+1 < len(req.Messages) && req.Messages[i+1].Role != "assistant") {
			return nil, nil, fmt.Errorf("inline system message %d is not between complete user and assistant/end", i)
		}
		var blocks []struct {
			// Type selects the public wire or event variant under inspection.
			Type string `json:"type"`
			// Tool carries the public call or result payload.
			Tool struct {
				// Type selects the public wire or event variant under inspection.
				Type string `json:"type"`
				// Name identifies the declared public tool.
				Name string `json:"name"`
				// Definition supplies the real inline tool declaration.
				Definition struct {
					// Name identifies the declared public tool.
					Name string `json:"name"`
					// Schema supplies the declaration argument schema.
					Schema json.RawMessage `json:"input_schema"`
				} `json:"definition"`
			} `json:"tool"`
		}
		if err := json.Unmarshal(m.Content, &blocks); err != nil {
			return nil, nil, err
		}
		for _, b := range blocks {
			switch b.Type {
			case "tool_addition":
				if b.Tool.Type != "tool_definition" || b.Tool.Definition.Name == "" || len(b.Tool.Definition.Schema) == 0 {
					return nil, nil, fmt.Errorf("malformed tool addition")
				}
				added[b.Tool.Definition.Name] = true
			case "tool_removal":
				if b.Tool.Type != "tool_reference" || b.Tool.Name == "" {
					return nil, nil, fmt.Errorf("malformed tool removal")
				}
				removed[b.Tool.Name] = true
			}
		}
	}
	return added, removed, nil
}
