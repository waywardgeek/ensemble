// The application supplies the workflow and role tools. Adding the third role
// changes this program only; Ensemble has no author/editor/reviewer knowledge.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ensemble/ensemble"
)

// A single writer owns stdout and the exercise log. Observe enqueues without
// waiting for either sink; reliable workflow completion does not consume progress.
type output struct {
	mu    sync.Mutex
	queue []any
	wake  chan struct{}
	done  chan struct{}
	file  *os.File
	err   error
}

func (o *output) push(v any) {
	o.mu.Lock()
	o.queue = append(o.queue, v)
	o.mu.Unlock()
	select {
	case o.wake <- struct{}{}:
	default:
	}
}
func (o *output) run() {
	defer close(o.done)
	writer := io.MultiWriter(os.Stdout, o.file)
	encoder := json.NewEncoder(writer)
	for {
		<-o.wake
		o.mu.Lock()
		q := o.queue
		o.queue = nil
		o.mu.Unlock()
		for _, v := range q {
			if v == nil {
				return
			}
			if err := encoder.Encode(v); err != nil {
				o.err = errors.Join(o.err, err)
			}
		}
	}
}

type observer struct {
	name   string
	output *output
}

// Observe offers progress without waiting for a consumer to receive it.
func (o observer) Observe(v ensemble.Observation) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return
	}
	var envelope map[string]any
	json.Unmarshal(bytes, &envelope)
	envelope["agent"] = o.name
	o.output.push(envelope)
}

// The shared manuscript is application data, synchronized because tool workers
// can outlive their turn. It is not copied into configuration or a core service.
type manuscript struct {
	mu   sync.Mutex
	text string
}

func (m *manuscript) read() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.text
}
func (m *manuscript) write(text string) {
	m.mu.Lock()
	m.text = text
	m.mu.Unlock()
}
func register(a ensemble.Agent, name, description, schema string, run func(ensemble.ToolContext, json.RawMessage) (string, error)) error {
	return a.RegisterTool(ensemble.ToolDeclaration{Name: name, Description: description, Schema: json.RawMessage(schema)}, run)
}
func roleTools(a ensemble.Agent, role string, draft *manuscript) error {
	textSchema := `{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}`
	put := func(_ ensemble.ToolContext, raw json.RawMessage) (string, error) {
		var in struct {
			// Text holds visible provider or user content, including an explicitly empty string.
			Text *string
		}
		if err := json.Unmarshal(raw, &in); err != nil {
			return "", err
		}
		if in.Text == nil {
			return "", errors.New("text is required")
		}
		draft.write(*in.Text)
		return "draft stored", nil
	}
	switch role {
	case "author":
		if err := register(a, "write_draft", "Store the authored draft", textSchema, put); err != nil {
			return err
		}
		if err := register(a, "word_count", "Count words in the stored draft", `{"type":"object","properties":{}}`, func(_ ensemble.ToolContext, _ json.RawMessage) (string, error) {
			return fmt.Sprint(len(strings.Fields(draft.read()))), nil
		}); err != nil {
			return err
		}
		// The published fixture's slow thought is local to this exercise. It proves
		// that an ordinary Go handler cannot block the actor's hint acknowledgement.
		return register(a, "think", "Pause to consider a thought", `{"type":"object","properties":{"seconds":{"type":"number"},"thought":{"type":"string"}},"required":["seconds","thought"]}`, func(_ ensemble.ToolContext, raw json.RawMessage) (string, error) {
			var in struct {
				// Seconds is the requested thought duration, validated as nonnegative.
				Seconds float64
				// Thought is the text retained after the exercise's slow thought completes.
				Thought string
			}
			if err := json.Unmarshal(raw, &in); err != nil {
				return "", err
			}
			if in.Seconds < 0 {
				return "", errors.New("seconds must be nonnegative")
			}
			time.Sleep(time.Duration(in.Seconds * float64(time.Second)))
			return in.Thought, nil
		})
	case "editor":
		if err := register(a, "read_draft", "Read the author's stored draft", `{"type":"object","properties":{}}`, func(_ ensemble.ToolContext, _ json.RawMessage) (string, error) { return draft.read(), nil }); err != nil {
			return err
		}
		return register(a, "edit_draft", "Replace the stored draft with an improved version", textSchema, put)
	case "reviewer":
		return register(a, "review", "Record accept or reject and notes", `{"type":"object","properties":{"decision":{"type":"string","enum":["accept","reject"]},"notes":{"type":"string"}},"required":["decision","notes"]}`, func(_ ensemble.ToolContext, raw json.RawMessage) (string, error) {
			var in struct {
				// Decision is the reviewer's explicit accept or reject result. Notes supplies the
				// reviewer's reason alongside its decision.
				Decision, Notes string
			}
			if err := json.Unmarshal(raw, &in); err != nil {
				return "", err
			}
			if in.Decision != "accept" && in.Decision != "reject" {
				return "", errors.New("decision must be accept or reject")
			}
			return in.Decision + ": " + in.Notes, nil
		})
	}
	return errors.New("unknown role")
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() (err error) {
	cfg, err := ensemble.ConfigFromEnv()
	if err != nil {
		return err
	}
	path := os.Getenv("CH06_LOG")
	if path == "" {
		path = "ch06.jsonl"
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	out := &output{wake: make(chan struct{}, 1), done: make(chan struct{}), file: file}
	go out.run()
	defer func() {
		out.push(nil)
		<-out.done
		err = errors.Join(err, out.err, file.Close())
	}()
	app := ensemble.New(os.Stderr)
	agents := map[string]ensemble.Agent{}
	draft := &manuscript{}
	instructions := map[string]string{
		"author":   "You are the author. Write a short draft about the user's topic using write_draft, check word_count, and return the draft. Use at most 100 words.",
		"editor":   "You are the editor. Call read_draft, improve clarity and imagery, store the improved draft using edit_draft, then return it. Use at most 100 words.",
		"reviewer": "You are the reviewer. Evaluate the supplied edited draft. Call review with accept or reject and concise notes, then state the decision.",
	}
	for _, name := range []string{"author", "editor", "reviewer"} {
		roleConfig := cfg
		roleConfig.SystemPrompt = instructions[name]
		roleConfig.DataDir = filepath.Join("cr", name)
		roleConfig.LogPath = path + "." + name
		a, createErr := app.NewAgent(roleConfig)
		if createErr != nil {
			return createErr
		}
		agents[name] = a
		defer func() { err = errors.Join(err, a.Shutdown()) }()
		if err := roleTools(a, name, draft); err != nil {
			return err
		}
		a.Observe(observer{name, out})
	}
	// One pipeline is the exercise, not a reusable orchestration language. Its
	// requests have reliable private replies; observations remain display only.
	var workflow sync.WaitGroup
	defer workflow.Wait()
	var workflowErr error
	started := false
	input := bufio.NewReader(os.Stdin)
	for {
		line, readErr := input.ReadBytes('\n')
		if readErr != nil && readErr != io.EOF {
			return readErr
		}
		if len(line) == 0 {
			break
		}
		var in struct {
			// Kind selects the input or observation protocol variant. Text holds visible
			// provider or user content, including an explicitly empty string. User retains the
			// legacy prompt spelling for earlier clients. Agent selects the named workflow role
			// receiving live input.
			Kind, Text, User, Agent string
		}
		if err := json.Unmarshal(line, &in); err != nil {
			return err
		}
		if in.Kind == "prompt" || (in.Kind == "" && in.User != "") {
			if started {
				return errors.New("one topic per workflow")
			}
			started = true
			topic := in.Text
			if topic == "" {
				topic = in.User
			}
			workflow.Add(1)
			go func() {
				defer workflow.Done()
				text := topic
				for _, name := range []string{"author", "editor", "reviewer"} {
					result := <-agents[name].Submit(context.Background(), text)
					if result.Err != nil {
						workflowErr = result.Err
						return
					}
					if name != "reviewer" {
						if stored := draft.read(); stored != "" {
							text = stored
						} else {
							text = result.Text
						}
					}
					out.push(map[string]string{"observation": "turn_ended", "agent": name})
					out.push(map[string]string{"pipeline": name + "_done"})
				}
			}()
		} else {
			target := in.Agent
			if target == "" {
				target = "author"
			}
			a := agents[target]
			if a == nil {
				return fmt.Errorf("unknown agent %q", target)
			}
			switch in.Kind {
			case "hint":
				a.Post(ensemble.Hint{Text: in.Text})
			case "interrupt":
				a.Post(ensemble.Interrupt{})
			default:
				return errors.New("unknown input kind")
			}
		}
		if readErr == io.EOF {
			break
		}
	}
	workflow.Wait()
	return workflowErr
}
