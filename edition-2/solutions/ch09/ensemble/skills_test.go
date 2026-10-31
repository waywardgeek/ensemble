package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"ensemble/ensemble"
)

func skillFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, text := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func skillAgent(t *testing.T, root *ensemble.Ensemble, dir, primary, url string) ensemble.Agent {
	t.Helper()
	a, err := root.NewAgent(ensemble.Config{Model: "fake", BaseURL: url, BuiltinTools: true, DataDir: t.TempDir()}, ensemble.AgentSpec{SkillsDir: dir, PrimarySkill: primary, Variables: map[string]func() string{"CUSTOM": func() string { return "creation" }}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := a.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	return a
}

func toolNames(a ensemble.Agent) []string {
	var names []string
	for _, declaration := range a.Declarations() {
		names = append(names, declaration.Name)
	}
	return names
}

// A real HTTP peer drives hidden dispatch, then loads instructions and observes
// the next request. Tool handlers, jobs, registry, reducer and vendor code all run.
func TestSkillsThroughLoop(t *testing.T) {
	dir := skillFiles(t, map[string]string{
		"base":   "---\nname: base\ntype: primary\ntools: read_file\nloadable-skills: code hidden\n---\nConstitution $CUSTOM. $TOOLS. Next: $SKILLS. $UNKNOWN\n---\nBody after rule.",
		"code":   "---\nname: code\ntype: loadable\ntools:\n  - write_file\ndepends:\n  - helper\nloadable-skills: next\n---\nCode $CUSTOM. $TOOLS. Next: $SKILLS.",
		"helper": "---\nname: helper\ntype: dependency\ntools: list_directory read_file\n---\nHelper instructions.",
		"next":   "---\nname: next\ntype: loadable\ntools: search_files\n---\nNext instructions.",
		"hidden": "---\nname: hidden\ntype: dependency\ntools: edit_file\n---\nHidden instructions.",
	})
	path := filepath.Join(t.TempDir(), "output.txt")
	var requests []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		requests = append(requests, body)
		var calls []map[string]any
		call := func(id, name string, input map[string]any) {
			calls = append(calls, map[string]any{"type": "tool_use", "id": id, "name": name, "input": input})
		}
		switch len(requests) {
		case 1:
			call("denied", "write_file", map[string]any{"path": path, "content": "forbidden"})
			call("load", "load_skill", map[string]any{"name": "code"})
		case 2:
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Error("hidden tool executed")
			}
			call("write", "write_file", map[string]any{"path": path, "content": "enabled"})
			call("next", "load_skill", map[string]any{"name": "next"})
			call("unload", "unload_skill", map[string]any{"name": "code"})
		case 3:
			_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
			return
		default:
			t.Error("unexpected extra request")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"content": calls, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer server.Close()
	root := ensemble.New(io.Discard)
	a := skillAgent(t, root, dir, "base", server.URL)
	other := skillAgent(t, root, dir, "base", server.URL)
	if got := toolNames(a); !reflect.DeepEqual(got, []string{"read_file", "load_skill", "unload_skill"}) {
		t.Fatalf("initial declarations: %v", got)
	}
	if got := a.Skills().LoadableSkills(); !reflect.DeepEqual(got, []string{"code"}) {
		t.Fatalf("initial disclosure: %v", got)
	}
	a.RegisterVar("CUSTOM", func() string { return "dynamic" })
	if answer, err := a.Ask(context.Background(), "exercise skills"); err != nil || answer != "done" {
		t.Fatalf("turn: %q %v", answer, err)
	}
	if len(requests) != 3 {
		t.Fatalf("requests: %d", len(requests))
	}
	initial := requests[0]["system"].(string)
	for _, fragment := range []string{"Constitution creation", "read_file", "Next: code", "$UNKNOWN", "---\nBody after rule."} {
		if !strings.Contains(initial, fragment) {
			t.Errorf("missing initial %q: %s", fragment, initial)
		}
	}
	for _, body := range requests[1:] {
		if body["system"] != initial {
			t.Error("system prompt changed after load")
		}
	}
	wire, _ := json.Marshal(requests[1])
	for _, fragment := range []string{"Code dynamic", "Helper instructions.", "Next: next", "unknown tool", "write_file", "list_directory"} {
		if !strings.Contains(string(wire), fragment) {
			t.Errorf("missing dynamic %q: %s", fragment, wire)
		}
	}
	for _, name := range []string{"write_file", "list_directory", "search_files"} {
		if !slices.Contains(toolNames(a), name) {
			t.Errorf("tool not enabled: %s", name)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "enabled" {
		t.Fatalf("enabled dispatch: %q %v", data, err)
	}
	if got := a.Skills().PendingUnload(); !reflect.DeepEqual(got, []string{"code"}) {
		t.Fatalf("pending: %v", got)
	}
	if len(a.Skills().LoadableSkills()) != 0 {
		t.Fatal("loaded or hidden skills still offered")
	}
	if !reflect.DeepEqual(toolNames(other), []string{"read_file", "load_skill", "unload_skill"}) {
		t.Fatal("skills leaked between Agents")
	}
	if meta, ok := a.Skills().ToolMeta("read_file"); !ok || meta.Source != ensemble.SourceInitial || meta.EventSeq != 0 {
		t.Fatalf("shared initial provenance changed: %+v", meta)
	}
	var seq uint64
	for _, event := range a.History().EventsAfter(0) {
		if event.Type == "tool_called" && event.Tool.CallID == "load" {
			seq = event.Seq
		}
	}
	if meta, ok := a.Skills().ToolMeta("write_file"); !ok || meta.Source != ensemble.SourceDynamic || meta.EventSeq != seq || seq == 0 {
		t.Fatalf("dynamic provenance: %+v, expected %d", meta, seq)
	}
}

// Failed recursive loads grant nothing; graph traversal catches cycles before
// revisiting a node and leaves the earlier accepted primary available.
func TestSkillLoadFailures(t *testing.T) {
	dir := skillFiles(t, map[string]string{
		"base":        "---\nname: base\ntype: primary\nloadable-skills: cycle missing\n---\nBase",
		"cycle":       "---\nname: cycle\ntype: loadable\ndepends: good loop\n---\nCycle",
		"good":        "---\nname: good\ntype: dependency\ntools: write_file\n---\nGood",
		"loop":        "---\nname: loop\ntype: dependency\ndepends: cycle\n---\nLoop",
		"missing":     "---\nname: missing\ntype: loadable\ndepends: good absent\n---\nMissing",
		"unreachable": "---\nname: unreachable\ntype: loadable\n---\nUnreachable",
	})
	a := skillAgent(t, ensemble.New(io.Discard), dir, "base", "")
	before := toolNames(a)
	for _, test := range []struct{ name, want string }{{"base", "not loadable"}, {"good", "not loadable"}, {"unreachable", "not available"}, {"absent", "unknown skill"}, {"cycle", "cycle"}, {"missing", "missing skill dependency"}} {
		if _, err := a.Skills().LoadDynamic(test.name, 10); err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("load %s: %v", test.name, err)
		}
		if !reflect.DeepEqual(toolNames(a), before) {
			t.Errorf("failed %s changed declarations", test.name)
		}
	}
	if err := a.Skills().MarkUnload("unreachable"); err == nil {
		t.Error("unloaded skill marked pending")
	}
}
