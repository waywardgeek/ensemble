package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestChapter20GuardsAndRoute(t *testing.T) {
	dir := t.TempDir()
	core := filepath.Join(dir, "core")
	if err := os.MkdirAll(filepath.Join(core, "gui"), 0700); err != nil {
		t.Fatal(err)
	}
	for p, s := range map[string]string{"go.mod": "module fixture.invalid/core\n\ngo 1.25\n", "gui/go.mod": "module fixture.invalid/gui\n\ngo 1.25\n"} {
		if err := os.WriteFile(filepath.Join(core, p), []byte(s), 0600); err != nil {
			t.Fatal(err)
		}
	}
	marker := filepath.Join(dir, "launched")
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("#!/bin/sh\nprintf '%s\\n' \"$*\" >> "+strconv.Quote(marker)+"\nexit 71\n"), 0700); err != nil {
		t.Fatal(err)
	}
	bad := [][]string{{"-ch", "20", target}, {"-ch", "20", "-core", core, target}, {"-ch", "20", "-core", filepath.Join(dir, "missing"), "-gui", target, target}, {"-ch", "20", "-core", core, "-gui", target}}
	for _, flag := range []string{"consumer", "consumer-case", "workflow", "media-models", "stream-exercise", "stream-model", "gui-model", "cache-prices", "cache-current", "cache-prior", "status-prices", "status-second-prices"} {
		bad = append(bad, []string{"-ch", "20", "-core", core, "-gui", target, "-" + flag, "", target})
	}
	for _, args := range bad {
		if code := run(args); code != 2 {
			t.Fatalf("%v exit %d", args, code)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatalf("invalid route launched: %v", args)
		}
	}
	out, err := os.CreateTemp(dir, "report")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = out
	code := run([]string{"-ch", "20", "-core", core, "-gui", target, "-json", target})
	os.Stdout = old
	out.Close()
	body, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Title    string
		MaxScore int `json:"max_score"`
		Checks   []struct{ ID string }
	}
	if err = json.Unmarshal(body, &report); err != nil {
		t.Fatalf("%v %s", err, body)
	}
	if code != 1 || !strings.Contains(report.Title, "Chapter 20") || report.MaxScore != 200 || len(report.Checks) != 14 {
		t.Fatalf("wrong route exit%d %s", code, body)
	}
	launches, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	deployed := false
	for _, line := range strings.Split(string(launches), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "--port" {
			deployed = true
		}
	}
	if !deployed {
		t.Fatalf("consumer compile failure suppressed independent deployed-GUI probe: %s", launches)
	}
}
