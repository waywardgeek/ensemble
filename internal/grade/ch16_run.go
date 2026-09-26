package grade

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Ch16Result carries one entry per check that ran. An empty string means
// the check passed; a missing key means it never ran, which is a failure
// of the harness rather than of the student.
type Ch16Result struct {
	Fatal string
	Errs  map[string]string
}

func (r *Ch16Result) ran(ids ...string) {
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	for _, id := range ids {
		if _, seen := r.Errs[id]; !seen {
			r.Errs[id] = ""
		}
	}
}

func (r *Ch16Result) fail(id, format string, a ...any) {
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	if msg, seen := r.Errs[id]; seen && msg != "" {
		return // keep the first, most specific failure
	}
	r.Errs[id] = fmt.Sprintf(format, a...)
}

// Settings that make the ladder observable in a few turns. The budgets are
// tiny on purpose: the point is to cross thresholds with scripted prompts
// rather than to wait for a real session to grow.
const ch16Settings = `{"context_target":200000,` +
	`"memory":{"conversation":{"budget":300},"session":{"budget":200},` +
	`"8x":{"budget":200},"64x":{"budget":200}}}`

// ch16SettingsOff disables the session band and the one above it.
const ch16SettingsNoSession = `{"context_target":200000,` +
	`"memory":{"conversation":{"budget":300},"session":{"budget":200,"disabled":true},` +
	`"8x":{"budget":200,"disabled":true},"64x":{"budget":200}}}`

const ch16Settings8xOnly = `{"context_target":200000,` +
	`"memory":{"conversation":{"budget":300},"session":{"budget":200,"disabled":true},` +
	`"8x":{"budget":200},"64x":{"budget":200}}}`

// A compressor that refuses to hand anything to the band above it.
const ch16SettingsNo8x = `{"context_target":200000,` +
	`"memory":{"conversation":{"budget":300},"session":{"budget":200},` +
	`"8x":{"budget":200,"disabled":true},"64x":{"budget":200}}}`

// ch16Memo is the text a scripted compressor submits for the nth memory.
// The number makes it possible to say which memories were folded and in
// what order without reading the student's directory.
func ch16Memo(n int) string {
	return fmt.Sprintf("MEMO-%02d I finished a piece of work and wrote it down.", n)
}

func ch16Fold(label string) string {
	return fmt.Sprintf("FOLD-%s I combined several memories into one.", label)
}

// ch16Router answers compressor requests. It hands out numbered memories
// in the order it is asked, and records every compressor body it saw.
type ch16Router struct {
	mu     sync.Mutex
	n      int
	bodies [][]byte
	// foldAt, when non-zero, means: from this call onward, answer with a
	// fold label instead of a session memory. Graduation compressors ask
	// later than conversation compressors, so this is how the grader
	// tells the two apart without reading the student's events.
	folds int
	// fail, when true, answers every compressor with a 500.
	fail bool
}

func (r *ch16Router) route(body []byte) *fakevendor.Reply {
	if !ch16IsCompressor(body) {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bodies = append(r.bodies, append([]byte(nil), body...))
	if r.fail {
		return &fakevendor.Reply{Status: 500, ErrBody: `{"error":"compressor unavailable"}`}
	}

	// A graduation compressor is recognisable: it is folding memories,
	// so what it was given already contains numbered memos.
	text := string(body)
	if strings.Count(text, "MEMO-") >= 4 {
		r.folds++
		return &fakevendor.Reply{
			ToolName: "submit",
			ToolID:   fmt.Sprintf("fold-%d", r.folds),
			ToolArgs: mustJSONArgs(ch16Fold(fmt.Sprintf("%02d", r.folds))),
			Usage:    fakevendor.Canonical{Input: 100, Output: 20},
		}
	}
	r.n++
	return &fakevendor.Reply{
		ToolName: "submit",
		ToolID:   fmt.Sprintf("memo-%d", r.n),
		ToolArgs: mustJSONArgs(ch16Memo(r.n)),
		Usage:    fakevendor.Canonical{Input: 100, Output: 20},
	}
}

func (r *ch16Router) seen() [][]byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.bodies
}

func mustJSONArgs(memory string) string {
	b, _ := json.Marshal(map[string]string{"memory": memory})
	return string(b)
}

// ch16Talk builds the script for a conversation of n turns. Each turn is
// a micro_handoff call followed by a short reply, which closes a segment
// and gives the compressor something whole to fold.
func ch16Talk(n int) []fakevendor.Reply {
	var out []fakevendor.Reply
	for i := 0; i < n; i++ {
		out = append(out,
			fakevendor.Reply{
				ToolName: "micro_handoff",
				ToolID:   fmt.Sprintf("mh-%d", i),
				ToolArgs: fmt.Sprintf(`{"text":"Checkpoint %d. I finished step %d and am moving on to the next piece of work."}`, i, i),
				Usage:    fakevendor.Canonical{Input: 200, Output: 30},
			},
			fakevendor.Reply{
				Text:  fmt.Sprintf("Step %d is recorded.", i),
				Usage: fakevendor.Canonical{Input: 220, Output: 20},
			},
		)
	}
	return out
}

// ch16Prompt is long enough that a few of them cross a 300 byte budget.
func ch16Prompt(i int, plant string) string {
	return fmt.Sprintf(
		"Work item %d. %s Please carry on with the task, keeping notes as you go, "+
			"because there is a lot of detail here and I want it all remembered properly.",
		i, plant)
}

const ch16Plant = "ZORBLAX-7741"

// Ch16Run drives the scenarios and returns one entry per check.
func Ch16Run(path string) Ch16Result {
	var res Ch16Result
	res.ran(ch16IDs()...)

	bin, cleanup, err := Build(path)
	if err != nil {
		res.Fatal = err.Error()
		return res
	}
	defer cleanup()

	skills, err := ch16Skills()
	if err != nil {
		res.Fatal = err.Error()
		return res
	}
	defer os.RemoveAll(skills)

	gui := filepath.Join(path, "web")

	ch16Compress(bin, skills, gui, &res)
	ch16Graduate(bin, skills, gui, &res)
	ch16Restore(bin, skills, gui, &res)
	ch16Force(bin, skills, gui, &res)
	ch16Replay(bin, skills, gui, &res)
	ch16Survive(bin, skills, gui, &res)
	ch16Neighbor(bin, skills, gui, &res)
	ch16Fresh(bin, skills, gui, &res)
	ch16Parity(path, &res)

	return res
}

func ch16IDs() []string {
	out := make([]string, 0, len(ch16Table))
	for _, c := range ch16Table {
		out = append(out, c.ID)
	}
	return out
}

// ----------------------------------------------------------------
// Scenario 7: a band will not graduate into a band that is switched off
// ----------------------------------------------------------------

func ch16Neighbor(bin, skills, gui string, res *Ch16Result) {
	dir, cleanup := freshRunDir("ch16-neighbor")
	defer cleanup()
	r := &ch16Router{}

	n := 16
	var prompts []string
	var expect []int
	for i := 0; i < n; i++ {
		prompts = append(prompts, ch16Prompt(i, ""))
		expect = append(expect, 2*(i+1))
	}
	out := ch16Launch(bin, skills, gui, ch16Opts{
		dir: dir, model: "claude-sonnet-5-course", prompts: prompts, expect: expect,
		replies: ch16Talk(n + 2), route: r.route, settings: ch16SettingsNo8x,
	})
	if out.fatal != "" {
		res.fail("disabled-neighbor-refused", "run did not complete: %s", out.fatal)
		return
	}
	if r.folds > 0 {
		res.fail("disabled-neighbor-refused",
			"the 8x band is switched off and the session band folded into it anyway (%d folds). "+
				"Graduating into a band nobody is reading destroys the memories.", r.folds)
		return
	}
	// Refusing must mean keeping, not discarding.
	last, ok := ch16Last(ch16Turns(out.reqs))
	if !ok {
		res.fail("disabled-neighbor-refused", "no turns recorded")
		return
	}
	if r.n == 0 {
		res.fail("disabled-neighbor-refused",
			"no session memories were written at all, so refusing to graduate proves nothing. "+
				"This scenario needs the session band to fill up first.")
		return
	}
	if !strings.Contains(ch16Whole(last.Body), ch16Memo(1)) {
		res.fail("disabled-neighbor-refused",
			"the 8x band is off, the session band did not fold, and the oldest memory vanished anyway. "+
				"Refusing to graduate has to mean keeping the memory, not dropping it.")
	}
}

// ----------------------------------------------------------------
// Scenario 8: memory files on disk are in context before the first turn
// ----------------------------------------------------------------

func ch16Fresh(bin, skills, gui string, res *Ch16Result) {
	dir, cleanup := freshRunDir("ch16-fresh")
	defer cleanup()
	r := &ch16Router{}

	// Build some memories, then throw away everything except the memory
	// directory. A fresh agent with memory files and no log must still
	// come up remembering.
	var prompts []string
	var expect []int
	for i := 0; i < 6; i++ {
		prompts = append(prompts, ch16Prompt(i, ""))
		expect = append(expect, 2*(i+1))
	}
	first := ch16Launch(bin, skills, gui, ch16Opts{
		dir: dir, model: "claude-sonnet-5-course", prompts: prompts, expect: expect,
		replies: ch16Talk(8), route: r.route, settings: ch16Settings,
	})
	if first.fatal != "" {
		res.fail("fresh-start-populates", "first run did not complete: %s", first.fatal)
		return
	}
	if r.n == 0 {
		res.fail("fresh-start-populates", "no memories were written, so there is nothing to come back to")
		return
	}

	// Remove the save file and the event log; keep whatever the student
	// wrote to disk as memory.
	for _, name := range []string{"save.json", "events.jsonl"} {
		_ = os.Remove(filepath.Join(dir, name))
	}

	r2 := &ch16Router{n: r.n}
	second := ch16Launch(bin, skills, gui, ch16Opts{
		dir: dir, model: "claude-sonnet-5-course", prompts: []string{"What were we doing?"}, expect: []int{1},
		replies: ch16Talk(2), route: r2.route, settings: ch16Settings,
	})
	if second.fatal != "" {
		res.fail("fresh-start-populates", "fresh start did not complete: %s", second.fatal)
		return
	}
	last, ok := ch16Last(ch16Turns(second.reqs))
	if !ok {
		res.fail("fresh-start-populates", "no turns recorded on the fresh start")
		return
	}
	if !strings.Contains(ch16Whole(last.Body), ch16Memo(1)) {
		res.fail("fresh-start-populates",
			"with the log deleted and the memory files left in place, the very first request "+
				"contained none of the memories. The files on disk are what the bands are made of: "+
				"an agent that only remembers by replaying its log has no memory, it has a transcript.")
	}
}

// ch16Skills writes the minimal skills tree the agent needs to boot.
func ch16Skills() (string, error) {
	dir, err := os.MkdirTemp("", "ch16-skills-")
	if err != nil {
		return "", err
	}
	d := filepath.Join(dir, "base")
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	body := `---
name: base
description: Base agent skill for the memory checks
type: primary
tools: read_file keep_tool_results micro_handoff
---

You are a helpful agent.
`
	if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(body), 0o644); err != nil {
		return "", err
	}
	return dir, nil
}

// ----------------------------------------------------------------
// Scenario 1: a checkpoint turns finished work into a memory
// ----------------------------------------------------------------

func ch16Compress(bin, skills, gui string, res *Ch16Result) {
	dir, cleanup := freshRunDir("ch16-compress")
	defer cleanup()
	r := &ch16Router{}

	var prompts []string
	for i := 0; i < 6; i++ {
		plant := ""
		if i == 0 {
			plant = ch16Plant
		}
		prompts = append(prompts, ch16Prompt(i, plant))
	}

	out := ch16Launch(bin, skills, gui, ch16Opts{
		dir:      dir,
		model:    "claude-sonnet-5-course",
		prompts:  prompts,
		expect:   []int{2, 4, 6, 8, 10, 12},
		replies:  ch16Talk(8),
		route:    r.route,
		settings: ch16Settings,
	})
	if out.fatal != "" {
		for _, id := range []string{"micro-handoff-compresses", "planted-fact-survives", "memory-is-data"} {
			res.fail(id, "run did not complete: %s", out.fatal)
		}
		return
	}

	comps := ch16Compressors(out.reqs)
	if len(comps) == 0 {
		res.fail("micro-handoff-compresses",
			"the conversation grew past its budget across %d turns and no compressor ever ran. "+
				"A request offering exactly one tool named submit is what this grader recognises as a compressor.",
			len(ch16Turns(out.reqs)))
		res.fail("planted-fact-survives", "no compressor ran, so nothing could be remembered")
		res.fail("memory-is-data", "no compressor ran, so no memory reached the context")
		return
	}

	// The compressor must have been shown the work it is folding.
	sawPlant := false
	for _, c := range comps {
		if strings.Contains(string(c.Body), ch16Plant) {
			sawPlant = true
			break
		}
	}
	if !sawPlant {
		res.fail("planted-fact-survives",
			"a compressor ran but was never shown the text it was folding: none of the %d compressor "+
				"requests contained %q, which was in the first prompt", len(comps), ch16Plant)
	}

	turns := ch16Turns(out.reqs)
	last, ok := ch16Last(turns)
	if !ok {
		res.fail("micro-handoff-compresses", "no turn requests recorded")
		return
	}

	// The memory the compressor submitted has to come back as context.
	if !strings.Contains(ch16Whole(last.Body), ch16Memo(1)) {
		res.fail("planted-fact-survives",
			"the compressor submitted a memory but it never came back: the last turn does not contain %q",
			ch16Memo(1))
	}

	// And the raw work it replaced has to be gone.
	if strings.Contains(ch16Conversation(last.Body), ch16Plant) {
		res.fail("micro-handoff-compresses",
			"the first prompt was compressed into a memory but its original text is still in the "+
				"conversation: %q appears in the last turn. Compaction has to replace the span, not annotate it.",
			ch16Plant)
	}

	// Memory arrives as something the agent is told, not as an instruction
	// bolted onto its constitution.
	if strings.Contains(ch16System(last.Body), ch16Memo(1)) {
		res.fail("memory-is-data",
			"the memory was placed in the system prompt. Memory is data in the conversation: "+
				"putting it in the system prompt makes every new memory a cache miss on the whole prefix, "+
				"and turns a recollection into an instruction.")
	}

	// A compressor writing in the third person produces memories that read
	// as a report about a stranger.
	firstPerson := false
	for _, c := range comps {
		s := strings.ToLower(ch16System(c.Body))
		if strings.Contains(s, "first person") || strings.Contains(s, "first-person") {
			firstPerson = true
			break
		}
	}
	if !firstPerson {
		res.fail("memory-is-data",
			"the compressor was never asked to write in the first person. A memory written about the "+
				"agent rather than by it reads as a briefing on somebody else.")
	}
}

// ----------------------------------------------------------------
// Scenario 2: memories fold upward, oldest first
// ----------------------------------------------------------------

func ch16Graduate(bin, skills, gui string, res *Ch16Result) {
	dir, cleanup := freshRunDir("ch16-graduate")
	defer cleanup()
	r := &ch16Router{}

	n := 16
	var prompts []string
	var expect []int
	for i := 0; i < n; i++ {
		prompts = append(prompts, ch16Prompt(i, ""))
		expect = append(expect, 2*(i+1))
	}

	out := ch16Launch(bin, skills, gui, ch16Opts{
		dir:      dir,
		model:    "claude-sonnet-5-course",
		prompts:  prompts,
		expect:   expect,
		replies:  ch16Talk(n + 2),
		route:    r.route,
		settings: ch16Settings,
	})
	if out.fatal != "" {
		res.fail("graduation-fires-oldest-first", "run did not complete: %s", out.fatal)
		return
	}

	if r.folds == 0 {
		res.fail("graduation-fires-oldest-first",
			"%d session memories were written and the session band's budget is 200 bytes, "+
				"but no compressor was ever asked to fold memories into the band above. "+
				"A band that only ever grows is not a ladder.", r.n)
		return
	}

	turns := ch16Turns(out.reqs)
	last, ok := ch16Last(turns)
	if !ok {
		res.fail("graduation-fires-oldest-first", "no turn requests recorded")
		return
	}
	whole := ch16Whole(last.Body)

	if !strings.Contains(whole, ch16Fold("01")) {
		res.fail("graduation-fires-oldest-first",
			"a fold happened but its result never reached the context: %q is missing from the last turn",
			ch16Fold("01"))
	}
	// The oldest memory must have been consumed by the fold.
	if strings.Contains(whole, ch16Memo(1)) {
		res.fail("graduation-fires-oldest-first",
			"memory 1 is still present after a fold. Graduation takes the OLDEST memories, "+
				"and leaving them behind means the band never shrinks.")
	}
	// And the newest must still be there, uncompressed.
	newest := ch16Memo(r.n)
	if r.n > 0 && !strings.Contains(whole, newest) {
		res.fail("graduation-fires-oldest-first",
			"the newest memory %q is gone. A fold should take the oldest memories, not the whole band.",
			newest)
	}
}

// ----------------------------------------------------------------
// Scenario 3: switching bands off and on
// ----------------------------------------------------------------

func ch16Restore(bin, skills, gui string, res *Ch16Result) {
	dir, cleanup := freshRunDir("ch16-restore")
	defer cleanup()
	r := &ch16Router{}

	run := func(settings string, prompts []string, expect []int) (ch16Out, bool) {
		o := ch16Launch(bin, skills, gui, ch16Opts{
			dir:      dir,
			model:    "claude-sonnet-5-course",
			prompts:  prompts,
			expect:   expect,
			replies:  ch16Talk(len(prompts) + 2),
			route:    r.route,
			settings: settings,
		})
		return o, o.fatal == ""
	}

	// Build up some memories.
	var prompts []string
	var expect []int
	for i := 0; i < 8; i++ {
		prompts = append(prompts, ch16Prompt(i, ""))
		expect = append(expect, 2*(i+1))
	}
	first, ok := run(ch16Settings, prompts, expect)
	if !ok {
		res.fail("disable-enable-idempotent", "first run did not complete: %s", first.fatal)
		res.fail("disabled-neighbor-refused", "first run did not complete: %s", first.fatal)
		return
	}
	baseTurns := ch16Turns(first.reqs)
	baseLast, _ := ch16Last(baseTurns)
	want := ch16Conversation(baseLast.Body)

	if r.n == 0 {
		res.fail("disable-enable-idempotent", "no memories were produced, so there is nothing to switch off")
		return
	}

	// Off. The memories must leave the context.
	off, ok := run(ch16SettingsNoSession, []string{"Carry on."}, []int{1})
	if !ok {
		res.fail("disable-enable-idempotent", "run with the session band off did not complete: %s", off.fatal)
		return
	}
	offLast, _ := ch16Last(ch16Turns(off.reqs))
	if strings.Contains(ch16Conversation(offLast.Body), ch16Memo(1)) {
		res.fail("disable-enable-idempotent",
			"the session band was switched off but its memories are still in the context. "+
				"A band that cannot be switched off is not a setting.")
		return
	}

	// While the band is off, edit a memory on disk. Re-enabling has to
	// re-read the files rather than restore a snapshot it kept in the log:
	// the files are the memory, and editing one is how a person corrects
	// something the agent got wrong.
	edited := ch16MarkLiveMemory(dir, want)

	// Back on, but the upper band first, so the restore order differs from
	// the order they were switched off in.
	mid, ok := run(ch16Settings8xOnly, []string{"Carry on."}, []int{1})
	if !ok {
		res.fail("disable-enable-idempotent", "partial restore did not complete: %s", mid.fatal)
		return
	}
	back, ok := run(ch16Settings, []string{"Carry on."}, []int{1})
	if !ok {
		res.fail("disable-enable-idempotent", "restore did not complete: %s", back.fatal)
		return
	}
	backLast, _ := ch16Last(ch16Turns(back.reqs))
	got := ch16Conversation(backLast.Body)

	if edited && !strings.Contains(got, "CORRECTED-BY-HAND") {
		res.fail("disable-enable-idempotent",
			"a memory file was edited on disk while its band was switched off, and switching the band "+
				"back on restored the old text. Re-enabling has to re-read the files. Keeping a copy in "+
				"the log and replaying that makes the files on disk decorative.")
		return
	}

	wantMem := ch16OnlyMemories(want)
	gotMem := ch16OnlyMemories(got)
	if edited {
		// The edited memo no longer matches its original text, so compare
		// only the memories that were not touched.
		wantMem = ch16DropFirst(wantMem)
		gotMem = ch16DropFirst(gotMem)
	}
	if wantMem != gotMem {
		res.fail("disable-enable-idempotent",
			"switching the bands off and back on - in a different order - did not restore the same memory.\n"+
				"  before: %s\n  after:  %s\n"+
				"The files on disk have not changed, so what is loaded from them must not depend on the order "+
				"the bands happened to come back in.", wantMem, gotMem)
	}
}

// ch16OnlyMemories reduces a rendered conversation to the memory markers
// it contains, in order. Comparing whole bodies would fail on the new
// turns each run adds; what has to be stable is which memories are
// present and in what order.
func ch16OnlyMemories(s string) string {
	var found []string
	for i := 1; i <= 40; i++ {
		m := ch16Memo(i)
		if idx := strings.Index(s, m); idx >= 0 {
			found = append(found, fmt.Sprintf("%d@%d", i, idx))
		}
	}
	for i := 1; i <= 10; i++ {
		f := ch16Fold(fmt.Sprintf("%02d", i))
		if idx := strings.Index(s, f); idx >= 0 {
			found = append(found, fmt.Sprintf("F%d@%d", i, idx))
		}
	}
	// Positions differ between runs; keep only the order.
	var names []string
	for _, f := range found {
		names = append(names, strings.Split(f, "@")[0])
	}
	return strings.Join(names, ",")
}

// ----------------------------------------------------------------
// Scenario 4: a full context takes the tools away
// ----------------------------------------------------------------

func ch16Force(bin, skills, gui string, res *Ch16Result) {
	dir, cleanup := freshRunDir("ch16-force")
	defer cleanup()
	r := &ch16Router{}

	// claude-ch16-course has a 4096 token window. Report usage that fills
	// it, and the agent should stop offering anything but the checkpoint.
	big := fakevendor.Canonical{Input: 3900, Output: 20}
	replies := []fakevendor.Reply{
		{Text: "Working.", Usage: big},
		{Text: "Still working.", Usage: big},
		{Text: "Still working.", Usage: big},
	}

	out := ch16Launch(bin, skills, gui, ch16Opts{
		dir:      dir,
		model:    "claude-ch16-course",
		prompts:  []string{"One.", "Two.", "Three."},
		expect:   []int{1, 2, 3},
		replies:  replies,
		route:    r.route,
		settings: ch16Settings,
	})
	if out.fatal != "" {
		res.fail("forced-handoff", "run did not complete: %s", out.fatal)
		return
	}

	turns := ch16Turns(out.reqs)
	if len(turns) < 2 {
		res.fail("forced-handoff", "expected at least 2 turns, saw %d", len(turns))
		return
	}
	last := turns[len(turns)-1]
	names := ch16ToolNames(last.Body)
	if len(names) == 0 {
		var all []string
		for i, t := range turns {
			all = append(all, fmt.Sprintf("req%d=%v", i+1, ch16ToolNames(t.Body)))
		}
		res.fail("forced-handoff",
			"the last request offered no tools at all; micro_handoff must remain. Tools per turn: %s",
			strings.Join(all, " "))
		return
	}
	hasHandoff := false
	extra := []string{}
	for _, n := range names {
		if n == "micro_handoff" {
			hasHandoff = true
		} else {
			extra = append(extra, n)
		}
	}
	if !hasHandoff {
		res.fail("forced-handoff",
			"the context was reported 95%% full and micro_handoff was not offered. "+
				"Taking away the only tool that can fix the problem leaves the agent stuck.")
	}
	if len(extra) > 0 {
		res.fail("forced-handoff",
			"the context was reported 95%% full but the agent was still offered %v. "+
				"Asking nicely does not work: the way to force a checkpoint is to make it the only move.",
			extra)
	}
}

// ----------------------------------------------------------------
// Scenario 5: replay costs nothing
// ----------------------------------------------------------------

func ch16Replay(bin, skills, gui string, res *Ch16Result) {
	dir, cleanup := freshRunDir("ch16-replay")
	defer cleanup()
	r := &ch16Router{}

	var prompts []string
	var expect []int
	for i := 0; i < 6; i++ {
		prompts = append(prompts, ch16Prompt(i, ""))
		expect = append(expect, 2*(i+1))
	}
	first := ch16Launch(bin, skills, gui, ch16Opts{
		dir: dir, model: "claude-sonnet-5-course", prompts: prompts, expect: expect,
		replies: ch16Talk(8), route: r.route, settings: ch16Settings,
	})
	if first.fatal != "" {
		res.fail("replay-needs-no-llm", "first run did not complete: %s", first.fatal)
		return
	}
	made := r.n
	if made == 0 {
		res.fail("replay-needs-no-llm", "no memories were produced, so replay proves nothing")
		return
	}

	// Restart. The agent rebuilds its context from the log and the files.
	r2 := &ch16Router{n: made}
	second := ch16Launch(bin, skills, gui, ch16Opts{
		dir: dir, model: "claude-sonnet-5-course", prompts: []string{"Carry on."}, expect: []int{1},
		replies: ch16Talk(2), route: r2.route, settings: ch16Settings,
	})
	if second.fatal != "" {
		res.fail("replay-needs-no-llm", "restart did not complete: %s", second.fatal)
		return
	}
	if got := len(ch16Compressors(second.reqs)); got > 0 {
		res.fail("replay-needs-no-llm",
			"restarting ran %d compressor requests. Rebuilding a context must replay recorded "+
				"decisions, not re-run the model that made them: replay would cost money, take time, "+
				"and produce a different context every time.", got)
	}
	last, ok := ch16Last(ch16Turns(second.reqs))
	if ok && !strings.Contains(ch16Whole(last.Body), ch16Memo(1)) {
		res.fail("replay-needs-no-llm",
			"after a restart the first memory is gone. Replay has to bring the bands back.")
	}
}

// ----------------------------------------------------------------
// Scenario 6: a failed or interrupted fold loses nothing
// ----------------------------------------------------------------

func ch16Survive(bin, skills, gui string, res *Ch16Result) {
	dir, cleanup := freshRunDir("ch16-survive")
	defer cleanup()

	// Every compressor call fails.
	r := &ch16Router{fail: true}
	var prompts []string
	var expect []int
	for i := 0; i < 5; i++ {
		prompts = append(prompts, ch16Prompt(i, ""))
		expect = append(expect, 2*(i+1))
	}
	out := ch16Launch(bin, skills, gui, ch16Opts{
		dir: dir, model: "claude-sonnet-5-course", prompts: prompts, expect: expect,
		replies: ch16Talk(7), route: r.route, settings: ch16Settings,
	})
	if out.fatal != "" {
		res.fail("abandon-on-restart",
			"the agent did not survive a compressor that returns 500: %s. "+
				"A failed fold is a normal event and must not end the turn.", out.fatal)
		return
	}
	turns := ch16Turns(out.reqs)
	last, ok := ch16Last(turns)
	if !ok {
		res.fail("abandon-on-restart", "no turns completed while the compressor was failing")
		return
	}
	// The work that could not be folded must still be there.
	if !strings.Contains(ch16Conversation(last.Body), "Work item 0") {
		res.fail("abandon-on-restart",
			"a compressor failed and the span it was folding disappeared anyway. "+
				"Nothing may be removed until the memory that replaces it exists.")
	}

	// Now let compaction succeed, and confirm nothing was double-folded.
	r2 := &ch16Router{}
	second := ch16Launch(bin, skills, gui, ch16Opts{
		dir: dir, model: "claude-sonnet-5-course", prompts: []string{"Carry on.", "And again."},
		expect:  []int{1, 2},
		replies: ch16Talk(3), route: r2.route, settings: ch16Settings,
	})
	if second.fatal != "" {
		res.fail("abandon-on-restart", "restart after failed folds did not complete: %s", second.fatal)
		return
	}
	lastOK, ok := ch16Last(ch16Turns(second.reqs))
	if ok {
		body := ch16Conversation(lastOK.Body)
		if n := strings.Count(body, ch16Memo(1)); n > 1 {
			res.fail("abandon-on-restart",
				"after retrying an abandoned fold the same memory appears %d times. "+
					"An interrupted fold must be abandoned, not resumed twice.", n)
		}
	}
}

// ----------------------------------------------------------------
// Parity
// ----------------------------------------------------------------

// ch16Parity is chapter 15's suite, run against this chapter's tree. The
// ladder and the journal still have to work.
func ch16Parity(path string, res *Ch16Result) {
	r := Ch15Run(path)
	if r.Fatal != "" {
		res.fail("ch15-parity", "chapter 15's checks could not run: %s", r.Fatal)
		return
	}
	if len(r.Errs) > 0 {
		var detail []string
		for id, msg := range r.Errs {
			if msg != "" {
				detail = append(detail, fmt.Sprintf("%s (%s)", id, msg))
			}
		}
		if len(detail) > 0 {
			sort.Strings(detail)
			res.fail("ch15-parity",
				"chapter 15 checks now fail: %s. Memory is added on top of the ladder, not instead of it.",
				strings.Join(detail, "; "))
		}
	}
}

// ch16MarkLiveMemory rewrites one memory file that is currently visible in
// the context, appending a line a human might add when correcting something
// the agent believes and should not.
//
// It does not assume a directory layout, a naming scheme, or which band the
// memory ended up in. It looks for a file whose text is present in live,
// which is the context as the model sees it, because that is the only
// property the chapter promises: what is on disk is what is in the context.
//
// Choosing a visible file matters. Uncompressed memories are never deleted,
// so once a memory has been folded upward its original stays on disk while
// the context shows the compressed version instead. Editing that superseded
// copy would correctly change nothing, and a check that did so would be
// testing the wrong file.
func ch16MarkLiveMemory(dir, live string) bool {
	var found string
	_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Size() > 1<<20 || found != "" {
			return nil
		}
		// Skip the event log and the save file: editing those is editing
		// the history, which is a different thing entirely.
		switch filepath.Base(p) {
		case "events.jsonl", "save.json", "settings.json":
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		// Match on the longest line in the file. A whole-body match is
		// brittle if the renderer adds a label or trims trailing space,
		// and a short line risks matching something unrelated.
		best := ""
		for _, ln := range strings.Split(string(b), "\n") {
			ln = strings.TrimSpace(ln)
			if len(ln) > len(best) {
				best = ln
			}
		}
		if len(best) >= 16 && strings.Contains(live, best) {
			found = p
		}
		return nil
	})
	if found == "" {
		return false
	}
	b, err := os.ReadFile(found)
	if err != nil {
		return false
	}
	out := string(b) + "\nCORRECTED-BY-HAND and rewritten on disk.\n"
	return os.WriteFile(found, []byte(out), 0o644) == nil
}


// ch16DropFirst removes the first entry from a comma separated list.
func ch16DropFirst(s string) string {
	parts := strings.Split(s, ",")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[1:], ",")
}
