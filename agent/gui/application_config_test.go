package gui

import (
	"encoding/json"
	"ensemble/ensemble"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func applicationFixture(t *testing.T) (string, *ApplicationConfigStore) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	data := `{"version":1,"provider":{"vendor":"openai","model":"test-model","billing_mode":"api","credential_file":"key.txt","wake_on_lan_mac":"02:00:00:00:00:01"}}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	store, err := LoadApplicationConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	return path, store
}

func TestApplicationConfigDefaultsAndRelativePaths(t *testing.T) {
	path, store := applicationFixture(t)
	base := filepath.Dir(path)
	c := store.Snapshot()
	if store.Path() != path {
		t.Fatalf("path = %q", store.Path())
	}
	for name, got := range map[string]string{".": c.Workspace, "cr": c.DataDir, "save.json": c.SavePath, "event.log": c.EventLog, "auth": c.AuthDirectory, "key.txt": c.Provider.CredentialFile} {
		if want := filepath.Join(base, name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	defaults := DefaultApplicationConfig()
	if c.GUI != defaults.GUI || c.GUI.Theme != "dark" || !c.GUI.TTSEnabled || c.GUI.TTSSpeed != 1 || c.GUI.FontSize != 16 {
		t.Fatalf("defaults = %+v", c.GUI)
	}
	c.SkillsDir = "skills"
	c.Workspace = "../work"
	c.DataDir = "state"
	c.SavePath = "snapshots/agent.json"
	c.EventLog = "logs/events.jsonl"
	c.AuthDirectory = "login"
	c.PrimarySkill = "coder"
	if err := store.Save(c); err != nil {
		t.Fatal(err)
	}
	next, err := LoadApplicationConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	c = next.Snapshot()
	if c.SkillsDir != filepath.Join(base, "skills") || c.Workspace != filepath.Join(base, "../work") || c.SavePath != filepath.Join(base, "snapshots/agent.json") || c.EventLog != filepath.Join(base, "logs/events.jsonl") || c.AuthDirectory != filepath.Join(base, "login") {
		t.Fatalf("resolved = %+v", c)
	}
}

func TestApplicationConfigStrictJSON(t *testing.T) {
	valid := `{"version":1,"provider":{"vendor":"openai","model":"test","billing_mode":"api","credential_file":"key"}}`
	cases := []string{
		valid + ` {}`, valid + ` garbage`, "null", `[]`,
		strings.Replace(valid, `"version":1`, `"version":2`, 1),
		strings.Replace(valid, `"version":1,`, "", 1),
		strings.Replace(valid, `"model":"test"`, `"model":"test","api_key":"secret"`, 1),
		strings.Replace(valid, `"model":"test"`, `"model":"test","access_token":"secret"`, 1),
		strings.Replace(valid, `"version":1`, `"version":1,"refresh_token":"secret"`, 1),
		strings.Replace(valid, `"version":1`, `"version":1,"unknown":true`, 1),
		strings.Replace(valid, `"version":1`, `"version":1,"gui":{"unknown":true}`, 1),
		strings.Replace(valid, `"version":1`, `"version":1,"gui":{"model":"session"}`, 1),
		strings.Replace(valid, `"version":1`, `"version":1,"gui":{"application":{}}`, 1),
		strings.Replace(valid, `"version":1`, `"version":1,"gui":{"memory":{"soul":{"secret":"token"}}}`, 1),
	}
	for _, data := range cases {
		t.Run(data, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadApplicationConfig(path); err == nil {
				t.Fatal("accepted invalid config")
			}
		})
	}
}

func TestApplicationConfigInvalidAndNormalizedSettings(t *testing.T) {
	_, store := applicationFixture(t)
	original := store.Snapshot()
	cases := map[string]func(*ApplicationConfig){
		"version":                  func(c *ApplicationConfig) { c.Version = 3 },
		"vendor":                   func(c *ApplicationConfig) { c.Provider.Vendor = "other" },
		"model":                    func(c *ApplicationConfig) { c.Provider.Model = " " },
		"billing":                  func(c *ApplicationConfig) { c.Provider.BillingMode = "" },
		"api missing credential":   func(c *ApplicationConfig) { c.Provider.CredentialFile = "" },
		"codex missing credential": func(c *ApplicationConfig) { c.Provider.BillingMode = "codex"; c.Provider.CredentialFile = "" },
		"codex wrong vendor":       func(c *ApplicationConfig) { c.Provider.BillingMode = "codex"; c.Provider.Vendor = "anthropic" },
		"chatgpt wrong vendor": func(c *ApplicationConfig) {
			c.Provider.BillingMode = "chatgpt"
			c.Provider.Vendor = "gemini"
			c.Provider.CredentialFile = ""
		},
		"chatgpt credential": func(c *ApplicationConfig) { c.Provider.BillingMode = "chatgpt" },
		"url scheme":         func(c *ApplicationConfig) { c.Provider.BaseURL = "file:///tmp/key" },
		"url relative":       func(c *ApplicationConfig) { c.Provider.BaseURL = "/v1" },
		"url userinfo":       func(c *ApplicationConfig) { c.Provider.BaseURL = "https://user:secret@example.com" },
		"url no host":        func(c *ApplicationConfig) { c.Provider.BaseURL = "https://" },
		"effort":             func(c *ApplicationConfig) { c.Provider.ReasoningEffort = "invalid" },
		"summary":            func(c *ApplicationConfig) { c.Provider.ReasoningSummary = "invalid" },
		"wake-on-LAN MAC":    func(c *ApplicationConfig) { c.Provider.WakeOnLANMAC = "02:00:00:00:00" },
		"negative memory":    func(c *ApplicationConfig) { c.GUI.Memory.Soul.Budget = -1 },
		"memory watermarks": func(c *ApplicationConfig) {
			c.GUI.Memory.Soul.LowWatermark = 300
			c.GUI.Memory.Soul.HighWatermark = 100
		},
		"path nul": func(c *ApplicationConfig) { c.DataDir = "a\x00b" },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			c := original
			change(&c)
			if err := store.Save(c); err == nil {
				t.Fatal("accepted invalid config")
			}
			if !reflect.DeepEqual(store.Snapshot(), original) {
				t.Fatal("failed validation changed snapshot")
			}
		})
	}
	c := original
	c.GUI.Theme = "invalid"
	c.GUI.TTSSpeed = 99
	c.GUI.FontSize = 1
	c.GUI.MaxToolRounds = -1
	c.GUI.ContextTarget = 10
	c.GUI.LogRetention = -1
	if err := store.Save(c); err != nil {
		t.Fatal(err)
	}
	gui := store.Snapshot().GUI
	if gui.Theme != "dark" || gui.TTSSpeed != 5 || gui.FontSize != 12 || gui.MaxToolRounds != 0 || gui.ContextTarget != 20000 || gui.LogRetention != 0 {
		t.Fatalf("settings not normalized: %+v", gui)
	}
}

func TestApplicationConfigRuntimeCredentialsAndHostileEnvironment(t *testing.T) {
	for _, name := range []string{"LLM_VENDOR", "LLM_MODEL", "LLM_API_KEY", "LLM_BASE_URL", "LLM_WOL_MAC", "OPENAI_API_KEY", "OPENAI_MODEL", "OPENAI_BILLING_MODE", "EN_CODEX_AUTH_FILE", "CODEX_HOME", "HOME", "ENSEMBLE_WORKSPACE", "EN_SKILLS_DIR", "EN_PRIMARY_SKILL", "EN_DATA_DIR", "EN_DISABLE_STREAMING", "EN_RECALL_DISABLED", "EN_RECALL_JUDGE_MODEL", "OPENAI_REASONING_EFFORT", "OPENAI_REASONING_SUMMARY"} {
		t.Setenv(name, "hostile-environment")
	}
	path, store := applicationFixture(t)
	c := store.Snapshot()
	if _, _, err := c.RuntimeConfig(); err == nil {
		t.Fatal("missing key did not fail runtime")
	}
	if err := store.Save(c); err != nil {
		t.Fatalf("save read credentials: %v", err)
	}
	if err := os.WriteFile(c.Provider.CredentialFile, []byte("  private-api-key\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c.GUI.MaxToolRounds = 12
	c.GUI.ContextTarget = 23000
	c.GUI.LogRetention = 30
	c.GUI.Memory.Soul.Disabled = true
	c.SkillsDir = filepath.Join(filepath.Dir(path), "skills")
	c.PrimarySkill = "coder"
	cfg, spec, err := c.RuntimeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Vendor != ensemble.OpenAI || cfg.Model != "test-model" || cfg.APIKey != "private-api-key" || cfg.BillingMode != ensemble.CredentialAPIKey || cfg.BaseURL != "" || cfg.WakeOnLANMAC != "02:00:00:00:00:01" || cfg.CodexAuthFile != "" || cfg.DisableStreaming || cfg.ReasoningEffort != "" || cfg.ReasoningSummary != "" || cfg.Recall.Disabled || cfg.Recall.JudgeModel != "" {
		t.Fatalf("runtime differs from file: %+v", cfg)
	}
	if !cfg.BuiltinTools || cfg.MaxToolRounds != 12 || cfg.ContextTarget != 23000 || cfg.LogRetention != 30 || !cfg.Memory.Soul.Disabled || cfg.Workspace != c.Workspace || cfg.DataDir != c.DataDir || cfg.SavePath != c.SavePath || cfg.LogPath != c.EventLog || spec.SkillsDir != c.SkillsDir || spec.PrimarySkill != "coder" || !spec.EnableWebSearch {
		t.Fatal("runtime mapping failed")
	}
	if err := os.WriteFile(c.Provider.CredentialFile, []byte("rotated-key"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, _, err = c.RuntimeConfig()
	if err != nil || cfg.APIKey != "rotated-key" {
		t.Fatal("runtime did not reread credentials")
	}
	if err := store.Save(c); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "private-api-key") || strings.Contains(string(data), "rotated-key") || strings.Contains(string(data), "hostile-environment") {
		t.Fatal("persisted a secret or environment setting")
	}
	c.Provider.BillingMode = "codex"
	c.Provider.CredentialFile = filepath.Join(t.TempDir(), "unread-auth.json")
	cfg, _, err = c.RuntimeConfig()
	if err != nil || cfg.APIKey != "" || cfg.CodexAuthFile != c.Provider.CredentialFile || cfg.BillingMode != ensemble.CredentialCodex {
		t.Fatalf("codex runtime: %v", err)
	}
	c.Provider.BillingMode = "chatgpt"
	c.Provider.CredentialFile = ""
	cfg, _, err = c.RuntimeConfig()
	if err != nil || cfg.APIKey != "" || cfg.CodexAuthFile != "" || cfg.BillingMode != ensemble.CredentialChatGPT {
		t.Fatalf("chatgpt runtime: %v", err)
	}
}

func TestApplicationConfigPrivateAtomicSave(t *testing.T) {
	path, store := applicationFixture(t)
	c := store.Snapshot()
	c.GUI.Model = "session-only-model"
	c.GUI.Application = &ApplicationSettingsView{}
	c.Provider.Model = "updated"
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	old, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer old.Close()
	if err := store.Save(c); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}
	oldInfo, err := old.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(oldInfo, info) {
		t.Fatal("save modified old inode instead of atomic replacement")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var disk map[string]json.RawMessage
	if err := json.Unmarshal(data, &disk); err != nil {
		t.Fatal(err)
	}
	var gui map[string]json.RawMessage
	if err := json.Unmarshal(disk["gui"], &gui); err != nil {
		t.Fatal(err)
	}
	if _, ok := gui["model"]; ok {
		t.Fatal("persisted GUI.Model")
	}
	if _, ok := gui["application"]; ok {
		t.Fatal("persisted GUI.Application")
	}
	if store.Snapshot().GUI.Model != "" || store.Snapshot().GUI.Application != nil {
		t.Fatal("snapshot retained session model")
	}
	original := store.Snapshot()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	c.Provider.Model = "rejected"
	if err := store.Save(c); err == nil {
		t.Fatal("rename onto directory succeeded")
	}
	if !reflect.DeepEqual(store.Snapshot(), original) {
		t.Fatal("failed save changed snapshot")
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".application-config-") {
			t.Fatal("temporary file leaked")
		}
	}
}

func TestApplicationConfigConcurrentSnapshots(t *testing.T) {
	_, store := applicationFixture(t)
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 8 {
				c := store.Snapshot()
				if err := store.Save(c); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
}

func TestApplicationConfigExplicitRuntimeOptions(t *testing.T) {
	path, store := applicationFixture(t)
	c := store.Snapshot()
	if c.GUIPort != 8088 || !c.EnableWebSearch {
		t.Fatal("invalid launch defaults")
	}
	c.Provider.BillingMode = "chatgpt"
	c.Provider.CredentialFile = ""
	c.DisableStreaming = true
	c.EnableWebSearch = false
	c.SafeMode = true
	c.SandboxRoot = "sandbox"
	c.GUIPort = 9999
	c.Recall = ensemble.RecallSettings{Disabled: true, DisableJudge: true, JudgeModel: "judge", MaxSnippets: 7}
	if err := store.Save(c); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadApplicationConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	c = loaded.Snapshot()
	cfg, spec, err := c.RuntimeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.DisableStreaming || !cfg.Recall.Disabled || !cfg.Recall.DisableJudge || cfg.Recall.JudgeModel != "judge" || cfg.Recall.MaxSnippets != 7 || spec.EnableWebSearch || !spec.SafeMode || spec.SandboxRoot != filepath.Join(filepath.Dir(path), "sandbox") || c.GUIPort != 9999 {
		t.Fatal("file runtime options were lost")
	}
	for _, port := range []int{-1, 0, 65536} {
		c.GUIPort = port
		if err := store.Save(c); err == nil {
			t.Fatalf("accepted port %d", port)
		}
	}
}
