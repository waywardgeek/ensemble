package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestInvalidInputsFailBeforeGrading(t *testing.T) {
	root := t.TempDir()
	module := filepath.Join(root, "module")
	if err := os.Mkdir(module, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte("module example.org/fixture\n\ngo 1.25\n"), 0600); err != nil {
		t.Fatal(err)
	}
	badCase := filepath.Join(root, "bad-case.json")
	if err := os.WriteFile(badCase, []byte("not JSON"), 0600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(root, "missing")
	marker := filepath.Join(root, "invoked")
	target := filepath.Join(root, "sentinel")
	if err := os.WriteFile(target, []byte("#!/bin/sh\nprintf invoked >> "+strconv.Quote(marker)+"\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	cases := [][]string{
		{}, {""}, {"-ch", "2", target}, {"-ch", "17", target}, {"-ch", "4"},
		{"-ch", "1", "-core", module, target},
		{"-ch", "1", "-consumer", "", target},
		{"-ch", "1", "-consumer-case", "", target},
		{"-ch", "4", target},
		{"-ch", "4", "-core", module, target},
		{"-ch", "4", "-consumer", module, target},
		{"-ch", "4", "-core", missing, "-consumer", module, target},
		{"-ch", "4", "-core", module, "-consumer", root, target},
		{"-ch", "4", "-core", module, "-consumer", module, "-consumer-case", badCase, target},
		{"-ch", "4", "-core", module, "-consumer", module, "-consumer-case", missing, target},
		{"-ch", "4", "-core", module, "-consumer", module, missing},
		{"-ch", "1", missing}, {"-ch", "1", target, target},
		{"-unknown", target},
	}
	for _, args := range cases {
		t.Run("args="+strings.Join(args, " "), func(t *testing.T) {
			got := run(args)
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("invalid input launched executable: marker error %v", err)
			}
			if got != 2 {
				t.Fatalf("run(%q) = %d, want 2", args, got)
			}
		})
	}
}

func TestHelpSucceedsWithoutLaunchingTarget(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "invoked")
	target := filepath.Join(root, "sentinel")
	if err := os.WriteFile(target, []byte("#!/bin/sh\nprintf invoked >> "+strconv.Quote(marker)+"\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"-h"}, {"-help"}, {"-ch", "1", "-h", target}, {"-ch", "4", "-h", target}, {"-ch", "5", "-h", target}, {"-ch", "6", "-h", target}, {"-ch", "7", "-h", target}, {"-ch", "8", "-h", target}, {"-ch", "9", "-h", target}, {"-ch", "10", "-h", target}, {"-ch", "11", "-h", target}, {"-ch", "12", "-h", target}, {"-ch", "13", "-h", target}, {"-ch", "14", "-h", target}} {
		if got := run(args); got != 0 {
			t.Errorf("run(%q) = %d, want 0", args, got)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatalf("help launched executable: marker error %v", err)
		}
	}
}

func TestChapter5ValidationBeforeLaunch(t *testing.T) {
	dir := t.TempDir()
	module := filepath.Join(dir, "module")
	if err := os.Mkdir(module, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte("module example.test/core\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	invalidJSON := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalidJSON, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	ran := filepath.Join(dir, "target-ran")
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("#!/bin/sh\nprintf ran > '"+ran+"'\nexit 99\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		flags []string
	}{
		{"executable wrong chapter", []string{"-ch", "4", "-executable-only"}},
		{"executable false", []string{"-ch", "5", "-executable-only=false"}},
		{"executable with core", []string{"-ch", "5", "-executable-only", "-core", module}},
		{"executable with workflow", []string{"-ch", "5", "-executable-only", "-workflow", module}},
		{"executable with unrelated flag", []string{"-ch", "5", "-executable-only", "-stream-model", "fake"}},
		{"executable bad media", []string{"-ch", "5", "-executable-only", "-media-models", invalidJSON}},
		{"executable missing media", []string{"-ch", "5", "-executable-only", "-media-models", filepath.Join(dir, "absent.json")}},
		{"missing core", []string{"-ch", "5", "-consumer", module, "-workflow", module}},
		{"missing consumer", []string{"-ch", "5", "-core", module, "-workflow", module}},
		{"missing workflow", []string{"-ch", "5", "-core", module, "-consumer", module}},
		{"workflow not module", []string{"-ch", "5", "-core", module, "-consumer", module, "-workflow", dir}},
		{"workflow missing", []string{"-ch", "5", "-core", module, "-consumer", module, "-workflow", filepath.Join(dir, "absent")}},
		{"bad media JSON", []string{"-ch", "5", "-core", module, "-consumer", module, "-workflow", module, "-media-models", invalidJSON}},
		{"missing media JSON", []string{"-ch", "5", "-core", module, "-consumer", module, "-workflow", module, "-media-models", filepath.Join(dir, "absent.json")}},
		{"bad consumer JSON", []string{"-ch", "5", "-core", module, "-consumer", module, "-workflow", module, "-consumer-case", invalidJSON}},
		{"ch1 workflow", []string{"-ch", "1", "-workflow="}},
		{"ch1 media", []string{"-ch", "1", "-media-models="}},
		{"ch4 workflow", []string{"-ch", "4", "-core", module, "-consumer", module, "-workflow="}},
		{"ch4 media", []string{"-ch", "4", "-core", module, "-consumer", module, "-media-models="}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if code := run(append(tc.flags, target)); code != 2 {
				t.Fatalf("exit=%d, want 2", code)
			}
			if _, err := os.Stat(ran); !os.IsNotExist(err) {
				t.Fatalf("invalid input launched target: %v", err)
			}
		})
	}
}

func TestChapter6ValidationBeforeLaunch(t *testing.T) {
	dir := t.TempDir()
	module := filepath.Join(dir, "module")
	if err := os.Mkdir(module, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte("module example.test/core\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	badJSON := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badJSON, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	badModels := filepath.Join(dir, "bad-models.json")
	if err := os.WriteFile(badModels, []byte(`{"anthropic":""}`), 0o644); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "invoked")
	target := filepath.Join(dir, "sentinel")
	if err := os.WriteFile(target, []byte("#!/bin/sh\nprintf invoked >> "+strconv.Quote(marker)+"\nexit 99\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	base := []string{"-ch", "6", "-core", module, "-consumer", module, "-workflow", module, "-stream-exercise", module, "-stream-model", "declared-model"}
	check := func(name string, flags []string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			args := append(append([]string(nil), flags...), target)
			if code := run(args); code != 2 {
				t.Fatalf("run(%q) exit=%d, want 2", args, code)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("invalid input launched target: %v", err)
			}
		})
	}
	for _, flag := range []string{"-core", "-consumer", "-workflow", "-stream-exercise", "-stream-model"} {
		args := make([]string, 0, len(base))
		for i := 0; i < len(base); i += 2 {
			if base[i] != flag {
				args = append(args, base[i], base[i+1])
			}
		}
		check("missing "+flag, args)
		check("empty "+flag, append(append([]string(nil), base...), flag, ""))
	}
	for _, flag := range []string{"-core", "-consumer", "-workflow", "-stream-exercise"} {
		check("non-module "+flag, append(append([]string(nil), base...), flag, dir))
		check("absent "+flag, append(append([]string(nil), base...), flag, filepath.Join(dir, "absent")))
	}
	check("blank model", append(append([]string(nil), base...), "-stream-model", " \t\n"))
	for _, flag := range []string{"-consumer-case", "-media-models"} {
		check("bad JSON "+flag, append(append([]string(nil), base...), flag, badJSON))
		check("missing JSON "+flag, append(append([]string(nil), base...), flag, filepath.Join(dir, "absent.json")))
	}
	check("invalid model schema", append(append([]string(nil), base...), "-media-models", badModels))
	for _, chapter := range []string{"1", "4", "5"} {
		args := []string{"-ch", chapter}
		if chapter != "1" {
			args = append(args, "-core", module, "-consumer", module)
		}
		if chapter == "5" {
			args = append(args, "-workflow", module)
		}
		for _, flag := range []string{"-stream-exercise", "-stream-model"} {
			check("older chapter "+chapter+" empty "+flag, append(append([]string(nil), args...), flag+"="))
			check("older chapter "+chapter+" nonempty "+flag, append(append([]string(nil), args...), flag, "not-valid-here"))
		}
	}
}

func TestChapter7ValidationBeforeLaunch(t *testing.T) {
	testGUIChapterValidationBeforeLaunch(t, "7")
}

func TestChapter8ValidationBeforeLaunch(t *testing.T) {
	testGUIChapterValidationBeforeLaunch(t, "8")
}

func TestChapter9ValidationBeforeLaunch(t *testing.T) {
	testGUIChapterValidationBeforeLaunch(t, "9")
}

func TestChapter10ValidationBeforeLaunch(t *testing.T) {
	testGUIChapterValidationBeforeLaunch(t, "10")
}

func TestChapter11ValidationBeforeLaunch(t *testing.T) {
	testGUIChapterValidationBeforeLaunch(t, "11")
}

func testGUIChapterValidationBeforeLaunch(t *testing.T, chapter string) {
	dir := t.TempDir()
	module := filepath.Join(dir, "module")
	if err := os.Mkdir(module, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte("module example.test/core\n\ngo 1.25\n"), 0600); err != nil {
		t.Fatal(err)
	}
	badModule := filepath.Join(dir, "directory-go-mod")
	if err := os.MkdirAll(filepath.Join(badModule, "go.mod"), 0700); err != nil {
		t.Fatal(err)
	}
	badJSON := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badJSON, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "invoked")
	target := filepath.Join(dir, "sentinel")
	if err := os.WriteFile(target, []byte("#!/bin/sh\nprintf invoked >> "+strconv.Quote(marker)+"\nexit 99\n"), 0700); err != nil {
		t.Fatal(err)
	}
	base := []string{"-ch", chapter, "-core", module, "-consumer", module, "-workflow", module, "-stream-exercise", module, "-stream-model", "declared-stream", "-gui", module, "-gui-model", "declared-gui"}
	check := func(name string, flags []string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			args := append(append([]string(nil), flags...), target)
			if code := run(args); code != 2 {
				t.Fatalf("run(%q) exit=%d, want 2", args, code)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("invalid input launched target: %v", err)
			}
		})
	}
	for _, flag := range []string{"-core", "-consumer", "-workflow", "-stream-exercise", "-stream-model", "-gui", "-gui-model"} {
		args := make([]string, 0, len(base))
		for i := 0; i < len(base); i += 2 {
			if base[i] != flag {
				args = append(args, base[i], base[i+1])
			}
		}
		check("missing "+flag, args)
		check("empty "+flag, append(append([]string(nil), base...), flag, ""))
	}
	for _, flag := range []string{"-core", "-consumer", "-workflow", "-stream-exercise", "-gui"} {
		for label, path := range map[string]string{"non-module": dir, "absent": filepath.Join(dir, "absent"), "nonregular go.mod": badModule} {
			check(label+" "+flag, append(append([]string(nil), base...), flag, path))
		}
	}
	for _, flag := range []string{"-stream-model", "-gui-model"} {
		check("blank "+flag, append(append([]string(nil), base...), flag, " \t\n"))
	}
	for _, flag := range []string{"-consumer-case", "-media-models"} {
		check("bad JSON "+flag, append(append([]string(nil), base...), flag, badJSON))
		check("missing JSON "+flag, append(append([]string(nil), base...), flag, filepath.Join(dir, "absent.json")))
	}
	for _, ch := range []string{"1", "4", "5", "6"} {
		args := []string{"-ch", ch}
		if ch != "1" {
			args = append(args, "-core", module, "-consumer", module)
		}
		if ch == "5" || ch == "6" {
			args = append(args, "-workflow", module)
		}
		if ch == "6" {
			args = append(args, "-stream-exercise", module, "-stream-model", "declared-stream")
		}
		for _, flag := range []string{"-gui", "-gui-model"} {
			check("older "+ch+" empty "+flag, append(append([]string(nil), args...), flag+"="))
			check("older "+ch+" nonempty "+flag, append(append([]string(nil), args...), flag, "not-valid-here"))
		}
	}
}

func TestChapter12ValidationBeforeLaunch(t *testing.T) {
	testGUIChapterValidationBeforeLaunch(t, "12")
}

func TestChapter13ValidationBeforeLaunch(t *testing.T) {
	testGUIChapterValidationBeforeLaunch(t, "13")
}

// Chapter 14 grades the GUI directly; none of the parity artifact flags apply.
func TestChapter14ValidationBeforeLaunch(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "invoked")
	target := filepath.Join(dir, "gui")
	if err := os.WriteFile(target, []byte("#!/bin/sh\nprintf invoked > "+strconv.Quote(marker)+"\nexit 99\n"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"core", "consumer", "workflow", "stream-exercise", "stream-model", "gui", "gui-model", "consumer-case", "media-models"} {
		for _, value := range []string{"", "unused"} {
			if code := run([]string{"-ch", "14", "-" + flag + "=" + value, target}); code != 2 {
				t.Fatalf("%s=%q exit %d", flag, value, code)
			}
		}
	}
	for _, args := range [][]string{{"-ch", "14"}, {"-ch", "14", filepath.Join(dir, "missing")}, {"-ch", "14", target, target}} {
		if code := run(args); code != 2 {
			t.Fatalf("%q exit %d", args, code)
		}
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("invalid input launched target: %v", err)
	}
}

func TestChapter15GuardsAndRoute(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "launched")
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("#!/bin/sh\nprintf launched >> "+strconv.Quote(marker)+"\nexit 71\n"), 0700); err != nil {
		t.Fatal(err)
	}
	cases := [][]string{
		{"-ch", "15", target},
		{"-ch", "15", "-cache-prices", "", target},
		{"-ch", "15", "-cache-prices", "1,2,3", target},
		{"-ch", "15", "-cache-prices", "NaN,2,3,4", target},
		{"-ch", "15", "-cache-prices", "1,2,3,4", "-cache-current", "", target},
		{"-ch", "15", "-cache-prices", "1,2,3,4", "-cache-prior", " ", target},
	}
	for _, flag := range []string{"core", "consumer", "consumer-case", "workflow", "media-models", "stream-exercise", "stream-model", "gui", "gui-model"} {
		cases = append(cases, []string{"-ch", "15", "-cache-prices", "1,2,3,4", "-" + flag, "", target})
	}
	for _, ch := range []string{"1", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14"} {
		for _, flag := range []string{"cache-current", "cache-prior", "cache-prices"} {
			cases = append(cases, []string{"-ch", ch, "-" + flag, "", target})
		}
	}
	for _, args := range cases {
		if got := run(args); got != 2 {
			t.Fatalf("%v status=%d", args, got)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatalf("invalid options launched target: %v", args)
		}
	}
	// Capture the real route's deliberately failing report, not a substitute grader.
	out, err := os.CreateTemp(dir, "report-")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = out
	status := run([]string{"-ch", "15", "-cache-prices", "2,2.5,.2,10", "-json", target})
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
		t.Fatalf("report: %v: %s", err, body)
	}
	if status != 1 || !strings.Contains(report.Title, "Chapter 15") || report.MaxScore != 100 || len(report.Checks) != 7 {
		t.Fatalf("wrong route/status: %d %s", status, body)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("route did not launch explicit target: %v", err)
	}
}

func TestChapter16GuardsAndRoute(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "launched")
	target := filepath.Join(dir, "gui")
	if err := os.WriteFile(target, []byte("#!/bin/sh\nprintf launched >> "+strconv.Quote(marker)+"\nexit 71\n"), 0700); err != nil {
		t.Fatal(err)
	}
	cases := [][]string{{"-ch", "16"}, {"-ch", "16", ""}, {"-ch", "16", filepath.Join(dir, "missing")}, {"-ch", "17", target}, {"-ch", "16", target, target}}
	for _, flag := range []string{"core", "consumer", "consumer-case", "workflow", "media-models", "stream-exercise", "stream-model", "gui", "gui-model", "cache-current", "cache-prior", "cache-prices"} {
		for _, value := range []string{"", "unused"} {
			cases = append(cases, []string{"-ch", "16", "-" + flag + "=" + value, target})
		}
	}
	for _, args := range cases {
		if got := run(args); got != 2 {
			t.Fatalf("%q exit %d", args, got)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatalf("invalid invocation launched target: %q (%v)", args, err)
		}
	}
	out, err := os.CreateTemp(dir, "report-")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = out
	status := run([]string{"-ch", "16", "-json", target})
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
		t.Fatalf("%v: %s", err, body)
	}
	want := []string{"credential-provider", "oauth-flow", "token-refresh", "responses-format", "cache-breakpoints", "reasoning-summaries", "billing-mode"}
	if status != 1 || !strings.Contains(report.Title, "Chapter 16") || report.MaxScore != 100 || len(report.Checks) != len(want) {
		t.Fatalf("wrong route: %d %s", status, body)
	}
	for i, id := range want {
		if report.Checks[i].ID != id {
			t.Fatalf("check %d: %s != %s", i, report.Checks[i].ID, id)
		}
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("explicit GUI never launched: %v", err)
	}
}
