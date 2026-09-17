package main

// ch06 — the actor upgrade: two seams and a loop.
//
//	./ch06              grader mode: JSON-lines protocol on stdin/stdout
//	./ch06 chat         the interactive loop
//	./ch06 render LOG   play LOG -> context -> render; print the request JSON
//	./ch06 dump         write the event log as JSON-lines
//	./ch06 --help       print the commands table

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/waywardgeek/coding-agents-course/agent/internal/common"
	"github.com/waywardgeek/coding-agents-course/agent/internal/jobs"
	"github.com/waywardgeek/coding-agents-course/agent/internal/llm"
	"github.com/waywardgeek/coding-agents-course/agent/internal/tools"
)

const fallbackName = "ch06"

func progName() string {
	base := filepath.Base(os.Args[0])
	if base == "." || base == string(os.PathSeparator) || base == "" {
		return fallbackName
	}
	return base
}

func defaultLogPath() string { return progName() + ".log" }

func usage(w io.Writer) {
	p := progName()
	fmt.Fprintf(w, `%[1]s — two seams and a loop.

usage:
  %[1]s                grader mode: JSON-lines protocol on stdin/stdout
  %[1]s chat           interactive loop; type a message, ctrl-D to exit
  %[1]s render LOG     play LOG -> context -> render; print the request JSON
  %[1]s dump           write the event log as JSON-lines
  %[1]s --help         print this table

environment:
  LLM_VENDOR           anthropic (default), openai or gemini
  LLM_MODEL            model id; overrides the vendor default
  LLM_API_KEY          API key (or ANTHROPIC_/OPENAI_/GEMINI_API_KEY)
  CH02_LOG             event log path (default %[2]s)
`, p, defaultLogPath())
}

const systemPrompt = "You are a helpful assistant."

func main() {
	args := os.Args[1:]
	mode := ""
	if len(args) > 0 {
		mode = args[0]
	}

	reg := tools.NewRegistry()
	cfg, err := configFromEnv(reg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(2)
	}
	logPath := envOr("CH02_LOG", defaultLogPath())

	switch mode {
	case "--help", "-h", "help":
		usage(os.Stdout)

	case "render":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "usage: %s render LOG\n", progName())
			os.Exit(2)
		}
		body, err := llm.RenderOnly(args[1], cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "render:", err)
			os.Exit(1)
		}
		os.Stdout.Write(body)
		if len(body) > 0 && body[len(body)-1] != '\n' {
			fmt.Println()
		}

	case "dump":
		log, err := common.LoadLogFile(logPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "dump:", err)
			os.Exit(1)
		}
		if err := log.Write(os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "dump:", err)
			os.Exit(1)
		}

	case "chat":
		if runLoop(cfg, logPath, true, reg) {
			os.Exit(1)
		}

	case "":
		if runActorLoop(cfg, logPath, reg) {
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", mode)
		usage(os.Stderr)
		os.Exit(2)
	}
}

// cliHost implements common.Host for the CLI.
type cliHost struct{}

func (cliHost) Logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// ----------------------------------------------------------------
// Actor-based loop for grader mode. Reads stdin on its own goroutine
// and processes messages through the mailbox.
// ----------------------------------------------------------------

// stdinMsg is the unified JSON input format.
type stdinMsg struct {
	// Chapter 6 protocol.
	Kind *string `json:"kind"`
	Text *string `json:"text"`
	// Backward compat with ch1–5.
	User      *string `json:"user"`
	Ephemeral *string `json:"ephemeral"`
}

func runActorLoop(cfg common.Config, logPath string, reg *tools.Reg) (vendorFailed bool) {
	host := cliHost{}
	j := jobs.NewJobs(host)
	eng := llm.NewEngine(cfg, logPath, j, reg, host)
	actor := llm.NewActor(eng, host)

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

	if isTerminal(os.Stdin) {
		fmt.Fprintf(os.Stderr, "%[1]s: reading JSON-lines on stdin\n", progName())
		fmt.Fprintf(os.Stderr, "for an interactive chat run `%[1]s chat`; `%[1]s --help` lists every command\n", progName())
	}

	// Attach an observer that emits observations as JSON on stdout.
	actor.Attach(stdoutObserver(func(obs common.Observation) {
		emitLocked(observationJSON(obs))
	}))

	// Start the actor loop.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go actor.Run(ctx)

	// Read stdin on this goroutine.
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	hinted := false

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}

		var msg stdinMsg
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			emitLocked(map[string]string{"error": "bad input: " + err.Error()})
			if !hinted {
				hinted = true
				fmt.Fprintf(os.Stderr, "\n%[1]s: that line is not JSON.\n", progName())
			}
			continue
		}

		// Chapter 6 protocol: {"kind":"prompt","text":"..."}
		if msg.Kind != nil {
			switch *msg.Kind {
			case "prompt":
				if msg.Text == nil {
					emitLocked(map[string]string{"error": "prompt requires text"})
					continue
				}
				actor.Send(common.UserMessage{Text: *msg.Text})
				// Wait for turn to end.
				obs, err := actor.Wait(ctx, func(o common.Observation) bool {
					_, ok := o.(common.TurnEnded)
					return ok
				})
				if err != nil {
					vendorFailed = true
					emitLocked(map[string]string{"error": err.Error()})
					continue
				}
				ended := obs.(common.TurnEnded)
				if ended.Err != "" {
					vendorFailed = true
					emitLocked(map[string]string{"error": ended.Err})
				} else {
					emitLocked(map[string]string{"assistant": ended.Text})
				}
			case "hint":
				if msg.Text == nil {
					emitLocked(map[string]string{"error": "hint requires text"})
					continue
				}
				actor.Send(common.Hint{Text: *msg.Text})
			case "interrupt":
				actor.Send(common.Interrupt{})
			default:
				emitLocked(map[string]string{"error": "unknown kind: " + *msg.Kind})
			}
			continue
		}

		// Backward compat: {"ephemeral":"..."}
		if msg.Ephemeral != nil {
			if err := eng.Attach(*msg.Ephemeral); err != nil {
				emitLocked(map[string]string{"error": err.Error()})
				continue
			}
			emitLocked(map[string]string{"ack": "ephemeral"})
			continue
		}

		// Backward compat: {"user":"..."}
		if msg.User != nil {
			reply, err := eng.Ask(*msg.User)
			if err != nil {
				vendorFailed = true
				emitLocked(map[string]string{"error": err.Error()})
				continue
			}
			emitLocked(map[string]string{"assistant": reply})
			continue
		}

		emitLocked(map[string]string{"error": "no recognized field"})
	}

	_ = actor.Shutdown()
	emitLocked(map[string]any{"usage": eng.Ctx.Usage})
	out.Flush()
	return vendorFailed
}

// stdoutObserver adapts a func to common.Observer.
type stdoutObserver func(common.Observation)

func (f stdoutObserver) Observe(o common.Observation) { f(o) }

// observationJSON converts an Observation to a JSON-friendly map.
func observationJSON(obs common.Observation) map[string]any {
	switch v := obs.(type) {
	case common.PartDelta:
		m := map[string]any{"observation": "part_delta", "part_id": v.PartID, "chunk": v.Chunk}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		return m
	case common.PartFinal:
		m := map[string]any{"observation": "part_final", "part_id": v.PartID, "seq": v.Seq}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		if tp, ok := v.Part.(common.TextPart); ok {
			m["text"] = tp.Text
		}
		return m
	case common.StateChanged:
		m := map[string]any{"observation": "state_changed", "from": v.From.String(), "to": v.To.String()}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		return m
	case common.TurnEnded:
		m := map[string]any{"observation": "turn_ended", "text": v.Text}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		if v.Err != "" {
			m["error"] = v.Err
		}
		return m
	}
	return map[string]any{"observation": "unknown"}
}

// ----------------------------------------------------------------
// Interactive chat mode (same as ch05, uses synchronous Ask).
// ----------------------------------------------------------------

func runLoop(cfg common.Config, logPath string, interactive bool, reg *tools.Reg) (vendorFailed bool) {
	host := cliHost{}
	j := jobs.NewJobs(host)
	eng := llm.NewEngine(cfg, logPath, j, reg, host)
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	if interactive {
		fmt.Fprintf(os.Stderr, "%s — type a message, ctrl-D to exit\n", progName())
		fmt.Fprint(os.Stderr, "> ")
	}

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			if interactive {
				fmt.Fprint(os.Stderr, "> ")
			}
			continue
		}

		reply, err := eng.Ask(line)
		if err != nil {
			vendorFailed = true
			fmt.Fprintln(os.Stderr, "error:", err)
			fmt.Fprint(os.Stderr, "> ")
			continue
		}

		fmt.Fprintln(os.Stdout, reply)
		out.Flush()
		fmt.Fprint(os.Stderr, "> ")
	}

	_ = eng.Shutdown()
	out.Flush()
	return vendorFailed
}

func emit(out *bufio.Writer, v any) {
	b, _ := json.Marshal(v)
	out.Write(b)
	out.WriteByte('\n')
	out.Flush()
}

func configFromEnv(reg *tools.Reg) (common.Config, error) {
	vendor, err := parseVendor(envOr("LLM_VENDOR", "anthropic"))
	if err != nil {
		return common.Config{}, err
	}
	cfg := common.Config{
		Vendor:       vendor,
		Surface:      common.DefaultSurface(vendor),
		SystemPrompt: systemPrompt,
		MaxTokens:    1024,
		Tools:        reg.Declarations(),
	}
	switch vendor {
	case common.VendorAnthropic:
		cfg.Model = pick("LLM_MODEL", "ANTHROPIC_MODEL", "claude-sonnet-5")
		cfg.BaseURL = pick("LLM_BASE_URL", "ANTHROPIC_BASE_URL", "https://api.anthropic.com")
		cfg.APIKey = pick("LLM_API_KEY", "ANTHROPIC_API_KEY", "")
	case common.VendorOpenAI:
		cfg.Model = pick("LLM_MODEL", "OPENAI_MODEL", "gpt-5")
		cfg.BaseURL = pick("LLM_BASE_URL", "OPENAI_BASE_URL", "https://api.openai.com")
		cfg.APIKey = pick("LLM_API_KEY", "OPENAI_API_KEY", "")
	case common.VendorGemini:
		cfg.Model = pick("LLM_MODEL", "GEMINI_MODEL", "gemini-3.8-flash")
		cfg.BaseURL = pick("LLM_BASE_URL", "GEMINI_BASE_URL", "https://generativelanguage.googleapis.com")
		cfg.APIKey = pick("LLM_API_KEY", "GEMINI_API_KEY", "")
	}
	return cfg, nil
}

func parseVendor(s string) (common.Vendor, error) {
	switch common.NormalizeName(s) {
	case "anthropic", "claude":
		return common.VendorAnthropic, nil
	case "openai":
		return common.VendorOpenAI, nil
	case "gemini":
		return common.VendorGemini, nil
	}
	return 0, fmt.Errorf("unknown vendor %q (want anthropic, openai or gemini)", s)
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func pick(primary, secondary, def string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return envOr(secondary, def)
}
