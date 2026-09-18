// ch08 — Everything Is an Artifact.
//
// This exercise binary starts an agent with an HTTP server, WebSocket
// transport, and browser GUI. It runs the existing CLI loop on stdin AND
// serves a browser interface on --port (default 8088).
//
//	./ch08                         grader mode: JSON-lines on stdin, WS on --port
//	./ch08 chat                    interactive terminal mode
//	./ch08 --port 9090             WS server on port 9090
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	agent "github.com/waywardgeek/ensemble/agent"
)

const defaultPort = "8088"

func main() {
	args := os.Args[1:]
	mode := ""
	port := ""
	guiDir := ""

	var filtered []string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--port" && i+1 < len(args):
			port = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--port="):
			port = strings.TrimPrefix(args[i], "--port=")
		case args[i] == "--gui-dir" && i+1 < len(args):
			guiDir = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--gui-dir="):
			guiDir = strings.TrimPrefix(args[i], "--gui-dir=")
		default:
			filtered = append(filtered, args[i])
		}
	}
	args = filtered
	if len(args) > 0 {
		mode = args[0]
	}
	if port == "" {
		port = envOr("EN_PORT", defaultPort)
	}
	if guiDir == "" {
		guiDir = "web/gui"
	}

	cfg, err := agent.ConfigFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(2)
	}
	cfg.SystemPrompt = "You are a helpful assistant."
	cfg.MaxTokens = 16384
	logPath := envOr("CH02_LOG", "ensemble.log")

	switch mode {
	case "--help", "-h", "help":
		fmt.Fprintf(os.Stdout, "ensemble — Everything Is an Artifact.\n\n")
		fmt.Fprintf(os.Stdout, "  ensemble                    grader mode + HTTP server\n")
		fmt.Fprintf(os.Stdout, "  ensemble chat               interactive terminal mode\n")
		fmt.Fprintf(os.Stdout, "  ensemble --port PORT         HTTP/WebSocket server port (default %s)\n", defaultPort)
		return
	case "chat":
		runChat(cfg, logPath)
	case "":
		runMain(cfg, logPath, port, guiDir)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", mode)
		os.Exit(2)
	}
}

func runMain(cfg agent.Config, logPath, port, guiDir string) {
	a := agent.NewAgent(cfg, logPath)
	defer a.Shutdown()

	// Register the "think" tool (same as ch06/ch07).
	a.RegisterTool("think", "Think for a specified number of seconds",
		json.RawMessage(`{"type":"object","properties":{"seconds":{"type":"number","description":"How long to think"},"thought":{"type":"string","description":"What to think about"}},"required":["seconds","thought"]}`),
		func(args json.RawMessage) (string, error) {
			var p struct {
				Seconds float64 `json:"seconds"`
				Thought string  `json:"thought"`
			}
			json.Unmarshal(args, &p)
			time.Sleep(time.Duration(p.Seconds * float64(time.Second)))
			return fmt.Sprintf("Thought about: %s", p.Thought), nil
		})

	actor := a.NewActor()
	gate := agent.NewPauseGate()
	actor.SetPauseGate(gate)

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	var outMu sync.Mutex

	emitLocked := func(v any) {
		outMu.Lock()
		defer outMu.Unlock()
		b, _ := json.Marshal(v)
		out.Write(b)
		out.WriteByte('\n')
		out.Flush()
	}

	// Attach stdout observer for grader.
	actor.Attach(observerFunc(func(obs agent.Observation) {
		emitLocked(observationJSON(obs))
	}))

	// Start HTTP/WebSocket server.
	hub := agent.NewWSHub(gate, func(msg agent.Inbound) {
		actor.Send(msg)
	}, "gui.log", a.EventLog())
	defer hub.Close()
	actor.Attach(hub)

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(guiDir))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		fs.ServeHTTP(w, r)
	}))
	mux.HandleFunc("/ws", hub.ServeWS)
	srv := &http.Server{Addr: ":" + port, Handler: mux}
	go srv.ListenAndServe()
	defer srv.Close()

	// Start actor loop.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go actor.Run(ctx)

	// Read stdin for grader protocol.
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		var msg stdinMsg
		if json.Unmarshal([]byte(line), &msg) != nil {
			emitLocked(map[string]string{"error": "bad input"})
			continue
		}
		if msg.Kind != nil {
			switch *msg.Kind {
			case "prompt":
				if msg.Text == nil {
					emitLocked(map[string]string{"error": "prompt requires text"})
					continue
				}
				actor.Send(agent.UserMessage{Text: *msg.Text})
				obs, err := actor.Wait(ctx, func(o agent.Observation) bool {
					_, ok := o.(agent.TurnEnded)
					return ok
				})
				if err != nil {
					emitLocked(map[string]string{"error": err.Error()})
					continue
				}
				ended := obs.(agent.TurnEnded)
				if ended.Err != "" {
					emitLocked(map[string]string{"error": ended.Err})
				} else {
					emitLocked(map[string]string{"assistant": ended.Text})
				}
			case "hint":
				if msg.Text != nil {
					actor.Send(agent.Hint{Text: *msg.Text})
				}
			case "interrupt":
				actor.Send(agent.Interrupt{})
			}
			continue
		}
		if msg.User != nil {
			reply, err := a.Ask(*msg.User)
			if err != nil {
				emitLocked(map[string]string{"error": err.Error()})
				continue
			}
			emitLocked(map[string]string{"assistant": reply})
			continue
		}
		emitLocked(map[string]string{"error": "no recognized field"})
	}

	_ = actor.Shutdown()
	emitLocked(map[string]any{"usage": a.Usage()})
}

func runChat(cfg agent.Config, logPath string) {
	a := agent.NewAgent(cfg, logPath)
	defer a.Shutdown()

	in := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	fmt.Fprint(os.Stderr, "ensemble — type a message, ctrl-D to exit\n> ")
	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			fmt.Fprint(os.Stderr, "> ")
			continue
		}
		reply, err := a.Ask(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		} else {
			fmt.Fprintln(out, reply)
			out.Flush()
		}
		fmt.Fprint(os.Stderr, "> ")
	}
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

func observationJSON(obs agent.Observation) map[string]any {
	switch v := obs.(type) {
	case agent.PartDelta:
		m := map[string]any{"observation": "part_delta", "part_id": v.PartID, "kind": v.Kind.String(), "chunk": v.Chunk}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		return m
	case agent.PartFinal:
		m := map[string]any{"observation": "part_final", "part_id": v.PartID, "seq": v.Seq}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		switch p := v.Part.(type) {
		case agent.TextPart:
			m["text"] = p.Text
		case agent.ToolCallPart:
			m["tool"] = p.Name
			m["args"] = string(p.Args)
		case agent.OpaquePart:
			m["opaque"] = true
		}
		return m
	case agent.StateChanged:
		m := map[string]any{"observation": "state_changed", "from": v.From.String(), "to": v.To.String()}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		return m
	case agent.TurnEnded:
		m := map[string]any{"observation": "turn_ended", "text": v.Text}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		if v.Err != "" {
			m["error"] = v.Err
		}
		return m
	case agent.ToolDispatched:
		m := map[string]any{"observation": "tool_dispatched", "call_id": v.CallID, "name": v.Name}
		if v.Input != nil {
			m["input"] = json.RawMessage(v.Input)
		}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		return m
	case agent.ToolFinished:
		m := map[string]any{"observation": "tool_finished", "call_id": v.CallID, "result": v.Result, "is_error": v.IsError}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		return m
	}
	return map[string]any{"observation": "unknown"}
}

type stdinMsg struct {
	Kind *string `json:"kind"`
	Text *string `json:"text"`
	User *string `json:"user"`
}

// observerFunc adapts a function to the Observer interface.
type observerFunc func(agent.Observation)

func (f observerFunc) Observe(o agent.Observation) { f(o) }

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
