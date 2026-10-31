package gradeedition2

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

func chapter7WireFixture() []map[string]any {
	return []map[string]any{
		{"type": "part_delta", "part_id": float64(1), "kind": "text", "chunk": guiIntro},
		{"type": "part_final", "part_id": float64(1), "part": map[string]any{"type": "text", "text": guiIntro}},
		{"type": "part_final", "part_id": float64(2), "part": map[string]any{"type": "tool_call", "name": "run_command", "call_id": "gui_call", "args": map[string]any{"command": "printf GUI_TOOL_RESULT"}}},
		{"type": "tool_dispatched", "part_id": float64(2), "call_id": "gui_call", "name": "run_command", "input": map[string]any{"command": "printf GUI_TOOL_RESULT"}},
		{"type": "tool_finished", "part_id": float64(3), "call_id": "gui_call", "result": "GUI_TOOL_RESULT", "is_error": false},
		{"type": "part_delta", "part_id": float64(4), "kind": "text", "chunk": guiAnswer},
		{"type": "part_final", "part_id": float64(4), "part": map[string]any{"type": "text", "text": guiAnswer}},
		{"type": "state_changed", "from": "idle", "to": "idle"},
		{"type": "turn_ended", "text": guiAnswer},
	}
}

func TestChapter7FinalMetadataUpdate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(map[string]any)
		pass   bool
	}{
		{"job metadata", func(p map[string]any) { p["job"] = map[string]any{"id": float64(7)} }, true},
		{"different call", func(p map[string]any) { p["call_id"] = "another_call" }, false},
		{"different arguments", func(p map[string]any) { p["args"] = map[string]any{"command": "wrong"} }, false},
		{"different kind", func(p map[string]any) { p["type"] = "thinking" }, false},
		{"different role", func(p map[string]any) { p["role"] = "human" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wire := chapter7WireFixture()
			b, err := json.Marshal(wire[2])
			if err != nil {
				t.Fatal(err)
			}
			var update map[string]any
			if err = json.Unmarshal(b, &update); err != nil {
				t.Fatal(err)
			}
			tc.change(object(update["part"]))
			wire = append(wire, update)
			c := grade.Check{Passed: true}
			checkGUIWire(&c, wire)
			if c.Passed != tc.pass {
				t.Fatalf("passed=%v want %v: %v", c.Passed, tc.pass, c.Details)
			}
		})
	}
}

func TestChapter7ReplayBatch(t *testing.T) {
	live := chapter7WireFixture()
	var events []any
	for _, m := range live {
		if m["type"] != "part_delta" {
			events = append(events, m)
		}
	}
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
		pass   bool
	}{
		{"complete", func(map[string]any) {}, true},
		{"missing events", func(d map[string]any) { delete(d, "events") }, false},
		{"missing final", func(d map[string]any) { a := d["events"].([]any); d["events"] = a[1:] }, false},
		{"stale partial", func(d map[string]any) {
			d["partials"] = []any{map[string]any{"type": "part_delta", "part_id": float64(1), "kind": "text", "chunk": "stale"}}
		}, false},
		{"changed content", func(d map[string]any) { a := d["events"].([]any); object(object(a[0])["part"])["text"] = "wrong" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(map[string]any{"events": events, "partials": []any{}, "agents": []any{}})
			if err != nil {
				t.Fatal(err)
			}
			var data map[string]any
			if err = json.Unmarshal(b, &data); err != nil {
				t.Fatal(err)
			}
			tc.mutate(data)
			frames := guiReplayFrames([]map[string]any{{"type": "replay", "data": data}})
			c := grade.Check{Passed: true}
			checkGUIReplay(&c, live, frames)
			if c.Passed != tc.pass {
				t.Fatalf("passed=%v want %v: %v", c.Passed, tc.pass, c.Details)
			}
			if tc.pass && len(oldMessages(frames)) != len(events) {
				t.Fatal("legacy evaluator did not receive batch contents")
			}
		})
	}
	// Older individual-frame replay is still accepted without changing its contents.
	frames := make([]map[string]any, 0, len(events))
	for _, e := range events {
		frames = append(frames, object(e))
	}
	c := grade.Check{Passed: true}
	checkGUIReplay(&c, live, guiReplayFrames(frames))
	if !c.Passed {
		t.Fatal(c.Details)
	}
}

func TestChapter7FinalOwnerIdentity(t *testing.T) {
	for _, key := range []string{"agent_id", "turn_id"} {
		t.Run(key, func(t *testing.T) {
			wire := chapter7WireFixture()
			wire[2][key] = float64(1)
			b, err := json.Marshal(wire[2])
			if err != nil {
				t.Fatal(err)
			}
			var update map[string]any
			if err = json.Unmarshal(b, &update); err != nil {
				t.Fatal(err)
			}
			update[key] = float64(2)
			c := grade.Check{Passed: true}
			checkGUIWire(&c, append(wire, update))
			if c.Passed {
				t.Fatalf("identical content reused under different %s passed", key)
			}
		})
	}
}
