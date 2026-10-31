// This external application observes delivery shape without knowing vendor wire
// formats. DisableStreaming changes request delivery, not the report's machinery.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"ensemble/ensemble"
)

type shape struct {
	mu     sync.Mutex
	start  time.Time
	first  time.Duration
	deltas map[string]int
	parts  map[string]int
}

// Observe counts in memory and returns promptly; terminal output waits until
// the request's reliable completion, never blocking the actor on an OS pipe.
func (s *shape) Observe(o ensemble.Observation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch p := o.(type) {
	case ensemble.PartDelta:
		if s.first == 0 {
			s.first = time.Since(s.start)
		}
		kind, _ := json.Marshal(p.Kind)
		s.deltas[strings.Trim(string(kind), "\"")]++
	case ensemble.PartFinal:
		if p.Part.Type != "tool_result" {
			s.parts[p.Part.Type]++
		}
	}
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := ensemble.ConfigFromEnv()
	if err != nil {
		return err
	}
	cfg.DisableStreaming = cfg.DisableStreaming || os.Getenv("CH07_NO_STREAM") == "1"
	cfg.BuiltinTools = true
	cfg.LogPath = "ch07-agent.jsonl"
	app := ensemble.New(os.Stderr)
	trace, err := os.Create("api.log")
	if err != nil {
		return err
	}
	defer trace.Close()
	app.SetAPILog(trace)
	a, err := app.NewAgent(cfg)
	if err != nil {
		return err
	}
	defer a.Shutdown()
	s := &shape{start: time.Now(), deltas: map[string]int{}, parts: map[string]int{}}
	a.Observe(s)
	prompt := "Use run_command to print STREAM_OK, then write two short sentences about a lighthouse keeper who collects fog."
	if len(os.Args) > 1 {
		prompt = strings.Join(os.Args[1:], " ")
	}
	answer, err := a.Ask(context.Background(), prompt)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"observation": "stream_stats", "streaming": ensemble.StreamingFor(cfg) != 0, "deltas": s.deltas, "parts": s.parts, "first_delta_ms": s.first.Milliseconds(), "elapsed_ms": time.Since(s.start).Milliseconds(), "assistant": answer, "usage": a.Engine().Usage()})
}
