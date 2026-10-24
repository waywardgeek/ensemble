package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

const fleetRoot = "FLEET_ROOT_22"
const fleetDone = "FLEET_COMPLETED_22"

// Route only external model replies. Every spawned child, tool job, question,
// saved conversation and idle correction belongs to the submitted executable.
type fleetDriver struct {
	bin     string
	current bool
}

func (driver fleetDriver) run(work string, route func([]byte) *fakevendor.Reply, steer func(*mcpCommand) error) (reqs []fakevendor.Recorded, err error) {
	skills := filepath.Join(work, "skills")
	err = recallWrite(skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Terminal fleet exercise\ntype: primary\ntools: spawn_sub_agent join_agents agents_status shutdown_agent send_message_to_parent respond_to_parent_message submit run_command wait_for_job tool_limits write_file read_file\n---\nMain and named-peer Agents finish in chat. Delegated Agents finish with submit. Questions are not submissions.\n"})
	if err != nil {
		return
	}
	provider := fakevendor.NewWithOptions(nil, fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		// Successful submit may end immediately or receive a final tool-free
		// assistant response. Tool absence alone does not identify a recall judge.
		if fleetHas(body, "finish") {
			return &fakevendor.Reply{Text: fleetDone}
		}
		for _, id := range []string{"child-submit", "mid-submit", "leaf-submit", "review-submit"} {
			if fleetHas(body, id) {
				return &fakevendor.Reply{Text: "Submitted work is complete."}
			}
		}
		var request map[string]any
		_ = json.Unmarshal(body, &request)
		if len(list(request["tools"])) == 0 {
			return &fakevendor.Reply{Text: "[]"}
		}
		reply := route(body)
		if driver.current {
			return chapter20ToolLimits(body, reply)
		}
		return reply
	}})
	defer provider.Close()
	var p *mcpCommand
	var e error
	if driver.current {
		p, e = startMCPCommandEnv(driver.bin, work, []string{"chat"}, chapter20Env(work, skills, provider.URL()))
	} else {
		p, e = startMCPCommand(driver.bin, work, []string{"chat"}, []string{"LLM_MODEL=claude-sonnet-5", "LLM_BASE_URL=" + provider.URL(), "EN_SKILLS_DIR=" + skills, "EN_SKILL_DIR=" + skills, "EN_PRIMARY_SKILL=base", "EN_RECALL_DISABLED=1"})
	}
	if e != nil {
		return nil, e
	}
	defer p.close()
	_, err = fmt.Fprintln(p.stdin, fleetRoot)
	if err == nil && steer != nil {
		err = steer(p)
	}
	if err == nil {
		for {
			var line string
			line, err = p.read()
			if err != nil || strings.Contains(line, fleetDone) {
				break
			}
		}
	}
	p.stdin.Close()
	select {
	case exit := <-p.done:
		p.done <- exit
		if err == nil && exit != nil {
			err = fmt.Errorf("CLI exit: %w", exit)
		}
	case <-time.After(5 * time.Second):
		p.cancel()
		exit := <-p.done
		p.done <- exit
		if err == nil {
			err = fmt.Errorf("CLI did not finish after EOF")
		}
	}
	if err != nil {
		err = fmt.Errorf("%w; stderr: %s", err, p.stderr.String())
	}
	return provider.Requests(), err
}

// The current public API uses the existing one-shot tool_limits owner rather
// than callback fields on workflow tools. Keep the operation and its ID intact.
func chapter20ToolLimits(body []byte, reply *fakevendor.Reply) *fakevendor.Reply {
	if reply == nil || (reply.ToolName != "join_agents" && reply.ToolName != "send_message_to_parent") || len(reply.Tools) != 0 {
		return reply
	}
	var args map[string]json.RawMessage
	if json.Unmarshal([]byte(reply.ToolArgs), &args) != nil {
		return reply
	}
	delay, ok := args["ai_callback_delay"]
	if !ok {
		return reply
	}
	delete(args, "ai_callback_delay")
	operation, _ := json.Marshal(args)
	limits, _ := json.Marshal(map[string]json.RawMessage{"ai_callback_delay": delay})
	adapted := *reply
	// A live question can interrupt an undispatched batch tail. Configure the
	// one-shot limit in its own response before requesting the real operation.
	if _, failed, err := anthropicToolText(body, reply.ToolID+"-limits"); err == nil {
		if failed {
			return &fakevendor.Reply{Text: "Controller tool_limits failed. " + fleetDone}
		}
		adapted.ToolArgs = string(operation)
		return &adapted
	}
	adapted.ToolID, adapted.ToolName, adapted.ToolArgs = reply.ToolID+"-limits", "tool_limits", string(limits)
	return &adapted
}

func fleetCall(id, name string, args any) *fakevendor.Reply {
	b, _ := json.Marshal(args)
	r := skillCall(id, name, string(b))
	return &r
}
func fleetHas(body []byte, id string) bool {
	_, _, err := anthropicToolText(body, id)
	return err == nil
}
func fleetFinish() *fakevendor.Reply {
	return &fakevendor.Reply{Text: fleetDone}
}

// A root's plain reply must reach actual turn completion without corrective
// wakes. Keep the socket open for a second turn: EOF must not hide a bad loop.
func fleetRootChatWith(bin string, c *grade.Check, current bool) {
	fleetTemp([]*grade.Check{c}, func(work string) {
		if err := recallWrite(filepath.Join(work, "skills"), map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Main chat control\ntype: primary\ntools: submit spawn_sub_agent\n---\nMain Agents finish in chat; delegated Agents submit.\n"}); err != nil {
			fail(c, "root chat setup: %v", err)
			return
		}
		provider := fakevendor.NewWithOptions(nil, fakevendor.Options{Route: func([]byte) *fakevendor.Reply {
			return &fakevendor.Reply{Text: "A normal main-chat answer.", Usage: fakevendor.Canonical{Input: 100, Output: 10}}
		}})
		defer provider.Close()
		p, peer, err := chapter20GUIStart(bin, work, provider.URL(), current)
		if err != nil {
			fail(c, "root chat launch: %v", err)
			return
		}
		defer p.close()
		defer peer.close()
		if err = peer.send("subscribe", ""); err != nil {
			fail(c, "root chat subscribe: %v", err)
			return
		}
		if _, err = readSettings(peer, "current_settings"); err != nil {
			fail(c, "root chat settings: %v", err)
			return
		}
		for i := 0; i < 2; i++ {
			if err = peer.send("prompt", fmt.Sprintf("Root chat turn %d.", i)); err != nil {
				fail(c, "root chat prompt: %v", err)
				return
			}
			frames, err := peer.until(3*time.Second, messageKind("turn_ended"))
			if err != nil {
				fail(c, "root chat did not finish normally: %v", err)
				return
			}
			last := frames[len(frames)-1]
			if problem, _ := last["error"].(string); problem != "" {
				fail(c, "root chat acquired submit failure: %s", problem)
				return
			}
			if len(provider.Requests()) != i+1 {
				fail(c, "root chat acquired an extra request/corrective wake")
				return
			}
		}
		if err = managementEOF(p); err != nil {
			fail(c, "root chat shutdown: %v", err)
		}
	})
}

// Result lookup is correlated to the scripted call, never a marker in its
// arguments. Repeated requests may replay the same immutable result.
func fleetResult(c *grade.Check, reqs []fakevendor.Recorded, id string, wantError bool) string {
	for _, r := range reqs {
		if fleetHas(r.Body, id) {
			return skillResult(c, r, id, wantError)
		}
	}
	fail(c, "no actual result for %s", id)
	return ""
}
func fleetTemp(checks []*grade.Check, run func(string)) {
	work, err := os.MkdirTemp("", "fleet-grade-")
	if err != nil {
		for i := range checks {
			fail(checks[i], "%v", err)
		}
		return
	}
	defer os.RemoveAll(work)
	run(work)
}
func fleetFailed(checks []grade.Check, err error) bool {
	if err == nil {
		return false
	}
	for i := range checks {
		fail(&checks[i], "fleet session: %v", err)
	}
	return true
}

func fleetChecksWith(driver fleetDriver, checks []grade.Check) {
	fleetTemp([]*grade.Check{&checks[0], &checks[1], &checks[2], &checks[3]}, func(work string) {
		reqs, err := driver.run(work, func(body []byte) *fakevendor.Reply {
			b := string(body)
			if !strings.Contains(b, fleetRoot) {
				payload := "CHILD_A_22"
				if strings.Contains(b, "TASK_B_22") {
					payload = "CHILD_B_22"
				}
				return fleetCall("child-submit", "submit", map[string]any{"result": payload})
			}
			for _, p := range []sandboxProbe{{"spawn-a", "spawn_sub_agent", map[string]any{"agent_id": "worker-a", "task": "TASK_A_22"}}, {"spawn-b", "spawn_sub_agent", map[string]any{"agent_id": "worker-b", "task": "TASK_B_22"}}, {"join-both", "join_agents", map[string]any{"agent_ids": []string{"worker-a", "worker-b"}, "ai_callback_delay": 5}}, {"status-before", "agents_status", map[string]any{}}, {"shutdown", "shutdown_agent", map[string]any{"agent_id": "worker-a"}}, {"status-after", "agents_status", map[string]any{}}} {
				if !fleetHas(body, p.id) {
					r := p.reply()
					return &r
				}
			}
			return fleetFinish()
		}, nil)
		if fleetFailed(checks[:4], err) {
			return
		}
		joined := fleetResult(&checks[0], reqs, "join-both", false)
		if !strings.Contains(joined, "CHILD_A_22") {
			fail(&checks[0], "join lost submitted child payload")
		}
		if !strings.Contains(joined, "CHILD_A_22") || !strings.Contains(joined, "CHILD_B_22") {
			fail(&checks[3], "one correlated join did not contain both submissions: %s", joined)
		}
		before := fleetResult(&checks[2], reqs, "status-before", false)
		after := fleetResult(&checks[2], reqs, "status-after", false)
		if !strings.Contains(before, "worker-a") || !strings.Contains(before, "worker-b") || strings.Contains(after, "worker-a") || !strings.Contains(after, "worker-b") {
			fail(&checks[2], "status did not remove exactly the shut-down child: before=%s after=%s", before, after)
		}
		fleetResult(&checks[2], reqs, "shutdown", false)
		// Validate separate real saves containing their own submitted facts. Merely
		// creating two empty files or observing two paths is not isolation evidence.
		owners := map[string]string{}
		_ = filepath.WalkDir(work, func(path string, d os.DirEntry, e error) error {
			if e != nil || d.IsDir() || !strings.HasPrefix(d.Name(), "save.json") {
				return nil
			}
			raw, e := os.ReadFile(path)
			if e != nil {
				return nil
			}
			// The original contract accepts a snapshot or its live journal.
			// Decode complete JSON values rather than awarding an arbitrary file.
			decoder := json.NewDecoder(bytes.NewReader(raw))
			var values []any
			for {
				var v any
				e := decoder.Decode(&v)
				if e == io.EOF {
					break
				}
				if e != nil {
					return nil
				}
				values = append(values, v)
			}
			for _, token := range []string{fleetRoot, "TASK_A_22", "TASK_B_22"} {
				if fleetExactText(values, token) {
					owners[token] = filepath.Dir(path)
				}
			}
			return nil
		})
		if owners[fleetRoot] == "" || owners["TASK_A_22"] == "" || owners["TASK_B_22"] == "" || owners[fleetRoot] == owners["TASK_A_22"] || owners[fleetRoot] == owners["TASK_B_22"] || owners["TASK_A_22"] == owners["TASK_B_22"] {
			fail(&checks[1], "distinct valid parent/child saved conversations absent: %v", owners)
		}
	})
	fleetMessagingWith(driver, &checks[4])
	fleetWake(driver, &checks[5])
	fleetIdleWith(driver, &checks[6], &checks[2])
	fleetStructuredWith(driver, checks)
}

func fleetAddressedReply(text string, current bool) (string, error) {
	var message map[string]any
	if err := json.Unmarshal([]byte(text), &message); err != nil {
		if current {
			return "", err
		}
		return text, nil
	}
	delete(message, "message_id")
	delete(message, "id")
	if current {
		for _, key := range []string{"request_id", "RequestID", "MessageID", "ID", "TurnID", "turn_id"} {
			delete(message, key)
		}
		senderValue, lower := message["sender"]
		upperValue, upper := message["Sender"]
		sender, lowerString := senderValue.(string)
		upperSender, upperString := upperValue.(string)
		if (lower && !lowerString) || (upper && !upperString) {
			return "", fmt.Errorf("reply sender must be a string")
		}
		if lower && upper && sender != upperSender {
			return "", fmt.Errorf("reply has conflicting sender fields")
		}
		if !lower {
			sender = upperSender
		}
		if sender != "middle" {
			return "", fmt.Errorf("reply lost actual middle sender: %v", message)
		}
	}
	encoded, err := json.Marshal(message)
	return string(encoded), err
}

func fleetMessagingWith(driver fleetDriver, c *grade.Check) {
	fleetTemp([]*grade.Check{c}, func(work string) {
		rootJoin := "join-mid"
		rootJoins := 1
		questionJoin := "question-join"
		questionJoins := 1
		reqs, err := driver.run(work, func(body []byte) *fakevendor.Reply {
			b := string(body)
			if strings.Contains(b, fleetRoot) {
				if !fleetHas(body, "spawn-mid") {
					return fleetCall("spawn-mid", "spawn_sub_agent", map[string]any{"agent_id": "middle", "task": "MIDDLE_TASK_22"})
				}
				if !fleetHas(body, rootJoin) {
					return fleetCall(rootJoin, "join_agents", map[string]any{"agent_ids": []string{"middle"}, "ai_callback_delay": 5})
				}
				if driver.current {
					text, _, _ := anthropicToolText(body, rootJoin)
					if !currentFleetSubmitted(text, "middle") {
						if rootJoins >= 32 {
							fail(c, "current controller exhausted 32 root fan-in observations without terminal middle: %s", text)
							return fleetFinish()
						}
						rootJoins++
						rootJoin = fmt.Sprintf("join-mid-%d", rootJoins)
						return fleetCall(rootJoin, "join_agents", map[string]any{"agent_ids": []string{"middle"}, "ai_callback_delay": 5})
					}
				}
				return fleetFinish()
			}
			if strings.Contains(b, "MIDDLE_TASK_22") {
				if !fleetHas(body, "spawn-leaf") {
					return fleetCall("spawn-leaf", "spawn_sub_agent", map[string]any{"agent_id": "leaf", "task": "LEAF_TASK_22"})
				}
				if !fleetHas(body, questionJoin) {
					return fleetCall(questionJoin, "join_agents", map[string]any{"agent_ids": []string{"leaf"}, "ai_callback_delay": 5})
				}
				if driver.current {
					text, failed, err := anthropicToolText(body, questionJoin)
					if err == nil && fleetUndispatchedJoin(text, failed) {
						if questionJoins >= 4 {
							fail(c, "four question-join observations were interrupted before dispatch")
							return fleetFinish()
						}
						questionJoins++
						questionJoin = fmt.Sprintf("question-join-%d", questionJoins)
						return fleetCall(questionJoin, "join_agents", map[string]any{"agent_ids": []string{"leaf"}, "ai_callback_delay": 5})
					}
				}
				if !fleetHas(body, "answer") {
					return fleetCall("answer", "respond_to_parent_message", map[string]any{"agent_id": "leaf", "response": "ANSWER_22"})
				}
				if !fleetHas(body, "leaf-join") {
					return fleetCall("leaf-join", "join_agents", map[string]any{"agent_ids": []string{"leaf"}, "ai_callback_delay": 5})
				}
				return fleetCall("mid-submit", "submit", map[string]any{"result": "MIDDLE_DONE_22"})
			}
			if !fleetHas(body, "ask") {
				return fleetCall("ask", "send_message_to_parent", map[string]any{"content": "QUESTION_22", "ai_callback_delay": 5})
			}
			return fleetCall("leaf-submit", "submit", map[string]any{"result": "LEAF_DONE_22"})
		}, nil)
		if err != nil {
			fail(c, "three-generation session: %v", err)
			return
		}
		for id, parts := range map[string][]string{questionJoin: {"QUESTION_22", "leaf"}, "ask": {"ANSWER_22", "middle"}, "leaf-join": {"LEAF_DONE_22"}, rootJoin: {"MIDDLE_DONE_22"}} {
			text := fleetResult(c, reqs, id, false)
			if driver.current {
				text = currentFleetJSON(text)
			}
			// A correlation ID can contain the parent's name even when the reply
			// loses its sender. Keep content/attribution, excluding those IDs.
			if id == "ask" {
				var parseErr error
				text, parseErr = fleetAddressedReply(text, driver.current)
				if parseErr != nil {
					fail(c, "invalid addressed reply: %v", parseErr)
					continue
				}
			}
			for _, part := range parts {
				if !strings.Contains(text, part) {
					fail(c, "%s lost %s: %s", id, part, text)
				}
			}
		}
		// Actual human terminal delivery/reply and same-name question isolation retain
		// their accepted built-CLI/source tests; this numeric case observes agent hops.
	})
}

var fleetHandle = regexp.MustCompile(`(?i)\b(?:handle|job)[" :=]+([0-9]+)`)

func fleetWake(driver fleetDriver, c *grade.Check) {
	fleetTemp([]*grade.Check{c}, func(work string) {
		start, release := filepath.Join(work, "started"), filepath.Join(work, "release")
		command := fmt.Sprintf("printf started > %q; while [ ! -f %q ]; do sleep 0.02; done; printf 'WAKE_RESULT_%%s' $((700+22))", start, release)
		var mu sync.Mutex
		woke := false
		handle := ""
		var defect string
		reqs, err := driver.run(work, func(body []byte) *fakevendor.Reply {
			if !fleetHas(body, "limits") {
				return fleetCall("limits", "tool_limits", map[string]any{"ai_callback_delay": 60})
			}
			if !fleetHas(body, "slow") {
				return fleetCall("slow", "run_command", map[string]any{"command": command})
			}
			text, _, _ := anthropicToolText(body, "slow")
			mu.Lock()
			defer mu.Unlock()
			if handle == "" {
				m := fleetHandle.FindStringSubmatch(text)
				if len(m) == 2 {
					handle = m[1]
				}
			}
			if strings.Contains(string(body), "STEER_WAKE_22") && !strings.Contains(text, "WAKE_RESULT_722") {
				woke = true
				_ = os.WriteFile(release, []byte("release"), 0600)
			}
			if fleetHas(body, "wait-same") {
				return fleetFinish()
			}
			if handle == "" {
				defect = "no running job handle"
				return fleetFinish()
			}
			var h int
			fmt.Sscan(handle, &h)
			return fleetCall("wait-same", "wait_for_job", map[string]any{"handle": h, "ai_callback_delay": 5})
		}, func(p *mcpCommand) error {
			deadline := time.Now().Add(3 * time.Second)
			for time.Now().Before(deadline) {
				if _, e := os.Stat(start); e == nil {
					_, e = fmt.Fprintln(p.stdin, "STEER_WAKE_22")
					return e
				}
				time.Sleep(10 * time.Millisecond)
			}
			return fmt.Errorf("real command never reached its pending barrier")
		})
		if err != nil {
			fail(c, "message-wake session: %v", err)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		if !woke || defect != "" {
			fail(c, "no causal message wake before real command release: woke=%v %s", woke, defect)
		}
		text := fleetResult(c, reqs, "wait-same", false)
		if !strings.Contains(text, "WAKE_RESULT_722") {
			fail(c, "same handle %s lost completed output: %s", handle, text)
		}
	})
}

func fleetIdleWith(driver fleetDriver, c, visibility *grade.Check) {
	fleetTemp([]*grade.Check{c}, func(work string) {
		release := filepath.Join(work, "idle-release")
		count := 0
		reqs, err := driver.run(work, func(body []byte) *fakevendor.Reply {
			if !strings.Contains(string(body), fleetRoot) {
				if !fleetHas(body, "child-hold") {
					return fleetCall("child-hold", "run_command", map[string]any{"command": fmt.Sprintf("while [ ! -f %q ]; do sleep 0.02; done", release), "ai_callback_delay": 60})
				}
				count++
				return &fakevendor.Reply{Text: "I stopped without submitting."}
			}
			for _, p := range []sandboxProbe{{"spawn-idle", "spawn_sub_agent", map[string]any{"agent_id": "drifter", "task": "IDLE_TASK_22"}}, {"status-live", "agents_status", map[string]any{}}, {"premature", "submit", map[string]any{"result": "PREMATURE_22"}}, {"release-idle", "write_file", map[string]any{"path": release, "content": "release"}}, {"idle-join", "join_agents", map[string]any{"agent_ids": []string{"drifter"}, "ai_callback_delay": 5}}, {"stop-idle", "shutdown_agent", map[string]any{"agent_id": "drifter"}}} {
				if !fleetHas(body, p.id) {
					r := p.reply()
					return &r
				}
			}
			return fleetFinish()
		}, nil)
		live := fleetResult(visibility, reqs, "status-live", false)
		if !strings.Contains(live, "drifter") || !strings.Contains(strings.ToLower(live), "running") {
			fail(visibility, "status omitted the held live child: %s", live)
		}
		if err != nil {
			fail(c, "idle session: %v", err)
			return
		}
		fleetResult(c, reqs, "premature", true)
		if count != 4 {
			fail(c, "unsubmitted child received %d completed prose responses, want its initial response plus exactly three wakes", count)
		}
		joined := ""
		for _, r := range reqs {
			if text, _, e := anthropicToolText(r.Body, "idle-join"); e == nil {
				joined = text
				break
			}
		}
		if !strings.Contains(joined, "drifter") || (!strings.Contains(strings.ToLower(joined), "fail") && !strings.Contains(strings.ToLower(joined), "without submit") && !strings.Contains(strings.ToLower(joined), "idle_without_submit")) {
			fail(c, "join lost actual exhausted-child escalation: %s", joined)
		}
	})
}

// Match stored text values, not a parent's echoed spawn task argument. The save
// remains public JSON; this does not impose a runtime struct or directory name.
func fleetExactText(v any, want string) bool {
	switch x := v.(type) {
	case map[string]any:
		if x["text"] == want {
			return true
		}
		for _, e := range x {
			if fleetExactText(e, want) {
				return true
			}
		}
	case []any:
		for _, e := range x {
			if fleetExactText(e, want) {
				return true
			}
		}
	}
	return false
}

// Later historical graders retain their original launch policy.
func fleetMessaging(bin string, c *grade.Check) { fleetMessagingWith(fleetDriver{bin: bin}, c) }
func fleetIdle(bin string, c, shutdown *grade.Check) {
	fleetIdleWith(fleetDriver{bin: bin}, c, shutdown)
}
func fleetRootChat(bin string, c *grade.Check) { fleetRootChatWith(bin, c, false) }

func fleetChecks(bin string, checks []grade.Check) { fleetChecksWith(fleetDriver{bin: bin}, checks) }

// Re-observe only a paired, explicitly undispatched observation. This is not a
// successful join or permission to retry a failed side-effecting operation.
func fleetUndispatchedJoin(text string, failed bool) bool {
	if !failed {
		return false
	}
	text = currentFleetJSON(text)
	return text == "interrupted, not executed: live input arrived before dispatch" || text == "interrupted, not executed: live input interrupted this pending batch"
}
