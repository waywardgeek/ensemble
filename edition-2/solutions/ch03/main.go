// The executable frames human chat and the exercise's JSON-lines protocol.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"ensemble/ensemble"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Generic settings select any provider; provider-prefixed settings retain the
// Chapter 1 interface and make interactive setup convenient for all three.
func config() (ensemble.Config, error) {
	vendor := os.Getenv("LLM_VENDOR")
	if vendor == "" {
		vendor = "anthropic"
	}
	cfg := ensemble.Config{BuiltinTools: true}
	switch vendor {
	case "anthropic":
		cfg.Vendor = ensemble.Anthropic
	case "openai":
		cfg.Vendor = ensemble.OpenAI
	case "gemini":
		cfg.Vendor = ensemble.Gemini
	default:
		return cfg, errors.New("unknown LLM_VENDOR")
	}
	env := func(name string) string {
		if value := os.Getenv("LLM_" + name); value != "" {
			return value
		}
		return os.Getenv(strings.ToUpper(vendor) + "_" + name)
	}
	cfg.BaseURL, cfg.APIKey, cfg.Model = env("BASE_URL"), env("API_KEY"), env("MODEL")
	cfg.ResolvedModel = os.Getenv("LLM_RESOLVED_MODEL")
	return cfg, nil
}
func run() error {
	// Command selection is separate from configuration. In particular, render
	// cannot accept model flags that would make byte-identity depend on a
	// second source of request settings.
	mode := "protocol"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	if !((mode == "protocol" && len(os.Args) == 1) || ((mode == "chat" || mode == "dump") && len(os.Args) == 2) || (mode == "render" && len(os.Args) == 3)) {
		return errors.New("usage: ch03 [chat | dump | render LOG]")
	}
	cfg, err := config()
	if err != nil {
		return err
	}
	// Dump only decodes the log; it does not require model setup to inspect facts.
	if mode == "dump" && cfg.Model == "" {
		cfg.Model = "offline"
	}
	app := ensemble.New(os.Stderr)
	a, err := app.NewAgent(cfg)
	if err != nil {
		return err
	}
	logPath := os.Getenv("CH02_LOG")
	if logPath == "" {
		logPath = "ch02.jsonl"
	}
	// These branches terminate before the live key check and input loop.
	// Dump reproduces recorded facts; render projects them onto today's target.
	// Neither command sends HTTP or consults a provider-owned conversation.
	if mode == "render" || mode == "dump" {
		path := logPath
		if mode == "render" {
			path = os.Args[2]
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		if err := a.History().Load(file); err != nil {
			return err
		}
		if mode == "dump" {
			return a.History().Dump(os.Stdout)
		}
		body, err := a.Engine().Render(a.History().Context())
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(os.Stdout, string(body))
		return err
	}
	if cfg.APIKey == "" {
		return errors.New("provider API key is not set")
	}
	// A session starts fresh; the standalone dump command loads its saved facts.
	// Rewrite after each completed synchronous input, including failed exchanges.
	// Crash-safe journaling and automatic resume belong to later chapters.
	save := func() error {
		file, err := os.Create(logPath)
		if err != nil {
			return err
		}
		err = a.History().Dump(file)
		closeErr := file.Close()
		if err != nil {
			return err
		}
		return closeErr
	}
	input := bufio.NewReader(os.Stdin)
	output := json.NewEncoder(os.Stdout)
	chat := mode == "chat"
	turns := 0
	if chat {
		fmt.Fprintf(os.Stderr, "talking to %s — Ctrl-D to quit\n", cfg.Model)
	}
	for {
		if chat {
			fmt.Fprint(os.Stdout, "you> ")
		}
		line, readErr := input.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return readErr
		}
		if len(line) == 0 && readErr == io.EOF {
			break
		}
		question := strings.TrimRight(line, "\r\n")
		if chat && strings.TrimSpace(question) == "" {
			continue
		}
		// Presence matters for directives: an empty ephemeral string is still a
		// directive to acknowledge, not an empty user question to send upstream.
		if !chat {
			var in struct {
				User      string  `json:"user"`
				Ephemeral *string `json:"ephemeral"`
			}
			if err := json.Unmarshal([]byte(line), &in); err != nil {
				return errors.New("invalid input JSON")
			}
			if in.Ephemeral != nil {
				if err := a.Ephemeral(*in.Ephemeral); err != nil {
					return err
				}
				if err := save(); err != nil {
					return err
				}
				if err := output.Encode(struct {
					Ack string `json:"ack"`
				}{"ephemeral"}); err != nil {
					return err
				}
				continue
			}
			question = in.User
		}
		// Save failure facts as well as successful replies. A client-visible error
		// must not erase the request from the independently inspectable record.
		answer, askErr := a.Ask(context.Background(), question)
		if err := save(); err != nil {
			return err
		}
		if askErr != nil {
			return askErr
		}
		turns += 2
		if chat {
			name := "assistant"
			if cfg.Vendor == ensemble.Anthropic {
				name = "claude"
			}
			fmt.Fprintf(os.Stdout, "\n%s> %s\n\n", name, answer)
			u := a.Engine().Usage()
			fmt.Fprintf(os.Stderr, "[%d turns | %d input tokens | %d output tokens | %d cache write | %d cache read]\n", turns, u.Input, u.Output, u.CacheWrite, u.CacheRead)
		} else if err := output.Encode(struct {
			Assistant string `json:"assistant"`
		}{answer}); err != nil {
			return err
		}
	}
	if err := save(); err != nil {
		return err
	}
	if !chat {
		return output.Encode(struct {
			Usage ensemble.Usage `json:"usage"`
		}{a.Engine().Usage()})
	}
	return nil
}
