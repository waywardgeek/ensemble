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
func TestArtifactReviewedRowsAreConservative(t *testing.T) {
	for _, tc := range []struct {
		name   string
		probes []ArtifactProbe
		failed bool
		want   int
	}{
		{"missing evidence", nil, false, 0},
		{"unrelated pass", []ArtifactProbe{{ID: "work-product-core", Passed: true}}, false, 0},
		{"incomplete Actions", []ArtifactProbe{{ID: "actions-browser", Passed: false}}, true, 0},
		{"complete Actions", []ArtifactProbe{{ID: "actions-browser", Passed: true}}, false, 30},
		{"only named viewer", []ArtifactProbe{{ID: "actions-browser", Passed: true}, {ID: "work-product-browser-unnamed=false", Passed: true}}, false, 30},
		{"only unnamed viewer", []ArtifactProbe{{ID: "actions-browser", Passed: true}, {ID: "work-product-browser-unnamed=true", Passed: true}}, false, 30},
		{"failed named viewer", []ArtifactProbe{{ID: "actions-browser", Passed: true}, {ID: "work-product-browser-unnamed=false", Passed: false}, {ID: "work-product-browser-unnamed=true", Passed: true}}, true, 30},
		{"failed unnamed viewer", []ArtifactProbe{{ID: "actions-browser", Passed: true}, {ID: "work-product-browser-unnamed=false", Passed: true}, {ID: "work-product-browser-unnamed=true", Passed: false}}, true, 30},
		{"complete reviewed rows", []ArtifactProbe{{ID: "actions-browser", Passed: true}, {ID: "work-product-browser-unnamed=false", Passed: true}, {ID: "work-product-browser-unnamed=true", Passed: true}}, false, 55},
		{"complete with independent failure", []ArtifactProbe{{ID: "actions-browser", Passed: true}, {ID: "work-product-browser-unnamed=false", Passed: true}, {ID: "work-product-browser-unnamed=true", Passed: true}, {ID: "timing-core", Passed: false}}, true, 55},
		{"viewers without Actions", []ArtifactProbe{{ID: "work-product-browser-unnamed=false", Passed: true}, {ID: "work-product-browser-unnamed=true", Passed: true}}, false, 0},
		{"independent failure", []ArtifactProbe{{ID: "actions-browser", Passed: true}, {ID: "timing-core", Passed: false}}, true, 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newArtifactReport()
			r.Probes = tc.probes
			if tc.failed {
				r.Status = "failed"
			}
			r.applyReviewedRows()
			r.applyReviewedRows() // Recomputing cannot double-count points.
			wantExit := 3
			if tc.failed {
				wantExit = 1
			}
			if r.Score != tc.want || r.ExitCode() != wantExit {
				t.Fatalf("unexpected report: %+v", r)
			}
			for _, row := range r.Rows {
				if (row.ID == "snippets" || row.ID == "interaction") && tc.want >= 30 {
					wantPoints := map[string]int{"snippets": 20, "interaction": 10}[row.ID]
					if row.Earned != wantPoints || row.Status != "passed" {
						t.Fatal(row)
					}
				} else if row.ID == "actions" && tc.want == 55 {
					if row.Earned != 25 || row.Status != "passed" {
						t.Fatal(row)
					}
				} else if row.Earned != 0 || row.Status != "ungraded" {
					t.Fatal(row)
				}
			}
		})
	}
}

func TestArtifactHostReceiptRequiresAllTerminalFacts(t *testing.T) {
	good := `{"requests":4,"returned":3,"offers":1,"failedOffers":1,"failure":null}`
	if err := checkArtifactHostReceipt([]byte(good)); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"requests", "returned", "offers", "failedOffers", "failure"} {
		for _, remove := range []bool{false, true} {
			var values map[string]any
			if err := json.Unmarshal([]byte(good), &values); err != nil {
				t.Fatal(err)
			}
			if remove {
				delete(values, field)
			} else if field == "failure" {
				values[field] = "fault"
			} else {
				values[field] = 0
			}
			bad, err := json.Marshal(values)
			if err != nil {
				t.Fatal(err)
			}
			if checkArtifactHostReceipt(bad) == nil {
				t.Fatalf("accepted incomplete terminal facts: %s", bad)
			}
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
	if r.ExitCode() != 3 || len(r.Probes) != 5 || r.Score != 55 {
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

func TestActionsReceiptsRequireEveryFact(t *testing.T) {
	good := map[string]any{"requests": 3, "executions": 1, "returned": 2, "running": 1, "done": 1, "streamRequests": 1, "streamCompleted": 1, "failure": nil, "streamFailure": nil, "stateRequests": 2, "stateExecutions": 2, "stateFailure": nil}
	data, _ := json.Marshal(good)
	if err := checkActionsHostReceipt(data); err != nil {
		t.Fatal(err)
	}
	for key, value := range good {
		delete(good, key)
		data, _ = json.Marshal(good)
		if checkActionsHostReceipt(data) == nil {
			t.Fatalf("missing %s accepted", key)
		}
		good[key] = value
	}
	encode := func(checks []string) []byte {
		data, _ := json.Marshal(map[string]any{"contract": "artifact-actions-v1", "passed": true, "checks": checks})
		return data
	}
	if err := checkActionsBrowserReceipt(encode(actionsBrowserChecks)); err != nil {
		t.Fatal(err)
	}
	for index, name := range actionsBrowserChecks {
		checks := append([]string{}, actionsBrowserChecks[:index]...)
		checks = append(checks, actionsBrowserChecks[index+1:]...)
		if checkActionsBrowserReceipt(encode(checks)) == nil {
			t.Fatalf("missing %s accepted", name)
		}
	}
}

func TestTimingCoreReceiptRequiresEveryCheck(t *testing.T) {
	encode := func(phase string, checks []string) []byte {
		data, err := json.Marshal(map[string]any{"contract": "artifact-timing-core-v1", "phase": phase, "passed": true, "checks": checks})
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	for _, phase := range []string{"writer", "reader"} {
		if err := checkTimingCoreReceipt(encode(phase, timingCoreChecks), phase); err != nil {
			t.Fatal(err)
		}
		for _, wrong := range []string{"", "wrong", "writer", "reader"} {
			if wrong != phase && checkTimingCoreReceipt(encode(wrong, timingCoreChecks), phase) == nil {
				t.Fatalf("%s accepted phase %q", phase, wrong)
			}
		}
		for i, name := range timingCoreChecks {
			checks := append([]string{}, timingCoreChecks[:i]...)
			checks = append(checks, timingCoreChecks[i+1:]...)
			if checkTimingCoreReceipt(encode(phase, checks), phase) == nil {
				t.Fatalf("%s accepted missing %s", phase, name)
			}
		}
		for _, data := range []string{`{}`, `{"contract":"artifact-timing-core-v1","passed":true}`, `{"contract":"wrong","passed":true}`, `{"contract":"artifact-timing-core-v1","passed":false}`} {
			if checkTimingCoreReceipt([]byte(data), phase) == nil {
				t.Fatalf("accepted incomplete receipt: %s", data)
			}
		}
	}
}
