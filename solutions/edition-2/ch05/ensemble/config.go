package ensemble

import (
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
	cfg := Config{Workspace: os.Getenv("ENSEMBLE_WORKSPACE")}
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
	cfg.ResolvedModel = os.Getenv("LLM_RESOLVED_MODEL")
	return cfg, nil
}
