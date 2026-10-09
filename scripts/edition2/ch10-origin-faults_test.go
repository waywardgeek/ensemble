package ensemble

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCh10AdmissionOriginFaults(t *testing.T) {
	a, _ := ch10RemainingOwner(t)
	x, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"origin-pass", "origin-short", "origin-write", "origin-sync", "origin-close"} {
		t.Run(mode, func(t *testing.T) {
			root := New(io.Discard)
			t.Cleanup(func() { _ = root.Close() })
			w := t.TempDir()
			o := SessionOptions{Config: Config{Vendor: "openai", Model: "fixture", APIKey: "origin-noncredential", BaseURL: "http://127.0.0.1:1", Workspace: w, DataDir: filepath.Join(w, mode)}}
			if _, err = root.InspectCheckpoint(x.Bytes); err != nil {
				t.Fatal("genuine origin parent refused", err)
			}
			b, err := root.ImportSession(x.Bytes, o)
			if mode == "origin-pass" {
				if err != nil {
					t.Fatal("no-fault origin wrapper refused", err)
				}
				if err = b.Close(); err != nil {
					t.Fatal(err)
				}
				if _, err = root.InspectSession(o.Config.DataDir); err != nil {
					t.Fatal(err)
				}
			} else {
				if b != nil || err == nil {
					if b != nil {
						_ = b.Close()
					}
					t.Fatal("origin fault exposed live Agent")
				}
				ch10RemainingCode(t, err, "session_io")
			}
			operations, err := os.ReadFile(filepath.Join(w, mode+".operations"))
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("mode=%s operations=%q", mode, operations)
			if strings.Count(string(operations), "open ") != 1 || strings.Count(string(operations), "write ") != 1 || strings.Count(string(operations), "close ") != 1 {
				t.Fatal("origin descriptor operation counts differ")
			}
			syncs := strings.Count(string(operations), "sync ")
			if (mode == "origin-short" || mode == "origin-write") && syncs != 0 {
				t.Fatal("failed origin write proceeded to sync")
			}
			if mode != "origin-short" && mode != "origin-write" && syncs != 1 {
				t.Fatal("origin sync boundary was not reached")
			}
			path := filepath.Join(o.Config.DataDir, "origin.json")
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want := x.Bytes
			if mode == "origin-short" || mode == "origin-write" {
				want = want[:len(want)/2]
			}
			if !bytes.Equal(original, want) {
				t.Fatal("origin partial bytes differ from actual writer effect")
			}
			if mode == "origin-pass" {
				return
			}
			for _, leaf := range []string{"events.log", "checkpoint.json"} {
				if _, err = os.Stat(filepath.Join(o.Config.DataDir, leaf)); !os.IsNotExist(err) {
					t.Fatal("origin fault admitted later durable construction", leaf, err)
				}
			}
			_, err = root.OpenSession(o)
			ch10RemainingCode(t, err, "session_corrupt")
			_, err = root.InspectSession(o.Config.DataDir)
			ch10RemainingCode(t, err, "session_corrupt")
			preserved, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(preserved, original) {
				t.Fatal("incomplete origin was rewritten after refusal")
			}
			// Same root and SessionID can import elsewhere: failed construction released
			// its identity reservation. The corrupt reopen above also proves path unlock.
			o.Config.DataDir = filepath.Join(w, "recovered")
			recovered, err := root.ImportSession(x.Bytes, o)
			if err != nil {
				t.Fatal("origin fault retained reservation", err)
			}
			if err = recovered.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
