package common

// EntryKind identifies what removes an entry. Survivors never contain tool parts.
type EntryKind uint8

const (
	// KindDialogue contains conversation and recoverable tool traffic.
	KindDialogue EntryKind = iota + 1
	// KindHandoff holds a checkpoint note after tool clearing.
	KindHandoff
	// KindSkill holds instructions independently of the loading tool's result.
	KindSkill
	// KindTools holds a declaration change at its original conversation position.
	KindTools
)

// SkillData captures rendered instructions, so replay never rereads a changed file.
type SkillData struct {
	// Name identifies the loaded skill.
	Name string `json:"name"`
	// Body is the rendered instruction text, including loaded dependencies.
	Body string `json:"body"`
}

// MicroHandoffData records the checkpoint text that survives its tool call.
type MicroHandoffData struct {
	// Text explains completed work, failed approaches and the next step.
	Text string `json:"text"`
}

// ToolsChangedData is a delta against the declarations immediately before it.
type ToolsChangedData struct {
	// Added contains full definitions, including replacement definitions by name.
	Added []ToolDeclaration `json:"added,omitempty"`
	// Removed names tools no longer available for dispatch.
	Removed []string `json:"removed,omitempty"`
}
