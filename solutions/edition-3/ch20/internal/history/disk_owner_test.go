package history

import (
	"ensemble/internal/agent"
	"ensemble/internal/common"
	"io"
	"log"
	"path/filepath"
	"testing"
)

// Disk tests use a real, unstarted Agent and its rooted Sandbox. No actor or
// provider is needed; memory-only replay tests can continue to use New(nil).
func diskHistoryOwner(t *testing.T, path string) common.Agent {
	t.Helper()
	root := filepath.Dir(path)
	owner, err := agent.New(diagnosticEnsemble{logger: log.New(io.Discard, "", 0)}, common.Config{Model: "fixture", BaseDir: root, DataDir: root, SavePath: path, SandboxRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := owner.Sandbox().Close(); err != nil {
			t.Error(err)
		}
	})
	return owner
}
func diskHistory(t *testing.T, path string) common.History {
	t.Helper()
	return New(diskHistoryOwner(t, path))
}
