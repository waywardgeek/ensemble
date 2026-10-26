package gui_test

import (
	"bytes"
	"encoding/json"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/websocket"
)

func TestDisplayNamesWireAndRestart(t *testing.T) {
	for _, mode := range []string{"legacy", "application"} {
		t.Run(mode, func(t *testing.T) {
			directory := t.TempDir()
			configPath := ""
			if mode == "application" {
				configPath = applicationFixture(t)
			}
			start := func() (*ensemble.Ensemble, ensemble.Agent, *httptest.Server, func()) {
				t.Helper()
				cfg := ensemble.Config{Model: "offline", Recall: ensemble.RecallSettings{Disabled: true}}
				var store *gui.ApplicationConfigStore
				if configPath != "" {
					var err error
					store, err = gui.LoadApplicationConfig(configPath)
					if err != nil {
						t.Fatal(err)
					}
				}
				root := ensemble.New(io.Discard)
				main, err := root.Start(directory, cfg, ensemble.AgentSpec{}, make(selectionInbox, 1))
				if err != nil {
					t.Fatal(err)
				}
				var hub *gui.Server
				if store == nil {
					hub, err = gui.New(root, main, io.Discard)
				} else {
					hub, err = gui.NewConfigured(root, main, io.Discard, store)
				}
				if err != nil {
					root.Close()
					t.Fatal(err)
				}
				server := httptest.NewServer(hub.Handler(""))
				closed := false
				stop := func() {
					if closed {
						return
					}
					closed = true
					hub.Close()
					server.Close()
					if err := root.Close(); err != nil {
						t.Error(err)
					}
				}
				t.Cleanup(stop)
				return root, main, server, stop
			}
			root, main, server, stop := start()
			peer, err := root.CreateAgent("peer", "offline", "")
			if err != nil {
				t.Fatal(err)
			}
			for _, agent := range []ensemble.Agent{main, peer} {
				if err := agent.Record(ensemble.Event{Type: "message_received", Message: &ensemble.MessageData{Actor: ensemble.Human, Parts: []ensemble.Part{{Type: "text", Text: ensemble.AgentKey(agent) + " history"}}}}); err != nil {
					t.Fatal(err)
				}
			}
			mainDir, peerDir := main.Config().DataDir, peer.Config().DataDir
			registry, err := os.ReadFile(filepath.Join(directory, "ensemble.json"))
			if err != nil {
				t.Fatal(err)
			}
			first, second := dial(t, server.URL), dial(t, server.URL)
			until(t, first, kind("current_settings"))
			until(t, second, kind("current_settings"))
			write := func(conn *websocket.Conn, frame any) {
				t.Helper()
				if err := conn.WriteJSON(frame); err != nil {
					t.Fatal(err)
				}
			}
			write(second, map[string]string{"type": "select_agent", "key": "peer"})
			until(t, second, kind("current_settings"))
			patch := func(key string, name any) {
				write(first, map[string]any{"type": "update_settings", "key": key, "settings": map[string]any{"display_name": name}})
			}
			checkTree := func(conn *websocket.Conn, names map[string]string) {
				t.Helper()
				frames := until(t, conn, func(frame map[string]any) bool {
					if frame["type"] != "agent_tree" {
						return false
					}
					matched := 0
					for _, value := range frame["root"].(map[string]any)["children"].([]any) {
						node := value.(map[string]any)
						key := node["key"].(string)
						if node["id"] != key {
							t.Fatal("display rename mutated identity", node)
						}
						if name, ok := names[key]; ok && node["display_name"] == name {
							matched++
						}
					}
					return matched == len(names)
				})
				if len(frames) == 0 {
					t.Fatal("missing tree")
				}
			}
			patch("main", "  Astra <main>  ")
			until(t, first, kind("settings_changed"))
			for _, conn := range []*websocket.Conn{first, second} {
				checkTree(conn, map[string]string{"main": "Astra <main>"})
			}
			patch("peer", "Research / peer")
			until(t, second, kind("settings_changed"))
			names := map[string]string{"main": "Astra <main>", "peer": "Research / peer"}
			for _, conn := range []*websocket.Conn{first, second} {
				checkTree(conn, names)
			}
			if configPath != "" {
				saved, err := gui.LoadApplicationConfig(configPath)
				if err != nil {
					t.Fatal(err)
				}
				if saved.Snapshot().GUI.DisplayName != names["main"] {
					t.Fatal("application config persistence dropped main display_name")
				}
			}
			for _, name := range []any{"", " \t\n ", nil} {
				patch("main", name)
				until(t, first, kind("settings_error"))
			}
			patch("missing", "Other")
			until(t, first, kind("settings_error"))
			if main.Config().DataDir != mainDir || peer.Config().DataDir != peerDir || ensemble.AgentKey(main) != "main" || ensemble.AgentKey(peer) != "peer" {
				t.Fatal("rename changed Agent identity or directory")
			}
			stop()
			root, main, server, _ = start()
			peer, err = root.FindAgent("peer", main)
			if err != nil {
				t.Fatal(err)
			}
			for _, agent := range []ensemble.Agent{main, peer} {
				data, err := json.Marshal(agent.History().EventsAfter(0))
				if err != nil || !bytes.Contains(data, []byte(ensemble.AgentKey(agent)+" history")) {
					t.Fatalf("history lost on restart: %s, %v", data, err)
				}
			}
			savedRegistry, err := os.ReadFile(filepath.Join(directory, "ensemble.json"))
			if err != nil || !bytes.Equal(savedRegistry, registry) {
				t.Fatal("GUI rename changed peer registry", err)
			}
			fresh := dial(t, server.URL)
			checkTree(fresh, names)
			frames := until(t, fresh, kind("current_settings"))
			if frames[len(frames)-1]["settings"].(map[string]any)["display_name"] != names["main"] {
				t.Fatal("main name not restored")
			}
			write(fresh, map[string]string{"type": "select_agent", "key": "peer"})
			frames = until(t, fresh, kind("current_settings"))
			if frames[len(frames)-1]["settings"].(map[string]any)["display_name"] != names["peer"] {
				t.Fatal("peer name not restored")
			}
		})
	}
}
