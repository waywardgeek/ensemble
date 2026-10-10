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

	"ensemble/ensemble"
)

// Display owns its slow output worker. Observe only offers a progress value;
// completion replies and persistence never depend on the terminal keeping up.
type display struct {
	queue   chan ensemble.Observation
	enabled atomic.Bool
}

// Observe offers progress without waiting for a consumer to receive it.
func (d *display) Observe(o ensemble.Observation) {
	if d.enabled.Load() {
		select {
		case d.queue <- o:
		default:
		}
	}
}
func converse(a ensemble.Agent, chat bool) (err error) {
	d := &display{queue: make(chan ensemble.Observation, 256)}
	d.enabled.Store(chat)
	a.Observe(d)
	var output sync.Mutex
	enc := json.NewEncoder(os.Stdout)
	emit := func(v any) {
		output.Lock()
		defer output.Unlock()
		enc.Encode(v)
	}
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		for o := range d.queue {
			if chat {
				output.Lock()
				switch v := o.(type) {
				case ensemble.PartDelta:
					if strings.HasPrefix(v.Chunk, "hint:") {
						fmt.Fprintln(os.Stdout, v.Chunk)
					}
				case ensemble.PartFinal:
					if v.Part.Type == "text" {
						fmt.Fprintf(os.Stdout, "assistant> %s\n", v.Part.Text)
					}
				case ensemble.StateChanged:
					if v.To == "interrupted" {
						fmt.Fprintln(os.Stdout, "[interrupted; other jobs continue]")
					}
				}
				output.Unlock()
			} else {
				emit(o)
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
		close(d.queue)
		<-finished
	}()
	if chat {
		fmt.Fprintln(os.Stderr, "Type a message, /interrupt to end this turn, Ctrl-D to quit. Messages while busy are hints.")
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
			if text == "/interrupt" {
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
	replies.Wait()
	if !chat {
		emit(map[string]any{"usage": a.Engine().Usage()})
	}
	return nil
}
