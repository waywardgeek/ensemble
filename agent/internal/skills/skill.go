// Package skills parses SKILL.md files and resolves the skill dependency
// graph. It is a spoke: it depends on the shared vocabulary in internal/common
// and on nothing else in the tree.
package skills

import (
	"fmt"
	"strings"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// ParseSkillType converts a frontmatter type value into a common.SkillType.
// An empty value defaults to loadable.
func ParseSkillType(raw string) (common.SkillType, error) {
	switch t := common.SkillType(strings.TrimSpace(raw)); t {
	case "":
		return common.SkillLoadable, nil
	case common.SkillLoadable, common.SkillPrimary, common.SkillDependency:
		return t, nil
	default:
		return "", fmt.Errorf("unknown skill type %q (want %q, %q, or %q)",
			raw, common.SkillLoadable, common.SkillPrimary, common.SkillDependency)
	}
}

// ParseSkillMD splits a SKILL.md into frontmatter fields and body.
// The --- split uses limit 3 so horizontal rules in the body are preserved.
func ParseSkillMD(content string) (*common.SkillProperties, error) {
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("SKILL.md must start with YAML frontmatter (---)")
	}
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("SKILL.md frontmatter not properly closed with ---")
	}

	frontmatter := strings.TrimSpace(parts[1])
	body := strings.TrimSpace(parts[2])

	props := &common.SkillProperties{
		RawBody: body,
		Body:    body,
	}

	// Parse frontmatter line by line (simple YAML subset: key: value).
	// Handles both inline lists ("a b c") and indented lists ("- a\n- b").
	// Also handles mcp_servers as nested object lists.
	var currentKey string
	var listItems []string
	inList := false
	var inMCPServers bool
	var currentMCP *common.MCPServerConfig

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

	flushMCP := func() {
		if currentMCP != nil && currentMCP.Name != "" {
			props.MCPServers = append(props.MCPServers, *currentMCP)
		}
		currentMCP = nil
	}

	for _, line := range strings.Split(frontmatter, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Inside mcp_servers block: parse nested objects.
		if inMCPServers {
			// "- name: value" starts a new server entry.
			if strings.HasPrefix(trimmed, "- ") {
				flushMCP()
				rest := strings.TrimPrefix(trimmed, "- ")
				currentMCP = &common.MCPServerConfig{}
				parseMCPField(currentMCP, rest)
				continue
			}
			// Indented key: value adds to current entry.
			if (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) && currentMCP != nil {
				parseMCPField(currentMCP, trimmed)
				continue
			}
			// Non-indented line: end of mcp_servers block.
			flushMCP()
			inMCPServers = false
			// Fall through to normal parsing.
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
		case "mcp_servers":
			inMCPServers = true
		}
	}
	flushList()
	flushMCP()

	if props.Name == "" {
		return nil, fmt.Errorf("SKILL.md missing required field: name")
	}
	if props.Description == "" {
		return nil, fmt.Errorf("SKILL.md missing required field: description")
	}
	if props.Type == "" {
		props.Type = common.SkillLoadable
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

// parseMCPField parses a "key: value" line into an common.MCPServerConfig.
func parseMCPField(c *common.MCPServerConfig, line string) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return
	}
	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])

	switch key {
	case "name":
		c.Name = value
	case "transport":
		c.Transport = value
	case "command":
		c.Command = value
	case "url":
		c.URL = value
	case "args":
		// Parse ["arg1", "arg2"] or bare words.
		value = strings.TrimPrefix(value, "[")
		value = strings.TrimSuffix(value, "]")
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			item = strings.Trim(item, `"'`)
			if item != "" {
				c.Args = append(c.Args, item)
			}
		}
	case "env":
		value = strings.TrimPrefix(value, "[")
		value = strings.TrimSuffix(value, "]")
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			item = strings.Trim(item, `"'`)
			if item != "" {
				c.Env = append(c.Env, item)
			}
		}
	}
}
