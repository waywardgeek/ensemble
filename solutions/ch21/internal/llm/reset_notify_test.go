package llm

import (
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

type captureObserver struct {
	got []common.Observation
}

func (c *captureObserver) Observe(o common.Observation) { c.got = append(c.got, o) }

// Reset is recorded as an event, so a client that replays the log rebuilds the
// cleared screen by itself. A client that is ALREADY connected replays
// nothing: it learns about the agent only through observations. Recording
// without notifying leaves the stale transcript on an open screen until the
// operator happens to refresh, which is exactly the bug this guards.
func TestResetTellsAlreadyConnectedClients(t *testing.T) {
	eng := NewEngine(common.Config{}, filepath.Join(t.TempDir(), "journal.jsonl"), nil, nil, nil)
	a := NewActor(eng, nil)

	obs := &captureObserver{}
	a.Attach(obs)

	a.handleReset()

	for _, o := range obs.got {
		if _, ok := o.(common.ConversationCleared); ok {
			return
		}
	}
	t.Fatalf("reset notified %v, want a ConversationCleared so open clients clear", obs.got)
}
