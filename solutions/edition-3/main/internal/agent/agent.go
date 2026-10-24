// Package agent owns turn orchestration and configuration, not HTTP or replay.
package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"ensemble/internal/common"
	"ensemble/internal/sandbox"
)

type agent struct {
	id            string
	parent        common.Ensemble
	config        common.Config
	configMu      sync.RWMutex
	saveMu        sync.Mutex
	recovery      common.Recovery
	maxToolRounds atomic.Int64
	engine        common.Engine
	history       common.History
	tools         common.ToolRegistry
	skills        common.SkillRegistry
	mcp           common.MCPManager
	jobs          common.Jobs
	sandbox       common.Sandbox
	box           mailbox
	done          chan struct{}
	shutdownErr   error
	liveCalls     map[string]bool
	jobParts      map[int]uint64
	workers       sync.WaitGroup
	observersMu   sync.Mutex
	observers     map[uint64]chan common.Observation
	queues        map[*observationQueue]struct{}
	nextObserver  uint64
	gateMu        sync.Mutex
	paused        bool
	// The fields below are accessed only by the actor.
	started            atomic.Bool
	handoffText        string
	handoffSavePending bool
	active             *turn
	pending            []*request
	nextTurn           uint64
	partID             uint64
	viewValue          atomic.Value
	viewState          common.ViewSnapshot
	viewParts          map[uint64]common.PartDelta
	state              common.TurnState
	terminal           *terminalCompletion
	stopping           bool
	speaking           bool
	gate               func()
	gateReason         string
}

// New validates configuration before any request and retains the actual owner.
// The composition root attaches Engine before publishing this Agent.
func New(parent common.Ensemble, config common.Config) (*agent, error) {
	if parent == nil {
		return nil, errors.New("Agent requires an Ensemble parent")
	}
	if config.ContextTarget > 0 {
		config.ContextTarget = effectiveTarget(config.ContextTarget)
	}
	if config.ContextTarget < 0 {
		return nil, fmt.Errorf("negative context_target")
	}
	if strings.TrimSpace(config.Model) == "" {
		return nil, errors.New("ANTHROPIC_MODEL is not set; ask GET /v1/models which models exist")
	}

	if err := common.NormalizeOpenAI(&config); err != nil {
		return nil, err
	}
	if config.Vendor == 0 {
		config.Vendor = common.Anthropic
	}
	switch config.Vendor {
	case common.Anthropic:
		if config.Surface == 0 {
			config.Surface = common.Messages
		}
		if config.BaseURL == "" {
			config.BaseURL = "https://api.anthropic.com"
		}
	case common.OpenAI:
		if config.Surface == 0 {
			config.Surface = common.ChatCompletions
		}
		if config.BaseURL == "" {
			config.BaseURL = "https://api.openai.com"
		}
	case common.Gemini:
		if config.Surface == 0 {
			config.Surface = common.GenerateContent
		}
		if config.BaseURL == "" {
			config.BaseURL = "https://generativelanguage.googleapis.com"
		}
	default:
		return nil, errors.New("unknown vendor")
	}
	if (config.Vendor == common.Anthropic && config.Surface != common.Messages) || (config.Vendor == common.OpenAI && config.Surface != common.ChatCompletions && config.Surface != common.Responses) || (config.Vendor == common.Gemini && config.Surface != common.GenerateContent) {
		return nil, errors.New("unsupported vendor surface")
	}
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")
	u, err := url.Parse(config.BaseURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		// Never echo a malformed URL: it might contain embedded credentials.
		return nil, errors.New("invalid ANTHROPIC_BASE_URL")
	}
	config.MCPServers = cloneMCPServers(config.MCPServers)
	config.Skills = cloneSkillSpec(config.Skills)
	config.StartupSkills = append([]string(nil), config.StartupSkills...)
	if config.Tools != nil {
		config.Tools = append([]string{}, config.Tools...)
	}
	if config.BaseDir == "" && config.SandboxRoot != "" {
		config.BaseDir = config.SandboxRoot
	}
	if config.BaseDir == "" {
		config.BaseDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	config.BaseDir, err = filepath.Abs(config.BaseDir)
	if err != nil {
		return nil, err
	}
	if config.DiagnosticMaxRecords == nil {
		n := 200
		config.DiagnosticMaxRecords = &n
	} else {
		n := *config.DiagnosticMaxRecords
		if n < 0 {
			return nil, errors.New("negative diagnostic_max_records")
		}
		config.DiagnosticMaxRecords = &n
	}
	if config.LogMode == "" {
		config.LogMode = common.LogFull
	}
	if config.LogMode != common.LogFull && config.LogMode != common.LogDialogueOnly {
		return nil, errors.New("invalid log mode")
	}
	if config.SavePath != "" && config.SavePath == config.LogPath {
		return nil, errors.New("save and legacy log paths must differ")
	}
	if config.MaxToolRounds < 0 {
		return nil, errors.New("MaxToolRounds must be nonnegative")
	}
	if config.MaxToolRounds == 0 {
		config.MaxToolRounds = common.DefaultMaxToolRounds
	}
	config.MCPServers = cloneMCPServers(config.MCPServers)
	config.Skills = cloneSkillSpec(config.Skills)
	config.StartupSkills = append([]string(nil), config.StartupSkills...)
	if config.Tools != nil {
		config.Tools = append([]string{}, config.Tools...)
	}
	a := &agent{parent: parent, config: config, box: mailbox{wake: make(chan struct{}, 1)}, done: make(chan struct{}), observers: make(map[uint64]chan common.Observation), queues: make(map[*observationQueue]struct{}), state: common.Idle, paused: config.PauseOnStart, speaking: config.Speaking}
	constructed := false
	defer func() {
		if !constructed && a.sandbox != nil {
			_ = a.sandbox.Close()
		}
	}()
	a.sandbox, err = sandbox.New(a)
	if err != nil {
		return nil, err
	}
	fail := func(err error) (*agent, error) { return nil, err }
	if a.sandbox.Root() != "" {
		config.SandboxRoot = a.sandbox.Root()
		config.BaseDir, err = a.sandbox.ValidatePath(config.BaseDir)
		if err != nil {
			return fail(fmt.Errorf("base directory: %w", err))
		}
	}
	a.config = config
	if config.DataDir == "" {
		if config.SandboxRoot != "" {
			config.DataDir = filepath.Join(config.SandboxRoot, ".ensemble")
		} else {
			config.DataDir, err = os.MkdirTemp("", "ensemble-agent-")
			if err != nil {
				return fail(err)
			}
		}
	}
	config.DataDir, err = a.sandbox.ValidatePath(config.DataDir)
	if err != nil {
		return fail(fmt.Errorf("agent data directory: %w", err))
	}
	if err = a.sandbox.MkdirAll(config.DataDir, 0700); err != nil {
		return fail(err)
	}
	if config.SavePath == "" {
		config.SavePath = filepath.Join(config.DataDir, "save.json")
	} else {
		// Explicit recovery paths belong to the caller's cwd, unlike relative
		// DataDir defaults. Resolve that choice before rooted validation: an
		// outside caller-selected path must be refused, never moved into BaseDir.
		var err error
		config.SavePath, err = filepath.Abs(config.SavePath)
		if err != nil {
			return nil, err
		}
	}
	for _, p := range []*string{&config.SavePath, &config.LogPath} {
		if *p != "" {
			*p, err = a.sandbox.ValidatePath(*p)
			if err != nil {
				return fail(fmt.Errorf("agent state path: %w", err))
			}
		}
	}
	if config.SavePath == config.LogPath {
		return fail(errors.New("save and legacy log paths must differ"))
	}
	a.config = config
	a.maxToolRounds.Store(int64(config.MaxToolRounds))
	a.viewState = common.ViewSnapshot{State: common.Idle, Paused: config.PauseOnStart, Speaking: config.Speaking}
	a.viewParts = make(map[uint64]common.PartDelta)
	a.jobParts = make(map[int]uint64)
	a.viewValue.Store(a.viewState)
	constructed = true
	return a, nil
}

// SetEngine completes the root's one-time wiring; it cannot replace an owner.
func (a *agent) SetEngine(engine common.Engine) {
	if engine == nil || a.engine != nil || engine.Parent() != a {
		panic("invalid Agent/Engine wiring")
	}
	a.engine = engine
}

// Parent exposes the actual Ensemble rather than a copied service bundle.
func (a *agent) Parent() common.Ensemble { return a.parent }
func (a *agent) ID() string              { return a.id }

// SetID is one-time root composition, before the actor starts.
func (a *agent) SetID(id string) {
	if a.id != "" || id == "" {
		panic("invalid Agent identity")
	}
	a.id = id
}

// Config returns a value copy of the Agent-owned configuration.
func (a *agent) Recovery() common.Recovery { return a.recovery }
func (a *agent) Config() common.Config {
	a.configMu.RLock()
	defer a.configMu.RUnlock()
	c := a.config
	c.SubmitSchema = append(json.RawMessage(nil), c.SubmitSchema...)
	if c.DiagnosticMaxRecords != nil {
		n := *c.DiagnosticMaxRecords
		c.DiagnosticMaxRecords = &n
	}
	c.MaxToolRounds = int(a.maxToolRounds.Load())
	c.MCPServers = cloneMCPServers(c.MCPServers)
	c.Skills = cloneSkillSpec(c.Skills)
	c.StartupSkills = append([]string(nil), c.StartupSkills...)
	if c.Tools != nil {
		c.Tools = append([]string{}, c.Tools...)
	}
	return c
}

// SetMaxToolRounds synchronously publishes next-turn policy; zero selects
// common.DefaultMaxToolRounds (200). The actor captures this once at admission;
// a later edit cannot move an active turn's boundary.
func (a *agent) SetMaxToolRounds(rounds int) error {
	if rounds < 0 {
		return errors.New("MaxToolRounds must not be negative")
	}
	a.maxToolRounds.Store(int64(rounds))
	return nil
}

func (a *agent) SetTools(t common.ToolRegistry) {
	if t == nil || a.tools != nil || t.Parent() != a {
		panic("invalid Agent/Tools wiring")
	}
	a.tools = t
}
func (a *agent) Skills() common.SkillRegistry { return a.skills }
func (a *agent) SetSkills(s common.SkillRegistry) {
	if s == nil || s.Parent() != a {
		panic("invalid skill registry attachment")
	}
	a.skills = s
}
func (a *agent) RegisterVar(name string, f func() string) { a.skills.RegisterVar(name, f) }
func (a *agent) Tools() common.ToolRegistry               { return a.tools }
func (a *agent) SetJobs(j common.Jobs) {
	if j == nil || a.jobs != nil || j.Parent() != a {
		panic("invalid Agent/Jobs wiring")
	}
	a.jobs = j
}
func (a *agent) Jobs() common.Jobs { return a.jobs }

// Engine exposes the owned transport and its parent relationship.
func (a *agent) Engine() common.Engine { return a.engine }

// SetHistory completes the other child edge at the composition root.
func (a *agent) SetHistory(h common.History) {
	if h == nil || a.history != nil || h.Parent() != a {
		panic("invalid Agent/History wiring")
	}
	a.history = h
}
func (a *agent) Log() common.History     { return a.history }
func (a *agent) Render() ([]byte, error) { return a.engine.Render(a.history.Context()) }
func (a *agent) ParseResponse(status int, b []byte) ([]common.Event, error) {
	return a.engine.Parse(&http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(b))}, common.ParseCallbacks{})
}

// History preserves Chapter 1's completed-text view, derived from the log.
func (a *agent) History() []common.Message {
	var messages, pending []common.Message
	events := a.history.Events()
	for i, e := range events {
		if e.Stale {
			continue
		}
		switch e.Type {
		case common.MessageReceived:
			if e.Message.Actor == common.Human {
				var b strings.Builder
				for _, p := range e.Message.Parts {
					if p.Type == "text" {
						b.WriteString(p.Text)
					}
				}
				pending = append(pending, common.Message{Role: "user", Content: b.String()})
			}
		case common.ResponseEnded:
			if i+1 < len(events) && events[i+1].Type == common.ErrorOccurred {
				pending = nil
				continue
			}
			hasCalls := false
			for _, p := range e.Response.Parts {
				if p.Type == "tool_call" {
					hasCalls = true
				}
			}
			if hasCalls {
				continue
			}
			var b strings.Builder
			for _, p := range e.Response.Parts {
				if p.Type == "text" {
					b.WriteString(p.Text)
				}
			}
			if strings.TrimSpace(b.String()) != "" {
				messages = append(messages, pending...)
				messages = append(messages, common.Message{Role: "assistant", Content: b.String()})
			}
			pending = nil
		case common.ErrorOccurred:
			pending = nil
		}
	}
	return messages
}
func (a *agent) Usage() common.Usage { return a.engine.Usage() }

func (a *agent) identifyCalls(events []common.Event) error {
	used := map[string]bool{}
	prior := a.history.Events()
	for _, entry := range a.history.Context().Dialogue {
		for _, p := range entry.Parts {
			if p.Type == "tool_call" && p.CallID != "" {
				used[p.CallID] = true
			}
		}
	}
	for _, e := range prior {
		if e.Response != nil {
			for _, p := range e.Response.Parts {
				if p.Type == "tool_call" && p.CallID != "" {
					used[p.CallID] = true
				}
			}
		}
	}
	// Reserve supplied IDs before synthesizing missing ones, preserving provenance
	// and all provider IDs. Duplicate supplied IDs cannot safely be dispatched.
	for _, e := range events {
		if e.Type != common.ResponseEnded {
			continue
		}
		for _, p := range e.Response.Parts {
			if p.Type != "tool_call" || p.CallID == "" {
				continue
			}
			if used[p.CallID] {
				return fmt.Errorf("duplicate tool call ID")
			}
			used[p.CallID] = true
		}
	}
	for i, e := range events {
		if e.Type != common.ResponseEnded {
			continue
		}
		for j := range e.Response.Parts {
			p := &e.Response.Parts[j]
			if p.Type != "tool_call" || p.CallID != "" {
				continue
			}
			base := fmt.Sprintf("live_call_%d_%d_%d", len(prior), i, j)
			id := base
			for suffix := 1; used[id]; suffix++ {
				id = fmt.Sprintf("%s_%d", base, suffix)
			}
			p.CallID = id
			used[id] = true
		}
	}
	return nil
}

func cloneSkillSpec(s *common.SkillSpec) *common.SkillSpec {
	if s == nil {
		return nil
	}
	out := *s
	out.Variables = map[string]func() string{}
	for k, v := range s.Variables {
		out.Variables[k] = v
	}
	return &out
}

func cloneMCPServers(in []common.MCPServer) []common.MCPServer {
	out := append([]common.MCPServer(nil), in...)
	for i := range out {
		out[i].Args = append([]string(nil), out[i].Args...)
	}
	return out
}

func (a *agent) SetMCP(m common.MCPManager) {
	if m == nil || a.mcp != nil {
		panic("invalid MCP attachment")
	}
	a.mcp = m
}
func (a *agent) MCP() common.MCPManager { return a.mcp }

func (a *agent) RegisterExternalTool(d common.ToolDefinition, h common.ToolHandler) error {
	return a.tools.Register(d, h)
}
func (a *agent) AttachMCPConnection(name string, t common.MCPTransport) error {
	return a.mcp.Attach(name, t)
}
func (a *agent) ServeMCPConnection(ctx context.Context, t common.MCPTransport) error {
	return a.mcp.Serve(ctx, t)
}

func (a *agent) Sandbox() common.Sandbox { return a.sandbox }
