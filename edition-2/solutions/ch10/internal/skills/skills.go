package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"ensemble/internal/common"
)

// State belongs to the Agent, not to the process or a shared skills directory.
// The lock permits observation while a tool worker loads instructions. Renderers
// run outside it because $TOOLS queries declarations through this same registry.
type registry struct {
	parent   common.Agent
	mu       sync.Mutex
	all      map[string]skill
	loaded   map[string]bool
	pending  map[string]bool
	tools    map[string]common.ToolMeta
	vars     map[string]func() string
	variable *regexp.Regexp
}

// New discovers files without enabling them; the root attaches this child before
// loading the primary so variable renderers can reach the complete ownership graph.
func New(parent common.Agent, directory string) (common.Skills, error) {
	if parent == nil {
		panic("Skills requires Agent")
	}
	r := &registry{parent: parent, all: make(map[string]skill), loaded: make(map[string]bool), pending: make(map[string]bool), tools: make(map[string]common.ToolMeta), vars: make(map[string]func() string), variable: regexp.MustCompile(`\$[A-Za-z_][A-Za-z0-9_]*`)}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(directory, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		s, err := r.parse(string(data))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if _, exists := r.all[s.name]; exists {
			return nil, fmt.Errorf("duplicate skill %q", s.name)
		}
		r.all[s.name] = s
	}
	// Lifecycle tools remain available even for a deliberately tiny primary.
	r.tools["load_skill"] = common.ToolMeta{Source: common.SourceInitial}
	r.tools["unload_skill"] = common.ToolMeta{Source: common.SourceInitial}
	r.RegisterVar("TOOLS", func() string {
		var lines []string
		for _, d := range r.parent.Declarations() {
			lines = append(lines, d.Name+": "+d.Description)
		}
		return strings.Join(lines, "\n")
	})
	r.RegisterVar("SKILLS", func() string { return strings.Join(r.LoadableSkills(), ", ") })
	return r, nil
}

// Agent exposes the required owning Agent rather than a sibling-service bundle.
func (r *registry) Agent() common.Agent { return r.parent }

// RegisterVar replaces a renderer for future loads; prior text stays frozen.
func (r *registry) RegisterVar(name string, render func() string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.vars[name] = render
}

func (r *registry) render(body string) string {
	r.mu.Lock()
	vars := make(map[string]func() string, len(r.vars))
	for name, render := range r.vars {
		vars[name] = render
	}
	r.mu.Unlock()
	// Unknown placeholders remain visible so missing application setup is legible.
	return r.variable.ReplaceAllStringFunc(body, func(token string) string {
		if render := vars[token[1:]]; render != nil {
			return render()
		}
		return token
	})
}

// LoadInitial is called only during construction, before the first request.
func (r *registry) LoadInitial(name string) (string, error) {
	return r.load(name, common.ToolMeta{Source: common.SourceInitial})
}

// LoadDynamic returns instructions for history, never a replacement system prompt.
func (r *registry) LoadDynamic(name string, seq uint64) (string, error) {
	return r.load(name, common.ToolMeta{Source: common.SourceDynamic, EventSeq: seq})
}

func (r *registry) load(name string, meta common.ToolMeta) (string, error) {
	r.mu.Lock()
	bodies, err := r.enable(name, meta)
	r.mu.Unlock()
	if err != nil {
		return "", err
	}
	return r.render(strings.Join(bodies, "\n\n")), nil
}

// Resolve the whole dependency graph before changing state. A missing dependency
// or cycle must not partially grant tools from an otherwise rejected load.
func (r *registry) enable(name string, meta common.ToolMeta) ([]string, error) {
	s, exists := r.all[name]
	if !exists {
		return nil, fmt.Errorf("unknown skill %q", name)
	}
	if meta.Source == common.SourceInitial {
		if s.kind != "primary" {
			return nil, fmt.Errorf("skill %q is not primary", name)
		}
	} else {
		if s.kind != "loadable" {
			return nil, fmt.Errorf("skill %q is not loadable", name)
		}
		if !r.loaded[name] && !r.loadable()[name] {
			return nil, fmt.Errorf("skill %q is not available to load", name)
		}
	}
	visiting, seen := make(map[string]bool), make(map[string]bool)
	var order []string
	var visit func(string) error
	visit = func(n string) error {
		if visiting[n] {
			return fmt.Errorf("skill dependency cycle at %q", n)
		}
		if r.loaded[n] || seen[n] {
			return nil
		}
		current, ok := r.all[n]
		if !ok {
			return fmt.Errorf("missing skill dependency %q", n)
		}
		visiting[n] = true
		for _, dependency := range current.depends {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		delete(visiting, n)
		seen[n] = true
		order = append(order, n)
		return nil
	}
	if err := visit(name); err != nil {
		return nil, err
	}
	if len(order) == 0 {
		return []string{"Skill " + name + " is already loaded."}, nil
	}
	var bodies []string
	for _, n := range order {
		current := r.all[n]
		r.loaded[n] = true
		for _, tool := range current.tools {
			if _, enabled := r.tools[tool]; !enabled {
				r.tools[tool] = meta
			}
		}
		bodies = append(bodies, current.body)
	}
	return bodies, nil
}

// loadable is called under the lock and exposes exactly one discovery level.
func (r *registry) loadable() map[string]bool {
	out := make(map[string]bool)
	for name := range r.loaded {
		for _, offered := range r.all[name].loadable {
			if s, ok := r.all[offered]; ok && s.kind == "loadable" && !r.loaded[offered] {
				out[offered] = true
			}
		}
	}
	return out
}

// LoadableSkills omits hidden dependencies, primaries and already-loaded skills.
func (r *registry) LoadableSkills() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var names []string
	for name := range r.loadable() {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ToolMeta is shared by declaration filtering and execution admission.
func (r *registry) ToolMeta(name string) (common.ToolMeta, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	meta, enabled := r.tools[name]
	return meta, enabled
}

// MarkUnload defers removal to the future compaction boundary.
func (r *registry) MarkUnload(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.loaded[name] {
		return fmt.Errorf("skill %q is not loaded", name)
	}
	r.pending[name] = true
	return nil
}

// PendingUnload exposes the deferred state without prematurely removing capability.
func (r *registry) PendingUnload() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var names []string
	for name := range r.pending {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
