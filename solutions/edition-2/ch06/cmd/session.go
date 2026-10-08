package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"example.com/ensemble"
)

type inputLine struct {
	line string
	err  error
}
type completedRequest struct {
	value  ensemble.Completion
	err    error
	legacy bool
}

func inputLines(owner ensemble.ClientOwner, input io.Reader, human bool, stop <-chan struct{}) <-chan inputLine {
	lines := make(chan inputLine)
	go func() {
		if human {
			r := bufio.NewReader(input)
			for {
				line, err := readChatLine(owner, r)
				select {
				case lines <- inputLine{line, err}:
				case <-stop:
					return
				}
				if err != nil {
					return
				}
			}
		}
		scanner := bufio.NewScanner(input)
		scanner.Buffer(make([]byte, 4096), 16*1024*1024)
		for scanner.Scan() {
			select {
			case lines <- inputLine{line: scanner.Text()}:
			case <-stop:
				return
			}
		}
		err := scanner.Err()
		if err == nil {
			err = io.EOF
		} else {
			err = fmt.Errorf("cannot read input")
		}
		select {
		case lines <- inputLine{err: err}:
		case <-stop:
			return
		}
	}()
	return lines
}
func watchCompletion(h ensemble.RequestHandle, ch chan<- completedRequest, legacy bool) {
	go func() { c, err := h.Wait(context.Background()); ch <- completedRequest{c, err, legacy} }()
}
func completionError(item completedRequest) error {
	if item.err != nil {
		return item.err
	}
	if item.value.Error != nil {
		return fmt.Errorf("%s: %s", item.value.Error.Code, item.value.Error.Message)
	}
	return nil
}
func runProtocol(owner ensemble.ClientOwner, agent *ensemble.Agent, input io.Reader, output io.Writer) error {
	return runProtocolObserved(owner, agent, input, output, false)
}
func runProtocolObserved(owner ensemble.ClientOwner, agent *ensemble.Agent, input io.Reader, output io.Writer, observe bool) error {
	done := make(chan struct{})
	defer close(done)
	lines := inputLines(owner, input, false, done)
	completed := make(chan completedRequest)
	pending := 0
	reading := true
	var fatal error
	encoder := json.NewEncoder(output)
	var progress *progress
	var wake <-chan struct{}
	if observe {
		var err error
		progress, err = newProgress(owner, agent.ID())
		if err != nil {
			return err
		}
		defer owner.Unsubscribe(progress.subscription)
		wake = progress.wake
	}
	show := func(o ensemble.Observation) error {
		return encoder.Encode(map[string]any{"observation": observationRecord(o)})
	}
	gap := func() error {
		return encoder.Encode(map[string]any{"observation_gap": map[string]string{"agent_id": agent.ID(), "reason": "overflow"}})
	}
	invalid := func(reason string) error {
		return encoder.Encode(map[string]any{"error": map[string]string{"code": "invalid_control", "message": reason}})
	}
	stop := func(err error) { fatal = err; reading = false; lines = nil; _ = agent.Close() }
	for reading || pending > 0 {
		select {
		case item := <-lines:
			if item.err != nil {
				reading = false
				lines = nil
				if item.err != io.EOF {
					stop(item.err)
				}
				continue
			}
			if strings.TrimSpace(item.line) == "" {
				continue
			}
			var fields map[string]json.RawMessage
			if json.Unmarshal([]byte(item.line), &fields) != nil || fields == nil {
				stop(fmt.Errorf("exactly one valid directive is required"))
				continue
			}
			if raw, ok := fields["kind"]; ok {
				var kind, text string
				if json.Unmarshal(raw, &kind) != nil {
					if err := invalid("kind must be a control name"); err != nil {
						return err
					}
					continue
				}
				count := 1
				if kind == "prompt" || kind == "hint" {
					count = 2
				}
				valid := len(fields) == count
				if count == 2 {
					valid = valid && json.Unmarshal(fields["text"], &text) == nil && strings.TrimSpace(text) != ""
				}
				if !valid || (kind != "prompt" && kind != "hint" && kind != "interrupt") {
					if err := invalid("unknown control or invalid operation fields"); err != nil {
						return err
					}
					continue
				}
				switch kind {
				case "prompt":
					h, err := agent.Submit(text)
					if err != nil {
						stop(err)
						continue
					}
					if err = encoder.Encode(map[string]string{"accepted": "prompt", "request_id": h.ID()}); err != nil {
						stop(err)
						continue
					}
					pending++
					watchCompletion(h, completed, false)
				case "hint":
					ack, err := agent.Hint(text)
					if err != nil {
						if err = invalid("hint requires an active turn"); err != nil {
							return err
						}
						continue
					}
					if err = encoder.Encode(map[string]any{"ack": "hint", "request_id": ack.RequestID, "seq": ack.Seq, "sent": ack.Sent}); err != nil {
						stop(err)
					}
				case "interrupt":
					ack, err := agent.Interrupt()
					if err != nil {
						stop(err)
						continue
					}
					if err = encoder.Encode(map[string]any{"ack": "interrupt", "request_id": ack.RequestID, "interrupted": ack.Interrupted}); err != nil {
						stop(err)
					}
				}
				continue
			}
			if len(fields) != 1 {
				stop(fmt.Errorf("exactly one valid directive is required"))
				continue
			}
			var request ensemble.ClientRequest
			request.AgentID = agent.ID()
			ack := ""
			switch {
			case fields["user"] != nil:
				var text string
				if string(fields["user"]) == "null" || json.Unmarshal(fields["user"], &text) != nil {
					stop(fmt.Errorf("invalid user input"))
					continue
				}
				h, err := agent.Submit(text)
				if err != nil {
					stop(err)
					continue
				}
				pending++
				watchCompletion(h, completed, true)
				continue
			case fields["ephemeral"] != nil:
				var text string
				if string(fields["ephemeral"]) == "null" || json.Unmarshal(fields["ephemeral"], &text) != nil {
					stop(fmt.Errorf("invalid ephemeral input"))
					continue
				}
				request.Ephemeral = &text
				ack = "ephemeral"
			case fields["redact"] != nil:
				var redaction ensemble.Redaction
				if json.Unmarshal(fields["redact"], &redaction) != nil {
					stop(fmt.Errorf("invalid redaction"))
					continue
				}
				request.Redact = &redaction
				ack = "redact"
			default:
				stop(fmt.Errorf("unknown directive"))
				continue
			}
			if _, err := owner.Submit(context.Background(), request); err != nil {
				stop(err)
				continue
			}
			if err := encoder.Encode(map[string]string{"ack": ack}); err != nil {
				stop(err)
			}
		case <-wake:
			if err := progress.drain(show, gap); err != nil {
				return err
			}
		case item := <-completed:
			if progress != nil {
				if err := progress.beforeCompletion(agent, item.value, show, gap); err != nil {
					return err
				}
			}
			pending--
			c := item.value
			if item.legacy {
				if c.Outcome == "success" {
					if err := encoder.Encode(map[string]string{"assistant": c.Text}); err != nil {
						stop(err)
					}
				}
			} else {
				record := map[string]any{"request_id": c.RequestID, "outcome": c.Outcome, "text": c.Text, "pending_hints": c.PendingHints}
				if c.Error != nil {
					record["error"] = c.Error
				}
				if err := encoder.Encode(map[string]any{"completion": record}); err != nil {
					stop(err)
				}
			}
			if c.Outcome == "error" || c.Outcome == "round_limit" || item.err != nil {
				stop(completionError(item))
			}
		}
	}
	if err := agent.Close(); fatal == nil {
		fatal = err
	}
	if fatal != nil {
		return fatal
	}
	return encoder.Encode(struct {
		Usage ensemble.Usage `json:"usage"`
	}{agent.Usage()})
}
