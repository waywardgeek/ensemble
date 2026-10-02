package cachelens

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Lens implements common.CacheLens.
//
// It keeps the previous request so each new one can be compared against it, and
// leaves both on disk as cache-ordered JSON so that `diff prior_request.json
// request.json` puts the cursor on the first byte that broke the cache.
type Lens struct {
	dir string

	mu        sync.Mutex
	prior     Sections
	havePrior bool

	// last is the divergence found for the request most recently observed. It
	// is held until ObserveUsage arrives, because the verdict needs both the
	// prefix we computed and the counts the provider reported, and those two
	// facts arrive on different calls.
	last     Divergence
	haveLast bool

	// lastModel is the model that request was sent to, held for the same
	// reason last is: the verdict needs the model's caching style and minimum
	// cacheable size, and the model name arrives with the request while the
	// counts arrive with the usage.
	lastModel string

	// report receives one line per analyzed request. Injected rather than
	// hardcoded to a logger so this package depends on nothing but common.
	report func(string)
}

// New returns a Lens writing its captures under dir.
//
// A nil report is replaced with a discard, so a caller that only wants the
// files on disk does not have to supply one.
func New(dir string, report func(string)) *Lens {
	if report == nil {
		report = func(string) {}
	}
	return &Lens{dir: dir, report: report}
}

// requestPath and priorPath are the two files a human diffs.
func (l *Lens) requestPath() string { return filepath.Join(l.dir, "request.json") }
func (l *Lens) priorPath() string   { return filepath.Join(l.dir, "prior_request.json") }

// ObserveRequest captures a request, rotates the previous capture, and records
// where the two diverge.
//
// Nothing here may fail loudly enough to affect the turn. The lens is a
// diagnostic: if it cannot parse or cannot write, the correct behavior is to
// say so once and let the conversation proceed. An agent that refuses to answer
// because its instrumentation is unhappy is worse than an uninstrumented one.
func (l *Lens) ObserveRequest(model string, body []byte) {
	cur, err := Split(model, body)
	if err != nil {
		l.report("cachelens: " + err.Error())
		return
	}

	l.mu.Lock()
	prior, havePrior := l.prior, l.havePrior
	l.prior, l.havePrior = cur, true
	l.lastModel = model
	l.mu.Unlock()

	// Rotate on disk before writing the new capture, so the pair on disk is
	// always the pair that was compared.
	if havePrior {
		_ = os.Rename(l.requestPath(), l.priorPath())
	}
	if err := l.write(l.requestPath(), Pretty(cur)); err != nil {
		l.report("cachelens: " + err.Error())
	}

	if !havePrior {
		l.mu.Lock()
		l.haveLast = false
		l.mu.Unlock()
		l.report(fmt.Sprintf("cachelens: first request captured, %d bytes, no prior to compare",
			len(Canonical(cur))))
		return
	}

	d := Compare(prior, cur)

	l.mu.Lock()
	l.last, l.haveLast = d, true
	l.mu.Unlock()

	l.report("cachelens: " + Describe(d))
}

// ObserveUsage compares what the prefix made possible against what the provider
// actually charged for.
//
// The comparison is a ratio, not a token count. Prefix length is bytes and the
// provider reports tokens; converting between them needs the vendor's
// tokenizer, and an estimate dressed up as a token count is a number that gets
// requoted later as though it had been measured. Ratios are comparable without
// pretending to a precision we do not have.
func (l *Lens) ObserveUsage(u common.Usage) {
	l.mu.Lock()
	d, have := l.last, l.haveLast
	model := l.lastModel
	l.haveLast = false
	l.mu.Unlock()

	if !have {
		return
	}

	// The four categories are disjoint, so their sum is everything that was fed
	// in for this request.
	totalIn := u.Input + u.CacheRead + u.CacheWrite
	if totalIn == 0 {
		return
	}

	actual := float64(u.CacheRead) / float64(totalIn)

	possible := 0.0
	if d.CurrentBytes > 0 {
		possible = float64(d.CacheableBytes) / float64(d.CurrentBytes)
	}

	l.report(fmt.Sprintf(
		"cachelens: cacheable prefix %.1f%% of request, provider served %.1f%% from cache "+
			"(read %d, write %d, fresh %d)",
		possible*100, actual*100, u.CacheRead, u.CacheWrite, u.Input))

	// Two very different things produce a gap between what the prefix allowed
	// and what the provider served, and they must not share an alarm.
	//
	// The first is ours: something above the dialogue changed between turns, so
	// the prefix moved and no cache entry could match. That is nondeterminism
	// in content that was supposed to be fixed, this tool can point at the
	// byte, and it is worth waking someone for.
	//
	// The second is not ours: the prefix held and the provider still missed,
	// because the entry expired, or the prefix is under the model's minimum
	// cacheable size, or the model caches implicitly and simply has not seen
	// this prefix before. Normal, frequent, and nothing to fix in our code.
	// Which of those it was is decided below from the model table rather than
	// listed as a set of guesses, because a diagnostic that offers three
	// possible causes has not diagnosed anything.
	//
	// Note what is deliberately NOT an alarm: the dialogue growing. That
	// happens every single turn, so firing on it would make the instrument
	// noise within a day, and a warning that always fires also trains you to
	// ignore the one that matters.
	if d.Unstable != "" {
		l.report(fmt.Sprintf(
			"cachelens: PREFIX DIVERGED — section %q changed at +%d, above the dialogue. "+
				"Everything after it is uncacheable. Run: diff %s %s",
			d.Unstable, d.UnstableOffset, l.priorPath(), l.requestPath()))
		return
	}

	const gapThreshold = 0.20
	if possible-actual <= gapThreshold {
		return
	}

	// The prefix held and the provider still served less than it could have.
	// Before guessing at a cause, consult what is actually known about this
	// model, because two of the three "likely causes" a generic message would
	// offer are decidable facts rather than guesses:
	//
	//   - Below the vendor's floor, the request was never a caching candidate.
	//     Reporting that as a miss is how a healthy request gets read as broken.
	//
	//   - "No breakpoint is set" is a DEFECT on an explicit-caching model and
	//     CORRECT BEHAVIOUR on an implicit one. Saying it unconditionally sends
	//     someone hunting for a marker that should not exist.
	f, known := common.LookupModel(model)
	if !known {
		l.report(fmt.Sprintf(
			"cachelens: CACHE MISSED ANYWAY — the prefix held, so this is not a prefix "+
				"bug, but %q is not in the model table, so its caching style and minimum "+
				"cacheable size are unknown and no better diagnosis is available.", model))
		return
	}

	if totalIn < f.MinCacheTokens {
		l.report(fmt.Sprintf(
			"cachelens: NOT ELIGIBLE — %d prompt tokens is under %s's %d-token minimum, "+
				"so this request was never a caching candidate and a zero here is the "+
				"correct answer. It means the prompt is too small, NOT that the model "+
				"cannot cache.", totalIn, model, f.MinCacheTokens))
		return
	}

	switch f.Caching {
	case common.CacheImplicit:
		l.report(fmt.Sprintf(
			"cachelens: MISS ON AN IMPLICIT-CACHING MODEL — the prefix held, and %s finds "+
				"the repeated prefix itself, so there is no marker we could have "+
				"forgotten: sending no cache directives is correct here, not an omission. "+
				"Expect up to TWO cold turns, not one. The fixed head (system prompt and "+
				"tools) is served from the second turn, but a conversation message must "+
				"be stable across two successive requests before it is cached, so a "+
				"growing history lags by an extra turn. Caching also commits in blocks of "+
				"a few thousand tokens, so a turn that adds only a little text may show no "+
				"increase at all. Worry only if the count is still flat after that.",
			model))
	case common.CacheExplicit:
		if d.Breakpoints == 0 {
			l.report(fmt.Sprintf(
				"cachelens: NO BREAKPOINT — the prefix held, but the request carries zero "+
					"cache_control markers and %s caches only what we mark. This one is "+
					"ours: an unmarked prefix is billed in full every turn, in silence.",
				model))
			return
		}
		l.report(fmt.Sprintf(
			"cachelens: CACHE MISSED ANYWAY — the prefix held and %d breakpoint(s) are "+
				"set, so this is not a prefix bug and not a missing marker. The entry "+
				"most likely expired between turns.", d.Breakpoints))
	}
}

func (l *Lens) write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
