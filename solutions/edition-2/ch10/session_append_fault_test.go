package ensemble

import (
	"bytes"
	"errors"
	"example.com/ensemble/internal/common"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
)

// The root's test adapter retains the real log and owner chain. It injects one
// actual partial physical write at the same append boundary used by public Append.
type partialSessionLog struct {
	common.EventLog
	parent common.Agent
	path   string
	closes atomic.Int32
}

func (l *partialSessionLog) Agent() common.Agent { return l.parent }
func (l *partialSessionLog) AppendPrepared(common.PreparedEvent) error {
	f, err := os.OpenFile(l.path, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	n, err := f.Write([]byte{'{'})
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if n != 1 {
		return errors.New("partial-write fixture did not write its byte")
	}
	return io.ErrShortWrite
}
func (l *partialSessionLog) Close() error { l.closes.Add(1); return l.EventLog.Close() }

type countedSessionCodec struct {
	common.SessionCodec
	encodes atomic.Int32
}

func (c *countedSessionCodec) Encode(cp common.Checkpoint) ([]byte, error) {
	c.encodes.Add(1)
	return c.SessionCodec.Encode(cp)
}

func TestSessionPublicAppendFaultClose(t *testing.T) {
	for _, saved := range []bool{false, true} {
		name := "no_checkpoint"
		if saved {
			name = "existing_checkpoint"
		}
		t.Run(name, func(t *testing.T) {
			root := New(io.Discard)
			defer root.Close()
			dir := t.TempDir()
			options := SessionOptions{Config: Config{DataDir: filepath.Join(dir, "session"), Workspace: dir, Model: "fixture", APIKey: "fixture"}}
			a, err := root.OpenSession(options)
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			if err = a.Append(Event{Type: "message_received", Message: &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text("retained")}}}); err != nil {
				t.Fatal(err)
			}
			exported, err := a.ExportCheckpoint()
			if err != nil {
				t.Fatal(err)
			}
			checkpoint := filepath.Join(options.Config.DataDir, "checkpoint.json")
			var beforeFile os.FileInfo
			var beforeBytes []byte
			if saved {
				if _, err = a.Checkpoint(); err != nil {
					t.Fatal(err)
				}
				held, err := os.Open(checkpoint)
				if err != nil {
					t.Fatal(err)
				}
				defer held.Close()
				beforeFile, err = held.Stat()
				if err != nil {
					t.Fatal(err)
				}
				beforeBytes, err = io.ReadAll(held)
				if err != nil {
					t.Fatal(err)
				}
			}
			before := a.Snapshot()
			beforeEvents := a.Events()
			logBefore, err := os.ReadFile(a.Config().LogPath)
			if err != nil {
				t.Fatal(err)
			}
			codec := &countedSessionCodec{SessionCodec: a.codec}
			a.codec = codec
			fault := &partialSessionLog{EventLog: a.log, parent: a, path: a.Config().LogPath}
			a.log = fault
			appendErr := a.Append(Event{Type: "message_received", Message: &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text("not accepted")}}})
			if !errors.Is(appendErr, io.ErrShortWrite) {
				t.Fatalf("append fault not reached: %v", appendErr)
			}
			if !reflect.DeepEqual(before, a.Snapshot()) || !reflect.DeepEqual(beforeEvents, a.Events()) {
				t.Error("partial append changed accepted state/history")
			}
			if _, err = a.Checkpoint(); err == nil {
				t.Error("checkpoint admitted after terminal append failure")
			}
			for i := 0; i < 2; i++ {
				if err = a.Close(); !errors.Is(err, appendErr) {
					t.Errorf("close %d did not retain original append error: %v", i, err)
				}
			}
			if codec.encodes.Load() != 0 {
				t.Errorf("faulted Agent attempted %d checkpoint encodes", codec.encodes.Load())
			}
			if fault.closes.Load() != 1 {
				t.Errorf("log closed %d times", fault.closes.Load())
			}
			afterFile, err := os.Stat(checkpoint)
			if saved {
				if err != nil || !os.SameFile(beforeFile, afterFile) {
					t.Errorf("checkpoint replaced despite fault: %v", err)
				}
				after, err := os.ReadFile(checkpoint)
				if err != nil || !bytes.Equal(beforeBytes, after) {
					t.Errorf("checkpoint bytes changed: %v", err)
				}
			} else if !os.IsNotExist(err) {
				t.Errorf("faulted close created checkpoint: %v", err)
			}
			logAfter, err := os.ReadFile(a.Config().LogPath)
			if err != nil || !bytes.Equal(logAfter, append(logBefore, '{')) {
				t.Errorf("partial log changed on close: %v", err)
			}
			_, err = root.OpenSession(options)
			var problem *SessionError
			if !errors.As(err, &problem) || problem.Code != "session_corrupt" {
				t.Errorf("lock/path not released or corrupt log accepted: %v", err)
			}
			options.Config.DataDir = filepath.Join(dir, "relocated")
			imported, err := root.ImportSession(exported.Bytes, options)
			if err != nil {
				t.Fatalf("SessionID reservation not released: %v", err)
			}
			if err = imported.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSessionPublicValidationRefusalStaysHealthy(t *testing.T) {
	root := New(io.Discard)
	defer root.Close()
	dir := t.TempDir()
	a, err := root.OpenSession(SessionOptions{Config: Config{DataDir: filepath.Join(dir, "session"), Workspace: dir, Model: "fixture", APIKey: "fixture"}})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	before := a.Snapshot()
	if err = a.Append(Event{Type: "tool_returned", Tool: &ToolEvent{CallID: "absent", Parts: []Part{Text("bad")}}}); err == nil {
		t.Fatal("invalid append accepted")
	}
	if !reflect.DeepEqual(before, a.Snapshot()) {
		t.Fatal("refusal changed accepted state")
	}
	if err = a.Append(Event{Type: "message_received", Message: &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text("valid after refusal")}}}); err != nil {
		t.Fatal("validation refusal faulted Agent", err)
	}
	if _, err = a.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err = a.Close(); err != nil {
			t.Fatal("healthy repeated close", err)
		}
	}
}
