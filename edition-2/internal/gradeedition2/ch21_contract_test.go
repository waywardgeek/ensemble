package gradeedition2

import "testing"

func TestChapter21ReplayEvents(t *testing.T) {
	text := map[string]any{"type": "part_final", "part": map[string]any{"type": "text", "text": "visible-result"}}
	state := map[string]any{"type": "state_changed", "to": "idle"}
	frames := managementReplayFrames([]map[string]any{
		{"type": "tool_called", "arguments": "hidden-marker"},
		{"type": "replay", "data": map[string]any{"events": []any{text, state}}},
		{"type": "current_settings"},
	})
	if !managementContains(frames, "visible-result") || managementState(frames) != "idle" {
		t.Fatalf("actual replay missing: %#v", frames)
	}
	if managementContains(frames, "hidden-marker") {
		t.Fatal("tool arguments credited as visible result")
	}
	if len(frames) != 5 {
		t.Fatalf("lost replay/trailer: %#v", frames)
	}
	direct := managementReplayFrames([]map[string]any{text, state})
	if len(direct) != 2 || !managementContains(direct, "visible-result") || managementState(direct) != "idle" {
		t.Fatal("legacy direct frames changed")
	}
}

func TestChapter21InheritedJoinShape(t *testing.T) {
	payload := fleetReviewPayload
	current := `{"outcomes":[{"agent_id":"worker23","parent_id":"manager","state":"submitted","submission":` + payload + `}]}`
	legacy := `[{"agent_id":"worker23","payload":` + payload + `}]`
	if !managementJoinedPayload(current, true) || !managementJoinedPayload(legacy, false) {
		t.Fatal("lost exact structured result")
	}
	for _, bad := range []string{`{"outcomes":[]}`, `{"outcomes":[{"submission":"encoded instead of typed"}]}`, `{"outcomes":[{"submission":null}]}`, `invalid`} {
		if managementJoinedPayload(bad, true) {
			t.Fatalf("accepted malformed/missing payload: %s", bad)
		}
	}
	if managementJoinedPayload(current, false) {
		t.Fatal("legacy parser silently changed")
	}
}
