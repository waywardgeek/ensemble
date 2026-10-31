package ensemble_test

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"ensemble/ensemble"
)

// Exercise the exported capability table and real Agent-owned engine; unknown
// aliases deliberately do not acquire policy from a provenance-only identity.
func TestVisibleReasoningCapabilities(t *testing.T) {
	if (ensemble.ModelFeatures{}).RequireVisibleReasoning {
		t.Fatal("zero-value capability must remain opt-in")
	}
	for model, want := range map[string]bool{
		"gpt-6-astra": true, "gpt-6.1-sol": true, "gpt-4o-mini": true,
		"gpt-5.6-sol": true, "gpt-5.4-mini": true, "gpt-4o": true,
		"gpt-ch19-course":      false,
		"claude-opus-5-course": false, "claude-sonnet-5-course": false,
		"claude-sonnet-4-6": false, "claude-opus-4-6": false, "claude-opus-5": false,
		"claude-sonnet-5": false, "claude-3-haiku-20240307": false,
		"gemini-3.8-flash": false, "gemini-2.5-flash": false, "gemini-3.1-pro-preview": false,
	} {
		t.Run(model, func(t *testing.T) {
			features, known := ensemble.LookupModel(model)
			if !known || features.RequireVisibleReasoning != want {
				t.Fatalf("%s: known=%t require=%t, want %t", model, known, features.RequireVisibleReasoning, want)
			}
		})
	}
	for _, model := range []string{"", "unknown", "gpt-unknown", "gpt-4o-alias"} {
		features, known := ensemble.LookupModel(model)
		if known || features.RequireVisibleReasoning {
			t.Fatalf("unknown %q acquired capability: %+v", model, features)
		}
	}
	a := visibleReasoningPromptAgent(t, ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o-alias", ResolvedModel: "gpt-4o", SystemPrompt: "Alias role"})
	if a.Engine().Features().RequireVisibleReasoning || a.Engine().SystemPrompt() != "Alias role" || a.Engine().Target().Model != "gpt-4o" {
		t.Fatal("resolved provenance must not replace selected-model capability policy")
	}
	b := visibleReasoningPromptAgent(t, ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o", ResolvedModel: "unknown"})
	if !b.Engine().Features().RequireVisibleReasoning {
		t.Fatal("resolved provenance disabled selected-model requirement")
	}
	fixture, _ := ensemble.LookupModel("gpt-ch19-course")
	if !fixture.Responses || !fixture.ExplicitCache {
		t.Fatal("synthetic legacy transport/cache capabilities changed")
	}
}

func visibleReasoningPromptAgent(t *testing.T, cfg ensemble.Config, specs ...ensemble.AgentSpec) ensemble.Agent {
	t.Helper()
	cfg.DataDir = t.TempDir()
	a, err := ensemble.New(io.Discard).NewAgent(cfg, specs...)
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

// Render through both OpenAI API surfaces and retain the role's exact prefix.
// Restoring an effective prompt must not stack another copy of the instruction.
func TestVisibleReasoningPrompt(t *testing.T) {
	for _, model := range []string{"gpt-4o", "gpt-6-astra"} {
		t.Run(model, func(t *testing.T) {
			cfg := ensemble.Config{Vendor: ensemble.OpenAI, Model: model, SystemPrompt: "Exact role instruction."}
			a := visibleReasoningPromptAgent(t, cfg)
			prompt := a.Engine().SystemPrompt()
			if !strings.HasPrefix(prompt, cfg.SystemPrompt+"\n\n") {
				t.Fatalf("role instruction lost: %q", prompt)
			}
			for _, phrase := range []string{"ordinary user-visible text", "same response before the tool calls", "what you will do and why", "Hidden thinking or a reasoning summary does not satisfy", "will be refused; retry"} {
				if !strings.Contains(prompt, phrase) {
					t.Errorf("missing prompt requirement %q: %q", phrase, prompt)
				}
			}
			if a.Engine().SystemPrompt() != prompt || a.Config().SystemPrompt != cfg.SystemPrompt {
				t.Fatal("prompt rendering mutated role state")
			}
			wire, err := a.Engine().Render(a.History().Context())
			if err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.Marshal(prompt)
			if strings.Count(string(wire), string(encoded)) != 1 {
				t.Fatalf("effective prompt absent or duplicated on wire: %s", wire)
			}
			cfg.SystemPrompt = prompt
			b := visibleReasoningPromptAgent(t, cfg)
			if b.Engine().SystemPrompt() != prompt {
				t.Fatal("restored effective prompt appended requirement twice")
			}
		})
	}
}

func TestVisibleReasoningPromptPreservesPrimaryAndDefault(t *testing.T) {
	dir := skillFiles(t, map[string]string{
		"empty": "---\nname: empty\ntype: primary\n---\n",
		"role":  "---\nname: role\ntype: primary\n---\nPrimary role wins.",
	})
	for _, model := range []string{"unknown", "claude-sonnet-5", "gemini-3.8-flash", "gpt-ch19-course", "gpt-4o"} {
		t.Run(model, func(t *testing.T) {
			cfg := ensemble.Config{Model: model, SystemPrompt: "Overridden config role"}
			a := visibleReasoningPromptAgent(t, cfg, ensemble.AgentSpec{SkillsDir: dir, PrimarySkill: "empty"})
			prompt := a.Engine().SystemPrompt()
			if model != "gpt-4o" {
				if prompt != "" {
					t.Fatalf("intentionally empty unrelated primary changed: %q", prompt)
				}
			} else if !strings.HasPrefix(prompt, "Before calling tools,") || strings.Contains(prompt, cfg.SystemPrompt) {
				t.Fatalf("empty primary must receive only capability policy: %q", prompt)
			}
			b := visibleReasoningPromptAgent(t, cfg, ensemble.AgentSpec{SkillsDir: dir, PrimarySkill: "role"})
			prompt = b.Engine().SystemPrompt()
			if !strings.HasPrefix(prompt, "Primary role wins.") || strings.Contains(prompt, cfg.SystemPrompt) {
				t.Fatalf("primary role replaced by config: %q", prompt)
			}
			if model != "gpt-4o" && prompt != "Primary role wins." {
				t.Fatalf("unrelated primary changed: %q", prompt)
			}
			c := visibleReasoningPromptAgent(t, cfg)
			if model != "gpt-4o" && c.Engine().SystemPrompt() != cfg.SystemPrompt {
				t.Fatal("unrelated explicit prompt changed")
			}
		})
	}
	base := visibleReasoningPromptAgent(t, ensemble.Config{Model: "unknown"}).Engine().SystemPrompt()
	required := visibleReasoningPromptAgent(t, ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o"})
	if base == "" || !strings.HasPrefix(required.Engine().SystemPrompt(), base+"\n\n") {
		t.Fatal("legacy default prompt was not preserved")
	}
	if err := required.SetModel("gpt-unknown"); err != nil {
		t.Fatal(err)
	}
	if required.Engine().SystemPrompt() != base {
		t.Fatal("capability instruction leaked after switching to unknown model")
	}
}

// Child role assembly must store authored instructions, not model-policy output.
// Otherwise a real OpenAI child keeps the gate instruction after model changes.
func TestVisibleReasoningPromptChildModelSwitch(t *testing.T) {
	root := ensemble.New(io.Discard)
	parent, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o", SystemPrompt: "Inherited role.", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = parent.Shutdown() })
	cfg := parent.Config()
	cfg.DataDir = t.TempDir()
	child, err := root.NewChildAgent(parent, cfg, ensemble.AgentSpec{ID: "child"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = child.Shutdown() })
	if strings.Contains(child.Config().SystemPrompt, "Before calling tools,") {
		t.Fatal("child role persisted model-specific policy")
	}
	if strings.Count(child.Engine().SystemPrompt(), "Before calling tools,") != 1 {
		t.Fatal("child missing or duplicating current-model requirement")
	}
	cfg = child.Config()
	cfg.DataDir = t.TempDir()
	grandchild, err := root.NewChildAgent(child, cfg, ensemble.AgentSpec{ID: "grandchild"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = grandchild.Shutdown() })
	if strings.Count(grandchild.Engine().SystemPrompt(), "You are a delegated sub-agent.") != 1 || strings.Count(grandchild.Engine().SystemPrompt(), "Before calling tools,") != 1 {
		t.Fatal("grandchild duplicated role or model instruction")
	}
	for _, a := range []ensemble.Agent{child, grandchild} {
		if err := a.SetModel("gpt-unknown"); err != nil {
			t.Fatal(err)
		}
		prompt := a.Engine().SystemPrompt()
		if strings.Contains(prompt, "Before calling tools,") || !strings.HasPrefix(prompt, "Inherited role.") || strings.Count(prompt, "You are a delegated sub-agent.") != 1 {
			t.Fatalf("model switch changed authored role or leaked policy: %q", prompt)
		}
	}
}
