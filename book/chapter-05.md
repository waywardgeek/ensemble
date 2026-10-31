# Chapter 5: Two Seams and a Loop

The historical CodeRhapsody incidents below retain the first edition's
measurements. The exercise and ownership rules below describe the C# implementation.
Historical Go runs are not evidence that this endpoint used a real provider.

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

Give each Agent one mailbox and one asynchronous loop that owns its conversation.
Callers send typed requests; model and tool workers return immutable outcomes.
Only the Agent applies those outcomes to History and decides what starts next.
A private completion task reliably completes each admitted request. A separate,
bounded observer stream reports progress without being allowed to stall work.

Pause stops scheduling, not background jobs. Interrupt ends the current turn,
not the Agent. Shutdown closes admission, drains workers and jobs, then destroys
the owning relationships. The old synchronous `SendAsync` becomes a convenience
wrapper over this same actor; it must not remain a second implementation.

The actor also gives local media and a small multi-Agent workflow one consistent
entry point. They are clients of the framework, not reasons to add a second
orchestrator in the CLI. Common declares the actual C# types, relationships, and
operations; partial implementations provide their bodies in the functional
modules.

## Shared contracts

The library remains one compiled C# assembly. Common declares the actual partial
objects, their persistent fields, and every operation used across modules.
Agent supplies orchestration; Engine renders and performs one frozen HTTP
exchange; History owns recorded facts and reduction; Jobs owns execution.
Neither the CLI nor a GUI belongs in Common or in the Agent's loop.

The public entry points have different promises:

| Operation | Promise |
| --- | --- |
| `SubmitAsync(parts)` | Queue a distinct request; return its reliable final text or error. |
| `SendAsync(text, cancellation)` | Submit text and wait; cancelling this wait does not cancel the turn. |
| `PostAsync(input)` | Acknowledge admission of a live message, hint, or control. |
| `DispatchAsync(call, cancellation)` | Dispatch a host tool through the actor; return its reliable report. |
| `RunAsync(input, cancellation)` | Sequential channel-based facade over the same actor. |
| `WaitAsync(predicate, cancellation)` | Consume matching lossy Agent progress. |
| `FlushAsync(cancellation)` | Wait for foreground quiescence without closing admission. |
| `ShutdownAsync()` | Close admission, cancel and join owned work; do not resurrect the actor. |

`AgentInput` is an immutable value carrying `Kind` and `Text`. `InputKind` contains
`Message`, `Hint`, `Pause`, `Resume`, `InterruptNow`, `InterruptAfterTool`, and
`Ephemeral`. A live message while busy becomes input to the current turn;
concurrent submissions remain distinct queued requests. The private mailbox
also carries HTTP/tool/control completions and lifetime operations.

`Observation` is a value with kind, event sequence, optional part, text,
state transition and nullable admission outcome. Kinds are `Input`,
`StateChanged`, `PartFinal`, `TurnEnd`, and `Error`. Each public submission,
live input and direct dispatch publishes exactly one accepted/refused input
outcome, including validation failures. A request's eventual completion is a
separate fact. Progress publication is nonblocking and bounded; the reliable
request task does not depend on progress being read.

Ensemble owns named or anonymous Agents and a merged progress stream.
`AgentObservation` contains the actual Agent reference and its observation.
`FindAgent(name)` searches the root's actual children, with ordinal unique names
within that root. `WaitAnyAsync` watches merged progress, not a group of results.
Each Agent retains independent History, Engine, Tools, Jobs and mailbox.

`BlobRef` and `ModelFeatures` are value-like declarations. Exact selected-model
capabilities govern local-file media projection. Logs retain locators; requests
inline bytes. Unknown models may send text, but not undeclared attachments.
URI and job-handle rendering remain deferred. See §5.9 for the one authoritative
capability policy.

**Verification is scoped.** The full retained Go grader includes Go consumer
imports and dependency checks. It is not a C# architecture test. For this
chapter's current executable contract, build the CLI and supply its executable
(or a wrapper that forwards arguments) to the explicit Edition 3 route:

```sh
EN_STARTUP_SKILLS=code-tools go run ./cmd/grade-edition3 -ch 5 -executable-only \
  -media-models /your/owned/media-models.json -json /your/owned/build/ch04
```

This is a **65-point** report: retained replay checks 15, retained media checks
10, and four real-CLI scenarios worth 10 each: early hints, already-aged hints,
Interrupt `now`, and Interrupt `after_tool`. The hint scenarios run gated shell
jobs plus a pending file write through a local fake provider. They inspect actual
effects, original call/result pairing, distinct running handles and subsequent
completion through those handles. They check current and remaining observation
deadlines, repeated hints, a shorter wait, a fast write, and a fresh next-batch
observation that must not inherit the cap. Interrupt scenarios require the
opposite tail behavior: no undispatched effects, explicit not-executed results,
and continued access to the already-admitted job. Timing windows allow ordinary
local scheduling jitter; use an otherwise quiet machine when investigating a
timing failure. These checks require a POSIX shell and local loopback HTTP.

For C#, `EN_STARTUP_SKILLS=code-tools` enables the execution tool used by the
retained replay fixture. Run in an isolated home without provider credentials.
The hint fixtures supply their own minimal environment and dummy provider key.
The model mapping selects fixture models; it does not add production aliases or
prove live capability:

```json
{"anthropic":"claude-3-haiku-20240307","openai":"gpt-4o","gemini":"gemini-2.5-flash"}
```

The unchanged Edition 2 `-ch 5 -executable-only` route still reports only its
historical 25 replay/media points. Default Edition 3 Chapter 5 grading retains
the historical Go consumer/workflow route; `-executable-only` deliberately selects
the language-neutral CLI checks and refuses those source-artifact flags.

**The remaining 35 points are not awarded by this executable route.** Its score
is not an architecture or full workflow certificate. Separately test held/frozen
HTTP, late-response refusal after Interrupt, pause, reliable replies under
observation loss, shutdown joins, complete ownership operations, and the real
external C# workflow. Review shared observation behavior for `wait_for_job` and
`send_input`, literal-next-call `tool_limits`, hints racing observer admission,
and the mandatory-completion host-adapter exception. Inspect Common and the
functional call paths; compilation alone cannot enforce the spoke dependency
rule. A marker check cannot prove that stored edits reached the reviewer or that
every manuscript tool ran. Keep those exercises and source review alongside—not
inside—the score.

## §5.1 The idea in plain words

Chapter 2 separated recorded facts from derived conversation context and vendor
requests. Chapter 4 lets a command keep running while the model receives a job
handle. But the Agent itself still follows a synchronous loop: ask the model,
wait for a tool, ask the model again. A second caller cannot reliably say “pause”
or add useful information while that loop is waiting.

The fix is not to scatter locks around that loop. Give conversation decisions
one owner, with three pieces that work together.

**An observer is a one-way window.** The Agent publishes state transitions,
completed parts, and progress. A display or logger watches them without owning
the conversation. Slow observation must not stop work. A bounded delivery queue
may discard a progress update; it must never be the only path by which a waiting
caller receives its answer. Each admitted request has its own reliable completion task
and exactly one terminal outcome.

**A mailbox is a one-way door.** Prompts, hints, controls, and worker completions
enter one ordered queue. In C# a Channel provides the ordinary asynchronous queue
and wake-up behavior. Posting does not wait for the tool or model to finish.
The queue synchronizes delivery; it does not give a sender permission to edit
History. A hint arriving during a request cannot rewrite the bytes already sent,
but it can affect the next safe scheduling decision.

**An actor loop drains the mailbox.** The Agent makes short decisions: record an
input, freeze a model request, admit a tool, apply a returned result, or settle a
turn. Work that waits on the network or a process runs outside that loop. Its
completion is another message, not a worker-side History write. The Agent's old
blocking API waits for a reply from this same actor. It is not an alternative
execution engine.

Ownership remains the model from Chapter 4. Ensemble owns Agents. Each Agent
owns its History, Engine, Tools, and Jobs; Jobs owns the running Job objects.
Pause and interrupt do not destroy that graph. A paused Agent still accepts
controls and worker completions. An interrupted turn must distinguish a tool
that started from a call that never reached dispatch: preserve the real job's
handle, and pair each skipped call with an explicit not-executed result. Do not
invent effects or silently replay the skipped tail on a later turn.

Shutdown is different. It closes admission, accounts for admitted work, drains
owned workers and processes, and only then detaches the children. That ordering
keeps the parent paths available while workers finish. Garbage collection cannot
perform this logical cleanup for us.

The media and multi-Agent examples use these same paths. Vendor requests may
contain inline media bytes, but the canonical log keeps the original locator.
A role tool reaches another Agent through its actual owning Ensemble and waits
on a private reply. There is no hidden workflow engine in the command-line host.

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

The actor publishes an immutable observation after recording the corresponding
fact. A display may receive a completed text, tool call, or result as a
`PartFinal`, not merely a string stripped of its arguments and correlation.
`StateChanged` reflects the reducer's actual transitions; `TurnEnd` describes a
terminal outcome. History remains the durable source for prompts, hints,
request boundaries and completed content. Observations cannot reconstruct it.

Each Agent owns a bounded channel configured to drop old progress when full.
Ensemble has a separate bounded merged channel. Publishing to both is a direct,
nonblocking operation through the Agent's real parent. It does not call an
arbitrary observer, create a task per event, or retain an unbounded callback queue.
A slow display may miss progress, but cannot delay recording or request replies.

This endpoint offers **shared consuming streams**, not broadcast subscriptions.
Two `WaitAsync` callers compete for Agent observations. A waiter also consumes
nonmatching values while looking for its predicate. Root and Agent streams are
separate views, but two readers of either one are not two subscribers. If a later
requirement needs independent subscribers, each needs its own bounded queue;
sharing a channel reader would not implement broadcast.

Admission publishes an `Input` outcome with `Accepted` true or false. This
includes invalid inputs and refusals after shutdown. The reliable task carries
the actual exception; an observer may miss its diagnostic. Shutdown closes work
admission, while observations remain readable until logical destruction. Calls
after destruction are disposed-object errors, not new events from a live Agent.

Chapter 5 does not stream HTTP tokens. Its external workflow displays a complete
text part as one delta followed by a finalizer. That is a presentation choice,
not evidence of incremental provider delivery. Chapter 6 supplies streaming.

## §5.4 The inbound seam

Chapter 4 already put a tool's execution outside the conversation loop. The
remaining deafness came from awaiting its report in the same loop that should
hear human input. Moving execution was not enough; move the wait too.

Use one `Channel<Mail>` with one reader. Public messages and private worker
outcomes enter that same queue. Channel delivery preserves enqueue order;
it does not invent a global physical order for simultaneous callers. Each
message is handled before the next scheduling decision.

The public vocabulary is intentionally smaller than the private mailbox:

```csharp
await agent.PostAsync(new AgentInput(InputKind.Hint, "Keep the examples short."));
await agent.PostAsync(new AgentInput(InputKind.Pause));
await agent.PostAsync(new AgentInput(InputKind.Resume));
```

These tasks acknowledge admission, not completion of a model turn. The actor
validates before changing conversation state or queuing effects. Refused input
must leave it usable for the next valid request. One lazy-started mailbox reader
belongs to the Agent; no separately owned Actor wrapper is needed.

HTTP and foreground-observation workers return immutable outcomes to this
mailbox. They do not append History, select the next tool, or schedule another
model request. A running Job owns its independent execution/output state;
its report is a snapshot, not evidence that execution has finished.

## §5.5 The actor loop

The implementation in `Agent/Actor.cs` follows three readable steps: receive a
mail message, apply its facts, then advance if scheduling is allowed. All lasting
scheduling fields are declared in `Common/Actor.cs`. Async methods provide waits;
there is no second scheduler framework hidden beside the Agent.

**Start a turn.** Record the message, render a frozen request string, record
`RequestSent`, then release HTTP to Engine's `SendAsync`. There is at most one
model request per Agent. Engine returns immutable, unnumbered response events;
the actor assigns their place in History. Engine neither reads live History nor
writes it. A hint received after freezing cannot change those request bytes.

**Apply a response.** Record returned facts before publishing progress. Validate
duplicate call IDs before any effect. Dispatch tool calls serially in response
order. Jobs reserves an ordinary job and records `ToolCalled` before starting
its handler. That synchronous admission prefix runs on the actor; the subsequent
observation wait runs outside it and posts its report back. Controls that do not
allocate jobs retain their Chapter 4 contracts.

**Record a report.** The actor appends `ToolReturned`, including the actual job
locator where applicable, before asking the model to consume it. Host-invoked
direct tools instead record `ToolObserved`: a real audit outcome that does not
invent an assistant tool call in model dialogue. Both preserve output evidence.
A returned `running` snapshot does not mean the worker ended.

**Apply a hint.** Record `HintReceived` without changing the active turn's reducer
state. During HTTP it affects the next request, never the frozen request already
in flight. During ordinary foreground tool observation, shorten the observation
ceiling to the default three seconds **measured from that observation's start**.
If three seconds have already elapsed, return promptly with the real report;
otherwise wait only for the remainder, unless a result or shorter deadline wins.
Do not cancel the Job or refuse undispatched calls merely because a hint arrived.
Dispatch every remaining finalized call normally, once, with the same observation
ceiling while the hint is pending. The next request sees the hint and all paired
results, including actual running-job handles. Idle hints are retained until a
proper trigger; pause still prevents new scheduling.

### Hear the human without throwing away the work

A finalized model response may contain a long file write or a carefully prepared
shell command followed by other calls. The human's hint asks the model to listen
soon, not to discard that prepared work. Shortening observation and cancelling
execution are different operations. Keep execution owned by the Job and make the
observer responsive; do not manufacture success or an `interrupted; not executed`
result for a call that should still run.

For an ordinary observation with start time `s`, requested delay `d`, and default
hint ceiling `D = 3 seconds`, a pending hint changes its deadline to:

```text
s + min(d, D)
```

The start is the start of this observation, not the Job's original launch and not
the hint's arrival. A hint during a twenty-minute observation that has already
lasted ten seconds therefore returns promptly. A hint after one second leaves at
most two seconds. Repeated hints cannot restart that clock. A requested half-second
wait stays half a second, and an immediately available result stays immediate.

Apply this rule both to a new ordinary tool's initial observation and to controls
that observe an existing Job, including `wait_for_job` and `send_input`. Preserve
the original call ID, output cursor/cap, real Job ID, state, and completed result.
One-shot `tool_limits` still belongs to the literal next call; the hint ceiling
shortens observation without rewriting arguments or leaking a limit into a later
call. No stale timer may produce an extra return after the observer has finished.

A pending hint also caps each remaining ordinary observation in the same finalized
batch. Two slow calls may therefore take roughly six seconds in total; a fast
write between them finishes normally. This is a per-observation bound, not a
three-second promise for an arbitrarily large batch. The hint is delivered with
the paired results on the next request. After delivery, a later observation uses
its own requested delay: neither an old wake signal nor a stale hint flag may
keep capping unrelated waits.

Keep the Agent authoritative for pending-hint state. A newly admitted observer
must notice a hint that arrived before it captured a wake signal; an active
observer must notice one that arrives afterward. A content-free wake signal can
shorten the current timer, but must not become a sticky substitute for the actual
mailbox message. Explicit host adapters declared `waitForCompletion` retain their
mandatory-completion contract when their consumer has no Job-wait interface;
this exception is not permission to bypass the cap for ordinary model tools.

**Interrupt remains different.** It deliberately ends the turn and can refuse its
undispatched tail. A hint preserves finalized pending calls and lets the model
choose what to change after receiving the hint. Neither operation silently kills
independent Jobs.

**Finish a turn.** Record terminal facts before completing the request task.
`TurnEnded` restores the reducer to Idle; `TurnInterrupted` records interruption.
These are replayable facts, not a fabricated error used to force a state change.
The Chapter 2 states remain Idle, InputPending, InFlight and ToolsPending.
Scheduling flags are transient actor control, not another canonical transcript.

### Reliable completion is not observation

`SubmitAsync` creates a `TaskCompletionSource<string>` with asynchronous
continuations and places it in the request message. The actor settles it exactly
once. A slow or cancelled caller cannot block the actor while it delivers the
outcome. Concurrent submissions queue separate turns; they do not read a shared
last-answer field or watch for an arbitrary Idle transition.

`SendAsync` waits on that same task. Caller cancellation detaches its wait;
explicit interrupt is how an application ends the turn. `RunAsync` is a sequential
input-channel adapter over submission, not a preserved copy of the old tool loop.
With logging enabled, append/flush precedes publication and completion. An append
failure cannot be reported as durable success; fault cleanup still joins workers
and detaches the graph.

### Limits and lifetime

A positive `ToolRoundLimit` permits that many tool **batches** per human turn.
Zero selects the named default of 200; a negative value disables the cap. After
the last permitted batch, the model may consume its results and answer. If it
requests another batch, execute none of it, pair every call as not executed, and
return a distinguishable limit error. Do not edit signed assistant content or
leave dangling results. The immutable Configuration supplies one policy per Agent.

Pause blocks new model/tool scheduling but continues processing inputs and
completions. Resume continues automatically. Handle-free job controls can act
while foreground work is held; they do not start a competing conversation turn.

Immediate interrupt cancels HTTP and the current foreground observation, not
independent Jobs. `after_tool` retains the current HTTP response or dispatched
tool boundary, then refuses the remaining calls. Both modes end the active turn
and leave the same Agent available. A late completion cannot dispatch stale calls
or settle a newer request. The actor collects its outstanding completion before
starting the next turn; independent Jobs keep their own output and handles.

Shutdown closes admission, cancels in-flight transport/observation and controls,
settles admitted requests, stops Jobs with reason `shutdown`, and joins workers.
History and the parent paths remain attached until this cleanup finishes.
Repeated shutdown joins the same lifetime. `Destroy` then removes relationships
through the owning parents' methods, even when a joined fault must propagate.

.NET cannot forcibly stop an arbitrary managed callback. The callback receives
cancellation and must cooperate; one that ignores it can delay shutdown.
PTY processes can be killed and drained. Do not claim a callback has vanished
merely because a cancellation token was signalled or its Job status is killed.

## §5.6 Wait and flush

Use `WaitAsync(predicate, cancellation)` for display progress. Use the relevant
request task for orchestration. A lossy state event cannot identify which of
several queued prompts completed. `Ensemble.WaitAnyAsync` applies the same rule
to the merged stream; it does not gather a group of reliable answers.

`FlushAsync` is useful at a CLI's EOF: wait until active/queued turns and
foreground controls settle. It does not close admission, resume paused work,
or wait for independent background Jobs. Later inputs may extend the wait.
Cancelling the waiter does not cancel those turns. Shutdown that discards
unfinished work faults the pending flush rather than claiming it drained.

This is deliberately less machinery than a workflow engine. The application
retains the author request task, then starts the editor after success, then the
reviewer. Observations explain progress; reliable tasks decide transitions.

## §5.7 Hints and interrupts

A hint changes information for the next safe decision. An interrupt ends the
current turn. Neither is permission to delete the Agent or kill independent
work. The same mailbox accepts both while HTTP or a tool observation is waiting.

Test the two controls separately with a held real Job and several finalized
calls in one batch. For a hint, assert that the current observation obeys its
original three-second ceiling, every remaining call executes once, and the next
request contains the hint with truthful results and stable running handles.
Include a short observation, a fast result, repeated hints, and a fresh next-batch
observation that must not inherit the ceiling. Release the held Jobs and retrieve
their real completions through the same handles.

For explicit Interrupt, assert that the first Job remains accessible but the
undispatched calls never run and receive not-executed results. Then submit a fresh
prompt and prove the Agent still answers. Repeat at the HTTP boundary: release a
late response containing calls and prove that an ended turn cannot cause their
effects. Use gates rather than sleeps to establish which boundary the test
actually reached.

### The stdin protocol

The CLI reads stdin asynchronously and posts to the Agent's mailbox. The protocol is JSON lines:

```json
{"kind":"prompt","text":"Write a haiku about refactoring"}
{"kind":"hint","text":"Use a metaphor about gardens"}
{"kind":"interrupt"}
```

Backward compatibility with the Chapter 1 format:

```json
{"user":"Write a haiku about refactoring"}
```

A message with a `"user"` key and no `"kind"` key queues a separate reliable
request. The CLI continues reading stdin and emits `{"assistant":"..."}` replies
in request order. Optional `parts` uses the public Part JSON format and takes
precedence over `user` text, including an empty array. Media bytes are resolved
only when projecting a request. Live `prompt` and `hint` inputs instead use the
actor's current turn state; they can become hints while the Agent is busy.

The C# CLI also accepts these direct actor operations:

```json
{"kind":"pause"}
{"kind":"resume"}
{"kind":"interrupt","mode":"after_tool"}
{"kind":"tool","id":"host-1","name":"list_directory","arguments":{"path":"."}}
{"kind":"shutdown"}
```

Interrupt mode defaults to `now`. Direct tool dispatch returns `tool_result`;
its audit result is not an invented model-requested call. Live inputs have an `ack` after admission; direct tools acknowledge queuing and
return their outcome later; shutdown acknowledges after joining. Live mode additionally
emits lossy `observation` records with kind, sequence, text, accepted flag, state
transition and optional part. These are progress, not request replies. Explicit
interrupt/shutdown produces an error outcome for an unfinished reliable request.

EOF calls `FlushAsync`: wait for active/queued turns and foreground controls, not
independent background Jobs. Then shutdown cancels and joins remaining Jobs.
Flush does not resume paused work; send `resume` or `shutdown` before EOF if work
is paused. Explicit shutdown cancels instead of draining queued turns. Cancelling
a flush waiter detaches only that waiter. New inputs may extend a flush wait;
it is a quiescence wait, not a frozen-prefix barrier.

## §5.8 Managing multiple agents

Ensemble already owns the peer Agents. It is not supplemented by another
coordinator root or mutable name registry. Construction supplies the real parent
and optional name; `AddAgent` validates ownership and name uniqueness before
changing either relationship end. `RemoveAgent` completes both ends. Destroying
one root cannot touch a same-named Agent belonging to another root.

```csharp
var author = new Agent(app, authorConfiguration, name: "author");
var editor = new Agent(app, editorConfiguration, name: "editor");
var reviewer = new Agent(app, reviewerConfiguration, name: "reviewer");
```

Each immutable Configuration supplies its role's `PrimaryInstructions`.
Installing tools appends tool-use guidance; it must not erase that role prompt.
Each Agent owns its own concrete Tool objects, History, Engine, Jobs and mailbox.
Two held requests on different Agents may progress concurrently; one Agent never
has two model requests in flight.

The merged observation envelope carries the actual source Agent. A hint addressed
to `author` uses `app.FindAgent("author")`, not a global fleet lookup followed by
filtering. A callback reaches diagnostics through Job → Jobs → Agent → Ensemble.
Application data such as the manuscript has one authoritative home, not one copy
per role.

This chapter's peers are constructed by application code. They are not children
spawned by a model tool. Adding the reviewer must not require a new library
workflow API; the application simply constructs another Agent and awaits its
reliable request. Sub-Agent ownership and spawning contracts remain later work.

## §5.9 Media capabilities

Image, audio, video, and document support are independent facts. `ModelFeatures`
uses separate flags for them. Do not infer audio support from image support, or
silently omit a part the selected model cannot accept.

`Engine.LookupModel(vendor, model, out features)` is an exact lookup. The small
Chapter 5 table declares images for `claude-3-haiku-20240307`, `gpt-4o`, and
`gpt-4o-mini`, and image/audio/video/PDF input for `gemini-2.5-flash`, each under
its own vendor. These are the renderer's supported declarations, not a claim
that a paid live request was run or that a model remains available indefinitely.
An unknown identifier still works for text; there is no default attachment row.

The renderer checks capabilities before reading a file or starting HTTP. An
unsupported attachment produces an error such as:

```text
Model 'claude-3-haiku-20240307' does not support audio media.
```

`BlobRef` is value-like: it has no owning parent because it describes an address,
not an owned file. Its `Kind` and `Locator` remain in canonical events and survive
as a locator when a blob is redacted. Rendering calls `Data()` only on local-path
references and creates disposable vendor JSON containing the actual bytes:

- Anthropic uses an image block with a base64 source and MIME type.
- OpenAI Chat Completions uses an `image_url` block containing a data URI.
- Gemini uses `inlineData` with its MIME type and base64 data.

The surrounding text parts retain their order. History does not acquire base64
content or a second authoritative file cache. A missing file is an error; a saved
locator cannot guarantee that external content remains available forever.

URI and job-handle attachments remain deferred. Unsupported tool-result media
placements also fail explicitly: this adapter cannot turn a blob into empty
text and pretend the model received it. Anthropic's nested result blocks can
carry supported images; the current OpenAI and Gemini tool-result projections
are text/structured-result projections, not media containers.

## §5.10 Exercise: Author, Editor, Reviewer

Build `Examples/Workflow/Workflow.csproj` as a separate C# assembly referencing
the public library, as the Chapter 4 consumer did. Three agents, three roles:

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
reviews. The application advances through reliable `SendAsync` or `SubmitAsync` completions.
`PostAsync` delivers live hints; `WaitAsync` watches progress.
Each agent uses the same vendor (the fake from Chapter 2) with a
different system prompt. The workflow is ordinary C# application code. Every new
collaboration pattern is a new application program. That is the limit this
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

The C# workflow is in `agent/Examples/Workflow`.
Keep the historical `CH06_LOG` environment name for protocol compatibility.
Build that project separately. This offline exercise requires a loopback
endpoint; it does not grant paid-provider access. The local provider fake supplies
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
payload fields are specified in the exercise walkthrough below. A hint acknowledgment is a delta whose
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

There is no streaming transport, GUI, model-driven spawning, grouped-results API,
URI fetching, or job-handle attachment resolver here. The workflow's whole-part
delta is not a token stream. `BlobRef` can describe URI and handle locators, but
this renderer rejects them explicitly. These omissions are boundaries, not
silent fallbacks.

The actor and progress seam provide foundations for later chapters. They do not
turn a lossy `Wait` into a future supervision contract. Keep the current surface
small and teach its exact promises before adding new ones.

## Taking it for a spin

Build the actual library tests, CLI and separate consumer using the approved
.NET 10 SDK. The working CLI project retains the historical `Ch04.csproj` and
`ch04` assembly names; that is a filename, not the chapter boundary.

```sh
dotnet run --project agent/Tests/Tests.csproj
dotnet build agent/Cli/Ch04.csproj
dotnet build agent/Examples/Workflow/Workflow.csproj
```

Use an owned loopback fake and dummy credentials for local exercises. Set
`LLM_VENDOR`, `LLM_MODEL`, `LLM_BASE_URL`, `LLM_API_KEY`, and an owned `CH06_LOG`
path for the workflow, then run its compiled DLL. The exercise fake must actually
call `think`, `write_draft`, `word_count`, `read_draft`, `edit_draft`, and `review`;
a fake that merely returns final prose does not test manuscript sharing.

Send the topic, wait until the author's real `think` Job starts, then send the
addressed hint shown in §5.10 and EOF. Inspect all three captured role prompts and
tool declarations. The next author request must contain the hint, the first
Job's real running report and actual results for every remaining finalized call.
Verify those calls execute exactly once; a hint must not produce a not-executed
substitute. Verify the
stored author's text reaches the editor, the edited text reaches the reviewer,
and the review decision and notes are actually returned by `review`. The three
reliable pipeline markers must occur in order before exit zero. Tagged progress
may be dropped; those markers do not depend on it.

Illustrative output shape, not a captured provider transcript:

```json
{"observation":"state_changed","agent":"author","from":"idle","to":"input_pending"}
{"observation":"part_delta","agent":"author","part_id":1,"chunk":"A draft about the ocean."}
{"observation":"part_final","agent":"author","seq":3,"part_id":1,"part":{"type":"text","text":"A draft about the ocean."}}
{"observation":"turn_ended","agent":"author"}
{"pipeline":"author_done"}
```

The consumer's `state_changed` fields are `from` and `to`; whole-text
`part_delta` carries `part_id` and `chunk`; `part_final` carries `seq`, `part_id`
and the complete serialized `part`. Display IDs are unique within this consumer
run; they are not durable History identities. Agent sequence numbers remain local
to that Agent. The shared application log records the same output lines, while
`CH06_LOG.author.jsonl`, `.editor.jsonl` and `.reviewer.jsonl` preserve separate
canonical histories.

Also exercise the single-Agent CLI: queued legacy replies, a live prompt and hint
during held HTTP, pause/resume, both interrupt modes, direct host tools, explicit
shutdown and EOF. Do not use a sleep alone as proof that a tool has started.
Check its real dispatch event or gated handler before sending the control.

For local media, `parts` takes precedence over fallback text. `kind:1` below is
the public codec's local-path Ref kind:

```json
{"user":"ignored fallback","parts":[{"type":"text","text":"Describe this."},{"type":"blob","mime":"image/png","ref":{"kind":1,"locator":"/your/owned/image.png"}}]}
```

With a declared image-capable model, inspect the captured request's actual bytes,
not just its MIME label. The saved event must retain only the locator. Unsupported
media on a text-only/unknown model must fail before a network request. Replay the
logged prefix preceding each `RequestSent` through the same vendor/model renderer
and compare it byte-for-byte with the live request. Replay must need no provider
key and must not rewrite canonical History.

The local verification of this C# endpoint covers actor/component behavior,
three-vendor image projection, Gemini WAV/PDF byte projection, representative
video-container bytes, the scoped executable grade, and the external three-role
workflow. Video bytes do not prove a decodable movie. Loopback fixtures do not
prove current model availability, remote acceptance, or live-provider behavior.
Real-provider exercises remain incomplete because the authorized allowance is
exhausted. No installation into the working Ensemble follows from these tests.

### Historical context, not carried evidence

The retained Go chapter described a Gemini ocean-writing workflow and an
Anthropic cartoon-image test, and linked a historical ocean conversation. Those
runs motivated the exercise, but they are not runs of this C# endpoint. The
important lesson survives: application code supplies roles and manuscript tools;
the actor handles transport, hints, jobs and reliable completion without knowing
what an author or editor is. Streaming, the GUI and model-created collaboration
still come later.
