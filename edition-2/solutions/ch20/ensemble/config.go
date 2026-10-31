package ensemble

import (
	"ensemble/internal/common"
	"errors"
	"os"
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
	cfg.ReasoningEffort = os.Getenv("OPENAI_REASONING_EFFORT")
	cfg.ReasoningSummary = os.Getenv("OPENAI_REASONING_SUMMARY")
	cfg.ResolvedModel = os.Getenv("LLM_RESOLVED_MODEL")
	return cfg, nil
}

// AgentSpecFromEnv translates CLI settings into explicit per-Agent identity.
// CUSTOM_VAR is the exercise's application variable, not a core environment reader.
func AgentSpecFromEnv() AgentSpec {
	return AgentSpec{EnableWebSearch: os.Getenv("EN_WEB_SEARCH") != "0", SandboxRoot: os.Getenv("EN_SANDBOX_ROOT"), SafeMode: os.Getenv("EN_SAFE_MODE") == "1", SkillsDir: os.Getenv("EN_SKILLS_DIR"), PrimarySkill: os.Getenv("EN_PRIMARY_SKILL"), Variables: map[string]func() string{
		"CUSTOM_VAR": func() string { return os.Getenv("EN_CUSTOM_VAR") },
	}}
}
