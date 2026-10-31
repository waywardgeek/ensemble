// The blind virtual user is a small external Ensemble client. Its five tools
// are supplied by the browser peer; it has no filesystem, shell or DOM tools.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"ensemble/ensemble"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (runErr error) {
	url := flag.String("agent-url", "http://localhost:8088", "target GUI address")
	persona := flag.String("persona", "blind", "the supported speech-only persona")
	task := flag.String("task", "", "task for the virtual listener")
	peer := flag.String("peer", "virtual-user.mjs", "browser MCP peer script")
	browser := flag.String("browser", "", "optional Chromium executable")
	flag.Parse()
	if *persona != "blind" || *task == "" {
		return errors.New("use --persona blind and --task")
	}
	cfg, err := ensemble.ConfigFromEnv()
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		return errors.New("provider API key is not set")
	}
	cfg.SystemPrompt = "You are a blind virtual user. Use type and submit to ask the target agent, wait for completion and hear its speech. Your only perception is the transcript returned by hear or wait. Report exactly what you heard, distinguishing text from thinking and tool announcements. Do not infer screen contents or audio quality."
	root := ensemble.New(os.Stderr)
	a, err := root.NewAgent(cfg) // No builtins or skill lifecycle capabilities.
	if err != nil {
		return err
	}
	defer func() {
		runErr = errors.Join(runErr, a.Shutdown())
		file, err := os.Create("virtual-user.jsonl")
		if err == nil {
			err = errors.Join(a.History().Dump(file), file.Close())
		}
		runErr = errors.Join(runErr, err)
	}()
	script, err := filepath.Abs(*peer)
	if err != nil {
		return err
	}
	args := []string{script, "--agent-url", *url, "--persona", *persona}
	if *browser != "" {
		args = append(args, "--browser", *browser)
	}
	cmd := exec.Command("node", args...)
	// The browser needs neither model credentials nor unrelated environment.
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return err
	}
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return err
	}
	// MCP close ends stdin, letting the peer close its browser before exit.
	defer cmd.Wait()
	transport := ensemble.NewMCPTransport(stdout, stdin, stdin)
	defer transport.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if err := a.MCP().Connect(ctx, "blind", transport); err != nil {
		return err
	}
	for _, tool := range a.Declarations() {
		switch tool.Name {
		case "hear", "type", "submit", "wait", "sleep":
		default:
			return fmt.Errorf("blind peer advertised forbidden tool %q", tool.Name)
		}
	}
	answer, err := a.Ask(ctx, *task)
	if err == nil {
		fmt.Println(answer)
	}
	return err
}
