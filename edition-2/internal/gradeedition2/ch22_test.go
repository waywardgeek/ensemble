package gradeedition2

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestChapter22Memory exercises the actual accepted GUI/Agent, owned settings,
// memory files, history and process lifetime. Only provider responses are fake.
func TestChapter22Memory(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch22/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	r := memoryChecks(bin)
	if r.Fatal != "" {
		t.Fatal(r.Fatal)
	}
	for id, e := range r.Errs {
		if e != "" {
			t.Errorf("%s: %s", id, e)
		}
	}
}

// TestMemoryBackground exercises the new scored scenario independently so a
// targeted runtime defect must fail that predicate rather than an older row.
func TestMemoryBackground(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch22/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if err = memoryBackground(bin, t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

// TestMemoryStale checks a real held graduation against a same-identity file edit
// and a completed off/on policy transition, preserving the exact source files.
func TestMemoryStale(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch22/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, edit := range []bool{false, true} {
		name := "policy"
		if edit {
			name = "source"
		}
		t.Run(name, func(t *testing.T) {
			if err := memoryStale(bin, t.TempDir(), edit); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestMemoryProtocolFailures exercises real bounded children against empty,
// wrong-type and syntactically complete provider-length-stop submissions.
func TestMemoryProtocolFailures(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch22/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	dir := t.TempDir()
	skills := filepath.Join(dir, "skills")
	if err := recallWrite(skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Memory failure fixture\ntype: primary\ntools: micro_handoff\n---\nFinish main work in chat.\n"}); err != nil {
		t.Fatal(err)
	}
	run := func(o memoryOpts) contextOut {
		o.model = "claude-sonnet-5"
		return memoryLaunch(bin, skills, dir, o)
	}
	for _, mode := range []string{"empty", "type", "length"} {
		t.Run(mode, func(t *testing.T) {
			if err := memoryProtocolFailure(run, t.TempDir(), mode); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestMemoryLifetime checks real explicit cancellation, owner shutdown and
// destination permission failure without losing the selected source or outcome.
func TestMemoryLifetime(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch22/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, mode := range []string{"kill", "shutdown", "write"} {
		t.Run(mode, func(t *testing.T) {
			if err := memoryLifetime(bin, t.TempDir(), mode); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestMemoryHeadroom checks actual near-window waiting, failure and interruption
// with a held worker and the owner's responsive public settings interface.
func TestMemoryHeadroom(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch22/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, mode := range []string{"success", "failure", "interrupt"} {
		t.Run(mode, func(t *testing.T) {
			if err := memoryHeadroom(bin, t.TempDir(), mode); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestMemoryAccounting checks measured successful and failed child usage through
// actual status tools, repeated main requests, and a fresh restored process.
func TestMemoryAccounting(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch22/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	dir := t.TempDir()
	skills := filepath.Join(dir, "skills")
	if err := recallWrite(skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Memory accounting fixture\ntype: primary\ntools: micro_handoff agent_status\n---\nFinish main work in chat.\n"}); err != nil {
		t.Fatal(err)
	}
	run := func(o memoryOpts) contextOut { return memoryLaunch(bin, skills, dir, o) }
	for _, failed := range []bool{false, true} {
		name := "success"
		if failed {
			name = "failure"
		}
		t.Run(name, func(t *testing.T) {
			accounting, replay := memoryAccounting(run, t.TempDir(), failed)
			if accounting != nil {
				t.Error(accounting)
			}
			if replay != nil {
				t.Error(replay)
			}
		})
	}
}

// TestMemoryRequestClassification keeps ordinary delegated submit traffic out of
// the memory route, including a task whose prose itself mentions memory.
func TestMemoryRequestClassification(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		memory     bool
	}{
		{"memory", `{"tools":[{"name":"submit","input_schema":{"type":"object","properties":{"memory":{"type":"string"}}}}]}`, true},
		{"ordinary", `{"messages":[{"content":"remember memory"}],"tools":[{"name":"submit","input_schema":{"type":"object","properties":{"result":{"type":"string"}}}}]}`, false},
		{"wrong type", `{"tools":[{"name":"submit","input_schema":{"properties":{"memory":{"type":"number"}}}}]}`, false},
		{"extra tool", `{"tools":[{"name":"submit","input_schema":{"properties":{"memory":{"type":"string"}}}},{"name":"read_file"}]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := memoryCompressor([]byte(tc.body)); got != tc.memory {
				t.Fatalf("memory classification = %v, want %v", got, tc.memory)
			}
		})
	}
}

// TestMemoryOutcomeBoundary checks actual snapshot/journal files: a truncated
// terminal line, main turn end or unrelated task cannot acknowledge a launch.
func TestMemoryOutcomeBoundary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	if err := os.WriteFile(path, []byte(`{"log":[{"type":"compactor_launched","compactor":{"task_id":"selected"}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, tail string
		pending    bool
	}{
		{"main end", "{\"type\":\"turn_ended\"}\n", true},
		{"other task", "{\"type\":\"compaction_outcome\",\"compaction_outcome\":{\"task_id\":\"other\",\"state\":\"committed\"}}\n", true},
		{"partial tail", `{"type":"compaction_outcome","compaction_outcome":{"task_id":"selected","state":"committed"}}`, true},
		{"complete", "{\"type\":\"compaction_outcome\",\"compaction_outcome\":{\"task_id\":\"selected\",\"state\":\"committed\"}}\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(path+".journal", []byte(tc.tail), 0600); err != nil {
				t.Fatal(err)
			}
			pending, err := memoryPending(dir)
			if err != nil || (len(pending) == 1 && pending[0] == "selected") != tc.pending {
				t.Fatalf("pending = %v, error = %v", pending, err)
			}
		})
	}
}
