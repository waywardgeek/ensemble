# Chapter 3 — seed notes (tools)

**Status:** no outline yet. This file is the raw material from the 2026-09-13
session with Bill, written down before it rots. Chapter 3 is *unwritten* but it
is not *unspecified*: four other documents have already made binding promises
about it, and those promises are the spec.

---

## 1. What chapter 3 has already been promised to do

Every item below is a commitment made in a committed document. Breaking one
leaves an earlier chapter holding a forward promise nobody collects, which is
the failure mode P9 exists to catch.

| # | promise | source |
|---|---|---|
| 1 | The response `type` filter becomes load-bearing. Chapter 3's wire **must** return a reply containing a non-text block. | `chapter-01-outline.md` §1.2 and §1.8 ("chapter 3, the first time a model asks to call a tool") |
| 2 | The event log gains **job events**. Note the noun: *job*, not *tool*. | `chapter-02-outline.md` ("Chapter 4 adds `Interrupted`. Chapter 3 adds job events.") |
| 3 | Chapter 3 must **time** a tool call. `ToolCalled` exists as an engine event partly for this. | `chapter-02-outline.md` ("so that Chapter 3 can time one and Chapter 4 can cancel one") |
| 4 | Tool output arrives **by the megabyte**, which is where JSON-lines greppability finally pays. | `chapter-02-outline.md` |
| 5 | Chapter 3's tools are **not instant** — that is precisely what makes chapter 4's mailbox earn its keep. | `chapter-04-actors-parking.md` |
| 6 | Chapter 3 is where a **deliberately-declined design decision** first appears (P6 anti-cheat). | `course-policy.md` P6 |

Promise 4 is a gift: chapter 2 built `RedactedPart` / `RedactSummary` and
nothing has yet *used* it. Megabyte tool output is its job. That is additive
under P1 and it collects on a debt.

---

## 2. The thesis — CORRECTED BY BILL, and the correction is the good part

**What I proposed (too broad, do not use):** "a tool call is a process you
start." Universal claim, all tools.

**Bill's correction, 2026-09-13, verbatim:**

> "There's not much reason to treat edit_file like a process, it works or fails
> quickly in every attempt the LLM ever made, and this is true of most of the
> locally executed tools that do not depend on anything out on the Internet.
> With MCP tools, we lost a lot of the control, and we have to assume any MCP
> tool we don't write ourselves could be flakey. They have to be handled like a
> process, just like run_command."

**The corrected thesis, which is a decision rule rather than a slogan:**

> The process model is earned by tools that **cross a boundary you do not
> control**. Local deterministic tools resolve or fail fast. `run_command` can
> hang. MCP tools are services someone else wrote, so you must assume they can
> wedge.

Why the narrow version is better:
- It is **true**, and Bill lived it; the broad version is retrospective tidying.
- It gives the reader a rule to apply on day one instead of a posture to adopt.
- It resists the over-engineering that "supervise everything" invites — which
  would have readers building job machinery around string replacement.
- It sets up chapter 5: MCP is where the process model stops being optional.

**Historical precision (Bill, 2026-09-13):** `run_command` has started *jobs*
since the StackAgent sprint in **July 2025** — `ai_callback_delay`, then check
status / wait / send input / kill, enough to drive `ed` interactively. Making
**all** tools manageable this way is new, "just in the last month or so."
Sub-agents were treated as processes rather than blocking calls from the start,
but were not introduced until **~December 2025 — UNVERIFIED, Bill would have to
check.** Do not print that date as-is.

**Write-once consequence.** Chapter 3 must ship both shapes: fast in-process
calls *and* supervised jobs. That is what CodeRhapsody actually is, and it keeps
chapter 4's mailbox purely additive, because the tools that need supervising are
exactly the slow ones.

---

## 3. The war story — RECEIPTED, two commits, both 2026-08-10

Repo `coderhapsody`, `internal/agent/watchdog.go`.

### 3a. `286bfd84` — "tool watchdog: a hung tool call must not be able to freeze the agent"

> Until now no tool call had any wall-clock bound. A handler that blocked
> forever took the whole agent with it: not crashed, not errored, just gone.
> The turn never completes, no observer fires, agents_status keeps reporting
> "processing", and a supervisor waiting on it waits forever.
>
> Hit live today by a `screenshot` call that never returned. The only recovery
> was killing CodeRhapsody -- which kills every agent in the tree, including
> sub-agents doing unrelated work. For one interactive session that costs an
> afternoon. For a fleet it is fatal, because the entire supervision model
> assumes turns end.

**The honesty beat, which is the best teaching in the whole commit:**

> Go cannot kill a goroutine. When the watchdog fires we STOP WAITING; the
> handler keeps running until it returns on its own or the process exits. This
> is containment, not cancellation, and the tool result says so in those words
> -- claiming otherwise would be exactly the "plausible default" anti-pattern we
> ban in compiler code.

Abandoned handlers are counted (`LeakedToolCalls`) so a human can decide to
restart. **What you buy is that the agent survives: a bounded, loud, recoverable
failure instead of an unbounded silent one.**

**MCP is explicitly the case inference cannot cover** — same point Bill made
tonight, written a month earlier:

> it cannot cover MCP tools at all, since those are third-party schemas we do
> not control -- and they are both the likeliest to wedge (a browser call on a
> page that never settles)

### 3b. `4bec4b96` — "watchdog: tools DECLARE their blocking contract; stop inferring it"

Same day. Two defects shipped in 3a: `send_secret` was missing from a
hand-written name→arg map and silently got the 3-minute default; and when the
deadline argument was omitted the budget fell back to 3 minutes even though the
tool's own default was longer. **Measured:** `join_agents` / `spawn_sub_agent`
(600s) and `send_message_to_parent` (300s) "would each have been abandoned
mid-wait while perfectly healthy."

**Bill's ruling:**

> conformance is the tool's job. The watchdog must not sniff schemas or parse
> English descriptions hunting for timeouts. Any tool that fails to declare its
> blocking behaviour and then trips the watchdog has a bug on its own side, and
> the loud abandonment message is the right way to learn that.

`ToolDefinition` gains `DeadlineArg` and `MaxBlockingTime`. Ten tools declare
contracts. The prose-parsing test is **deleted**; nothing at runtime reads a
description or scans a schema. This is "annotate the producer, don't detect from
the consumer."

**And the line that rhymes with this whole book's method:**

> A ruling written down and not wired up is invisible: the sentence reads like a
> reason and nothing tests it.

That is the same lesson as P9 and the chapter 1 grader audit, arrived at
independently a month earlier in a different subsystem. Worth saying out loud in
the chapter.

---

## 4. The deliberately-declined decision (P6) — CANDIDATE

**Rejected candidate:** "run tool calls in parallel or sequentially?" Bill runs
them **sequentially on purpose**. It is not an open question in this book and
pretending otherwise would contradict the author.

**Proposed candidate: what do you do when a supervised job never returns?**

It is genuinely unsettled, by Bill's own account (2026-09-13): budgets and tool
call limits are "future work"; his instinct is to **pause** the agent that hit a
limit rather than kill it, let anything waiting on it wake and continue, and let
the parent resume it with more budget.

The knob has **two opposite failure modes and no correct setting**, and we have
receipts for both:
- too short → healthy work abandoned (`join_agents` at 600s, measured)
- too long → the freeze of 3a (a `screenshot` that never returned)

So the chapter can honestly say: we do not know the right answer, here are three
(kill / abandon-and-count / pause-and-ask-the-parent), pick one and make your
event log show which you chose. Nothing to paste, because the text declines to
answer, and the decision is real rather than a riddle.

---

## 5. Open for Bill

1. Confirm the corrected thesis is how it actually went, not tidying.
2. Verify the sub-agent introduction date (~Dec 2025?).
3. The PTY trouble with `run_command` — Bill recalls "a lot of trouble" and that
   it is probably a real incident, but details need a history-log search. Not yet
   done. Memory fragment to check: PTY per **job** not per **session**, because
   shared mutable state destroys replay.
4. Does chapter 3 open on the `screenshot` hang, or on an MCP hang? The
   screenshot one is better documented; the MCP one is more representative of
   the thesis and points at chapter 5.
