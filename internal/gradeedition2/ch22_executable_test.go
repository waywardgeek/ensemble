package gradeedition2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

func TestMemoryExecutableScope(t *testing.T) {
	checks := memoryExecutableChecks()
	total := 0
	for _, c := range checks {
		total += c.Points
		if !c.Passed || c.Earned != c.Points {
			t.Fatal("uninitialized check", c)
		}
		if c.ID == "micro-handoff-compresses" || c.ID == "forced-handoff" {
			t.Fatal("unadapted row credited")
		}
	}
	if len(checks) != 5 || total != 40 {
		t.Fatalf("scope %d rows/%d points", len(checks), total)
	}
}

func TestMemoryExecutableSettings(t *testing.T) {
	raw, err := memoryExecutableSettings(memorySettings(500, "8x"))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Version int `json:"version"`
		Memory  map[string]struct {
			Budget   int  `json:"budget"`
			Disabled bool `json:"disabled"`
		} `json:"memory"`
	}
	if err = json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	if document.Version != 1 || len(document.Memory) != 6 || document.Memory["session"].Budget != 500 || !document.Memory["archive8"].Disabled || document.Memory["session"].Disabled {
		t.Fatalf("wrong policy translation: %s", raw)
	}
	if _, err = memoryExecutableSettings(map[string]any{"memory": map[string]any{"8x": map[string]any{"disabled": "false"}}}); err == nil {
		t.Fatal("malformed fixture accepted")
	}
}

func TestMemoryExecutableOutcomes(t *testing.T) {
	launch := "{\"compression\":{\"task_id\":\"one\"}}\n"
	done := "{\"outcome\":{\"task_id\":\"one\",\"state\":\"committed\"}}\n"
	for _, tc := range []struct {
		text         string
		pending, bad bool
	}{
		{"", false, false}, {launch, true, false}, {launch + done, false, false},
		{launch + "{\"outcome\":", true, false}, // unfinished append is not a terminal fact
		{done, false, true}, {launch + done + launch, false, true}, {launch + launch, false, true}, {"not-json\n", false, true},
		{launch + "{\"outcome\":{\"task_id\":\"one\",\"state\":\"failed\"}}\n", false, true},
	} {
		pending, err := memoryExecutablePending([]byte(tc.text))
		if (err != nil) != tc.bad || (err == nil && pending != tc.pending) {
			t.Fatalf("%q: pending %v, error %v", tc.text, pending, err)
		}
	}
}

func TestMemoryExecutableFreshCorpusOnly(t *testing.T) {
	base, fresh := t.TempDir(), t.TempDir()
	for _, relative := range []string{"memory/20260101-1.md", "memory/20260101-2.md.retired", "children/compress-test/submit.md", "save.json"} {
		path := filepath.Join(base, "agents", "root", relative)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(memoryMemo(1)), 0600); err != nil {
			t.Fatal(err)
		}
	}
	calls := 0
	run := func(o memoryOpts) contextOut {
		calls++
		var files []string
		err := filepath.Walk(filepath.Join(o.dir, "agents", "root"), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				rel, _ := filepath.Rel(filepath.Join(o.dir, "agents", "root"), path)
				files = append(files, rel)
			}
			return nil
		})
		if err != nil || len(files) != 1 || files[0] != "memory/20260101-1.md" {
			t.Fatalf("copied %v: %v", files, err)
		}
		body, _ := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": memoryMemo(1)}}})
		return contextOut{reqs: []fakevendor.Recorded{{Body: body}}}
	}
	result := &contextResult{Errs: map[string]string{}}
	memoryExecutableFresh(run, base, fresh, result)
	if calls != 1 || result.Errs["fresh-start-populates"] != "" {
		t.Fatalf("calls %d, errors %v", calls, result.Errs)
	}
}
