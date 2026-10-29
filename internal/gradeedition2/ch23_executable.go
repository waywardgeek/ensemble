package gradeedition2

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter23Executable retains seven complete original wire-level rows. Corpus
// graduation/restart, cut lifetime and Go topology are excluded, not inferred
// from independent library tests or credited for only their simpler subchecks.
func Chapter23Executable(bin string) (grade.Report, error) {
	work, err := os.MkdirTemp("", "c23-executable-")
	if err != nil {
		return grade.Report{}, err
	}
	s := &recallSuite{bin: bin, dir: work, skills: filepath.Join(work, "skills"), data: "agents/root"}
	if err = recallWrite(s.skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Local recall grading fixture\ntype: primary\ntools: read_file\n---\nFinish main work in chat.\n"}); err != nil {
		return grade.Report{}, err
	}
	s.launch = func(o recallOpts) recallOut { return recallExecutableLaunch(s.bin, s.skills, o) }
	r := contextResult{Errs: map[string]string{}}
	recallRanking(s, &r)
	recallChunking(s, &r)
	recallQuotas(s, &r)
	recallJudging(s, &r)
	recallFallback(s, &r, true)
	recallCap(s, &r)
	checks := recallExecutableChecks()
	for i := range checks {
		if problem := r.Errs[checks[i].ID]; problem != "" {
			fail(&checks[i], "%s", problem)
		}
		checks[i].Details = append(checks[i].Details, "retained artifacts: "+work)
	}
	return grade.NewTitledReport("Edition 3, Chapter 23 — C# executable scope (70 points; corpus graduation, cut lifetime and Go topology rows excluded)", checks, ""), nil
}

func recallExecutableChecks() []grade.Check {
	included := map[string]bool{"bm25-scores": true, "bm25-stop-words": true, "chunking-splits": true, "per-source-quota": true, "judge-filters": true, "judge-fallback": true, "injection-capped": true}
	var checks []grade.Check
	for _, c := range grade.Ch17Checks(grade.Ch17Result{}) {
		if included[c.ID] {
			c.Passed, c.Earned, c.Details = true, c.Points, nil
			checks = append(checks, c)
		}
	}
	return checks
}

// Seed the real strict per-Agent document. Archive fixtures must be searched,
// not loaded into live memory before retrieval. No environment setting stands
// in for persisted policy; this host reads the document through its normal API.
func recallExecutableSettings() ([]byte, error) {
	memory := map[string]any{}
	for _, name := range []string{"conversation", "soul", "memory", "archive64", "archive8", "session"} {
		memory[name] = map[string]any{"budget": 5000, "high_watermark": 10000, "low_watermark": 5000, "disabled": true}
	}
	recall := map[string]any{"disabled": false, "disable_judge": false, "judge_model": "", "min_score": 3, "max_snippets": 3, "min_query_length": 80, "candidates": 20, "context_messages": 3, "max_recall_bytes": 6144}
	return json.Marshal(map[string]any{"version": 1, "memory": memory, "recall": recall})
}

func recallExecutableLaunch(bin, skills string, o recallOpts) (out recallOut) {
	home := filepath.Join(o.dir, "agents", "root")
	if err := os.MkdirAll(home, 0700); err != nil {
		out.fatal = err.Error()
		return
	}
	settings, err := recallExecutableSettings()
	if err == nil {
		err = os.WriteFile(filepath.Join(home, "settings.json"), settings, 0600)
	}
	if err != nil {
		out.fatal = err.Error()
		return
	}
	srv := newRecallPeer(o)
	defer srv.Close()
	defer srv.release()
	defer func() {
		out.reqs = srv.Requests()
		raw, _ := json.MarshalIndent(out.reqs, "", "  ")
		_ = os.WriteFile(filepath.Join(o.dir, "provider-requests.json"), raw, 0600)
	}()
	env := webStatusEnv(os.Environ(), o.dir, skills, srv.URL, o.model, "base")
	args := []string{"--no-stream", "--data-dir", o.dir, "--agent-name", "root", "--skills-dir", skills, "--primary-skill", "base"}
	p, err := startMCPCommandEnvTimeout(bin, o.dir, args, env, 90*time.Second)
	if err != nil {
		out.fatal = err.Error()
		return
	}
	defer p.close()
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
		out.events, err = recallExecutableEvents(filepath.Join(home, "events.jsonl"))
		if err != nil && out.fatal == "" {
			out.fatal = "read events: " + err.Error()
		}
	}()
	for i, prompt := range o.prompts {
		if err = p.send(map[string]any{"kind": "prompt", "text": prompt}); err != nil {
			out.fatal = err.Error()
			return
		}
		want := i + 1
		if i < len(o.expect) {
			want = o.expect[i]
		}
		final := "Acknowledged."
		if want > 0 && want <= len(o.replies) {
			final = o.replies[want-1].Text
		}
		if final == "" {
			out.fatal = "fixture lacks final reply"
			return
		}
		if err = c20ReadFinal(p, final); err != nil {
			out.fatal = err.Error()
			return
		}
		if o.after != nil {
			if err = o.after(i); err != nil {
				out.fatal = err.Error()
				return
			}
		}
		if n := len(recallTurns(srv.Requests())); n != want {
			out.fatal = fmt.Sprintf("turn %d made %d main requests, want %d", i+1, n, want)
			return
		}
	}
	return
}

// Hosts have already joined. Unlike a live-tail reader, a partial final line
// here is a damaged log, not permission to ignore a missing attachment.
func recallExecutableEvents(path string) ([]map[string]json.RawMessage, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) != 0 && raw[len(raw)-1] != '\n' {
		return nil, fmt.Errorf("incomplete joined event log")
	}
	var events []map[string]json.RawMessage
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 4096), 16*1024*1024)
	for scanner.Scan() {
		var event map[string]json.RawMessage
		if err = json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	var kind string
	if len(events) > 0 {
		_ = json.Unmarshal(events[len(events)-1]["type"], &kind)
	}
	if kind != "agent_shutdown" {
		return nil, fmt.Errorf("joined log lacks agent_shutdown")
	}
	return events, nil
}
