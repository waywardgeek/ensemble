// The executable frames human chat and the exercise's JSON-lines protocol.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	args := os.Args[1:]
	savePath := "save.json"
	// Accept the location flag beside any live command, without changing the
	// existing positional protocol or making load an opt-in operation.
	var positional []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--save" {
			i++
			if i == len(args) || args[i] == "" {
				return errors.New("--save requires a path")
			}
			savePath = args[i]
		} else if strings.HasPrefix(args[i], "--save=") {
			savePath = strings.TrimPrefix(args[i], "--save=")
			if savePath == "" {
				return errors.New("--save requires a path")
			}
		} else {
			positional = append(positional, args[i])
		}
	}
	mode := "protocol"
	if len(positional) > 0 {
		mode = positional[0]
	}
	if !((mode == "protocol" && len(positional) == 0) || ((mode == "chat" || mode == "dump" || mode == "--mcp-pipe") && len(positional) == 1) || ((mode == "render" || mode == "prompt") && len(positional) == 2)) {
		return errors.New("usage: ensemble [--save PATH] [chat | dump | render LOG | prompt TEXT | --mcp-pipe]")
	}
	cfg, err := ensemble.ConfigFromEnv()
	cfg.BuiltinTools, cfg.DataDir = true, "cr"
	if err != nil {
		return err
	}
	// Dump only decodes the log; it does not require model setup to inspect facts.
	if (mode == "dump" || mode == "--mcp-pipe") && cfg.Model == "" {
		cfg.Model = "offline"
	}
	logPath := os.Getenv("CH07_AGENT_LOG")
	if logPath == "" {
		logPath = os.Getenv("CH02_LOG")
	}
	if logPath == "" {
		logPath = "ch02.jsonl"
	}
	cfg.LogPath = logPath
	// This single-Agent application chooses cwd/save.json. Library consumers
	// default to their Agent's DataDir; the path remains owned by its configuration.
	cfg.SavePath, err = filepath.Abs(savePath)
	if err != nil {
		return err
	}
	app := ensemble.New(os.Stderr)
	if mode == "chat" || mode == "protocol" || mode == "prompt" {
		trace, err := os.Create("api.log")
		if err != nil {
			return err
		}
		defer trace.Close()
		app.SetAPILog(trace)
	}
	a, err := app.NewAgent(cfg, ensemble.AgentSpecFromEnv())
	if err != nil {
		return err
	}
	// These branches terminate before the live key check and input loop.
	// Dump reproduces recorded facts; render projects them onto today's target.
	// Neither command sends HTTP or consults a provider-owned conversation.
	if mode == "render" || mode == "dump" {
		path := logPath
		if mode == "render" {
			path = positional[1]
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
	if cfg.APIKey == "" && mode != "--mcp-pipe" {
		return errors.New("provider API key is not set")
	}
	if err := a.Load(); err != nil {
		return err
	}
	// Register saving only after a successful load. Joining first preserves final
	// job facts and prevents a bad input file from ever reaching replacement.
	defer func() { runErr = errors.Join(runErr, a.Shutdown(), a.Save()) }()
	if mode == "--mcp-pipe" {
		transport := ensemble.NewMCPTransport(os.Stdin, os.Stdout, pipeBoundary{})
		if err := a.MCP().Connect(context.Background(), "pipe", transport); err != nil {
			return err
		}
		<-a.MCP().Done("pipe")
		return nil
	}
	if mode == "prompt" {
		answer, err := a.Ask(context.Background(), positional[1])
		if err == nil {
			fmt.Fprintln(os.Stdout, answer)
		}
		return err
	}
	return converse(a, mode == "chat")
}

// The dedicated protocol owns both streams; closing only stdin could leave a
// blocked reverse reply holding shutdown open when its peer has disconnected.
type pipeBoundary struct{}

// Close interrupts both directions of the dedicated command's protocol pipe.
func (pipeBoundary) Close() error { return errors.Join(os.Stdin.Close(), os.Stdout.Close()) }
