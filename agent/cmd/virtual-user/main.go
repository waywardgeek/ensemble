// Command virtual-user drives the coding agent's GUI as an automated tester.
//
// It connects to the hub's WebSocket, discovers GUI tools from the browser's
// MCP server, and runs an LLM conversation loop that sees the DOM, clicks
// buttons, types into fields, and files bug reports. It sees what a human
// sees and interacts how a human would — nothing more.
//
// Usage:
//
//	ANTHROPIC_API_KEY=... virtual-user --task "test the prompt flow"
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	agent "github.com/waywardgeek/ensemble/agent"
)

const virtualUserPrompt = `You are a virtual user testing a coding agent through its GUI. You can see the GUI DOM, click buttons, type into text fields, and hear TTS output.

You CANNOT read files, edit code, or run commands. You can only interact through the GUI, exactly as a human user would.

Your tools:
- gui_snapshot: see the current GUI state (auto-injected each round)
- gui_click(selector): click a button, tab, or link
- gui_input(selector, text): type text into an input field
- tts_queue: check what's being spoken via TTS
- file_report(title, body): file a bug report or observation

Your task will be given to you as a prompt. Explore the GUI methodically:
1. Start by taking a snapshot to understand the current state
2. Try interacting with the elements you see
3. Report any bugs, confusing behavior, or missing functionality
4. Be thorough — test edge cases, empty inputs, rapid clicks`

func main() {
	agentURL := flag.String("agent-url", "ws://localhost:8082/ws", "hub WebSocket URL")
	reportFile := flag.String("report-file", "reports.md", "path for file_report output")
	task := flag.String("task", "", "task prompt for the virtual user")
	flag.Parse()

	// Accept task from --task or remaining args.
	taskText := *task
	if taskText == "" && flag.NArg() > 0 {
		taskText = strings.Join(flag.Args(), " ")
	}
	if taskText == "" {
		fmt.Fprintln(os.Stderr, "usage: virtual-user --task 'test the prompt flow'")
		os.Exit(1)
	}

	// Build config from env vars (ANTHROPIC_API_KEY, MODEL, etc.).
	cfg, err := agent.ConfigFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	cfg.SystemPrompt = virtualUserPrompt

	// Create a bare agent — no builtin tools. All tools come from MCP
	// or explicit registration below.
	a := agent.NewBareAgent(cfg, "virtual-user.jsonl")
	defer a.Shutdown()

	// Register file_report: appends markdown sections to the report file.
	reportPath := *reportFile
	a.RegisterTool("file_report",
		"File a bug report or observation about the GUI",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"title": {"type": "string", "description": "Short title for the report"},
				"body":  {"type": "string", "description": "Detailed description of the bug or observation"}
			},
			"required": ["title", "body"]
		}`),
		func(args json.RawMessage) (string, error) {
			var params struct {
				Title string `json:"title"`
				Body  string `json:"body"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", fmt.Errorf("parse args: %w", err)
			}
			f, err := os.OpenFile(reportPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return "", fmt.Errorf("open report file: %w", err)
			}
			defer f.Close()
			fmt.Fprintf(f, "## %s\n\n%s\n\n---\n\n", params.Title, params.Body)
			return "Report filed: " + params.Title, nil
		},
	)

	// Connect to the hub WebSocket with source tag "vu".
	log.Printf("connecting to %s...", *agentURL)
	transport, err := agent.NewMCPClientWSTransport(*agentURL, "vu")
	if err != nil {
		log.Fatalf("connect to hub: %v", err)
	}
	log.Println("WebSocket connected, starting MCP handshake...")

	if err := a.ConnectMCP(transport); err != nil {
		log.Fatalf("MCP handshake: %v", err)
	}

	log.Printf("virtual user connected, MCP tools discovered")
	log.Printf("task: %s", taskText)

	// Run the conversation loop. The agent discovers gui_snapshot,
	// gui_click, gui_input, and tts_queue from the browser's MCP server.
	// gui_snapshot is ephemeral: the engine auto-injects it each round.
	reply, err := a.Ask(taskText)
	if err != nil {
		log.Fatalf("conversation failed: %v", err)
	}
	fmt.Println(reply)
}
