package common

// Skill vocabulary. These are the shared nouns of the skill system: the parsed
// shape of a SKILL.md and the runtime state of a loaded skill. The behavior that
// produces and manipulates them - the frontmatter parser, the dependency
// resolver, the variable renderer - lives in the internal/skills spoke.
//
// The types live here, and not in that spoke, because the interfaces below
// mention them. A hub that names a spoke's types would have to import the
// spoke, which is the one thing the star topology forbids.

// SkillType classifies how a skill is used.
type SkillType string

const (
	SkillLoadable   SkillType = "loadable"   // Dynamic, offered in system prompt.
	SkillPrimary    SkillType = "primary"    // Agent identity, set at creation.
	SkillDependency SkillType = "dependency" // Pulled in via depends:, invisible.
)

// LoadState tracks a skill's runtime lifecycle.
type LoadState string

const (
	LoadInitial       LoadState = "initial"        // Loaded at agent creation.
	LoadDynamic       LoadState = "dynamic"        // Loaded via load_skill.
	LoadPendingUnload LoadState = "pending-unload" // Marked for removal at compaction.
	LoadAvailable     LoadState = "available"      // Parsed but not loaded.
)

// MCPServerConfig describes an MCP server declared in a skill's frontmatter.
type MCPServerConfig struct {
	Name      string   // server name (for logging/identification)
	Transport string   // "stdio", "websocket", or "url"
	Command   string   // for stdio: the command to run
	Args      []string // for stdio: command arguments
	Env       []string // for stdio: environment variables
	URL       string   // for url: the server URL (future)
}

// SkillProperties is the parsed representation of a SKILL.md file.
type SkillProperties struct {
	Name           string            // kebab-case identifier
	Description    string            // what the skill does
	Type           SkillType         // primary, loadable, or dependency
	Tools          []string          // tool names this skill enables
	Dependencies   []string          // skills auto-loaded when this loads
	LoadableSkills []string          // skills this skill makes available for load_skill
	MCPServers     []MCPServerConfig // MCP servers to start when skill loads
	Body           string            // markdown instructions (post variable substitution)
	RawBody        string            // markdown instructions (pre variable substitution)
	FilePath       string            // path to SKILL.md on disk
}

// SkillEntry is a registered skill plus its runtime state.
type SkillEntry struct {
	Props    *SkillProperties
	State    LoadState
	EventSeq int // sequence number of the event that loaded it
}

// SkillSummary is the compact form offered to the model for load_skill.
type SkillSummary struct {
	Name        string
	Description string
}

// Vars renders $VAR references in skill text. The tool layer holds one so it
// can expand a skill body at load time; internal/skills implements it.
type Vars interface {
	Render(text string) string
}

// Skills is the tool layer's view of the skill system: enough to gate tools by
// the currently loaded set, and to drive load_skill and unload_skill. It is
// deliberately the whole seam - internal/tools calls nothing else on the
// registry, so anything wider would overstate the coupling.
type Skills interface {
	// IsToolEnabled reports whether a tool is exposed by a loaded skill.
	IsToolEnabled(name string) bool
	// Get returns a registered skill, or nil if there is no such skill.
	Get(name string) *SkillEntry
	// LoadDynamic loads a skill and its dependencies, returning the rendered body.
	LoadDynamic(name string, eventSeq int, vars Vars) (string, error)
	// MarkUnload flags a loaded skill for removal at the next compaction.
	MarkUnload(name string) error
	// LoadableSkills lists the skills currently offered for load_skill.
	LoadableSkills() []SkillSummary
}
