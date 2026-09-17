# Parking file — material for Chapter 6 (the agent framework / two seams and a loop)

**Provenance:** these sections were written for Chapter 2 and survived Opus 5's review
(findings M1, M2, M4, M8, E4, E7). When Chapter 2 became the LLM-seam chapter, hints
and interrupts moved here, where turns have real middles because Chapter 4 gave them
real tools. Nothing here is retracted — it is relocated. The `agent_status` toy tool
is expected to DISSOLVE in the move: it existed only to give a Chapter 2 turn a middle.

**Note (2026-09-17):** This was originally labeled "Chapter 5" parking. The
refactoring chapter is now Chapter 5; the actor/framework chapter is Chapter 6.
§2.6a (hints) and the deafness analysis are core Chapter 6 material. §2.8
(stateful actors / deployment topology) may belong in a LATER chapter (gateway
or deployment) — flag for Bill's ruling.

---

## §2.6a Real-time hints — steering an agent mid-turn

*(Author note: likely renumbers to §2.7 when prose is written. Placed here
deliberately — it is the renderer section's payoff and must follow it.)*

The agent is three tool calls into a refactor and heading somewhere you don't
want it to go. You type: *"stop, the bug is in the parser."* Those words reach
the model **during** the current turn — not after it finishes, not as the next
question. The agent pivots mid-chain, without the tool chain breaking.

This is the capability Chapter 1 §1.1 named first when it argued that
frameworks hard-code delivery. Here the student builds it.

**Why this chapter needs one tool.** A hint is a message that arrives *during*
a turn — so a turn has to be long enough to have a middle. With no tools a
turn is a single round trip: request, response, done. There is no middle, and
mid-turn steering degenerates into "your next question." That is why the
exercise hardcodes exactly one trivial tool (`agent_status`, below). It is
not a tool chapter; it is the smallest possible thing that makes a turn last
long enough for a human to interrupt it.

**The demonstration.** The model calls `agent_status`. You return the status.
It calls it again. And again — a loop that will happily run until something
stops it. Mid-loop, you type:

> **please stop**

The words reach the model *inside* the turn, attached to the very next tool
result. The loop ends. You did not kill the process, you did not wait for
your turn — you steered a running agent, and you built every part of the path
those words travelled.

In the REPL, try **"Please speak like a pirate"** mid-turn instead, and watch
the rest of the turn come back in pirate. Same mechanism, more fun, and it
makes the timing vivid: everything before the hint is normal, everything
after is piratical, in a single unbroken turn.

**Provenance.** As far as we can determine, this book's author and his agent
invented this in **July 2025** — the first known implementation of mid-turn
human steering in an agentic loop. It came from noticing that the Messages API
would accept a user `text` block in the *same message* as `tool_result`
blocks, which meant there was a place to put words that the model would read
before deciding its next move. Frontier products have since shipped
steering in their own UIs, and the API has since grown a first-class mechanism
for it (below). *(Publication note: the priority claim is the author's, dated
and falsifiable. Verify before print; "as far as we can determine" stays
either way.)*

### Why it belongs in this chapter, and not in the chapter about tools

Because the hard part is not the network call — it is that a hint is
**the same event as a prompt, distinguished only by state.**

§2.3's reducer table already says so: `InFlight × MessageReceived → InFlight`,
classified as a hint. There is no `Hint` event type and there must not be one.
The identical bytes typed by the identical human are a *prompt* when the turn
is `Idle` and a *hint* when it is `InFlight`. Only the reducer knows which,
because only the reducer holds the state. Classify at capture time and you
will be wrong every time the human types fast.

The second half is delivery, and delivery is the **renderer's** problem —
which is what makes this the cleanest demonstration of §2.6's rule. The
context records a fact about the conversation: *a hint is pending, not yet
carried to the model.* It does not record how to carry it. That separation is
not an aesthetic preference; it is load-bearing, because there are currently
**two** ways to carry a hint and which one you use depends on the model:

| carriage | how it is carried | availability |
|---|---|---|
| **Mid-turn `system` message** | a `{"role":"system"}` entry inside `messages`, placed after the current tool results | accepted by every current Claude model we have tested, `claude-sonnet-5` and `claude-opus-4-8` included |
| **The appended text block** (the original 2025 hack) | the hint as a `text` block placed **after** the `tool_result` blocks in the same user message | everywhere, by construction — it is an ordinary user message |

Both are legal, both steer the model, and **the renderer chooses**. Notice what
supporting both costs: a renderer that knows which carriage it is using, and a
context that never had to care. Put the carriage decision inside your HTTP
code, where it will feel natural, and you can serve exactly one shape of model.

> **A note on what this table used to say.** An earlier draft claimed the
> mid-turn `system` message worked on `claude-opus-4-8` and newer but *not* on
> `claude-sonnet-5`, and that the hack cost you a round of lag. Measured
> against the live API, that is wrong on both counts: `claude-sonnet-5`
> accepts the mid-turn `system` entry and obeys an instruction that appears
> nowhere else in the request, and both carriages steer the model on the very
> first response after the hint is carried. The lag was real. It was not where
> we thought. Chasing it down is the rest of this section — and the reason the
> table above deliberately does not tell you which models support what. Any
> such table is a fact about one week in the history of an API. The mechanism
> below outlives it.

**The ordering is load-bearing.** In the hack, the hint text goes *after* the
tool results in the message. A hint placed before them reads as a comment on
nothing — the model has not yet seen what it is being steered about. This is a
real rule with a real failure mode, it is graded, and it is the reason
`agent_status` had to exist at all: an ordering rule needs two things to order.

**A pending hint is not in the dialogue yet.** This is the one place where a
student who has understood everything else will still fail, so it is worth
being slow and explicit.

Look at the timing. The hint arrives *during* request N — that is the
definition of a hint. But it must be carried *after the tool results* of
request N+1, and at the moment it arrives, those results **do not exist yet**.
In `Seq` order, the hint sits before the `ToolCalled` and `ToolReturned` events
of the very message it is supposed to follow.

So the obvious, correct-sounding thing — play the log in order, render the
dialogue in order — puts the hint in the wrong place. The student's reducer is
right. Their renderer is right. Their reading of §2.2 is right. The result is
wrong.

The resolution is §2.6's rule paying for itself in cash:

> A pending hint is **not in the dialogue**. The context records only that a
> hint is *pending*. The renderer places it at the end of the final user
> message of whichever request carries it — and `RequestSent` is the event
> that moves it into the dialogue, at that same end position. From that moment
> it is ordinary history and never moves again.

That single decision buys three things at once: the hint lands in the right
position on delivery, the history is stable afterwards (which prefix caching
will demand in a later chapter), and replay puts it in exactly the same place
every time.

**Render, then record.** A direct corollary, and worth stating as a rule
because the failure is so well disguised: the request is rendered **first**,
and `RequestSent` is recorded **after**. Rendering is what *carries* the
pending hint and the pending ephemera; `RequestSent` is what *consumes* them.
Do it in the other order and the hint goes out one round late — which is
indistinguishable, from the outside, from a model that lags. A student who hits
this will blame the vendor. It is in their own engine, four lines apart.

**A hint is history, not ephemera.** This is the distinction students most
reliably get backwards, and the two graded properties need saying precisely,
because read casually they contradict each other:

- **Delivered once** means the hint appears **at most once within any single
  request** — not duplicated inside a request, and not re-attached as a *fresh*
  delivery on each subsequent round.
- **Retained** means that once delivered it stays in the dialogue, and is still
  there in a request sent many turns later.

Both are true simultaneously because the hint becomes an ordinary part of a
user message, and history is re-sent in full every round (§2.1). Its words
therefore *do* appear in every later request, forever — and that is not the
ephemera mistake, that is the definition of retention. Contrast a stale
timestamp, which must vanish after one delivery because on the next round it
is simply a lie. **A hint is delivered once and remembered always.**

**What actually determines responsiveness.** Here is the mechanism the model
table cannot give you:

> A hint is carried by **the next request the engine sends**. Anything that
> delays that request delays the steer.

That reframes the question from "which carriage?" to "what is my engine doing
right now?", and there are only three answers:

1. **A response is already in flight when you type.** Irreducible. The HTTP
   request has left; nothing can overtake it. You wait for that one response.
   No carriage fixes this, and no carriage needs to — it is one response, not
   one round.
2. **No further request is coming.** If the turn is ending, there is no next
   request to carry anything, and the hint becomes what it always was for a
   plain chatbot: your next message.
3. **The engine cannot hear you while it works.** This is the big one, and it
   is the one that masquerades as vendor lag.

That third case is measurable. Patch the chapter's one tool to sleep fifteen
seconds and type a hint five seconds in:

| carriage | hint typed | hint *received* | mailbox blackout |
|---|---|---|---|
| appended text block | t+5.0s | t+16.4s | **+11.4s** |
| mid-turn `system` | t+5.0s | t+16.3s | **+11.3s** |

Identical, because the delay is not in the API at all. It is in the engine: if
tools execute on the same goroutine that drains the mailbox, then for the whole
duration of a tool the actor is **deaf**. Both carriages then deliver on the
first request after the tool returns, promptly and equally.

(These are single runs against one vendor on one day, quoted to show a shape,
not to characterise a model. Re-measure before you trust any digit here.)

Which lands us somewhere better than a compatibility table. An actor whose
mailbox goes deaf whenever it does work does not really have a mailbox — it has
an inbox it checks between chores. Chapter 2's tool is instant, so the chapter
cannot show you this failure. Chapter 4's tools are not, and that is when the
mailbox starts to earn its keep.


---

## §2.8 Actors are stateful — deployment follows the data structures

- The context is live, expensive-to-rebuild state. Where it LIVES is a
  property of the data structures, not a free deployment choice.
- Two topologies seen in the wild:
  1. **Stateless workers**: every message rehydrates the full conversation
     from durable transactional storage, processes, writes back. (Some
     frameworks require this.)
  2. **Stateful actors** (ruling: preferred): the conversation lives in
     memory on one server; the load balancer routes a conversation's
     messages consistently to the same machine.
- Why stateful wins: no per-message rehydration cost — and, decisively,
  **ancillary local state**. Example: a tool result too large for the
  context is not dropped; it's kept on local disk, a stub goes in the
  context, and the LLM can scan the full output later via tool calls. With
  random routing, the file is on the wrong machine and this pattern —
  one of the most valuable in a real coding agent — becomes a distributed-
  systems problem.
- **Actors persist. They migrate, and migration is hard** — snapshot the
  context + blob table + ancillary files, move, reattach — but that is the
  primitive worth building, not a reason to go stateless.
- Local single-machine deployment (how this course builds) is the
  degenerate case where stickiness is free. The cloud consequences are
  named here so the student knows what the local design is secretly
  deciding.

