package gradeedition2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

func TestChapter14InstructionPlacement(t *testing.T) {
	for _, tc := range []struct {
		name, system, content string
		loaded, good          bool
	}{
		{"startup", contextPrimaryMark, "hello", false, true},
		{"dynamic", contextPrimaryMark + "\n" + contextManualMark, "loaded", true, true},
		{"old frozen policy", contextPrimaryMark, contextManualMark, true, false},
		{"missing primary", contextManualMark, "loaded", true, false},
		{"duplicate manual", contextPrimaryMark + contextManualMark + contextManualMark, "loaded", true, false},
		{"duplicate dialogue", contextPrimaryMark + contextManualMark, contextManualMark, true, false},
		{"premature manual", contextPrimaryMark + contextManualMark, "hello", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, system := range []any{tc.system, []any{map[string]string{"type": "text", "text": tc.system}}} {
				body, err := json.Marshal(map[string]any{"system": system, "messages": []any{map[string]any{"role": "user", "content": tc.content}}})
				if err != nil {
					t.Fatal(err)
				}
				req, err := contextParse(body)
				if err != nil {
					t.Fatal(err)
				}
				got := contextInstructions(req, tc.loaded)
				if (got == "") != tc.good {
					t.Fatalf("good=%v: %s", tc.good, got)
				}
			}
		})
	}
}

func TestChapter14PrimaryGrants(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "skills")
	if err := contextCreateSkills(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "base", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "tools: read_file load_skill unload_skill keep_tool_results micro_handoff\n") || !strings.Contains(string(data), contextPrimaryMark) {
		t.Fatalf("fixture omits declared lifecycle grants or primary marker")
	}
}

// This tests the grader's orchestration using supplied request records. It does
// not establish that any student can recover a backup or append a journal.
func TestChapter14CrashVariants(t *testing.T) {
	body := strings.Join(append(append([]string{}, contextCrashPrompts...), "REPLY-ONE", "REPLY-TWO", "REPLY-THREE"), "\n")
	for _, bad := range []string{"", "lost", "duplicate", "different", "fatal"} {
		t.Run("fault-"+bad, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "C")
			if err := os.Mkdir(dir, 0700); err != nil {
				t.Fatal(err)
			}
			starts := 0
			recoveries := 0
			run := func(o contextOpts) contextOut {
				starts++
				if starts == 1 {
					for name, text := range map[string]string{"save.json": "primary", "save.json.bak": "backup", "tail": "retained tail"} {
						if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
							t.Fatal(err)
						}
					}
					return contextOut{}
				}
				if starts == 2 {
					if !o.kill {
						t.Fatal("missing killed-session step")
					}
					return contextOut{}
				}
				recoveries++
				tail, err := os.ReadFile(filepath.Join(o.dir, "tail"))
				if err != nil || string(tail) != "retained tail" {
					t.Fatalf("tail not cloned: %q %v", tail, err)
				}
				b := body
				if strings.HasSuffix(o.dir, "-missing") {
					if _, err := os.Stat(filepath.Join(o.dir, "save.json")); !os.IsNotExist(err) {
						t.Fatalf("primary not missing: %v", err)
					}
					switch bad {
					case "lost":
						b = strings.ReplaceAll(b, "REPLY-TWO", "")
					case "duplicate":
						b += "\nREPLY-TWO"
					case "different":
						b += "\nEXTRA"
					case "fatal":
						return contextOut{fatal: "refused fallback"}
					}
				}
				if strings.HasSuffix(o.dir, "-corrupt") {
					data, err := os.ReadFile(filepath.Join(o.dir, "save.json"))
					if err != nil || string(data) != "{broken primary" {
						t.Fatalf("primary not corrupt: %q %v", data, err)
					}
				}
				return contextOut{reqs: []fakevendor.Recorded{{Body: []byte(b)}}}
			}
			r := contextResult{Errs: map[string]string{}}
			contextCrash(run, dir, &r)
			errText, ran := r.Errs["crash-recovery"]
			if !ran || (errText == "") != (bad == "") {
				t.Fatalf("bad=%q result=%q ran=%v", bad, errText, ran)
			}
			if bad == "" && recoveries != 3 {
				t.Fatalf("only %d recovery variants", recoveries)
			}
		})
	}
}

func TestChapter14TotalCannotFallback(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "source")
	if err := os.MkdirAll(src, 0700); err != nil {
		t.Fatal(err)
	}
	if err := contextCreateSkills(filepath.Join(root, "skills")); err != nil {
		t.Fatal(err)
	}
	archive := `{"context":{},"log":[{"seq":1,"type":"user_message","user":{"text":"` + contextCrashLastPrompt + `"}}]}`
	for name, data := range map[string]string{"save.json": archive, "save.json.bak": "valid backup stand-in", "journal.jsonl": "complete tail stand-in"} {
		if err := os.WriteFile(filepath.Join(src, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	dst := filepath.Join(t.TempDir(), "probe")
	if err := contextCopyReducerInput(src, dst); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "save.json" {
		t.Fatalf("alternative recovery inputs copied: %v", entries)
	}
	r := contextResult{Errs: map[string]string{}}
	called := false
	contextTotal(func(o contextOpts) contextOut {
		called = true
		data, err := os.ReadFile(filepath.Join(o.dir, "save.json"))
		if err != nil {
			t.Fatal(err)
		}
		var top map[string]json.RawMessage
		if err := json.Unmarshal(data, &top); err != nil {
			t.Fatal(err)
		}
		if string(top["context"]) != "null" || !strings.Contains(string(top["log"]), `"type":"skill_loaded"`) {
			t.Fatal("planted replay input missing")
		}
		return contextOut{fatal: "strict known-payload rejection"}
	}, dst, &r)
	if !called || r.Errs["total-reducer"] == "" {
		t.Fatal("strict refusal passed total reducer")
	}
}

func TestChapter14StaticFixture(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "gui")
	if err := contextStaticFixture(dir); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 3 {
		t.Fatalf("fixture entries: %v %v", entries, err)
	}
	for _, name := range []string{"index.html", "app.js", "style.css"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || len(b) == 0 {
			t.Fatalf("fixture %s: %v", name, err)
		}
	}
	if err := contextStaticFixture(filepath.Join(dir, "index.html", "child")); err == nil {
		t.Fatal("accepted a non-directory parent")
	}
	blocked := t.TempDir()
	if err := os.Mkdir(filepath.Join(blocked, "app.js"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := contextStaticFixture(blocked); err == nil {
		t.Fatal("ignored an asset write failure")
	}
}
