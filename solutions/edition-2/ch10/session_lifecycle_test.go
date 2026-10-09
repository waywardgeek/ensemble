package ensemble

import (
	"context"
	"errors"
	"example.com/ensemble/internal/common"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
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

func TestSessionUnsupportedWatermarksRefuseBeforeImport(t *testing.T) {
	root := New(io.Discard)
	defer root.Close()
	dir := t.TempDir()
	options := SessionOptions{Config: Config{DataDir: filepath.Join(dir, "original"), Model: "fixture", APIKey: "fixture", Workspace: dir}}
	a, err := root.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	export, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = root.InspectCheckpoint(export.Bytes); err != nil {
		t.Fatal("positive parent", err)
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"job", "activation"} {
		t.Run(field, func(t *testing.T) {
			cp, err := a.codec.Decode(export.Bytes)
			if err != nil {
				t.Fatal(err)
			}
			if field == "job" {
				cp.HighWatermarks.Job = 1
			} else {
				cp.HighWatermarks.Activation = 1
			}
			cp.State.Session.HighWatermarks = cp.HighWatermarks
			raw, err := a.codec.Encode(cp)
			if err != nil {
				t.Fatal(err)
			}
			var problem *SessionError
			if _, err = root.InspectCheckpoint(raw); !errors.As(err, &problem) || problem.Code != "session_corrupt" {
				t.Fatalf("unsupported maximum: %v", err)
			}
			dest := filepath.Join(dir, field)
			options.Config.DataDir = dest
			if _, err = root.ImportSession(raw, options); !errors.As(err, &problem) || problem.Code != "session_corrupt" {
				t.Fatalf("import: %v", err)
			}
			for _, name := range []string{"origin.json", "checkpoint.json", "events.log"} {
				if _, err = os.Stat(filepath.Join(dest, name)); !os.IsNotExist(err) {
					t.Fatalf("refused import wrote %s: %v", name, err)
				}
			}
		})
	}
}

func TestSessionAcceptedJobTailDuringCheckpoint(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			io.WriteString(w, `{"model":"fixture","content":[{"type":"tool_use","id":"worker","name":"run_command","input":{"command":"while [ ! -e release-worker ]; do sleep 0.01; done; printf finished","ai_callback_delay":0}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			io.WriteString(w, `{"model":"fixture","content":[{"type":"text","text":"job retained"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer server.Close()
	root := New(io.Discard)
	defer root.Close()
	dir := t.TempDir()
	a, err := root.OpenSession(SessionOptions{Config: Config{DataDir: filepath.Join(dir, "session"), Workspace: dir, Vendor: "anthropic", Model: "fixture", APIKey: "fixture", BaseURL: server.URL, DisableStreaming: true, Builtins: []string{"run_command"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Ask(context.Background(), "start controlled worker"); err != nil {
		t.Fatal(err)
	}
	before := a.Snapshot()
	if len(before.Jobs) != 1 {
		t.Fatal("no actual job")
	}
	var handle uint64
	for h, j := range before.Jobs {
		handle = h
		if j.Status != "running" {
			t.Fatal("positive parent is not running")
		}
	}
	gate := &gatedSessionCodec{SessionCodec: a.codec, entered: make(chan struct{}), release: make(chan struct{})}
	a.codec = gate
	released := false
	defer func() {
		if !released {
			close(gate.release)
		}
	}()
	saved := make(chan error, 1)
	go func() { _, err := a.Checkpoint(); saved <- err }()
	<-gate.entered
	if err = os.WriteFile(filepath.Join(dir, "release-worker"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(3 * time.Second)
	for a.Snapshot().Jobs[handle].Status == "running" {
		select {
		case <-deadline:
			t.Fatal("accepted job event parked behind checkpoint worker")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	if a.Snapshot().LastSeq <= before.LastSeq {
		t.Fatal("worker state was mistaken for accepted tail")
	}
	close(gate.release)
	released = true
	if err = <-saved; err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "session/checkpoint.json"))
	if err != nil {
		t.Fatal(err)
	}
	inspected, err := root.InspectCheckpoint(raw)
	if err != nil {
		t.Fatal(err)
	}
	if inspected.Boundary().LogSeq != before.LastSeq || inspected.Snapshot().Jobs[handle].Status != "running" {
		t.Fatal("checkpoint captured racing worker state")
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 2 {
		t.Fatal("unexpected model requests", requests.Load())
	}
}
