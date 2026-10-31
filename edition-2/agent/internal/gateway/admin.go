package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	wire "ensemble/gateway"
)

// request holds policy ownership through validation and mutation. Replies are
// written after releasing the lock so a slow administrator cannot stop routing.
func (s *Server) request(id string, r wire.RPC) *wire.RPC {
	reply := &wire.RPC{JSONRPC: "2.0", ID: r.ID}
	fail := func(code int, err error) *wire.RPC {
		reply.Error = &wire.RPCError{Code: code, Message: err.Error()}
		return reply
	}
	if r.JSONRPC != "2.0" {
		return fail(-32600, errors.New("expected JSON-RPC 2.0"))
	}
	var result any
	switch r.Method {
	// MCP runs inside the already authenticated connection. There is no second
	// admin listener or bearer-token shortcut with a different trust policy.
	case "notifications/initialized":
		return nil
	case "initialize":
		result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]string{"name": "ensemble-gateway", "version": "1"}}
	case "tools/list":
		result = map[string]any{"tools": toolSchemas()}
	case "tools/call":
		var call struct {
			// Name selects a published gateway operation.
			Name string `json:"name"`
			// Arguments is decoded only by the selected operation.
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(r.Params, &call); err != nil {
			return fail(-32602, err)
		}
		s.mu.Lock()
		value, err := s.call(id, call.Name, call.Arguments)
		s.mu.Unlock()
		// Tool refusal is an ordinary MCP execution result. JSON-RPC errors
		// describe malformed/unknown protocol requests, not a denied group grant.
		if err != nil {
			result = map[string]any{"content": []map[string]string{{"type": "text", "text": err.Error()}}, "isError": true}
			break
		}
		encoded, _ := json.Marshal(value)
		result = map[string]any{"content": []map[string]string{{"type": "text", "text": string(encoded)}}, "isError": false}
	default:
		return fail(-32601, errors.New("unknown MCP method"))
	}
	reply.Result, _ = json.Marshal(result)
	return reply
}

func toolSchemas() []map[string]any {
	// These schemas describe the actual administration contract, rather than
	// exposing internal structs or accepting arbitrary unvalidated command blobs.
	schema := func(name, description string, properties map[string]any, required ...string) map[string]any {
		if required == nil {
			required = []string{}
		}
		return map[string]any{"name": name, "description": description, "inputSchema": map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}}
	}
	text := map[string]any{"type": "string"}
	groups := map[string]any{"type": "array", "items": text}
	return []map[string]any{
		schema("send-message", "Route a message under the recipient's group policy.", map[string]any{"to_agent": text, "content": text, "conversation_id": text, "response_mode": text}, "to_agent", "content"),
		schema("declare-contactable", "Replace this identity's accepted sender groups.", map[string]any{"groups": groups}, "groups"),
		schema("register-child", "Register an unused X25519 identity with narrow-only groups.", map[string]any{"agent_id": text, "public_key": text, "groups": groups}, "agent_id", "public_key", "groups"),
		schema("list-groups", "Inspect group definitions and your effective memberships.", map[string]any{}),
		schema("put-group", "Create or edit a runtime: group; requires group-admin.", map[string]any{"name": text, "description": text, "capabilities": groups, "trust": text}, "name", "capabilities", "trust"),
		schema("delete-group", "Delete a runtime: group and remove its dynamic grants.", map[string]any{"name": text}, "name"),
		schema("set-groups", "Replace an agent's additional runtime memberships; requires group-admin.", map[string]any{"agent_id": text, "groups": groups}, "agent_id", "groups"),
	}
}

// The caller identity is supplied by the authenticated connection, never by
// tool arguments. Each mutation validates every proposed grant before publishing
// any of it, so an over-grant cannot leave a partially registered child behind.
func (s *Server) call(id, name string, args json.RawMessage) (any, error) {
	switch name {
	case "send-message":
		var message wire.OutboundMessage
		if err := json.Unmarshal(args, &message); err != nil {
			return nil, err
		}
		return "queued", s.route(id, message)
	case "declare-contactable":
		if !s.capable(id, "declare-contactable") {
			return nil, errors.New("declare-contactable capability required")
		}
		var a struct {
			// Groups replaces this caller's accepted senders, including an empty set.
			Groups []string `json:"groups"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return nil, err
		}
		for _, name := range a.Groups {
			if _, ok := s.group(name); !ok {
				return nil, errors.New("unknown contact group")
			}
		}
		s.state.Contactable[id] = slices.Clone(a.Groups)
		return "contactability updated", nil
	case "register-child":
		var child wire.ChildRegistration
		if err := json.Unmarshal(args, &child); err != nil {
			return nil, err
		}
		if !validName(child.AgentID) {
			return nil, errors.New("invalid child ID")
		}
		if _, exists := s.identity(child.AgentID); exists {
			return nil, errors.New("identity already registered")
		}
		if len(child.PublicKey) != 32 {
			return nil, errors.New("child requires a 32-byte X25519 public key")
		}
		// Child registration is deliberately not a privilege reserved for admins.
		// Its safety comes from attenuation: every group is already held by this
		// authenticated parent. An existing identity can never be taken over.
		for _, g := range child.Groups {
			if !slices.Contains(s.memberships(id), g) {
				return nil, errors.New("child group grant would widen parent authority")
			}
		}
		s.state.Children[child.AgentID] = AgentConfig{PublicKey: wire.EncodePublicKey(child.PublicKey), Groups: slices.Clone(child.Groups)}
		s.queues[child.AgentID] = &queue{}
		return "child registered", nil
	case "list-groups":
		// Listing definitions does not itself grant authority. Mutation paths
		// below re-evaluate the authenticated caller's group-admin capability.
		all := map[string]Group{}
		for n, g := range s.config.Groups {
			all[n] = g
		}
		for n, g := range s.state.Groups {
			all[n] = g
		}
		return map[string]any{"groups": all, "memberships": s.memberships(id)}, nil
	case "put-group", "delete-group", "set-groups":
		if !s.capable(id, "group-admin") {
			return nil, errors.New("group-admin capability required")
		}
		return s.admin(name, args)
	default:
		return nil, errors.New("unknown gateway tool")
	}
}
func (s *Server) admin(name string, args json.RawMessage) (any, error) {
	var a struct {
		// Name selects the runtime group being changed.
		Name string `json:"name"`
		// AgentID selects a membership target; it cannot create an identity.
		AgentID string `json:"agent_id"`
		// Groups replaces only the target's additional runtime memberships.
		Groups []string `json:"groups"`
		// Description records the administrator's explanation of this group.
		Description string `json:"description"`
		// Capabilities grants operations independently of contactability.
		Capabilities []string `json:"capabilities"`
		// Trust determines provenance when this group authorizes delivery.
		Trust string `json:"trust"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	// Administrator changes live in the runtime namespace. Base memberships in
	// the operator's file remain facts of that file across restarts.
	if name == "set-groups" {
		if _, ok := s.identity(a.AgentID); !ok {
			return nil, errors.New("unknown agent")
		}
		for _, g := range a.Groups {
			if _, ok := s.state.Groups[g]; !ok {
				return nil, errors.New("only existing runtime groups can be assigned")
			}
		}
		s.state.Memberships[a.AgentID] = slices.Clone(a.Groups)
		return "memberships updated", nil
	}
	if !validName(a.Name) || !isRuntime(a.Name) {
		return nil, errors.New("config groups are immutable; use the runtime: namespace")
	}
	if name == "put-group" {
		g := Group{Description: a.Description, Capabilities: a.Capabilities, Trust: a.Trust}
		if err := validGroup(g); err != nil {
			return nil, err
		}
		s.state.Groups[a.Name] = g
		return "group updated", nil
	}
	if _, ok := s.state.Groups[a.Name]; !ok {
		return nil, errors.New("unknown runtime group")
	}
	delete(s.state.Groups, a.Name)
	// Removing all references prevents a later group with the same name from
	// accidentally restoring grants that an administrator explicitly revoked.
	without := func(groups []string) []string {
		return slices.DeleteFunc(slices.Clone(groups), func(g string) bool { return g == a.Name })
	}
	for id, groups := range s.state.Memberships {
		s.state.Memberships[id] = without(groups)
	}
	for id, groups := range s.state.Contactable {
		s.state.Contactable[id] = without(groups)
	}
	for id, child := range s.state.Children {
		child.Groups = without(child.Groups)
		s.state.Children[id] = child
	}
	return "group deleted", nil
}

// Routing checks two different authorities: may this identity send at all,
// and has this recipient accepted one of its groups? Possession of a valid key
// proves identity; it does not imply either permission.
func (s *Server) route(from string, message wire.OutboundMessage) error {
	if !s.capable(from, "send-message") {
		return errors.New("send-message capability required")
	}
	_, exists := s.identity(message.ToAgent)
	if !exists {
		return errors.New("unknown recipient")
	}
	// Choose the strongest ACCEPTED group, not an unrelated admin membership.
	// Otherwise contactability would authorize one group while stamping another.
	rank := map[string]int{"external": 1, "trusted": 2, "owner": 3}
	selected := ""
	trust := ""
	for _, name := range s.memberships(from) {
		if !slices.Contains(s.state.Contactable[message.ToAgent], name) {
			continue
		}
		g, ok := s.group(name)
		if ok && rank[g.Trust] > rank[trust] {
			selected, trust = name, g.Trust
		}
	}
	if selected == "" {
		return errors.New("sender groups are not accepted by recipient")
	}
	sender, _ := s.identity(from)
	source := wire.Source{Type: "agent", ChannelID: selected, UserID: from, DisplayName: sender.DisplayName, Trust: trust}
	q := s.queues[message.ToAgent]
	if q == nil {
		q = &queue{}
		s.queues[message.ToAgent] = q
	}
	// Allocate under the same lock as authorization and admission. Sequence
	// identity belongs to the recipient, including messages admitted offline.
	q.Next++
	env := wire.Envelope{ID: fmt.Sprintf("%s:%d", message.ToAgent, q.Next), Seq: q.Next, ToAgent: message.ToAgent, Source: source, Timestamp: time.Now().UTC(), ConversationID: message.ConversationID, ResponseMode: message.ResponseMode, Content: wire.ProvenanceHeader(source) + "\n\n" + message.Content}
	q.Messages = append(q.Messages, env)
	if c := s.active[message.ToAgent]; c != nil {
		c.notify()
	}
	return nil
}
