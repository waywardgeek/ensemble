package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"ensemble/internal/common"
)

type remoteTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
	Ephemeral   string          `json:"ephemeral,omitempty"`
}
type connection struct {
	peer   *peer
	config common.MCPServer
	tools  []remoteTool
	ready  chan struct{}
	err    error
}
type manager struct {
	parent      common.Agent
	mu          sync.Mutex
	connections map[string]*connection
	active      map[string]common.MCPServer
	closed      bool
	ctx         context.Context
	cancel      context.CancelFunc
}

func New(a common.Agent) common.MCPManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &manager{parent: a, connections: map[string]*connection{}, active: map[string]common.MCPServer{}, ctx: ctx, cancel: cancel}
}
func (m *manager) endpoint(d common.MCPServer) common.MCPServer {
	for _, c := range m.parent.Config().MCPServers {
		if c.Name == d.Name {
			return c
		}
	}
	return d
}
func (m *manager) initialize(ctx context.Context, c *connection) error {
	var init struct {
		ProtocolVersion string                     `json:"protocolVersion"`
		Capabilities    map[string]json.RawMessage `json:"capabilities"`
	}
	if err := c.peer.call(ctx, "initialize", map[string]any{"protocolVersion": "2025-03-26", "capabilities": map[string]any{}, "clientInfo": map[string]string{"name": "ensemble", "version": "edition-3-ch18"}}, &init); err != nil {
		return err
	}
	switch init.ProtocolVersion {
	case "2024-11-05", "2025-03-26", "2025-06-18":
	default:
		return errors.New("MCP unsupported protocol version")
	}
	if t, ok := c.peer.transport.(*httpTransport); ok {
		t.setProtocol(init.ProtocolVersion)
	}
	if err := c.peer.notify(ctx, "notifications/initialized", map[string]any{}); err != nil {
		return err
	}
	if _, ok := init.Capabilities["tools"]; !ok {
		return nil
	}
	cursor := ""
	seen := map[string]bool{}
	for {
		var list struct {
			Tools      []remoteTool `json:"tools"`
			NextCursor string       `json:"nextCursor"`
		}
		params := map[string]any{}
		if cursor != "" {
			params["cursor"] = cursor
		}
		if err := c.peer.call(ctx, "tools/list", params, &list); err != nil {
			return err
		}
		c.tools = append(c.tools, list.Tools...)
		if list.NextCursor == "" {
			break
		}
		if seen[list.NextCursor] {
			return errors.New("MCP repeated tools/list cursor")
		}
		seen[list.NextCursor] = true
		cursor = list.NextCursor
	}
	return nil
}
func (m *manager) connect(ctx context.Context, d common.MCPServer) (*connection, error) {
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(m.ctx, cancel)
	defer stop()
	defer cancel()
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, common.ErrStopped
	}
	if c := m.connections[d.Name]; c != nil {
		m.mu.Unlock()
		select {
		case <-c.ready:
			return c, c.err
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-m.ctx.Done():
			return nil, common.ErrStopped
		}
	}
	c := &connection{config: m.endpoint(d), ready: make(chan struct{})}
	m.connections[d.Name] = c
	m.mu.Unlock()
	t, err := dialPrepared(ctx, c.config, m.parent.Config().BaseDir, m.parent.Parent().Logger().Writer(), m.parent.Sandbox().PrepareCommand)
	if err == nil {
		c.peer = newPeer(t, m.localHandler(true), false)
		err = m.initialize(ctx, c)
	}
	m.mu.Lock()
	c.err = err
	close(c.ready)
	if err != nil && m.connections[d.Name] == c {
		delete(m.connections, d.Name)
	}
	closed := m.closed
	m.mu.Unlock()
	if err != nil || closed {
		if c.peer != nil {
			c.peer.close(errors.New("MCP connection closed"))
		}
		if err == nil {
			err = common.ErrStopped
		}
	}
	return c, err
}
func (m *manager) Attach(name string, t common.MCPTransport) error { return m.Connect(m.ctx, name, t) }
func (m *manager) Connect(ctx context.Context, name string, t common.MCPTransport) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(m.ctx, cancel)
	defer stop()
	// With skills enabled, a supplied transport still requires a named declaration.
	// Without skills, Connect is the application owner's explicit admission.
	var declaration *common.MCPServer
	for _, s := range m.parent.Skills().Snapshot().Skills {
		for _, d := range s.MCPServers {
			if d.Name == name {
				copy := d
				declaration = &copy
			}
		}
	}
	if declaration == nil && m.parent.Config().Skills == nil {
		declaration = &common.MCPServer{Name: name, Transport: "application"}
	}
	if declaration == nil {
		_ = t.Close()
		return fmt.Errorf("undeclared MCP server %q", name)
	}
	c := &connection{config: *declaration, ready: make(chan struct{})}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		_ = t.Close()
		return common.ErrStopped
	}
	if m.connections[name] != nil {
		m.mu.Unlock()
		_ = t.Close()
		return errors.New("MCP server already attached")
	}
	m.connections[name] = c
	m.mu.Unlock()
	c.peer = newPeer(t, m.localHandler(true), false)
	err := m.initialize(ctx, c)
	if err == nil {
		catalog := map[string]remoteTool{}
		for _, tool := range c.tools {
			schema, e := normalizeSchema(tool.InputSchema)
			if e != nil {
				err = e
				break
			}
			tool.InputSchema = schema
			catalog[tool.Name] = tool
		}
		// Validate admitted expectations, including cold browser definitions,
		// rather than imposing the bundled catalog on every declared peer.
		for _, tool := range m.parent.Skills().Snapshot().Tools {
			if tool.Source != "mcp:"+name {
				continue
			}
			if !bytes.Equal(tool.Parameters, catalog[tool.Name].InputSchema) {
				err = fmt.Errorf("MCP schema drift on reconnect: %s", tool.Name)
				break
			}
			if catalog[tool.Name].Ephemeral != tool.Ephemeral {
				err = fmt.Errorf("MCP automatic mode drift on reconnect: %s", tool.Name)
				break
			}
		}
	}
	if err == nil && m.parent.Config().Skills == nil {
		var specs []common.ToolSpec
		for _, tool := range c.tools {
			var spec common.ToolSpec
			spec, err = m.toolSpec(name, tool)
			if err != nil {
				break
			}
			specs = append(specs, spec)
		}
		if err == nil {
			err = m.parent.Skills().RegisterExternal(specs)
		}
		if err == nil {
			m.mu.Lock()
			m.active[name] = c.config
			m.mu.Unlock()
		}
	}

	m.mu.Lock()
	c.err = err
	close(c.ready)
	if err != nil && m.connections[name] == c {
		delete(m.connections, name)
	}
	m.mu.Unlock()
	if err != nil {
		c.peer.close(err)
		return err
	}
	go func() {
		<-c.peer.done
		m.mu.Lock()
		if m.connections[name] == c {
			delete(m.connections, name)
		}
		m.mu.Unlock()
	}()
	return nil
}
func (m *manager) Discover(ctx context.Context, skills []common.Skill) ([]common.ToolSpec, error) {
	// A declaration admits only tool names actually named by that skill. Listing
	// other remote tools is discovery, never a grant of execution authority.
	type allowed struct {
		config common.MCPServer
		names  map[string]bool
	}
	declarations := map[string]*allowed{}
	for _, s := range skills {
		for _, d := range s.MCPServers {
			if d.Name == "" {
				return nil, errors.New("MCP declaration requires name")
			}
			a := declarations[d.Name]
			if a == nil {
				a = &allowed{d, map[string]bool{}}
				declarations[d.Name] = a
			}
			for _, name := range s.Tools {
				a.names[name] = true
			}
		}
	}
	names := make([]string, 0, len(declarations))
	for name := range declarations {
		names = append(names, name)
	}
	sort.Strings(names)
	var specs []common.ToolSpec
	owners := map[string]string{}
	for _, name := range names {
		a := declarations[name]
		// A real attached peer owns its catalog. Only an unattached bundled
		// browser uses cold definitions; unknown required names still fail
		// normal skill admission, and endpoint overrides discover normally.
		m.mu.Lock()
		attached := m.connections[name] != nil
		m.mu.Unlock()
		tools := browserDefinitions(m.endpoint(a.config))
		if attached || tools == nil {
			c, err := m.connect(ctx, a.config)
			if err != nil {
				return nil, fmt.Errorf("MCP %s: %w", name, err)
			}
			tools = c.tools
		}
		for _, t := range tools {
			if t.Ephemeral != "" && t.Ephemeral != "round" && t.Ephemeral != "turn" {
				return nil, fmt.Errorf("invalid ephemeral mode %q", t.Ephemeral)
			}
			if !a.names[t.Name] {
				continue
			}
			if prior := owners[t.Name]; prior != "" {
				return nil, fmt.Errorf("MCP tool %q listed twice (%s, %s)", t.Name, prior, name)
			}
			owners[t.Name] = name
			spec, err := m.toolSpec(name, t)
			if err != nil {
				return nil, err
			}
			specs = append(specs, spec)
		}
	}
	return specs, nil
}
func (m *manager) toolSpec(server string, t remoteTool) (common.ToolSpec, error) {
	if t.Name == "" || strings.TrimSpace(t.Description) == "" {
		return common.ToolSpec{}, errors.New("MCP tool requires name and description")
	}
	if t.Ephemeral != "" && t.Ephemeral != "round" && t.Ephemeral != "turn" {
		return common.ToolSpec{}, fmt.Errorf("invalid ephemeral mode %q", t.Ephemeral)
	}
	if t.Ephemeral != "" && (t.Name == "gui_snapshot" || t.Name == "tts_queue") {
		return common.ToolSpec{}, fmt.Errorf("MCP observation tool %s must not be automatic", t.Name)
	}
	schema, err := normalizeSchema(t.InputSchema)
	if err != nil {
		return common.ToolSpec{}, fmt.Errorf("MCP %s/%s: %w", server, t.Name, err)
	}
	return common.ToolSpec{Ephemeral: t.Ephemeral, ToolDefinition: common.ToolDefinition{Name: t.Name, Description: t.Description, Parameters: schema}, Source: "mcp:" + server, Handler: func(ctx context.Context, owner common.ToolCall, args json.RawMessage) error {
		return m.callTool(ctx, server, t.Name, args, owner)
	}}, nil
}
func (m *manager) callTool(ctx context.Context, server, tool string, args json.RawMessage, out io.Writer) error {
	m.mu.Lock()
	c := m.connections[server]
	_, enabled := m.active[server]
	m.mu.Unlock()
	if !enabled || c == nil {
		return errors.New("MCP server is not active")
	}
	// Startup admission can precede the browser handshake. Never call through
	// a partially initialized peer or a connection removed by pending unload.
	select {
	case <-c.ready:
	case <-ctx.Done():
		return ctx.Err()
	case <-m.ctx.Done():
		return common.ErrStopped
	}
	if c.err != nil {
		return c.err
	}
	m.mu.Lock()
	_, enabled = m.active[server]
	current := m.connections[server] == c
	m.mu.Unlock()
	if !enabled || !current {
		return errors.New("MCP server is not active")
	}
	var result common.MCPCallResult
	if err := c.peer.call(ctx, "tools/call", map[string]any{"name": tool, "arguments": args}, &result); err != nil {
		return err
	}
	if err := writeContent(out, result.Content); err != nil {
		return err
	}
	if result.IsError {
		return errors.New("MCP tool reported isError")
	}
	return nil
}
func writeContent(w io.Writer, content []json.RawMessage) error {
	for i, b := range content {
		if i > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		var c common.MCPContent
		if err := json.Unmarshal(b, &c); err != nil || c.Type == "" {
			return errors.New("invalid MCP content block")
		}
		text := string(b)
		if c.Type == "text" {
			text = c.Text
		}
		if _, err := io.WriteString(w, text); err != nil {
			return err
		}
	}
	return nil
}
func (m *manager) Reconcile(snapshot common.CapabilitySnapshot) {
	active := map[string]common.MCPServer{}
	for _, s := range snapshot.Skills {
		if s.Enabled && !s.PendingUnload {
			for _, d := range s.MCPServers {
				active[d.Name] = d
			}
		}
	}
	m.mu.Lock()
	old := m.active
	m.active = active
	var closing []*connection
	for name, c := range m.connections {
		if _, ok := active[name]; !ok && (c.config.Transport != "websocket" || old[name].Name != "") {
			closing = append(closing, c)
			delete(m.connections, name)
		}
	}
	m.mu.Unlock()
	// Shutdown and skill updates must not wait for transport I/O or child reaping.
	for _, c := range closing {
		go func(c *connection) {
			<-c.ready
			if c.peer != nil {
				c.peer.close(errors.New("MCP skill unloaded"))
			}
		}(c)
	}
}

// Gather invokes declared hidden tools through ordinary job admission, not a
// second executor. A cancelled observer leaves its admitted job running.
func (m *manager) Gather(ctx context.Context, includeTurn bool) (round, turn common.PartList) {
	for _, t := range m.parent.Skills().Snapshot().Tools {
		if !t.Enabled || t.Ephemeral == "" || (t.Ephemeral == "turn" && !includeTurn) {
			continue
		}
		if ctx.Err() != nil {
			break
		}
		parts, err := m.parent.Tools().Collect(ctx, t.Name)
		if ctx.Err() != nil {
			break
		}
		if err != nil {
			parts = common.PartList{common.Text(fmt.Sprintf("[automatic context %s unavailable: %v]", t.Name, err))}
		}
		parts = append(common.PartList{common.Text("[automatic context " + t.Name + "]")}, parts...)
		if t.Ephemeral == "turn" {
			turn = append(turn, parts...)
		} else {
			round = append(round, parts...)
		}
	}
	return
}
func (m *manager) Close() error {
	m.cancel()
	m.mu.Lock()
	m.closed = true
	connections := m.connections
	m.connections = map[string]*connection{}
	m.mu.Unlock()
	for _, c := range connections {
		<-c.ready
		if c.peer != nil {
			c.peer.close(common.ErrStopped)
		}
	}
	return nil
}

// localHandler is per peer: initialization state never leaks across connections.
func (m *manager) localHandler(reverse bool) handler {
	var mu sync.Mutex
	initialized, ready := false, reverse
	return func(ctx context.Context, method string, params json.RawMessage) (any, error) {
		mu.Lock()
		switch method {
		case "initialize":
			if initialized {
				mu.Unlock()
				return nil, &rpcError{-32600, "already initialized"}
			}
			initialized = true
			mu.Unlock()
			return map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]string{"name": "ensemble", "version": "edition-3-ch11"}}, nil
		case "notifications/initialized":
			ready = initialized
			mu.Unlock()
			return nil, nil
		}
		ok := ready
		mu.Unlock()
		// Outgoing sessions can receive the historical reverse-tool extension after
		// our initialize; it is still restricted to the enabled local registry.
		if !ok {
			return nil, &rpcError{-32002, "not initialized"}
		}
		switch method {
		case "ping":
			return map[string]any{}, nil
		case "tools/list", "tools/builtin/list":
			tools := []remoteTool{}
			for _, t := range m.parent.Skills().Snapshot().Tools {
				if t.Enabled && t.Ephemeral == "" && !strings.HasPrefix(t.Source, "mcp:") {
					tools = append(tools, remoteTool{Name: t.Name, Description: t.Description, InputSchema: t.Parameters})
				}
			}
			return map[string]any{"tools": tools}, nil
		case "tools/call", "tools/builtin/call":
			var p struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			if json.Unmarshal(params, &p) != nil || p.Name == "" {
				return nil, &rpcError{-32602, "invalid tool call"}
			}
			if len(p.Arguments) == 0 {
				p.Arguments = json.RawMessage(`{}`)
			}
			allowed := false
			for _, t := range m.parent.Skills().Snapshot().Tools {
				if t.Enabled && t.Ephemeral == "" && t.Name == p.Name && !strings.HasPrefix(t.Source, "mcp:") {
					allowed = true
				}
			}
			refusal := ""
			if !allowed {
				refusal = "MCP local tool is unknown or disabled"
			}
			result := m.parent.Tools().Prepare(ctx, common.Part{Name: p.Name, Args: p.Arguments}, refusal, 0).Await(ctx)
			blocks := []json.RawMessage{}
			for _, part := range result.Parts {
				b, _ := json.Marshal(common.MCPContent{Type: "text", Text: part.Text})
				blocks = append(blocks, b)
			}
			return common.MCPCallResult{Content: blocks, IsError: result.IsError}, nil
		default:
			return nil, &rpcError{-32601, "method not found"}
		}
	}
}
func (m *manager) Serve(ctx context.Context, t common.MCPTransport) error {
	p := newPeer(t, m.localHandler(false), true)
	defer p.close(io.EOF)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.ctx.Done():
		return common.ErrStopped
	case <-p.done:
		p.mu.Lock()
		err := p.err
		p.mu.Unlock()
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
}
