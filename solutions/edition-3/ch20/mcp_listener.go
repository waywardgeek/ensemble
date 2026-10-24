package ensemble

import (
	"context"
	"errors"
	"net"
	"sync"
)

// MCPListener belongs to its embedding command, independently of Agent lifetime.
// Close joins accepted sessions before the caller shuts down the Agent.
type MCPListener struct {
	listener net.Listener
	cancel   context.CancelFunc
	done     chan struct{}
	err      error // published by closing done
}

func StartMCPListener(ctx context.Context, a Agent, address string) (*MCPListener, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, errors.New("MCP listener requires a loopback host and port")
	}
	if host == "localhost" {
		host = "127.0.0.1"
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return nil, errors.New("MCP listener must bind loopback")
	}
	l, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	s := &MCPListener{listener: l, cancel: cancel, done: make(chan struct{})}
	go func() {
		stop := context.AfterFunc(ctx, func() { _ = l.Close() })
		defer stop()
		var sessions sync.WaitGroup
		defer close(s.done)
		defer sessions.Wait()
		defer cancel()
		for {
			c, err := l.Accept()
			if err != nil {
				if ctx.Err() == nil {
					s.err = err
				}
				return
			}
			sessions.Add(1)
			go func() { defer sessions.Done(); _ = ServeMCP(ctx, a, NewMCPStream(c)) }()
		}
	}()
	return s, nil
}
func (s *MCPListener) Addr() net.Addr { return s.listener.Addr() }
func (s *MCPListener) Wait() error    { <-s.done; return s.err }
func (s *MCPListener) Close() error   { s.cancel(); _ = s.listener.Close(); return s.Wait() }
