// Package cli adapts process I/O to the shared core; it owns no conversation logic.
package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"ensemble"
)

func Run(ctx context.Context, args []string, in io.Reader, out, diagnostics io.Writer) (result error) {
	if len(args) > 0 && args[0] == "auth" {
		return ensemble.RunAuthCommand(ctx, ensemble.NewAuth(ensemble.AuthOptionsFromEnv(os.Getenv)), args[1:], out, nil)
	}
	flags := flag.NewFlagSet("ch01", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	mcpListen := flags.String("mcp-listen", "", "serve a sibling MCP listener on loopback while the CLI runs")
	mcpPipe := flags.Bool("mcp-pipe", false, "serve MCP on stdin/stdout until EOF")
	mcpSocket := flags.String("mcp-socket", "", "serve MCP on a dedicated TCP address")
	guiDebug := flags.Bool("gui-debug", false, "load gui-debug at startup")
	skillsDir := flags.String("skills-dir", "", "skill directory (default: ./skills when ensemble is present)")
	safe := flags.Bool("safe-mode", false, "disable command and job tools")
	web := flags.Bool("no-web-search", false, "disable web-search skill")
	root := flags.String("sandbox", "", "confine Agent files and commands to this directory")
	savePath := flags.String("save", "", "recovery archive (default: save.json)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	args = flags.Args()
	protocolMode := *mcpPipe || *mcpSocket != ""
	if protocolMode && *mcpListen != "" {
		return errors.New("embedded and dedicated MCP modes are mutually exclusive")
	}
	if (*mcpPipe && *mcpSocket != "") || (protocolMode && len(args) > 0) {
		return errors.New("MCP modes are mutually exclusive and take no chat/render arguments")
	}
	chat := len(args) == 1 && args[0] == "chat"
	renderMode := len(args) == 2 && args[0] == "render"
	dumpMode := len(args) == 1 && args[0] == "dump"
	if len(args) != 0 && !chat && !renderMode && !dumpMode {
		return errors.New("usage: ch01 [--safe-mode] [--no-web-search] [--sandbox PATH] [--save PATH] [--gui-debug] [--skills-dir PATH] [chat | render LOG | dump]")
	}
	config, err := ensemble.ConfigFromEnv(os.Getenv)
	if err != nil {
		return err
	}
	if err := ConfigureSkills(&config, *skillsDir, *guiDebug); err != nil {
		return err
	}
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "save":
			config.SavePath = *savePath
		case "safe-mode":
			config.SafeMode = *safe
		case "no-web-search":
			config.EnableWebSearch = !*web
		case "sandbox":
			config.SandboxRoot = *root
		}
	})
	if protocolMode && config.Model == "" {
		config.Model = "offline-mcp"
	}
	if dumpMode && config.Model == "" {
		config.Model = "offline-dump"
	}
	path := os.Getenv("CH07_AGENT_LOG")
	if path == "" {
		path = os.Getenv("CH02_LOG")
	}
	if path == "" {
		path = "ch02.jsonl"
	}
	if !renderMode && !dumpMode {
		config.LogPath = path
	}
	config.DataDir = "cr"
	if config.SandboxRoot != "" {
		config.BaseDir = config.SandboxRoot
	}
	if config.SavePath == "" {
		config.SavePath = "save.json"
	}
	// Offline log tools must not load or overwrite the conversational save.
	if renderMode || dumpMode {
		parent := ""
		if config.SandboxRoot != "" {
			parent = config.SandboxRoot
		}
		dir, err := os.MkdirTemp(parent, "ensemble-log-tool-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
		config.SavePath = dir + "/save.json"
	}
	app := ensemble.New(log.New(diagnostics, "", log.LstdFlags))
	if err := app.SetAuth(ensemble.NewAuth(ensemble.AuthOptionsFromEnv(os.Getenv))); err != nil {
		return err
	}
	tracePath := os.Getenv("E3_TRACE")
	if tracePath == "" && !renderMode && !dumpMode {
		tracePath = "api.log"
	}
	var trace *ensemble.DiagnosticLog
	ready := false
	defer func() {
		shutdownErr := app.Shutdown()
		if result == nil && shutdownErr == nil && ready && trace != nil {
			result = trace.Retain()
		}
		if trace != nil {
			result = errors.Join(result, trace.Close())
		}
		result = errors.Join(result, shutdownErr)
	}()
	if err := installAgentFactory(app); err != nil {
		return err
	}
	a, err := app.CreateAgent(ctx, config)
	if err != nil {
		return err
	}
	if path := tracePath; path != "" {
		trace, err = ensemble.OpenAgentDiagnosticLog(a, path, func() int {
			agents := app.Agents()
			if len(agents) > 0 {
				return *agents[0].Config().DiagnosticMaxRecords
			}
			return 200
		})
		if err != nil {
			return errors.New("opening trace failed")
		}

		app.SetTrace(trace)
	}

	if *mcpListen != "" && !renderMode && !dumpMode {
		listener, err := ensemble.StartMCPListener(ctx, a, *mcpListen)
		if err != nil {
			return err
		}
		defer listener.Close()
	}
	if protocolMode {
		ready = true
		if *mcpPipe {
			return ensemble.ServeMCP(ctx, a, ensemble.NewMCPStream(ensemble.SplitReadWriteCloser{Reader: in, Writer: out, CloseFunc: func() error {
				if closer, ok := in.(io.Closer); ok {
					return closer.Close()
				}
				return nil
			}}))
		}
		return ensemble.ListenMCP(ctx, a, *mcpSocket)
	}
	if r := a.Recovery(); r.Path != "" {
		fmt.Fprintf(diagnostics, "loaded %s through sequence %d (backup=%t)\n", r.Path, r.AsOf, r.Backup)
	}
	if renderMode || dumpMode {
		if renderMode {
			path = args[1]
		}
		f, err := a.Sandbox().OpenFile(path, os.O_RDONLY, 0)
		if err != nil {
			return errors.New("opening history failed")
		}
		defer f.Close()
		if err := a.Log().Load(f); err != nil {
			return err
		}
		if dumpMode {
			return a.Log().Dump(out)
		}
		b, err := a.Render()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, string(b))
		return err
	}
	// Prune only after a usable recovery boundary exists, never on failed load
	// or failed fresh save. Quiet startup must apply policy even without requests.
	if a.Recovery().Path == "" {
		if err = a.Save(); err != nil {
			return err
		}
	}
	if trace != nil {
		if err = trace.Retain(); err != nil {
			return err
		}
	}
	ready = true
	if chat {
		return runChat(ctx, a, in, out, diagnostics)
	}
	// ReadString grows to fit a physical line instead of imposing Scanner's
	// default 64 KiB ceiling on otherwise valid questions.
	reader := bufio.NewReader(in)
	encoder := json.NewEncoder(out)
	progress := a.Observe()
	defer progress.Close()
	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return errors.New("reading input failed")
		}
		if readErr == io.EOF && len(line) == 0 {
			break
		}
		var input struct {
			User      string  `json:"user"`
			Kind      string  `json:"kind"`
			Text      string  `json:"text"`
			Ephemeral *string `json:"ephemeral"`
		}
		if err := json.Unmarshal([]byte(line), &input); err != nil {
			return errors.New("invalid input JSON")
		}
		if input.Kind == "status" || input.Kind == "model" {
			if input.Kind == "model" {
				if err := a.SetModel(input.Text); err != nil {
					if err := encoder.Encode(map[string]string{"error": err.Error()}); err != nil {
						return err
					}
					continue
				}
			}
			if err := encoder.Encode(a.Engine().Status()); err != nil {
				return err
			}
			continue
		}
		if input.Kind == "save" {
			if err := a.Save(); err != nil {
				return err
			}
			if err := encoder.Encode(map[string]string{"ack": "save"}); err != nil {
				return err
			}
			continue
		}
		if input.Kind == "prompt" {
			input.User = input.Text
		}
		if input.Kind == "hint" {
			if err := a.Post(ensemble.Hint{Text: input.Text}); err != nil {
				return err
			}
			if err := encoder.Encode(map[string]string{"ack": "hint"}); err != nil {
				return err
			}
			continue
		}
		if input.Ephemeral != nil {
			if err := a.Post(ensemble.Ephemeral{Text: *input.Ephemeral}); err != nil {
				return err
			}
			if err := encoder.Encode(struct {
				Ack string `json:"ack"`
			}{"ephemeral"}); err != nil {
				return err
			}
			continue
		}
		reply, err := protocolTurn(ctx, a, input.User, progress, encoder, input.Kind == "prompt" || a.Config().Streaming())
		if err != nil {
			return err
		}
		if err := encoder.Encode(map[string]string{"assistant": reply}); err != nil {
			return err
		}
	}
	return encoder.Encode(map[string]ensemble.Usage{"usage": a.Usage()})
}

// Deltas are optional progress; the answer still comes from Submit's private
// reply. Drain again after completion so scheduling cannot overtake queued data.
func protocolTurn(ctx context.Context, a ensemble.Agent, text string, q ensemble.ObservationQueue, encoder *json.Encoder, observations bool) (string, error) {
	drain := func() error {
		for _, o := range q.Drain() {
			if !observations {
				continue
			}
			var record any
			switch v := o.(type) {
			case ensemble.BillingSelected:
				record = struct {
					Observation string `json:"observation"`
					ensemble.BillingSelected
				}{"billing_selected", v}
			case ensemble.StateChanged:
				record = struct {
					Observation string `json:"observation"`
					ensemble.StateChanged
				}{"state_changed", v}
			case ensemble.PartDelta:
				record = struct {
					Observation string `json:"observation"`
					ensemble.PartDelta
				}{"part_delta", v}
			case ensemble.PartFinal:
				record = struct {
					Observation string `json:"observation"`
					ensemble.PartFinal
					Text string `json:"text,omitempty"`
					Tool string `json:"tool,omitempty"`
				}{"part_final", v, v.Part.Text, v.Part.Name}
			default:
				continue
			}
			if err := encoder.Encode(record); err != nil {
				return err
			}
		}
		return nil
	}
	reply := a.Submit(ctx, text)
	for {
		select {
		case <-q.Wake():
			if err := drain(); err != nil {
				return "", err
			}
		case result := <-reply:
			if err := drain(); err != nil {
				return "", err
			}
			return result.Text, result.Err
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}
