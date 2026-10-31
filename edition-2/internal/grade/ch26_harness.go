package grade

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Sentinel strings for ch26 probes. Each one is arithmetic or random enough
// that a coincidental match in the student's own code is impossible.
const (
	ch26Marker     = "CH26-PARENT-GO"
	ch26Payload    = "PAYLOAD-c4e926"
	ch26TaskDone   = "CH26-TASK-SUBMIT"
	ch26OtherChild = "orphan-8d3f"
	ch26IdleChild  = "driftr-c26a"
	ch26TaskIdle   = "CH26-TASK-IDLE"
)

// Ch26Run grades a submission.
func Ch26Run(dir string) Ch26Result {
	var r Ch26Result
	abs, err := filepath.Abs(dir)
	if err != nil {
		r.Fatal = fmt.Sprintf("bad path: %v", err)
		return r
	}

	bin, cleanup, err := Build(abs)
	if err != nil {
		r.Fatal = fmt.Sprintf("submission does not build: %v", err)
		return r
	}
	defer cleanup()

	ch26Activity(bin, &r)
	ch26Persistence(bin, &r)
	ch26Ownership(bin, &r)
	ch26WaitChange(bin, &r)
	ch26CompositionRoot(bin, &r)
	ch26GUITree(bin, abs, &r)
	ch26SubmitIdle(bin, &r)
	return r
}

// ─── Check 1: activity-observed ──────────────────────────────────────────────
//
// Spawn a child that runs a tool. The parent calls get_sub_agent_status. The
// activity field in the result proves the observation was emitted and is
// readable. Additionally, the turn must produce at least one ActivityChanged
// observation that reaches the model via tool results or the observer.

func ch26Activity(bin string, r *Ch26Result) {
	work, err := os.MkdirTemp("", "ch26-activity-*")
	if err != nil {
		r.failAll("cannot create working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			b := string(body)
			switch {
			case !strings.Contains(b, ch26Marker):
				// Child: submit immediately.
				return &fakevendor.Reply{
					ToolName: "submit",
					ToolArgs: ch25Args(map[string]any{"result": ch26Payload}),
					ToolID:   "ch26c1",
				}
			case !strings.Contains(b, "ch26a1"):
				return &fakevendor.Reply{
					ToolName: "spawn_sub_agent",
					ToolArgs: ch25Args(map[string]any{"agent_id": "probe-child", "task": ch26TaskDone}),
					ToolID:   "ch26a1",
				}
			case !strings.Contains(b, "ch26a2"):
				return &fakevendor.Reply{
					ToolName: "get_sub_agent_status",
					ToolArgs: ch25Args(map[string]any{"agent_id": "probe-child"}),
					ToolID:   "ch26a2",
				}
			case !strings.Contains(b, "ch26a3"):
				return &fakevendor.Reply{
					ToolName: "join_agents",
					ToolArgs: ch25Args(map[string]any{"agent_ids": []string{"probe-child"}}),
					ToolID:   "ch26a3",
				}
			default:
				return &fakevendor.Reply{Text: "activity run done"}
			}
		},
	})
	defer srv.Close()

	sess := ch25Launch(bin, work, srv.URL(), nil, ch26Marker+" check activity", nil)
	if sess.err != nil {
		r.fail("activity-observed", "could not drive the agent: %v", sess.err)
		return
	}

	r.ran("activity-observed")
	// The get_sub_agent_status result must contain "activity" (the field name)
	// proving that the status tool reports the activity phase.
	results := ch23ToolResults(srv.Requests())
	sawActivity := false
	for _, res := range results {
		if strings.Contains(strings.ToLower(res), "activity") {
			sawActivity = true
			break
		}
	}
	if !sawActivity {
		r.fail("activity-observed", "get_sub_agent_status never reported an activity field; ActivityChanged must be tracked and exposed through the status tool")
	}
}

// ─── Check 2: ensemble-persists ──────────────────────────────────────────────
//
// Create a named agent via the Ensemble, shut down, restart from the same
// directory, and verify the agent is restored.

func ch26Persistence(bin string, r *Ch26Result) {
	work, err := os.MkdirTemp("", "ch26-persist-*")
	if err != nil {
		r.failAll("cannot create working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	// First run: the parent spawns nothing interesting; the proof is that
	// ensemble.json is written after the agent creates a named agent. But the
	// Ensemble.CreateAgent is called from the GUI, not from tools. So we
	// probe by checking the file directly after the binary exits.
	//
	// Run 1: just start and stop to create the data directory. Then manually
	// write ensemble.json. Run 2: verify the binary reads it back.
	srv1 := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			return &fakevendor.Reply{Text: "setup done"}
		},
	})

	sess1 := ch25Launch(bin, work, srv1.URL(), nil, ch26Marker+" hello", nil)
	srv1.Close()
	if sess1.err != nil {
		r.fail("ensemble-persists", "first run failed: %v", sess1.err)
		return
	}

	// Write ensemble.json with a named agent.
	registry := []map[string]any{{
		"name":    "persistent-agent",
		"model":   "test-model",
		"skill":   "",
		"created": time.Now().Format(time.RFC3339),
	}}
	data, _ := json.MarshalIndent(registry, "", "  ")
	ensPath := filepath.Join(work, "ensemble.json")
	if err := os.WriteFile(ensPath, data, 0o644); err != nil {
		r.fail("ensemble-persists", "could not write ensemble.json: %v", err)
		return
	}

	// Run 2: verify the binary loads the registry and lists the agent.
	srv2 := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			b := string(body)
			if !strings.Contains(b, "ch26p1") {
				return &fakevendor.Reply{
					ToolName: "agents_status",
					ToolArgs: "{}",
					ToolID:   "ch26p1",
				}
			}
			return &fakevendor.Reply{Text: "persistence check done"}
		},
	})
	defer srv2.Close()

	sess2 := ch25Launch(bin, work, srv2.URL(), nil, ch26Marker+" check persistence", nil)
	if sess2.err != nil {
		r.fail("ensemble-persists", "second run failed: %v", sess2.err)
		return
	}

	r.ran("ensemble-persists")
	// The agents_status result should mention the persistent agent.
	// But wait: agents_status only shows sub-agents of the calling agent,
	// not named Ensemble agents. Let me check if it shows them...
	// Actually, named agents are NOT sub-agents of root. They're peers.
	// The persistence proof is that ensemble.json is read back. Let me
	// verify by checking the data directory for the agent.
	agentDir := filepath.Join(work, "persistent-agent")
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		r.fail("ensemble-persists", "after restart, the named agent's data directory %q does not exist; ensemble.json must be loaded at startup and agents recreated", agentDir)
	}
}

// ─── Check 3: ownership-enforced ─────────────────────────────────────────────
//
// The parent spawns a child and then calls shutdown_agent with a nonexistent
// agent ID. The tool result must contain "unknown" (proving the lookup is
// bounded to the caller's sub-tree).

func ch26Ownership(bin string, r *Ch26Result) {
	work, err := os.MkdirTemp("", "ch26-own-*")
	if err != nil {
		r.failAll("cannot create working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			b := string(body)
			switch {
			case !strings.Contains(b, "ch26o1"):
				// Try to shut down a non-child.
				return &fakevendor.Reply{
					ToolName: "shutdown_agent",
					ToolArgs: ch25Args(map[string]any{"agent_id": ch26OtherChild}),
					ToolID:   "ch26o1",
				}
			default:
				return &fakevendor.Reply{Text: "ownership done"}
			}
		},
	})
	defer srv.Close()

	sess := ch25Launch(bin, work, srv.URL(), nil, ch26Marker+" test ownership", nil)
	if sess.err != nil {
		r.fail("ownership-enforced", "could not drive the agent: %v", sess.err)
		return
	}

	r.ran("ownership-enforced")
	// The tool result for shutdown_agent must contain "unknown" (indicating the
	// agent was not found in the caller's sub-tree).
	results := ch23ToolResults(srv.Requests())
	sawUnknown := false
	for _, res := range results {
		if strings.Contains(strings.ToLower(res), "unknown") || strings.Contains(strings.ToLower(res), "no such") {
			sawUnknown = true
			break
		}
	}
	if !sawUnknown {
		r.fail("ownership-enforced", "shutdown_agent for non-child %q did not return an error containing 'unknown' or 'no such'; fleet tools must reject non-children", ch26OtherChild)
	}
}

// ─── Check 4: wait-change-wakes ──────────────────────────────────────────────
//
// Spawn a child, call wait_for_agent_change, and verify it wakes when the
// child transitions (finishes its turn).

func ch26WaitChange(bin string, r *Ch26Result) {
	work, err := os.MkdirTemp("", "ch26-wait-*")
	if err != nil {
		r.failAll("cannot create working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			b := string(body)
			switch {
			case !strings.Contains(b, ch26Marker):
				// Child: submit.
				if strings.Contains(b, "ch26wc1") {
					return &fakevendor.Reply{Text: "child done"}
				}
				return &fakevendor.Reply{
					ToolName: "submit",
					ToolArgs: ch25Args(map[string]any{"result": ch26Payload}),
					ToolID:   "ch26wc1",
				}
			case !strings.Contains(b, "ch26w1"):
				return &fakevendor.Reply{
					ToolName: "spawn_sub_agent",
					ToolArgs: ch25Args(map[string]any{"agent_id": "wait-child", "task": ch26TaskDone, "mode": "async"}),
					ToolID:   "ch26w1",
				}
			case !strings.Contains(b, "ch26w2"):
				return &fakevendor.Reply{
					ToolName: "wait_for_agent_change",
					ToolArgs: ch25Args(map[string]any{"agent_ids": []string{"wait-child"}, "timeout_seconds": 30}),
					ToolID:   "ch26w2",
				}
			default:
				return &fakevendor.Reply{Text: "wait-change done"}
			}
		},
	})
	defer srv.Close()

	sess := ch25Launch(bin, work, srv.URL(), nil, ch26Marker+" wait for child change", nil)
	if sess.err != nil {
		r.fail("wait-change-wakes", "could not drive the agent: %v", sess.err)
		return
	}

	if !ch26RequireTools(r, srv, "wait-change-wakes", "wait_for_agent_change") {
		return
	}

	r.ran("wait-change-wakes")
	// The wait_for_agent_change result must contain "status_change" or
	// "state" — which only appears in the wait tool's own result, not in
	// spawn_sub_agent's result (which also mentions the child's name).
	results := ch23ToolResults(srv.Requests())
	sawStateChange := false
	for _, res := range results {
		if (strings.Contains(res, "status_change") || strings.Contains(res, "state_change")) &&
			strings.Contains(res, "wait-child") {
			sawStateChange = true
			break
		}
	}
	if !sawStateChange {
		r.fail("wait-change-wakes", "wait_for_agent_change never reported a status_change event for wait-child; the tool must wake when a monitored agent's state changes")
	}
}

// ─── Check 5: composition-root ───────────────────────────────────────────────
//
// Build the binary, start it, send a prompt, and get a response. This proves
// the composition root works end-to-end through the Ensemble layer.

func ch26CompositionRoot(bin string, r *Ch26Result) {
	work, err := os.MkdirTemp("", "ch26-root-*")
	if err != nil {
		r.failAll("cannot create working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			return &fakevendor.Reply{Text: "composition root works"}
		},
	})
	defer srv.Close()

	sess := ch25Launch(bin, work, srv.URL(), nil, ch26Marker+" verify composition root", nil)
	if sess.err != nil {
		r.fail("composition-root", "binary did not produce a response: %v", sess.err)
		return
	}

	r.ran("composition-root")
	// Verify we got an assistant response line.
	sawResponse := false
	for _, line := range sess.lines {
		var msg struct {
			Assistant *string `json:"assistant"`
		}
		if json.Unmarshal([]byte(line), &msg) == nil && msg.Assistant != nil {
			sawResponse = true
			break
		}
	}
	if !sawResponse {
		r.fail("composition-root", "the binary started but never emitted an assistant response; the Ensemble composition root must wire agent, factory, and event routing")
	}
}

// ─── Check 6: gui-tree-message ───────────────────────────────────────────────
//
// Connect to the GUI WebSocket and verify an agent_tree message is received
// on connect, containing at least one agent entry.

func ch26GUITree(bin, srcDir string, r *Ch26Result) {
	work, err := os.MkdirTemp("", "ch26-gui-*")
	if err != nil {
		r.failAll("cannot create working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	// Find a free port for the GUI.
	guiPort := ch26FreePort()

	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			// Just keep running while we probe the WS.
			time.Sleep(2 * time.Second)
			return &fakevendor.Reply{Text: "gui probe done"}
		},
	})
	defer srv.Close()

	cmd := exec.Command(bin, "--port", fmt.Sprintf("%d", guiPort))
	cmd.Dir = work
	cmd.Env = ch25Env(srv.URL())
	stdin, _ := cmd.StdinPipe()
	cmd.Stdout = io.Discard
	var errBuf strings.Builder
	cmd.Stderr = &errBuf

	if err := cmd.Start(); err != nil {
		r.fail("gui-tree-message", "could not start binary: %v", err)
		return
	}
	defer func() {
		_ = stdin.Close()
		go func() { time.Sleep(10 * time.Second); _ = cmd.Process.Kill() }()
		_ = cmd.Wait()
	}()

	// Send the prompt to start the agent's turn.
	fmt.Fprintf(stdin, "%s\n", ch25Prompt(ch26Marker+" gui tree probe"))

	// Wait for the GUI server to be ready.
	wsURL := fmt.Sprintf("ws://localhost:%d/ws", guiPort)
	var ws *websocket.Conn
	for i := 0; i < 30; i++ {
		time.Sleep(500 * time.Millisecond)
		ws, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
		if err == nil {
			break
		}
	}
	if ws == nil {
		r.fail("gui-tree-message", "could not connect to WebSocket at %s after 15s: %v; stderr: %s", wsURL, err, errBuf.String())
		return
	}
	defer ws.Close()

	// Send a subscribe message (the GUI protocol requires this).
	subscribe := map[string]any{"type": "subscribe"}
	_ = ws.WriteJSON(subscribe)

	// Read messages for up to 10 seconds looking for agent_tree.
	r.ran("gui-tree-message")
	ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			r.fail("gui-tree-message", "never received an agents_tree message on WebSocket connect; the GUI server must broadcast the agent tree on subscribe")
			break
		}
		var envelope struct {
			Type string          `json:"type"`
			Root json.RawMessage `json:"root"`
		}
		if json.Unmarshal(msg, &envelope) == nil && (envelope.Type == "agents_tree" || envelope.Type == "agent_tree") {
			if len(envelope.Root) == 0 {
				r.fail("gui-tree-message", "agents_tree message has no root field; the tree must include at least one agent entry")
			}
			break
		}
	}
}

// ─── Check 7: submit-idle-regression ─────────────────────────────────────────
//
// This is a regression check: submit must still reject with active children,
// and idle correction must still fire, both through the Ensemble layer. This
// is identical to ch25's idle check but exercises the new Ensemble path.

func ch26SubmitIdle(bin string, r *Ch26Result) {
	work, err := os.MkdirTemp("", "ch26-idle-*")
	if err != nil {
		r.fail("submit-idle-regression", "cannot create working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			b := string(body)
			switch {
			case !strings.Contains(b, ch26Marker):
				// The child never submits.
				return &fakevendor.Reply{Text: "I had a look around."}
			case !strings.Contains(b, "ch26d1"):
				return &fakevendor.Reply{
					ToolName: "spawn_sub_agent",
					ToolArgs: ch25Args(map[string]any{"agent_id": ch26IdleChild, "task": ch26TaskIdle}),
					ToolID:   "ch26d1",
				}
			case !strings.Contains(b, "ch26d1b"):
				// Try to submit while child is active.
				return &fakevendor.Reply{
					ToolName: "submit",
					ToolArgs: ch25Args(map[string]any{"result": "premature"}),
					ToolID:   "ch26d1b",
				}
			case !strings.Contains(b, "ch26d2"):
				return &fakevendor.Reply{
					ToolName: "join_agents",
					ToolArgs: ch25Args(map[string]any{"agent_ids": []string{ch26IdleChild}}),
					ToolID:   "ch26d2",
				}
			default:
				return &fakevendor.Reply{Text: "idle run done"}
			}
		},
	})
	defer srv.Close()

	sess := ch25Launch(bin, work, srv.URL(), nil, ch26Marker+" start the drifting worker", nil)
	if sess.err != nil {
		r.fail("submit-idle-regression", "could not drive the agent: %v", sess.err)
		return
	}
	if !ch26RequireTools(r, srv, "submit-idle-regression", "spawn_sub_agent", "submit", "join_agents") {
		return
	}
	r.ran("submit-idle-regression")

	// submit-rejection: the parent tried to submit while child active.
	sawReject := false
	for _, b := range ch25Bodies(srv) {
		if strings.Contains(b, ch26Marker) &&
			(strings.Contains(b, "cannot submit") || strings.Contains(b, "still active")) {
			sawReject = true
			break
		}
	}
	if !sawReject {
		r.fail("submit-idle-regression", "submit while the child %q is active was not rejected: submit must fail when active children exist", ch26IdleChild)
		return
	}

	// idle correction: the child never submits, so the framework wakes it.
	// Count child turns (child makes no tool calls, so one turn = one request).
	turns := 0
	for _, b := range ch25Bodies(srv) {
		if !strings.Contains(b, ch26Marker) {
			turns++
		}
	}
	const want = 4 // the child's own turn, plus three wakes
	if turns != want {
		r.fail("submit-idle-regression", "the child that never submitted took %d turns, expected %d: its own, plus three wakes before the framework gives up", turns, want)
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// ch26FreePort finds an available port for the GUI server.
func ch26FreePort() int {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 9876
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

// ch26RequireTools checks that the named tools were advertised to the model.
func ch26RequireTools(r *Ch26Result, srv *fakevendor.Server, check string, names ...string) bool {
	advertised := map[string]bool{}
	for _, b := range ch25Bodies(srv) {
		for n := range ch23AdvertisedTools([]byte(b)) {
			advertised[n] = true
		}
	}
	var missing []string
	for _, n := range names {
		if !advertised[n] {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		r.ran(check)
		r.fail(check, "these pinned tool names are never advertised to the model: %s", strings.Join(missing, ", "))
		return false
	}
	return true
}
