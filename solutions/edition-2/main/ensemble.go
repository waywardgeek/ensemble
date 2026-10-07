// Package ensemble provides a headless event-based conversation library.
package ensemble

import (
	"context"
	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/eventlog"
	"example.com/ensemble/internal/jobs"
	"example.com/ensemble/internal/llm"
	"example.com/ensemble/internal/tools"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Context = common.Context
type Config = common.Config
type Usage = common.Usage
type Message = common.Message
type Conversation = common.Conversation
type Provenance = common.Provenance
type Part = common.Part
type Ref = common.Ref
type Entry = common.Entry
type Event = common.Event
type Response = common.Response
type ToolEvent = common.ToolEvent
type JobSnapshot = common.JobSnapshot
type ToolDefinition = common.ToolDefinition
type Redaction = common.Redaction
type Observation = common.Observation
type Observer = common.Observer
type ClientOwner = common.ClientOwner
type ClientRequest = common.ClientRequest
type ClientResult = common.ClientResult

type subscription struct {
	agentID  string
	observer Observer
}
type Ensemble struct {
	logger                  *log.Logger
	mu                      sync.Mutex
	agents                  map[string]*Agent
	nextAgent, nextObserver uint64
	nextHandle              uint64
	closed                  bool
	observers               map[uint64]subscription
}

func New(diagnostics io.Writer) *Ensemble {
	if diagnostics == nil {
		diagnostics = io.Discard
	}
	return &Ensemble{logger: log.New(diagnostics, "ensemble: ", 0), agents: map[string]*Agent{}, observers: map[uint64]subscription{}}
}
func (e *Ensemble) Logf(format string, args ...any) { e.logger.Printf(format, args...) }
func normalize(owner common.Ensemble, config Config) Config {
	if config.Vendor == "" {
		config.Vendor = "anthropic"
	}
	if config.BaseURL == "" {
		switch config.Vendor {
		case "anthropic":
			config.BaseURL = "https://api.anthropic.com"
		case "openai":
			config.BaseURL = "https://api.openai.com"
		case "gemini":
			config.BaseURL = "https://generativelanguage.googleapis.com"
		}
	}
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")
	if config.System == "" {
		config.System = "Answer helpfully and concisely. Retain the conversation's details."
	}
	if config.MaxTokens <= 0 {
		config.MaxTokens = 512
	}
	return config
}
func (e *Ensemble) construct(config Config) (*Agent, error) {
	a := &Agent{parent: e, config: normalize(e, config)}
	a.engine = llm.New(turnAgent{a})
	owned, err := llm.Clone(a.engine, a.config)
	if err != nil {
		return nil, err
	}
	a.config = owned
	a.config.Workspace, err = filepath.Abs(a.config.Workspace)
	if err != nil {
		return nil, fmt.Errorf("cannot capture Agent workspace: %w", err)
	}
	a.jobs = jobs.New(jobAgent{a})
	a.registry, err = tools.New(toolAgent{a}, a.config.Builtins)
	if err != nil {
		return nil, err
	}
	return a, nil
}
func (e *Ensemble) publishAgent(a *Agent) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return fmt.Errorf("Ensemble is closed")
	}
	e.nextAgent++
	a.id = fmt.Sprintf("agent-%d", e.nextAgent)
	e.agents[a.id] = a
	return nil
}
func (e *Ensemble) NewAgent(config Config) (*Agent, error) {
	a, err := e.construct(config)
	if err != nil {
		return nil, err
	}
	if err = llm.ValidateConfig(a.engine, a.config, true); err != nil {
		return nil, err
	}
	if !a.registry.Match(a.config.Tools) {
		return nil, fmt.Errorf("live tool declarations must match selected builtins")
	}
	a.config.Tools = a.registry.Declarations()
	if a.config.LogPath == "" {
		return nil, fmt.Errorf("a fresh LogPath is required")
	}
	a.log, err = eventlog.New(a, a.config.LogPath)
	if err != nil {
		return nil, err
	}
	if err = e.publishAgent(a); err != nil {
		_ = a.Close()
		return nil, err
	}
	return a, nil
}
func (e *Ensemble) Load(path string, config Config) (*Agent, error) {
	a, err := e.construct(config)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open event log")
	}
	defer f.Close()
	events, lines, err := eventlog.Read(a, f)
	if err != nil {
		return nil, err
	}
	for i, event := range events {
		if err = a.append(event, false, false); err != nil {
			return nil, llm.ParseEventError(a.engine, lines[i], err)
		}
	}
	if err = e.publishAgent(a); err != nil {
		_ = a.Close()
		return nil, err
	}
	return a, nil
}
func (e *Ensemble) Agent(id string) (*Agent, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	a := e.agents[id]
	if a == nil {
		return nil, fmt.Errorf("unknown Agent")
	}
	return a, nil
}
func (e *Ensemble) Submit(ctx context.Context, r ClientRequest) (ClientResult, error) {
	a, err := e.Agent(r.AgentID)
	if err != nil {
		return ClientResult{}, err
	}
	n := 0
	for _, b := range []bool{r.Prompt != nil, r.Ephemeral != nil, r.Redact != nil} {
		if b {
			n++
		}
	}
	if n != 1 {
		return ClientResult{}, fmt.Errorf("exactly one request directive is required")
	}
	if r.Prompt != nil {
		return a.Prompt(ctx, *r.Prompt)
	}
	if r.Ephemeral != nil {
		err = a.Ephemeral(*r.Ephemeral)
	} else {
		err = a.Redact(*r.Redact)
	}
	return ClientResult{Usage: a.Usage()}, err
}
func (e *Ensemble) Subscribe(agentID string, observer Observer) (uint64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.agents[agentID] == nil || observer == nil {
		return 0, fmt.Errorf("invalid observer subscription")
	}
	e.nextObserver++
	e.observers[e.nextObserver] = subscription{agentID, observer}
	return e.nextObserver, nil
}
func (e *Ensemble) Unsubscribe(id uint64) { e.mu.Lock(); defer e.mu.Unlock(); delete(e.observers, id) }
func (e *Ensemble) Publish(agentID string, event Event) {
	e.mu.Lock()
	listeners := []Observer{}
	a := e.agents[agentID]
	for _, s := range e.observers {
		if s.agentID == agentID {
			listeners = append(listeners, s.observer)
		}
	}
	e.mu.Unlock()
	// Publication follows durable application. Give each recipient its own copy so
	// one observer cannot change either history or another observer's notification.
	for _, listener := range listeners {
		copy, err := llm.Clone(a.engine, event)
		if err == nil {
			listener.Observe(Observation{AgentID: agentID, Seq: event.Seq, Kind: event.Type, Event: copy})
		}
	}
}

type Agent struct {
	parent    common.Ensemble
	id        string
	config    Config
	mu        sync.Mutex
	appendMu  sync.Mutex
	closeMu   sync.Mutex
	closed    bool
	jobs      *jobs.Service
	operation sync.Mutex
	events    []Event
	context   common.Context
	registry  *tools.Registry
	engine    *llm.Engine
	log       *eventlog.Log
	faulted   bool
}

func (a *Agent) ID() string                { return a.id }
func (a *Agent) Ensemble() common.Ensemble { return a.parent }

// Workspace returns the Agent-owned scalar without cloning unrelated declarations.
func (a *Agent) Workspace() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.config.Workspace
}
func (a *Agent) Config() Config {
	a.mu.Lock()
	defer a.mu.Unlock()
	out, _ := llm.Clone(a.engine, a.config)
	return out
}
func (a *Agent) SetConfig(config Config) error {
	if !a.operation.TryLock() {
		return fmt.Errorf("Agent busy")
	}
	defer a.operation.Unlock()
	config = normalize(a.parent, config)
	current := a.Config()
	if config.Workspace == "" {
		config.Workspace = current.Workspace
	}
	workspace, err := filepath.Abs(config.Workspace)
	if err != nil {
		return err
	}
	if workspace != current.Workspace || !reflect.DeepEqual(config.Builtins, current.Builtins) {
		return fmt.Errorf("workspace and builtin selection are fixed when the Agent is created")
	}
	config.Workspace = workspace
	if !a.registry.Match(config.Tools) {
		return fmt.Errorf("live tool declarations must match selected builtins")
	}
	config.Tools = a.registry.Declarations()
	if err := llm.ValidateConfig(a.engine, config, true); err != nil {
		return err
	}
	owned, err := llm.Clone(a.engine, config)
	if err != nil {
		return err
	}
	a.mu.Lock()
	a.config = owned
	a.mu.Unlock()
	return nil
}
func (a *Agent) Usage() Usage { a.mu.Lock(); defer a.mu.Unlock(); return a.engine.Usage() }
func (a *Agent) UsageByModel() map[Provenance]Usage {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.engine.UsageByModel()
}
func (a *Agent) Snapshot() common.Context {
	a.mu.Lock()
	defer a.mu.Unlock()
	out, _ := llm.Clone(a.engine, a.context)
	return out
}

// Internal checks borrow state only while locked; public snapshots still own copies.
func (a *Agent) canPrompt() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed || a.faulted {
		return fmt.Errorf("Agent is closed or faulted")
	}
	return llm.CanPrompt(a.engine, a.context)
}
func (a *Agent) nextSeq() uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.context.LastSeq + 1
}
func (a *Agent) Events() []Event {
	a.mu.Lock()
	defer a.mu.Unlock()
	out, _ := llm.Clone(a.engine, a.events)
	return out
}
func (a *Agent) History() Conversation {
	c := a.Snapshot()
	out := Conversation{}
	for _, entry := range c.Entries {
		role := entry.Actor
		if role == "human" {
			role = "user"
		}
		if role == "agent" {
			role = "assistant"
		}
		out = append(out, Message{Role: role, Content: llm.TextAnswer(a.engine, entry.Parts)})
	}
	return out
}
func (a *Agent) Render(config Config) ([]byte, error) {
	return llm.Render(a.engine, a.Snapshot(), normalize(a.parent, config))
}
func (a *Agent) Dump() ([]byte, error) { return eventlog.Dump(a, a.Events()) }
func (a *Agent) Close() error {
	a.closeMu.Lock()
	defer a.closeMu.Unlock()
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil
	}
	a.closed = true
	a.mu.Unlock()
	err := a.jobs.Close()
	a.appendMu.Lock()
	defer a.appendMu.Unlock()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.faulted = true
	if a.log != nil {
		if closeErr := a.log.Close(); err == nil {
			err = closeErr
		}
	}
	return err
}
func Text(text string) Part { return llm.Text(text) }
func (a *Agent) Append(event Event) error {
	if !a.operation.TryLock() {
		return fmt.Errorf("Agent busy")
	}
	defer a.operation.Unlock()
	return a.append(event, true, true)
}
func (a *Agent) append(event Event, persist, notify bool) error {
	a.appendMu.Lock()
	defer a.appendMu.Unlock()
	a.mu.Lock()
	if persist && (a.faulted || a.log == nil) {
		a.mu.Unlock()
		return fmt.Errorf("Agent log is read-only or faulted")
	}
	if persist {
		event.Seq = a.context.LastSeq + 1
		event.Time = time.Now().UTC().Format(time.RFC3339Nano)
	}
	owned, err := llm.Clone(a.engine, event)
	if err == nil && persist && owned.Type == "response_ended" {
		for _, index := range event.Response.MissingCallIDs {
			owned.Response.Parts[index].CallID = fmt.Sprintf("call-%d-%d", owned.Seq, index)
		}
	}
	if err == nil {
		err = llm.Validate(a.engine, a.context, &owned)
	}
	var applied Event
	if err == nil {
		applied, err = llm.Clone(a.engine, owned)
	}
	if err == nil && persist {
		err = a.log.Append(owned)
		if err != nil {
			a.faulted = true
		}
	}
	if err != nil {
		a.mu.Unlock()
		return err
	}
	a.events = append(a.events, owned)
	llm.Apply(a.engine, &a.context, applied)
	if owned.Type == "response_ended" {
		a.engine.Account(*owned.Response)
	}
	a.mu.Unlock()
	if notify {
		a.parent.Publish(a.id, owned)
	}
	return nil
}
func (a *Agent) Ephemeral(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("ephemeral text must be nonempty")
	}
	return a.Append(Event{Type: "message_received", Message: &Entry{Actor: "system", Purpose: "ephemeral", Parts: []Part{Text(text)}}})
}
func (a *Agent) Redact(r Redaction) error { return a.Append(Event{Type: "redacted", Redact: &r}) }
func (a *Agent) Ask(ctx context.Context, question string) (string, error) {
	result, err := a.Prompt(ctx, question)
	return result.Text, err
}
func (a *Agent) Prompt(ctx context.Context, question string) (ClientResult, error) {
	if !a.operation.TryLock() {
		return ClientResult{}, fmt.Errorf("Agent busy")
	}
	defer a.operation.Unlock()
	if strings.TrimSpace(question) == "" {
		return ClientResult{}, fmt.Errorf("question must be nonempty")
	}
	config := a.Config()
	if err := llm.ValidateConfig(a.engine, config, true); err != nil {
		return ClientResult{}, err
	}
	if err := a.canPrompt(); err != nil {
		return ClientResult{}, err
	}
	if err := a.append(Event{Type: "message_received", Message: &Entry{Actor: "human", Purpose: "dialogue", Parts: []Part{Text(question)}}}, true, true); err != nil {
		return ClientResult{}, err
	}
	return a.engine.Turn(ctx)
}

// Engine receives the Agent interface through a private admission adapter. These
// methods assume Prompt holds operation; public Append still acquires that lock.
type turnAgent struct{ *Agent }

func (a turnAgent) TurnSnapshot() common.Context        { return a.Snapshot() }
func (a turnAgent) RecordTurn(event common.Event) error { return a.append(event, true, true) }
func (a turnAgent) NextSequence() uint64                { return a.nextSeq() }
func (a turnAgent) Registry() common.Registry           { return a.registry }

func (e *Ensemble) AllocateHandle() uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextHandle++
	return e.nextHandle
}
func (e *Ensemble) Close() error {
	e.mu.Lock()
	e.closed = true
	agents := make([]*Agent, 0, len(e.agents))
	for _, a := range e.agents {
		agents = append(agents, a)
	}
	e.mu.Unlock()
	var first error
	for _, a := range agents {
		if err := a.Close(); first == nil {
			first = err
		}
	}
	return first
}

// Jobs has its own admission path: background completion cannot wait for Prompt.
type jobAgent struct{ *Agent }

func (a jobAgent) RecordJob(event common.Event) error { return a.append(event, true, true) }
func (a jobAgent) Registry() common.Registry          { return a.registry }
func (a jobAgent) Fault(err error) {
	a.mu.Lock()
	a.faulted = true
	a.mu.Unlock()
	a.parent.Logf("jobs faulted: %v", err)
}
func (a turnAgent) Jobs() common.Jobs { return a.jobs }

// Tool operations follow their Agent parent to Jobs and durable publication.
type toolAgent struct{ *Agent }

func (a toolAgent) Jobs() common.Jobs                   { return a.jobs }
func (a toolAgent) RecordTool(event common.Event) error { return a.append(event, true, true) }
