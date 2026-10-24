package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Child preserves the public Config owner. Call Clamp after applying child
// overrides. Task identity and submission contracts do not inherit.
func (c Config) Child(dataDir string) Config {
	child := c
	child.ID, child.ParentID, child.Prompt = "", "", ""
	child.SubmitSchema = nil
	child.DataDir, child.LogPath, child.SavePath = dataDir, "", ""
	child.Tools = append([]string(nil), c.Tools...)
	if c.Tools != nil && child.Tools == nil {
		child.Tools = []string{}
	}
	child.StartupSkills = append([]string(nil), c.StartupSkills...)
	child.MCPServers = append([]MCPServer(nil), c.MCPServers...)
	for i := range child.MCPServers {
		child.MCPServers[i].Args = append([]string(nil), c.MCPServers[i].Args...)
	}
	if c.Skills != nil {
		s := *c.Skills
		s.Variables = map[string]func() string{}
		for k, v := range c.Skills.Variables {
			s.Variables[k] = v
		}
		child.Skills = &s
	}
	return child
}

// Clamp always returns the narrowed policy, even when the error reports an
// attempted widening. The alarm must not be confused with the security wall.
func (child Config) Clamp(parent Config) (Config, error) {
	var widened []string
	if parent.SafeMode && !child.SafeMode {
		child.SafeMode = true
		widened = append(widened, "safe mode")
	}
	if !parent.EnableWebSearch && child.EnableWebSearch {
		child.EnableWebSearch = false
		widened = append(widened, "web search")
	}
	if parent.SandboxRoot != "" {
		root, err := CanonicalPath(parent.SandboxRoot)
		if err != nil {
			child.SandboxRoot = parent.SandboxRoot
			return child, fmt.Errorf("parent sandbox root: %w", err)
		}
		candidate, err := CanonicalPath(child.SandboxRoot)
		rel, relErr := filepath.Rel(root, candidate)
		if child.SandboxRoot == "" || err != nil || relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
			child.SandboxRoot = root
			widened = append(widened, "sandbox root")
		} else {
			child.SandboxRoot = candidate
		}
	}
	if len(widened) > 0 {
		return child, fmt.Errorf("child permissions narrowed: %s", strings.Join(widened, ", "))
	}
	return child, nil
}
func CanonicalPath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var tail []string
	for {
		resolved, err := filepath.EvalSymlinks(path)
		if err == nil {
			for i := len(tail) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, tail[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		if fi, e := os.Lstat(path); e == nil && fi.Mode()&os.ModeSymlink != 0 {
			return "", err
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", err
		}
		tail = append(tail, filepath.Base(path))
		path = parent
	}
}
