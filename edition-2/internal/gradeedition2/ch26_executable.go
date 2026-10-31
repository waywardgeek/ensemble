package gradeedition2

import (
	"os"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter26Executable runs the original outside peer against the submitted
// server executable. peer is an explicitly selected independent Go reference,
// NOT the submitted C# library. Consequently its own crypto, callback/receipt
// behavior and import graph cannot earn C1, C5 or C6. The full scenario is retained
// because C3 includes both empty and nonempty authenticated welcome counts.
func Chapter26Executable(bin, peer string) (grade.Report, error) {
	work, err := os.MkdirTemp("", "c26-executable-")
	if err != nil {
		return grade.Report{}, err
	}
	result := grade.NewCh27Result()
	if err = gatewayExercise(bin, peer, work, "QueueDepth", &result); err != nil {
		result.Fatal = err.Error()
	}
	checks := gatewayExecutableChecks(result)
	for i := range checks {
		checks[i].Details = append(checks[i].Details,
			"Independent Go peer; only submitted server behavior credited.",
			"retained artifacts: "+work)
	}
	return grade.NewTitledReport("Edition 3, Chapter 26 — C# executable server scope (60 points; C1 Noise wrapper, C5 complete client delivery contract and C6 public client boundary excluded)", checks, ""), nil
}

func gatewayExecutableChecks(result grade.Ch27Result) []grade.Check {
	var checks []grade.Check
	for _, check := range grade.Ch27Checks(result) {
		switch check.ID {
		case "C2", "C3", "C4", "C7":
			if check.ID == "C3" {
				check.Title = "Actual Gateway binary: authenticated identities and empty/nonempty welcome queue counts"
			}
			checks = append(checks, check)
		}
	}
	return checks
}
