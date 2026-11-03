package gradeedition3

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHostPatternFiles(t *testing.T) {
	directory := t.TempDir()
	pattern, err := WriteHostPattern(directory, 2048, 1152)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "pattern.json")
	parsed, err := ReadHostPattern(path)
	if err != nil || parsed != pattern {
		t.Fatalf("manifest mismatch: %v", err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := WriteHostPattern(directory, 2048, 1152); err == nil {
		t.Fatal("overwrote existing pattern")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("existing pattern changed")
	}
	html, err := os.ReadFile(filepath.Join(directory, "display.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(html), "http:") || strings.Contains(string(html), "https:") || !strings.Contains(string(html), "requestFullscreen") {
		t.Fatal("unexpected display fixture dependencies")
	}
}
func TestHostPatternManifestRefusals(t *testing.T) {
	valid := `{"prepared_width":64,"prepared_height":32,"colors":[[1,2,3],[4,5,6],[7,8,9],[10,11,12]]}`
	cases := []string{"null", "{}", valid + "{}", strings.Repeat(" ", 8193),
		strings.Replace(valid, `"prepared_width":64`, `"prepared_width":64,"prepared_width":64`, 1),
		strings.Replace(valid, `"prepared_width":64`, `"width":64`, 1),
		strings.Replace(valid, `[1,2,3]`, `[256,2,3]`, 1),
		strings.Replace(valid, `[1,2,3]`, `[1,2]`, 1),
		strings.Replace(valid, `[4,5,6]`, `[1,2,3]`, 1)}
	path := filepath.Join(t.TempDir(), "pattern.json")
	for _, text := range cases {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadHostPattern(path); err == nil {
			t.Fatalf("accepted invalid manifest %q", text)
		}
	}
}
