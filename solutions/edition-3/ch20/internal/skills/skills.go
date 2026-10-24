package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type skillState struct {
	info     common.SkillInfo
	rendered string
}
type registry struct {
	parent common.Agent
	// mu protects capability truth; no application callback runs under it.
	mu           sync.Mutex
	tools        []common.ToolSpec
	skills       map[string]*skillState
	active       []string
	vars         map[string]func() string
	listeners    map[uint64]func()
	nextListener uint64
}

func New(parent common.Agent, tools []common.ToolSpec) (common.SkillRegistry, error) {
	if parent == nil {
		return nil, fmt.Errorf("skills require Agent parent")
	}
	r := &registry{parent: parent, skills: map[string]*skillState{}, vars: map[string]func() string{}, listeners: map[uint64]func(){}}
	for _, t := range tools {
		if err := r.addTool(t); err != nil {
			return nil, err
		}
	}
	if spec := parent.Config().Skills; spec != nil {
		if spec.Directory == "" || spec.Primary == "" {
			return nil, fmt.Errorf("skills require directory and primary")
		}
		if parent.Config().Tools != nil {
			return nil, fmt.Errorf("Tools and Skills cannot both select startup capabilities")
		}
		if err := r.discover(spec.Directory); err != nil {
			return nil, err
		}
		for name, f := range spec.Variables {
			r.vars[name] = f
		}
	}
	return r, nil
}
func (r *registry) Parent() common.Agent { return r.parent }
func (r *registry) RegisterTool(t common.ToolSpec) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.addTool(t); err != nil {
		return err
	}
	return r.publishDeclarationsLocked()
}

// Construction populates the catalogue before startup/recovered grants exist.
// It must not publish live deltas against an already initialized History. Other
// callers hold mu and publish only after their complete registration succeeds.
func (r *registry) addTool(t common.ToolSpec) error {
	if strings.TrimSpace(t.Name) == "" || strings.TrimSpace(t.Description) == "" || t.Handler == nil {
		return fmt.Errorf("tool requires name, description and handler")
	}
	if len(t.Parameters) == 0 {
		t.Parameters = json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)
	}
	var schema map[string]json.RawMessage
	if json.Unmarshal(t.Parameters, &schema) != nil || schema == nil {
		return fmt.Errorf("tool parameters must be an object schema")
	}
	var schemaType string
	if json.Unmarshal(schema["type"], &schemaType) != nil || schemaType != "object" {
		return fmt.Errorf("tool parameters must have explicit object type")
	}
	t.Parameters = append(json.RawMessage(nil), t.Parameters...)
	if t.Source == "" {
		t.Source = "application"
	}
	if t.Enabled && t.Origin == "" {
		t.Origin = "manual"
	}
	if r.toolIndex(t.Name) >= 0 {
		return fmt.Errorf("tool %q already registered", t.Name)
	}
	r.tools = append(r.tools, t)
	return nil
}

// RegisterExternal publishes an explicit application's discovered catalogue as
// one batch. Skill-configured Agents instead admit through load's grant check.
func (r *registry) RegisterExternal(specs []common.ToolSpec) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.parent.Config().Skills != nil {
		return fmt.Errorf("external capabilities require a skill grant")
	}
	names := map[string]bool{}
	for _, t := range specs {
		if t.Name == "" || t.Handler == nil || !json.Valid(t.Parameters) {
			return fmt.Errorf("invalid external capability")
		}
		if names[t.Name] || r.toolIndex(t.Name) >= 0 {
			return fmt.Errorf("tool already registered: %s", t.Name)
		}
		names[t.Name] = true
	}
	for _, t := range specs {
		t.Enabled = r.parent.Config().Tools == nil
		for _, name := range r.parent.Config().Tools {
			if name == t.Name {
				t.Enabled = true
			}
		}
		t.Origin = "application"
		r.tools = append(r.tools, t)
	}
	return r.publishDeclarationsLocked()
}

func (r *registry) toolIndex(name string) int {
	for i, t := range r.tools {
		if t.Name == name {
			return i
		}
	}
	return -1
}
func (r *registry) RegisterVar(name string, f func() string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if f == nil {
		delete(r.vars, name)
	} else {
		r.vars[name] = f
	}
}
func (r *registry) LoadInitial() error {
	cfg := r.parent.Config()
	if cfg.Skills != nil {
		if err := r.load(cfg.Skills.Primary, "initial", 0); err != nil {
			return err
		}
		for _, name := range cfg.StartupSkills {
			if err := r.load(name, "initial", 0); err != nil {
				return fmt.Errorf("startup skill %q: %w", name, err)
			}
		}
		return nil
	}
	if len(cfg.StartupSkills) != 0 {
		return fmt.Errorf("startup skills require a skill directory and primary")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	names := cfg.Tools
	if names == nil {
		for _, t := range r.tools {
			if t.Name != "load_skill" && t.Name != "unload_skill" {
				names = append(names, t.Name)
			}
		}
	}
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			return fmt.Errorf("duplicate enabled tool %q", name)
		}
		seen[name] = true
		if r.toolIndex(name) < 0 {
			return fmt.Errorf("unknown enabled tool %q", name)
		}
	}
	for _, name := range names {
		t := &r.tools[r.toolIndex(name)]
		t.Enabled = true
		t.Origin = "initial"
	}
	return nil
}
func (r *registry) loadable(enabled map[string]bool) []common.SkillInfo {
	offered := map[string]bool{}
	for name, s := range r.skills {
		if enabled[name] {
			for _, n := range s.info.Offers {
				offered[n] = true
			}
		}
	}
	names := []string{}
	for name := range offered {
		if name == "web-search" && !r.parent.Config().EnableWebSearch {
			continue
		}
		if s := r.skills[name]; s != nil && s.info.Type == "loadable" && !enabled[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	out := []common.SkillInfo{}
	for _, n := range names {
		out = append(out, cloneSkill(r.skills[n].info))
	}
	return out
}
func (r *registry) enabled() map[string]bool {
	out := map[string]bool{}
	for n, s := range r.skills {
		out[n] = s.info.Enabled
	}
	return out
}
func (r *registry) LoadableSkills() []common.SkillInfo {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.loadable(r.enabled())
}
func (r *registry) LoadSkill(name string) error { return r.load(name, "manual", 0) }
func (r *registry) LoadDynamic(name string, seq common.Seq) error {
	if seq == 0 {
		return fmt.Errorf("dynamic skill requires tool_called sequence")
	}
	return r.load(name, "tool", seq)
}
func (r *registry) load(name, origin string, seq common.Seq) error {
	callbacks, err := r.activate(name, origin, seq)
	if r.parent.MCP() != nil {
		r.parent.MCP().Reconcile(r.Snapshot())
	}
	for _, f := range callbacks {
		f()
	}
	return err
}
func (r *registry) activate(name, origin string, seq common.Seq) ([]func(), error) {
	r.mu.Lock()
	fail := func(err error) ([]func(), error) { r.mu.Unlock(); return nil, err }
	root := r.skills[name]
	if root == nil {
		return fail(fmt.Errorf("unknown skill %q", name))
	}
	if origin == "initial" && name == r.parent.Config().Skills.Primary {
		if root.info.Type != "primary" {
			return fail(fmt.Errorf("initial skill %q is not primary", name))
		}
	} else {
		if root.info.Type != "loadable" {
			return fail(fmt.Errorf("skill %q is not loadable", name))
		}
		if !root.info.Enabled {
			found := false
			for _, s := range r.loadable(r.enabled()) {
				if s.Name == name {
					found = true
				}
			}
			if !found {
				return fail(fmt.Errorf("skill %q is not offered", name))
			}
		}
	}
	order := []string{}
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var visit func(string) error
	visit = func(name string) error {
		if name == "web-search" && !r.parent.Config().EnableWebSearch {
			return fmt.Errorf("web search is disabled")
		}
		s := r.skills[name]
		if s == nil {
			return fmt.Errorf("missing dependency %q", name)
		}
		if visiting[name] {
			return fmt.Errorf("skill dependency cycle at %q", name)
		}
		if visited[name] {
			return nil
		}
		visiting[name] = true
		for _, dep := range s.info.Depends {
			if err := visit(dep); err != nil {
				return err
			}
		}
		for _, tool := range s.info.Tools {
			if r.toolIndex(tool) < 0 && len(s.info.MCPServers) == 0 {
				return fmt.Errorf("skill %q requires missing tool %q", name, tool)
			}
		}
		visiting[name] = false
		visited[name] = true
		order = append(order, name)
		return nil
	}
	if err := visit(name); err != nil {
		return fail(err)
	}
	vars := map[string]func() string{}
	for n, f := range r.vars {
		vars[n] = f
	}
	planned := []common.Skill{}
	for _, n := range order {
		planned = append(planned, cloneSkill(r.skills[n].info).Skill)
	}
	r.mu.Unlock()
	var remote []common.ToolSpec
	if r.parent.MCP() != nil {
		var err error
		remote, err = r.parent.MCP().Discover(context.Background(), planned)
		if err != nil {
			return nil, err
		}
	}
	// Graph identities are immutable and tool registration is append-only. Resolve
	// application values outside every lock, then publish against the current set.
	values := map[string]string{}
	for n, f := range vars {
		if n != "TOOLS" && n != "SKILLS" && f != nil {
			values[n] = f()
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	// Stage the entire remote catalog before changing capability truth.
	for _, t := range remote {
		if i := r.toolIndex(t.Name); i >= 0 && (r.tools[i].Source != t.Source || r.tools[i].Ephemeral != t.Ephemeral || !bytes.Equal(r.tools[i].Parameters, t.Parameters)) {
			return nil, fmt.Errorf("MCP tool collision or schema drift: %s", t.Name)
		}
	}
	available := map[string]bool{}
	for _, t := range r.tools {
		available[t.Name] = true
	}
	for _, t := range remote {
		available[t.Name] = true
	}
	for _, s := range planned {
		for _, name := range s.Tools {
			if !available[name] {
				return nil, fmt.Errorf("skill %q requires missing tool %q", s.Name, name)
			}
			source := ""
			if i := r.toolIndex(name); i >= 0 {
				source = r.tools[i].Source
			}
			for _, t := range remote {
				if t.Name == name {
					source = t.Source
				}
			}
			if strings.HasPrefix(source, "mcp:") {
				declared := false
				for _, d := range s.MCPServers {
					if source == "mcp:"+d.Name {
						declared = true
					}
				}
				if !declared {
					return nil, fmt.Errorf("skill %q has no server declaration for %q", s.Name, name)
				}
			}
		}
	}
	for _, t := range remote {
		if r.toolIndex(t.Name) < 0 {
			r.tools = append(r.tools, t)
		}
	}
	enabled := r.enabled()
	toolSet := map[string]bool{}
	for _, t := range r.tools {
		if t.Enabled && t.Ephemeral == "" && common.ToolPolicy(r.parent.Config(), t.Name) == nil {
			toolSet[t.Name] = true
		}
	}
	if origin == "initial" {
		toolSet["load_skill"] = true
		toolSet["unload_skill"] = true
		toolSet["compact_context"], toolSet["micro_handoff"], toolSet["keep_tool_results"] = true, true, true
	}
	for _, n := range order {
		enabled[n] = true
		for _, tool := range r.skills[n].info.Tools {
			if i := r.toolIndex(tool); i >= 0 && r.tools[i].Ephemeral != "" {
				continue
			}
			toolSet[tool] = true
		}
	}
	tools := []string{}
	for n := range toolSet {
		tools = append(tools, n)
	}
	sort.Strings(tools)
	loadables := []string{}
	for _, s := range r.loadable(enabled) {
		loadables = append(loadables, s.Name+": "+s.Description)
	}
	values["TOOLS"] = strings.Join(tools, ", ")
	values["SKILLS"] = strings.Join(loadables, "\n")
	changed := false
	if origin == "initial" {
		for _, n := range []string{"load_skill", "unload_skill", "compact_context", "micro_handoff", "keep_tool_results"} {
			t := &r.tools[r.toolIndex(n)]
			t.Enabled = true
			t.Origin = "initial"
		}
	}
	for _, n := range order {
		s := r.skills[n]
		if s.info.PendingUnload {
			if err := r.parent.RecordFact(common.Event{Type: common.SkillLoaded, Skill: &common.SkillData{Name: n, Body: s.rendered, Origin: origin, ActiveSeq: seq}}); err != nil {
				return nil, err
			}
			s.info.PendingUnload = false
			changed = true
		}
		if s.info.Enabled {
			continue
		}
		body := expand(s.info.Instructions, values)
		if err := r.parent.RecordFact(common.Event{Type: common.SkillLoaded, Skill: &common.SkillData{Name: n, Body: body, Origin: origin, ActiveSeq: seq}}); err != nil {
			return nil, err
		}
		s.info.Enabled = true
		s.info.Origin = origin
		s.info.EventSeq = seq
		s.rendered = body
		r.active = append(r.active, n)
		changed = true
		for _, n := range s.info.Tools {
			t := &r.tools[r.toolIndex(n)]
			if t.Enabled {
				continue
			}
			t.Enabled = true
			t.Skill = s.info.Name
			t.Origin = origin
			t.EventSeq = seq
		}
	}
	r.reconcileGrantsLocked()
	if err := r.publishDeclarationsLocked(); err != nil {
		return nil, err
	}
	return r.callbacks(changed), nil
}
func (r *registry) UnloadSkill(name string) error {
	r.mu.Lock()
	s := r.skills[name]
	if s == nil || !s.info.Enabled {
		r.mu.Unlock()
		return fmt.Errorf("skill %q is not enabled", name)
	}
	changed := !s.info.PendingUnload
	if changed {
		if err := r.parent.RecordFact(common.Event{Type: common.SkillUnloaded, Skill: &common.SkillData{Name: name}}); err != nil {
			r.mu.Unlock()
			return err
		}
	}
	s.info.PendingUnload = true
	r.reconcileGrantsLocked()
	if err := r.publishDeclarationsLocked(); err != nil {
		r.mu.Unlock()
		return err
	}
	callbacks := r.callbacks(changed)
	r.mu.Unlock()
	if m := r.parent.MCP(); m != nil {
		m.Reconcile(r.Snapshot())
	}
	for _, f := range callbacks {
		f()
	}
	return nil
}

// RefreshPolicy withdraws web grants using the existing pending-unload state.
// Re-enabling exposes the declaration for an explicit reload, not a silent dial.
func (r *registry) RefreshPolicy() error {
	r.mu.Lock()
	web := r.skills["web-search"]
	unload := !r.parent.Config().EnableWebSearch && web != nil && web.info.Enabled && !web.info.PendingUnload
	r.mu.Unlock()
	var err error
	if unload {
		err = r.UnloadSkill("web-search")
	}
	r.mu.Lock()
	if err == nil {
		err = r.publishDeclarationsLocked()
	}
	callbacks := r.callbacks(true)
	r.mu.Unlock()
	if m := r.parent.MCP(); m != nil {
		m.Reconcile(r.Snapshot())
	}
	for _, f := range callbacks {
		f()
	}
	return err
}

func (r *registry) callbacks(changed bool) []func() {
	var out []func()
	if changed {
		for _, f := range r.listeners {
			out = append(out, f)
		}
	}
	return out
}
func (r *registry) SubscribeSkills(f func()) func() {
	if f == nil {
		return func() {}
	}
	r.mu.Lock()
	r.nextListener++
	id := r.nextListener
	r.listeners[id] = f
	r.mu.Unlock()
	return func() { r.mu.Lock(); delete(r.listeners, id); r.mu.Unlock() }
}
func cloneSkill(s common.SkillInfo) common.SkillInfo {
	s.Tools = append([]string{}, s.Tools...)
	s.Depends = append([]string{}, s.Depends...)
	s.Offers = append([]string{}, s.Offers...)
	s.MCPServers = append([]common.MCPServer(nil), s.MCPServers...)
	for i := range s.MCPServers {
		s.MCPServers[i].Args = append([]string(nil), s.MCPServers[i].Args...)
	}
	return s
}

// MCPAuthEnvNames reads known declarations, not capability visibility. Disabling
// a skill cannot make its already-known credentials ordinary subprocess inputs.
// Return names only; this does not grant or advertise a skill or tool.
func (r *registry) MCPAuthEnvNames() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	known := map[string]bool{}
	for _, skill := range r.skills {
		for _, server := range skill.info.MCPServers {
			if server.AuthEnv != "" {
				known[server.AuthEnv] = true
			}
		}
	}
	names := make([]string, 0, len(known))
	for name := range known {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *registry) ListSkills() []common.SkillInfo { return r.Snapshot().Skills }
func (r *registry) Snapshot() common.CapabilitySnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := common.CapabilitySnapshot{Definitions: []common.ToolDefinition{}, Tools: []common.ToolSpec{}, Skills: []common.SkillInfo{}}
	for _, t := range r.tools {
		t.Parameters = append(json.RawMessage(nil), t.Parameters...)
		if common.ToolPolicy(r.parent.Config(), t.Name) != nil || (t.Skill == "web-search" && !r.parent.Config().EnableWebSearch) {
			t.Enabled = false
		}
		out.Tools = append(out.Tools, t)
		if t.Enabled && t.Ephemeral == "" && common.ToolPolicy(r.parent.Config(), t.Name) == nil {
			out.Definitions = append(out.Definitions, t.ToolDefinition)
		}
	}
	names := []string{}
	for n := range r.skills {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if n != "web-search" || r.parent.Config().EnableWebSearch {
			out.Skills = append(out.Skills, cloneSkill(r.skills[n].info))
		}
	}
	for _, n := range r.active {
		if n != "web-search" || r.parent.Config().EnableWebSearch {
			out.Instructions += "\n\n" + r.skills[n].rendered
		}
	}
	return out
}
func expand(text string, values map[string]string) string {
	var out strings.Builder
	for i := 0; i < len(text); {
		if text[i] != '$' {
			out.WriteByte(text[i])
			i++
			continue
		}
		end := i + 1
		for end < len(text) && (text[end] >= 'A' && text[end] <= 'Z' || text[end] >= 'a' && text[end] <= 'z' || text[end] >= '0' && text[end] <= '9' || text[end] == '_') {
			end++
		}
		name := text[i+1 : end]
		if value, ok := values[name]; ok {
			out.WriteString(value)
		} else {
			out.WriteString(text[i:end])
		}
		i = end
	}
	return out.String()
}

// Instruction survival is separate from grants: pending unload withdraws its
// compiled and external grants while independent live grants remain enabled.
func (r *registry) reconcileGrantsLocked() {
	for i := range r.tools {
		t := &r.tools[i]
		external := strings.HasPrefix(t.Source, "mcp:")
		if !external && t.Skill == "" {
			continue
		} // independent manual/startup grant
		t.Enabled = false
		for _, s := range r.skills {
			if !s.info.Enabled || s.info.PendingUnload {
				continue
			}
			if !external {
				for _, name := range s.info.Tools {
					if name == t.Name {
						t.Enabled = true
					}
				}
				continue
			}
			for _, d := range s.info.MCPServers {
				if t.Source != "mcp:"+d.Name {
					continue
				}
				allowed := false
				for _, name := range s.info.Tools {
					if name == t.Name {
						allowed = true
					}
				}
				if allowed {
					t.Enabled = true
					if t.Origin == "" {
						t.Origin = s.info.Origin
						t.Skill = s.info.Name
						t.EventSeq = s.info.EventSeq
					}
				}
			}
		}
	}
}

func (r *registry) RestoreContext(c common.Context) error {
	if r.parent.Config().Skills == nil {
		return r.LoadInitial()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, saved := range c.ActiveSkills {
		if saved.Name == "web-search" && !r.parent.Config().EnableWebSearch {
			return fmt.Errorf("saved web-search skill is disabled")
		}
		if r.skills[saved.Name] == nil {
			return fmt.Errorf("saved skill %q is unavailable", saved.Name)
		}
	}
	for _, n := range []string{"load_skill", "unload_skill", "compact_context", "micro_handoff", "keep_tool_results"} {
		if i := r.toolIndex(n); i >= 0 {
			r.tools[i].Enabled = true
			r.tools[i].Origin = "initial"
		}
	}
	for _, saved := range c.ActiveSkills {
		if saved.Name == "web-search" && !r.parent.Config().EnableWebSearch {
			return fmt.Errorf("saved web-search skill is disabled")
		}
		s := r.skills[saved.Name]
		if s == nil {
			return fmt.Errorf("saved skill %q is unavailable", saved.Name)
		}
		s.info.Enabled, s.info.PendingUnload = true, saved.PendingUnload
		s.info.Origin, s.info.EventSeq, s.rendered = saved.Origin, saved.ActiveSeq, saved.Body
		r.active = append(r.active, saved.Name)
		for _, name := range s.info.Tools {
			if i := r.toolIndex(name); i >= 0 && r.tools[i].Origin == "" {
				t := &r.tools[i]
				t.Skill, t.Origin, t.EventSeq = saved.Name, saved.Origin, saved.ActiveSeq
			}
		}
	}
	r.reconcileGrantsLocked()
	return nil
}

// publishDeclarationsLocked is called after a successful grant mutation. It only
// submits a fact; the actor does not call back into the locked registry.
func (r *registry) publishDeclarationsLocked() error {
	c := r.parent.Log().Context()
	if !c.ToolsInitialized {
		return nil
	} // Root freezes the complete startup set.
	before := map[string]common.ToolDeclaration{}
	for _, d := range c.Tools {
		before[d.Name] = d
	}
	current := map[string]bool{}
	delta := common.ToolsData{}
	for _, tool := range r.tools {
		if !tool.Enabled || tool.Ephemeral != "" || common.ToolPolicy(r.parent.Config(), tool.Name) != nil || (tool.Skill == "web-search" && !r.parent.Config().EnableWebSearch) {
			continue
		}
		d := common.ToolDeclaration{Name: tool.Name, Description: tool.Description, Schema: tool.Parameters}
		current[d.Name] = true
		old, ok := before[d.Name]
		if !ok || old.Description != d.Description || !bytes.Equal(old.Schema, d.Schema) {
			delta.Added = append(delta.Added, d)
		}
	}
	for _, d := range c.Tools {
		if !current[d.Name] {
			delta.Removed = append(delta.Removed, d.Name)
		}
	}
	if len(delta.Added) == 0 && len(delta.Removed) == 0 {
		return nil
	}
	return r.parent.RecordFact(common.Event{Type: common.ToolsChanged, Tools: &delta})
}
