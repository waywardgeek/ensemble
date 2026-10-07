// Package gui is a client-boundary stub, not a browser or WebSocket server.
// A future WebSocket transport attaches here and translates browser requests
// into Submit calls; the core library remains independent of that transport.
package gui

import (
	"context"
	"encoding/json"
	"example.com/ensemble"
	"fmt"
	"sync"
)

type Client struct {
	parent       ensemble.ClientOwner
	agentID      string
	subscription uint64
	mu           sync.Mutex
	closed       bool
	events       []ensemble.Observation
}

func New(parent ensemble.ClientOwner, agentID string) (*Client, error) {
	c := &Client{parent: parent, agentID: agentID}
	id, err := parent.Subscribe(agentID, c)
	if err != nil {
		return nil, err
	}
	c.subscription = id
	return c, nil
}
func (c *Client) Ensemble() ensemble.ClientOwner { return c.parent }
func (c *Client) Observe(event ensemble.Observation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed && event.AgentID == c.agentID {
		c.events = append(c.events, event)
	}
}
func (c *Client) Submit(ctx context.Context, request ensemble.ClientRequest) (ensemble.ClientResult, error) {
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		c.parent.Logf("GUI client closed")
		return ensemble.ClientResult{}, fmt.Errorf("GUI client closed")
	}
	request.AgentID = c.agentID
	return c.parent.Submit(ctx, request)
}
func (c *Client) Observations() []ensemble.Observation {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []ensemble.Observation
	data, _ := json.Marshal(c.events)
	_ = json.Unmarshal(data, &out)
	return out
}
func (c *Client) Close() {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	c.parent.Unsubscribe(c.subscription)
}
