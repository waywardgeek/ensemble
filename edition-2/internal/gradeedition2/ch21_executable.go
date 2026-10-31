package gradeedition2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Keep the historical Go/GUI route intact. These five original rows are
// observable through the C# CLI; the two GUI/persistence rows are not credited
// by library tests or by earlier browser receipts.
func Chapter21Executable(bin, skills string) (grade.Report, error) {
	if !filepath.IsAbs(skills) {
		return grade.Report{}, fmt.Errorf("-fleet-skills must be an absolute shipped skills directory")
	}
	if _, err := os.Stat(filepath.Join(skills, "ensemble", "SKILL.md")); err != nil {
		return grade.Report{}, err
	}
	checks := grade.Ch26Checks(grade.Ch26Result{})
	for i := range checks {
		checks[i].Passed = true
		checks[i].Earned = checks[i].Points
		checks[i].Details = nil
	}
	checks[0].Title = "Actor status distinguishes held inference from idle completion without invented usage"
	checks[2].Title = "Owned-child status succeeds; cross-tree and missing targets refuse without effects"
	checks[4].Title = "Queued questions, fresh output cursors, typed retrieval and nested questions compose"
	checks[3].Title = "Change observation waits on selected children, yields on progress, and times out without killing work"
	driver := c20Executable{bin: bin, skills: skills}
	c21Management(driver, checks)
	c21Nested(driver, &checks[6])
	c20Messaging(driver, &checks[4])
	c21Idle(driver, &checks[6])
	c20RootChat(driver, &checks[6])
	selected := []grade.Check{checks[0], checks[2], checks[3], checks[4], checks[6]}
	return grade.NewTitledReport("Edition 3, Chapter 21 — C# executable scope (70 points; GUI/persistent-peer rows excluded)", selected, ""), nil
}

// One correlated external consumer of the real host's JSON control protocol.
// Job cursors can split output, so completion accumulates every observation.
// The transcript is retained even when an assertion or the process fails.
type c21Controls struct {
	p      *mcpCommand
	serial int
	rows   []string
}
type c21Reply struct {
	text string
	bad  bool
}

func (c *c21Controls) call(name string, args map[string]any) (c21Reply, error) {
	c.serial++
	id := fmt.Sprintf("management-%d", c.serial)
	if err := c.p.send(map[string]any{"kind": "tool", "id": id, "name": name, "arguments": args}); err != nil {
		return c21Reply{}, err
	}
	for {
		line, err := c.p.readWithin(20 * time.Second)
		if err != nil {
			return c21Reply{}, err
		}
		c.rows = append(c.rows, line)
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return c21Reply{}, err
		}
		if row["error"] != nil {
			return c21Reply{}, fmt.Errorf("host error: %v", row["error"])
		}
		result := object(row["tool_result"])
		if result["call_id"] != id {
			continue
		}
		text := ""
		for _, value := range list(result["parts"]) {
			part := object(value)
			if part["type"] == "text" {
				text += fmt.Sprint(part["text"])
			}
		}
		return c21Reply{text: text, bad: result["is_error"] == true}, nil
	}
}
func (c *c21Controls) finish(reply c21Reply) (c21Reply, error) {
	text := reply.text
	for attempts := 0; attempts < 100; attempts++ {
		match := c20Job.FindStringSubmatch(reply.text)
		if reply.bad || len(match) == 0 || match[2] != "running" {
			reply.text = text
			return reply, nil
		}
		var handle int
		if _, err := fmt.Sscanf(match[1], "%d", &handle); err != nil {
			return reply, err
		}
		next, err := c.call("wait_for_job", map[string]any{"handle": handle, "ai_callback_delay": 0.2})
		if err != nil {
			return reply, err
		}
		text += "\n" + next.text
		reply = next
	}
	return reply, fmt.Errorf("Job failed to finish within bounded observations")
}
func (c *c21Controls) completed(name string, args map[string]any) (c21Reply, error) {
	reply, err := c.call(name, args)
	if err != nil {
		return reply, err
	}
	return c.finish(reply)
}
func c21OK(reply c21Reply, err error) (map[string]any, error) {
	if err != nil {
		return nil, err
	}
	if reply.bad {
		return nil, fmt.Errorf("tool refused: %s", reply.text)
	}
	row := c20JSON(reply.text)
	if row == nil {
		return nil, fmt.Errorf("missing JSON result: %s", reply.text)
	}
	return row, nil
}
func (c *c21Controls) save(work string) {
	_ = os.WriteFile(filepath.Join(work, "controls.jsonl"), []byte(strings.Join(c.rows, "\n")+"\n"), 0600)
}

func c21Management(d c20Executable, checks []grade.Check) {
	selected := []*grade.Check{&checks[0], &checks[2], &checks[4], &checks[3]}
	c20Temp(selected, func(work string) {
		entered, release := make(chan struct{}), make(chan struct{})
		var started, released sync.Once
		unblock := func() { released.Do(func() { close(release) }) }
		defer unblock()
		requests, err := d.run(work, func(body []byte) *fakevendor.Reply {
			if c20Model(body) == "held-fixture" {
				if !fleetHas(body, "foreign-status") {
					started.Do(func() { close(entered) })
					select {
					case <-release:
					case <-time.After(20 * time.Second):
						return &fakevendor.Reply{Text: "FIXTURE_GATE_TIMEOUT"}
					}
					return fleetCall("foreign-status", "get_sub_agent_status", map[string]any{"agent_id": "sibling"})
				}
			}
			// Separate accepted turns use a fresh tool ID, not a prior saved completion.
			id, answer := "submitted", "FIRST_FINAL"
			if strings.Contains(memoryConversation(body), "QUEUED_QUESTION") {
				id, answer = "asked-submit", "QUEUED_ANSWER"
			}
			if strings.Contains(memoryConversation(body), "LATER_MESSAGE") {
				id, answer = "message-submit", "MESSAGE_ANSWER"
			}
			if !fleetHas(body, id) {
				return fleetCall(id, "submit", map[string]any{"result": answer})
			}
			return &fakevendor.Reply{Text: answer}
		}, func(p *mcpCommand) error {
			defer unblock()
			c := &c21Controls{p: p}
			defer c.save(work)
			if _, e := c21OK(c.completed("spawn_sub_agent", map[string]any{"agent_id": "sibling", "model": "plain-fixture", "task": "Submit normally."})); e != nil {
				return e
			}
			spawn, e := c.call("spawn_sub_agent", map[string]any{"agent_id": "held", "model": "held-fixture", "task": "Wait for the controlled response.", "ai_callback_delay": 0.05})
			if e != nil {
				return e
			}
			select {
			case <-entered:
			case <-time.After(10 * time.Second):
				return fmt.Errorf("held inference never entered")
			}
			busy, e := c21OK(c.completed("get_sub_agent_status", map[string]any{"agent_id": "held"}))
			if e != nil {
				return e
			}
			c20Require(&checks[0], busy["activity"] == "busy" && busy["work_phase"] == "thinking" && busy["submitted"] == false, "held model work did not report thinking/busy")
			c20Require(&checks[0], busy["estimated_cost_usd"] == float64(0) && busy["last_response_usage"] == nil && len(list(busy["measured_models"])) == 0, "in-flight response acquired completed usage")
			for _, name := range []string{"get_sub_agent_status", "send_message", "ask_agent", "wait_for_agent", "get_submission", "interrupt_agent", "shutdown_agent", "kill_agent"} {
				args := map[string]any{"agent_id": "missing", "content": "must not run"}
				// Exact declarations reject extra fields before dispatch; test real ownership,
				// not that generic schema guard, by supplying only each tool's own fields.
				if name != "send_message" && name != "ask_agent" {
					delete(args, "content")
				}
				refused, e := c.completed(name, args)
				if e != nil {
					return e
				}
				c20Require(&checks[2], refused.bad, "unowned target accepted by "+name)
			}
			timeout, e := c21OK(c.completed("wait_for_agent_change", map[string]any{"agent_ids": []string{"sibling"}, "timeout_seconds": 0.2, "stuck_seconds": 0}))
			if e != nil {
				return e
			}
			c20Require(&checks[3], timeout["reason"] == "timeout", "unchanged idle child did not time out")
			waiting, e := c.call("wait_for_agent_change", map[string]any{"agent_ids": []string{"held"}, "timeout_seconds": 10, "stuck_seconds": 0, "ai_callback_delay": 0.05})
			if e != nil {
				return e
			}
			match := c20Job.FindStringSubmatch(waiting.text)
			c20Require(&checks[3], !waiting.bad && len(match) > 0 && match[2] == "running", "change wait returned before held progress")
			unblock()
			changed, e := c21OK(c.finish(waiting))
			if e != nil {
				return e
			}
			c20Require(&checks[3], changed["reason"] == "changed" && len(list(changed["agents"])) == 1, "selected progress did not produce one changed child")
			if _, e = c21OK(c.finish(spawn)); e != nil {
				return e
			}
			idle, e := c21OK(c.completed("get_sub_agent_status", map[string]any{"agent_id": "held"}))
			if e != nil {
				return e
			}
			c20Require(&checks[0], idle["estimated_cost_usd"] == nil && len(list(idle["measured_models"])) > 0, "unknown completed model acquired an invented price")
			c20Require(&checks[0], idle["activity"] == "idle" && idle["work_phase"] == "idle" && idle["submitted"] == true, "completed child did not report idle/submitted")
			output, e := c21OK(c.completed("wait_for_agent", map[string]any{"agent_id": "held"}))
			if e != nil {
				return e
			}
			c20Require(&checks[4], strings.Contains(fmt.Sprint(output["output"]), "FIRST_FINAL"), "finalized output missing")
			again, e := c21OK(c.completed("wait_for_agent", map[string]any{"agent_id": "held", "cursor": output["cursor"]}))
			if e != nil {
				return e
			}
			c20Require(&checks[4], again["output"] == "", "output cursor repeated old content")
			answer, e := c21OK(c.completed("ask_agent", map[string]any{"agent_id": "held", "content": "QUEUED_QUESTION"}))
			if e != nil {
				return e
			}
			c20Require(&checks[4], answer["response"] == "QUEUED_ANSWER", "ask consumed an older turn reply")
			if _, e = c21OK(c.completed("send_message", map[string]any{"agent_id": "held", "content": "LATER_MESSAGE", "cursor": output["cursor"]})); e != nil {
				return e
			}
			if _, e = c21OK(c.completed("shutdown_agent", map[string]any{"agent_id": "held"})); e != nil {
				return e
			}
			artifact, e := c21OK(c.completed("get_submission", map[string]any{"agent_id": "held"}))
			if e != nil {
				return e
			}
			c20Require(&checks[4], artifact["historical"] == true && artifact["parent_address"] == "root" && object(artifact["payload"])["result"] == "MESSAGE_ANSWER", "fresh retained result lost after joined shutdown")
			remaining, e := c21OK(c.completed("agents_status", map[string]any{}))
			if e != nil {
				return e
			}
			c20Require(&checks[2], len(list(remaining["agents"])) == 1 && object(list(remaining["agents"])[0])["agent_id"] == "sibling", "removal damaged sibling or retrieval resurrected child")
			return nil
		})
		if err != nil {
			for _, c := range selected {
				fail(c, "management session: %v", err)
			}
			return
		}
		crossRefused := false
		for _, r := range requests {
			if c20Model(r.Body) == "held-fixture" {
				_, bad, found := c20Result(r.Body, "foreign-status")
				crossRefused = crossRefused || (found && bad)
			}
		}
		c20Require(&checks[2], crossRefused, "child could observe a real sibling owned by its parent")
	})
}

func c21Nested(d c20Executable, c *grade.Check) {
	c20Temp([]*grade.Check{c}, func(work string) {
		requests, err := d.run(work, func(body []byte) *fakevendor.Reply {
			if c20Model(body) == "leaf-fixture" {
				if reply, _, _ := c20Await(body, "leaf-question", "send_message_to_parent", map[string]any{"content": "Release the leaf after testing premature submission.", "ai_callback_delay": 30}); reply != nil {
					return reply
				}
				if !fleetHas(body, "leaf-submit") {
					return fleetCall("leaf-submit", "submit", map[string]any{"result": "leaf done"})
				}
				return &fakevendor.Reply{Text: "Leaf finished."}
			}
			if !fleetHas(body, "leaf-spawn") {
				return fleetCall("leaf-spawn", "spawn_sub_agent", map[string]any{"agent_id": "leaf", "model": "leaf-fixture", "task": "Ask the parent then submit.", "ai_callback_delay": 0.05})
			}
			if reply, _, _ := c20Await(body, "premature", "submit", map[string]any{"result": "too early"}); reply != nil {
				return reply
			}
			if !fleetHas(body, "leaf-answer") {
				reply, text, bad := c20Await(body, "question-join", "join_agents", map[string]any{"agent_ids": []string{"leaf"}, "ai_callback_delay": 5})
				if !bad && c20JSON(text)["pending_question"] != nil {
					return fleetCall("leaf-answer", "respond_to_parent_message", map[string]any{"agent_id": "leaf", "response": "Proceed."})
				}
				if reply != nil {
					return reply
				}
				return c20Done()
			}
			if reply, _, _ := c20Await(body, "nested-join", "join_agents", map[string]any{"agent_ids": []string{"leaf"}, "ai_callback_delay": 5}); reply != nil {
				return reply
			}
			if !fleetHas(body, "parent-submit") {
				return fleetCall("parent-submit", "submit", map[string]any{"result": "parent done"})
			}
			return &fakevendor.Reply{Text: "Parent finished."}
		}, func(p *mcpCommand) error {
			control := &c21Controls{p: p}
			defer control.save(work)
			if _, e := c21OK(control.completed("spawn_sub_agent", map[string]any{"agent_id": "branch", "model": "branch-fixture", "task": "Complete the nested work."})); e != nil {
				return e
			}
			joined, e := c21OK(control.completed("join_agents", map[string]any{"agent_ids": []string{"branch"}}))
			if e != nil {
				return e
			}
			c20Require(c, object(object(joined["submissions"])["branch"])["result"] == "parent done", "bottom-up typed payload did not reach owner")
			return nil
		})
		if err != nil {
			fail(c, "nested session: %v", err)
			return
		}
		refused := false
		for _, r := range requests {
			text, bad, found := c20Result(r.Body, "premature")
			refused = refused || (found && bad && strings.Contains(text, "Join or shut down unresolved children before submitting: leaf"))
		}
		c20Require(c, refused, "parent submitted while owned child was unfinished")
		for _, rel := range []string{"branch", "branch/children/leaf"} {
			bytes, e := os.ReadFile(filepath.Join(work, "agents", "root", "children", rel, "submit.md"))
			var row map[string]any
			decode := json.Unmarshal(bytes, &row)
			c20Require(c, e == nil && decode == nil && row["version"] == float64(2) && row["payload"] != nil, "missing version-2 owned artifact: "+rel)
		}
	})
}

func c21Idle(d c20Executable, c *grade.Check) {
	c20Temp([]*grade.Check{c}, func(work string) {
		requests, err := d.run(work, func([]byte) *fakevendor.Reply { return &fakevendor.Reply{Text: "No submission."} }, func(p *mcpCommand) error {
			control := &c21Controls{p: p}
			defer control.save(work)
			failed, e := control.completed("spawn_sub_agent", map[string]any{"agent_id": "idle", "task": "Submit rather than stopping."})
			if e != nil {
				return e
			}
			c20Require(c, failed.bad, "unsubmitted exhaustion reported successful task")
			state, e := c21OK(control.completed("get_sub_agent_status", map[string]any{"agent_id": "idle"}))
			if e != nil {
				return e
			}
			c20Require(c, state["submitted"] == false && state["failure"] != nil, "idle was treated as submitted")
			return nil
		})
		if err != nil {
			fail(c, "idle session: %v", err)
			return
		}
		c20Require(c, len(requests) == 4, "expected initial response and exactly three correction attempts")
		_, err = os.Stat(filepath.Join(work, "agents", "root", "children", "idle", "submit.md"))
		c20Require(c, os.IsNotExist(err), "unsubmitted child acquired durable success")
	})
}
