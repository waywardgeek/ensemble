package common

import (
	"encoding/json"
	"os"
	"sync"
)

// Settings holds GUI-editable configuration. Fields that affect agent
// behavior (Model, Temperature, MaxTokens, ThinkingBudget, MaxToolRounds,
// SystemPrompt) are applied to the engine's Config on change. Fields that
// are purely cosmetic (Theme, TTS*, FontSize) are stored and broadcast
// to clients but do not touch the agent.
//
// Zero values mean "not set" — a patch with omitempty only overwrites
// fields the client explicitly included.
// Settings is the full, authoritative state of every user preference.
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
func (s *Settings) clamp() {
	if s.Temperature < MinTemperature {
		s.Temperature = MinTemperature
	}
	if s.Temperature > MaxTemperature {
		s.Temperature = MaxTemperature
	}
	if s.MaxTokens < 0 {
		s.MaxTokens = 0
	}
	if s.MaxTokens > MaxTokensCeiling {
		s.MaxTokens = MaxTokensCeiling
	}
	if s.ThinkingBudget < 0 {
		s.ThinkingBudget = 0
	}
	if s.ThinkingBudget > MaxThinkingBudget {
		s.ThinkingBudget = MaxThinkingBudget
	}
	if s.MaxToolRounds < 0 {
		s.MaxToolRounds = 0
	}
	if s.MaxToolRounds > MaxToolRoundsCap {
		s.MaxToolRounds = MaxToolRoundsCap
	}
	if s.FontSize < 0 {
		s.FontSize = 0
	} else if s.FontSize > 0 && s.FontSize < MinFontSize {
		s.FontSize = MinFontSize
	}
	if s.FontSize > MaxFontSize {
		s.FontSize = MaxFontSize
	}
	if s.TTSSpeed < 0 {
		s.TTSSpeed = 0
	} else if s.TTSSpeed > 0 && s.TTSSpeed < MinTTSSpeed {
		s.TTSSpeed = MinTTSSpeed
	}
	if s.TTSSpeed > MaxTTSSpeed {
		s.TTSSpeed = MaxTTSSpeed
	}
	switch s.Theme {
	case "", "dark", "light", "system":
	default:
		s.Theme = ""
	}
}

// SettingsStore is a thread-safe, persistent settings holder.
// The hub owns one; clients read and patch through it.
type SettingsStore struct {
	mu   sync.Mutex
	data Settings
	path string // empty = no persistence
}

// NewSettingsStore creates a store. If path is non-empty and the file
// exists, settings are loaded from it.
func NewSettingsStore(path string) *SettingsStore {
	s := &SettingsStore{path: path}
	if path != "" {
		if raw, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(raw, &s.data)
			s.data.clamp()
		}
	}
	return s
}

// Get returns a snapshot of the current settings.
func (s *SettingsStore) Get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}


// ApplyRaw merges a raw JSON patch. This preserves bool false values
// that omitempty would skip in a struct-level merge.
func (s *SettingsStore) ApplyRaw(raw json.RawMessage) Settings {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Unmarshal into a map to detect which fields were actually set.
	var patch map[string]json.RawMessage
	if json.Unmarshal(raw, &patch) != nil {
		return s.data
	}

	if v, ok := patch["model"]; ok {
		var str string
		if json.Unmarshal(v, &str) == nil && str != "" {
			s.data.Model = str
		}
	}
	if v, ok := patch["temperature"]; ok {
		var f float64
		if json.Unmarshal(v, &f) == nil {
			s.data.Temperature = f
		}
	}
	if v, ok := patch["max_tokens"]; ok {
		var n int
		if json.Unmarshal(v, &n) == nil {
			s.data.MaxTokens = n
		}
	}
	if v, ok := patch["thinking_budget"]; ok {
		var n int
		if json.Unmarshal(v, &n) == nil {
			s.data.ThinkingBudget = n
		}
	}
	if v, ok := patch["max_tool_rounds"]; ok {
		var n int
		if json.Unmarshal(v, &n) == nil {
			s.data.MaxToolRounds = n
		}
	}
	if v, ok := patch["system_prompt"]; ok {
		var str string
		if json.Unmarshal(v, &str) == nil {
			s.data.SystemPrompt = str
		}
	}
	if v, ok := patch["theme"]; ok {
		var str string
		if json.Unmarshal(v, &str) == nil && str != "" {
			s.data.Theme = str
		}
	}
	if v, ok := patch["font_size"]; ok {
		var n int
		if json.Unmarshal(v, &n) == nil {
			s.data.FontSize = n
		}
	}
	if v, ok := patch["tts_enabled"]; ok {
		var b bool
		if json.Unmarshal(v, &b) == nil {
			s.data.TTSEnabled = b
		}
	}
	if v, ok := patch["tts_speed"]; ok {
		var f float64
		if json.Unmarshal(v, &f) == nil {
			s.data.TTSSpeed = f
		}
	}

	s.data.clamp()
	s.persist()
	return s.data
}

// persist writes the current settings to disk. Caller must hold mu.
func (s *SettingsStore) persist() {
	if s.path == "" {
		return
	}
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(s.path, data, 0644)
}
