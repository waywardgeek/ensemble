package common

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"
)

// ---------------------------------------------------------------------------
// Job-related types shared across packages.
// ---------------------------------------------------------------------------

// JobStatus is the whole life of a job.
type JobStatus int

const (
	StatusRunning JobStatus = iota + 1
	StatusDone
	StatusKilled
)

func (s JobStatus) String() string {
	switch s {
	case StatusRunning:
		return "running"
	case StatusDone:
		return "done"
	case StatusKilled:
		return "killed"
	}
	return fmt.Sprintf("JobStatus(%d)", int(s))
}

func (s JobStatus) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }

func (s *JobStatus) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch str {
	case "running":
		*s = StatusRunning
	case "done":
		*s = StatusDone
	case "killed":
		*s = StatusKilled
	default:
		return fmt.Errorf("unknown job status %q", str)
	}
	return nil
}

// Limits controls how the dispatcher waits on a job.
type Limits struct {
	Delay     time.Duration
	Pattern   *regexp.Regexp
	MaxOutput int
}

const (
	DefaultDelay     = 3 * time.Second
	DefaultMaxOutput = 16 * 1024
)

func DefaultLimits() Limits { return Limits{Delay: DefaultDelay, MaxOutput: DefaultMaxOutput} }

func (l Limits) String() string {
	s := fmt.Sprintf("ai_callback_delay %s, max_output_bytes %d", l.Delay, l.MaxOutput)
	if l.Pattern != nil {
		s += fmt.Sprintf(", ai_callback_pattern %q", l.Pattern.String())
	}
	return s
}

// WakeReason reports WHY a wait ended.
type WakeReason int

const (
	WokeDone WakeReason = iota + 1
	WokeDelay
	WokePattern
)

// ---------------------------------------------------------------------------
// Interfaces that packages use to reach each other through common.
// ---------------------------------------------------------------------------

// JobHandle is what a running tool sees of its own job.
type JobHandle interface {
	io.Writer
	Attach(proc *os.Process, stdin interface{ Write([]byte) (int, error) })
	SetExit(code int)
	SetCwd(dir string)
	Status() JobStatus
	Data() *JobData
	HasProcess() bool
	Wait(l Limits) WakeReason
	Report(reason WakeReason, l Limits) string
	Kill(reason string) bool
	// StopProcess kills an attached process (or one attached later), without
	// discarding a non-process tool's eventual return value during shutdown.
	StopProcess(reason string)
	SendInput(text string) error
	Bytes() int
	Err() error
	Finish(result string, err error)
}

// JobManager manages the set of running jobs. It embeds Agent so that job
// supervision code can log through the parent chain.
type JobManager interface {
	Agent
	Start(tool, callID string) (JobHandle, error)
	Get(h int) (JobHandle, bool)
	Handles() []int
	Running() []JobHandle
	SetNext(l Limits)
	Take(args json.RawMessage) (Limits, bool, error)
}

// ToolRegistry is what the engine uses to dispatch tool calls.
type ToolRegistry interface {
	Lookup(name string) (Tool, error)
	Declarations() []ToolDecl
	EphemeralTools(mode string) []Tool

	// IsEnabled reports whether a tool may be called right now. Lookup
	// answers "does this exist"; a tool can exist and still be off limits.
	// A bridged MCP server registers every tool it advertises, but only the
	// ones its SKILL.md lists are enabled, and dispatch checks this before
	// running anything the model asked for by name.
	IsEnabled(name string) bool

	// SyncModelGatedTools declares or withdraws the tools whose meaning
	// depends on the model in use, so that switching models leaves the
	// advertised tool set honest. Called at startup and on every switch.
	SyncModelGatedTools(model string)
}

// ---------------------------------------------------------------------------
// Tool types shared between engine and tool implementations.
// ---------------------------------------------------------------------------

// Agent is the root interface every object can reach through its parent chain.
// It provides access to the logger and any other top-level facilities.
// Agent is the back-pointer interface every object in the library holds to
// the agent that created it, directly or through its own parent.
//
// It was called Host until Chapter 22, named for its first capability —
// logging — rather than for the object it points at. The name mattered more
// than it looks. Nobody thinks to add Model() to "the logging thing", so when
// a later object needed the model it got a closure stapled to it at the
// wiring site instead, and the chain stopped growing the day it was named.
//
// Usage reporting used to live here too, with a comment explaining that it
// belonged on Host "because Host is already the thing every object can
// reach". That is a reason for where it was EASY to put, not for where it
// BELONGED, and it is the same mistake in a quieter register: reachable had
// become the selection criterion, so the parent interface accreted a
// capability that has nothing to do with being a parent. Token counts now
// live on Engine, which is the object that spends them and the only one that
// knows which model did.
type Agent interface {
	Logf(format string, args ...any)
	// APILogf logs LLM API wire traffic (JSON requests and responses).
	APILogf(format string, args ...any)
	// Debugf logs to both the terminal and the debug log file.
	Debugf(format string, args ...any)
}

// Engine is the back-pointer interface for the object that runs model
// requests: it owns the configuration, spends the tokens, and knows which
// model spent them.
//
// It exists because the engine was always reachable from the tool dispatch
// site and never exposed. The line that builds a Call reached into the engine
// to pull out one field and discarded the rest, so a tool that needed the
// model name had no route to it even though the engine was sitting in scope
// one identifier away. Everything added here was already available at the
// moment the call was constructed.
//
// Agent() is the parent accessor. An object walks UP the chain — Engine to
// Agent — rather than holding a separate direct line to the root, because two
// routes to the same object drift, and the one that drifts is always the one
// missing the capability you need.
type Engine interface {
	// Agent returns the engine's parent.
	Agent() Agent

	// Model is the model the engine will send the next request to.
	Model() string

	// Pricing is the price table for the current model. Prices are returned
	// rather than costs: money is derived at the point of display, never
	// stored, because a stored cost is wrong the day a price changes.
	Pricing() Pricing

	// Usage reports token counts, including the per-model split needed to
	// price a session that switched models.
	Usage() UsageSource
}

// ToolFunc executes one tool call and returns text the model will see.
type ToolFunc func(c *Call, args json.RawMessage) (string, error)

// Tool is a named, documented function the model can ask for.
//
// Ephemeral is "round", "turn", or "" — matching the ToolDecl field. It
// controls whether the engine auto-calls this tool and routes its output to
// context.Ephemera instead of dialogue.
type Tool struct {
	Name        string
	Description string
	Schema      json.RawMessage
	Run         ToolFunc
	NoJob       bool
	Ephemeral   string // "round", "turn", or ""
}

// Call is what a tool is handed besides its arguments. The embedded Agent
// gives every tool trivial access to the logger through the parent chain.
type Call struct {
	Agent
	// Engine is the back-pointer to the object running the model request
	// this tool call belongs to. A tool reaches the agent through it with
	// Engine.Agent(), rather than through a second field pointing at the
	// root: one route, so there is one place to extend.
	//
	// Nothing new had to be plumbed to make this available. The engine was
	// already in scope at the line that builds a Call — that line reached
	// into it to pull out Jobs and threw the rest away.
	Engine Engine
	Job    JobHandle
	Jobs   JobManager
	Limits Limits
	Events []Event
	// DeferFinish tells the dispatcher not to call Job.Finish when the
	// tool returns — the tool has spawned a background goroutine that
	// will call Finish itself (e.g. an interactive PTY reader). When
	// false (the default), the dispatcher calls Finish as usual.
	DeferFinish bool
}

// ---------------------------------------------------------------------------
// Limit wire format and helpers (methods on Limits must live here).
// ---------------------------------------------------------------------------

// LimitArgs is the wire spelling of Limits. Every tool that waits on a job
// declares these three in its schema; `tool_limits` accepts them for the
// tools that do not.
type LimitArgs struct {
	Delay     *float64 `json:"ai_callback_delay"`
	Pattern   *string  `json:"ai_callback_pattern"`
	MaxOutput *int     `json:"max_output_bytes"`
}

// Overlay applies whichever of the three the arguments set.
func (l Limits) Overlay(a LimitArgs) (Limits, error) {
	if a.Delay != nil {
		if *a.Delay < 0 {
			return l, fmt.Errorf("ai_callback_delay must be >= 0, got %v", *a.Delay)
		}
		l.Delay = time.Duration(*a.Delay * float64(time.Second))
	}
	if a.Pattern != nil {
		if *a.Pattern == "" {
			l.Pattern = nil
		} else {
			re, err := regexp.Compile(*a.Pattern)
			if err != nil {
				return l, fmt.Errorf("ai_callback_pattern: bad pattern %q: %v", *a.Pattern, err)
			}
			l.Pattern = re
		}
	}
	if a.MaxOutput != nil {
		if *a.MaxOutput <= 0 {
			return l, fmt.Errorf("max_output_bytes must be > 0, got %d", *a.MaxOutput)
		}
		l.MaxOutput = *a.MaxOutput
	}
	return l, nil
}

// LimitsInArgs peeks at a call's arguments for the three limit keys.
func LimitsInArgs(args json.RawMessage) (LimitArgs, error) {
	var a LimitArgs
	if len(args) == 0 {
		return a, nil
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return LimitArgs{}, nil
	}
	return a, nil
}

// ---------------------------------------------------------------------------
// JobData: the snapshot a job exposes for reporting.
// (Defined in event.go with the rest of the event vocabulary.)
// ---------------------------------------------------------------------------
