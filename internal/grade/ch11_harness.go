package grade

// ch11_harness.go — Persistence grader.
//
// The chapter's claim is that an agent can be written to disk and read
// back without losing or duplicating a single turn. This harness tests
// that claim the only way it can be tested honestly: from outside.
//
// Two rules govern everything below.
//
// It grades BEHAVIOR, not code. The only things it inspects are the
// bytes of the save file's top-level fields and the bodies of the HTTP
// requests the agent sends to the vendor. It never reads the student's
// source, never calls into their packages, and never trusts a verdict
// the student's own program printed — a `verify` that prints MATCH
// unconditionally earns nothing here.
//
// It never parses event internals. Events are the student's design:
// their field names, their shapes, their meanings are all "Yours".
// Where a check needs an older snapshot attached to a newer log, it
// SPLICES the top-level JSON fields of two save files and hands the
// result back to the agent. The grader has no opinion about what an
// event contains; it only has an opinion about what the agent must do
// with one.

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

// Ch11Result carries the outcomes of the persistence checks.
type Ch11Result struct {
	SaveShapeOK  bool
	SaveShapeErr string

	DefaultLoadOK  bool
	DefaultLoadErr string

	ReplayEqualsSnapshotOK  bool
	ReplayEqualsSnapshotErr string

	TailAppliedOnceOK  bool
	TailAppliedOnceErr string

	LogNotNeededOK  bool
	LogNotNeededErr string

	BadSaveRefusedOK  bool
	BadSaveRefusedErr string

	Ch10Parity    bool
	Ch10ParityErr string

	// asWrittenBody is the request produced by loading the turn-3 save
	// exactly as the agent wrote it. replay-equals-snapshot has to
	// produce it anyway, and tail-applied-once needs the same thing to
	// compare against, so it is passed along rather than re-measured.
	asWrittenBody string
}

// The prompts of the first session. Two checks look for these strings
// inside a later request body, which is what "the agent remembered"
// looks like from the vendor's side of the wire.
const (
	ch11P1 = "Hello, who are you?"
	ch11P2 = "What is the meaning of life?"
	ch11P3 = "Goodbye for now!"
	ch11P4 = "Do you remember me?"
)

// ch11Reply is the fake vendor's answer. Every run that is compared
// byte-for-byte against another run must be fed the same replies, or
// the difference in the transcript would masquerade as a replay bug.
func ch11Reply(text string) fakevendor.Reply {
	return fakevendor.Reply{
		Text:  text,
		Usage: fakevendor.Canonical{Input: 100, Output: 30},
	}
}

// ch11Opts describes one launch of the student's agent.
type ch11Opts struct {
	dir     string   // working directory: where ./save.json lives
	args    []string // extra flags beyond --port and --gui-dir
	prompts []string
	replies []fakevendor.Reply
	// noServe expects the agent to die before the GUI is up, so the
	// harness must not wait for HTTP or it would wait out the timeout.
	noServe bool
}

// ch11Out is what one launch produced.
type ch11Out struct {
	reqs     []fakevendor.Recorded
	exitCode int
	fatal    string // harness-level failure; the check cannot be judged
}

// ch11Launch runs the agent once and records every request it made.
//
// Each launch gets its own working directory from the caller. That
// matters more in this chapter than in any other: the agent now loads
// ./save.json with no flag at all, so two runs sharing a directory
// would load each other's conversations, and a check would pass or
// fail for reasons that have nothing to do with what it tests.
func ch11Launch(bin, skillsDir, guiDir string, o ch11Opts) ch11Out {
	srv := fakevendor.New(o.replies)
	defer srv.Close()

	port := freePort()
	args := append([]string{"--port", port, "--gui-dir", guiDir}, o.args...)
	cmd := exec.Command(bin, args...)
	cmd.Dir = o.dir
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_MODEL=fake-model",
		"LLM_VENDOR=anthropic",
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"CH02_LOG="+filepath.Join(o.dir, "events.jsonl"),
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ch11Out{fatal: fmt.Sprintf("stdin pipe: %v", err)}
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return ch11Out{fatal: fmt.Sprintf("start: %v", err)}
	}

	done := make(chan error, 1)

	if o.noServe {
		// The agent is expected to refuse and exit. Give it the same
		// 10 seconds rule 6 allows for an orderly shutdown.
		go func() { done <- cmd.Wait() }()
		var werr error
		select {
		case werr = <-done:
		case <-time.After(10 * time.Second):
			cmd.Process.Kill()
			<-done
			stdin.Close()
			return ch11Out{fatal: "agent did not exit within 10s"}
		}
		stdin.Close()
		return ch11Out{reqs: srv.Requests(), exitCode: ch11ExitCode(werr)}
	}

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		return ch11Out{fatal: "GUI server did not start"}
	}

	for _, p := range o.prompts {
		line, _ := json.Marshal(map[string]string{"kind": "prompt", "text": p})
		fmt.Fprintln(stdin, string(line))
		time.Sleep(500 * time.Millisecond)
	}
	// Let the last reply land before stdin closes.
	time.Sleep(2 * time.Second)

	// Rule 6: closing stdin is the save trigger, and the agent has ten
	// seconds to write the file and go.
	stdin.Close()
	go func() { done <- cmd.Wait() }()
	var werr error
	select {
	case werr = <-done:
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		<-done
		return ch11Out{reqs: srv.Requests(), fatal: "agent did not save and exit within 10s of stdin closing"}
	}
	return ch11Out{reqs: srv.Requests(), exitCode: ch11ExitCode(werr)}
}

func ch11ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	return -1
}

// ch11LastBody is the body of the final request the agent sent.
func ch11LastBody(o ch11Out) string {
	if len(o.reqs) == 0 {
		return ""
	}
	return string(o.reqs[len(o.reqs)-1].Body)
}

// Ch11Run drives all persistence checks.
func Ch11Run(path string) Ch11Result {
	r := Ch11Result{}

	bin, cleanup, err := Build(path)
	if err != nil {
		r.setAll("build: " + err.Error())
		return r
	}
	defer cleanup()

	tmp, err := os.MkdirTemp("", "ch11-grade-*")
	if err != nil {
		r.setAll("tmpdir: " + err.Error())
		return r
	}
	defer os.RemoveAll(tmp)

	guiDir := filepath.Join(tmp, "gui")
	os.MkdirAll(guiDir, 0755)
	skillsDir := filepath.Join(tmp, "skills")
	createCh11Skills(skillsDir)

	// ---- Fixture: two sessions in ONE directory, no flags ----------
	//
	// This is both the default-load check and the source of the two
	// save files every later check is built from. Session 1 holds two
	// turns; session 2 starts in the same directory with no flags, so
	// the only way it can know about session 1 is by having loaded
	// ./save.json of its own accord.
	dirA := filepath.Join(tmp, "session")
	if err := os.MkdirAll(dirA, 0755); err != nil {
		r.setAll("mkdir session: " + err.Error())
		return r
	}
	savedA := filepath.Join(dirA, "save.json")

	run1 := ch11Launch(bin, skillsDir, guiDir, ch11Opts{
		dir:     dirA,
		prompts: []string{ch11P1, ch11P2},
		replies: []fakevendor.Reply{
			ch11Reply("Hello! I am your assistant."),
			ch11Reply("The meaning of life is 42."),
		},
	})
	if run1.fatal != "" {
		r.setAll("session 1: " + run1.fatal)
		return r
	}
	s2Bytes, err := os.ReadFile(savedA)
	if err != nil {
		r.setAll(fmt.Sprintf("session 1 wrote no %s: %v — rule 1 says the "+
			"save file is save.json in the working directory, with no flag", savedA, err))
		return r
	}

	run2 := ch11Launch(bin, skillsDir, guiDir, ch11Opts{
		dir:     dirA,
		prompts: []string{ch11P3},
		replies: []fakevendor.Reply{ch11Reply("Goodbye, and thanks for all the fish.")},
	})
	if run2.fatal != "" {
		r.setAll("session 2: " + run2.fatal)
		return r
	}
	s3Bytes, err := os.ReadFile(savedA)
	if err != nil {
		r.setAll(fmt.Sprintf("session 2 wrote no %s: %v", savedA, err))
		return r
	}

	ch11CheckDefaultLoad(run2, &r)
	ch11CheckSaveShape(s3Bytes, &r)
	ch11CheckReplayEqualsSnapshot(bin, skillsDir, guiDir, tmp, s3Bytes, &r)
	ch11CheckTailAppliedOnce(bin, skillsDir, guiDir, tmp, s2Bytes, s3Bytes, &r)
	ch11CheckLogNotNeeded(bin, skillsDir, guiDir, tmp, s3Bytes, &r)
	ch11CheckBadSaveRefused(bin, skillsDir, guiDir, tmp, &r)

	return r
}

func (r *Ch11Result) setAll(msg string) {
	r.SaveShapeErr = msg
	r.DefaultLoadErr = msg
	r.ReplayEqualsSnapshotErr = msg
	r.TailAppliedOnceErr = msg
	r.LogNotNeededErr = msg
	r.BadSaveRefusedErr = msg
}

// ---------------------------------------------------------------- //
// default-load (20)
// ---------------------------------------------------------------- //

// The second session was started in the first session's directory with
// no flags at all. If its very first request to the vendor carries the
// first session's prompts, the agent loaded by default. Nothing else
// could have put those words on the wire.
func ch11CheckDefaultLoad(run2 ch11Out, r *Ch11Result) {
	if len(run2.reqs) == 0 {
		r.DefaultLoadErr = "second run sent no request to the vendor"
		return
	}
	first := string(run2.reqs[0].Body)
	for _, want := range []string{ch11P1, ch11P2} {
		if !strings.Contains(first, want) {
			r.DefaultLoadErr = fmt.Sprintf(
				"second run in the same directory, no flags: first request does not "+
					"carry the earlier prompt %q — the agent did not load ./save.json at start", want)
			return
		}
	}
	if !strings.Contains(first, ch11P3) {
		r.DefaultLoadErr = "second run's first request is missing its own new prompt"
		return
	}
	r.DefaultLoadOK = true
}

// ---------------------------------------------------------------- //
// save-shape (15)
// ---------------------------------------------------------------- //

// ch11SaveTop is the save file read the way the grader is allowed to
// read it: top-level fields only, with the log left as opaque blobs.
type ch11SaveTop struct {
	Config struct {
		Model        string `json:"model"`
		Vendor       string `json:"vendor"`
		SystemPrompt string `json:"system_prompt"`
		Tools        []struct {
			Name string `json:"name"`
		} `json:"tools"`
	} `json:"config"`
	AsOf    int64             `json:"as_of"`
	Context json.RawMessage   `json:"context"`
	Log     []json.RawMessage `json:"log"`
}

// ch11LogSeqs pulls the one event field the contract names — seq — and
// nothing else. Rule 6 is a statement about ordering, and ordering is
// the grader's business; what an event MEANS is not.
func ch11LogSeqs(log []json.RawMessage) ([]int64, error) {
	seqs := make([]int64, 0, len(log))
	for i, raw := range log {
		var e struct {
			Seq int64 `json:"seq"`
		}
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("log[%d]: %v", i, err)
		}
		seqs = append(seqs, e.Seq)
	}
	return seqs, nil
}

func ch11CheckSaveShape(data []byte, r *Ch11Result) {
	var sf ch11SaveTop
	if err := json.Unmarshal(data, &sf); err != nil {
		r.SaveShapeErr = fmt.Sprintf("save file does not parse as a SaveFile: %v", err)
		return
	}
	for _, f := range []struct{ name, val string }{
		{"config.model", sf.Config.Model},
		{"config.vendor", sf.Config.Vendor},
		{"config.system_prompt", sf.Config.SystemPrompt},
	} {
		if f.val == "" {
			r.SaveShapeErr = fmt.Sprintf("%s is empty — rule 7: the save records what shaped the wire", f.name)
			return
		}
	}
	names := map[string]bool{}
	for _, t := range sf.Config.Tools {
		names[t.Name] = true
	}
	for _, want := range []string{"read_file", "think"} {
		if !names[want] {
			r.SaveShapeErr = fmt.Sprintf("config.tools is missing %q", want)
			return
		}
	}

	if len(sf.Log) == 0 {
		r.SaveShapeErr = "save file has an empty log after a two-session conversation — " +
			"rule 6: the saved log is the loaded log plus every event created since"
		return
	}
	seqs, err := ch11LogSeqs(sf.Log)
	if err != nil {
		r.SaveShapeErr = "every log event needs a seq: " + err.Error()
		return
	}
	for i := 1; i < len(seqs); i++ {
		if seqs[i] <= seqs[i-1] {
			r.SaveShapeErr = fmt.Sprintf(
				"log seq not strictly increasing: log[%d].seq=%d follows log[%d].seq=%d",
				i, seqs[i], i-1, seqs[i-1])
			return
		}
	}
	last := seqs[len(seqs)-1]
	if sf.AsOf != last {
		r.SaveShapeErr = fmt.Sprintf(
			"as_of is %d but the last log event has seq %d — rule 6: as_of is the Seq "+
				"of the last event folded into the saved context", sf.AsOf, last)
		return
	}
	r.SaveShapeOK = true
}

// ---------------------------------------------------------------- //
// replay-equals-snapshot (20)
// ---------------------------------------------------------------- //

// ch11LoadAndAsk drops a save file into a fresh directory, starts the
// agent there with no flags, asks one question, and returns the run.
// A fresh directory per call is what keeps these runs independent.
func ch11LoadAndAsk(bin, skillsDir, guiDir, tmp, name string, save []byte, prompt string) (ch11Out, error) {
	dir := filepath.Join(tmp, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ch11Out{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, "save.json"), save, 0644); err != nil {
		return ch11Out{}, err
	}
	out := ch11Launch(bin, skillsDir, guiDir, ch11Opts{
		dir:     dir,
		prompts: []string{prompt},
		replies: []fakevendor.Reply{ch11Reply("I remember our conversation!")},
	})
	return out, nil
}

// ch11WithNullContext rewrites only the top-level "context" field to
// null, leaving every other byte of the file — including the log and
// the anchor — exactly as the agent wrote it.
func ch11WithNullContext(data []byte) ([]byte, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, err
	}
	top["context"] = json.RawMessage("null")
	return json.MarshalIndent(top, "", "  ")
}

func ch11CheckReplayEqualsSnapshot(bin, skillsDir, guiDir, tmp string, s3 []byte, r *Ch11Result) {
	nulled, err := ch11WithNullContext(s3)
	if err != nil {
		r.ReplayEqualsSnapshotErr = fmt.Sprintf("save file does not parse: %v", err)
		return
	}

	asWritten, err := ch11LoadAndAsk(bin, skillsDir, guiDir, tmp, "replay-snapshot", s3, ch11P4)
	if err != nil {
		r.ReplayEqualsSnapshotErr = "harness: " + err.Error()
		return
	}
	fromLog, err := ch11LoadAndAsk(bin, skillsDir, guiDir, tmp, "replay-null", nulled, ch11P4)
	if err != nil {
		r.ReplayEqualsSnapshotErr = "harness: " + err.Error()
		return
	}
	if asWritten.fatal != "" {
		r.ReplayEqualsSnapshotErr = "loading the save as written: " + asWritten.fatal
		return
	}
	if fromLog.fatal != "" {
		r.ReplayEqualsSnapshotErr = "loading the save with context null: " + fromLog.fatal
		return
	}

	a, b := ch11LastBody(asWritten), ch11LastBody(fromLog)
	if a == "" {
		r.ReplayEqualsSnapshotErr = "loading the save as written sent no request to the vendor"
		return
	}
	if b == "" {
		r.ReplayEqualsSnapshotErr = "loading the save with context null sent no request to the vendor"
		return
	}
	if a != b {
		r.ReplayEqualsSnapshotErr = "rule 8: loading the save as written and loading it with " +
			"context null produced DIFFERENT vendor requests.\n" +
			ch11Diff("snapshot", a, "rebuilt-from-log", b)
		return
	}
	r.ReplayEqualsSnapshotOK = true
	// Handed on so tail-applied-once can compare against the same
	// as-written request instead of paying for a second identical run.
	r.asWrittenBody = a
}

// ---------------------------------------------------------------- //
// tail-applied-once (15)
// ---------------------------------------------------------------- //

// The splice: the turn-2 save's snapshot and anchor, carrying the
// turn-3 save's log. An agent that honours as_of applies only the
// events above it and lands on the turn-3 context. An agent that skips
// the tail is a turn behind; an agent that replays from the start over
// the snapshot says everything twice. Both are visible on the wire,
// and neither requires the grader to know what an event is.
func ch11Splice(s2, s3 []byte) ([]byte, error) {
	var a, b map[string]json.RawMessage
	if err := json.Unmarshal(s2, &a); err != nil {
		return nil, fmt.Errorf("turn-2 save: %v", err)
	}
	if err := json.Unmarshal(s3, &b); err != nil {
		return nil, fmt.Errorf("turn-3 save: %v", err)
	}
	for _, k := range []string{"as_of", "context"} {
		if _, ok := a[k]; !ok {
			return nil, fmt.Errorf("turn-2 save has no %q field", k)
		}
	}
	if _, ok := b["log"]; !ok {
		return nil, fmt.Errorf("turn-3 save has no %q field", "log")
	}
	return json.MarshalIndent(map[string]json.RawMessage{
		"config":  b["config"],
		"as_of":   a["as_of"],
		"context": a["context"],
		"log":     b["log"],
	}, "", "  ")
}

func ch11CheckTailAppliedOnce(bin, skillsDir, guiDir, tmp string, s2, s3 []byte, r *Ch11Result) {
	spliced, err := ch11Splice(s2, s3)
	if err != nil {
		r.TailAppliedOnceErr = "splice: " + err.Error()
		return
	}

	// Guard against a fixture that cannot tell the two apart: if the
	// turn-2 anchor already covers the whole turn-3 log, there is no
	// tail and the check would pass for free.
	var a, b ch11SaveTop
	if json.Unmarshal(s2, &a) == nil && json.Unmarshal(s3, &b) == nil {
		if seqs, err := ch11LogSeqs(b.Log); err == nil && len(seqs) > 0 {
			if seqs[len(seqs)-1] <= a.AsOf {
				r.TailAppliedOnceErr = fmt.Sprintf(
					"fixture has no tail to apply: turn-2 as_of is %d and the turn-3 log "+
						"ends at seq %d", a.AsOf, seqs[len(seqs)-1])
				return
			}
		}
	}

	out, err := ch11LoadAndAsk(bin, skillsDir, guiDir, tmp, "tail-splice", spliced, ch11P4)
	if err != nil {
		r.TailAppliedOnceErr = "harness: " + err.Error()
		return
	}
	if out.fatal != "" {
		r.TailAppliedOnceErr = "loading the spliced save: " + out.fatal
		return
	}
	got := ch11LastBody(out)
	if got == "" {
		r.TailAppliedOnceErr = "loading the spliced save sent no request to the vendor"
		return
	}

	want := r.asWrittenBody
	if want == "" {
		// replay-equals-snapshot did not get far enough to leave one.
		ref, err := ch11LoadAndAsk(bin, skillsDir, guiDir, tmp, "tail-direct", s3, ch11P4)
		if err != nil {
			r.TailAppliedOnceErr = "harness: " + err.Error()
			return
		}
		if ref.fatal != "" {
			r.TailAppliedOnceErr = "loading the turn-3 save: " + ref.fatal
			return
		}
		want = ch11LastBody(ref)
	}
	if want == "" {
		r.TailAppliedOnceErr = "loading the turn-3 save sent no request to the vendor"
		return
	}
	if got != want {
		r.TailAppliedOnceErr = "rule 3: the turn-2 snapshot spliced onto the turn-3 log did " +
			"not produce the same request as the turn-3 save.\n" +
			ch11Diff("spliced", got, "turn-3-save", want)
		return
	}
	r.TailAppliedOnceOK = true
}

// ---------------------------------------------------------------- //
// log-not-needed (10)
// ---------------------------------------------------------------- //

func ch11CheckLogNotNeeded(bin, skillsDir, guiDir, tmp string, s3 []byte, r *Ch11Result) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(s3, &top); err != nil {
		r.LogNotNeededErr = fmt.Sprintf("save file does not parse: %v", err)
		return
	}
	var orig ch11SaveTop
	if err := json.Unmarshal(s3, &orig); err != nil {
		r.LogNotNeededErr = fmt.Sprintf("save file does not parse: %v", err)
		return
	}
	top["log"] = json.RawMessage("[]")
	trimmed, err := json.MarshalIndent(top, "", "  ")
	if err != nil {
		r.LogNotNeededErr = "harness: " + err.Error()
		return
	}

	dir := filepath.Join(tmp, "log-not-needed")
	if err := os.MkdirAll(dir, 0755); err != nil {
		r.LogNotNeededErr = "harness: " + err.Error()
		return
	}
	savePath := filepath.Join(dir, "save.json")
	if err := os.WriteFile(savePath, trimmed, 0644); err != nil {
		r.LogNotNeededErr = "harness: " + err.Error()
		return
	}
	out := ch11Launch(bin, skillsDir, guiDir, ch11Opts{
		dir:     dir,
		prompts: []string{ch11P4},
		replies: []fakevendor.Reply{ch11Reply("Of course I remember you.")},
	})
	if out.fatal != "" {
		r.LogNotNeededErr = "loading a save with an empty log: " + out.fatal
		return
	}

	// Rule 4: the vendor sees the context, never the log. An empty log
	// beside a real snapshot must still carry the conversation.
	body := ch11LastBody(out)
	if body == "" {
		r.LogNotNeededErr = "no request sent to the vendor"
		return
	}
	if !strings.Contains(body, ch11P1) {
		r.LogNotNeededErr = fmt.Sprintf(
			"rule 4: a save with %q and a non-null context is complete, but the request "+
				"does not carry the earlier prompt %q", `"log": []`, ch11P1)
		return
	}

	// Rule 5: with no log to read a last Seq from, numbering has to
	// continue from the anchor. Every event written this session must
	// sit strictly above it.
	after, err := os.ReadFile(savePath)
	if err != nil {
		r.LogNotNeededErr = fmt.Sprintf("no save file written at exit: %v", err)
		return
	}
	var post ch11SaveTop
	if err := json.Unmarshal(after, &post); err != nil {
		r.LogNotNeededErr = fmt.Sprintf("save written at exit does not parse: %v", err)
		return
	}
	if len(post.Log) == 0 {
		r.LogNotNeededErr = "rule 6: this session created events, but the saved log is empty"
		return
	}
	seqs, err := ch11LogSeqs(post.Log)
	if err != nil {
		r.LogNotNeededErr = "every log event needs a seq: " + err.Error()
		return
	}
	for i, s := range seqs {
		if s <= orig.AsOf {
			r.LogNotNeededErr = fmt.Sprintf(
				"rule 5: the loaded save was anchored at as_of=%d, so the first new event "+
					"must have seq %d; log[%d].seq is %d, which reuses a number already "+
					"inside the snapshot", orig.AsOf, orig.AsOf+1, i, s)
			return
		}
	}
	r.LogNotNeededOK = true
}

// ---------------------------------------------------------------- //
// bad-save-refused (5)
// ---------------------------------------------------------------- //

// Rule 2. The danger is not the crash, it is the silent recovery:
// an agent that shrugs at an unreadable save and starts fresh will
// write an empty conversation over the user's history ten seconds
// later. So the bytes are checked as carefully as the exit code.
func ch11CheckBadSaveRefused(bin, skillsDir, guiDir, tmp string, r *Ch11Result) {
	dir := filepath.Join(tmp, "bad-save")
	if err := os.MkdirAll(dir, 0755); err != nil {
		r.BadSaveRefusedErr = "harness: " + err.Error()
		return
	}
	garbage := []byte("{\"config\": {\"model\": \"fake-model\"}, \"log\": [ this is not JSON\n")
	savePath := filepath.Join(dir, "save.json")
	if err := os.WriteFile(savePath, garbage, 0644); err != nil {
		r.BadSaveRefusedErr = "harness: " + err.Error()
		return
	}

	out := ch11Launch(bin, skillsDir, guiDir, ch11Opts{
		dir:     dir,
		noServe: true,
		replies: []fakevendor.Reply{ch11Reply("should never be asked")},
	})
	if out.fatal != "" {
		r.BadSaveRefusedErr = "rule 2: a save file that does not parse must be fatal, but " + out.fatal
		return
	}
	if out.exitCode == 0 {
		r.BadSaveRefusedErr = "rule 2: agent exited 0 on a save file that does not parse; " +
			"it must exit non-zero rather than start fresh over the user's history"
		return
	}

	after, err := os.ReadFile(savePath)
	if err != nil {
		r.BadSaveRefusedErr = fmt.Sprintf("rule 2: the unreadable save file was removed: %v", err)
		return
	}
	if string(after) != string(garbage) {
		r.BadSaveRefusedErr = "rule 2: the agent exited non-zero but rewrote the save file; " +
			"its bytes must be left untouched so a human can recover them"
		return
	}
	r.BadSaveRefusedOK = true
}

// ---------------------------------------------------------------- //

// ch11Diff reports the first byte at which two request bodies part
// company, with a window either side. Two 8KB JSON blobs that differ in
// one turn are unreadable side by side; the offset is the whole story.
func ch11Diff(nameA, a, nameB, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	lo := i - 60
	if lo < 0 {
		lo = 0
	}
	win := func(s string) string {
		hi := i + 120
		if hi > len(s) {
			hi = len(s)
		}
		return s[lo:hi]
	}
	return fmt.Sprintf("first difference at byte %d (lengths %d and %d)\n  %s: ...%s...\n  %s: ...%s...",
		i, len(a), len(b), nameA, win(a), nameB, win(b))
}

func createCh11Skills(dir string) {
	os.MkdirAll(filepath.Join(dir, "base"), 0755)
	os.WriteFile(filepath.Join(dir, "base", "SKILL.md"), []byte(`---
name: base
description: Base assistant skill
type: primary
tools: read_file think
---

You are a helpful assistant.
`), 0644)
}
