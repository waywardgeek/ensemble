package gradeedition2

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestGateway26 runs only when an actual isolated submission is selected. It
// exercises built processes and outside consumers, including mutant copies.
func TestGateway26(t *testing.T) {
	core := os.Getenv("GATEWAY_SUBMISSION")
	if core == "" {
		t.Skip("set GATEWAY_SUBMISSION to exercise a real submitted Gateway")
	}
	core, err := filepath.Abs(core)
	if err != nil {
		t.Fatal(err)
	}
	bin, cleanup, err := grade.Build(filepath.Join(core, "cmd", "gateway"))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	field := os.Getenv("GATEWAY_WELCOME")
	if field == "" {
		field = "QueueDepth"
	}
	report, err := Chapter26(bin, core, field)
	if err != nil {
		t.Fatal(err)
	}
	report.WriteText(os.Stdout)
	if !report.Passed {
		t.Fail()
	}
}
