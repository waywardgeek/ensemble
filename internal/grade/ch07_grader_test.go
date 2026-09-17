package grade

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// Chapter 7 deletion audit (policy P9).
//
// The question this file answers is not "do the checks pass?" but "would the
// checks NOTICE if the behavior were gone?" A check that cannot fail is not a
// check, and the only way to find out is to delete the behavior from the
// REFERENCE and watch the score drop by exactly the expected amount.
//
// Chapter 7's audit carries one mutant that matters more than the rest:
// per-chunk-part-ids. Streaming is easy to fake. A submission that emits a
// fresh PartID for every chunk still streams text, still streams thinking,
// still streams tool arguments, and still produces a final answer identical
// to the reference. Everything a human would eyeball looks right. The only
// thing it destroys is the ability to correlate a run of deltas with the part
// they became, which is the one property the chapter exists to teach and the
// one property a GUI cannot live without. If deltas-match-final does not fail
// under that mutant, the check is decoration.

type ch7edit struct {
	// relPath is relative to the project root (e.g., "agent/internal/llm/claude.go").
	relPath string
	find    string // regexp, must match EXACTLY once
	replace string
}

type ch7mutation struct {
	name     string
	wantFail []string
	edits    []ch7edit
	why      string
}

func ch7ProjectRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	// thisFile is in internal/grade/; project root is two levels up.
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// buildCh7Mutant copies the reference into a temp dir and applies the edits.
//
// The exactly-once requirement on every pattern is the load-bearing part. A
// regexp written from memory that matches zero times produces a mutant
// identical to the reference, which then scores 100 and reports a PASSING
// audit for a behavior nobody deleted. That failure mode is silent, it is
// easy, and it has already happened once on this book: the Chapter 6 coder
// wrote three patterns against remembered code and they matched nothing.
// So a pattern that does not land is a hard failure here, not a skip.
func buildCh7Mutant(t *testing.T, m ch7mutation) string {
	t.Helper()
	root := ch7ProjectRoot(t)
	dir := t.TempDir()

	// ch6-parity re-runs Chapter 6's grader, which drives ch06/ and reads
	// ch05/ for its own parity check, so the mutant needs all four trees.
	for _, sub := range []string{"agent", "ch05", "ch06", "ch07"} {
		if err := copyDir(filepath.Join(root, sub), filepath.Join(dir, sub)); err != nil {
			t.Fatalf("copy %s: %v", sub, err)
		}
	}

	for _, e := range m.edits {
		p := filepath.Join(dir, e.relPath)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("mutant %s: read %s: %v", m.name, e.relPath, err)
		}
		re, err := regexp.Compile(e.find)
		if err != nil {
			t.Fatalf("mutant %s: bad regexp %q: %v", m.name, e.find, err)
		}
		if n := len(re.FindAllIndex(b, -1)); n != 1 {
			t.Fatalf("mutant %s: pattern %q matched %d times in %s; want exactly 1.\n"+
				"THE MUTATION DID NOT LAND, so this audit would prove nothing. Fix the pattern against the real source.",
				m.name, e.find, n, e.relPath)
		}
		if err := os.WriteFile(p, re.ReplaceAll(b, []byte(e.replace)), 0o644); err != nil {
			t.Fatalf("mutant %s: write %s: %v", m.name, e.relPath, err)
		}
	}
	return dir
}

func scoreCh7(t *testing.T, dir string) (total int, failed []string) {
	t.Helper()
	res, err := Ch7Run(dir)
	if err != nil {
		t.Fatalf("Ch7Run: %v", err)
	}
	for _, c := range Ch7Evaluate(res) {
		total += c.Earned
		if !c.Passed {
			failed = append(failed, c.ID)
			// Log why, not just which. A mutant whose blast radius is
			// wider than expected is usually telling you something true
			// about the code, and the detail is how you find out whether
			// the extra failure is real or accidental.
			t.Logf("  %-22s %s", c.ID, firstOr(c.Details, "no detail"))
		}
	}
	sort.Strings(failed)
	return total, failed
}

// --- the mutants -----------------------------------------------------------

func ch7Mutants() []ch7mutation {
	return []ch7mutation{
		// ---- the whole chapter: ask for the stream at all ----------------
		{
			name: "no-stream-flag",
			why: "Render never sets stream:true, so the vendor answers with one " +
				"JSON blob. Note what survives: the final text is identical, the " +
				"parts are identical, the event log is identical. Only the SHAPE " +
				"of arrival changed. Every delta-counting check should notice and " +
				"nothing else should, which is the chapter's claim that streaming " +
				"is a delivery property and not a content property.",
			wantFail: []string{"deltas-match-final", "delivery-not-content", "stream-deltas", "thinking-streamed", "tool-params-streamed"},
			edits: []ch7edit{{
				relPath: "agent/internal/llm/claude.go",
				find:    `Stream:\s+common\.StreamingFor\(cfg\) != 0,`,
				replace: `Stream:    false,`,
			}},
		},

		// ---- the mutant that justifies the chapter -----------------------
		{
			name: "per-chunk-part-ids",
			why: "Text deltas carry an id derived from the chunk instead of the id " +
				"of the block they belong to, so a fresh id effectively arrives with " +
				"every chunk. The stream still flows, every kind still arrives, the " +
				"answer is still correct, and the finalized parts are byte-identical " +
				"to the reference. What breaks is correlation: no run of deltas can " +
				"be tied to the part it became. A grader that concatenated every text " +
				"delta in the turn and compared it to the final message would pass " +
				"this mutant happily, which is why deltas-match-final groups BY id.\n" +
				"Note the shape of this edit. An earlier version of this mutant used a " +
				"package-level counter, which was a better story but a worse test: it " +
				"also tripped Chapter 5's no-mutable-globals check through ch6-parity, " +
				"so the mutant failed two checks and only one of them was about " +
				"streaming. A mutation must delete exactly one behavior.",
			wantFail: []string{"deltas-match-final"},
			edits: []ch7edit{{
				relPath: "agent/internal/llm/claude.go",
				find:    `cb\.Delta\(partID\(ev\.Index\), common\.DeltaText, ev\.Delta\.Text\)`,
				replace: `cb.Delta(partID(ev.Index)*1000+uint64(len(ev.Delta.Text)), common.DeltaText, ev.Delta.Text)`,
			}},
		},

		// ---- ids must not cross kinds ----------------------------------------
		{
			name: "cross-kind-part-ids",
			why: "Thinking deltas are assigned the next block's part id, so " +
				"they share an id with the text block. The stream still flows, " +
				"every kind still arrives, the answer is still correct, and " +
				"deltas-match-final passes because it groups text by kind before " +
				"reassembly. What breaks is the cross-kind invariant: a part id " +
				"must belong to exactly one kind.\n" +
				"The author predicted (§7.2 ruling) that the per-chunk-part-ids " +
				"mutant would fail more checks once this invariant was enforced. " +
				"That prediction was false: per-chunk-part-ids only perturbs TEXT " +
				"delta ids, and the per-kind checks are indifferent to it. This " +
				"dedicated mutant exists because the deletion audit requires every " +
				"enforced behaviour to have a mutant that deletes it.",
			wantFail: []string{"thinking-streamed"},
			edits: []ch7edit{{
				relPath: "agent/internal/llm/claude.go",
				find:    `cb\.Delta\(partID\(ev\.Index\), common\.DeltaThinking, ev\.Delta\.Thinking\)`,
				replace: `cb.Delta(partID(ev.Index+1), common.DeltaThinking, ev.Delta.Thinking)`,
			}},
		},

		// ---- reasoning is a kind, not an afterthought --------------------
		{
			name: "no-thinking-deltas",
			why: "thinking_delta events are consumed and dropped rather than " +
				"reported. The reasoning block still arrives in the final parts, " +
				"so the transcript is complete; it simply arrives all at once at " +
				"the end, which is exactly the experience streaming is supposed to " +
				"remove.\n" +
				"Only thinking-streamed notices, and that is worth knowing: " +
				"deltas-match-final walks the TEXT parts, so it is indifferent to " +
				"the reasoning stream. The two checks cover different kinds on " +
				"purpose rather than by accident.",
			wantFail: []string{"thinking-streamed"},
			edits: []ch7edit{{
				relPath: "agent/internal/llm/claude.go",
				find:    `cb\.Delta\(partID\(ev\.Index\), common\.DeltaThinking, ev\.Delta\.Thinking\)`,
				replace: `_ = ev.Delta.Thinking // deltas dropped`,
			}},
		},

		// ---- tool arguments are the hardest kind to stream ---------------
		{
			name: "no-tool-arg-deltas",
			why: "Partial tool-call JSON is accumulated for the final part but never " +
				"reported as it arrives. This is the kind most often skipped in real " +
				"implementations, because partial JSON is not parseable and looks " +
				"useless. It is not useless: it is how a GUI shows which file the " +
				"agent is about to edit before the call is complete.",
			wantFail: []string{"tool-params-streamed"},
			edits: []ch7edit{{
				relPath: "agent/internal/llm/claude.go",
				find:    `cb\.Delta\(partID\(ev\.Index\), common\.DeltaToolCall, ev\.Delta\.PartialJSON\)`,
				replace: `_ = ev.Delta.PartialJSON // deltas dropped`,
			}},
		},

		// ---- everything Chapter 6 promised still has to work -------------
		{
			name: "no-model-validation",
			why: "Turn() stops refusing unknown models. Chapter 7 touched Render " +
				"and Parse in all three vendors, which is exactly the blast radius " +
				"where a Chapter 6 guarantee would get quietly dropped. Parity is " +
				"the check that notices.",
			wantFail: []string{"ch6-parity"},
			edits: []ch7edit{{
				relPath: "agent/internal/llm/engine.go",
				find:    `if _, known := common\.LookupModel\(e\.Cfg\.Model\); !known \{`,
				replace: `if false { // model validation removed`,
			}},
		},
	}
}

// --- the tests -------------------------------------------------------------

func TestCh7ReferenceScores100(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the reference solution")
	}
	root := ch7ProjectRoot(t)
	total, failed := scoreCh7(t, root)
	if total != 100 || len(failed) > 0 {
		t.Fatalf("reference scored %d/100, failing %v; want 100 and nothing failing", total, failed)
	}
}

func TestCh7PointsSumTo100(t *testing.T) {
	res := &Ch7Result{}
	sum := 0
	seen := map[string]bool{}
	for _, c := range Ch7Evaluate(res) {
		if seen[c.ID] {
			t.Errorf("duplicate check id %q", c.ID)
		}
		seen[c.ID] = true
		sum += c.Points
		if c.Points <= 0 {
			t.Errorf("check %q carries %d points; a zero-point check is invisible to the deletion audit", c.ID, c.Points)
		}
	}
	if sum != 100 {
		t.Fatalf("checks sum to %d, want exactly 100", sum)
	}
	if len(seen) != 6 {
		t.Fatalf("got %d checks, want 6", len(seen))
	}
}

// TestCh7DeletionAudit: delete each protected behavior from the REFERENCE
// and confirm the score drops by exactly the checks that should notice.
func TestCh7DeletionAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and grades one mutant per case")
	}
	names := map[string]bool{}
	for _, m := range ch7Mutants() {
		if names[m.name] {
			t.Fatalf("duplicate mutant name %q", m.name)
		}
		names[m.name] = true
	}
	for _, m := range ch7Mutants() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			dir := buildCh7Mutant(t, m)
			total, failed := scoreCh7(t, dir)
			t.Logf("AUDIT %-40s score=%3d/100 failing=%v", m.name, total, failed)

			want := append([]string{}, m.wantFail...)
			sort.Strings(want)
			if strings.Join(failed, ",") != strings.Join(want, ",") {
				t.Fatalf("%s\nwhy: %s\nfailing checks = %v\nwant            = %v\nscore %d/100",
					m.name, m.why, failed, want, total)
			}
			if len(want) == 0 {
				t.Fatalf("%s: every Chapter 7 mutant deletes a graded behavior; an empty wantFail is a mistake here", m.name)
			}
			if total >= 100 {
				t.Fatalf("%s: score did not drop (%d/100). A check that cannot fail is not a check.", m.name, total)
			}
		})
	}
}
