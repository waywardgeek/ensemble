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
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	agent "github.com/waywardgeek/ensemble/agent"
	"github.com/waywardgeek/ensemble/agent/internal/common"
	"github.com/waywardgeek/ensemble/agent/internal/llm"
	"github.com/waywardgeek/ensemble/agent/internal/mcp"
	"github.com/waywardgeek/ensemble/agent/internal/settings"
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
                       /hint <text> steers a running turn, /interrupt stops it
                       --verbose also shows state changes and tool arguments
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
	// Rule 1: the save file is save.json in the working directory,
	// beside settings.json. --save PATH names a different file. The
	// flag changes WHERE, never WHETHER: this one path is both the
	// file loaded at start and the file written at exit. There is no
	// --load, because a flag that loads without saving (or saves
	// without loading) is how you lose an afternoon of conversation.
	savePath := "save.json"
	mcpPipe := false
	guiDebug := false
	// Exposes the browser's MCP server on a loopback TCP port, so an agent in
	// another process can see and drive this GUI. Off unless asked for: anyone
	// who can reach the port can click and type in the GUI.
	mcpPort := ""
	skillsDir := ""
	// On by default, like api.log and debug.log. Speech is the one channel you
	// cannot scroll back through, so it is the one that most needs a record.
	ttsLogPath := "tts.log"

	// Print the authorize URL instead of opening a browser. This is not a
	// test hook: an agent running over SSH, in a container, or on a headless
	// box has no browser to open, and the operator needs the URL to paste
	// into one somewhere else.
	printURL := false

	// Shows state transitions, streaming tool-argument fragments and opaque
	// parts in the terminal front end. Off by default because the interesting
	// output is the conversation, not the machinery around it.
	verbose := false

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
		case args[i] == "--save" && i+1 < len(args):
			savePath = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--save="):
			savePath = strings.TrimPrefix(args[i], "--save=")
		case args[i] == "--print-url":
			printURL = true
		case args[i] == "--verbose":
			verbose = true
		case args[i] == "--mcp-pipe":
			mcpPipe = true
		case args[i] == "--gui-debug":
			guiDebug = true
		case args[i] == "--mcp-port" && i+1 < len(args):
			mcpPort = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--mcp-port="):
			mcpPort = strings.TrimPrefix(args[i], "--mcp-port=")
		case args[i] == "--skills-dir" && i+1 < len(args):
			skillsDir = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--skills-dir="):
			skillsDir = strings.TrimPrefix(args[i], "--skills-dir=")
		case args[i] == "--tts-log" && i+1 < len(args):
			ttsLogPath = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--tts-log="):
			ttsLogPath = strings.TrimPrefix(args[i], "--tts-log=")
		default:
			filtered = append(filtered, args[i])
		}
	}
	args = filtered
	if mcpPort != "" && port == "" {
		fmt.Fprintln(os.Stderr, "--mcp-port relays to the GUI, so it needs --port as well")
		os.Exit(2)
	}
	if len(args) > 0 {
		mode = args[0]
	}

	cfg, err := configFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(2)
	}
	logPath := envOr("CH02_LOG", defaultLogPath())

	switch mode {
	case "--help", "-h", "help":
		usage(os.Stdout)

	case "auth":
		sub := ""
		if len(args) > 1 {
			sub = args[1]
		}
		if err := runAuth(sub, printURL); err != nil {
			fmt.Fprintln(os.Stderr, "auth:", err)
			os.Exit(2)
		}

	case "verify":
		sf, err := llm.Load(savePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load: %v\n", err)
			os.Exit(2)
		}
		// Rebuild from the log alone and compare against the saved
		// snapshot. A save with a snapshot and a trimmed log has
		// nothing to compare, so this only means something for a save
		// that still carries its whole history. verify is a debugging
		// aid and no part of the contract: nothing grades it.
		full := &llm.SaveFile{Log: sf.Log}
		rebuilt := full.Restore(func(err error) {
			fmt.Fprintf(os.Stderr, "rebuild: %v\n", err)
		})
		savedJSON, _ := json.MarshalIndent(sf.Context, "", "  ")
		rebuiltJSON, _ := json.MarshalIndent(rebuilt, "", "  ")
		if string(savedJSON) == string(rebuiltJSON) {
			fmt.Println("MATCH")
		} else {
			fmt.Println("MISMATCH")
			// Find first difference
			for i := 0; i < len(savedJSON) && i < len(rebuiltJSON); i++ {
				if savedJSON[i] != rebuiltJSON[i] {
					fmt.Fprintf(os.Stderr, "first diff at byte %d\n", i)
					break
				}
			}
			os.Exit(1)
		}

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
		log, err := llm.LoadLogFile(logPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "dump:", err)
			os.Exit(1)
		}
		if err := llm.Write(log, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "dump:", err)
			os.Exit(1)
		}

	case "chat":
		// Chat is the SAME actor loop the GUI drives, with a text front end
		// attached instead of a browser. It used to be a bare engine with no
		// actor, skills, recall or save file, which meant the terminal could
		// not reproduce a single GUI-reported bug. Pass --port as well to run
		// both front ends against one agent.
		if runActorLoop(cfg, logPath, port, guiDir, savePath, mcpPipe, guiDebug, skillsDir, ttsLogPath, mcpPort, true, verbose) {
			os.Exit(1)
		}

	case "":
		if runActorLoop(cfg, logPath, port, guiDir, savePath, mcpPipe, guiDebug, skillsDir, ttsLogPath, mcpPort, false, verbose) {
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", mode)
		usage(os.Stderr)
		os.Exit(2)
	}
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

func runActorLoop(cfg common.Config, logPath string, port string, guiDir string, savePath string, mcpPipe bool, guiDebug bool, skillsDir string, ttsLogPath string, mcpPort string, textMode, verboseText bool) (vendorFailed bool) {
	// One agent, built by the library's own constructor.
	//
	// Everything that follows used to be assembled here by hand: a second
	// logging host, a second job table, a second tool registry, a second
	// engine. None of it was deliberate. NewAgent had no caller, so when a
	// capability arrived it was easier to staple it onto the engine here than
	// to teach a constructor nobody was using.
	skillDir := envOr("EN_SKILLS_DIR", "skills")
	if skillsDir != "" {
		skillDir = skillsDir
	}
	primaryName := envOr("EN_PRIMARY_SKILL", "ensemble")

	a, err := agent.NewAgent(cfg, agent.AgentSpec{
		DataDir:  ".",
		SkillDir: skillDir,
		Skills:   []string{primaryName},
		LogPath:  logPath,
		SavePath: savePath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	host := a
	j := a.Jobs()
	reg := a.Registry()
	sr := a.Skills()
	vars := a.Vars()
	eng := a.Engine()

	// Application-specific variable renderers, from the environment. These are
	// registered after construction because only skills loaded later in the
	// session can reference them; the primary skill body cannot, since it was
	// rendered during NewAgent.
	if cv := os.Getenv("EN_CUSTOM_VAR"); cv != "" {
		vars.Register("CUSTOM_VAR", func() string { return cv })
	}

	attachCredentials(eng, cfg)
	// The engine's capabilities -- the cache lens, restored context, the
	// journal, settings, memory, memory bands and recall -- are built by
	// NewAgent now, inside this agent's own data directory.
	//
	// All of them used to be stapled onto the engine here, one assignment at
	// a time, over five chapters. That is precisely how the library's
	// constructor fell sixteen chapters behind without anyone noticing: a
	// capability added here cost one line, and a capability added to NewAgent
	// would have had to be added to a function with no callers.
	settingsStore := a.Settings()
	actor := a.Actor()
	defer a.Journal().Close()

	// Wire load_skill/unload_skill now that we have the event log.
	reg.WireSkills(sr, vars, eng.Log)
	// Re-derive declarations so load_skill and unload_skill appear. This is
	// the last time eng.Cfg.Tools is written once a session can start:
	// Chapter 15 rule 1 freezes the startup declarations. A skill loaded
	// later records ToolsChanged instead, and the renderer carries the
	// change in the dialog (rule 2). Chapter 10 used to rewrite
	// eng.Cfg.Tools here on every load, which changed the prefix and
	// missed the cache on every request after it.
	eng.Cfg.Tools = reg.Declarations()

	// Track MCP clients per skill for lifecycle management.
	skillMCPClients := make(map[string][]*mcp.Client)
	skillMCPToolNames := make(map[string][]string)

	// Skill-based MCP lifecycle: connect when a skill is loaded.
	reg.SetOnSkillMCPConnect(func(skill string, servers []common.MCPServerConfig) ([]string, error) {
		var allToolNames []string
		for _, srv := range servers {
			var t mcp.Transport
			var tErr error
			switch srv.Transport {
			case "stdio":
				t, tErr = mcp.NewStdioTransport(srv.Command, srv.Args, srv.Env)
			case "url":
				// Streamable HTTP. The Transport interface was written with
				// this in mind -- "a future fourth (URL/port-based) can be
				// added later" -- and MCPServerConfig has carried a URL field
				// since Chapter 12. This case is the whole host-side cost of
				// reaching a hosted MCP server.
				//
				// auth-env names an environment variable; the skill file never
				// holds the key itself. If the variable is unset we still
				// connect, because this endpoint also answers keyless requests
				// -- at a lower daily limit. Degrading to anonymous is safe;
				// doing it silently is not, so say so.
				var headers map[string]string
				if srv.AuthEnv != "" {
					if token := os.Getenv(srv.AuthEnv); token != "" {
						headers = map[string]string{
							"Authorization": "Bearer " + token,
						}
					} else {
						host.Logf("skill %s MCP %s: $%s is unset; connecting without a key",
							skill, srv.Name, srv.AuthEnv)
					}
				}
				t, tErr = mcp.NewHTTPTransport(srv.URL, headers)
			default:
				host.Logf("skill %s: unsupported MCP transport %q, skipping", skill, srv.Transport)
				continue
			}
			if tErr != nil {
				return nil, fmt.Errorf("skill %s MCP %s: %w", skill, srv.Name, tErr)
			}

			client := mcp.NewClient(t)
			if err := client.Initialize(context.Background()); err != nil {
				client.Close()
				return nil, fmt.Errorf("skill %s MCP %s init: %w", skill, srv.Name, err)
			}

			mcpTools, err := client.ListTools(context.Background())
			if err != nil {
				client.Close()
				return nil, fmt.Errorf("skill %s MCP %s list: %w", skill, srv.Name, err)
			}

			bridged := mcp.Bridge(client, mcpTools)
			for _, bt := range bridged {
				reg.RegisterTool(bt)
				allToolNames = append(allToolNames, bt.Name)
			}

			client.SetReverseHandler(func(name string, args json.RawMessage) (string, error) {
				tool, lErr := reg.Lookup(name)
				if lErr != nil {
					return "", lErr
				}
				return tool.Run(&common.Call{Agent: host, Jobs: j}, args)
			})

			skillMCPClients[skill] = append(skillMCPClients[skill], client)
		}
		skillMCPToolNames[skill] = allToolNames
		return allToolNames, nil
	})

	// Skill-based MCP lifecycle: disconnect when a skill is unloaded.
	reg.SetOnSkillMCPDisconnect(func(skill string) error {
		// Remove bridged tools from registry.
		for _, name := range skillMCPToolNames[skill] {
			reg.RemoveTool(name)
		}
		delete(skillMCPToolNames, skill)

		// Close MCP clients.
		for _, client := range skillMCPClients[skill] {
			client.Close()
		}
		delete(skillMCPClients, skill)
		return nil
	})

	// --gui-debug: auto-load the gui-debug skill on startup.
	if guiDebug {
		loadTool, lErr := reg.Lookup("load_skill")
		if lErr != nil {
			fmt.Fprintf(os.Stderr, "gui-debug: load_skill not found: %v\n", lErr)
			return true
		}
		args, _ := json.Marshal(map[string]string{"name": "gui-debug"})
		// A real Call, so the skill's SkillLoaded event goes through
		// Record like every other event (ch15: the reducer turns it into
		// the Skill entry). Its ToolsChanged is dropped: this is startup,
		// the prefix is not frozen yet, and the declarations are simply
		// re-derived below.
		c := &common.Call{Agent: host, Jobs: j}
		if _, lErr = loadTool.Run(c, args); lErr != nil {
			fmt.Fprintf(os.Stderr, "gui-debug: %v\n", lErr)
			return true
		}
		for _, ev := range c.Events {
			if ev.Type == common.ToolsChanged {
				continue
			}
			if rErr := eng.Record(ev); rErr != nil {
				fmt.Fprintf(os.Stderr, "gui-debug: %v\n", rErr)
				return true
			}
		}
		eng.Cfg.Tools = reg.Declarations()
	}

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

	if textMode {
		if isTerminal(os.Stdin) {
			fmt.Fprintf(os.Stderr, "%[1]s: interactive chat. type a message and press enter.\n", progName())
			fmt.Fprintf(os.Stderr, "/hint <text> steers the turn in flight, /interrupt stops it, /quit exits; `%[1]s --help` lists every command\n", progName())
		}
		// The terminal front end is an Observer, exactly like the GUI. Both
		// watch the same actor, so the CLI reproduces agent bugs without a
		// browser, and a bug that appears in only one of them is a front-end
		// bug.
		actor.Attach(newTextObserver(os.Stdout, os.Stderr, verboseText))
	} else {
		if isTerminal(os.Stdin) {
			fmt.Fprintf(os.Stderr, "%[1]s: reading JSON-lines on stdin\n", progName())
			fmt.Fprintf(os.Stderr, "for an interactive chat run `%[1]s chat`; `%[1]s --help` lists every command\n", progName())
		}

		// Attach an observer that emits observations as JSON on stdout.
		actor.Attach(stdoutObserver(func(obs common.Observation) {
			emitLocked(observationJSON(obs))
		}))
	}

	// Start the HTTP/WebSocket server if --port is set.
	if port != "" {
		// The same store the context policy reads, so a target changed in
		// the Context Management tab applies from the next request.
		//
		// Usage comes from the engine, which spends the tokens and is the
		// only object that knows which model spent them. It used to come
		// from host, back when the counter was embedded on the agent
		// because the agent was the thing everything could reach.
		//
		// Note what this line cannot protect against. A usage source handed
		// over once at construction is resolved now and never again, so
		// moving the counter elsewhere silently aims the meter at an object
		// nobody writes to: nothing fails to compile and the meter reads
		// zero. A back-pointer is re-resolved at the moment of asking and
		// cannot go stale that way.
		hub := ws.NewHub(gate, func(msg common.Inbound) {
			actor.Send(msg)
		}, "gui.log", eng.Log, settingsStore, eng.Usage())
		// view_gui: the agent looks at its own GUI when it chooses to. Registered
		// here because it needs the hub, and before the actor starts, so it is in
		// the startup declarations like every builtin (the --gui-debug load above
		// re-derives them at this stage for the same reason).
		reg.RegisterInitial(hub.ViewGUITool())
		eng.Cfg.Tools = reg.Declarations()
		// The engine's config is the authority on which model is serving
		// turns: it is validated at startup and refuses a model it does not
		// know. The settings store can be empty, and pricing an empty name
		// silently reports a paid model as free.
		hub.Model = func() string { return eng.Cfg.Model }
		defer hub.Close()
		hub.SetTTSLog(ttsLogPath)
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

		if mcpPort != "" {
			// Loopback only, unlike the GUI server: this port takes clicks
			// and keystrokes, not just page loads.
			ln, lErr := net.Listen("tcp", "127.0.0.1:"+mcpPort)
			if lErr != nil {
				fmt.Fprintf(os.Stderr, "mcp-port: %v\n", lErr)
				os.Exit(2)
			}
			defer ln.Close()
			go hub.ServeMCP(ln)
			fmt.Fprintf(os.Stderr, "GUI MCP server on %s (connect with mcp-connect)\n", ln.Addr())
		}
	}

	// Start the actor loop.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go actor.Run(ctx)

	// MCP pipe mode: connect the MCP client to stdin/stdout.
	// The grader (or another MCP server) drives the MCP protocol externally.
	if mcpPipe {
		t := mcp.NewRawTransport(os.Stdin, os.Stdout)
		client := mcp.NewClient(t)
		defer client.Close()

		if err := client.Initialize(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "mcp init: %v\n", err)
			os.Exit(1)
		}
		mcpTools, err := client.ListTools(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mcp tools/list: %v\n", err)
			os.Exit(1)
		}

		// Bridge and register MCP tools.
		for _, bt := range mcp.Bridge(client, mcpTools) {
			reg.RegisterTool(bt)
		}

		// Set up reverse handler.
		client.SetReverseHandler(func(name string, args json.RawMessage) (string, error) {
			tool, lookupErr := reg.Lookup(name)
			if lookupErr != nil {
				return "", lookupErr
			}
			c := &common.Call{Agent: host, Jobs: j}
			return tool.Run(c, args)
		})

		// Update tool declarations.
		eng.Cfg.Tools = reg.Declarations()

		// Block until the MCP transport closes (EOF from the other side).
		<-client.Done()
		cancel()
		_ = actor.Shutdown()
		return false
	}

	// Read stdin on its own goroutine so the main goroutine can wait on
	// either end-of-input or a signal. A blocking Scan cannot be cancelled,
	// so on Ctrl-C we abandon this goroutine and leave through the normal
	// path below — which is the whole point, because that path holds the save.
	stdinDone := make(chan struct{})
	go func() {
		defer close(stdinDone)

		in := bufio.NewScanner(os.Stdin)
		in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
		hinted := false

		for {
			if textMode && isTerminal(os.Stdin) {
				fmt.Fprint(os.Stdout, "\n> ")
			}
			if !in.Scan() {
				break
			}
			line := strings.TrimSpace(in.Text())
			if line == "" {
				continue
			}

			var msg stdinMsg
			if textMode {
				// A terminal user types prose, not protocol. Slash commands
				// carry the two things prose cannot express: steering a turn
				// that is already running, and stopping it.
				parsed, quit, perr := parseTextLine(line)
				if quit {
					return
				}
				if perr != nil {
					fmt.Fprintf(os.Stdout, "%s\n", perr)
					continue
				}
				msg = parsed
			} else if err := json.Unmarshal([]byte(line), &msg); err != nil {
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
					reply, err := actor.AskContext(ctx, *msg.Text)
					if err != nil {
						vendorFailed = true
						if !textMode {
							emitLocked(map[string]string{"error": err.Error()})
						}
					} else if !textMode {
						// The text observer already streamed it; never echo twice.
						emitLocked(map[string]string{"assistant": reply})
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
				if err := actor.AttachContext(ctx, *msg.Ephemeral); err != nil {
					emitLocked(map[string]string{"error": err.Error()})
					continue
				}
				emitLocked(map[string]string{"ack": "ephemeral"})
				continue
			}

			// Backward compat: {"user":"..."}
			if msg.User != nil {
				reply, err := actor.AskContext(ctx, *msg.User)
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
	}()

	// Ctrl-C and SIGTERM have to reach the same save that a clean exit runs.
	// Without this the process dies with the conversation live only in the
	// journal, and save.json is never written at all.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stdinDone:
	case sig := <-sigCh:
		fmt.Fprintf(os.Stderr, "\n%v received, saving conversation...\n", sig)
	}

	_ = actor.Shutdown()

	// Save state if --save was given.
	// Save at exit (rule 6). Every event the engine records is folded
	// into Ctx as it is appended, so the last event in the log is
	// exactly the last event inside the snapshot: that is the anchor.
	// An empty log means nothing has been folded in, so the anchor is
	// whatever it was when we loaded — which ResetSeq left one below
	// the next Seq.
	asOf := common.Seq(0)
	if n := len(eng.Log.Events); n > 0 {
		asOf = eng.Log.Events[n-1].Seq
	} else {
		asOf = llm.NextSeq(eng.Log) - 1
	}
	// Chapter 15 rule 9: snapshot first, then empty the journal, and only
	// if the snapshot landed. The other order loses the tail to a crash
	// between the two; a failed save that still emptied the journal loses
	// it outright.
	if err := llm.SaveRetaining(savePath, asOf, eng.Ctx, eng.Log, eng.Cfg, settingsStore.Get().LogRetention); err != nil {
		fmt.Fprintf(os.Stderr, "save: %v\n", err)
	} else if err := a.Journal().Reset(); err != nil {
		fmt.Fprintf(os.Stderr, "save: %v\n", err)
	}

	if textMode {
		// Usage is commentary, not the answer. Printing it on stdout would put
		// a JSON line in the middle of piped reply text.
		if b, err := json.Marshal(map[string]any{"usage": eng.Ctx.Usage}); err == nil {
			fmt.Fprintf(os.Stderr, "%s\n", b)
		}
	} else {
		emitLocked(map[string]any{"usage": eng.Ctx.Usage})
	}
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

func emit(out *bufio.Writer, v any) {
	b, _ := json.Marshal(v)
	out.Write(b)
	out.WriteByte('\n')
	out.Flush()
}

func configFromEnv() (common.Config, error) {
	// Settle the model before the vendor, because the vendor is a property of
	// the model and not an independent choice. Resolving the vendor first and
	// then choosing a model inside it means a model named anywhere other than
	// LLM_MODEL reaches whichever vendor happened to be the default. That does
	// not fail: an OpenAI id sent to Anthropic comes back answered by Sonnet,
	// so the session looks healthy and is simply the wrong model. The GUI
	// never hit this because it sends the model again after connecting; the
	// CLI has nothing to send it, so it ran the default every time.
	named := os.Getenv("LLM_MODEL")
	if named == "" {
		named = homeSetting("LLM_MODEL")
	}
	if named == "" {
		named = localSettingModel()
	}

	vendorNamed := os.Getenv("LLM_VENDOR")
	if vendorNamed == "" {
		vendorNamed = homeSetting("LLM_VENDOR")
	}

	var vendor common.Vendor
	switch {
	case vendorNamed != "":
		// An explicit vendor is an instruction and outranks the table, which
		// is what lets a local OpenAI-compatible endpoint serve a model id
		// the table has never heard of.
		v, err := parseVendor(vendorNamed)
		if err != nil {
			return common.Config{}, err
		}
		vendor = v
	default:
		vendor = common.VendorAnthropic
		if feat, ok := common.LookupModel(named); ok && named != "" {
			// Only a model the table knows may choose its own vendor. An
			// unknown id keeps the old default rather than erroring, because
			// the graders drive custom ids at fake endpoints and a hard
			// failure here would stop them dead.
			vendor = feat.Vendor
		}
	}
	cfg := common.Config{
		Vendor:       vendor,
		SystemPrompt: systemPrompt,
		MaxTokens:    16384,
	}
	cfg.Endpoints = common.ResolveEndpoints(pick)
	active := cfg.Endpoints[vendor]
	cfg.BaseURL, cfg.APIKey = active.BaseURL, active.APIKey
	switch vendor {
	case common.VendorAnthropic:
		cfg.Model = pick("LLM_MODEL", "ANTHROPIC_MODEL", "claude-sonnet-5")
	case common.VendorOpenAI:
		cfg.Model = pick("LLM_MODEL", "OPENAI_MODEL", "gpt-6.1-sol")
	case common.VendorGemini:
		cfg.Model = pick("LLM_MODEL", "GEMINI_MODEL", "gemini-3.8-flash")
	}

	// After the model is known, never before: the surface is a property of
	// the model, and resolving it from the vendor alone would put every
	// OpenAI model on whichever endpoint the newest one happens to use.
	// A model named outside the environment, which is to say the one the GUI
	// last wrote to settings.json, wins over the per-vendor default picked
	// above. Without this the CLI silently ignores the model on display.
	if named != "" {
		cfg.Model = named
	}

	cfg.Surface = common.SurfaceForModel(cfg.Model, vendor)

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

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	if v := homeSetting(k); v != "" {
		return v
	}
	return def
}

func pick(primary, secondary, def string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	if v := os.Getenv(secondary); v != "" {
		return v
	}
	if v := homeSetting(primary); v != "" {
		return v
	}
	if v := homeSetting(secondary); v != "" {
		return v
	}
	return def
}

// homeSetting reads one value from ~/.en/settings.json.
//
// This re-reads the file on every lookup rather than caching it in a package
// variable, which is deliberate. A package-level var holding process-wide
// configuration is the mutable global the architecture forbids: it makes tests
// order-dependent, it cannot be overridden per agent, and it is invisible to
// the composition root. The file holds a handful of keys and is consulted only
// while assembling the startup config, so the cost of not caching it is a few
// microseconds once. That is a good trade for keeping the global out.
// localSettingModel reports the model recorded in the settings file the GUI
// writes beside the agent. It reuses the settings reader rather than parsing
// the file again so that the shape of settings.json is known in one place.
// An absent or unreadable file yields the empty string, which simply means
// "nothing chosen" and leaves the per-vendor default in force.
func localSettingModel() string {
	return settings.NewSettingsStore(filepath.Join(".", "settings.json")).Get().Model
}

func homeSetting(key string) string {
	if v, ok := loadHomeSettings()[strings.ToLower(key)]; ok {
		return v
	}
	return ""
}

func loadHomeSettings() map[string]string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(home, ".en", "settings.json"))
	if err != nil {
		return nil
	}
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Fprintf(os.Stderr, "warning: ~/.en/settings.json: %v\n", err)
		return nil
	}
	return raw
}
