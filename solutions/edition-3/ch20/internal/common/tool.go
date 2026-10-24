package common

import (
	"context"
	"encoding/json"
	"io"
)

// ToolDefinition is a wire-neutral declaration, not a handler or runtime state.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// ToolRegistry owns one Agent's built-ins. Dispatch returns a result even on failure.
// Every ordinary dispatch uses the Agent-owned Jobs collection.
type ToolRegistry interface {
	// Invoke is the compatibility recording path, admitted by the Agent actor.
	Invoke(context.Context, Part, string) error
	Parent() Agent
	Catalog() []ToolSpec
	Snapshot() CapabilitySnapshot
	Definitions() []ToolDefinition
	Collect(context.Context, string) (PartList, error)
	Dispatch(context.Context, string, json.RawMessage) (PartList, bool)
	// Prepare reserves an ordinary job; Await obtains actor admission if needed.
	Prepare(context.Context, Part, string, uint64) ToolOperation
	Register(ToolDefinition, ToolHandler) error
	RegisterTool(ToolSpec) error
}

// ToolCall retains the actual Agent and streams output into its ordinary Job.
// A handler must cooperate with cancellation; late writes after kill are ignored.
type ToolCall interface {
	io.Writer
	Parent() Agent
	StartProcess(func() (io.Writer, func() error, error)) error
	SetExitCode(int)
}

type ToolHandler func(context.Context, ToolCall, json.RawMessage) error

// ToolOperation separates serial admission from a cancellable observation.
// Cancellation of Await releases observation, never execution.
type ToolOperation interface {
	Called() ToolData
	SetEventSeq(Seq)
	Dispatched() bool
	Release()
	ReleaseAfter(func() error) error
	Abort(error)
	Await(context.Context) ToolData
}
