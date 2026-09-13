package main

// ch02 — one log, three vendors.
//
//	./ch02              grader mode: the Chapter 1 stdio protocol, unchanged
//	./ch02 chat         the interactive loop from Chapter 1
//	./ch02 render LOG   play LOG -> context -> render; print the request JSON
//	./ch02 dump         write the event log as JSON-lines
//	./ch02 --help       print the commands table

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// fallbackName is used only when os.Args[0] is empty or degenerate, which
// happens when a process is exec'd with an empty argv. It is not the normal
// path and never the one a student sees.
const fallbackName = "ch02"

// progName is the name this binary was invoked as. Deriving it once means the
// usage text, the `render` usage line and the default log name all agree with
// whatever the executable is actually called — so a binary built from
// solutions/ch03 calls itself ch03 and writes ch03.log, instead of claiming to
// be ch02 and scribbling into ch02.log.
func progName() string {
	base := filepath.Base(os.Args[0])
	if base == "." || base == string(os.PathSeparator) || base == "" {
		return fallbackName
	}
	return base
}

// defaultLogPath is the log used when CH02_LOG is unset. The environment
// variable keeps its name: students who passed Chapter 2 read CH02_LOG, and
// renaming it would break them for no gain. Only the DEFAULT changes.
func defaultLogPath() string { return progName() + ".log" }

// usage prints the commands table. It goes to stdout when the user asked for
// it and to stderr when they got here by making a mistake, which is the
// ordinary Unix split: asked-for output is data, unasked-for output is
// diagnostics.
func usage(w io.Writer) {
	p := progName()
	fmt.Fprintf(w, `%[1]s — one log, three vendors.

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

// The system prompt is a CONSTANT, computed nowhere and stored nowhere. For
// this chapter a constant is a perfectly good renderer. The rule is only about
// where it comes from: it is an output, not a value someone appends to.
const systemPrompt = "You are a helpful assistant."

func main() {
	args := os.Args[1:]
	mode := ""
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
		// Asked for, so it is data: stdout, exit 0. A --help that exits
		// nonzero cannot be piped into a pager without the shell complaining.
		usage(os.Stdout)

	case "render":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "usage: %s render LOG\n", progName())
			os.Exit(2)
		}
		// `render` takes NO FLAGS. Vendor and model come from the
		// environment, exactly as in grader mode. The moment rendering accepts
		// --model, byte-identity becomes a property of how you invoked the
		// command rather than of the log.
		body, err := RenderOnly(args[1], cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "render:", err)
			os.Exit(1)
		}
		os.Stdout.Write(body)
		if len(body) > 0 && body[len(body)-1] != '\n' {
			fmt.Println()
		}

	case "dump":
		log, err := LoadLogFile(logPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "dump:", err)
			os.Exit(1)
		}
		if err := log.Write(os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "dump:", err)
			os.Exit(1)
		}

	case "chat":
		if runLoop(cfg, logPath, true) {
			os.Exit(1)
		}

	case "":
		if runLoop(cfg, logPath, false) {
			os.Exit(1)
		}

	default:
		// Naming the bad token and then showing what the valid ones ARE is the
		// whole difference between a diagnostic and a shrug.
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", mode)
		usage(os.Stderr)
		os.Exit(2)
	}
}

// runLoop is both the interactive REPL and the grader protocol. They differ
// only in how a line is read and how a reply is printed, which is the honest
// amount of difference between them.
//
// runLoop reports whether any vendor call failed during the session.
//
// The failure is also printed, as a JSON error line, and printing feels like
// reporting. It is not. Nothing outside this process reads that line: a shell,
// a CI job and scripts/live.sh all ask the same question, and the only answer
// they get is the exit status. A session that could not reach the vendor and
// still exits 0 is a green dashboard — it reports success it did not have.
func runLoop(cfg Config, logPath string, interactive bool) (vendorFailed bool) {
	eng := NewEngine(cfg, logPath)
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	if interactive {
		fmt.Fprintf(os.Stderr, "%s — type a message, ctrl-D to exit\n", progName())
		fmt.Fprint(os.Stderr, "> ")
	} else if isTerminal(os.Stdin) {
		// A human typed the bare command and is now looking at a blank line,
		// wondering whether it hung. Say what this mode is BEFORE they type
		// prose at it, not after. Grader stdin is a pipe, so this never fires
		// under grading.
		fmt.Fprintf(os.Stderr, "%[1]s: reading a JSON-lines log on stdin, one object per line, e.g. {\"user\":\"Hi.\"}\n", progName())
		fmt.Fprintf(os.Stderr, "for an interactive chat run `%[1]s chat`; `%[1]s --help` lists every command\n", progName())
	}

	// hinted keeps the long explanation to once per session: a malformed
	// 10,000-line file should not print 10,000 identical paragraphs.
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
				// The stdout line is the machine protocol and is unchanged,
				// byte for byte. The EXPLANATION goes to stderr, where a human
				// reading their terminal sees it and a program parsing stdout
				// does not. "invalid character 'H'" is accurate and useless;
				// it names the symptom and not the mistake.
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
			// A directive is acknowledged so that one unhandled directive
			// fails one check with one cause, instead of four checks with
			// four mysteries.
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

	_ = eng.Save()

	if !interactive {
		// Chapter 1's closing line, unchanged in shape. The extra categories
		// are additive: input and output still mean what they meant.
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

func configFromEnv() (Config, error) {
	vendor, err := parseVendor(envOr("LLM_VENDOR", "anthropic"))
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Vendor:       vendor,
		Surface:      DefaultSurface(vendor),
		SystemPrompt: systemPrompt,
		MaxTokens:    1024,
	}
	switch vendor {
	case VendorAnthropic:
		cfg.Model = pick("LLM_MODEL", "ANTHROPIC_MODEL", "claude-sonnet-5")
		cfg.BaseURL = pick("LLM_BASE_URL", "ANTHROPIC_BASE_URL", "https://api.anthropic.com")
		cfg.APIKey = pick("LLM_API_KEY", "ANTHROPIC_API_KEY", "")
	case VendorOpenAI:
		cfg.Model = pick("LLM_MODEL", "OPENAI_MODEL", "gpt-5")
		cfg.BaseURL = pick("LLM_BASE_URL", "OPENAI_BASE_URL", "https://api.openai.com")
		cfg.APIKey = pick("LLM_API_KEY", "OPENAI_API_KEY", "")
	case VendorGemini:
		cfg.Model = pick("LLM_MODEL", "GEMINI_MODEL", "gemini-3.8-flash")
		cfg.BaseURL = pick("LLM_BASE_URL", "GEMINI_BASE_URL", "https://generativelanguage.googleapis.com")
		cfg.APIKey = pick("LLM_API_KEY", "GEMINI_API_KEY", "")
	}
	return cfg, nil
}

func parseVendor(s string) (Vendor, error) {
	switch normalizeName(s) {
	case "anthropic", "claude":
		return VendorAnthropic, nil
	case "openai":
		return VendorOpenAI, nil
	case "gemini":
		return VendorGemini, nil
	}
	return 0, fmt.Errorf("unknown vendor %q (want anthropic, openai or gemini)", s)
}

// isTerminal reports whether f is attached to a terminal rather than a pipe or
// a file. It is the difference between "a person is typing at me" and "a
// program is feeding me", and it is the only thing that decides whether the
// bare-invocation banner prints. Under grading stdin is always a pipe.
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
