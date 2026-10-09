package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// replyRouter matches browser replies to the requests that asked for them.
//
// A request to the browser is a broadcast: every live tab runs mcp.js and
// every one of them answers. The first reply wins and is delivered to whoever
// registered the id; later replies with the same id find nothing waiting and
// are dropped. Both the TCP relay and CallGUI route through this, so they
// cannot disagree about what a duplicate is.
type replyRouter struct {
	mu      sync.Mutex
	pending map[string]func(json.RawMessage)
}

func newReplyRouter() *replyRouter {
	return &replyRouter{pending: map[string]func(json.RawMessage){}}
}

// expect registers deliver as the destination of the reply with this id.
// It must be called before the request is sent, or a fast reply is lost.
func (r *replyRouter) expect(id json.RawMessage, deliver func(json.RawMessage)) {
	r.mu.Lock()
	r.pending[string(id)] = deliver
	r.mu.Unlock()
}

// forget drops a request that will never be answered.
func (r *replyRouter) forget(id json.RawMessage) {
	r.mu.Lock()
	delete(r.pending, string(id))
	r.mu.Unlock()
}

// route delivers a reply to its waiter, once. It reports false for a reply
// nobody is waiting for: a duplicate from a second tab, or a late reply to a
// request that already gave up.
func (r *replyRouter) route(id json.RawMessage, payload json.RawMessage) bool {
	r.mu.Lock()
	deliver := r.pending[string(id)]
	delete(r.pending, string(id))
	r.mu.Unlock()
	if deliver == nil {
		return false
	}
	deliver(payload)
	return true
}

// guiError is a string type so its values can be constants: a package-level
// error variable is shared mutable state, and errors.Is still works on a
// comparable defined type.
type guiError string

func (e guiError) Error() string { return string(e) }

// ErrNoGUI means no browser had the GUI open, so nothing will ever answer.
const ErrNoGUI guiError = "no GUI connected: open the ensemble GUI in a browser"

// ErrGUITimeout means a browser had the GUI open and did not answer in time.
const ErrGUITimeout guiError = "the GUI did not answer in time"

// selfSource is the hub routing tag for requests the agent makes of its own
// GUI, distinct from the in-process MCP client (""), the virtual user ("vu")
// and TCP clients ("tcp-N").
const selfSource = "self"

// guiCallTimeout bounds a CallGUI. A snapshot is a synchronous walk of the
// DOM, so a browser that has not answered in this long is not going to.
const guiCallTimeout = 10 * time.Second

// CallGUI sends one JSON-RPC request to the browser's MCP server and returns
// the result of the first reply. mcp.js keeps no per-client state, so no
// initialize handshake is needed first.
func (h *Server) CallGUI(ctx context.Context, method string, params any) (json.RawMessage, error) {
	id, _ := json.Marshal(h.selfSeq.Add(1))
	got := make(chan json.RawMessage, 1)
	h.selfReplies.expect(id, func(p json.RawMessage) { got <- p })

	msg, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      json.RawMessage(id),
		"method":  method,
		"params":  params,
	})
	if err != nil {
		h.selfReplies.forget(id)
		return nil, err
	}
	if h.BroadcastJSONRPC(msg, selfSource) == 0 {
		h.selfReplies.forget(id)
		return nil, ErrNoGUI
	}

	var reply json.RawMessage
	select {
	case reply = <-got:
	case <-ctx.Done():
		h.selfReplies.forget(id)
		return nil, ctx.Err()
	case <-time.After(guiCallTimeout):
		h.selfReplies.forget(id)
		return nil, ErrGUITimeout
	}

	var resp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(reply, &resp); err != nil {
		return nil, fmt.Errorf("gui reply: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("gui: %s", resp.Error.Message)
	}
	return resp.Result, nil
}

// ViewGUITool lets the agent look at its own GUI when it chooses to.
//
// It is an ordinary tool, not an ephemeral one, and that is the point. Its
// result is appended to the conversation like any other tool result, so it
// costs tokens only when the model decides to look, it works on models that
// forbid ephemeral content, and it never changes the cached prefix. Because it
// does no handshake at startup, it does not care whether a browser is open
// yet; a call with none open says so.
func (h *Server) ViewGUITool() common.Tool {
	return common.Tool{
		Name: "view_gui",
		Description: "Look at your own GUI as the user currently sees it in the browser. " +
			"Returns a markdown snapshot: the panes, every interactive element with a CSS " +
			"selector, and previews of the artifacts on screen. Use it when the user refers " +
			"to something on screen, or to check that a change to the GUI took effect.",
		Schema: json.RawMessage(`{"type":"object","properties":{}}`),
		Run: func(c *common.Call, args json.RawMessage) (string, error) {
			res, err := h.CallGUI(context.Background(), "tools/call", map[string]any{
				"name":      "gui_snapshot",
				"arguments": map[string]any{},
			})
			if err != nil {
				return "", err
			}
			var tr struct {
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
				IsError bool `json:"isError"`
			}
			if err := json.Unmarshal(res, &tr); err != nil {
				return "", fmt.Errorf("gui_snapshot result: %w", err)
			}
			var parts []string
			for _, item := range tr.Content {
				if item.Type == "text" {
					parts = append(parts, item.Text)
				}
			}
			text := strings.Join(parts, "\n")
			if tr.IsError {
				return "", fmt.Errorf("gui_snapshot: %s", text)
			}
			return text, nil
		},
	}
}

// receiveSelf is the hub receiver for selfSource: browser replies to CallGUI.
func (h *Server) receiveSelf(payload json.RawMessage) {
	var head rpcHead
	if json.Unmarshal(payload, &head) != nil || head.ID == nil || head.Method != "" {
		return
	}
	h.selfReplies.route(head.ID, payload)
}
