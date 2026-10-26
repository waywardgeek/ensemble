package gui

import (
	"bytes"
	"encoding/json"
	"ensemble/ensemble"
	"io"
	"reflect"
	"strings"
	"testing"
)

// TestTurnTimingProjection keeps measured zeros, dispatch counts and admission
// time intact across the shared live/replay wire projection.
func TestTurnTimingProjection(t *testing.T) {
	for _, timing := range []string{
		`{"started":"2026-10-25T12:00:00Z","total_ms":800,"model_ms":500,"tool_ms":200,"other_ms":100,"calls":3,"requests":2,"interrupted":true}`,
		`{"started":"2026-10-25T12:00:00Z","total_ms":0,"model_ms":0,"tool_ms":0,"other_ms":0,"calls":1,"requests":1,"interrupted":false}`,
		`{"started":"2026-10-25T12:00:00Z","total_ms":17,"model_ms":0,"tool_ms":0,"other_ms":17,"calls":0,"requests":0,"interrupted":false}`,
		`{"started":"2026-10-25T12:00:00Z","total_ms":17,"model_ms":0,"tool_ms":0,"other_ms":17,"other_phases_ms":{"context_copy":12,"unattributed":5},"calls":0,"requests":0,"interrupted":false}`,
	} {
		t.Run(timing, func(t *testing.T) {
			var event ensemble.Event
			if err := json.Unmarshal([]byte(`{"seq":42,"type":"turn_ended","time":"2026-10-26T01:00:00Z","turn":{"error":"fixture error","timing":`+timing+`}}`), &event); err != nil {
				t.Fatal(err)
			}
			projected := eventWire(event)
			if len(projected) != 1 {
				t.Fatalf("want one timing summary, got %v", projected)
			}
			encoded, err := json.Marshal(projected[0])
			if err != nil {
				t.Fatal(err)
			}
			var got, want map[string]any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(timing), &want); err != nil {
				t.Fatal(err)
			}
			want["type"], want["turn_seq"], want["error"] = "turn_timing", float64(42), "fixture error"
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("timing projection = %v, want %v", got, want)
			}
		})
	}
}

// TestTurnTimingLoadedMalformedHistory exercises retained malformed facts through
// real history loading, then dump/reload into a fresh Agent as on restart.
func TestTurnTimingLoadedMalformedHistory(t *testing.T) {
	valid := `{"started":"2026-10-25T12:00:00Z","total_ms":10,"model_ms":4,"tool_ms":3,"other_ms":3,"calls":1,"requests":1,"interrupted":false}`
	for _, test := range []struct {
		name, old, replacement string
	}{
		{"model exceeds total", `"model_ms":4`, `"model_ms":1000`},
		{"sum differs", `"other_ms":3`, `"other_ms":0`},
		{"negative total", `"total_ms":10`, `"total_ms":-1`},
		{"negative model", `"model_ms":4`, `"model_ms":-1`},
		{"negative tool", `"tool_ms":3`, `"tool_ms":-1`},
		{"negative other", `"other_ms":3`, `"other_ms":-1`},
		{"negative calls", `"calls":1`, `"calls":-1`},
		{"negative requests", `"requests":1`, `"requests":-1`},
		{"missing total", `"total_ms":10,`, ``},
		{"missing start", `"started":"2026-10-25T12:00:00Z",`, ``},
		{"zero start", `2026-10-25T12:00:00Z`, `0001-01-01T00:00:00Z`},
		{"model without requests", `"requests":1`, `"requests":0`},
		{"tools without calls", `"calls":1`, `"calls":0`},
		{"unsafe calls", `"calls":1`, `"calls":9007199254740992`},
		{"unsafe requests", `"requests":1`, `"requests":9007199254740992`},
		{"unsafe duration", `"total_ms":10,"model_ms":4,"tool_ms":3,"other_ms":3`, `"total_ms":9007199254740992,"model_ms":0,"tool_ms":0,"other_ms":9007199254740992`},
		{"phase sum differs", `"other_ms":3`, `"other_ms":3,"other_phases_ms":{"context_copy":2}`},
		{"negative phase", `"other_ms":3`, `"other_ms":3,"other_phases_ms":{"context_copy":-1,"unattributed":4}`},
		{"empty phase name", `"other_ms":3`, `"other_ms":3,"other_phases_ms":{"":3}`},
		{"phase exceeds other", `"other_ms":3`, `"other_ms":3,"other_phases_ms":{"context_copy":9007199254740992}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			timing := strings.Replace(valid, test.old, test.replacement, 1)
			log := `{"seq":42,"type":"turn_ended","turn":{"timing":` + timing + `}}` + "\n" +
				`{"seq":43,"type":"turn_ended"}` + "\n" +
				`{"seq":44,"type":"turn_ended","turn":{"timing":{"started":"2026-10-25T12:00:00Z","total_ms":0,"model_ms":0,"tool_ms":0,"other_ms":0,"calls":1,"requests":1,"interrupted":false}}}`
			for restart := 0; restart < 2; restart++ {
				root := ensemble.New(io.Discard)
				agent, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "local-fixture", DataDir: t.TempDir()})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { agent.Shutdown() })
				if err := agent.History().Load(strings.NewReader(log)); err != nil {
					t.Fatal(err)
				}
				events := agent.History().EventsAfter(0)
				if len(events) != 3 || events[0].Turn == nil || events[0].Turn.Timing == nil {
					t.Fatalf("malformed fact not retained by History.Load: %v", events)
				}
				for _, event := range events[:2] {
					if got := eventWire(event); len(got) != 0 {
						t.Fatalf("restart %d projected invalid/legacy timing: %v", restart, got)
					}
				}
				if got := eventWire(events[2]); len(got) != 1 || got[0]["total_ms"] != int64(0) {
					t.Fatalf("restart %d lost measured zero: %v", restart, got)
				}
				var saved bytes.Buffer
				if err := agent.History().Dump(&saved); err != nil {
					t.Fatal(err)
				}
				log = saved.String()
			}
		})
	}
}

// TestTurnTimingLegacyAbsent prevents capture timestamps or old closing events
// from being presented as runtime measurements.
func TestTurnTimingLegacyAbsent(t *testing.T) {
	for _, payload := range []string{
		`{"seq":1,"type":"turn_ended","time":"2026-10-25T12:00:00Z"}`,
		`{"seq":2,"type":"turn_ended","turn":{"text":"legacy","error":"legacy error"}}`,
	} {
		var event ensemble.Event
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			t.Fatal(err)
		}
		if got := eventWire(event); len(got) != 0 {
			t.Fatalf("legacy turn acquired inferred timing: %v", got)
		}
	}
}
