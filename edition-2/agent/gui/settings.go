package gui

import (
	"bytes"
	"encoding/json"
	"ensemble/ensemble"
	"fmt"
	"os"
	"path/filepath"
)

// Settings is a full snapshot of the GUI preferences and the Agent's saved turn budget. No field
// uses omitempty: false and zero are real values that every client must receive.
// ApplicationSettingsView distinguishes saved next-launch configuration from
// the provider selected when this process started. No credential contents appear.
type ApplicationSettingsView struct {
	Path            string           `json:"path"`
	Provider        ProviderSettings `json:"provider"`
	ActiveProvider  ProviderSettings `json:"active_provider"`
	RestartRequired bool             `json:"restart_required"`
}

type Settings struct {
	Application *ApplicationSettingsView `json:"application,omitempty"`
	// Model is the effective session selection; launch configuration wins on restart.
	Model string `json:"model"`
	// Memory controls the Agent corpus; full snapshots preserve nested false and zero values.
	Memory ensemble.MemorySettings `json:"memory"`
	// Theme selects dark, light, or the browser's system preference.
	Theme string `json:"theme"`
	// TTSEnabled controls automatic speech in connected browsers.
	TTSEnabled bool `json:"tts_enabled"`
	// TTSSpeed is the approximate speed multiplier, bounded to 0.5 through 5.
	// The browser calibrates its native rate for Eloquence above 2x.
	TTSSpeed float64 `json:"tts_speed"`
	// FontSize is the interface text size in pixels, bounded to 12 through 24.
	FontSize int `json:"font_size"`
	// MaxToolRounds is a legacy saved value, accepted but ignored by execution.
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
	current.TTSSpeed = max(0.5, min(5, current.TTSSpeed))
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
func (s *agentView) snapshotSettings() Settings {
	snapshot := s.settings
	snapshot.Model = s.agent.Config().Model
	snapshot.MaxToolRounds = s.agent.Config().MaxToolRounds
	snapshot.ContextTarget = s.agent.Config().ContextTarget
	snapshot.LogRetention = s.agent.Config().LogRetention
	snapshot.Memory = s.agent.Config().Memory
	snapshot.Application = nil
	if store := s.applicationStore(); store != nil {
		saved, active := store.Snapshot().Provider, s.parent.activeProvider
		active.Model = snapshot.Model
		snapshot.Application = &ApplicationSettingsView{Path: store.Path(), Provider: saved, ActiveProvider: active, RestartRequired: saved != active}
	}
	return snapshot
}
func (s *agentView) loadSettings() error {
	s.settings = Settings{Theme: "dark", TTSEnabled: true, TTSSpeed: 1, FontSize: 16}
	var data []byte
	var err error
	if store := s.applicationStore(); store != nil {
		data, err = json.Marshal(store.Snapshot().GUI)
	} else {
		data, err = os.ReadFile(s.settingsPath())
	}
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
	if err := s.agent.SetMemorySettings(next.Memory); err != nil {
		return err
	}
	return s.agent.SetContextSettings(next.ContextTarget, next.LogRetention)
}
func (s *agentView) applicationStore() *ApplicationConfigStore {
	if s.agent == s.parent.agent {
		return s.parent.application
	}
	return nil
}

func (s *agentView) settingsPath() string {
	return filepath.Join(s.agent.Config().DataDir, "settings.json")
}

// Called under settingsMu, together with broadcast and subscription. Never hold
// the observer mutex while waiting for the actor: streamed progress needs it.
// Persist before applying/publishing so a failed write leaves accepted state alone.
func (s *agentView) updateSettings(raw json.RawMessage) (Settings, error) {
	// Model selection is an immediate session operation, not a persisted GUI
	// preference. A separate patch avoids partially accepting a mixed update
	// when an active turn rejects the model change.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return Settings{}, err
	}
	if value, present := fields["provider"]; present {
		store := s.applicationStore()
		if store == nil {
			return Settings{}, fmt.Errorf("provider settings require a file-configured main agent")
		}
		if len(fields) != 1 {
			return Settings{}, fmt.Errorf("send provider settings separately from immediate preferences")
		}
		next := store.Snapshot()
		decoder := json.NewDecoder(bytes.NewReader(value))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&next.Provider); err != nil {
			return Settings{}, fmt.Errorf("invalid provider settings: %w", err)
		}
		if err := store.Save(next); err != nil {
			return Settings{}, err
		}
		s.parent.mu.Lock()
		snapshot := s.snapshotSettings()
		s.parent.mu.Unlock()
		return snapshot, nil
	}
	if _, present := fields["application"]; present {
		return Settings{}, fmt.Errorf("application metadata is read-only; send provider settings instead")
	}
	if value, present := fields["model"]; present {
		if s.applicationStore() != nil {
			return Settings{}, fmt.Errorf("edit provider.model; file-configured model changes apply after restart")
		}
		if len(fields) != 1 {
			return Settings{}, fmt.Errorf("send model selection separately from saved preferences")
		}
		var model string
		if err := json.Unmarshal(value, &model); err != nil {
			return Settings{}, err
		}
		if err := s.agent.SetModel(model); err != nil {
			return Settings{}, err
		}
		s.parent.mu.Lock()
		snapshot := s.snapshotSettings()
		s.parent.mu.Unlock()
		return snapshot, nil
	}
	s.parent.mu.Lock()
	current := s.snapshotSettings()
	s.parent.mu.Unlock()
	next, err := settingsPatch(current, raw)
	if err != nil {
		return Settings{}, err
	}
	// Validate before persistence; a rejected patch must not change the saved file.
	if err := ensemble.ValidateMemorySettings(next.Memory); err != nil {
		return Settings{}, err
	}
	// Shadow the display-only model field when writing preferences; a stored
	// model must not override an explicitly selected launch provider/credential.
	data, err := json.MarshalIndent(struct {
		// Settings supplies the existing persisted preferences.
		Settings
		// Model suppresses the embedded, session-only display field on disk.
		Model *string `json:"model,omitempty"`
	}{Settings: next}, "", "  ")
	if err != nil {
		return Settings{}, err
	}
	if store := s.applicationStore(); store != nil {
		config := store.Snapshot()
		config.GUI = next
		config.GUI.Application = nil
		config.GUI.Model = ""
		if err := store.Save(config); err != nil {
			return Settings{}, err
		}
	} else {
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
	}
	s.agent.SetMaxToolRounds(next.MaxToolRounds) // Already clamped by settingsPatch.
	s.agent.SetContextSettings(next.ContextTarget, next.LogRetention)
	if err := s.agent.SetMemorySettings(next.Memory); err != nil {
		return Settings{}, err
	}
	s.parent.mu.Lock()
	s.settings = next
	next = s.snapshotSettings()
	s.parent.mu.Unlock()
	return next, nil
}
