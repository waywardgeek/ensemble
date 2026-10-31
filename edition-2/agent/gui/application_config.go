package gui

import (
	"bytes"
	"encoding/json"
	"ensemble/ensemble"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ProviderSettings selects transport and billing without containing credentials.
// CredentialFile is a reference, never the contents of a credential file.
type ProviderSettings struct {
	Vendor           string `json:"vendor"`
	Model            string `json:"model"`
	BillingMode      string `json:"billing_mode"`
	BaseURL          string `json:"base_url"`
	CredentialFile   string `json:"credential_file"`
	ReasoningEffort  string `json:"reasoning_effort"`
	ReasoningSummary string `json:"reasoning_summary"`
}

// ApplicationConfig is the versioned, secret-free launch configuration.
// Relative paths are interpreted against the configuration file's directory.
type ApplicationConfig struct {
	Version          int                     `json:"version"`
	Provider         ProviderSettings        `json:"provider"`
	Workspace        string                  `json:"workspace"`
	SkillsDir        string                  `json:"skills_dir"`
	PrimarySkill     string                  `json:"primary_skill"`
	DataDir          string                  `json:"data_dir"`
	SavePath         string                  `json:"save_path"`
	EventLog         string                  `json:"event_log"`
	AuthDirectory    string                  `json:"auth_directory"`
	AuthIssuer       string                  `json:"auth_issuer"`
	GUI              Settings                `json:"gui"`
	DisableStreaming bool                    `json:"disable_streaming"`
	Recall           ensemble.RecallSettings `json:"recall"`
	EnableWebSearch  bool                    `json:"enable_web_search"`
	SandboxRoot      string                  `json:"sandbox_root"`
	SafeMode         bool                    `json:"safe_mode"`
	GUIPort          int                     `json:"gui_port"`
}

// The disk boundary deliberately lists preferences, excluding session model and
// application status. New display-only Settings fields cannot leak into files.
type applicationGUISettings struct {
	Memory        ensemble.MemorySettings `json:"memory"`
	Theme         string                  `json:"theme"`
	TTSEnabled    bool                    `json:"tts_enabled"`
	TTSSpeed      float64                 `json:"tts_speed"`
	FontSize      int                     `json:"font_size"`
	MaxToolRounds int                     `json:"max_tool_rounds"`
	ContextTarget int                     `json:"context_target"`
	LogRetention  int                     `json:"log_retention"`
}

func diskGUI(s Settings) applicationGUISettings {
	return applicationGUISettings{s.Memory, s.Theme, s.TTSEnabled, s.TTSSpeed, s.FontSize, s.MaxToolRounds, s.ContextTarget, s.LogRetention}
}
func (s applicationGUISettings) settings() Settings {
	return Settings{Memory: s.Memory, Theme: s.Theme, TTSEnabled: s.TTSEnabled, TTSSpeed: s.TTSSpeed, FontSize: s.FontSize, MaxToolRounds: s.MaxToolRounds, ContextTarget: s.ContextTarget, LogRetention: s.LogRetention}
}

type applicationConfigJSON ApplicationConfig

// MarshalJSON also protects callers that serialize snapshots outside Save.
func (c ApplicationConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		applicationConfigJSON
		GUI applicationGUISettings `json:"gui"`
	}{applicationConfigJSON(c), diskGUI(c.GUI)})
}

// DefaultApplicationConfig returns deterministic defaults, without consulting
// environment variables or other applications' configuration and credentials.
// Vendor and model must be supplied explicitly. Zero memory sizes select the
// library's common memory defaults, as they do in existing GUI preferences.
func DefaultApplicationConfig() ApplicationConfig {
	return ApplicationConfig{Version: 1, GUIPort: 8088, EnableWebSearch: true, Provider: ProviderSettings{BillingMode: "api"}, Workspace: ".", DataDir: "cr", SavePath: "save.json", EventLog: "event.log", AuthDirectory: "auth", GUI: Settings{Theme: "dark", TTSEnabled: true, TTSSpeed: 1, FontSize: 16}}
}

// ApplicationConfigStore owns a validated snapshot and serializes atomic saves.
type ApplicationConfigStore struct {
	mu     sync.RWMutex
	path   string
	config ApplicationConfig
}

// LoadApplicationConfig reads only this file, not credential files or the environment.
func LoadApplicationConfig(path string) (*ApplicationConfigStore, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("configuration path is required")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := DefaultApplicationConfig()
	c.Version = 0 // Version must be explicitly present in the document.
	wire := struct {
		*applicationConfigJSON
		GUI applicationGUISettings `json:"gui"`
	}{(*applicationConfigJSON)(&c), diskGUI(c.GUI)}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return nil, fmt.Errorf("invalid application config: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("application config must contain exactly one JSON object")
	}
	c.GUI = wire.GUI.settings()
	c, err = normalizeApplicationConfig(c, filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	return &ApplicationConfigStore{path: path, config: c}, nil
}

// Path returns the absolute configuration filename.
func (s *ApplicationConfigStore) Path() string { return s.path }

// Snapshot returns an independent copy of the last successfully saved configuration.
func (s *ApplicationConfigStore) Snapshot() ApplicationConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// Save validates before replacing the file with a private, same-directory
// temporary file. A failed write or rename never publishes the new snapshot.
func (s *ApplicationConfigStore) Save(c ApplicationConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := normalizeApplicationConfig(c, filepath.Dir(s.path))
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".application-config-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), s.path); err != nil {
		return err
	}
	s.config = c
	return nil
}

func validateApplicationURL(name, value string) error {
	if value == "" {
		return nil
	}
	u, err := url.Parse(value)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" {
		return fmt.Errorf("%s must be an http(s) URL without user information or fragment", name)
	}
	return nil
}

func normalizeApplicationConfig(c ApplicationConfig, base string) (ApplicationConfig, error) {
	if c.Version != 1 {
		return c, fmt.Errorf("unsupported application config version (expected 1)")
	}
	if c.GUIPort < 1 || c.GUIPort > 65535 {
		return c, fmt.Errorf("gui_port must be between 1 and 65535")
	}
	p := c.Provider
	switch p.Vendor {
	case "openai", "anthropic", "gemini":
	default:
		return c, fmt.Errorf("provider.vendor must be openai, anthropic, or gemini")
	}
	if strings.TrimSpace(p.Model) == "" {
		return c, fmt.Errorf("provider.model is required")
	}
	switch p.BillingMode {
	case "api":
		if strings.TrimSpace(p.CredentialFile) == "" {
			return c, fmt.Errorf("api billing requires credential_file")
		}
	case "codex":
		if p.Vendor != "openai" || strings.TrimSpace(p.CredentialFile) == "" {
			return c, fmt.Errorf("codex billing requires openai and credential_file")
		}
	case "chatgpt":
		if p.Vendor != "openai" || p.CredentialFile != "" {
			return c, fmt.Errorf("chatgpt billing requires openai and no credential_file")
		}
	default:
		return c, fmt.Errorf("provider.billing_mode must be api, codex, or chatgpt")
	}
	switch p.ReasoningEffort {
	case "", "none", "minimal", "low", "medium", "high", "xhigh":
	default:
		return c, fmt.Errorf("invalid provider.reasoning_effort")
	}
	switch p.ReasoningSummary {
	case "", "auto", "concise", "detailed":
	default:
		return c, fmt.Errorf("invalid provider.reasoning_summary")
	}
	if err := validateApplicationURL("provider.base_url", p.BaseURL); err != nil {
		return c, err
	}
	if err := validateApplicationURL("auth_issuer", c.AuthIssuer); err != nil {
		return c, err
	}
	if c.Workspace == "" {
		c.Workspace = "."
	}
	if c.DataDir == "" {
		c.DataDir = "cr"
	}
	if c.SavePath == "" {
		c.SavePath = "save.json"
	}
	if c.EventLog == "" {
		c.EventLog = "event.log"
	}
	if c.AuthDirectory == "" {
		c.AuthDirectory = "auth"
	}
	for _, path := range []*string{&c.Workspace, &c.SandboxRoot, &c.SkillsDir, &c.DataDir, &c.SavePath, &c.EventLog, &c.AuthDirectory, &c.Provider.CredentialFile} {
		if strings.ContainsRune(*path, 0) || (*path != "" && strings.TrimSpace(*path) == "") {
			return c, fmt.Errorf("invalid empty or NUL-containing configuration path")
		}
		if *path != "" && base != "" && !filepath.IsAbs(*path) {
			*path = filepath.Join(base, *path)
		}
	}
	// Rebuild only persistent fields so session model/application views are cleared.
	c.GUI = diskGUI(c.GUI).settings()
	raw, err := json.Marshal(diskGUI(c.GUI))
	if err != nil {
		return c, err
	}
	c.GUI, err = settingsPatch(c.GUI, raw)
	if err != nil {
		return c, err
	}
	if err = ensemble.ValidateMemorySettings(c.GUI.Memory); err != nil {
		return c, err
	}
	return c, nil
}

// RuntimeConfig maps the explicit configuration into the library. Only this
// operation reads API key files; Codex and ChatGPT credentials remain owned by
// their respective credential services. ConfigureAuth separately with the
// snapshot's AuthDirectory and AuthIssuer before constructing Agents.
func (c ApplicationConfig) RuntimeConfig() (ensemble.Config, ensemble.AgentSpec, error) {
	c, err := normalizeApplicationConfig(c, "")
	if err != nil {
		return ensemble.Config{}, ensemble.AgentSpec{}, err
	}
	vendor := map[string]ensemble.Vendor{"openai": ensemble.OpenAI, "anthropic": ensemble.Anthropic, "gemini": ensemble.Gemini}[c.Provider.Vendor]
	cfg := ensemble.Config{Vendor: vendor, Model: c.Provider.Model, BaseURL: c.Provider.BaseURL, BillingMode: ensemble.CredentialKind(c.Provider.BillingMode), ReasoningEffort: c.Provider.ReasoningEffort, ReasoningSummary: c.Provider.ReasoningSummary, Workspace: c.Workspace, DataDir: c.DataDir, SavePath: c.SavePath, LogPath: c.EventLog, BuiltinTools: true, MaxToolRounds: c.GUI.MaxToolRounds, ContextTarget: c.GUI.ContextTarget, LogRetention: c.GUI.LogRetention, Memory: c.GUI.Memory, DisableStreaming: c.DisableStreaming, Recall: c.Recall}
	if c.Provider.BillingMode == "codex" {
		cfg.CodexAuthFile = c.Provider.CredentialFile
	}
	if c.Provider.BillingMode == "api" && c.Provider.CredentialFile != "" {
		key, err := os.ReadFile(c.Provider.CredentialFile)
		if err != nil {
			return ensemble.Config{}, ensemble.AgentSpec{}, fmt.Errorf("cannot read provider credential_file: %w", err)
		}
		cfg.APIKey = strings.TrimSpace(string(key))
		if cfg.APIKey == "" {
			return ensemble.Config{}, ensemble.AgentSpec{}, fmt.Errorf("provider credential_file is empty")
		}
	}
	spec := ensemble.AgentSpec{SkillsDir: c.SkillsDir, PrimarySkill: c.PrimarySkill, EnableWebSearch: c.EnableWebSearch, SandboxRoot: c.SandboxRoot, SafeMode: c.SafeMode}
	return cfg, spec, nil
}
