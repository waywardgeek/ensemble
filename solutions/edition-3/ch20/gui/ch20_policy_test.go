package gui

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"reflect"
	"testing"

	"ensemble"
)

func TestCh20SelectedViewPolicyIsolation(t *testing.T) {
	root := ensemble.New(nil)
	defer root.Shutdown()
	base := t.TempDir()
	makeAgent := func(id string, safe, web bool) ensemble.Agent {
		t.Helper()
		cfg := ensemble.Config{ID: id, Model: "offline", BaseDir: base, DataDir: filepath.Join(base, id), SavePath: filepath.Join(base, id, "save.json"), SafeMode: safe, EnableWebSearch: web, ContextTarget: 4096, MaxToolRounds: 7}
		a, err := root.NewAgent(cfg)
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	first := makeAgent("first", true, false)
	selected := makeAgent("selected", false, true)
	before := first.Config()
	view, err := New(root, selected, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer view.Close()
	if got := view.Hub.currentSettings(); got.SafeMode || !got.EnableWebSearch {
		t.Fatalf("settings came from first Agent: %+v", got)
	}
	if !reflect.DeepEqual(before, first.Config()) {
		t.Fatal("constructor mutated unrelated Agent")
	}
	cfg := first.Config().Child(filepath.Join(base, "late"))
	cfg.ID, cfg.ParentID = "late", first.ID()
	late, err := root.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	lateBefore := late.Config()
	view.Hub.mu.Lock()
	view.Hub.view(late)
	view.Hub.mu.Unlock()
	client := view.Hub.connect()
	defer client.stop()
	if !reflect.DeepEqual(lateBefore, late.Config()) {
		t.Fatal("late view mutated delegated policy")
	}
	if err := view.Hub.updateSettings(json.RawMessage(`{"safe_mode":true,"enable_web_search":false,"max_tool_rounds":9}`)); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, first.Config()) || !reflect.DeepEqual(lateBefore, late.Config()) {
		t.Fatal("selected settings changed unrelated Agent")
	}
	if !selected.Config().SafeMode || selected.Config().EnableWebSearch || selected.Config().MaxToolRounds != 9 {
		t.Fatal("selected settings were not applied")
	}
	// The embedded application must consume the same public modules embedders use.
	raw, err := Assets().Open("app.js")
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	content, _ := io.ReadAll(raw)
	for _, module := range []string{"./artifact-scroll.js", "new ArtifactScroll(", "artifactScroll.receive(frame)"} {
		if !bytes.Contains(content, []byte(module)) {
			t.Fatal("default application does not consume", module)
		}
	}
}
