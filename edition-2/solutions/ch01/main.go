// The executable supplies two clients of the same public Ensemble API.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"ensemble/ensemble"
)

func main() {
	// Failures stop this naive client. Continuing would conceal a failed turn;
	// the diagnostic belongs on stderr so stdout remains machine-readable.
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// Default startup is silent because the grader immediately sends JSON lines.
	chat := len(os.Args) == 2 && os.Args[1] == "chat"
	if len(os.Args) > 1 && !chat {
		return errors.New("usage: ch01 [chat]")
	}
	app := ensemble.New(os.Stderr)
	a, err := app.NewAgent(ensemble.Config{
		BaseURL: os.Getenv("ANTHROPIC_BASE_URL"),
		APIKey:  os.Getenv("ANTHROPIC_API_KEY"),
		Model:   os.Getenv("ANTHROPIC_MODEL"),
	})
	if err != nil {
		return err
	}
	// Both front ends call the same Agent; only their input/output framing differs.
	input := bufio.NewReader(os.Stdin)
	output := json.NewEncoder(os.Stdout)
	turns := 0
	if chat {
		fmt.Fprintf(os.Stderr, "talking to %s — Ctrl-D to quit\n", a.Config().Model)
	}
	for {
		if chat {
			fmt.Fprint(os.Stdout, "you> ")
		}
		// ReadString handles a last line without a newline and has no Scanner token cap.
		line, readErr := input.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return readErr
		}
		if len(line) == 0 && readErr == io.EOF {
			break
		}
		// Remove line framing, preserving the user text that will be replayed later.
		question := strings.TrimRight(line, "\r\n")
		if chat && strings.TrimSpace(question) == "" {
			continue
		}
		// Decode one complete line: two objects on the same line are not two turns.
		if !chat {
			var in struct {
				User string `json:"user"`
			}
			if err := json.Unmarshal([]byte(line), &in); err != nil {
				return errors.New("invalid input JSON")
			}
			question = in.User
		}
		answer, err := a.Ask(context.Background(), question)
		if err != nil {
			return err
		}
		// Two messages were retained: one question and one provider response.
		turns += 2
		if chat {
			fmt.Fprintf(os.Stdout, "\nclaude> %s\n\n", answer)
			// These are provider totals, separate from the local message count.
			u := a.Engine().Usage()
			fmt.Fprintf(os.Stderr, "[%d turns | %d input tokens | %d output tokens]\n", turns, u.Input, u.Output)
		} else if err := output.Encode(struct {
			Assistant string `json:"assistant"`
		}{answer}); err != nil {
			return err
		}
	}
	// In default mode stdout is exclusively the grader's JSON-lines protocol.
	if !chat {
		return output.Encode(struct {
			Usage ensemble.Usage `json:"usage"`
		}{a.Engine().Usage()})
	}
	return nil
}
