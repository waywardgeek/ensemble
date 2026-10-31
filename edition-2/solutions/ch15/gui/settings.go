package gui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Settings is a full snapshot of the GUI preferences and the Agent's saved turn budget. No field
// uses omitempty: false and zero are real values that every client must receive.
type Settings struct {
	// Theme selects dark, light, or the browser's system preference.
	Theme string `json:"theme"`
	// TTSEnabled controls automatic speech in connected browsers.
	TTSEnabled bool `json:"tts_enabled"`
	// TTSSpeed is the native speech rate, bounded to 0.5 through 2.
	TTSSpeed float64 `json:"tts_speed"`
	// FontSize is the interface text size in pixels, bounded to 12 through 24.
	FontSize int `json:"font_size"`
	// MaxToolRounds permits N tool batches per turn; zero selects default 200.
	MaxToolRounds int `json:"max_tool_rounds"`
	// ContextTarget sets the tool-byte target; zero selects 400,000.
	ContextTarget int `json:"context_target"`
	// LogRetention keeps this many saved events; zero keeps all.
	LogRetention int `json:"log_retention"`
}

// Decode onto a copy: omitted fields retain their value, explicit false/zero
// replace it, and a type error cannot leave a partially applied patch behind.
// The same boundary serves disk and wire so saved bad values cannot bypass it.
func settingsPatch(current Settings, raw json.RawMessage) (Settings, error) {
	if raw = bytes.TrimSpace(raw); len(raw) == 0 || raw[0] != '{' {
		return current, fmt.Errorf("settings must be a JSON object")
	}
	if err := json.Unmarshal(raw, &current); err != nil {
		return current, fmt.Errorf("invalid settings: %w", err)
	}
	switch current.Theme {
	case "dark", "light", "system":
	default:
		current.Theme = "dark"
	}
	current.TTSSpeed = max(0.5, min(2, current.TTSSpeed))
	current.FontSize = max(12, min(24, current.FontSize))
	current.MaxToolRounds = max(0, current.MaxToolRounds)
	current.ContextTarget = max(0, current.ContextTarget)
	if current.ContextTarget > 0 {
		current.ContextTarget = max(20000, current.ContextTarget)
	}
	current.LogRetention = max(0, current.LogRetention)
	return current, nil
}

// The optional GUI owns browser preferences and the startup envelope. The
// Agent owns the effective budget; snapshots always read it from that owner.
func (s *Server) snapshotSettings() Settings {
	snapshot := s.settings
	snapshot.MaxToolRounds = s.agent.Config().MaxToolRounds
	snapshot.ContextTarget = s.agent.Config().ContextTarget
	snapshot.LogRetention = s.agent.Config().LogRetention
	return snapshot
}
func (s *Server) loadSettings() error {
	s.settings = Settings{Theme: "dark", TTSEnabled: true, TTSSpeed: 1, FontSize: 16}
	data, err := os.ReadFile(s.settingsPath())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	next, err := settingsPatch(s.snapshotSettings(), data)
	if err != nil {
		return err
	}
	s.settings = next
	if err := s.agent.SetMaxToolRounds(next.MaxToolRounds); err != nil {
		return err
	}
	return s.agent.SetContextSettings(next.ContextTarget, next.LogRetention)
}
func (s *Server) settingsPath() string {
	return filepath.Join(s.agent.Config().DataDir, "settings.json")
}

// Called under Server.mu, together with broadcast and subscription. Persist
// before publishing so a failed write cannot claim a successful saved update.
func (s *Server) updateSettings(raw json.RawMessage) (Settings, error) {
	next, err := settingsPatch(s.snapshotSettings(), raw)
	if err != nil {
		return Settings{}, err
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return Settings{}, err
	}
	path := s.settingsPath()
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return Settings{}, err
	}
	// A short replacement write leaves the old file intact if writing fails.
	if err = os.WriteFile(path+".tmp", data, 0600); err == nil {
		err = os.Rename(path+".tmp", path)
	}
	if err != nil {
		os.Remove(path + ".tmp")
		return Settings{}, err
	}
	s.settings = next
	s.agent.SetMaxToolRounds(next.MaxToolRounds) // Already clamped by settingsPatch.
	s.agent.SetContextSettings(next.ContextTarget, next.LogRetention)
	return next, nil
}
