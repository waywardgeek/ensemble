package gradeedition2

import (
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

func TestGatewayExecutableScope(t *testing.T) {
	result := grade.NewCh27Result()
	for _, id := range []string{"C2", "C3", "C4", "C7"} {
		result.Errs[id] = ""
	}
	// Neither a successful nor a failed check of the reference peer earns or
	// removes credit for the submitted server's four complete rows.
	result.Errs["C1"] = "reference peer crypto"
	result.Errs["C5"] = "reference peer callback"
	result.Errs["C6"] = "reference peer imports"
	checks := gatewayExecutableChecks(result)
	total := 0
	if len(checks) != 4 {
		t.Fatalf("got %d rows, want four", len(checks))
	}
	for _, c := range checks {
		total += c.Points
		if !c.Passed || c.Earned != c.Points {
			t.Fatalf("complete server row failed: %+v", c)
		}
	}
	if total != 60 {
		t.Fatalf("got %d points, want 60", total)
	}
	delete(result.Errs, "C7")
	for _, c := range gatewayExecutableChecks(result) {
		if c.ID == "C7" && (c.Passed || c.Earned != 0) {
			t.Fatal("unexercised child registration received credit")
		}
	}
	result.Fatal = "exercise aborted"
	for _, c := range gatewayExecutableChecks(result) {
		if c.Passed || c.Earned != 0 {
			t.Fatal("aborted exercise received credit")
		}
	}
}
