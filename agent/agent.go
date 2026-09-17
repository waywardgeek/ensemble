// Package agent provides a reusable framework for building LLM-powered coding
// agents. External programs import this package to create agents, register
// custom tools, and run interactive or scripted sessions.
//
// The shared type vocabulary lives in agent/internal/common and is re-exported
// here so that external callers never import internal/ directly.
package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/waywardgeek/coding-agents-course/agent/internal/common"
	"github.com/waywardgeek/coding-agents-course/agent/internal/jobs"
	"github.com/waywardgeek/coding-agents-course/agent/internal/llm"
	"github.com/waywardgeek/coding-agents-course/agent/internal/tools"
)

// Re-export the types external programs need.
type Config = common.Config
type ToolDecl = common.ToolDecl
type Vendor = common.Vendor
type Surface = common.Surface
type Usage = common.Usage

const (
	VendorAnthropic = common.VendorAnthropic
	VendorOpenAI    = common.VendorOpenAI
	VendorGemini    = common.VendorGemini
)

// DefaultSurface returns the default API surface for a vendor.
var DefaultSurface = common.DefaultSurface

// ToolHandler is the signature for a custom tool: given JSON arguments,
// return the text the model will see.
type ToolHandler func(args json.RawMessage) (string, error)

// Agent is the public handle to a running agent.
type Agent struct {
	eng *llm.Engine
}

// NewAgent creates a new agent with the given config and log path. It wires
// up the builtin tools (run_command, read_file, etc.) and any custom tools
// registered before this call. Constructors take interfaces to parents: the
// engine receives the job manager and tool registry as interfaces, not
// concrete types.
func NewAgent(cfg Config, logPath string) *Agent {
	j := jobs.NewJobs()
	reg := tools.NewRegistry()
	return &Agent{eng: llm.NewEngine(cfg, logPath, j, reg)}
}

// RegisterTool adds a custom tool to the global registry. Call this before
// NewAgent. The schema is the JSON Schema for the tool's input parameters.
func RegisterTool(name, description string, schema json.RawMessage, handler ToolHandler) {
	tools.Register(name, description, schema, handler)
}

// Ask sends a prompt and runs the full tool loop until the model replies.
func (a *Agent) Ask(prompt string) (string, error) {
	return a.eng.Ask(prompt)
}

// Shutdown ends all running jobs and saves the log.
func (a *Agent) Shutdown() error {
	return a.eng.Shutdown()
}

// Usage returns the token counts accumulated across all requests.
func (a *Agent) Usage() Usage {
	return a.eng.Ctx.Usage
}

// ConfigFromEnv builds a Config from environment variables (LLM_VENDOR,
// LLM_MODEL, LLM_API_KEY, LLM_BASE_URL). This is the standard way for an
// external program to configure the framework without hardcoding vendor
// details.
func ConfigFromEnv() (Config, error) {
	vendor, err := parseVendor(envOr("LLM_VENDOR", "anthropic"))
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Vendor:    vendor,
		Surface:   common.DefaultSurface(vendor),
		MaxTokens: 1024,
		Tools:     tools.Declarations(),
	}
	switch vendor {
	case VendorAnthropic:
		cfg.Model = pick("LLM_MODEL", "ANTHROPIC_MODEL", "claude-sonnet-5")
		cfg.BaseURL = pick("LLM_BASE_URL", "ANTHROPIC_BASE_URL", "https://api.anthropic.com")
		cfg.APIKey = pick("LLM_API_KEY", "ANTHROPIC_API_KEY", "")
	case VendorOpenAI:
		cfg.Model = pick("LLM_MODEL", "OPENAI_MODEL", "gpt-5")
		cfg.BaseURL = pick("LLM_BASE_URL", "OPENAI_BASE_URL", "https://api.openai.com")
		cfg.APIKey = pick("LLM_API_KEY", "OPENAI_API_KEY", "")
	case VendorGemini:
		cfg.Model = pick("LLM_MODEL", "GEMINI_MODEL", "gemini-3.8-flash")
		cfg.BaseURL = pick("LLM_BASE_URL", "GEMINI_BASE_URL", "https://generativelanguage.googleapis.com")
		cfg.APIKey = pick("LLM_API_KEY", "GEMINI_API_KEY", "")
	}
	return cfg, nil
}

func parseVendor(s string) (Vendor, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "anthropic", "claude":
		return VendorAnthropic, nil
	case "openai":
		return VendorOpenAI, nil
	case "gemini":
		return VendorGemini, nil
	}
	return 0, fmt.Errorf("unknown vendor %q (want anthropic, openai or gemini)", s)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func pick(primary, secondary, def string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return envOr(secondary, def)
}
