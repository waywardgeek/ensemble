// This separate module is a consumer, not part of Ensemble's internal tree.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"ensemble/ensemble"
)

// An application can extend one Agent without teaching the framework its tool.
// The call's owner chain also gives the tool access to the application's logger.
func registerShout(a ensemble.Agent) error {
	return a.RegisterTool(ensemble.ToolDeclaration{
		Name: "shout", Description: "Return text converted to uppercase",
		Schema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}`),
	}, func(call ensemble.ToolContext, args json.RawMessage) (string, error) {
		// Schema guides the model; the handler still checks required input.
		// A pointer distinguishes missing text from a valid empty string.
		var input struct {
			Text *string `json:"text"`
		}
		if err := json.Unmarshal(args, &input); err != nil {
			return "", err
		}
		if input.Text == nil {
			return "", errors.New("text is required")
		}
		call.Engine().Agent().Ensemble().Logf("shout called")
		return strings.ToUpper(*input.Text), nil
	})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (runErr error) {
	// Applications share provider parsing with the CLI, while choosing their
	// own tool visibility. The zero BuiltinTools setting leaves only shout.
	cfg, err := ensemble.ConfigFromEnv()
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		return errors.New("provider API key is not set")
	}
	app := ensemble.New(os.Stderr)
	a, err := app.NewAgent(cfg)
	if err != nil {
		return err
	}
	// Save the actual calls/results for inspection; the model's final words alone
	// do not prove that our handler ran. Shutdown records any outstanding job kill.
	defer func() {
		runErr = errors.Join(runErr, a.Shutdown())
		file, err := os.Create("events.jsonl")
		if err == nil {
			err = errors.Join(a.History().Dump(file), file.Close())
		}
		runErr = errors.Join(runErr, err)
	}()
	// Install the declaration before the first request so visibility and
	// dispatch use the same per-Agent registry throughout the conversation.
	if err := registerShout(a); err != nil {
		return err
	}
	prompt := strings.Join(os.Args[1:], " ")
	if prompt == "" {
		prompt = "Call shout with text hello framework, then report its result."
	}
	// Bound this one-shot demonstration; library callers choose their own context.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	fmt.Printf("you> %s\n", prompt)
	answer, err := a.Ask(ctx, prompt)
	if err != nil {
		return err
	}
	fmt.Printf("assistant> %s\n", answer)
	fmt.Fprintf(os.Stderr, "usage: %+v\n", a.Engine().Usage())
	return nil
}
