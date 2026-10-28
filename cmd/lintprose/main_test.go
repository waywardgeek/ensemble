package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintAcceptsAnyChapterLength(t *testing.T) {
	// Each paragraph avoids unrelated style warnings. These cases fall below,
	// inside, and well above the former 4,000–7,500-word range.
	const paragraph = "Bill checks the output before deciding which implementation details the reader needs to understand.\n\n"
	for _, tc := range []struct {
		name       string
		paragraphs int
	}{
		{"very short", 1},
		{"formerly in range", 300},
		{"formerly over maximum", 600},
		{"very long", 10000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "chapter-01.md")
			if err := os.WriteFile(path, []byte(strings.Repeat(paragraph, tc.paragraphs)), 0600); err != nil {
				t.Fatal(err)
			}
			results, err := lint(path, false)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range results {
				if r.name == "words" {
					t.Errorf("chapter length should not produce a diagnostic: %+v", r)
				}
				if r.failed {
					t.Errorf("unexpected diagnostic for clean prose: %+v", r)
				}
			}
		})
	}
}

func TestLintStillChecksStyle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chapter-01.md")
	if err := os.WriteFile(path, []byte("Bill checks the output—but the process still needs a clear explanation.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	results, err := lint(path, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.name == "em-dashes" {
			if !r.failed || !r.hard || r.count != 1 {
				t.Fatalf("expected the existing em-dash check to fail: %+v", r)
			}
			return
		}
	}
	t.Fatal("missing em-dash diagnostic")
}
