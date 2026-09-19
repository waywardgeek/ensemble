package grade

// Chapter 10 harness — Skills engine.
//
// Tests skill loading, dependency resolution, progressive disclosure,
// variable substitution, and blocked-skill rejection.  Drives the binary
// through the same WebSocket+fakevendor path as ch8/ch9 and inspects the
// fakevendor's recorded requests for tool declarations.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// ----------------------------------------------------------------
// Result type
// ----------------------------------------------------------------

// Ch10Result holds evidence for the ch10 checks.
type Ch10Result struct {
	Base string

	BuildOK  bool
	BuildErr string

	// Initial tool declarations from the first request.
	InitialTools    []string
	InitialToolsErr string

	// System prompt from the first request.
	SystemPrompt    string
	SystemPromptErr string

	// load_skill("code-tools") — verify new tools appear.
	LoadSkillOK       bool
	LoadSkillErr      string
	PostLoadTools     []string // tools after loading code-tools
	LoadSkillResult   string   // tool result content

	// Progressive disclosure — loading code-tools makes search-tools loadable.
	// Then loading search-tools auto-loads search-helpers (depends chain).
	DisclosureOK       bool
	DisclosureErr      string
	PostDisclosureTools []string
	DisclosureResult   string

	// depends-autoload — search-helpers' tools appear after loading search-tools.
	DependsOK  bool
	DependsErr string

	// Variable substitution — search-tools body contains $CUSTOM_VAR.
	VarSubOK   bool
	VarSubErr  string
	VarSubBody string

	// Blocked skill — "blocked" is not loadable.
	BlockedOK  bool
	BlockedErr string
	BlockedResult string

	// System prompt check.
	SystemPromptOK  bool

	// Ensemble primary skill test.
	EnsembleTools    []string
	EnsemblePrompt   string
	EnsembleToolsErr string

	// Ch9 parity.
	Ch9Result *Ch9Result
	Ch9Err    string

	HelpersErr string
}

// ----------------------------------------------------------------
// Fixture skills
// ----------------------------------------------------------------

func ch10CreateFixtureSkills(dir string) error {
	skills := map[string]string{
		"base": `---
name: base
description: Base agent skill with core tools
type: primary
tools: read_file think
loadable-skills: code-tools
---
You are a helpful agent.

Available skills to load: $SKILLS
`,
		"code-tools": `---
name: code-tools
description: File editing tools for code modification
type: loadable
tools: edit_file write_file
loadable-skills: search-tools
---
You now have code editing capabilities.
`,
		"search-tools": `---
name: search-tools
description: Code search and discovery tools
type: loadable
tools: search_files
depends: search-helpers
---
Search tools loaded. Custom value: $CUSTOM_VAR
`,
		"search-helpers": `---
name: search-helpers
description: Internal helpers for search indexing
type: dependency
tools: list_directory
---
Search helper internals loaded.
`,
		"blocked": `---
name: blocked
description: A skill that is never made loadable
type: loadable
tools: blocked_tool
---
This skill should never be loadable.
`,
		"ensemble": `---
name: ensemble
description: Full Ensemble agent with all core tools
type: primary
tools: run_command wait_for_job send_input kill_job read_file write_file edit_file list_directory search_files think load_skill unload_skill
loadable-skills: code-tools search-tools
---
You are the Ensemble agent. You have full access to all tools.

Available skills to load: $SKILLS
`,
	}

	for name, content := range skills {
		skillDir := filepath.Join(dir, name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------
// Fake vendor replies
// ----------------------------------------------------------------

// ch10InitialReplies: the LLM says hello, then calls load_skill("code-tools").
func ch10InitialReplies() []fakevendor.Reply {
	return []fakevendor.Reply{
		{
			// Round 1: call load_skill to load code-tools.
			ToolName: "load_skill",
			ToolArgs: `{"name":"code-tools"}`,
			ToolID:   "call_load_1",
			Usage:    fakevendor.Canonical{Input: 100, Output: 30},
		},
		{
			// Round 2: after tool result, say done.
			Text:  "Code-tools loaded successfully.",
			Usage: fakevendor.Canonical{Input: 200, Output: 20},
		},
	}
}

// ch10DisclosureReplies: load code-tools, then search-tools (progressive disclosure + depends).
func ch10DisclosureReplies() []fakevendor.Reply {
	return []fakevendor.Reply{
		{
			ToolName: "load_skill",
			ToolArgs: `{"name":"code-tools"}`,
			ToolID:   "call_disc_1",
			Usage:    fakevendor.Canonical{Input: 100, Output: 30},
		},
		{
			// After code-tools loaded, load search-tools (now available).
			ToolName: "load_skill",
			ToolArgs: `{"name":"search-tools"}`,
			ToolID:   "call_disc_2",
			Usage:    fakevendor.Canonical{Input: 200, Output: 30},
		},
		{
			Text:  "All skills loaded.",
			Usage: fakevendor.Canonical{Input: 300, Output: 20},
		},
	}
}

// ch10BlockedReplies: try to load "blocked" — should fail.
func ch10BlockedReplies() []fakevendor.Reply {
	return []fakevendor.Reply{
		{
			ToolName: "load_skill",
			ToolArgs: `{"name":"blocked"}`,
			ToolID:   "call_block_1",
			Usage:    fakevendor.Canonical{Input: 100, Output: 30},
		},
		{
			Text:  "Understood, that skill is not available.",
			Usage: fakevendor.Canonical{Input: 200, Output: 20},
		},
	}
}

// ----------------------------------------------------------------
// Main harness
// ----------------------------------------------------------------

func Ch10Run(dir string) (*Ch10Result, error) {
	dir, _ = filepath.Abs(dir)
	r := &Ch10Result{Base: dir}

	// Build.
	bin, cleanup, err := Build(dir)
	if err != nil {
		r.BuildErr = err.Error()
		return r, nil
	}
	defer cleanup()
	r.BuildOK = true

	guiDir := filepath.Join(dir, "web", "gui")

	// Phase 1: Initial tools + load_skill.
	ch10DriveInitialAndLoad(r, bin, guiDir)

	// System prompt check — verify the primary skill's body is in the system prompt.
	if r.SystemPrompt != "" {
		// The base skill body starts with "You are a helpful agent."
		if strings.Contains(r.SystemPrompt, "You are a helpful agent") {
			r.SystemPromptOK = true
		} else {
			r.SystemPromptErr = fmt.Sprintf("system prompt does not contain base skill body (len=%d)", len(r.SystemPrompt))
		}
	} else if r.SystemPromptErr == "" {
		r.SystemPromptErr = "system prompt not captured from first request"
	}

	// Phase 2: Progressive disclosure + depends.
	ch10DriveDisclosure(r, bin, guiDir)

	// Phase 3: Blocked skill.
	ch10DriveBlocked(r, bin, guiDir)

	// Phase 4: Ch9 parity.
	ch9r, err := Ch9Run(dir)
	if err != nil {
		r.Ch9Err = fmt.Sprintf("ch9 harness error: %v", err)
	} else {
		r.Ch9Result = ch9r
	}

	// Phase 5: Ensemble primary skill — all core tools visible from the start.
	ch10DriveEnsemble(r, bin, guiDir)

	return r, nil
}

// ----------------------------------------------------------------
// Phase 1: Initial tools + load_skill
// ----------------------------------------------------------------

func ch10DriveInitialAndLoad(r *Ch10Result, bin, guiDir string) {
	tmp, _ := os.MkdirTemp("", "ch10-init-*")
	defer os.RemoveAll(tmp)

	skillsDir := filepath.Join(tmp, "skills")
	ch10CreateFixtureSkills(skillsDir)

	port := freePort()
	srv := fakevendor.New(ch10InitialReplies())
	defer srv.Close()

	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"EN_CUSTOM_VAR=hello-from-grader",
		"CH02_LOG="+filepath.Join(tmp, "init.log"),
	)
	cmd.Stderr = os.Stderr
	stdin, _ := cmd.StdinPipe()
	if err := cmd.Start(); err != nil {
		r.HelpersErr = "start: " + err.Error()
		return
	}
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		r.HelpersErr = "HTTP server did not start"
		return
	}

	// Connect via WebSocket and send a prompt.
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:"+port+"/ws", nil)
	if err != nil {
		r.HelpersErr = "ws dial: " + err.Error()
		return
	}
	defer conn.Close()

	conn.WriteJSON(map[string]any{"type": "subscribe"})
	conn.WriteJSON(map[string]any{"type": "prompt", "text": "Load code-tools for me."})

	// Wait for the turn to complete.
	readWSMessages(conn, 30*time.Second, func(msgs []Ch8WsMsg) bool {
		for _, m := range msgs {
			if m.Type == "turn_ended" {
				return true
			}
		}
		return false
	})

	// Inspect the fakevendor's recorded requests.
	reqs := srv.Requests()
	if len(reqs) == 0 {
		r.InitialToolsErr = "no requests recorded by fakevendor"
		return
	}

	// Request 1: initial tool declarations.
	r.InitialTools = extractToolNamesFromRequest(reqs[0])
	r.SystemPrompt = extractSystemPromptFromRequest(reqs[0])

	if len(r.InitialTools) == 0 {
		r.InitialToolsErr = "no tools declared in first request"
	}

	// Request 2 (after load_skill result): check for new tools.
	if len(reqs) >= 2 {
		r.PostLoadTools = extractToolNamesFromRequest(reqs[1])
		r.LoadSkillResult = extractToolResultFromRequest(reqs[1])

		if len(r.PostLoadTools) > len(r.InitialTools) {
			r.LoadSkillOK = true
		} else {
			r.LoadSkillErr = fmt.Sprintf("tool count did not increase: initial=%d, after=%d",
				len(r.InitialTools), len(r.PostLoadTools))
		}
	} else {
		r.LoadSkillErr = "only one request recorded — load_skill never executed"
	}
}

// ----------------------------------------------------------------
// Phase 2: Progressive disclosure + depends
// ----------------------------------------------------------------

func ch10DriveDisclosure(r *Ch10Result, bin, guiDir string) {
	tmp, _ := os.MkdirTemp("", "ch10-disc-*")
	defer os.RemoveAll(tmp)

	skillsDir := filepath.Join(tmp, "skills")
	ch10CreateFixtureSkills(skillsDir)

	port := freePort()
	srv := fakevendor.New(ch10DisclosureReplies())
	defer srv.Close()

	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"EN_CUSTOM_VAR=hello-from-grader",
		"CH02_LOG="+filepath.Join(tmp, "disc.log"),
	)
	cmd.Stderr = os.Stderr
	stdin, _ := cmd.StdinPipe()
	if err := cmd.Start(); err != nil {
		r.DisclosureErr = "start: " + err.Error()
		return
	}
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		r.DisclosureErr = "HTTP server did not start"
		return
	}

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:"+port+"/ws", nil)
	if err != nil {
		r.DisclosureErr = "ws dial: " + err.Error()
		return
	}
	defer conn.Close()

	conn.WriteJSON(map[string]any{"type": "subscribe"})
	conn.WriteJSON(map[string]any{"type": "prompt", "text": "Load code-tools then search-tools."})

	readWSMessages(conn, 30*time.Second, func(msgs []Ch8WsMsg) bool {
		for _, m := range msgs {
			if m.Type == "turn_ended" {
				return true
			}
		}
		return false
	})

	reqs := srv.Requests()

	// Need 3 requests: initial, after code-tools, after search-tools.
	if len(reqs) < 3 {
		r.DisclosureErr = fmt.Sprintf("expected 3 requests, got %d", len(reqs))
		return
	}

	// After search-tools: check for search_files and list_directory.
	r.PostDisclosureTools = extractToolNamesFromRequest(reqs[2])
	r.DisclosureResult = extractToolResultFromRequest(reqs[2])

	foundSearch := false
	foundListDir := false
	for _, t := range r.PostDisclosureTools {
		if t == "search_files" {
			foundSearch = true
		}
		if t == "list_directory" {
			foundListDir = true
		}
	}

	if foundSearch {
		r.DisclosureOK = true
	} else {
		r.DisclosureErr = fmt.Sprintf("search_files not in tools after loading search-tools: %v",
			r.PostDisclosureTools)
	}

	if foundListDir {
		r.DependsOK = true
	} else {
		r.DependsErr = fmt.Sprintf("list_directory not in tools — search-helpers not auto-loaded: %v",
			r.PostDisclosureTools)
	}

	// Check variable substitution in the search-tools tool result.
	if strings.Contains(r.DisclosureResult, "hello-from-grader") {
		r.VarSubOK = true
		r.VarSubBody = r.DisclosureResult
	} else {
		r.VarSubErr = fmt.Sprintf("$CUSTOM_VAR not rendered — got: %s", r.DisclosureResult)
	}
}

// ----------------------------------------------------------------
// Phase 3: Blocked skill
// ----------------------------------------------------------------

func ch10DriveBlocked(r *Ch10Result, bin, guiDir string) {
	tmp, _ := os.MkdirTemp("", "ch10-block-*")
	defer os.RemoveAll(tmp)

	skillsDir := filepath.Join(tmp, "skills")
	ch10CreateFixtureSkills(skillsDir)

	port := freePort()
	srv := fakevendor.New(ch10BlockedReplies())
	defer srv.Close()

	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"EN_CUSTOM_VAR=hello-from-grader",
		"CH02_LOG="+filepath.Join(tmp, "block.log"),
	)
	cmd.Stderr = os.Stderr
	stdin, _ := cmd.StdinPipe()
	if err := cmd.Start(); err != nil {
		r.BlockedErr = "start: " + err.Error()
		return
	}
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		r.BlockedErr = "HTTP server did not start"
		return
	}

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:"+port+"/ws", nil)
	if err != nil {
		r.BlockedErr = "ws dial: " + err.Error()
		return
	}
	defer conn.Close()

	conn.WriteJSON(map[string]any{"type": "subscribe"})
	conn.WriteJSON(map[string]any{"type": "prompt", "text": "Try to load blocked."})

	readWSMessages(conn, 30*time.Second, func(msgs []Ch8WsMsg) bool {
		for _, m := range msgs {
			if m.Type == "turn_ended" {
				return true
			}
		}
		return false
	})

	reqs := srv.Requests()
	if len(reqs) < 2 {
		r.BlockedErr = fmt.Sprintf("expected 2 requests, got %d", len(reqs))
		return
	}

	// The tool result should indicate an error.
	r.BlockedResult = extractToolResultFromRequest(reqs[1])
	lower := strings.ToLower(r.BlockedResult)
	if strings.Contains(lower, "error") ||
		strings.Contains(lower, "not loadable") ||
		strings.Contains(lower, "not available") ||
		strings.Contains(lower, "cannot") ||
		strings.Contains(lower, "not found") {
		r.BlockedOK = true
	} else {
		r.BlockedErr = fmt.Sprintf("load_skill(blocked) did not return an error — got: %s", r.BlockedResult)
	}
}

// ----------------------------------------------------------------
// Request inspection helpers
// ----------------------------------------------------------------

// extractToolNamesFromRequest extracts tool names from the "tools" array
// in an Anthropic-format request body.
func extractToolNamesFromRequest(rec fakevendor.Recorded) []string {
	m := rec.JSON()
	if m == nil {
		return nil
	}

	tools, ok := m["tools"].([]any)
	if !ok {
		return nil
	}

	var names []string
	for _, t := range tools {
		tm, ok := t.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := tm["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names
}

// extractSystemPromptFromRequest extracts the system prompt text from
// the "system" field of an Anthropic-format request.
func extractSystemPromptFromRequest(rec fakevendor.Recorded) string {
	m := rec.JSON()
	if m == nil {
		return ""
	}

	// The system field can be a string or an array of content blocks.
	switch sys := m["system"].(type) {
	case string:
		return sys
	case []any:
		var parts []string
		for _, block := range sys {
			if bm, ok := block.(map[string]any); ok {
				if text, ok := bm["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

// extractToolResultFromRequest extracts the last tool_result content from
// the messages array in an Anthropic-format request.
func extractToolResultFromRequest(rec fakevendor.Recorded) string {
	m := rec.JSON()
	if m == nil {
		return ""
	}

	msgs, ok := m["messages"].([]any)
	if !ok {
		return ""
	}

	// Walk backwards to find the last tool_result.
	for i := len(msgs) - 1; i >= 0; i-- {
		msg, ok := msgs[i].(map[string]any)
		if !ok {
			continue
		}
		if msg["role"] != "user" {
			continue
		}
		content, ok := msg["content"].([]any)
		if !ok {
			continue
		}
		for j := len(content) - 1; j >= 0; j-- {
			block, ok := content[j].(map[string]any)
			if !ok {
				continue
			}
			if block["type"] == "tool_result" {
				// Content can be a string or an array of text blocks.
				switch c := block["content"].(type) {
				case string:
					return c
				case []any:
					var parts []string
					for _, p := range c {
						if pm, ok := p.(map[string]any); ok {
							if text, ok := pm["text"].(string); ok {
								parts = append(parts, text)
							}
						}
					}
					return strings.Join(parts, "\n")
				}
			}
		}
	}
	return ""
}

// ch10DriveEnsemble starts the binary with EN_PRIMARY_SKILL=ensemble and
// verifies that all core tools are visible from the start.
func ch10DriveEnsemble(r *Ch10Result, bin, guiDir string) {
	replies := []fakevendor.Reply{
		{Text: "Ensemble agent ready with all tools."},
	}
	srv := fakevendor.New(replies)
	defer srv.Close()

	tmp, err := os.MkdirTemp("", "ch10-ensemble-*")
	if err != nil {
		r.EnsembleToolsErr = fmt.Sprintf("temp dir: %v", err)
		return
	}
	defer os.RemoveAll(tmp)

	skillsDir := filepath.Join(tmp, "skills")
	if err := ch10CreateFixtureSkills(skillsDir); err != nil {
		r.EnsembleToolsErr = fmt.Sprintf("create skills: %v", err)
		return
	}

	port := freePort()

	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=test-key",
		"EN_LOG_DIR="+tmp,
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=ensemble",
		"EN_CUSTOM_VAR=hello-from-grader",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		r.EnsembleToolsErr = fmt.Sprintf("stdin pipe: %v", err)
		return
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		r.EnsembleToolsErr = fmt.Sprintf("start: %v", err)
		return
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		r.EnsembleToolsErr = "GUI server did not start"
		return
	}

	// Wait a moment for the actor loop to be ready.
	time.Sleep(500 * time.Millisecond)

	// Send a prompt via stdin.
	fmt.Fprintln(stdin, `{"kind":"prompt","text":"hello"}`)

	// Wait for the request to reach fakevendor.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if len(srv.Requests()) > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	reqs := srv.Requests()
	if len(reqs) == 0 {
		r.EnsembleToolsErr = "no requests to fakevendor"
		return
	}

	r.EnsembleTools = extractToolNamesFromRequest(reqs[0])
	r.EnsemblePrompt = extractSystemPromptFromRequest(reqs[0])
}

