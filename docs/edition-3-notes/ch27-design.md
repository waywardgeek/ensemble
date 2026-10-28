# Chapter 27 design — Goals

Status: design recorded before implementation; current-runtime source is now implemented, locally tested, and directly self-reviewed. Bill authorized Cadence to perform this exception without new sub-agents. The chapter body was written only after final local validation. No installation, restart, live activation, paid provider exercise, commit, or student Chapter 27 acceptance is claimed. Chapter 20 workers remain separate and untouched.

## Why this feature

A useful long-running assignment should not depend on Bill repeatedly saying “please continue.” Sometimes the model finishes a response prematurely; sometimes a request ends in an error. Instructions can improve the first case, but cannot run themselves after the turn ends. An editable goal in SYSTEM supplies persistent direction during requests; an opt-in idle reminder supplies another opportunity to act.

This is not a diagnosis of the recent Chapter20 observation discrepancy. Cadence prematurely declared a blocked student, and Bill reported it was running. That claim was withdrawn. A goal timer does not excuse that judgment error, prove a network failure, or detect a request that remains active forever.

## Boundaries and user control

There are three separate facts: the goal file exists; this Agent has configured that file as its goal; its idle reminders are enabled. None establishes semantic completion. Loading a skill exposes a capability; it does not activate it. No existing session, child, startup configuration, primary tool grant, or live runtime is silently opted in.

The dedicated `goals` skill teaches use of `goal_control`. The compiled control is not declared or callable before its skill grant is active, including legacy/no-primary-skill operation. Listing `goals` as available to load is not loading it. Cadence loads and activates it only on Bill's request, not as automatic error recovery.

`goal_control` accepts optional `file`, `enabled`, and `idle_seconds` fields. No fields reads status. First enabling requires a configured readable, nonempty goal file. The default idle delay is 480 seconds (eight minutes); explicit delays must be positive, finite, and representable. Omitted fields preserve current values. `enabled:false` must work even when the configured file is missing or invalid. `file:""` clears the configuration and disables reminders. Invalid updates preserve the prior configuration rather than half-applying it.

Configuration is per-Agent and process-local. It is not inherited by children or automatically restored on restart. The file is durable, but a restarted Agent requires explicit configuration. This is a deliberate limit of this first feature, not a durable scheduler. Compaction workers and terminal/submitted actors cannot acquire autonomous reminder lifetimes.

Turning reminders off retains the configured goal in SYSTEM. Clearing configuration removes it. Unloading the `goals` skill removes its control grant and clears its active goal configuration/reminder; the file itself is never deleted. Reloading does not reactivate an old timer.

## File and request ownership

The Agent owns the selected file and all filesystem operations go through its existing Sandbox. Relative and absolute names retain that owner's established semantics; there is no host-open bypass, new data directory, or parallel file API. The tool returns the configured path, delay, and status, not the goal's full contents.

Read the configured file afresh for every model request, including subsequent requests within one tool-using turn. Form a request-local SYSTEM contribution once. It is separate from primary instructions and skill manuals, not a tool result, ordinary user message, conversation summary, or content copied into a child. Existing provider renderers preserve their SYSTEM/developer conventions. A file edit changes the next request; an already sent request keeps its original bytes. Do not rewrite canonical History or make a second copy during save, replay, micro_handoff, or compaction. The goal remains available because its source is outside the compacted conversation, not because it is an immortal History event.

A read error or newly empty configured file is explicit in status. It disarms automatic reminders and is not interpreted as completion. For an ordinary human-initiated request, include a clearly labelled unavailable-goal notice in SYSTEM rather than silently omitting the goal or pretending stale content is current. This permits the agent to help repair/clear the configuration. An automatic wake with an unreadable/empty goal does not start a model request. Fixing the file permits fresh reads; reminders require explicit re-enabling after this safety stop. No silent truncation or invented fallback goal.

## Actor-owned idle lifecycle

Reuse the existing actor mailbox and turn queue. The timer callback only posts a wake message with its generation. It does not call the Engine, mutate History, run a tool, replay a command, or manipulate another Agent. The actor verifies current configuration, generation, pause/stop state and eligibility before admitting the reminder.

Idle means no active model/tool turn and no pending ordinary turn. Existing background jobs and children may still be running; checking them is one legitimate purpose of a reminder. An active HTTP request or a turn waiting in an owned tool is not idle. The feature must not classify a quiet active turn as failed.

When an enabled, eligible Agent becomes idle, arm one timer for the configured delay and publish its due time. Activity cancels the timer and invalidates already queued callbacks. Once that activity finishes, a fresh full delay starts. Queued ordinary input takes precedence over an idle reminder. Repeated timer callbacks cannot create duplicate or overlapping turns.

On expiry, read the current goal, admit one clearly labelled automatic reminder through the normal event loop, and let the existing actor own History, rendering, tools, accounting, cancellation, and completion. The reminder asks the agent to check the goal and actual jobs/children/results, continue authorized unfinished work, or turn reminders off and explain why it needs the human. It must not impersonate a newly authored human command or claim that a prior tool did or did not execute. The goal itself stays in SYSTEM rather than being repeated in the wake text.

A normal end or transport/inference error returns to the same idle policy: a subsequent attempt is separated by the full delay. This is a new assessment turn, not a retry of an old request body. Explicit pause suspends wake-ups and resuming begins a new full idle delay. Explicit Interrupt disables reminders; human stop must not be undone by a later timer. Shutdown/unload/clear invalidate callbacks and leave no reminder-owned goroutine, turn, or timer lifetime behind. None of this repairs or times out a request that still reports itself active.

## Public state and GUI

Expose one immutable snapshot from the owning Agent, plus an observation when it changes. Include configured file, enabled flag, idle delay, current phase, optional next-reminder time, and a useful error. Publish only after the actor has accepted the state change. Getter/replay reads must be concurrency-safe and must not start or re-arm timers.

The GUI subscribes to the existing observer path and includes this snapshot in initial/reconnect/Agent-selection replay. It must show the selected Agent's state, not a last message from a different Agent. No GUI timer decides when the core should wake. An optional displayed time is derived from the core's due time, never authority to call the model.

Use a visible text label with accessible text, not color alone: “No goal,” “Goal: working,” “Goal: waiting,” “Goal: paused,” “Goal: reminders off,” or “Goal: error,” with the reminder-enabled distinction preserved while working. Show the configured file/delay/next time/error in readable details. “Reminders off” must never become “completed”: runtime does not infer whether the goal has been achieved. Do not add automatic speech, a second narrator, flashing countdowns, or a goal-management dashboard. The existing Conversation and Actions panes remain available.

## Implementation and review plan

Extend the actual actor/mailbox owner and existing SYSTEM assembly, skill registry, public observation aliases, and GUI projection. Use one configuration/control path, one timer owner, and request-local prompt data. No task hierarchy, dependency scheduler, watchdog, new persistence format, global model retry, synthetic tool results, duplicate actor loop, or duplicated credentials/configuration system.

Cadence implements and reviews this exception directly. Tests use existing local fake-provider and GUI helpers, owned files/processes/ports, the established sequential offline Go environment, and no real keys. Focused tests/race checks and affected compilation are authorized; installation, running-session activation, settings changes, restart, publication, and paid traffic are not implied.

Acceptance follows the exercises in Chapter27. Review source as well as results: particularly lock ownership, timer generations, turn/error transitions, per-request SYSTEM placement, default disclosure, unload, and GUI replay. Keep preparation mistakes distinct from failed behavior. If implementation exposes a simpler or necessary adjustment, record the reason and update the pre-code contract explicitly rather than silently redefining it to pass a test.

Only after implementation, exercise results and direct self-review will Cadence write the chapter body. The crossover account may explain why the feature was added during use, but must distinguish the proposed mechanism, implemented source, passing local checks, installed binary, enabled session, and observed real behavior.

## Design refinements before completion

Bill replaced the proposed exponential poke schedule with a fixed eight-minute idle reminder. The exponential edit was interrupted before execution; no backoff counter or exponential timer state was implemented. Each subsequent return to eligible idle state uses the same configured delay.

Bill also reported launching a separate top-level companion-guide agent to explain *The Art of Building AI Coding Agents* to humans who want the knowledge without executing the book. That assignment is independent, not a replacement for this chapter's executable contract or implementation evidence. Preserve its concurrent work; Cadence is not duplicating or directing that assignment.


## Implementation and direct self-review

I reviewed the actual actor, Engine renderers, registry/unload hook, tool admission, and GUI projection against the eight exercises. The design remains one file, one actor-owned timer, and request-local SYSTEM text. No new scheduler, retry service, history format, or child inheritance was added. Public GoalController is optional, preserving existing Agent implementers; dormant control installation does not make it a default grant.

I corrected three demonstrated ownership/wire defects during this continuation: a synchronous reminder-start failure falsely published working without rearming; Codex Responses duplicated the goal through developer input and top-level instructions; and goal publication during discarded child construction deadlocked on its parent's reservation lock. The fixes stay in those existing owners. The last uses the existing rejected-child GUI regression rather than another harness. Earlier lazy-actor startup and skill-grant installation defects remain preserved in the initial receipts. Several new test compilations also failed because I guessed current-runtime API names from the student tree; those were my setup errors, not product findings.

Evidence is retained under `/private/tmp/ch27-goals.5vUy1q` (pointer `/private/tmp/ch27-goals-current`). `goals-preflight-confirmed.log`, `goals-codex-before.log`, and `goals-final-validation.log` preserve behavioral failures; earlier setup logs remain there too. The successful final sequential receipt is `goals-final-validation-second.log` with exit 0: all 15 goal tests under race (including five provider-route subtests), the full GUI Go suite under race, JavaScript status checks, the isolated real-GUI browser exercise, affected vet, and root compile-only sweep. Existing Engine and skill-loop behavior from `goals-regression-third.log` and the Engine result in the first final batch remains applicable; the final fixes changed neither their assertions nor their exercised non-goal rendering. Do not relabel no-test packages or the final compile sweep as behavioral tests.

The final browser used a freshly built owned binary, synthetic Anthropic responses, an owned dynamic port/profile, and transcript-only speech. It established status text/accessible labels/layout, unchanged reconnect due time, child selection isolation, Interrupt, foreign-frame rejection, selection loading, and disconnect after owned server termination. It did not establish human visual quality, hearing, provider judgment, normal EOF, live installation, or behavior on port 8088. No credentials were read and no real provider was called. Tests use short explicit intervals; eight-minute default behavior is configuration-asserted, not a wall-clock eight-minute observation. Finite lifecycle tests and source review do not prove every possible callback interleaving.

I preserved the concurrent companion guide, Chapter 20 student source/contexts, and pre-existing runtime handoff amendments. No broad staging or commit was performed. The completed body received a direct factual/voice review. The first prose lint caught two em dashes in the pre-code TL;DR; those were corrected. Final lint has no hard violations, with soft negation-density and person-gap warnings retained rather than adding filler or weakening safety distinctions. This section records self-review rather than independent reviewer acceptance.
