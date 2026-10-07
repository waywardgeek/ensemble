// Package llm owns model exchange behavior, HTTP transport, and accounting.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"example.com/ensemble/internal/common"
)

type Engine struct {
	parent common.Agent
	client *http.Client
	usage  common.Usage
}

func New(parent common.Agent) *Engine {
	return &Engine{parent: parent, client: &http.Client{
		Timeout: 60 * time.Second,
		// A nil Transport would silently share http.DefaultTransport across Agents.
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		// Redirects could leak credentials and would violate one request per question.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

func (e *Engine) Agent() common.Agent { return e.parent }
func (e *Engine) Usage() common.Usage { return e.usage }

func (e *Engine) Ask(ctx context.Context, question string) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "", failure(e, "question must be nonempty")
	}
	body, err := buildRequest(e, e.parent.History(), question)
	if err != nil {
		return "", err
	}
	config := e.parent.Config()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", failure(e, "cannot construct model request")
	}
	req.Header.Set("x-api-key", config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")
	response, err := e.client.Do(req)
	if err != nil {
		return "", requestFailure(e, err, "model transport failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", failure(e, "model returned HTTP %d", response.StatusCode)
	}
	answer, usage, err := parseResponse(e, response.Body)
	if err != nil {
		return "", err
	}
	// Commit only a validated pair and its accounting: a rejected response must
	// not leave an unanswered user turn or charge usage to the next exchange.
	e.parent.Commit(common.Message{Role: "user", Content: question}, common.Message{Role: "assistant", Content: answer})
	e.usage.Input += usage.Input
	e.usage.Output += usage.Output
	return answer, nil
}

// Request behavior stays here even though Conversation is declared in common.
func buildRequest(owner common.Engine, history common.Conversation, question string) ([]byte, error) {
	messages := append(append(common.Conversation(nil), history...), common.Message{Role: "user", Content: question})
	request := struct {
		Model     string              `json:"model"`
		MaxTokens int                 `json:"max_tokens"`
		System    string              `json:"system"`
		Messages  common.Conversation `json:"messages"`
	}{
		Model:     owner.Agent().Config().Model,
		MaxTokens: 512,
		System:    "Answer helpfully and concisely. Retain the conversation's details.",
		Messages:  messages,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, failure(owner, "cannot encode model request")
	}
	return body, nil
}

func parseResponse(owner common.Engine, reader io.Reader) (string, common.Usage, error) {
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage *struct {
			Input  *int64 `json:"input_tokens"`
			Output *int64 `json:"output_tokens"`
		} `json:"usage"`
	}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&response); err != nil {
		return "", common.Usage{}, requestFailure(owner, err, "malformed model response JSON")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return "", common.Usage{}, requestFailure(owner, err, "trailing data in model response")
	}
	var answer strings.Builder
	for _, block := range response.Content {
		if block.Type == "text" {
			answer.WriteString(block.Text)
		}
	}
	if answer.Len() == 0 {
		return "", common.Usage{}, failure(owner, "model response has no text answer")
	}
	if response.Usage == nil || response.Usage.Input == nil || response.Usage.Output == nil || *response.Usage.Input < 0 || *response.Usage.Output < 0 {
		return "", common.Usage{}, failure(owner, "model response has missing or invalid usage")
	}
	return answer.String(), common.Usage{Input: *response.Usage.Input, Output: *response.Usage.Output}, nil
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
