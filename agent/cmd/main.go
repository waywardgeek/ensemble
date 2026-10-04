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
	"github.com/waywardgeek/ensemble/agent/internal/cachelens"
	"github.com/waywardgeek/ensemble/agent/internal/common"
	"github.com/waywardgeek/ensemble/agent/internal/jobs"
	"github.com/waywardgeek/ensemble/agent/internal/llm"
	"github.com/waywardgeek/ensemble/agent/internal/mcp"
	"github.com/waywardgeek/ensemble/agent/internal/recall"
	"github.com/waywardgeek/ensemble/agent/internal/settings"
	"github.com/waywardgeek/ensemble/agent/internal/skills"
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
		if runActorLoop(cfg, logPath, reg, port, guiDir, savePath, mcpPipe, guiDebug, skillsDir, ttsLogPath, mcpPort, true, verbose) {
			os.Exit(1)
		}

	case "":
		if runActorLoop(cfg, logPath, reg, port, guiDir, savePath, mcpPipe, guiDebug, skillsDir, ttsLogPath, mcpPort, false, verbose) {
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
	// UsageCounter holds this run's token totals. cliHost is the composition
	// root for the server, so its lifespan is the process — which is exactly
	// the session the header reports on.
	common.UsageCounter

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

func runActorLoop(cfg common.Config, logPath string, reg *tools.Reg, port string, guiDir string, savePath string, mcpPipe bool, guiDebug bool, skillsDir string, ttsLogPath string, mcpPort string, textMode, verboseText bool) (vendorFailed bool) {
	host := newCLIHost()
	j := jobs.NewJobs(host)

	// Set up skills.
	skillDir := envOr("EN_SKILLS_DIR", "skills")
	if skillsDir != "" {
		skillDir = skillsDir
	}
	sr := skills.NewSkillRegistry()
	vars := skills.NewVarRegistry()

	// Built-in variable renderers.
	vars.Register("TOOLS", skills.BuiltinToolsRenderer(sr, reg))
	vars.Register("SKILLS", skills.BuiltinSkillsRenderer(sr))

	// Application-specific variable renderers (from environment).
	if cv := os.Getenv("EN_CUSTOM_VAR"); cv != "" {
		vars.Register("CUSTOM_VAR", func() string { return cv })
	}

	// Discover and load skills.
	if err := sr.DiscoverSkills(skillDir); err != nil {
		fmt.Fprintf(os.Stderr, "skills: %v\n", err)
	}

	// Load the primary skill to set the system prompt. Defaulting to "ensemble"
	// means a bare `./ensemble --port 8084` gets a full toolset: the tool filter
	// enables only what a loaded skill declares, so an agent with no skill
	// loaded would be left with just load_skill and unload_skill. This is a
	// warning rather than fatal because the agent is also run without any
	// skills directory at all, where no filtering is the correct behaviour.
	primaryName := envOr("EN_PRIMARY_SKILL", "ensemble")
	if err := sr.LoadInitial(primaryName, vars); err != nil {
		fmt.Fprintf(os.Stderr, "primary skill %q: %v\n", primaryName, err)
	}
	// Build the system prompt from initial skill bodies.
	bodies := sr.InitialBodies()
	if len(bodies) > 0 {
		cfg.SystemPrompt = strings.Join(bodies, "\n\n---\n\n")
	}

	// Wire skill-based tool filtering and load_skill/unload_skill tools.
	reg.SetSkillRegistry(sr)

	// keep_tool_results means nothing on a model whose features row does
	// not enable per-round-trip stubbing (ch15 rule 6), so it is not
	// declared there: a tool that silently does nothing is a lie in the
	// prompt.
	//
	// This is the same call the actor makes on a model switch, so startup
	// and switch cannot drift apart: one function decides, in both cases.
	reg.SyncModelGatedTools(cfg.Model)

	// Rebuild tool declarations with skill filtering applied.
	cfg.Tools = reg.Declarations()

	eng := llm.NewEngine(cfg, logPath, j, reg, host)
	attachCredentials(eng, cfg)

	// Cache analysis is on by default, like api.log and debug.log.
	//
	// A cache miss is the one failure in this program that produces a correct
	// answer, so nothing else will ever report it: no error, no exception, no
	// failing test. It shows up only on the invoice, thirty days later,
	// aggregated past the point where you could tell which change caused it. A
	// failure with no natural signal needs a manufactured one, and a signal
	// that has to be switched on is off on the day you need it.
	//
	// The captures land in the working directory beside the other two logs.
	eng.Cache = cachelens.New(".", func(s string) { host.Debugf("%s", s) })

	// Load at start (rule 2). No flag decides this; the files' existence
	// does. Chapter 15 adds the journal: every event recorded since the
	// last snapshot, so a session that was killed resumes where it died.
	// Recover assembles snapshot + log + journal tail; Restore applies the
	// tail, skipping (and logging) any event it cannot apply.
	sf, err := llm.Recover(savePath, func(err error) { host.Logf("load: %v", err) })
	if err != nil {
		// A file that exists but does not parse is fatal. We must
		// NOT start fresh over it: the agent saves at exit, and a
		// fresh start would write an empty conversation over the
		// user's history. Name the file, touch none of its bytes,
		// and let a human look at it.
		fmt.Fprintf(os.Stderr, "agent: cannot load save file %s: %v\n", savePath, err)
		os.Exit(1)
	}
	if len(sf.Log) > 0 || sf.Context != nil {
		// Rule 3: install the snapshot, then apply only the tail above
		// its anchor. Restore handles the null-snapshot case too.
		eng.Ctx = sf.Restore(func(err error) { host.Logf("load: %v", err) })
		// The loaded log is kept whole: rule 6 saves it back plus
		// whatever this session adds. Restoring the context consumed
		// the tail; it did not consume the history.
		eng.Log.Events = sf.Log
		// Rule 5: numbering continues past both the anchor and the log,
		// which is why NextSeq looks at both.
		llm.ResetSeq(eng.Log, sf.NextSeq())
		// Rule 7 is enforced by what is ABSENT here: nothing copies
		// sf.Config back into cfg. The save records what shaped the
		// wire so a human can read it; the running agent's own model,
		// vendor, prompt and tools win. The ch2 context is
		// vendor-independent, so a conversation saved against one
		// vendor resumes against another.
	}
	journal, err := llm.OpenJournal(llm.JournalPath(savePath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent: %v\n", err)
		os.Exit(1)
	}
	defer journal.Close()
	eng.Journal = journal

	// Chapter 15 rule 10: the context target comes from settings.json,
	// read whether or not the GUI is up, and re-read on every request so a
	// change applies to the next cut and never to a recorded one.
	settingsStore := settings.NewSettingsStore(filepath.Join(".", "settings.json"))
	eng.Target = func() int { return settingsStore.Get().ContextTarget }
	eng.ToolRoundLimit = func() int { return settingsStore.Get().MaxToolRounds }

	// Chapter 16. The memory directory sits beside the save file, because
	// it is the same kind of thing: the part of this agent that outlives
	// the process. Bands are re-read on every check for the same reason
	// the context target is, so switching a band off in the GUI takes
	// effect on the next turn rather than on the next launch.
	eng.Memory = llm.NewStore(filepath.Join(".", "memory"))
	eng.Bands = func() common.BandConfig { return settingsStore.Get().Memory.Normalized() }

	// Chapter 17. Recall is assembled here, at the top, because this is the
	// only place in the program that is allowed to know about both halves.
	//
	// internal/recall knows how to score text and nothing about models.
	// internal/llm knows how to call a model and nothing about scoring. They
	// do not import each other and could not — they are both spokes, and a
	// spoke importing a spoke is how a star quietly becomes a graph. The two
	// are joined by two interfaces that live in the hub: the engine accepts a
	// common.Recaller, and the recaller accepts a common.SnippetJudge. Main
	// is where the concrete types on either side of those interfaces are
	// finally allowed to meet.
	//
	// The judge is a plain function call that happens to be evaluated by a
	// language model. It is not a sub-agent: it has no dialogue, no tools, no
	// memory of the last time it was asked, and no ability to do anything
	// except return text. Every call starts from nothing.
	recaller := recall.New(
		recall.DefaultSources(".", skillDir),
		llm.NewJudge(eng),
		recall.DefaultConfig(),
	)
	if recaller.Indexed() > 0 {
		// Left nil when there is nothing archived, which is the state of
		// every agent on its first run. Nil is not a degraded mode to be
		// apologised for: with an empty archive, every retrieval would score
		// nothing and the only observable effect of wiring it up would be a
		// judge call per turn that can only ever answer NONE.
		eng.Recall = recaller
	}

	// Populate the bands from what is on disk before the first turn, so a
	// restart comes back with the same memory it went down with. A failure
	// here is worth saying out loud and not worth dying over: an agent
	// with no memory can still work, it just cannot remember having done so.
	if err := eng.SyncBands("startup"); err != nil {
		host.Logf("memory: could not load bands at startup: %v", err)
	}

	actor := llm.NewActor(eng, host)

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
				return tool.Run(&common.Call{Host: host, Jobs: j}, args)
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
		c := &common.Call{Host: host, Jobs: j}
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

		hub := ws.NewHub(gate, func(msg common.Inbound) {
			actor.Send(msg)
		}, "gui.log", eng.Log, settingsStore, host)
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
			c := &common.Call{Host: host, Jobs: j}
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
	} else if err := journal.Reset(); err != nil {
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

func configFromEnv(reg *tools.Reg) (common.Config, error) {
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
		Tools:        reg.Declarations(),
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
