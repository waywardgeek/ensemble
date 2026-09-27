package common

import "encoding/json"

// Settings is the full, authoritative state of every user preference.
//
// What is actually wired, as of 2026-09-26: ContextTarget and LogRetention
// reach the agent, both as live pulls set up in cmd/main.go. Theme, FontSize
// and the TTS fields are broadcast to clients and never touch the agent. The
// remaining six — Model, Temperature, MaxTokens, ThinkingBudget,
// MaxToolRounds and SystemPrompt — are stored, clamped and persisted, and
// nothing reads them. SettingsStore holds no engine reference and ApplyRaw
// cannot reach one, so changing those six has no effect; see TODO.md.
//
// An earlier version of this comment claimed those six were "applied to the
// engine's Config on change". That was never true, and it is recorded here
// because a comment asserting wiring that does not exist is worse than no
// comment: it stops the reader from checking.
//
// No field is omitempty, deliberately. Settings travels to the client as
// a complete snapshot, and a zero value is a real value: temperature 0,
// TTS disabled, a font size the server reset after rejecting a bad one.
// With omitempty those states vanish from the JSON and the client cannot
// distinguish "the server set this to zero" from "the server said
// nothing about this", so a corrected field silently keeps its old
// displayed value. Patches are expressed by key presence in ApplyRaw,
// never by zero, which is why nothing here needs to be omitted.
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

// Bounds for settings that reach the model API or the renderer. A value
// outside these ranges is not a preference, it is a bug or an attack: the
// API rejects temperature -5, and font size 0 renders nothing.
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

// clamp forces every field into its legal range. It is called on every
// path that can change settings: both merge functions and the load from
// disk, so a hand-edited settings.json cannot smuggle a bad value past
// the checks that the WebSocket patches go through.
//
// Zero means "not set" for most fields, and clamp preserves that. A
// negative value collapses to zero (unset, so the default applies)
// rather than to the minimum, because a client sending -5 has told us
// nothing about what it actually wants.

// SettingsSource is the websocket layer's view of the settings store: read the
// current settings, and apply a patch arriving from the client. The store
// itself - file I/O, clamping, persistence - lives in internal/settings.
type SettingsSource interface {
	Get() Settings
	ApplyRaw(raw json.RawMessage) Settings
}
