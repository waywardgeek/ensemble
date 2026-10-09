package common

import (
	"context"
	"encoding/json"
)

const MCPVersion = "2026-07-28"
const MCPMessageLimit = 8 << 20

type MCPError struct{ Code, Message string }

func (e *MCPError) Error() string { return e.Code + ": " + e.Message }

type MCPMessage struct {
	RequestID string
	JSON      []byte
}
type MCPTransportConstructor interface {
	Open(context.Context, MCPConnection) (MCPTransport, error)
}
type MCPTransport interface {
	Connection() MCPConnection
	Send(context.Context, MCPMessage) error
	Receive(context.Context) ([]byte, error)
	Abandon(context.Context, string, []byte) error
	Diagnostics() MCPDiagnostics
	Close() error
}
type MCPConnection interface {
	Service() MCPService
	Snapshot() MCPConnectionSnapshot
}
type MCPConnectionSpec struct {
	Key                string
	Transport          MCPTransportConstructor
	CallTimeoutSeconds int
}
type MCPSelection struct {
	Alias      string `json:"alias"`
	Connection string `json:"connection"`
	RemoteName string `json:"remote_name"`
}
type MCPDefinition struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	InputSchema  json.RawMessage `json:"inputSchema"`
	OutputSchema json.RawMessage `json:"outputSchema"`
}
type MCPBinding struct {
	Alias      string        `json:"alias"`
	Connection string        `json:"connection"`
	RemoteName string        `json:"remote_name"`
	Definition MCPDefinition `json:"definition"`
}
type MCPConnectionSnapshot struct {
	Key             string `json:"key"`
	State           string `json:"state"`
	Generation      uint64 `json:"generation"`
	ProtocolVersion string `json:"protocol_version"`
	DiscoveredCount int    `json:"discovered_count"`
	ErrorCode       string `json:"error_code"`
}
type MCPBindingSnapshot struct {
	Alias           string  `json:"alias"`
	Connection      string  `json:"connection"`
	RemoteName      string  `json:"remote_name"`
	State           string  `json:"state"`
	Generation      *uint64 `json:"generation"`
	ProtocolVersion string  `json:"protocol_version"`
	ErrorCode       string  `json:"error_code"`
}
type MCPDiagnostics struct {
	StderrTail []byte
	Truncated  bool
}
type MCPService interface {
	Ensemble() Ensemble
	Configure([]MCPConnectionSpec) error
	Prepare(context.Context, string) (MCPConnectionSnapshot, error)
	Reopen(context.Context, string) (MCPConnectionSnapshot, error)
	CloseConnection(string) error
	Connections() []MCPConnectionSnapshot
	Definitions(string) ([]MCPDefinition, error)
	Diagnostics(string) (MCPDiagnostics, error)
	Close() error
}

// MCPRuntime is reached only through the application owner. It operates on
// frozen bindings and Jobs; Registry/Actor remain admission authority.
type MCPRuntime interface {
	MCPService
	ValidateSelections([]MCPSelection) error
 Freeze([]MCPSelection) ([]MCPBinding, error)
	ValidateBindings([]MCPBinding) error
	Attach(MCPAgent) error
	Detach(MCPAgent)
	Invoke(context.Context, Job, MCPBinding, json.RawMessage) ExecutionResult
	BindingState([]MCPBinding, bool) []MCPBindingSnapshot
}
type MCPAgent interface {
	Agent
	MCPBindings() ([]MCPBinding, error)
	MCPStateChanged()
}
type MCPStdioOptions struct {
	Command, CWD       string
	Args, EnvAllowlist []string
}
type MCPMessageEndpoint interface {
	Send(context.Context, []byte) error
	Receive(context.Context) ([]byte, error)
	Close() error
}
type MCPMessageEndpointSource interface {
	Open(context.Context) (MCPMessageEndpoint, error)
}

// MCPConnectionOwner is the actual service parent of live connections.
type MCPConnectionOwner interface {
 MCPService
 ConnectionChanged()
}

type MCPStdioSpec struct { Key string; CallTimeoutSeconds int; Options MCPStdioOptions }
type MCPFileConfig struct { Connections []MCPStdioSpec; Bindings []MCPSelection }
