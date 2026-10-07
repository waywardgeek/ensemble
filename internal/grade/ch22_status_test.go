package grade

import (
	"fmt"
	"strings"
	"testing"
)

// Sharpness tests for the status report assertions.
//
// The behavioural checks run the real binary, which means that when they
// pass, they pass for reasons that are expensive to inspect. These tests
// feed the assertions a hand-written report instead, so there is direct
// evidence that a wrong figure is actually caught. A check that has never
// been observed to fail is not evidence of anything.

// ch22Report renders a status report in the reference tree's wording. The
// wording is not what is being tested -- the numbers are -- but a fixture
// has to pick some wording.
func ch22Report(model string, last, tot ch22Usage, ratePct, cost float64) string {
	line := func(u ch22Usage) string {
		return fmt.Sprintf("%d in, %d out, %d cache read, %d cache write",
			u.in, u.out, u.cacheRead, u.cacheWrite)
	}
	return fmt.Sprintf(
		"model: %s\nlast response: %s\nsession total: %s\n"+
			"cache hit rate: %.1f%% (%d of %d input tokens served from cache)\n"+
			"session cost: ~$%.4f\n",
		model, line(last), line(tot), ratePct, tot.cacheRead, tot.in+tot.cacheRead, cost)
}

// ch22Scenario is the session the behavioural checks drive, with the first
// two turns billed to the dear model and the third to the cheap one.
func ch22Scenario() ch22Expect {
	return ch22Expected([]string{ch22ModelDear, ch22ModelDear, ch22ModelCheap, ch22ModelCheap})
}

func TestCh22StatusAcceptsACorrectReport(t *testing.T) {
	e := ch22Scenario()
	rep := ch22Report(e.model,
		ch22Usage{e.lastIn, e.lastOut, e.lastRead, e.lastWrite},
		ch22Usage{e.totIn, e.totOut, e.totRead, e.totWrite},
		e.cacheRatePct, e.costCorrect)

	if bad := ch22CheckStatus(rep, e); len(bad) > 0 {
		t.Fatalf("correct report rejected: %v", bad)
	}
	bad, broken := ch22CheckCost(rep, e)
	if broken != "" {
		t.Fatalf("grader reported itself broken on a valid scenario: %s", broken)
	}
	if len(bad) > 0 {
		t.Fatalf("correct cost rejected: %v", bad)
	}
}

// The bug this chapter fixes: the whole session priced at whichever model
// is current. If this scenario cannot tell that apart from the right
// answer, the fifteen points are decorative.
func TestCh22CostCatchesTheFlatRateBug(t *testing.T) {
	e := ch22Scenario()
	if fmt.Sprintf("%.4f", e.costCorrect) == fmt.Sprintf("%.4f", e.costIfOneRate) {
		t.Fatalf("scenario cannot distinguish the two formulas: both are $%.4f", e.costCorrect)
	}
	rep := ch22Report(e.model,
		ch22Usage{e.lastIn, e.lastOut, e.lastRead, e.lastWrite},
		ch22Usage{e.totIn, e.totOut, e.totRead, e.totWrite},
		e.cacheRatePct, e.costIfOneRate) // <- the bug

	bad, broken := ch22CheckCost(rep, e)
	if broken != "" {
		t.Fatalf("grader reported itself broken: %s", broken)
	}
	if len(bad) == 0 {
		t.Fatal("the flat-rate cost was accepted; the check cannot kill the bug it exists for")
	}
	if !strings.Contains(bad[0], "current") {
		t.Errorf("message does not name the fault:\n%s", bad[0])
	}
}

// A session that only ever used one model cannot distinguish per-model
// pricing from one flat rate. The grader must say so about ITSELF rather
// than award the points.
func TestCh22CostRefusesAVacuousSession(t *testing.T) {
	e := ch22Expected([]string{ch22ModelCheap, ch22ModelCheap, ch22ModelCheap})
	rep := ch22Report(e.model,
		ch22Usage{e.lastIn, e.lastOut, e.lastRead, e.lastWrite},
		ch22Usage{e.totIn, e.totOut, e.totRead, e.totWrite},
		e.cacheRatePct, e.costCorrect)

	bad, broken := ch22CheckCost(rep, e)
	if broken == "" {
		t.Fatalf("a single-model session was graded as if it proved something (bad=%v)", bad)
	}
}

// Each reported quantity must be load-bearing: corrupt it, and the check
// must notice. A report nobody checks is prose.
func TestCh22StatusCatchesEachWrongFigure(t *testing.T) {
	e := ch22Scenario()
	good := ch22Usage{e.lastIn, e.lastOut, e.lastRead, e.lastWrite}
	tot := ch22Usage{e.totIn, e.totOut, e.totRead, e.totWrite}

	cases := []struct {
		label string
		rep   string
		want  string
	}{
		{"wrong model", ch22Report("some-other-model", good, tot, e.cacheRatePct, e.costCorrect), "model"},
		{"wrong last-response input",
			ch22Report(e.model, ch22Usage{good.in + 7, good.out, good.cacheRead, good.cacheWrite},
				tot, e.cacheRatePct, e.costCorrect), "last response input"},
		{"wrong session total output",
			ch22Report(e.model, good, ch22Usage{tot.in, tot.out + 7, tot.cacheRead, tot.cacheWrite},
				e.cacheRatePct, e.costCorrect), "session total output"},
		{"wrong cache hit rate",
			ch22Report(e.model, good, tot, e.cacheRatePct+9, e.costCorrect), "cache hit rate"},
		{"report absent", "", "model"},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			bad := ch22CheckStatus(tc.rep, e)
			if len(bad) == 0 {
				t.Fatal("corrupted report accepted")
			}
			if !strings.Contains(strings.Join(bad, "\n"), tc.want) {
				t.Errorf("complaint does not name %q:\n%v", tc.want, bad)
			}
		})
	}
}

// The shipped skill file must declare the tool, or the feature is dead on
// arrival however well it is written -- chapter 21's lesson.
func TestCh22StatusToolIsDeclaredInTheShippedSkill(t *testing.T) {
	ok, detail := ch22StatusToolDeclared("../../agent")
	if !ok {
		t.Fatalf("reference tree: %s", detail)
	}
	t.Log(detail)

	if ok, _ := ch22StatusToolDeclared(t.TempDir()); ok {
		t.Fatal("an empty tree was reported as declaring the tool")
	}
}
