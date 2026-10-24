package common

import "encoding/json"

// LogMode controls archive retention, not provider requests or live diagnostics.
type LogMode string

const (
	// LogFull retains causal events alongside the recovery context.
	LogFull LogMode = "full"
	// LogDialogueOnly omits the event archive but retains the complete context.
	LogDialogueOnly LogMode = "dialogue_only"
)

// SavedConfig describes the effective Agent at save time. It is never restored
// over caller configuration and deliberately has no credential or endpoint field.
type SavedConfig struct {
	AuthMode               AuthMode          `json:"auth_mode,omitempty"`
	ReasoningEffort        string            `json:"reasoning_effort,omitempty"`
	ReasoningSummary       string            `json:"reasoning_summary,omitempty"`
	Vendor                 Vendor            `json:"vendor"`
	Surface                Surface           `json:"surface"`
	Model                  string            `json:"model"`
	ResolvedModel          string            `json:"resolved_model,omitempty"`
	SystemPrompt           string            `json:"system_prompt"`
	ContextTarget          int               `json:"context_target,omitempty"`
	MaxToolRounds          int               `json:"max_tool_rounds"`
	DisableStreaming       bool              `json:"disable_streaming"`
	StreamCapability       StreamCapability  `json:"stream_capability"`
	RequireToolExplanation bool              `json:"require_tool_explanation"`
	Tools                  []ToolDeclaration `json:"tools"`
	Skills                 []SkillInfo       `json:"skills"`
}

// SaveFile is the versioned, coherent History archive. AsOf is the context
// checkpoint; only Log events strictly after it are replayed on load.
type SaveFile struct {
	Config      SavedConfig      `json:"config"`
	LogMode     LogMode          `json:"log_mode"`
	AsOf        Seq              `json:"as_of"`
	Context     *Context         `json:"context"`
	Log         []Event          `json:"log"`
	DialogueLog []DialogueRecord `json:"dialogue_log,omitempty"`
}

// Recovery names the validated artifact and resulting sequence boundary.
// Empty Path means both candidates were absent and a fresh history was used.
type Recovery struct {
	Path   string
	AsOf   Seq
	Backup bool
}

// ToolDeclaration describes capability, never an executable handler.
type ToolDeclaration struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
}

// DialogueRecord is a derived archive, not an event or replay input.
type DialogueRecord struct {
	SourceSeq Seq    `json:"source_seq"`
	Role      string `json:"role"`
	Text      string `json:"text"`
}
