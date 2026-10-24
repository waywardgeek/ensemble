package gui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"

	"ensemble"
)

type relaySession struct {
	stream  ensemble.MCPTransport
	browser *browserTransport
}

// ServeMCP relays newline-delimited JSON-RPC on an application-supplied listener
// to a connected browser. It neither opens a listener nor starts the HTTP server.
// Each TCP session binds to one browser and has a private routing key; caller IDs
// remain unchanged. Server.Close releases listeners and all outstanding sessions.
func (s *Server) ServeMCP(listener net.Listener) error {
	if listener == nil {
		return errors.New("MCP relay requires a listener")
	}
	s.relayMu.Lock()
	if s.relayClosed {
		s.relayMu.Unlock()
		return errors.New("GUI server is closed")
	}
	if s.listeners == nil {
		s.listeners = map[net.Listener]bool{}
	}
	s.listeners[listener] = true
	s.relayMu.Unlock()
	defer func() { listener.Close(); s.relayMu.Lock(); delete(s.listeners, listener); s.relayMu.Unlock() }()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		s.relayMu.Lock()
		if s.relayClosed {
			s.relayMu.Unlock()
			conn.Close()
			return nil
		}
		s.relayNext++
		key := fmt.Sprintf("browser-relay-%d", s.relayNext)
		session := &relaySession{stream: ensemble.NewMCPStream(conn)}
		s.Hub.mu.Lock()
		for c := range s.Hub.clients {
			select {
			case <-c.done:
				continue
			default:
			}
			session.browser = &browserTransport{client: c, agent: key, incoming: make(chan json.RawMessage, 64), done: make(chan struct{})}
			break
		}
		s.Hub.mu.Unlock()
		if s.relays == nil {
			s.relays = map[string]*relaySession{}
		}
		s.relays[key] = session
		s.relayWG.Add(1)
		s.relayMu.Unlock()
		go s.runRelay(key, session)
	}
}
func (s *Server) runRelay(key string, session *relaySession) {
	defer s.relayWG.Done()
	defer func() {
		session.stream.Close()
		if session.browser != nil {
			session.browser.Close()
		}
		s.relayMu.Lock()
		delete(s.relays, key)
		s.relayMu.Unlock()
	}()
	if session.browser == nil {
		for {
			raw, err := session.stream.Recv()
			if err != nil {
				return
			}
			var request struct {
				ID json.RawMessage `json:"id"`
			}
			if json.Unmarshal(raw, &request) != nil || len(request.ID) == 0 {
				continue
			}
			reply, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": request.ID, "error": map[string]any{"code": -32000, "message": "no browser connected"}})
			if session.stream.Send(reply) != nil {
				return
			}
		}
	}
	// Both directions share existing MCP framing and the browser's socket writer.
	// Closing either end releases the other direction, including an unanswered RPC.
	returned := make(chan struct{})
	go func() {
		defer close(returned)
		defer session.stream.Close()
		for {
			raw, err := session.browser.Recv()
			if err != nil {
				return
			}
			if session.stream.Send(raw) != nil {
				return
			}
		}
	}()
	defer func() { session.browser.Close(); session.stream.Close(); <-returned }()
	for {
		raw, err := session.stream.Recv()
		if err != nil {
			return
		}
		if session.browser.Send(raw) != nil {
			return
		}
	}
}
func (s *Server) relayResponse(c *client, cmd command) (bool, error) {
	s.relayMu.Lock()
	session := s.relays[cmd.AgentID]
	s.relayMu.Unlock()
	if session == nil {
		return false, nil
	}
	if session.browser == nil || session.browser.client != c {
		return true, errors.New("MCP relay belongs to another browser")
	}
	if !json.Valid(cmd.Payload) {
		return true, errors.New("invalid JSON-RPC response")
	}
	t := session.browser
	select {
	case <-t.done:
		return true, io.EOF
	case <-c.done:
		return true, io.EOF
	case t.incoming <- append(json.RawMessage(nil), cmd.Payload...):
		return true, nil
	default:
		t.Close()
		return true, errors.New("MCP relay inbound queue overflow")
	}
}
func (s *Server) closeRelays() {
	s.relayMu.Lock()
	s.relayClosed = true
	for listener := range s.listeners {
		listener.Close()
	}
	for _, session := range s.relays {
		session.stream.Close()
		if session.browser != nil {
			session.browser.Close()
		}
	}
	s.relayMu.Unlock()
	s.relayWG.Wait()
}
