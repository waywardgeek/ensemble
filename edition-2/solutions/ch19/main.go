// The executable frames human chat and the exercise's JSON-lines protocol.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ensemble/authcli"
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
	if len(args) > 0 && args[0] == "auth" {
		app := ensemble.New(os.Stderr)
		if err := authcli.Configure(app); err != nil {
			return err
		}
		return authcli.Run(context.Background(), app, args[1:], os.Stdout)
	}
	savePath := "save.json"
	spec := ensemble.AgentSpecFromEnv()
	guiDebug := false
	// Accept the location flag beside any live command, without changing the
	// existing positional protocol or making load an opt-in operation.
	var positional []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--safe-mode" {
			spec.SafeMode = true
		} else if args[i] == "--no-web-search" {
			spec.EnableWebSearch = false
		} else if args[i] == "--sandbox" {
			i++
			if i == len(args) || args[i] == "" {
				return errors.New("--sandbox requires a path")
			}
			spec.SandboxRoot = args[i]
		} else if args[i] == "--gui-debug" {
			guiDebug = true
		} else if args[i] == "--save" || args[i] == "--skills-dir" {
			name := args[i]
			i++
			if i == len(args) || args[i] == "" {
				return fmt.Errorf("%s requires a path", name)
			}
			if name == "--save" {
				savePath = args[i]
			} else {
				spec.SkillsDir = args[i]
			}
		} else if strings.HasPrefix(args[i], "--save=") || strings.HasPrefix(args[i], "--skills-dir=") {
			name, value, _ := strings.Cut(args[i], "=")
			if value == "" {
				return fmt.Errorf("%s requires a path", name)
			}
			if name == "--save" {
				savePath = value
			} else {
				spec.SkillsDir = value
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
		return errors.New("usage: ensemble [--save PATH] [--gui-debug] [--skills-dir PATH] [chat | dump | render LOG | prompt TEXT | --mcp-pipe]")
	}
	if guiDebug && spec.SkillsDir == "" {
		spec.SkillsDir = "skills"
	}
	if spec.SkillsDir != "" && spec.PrimarySkill == "" {
		spec.PrimarySkill = "ensemble"
	}
	if guiDebug {
		spec.StartupSkills = []string{"gui-debug"}
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
	if spec.SandboxRoot != "" {
		root, err := filepath.Abs(spec.SandboxRoot)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(root, 0700); err != nil {
			return err
		}
		spec.SandboxRoot = root
		cfg.DataDir = filepath.Join(root, "cr")
		logPath = filepath.Join(root, filepath.Base(logPath))
		savePath = filepath.Join(root, filepath.Base(savePath))
	}
	cfg.LogPath = logPath
	// This single-Agent application chooses cwd/save.json. Library consumers
	// default to their Agent's DataDir; the path remains owned by its configuration.
	cfg.SavePath, err = filepath.Abs(savePath)
	if err != nil {
		return err
	}
	app := ensemble.New(os.Stderr)
	if err := authcli.Configure(app); err != nil {
		return err
	}
	a, err := app.NewAgent(cfg, spec)
	if err != nil {
		return err
	}
	if mode == "chat" || mode == "protocol" || mode == "prompt" {
		tracePath := "api.log"
		trace, err := a.Sandbox().OpenFile(tracePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			a.Shutdown()
			return err
		}
		defer trace.Close()
		app.SetAPILog(trace)
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
	if a.Config().BillingMode != ensemble.CredentialChatGPT && cfg.APIKey == "" && mode != "--mcp-pipe" {
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
