// The executable frames human chat and the exercise's JSON-lines protocol.
package main

import (
	"errors"
	"fmt"
	"os"

	"ensemble/ensemble"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (runErr error) {
	// Command selection is separate from configuration. In particular, render
	// cannot accept model flags that would make byte-identity depend on a
	// second source of request settings.
	mode := "protocol"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	if !((mode == "protocol" && len(os.Args) == 1) || ((mode == "chat" || mode == "dump") && len(os.Args) == 2) || (mode == "render" && len(os.Args) == 3)) {
		return errors.New("usage: ensemble [chat | dump | render LOG]")
	}
	cfg, err := ensemble.ConfigFromEnv()
	cfg.BuiltinTools, cfg.DataDir = true, "cr"
	if err != nil {
		return err
	}
	// Dump only decodes the log; it does not require model setup to inspect facts.
	if mode == "dump" && cfg.Model == "" {
		cfg.Model = "offline"
	}
	logPath := os.Getenv("CH02_LOG")
	if logPath == "" {
		logPath = "ch02.jsonl"
	}
	cfg.LogPath = logPath
	app := ensemble.New(os.Stderr)
	a, err := app.NewAgent(cfg)
	if err != nil {
		return err
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
	return converse(a, mode == "chat")
}
