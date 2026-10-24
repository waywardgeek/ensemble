package mcp

import (
	"encoding/json"
	"testing"
)

func TestBrowserObservationModes(t *testing.T) {
	m := &manager{}
	for _, name := range []string{"gui_snapshot", "tts_queue", "gui_click", "gui_input", "peer_context"} {
		for _, mode := range []string{"", "round", "turn"} {
			t.Run(name+"/"+mode, func(t *testing.T) {
				tool, err := m.toolSpec("browser-debug", remoteTool{Name: name, Description: "peer tool", InputSchema: json.RawMessage(`{"type":"object"}`), Ephemeral: mode})
				wantError := mode != "" && (name == "gui_snapshot" || name == "tts_queue")
				if (err != nil) != wantError {
					t.Fatalf("mode %q: err=%v, want error=%v", mode, err, wantError)
				}
				if err == nil && tool.Ephemeral != mode {
					t.Fatalf("automatic identity changed: %q -> %q", mode, tool.Ephemeral)
				}
			})
		}
	}
}
