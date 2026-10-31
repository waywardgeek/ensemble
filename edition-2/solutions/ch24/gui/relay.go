package gui

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync"
)

// ServeMCP relays JSONL requests to one connected browser. The application
// supplies a listener (normally loopback) and chooses who can reach it. Close
// stops this listener and its clients. Without a browser, requests receive a
// correlated error; a silent browser waits until it, the client or Server closes.
func (s *Server) ServeMCP(listener net.Listener) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		listener.Close()
		return net.ErrClosed
	}
	s.relays[listener] = true
	s.mu.Unlock()
	defer func() {
		listener.Close()
		s.mu.Lock()
		delete(s.relays, listener)
		s.mu.Unlock()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			conn.Close()
			return net.ErrClosed
		}
		s.relayConnections[conn] = true
		s.workers.Add(1)
		s.mu.Unlock()
		go s.serveMCPConnection(conn)
	}
}

func (s *Server) serveMCPConnection(conn net.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	var requests sync.WaitGroup
	var writes sync.Mutex
	defer func() {
		cancel()
		conn.Close()
		requests.Wait()
		s.mu.Lock()
		delete(s.relayConnections, conn)
		s.mu.Unlock()
		s.workers.Done()
	}()
	decoder, encoder := json.NewDecoder(conn), json.NewEncoder(conn)
	for {
		var request map[string]json.RawMessage
		if decoder.Decode(&request) != nil {
			return
		}
		// Keep reading while requests wait: EOF must cancel them, and replies
		// may arrive out of order. Only complete output messages share a lock.
		requests.Add(1)
		go func() {
			defer requests.Done()
			id := request["id"]
			response, err := s.browserRequest(ctx, request)
			if id == nil {
				return
			}
			if err != nil {
				response, _ = json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id,
					"error": map[string]any{"code": -32000, "message": err.Error()}})
			}
			writes.Lock()
			defer writes.Unlock()
			if encoder.Encode(response) != nil {
				conn.Close()
			}
		}()
	}
}

func (s *Server) browserRequest(ctx context.Context, request map[string]json.RawMessage) (json.RawMessage, error) {
	s.mu.Lock()
	var browser *WSTransport
	for _, tunnel := range s.tunnels {
		browser = tunnel
		break
	}
	s.mu.Unlock()
	if browser == nil {
		return nil, errors.New("no browser connected")
	}
	return browser.relay(ctx, request)
}
