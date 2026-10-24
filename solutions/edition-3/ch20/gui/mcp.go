package gui

import (
	"encoding/json"
	"errors"
	"io"
	"sync"

	"ensemble"
)

// browserTransport shares the existing socket writer, never replay storage.
// ready publishes the result of this particular tunnel's initialization.
type browserTransport struct {
	client      *client
	agent       string
	incoming    chan json.RawMessage
	done, ready chan struct{}
	err         error
	once        sync.Once
}

func (t *browserTransport) Recv() (json.RawMessage, error) {
	select {
	case b := <-t.incoming:
		return b, nil
	case <-t.done:
		return nil, io.EOF
	case <-t.client.done:
		return nil, io.EOF
	}
}
func (t *browserTransport) Send(b json.RawMessage) error {
	wire, err := json.Marshal(Wire{Type: "jsonrpc", AgentID: t.agent, Payload: b})
	if err != nil {
		return err
	}
	select {
	case <-t.done:
		return io.EOF
	case <-t.client.done:
		return io.EOF
	case t.client.queue <- wire:
		return nil
	}
}
func (t *browserTransport) Close() error {
	t.once.Do(func() {
		close(t.done)
		wire, _ := json.Marshal(Wire{Type: "mcp_closed", AgentID: t.agent})
		select {
		case <-t.client.done:
		case t.client.queue <- wire:
		default:
			t.client.stop()
		}
	})
	return nil
}

// The client may override a skill's browser declaration with a pipe or socket.
// Only an effective WebSocket endpoint belongs to this GUI tunnel.
func browserWebSocket(a ensemble.Agent, declaration ensemble.MCPServer) bool {
	if declaration.Name != "browser-debug" {
		return false
	}
	for _, endpoint := range a.Config().MCPServers {
		if endpoint.Name == declaration.Name {
			return endpoint.Transport == "websocket"
		}
	}
	return declaration.Transport == "websocket"
}

func (s *Server) openMCP(c *client, a ensemble.Agent) (*browserTransport, error) {
	// Browser-originated names never choose an endpoint or grant authority.
	declared := false
	for _, skill := range a.Skills().Snapshot().Skills {
		for _, d := range skill.MCPServers {
			if browserWebSocket(a, d) {
				declared = true
			}
		}
	}
	if !declared {
		return nil, errors.New("browser-debug is not declared")
	}
	c.mcpMu.Lock()
	defer c.mcpMu.Unlock()
	t := c.mcp[a.ID()]
	if t != nil {
		select {
		case <-t.done:
			t = nil
		default:
		}
	}
	if t != nil {
		return t, nil
	}
	t = &browserTransport{client: c, agent: a.ID(), incoming: make(chan json.RawMessage, 64), done: make(chan struct{}), ready: make(chan struct{})}
	if c.mcp == nil {
		c.mcp = map[string]*browserTransport{}
	}
	c.mcp[a.ID()] = t
	go func() {
		t.err = a.AttachMCPConnection("browser-debug", t)
		close(t.ready)
		if t.err != nil {
			s.Hub.sendError(c, a.ID(), t.err)
		}
	}()
	return t, nil
}
func (s *Server) mcp(c *client, cmd command) error {
	if !s.Hub.allows(cmd.AgentID) {
		return errors.New("Agent is outside this GUI view")
	}
	a, ok := s.Root.Agent(cmd.AgentID)
	if !ok {
		return errors.New("unknown agent")
	}
	if cmd.Type == "mcp_open" {
		_, err := s.openMCP(c, a)
		return err
	}
	c.mcpMu.Lock()
	t := c.mcp[cmd.AgentID]
	c.mcpMu.Unlock()
	if t == nil {
		return errors.New("MCP tunnel is not open")
	}
	if !json.Valid(cmd.Payload) {
		return errors.New("invalid JSON-RPC envelope")
	}
	select {
	case <-t.done:
		return io.EOF
	case <-c.done:
		return io.EOF
	case t.incoming <- append(json.RawMessage(nil), cmd.Payload...):
		return nil
	default:
		t.Close()
		return errors.New("MCP inbound queue overflow")
	}
}

// Loading can need responses on this socket. Only the existing command worker
// waits for discovery, never its reader or the Agent actor.
func (s *Server) prepareBrowserSkill(c *client, a ensemble.Agent, name string) error {
	for _, skill := range a.Skills().Snapshot().Skills {
		if skill.Name != name {
			continue
		}
		for _, d := range skill.MCPServers {
			if !browserWebSocket(a, d) {
				continue
			}
			if skill.PendingUnload {
				c.mcpMu.Lock()
				old := c.mcp[a.ID()]
				c.mcpMu.Unlock()
				if old != nil {
					old.Close()
				}
			}
			t, err := s.openMCP(c, a)
			if err != nil {
				return err
			}
			select {
			case <-t.ready:
				return t.err
			case <-c.done:
				return io.EOF
			}
		}
	}
	return nil
}

var _ ensemble.MCPTransport = (*browserTransport)(nil)
