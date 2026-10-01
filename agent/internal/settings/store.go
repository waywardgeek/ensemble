// Package settings owns the persisted settings store: defaults, clamping,
// and reading and writing settings.json. It is a spoke; the common.Settings value it
// produces is shared vocabulary and lives in internal/common.
package settings

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

func clamp(s *common.Settings) {
	if s.Temperature < common.MinTemperature {
		s.Temperature = common.MinTemperature
	}
	if s.Temperature > common.MaxTemperature {
		s.Temperature = common.MaxTemperature
	}
	if s.MaxTokens < 0 {
		s.MaxTokens = 0
	}
	if s.MaxTokens > common.MaxTokensCeiling {
		s.MaxTokens = common.MaxTokensCeiling
	}
	if s.ThinkingBudget < 0 {
		s.ThinkingBudget = 0
	}
	if s.ThinkingBudget > common.MaxThinkingBudget {
		s.ThinkingBudget = common.MaxThinkingBudget
	}
	if s.MaxToolRounds < 0 {
		s.MaxToolRounds = 0
	}
	if s.MaxToolRounds > common.MaxToolRoundsCap {
		s.MaxToolRounds = common.MaxToolRoundsCap
	}
	if s.FontSize < 0 {
		s.FontSize = 0
	} else if s.FontSize > 0 && s.FontSize < common.MinFontSize {
		s.FontSize = common.MinFontSize
	}
	if s.FontSize > common.MaxFontSize {
		s.FontSize = common.MaxFontSize
	}
	// Unset resolves to the default here rather than in the client, so the
	// broadcast always carries a concrete number and the GUI never has to
	// keep its own copy of the default.
	if s.MaxEvents <= 0 {
		s.MaxEvents = common.DefaultMaxEvents
	} else if s.MaxEvents < common.MinMaxEvents {
		s.MaxEvents = common.MinMaxEvents
	} else if s.MaxEvents > common.MaxMaxEventsCap {
		s.MaxEvents = common.MaxMaxEventsCap
	}
	if s.TTSSpeed < 0 {
		s.TTSSpeed = 0
	} else if s.TTSSpeed > 0 && s.TTSSpeed < common.MinTTSSpeed {
		s.TTSSpeed = common.MinTTSSpeed
	}
	if s.TTSSpeed > common.MaxTTSSpeed {
		s.TTSSpeed = common.MaxTTSSpeed
	}
	if s.ContextTarget < 0 {
		s.ContextTarget = 0
	} else if s.ContextTarget > 0 && s.ContextTarget < common.MinContextTarget {
		s.ContextTarget = common.MinContextTarget
	}
	if s.LogRetention < 0 {
		s.LogRetention = 0
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
	data common.Settings
	path string // empty = no persistence
}

// NewSettingsStore creates a store. If path is non-empty and the file
// exists, settings are loaded from it.
func NewSettingsStore(path string) *SettingsStore {
	s := &SettingsStore{path: path}
	if path != "" {
		if raw, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(raw, &s.data)
			clamp(&s.data)
		}
	}
	return s
}

// Get returns a snapshot of the current settings.
func (s *SettingsStore) Get() common.Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

// ApplyRaw merges a raw JSON patch. This preserves bool false values
// that omitempty would skip in a struct-level merge.
func (s *SettingsStore) ApplyRaw(raw json.RawMessage) common.Settings {
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
	if v, ok := patch["max_events"]; ok {
		var n int
		if json.Unmarshal(v, &n) == nil {
			s.data.MaxEvents = n
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
	if v, ok := patch["context_target"]; ok {
		var n int
		if json.Unmarshal(v, &n) == nil {
			s.data.ContextTarget = n
		}
	}
	if v, ok := patch["log_retention"]; ok {
		var n int
		if json.Unmarshal(v, &n) == nil {
			s.data.LogRetention = n
		}
	}

	clamp(&s.data)
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
