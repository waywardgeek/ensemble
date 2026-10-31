package common

// SaveFile keeps the projection and its anchor beside the original facts.
// A non-nil Context is sufficient even when the log has been trimmed.
type SaveFile struct {
	// Config records request identity, never transport credentials or endpoints.
	Config SaveConfig `json:"config"`
	// AsOf is the last event folded into Context, not the length of Log.
	AsOf uint64 `json:"as_of"`
	// Context is the neutral snapshot; nil requests a complete log rebuild.
	Context *Context `json:"context"`
	// Log retains recorded facts in strictly increasing sequence order.
	Log []Event `json:"log"`
}

// SaveConfig is historical metadata; loading never replaces runtime settings.
type SaveConfig struct {
	// Model records the configured request model.
	Model string `json:"model"`
	// Vendor records the request translator.
	Vendor Vendor `json:"vendor"`
	// SystemPrompt records the Agent's current constitution.
	SystemPrompt string `json:"system_prompt"`
	// Tools records declarations, excluding executable handlers.
	Tools []ToolDeclaration `json:"tools"`
}
