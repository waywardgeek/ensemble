package grade

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Ch18Result records which checks ran and which failed.
type Ch18Result struct {
	Errs  map[string]string
	Fatal string
}

func (r *Ch18Result) ran(id string) {
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	if _, ok := r.Errs[id]; !ok {
		r.Errs[id] = ""
	}
}

func (r *Ch18Result) fail(id, format string, args ...any) {
	r.ran(id)
	if r.Errs[id] == "" {
		r.Errs[id] = fmt.Sprintf(format, args...)
	}
}

// Priced models, one per vendor. The cost checks need a model with a real
// price sheet: an unpriced model is not free, it is unpriced, and it reports a
// dash rather than a number.
const (
	ch18Anthropic = "claude-sonnet-5"
	ch18OpenAI    = "gpt-6-astra"
	ch18Gemini    = "gemini-3.8-flash"
)

// The two turns of the main scenario, chosen so that every number the grader
// asserts is distinguishable from every other. A cold turn writes the cache
// and reads nothing; a warm turn reads most of its prefix back.
var (
	ch18Cold = fakevendor.Canonical{Input: 1000, CacheWrite: 500, CacheRead: 0, Output: 10}
	ch18Warm = fakevendor.Canonical{Input: 100, CacheWrite: 0, CacheRead: 1400, Output: 20}
	// The restart turn. Small and coprime to the numbers above, so a meter
	// that leaks the restored lifetime total cannot coincidentally match.
	ch18Restart = fakevendor.Canonical{Input: 7, CacheWrite: 0, CacheRead: 0, Output: 3}
	// One turn per vendor, four distinct values so a parser that crosses two
	// fields is caught rather than averaged out.
	ch18PerVendor = fakevendor.Canonical{Input: 11, CacheWrite: 22, CacheRead: 33, Output: 44}
)

func Ch18Run(dir string) Ch18Result {
	var res Ch18Result

	bin, cleanup, err := Build(dir)
	if err != nil {
		res.Fatal = fmt.Sprintf("build failed: %v", err)
		return res
	}
	defer cleanup()

	skills, skillsCleanup := ch18Skills()
	defer skillsCleanup()
	gui := filepath.Join(dir, "web")

	ch18Prefix(&res, bin, skills, gui)
	ch18Vendors(&res, bin, skills, gui)

	return res
}

// ch18Skills gives the agent a minimal primary skill, so the system prompt has
// stable content to cache without dragging in a chapter's worth of skill text.
func ch18Skills() (string, func()) {
	dir, err := os.MkdirTemp("", "ch18-skills-")
	if err != nil {
		return "", func() {}
	}
	base := filepath.Join(dir, "base")
	os.MkdirAll(base, 0o755)
	os.WriteFile(filepath.Join(base, "SKILL.md"), []byte(
		"---\nname: base\ndescription: baseline skill\n---\n\nAnswer briefly.\n"), 0o644)
	return dir, func() { os.RemoveAll(dir) }
}

// ---------------------------------------------------------------------------
// Scenario 1: two turns, then a restart in the same directory
// ---------------------------------------------------------------------------

func ch18Prefix(res *Ch18Result, bin, skills, gui string) {
	ids := []string{
		"cache-lens-exists", "prefix-is-stable", "system-has-breakpoint",
		"cost-is-computed-not-stored", "cache-read-rate", "usage-is-session-scoped",
	}
	for _, id := range ids {
		res.ran(id)
	}

	work, cleanup := freshRunDir("ch18-prefix")
	defer cleanup()

	out := ch18Launch(bin, skills, gui, ch18Opts{
		dir:     work,
		model:   ch18Anthropic,
		vendor:  "anthropic",
		prompts: []string{"first question", "second question"},
		replies: []fakevendor.Reply{
			{Text: "cold answer", Usage: ch18Cold},
			{Text: "warm answer", Usage: ch18Warm},
		},
	})
	if out.fatal != "" {
		for _, id := range ids {
			res.fail(id, "%s", out.fatal)
		}
		return
	}

	ch18CheckLens(res, out)
	ch18CheckStable(res, out)
	ch18CheckSystemBreakpoint(res, out)
	ch18CheckCost(res, out)
	ch18CheckCacheRead(res, out)

	// The restart reuses the working directory, so save.json is there to be
	// restored. That is the whole point: the lifetime total exists and the
	// meter must still report only this session.
	ch18CheckSessionScope(res, out, bin, skills, gui, work)
}

// ch18CheckLens: prior_request.json cannot exist unless the lens observed a
// SECOND request and rotated the first. That is the behaviour the check wants,
// and it is invisible to any amount of code inspection.
func ch18CheckLens(res *Ch18Result, out ch18Out) {
	const id = "cache-lens-exists"
	for _, name := range ch18LensNames {
		if len(out.lensFiles[name]) == 0 {
			res.fail(id, "the lens left no %s after two turns: nothing observed consecutive requests, "+
				"so a broken prefix would be undetectable", name)
			return
		}
	}
	// Both files must contain the request material itself, split into the
	// sections that lay out in cache order. A lens that writes only a verdict
	// cannot put a diff on the first broken byte, which is the deliverable:
	// the point of rotating two files is that plain diff does the last mile.
	for _, name := range ch18LensNames {
		body := string(out.lensFiles[name])
		if !ch18HasSection(body, "messages") && !ch18HasSection(body, "contents") {
			res.fail(id, "%s contains no messages section, so the lens did not capture a turn "+
				"request in a form a diff can read", name)
			return
		}
		if !ch18HasSection(body, "system") && !ch18HasSection(body, "systemInstruction") {
			res.fail(id, "%s contains no system section: the largest stable part of the prefix "+
				"is exactly the part worth watching", name)
			return
		}
	}
	if string(out.lensFiles["request.json"]) == string(out.lensFiles["prior_request.json"]) {
		res.fail(id, "request.json and prior_request.json are identical: "+
			"the lens is not rotating, so every comparison is a request against itself")
	}
}

// ch18HasSection reports whether a rotated lens file carries a named section.
// Matched loosely on the name rather than on a framing style, because the
// requirement is that a human can diff the file, not that they can diff it in
// one particular decoration.
func ch18HasSection(body, name string) bool {
	return strings.Contains(body, name)
}

// ch18CheckStable measures the prefix itself rather than reading the lens's
// verdict, because the lens is under test. An instrument that reports
// "identical" unconditionally would otherwise score full marks here.
func ch18CheckStable(res *Ch18Result, out ch18Out) {
	const id = "prefix-is-stable"
	turns := ch18TurnBodies(out.reqs)
	if len(turns) < 2 {
		res.fail(id, "only %d turn requests reached the vendor, need 2 to compare a prefix", len(turns))
		return
	}
	a, b := ch18Sections(turns[0]), ch18Sections(turns[1])
	if a == nil || b == nil {
		res.fail(id, "could not parse the request bodies as JSON objects")
		return
	}
	for _, name := range []string{"tools", "system"} {
		av, aok := a[name]
		bv, bok := b[name]
		if !aok || !bok {
			// A vendor that declares tools inline has no tools array; not a
			// failure, just nothing to compare.
			continue
		}
		if av != bv {
			res.fail(id, "the %s section changed between turn 1 and turn 2 (%d bytes then %d bytes). "+
				"An unstable prefix cannot be cached, and no amount of breakpoints fixes it",
				name, len(av), len(bv))
			return
		}
	}
}

func ch18CheckSystemBreakpoint(res *Ch18Result, out ch18Out) {
	const id = "system-has-breakpoint"
	turns := ch18TurnBodies(out.reqs)
	if len(turns) == 0 {
		res.fail(id, "no turn requests reached the vendor")
		return
	}
	if !ch18HasSystemBreakpoint(turns[0]) {
		res.fail(id, "the system section carries no cache_control marker on the wire. "+
			"Without it the whole system prompt is re-read at full price every turn, "+
			"and nothing in the agent's behaviour changes to say so")
	}
}

// ch18CheckCost enforces the division of labour: the log holds counts, the
// display holds money. Prices change and counts are history, so a stored
// dollar amount is wrong the moment a vendor reprices.
func ch18CheckCost(res *Ch18Result, out ch18Out) {
	const id = "cost-is-computed-not-stored"

	if len(out.rawRespUsage) == 0 {
		res.fail(id, "no usage was recorded in the event log, so there is nothing to derive a cost from")
		return
	}
	for _, u := range out.rawRespUsage {
		for k, v := range u {
			if ch18IsMoneyKey(k) {
				res.fail(id, "the recorded usage carries a monetary field %q: prices change and counts "+
					"are history, so cost must be derived at display time, never stored", k)
				return
			}
			if f, ok := v.(float64); ok && f != float64(int64(f)) {
				res.fail(id, "the recorded usage field %q holds a fractional value (%v): "+
					"usage counts tokens, never money", k, v)
				return
			}
		}
	}

	if len(out.usage) == 0 {
		res.fail(id, "the meter never reported usage, so the cost is invisible at the moment it is incurred")
		return
	}
	last := out.usage[len(out.usage)-1]
	if !last.Priced {
		res.fail(id, "the meter reports model %s as unpriced, but it has a price sheet", ch18Anthropic)
		return
	}
	if last.CostUSD <= 0 {
		res.fail(id, "the meter reports a cost of %v for %d input and %d output tokens on a priced model",
			last.CostUSD, last.Input, last.Output)
	}
}

func ch18IsMoneyKey(k string) bool {
	k = strings.ToLower(k)
	for _, bad := range []string{"cost", "dollar", "usd", "price", "spend", "charge"} {
		if strings.Contains(k, bad) {
			return true
		}
	}
	return false
}

// ch18CheckCacheRead proves the agent reports what the provider actually said,
// rather than a plausible-looking number of its own.
func ch18CheckCacheRead(res *Ch18Result, out ch18Out) {
	const id = "cache-read-rate"
	got := ch18Responses(out.events)
	if len(got) < 2 {
		res.fail(id, "the event log holds %d responses with usage, need 2", len(got))
		return
	}
	if got[0].CacheRead != 0 {
		res.fail(id, "the first turn reports %d cache-read tokens. The first request of a run "+
			"has nothing to read: a nonzero figure here means the number is invented, not parsed",
			got[0].CacheRead)
		return
	}
	if got[1].CacheRead != ch18Warm.CacheRead {
		res.fail(id, "the warm turn reports %d cache-read tokens, the provider said %d. "+
			"A cache hit that goes unreported is indistinguishable from no cache at all",
			got[1].CacheRead, ch18Warm.CacheRead)
		return
	}
	if len(out.usage) > 0 {
		if last := out.usage[len(out.usage)-1]; last.HitRate <= 0 {
			res.fail(id, "the meter reports a hit rate of %v after a turn that read %d tokens from cache",
				last.HitRate, ch18Warm.CacheRead)
		}
	}
}

// ch18CheckSessionScope restarts the agent in the same directory. The lifetime
// total is on disk and will be restored into the context; the meter must still
// report only what this session spent, because "session cost" means the cost
// since you sat down.
func ch18CheckSessionScope(res *Ch18Result, first ch18Out, bin, skills, gui, work string) {
	const id = "usage-is-session-scoped"

	// Guard the fixture: if usage never accumulated in the first place, a
	// meter reading zero after restart would pass while proving nothing.
	if len(first.usage) == 0 {
		res.fail(id, "the meter never reported usage during the first run")
		return
	}
	before := first.usage[len(first.usage)-1]
	wantFirst := ch18Cold.Input + ch18Warm.Input
	if before.Input != wantFirst {
		res.fail(id, "after two turns the meter reports %d input tokens, expected the session total %d. "+
			"Session scoping cannot be judged until accumulation works", before.Input, wantFirst)
		return
	}
	if _, err := os.Stat(filepath.Join(work, "save.json")); err != nil {
		res.fail(id, "no save.json after the first run, so there is no restored lifetime total "+
			"for the meter to leak and this check cannot distinguish the two sources")
		return
	}

	out := ch18Launch(bin, skills, gui, ch18Opts{
		dir:     work,
		model:   ch18Anthropic,
		vendor:  "anthropic",
		prompts: []string{"after restart"},
		replies: []fakevendor.Reply{{Text: "fresh answer", Usage: ch18Restart}},
	})
	if out.fatal != "" {
		res.fail(id, "restart run failed: %s", out.fatal)
		return
	}
	if len(out.usage) == 0 {
		res.fail(id, "the meter reported nothing after the restart")
		return
	}
	last := out.usage[len(out.usage)-1]
	if last.Input == ch18Restart.Input {
		return // session-scoped, as required
	}
	if last.Input >= wantFirst {
		res.fail(id, "after a restart the meter reports %d input tokens, but this session spent only %d. "+
			"It is reading the lifetime total restored from save.json, so the number grows forever "+
			"and never answers what this sitting cost", last.Input, ch18Restart.Input)
		return
	}
	res.fail(id, "after a restart the meter reports %d input tokens, expected %d",
		last.Input, ch18Restart.Input)
}

// ---------------------------------------------------------------------------
// Scenario 2: one turn per vendor
// ---------------------------------------------------------------------------

// ch18CheckVendors runs the same turn against all three providers. Each
// reports usage in its own convention, and the disagreement is the point:
// Anthropic's input count EXCLUDES cache reads, OpenAI's prompt total
// INCLUDES them. A parser that copies the field across without normalising
// produces a number that is wrong by exactly the size of the cache hit, which
// is largest precisely when caching is working best.
func ch18Vendors(res *Ch18Result, bin, skills, gui string) {
	const id = "all-vendors-report-usage"
	res.ran(id)

	for _, v := range []struct {
		vendor string
		model  string
		// cacheWrite is false where the provider has no cache-write bucket on
		// the wire. Grading a field a vendor never reports would demand an
		// invented number, which is the opposite of this chapter's lesson.
		cacheWrite bool
	}{
		{"anthropic", ch18Anthropic, true},
		{"openai", ch18OpenAI, true},
		{"gemini", ch18Gemini, false},
	} {
		out := ch18Launch(bin, skills, gui, ch18Opts{
			dir:     ch18TempDir(res, id, v.vendor),
			model:   v.model,
			vendor:  v.vendor,
			prompts: []string{"one question"},
			replies: []fakevendor.Reply{{Text: "an answer", Usage: ch18PerVendor}},
		})
		if out.fatal != "" {
			res.fail(id, "%s run failed: %s", v.vendor, out.fatal)
			return
		}
		got := ch18Responses(out.events)
		if len(got) == 0 {
			res.fail(id, "the %s parser recorded no usage at all", v.vendor)
			return
		}
		u := got[0]
		if u.Input == 0 && u.CacheWrite == 0 && u.CacheRead == 0 && u.Output == 0 {
			res.fail(id, "the %s parser returned zero for all four usage categories on a priced model. "+
				"A vendor that reports nothing makes every cost figure downstream a guess", v.vendor)
			return
		}
		want := ch18PerVendor
		if u.Input != want.Input {
			res.fail(id, "%s: input is %d, the provider reported %d. Note that this provider's prompt "+
				"count and the agent's Input field are not the same quantity when a cache hit occurs",
				v.vendor, u.Input, want.Input)
			return
		}
		if u.CacheRead != want.CacheRead {
			res.fail(id, "%s: cache-read is %d, the provider reported %d", v.vendor, u.CacheRead, want.CacheRead)
			return
		}
		if u.Output != want.Output {
			res.fail(id, "%s: output is %d, the provider reported %d", v.vendor, u.Output, want.Output)
			return
		}
		if v.cacheWrite && u.CacheWrite != want.CacheWrite {
			res.fail(id, "%s: cache-write is %d, the provider reported %d", v.vendor, u.CacheWrite, want.CacheWrite)
			return
		}
	}
}

func ch18TempDir(res *Ch18Result, id, name string) string {
	dir, err := os.MkdirTemp("", "ch18-"+name+"-")
	if err != nil {
		res.fail(id, "temp dir: %v", err)
		return ""
	}
	return dir
}

// ---------------------------------------------------------------------------

// ch18TurnBodies returns the bodies of requests that declare tools, which is
// what distinguishes a turn from the judge and compressor calls that chapters
// 16 and 17 added. Classifying by shape rather than by position keeps these
// assertions true however many judge calls a student's recall decides to make.
func ch18TurnBodies(reqs []fakevendor.Recorded) [][]byte {
	var out [][]byte
	for _, r := range reqs {
		var probe struct {
			Tools []json.RawMessage `json:"tools"`
		}
		if json.Unmarshal(r.Body, &probe) != nil {
			continue
		}
		if len(probe.Tools) == 0 {
			continue
		}
		out = append(out, r.Body)
	}
	return out
}
