package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"example.com/ensemble"
	gui "example.com/ensemble-gui"
	"example.com/ensemble/cli"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	port := flag.Int("port", 8088, "local listening port (0 chooses a free port)")
	terminal := flag.Bool("terminal", false, "attach the human terminal to this Agent")
	tracePath := flag.String("gui-log", "", "optional conversation trace path")
	preferencesPath := flag.String("preferences", ".ensemble/gui-preferences.json", "display preferences file")
	policyPath := flag.String("policy", ".ensemble/agent-policy.json", "Agent execution policy file")
	flag.Parse()
	if *port < 0 || *port > 65535 {
		return fmt.Errorf("invalid port")
	}
	if value := os.Getenv("EN_DISABLE_STREAMING"); value != "" && value != "0" && value != "1" {
		return fmt.Errorf("EN_DISABLE_STREAMING must be 0 or 1")
	}
	app := ensemble.New(os.Stderr)
	defer app.Close()
	config := cli.Configuration(app)
	config.PolicyPath = *policyPath
	prefResolved, err := filepath.Abs(*preferencesPath)
	if err != nil {
		return fmt.Errorf("invalid preferences path")
	}
	policyResolved, err := filepath.Abs(*policyPath)
	if err != nil {
		return fmt.Errorf("invalid policy path")
	}
	if prefResolved == policyResolved {
		return fmt.Errorf("preferences and policy need separate paths")
	}
	config.PolicyPath = policyResolved
	config.Builtins = []string{"read_file", "list_directory", "search_files", "write_file", "edit_file", "run_command", "wait_for_job", "send_input", "kill_job", "tool_limits"}
	if config.Skills != nil {
		config.Builtins = append(config.Builtins, "load_skill", "unload_skill")
	}
	config.MaxTokens = 4096
	a, err := app.NewAgent(config)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		return err
	}
	defer listener.Close()
	origin := "http://" + listener.Addr().String()
	var trace *os.File
	if *tracePath != "" {
		trace, err = os.OpenFile(*tracePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("cannot create GUI trace")
		}
		defer trace.Close()
	}
	var traceWriter io.Writer
	if trace != nil {
		traceWriter = trace
	}
	server, err := gui.NewServer(app, a.ID(), origin, traceWriter, gui.ServerOptions{PreferencesPath: prefResolved})
	if err != nil {
		return err
	}
	defer server.Close()
	httpServer := &http.Server{Handler: server}
	defer httpServer.Close()
	failures := make(chan error, 2)
	go func() { failures <- httpServer.Serve(listener) }()
	terminalDone := make(chan struct{})
	if *terminal {
		defer func() { _ = os.Stdin.Close(); _ = a.Close(); <-terminalDone }()
		go func() {
			defer close(terminalDone)
			err := cli.Chat(app, a, os.Stdin, os.Stdout, true)
			if err != nil {
				failures <- err
			}
		}()
	}
	fmt.Fprintln(os.Stdout, origin)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	select {
	case <-stop:
		return nil
	case err := <-failures:
		if errors.Is(err, cli.ErrQuit) || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
