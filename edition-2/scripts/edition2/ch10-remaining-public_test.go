package ensemble

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"example.com/ensemble/internal/common"
)

func TestCh10RemainingPublicConstructionFacts(t *testing.T) {
	a, _ := ch10RemainingOwner(t)
	x, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	initializer := a.Events()[0]
	e := New(io.Discard)
	t.Cleanup(func() { _ = e.Close() })
	w := t.TempDir()
	imported, err := e.ImportSession(x.Bytes, SessionOptions{Config: Config{Vendor: "openai", Model: "remaining-local", APIKey: "fixture-not-a-credential", BaseURL: "http://127.0.0.1:1", Workspace: w, DataDir: filepath.Join(w, "imported")}})
	if err != nil {
		t.Fatal(err)
	}
	anchor := imported.Events()[0]
	if initializer.Type != "session_initialized" || anchor.Type != "session_anchor" {
		t.Fatal("genuine construction facts missing")
	}
	for _, event := range []Event{initializer, anchor} {
		for _, mounted := range []bool{false, true} {
			name := event.Type
			if mounted {
				name += "-mounted"
			} else {
				name += "-fresh"
			}
			t.Run(name, func(t *testing.T) {
				workspace := t.TempDir()
				config := Config{Vendor: "openai", Model: "remaining-local", APIKey: "fixture-not-a-credential", BaseURL: "http://127.0.0.1:1", Workspace: workspace}
				var target *Agent
				var err error
				if mounted {
					config.DataDir = filepath.Join(workspace, "store")
					target, err = e.OpenSession(SessionOptions{Config: config})
				} else {
					config.LogPath = filepath.Join(workspace, "standalone.log")
					target, err = e.NewAgent(config)
				}
				if err != nil {
					t.Fatal(err)
				}
				path := target.Config().LogPath
				before, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				count := len(target.Events())
				if err = target.Append(event); err == nil {
					t.Fatal("public append adopted a construction-only session fact")
				}
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(before, after) || len(target.Events()) != count {
					t.Fatal("refused construction fact changed accepted/durable state")
				}
				if err = target.Append(Event{Type: "message_received", Message: &common.Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text("still usable")}}}); err != nil {
					t.Fatalf("construction refusal became terminal: %v", err)
				}
				if err = target.Close(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestCh10RemainingStoreLeaves(t *testing.T) {
	a, _ := ch10RemainingOwner(t)
	x, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	for _, leaf := range []string{"events.log", "checkpoint.json", "origin.json", "owner.lock"} {
		for _, kind := range []string{"symlink", "directory"} {
			t.Run(leaf+"-"+kind, func(t *testing.T) {
				e := New(io.Discard)
				t.Cleanup(func() { _ = e.Close() })
				w := t.TempDir()
				o := SessionOptions{Config: Config{Vendor: "openai", Model: "remaining-local", APIKey: "fixture-not-a-credential", BaseURL: "http://127.0.0.1:1", Workspace: w, DataDir: filepath.Join(w, "store")}}
				parent, err := e.ImportSession(x.Bytes, o)
				if err != nil {
					t.Fatal(err)
				}
				if err = parent.Close(); err != nil {
					t.Fatal(err)
				}
				if _, err = e.InspectSession(o.Config.DataDir); err != nil {
					t.Fatalf("genuine imported parent: %v", err)
				}
				path := filepath.Join(o.Config.DataDir, leaf)
				before, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				preserved := path + ".preserved"
				if err = os.Rename(path, preserved); err != nil {
					t.Fatal(err)
				}
				if kind == "symlink" {
					err = os.Symlink(preserved, path)
				} else {
					err = os.Mkdir(path, 0700)
				}
				if err != nil {
					t.Fatal(err)
				}
				_, err = e.OpenSession(o)
				ch10RemainingCode(t, err, "session_corrupt")
				_, err = e.InspectSession(o.Config.DataDir)
				ch10RemainingCode(t, err, "session_corrupt")
				after, err := os.ReadFile(preserved)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatal("leaf refusal rewrote source")
				}
				if err = os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err = os.Rename(preserved, path); err != nil {
					t.Fatal(err)
				}
				recovered, err := e.OpenSession(o)
				if err != nil {
					t.Fatalf("failed construction retained reservation: %v", err)
				}
				if err = recovered.Close(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
