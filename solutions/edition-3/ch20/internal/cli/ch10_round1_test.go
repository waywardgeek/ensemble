package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCh10R1CLIOverrideAndRetention(t *testing.T) {
	for _, mode := range []string{"chat", "protocol"} {
		t.Run(mode, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("LLM_MODEL", "fixture")
			t.Setenv("LLM_TOOLS", "none")
			path := filepath.Join(t.TempDir(), "chosen.json")
			if err := os.WriteFile("api.log", []byte(strings.Repeat("old\n", 205)), 0600); err != nil {
				t.Fatal(err)
			}
			var out, errout bytes.Buffer
			if code := Run(context.Background(), append([]string{"--save", path}, func() []string {
				if mode == "chat" {
					return []string{"chat"}
				}
				return nil
			}()...), strings.NewReader(""), &out, &errout); code != nil {
				t.Fatalf("CLI error %v: %s", code, errout.String())
			}
			if _, err := os.Stat(path); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat("save.json"); !os.IsNotExist(err) {
				t.Fatalf("default save touched: %v", err)
			}
			b, err := os.ReadFile("api.log")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(b), "\n") != 200 {
				t.Fatalf("quiet startup/exit retention records=%d", strings.Count(string(b), "\n"))
			}
		})
	}
}
func TestCh10R1FailedLoadKeepsDiagnostics(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("LLM_MODEL", "fixture")
	t.Setenv("LLM_TOOLS", "none")
	before := []byte(strings.Repeat("old\n", 205))
	_ = os.WriteFile("api.log", before, 0600)
	_ = os.WriteFile("save.json", []byte("corrupt"), 0600)
	var out, errout bytes.Buffer
	if code := Run(context.Background(), []string{"chat"}, strings.NewReader(""), &out, &errout); code == nil {
		t.Fatal("accepted bad save")
	}
	after, err := os.ReadFile("api.log")
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("failed load changed diagnostic: %v", err)
	}
}

// The line reader is reached only after startup recovery/save and retention.
// Inject a normal-exit save failure there, after adding fresh diagnostics.
type ch10FailSaveInput struct {
	t    *testing.T
	once bool
	path string
}

func (r *ch10FailSaveInput) Read(p []byte) (int, error) {
	if !r.once {
		r.once = true
		for _, path := range []string{r.path, r.path + ".bak"} {
			if err := os.WriteFile(path, []byte("corrupt"), 0600); err != nil {
				r.t.Fatal(err)
			}
		}
		f, err := os.OpenFile("api.log", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			r.t.Fatal(err)
		}
		_, err = f.WriteString(strings.Repeat("after-startup\n", 10))
		if err != nil {
			r.t.Fatal(err)
		}
		if err = f.Close(); err != nil {
			r.t.Fatal(err)
		}
	}
	return 0, io.EOF
}
func TestCh10R1FailedExitSaveKeepsDiagnostics(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("LLM_MODEL", "fixture")
	t.Setenv("LLM_TOOLS", "none")
	if err := os.WriteFile("api.log", []byte(strings.Repeat("old\n", 205)), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errout bytes.Buffer
	err := Run(context.Background(), nil, &ch10FailSaveInput{t: t, path: "save.json"}, &out, &errout)
	if err == nil || (!strings.Contains(err.Error(), "neither recovery candidate") && !strings.Contains(err.Error(), "reading archived prefix")) {
		t.Fatalf("expected save refusal: %v", err)
	}
	b, err := os.ReadFile("api.log")
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != strings.Repeat("old\n", 200)+strings.Repeat("after-startup\n", 10) {
		t.Fatalf("failed save trimmed diagnostics: records=%d", strings.Count(string(b), "\n"))
	}
	for _, path := range []string{"save.json", "save.json.bak"} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != "corrupt" {
			t.Fatalf("refused save changed %s: %q, %v", path, data, err)
		}
	}
}

func TestCh10R1CLIReadOnlyModes(t *testing.T) {
	for _, mode := range []string{"render", "dump"} {
		t.Run(mode, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("LLM_MODEL", "fixture")
			t.Setenv("LLM_TOOLS", "none")
			const untouched = "not a conversation archive: leave intact"
			if err := os.WriteFile("chosen.json", []byte(untouched), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile("ch02.jsonl", nil, 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"--save", "chosen.json", mode}
			if mode == "render" {
				if err := os.WriteFile("events.jsonl", nil, 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "events.jsonl")
			}
			var out, errout bytes.Buffer
			if err := Run(context.Background(), args, strings.NewReader(""), &out, &errout); err != nil {
				t.Fatalf("%s: %v %s", mode, err, errout.String())
			}
			raw, err := os.ReadFile("chosen.json")
			if err != nil || string(raw) != untouched {
				t.Fatalf("offline mode changed archive: %q %v", raw, err)
			}
		})
	}
}
