// A public consumer verifies independent operation identities and complete finals
// without depending on provisional display for request completion.
package main

import (
	"context"
	"encoding/json"
	"example.com/ensemble"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type observer struct {
	mu         sync.Mutex
	values     []ensemble.Observation
	terminal   chan struct{}
	once       sync.Once
	finalsOnly bool
}

func (o *observer) Observe(v ensemble.Observation) {
	o.mu.Lock()
	if !o.finalsOnly || v.Kind == "part_final" {
		o.values = append(o.values, v)
	}
	o.mu.Unlock()
	if v.Kind == "turn_ended" {
		o.once.Do(func() { close(o.terminal) })
	}
}

type slowObserver struct{ release <-chan struct{} }

func (s slowObserver) Observe(ensemble.Observation) { <-s.release }
func run() error {
	directory := os.Getenv("ENSEMBLE_RUN_DIRECTORY")
	if directory == "" {
		return fmt.Errorf("ENSEMBLE_RUN_DIRECTORY required")
	}
	app := ensemble.New(os.Stderr)
	defer app.Close()
	release := make(chan struct{})
	defer close(release)
	observers := []*observer{}
	finals := []*observer{}
	handles := []ensemble.RequestHandle{}
	slowIDs := []uint64{}
	for i := 0; i < 2; i++ {
		a, err := app.NewAgent(ensemble.Config{Vendor: os.Getenv("LLM_VENDOR"), Model: os.Getenv("LLM_MODEL"), ResolvedModel: os.Getenv("LLM_RESOLVED_MODEL"), APIKey: os.Getenv("LLM_API_KEY"), BaseURL: os.Getenv("LLM_BASE_URL"), LogPath: filepath.Join(directory, fmt.Sprintf("agent-%d.log", i+1)), MaxTokens: 512})
		if err != nil {
			return err
		}
		o := &observer{terminal: make(chan struct{})}
		f := &observer{terminal: make(chan struct{}), finalsOnly: true}
		observers = append(observers, o)
		finals = append(finals, f)
		app.Subscribe(a.ID(), o)
		app.Subscribe(a.ID(), f)
		id, _ := app.Subscribe(a.ID(), slowObserver{release})
		slowIDs = append(slowIDs, id)
		h, err := a.Submit(fmt.Sprintf("Start with AGENT-%d. Then explain in about 80 words why a streamed proposal must wait for complete validation before any tool effect. Do not use tools.", i+1))
		if err != nil {
			return err
		}
		handles = append(handles, h)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	completions := []ensemble.Completion{}
	for i, h := range handles {
		c, err := h.Wait(ctx)
		if err != nil {
			return err
		}
		if c.Outcome != "success" {
			return fmt.Errorf("public request %s: %s", c.RequestID, c.Outcome)
		}
		completions = append(completions, c)
		select {
		case <-observers[i].terminal:
		case <-ctx.Done():
			return ctx.Err()
		}
		select {
		case <-finals[i].terminal:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	all := [][]ensemble.Observation{}
	onlyFinals := [][]ensemble.Observation{}
	statuses := []string{}
	for i, o := range observers {
		o.mu.Lock()
		all = append(all, append([]ensemble.Observation(nil), o.values...))
		o.mu.Unlock()
		f := finals[i]
		f.mu.Lock()
		onlyFinals = append(onlyFinals, append([]ensemble.Observation(nil), f.values...))
		f.mu.Unlock()
		statuses = append(statuses, app.SubscriptionStatus(slowIDs[i]))
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"completions": completions, "observations": all, "finals_only": onlyFinals, "slow_subscription_status": statuses, "completion_before_slow_release": true})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
