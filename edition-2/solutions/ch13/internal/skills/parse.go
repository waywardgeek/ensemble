// Package skills discovers and renders per-Agent capabilities; it owns no tools.
package skills

import (
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"strings"
)

// Parse only the chapter's frontmatter fields and MCP server list, not arbitrary YAML.
type skill struct {
	name, description, kind, body string
	tools, depends, loadable      []string
	servers                       []common.MCPServer
}

func (r *registry) parse(text string) (skill, error) {
	parts := strings.SplitN(text, "---", 3)
	if len(parts) != 3 || strings.TrimSpace(parts[0]) != "" {
		return skill{}, fmt.Errorf("skill requires --- frontmatter markers")
	}
	fields := make(map[string][]string)
	key := ""
	var servers []common.MCPServer
	inServers := false
	for _, line := range strings.Split(parts[1], "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "mcp_servers:" {
			inServers = true
			continue
		}
		if inServers && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
				if strings.HasPrefix(trimmed, "- ") {
					servers = append(servers, common.MCPServer{})
					trimmed = strings.TrimPrefix(trimmed, "- ")
				}
				if len(servers) == 0 {
					return skill{}, fmt.Errorf("MCP server list requires - name")
				}
				field, value, ok := strings.Cut(trimmed, ":")
				if !ok {
					return skill{}, fmt.Errorf("invalid MCP server field")
				}
				value = strings.TrimSpace(value)
				server := &servers[len(servers)-1]
				switch field {
				case "name":
					server.Name = strings.Trim(value, "\"'")
				case "transport":
					server.Transport = strings.Trim(value, "\"'")
				case "command":
					server.Command = strings.Trim(value, "\"'")
				case "args":
					if err := json.Unmarshal([]byte(value), &server.Args); err != nil {
						return skill{}, fmt.Errorf("MCP args must be a JSON-style string array: %w", err)
					}
				default:
					return skill{}, fmt.Errorf("unknown MCP server field %q", field)
				}
				continue
			}
			inServers = false
		}
		line = trimmed
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
	s := skill{servers: servers, name: one("name"), description: one("description"), kind: one("type"), body: strings.TrimSpace(parts[2]), tools: fields["tools"], depends: fields["depends"], loadable: fields["loadable-skills"]}
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
