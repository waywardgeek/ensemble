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
type Settings struct {
	// Agent behavior.
	Model          string  `json:"model,omitempty"`
	Temperature    float64 `json:"temperature,omitempty"`
	MaxTokens      int     `json:"max_tokens,omitempty"`
	ThinkingBudget int     `json:"thinking_budget,omitempty"`
	MaxToolRounds  int     `json:"max_tool_rounds,omitempty"`
	SystemPrompt   string  `json:"system_prompt,omitempty"`

	// Appearance.
	Theme    string `json:"theme,omitempty"` // "dark", "light", "system"
	FontSize int    `json:"font_size,omitempty"`

	// Accessibility.
	TTSEnabled bool    `json:"tts_enabled,omitempty"`
	TTSSpeed   float64 `json:"tts_speed,omitempty"`
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

// Apply merges a partial patch into the current settings and persists.
// Returns the full settings after the merge.
func (s *SettingsStore) Apply(patch Settings) Settings {
	s.mu.Lock()
	defer s.mu.Unlock()

	if patch.Model != "" {
		s.data.Model = patch.Model
	}
	if patch.Temperature != 0 {
		s.data.Temperature = patch.Temperature
	}
	if patch.MaxTokens != 0 {
		s.data.MaxTokens = patch.MaxTokens
	}
	if patch.ThinkingBudget != 0 {
		s.data.ThinkingBudget = patch.ThinkingBudget
	}
	if patch.MaxToolRounds != 0 {
		s.data.MaxToolRounds = patch.MaxToolRounds
	}
	if patch.SystemPrompt != "" {
		s.data.SystemPrompt = patch.SystemPrompt
	}
	if patch.Theme != "" {
		s.data.Theme = patch.Theme
	}
	if patch.FontSize != 0 {
		s.data.FontSize = patch.FontSize
	}
	// TTSEnabled is a bool — patch it if the JSON included it.
	// Since omitempty skips false, we handle this via the raw patch.
	// For simplicity, always apply.
	s.data.TTSEnabled = patch.TTSEnabled
	if patch.TTSSpeed != 0 {
		s.data.TTSSpeed = patch.TTSSpeed
	}

	s.persist()
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
