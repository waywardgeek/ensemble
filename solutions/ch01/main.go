// Chapter 1 — A Conversation, the Obvious Way.
//
// A complete multi-round chat client for the Anthropic Messages API in one
// file, with no SDK and no framework. Every byte of every request is put there
// by code you can read.
//
// Two modes, one conversation core:
//
//	./ch01          grader mode — JSON lines on stdin/stdout
//	./ch01 chat     interactive mode — talk to it yourself
//
// Environment:
//
//	ANTHROPIC_API_KEY    required
//	ANTHROPIC_BASE_URL   optional, defaults to https://api.anthropic.com
//	                     (the grader and the course proxy both live here)
//	ANTHROPIC_MODEL      required — there is deliberately no default, because a
//	                     remembered model ID is exactly the failure mode this
//	                     chapter is about. Ask GET /v1/models what exists today.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// --- the obvious data structure --------------------------------------------

// Message is one turn. Role is "user" or "assistant"; the API requires that
// they strictly alternate.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Conversation is the entire history. This is the whole state of the chat —
// which is the deep fact of this chapter: the API is stateless, so the
// conversation lives in this slice or it lives nowhere.
type Conversation []Message

// systemPrompt is the one fixed line of instruction that rides outside the
// messages array.
const systemPrompt = "You are a helpful assistant built from raw HTTP calls in Chapter 1 of Building Advanced AI Coding Agents. Answer briefly."

// --- the wire format --------------------------------------------------------

type request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
}

type response struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// --- the client -------------------------------------------------------------

// Client holds the connection details and the running bill.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client

	// Usage is money. Accumulated from the first request onward.
	InputTokens  int
	OutputTokens int
}

func newClient() (*Client, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}
	base := os.Getenv("ANTHROPIC_BASE_URL")
	if base == "" {
		base = "https://api.anthropic.com"
	}
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		// No default on purpose. Any ID written here would be a soft hardcode:
		// correct the day it was typed, wrong some later day, and invisible to
		// the grader, which always sets the variable. Refuse instead, and tell
		// the reader where the real answer lives:
		//   curl -s https://api.anthropic.com/v1/models \
		//     -H "x-api-key: $ANTHROPIC_API_KEY" \
		//     -H "anthropic-version: 2023-06-01"
		return nil, fmt.Errorf("ANTHROPIC_MODEL is not set; ask GET /v1/models which models exist")
	}
	return &Client{
		BaseURL: strings.TrimSuffix(base, "/"),
		APIKey:  key,
		Model:   model,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}, nil
}

// Send POSTs the entire conversation and returns the assistant's reply text.
// Note what is *not* here: no retry, no backoff, no streaming. A 429 kills
// this program, and that is the honest state of a naive client.
func (c *Client) Send(conv Conversation) (string, error) {
	body, err := json.Marshal(request{
		Model:     c.Model,
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages:  conv,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}

	var parsed response
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	c.InputTokens += parsed.Usage.InputTokens
	c.OutputTokens += parsed.Usage.OutputTokens

	var text strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	return text.String(), nil
}

// Ask is the loop of the whole chapter: append the question, send everything,
// append the answer, hand it back.
func (c *Client) Ask(conv *Conversation, question string) (string, error) {
	*conv = append(*conv, Message{Role: "user", Content: question})
	reply, err := c.Send(*conv)
	if err != nil {
		return "", err
	}
	*conv = append(*conv, Message{Role: "assistant", Content: reply})
	return reply, nil
}

// --- mode 1: the grader contract --------------------------------------------

type graderIn struct {
	User string `json:"user"`
}

type graderOut struct {
	Assistant string `json:"assistant"`
}

type usageOut struct {
	Usage struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"usage"`
}

func runGrader(c *Client) error {
	var conv Conversation
	out := json.NewEncoder(os.Stdout)

	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		var msg graderIn
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return fmt.Errorf("bad input line: %w", err)
		}
		reply, err := c.Ask(&conv, msg.User)
		if err != nil {
			return err
		}
		if err := out.Encode(graderOut{Assistant: reply}); err != nil {
			return err
		}
	}
	if err := in.Err(); err != nil {
		return err
	}

	// stdin closed: report the bill.
	var u usageOut
	u.Usage.Input = c.InputTokens
	u.Usage.Output = c.OutputTokens
	return out.Encode(u)
}

// --- mode 2: chat with it ---------------------------------------------------

func runChat(c *Client) error {
	fmt.Fprintf(os.Stderr, "talking to %s at %s — Ctrl-D to quit\n\n", c.Model, c.BaseURL)
	var conv Conversation
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for {
		fmt.Print("you> ")
		if !in.Scan() {
			break
		}
		question := strings.TrimSpace(in.Text())
		if question == "" {
			continue
		}
		reply, err := c.Ask(&conv, question)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		fmt.Printf("\nclaude> %s\n\n", reply)
		// Usage is money: watch the input count climb every round as the
		// resent history gets longer.
		fmt.Fprintf(os.Stderr, "[%d turns | %d input tokens | %d output tokens]\n\n",
			len(conv), c.InputTokens, c.OutputTokens)
	}
	fmt.Fprintf(os.Stderr, "\ntotal: %d input tokens, %d output tokens\n", c.InputTokens, c.OutputTokens)
	return nil
}

func main() {
	c, err := newClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ch01: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "chat" {
		err = runChat(c)
	} else {
		err = runGrader(c)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ch01: %v\n", err)
		os.Exit(1)
	}
}
