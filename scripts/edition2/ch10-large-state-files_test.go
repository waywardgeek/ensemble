package ensemble

// These tests allocate real 256/512 MiB inputs and must run alone after the
// coordinator releases a sufficiently sized disk/build slot. No 1 GiB file.
import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"example.com/ensemble/internal/common"
)

func ch10LargeHash(t *testing.T, path string) string {
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

func ch10LargePad(t *testing.T, path string, prefix []byte, size int64) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if n, err := f.Write(prefix); err != nil || n != len(prefix) {
		t.Fatal("fixture prefix write failed", err)
	}
	remaining := size - int64(len(prefix))
	if remaining < 0 {
		t.Fatal("padding fixture exceeds target before padding")
	}
	chunk := []byte(strings.Repeat(" ", 1<<20))
	for remaining > 0 {
		n := int64(len(chunk))
		if remaining < n {
			n = remaining
		}
		written, err := f.Write(chunk[:n])
		if err != nil || int64(written) != n {
			t.Fatal("fixture whitespace write failed", err)
		}
		remaining -= n
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil || st.Size() != size {
		t.Fatal("fixture physical size differs from intended boundary")
	}
}

func TestCh10LargeCanonicalState(t *testing.T) {
	a, _ := ch10RemainingOwner(t)
	for i := 0; i < 104; i++ {
		if err := a.Append(Event{Type: "message_received", Message: &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text("")}}}); err != nil {
			t.Fatal(err)
		}
	}
	x, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	cp, err := a.Codec().Decode(x.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if len(cp.State.Context.Instructions) != 104 || len(cp.State.Window.Events) != 100 || cp.AsOf != 105 {
		t.Fatal("genuine104-entry/window parent differs")
	}
	var envelope map[string]json.RawMessage
	if err = json.Unmarshal(x.Bytes, &envelope); err != nil {
		t.Fatal(err)
	}
	small, err := a.Codec().Canonical(envelope["state"])
	if err != nil {
		t.Fatal(err)
	}
	r := New(io.Discard)
	t.Cleanup(func() { _ = r.Close() })
	options := func() SessionOptions {
		w := t.TempDir()
		return SessionOptions{Config: Config{Vendor: "openai", Model: "remaining-local", APIKey: "large-noncredential", BaseURL: "http://127.0.0.1:1", Workspace: w, DataDir: filepath.Join(w, "imported")}}
	}
	// A real small import measures the anchor's semantic overhead. The live
	// large import must reserve this space; import does not preserve state size.
	calibration, err := r.ImportSession(x.Bytes, options())
	if err != nil {
		t.Fatal(err)
	}
	calibrated, err := calibration.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	var calibratedEnvelope map[string]json.RawMessage
	if err = json.Unmarshal(calibrated.Bytes, &calibratedEnvelope); err != nil {
		t.Fatal(err)
	}
	afterAnchor, err := a.Codec().Canonical(calibratedEnvelope["state"])
	if err != nil {
		t.Fatal(err)
	}
	anchorDelta := len(afterAnchor) - len(small)
	if anchorDelta < 0 || anchorDelta > 1<<20 {
		t.Fatalf("unexpected genuine anchor size delta: %d", anchorDelta)
	}
	t.Logf("genuine small parent canonical_bytes=%d anchor_delta=%d", len(small), anchorDelta)
	if err = calibration.Close(); err != nil {
		t.Fatal(err)
	}
	// The separately verified canonical codec sizes only this small parent.
	// Each inserted ASCII x adds exactly one canonical byte. Oldest four entries
	// are outside the genuine100-event window, so their bodies are not duplicated.
	remaining := (256 << 20) - len(small)
	if remaining < 4 {
		t.Fatal("unexpected large baseline")
	}
	for i := 0; i < 4; i++ {
		n := remaining / (4 - i)
		remaining -= n
		entry := &cp.State.Context.Instructions[i]
		body := strings.Repeat("x", n)
		entry.Parts[0].Text = &body
		probe := *entry
		probe.Seq = 0
		probe.Parts = []Part{Text("")}
		encoded, err := json.Marshal(Event{Seq: entry.Seq, Type: "message_received", Time: "2026-10-08T00:00:00Z", Message: &probe})
		if err != nil || len(encoded)+n+1 > 64<<20 {
			t.Fatal("large semantic fixture requires an invalid physical event")
		}
	}
	if remaining != 0 {
		t.Fatal("canonical byte arithmetic did not consume exact budget")
	}
	high, err := a.Codec().Encode(cp)
	if err != nil {
		t.Fatalf("exact256MiB canonical state refused: %v", err)
	}
	if _, err = r.InspectCheckpoint(high); err != nil {
		t.Fatalf("exact canonical public inspection refused: %v", err)
	}
	t.Log("exact256MiB canonical state encoded and publicly inspected")
	before := *cp.State.Context.Instructions[0].Parts[0].Text
	oneOver := before + "x"
	cp.State.Context.Instructions[0].Parts[0].Text = &oneOver
	_, err = a.Codec().Encode(cp)
	if err == nil {
		t.Fatal("canonical state one-over limit accepted")
	}
	ch10RemainingCode(t, err, "session_limit")
	trimmed := before[:len(before)-anchorDelta]
	cp.State.Context.Instructions[0].Parts[0].Text = &trimmed
	high, err = a.Codec().Encode(cp)
	if err != nil {
		t.Fatal(err)
	}
	o := options()
	b, err := r.ImportSession(high, o)
	if err != nil {
		t.Fatalf("exact canonical public import refused: %v", err)
	}
	high, cp.State, before, oneOver, trimmed = nil, nil, "", "", ""
	runtime.GC()
	ack, err := b.Checkpoint()
	if err != nil || ack.AsOf != 106 {
		t.Fatalf("exact canonical imported checkpoint refused: %+v %v", ack, err)
	}
	t.Log("calibrated exact256MiB imported state checkpoint committed")
	checkpoint := filepath.Join(o.Config.DataDir, "checkpoint.json")
	origin := filepath.Join(o.Config.DataDir, "origin.json")
	checkpointHash, originHash := ch10LargeHash(t, checkpoint), ch10LargeHash(t, origin)
	appendInstruction := func(text string) {
		if err := b.Append(Event{Type: "message_received", Message: &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text(text)}}}); err != nil {
			t.Fatal(err)
		}
	}
	appendInstruction("one more accepted entry")
	logPath := filepath.Join(o.Config.DataDir, "events.log")
	logHash := ch10LargeHash(t, logPath)
	_, err = b.ExportCheckpoint()
	ch10RemainingCode(t, err, "session_limit")
	_, err = b.Checkpoint()
	ch10RemainingCode(t, err, "session_limit")
	if ch10LargeHash(t, checkpoint) != checkpointHash || ch10LargeHash(t, origin) != originHash || ch10LargeHash(t, logPath) != logHash {
		t.Fatal("snapshot limit damaged prior checkpoint, origin or log")
	}
	appendInstruction("still healthy after refused checkpoint")
	ch10RemainingCode(t, b.Close(), "session_limit")
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCh10LargePhysicalStateFiles(t *testing.T) {
	for _, leaf := range []string{"checkpoint.json", "origin.json"} {
		t.Run(leaf, func(t *testing.T) {
			a, _ := ch10RemainingOwner(t)
			x, err := a.ExportCheckpoint()
			if err != nil {
				t.Fatal(err)
			}
			directory := a.Config().DataDir
			if err = a.Close(); err != nil {
				t.Fatal(err)
			}
			r := New(io.Discard)
			t.Cleanup(func() { _ = r.Close() })
			if leaf == "origin.json" {
				w := t.TempDir()
				o := SessionOptions{Config: Config{Vendor: "openai", Model: "remaining-local", APIKey: "large-noncredential", BaseURL: "http://127.0.0.1:1", Workspace: w, DataDir: filepath.Join(w, "imported")}}
				b, err := r.ImportSession(x.Bytes, o)
				if err != nil {
					t.Fatal(err)
				}
				if err = b.Close(); err != nil {
					t.Fatal(err)
				}
				directory = o.Config.DataDir
			}
			if _, err = r.InspectSession(directory); err != nil {
				t.Fatalf("small genuine store parent refused: %v", err)
			}
			path := filepath.Join(directory, leaf)
			prefix, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			logPath := filepath.Join(directory, "events.log")
			for _, size := range []int64{512 << 20, (512 << 20) + 1} {
				ch10LargePad(t, path, prefix, size)
				fileHash := ch10LargeHash(t, path)
				if leaf == "origin.json" {
					log, err := os.ReadFile(logPath)
					if err != nil {
						t.Fatal(err)
					}
					lines := strings.Split(strings.TrimSuffix(string(log), "\n"), "\n")
					if len(lines) != 2 {
						t.Fatal("genuine empty imported log is not header plus anchor")
					}
					var anchor common.Event
					if err = json.Unmarshal([]byte(lines[1]), &anchor); err != nil || anchor.Type != "session_anchor" || anchor.Session == nil {
						t.Fatal("genuine import anchor unavailable", err)
					}
					anchor.Session.OriginSHA256 = fileHash
					raw, err := json.Marshal(anchor)
					if err != nil {
						t.Fatal(err)
					}
					if err = os.WriteFile(logPath, []byte(lines[0]+"\n"+string(raw)+"\n"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				logHash := ch10LargeHash(t, logPath)
				_, err = r.InspectSession(directory)
				if size == 512<<20 {
					if err != nil {
						t.Fatalf("exact512MiB %s refused: %v", leaf, err)
					}
				} else {
					if err == nil {
						t.Fatal("physical state file one-over limit accepted")
					}
					ch10RemainingCode(t, err, "session_corrupt")
					if !strings.Contains(err.Error(), "512 MiB") {
						t.Fatalf("file bound not identified: %v", err)
					}
				}
				t.Logf("physical leaf=%s bytes=%d sha256=%s inspection_error=%v", leaf, size, fileHash, err)
				if ch10LargeHash(t, path) != fileHash || ch10LargeHash(t, logPath) != logHash {
					t.Fatal("file-limit inspection rewrote source")
				}
				runtime.GC()
			}
		})
	}
}
