package agent

import (
	"testing"
)

// Safe mode and the web-search gate are both withholding, and withholding is
// the kind of behavior that rots quietly. Nothing fails when a withheld tool
// comes back: the agent works better. So these tests assert absence, and
// each one is paired with a positive control asserting the same thing is
// present when it should be, because a constructor that produced an agent
// with no tools at all would satisfy every absence test here.

func TestSafeModeWithholdsExecTools(t *testing.T) {
	a, err := NewAgent(testSpecConfig(), AgentSpec{
		DataDir:  t.TempDir(),
		SafeMode: true,
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}
	defer a.Shutdown()

	for _, name := range SafeModeWithheldTools() {
		if _, err := a.Registry().Lookup(name); err == nil {
			t.Errorf("safe mode left %q in the registry", name)
		}
	}
}

// POSITIVE CONTROL for the test above.
func TestWithoutSafeModeExecToolsArePresent(t *testing.T) {
	a, err := NewAgent(testSpecConfig(), AgentSpec{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}
	defer a.Shutdown()

	for _, name := range SafeModeWithheldTools() {
		if _, err := a.Registry().Lookup(name); err != nil {
			t.Errorf("%q is missing outside safe mode: %v", name, err)
		}
	}
}

// Safe mode withholds the exec tools and nothing else. An implementation
// that emptied the registry would pass the absence test above; this one
// says the agent can still read and write files, which is the whole point
// of a mode you would willingly run in.
func TestSafeModeKeepsOrdinaryTools(t *testing.T) {
	a, err := NewAgent(testSpecConfig(), AgentSpec{
		DataDir:  t.TempDir(),
		SafeMode: true,
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}
	defer a.Shutdown()

	for _, name := range []string{"read_file", "write_file", "edit_file", "list_directory"} {
		if _, err := a.Registry().Lookup(name); err != nil {
			t.Errorf("safe mode removed %q, which it should keep: %v", name, err)
		}
	}
}

func TestWebSearchDisabledForgetsTheSkill(t *testing.T) {
	a, err := NewAgent(testSpecConfig(), AgentSpec{
		DataDir:         t.TempDir(),
		SkillDir:        "skills",
		EnableWebSearch: false,
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}
	defer a.Shutdown()

	if a.Skills().Get(WebSearchSkill) != nil {
		t.Error("web-search skill is still registered with web search disabled")
	}
	for _, s := range a.Skills().LoadableSkills() {
		if s.Name == WebSearchSkill {
			t.Error("web-search is still offered as loadable")
		}
	}
}

// POSITIVE CONTROL. Without this, a typo in the skill directory would make
// the test above pass for the wrong reason: no skills discovered at all.
func TestWebSearchEnabledKeepsTheSkill(t *testing.T) {
	a, err := NewAgent(testSpecConfig(), AgentSpec{
		DataDir:         t.TempDir(),
		SkillDir:        "skills",
		EnableWebSearch: true,
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}
	defer a.Shutdown()

	if a.Skills().Get(WebSearchSkill) == nil {
		t.Fatal("web-search skill was not discovered even with web search enabled; " +
			"the disabled test above may be passing for the wrong reason")
	}
}
