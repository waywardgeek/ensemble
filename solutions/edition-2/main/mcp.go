package ensemble

import (
	"context"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/mcpmemory"
 "example.com/ensemble/internal/mcpconfig"
	"example.com/ensemble/internal/mcpstdio"
	"example.com/ensemble/internal/tools"
	"fmt"
)

type MCPRoot = common.Ensemble
type JSONService = common.JSONService
type JSONBounds = common.JSONBounds
type MCPService = common.MCPService
type MCPConnection = common.MCPConnection
type MCPTransport = common.MCPTransport
type MCPTransportConstructor = common.MCPTransportConstructor
type MCPMessage = common.MCPMessage
type MCPMessageEndpoint = common.MCPMessageEndpoint
type MCPMessageEndpointSource = common.MCPMessageEndpointSource
type MCPConnectionSpec = common.MCPConnectionSpec
type MCPConnectionSnapshot = common.MCPConnectionSnapshot
type MCPBindingSnapshot = common.MCPBindingSnapshot
type MCPBinding = common.MCPBinding
type MCPSelection = common.MCPSelection
type MCPDefinition = common.MCPDefinition
type MCPDiagnostics = common.MCPDiagnostics
type MCPStdioOptions = common.MCPStdioOptions
type MCPError = common.MCPError

func (e *Ensemble) MCP() common.MCPService                                    { return e.mcp }
func (e *Ensemble) MCPRuntime() common.MCPRuntime                             { return e.mcp }
func NewStdioMCPTransport(o MCPStdioOptions) (MCPTransportConstructor, error) { return mcpstdio.New(o) }
func NewMemoryMCPTransport(s MCPMessageEndpointSource) MCPTransportConstructor {
	return mcpmemory.New(s)
}
func (a *Agent) MCPBindings() ([]MCPBinding, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := append([]MCPBinding{}, a.bindings...)
	for i := range out {
		out[i].Definition.InputSchema = append(json.RawMessage(nil), out[i].Definition.InputSchema...)
		out[i].Definition.OutputSchema = append(json.RawMessage(nil), out[i].Definition.OutputSchema...)
	}
	return out, nil
}
func (a *Agent) MCPState() ([]MCPBindingSnapshot, error) {
	if a.actor != nil {
		return a.actor.MCPState()
	}
	return a.mcpView(false), nil
}
func (a *Agent) mcpView(live bool) []MCPBindingSnapshot {
	b, _ := a.MCPBindings()
	return a.parent.MCPRuntime().BindingState(b, live)
}
func (a turnAgent) MCPView() []common.MCPBindingSnapshot { return a.mcpView(true) }
func (a *Agent) MCPStateChanged() {
	if a.actor != nil {
		_ = a.actor.MCPChanged()
	}
}
func (a *Agent) installMCP(bindings []MCPBinding) error {
	if err := a.parent.MCPRuntime().ValidateBindings(bindings); err != nil {
		return err
	}
	if len(bindings) == 0 { bindings = nil }; a.bindings = bindings
	registry, err := tools.New(toolAgent{a}, a.config.Builtins)
	if err != nil {
		return err
	}
	a.registry = registry
	return nil
}
func (a *Agent) prepareMCP(ctx context.Context, recorded []MCPBinding) error {
	if err:=a.parent.MCPRuntime().ValidateSelections(a.config.MCPBindings);err!=nil{return err}
 keys := map[string]bool{}
	for _, s := range a.config.MCPBindings {
		keys[s.Connection] = true
	}
	// Freeze validates alias grammar, names and all ready selections after prepare.
	ordered := make([]string, 0, len(keys))
	for k := range keys {
		ordered = append(ordered, k)
	}
	sortStrings(ordered)
	for _, k := range ordered {
		if _, err := a.parent.MCP().Prepare(ctx, k); err != nil {
			return err
		}
	}
	b, err := a.parent.MCPRuntime().Freeze(a.config.MCPBindings)
	if err != nil {
		return err
	}
	if recorded != nil {
		left, _ := json.Marshal(recorded)
		right, _ := json.Marshal(b)
		if !a.codec.EqualJSON(left, right) {
			return sessionError("session_incompatible", "remote definitions differ")
		}
	}
	return a.installMCP(b)
}
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
func (a *Agent) freezePreparedMCP() error {
	b, err := a.parent.MCPRuntime().Freeze(a.config.MCPBindings)
	if err != nil {
		return err
	}
	return a.installMCP(b)
}
func (a *Agent) checkMCPSelections(recorded []MCPBinding) error {
	if len(recorded) != len(a.config.MCPBindings) {
		return sessionError("session_incompatible", "remote selections differ")
	}
	seen := map[string]bool{}
	for _, x := range a.config.MCPBindings {
		if seen[x.Alias] {
			return fmt.Errorf("duplicate remote alias")
		}
		seen[x.Alias] = true
		found := false
		for _, b := range recorded {
			if x.Alias == b.Alias && x.Connection == b.Connection && x.RemoteName == b.RemoteName {
				found = true
				break
			}
		}
		if !found {
			return sessionError("session_incompatible", "remote selections differ")
		}
	}
	return a.installMCP(recorded)
}

// LoadMCPConfig validates the whole file before registering any connection. It
// neither samples environment values nor opens a transport.
func (e *Ensemble) LoadMCPConfig(path string) ([]MCPSelection,error) {
 file,err:=mcpconfig.New(e).Read(path);if err!=nil{return nil,err}
 specs:=[]MCPConnectionSpec{}
 for _,c:=range file.Connections { constructor,err:=NewStdioMCPTransport(c.Options);if err!=nil{return nil,err};specs=append(specs,MCPConnectionSpec{Key:c.Key,Transport:constructor,CallTimeoutSeconds:c.CallTimeoutSeconds}) }
 if err=e.MCP().Configure(specs);err!=nil{return nil,err};return file.Bindings,nil
}
// PrepareMCPSelections prepares only selected logical keys, in key order.
func (e *Ensemble) PrepareMCPSelections(ctx context.Context,selections []MCPSelection) error {
 if err:=e.mcp.ValidateSelections(selections);err!=nil{return err}
 keys:=map[string]bool{};for _,s:=range selections{keys[s.Connection]=true};ordered:=[]string{};for k:=range keys{ordered=append(ordered,k)};sortStrings(ordered)
 for _,k:=range ordered{if _,err:=e.MCP().Prepare(ctx,k);err!=nil{return err}};return nil
}
