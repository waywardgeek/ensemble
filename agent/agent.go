// Package agent provides a reusable framework for building LLM-powered coding
// agents. External programs import this package to create agents, register
// custom tools, and run interactive or scripted sessions.
//
// The shared type vocabulary lives in agent/internal/common and is re-exported
// here so that external callers never import internal/ directly.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/waywardgeek/ensemble/agent/internal/common"
	"github.com/waywardgeek/ensemble/agent/internal/jobs"
	"github.com/waywardgeek/ensemble/agent/internal/llm"
	"github.com/waywardgeek/ensemble/agent/internal/mcp"
	"github.com/waywardgeek/ensemble/agent/internal/settings"
	"github.com/waywardgeek/ensemble/agent/internal/skills"
	"github.com/waywardgeek/ensemble/agent/internal/tools"
	"github.com/waywardgeek/ensemble/agent/internal/ws"
)

// Re-export the types external programs need.
type Config = common.Config
type ToolDecl = common.ToolDecl
type Vendor = common.Vendor
type Surface = common.Surface
type Usage = common.Usage

// Skill types.
type SkillProperties = common.SkillProperties
type MCPServerConfig = common.MCPServerConfig
type SkillRegistry = skills.SkillRegistry
type VarRenderer = skills.VarRenderer
type VarRegistry = skills.VarRegistry

// MCP types.
type MCPTransport = mcp.Transport
type MCPClient = mcp.Client

// NewMCPPipeTransport creates a pair of connected in-process transports.
func NewMCPPipeTransport() (client MCPTransport, server MCPTransport) {
	return mcp.NewPipeTransport()
}

// NewMCPStdioTransport spawns a subprocess and talks JSON-RPC over stdin/stdout.
func NewMCPStdioTransport(command string, args []string, env []string) (MCPTransport, error) {
	return mcp.NewStdioTransport(command, args, env)
}

// NewMCPWSTransport creates a transport tunneled over the WebSocket hub.
func NewMCPWSTransport(hub *WSHub) MCPTransport {
	wst := mcp.NewWSTransport(func(data json.RawMessage) {
		hub.BroadcastJSONRPC(data, "")
	})
	hub.SetMCPReceiver("", wst.Deliver)
	return wst
}

// NewMCPRawTransport creates a transport from raw io.Reader/io.Writer.
// Used for --mcp-pipe mode where stdin/stdout become the MCP wire.
func NewMCPRawTransport(r io.Reader, w io.Writer) MCPTransport {
	return mcp.NewRawTransport(r, w)
}

// NewMCPClientWSTransport dials a WebSocket hub and creates a transport
// that tags all messages with the given source. Used by the virtual user
// to connect to the browser's MCP server through the hub.
func NewMCPClientWSTransport(url, source string) (MCPTransport, error) {
	return mcp.NewClientWSTransport(url, source)
}

// Chapter 6: observer and mailbox types.
type Observer = common.Observer
type Observation = common.Observation
type PartDelta = common.PartDelta
type PartFinal = common.PartFinal
type StateChanged = common.StateChanged
type TurnEnded = common.TurnEnded
type AgentID = common.AgentID
type Inbound = common.Inbound
type UserMessage = common.UserMessage
type Hint = common.Hint
type ToolCompleted = common.ToolCompleted
type Interrupt = common.Interrupt
type Mailbox = common.Mailbox
type Media = common.Media
type ModelFeatures = common.ModelFeatures
type TurnState = common.TurnState
type TextPart = common.TextPart

// Chapter 8: GUI observations and pause gate.
type ToolDispatched = common.ToolDispatched
type ToolFinished = common.ToolFinished
type PauseGate = common.PauseGate

// Settings and SettingsStore are re-exported for the WebSocket hub
// constructor and the settings management protocol.
type Settings = common.Settings
type SettingsStore = settings.SettingsStore

// Log is the append-only event log. Exported so the WebSocket hub can read
// it for reconnection without copying.
type Log = common.Log

// ToolCallPart and OpaquePart are re-exported because PartFinal now reports
// EVERY part, not just text. A consumer switching on a final needs the types
// to switch on, and before streaming there was nothing but text to see. The
// internal/ wall is what surfaced this: the ch07 exercise would not compile
// without them, which is the whole reason the exercise lives outside the
// module.
type ToolCallPart = common.ToolCallPart
type OpaquePart = common.OpaquePart

// Streaming. DeltaKind says what sort of content a chunk is; Stream is the
// set of kinds a model can actually deliver incrementally. StreamCallbacks is
// the parse-side half of the seam, exported because a caller embedding this
// framework may want to drive an Engine turn directly.
type DeltaKind = common.DeltaKind
type Stream = common.Stream
type StreamCallbacks = common.StreamCallbacks

const (
	VendorAnthropic = common.VendorAnthropic
	VendorOpenAI    = common.VendorOpenAI
	VendorGemini    = common.VendorGemini
)

// TurnState constants.
const (
	Idle             = common.Idle
	InputPending     = common.InputPending
	InFlight         = common.InFlight
	ToolsPending     = common.ToolsPending
	StateInterrupted = common.Interrupted
)

// Media constants.
const (
	MediaImage    = common.MediaImage
	MediaAudio    = common.MediaAudio
	MediaVideo    = common.MediaVideo
	MediaDocument = common.MediaDocument
)

// DeltaKind constants.
const (
	DeltaThinking = common.DeltaThinking
	DeltaText     = common.DeltaText
	DeltaToolCall = common.DeltaToolCall
)

// Stream constants: the delta kinds a model can deliver incrementally.
const (
	StreamText     = common.StreamText
	StreamThinking = common.StreamThinking
	StreamToolArgs = common.StreamToolArgs
	StreamAll      = common.StreamAll
)

// StreamingFor reports which delta kinds may be streamed for a config.
func StreamingFor(cfg Config) Stream { return common.StreamingFor(cfg) }

// DefaultSurface returns the default API surface for a vendor.
func DefaultSurface(v Vendor) Surface { return common.DefaultSurface(v) }

// SurfaceForModel reports which endpoint dialect a model is spoken to,
// consulting the model table before falling back to the vendor default.
// Prefer it to DefaultSurface: one vendor can serve two surfaces at once.
func SurfaceForModel(model string, v Vendor) Surface {
	return common.SurfaceForModel(model, v)
}

// ToolHandler is the signature for a custom tool: given JSON arguments,
// return the text the model will see.
type ToolHandler func(args json.RawMessage) (string, error)

// Agent is the public handle to a running agent. It implements common.Agent,
// providing the logger that all internal code reaches through parent interfaces.
type Agent struct {
	// UsageCounter holds this run's token totals. Embedded so Agent satisfies
	// the usage half of common.Agent, which is how anything downstream reports
	// spend without being handed a reporter.
	common.UsageCounter

	eng        *llm.Engine
	actor      *llm.Actor
	reg        *tools.Reg
	skills     *skills.SkillRegistry
	vars       *skills.VarRegistry
	Logger     *Logger
	mcpClients []*mcp.Client // active MCP connections for cleanup
}

// NewAgent creates a new agent with the given config and log path. It wires
// up the builtin tools (run_command, read_file, etc.) and any custom tools
// registered before this call. Constructors take interfaces to parents: the
// engine receives the job manager and tool registry as interfaces, not
// concrete types.
func NewAgent(cfg Config, logPath string) *Agent {
	a := &Agent{
		Logger: DefaultLogger(),
		skills: skills.NewSkillRegistry(),
		vars:   skills.NewVarRegistry(),
	}
	j := jobs.NewJobs(a)
	a.reg = tools.NewRegistry()

	// Register built-in variable renderers.
	a.vars.Register("TOOLS", skills.BuiltinToolsRenderer(a.skills, a.reg))
	a.vars.Register("SKILLS", skills.BuiltinSkillsRenderer(a.skills))

	cfg.Tools = a.reg.Declarations()
	a.eng = llm.NewEngine(cfg, logPath, j, a.reg, a)
	a.actor = llm.NewActor(a.eng, a)
	return a
}

// NewBareAgent creates an agent with NO builtin tools. All tools come from
// MCP or explicit RegisterTool calls. Used for auxiliary agents like the
// virtual user that interact only through GUI tools.
func NewBareAgent(cfg Config, logPath string) *Agent {
	a := &Agent{
		Logger: DefaultLogger(),
		skills: skills.NewSkillRegistry(),
		vars:   skills.NewVarRegistry(),
	}
	j := jobs.NewJobs(a)
	a.reg = tools.NewBareRegistry()

	cfg.Tools = a.reg.Declarations()
	a.eng = llm.NewEngine(cfg, logPath, j, a.reg, a)
	a.actor = llm.NewActor(a.eng, a)
	return a
}

// RegisterTool adds a custom tool to this agent's registry. Call this before
// the first Ask. The schema is the JSON Schema for the tool's input parameters.
func (a *Agent) RegisterTool(name, description string, schema json.RawMessage, handler ToolHandler) {
	a.reg.Register(name, description, schema, handler)
	// Update the engine's config so declarations include the new tool.
	a.eng.Cfg.Tools = a.reg.Declarations()
}

// RemoveTool deletes a tool from this agent's registry, including tools bridged in
// from an MCP server. Denying a capability means removing the tool rather than
// instructing the model to avoid it: an observer that still holds a channel will
// use it, and a task completed through the wrong channel tests the wrong channel.
func (a *Agent) RemoveTool(name string) {
	a.reg.RemoveTool(name)
	a.eng.Cfg.Tools = a.reg.Declarations()
}

// RegisterVar adds a custom variable renderer for $VAR substitution in skill
// bodies. Call this before DiscoverSkills / LoadSkill so variables are available
// at render time. Built-in renderers ($TOOLS, $SKILLS) are registered automatically.
func (a *Agent) RegisterVar(name string, renderer VarRenderer) {
	a.vars.Register(name, renderer)
}

// DiscoverSkills scans a directory for skill subdirectories (each containing
// a SKILL.md). Skills are registered as Available but not loaded.
func (a *Agent) DiscoverSkills(dir string) error {
	return a.skills.DiscoverSkills(dir)
}

// LoadSkill loads a skill by name, marking it as initial (loaded at creation).
// Resolves depends: chain. Call after DiscoverSkills and before the first Ask
// to set up the primary skill and any initial loadable skills. The rendered
// body is appended to cfg.SystemPrompt.
func (a *Agent) LoadSkill(name string) error {
	if err := a.skills.LoadInitial(name, a.vars); err != nil {
		return err
	}
	// Rebuild system prompt from initial skill bodies.
	bodies := a.skills.InitialBodies()
	if len(bodies) > 0 {
		a.eng.Cfg.SystemPrompt = strings.Join(bodies, "\n\n---\n\n")
	}
	// Rebuild tool declarations with skill filtering.
	a.eng.Cfg.Tools = a.reg.Declarations()
	return nil
}

// WireSkillTools registers the load_skill and unload_skill tools. Call this
// after DiscoverSkills and initial LoadSkill calls, before the first Ask.
func (a *Agent) WireSkillTools() {
	a.reg.WireSkills(a.skills, a.vars, a.eng.Log)
	a.eng.Cfg.Tools = a.reg.Declarations()
	// When a skill is loaded dynamically, update the config's tool declarations.
	a.reg.SetOnToolsChanged(func() {
		a.eng.Cfg.Tools = a.reg.Declarations()
	})
}

// Logf logs a message through the agent's logger. This makes Agent satisfy
// common.Agent, so any code that holds a Agent can log.
func (a *Agent) Logf(format string, args ...any)    { a.Logger.Logf(format, args...) }
func (a *Agent) APILogf(format string, args ...any) { a.Logger.APILogf(format, args...) }
func (a *Agent) Debugf(format string, args ...any)  { a.Logger.Debugf(format, args...) }

// Ask sends a prompt and runs the full tool loop until the model replies.
func (a *Agent) Ask(prompt string) (string, error) {
	return a.actor.Ask(prompt)
}

// Shutdown ends all running jobs, closes MCP connections, and saves the log.
func (a *Agent) Shutdown() error {
	err := a.actor.Shutdown()
	for _, c := range a.mcpClients {
		c.Close()
	}
	a.mcpClients = nil
	return err
}

// ConnectMCP connects to an MCP server over the given transport.
// It performs the MCP handshake, discovers tools, bridges them into the
// agent's registry, and sets up reverse tool handling.
func (a *Agent) ConnectMCP(t mcp.Transport) error {
	client := mcp.NewClient(t)

	if err := client.Initialize(context.Background()); err != nil {
		client.Close()
		return fmt.Errorf("mcp connect: %w", err)
	}

	tools, err := client.ListTools(context.Background())
	if err != nil {
		client.Close()
		return fmt.Errorf("mcp connect: %w", err)
	}

	// Bridge discovered tools into the agent's registry.
	bridged := mcp.Bridge(client, tools)
	for _, t := range bridged {
		a.reg.RegisterTool(t)
	}

	// Set up reverse tool handler: MCP server can call agent tools.
	client.SetReverseHandler(func(name string, args json.RawMessage) (string, error) {
		tool, err := a.reg.Lookup(name)
		if err != nil {
			return "", err
		}
		c := &common.Call{Agent: a, Jobs: a.eng.Jobs}
		return tool.Run(c, args)
	})

	a.mcpClients = append(a.mcpClients, client)

	// Update tool declarations for the engine.
	a.eng.Cfg.Tools = a.reg.Declarations()
	return nil
}

// Usage returns the token counts accumulated across all requests.
func (a *Agent) Usage() Usage {
	return a.eng.Ctx.Usage
}

// EventLog returns the append-only event log. The hub reads this for
// reconnection; callers must not mutate the returned log.
func (a *Agent) EventLog() *Log {
	return a.eng.Log
}

// ----------------------------------------------------------------
// Chapter 6: Actor-based API
// ----------------------------------------------------------------

// Actor is the public handle to an actor-based agent.
type Actor = llm.Actor

// ToolRoundLimitError reports that a turn requested more tool batches than
// permitted. Rejected calls have paired not-executed results in the history.
type ToolRoundLimitError = llm.ToolRoundLimitError

// ErrActorStopped is returned when a request cannot run because its actor stopped.
const ErrActorStopped = llm.ErrActorStopped

// Framework manages multiple actors.
type Framework = llm.Framework

// NewActor returns this agent's single execution owner.
func (a *Agent) NewActor() *Actor {
	return a.actor
}

// NewFramework creates a multi-agent framework.
func NewFramework(host common.Agent) *Framework {
	return llm.NewFramework(host)
}

// NewMailbox creates a new mailbox.
func NewMailbox() *Mailbox {
	return common.NewMailbox()
}

// NewPauseGate creates an unpaused PauseGate.
func NewPauseGate() *PauseGate {
	return common.NewPauseGate()
}

// NewSkillRegistry creates an empty skill registry.
func NewSkillRegistry() *SkillRegistry {
	return skills.NewSkillRegistry()
}

// NewVarRegistry creates an empty variable registry.
func NewVarRegistry() *VarRegistry {
	return skills.NewVarRegistry()
}

// NewSettingsStore creates a settings store. If path is non-empty and the
// file exists, settings are loaded from it.
func NewSettingsStore(path string) *SettingsStore {
	return settings.NewSettingsStore(path)
}

// LookupModel returns model features.
func LookupModel(model string) (ModelFeatures, bool) {
	return common.LookupModel(model)
}

// ConfigFromEnv builds a Config from environment variables (LLM_VENDOR,
// LLM_MODEL, LLM_API_KEY, LLM_BASE_URL). This is the standard way for an
// external program to configure the framework without hardcoding vendor
// details.
func ConfigFromEnv() (Config, error) {
	vendor, err := parseVendor(envOr("LLM_VENDOR", "anthropic"))
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Vendor:    vendor,
		MaxTokens: 1024,
	}
	cfg.Endpoints = common.ResolveEndpoints(pick)
	active := cfg.Endpoints[vendor]
	cfg.BaseURL, cfg.APIKey = active.BaseURL, active.APIKey
	switch vendor {
	case VendorAnthropic:
		cfg.Model = pick("LLM_MODEL", "ANTHROPIC_MODEL", "claude-sonnet-5")
	case VendorOpenAI:
		cfg.Model = pick("LLM_MODEL", "OPENAI_MODEL", "gpt-6.1-sol")
	case VendorGemini:
		cfg.Model = pick("LLM_MODEL", "GEMINI_MODEL", "gemini-3.8-flash")
	}

	// After the model is known, never before: the surface is a property of
	// the model, and resolving it from the vendor alone would put every
	// OpenAI model on whichever endpoint the newest one happens to use.
	cfg.Surface = common.SurfaceForModel(cfg.Model, vendor)
	return cfg, nil
}

func parseVendor(s string) (Vendor, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "anthropic", "claude":
		return VendorAnthropic, nil
	case "openai":
		return VendorOpenAI, nil
	case "gemini":
		return VendorGemini, nil
	}
	return 0, fmt.Errorf("unknown vendor %q (want anthropic, openai or gemini)", s)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func pick(primary, secondary, def string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return envOr(secondary, def)
}

// ----------------------------------------------------------------
// Chapter 8: WebSocket hub
// ----------------------------------------------------------------

// WSHub is the WebSocket fan-out hub. It implements Observer.
type WSHub = ws.Hub

// NewWSHub creates a hub that fans observations out to WebSocket clients.
// send is called for every prompt/hint/interrupt from a browser; gate
// controls tool-dispatch pausing; guiLogPath is the path to gui.log;
// eventLog provides read access to the append-only event log for
// reconnection; settings provides the GUI-editable settings store (may be nil).
//
// The session usage meter is passed as nil here deliberately. This signature
// is chapter 8's published API, and the meter arrived long afterwards; the
// server in cmd/ calls ws.NewHub directly to supply one. Growing this
// function's arity would rewrite a chapter's contract for a later feature.
func NewWSHub(gate *PauseGate, send func(Inbound), guiLogPath string, eventLog *Log, settings *SettingsStore) *WSHub {
	return ws.NewHub(gate, send, guiLogPath, eventLog, settings, nil)
}

// ServeHTTP starts an HTTP server that serves static files from staticDir
// and upgrades /ws to a WebSocket connection handled by hub.
func ServeHTTP(addr string, staticDir string, hub *WSHub) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	mux.HandleFunc("/ws", hub.ServeWS)
	srv := &http.Server{Addr: addr, Handler: mux}
	go srv.ListenAndServe()
	return srv
}
