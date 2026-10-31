// Package skills discovers and renders per-Agent capabilities; it owns no tools.
package skills

import (
	"fmt"
	"strings"
)

// The chapter's frontmatter is deliberately a flat format, not arbitrary YAML.
// Keeping its two list forms here avoids making a general document processor.
type skill struct {
	name, description, kind, body string
	tools, depends, loadable      []string
}

func (r *registry) parse(text string) (skill, error) {
	parts := strings.SplitN(text, "---", 3)
	if len(parts) != 3 || strings.TrimSpace(parts[0]) != "" {
		return skill{}, fmt.Errorf("skill requires --- frontmatter markers")
	}
	fields := make(map[string][]string)
	key := ""
	for _, line := range strings.Split(parts[1], "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			if key != "tools" && key != "depends" && key != "loadable-skills" {
				return skill{}, fmt.Errorf("list item outside a list field")
			}
			fields[key] = append(fields[key], strings.TrimSpace(strings.TrimPrefix(line, "- ")))
			continue
		}
		var value string
		var ok bool
		key, value, ok = strings.Cut(line, ":")
		if !ok {
			return skill{}, fmt.Errorf("invalid frontmatter line %q", line)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch key {
		case "tools", "depends", "loadable-skills":
			fields[key] = strings.Fields(value)
		default:
			fields[key] = []string{strings.Trim(value, "\"'")}
		}
	}
	one := func(key string) string { return strings.Join(fields[key], " ") }
	s := skill{name: one("name"), description: one("description"), kind: one("type"), body: strings.TrimSpace(parts[2]), tools: fields["tools"], depends: fields["depends"], loadable: fields["loadable-skills"]}
	if s.name == "" {
		return s, fmt.Errorf("skill name is required")
	}
	switch s.kind {
	case "primary", "loadable", "dependency":
	default:
		return s, fmt.Errorf("skill %q has invalid type %q", s.name, s.kind)
	}
	return s, nil
}
