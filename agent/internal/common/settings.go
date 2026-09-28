package common

import "encoding/json"

// Settings is the full, authoritative state of every user preference,
// held flat because this struct is the WebSocket wire contract.
//
// No field is omitempty, deliberately. Settings travels to the client as
// a complete snapshot, and a zero value is a real value: temperature 0,
// TTS disabled, a font size the server reset after rejecting a bad one.
type Settings struct {
	// Agent behavior.
	Model          string  `json:"model"`
	Temperature    float64 `json:"temperature"`
	MaxTokens      int     `json:"max_tokens"`
	ThinkingBudget int     `json:"thinking_budget"`
	MaxToolRounds  int     `json:"max_tool_rounds"`
	SystemPrompt   string  `json:"system_prompt"`

	// Appearance.
	Theme    string `json:"theme"` // "dark", "light", "system"
	FontSize int    `json:"font_size"`

	// Accessibility.
	TTSEnabled bool    `json:"tts_enabled"`
	TTSSpeed   float64 `json:"tts_speed"`

	// Context management (Chapter 15 rule 10). ContextTarget is the one
	// knob: the target context size in bytes, from which the stub
	// threshold and both ladder bands derive (see BudgetsFor); zero means
	// DefaultContextTarget. LogRetention is how many events the save file
	// keeps; zero keeps them all.
	ContextTarget int `json:"context_target"`
	LogRetention  int `json:"log_retention"`

	// Memory is the per-band configuration from chapter 16: which bands
	// are on, and how many bytes each may hold before it folds upward. A
	// settings file written before chapter 16 existed has no "memory" key,
	// so it deserializes to the zero value, which normalizes to every band
	// on at its default budget. That is why BandSettings says Disabled
	// rather than Enabled.
	Memory BandConfig `json:"memory"`
}

// Bounds for settings that reach the model API or the renderer.
const (
	MinTemperature    = 0.0
	MaxTemperature    = 2.0
	MaxTokensCeiling  = 1000000
	MaxThinkingBudget = 200000
	MaxToolRoundsCap  = 10000
	MinFontSize       = 8
	MaxFontSize       = 72
	MinTTSSpeed       = 0.1
	MaxTTSSpeed       = 10.0
)

// SettingsSource is the websocket layer's view of the settings store.
type SettingsSource interface {
	Get() Settings
	ApplyRaw(raw json.RawMessage) Settings
}
