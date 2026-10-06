package llm

import (
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// TestMeasureMemory prints the numbers chapter 16 is supposed to print, rather
// than asserting anything. Run it with:
//
//	go test ./internal/llm/ -run TestMeasureMemory -v
//
// What it can and cannot measure is worth being clear about. The compressor
// here is a fake that answers with a fixed string, so the COMPRESSION RATIO it
// achieves is a property of the fake and means nothing. What is real is the
// shape of the system around it: how many memories a session produces, how
// often the cascade fires, what each band ends up holding against its budget,
// and how much smaller the context is with memory switched on than with it
// switched off. Those are structural and do not depend on what the compressor
// writes.
func TestMeasureMemory(t *testing.T) {
	type result struct {
		bands     map[common.Band]int
		convo     int
		folds     int
		memories  int
		turns     int
		budgetSes int
		budgetCon int
	}

	run := func(on bool) result {
		f := newFakeCompressor(t)
		cfg := common.DefaultBandConfig()
		cfg.Conversation.Budget = 3000
		s := cfg.For(common.BandSession)
		s.Budget = 1500
		cfg.Set(common.BandSession, s)
		if !on {
			for _, b := range common.AllBands() {
				bs := cfg.For(b)
				bs.Disabled = true
				cfg.Set(b, bs)
			}
		}

		e, _, _ := compactEngine(t, f, cfg)
		const turns = 40
		for i := 0; i < turns; i++ {
			say(e, 400)
			if i%4 == 3 {
				checkpoint(e)
			}
			e.maybeCompact()
		}

		r := result{
			bands:     map[common.Band]int{},
			convo:     ConversationBytes(e.Ctx),
			turns:     turns,
			budgetSes: cfg.For(common.BandSession).Budget,
			budgetCon: cfg.Conversation.Budget,
		}
		for _, b := range common.AllBands() {
			r.bands[b] = BandBytes(e.Ctx, b)
		}
		for _, ev := range e.Log.Events {
			switch ev.Type {
			case common.CompactorLaunched:
				r.folds++
			case common.BandPopulated:
				r.memories++
			}
		}
		return r
	}

	on := run(true)
	off := run(false)

	total := func(r result) int {
		n := r.convo
		for _, b := range common.AllBands() {
			n += r.bands[b]
		}
		return n
	}

	t.Logf("scripted session: %d turns, a checkpoint every 4 turns, 400 bytes of talk per turn", on.turns)
	t.Logf("budgets: conversation %d bytes (high %d), session band %d bytes (high %d)",
		on.budgetCon, on.budgetCon*2, on.budgetSes, on.budgetSes*2)
	t.Logf("")
	t.Logf("with memory ON:")
	for _, b := range common.AllBands() {
		if on.bands[b] > 0 {
			t.Logf("    band %-8s %6d bytes", b.String(), on.bands[b])
		}
	}
	t.Logf("    conversation %6d bytes", on.convo)
	t.Logf("    memories written: %d   cascade launches: %d", on.memories, on.folds)
	t.Logf("    TOTAL %d bytes", total(on))
	t.Logf("")
	t.Logf("with memory OFF (every band disabled):")
	t.Logf("    conversation %6d bytes", off.convo)
	t.Logf("    TOTAL %d bytes", total(off))
	t.Logf("")
	if total(off) > 0 {
		t.Logf("memory ON is %.1f%% the size of memory OFF over the same session",
			100*float64(total(on))/float64(total(off)))
	}
	t.Logf("cascade launches per checkpoint: %.2f", float64(on.folds)/float64(on.turns/4))
}
