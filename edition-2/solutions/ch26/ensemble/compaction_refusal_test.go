package ensemble_test

import (
	"ensemble/ensemble"
	"ensemble/internal/common"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestRootFactoryPurposeRefusal checks refusal without starting inference.
func TestRootFactoryPurposeRefusal(t *testing.T) {
	for _, managed := range []bool{true, false} {
		name := "managed-control"
		if !managed {
			name = "ordinary-return"
		}
		t.Run(name, func(t *testing.T) {
			app := ensemble.New(io.Discard)
			oldLog := filepath.Join(t.TempDir(), "events.jsonl")
			before := "prior owner log bytes\n"
			if err := os.WriteFile(oldLog, []byte(before), 0600); err != nil {
				t.Fatal(err)
			}
			var child ensemble.Agent
			app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
				spec.Compaction = managed
				spec.LogPath = oldLog
				a, err := app.NewChildAgent(parent, parent.Config(), spec)
				child = a
				return a, err
			})
			parent, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: "http://127.0.0.1:1", DataDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { parent.Shutdown() })
			if _, _, err := parent.Fleet().StartCompaction("fresh", "source"); err == nil {
				t.Fatal("invalid factory accepted")
			}
			if child == nil {
				t.Fatal("factory did not construct child")
			}
			after, err := os.ReadFile(oldLog)
			if err != nil || string(after) != before {
				t.Errorf("refusal overwrote existing log: got %q, err %v", after, err)
			}
			if app.ContainsAgent(child) {
				t.Error("refusal retained rejected child in root")
			}
		})
	}
}

// TestOrdinaryShutdownStillPersists distinguishes refusal disposal from the
// existing valid ordinary child shutdown and root-retention contract.
func TestOrdinaryShutdownStillPersists(t *testing.T) {
	app := ensemble.New(io.Discard)
	parent, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { parent.Shutdown() })
	child, err := app.CreateChild(parent, common.AgentSpec{ID: "ordinary", DataDir: filepath.Join(parent.Config().DataDir, "ordinary"), LogPath: filepath.Join(parent.Config().DataDir, "ordinary.log")})
	if err != nil {
		t.Fatal(err)
	}
	if err := child.Shutdown(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(child.Config().LogPath)
	if err != nil || len(data) == 0 {
		t.Fatalf("ordinary shutdown did not save: %q %v", data, err)
	}
	if !app.ContainsAgent(child) {
		t.Fatal("ordinary shutdown removed root child")
	}
}
