package ensemble

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"path/filepath"
)

type PolicySnapshot = common.PolicySnapshot
type PolicyAck = common.PolicyAck
type TurnPolicy = common.TurnPolicy
type SettingsError = common.SettingsError

const DefaultMaxModelRequests = common.DefaultMaxModelRequests

func (e *Ensemble) ClaimSettingsPath(path string) (string, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid settings path")
	}
	// Resolve existing symlink ancestors as well as an existing destination.
	suffix := []string{}
	root := resolved
	for {
		canonical, err := filepath.EvalSymlinks(root)
		if err == nil {
			resolved = canonical
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			break
		}
		next := filepath.Dir(root)
		if next == root {
			return "", fmt.Errorf("cannot resolve settings path")
		}
		suffix = append(suffix, filepath.Base(root))
		root = next
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return "", fmt.Errorf("application closed")
	}
	if e.settingsPaths == nil {
		e.settingsPaths = map[string]bool{}
	}
	if e.settingsPaths[resolved] {
		return "", fmt.Errorf("settings path already has a live owner")
	}
	if e.sessionReservations["path:"+filepath.Dir(resolved)] {
		switch filepath.Base(resolved) {
		case "owner.lock", "events.log", "checkpoint.json", "origin.json":
			return "", fmt.Errorf("settings path conflicts with session storage")
		}
	}
	e.settingsPaths[resolved] = true
	return resolved, nil
}
func (e *Ensemble) ReleaseSettingsPath(path string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.settingsPaths, path)
}
func (a *Agent) ExecutionPolicy() PolicySnapshot {
	if a.policy == nil {
		return PolicySnapshot{EffectiveMaxModelRequests: DefaultMaxModelRequests}
	}
	return a.policy.Snapshot()
}
func (a *Agent) UpdatePolicy(base uint64, patch json.RawMessage) (PolicyAck, error) {
	if a.actor == nil {
		return PolicyAck{}, &SettingsError{Code: "settings_closed", Message: "Agent is read-only"}
	}
	return a.actor.UpdatePolicy(base, patch)
}
func (e *Ensemble) ExecutionPolicy(id string) (PolicySnapshot, error) {
	a, err := e.Agent(id)
	if err != nil {
		return PolicySnapshot{}, err
	}
	return a.ExecutionPolicy(), nil
}
func (e *Ensemble) UpdatePolicy(id string, base uint64, patch json.RawMessage) (PolicyAck, error) {
	a, err := e.Agent(id)
	if err != nil {
		return PolicyAck{}, err
	}
	return a.UpdatePolicy(base, patch)
}
