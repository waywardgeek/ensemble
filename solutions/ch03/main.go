package main

// ch02 — one log, three vendors.
//
//	./ch02              grader mode: the Chapter 1 stdio protocol, unchanged
//	./ch02 chat         the interactive loop from Chapter 1
//	./ch02 render LOG   play LOG -> context -> render; print the request JSON
//	./ch02 dump         write the event log as JSON-lines

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

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
	logPath := envOr("CH02_LOG", "ch02.log")

	switch mode {
	case "render":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: ch02 render LOG")
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
		runLoop(cfg, logPath, true)

	case "":
		runLoop(cfg, logPath, false)

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", mode)
		os.Exit(2)
	}
}

// runLoop is both the interactive REPL and the grader protocol. They differ
// only in how a line is read and how a reply is printed, which is the honest
// amount of difference between them.
func runLoop(cfg Config, logPath string, interactive bool) {
	eng := NewEngine(cfg, logPath)
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	if interactive {
		fmt.Fprintln(os.Stderr, "ch02 — type a message, ctrl-D to exit")
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

		prompt := line
		if !interactive {
			var msg struct {
				User      *string `json:"user"`
				Ephemeral *string `json:"ephemeral"`
			}
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				emit(out, map[string]string{"error": "bad input: " + err.Error()})
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
		cfg.Model = pick("LLM_MODEL", "GEMINI_MODEL", "gemini-3-pro")
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
