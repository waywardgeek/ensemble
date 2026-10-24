package gui

import (
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"ensemble"
)

func TestCh20Round2SelectedScope(t *testing.T) {
	root := ensemble.New(nil)
	defer root.Shutdown()
	base := t.TempDir()
	create := func(id string) ensemble.Agent {
		a, err := root.NewAgent(ensemble.Config{ID: id, Model: "offline", BaseDir: base, DataDir: filepath.Join(base, id), SavePath: filepath.Join(base, id, "save.json")})
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	first, selected := create("first"), create("selected")
	view, err := New(root, selected, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer view.Close()
	checkSnapshot := func(c *client, want int) {
		t.Helper()
		var frame struct {
			Type string
			Data struct {
				Agents           []struct{ ID string }
				Events, Partials []Wire
			}
		}
		if err := json.Unmarshal(<-c.queue, &frame); err != nil {
			t.Fatal(err)
		}
		if frame.Type != "replay" || len(frame.Data.Agents) != want {
			t.Fatalf("snapshot %+v", frame)
		}
		if want == 1 {
			if frame.Data.Agents[0].ID != selected.ID() {
				t.Fatal("wrong selected Agent")
			}
			for _, w := range append(frame.Data.Events, frame.Data.Partials...) {
				if w.AgentID != selected.ID() {
					t.Fatalf("foreign replay %+v", w)
				}
			}
		}
	}
	c := view.Hub.connect()
	checkSnapshot(c, 1)
	drain(c)
	late := create("late")
	for _, a := range []ensemble.Agent{first, late} {
		view.Hub.observe(ensemble.AgentObservation{Agent: a})
	}
	view.Hub.mu.Lock()
	view.Hub.broadcast(Wire{Type: "state", AgentID: first.ID()})
	_, hasFirst := view.Hub.agents[first.ID()]
	_, hasLate := view.Hub.agents[late.ID()]
	view.Hub.mu.Unlock()
	if hasFirst || hasLate {
		t.Fatal("selected Hub projected another Agent")
	}
	for _, w := range drain(c) {
		if w.AgentID != "" && w.AgentID != selected.ID() {
			t.Fatalf("foreign live frame %+v", w)
		}
	}
	for _, kind := range []string{"prompt", "pause", "interrupt", "clear", "update_settings", "answer_human", "mcp_open"} {
		err := view.command(c, command{Type: kind, AgentID: first.ID(), Text: "must not route", Settings: json.RawMessage(`{"safe_mode":false}`)})
		if err == nil || !strings.Contains(err.Error(), "outside") {
			t.Fatalf("foreign command %s: %v", kind, err)
		}
	}
	if err := view.mcp(c, command{Type: "mcp_open", AgentID: first.ID()}); err == nil {
		t.Fatal("foreign MCP attachment accepted")
	}
	filtered := view.Hub.connectAt(map[string]ensemble.ViewSnapshot{first.ID(): first.View(), selected.ID(): selected.View(), late.ID(): late.View()})
	checkSnapshot(filtered, 1)
	global := NewHub(root)
	defer global.Close()
	checkSnapshot(global.connect(), 3)
}
