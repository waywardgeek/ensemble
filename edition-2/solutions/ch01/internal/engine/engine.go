// Package engine owns HTTP exchanges and measured usage, not conversation history.
package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ensemble/internal/common"
)

// The model sees only this instruction and the messages we explicitly send.
// Application state and provider credentials never enter the conversation.
const systemPrompt = "You are a helpful assistant built from raw HTTP calls in Chapter 1 of The Art of Building AI Coding Agents. Answer briefly."

// Transport and accounting share the Engine lifetime. Configuration stays with
// the parent, so each request reads the settings for this particular Agent.
type engine struct {
	parent common.Agent
	client *http.Client
	usage  common.Usage
}

func New(parent common.Agent) common.Engine {
	if parent == nil {
		panic("Engine requires Agent")
	}
	// One bounded exchange; this chapter has no retries or recovery policy.
	return &engine{parent: parent, client: &http.Client{Timeout: 60 * time.Second}}
}
func (e *engine) Agent() common.Agent { return e.parent }
func (e *engine) Usage() common.Usage { return e.usage }

func (e *engine) Send(ctx context.Context, messages []common.Message) (answer string, err error) {
	// Diagnostics follow ownership to the application's logger. No request bodies,
	// credentials or provider error bodies are logged, even on unsuccessful calls.
	defer func() {
		if err != nil {
			e.parent.Ensemble().Logger().Printf("Anthropic request failed: %v", err)
		}
	}()
	// A vendor payload is temporary; serializing it does not alter Agent history.
	cfg := e.parent.Config()
	body, err := json.Marshal(struct {
		Model string `json:"model"`
		// Anthropic requires an explicit output budget, even for a short reply.
		MaxTokens int              `json:"max_tokens"`
		System    string           `json:"system"`
		Messages  []common.Message `json:"messages"`
	}{cfg.Model, 1024, systemPrompt, messages})
	if err != nil {
		return "", errors.New("encoding request failed")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cfg.BaseURL, "/")+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", errors.New("invalid Anthropic endpoint")
	}
	// The three headers are part of the wire contract, including for a proxy.
	req.Header.Set("x-api-key", cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		// URL errors can contain credentials. Preserve cancellation/timeout identity
		// without returning the URL or remote diagnostic text to clients.
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return "", context.DeadlineExceeded
		}
		return "", errors.New("Anthropic transport failed")
	}
	// Release every response body, including rejected statuses, for connection reuse.
	defer resp.Body.Close()
	// A status is useful diagnostic information without copying an untrusted
	// error body that could echo authorization data. Do not retry this exchange.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic HTTP status %d", resp.StatusCode)
	}

	// The response uses typed blocks even though this chapter sends bare strings.
	// Pointer counts distinguish a real measured zero from missing accounting.
	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			Input  *int `json:"input_tokens"`
			Output *int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", errors.New("decoding Anthropic response failed")
	}
	if parsed.Usage.Input == nil || parsed.Usage.Output == nil {
		return "", errors.New("Anthropic response missing usage")
	}
	// Account for a measured response even if it supplies no usable dialogue.
	// Cumulative input grows as the same retained history is billed again.
	e.usage.Input += *parsed.Usage.Input
	e.usage.Output += *parsed.Usage.Output
	// A reply can contain multiple text blocks; taking block zero loses words.
	var text strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	// Empty answers cannot be replayed as valid messages on the next turn.
	if strings.TrimSpace(text.String()) == "" {
		return "", errors.New("Anthropic response has no text")
	}
	return text.String(), nil
}
