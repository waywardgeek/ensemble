// Command fakevendor serves the course's fake vendor API on a local port so a
// chapter binary — ours or yours — can be driven by hand instead of only
// through the grader.
//
// The grader mounts this same fake in-process. This command exists because
// being scored by the loop and watching it run are different things, and
// until it existed only the first was possible.
//
// One server serves all three vendor dialects and routes on the request path
// (/v1/messages, /chat/completions, :generateContent), so the dialect is
// chosen by the CLIENT. The -vendor flag only decides which environment block
// is printed or handed to the chapter you asked it to run.
//
// Two ways to use it:
//
//	# 1. Run a reference solution against the fake, and watch it:
//	go run ./cmd/fakevendor -ch 3 chat
//	go run ./cmd/fakevendor -ch 3 -vendor gemini chat
//	echo '{"user":"what is here?"}' | go run ./cmd/fakevendor -ch 3 -vendor openai
//
//	# 2. Run YOUR agent against it: serve, and paste the printed env block.
//	go run ./cmd/fakevendor -vendor openai
//
// With -ch the fake is started, the chapter is run with its environment
// pointed at the fake (stdin/stdout are yours), and the fake is stopped when
// the chapter exits. -solution overrides the directory, so -ch 3 -solution
// ./my-agent runs your chapter 3 instead of ours.
//
// Every request the fake receives is traced on stderr: which dialect's
// endpoint it hit, whether it declared tools, whether it carried tool
// results, and which scripted reply was served. That trace is the loop
// turning: request 1 gets a tool call, request 2 carries the result, and the
// reply to request 2 is the model's answer.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

func main() {
	flag.Usage = usage
	vendor := flag.String("vendor", "anthropic", "dialect the chapter should speak: anthropic, openai, or gemini")
	ch := flag.Int("ch", 0, "run this chapter's solution against the fake (1, 2 or 3); 0 = serve only")
	solution := flag.String("solution", "", "solution directory to run (default ./solutions/chNN)")
	scriptName := flag.String("script", "", "reply script: tools (a tool call, then an answer) or text (answers only); default by chapter")
	addr := flag.String("addr", "", "fixed listen address, e.g. 127.0.0.1:8089 (default: a free port)")
	quiet := flag.Bool("q", false, "do not trace requests on stderr")
	flag.Parse()

	switch *vendor {
	case "anthropic", "openai", "gemini":
	default:
		fatalf("unknown vendor %q: want anthropic, openai or gemini", *vendor)
	}
	if *ch == 1 && *vendor != "anthropic" {
		fatalf("chapter 1 speaks only the Anthropic dialect; -vendor %s needs -ch 2 or later", *vendor)
	}
	if *scriptName == "" {
		*scriptName = "tools"
		if *ch == 1 || *ch == 2 {
			*scriptName = "text" // neither chapter executes a tool call
		}
	}
	replies, ok := scripts[*scriptName]
	if !ok {
		fatalf("unknown script %q: want tools or text", *scriptName)
	}

	var trace *os.File
	if !*quiet {
		trace = os.Stderr
	}
	srv := fakevendor.NewWithOptions(replies, fakevendor.Options{Cycle: true, Trace: trace, Addr: *addr})
	defer srv.Close()

	env := envFor(*vendor, srv.URL())

	if *ch == 0 {
		fmt.Printf("fake vendor listening at %s  (script: %s, cycling)\n", srv.URL(), *scriptName)
		fmt.Printf("all three dialects are served; the path you POST to picks the one you get.\n\n")
		fmt.Printf("# paste this in the shell where your agent runs:\n")
		for _, kv := range env {
			fmt.Printf("export %s\n", kv)
		}
		fmt.Printf("\n# then, for example:\n")
		fmt.Printf("#   go run ./solutions/ch03 chat\n")
		fmt.Printf("#   echo '{\"user\":\"what is here?\"}' | go run ./solutions/ch03\n\n")
		fmt.Printf("Ctrl-C to stop.\n")
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		fmt.Fprintln(os.Stderr, "\nfake: stopped.")
		return
	}

	dir := *solution
	if dir == "" {
		dir = fmt.Sprintf("./solutions/ch%02d", *ch)
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		fatalf("no solution directory at %s (run from the repository root, or pass -solution)", dir)
	}

	args := append([]string{"run", dir}, flag.Args()...)
	cmd := exec.Command("go", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), env...)
	fmt.Fprintf(os.Stderr, "fake: %s dialect at %s, script %q; running: go %s\n",
		*vendor, srv.URL(), *scriptName, strings.Join(args, " "))

	// Forward Ctrl-C to the child so a REPL exits the way it would on its own,
	// and the fake still gets closed by the deferred Close.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		for s := range sig {
			if cmd.Process != nil {
				_ = cmd.Process.Signal(s)
			}
		}
	}()

	err := cmd.Run()
	signal.Stop(sig)
	srv.Close()
	if ee, ok := err.(*exec.ExitError); ok {
		os.Exit(ee.ExitCode())
	}
	if err != nil {
		fatalf("run %s: %v", dir, err)
	}
}

// envFor is the environment a chapter needs to talk to the fake as the given
// vendor. LLM_* is what chapters 2+ read first; the vendor-prefixed names are
// their fallback and the only names chapter 1 knows. Both are set so the same
// block works for every chapter.
func envFor(vendor, url string) []string {
	prefix := map[string]string{"anthropic": "ANTHROPIC", "openai": "OPENAI", "gemini": "GEMINI"}[vendor]
	return []string{
		"LLM_VENDOR=" + vendor,
		"LLM_BASE_URL=" + url,
		"LLM_API_KEY=fake",
		"LLM_MODEL=fake-model",
		prefix + "_BASE_URL=" + url,
		prefix + "_API_KEY=fake",
		prefix + "_MODEL=fake-model",
	}
}

// scripts are the reply sequences the fake cycles through. They are fixed:
// the fake does not read your prompt, so its answers will not relate to what
// you typed — what you are watching is the protocol, not the conversation.
var scripts = map[string][]fakevendor.Reply{
	// tools: one chapter 3 turn. The model asks to see the directory, then to
	// read a file it "found", then answers. The chapter's loop must execute
	// both calls and send both results back before it gets the answer. Three
	// requests per user turn; with cycling, every turn does the same dance.
	"tools": {
		{ToolName: "list_directory", ToolArgs: `{"path":"."}`, ToolID: "call_ls",
			Usage: fakevendor.Canonical{Input: 220, Output: 30}},
		{ToolName: "read_file", ToolArgs: `{"path":"go.mod","start_line":1,"end_line":3}`, ToolID: "call_read",
			Usage: fakevendor.Canonical{Input: 410, Output: 40}},
		{Text: "I listed the directory and read the top of go.mod. That is the tool loop: I asked, your agent ran the tool and reported back, and now I am answering.",
			Usage: fakevendor.Canonical{Input: 560, Output: 45}},
	},
	// text: for chapters 1 and 2, which record a tool call but do not run one.
	"text": {
		{Text: "Hello from the fake vendor. I am a scripted reply and I did not read what you said.",
			Usage: fakevendor.Canonical{Input: 120, Output: 25}},
		{Text: "Still the fake. Every request you send is traced on stderr, so you can see the whole history being resent each turn.",
			Usage: fakevendor.Canonical{Input: 260, Output: 30}},
	},
}

func usage() {
	fmt.Fprintf(os.Stderr, `usage: go run ./cmd/fakevendor [-vendor V] [-ch N [-solution DIR] [chapter args...]] [-script tools|text] [-addr A] [-q]

  go run ./cmd/fakevendor -ch 3 chat                  run solutions/ch03 as a REPL against the fake (Anthropic dialect)
  go run ./cmd/fakevendor -ch 3 -vendor gemini chat   same, Gemini dialect
  go run ./cmd/fakevendor -ch 2 -vendor openai        chapter 2, JSON lines on stdin, OpenAI dialect
  go run ./cmd/fakevendor                             serve only; prints the env block for your own agent

flags:
`)
	flag.PrintDefaults()
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fakevendor: "+format+"\n", args...)
	os.Exit(2)
}
