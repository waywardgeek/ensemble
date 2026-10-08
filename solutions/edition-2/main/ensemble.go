// Package ensemble provides a headless event-based conversation library.
package ensemble

import (
	"context"
	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/eventlog"
	"example.com/ensemble/internal/jobs"
	"example.com/ensemble/internal/llm"
	"example.com/ensemble/internal/policy"
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
type TurnEvent = common.TurnEvent
type HintEvent = common.HintEvent
type RequestEvent = common.RequestEvent
type EventError = common.EventError
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
type RequestHandle = common.RequestHandle
type StoppedError = common.StoppedError
type Completion = common.Completion
type ControlAck = common.ControlAck
type Collection = common.Collection

type subscription struct {
	parent   common.Ensemble
	agentID  string
	observer Observer
	queue    chan Observation
	done     chan struct{}
	reason   string
}
type Ensemble struct {
	settingsPaths           map[string]bool
	logger                  *log.Logger
	mu                      sync.Mutex
	agents                  map[string]*Agent
	nextAgent, nextObserver uint64
	nextHandle              uint64
	closed                  bool
	observers               map[uint64]*subscription
}

func New(diagnostics io.Writer) *Ensemble {
	if diagnostics == nil {
		diagnostics = io.Discard
	}
	return &Ensemble{logger: log.New(diagnostics, "ensemble: ", 0), agents: map[string]*Agent{}, observers: map[uint64]*subscription{}}
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
	a.config.LogPath, err = filepath.Abs(a.config.LogPath)
	if err != nil {
		return nil, err
	}
	a.log, err = eventlog.New(a, a.config.LogPath)
	if err != nil {
		return nil, err
	}
	a.policy, err = policy.New(a, a.config.PolicyPath)
	if err != nil {
		_ = a.Close()
		return nil, err
	}
	a.actor = llm.NewActor(turnAgent{a})
	if err = e.publishAgent(a); err != nil {
		_ = a.Close()
		return nil, err
	}
	return a, nil
}
func (e *Ensemble) Load(path string, config Config) (*Agent, error) {
	var err error
	config.LogPath, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
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
	if e.closed {
		return 0, common.StoppedError{}
	}
	e.nextObserver++
	s := &subscription{parent: e, agentID: agentID, observer: observer, queue: make(chan Observation, 256), done: make(chan struct{})}
	e.observers[e.nextObserver] = s
	go func() {
		defer func() {
			e.mu.Lock()
			s.observer = nil
			s.queue = nil
			close(s.done)
			e.mu.Unlock()
		}()
		for observation := range s.queue {
			s.observer.Observe(observation)
		}
	}()
	return e.nextObserver, nil
}
func (e *Ensemble) Unsubscribe(id uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if s := e.observers[id]; s != nil && s.reason == "" {
		s.reason = "unsubscribed"
		close(s.queue)
	}
}
func (e *Ensemble) SubscriptionStatus(id uint64) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	if s := e.observers[id]; s != nil {
		return s.reason
	}
	return "unknown subscription"
}
func (e *Ensemble) Observe(o Observation) {
	e.mu.Lock()
	defer e.mu.Unlock()
	a := e.agents[o.AgentID]
	if a == nil {
		return
	}
	for _, s := range e.observers {
		if s.agentID != o.AgentID || s.reason != "" {
			continue
		}
		owned, err := llm.Clone(a.engine, o)
		if err != nil {
			continue
		}
		select {
		case s.queue <- owned:
		default:
			s.reason = "overflow"
			close(s.queue)
		}
	}
}
func (e *Ensemble) Publish(agentID string, event Event) {
	e.Observe(Observation{AgentID: agentID, Seq: event.Seq, Kind: event.Type, Event: event})
}

type Agent struct {
	policy    *policy.Service
	actor     *llm.Actor
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
	log       common.EventLog
	faulted   bool
}

func (a *Agent) ModelReady(op common.ModelOperation) { a.actor.ModelReady(op) }

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
	a.mu.Lock()
	stopped := a.closed || a.faulted
	a.mu.Unlock()
	if stopped {
		return fmt.Errorf("Agent stopped")
	}
	if !a.operation.TryLock() {
		return fmt.Errorf("Agent busy")
	}
	defer a.operation.Unlock()
	config = normalize(a.parent, config)
	current := a.Config()
	if config.PolicyPath == "" {
		config.PolicyPath = current.PolicyPath
	}
	if config.PolicyPath != current.PolicyPath {
		return fmt.Errorf("policy destination is fixed when the Agent is created")
	}
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
	if config.LogPath == "" {
		config.LogPath = current.LogPath
	}
	logPath, err := filepath.Abs(config.LogPath)
	if err != nil {
		return err
	}
	if logPath != current.LogPath {
		return fmt.Errorf("log destination is fixed when the Agent is created")
	}
	config.LogPath = logPath
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
func (a *Agent) Usage() Usage { return a.engine.Usage() }
func (a *Agent) UsageByModel() map[Provenance]Usage {
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
	if a.actor != nil {
		return a.actor.Close()
	}
	_ = a.jobs.Close()
	return a.finishClose()
}
func (a *Agent) finishClose() error {
	if a.policy != nil {
		a.policy.Close()
	}
	a.engine.Close()
	a.closeMu.Lock()
	defer a.closeMu.Unlock()
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil
	}
	a.closed = true
	a.mu.Unlock()
	var err error
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
	if a.actor != nil {
		return a.actor.Append(event)
	}
	if !a.operation.TryLock() {
		return fmt.Errorf("Agent busy")
	}
	defer a.operation.Unlock()
	return a.append(event, true, true)
}
func (a *Agent) append(event Event, persist, notify bool) error {
	return a.appendPrepared(event, persist, notify, nil)
}

func (a *Agent) appendPrepared(event Event, persist, notify bool, missing []int) error {
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
	if err == nil && len(missing) != 0 {
		for _, index := range missing {
			if !persist || owned.Type != "response_ended" || owned.Response == nil || index < 0 || index >= len(owned.Response.Parts) || owned.Response.Parts[index].Type != "tool_call" {
				err = fmt.Errorf("invalid parsed response call index")
				break
			}
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
		if a.actor != nil {
			a.actor.PublishDurable(owned)
		} else {
			a.parent.Publish(a.id, owned)
		}
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
func (a *Agent) Submit(question string) (RequestHandle, error) {
	if a.actor == nil {
		return nil, fmt.Errorf("Agent is read-only")
	}
	return a.actor.Submit(question)
}
func (a *Agent) Hint(text string) (ControlAck, error) {
	if a.actor == nil {
		return ControlAck{}, fmt.Errorf("Agent is read-only")
	}
	return a.actor.Hint(text)
}
func (a *Agent) Interrupt() (ControlAck, error) {
	if a.actor == nil {
		return ControlAck{}, fmt.Errorf("Agent is read-only")
	}
	return a.actor.Interrupt()
}
func (a *Agent) Cancel(id string) error {
	if a.actor == nil {
		return fmt.Errorf("Agent is read-only")
	}
	return a.actor.Cancel(id)
}
func (a *Agent) Prompt(ctx context.Context, question string) (ClientResult, error) {
	h, err := a.Submit(question)
	if err != nil {
		return ClientResult{}, err
	}
	c, err := h.Wait(ctx)
	if err != nil {
		if err == context.DeadlineExceeded {
			a.parent.Logf("blocking request timed out")
		} else {
			a.parent.Logf("blocking request canceled")
		}
		_ = h.Cancel()
		return ClientResult{}, err
	}
	result := ClientResult{Text: c.Text, Parts: c.Parts, Usage: a.Usage()}
	if c.Error != nil {
		return result, fmt.Errorf("%s: %s", c.Error.Code, c.Error.Message)
	}
	return result, nil
}

// The private adapter exposes actor-only persistence and lifecycle operations.
type turnAgent struct{ *Agent }

func (a turnAgent) TurnSnapshot() common.Context        { return a.Snapshot() }
func (a turnAgent) RecordTurn(event common.Event) error { return a.append(event, true, true) }
func (a turnAgent) RecordResponse(parsed common.ParsedResponse) error {
	return a.appendPrepared(Event{Type: "response_ended", Response: &parsed.Response}, true, true, parsed.MissingCallIDs)
}
func (a turnAgent) Engine() common.ModelEngine   { return a.engine }
func (a turnAgent) Policy() common.PolicyService { return a.policy }
func (a turnAgent) BeginTurn() error {
	a.operation.Lock()
	if err := a.canPrompt(); err != nil {
		a.operation.Unlock()
		return err
	}
	return nil
}
func (a turnAgent) EndTurn()                  { a.operation.Unlock() }
func (a turnAgent) FinishClose() error        { return a.finishClose() }
func (a turnAgent) NextSequence() uint64      { return a.nextSeq() }
func (a turnAgent) Registry() common.Registry { return a.registry }

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
	e.mu.Lock()
	for _, s := range e.observers {
		if s.reason == "" {
			s.reason = "closed"
			close(s.queue)
		}
	}
	e.mu.Unlock()
	return first
}

// Jobs has its own admission path: background completion cannot wait for Prompt.
type jobAgent struct{ *Agent }

func (a jobAgent) RecordJob(event common.Event) error {
	if a.actor != nil {
		return a.actor.Record(event)
	}
	return a.append(event, true, true)
}
func (a jobAgent) Registry() common.Registry { return a.registry }
func (a jobAgent) Fault(err error) {
	a.mu.Lock()
	a.faulted = true
	a.mu.Unlock()
	a.parent.Logf("jobs faulted: %v", err)
	if a.actor != nil {
		a.actor.Fault(err)
	}
}
func (a turnAgent) Jobs() common.Jobs { return a.jobs }

// Tool operations follow their Agent parent to Jobs and durable publication.
type toolAgent struct{ *Agent }

func (a toolAgent) Jobs() common.Jobs                   { return a.jobs }
func (a toolAgent) RecordTool(event common.Event) error { return a.append(event, true, true) }

func (e *Ensemble) SubmitPrompt(id, text string) (RequestHandle, error) {
	a, err := e.Agent(id)
	if err != nil {
		return nil, err
	}
	return a.Submit(text)
}
func (e *Ensemble) Hint(id, text string) (ControlAck, error) {
	a, err := e.Agent(id)
	if err != nil {
		return ControlAck{}, err
	}
	return a.Hint(text)
}
func (e *Ensemble) Interrupt(id string) (ControlAck, error) {
	a, err := e.Agent(id)
	if err != nil {
		return ControlAck{}, err
	}
	return a.Interrupt()
}

// ReconstructRequest renders the exact prefix/configuration of a captured send.
// It performs no model or tool effect and does not consume pending guidance.
func (a *Agent) ReconstructRequest(sequence uint64) ([]byte, error) {
	state := common.Context{}
	for _, event := range a.Events() {
		if event.Seq == sequence {
			if event.Type != "request_sent" || event.Request.Configuration == nil {
				return nil, fmt.Errorf("sequence lacks captured request configuration")
			}
			request := event.Request
			captured := request.Configuration
			config := Config{DisableStreaming: request.Delivery != "stream", Vendor: request.To.Vendor, Model: request.To.Model, System: captured.System, MaxTokens: captured.MaxTokens, Tools: captured.Tools, ResolvedModel: captured.ResolvedModel}
			return llm.Render(a.engine, state, config)
		}
		llm.Apply(a.engine, &state, event)
	}
	return nil, fmt.Errorf("request sequence not found")
}
