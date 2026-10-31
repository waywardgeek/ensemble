// This optional application owns one Ensemble fleet and its terminal/browser views.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	gui "ensemble-gui-stub"
	"ensemble/authcli"
	"ensemble/ensemble"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() (runErr error) {
	if len(os.Args) > 1 && os.Args[1] == "auth" {
		root := ensemble.New(os.Stderr)
		if err := authcli.Configure(root); err != nil {
			return err
		}
		return authcli.Run(context.Background(), root, os.Args[2:], os.Stdout)
	}
	// Flags are the public configuration surface. Existing provider environment
	// defaults retain the earlier CLI setup and the original grader contract.
	cfg, err := ensemble.ConfigFromEnv()
	if err != nil {
		return err
	}
	spec := ensemble.AgentSpecFromEnv()
	guiDebug := flag.Bool("gui-debug", false, "load the gui-debug skill before accepting input")
	flag.StringVar(&spec.SkillsDir, "skills-dir", spec.SkillsDir, "skill directory (gui-debug defaults to skills)")
	port := flag.Int("port", 8088, "localhost HTTP port")
	directory := flag.String("gui-dir", "", "override embedded GUI assets with a disk directory")
	vendor := flag.String("vendor", "", "anthropic, openai or gemini")
	flag.StringVar(&cfg.Model, "model", cfg.Model, "discovered provider model")
	flag.StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL, "provider API base URL")
	flag.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "tool working directory")
	flag.StringVar(&cfg.APIKey, "api-key", cfg.APIKey, "provider credential (prefer environment for secrets)")
	flag.StringVar(&cfg.LogPath, "event-log", "event.log", "event history destination")
	flag.BoolVar(&cfg.DisableStreaming, "no-stream", cfg.DisableStreaming, "disable incremental delivery")
	flag.StringVar(&cfg.SavePath, "save", "save.json", "save loaded at startup and written at exit")
	flag.Parse()
	if *guiDebug && spec.SkillsDir == "" {
		spec.SkillsDir = "skills"
	}
	if spec.SkillsDir != "" && spec.PrimarySkill == "" {
		spec.PrimarySkill = "ensemble"
	}
	if *guiDebug {
		spec.StartupSkills = []string{"gui-debug"}
	}
	cfg.SavePath, err = filepath.Abs(cfg.SavePath)
	if err != nil {
		return err
	}
	switch *vendor {
	case "":
	case "anthropic":
		cfg.Vendor = ensemble.Anthropic
	case "openai":
		cfg.Vendor = ensemble.OpenAI
	case "gemini":
		cfg.Vendor = ensemble.Gemini
	default:
		return fmt.Errorf("unknown vendor")
	}
	cfg.BuiltinTools = true
	cfg.DataDir = "cr"
	// Each log records a different boundary. Keep provider frames, transport
	// messages and diagnostics separate so reconnect/debugging remains legible.
	debug, err := os.Create("debug.log")
	if err != nil {
		return err
	}
	defer debug.Close()
	api, err := os.Create("api.log")
	if err != nil {
		return err
	}
	defer api.Close()
	trace, err := os.Create("gui.log")
	if err != nil {
		return err
	}
	defer trace.Close()
	root := ensemble.New(debug)
	if err := authcli.Configure(root); err != nil {
		return err
	}
	root.SetAPILog(api)
	terminal := &terminal{events: make(chan ensemble.AgentObservation, 256)}
	root.AddObserver(terminal)
	displayDone := make(chan struct{})
	defer close(displayDone)
	go terminal.display(displayDone)
	a, err := root.Start(".", cfg, spec, terminal)
	if err != nil {
		return err
	}
	if a.Config().BillingMode != ensemble.CredentialChatGPT && cfg.APIKey == "" {
		return fmt.Errorf("provider API key is not set")
	}
	// Shutdown joins actual jobs. Defer order first detaches the HTTP receiver,
	// then ends Agent lifetime, then closes the application-owned log writers.
	defer func() { runErr = errors.Join(runErr, root.Close()) }()
	hub, err := gui.New(root, a, trace)
	if err != nil {
		return err
	}
	defer hub.Close()
	server := &http.Server{Addr: "127.0.0.1:" + strconv.Itoa(*port), Handler: hub.Handler(*directory)}
	// Report listener failure through its own completion channel: a bind error
	// must not leave a seemingly live application waiting only for a signal.
	failures := make(chan error, 1)
	go func() { failures <- server.ListenAndServe() }()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Terminal input shares the actor with browsers. EOF now ends the application
	// after queued work is flushed, making save-on-exit usable from a pipe.
	inputDone := make(chan error, 1)
	go func() {
		input := bufio.NewScanner(os.Stdin)
		for input.Scan() {
			text := input.Text()
			// Keep the human terminal usable while accepting the chapter's prompt
			// envelope on redirected stdin, just as the headless client does.
			var prompt struct {
				// Kind recognizes redirected prompt input without consuming ordinary prose.
				Kind string `json:"kind"`
				// Text is the human message inside that prompt envelope.
				Text string `json:"text"`
			}
			if json.Unmarshal([]byte(text), &prompt) == nil && prompt.Kind == "prompt" {
				text = prompt.Text
			}
			switch text {
			case "/interrupt":
				a.Post(ensemble.Interrupt{})
			default:
				if strings.TrimSpace(text) != "" {
					a.Post(ensemble.UserMessage{Text: text})
				}
			}
		}
		inputDone <- input.Err()
	}()
	fmt.Fprintf(os.Stderr, "Open http://localhost:%d; terminal messages while busy are hints. /interrupt ends the turn.\n", *port)
	select {
	case err := <-failures:
		return err
	case <-ctx.Done():
		return server.Close()
	case err := <-inputDone:
		return errors.Join(err, a.Flush(context.Background()), server.Close())
	}
}

// Terminal queues root observations; a slow stdout pipe cannot stop the actor.
// This is a transport receiver, not another conversation coordinator.
type terminal struct {
	events chan ensemble.AgentObservation
}

// ObserveAgent offers identified progress to the terminal's display worker.
func (t *terminal) ObserveAgent(a ensemble.Agent, o ensemble.Observation) {
	select {
	case t.events <- ensemble.AgentObservation{Agent: a, Observation: o}:
	default:
	}
}
func (t *terminal) display(done <-chan struct{}) {
	for {
		select {
		case event := <-t.events:
			switch o := event.Observation.(type) {
			case ensemble.BillingSelected:
				fmt.Fprintf(os.Stderr, "[billing: %s]\n", o.Source)
			case ensemble.PartDelta:
				fmt.Print(o.Chunk)
			case ensemble.PartFinal:
				// Deltas already printed the content; finals separate cards without
				// reading the same sentence aloud or printing its text twice.
				if o.Part.Type == "text" {
					fmt.Println()
				}
			case ensemble.ToolDispatched:
				fmt.Fprintf(os.Stderr, "[%s]\n", o.Name)
			case ensemble.ToolFinished:
				fmt.Fprintln(os.Stderr, o.Result)
			case ensemble.TurnEnded:
				if o.Error != "" {
					fmt.Fprintln(os.Stderr, o.Error)
				}
			}
		case <-done:
			return
		}
	}
}

// Deliver presents an addressed human question without blocking any Agent actor.
func (t *terminal) Deliver(message ensemble.FleetMessage) error {
	fmt.Fprintf(os.Stderr, "[from %s, message %s] %s\n", message.From, message.ID, message.Content)
	return nil
}
