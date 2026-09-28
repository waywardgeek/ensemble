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
	// because the entry expired, or nothing marks a breakpoint, or the prefix
	// is under the model's minimum cacheable size. Normal, frequent, and
	// nothing to fix in our code.
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
	if possible-actual > gapThreshold {
		l.report("cachelens: CACHE MISSED ANYWAY — the prefix held, so this is not a " +
			"prefix bug. Likely no breakpoint is set, the entry expired, or the " +
			"prefix is below the model's minimum cacheable size.")
	}
}

func (l *Lens) write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
