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

// blindUserPrompt drives an observer that perceives the agent only through speech.
//
// The point is not politeness. A sighted observer completes every task through the
// DOM, so the speech channel is never load bearing and a defect in it can never
// fail a run. Removing sight makes speech the only path to success, which turns a
// silent omission into a task failure instead of a detail nobody checks.
const blindUserPrompt = `You are testing an AI coding agent's interface as a user who cannot see the screen.

You have no vision. There is no snapshot tool and no way to read the page. Everything
you know about what the agent is doing arrives as speech, which you read with
tts_transcript.

Your tools:
- tts_bypass(enabled): call this FIRST with enabled=true, so speech is recorded
  without playing audio and you are not forced to wait in real time
- tts_transcript(since): everything spoken so far, in order. Pass since=<last_seq>
  to receive only what is new
- gui_input(selector, text): type text into an input field
- gui_submit(selector): submit the input
- wait_for_idle(timeout_seconds): wait for the agent to finish its turn
- sleep(seconds): wait
- file_report(title, body): file your findings

How to work:
1. Enable tts_bypass.
2. Type your prompt into the chat input and submit it.
3. Wait for the agent to go idle, reading tts_transcript as you go.
4. Answer the question you were asked using only what you heard.

The rule that matters most: if the information you needed was never spoken, then you
did not learn it. Say so plainly, and report exactly what you did hear instead. Never
infer, guess, or reconstruct a plausible answer from context. "I could not hear X" is
a correct and useful result. A confident answer you did not actually hear is the worst
possible outcome, because it hides a real defect from the people who could fix it.`

func main() {
	agentURL := flag.String("agent-url", "ws://localhost:8082/ws", "hub WebSocket URL")
	reportFile := flag.String("report-file", "reports.md", "path for file_report output")
	task := flag.String("task", "", "task prompt for the virtual user")
	persona := flag.String("persona", "sighted",
		"sighted (reads the DOM) or blind (perceives only speech output)")
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
	switch *persona {
	case "sighted":
		cfg.SystemPrompt = virtualUserPrompt
	case "blind":
		cfg.SystemPrompt = blindUserPrompt
	default:
		fmt.Fprintf(os.Stderr, "unknown persona %q: want sighted or blind\n", *persona)
		os.Exit(1)
	}

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

	// A blind observer has to actually lose the capability, not merely be asked not
	// to use it. An agent that can still see will finish the task by seeing, and the
	// speech channel it exists to exercise stays untested. gui_snapshot matters most
	// here: it is ephemeral, so the engine injects the whole DOM every round unasked.
	if *persona == "blind" {
		for _, name := range []string{"gui_snapshot", "gui_click"} {
			a.RemoveTool(name)
		}
		log.Printf("persona=blind: removed gui_snapshot and gui_click; perception is speech only")
	}

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
