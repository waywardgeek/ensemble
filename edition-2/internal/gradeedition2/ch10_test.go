package gradeedition2

import (
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestChapter10Persistence runs both actual application clients. The provider
// alone is fake; files, restart, reducer replay and exit saving remain real.
func TestChapter10Persistence(t *testing.T) {
	core, err := filepath.Abs("../../solutions/edition-3/ch10")
	if err != nil {
		t.Fatal(err)
	}
	bin, cleanup, err := grade.Build(filepath.Join(core, "cmd", "ch01"))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, c := range persistenceChecks(bin, filepath.Join(core, "gui")) {
		if !c.Passed {
			t.Errorf("%s: %v", c.ID, c.Details)
		}
	}
}
