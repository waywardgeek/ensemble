package agent

import (
	"os"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

func testSpecConfig() Config {
	return Config{Model: "claude-sonnet-5", Vendor: common.VendorAnthropic, APIKey: "test-key"}
}

// TestNewAgentBuildsCompleteAgent asserts that the library constructor
// produces an agent that could actually run.
//
// This test exists because it did not. NewAgent went sixteen chapters without
// a single caller, and in that time the command-line root taught itself nine
// capabilities the constructor never learned: restored context, the journal,
// settings, the context target, the tool-round limit, memory, the memory
// bands, recall, and the cache lens. Nothing failed, because nothing asked.
// A constructor with no caller has no way to be wrong.
func TestNewAgentBuildsCompleteAgent(t *testing.T) {
	a, err := NewAgent(testSpecConfig(), AgentSpec{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}
	defer a.Shutdown()

	parts := []struct {
		name string
		ok   bool
	}{
		{"Engine", a.Engine() != nil},
		{"Actor", a.Actor() != nil},
		{"Jobs", a.Jobs() != nil},
		{"Registry", a.Registry() != nil},
		{"Skills", a.Skills() != nil},
		{"Vars", a.Vars() != nil},
		{"Settings", a.Settings() != nil},
		{"SaveFile", a.SaveFile() != nil},
	}
	for _, p := range parts {
		if !p.ok {
			t.Errorf("agent has no %s", p.name)
		}
	}

	if a.Engine() == nil {
		t.Fatal("cannot check engine capabilities without an engine")
	}
	eng := a.Engine()
	caps := []struct {
		name string
		ok   bool
	}{
		{"Ctx", eng.Ctx != nil},
		{"Journal", eng.Journal != nil},
		{"Cache", eng.Cache != nil},
		{"Memory", eng.Memory != nil},
		{"Target", eng.Target != nil},
		{"ToolRoundLimit", eng.ToolRoundLimit != nil},
		{"Bands", eng.Bands != nil},
		// Recall is deliberately absent from this list. It is left nil when
		// nothing is archived, which is the state of every agent on its first
		// run, so a fresh data directory must NOT have one. Asserting it
		// non-nil here would be asserting a bug.
	}
	for _, c := range caps {
		if !c.ok {
			t.Errorf("engine capability %s is nil: NewAgent does not build a complete agent", c.name)
		}
	}
}

// TestAgentFilesLiveInItsDataDirectory is the property that makes more than
// one agent possible. Before the data directory existed, every one of these
// files was created in the process working directory, so a second agent in
// the same process would have written over the first one's save file.
func TestAgentFilesLiveInItsDataDirectory(t *testing.T) {
	dirA, dirB := t.TempDir(), t.TempDir()

	a, err := NewAgent(testSpecConfig(), AgentSpec{DataDir: dirA})
	if err != nil {
		t.Fatalf("NewAgent A: %v", err)
	}
	defer a.Shutdown()

	b, err := NewAgent(testSpecConfig(), AgentSpec{DataDir: dirB})
	if err != nil {
		t.Fatalf("NewAgent B: %v", err)
	}
	defer b.Shutdown()

	// Each agent put something in its own directory.
	for _, d := range []struct {
		label string
		path  string
	}{{"A", dirA}, {"B", dirB}} {
		entries, err := os.ReadDir(d.path)
		if err != nil {
			t.Fatalf("read %s: %v", d.path, err)
		}
		if len(entries) == 0 {
			t.Errorf("agent %s created no files in its data directory; "+
				"they are going somewhere else, most likely the working directory", d.label)
		}
	}

	// The two agents agree on nothing but the file names.
	if a.Spec().Path("save.json") == b.Spec().Path("save.json") {
		t.Error("two agents resolved the same save file path")
	}
}
