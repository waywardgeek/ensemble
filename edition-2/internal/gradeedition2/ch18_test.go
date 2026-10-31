package gradeedition2

import (
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestChapter18Combined drives both shipped clients against actual HTTP peers.
// It targets the accepted restart snapshot and independently supplied Sonnet 4.6
// rates; the middle-model control remains deliberately unpriced.
func TestChapter18Combined(t *testing.T) {
	core := "../../solutions/edition-3/ch18"
	cli, cleanCLI, err := grade.Build(core + "/cmd/ch01")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanCLI()
	gui, cleanGUI, err := grade.Build(core + "/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanGUI()
	report, err := Chapter18Targets(cli, gui, core, [4]float64{3, 3.75, .3, 15}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range report.Checks {
		if !c.Passed {
			t.Errorf("%s: %v", c.ID, c.Details)
		}
	}
}
