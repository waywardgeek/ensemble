# Review of `chapter-02-outline.md`

> **SUPERSEDED (2026-09-12).** This reviews **Draft 3**, which was
> Anthropic-only and contained hints, interrupts, a mailbox and the
> `agent_status` tool. Draft 4 is the LLM seam chapter. The current review is
> **`book/review-ch02-draft4.md`**.
>
> This file is kept because `book/chapter-05-actors-parking.md` cites its
> findings M1, M2, M4, E4 and E7, which remain correct and are simply not due
> until Chapter 4. See `book/brief-ch02-code.md` for the applied / moved /
> obsolete disposition of every finding below.

**Reviewer:** the coder (Opus 5), writing from having built the Chapter 2
grader and the reference solution against this outline.
**Date:** 2026-09-11.
**Status of the artifact reviewed:** `book/chapter-02-outline.md`, 689 lines,
"Draft 2", all eleven open questions ruled.

This is an empirical review, not an editorial one. Everything in Part 1 is a
place where I could not build what the outline described, or where I *could*
build it but a competent student would fail for a reason the chapter never
states. Chapter 1's review earned its keep by finding four of those; this
chapter had eight, which is not a criticism — Chapter 2 is a much harder
exercise, and it caught them before a student did.

The grader and reference solution are committed at `90db7f3`. The reference
solution scores 100/100. Thirteen mutations assert the exact set of checks
each defect is supposed to break.

---

## Part 1 — Must-fix: graded, but not stated

### M1. A hint is not delivered where it arrives, and the chapter does not say so

**This is the important one.** It is the single place where I expect a careful
student, who has understood the whole chapter, to fail.

The sequence is unavoidable. The hint arrives *during* request N — that is the
definition of a hint. But it must be carried *after the tool results* of
request N+1, and those results do not exist yet at the moment it arrives. In
`Seq` order the hint sits before the `ToolCalled` and `ToolReturned` events
that will end up in the very message it is supposed to follow.

So a student who does the obvious, correct-sounding thing — play the log in
order, render the dialogue in order — puts the hint **before** the last tool
results and fails `positioned`. Their reducer is right. Their renderer is
right. Their reading of §2.2 ("ordering is primary") is right. They fail
anyway.

The resolution is a genuinely lovely piece of design and it belongs in the
prose, because it is §2.6's rule paying for itself:

> A pending hint is **not in the dialogue yet**. The context records only that
> a hint is pending. The renderer places it at the end of the final user
> message of the request it is carried in — and `RequestSent` is the event
> that moves it into the dialogue, at that same end position. From then on it
> is ordinary history and never moves again.

That gives, for free: correct position on delivery, stable history afterwards
(which prefix caching will need in a later chapter), and identical behavior on
replay. It is also the concrete answer to "the context records a fact, the
renderer decides how to carry it" — the abstract statement in §2.6a is true
but a student cannot act on it without this paragraph.

Recommend: add it to §2.6a, next to "the ordering is load-bearing".

### M2. "Delivered once" and "retained" contradict each other as written

§2.6a grades *delivered once* as "carried in exactly one request, not re-sent
every round" and *retained* as "still in the dialogue ten rounds later".

Under the hack both cannot be literally true. The hint is a text block inside a
user message; that message is history; history is re-sent in full on every
subsequent request. The words therefore appear in every later request, forever.
That is not the ephemera mistake — it is the definition of retention.

The gradeable reading, which is what I implemented:

- **delivered once** — the hint appears **at most once within any single
  request**. It is not duplicated, and it is not re-attached as a *new*
  delivery each round.
- **retained** — it is still present in a request sent several turns later,
  and it is in the log as a `MessageReceived`.

Recommend: state both in those terms. As written, a student who reasons
carefully about the contradiction will pick one and lose either way.

### M3. The log's event-type names are a graded vocabulary, but §2.3 calls them "v0"

The `interrupt`, `redaction` and `usage` checks all assert on the *contents* of
the student's dumped log: that an `Interrupted` event exists, that a
`ToolCalled` follows it, that no `ToolReturned` does, that a `Redacted` event
names its target, that `ResponseEnded` events exist. That is only possible if
the grader and the student agree on names — but §2.3 explicitly says
"names will evolve with the book, per ruling (taxonomy v0)".

I normalized aggressively: type names are compared lowercased with punctuation
stripped, so `ToolCalled`, `tool_called` and `TOOL-CALLED` are the same event,
and field lookup does the same for `target_seq` / `targetSeq`. That removes the
spelling argument. It does not remove the vocabulary requirement.

Recommend: one short paragraph in the exercise's "Log serialization" section
freezing the *set* of names for the exercise (the §2.3 taxonomy, spelled
however you like), and saying the grader matches case- and
punctuation-insensitively. Otherwise this is M-class: graded, unstated.

### M4. An interrupted round's `{"assistant"}` obligation is undefined

Chapter 1's contract is exactly one `{"assistant": ...}` per `{"user": ...}`.
Chapter 2 kills a turn in the middle. Does that round still owe its line?

The outline does not say, and the two answers desynchronize the stream against
each other — which turns one defect into a cascade of timeouts, the exact
failure mode Chapter 1's review warned about.

My rig accepts **both**: after injecting the interrupt it drains for a short
window, accepts an assistant line if one comes, and notes it without penalty.
The reference solution emits nothing, which I think is the better answer (a
killed turn produced no answer).

Recommend: state the rule — "a turn killed by an interrupt produces no
`{"assistant"}` line; the grader tolerates one if your design prefers to emit
partial text" — so the student can choose deliberately.

### M5. The request must declare the tool, and "no tool registry" may read as "no tools field"

The exercise says, emphatically, what the student is *not* building: no
registry, no dispatch table, no schemas, no argument validation. A reasonable
student may conclude the `tools` array is also out of scope. It is not — a real
Messages API never emits a `tool_use` for a tool the request did not declare,
and a submission that omits it works only through the fake's generosity.

Recommend: one sentence — "the request declares exactly one tool; the schema is
three lines and hardcoded" — inside the same paragraph that forbids the
registry, so the two are impossible to confuse.

### M6. There is no protocol check in the point table, and the protocol grew this chapter

Chapter 1 had `protocol` worth 15 points. Chapter 2 *extends* the stdio
contract with five directives and an acknowledgement, and then has no check
that the stream behaved. A submission that fails to acknowledge a directive
currently fails `logdump`, `redaction`, `ephemera`, `hint` and `interrupt`
simultaneously, and the student has to work backwards to the cause.

I added a **zero-point `session` check** — it keeps the table at exactly 100,
reports protocol violations, unexpected stdout lines and a census of every
request the fake saw (which turn, which loop iteration), and it can still sink
a submission. That census is also what found my own worst grader bug (E1).

Recommend: adopt it in the table, at 0 or at real points. If real points, take
them from `ch1parity`, which is the thing it most resembles.

### M7. `render`'s output contract is unspecified, and byte-identity is graded

The chapter says `render` prints "the vendor request JSON that *would* be
sent". Since two invocations are compared **byte for byte**, the contract needs
three more words: to **stdout**, **nothing else on stdout**, exit **0**. It
should also say that the model id comes from the environment like everywhere
else, so that "which model" is not a source of divergence.

### M8. The reducer table has two reachable gaps, in a section that claims totality

§2.3 says the reducer is total and §2.4 prints the table. Two reachable pairs
are missing:

- `ToolsPending × MessageReceived` — a human typing while tools are executing.
  Reachable in this very exercise if the grader's timing shifts slightly; it
  must be a hint, same as `InFlight`.
- `Idle × Interrupted` — interrupting when nothing is running. Must be a
  defined no-op; it is the natural companion to the `Interrupted × Interrupted`
  identity the table already shows.

Recommend: add both rows, or add a line under the table saying "every pair not
listed is identity". The second is cheaper and stronger, and it is what makes
"totality" a claim rather than a table length.

---

## Part 2 — Enrichment: things building it revealed

### E1. The "is this request mid-turn?" trap is a gift to the chapter

My grader's first serious bug. To decide whether an incoming request was a new
turn or a continuation of a tool loop, I asked the obvious question: *does it
contain `tool_result` blocks?*

That is wrong, and wrong in exactly the way this chapter is about. Once a tool
loop has happened, **every later request contains those tool results forever**,
because history is re-sent in full. My fake classified four subsequent
new-turn requests as continuations of a loop that had ended three turns
earlier. The census line showed it instantly: `7:hint-loop/cont0
8:hint-loop/cont0` for requests that were nothing of the kind.

It gets better: the fix of matching on the freshly-typed prompt is also forced
by the interrupt, because the synthesized `interrupted by user` tool_result
makes a *new* turn's request look like an answer to a dangling call.

This is a two-paragraph sidebar that teaches "full re-send is the price of
ownership" (§2.1) with a real consequence instead of an assertion, and it costs
the chapter nothing because the student's own code faces the same question.

### E2. The post-interrupt request is a better §2.6 example than the one in the outline

§2.6 describes the dangling-tool-call quirk abstractly. What actually comes out
the other end is concrete and slightly startling: the final user message of the
next request contains a synthesized `tool_result` saying "interrupted by user"
**and** the human's new question, merged into one message, because Anthropic
will not accept two user messages in a row. One rendered JSON snippet would
make §2.6 land harder than the paragraph does.

### E3. Name the four ways non-determinism gets into a renderer

`replay` is the check most likely to fail for a reason a student cannot guess.
The causes are enumerable and short: the clock, a random id, Go's randomized
map iteration order, and set iteration. The reference solution uses structs and
never maps for the wire types specifically because of the third. One sentence
in the prose converts an afternoon of bafflement into a checklist.

### E4. Render before you record

Worth stating explicitly in §2.6a: the request is rendered **first**, and the
`RequestSent` event is recorded **after**. Render is what carries the pending
hint and the pending ephemera; `RequestSent` is what consumes them. Do it in
the other order and the hint is delivered one round late — which is precisely
the one-round-lag symptom the chapter warns about under "Testing note", from a
completely different cause. A student who hits it will misdiagnose it as the
model.

### E5. Show what `agent_status` returns

`{"highest_seq":41,"turn_state":"ToolsPending"}` is three lines of code and it
makes the self-reference concrete. It is also the chapter's quiet proof that
the tool result must be a deterministic function of the log, since `replay`
byte-compares it.

### E6. Small note on `Apply`

§2.3 writes the reducer as `newContext = Apply(context, event)`. The reference
solution uses a pointer receiver that mutates in place. Value semantics over a
struct containing slices invites aliasing bugs that look exactly like replay
non-determinism, which would be a cruel thing to hand a student in the same
chapter that grades replay. Recommend the notation stay (it describes
information flow) with a half-sentence noting that in Go the natural
implementation mutates.

---

## Part 3 — Do NOT add

These protect rulings that are already correct, against improvements that look
attractive from a distance.

1. **Do not make the fake enforce hint *position* behaviorally.** The outline
   says the fake loops "until it sees that text correctly delivered". I
   implemented "until it sees that text", and grade position separately, on
   purpose. Enforcing position in the loop means a student who is 90% right —
   hint delivered, one block too early — gets an infinite loop and a bare
   timeout instead of a named property. The behavioral signal you actually want
   (drop the hint and the agent keeps running) is fully preserved by stopping
   on presence.

2. **Do not add a second tool.** One tool is the minimum that gives a turn a
   middle. Two is a dispatch table, and then it is the tools chapter. The
   reversal recorded in question 10 got this exactly right.

3. **Do not require a particular hint carriage.** The grader asserts the
   property (received while blocked, classified, positioned, once, retained)
   and accepts both the mid-turn system message and the appended text block.
   Requiring one would date the chapter to a model generation.

4. **Do not give `render` flags for model or max_tokens.** Environment only.
   The moment rendering takes CLI parameters, byte-identity becomes a property
   of how you invoked it rather than of the log.

5. **Do not lower `ch1parity` any further.** At 25 it already says what the
   reweighting note wanted it to say. Below that, a rewrite that silently
   breaks Chapter 1's contract starts to look survivable.

6. **Do not drop the `{"ok": true}` acknowledgement.** It is what turns "the
   submission hung" into "the submission hung on the redact directive". I
   would have lost an hour without it.

---

## Part 4 — Facts available to cite

**Verified by building and running** (all reproducible with `make grade2` and
`go test ./...`):

- Nine checks, summing to 100: `session` 0, `ch1parity` 25, `logdump` 5,
  `replay` 15, `redaction` 15, `ephemera` 10, `hint` 15, `interrupt` 10,
  `usage` 5.
- The reference solution's graded session: **12 requests, 52 events**, with the
  tool loop running three continuations before the hint lands.
- The hint was acknowledged **tens of microseconds** into a request the program
  had already sent and not yet received an answer to (15µs and 21µs on two
  runs; it varies, so quote the order of magnitude, not the digit). That
  measurement is the whole argument for the mailbox, and it is measured rather
  than asserted.
- **13 mutations**, each asserting the exact set of failing check ids. Full
  suite: **53 seconds**.
- One prediction was wrong: I expected a no-mailbox submission to fail `hint`
  alone. It fails `hint`, `interrupt` and `session`, because a program that
  cannot read stdin while blocked cannot receive the *interrupt* directive
  either, so the late tool call executes. The grader was right and my
  expectation was corrected — the same shape of lesson as Chapter 1's
  `twocalls` mutation.
- The Go tool ignores `testdata/`, so mutant scratch directories inside the
  module do not break `go build ./...`.

**Measured live against the real API on 2026-09-11** (four runs of
`scripts/live_hint.py`, plus a bare curl probe; logs in `/tmp/live-hint/`).
Every cell is n=1 — treat these as observations that falsify a claim, not as
characterisation of the models.

The setup: ask the model to call `agent_status` repeatedly, narrating between
calls, so the turn has a middle. Five seconds in, type
"Please speak like a pirate for the rest of this turn." Then read the dumped
log and count how many assistant messages happen between the hint being
*carried* and the model's behavior changing.

| model | carriage | first piratical reply after the hint was carried |
|---|---|---|
| `claude-sonnet-5` | appended text block (the hack) | #1 |
| `claude-sonnet-5` | mid-turn `system` message | #1 |
| `claude-opus-4-8` | appended text block (the hack) | #1 |
| `claude-opus-4-8` | mid-turn `system` message | #1 |

### M9. The model-support table in §2.6a is falsified as written — replace it with a mechanism

§2.6a states that mid-turn `system` messages are supported on
`claude-opus-4-8` and newer but **not** on `claude-sonnet-5`. That is not what
happens. `claude-sonnet-5` accepted a `{"role":"system"}` entry inside
`messages` and obeyed an instruction that appeared nowhere else in the request
— confirmed twice, once through the full rig and once through a bare curl that
returned HTTP 200 with the model replying in dialect.

A supporting inference: the trailing `system` message was accepted directly
after a user message carrying tool results. Were the API silently coercing that
role to `user`, the request would have held two consecutive user messages,
which Anthropic classically rejects. It did not. That is evidence of genuine
special handling rather than coercion — not proof, absent vendor docs.

**And the lag is not a vendor property at all.** Repeating the experiment with
the one tool patched to sleep 15 seconds:

| carriage | hint typed | hint *received* | mailbox blackout |
|---|---|---|---|
| appended text block | t+5.0s | t+16.4s | **+11.4s** |
| mid-turn `system` | t+5.0s | t+16.3s | **+11.3s** |

Identical, because the delay is not in the API. It is in the engine: the
reference solution executes its tool on the same goroutine that drains the
mailbox, so while a tool runs the actor is deaf. With Chapter 2's instant
`agent_status` this never shows; with a build or a test run it is the whole
lag. Both carriages then deliver on the first request after the tool finishes.

The only irreducible lag is the response **already in flight** when you type.
No carriage can fix that one, because the HTTP request has already left.

Recommend replacing the support table with the mechanism, which outlives any
model release: *a hint is carried by the next request the engine sends. The
hack needs somewhere to attach — the next tool results — so anything that
delays the next request delays the steer. What actually determines
responsiveness is whether your engine can still hear you while it works.*

Keep the two carriages in the chapter; they are both real and the renderer
should choose. Drop the claim about which models accept which, or re-verify it
under the author's own conditions and state those conditions.

### E7. The reference solution runs tools on the mailbox goroutine, and the grader cannot catch it

Found by the experiment above, and worth the chapter's attention because it
undercuts §2.5's own thesis: an actor whose mailbox goes deaf whenever it does
work does not really have one.

The exercise cannot detect this. Its single tool is instant *by specification*
(it must be, because `replay` byte-compares its output), so a submission that
runs tools synchronously passes every check. The grader's gate proves only that
the program stays responsive while blocked on **HTTP**, which is the easier
half.

Recommend: say so in the prose as a known boundary with a forward reference —
"Chapter 3 runs tools off the engine goroutine, and that is when the mailbox
starts earning its keep" — rather than complicating this chapter's listing with
tool cancellation. I did not change the reference solution, because the fix
belongs with the tool loop and this is not the tool chapter. That is a judgement
call and it is reversible.

**Still NOT verified, and still resting on the author's direct experience:**

- The **July 2025 priority claim** for mid-turn hints. Unchanged, and outside
  what any experiment here could settle.
- The author's own observation of a lag that disappeared on switching to the
  `system` carriage. Nothing above contradicts it — the runs here reproduce a
  multi-second lag with exactly the right shape, but locate its cause in the
  engine rather than the API. If CodeRhapsody executes tools off its main loop,
  then its lag has a different cause and is worth finding before print.

Model **ids** were confirmed against `GET /v1/models`; both
`claude-opus-4-8` and `claude-sonnet-5` exist.
