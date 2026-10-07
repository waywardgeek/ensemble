package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"example.com/ensemble"
)

func main() {
	if err := run(os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(input io.Reader, output, diagnostics io.Writer) error {
	if len(os.Args) != 1 {
		return fmt.Errorf("usage: ensemble (JSON lines on stdin)")
	}
	app := ensemble.New(diagnostics)
	agent, err := app.NewAgent(ensemble.Config{APIKey: os.Getenv("ANTHROPIC_API_KEY"), Model: os.Getenv("ANTHROPIC_MODEL"), BaseURL: os.Getenv("ANTHROPIC_BASE_URL")})
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 16*1024*1024)
	encoder := json.NewEncoder(output)
	for scanner.Scan() {
		line := scanner.Bytes()
		if strings.TrimSpace(string(line)) == "" {
			continue
		}
		var request struct {
			User string `json:"user"`
		}
		if err := json.Unmarshal(line, &request); err != nil {
			return fmt.Errorf("malformed input JSON")
		}
		if strings.TrimSpace(request.User) == "" {
			return fmt.Errorf("question must be nonempty")
		}
		answer, err := agent.Ask(context.Background(), request.User)
		if err != nil {
			return err
		}
		if err := encoder.Encode(struct {
			Assistant string `json:"assistant"`
		}{answer}); err != nil {
			return fmt.Errorf("write answer: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	return encoder.Encode(struct {
		Usage ensemble.Usage `json:"usage"`
	}{agent.Usage()})
}
