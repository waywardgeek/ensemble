package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTPTransport speaks MCP over Streamable HTTP: every client message is a
// POST, and the reply arrives either as a JSON body or as a text/event-stream
// carrying one or more "data:" frames.
//
// The other three transports are duplex streams, so Recv blocks on a pipe that
// the peer writes to whenever it likes. HTTP has no such channel: a reply only
// exists as the body of a request the client made. Send therefore parses each
// reply and queues the JSON-RPC messages it found, and Recv drains that queue.
// To the Codec's read loop this is indistinguishable from a stream.
//
// The queue is a slice guarded by a condition variable rather than a buffered
// channel. A notification POST returns 202 with no body and so enqueues
// nothing, while a single POST may yield several frames; with a fixed capacity
// either case can wedge a concurrent Send. An unbounded queue cannot.
type HTTPTransport struct {
	url     string
	client  *http.Client
	headers map[string]string

	mu        sync.Mutex
	cond      *sync.Cond
	queue     []json.RawMessage
	closed    bool
	parked    int // Recv calls currently blocked in cond.Wait; tests observe it
	sessionID string
}

// NewHTTPTransport dials nothing — HTTP is connectionless, so the first POST
// is the first contact. headers may carry authorization for servers that want
// it; a keyless server needs none.
func NewHTTPTransport(url string, headers map[string]string) (*HTTPTransport, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("mcp http transport: url %q must start with http:// or https://", url)
	}
	t := &HTTPTransport{
		url:     url,
		client:  &http.Client{Timeout: 120 * time.Second},
		headers: headers,
	}
	t.cond = sync.NewCond(&t.mu)
	return t, nil
}

// Send POSTs one JSON-RPC message and queues whatever came back.
func (t *HTTPTransport) Send(msg json.RawMessage) error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return fmt.Errorf("mcp http transport: closed")
	}
	session := t.sessionID
	t.mu.Unlock()

	req, err := http.NewRequest(http.MethodPost, t.url, bytes.NewReader(msg))
	if err != nil {
		return fmt.Errorf("mcp http transport: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Accept both shapes: the spec lets the server answer either way, and
	// servers have been observed to switch between them per method.
	req.Header.Set("Accept", "application/json, text/event-stream")
	if session != "" {
		req.Header.Set("Mcp-Session-Id", session)
	}
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("mcp http transport: post: %w", err)
	}
	defer resp.Body.Close()

	// A stateful server hands out a session on initialize and expects it back
	// on every later request. A stateless one never sends the header.
	if id := resp.Header.Get("Mcp-Session-Id"); id != "" {
		t.mu.Lock()
		t.sessionID = id
		t.mu.Unlock()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("mcp http transport: read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Surface the status AND the body. An MCP server reports auth and
		// quota problems in the body, and dropping it leaves the caller
		// guessing why a tool stopped working.
		return fmt.Errorf("mcp http transport: %s: %s", resp.Status, snippet(body))
	}

	for _, m := range parseHTTPBody(resp.Header.Get("Content-Type"), body) {
		t.mu.Lock()
		t.queue = append(t.queue, m)
		t.mu.Unlock()
		t.cond.Signal()
	}
	return nil
}

// Recv returns the next queued message, blocking until one arrives or the
// transport closes.
func (t *HTTPTransport) Recv() (json.RawMessage, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for len(t.queue) == 0 && !t.closed {
		t.parked++
		t.cond.Wait()
		t.parked--
	}
	if len(t.queue) == 0 {
		return nil, io.EOF
	}
	msg := t.queue[0]
	t.queue = t.queue[1:]
	return msg, nil
}

// Close wakes every blocked Recv. There is no connection to tear down.
func (t *HTTPTransport) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	t.mu.Unlock()
	t.cond.Broadcast()
	return nil
}

// parseHTTPBody extracts JSON-RPC messages from a reply body. An SSE body
// carries them in "data:" lines, one message per line; a JSON body is a single
// message. An empty body (202 for a notification) yields none.
func parseHTTPBody(contentType string, body []byte) []json.RawMessage {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil
	}
	if strings.Contains(contentType, "text/event-stream") {
		var out []json.RawMessage
		for _, line := range strings.Split(string(trimmed), "\n") {
			line = strings.TrimSpace(line)
			data, ok := strings.CutPrefix(line, "data:")
			if !ok {
				continue // "event:", "id:", ":" comments, blank separators
			}
			data = strings.TrimSpace(data)
			if data == "" || data == "[DONE]" {
				continue
			}
			out = append(out, json.RawMessage(data))
		}
		return out
	}
	return []json.RawMessage{json.RawMessage(trimmed)}
}

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 300 {
		return s[:300] + "..."
	}
	if s == "" {
		return "(empty body)"
	}
	return s
}
