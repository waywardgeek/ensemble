package ensemble

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/eventlog"
)

// These controls wrap real files. Every injected failure/barrier must be hit,
// and a successful parent and deliberate retry establish the same path works.
type ch10ReviewIO struct {
	common.CheckpointIO
	mu               sync.Mutex
	fail, gate       string
	hits             map[string]int
	entered, release chan struct{}
	once             sync.Once
}

func (d *ch10ReviewIO) step(name string) error {
	d.mu.Lock()
	d.hits[name]++
	fail, gate := d.fail == name, d.gate == name
	d.mu.Unlock()
	if gate {
		d.once.Do(func() { close(d.entered); <-d.release })
	}
	if fail {
		return errors.New("independent injected checkpoint failure")
	}
	return nil
}
func (d *ch10ReviewIO) count(name string) int { d.mu.Lock(); defer d.mu.Unlock(); return d.hits[name] }
func (d *ch10ReviewIO) CreateTemp(dir string) (common.CheckpointFile, error) {
	if err := d.step("create"); err != nil {
		return nil, err
	}
	f, err := d.CheckpointIO.CreateTemp(dir)
	if err != nil {
		return nil, err
	}
	return &ch10ReviewFile{CheckpointFile: f, parent: d}, nil
}
func (d *ch10ReviewIO) Rename(from, to string) error {
	if err := d.step("replace"); err != nil {
		return err
	}
	return d.CheckpointIO.Rename(from, to)
}

type ch10ReviewFile struct {
	common.CheckpointFile
	parent *ch10ReviewIO
}

func (f *ch10ReviewFile) IO() common.CheckpointIO { return f.parent }
func (f *ch10ReviewFile) Write(b []byte) (int, error) {
	if err := f.parent.step("write"); err != nil {
		return 0, err
	}
	f.parent.mu.Lock()
	short := f.parent.fail == "short"
	f.parent.mu.Unlock()
	if short {
		_ = f.parent.step("short")
		return f.CheckpointFile.Write(b[:len(b)-1])
	}
	return f.CheckpointFile.Write(b)
}
func (f *ch10ReviewFile) Sync() error {
	if err := f.parent.step("sync"); err != nil {
		return err
	}
	return f.CheckpointFile.Sync()
}
func (f *ch10ReviewFile) Close() error {
	err := f.CheckpointFile.Close()
	if injected := f.parent.step("close"); injected != nil {
		return injected
	}
	return err
}
func ch10ReviewOpen(t *testing.T) (*Ensemble, *Agent, SessionOptions) {
	t.Helper()
	dir := t.TempDir()
	r := New(io.Discard)
	o := SessionOptions{Config: Config{DataDir: filepath.Join(dir, "session"), Workspace: dir, Model: "fixture", APIKey: "not-a-real-key"}}
	a, err := r.OpenSession(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r, a, o
}
func ch10ReviewAppend(a *Agent, text string) error {
	return a.Append(Event{Type: "message_received", Message: &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text(text)}}})
}
func ch10ReviewBytes(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func ch10ReviewCode(t *testing.T, err error, want string) {
	t.Helper()
	var e *SessionError
	if !errors.As(err, &e) || e.Code != want {
		t.Fatalf("wanted %s, got %v", want, err)
	}
}
func ch10ReviewWait[T any](t *testing.T, ch <-chan T, label string) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(4 * time.Second):
		t.Fatalf("barrier did not complete: %s", label)
		var zero T
		return zero
	}
}
func ch10ReviewWrap(a *Agent, fail, gate string) *ch10ReviewIO {
	d := &ch10ReviewIO{fail: fail, gate: gate, hits: map[string]int{}, entered: make(chan struct{}), release: make(chan struct{})}
	a.store.Ch10ReviewWrapIO(func(real common.CheckpointIO) common.CheckpointIO { d.CheckpointIO = real; return d })
	return d
}

func TestCh10ReviewCheckpointFaults(t *testing.T) {
	for _, point := range []string{"none", "create", "write", "short", "sync", "close", "replace"} {
		t.Run(point, func(t *testing.T) {
			r, a, o := ch10ReviewOpen(t)
			first, err := a.Checkpoint()
			if err != nil {
				t.Fatal("positive parent", err)
			}
			path := filepath.Join(o.Config.DataDir, "checkpoint.json")
			before := ch10ReviewBytes(t, path)
			if err = ch10ReviewAppend(a, "accepted after original checkpoint"); err != nil {
				t.Fatal(err)
			}
			d := ch10ReviewWrap(a, point, "")
			ack, err := a.Checkpoint()
			if point == "none" {
				if err != nil || ack.AsOf <= first.AsOf || d.count("replace") != 1 {
					t.Fatalf("positive write control: %+v %v", ack, err)
				}
			} else {
				ch10ReviewCode(t, err, "session_io")
				if d.count(point) != 1 {
					t.Fatalf("fault hook %s was not hit exactly once", point)
				}
				if !bytes.Equal(before, ch10ReviewBytes(t, path)) {
					t.Fatal("failed checkpoint changed committed bytes")
				}
				state, err := a.Session()
				if err != nil || state.CheckpointSeq == nil || *state.CheckpointSeq != first.AsOf {
					t.Fatalf("failed checkpoint claimed an anchor: %+v %v", state, err)
				}
				d.mu.Lock()
				d.fail = ""
				d.mu.Unlock()
				if err = ch10ReviewAppend(a, "healthy log after failed save"); err != nil {
					t.Fatal(err)
				}
				ack, err = a.Checkpoint()
				if err != nil || ack.AsOf <= first.AsOf {
					t.Fatalf("deliberate retry: %+v %v", ack, err)
				}
			}
			inspection, err := r.InspectCheckpoint(ch10ReviewBytes(t, path))
			if err != nil || inspection.Boundary().LogSeq != ack.AsOf {
				t.Fatalf("new committed checkpoint: %v", err)
			}
			left, err := filepath.Glob(filepath.Join(o.Config.DataDir, ".checkpoint-*"))
			if err != nil || len(left) != 0 {
				t.Fatalf("uncommitted temporaries retained: %v %v", left, err)
			}
		})
	}
}

func TestCh10ReviewCanceledWaitKeepsCapturedWrite(t *testing.T) {
	r, a, o := ch10ReviewOpen(t)
	if err := ch10ReviewAppend(a, "inside captured prefix"); err != nil {
		t.Fatal(err)
	}
	boundary := a.Snapshot().LastSeq
	d := ch10ReviewWrap(a, "", "replace")
	var release sync.Once
	unblock := func() { release.Do(func() { close(d.release) }) }
	defer unblock()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := a.CheckpointContext(ctx); done <- err }()
	ch10ReviewWait(t, d.entered, "actual replacement boundary")
	_, err := a.Checkpoint()
	ch10ReviewCode(t, err, "session_busy")
	_, err = a.ExportCheckpoint()
	ch10ReviewCode(t, err, "session_busy")
	cancel()
	if err = ch10ReviewWait(t, done, "canceled caller"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel waiter: %v", err)
	}
	appendDone := make(chan error, 1)
	go func() { appendDone <- ch10ReviewAppend(a, "outside captured prefix") }()
	if err = ch10ReviewWait(t, appendDone, "same Agent actor responsiveness"); err != nil {
		t.Fatal(err)
	}
	otherDone := make(chan error, 1)
	go func() {
		other := o
		other.Config.DataDir = filepath.Join(o.Config.Workspace, "other")
		b, e := r.OpenSession(other)
		if e == nil {
			_, e = b.Checkpoint()
		}
		otherDone <- e
	}()
	if err = ch10ReviewWait(t, otherDone, "other Agent store responsiveness"); err != nil {
		t.Fatal(err)
	}
	unblock()
	deadline := time.Now().Add(4 * time.Second)
	for {
		s, e := a.Session()
		if e != nil {
			t.Fatal(e)
		}
		if s.CheckpointSeq != nil {
			if *s.CheckpointSeq != boundary {
				t.Fatal("save included later tail")
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("canceled wait canceled committed worker")
		}
		time.Sleep(time.Millisecond)
	}
	inspection, err := r.InspectCheckpoint(ch10ReviewBytes(t, filepath.Join(o.Config.DataDir, "checkpoint.json")))
	if err != nil || inspection.Boundary().LogSeq != boundary || len(inspection.Snapshot().Instructions) != 1 {
		t.Fatalf("owned prefix changed while writer blocked: %v", err)
	}
	if a.Snapshot().LastSeq <= boundary || len(a.Snapshot().Instructions) != 2 {
		t.Fatal("accepted tail lost")
	}
	if d.count("replace") != 1 {
		t.Fatal("barrier did not hit one real write")
	}
}

func TestCh10ReviewCloseJoinsWriterBeforeUnlock(t *testing.T) {
	_, a, o := ch10ReviewOpen(t)
	d := ch10ReviewWrap(a, "", "replace")
	var release sync.Once
	unblock := func() { release.Do(func() { close(d.release) }) }
	defer unblock()
	save := make(chan error, 1)
	go func() { _, err := a.Checkpoint(); save <- err }()
	ch10ReviewWait(t, d.entered, "checkpoint replace before close")
	_, watch, err := a.Watch()
	if err != nil {
		t.Fatal(err)
	}
	defer watch.Close()
	closed := make(chan error, 1)
	go func() { closed <- a.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	for {
		record, e := watch.Next(ctx)
		if e != nil {
			t.Fatal("close did not publish its admission", e)
		}
		if record.Observation.Kind == "state" && record.Observation.State == "stopping" {
			break
		}
	}
	// The competitor is an actual second root using the OS lock, not an in-root
	// reservation check. Completion of this refusal is the ordering observation.
	competitor := New(io.Discard)
	defer competitor.Close()
	_, err = competitor.OpenSession(o)
	ch10ReviewCode(t, err, "session_in_use")
	select {
	case err := <-closed:
		t.Fatalf("close returned before held writer: %v", err)
	default:
	}
	unblock()
	if err = ch10ReviewWait(t, save, "first save completion"); err != nil {
		t.Fatal(err)
	}
	if err = ch10ReviewWait(t, closed, "close completion"); err != nil {
		t.Fatal(err)
	}
	if err = a.Close(); err != nil {
		t.Fatal("repeat close", err)
	}
	b, err := competitor.OpenSession(o)
	if err != nil {
		t.Fatal("lock not released after join", err)
	}
	if s, e := b.Session(); e != nil || s.CheckpointSeq == nil {
		t.Fatalf("final checkpoint absent: %+v %v", s, e)
	}
	if d.count("replace") < 2 {
		t.Fatal("close omitted its final checkpoint after joining the prior one")
	}
}

type ch10ReviewPartialWriter struct {
	io.WriteCloser
	calls int
}

func (w *ch10ReviewPartialWriter) Write(p []byte) (int, error) {
	w.calls++
	return w.WriteCloser.Write(p[:1])
}
func TestCh10ReviewAppendFailurePreservesCheckpoint(t *testing.T) {
	_, a, o := ch10ReviewOpen(t)
	if _, err := a.Checkpoint(); err != nil {
		t.Fatal("positive parent", err)
	}
	path := filepath.Join(o.Config.DataDir, "checkpoint.json")
	before := ch10ReviewBytes(t, path)
	snapshot := a.Snapshot()
	events := a.Events()
	disk := ch10ReviewWrap(a, "", "")
	log, ok := a.log.(*eventlog.Log)
	if !ok {
		t.Fatal("review adapter no longer matches real event log")
	}
	var partial *ch10ReviewPartialWriter
	log.Ch10ReviewWrapWriter(func(real io.WriteCloser) io.WriteCloser {
		partial = &ch10ReviewPartialWriter{WriteCloser: real}
		return partial
	})
	if err := ch10ReviewAppend(a, "must not become accepted"); err == nil {
		t.Fatal("short append succeeded")
	}
	if partial.calls != 1 {
		t.Fatal("actual partial write not injected exactly once")
	}
	if !reflect.DeepEqual(snapshot, a.Snapshot()) || !reflect.DeepEqual(events, a.Events()) {
		t.Fatal("failed append changed accepted state/history")
	}
	if err := a.Close(); err == nil {
		t.Error("faulted close hid persistence failure")
	}
	if disk.count("create") != 0 || disk.count("replace") != 0 {
		t.Errorf("faulted close attempted final checkpoint: create=%d replace=%d", disk.count("create"), disk.count("replace"))
	}
	if !bytes.Equal(before, ch10ReviewBytes(t, path)) {
		t.Fatal("faulted close replaced last usable checkpoint")
	}
	raw := ch10ReviewBytes(t, filepath.Join(o.Config.DataDir, "events.log"))
	if raw[len(raw)-1] != '{' {
		t.Fatal("physical failed-write artifact missing")
	}
	other := New(io.Discard)
	defer other.Close()
	_, err := other.OpenSession(o)
	ch10ReviewCode(t, err, "session_corrupt")
	if !bytes.Equal(raw, ch10ReviewBytes(t, filepath.Join(o.Config.DataDir, "events.log"))) {
		t.Fatal("failed resume rewrote partial record")
	}
}
