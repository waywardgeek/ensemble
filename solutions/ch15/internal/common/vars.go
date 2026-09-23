package common

import (
	"fmt"
	"sort"
	"strings"
)

// VarRenderer produces the current value of a template variable.
// Called at render time: once at agent creation for the initial system prompt,
// and again at load time for dynamically loaded skill bodies.
type VarRenderer func() string

// VarRegistry holds named variable renderers for $VAR substitution.
type VarRegistry struct {
	renderers map[string]VarRenderer
}

// NewVarRegistry creates an empty variable registry.
func NewVarRegistry() *VarRegistry {
	return &VarRegistry{
		renderers: make(map[string]VarRenderer),
	}
}

// Register adds a variable renderer. The name should be uppercase
// (e.g., "TOOLS", "SKILLS", "MEMORY").
func (v *VarRegistry) Register(name string, renderer VarRenderer) {
	v.renderers[name] = renderer
}

// Render performs variable substitution on text. Scans for $IDENTIFIER
// tokens (uppercase letters, digits, underscores). Known variables are
// replaced with their rendered values. Unknown variables are left as-is.
func (v *VarRegistry) Render(text string) string {
	if !strings.Contains(text, "$") {
		return text
	}

	var result strings.Builder
	result.Grow(len(text))

	i := 0
	for i < len(text) {
		if text[i] != '$' {
			result.WriteByte(text[i])
			i++
			continue
		}

		// Found a $. Read the identifier.
		j := i + 1
		for j < len(text) && isVarChar(text[j]) {
			j++
		}

		if j == i+1 {
			// Bare $ with no identifier — keep it
			result.WriteByte('$')
			i++
			continue
		}

		name := text[i+1 : j]
		if renderer, ok := v.renderers[name]; ok {
			result.WriteString(renderer())
		} else {
			// Unknown variable — leave as-is
			result.WriteString(text[i:j])
		}
		i = j
	}

	return result.String()
}

// isVarChar returns true for characters allowed in a variable name:
// uppercase letters, digits, and underscores.
func isVarChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// BuiltinToolsRenderer creates a VarRenderer for $TOOLS that lists
// the currently-enabled tool declarations.
func BuiltinToolsRenderer(skills *SkillRegistry, toolReg interface{ Declarations() []ToolDecl }) VarRenderer {
	return func() string {
		enabled := skills.LoadedTools()
		if len(enabled) == 0 {
			return "(no tools available)"
		}
		set := make(map[string]bool, len(enabled))
		for _, name := range enabled {
			set[name] = true
		}

		var lines []string
		for _, decl := range toolReg.Declarations() {
			if set[decl.Name] {
				lines = append(lines, fmt.Sprintf("- **%s**: %s", decl.Name, decl.Description))
			}
		}
		sort.Strings(lines)
		return strings.Join(lines, "\n")
	}
}

// BuiltinSkillsRenderer creates a VarRenderer for $SKILLS that lists
// the currently-loadable skills (name + description).
func BuiltinSkillsRenderer(skills *SkillRegistry) VarRenderer {
	return func() string {
		loadable := skills.LoadableSkills()
		if len(loadable) == 0 {
			return "(no skills available to load)"
		}
		var lines []string
		for _, s := range loadable {
			lines = append(lines, fmt.Sprintf("- **%s**: %s", s.Name, s.Description))
		}
		return strings.Join(lines, "\n")
	}
}
