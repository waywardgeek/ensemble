package common

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SkillEntry is a skill in the registry with its runtime load state.
type SkillEntry struct {
	Props    *SkillProperties
	State    LoadState
	EventSeq int // 0 for initial/available; event seq for dynamic loads
}

// SkillRegistry manages all known skills and their load states.
type SkillRegistry struct {
	skills map[string]*SkillEntry
}

// NewSkillRegistry creates an empty registry.
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]*SkillEntry),
	}
}

// DiscoverSkills scans a directory for subdirectories containing SKILL.md.
// Each valid skill is registered as Available (parsed but not loaded).
func (r *SkillRegistry) DiscoverSkills(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No skills directory is fine
		}
		return fmt.Errorf("reading skills dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillDir := filepath.Join(dir, entry.Name())
		skillPath := filepath.Join(skillDir, "SKILL.md")
		content, err := os.ReadFile(skillPath)
		if err != nil {
			continue // Not a skill directory
		}
		props, err := ParseSkillMD(string(content))
		if err != nil {
			return fmt.Errorf("parsing %s: %w", skillPath, err)
		}
		props.FilePath = skillPath
		r.skills[props.Name] = &SkillEntry{
			Props: props,
			State: LoadAvailable,
		}
	}
	return nil
}

// Register adds a skill to the registry. Used for programmatic registration
// (e.g., in tests or for embedded skills).
func (r *SkillRegistry) Register(props *SkillProperties) {
	r.skills[props.Name] = &SkillEntry{
		Props: props,
		State: LoadAvailable,
	}
}

// Get returns a skill entry by name, or nil if not found.
func (r *SkillRegistry) Get(name string) *SkillEntry {
	return r.skills[name]
}

// LoadInitial loads a skill and marks it as initial (loaded at creation).
// Resolves depends: chain. Panics on circular dependencies.
func (r *SkillRegistry) LoadInitial(name string, vars *VarRegistry) error {
	return r.load(name, LoadInitial, 0, vars, nil)
}

// LoadDynamic loads a skill and marks it as dynamic (loaded via load_skill).
// Resolves depends: chain. Returns the rendered body.
func (r *SkillRegistry) LoadDynamic(name string, eventSeq int, vars *VarRegistry) (string, error) {
	// Check the skill is loadable
	entry := r.skills[name]
	if entry == nil {
		return "", fmt.Errorf("unknown skill %q", name)
	}
	if entry.Props.Type == SkillPrimary {
		return "", fmt.Errorf("skill %q is primary and cannot be loaded dynamically", name)
	}
	if entry.State == LoadDynamic || entry.State == LoadInitial {
		return "", fmt.Errorf("skill %q is already loaded", name)
	}

	// Check it's in the loadable set
	if !r.IsLoadable(name) {
		return "", fmt.Errorf("skill %q is not available for loading (not declared by any loaded skill's loadable-skills)", name)
	}

	if err := r.load(name, LoadDynamic, eventSeq, vars, nil); err != nil {
		return "", err
	}
	return entry.Props.Body, nil
}

// load is the internal loader. visiting tracks the dependency chain for
// cycle detection.
func (r *SkillRegistry) load(name string, state LoadState, eventSeq int, vars *VarRegistry, visiting map[string]bool) error {
	if visiting == nil {
		visiting = make(map[string]bool)
	}
	if visiting[name] {
		// Build the cycle path for the error message
		var path []string
		for n := range visiting {
			path = append(path, n)
		}
		sort.Strings(path)
		panic(fmt.Sprintf("circular skill dependency: %s -> %s", strings.Join(path, " -> "), name))
	}

	entry := r.skills[name]
	if entry == nil {
		return fmt.Errorf("unknown skill %q", name)
	}

	// Already loaded — skip (dependency pulled in by multiple paths)
	if entry.State == LoadInitial || entry.State == LoadDynamic {
		return nil
	}

	visiting[name] = true

	// Load dependencies first
	for _, dep := range entry.Props.Dependencies {
		if err := r.load(dep, state, eventSeq, vars, visiting); err != nil {
			return fmt.Errorf("loading dependency %q of %q: %w", dep, name, err)
		}
	}

	delete(visiting, name)

	// Render variable substitutions in the body
	if vars != nil {
		entry.Props.Body = vars.Render(entry.Props.RawBody)
	}

	entry.State = state
	entry.EventSeq = eventSeq
	return nil
}

// MarkUnload marks a skill for removal at the next context compaction.
func (r *SkillRegistry) MarkUnload(name string) error {
	entry := r.skills[name]
	if entry == nil {
		return fmt.Errorf("unknown skill %q", name)
	}
	if entry.State != LoadDynamic && entry.State != LoadInitial {
		return fmt.Errorf("skill %q is not loaded", name)
	}
	entry.State = LoadPendingUnload
	return nil
}

// IsLoadable returns true if the named skill can be loaded via load_skill.
// A skill is loadable if: (a) it exists, (b) its type is loadable, and
// (c) it appears in the loadable-skills closure of some currently-loaded skill.
func (r *SkillRegistry) IsLoadable(name string) bool {
	entry := r.skills[name]
	if entry == nil {
		return false
	}
	if entry.Props.Type != SkillLoadable {
		return false
	}
	// Check if any loaded skill declares this in its loadable-skills
	for _, e := range r.skills {
		if e.State != LoadInitial && e.State != LoadDynamic {
			continue
		}
		for _, ls := range e.Props.LoadableSkills {
			if ls == name {
				return true
			}
		}
	}
	return false
}

// LoadableSkills returns the names and descriptions of skills currently
// available for load_skill, sorted by name.
func (r *SkillRegistry) LoadableSkills() []SkillSummary {
	// First, collect all names declared loadable by loaded skills
	available := make(map[string]bool)
	for _, e := range r.skills {
		if e.State != LoadInitial && e.State != LoadDynamic {
			continue
		}
		for _, ls := range e.Props.LoadableSkills {
			available[ls] = true
		}
	}

	var result []SkillSummary
	for name := range available {
		entry := r.skills[name]
		if entry == nil {
			continue
		}
		// Only show skills that are loadable type and not already loaded
		if entry.Props.Type != SkillLoadable {
			continue
		}
		if entry.State == LoadInitial || entry.State == LoadDynamic {
			continue // Already loaded
		}
		result = append(result, SkillSummary{
			Name:        entry.Props.Name,
			Description: entry.Props.Description,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// LoadedTools returns the tool names enabled by all currently-loaded skills,
// deduplicated and sorted.
func (r *SkillRegistry) LoadedTools() []string {
	seen := make(map[string]bool)
	for _, e := range r.skills {
		if e.State != LoadInitial && e.State != LoadDynamic {
			continue
		}
		for _, t := range e.Props.Tools {
			seen[t] = true
		}
	}
	var result []string
	for t := range seen {
		result = append(result, t)
	}
	sort.Strings(result)
	return result
}

// IsToolEnabled returns true if the named tool is enabled by any loaded skill,
// or if it's a framework tool (load_skill, unload_skill).
func (r *SkillRegistry) IsToolEnabled(name string) bool {
	// Framework tools are always enabled when skills are active.
	if name == "load_skill" || name == "unload_skill" {
		return true
	}
	for _, e := range r.skills {
		if e.State != LoadInitial && e.State != LoadDynamic {
			continue
		}
		for _, t := range e.Props.Tools {
			if t == name {
				return true
			}
		}
	}
	return false
}

// SkillSummary is a lightweight representation for listing.
type SkillSummary struct {
	Name        string
	Description string
}

// InitialBodies returns the rendered bodies of all initially-loaded skills,
// in alphabetical order by name. Used to build the system prompt at creation.
func (r *SkillRegistry) InitialBodies() []string {
	var entries []*SkillEntry
	for _, e := range r.skills {
		if e.State == LoadInitial {
			entries = append(entries, e)
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Props.Name < entries[j].Props.Name
	})
	var bodies []string
	for _, e := range entries {
		if e.Props.Body != "" {
			bodies = append(bodies, e.Props.Body)
		}
	}
	return bodies
}
