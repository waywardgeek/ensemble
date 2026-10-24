package ensemble

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCh10R2DirectDispatchArchive(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{Model: "fixture", BaseDir: dir, DataDir: "cr", SavePath: filepath.Join(dir, "archive.json")}
	root := New(nil)
	defer root.Shutdown()
	a, err := root.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, call := range []struct {
		name, args string
		failure    bool
	}{
		{"write_file", `{"path":"effect.txt","content":"saved effect"}`, false},
		{"tool_limits", `{"max_output_bytes":4096}`, false},
		{"read_file", `{"path":"missing.txt"}`, true},
	} {
		text, failed := a.Tools().Dispatch(context.Background(), call.name, json.RawMessage(call.args))
		if failed != call.failure {
			t.Fatalf("Dispatch %s: error=%v text=%v", call.name, failed, text)
		}
	}
	// Supplement direct Dispatch with a genuinely supplied correlation ID. Do not
	// replace the ID-less public calls above with Invoke or synthesize their IDs.
	args := json.RawMessage(`{"path":"effect.txt"}`)
	if err = a.Tools().Invoke(context.Background(), Part{Type: "tool_call", Name: "read_file", CallID: "native-unchanged", Args: args}, ""); err != nil {
		t.Fatal(err)
	}
	if err = root.Shutdown(); err != nil {
		t.Fatalf("dispatch history cannot be saved: %v", err)
	}
	original := a.Log().Events()
	ordinary, control, failed, supplied := false, false, false, 0
	var suppliedPart uint64
	for _, e := range original {
		if e.Tool == nil {
			continue
		}
		if e.Tool.CallID == "native-unchanged" {
			supplied++
			if suppliedPart == 0 {
				suppliedPart = e.Tool.CallPartID
			}
			if string(e.Tool.Args) != string(args) || suppliedPart == 0 || e.Tool.CallPartID != suppliedPart {
				t.Fatalf("supplied correlation changed: %+v", e.Tool)
			}
		} else if e.Tool.CallID != "" || e.TurnID != 0 {
			t.Fatalf("fabricated direct identity: %+v", e)
		}
		if e.Type == "tool_returned" {
			switch e.Tool.Name {
			case "write_file":
				ordinary = !e.Tool.IsError
			case "tool_limits":
				control = !e.Tool.IsError
			case "read_file":
				failed = failed || e.Tool.IsError
			}
		}
	}
	if !ordinary || !control || !failed || supplied != 3 {
		t.Fatalf("missing real operation history: ordinary=%v control=%v failed=%v supplied=%d", ordinary, control, failed, supplied)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "effect.txt"))
	if err != nil || string(raw) != "saved effect" {
		t.Fatalf("effect: %q %v", raw, err)
	}
	reopened := New(nil)
	defer reopened.Shutdown()
	b, err := reopened.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(original)
	after, _ := json.Marshal(b.Log().Events())
	if string(before) != string(after) {
		t.Fatalf("canonical tags/IDs/args/anchors changed on reopen\nbefore=%s\nafter=%s", before, after)
	}
	if b.Recovery().AsOf != a.Log().LastSeq() || b.Log().LastSeq() != a.Log().LastSeq() {
		t.Fatal("recovery boundary changed")
	}
	if err = reopened.Shutdown(); err != nil {
		t.Fatal(err)
	}
}

func TestCh10R2BaseRelativeRecoveryIsolation(t *testing.T) {
	launch := t.TempDir()
	t.Chdir(launch) // A fresh launch directory, not the stopped run's cr/.
	relativeBase, err := filepath.Rel(launch, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	initialEmptyTools := func(events []Event) bool {
		return len(events) == 1 && events[0].Seq == 1 && events[0].Type == ToolsChanged && events[0].Tools != nil && len(events[0].Tools.Added) == 0 && len(events[0].Tools.Removed) == 0
	}
	configs := []Config{{Model: "fixture", BaseDir: t.TempDir(), DataDir: "cr", Tools: []string{}}, {Model: "fixture", BaseDir: relativeBase, DataDir: "cr", Tools: []string{}}}
	for i, cfg := range configs {
		root := New(nil)
		defer root.Shutdown()
		a, err := root.NewAgent(cfg)
		if err != nil {
			t.Fatal(err)
		}
		want, err := filepath.Abs(filepath.Join(cfg.BaseDir, "cr", "agent-1", "save.json"))
		if err != nil {
			t.Fatal(err)
		}
		if a.Config().SavePath != want {
			t.Fatalf("default archive escaped Agent data ownership: got %q want %q", a.Config().SavePath, want)
		}
		wantDataDir, err := filepath.Abs(filepath.Join(cfg.BaseDir, "cr"))
		if err != nil {
			t.Fatal(err)
		}
		if a.Config().DataDir != wantDataDir || !initialEmptyTools(a.Log().Events()) || a.Recovery().Path != "" || len(a.Log().Context().ActiveSkills) != 0 || a.Log().Context().Handoff != "" {
			t.Fatal("fresh root inherited another root's state")
		}
		text := []string{"first owner", "second owner"}[i]
		if err = a.Log().Append(Event{Type: "message_received", Message: &MessageData{Actor: "human", Parts: PartList{{Type: "text", Text: text}}}}); err != nil {
			t.Fatal(err)
		}
		if err = root.Shutdown(); err != nil {
			t.Fatal(err)
		}
	}
	for i, cfg := range configs {
		root := New(nil)
		defer root.Shutdown()
		a, err := root.NewAgent(cfg)
		if err != nil {
			t.Fatal(err)
		}
		es := a.Log().Events()
		want := []string{"first owner", "second owner"}[i]
		if len(es) != 2 || !initialEmptyTools(es[:1]) || es[1].Message == nil || es[1].Message.Actor != "human" || len(es[1].Message.Parts) != 1 || es[1].Message.Parts[0].Text != want || a.Recovery().Path != a.Config().SavePath {
			t.Fatalf("cross-root recovery: %+v", es)
		}
		if err = root.Shutdown(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat("cr"); !os.IsNotExist(err) {
		t.Fatalf("default save touched process cwd: %v", err)
	}
}

func TestCh10R2ExplicitRecoveryPathAndOverlap(t *testing.T) {
	launch := t.TempDir()
	t.Chdir(launch)
	root := New(nil)
	defer root.Shutdown()
	cfg := Config{Model: "fixture", BaseDir: t.TempDir(), DataDir: "cr", SavePath: "explicit.json", Tools: []string{}}
	a, err := root.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if a.Config().SavePath != filepath.Join(launch, "explicit.json") {
		t.Fatal("relocated caller-selected explicit path")
	}
	if err = a.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(launch, "explicit.json")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"explicit.json", "explicit.json.bak"} {
		cfg.SavePath = path
		if _, err = root.NewAgent(cfg); err == nil {
			t.Fatalf("allowed overlapping recovery path %s", path)
		}
	}
}
