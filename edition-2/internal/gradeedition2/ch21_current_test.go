package gradeedition2

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestChapter21CurrentProjection(t *testing.T) {
	root := map[string]any{"id": "main", "state": map[string]any{"turn": "idle"}}
	child := map[string]any{"id": "child", "parent_id": "main", "state": map[string]any{"turn": "in_flight"}}
	replay := map[string]any{"type": "replay", "data": map[string]any{"selected_agent_id": "main", "agents": []any{root, child}}}
	tree := map[string]any{"type": "agent_tree", "data": []any{root, child}}
	if !managementTree(tree, true) || managementTreeNode(tree, "child", true)["parent_id"] != "main" {
		t.Fatal("lost parent-linked tree")
	}
	if managementTreeNode(tree, "foreign", true) != nil {
		t.Fatal("invented foreign child")
	}
	frames := []map[string]any{replay, {"type": "state_changed", "agent_id": "child", "to": "in_flight"}}
	if managementStateWith(frames, true) != "idle" {
		t.Fatal("foreign state replaced selected replay")
	}
	frames = append(frames, map[string]any{"type": "state_changed", "agent_id": "main", "to": "in_flight"})
	if managementStateWith(frames, true) != "in_flight" {
		t.Fatal("selected live state missing")
	}
	for _, bad := range []map[string]any{
		{"type": "tool_called", "arguments": tree},
		{"type": "agent_tree", "data": []any{root, root}},
		{"type": "agent_tree", "data": []any{map[string]any{"id": 7}}},
		{"type": "agent_tree", "data": []any{map[string]any{"id": "x", "parent_id": 7}}},
	} {
		if managementTree(bad, true) {
			t.Fatalf("malformed or forged roster accepted: %#v", bad)
		}
	}
	if !managementContains([]map[string]any{{"type": "part_final", "kind": "text", "text": "visible"}}, "visible") {
		t.Fatal("flat text missing")
	}
	if managementContains([]map[string]any{{"type": "part_final", "kind": "tool_call", "text": "not-visible"}}, "not-visible") {
		t.Fatal("non-text credited")
	}
	legacy := map[string]any{"type": "agent_tree", "root": map[string]any{"id": "main", "children": []any{map[string]any{"id": "child"}}}}
	if managementTreeNode(legacy, "child", false) == nil || managementTreeNode(legacy, "child", true) != nil {
		t.Fatal("legacy/current tree boundaries changed")
	}
}

func TestChapter21CurrentStatus(t *testing.T) {
	valid := `{"agent_id":"worker23","parent_id":"main","state":"submitted","outcome":{"agent_id":"worker23","parent_id":"main","state":"submitted","submission":` + fleetReviewPayload + `}}`
	row, err := managementStatus(valid, true)
	if err != nil || !reflect.DeepEqual(row["payload"], fleetJSON(fleetReviewPayload)) {
		t.Fatalf("typed result: %#v %v", row, err)
	}
	for _, bad := range []string{
		`null`, `[]`, `{"outcome":"encoded"}`,
		`{"agent_id":"worker23","state":"running","outcome":{"agent_id":"worker23","state":"submitted"}}`,
		`{"agent_id":"worker23","state":"submitted","outcome":{"agent_id":"foreign","state":"submitted"}}`,
	} {
		if _, err := managementStatus(bad, true); err == nil {
			t.Fatalf("bad status accepted: %s", bad)
		}
	}
	row, err = managementStatus(`{"agent_id":"worker23","state":"running"}`, true)
	if err != nil || row["payload"] != nil {
		t.Fatal("running task fabricated submission")
	}
}

func TestChapter21CurrentWaitArguments(t *testing.T) {
	original := fleetCall("wait-change", "wait_for_agent_change", map[string]any{"agent_ids": []string{"worker23"}, "timeout_seconds": 5, "ai_callback_delay": 10})
	before := original.ToolArgs
	adapted := chapter21ToolReply(nil, original)
	var args map[string]any
	if err := json.Unmarshal([]byte(adapted.ToolArgs), &args); err != nil {
		t.Fatal(err)
	}
	if adapted.ToolID != original.ToolID || adapted.ToolName != original.ToolName || args["timeout_seconds"] != float64(5) || !reflect.DeepEqual(args["agent_ids"], []any{"worker23"}) {
		t.Fatalf("changed wait: %#v", adapted)
	}
	if _, ok := args["ai_callback_delay"]; ok {
		t.Fatal("obsolete callback argument retained")
	}
	if original.ToolArgs != before {
		t.Fatal("mutated scripted source reply")
	}
}

func TestChapter21CurrentFileLaunch(t *testing.T) {
	work := t.TempDir()
	bin := filepath.Join(work, "stub")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > launch-args\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if p, peer, err := chapter21GUIStart(bin, work, "http://127.0.0.1:1"); err == nil {
		peer.close()
		p.close()
		t.Fatal("exited stub counted as ready GUI")
	}
	args, err := os.ReadFile(filepath.Join(work, "launch-args"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(work, "application.json")
	fields := strings.Fields(string(args))
	if len(fields) != 4 || fields[0] != "--config" || fields[1] != path || fields[2] != "--port" {
		t.Fatalf("not published file-mode launch: %q", args)
	}
	port, err := strconv.Atoi(fields[3])
	if err != nil || port < 1 || port > 65535 {
		t.Fatalf("invalid explicit port: %q", fields[3])
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err = json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	if config["base_url"] != "http://127.0.0.1:1" || config["api_key_file"] != filepath.Join(work, "fixture-key") || config["base_dir"] != work || config["version"] != float64(1) {
		t.Fatalf("wrong isolated config: %s", raw)
	}
	for _, field := range []string{"api_key", "api_key_env", "port", "streaming", "safe_mode", "preferences"} {
		if _, exists := config[field]; exists {
			t.Fatalf("unnecessary or unpublished launch field %q", field)
		}
	}
	keyPath := filepath.Join(work, "fixture-key")
	key, err := os.ReadFile(keyPath)
	if err != nil || string(key) != "local-fixture-only" {
		t.Fatalf("missing isolated dummy key: %v", err)
	}
	info, err := os.Stat(keyPath)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("dummy key must be private: %v", err)
	}
}

func TestChapter21CurrentSelectionCorrelation(t *testing.T) {
	replay := func(id, text string) map[string]any {
		return map[string]any{"type": "replay", "data": map[string]any{
			"selected_agent_id": id,
			"agents":            []any{map[string]any{"id": id, "state": map[string]any{"turn": "idle"}}},
			"events":            []any{map[string]any{"type": "part_final", "agent_id": id, "kind": "text", "text": text}},
		}}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		for _, frame := range []map[string]any{replay("main", "FOREIGN_HISTORY"), {"type": "current_settings", "agent_id": "main"}} {
			if err := conn.WriteJSON(frame); err != nil {
				t.Error(err)
				return
			}
		}
		var command map[string]any
		if err := conn.ReadJSON(&command); err != nil {
			t.Error(err)
			return
		}
		if command["type"] != "select_agent" || command["agent_id"] != "named23" {
			t.Errorf("wrong selection: %#v", command)
			return
		}
		for _, frame := range []map[string]any{
			{"type": "current_settings", "agent_id": "named23"}, // A trailer alone is not replay.
			{"type": "state_changed", "agent_id": "main", "to": "in_flight"},
			{"type": "part_final", "agent_id": "main", "kind": "text", "text": "FOREIGN_LIVE"},
			{"type": "current_settings", "agent_id": "main"},
			replay("named23", "REQUESTED_HISTORY"),
		} {
			if err := conn.WriteJSON(frame); err != nil {
				t.Error(err)
				return
			}
		}
		// Keep the connection alive until the client has consumed its frames.
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()
	peer, err := openGUIPeer("ws" + strings.TrimPrefix(server.URL, "http"))
	if err != nil {
		t.Fatal(err)
	}
	defer peer.close()
	frames, err := managementSelectWith(peer, "named23", true)
	if err != nil {
		t.Fatal(err)
	}
	if managementStateWith(frames, true) != "idle" || !managementContains(frames, "REQUESTED_HISTORY") || managementContains(frames, "FOREIGN_HISTORY") || managementContains(frames, "FOREIGN_LIVE") {
		t.Fatalf("selection crossed identities or ended early: %#v", frames)
	}
	if last := frames[len(frames)-1]; last["type"] != "replay" || managementSelected(last) != "named23" {
		t.Fatalf("wrong selection boundary: %#v", last)
	}
}

func TestChapter21CurrentAcceptedOutcomeIdentity(t *testing.T) {
	for _, fields := range []string{
		``, `"agent_id":"worker23",`, `"state":"submitted",`,
		`"agent_id":"foreign","state":"submitted",`,
		`"agent_id":"worker23","state":"running",`,
		`"agent_id":"worker23","state":"failed",`,
	} {
		text := `{"outcomes":[{` + fields + `"submission":` + fleetReviewPayload + `}]}`
		if managementJoinedPayload(text, true) {
			t.Fatalf("unaccepted join credited: %s", text)
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(`{`+fields+`"payload":`+fleetReviewPayload+`}`), &row); err != nil {
			t.Fatal(err)
		}
		if managementSubmitted(row, "payload", true) {
			t.Fatalf("unaccepted status credited: %#v", row)
		}
	}
	row := map[string]any{"agent_id": "worker23", "state": "submitted", "payload": fleetJSON(fleetReviewPayload)}
	if !managementSubmitted(row, "payload", true) {
		t.Fatal("accepted exact outcome refused")
	}
}

func TestChapter21CurrentNamedCommands(t *testing.T) {
	for _, kind := range []string{"create_agent", "delete_agent"} {
		for _, current := range []bool{false, true} {
			field := "name"
			if current {
				field = "agent_id"
			}
			want := map[string]string{"type": kind, field: "named23"}
			if got := managementNamedCommand(kind, "named23", current); !reflect.DeepEqual(got, want) {
				t.Fatalf("%s current=%v: got %#v, want %#v", kind, current, got, want)
			}
		}
	}
}

// The GUI is observed over WebSocket, but its terminal output still needs a
// consumer: mcpCommand's stdout scanner owns process reaping after stdout EOF.
func TestChapter21CurrentGUIOutputDrain(t *testing.T) {
	for _, exit := range []string{"0", "7"} {
		t.Run(exit, func(t *testing.T) {
			work := t.TempDir()
			bin := filepath.Join(work, "verbose-gui-stub")
			script := "#!/bin/sh\ni=0\nwhile [ \"$i\" -lt 256 ]; do printf 'diagnostic %s\\n' \"$i\"; i=$((i+1)); done\nprintf 'EXIT_CONTROL\\n' >&2\nexit " + exit + "\n"
			if err := os.WriteFile(bin, []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			p, peer, err := chapter21GUIStart(bin, work, "http://127.0.0.1:1")
			if err == nil {
				peer.close()
				p.close()
				t.Fatal("exited stub counted as ready GUI")
			}
			if !strings.Contains(err.Error(), "GUI child exited before readiness:") || !strings.Contains(err.Error(), "EXIT_CONTROL") {
				t.Fatalf("stdout backpressure hid actual child exit: %v", err)
			}
			if exit == "7" && !strings.Contains(err.Error(), "exit status 7") {
				t.Fatalf("lost nonzero exit: %v", err)
			}
		})
	}
}

func TestChapter21CurrentSavedSubmissionWithoutRespawn(t *testing.T) {
	messages := []any{map[string]any{"role": "user", "content": "RESTORE_ROOT_23"}}
	body, _ := json.Marshal(map[string]any{"messages": messages})
	current := managementReply(body, t.TempDir(), true)
	if current.ToolName != "get_submission" || current.ToolID != "durable-retrieve" || current.ToolArgs != `{"agent_id":"worker23"}` {
		t.Fatalf("current retrieval reconstructed a child or changed target: %#v", current)
	}
	legacy := managementReply(body, t.TempDir(), false)
	if legacy.ToolName != "spawn_sub_agent" || legacy.ToolID != "respawn-worker" {
		t.Fatalf("legacy initial step changed: %#v", legacy)
	}
	messages = append(messages, map[string]any{"role": "user", "content": []any{map[string]any{
		"type": "tool_result", "tool_use_id": "durable-retrieve", "content": fleetReviewPayload,
	}}})
	body, _ = json.Marshal(map[string]any{"messages": messages})
	current = managementReply(body, t.TempDir(), true)
	if current.Text != "RESTORED_RESULT_23" || current.ToolName != "" || len(current.Tools) != 0 {
		t.Fatalf("saved-answer read requested extra lifecycle work: %#v", current)
	}
	// Historical drivers keep their original ordered respawn/read/status/kill
	// protocol; the current simplification must not silently change that route.
	messages = messages[:1]
	for _, step := range []struct{ id, name string }{
		{"respawn-worker", "spawn_sub_agent"},
		{"durable-retrieve", "get_submission"},
		{"durable-status", "get_sub_agent_status"},
		{"stop-reopened", "kill_agent"},
	} {
		body, _ = json.Marshal(map[string]any{"messages": messages})
		reply := managementReply(body, t.TempDir(), false)
		if reply.ToolID != step.id || reply.ToolName != step.name {
			t.Fatalf("legacy step %s changed: %#v", step.id, reply)
		}
		messages = append(messages, map[string]any{"role": "user", "content": []any{map[string]any{
			"type": "tool_result", "tool_use_id": step.id, "content": "ok",
		}}})
	}
	body, _ = json.Marshal(map[string]any{"messages": messages})
	if reply := managementReply(body, t.TempDir(), false); reply.Text != "RESTORED_RESULT_23" || reply.ToolName != "" {
		t.Fatalf("legacy completion changed: %#v", reply)
	}
}

// These are public wire shapes, not the fixture's former private schema.
func TestChapter21CurrentPublicProjection(t *testing.T) {
	root := map[string]any{"id": "main", "agent_id": "main", "state": "idle"}
	child := map[string]any{"id": "child", "agent_id": "child", "parent_id": "main", "state": "running"}
	rows := []any{root, child}
	for _, data := range []any{rows, map[string]any{"agents": rows}, map[string]any{"agent_tree": rows, "agents": rows}} {
		frame := map[string]any{"type": "agent_tree", "data": data}
		if node := managementTreeNode(frame, "child", true); node == nil || node["parent_id"] != "main" {
			t.Errorf("lost public roster: %#v", frame)
		}
	}
	replay := map[string]any{"type": "replay", "data": map[string]any{
		"selected_agent": "child", "agent_tree": rows,
		"agents": []any{map[string]any{"id": "child", "state": map[string]any{"turn": "in_flight"}}},
		"events": []any{
			map[string]any{"type": "part_final", "agent_id": "child", "kind": "text", "text": "CHILD"},
			map[string]any{"type": "part_final", "agent_id": "main", "kind": "text", "text": "FOREIGN"},
		},
	}}
	if managementTreeNode(replay, "main", true) == nil {
		t.Error("selected conversation rows hid full roster")
	}
	frames := managementSelectionFrames([]map[string]any{replay}, "child")
	if managementStateWith(frames, true) != "in_flight" || !managementContains(frames, "CHILD") || managementContains(frames, "FOREIGN") {
		t.Errorf("lost selection/turn isolation: %#v", frames)
	}
	// A lifecycle string is not proof that a model request is in flight.
	object(replay["data"])["agents"] = rows
	if state := managementStateWith([]map[string]any{replay}, true); state != "" {
		t.Errorf("lifecycle invented turn state: %q", state)
	}
	object(replay["data"])["selected_agent_id"] = "main"
	if got := managementSelectionFrames([]map[string]any{replay}, "child"); len(got) != 0 {
		t.Errorf("conflicting selected identities accepted: %#v", got)
	}
	for _, bad := range []any{
		[]any{root, root}, []any{map[string]any{"id": "child", "agent_id": "other"}},
		[]any{map[string]any{"id": "child", "parent_id": 42}}, "not rows",
	} {
		frame := map[string]any{"type": "agent_tree", "data": map[string]any{"agent_tree": bad, "agents": rows}}
		if managementTree(frame, true) {
			t.Errorf("malformed authoritative roster fell back to other data: %#v", frame)
		}
	}
}

func TestChapter21CurrentPublicSelection(t *testing.T) {
	// Selection replay is self-identifying. A settings trailer is not required.
	for _, field := range []string{"selected_agent", "selected_agent_id", ""} {
		t.Run(field, func(t *testing.T) {
			replay := func(id, text string) map[string]any {
				data := map[string]any{
					"agents": []any{map[string]any{"id": id, "state": map[string]any{"turn": "idle"}}},
					"events": []any{map[string]any{"type": "part_final", "agent_id": id, "kind": "text", "text": text}},
				}
				if field != "" {
					data[field] = id
				}
				return map[string]any{"type": "replay", "data": data}
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
				if err != nil {
					t.Error(err)
					return
				}
				defer conn.Close()
				var command map[string]any
				if err = conn.ReadJSON(&command); err != nil {
					t.Error(err)
					return
				}
				if command["agent_id"] != "named23" {
					t.Errorf("wrong target: %#v", command)
					return
				}
				for _, frame := range []map[string]any{
					replay("main", "FOREIGN"), {"type": "current_settings"},
					{"type": "current_settings", "agent_id": "named23"},
					{"type": "replay", "data": map[string]any{field: "named23", "agents": []any{map[string]any{"id": "main"}}}},
					replay("named23", "REQUESTED"),
				} {
					if err = conn.WriteJSON(frame); err != nil {
						t.Error(err)
						return
					}
				}
				_, _, _ = conn.ReadMessage()
			}))
			defer server.Close()
			peer, err := openGUIPeer("ws" + strings.TrimPrefix(server.URL, "http"))
			if err != nil {
				t.Fatal(err)
			}
			defer peer.close()
			frames, err := managementSelectWith(peer, "named23", true)
			if err != nil || !managementContains(frames, "REQUESTED") || managementContains(frames, "FOREIGN") || managementStateWith(frames, true) != "idle" {
				t.Fatalf("selection: %#v (%v)", frames, err)
			}
		})
	}
}

func TestChapter21CurrentActivitySpellings(t *testing.T) {
	for _, tc := range []struct{ wire, want string }{
		{"waiting_for_model", "thinking"}, {"running_tools", "tool_call"}, {"idle", "idle"},
		{"thinking", "thinking"}, {"tool_call", "tool_call"}, {"waiting", "waiting"},
		{"preparing", "preparing"}, {"stopped", "stopped"},
	} {
		activity := map[string]any{"agent_id": "child", "parent_id": "main", "phase": tc.wire}
		row := map[string]any{"id": "child", "parent_id": "main", "activity": activity}
		for _, frame := range []map[string]any{
			{"type": "agent_status", "agent_id": "child", "activity": activity},
			{"type": "agent_activity", "key": "main/child", "activity": activity},
			{"type": "agent_tree", "data": map[string]any{"agent_tree": []any{row}}},
		} {
			if got := managementActivityPhase(frame, "child", "main"); got != tc.want {
				t.Errorf("%#v -> %q, want %q", frame, got, tc.want)
			}
		}
	}
	for _, frame := range []map[string]any{
		{"type": "agent_status", "agent_id": "child", "replay": true, "activity": map[string]any{"phase": "waiting_for_model"}},
		{"type": "agent_status", "agent_id": "foreign", "activity": map[string]any{"phase": "waiting_for_model"}},
		{"type": "agent_status", "agent_id": "child", "activity": map[string]any{"agent_id": "foreign", "phase": "waiting_for_model"}},
		{"type": "agent_status", "agent_id": "child", "activity": map[string]any{"parent_id": "foreign", "phase": "waiting_for_model"}},
		{"type": "agent_activity", "key": "foreign/child", "activity": "thinking"},
		{"type": "tool_called", "arguments": map[string]any{"agent_id": "child", "phase": "thinking"}},
		{"type": "agent_tree", "data": []any{map[string]any{"id": "child", "parent_id": "foreign", "activity": "thinking"}}},
	} {
		if got := managementActivityPhase(frame, "child", "main"); got != "" {
			t.Errorf("foreign/forged phase credited: %#v -> %q", frame, got)
		}
	}
}

// Fresh2's public browser capture uses a nested top-level roster and one
// selected conversation snapshot, without a redundant selected-ID field.
func TestChapter21CurrentNestedRosterAndSnapshot(t *testing.T) {
	child := map[string]any{"id": "child", "parent_id": "main", "activity": map[string]any{"phase": "waiting"}}
	root := map[string]any{"id": "main", "children": []any{child}}
	frame := map[string]any{"type": "agent_tree", "agents": []any{root}}
	if managementTreeNode(frame, "child", true) == nil || managementActivityPhase(frame, "child", "main") != "waiting" {
		t.Fatal("lost nested public roster/activity")
	}
	for _, bad := range []any{
		[]any{root, child},
		[]any{map[string]any{"id": "main", "children": []any{map[string]any{"id": "child", "parent_id": "foreign"}}}},
		[]any{map[string]any{"id": "main", "children": "invalid"}},
		[]any{map[string]any{"id": "main", "children": []any{map[string]any{"id": "child", "agent_id": "foreign"}}}},
	} {
		if managementTree(map[string]any{"type": "agent_tree", "agents": bad}, true) {
			t.Fatalf("malformed nested roster accepted: %#v", bad)
		}
	}
	replay := map[string]any{"type": "replay", "agent_id": "", "data": map[string]any{
		"agents": []any{map[string]any{"id": "child", "state": map[string]any{"turn": "idle"}}},
		"events": []any{map[string]any{"type": "part_final", "agent_id": "child", "part": map[string]any{"type": "text", "text": "SELECTED"}}, map[string]any{"type": "part_final", "agent_id": "foreign", "part": map[string]any{"type": "text", "text": "FOREIGN"}}},
	}}
	selected := managementSelectionFrames([]map[string]any{replay}, "child")
	if managementSelected(replay) != "child" || managementStateWith(selected, true) != "idle" || !managementContains(selected, "SELECTED") || managementContains(selected, "FOREIGN") {
		t.Fatalf("lost single-snapshot selection: %#v", selected)
	}
	data := object(replay["data"])
	for _, bad := range []any{
		[]any{}, []any{child},
		[]any{map[string]any{"id": "child", "state": map[string]any{"turn": "idle"}}, map[string]any{"id": "other", "state": map[string]any{"turn": "idle"}}},
		[]any{map[string]any{"id": "child", "agent_id": "other", "state": map[string]any{"turn": "idle"}}},
	} {
		data["agents"] = bad
		if managementSelected(replay) != "" {
			t.Fatalf("ambiguous/non-conversation snapshot selected: %#v", bad)
		}
	}
	data["agents"] = []any{map[string]any{"id": "child", "state": map[string]any{"turn": "idle"}}}
	data["selected_agent"] = "foreign"
	if len(managementSelectionFrames([]map[string]any{replay}, "child")) != 0 {
		t.Fatal("explicit foreign selection was replaced by inferred identity")
	}
	data["selected_agent"] = 42
	if managementSelected(replay) != "" {
		t.Fatal("malformed explicit selection fell back")
	}
}

func TestChapter21CurrentInputCommand(t *testing.T) {
	for _, current := range []bool{false, true} {
		kind := "prompt"
		if current {
			kind = "input"
		}
		want := map[string]string{"type": kind, "text": "unchanged input bytes"}
		if got := managementInputCommand("unchanged input bytes", current); !reflect.DeepEqual(got, want) {
			t.Fatalf("current=%v: %#v", current, got)
		}
	}
}

// GUI identities are roster identities, not the parent's local tool names.
func TestChapter21CurrentCanonicalSelection(t *testing.T) {
	const id = "main/worker23"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		var command map[string]any
		if err := conn.ReadJSON(&command); err != nil {
			t.Error(err)
			return
		}
		if command["type"] != "select_agent" || command["agent_id"] != id {
			t.Errorf("advertised identity rewritten: %#v", command)
			return
		}
		if err := conn.WriteJSON(map[string]any{"type": "replay", "data": map[string]any{
			"selected_agent_id": id,
			"agents":            []any{map[string]any{"id": id, "parent_id": "main", "state": map[string]any{"turn": "idle"}}},
		}}); err != nil {
			t.Error(err)
			return
		}
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()
	peer, err := openGUIPeer("ws" + strings.TrimPrefix(server.URL, "http"))
	if err != nil {
		t.Fatal(err)
	}
	defer peer.close()
	frames, err := managementSelectWith(peer, id, true)
	if err != nil || managementStateWith(frames, true) != "idle" {
		t.Fatalf("canonical selection: %#v (%v)", frames, err)
	}
}

func TestChapter21CurrentFlatActivity(t *testing.T) {
	const id = "main/helper"
	frame := func(rows ...any) map[string]any {
		return map[string]any{"type": "agents_update", "agent_id": "", "data": rows}
	}
	row := func(phase string) map[string]any {
		return map[string]any{"agent_id": id, "parent_id": "main", "state": "running", "phase": phase}
	}
	for phase, want := range map[string]string{"in_flight": "thinking", "executing_tools": "tool_call", "waiting": "waiting", "idle": "idle", "tools_pending": "tools_pending"} {
		if got := managementActivityPhase(frame(row(phase)), id, "main"); got != want {
			t.Errorf("phase %s = %q, want %q", phase, got, want)
		}
	}
	conflict := row("in_flight")
	conflict["id"] = "main/foreign"
	foreign := row("in_flight")
	foreign["agent_id"] = "main/foreign"
	replay := frame(row("in_flight"))
	replay["replay"] = true
	history := frame(row("in_flight"))
	history["type"] = "replay"
	for name, f := range map[string]map[string]any{
		"foreign": frame(foreign), "conflicting": frame(conflict), "duplicate": frame(row("in_flight"), row("in_flight")), "replayed": replay, "history": history,
	} {
		if got := managementActivityPhase(f, id, "main"); got != "" {
			t.Errorf("%s gave live phase %q", name, got)
		}
	}
}

func TestChapter21CurrentFlatRosterIdentity(t *testing.T) {
	row := func(id, parent string) map[string]any {
		return map[string]any{"agent_id": id, "parent_id": parent, "phase": "idle"}
	}
	frame := func(rows ...any) map[string]any { return map[string]any{"type": "agents_update", "data": rows} }
	for _, id := range []string{"helper", "main/helper"} {
		original := row(id, "main")
		f := frame(row("main", ""), original, row("other/helper", "other"))
		if !managementTree(f, true) || managementOwnedChildID(f, "helper", "main") != id {
			t.Fatalf("lost advertised ID %s: %#v", id, f)
		}
		if _, changed := original["id"]; changed {
			t.Fatal("normalizer mutated received frame")
		}
	}
	for name, f := range map[string]map[string]any{
		"foreign":      frame(row("other/helper", "other")),
		"wrong-parent": frame(row("main/helper", "other")),
		"ambiguous":    frame(row("helper", "main"), row("main/helper", "main")),
		"duplicate":    frame(row("main/helper", "main"), row("main/helper", "main")),
		"bad-parent":   frame(map[string]any{"agent_id": "main/helper", "parent_id": 7}),
		"conflicting":  frame(map[string]any{"id": "main/helper", "agent_id": "main/other", "parent_id": "main"}),
	} {
		if id := managementOwnedChildID(f, "helper", "main"); id != "" {
			t.Errorf("%s resolved %q", name, id)
		}
	}
	replay := map[string]any{"type": "replay", "data": map[string]any{"agents": []any{map[string]any{"id": "main", "state": map[string]any{"turn": "idle"}}}}}
	if managementTree(replay, true) {
		t.Fatal("selected-only replay credited as a complete roster")
	}
	object(replay["data"])["agent_tree"] = []any{row("main", ""), row("main/helper", "main")}
	if managementOwnedChildID(replay, "helper", "main") != "" {
		t.Fatal("replay credited as live child membership")
	}
	if !managementTree(replay, true) {
		t.Fatal("explicit full replay roster lost")
	}
}

func TestChapter21CurrentWaitFrameIdentity(t *testing.T) {
	call := func(agent, id, kind string) map[string]any {
		return map[string]any{"type": "part_final", "agent_id": agent, "part": map[string]any{"type": kind, "call_id": id, "name": "wait_for_agent_change"}}
	}
	for _, kind := range []string{"tool_call", "tool_result"} {
		good := call("main", "wait-change", kind)
		if managementLiveToolPart(good, "main", "wait-change", kind) == nil {
			t.Fatal("valid live card lost")
		}
		replay := call("main", "wait-change", kind)
		replay["replay"] = true
		text := call("main", "wait-change", kind)
		object(text["part"])["type"] = "text"
		for _, bad := range []map[string]any{replay, text, call("foreign", "wait-change", kind), call("main", "other", kind), {"type": "tool_called", "arguments": good}, {"type": "replay", "data": good}} {
			if managementLiveToolPart(bad, "main", "wait-change", kind) != nil {
				t.Fatalf("forged/stale/foreign tool card accepted: %#v", bad)
			}
		}
	}
	waiting := map[string]any{"type": "agents_update", "data": []any{map[string]any{"agent_id": "main", "phase": "waiting"}}}
	if managementActivityPhase(waiting, "main", "") != "waiting" {
		t.Fatal("flat root waiting observation lost")
	}
	object(waiting["data"].([]any)[0])["parent_id"] = "foreign"
	if managementActivityPhase(waiting, "main", "") != "" {
		t.Fatal("foreign parent credited as root waiting")
	}
	// Canonical legacy-key activity must not acquire a second parent prefix.
	activity := map[string]any{"type": "agent_activity", "key": "main/helper", "activity": map[string]any{"phase": "thinking"}}
	if managementActivityPhase(activity, "main/helper", "main") != "thinking" {
		t.Fatal("canonical activity key lost")
	}
}
