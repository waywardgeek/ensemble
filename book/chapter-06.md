# Chapter 6: Two Seams and a Loop

The agent framework from Chapter 5 is good enough to write an
author-editor orchestrator. Imagine what it could do with a GUI
built exactly the way you want it. Your layout, your keybindings,
your idea of what coding with AI should feel like. That is where this
book is headed. But the tools that shipped GUI-first paid for it:
the GUI imported the framework, the framework imported the GUI, and
by the time anyone noticed the cycle it was load-bearing.
CodeRhapsody made the same mistake, and refactoring it out took weeks
of focused work. This chapter exists so you do not repeat it. By the
end, your framework will have an API clean enough to host any GUI,
any gateway, any agentic application. The GUI chapter comes next.
This one builds the surface it plugs into.

That surface is two seams and a loop. Right now, ask a question while
a tool runs and the agent does not notice until the tool finishes.
Send a hint and it arrives one round too late. Run two agents at once
and the second one blocks until the first is done. The package
structure is right and the engine is deaf.

This chapter fixes the deafness. Three additions, no existing code
removed: an outbound seam so the framework tells the world what
happened instead of the world reaching in to ask; an inbound queue so
prompts, hints, tool completions, and interrupts enter through one
door; and a loop between them that drains the queue and notifies
observers on every event. A framework is two seams and a loop. By
the end, the same core runs a three-agent workflow where each agent
has its own tools, its own prompt, and its own observers, and a hint
mid-tool-call arrives before the tool finishes.

## TL;DR

**What you build.** Three things added to Chapter 5's star topology:

1. An **observer** interface in the hub (outbound seam).
2. A **mailbox** queue in a new implementation package (inbound seam).
3. An **actor loop** in the engine that drains the mailbox and notifies
   observers.

Plus a multi-agent coordinator, `Ref` and `ModelFeatures` types in the
hub, and a three-agent exercise that proves the framework can run
completely different agents without changing a line inside it.

**The observer seam.** One interface, three event types:

```go
// Observer receives notifications. Observe must not block.
type Observer interface {
    Observe(Observation)
}

type Observation interface{ observation() }

type PartDelta struct {          // streaming chunk
    PartID uint64 `json:"part_id"`
    Chunk  string `json:"chunk"`
}

type PartFinal struct {          // completed part
    Seq    Seq      `json:"seq"`
    PartID uint64   `json:"part_id"`
    Part   TextPart `json:"part"`
}

type StateChanged struct {       // turn-state transition
    From   TurnState `json:"from"`
    To     TurnState `json:"to"`
}

type AgentID string
```

`Observe` must not block: a slow observer that holds up the loop
recreates the deafness this chapter exists to fix. Single-agent logs
stay unchanged from Chapter 5; the framework tags agent identity
externally when coordinating multiple agents.

**The mailbox.** One inbound queue carries everything:

```go
type Inbound interface{ inbound() }

type UserMessage struct{ Text string }
type Hint        struct{ Text string }
type Interrupt   struct{}
type ToolCompleted struct {
    CallID  string
    Result  string
    IsError bool
}
```

A prompt, a hint, a tool finishing, and an interrupt are the same
kind of event to the loop. `Post(msg Inbound)` must not block or
drop. The queue is a mutex-guarded slice with a signal channel;
nothing in it is clever.

**The actor loop.** The engine runs on its own goroutine. When the
mailbox has a message, the loop drains it:

- `UserMessage`: start a new turn (render context, send to vendor, record
  events).
- `Hint`: attach to the current turn's pending context.
- `ToolCompleted`: record the result, check for outstanding calls, continue
  the turn or go idle.
- `Interrupt`: set state to `Interrupted`, stop processing.

The synchronous `Ask(text) (string, error)` still exists. It posts a
`UserMessage` and waits for the turn to end. Callers that do not need
blocking use `Post` directly and watch through an observer.

**The Wait primitive.**

```go
func (a *Actor) Wait(ctx context.Context,
    pred func(Observation) bool) (Observation, error)
```

Blocks until an observation satisfies `pred` or the context cancels.
Every higher-level operation derives from this: blocking send is
`Post` plus `Wait` for idle; waiting for two agents is `Wait` on
whichever fires first.

**Ref and ModelFeatures.** `BlobPart` gains a `Ref` instead of a bare
path string:

```go
type RefKind uint8
const (
    RefPath   RefKind = iota + 1
    RefURI
    RefHandle
)

type Ref struct {
    Kind    RefKind `json:"kind"`
    Locator string  `json:"locator"`
}

func (r Ref) Zero() bool { return r.Kind == 0 }
```

`ModelFeatures` declares what media a model accepts:

```go
type Media uint8
const (
    MediaImage    Media = 1 << iota
    MediaAudio
    MediaVideo
    MediaDocument
)

type ModelFeatures struct {
    Media Media
}

func LookupModel(model string) (ModelFeatures, bool)
```

No default row. An unknown model returns `false`, and the renderer
refuses with an error naming the model and the unsupported media type.
Dropping a part silently is the Chapter 1 mistake: a format the model
cannot read is a lie, not a degradation.

**Rules.** Seven checks.

1. **Chapter 5 still passes.** The actor upgrade changes no observable
   behavior from Chapter 5's star topology. (`ch5-parity`, 10)
2. **The agent hears hints mid-tool.** A slow tool runs for three
   seconds. One second in, a hint arrives. The observer stream shows
   the hint *before* the tool completion. (`not-deaf`, 25)
3. **Replay matches live.** Truncate the event log at the last
   `request_sent`, replay it, and the resulting context matches the
   live context byte for byte. (`replay-is-live`, 15)
4. **Observers fire.** The observation stream contains both state
   transitions (`StateChanged`) and completed content (`PartFinal`).
   (`observer-fires`, 15)
5. **Two agents, one wakeup.** Two agents finish in the same turn
   window. The parent's observation stream shows one wakeup, not two.
   (`wake-once`, 15)
6. **Unsupported media refused.** Present media the model does not
   support. The renderer returns an error naming the model and the
   media type, not a silent drop. (`loud-refusal`, 10)
7. **Hub stays clean.** `go list -deps` on `internal/common/` shows
   only the standard library and first-party leaf packages.
   (`hub-clean`, 10)

Seven checks, sum 100: `ch5-parity` 10, `not-deaf` 25,
`replay-is-live` 15, `observer-fires` 15, `wake-once` 15,
`loud-refusal` 10, `hub-clean` 10.

**Yours.** The shape of the mailbox, the structure of the actor loop,
how many goroutines the framework uses. The grader checks the
properties, not the implementation.

**Exercise.** Build a three-agent workflow in `ch06/`. An author
writes a draft, an editor improves it, a reviewer approves it. Each
agent has its own prompt and tools. The framework coordinates them.
A single binary reads JSON events on stdin, writes observations on
stdout, and logs to `CH06_LOG`.

```sh
make grade-dir CH=6 DIR=path/to/yours
make grade6
```

## §6.1 The idea in plain words

Chapter 2 built a renderer that turned the event log into
`history.md`. It watched the stream and produced a view. That is an
observer. The reader who built it has already written one and can now
name it.

A deaf loop processes one request at a time. The model sends back a
tool call, the engine runs the tool, and nothing else can happen
until the tool finishes. A user typing a hint while a three-second
build runs gets no acknowledgment. A second agent waiting for its
turn gets nothing at all. The engine is a for-loop, and a for-loop
has no ears.

The fix is three pieces that work together.

**An observer is a one-way window.** The engine notifies observers
when something happens: a state transition, a completed response, a
streaming chunk. Observers cannot call back into the engine. They
watch and react. A logger is an observer. A GUI is an observer. A
parent managing a child agent is an observer. The observer pattern
itself is nothing new. What matters is that the framework defines the
interface, observers implement it, and nothing inside the framework
knows what the observers do. A GUI, a gateway, and a sub-agent
supervisor all plug in here without any of them changing a line of
framework code.

**A mailbox is a one-way door.** The engine accepts messages through
a queue: prompts, hints, tool completions, interrupts. `Post` never
blocks. The queue is a slice guarded by a mutex with a channel to
signal "something arrived." Prompts, hints, and tool completions
enter through the same door because the loop that processes them
needs to see them in order. A hint that arrives while a tool is
running goes into the queue after the prompt that started the turn
and before the tool completion that ends it. The ordering is the
whole point. A separate queue per message type destroys it.

**An actor loop drains the mailbox.** The engine runs on its own
goroutine. It blocks on the mailbox's signal channel until something
arrives, drains the queue, and processes each message in order. After
processing, it notifies observers. The loop is the smallest possible
concurrency primitive: one goroutine, one queue, one notify step. It
replaces the synchronous for-loop from Chapter 4 with something that
can hear, without introducing locks on any shared state. The only
lock is inside the mailbox itself.

These three pieces are independent additions to the star topology
from Chapter 5. The observer interface and the mailbox message types
go into `internal/common/` as two new interfaces in the hub. The
mailbox implementation goes into a new spoke or an existing one. The
actor loop replaces the engine's synchronous ask. No existing spoke
changes.

## §6.2 The framework that imported its own GUI

CodeRhapsody built the GUI before it had a framework seam. The React
frontend connected to a Go server, the Go server imported the
framework, and the framework reached back into the server for test
scaffolding. `test_support.go` built a real GUI server for the
twenty-three tests that needed a UI attached. The import graph formed
a cycle, and the cycle hid inside convenience: nobody noticed because
every test passed.

The dependency surfaced as a question: can the framework run without
the GUI? No. It compiled without it, but tests could not run.
Untangling it took five commits across two days. The load-bearing
commit was titled *"Test scaffolding takes a UIObserver, not a GUI
server."*

The fix was not deletion. The GUI was demoted to an observer. Instead
of the framework importing the GUI, the framework broadcasts events,
and the GUI subscribes to them. The import arrow reversed direction.
After the fix, the framework does not know the GUI exists, the GUI
is one of several observers, and tests run without building a server.

> The guard that enforces this boundary exempts test files and ratchets
> them downward, on the principle that a guard that fails today guards
> nothing. Twenty-three tests needed a GUI server. Now thirteen do,
> and the ratchet prevents the number from climbing back. The boundary
> found problems, it did not create them: two encapsulation reaches
> appeared as compile errors the moment the `internal/` wall went up,
> one where a tool file grew a method on a job type and another where
> the engine called an unexported function on a job.

The fix's shape is this chapter's thesis. The GUI went from a
component the framework knows about to an observer the framework
broadcasts at. Chapter 5's facade measured who actually uses a
framework: two-thirds of the API surface serves tool authoring, five
percent serves agent lifecycle. Design the observer seam for the
ninety-five percent of the world that is not the engine.

## §6.3 The outbound seam

An observer receives events. It does not ask for them, it does not
call back into the engine, and it does not block. Those three
constraints are the interface:

```go
type Observer interface {
    Observe(Observation)
}
```

`Observe` takes one argument. The argument is a sealed interface:
the framework defines every type that implements it, and external
code cannot add new ones. Three types carry the information:

```go
type PartDelta struct {          // streaming chunk
    PartID uint64 `json:"part_id"`
    Chunk  string `json:"chunk"`
}
```

A `PartDelta` arrives for every streaming chunk the vendor sends.
`PartID` identifies which part is being streamed so that an observer
can assemble the complete response without buffering. `Chunk` is a
string, not `[]byte`, because the event log is JSON and base64
doubles the size of everything.

The streaming rule: do not model streaming as a mode. Model
non-streaming as a stream of length one. A non-streaming vendor emits
one delta plus the finalizer. Observers that care about liveness
render deltas; observers that do not ignore them and act on
finalizers. Adding streaming later therefore adds no new event kinds.

Tool results are the proof the rule is right: they are observable but
never stream, which needs no special case. They are always length one.

Deltas are transient. They are never appended to the durable log;
only finalized parts are recorded. The log is the state. The stream
is the experience. Replaying the log must produce the same context
as the live run, minus the animation. The `replay-is-live` check
tests exactly this: truncate the log at the last `request_sent`,
replay it, and compare byte for byte against what the live engine
sent to the vendor.

```go
type PartFinal struct {          // completed part
    Seq    Seq      `json:"seq"`
    PartID uint64   `json:"part_id"`
    Part   TextPart `json:"part"`
}
```

A `PartFinal` arrives when a part is complete: the full `TextPart` or
`ToolCallPart`, with its sequence number and position. This is what
replay uses. If you can reconstruct the context from `PartFinal`
events alone, the observer seam is complete.

```go
type StateChanged struct {       // turn-state transition
    From   TurnState `json:"from"`
    To     TurnState `json:"to"`
}
```

A `StateChanged` fires on every transition: `Idle` to
`InputPending`, `InputPending` to `InFlight`, `InFlight` to
`ToolsPending`, any terminal state back to `Idle`. The GUI uses this
to show a spinner. The parent uses this to know when a child finished.
The grader uses this to verify the observer seam works.

The observation types carry no agent identity. For a single agent,
that is all you need. For a multi-agent framework, the coordinator
tags agent identity externally when it routes observations. The type
is cheap because external tagging is just a wrapper.

The observer is told, never asked. The engine calls `Observe` after
every action: after recording a response, after transitioning state,
after a tool completes. Observers see the same events whether they
are attached to a running agent or to a freshly started one, because
the events are generated from the same code path that updates the
context. Replay and live are the same sequence, and the
`replay-is-live` check tests exactly that.

### Why Observe must not block

A blocking observer recreates the deafness. The engine calls
`Observe` on the actor goroutine. If `Observe` takes a lock, writes
to a slow network, or waits for a response, the actor goroutine
stalls. The mailbox fills up. Hints arrive and wait. The queue that
was supposed to fix the problem becomes the problem.

If an observer needs to do slow work, it copies the observation and
posts it to its own queue. The observer's goroutine drains its own
queue. The engine never waits.

## §6.4 The inbound seam

The agent is deaf, but the obvious explanation is wrong. It is not
that the tool runs inline. The tool already runs on its own goroutine
since Chapter 4. It is that the engine parks in a `select` awaiting
the tool's result, and that engine is the same loop that would drain
inbound events. The drainer is parked. Hints arrive and sit in a
channel that nobody reads until the tool finishes.

The fix: deliver tool completions INTO the queue instead of awaiting
them in a `select`. The mailbox carries everything the engine can
hear:

```go
type Inbound interface{ inbound() }

type UserMessage   struct{ Text string }
type Hint          struct{ Text string }
type ToolCompleted struct {
    CallID  string
    Result  string
    IsError bool
}
type Interrupt struct{}
```

One queue. One `Post` method. Four message types. The implementation
is a mutex-guarded slice with a buffered channel of capacity one as
the signal:

```go
type Box struct {
    mu    sync.Mutex
    queue []Inbound
    wake  chan struct{}
}

func (b *Box) Post(msg Inbound) {
    b.mu.Lock()
    b.queue = append(b.queue, msg)
    b.mu.Unlock()
    select {
    case b.wake <- struct{}{}:
    default:
    }
}

func (b *Box) Drain() []Inbound {
    b.mu.Lock()
    q := b.queue
    b.queue = nil
    b.mu.Unlock()
    return q
}
```

`Post` never blocks. If the channel is full, the signal is already
pending and the loop will drain the queue. `Drain` returns everything
and clears the queue. The mutex protects the slice; the channel is
just a wake signal. No priorities, no reordering, no cleverness.

### Why one queue

A prompt, a hint, a tool completion, and an interrupt arrive through
the same door because the order they arrived in is the order they
matter in. A hint that arrived during a tool call matters now. A
tool completion that arrived after an interrupt matters never.

Separate queues destroy this ordering. If prompts and hints go to
different channels, a `select` chooses between them, and Go's
`select` is pseudo-random when both are ready. The hint might be
processed before the prompt that it was meant to modify, or after
the tool that it was meant to interrupt. A single FIFO preserves
arrival order, and arrival order is causal order.

### Tool completions enter the mailbox

This is the change that breaks the deafness. In Chapter 4, the engine
calls `Execute(call)` and blocks until the tool finishes. The return
value is the tool result, and the engine records it immediately.

In Chapter 6, the engine dispatches the tool call on a separate
goroutine. When the tool finishes, it posts a `ToolCompleted` to the
mailbox. The engine is back at its mailbox drain loop, free to
process hints and interrupts while the tool runs. The tool result
arrives as a message, not a return value, and the engine processes it
in queue order alongside everything else.

## §6.5 The actor loop

The actor is a goroutine with a mailbox. It blocks on the signal
channel, drains the queue, and processes each message. After
processing, it notifies observers.

```go
func (e *Engine) run() {
    for {
        <-e.box.wake

        for _, msg := range e.box.Drain() {
            switch m := msg.(type) {
            case UserMessage:
                e.startTurn(m.Text)
            case Hint:
                e.applyHint(m.Text)
            case ToolCompleted:
                e.recordResult(m)
            case Interrupt:
                e.interrupt()
                return
            }
            e.notifyObservers()
        }
    }
}
```

The pseudocode above omits error handling, round limits, and the
multi-turn tool loop. The real implementation is longer. The shape
is the same: drain, switch, notify.

**`startTurn`** records the prompt as a `MessageReceived` event,
renders the context, sends it to the vendor, and records the
response. If the response contains tool calls, the engine dispatches
each one on its own goroutine and sets the turn state to
`ToolsPending`. If no tool calls, the turn ends and the state goes
to `Idle`.

**`applyHint`** records the hint as a `HintReceived` event and
attaches it to the pending context. The next vendor request will
include it. The hint does not start a new turn and does not change
the turn state.

**`recordResult`** records the tool result as a `ToolReturned` event,
checks whether all outstanding calls have results, and if so, sends
the accumulated results back to the vendor for the next round.

**`interrupt`** sets the turn state to `Interrupted` and stops
processing. Outstanding tool calls may still be running, but their
results will be discarded when they post to the mailbox of a stopped
engine.

### Ask, rebuilt

The synchronous `Ask(text) (string, error)` from Chapter 4 still
works. Its implementation changes:

```go
func (a *Agent) Ask(text string) (string, error) {
    a.Post(Prompt{Text: text})
    obs, err := a.Wait(context.Background(), func(o Observation) bool {
        sc, ok := o.(StateChanged)
        return ok && sc.To == Idle
    })
    if err != nil {
        return "", err
    }
    return a.LastText(), nil
}
```

Post the prompt. Wait for the state to reach `Idle`. Return the last
agent text. The blocking is in `Wait`, not in the engine. The engine
is free to hear hints while the caller waits.

### Where is the concurrency?

Exactly two goroutines per agent: the actor loop and the current tool
(if any). The mailbox has one lock. The observer list is set at
creation and never modified. There is no shared mutable state between
the actor and the tool except the mailbox itself, and the mailbox
is the one lock.

This is not accidental minimalism. Every lock is a place where two
goroutines disagree about what is happening. Two goroutines can
disagree in testable ways. Ten goroutines with a shared map disagree
in ways that show up in production at 3 AM on a Saturday.

> CodeRhapsody's Chapter 4 grader exposed five real concurrency bugs
> in 120 parallel test runs. The root cause of every one was a parent
> holding stale facts about a child: reading a field on one goroutine
> while the actor goroutine wrote it. The fix each time was the same:
> move the read to the actor goroutine, or make the mailbox the only
> path between them. Two goroutines with one lock is not a style
> preference. It is what survived the race detector.

The goal is the smallest number of goroutines that fixes the
deafness, and that number is two.

## §6.6 The Wait primitive

This chapter does not introduce agent state. It exposes it. The
reader has had a state machine since Chapter 2: `TurnState` with
`Idle`, `InputPending`, `InFlight`, `ToolsPending`. Chapter 4 added
`Interrupted`, deliberately as a state rather than a flag, because
replay re-executes tool calls that were cancelled if `Interrupted`
is not terminal. The state machine is a fact. The observer seam
makes it visible. `Wait` makes it waitable.

```go
func (a *Agent) Wait(ctx context.Context,
    pred func(Observation) bool) (Observation, error)
```

`Wait` blocks until an observation satisfies `pred` or the context
cancels. The implementation registers a temporary observer that
checks every observation against the predicate, and signals a
condition when one matches.

Every higher-level wait derives from this.

**Blocking send**: post a `UserMessage`, then `Wait` for `StateChanged`
where `To` is `Idle` or `Interrupted`.

**Wait for agent**: `Wait` for `StateChanged` where `To` is `Idle`
and `Agent` matches.

**Join**: `Wait` for a predicate that tracks N agent IDs and returns
true when all have reached `Idle`.

**Wake-any**: `Wait` for `StateChanged` where `To` is `Idle` and
`Agent` is any of a set.

The `wake-once` grader check tests this directly. Two agents finish
in the same turn window. The parent waits for both. The observation
stream shows one wakeup that delivers both completions, not two
separate wakeups. If the implementation polls or uses one channel per
agent, the check fails.

## §6.7 Hints and interrupts

A hint is the same event as a prompt, distinguished by turn state.
A message that arrives while the turn is idle starts a new turn. A
message that arrives while the turn is in flight or tools-pending
is a hint. The classification happens in the reducer (`Apply`), not
at the capture site, because only the reducer holds the state that
makes the decision correct.

The `HintReceived` event type was added in Chapter 5's refactoring
for exactly this reason: without it, the reducer cannot distinguish
"a new prompt arrived" from "a hint arrived during an active turn."
Both carry text. The difference is when they arrived relative to
the turn, and the turn state is the reducer's business.

> Bill's ruling on callbacks is absolute: a callback added to break
> a Go dependency cycle is a red flag. The mailbox breaks no cycles.
> It is not a workaround for a dependency the compiler rejected. It
> is the mechanism by which concurrent events enter a sequential
> loop. The engine has one goroutine, one queue, and one notify step.
> If you find yourself adding function-pointer fields to break a
> compile error, the dependency is real, and the fix is to move the
> interface to the hub.

The interrupt is simpler. An `Interrupt` message arrives, the engine
sets the turn state to `Interrupted`, and the loop exits. Tools that
are still running will complete, and their `ToolCompleted` messages will
arrive at a mailbox that nobody is draining. That is fine. A killed
goroutine's output is garbage, and treating it otherwise is a
different bug.

### The stdin protocol

`cmd/main.go` reads stdin on its own goroutine and posts to the
agent's mailbox. The protocol is JSON lines:

```json
{"kind":"prompt","text":"Write a haiku about refactoring"}
{"kind":"hint","text":"Use a metaphor about gardens"}
{"kind":"interrupt"}
```

Backward compatibility with the Chapter 1 format:

```json
{"user":"Write a haiku about refactoring"}
```

A message with a `"user"` key and no `"kind"` key is treated as a
prompt. This keeps every grader from Chapters 1 through 5 working
without changes.

## §6.8 Managing multiple agents

A single-agent framework is a special case. The same engine, the same
mailbox, the same observer seam scales to N agents with one addition:
a coordinator that tracks which agents exist and routes observations.

```go
type Framework struct {
    actors    map[AgentID]*Actor
    host      Host
    merged    chan Observation
}

func (f *Framework) Add(id AgentID, actor *Actor)
func (f *Framework) Remove(id AgentID)
func (f *Framework) WaitAny(ctx context.Context,
    pred func(Observation) bool) (Observation, error)
```

`Add` registers an actor with its own mailbox and its own actor
goroutine. The framework attaches itself as an observer on every
actor it manages, multiplexing their observations into a single
merged channel. `WaitAny` on the framework blocks until any actor
fires an observation that satisfies the predicate.

A Go program constructing three agents and passing messages between
them is not an agent spawning children through its own tool surface.
Multi-agent does not require a sub-agent API. It requires a
framework, a prompt, and tools. This exercise is just Go. The
distinction matters because a sub-agent chapter adds a tool surface
for spawning; this chapter proves it is not necessary.

> A framework that manages agents is a parent. The
> observation stream is how the parent watches its children. An
> agent that spawns another watches it through this same observer
> seam and decides what to do based on what it sees. Nothing needs
> to be added to the observer interface. Everything is already here.
> The alternative is reading the child's history file. One
> supervision call built that way returned 307,984 bytes, roughly
> 13% of a context window, because it had no volume contract. The
> observer seam, delivering events as they happen, returned 3,447
> bytes for the same purpose.

Three agents prove the cost was paid once. The first agent might work
because the framework was tested with it. The second agent might work
because the code was debugged for two. The third agent works because
the framework is general.

## §6.9 Media capabilities

A model that accepts images does not accept audio. A model that
accepts audio does not accept video. A renderer that silently drops
an unsupported part is the Chapter 1 mistake: the model does not see
what the programmer sent, and neither one knows.

`ModelFeatures` declares what a model accepts:

```go
type Media uint8
const (
    MediaImage    Media = 1 << iota
    MediaAudio
    MediaVideo
    MediaDocument
)

type ModelFeatures struct {
    Media Media
}
```

`LookupModel(model) (ModelFeatures, bool)` has no default row.
An unknown model returns `false`, and the renderer refuses loudly:

```
error: model "claude-3-haiku-20240307" does not support audio;
       cannot render BlobPart with Ref{Kind:RefPath, Locator:"recording.mp3"}
```

The error names the model and the media type. The programmer reads
the error and knows what to change. A silent drop would do nothing,
and the programmer would debug the prompt for an hour before
discovering the audio was never sent.

`Ref` replaces the bare path string on `BlobPart`:

```go
type RefKind uint8
const (
    RefPath   RefKind = iota + 1
    RefURI
    RefHandle
)

type Ref struct {
    Kind    RefKind `json:"kind"`
    Locator string  `json:"locator"`
}
```

Three locator kinds cover the three ways content arrives: a local
file path, a remote URI, and a handle to a job's output. The Ref
carries through redaction: when a `BlobPart` is superseded, its `Ref`
survives in the `RedactedPart`, so the content is recoverable by
construction.

## §6.10 Exercise: Author, Editor, Reviewer

Build `ch06/main.go`. Three agents, three roles:

**The author** receives a topic and writes a first draft. Its tools
are `write_draft` (stores text) and `word_count` (returns the count).

**The editor** receives the author's draft and improves it. Its tools
are `read_draft` (retrieves text) and `edit_draft` (replaces text).

**The reviewer** receives the editor's draft and approves or rejects
it. Its tool is `review` (returns accept or reject with notes).

The workflow is sequential: author writes, editor edits, reviewer
reviews. The framework coordinates them through `Post` and `Wait`.
Each agent uses the same vendor (the fake from Chapter 2) with a
different system prompt. The workflow is hardcoded Go. Every new
collaboration pattern is a new Go program. That is the limit this
chapter reaches and the next several chapters work to remove. The
framework works; the rigidity is in the glue code, not in the
framework itself.

Three agents prove the cost was paid once. The first agent might work
because the framework was tested with it. The second might work
because the code was debugged for two. The third works because the
framework is general. If adding the reviewer required a single line
inside the framework package, the seam is wrong.

The binary reads JSON events on stdin, writes observations as JSON
lines on stdout, and logs to the path in `CH06_LOG`. The grader
drives the fake vendor to script specific responses for each agent.

```sh
CH06_LOG=/tmp/ch06.jsonl LLM_VENDOR=fake \
  LLM_BASE_URL=http://localhost:PORT ./ch06
```

## §6.11 What this chapter does not build

Streaming observations arrive as `PartDelta`, but the vendor seam
from Chapter 2 does not stream yet. It returns a complete response in
one block. The `PartDelta` type exists so that when streaming lands,
observers get incremental updates without an interface change.

The `Ref` type supports `RefURI` and `RefHandle` alongside `RefPath`,
but the renderers only handle `RefPath` so far. A vendor that accepts
a URL instead of inlined bytes needs the URI form; a tool whose
output is too large for inline needs the handle form. Both are future
extensions to the renderer, not to the Ref type.

Sub-agent spawning, where one agent creates and supervises another
through the observer seam, uses everything built here and adds
nothing to the framework's interfaces. The observer is already the
parent's view of the child. The mailbox is already the child's
inbox. The Wait primitive is already the join. You now have a hint
channel and nowhere to type into it.

## Taking it for a spin

Run the exercise against the fake vendor and watch the observations
stream on stdout. Three agents start, each with its own state
transitions:

```
{"agent":"author","from":"idle","to":"input_pending"}
{"agent":"author","from":"input_pending","to":"in_flight"}
{"agent":"author","part_id":1,"chunk":"Let me write about "}
{"agent":"author","part_id":1,"chunk":"refactoring..."}
{"agent":"author","seq":3,"part_id":1,"part":{"type":"text","text":"Let me write about refactoring..."}}
{"agent":"author","from":"in_flight","to":"tools_pending"}
...
{"agent":"author","from":"tools_pending","to":"idle"}
{"agent":"editor","from":"idle","to":"input_pending"}
...
{"agent":"reviewer","from":"idle","to":"input_pending"}
...
{"agent":"reviewer","from":"in_flight","to":"idle"}
```

The observation stream is the whole story. No polling, no callbacks,
no reaching into agent internals. The framework tells you what
happened, in order, and you decide what it means. A logger writes
it to a file. A GUI renders it as a chat. A parent agent uses it to
decide when to send the next prompt. The observer seam carries all
three without knowing about any of them.

Now try it with a real vendor. Build the binary and export your
credentials:

```bash
cd agent && go build -o bin ./cmd/
export LLM_API_KEY="your-anthropic-api-key"
export LLM_MODEL="claude-sonnet-4-20250514"
```

`LLM_VENDOR` defaults to `anthropic`. For Gemini, set it to `google`
and point `LLM_BASE_URL` at the Gemini endpoint. For OpenAI, set it
to `openai`.

Start an interactive session. The binary reads one JSON line per
stdin line and streams observations to stdout:

```bash
./bin
```

Type a prompt and press enter:

```
{"kind":"prompt","text":"You are a pirate. Respond only in pirate speak. Tell me about your ship."}
```

Observations stream back as JSON. While the model is still
responding, type a hint on the next line and press enter:

```
{"kind":"hint","text":"Actually, make it a space pirate. Your ship is a starship."}
```

The hint lands mid-turn. Watch the observation stream: a `part_delta`
containing `hint:` appears between the streaming chunks, proving
the agent heard it while the model was still talking. The model
picks it up on its next round and pivots to space piracy.

When you are done, send an interrupt or press Ctrl-C:

```
{"kind":"interrupt"}
```

The agent was deaf. Now it listens.
