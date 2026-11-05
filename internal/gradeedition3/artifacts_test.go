package gradeedition3

import (
	"encoding/json"
	"os"
	"testing"
)

func TestArtifactIncompleteReportDoesNotPass(t *testing.T) {
	r := newArtifactReport()
	r.Probes = []ArtifactProbe{{ID: "partial", Passed: true}}
	if r.ExitCode() != 3 || r.Status != "ungraded" || r.Score != 0 || r.MaxScore != 100 {
		t.Fatalf("partial probes became a grade: %+v", r)
	}
	total := 0
	for _, row := range r.Rows {
		total += row.Points
		if row.Status != "ungraded" || row.Earned != 0 {
			t.Fatal(row)
		}
	}
	if total != 100 {
		t.Fatal(total)
	}
	r.Status = "failed"
	if r.ExitCode() != 1 {
		t.Fatal("failed probe exit")
	}
}
func TestArtifactHostReceiptRequiresAllTerminalFacts(t *testing.T) {
	good := `{"requests":3,"returned":2,"offers":1,"failure":null}`
	if err := checkArtifactHostReceipt([]byte(good)); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{`{}`, `{"requests":3,"returned":2,"offers":1}`, `{"requests":4,"returned":2,"offers":1,"failure":null}`, `{"requests":3,"returned":2,"offers":0,"failure":null}`, `{"requests":3,"returned":2,"offers":1,"failure":"fault"}`} {
		if checkArtifactHostReceipt([]byte(bad)) == nil {
			t.Fatalf("accepted incomplete terminal facts: %s", bad)
		}
	}
}
func TestArtifactCoreReceiptCannotClaimBareSuccess(t *testing.T) {
	data, _ := json.Marshal(map[string]any{"contract": "artifact-work-product-core-v1", "passed": true, "checks": []string{}})
	if checkArtifactCoreReceipt(data) == nil {
		t.Fatal("empty core receipt accepted")
	}
}
func TestArtifactExternalCandidate(t *testing.T) {
	candidate := os.Getenv("ARTIFACT_RUNTIME")
	if candidate == "" {
		t.Skip("explicit runtime and existing tools required")
	}
	r, err := GradeArtifacts(ArtifactGradeConfig{candidate, os.Getenv("ARTIFACT_DOTNET"), os.Getenv("ARTIFACT_NODE"), os.Getenv("ARTIFACT_PLAYWRIGHT")})
	if err != nil {
		t.Fatal(err)
	}
	if r.ExitCode() != 3 || len(r.Probes) != 3 {
		t.Fatalf("unexpected partial report: %+v", r)
	}
	for _, p := range r.Probes {
		if !p.Passed {
			t.Fatalf("probe failed: %+v", p)
		}
	}
}

func TestArtifactBrowserReceiptRequiresEveryCheck(t *testing.T) {
	encode := func(checks []string) []byte {
		data, err := json.Marshal(map[string]any{
			"contract": "artifact-browser-work-product-v1", "passed": true, "checks": checks,
		})
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	if err := checkArtifactBrowserReceipt(encode(artifactBrowserChecks)); err != nil {
		t.Fatal(err)
	}
	if checkArtifactBrowserReceipt(encode(nil)) == nil {
		t.Fatal("bare success accepted")
	}
	for index, name := range artifactBrowserChecks {
		checks := append([]string{}, artifactBrowserChecks[:index]...)
		checks = append(checks, artifactBrowserChecks[index+1:]...)
		if checkArtifactBrowserReceipt(encode(checks)) == nil {
			t.Fatalf("missing %s accepted", name)
		}
	}
}
