package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// A client can send anything. These are values a hostile or buggy client
// actually sends, and none of them may reach the model API.
func TestApplyRawClampsOutOfRangeValues(t *testing.T) {
	s := NewSettingsStore("")

	got := s.ApplyRaw(json.RawMessage(`{
		"temperature": -5,
		"max_tokens": -100,
		"thinking_budget": -1,
		"max_tool_rounds": -7,
		"font_size": -3,
		"tts_speed": -2,
		"theme": "'; DROP TABLE settings;--"
	}`))

	if got.Temperature != MinTemperature {
		t.Errorf("Temperature = %v, want %v", got.Temperature, MinTemperature)
	}
	if got.MaxTokens != 0 {
		t.Errorf("MaxTokens = %v, want 0", got.MaxTokens)
	}
	if got.ThinkingBudget != 0 {
		t.Errorf("ThinkingBudget = %v, want 0", got.ThinkingBudget)
	}
	if got.MaxToolRounds != 0 {
		t.Errorf("MaxToolRounds = %v, want 0", got.MaxToolRounds)
	}
	if got.FontSize != 0 {
		t.Errorf("FontSize = %v, want 0 (unset)", got.FontSize)
	}
	if got.TTSSpeed != 0 {
		t.Errorf("TTSSpeed = %v, want 0 (unset)", got.TTSSpeed)
	}
	if got.Theme != "" {
		t.Errorf("Theme = %q, want \"\" (unknown theme rejected)", got.Theme)
	}
}

func TestApplyRawClampsAboveMaximum(t *testing.T) {
	s := NewSettingsStore("")

	got := s.ApplyRaw(json.RawMessage(`{
		"temperature": 99,
		"max_tokens": 1073741824,
		"thinking_budget": 1073741824,
		"max_tool_rounds": 1073741824,
		"font_size": 9000,
		"tts_speed": 1000
	}`))

	if got.Temperature != MaxTemperature {
		t.Errorf("Temperature = %v, want %v", got.Temperature, MaxTemperature)
	}
	if got.MaxTokens != MaxTokensCeiling {
		t.Errorf("MaxTokens = %v, want %v", got.MaxTokens, MaxTokensCeiling)
	}
	if got.ThinkingBudget != MaxThinkingBudget {
		t.Errorf("ThinkingBudget = %v, want %v", got.ThinkingBudget, MaxThinkingBudget)
	}
	if got.MaxToolRounds != MaxToolRoundsCap {
		t.Errorf("MaxToolRounds = %v, want %v", got.MaxToolRounds, MaxToolRoundsCap)
	}
	if got.FontSize != MaxFontSize {
		t.Errorf("FontSize = %v, want %v", got.FontSize, MaxFontSize)
	}
	if got.TTSSpeed != MaxTTSSpeed {
		t.Errorf("TTSSpeed = %v, want %v", got.TTSSpeed, MaxTTSSpeed)
	}
}

// A font size of 3 is not zero, so it is a real request, but it renders
// nothing. Small positive values rise to the minimum rather than
// collapsing to unset.
func TestApplyRawRaisesSmallPositivesToMinimum(t *testing.T) {
	s := NewSettingsStore("")
	got := s.ApplyRaw(json.RawMessage(`{"font_size": 3, "tts_speed": 0.01}`))

	if got.FontSize != MinFontSize {
		t.Errorf("FontSize = %v, want %v", got.FontSize, MinFontSize)
	}
	if got.TTSSpeed != MinTTSSpeed {
		t.Errorf("TTSSpeed = %v, want %v", got.TTSSpeed, MinTTSSpeed)
	}
}

// Legal values must survive untouched, or the clamp is just breakage.
func TestApplyRawLeavesLegalValuesAlone(t *testing.T) {
	s := NewSettingsStore("")
	want := Settings{
		Model:          "claude-sonnet-4-5",
		Temperature:    0.7,
		MaxTokens:      4096,
		ThinkingBudget: 8000,
		MaxToolRounds:  50,
		SystemPrompt:   "be brief",
		Theme:          "dark",
		FontSize:       18,
		TTSEnabled:     true,
		TTSSpeed:       1.5,
	}
	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.ApplyRaw(raw); got != want {
		t.Errorf("legal settings were modified:\n got %+v\nwant %+v", got, want)
	}
}

// Only the keys present in the patch may change. Absence means "leave it
// alone", which is why patches are presence-based rather than using zero
// as a sentinel.
func TestApplyRawPatchesOnlyPresentKeys(t *testing.T) {
	s := NewSettingsStore("")
	s.ApplyRaw(json.RawMessage(`{"theme": "dark", "font_size": 20}`))

	got := s.ApplyRaw(json.RawMessage(`{"font_size": 14}`))

	if got.Theme != "dark" {
		t.Errorf("Theme = %q, want \"dark\" (absent from patch, must persist)", got.Theme)
	}
	if got.FontSize != 14 {
		t.Errorf("FontSize = %v, want 14", got.FontSize)
	}
}

// Zero is a real value, not an absence. A client must be able to turn TTS
// off and set temperature to exactly 0.
func TestApplyRawAcceptsMeaningfulZeroes(t *testing.T) {
	s := NewSettingsStore("")
	s.ApplyRaw(json.RawMessage(`{"tts_enabled": true, "temperature": 1.5}`))

	got := s.ApplyRaw(json.RawMessage(`{"tts_enabled": false, "temperature": 0}`))

	if got.TTSEnabled {
		t.Error("TTSEnabled = true, want false (client explicitly disabled it)")
	}
	if got.Temperature != 0 {
		t.Errorf("Temperature = %v, want 0 (an explicit, legal choice)", got.Temperature)
	}
}

// The snapshot sent to the client must carry every field. If a corrected
// field were omitted, the client could not tell "server set this to zero"
// from "server said nothing", and would keep displaying the bad value the
// user typed.
func TestSettingsJSONAlwaysIncludesEveryField(t *testing.T) {
	raw, err := json.Marshal(Settings{})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{
		"model", "temperature", "max_tokens", "thinking_budget",
		"max_tool_rounds", "system_prompt", "theme", "font_size",
		"tts_enabled", "tts_speed",
	} {
		if _, ok := fields[key]; !ok {
			t.Errorf("zero-valued Settings omits %q from JSON; the client "+
				"cannot learn the server reset this field", key)
		}
	}
}

// Validating only the WebSocket patch leaves a hole: settings.json is a
// file, and files get hand-edited.
func TestLoadFromDiskClamps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	raw, err := json.Marshal(map[string]any{
		"temperature": -5,
		"font_size":   9000,
		"theme":       "nonsense",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	got := NewSettingsStore(path).Get()

	if got.Temperature != MinTemperature {
		t.Errorf("Temperature = %v, want %v", got.Temperature, MinTemperature)
	}
	if got.FontSize != MaxFontSize {
		t.Errorf("FontSize = %v, want %v", got.FontSize, MaxFontSize)
	}
	if got.Theme != "" {
		t.Errorf("Theme = %q, want \"\"", got.Theme)
	}
}
