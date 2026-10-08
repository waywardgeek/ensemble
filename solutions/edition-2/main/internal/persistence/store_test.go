package persistence

import (
	"bytes"
	"errors"
	"example.com/ensemble/internal/common"
	"os"
	"path/filepath"
	"testing"
)

type storeAgent struct {
	common.SessionAgent
	codec common.SessionCodec
}

func (a *storeAgent) Ensemble() common.Ensemble  { return &testRoot{} }
func (a *storeAgent) Codec() common.SessionCodec { return a.codec }

type faultIO struct {
	common.CheckpointIO
	parent           common.SessionStore
	mode             string
	entered, release chan struct{}
}

func (d *faultIO) Store() common.SessionStore { return d.parent }
func (d *faultIO) CreateTemp(path string) (common.CheckpointFile, error) {
	if d.mode == "create" {
		return nil, errors.New("injected create")
	}
	f, err := d.CheckpointIO.CreateTemp(path)
	if err != nil {
		return nil, err
	}
	return &faultFile{CheckpointFile: f, parent: d}, nil
}
func (d *faultIO) Rename(from, to string) error {
	if d.mode == "rename" {
		return errors.New("injected rename")
	}
	return d.CheckpointIO.Rename(from, to)
}

type faultFile struct {
	common.CheckpointFile
	parent *faultIO
}

func (f *faultFile) IO() common.CheckpointIO { return f.parent }
func (f *faultFile) Write(data []byte) (int, error) {
	if f.parent.entered != nil {
		close(f.parent.entered)
		<-f.parent.release
	}
	switch f.parent.mode {
	case "write":
		return 0, errors.New("injected write")
	case "short":
		return len(data) - 1, nil
	}
	return f.CheckpointFile.Write(data)
}
func (f *faultFile) Sync() error {
	if f.parent.mode == "sync" {
		return errors.New("injected sync")
	}
	return f.CheckpointFile.Sync()
}
func (f *faultFile) Close() error {
	err := f.CheckpointFile.Close()
	if f.parent.mode == "close" {
		return errors.New("injected close")
	}
	return err
}
func storeFixture(t *testing.T) (*Store, common.Checkpoint) {
	t.Helper()
	a := &storeAgent{}
	a.codec = NewCodec(a)
	s, err := Open(a, t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	system := "fixture"
	cp := common.Checkpoint{Version: 1, StateVersion: 1, AsOf: 1, SessionID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Identity: common.SessionIdentity{Mode: "plain", System: &system, Handlers: []common.HandlerIdentity{}}, HighWatermarks: common.Watermarks{Event: 1}}
	return s, cp
}
func TestCheckpointWriteFaultsPreserveCommit(t *testing.T) {
	for _, mode := range []string{"create", "write", "short", "sync", "close", "rename"} {
		t.Run(mode, func(t *testing.T) {
			s, cp := storeFixture(t)
			done, err := s.Begin(cp, true)
			if err != nil {
				t.Fatal(err)
			}
			r := <-done
			if r.Error != nil || !r.Saved {
				t.Fatal(r.Error)
			}
			s.Applied(r)
			path := filepath.Join(s.path, "checkpoint.json")
			before, _ := os.ReadFile(path)
			original := s.io
			s.io = &faultIO{CheckpointIO: original, parent: s, mode: mode}
			cp.AsOf = 2
			cp.HighWatermarks.Event = 2
			done, err = s.Begin(cp, true)
			if err != nil {
				t.Fatal(err)
			}
			r = <-done
			s.Applied(r)
			after, _ := os.ReadFile(path)
			if r.Error == nil || r.Saved || !bytes.Equal(before, after) || *s.CheckpointSequence() != 1 {
				t.Fatal("failed replacement changed commit")
			}
			s.io = original
			done, err = s.Begin(cp, true)
			if err != nil {
				t.Fatal(err)
			}
			r = <-done
			s.Applied(r)
			if r.Error != nil || !r.Saved || *s.CheckpointSequence() != 2 {
				t.Fatal("failure stranded worker gate")
			}
		})
	}
}
func TestCheckpointGateAndLockLifetime(t *testing.T) {
	s, cp := storeFixture(t)
	d := &faultIO{CheckpointIO: s.io, parent: s, entered: make(chan struct{}), release: make(chan struct{})}
	s.io = d
	done, err := s.Begin(cp, true)
	if err != nil {
		t.Fatal(err)
	}
	<-d.entered
	if _, err = s.Begin(cp, false); err == nil {
		t.Fatal("export bypassed active save")
	}
	if err = s.Close(); err == nil {
		t.Fatal("close released active worker")
	}
	if other, err := Open(s.Agent(), s.path, false); err == nil {
		other.Close()
		t.Fatal("second lock owner accepted")
	}
	close(d.release)
	r := <-done
	s.Applied(r)
	if r.Error != nil {
		t.Fatal(r.Error)
	}
	before, err := os.Stat(filepath.Join(s.path, "owner.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	other, err := Open(s.Agent(), s.path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	after, _ := os.Stat(filepath.Join(s.path, "owner.lock"))
	if !os.SameFile(before, after) {
		t.Fatal("lock inode replaced")
	}
}
