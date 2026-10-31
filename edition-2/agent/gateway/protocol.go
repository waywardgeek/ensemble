package gateway

import (
	"encoding/json"
	"fmt"
	"time"
)

// ResponseAsync asks for a routed reply, without holding the sender's Send call.
const ResponseAsync = "async"

// Source is provenance assigned by the gateway after authentication and policy checks.
type Source struct {
	// Type identifies an authenticated agent rather than sender-authored content.
	Type string `json:"type"`
	// ChannelID names the accepted group supplying this message's authority.
	ChannelID string `json:"channel_id"`
	// UserID is the authenticated sending agent's identifier.
	UserID string `json:"user_id"`
	// DisplayName is the operator-configured readable sender name.
	DisplayName string `json:"display_name,omitempty"`
	// Trust is owner, trusted, or external, derived from the accepted group.
	Trust string `json:"trust"`
}

// Envelope is a routed message; Seq is scoped to its recipient and never reused.
type Envelope struct {
	// ID is the gateway's message identity for diagnostics and correlation.
	ID string `json:"id"`
	// Seq identifies receipt progress in the recipient's queue.
	Seq uint64 `json:"seq"`
	// ToAgent is the authorized recipient.
	ToAgent string `json:"to_agent"`
	// Source contains authenticated provenance, independent of the content.
	Source Source `json:"source"`
	// Timestamp records gateway admission time in UTC.
	Timestamp time.Time `json:"timestamp"`
	// ConversationID is preserved across a request and its response.
	ConversationID string `json:"conversation_id,omitempty"`
	// Content includes the gateway provenance header above the original text.
	Content string `json:"content"`
	// ResponseMode indicates whether the sender requested an asynchronous answer.
	ResponseMode string `json:"response_mode,omitempty"`
}

// ProvenanceHeader renders routing facts, not an interpretation of message content.
func ProvenanceHeader(source Source) string {
	return fmt.Sprintf("[GATEWAY source=%s user=%s trust=%s]", source.ChannelID, source.UserID, source.Trust)
}

// OutboundMessage deliberately has no Source field: callers cannot choose trust.
type OutboundMessage struct {
	// ToAgent names the intended recipient, whose policy is checked by the gateway.
	ToAgent string `json:"to_agent"`
	// Content is untrusted application text until stamped at the routing boundary.
	Content string `json:"content"`
	// ConversationID correlates application requests and answers.
	ConversationID string `json:"conversation_id,omitempty"`
	// ResponseMode preserves the caller's requested reply behavior.
	ResponseMode string `json:"response_mode,omitempty"`
}

// ChildRegistration grants a new independently authenticated identity a subset of groups.
type ChildRegistration struct {
	// AgentID must be an unused identity, never a replacement of an existing peer.
	AgentID string `json:"agent_id"`
	// PublicKey is the child's X25519 public half; no secret crosses this boundary.
	PublicKey []byte `json:"public_key"`
	// Groups must all be held by the authenticated registering parent.
	Groups []string `json:"groups"`
}

// WelcomePayload confirms authenticated readiness before Connect returns.
type WelcomePayload struct {
	// AgentID is the identity verified during the handshake and hello exchange.
	AgentID string `json:"agent_id"`
	// QueueDepth counts outstanding messages after applying the resume receipt.
	QueueDepth int `json:"queue_depth"`
	// ServerVersion identifies this gateway protocol implementation.
	ServerVersion string `json:"server_version"`
}

// Frame is the wire vocabulary shared by independent client and server packages.
// Application operations use MCP requests; connection/delivery messages stay small.
type Frame struct {
	// Type discriminates hello, welcome, deliver, ack, response, ping/pong, error or MCP.
	Type string `json:"type"`
	// AgentID identifies hello; authentication still comes from Noise, never this string.
	AgentID string `json:"agent_id,omitempty"`
	// Version identifies the protocol spoken by hello.
	Version int `json:"version,omitempty"`
	// Seq is the cumulative ack or hello's last successfully handled sequence.
	Seq uint64 `json:"last_acked_seq,omitempty"`
	// Welcome is present only on an accepted connection.
	Welcome *WelcomePayload `json:"welcome,omitempty"`
	// Envelope is present on a routed delivery.
	Envelope *Envelope `json:"envelope,omitempty"`
	// Message is an outbound asynchronous response, checked like any other send.
	Message *OutboundMessage `json:"message,omitempty"`
	// RPC carries a JSON-RPC 2.0 MCP request or response under authenticated transport.
	RPC *RPC `json:"rpc,omitempty"`
	// Error contains a connection-level diagnostic, never keys or message bodies.
	Error string `json:"error,omitempty"`
}

// RPC is a JSON-RPC 2.0 envelope used for the tunneled MCP administration surface.
type RPC struct {
	// JSONRPC is the protocol version, always 2.0.
	JSONRPC string `json:"jsonrpc"`
	// ID correlates a response with its request; zero is reserved for notifications.
	ID uint64 `json:"id,omitempty"`
	// Method is initialize, tools/list, tools/call or notifications/initialized.
	Method string `json:"method,omitempty"`
	// Params is the method-specific input object.
	Params json.RawMessage `json:"params,omitempty"`
	// Result is a successful method-specific response.
	Result json.RawMessage `json:"result,omitempty"`
	// Error carries a failed request's JSON-RPC error.
	Error *RPCError `json:"error,omitempty"`
}

// RPCError is a correlated protocol or authorization refusal.
type RPCError struct {
	// Code distinguishes invalid input from execution refusal.
	Code int `json:"code"`
	// Message is a safe explanation suitable for the caller.
	Message string `json:"message"`
}
