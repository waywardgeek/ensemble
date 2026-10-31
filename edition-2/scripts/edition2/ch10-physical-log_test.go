package ensemble

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
)

func ch10PhysicalLogHash(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TestCh10PhysicalLogBoundary(t *testing.T) {
	a, _ := ch10RemainingOwner(t)
	dir := a.Config().DataDir
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	root := New(io.Discard)
	t.Cleanup(func() { _ = root.Close() })
	parent, err := root.InspectSession(dir)
	if err != nil {
		t.Fatal("genuine closed log parent refused", err)
	}
	log := filepath.Join(dir, "events.log")
	checkpoint := filepath.Join(dir, "checkpoint.json")
	checkpointHash := ch10PhysicalLogHash(t, checkpoint)
	info, err := os.Stat(log)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatal("parent log is not a regular file", err)
	}
	prefixSize := info.Size()
	f, err := os.OpenFile(log, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		t.Fatal(err)
	}
	// Every added byte is physically written. Each whitespace-only record is
	// at most 1 MiB, well below the unchanged inherited blank-record bound.
	chunk := bytes.Repeat([]byte{' '}, 1<<20)
	var written int64
	for remaining := int64(1<<30) - prefixSize; remaining > 0; {
		n := min(int64(len(chunk)), remaining)
		chunk[n-1] = '\n'
		count, writeErr := f.Write(chunk[:n])
		if writeErr != nil || count != int(n) {
			_ = f.Close()
			t.Fatal("physical fixture write failed", count, writeErr)
		}
		written += int64(count)
		remaining -= int64(count)
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	info, err = os.Stat(log)
	if err != nil || !info.Mode().IsRegular() || info.Size() != 1<<30 || prefixSize+written != 1<<30 {
		t.Fatal("physical exact-size fixture differs", err)
	}
	t.Logf("regular_file=%s logical_bytes=%d allocated_bytes=%d sequential_padding_bytes=%d", log, info.Size(), info.Sys().(*syscall.Stat_t).Blocks*512, written)
	for _, extra := range []bool{false, true} {
		if extra {
			f, err = os.OpenFile(log, os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				t.Fatal(err)
			}
			if n, writeErr := f.Write([]byte{' '}); writeErr != nil || n != 1 {
				_ = f.Close()
				t.Fatal("one-over fixture write failed", writeErr)
			}
			if err = f.Sync(); err != nil {
				_ = f.Close()
				t.Fatal(err)
			}
			if err = f.Close(); err != nil {
				t.Fatal(err)
			}
		}
		before := ch10PhysicalLogHash(t, log)
		got, inspectErr := root.InspectSession(dir)
		info, err = os.Stat(log)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("physical_bytes=%d sha256=%s inspect_error=%v", info.Size(), before, inspectErr)
		if !extra {
			if inspectErr != nil || !reflect.DeepEqual(got.Session(), parent.Session()) || !reflect.DeepEqual(got.Boundary(), parent.Boundary()) || !reflect.DeepEqual(got.Events(), parent.Events()) || !reflect.DeepEqual(got.Snapshot(), parent.Snapshot()) {
				t.Fatal("physical exact log boundary parent refused or changed semantics", inspectErr)
			}
		} else {
			if info.Size() != (1<<30)+1 {
				t.Fatal("physical one-over size differs")
			}
			if inspectErr == nil {
				t.Fatal("physical log one-over limit accepted")
			}
			ch10RemainingCode(t, inspectErr, "session_corrupt")
			if !strings.Contains(inspectErr.Error(), "session log exceeds 1 GiB") {
				t.Fatal("physical log bound not diagnosed", inspectErr)
			}
		}
		if ch10PhysicalLogHash(t, log) != before || ch10PhysicalLogHash(t, checkpoint) != checkpointHash {
			t.Fatal("physical inspection rewrote source files")
		}
	}
}
