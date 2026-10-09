# Public MCP API, Chapter 11

Implementation contract; runtime validation status is recorded in evidence/ch11.
Public aliases expose common declarations through example.com/ensemble. No GUI
or internal imports are needed by an external transport. This extends accepted
Chapter 10; persistence-format.md specifies the exact version-2 codec.

```go
func (e *Ensemble) MCP() MCPService
func (e *Ensemble) JSON() JSONService
func NewStdioMCPTransport(MCPStdioOptions) (MCPTransportConstructor, error)
func NewMemoryMCPTransport(MCPMessageEndpointSource) MCPTransportConstructor
func (a *Agent) MCPBindings() ([]MCPBinding, error)
func (a *Agent) MCPState() ([]MCPBindingSnapshot, error)
func (e *Ensemble) OpenSessionContext(context.Context, SessionOptions) (*Agent,error)
func (e *Ensemble) ImportSessionContext(context.Context, []byte, SessionOptions) (*Agent,error)
```

MCPRoot aliases common.Ensemble, the coherent parent interface exposing Logf,
JSON, MCP and existing root facilities. MCPService.Ensemble returns that same
interface. MCPConnection has Service() MCPService and Snapshot()
MCPConnectionSnapshot. MCPTransportConstructor.Open(context.Context,
MCPConnection) returns (MCPTransport,error) and receives the actual new parent.

MCPTransport has Connection() MCPConnection, Send(context.Context,MCPMessage)
error, Receive(context.Context)([]byte,error), Abandon(context.Context,string,
[]byte) error, Diagnostics() MCPDiagnostics and Close() error. MCPMessage is
{RequestID string; JSON []byte}. Send completes delivery, not merely insertion
into an unbounded private queue. All retained inputs/returned messages own bytes.
Abandon receives the issued ID and already encoded cancellation notification;
stdio/memory deliver it unchanged. Transports do not parse protocol envelopes.
Close wakes blocked operations before joining owned workers and is idempotent.
Open must honor cancellation and clean partial resources on error.

MCPService has Configure([]MCPConnectionSpec) error; Prepare/Reopen(context.Context,
string)(MCPConnectionSnapshot,error); CloseConnection(string) error;
Connections() []MCPConnectionSnapshot; Definitions(string)([]MCPDefinition,error);
Diagnostics(string)(MCPDiagnostics,error); Close() error. Configure is atomic,
launches nothing and rejects duplicate keys. Prepare reuses ready state; explicit
Reopen closes/joins the old generation first. Generation is allocated before
transport construction and never reused. Configured unopened keys have no
Connection snapshot. A service retains at most 32 keys. Close prevents admission,
unblocks operations, closes transports and joins owned workers.

MCPConnectionSpec is {Key string; Transport MCPTransportConstructor;
CallTimeoutSeconds int}. Public zero timeout selects 120; otherwise 1..600.
MCPStdioOptions is {Command,CWD string; Args,EnvAllowlist []string}; paths are
absolute at construction, no shell/PATH lookup. Only named present environment
values are captured when preparing. macOS/Linux use the specified bounded
process-group shutdown; other process platforms refuse.
MCPMessageEndpointSource.Open(context.Context) returns (MCPMessageEndpoint,error)
for a fresh generation. Endpoint Send(context.Context,[]byte), Receive and Close
have the same complete-message, owned-buffer and unblock contracts. This external
resource does not replace the adapter's Connection parent. Memory does no framing.

Config gains creation-only MCPBindings []MCPSelection, whose fields are Alias,
Connection, RemoteName (strings; JSON alias,connection,remote_name). Fresh usage:
configure root, explicitly prepare selected keys, then NewAgent with selections.
Session opening performs local/store validation before preparing any new endpoint.
The context-free session APIs wrap the context variants. Each Agent freezes its
bindings and ordinary grant/Registry rules govern dispatch. Closing one Agent
cancels only its remote Jobs; shared transports remain root-owned.

MCPBinding has Alias,Connection,RemoteName and Definition MCPDefinition.
MCPDefinition has Name,Description strings, InputSchema,OutputSchema json.RawMessage;
nil output is absent, boolean false is present. JSON definitions use name,
description,inputSchema,outputSchema. Getters return deep copies; schemas cannot
mutate frozen facts. A matching definition does not certify implementation code.

Connection snapshot fields are Key,State,Generation uint64,ProtocolVersion,
DiscoveredCount int,ErrorCode; all other fields strings. Safe Agent snapshots are
exactly alias,connection,remote_name,state,generation,protocol_version,error_code
with positive uint64 generation or null. State is preparing/ready/closed/failed.
Protocol version is 2026-07-28. WatchState requires mcp sorted by alias; changes
are actor-ordered mcp_changed observations. No paths/env/stderr appear there.
MCPDiagnostics has StderrTail []byte and Truncated bool, explicitly requested
application data only. Its tail is bounded to 16 KiB. Errors use *MCPError with
Code,Message strings; untrusted error data/stderr are not copied into messages.

JSONService is root-owned in the neutral jsonvalue spoke. Its Ensemble() parent
returns MCPRoot; Parse([]byte,JSONBounds)(any,error), Encode(any,bool,int)([]byte,error),
ScalarEscapes([]byte) error, Number(string) string, Compare(json.Number,json.Number)
int and Integral(json.Number) bool operate on validated JSON values/numbers.
JSONBounds has Bytes,Depth,Nodes,Collection int and Scalars bool. Positive byte,
depth and collection bounds are mandatory; zero Nodes means no extra total-node
bound for inherited session callers. Scalars enables strict surrogate validation.
MCP always requests it and 8 MiB/64 levels/100,000 nodes. Encode's Boolean selects
canonical number normalization; false preserves number lexemes. Original replay
raw bytes are retained separately. Protocol/session policy remains with callers.

Tools reaches internal runtime execution through Job→Jobs→Agent→Ensemble.
No direct public call-by-alias helper bypasses Registry/Actor admission. Internal
runtime facilities and application parent methods are not server authority.
Public embedding tests implement their own constructor/transport and use ordinary
Agents for calls. Exhaustion controls remain owner-local tests, never setters.
