package gradeedition2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter14 grades recorded context management through the actual GUI command.
// The original eight weights total 100; this chapter has no inherited parity row.
func Chapter14(bin string) (grade.Report, error) {
	r := contextChecks(bin)
	checks := grade.Ch15Checks(grade.Ch15Result{Fatal: r.Fatal, Errs: r.Errs})
	checks[4].Title = "Stable primary and fixture tools, dynamic SYSTEM skills, inline MCP invocation and removal"
	return grade.NewTitledReport("Edition 2, Chapter 14 — Context Management", checks, ""), nil
}

// contextResult maps each check to its failure. A check present with an
// empty string passed; a check absent never ran.
type contextResult struct {
	// executable selects the C# host contract for this run only.
	executable bool
	// Fatal reports a launch/setup failure shared by every check.
	Fatal string
	// Errs retains the original eight check outcomes; absent means not run.
	Errs map[string]string
}

func (r *contextResult) set(name, errText string) { r.Errs[name] = errText }

func (r *contextResult) fail(name, format string, a ...any) {
	r.Errs[name] = fmt.Sprintf(format, a...)
}

// The session each check is built on. The scenarios share state on
// purpose: the save written by one session is the input of the next,
// which is exactly how a student's agent will meet its own files.
//
//	ladder   L: load manual, read 12 files at the smallest target
//	restart  A: L again, same target     B: L again, target raised
//	handoff  H: L, read one more file, micro_handoff
//	replay   R1: H as written            R2: H with context nulled
//	crash    C: one clean session, then two turns and SIGKILL, restart
//	total    T: C's final save with two unappliable events planted
//	prefix   F: stub model, load manual then mcp-tools
//	keep     K: stub model, keep one batch and not the next; K' no-stub
func contextChecks(bin string) contextResult { return contextChecksFor(bin, false) }

func contextChecksFor(bin string, executable bool) contextResult {
	r := contextResult{Errs: map[string]string{}, executable: executable}
	tmp, err := os.MkdirTemp("", "context-grade-*")
	if err != nil {
		r.Fatal = err.Error()
		return r
	}
	defer func() {
		if executable {
			failed := r.Fatal != ""
			for _, message := range r.Errs {
				failed = failed || message != ""
			}
			if failed {
				fmt.Fprintln(os.Stderr, "Failed Chapter 14 executable fixtures retained:", tmp)
				return
			}
		}
		os.RemoveAll(tmp)
	}()

	skillsDir := filepath.Join(tmp, "skills")
	if err := contextCreateSkills(skillsDir); err != nil {
		r.Fatal = err.Error()
		return r
	}
	guiDir := filepath.Join(tmp, "gui")
	if err := contextStaticFixture(guiDir); err != nil {
		r.Fatal = fmt.Sprintf("prepare GUI static fixture: %v", err)
		return r
	}

	if executable {
		// The optional keep control is granted by C# startup only when opted in.
		path := filepath.Join(skillsDir, "base", "SKILL.md")
		body, e := os.ReadFile(path)
		if e == nil {
			e = os.WriteFile(path, []byte(strings.Replace(string(body), " keep_tool_results", "", 1)), 0600)
		}
		if e != nil {
			r.Fatal = e.Error()
			return r
		}
	}
	run := func(o contextOpts) contextOut {
		if executable {
			return contextExecutableLaunch(bin, skillsDir, guiDir, o)
		}
		return contextLaunch(bin, skillsDir, guiDir, o)
	}
	newDir := func(name string) string {
		d := filepath.Join(tmp, name)
		os.MkdirAll(d, 0o755)
		return d
	}
	fork := func(from, name string) string {
		d := newDir(name)
		contextCopyDir(from, d)
		return d
	}

	ladder := contextLadder(run, newDir("L"), &r)
	if ladder != "" {
		contextRestart(run, ladder, fork, &r)
		handoff := contextHandoff(run, fork(ladder, "H"), &r)
		if handoff != "" {
			contextReplay(run, handoff, fork, &r)
		} else {
			r.set("replay-equals-snapshot", "the micro_handoff session it replays did not complete")
		}
	} else {
		for _, n := range []string{"ladder-is-recorded", "micro-handoff-shape", "replay-equals-snapshot"} {
			r.set(n, "the ladder session they build on did not complete")
		}
	}
	crash := contextCrash(run, newDir("C"), &r)
	if crash != "" {
		probe := newDir("T")
		if err := contextCopyReducerInput(crash, probe); err != nil {
			r.fail("total-reducer", "prepare isolated primary: %v", err)
		} else {
			contextTotal(run, probe, &r)
		}
	} else {
		r.set("total-reducer", "the crash-recovery session it builds on did not complete")
	}
	if executable {
		contextExecutablePrefix(run, newDir("F"), &r)
	} else {
		contextPrefix(run, newDir("F"), &r)
	}
	contextKeep(run, newDir("K"), newDir("K2"), &r)
	return r
}

// ----------------------------------------------------------------
// skill-survives-the-ladder (L) — rules 3, 4, 7
// ----------------------------------------------------------------

const contextLadderFiles = 12

func contextLadder(run func(contextOpts) contextOut, dir string, r *contextResult) string {
	const name = "skill-survives-the-ladder"
	for n := 1; n <= contextLadderFiles; n++ {
		if err := contextWriteFile(dir, n, 1200); err != nil {
			r.fail(name, "plant file: %v", err)
			return ""
		}
	}

	replies := []fakevendor.Reply{contextCall("L0", "load_skill", `{"name":"manual"}`)}
	for n := 1; n <= contextLadderFiles; n++ {
		replies = append(replies, contextRead(fmt.Sprintf("L%d", n), n))
	}
	replies = append(replies, contextText("All files read."))
	want := len(replies)
	out := run(contextOpts{dir: dir, target: contextSmallTarget, model: contextNoStubModel,
		prompts: []string{"Read the manual, then every file."},
		expect:  []int{want}, replies: replies})
	if out.fatal != "" {
		r.fail(name, "%s", out.fatal)
		return ""
	}
	if len(out.reqs) < want {
		r.fail(name, "expected %d requests, got %d", want, len(out.reqs))
		return ""
	}
	// Every request after the load carries the manual exactly once,
	// in SYSTEM, never duplicated in dialogue or a tool result.
	for i := 1; i < want; i++ {
		if msg := contextSkillOnce(out.reqs[i].Body); msg != "" {
			r.fail(name, "request %d: %s", i+1, msg)
			return dir
		}
	}
	// A missing old marker proves a cut only after the corresponding successful
	// result really reached the model. Inspect results, not echoed call arguments.
	for n := 1; n <= contextLadderFiles; n++ {
		text, failed, err := anthropicToolText(out.reqs[n+1].Body, fmt.Sprintf("L%d", n))
		full, readErr := os.ReadFile(filepath.Join(dir, contextPath(n)))
		if err != nil || readErr != nil || failed || !strings.Contains(text, string(full)) {
			r.fail(name, "read L%d was not delivered in full successfully before trimming: error=%t %v", n, failed, err)
			return dir
		}
	}
	last, _ := contextParse(out.reqs[want-1].Body)
	text := contextMessagesText(last)
	if strings.Contains(text, contextFileMark(1)) {
		r.fail(name, "the ladder never fired: at target %d the first file's result is still in the last request after %d reads",
			contextSmallTarget, contextLadderFiles)
		return dir
	}
	// Both watermarks fired: the oldest read lost its call (RedactTool),
	// and some later read kept its call with the result stubbed
	// (RedactResult).
	lb := contextBlocks(last)
	if contextHasUse(lb, "L1") {
		r.fail(name, "the dialogue watermark never fired: the first read's call is still in the last request")
		return dir
	}
	middle := false
	for n := 2; n < contextLadderFiles; n++ {
		id := fmt.Sprintf("L%d", n)
		if res, ok := contextResultFor(lb, id); ok && contextHasUse(lb, id) && !strings.Contains(res, contextFileMark(n)) {
			middle = true
		}
	}
	if !middle {
		r.fail(name, "no read sits between the watermarks with its call kept and its result stubbed")
		return dir
	}
	if !strings.Contains(text, contextFileMark(contextLadderFiles)) {
		r.fail(name, "the newest file's result is missing from the request that follows it")
		return dir
	}
	if bad := contextPairing(contextBlocks(last)); len(bad) > 0 {
		r.fail(name, "the ladder broke tool pairing: %s", strings.Join(bad, "; "))
		return dir
	}
	r.set(name, "")
	return dir
}

// contextSkillOnce checks durable instruction projection, not tool-result prose.
func contextSkillOnce(body []byte) string {
	req, err := contextParse(body)
	if err != nil {
		return "body is not JSON: " + err.Error()
	}
	return contextInstructions(req, true)
}

// The primary remains stable while a loaded skill intentionally changes SYSTEM.
// Raw marker counts work for both Anthropic string and text-block SYSTEM forms.
func contextInstructions(req contextReq, loaded bool) string {
	if strings.Count(string(req.System), contextPrimaryMark) != 1 {
		return "primary instructions must occur exactly once in SYSTEM"
	}
	want := 0
	if loaded {
		want = 1
	}
	if n := strings.Count(string(req.System), contextManualMark); n != want {
		return fmt.Sprintf("skill body occurs %d times in SYSTEM, want %d", n, want)
	}
	if strings.Contains(contextMessagesText(req), contextManualMark) {
		return "skill body duplicated into dialogue/tool content instead of instruction projection"
	}
	return ""
}

// ----------------------------------------------------------------
// ladder-is-recorded (A, B) — rule 7
// ----------------------------------------------------------------

func contextRestart(run func(contextOpts) contextOut, ladder string, fork func(string, string) string, r *contextResult) {
	const name = "ladder-is-recorded"
	a := fork(ladder, "A")
	b := fork(ladder, "B")

	o := func(dir string) contextOpts {
		return contextOpts{dir: dir, startupManual: r.executable, model: contextNoStubModel, prompts: []string{"Still there?"},
			expect: []int{1}, replies: []fakevendor.Reply{contextText("Yes.")}}
	}
	aOpts, bOpts := o(a), o(b)
	aOpts.target, bOpts.target = contextSmallTarget, 400000
	oa, ob := run(aOpts), run(bOpts)
	for _, x := range []struct {
		n string
		o contextOut
	}{{"same target", oa}, {"raised target", ob}} {
		if x.o.fatal != "" || len(x.o.reqs) == 0 {
			r.fail(name, "restart with %s: %s (requests: %d)", x.n, x.o.fatal, len(x.o.reqs))
			return
		}
	}
	ba, bb := string(oa.reqs[0].Body), string(ob.reqs[0].Body)
	if !strings.Contains(ba, "Read the manual") {
		r.fail(name, "the restart did not load the ladder session")
		return
	}
	if strings.Contains(ba, contextFileMark(1)) {
		r.fail(name, "after restart the first file's result is back: the cuts were not replayed")
		return
	}
	if ba != bb {
		r.fail(name, "raising the target changed history that was already cut: %s",
			contextDiff("same target", ba, "raised target", bb))
		return
	}
	r.set(name, "")
}

// ----------------------------------------------------------------
// micro-handoff-shape (H) — rule 5
// ----------------------------------------------------------------

func contextHandoff(run func(contextOpts) contextOut, dir string, r *contextResult) string {
	const name = "micro-handoff-shape"
	args, _ := json.Marshal(map[string]string{"text": contextHandoffText})
	replies := []fakevendor.Reply{
		contextRead("H1", contextLadderFiles),
		{Tools: []fakevendor.ToolCall{
			{ID: "H2", Name: "micro_handoff", Args: string(args)},
			{ID: "H3", Name: "read_file", Args: contextRead("H3", contextLadderFiles).ToolArgs},
		}},
		contextText("Checkpoint written."),
	}
	out := run(contextOpts{dir: dir, startupManual: r.executable, model: contextNoStubModel, prompts: []string{"Checkpoint now."},
		expect: []int{3}, replies: replies})
	if out.fatal != "" || len(out.reqs) < 3 {
		r.fail(name, "handoff session: %s (requests: %d)", out.fatal, len(out.reqs))
		return ""
	}
	req, err := contextParse(out.reqs[2].Body)
	if err != nil {
		r.fail(name, "request after the handoff is not JSON: %v", err)
		return dir
	}
	blocks := contextBlocks(req)
	if bad := contextPairing(blocks); len(bad) > 0 {
		r.fail(name, "orphaned tool parts after the handoff: %s", strings.Join(bad, "; "))
		return dir
	}
	for _, b := range blocks {
		if b.Type == "tool_use" || b.Type == "tool_result" {
			r.fail(name, "a %s block survived the handoff (id %s%s)", b.Type, b.ID, b.ToolUseID)
			return dir
		}
	}
	if n := strings.Count(contextMessagesText(req), contextHandoffText); n != 1 {
		r.fail(name, "the handoff text appears %d times in the next request, want exactly 1", n)
		return dir
	}
	// The request before the handoff shows the tool traffic it removed,
	// so a pass here is not a session that never had tool parts.
	prev, _ := contextParse(out.reqs[1].Body)
	if !contextHasUse(contextBlocks(prev), "H1") {
		r.fail(name, "the read before the handoff never reached the vendor")
		return dir
	}
	if msg := contextSkillOnce(out.reqs[2].Body); msg != "" {
		r.fail(name, "after the handoff: %s", msg)
		return dir
	}
	if msg := contextHandoffLogFor(dir, r.executable); msg != "" {
		r.fail(name, "%s", msg)
		return dir
	}
	r.set(name, "")
	return dir
}

// ----------------------------------------------------------------
// replay-equals-snapshot (R1, R2) — rules 8, 9
// ----------------------------------------------------------------

func contextReplay(run func(contextOpts) contextOut, handoff string, fork func(string, string) string, r *contextResult) {
	const name = "replay-equals-snapshot"
	data, err := os.ReadFile(filepath.Join(handoff, "save.json"))
	if err != nil {
		r.fail(name, "the handoff session wrote no save.json: %v", err)
		return
	}
	var top struct {
		// Context is the disposable snapshot compared with full-log replay.
		Context json.RawMessage `json:"context"`
		// Log preserves the durable public event sequence.
		Log []json.RawMessage `json:"log"`
	}
	if err := json.Unmarshal(data, &top); err != nil {
		r.fail(name, "save.json does not parse: %v", err)
		return
	}
	if len(top.Context) == 0 || string(top.Context) == "null" {
		r.fail(name, "save.json has no context snapshot to compare against")
		return
	}
	seen := map[string]bool{}
	for _, raw := range top.Log {
		var e struct {
			// Type selects the public wire or event variant under inspection.
			Type string `json:"type"`
		}
		json.Unmarshal(raw, &e)
		seen[e.Type] = true
	}
	handoffKind := "micro_handoff"
	if r.executable {
		handoffKind = "handoff_recorded"
	}
	for _, t := range []string{"skill_loaded", "redacted", handoffKind} {
		if !seen[t] {
			r.fail(name, "the save's log has no %q event; the session should have produced one", t)
			return
		}
	}
	r1 := fork(handoff, "R1")
	r2 := fork(handoff, "R2")
	nulled, err := contextNullSnapshot(data)
	if err != nil {
		r.fail(name, "null the context: %v", err)
		return
	}
	os.WriteFile(filepath.Join(r2, "save.json"), nulled, 0o644)
	o := func(dir string) contextOpts {
		return contextOpts{dir: dir, startupManual: r.executable, model: contextNoStubModel, prompts: []string{"Continue."},
			expect: []int{1}, replies: []fakevendor.Reply{contextText("Continuing.")}}
	}
	o1, o2 := run(o(r1)), run(o(r2))
	if o1.fatal != "" || o2.fatal != "" || len(o1.reqs) == 0 || len(o2.reqs) == 0 {
		r.fail(name, "as written: %q, replayed: %q", o1.fatal, o2.fatal)
		return
	}
	b1, b2 := string(o1.reqs[0].Body), string(o2.reqs[0].Body)
	if !strings.Contains(b1, contextHandoffText) {
		r.fail(name, "loading the save lost the handoff entry")
		return
	}
	if b1 != b2 {
		r.fail(name, "replaying the log differs from the snapshot: %s", contextDiff("snapshot", b1, "replay", b2))
		return
	}
	r.set(name, "")
}

// ----------------------------------------------------------------
// crash-recovery (C) — rule 9
// ----------------------------------------------------------------

const contextCrashLastPrompt = "Do you remember all three?"

var contextCrashPrompts = []string{"First, a clean turn.", "Second turn, before the crash.", "Third turn, then the power goes."}

func contextCrash(run func(contextOpts) contextOut, dir string, r *contextResult) string {
	const name = "crash-recovery"
	// A clean session first, so the crash leaves a tail after an anchor
	// rather than a directory with no snapshot at all.
	o1 := run(contextOpts{dir: dir, model: contextNoStubModel, prompts: contextCrashPrompts[:1],
		expect: []int{1}, replies: []fakevendor.Reply{contextText("REPLY-ONE")}})
	if o1.fatal != "" {
		r.fail(name, "clean session: %s", o1.fatal)
		return ""
	}
	if r.executable {
		// C# snapshots on joined shutdown, not every response. Establish a
		// second clean save so backup fallback has a genuinely older anchor.
		anchor := run(contextOpts{dir: dir, model: contextNoStubModel,
			prompts: []string{"Establish the second recovery anchor."}, expect: []int{1},
			replies: []fakevendor.Reply{contextText("Second anchor complete.")}})
		if anchor.fatal != "" {
			r.fail(name, "second anchor: %s", anchor.fatal)
			return ""
		}
	}
	o2 := run(contextOpts{dir: dir, model: contextNoStubModel, prompts: contextCrashPrompts[1:],
		expect: []int{1, 2}, kill: true,
		replies: []fakevendor.Reply{contextText("REPLY-TWO"), contextText("REPLY-THREE")}})
	if o2.fatal != "" {
		r.fail(name, "session killed mid-way: %s", o2.fatal)
		return ""
	}
	// Clone before any recovery save can rotate the files under comparison.
	primary, err := os.ReadFile(filepath.Join(dir, "save.json"))
	if err != nil {
		r.fail(name, "primary before recovery: %v", err)
		return ""
	}
	backup, err := os.ReadFile(filepath.Join(dir, "save.json.bak"))
	if err != nil || string(primary) == string(backup) {
		r.fail(name, "need a distinct prior backup before fallback recovery: %v", err)
		return ""
	}
	for _, variant := range []string{"missing", "corrupt"} {
		dst := dir + "-" + variant
		if err := contextCopyDir(dir, dst); err != nil {
			r.fail(name, "copy %s: %v", variant, err)
			return ""
		}
		path := filepath.Join(dst, "save.json")
		if variant == "missing" {
			err = os.Remove(path)
		} else {
			err = os.WriteFile(path, []byte("{broken primary"), 0600)
		}
		if err != nil {
			r.fail(name, "%s primary: %v", variant, err)
			return ""
		}
	}
	var expected string
	for _, variant := range []string{"intact", "missing", "corrupt"} {
		d := dir
		if variant != "intact" {
			d += "-" + variant
		}
		o3 := run(contextOpts{dir: d, model: contextNoStubModel, prompts: []string{contextCrashLastPrompt},
			expect: []int{1}, replies: []fakevendor.Reply{contextText("REPLY-FOUR")}})
		if o3.fatal != "" || len(o3.reqs) != 1 {
			r.fail(name, "%s primary recovery: %s (requests %d)", variant, o3.fatal, len(o3.reqs))
			return ""
		}
		body := string(o3.reqs[0].Body)
		if msg := contextCrashContents(body); msg != "" {
			r.fail(name, "%s primary: %s", variant, msg)
			return ""
		}
		if variant == "intact" {
			expected = body
		} else if body != expected {
			r.fail(name, "%s backup-tail replay differs: %s", variant, contextDiff("primary", expected, "backup", body))
			return ""
		}
	}
	r.set(name, "")
	return dir
}

// Counting each turn exactly once rejects both lost and replayed-overlap dialogue.
func contextCrashContents(body string) string {
	for _, want := range append(append([]string{}, contextCrashPrompts...), "REPLY-ONE", "REPLY-TWO", "REPLY-THREE") {
		if n := strings.Count(body, want); n != 1 {
			return fmt.Sprintf("recovered %q %d times, want exactly once", want, n)
		}
	}
	return ""
}

// ----------------------------------------------------------------
// total-reducer (T) — rule 8
// ----------------------------------------------------------------

// Only the complete edited archive may supply this probe's history. A valid
// backup and journal could otherwise hide refusal of the planted known events.
func contextCopyReducerInput(src, dst string) error {
	// The launch closure supplies shared skills; only the edited primary belongs
	// here. Copying recovery side inputs could conceal a strict reducer refusal.
	if err := os.MkdirAll(dst, 0o700); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(src, "save.json"))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, "save.json"), data, 0600)
}

func contextTotal(run func(contextOpts) contextOut, dir string, r *contextResult) {
	const name = "total-reducer"
	path := filepath.Join(dir, "save.json")
	data, err := os.ReadFile(path)
	if err != nil {
		r.fail(name, "the crash-recovery restart wrote no save.json: %v", err)
		return
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		r.fail(name, "save.json does not parse: %v", err)
		return
	}
	var log []map[string]json.RawMessage
	if err := json.Unmarshal(top["log"], &log); err != nil {
		r.fail(name, "save.json log is not a list of events: %v", err)
		return
	}
	// The bad events go in the middle of the log, before the last turn,
	// and the context is nulled so loading replays through them. An
	// agent that stops at the first bad event instead of continuing then
	// loses the turn after it, which rule 8 forbids.
	last := contextCrashLastPrompt
	idx := -1
	for i, ev := range log {
		b, _ := json.Marshal(ev)
		if strings.Contains(string(b), last) {
			idx = i
			break
		}
	}
	if idx < 0 {
		r.fail(name, "save.json log has no event carrying the prompt %q", last)
		return
	}
	var seq0 uint64
	if err := json.Unmarshal(log[idx]["seq"], &seq0); err != nil {
		r.fail(name, "event %d has no numeric seq: %v", idx, err)
		return
	}
	for i := idx; i < len(log); i++ {
		var s uint64
		json.Unmarshal(log[i]["seq"], &s)
		log[i]["seq"], _ = json.Marshal(s + 2)
	}
	// Two events that parse and cannot apply: a redaction of entries
	// that never existed, and a skill load with no payload.
	var bad []map[string]json.RawMessage
	for _, b := range []string{
		fmt.Sprintf(`{"seq":%d,"type":"redacted","time":"2026-01-01T00:00:00Z","redact":{"from":900000,"to":900001,"level":"redact_result","reason":"planted"}}`, seq0),
		fmt.Sprintf(`{"seq":%d,"type":"skill_loaded","time":"2026-01-01T00:00:00Z"}`, seq0+1),
	} {
		var m map[string]json.RawMessage
		json.Unmarshal([]byte(b), &m)
		bad = append(bad, m)
	}
	log = append(log[:idx], append(bad, log[idx:]...)...)
	top["log"], _ = json.Marshal(log)
	top["context"] = json.RawMessage("null")
	out, err := json.Marshal(top)
	if err != nil {
		r.fail(name, "encode planted events: %v", err)
		return
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		r.fail(name, "write planted events: %v", err)
		return
	}

	o := run(contextOpts{dir: dir, model: contextNoStubModel, prompts: []string{"Are you still whole?"},
		expect: []int{1}, replies: []fakevendor.Reply{contextText("Whole.")}})
	if o.fatal != "" || len(o.reqs) == 0 {
		r.fail(name, "with two unappliable events planted the agent did not start and ask: %s", o.fatal)
		return
	}
	req, err := contextParse(o.reqs[0].Body)
	if err != nil || len(req.Messages) == 0 {
		r.fail(name, "the request after loading is malformed: %v", err)
		return
	}
	body := string(o.reqs[0].Body)
	for _, want := range append(append([]string{}, contextCrashPrompts...), contextCrashLastPrompt, "REPLY-FOUR") {
		if !strings.Contains(body, want) {
			r.fail(name, "loading skipped more than the bad events: %q is gone", want)
			return
		}
	}
	r.set(name, "")
}

// ----------------------------------------------------------------
// frozen-prefix (F) — rules 1, 2
// ----------------------------------------------------------------

// ----------------------------------------------------------------
// keep-or-stub (K, K2) — rule 6
// ----------------------------------------------------------------

func contextKeep(run func(contextOpts) contextOut, stubDir, plainDir string, r *contextResult) {
	const name = "keep-or-stub"
	script := func(dir, model string) contextOut {
		for _, n := range []int{21, 22, 23} {
			contextWriteFile(dir, n, 6000)
		}
		replies := []fakevendor.Reply{
			contextRead("K1", 21),
			{Tools: []fakevendor.ToolCall{
				{ID: "K2keep", Name: "keep_tool_results", Args: `{}`},
				{ID: "K2", Name: "read_file", Args: contextRead("K2", 22).ToolArgs},
			}, Usage: fakevendor.Canonical{Input: 100, Output: 30}},
			contextRead("K3", 23),
			contextText("Three files read."),
		}
		return run(contextOpts{dir: dir, model: model, prompts: []string{"Read three files."},
			expect: []int{4}, replies: replies})
	}
	full := func(blocks []contextBlock, id string, n int) (bool, string) {
		res, ok := contextResultFor(blocks, id)
		if !ok {
			return false, "no result for " + id
		}
		return strings.Contains(res, contextFileMark(n)), ""
	}

	out := script(stubDir, contextStubModel)
	if out.fatal != "" || len(out.reqs) < 4 {
		r.fail(name, "%s session: %s (requests: %d)", contextStubModel, out.fatal, len(out.reqs))
		return
	}
	req3, _ := contextParse(out.reqs[2].Body)
	req4, _ := contextParse(out.reqs[3].Body)
	b3, b4 := contextBlocks(req3), contextBlocks(req4)
	if ok, msg := full(b3, "K2", 22); !ok {
		r.fail(name, "the request right after a read must carry its full result: %s", msg)
		return
	}
	if ok, msg := full(b4, "K1", 21); !ok || msg != "" {
		r.fail(name, "a batch kept with keep_tool_results was stubbed anyway %s", msg)
		return
	}
	if ok, msg := full(b4, "K2", 22); ok || msg != "" {
		if msg != "" {
			r.fail(name, "the unkept result lost its tool_result block: %s", msg)
		} else {
			r.fail(name, "on %s the unkept result was not stubbed in the request after the one that carried it", contextStubModel)
		}
		return
	}
	if !contextHasUse(b4, "K2") {
		r.fail(name, "stubbing removed the call; the call must survive")
		return
	}
	if ok, msg := full(b4, "K3", 23); !ok {
		r.fail(name, "the newest result was stubbed before the model saw it %s", msg)
		return
	}

	plain := script(plainDir, contextNoStubModel)
	if plain.fatal != "" || len(plain.reqs) < 4 {
		r.fail(name, "%s session: %s (requests: %d)", contextNoStubModel, plain.fatal, len(plain.reqs))
		return
	}
	p4, _ := contextParse(plain.reqs[3].Body)
	if ok, msg := full(contextBlocks(p4), "K2", 22); !ok {
		r.fail(name, "%s has no stubbing column, but a result was stubbed per round trip %s", contextNoStubModel, msg)
		return
	}
	r.set(name, "")
}

// contextDiff reports differing request sizes without flooding the scored report.
func contextDiff(nameA, a, nameB, b string) string {
	return fmt.Sprintf("%s (%d bytes) differs from %s (%d bytes)", nameA, len(a), nameB, len(b))
}

func contextPrefix(run func(contextOpts) contextOut, dir string, r *contextResult) {
	const name = "frozen-prefix"
	replies := []fakevendor.Reply{
		contextCall("F1", "load_skill", `{"name":"manual"}`),
		contextCall("F2", "load_skill", `{"name":"mcp-tools"}`),
		contextCall("F3", "gui_input", `{"selector":"#search-box","text":"INPUT_ARGUMENT_ONLY"}`),
		contextCall("F4", "unload_skill", `{"name":"mcp-tools"}`),
		contextText("MCP lifecycle complete."),
	}
	out := run(contextOpts{dir: dir, model: contextStubModel, prompts: []string{"Load the manual, use MCP, then unload it."}, expect: []int{5}, replies: replies})
	if out.fatal != "" || len(out.reqs) != 5 {
		r.fail(name, "session: %s (requests %d)", out.fatal, len(out.reqs))
		return
	}
	first, err := contextParse(out.reqs[0].Body)
	if err != nil {
		r.fail(name, "%v", err)
		return
	}
	// Dynamic tools must be absent initially, then present as real declarations.
	var initial []struct {
		// Name identifies the declared public tool.
		Name string `json:"name"`
	}
	if err = json.Unmarshal(first.Tools, &initial); err != nil {
		r.fail(name, "startup tools: %v", err)
		return
	}
	for _, tool := range initial {
		if tool.Name == "list_directory" || strings.HasPrefix(tool.Name, "gui_") {
			r.fail(name, "dynamic %s already declared at startup", tool.Name)
			return
		}
	}
	for i, q := range out.reqs {
		req, e := contextParse(q.Body)
		if e != nil || string(req.Tools) != string(first.Tools) {
			r.fail(name, "request %d changed/malformed startup fixture tools: %v", i+1, e)
			return
		}
		if msg := contextInstructions(req, i > 0); msg != "" {
			r.fail(name, "request %d: %s", i+1, msg)
			return
		}
		added, removed, e := contextInline(req)
		if e != nil {
			r.fail(name, "request %d: %v", i+1, e)
			return
		}
		if i >= 1 && !added["list_directory"] {
			r.fail(name, "request %d lacks manual's inline list_directory declaration", i+1)
			return
		}
		if i >= 2 {
			for _, n := range []string{"gui_input", "gui_click"} {
				if !added[n] {
					r.fail(name, "request %d lacks MCP inline %s declaration", i+1, n)
					return
				}
			}
		}
		if i == 4 {
			for _, n := range []string{"gui_input", "gui_click"} {
				if !removed[n] {
					r.fail(name, "unload lacks inline removal of %s", n)
					return
				}
			}
		}
		// Chapter 12's peer is on-demand. No unsolicited GUI observations
		// belong in SYSTEM or plain messages merely because it was loaded.
		text := guiDebugText(q)
		if strings.Contains(text, "SNAPSHOT_ONLY_") || strings.Contains(text, "TTS_ONLY_") {
			r.fail(name, "request %d contains an unrequested GUI observation", i+1)
			return
		}
	}
	text, failed, e := anthropicToolText(out.reqs[3].Body, "F3")
	if e != nil || failed || !strings.Contains(text, "INPUT_RESULT_ONLY:INPUT_ARGUMENT_ONLY") {
		r.fail(name, "correlated gui_input result missing/wrong: error=%t %v %q", failed, e, text)
		return
	}
	r.set(name, "")
}

// contextStaticFixture satisfies the GUI launch prerequisite only. These
// checks use stdin and /ws; they neither fetch nor execute a page. Inert fixture
// assets are not evidence about the student's browser UI or shipped assets.
func contextStaticFixture(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, f := range []struct{ name, body string }{
		{"index.html", "<!doctype html><title>Context WebSocket fixture</title>\n"},
		{"app.js", "// Inert: Chapter 14 uses the public WebSocket directly.\n"},
		{"style.css", "/* Inert static launch fixture. */\n"},
	} {
		if err := os.WriteFile(filepath.Join(dir, f.name), []byte(f.body), 0o644); err != nil {
			return err
		}
	}
	return nil
}
