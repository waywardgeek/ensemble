package cli

import (
	"example.com/ensemble"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Run preserves the standalone CLI's human and protocol modes.
func Run(args []string, input io.Reader, output, diagnostics io.Writer) error {
	return runArgs(args, input, output, diagnostics)
}

// Configuration reads the inherited environment without printing credentials.
func Configuration(owner ensemble.ClientOwner) ensemble.Config { return configuration(owner) }
func configuration(owner ensemble.ClientOwner) ensemble.Config {
	vendor := os.Getenv("LLM_VENDOR")
	if vendor == "" {
		vendor = "anthropic"
	}
	value := func(name string) string {
		if v := os.Getenv("LLM_" + name); v != "" {
			return v
		}
		return os.Getenv(strings.ToUpper(vendor) + "_" + name)
	}
	path := os.Getenv("CH02_LOG")
	if path == "" {
		path = filepath.Base(os.Args[0]) + ".log"
	}
	return ensemble.Config{DisableStreaming: os.Getenv("EN_DISABLE_STREAMING") == "1", Vendor: vendor, APIKey: value("API_KEY"), Model: value("MODEL"), BaseURL: value("BASE_URL"), ResolvedModel: os.Getenv("LLM_RESOLVED_MODEL"), LogPath: path}
}
func run(input io.Reader, output, diagnostics io.Writer) error {
	return runArgs(os.Args[1:], input, output, diagnostics)
}
func runArgs(args []string, input io.Reader, output, diagnostics io.Writer) error {
	if value := os.Getenv("EN_DISABLE_STREAMING"); value != "" && value != "0" && value != "1" {
		return fmt.Errorf("EN_DISABLE_STREAMING must be 0 or 1")
	}
	app := ensemble.New(diagnostics)
	config := configuration(app)
	observe := false
	if len(args) == 2 && args[0] == "protocol" && args[1] == "--observe" {
		observe = true
		args = args[:1]
	}
	mode := "protocol"
	if isTerminal(app, input) && isTerminal(app, output) {
		mode = "chat"
	}
	if len(args) == 1 && (args[0] == "chat" || args[0] == "protocol") {
		mode = args[0]
	} else if len(args) > 0 {
		var path string
		switch {
		case len(args) == 1 && args[0] == "dump":
			path = config.LogPath
		case len(args) == 3 && args[0] == "replay":
			path = args[1]
		case len(args) == 2 && args[0] == "render":
			path = args[1]
		default:
			return fmt.Errorf("usage: ensemble [chat | protocol | dump | render LOG | replay LOG SEQ]")
		}
		agent, err := app.Load(path, config)
		if err != nil {
			return err
		}
		var data []byte
		if args[0] == "replay" {
			var sequence uint64
			sequence, err = strconv.ParseUint(args[2], 10, 64)
			if err == nil {
				data, err = agent.ReconstructRequest(sequence)
			}
		} else if args[0] == "dump" {
			data, err = agent.Dump()
		} else {
			data, err = agent.Render(config)
		}
		if err != nil {
			return err
		}
		if len(data) == 0 || data[len(data)-1] != '\n' {
			data = append(data, '\n')
		}
		_, err = output.Write(data)
		return err
	}
	config.Builtins = []string{"read_file", "list_directory", "search_files", "write_file", "edit_file", "run_command", "wait_for_job", "send_input", "kill_job", "tool_limits"}
	config.MaxTokens = 4096
	agent, err := app.NewAgent(config)
	if err != nil {
		return err
	}
	defer agent.Close()
	if mode == "chat" {
		return runChat(app, agent, input, output)
	}
	return runProtocolObserved(app, agent, input, output, observe)
}
