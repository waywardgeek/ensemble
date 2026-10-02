package ws

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// fakeSettings is a SettingsSource holding one value.
type fakeSettings struct{ s common.Settings }

func (f *fakeSettings) Get() common.Settings { return f.s }
func (f *fakeSettings) ApplyRaw(raw json.RawMessage) common.Settings {
	return f.s
}

// TestSettingsForDisplayFillsUnsetModel pins the rule that the GUI is never
// told an empty model.
//
// An unset model means "follow the startup default". A picker handed an empty
// string falls back to its first option, which is some other vendor's model
// entirely, and then reports a model the agent is not running. The usage meter
// had this exact failure: it confirmed a model every request ignored.
func TestSettingsForDisplayFillsUnsetModel(t *testing.T) {
	h := &Hub{
		settings: &fakeSettings{s: common.Settings{Model: ""}},
		Model:    func() string { return "gpt-6.1-sol" },
	}

	got := h.settingsForDisplay(h.settings.Get())

	if got.Model != "gpt-6.1-sol" {
		t.Fatalf("unset model should be filled with the model in force, got %q", got.Model)
	}
}

// TestSettingsForDisplayKeepsChosenModel pins the other half: a model the
// operator already chose is left alone.
//
// The actor applies a model switch asynchronously, so for a moment after the
// change the live model is still the OLD one. Overwriting unconditionally
// would broadcast that stale value and snap the picker back to the previous
// model immediately after the operator changed it.
func TestSettingsForDisplayKeepsChosenModel(t *testing.T) {
	h := &Hub{
		settings: &fakeSettings{s: common.Settings{Model: "claude-opus-5-5"}},
		Model:    func() string { return "gpt-6.1-sol" }, // switch not applied yet
	}

	got := h.settingsForDisplay(h.settings.Get())

	if got.Model != "claude-opus-5-5" {
		t.Fatalf("a chosen model must survive; stale live model overwrote it: %q", got.Model)
	}
}
