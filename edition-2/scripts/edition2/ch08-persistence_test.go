package gui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"example.com/ensemble"
	"github.com/gorilla/websocket"
)

type ch08Domain struct {
	name, path          string
	agent               *ensemble.Agent
	server              *Server
	arm                 func(string, bool) (<-chan struct{}, chan struct{})
	update              func(uint64, json.RawMessage) error
	snapshot            func() any
	close               func()
	seed, change, retry json.RawMessage
}

func ch08Value(t *testing.T, v any) map[string]any {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	var out map[string]any
	if e = json.Unmarshal(b, &out); e != nil {
		t.Fatal(e)
	}
	return out
}
func ch08DomainFor(t *testing.T, name string) *ch08Domain {
	t.Helper()
	dir := t.TempDir()
	app := ensemble.New(io.Discard)
	a, e := app.NewAgent(ensemble.Config{Vendor: "openai", Model: "fixture-ch08-disk", APIKey: "LOCAL-ONLY", Workspace: dir, LogPath: filepath.Join(dir, "events.log"), PolicyPath: filepath.Join(dir, "policy.json")})
	if e != nil {
		t.Fatal(e)
	}
	s, e := NewServer(app, a.ID(), "http://127.0.0.1:1234", io.Discard, ServerOptions{PreferencesPath: filepath.Join(dir, "preferences.json")})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close(); app.Close() })
	h := &ch08Domain{name: name, path: filepath.Join(dir, name+".json"), agent: a, server: s}
	if name == "policy" {
		h.arm = func(stage string, fail bool) (<-chan struct{}, chan struct{}) {
			c := a.Ch08ArmPolicy(stage, fail)
			return c.Entered, c.Release
		}
		h.update = func(base uint64, p json.RawMessage) error { _, e := a.UpdatePolicy(base, p); return e }
		h.snapshot = func() any { return a.ExecutionPolicy() }
		h.close = func() { _ = a.Close() }
		h.seed = json.RawMessage(`{"max_model_requests":1}`)
		h.change = json.RawMessage(`{"max_model_requests":2}`)
		h.retry = json.RawMessage(`{"max_model_requests":3}`)
	} else {
		h.arm = func(stage string, fail bool) (<-chan struct{}, chan struct{}) {
			c := s.preferences.Ch08Arm(stage, fail)
			return c.Entered, c.Release
		}
		h.update = func(base uint64, p json.RawMessage) error { _, e := s.Preferences().Update(base, p); return e }
		h.snapshot = func() any { return s.Preferences().Snapshot() }
		h.close = func() { s.Preferences().Close() }
		h.seed = json.RawMessage(`{"theme":"light"}`)
		h.change = json.RawMessage(`{"font_size":22}`)
		h.retry = json.RawMessage(`{"actions_width":440}`)
	}
	if e = h.update(0, h.seed); e != nil {
		t.Fatal(e)
	}
	return h
}
func ch08Hold(t *testing.T, h *ch08Domain, stage string, fail bool) (<-chan struct{}, chan struct{}) {
	t.Helper()
	entered, release := h.arm(stage, fail)
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	return entered, release
}
func ch08Wait(t *testing.T, done <-chan struct{}, reason string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal(reason)
	}
}
func ch08Result(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case e := <-done:
		return e
	case <-time.After(time.Second):
		t.Fatal("settings operation did not settle")
		return nil
	}
}
func ch08Code(t *testing.T, e error, want string) {
	t.Helper()
	var value *ensemble.SettingsError
	if !errors.As(e, &value) || value.Code != want {
		t.Fatalf("wanted %s, got %v", want, e)
	}
}
func ch08Read(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func ch08Revision(t *testing.T, h *ch08Domain, n int) {
	t.Helper()
	if ch08Value(t, h.snapshot())["revision"] != float64(n) {
		t.Fatalf("%s applied revision differs", h.name)
	}
}

func TestCh08PersistenceStages(t *testing.T) {
	for _, name := range []string{"policy", "preferences"} {
		for _, stage := range []string{"create", "write", "short-write", "sync", "close", "replace"} {
			positive := false
			for _, fail := range []bool{false, true} {
				label := "positive"
				if fail {
					label = "fault"
				}
				t.Run(name+"/"+stage+"/"+label, func(t *testing.T) {
					if fail && !positive {
						t.Fatal("fault requires the matching successful real disk stage")
					}
					h := ch08DomainFor(t, name)
					old := ch08Read(t, h.path)
					snapshot := ch08Value(t, h.snapshot())
					entered, release := ch08Hold(t, h, stage, fail)
					done := make(chan error, 1)
					go func() { done <- h.update(1, h.change) }()
					ch08Wait(t, entered, "writer never reached intended disk stage")
					if !reflect.DeepEqual(snapshot, ch08Value(t, h.snapshot())) || string(ch08Read(t, h.path)) != string(old) {
						t.Fatal("candidate applied or original changed before replace")
					}
					close(release)
					err := ch08Result(t, done)
					if fail {
						ch08Code(t, err, "settings_persist_failed")
						if !reflect.DeepEqual(snapshot, ch08Value(t, h.snapshot())) || string(ch08Read(t, h.path)) != string(old) {
							t.Fatal("pre-replace failure changed applied state or original bytes")
						}
					} else {
						if err != nil {
							t.Fatal(err)
						}
						ch08Revision(t, h, 2)
						if string(ch08Read(t, h.path)) == string(old) {
							t.Fatal("positive did not replace file")
						}
						positive = true
					}
					pending, e := filepath.Glob(filepath.Join(filepath.Dir(h.path), "."+name+"-*"))
					if e != nil || len(pending) != 0 {
						t.Fatal("uncommitted temporary file retained", pending, e)
					}
				})
			}
		}
	}
}

func TestCh08PersistenceHeldControlsAndRetry(t *testing.T) {
	for _, name := range []string{"policy", "preferences"} {
		t.Run(name, func(t *testing.T) {
			h := ch08DomainFor(t, name)
			entered, release := ch08Hold(t, h, "replace", false)
			done := make(chan error, 1)
			go func() { done <- h.update(1, h.change) }()
			ch08Wait(t, entered, "replace barrier missing")
			busy := make(chan error, 1)
			go func() { busy <- h.update(1, h.retry) }()
			ch08Code(t, ch08Result(t, busy), "settings_busy")
			ch08Revision(t, h, 1)
			controls := make(chan error, 1)
			go func() {
				_, watch, e := h.agent.Watch()
				if e != nil {
					controls <- e
					return
				}
				defer watch.Close()
				pause, e := h.agent.RegisterPause()
				if e != nil {
					controls <- e
					return
				}
				defer pause.Close()
				_, e = pause.Update(true, false)
				if e == nil {
					_, e = h.agent.Interrupt()
				}
				controls <- e
			}()
			if e := ch08Result(t, controls); e != nil {
				t.Fatal(e)
			}
			independent := make(chan error, 1)
			go func() {
				if name == "policy" {
					_, e := h.server.Preferences().Update(0, json.RawMessage(`{"theme":"light"}`))
					independent <- e
				} else {
					_, e := h.agent.UpdatePolicy(0, json.RawMessage(`{"max_model_requests":4}`))
					independent <- e
				}
			}()
			if e := ch08Result(t, independent); e != nil {
				t.Fatal("other domain did not progress", e)
			}
			select {
			case <-done:
				t.Fatal("acknowledged while replacement still held")
			default:
			}
			close(release)
			if e := ch08Result(t, done); e != nil {
				t.Fatal(e)
			}
			ch08Revision(t, h, 2)
			ch08Code(t, h.update(1, h.retry), "revision_conflict")
			if e := h.update(2, h.retry); e != nil {
				t.Fatal(e)
			}
			ch08Revision(t, h, 3)
			if name == "preferences" {
				v := h.server.Preferences().Snapshot().Preferences
				if v.Theme != "light" || v.FontSize != 22 || v.ActionsWidth != 440 {
					t.Fatal("fresh-base retry erased earlier sparse update")
				}
			}
		})
	}
}

func TestCh08PersistenceCloseJoinsActualOutcome(t *testing.T) {
	for _, name := range []string{"policy", "preferences"} {
		for _, mode := range []string{"replace-success", "replace-failure", "committed"} {
			t.Run(name+"/"+mode, func(t *testing.T) {
				h := ch08DomainFor(t, name)
				old := ch08Read(t, h.path)
				stage := "replace"
				if mode == "committed" {
					stage = mode
				}
				entered, release := ch08Hold(t, h, stage, mode == "replace-failure")
				done := make(chan error, 1)
				go func() { done <- h.update(1, h.change) }()
				ch08Wait(t, entered, "positive close/write barrier absent")
				ch08Revision(t, h, 1)
				if mode == "committed" && string(ch08Read(t, h.path)) == string(old) {
					t.Fatal("post-commit fixture did not reach real replacement")
				}
				closed := make(chan struct{})
				go func() { h.close(); close(closed) }()
				deadline := time.Now().Add(time.Second)
				for {
					e := h.update(1, h.retry)
					var setting *ensemble.SettingsError
					if errors.As(e, &setting) && setting.Code == "settings_closed" {
						break
					}
					if !errors.As(e, &setting) || setting.Code != "settings_busy" {
						t.Fatal("close admission gave unintended result", e)
					}
					if time.Now().After(deadline) {
						t.Fatal("close did not refuse new writes")
					}
					time.Sleep(time.Millisecond)
				}
				select {
				case <-closed:
					t.Fatal("close returned before owned writer joined")
				default:
				}
				close(release)
				err := ch08Result(t, done)
				ch08Wait(t, closed, "close did not join released writer")
				if mode == "replace-failure" {
					ch08Code(t, err, "settings_persist_failed")
					ch08Revision(t, h, 1)
					if string(ch08Read(t, h.path)) != string(old) {
						t.Fatal("failed writer replaced old file")
					}
				} else {
					if err != nil {
						t.Fatal("committed write falsely reported rollback", err)
					}
					ch08Revision(t, h, 2)
					if string(ch08Read(t, h.path)) == string(old) {
						t.Fatal("committed write disappeared on close")
					}
				}
			})
		}
	}
}

func TestCh08PersistencePreferenceHandoff(t *testing.T) {
	h := ch08DomainFor(t, "preferences")
	entered, release := ch08Hold(t, h, "replace", false)
	done := make(chan error, 1)
	go func() { done <- h.update(1, h.change) }()
	ch08Wait(t, entered, "positive handoff barrier missing")
	before, watch, e := h.server.Preferences().Subscribe()
	if e != nil {
		t.Fatal(e)
	}
	defer watch.Close()
	if before.Revision != 1 {
		t.Fatal("subscription observed uncommitted candidate")
	}
	close(release)
	if e = ch08Result(t, done); e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	next, e := watch.Next(ctx)
	if e != nil || next.Revision != 2 || next.Preferences.FontSize != 22 {
		t.Fatal("old snapshot/new tail handoff lost committed change", next, e)
	}
	after, tail, e := h.server.Preferences().Subscribe()
	if e != nil {
		t.Fatal(e)
	}
	defer tail.Close()
	if after.Revision != 2 || after.Preferences.FontSize != 22 {
		t.Fatal("new subscriber regressed after ack")
	}
	unchanged, e := h.server.Preferences().Update(2, h.change)
	if e != nil || unchanged.Revision != 2 {
		t.Fatal("no-change preference update advanced revision", e)
	}
	if _, e = h.server.Preferences().Update(2, json.RawMessage(`{"font_size":24}`)); e != nil {
		t.Fatal(e)
	}
	next, e = tail.Next(ctx)
	if e != nil || next.Revision != 3 || next.Preferences.FontSize != 24 {
		t.Fatal("no-change broadcast preceded next actual preference change", next, e)
	}
}

func TestCh08PersistencePreferenceOverflowAndIdleClose(t *testing.T) {
	h := ch08DomainFor(t, "preferences")
	_, slow, e := h.server.Preferences().Subscribe()
	if e != nil {
		t.Fatal(e)
	}
	defer slow.Close()
	_, fast, e := h.server.Preferences().Subscribe()
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Chapter 7 specifies 256 retained observations. Use its published bound,
	// rather than reading the queue's implementation capacity.
	for i := 0; i < 256; i++ {
		font := 14
		if i%2 == 1 {
			font = 20
		}
		if _, e = h.server.Preferences().Update(uint64(i+1), json.RawMessage(fmt.Sprintf(`{"font_size":%d}`, font))); e != nil {
			t.Fatal(e)
		}
		v, e := fast.Next(ctx)
		if e != nil || v.Revision != uint64(i+2) {
			t.Fatal("healthy preference peer lost revision", v, e)
		}
	}
	select {
	case <-slow.Done():
		t.Fatal("exact-capacity preference watch closed early")
	default:
	}
	if _, e = h.server.Preferences().Update(257, json.RawMessage(`{"font_size":24}`)); e != nil {
		t.Fatal(e)
	}
	select {
	case <-slow.Done():
	default:
		t.Fatal("overflowing preference watch silently remained current")
	}
	if _, e = slow.Next(ctx); e == nil {
		t.Fatal("overflowed preference watch delivered incomplete state")
	}
	v, e := fast.Next(ctx)
	if e != nil || v.Revision != 258 {
		t.Fatal("overflow affected healthy preference peer", v, e)
	}
	latest, joined, e := h.server.Preferences().Subscribe()
	if e != nil {
		t.Fatal(e)
	}
	if latest.Revision != 258 || latest.Preferences.FontSize != 24 {
		t.Fatal("resubscribe failed to recover authoritative preferences")
	}
	fast.Close()
	joined.Close()
	slow.Close()
	if h.server.preferences.Ch08WatchCount() != 0 {
		t.Fatal("closed preference recipient retained without further publication")
	}
}

func TestCh08PersistenceActualSocketReadLoop(t *testing.T) {
	for _, name := range []string{"policy", "preferences"} {
		t.Run(name, func(t *testing.T) {
			h := ch08DomainFor(t, name)
			httpServer := httptest.NewUnstartedServer(h.server)
			// Establish the creation identity before serving. A fixed test port
			// would fail the inherited Host guard before reaching the read loop.
			h.server.origin = "http://" + httpServer.Listener.Addr().String()
			httpServer.Start()
			defer httpServer.Close()
			socket, _, e := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(httpServer.URL, "http")+"/ws", http.Header{"Origin": []string{httpServer.URL}})
			if e != nil {
				t.Fatal(e)
			}
			defer socket.Close()
			socket.SetReadDeadline(time.Now().Add(4 * time.Second))
			send := func(v any) {
				t.Helper()
				if e := socket.WriteJSON(v); e != nil {
					t.Fatal(e)
				}
			}
			until := func(id, kind string) map[string]any {
				t.Helper()
				for {
					var v map[string]any
					if e := socket.ReadJSON(&v); e != nil {
						t.Fatal(e)
					}
					if (id == "" || v["id"] == id) && (kind == "" || v["type"] == kind) {
						return v
					}
				}
			}
			send(map[string]any{"type": "subscribe", "id": "sub"})
			until("", "snapshot_end")
			entered, release := ch08Hold(t, h, "replace", false)
			send(map[string]any{"type": name + "_update", "id": "pending", "base_revision": 1, "patch": h.change})
			ch08Wait(t, entered, "socket update did not reach disk")
			send(map[string]any{"type": "pause", "id": "pause", "typing": true, "speaking": false})
			ack := until("pause", "ack")
			if ack["typing_clients"] != float64(1) {
				t.Fatal("pause read loop did not progress while disk held")
			}
			send(map[string]any{"type": name + "_update", "id": "busy", "base_revision": 1, "patch": h.retry})
			refusal := until("busy", "error")
			if refusal["code"] != "settings_busy" {
				t.Fatal("busy socket update not promptly refused", refusal)
			}
			other := "policy"
			patch := json.RawMessage(`{"max_model_requests":4}`)
			if name == "policy" {
				other = "preferences"
				patch = json.RawMessage(`{"theme":"light"}`)
			}
			send(map[string]any{"type": other + "_update", "id": "other", "base_revision": 0, "patch": patch})
			until("other", other+"_ack")
			// A disconnected sender loses its reply, not a candidate already owned by
			// the service. Do not resend it to make this case pass.
			_ = socket.Close()
			close(release)
			deadline := time.Now().Add(time.Second)
			for ch08Value(t, h.snapshot())["revision"] != float64(2) {
				if time.Now().After(deadline) {
					t.Fatal("disconnect discarded accepted settings write")
				}
				time.Sleep(time.Millisecond)
			}
		})
	}
}
