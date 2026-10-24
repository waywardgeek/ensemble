package history

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"ensemble/internal/common"
)

func TestChapter14ConcurrentAppendSaveMissingPrimary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.save")
	h := diskHistory(t, path)
	if _, err := h.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	appendRecord := func(text string) error {
		return h.Append(common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text(text)}}})
	}
	if err := appendRecord("initial"); err != nil {
		t.Fatal(err)
	}
	if err := h.SaveFile(path, archiveConfig(), common.LogFull); err != nil {
		t.Fatal(err)
	}
	stagedPath := filepath.Join(t.TempDir(), "admission.dump")
	start, done := make(chan struct{}), make(chan error, 2)
	go func() {
		<-start
		for i := 0; i < 48; i++ {
			text := fmt.Sprintf("writer-%02d", i)
			var err error
			if i%2 == 0 {
				event := common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text(text)}}}
				err = h.AppendSaved(event, func(dump func(io.Writer) error) error {
					// Even while another snapshot writer is queued, staging must leave
					// committed readers at the OLD frontier and keep its new fact private.
					if h.LastSeq() != common.Seq(i+1) || len(h.Events()) != i+1 {
						return fmt.Errorf("staged admission changed committed frontier at writer %d", i)
					}
					f, err := os.OpenFile(stagedPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
					if err != nil {
						return err
					}
					err = dump(f)
					if err == nil {
						err = f.Sync()
					}
					closeErr := f.Close()
					if err != nil {
						return err
					}
					return closeErr
				})
			} else {
				err = appendRecord(text)
			}
			if err != nil {
				done <- err
				return
			}
			runtime.Gosched()
		}
		done <- nil
	}()
	go func() {
		<-start
		for i := 0; i < 12; i++ {
			if err := h.SaveFile(path, archiveConfig(), common.LogFull); err != nil {
				done <- err
				return
			}
			runtime.Gosched()
		}
		done <- nil
	}()
	close(start)
	// Join both actual owner operations before inspecting their disk state.
	first, second := <-done, <-done
	if first != nil || second != nil {
		t.Fatalf("concurrent append/save: %v; %v", first, second)
	}
	if h.LastSeq() != 49 || len(h.Context().Dialogue) != 49 {
		t.Fatal("concurrent operations lost or duplicated events")
	}
	staged, err := os.ReadFile(stagedPath)
	if err != nil {
		t.Fatal(err)
	}
	prepared := New(nil)
	if err := prepared.Load(bytes.NewReader(staged)); err != nil {
		t.Fatal(err)
	}
	if prepared.LastSeq() != 48 {
		t.Fatalf("last staged file ended at %d, want 48", prepared.LastSeq())
	}
	// Make the backup older than the primary, with another event beyond both.
	if err := h.SaveFile(path, archiveConfig(), common.LogFull); err != nil {
		t.Fatal(err)
	}
	if err := appendRecord("covered-only-by-primary"); err != nil {
		t.Fatal(err)
	}
	if err := h.SaveFile(path, archiveConfig(), common.LogFull); err != nil {
		t.Fatal(err)
	}
	if err := appendRecord("journal-only"); err != nil {
		t.Fatal(err)
	}
	want := h.Context()
	check := func(restored common.History) {
		t.Helper()
		if restored.LastSeq() != 51 || !reflect.DeepEqual(restored.Context(), want) {
			t.Fatalf("restart differs: seq=%d dialogue=%d want seq=51 dialogue=%d", restored.LastSeq(), len(restored.Context().Dialogue), len(want.Dialogue))
		}
	}
	primary := diskHistory(t, path)
	if r, err := primary.LoadFile(path); err != nil || r.Backup {
		t.Fatalf("primary restart: %+v %v", r, err)
	}
	check(primary)
	// Preserve the owned primary bytes while simulating a missing pathname.
	if err := os.Rename(path, path+".held-primary"); err != nil {
		t.Fatal(err)
	}
	fallback := diskHistory(t, path)
	if r, err := fallback.LoadFile(path); err != nil || !r.Backup {
		t.Fatalf("missing-primary fallback: %+v %v", r, err)
	}
	check(fallback)
	if err := fallback.Append(common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text("after-restart")}}}); err != nil {
		t.Fatal(err)
	}
	again := diskHistory(t, path)
	if _, err := again.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if again.LastSeq() != 52 || !reflect.DeepEqual(again.Context(), fallback.Context()) {
		t.Fatal("continued append did not replay exactly once after fallback")
	}
}
