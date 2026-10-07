// Package llm owns conversation interpretation, model exchanges and accounting.
package llm

import (
	"bytes"
	"context"
	"errors"
	"example.com/ensemble/internal/common"
	"fmt"
	"io"
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
	config := e.parent.Config()
	path := "/v1/messages"
	if config.Vendor == "openai" {
		path = "/v1/chat/completions"
	}
	if config.Vendor == "gemini" {
		path = "/v1beta/models/" + url.PathEscape(strings.TrimPrefix(config.Model, "models/")) + ":generateContent"
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
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return common.ParsedResponse{}, requestFailure(e, err, "model response read failed")
	}
	return Parse(e, config, data, responseSeq)
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
