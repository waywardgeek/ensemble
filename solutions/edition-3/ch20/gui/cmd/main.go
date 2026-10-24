// gui hosts an optional browser and terminal view through Ensemble's public API.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ensemble"
	"ensemble/gui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "auth" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := ensemble.RunAuthCommand(ctx, ensemble.NewAuth(ensemble.AuthOptionsFromEnv(os.Getenv)), os.Args[2:], os.Stdout, nil); err != nil {
			log.Print(err)
			os.Exit(1)
		}
		return
	}
	guiDebug := flag.Bool("gui-debug", false, "load gui-debug at startup")
	skillsDir := flag.String("skills-dir", "", "skill directory (default: ./skills when ensemble is present)")
	port := flag.Int("port", 8088, "loopback HTTP port")
	assets := flag.String("gui-dir", "", "GUI asset directory (absolute paths work outside cwd)")
	safe := flag.Bool("safe-mode", false, "disable command and job tools")
	web := flag.Bool("no-web-search", false, "disable web-search skill")
	sandboxRoot := flag.String("sandbox", "", "confine Agent files and commands to this directory")
	settings := flag.String("settings", "", "GUI settings JSON path (default: cr/settings.json under the working directory)")
	save := flag.String("save", "", "recovery archive (default: save.json)")
	logMode := flag.String("log-mode", "", "archive retention: full or dialogue_only")
	pause := flag.Bool("pause-on-start", false, "pause before the first model request")
	speaking := flag.Bool("speaking", false, "enable Agent speech requests")
	mcpListen := flag.String("mcp-listen", "", "serve a sibling MCP listener on loopback while the GUI runs")
	mcpPipe := flag.Bool("mcp-pipe", false, "serve MCP on stdin/stdout until EOF")
	mcpSocket := flag.String("mcp-socket", "", "serve MCP on a dedicated TCP address")
	flag.Parse()
	protocolMode := *mcpPipe || *mcpSocket != ""
	if (*mcpPipe && *mcpSocket != "") || (protocolMode && *mcpListen != "") || flag.NArg() != 0 {
		log.Print("usage: gui [--mcp-pipe | --mcp-socket ADDRESS]")
		return
	}
	config, err := ensemble.ConfigFromEnv(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	if err := configureSkills(&config, *skillsDir, *guiDebug); err != nil {
		log.Fatal(err)
	}
	if protocolMode && config.Model == "" {
		config.Model = "offline-mcp"
	}
	if config.DataDir == "" {
		config.DataDir = "cr"
	}
	if config.SavePath == "" {
		config.SavePath = "save.json"
	}
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "safe-mode":
			config.SafeMode = *safe
		case "no-web-search":
			config.EnableWebSearch = !*web
		case "sandbox":
			config.SandboxRoot = *sandboxRoot
		case "save":
			config.SavePath = *save
		case "log-mode":
			config.LogMode = ensemble.LogMode(*logMode)
		case "pause-on-start":
			config.PauseOnStart = *pause
		case "speaking":
			config.Speaking = *speaking
		}
	})
	if config.SandboxRoot != "" {
		config.BaseDir = config.SandboxRoot
	}
	root := ensemble.New(log.Default())
	if err := root.SetAuth(ensemble.NewAuth(ensemble.AuthOptionsFromEnv(os.Getenv))); err != nil {
		log.Fatal(err)
	}
	var trace *ensemble.DiagnosticLog
	ready, normalExit := false, false
	defer func() {
		if trace != nil {
			defer trace.Close()
		}
		if err := root.Shutdown(); err != nil {
			log.Printf("shutdown/save: %v", err)
			return
		}
		if ready && normalExit && trace != nil {
			if err := trace.Retain(); err != nil {
				log.Printf("diagnostic retention: %v", err)
			}
		}
	}()
	if err := installAgentFactory(root); err != nil {
		log.Fatal(err)
	}
	var agent ensemble.Agent
	var app *gui.Server
	if protocolMode {
		agent, err = root.CreateAgent(context.Background(), config)
	} else {
		app, agent, err = gui.NewServerFromConfig(root, config, *assets, *settings)
		if err == nil {
			root.SetHumanQuestionSink(app.Hub.PresentHumanQuestion)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
	if protocolMode {
		ready = true
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		var err error
		if *mcpPipe {
			err = ensemble.ServeMCP(ctx, agent, ensemble.NewMCPStream(ensemble.SplitReadWriteCloser{Reader: os.Stdin, Writer: os.Stdout, CloseFunc: os.Stdin.Close}))
		} else {
			err = ensemble.ListenMCP(ctx, agent, *mcpSocket)
		}
		normalExit = err == nil
		if err != nil && ctx.Err() == nil {
			log.Print(err)
		}
		return
	}
	if r := agent.Recovery(); r.Path != "" {
		log.Printf("loaded %s through sequence %d (backup=%t)", r.Path, r.AsOf, r.Backup)
	}
	defer app.Close()
	app.DefaultAgent = agent.ID()
	trace, err = ensemble.OpenAgentDiagnosticLog(agent, "gui.log", func() int { return *agent.Config().DiagnosticMaxRecords })
	if err != nil {
		log.Print(err)
		return
	}

	app.Trace = trace
	if agent.Recovery().Path == "" {
		if err = agent.Save(); err != nil {
			log.Printf("startup save: %v", err)
			return
		}
	}
	if err = trace.Retain(); err != nil {
		log.Print(err)
		return
	}
	ready = true
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *mcpListen != "" {
		listener, err := ensemble.StartMCPListener(ctx, agent, *mcpListen)
		if err != nil {
			log.Print(err)
			return
		}
		defer listener.Close()
	}
	server := &http.Server{Addr: addr, Handler: app.Handler(app.Assets), ReadHeaderTimeout: 5 * time.Second}
	httpListener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Print(err)
		return
	}
	defer httpListener.Close()
	terminalDone := app.Terminal(ctx, os.Stdin, os.Stdout)
	drained := make(chan bool, 1)
	go func() {
		completed := false
		select {
		case <-ctx.Done():
		case <-terminalDone:
			completed = ctx.Err() == nil
		}
		drained <- completed
		stop()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	log.Printf("Ensemble GUI http://%s · assets %s · terminal EOF saves and exits", httpListener.Addr().String(), app.Assets)
	if err := server.Serve(httpListener); err != nil && err != http.ErrServerClosed {
		log.Print(err)
	} else {
		normalExit = <-drained
	}
}
