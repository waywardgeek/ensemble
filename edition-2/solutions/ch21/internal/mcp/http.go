package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"ensemble/internal/common"
)

// HTTP is a boundary resource owned by the MCP manager, like its stdio peers.
// One POST may produce no messages or many. A queue decouples senders from the
// protocol reader; a bounded channel here could deadlock discovery or replies.
type httpTransport struct {
	parent                   common.MCP
	endpoint, token, session string
	client                   *http.Client
	ctx                      context.Context
	cancel                   context.CancelFunc
	mu                       sync.Mutex
	ready                    *sync.Cond
	queue                    []json.RawMessage
	closed                   bool
}

// NewHTTP creates a boundary owned by the supplied manager. Credentials resolve
// once at connection creation, never from model-visible skill instructions.
func NewHTTP(parent common.MCP, server common.MCPServer) *httpTransport {
	if parent == nil {
		panic("HTTP MCP requires its manager")
	}
	ctx, cancel := context.WithCancel(context.Background())
	t := &httpTransport{parent: parent, endpoint: server.URL, ctx: ctx, cancel: cancel,
		// Redirects are not part of the declared endpoint contract. In particular,
		// never forward its credential to a location supplied by another server.
		client: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	t.ready = sync.NewCond(&t.mu)
	if server.AuthEnv != "" {
		t.token = os.Getenv(server.AuthEnv)
		if t.token == "" {
			parent.Agent().Ensemble().Logf("MCP %s: auth environment %s unset; connecting without authentication", server.Name, server.AuthEnv)
		}
	}
	return t
}

// MCP exposes the connection owner rather than copying its logging services.
func (t *httpTransport) MCP() common.MCP { return t.parent }

// Send supports the common message boundary for notifications and reverse replies.
func (t *httpTransport) Send(message json.RawMessage) error {
	return t.SendContext(context.Background(), message)
}

// SendContext binds a POST to both the actual job and the connection lifetime.
// Callback expiry does not cancel the job; Close cancels every active POST.
func (t *httpTransport) SendContext(ctx context.Context, message json.RawMessage) error {
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(t.ctx, cancel)
	defer stop()
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint, bytes.NewReader(message))
	if err != nil {
		return t.diagnostic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	if t.token != "" {
		request.Header.Set("Authorization", "Bearer "+t.token)
	}
	t.mu.Lock()
	session, closed := t.session, t.closed
	t.mu.Unlock()
	if closed {
		return errClosed
	}
	if session != "" {
		request.Header.Set("Mcp-Session-Id", session)
	}
	response, err := t.client.Do(request)
	if err != nil {
		return t.diagnostic(err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return t.diagnostic(fmt.Errorf("MCP HTTP %s: %s (reading body: %w)", response.Status, body, readErr))
		}
		return t.diagnostic(fmt.Errorf("MCP HTTP %s: %s", response.Status, body))
	}
	if session := response.Header.Get("Mcp-Session-Id"); session != "" {
		t.mu.Lock()
		t.session = session
		t.mu.Unlock()
	}
	if response.StatusCode == http.StatusAccepted {
		return nil
	}
	if strings.HasPrefix(response.Header.Get("Content-Type"), "text/event-stream") {
		return t.events(response.Body)
	}
	var reply json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&reply); err != nil {
		return t.diagnostic(err)
	}
	return t.enqueue(reply)
}

// SSE data lines form one JSON message at a blank line; event names and comments
// are framing, not JSON. ReadString avoids a scanner's accidental 64 KiB cap on
// scraped pages. Enqueue each frame immediately so reverse requests can run.
func (t *httpTransport) events(body io.Reader) error {
	reader := bufio.NewReader(body)
	var data []string
	flush := func() error {
		if len(data) == 0 {
			return nil
		}
		raw := json.RawMessage(strings.Join(data, "\n"))
		data = nil
		if !json.Valid(raw) {
			return fmt.Errorf("invalid MCP SSE JSON")
		}
		return t.enqueue(raw)
	}
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if line == "" {
			if e := flush(); e != nil {
				return e
			}
		} else if value, ok := strings.CutPrefix(line, "data:"); ok {
			data = append(data, strings.TrimPrefix(value, " "))
		}
		if err == io.EOF {
			return flush()
		}
		if err != nil {
			return t.diagnostic(err)
		}
	}
}

func (t *httpTransport) enqueue(message json.RawMessage) error {
	// A peer can reflect its Authorization value in a JSON-RPC error or even
	// a successful result. Keep that transport secret out of model context too.
	if t.token != "" {
		message = bytes.ReplaceAll(message, []byte(t.token), []byte("[REDACTED]"))
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return errClosed
	}
	t.queue = append(t.queue, message)
	t.ready.Signal()
	return nil
}

// Recv waits without polling. Close wakes it even when no POST ever replied.
func (t *httpTransport) Recv() (json.RawMessage, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for len(t.queue) == 0 && !t.closed {
		t.ready.Wait()
	}
	if t.closed {
		return nil, errClosed
	}
	message := t.queue[0]
	t.queue[0] = nil
	t.queue = t.queue[1:]
	return message, nil
}

// Close releases pending I/O before the Agent joins its jobs during shutdown.
func (t *httpTransport) Close() error {
	t.mu.Lock()
	t.closed = true
	t.queue = nil
	t.ready.Broadcast()
	t.mu.Unlock()
	t.cancel()
	t.client.CloseIdleConnections()
	return nil
}

// Preserve server diagnostics, but never return our Bearer token if a peer or
// proxy reflects it in an error. Other diagnostic bytes remain intact.
func (t *httpTransport) diagnostic(err error) error {
	if t.token == "" {
		return err
	}
	return fmt.Errorf("%s", strings.ReplaceAll(err.Error(), t.token, "[REDACTED]"))
}
