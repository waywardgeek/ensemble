package mcp

import "encoding/json"

// Transport is the wire layer for JSON-RPC messages.
//
// Three implementations cover every MCP server scenario:
//   - PipeTransport: in-process testing (grader, unit tests)
//   - StdioTransport: spawn a local subprocess (Python/Go MCP server)
//   - WSTransport: tunnel over the GUI WebSocket
//
// A future fourth (URL/port-based) can be added later — the interface is the seam.
type Transport interface {
	Send(msg json.RawMessage) error
	Recv() (json.RawMessage, error)
	Close() error
}
