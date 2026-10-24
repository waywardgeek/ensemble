package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"sync"

	"ensemble/internal/common"
)

const maxHTTPMessage = 16 << 20

type httpDelivery struct {
	data json.RawMessage
	err  error
}

// httpTransport only frames messages. Correlation, tool registration and
// capability grants remain with peer, manager and Skills respectively.
// There is intentionally no Client.Timeout: delayed tools own their lifetime.
type httpTransport struct {
	ctx               context.Context
	cancel            context.CancelFunc
	client            *http.Client
	endpoint, token   string
	mu                sync.Mutex
	closed            bool
	requests          sync.WaitGroup
	session, protocol string
	incoming          chan httpDelivery
}

func newHTTPTransport(c common.MCPServer, diagnostics io.Writer) (*httpTransport, error) {
	if err := common.ValidateMCPURL(c.URL); err != nil {
		return nil, err
	}
	token := os.Getenv(c.AuthEnv)
	if strings.TrimSpace(token) == "" {
		token = ""
	}
	if c.AuthEnv != "" && token == "" && diagnostics != nil {
		fmt.Fprintf(diagnostics, "MCP credential %s is not set; attempting anonymous access\n", c.AuthEnv)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &httpTransport{ctx: ctx, cancel: cancel, endpoint: c.URL, token: token,
		client: &http.Client{Transport: http.DefaultTransport.(*http.Transport).Clone(), CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, incoming: make(chan httpDelivery, 32)}, nil
}
func (t *httpTransport) setProtocol(v string) { t.mu.Lock(); t.protocol = v; t.mu.Unlock() }
func (t *httpTransport) Close() error {
	t.mu.Lock()
	t.closed = true
	t.cancel()
	t.mu.Unlock()
	t.requests.Wait()
	t.client.CloseIdleConnections()
	return nil
}
func (t *httpTransport) Recv() (json.RawMessage, error) {
	select {
	case d := <-t.incoming:
		return d.data, d.err
	case <-t.ctx.Done():
		return nil, io.EOF
	}
}
func (t *httpTransport) deliver(ctx context.Context, d httpDelivery) {
	select {
	case t.incoming <- d:
	case <-ctx.Done():
	case <-t.ctx.Done():
	}
}
func (t *httpTransport) Send(raw json.RawMessage) error { return t.SendContext(t.ctx, raw) }
func (t *httpTransport) SendContext(ctx context.Context, raw json.RawMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return io.ErrClosedPipe
	}
	t.requests.Add(1)
	t.mu.Unlock()
	launched := false
	defer func() {
		if !launched {
			t.requests.Done()
		}
	}()
	var m message
	if json.Unmarshal(raw, &m) != nil {
		return errors.New("invalid MCP request")
	}
	requestCtx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(t.ctx, cancel)
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, t.endpoint, bytes.NewReader(raw))
	if err != nil {
		stop()
		cancel()
		return errors.New("MCP HTTP request creation failed")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	t.mu.Lock()
	if t.session != "" {
		req.Header.Set("Mcp-Session-Id", t.session)
	}
	if t.protocol != "" {
		req.Header.Set("MCP-Protocol-Version", t.protocol)
	}
	t.mu.Unlock()
	if m.Method == "" || !validID(m.ID) {
		defer cancel()
		defer stop()
		return t.exchange(requestCtx, req, m)
	}
	launched = true
	go func() {
		defer t.requests.Done()
		defer cancel()
		defer stop()
		err := t.exchange(requestCtx, req, m)
		if err != nil && requestCtx.Err() == nil {
			t.deliver(requestCtx, httpDelivery{err: err})
		}
	}()
	return nil
}
func (t *httpTransport) exchange(ctx context.Context, req *http.Request, sent message) error {
	resp, err := t.client.Do(req)
	if err != nil {
		return errors.New("MCP HTTP exchange failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read past the display boundary by an escaped token length so truncation cannot
		// disclose a prefix of a credential spanning that boundary.
		const limit = 4096
		body, _ := io.ReadAll(io.LimitReader(resp.Body, int64(limit+6*len(t.token))))
		reason := strings.TrimSpace(string(redactHTTPError(body, t.token)))
		if len(reason) > limit {
			reason = reason[:limit] + "…"
		}
		if reason == "" {
			reason = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("MCP HTTP status %d: %s", resp.StatusCode, reason)
	}
	if session := resp.Header.Get("Mcp-Session-Id"); session != "" {
		if sent.Method != "initialize" {
			return errors.New("MCP session header outside initialization")
		}
		for _, c := range session {
			if c < 0x21 || c > 0x7e {
				return errors.New("invalid MCP session header")
			}
		}
		t.mu.Lock()
		t.session = session
		t.mu.Unlock()
	}
	expectsReply := sent.Method != "" && validID(sent.ID)
	if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusNoContent {
		if expectsReply {
			return errors.New("MCP request accepted without a response")
		}
		return nil
	}
	contentType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return errors.New("MCP response has invalid content type")
	}
	matched := false
	emit := func(raw []byte) error {
		var m message
		if len(raw) > maxHTTPMessage || json.Unmarshal(raw, &m) != nil || m.JSONRPC != "2.0" {
			return errors.New("invalid MCP HTTP message")
		}
		if m.Method == "" && (!validID(m.ID) || (len(m.Result) == 0) == (m.Error == nil)) {
			return errors.New("invalid MCP HTTP response")
		}
		if expectsReply && m.Method == "" && bytes.Equal(m.ID, sent.ID) {
			matched = true
		}
		// Redact decoded error strings, not JSON spellings: escaped credentials
		// must not survive into peer errors or error-valued tool content. Leave
		// IDs, successful content, and unrelated RPC fields untouched.
		var result struct {
			IsError bool `json:"isError"`
		}
		_ = json.Unmarshal(m.Result, &result)
		if t.token != "" && (m.Error != nil || result.IsError) {
			var fields map[string]json.RawMessage
			_ = json.Unmarshal(raw, &fields)
			if m.Error != nil {
				fields["error"] = redactHTTPError(fields["error"], t.token)
			}
			if result.IsError {
				fields["result"] = redactHTTPError(fields["result"], t.token)
			}
			raw, _ = json.Marshal(fields)
		}
		t.deliver(ctx, httpDelivery{data: append(json.RawMessage(nil), raw...)})
		return nil
	}
	switch contentType {
	case "application/json":
		raw, err := io.ReadAll(io.LimitReader(resp.Body, maxHTTPMessage+1))
		if err != nil {
			return errors.New("MCP HTTP response read failed")
		}
		if len(bytes.TrimSpace(raw)) == 0 && !expectsReply {
			return nil
		}
		if err = emit(raw); err != nil {
			return err
		}
	case "text/event-stream":
		err := readMCPEvents(resp.Body, func(raw []byte) error {
			if err := emit(raw); err != nil {
				return err
			}
			if matched {
				return io.EOF
			}
			return nil
		})
		if err != nil && err != io.EOF {
			return err
		}
	default:
		return errors.New("unsupported MCP HTTP response content type")
	}
	if expectsReply && !matched {
		return errors.New("MCP HTTP stream ended without matching response")
	}
	return nil
}

// Errors may be text or structured JSON. Preserve number precision and content
// structure while replacing credential occurrences in decoded string values.
func redactHTTPError(raw []byte, token string) []byte { return common.RedactDiagnostic(raw, token) }

// SSE dispatches on a blank line, joining data fields exactly as the protocol
// specifies. Comments, event names and keepalives never reach the JSON peer.
func readMCPEvents(r io.Reader, emit func([]byte) error) error {
	scan := bufio.NewScanner(r)
	scan.Buffer(make([]byte, 4096), maxHTTPMessage)
	var data []string
	size := 0
	dispatch := func() error {
		if len(data) == 0 {
			return nil
		}
		raw := strings.Join(data, "\n")
		data = nil
		size = 0
		return emit([]byte(raw))
	}
	for scan.Scan() {
		line := scan.Text()
		if line == "" {
			if err := dispatch(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		if field == "data" {
			size += len(value) + 1
			if size > maxHTTPMessage {
				return errors.New("MCP SSE event too large")
			}
			data = append(data, value)
		}
	}
	if scan.Err() != nil {
		return errors.New("MCP SSE read failed")
	}
	if len(data) > 0 {
		return errors.New("MCP SSE ended inside an event")
	}
	return nil
}
