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

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Actor wraps an Engine with a mailbox-driven event loop.
type Actor struct {
	eng  *Engine
	mb   *common.Mailbox
	host common.Agent
	gate *common.PauseGate // nil means never pause

	mu          sync.Mutex
	observers   []observerEntry
	obsIDSeq    uint64
	state       common.TurnState
	partSeq     uint64
	ctx         context.Context // actor lifetime; set once before its goroutine starts
	turnCtx     context.Context
	pending     []common.Inbound // requests/settings deferred while a tool runs
	result      common.TurnResult
	workers     sync.WaitGroup
	lifeMu      sync.Mutex
	started     bool
	cancel      context.CancelFunc
	done        chan struct{}
	shutdownErr error

	// obs is a buffered channel of observations that Wait can select on.
	obs chan common.Observation
}

type observerEntry struct {
	o  common.Observer
	id uint64
}

// NewActor creates an Actor wrapping the given Engine.
func NewActor(eng *Engine, host common.Agent) *Actor {
	return &Actor{
		eng:     eng,
		mb:      common.NewMailbox(),
		host:    host,
		state:   common.Idle,
		obs:     make(chan common.Observation, 256),
		done:    make(chan struct{}),
		turnCtx: context.Background(),
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

// drain processes all queued messages.
func (a *Actor) drain(ctx context.Context) {
	for {
		msgs := a.pending
		a.pending = nil
		msgs = append(msgs, a.mb.Drain()...)
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
	case common.Request:
		a.handleRequest(m)
	case common.UserMessage:
		a.turnCtx = a.ctx
		a.handleUserMessage(m)
	case common.Hint:
		a.handleHint(m)
	case common.ToolCompleted:
		if err := a.completeTool(m); err != nil {
			a.eng.logf("tool completion: %v", err)
		}
	case common.ToolEvents:
		if err := a.recordToolEvents(m.Events); err != nil {
			a.eng.logf("tool effects: %v", err)
		}
	case common.Interrupt:
		a.handleInterrupt()
	case common.Reset:
		a.handleReset()
	case common.SetModel:
		a.handleSetModel(m)
	}
}

// handleUserMessage is the single human-turn setup path for every caller.
func (a *Actor) handleUserMessage(m common.UserMessage) {
	a.setState(common.InputPending)

	if err := a.eng.Say(m.Text); err != nil {
		a.finishTurn("", err)
		return
	}

	// Recall runs AFTER the user's message has landed and BEFORE the model is
	// asked anything. Both halves of that sentence are load-bearing.
	//
	// After, because the retrieved material is a response to what was just
	// said, and an entry that landed first would read as context the user was
	// replying to rather than context fetched on their behalf.
	//
	// Before, because the whole point is that the model sees the material on
	// the turn where it is relevant. Attaching it afterwards would be an
	// elaborate way of answering the previous question.
	a.eng.attachRecall(m.Text)

	// "turn" ephemeral tools are called once when the turn starts.
	if err := a.eng.CallEphemeral("turn"); err != nil {
		a.finishTurn("", err)
		return
	}

	a.runTurnLoop(a.toolRoundLimit())
}

// runTurnLoop is the sole multi-round orchestration loop.
func (a *Actor) runTurnLoop(limit int) {
	rounds := 0
	for {
		a.setState(common.InFlight)

		// "round" ephemeral tools are called before every round.
		if err := a.eng.CallEphemeral("round"); err != nil {
			a.finishTurn("", err)
			return
		}

		reply, err := a.eng.TurnContext(a.turnCtx, a.streamWatch())
		if err != nil {
			a.finishTurn("", err)
			return
		}

		calls := a.eng.PendingCalls()
		if len(calls) == 0 {
			a.finishTurn(reply, nil)
			return
		}

		if rounds >= limit {
			err := &ToolRoundLimitError{Limit: limit}
			// Close recorded calls honestly: no dispatch occurred.
			for _, call := range calls {
				if recErr := a.eng.Record(common.Event{Type: common.ToolReturned, Tool: &common.ToolData{
					CallID: call.CallID, Name: call.Name, Args: call.Args, IsError: true,
					Parts: common.PartList{common.TextPart{Text: "Not executed: " + err.Error() + ". Ask to continue in a new turn."}},
				}}); recErr != nil {
					a.finishTurn(reply, recErr)
					return
				}
			}
			_ = a.eng.Record(common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{Message: err.Error()}})
			a.finishTurn(reply, err)
			return
		}

		// A model that acts without narrating cannot be supervised, and some
		// models will not narrate unless required to. Refuse the batch rather
		// than run tools whose effects nobody watched being decided.
		if a.enforceVisibleReasoning(reply, calls) {
			continue
		}

		rounds++
		a.setState(common.ToolsPending)

		// Dispatch and wait for tools one at a time. Serial dispatch lets
		// the pause gate hold execution between tools: if a client pauses
		// after the first tool finishes, the second never starts.
		for _, call := range calls {
			if a.gate != nil && !a.gate.WaitIfPaused(a.turnCtx) {
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
			if !a.waitForTool(call.CallID) {
				return // interrupted
			}
		}
	}
}

// dispatchTool dispatches a single tool call without blocking.
// visibleReasoningRequired is returned in place of every result in a batch of
// tool calls that arrived with no narration.
//
// It is long on purpose. It is read by a model that has just been stopped, and
// it has to do three things: say plainly that nothing ran, explain why the rule
// exists well enough that the model cooperates rather than works around it, and
// give an unambiguous next step.
const visibleReasoningRequired = `VISIBLE REASONING REQUIRED. Nothing was executed. The files on disk are unchanged.

You called a tool without first saying, in ordinary chat text, what you were
about to do and why.

HOW THIS COLLABORATION WORKS

Your human partner reads your narration as it streams, at roughly 750 words per
minute, through a screen reader. He has 20/180 vision. Reading a diff after the
fact is slow and painful for him in a way it is not for a sighted reviewer.

Your narration is not a courtesy and not a style preference. It IS the code
review, and it happens live, while you work. When you narrate as you go, he
follows the change as it is made, stops you mid-task when an assumption is
wrong, and by the time you finish he already understands the change and does not
need to review it at all.

When you work in silence, he cannot tell what you did, what you touched, or what
you broke. The only safe move left to him is to throw the entire run away with a
git reset, including the parts that were correct. Silence does not save time. It
destroys the whole session's output.

YOUR INTERNAL REASONING DOES NOT COUNT

It may be summarized, delayed, withheld, or never shown. He cannot read it, and
it does not say what you are about to do to the files on disk. Only ordinary
assistant text in the chat counts as narration.

REPORT EVERYTHING YOU RUN INTO

Everything you encounter matters, and the things that go wrong matter most. If a
tool did not behave as you expected, say so explicitly: what you expected, what
you got instead, and what you concluded from the difference. Do not quietly work
around it. Do not retry in silence and move on.

This agent is still being built, and your partner improves it using exactly
these reports. A surprise you absorb silently is a bug he never learns about; a
surprise you describe is a fix. If a tool is confusing, badly named, returns the
wrong shape, or omits something you needed, say that too. You are the only one
who can see it from the inside.

WHAT TO DO NOW

1. Say what you are about to do, and why, in one or two plain sentences. Name
   the file. Name the change. Say what you expect to happen.
2. Make the same tool call again. It will run.
3. Keep doing this before every tool call, and explain each edit as you make it,
   in enough detail that someone LISTENING rather than reading understands the
   change without opening the file.
4. When a result surprises you, say so before you continue.

For example:
   "I'll read internal/llm/actor.go around the dispatch loop to find where tool
    calls are gathered, since that is where the check has to go."
   followed by the tool call.

Do not apologize, do not quote this message back, and do not stop working.
Narrate, then carry on.`

// enforceVisibleReasoning refuses a batch of tool calls that arrived with no
// narration, for models that have to be told.
//
// It reports whether the batch was refused. Every call in the batch gets the
// same error result, so the model sees one consistent answer however many tools
// it asked for, and the next round carries the explanation in context.
//
// Nothing is executed. That is the point. A tool that has already run cannot be
// un-run by complaining afterwards, and effects the supervisor could not watch
// are precisely what makes a silent run unreviewable.
func (a *Actor) enforceVisibleReasoning(reply string, calls []common.ToolCallPart) bool {
	if strings.TrimSpace(reply) != "" {
		return false
	}
	f, ok := common.LookupModel(a.eng.Cfg.Model)
	if !ok || !f.RequiresVisibleReasoning {
		return false
	}
	for _, c := range calls {
		a.eng.Record(common.Event{Type: common.ToolReturned, Tool: &common.ToolData{
			CallID:  c.CallID,
			Name:    c.Name,
			Args:    c.Args,
			Parts:   common.PartList{common.TextPart{Text: visibleReasoningRequired}},
			IsError: true,
		}})
	}
	return true
}

// dispatchTool runs one tool call and records its result.
//
// EVERY call records a common.ToolReturned event, including the ones that fail. A
// failure is a RESULT, not an absence: the model asked a question and the
// answer is "that did not work, here is why". Dropping the result instead —
// or panicking, or ending the turn — leaves the model waiting for an answer to
// a question it can see it asked, and is the single most common way an agent
// locks up.
//
// THIS IS THE DISPATCH SITE, and the job model lives here and nowhere else.
// Before the tool runs it has a handle, an output file and a status; the tool
// runs on its own goroutine; and this goroutine waits on the JOB — until it
// is done, or the delay passes, or the pattern appears — rather than in the
// tool. Compare Chapter 3's version: the handler ran right here, and if it
// never came back, neither did the agent. Nothing about any of the six tools
// made that so. This function did.
//
// The supervision tools and tool_limits are the exception, marked NoJob: they
// act on jobs rather than being jobs, and they run inline.
func (a *Actor) dispatchTool(call common.ToolCallPart) error {
	// Resolve limits before ToolCalled so even a missing tool consumes a
	// pending one-shot setting. Bad patterns are tool errors, not turn errors.
	limits, fromPending, limErr := a.eng.Jobs.Take(call.Args)

	tool, err := a.eng.Tools.Lookup(call.Name)
	if err == nil && !a.eng.Tools.IsEnabled(call.Name) {
		// Registered but not enabled by any loaded skill. A bridged MCP
		// server registers everything it advertises; the SKILL.md decides
		// what may actually be called. Report it the way an unknown tool is
		// reported, because to this conversation it is one -- it was never
		// declared, and naming it as "disabled" would tell the model that
		// something it cannot see is there to be unlocked.
		err = fmt.Errorf("unknown tool: %s", call.Name)
	}
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
		c := &common.Call{Agent: a.host, Jobs: a.eng.Jobs, Limits: limits}
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

	// Allocate a handle for every tool before knowing what it will do.
	// read_file on an unavailable NFS mount hangs as well as a shell command.
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
	a.workers.Add(2)
	go func() {
		defer a.workers.Done()
		c := &common.Call{Agent: a.eng.Agent, Job: job, Jobs: a.eng.Jobs, Limits: limits}
		// No recover: a tool panic is an invariant violation, not model output.
		out, err := tool.Run(c, callCopy.Args)
		// Publish effects before Finish wakes a normal completion waiter.
		// Only the execution goroutine may read Call.Events while a tool runs.
		if len(c.Events) != 0 {
			mb.Post(common.ToolEvents{Events: c.Events})
		}
		// DeferFinish: the tool spawned a background goroutine that will
		// call Finish itself (e.g. an interactive PTY reader). Skip it
		// here so the job stays Running and Wait honours the delay/pattern.
		if !c.DeferFinish {
			job.Finish(out, err)
		}
	}()
	go func() {
		defer a.workers.Done()
		// Wait on the JOB, not the tool. A callback deadline returns control
		// even when a Go tool has not returned; it does not cancel that tool.
		reason := job.Wait(limits)
		result := job.Report(reason, limits)
		if fromPending {
			result = pendingNote(callCopy.Name, limits) + result
		}
		isError := job.Status() == common.StatusDone && job.Err() != nil

		// The actor owns log/context mutation, including late completions.
		mb.Post(common.ToolCompleted{
			CallID: callCopy.CallID, Result: result, IsError: isError,
			Tool: &common.ToolData{CallID: callCopy.CallID, Name: callCopy.Name,
				Args: callCopy.Args, Parts: common.PartList{common.TextPart{Text: result}},
				IsError: isError, Job: job.Data()},
		})
	}()

	return nil
}

// waitForTool waits for the dispatched call, processing hints along the way.
// A late completion from an interrupted turn still updates its tool card, but
// cannot release this call's wait. Returns false if interrupted.
func (a *Actor) waitForTool(callID string) bool {
	completed := false
	for !completed {
		select {
		case <-a.turnCtx.Done():
			a.finishTurn("", a.turnCtx.Err())
			return false
		case <-a.mb.Signal():
			msgs := a.mb.Drain()
			for i, msg := range msgs {
				switch m := msg.(type) {
				case common.ToolCompleted:
					if m.CallID == callID {
						completed = true
					}
					if err := a.completeTool(m); err != nil {
						a.pending = append(a.pending, msgs[i+1:]...)
						a.finishTurn("", err)
						return false
					}
				case common.ToolEvents:
					if err := a.recordToolEvents(m.Events); err != nil {
						a.pending = append(a.pending, msgs[i+1:]...)
						a.finishTurn("", err)
						return false
					}
				case common.Hint:
					a.handleHint(m)
				case common.Interrupt:
					a.pending = append(a.pending, msgs[i+1:]...)
					a.handleInterrupt()
					return false
				case common.UserMessage:
					// The same event, classified by where it arrives. Reaching THIS
					// drain means a turn is already running, and that is what makes
					// the message a hint rather than the start of a new turn. The
					// distinction is turn state, and this is the only place that knows
					// the turn state for certain.
					//
					// It used to be re-posted and queued for the next turn instead.
					// That was invisible to whoever sent it: a message typed while a
					// tool was running did not steer the work it was about. It waited,
					// then started a turn of its own, by which time the thing it was
					// about had already finished.
					a.handleHint(common.Hint{Text: m.Text})
				default:
					a.pending = append(a.pending, msg)
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
// handleReset clears the conversation and the memories recalled to serve it.
//
// The clearing itself is not done here. Record owns it, because the reducer is
// what defines a ConversationReset means and the journal is what makes it
// survive a restart: the journal replays on load, so the reset re-applies
// without anyone having to remember to write a snapshot first. Clearing the
// context directly would work until the next restart and then quietly undo
// itself, which is the worst failure shape available.
func (a *Actor) handleReset() {
	if err := a.eng.Record(common.Event{Type: common.ConversationReset}); err != nil {
		a.notify(common.TurnEnded{Err: "reset failed: " + err.Error()})
		return
	}
	a.notify(common.ConversationCleared{})
	a.setState(common.Idle)
	_ = a.eng.Save()
}

// handleSetModel switches the model that subsequent requests are rendered
// for. It runs on the actor goroutine between turns, so it can never land
// partway through rendering a request.
//
// Almost nothing needs rebuilding. The renderer and curate() both resolve
// features from Cfg.Model on every call, so they follow a switch on their
// own. Two things do not follow. Vendor and Surface are separate config
// fields and must move with the model, or the next request is rendered in
// one vendor's dialect and posted to another vendor's endpoint. And the
// model-gated tool set has to be re-resolved.
//
// An unknown model is refused rather than guessed at. Guessing would dial
// the wrong vendor and fail at the API with an error that says nothing
// about the real cause.
func (a *Actor) handleSetModel(m common.SetModel) {
	if m.Model == a.eng.Cfg.Model {
		return
	}
	v, ok := common.VendorFor(m.Model)
	if !ok {
		a.notify(common.TurnEnded{Err: "unknown model: " + m.Model})
		return
	}

	// The endpoint moves with the model. The renderer follows Vendor, so an
	// endpoint left behind renders one vendor's dialect and posts it to
	// another vendor's host, carrying that vendor's key. Resolve and refuse
	// before mutating anything: a half applied switch is worse than none.
	ep, ok := a.eng.Cfg.Endpoints[v]
	if !ok {
		a.notify(common.TurnEnded{Err: fmt.Sprintf("no endpoint configured for vendor %v, required by model %s", v, m.Model)})
		return
	}

	a.eng.Cfg.Model = m.Model
	a.eng.Cfg.Vendor = v
	a.eng.Cfg.Surface = common.SurfaceForModel(m.Model, v)
	a.eng.Cfg.BaseURL = ep.BaseURL
	a.eng.Cfg.APIKey = ep.APIKey
	// The credential provider moves with the vendor for the same reason the
	// endpoint does. Absent means "use Cfg.APIKey", so a vendor with no
	// provider correctly reverts to its metered key rather than inheriting
	// the previous vendor's bearer token.
	a.eng.Creds = a.eng.CredsByVendor[v]
	a.eng.logf("credential: vendor=%v now using %s", v, common.DescribeCredential(a.eng.Creds))
	if a.eng.Tools != nil {
		a.eng.Tools.SyncModelGatedTools(m.Model)
	}
	_ = a.eng.Save()
}

func (a *Actor) handleInterrupt() {
	// Interrupted, not Idle. The turn did not finish, and a state that said
	// "idle" would erase that: the next thing to read this could not tell an
	// aborted turn from one that ran to completion.
	//
	// It is safe to rest here because nothing gates on the state any more.
	// The GUI used to decide, from this value, whether a typed message was a
	// prompt or a hint -- which is what made a stale state unrecoverable. That
	// decision has moved to the actor, where the turn state is known for
	// certain, so a status line is now a status line and nothing else.
	a.setState(common.Interrupted)
	a.endTurn("", fmt.Errorf("interrupted"))
}

// finishTurn transitions to Idle and notifies observers.
func (a *Actor) finishTurn(text string, err error) {
	a.setState(common.Idle)
	a.endTurn(text, err)
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
		AllocPartID: func() uint64 {
			return atomic.AddUint64(&a.partSeq, 1)
		},
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
	host   common.Agent

	// merged observation channel for wake-once semantics
	merged chan common.Observation
}

// NewFramework creates a multi-agent framework.
func NewFramework(host common.Agent) *Framework {
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

// pendingNote heads the report of whichever call consumed a pending
// tool_limits. The limits are one-shot and land on the next call whatever it
// is. That is the design, and the footgun in it is that landing on the wrong
// call used to be silent: the model saw a truncated build two calls later
// with no cause in sight. This line puts the cause in the very result it
// produced. Loud, not different.
func pendingNote(tool string, l common.Limits) string {
	return fmt.Sprintf("[tool_limits consumed by this %s call: %s]\n", tool, l)
}
