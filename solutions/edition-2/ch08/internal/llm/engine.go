// Package llm owns conversation interpretation, model exchanges and accounting.
package llm

import (
	"bytes"
	"context"
	"errors"
	"example.com/ensemble/internal/common"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Engine struct {
	parent common.Agent
	client *http.Client
	// Accounting locks never call back into Agent. Agent may hold its state lock
	// while committing usage exactly once after durable append.
	usageMu sync.Mutex
	usage   map[common.Provenance]common.Usage
}

func New(parent common.Agent) *Engine {
	return &Engine{parent: parent, usage: map[common.Provenance]common.Usage{}, client: &http.Client{
		Timeout: 60 * time.Second,
		// A nil Transport would silently share http.DefaultTransport across Agents.
		Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, IdleConnTimeout: 90 * time.Second, TLSHandshakeTimeout: 10 * time.Second},
		// Redirects could leak credentials and would violate one request per question.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}
func (e *Engine) Agent() common.Agent { return e.parent }
func (e *Engine) Account(response common.Response) {
	e.usageMu.Lock()
	defer e.usageMu.Unlock()
	u := e.usage[response.From]
	e.add(&u, *response.Usage)
	e.usage[response.From] = u
}
func (e *Engine) Usage() common.Usage {
	e.usageMu.Lock()
	defer e.usageMu.Unlock()
	var out common.Usage
	for _, u := range e.usage {
		e.add(&out, u)
	}
	return out
}
func (e *Engine) UsageByModel() map[common.Provenance]common.Usage {
	e.usageMu.Lock()
	defer e.usageMu.Unlock()
	out := map[common.Provenance]common.Usage{}
	for p, u := range e.usage {
		out[p] = u
	}
	return out
}
func (e *Engine) add(total *common.Usage, u common.Usage) {
	total.Input += u.Input
	total.CacheWrite += u.CacheWrite
	total.CacheRead += u.CacheRead
	total.Output += u.Output
}

func (e *Engine) Exchange(ctx context.Context, body []byte, responseSeq uint64) (common.ParsedResponse, error) {
	return e.ExchangeConfig(ctx, body, e.parent.Config())
}
func (e *Engine) ExchangeConfig(ctx context.Context, body []byte, config common.Config) (common.ParsedResponse, error) {
	return e.ExchangeOperation(ctx, nil, body, config)
}
func (e *Engine) ExchangeOperation(ctx context.Context, op common.ModelOperation, body []byte, config common.Config) (common.ParsedResponse, error) {
	path := "/v1/messages"
	if config.Vendor == "openai" {
		path = "/v1/chat/completions"
	}
	if config.Vendor == "gemini" {
		path = "/v1beta/models/" + url.PathEscape(strings.TrimPrefix(config.Model, "models/")) + ":generateContent"
		if !config.DisableStreaming {
			path = strings.TrimSuffix(path, ":generateContent") + ":streamGenerateContent?alt=sse"
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return common.ParsedResponse{}, failure(e, "cannot construct model request")
	}
	req.Header.Set("content-type", "application/json")
	switch config.Vendor {
	case "anthropic":
		req.Header.Set("x-api-key", config.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	case "openai":
		req.Header.Set("authorization", "Bearer "+config.APIKey)
	case "gemini":
		req.Header.Set("x-goog-api-key", config.APIKey)
	}
	response, err := e.client.Do(req)
	if err != nil {
		return common.ParsedResponse{}, requestFailure(e, err, "model transport failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return common.ParsedResponse{}, failure(e, "model returned HTTP %d", response.StatusCode)
	}
	if !config.DisableStreaming {
		media, _, mediaErr := mime.ParseMediaType(response.Header.Get("Content-Type"))
		if mediaErr != nil || media != "text/event-stream" {
			return common.ParsedResponse{}, failure(e, "stream response requires text/event-stream")
		}
		if op == nil {
			return common.ParsedResponse{}, failure(e, "stream requires owned model operation")
		}
		return parseStream(ctx, op, config, response.Body)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, responseLimit+1))
	if len(data) > responseLimit {
		return common.ParsedResponse{}, failure(e, "model response exceeds 16 MiB")
	}
	if err != nil {
		return common.ParsedResponse{}, requestFailure(e, err, "model response read failed")
	}
	return Parse(e, config, data, 0)
}
func requestFailure(owner common.Engine, cause error, fallback string) error {
	// Wrap only known safe sentinels, never raw transport errors: those can
	// contain a credential-bearing URL or arbitrary provider-controlled text.
	if errors.Is(cause, context.Canceled) {
		return failure(owner, "model request canceled: %w", context.Canceled)
	}
	if errors.Is(cause, context.DeadlineExceeded) {
		return failure(owner, "model request timed out: %w", context.DeadlineExceeded)
	}
	var networkError net.Error
	if errors.As(cause, &networkError) && networkError.Timeout() {
		return failure(owner, "model request timed out")
	}
	return failure(owner, "%s", fallback)
}
func failure(owner common.Engine, format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	owner.Agent().Ensemble().Logf("%s", err)
	return err
}

func (e *Engine) Close() { e.client.CloseIdleConnections() }
