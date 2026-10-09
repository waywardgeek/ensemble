// Package mcp owns discovery, protocol correlation, schemas and remote results.
package mcp

import (
	"context"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"regexp"
	"sort"
	"sync"
	"time"
)

type registration struct {
	gate       sync.Mutex
	spec       common.MCPConnectionSpec
	connection *connection
}
type Service struct {
	parent     common.Ensemble
	mu         sync.Mutex
	entries    map[string]*registration
	agents     map[common.MCPAgent][]common.MCPBinding
	generation uint64
	closed     bool
}

func New(parent common.Ensemble) *Service {
	return &Service{parent: parent, entries: map[string]*registration{}, agents: map[common.MCPAgent][]common.MCPBinding{}}
}
func (s *Service) Ensemble() common.Ensemble { return s.parent }
func (s *Service) json() common.JSONService  { return s.parent.JSON() }
func failure(code, message string) *common.MCPError {
	return &common.MCPError{Code: code, Message: message}
}
func (s *Service) Configure(specs []common.MCPConnectionSpec) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return failure("mcp_unavailable", "MCP service is closed.")
	}
	if len(s.entries)+len(specs) > 32 {
		return failure("mcp_limit", "Connection limit.")
	}
	seen := map[string]bool{}
	for _, x := range specs {
		if !validKey(x.Key) || seen[x.Key] || s.entries[x.Key] != nil || x.Transport == nil {
			return failure("mcp_unavailable", "Invalid or duplicate connection selection.")
		}
		if x.CallTimeoutSeconds < 0 || x.CallTimeoutSeconds > 600 {
			return failure("mcp_limit", "Call timeout must be 1–600 seconds.")
		}
		seen[x.Key] = true
	}
	for _, x := range specs {
		if x.CallTimeoutSeconds == 0 {
			x.CallTimeoutSeconds = 120
		}
		s.entries[x.Key] = &registration{spec: x}
	}
	return nil
}
func validKey(s string) bool    { ok, _ := regexp.MatchString(`^[a-z][a-z0-9_]{0,63}$`, s); return ok }
func validRemote(s string) bool { ok, _ := regexp.MatchString(`^[A-Za-z0-9_.-]{1,128}$`, s); return ok }
func (s *Service) entry(key string) (*registration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.entries[key]
	if r == nil || s.closed {
		return nil, failure("mcp_unavailable", "Connection is unavailable.")
	}
	return r, nil
}
func (s *Service) Prepare(ctx context.Context, key string) (common.MCPConnectionSnapshot, error) {
	return s.prepare(ctx, key, false)
}
func (s *Service) Reopen(ctx context.Context, key string) (common.MCPConnectionSnapshot, error) {
	return s.prepare(ctx, key, true)
}
func (s *Service) prepare(ctx context.Context, key string, reopen bool) (common.MCPConnectionSnapshot, error) {
	var zero common.MCPConnectionSnapshot
	r, err := s.entry(key)
	if err != nil {
		return zero, err
	}
	r.gate.Lock()
	defer r.gate.Unlock()
	s.mu.Lock()
	old := r.connection
	closed := s.closed
	s.mu.Unlock()
	if closed {
		return zero, failure("mcp_unavailable", "MCP service is closed.")
	}
	if old != nil {
		state := old.Snapshot()
		if !reopen {
			if state.State == "ready" {
				return state, nil
			}
			return state, failure("mcp_unavailable", "Explicit reopen is required.")
		}
		old.stop("closed", "mcp_unavailable")
		<-old.joined
	}
	s.mu.Lock()
	if s.closed || s.generation == ^uint64(0) {
		s.mu.Unlock()
		return zero, failure("mcp_limit", "Connection generation exhausted or service closed.")
	}
	s.generation++
	c := newConnection(s, r.spec, s.generation)
	r.connection = c
	s.mu.Unlock()
	s.notify()
	deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	t, err := r.spec.Transport.Open(deadline, c)
	if err != nil {
		c.failBeforeStart("mcp_transport")
		return c.Snapshot(), failure("mcp_transport", "Transport preparation failed.")
	}
	c.start(t)
	if err = c.discover(deadline); err != nil {
		code := "mcp_protocol"
		if e, ok := err.(*common.MCPError); ok {
			code = e.Code
		}
		c.stop("failed", code)
		<-c.joined
		return c.Snapshot(), err
	}
	s.mu.Lock()
	for _, bindings := range s.agents {
		for _, b := range bindings {
			if b.Connection == key && !c.matches(b) {
				err = failure("mcp_unavailable", "Frozen remote definition changed.")
				break
			}
		}
		if err != nil {
			break
		}
	}
	if s.closed {
		err = failure("mcp_unavailable", "MCP service closed during preparation.")
	}
	if err == nil {
		c.mu.Lock()
		if c.state == "preparing" {
			c.state = "ready"
		} else {
			err = failure("mcp_transport", "Transport closed during preparation.")
		}
		c.mu.Unlock()
	}
	s.mu.Unlock()
	if err != nil {
		c.stop("failed", "mcp_unavailable")
		<-c.joined
		return c.Snapshot(), err
	}
	s.notify()
	return c.Snapshot(), nil
}
func (s *Service) CloseConnection(key string) error {
	r, e := s.entry(key)
	if e != nil {
		return e
	}
	r.gate.Lock()
	defer r.gate.Unlock()
	s.mu.Lock()
	c := r.connection
	s.mu.Unlock()
	if c != nil {
		c.stop("closed", "mcp_unavailable")
		<-c.joined
	}
	return nil
}
func (s *Service) Close() error {
	s.mu.Lock()
	s.closed = true
	list := make([]*registration, 0, len(s.entries))
	for _, r := range s.entries {
		list = append(list, r)
	}
	s.mu.Unlock()
	for _, r := range list {
		r.gate.Lock()
		s.mu.Lock()
		c := r.connection
		s.mu.Unlock()
		if c != nil {
			c.stop("closed", "mcp_unavailable")
			<-c.joined
		}
		r.gate.Unlock()
	}
	return nil
}
func (s *Service) Connections() []common.MCPConnectionSnapshot {
	s.mu.Lock()
	cs := []*connection{}
	for _, r := range s.entries {
		if r.connection != nil {
			cs = append(cs, r.connection)
		}
	}
	s.mu.Unlock()
	out := []common.MCPConnectionSnapshot{}
	for _, c := range cs {
		out = append(out, c.Snapshot())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}
func cloneDefinition(d common.MCPDefinition) common.MCPDefinition {
	d.InputSchema = append(json.RawMessage(nil), d.InputSchema...)
	d.OutputSchema = append(json.RawMessage(nil), d.OutputSchema...)
	return d
}
func cloneBindings(b []common.MCPBinding) []common.MCPBinding {
	out := append([]common.MCPBinding{}, b...)
	for i := range out {
		out[i].Definition = cloneDefinition(out[i].Definition)
	}
	return out
}
func (s *Service) Definitions(key string) ([]common.MCPDefinition, error) {
	r, e := s.entry(key)
	if e != nil {
		return nil, e
	}
	s.mu.Lock()
	c := r.connection
	s.mu.Unlock()
	if c == nil {
		return nil, failure("mcp_unavailable", "Connection has not been prepared.")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != "ready" {
		return nil, failure("mcp_unavailable", "Connection is not ready.")
	}
	out := []common.MCPDefinition{}
	for _, d := range c.definitions {
		out = append(out, cloneDefinition(d))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
func (s *Service) Diagnostics(key string) (common.MCPDiagnostics, error) {
	r, e := s.entry(key)
	if e != nil {
		return common.MCPDiagnostics{}, e
	}
	s.mu.Lock()
	c := r.connection
	s.mu.Unlock()
	if c == nil {
		return common.MCPDiagnostics{}, nil
	}
	c.mu.Lock()
	t := c.transport
	c.mu.Unlock()
	if t == nil {
		return common.MCPDiagnostics{}, nil
	}
	d := t.Diagnostics()
	d.StderrTail = append([]byte(nil), d.StderrTail...)
	return d, nil
}
func (s *Service) Freeze(selections []common.MCPSelection) ([]common.MCPBinding, error) {
	if len(selections) > 1024 {
		return nil, failure("mcp_limit", "Remote alias limit.")
	}
	out := []common.MCPBinding{}
	seen := map[string]bool{}
	for _, x := range selections {
		if !validKey(x.Alias) || !validKey(x.Connection) || !validRemote(x.RemoteName) || seen[x.Alias] {
			return nil, failure("mcp_unavailable", "Invalid or duplicate binding.")
		}
		seen[x.Alias] = true
		defs, err := s.Definitions(x.Connection)
		if err != nil {
			return nil, err
		}
		found := false
		for _, d := range defs {
			if d.Name == x.RemoteName {
				out = append(out, common.MCPBinding{Alias: x.Alias, Connection: x.Connection, RemoteName: x.RemoteName, Definition: d})
				found = true
				break
			}
		}
		if !found {
			return nil, failure("mcp_unavailable", "Selected remote tool is missing.")
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Alias < out[j].Alias })
	return out, nil
}
func (s *Service) ValidateBindings(bindings []common.MCPBinding) error {
	if len(bindings) > 1024 {
		return failure("mcp_limit", "Remote alias limit.")
	}
	keys := map[string]bool{}
	for i, b := range bindings {
		if !validKey(b.Alias) || !validKey(b.Connection) || !validRemote(b.RemoteName) || b.Definition.Name != b.RemoteName || i > 0 && bindings[i-1].Alias >= b.Alias {
			return failure("mcp_protocol", "Invalid frozen binding.")
		}
		keys[b.Connection] = true
		if _, err := validateDefinition(s, b.Definition); err != nil {
			return err
		}
	}
	if len(keys) > 32 {
		return failure("mcp_limit", "Connection limit.")
	}
	return nil
}
func (s *Service) Attach(a common.MCPAgent) error {
	b, e := a.MCPBindings()
	if e != nil {
		return e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return failure("mcp_unavailable", "MCP service is closed.")
	}
	for _, x := range b {
		r := s.entries[x.Connection]
		if r == nil || r.connection == nil || r.connection.Snapshot().State != "ready" || !r.connection.matches(x) {
			return failure("mcp_unavailable", "Connection changed before Agent publication.")
		}
	}
	s.agents[a] = cloneBindings(b)
	return nil
}
func (s *Service) Detach(a common.MCPAgent) { s.mu.Lock(); delete(s.agents, a); s.mu.Unlock() }
func (s *Service) notify() {
	s.mu.Lock()
	list := []common.MCPAgent{}
	for a := range s.agents {
		list = append(list, a)
	}
	s.mu.Unlock()
	for _, a := range list {
		a.MCPStateChanged()
	}
}
func (s *Service) BindingState(b []common.MCPBinding, live bool) []common.MCPBindingSnapshot {
	out := []common.MCPBindingSnapshot{}
	states := map[string]common.MCPConnectionSnapshot{}
	if live {
		for _, x := range s.Connections() {
			states[x.Key] = x
		}
	}
	for _, x := range b {
		v := common.MCPBindingSnapshot{Alias: x.Alias, Connection: x.Connection, RemoteName: x.RemoteName, State: "closed", ProtocolVersion: common.MCPVersion, ErrorCode: "mcp_unavailable"}
		if state, ok := states[x.Connection]; ok {
			g := state.Generation
			v.Generation = &g
			v.State = state.State
			v.ErrorCode = state.ErrorCode
		}
		out = append(out, v)
	}
	return out
}
func (s *Service) Invoke(ctx context.Context, job common.Job, b common.MCPBinding, args json.RawMessage) common.ExecutionResult {
	bad := func(err error) common.ExecutionResult {
		e, ok := err.(*common.MCPError)
		if !ok {
			e = failure("mcp_invalid_result", "Remote result is invalid.")
		}
		data, _ := s.json().Encode(map[string]any{"error": e.Code, "message": e.Message}, true, 4096)
		return common.ExecutionResult{Text: string(data) + "\n", IsError: true}
	}
	// Arguments are validated locally before capacity, IDs or transport delivery.
	schemas, err := validateDefinition(s, b.Definition)
	if err != nil {
		return bad(err)
	}
	v, err := s.json().Parse(args, messageBounds())
	if err != nil {
		return common.ExecutionResult{Text: "invalid arguments: expected unambiguous JSON object", IsError: true}
	}
	if _, ok := v.(map[string]any); !ok {
		return common.ExecutionResult{Text: "invalid arguments: expected object", IsError: true}
	}
	if err = schemas.input.validate(v); err != nil {
		return common.ExecutionResult{Text: "invalid arguments: remote input schema refused the complete object", IsError: true}
	}
	r, err := s.entry(b.Connection)
	if err != nil {
		return bad(err)
	}
	s.mu.Lock()
	c := r.connection
	s.mu.Unlock()
	if c == nil || c.Snapshot().State != "ready" || !c.matches(b) {
		return bad(failure("mcp_unavailable", "Frozen connection is unavailable."))
	}
	deadline, cancel := context.WithTimeout(ctx, time.Duration(r.spec.CallTimeoutSeconds)*time.Second)
	defer cancel()
	result, err := c.rpc(deadline, "tools/call", map[string]any{"name": b.RemoteName, "arguments": v})
	if err != nil {
		return bad(err)
	}
	text, isError, err := s.result(result, schemas.output)
	if err != nil {
		return bad(err)
	}
	return common.ExecutionResult{Text: text, IsError: isError}
}

type definitionSchemas struct{ input, output *schema }

func validateDefinition(s common.MCPService, d common.MCPDefinition) (definitionSchemas, error) {
	out := definitionSchemas{}
	if !validRemote(d.Name) || len(d.Description) > 64<<10 {
		return out, failure("mcp_protocol", "Invalid remote definition.")
	}
	var err error
	out.input, err = compile(s, d.InputSchema, true)
	if err != nil {
		return out, err
	}
	if d.OutputSchema != nil {
		out.output, err = compile(s, d.OutputSchema, false)
	}
	return out, err
}
func equalDefinition(s common.MCPService, a, b common.MCPDefinition) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	vx, e := s.Ensemble().JSON().Parse(x, messageBounds())
	if e != nil {
		return false
	}
	vy, e := s.Ensemble().JSON().Parse(y, messageBounds())
	if e != nil {
		return false
	}
	cx, e := s.Ensemble().JSON().Encode(vx, true, common.MCPMessageLimit)
	if e != nil {
		return false
	}
	cy, e := s.Ensemble().JSON().Encode(vy, true, common.MCPMessageLimit)
	return e == nil && string(cx) == string(cy)
}
func messageBounds() common.JSONBounds {
	return common.JSONBounds{Bytes: common.MCPMessageLimit, Depth: 64, Nodes: 100000, Collection: 100000, Scalars: true}
}

func (s *Service) ConnectionChanged() { s.notify() }

func (s *Service) ValidateSelections(selections []common.MCPSelection) error {
 s.mu.Lock();defer s.mu.Unlock()
 if len(selections)>1024{return failure("mcp_limit","Remote alias limit.")}
 seen:=map[string]bool{}
 for _,b:=range selections {if !validKey(b.Alias)||!validKey(b.Connection)||!validRemote(b.RemoteName)||seen[b.Alias]||s.entries[b.Connection]==nil{return failure("mcp_unavailable","Invalid or missing selection.")};seen[b.Alias]=true}
 return nil
}

func (s *Service) StopAdmission() { s.mu.Lock(); s.closed=true; s.mu.Unlock() }
