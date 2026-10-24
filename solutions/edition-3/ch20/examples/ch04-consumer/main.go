// The consumer imports only the public library and uses its existing Agent loop.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"ensemble"
)

func shout(_ context.Context, call ensemble.ToolCall, raw json.RawMessage) error {
	var args struct {
		Text *string `json:"text"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&args); err != nil {
		return err
	}
	if args.Text == nil {
		return errors.New("text is required")
	}
	call.Parent().Parent().Logger().Print("shout called")
	_, err := io.WriteString(call, strings.ToUpper(*args.Text))
	return err
}
func run() (result error) {
	config, err := ensemble.ConfigFromEnv(os.Getenv)
	if err != nil {
		return err
	}
	app := ensemble.New(log.New(os.Stderr, "", log.LstdFlags))
	defer func() { result = errors.Join(result, app.Shutdown()) }()
	agent, err := app.NewAgent(config)
	if err != nil {
		return err
	}
	err = agent.Tools().Register(ensemble.ToolDefinition{Name: "shout", Description: "Uppercase the supplied text.", Parameters: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`)}, shout)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	answer, askErr := agent.Ask(ctx, `Use shout with text "hello framework", then give its result.`)
	if err = app.Shutdown(); err != nil {
		return err
	}
	if path := os.Getenv("CH02_LOG"); path != "" {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		dumpErr := agent.Log().Dump(f)
		closeErr := f.Close()
		if err = errors.Join(dumpErr, closeErr); err != nil {
			return err
		}
	}
	if askErr != nil {
		return askErr
	}
	_, err = fmt.Fprintln(os.Stdout, answer)
	return err
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
