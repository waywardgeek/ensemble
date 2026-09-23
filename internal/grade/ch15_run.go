package grade

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Ch15Result maps each check to its failure. A check present with an
// empty string passed; a check absent never ran.
type Ch15Result struct {
	Fatal string
	Errs  map[string]string
}

func (r *Ch15Result) set(name, errText string) { r.Errs[name] = errText }

func (r *Ch15Result) fail(name, format string, a ...any) {
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
func Ch15Run(path string) Ch15Result {
	r := Ch15Result{Errs: map[string]string{}}
	tmp, err := os.MkdirTemp("", "ch15-grade-*")
	if err != nil {
		r.Fatal = err.Error()
		return r
	}
	defer os.RemoveAll(tmp)

	bin, cleanup, err := Build(path)
	if err != nil {
		r.Fatal = "build: " + err.Error()
		return r
	}
	defer cleanup()
	gradeBin, err := os.Executable()
	if err != nil {
		r.Fatal = err.Error()
		return r
	}
	skillsDir := filepath.Join(tmp, "skills")
	if err := ch15CreateSkills(skillsDir, gradeBin); err != nil {
		r.Fatal = err.Error()
		return r
	}
	guiDir := filepath.Join(tmp, "gui")
	os.MkdirAll(guiDir, 0o755)

	run := func(o ch15Opts) ch15Out { return ch15Launch(bin, skillsDir, guiDir, o) }
	newDir := func(name string) string {
		d := filepath.Join(tmp, name)
		os.MkdirAll(d, 0o755)
		return d
	}
	fork := func(from, name string) string {
		d := newDir(name)
		ch15CopyDir(from, d)
		return d
	}

	ladder := ch15Ladder(run, newDir("L"), &r)
	if ladder != "" {
		ch15Restart(run, ladder, fork, &r)
		handoff := ch15Handoff(run, fork(ladder, "H"), &r)
		if handoff != "" {
			ch15Replay(run, handoff, fork, &r)
		} else {
			r.set("replay-equals-snapshot", "the micro_handoff session it replays did not complete")
		}
	} else {
		for _, n := range []string{"ladder-is-recorded", "micro-handoff-shape", "replay-equals-snapshot"} {
			r.set(n, "the ladder session they build on did not complete")
		}
	}
	crash := ch15Crash(run, newDir("C"), &r)
	if crash != "" {
		ch15Total(run, fork(crash, "T"), &r)
	} else {
		r.set("total-reducer", "the crash-recovery session it builds on did not complete")
	}
	ch15Prefix(run, newDir("F"), &r)
	ch15Keep(run, newDir("K"), newDir("K2"), &r)
	return r
}

// ----------------------------------------------------------------
// skill-survives-the-ladder (L) — rules 3, 4, 7
// ----------------------------------------------------------------

const ch15LadderFiles = 12

func ch15Ladder(run func(ch15Opts) ch15Out, dir string, r *Ch15Result) string {
	const name = "skill-survives-the-ladder"
	for n := 1; n <= ch15LadderFiles; n++ {
		if err := ch15WriteFile(dir, n, 1200); err != nil {
			r.fail(name, "plant file: %v", err)
			return ""
		}
	}
	ch15WriteSettings(dir, ch15SmallTarget)
	replies := []fakevendor.Reply{ch15Call("L0", "load_skill", `{"name":"manual"}`)}
	for n := 1; n <= ch15LadderFiles; n++ {
		replies = append(replies, ch15Read(fmt.Sprintf("L%d", n), n))
	}
	replies = append(replies, ch15Text("All files read."))
	want := len(replies)
	out := run(ch15Opts{dir: dir, model: ch15NoStubModel,
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
	// never inside a tool result and never in the system prompt.
	for i := 1; i < want; i++ {
		if msg := ch15SkillOnce(out.reqs[i].Body); msg != "" {
			r.fail(name, "request %d: %s", i+1, msg)
			return dir
		}
	}
	last, _ := ch15Parse(out.reqs[want-1].Body)
	text := ch15MessagesText(last)
	if strings.Contains(text, ch15FileMark(1)) {
		r.fail(name, "the ladder never fired: at target %d the first file's result is still in the last request after %d reads",
			ch15SmallTarget, ch15LadderFiles)
		return dir
	}
	// Both watermarks fired: the oldest read lost its call (RedactTool),
	// and some later read kept its call with the result stubbed
	// (RedactResult).
	lb := ch15Blocks(last)
	if ch15HasUse(lb, "L1") {
		r.fail(name, "the dialogue watermark never fired: the first read's call is still in the last request")
		return dir
	}
	middle := false
	for n := 2; n < ch15LadderFiles; n++ {
		id := fmt.Sprintf("L%d", n)
		if res, ok := ch15ResultFor(lb, id); ok && ch15HasUse(lb, id) && !strings.Contains(res, ch15FileMark(n)) {
			middle = true
		}
	}
	if !middle {
		r.fail(name, "no read sits between the watermarks with its call kept and its result stubbed")
		return dir
	}
	if !strings.Contains(text, ch15FileMark(ch15LadderFiles)) {
		r.fail(name, "the newest file's result is missing from the request that follows it")
		return dir
	}
	if bad := ch15Pairing(ch15Blocks(last)); len(bad) > 0 {
		r.fail(name, "the ladder broke tool pairing: %s", strings.Join(bad, "; "))
		return dir
	}
	r.set(name, "")
	return dir
}

// ch15SkillOnce says what is wrong with the manual's placement in one
// request body, or "" if it appears once, outside tools and system.
func ch15SkillOnce(body []byte) string {
	req, err := ch15Parse(body)
	if err != nil {
		return "body is not JSON: " + err.Error()
	}
	if strings.Contains(string(req.System), ch15ManualMark) {
		return "the skill body is in the system prompt; rule 1 freezes it"
	}
	n := strings.Count(ch15MessagesText(req), ch15ManualMark)
	if n != 1 {
		return fmt.Sprintf("the skill body appears %d times in the messages, want exactly 1", n)
	}
	for _, b := range ch15Blocks(req) {
		if (b.Type == "tool_result" || b.Type == "tool_use") && strings.Contains(string(b.Raw), ch15ManualMark) {
			return "the skill body rides inside a tool part; rule 4 makes it an entry"
		}
	}
	return ""
}

// ----------------------------------------------------------------
// ladder-is-recorded (A, B) — rule 7
// ----------------------------------------------------------------

func ch15Restart(run func(ch15Opts) ch15Out, ladder string, fork func(string, string) string, r *Ch15Result) {
	const name = "ladder-is-recorded"
	a := fork(ladder, "A")
	b := fork(ladder, "B")
	ch15WriteSettings(a, ch15SmallTarget)
	ch15WriteSettings(b, 400000)
	o := func(dir string) ch15Opts {
		return ch15Opts{dir: dir, model: ch15NoStubModel, prompts: []string{"Still there?"},
			expect: []int{1}, replies: []fakevendor.Reply{ch15Text("Yes.")}}
	}
	oa, ob := run(o(a)), run(o(b))
	for _, x := range []struct {
		n string
		o ch15Out
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
	if strings.Contains(ba, ch15FileMark(1)) {
		r.fail(name, "after restart the first file's result is back: the cuts were not replayed")
		return
	}
	if ba != bb {
		r.fail(name, "raising the target changed history that was already cut: %s",
			ch15Diff("same target", ba, "raised target", bb))
		return
	}
	r.set(name, "")
}

// ----------------------------------------------------------------
// micro-handoff-shape (H) — rule 5
// ----------------------------------------------------------------

func ch15Handoff(run func(ch15Opts) ch15Out, dir string, r *Ch15Result) string {
	const name = "micro-handoff-shape"
	args, _ := json.Marshal(map[string]string{"text": ch15HandoffText})
	replies := []fakevendor.Reply{
		ch15Read("H1", ch15LadderFiles),
		ch15Call("H2", "micro_handoff", string(args)),
		ch15Text("Checkpoint written."),
	}
	out := run(ch15Opts{dir: dir, model: ch15NoStubModel, prompts: []string{"Checkpoint now."},
		expect: []int{3}, replies: replies})
	if out.fatal != "" || len(out.reqs) < 3 {
		r.fail(name, "handoff session: %s (requests: %d)", out.fatal, len(out.reqs))
		return ""
	}
	req, err := ch15Parse(out.reqs[2].Body)
	if err != nil {
		r.fail(name, "request after the handoff is not JSON: %v", err)
		return dir
	}
	blocks := ch15Blocks(req)
	if bad := ch15Pairing(blocks); len(bad) > 0 {
		r.fail(name, "orphaned tool parts after the handoff: %s", strings.Join(bad, "; "))
		return dir
	}
	for _, b := range blocks {
		if b.Type == "tool_use" || b.Type == "tool_result" {
			r.fail(name, "a %s block survived the handoff (id %s%s)", b.Type, b.ID, b.ToolUseID)
			return dir
		}
	}
	if n := strings.Count(ch15MessagesText(req), ch15HandoffText); n != 1 {
		r.fail(name, "the handoff text appears %d times in the next request, want exactly 1", n)
		return dir
	}
	// The request before the handoff shows the tool traffic it removed,
	// so a pass here is not a session that never had tool parts.
	prev, _ := ch15Parse(out.reqs[1].Body)
	if !ch15HasUse(ch15Blocks(prev), "H1") {
		r.fail(name, "the read before the handoff never reached the vendor")
		return dir
	}
	if msg := ch15SkillOnce(out.reqs[2].Body); msg != "" {
		r.fail(name, "after the handoff: %s", msg)
		return dir
	}
	r.set(name, "")
	return dir
}

// ----------------------------------------------------------------
// replay-equals-snapshot (R1, R2) — rules 8, 9
// ----------------------------------------------------------------

func ch15Replay(run func(ch15Opts) ch15Out, handoff string, fork func(string, string) string, r *Ch15Result) {
	const name = "replay-equals-snapshot"
	data, err := os.ReadFile(filepath.Join(handoff, "save.json"))
	if err != nil {
		r.fail(name, "the handoff session wrote no save.json: %v", err)
		return
	}
	var top struct {
		Context json.RawMessage   `json:"context"`
		Log     []json.RawMessage `json:"log"`
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
			Type string `json:"type"`
		}
		json.Unmarshal(raw, &e)
		seen[e.Type] = true
	}
	for _, t := range []string{"skill_loaded", "redacted", "micro_handoff"} {
		if !seen[t] {
			r.fail(name, "the save's log has no %q event; the session should have produced one", t)
			return
		}
	}
	r1 := fork(handoff, "R1")
	r2 := fork(handoff, "R2")
	nulled, err := ch11WithNullContext(data)
	if err != nil {
		r.fail(name, "null the context: %v", err)
		return
	}
	os.WriteFile(filepath.Join(r2, "save.json"), nulled, 0o644)
	o := func(dir string) ch15Opts {
		return ch15Opts{dir: dir, model: ch15NoStubModel, prompts: []string{"Continue."},
			expect: []int{1}, replies: []fakevendor.Reply{ch15Text("Continuing.")}}
	}
	o1, o2 := run(o(r1)), run(o(r2))
	if o1.fatal != "" || o2.fatal != "" || len(o1.reqs) == 0 || len(o2.reqs) == 0 {
		r.fail(name, "as written: %q, replayed: %q", o1.fatal, o2.fatal)
		return
	}
	b1, b2 := string(o1.reqs[0].Body), string(o2.reqs[0].Body)
	if !strings.Contains(b1, ch15HandoffText) {
		r.fail(name, "loading the save lost the handoff entry")
		return
	}
	if b1 != b2 {
		r.fail(name, "replaying the log differs from the snapshot: %s", ch15Diff("snapshot", b1, "replay", b2))
		return
	}
	r.set(name, "")
}

// ----------------------------------------------------------------
// crash-recovery (C) — rule 9
// ----------------------------------------------------------------

const ch15CrashLastPrompt = "Do you remember all three?"

var ch15CrashPrompts = []string{"First, a clean turn.", "Second turn, before the crash.", "Third turn, then the power goes."}

func ch15Crash(run func(ch15Opts) ch15Out, dir string, r *Ch15Result) string {
	const name = "crash-recovery"
	// A clean session first, so the crash leaves a tail after an anchor
	// rather than a directory with no snapshot at all.
	o1 := run(ch15Opts{dir: dir, model: ch15NoStubModel, prompts: ch15CrashPrompts[:1],
		expect: []int{1}, replies: []fakevendor.Reply{ch15Text("REPLY-ONE")}})
	if o1.fatal != "" {
		r.fail(name, "clean session: %s", o1.fatal)
		return ""
	}
	o2 := run(ch15Opts{dir: dir, model: ch15NoStubModel, prompts: ch15CrashPrompts[1:],
		expect: []int{1, 2}, kill: true,
		replies: []fakevendor.Reply{ch15Text("REPLY-TWO"), ch15Text("REPLY-THREE")}})
	if o2.fatal != "" {
		r.fail(name, "session killed mid-way: %s", o2.fatal)
		return ""
	}
	o3 := run(ch15Opts{dir: dir, model: ch15NoStubModel, prompts: []string{ch15CrashLastPrompt},
		expect: []int{1}, replies: []fakevendor.Reply{ch15Text("REPLY-FOUR")}})
	if o3.fatal != "" || len(o3.reqs) == 0 {
		r.fail(name, "restart after SIGKILL: %s (requests: %d)", o3.fatal, len(o3.reqs))
		return ""
	}
	body := string(o3.reqs[0].Body)
	for i, want := range append(append([]string{}, ch15CrashPrompts...), "REPLY-ONE", "REPLY-TWO", "REPLY-THREE") {
		if !strings.Contains(body, want) {
			what := "turn"
			if i >= len(ch15CrashPrompts) {
				what = "reply"
			}
			r.fail(name, "after SIGKILL and restart the %s %q is gone", what, want)
			return dir
		}
	}
	r.set(name, "")
	return dir
}

// ----------------------------------------------------------------
// total-reducer (T) — rule 8
// ----------------------------------------------------------------

func ch15Total(run func(ch15Opts) ch15Out, dir string, r *Ch15Result) {
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
	last := ch15CrashLastPrompt
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
	out, _ := json.Marshal(top)
	os.WriteFile(path, out, 0o644)

	o := run(ch15Opts{dir: dir, model: ch15NoStubModel, prompts: []string{"Are you still whole?"},
		expect: []int{1}, replies: []fakevendor.Reply{ch15Text("Whole.")}})
	if o.fatal != "" || len(o.reqs) == 0 {
		r.fail(name, "with two unappliable events planted the agent did not start and ask: %s", o.fatal)
		return
	}
	req, err := ch15Parse(o.reqs[0].Body)
	if err != nil || len(req.Messages) == 0 {
		r.fail(name, "the request after loading is malformed: %v", err)
		return
	}
	body := string(o.reqs[0].Body)
	for _, want := range append(append([]string{}, ch15CrashPrompts...), ch15CrashLastPrompt, "REPLY-FOUR") {
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

func ch15Prefix(run func(ch15Opts) ch15Out, dir string, r *Ch15Result) {
	const name = "frozen-prefix"
	replies := []fakevendor.Reply{
		ch15Call("F1", "load_skill", `{"name":"manual"}`),
		ch15Call("F2", "load_skill", `{"name":"mcp-tools"}`),
		ch15Text("Both loaded."),
	}
	out := run(ch15Opts{dir: dir, model: ch15StubModel, prompts: []string{"Load the manual and the MCP tools."},
		expect: []int{3}, replies: replies})
	if out.fatal != "" || len(out.reqs) < 3 {
		r.fail(name, "session: %s (requests: %d)", out.fatal, len(out.reqs))
		return
	}
	first, err := ch15Parse(out.reqs[0].Body)
	if err != nil {
		r.fail(name, "first request is not JSON: %v", err)
		return
	}
	for i, q := range out.reqs[1:] {
		req, err := ch15Parse(q.Body)
		if err != nil {
			r.fail(name, "request %d is not JSON: %v", i+2, err)
			return
		}
		if string(req.System) != string(first.System) {
			r.fail(name, "the system prompt changed at request %d: %s", i+2,
				ch15Diff("request 1", string(first.System), fmt.Sprintf("request %d", i+2), string(req.System)))
			return
		}
		if string(req.Tools) != string(first.Tools) {
			r.fail(name, "the startup tool declarations changed at request %d: %s", i+2,
				ch15Diff("request 1", string(first.Tools), fmt.Sprintf("request %d", i+2), string(req.Tools)))
			return
		}
	}
	last, _ := ch15Parse(out.reqs[2].Body)
	msgs := ch15MessagesText(last)
	for _, tool := range []string{"think", ch15MCPTool} {
		quoted := fmt.Sprintf(`"name":"%s"`, tool)
		if strings.Contains(string(first.Tools), quoted) {
			r.fail(name, "%s was declared at startup; the fixture expects it only after a load", tool)
			return
		}
		if !strings.Contains(strings.ReplaceAll(msgs, `\"`, `"`), quoted) {
			r.fail(name, "the tool %s never reached the dialog after its skill loaded", tool)
			return
		}
	}
	r.set(name, "")
}

// ----------------------------------------------------------------
// keep-or-stub (K, K2) — rule 6
// ----------------------------------------------------------------

func ch15Keep(run func(ch15Opts) ch15Out, stubDir, plainDir string, r *Ch15Result) {
	const name = "keep-or-stub"
	script := func(dir, model string) ch15Out {
		for _, n := range []int{21, 22, 23} {
			ch15WriteFile(dir, n, 6000)
		}
		replies := []fakevendor.Reply{
			ch15Read("K1", 21),
			{Tools: []fakevendor.ToolCall{
				{ID: "K2keep", Name: "keep_tool_results", Args: `{}`},
				{ID: "K2", Name: "read_file", Args: ch15Read("K2", 22).ToolArgs},
			}, Usage: fakevendor.Canonical{Input: 100, Output: 30}},
			ch15Read("K3", 23),
			ch15Text("Three files read."),
		}
		return run(ch15Opts{dir: dir, model: model, prompts: []string{"Read three files."},
			expect: []int{4}, replies: replies})
	}
	full := func(blocks []ch15Block, id string, n int) (bool, string) {
		res, ok := ch15ResultFor(blocks, id)
		if !ok {
			return false, "no result for " + id
		}
		return strings.Contains(res, ch15FileMark(n)), ""
	}

	out := script(stubDir, ch15StubModel)
	if out.fatal != "" || len(out.reqs) < 4 {
		r.fail(name, "%s session: %s (requests: %d)", ch15StubModel, out.fatal, len(out.reqs))
		return
	}
	req3, _ := ch15Parse(out.reqs[2].Body)
	req4, _ := ch15Parse(out.reqs[3].Body)
	b3, b4 := ch15Blocks(req3), ch15Blocks(req4)
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
			r.fail(name, "on %s the unkept result was not stubbed in the request after the one that carried it", ch15StubModel)
		}
		return
	}
	if !ch15HasUse(b4, "K2") {
		r.fail(name, "stubbing removed the call; the call must survive")
		return
	}
	if ok, msg := full(b4, "K3", 23); !ok {
		r.fail(name, "the newest result was stubbed before the model saw it %s", msg)
		return
	}

	plain := script(plainDir, ch15NoStubModel)
	if plain.fatal != "" || len(plain.reqs) < 4 {
		r.fail(name, "%s session: %s (requests: %d)", ch15NoStubModel, plain.fatal, len(plain.reqs))
		return
	}
	p4, _ := ch15Parse(plain.reqs[3].Body)
	if ok, msg := full(ch15Blocks(p4), "K2", 22); !ok {
		r.fail(name, "%s has no stubbing column, but a result was stubbed per round trip %s", ch15NoStubModel, msg)
		return
	}
	r.set(name, "")
}

// ch15Diff keeps failure messages short: ch11's diff centred on the
// first differing byte.
func ch15Diff(nameA, a, nameB, b string) string { return ch11Diff(nameA, a, nameB, b) }
