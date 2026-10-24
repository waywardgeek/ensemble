package gui

import (
	"encoding/json"
	"ensemble"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// Settings is a complete GUI-local snapshot. False and zero stay on the wire.
// Read Aloud controls playback, not capture; Agent Interrupt remains separate.
type Settings struct {
	SafeMode             bool             `json:"safe_mode"`
	EnableWebSearch      bool             `json:"enable_web_search"`
	Theme                string           `json:"theme"`
	PauseWhileSpeaking   bool             `json:"pause_while_speaking"`
	TTSEnabled           bool             `json:"tts_enabled"`
	TTSSpeed             float64          `json:"tts_speed"`
	FontSize             int              `json:"font_size"`
	LogMode              ensemble.LogMode `json:"log_mode"`
	DiagnosticMaxRecords int              `json:"diagnostic_max_records"`
	ContextTarget        int              `json:"context_target"`
	MaxToolRounds        int              `json:"max_tool_rounds"`
}

func DefaultSettings() Settings {
	return Settings{EnableWebSearch: true, Theme: "dark", TTSEnabled: true, PauseWhileSpeaking: true, TTSSpeed: 1, FontSize: 16, LogMode: ensemble.LogFull, DiagnosticMaxRecords: 200}
}

// A sparse patch is decoded onto a copy, then validated as a whole. The same
// path clamps wire and disk input; a type error cannot publish half an update.
func patchSettings(current Settings, raw json.RawMessage) (Settings, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return current, errors.New("settings must be a JSON object")
	}
	next := current
	for key, value := range fields {
		if string(value) == "null" {
			return current, fmt.Errorf("%s cannot be null", key)
		}
		var target any
		switch key {
		case "safe_mode":
			target = &next.SafeMode
		case "enable_web_search":
			target = &next.EnableWebSearch
		case "theme":
			target = &next.Theme
		case "pause_while_speaking":
			target = &next.PauseWhileSpeaking
		case "tts_enabled":
			target = &next.TTSEnabled
		case "tts_speed":
			target = &next.TTSSpeed
		case "font_size":
			target = &next.FontSize
		case "log_mode":
			target = &next.LogMode
		case "diagnostic_max_records":
			target = &next.DiagnosticMaxRecords
		case "context_target":
			target = &next.ContextTarget
		case "max_tool_rounds":
			target = &next.MaxToolRounds
		default:
			return current, fmt.Errorf("unknown settings field %q", key)
		}
		if err := json.Unmarshal(value, target); err != nil {
			return current, fmt.Errorf("invalid value for %s", key)
		}
	}
	switch next.Theme {
	case "dark", "light", "system":
	default:
		next.Theme = "dark"
	}
	if math.IsNaN(next.TTSSpeed) || math.IsInf(next.TTSSpeed, 0) {
		return current, errors.New("tts_speed must be finite")
	}
	next.TTSSpeed = max(0.5, min(5, next.TTSSpeed))
	next.FontSize = max(12, min(24, next.FontSize))
	next.MaxToolRounds = max(0, next.MaxToolRounds)
	if next.ContextTarget > 0 {
		next.ContextTarget = max(20000, next.ContextTarget)
	}
	if next.ContextTarget < 0 {
		return current, errors.New("context_target must be nonnegative")
	}
	if next.DiagnosticMaxRecords < 0 {
		return current, errors.New("diagnostic_max_records must be nonnegative")
	}
	if next.LogMode != ensemble.LogFull && next.LogMode != ensemble.LogDialogueOnly {
		return current, errors.New("invalid log_mode")
	}
	return next, nil
}

// The hub mutex serializes save, runtime application and publication. This is
// atomic replacement of one file, not a claim of power-loss durability.
type settingsStore struct {
	path  string
	agent ensemble.Agent
}

// Only absent settings fields inherit launch policy. Both startup preload and
// already-attached stores use this default and the same patchSettings parser.
func settingsForLaunch(config ensemble.Config) Settings {
	defaults := DefaultSettings()
	defaults.SafeMode, defaults.EnableWebSearch = config.SafeMode, config.EnableWebSearch
	return defaults
}

func (s settingsStore) load() (Settings, error) {
	defaults := DefaultSettings()
	// Only missing fields inherit launch policy; explicit persisted choices win.
	if s.agent != nil {
		defaults = settingsForLaunch(s.agent.Config())
	}
	if s.path == "" {
		return defaults, nil
	}
	readFile := os.ReadFile
	if s.agent != nil {
		readFile = s.agent.Sandbox().ReadFile
	}
	raw, err := readFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return defaults, nil
	}
	if err != nil {
		return defaults, err
	}
	return patchSettings(defaults, raw)
}
func (s settingsStore) save(value Settings) error {
	if s.path == "" {
		return nil
	} // Standalone in-memory hubs have no persistence path.
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	mkdir, createTemp, remove, rename := os.MkdirAll, os.CreateTemp, os.Remove, os.Rename
	if s.agent != nil {
		files := s.agent.Sandbox()
		mkdir = files.MkdirAll
		createTemp = files.CreateTemp
		remove = files.Remove
		rename = files.Rename
	}
	dir := filepath.Dir(s.path)
	if err = mkdir(dir, 0700); err != nil {
		return err
	}
	file, err := createTemp(dir, ".settings-")
	if err != nil {
		return err
	}
	name := file.Name()
	defer remove(name)
	if _, err = file.Write(append(data, '\n')); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return rename(name, s.path)
}

// Read budget policy back from its owner, not a second GUI runtime budget.
// A GUI view never owns policy for unrelated roots or delegated Agents.
func (h *Hub) currentSettings() Settings {
	value := h.settings
	if owner := h.settingsAgent(); owner != nil {
		cfg := owner.Config()
		value.SafeMode, value.EnableWebSearch = cfg.SafeMode, cfg.EnableWebSearch
		value.ContextTarget = cfg.ContextTarget
		value.MaxToolRounds = cfg.MaxToolRounds
		value.LogMode = cfg.LogMode
		value.DiagnosticMaxRecords = *cfg.DiagnosticMaxRecords
	}
	return value
}
func (h *Hub) updateSettings(raw json.RawMessage) error {
	h.settingsMu.Lock()
	defer h.settingsMu.Unlock()
	h.mu.Lock()
	next, err := patchSettings(h.currentSettings(), raw)
	if err != nil {
		h.mu.Unlock()
		return err
	}
	owner := h.settingsAgent()
	if owner != nil && owner.Config().ParentID != "" {
		if parent, ok := h.en.Agent(owner.Config().ParentID); ok {
			inherited := parent.Config()
			next.SafeMode = next.SafeMode || inherited.SafeMode
			next.EnableWebSearch = next.EnableWebSearch && inherited.EnableWebSearch
		}
	}
	if err = h.store.save(next); err != nil {
		h.mu.Unlock()
		return fmt.Errorf("could not save settings: %w", err)
	}
	h.settings = next
	var agents []ensemble.Agent
	if owner != nil {
		agents = []ensemble.Agent{owner}
	}
	h.mu.Unlock()
	// Capability notifications re-enter Hub.mu. Keep policy setters outside that
	// lock while settingsMu preserves save/apply/publication ordering.
	for _, a := range agents {
		a.SetSecurityPolicy(next.SafeMode, next.EnableWebSearch)
		_ = a.SetContextTarget(next.ContextTarget)
		_ = a.SetMaxToolRounds(next.MaxToolRounds)
		_ = a.SetPersistencePolicy(next.LogMode, next.DiagnosticMaxRecords)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	settings := h.currentSettings()
	h.broadcast(Wire{Type: "settings_changed", Settings: &settings})
	return nil
}

func (h *Hub) settingsAgent() ensemble.Agent {
	if h.store.agent != nil {
		return h.store.agent
	}
	agents := h.en.Agents()
	if len(agents) > 0 {
		return agents[0]
	}
	return nil
}
