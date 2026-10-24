package common

import (
	"context"
	"encoding/json"
	"time"
)

const RefHandle = 3

type JobStatus string

const (
	JobRunning      JobStatus = "running"
	JobDone         JobStatus = "done"
	JobKilledStatus JobStatus = "killed"
)

// JobData is a snapshot, never the live runtime record.
type JobData struct {
	CallPartID uint64          `json:"call_part_id,omitempty"`
	TurnID     uint64          `json:"turn_id,omitempty"`
	CallID     string          `json:"call_id,omitempty"`
	Name       string          `json:"name,omitempty"`
	Args       json.RawMessage `json:"args,omitempty"`
	Handle     int             `json:"handle"`
	Status     JobStatus       `json:"status"`
	Output     Ref             `json:"output"`
	Bytes      int64           `json:"bytes"`
	ExitCode   *int            `json:"exit_code,omitempty"`
	Cwd        string          `json:"cwd,omitempty"`
	Reason     string          `json:"reason,omitempty"`
}

type ToolLimits struct {
	Delay          time.Duration
	Pattern        string
	MaxOutputBytes int
}

func DefaultToolLimits() ToolLimits {
	return ToolLimits{Delay: 3 * time.Second, MaxOutputBytes: 16 * 1024}
}

type JobReport struct {
	Job     JobData
	Text    string
	IsError bool
}

type Job interface {
	Parent() Jobs
	Data() JobData
	// Start releases a reservation; Abort settles one without handler effects.
	Start() bool
	// StartAfter holds reservation admission through a save/publication boundary.
	StartAfter(func() error) (bool, error)
	Abort(error)
	InitialReport(context.Context, ToolLimits) (JobReport, error)
	Complete(context.Context) (JobReport, error)
	Wait(context.Context, ToolLimits) (JobReport, error)
	SendInput(context.Context, string, ToolLimits) (JobReport, error)
	Kill(string) error
	Shutdown() error
}

// Jobs belongs to one Agent. The Ensemble's table allocates all handles.
type Jobs interface {
	Parent() Agent
	Start(ctx context.Context, handler ToolHandler, args []byte, cwd string, origin JobData) (Job, error)
	Reserve(ctx context.Context, handler ToolHandler, args []byte, cwd string, origin JobData) (Job, error)
	Get(int) (Job, error)
	Shutdown() error
}
