package grade

// Chapter 6's grader audit (course-policy P9).
//
// Every mutant below deletes ONE behavior from the reference solution and
// asserts the EXACT set of check ids that fails. A mutation that does not
// apply is a test failure, not a silent pass: an unapplied mutation scores
// 100 and manufactures a fake finding.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type ch6edit struct {
	// relPath is relative to the project root (e.g., "agent/internal/llm/actor.go").
	relPath string
	find    string // regexp, must match EXACTLY once
	replace string
}

type ch6mutation struct {
	name     string
	wantFail []string
	edits    []ch6edit
	why      string
}

func ch6ProjectRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	// ch6 grades the frozen snapshot, not the repo root.
	//
	// This used to return the repo root two levels up, which was correct until
	// 34ef50b ("Unify grading: every chapter grades one directory") moved the
	// exercise trees under solutions/. The root-level ch05/ and ch06/ that this
	// test copies have not existed since, so both TestCh6ReferenceScores100 and
	// TestCh6DeletionAudit have failed continuously while gradesweep.sh, which
	// already pointed at solutions/ch06, reported 100. That split is the whole
	// reason the TODO carried "ch6 reference scores 0/100" as a mystery.
	//
	// A chapter's score is only meaningful at its canonical target, and the
	// canonical target list lives in scripts/gradesweep.sh. Keep these in step.
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "solutions", "ch06")
	for _, sub := range []string{"agent", "ch05", "ch06"} {
		if _, err := os.Stat(filepath.Join(root, sub)); err != nil {
			t.Fatalf("ch6 target %s is missing %s/: %v\n"+
				"The grader copies these three directories; a missing one silently "+
				"produces an unbuildable tree and a 0/100 that looks like a real "+
				"architecture failure.", root, sub, err)
		}
	}
	return root
}

// copyDir recursively copies src to dst.
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			// Skip .git and vendor dirs.
			if e.Name() == ".git" || e.Name() == "vendor" || e.Name() == "node_modules" {
				continue
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			b, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, b, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

// buildCh6Mutant copies the relevant project dirs, applies edits, and builds.
func buildCh6Mutant(t *testing.T, m ch6mutation) string {
	t.Helper()
	root := ch6ProjectRoot(t)
	dir := t.TempDir()

	// Copy the dirs that the grader needs: agent/, ch05/, and ch06/.
	for _, sub := range []string{"agent", "ch05", "ch06"} {
		src := filepath.Join(root, sub)
		dst := filepath.Join(dir, sub)
		if err := copyDir(src, dst); err != nil {
			t.Fatalf("copy %s: %v", sub, err)
		}
	}

	// Apply edits.
	for _, ed := range m.edits {
		path := filepath.Join(dir, ed.relPath)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: mutation targets %s, which does not exist", m.name, ed.relPath)
		}
		re, err := regexp.Compile(ed.find)
		if err != nil {
			t.Fatalf("%s: bad pattern %q: %v", m.name, ed.find, err)
		}
		hits := re.FindAllIndex(b, -1)
		if len(hits) != 1 {
			t.Fatalf("MUTATION DID NOT LAND: %s: pattern %q matched %d times in %s, want exactly 1. "+
				"An unapplied mutation scores 100 and manufactures a fake finding.",
				m.name, ed.find, len(hits), ed.relPath)
		}
		mutated := re.ReplaceAll(b, []byte(ed.replace))
		if string(mutated) == string(b) {
			t.Fatalf("MUTATION DID NOT LAND: %s: replacing %q in %s changed no bytes", m.name, ed.find, ed.relPath)
		}
		if err := os.WriteFile(path, mutated, 0o644); err != nil {
			t.Fatalf("%s: write %s: %v", m.name, ed.relPath, err)
		}
	}

	// Verify the agent module compiles.
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = filepath.Join(dir, "agent")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s: mutant agent does not compile: %v\n%s", m.name, err, out)
	}

	return dir
}

func scoreCh6(t *testing.T, dir string) (total int, failed []string) {
	t.Helper()
	// Ch6Run wants the AGENT MODULE directory, not the snapshot root: it runs
	// DiscoverBase and Build against what it is given, and the snapshot root
	// has no go.mod. scripts/gradesweep.sh already passes solutions/ch06/agent.
	// Passing the root here made the build fail, which zeroed every check at
	// once and read as a catastrophic architecture failure rather than a bad
	// path. The exercise dirs ch05/ and ch06/ still have to be copied as
	// siblings, which is why dir is the root everywhere else in this file.
	res, err := Ch6Run(filepath.Join(dir, "agent"))
	if err != nil {
		t.Fatalf("Ch6Run: %v", err)
	}
	for _, c := range Ch6Evaluate(res) {
		total += c.Earned
		if !c.Passed {
			failed = append(failed, c.ID)
		}
	}
	sort.Strings(failed)
	return total, failed
}

// --- the mutants -----------------------------------------------------------

func ch6Mutants() []ch6mutation {
	return []ch6mutation{
		// ---- not-deaf: hint delivery during tool execution ---------------
		{
			name: "hint-not-emitted-as-observation",
			why: "The actor processes the hint (attaches it to the engine) but never " +
				"emits an observation for it, so the grader never sees the hint in the " +
				"observation stream.",
			wantFail: []string{"not-deaf"},
			edits: []ch6edit{{
				relPath: "agent/internal/llm/actor.go",
				find:    `a\.notify\(common\.PartDelta\{\n\t\tPartID: atomic\.AddUint64\(&a\.partSeq, 1\),\n\t\tChunk:  "hint:" \+ m\.Text,\n\t\}\)`,
				replace: `_ = m.Text // observation suppressed`,
			}},
		},
		{
			name: "hints-ignored-in-waitForTools",
			why: "The mailbox drain in waitForTools handles ToolCompleted but treats " +
				"Hint as an unrecognized type, silently discarding it.",
			wantFail: []string{"not-deaf"},
			edits: []ch6edit{{
				relPath: "agent/internal/llm/actor.go",
				find:    `case common\.Hint:\n\t\t\t\t\ta\.handleHint\(m\)`,
				replace: `case common.Hint:
					_ = m // ignored`,
			}},
		},

		// ---- replay-is-live: event log -----------------------------------
		{
			name: "no-event-log",
			why: "The exercise pipeline runs correctly but writes no event log, so " +
				"the replay check finds nothing to compare, and the log-derived " +
				"hint and wake checks lose their evidence. " +
				"NOT observer-fires: the why-text here used to claim the grader " +
				"'cannot verify observer output' without the log, and that is " +
				"false. Observer output travels on the agent's STDOUT and is read " +
				"by parseObservations into r.StateChanges and r.ContentObs " +
				"(ch06_harness.go), which this mutant never touches. The claim " +
				"went unchallenged because this audit could not run at all between " +
				"34ef50b and 2026-09-27. observer-fires keeps its own dedicated " +
				"mutant, no-state-change-notifications, which kills it exactly and " +
				"alone, so nothing is left unaudited by correcting this set.",
			wantFail: []string{"not-deaf", "replay-is-live", "wake-once"},
			edits: []ch6edit{{
				relPath: "ch06/main.go",
				find:    `if p := os\.Getenv\("CH06_LOG"\); p != "" \{`,
				replace: `if false {`,
			}},
		},

		// ---- observer-fires: state changes + content --------------------
		{
			name: "no-state-change-notifications",
			why: "The actor changes state internally but never notifies observers, " +
				"so no state_changed observations reach the grader.",
			wantFail: []string{"observer-fires"},
			edits: []ch6edit{{
				relPath: "agent/internal/llm/actor.go",
				find:    `func \(a \*Actor\) setState\(s common\.TurnState\) \{\n\ta\.mu\.Lock\(\)\n\told := a\.state\n\ta\.state = s\n\ta\.mu\.Unlock\(\)\n\tif old != s \{\n\t\ta\.notify\(common\.StateChanged\{From: old, To: s\}\)\n\t\}\n\}`,
				replace: `func (a *Actor) setState(s common.TurnState) {
	a.mu.Lock()
	a.state = s
	a.mu.Unlock()
}`,
			}},
		},
		// ---- wake-once: multi-agent -----------------------------------
		{
			name: "single-agent-pipeline",
			why: "The exercise uses only one agent for all three steps instead of " +
				"three separate agents, so observations come from a single agent ID.",
			wantFail: []string{"replay-is-live", "wake-once"},
			edits: []ch6edit{{
				relPath: "ch06/main.go",
				find:    `fw\.Add\(editorName,`,
				replace: `fw.Add(authorName, /* was editor */`,
			}, {
				relPath: "ch06/main.go",
				find:    `fw\.Add\(reviewerName,`,
				replace: `fw.Add(authorName, /* was reviewer */`,
			}},
		},

		// ---- loud-refusal: model validation ----------------------------
		{
			name: "no-model-validation",
			why: "Turn() skips the LookupModel check, so unknown models are accepted " +
				"silently and the request goes to the wire.",
			wantFail: []string{"loud-refusal"},
			edits: []ch6edit{{
				relPath: "agent/internal/llm/engine.go",
				find:    `if _, known := common\.LookupModel\(e\.Cfg\.Model\); !known \{`,
				replace: `if false { // model validation removed`,
			}},
		},

		// ---- hub-clean: star topology ---------------------------------
		{
			name: "spoke-imports-spoke",
			why: "internal/llm imports internal/tools directly, breaking the star " +
				"topology where spokes should only import internal/common.",
			wantFail: []string{"ch5-parity", "hub-clean"},
			edits: []ch6edit{{
				relPath: "agent/internal/llm/engine.go",
				find:    `"github\.com/waywardgeek/ensemble/agent/internal/common"`,
				replace: `"github.com/waywardgeek/ensemble/agent/internal/common"
	_ "github.com/waywardgeek/ensemble/agent/internal/tools"`,
			}},
		},
	}
}

// --- the tests -------------------------------------------------------------

func TestCh6ReferenceScores100(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the reference solution")
	}
	root := ch6ProjectRoot(t)
	total, failed := scoreCh6(t, root)
	if total != 100 || len(failed) > 0 {
		t.Fatalf("reference scored %d/100, failing %v; want 100 and nothing failing", total, failed)
	}
}

func TestCh6PointsSumTo100(t *testing.T) {
	res := &Ch6Result{}
	sum := 0
	seen := map[string]bool{}
	for _, c := range Ch6Evaluate(res) {
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
	if len(seen) != 7 {
		t.Fatalf("got %d checks, want 7", len(seen))
	}
}

// TestCh6DeletionAudit: delete each protected behavior from the REFERENCE
// and confirm the score drops by exactly the checks that should notice.
func TestCh6DeletionAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and grades one mutant per case")
	}
	names := map[string]bool{}
	for _, m := range ch6Mutants() {
		if names[m.name] {
			t.Fatalf("duplicate mutant name %q", m.name)
		}
		names[m.name] = true
	}
	for _, m := range ch6Mutants() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			dir := buildCh6Mutant(t, m)
			total, failed := scoreCh6(t, dir)
			t.Logf("AUDIT %-40s score=%3d/100 failing=%v", m.name, total, failed)

			want := append([]string{}, m.wantFail...)
			sort.Strings(want)
			if strings.Join(failed, ",") != strings.Join(want, ",") {
				t.Fatalf("%s\nwhy: %s\nfailing checks = %v\nwant            = %v\nscore %d/100",
					m.name, m.why, failed, want, total)
			}
			if len(want) == 0 {
				t.Fatalf("%s: every Chapter 6 mutant deletes a graded behavior; an empty wantFail is a mistake here", m.name)
			}
			if total >= 100 {
				t.Fatalf("%s: score did not drop (%d/100). A check that cannot fail is not a check.", m.name, total)
			}
		})
	}
}
