package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter22Executable shares the original wire/file assertions, not Go's GUI,
// event decoder, or library import exercises. The other 60 points are excluded;
// separate C# component/GUI tests must not silently earn executable credit.
func Chapter22Executable(bin string) (grade.Report, error) {
	checks := memoryExecutableChecks()
	work, err := os.MkdirTemp("", "c22-executable-")
	if err != nil {
		return grade.Report{}, err
	}
	result := &contextResult{Errs: map[string]string{}}
	run := func(o memoryOpts) contextOut { return memoryExecutableRun(bin, o) }
	base := filepath.Join(work, "facts")
	memoryCompress(run, base, result)
	memoryGraduate(run, filepath.Join(work, "graduation"), result)
	memoryNeighbor(run, filepath.Join(work, "neighbor"), result)
	memoryExecutableFresh(run, base, filepath.Join(work, "fresh"), result)
	for i := range checks {
		if problem := result.Errs[checks[i].ID]; problem != "" {
			fail(&checks[i], "%s", problem)
		}
		checks[i].Details = append(checks[i].Details, "retained artifacts: "+work)
	}
	return grade.NewTitledReport("Edition 3, Chapter 22 — C# executable scope (40 points; Go GUI/core/lifecycle/pressure rows excluded)", checks, ""), nil
}

func memoryExecutableChecks() []grade.Check {
	included := map[string]bool{
		"planted-fact-survives": true, "graduation-fires-oldest-first": true,
		"disabled-neighbor-refused": true, "fresh-start-populates": true, "memory-is-data": true,
	}
	var checks []grade.Check
	for _, check := range grade.Ch16Checks(grade.Ch16Result{}) {
		if included[check.ID] {
			check.Passed, check.Earned, check.Details = true, check.Points, nil
			checks = append(checks, check)
		}
	}
	return checks
}

// Translate the historical fixture's policy names into the exact six-band
// Agent-local document. This is fixture input, never a second runtime authority.
func memoryExecutableSettings(settings map[string]any) ([]byte, error) {
	bands := map[string]any{}
	input := object(settings["memory"])
	names := map[string]string{"conversation": "conversation", "soul": "soul", "memory": "memory", "archive64": "64x", "archive8": "8x", "session": "session"}
	for name, old := range names {
		source := object(input[old])
		budget := any(1500)
		if name == "conversation" {
			budget = 3000
		}
		if value, found := source["budget"]; found {
			budget = value
		}
		disabled := false
		if value, found := source["disabled"]; found {
			var ok bool
			disabled, ok = value.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid fixture disabled value for %s", name)
			}
		}
		bands[name] = map[string]any{"budget": budget, "high_watermark": 0, "low_watermark": 0, "disabled": disabled}
	}
	return json.Marshal(map[string]any{"version": 1, "memory": bands})
}

// Only complete lines are evidence. The host may be appending the final line.
// A terminal failure is not quiescent success, and an unlaunched outcome refuses.
func memoryExecutablePending(data []byte) (bool, error) {
	pending := map[string]bool{}
	lines := bytes.Split(data, []byte{'\n'})
	for _, line := range lines[:len(lines)-1] {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var row map[string]json.RawMessage
		if err := json.Unmarshal(line, &row); err != nil {
			return false, err
		}
		if raw := row["compression"]; raw != nil {
			var launch struct {
				ID string `json:"task_id"`
			}
			if err := json.Unmarshal(raw, &launch); err != nil {
				return false, err
			}
			_, exists := pending[launch.ID]
			if launch.ID == "" || exists {
				return false, fmt.Errorf("invalid/duplicate launch")
			}
			pending[launch.ID] = true
		}
		if raw := row["outcome"]; raw != nil {
			var outcome struct {
				ID          string `json:"task_id"`
				Disposition string `json:"state"`
			}
			if err := json.Unmarshal(raw, &outcome); err != nil {
				return false, err
			}
			if !pending[outcome.ID] || outcome.Disposition != "committed" {
				return false, fmt.Errorf("compression outcome %s: %s", outcome.ID, outcome.Disposition)
			}
			pending[outcome.ID] = false
		}
	}
	for _, active := range pending {
		if active {
			return true, nil
		}
	}
	return false, nil
}

func memoryExecutableDrain(control *c21Controls, path string) error {
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		pending, err := memoryExecutablePending(data)
		if err != nil {
			return err
		}
		if !pending {
			// The actor can launch the next waterfall attempt immediately after writing
			// an outcome. A real control barrier closes that publication/read race.
			reply, err := control.completed("agent_status", map[string]any{})
			if err != nil {
				return err
			}
			if reply.bad {
				return fmt.Errorf("status barrier: %s", reply.text)
			}
			data, err = os.ReadFile(path)
			if err != nil {
				return err
			}
			pending, err = memoryExecutablePending(data)
			if err != nil {
				return err
			}
			if !pending {
				return nil
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("compression did not reach durable finalization")
}

func memoryExecutableRun(bin string, o memoryOpts) (out contextOut) {
	if o.pending != nil || o.release != nil || o.unblock != nil || o.lengthStop || o.accounting {
		return contextOut{fatal: "unsupported mode in scoped C# runner"}
	}
	home := filepath.Join(o.dir, "agents", "root")
	skills := filepath.Join(o.dir, "fixture-skills")
	if err := os.MkdirAll(home, 0700); err != nil {
		return contextOut{fatal: err.Error()}
	}
	if err := os.MkdirAll(filepath.Join(skills, "memory-grade"), 0700); err != nil {
		return contextOut{fatal: err.Error()}
	}
	manual := "---\nname: memory-grade\ndescription: Local Chapter 22 grading fixture.\ntype: primary\ntools: micro_handoff agent_status wait_for_job\n---\nExplain tools before using them. Keep memory as ordinary data.\n"
	if err := os.WriteFile(filepath.Join(skills, "memory-grade", "SKILL.md"), []byte(manual), 0600); err != nil {
		return contextOut{fatal: err.Error()}
	}
	settings, err := memoryExecutableSettings(o.settings)
	if err == nil {
		err = os.WriteFile(filepath.Join(home, "settings.json"), settings, 0600)
	}
	if err != nil {
		return contextOut{fatal: err.Error()}
	}
	replies := append([]fakevendor.Reply(nil), o.replies...)
	// Restricted C# workers require an ordinary explanation before submit.
	// Preserve the shared fake's memory content and route classification.
	route := func(body []byte) *fakevendor.Reply {
		if o.route == nil {
			return nil
		}
		original := o.route(body)
		if original == nil {
			return nil
		}
		reply := *original
		if (reply.ToolName != "" || len(reply.Tools) != 0) && reply.Text == "" {
			reply.Text = "I will retain the supplied decisions and reasons in memory, then submit the compressed text."
		}
		return &reply
	}
	provider := fakevendor.NewWithOptions(replies, fakevendor.Options{Route: route})
	defer provider.Close()
	defer func() {
		out.reqs = provider.Requests()
		data, _ := json.MarshalIndent(out.reqs, "", "  ")
		_ = os.WriteFile(filepath.Join(o.dir, "provider-requests.json"), data, 0600)
	}()
	env := webStatusEnv(os.Environ(), o.dir, skills, provider.URL(), "memory-local-fixture", "memory-grade")
	args := []string{"--no-stream", "--data-dir", o.dir, "--agent-name", "root", "--skills-dir", skills, "--primary-skill", "memory-grade"}
	p, err := startMCPCommandEnvTimeout(bin, o.dir, args, env, 90*time.Second)
	if err != nil {
		return contextOut{fatal: err.Error()}
	}
	defer p.close()
	control := &c21Controls{p: p}
	defer control.save(o.dir)
	defer func() {
		p.stdin.Close()
		select {
		case exit := <-p.done:
			p.done <- exit
			if exit != nil && out.fatal == "" {
				out.fatal = fmt.Sprintf("CLI exit: %v", exit)
			}
		case <-time.After(5 * time.Second):
			p.cancel()
			exit := <-p.done
			p.done <- exit
			if out.fatal == "" {
				out.fatal = "CLI failed joined EOF shutdown"
			}
		}
		_ = os.WriteFile(filepath.Join(o.dir, "host.stderr"), []byte(p.stderr.String()), 0600)
	}()
	for i, prompt := range o.prompts {
		if err = p.send(map[string]any{"kind": "prompt", "text": prompt}); err != nil {
			return contextOut{fatal: err.Error()}
		}
		want := i + 1
		if i < len(o.expect) {
			want = o.expect[i]
		}
		if want < 1 || want > len(replies) {
			return contextOut{fatal: "fixture lacks final reply"}
		}
		if err = c20ReadFinal(p, replies[want-1].Text); err != nil {
			return contextOut{fatal: err.Error()}
		}
		if n := len(memoryTurns(provider.Requests())); n != want {
			return contextOut{fatal: fmt.Sprintf("turn %d made %d main requests, want %d", i, n, want)}
		}
		if err = memoryExecutableDrain(control, filepath.Join(home, "events.jsonl")); err != nil {
			return contextOut{fatal: err.Error()}
		}
	}
	return out
}

// Copy only live corpus Markdown, never a conversation, settings document,
// child submit.md, retired source, or hidden transaction receipt.
func memoryExecutableFresh(run memoryRunner, base, dir string, r *contextResult) {
	source := filepath.Join(base, "agents", "root")
	target := filepath.Join(dir, "agents", "root")
	copied := 0
	for _, pattern := range []string{"SOUL.md", "MEMORY.md", "memory/*.md", "bucket-0/*.md", "bucket-1/*.md"} {
		paths, err := filepath.Glob(filepath.Join(source, pattern))
		if err != nil {
			r.fail("fresh-start-populates", "corpus paths: %v", err)
			return
		}
		for _, path := range paths {
			body, err := os.ReadFile(path)
			if err != nil {
				r.fail("fresh-start-populates", "corpus read: %v", err)
				return
			}
			if memoryOrder(string(body)) == "" {
				continue
			}
			relative, err := filepath.Rel(source, path)
			if err != nil {
				r.fail("fresh-start-populates", "corpus relative path: %v", err)
				return
			}
			destination := filepath.Join(target, relative)
			if err = os.MkdirAll(filepath.Dir(destination), 0700); err == nil {
				err = os.WriteFile(destination, body, 0600)
			}
			if err != nil {
				r.fail("fresh-start-populates", "corpus copy: %v", err)
				return
			}
			copied++
		}
	}
	if copied == 0 {
		r.fail("fresh-start-populates", "no actual completed corpus to copy")
		return
	}
	out := run(memoryOpts{dir: dir, prompts: []string{"Read retained memory."}, expect: []int{1}, replies: []fakevendor.Reply{{Text: "Memory loaded."}}, settings: memorySettings(5000)})
	if !memoryCompleted(r, out, "fresh-start-populates") {
		return
	}
	turns := memoryTurns(out.reqs)
	if len(out.reqs) != 1 || len(turns) != 1 || !strings.Contains(memoryConversation(turns[0].Body), memoryMemo(1)) {
		r.fail("fresh-start-populates", "first request without a save omitted disk memory or invoked a compressor")
	}
}
