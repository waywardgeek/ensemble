// This optional application owns one Ensemble, one Agent and two human views.
package main

import (
	"bufio"
	"context"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
func run() error {
	// Flags are the public configuration surface. Existing provider environment
	// defaults retain the earlier CLI setup and the original grader contract.
	cfg, err := ensemble.ConfigFromEnv()
	if err != nil {
		return err
	}
	port := flag.Int("port", 8088, "localhost HTTP port")
	directory := flag.String("gui-dir", "web/gui", "static GUI directory")
	vendor := flag.String("vendor", "", "anthropic, openai or gemini")
	flag.StringVar(&cfg.Model, "model", cfg.Model, "discovered provider model")
	flag.StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL, "provider API base URL")
	flag.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "tool working directory")
	flag.StringVar(&cfg.APIKey, "api-key", cfg.APIKey, "provider credential (prefer environment for secrets)")
	flag.StringVar(&cfg.LogPath, "event-log", "event.log", "event history destination")
	flag.BoolVar(&cfg.DisableStreaming, "no-stream", cfg.DisableStreaming, "disable incremental delivery")
	flag.Parse()
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
	if cfg.APIKey == "" {
		return fmt.Errorf("provider API key is not set")
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
	root.SetAPILog(api)
	terminal := &terminal{events: make(chan ensemble.AgentObservation, 256)}
	root.AddObserver(terminal)
	displayDone := make(chan struct{})
	defer close(displayDone)
	go terminal.display(displayDone)
	a, err := root.NewAgent(cfg, ensemble.AgentSpecFromEnv())
	if err != nil {
		return err
	}
	// Shutdown joins actual jobs. Defer order first detaches the HTTP receiver,
	// then ends Agent lifetime, then closes the application-owned log writers.
	defer a.Shutdown()
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
	// Terminal input never runs a second model loop. EOF detaches this view;
	// browsers may continue until the application receives its shutdown signal.
	go func() {
		input := bufio.NewScanner(os.Stdin)
		for input.Scan() {
			text := input.Text()
			switch text {
			case "/interrupt":
				a.Post(ensemble.Interrupt{})
			default:
				if strings.TrimSpace(text) != "" {
					a.Post(ensemble.UserMessage{Text: text})
				}
			}
		}
	}()
	fmt.Fprintf(os.Stderr, "Open http://localhost:%d; terminal messages while busy are hints. /interrupt ends the turn.\n", *port)
	select {
	case err := <-failures:
		return err
	case <-ctx.Done():
		return server.Close()
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
