package main

// ch05 — the same agent as ch04, refactored into packages.
//
//	./ch05              grader mode: the Chapter 1 stdio protocol, unchanged
//	./ch05 chat         the interactive loop from Chapter 1
//	./ch05 render LOG   play LOG -> context -> render; print the request JSON
//	./ch05 dump         write the event log as JSON-lines
//	./ch05 --help       print the commands table

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/coding-agents-course/agent/internal/common"
	"github.com/waywardgeek/coding-agents-course/agent/internal/jobs"
	"github.com/waywardgeek/coding-agents-course/agent/internal/llm"
	"github.com/waywardgeek/coding-agents-course/agent/internal/tools"
)

const fallbackName = "ch05"

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
	fmt.Fprintf(w, `%[1]s — one log, three vendors, and a tool loop.

usage:
  %[1]s                grader mode: read a JSON-lines log on stdin
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
		if runLoop(cfg, logPath, false, reg) {
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", mode)
		usage(os.Stderr)
		os.Exit(2)
	}
}

// cliHost implements common.Host for the CLI — a simple stderr logger.
type cliHost struct{}

func (cliHost) Logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

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
	} else if isTerminal(os.Stdin) {
		fmt.Fprintf(os.Stderr, "%[1]s: reading a JSON-lines log on stdin, one object per line, e.g. {\"user\":\"Hi.\"}\n", progName())
		fmt.Fprintf(os.Stderr, "for an interactive chat run `%[1]s chat`; `%[1]s --help` lists every command\n", progName())
	}

	hinted := false

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			if interactive {
				fmt.Fprint(os.Stderr, "> ")
			}
			continue
		}

		prompt := line
		if !interactive {
			var msg struct {
				User      *string `json:"user"`
				Ephemeral *string `json:"ephemeral"`
			}
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				emit(out, map[string]string{"error": "bad input: " + err.Error()})
				if !hinted {
					hinted = true
					fmt.Fprintf(os.Stderr,
						"\n%[1]s: that line is not JSON.\n"+
							"This mode reads a JSON-lines log on stdin — one object per line, e.g.\n"+
							"    {\"user\":\"Hi.\"}\n"+
							"To type messages yourself, run `%[1]s chat`.\n"+
							"`%[1]s --help` lists every command.\n\n", progName())
				}
				continue
			}
			if msg.Ephemeral != nil {
				if err := eng.Attach(*msg.Ephemeral); err != nil {
					emit(out, map[string]string{"error": err.Error()})
					continue
				}
				emit(out, map[string]string{"ack": "ephemeral"})
				continue
			}
			if msg.User == nil {
				emit(out, map[string]string{"error": "no user field"})
				continue
			}
			prompt = *msg.User
		}

		reply, err := eng.Ask(prompt)
		if err != nil {
			vendorFailed = true
			if interactive {
				fmt.Fprintln(os.Stderr, "error:", err)
				fmt.Fprint(os.Stderr, "> ")
				continue
			}
			emit(out, map[string]string{"error": err.Error()})
			continue
		}

		if interactive {
			fmt.Fprintln(os.Stdout, reply)
			out.Flush()
			fmt.Fprint(os.Stderr, "> ")
			continue
		}
		emit(out, map[string]string{"assistant": reply})
	}

	_ = eng.Shutdown()

	if !interactive {
		emit(out, map[string]any{"usage": eng.Ctx.Usage})
	}
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
