package common

import "encoding/json"

// AgentSpec supplies per-Agent identity independently of provider configuration.
// The composition root chooses full-access defaults when the caller omits a spec.
type AgentSpec struct {
	// Compaction selects a fresh, single-task worker with immutable submit-only admission.
	// Ordinary child spawning resets this purpose instead of inheriting it.
	Compaction bool
	// ID is this Agent's address within its parent's children.
	ID string
	// ParentID names the spawning Agent or the configured human address.
	ParentID string
	// Model optionally requests a model from the application child factory.
	Model string
	// SubmitSchema describes the entire submit argument object; absent uses result text.
	// Construction owns a copy; this contract is not inherited by ordinary children.
	SubmitSchema json.RawMessage
	// DataDir optionally selects this Agent's state directory.
	DataDir string
	// LogPath optionally selects its history log.
	LogPath string
	// SavePath optionally selects its snapshot.
	SavePath string
	// SandboxRoot confines file tools and command descendants when nonempty.
	SandboxRoot string
	// SafeMode removes shell execution and its attendant tools.
	SafeMode bool
	// EnableWebSearch permits discovery and loading of the web-search skill.
	EnableWebSearch bool
	// SkillsDir contains one named directory per SKILL.md.
	SkillsDir string
	// PrimarySkill selects the identity loaded before the first request.
	PrimarySkill string
	// StartupSkills loads reachable optional skills during construction. Their
	// manuals join the frozen startup prompt; no model-authored call is invented.
	StartupSkills []string
	// Variables supplies application renderers before the primary body is frozen.
	Variables map[string]func() string
}

// ToolSource distinguishes creation-time capabilities from later additions.
type ToolSource string

const (
	// SourceInitial marks tools enabled before any conversation events.
	SourceInitial ToolSource = "initial"
	// SourceDynamic marks tools enabled by a recorded load_skill call.
	SourceDynamic ToolSource = "dynamic"
)

// ToolMeta retains the first enabling source, even if another skill shares a tool.
type ToolMeta struct {
	// Source identifies creation-time versus dynamically loaded tools.
	Source ToolSource
	// EventSeq is the enabling tool_called sequence, or zero for initial tools.
	EventSeq uint64
}

// Skills owns one Agent's discovered capabilities and point-of-use renderers.
// Compiled instruction removal is pending; MCP servers stop on unload.
type Skills interface {
	// Directory identifies this Agent's skill corpus for passive discovery.
	Directory() string
	// Agent returns the owner whose declarations and services the registry uses.
	Agent() Agent
	// LoadInitial enables the primary and dependencies and renders their bodies once.
	LoadInitial(string) (string, error)
	// LoadDynamic enables a reachable loadable skill and returns rendered instructions.
	LoadDynamic(string, uint64) (string, error)
	// MarkUnload stops this skill's MCP servers and defers compiled/body removal.
	MarkUnload(string) error
	// PendingUnload returns sorted skill names awaiting future compaction.
	PendingUnload() []string
	// LoadableSkills returns sorted reachable, unloaded loadable skill names.
	LoadableSkills() []string
	// ToolMeta reports whether a tool is enabled and the event that first enabled it.
	ToolMeta(string) (ToolMeta, bool)
	// RegisterVar installs a renderer; already-rendered instructions never change.
	RegisterVar(string, func() string)
}
