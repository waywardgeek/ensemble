package gradeedition2

import (
	"reflect"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// Run the real accepted program against the original external-provider fake.
// Changing one captured header then isolates the correction from legacy checks.
func TestChapter1VersionAndPreservedChecks(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch01")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	res, err := grade.Run(bin)
	if err != nil {
		t.Fatal(err)
	}
	baseline := Chapter1(res)
	if !baseline.Passed || baseline.Score != 100 || !reflect.DeepEqual(baseline.Checks, grade.Evaluate(res)) {
		t.Fatalf("unchanged checks failed: %+v", baseline)
	}
	for _, version := range []string{"wrong-version", ""} {
		res.Records[0].Header.Set("anthropic-version", version)
		report := Chapter1(res)
		if report.Passed || report.Score != 85 {
			t.Fatalf("version %q: score=%d, passed=%v", version, report.Score, report.Passed)
		}
		for i, check := range report.Checks {
			if check.ID == "wire" {
				if check.Passed || !strings.Contains(strings.Join(check.Details, " "), "expected 2023-06-01") {
					t.Fatalf("missing exact-version diagnosis: %+v", check)
				}
			} else if !reflect.DeepEqual(check, baseline.Checks[i]) {
				t.Fatalf("version correction changed unrelated check: %+v", check)
			}
		}
	}
}
