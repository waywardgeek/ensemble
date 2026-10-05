# Retire the synchronous agent loop

Plan written 2026-10-04, at Bill's request. Implementation has not begun.

## Goal

One production orchestration loop, owned by the actor. Synchronous APIs are
send-and-wait adapters, not separate agents with subtly different behavior.
Leave the implementation cleaner as we work through the codebook; do not keep
obsolete machinery merely because an earlier chapter introduced it.

Chapter 6 already states this contract: synchronous `Ask` posts a user message
and waits for its turn to end. Preserve public behavior where possible, not the
old implementation. Do not introduce a third generic loop under two wrappers.

## Verified starting point

- `internal/llm/engine.go`: `Engine.Ask` / `AskWatching` own a synchronous
  model → execute tools → model loop.
- `internal/llm/actor.go`: `handleUserMessage` / `runTurnLoop` duplicate turn
  setup and orchestration, adding hints, interrupts, pause gates, asynchronous
  tool completion, visible-reasoning enforcement and observations.
- Both loops reference one `MaxToolRounds = 16` constant. The persisted
  `max_tool_rounds` preference is not connected to either loop. Bill set it to
  200, but that does not yet change execution.
- Public `Agent.Ask`, and therefore `cmd/virtual-user`, use the engine loop.
- The current CLI's legacy `{"user":"..."}` branch also calls `eng.Ask`, even
  though `runActorLoop` already has an actor running over that engine. Its
  `{"kind":"prompt",...}` and text-chat paths use the actor.
- `Actor.Ask` exists, but waits on the shared 256-entry observation channel.
  `notify` drops observations when that channel is full; a terminal event can
  be lost, and multiple waiters can consume one another's events. This is not
  a reliable request-completion channel.
- `Actor.Shutdown` calls engine shutdown; it does not stop and join `Run`.
  `finishTurn` notifies completion before saving and ignores the save error.
  A direct reroute of synchronous callers would change durability/error behavior.
- Public `Agent.NewActor` currently creates another wrapper on every call.
  The migration must not create two mailbox consumers or owners of one engine.

## Target contracts

### One owner and one loop

The actor owns multi-round execution and runtime mutations of its engine.
`Engine` keeps one-request rendering/parsing, event reduction, tool execution
primitives and persistence. It no longer owns a user-to-final-answer loop.

One `Agent` owns one actor. Blocking `Ask` must work without requiring callers
to manually start a goroutine; explicit actor clients must keep working too.
Both APIs reach the same actor. Starting/running twice must not start a second
consumer. Define and test start, stop, join, repeated shutdown and use-after-stop
behavior before switching callers. Keep this lifecycle machinery small.

Shutdown stops acceptance of new work, resolves pending blocking callers, and
joins the actor before shared state is saved or closed. Audit in-flight model
requests and tool workers: do not claim cancellation or joining merely because
`Run` exited. Preserve results of work that actually completed; do not fabricate
success for work stopped or never dispatched. Make any non-cancellable-tool
limitation explicit rather than expanding this task into a job-system rewrite.

### Reliable synchronous completion

A blocking request receives its own final text/error exactly once, independently
of streaming traffic and other observers. Use a request-scoped completion path
through the mailbox, not a wait for any `TurnEnded` on a shared lossy channel.
Keep this control plumbing out of the durable conversation event format.

Concurrent blocking requests are queued as distinct turns, not accidentally
converted into hints or returned another turn's answer. GUI hints retain their
existing mid-turn behavior. Cancellation/shutdown must resolve the correct
caller and must not leave a waiter stranded. A cancelled caller must not leave
an unbuffered reply send blocking the actor.

Completion must account for required persistence: surface save failures to the
caller and observers, and do not announce durable success before saving. Leave
streaming observers non-blocking; reliable completion is not a reason to make
every delta lossless or introduce a general event-bus framework.

### Tool-round limit

Proposed semantics (part of this plan, not implemented):

- A positive value N allows N dispatched tool batches per user turn, not N
  individual calls. Parallel/wrapped tool calls do not each consume a round.
- Zero selects a named default of 200. It does not mean unlimited execution.
  Document this in the settings contract and GUI. Retain the existing settings
  validation bound unless a separate requirement changes it.
- The actor reads the effective value once at the beginning of a turn. A GUI
  edit while a turn is active takes effect on the next turn. Load the persisted
  value before the first turn; use the same source for runtime updates. Public
  library callers get an explicit configuration path and the same default.
- After N batches, allow the model request that consumes their results and may
  finish normally. If it requests another batch, dispatch none of that batch.
  Record an explicit not-executed result for each call, with the limit reason,
  then finish with a distinguishable limit error. These are known non-executions,
  not synthetic unknown/lost outcomes. Do not alter signed assistant content.
- Persist a well-formed history. A later user instruction, including an increased
  limit, can start another turn without duplicates, dangling calls, or silently
  executing abandoned calls. The model can reissue the work when asked to resume.
- Count dispatched batches, not retries caused solely by narration enforcement.
  Keep the existing tool-ordering contract within a batch.

There will be one enforcement point in the actor. Delete the old hard-coded
loop limit rather than wiring the setting separately into both loops.

## Implementation sequence

### 1. Lock down behavior and caller inventory

Read the relevant chapter 5–9 contracts and identify every direct `Engine.Ask`,
`AskWatching`, public `Agent.Ask`, `Actor.Ask`, actor-start and shutdown caller,
including exercise/grader code outside `agent/`. The inventory above is the
production starting point, not a claim that the external caller audit is done.

Add integration tests against a local fake vendor and real tool registry for
public blocking Ask, explicit actor use, current CLI prompts and legacy CLI
prompts. Assert observable output, tool order, transcript and save behavior.
Include recall and turn/round ephemeral hooks firing exactly once at the right
time. Audit visible-reasoning enforcement: do not silently disable supervision
on synchronous calls just to retain the old bypass.

### 2. Make the actor safe as the sole execution owner

Implement the lifecycle and request-scoped completion contracts above. Test
completion under more than 256 observations, multiple queued requests,
interruption, cancellation, shutdown during tool work, and persistence failure.
Keep the existing public non-blocking observer seam; avoid unrelated API churn.

### 3. Migrate callers, then delete

- Route public `Agent.Ask` through its owned actor; `NewActor` exposes that same
  owner rather than creating an independent runtime over the engine.
- Translate both CLI prompt formats into the same actor request path. Preserve
  documented JSON-lines output and exit status, and avoid echoing streamed text
  twice. Audit adjacent legacy `ephemeral` handling for direct context mutation
  from the stdin goroutine; route runtime mutations through the owner.
- Verify the virtual user and other embedding programs still work through their
  existing public API, including cleanup and errors.
- Delete `Engine.Ask`, `Engine.AskWatching` and their orchestration-only code once
  no current callers remain. If a supported caller needs a streaming convenience,
  adapt actor observations; do not retain an engine loop for it.
- Consolidate recall/ephemeral turn setup in the actor. Keep execution primitives
  only where the actor still needs them. Remove stale comments about two paths.

### 4. Connect the round-limit preference to the remaining loop

Implement the setting's startup/runtime wiring and the boundary behavior above.
Keep settings persistence separate from execution; transfer/read the small policy
needed by the owner rather than letting the WebSocket goroutine mutate engine
state. Do not bundle fixing every other dead setting into this change.

### 5. Prove equivalence and finish the codebook migration

Run the full Go suite, affected chapter checks, and race tests on the actor,
public API, CLI and WebSocket paths. Use deterministic local HTTP fakes and real
tools; no paid model is required to test 200 rounds.

Required regression cases:

- Both public blocking and actor APIs execute the same tool workflow and preserve
  exact final text, tool ordering, recall and ephemeral hook counts.
- Legacy JSON prompt, current JSON prompt and CLI chat use the same runtime;
  each emits one final answer and reports failures correctly.
- More than 16 batches succeed with a configured limit of 200. At the exact
  boundary, a final answer succeeds, but another tool batch is not dispatched.
- A small limit, zero/default, an explicit library value, GUI updates and a
  persisted value after restart all behave as documented.
- A limit stop saves valid tool-call/result pairs, reports a non-success outcome,
  and can resume on a subsequent user turn. Live and replayed context agree.
- Saturated observation traffic cannot lose the blocking caller's completion;
  two queued blocking requests receive their own answers.
- Shutdown/cancellation does not hang callers, mutate context after final save,
  or start duplicate consumers. Save errors are returned, not swallowed.
- Existing stale tool-completion, lost-result repair, hint, interrupt, model-switch,
  GUI replay and streaming tests remain intact.

Update the current codebook instructions so chapter 6 explicitly retires the
pre-actor orchestration loop while retaining the synchronous API as an adapter.
Correct later instructions that assume two paths. Do not edit frozen early
`solutions/chNN` snapshots to pretend actors existed before their introduction;
run their compatibility checks and document any actual contract conflict rather
than weakening a grader. Record the migration in the crossover changelog and
close the round-limit TODO only after tests prove it is live.

## Done means

- One multi-round orchestration loop in current production code.
- No production call path bypasses the actor to run a second loop.
- `Ask` is a small adapter with reliable completion and explicit lifecycle.
- Bill's saved value of 200 actually governs subsequent turns.
- Limit stops, save failures and interruption are observable and resumable.
- Obsolete methods, duplicated setup and misleading comments are deleted.
- Tests prove the contracts; chapter guidance teaches the cleaned-up design.

## Not part of this change

The separate Sonnet signed-thinking rejection, GUI model restoration, broad
settings overhaul, full observer-delivery redesign, and parallel tool scheduling.
Avoid a universal scheduler, strategy-based loop framework or speculative APIs.

## Implementation checkpoint: ownership and settings migrated

Implemented in the working tree, not yet a claim that every acceptance item is
closed:

- `Engine.Ask`, `AskWatching`, duplicate `Execute` and duplicate `Shutdown` are
  retired. `Agent.Ask`, public actors, virtual-user callers and both CLI prompt
  formats use one actor. Blocking requests have private buffered replies;
  observations remain bounded and non-blocking.
- Workers post tool outcomes and context effects to the mailbox. Tool execution
  and job reporting are separate joined workers: a callback deadline can return
  a running-job report before a Go tool returns. Shutdown joins non-cancellable
  Go handlers without suppressing their actual results; managed-process stop
  requests also apply when a process attaches after shutdown begins.
- Saved settings are exercised through a compiled CLI and a fake HTTP provider,
  not merely by setting Config in a unit test. Both JSON prompt formats permit
  18 batches with saved 200, stop with paired refused calls at saved 1, and resume
  after process restart. Legacy ephemeral attachment reaches the model. A
  runtime-store change during a turn is observed on the next human turn only.
- External callers can inspect `ToolRoundLimitError` and `ErrActorStopped`.
  The latter is an immutable typed constant: the architecture grader caught
  the initial mutable sentinel.
- Chapter 6 now teaches request-specific completion and retirement rather than
  shared idle waits and preserving both loops. Chapter 9 includes settings
  execution tests. Fixed counts of goroutines and advice to discard late Go
  tool output no longer describe the runtime and have been replaced with the
  ownership/join contract.

Evidence and outstanding checks:

- Agent full suite and full race suite passed after the tool dispatch migration;
  the newer runtime-setting boundary test separately passed five race runs.
  Run both complete suites again for the final tree.
- The first root run used a five-minute timeout; chapter 4's audit explicitly
  documents about six minutes. The normal-timeout run completed in about 8.5
  minutes but failed the mutable-global audit and exposed stale grader code.
- Chapter 7 mutations had stale parser anchors (`partID` versus `pm.id`);
  their expected failures were retained. The reference and deletion audit now
  pass. Chapter 8's replay harness repeatedly fetched the same range and dropped
  namespaced string part IDs during numeric-only decoding. Fixed both harness
  defects; kept the final-part-plus-tool-event requirement. Chapter 8 now passes.
- Still required before declaring the entire plan complete: final full root
  gate, explicit recall/turn-and-round-ephemera parity evidence, and final public
  API/comment audit. No frozen chapter solution has been rewritten.
