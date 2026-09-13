// Command fakevendor serves the course's fake vendor API on a local port so a
// chapter binary can be driven by hand, instead of only through the grader.
//
// The grader mounts this same fake in-process. This command exists because
// being scored by the loop and watching it run are different things, and until
// now only the first was possible.
//
// One server serves all three vendor dialects; it routes on the request path,
// so the vendor is chosen by the client, not here. The -vendor flag only
// decides which environment block gets printed for you to paste.
//
//	go run ./cmd/fakevendor
//
// then, in another shell, paste the printed block and run a chapter:
//
//	echo 'what files are here?' | go run ./solutions/ch03 chat
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/waywardgeek/coding-agents-course/internal/fakevendor"
)

func main() {
	vendor := flag.String("vendor", "anthropic", "vendor dialect to print env for: anthropic, openai, or gemini")
	flag.Parse()

	srv := fakevendor.New(script())
	defer srv.Close()

	fmt.Printf("fake vendor listening at %s\n\n", srv.URL())
	fmt.Printf("export LLM_BASE_URL=%s\n", srv.URL())
	fmt.Printf("export LLM_API_KEY=dummy\n")
	fmt.Printf("export LLM_VENDOR=%s\n\n", *vendor)
	fmt.Printf("then:  echo 'what files are here?' | go run ./solutions/ch03 chat\n\n")
	fmt.Printf("Ctrl-C to stop.\n")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	fmt.Println("\nstopped.")
}

// script is a three-turn session that exercises the chapter 3 loop: the model
// asks for a tool, the agent runs it and reports back, the model answers.
//
// The third reply reports zero usage and is the terminator. A fake that runs
// out of replies repeats its last one, and a loop that executes tool calls will
// answer a repeated tool call forever — so the last reply must be one that ends
// the conversation rather than continuing it.
func script() []fakevendor.Reply {
	return []fakevendor.Reply{
		{
			ToolName: "list_directory",
			ToolArgs: `{"path":"."}`,
			ToolID:   "call_1",
			Usage:    fakevendor.Canonical{Input: 120, Output: 30},
		},
		{
			Text:  "Those are the files in the current directory.",
			Usage: fakevendor.Canonical{Input: 180, Output: 12},
		},
		{
			Text: "Nothing further.",
		},
	}
}
