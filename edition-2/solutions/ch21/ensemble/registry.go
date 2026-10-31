package ensemble

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// NamedAgent is the durable identity of a peer, not an ephemeral task child.
type NamedAgent struct {
	// Name is a single directory-safe peer address.
	Name string `json:"name"`
	// Model selects the peer's model at restart.
	Model string `json:"model"`
	// Skill selects the peer's defining skill from the configured skills directory.
	Skill string `json:"skill"`
	// Created records the original creation time in RFC3339 form.
	Created time.Time `json:"created"`
}

// Start constructs the fleet application through one public composition path.
// Authentication and diagnostics are configured on New's root before this call.
// Existing NewAgent consumers remain free to use the smaller non-fleet protocol.
func (e *Ensemble) Start(dir string, cfg Config, spec AgentSpec, sink UserSink) (Agent, error) {
	e.registryMu.Lock()
	defer e.registryMu.Unlock()
	if e.root != nil {
		return nil, errors.New("Ensemble already started")
	}
	if sink == nil {
		return nil, errors.New("fleet requires a user sink")
	}
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	var entries []NamedAgent
	data, err := os.ReadFile(filepath.Join(absolute, "ensemble.json"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		if err = json.Unmarshal(data, &entries); err != nil {
			return nil, err
		}
	}
	if cfg.DataDir == "" {
		cfg.DataDir = filepath.Join(absolute, "main")
	} else if !filepath.IsAbs(cfg.DataDir) {
		cfg.DataDir = filepath.Join(absolute, cfg.DataDir)
	}
	e.directory, e.defaults, e.defaultSpec = absolute, cfg, spec
	e.named = map[string]Agent{}
	e.EnableFleet(sink)
	root, err := e.NewAgent(cfg, spec)
	if err != nil {
		return nil, err
	}
	if err = root.Load(); err != nil {
		root.Shutdown()
		return nil, err
	}
	e.root = root
	for _, entry := range entries {
		if _, err = e.createPeer(entry); err != nil {
			for _, a := range e.named {
				a.Shutdown()
			}
			root.Shutdown()
			e.root = nil
			return nil, err
		}
	}
	e.registry = entries
	return root, nil
}

// EnableFleet installs the recursive factory and human route before construction.
// Start also calls it; offline consumers may opt into the protocol without loading saves.
func (e *Ensemble) EnableFleet(sink UserSink) {
	e.SetUserSink("user", sink)
	e.SetAgentFactory(func(parent Agent, child AgentSpec) (Agent, error) {
		config := parent.Config()
		if child.Model != "" {
			config.Model = child.Model
		}
		return e.NewChildAgent(parent, config, child)
	})
}

func validName(name string) bool {
	return strings.TrimSpace(name) != "" && name != "." && name != ".." && !strings.ContainsAny(name, "/\\")
}
func (e *Ensemble) createPeer(entry NamedAgent) (Agent, error) {
	if !validName(entry.Name) || entry.Name == e.UserID() || entry.Name == e.root.Spec().ID {
		return nil, errors.New("invalid or reserved agent name")
	}
	if e.named[entry.Name] != nil {
		return nil, fmt.Errorf("agent %q already exists", entry.Name)
	}
	cfg, spec := e.defaults, e.defaultSpec
	cfg.Model = entry.Model
	spec.ID, spec.PrimarySkill = entry.Name, entry.Skill
	spec.Model = entry.Model
	spec.DataDir = filepath.Join(e.directory, "agents", entry.Name)
	spec.LogPath = filepath.Join(spec.DataDir, "events.jsonl")
	spec.SavePath = filepath.Join(spec.DataDir, "save.json")
	a, err := e.NewAgent(cfg, spec)
	if err != nil {
		return nil, err
	}
	if err = a.Load(); err == nil {
		err = a.Save()
	}
	if err != nil {
		a.Shutdown()
		return nil, err
	}
	e.named[entry.Name] = a
	return a, nil
}

// CreateAgent creates a named peer and persists its identity before announcing it.
// Empty model/skill inherit the application defaults; children are not registered.
func (e *Ensemble) CreateAgent(name, model, skill string) (Agent, error) {
	e.registryMu.Lock()
	defer e.registryMu.Unlock()
	if e.root == nil {
		return nil, errors.New("Ensemble is not started")
	}
	if model == "" {
		model = e.defaults.Model
	}
	if skill == "" {
		skill = e.defaultSpec.PrimarySkill
	}
	entry := NamedAgent{Name: name, Model: model, Skill: skill, Created: time.Now().UTC()}
	a, err := e.createPeer(entry)
	if err != nil {
		return nil, err
	}
	entries := append(append([]NamedAgent(nil), e.registry...), entry)
	if err = e.saveRegistry(entries); err != nil {
		delete(e.named, name)
		a.Shutdown()
		return nil, err
	}
	e.registry = entries
	return a, nil
}

// DeleteAgent accepts only named peers and refuses to orphan unfinished children.
// Application shutdown is separate: it preserves every peer's save and registry.
func (e *Ensemble) DeleteAgent(name string) error {
	e.registryMu.Lock()
	defer e.registryMu.Unlock()
	a := e.named[name]
	if a == nil {
		return fmt.Errorf("unknown agent %q", name)
	}
	if err := a.Fleet().PrepareDelete(); err != nil {
		return err
	}
	// Stop admission before deleting identity or files. No new children can appear
	// after this boundary, even if a browser races a model's final tool call.
	if err := a.Shutdown(); err != nil {
		return err
	}
	entries := make([]NamedAgent, 0, len(e.registry)-1)
	for _, entry := range e.registry {
		if entry.Name != name {
			entries = append(entries, entry)
		}
	}
	if err := e.saveRegistry(entries); err != nil {
		return err
	}
	e.registry = entries
	delete(e.named, name)
	return os.RemoveAll(a.Config().DataDir)
}

func (e *Ensemble) saveRegistry(entries []NamedAgent) error {
	if err := os.MkdirAll(e.directory, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(e.directory, ".ensemble-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = errors.Join(json.NewEncoder(file).Encode(entries), file.Close())
	if err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(e.directory, "ensemble.json"))
}

// Close stops the complete application lifetime and saves named conversations.
// Registry identities survive; explicit DeleteAgent is the only removal operation.
func (e *Ensemble) Close() error {
	e.registryMu.Lock()
	peers := []Agent{}
	if e.root != nil {
		peers = append(peers, e.root)
	}
	for _, a := range e.named {
		peers = append(peers, a)
	}
	e.registryMu.Unlock()
	var errs []error
	for _, a := range peers {
		errs = append(errs, a.Shutdown(), a.Save())
	}
	return errors.Join(errs...)
}

// AgentNode is the browser's recursive view. Key follows actual ownership so
// identical local names in different branches remain independently selectable.
type AgentNode struct {
	// ID is this Agent's local name, or ensemble for the synthetic container.
	ID string `json:"id"`
	// Key identifies one owned path, independent of its display label.
	Key string `json:"key"`
	// Parent is the containing Agent's key, empty for top-level peers.
	Parent string `json:"parent"`
	// Persistent distinguishes named peers/root from task children.
	Persistent bool `json:"persistent"`
	// Deletable is true only for named peers created through the registry.
	Deletable bool `json:"deletable"`
	// State reports lifecycle separately from current work.
	State string `json:"state"`
	// Activity identifies current provider, tool or waiting work.
	Activity Activity `json:"activity"`
	// Children contains directly owned descendants in name order.
	Children []AgentNode `json:"children"`
}

// AgentKey derives identity from actual parents, never from a global ID lookup.
func AgentKey(a Agent) string {
	if parent := a.ParentAgent(); parent != nil {
		return AgentKey(parent) + "/" + a.Spec().ID
	}
	return a.Spec().ID
}
func treeNode(a Agent, persistent, deletable bool) AgentNode {
	s := a.Fleet().Status()
	n := AgentNode{ID: a.Spec().ID, Key: AgentKey(a), Persistent: persistent, Deletable: deletable, State: s.State, Activity: a.Activity(), Children: []AgentNode{}}
	if p := a.ParentAgent(); p != nil {
		n.Parent = AgentKey(p)
	}
	for _, child := range a.Fleet().Children() {
		n.Children = append(n.Children, treeNode(child, false, false))
	}
	return n
}

// TreeSnapshot walks owned children; the synthetic root also contains named peers.
// The fallback preserves the existing GUI consumer built with NewAgent.
func (e *Ensemble) TreeSnapshot(fallback Agent) AgentNode {
	e.registryMu.Lock()
	defer e.registryMu.Unlock()
	root := e.root
	if root == nil {
		root = fallback
	}
	n := AgentNode{ID: "ensemble", Key: "", State: "running", Children: []AgentNode{}}
	if root != nil {
		n.Children = append(n.Children, treeNode(root, true, false))
	}
	names := make([]string, 0, len(e.named))
	for name := range e.named {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		n.Children = append(n.Children, treeNode(e.named[name], true, true))
	}
	return n
}

// FindAgent is a trusted client operation; model tools never call this tree walk.
func (e *Ensemble) FindAgent(key string, fallback Agent) (Agent, error) {
	e.registryMu.Lock()
	defer e.registryMu.Unlock()
	parts := strings.Split(key, "/")
	root := e.root
	if root == nil {
		root = fallback
	}
	var a Agent
	if root != nil && root.Spec().ID == parts[0] {
		a = root
	} else {
		a = e.named[parts[0]]
	}
	for _, id := range parts[1:] {
		if a == nil {
			break
		}
		a, _ = a.Fleet().Find(id)
	}
	if a == nil {
		return nil, errors.New("unknown agent")
	}
	return a, nil
}
