// Package skills owns catalog parsing, graph resolution and recorded capability
// material. It reaches configuration and diagnostics through its Agent parent.
package skills

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"example.com/ensemble/internal/common"
)

const sourceLimit = 65536
const catalogLimit = 8388608
const definitionLimit = 256

func (s *Service) failure(code, name, detail string) error {
	if !skillID(name) {
		name = ""
	}
	revision := uint64(0)
	if s.committed != nil {
		revision = s.committed.state.Revision
	}
	s.parent.Ensemble().Logf("skills %s: %s (%s)", name, detail, code)
	return &common.SkillError{Code: code, Name: name, Revision: revision, Detail: detail}
}
func skillID(v string) bool    { return identifier(v, '-', false) }
func toolID(v string) bool     { return identifier(v, '_', false) }
func variableID(v string) bool { return identifier(v, '_', true) }
func identifier(v string, separator byte, upper bool) bool {
	letter := func(b byte) bool {
		if upper {
			return b >= 'A' && b <= 'Z'
		}
		return b >= 'a' && b <= 'z'
	}
	if len(v) == 0 || len(v) > 64 || !letter(v[0]) {
		return false
	}
	for i := 1; i < len(v); i++ {
		b := v[i]
		if !letter(b) && (b < '0' || b > '9') && b != separator {
			return false
		}
	}
	return true
}

// CopyConfiguration runs before publication. It reads no catalog files and
// retains no caller-owned map or byte slice. Resolution never changes cwd.
func CopyConfiguration(owner common.Agent, input *common.SkillConfig) (*common.SkillConfig, error) {
	if input == nil {
		return nil, nil
	}
	bad := func(detail string) (*common.SkillConfig, error) {
		owner.Ensemble().Logf("invalid skill configuration: %s", detail)
		return nil, fmt.Errorf("invalid skill configuration: %s", detail)
	}
	if (input.Directory != "") == (input.Catalog != nil) {
		return bad("select exactly one catalog source")
	}
	if !skillID(input.Primary) {
		return bad("invalid primary identifier")
	}
	out := &common.SkillConfig{Directory: input.Directory, Primary: input.Primary}
	if out.Directory != "" {
		if !filepath.IsAbs(out.Directory) {
			out.Directory = filepath.Join(owner.Workspace(), out.Directory)
		}
		resolved, err := filepath.Abs(out.Directory)
		if err != nil {
			return bad("cannot resolve catalog directory")
		}
		out.Directory = resolved
	}
	if input.Catalog != nil {
		if len(input.Catalog) > definitionLimit {
			return bad("too many skill definitions")
		}
		total := 0
		out.Catalog = make(map[string][]byte, len(input.Catalog))
		for name, data := range input.Catalog {
			if !skillID(name) {
				return bad("invalid catalog identifier")
			}
			if len(data) > sourceLimit || len(data) > catalogLimit-total {
				return bad("skill source byte limit exceeded")
			}
			total += len(data)
			out.Catalog[name] = bytes.Clone(data)
		}
	}
	if len(input.Variables) > 32 {
		return bad("too many scalar bindings")
	}
	out.Variables = make(map[string]string, len(input.Variables))
	total := 0
	for name, value := range input.Variables {
		if !variableID(name) || name == "TOOLS" || name == "SKILLS" {
			return bad("invalid custom variable name")
		}
		if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
			return bad("invalid scalar text")
		}
		if len(value) > 4096 || len(value) > 65536-total {
			return bad("scalar byte limit exceeded")
		}
		total += len(value)
		out.Variables[name] = value
	}
	return out, nil
}

func (s *Service) readCatalog() error {
	config := s.parent.SkillConfiguration()
	if config == nil {
		return s.failure("skills_disabled", "", "skills disabled")
	}
	sources := config.Catalog
	if config.Directory != "" {
		sources = map[string][]byte{}
		entries, err := os.ReadDir(config.Directory)
		if err != nil {
			return s.failure("skill_unavailable", "", "cannot read catalog directory")
		}
		total := 0
		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink != 0 {
				return s.failure("skill_unavailable", entry.Name(), "symlinked catalog child")
			}
			if entry.Type().IsRegular() {
				continue
			}
			if !entry.IsDir() || !skillID(entry.Name()) {
				return s.failure("skill_unavailable", "", "invalid catalog child")
			}
			if len(sources) == definitionLimit {
				return s.failure("skill_too_large", "", "too many skill definitions")
			}
			path := filepath.Join(config.Directory, entry.Name(), "SKILL.md")
			info, err := os.Lstat(path)
			if err != nil || !info.Mode().IsRegular() {
				return s.failure("skill_unavailable", entry.Name(), "skill requires an ordinary readable file")
			}
			if info.Size() > sourceLimit {
				return s.failure("skill_too_large", entry.Name(), "skill source byte limit exceeded")
			}
			file, err := os.Open(path)
			if err != nil {
				return s.failure("skill_unavailable", entry.Name(), "cannot read skill file")
			}
			data, readErr := io.ReadAll(io.LimitReader(file, sourceLimit+1))
			closeErr := file.Close()
			if readErr != nil || closeErr != nil {
				return s.failure("skill_unavailable", entry.Name(), "cannot read skill file")
			}
			if len(data) > sourceLimit || len(data) > catalogLimit-total {
				return s.failure("skill_too_large", entry.Name(), "skill source byte limit exceeded")
			}
			total += len(data)
			sources[entry.Name()] = data
		}
	}
	if len(sources) > definitionLimit {
		return s.failure("skill_too_large", "", "too many skill definitions")
	}
	names := make([]string, 0, len(sources))
	for name := range sources {
		names = append(names, name)
	}
	sort.Strings(names)
	total := 0
	for _, name := range names {
		data := sources[name]
		if len(data) > catalogLimit-total {
			return s.failure("skill_too_large", name, "catalog source byte limit exceeded")
		}
		total += len(data)
		definition, err := s.parse(name, data)
		if err != nil {
			return err
		}
		s.catalog[name] = definition
	}
	return nil
}

func (s *Service) parse(name string, source []byte) (common.SkillDefinition, error) {
	bad := func(detail string) (common.SkillDefinition, error) {
		return common.SkillDefinition{}, s.failure("skill_unavailable", name, detail)
	}
	if !skillID(name) {
		return bad("invalid catalog identifier")
	}
	if len(source) > sourceLimit {
		return common.SkillDefinition{}, s.failure("skill_too_large", name, "skill source byte limit exceeded")
	}
	if !utf8.Valid(source) || bytes.Contains(source, []byte{0}) || bytes.Contains(source, []byte{0xef, 0xbb, 0xbf}) {
		return bad("invalid skill source text")
	}
	for i, b := range source {
		if b == '\r' && (i+1 == len(source) || source[i+1] != '\n') {
			return bad("bare CR in skill source")
		}
	}
	text := string(source)
	// Offsets stay in the original bytes: body CRLF, blank lines and horizontal
	// rules are material, not parser whitespace to normalize away.
	line := func(start int) (string, int) {
		n := strings.IndexByte(text[start:], '\n')
		if n < 0 {
			return text[start:], len(text)
		}
		end := start + n
		return strings.TrimSuffix(text[start:end], "\r"), end + 1
	}
	first, pos := line(0)
	if first != "---" || pos == len(text) {
		return bad("missing frontmatter delimiters")
	}
	headers := []string{}
	body := ""
	closed := false
	for pos < len(text) {
		value, next := line(pos)
		pos = next
		if value == "---" {
			body = text[pos:]
			closed = true
			break
		}
		headers = append(headers, value)
	}
	if !closed {
		return bad("missing closing frontmatter delimiter")
	}
	fields := map[string]string{}
	lists := map[string][]string{}
	for i := 0; i < len(headers); i++ {
		row := headers[i]
		if strings.ContainsRune(row, '\t') {
			return bad("tabs in frontmatter")
		}
		if strings.Trim(row, " ") == "" {
			continue
		}
		key, value, ok := strings.Cut(row, ":")
		if !ok {
			return bad("invalid frontmatter mapping")
		}
		if _, exists := fields[key]; exists {
			return bad("duplicate frontmatter field")
		}
		switch key {
		case "name", "description", "type", "tools", "depends", "loadable-skills":
		default:
			return bad("unknown frontmatter field")
		}
		fields[key] = value
		isList := key == "tools" || key == "depends" || key == "loadable-skills"
		if !isList {
			decoded, quoted, err := s.headerString(value)
			if err != nil {
				return bad("invalid header string")
			}
			if key == "description" && !quoted {
				r, _ := utf8.DecodeRuneInString(decoded)
				lower := strings.ToLower(decoded)
				if !unicode.IsLetter(r) || strings.ContainsRune(decoded, '#') || lower == "true" || lower == "false" || lower == "null" {
					return bad("invalid unquoted description")
				}
			}
			fields[key] = decoded
			continue
		}
		items := []string{}
		value = strings.Trim(value, " ")
		if value == "" {
			for i+1 < len(headers) {
				next := headers[i+1]
				if strings.Trim(next, " ") == "" {
					i++
					continue
				}
				if !strings.HasPrefix(next, "  - ") {
					break
				}
				i++
				item, _, err := s.headerString(next[4:])
				if err != nil {
					return bad("invalid block list item")
				}
				items = append(items, item)
			}
			if len(items) == 0 {
				return bad("empty block list")
			}
		} else if value != "[]" {
			decoded, _, err := s.headerString(value)
			if err != nil {
				return bad("invalid list scalar")
			}
			items = strings.Split(decoded, " ")
		}
		seen := map[string]bool{}
		for _, item := range items {
			valid := skillID(item)
			if key == "tools" {
				valid = toolID(item)
			}
			if !valid || seen[item] {
				return bad("invalid or duplicate list identifier")
			}
			seen[item] = true
		}
		sort.Strings(items)
		lists[key] = items
	}
	if fields["name"] != name {
		return bad("name must equal catalog identifier")
	}
	if fields["description"] == "" || len(fields["description"]) > 256 {
		return bad("invalid description size")
	}
	kind := fields["type"]
	if kind != "primary" && kind != "loadable" && kind != "dependency" {
		return bad("invalid skill type")
	}
	if kind == "primary" && body == "" {
		return bad("primary body is empty")
	}
	return common.SkillDefinition{Name: name, Description: fields["description"], Type: kind, Body: body, Tools: lists["tools"], Depends: lists["depends"], Loadable: lists["loadable-skills"]}, nil
}

// Header syntax is data decoding only. The owning parser supplies diagnostics.
func (s *Service) headerString(input string) (string, bool, error) {
	input = strings.Trim(input, " ")
	bad := func() (string, bool, error) { return "", false, fmt.Errorf("invalid header string") }
	if input == "" || strings.ContainsAny(input, "\t\r\n\x00") {
		return bad()
	}
	quoted := false
	value := input
	if input[0] == '"' {
		quoted = true
		if json.Unmarshal([]byte(input), &value) != nil {
			return bad()
		}
	} else if input[0] == '\'' {
		quoted = true
		var out strings.Builder
		closed := false
		for i := 1; i < len(input); i++ {
			if input[i] != '\'' {
				out.WriteByte(input[i])
				continue
			}
			if i+1 < len(input) && input[i+1] == '\'' {
				out.WriteByte('\'')
				i++
				continue
			}
			if i != len(input)-1 {
				return bad()
			}
			closed = true
		}
		if !closed {
			return bad()
		}
		value = out.String()
	}
	if strings.ContainsAny(value, "\r\n\x00\t") || !utf8.ValidString(value) {
		return bad()
	}
	return value, quoted, nil
}
