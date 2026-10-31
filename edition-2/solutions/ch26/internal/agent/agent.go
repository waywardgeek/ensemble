// Package agent owns conversation lifetime and each Agent's configuration.
package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ensemble/internal/common"
)

// The root retains this Agent; the Agent retains its children. Neither child
// imports the other implementation package. Their shared interfaces let the
// owner coordinate one actor without a second orchestration path.
type agent struct {
	activityMu   sync.Mutex
	activity     common.Activity
	progressTime time.Time
	parent       common.Ensemble
	parentAgent  common.Agent
	fleet        common.SubAgentManager
	spec         common.AgentSpec
	sandbox      common.Sandbox
	config       common.Config
	configMu     sync.Mutex
	engine       common.Engine
	history      common.History
	tools        []common.ToolDefinition
	jobs         common.Jobs
	skills       common.Skills
	mcp          common.MCP
	memory       common.Memory
	recaller     common.Recaller
	forceHandoff atomic.Bool
	once         sync.Once
	box          *box
	done         chan struct{}
	observersMu  sync.Mutex
	observers    []common.Observer
	progress     chan common.Observation
	stopErr      error
	lifecycle    sync.Mutex
	stopping     bool
	pauseMu      sync.Mutex
	resume       chan struct{}
}

// New creates an Agent with a required Ensemble parent; the root installs its
// children before use.
func New(parent common.Ensemble, cfg common.Config, spec common.AgentSpec, parents ...common.Agent) *agent {
	if parent == nil {
		panic("Agent requires Ensemble")
	}
	var spawning common.Agent
	if len(parents) > 0 {
		spawning = parents[0]
		if spawning == nil {
			panic("child Agent requires spawning parent")
		}
	}
	return &agent{parentAgent: spawning, parent: parent, config: cfg, spec: spec, box: newBox(), done: make(chan struct{}), progress: make(chan common.Observation, 256)}
}

// Ensemble returns the owning application root, including its logging services.
func (a *agent) Ensemble() common.Ensemble { return a.parent }

// Config returns request/workspace configuration with the current turn policy.
func (a *agent) Config() common.Config {
	a.configMu.Lock()
	defer a.configMu.Unlock()
	return a.config
}

// Engine reaches the transport and measured usage through the owning Agent.
func (a *agent) Engine() common.Engine { return a.engine }

// History exposes offline replay and snapshots of this Agent's facts.
func (a *agent) History() common.History { return a.history }

// The composition root installs this Agent's visible registry. Keeping one
// ordered slice makes declaration order stable and dispatch agree with visibility.
func (a *agent) SetTools(definitions []common.ToolDefinition) {
	a.tools = nil
	for _, definition := range definitions {
		if a.allowedTool(definition.Name) {
			a.tools = append(a.tools, definition)
		}
	}
}

// Declarations returns the ordered tools visible to this Agent's model.
func (a *agent) Declarations() []common.ToolDeclaration {
	var declarations []common.ToolDeclaration
	for _, tool := range a.definitions() {
		if tool.Ephemeral != "" || !a.toolEnabled(tool.Name) {
			continue
		}
		declaration := tool.ToolDeclaration
		declaration.Schema = append(json.RawMessage(nil), declaration.Schema...)
		declarations = append(declarations, declaration)
	}
	return declarations
}

// RegisterTool extends this Agent only. Register before Ask; conversation and
// registration are synchronous application operations, not concurrent inputs.
// Reject duplicate names so declarations and dispatch cannot silently disagree.
func (a *agent) RegisterTool(declaration common.ToolDeclaration, handler func(common.ToolContext, json.RawMessage) (string, error)) error {
	if !a.allowedTool(declaration.Name) {
		return errors.New("tool unavailable in safe mode")
	}
	if strings.TrimSpace(declaration.Name) == "" || handler == nil || !json.Valid(declaration.Schema) {
		return errors.New("tool requires a name, JSON schema and handler")
	}
	for _, tool := range a.tools {
		if tool.Name == declaration.Name {
			return fmt.Errorf("tool %q is already registered", declaration.Name)
		}
	}
	// Own the schema bytes: callers may reuse their input buffer after setup.
	declaration.Schema = append(json.RawMessage(nil), declaration.Schema...)
	a.tools = append(a.tools, common.ToolDefinition{ToolDeclaration: declaration, Run: handler})
	return nil
}

// A call has its actual Agent as parent; tools reach Engine and job services
// through that owner. Only the dispatch-created job is specific to this call.
type callContext struct {
	parent common.Agent
	job    common.Job
	limits common.Limits
	seq    uint64
}

// Agent returns the required owning Agent rather than a copied service bundle.
func (c *callContext) Agent() common.Agent { return c.parent }

// Engine reaches the transport and measured usage through the owning Agent.
func (c *callContext) Engine() common.Engine { return c.parent.Engine() }

// Job returns this call's execution handle; supervision calls have no new job.
func (c *callContext) Job() common.Job { return c.job }

// Limits returns this call's wait/output policy, never an execution deadline.
func (c *callContext) Limits() common.Limits { return c.limits }

// Jobs reaches this Agent's collection of continuing and completed executions.
func (a *agent) Jobs() common.Jobs { return a.jobs }

// AttachJobs installs the job collection after checking its parent is this Agent.
func (a *agent) AttachJobs(j common.Jobs) {
	if j == nil || j.Agent() != a {
		panic("Jobs must belong to this Agent")
	}
	a.jobs = j
}

// The root constructs children; attachment verifies both sides of ownership
// before exposing the assembled Agent to clients.
func (a *agent) AttachEngine(e common.Engine) {
	if e == nil || e.Agent() != a {
		panic("Engine must belong to this Agent")
	}
	a.engine = e
}

// AttachHistory installs history after checking its parent is this Agent.
func (a *agent) AttachHistory(h common.History) {
	if h == nil || h.Agent() != a {
		panic("History must belong to this Agent")
	}
	a.history = h
}

// SetMaxToolRounds changes the next turn's budget, preserving the active turn's
// snapshot. Zero selects the named default; negative budgets are invalid.
func (a *agent) SetMaxToolRounds(limit int) error {
	if limit < 0 {
		return fmt.Errorf("MaxToolRounds must be nonnegative")
	}
	a.configMu.Lock()
	defer a.configMu.Unlock()
	a.config.MaxToolRounds = limit
	return nil
}

// EventSeq identifies the dispatch fact that enabled a dynamically loaded tool.
func (c *callContext) EventSeq() uint64 { return c.seq }

// Skills returns this Agent's optional capability registry.
func (a *agent) Skills() common.Skills { return a.skills }

// AttachSkills establishes the Agent side of the registry's ownership link.
func (a *agent) AttachSkills(s common.Skills) {
	if s == nil || s.Agent() != a {
		panic("Skills must belong to this Agent")
	}
	a.skills = s
}

// RegisterVar affects future dynamic loads; construction already froze the primary.
func (a *agent) RegisterVar(name string, render func() string) {
	if a.skills != nil {
		a.skills.RegisterVar(name, render)
	}
}

// SetSystemPrompt installs the rendered constitution during construction only.
func (a *agent) SetSystemPrompt(prompt string) {
	a.configMu.Lock()
	defer a.configMu.Unlock()
	a.config.SystemPrompt = prompt
}

func (a *agent) allowedTool(name string) bool {
	if a.spec.Compaction {
		return name == "submit"
	}
	if a.spec.SafeMode {
		switch name {
		case "run_command", "send_input", "wait_for_job", "kill_job", "tool_limits":
			return false
		}
	}
	return true
}

func (a *agent) toolEnabled(name string) bool {
	if !a.allowedTool(name) {
		return false
	}
	if a.forceHandoff.Load() && name != "micro_handoff" {
		return false
	}
	if a.skills == nil {
		return true
	}
	_, enabled := a.skills.ToolMeta(name)
	return enabled
}

// MCP exposes this Agent's connection and discovered-tool lifetimes.
func (a *agent) MCP() common.MCP { return a.mcp }

// AttachMCP establishes both sides before connections or skills are loaded.
func (a *agent) AttachMCP(m common.MCP) {
	if m == nil || m.Agent() != a {
		panic("MCP must belong to this Agent")
	}
	a.mcp = m
}

func (a *agent) definitions() []common.ToolDefinition {
	definitions := append([]common.ToolDefinition(nil), a.tools...)
	if a.mcp != nil {
		// An external server cannot shadow a compiled capability. One name has
		// one declaration and one dispatch target, even across multiple peers.
		names := make(map[string]bool)
		for _, definition := range definitions {
			names[definition.Name] = true
		}
		for _, definition := range a.mcp.Tools() {
			if !names[definition.Name] {
				definitions = append(definitions, definition)
				names[definition.Name] = true
			}
		}
	}
	return definitions
}
func (a *agent) definition(name string, automatic bool) *common.ToolDefinition {
	for _, tool := range a.definitions() {
		if tool.Name == name && a.toolEnabled(name) && (automatic || tool.Ephemeral == "") {
			return &tool
		}
	}
	return nil
}

// SetContextSettings preserves zero defaults and clamps positive targets to the teaching floor.
func (a *agent) SetContextSettings(target, retention int) error {
	if target < 0 || retention < 0 {
		return fmt.Errorf("context settings must be nonnegative")
	}
	if target > 0 {
		target = max(20000, target)
	}
	a.configMu.Lock()
	defer a.configMu.Unlock()
	a.config.ContextTarget, a.config.LogRetention = target, retention
	return nil
}

// SetMemorySettings records content changes through the actor before acknowledging
// the patch. A disabled band disappears even while the Agent is idle.
func (a *agent) SetMemorySettings(settings common.MemorySettings) error {
	if a.spec.Compaction {
		return errors.New("compaction worker memory policy is immutable")
	}
	if err := a.memory.Validate(settings); err != nil {
		return err
	}
	reply := make(chan error, 1)
	if err := a.post(memorySettings{settings, reply}); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-a.done:
		return common.ErrStopped
	}
}

// AttachMemory establishes both sides of corpus ownership before runtime use.
func (a *agent) AttachMemory(m common.Memory) {
	if m == nil || m.Agent() != a {
		panic("Memory must belong to this Agent")
	}
	a.memory = m
}

// Memory returns the corpus service owned by this Agent.
func (a *agent) Memory() common.Memory { return a.memory }

// Spec returns identity and restrictions owned by this Agent.
func (a *agent) Spec() common.AgentSpec {
	spec := a.spec
	spec.SubmitSchema = append(json.RawMessage(nil), spec.SubmitSchema...)
	return spec
}

// Sandbox reaches the child enforcing this Agent's policy.
func (a *agent) Sandbox() common.Sandbox { return a.sandbox }

// AttachSandbox establishes both sides of the required ownership link.
func (a *agent) AttachSandbox(s common.Sandbox) {
	if s == nil || s.Agent() != a {
		panic("Sandbox must belong to this Agent")
	}
	a.sandbox = s
}

// ParentAgent exposes the spawning owner independently of Ensemble ownership.
func (a *agent) ParentAgent() common.Agent { return a.parentAgent }

// Fleet returns the manager whose lifetime follows this Agent.
func (a *agent) Fleet() common.SubAgentManager { return a.fleet }

// AttachFleet establishes the manager side of the required ownership link.
func (a *agent) AttachFleet(f common.SubAgentManager) {
	if f == nil || f.Agent() != a {
		panic("fleet must belong to Agent")
	}
	a.fleet = f
}
