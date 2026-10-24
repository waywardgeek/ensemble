package common

import (
	"context"
	"fmt"
	"time"
)

// Auth is an application service, not conversation state or a model tool.
// Resolve is called at the HTTP boundary; credentials must not enter Config.
type Auth interface {
	Login(context.Context, func(string) error) (AuthStatus, error)
	Status(context.Context) (AuthStatus, error)
	Logout(context.Context, string, bool) error
	Resolve(context.Context, Config) (Credential, error)
	Reject(context.Context, Credential) error
	Mode(context.Context) (AuthMode, error)
	Models(context.Context, Config) ([]AuthModel, error)
}
type AuthMode string

const (
	AuthChatGPT AuthMode = "chatgpt"
	AuthAPI     AuthMode = "api"
)

type AuthModel struct {
	Slug string `json:"slug"`
	Name string `json:"display_name"`
}
type AuthStatus struct {
	Mode      AuthMode  `json:"mode"`
	Account   string    `json:"account,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Eligible  bool      `json:"eligible"`
	Reason    string    `json:"reason,omitempty"`
}

// Credential is short-lived transport data. It is deliberately not serializable.
type Credential struct {
	Mode       AuthMode `json:"-"`
	Token      string   `json:"-"`
	Account    string   `json:"-"`
	Profile    string   `json:"-"`
	Generation uint64   `json:"-"`
}

// responsesModel uses the same exact-row contract as media and streaming.
func responsesModel(c Config) *ResponsesFeatures {
	model := c.ResolvedModel
	if model == "" {
		model = c.Model
	}
	features, _ := LookupModel(model)
	return features.Responses
}

// ResponsesCapabilities never infers support from a provider or model prefix.
func ResponsesCapabilities(c Config) (effort, summary bool) {
	if c.Vendor != OpenAI || c.Surface != Responses {
		return false, false
	}
	f := responsesModel(c)
	if f == nil {
		return false, false
	}
	return f.DefaultEffort != "", f.Summary
}
func NormalizeOpenAI(c *Config) error {
	switch c.AuthMode {
	case "", AuthChatGPT, AuthAPI:
	default:
		return fmt.Errorf("OpenAI billing mode must be api or chatgpt; automatic fallback is not implemented")
	}
	if c.AuthMode == AuthChatGPT && c.Vendor == 0 {
		c.Vendor = OpenAI
	}
	if c.AuthMode == AuthChatGPT && c.Vendor != OpenAI {
		return fmt.Errorf("ChatGPT auth requires OpenAI Responses")
	}
	if c.Vendor != OpenAI {
		return nil
	}
	if c.Surface == 0 && c.AuthMode == AuthChatGPT {
		c.Surface = Responses
	}
	if c.Surface == 0 {
		if responsesModel(*c) != nil {
			c.Surface = Responses
		} else {
			c.Surface = ChatCompletions
		}
	}
	if c.AuthMode == AuthChatGPT && c.Surface != Responses {
		return fmt.Errorf("ChatGPT auth requires OpenAI Responses")
	}
	if c.AuthMode == AuthChatGPT {
		if c.DisableStreaming || c.StreamCapability == StreamPlain {
			return fmt.Errorf("ChatGPT plan inference requires streaming; select API billing for stream-off")
		}
		c.StreamCapability = StreamSSE
	}
	e, s := ResponsesCapabilities(*c)
	if c.Surface == Responses {
		if f := responsesModel(*c); f != nil {
			c.RequireToolExplanation = f.RequireToolExplanation
		}
	}
	// Exact public capability identities are not unresolved routing aliases.
	// Unknown models still require explicit ResolvedModel evidence for replay.
	if e && c.ResolvedModel == "" {
		c.ResolvedModel = c.Model
	}
	switch c.ReasoningEffort {
	case "", "none", "minimal", "low", "medium", "high", "xhigh":
	default:
		return fmt.Errorf("invalid reasoning effort")
	}
	if c.ReasoningSummary != "" && c.ReasoningSummary != "none" && c.ReasoningSummary != "auto" && c.ReasoningSummary != "concise" && c.ReasoningSummary != "detailed" {
		return fmt.Errorf("invalid reasoning summary")
	}
	if e && c.ReasoningEffort == "" {
		c.ReasoningEffort = responsesModel(*c).DefaultEffort
	}
	if s && c.ReasoningSummary == "" {
		c.ReasoningSummary = "auto"
	}
	return nil
}
