package grade_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/waywardgeek/coding-agents-course/internal/grade"
)

// repoRoot walks up from the test's working directory to the module root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find module root")
		}
		dir = parent
	}
}

func runAgainst(t *testing.T, pkgDir string) grade.Report {
	t.Helper()
	bin, cleanup, err := grade.Build(pkgDir)
	if err != nil {
		t.Fatalf("building %s: %v", pkgDir, err)
	}
	defer cleanup()
	res, err := grade.Run(bin)
	if err != nil {
		t.Fatalf("running %s: %v", pkgDir, err)
	}
	return grade.NewReport(grade.Evaluate(res), res.Stderr)
}

func failedIDs(r grade.Report) []string {
	var out []string
	for _, c := range r.Checks {
		if !c.Passed {
			out = append(out, c.ID)
		}
	}
	sort.Strings(out)
	return out
}

// TestReferenceSolutionPasses is the positive control: if this fails, every
// other assertion in this file is meaningless.
func TestReferenceSolutionPasses(t *testing.T) {
	os.Unsetenv("COURSE_MUTATION")
	r := runAgainst(t, filepath.Join(repoRoot(t), "solutions", "ch01"))
	if !r.Passed {
		t.Errorf("reference solution failed checks %v", failedIDs(r))
		for _, c := range r.Checks {
			if !c.Passed {
				t.Errorf("  %s: %s", c.ID, strings.Join(c.Details, "; "))
			}
		}
	}
	if r.Score != r.MaxScore {
		t.Errorf("score %d/%d, want full marks", r.Score, r.MaxScore)
	}
}

// TestMutationsAreCaught is the sensitivity proof. Each mutation breaks one
// property; the assertion is on the EXACT set of checks that fail, because a
// grader that fails everything on any defect is as useless as one that fails
// nothing — it cannot tell the student what is wrong.
func TestMutationsAreCaught(t *testing.T) {
	mutantDir := filepath.Join(repoRoot(t), "testdata", "students", "mutant")

	cases := []struct {
		mutation string
		wantFail []string
		why      string
	}{
		{"none", nil,
			"the mutant with no defect must pass, or the defects prove nothing"},
		{"amnesiac", []string{"growth", "memory"},
			"a fresh single-message call per round: no conversation exists"},
		{"useronly", []string{"memory", "wire"},
			"appends questions but never the model's replies: roles stop alternating"},
		{"nousage", []string{"usage"},
			"never reports the bill"},
		{"fakeusage", []string{"usage"},
			"invents usage numbers instead of summing the responses"},
		{"noversion", []string{"wire"},
			"omits the anthropic-version header"},
		{"nomaxtokens", []string{"wire"},
			"omits max_tokens, which the API will not guess"},
		{"chatty", []string{"protocol"},
			"writes diagnostics to stdout, corrupting the protocol stream"},
		{"twocalls", []string{"calls", "replies"},
			"two API calls per round: the answers come from the wrong response"},
		{"fabricate", []string{"calls", "growth", "memory", "replies", "usage", "wire"},
			"never calls the API at all"},
		{"firstblock", []string{"memory", "replies"},
			"reads content[0].text instead of walking the block list and concatenating"},
		{"hardkey", []string{"wire"},
			"hardcodes an API key instead of reading ANTHROPIC_API_KEY"},
	}

	for _, tc := range cases {
		t.Run(tc.mutation, func(t *testing.T) {
			t.Setenv("COURSE_MUTATION", tc.mutation)
			r := runAgainst(t, mutantDir)

			got := failedIDs(r)
			want := append([]string(nil), tc.wantFail...)
			sort.Strings(want)

			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Errorf("mutation %q (%s)\n  failing checks = %v\n  want           = %v",
					tc.mutation, tc.why, got, want)
				for _, c := range r.Checks {
					if !c.Passed {
						t.Logf("  %s: %s", c.ID, strings.Join(c.Details, " | "))
					}
				}
			}
		})
	}
}
