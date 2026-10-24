package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"ensemble"
)

func TestChapter14ReopenPreservesFactsAndEffectiveMenus(t *testing.T) {
	for _, tc := range []struct {
		name   string
		tools  []string
		skills bool
	}{
		{name: "default"}, {name: "selected", tools: []string{"read_file"}}, {name: "empty", tools: []string{}}, {name: "recovered-skills", skills: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4", APIKey: "local-fixture-only", BaseDir: t.TempDir(), DataDir: t.TempDir(), SavePath: filepath.Join(t.TempDir(), "session.save"), Tools: tc.tools}
			if tc.skills {
				cfg.Skills = &ensemble.SkillSpec{Directory: c9Fixture(t), Primary: "base"}
			}
			first := ensemble.New(nil)
			firstClosed := false
			defer func() {
				if !firstClosed {
					first.Shutdown()
				}
			}()
			a, err := first.NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if tc.skills {
				if err := a.Skills().LoadSkill("writer"); err != nil {
					t.Fatal(err)
				}
			}
			if err := a.Save(); err != nil {
				t.Fatal(err)
			}
			encode := func(v any) []byte {
				t.Helper()
				data, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				return data
			}
			events, projection, menu := encode(a.Log().Events()), encode(a.Log().Context()), encode(a.Tools().Definitions())
			if err := first.Shutdown(); err != nil {
				t.Fatal(err)
			}
			firstClosed = true
			second := ensemble.New(nil)
			defer second.Shutdown()
			restored, err := second.NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if got := encode(restored.Log().Events()); !bytes.Equal(got, events) {
				t.Fatalf("construction changed canonical facts:\nbefore=%s\nafter=%s", events, got)
			}
			if got := encode(restored.Log().Context()); !bytes.Equal(got, projection) {
				t.Fatalf("construction changed recovered projection:\nbefore=%s\nafter=%s", projection, got)
			}
			if got := encode(restored.Tools().Definitions()); !bytes.Equal(got, menu) {
				t.Fatalf("recovered effective menu changed:\nbefore=%s\nafter=%s", menu, got)
			}
			// Suppressing construction publication must not suppress later live changes.
			boundary := restored.Log().LastSeq()
			err = restored.Tools().Register(ensemble.ToolDefinition{Name: "reopen_probe", Description: "Local registration probe", Parameters: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)}, func(context.Context, ensemble.ToolCall, json.RawMessage) error { return nil })
			if err != nil {
				t.Fatal(err)
			}
			delta := restored.Log().EventsRange(boundary, restored.Log().LastSeq())
			if len(delta) != 1 || delta[0].Type != ensemble.ToolsChanged || delta[0].Tools == nil || len(delta[0].Tools.Removed) != 0 || len(delta[0].Tools.Added) != 1 || delta[0].Tools.Added[0].Name != "reopen_probe" {
				t.Fatalf("live registration lost or polluted its delta: %+v", delta)
			}
		})
	}
}
