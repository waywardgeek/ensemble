package gradeedition2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter20Executable scores only contracts observable through the C# host.
// The historical Go GUI/import exercise (100) stays on its original route.
// Separate C# library tests do not earn those points.
func Chapter20Executable(bin, skills string) (grade.Report, error) {
	if !filepath.IsAbs(skills) {
		return grade.Report{}, fmt.Errorf("-fleet-skills must be an absolute shipped skills directory")
	}
	if _, err := os.Stat(filepath.Join(skills, "ensemble", "SKILL.md")); err != nil {
		return grade.Report{}, err
	}
	checks := grade.Ch25Checks(grade.Ch25Result{})
	for i := range checks {
		checks[i].Passed = true
		checks[i].Earned = checks[i].Points
		checks[i].Details = nil
	}
	checks[1].Title = "Parent and independently constructed children retain isolated durable conversations"
	checks[6].Title = "Delegated idle work receives exactly three corrections; ordinary root chat remains ordinary chat"
	driver := c20Executable{bin: bin, skills: skills}
	c20Structured(driver, checks)
	c20Messaging(driver, &checks[4])
	c20Wake(driver, &checks[5])
	c20Idle(driver, &checks[6])
	c20RootChat(driver, &checks[6])
	return grade.NewTitledReport("Edition 3, Chapter 20 — C# executable scope (100 points; Go GUI/reuse exercise excluded)", checks, ""), nil
}

// Retain this adapter's owned evidence on success and failure. The historical
// fleet helper intentionally removes its tree and must not be reused here.
func c20Temp(checks []*grade.Check, run func(string)) {
	work, err := os.MkdirTemp("", "c20-executable-")
	if err != nil {
		for _, c := range checks {
			fail(c, "%v", err)
		}
		return
	}
	run(work)
	for _, c := range checks {
		if !c.Passed {
			c.Details = append(c.Details, "artifacts: "+work)
		}
	}
}

type c20Executable struct{ bin, skills string }

// Sessions own only temporary directories, the fake HTTP boundary and one CLI.
// Keep requests and process diagnostics even on failure. EOF must join normally;
// killing a timed-out process is cleanup, never a successful shutdown observation.
func (d c20Executable) run(work string, route func([]byte) *fakevendor.Reply, steer func(*mcpCommand) error) ([]fakevendor.Recorded, error) {
	provider := fakevendor.NewWithOptions(nil, fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		reply := route(body)
		if reply != nil && reply.ToolName != "" && reply.Text == "" {
			copy := *reply
			copy.Text = "I will use " + reply.ToolName + " to check this local workflow."
			reply = &copy
		}
		return reply
	}})
	defer provider.Close()
	env := webStatusEnv(os.Environ(), work, d.skills, provider.URL(), "root-fixture", "ensemble")
	args := []string{"--no-stream", "--skills-dir", d.skills, "--primary-skill", "ensemble", "--startup-skills", "code-tools", "--save", filepath.Join(work, "save.json")}
	p, err := startMCPCommandEnvTimeout(d.bin, work, args, env, 60*time.Second)
	if err != nil {
		return nil, err
	}
	defer p.close()
	if steer != nil {
		err = steer(p)
	} else {
		err = p.send(map[string]any{"kind": "prompt", "text": fleetRoot})
		if err == nil {
			err = c20ReadFinal(p, fleetDone)
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
			err = fmt.Errorf("CLI did not join after EOF")
		}
	}
	requests := provider.Requests()
	data, _ := json.MarshalIndent(requests, "", "  ")
	_ = os.WriteFile(filepath.Join(work, "provider-requests.json"), data, 0600)
	_ = os.WriteFile(filepath.Join(work, "host.stderr"), []byte(p.stderr.String()), 0600)
	if err != nil {
		err = fmt.Errorf("%w; artifacts %s; stderr: %s", err, work, p.stderr.String())
	}
	return requests, err
}

func c20ReadFinal(p *mcpCommand, want string) error {
	for {
		line, err := p.readWithin(35 * time.Second)
		if err != nil {
			return err
		}
		var row map[string]any
		if json.Unmarshal([]byte(line), &row) != nil {
			return fmt.Errorf("non-JSON host output: %q", line)
		}
		if row["error"] != nil {
			return fmt.Errorf("host error: %v", row["error"])
		}
		if row["assistant"] == want {
			return nil
		}
	}
}

// Completed inline tools return JSON directly; a later Job observation prepends
// its documented status header. Parse JSON rows, not substrings of escaped text.
func c20JSON(text string) map[string]any {
	var result map[string]any
	for _, line := range strings.Split(text, "\n") {
		var row map[string]any
		if json.Unmarshal([]byte(line), &row) == nil && row != nil {
			result = row
		}
	}
	return result
}

var c20Job = regexp.MustCompile(`(?m)^job ([0-9]+), status: (running|done|cancelled|error),`)

// A correlated wait continues the same Job, never recreates a join. Repeated
// tool IDs across responses are legal; select the latest occurrence explicitly.
func c20Result(body []byte, id string) (string, bool, bool) {
	var request map[string]any
	_ = json.Unmarshal(body, &request)
	var text string
	var bad, found bool
	for _, message := range list(request["messages"]) {
		for _, part := range list(object(message)["content"]) {
			row := object(part)
			if row["type"] == "tool_result" && row["tool_use_id"] == id {
				raw, _ := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": []any{row}}}})
				text, bad, _ = anthropicToolText(raw, id)
				found = true
			}
		}
	}
	return text, bad, found
}
func c20Await(body []byte, id, name string, args any) (*fakevendor.Reply, string, bool) {
	text, bad, found := c20Result(body, id+"-wait")
	if !found {
		text, bad, found = c20Result(body, id)
	}
	if !found {
		return fleetCall(id, name, args), "", false
	}
	// Incoming hints may prevent a model-proposed observation from dispatching.
	// Retry only this exact admission refusal; real tool failures remain failures.
	if bad && text == "Hint received before dispatch; not executed." {
		original, _, _ := c20Result(body, id)
		if handle := c20Job.FindStringSubmatch(original); len(handle) != 0 {
			var number int
			_, _ = fmt.Sscan(handle[1], &number)
			return fleetCall(id+"-wait", "wait_for_job", map[string]any{"handle": number, "ai_callback_delay": 30}), text, false
		}
		return fleetCall(id, name, args), text, false
	}
	match := c20Job.FindStringSubmatch(text)
	if !bad && len(match) != 0 && match[2] == "running" {
		var handle int
		_, _ = fmt.Sscan(match[1], &handle)
		return fleetCall(id+"-wait", "wait_for_job", map[string]any{"handle": handle, "ai_callback_delay": 30}), text, false
	}
	return nil, text, bad
}
func c20Model(body []byte) string {
	var request map[string]any
	_ = json.Unmarshal(body, &request)
	value, _ := request["model"].(string)
	return value
}
func c20Done() *fakevendor.Reply { return &fakevendor.Reply{Text: fleetDone} }
func c20Require(c *grade.Check, ok bool, why string) {
	if !ok {
		fail(c, "%s", why)
	}
}
func c20ToolOK(c *grade.Check, requests []fakevendor.Recorded, id string) string {
	for i := len(requests) - 1; i >= 0; i-- {
		text, bad, found := c20Result(requests[i].Body, id)
		if found {
			c20Require(c, !bad, "tool "+id+" refused: "+text)
			return text
		}
	}
	fail(c, "missing correlated tool result %s", id)
	return ""
}
func c20Calls(requests []fakevendor.Recorded, model string) [][]byte {
	var out [][]byte
	for _, r := range requests {
		if c20Model(r.Body) == model {
			out = append(out, r.Body)
		}
	}
	return out
}

func c20Structured(d c20Executable, checks []grade.Check) {
	selected := []*grade.Check{&checks[0], &checks[1], &checks[2], &checks[3]}
	c20Temp(selected, func(work string) {
		invalid := []any{map[string]any{"summary": "missing"}, map[string]any{"summary": 7, "findings": []any{}, "follow_up": nil}, map[string]any{"summary": "nested", "findings": []any{map[string]any{"path": "main.cs"}}, "follow_up": nil}, map[string]any{"summary": "extra", "findings": []any{}, "follow_up": nil, "unexpected": true}}
		requests, err := d.run(work, func(body []byte) *fakevendor.Reply {
			model := c20Model(body)
			if model == "leaf-fixture" || model == "plain-fixture" {
				if !fleetHas(body, "child-submit") {
					return fleetCall("child-submit", "submit", map[string]any{"result": model + " complete"})
				}
				return &fakevendor.Reply{Text: "Submitted."}
			}
			if model == "review-fixture" {
				if !fleetHas(body, "leaf-spawn") {
					return fleetCall("leaf-spawn", "spawn_sub_agent", map[string]any{"agent_id": "leaf", "model": "leaf-fixture", "task": "LEAF_ONLY_TASK"})
				}
				if reply, _, _ := c20Await(body, "leaf-join", "join_agents", map[string]any{"agent_ids": []string{"leaf"}, "ai_callback_delay": 30}); reply != nil {
					return reply
				}
				for i, args := range invalid {
					id := fmt.Sprintf("invalid-%d", i)
					if !fleetHas(body, id) {
						return fleetCall(id, "submit", args)
					}
				}
				if !fleetHas(body, "review-submit") {
					return fleetCall("review-submit", "submit", fleetJSON(fleetReviewPayload))
				}
				return &fakevendor.Reply{Text: "Submitted."}
			}
			if !fleetHas(body, "bad-schema") {
				return fleetCall("bad-schema", "spawn_sub_agent", map[string]any{"agent_id": "bad", "task": "Must not run", "submit_schema": map[string]any{"unsupported": true}})
			}
			if !fleetHas(body, "review-spawn") {
				return fleetCall("review-spawn", "spawn_sub_agent", map[string]any{"agent_id": "review", "model": "review-fixture", "task": "REVIEW_ONLY_TASK", "submit_schema": fleetJSON(fleetReviewSchema)})
			}
			if !fleetHas(body, "plain-spawn") {
				return fleetCall("plain-spawn", "spawn_sub_agent", map[string]any{"agent_id": "plain", "model": "plain-fixture", "task": "PLAIN_ONLY_TASK"})
			}
			if reply, _, _ := c20Await(body, "joined", "join_agents", map[string]any{"agent_ids": []string{"review", "plain"}, "ai_callback_delay": 30}); reply != nil {
				return reply
			}
			if !fleetHas(body, "status-before") {
				return fleetCall("status-before", "agents_status", map[string]any{})
			}
			if !fleetHas(body, "remove-review") {
				return fleetCall("remove-review", "shutdown_agent", map[string]any{"agent_id": "review"})
			}
			if !fleetHas(body, "remove-plain") {
				return fleetCall("remove-plain", "shutdown_agent", map[string]any{"agent_id": "plain"})
			}
			if !fleetHas(body, "status-after") {
				return fleetCall("status-after", "agents_status", map[string]any{})
			}
			return c20Done()
		}, nil)
		if err != nil {
			for _, c := range selected {
				fail(c, "structured workflow: %v", err)
			}
			return
		}
		for _, id := range []string{"review-spawn", "plain-spawn"} {
			c20ToolOK(&checks[0], requests, id)
		}
		for _, id := range []string{"remove-review", "remove-plain"} {
			c20ToolOK(&checks[2], requests, id)
		}
		before := c20JSON(c20ToolOK(&checks[2], requests, "status-before"))
		after := c20JSON(c20ToolOK(&checks[2], requests, "status-after"))
		c20Require(&checks[2], len(list(before["agents"])) == 2 && len(list(after["agents"])) == 0, "status did not list two children then remove them")
		for _, v := range list(before["agents"]) {
			c20Require(&checks[2], object(v)["submitted"] == true, "status lost authoritative submission")
		}
		var joined map[string]any
		for i := len(requests) - 1; i >= 0; i-- {
			for _, id := range []string{"joined-wait", "joined"} {
				text, bad, found := c20Result(requests[i].Body, id)
				if found && !bad && c20JSON(text)["submissions"] != nil {
					joined = c20JSON(text)
					break
				}
			}
			if joined != nil {
				break
			}
		}
		results := object(joined["submissions"])
		c20Require(&checks[3], len(results) == 2 && reflect.DeepEqual(results["review"], fleetJSON(fleetReviewPayload)) && object(results["plain"])["result"] == "plain-fixture complete", "join lost a child's typed payload")
		c20Require(&checks[0], len(results) == 2, "spawned work did not reach parent")
		seenBad := false
		for _, r := range requests {
			_, bad, found := c20Result(r.Body, "bad-schema")
			seenBad = seenBad || (found && bad)
		}
		c20Require(&checks[0], seenBad, "unsupported schema was not refused at launch")
		for i := range invalid {
			id := fmt.Sprintf("invalid-%d", i)
			seen := false
			for _, r := range requests {
				_, bad, found := c20Result(r.Body, id)
				seen = seen || (found && bad)
			}
			c20Require(&checks[3], seen, "invalid payload accepted: "+id)
		}
		for model, want := range map[string]any{"review-fixture": fleetJSON(fleetReviewSchema), "leaf-fixture": fleetJSON(fleetDefaultSchema), "plain-fixture": fleetJSON(fleetDefaultSchema)} {
			calls := c20Calls(requests, model)
			c20Require(&checks[0], len(calls) > 0, "child made no request: "+model)
			for _, body := range calls {
				var request map[string]any
				_ = json.Unmarshal(body, &request)
				var schema any
				for _, value := range list(request["tools"]) {
					tool := object(value)
					if tool["name"] == "submit" {
						schema = tool["input_schema"]
					}
				}
				properties := object(object(schema)["properties"])
				for _, key := range []string{"ai_callback_delay", "ai_callback_pattern", "max_output_bytes"} {
					delete(properties, key)
				}
				c20Require(&checks[0], schema != nil && reflect.DeepEqual(schema, want), "child submit schema differs: "+model)
			}
			if len(calls) > 0 {
				text := memoryConversation(calls[0])
				c20Require(&checks[1], !strings.Contains(text, fleetRoot), "child inherited parent dialogue")
			}
		}
		for rel, want := range map[string]any{"children/review": fleetJSON(fleetReviewPayload), "children/review/children/leaf": map[string]any{"result": "leaf-fixture complete"}, "children/plain": map[string]any{"result": "plain-fixture complete"}} {
			data, e := os.ReadFile(filepath.Join(work, rel, "submit.md"))
			var row map[string]any
			decode := json.Unmarshal(data, &row)
			c20Require(&checks[1], e == nil && decode == nil && row["version"] == float64(1) && row["agent_id"] == filepath.Base(rel) && reflect.DeepEqual(row["payload"], want), "missing/malformed durable submission: "+rel)
			data, e = os.ReadFile(filepath.Join(work, rel, "save.json"))
			c20Require(&checks[1], e == nil && json.Valid(data), "missing independent child save: "+rel)
		}
		data, e := os.ReadFile(filepath.Join(work, "save.json"))
		c20Require(&checks[1], e == nil && json.Valid(data), "parent conversation was not saved")
	})
}

func c20Messaging(d c20Executable, c *grade.Check) {
	c20Temp([]*grade.Check{c}, func(work string) {
		requests, err := d.run(work, func(body []byte) *fakevendor.Reply {
			switch c20Model(body) {
			case "leaf-fixture":
				if reply, _, _ := c20Await(body, "leaf-question", "send_message_to_parent", map[string]any{"content": "LEAF_NEEDS_PARENT", "ai_callback_delay": 30}); reply != nil {
					return reply
				}
				if !fleetHas(body, "leaf-submit") {
					return fleetCall("leaf-submit", "submit", map[string]any{"result": "leaf done"})
				}
				return &fakevendor.Reply{Text: "Submitted."}
			case "mid-fixture":
				if !fleetHas(body, "leaf-spawn") {
					return fleetCall("leaf-spawn", "spawn_sub_agent", map[string]any{"agent_id": "leaf", "model": "leaf-fixture", "task": "Ask mid before submitting."})
				}
				if !fleetHas(body, "leaf-answer") {
					reply, text, bad := c20Await(body, "leaf-join", "join_agents", map[string]any{"agent_ids": []string{"leaf"}, "ai_callback_delay": 30})
					if !bad && c20JSON(text)["pending_question"] != nil {
						return fleetCall("leaf-answer", "respond_to_parent_message", map[string]any{"agent_id": "leaf", "response": "MID_APPROVED"})
					}
					if reply != nil {
						return reply
					}
					return c20Done() // Assertions retain the failed/missing question.
				}
				if reply, _, _ := c20Await(body, "leaf-join", "join_agents", map[string]any{"agent_ids": []string{"leaf"}, "ai_callback_delay": 30}); reply != nil {
					return reply
				}
				if reply, _, _ := c20Await(body, "mid-question", "send_message_to_parent", map[string]any{"content": "MID_NEEDS_PARENT", "ai_callback_delay": 30}); reply != nil {
					return reply
				}
				if !fleetHas(body, "mid-submit") {
					return fleetCall("mid-submit", "submit", map[string]any{"result": "mid done"})
				}
				return &fakevendor.Reply{Text: "Submitted."}
			}
			if reply, _, _ := c20Await(body, "human-question", "send_message_to_parent", map[string]any{"content": "HUMAN_APPROVAL", "ai_callback_delay": 30}); reply != nil {
				return reply
			}
			if !fleetHas(body, "mid-spawn") {
				return fleetCall("mid-spawn", "spawn_sub_agent", map[string]any{"agent_id": "mid", "model": "mid-fixture", "task": "Run the nested question workflow."})
			}
			if !fleetHas(body, "mid-answer") {
				reply, text, bad := c20Await(body, "mid-join", "join_agents", map[string]any{"agent_ids": []string{"mid"}, "ai_callback_delay": 30})
				if !bad && c20JSON(text)["pending_question"] != nil {
					return fleetCall("mid-answer", "respond_to_parent_message", map[string]any{"agent_id": "mid", "response": "ROOT_APPROVED"})
				}
				if reply != nil {
					return reply
				}
				return c20Done() // Assertions retain the failed/missing question.
			}
			if reply, _, _ := c20Await(body, "mid-join", "join_agents", map[string]any{"agent_ids": []string{"mid"}, "ai_callback_delay": 30}); reply != nil {
				return reply
			}
			return c20Done()
		}, func(p *mcpCommand) error {
			if err := p.send(map[string]any{"kind": "prompt", "text": fleetRoot}); err != nil {
				return err
			}
			for {
				line, err := p.readWithin(15 * time.Second)
				if err != nil {
					return err
				}
				var row map[string]any
				if json.Unmarshal([]byte(line), &row) != nil {
					return fmt.Errorf("invalid host JSON")
				}
				if question := object(row["human_question"]); question != nil {
					if question["content"] != "HUMAN_APPROVAL" {
						return fmt.Errorf("wrong human question: %s", line)
					}
					if err = p.send(map[string]any{"kind": "reply", "message_id": question["message_id"], "text": "HUMAN_APPROVED"}); err != nil {
						return err
					}
					break
				}
			}
			return c20ReadFinal(p, fleetDone)
		})
		if err != nil {
			fail(c, "messaging: %v", err)
			return
		}
		for _, tc := range []struct{ id, sender, response string }{{"leaf-question", "mid", "MID_APPROVED"}, {"mid-question", "root", "ROOT_APPROVED"}, {"human-question", "user", "HUMAN_APPROVED"}} {
			found := false
			for _, r := range requests {
				for _, id := range []string{tc.id, tc.id + "-wait"} {
					text, bad, exists := c20Result(r.Body, id)
					row := c20JSON(text)
					if exists && !bad && row["response"] == tc.response {
						found = true
						c20Require(c, row["sender"] == tc.sender, "reply sender missing/wrong: "+tc.id)
					}
				}
			}
			c20Require(c, found, "addressed reply absent: "+tc.id)
		}
		for _, id := range []string{"leaf-join", "mid-join"} {
			seen := false
			for _, request := range requests {
				for _, observation := range []string{id, id + "-wait"} {
					text, bad, found := c20Result(request.Body, observation)
					seen = seen || (found && !bad && c20JSON(text)["pending_question"] != nil)
				}
			}
			c20Require(c, seen, "join did not expose pending question: "+id)
		}
		for _, id := range []string{"leaf-answer", "mid-answer", "leaf-submit", "mid-submit"} {
			c20ToolOK(c, requests, id)
		}
	})
}

func c20Wake(d c20Executable, c *grade.Check) {
	c20Temp([]*grade.Check{c}, func(work string) {
		started := filepath.Join(work, "started")
		release := filepath.Join(work, "release")
		wokeBeforeRelease := false
		requests, err := d.run(work, func(body []byte) *fakevendor.Reply {
			if c20Model(body) == "wake-fixture" {
				// Keep the fake provider responsive. A real child Job waits for
				// the unrelated parent shell's marker, then admits the question.
				gate := map[string]any{"command": fmt.Sprintf("while [ ! -f %q ]; do sleep 0.01; done", started), "ai_callback_delay": 30}
				if reply, _, _ := c20Await(body, "start-gate", "run_command", gate); reply != nil {
					return reply
				}
				if !fleetHas(body, "wake-question") {
					return fleetCall("wake-question", "send_message_to_parent", map[string]any{"content": "WAKE_PARENT_NOW", "ai_callback_delay": 30})
				}
				if reply, _, _ := c20Await(body, "wake-question", "send_message_to_parent", nil); reply != nil {
					return reply
				}
				if !fleetHas(body, "wake-submit") {
					return fleetCall("wake-submit", "submit", map[string]any{"result": "wake complete"})
				}
				return &fakevendor.Reply{Text: "Submitted."}
			}
			if !fleetHas(body, "wake-spawn") {
				return fleetCall("wake-spawn", "spawn_sub_agent", map[string]any{"agent_id": "wake", "model": "wake-fixture", "task": "Ask when the unrelated shell has started."})
			}
			if !fleetHas(body, "slow-job") {
				return fleetCall("slow-job", "run_command", map[string]any{"command": "printf started > started; while [ ! -f release ]; do sleep 0.01; done; printf BACKGROUND_SURVIVED", "ai_callback_delay": 120, "ai_callback_pattern": "NEVER_MATCH"})
			}
			if !fleetHas(body, "wake-answer") {
				text, bad, _ := c20Result(body, "slow-job")
				match := c20Job.FindStringSubmatch(text)
				_, e := os.Stat(release)
				wokeBeforeRelease = !bad && len(match) > 0 && match[2] == "running" && os.IsNotExist(e) && strings.Contains(memoryConversation(body), "WAKE_PARENT_NOW")
				return fleetCall("wake-answer", "respond_to_parent_message", map[string]any{"agent_id": "wake", "response": "WAKE_APPROVED"})
			}
			if !fleetHas(body, "release-job") {
				return fleetCall("release-job", "write_file", map[string]any{"path": "release", "content": "continue"})
			}
			if reply, _, _ := c20Await(body, "slow-job", "run_command", nil); reply != nil {
				return reply
			}
			if reply, _, _ := c20Await(body, "wake-join", "join_agents", map[string]any{"agent_ids": []string{"wake"}, "ai_callback_delay": 30}); reply != nil {
				return reply
			}
			return c20Done()
		}, nil)
		if err != nil {
			fail(c, "incoming wake: %v", err)
			return
		}
		c20Require(c, wokeBeforeRelease, "incoming child message did not wake the unrelated running Job before release")
		text := c20ToolOK(c, requests, "slow-job-wait")
		delivered := false
		for _, request := range requests {
			output, bad, found := c20Result(request.Body, "slow-job-wait")
			delivered = delivered || (found && !bad && strings.Contains(output, "BACKGROUND_SURVIVED"))
		}
		final := c20Job.FindStringSubmatch(text)
		c20Require(c, delivered && len(final) != 0 && final[2] == "done" && strings.Contains(text, "exit_code: 0"), "background Job did not survive and complete")
		c20ToolOK(c, requests, "wake-answer")
		c20ToolOK(c, requests, "wake-submit")
	})
}

func c20Idle(d c20Executable, c *grade.Check) {
	c20Temp([]*grade.Check{c}, func(work string) {
		requests, err := d.run(work, func(body []byte) *fakevendor.Reply {
			if c20Model(body) == "idle-fixture" {
				return &fakevendor.Reply{Text: "I stopped without submitting."}
			}
			if !fleetHas(body, "idle-spawn") {
				return fleetCall("idle-spawn", "spawn_sub_agent", map[string]any{"agent_id": "idle", "model": "idle-fixture", "task": "Must submit, not merely stop."})
			}
			if reply, _, _ := c20Await(body, "idle-join", "join_agents", map[string]any{"agent_ids": []string{"idle"}, "ai_callback_delay": 30}); reply != nil {
				return reply
			}
			if !fleetHas(body, "idle-status") {
				return fleetCall("idle-status", "agents_status", map[string]any{})
			}
			return c20Done()
		}, nil)
		if err != nil {
			fail(c, "idle correction: %v", err)
			return
		}
		c20Require(c, len(c20Calls(requests, "idle-fixture")) == 4, "expected initial response plus exactly three corrective wakes")
		refused := false
		for _, r := range requests {
			for _, id := range []string{"idle-join", "idle-join-wait"} {
				_, bad, found := c20Result(r.Body, id)
				refused = refused || (found && bad)
			}
		}
		c20Require(c, refused, "join silently accepted idle unsubmitted child")
		row := c20JSON(c20ToolOK(c, requests, "idle-status"))
		children := list(row["agents"])
		if len(children) != 1 {
			fail(c, "failed child disappeared from status")
		} else {
			child := object(children[0])
			c20Require(c, child["submitted"] == false && child["failure"] != nil && child["failure"] != "", "idle exhaustion was not explicit failure")
		}
		_, err = os.Stat(filepath.Join(work, "children", "idle", "submit.md"))
		c20Require(c, os.IsNotExist(err), "plain text was written as a submission")
	})
}

func c20RootChat(d c20Executable, c *grade.Check) {
	c20Temp([]*grade.Check{c}, func(work string) {
		const answer = "ORDINARY_ROOT_ANSWER"
		requests, err := d.run(work, func([]byte) *fakevendor.Reply { return &fakevendor.Reply{Text: answer} }, func(p *mcpCommand) error {
			for _, text := range []string{"First ordinary question", "Second ordinary question"} {
				if err := p.send(map[string]any{"kind": "prompt", "text": text}); err != nil {
					return err
				}
				if err := c20ReadFinal(p, answer); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			fail(c, "root chat: %v", err)
			return
		}
		c20Require(c, len(requests) == 2, "ordinary root chat was subjected to delegated submit correction")
	})
}
