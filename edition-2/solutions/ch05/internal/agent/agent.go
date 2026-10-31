// Package agent owns conversation lifetime and each Agent's configuration.
package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"ensemble/internal/common"
)

// The root retains this Agent; the Agent retains its children. Neither child
// imports the other implementation package. Their shared interfaces let the
// owner coordinate one actor without a second orchestration path.
type agent struct {
	parent      common.Ensemble
	config      common.Config
	engine      common.Engine
	history     common.History
	tools       []common.ToolDefinition
	jobs        common.Jobs
	once        sync.Once
	box         *box
	done        chan struct{}
	observersMu sync.Mutex
	observers   []common.Observer
	progress    chan common.Observation
	stopErr     error
	lifecycle   sync.Mutex
	stopping    bool
}

// New creates an Agent with a required Ensemble parent; the root installs its
// children before use.
func New(parent common.Ensemble, cfg common.Config) *agent {
	if parent == nil {
		panic("Agent requires Ensemble")
	}
	return &agent{parent: parent, config: cfg, box: newBox(), done: make(chan struct{}), progress: make(chan common.Observation, 256)}
}

// Ensemble returns the owning application root, including its logging services.
func (a *agent) Ensemble() common.Ensemble { return a.parent }

// Config returns this Agent's fixed request and workspace configuration.
func (a *agent) Config() common.Config { return a.config }

// Engine reaches the transport and measured usage through the owning Agent.
func (a *agent) Engine() common.Engine { return a.engine }

// History exposes offline replay and snapshots of this Agent's facts.
func (a *agent) History() common.History { return a.history }

// The composition root installs this Agent's visible registry. Keeping one
// ordered slice makes declaration order stable and dispatch agree with visibility.
func (a *agent) SetTools(definitions []common.ToolDefinition) { a.tools = definitions }

// Declarations returns the ordered tools visible to this Agent's model.
func (a *agent) Declarations() []common.ToolDeclaration {
	var declarations []common.ToolDeclaration
	for _, tool := range a.tools {
		declarations = append(declarations, tool.ToolDeclaration)
	}
	return declarations
}

// RegisterTool extends this Agent only. Register before Ask; conversation and
// registration are synchronous application operations, not concurrent inputs.
// Reject duplicate names so declarations and dispatch cannot silently disagree.
func (a *agent) RegisterTool(declaration common.ToolDeclaration, handler func(common.ToolContext, json.RawMessage) (string, error)) error {
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
