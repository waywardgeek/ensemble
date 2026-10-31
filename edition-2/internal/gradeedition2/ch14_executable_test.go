package gradeedition2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestChapter14HandoffRecords(t *testing.T) {
	for _, executable := range []bool{false, true} {
		for _, wrong := range []string{"", "missing", "late", "failed", "wrong-kind"} {
			name := "go/" + wrong
			if executable {
				name = "csharp/" + wrong
			}
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				note := map[string]any{"seq": 4, "type": "micro_handoff", "handoff": map[string]string{"text": contextHandoffText}}
				if executable {
					kind := "handoff"
					if wrong == "wrong-kind" {
						kind = "text"
					}
					note = map[string]any{"seq": 4, "type": "handoff_recorded", "parts": []any{map[string]string{"type": kind, "text": contextHandoffText}}}
				} else if wrong == "wrong-kind" {
					note["type"] = "handoff_recorded"
				}
				events := []any{
					map[string]any{"seq": 1, "type": "tool_called", "tool": map[string]any{"call_id": "H2"}},
					map[string]any{"seq": 2, "type": "tool_returned", "tool": map[string]any{"call_id": "H2"}},
				}
				if wrong != "missing" {
					seq := 3
					if wrong == "late" {
						seq = 5
					}
					events = append(events, map[string]any{"seq": seq, "type": "tool_returned", "tool": map[string]any{
						"call_id": "H3", "is_error": wrong == "failed", "parts": []any{map[string]string{"type": "text", "text": contextFileMark(contextLadderFiles)}}}})
				}
				events = append(events, note)
				encoded, err := json.Marshal(map[string]any{"log": events})
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(dir, "save.json"), encoded, 0600); err != nil {
					t.Fatal(err)
				}
				result := contextHandoffLogFor(dir, executable)
				if (result == "") != (wrong == "") {
					t.Fatalf("wrong=%q, result=%q", wrong, result)
				}
			})
		}
	}
}
