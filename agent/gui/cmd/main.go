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
	configPath, err := configPathFromArgs(os.Args[1:])
	if err != nil {
		return err
	}
	var application *gui.ApplicationConfigStore
	var cfg ensemble.Config
	var spec ensemble.AgentSpec
	if configPath != "" {
		application, err = gui.LoadApplicationConfig(configPath)
		if err != nil {
			return err
		}
		cfg, spec, err = application.Snapshot().RuntimeConfig()
	} else {
		cfg, err = ensemble.ConfigFromEnv()
		spec = ensemble.AgentSpecFromEnv()
	}
	if err != nil {
		return err
	}
	flag.String("config", configPath, "explicit application settings file; ignores provider environment")
	guiDebug := flag.Bool("gui-debug", false, "load the gui-debug skill before accepting input")
	flag.StringVar(&spec.SkillsDir, "skills-dir", spec.SkillsDir, "skill directory (gui-debug defaults to skills)")
	defaultPort := 8088
	if application != nil {
		defaultPort = application.Snapshot().GUIPort
	}
	port := flag.Int("port", defaultPort, "localhost HTTP port")
	directory := flag.String("gui-dir", "", "override embedded GUI assets with a disk directory")
	vendor := flag.String("vendor", "", "anthropic, openai or gemini")
	flag.StringVar(&cfg.Model, "model", cfg.Model, "discovered provider model")
	flag.StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL, "provider API base URL")
	flag.StringVar(&cfg.WakeOnLANMAC, "wol-mac", cfg.WakeOnLANMAC, "Wake-on-LAN MAC for an unreachable local provider")
	flag.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "tool working directory")
	// Never expose a loaded credential as a flag default: Go prints defaults
	// in both help and flag-error output.
	apiKeyOverride := flag.String("api-key", "", "legacy provider credential override (prefer a credential file)")
	if cfg.LogPath == "" {
		cfg.LogPath = "event.log"
	}
	flag.StringVar(&cfg.LogPath, "event-log", cfg.LogPath, "event history destination")
	flag.BoolVar(&cfg.DisableStreaming, "no-stream", cfg.DisableStreaming, "disable incremental delivery")
	if cfg.SavePath == "" {
		cfg.SavePath = "save.json"
	}
	flag.StringVar(&cfg.SavePath, "save", cfg.SavePath, "save loaded at startup and written at exit")
	flag.Parse()
	if *apiKeyOverride != "" && application == nil {
		cfg.APIKey = *apiKeyOverride
	}
	if application != nil {
		var incompatible string
		flag.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "config":
			case "port", "gui-dir", "gui-debug":
				fmt.Fprintf(os.Stderr, "GUI launch override: --%s=%s\n", f.Name, f.Value.String())
			default:
				incompatible = f.Name
			}
		})
		if incompatible != "" {
			return fmt.Errorf("--%s cannot override --config; edit the settings file", incompatible)
		}
		if len(flag.Args()) != 0 {
			return fmt.Errorf("unexpected arguments after --config launch")
		}
		fmt.Fprintf(os.Stderr, "Application settings: %s (provider environment ignored)\n", application.Path())
	}
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
	registryDirectory := "."
	if application == nil {
		cfg.DataDir = "cr"
	} else {
		registryDirectory = filepath.Dir(application.Path())
	}
	// Each log records a different boundary. Keep provider frames, transport
	// messages and diagnostics separate so reconnect/debugging remains legible.
	debug, err := os.Create(filepath.Join(registryDirectory, "debug.log"))
	if err != nil {
		return err
	}
	defer debug.Close()
	api, err := os.Create(filepath.Join(registryDirectory, "api.log"))
	if err != nil {
		return err
	}
	defer api.Close()
	trace, err := os.Create(filepath.Join(registryDirectory, "gui.log"))
	if err != nil {
		return err
	}
	defer trace.Close()
	root := ensemble.New(debug)
	if application != nil {
		settings := application.Snapshot()
		root.ConfigureAuth(ensemble.AuthOptions{Directory: settings.AuthDirectory, Issuer: settings.AuthIssuer})
	} else if err := authcli.Configure(root); err != nil {
		return err
	}
	root.SetAPILog(api)
	terminal := &terminal{events: make(chan ensemble.AgentObservation, 256)}
	root.AddObserver(terminal)
	displayDone := make(chan struct{})
	defer close(displayDone)
	go terminal.display(displayDone)
	a, err := root.Start(registryDirectory, cfg, spec, terminal)
	if err != nil {
		return err
	}
	if a.Config().Credentials.Kind() == ensemble.CredentialAPIKey && cfg.APIKey == "" {
		return fmt.Errorf("provider API key is not set")
	}
	// Shutdown joins actual jobs. Defer order first detaches the HTTP receiver,
	// then ends Agent lifetime, then closes the application-owned log writers.
	defer func() { runErr = errors.Join(runErr, root.Close()) }()
	var hub *gui.Server
	if application == nil {
		hub, err = gui.New(root, a, trace)
	} else {
		hub, err = gui.NewConfigured(root, a, trace, application)
	}
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
