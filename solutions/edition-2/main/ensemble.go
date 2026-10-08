// Package ensemble provides a headless event-based conversation library.
package ensemble

import (
	"context"
	"errors"
	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/eventlog"
	"example.com/ensemble/internal/jobs"
	"example.com/ensemble/internal/llm"
	"example.com/ensemble/internal/persistence"
	"example.com/ensemble/internal/policy"
	"example.com/ensemble/internal/skills"
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
type RequestConfig = common.RequestConfig
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
	sessionReservations     map[string]bool
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
	if config.System == "" && config.Skills == nil {
		config.System = "Answer helpfully and concisely. Retain the conversation's details."
	}
	if config.MaxTokens <= 0 {
		config.MaxTokens = 512
	}
	return config
}
func (e *Ensemble) construct(config Config) (*Agent, error) {
	a := &Agent{parent: e, config: normalize(e, config)}
	a.codec = persistence.NewCodec(a)
	a.engine = llm.New(turnAgent{a})
	if config.Skills != nil {
		copied, err := skills.CopyConfiguration(skillAgent{a}, config.Skills)
		if err != nil {
			return nil, err
		}
		a.config.Skills = copied
	}
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
	if config.DataDir != "" {
		return nil, sessionError("session_conflict", "use OpenSession for DataDir")
	}
	if config.Skills != nil && config.System != "" {
		return nil, fmt.Errorf("System conflicts with selected primary skill")
	}
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
	var initial common.SkillCandidate
	if a.config.Skills != nil {
		a.skills, err = skills.New(skillAgent{a})
		if err != nil {
			return nil, err
		}
		initial, err = a.skills.Prepare(common.SkillOperation{Action: "initialize", Name: a.config.Skills.Primary})
		if err != nil {
			return nil, err
		}
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
	if initial != nil {
		fact := initial.Transition()
		if err = a.appendPrepared(Event{Type: "skills_initialized", Skills: &fact}, true, false, nil, initial); err != nil {
			_ = a.Close()
			return nil, err
		}
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
	// An offline reader never resolves a live source or scalar environment.
	config.Skills = nil
	var err error
	config.LogPath, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	a, err := e.construct(config)
	if err != nil {
		return nil, err
	}
	a.skills = skills.Historical(skillAgent{a})
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
		if i == 0 && event.Type == "session_anchor" {
			return nil, sessionError("session_origin_required", "anchor log requires session inspection with origin")
		}
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
	store               *persistence.Store
	sessionID           string
	identity            common.SessionIdentity
	resumed             bool
	reservedPath        string
	reservedID          bool
	origin              *common.SemanticState
	recent              []Event
	renderableCount     uint64
	eventCount          uint64
	inspectedCheckpoint *uint64
	codec               common.SessionCodec
	skills              *skills.Service
	policy              *policy.Service
	actor               *llm.Actor
	parent              common.Ensemble
	id                  string
	config              Config
	mu                  sync.Mutex
	appendMu            sync.Mutex
	closeMu             sync.Mutex
	closed              bool
	jobs                *jobs.Service
	operation           sync.Mutex
	events              []Event
	context             common.Context
	registry            *tools.Registry
	engine              *llm.Engine
	log                 common.EventLog
	faulted             bool
}

func (a *Agent) ModelReady(op common.ModelOperation) { a.actor.ModelReady(op) }

func (a *Agent) Codec() common.SessionCodec { return a.codec }

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
	out, _ := llm.Clone(a.engine, a.config)
	mode := a.context.SkillMode
	liveSkills := a.config.Skills != nil
	if mode {
		out.System = a.context.SkillPrimary
	}
	a.mu.Unlock()
	if mode && liveSkills {
		out.Tools = a.registry.Declarations()
	}
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
	current := a.Config()
	if config.DataDir != current.DataDir {
		return sessionError("session_conflict", "DataDir is creation-only")
	}
	if a.store != nil && a.identity.Mode == "plain" && config.System != current.System {
		return sessionError("session_incompatible", "base System is fixed for a session")
	}
	if !reflect.DeepEqual(config.Skills, current.Skills) {
		return fmt.Errorf("skill configuration is creation-only")
	}
	if current.Skills != nil {
		if config.System != "" && config.System != current.System {
			return fmt.Errorf("System cannot replace primary skill")
		}
		config.System = current.System
	}
	config = normalize(a.parent, config)
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
	c := a.Snapshot()
	if c.Session != nil && c.Session.Identity != nil && c.Session.Identity.Mode == "plain" {
		base := *c.Session.Identity.System
		if config.System != "" && config.System != base {
			return nil, sessionError("session_incompatible", "render base differs")
		}
		config.System = base
	}
	if c.SkillMode {
		if config.System != "" && config.System != c.SkillPrimary {
			return nil, fmt.Errorf("render System conflicts with recorded primary")
		}
		config.System = c.SkillPrimary
	}
	return llm.Render(a.engine, c, normalize(a.parent, config))
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
	if a.store != nil {
		if closeErr := a.store.Close(); err == nil {
			err = closeErr
		}
	}
	if a.reservedID {
		a.parent.ReleaseSession("id:" + a.sessionID)
		a.reservedID = false
	}
	if a.reservedPath != "" {
		a.parent.ReleaseSession("path:" + a.reservedPath)
		a.reservedPath = ""
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
	return a.appendPrepared(event, persist, notify, nil, nil)
}

func (a *Agent) appendPrepared(event Event, persist, notify bool, missing []int, prepared ...common.SkillCandidate) error {
	a.appendMu.Lock()
	defer a.appendMu.Unlock()
	a.mu.Lock()
	if persist && (a.faulted || a.log == nil) {
		a.mu.Unlock()
		return fmt.Errorf("Agent log is read-only or faulted")
	}
	if persist {
		if a.context.LastSeq == ^uint64(0) {
			a.faulted = true
			a.mu.Unlock()
			return sessionError("session_limit", "event identities exhausted")
		}
		event.Seq = a.context.LastSeq + 1
		event.Time = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if persist && a.store == nil {
		if err := eventlog.CheckSkillRecord(a, event); err != nil {
			a.mu.Unlock()
			return err
		}
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
	if err == nil && owned.Session != nil && persist && a.id != "" {
		err = sessionError("session_conflict", "session facts are construction-only")
	}
	if err == nil {
		err = a.jobs.ValidateLimitEvent(a.context, owned)
	}
	if err == nil && owned.Type == "response_ended" {
		err = a.engine.ValidateAccount(*owned.Response)
	}
	var skillCandidate common.SkillCandidate
	if err == nil && owned.Skills != nil {
		if a.skills == nil || persist && owned.Type == "skills_initialized" && a.id != "" {
			err = fmt.Errorf("skill initialization is construction-only")
		} else if persist {
			if len(prepared) > 0 {
				skillCandidate = prepared[0]
			}
			if skillCandidate == nil {
				skillCandidate, err = a.skills.Prepare(common.SkillOperation{Action: owned.Skills.Action, Name: owned.Skills.Name})
			}
			if err == nil && (!skillCandidate.Result().Changed || !reflect.DeepEqual(skillCandidate.Transition(), *owned.Skills)) {
				err = fmt.Errorf("skill fact differs from frozen catalog candidate")
			}
		} else {
			skillCandidate, err = a.skills.PrepareRecorded(*owned.Skills)
		}
	}
	if err == nil {
		err = llm.ValidateSessionCollections(a.engine, a.context, owned)
	}
	var encoded common.PreparedEvent
	if err == nil && persist && a.store != nil {
		encoded, err = a.log.Prepare(owned)
		if err == nil {
			owned = encoded.Event()
		} else {
			var problem *common.SessionError
			if errors.As(err, &problem) && problem.Code == "session_limit" {
				a.faulted = true
			}
		}
	}
	var applied Event
	if err == nil {
		applied, err = llm.Clone(a.engine, owned)
	}
	if err == nil && persist {
		if encoded != nil {
			err = a.log.AppendPrepared(encoded)
		} else {
			err = a.log.Append(owned)
		}
		if err != nil {
			a.faulted = true
		}
	}
	if err != nil {
		var storage *common.SessionError
		if persist && errors.As(err, &storage) && storage.Code == "session_limit" {
			a.faulted = true
		}
		a.mu.Unlock()
		return err
	}
	if skillCandidate != nil {
		a.skills.Apply(skillCandidate, owned.Seq)
	}
	a.events = append(a.events, owned)
	a.eventCount++
	if llm.RenderableEvent(a.engine, owned.Type) {
		a.renderableCount++
		a.recent = append(a.recent, owned)
		if len(a.recent) > 100 {
			a.recent = append([]Event{}, a.recent[len(a.recent)-100:]...)
		}
	}
	a.jobs.ApplyLimitEvent(owned)
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
	return a.appendPrepared(Event{Type: "response_ended", Response: &parsed.Response}, true, true, parsed.MissingCallIDs, nil)
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
	if e.nextHandle == ^uint64(0) {
		return 0
	}
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
	if a.origin != nil {
		if sequence <= a.origin.Session.AsOf {
			return nil, sessionError("history_unavailable", "request precedes available raw history")
		}
		state, _ = llm.Clone(a.engine, a.origin.Context)
		if a.origin.Skills != nil {
			for _, m := range a.origin.Skills.Material {
				if m.Record.Type == "primary" {
					state.SkillPrimary = m.Record.Body
				}
				for i := range state.Entries {
					entry := &state.Entries[i]
					if entry.Purpose == "skill" && entry.Activation == m.Record.Activation && m.Record.Body != "" {
						entry.Parts = []Part{Text(fmt.Sprintf("[skill %s activation %d]\n%s\n[/skill]", m.Record.Name, m.Record.Activation, m.Record.Body))}
					}
				}
			}
		}
	}
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
