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
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	agent "github.com/waywardgeek/ensemble/agent"
	"github.com/waywardgeek/ensemble/agent/internal/common"
	"github.com/waywardgeek/ensemble/agent/internal/jobs"
	"github.com/waywardgeek/ensemble/agent/internal/llm"
	"github.com/waywardgeek/ensemble/agent/internal/tools"
	"github.com/waywardgeek/ensemble/agent/internal/ws"
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
	port := ""
	guiDir := ""

	// Parse flags manually to keep backward compat with positional commands.
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
		if runActorLoop(cfg, logPath, reg, port, guiDir) {
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", mode)
		usage(os.Stderr)
		os.Exit(2)
	}
}

// cliHost implements common.Host for the CLI with three log destinations.
type cliHost struct {
	logger *agent.Logger
}

func newCLIHost() *cliHost {
	logger := agent.DefaultLogger()

	// API log: raw JSON wire traffic.
	if f, err := os.Create("api.log"); err == nil {
		logger.SetAPILog(f)
	}

	// Debug log: arbitrary text, also printed to stderr.
	if f, err := os.Create("debug.log"); err == nil {
		logger.SetDebugLog(f)
	}

	return &cliHost{logger: logger}
}

func (h *cliHost) Logf(format string, args ...any) {
	h.logger.Logf(format, args...)
}
func (h *cliHost) APILogf(format string, args ...any) {
	h.logger.APILogf(format, args...)
}
func (h *cliHost) Debugf(format string, args ...any) {
	h.logger.Debugf(format, args...)
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

func runActorLoop(cfg common.Config, logPath string, reg *tools.Reg, port string, guiDir string) (vendorFailed bool) {
	host := newCLIHost()
	j := jobs.NewJobs(host)

	// Set up skills.
	skillDir := envOr("EN_SKILLS_DIR", "skills")
	sr := common.NewSkillRegistry()
	vars := common.NewVarRegistry()

	// Built-in variable renderers.
	vars.Register("TOOLS", common.BuiltinToolsRenderer(sr, reg))
	vars.Register("SKILLS", common.BuiltinSkillsRenderer(sr))

	// Application-specific variable renderers (from environment).
	if cv := os.Getenv("EN_CUSTOM_VAR"); cv != "" {
		vars.Register("CUSTOM_VAR", func() string { return cv })
	}

	// Discover and load skills.
	if err := sr.DiscoverSkills(skillDir); err != nil {
		fmt.Fprintf(os.Stderr, "skills: %v\n", err)
	}

	// Load the primary skill (if any) to set the system prompt.
	primaryName := envOr("EN_PRIMARY_SKILL", "")
	if primaryName != "" {
		if err := sr.LoadInitial(primaryName, vars); err != nil {
			fmt.Fprintf(os.Stderr, "primary skill %q: %v\n", primaryName, err)
			os.Exit(1)
		}
		// Build the system prompt from initial skill bodies.
		bodies := sr.InitialBodies()
		if len(bodies) > 0 {
			cfg.SystemPrompt = strings.Join(bodies, "\n\n---\n\n")
		}
	}

	// Wire skill-based tool filtering and load_skill/unload_skill tools.
	reg.SetSkillRegistry(sr)

	// Rebuild tool declarations with skill filtering applied.
	cfg.Tools = reg.Declarations()

	eng := llm.NewEngine(cfg, logPath, j, reg, host)
	actor := llm.NewActor(eng, host)

	// Wire load_skill/unload_skill now that we have the event log.
	reg.WireSkills(sr, vars, eng.Log)
	// Re-derive declarations so load_skill and unload_skill appear.
	eng.Cfg.Tools = reg.Declarations()
	// When a skill is loaded dynamically, update the config's tool declarations.
	reg.SetOnToolsChanged(func() {
		eng.Cfg.Tools = reg.Declarations()
	})

	// Pause gate: shared between the actor and the WS hub.
	gate := common.NewPauseGate()
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

	if isTerminal(os.Stdin) {
		fmt.Fprintf(os.Stderr, "%[1]s: reading JSON-lines on stdin\n", progName())
		fmt.Fprintf(os.Stderr, "for an interactive chat run `%[1]s chat`; `%[1]s --help` lists every command\n", progName())
	}

	// Attach an observer that emits observations as JSON on stdout.
	actor.Attach(stdoutObserver(func(obs common.Observation) {
		emitLocked(observationJSON(obs))
	}))

	// Start the HTTP/WebSocket server if --port is set.
	if port != "" {
		settingsPath := filepath.Join(".", "settings.json")
		settingsStore := common.NewSettingsStore(settingsPath)

		hub := ws.NewHub(gate, func(msg common.Inbound) {
			actor.Send(msg)
		}, "gui.log", eng.Log, settingsStore)
		defer hub.Close()
		actor.Attach(hub)

		staticDir := guiDir
		if staticDir == "" {
			staticDir = "web/gui"
		}
		mux := http.NewServeMux()
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			fs.ServeHTTP(w, r)
		}))
		mux.HandleFunc("/ws", hub.ServeWS)
		srv := &http.Server{Addr: ":" + port, Handler: mux}
		go srv.ListenAndServe()
		defer srv.Close()
	}

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
				// The guidance below also exists in the TTY banner, which a
				// piped caller never sees. Someone feeding this binary the
				// wrong thing is exactly the person who needs to be told
				// what the right thing is, so say it here too.
				fmt.Fprintf(os.Stderr, "\n%[1]s: that line is not JSON.\n", progName())
				fmt.Fprintf(os.Stderr, "%[1]s reads a JSON-lines log on stdin, one message per line.\n", progName())
				fmt.Fprintf(os.Stderr, "for an interactive chat run `%[1]s chat`; `%[1]s --help` lists every command\n", progName())
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
		m := map[string]any{"observation": "part_delta", "part_id": v.PartID, "kind": v.Kind.String(), "chunk": v.Chunk}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		return m
	case common.PartFinal:
		m := map[string]any{"observation": "part_final", "part_id": v.PartID, "seq": v.Seq}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		switch p := v.Part.(type) {
		case common.TextPart:
			m["text"] = p.Text
		case common.ToolCallPart:
			// Finals now fire for tool calls and reasoning, not just text.
			// That content is what an Actions pane is made of, and the old
			// text-only path emitted none of it.
			m["tool"] = p.Name
			m["args"] = string(p.Args)
		case common.OpaquePart:
			m["opaque"] = true
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
	case common.ToolDispatched:
		m := map[string]any{"observation": "tool_dispatched", "call_id": v.CallID, "name": v.Name}
		if v.Input != nil {
			m["input"] = json.RawMessage(v.Input)
		}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		return m
	case common.ToolFinished:
		m := map[string]any{"observation": "tool_finished", "call_id": v.CallID, "result": v.Result, "is_error": v.IsError}
		if v.Agent != "" {
			m["agent"] = string(v.Agent)
		}
		return m
	}
	return map[string]any{"observation": "unknown"}
}

// ----------------------------------------------------------------
// Interactive chat mode (same as ch05, uses synchronous Ask).
// ----------------------------------------------------------------

func runLoop(cfg common.Config, logPath string, interactive bool, reg *tools.Reg) (vendorFailed bool) {
	host := newCLIHost()
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

		if _, err := eng.AskWatching(line, chatStream(out)); err != nil {
			vendorFailed = true
			fmt.Fprintln(os.Stderr, "error:", err)
			fmt.Fprint(os.Stderr, "> ")
			continue
		}

		// The reply already streamed to stdout through the callbacks above.
		// Printing it again here is the obvious mistake: it duplicates every
		// answer, and it looks correct until the first long one.
		fmt.Fprintln(out)
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
		MaxTokens:    16384,
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

	// Allow disabling streaming for testing.
	if os.Getenv("EN_DISABLE_STREAMING") == "1" {
		cfg.DisableStreaming = true
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

// chatStream prints a turn as it arrives.
//
// Reply text goes to STDOUT, because that is the answer and stdout is where
// an answer belongs. Reasoning and tool calls go to STDERR as commentary,
// dimmed and yellow. Piping the binary therefore still yields exactly what
// the assistant said and nothing else, while a human at a terminal sees the
// whole turn being built.
//
// Colour is chosen by whether STDERR is a terminal, not by a flag. Writing
// escape codes into a pipe is how a log file ends up full of \033[2m.
func chatStream(out *bufio.Writer) common.StreamCallbacks {
	const (
		reset  = "\033[0m"
		dim    = "\033[2m"
		yellow = "\033[33m"
	)
	color := isTerminal(os.Stderr)
	open := false

	paint := func(code string) {
		if color && !open {
			fmt.Fprint(os.Stderr, code)
			open = true
		}
	}
	clear := func() {
		if open {
			fmt.Fprint(os.Stderr, reset)
			open = false
		}
	}

	return common.StreamCallbacks{
		OnDelta: func(_ uint64, kind common.DeltaKind, chunk string) {
			switch kind {
			case common.DeltaText:
				clear()
				fmt.Fprint(out, chunk)
				// Flush per chunk. Without it the buffer holds the whole
				// answer and releases it in one lump at the end, which looks
				// exactly like streaming having no effect.
				out.Flush()
			case common.DeltaThinking:
				paint(dim)
				fmt.Fprint(os.Stderr, chunk)
			case common.DeltaToolCall:
				paint(yellow)
				fmt.Fprint(os.Stderr, chunk)
			}
		},
		OnPartFinal: func(_ uint64, _ common.Part) {
			// Only break the line if commentary was being written, so a
			// plain text answer does not collect blank lines after it.
			if open {
				clear()
				fmt.Fprintln(os.Stderr)
			}
		},
	}
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
