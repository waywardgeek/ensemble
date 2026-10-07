package main

import (
	"bufio"
	"context"
	"encoding/json"
	"example.com/ensemble"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
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
	return ensemble.Config{Vendor: vendor, APIKey: value("API_KEY"), Model: value("MODEL"), BaseURL: value("BASE_URL"), ResolvedModel: os.Getenv("LLM_RESOLVED_MODEL"), LogPath: path}
}
func run(input io.Reader, output, diagnostics io.Writer) error {
	return runArgs(os.Args[1:], input, output, diagnostics)
}
func runArgs(args []string, input io.Reader, output, diagnostics io.Writer) error {
	app := ensemble.New(diagnostics)
	config := configuration(app)
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
		case len(args) == 2 && args[0] == "render":
			path = args[1]
		default:
			return fmt.Errorf("usage: ensemble [chat | protocol | dump | render LOG]")
		}
		agent, err := app.Load(path, config)
		if err != nil {
			return err
		}
		var data []byte
		if args[0] == "dump" {
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
	return runProtocol(app, agent, input, output)
}

func runProtocol(app ensemble.ClientOwner, agent *ensemble.Agent, input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 16*1024*1024)
	encoder := json.NewEncoder(output)
	for scanner.Scan() {
		line := scanner.Bytes()
		if strings.TrimSpace(string(line)) == "" {
			continue
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(line, &fields) != nil || len(fields) != 1 {
			return fmt.Errorf("exactly one valid directive is required")
		}
		request := ensemble.ClientRequest{AgentID: agent.ID()}
		ack := ""
		switch {
		case fields["user"] != nil:
			var text string
			if string(fields["user"]) == "null" || json.Unmarshal(fields["user"], &text) != nil {
				return fmt.Errorf("invalid user input")
			}
			request.Prompt = &text
		case fields["ephemeral"] != nil:
			var text string
			if string(fields["ephemeral"]) == "null" || json.Unmarshal(fields["ephemeral"], &text) != nil {
				return fmt.Errorf("invalid ephemeral input")
			}
			request.Ephemeral = &text
			ack = "ephemeral"
		case fields["redact"] != nil:
			var r ensemble.Redaction
			if json.Unmarshal(fields["redact"], &r) != nil {
				return fmt.Errorf("invalid redaction")
			}
			request.Redact = &r
			ack = "redact"
		default:
			return fmt.Errorf("unknown directive")
		}
		result, err := app.Submit(context.Background(), request)
		if err != nil {
			return err
		}
		if ack != "" {
			err = encoder.Encode(map[string]string{"ack": ack})
		} else {
			err = encoder.Encode(map[string]string{"assistant": result.Text})
		}
		if err != nil {
			return fmt.Errorf("cannot write response")
		}
	}
	if scanner.Err() != nil {
		return fmt.Errorf("cannot read input")
	}
	if err := agent.Close(); err != nil {
		return err
	}
	return encoder.Encode(struct {
		Usage ensemble.Usage `json:"usage"`
	}{agent.Usage()})
}
