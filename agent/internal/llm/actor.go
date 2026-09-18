package llm

// The actor: a loop with a mailbox.
//
// The loop drains the mailbox and processes messages in order. UserMessage
// starts a turn, Hint attaches context, ToolCompleted records a result,
// Interrupt stops the turn. Observers fire on every state change and content
// event.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/waywardgeek/coding-agents-course/agent/internal/common"
)

// Actor wraps an Engine with a mailbox-driven event loop.
type Actor struct {
	eng  *Engine
	mb   *common.Mailbox
	host common.Host
	gate *common.PauseGate // nil means never pause

	mu        sync.Mutex
	observers []observerEntry
	obsIDSeq  uint64
	state     common.TurnState
	partSeq   uint64
	ctx       context.Context // set by Run, used by WaitIfPaused

	// obs is a buffered channel of observations that Wait can select on.
	obs chan common.Observation
}

type observerEntry struct {
	o  common.Observer
	id uint64
}

// NewActor creates an Actor wrapping the given Engine.
func NewActor(eng *Engine, host common.Host) *Actor {
	return &Actor{
		eng:   eng,
		mb:    common.NewMailbox(),
		host:  host,
		state: common.Idle,
		obs:   make(chan common.Observation, 256),
	}
}

// Send delivers into the mailbox. It NEVER blocks.
func (a *Actor) Send(msg common.Inbound) {
	a.mb.Post(msg)
}

// SetPauseGate attaches a pause gate that is checked before each tool
// dispatch. Pass nil to disable pausing.
func (a *Actor) SetPauseGate(g *common.PauseGate) {
	a.gate = g
}

// Attach registers an observer. Detach is by the returned func.
func (a *Actor) Attach(o common.Observer) (detach func()) {
	a.mu.Lock()
	a.obsIDSeq++
	id := a.obsIDSeq
	a.observers = append(a.observers, observerEntry{o: o, id: id})
	a.mu.Unlock()
	return func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		for i, e := range a.observers {
			if e.id == id {
				a.observers = append(a.observers[:i], a.observers[i+1:]...)
				return
			}
		}
	}
}

// State reports the current state without waiting.
func (a *Actor) State() common.TurnState {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

func (a *Actor) setState(s common.TurnState) {
	a.mu.Lock()
	old := a.state
	a.state = s
	a.mu.Unlock()
	if old != s {
		a.notify(common.StateChanged{From: old, To: s})
	}
}

// notify fires an observation to all observers (non-blocking) and to the obs
// channel for Wait consumers.
func (a *Actor) notify(obs common.Observation) {
	a.mu.Lock()
	observers := make([]common.Observer, len(a.observers))
	for i, e := range a.observers {
		observers[i] = e.o
	}
	a.mu.Unlock()
	for _, o := range observers {
		o.Observe(obs)
	}
	// Non-blocking send to the obs channel for Wait.
	select {
	case a.obs <- obs:
	default:
	}
}

// Wait blocks until an observation satisfies pred, or ctx is done.
func (a *Actor) Wait(ctx context.Context, pred func(common.Observation) bool) (common.Observation, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case obs := <-a.obs:
			if pred(obs) {
				return obs, nil
			}
		}
	}
}

// Ask posts a UserMessage and waits for the turn to end. This is the
// synchronous convenience: post + wait for turn-end.
func (a *Actor) Ask(text string) (string, error) {
	a.Send(common.UserMessage{Text: text})
	ctx := context.Background()
	obs, err := a.Wait(ctx, func(o common.Observation) bool {
		_, ok := o.(common.TurnEnded)
		return ok
	})
	if err != nil {
		return "", err
	}
	ended := obs.(common.TurnEnded)
	if ended.Err != "" {
		return ended.Text, fmt.Errorf("%s", ended.Err)
	}
	return ended.Text, nil
}

// Run starts the actor loop. It blocks until ctx is cancelled.
func (a *Actor) Run(ctx context.Context) {
	a.ctx = ctx
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.mb.Signal():
			a.drain(ctx)
		}
	}
}

// drain processes all queued messages.
func (a *Actor) drain(ctx context.Context) {
	for {
		msgs := a.mb.Drain()
		if len(msgs) == 0 {
			return
		}
		for _, msg := range msgs {
			select {
			case <-ctx.Done():
				return
			default:
			}
			a.handle(msg)
		}
	}
}

// handle processes a single inbound message.
func (a *Actor) handle(msg common.Inbound) {
	switch m := msg.(type) {
	case common.UserMessage:
		a.handleUserMessage(m)
	case common.Hint:
		a.handleHint(m)
	case common.ToolCompleted:
		// Stray completions outside of waitForTools — requeue them.
		a.mb.Post(m)
	case common.Interrupt:
		a.handleInterrupt()
	}
}

// handleUserMessage starts a new turn: Say, then loop Turn→Execute.
func (a *Actor) handleUserMessage(m common.UserMessage) {
	a.setState(common.InputPending)

	if err := a.eng.Say(m.Text); err != nil {
		a.finishTurn("", err)
		return
	}

	a.runTurnLoop()
}

// runTurnLoop is the actor version of Engine.Ask's loop. Instead of blocking
// on Execute, it dispatches tools and waits for ToolCompleted messages.
func (a *Actor) runTurnLoop() {
	for round := 0; ; round++ {
		a.setState(common.InFlight)

		reply, err := a.eng.Turn(a.streamWatch())
		if err != nil {
			a.finishTurn("", err)
			return
		}

		calls := a.eng.PendingCalls()
		if len(calls) == 0 {
			a.finishTurn(reply, nil)
			return
		}

		if round >= MaxToolRounds {
			_ = a.eng.Record(common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{
				Message: fmt.Sprintf("stopped after %d rounds of tool calls", MaxToolRounds),
			}})
			a.finishTurn(reply, nil)
			return
		}

		a.setState(common.ToolsPending)

		// Dispatch and wait for tools one at a time. Serial dispatch lets
		// the pause gate hold execution between tools: if a client pauses
		// after the first tool finishes, the second never starts.
		for _, call := range calls {
			if a.gate != nil && !a.gate.WaitIfPaused(a.ctx) {
				a.handleInterrupt()
				return
			}
			a.notify(common.ToolDispatched{
				CallID: call.CallID,
				Name:   call.Name,
				Input:  json.RawMessage(call.Args),
			})
			if err := a.dispatchTool(call); err != nil {
				a.finishTurn("", err)
				return
			}
			if !a.waitForTools(1) {
				return // interrupted
			}
		}
	}
}

// dispatchTool dispatches a single tool call without blocking.
func (a *Actor) dispatchTool(call common.ToolCallPart) error {
	// Resolve limits before dispatch.
	limits, fromPending, limErr := a.eng.Jobs.Take(call.Args)

	tool, err := a.eng.Tools.Lookup(call.Name)
	if err == nil {
		err = limErr
	}

	// NoJob tools run inline (they're supervision tools).
	if err != nil || tool.NoJob {
		if err := a.eng.Record(common.Event{Type: common.ToolCalled, Tool: &common.ToolData{
			CallID: call.CallID, Name: call.Name, Args: call.Args,
		}}); err != nil {
			return err
		}
		c := &common.Call{Host: a.host, Jobs: a.eng.Jobs, Limits: limits}
		var out string
		if err == nil {
			out, err = tool.Run(c, call.Args)
		}
		isError := false
		if err != nil {
			isError, out = true, err.Error()
		}
		if fromPending {
			out = pendingNote(call.Name, limits) + out
		}
		if err := a.eng.Record(common.Event{Type: common.ToolReturned, Tool: &common.ToolData{
			CallID: call.CallID, Name: call.Name, Args: call.Args,
			Parts: common.PartList{common.TextPart{Text: out}}, IsError: isError,
		}}); err != nil {
			return err
		}
		for _, ev := range c.Events {
			if err := a.eng.Record(ev); err != nil {
				return err
			}
		}
		// Post inline completion to mailbox so the counter works.
		a.mb.Post(common.ToolCompleted{CallID: call.CallID, Result: out, IsError: isError})
		return nil
	}

	// Normal tool: start a job, run on goroutine, post completion to mailbox.
	job, err := a.eng.Jobs.Start(call.Name, call.CallID)
	if err != nil {
		return err
	}
	if err := a.eng.Record(common.Event{Type: common.ToolCalled, Tool: &common.ToolData{
		CallID: call.CallID, Name: call.Name, Args: call.Args, Job: job.Data(),
	}}); err != nil {
		return err
	}

	mb := a.mb
	callCopy := call
	go func() {
		out, err := tool.Run(&common.Call{Host: a.eng.Host, Job: job, Jobs: a.eng.Jobs, Limits: limits}, callCopy.Args)
		job.Finish(out, err)

		reason := job.Wait(limits)
		result := job.Report(reason, limits)
		if fromPending {
			result = pendingNote(callCopy.Name, limits) + result
		}
		isError := job.Status() == common.StatusDone && job.Err() != nil

		// Record the tool result in the engine log.
		a.eng.Record(common.Event{Type: common.ToolReturned, Tool: &common.ToolData{
			CallID:  callCopy.CallID,
			Name:    callCopy.Name,
			Args:    callCopy.Args,
			Parts:   common.PartList{common.TextPart{Text: result}},
			IsError: isError,
			Job:     job.Data(),
		}})

		// Post to mailbox — the actor loop will process this.
		mb.Post(common.ToolCompleted{
			CallID:  callCopy.CallID,
			Result:  result,
			IsError: isError,
		})
	}()

	return nil
}

// waitForTools waits for `count` tool completions, processing hints along
// the way. Returns false if interrupted.
func (a *Actor) waitForTools(count int) bool {
	completed := 0
	for completed < count {
		select {
		case <-a.mb.Signal():
			msgs := a.mb.Drain()
			for _, msg := range msgs {
				switch m := msg.(type) {
				case common.ToolCompleted:
					completed++
					a.notify(common.ToolFinished{
						CallID:  m.CallID,
						Result:  m.Result,
						IsError: m.IsError,
					})
					a.notify(common.PartFinal{
						Seq:    common.Seq(len(a.eng.Log.Events)),
						PartID: atomic.AddUint64(&a.partSeq, 1),
						Part:   common.TextPart{Text: m.Result},
					})
				case common.Hint:
					a.handleHint(m)
				case common.Interrupt:
					a.handleInterrupt()
					return false
				case common.UserMessage:
					// Queue for next turn — don't process mid-tools.
					a.mb.Post(m)
				}
			}
		}
	}
	return true
}

// handleHint attaches a hint to the current turn's context.
func (a *Actor) handleHint(m common.Hint) {
	_ = a.eng.Attach(m.Text)
	a.notify(common.PartDelta{
		PartID: atomic.AddUint64(&a.partSeq, 1),
		Chunk:  "hint:" + m.Text,
	})
}

// handleInterrupt sets the interrupted state.
func (a *Actor) handleInterrupt() {
	a.setState(common.Interrupted)
	a.notify(common.TurnEnded{Err: "interrupted"})
}

// finishTurn transitions to Idle and notifies observers.
func (a *Actor) finishTurn(text string, err error) {
	a.setState(common.Idle)
	ended := common.TurnEnded{Text: text}
	if err != nil {
		ended.Err = err.Error()
	}
	a.notify(ended)
	_ = a.eng.Save()
}

// notifyContent is gone, and its absence is the point.
//
// It walked the finished dialogue entry and emitted a PartFinal per text
// part, minting a fresh PartID for each from a counter. That could never
// correlate with anything: the deltas did not exist yet, and when they did
// they would have had ids from the same counter, one per chunk. The parser
// now reports both ends of the stream, so the id that labelled the chunks is
// by construction the id that labels the finished part.
//
// It also emitted finals only for TextPart. Tool calls and reasoning got
// nothing, which is exactly the content an Actions pane most wants.

// streamWatch is the actor's half of the stream.
//
// The engine fills in the logging half. This half turns what the parser saw
// into observations, and it runs on the actor goroutine because Parse is
// called synchronously from the turn loop — the same reason it is safe to
// read the context here.
func (a *Actor) streamWatch() common.StreamCallbacks {
	return common.StreamCallbacks{
		OnDelta: func(partID uint64, kind common.DeltaKind, chunk string) {
			a.notify(common.PartDelta{PartID: partID, Kind: kind, Chunk: chunk})
		},
		OnPartFinal: func(partID uint64, part common.Part) {
			a.notify(common.PartFinal{Seq: a.entrySeq(), PartID: partID, Part: part})
		},
	}
}

// entrySeq is the Seq of the dialogue entry the response was folded into.
//
// Correct only because finals are reported AFTER the event that carries
// them, so the entry exists by the time this is asked.
func (a *Actor) entrySeq() common.Seq {
	entries := a.eng.Ctx.Dialogue
	if len(entries) == 0 {
		return 0
	}
	return entries[len(entries)-1].Seq
}

// Engine returns the underlying engine for log/context access.
func (a *Actor) Engine() *Engine { return a.eng }

// Shutdown ends the actor's engine.
func (a *Actor) Shutdown() error { return a.eng.Shutdown() }

// lastAgentText returns the last text the agent said.
func (a *Actor) lastAgentText() string {
	for i := len(a.eng.Ctx.Dialogue) - 1; i >= 0; i-- {
		if a.eng.Ctx.Dialogue[i].Actor != common.ActorAgent {
			continue
		}
		var b strings.Builder
		for _, p := range a.eng.Ctx.Dialogue[i].Parts {
			if t, ok := p.(common.TextPart); ok {
				b.WriteString(t.Text)
			}
		}
		return b.String()
	}
	return ""
}

// AgentID returns the agent ID for this actor (empty for single-agent).
func (a *Actor) AgentID() common.AgentID { return "" }

// ----------------------------------------------------------------
// Framework manages multiple actors.
// ----------------------------------------------------------------

// Framework manages multiple actors. It is minimal — just enough to support
// the exercise binary's three-agent workflow.
type Framework struct {
	mu     sync.Mutex
	actors map[common.AgentID]*Actor
	host   common.Host

	// merged observation channel for wake-once semantics
	merged chan common.Observation
}

// NewFramework creates a multi-agent framework.
func NewFramework(host common.Host) *Framework {
	return &Framework{
		actors: make(map[common.AgentID]*Actor),
		host:   host,
		merged: make(chan common.Observation, 256),
	}
}

// Add registers an actor with a given ID. It attaches a bridging observer
// that tags observations with the agent ID and feeds them to the merged
// channel.
func (f *Framework) Add(id common.AgentID, actor *Actor) {
	f.mu.Lock()
	f.actors[id] = actor
	f.mu.Unlock()

	// Bridge: tag observations with the agent ID and forward to merged.
	actor.Attach(ObserverFunc(func(obs common.Observation) {
		tagged := tagObservation(obs, id)
		select {
		case f.merged <- tagged:
		default:
		}
	}))
}

// Get returns the actor for a given ID.
func (f *Framework) Get(id common.AgentID) (*Actor, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.actors[id]
	return a, ok
}

// WaitAny blocks until any observation from any agent satisfies pred.
func (f *Framework) WaitAny(ctx context.Context, pred func(common.Observation) bool) (common.Observation, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case obs := <-f.merged:
			if pred(obs) {
				return obs, nil
			}
		}
	}
}

// Shutdown shuts down all actors.
func (f *Framework) Shutdown() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	var firstErr error
	for _, a := range f.actors {
		if err := a.Shutdown(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// ObserverFunc adapts a plain function to the Observer interface.
type ObserverFunc func(common.Observation)

// Observe implements common.Observer.
func (f ObserverFunc) Observe(o common.Observation) { f(o) }

// tagObservation sets the Agent field on an observation.
func tagObservation(obs common.Observation, id common.AgentID) common.Observation {
	switch v := obs.(type) {
	case common.PartDelta:
		v.Agent = id
		return v
	case common.PartFinal:
		v.Agent = id
		return v
	case common.StateChanged:
		v.Agent = id
		return v
	case common.TurnEnded:
		v.Agent = id
		return v
	case common.ToolDispatched:
		v.Agent = id
		return v
	case common.ToolFinished:
		v.Agent = id
		return v
	}
	return obs
}

// Timeout helper for use in exercises and graders.
func WithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}
