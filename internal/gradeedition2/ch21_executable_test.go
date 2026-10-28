package gradeedition2

import (
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

func TestChapter21ExecutableScope(t *testing.T) {
	checks := grade.Ch26Checks(grade.Ch26Result{})
	if checks[1].ID != "ensemble-persists" || checks[5].ID != "gui-tree-message" || checks[2].ID != "ownership-enforced" || checks[3].ID != "wait-change-wakes" {
		t.Fatal("score row identity drifted")
	}
	total := 0
	for _, i := range []int{0, 2, 3, 4, 6} {
		total += checks[i].Points
	}
	if total != 70 || checks[1].Points+checks[5].Points != 30 {
		t.Fatal("Chapter 21 scope drifted")
	}
	for _, skills := range []string{"", "relative", filepath.Join(t.TempDir(), "missing")} {
		if _, err := Chapter21Executable("must-not-launch", skills); err == nil {
			t.Fatalf("accepted invalid scope %q", skills)
		}
	}
}
func TestChapter21ExecutableResults(t *testing.T) {
	for _, reply := range []c21Reply{{text: "ordinary prose"}, {text: `{"activity":"idle"}`, bad: true}} {
		if _, err := c21OK(reply, nil); err == nil {
			t.Fatal("invalid result accepted")
		}
	}
	row, err := c21OK(c21Reply{text: "job 7, status: done, bytes: 20\n{\"historical\":true}\n"}, nil)
	if err != nil || row["historical"] != true {
		t.Fatal("completed result lost")
	}
	for _, text := range []string{"{\"activity\":\"idle\"}", "job 2, status: done, bytes: 5\n{}"} {
		control := &c21Controls{}
		result, err := control.finish(c21Reply{text: text})
		if err != nil || result.text != text {
			t.Fatal("completed output incorrectly polled")
		}
	}
}
