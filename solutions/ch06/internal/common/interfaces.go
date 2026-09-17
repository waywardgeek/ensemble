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
	WokeDone    WakeReason = iota + 1
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
	SendInput(text string) error
	Bytes() int
	Err() error
	Finish(result string, err error)
}

// JobManager manages the set of running jobs. It embeds Host so that job
// supervision code can log through the parent chain.
type JobManager interface {
	Host
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
}

// ---------------------------------------------------------------------------
// Tool types shared between engine and tool implementations.
// ---------------------------------------------------------------------------

// Host is the root interface every object can reach through its parent chain.
// It provides access to the logger and any other top-level facilities.
type Host interface {
	Logf(format string, args ...any)
}

// ToolFunc executes one tool call and returns text the model will see.
type ToolFunc func(c *Call, args json.RawMessage) (string, error)

// Tool is a named, documented function the model can ask for.
type Tool struct {
	Name        string
	Description string
	Schema      json.RawMessage
	Run         ToolFunc
	NoJob       bool
}

// Call is what a tool is handed besides its arguments. The embedded Host
// gives every tool trivial access to the logger through the parent chain.
type Call struct {
	Host
	Job    JobHandle
	Jobs   JobManager
	Limits Limits
	Events []Event
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
