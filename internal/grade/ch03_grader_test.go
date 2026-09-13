package grade

// Chapter 3's grader audit (course-policy P9).
//
// Two different tests live here, and the brief is explicit that only the
// second one has ever found a real hole:
//
//   - Mutating a deliberately-broken student proves the CHECKS fire.
//   - Deleting behavior from the REFERENCE proves the CHAPTER'S PROMISES are
//     collected.
//
// Everything below is the second kind. Each mutant removes one behavior from
// the reference solution and asserts the EXACT set of check ids that fails. A
// mutant whose expected set is empty is a deliberate claim that the variant is
// ACCEPTABLE — that is how the declined decision is proved to accept all three
// answers rather than quietly preferring one.
//
// A mutation that fails to apply is a test failure, not a silent pass. A
// silently-unapplied mutation scores 100 and manufactures a fake finding.

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

// --- the mutation rig ------------------------------------------------------

type ch3edit struct {
	file    string
	find    string // regexp, must match EXACTLY once
	replace string
}

type ch3mutation struct {
	name string
	// wantFail is the exact set of check ids expected to fail. Empty means the
	// mutant must still score 100: an accepted variant, not a missed bug.
	wantFail []string
	edits    []ch3edit
	why      string
}

func ch3ReferenceDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "solutions", "ch03")
}

// buildCh3Mutant copies the reference solution, applies the edits, and builds
// it. Every edit must match exactly once, so a pattern that has drifted out of
// date fails loudly instead of producing a green run against unmutated code.
func buildCh3Mutant(t *testing.T, m ch3mutation) string {
	t.Helper()
	ref := ch3ReferenceDir(t)
	dir := t.TempDir()

	entries, err := os.ReadDir(ref)
	if err != nil {
		t.Fatalf("read reference: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(ref, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), b, 0o644); err != nil {
			t.Fatalf("write %s: %v", e.Name(), err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module mutant\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	for _, ed := range m.edits {
		path := filepath.Join(dir, ed.file)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: mutation targets %s, which does not exist", m.name, ed.file)
		}
		re, err := regexp.Compile(ed.find)
		if err != nil {
			t.Fatalf("%s: bad pattern %q: %v", m.name, ed.find, err)
		}
		hits := re.FindAllIndex(b, -1)
		if len(hits) != 1 {
			t.Fatalf("MUTATION DID NOT LAND: %s: pattern %q matched %d times in %s, want exactly 1. "+
				"An unapplied mutation scores 100 and manufactures a fake finding.",
				m.name, ed.find, len(hits), ed.file)
		}
		if err := os.WriteFile(path, re.ReplaceAll(b, []byte(ed.replace)), 0o644); err != nil {
			t.Fatalf("%s: write %s: %v", m.name, ed.file, err)
		}
	}

	bin := filepath.Join(dir, "mutant-bin")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s: mutant does not compile: %v\n%s", m.name, err, out)
	}
	return bin
}

// scoreCh3 runs the whole Chapter 3 grader against a binary.
func scoreCh3(t *testing.T, bin string) (total int, failed []string) {
	t.Helper()
	res, err := Ch3Run(bin)
	if err != nil {
		t.Fatalf("Ch3Run: %v", err)
	}
	for _, c := range Ch3Evaluate(res) {
		total += c.Earned
		if !c.Passed {
			failed = append(failed, c.ID)
		}
	}
	sort.Strings(failed)
	return total, failed
}

// --- the mutants -----------------------------------------------------------

func ch3Mutants() []ch3mutation {
	return []ch3mutation{
		// ---- Chapter 1's deferred type filter, collected here -------------
		{
			name: "blind-walk-no-type-filter",
			why: "Chapter 1 promised that filtering content blocks by `type` starts paying here. " +
				"This walks every block as if it were text, which is the naive implementation the promise warns about.",
			wantFail: []string{"ch2parity", "editcontract", "localtools", "multiblock", "runcommand", "toolerror", "toolloop"},
			edits: []ch3edit{{
				file: "anthropic.go", find: `switch h\.Type \{`, replace: `switch "text" {`,
			}},
		},
		{
			name:     "tool-use-blocks-ignored",
			why:      "An implementation that only looks at text blocks never sees the call and silently does nothing.",
			wantFail: []string{"ch2parity", "editcontract", "localtools", "multiblock", "runcommand", "toolerror", "toolloop"},
			edits: []ch3edit{{
				file: "anthropic.go", find: `case "tool_use":`, replace: `case "tool_use_never_matches":`,
			}},
		},
		{
			name:     "text-blocks-dropped",
			why:      "The other half of dispatching on type: the prose in a mixed message must be recorded too.",
			wantFail: []string{"ch2parity", "multiblock", "toolerror", "toolloop"},
			edits: []ch3edit{{
				file: "anthropic.go", find: `case "text":`, replace: `case "text_never_matches":`,
			}},
		},

		// ---- the loop ------------------------------------------------------
		{
			name:     "loop-runs-one-round",
			why:      "A tool call is the middle of a turn, not the end. This executes the tools but never goes back.",
			wantFail: []string{"editcontract", "localtools", "multiblock", "runcommand", "toolerror", "toolloop"},
			edits: []ch3edit{{
				file: "engine.go", find: `for round := 0; ; round\+\+ \{`, replace: `for round := 0; round < 1; round++ {`,
			}},
		},
		{
			name: "results-keyed-by-position",
			why: "Answers the first outstanding call every time. With one call outstanding this is indistinguishable " +
				"from correct, which is exactly why the multiblock fixture issues two.",
			wantFail: []string{"multiblock"},
			edits: []ch3edit{
				{file: "engine.go", find: `for _, call := range calls \{`, replace: `for range calls {`},
				{file: "engine.go", find: `if err := e\.Execute\(call\); err != nil \{`, replace: `if err := e.Execute(calls[0]); err != nil {`},
			},
		},

		// ---- failure is a result ------------------------------------------
		{
			name:     "tool-error-result-dropped",
			why:      "The single most common way a student's agent locks up: a failure that returns nothing at all.",
			wantFail: []string{"editcontract", "toolerror"},
			edits: []ch3edit{{
				file: "engine.go", find: `isError, out = true, err\.Error\(\)`, replace: `return nil`,
			}},
		},
		{
			name:     "tool-error-not-marked",
			why:      "The result comes back, but the model is not told it was a failure.",
			wantFail: []string{"editcontract", "toolerror"},
			edits: []ch3edit{{
				file: "engine.go", find: `isError, out = true, err\.Error\(\)`, replace: `isError, out = false, err.Error()`,
			}},
		},
		{
			name:     "unknown-tool-panics",
			why:      "A tool that does not exist must be an error result, not a crash.",
			wantFail: []string{"toolerror"},
			edits: []ch3edit{{
				file:    "tools.go",
				find:    `return "", fmt\.Errorf\("no such tool %q; available tools: %s",\n\t\t\tname, strings\.Join\(ToolNames\(\), ", "\)\)`,
				replace: `panic("no such tool " + name)`,
			}},
		},
		{
			name: "nonzero-exit-marked-as-error",
			why: "NEGATIVE CONTROL for toolerror. A command that exits 7 RAN; calling that a tool error is a false " +
				"claim about the call. Without this mutant, marking everything as an error would score full marks.",
			wantFail: []string{"runcommand", "toolerror"},
			edits: []ch3edit{{
				file: "tools.go", find: `exitCode = ee\.ExitCode\(\)`,
				replace: `return "", fmt.Errorf("run_command: exit %d", ee.ExitCode())`,
			}},
		},

		// ---- run_command ---------------------------------------------------
		{
			name:     "exit-code-not-reported",
			why:      "The exit code never reaches the model.",
			wantFail: []string{"runcommand", "toolerror"},
			edits: []ch3edit{{
				file: "tools.go", find: `fmt\.Fprintf\(&b, "exit_code: %d\\n", exitCode\)`, replace: `_ = exitCode`,
			}},
		},
		{
			name:     "stderr-dropped",
			why:      "Half of what a failing command tells you is on stderr.",
			wantFail: []string{"runcommand"},
			edits: []ch3edit{{
				file: "tools.go", find: `fmt\.Fprintf\(&b, "stderr:\\n%s", stderr\.String\(\)\)`, replace: `fmt.Fprintf(&b, "stderr:")`,
			}},
		},

		// ---- the local file tools, one at a time ---------------------------
		{
			name:     "read-range-ignored",
			why:      "read_file's line range is the whole point of the tool: it decides how much context the turn costs.",
			wantFail: []string{"localtools", "multiblock", "toolloop"},
			edits: []ch3edit{{
				file: "tools.go", find: `if a\.StartLine > 0 \|\| a\.EndLine > 0 \{`, replace: `if false {`,
			}},
		},
		{
			name:     "write-file-writes-nothing",
			why:      "Reports success, writes an empty file. Only the disk can catch this.",
			wantFail: []string{"localtools"},
			edits: []ch3edit{{
				file: "tools.go", find: `os\.WriteFile\(a\.Path, \[\]byte\(a\.Content\), 0o644\)`,
				replace: `os.WriteFile(a.Path, []byte(""), 0o644)`,
			}},
		},
		{
			name:     "edit-file-does-not-change-the-file",
			why:      "Reports the edit, writes the original bytes back.",
			wantFail: []string{"localtools"},
			edits: []ch3edit{{
				file: "tools.go", find: `updated := strings\.Replace\(text, a\.OldText, a\.NewText, 1\)`,
				replace: `updated := text`,
			}},
		},
		{
			name:     "list-directory-hides-files",
			why:      "An agent that cannot see the tree guesses at paths.",
			wantFail: []string{"localtools"},
			edits: []ch3edit{{
				file: "tools.go", find: `fmt\.Fprintf\(&b, "file  %s \(%d bytes\)\\n", e\.Name\(\), size\)`, replace: `_ = size`,
			}},
		},
		{
			name:     "search-files-finds-nothing",
			why:      "An agent that cannot grep cannot find what to read.",
			wantFail: []string{"localtools"},
			edits: []ch3edit{{
				file: "tools.go", find: `if re\.MatchString\(line\) \{`, replace: `if false && re.MatchString(line) {`,
			}},
		},

		// ---- the declined decision (P6) ------------------------------------
		// The reference REFUSES. These two mutants switch it to the other two
		// defensible answers and assert the score does NOT move. If either of
		// them ever fails a check, a preference has leaked into the grader.
		{
			name:     "editfile-answer-fuzzy-match",
			why:      "DECLINED DECISION: fuzzy-match must be accepted exactly as readily as refusing.",
			wantFail: nil,
			edits: []ch3edit{{
				file: "tools.go", find: `(?s)case n == 0:.*?case n > 1:`,
				replace: `case n == 0:
		lines := strings.Split(text, "\n")
		lines[0] = a.NewText
		if err := os.WriteFile(a.Path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
			return "", err
		}
		return fmt.Sprintf("edit_file: exact anchor not found in %s; applied the closest match instead", a.Path), nil
	case n > 1:`,
			}},
		},
		{
			name:     "editfile-answer-rewrite",
			why:      "DECLINED DECISION: falling back to rewriting the file must also be accepted.",
			wantFail: nil,
			edits: []ch3edit{{
				file: "tools.go", find: `(?s)case n == 0:.*?case n > 1:`,
				replace: `case n == 0:
		if err := os.WriteFile(a.Path, []byte(a.NewText), 0o644); err != nil {
			return "", err
		}
		return fmt.Sprintf("edit_file: exact anchor not found in %s; rewrote the file", a.Path), nil
	case n > 1:`,
			}},
		},
		{
			name: "editfile-claims-success-changes-nothing",
			why: "The one outcome NO answer defends: the model is told the edit worked and the file is untouched. " +
				"This is what stops editcontract from being satisfied by any string at all.",
			wantFail: []string{"editcontract"},
			edits: []ch3edit{{
				file: "tools.go", find: `(?s)case n == 0:.*?case n > 1:`,
				replace: `case n == 0:
		return "ok", nil
	case n > 1:`,
			}},
		},
		{
			name: "editfile-refusal-says-nothing",
			why: "A refusal that does not say what it saw is only half a loud failure: it declines to guess, " +
				"and then costs a round trip to find out why.",
			wantFail: []string{"editcontract"},
			edits: []ch3edit{{
				file: "tools.go", find: `(?s)case n == 0:.*?case n > 1:`,
				replace: `case n == 0:
		return "", fmt.Errorf("edit failed")
	case n > 1:`,
			}},
		},

		// ---- the Chapter 2 regression guard --------------------------------
		{
			name:     "ch2-usage-accounting-broken",
			why:      "ch2parity must be able to fail, or it is a decoration rather than a regression guard.",
			wantFail: []string{"ch2parity"},
			edits: []ch3edit{{
				file: "anthropic.go", find: `CacheRead:  resp\.Usage\.CacheReadTokens,`, replace: `CacheRead:  0,`,
			}},
		},
		{
			name: "tool-results-not-spliced-first",
			why: "Anthropic requires the tool_result to come FIRST in the content array of the message answering it. " +
				"No scripted session can exercise this, because no message of theirs mixes a tool_result with " +
				"anything else — a prompt cannot arrive while the loop is blocked on a tool, which is Chapter 5's " +
				"mailbox. It is graded on a RENDER of the exhibit log, whose tool return is followed by a human turn. " +
				"Before that render was added this mutant scored 100/100 and nothing anywhere caught it.",
			wantFail: []string{"toolloop"},
			edits: []ch3edit{{
				file: "anthropic.go", find: `appendBlocks\("user", blocks, true\)`, replace: `appendBlocks("user", blocks, false)`,
			}},
		},
	}
}

// --- the tests -------------------------------------------------------------

func TestCh3ReferenceScores100(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the reference solution")
	}
	ref := ch3ReferenceDir(t)
	bin := filepath.Join(t.TempDir(), "ch03")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = ref
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build reference: %v\n%s", err, out)
	}
	total, failed := scoreCh3(t, bin)
	if total != 100 || len(failed) > 0 {
		t.Fatalf("reference scored %d/100, failing %v; want 100 and nothing failing", total, failed)
	}
}

func TestCh3PointsSumTo100(t *testing.T) {
	// Deliberately summed from the checks themselves rather than from the
	// outline's table, so the code cannot drift away from its own arithmetic.
	res := &Ch3Result{Loop: map[string]*Ch3Session{}}
	sum := 0
	seen := map[string]bool{}
	for _, c := range Ch3Evaluate(res) {
		if seen[c.ID] {
			t.Errorf("duplicate check id %q", c.ID)
		}
		seen[c.ID] = true
		sum += c.Points
	}
	if sum != 100 {
		t.Fatalf("checks sum to %d, want exactly 100", sum)
	}
	if len(seen) != 7 {
		t.Fatalf("got %d checks, want 7", len(seen))
	}
}

// TestCh3DeletionAudit is the test the brief calls for: delete each protected
// behavior from the REFERENCE and confirm the score drops.
//
// A row reading 100 -> 100 where a drop was expected is the finding, and it
// fails the test rather than being written off as a quirk.
func TestCh3DeletionAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and grades one mutant per case")
	}
	for _, m := range ch3Mutants() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			bin := buildCh3Mutant(t, m)
			total, failed := scoreCh3(t, bin)
			t.Logf("AUDIT %-38s score=%3d/100 failing=%v", m.name, total, failed)

			want := append([]string{}, m.wantFail...)
			sort.Strings(want)

			if strings.Join(failed, ",") != strings.Join(want, ",") {
				t.Fatalf("%s\nwhy: %s\nfailing checks = %v\nwant            = %v\nscore %d/100",
					m.name, m.why, failed, want, total)
			}
			if len(want) == 0 {
				if total != 100 {
					t.Fatalf("%s: accepted variant scored %d/100; it must be worth exactly as much as the reference answer", m.name, total)
				}
				return
			}
			if total >= 100 {
				t.Fatalf("%s: score did not drop (%d/100). A check that cannot fail is not a check.", m.name, total)
			}
		})
	}
}
