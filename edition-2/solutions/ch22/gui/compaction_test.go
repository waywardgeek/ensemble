package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"ensemble/ensemble"
	"github.com/gorilla/websocket"
)

type compactionObserver func(ensemble.Agent, ensemble.Observation)

// ObserveAgent retains an actual observation for delayed-delivery coverage.
func (f compactionObserver) ObserveAgent(a ensemble.Agent, o ensemble.Observation) { f(a, o) }

func compactionFrame(t *testing.T, c *websocket.Conn, kind string) map[string]any {
	t.Helper()
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		var frame map[string]any
		if err := c.ReadJSON(&frame); err != nil {
			t.Fatal(err)
		}
		if frame["type"] == "agent_error" {
			t.Fatalf("GUI error: %v", frame)
		}
		if frame["type"] == kind {
			return frame
		}
	}
}

func compactionTreeHas(node map[string]any, key string) bool {
	if node["key"] == key {
		return true
	}
	children, _ := node["children"].([]any)
	for _, child := range children {
		if compactionTreeHas(child.(map[string]any), key) {
			return true
		}
	}
	return false
}

// TestCompactionViewRemoval checks real socket selection, owner settings/replay,
// repeated reference cleanup and an actual late observation after worker removal.
func TestCompactionViewRemoval(t *testing.T) {
	for _, nested := range []bool{false, true} {
		t.Run(fmt.Sprint(nested), func(t *testing.T) {
			var boundary sync.Mutex
			var entered, release chan struct{}
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					Tools []struct {
						Name   string `json:"name"`
						Schema struct {
							Properties map[string]json.RawMessage `json:"properties"`
						} `json:"input_schema"`
					} `json:"tools"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					http.Error(w, err.Error(), 400)
					return
				}
				payload := map[string]string{"result": "ordinary parent complete"}
				if len(request.Tools) == 1 && request.Tools[0].Name == "submit" && request.Tools[0].Schema.Properties["memory"] != nil {
					boundary.Lock()
					start, finish := entered, release
					boundary.Unlock()
					close(start)
					select {
					case <-finish:
					case <-r.Context().Done():
						return
					}
					payload = map[string]string{"memory": "I kept why the decision was made."}
				}
				json.NewEncoder(w).Encode(map[string]any{"content": []any{map[string]any{"type": "tool_use", "id": "submit", "name": "submit", "input": payload}}, "usage": map[string]int{"input_tokens": 10, "output_tokens": 5}})
			}))
			t.Cleanup(peer.Close)
			t.Cleanup(func() {
				boundary.Lock()
				defer boundary.Unlock()
				if release != nil {
					select {
					case <-release:
					default:
						close(release)
					}
				}
			})
			app := ensemble.New(io.Discard)
			app.EnableFleet(nil)
			root, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: peer.URL, DataDir: t.TempDir(), DisableStreaming: true})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { root.Shutdown() })
			owner := root
			wantFont := float64(16)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if nested {
				if _, err := root.Fleet().Spawn("owner", "ordinary task", ""); err != nil {
					t.Fatal(err)
				}
				owner, err = root.Fleet().Find("owner")
				if err != nil {
					t.Fatal(err)
				}
				if err := owner.Flush(ctx); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(owner.Config().DataDir, "settings.json"), []byte(`{"font_size":22}`), 0600); err != nil {
					t.Fatal(err)
				}
				wantFont = 22
			}
			hub, err := New(app, root, io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(hub.Close)
			server := httptest.NewServer(hub.Handler(""))
			t.Cleanup(server.Close)
			conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { conn.Close() })
			var observed sync.Mutex
			var lastAgent ensemble.Agent
			var last ensemble.Observation
			app.AddObserver(compactionObserver(func(a ensemble.Agent, o ensemble.Observation) {
				if _, removed := o.(ensemble.AgentRemoved); removed {
					return
				}
				if a.Spec().Compaction {
					observed.Lock()
					lastAgent, last = a, o
					observed.Unlock()
				}
			}))
			for i := 0; i < 3; i++ {
				boundary.Lock()
				entered, release = make(chan struct{}), make(chan struct{})
				start, finish := entered, release
				boundary.Unlock()
				child, done, err := owner.Fleet().StartCompaction(fmt.Sprintf("worker-%d", i), "source data")
				if err != nil {
					t.Fatal(err)
				}
				select {
				case <-start:
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
				if err := conn.WriteJSON(map[string]string{"type": "select_agent", "key": ensemble.AgentKey(child)}); err != nil {
					t.Fatal(err)
				}
				if got := compactionFrame(t, conn, "agent_selected")["key"]; got != ensemble.AgentKey(child) {
					t.Fatalf("selected %v", got)
				}
				compactionFrame(t, conn, "current_settings")
				close(finish)
				select {
				case out := <-done:
					if out.Result.Err != nil || out.CleanupErr != nil {
						t.Fatalf("worker: %+v", out)
					}
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
				if got := compactionFrame(t, conn, "agent_selected")["key"]; got != ensemble.AgentKey(owner) {
					t.Fatalf("fallback selected %v", got)
				}
				tree := compactionFrame(t, conn, "agent_tree")["root"].(map[string]any)
				if compactionTreeHas(tree, ensemble.AgentKey(child)) || !compactionTreeHas(tree, ensemble.AgentKey(owner)) {
					t.Fatalf("fallback tree lost owner or retained worker: %v", tree)
				}
				if nested {
					if got := compactionFrame(t, conn, "tool_finished")["result"]; got != "submitted: ordinary parent complete" {
						t.Fatalf("fallback lost owner replay: %v", got)
					}
				}
				settings := compactionFrame(t, conn, "current_settings")["settings"].(map[string]any)
				if settings["font_size"] != wantFont {
					t.Fatalf("fallback lost owner's saved settings: %v", settings)
				}
				observed.Lock()
				lateAgent, late := lastAgent, last
				observed.Unlock()
				if lateAgent != child || late == nil {
					t.Fatal("no actual late observation captured")
				}
				app.Observe(child, late)
				app.ForgetCompaction(child)
				hub.mu.Lock()
				_, retained := hub.views[child]
				count := len(hub.views)
				selected := false
				for _, v := range hub.selected {
					selected = selected || v.agent == child
				}
				hub.mu.Unlock()
				if retained || selected || count > 2 {
					t.Fatalf("removed worker retained: view=%v selection=%v views=%d", retained, selected, count)
				}
			}
		})
	}
}

// TestCompactionRemovedDuringSelection holds a real settings read while the
// selected worker finishes, so its delayed selection must fall back to the owner.
func TestCompactionRemovedDuringSelection(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		close(entered)
		<-release
		json.NewEncoder(w).Encode(map[string]any{"content": []any{map[string]any{"type": "tool_use", "id": "submit", "name": "submit", "input": map[string]string{"memory": "I kept the reason."}}}, "usage": map[string]int{"input_tokens": 10, "output_tokens": 5}})
	}))
	t.Cleanup(peer.Close)
	app := ensemble.New(io.Discard)
	root, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: peer.URL, DataDir: t.TempDir(), DisableStreaming: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Shutdown() })
	hub, err := New(app, root, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(hub.Close)
	server := httptest.NewServer(hub.Handler(""))
	t.Cleanup(server.Close)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	child, done, err := root.Fleet().StartCompaction("selecting", "source data")
	if err != nil {
		t.Fatal(err)
	}
	<-entered
	path := filepath.Join(child.Config().DataDir, "settings.json")
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	reading := make(chan error, 1)
	finishRead := make(chan struct{})
	readDone := make(chan error, 1)
	go func() {
		f, err := os.OpenFile(path, os.O_WRONLY, 0600)
		reading <- err
		if err != nil {
			return
		}
		<-finishRead
		_, err = f.Write([]byte(`{}`))
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
		readDone <- err
	}()
	if err := conn.WriteJSON(map[string]string{"type": "select_agent", "key": ensemble.AgentKey(child)}); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-reading:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("selection did not start settings read")
	}
	close(release)
	select {
	case out := <-done:
		if out.Result.Err != nil || out.CleanupErr != nil {
			t.Fatalf("worker: %+v", out)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("settings selection blocked worker completion")
	}
	close(finishRead)
	if err := <-readDone; err != nil {
		t.Fatal(err)
	}
	if got := compactionFrame(t, conn, "agent_selected")["key"]; got != ensemble.AgentKey(root) {
		t.Fatalf("late selection resurrected %v", got)
	}
	compactionFrame(t, conn, "current_settings")
	hub.mu.Lock()
	_, retained := hub.views[child]
	hub.mu.Unlock()
	if retained {
		t.Fatal("settings load recreated removed view")
	}
}

// TestRejectedOrdinaryChildSelection covers a GUI selection that already resolved
// a factory child when its managed-purpose check refuses and removes that child.
func TestRejectedOrdinaryChildSelection(t *testing.T) {
	app := ensemble.New(io.Discard)
	root, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Shutdown() })
	hub, err := New(app, root, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(hub.Close)
	server := httptest.NewServer(hub.Handler(""))
	t.Cleanup(server.Close)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := conn.WriteJSON(map[string]string{"type": "subscribe"}); err != nil {
		t.Fatal(err)
	}
	compactionFrame(t, conn, "agent_selected")
	compactionFrame(t, conn, "current_settings")
	var rejected ensemble.Agent
	app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
		spec.Compaction = false
		child, err := app.NewChildAgent(parent, parent.Config(), spec)
		rejected = child
		return child, err
	})
	if _, _, err := root.Fleet().StartCompaction("rejected", "source"); err == nil {
		t.Fatal("wrong purpose accepted")
	}
	if rejected == nil || app.ContainsAgent(rejected) {
		t.Fatal("actual refused child was not removed")
	}
	// The connection is real; calling the shared selection continuation models
	// having resolved the old pointer before removal, without a runtime test hook.
	hub.mu.Lock()
	var connection *websocket.Conn
	var queue chan [][]byte
	for current, q := range hub.clients {
		connection, queue = current, q
	}
	hub.mu.Unlock()
	if connection == nil {
		t.Fatal("socket not registered")
	}
	if err := hub.selectConversation(connection, queue, rejected, true, nil); err != nil {
		t.Fatal(err)
	}
	if got := compactionFrame(t, conn, "agent_selected")["key"]; got != ensemble.AgentKey(root) {
		t.Fatalf("selected rejected child: %v", got)
	}
	compactionFrame(t, conn, "current_settings")
	hub.mu.Lock()
	_, retained := hub.views[rejected]
	selected := hub.selected[connection].agent
	hub.mu.Unlock()
	if retained || selected != root {
		t.Fatal("refused ordinary child retained by GUI")
	}
}
