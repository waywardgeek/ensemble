# Chapter 5: Two Seams and a Loop

The historical CodeRhapsody incidents below retain the first edition's
measurements. The exercise and ownership rules incorporate the accepted
Ensemble implementation.

The agent framework from Chapter 4 is good enough to write an
author-editor orchestrator. Imagine what it could do with a GUI
built exactly the way you want it. Your layout, your keybindings,
your idea of what coding with AI should feel like. That is where this
book is headed. But the tools that shipped GUI-first paid for it:
the GUI imported the framework, the framework imported the GUI, and
by the time anyone noticed the cycle it was load-bearing.
CodeRhapsody made the same mistake, and refactoring it out took weeks
of focused work. This chapter exists so you do not repeat it. By the
end, your framework will have an API clean enough to host any GUI,
any gateway, any agentic application. Streaming comes next, followed by the GUI.
This one builds the surface it plugs into.

That surface is two seams and a loop. Right now, ask a question while
a tool runs and the agent does not notice until the tool finishes.
Send a hint and it arrives one round too late. Run two agents through one blocking application loop
and the second waits until the first is done. The package
structure is right and the engine is deaf.

This chapter fixes the deafness. Three additions and one retirement:
an outbound seam so the framework tells the world what
happened instead of the world reaching in to ask; an inbound queue so
prompts, hints, tool completions, and interrupts enter through one
door; and a loop between them that drains the queue and notifies
observers on every event. Retire the old synchronous orchestration loop;
keep its public API as an adapter to the actor. Two loops over one engine
are two owners of mutable state, not backward compatibility.
A framework is two seams and a loop. By the end, the same core runs
a three-agent workflow where each agent
has its own tools, its own prompt, and its own observers, and a hint
mid-tool-call arrives before the tool finishes.

## TL;DR

**What you build.** Three things added to Chapter 4's star topology:

1. An **observer** interface in the hub (outbound seam).
2. A **mailbox** queue in the Agent implementation or its own spoke (inbound seam).
3. An **actor loop** owned by the Agent that drains the mailbox and notifies
   observers. Engine remains its single-request transport child.

Plus identity-tagged observations at the existing Ensemble root, media
capabilities alongside the existing `Ref` vocabulary in the hub, and a
three-agent exercise that proves the framework can run completely different agents without changing a line inside it.

**The observer seam.** One interface, three event types. The following
shapes show the shared vocabulary; public aliases expose it to external
clients without importing common:

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
    Seq    uint64   `json:"seq"`
    PartID uint64   `json:"part_id"`
    Part   Part     `json:"part"`
}

type StateChanged struct {       // turn-state transition
    From   string `json:"from"`
    To     string `json:"to"`
}

type AgentID string
```

`Observe` must not block: a slow observer that holds up the loop
recreates the deafness this chapter exists to fix. Agent identity remains
outside each Agent's event log; the framework tags observations externally
when coordinating multiple agents.

**The mailbox.** One inbound queue carries everything:

```go
type Inbound interface{ inbound() }

type UserMessage struct{ Text string }
type Hint        struct{ Text string }
type Interrupt   struct{}
```

A prompt, a hint, a tool finishing, and an interrupt are the same
kind of event to the loop. Public `Post` must not wait for execution or
drop an admitted message. The queue is a mutex-guarded slice with a signal channel;
nothing in it is clever. The worker outcome is an internal message with
its originating turn and full result metadata, not a text-only public shortcut.

**The actor loop.** The Agent owns one conversation goroutine. When the
mailbox has a message, the loop drains it:

- Live `UserMessage` or `Hint`: start a turn when idle; while active, attach
  text to the current turn's pending context.
- Worker report: record the result, check for outstanding calls, continue
  the turn or go idle.
- `Interrupt`: end the active turn, keep the actor receiving input and retain
  real results from jobs that continue.

The synchronous Ask API still exists. It posts a
request with its own buffered reply channel and waits for that request's
text or error. Concurrent blocking requests queue as separate turns;
they are not hints for the active turn. Callers that do not need blocking
use `Post` directly and watch through an observer. One agent owns one
actor, whether a caller uses the blocking or non-blocking API.

**The Wait primitive.**

```go
func (a *agent) Wait(ctx context.Context,
    pred func(Observation) bool) (Observation, error)
```

Blocks until an observation satisfies `pred` or the context cancels.
Use it for observing progress, not acknowledging a blocking request.
The observation stream is bounded and may drop events; waiters can also
consume one another's events. Neither an idle notification nor shared
last-text state identifies which request completed. Request replies are
reliable control messages, separate from this non-blocking observation seam.

**Ref and ModelFeatures.** The existing blob part already carries a `Ref`.
This chapter gives local-path media a real renderer:

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

No default media row. An unknown model returns `false`; attaching media
then fails with an error naming the model and the unsupported media type.
Ordinary text does not require a row in a media capability table. Lookup
policy belongs with rendering; common holds the shared types.
Dropping a part silently is the Chapter 1 mistake: a format the model
cannot read is a lie, not a degradation.

**Rules.** Seven checks.

1. **Chapter 4 still passes.** The actor upgrade changes no observable
   behavior from Chapter 4's star topology. (`ch4-parity`, 10)
2. **The agent hears hints mid-tool.** A slow tool runs for three
   seconds. During that work, a hint arrives. The observer stream shows
   the hint *before* the correlated tool report. A callback may report a
   still-running job; the check does not confuse that report with execution ending. (`not-deaf`, 25)
3. **Replay matches live.** Replay the event-log prefix just before the last
   `request_sent`; rendering that prefix matches the captured live
   vendor request byte for byte. (`replay-is-live`, 15)
4. **Observers fire.** The observation stream contains both state
   transitions (`StateChanged`) and completed content (`PartFinal`).
   (`observer-fires`, 15)
5. **Multiple Agents remain distinguishable.** The application runs author,
   editor and reviewer in sequence. Observations identify their Agent; the
   application advances only after that request actually completes.
   (`multi-agent-observations`, 15)
6. **Unsupported media refused.** Present media the model does not
   support. The renderer returns an error naming the model and the
   media type, not a silent drop. (`loud-refusal`, 10)
7. **Hub stays clean.** `go list -deps` on `internal/common/` shows
   no implementation-spoke dependencies. Shared vocabulary does not import
   its implementations.
   (`hub-clean`, 10)

Seven checks, sum 100: `ch4-parity` 10, `not-deaf` 25,
`replay-is-live` 15, `observer-fires` 15, `multi-agent-observations` 15,
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
go run ./cmd/grade-edition2 -ch 5 path/to/yours
```

The grader finds the `ch06` workflow module inside the submitted core or
beside it; `-workflow path/to/workflow` selects another location. It builds
and runs that actual application, without falling back to the core CLI.
Chapter 4's `-consumer` and `-consumer-case` flags remain available for the
inherited external-consumer check.

Media probes render a local image through each adapter, refuse unsupported
audio and refuse an image on an unknown model. The Anthropic fixture uses the
older Sonnet 4.6 row declared by this chapter’s accepted solution. To test another
implemented capability table, save this grader-only mapping
as `media-models.json` and pass `-media-models media-models.json`:

```json
{"anthropic":"claude-sonnet-4-6","openai":"gpt-5.4-mini","gemini":"gemini-3.8-flash"}
```

The replay check captures real HTTP requests and compares them with rendering
the corresponding pre-request log prefixes. The workflow fake calls `think`
and returns three role texts. Its observations and ordered completion markers
do not, by themselves, establish that every manuscript tool works, stored edits
reach the reviewer, or a saturated progress stream cannot lose request replies.
Keep those required behaviors covered by real public-API integration tests and
source review. Passing a marker check does not waive them.


## §5.1 The idea in plain words

Chapter 2 built a renderer that turned the event log into
`history.md`. It watched the stream and produced a view: an
observer. The reader has already built one and can now name it.

A deaf loop processes one request at a time. The model sends back a
tool call, the engine runs the tool, and nothing else can happen
until the tool finishes. A user typing a hint while a three-second
build runs gets no acknowledgment. A second agent waiting for its
turn gets nothing at all. The engine is a for-loop, and a for-loop
has no ears.

The fix is three pieces that work together.

**An observer is a one-way window.** The Agent notifies observers
when something happens: a state transition, a completed response, a
streaming chunk. Observers cannot call back into the engine. They
watch and react. A logger is an observer. A GUI is an observer. A
parent managing a child agent is an observer. What matters is that
the framework defines the interface, observers implement it, and
nothing inside the framework knows what the observers do. A GUI, a gateway, and a sub-agent
supervisor all plug in here without any of them changing a line of
framework code.

**A mailbox is a one-way door.** The Agent accepts messages through
a queue: prompts, hints, tool completions, interrupts. `Post` does not wait
for execution or a receiver. The queue is a slice guarded by a mutex with a channel to
signal "something arrived." Prompts, hints, and tool completions
enter through the same door because the loop that processes them
needs to see them in order. A hint that arrives while a tool is
running goes into the queue after the prompt that started the turn
and before the tool completion that ends it. The ordering is the
whole point. A separate queue per message type destroys it.

**An actor loop drains the mailbox.** The Agent runs its actor on its own
goroutine. It blocks on the mailbox's signal channel until something
arrives, drains the queue, and processes each message in order. After
processing, it notifies observers. The loop is the smallest possible
concurrency primitive: one goroutine, one queue, one notify step. It
replaces the synchronous for-loop from Chapter 4 with something that
can hear. The mailbox synchronizes delivery; the job subsystem synchronizes
worker state. Neither gives workers permission to mutate the conversation.

These three pieces are independent additions to the star topology
from Chapter 4. The observer interface and public mailbox input types
go into `internal/common/`; private worker messages stay with their owner. The
mailbox implementation goes into a new spoke or an existing one. The
Agent-owned actor replaces the engine's synchronous ask while preserving the public
blocking API. Internal orchestration helpers are retired, not retained as a
second execution path.

## §5.2 The framework that imported its own GUI

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
broadcasts at. The first-edition facade measured who actually used its
framework: two-thirds of the API surface serves tool authoring, five
percent serves agent lifecycle. Design the observer seam for the
ninety-five percent of the world that is not the engine.

## §5.3 The outbound seam

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
can assemble the complete response without confusing concurrent parts. `Chunk` is a
string because it is text. A `[]byte` JSON field would base64-encode it,
adding about a third to the byte count before JSON overhead.

The streaming rule: do not model streaming as a mode. Model
non-streaming as a stream of length one. A non-streaming vendor emits
one delta plus the finalizer. Observers that care about liveness
render deltas; observers that do not ignore them and act on
finalizers. Adding streaming later therefore adds no new event kinds.

Tool results are the proof the rule is right: they are observable but
never stream, which needs no special case. They are always length one.

Deltas are transient. They are never appended to the durable log;
finalized content and other conversation facts are recorded. The log is the state. The stream
is the experience. Replaying the log must produce the same context
as the live run, minus the animation. The `replay-is-live` check
tests exactly this: replay the prefix just before the last `request_sent`
and compare the rendered request byte for byte with the captured request.
The marker records delivery and consumes pending ephemera, so including it
would reconstruct the state after that request was sent.

```go
type PartFinal struct {          // completed part
    Seq    uint64   `json:"seq"`
    PartID uint64   `json:"part_id"`
    Part   Part     `json:"part"`
}
```

A `PartFinal` arrives when a part is complete: the full neutral part,
including text, a tool call or a tool result, with its sequence number
and display identity. A text-only field loses the call arguments and
correlation that an observer needs. Durable replay uses the event log,
which also contains prompts, hints and request boundaries; the lossy
observation stream cannot replace it.

```go
type StateChanged struct {       // turn-state transition
    From   string `json:"from"`
    To     string `json:"to"`
}
```

A `StateChanged` fires on every transition: `Idle` to
`InputPending`, `InputPending` to `InFlight`, `InFlight` to
`ToolsPending`, any terminal state back to `Idle`. The GUI uses this
to show a spinner. An application uses this to display progress; request completion tells it
when a particular submitted turn finished.
The grader uses this to verify the observer seam works.

The observation types carry no agent identity. For a single agent,
that is all you need. For a multi-agent framework, the coordinator
tags agent identity externally when it routes observations. The type
is cheap because external tagging is just a wrapper.

The observer is told, never asked. The Agent calls `Observe` after
every action: after recording a response, after transitioning state,
after a tool completes. Observers receive future events from the same code path that updates
the context. Attaching late does not replay earlier observations; history
remains the separate source for past facts.

### Why Observe must not block

A blocking observer recreates the deafness. The Agent calls
`Observe` on the actor goroutine. If `Observe` takes a lock, writes
to a slow network, or waits for a response, the actor goroutine
stalls. The mailbox fills up. Hints arrive and wait. The queue that
was supposed to fix the problem becomes the problem.

If an observer needs to do slow work, it copies the observation and
posts it to its own queue. The observer's goroutine drains its own
queue. The actor never waits for that slow work.

## §5.4 The inbound seam

The agent is deaf, but the obvious explanation is wrong. It is not
that the tool runs inline. The tool already runs on its own goroutine
since Chapter 4. It is that the conversation loop parks in a `select` awaiting
the job report, and that is the same loop that would drain
inbound events. The drainer is parked. Hints arrive and sit in a
channel that nobody reads until the tool finishes.

The fix: deliver tool completions INTO the queue instead of awaiting
them in a `select`. The mailbox carries everything the Agent can
hear:

```go
type Inbound interface{ inbound() }

type UserMessage   struct{ Text string }
type Hint          struct{ Text string }
type Interrupt struct{}
```

One queue carries public input and private worker messages. The small
mailbox sketch below uses `any` for those internal messages; public `Post`
accepts only the public input vocabulary. The implementation
is a mutex-guarded slice with a buffered channel of capacity one as
the signal:

```go
type Box struct {
    mu    sync.Mutex
    queue []any
    Wake  chan struct{}
}

func (b *Box) Post(msg any) {
    b.mu.Lock()
    b.queue = append(b.queue, msg)
    b.mu.Unlock()
    select {
    case b.Wake <- struct{}{}:
    default:
    }
}

func (b *Box) Drain() []any {
    b.mu.Lock()
    q := b.queue
    b.queue = nil
    b.mu.Unlock()
    return q
}
```

`Post` does not wait for a receiver. Its brief mutex protects enqueueing;
it is not a lock-free guarantee. If the channel is full, the signal is already
pending and the loop will drain the queue. `Drain` returns everything
and clears the queue. The mutex protects the slice; the channel is
just a wake signal. The queue preserves arrival order and nothing else.

### Why one queue

A prompt, a hint, a tool completion, and an interrupt arrive through
the same door because the order they arrived in is the order they
matter in. A hint that arrived during a tool call matters now. A
tool completion that arrived after an interrupt still contains real output.
Its former turn is over; its job may still be alive.

Separate queues destroy this ordering. If prompts and hints go to
different channels, a `select` chooses between them, and Go's
`select` is pseudo-random when both are ready. The hint might be
processed before the prompt that it was meant to modify, or after
the tool that it was meant to interrupt. A single FIFO preserves
arrival order, and arrival order is causal order.

### Tool completions enter the mailbox

This is the change that breaks the deafness. In Chapter 4, the Agent
dispatcher waits for the job report. The wait has a limit, but the same
conversation loop cannot process human input while it is parked.

In Chapter 5, workers execute the tool and wait for its report.
They post their outcomes to the mailbox. The Agent is back at its
mailbox drain loop, free to
process hints and interrupts while the tool runs. The tool result
arrives as a message, not a return value, and the actor processes it
in queue order alongside everything else.

## §5.5 The actor loop

The actor is the Agent's conversation goroutine with a mailbox. It blocks on the signal
channel, drains the queue, and processes each message. After
processing, it notifies observers.

```go
func (r *runtime) run() {
    for {
        <-r.owner.box.Wake
        for _, message := range r.owner.box.Drain() {
            switch m := message.(type) {
            case common.UserMessage:
                r.input(m.Text)
            case common.Hint:
                r.input(m.Text)
            case response:
                r.receive(m)
            case report:
                // Record the real result, including reports from older turns.
                r.recordReport(m)
            case common.Interrupt:
                r.interrupt(common.ErrInterrupted)
                // Keep draining: interruption does not end the actor.
            }
        }
    }
}
```

This is schematic: `runtime` belongs to the Agent, and `recordReport`
stands for its report-handling branch. The excerpt omits request replies,
shutdown, error handling and round limits. Notifications follow successful
reduction, so observers see stored facts. The shape remains drain, switch,
notify; method names are not the exercise contract.

**Starting a turn** records the prompt as a `MessageReceived` event,
renders the context and records the request boundary before handing
HTTP to a worker. That worker returns response facts through the mailbox.
A synchronous HTTP call on the actor would merely move the deafness from
tools to the network. Tool calls enter `ToolsPending`; dispatch remains
serial within their batch, with each wait performed outside the actor.
When no further calls or pending hints remain, the turn can finish.

**Applying a hint** records the hint as a `HintReceived` event and
attaches it to the pending context. The next vendor request will
include it. A hint arriving during HTTP cannot rewrite bytes already sent;
record the request boundary before releasing the transport worker. The hint does not start a new turn and does not change
the turn state.

**Recording a result** records the tool result as a `ToolReturned` event,
checks whether all outstanding calls have results, and if so, sends
the accumulated results back to the vendor for the next round.

**`interrupt`** ends the active turn and cancels its in-flight model
request. The Agent remains usable, and other jobs continue. Pair outstanding
calls honestly: a wait was interrupted, not necessarily its job. Refuse calls
that were never dispatched. A later real report must not become a duplicate
vendor result or finish a newer turn. The accepted implementation retains it
as `job_report`, with the bytes still in the job output file.

### Ask, rebuilt

The synchronous Ask API from Chapter 4 still
works; the accepted public method also takes a caller context. It delegates to the agent's owned actor:

```go
func (a *agent) Ask(ctx context.Context, text string) (string, error) {
    reply := a.Submit(ctx, text)
    select {
    case result := <-reply:
        return result.Text, result.Err
    case <-ctx.Done():
        return "", ctx.Err()
    }
}
```

`Submit` creates a one-element buffered reply channel, posts the request,
and returns that channel. `Ask` waits for its reply, caller cancellation,
or actor shutdown. The actor continues processing its mailbox. The buffer
lets the actor finish even if the caller has cancelled.
The actor returns that request's text and error, not shared last-text state.
When persistence is configured, save before replying; a failed save is a
failed completion, not durable success. The accepted library's empty `LogPath`
leaves persistence with its embedding caller. Notify observers of the same outcome without making observation
delivery a prerequisite for returning to the caller.

The excerpt shows the request-specific reply; production also handles
shutdown without losing an already completed reply. Starting the actor is lazy
and idempotent. Repeated callers use the same Agent-owned mailbox consumer;
an additional public Actor wrapper is unnecessary.
Shutdown stops new work, cancels in-flight model requests, stops managed
processes, joins workers, records their outcomes and saves. An arbitrary Go
tool may not be cancellable: shutdown must wait for it, not discard its real
result and pretend it stopped. Repeated shutdown is safe; later requests
return a distinguishable stopped error.

Delete `Engine.Ask` and any sibling orchestration loop after migrating their
callers, including legacy CLI prompt formats and embedding programs. Keep
one-request rendering and transport in Engine and reduction/persistence
in History. The Agent actor orders multi-round orchestration, turn setup,
runtime context mutation and save-before-reply. Workers send facts through
the mailbox; they do not append directly to live History. The synchronous API survives; the duplicate implementation does not.

The same ownership rule applies to the tool-round limit. A positive N permits
N dispatched tool batches in one human turn. Zero selects the named default
of 200, not infinity. Read the effective setting once at turn start; chapter
8's persisted setting must feed this same policy. After the Nth batch, allow
the model to consume its results and finish. If it asks for another batch,
execute none of it, record explicit not-executed results for every call and
return a distinguishable limit error. Do not leave dangling calls or edit
signed assistant content. A later instruction can start a new turn.

### Where is the concurrency?

One actor goroutine owns the engine. A tool execution worker runs the
handler; a separate job-report worker waits for completion, the callback
deadline or an output pattern. These must be separate: a blocked handler
must not prevent the model from receiving a running-job report. A callback
is not cancellation. Both workers belong to the shutdown join.

Workers return results and context effects through the mailbox rather than
mutating the engine. The job subsystem synchronizes process state, output
and completion; the actor owns the conversation. Count ownership boundaries,
not a fixed number of goroutines.

The minimalism is deliberate. Every lock is a place where two
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

That historical two-goroutine fix established the ownership boundary.
The job-report worker preserves it while allowing callbacks before a tool
returns. The goal is still minimal concurrency, with one execution owner
and explicit messages across every boundary.

## §5.6 The Wait primitive

This chapter does not introduce agent state. It exposes it. The
reader has had a state machine since Chapter 2: `TurnState` with
`Idle`, `InputPending`, `InFlight`, `ToolsPending`. This chapter adds
`Interrupted`, deliberately as a state rather than a flag, because
the interrupted turn must not resume dispatch when a late result arrives.
Replay reduces recorded facts; it does not execute tools. The state machine is a fact. The observer seam
makes it visible. `Wait` makes it waitable.

```go
func (a *agent) Wait(ctx context.Context,
    pred func(Observation) bool) (Observation, error)
```

`Wait` blocks until an observation satisfies `pred` or the context
cancels. The current actor consumes a bounded observation channel; it is
not a subscription replaying every event to every waiter.

**Blocking send** uses the request-scoped completion described above, not
`Wait` for a shared idle transition.

**Wait for agent**, **join**, and **wake-any** can use observations to watch
agent state. But an orchestration decision that requires completion cannot
rest on a lossy progress stream. Track the relevant request completions
reliably, then publish the resulting progress. A shared idle event cannot
acknowledge a particular prompt, and a slow observer must not stall the
actor just to prevent a dropped progress notification.

The original check named `wake-once` verified observations and turn-ended
records from distinct Agents. It did not measure wakeup count or simultaneous
completion. The Edition 2 label is `multi-agent-observations`. This application
coordinates stages sequentially; reliable request completion tells it when
to pass a result to the next Agent. No grouped-results API is required.

## §5.7 Hints and interrupts

A hint is the same event as a prompt, distinguished by turn state.
A message that arrives while the turn is idle starts a new turn. A
message that arrives while the turn is in flight or tools-pending
is a hint. Classification belongs to the Agent
conversation owner using the current reduced state, not the stdin reader
whose view can already be stale. Blocking `Ask`/`Submit` calls deliberately
queue separate requests instead of becoming hints.

The `HintReceived` event type exists for exactly this reason:
without it, the reducer cannot distinguish
"a new prompt arrived" from "a hint arrived during an active turn."
Both carry text. The difference is when they arrived relative to
the turn, and the turn state is the reducer's business.

> Bill's ruling on callbacks is absolute: a callback added to break
> a Go dependency cycle is a red flag. The mailbox breaks no cycles.
> It is the mechanism by which concurrent events enter a sequential
> loop, not a workaround for a dependency the compiler rejected.
> The Agent owns the conversation goroutine, queue and notification step.
> If you find yourself adding function-pointer fields to break a
> compile error, the dependency is real, and the fix is to move the
> interface to the hub.

An `Interrupt` ends the active turn, not the actor's lifetime. The actor
continues draining its mailbox and recording results. A Go handler may
still complete after the interrupt; its real output is not garbage. An
explicit job kill has different semantics: it suppresses late output for
that killed job. Shutdown must stop managed processes, join workers and
save their outcomes rather than silently applying kill semantics to every
non-cancellable Go function.

### The stdin protocol

The CLI reads stdin on its own goroutine and posts to the
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
separate request through the blocking adapter. This preserves the older
protocol's request/reply behavior. Explicit live `prompt` or `hint` messages
use current turn state; they can become hints while the Agent is busy.

## §5.8 Managing multiple agents

A single-agent framework is a special case. The same engine, the same
mailbox, the same observer seam scales to N agents with one addition:
a coordinator that tracks which agents exist and routes observations.

```go
// Ensemble already owns the peer Agents; it is not a second coordinator root.
type AgentObservation struct {
    Agent       Agent
    Observation Observation
}

func (e *Ensemble) WaitAny(ctx context.Context,
    pred func(AgentObservation) bool) (AgentObservation, error)
```

Each Agent has its own mailbox and execution loop. It reports observations
through its Ensemble parent, which supplies the source identity in an envelope.
`WaitAny` consumes one matching observation from that merged progress stream.
It does not gather a group of request results. The application retains each
request's own completion for decisions that must not be lost.

A Go program constructing three agents and passing messages between
them is not an agent spawning children through its own tool surface.
Multi-agent requires nothing beyond a framework, a prompt, and
tools. This exercise is just Go. The
distinction matters because a sub-agent chapter adds a tool surface
for spawning; this chapter proves it is not necessary.

> A framework that manages agents is a parent. The
> observation stream is how the parent watches its children. A
> future agent supervisor can use progress from this seam alongside reliable
> completion. Its spawning and waiting contract is not defined by this exercise.
> The alternative is reading the child's history file. One
> supervision call built that way returned 307,984 bytes, roughly
> 13% of a context window, because it had no volume contract. The
> observer seam, delivering events as they happen, returned 3,447
> bytes for the same purpose.

The third agent is the proof. If adding it requires a single
change inside the framework package, the seam is incomplete.

## §5.9 Media capabilities

Accepting images does not imply accepting audio. Accepting audio does
not imply accepting video. A renderer that silently drops
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
An unknown model returns `false`. If a blob needs rendering, the renderer
refuses rather than guessing its support; text alone is still allowed. For
an unsupported audio attachment, the diagnostic names both facts:

```
error: model "claude-3-haiku-20240307" does not support audio;
       cannot render BlobPart with Ref{Kind:RefPath, Locator:"recording.mp3"}
```

The error names the model and the media type. The programmer reads
the error and knows what to change. A silent drop would do nothing,
and the programmer would debug the prompt for an hour before
discovering the audio was never sent.

The blob part uses the `Ref` already carried by the event log:

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
survives in the `RedactedPart`, so the address survives. This does not guarantee that an external file
or job output remains available forever. Local-path rendering reads the actual
file bytes into the vendor attachment; preserving a locator alone is not media
support. URI and handle attachment rendering remain deferred.

## §5.10 Exercise: Author, Editor, Reviewer

Build `ch06/main.go` in a separate Go module importing the public core,
as the Chapter 4 consumer did. Three agents, three roles:

**The author** receives a topic and writes a first draft. Its tools
are `write_draft` (stores text) and `word_count` (returns the count).

**The editor** receives the author's draft and improves it. Its tools
are `read_draft` (retrieves text) and `edit_draft` (replaces text).

**The reviewer** receives the editor's draft and approves or rejects
it. Its tool is `review` (returns accept or reject with notes).

The accepted example uses string `text` arguments for `write_draft` and
`edit_draft`, empty objects for `word_count` and `read_draft`, and string
`decision` (`accept` or `reject`) plus `notes` for `review`. Its word count
counts whitespace-separated words in the stored manuscript. These are small
application-level choices to document in the tool schemas, not new framework
builtins. The stored draft is shared application data; synchronize access
because a tool worker can outlive the turn that launched it.

The slow-tool fixture also calls `think` with this argument object:

```json
{"seconds":3,"thought":"composing draft"}
```

It waits the numeric number of seconds and returns the thought. Register it
only on the author in this application; it is not an Ensemble builtin or a
substitute for the actual manuscript tools.

The workflow is sequential: author writes, editor edits, reviewer
reviews. The application advances through reliable `Ask` or `Submit` completions.
`Post` delivers live hints; `Wait` watches progress.
Each agent uses the same vendor (the fake from Chapter 2) with a
different system prompt. The workflow is hardcoded Go. Every new
collaboration pattern is a new Go program. That is the limit this
chapter reaches; agent-created collaboration remains later work. The
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

The reference workflow is in
[`solutions/ch05/ch06/`](../solutions/ch05/ch06/).
Keep the historical `ch06` application directory and `CH06_LOG` name for
compatibility. Build that module separately. The local provider fake supplies
the ordinary Anthropic adapter environment (`LLM_VENDOR=anthropic`, a model,
a dummy `LLM_API_KEY`, and its loopback `LLM_BASE_URL`), not a new vendor enum.
The grader starts the application without arguments in a disposable directory.
It sends a prompt, a hint addressed to the author during `think`, then EOF:

```json
{"kind":"prompt","text":"Write about the ocean"}
{"kind":"hint","text":"make it vivid","agent":"author"}
```

Emit observations with the `observation` discriminator and external `agent`
tag. The kinds are `state_changed`, `part_delta` and `part_final`; their
payload fields are shown above. A hint acknowledgment is a delta whose
`chunk` begins `hint:`. After each actual request completes, the application
emits its completion record and stage marker:

```json
{"observation":"turn_ended","agent":"author"}
{"pipeline":"author_done"}
```

Use the corresponding editor/reviewer names for their stages. These extra
records belong to the application protocol, not new sealed core observation
types. Log them and the observations to `CH06_LOG`; keep each Agent's durable
conversation history separately so one Agent's sequence numbers do not become
another's. EOF waits for the submitted workflow before shutdown.

## §5.11 What this chapter does not build

Streaming observations arrive as `PartDelta`, but the vendor seam
from Chapter 2 does not stream yet. It returns a complete response in
one block. The `PartDelta` type exists so that when streaming lands,
observers get incremental updates without an interface change.

The `Ref` type supports `RefURI` and `RefHandle` alongside `RefPath`,
but the renderers only handle `RefPath` so far. A vendor that accepts
a URL instead of inlined bytes needs the URI form; a tool whose
output is too large for inline needs the handle form. Both are future
extensions to the renderer, not to the Ref type.

Sub-agent spawning and its supervision contract remain future work.
Today the Go application constructs peer Agents and passes results between
them. These seams provide useful foundations without claiming that a lossy
`Wait` supplies every future join contract. You now have a hint
channel and a terminal to type into. Streaming and the GUI still lie ahead.

## Taking it for a spin

Run the exercise against the fake vendor and watch the observations
stream on stdout. Three agents have their own state transitions. The
following abridged shape is illustrative, not a captured provider transcript:

```json
{"observation":"state_changed","agent":"author","from":"idle","to":"input_pending"}
{"observation":"state_changed","agent":"author","from":"input_pending","to":"in_flight"}
{"observation":"part_delta","agent":"author","part_id":1,"chunk":"A draft about the ocean."}
{"observation":"part_final","agent":"author","seq":3,"part_id":1,"part":{"type":"text","text":"A draft about the ocean."}}
{"observation":"turn_ended","agent":"author"}
{"pipeline":"author_done"}
```

The editor and reviewer produce their own tagged progress and completion
records after their requests finish. The non-streaming response produces a
whole-part delta and finalizer; there are no live token chunks yet.

The observation stream is the display story; request replies carry completion.
The framework tells you
what happened, in order, and you decide what it means. A logger writes
it to a file. A GUI renders it as a chat. The application displays progress
and uses the separate request reply
to decide when to send the next prompt. The observer seam carries all
three without knowing about any of them.

Now try it with a real vendor. Build your own core from its module root:

```sh
go build -o /tmp/ch5agent .
```

Use the credential procedure from Chapter 0: configure the intended child's
`LLM_VENDOR`, discovered `LLM_MODEL`, and `LLM_API_KEY` without printing keys
or putting them in command arguments. Use a scratch working directory for
jobs. `CH02_LOG` selects the core event log; the separate workflow uses
`CH06_LOG`. Start the actual human interface:

```sh
CH02_LOG=/tmp/ch5-chat.jsonl /tmp/ch5agent chat
```

Type ordinary text. Ask for a bounded job such as a shell sleep followed by
printing a marker. While it runs, type a hint. The hint acknowledgment should
arrive before the tool finishes. Type `/interrupt`, then a new question.
The new turn should answer while the older job continues. Inspect its retained
output afterward. Ctrl-C also interrupts the current turn; Ctrl-D ends the
session and runs shutdown. An interrupt is not a way to quit the Agent.

The accepted implementation did this on October 9 with
`claude-haiku-5-5`. A 12-second shell job was still running when the CLI
acknowledged `Hint: remember the marker COMET.` at 22:45:22.121672 UTC.
Interrupt followed at 22:45:22.172780. The new turn answered `COMET` at
22:45:23.081200; the job's real report arrived at 22:45:34.088930 with
`REAL_LATE_RESULT`. The interface was usable roughly eleven seconds before
the old work finished. That is the point of keeping the actor alive.

The [retained CLI transcript](../solutions/evidence/ch06/live-cli-transcript.json)
and event log preserve one correction: the initial implementation emitted
`job_finished` with status `running` when the shell handler returned after
launch. The final guard emits that event only for a terminal job snapshot.
The real late output in the demonstration was correct; the record was kept
and no paid run was repeated for the metadata repair.

Build the three-role application from its own module root:

```sh
go build -o /tmp/ch5workflow .
```

Run it from a scratch directory with the same provider setup, then supply
one topic as JSON. All three roles use the same vendor and different prompts:

```sh
CH06_LOG=/tmp/ch5-workflow.jsonl /tmp/ch5workflow
```

```json
{"kind":"prompt","text":"Write about the ocean at dawn"}
```

EOF waits for this workflow. An interruption of a role returns that request's
error rather than advancing with an incomplete draft. The historical example
module remains named `ch06`; the book's new chapter number does not rename
its source or evidence paths.

The October 9 Gemini run used `gemini-3.8-flash` for all three roles. The
author called `write_draft`, `word_count`, and `think`. Asked for 60 words,
it first stored 59, saw the actual count, added the word `night`, rewrote the
draft and counted 60. The editor called `read_draft` and `edit_draft`; the
reviewer called `review` and accepted the
edited draft. The reviewer received the stored edited text, not merely a
claim that an editor had run. The full tagged observations and three separate
Agent logs remain in [the workflow evidence](../solutions/evidence/ch06/live-workflow-transcript.txt).

Media received its own, different demonstration. A separate public-API client
sent a local 96×96 red PNG to `gpt-5.4-mini`, which answered `red`. The same
renderer refused an audio part with `model "gpt-5.4-mini" does not support audio
media`. These [media receipts](../solutions/evidence/ch06/live-media-transcript.txt)
show actual image consumption and a local refusal; audio, video and PDF
positive cases were checked deterministically at the rendering boundary,
not presented as additional live model runs.

| Retained run | Input tokens | Output tokens |
| --- | ---: | ---: |
| Anthropic human CLI | 4,431 | 177 |
| Gemini author | 5,695 | 1,157 |
| Gemini editor | 1,310 | 591 |
| Gemini reviewer | 663 | 324 |
| OpenAI image client | 63 | 4 |

All recorded cache-read and cache-write counts were zero. This was one
feature-specific demonstration per vendor, not every feature on every vendor.
Original grading remained **80/100 FAIL**: inherited logger-shape parity and
a text-only unknown-model refusal probe failed. Independent review accepted
those documented exceptions; neither was evidence that unsupported media had
been silently dropped.

One review mistake also belongs in the lesson. The initial reviewer demanded
grouped completion from the misleading one-wakeup paragraph, and the student
agreed before stopping at the undefined contract. Bill withdrew that requirement.
A real correction remained: changing Gemini's `mimeType` key survived a test
that only looked for bytes somewhere in the request. The repaired test decoded
the actual attachment fields and caught both wrong MIME and wrong data keys.
A mutation can reveal a weak test; an agreeable reviewer cannot settle missing
scope. The [chapter record](../docs/edition-2-notes/ch06.md) preserves both
corrections and the original results.

The agent was deaf. Now it listens.
