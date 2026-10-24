package ensemble

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCh10R1DefaultAgentPaths(t *testing.T) {
	root := New(nil)
	defer root.Shutdown()
	var paths []string
	for i := 0; i < 2; i++ {
		// Unnamed Agents still derive saves, but each owns independent state.
		dir := t.TempDir()
		a, err := root.NewAgent(Config{Model: "fixture", DataDir: dir, Tools: []string{}})
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, a.Config().SavePath)
		if err = a.Save(); err != nil {
			t.Errorf("fresh Agent save: %v", err)
		}
	}
	if paths[0] == paths[1] {
		t.Fatal("shared default save")
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Error(err)
		}
	}
}
func TestCh10R1DiagnosticAppendDoesNotTrim(t *testing.T) {
	path := filepath.Join(t.TempDir(), "api.log")
	d, err := OpenDiagnosticLog(path, func() int { return 2 })
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"one\n", "two\n", "three\n"} {
		if _, err = d.Write([]byte(s)); err != nil {
			t.Fatal(err)
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(b), "\n") != 3 {
		t.Fatalf("append destructively retained before recovery boundary: %q", b)
	}
}

func TestCh10R1DiagnosticRetentionBoundaries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "api.log")
	limit := 2
	if err := os.WriteFile(path, []byte("one\ntwo\nthree\n"), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := OpenDiagnosticLog(path, func() int { return limit })
	if err != nil {
		t.Fatal(err)
	}
	if err = d.Retain(); err != nil {
		t.Fatal(err)
	}
	if _, err = d.Write([]byte("four\n")); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "two\nthree\nfour\n" {
		t.Fatalf("append offset after replacement: %q %v", b, err)
	}
	limit = 0
	if err = d.Retain(); err != nil {
		t.Fatal(err)
	}
	zero, _ := os.ReadFile(path)
	if string(zero) != string(b) {
		t.Fatal("zero pruned")
	}
	limit = -1
	if err = d.Retain(); err == nil {
		t.Fatal("accepted negative limit")
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(b) {
		t.Fatal("invalid policy pruned")
	}
}
