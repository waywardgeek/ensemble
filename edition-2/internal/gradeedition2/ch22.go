package gradeedition2

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter22 retains original16's eleven weights, with corrected Chapter14 parity.
// Numeric memory checks establish delivered bytes and lifecycle mechanics, not
// semantic compression quality or a model's willingness to obey a checkpoint.
func Chapter22(bin string) (grade.Report, error) {
	r := memoryChecks(bin)
	previous, err := Chapter14(bin)
	if err != nil {
		return grade.Report{}, err
	}
	r.Errs["ch15-parity"] = ""
	for _, c := range previous.Checks {
		if !c.Passed {
			r.fail("ch15-parity", "corrected Chapter14: %s: %s", c.ID, strings.Join(c.Details, "; "))
		}
	}
	checks := grade.Ch16Checks(grade.Ch16Result{Fatal: r.Fatal, Errs: r.Errs})
	checks[len(checks)-1].ID = "ch14-parity"
	checks[len(checks)-1].Title = "Chapter 14 context management and journaling remain functional."
	return grade.NewTitledReport("Edition 2, Chapter 22 — Forgetting on Purpose", checks, ""), nil
}

// Public settings, child selection and tree dismissal use the optional GUI.
// All memory work still runs in the submitted Agent; the provider is the sole fake.
// These tools are declared by both implementations, unlike the old think fixture.
func memoryChecks(bin string) contextResult {
	r := contextResult{Errs: map[string]string{}}
	for _, c := range grade.Ch16Checks(grade.Ch16Result{}) {
		if c.ID != "ch15-parity" {
			r.set(c.ID, "")
		}
	}
	tmp, err := os.MkdirTemp("", "memory-grade-*")
	if err != nil {
		r.Fatal = err.Error()
		return r
	}
	defer os.RemoveAll(tmp)
	skills, gui := filepath.Join(tmp, "skills"), filepath.Join(tmp, "gui")
	os.MkdirAll(filepath.Join(skills, "base"), 0755)
	os.MkdirAll(gui, 0755)
	err = os.WriteFile(filepath.Join(skills, "base", "SKILL.md"), []byte("---\nname: base\ndescription: Memory fixture\ntype: primary\ntools: read_file keep_tool_results micro_handoff agent_status\n---\nYou are a helpful agent.\n"), 0644)
	if err != nil {
		r.Fatal = err.Error()
		return r
	}
	run := func(o memoryOpts) contextOut {
		if o.model == "" {
			o.model = "claude-sonnet-5"
		}
		if o.settings == nil {
			o.settings = memorySettings(5000)
		}
		return memoryLaunch(bin, skills, gui, o)
	}
	newDir := func(name string) string { d := filepath.Join(tmp, name); os.MkdirAll(d, 0755); return d }
	base := newDir("compress")
	memoryCompress(run, base, &r)
	if err := memoryBackground(bin, newDir("background")); err != nil {
		r.fail("micro-handoff-compresses", "background lifecycle: %v", err)
	}
	memoryGraduate(run, newDir("graduate"), &r)
	memoryRestore(run, newDir("toggle"), &r)
	for _, edit := range []bool{false, true} {
		if err := memoryStale(bin, newDir(fmt.Sprintf("stale-%t", edit)), edit); err != nil {
			r.fail("disable-enable-idempotent", "held candidate edit=%t: %v", edit, err)
		}
	}
	memoryForce(run, newDir("force"), &r)
	for _, mode := range []string{"success", "failure", "interrupt"} {
		if err := memoryHeadroom(bin, newDir("headroom-"+mode), mode); err != nil {
			r.fail("forced-handoff", "%s pending headroom: %v", mode, err)
		}
	}
	for _, failed := range []bool{false, true} {
		usageErr, replayErr := memoryAccounting(run, newDir(fmt.Sprintf("accounting-%t", failed)), failed)
		if usageErr != nil {
			r.fail("forced-handoff", "final child accounting failed=%t: %v", failed, usageErr)
		}
		if replayErr != nil {
			r.fail("replay-needs-no-llm", "historical child accounting failed=%t: %v", failed, replayErr)
		}
	}
	memoryReplay(run, base, newDir, &r)
	memoryAbandon(run, newDir, &r)
	for _, mode := range []string{"kill", "shutdown", "write"} {
		if err := memoryLifetime(bin, newDir("lifetime-"+mode), mode); err != nil {
			r.fail("abandon-on-restart", "%s lifetime: %v", mode, err)
		}
	}
	for _, mode := range []string{"empty", "type", "length"} {
		if err := memoryProtocolFailure(run, newDir("protocol-"+mode), mode); err != nil {
			r.fail("abandon-on-restart", "%s candidate: %v", mode, err)
		}
	}
	memoryNeighbor(run, newDir("neighbor"), &r)
	memoryFresh(run, base, newDir("fresh"), &r)
	return r
}

type memoryRunner func(memoryOpts) contextOut

// Each scripted turn has one checkpoint and one final reply. Auxiliary calls
// do not advance these main-request counts, even when a fold fails.
func memoryOptions(dir string, n int, settings map[string]any, p *memoryRouter) memoryOpts {
	expect := make([]int, n)
	for i := range expect {
		expect[i] = 2 * (i + 1)
	}
	// The last checkpoint's main continuation may precede background acceptance.
	// Read its completed corpus in a separate, explicitly marked verification turn.
	prompts := append(memoryPrompts(n), "Verify the completed memory without making another checkpoint.")
	expect = append(expect, 2*n+1)
	replies := append(memoryTalk(n), fakevendor.Reply{Text: "Memory verified.", Usage: fakevendor.Canonical{Input: 100}})
	return memoryOpts{dir: dir, prompts: prompts, expect: expect, replies: replies, settings: settings, route: p.route}
}
func memoryCompleted(r *contextResult, out contextOut, ids ...string) bool {
	if out.fatal == "" && len(memoryTurns(out.reqs)) > 0 {
		return true
	}
	for _, id := range ids {
		r.fail(id, "session failed or made no main request: %s", out.fatal)
	}
	return false
}

// Keep this band large so returned session bytes remain available to compare.
// The fake preserves the planted fact only if the actual compressor receives it;
// this establishes delivery through both boundaries without pretending to summarize.
func memoryCompress(run memoryRunner, dir string, r *contextResult) {
	const cut = "micro-handoff-compresses"
	const fact = "planted-fact-survives"
	const data = "memory-is-data"
	p := &memoryRouter{}
	o := memoryOptions(dir, 6, memorySettings(5000), p)
	o.prompts[0] += " The exact fact is " + memoryPlant + "."
	out := run(o)
	if !memoryCompleted(r, out, cut, fact, data) {
		return
	}
	if p.n == 0 || len(p.bodies) == 0 {
		for _, id := range []string{cut, fact, data} {
			r.fail(id, "no submit-only compressor request")
		}
		return
	}
	last := memoryConversation(memoryLast(out))
	first := memoryConversation(p.bodies[0])
	if !strings.Contains(first, "Work item 0.") || !strings.Contains(first, "END-WORK-0.") {
		r.fail(cut, "first compressor did not receive both ends of completed work")
	}
	if strings.Contains(last, "Work item 0.") || strings.Contains(last, "END-WORK-0.") {
		r.fail(cut, "compressed raw work remains in conversation")
	}
	if !strings.Contains(first, memoryPlant) || !strings.Contains(last, memoryMemo(1)+" I remember "+memoryPlant+".") {
		r.fail(fact, "planted fact did not reach compressor and return as its exact submitted memory")
	}
	if !strings.Contains(last, memoryMemo(1)) || strings.Contains(memorySystem(memoryLast(out)), "MEMO-") {
		r.fail(data, "memory missing from conversation or inserted into system")
	}
}

// With these byte sizes, graduation must leave newer files behind. Folding the
// entire band would make an oldest/newest mutation observationally equivalent.
// The source tests separately verify low-watermark arithmetic and higher tiers.
func memoryGraduate(run memoryRunner, dir string, r *contextResult) {
	const id = "graduation-fires-oldest-first"
	p := &memoryRouter{}
	out := run(memoryOptions(dir, 8, memorySettings(500), p))
	if !memoryCompleted(r, out, id) {
		return
	}
	if p.folds == 0 || p.firstFoldN < 3 {
		r.fail(id, "no distinguishing graduation after successful session populations")
		return
	}
	first := memoryConversation(p.foldBodies[0])
	last := memoryConversation(memoryLast(out))
	if !strings.Contains(first, memoryMemo(1)) || strings.Contains(first, memoryMemo(p.firstFoldN)) {
		r.fail(id, "first fold must include oldest and leave the newest of %d populated files", p.firstFoldN)
	}
	if !strings.Contains(last, memoryFold(1)) || strings.Contains(last, memoryMemo(1)) || !strings.Contains(last, memoryMemo(p.n)) {
		r.fail(id, "fold result, retired oldest or retained newest is wrong")
	}
}

// Compare marker positions in actual provider order, including duplicates. The
// old helper iterated numeric IDs and therefore returned its own order instead.
func memoryOrder(text string) string {
	return strings.Join(regexp.MustCompile(`(?:MEMO|FOLD)-[0-9]+`).FindAllString(text, -1), ",")
}

// Edit a real live corpus file whose complete bytes were delivered. Never choose
// an arbitrary longest line or silently skip the test when no file can be found.
func memoryEdit(dir, live, marker string) error {
	found := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if found || info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		text := strings.TrimSpace(string(b))
		if memoryOrder(text) == "" || !strings.Contains(live, text) {
			return nil
		}
		if e = os.WriteFile(path, append(b, []byte("\n"+marker)...), info.Mode().Perm()); e != nil {
			return e
		}
		found = true
		return nil
	})
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("no live memory file could be edited")
	}
	return nil
}

// Ordinary query replies cannot create a new checkpoint. Thus changes seen here
// come from disabling, disk edits and startup/re-enable, not another compression.
// Re-enable the upper band first to challenge order based on arrival time.
func memoryRestore(run memoryRunner, dir string, r *contextResult) {
	const id = "disable-enable-idempotent"
	p := &memoryRouter{}
	out := run(memoryOptions(dir, 8, memorySettings(500), p))
	if !memoryCompleted(r, out, id) {
		return
	}
	want := memoryConversation(memoryLast(out))
	order := memoryOrder(want)
	if !strings.Contains(order, "FOLD-") || !strings.Contains(order, "MEMO-") {
		r.fail(id, "both memory bands were not populated")
		return
	}
	plain := func(settings map[string]any) contextOut {
		return run(memoryOpts{dir: dir, settings: settings, prompts: []string{"Inspect the current memory."}, replies: []fakevendor.Reply{{Text: "Observed.", Usage: fakevendor.Canonical{Input: 100}}}, route: p.route})
	}
	off := plain(memorySettings(500, "session", "8x"))
	if !memoryCompleted(r, off, id) {
		return
	}
	if memoryOrder(memoryConversation(memoryLast(off))) != "" {
		r.fail(id, "disabled bands remain in conversation")
		return
	}
	if e := memoryEdit(dir, want, "CORRECTED-BY-HAND"); e != nil {
		r.fail(id, "disk edit required: %v", e)
		return
	}
	mid := plain(memorySettings(500, "session"))
	if !memoryCompleted(r, mid, id) {
		return
	}
	if s := memoryOrder(memoryConversation(memoryLast(mid))); !strings.Contains(s, "FOLD-") || strings.Contains(s, "MEMO-") {
		r.fail(id, "partial restore did not expose only upper band")
		return
	}
	back := plain(memorySettings(500))
	if !memoryCompleted(r, back, id) {
		return
	}
	got := memoryConversation(memoryLast(back))
	if memoryOrder(got) != order || !strings.Contains(got, "CORRECTED-BY-HAND") {
		r.fail(id, "restored memory order/duplicates or actual disk edit differs: before %s after %s", order, memoryOrder(got))
		return
	}
	if e := memoryEdit(dir, got, "RESTART-EDIT-WINS"); e != nil {
		r.fail(id, "second disk edit required: %v", e)
		return
	}
	again := plain(memorySettings(500))
	if !memoryCompleted(r, again, id) {
		return
	}
	if !strings.Contains(memoryConversation(memoryLast(again)), "RESTART-EDIT-WINS") {
		r.fail(id, "ordinary enabled restart ignored edited disk bytes")
	}
	before := len(p.bodies)
	quiet := run(memoryOptions(dir, 3, memorySettings(500, "session", "8x"), p))
	if !memoryCompleted(r, quiet, id) {
		return
	}
	if len(p.bodies) != before {
		r.fail(id, "disabled destination still invokes compressor")
	}
}

func memoryForce(run memoryRunner, dir string, r *contextResult) {
	const id = "forced-handoff"
	// Anthropic's three raw input buckets are disjoint. Output does not contribute.
	// A completed handoff clears the latch; smaller subsequent MAIN usage must
	// leave it clear even though both lifetime and auxiliary usage have grown.
	replies := []fakevendor.Reply{
		{Text: "Low occupancy.", Usage: fakevendor.Canonical{Input: 100, Output: 9000}},
		{Text: "Near limit.", Usage: fakevendor.Canonical{Input: 100, CacheWrite: 1800, CacheRead: 1800}},
		{Text: "At limit.", Usage: fakevendor.Canonical{Input: 100, CacheWrite: 1800, CacheRead: 2050}},
		{ToolName: "micro_handoff", ToolID: "forced", ToolArgs: `{"text":"I completed this checkpoint."}`, Usage: fakevendor.Canonical{Input: 100, CacheRead: 3850}},
		{Text: "Small again.", Usage: fakevendor.Canonical{Input: 100}},
		{Text: "Still small.", Usage: fakevendor.Canonical{Input: 100}},
	}
	p := &memoryRouter{usage: fakevendor.Canonical{Input: 12000}}
	out := run(memoryOpts{dir: dir, model: "claude-ch16-course", prompts: []string{"One.", "Two.", "Three.", "Four.", "Five."}, expect: []int{1, 2, 3, 5, 6}, replies: replies, route: p.route})
	if !memoryCompleted(r, out, id) {
		return
	}
	qs := memoryTurns(out.reqs)
	if len(qs) != 6 {
		r.fail(id, "got %d main requests, want6", len(qs))
		return
	}
	for _, i := range []int{0, 1, 2, 5} {
		if len(memoryTools(qs[i].Body)) < 2 {
			r.fail(id, "premature/sticky forcing at main request%d", i+1)
		}
	}
	// The first continuation after handoff still follows a high-input response.
	// Implementations may release on that checkpoint or await its smaller reply;
	// request six must expose ordinary tools after that smaller reply completes.
	if n := memoryTools(qs[3].Body); len(n) != 1 || n[0] != "micro_handoff" {
		r.fail(id, "95%% does not restrict to handoff: %v", n)
	}
	// The chapter leaves warning wording free. Remove only the exact fixture
	// conversation, then observe new nonempty text at the warning boundary.
	// This proves delivery, not whether arbitrary added prose is useful advice.
	extra := func(body []byte) string {
		text := memoryConversation(body)
		for _, known := range []string{"One.", "Two.", "Three.", "Low occupancy.", "Near limit."} {
			text = strings.Replace(text, known, "", 1)
		}
		return strings.Join(strings.Fields(text), " ")
	}
	before, after := extra(qs[1].Body), extra(qs[2].Body)
	if after == "" || after == before {
		r.fail(id, "warning boundary added no message text")
	}
	// Force an actual auxiliary call while main replies remain small; its high
	// reported input must not narrow the next ordinary request's declarations.
	aux := run(memoryOpts{dir: dir, model: "claude-ch16-course", prompts: memoryPrompts(2), expect: []int{2, 4}, replies: memoryTalk(2), route: p.route})
	if !memoryCompleted(r, aux, id) {
		return
	}
	if p.n == 0 {
		r.fail(id, "no actual auxiliary call for occupancy exclusion")
		return
	}
	for i, q := range memoryTurns(aux.reqs) {
		if len(memoryTools(q.Body)) < 2 {
			r.fail(id, "auxiliary usage restricted main request%d", i+1)
		}
	}
}

// Normal startup intentionally refreshes enabled disk files. This checks the
// executable restoration path and actual next-request equality; pure reducer
// reconstruction without filesystem access retains its separate source proof.
func memoryReplay(run memoryRunner, base string, newDir func(string) string, r *contextResult) {
	const id = "replay-needs-no-llm"
	a, b := newDir("snapshot"), newDir("replay")
	if e := contextCopyDir(base, a); e != nil {
		r.fail(id, "copy: %v", e)
		return
	}
	if e := contextCopyDir(base, b); e != nil {
		r.fail(id, "copy: %v", e)
		return
	}
	save := filepath.Join(b, "save.json")
	data, e := os.ReadFile(save)
	if e != nil {
		r.fail(id, "save: %v", e)
		return
	}
	data, e = contextNullSnapshot(data)
	if e != nil {
		r.fail(id, "snapshot: %v", e)
		return
	}
	if e = os.WriteFile(save, data, 0644); e != nil {
		r.fail(id, "null snapshot: %v", e)
		return
	}
	// A startup-only launch reaches the public settings barrier and exits without
	// any vendor request. Then identical explicit prompts compare actual wire bytes.
	idle := run(memoryOpts{dir: b})
	if idle.fatal != "" || len(idle.reqs) != 0 {
		r.fail(id, "startup replay contacted vendor or failed: %s requests%d", idle.fatal, len(idle.reqs))
		return
	}
	query := func(dir string) contextOut {
		return run(memoryOpts{dir: dir, prompts: []string{"What is remembered?"}, replies: []fakevendor.Reply{{Text: "Done."}}})
	}
	left, right := query(a), query(b)
	if !memoryCompleted(r, left, id) || !memoryCompleted(r, right, id) {
		return
	}
	if len(left.reqs) != 1 || len(right.reqs) != 1 || !bytes.Equal(left.reqs[0].Body, right.reqs[0].Body) {
		r.fail(id, "snapshot and full-log replay produced different requests or auxiliary calls")
	}
	if !strings.Contains(memoryConversation(memoryLast(right)), memoryMemo(1)) {
		r.fail(id, "replay did not retain returned memory")
	}
}

// Exceed the unchanged high watermark while the destination is disabled, then
// require every submitted source memory exactly once. The warning diagnostic is
// covered by the existing runtime source test, not inferred from this absence.
func memoryNeighbor(run memoryRunner, dir string, r *contextResult) {
	const id = "disabled-neighbor-refused"
	p := &memoryRouter{}
	out := run(memoryOptions(dir, 8, memorySettings(500, "8x"), p))
	if !memoryCompleted(r, out, id) {
		return
	}
	if p.n < 6 || p.n*len(memoryMemo(1)) <= 1000 {
		r.fail(id, "session band never crossed the high watermark")
		return
	}
	if p.folds != 0 {
		r.fail(id, "graduated into disabled neighbor")
		return
	}
	text := memoryConversation(memoryLast(out))
	for i := 1; i <= p.n; i++ {
		if strings.Count(text, memoryMemo(i)) != 1 {
			r.fail(id, "refused graduation lost/duplicated memory%d", i)
			return
		}
	}
}

func memoryFresh(run memoryRunner, base, dir string, r *contextResult) {
	const id = "fresh-start-populates"
	count := 0
	// Copy only the actual generated live Markdown corpus, preserving its own
	// relative paths. No save, journal or historical transcript can restore it.
	err := filepath.Walk(base, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if memoryOrder(string(b)) == "" {
			return nil
		}
		rel, _ := filepath.Rel(base, path)
		dst := filepath.Join(dir, rel)
		if e = os.MkdirAll(filepath.Dir(dst), 0755); e != nil {
			return e
		}
		count++
		return os.WriteFile(dst, b, 0644)
	})
	if err != nil || count == 0 {
		r.fail(id, "no actual corpus copied: %d %v", count, err)
		return
	}
	out := run(memoryOpts{dir: dir, prompts: []string{"What is remembered?"}, replies: []fakevendor.Reply{{Text: "Done."}}})
	if !memoryCompleted(r, out, id) {
		return
	}
	qs := memoryTurns(out.reqs)
	if len(out.reqs) != 1 || !strings.Contains(memoryConversation(qs[0].Body), memoryMemo(1)) {
		r.fail(id, "first request without a save omitted disk memory or invoked a compressor")
	}
}

// Populate successfully before either failure mode. Returning errors from every
// compressor would leave an empty band and never exercise abandoned graduation.
// The two runs distinguish an actual HTTP error from a request still pending
// when the process dies; neither may spend another call merely to restart.
func memoryAbandon(run memoryRunner, newDir func(string) string, r *contextResult) {
	const id = "abandon-on-restart"
	for _, pending := range []bool{false, true} {
		dir := newDir(fmt.Sprintf("abandon-%v", pending))
		p := &memoryRouter{}
		first := run(memoryOptions(dir, 3, memorySettings(5000), p))
		if !memoryCompleted(r, first, id) {
			return
		}
		if p.n < 3 || p.folds != 0 {
			r.fail(id, "could not populate session before graduation failure")
			return
		}
		p.failFold = true
		o := memoryOptions(dir, 1, memorySettings(250), p)
		if pending {
			p.pending = make(chan struct{})
			p.release = make(chan struct{})
			o.pending = p.pending
			o.release = p.release
			o.unblock = p.unblock
			// Kill at the actual checkpoint, before the ordinary verification turn.
			o.prompts = o.prompts[:1]
		}
		failed := run(o)
		if failed.fatal != "" || p.folds == 0 {
			r.fail(id, "actual failed/pending fold not established: pending%v folds%d %s", pending, p.folds, failed.fatal)
			return
		}
		// No response from the blocked peer can have installed a replacement before
		// SIGKILL. Startup must abandon that attempt and expose each source once.
		resumed := run(memoryOpts{dir: dir, settings: memorySettings(250), prompts: []string{"Inspect retained memories."}, replies: []fakevendor.Reply{{Text: "Retained."}}})
		if !memoryCompleted(r, resumed, id) {
			return
		}
		if len(resumed.reqs) != 1 {
			r.fail(id, "restart reran abandoned compressor")
			return
		}
		text := memoryConversation(resumed.reqs[0].Body)
		for i := 1; i <= p.n; i++ {
			if strings.Count(text, memoryMemo(i)) != 1 {
				r.fail(id, "pending%v source memory%d lost/duplicated after restart", pending, i)
				return
			}
		}
		if strings.Contains(text, "FOLD-") {
			r.fail(id, "failed fold was populated")
			return
		}
		recovered := &memoryRouter{n: p.n}
		success := run(memoryOptions(dir, 1, memorySettings(250), recovered))
		if !memoryCompleted(r, success, id) {
			return
		}
		if recovered.folds == 0 || strings.Count(memoryConversation(memoryLast(success)), memoryFold(1)) != 1 {
			r.fail(id, "next actual checkpoint failed to recover exactly one fold")
			return
		}
	}
}
