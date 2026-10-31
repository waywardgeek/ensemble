package gradeedition3

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// These are external-provider fakes around real processes, not mocks of the
// client's HTTP behavior. Only the one known timeout expression is changed in
// scratch copies; the historical answer and its evidence remain untouched.
func TestGenerationLifetimeControls(t *testing.T) {
	for _, control := range []struct {
		name     string
		deadline string
		wantPass bool
	}{
		{"no_blanket_deadline", "0 * time.Second", true},
		{"blanket_deadline", "25 * time.Millisecond", false},
	} {
		t.Run(control.name, func(t *testing.T) {
			bin := buildTimeoutControl(t, control.deadline)
			for _, body := range []bool{false, true} {
				phase := "headers"
				if body {
					phase = "body"
				}
				t.Run(phase, func(t *testing.T) {
					check, err := checkGenerationLifetime(bin, body, 250*time.Millisecond)
					if err != nil {
						t.Fatal(err)
					}
					if check.ID != "generation_"+phase || check.Passed != control.wantPass {
						t.Fatalf("wrong lifetime verdict: %+v", check)
					}
					if !control.wantPass {
						detail := strings.Join(check.Details, "\n")
						// A build failure, zero API calls, or unrelated protocol fault
						// must not count as killing the timeout mutation.
						if !strings.Contains(detail, "got 0 complete answers") || strings.Contains(detail, "provider received") || strings.Contains(detail, "watchdog expired") {
							t.Fatalf("mutation failed for an unintended reason: %+v", check)
						}
					}
				})
			}
		})
	}
}

// Pin the actual long probe above the known 60-second carryover. The short
// mutation test proves sensitivity, not coverage of arbitrary long deadlines.
func TestHistoricalTimeoutIsExceeded(t *testing.T) {
	if generationDelay <= 60*time.Second {
		t.Fatalf("probe %s would miss the historical 60-second timeout", generationDelay)
	}
}

func buildTimeoutControl(t *testing.T, deadline string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(filepath.Join("..", "..", "edition-2", "solutions", "ch01"))); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "internal", "engine", "engine.go")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const anchor = "Timeout: 60 * time.Second"
	if strings.Count(string(source), anchor) != 1 {
		t.Fatal("historical control changed: expected exactly one known timeout expression")
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(source), anchor, "Timeout: "+deadline, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	bin, cleanup, err := grade.Build(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	return bin
}
