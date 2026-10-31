package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"ensemble/internal/common"
)

type connection struct {
	skill   string
	server  common.MCPServer
	allowed []string
	client  *client
	ready   bool
}
type manager struct {
	parent      common.Agent
	mu          sync.Mutex
	connections []*connection
	closed      bool
	browser     common.MCPTransport
}

// New creates the Agent's connection owner; transports remain replaceable I/O.
func New(parent common.Agent) common.MCP {
	if parent == nil {
		panic("MCP requires Agent")
	}
	return &manager{parent: parent}
}

// Agent exposes the owner rather than storing copies of its registry or logger.
func (m *manager) Agent() common.Agent { return m.parent }

// StartSkill records grants before connecting. Discovery never expands allowed.
// WebSocket declarations wait until a real browser supplies their transport.
func (m *manager) StartSkill(skill string, servers []common.MCPServer, allowed []string) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return errClosed
	}
	for _, server := range servers {
		if server.Name == "" || (server.Transport != "stdio" && server.Transport != "websocket" && server.Transport != "url") || (server.Transport == "stdio" && server.Command == "") || (server.Transport == "url" && server.URL == "") {
			m.mu.Unlock()
			return fmt.Errorf("invalid MCP server declaration")
		}
		for _, existing := range m.connections {
			if existing.server.Name == server.Name {
				m.mu.Unlock()
				return fmt.Errorf("MCP server %q already declared", server.Name)
			}
		}
	}
	for _, server := range servers {
		m.connections = append(m.connections, &connection{skill: skill, server: server, allowed: append([]string{}, allowed...)})
	}
	m.mu.Unlock()
	for _, server := range servers {
		var t common.MCPTransport
		var err error
		if server.Transport == "websocket" {
			m.mu.Lock()
			t = m.browser
			m.mu.Unlock()
			if t == nil {
				continue
			}
		} else if server.Transport == "url" {
			t = NewHTTP(m, server)
		} else {
			t, err = startStdio(m.parent, server)
		}
		if err == nil {
			err = m.Connect(context.Background(), server.Name, t)
		}
		if err != nil {
			m.StopSkill(skill)
			return err
		}
	}
	return nil
}

// Connect joins an external transport to a declared peer (or an unskilled Agent).
// The client is published only after discovery; incomplete tools never leak out.
func (m *manager) Connect(ctx context.Context, name string, transport common.MCPTransport) error {
	m.mu.Lock()
	var entry *connection
	for _, c := range m.connections {
		if c.server.Name == name {
			entry = c
			break
		}
	}
	if m.closed || (entry == nil && m.parent.Skills() != nil) {
		m.mu.Unlock()
		transport.Close()
		return fmt.Errorf("MCP server %q is not admitted", name)
	}
	if entry == nil {
		entry = &connection{server: common.MCPServer{Name: name}}
		m.connections = append(m.connections, entry)
	}
	if entry.client != nil {
		select {
		case <-entry.client.done:
		default:
			m.mu.Unlock()
			transport.Close()
			return fmt.Errorf("MCP server %q is connected", name)
		}
	}
	c := newClient(m, transport)
	entry.client = c
	entry.ready = false
	m.mu.Unlock()
	if err := c.initialize(ctx); err != nil {
		c.close()
		return err
	}
	m.mu.Lock()
	// The lock publishes discovered schemas to concurrent renderer snapshots.
	// Unload may have closed this connection during its handshake.
	entry.ready = true
	m.mu.Unlock()
	return nil
}

// Tools returns copies so a caller never observes the manager's mutable slice.
func (m *manager) Tools() []common.ToolDefinition {
	m.mu.Lock()
	defer m.mu.Unlock()
	var tools []common.ToolDefinition
	for _, entry := range m.connections {
		c := entry.client
		if c == nil || !entry.ready {
			continue
		}
		select {
		case <-c.done:
			continue
		default:
		}
		for _, info := range c.tools {
			allowed := entry.allowed == nil
			for _, name := range entry.allowed {
				if name == info.Name {
					allowed = true
				}
			}
			if !allowed {
				continue
			}
			info := info
			tools = append(tools, common.ToolDefinition{ToolDeclaration: common.ToolDeclaration{Name: info.Name, Description: info.Description, Schema: info.Schema}, Ephemeral: info.Ephemeral,
				Run: func(call common.ToolContext, args json.RawMessage) (string, error) {
					return c.call(call.Job().Context(), info.Name, args)
				}})
		}
	}
	return tools
}

// StopSkill detaches tools before closing peers. Compiled skill capabilities and
// retained instructions still follow the earlier pending-unload policy.
func (m *manager) StopSkill(skill string) error {
	m.mu.Lock()
	var stopped []*client
	var retained []*connection
	for _, entry := range m.connections {
		if entry.skill == skill {
			if entry.client != nil {
				stopped = append(stopped, entry.client)
			}
		} else {
			retained = append(retained, entry)
		}
	}
	m.connections = retained
	m.mu.Unlock()
	var failures []error
	for _, c := range stopped {
		failures = append(failures, c.close())
	}
	return errors.Join(failures...)
}

// Close releases peers before Agent joins its outstanding RPC job workers.
func (m *manager) Close() error {
	m.mu.Lock()
	m.closed = true
	entries := m.connections
	m.connections = nil
	m.mu.Unlock()
	var failures []error
	for _, entry := range entries {
		if entry.client != nil {
			failures = append(failures, entry.client.close())
		}
	}
	return errors.Join(failures...)
}

// Done lets the dedicated pipe command wait for its peer without reading its wire.
func (m *manager) Done(name string) <-chan struct{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.connections {
		if entry.server.Name == name && entry.client != nil {
			return entry.client.done
		}
	}
	done := make(chan struct{})
	close(done)
	return done
}

// AttachBrowser retains the first live GUI boundary until its connection closes.
// With skills, discovery waits for an actual websocket declaration.
func (m *manager) AttachBrowser(ctx context.Context, transport common.MCPTransport) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		transport.Close()
		return errClosed
	}
	name := ""
	if m.parent.Skills() == nil {
		name = "browser"
	} else {
		for _, entry := range m.connections {
			if entry.server.Transport == "websocket" {
				name = entry.server.Name
				break
			}
		}
	}
	if name == "" {
		previous := m.browser
		m.browser = transport
		m.mu.Unlock()
		if previous != nil {
			previous.Close()
		}
		return nil
	}
	m.mu.Unlock()
	if err := m.Connect(ctx, name, transport); err != nil {
		return err
	}
	m.mu.Lock()
	m.browser = transport
	m.mu.Unlock()
	return nil
}
