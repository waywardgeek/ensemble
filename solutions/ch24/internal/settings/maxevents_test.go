package settings

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// MaxEvents caps how many events the GUI keeps on screen. Unlike temperature
// or tts_enabled, zero is not a meaningful choice here: a screen showing no
// events is useless, so unset resolves to the default rather than being taken
// literally. "Effectively unlimited" is expressed as the cap, not as zero.
//
// Resolving it on the server rather than in the client means the broadcast
// always carries a concrete number, so the GUI never keeps its own copy of
// the default and the two cannot disagree.
func TestMaxEventsClamp(t *testing.T) {
	cases := []struct {
		name  string
		patch string
		want  int
	}{
		{"unset resolves to the default", `{}`, common.DefaultMaxEvents},
		{"explicit zero is not a useful screen", `{"max_events": 0}`, common.DefaultMaxEvents},
		{"negative is nonsense", `{"max_events": -5}`, common.DefaultMaxEvents},
		{"below the floor is raised", `{"max_events": 3}`, common.MinMaxEvents},
		{"a legal value survives untouched", `{"max_events": 250}`, 250},
		{"above the ceiling is capped", `{"max_events": 999999}`, common.MaxMaxEventsCap},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSettingsStore("")
			got := s.ApplyRaw(json.RawMessage(tc.patch))
			if got.MaxEvents != tc.want {
				t.Errorf("MaxEvents = %d, want %d", got.MaxEvents, tc.want)
			}
		})
	}
}
