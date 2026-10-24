// Command ch07 measures delivery through the public library, not core internals.
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
	"time"

	"ensemble"
)

type counts struct {
	Text     int `json:"text"`
	Thinking int `json:"thinking"`
	ToolCall int `json:"tool_call"`
}

func (c *counts) add(kind ensemble.DeltaKind) {
	switch kind {
	case ensemble.DeltaText:
		c.Text++
	case ensemble.DeltaThinking:
		c.Thinking++
	case ensemble.DeltaToolCall:
		c.ToolCall++
	}
}

type stats struct {
	Observation  string         `json:"observation"`
	Deltas       counts         `json:"deltas"`
	Parts        counts         `json:"parts"`
	FirstDeltaMS float64        `json:"first_delta_ms"`
	ElapsedMS    float64        `json:"elapsed_ms"`
	Streaming    bool           `json:"streaming"`
	Usage        ensemble.Usage `json:"usage"`
	Assistant    string         `json:"assistant"`
}

func run(ctx context.Context, args []string, out, diagnostics io.Writer) (err error) {
	prompt := strings.Join(args, " ")
	if prompt == "" {
		return errors.New("usage: ch07 PROMPT")
	}
	config, err := ensemble.ConfigFromEnv(os.Getenv)
	if err != nil {
		return err
	}
	// This example has its own switch; it does not depend on the core CLI's.
	config.DisableStreaming = os.Getenv("CH07_NO_STREAM") == "1"
	config.LogPath = os.Getenv("CH07_AGENT_LOG")
	if config.LogPath == "" {
		config.LogPath = "ch07-agent.jsonl"
	}
	app := ensemble.New(log.New(diagnostics, "", 0))
	trace, err := os.OpenFile("api.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer trace.Close()
	app.SetTrace(trace)
	defer func() {
		if e := app.Shutdown(); err == nil {
			err = e
		}
	}()
	a, err := app.NewAgent(config)
	if err != nil {
		return err
	}
	q := a.Observe()
	defer q.Close()
	start := time.Now()
	report := stats{Observation: "stream_stats", FirstDeltaMS: -1, Streaming: a.Config().Streaming()}
	type key struct{ turn, response, part uint64 }
	kinds := map[key]ensemble.DeltaKind{}
	drain := func() {
		for _, o := range q.Drain() {
			switch v := o.(type) {
			case ensemble.PartDelta:
				if report.FirstDeltaMS < 0 {
					report.FirstDeltaMS = float64(time.Since(start)) / float64(time.Millisecond)
				}
				report.Deltas.add(v.Kind)
				kinds[key{v.TurnID, v.ResponseID, v.PartID}] = v.Kind
			case ensemble.PartFinal:
				// Tool results are durable finals but not model-response parts.
				if v.Part.Type == "tool_result" {
					continue
				}
				kind := kinds[key{v.TurnID, v.ResponseID, v.PartID}]
				if kind == "" {
					kind = ensemble.DeltaKind(v.Part.Type)
				}
				report.Parts.add(kind)
			}
		}
	}
	reply := a.Submit(ctx, prompt)
	for {
		select {
		case <-q.Wake():
			drain()
		case result := <-reply:
			drain()
			if result.Err != nil {
				return result.Err
			}
			report.Assistant = result.Text
			report.Usage = a.Usage()
			report.ElapsedMS = float64(time.Since(start)) / float64(time.Millisecond)
			return json.NewEncoder(out).Encode(report)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
