package ensemble

// Reachable public construction and fault combinations. The model is a local
// deterministic HTTP peer; source stores and descriptors remain real.
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"example.com/ensemble/internal/eventlog"
)

type ch10EdgeCall struct{ name, args string }

func ch10EdgeServer(t *testing.T, calls ...ch10EdgeCall) (*httptest.Server, *atomic.Int32) {
	return ch10EdgeNamedServer(t, "edge", calls...)
}

func ch10EdgeNamedServer(t *testing.T, prefix string, calls ...ch10EdgeCall) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	count := &atomic.Int32{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		n := int(count.Add(1))
		w.Header().Set("Content-Type", "application/json")
		if n%2 == 0 {
			_, _ = io.WriteString(w, `{"model":"gpt-4.1-mini-2025-04-14","choices":[{"index":0,"message":{"role":"assistant","content":"complete"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
			return
		}
		i := (n - 1) / 2
		if i >= len(calls) {
			http.Error(w, "unexpected extra request", 500)
			return
		}
		c := calls[i]
		_ = json.NewEncoder(w).Encode(map[string]any{"model": "gpt-4.1-mini-2025-04-14", "choices": []any{map[string]any{"index": 0, "message": map[string]any{"role": "assistant", "content": "", "tool_calls": []any{map[string]any{"id": fmt.Sprintf("%s-%d", prefix, i), "type": "function", "function": map[string]any{"name": c.name, "arguments": c.args}}}}, "finish_reason": "tool_calls"}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1}})
	}))
	t.Cleanup(s.Close)
	return s, count
}

func ch10EdgeOptions(t *testing.T, url string, builtins ...string) SessionOptions {
	t.Helper()
	w := t.TempDir()
	return SessionOptions{Config: Config{Vendor: "openai", Model: "gpt-4.1-mini-2025-04-14", APIKey: "edge-noncredential", BaseURL: url, DisableStreaming: true, Workspace: w, DataDir: filepath.Join(w, "session"), Builtins: builtins}}
}

func ch10EdgeTurn(t *testing.T, a *Agent, success bool) Completion {
	t.Helper()
	h, err := a.Submit("Perform the next controlled local attempt.")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	c, err := h.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if success && c.Outcome != "success" || !success && c.Outcome == "success" {
		t.Fatalf("unexpected turn outcome: %+v", c)
	}
	return c
}

type ch10EdgeWriter struct {
	io.WriteCloser
	fail bool
	hits int
}

func (w *ch10EdgeWriter) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte(`"type":"tool_called"`)) {
		w.hits++
		if w.fail {
			return 0, errors.New("review failure after durable limit consumption")
		}
	}
	return w.WriteCloser.Write(p)
}

func TestCh10EdgesConsumedLimitsAppendFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			server, count := ch10EdgeServer(t, ch10EdgeCall{"tool_limits", `{"max_output_bytes":7}`}, ch10EdgeCall{"not_installed", `{}`})
			r := New(io.Discard)
			t.Cleanup(func() { _ = r.Close() })
			o := ch10EdgeOptions(t, server.URL, "tool_limits")
			a, err := r.OpenSession(o)
			if err != nil {
				t.Fatal(err)
			}
			ch10EdgeTurn(t, a, true)
			if a.jobs.PendingLimits() == nil {
				t.Fatal("genuine setter did not establish pending limits")
			}
			if _, err = a.Checkpoint(); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(o.Config.DataDir, "checkpoint.json")
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var writer *ch10EdgeWriter
			a.log.(*eventlog.Log).Ch10ReviewWrapWriter(func(real io.WriteCloser) io.WriteCloser {
				writer = &ch10EdgeWriter{WriteCloser: real, fail: fail}
				return writer
			})
			ch10EdgeTurn(t, a, !fail)
			if writer.hits != 1 {
				t.Fatalf("intended post-consumption append hit count: %d", writer.hits)
			}
			if a.jobs.PendingLimits() != nil {
				t.Fatal("durably consumed setting became pending after attempted dispatch")
			}
			consumed, called := 0, 0
			for _, e := range a.Events() {
				if e.Type == "tool_limits_consumed" && e.Limits.CallID == "edge-1" {
					consumed++
				}
				if e.Type == "tool_called" && e.Tool.CallID == "edge-1" {
					called++
				}
			}
			if consumed != 1 || fail && called != 0 || !fail && called != 1 {
				t.Fatalf("wrong accepted consumption/dispatch: %d/%d", consumed, called)
			}
			if !fail {
				if count.Load() != 4 {
					t.Fatal("positive did not continue after refused unknown tool")
				}
				if err = a.Close(); err != nil {
					t.Fatal(err)
				}
				return
			}
			if count.Load() != 3 {
				t.Fatal("append failure sent another model request")
			}
			if _, err = a.Submit("must remain terminal"); err == nil {
				t.Fatal("failed append retained admission")
			}
			if err = a.Close(); err == nil {
				t.Fatal("append failure disappeared at close")
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("append failure replaced prior checkpoint")
			}
			inspection, err := r.InspectSession(o.Config.DataDir)
			if err != nil || inspection.Boundary().Settled {
				t.Fatalf("complete prefix must remain inspectable and unfinished: %v", err)
			}
			if inspection.agent.jobs.PendingLimits() != nil {
				t.Fatal("offline replay restored consumed setting")
			}
			_, err = r.OpenSession(o)
			ch10RemainingCode(t, err, "session_unfinished")
			if count.Load() != 3 {
				t.Fatal("unfinished resume sent HTTP")
			}
		})
	}
}

func TestCh10EdgesRootJobFloor(t *testing.T) {
	server, count := ch10EdgeServer(t, ch10EdgeCall{"run_command", `{"command":"printf floor-done"}`})
	origin := New(io.Discard)
	t.Cleanup(func() { _ = origin.Close() })
	for i := uint64(1); i <= 100; i++ {
		if origin.AllocateHandle() != i {
			t.Fatal("public root allocation control")
		}
	}
	o := ch10EdgeOptions(t, server.URL, "run_command")
	a, err := origin.OpenSession(o)
	if err != nil {
		t.Fatal(err)
	}
	ch10EdgeTurn(t, a, true)
	if len(a.Snapshot().Jobs) != 1 || a.Snapshot().Jobs[101].Status != "done" {
		t.Fatal("real job did not establish durable handle101")
	}
	x, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	for _, burned := range []uint64{0, 200} {
		r := New(io.Discard)
		t.Cleanup(func() { _ = r.Close() })
		for i := uint64(1); i <= burned; i++ {
			if r.AllocateHandle() != i {
				t.Fatal("target root allocation control")
			}
		}
		selected := ch10EdgeOptions(t, server.URL, "run_command")
		b, err := r.ImportSession(x.Bytes, selected)
		if err != nil {
			t.Fatal(err)
		}
		want := max(burned, 101) + 1
		if got := r.AllocateHandle(); got != want {
			t.Fatalf("root allocator floor: got%d want%d", got, want)
		}
		if b.Snapshot().Jobs[101].Status != "done" {
			t.Fatal("historical job changed during import")
		}
		if err = b.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if count.Load() != 2 {
		t.Fatal("historical import sent HTTP")
	}
}

func TestCh10EdgesPartialStoreTopology(t *testing.T) {
	a, _ := ch10RemainingOwner(t)
	x, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	full := a.Config().DataDir
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	r := New(io.Discard)
	t.Cleanup(func() { _ = r.Close() })
	o := ch10EdgeOptions(t, "http://127.0.0.1:1")
	imported, err := r.ImportSession(x.Bytes, o)
	if err != nil {
		t.Fatal(err)
	}
	if err = imported.Close(); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{full, o.Config.DataDir} {
		if _, err = r.InspectSession(source); err != nil {
			t.Fatalf("genuine topology parent: %v", err)
		}
	}
	cases := []struct{ name, source, remove, add, code string }{
		{"checkpoint-without-log", full, "events.log", "", "session_corrupt"},
		{"origin-without-log", o.Config.DataDir, "events.log", "", "session_corrupt"},
		{"anchor-without-origin", o.Config.DataDir, "origin.json", "", "session_origin_required"},
		{"full-origin-with-extra-origin", full, "", "origin.json", "session_corrupt"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := filepath.Join(t.TempDir(), "store")
			if err := os.Mkdir(d, 0700); err != nil {
				t.Fatal(err)
			}
			for _, leaf := range []string{"events.log", "checkpoint.json", "origin.json"} {
				b, err := os.ReadFile(filepath.Join(c.source, leaf))
				if os.IsNotExist(err) {
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(d, leaf), b, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if c.remove != "" {
				if err := os.Remove(filepath.Join(d, c.remove)); err != nil {
					t.Fatal(err)
				}
			}
			if c.add != "" {
				if err := os.WriteFile(filepath.Join(d, c.add), x.Bytes, 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := map[string][]byte{}
			for _, leaf := range []string{"events.log", "checkpoint.json", "origin.json"} {
				b, err := os.ReadFile(filepath.Join(d, leaf))
				if os.IsNotExist(err) {
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				before[leaf] = b
			}
			selected := ch10EdgeOptions(t, "http://127.0.0.1:1")
			selected.Config.DataDir = d
			_, err = r.OpenSession(selected)
			ch10RemainingCode(t, err, c.code)
			_, err = r.InspectSession(d)
			ch10RemainingCode(t, err, c.code)
			for _, leaf := range []string{"events.log", "checkpoint.json", "origin.json"} {
				b, err := os.ReadFile(filepath.Join(d, leaf))
				original, exists := before[leaf]
				if !exists {
					if !os.IsNotExist(err) {
						t.Fatalf("refusal created %s", leaf)
					}
					continue
				}
				if err != nil || !bytes.Equal(b, original) {
					t.Fatalf("refusal rewrote %s", leaf)
				}
			}
			// Repair only the fixture mutation and prove construction reservations
			// were released. The runtime never repairs the refused source itself.
			if c.add != "" {
				if err = os.Remove(filepath.Join(d, c.add)); err != nil {
					t.Fatal(err)
				}
			}
			if c.remove != "" {
				b, err := os.ReadFile(filepath.Join(c.source, c.remove))
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(d, c.remove), b, 0600); err != nil {
					t.Fatal(err)
				}
			}
			recovered, err := r.OpenSession(selected)
			if err != nil {
				t.Fatalf("failed topology mount retained reservation: %v", err)
			}
			if err = recovered.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
