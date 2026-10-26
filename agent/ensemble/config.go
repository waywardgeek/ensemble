package ensemble

import (
	"ensemble/internal/common"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ConfigFromEnv reads provider settings without sending a request. Model IDs
// are explicit: applications choose a discovered model rather than a stale
// built-in default. Builtin tools and data directories are application choices.
// Generic settings select any provider; provider-prefixed settings retain the
// Chapter 1 interface and make interactive setup convenient for all three.
func ConfigFromEnv() (Config, error) {
	vendor := os.Getenv("LLM_VENDOR")
	if vendor == "" {
		vendor = "anthropic"
	}
	cfg := Config{Workspace: os.Getenv("ENSEMBLE_WORKSPACE"), DisableStreaming: os.Getenv("EN_DISABLE_STREAMING") == "1"}
	if value := os.Getenv("EN_STREAMING"); value != "" {
		streams, err := ParseStream(value)
		if err != nil {
			return cfg, err
		}
		cfg.Streaming = streams
	}
	switch vendor {
	case "anthropic":
		cfg.Vendor = Anthropic
	case "openai":
		cfg.Vendor = OpenAI
	case "gemini":
		cfg.Vendor = Gemini
	default:
		return cfg, errors.New("unknown LLM_VENDOR")
	}
	env := func(name string) string {
		if value := os.Getenv("LLM_" + name); value != "" {
			return value
		}
		return os.Getenv(strings.ToUpper(vendor) + "_" + name)
	}
	cfg.BaseURL, cfg.APIKey, cfg.Model = env("BASE_URL"), env("API_KEY"), env("MODEL")
	cfg.BillingMode = common.CredentialKind(os.Getenv("OPENAI_BILLING_MODE"))
	// External credentials are consulted only after explicit route selection.
	if cfg.BillingMode == common.CredentialCodex {
		cfg.CodexAuthFile = os.Getenv("EN_CODEX_AUTH_FILE")
		if cfg.CodexAuthFile == "" {
			directory := os.Getenv("CODEX_HOME")
			if directory == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return cfg, err
				}
				directory = filepath.Join(home, ".codex")
			}
			cfg.CodexAuthFile = filepath.Join(directory, "auth.json")
		}
	}
	cfg.ReasoningEffort = os.Getenv("OPENAI_REASONING_EFFORT")
	cfg.ReasoningSummary = os.Getenv("OPENAI_REASONING_SUMMARY")
	cfg.ResolvedModel = os.Getenv("LLM_RESOLVED_MODEL")
	// Recall uses this provider by default; no second credential source is hidden.
	cfg.Recall.JudgeModel = os.Getenv("EN_RECALL_JUDGE_MODEL")
	cfg.Recall.Disabled = os.Getenv("EN_RECALL_DISABLED") == "1"
	cfg.Recall.DisableJudge = os.Getenv("EN_RECALL_DISABLE_JUDGE") == "1"
	return cfg, nil
}

// ParseStream accepts "all" to enable incremental delivery or empty input to
// use the model default. Delivery does not filter the provider's output channels.
func ParseStream(value string) (Stream, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return 0, nil
	case "all":
		return StreamAll, nil
	default:
		return 0, errors.New("invalid streaming selection: use all or leave empty for the model default")
	}
}

// AgentSpecFromEnv translates CLI settings into explicit per-Agent identity.
// CUSTOM_VAR is the exercise's application variable, not a core environment reader.
func AgentSpecFromEnv() AgentSpec {
	return AgentSpec{EnableWebSearch: os.Getenv("EN_WEB_SEARCH") != "0", SandboxRoot: os.Getenv("EN_SANDBOX_ROOT"), SafeMode: os.Getenv("EN_SAFE_MODE") == "1", SkillsDir: os.Getenv("EN_SKILLS_DIR"), PrimarySkill: os.Getenv("EN_PRIMARY_SKILL"), Variables: map[string]func() string{
		"CUSTOM_VAR": func() string { return os.Getenv("EN_CUSTOM_VAR") },
	}}
}
