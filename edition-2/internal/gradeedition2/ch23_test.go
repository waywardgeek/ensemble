package gradeedition2

import (
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestChapter23Recall exercises the accepted GUI/Agent with actual files and
// provider HTTP. The full command separately measures the real 15-second timeout;
// focused regressions reuse that evidence instead of paying the delay each time.
func TestChapter23Recall(t *testing.T) {
	bin, clean, err := grade.Build("../../solutions/ch23/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer clean()
	r := recallChecks(bin, "cr", false)
	if r.Fatal != "" {
		t.Fatal(r.Fatal)
	}
	for id, e := range r.Errs {
		if e != "" {
			t.Errorf("%s: %s", id, e)
		}
	}
	if err := recallTopology("../../solutions/ch23", "../../solutions/ch23/gui"); err != "" {
		t.Error(err)
	}
}

// TestRecallCompaction exercises actual graduation and failed-attempt restarts,
// then checks retired-source recall and durable attachment bytes.
func TestRecallCompaction(t *testing.T) {
	bin, clean, err := grade.Build("../../solutions/ch23/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer clean()
	dir := t.TempDir()
	s := recallSuite{bin: bin, dir: dir, skills: filepath.Join(dir, "skills"), gui: dir, data: "cr"}
	if err := recallWrite(s.skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Recall compaction fixture\ntype: primary\ntools: read_file micro_handoff\n---\nFinish main work in chat.\n"}); err != nil {
		t.Fatal(err)
	}
	r := contextResult{Errs: map[string]string{}}
	recallCompaction(&s, &r)
	for id, err := range r.Errs {
		if err != "" {
			t.Errorf("%s: %s", id, err)
		}
	}
}
