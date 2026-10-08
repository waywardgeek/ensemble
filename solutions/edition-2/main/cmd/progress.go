package main

import (
	"bufio"
	"fmt"
	"sync"
	"time"

	"example.com/ensemble"
)

// This client-owned bridge bounds its own storage too. A slow terminal can lose
// display, but the reliable handle is independent and recovers the final answer.
type progress struct {
	parent             ensemble.ClientOwner
	agentID            string
	subscription       uint64
	mu                 sync.Mutex
	queue              []ensemble.Observation
	wake               chan struct{}
	overflow, reported bool
	last               string
	ended              map[string]bool
	streamed           map[string]bool
}

func newProgress(owner ensemble.ClientOwner, agentID string) (*progress, error) {
	p := &progress{parent: owner, agentID: agentID, wake: make(chan struct{}, 1), ended: map[string]bool{}, streamed: map[string]bool{}}
	id, err := owner.Subscribe(agentID, p)
	p.subscription = id
	return p, err
}
func (p *progress) Observe(o ensemble.Observation) {
	switch o.Kind {
	case "model_begin", "part_delta", "part_final", "model_end", "turn_ended":
	default:
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.overflow {
		return
	}
	if len(p.queue) == 256 {
		p.overflow = true
		p.queue = nil
	} else {
		p.queue = append(p.queue, o)
	}
	select {
	case p.wake <- struct{}{}:
	default:
	}
}
func (p *progress) drain(show func(ensemble.Observation) error, gap func() error) error {
	p.mu.Lock()
	queue := p.queue
	p.queue = nil
	overflow := p.overflow
	p.mu.Unlock()
	overflow = overflow || p.parent.SubscriptionStatus(p.subscription) == "overflow"
	if overflow && !p.reported {
		p.reported = true
		p.parent.Unsubscribe(p.subscription)
		return gap()
	}
	if p.reported {
		return nil
	}
	for _, o := range queue {
		if o.Kind == "turn_ended" {
			p.ended[o.Event.Turn.RequestID] = true
			continue
		}
		if o.Kind == "model_begin" {
			p.streamed[o.RequestID] = o.Delivery == "stream"
		}
		if err := show(o); err != nil {
			return err
		}
	}
	return nil
}
func (p *progress) beforeCompletion(agent *ensemble.Agent, c ensemble.Completion, show func(ensemble.Observation) error, gap func() error) error {
	// Queued requests have no turn. Active completions always follow turn_ended;
	// persistence faults may lack it, and recover through the reliable handle.
	needsEnd := false
	for _, event := range agent.Events() {
		if event.Type == "turn_ended" && event.Turn.RequestID == c.RequestID {
			needsEnd = true
			break
		}
	}
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := p.drain(show, gap); err != nil {
			return err
		}
		if p.reported || p.ended[c.RequestID] || !needsEnd {
			return nil
		}
		select {
		case <-p.wake:
		case <-ticker.C:
		}
	}
}
func (p *progress) chat(out *bufio.Writer, o ensemble.Observation) error {
	switch o.Kind {
	case "model_begin":
		fmt.Fprintf(out, "\nRequest %s / %s (%s)\n", o.RequestID, o.OperationID, o.Delivery)
	case "part_delta":
		label := "Assistant"
		switch o.Channel {
		case "thinking":
			label = "thinking"
		case "tool_name", "tool_args":
			label = "proposed tool"
		}
		key := fmt.Sprintf("%s/%d/%s", o.OperationID, o.PartID, o.Channel)
		if p.last != key {
			fmt.Fprintf(out, "\n[%s %s part %d] ", o.RequestID, label, o.PartID)
			p.last = key
		}
		fmt.Fprint(out, o.Text)
	case "part_final": // Plain chat prints the reliable complete answer once.
	case "model_end":
		p.last = ""
		if o.Accepted {
			fmt.Fprintln(out)
		} else {
			fmt.Fprintf(out, "\nIncomplete display: %s: %s\n", o.Code, o.Message)
		}
	}
	return out.Flush()
}
func observationRecord(o ensemble.Observation) map[string]any {
	m := map[string]any{"kind": o.Kind, "agent_id": o.AgentID, "request_id": o.RequestID, "operation_id": o.OperationID}
	switch o.Kind {
	case "model_begin":
		m["delivery"] = o.Delivery
	case "part_delta":
		m["part_id"] = o.PartID
		m["channel"] = o.Channel
		m["text"] = o.Text
	case "part_final":
		m["part_id"] = o.PartID
		m["response_seq"] = o.ResponseSeq
		m["part_index"] = o.PartIndex
		m["part"] = o.Part
	case "model_end":
		m["accepted"] = o.Accepted
		if o.Accepted {
			m["response_seq"] = o.ResponseSeq
		} else {
			m["code"] = o.Code
			m["message"] = o.Message
		}
	}
	return m
}
