package ensemble

import (
	"ensemble/internal/sandbox"
	"fmt"
	"strings"
)

// Child derives restrictions before overriding the child's state directory.
// Explicit log/save locations belong to the parent and must not be inherited.
// Submission contracts and managed purpose apply only to the original task.
func (s AgentSpec) Child(dataDir string) AgentSpec {
	s.DataDir, s.LogPath, s.SavePath = dataDir, "", ""
	s.SubmitSchema = nil
	s.Compaction = false
	return s
}

// Clamp always returns a narrowed spec. Its diagnostic is an alarm, never the
// enforcement: ignoring the error must not restore a requested wider boundary.
func (s AgentSpec) Clamp(parent AgentSpec) (AgentSpec, error) {
	var widened []string
	// SafeMode is a restriction (OR); web search is a capability (AND).
	if parent.SafeMode && !s.SafeMode {
		s.SafeMode = true
		widened = append(widened, "safe mode")
	}
	if !parent.EnableWebSearch && s.EnableWebSearch {
		s.EnableWebSearch = false
		widened = append(widened, "web search")
	}
	if parent.SandboxRoot != "" && !sandbox.Within(parent.SandboxRoot, s.SandboxRoot) {
		s.SandboxRoot = parent.SandboxRoot
		widened = append(widened, "sandbox root")
	}
	if len(widened) != 0 {
		return s, fmt.Errorf("child permissions narrowed: %s", strings.Join(widened, ", "))
	}
	return s, nil
}
