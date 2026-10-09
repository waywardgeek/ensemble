package mcp

import (
	"context"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type operation struct {
	ctx                                       context.Context
	method                                    string
	params                                    map[string]any
	id                                        string
	staged                                    time.Time
	noticeAt                                  time.Time
	issued, sending, sent, cancelled, settled bool
	done                                      chan struct{}
	result                                    map[string]any
	err                                       error
}
type connection struct {
	parent          common.MCPConnectionOwner
	spec            common.MCPConnectionSpec
	generation      uint64
	mu              sync.Mutex
	state, code     string
	definitions     map[string]common.MCPDefinition
	issued          uint64
	slots           [64]*operation
	pending         map[string]*operation
	transport       common.MCPTransport
	wake            chan struct{}
	stopped, joined chan struct{}
	once            sync.Once
	workers         sync.WaitGroup
	stale           uint64
}

func newConnection(s common.MCPConnectionOwner, spec common.MCPConnectionSpec, g uint64) *connection {
	return &connection{parent: s, spec: spec, generation: g, state: "preparing", definitions: map[string]common.MCPDefinition{}, pending: map[string]*operation{}, wake: make(chan struct{}, 1), stopped: make(chan struct{}), joined: make(chan struct{})}
}
func (c *connection) Service() common.MCPService { return c.parent }
func (c *connection) Snapshot() common.MCPConnectionSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := 0
	if c.state == "ready" {
		count = len(c.definitions)
	}
	return common.MCPConnectionSnapshot{Key: c.spec.Key, State: c.state, Generation: c.generation, ProtocolVersion: common.MCPVersion, DiscoveredCount: count, ErrorCode: c.code}
}
func (c *connection) matches(b common.MCPBinding) bool {
	c.mu.Lock()
	d, ok := c.definitions[b.RemoteName]
	c.mu.Unlock()
	return ok && equalDefinition(c.parent, d, b.Definition)
}
func (c *connection) signal() {
	select {
	case c.wake <- struct{}{}:
	default:
	}
}
func (c *connection) start(t common.MCPTransport) {
	c.mu.Lock()
	c.transport = t
	c.mu.Unlock()
	c.workers.Add(3)
	go c.read()
	go c.write()
	go c.deadlines()
	go func() {
		<-c.stopped
		_ = t.Close()
		c.workers.Wait()
		c.mu.Lock()
		for i := range c.slots {
			c.slots[i] = nil
		}
		c.pending = map[string]*operation{}
		c.mu.Unlock()
		c.parent.ConnectionChanged()
		close(c.joined)
	}()
}
func (c *connection) failBeforeStart(code string) {
	c.stop("failed", code)
	c.parent.ConnectionChanged()
	close(c.joined)
}
func (c *connection) stop(state, code string) {
	c.once.Do(func() {
		c.mu.Lock()
		c.state = state
		c.code = code
		for _, o := range c.slots {
			if o != nil && !o.settled {
				message := "MCP connection closed."
				if o.issued {
					message += " The external outcome may be unknown."
				}
				c.settle(o, nil, failure(code, message))
			}
		}
		close(c.stopped)
		c.mu.Unlock()
		c.signal()
	})
}
func (c *connection) settle(o *operation, result map[string]any, err error) {
	if o.settled {
		return
	}
	o.settled = true
	o.result = result
	o.err = err
	if o.id != "" {
		delete(c.pending, o.id)
	}
	close(o.done)
}
func (c *connection) release(o *operation) {
	for i, v := range c.slots {
		if v == o {
			c.slots[i] = nil
			return
		}
	}
}
func (c *connection) cancel(o *operation, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if o.settled {
		return
	}
	code := "mcp_unavailable"
	msg := "Remote call canceled."
	if err == context.DeadlineExceeded {
		code = "mcp_timeout"
		msg = "Remote call deadline exceeded. The external outcome may be unknown."
	}
	o.cancelled = o.issued
	o.noticeAt = time.Now()
	c.settle(o, nil, failure(code, msg))
	if !o.issued {
		c.release(o)
	}
	c.signal()
}
func (c *connection) rpc(ctx context.Context, method string, params map[string]any) (map[string]any, error) {
	c.mu.Lock()
	if c.state != "ready" && c.state != "preparing" {
		c.mu.Unlock()
		return nil, failure("mcp_unavailable", "Connection is unavailable.")
	}
	index := -1
	for i, o := range c.slots {
		if o == nil {
			index = i
			break
		}
	}
	if index < 0 {
		c.mu.Unlock()
		return nil, failure("mcp_capacity", "Connection operation capacity reached.")
	}
	if err := ctx.Err(); err != nil {
		c.mu.Unlock()
		return nil, failure("mcp_timeout", "Remote call canceled before issue. The external outcome may be unknown.")
	}
	o := &operation{ctx: ctx, method: method, params: params, staged: time.Now(), done: make(chan struct{})}
	c.slots[index] = o
	c.mu.Unlock()
	c.signal()
	select {
	case <-o.done:
	case <-ctx.Done():
		c.cancel(o, ctx.Err())
		<-o.done
	}
	c.mu.Lock()
	result, err := o.result, o.err
	c.mu.Unlock()
	return result, err
}
func (c *connection) write() {
	defer c.workers.Done()
	for {
		select {
		case <-c.stopped:
			return
		default:
		}
		c.mu.Lock()
		var o *operation
		notice := false
		for _, x := range c.slots {
			if x != nil && !x.sending && (x.cancelled && x.sent || !x.issued && !x.settled) {
				o = x
				notice = x.cancelled
				break
			}
		}
		if o == nil {
			c.mu.Unlock()
			select {
			case <-c.wake:
				continue
			case <-c.stopped:
				return
			}
		}
		if !notice && o.ctx.Err() != nil {
			c.mu.Unlock()
			c.cancel(o, o.ctx.Err())
			continue
		}
		var data []byte
		var err error
		deadline := o.staged.Add(time.Second)
		if notice {
			deadline = o.noticeAt.Add(time.Second)
			data = []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":%q,"reason":"Cancelled by client"}}`, o.id))
		} else {
			if c.issued == ^uint64(0) {
				c.settle(o, nil, failure("mcp_limit", "RPC identities exhausted; explicitly reopen."))
				c.release(o)
				c.mu.Unlock()
				continue
			}
			id := "rpc-" + strconv.FormatUint(c.issued+1, 10)
			params := map[string]any{}
			for k, v := range o.params {
				params[k] = v
			}
			params["_meta"] = map[string]any{"io.modelcontextprotocol/protocolVersion": common.MCPVersion, "io.modelcontextprotocol/clientCapabilities": map[string]any{}, "io.modelcontextprotocol/clientInfo": map[string]any{"name": "Ensemble", "version": "edition-2-ch11"}}
			data, err = json.Marshal(struct {
				JSONRPC string `json:"jsonrpc"`
				ID      string `json:"id"`
				Method  string `json:"method"`
				Params  any    `json:"params"`
			}{"2.0", id, o.method, params})
			if err==nil { _,err=c.parent.Ensemble().JSON().Parse(data,messageBounds()) }
 if err != nil || len(data) > common.MCPMessageLimit {
				c.settle(o, nil, failure("mcp_limit", "Outgoing message limit."))
				c.release(o)
				c.mu.Unlock()
				continue
			}
			if o.ctx.Err()!=nil { c.mu.Unlock(); c.cancel(o,o.ctx.Err());continue }
 if time.Now().After(deadline) {
				c.mu.Unlock()
				c.stop("failed", "mcp_transport")
				return
			}
			c.issued++
			o.id = id
			o.issued = true
			c.pending[id] = o
		}
		o.sending = true
		c.mu.Unlock()
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		if notice {
			err = c.transport.Abandon(ctx, o.id, data)
		} else {
			err = c.transport.Send(ctx, common.MCPMessage{RequestID: o.id, JSON: data})
		}
		cancel()
		if err != nil {
			c.stop("failed", "mcp_transport")
			return
		}
		c.mu.Lock()
		o.sending = false
		if notice {
			o.cancelled = false
			c.release(o)
		} else {
			o.sent = true
			if o.settled && !o.cancelled {
				c.release(o)
			}
		}
		c.mu.Unlock()
		c.signal()
	}
}
func (c *connection) deadlines() {
	defer c.workers.Done()
	timer := time.NewTicker(5 * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case <-c.stopped:
			return
		case now := <-timer.C:
			c.mu.Lock()
			stall := false
			for _, o := range c.slots {
				if o == nil {
					continue
				}
				if !o.sent && now.Sub(o.staged) >= time.Second || o.cancelled && now.Sub(o.noticeAt) >= time.Second {
					stall = true
					break
				}
			}
			c.mu.Unlock()
			if stall {
				c.stop("failed", "mcp_transport")
				return
			}
		}
	}
}
func (c *connection) read() {
	defer c.workers.Done()
	for {
		raw, err := c.transport.Receive(context.Background())
		if err != nil {
			c.stop("failed", "mcp_transport")
			return
		}
		v, err := c.parent.Ensemble().JSON().Parse(raw, messageBounds())
		if err != nil {
			c.stop("failed", "mcp_protocol")
			return
		}
		m, ok := v.(map[string]any)
		if !ok || m["jsonrpc"] != "2.0" {
			c.stop("failed", "mcp_protocol")
			return
		}
		if method, exists := m["method"]; exists {
			_, hasID := m["id"]
			_, stringMethod := method.(string)
			_, objectParams := m["params"].(map[string]any)
			if hasID || !stringMethod || !objectParams || !fields(m, "jsonrpc", "method", "params") {
				c.stop("failed", "mcp_protocol")
				return
			}
			continue
		}
		id, ok := m["id"].(string)
		if !ok || !fields(m, "jsonrpc", "id", "result", "error") {
			c.stop("failed", "mcp_protocol")
			return
		}
		n, err := strconv.ParseUint(strings.TrimPrefix(id, "rpc-"), 10, 64)
		if err != nil || n == 0 || id != "rpc-"+strconv.FormatUint(n, 10) {
			c.stop("failed", "mcp_protocol")
			return
		}
		result, hasResult := m["result"]
		remote, hasError := m["error"]
		if hasResult == hasError {
			c.stop("failed", "mcp_protocol")
			return
		}
		var out map[string]any
		var opErr error
		if hasResult {
			out, ok = result.(map[string]any)
			if !ok {
				c.stop("failed", "mcp_protocol")
				return
			}
			if typ, present := out["resultType"]; present && typ != "complete" {
				if typ == "input_required" {
					opErr = failure("mcp_unsupported_result", "Remote input_required result is unsupported.")
				} else {
					c.stop("failed", "mcp_protocol")
					return
				}
			}
		} else {
			obj, ok := remote.(map[string]any)
			if !ok || !fields(obj, "code", "message", "data") {
				c.stop("failed", "mcp_protocol")
				return
			}
			code, ok := obj["code"].(json.Number)
			_, text := obj["message"].(string)
			j := c.parent.Ensemble().JSON()
			if !ok || !text || !j.Integral(code) || j.Compare(code, "-2147483648") < 0 || j.Compare(code, "2147483647") > 0 {
				c.stop("failed", "mcp_protocol")
				return
			}
			// The range check above bounds decimal expansion to ten digits.
			canonical := j.Number(string(code))
			parts := strings.Split(canonical, "e")
			decimal := parts[0]
			if len(parts) == 2 {
				power, _ := strconv.Atoi(parts[1])
				decimal += strings.Repeat("0", power)
			}
			opErr = failure("mcp_remote_error", "Remote JSON-RPC error "+decimal+".")
		}
		c.mu.Lock()
		if n > c.issued {
			c.mu.Unlock()
			c.stop("failed", "mcp_protocol")
			return
		}
		o := c.pending[id]
		if o != nil {
			c.settle(o, out, opErr)
			if o.sent && !o.sending {
				c.release(o)
			}
		} else {
			if c.stale < ^uint64(0) {
				c.stale++
			}
		}
		c.mu.Unlock()
	}
}
func fields(m map[string]any, allowed ...string) bool {
	for k := range m {
		found := false
		for _, a := range allowed {
			if k == a {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
