package skills

import (
	"bufio"
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Frontmatter deliberately accepts scalar strings, folded/literal description
// blocks and string lists (inline or indented). It is not a general YAML engine.
func parseSkill(path string) (common.Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return common.Skill{}, err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	s := common.Skill{Source: path, Instructions: text}
	if !strings.HasPrefix(text, "---\n") {
		return s, fmt.Errorf("%s: missing frontmatter", path)
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return s, fmt.Errorf("%s: unclosed frontmatter", path)
	}
	header := text[4 : 4+end]
	s.Instructions = text[4+end+5:]
	var key, block string
	lists := map[string]*[]string{"tools": &s.Tools, "depends": &s.Depends, "loadable-skills": &s.Offers}
	scalar := func(v string) string {
		v = strings.TrimSpace(v)
		if len(v) > 1 && v[0] == '"' {
			var q string
			if json.Unmarshal([]byte(v), &q) == nil {
				return q
			}
		}
		if len(v) > 1 && v[0] == '\'' && v[len(v)-1] == '\'' {
			return strings.ReplaceAll(v[1:len(v)-1], "''", "'")
		}
		return v
	}
	scan := bufio.NewScanner(strings.NewReader(header))
	scan.Buffer(make([]byte, 4096), 1024*1024)
	for scan.Scan() {
		line := scan.Text()
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if key == "mcp_servers" {
				if strings.HasPrefix(trim, "- ") {
					s.MCPServers = append(s.MCPServers, common.MCPServer{})
					trim = strings.TrimSpace(trim[2:])
				}
				if len(s.MCPServers) == 0 {
					return s, fmt.Errorf("%s: mcp_servers requires entries", path)
				}
				k, v, ok := strings.Cut(trim, ":")
				if !ok {
					return s, fmt.Errorf("%s: invalid MCP declaration", path)
				}
				d := &s.MCPServers[len(s.MCPServers)-1]
				v = scalar(v)
				switch k {
				case "name":
					d.Name = v
				case "transport":
					d.Transport = v
				case "url":
					d.URL = v
				case "auth-env":
					d.AuthEnv = v
				case "command":
					d.Command = v
				case "address":
					d.Address = v
				case "args":
					if err := json.Unmarshal([]byte(strings.TrimSpace(strings.SplitN(trim, ":", 2)[1])), &d.Args); err != nil {
						return s, fmt.Errorf("%s: MCP args must be a JSON string array", path)
					}
				default:
					return s, fmt.Errorf("%s: unknown MCP field %q", path, k)
				}
				continue
			}
			if block != "" {
				sep := "\n"
				if block == ">" {
					sep = " "
				}
				if s.Description != "" {
					s.Description += sep
				}
				s.Description += trim
				continue
			}
			if list := lists[key]; list != nil && strings.HasPrefix(trim, "- ") {
				*list = append(*list, scalar(trim[2:]))
				continue
			}
			return s, fmt.Errorf("%s: unsupported frontmatter indentation", path)
		}
		k, v, ok := strings.Cut(trim, ":")
		if !ok {
			return s, fmt.Errorf("%s: invalid frontmatter", path)
		}
		key = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		block = ""
		if list := lists[key]; list != nil {
			if v == "" {
				continue
			}
			if !strings.HasPrefix(v, "[") {
				*list = append(*list, strings.Fields(v)...)
				continue
			}
			if !strings.HasSuffix(v, "]") {
				return s, fmt.Errorf("%s: %s has unclosed list", path, key)
			}
			body := strings.TrimSpace(v[1 : len(v)-1])
			if body != "" {
				for _, x := range strings.Split(body, ",") {
					*list = append(*list, scalar(x))
				}
			}
			continue
		}
		switch key {
		case "name":
			s.Name = scalar(v)
		case "type":
			s.Type = scalar(v)
		case "description":
			if v == "|" || v == ">" {
				block = v
			} else {
				s.Description = scalar(v)
			}
		default: /* metadata unrelated to this runtime is ignored */
		}
	}
	if err := scan.Err(); err != nil {
		return s, err
	}
	if strings.TrimSpace(s.Name) == "" {
		return s, fmt.Errorf("%s: empty skill name", path)
	}
	if s.Type != "primary" && s.Type != "loadable" && s.Type != "dependency" {
		return s, fmt.Errorf("%s: invalid skill type %q", path, s.Type)
	}
	for _, d := range s.MCPServers {
		if d.Transport == "url" {
			if err := common.ValidateMCPURL(d.URL); err != nil {
				return s, err
			}
		}
		if d.Name == "" || d.Transport == "" {
			return s, fmt.Errorf("%s: MCP declaration requires name and transport", path)
		}
	}
	for _, list := range lists {
		for _, v := range *list {
			if strings.TrimSpace(v) == "" {
				return s, fmt.Errorf("%s: empty list member", path)
			}
		}
	}
	return s, nil
}

// Discovery grants no capabilities. Identity is explicit in frontmatter, not
// inferred from a directory name; duplicates are errors.
// The application-selected corpus is a separate trusted input, not workspace
// data. Host reads here must not become a runtime file-access bypass: later
// loads resolve only these retained declarations by name, never caller paths.
func (r *registry) discover(root string) error {
	if !filepath.IsAbs(root) {
		root = filepath.Join(r.parent.Config().BaseDir, root)
	}
	files, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		path := filepath.Join(root, f.Name(), "SKILL.md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		skill, err := parseSkill(path)
		if err != nil {
			return err
		}
		if r.skills[skill.Name] != nil {
			return fmt.Errorf("%s: duplicate skill %q", path, skill.Name)
		}
		r.skills[skill.Name] = &skillState{info: common.SkillInfo{Skill: skill}}
	}
	return nil
}
