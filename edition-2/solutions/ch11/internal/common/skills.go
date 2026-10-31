package common

// AgentSpec supplies per-Agent identity independently of provider configuration.
// Omit it to retain the unrestricted tool setup used before skills.
type AgentSpec struct {
	// SkillsDir contains one named directory per SKILL.md.
	SkillsDir string
	// PrimarySkill selects the identity loaded before the first request.
	PrimarySkill string
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
