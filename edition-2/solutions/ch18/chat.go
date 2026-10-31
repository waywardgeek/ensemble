package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ensemble/ensemble"
	"golang.org/x/term"
)

// Display owns its slow output worker. Observe only offers a progress value;
// completion replies and persistence never depend on the terminal keeping up.
type display struct {
	queue   []ensemble.Observation
	mu      sync.Mutex
	wake    chan struct{}
	closed  bool
	enabled atomic.Bool
}

// Observe offers progress without waiting for a consumer to receive it.
func (d *display) Observe(o ensemble.Observation) {
	if !d.enabled.Load() {
		return
	}
	d.mu.Lock()
	d.queue = append(d.queue, o)
	d.mu.Unlock()
	select {
	case d.wake <- struct{}{}:
	default:
	}
}

// Shutdown the display only after the Agent joins all producers. A queue rather
// than a bounded channel preserves the incremental reply under a slow terminal.
func (d *display) close() {
	d.mu.Lock()
	d.closed = true
	d.mu.Unlock()
	select {
	case d.wake <- struct{}{}:
	default:
	}
}
func converse(a ensemble.Agent, chat bool) (err error) {
	d := &display{wake: make(chan struct{}, 1)}
	d.enabled.Store(chat)
	a.Observe(d)
	var output sync.Mutex
	writer := io.Writer(os.Stdout)
	if path := os.Getenv("CH07_LOG"); path != "" && !chat {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = io.MultiWriter(writer, file)
	}
	enc := json.NewEncoder(writer)
	emit := func(v any) {
		output.Lock()
		defer output.Unlock()
		enc.Encode(v)
	}
	started := time.Now()
	var first time.Duration
	counts, parts := map[string]int{}, map[string]int{}
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		colour := term.IsTerminal(int(os.Stderr.Fd()))
		for {
			<-d.wake
			d.mu.Lock()
			queue, closed := d.queue, d.closed
			d.queue = nil
			d.mu.Unlock()
			for _, o := range queue {
				switch v := o.(type) {
				case ensemble.PartDelta:
					if first == 0 {
						first = time.Since(started)
					}
					kind, _ := json.Marshal(v.Kind)
					counts[strings.Trim(string(kind), "\"")]++
				case ensemble.PartFinal:
					if v.Part.Type != "tool_result" {
						parts[v.Part.Type]++
					}
				}
				if !chat {
					emit(o)
					continue
				}
				output.Lock()
				switch v := o.(type) {
				case ensemble.PartDelta:
					if v.Hint {
						fmt.Fprintln(os.Stderr, v.Chunk)
						break
					}
					switch v.Kind {
					case ensemble.DeltaText:
						// os.File is unbuffered: each write reaches the pipe/terminal now.
						fmt.Fprint(os.Stdout, v.Chunk)
					case ensemble.DeltaThinking, ensemble.DeltaReasoningSummary, ensemble.DeltaToolCall:
						style := "\x1b[2m"
						if v.Kind == ensemble.DeltaToolCall {
							style = "\x1b[33m"
						}
						if colour {
							fmt.Fprint(os.Stderr, style)
						}
						fmt.Fprint(os.Stderr, v.Chunk)
						if colour {
							fmt.Fprint(os.Stderr, "\x1b[0m")
						}
					}
				case ensemble.PartFinal:
					// Finals correct widgets in richer clients. The terminal only separates
					// completed pieces; printing their content would duplicate every delta.
					if v.Part.Type == "text" {
						fmt.Fprintln(os.Stdout)
					} else if v.Part.Type == "opaque" || v.Part.Type == "tool_call" {
						fmt.Fprintln(os.Stderr)
					}
				case ensemble.StateChanged:
					if v.To == "interrupted" {
						fmt.Fprintln(os.Stderr, "[interrupted; other jobs continue]")
					}
				}
				output.Unlock()
			}
			if closed {
				return
			}
		}
	}()
	// stdin is independent of all turns. Legacy requests queue individually;
	// human messages and the new protocol are classified against live turn state.
	// Completion order follows input order even if Go schedules reply printers
	// differently. This chain never blocks the stdin reader or the actor.
	previous := make(chan struct{})
	close(previous)
	var replies sync.WaitGroup
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	stopSignals := make(chan struct{})
	defer func() { signal.Stop(signals); close(stopSignals) }()
	go func() {
		for {
			select {
			case <-signals:
				a.Post(ensemble.Interrupt{})
			case <-stopSignals:
				return
			}
		}
	}()
	defer func() {
		err = errors.Join(err, a.Shutdown())
		replies.Wait()
		d.close()
		<-finished
		// Drain the display before its summary so the answer cannot overtake
		// the stream records that explain it.
		if err == nil && !chat {
			if d.enabled.Load() {
				emit(map[string]any{"observation": "stream_stats", "deltas": counts, "parts": parts, "first_delta_ms": first.Milliseconds()})
				// Prompt/hint mode has no private request reply. At the EOF barrier its
				// final assistant entry is authoritative, independent of lossy progress.
				dialogue := a.History().Context().Dialogue
				if len(dialogue) > 0 {
					var answer strings.Builder
					for _, p := range dialogue[len(dialogue)-1].Parts {
						if p.Type == "text" {
							answer.WriteString(p.Text)
						}
					}
					emit(map[string]string{"assistant": answer.String()})
				}
			}
			emit(map[string]any{"usage": a.Engine().Usage()})
		}
	}()
	if chat {
		fmt.Fprintln(os.Stderr, "Type a message, /model NAME between turns, /interrupt to end this turn, Ctrl-D to quit. Messages while busy are hints.")
	}
	input := bufio.NewReader(os.Stdin)
	for {
		line, readErr := input.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return readErr
		}
		if len(line) == 0 && readErr == io.EOF {
			break
		}
		text := strings.TrimRight(line, "\r\n")
		if chat {
			if strings.HasPrefix(text, "/model ") {
				if err := a.SetModel(strings.TrimSpace(strings.TrimPrefix(text, "/model "))); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					fmt.Fprintln(os.Stderr, "Model:", a.Config().Model)
				}
			} else if text == "/interrupt" {
				a.Post(ensemble.Interrupt{})
			} else if strings.TrimSpace(text) != "" {
				a.Post(ensemble.UserMessage{Text: text})
			}
			continue
		}
		var in struct {
			// Kind selects the input or observation protocol variant. Text holds visible
			// provider or user content, including an explicitly empty string. User retains the
			// legacy prompt spelling for earlier clients.
			Kind, Text, User string
			// Ephemeral distinguishes a present instruction directive from a normal prompt.
			Ephemeral *string `json:"ephemeral"`
		}
		if json.Unmarshal([]byte(line), &in) != nil {
			return errors.New("invalid input JSON")
		}
		if in.Ephemeral != nil {
			if err := a.Ephemeral(*in.Ephemeral); err != nil {
				return err
			}
			emit(map[string]string{"ack": "ephemeral"})
			continue
		}
		if in.Kind != "" {
			d.enabled.Store(true)
			switch in.Kind {
			case "prompt":
				a.Post(ensemble.UserMessage{Text: in.Text})
			case "hint":
				a.Post(ensemble.Hint{Text: in.Text})
			case "interrupt":
				a.Post(ensemble.Interrupt{})
			default:
				return errors.New("unknown input kind")
			}
		} else {
			reply := a.Submit(context.Background(), in.User)
			before := previous
			previous = make(chan struct{})
			after := previous
			replies.Add(1)
			go func() {
				defer close(after)
				<-before
				defer replies.Done()
				result := <-reply
				if result.Err != nil {
					emit(map[string]string{"error": result.Err.Error()})
				} else {
					emit(map[string]string{"assistant": result.Text})
				}
			}()
		}
	}
	if err := a.Flush(context.Background()); err != nil {
		return err
	}

	return nil
}
