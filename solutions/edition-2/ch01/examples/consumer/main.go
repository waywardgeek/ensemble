// This separate module exercises only Ensemble's public API against a real model.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"example.com/ensemble"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var diagnostics bytes.Buffer
	app := ensemble.New(&diagnostics)
	config := ensemble.Config{APIKey: os.Getenv("ANTHROPIC_API_KEY"), Model: os.Getenv("ANTHROPIC_MODEL"), BaseURL: os.Getenv("ANTHROPIC_BASE_URL")}
	a, err := app.NewAgent(config)
	if err != nil {
		return err
	}
	b, err := app.NewAgent(config)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(os.Stdout)
	turns := []struct {
		agent          *ensemble.Agent
		name, question string
	}{
		{a, "A", "My secret code is CORAL-271. Acknowledge it briefly."},
		{b, "B", "My secret code is HERON-839. Acknowledge it briefly."},
		{a, "A", "What is my secret code? Reply only with that code."},
		{b, "B", "What is my secret code? Reply only with that code."},
	}
	for _, turn := range turns {
		answer, err := turn.agent.Ask(context.Background(), turn.question)
		if err != nil {
			return err
		}
		if err := encoder.Encode(struct {
			Agent, Question, Answer string
			Usage                   ensemble.Usage
		}{turn.name, turn.question, answer, turn.agent.Usage()}); err != nil {
			return err
		}
	}
	for _, check := range []struct {
		agent      *ensemble.Agent
		own, other string
	}{{a, "CORAL-271", "HERON-839"}, {b, "HERON-839", "CORAL-271"}} {
		history := check.agent.History()
		if len(history) != 4 || !strings.Contains(history[3].Content, check.own) {
			return fmt.Errorf("live recall did not contain expected code")
		}
		for _, message := range history {
			if strings.Contains(message.Content, check.other) {
				return fmt.Errorf("independent Agent history leaked")
			}
		}
	}
	usageA, usageB := a.Usage(), b.Usage()
	// Local fault injection accompanies the live run; this is not a provider failure.
	broken, err := app.NewAgent(ensemble.Config{APIKey: "local-test", Model: config.Model, BaseURL: "http://127.0.0.1:1"})
	if err != nil {
		return err
	}
	_, err = broken.Ask(context.Background(), "exercise local transport failure")
	if err == nil || !strings.Contains(diagnostics.String(), "model transport failed") {
		return fmt.Errorf("local failure did not reach root logger")
	}
	if a.Usage() != usageA || b.Usage() != usageB || broken.Usage() != (ensemble.Usage{}) {
		return fmt.Errorf("usage crossed Agent boundaries")
	}
	return encoder.Encode(struct {
		Independent     bool
		UsageA, UsageB  ensemble.Usage
		LocalFailureLog string
	}{true, usageA, usageB, diagnostics.String()})
}
