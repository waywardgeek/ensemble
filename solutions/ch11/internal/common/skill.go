package common

import (
	"fmt"
	"strings"
)

// SkillType classifies how a skill is used.
type SkillType string

const (
	SkillLoadable   SkillType = "loadable"   // Dynamic, offered in system prompt.
	SkillPrimary    SkillType = "primary"     // Agent identity, set at creation.
	SkillDependency SkillType = "dependency"  // Pulled in via depends:, invisible.
)

// ParseSkillType converts a frontmatter value. Empty defaults to loadable.
func ParseSkillType(raw string) (SkillType, error) {
	switch t := SkillType(strings.TrimSpace(raw)); t {
	case "", SkillLoadable:
		return SkillLoadable, nil
	case SkillPrimary, SkillDependency:
		return t, nil
	default:
		return "", fmt.Errorf("unknown skill type %q (want %q, %q or %q)",
			raw, SkillPrimary, SkillDependency, SkillLoadable)
	}
}

// LoadState tracks a skill's runtime lifecycle.
type LoadState string

const (
	LoadInitial       LoadState = "initial"        // Loaded at agent creation.
	LoadDynamic       LoadState = "dynamic"        // Loaded via load_skill.
	LoadPendingUnload LoadState = "pending-unload"  // Marked for removal at compaction.
	LoadAvailable     LoadState = "available"       // Parsed but not loaded.
)

// SkillProperties is the parsed representation of a SKILL.md file.
type SkillProperties struct {
	Name           string    // kebab-case identifier
	Description    string    // what the skill does
	Type           SkillType // primary, loadable, or dependency
	Tools          []string  // tool names this skill enables
	Dependencies   []string  // skills auto-loaded when this loads
	LoadableSkills []string  // skills this skill makes available for load_skill
	Body           string    // markdown instructions (post variable substitution)
	RawBody        string    // markdown instructions (pre variable substitution)
	FilePath       string    // path to SKILL.md on disk
}

// ParseSkillMD splits a SKILL.md into frontmatter fields and body.
// The --- split uses limit 3 so horizontal rules in the body are preserved.
func ParseSkillMD(content string) (*SkillProperties, error) {
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("SKILL.md must start with YAML frontmatter (---)")
	}
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("SKILL.md frontmatter not properly closed with ---")
	}

	frontmatter := strings.TrimSpace(parts[1])
	body := strings.TrimSpace(parts[2])

	props := &SkillProperties{
		RawBody: body,
		Body:    body,
	}

	// Parse frontmatter line by line (simple YAML subset: key: value).
	// Handles both inline lists ("a b c") and indented lists ("- a\n- b").
	var currentKey string
	var listItems []string
	inList := false

	flushList := func() {
		if !inList || currentKey == "" {
			return
		}
		switch currentKey {
		case "tools":
			props.Tools = listItems
		case "depends":
			props.Dependencies = listItems
		case "loadable-skills":
			props.LoadableSkills = listItems
		}
		listItems = nil
		inList = false
		currentKey = ""
	}

	for _, line := range strings.Split(frontmatter, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for list item: "  - value"
		if strings.HasPrefix(trimmed, "- ") {
			item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if item != "" {
				listItems = append(listItems, item)
				inList = true
			}
			continue
		}

		// Flush any pending list before processing a new key
		flushList()

		// Key: value line
		idx := strings.Index(trimmed, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		value := strings.TrimSpace(trimmed[idx+1:])

		switch key {
		case "name":
			props.Name = value
		case "description":
			props.Description = value
		case "type":
			t, err := ParseSkillType(value)
			if err != nil {
				return nil, err
			}
			props.Type = t
		case "tools", "depends", "loadable-skills":
			currentKey = key
			if value != "" {
				// Inline: "tools: read_file write_file"
				items := strings.Fields(value)
				switch key {
				case "tools":
					props.Tools = items
				case "depends":
					props.Dependencies = items
				case "loadable-skills":
					props.LoadableSkills = items
				}
			} else {
				// Value is empty — expect indented list items to follow
				inList = true
				listItems = nil
			}
		}
	}
	flushList()

	if props.Name == "" {
		return nil, fmt.Errorf("SKILL.md missing required field: name")
	}
	if props.Description == "" {
		return nil, fmt.Errorf("SKILL.md missing required field: description")
	}
	if props.Type == "" {
		props.Type = SkillLoadable
	}

	return props, nil
}

// flexibleList parses a value that may be space-separated or a YAML list.
// This is used by the registry when loading from frontmatter.
func FlexibleList(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Fields(s)
}
