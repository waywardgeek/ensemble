// Command ch06 is application glue: three peer Agents and one shared manuscript.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"ensemble"
)

type manuscript struct {
	mu                     sync.Mutex
	text, decision, notes  string
	writes, edits, reviews int
}

func (m *manuscript) snapshot() (string, int, int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.text, m.writes, m.edits, m.reviews
}
func arguments(raw json.RawMessage, v any) error {
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return errors.New("one argument object required")
	}
	return nil
}
func (m *manuscript) replace(edit bool) ensemble.ToolHandler {
	return func(_ context.Context, call ensemble.ToolCall, raw json.RawMessage) error {
		var arg struct {
			Text *string `json:"text"`
		}
		if err := arguments(raw, &arg); err != nil {
			return err
		}
		if arg.Text == nil {
			return errors.New("text is required")
		}
		m.mu.Lock()
		m.text = *arg.Text
		if edit {
			m.edits++
		} else {
			m.writes++
		}
		m.mu.Unlock()
		_, err := io.WriteString(call, "draft stored")
		return err
	}
}
func (m *manuscript) read(count bool) ensemble.ToolHandler {
	return func(_ context.Context, call ensemble.ToolCall, raw json.RawMessage) error {
		var arg struct{}
		if err := arguments(raw, &arg); err != nil {
			return err
		}
		text, _, _, _ := m.snapshot()
		if count {
			text = fmt.Sprint(len(strings.Fields(text)))
		}
		_, err := io.WriteString(call, text)
		return err
	}
}
func (m *manuscript) review(_ context.Context, call ensemble.ToolCall, raw json.RawMessage) error {
	var arg struct {
		Decision string  `json:"decision"`
		Notes    *string `json:"notes"`
	}
	if err := arguments(raw, &arg); err != nil {
		return err
	}
	if (arg.Decision != "accept" && arg.Decision != "reject") || arg.Notes == nil {
		return errors.New("decision must be accept or reject; notes is required")
	}
	m.mu.Lock()
	m.decision = arg.Decision
	m.notes = *arg.Notes
	m.reviews++
	m.mu.Unlock()
	_, err := fmt.Fprintf(call, "%s: %s", arg.Decision, *arg.Notes)
	return err
}
func think(ctx context.Context, call ensemble.ToolCall, raw json.RawMessage) error {
	var arg struct {
		Seconds *float64 `json:"seconds"`
		Thought *string  `json:"thought"`
	}
	if err := arguments(raw, &arg); err != nil {
		return err
	}
	if arg.Seconds == nil || *arg.Seconds < 0 || *arg.Seconds > float64(int64(^uint64(0)>>1))/float64(time.Second) || arg.Thought == nil {
		return errors.New("nonnegative seconds and thought are required")
	}
	timer := time.NewTimer(time.Duration(*arg.Seconds * float64(time.Second)))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	_, err := io.WriteString(call, *arg.Thought)
	return err
}
func register(a ensemble.Agent, name, description, schema string, h ensemble.ToolHandler) error {
	return a.Tools().Register(ensemble.ToolDefinition{Name: name, Description: description, Parameters: json.RawMessage(schema)}, h)
}

const textSchema = `{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`
const emptySchema = `{"type":"object","properties":{},"additionalProperties":false}`

func roleTools(a ensemble.Agent, role string, m *manuscript) error {
	switch role {
	case "author":
		if err := register(a, "write_draft", "Store the complete draft text.", textSchema, m.replace(false)); err != nil {
			return err
		}
		if err := register(a, "word_count", "Count whitespace-separated words in the stored draft.", emptySchema, m.read(true)); err != nil {
			return err
		}
		return register(a, "think", "Wait seconds, then return thought.", `{"type":"object","properties":{"seconds":{"type":"number","minimum":0},"thought":{"type":"string"}},"required":["seconds","thought"],"additionalProperties":false}`, think)
	case "editor":
		if err := register(a, "read_draft", "Read the current stored draft.", emptySchema, m.read(false)); err != nil {
			return err
		}
		return register(a, "edit_draft", "Replace the complete stored draft with improved text.", textSchema, m.replace(true))
	case "reviewer":
		return register(a, "review", "Record accept or reject and review notes for the edited draft.", `{"type":"object","properties":{"decision":{"type":"string","enum":["accept","reject"]},"notes":{"type":"string"}},"required":["decision","notes"],"additionalProperties":false}`, m.review)
	}
	return fmt.Errorf("unknown role %q", role)
}

// The app owns formatting and external identity, never Agent history mutation.
type output struct {
	mu          sync.Mutex
	stdout, log *json.Encoder
	err         error
}

func (o *output) emit(v any) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.stdout.Encode(v); err != nil {
		o.err = errors.Join(o.err, err)
	}
	if err := o.log.Encode(v); err != nil {
		o.err = errors.Join(o.err, err)
	}
}
func (o *output) failure() error { o.mu.Lock(); defer o.mu.Unlock(); return o.err }

type observer struct {
	role   string
	output *output
}

func (o observer) Observe(v ensemble.Observation) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	fields := map[string]any{}
	if json.Unmarshal(b, &fields) != nil {
		return
	}
	fields["agent"] = o.role
	switch v.(type) {
	case ensemble.PartDelta:
		fields["observation"] = "part_delta"
	case ensemble.PartFinal:
		fields["observation"] = "part_final"
	case ensemble.StateChanged:
		fields["observation"] = "state_changed"
	default:
		return
	}
	o.output.emit(fields)
}
func pipeline(ctx context.Context, topic string, agents map[string]ensemble.Agent, m *manuscript, out *output) error {
	_, writes, edits, reviews := m.snapshot()
	for _, role := range []string{"author", "editor", "reviewer"} {
		prompt := "Write a draft about: " + topic + ". Store it with write_draft; use word_count to check it."
		if role != "author" {
			draft, _, _, _ := m.snapshot()
			if role == "editor" {
				prompt = "Read and improve this stored draft using read_draft and edit_draft:\n" + draft
			} else {
				prompt = "Review this edited draft. Call review with accept or reject and notes:\n" + draft
			}
		}
		result := <-agents[role].Submit(ctx, prompt)
		if result.Err != nil {
			return fmt.Errorf("%s: %w", role, result.Err)
		}
		_, w, e, r := m.snapshot()
		if role == "author" && w == writes || role == "editor" && e == edits || role == "reviewer" && r == reviews {
			return fmt.Errorf("%s completed without its required manuscript tool effect", role)
		}
		out.emit(map[string]string{"observation": "turn_ended", "agent": role})
		out.emit(map[string]string{"pipeline": role + "_done"})
	}
	return out.failure()
}

type input struct{ Kind, Text, Agent string }
type readResult struct {
	input input
	err   error
}

func readInputs(in io.Reader, ch chan<- readResult, done <-chan struct{}) {
	defer close(ch)
	r := bufio.NewReader(in)
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 {
			var v input
			decodeErr := json.Unmarshal([]byte(line), &v)
			select {
			case ch <- readResult{v, decodeErr}:
			case <-done:
				return
			}
			if decodeErr != nil {
				return
			}
		}
		if err != nil {
			if err != io.EOF {
				select {
				case ch <- readResult{err: err}:
				case <-done:
				}
			}
			return
		}
	}
}
func run() error { return runIO(os.Stdin, os.Stdout, os.Stderr, os.Getenv) }
func runIO(in io.Reader, stdout, diagnostics io.Writer, getenv func(string) string) (result error) {
	cfg, err := ensemble.ConfigFromEnv(getenv)
	if err != nil {
		return err
	}
	path := getenv("CH06_LOG")
	if path == "" {
		path = "ch06.jsonl"
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, f.Close()) }()
	out := &output{stdout: json.NewEncoder(stdout), log: json.NewEncoder(f)}
	app := ensemble.New(log.New(diagnostics, "", log.LstdFlags))
	defer func() { result = errors.Join(result, app.Shutdown(), out.failure()) }()
	agents := map[string]ensemble.Agent{}
	draft := &manuscript{}
	for _, role := range []string{"author", "editor", "reviewer"} {
		c := cfg
		c.Tools = []string{"tool_limits", "wait_for_job"}
		c.LogPath = path + "." + role + ".jsonl"
		c.SystemPrompt = "You are the " + role + " in a sequential author/editor/reviewer workflow. Use your role tools to update or review the shared manuscript. Finish only after the required tool work has completed."
		a, err := app.NewAgent(c)
		if err != nil {
			return err
		}
		agents[role] = a
		if err := roleTools(a, role, draft); err != nil {
			return err
		}
		a.Subscribe(observer{role, out})
	}
	done := make(chan struct{})
	defer close(done)
	inputs := make(chan readResult)
	go readInputs(in, inputs, done)
	completed := make(chan error, 1)
	var pending []string
	active := false
	for inputs != nil || active || len(pending) > 0 {
		if !active && len(pending) > 0 {
			topic := pending[0]
			pending = pending[1:]
			active = true
			go func() { completed <- pipeline(context.Background(), topic, agents, draft, out) }()
		}
		select {
		case r, ok := <-inputs:
			if !ok {
				inputs = nil
				continue
			}
			if r.err != nil {
				return r.err
			}
			v := r.input
			switch v.Kind {
			case "prompt":
				pending = append(pending, v.Text)
			case "hint":
				a := agents[v.Agent]
				if a == nil {
					return fmt.Errorf("unknown agent %q", v.Agent)
				}
				if err := a.Post(ensemble.Hint{Text: v.Text}); err != nil {
					return err
				}
				out.emit(map[string]any{"observation": "part_delta", "agent": v.Agent, "part_id": 0, "chunk": "hint: " + v.Text})
			case "interrupt":
				a := agents[v.Agent]
				if a == nil {
					return fmt.Errorf("unknown agent %q", v.Agent)
				}
				if err := a.Post(ensemble.Interrupt{}); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown input kind %q", v.Kind)
			}
		case err := <-completed:
			active = false
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
