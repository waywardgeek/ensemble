package ensemble

import (
	"context"
	"errors"
	"example.com/ensemble/internal/common"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// The composition root replaces its own codec with a gated test double. Store
// still reaches it through its actual Agent parent, without injected siblings.
type gatedSessionCodec struct {
	common.SessionCodec
	once             sync.Once
	entered, release chan struct{}
}

func (c *gatedSessionCodec) Encode(cp common.Checkpoint) ([]byte, error) {
	c.once.Do(func() { close(c.entered); <-c.release })
	return c.SessionCodec.Encode(cp)
}
func TestSessionCaptureTailAndCanceledWait(t *testing.T) {
	root := New(io.Discard)
	defer root.Close()
	dir := t.TempDir()
	a, err := root.OpenSession(SessionOptions{Config: Config{DataDir: filepath.Join(dir, "session"), Model: "fixture", APIKey: "fixture", Workspace: dir}})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Append(Event{Type: "message_received", Message: &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text("before")}}}); err != nil {
		t.Fatal(err)
	}
	gate := &gatedSessionCodec{SessionCodec: a.codec, entered: make(chan struct{}), release: make(chan struct{})}
	a.codec = gate
	before := a.Snapshot().LastSeq
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := a.CheckpointContext(ctx); done <- err }()
	<-gate.entered
	if _, err = a.ExportCheckpoint(); err == nil {
		t.Fatal("second capture admitted")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled waiter retained by store")
	}
	appendDone := make(chan error, 1)
	go func() {
		appendDone <- a.Append(Event{Type: "message_received", Message: &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text("accepted tail")}}})
	}()
	select {
	case err := <-appendDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("checkpoint worker parked Actor")
	}
	close(gate.release)
	deadline := time.After(time.Second)
	for {
		s, err := a.Session()
		if err != nil {
			t.Fatal(err)
		}
		if s.CheckpointSeq != nil {
			if *s.CheckpointSeq != before {
				t.Fatal("ack included later tail")
			}
			break
		}
		select {
		case <-deadline:
			t.Fatal("save did not finish after canceled wait")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, "session/checkpoint.json"))
	if err != nil {
		t.Fatal(err)
	}
	inspection, err := root.InspectCheckpoint(raw)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Boundary().LogSeq != before || len(inspection.Snapshot().Instructions) != 1 {
		t.Fatal("captured value raced with later instruction")
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	if err = a.Close(); err != nil {
		t.Fatal("repeated close", err)
	}
}
