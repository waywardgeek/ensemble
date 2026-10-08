//go:build ignore

// A standalone public consumer for actual native speech over retained provider
// facts. No model request can reach a provider. Adapted from the accepted Chapter
// 7 student's replay helper, not a reviewer or historical implementation.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"example.com/ensemble"
	gui "example.com/ensemble-gui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("supply provider and retained event log")
	}
	app := ensemble.New(os.Stderr)
	defer app.Close()
	workspace, err := os.Getwd()
	if err != nil {
		return err
	}
	agent, err := app.NewAgent(ensemble.Config{Vendor: os.Args[1], Model: "retained provider text — no model work", APIKey: "disabled", BaseURL: "http://127.0.0.1:1", Workspace: workspace, LogPath: filepath.Join(workspace, "readmitted.jsonl")})
	if err != nil {
		return err
	}
	if err := admit(agent, os.Args[2]); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()
	origin := "http://" + listener.Addr().String()
	client, err := gui.NewServer(app, agent.ID(), origin, nil, gui.ServerOptions{PreferencesPath: filepath.Join(workspace, "preferences.json")})
	if err != nil {
		return err
	}
	defer client.Close()
	server := &http.Server{Handler: client}
	defer server.Close()
	go server.Serve(listener)
	fmt.Println(origin)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	<-stop
	return nil
}
func admit(agent *ensemble.Agent, path string) error {
	input, err := os.Open(path)
	if err != nil {
		return err
	}
	defer input.Close()
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	if !scanner.Scan() {
		return fmt.Errorf("missing retained log header")
	}
	var header struct {
		Version int `json:"log_version"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil || header.Version != 1 {
		return fmt.Errorf("invalid retained log header")
	}
	for scanner.Scan() {
		var event ensemble.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return err
		}
		if err := agent.Append(event); err != nil {
			return err
		}
	}
	return scanner.Err()
}
