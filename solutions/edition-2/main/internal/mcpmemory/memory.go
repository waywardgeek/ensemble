// Package mcpmemory adapts an application-owned complete-message endpoint.
package mcpmemory

import (
	"context"
	"example.com/ensemble/internal/common"
	"fmt"
	"sync"
)

type Constructor struct {
	source common.MCPMessageEndpointSource
}

func New(source common.MCPMessageEndpointSource) *Constructor { return &Constructor{source: source} }
func (c *Constructor) Open(ctx context.Context, parent common.MCPConnection) (common.MCPTransport, error) {
	if c.source == nil {
		return nil, fmt.Errorf("memory endpoint source required")
	}
	e, err := c.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &transport{parent: parent, endpoint: e}, nil
}

type transport struct {
	parent   common.MCPConnection
	endpoint common.MCPMessageEndpoint
	mu       sync.Mutex
	closed   bool
	wg       sync.WaitGroup
	once     sync.Once
	err      error
}

func (t *transport) Connection() common.MCPConnection { return t.parent }
func (t *transport) begin() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return false
	}
	t.wg.Add(1)
	return true
}
func (t *transport) Send(ctx context.Context, m common.MCPMessage) error {
	if !t.begin() {
		return fmt.Errorf("memory channel closed")
	}
	defer t.wg.Done()
	if len(m.JSON) > common.MCPMessageLimit {
		return fmt.Errorf("message limit")
	}
	return t.endpoint.Send(ctx, append([]byte(nil), m.JSON...))
}
func (t *transport) Receive(ctx context.Context) ([]byte, error) {
	if !t.begin() {
		return nil, fmt.Errorf("memory channel closed")
	}
	defer t.wg.Done()
	b, e := t.endpoint.Receive(ctx)
	if len(b) > common.MCPMessageLimit {
		return nil, fmt.Errorf("message limit")
	}
	return append([]byte(nil), b...), e
}
func (t *transport) Abandon(ctx context.Context, id string, b []byte) error {
	return t.Send(ctx, common.MCPMessage{RequestID: id, JSON: b})
}
func (t *transport) Diagnostics() common.MCPDiagnostics { return common.MCPDiagnostics{} }
func (t *transport) Close() error {
	t.once.Do(func() { t.mu.Lock(); t.closed = true; t.mu.Unlock(); t.err = t.endpoint.Close(); t.wg.Wait() })
	return t.err
}
