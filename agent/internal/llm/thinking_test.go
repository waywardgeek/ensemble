package llm

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/coding-agents-course/agent/internal/common"
)

// TestAnthropicThinkingRendered verifies that a rendered Anthropic request
// contains a thinking block and the correct wire format for the model.
//
// Claude-sonnet-5 and claude-opus-5 use ADAPTIVE thinking:
//   thinking: { "type": "adaptive" }
//   output_config: { "effort": "high" }
//
// This is a regression guard for the bug where Claude showed NO thinking
// because anthRequest had no thinking field. The fake vendor concealed it
// by scripting thinking unconditionally, while real Claude never thinks
// unless the request asks.
func TestAnthropicThinkingRendered(t *testing.T) {
	ctx := &common.Context{
		Dialogue: []common.Entry{{
			Actor: common.ActorHuman,
			Parts: common.PartList{common.TextPart{Text: "hello"}},
		}},
	}
	cfg := common.Config{
		Vendor:  common.VendorAnthropic,
		Model:   "claude-sonnet-5",
		BaseURL: "https://api.anthropic.com",
		APIKey:  "test-key",
		// Thinking unset (zero) → resolves to ThinkingHigh.
	}

	req, err := (anthropicSeam{}).Render(ctx, cfg)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	body, err := common.BodyOf(req)
	if err != nil {
		t.Fatalf("BodyOf: %v", err)
	}

	var wire struct {
		MaxTokens int `json:"max_tokens"`
		Thinking  *struct {
			Type         string `json:"type"`
			Display      string `json:"display"`
			BudgetTokens int    `json:"budget_tokens"`
		} `json:"thinking"`
		OutputConfig *struct {
			Effort string `json:"effort"`
		} `json:"output_config"`
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	// 1. Thinking block must be present.
	if wire.Thinking == nil {
		t.Fatal("thinking block is nil — the request would not trigger thinking")
	}

	// 2. claude-sonnet-5 is adaptive: type must be "adaptive".
	if wire.Thinking.Type != "adaptive" {
		t.Errorf("thinking.type = %q, want %q (claude-sonnet-5 uses adaptive thinking)",
			wire.Thinking.Type, "adaptive")
	}

	// 3. Adaptive models must NOT have budget_tokens.
	if wire.Thinking.BudgetTokens != 0 {
		t.Errorf("thinking.budget_tokens = %d, want 0 (adaptive models don't use budget_tokens)",
			wire.Thinking.BudgetTokens)
	}

	// 3b. display MUST be "summarized", and this is the assertion most worth
	// having. Its absence has no visible symptom: the response still carries
	// a signed thinking block, the thinking tokens are still generated and
	// BILLED, and only the text is missing. A regression here looks exactly
	// like a model that chose not to think, and is discovered on the invoice.
	if wire.Thinking.Display != "summarized" {
		t.Errorf("thinking.display = %q, want %q — without it the thinking block comes back EMPTY and is still billed",
			wire.Thinking.Display, "summarized")
	}

	// 4. output_config.effort must be present for adaptive models.
	if wire.OutputConfig == nil {
		t.Fatal("output_config is nil — adaptive thinking needs effort level")
	}
	if wire.OutputConfig.Effort != "high" {
		t.Errorf("output_config.effort = %q, want %q", wire.OutputConfig.Effort, "high")
	}
}

// TestAnthropicManualThinkingRendered verifies that a non-adaptive model
// uses the manual thinking wire format: type:"enabled" + budget_tokens.
func TestAnthropicManualThinkingRendered(t *testing.T) {
	// Use fake-model which has AdaptiveThinking: false.
	ctx := &common.Context{
		Dialogue: []common.Entry{{
			Actor: common.ActorHuman,
			Parts: common.PartList{common.TextPart{Text: "hello"}},
		}},
	}
	cfg := common.Config{
		Vendor:  common.VendorAnthropic,
		Model:   "fake-model",
		BaseURL: "https://api.anthropic.com",
		APIKey:  "test-key",
		// Thinking unset (zero) → resolves to ThinkingHigh.
	}

	req, err := (anthropicSeam{}).Render(ctx, cfg)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	body, err := common.BodyOf(req)
	if err != nil {
		t.Fatalf("BodyOf: %v", err)
	}

	var wire struct {
		MaxTokens int `json:"max_tokens"`
		Thinking  *struct {
			Type         string `json:"type"`
			BudgetTokens int    `json:"budget_tokens"`
		} `json:"thinking"`
		OutputConfig *json.RawMessage `json:"output_config"`
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	// 1. Thinking block must be present.
	if wire.Thinking == nil {
		t.Fatal("thinking block is nil — the request would not trigger thinking")
	}

	// 2. Manual model: type must be "enabled".
	if wire.Thinking.Type != "enabled" {
		t.Errorf("thinking.type = %q, want %q", wire.Thinking.Type, "enabled")
	}

	// 3. Budget must be >= 1024 (Anthropic minimum).
	if wire.Thinking.BudgetTokens < 1024 {
		t.Errorf("thinking.budget_tokens = %d, want >= 1024", wire.Thinking.BudgetTokens)
	}

	// 4. max_tokens must strictly exceed budget_tokens.
	if wire.MaxTokens <= wire.Thinking.BudgetTokens {
		t.Errorf("max_tokens (%d) must be strictly greater than budget_tokens (%d)",
			wire.MaxTokens, wire.Thinking.BudgetTokens)
	}

	// 5. Manual models must NOT have output_config.
	if wire.OutputConfig != nil {
		t.Errorf("output_config should be absent for manual-thinking models, got: %s", *wire.OutputConfig)
	}
}

// TestAnthropicThinkingOff verifies that setting ThinkingOff produces a
// request with NO thinking block — the opt-out must work too.
func TestAnthropicThinkingOff(t *testing.T) {
	ctx := &common.Context{
		Dialogue: []common.Entry{{
			Actor: common.ActorHuman,
			Parts: common.PartList{common.TextPart{Text: "hello"}},
		}},
	}
	cfg := common.Config{
		Vendor:   common.VendorAnthropic,
		Model:    "claude-sonnet-5",
		BaseURL:  "https://api.anthropic.com",
		APIKey:   "test-key",
		Thinking: common.ThinkingOff,
	}

	req, err := (anthropicSeam{}).Render(ctx, cfg)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	body, err := common.BodyOf(req)
	if err != nil {
		t.Fatalf("BodyOf: %v", err)
	}

	var wire struct {
		Thinking *json.RawMessage `json:"thinking"`
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if wire.Thinking != nil {
		t.Errorf("thinking block should be absent when ThinkingOff, got: %s", *wire.Thinking)
	}
}

// TestGeminiThinkingRendered verifies Gemini gets thinkingConfig with
// includeThoughts: true.
func TestGeminiThinkingRendered(t *testing.T) {
	ctx := &common.Context{
		Dialogue: []common.Entry{{
			Actor: common.ActorHuman,
			Parts: common.PartList{common.TextPart{Text: "hello"}},
		}},
	}
	cfg := common.Config{
		Vendor:  common.VendorGemini,
		Model:   "gemini-3.8-flash",
		BaseURL: "https://generativelanguage.googleapis.com",
		APIKey:  "test-key",
	}

	req, err := (geminiSeam{}).Render(ctx, cfg)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	body, err := common.BodyOf(req)
	if err != nil {
		t.Fatalf("BodyOf: %v", err)
	}

	var wire struct {
		GenerationConfig *struct {
			ThinkingConfig *struct {
				ThinkingBudget  int  `json:"thinkingBudget"`
				IncludeThoughts bool `json:"includeThoughts"`
			} `json:"thinkingConfig"`
			MaxOutputTokens int `json:"maxOutputTokens"`
		} `json:"generationConfig"`
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if wire.GenerationConfig == nil {
		t.Fatal("generationConfig is nil")
	}
	if wire.GenerationConfig.ThinkingConfig == nil {
		t.Fatal("thinkingConfig is nil — Gemini would not include thoughts")
	}
	if !wire.GenerationConfig.ThinkingConfig.IncludeThoughts {
		t.Error("includeThoughts must be true — otherwise you pay for thinking you never receive")
	}
	if wire.GenerationConfig.ThinkingConfig.ThinkingBudget < 1024 {
		t.Errorf("thinkingBudget = %d, want >= 1024", wire.GenerationConfig.ThinkingConfig.ThinkingBudget)
	}
}

// TestOpenAIReasoningEffort verifies OpenAI gets reasoning_effort.
func TestOpenAIReasoningEffort(t *testing.T) {
	ctx := &common.Context{
		Dialogue: []common.Entry{{
			Actor: common.ActorHuman,
			Parts: common.PartList{common.TextPart{Text: "hello"}},
		}},
	}
	cfg := common.Config{
		Vendor:  common.VendorOpenAI,
		Model:   "gpt-6-astra",
		BaseURL: "https://api.openai.com",
		APIKey:  "test-key",
	}

	req, err := (openAISeam{}).Render(ctx, cfg)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	body, err := common.BodyOf(req)
	if err != nil {
		t.Fatalf("BodyOf: %v", err)
	}

	var wire struct {
		ReasoningEffort string `json:"reasoning_effort"`
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if wire.ReasoningEffort != "high" {
		t.Errorf("reasoning_effort = %q, want %q", wire.ReasoningEffort, "high")
	}
}
