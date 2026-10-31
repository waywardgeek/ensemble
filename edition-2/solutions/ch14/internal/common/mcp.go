package common

import (
	"context"
	"encoding/json"
)

// MCPTransport is an external message boundary, independent of GUI framing.
// Close must unblock Recv and Send; one receiver and concurrent senders use it.
type MCPTransport interface {
	// Send writes one complete JSON-RPC message.
	Send(json.RawMessage) error
	// Recv reads one complete JSON-RPC message.
	Recv() (json.RawMessage, error)
	// Close releases the boundary and wakes blocked I/O.
	Close() error
}

// MCPServer is a skill's launch declaration, not an enabled-tool grant.
type MCPServer struct {
	// Name identifies the connection for its owning skill.
	Name string
	// Transport selects stdio or websocket.
	Transport string
	// Command names the executable for a stdio server.
	Command string
	// Args are literal command arguments, without shell expansion.
	Args []string
}

// MCP owns one Agent's external clients and their discovered tool lifetimes.
type MCP interface {
	// Agent exposes the owner used for tool admission, jobs and diagnostics.
	Agent() Agent
	// Connect initializes an externally supplied transport and discovers tools.
	// With skills enabled, name must match a declared server.
	Connect(context.Context, string, MCPTransport) error
	// Done closes when the named peer disconnects; unknown peers are already done.
	Done(string) <-chan struct{}
	// AttachBrowser offers a GUI tunnel; a declared skill can start it later.
	AttachBrowser(context.Context, MCPTransport) error
	// StartSkill starts stdio peers and records websocket peers awaiting a browser.
	StartSkill(string, []MCPServer, []string) error
	// StopSkill closes only the connections declared by this skill.
	StopSkill(string) error
	// Tools returns active, skill-admitted definitions in discovery order.
	Tools() []ToolDefinition
	// Close stops all connections during Agent shutdown.
	Close() error
}

// MCPTransportOwner provides the endpoint's actual owner without GUI dependencies.
type MCPTransportOwner interface {
	// Agent reaches the conversation served by this endpoint.
	Agent() Agent
}
