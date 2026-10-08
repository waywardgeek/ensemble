// Package gui provides optional browser transport and reusable public clients.
// Client retains the earlier observation adapter; Server and Connector provide
// the atomic watch browser path. The core library remains independent of both.
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
	updates      chan struct{}
}

func New(parent ensemble.ClientOwner, agentID string) (*Client, error) {
	c := &Client{parent: parent, agentID: agentID, updates: make(chan struct{}, 1)}
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
		select {
		case c.updates <- struct{}{}:
		default:
		}
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

func (c *Client) SubmitPrompt(text string) (ensemble.RequestHandle, error) {
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		return nil, fmt.Errorf("GUI client closed")
	}
	return c.parent.SubmitPrompt(c.agentID, text)
}
func (c *Client) Hint(text string) (ensemble.ControlAck, error) {
	return c.parent.Hint(c.agentID, text)
}
func (c *Client) Interrupt() (ensemble.ControlAck, error) { return c.parent.Interrupt(c.agentID) }
func (c *Client) Status() string                          { return c.parent.SubscriptionStatus(c.subscription) }

func (c *Client) Updates() <-chan struct{} { return c.updates }
