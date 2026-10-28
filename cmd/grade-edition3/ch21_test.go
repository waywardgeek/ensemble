package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestChapter21GuardsAndRoute(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "launched")
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("#!/bin/sh\nprintf '%s\\n' \"$*\" >> "+strconv.Quote(marker)+"\nexit 71\n"), 0700); err != nil {
		t.Fatal(err)
	}
	bad := [][]string{{"-ch", "21", target}, {"-ch", "21", "-gui", "", target}, {"-ch", "21", "-gui", target}}
	for _, flag := range []string{"core", "consumer", "consumer-case", "workflow", "media-models", "stream-exercise", "stream-model", "gui-model", "cache-prices", "cache-current", "cache-prior", "status-prices", "status-second-prices"} {
		bad = append(bad, []string{"-ch", "21", "-gui", target, "-" + flag, "", target})
	}
	for _, args := range bad {
		if code := run(args); code != 2 {
			t.Fatalf("%v exit %d", args, code)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatalf("invalid route launched: %v (%v)", args, err)
		}
	}
	out, err := os.CreateTemp(dir, "report")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	defer func() { os.Stdout = old }()
	os.Stdout = out
	code := run([]string{"-ch", "21", "-gui", target, "-json", target})
	os.Stdout = old
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Title    string
		MaxScore int `json:"max_score"`
		Checks   []struct{ ID string }
	}
	if err := json.Unmarshal(body, &report); err != nil {
		t.Fatalf("%v %s", err, body)
	}
	if code != 1 || !strings.Contains(report.Title, "Edition 3, Chapter 21") || report.MaxScore != 100 || len(report.Checks) != 7 {
		t.Fatalf("wrong route exit%d %s", code, body)
	}
	if body, err := os.ReadFile(marker); err != nil || !strings.Contains(string(body), "--port") {
		t.Fatalf("GUI route not attempted: %s %v", body, err)
	}
}
