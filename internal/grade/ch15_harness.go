package grade

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Chapter 15 grades context management from the vendor's side of the
// wire. Every check reads the requests the fake vendor records, and a
// few also read or plant events in ch11's save.json. None of them read
// the student's own log layout, which rule 9 leaves to the student.

// Models the grader launches under. Rule 6 is gated on a features row,
// so the grader needs one model that stubs per round trip and one that
// does not. Both rows are part of the student's table; the TL;DR names
// them.
const (
	ch15StubModel   = "claude-opus-5-course"   // stubs per round trip, inline tools
	ch15NoStubModel = "claude-sonnet-5-course" // neither
)

// ch15SmallTarget is the smallest target the settings accept. At this
// size the ladder must fire within a dozen file reads, which keeps the
// session short enough to grade in seconds.
const ch15SmallTarget = 20000

// Markers. Each is unique enough that a substring count in a request
// body means exactly what the check says it means.
const (
	ch15ManualMark  = "MANUAL-BODY-7f3a91"
	ch15HandoffText = "HANDOFF-NOTE-c41e: reads done, ladder fired, next step is the summary."
	ch15MCPTool     = "gui_click"
)

// ch15FileMark is the marker repeated through planted file n.
func ch15FileMark(n int) string { return fmt.Sprintf("FILE%02d-CONTENT-", n) }

// ch15WriteFile plants a workspace file of about size bytes whose every
// line carries its marker, so any surviving fragment of the result is
// still recognisable.
func ch15WriteFile(dir string, n, size int) error {
	var b strings.Builder
	for b.Len() < size {
		fmt.Fprintf(&b, "%sline %04d of planted file %02d\n", ch15FileMark(n), b.Len(), n)
	}
	return os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%02d.txt", n)), []byte(b.String()), 0o644)
}

// ch15CreateSkills writes the fixture skills. "base" is the primary
// skill; "manual" is loadable and adds the think tool; "mcp-tools" is
// loadable and connects ch13's fake MCP server.
func ch15CreateSkills(dir, gradeBin string) error {
	skills := map[string]string{
		"base": `---
name: base
description: Base agent skill for the context-management checks
type: primary
tools: read_file keep_tool_results micro_handoff
loadable-skills: manual mcp-tools
---
You are a helpful agent.
`,
		"manual": `---
name: manual
description: A manual that must outlive tool clearing
type: loadable
tools: think
---
` + ch15ManualMark + `
Always read the manual before editing. This body is the whole manual.
`,
	}
	mcpSkill := fmt.Sprintf(`---
name: mcp-tools
description: Connects a fake MCP server mid-session
type: loadable
mcp_servers:
  - name: fakemcp
    transport: stdio
    command: %s
    args: ["--fake-mcp"]
---
Tools from the fake MCP server.
`, gradeBin)
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

// ch15Opts describes one launch.
type ch15Opts struct {
	dir     string
	model   string
	prompts []string
	// expect[i] is how many requests the vendor must have recorded once
	// prompt i's turn is over. The harness waits for that count instead
	// of sleeping, so a slow machine is not a failing student.
	expect  []int
	replies []fakevendor.Reply
	// kill ends the run with SIGKILL instead of closing stdin, so the
	// agent gets no chance to snapshot.
	kill bool
}

type ch15Out struct {
	reqs  []fakevendor.Recorded
	fatal string
}

func ch15Launch(bin, skillsDir, guiDir string, o ch15Opts) ch15Out {
	srv := fakevendor.New(o.replies)
	defer srv.Close()

	port := freePort()
	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = o.dir
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_MODEL="+o.model,
		"LLM_VENDOR=anthropic",
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"CH02_LOG="+filepath.Join(o.dir, "events.jsonl"),
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ch15Out{fatal: fmt.Sprintf("stdin pipe: %v", err)}
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return ch15Out{fatal: fmt.Sprintf("start: %v", err)}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	stop := func() {
		cmd.Process.Kill()
		<-done
	}

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		stop()
		return ch15Out{fatal: "GUI server did not start"}
	}

	for i, p := range o.prompts {
		line, _ := json.Marshal(map[string]string{"kind": "prompt", "text": p})
		fmt.Fprintln(stdin, string(line))
		want := i + 1
		if i < len(o.expect) {
			want = o.expect[i]
		}
		if !ch15WaitRequests(srv, want, 20*time.Second) {
			stop()
			return ch15Out{reqs: srv.Requests(), fatal: fmt.Sprintf(
				"after prompt %d the vendor saw %d requests, expected %d", i+1, len(srv.Requests()), want)}
		}
		// The last request's reply still has to land and end the turn.
		time.Sleep(400 * time.Millisecond)
	}

	if o.kill {
		stop()
		stdin.Close()
		return ch15Out{reqs: srv.Requests()}
	}
	stdin.Close()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		<-done
		return ch15Out{reqs: srv.Requests(), fatal: "agent did not save and exit within 10s of stdin closing"}
	}
	return ch15Out{reqs: srv.Requests()}
}

func ch15WaitRequests(srv *fakevendor.Server, n int, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if len(srv.Requests()) >= n {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return len(srv.Requests()) >= n
}

// ----------------------------------------------------------------
// Reading an Anthropic request body
// ----------------------------------------------------------------

// ch15Req is the part of an Anthropic Messages request the checks read.
// Anything else in the body is ignored.
type ch15Req struct {
	System   json.RawMessage `json:"system"`
	Tools    json.RawMessage `json:"tools"`
	Messages []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"messages"`
}

// ch15Block is one content block. Content may be a string or a list of
// blocks; a string is treated as one text block.
type ch15Block struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	ToolUseID string          `json:"tool_use_id"`
	Raw       json.RawMessage `json:"-"`
}

func ch15Parse(body []byte) (ch15Req, error) {
	var r ch15Req
	err := json.Unmarshal(body, &r)
	return r, err
}

// ch15Blocks flattens every message's content into blocks, keeping each
// block's raw bytes so a check can ask what a tool result contains.
func ch15Blocks(r ch15Req) []ch15Block {
	var out []ch15Block
	for _, m := range r.Messages {
		var s string
		if json.Unmarshal(m.Content, &s) == nil {
			raw, _ := json.Marshal(map[string]string{"type": "text", "text": s})
			out = append(out, ch15Block{Type: "text", Raw: raw})
			continue
		}
		var list []json.RawMessage
		if json.Unmarshal(m.Content, &list) != nil {
			continue
		}
		for _, raw := range list {
			var b ch15Block
			json.Unmarshal(raw, &b)
			b.Raw = raw
			out = append(out, b)
		}
	}
	return out
}

// ch15MessagesText is the messages array alone, as bytes, so a marker
// count excludes the system prompt and the tools array.
func ch15MessagesText(r ch15Req) string {
	b, _ := json.Marshal(r.Messages)
	return string(b)
}

// ch15Pairing reports every tool_use without a tool_result and every
// tool_result without a tool_use. Empty means well paired.
func ch15Pairing(blocks []ch15Block) []string {
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

// ch15ResultFor returns the raw tool_result block for a call id.
func ch15ResultFor(blocks []ch15Block, id string) (string, bool) {
	for _, b := range blocks {
		if b.Type == "tool_result" && b.ToolUseID == id {
			return string(b.Raw), true
		}
	}
	return "", false
}

func ch15HasUse(blocks []ch15Block, id string) bool {
	for _, b := range blocks {
		if b.Type == "tool_use" && b.ID == id {
			return true
		}
	}
	return false
}

// ch15Replies builders.
func ch15Text(s string) fakevendor.Reply {
	return fakevendor.Reply{Text: s, Usage: fakevendor.Canonical{Input: 100, Output: 30}}
}

func ch15Call(id, name, args string) fakevendor.Reply {
	return fakevendor.Reply{ToolID: id, ToolName: name, ToolArgs: args,
		Usage: fakevendor.Canonical{Input: 100, Output: 30}}
}

func ch15Read(id string, n int) fakevendor.Reply {
	return ch15Call(id, "read_file", fmt.Sprintf(`{"path":"f%02d.txt"}`, n))
}

// ch15CopyDir copies a run directory's regular files (save.json, its
// backup, settings, the workspace files, and whatever the student's
// log layout put there) so a later launch starts from the same state.
func ch15CopyDir(src, dst string) error {
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

func ch15WriteSettings(dir string, target int) error {
	b, _ := json.Marshal(map[string]int{"context_target": target})
	return os.WriteFile(filepath.Join(dir, "settings.json"), b, 0o644)
}
