# Chapter 2 — One Log, Three Vendors

*The data structures, and the seam they exist to make possible.*

**Status:** Draft 4, 2026-09-12. Outline only; prose not yet written.

---

## Draft 4 — what changed and why

Draft 3 was "the real data structures," with the vendor seam present as an
*idea* and the exercise targeting Anthropic alone. That is the same mistake the
chapter now opens by describing: a seam with one implementation is not a seam,
it is a naming convention. Draft 4 makes the seam the subject, and makes the
student build three renderers and three parsers over one context.

Three structural changes:

1. **The LLM seam is the chapter.** The data structures are motivated *by* the
   seam rather than the seam being one of their benefits. Three vendors —
   Anthropic, OpenAI, Gemini — request rendering **and** response parsing.
2. **Hints and interrupts moved to Chapter 5.** They are about time and
   concurrency, not about vendors. Nothing is retracted; the material is
   preserved verbatim in `book/chapter-05-actors-parking.md`, including review
   findings M1, M2, M4, E4 and E7.
3. **The `agent_status` toy tool is gone.** It existed only so a Chapter 2 turn
   would have a middle for a hint to land in. With tools in Chapter 3 and hints
   in Chapter 5, the constraint that forced it no longer exists. Chapter 2
   *renders* logs containing tool events without executing any.

**The write-once constraint now governs the book.** Chapter 1 is the single
sacrificial chapter. From Chapter 2 on, every chapter is strictly additive —
new events, new tools, new seams, never "delete what you built." This is why
Draft 4 exists at all: under write-once, whatever Chapter 2 gets wrong is
inherited by every chapter after it.

**The escape hatch, stated plainly to the reader.** Write-once binds *us*, not
you. Every chapter's reference solution is public from day one — read it
whenever you like, and **start any chapter from ours instead of your own.**
Failing a check in Chapter 2 must never end your course in Chapter 4. This is
not a grudging concession — it is what makes write-once safe to promise.
Additive means the book never demolishes code you wrote; rebasing means one bad
structural choice never strands you. The graders cooperate: they run your
binary and read what it emits, never your source and never its history. No
check asks whose code it is.

**Claim status:**

- **FALSIFIED and removed** (Draft 3, finding M9): the claim that mid-turn
  `system` messages work on `claude-opus-4-8`+ but not `claude-sonnet-5`.
  Measured live; Sonnet 5 accepts and obeys them. That material now lives in
  the Chapter 5 parking file with the correction applied.
- **STILL UNVERIFIED:** the July 2025 priority claim for mid-turn hints — now
  a Chapter 5 problem, not this chapter's.
- **NEEDS VERIFICATION BEFORE PRINT:** every wire-format detail in §2.6. The
  three request and response shapes are quoted from working knowledge and must
  be checked against current vendor documentation by the coder. Wire formats
  drift, and this chapter is nothing but wire formats.

---

## §2.0 Cold open — the seam I got wrong, and what it cost

Open with my own failure, told plainly, because it is the most expensive
mistake in this book and it looks completely reasonable while you are making
it.

The sequence:

1. Build an agent against Anthropic. It works.
2. Extract an interface, `AIClientInterface`, with all the methods needed,
   as they were needed, by `ClaudeClient`.
3. Add a second vendor. The interface does not fit, because it was never
   vendor-shaped; it was Claude-shaped with an interface keyword in front of
   it. So: copy `ClaudeClient`, paste, edit until Gemini works.
4. Add a third. Copy, paste, edit until OpenAI works.
5. Discover, roughly 30,000 lines later, that three near-identical clients
   drift independently, that every bug must be fixed three times, and that
   two of the three fixes will be forgotten.

**The remedy was worse than the disease.**
The correct seam was eventually designed (one context, one renderer per
vendor, one parser per vendor) and delivered as a big-bang rewrite. A year
later the migration is still not finished. The product works. It is also
semi-broken in ways I have not finished cataloging, and some bugs have
not yet been reported to anyone, including me.

The lesson has two halves:

> **The seam was right. Shipping it as a rewrite was the mistake.**
> Cutting a seam late does not cost you one refactor. It costs you a tail —
> and the tail is paid by whoever is using the product while you migrate.

**This is why the chapter charges an hour for something a reader would rather
skip.** Writing three renderers on day one feels like over-engineering. It is
the cheapest hour in the book: it buys the shape of an interface that was
*derived from three implementations* instead of extrapolated from one.

### The demolition (carried forward from Draft 3)

Chapter 1 ended with a working agent and a deliberate attachment to it. This
section takes it apart. The `[]{role, content}` array cannot express: what the
model actually returned versus what we chose to send, tool calls and their
results, a redaction, token accounting, or who said a thing and why. Chapter 1
was "total garbage, but the student learns the basics" — and the student now
knows enough to see why.

State the promise honestly, because Draft 3 learned this the hard way when
scope changed: **the rewrite is observably identical for everything Chapter 1
could already do.** It then gains something Chapter 1 could not express at any
price — the same conversation, correctly, to three different vendors.

### The contract, stated once

This is the last time you will be asked to throw code away.

Chapter 1 was sacrificial on purpose: you had to feel a wrong data structure
fail before a right one could mean anything. **From here on, every chapter is
additive.** New events, new tools, new seams — but nothing in this chapter gets
deleted in the next one, or the one after that.

That is a promise with a practical consequence, and it is the reason to state
it rather than let the reader infer it: **build the simplest thing that
satisfies this chapter.** Do not leave room for the tool loop. Do not
generalize for concurrency. Do not invent a plugin system for skills. Those are
all coming, and the chapters are sequenced so that each arrives *before* the
weight that would have made it painful — which is precisely the lesson §2.0
just paid 30,000 lines to learn.

A reader who does not trust this promise will over-engineer defensively, and
defensive over-engineering is the failure mode this book argues against
everywhere else. If a later chapter makes you delete something
from this one, that is our bug, not yours.

### "But I only use one vendor"

The obvious objection, and the honest answer is not the one the chapter title
suggests. This seam is not mainly for people running three models. It is for
people running one.

You do not have to add a competitor for the wire format underneath you to
change. **As of September 2026, the Gemini API this project uses is
deprecated, and its replacement, the Interactions API, is not yet available on
Vertex AI, which is the access path many corporate users are required to
take.** The old surface is marked for removal and the new one cannot be
reached from where they stand. That is not a hypothetical migration used to
motivate a design. It is a live one, and nobody involved chose it.

With a seam, that is an afternoon: write a renderer for the new surface, keep
the old one until it dies, switch on a field, and let the logs replay
unchanged. Without one (with a data structure shaped like a particular
vendor's request body, which is §2.0's mistake) it is a rewrite, and §2.0 has
already told you how those go.

This is also why `Provenance` records a **surface** (§2.4a). The vendor is not
the unit of compatibility. One company, one model, two incompatible surfaces
is an ordinary Tuesday, and replayed material is bound to the surface that
produced it. `SurfaceInteractions` is in the enum because this was
foreseeable.

> *Authorial note: keep this receipt-shaped and dated, per `voice.md`. The
> verifiable facts, a deprecated surface and a replacement absent from Vertex
> AI, are devastating on their own and need no help. Name the API, never the
> company — the public record carries the claim, and a reader can check it
> without trusting us.*

---

## §2.1 History ≠ Context

- **History** is an append-only event log. What happened, in order, forever.
  It is the truth and it is never edited.
- **Context** is the vendor-independent state you get by replaying that log.
  It is derived, reconstructible, and disposable.
- **The request** is what a *renderer* makes from context for one specific
  vendor. It is disposable and it is a lie by omission — necessarily.

Three consumers, two needs: the renderer reads the context; the GUI and the
auditor read the log.

**Full re-send is the price of ownership.** Every request carries the entire
conversation. You pay for it in tokens (largely refunded by prefix caching, a
later chapter) and you buy the ability to edit history, which is the core
capability of a coding agent and the reason §2.6 declines vendor stateful
conversation APIs.

> **Sidebar: "Is this request mid-turn?" — a bug from building this chapter.**
>
> Draft 3's grader needed to know whether an incoming request began a new turn
> or continued a tool loop. The obvious test: *does it contain `tool_result`
> blocks?*
>
> Wrong, and wrong in exactly the way this section is about. Once a tool loop
> has happened, **every later request contains those tool results forever**,
> because history is re-sent in full. Four new-turn requests were classified as
> continuations of a loop that had ended three turns earlier.
>
> The general lesson: **a request is not a description of the current moment.**
> It is the entire history, re-sent, with a little new material on the end. Any
> question shaped like "what is happening right now?" must be asked of the
> *end* of the request, or of the log — never of the whole.

---

## §2.2 The event log

- Append-only. Monotonic `Seq`. **Ordering is primary**; wall-clock time is
  metadata and may be wrong, duplicated, or non-monotonic across machines.
- Never edited, never reordered, never deleted in place. A redaction is a new
  event that supersedes, not a mutation of an old one (§2.6).
- Serialized as JSON-lines so it is greppable with ordinary tools, a property
  that becomes load-bearing in Chapter 4, when tool output starts arriving by
  the megabyte.

---

## §2.3 Events

### The self-contained event rule (load-bearing)

Given the current context and just the next event (nothing else) we can
correctly compute the new context. Write the rule as:

```
newContext = Apply(context, event)
```

That notation describes **information flow**. It is the chapter's central
claim: everything needed to advance the context is in the context plus one
event. It is *not* a demand for value semantics. In Go the natural
implementation is a pointer receiver mutating in place, and that is the one to
write: a `Context` full of slices copied by value gives you two contexts
sharing one backing array, and that bug is indistinguishable from renderer
non-determinism (§2.7).

**Classification is the reducer's job, not the capture site's.** The same
arriving bytes mean different things depending on turn state. Decide at capture
time and you are wrong every time the human types quickly. Only the reducer
holds the state that makes the decision correct. (Chapter 5 makes this vivid:
the *same* event is a prompt or a hint depending solely on turn state.)

### The taxonomy for this chapter

`MessageReceived`, `RequestSent`, `ResponseStarted`, `ResponseEnded`,
`ToolCalled`, `ToolReturned`, `Redacted`, `ErrorOccurred`.

Chapter 5 adds `Interrupted`. Chapter 4 adds job events. **Additive, always**:
this is the first place the write-once discipline is visible to the reader.

Three notes:

- **A single `ErrorOccurred`.** Infrastructure errors change turn state;
  semantic errors (a tool that ran and failed) are ordinary tool *content*.
  Conflating them is why agents get stuck retrying a compile error as though it
  were a network outage.
- **Thinking text is log-only.** The context carries opaque replay material
  (a signature, a redacted block, an id) tagged with the exact model that
  produced it, and the renderer decides whether that model wants it back. Never
  reconstruct reasoning as prose and feed it to a different model as though it
  were your own.
- **`ResponseEnded` carries the content; `ToolCalled` records the dispatch.**
  Say this explicitly, because two coherent readings exist and they are not
  compatible. `ResponseEnded.Parts` holds everything the assistant produced:
  text and `ToolCallPart`s together, in the order it produced them. `ToolCalled`
  is an **engine** event: it contributes no dialogue content and records that a
  call was actually dispatched, so that Chapter 4 can time one and Chapter 5 can
  cancel one. The alternative (a `ToolCalled` per call, with `ResponseEnded`
  carrying only text) throws away the ordering of text relative to calls within
  a single turn, which is why it is not what we do. This is also how to read
  `InFlight × ResponseEnded (tool calls)` in the table below: *inspect the
  response's parts.* Chapter 3 inherits this choice under write-once, so it
  belongs here rather than in an implementer's head.

### Tool events without a tool loop

Chapter 2 executes no tools. It **renders logs that contain tool events**,
supplied by the exercise. The student therefore writes a reducer that handles
events it cannot yet produce.

That is deliberate. You are building the shape before the capability, because
the shape determines whether the capability can be added without a rewrite.

### Turn states

`Idle`, `InputPending`, `InFlight`, `ToolsPending`. (`Interrupted` arrives in
Chapter 5; and it must be a *state*, not a flag, or replay re-executes tool
calls that were canceled.)

| transition | result | note |
|---|---|---|
| Idle × MessageReceived | InputPending | ordinary prompt |
| InputPending × RequestSent | InFlight | |
| InFlight × ResponseEnded (tool calls) | ToolsPending | |
| InFlight × ResponseEnded (no tool calls) | Idle | turn complete |
| ToolsPending × ToolReturned (last) | InputPending | loop continues |
| InFlight × ErrorOccurred | Idle | infrastructure failure ends the turn |

**Every pair not listed is identity.** That sentence, not the length of the
table, is what makes the reducer total. A table enumerates the transitions we
thought of; the default covers the ones we did not. Write it as the `default`
arm of the switch, **not** as a `panic`.

---

## §2.4 The context

The context is **vendor-independent by construction**, and this chapter is the
only one that can prove it.

Contents: the dialogue (ordered, actor-attributed, parts-structured), pending
ephemera, token accounting, and opaque per-model replay
material carried but never interpreted. Redacted content is not tracked
separately; it is replaced in place by the reducer (§2.4a).

**Content is Parts, not a string.** Text, tool calls, tool results, images,
audio, and vendor-opaque blobs. A string is the Chapter 1 mistake wearing a
struct.

### The system prompt is rendered, not stored

*The inoculation. It teaches no prompt content and prevents the most common
architectural mess in the field.*

The system prompt is **output of the renderer**, computed from context plus
configuration. It is not a blob of text living in the log, and it is not a
field on the context that someone appends to.

For now, a constant string is a perfectly good renderer. The rule is only about
**where it comes from**.

The system prompt is the easiest surface in an agent to abuse, and the abuse
has a predictable
shape. First someone describes the tools in it by hand. Then the descriptions
drift from the actual tools. Then part of it is generated and part is
hand-written, and no one can say which. By the time it is 400 lines nobody will
delete a word, because nobody can prove which words are load-bearing.

Chapter 6 replaces the constant with generation from skills. Under write-once
that must be a pure addition — and it is, **provided the system prompt was
never a stored value in the first place.**

The three vendors make the point concrete before the reader can form a bad
habit: Anthropic takes a top-level `system` parameter, OpenAI takes a `system`
(or `developer`) message inside the array, Gemini takes a separate
`systemInstruction` object. One fact; three placements; a renderer's problem.
Store it and you have just picked a vendor.

---

## §2.4a The types, in one place

*Gathered here so they can be reviewed as a set. In the prose each type is
introduced where it is motivated — this section is the contract with the
grader and the reference solution.*

Shapes, not implementations. Field names are illustrative; the grader
normalizes names case- and punctuation-insensitively, and must never grade on
Go identifiers.

### The log

```go
type Seq uint64

type Event struct {
    Seq  Seq
    Type EventType
    Time time.Time // METADATA. Ordering comes from Seq, never from this.

    // Exactly one is non-nil, selected by Type. Verbose on purpose:
    // it round-trips as JSON with no registry, and it makes the
    // reducer's switch exhaustive by construction.
    Message  *MessageData
    Request  *RequestData
    Response *ResponseData
    Tool     *ToolData
    Redact   *RedactData
    Error    *ErrorData
}

type MessageData struct {
    Actor Actor
    Parts []Part
}

type ResponseData struct {
    Parts []Part
    Usage Usage
    From  Provenance // NOT a vendor string. See below — this is load-bearing.
}

// Provenance records who produced content. Recorded at WRITE time by the
// client that produced it, NEVER inferred later.
type Provenance struct {
    Vendor  Vendor
    Model   string // OPEN set. New models ship weekly. Never switch on it.
    Surface Surface
}

type Vendor uint8
type Surface uint8

// iota + 1 on purpose: the zero value is INVALID, so a Provenance that was
// never populated is detectable instead of silently meaning "Anthropic".
const (
    VendorAnthropic Vendor = iota + 1
    VendorGemini
    VendorOpenAI
)

const (
    SurfaceMessages Surface = iota + 1 // Anthropic
    SurfaceChatCompletions             // OpenAI — what Exhibits A–C speak
    SurfaceGenerateContent             // Gemini — what Exhibits A–C speak
    SurfaceInteractions                // Gemini's replacement surface
    SurfaceResponses                   // OpenAI's newer surface
)

type RedactData struct {
    From, To    Seq        // the SPAN superseded, inclusive. Not one event.
    Level       Redaction  // what, within that span, is removed
    Replacement []Part     // RedactSummary only; purges synthesize their stubs
    Reason      string
}

// Redaction levels, weakest first. Ch2 exercises only RedactResult; the rest
// arrive with context engineering. Span says WHERE, level says WHAT.
type Redaction uint8
const (
    RedactResult   Redaction = iota + 1 // tool result content -> stub; the call survives
    RedactTool                          // call and result both go; visible reasoning survives
    RedactDialogue                      // prose and reasoning go; survivors are defined by the compaction policy
    RedactSummary                       // the whole span collapses to ONE entry of compressed prose
)
```

### Content

```go
type Part interface{ isPart() }

type TextPart   struct{ Text string }
type BlobPart   struct{ MIME, Path string } // on disk, never inline
type OpaquePart struct{ From Provenance; Data json.RawMessage }

// The RESULT of applying a Redacted event. Replaces the parts it supersedes;
// nothing records "a redaction happened" separately — the log already does.
// The stub is informative on purpose: what it was, and how to get it back.
type RedactedPart struct{ Stub string }

type ToolCallPart struct {
    CallID string // the ID AS ISSUED, by the model named in From
    From   Provenance
    Name   string
    Args   json.RawMessage
    Opaque json.RawMessage // replay material bound to THIS CALL
}

type ToolResultPart struct {
    CallID  string
    Parts   []Part
    IsError bool // a tool that ran and failed is CONTENT, not ErrorOccurred
}
```

**`ToolCallPart.Opaque` is the field this chapter did not want to need.**
`OpaquePart` above is standalone: it floats in the parts list, associated with
nothing. That is right for a thinking block, which belongs to the turn. It is
useless for Gemini's `thoughtSignature`, which arrives as a *sibling key of
`functionCall`* and is bound to that one call. A context with nowhere to put
per-call replay material cannot produce a valid Gemini 3.x request after a tool
call at all — the API answers 400. So the field is here. §2.6 tells the story of
how it got here, because the chapter bet against needing it and lost.

The naive version of this field is `Vendor string`. That is wrong, and it is
wrong in a way you will not discover until a user switches models mid
conversation.

**Thinking signatures, the encrypted reasoning material, are bound to the
model, not the vendor.** But *what* each vendor validates, and how it fails, is
not the same thing — and that difference is worth more than the simpler table
it replaces.

Measured 2026-09-12: signatures harvested from four Gemini models and replayed
across all sixteen pairings were **accepted without error, 16 of 16**. So
"Gemini rejects another model's thinking" is simply false. What Gemini rejects
is a signature that is **corrupt**, or one **missing** from a replayed
`functionCall` (Gemini 3.x; 2.5 returns 200). Anthropic's drop is real but
**directional**: it reads its own thinking and that of *earlier* models, and
silently discards a *newer* model's, while returning 400 for one that has been
modified.

| vendor | what it validates | how it fails |
|---|---|---|
| Gemini | signature **integrity** | loud — 400 on corrupt or missing |
| Anthropic | model **binding** | quiet — drops what this model cannot read |

The contrast is therefore about **integrity, not authorship** — and the rule of
this book survives it intact: **the loud failure is the good one.** Gemini's 400
costs you an afternoon. Anthropic's silence costs you a subtly worse agent that
still passes every test: reasoning quietly discarded, nothing in the logs, no
way to tell from the outside. Exactly the shape §2.6 forbids when it insists
that media asymmetry must raise rather than drop.

*(16 of 16 is HTTP-level acceptance. Whether the Gemini backend* honors *a foreign
signature is not observable from outside the API, so the claim in print is
"accepted without error" — never "honored".)*

And `Surface` earns its place for the same reason. Signature validity is scoped
to *(vendor, model, surface)*:

- Gemini **Interactions** attaches signatures to thought steps and built-in
  tool steps, but never to standard function calls.
- Gemini legacy **generateContent** returns a 400 if you replay a
  `functionCall` *without* its signature.

So a coarse vendor tag cannot even decide whether to *include* the material,
let alone whether it is replayable. **Without the exact model, a Gemini
renderer cannot construct a valid request at all.**

Hence the rule, which is about capture rather than rendering:

> Provenance is recorded at **write time**, by the client that produced the
> content, and is never inferred afterwards. A renderer may read it. Nothing
> may reconstruct it.

Inference is impossible in principle here: by the time you are rendering, the
model that produced a signature three turns ago is simply not derivable from
anything else in the context. Miss it at capture and the information is gone.

### Enum or string? The rule, and why `Model` is the odd one out

Two of `Provenance`'s three fields are enums and one is a string, which looks
inconsistent until you have the rule:

> **Enum when the code must exhaustively handle every case. String when the
> value is only compared for equality and the set is open.**

`Vendor` and `Surface` are closed sets the renderer *switches on*: there is
exactly one renderer and parser per vendor, compiled in. A typo like
`"Messages"` in a string field is a runtime surprise; as an enum it does not
compile. `Model` is the opposite: an open set that gains members weekly, never
switched on, only ever compared — *is this the same model that issued that
signature?* Make it an enum and you need a rebuild to record a model you have
no other opinion about.

Two details that are easy to get wrong:

**Start the constants at `iota + 1`.** The zero value must be invalid.
Otherwise a `Provenance` that nobody populated is indistinguishable from a
genuine Anthropic/messages one. And since provenance must be captured at write
time and can never be reconstructed, "nobody populated it" is precisely the bug
you need to be loud. A zero value that silently means something is a default
wearing a disguise.

**Marshal them as readable strings, and refuse unknown ones on the way in.**
The log is JSON-lines so that ordinary tools can grep it (§2.2); `"vendor":2`
destroys that for no gain. An unrecognized surface on read is a **loud
refusal**, exactly as §2.7 requires for an unknown event type. Not a default,
not a skip — the same discipline, for the same reason: a value you silently
coerce is a value you will debug in production.

### The tool-call id, and where it collides with replay

The same field carries a second load. A tool-call id is the one piece of vendor
vocabulary that legitimately enters the context: you cannot answer a call
without quoting the id that made it.

Render a conversation holding an Anthropic `toolu_…` id to OpenAI and it is
meaningless, so the renderer must synthesize one. It must be **derived from
`Seq`**, not generated randomly, because `replay` compares bytes and a random
id is one of the four non-determinism sources named in §2.7. The seam and the
determinism rule meet at this field, and students who wire them up
independently will collide.

### The context

```go
type Context struct {
    Turn     TurnState
    Dialogue []Entry
    Ephemera []Part // pending; delivered once, then cleared
    Usage    Usage  // running totals, vendor-normalized
}

// Token counts, NEVER money. All four fields are DISJOINT — they sum to the
// billable total. Vendors disagree about whether that is true of their own
// reporting; making it true is the parser's job.
type Usage struct {
    Input      int // input that was neither read from nor written to cache
    CacheWrite int // cache creation. Typically costs MORE than plain input.
    CacheRead  int // typically costs an order of magnitude LESS.
    Output     int
}

type Entry struct {
    Seq   Seq // log position of the event that produced this entry
    Actor Actor
    Parts []Part
}
```

**`Entry.Seq` is not decoration.** `RedactData` names a *span* of `Seq` numbers,
so an entry carrying no `Seq` gives a redaction nothing to match against and the
entire family becomes inapplicable. The reflex fix, when you hit this while
building, is a `map[Seq]bool` kept off to the side, which is precisely the
unbounded field the next rule exists to delete. One fixed-size field per entry
costs nothing, and entries are already bounded by the compaction policy. Watch
for that pattern: the rule you are about to read will try to reassert itself in
disguise, and it will look like a reasonable local fix every time.

**No field in the context may grow without bound.** State it as a rule, because
it is cheap to honor now and very expensive to retrofit.

The context is not a request buffer. It is the current state of an actor that
may run for **years**: memory, identity, recent conversation, everything the
model knows about itself. Anything in it that only ever accumulates is a slow
leak with a long fuse, and the fuse burns in production, on the agent you care
most about, long after the design decision is unrecoverable.

An earlier draft of this chapter had `Redacted map[Seq]bool` in the context, to
remember which events had been superseded. It is a natural thing to write and
it is wrong twice over:

- **It grows forever.** One entry per redaction, retained for the life of the
  actor, and nothing ever removes them.
- **It is redundant.** The log already records every `Redacted` event,
  permanently. The context does not need to remember that a redaction
  *happened*; it needs to hold the content that redaction *produced*.

Deleting it removes a field and a failure mode at the same time. A redaction is not metadata about
content — **it is content**, and the context holds the result of replaying the
log, exactly as §2.1 promised.

Apply the same lens to every field and one survivor stands out: `Dialogue`
grows too. It is bounded by a *policy* (compaction and retention, a later
chapter) rather than by its shape, and the shape survives compaction unchanged
because compaction replaces entries with a summary entry. That is the
distinction to hold: `Dialogue` grows and has a plan; the map grew and had
none.

### Redaction is a family, not a flag

The shape above (a **span**, a **level**, and an optional replacement) looks
like over-modeling for a chapter that only ever stubs a tool result. It is
here because the alternative is demolishing it later, and because the thing it
grows into is the mechanism that keeps an agent alive past its context window.

The naive design is `Target Seq` plus a boolean: this event was redacted. It
cannot express "remove every tool call and result older than the last
`save_memory`," which is the compaction this chapter's students will reach
for first.

**Compaction by position versus compaction by category.** The common framework
approach — one widely used agent SDK does this — is to replace the oldest *portion* of
history with an LLM-written summary. That is compaction by **position**: it
discards whatever happens to be old, valuable or not, and what it loses is
unpredictable, because a summary is lossy in ways nobody enumerated.

Compaction by **category** discards a *kind* of content wherever it appears.
And the categories are wildly unequal. Measured across CodeRhapsody coding
sessions in 2026: tool results were about **42%** of conversation history, and
tool-call arguments another **30%**. Roughly three-quarters of the tokens, and
they carry almost none of the continuity. The agent's reasoning, its decisions,
its sense of what it is doing — those are cheap and they are the part you
cannot regenerate.

Which yields the rule:

> **Know what you are throwing away.** Purge categories first, summarize last.
> A category purge is lossy in a way you can name and have measured. A summary
> is lossy in a way you will discover later, in production, as a personality
> change.

Hence the levels, weakest first: stub the tool *results* but keep the calls;
remove calls and results entirely but keep visible reasoning; remove prose and
reasoning but **never** the goal stack; and only when compacted records have
themselves piled up, summarize.

**Stubs are synthesized, not stored.** A `RedactResult` stub is computed by the
reducer from the event it supersedes (the tool's name, the size, the path the
output still lives at), which makes it deterministic (so replay is stable),
recoverable (§2.2's greppable log, and Chapter 4's on-disk tool output), and
free of storage that grows. Only `RedactSummary` stores a `Replacement`,
because only there is the new content something an LLM wrote and nobody can
recompute.

**A summary is a fold; the other three levels are filters.** `RedactResult`,
`RedactTool` and `RedactDialogue` rewrite each entry in the span
independently: N entries in, N entries out. `RedactSummary` collapses the span
to a **single** entry carrying the `Replacement`. That distinction is worth
stating because the tidy implementation is the wrong one: four levels, one loop
over the span, one `case` each. Written that way the summary is copied into
every entry it was meant to replace, and compaction *grows* the context it was
called to shrink. (Measured on the reference solution before the fix: a
three-entry span produced three copies of its own summary.) The collapsed entry
takes `Seq = From`, the span's own start, which is already in the event and
therefore deterministic under replay. Its `Actor` is `System`, because a span
can cross Human, Agent and Tool, and a summary of several speakers is not any
of their speech. It is compaction output.

**And compaction is an event.** It goes in the log like everything else, which
is what lets both of this chapter's promises hold at once: the log stays
complete and append-only, replay reproduces the *compacted* context exactly,
and the context stays bounded. A framework that compacts by mutating its
in-memory history has quietly given up on replay, and will not notice until it
needs to debug a session it can no longer reconstruct.

The policy (what thresholds trigger which level, and where the boundaries fall)
is context engineering, and it gets its own chapter. Chapter 2 owes it only a
shape it will not have to break.

### Usage is four numbers, not two — and they are not disjoint the same way

`Usage` looks like bookkeeping and is not. It is the **instrument** that makes
everything in the previous section tunable: you cannot set a token threshold
you cannot measure, and you cannot justify a stable prefix without knowing what
a cache read costs relative to a cache write.

Four categories, at genuinely different prices:

| category | rough price relative to plain input |
|---|---|
| `Input` | 1× |
| `CacheWrite` | **more** than 1× — you pay a premium to create the entry |
| `CacheRead` | **far less** — roughly an order of magnitude cheaper |
| `Output` | several× |

That `CacheRead` row is the whole economic argument for the volatility ordering
in the context-engineering chapter. Put your most-changing content at the front
of the prefix and you convert the cheapest category into the most expensive one,
on every single request, forever. It is the most costly one-line mistake in
agent engineering, and it is invisible without this struct.

**The trap: vendors disagree about whether their own categories overlap.**

Some report cached tokens as a **subset** of the prompt total. Others report
them as **disjoint** additions alongside it. Normalize naively (sum everything
you are given) and you double-count on one vendor and undercount on another,
producing a cost figure that is confidently wrong in opposite directions
depending on which model you are talking to.

This is the purest seam bug in the chapter. Nothing crashes. No test fails. The
number is simply not the number, and you will act on it for months.

> **Canonical form: the four fields are disjoint and sum to the billable
> total.** Whatever a vendor reports, the parser converts it. Where a vendor's
> convention is a subset, subtract; where it is already disjoint, pass through.

*(For the coder: the per-vendor conventions must be verified against current
documentation before print — which is subset, which is disjoint, and whether
either has changed. This is exactly the class of claim that was falsified in
Draft 3. State the date of verification in the text.)*

**Record counts, never money.** Prices change; token counts are history. A
dollar amount in the log is wrong the moment a vendor reprices, and it destroys
your ability to re-cost historical sessions under new rates. Pricing is
configuration and belongs beside the model id; usage is a fact and belongs in
the log. Same distinction as `Provenance` versus format vocabulary: record what
happened, compute what it means.

**A known gap, flagged rather than solved.** At least one vendor bills cache
*storage* by duration, a real cost with no token count attached. A struct of
pure counts cannot express it, and cache lifetime is a concept Chapter 2 does
not have. Noted here so that the later caching chapter adds it deliberately,
rather than discovering that `Usage` was the wrong shape all along.

**Note what is absent: there is nowhere to put system prompt text.** That is
deliberate and structural. The system prompt is an *output*, computed by the
renderer from `Config` (§2.4). If a reader wants to store it, they will have to
add a field — and adding it is the moment to stop and re-read §2.4.

Equally absent: `role`, `content`, `tool_use_id`, `assistant`.

**But note the distinction `Provenance` forces, because it looks like a
violation and is not.** Storing `"anthropic"` or `"claude-opus-5"` is recording
a *fact about where bytes came from*. Storing `role` or `tool_use_id` would be
adopting a vendor's *description of what the bytes are*. The first is history:
it happened, it is not re-derivable, and throwing it away is lossy. The second
is a format decision, and format decisions belong in the renderer.

The test to apply to any field you are tempted to add: **could this have been
different if the same conversation had happened against another vendor?** If
yes, it is provenance and it belongs. If it is just that vendor's word for
something you already model, it has leaked.

### The seam

```go
type Renderer interface {
    Render(*Context, Config) (*http.Request, error)
}

type Parser interface {
    Parse(status int, body []byte) ([]Event, error)
}
```

**Two methods. No vendor types in either signature.** That is the whole
remedy, and its smallness is the point.

`Parse` returns **events**, not a message or a context. There is exactly one
path into the context (append events, run the reducer) so a vendor response
and a human keystroke enter by the same door. Give the parser the power to
mutate context directly and you have quietly created a second reducer that
nobody will remember to keep total.

Contrast the shape that cost 30,000 lines:

```go
// DON'T. This is the mistake in §2.0, in its natural habitat.
type AIClientInterface interface {
    SendMessage(msgs []ClaudeMessage) (*ClaudeResponse, error)
    CountTokens(msgs []ClaudeMessage) (int, error)
    // ...eleven more methods, each shaped by what ClaudeClient
    //    already happened to do
}
```

The tell is visible without knowing the story: **vendor types in the
signature.** `ClaudeMessage` in the interface means the interface *is* the
Claude client, and the second implementation can only be a copy-paste. An
interface extracted from one implementation records that implementation's
accidents as though they were requirements.

### What exists now for a later chapter

Under write-once, readers will reasonably ask why a field exists that nothing
uses yet. Answer honestly, in the text:

| element | used in Ch2 | exists for |
|---|---|---|
| `ToolCallPart` / `ToolResultPart` | rendering supplied logs | Ch3, where tools are executed |
| `BlobPart.Path` | not exercised | Ch4, when tool output arrives by the megabyte |
| `OpaquePart` | thinking signatures | every chapter; never interpreted |
| `TurnState.ToolsPending` | replaying supplied logs | Ch3 |
| `RedactedPart` | fully exercised | Ch7 |
| `Redaction` levels above `RedactResult` | not exercised | context engineering — the compaction gradient |
| `RedactData.From`/`To` as a span | spans of one | context engineering — "every tool call before the last save_memory" |

This table is the write-once discipline made visible, and it is the chapter's
answer to "am I over-engineering?" — no, and here is the receipt.

---

## §2.5 Actors and rooms

Actors: `Human`, `Agent`, `System`, `Tool`. Rooms group a conversation. There
is deliberately **no `To` field**: addressing is a property of the room, not
of the message, and adding `To` invites a routing layer the book does not want.

The `Tool` actor looks like over-modeling until §2.6.

---

## §2.6 The seam — renderers and parsers

> The context is the truth. A **renderer** turns truth into one vendor's
> request. A **parser** turns one vendor's response back into truth.
> Distortion lives in those two places and nowhere else.

The seam is **bidirectional**, and rendering is the easy half. `AIClientInterface`
did not fail because request formatting was hard; it failed because
vendor-shaped thinking hid in the response path, in retries, in errors, in
streaming, in token accounting.

### Exhibit A — one tool result, three authorships

The demonstration the chapter is built around. A single `ToolReturned` event,
`Actor: Tool`, rendered three ways.

**Anthropic**, a `tool_result` block inside a **user** message:

```json
{ "role": "user",
  "content": [ { "type": "tool_result", "tool_use_id": "toolu_…",
                 "content": "ok" } ] }
```

**OpenAI**, its own message with a **tool** role:

```json
{ "role": "tool", "tool_call_id": "call_…", "content": "ok" }
```

**Gemini**, a `functionResponse` part in a **user** turn:

```json
{ "role": "user",
  "parts": [ { "functionResponse": { "name": "…", "response": { … } } } ] }
```

Three vendors cannot agree on who said it. Anthropic says the human did, which
is false, and is the tidiest available lie under a format that demands strict
user/assistant alternation. OpenAI invents a role. Gemini splits the
difference.

**The context is right and all three wire formats are compromises, in different
directions.** That is the entire argument for the seam: the student will watch
one `Actor: Tool` event become three different claims about authorship, and
none of the three is worth storing.

Corollary, and the reason §2.5 models a `Tool` actor at all: **authorship is a
rendering decision.** If your context stores `role: "user"` for a tool result
because that is what Anthropic wanted, you have already lost, and you will
discover it in the copy-paste.

### Exhibit B — the merged message

Anthropic will not accept two consecutive user messages. So a tool result plus
the human's next instruction merge into one message:

```json
{ "role": "user",
  "content": [
    { "type": "tool_result", "tool_use_id": "toolu_…", "content": "ok" },
    { "type": "text", "text": "now check the config instead" }
  ] }
```

Nothing in the context looks like this. Two honest, separate, ordered facts,
fused because one vendor demands alternation. Render the same log for OpenAI
and they stay separate. The merge is not a fact about the conversation; it is a
fact about a wire format, and it belongs in exactly one function.

### Exhibit C — parsing back

Three response shapes normalize to identical context:

| vendor | assistant text at | tool calls at | stop signal |
|---|---|---|---|
| Anthropic | `content[]` blocks | `tool_use` blocks | `stop_reason` |
| OpenAI | `choices[0].message.content` | `.tool_calls[]` | `finish_reason` |
| Gemini | `candidates[0].content.parts[]` | `functionCall` parts | `finishReason` |

The grader's real question: **feed all three responses, get contexts that are
byte-identical apart from `Provenance`.** Everything the model *said* must
normalize; exactly one thing must survive — the record of who said it, and with
which model, on which surface.

That "apart from" is not a loophole, it is the whole distinction of §2.4a: the
content normalizes because content is ours, the provenance persists because it
is history and is not re-derivable. A submission whose three contexts are
*fully* identical has thrown provenance away and will be unable to render a
valid Gemini request later. A submission whose contexts differ anywhere else
has leaked vendor shape past the parser — and leaked vendor shape is precisely
what makes the second implementation a copy-paste.

Also normalized here: token accounting (`usage.input_tokens` /
`prompt_tokens` / `usageMetadata.promptTokenCount`), and errors: an HTTP 429
is an `ErrorOccurred`, not a response.

### Rules the seam has to hold

- **Decline vendor stateful conversation APIs.** Server-side threads (or
  `previous_response_id`-style continuations) trade away the ability to edit
  history. Editing history is a coding agent's core tool: redaction, replay,
  compaction. Own the history or you cannot build the product.
- **Media asymmetry is a LOUD error.** An audio part rendered for a text-only
  model raises; it never silently drops. Fallbacks convert an invariant
  violation into silently-wrong output.
- **Empty is not absent, and the seam must keep them apart.** A vendor may
  legally return `content: ""` — an assistant turn that genuinely produced no
  text. If the parser decides an empty string is not worth recording, the turn
  becomes structurally empty, and the renderer, asked to serialize *nothing*,
  reaches for `null`. We shipped exactly this bug. OpenAI returned

      "message": {"role":"assistant","content":"","refusal":null},
      "finish_reason": "length",
      "usage": {"completion_tokens_details":{"reasoning_tokens":1024}}

  and our very next request on the same wire said

      {"role":"assistant","content":null}

  which that API rejects on a bare assistant message. Note what this is *not*:
  the vendor was consistent, accepting `""` and refusing `null` throughout. The
  round trip lost the distinction, and a value the vendor never sent came back
  to it as an error. The other two renderers were already correct, which is the
  tell — when one of three implementations of a seam is wrong, the seam is
  usually fine and the implementation is lazy.
- **Opaque replay material is carried, never interpreted.** Thinking
  signatures, tool-use ids, cache markers: store them, hand them back to the
  **exact model** that issued them, and never to a different one. Match the
  grain to the model (see `Provenance` in §2.4a).
- **The context never learns a vendor's vocabulary.** If the word `assistant`,
  `toolu_`, or `functionCall` appears in your context types, the seam has
  already leaked.

### The prediction the chapter makes out loud

The second renderer costs real work. **The third should be nearly free.** If it
is not, the seam is wrong — and the reader will discover that in an hour
instead of in 30,000 lines.

This is the chapter's falsifiable claim about its own design, and it is the
only place in the book where the reader can run the experiment themselves. So
the **order is fixed, and it is fixed to make the test honest**:

1. **Anthropic**: the baseline. Everything the student already has.
2. **OpenAI**, a moderate difference: a `tool` role of its own, a flat message
   list, `tool_calls` as an array. Enough divergence to force a real
   abstraction rather than a rename.
3. **Gemini**: the genuinely alien one. `contents` rather than `messages`,
   `parts` rather than blocks, `role: "model"`, `systemInstruction` hoisted
   clean out of the message list, `functionCall`/`functionResponse`.

**The hardest vendor goes last on purpose.** The tempting order puts the most
different one second and the most familiar one third — and then "the third was
nearly free" is true because the third was *easy*, not because the seam was
right. The claim would pass for the wrong reason. Put the alien one last and
the prediction is tested in the direction that can actually falsify it.

Tell the reader to notice what each one costs them — and be exact about the
unit, because the obvious one is wrong. **The bet is about context changes, not
clock time.** If a renderer lands without sending them back into `Context` to
add a field, the seam held for that vendor. If one forces a field in, the seam
was missing something, they have learned it on day one, and the chapter has
done its job by losing its own bet.

**Do not let them measure this in hours.** The third implementation will take
longer than the second no matter how good their seam is, and a student timing
themselves will draw exactly the wrong conclusion from that. Hours are evidence
about the vendor, not about the design. Context diffs are the only honest
instrument, which is convenient, because they are also the only one the grader
can read.

> **Authorial note (2026-09-12).** Bill ruled: leave the criticism in. It is
> now receipted by the deprecation paragraph below, which costs a whole
> edition and is checkable by anyone with a Vertex account. Keep the
> criticism and the receipt adjacent — the characterization earns its place
> because the paragraph under it proves the pattern, not because it is vivid.
> Do not let a later editing pass separate them.

**The Gemini API fails in ways that do not announce themselves:** a request
that returns nothing at all, an error that describes a problem you do not have,
a silence indistinguishable from a bug in your own assembly code.

**The receipt is this book.** As of September 2026, the Gemini surface this
chapter teaches is deprecated. Its replacement, the Interactions API, is not
available on Vertex AI, the access path many corporate readers of this book are
required to take. So the chapter documents the deprecated surface, because that
is the one they can actually reach from where they stand. When Vertex catches up, this book will need a second edition: not because
anything about agent architecture changed, but because a vendor deprecated a
surface before shipping the replacement to its own enterprise platform.

**And there is no event to subscribe to.** The old surface was deprecated
without shipping any way to learn when the new one reaches Vertex AI. So the
migration path is a polling loop with a human in it. I checked a week ago. What
is the correct interval for polling a vendor's roadmap (weekly? monthly?) is
left as an exercise to the reader, and it is the only exercise in this book with
no defensible answer.

Which is the real argument for the seam, made better than any diagram could
make it. You are not building it to add vendors. You are building it because
the one vendor you already use will move the ground under you on a schedule
you do not control, announce it in a changelog, and leave you to find out
whether your abstraction was real.

**This is the second reason the order is fixed, and the more useful one.**
Building against the most honest API first is not a difficulty ramp, it is
establishing a control. When a vendor's failure is ambiguous (and one of them
always is) you need to already know that your context assembly is correct, or
you cannot tell their bug from yours and will spend the afternoon apologizing
to a machine that was wrong. Order your implementations so the
ambiguous failures arrive *after* you have something trustworthy to bisect
against. That habit outlives every vendor named in this chapter.

Built in the fixed order, verified against live APIs on 2026-09-12:

| renderer | context changes forced | vendor file |
|---|---|---|
| Anthropic | defined the core | 271 lines |
| OpenAI | **none** | 235 lines |
| Gemini | **one field** — `ToolCallPart.Opaque` | 258 lines |

The seam held on the harder half. OpenAI's surface disagrees with Anthropic's
about tool-result authorship, id handling *and* usage conventions, and it cost
the context nothing at all. Gemini then forced exactly one field, for the reason
§2.4a gives: `thoughtSignature` arrives bound to a particular call, and
`OpaquePart` had nowhere to put it.

Bent, then, not broken — so the chapter ships that field in §2.4a and tells you
it lost, instead of letting you meet it as a 400 on a Tuesday.

Two points always fit a line. A student can shape the interface around vendor
A, bend vendor B to fit it, and call the result a seam. The third
implementation is what separates an abstraction from a bridge between two
specific things.

---

## §2.7 Versioning and replay

- Replay with **current code**, not with historical code. Log format carries a
  semantic version.
- **Where the version lives, and the asymmetry that governs it.** A version is
  not an event, so it does not get a `Seq`: emit it as a header line,
  `{"log_version":1}`, ahead of the events. Then be **lenient about it**: a log
  with no header is assumed current, and the grader ignores the line entirely,
  so omitting it costs nothing. Reserve strictness for what you must
  *interpret*. That split is the rule worth carrying: **be forgiving about
  metadata you control, unforgiving about anything whose meaning you would have
  to guess.**
- **An unknown event type is a refusal to load, loudly.** Not a skip. Skipping
  an unknown event silently produces a context that is wrong in a way nothing
  downstream can detect.
- Retention is **policy**, not architecture: the log is complete; what you keep
  is a separate decision.

**Four ways non-determinism gets into a renderer**, named because the failure
message only tells you *that* two renders differed: the clock, a randomly
generated id, Go's deliberately randomized map iteration order, and iteration
over a set. The last two are the same bug, and they are why
wire types should be structs with ordered fields rather than `map[string]any` —
the map serializes differently on some future run, on some future machine, and
never on the one where you tested.

---

## §2.8 The exercise

### Commands

| command | behavior |
|---|---|
| `./ch02 chat` | Chapter 1's interactive loop, unchanged in observable behavior |
| `./ch02 render LOG` | play `LOG` → context → render; print the vendor request JSON that *would* be sent to stdout, and **nothing else on stdout**; exit `0`. **Makes no network call.** |
| `./ch02 dump` | write the event log as JSON-lines |

`render` is the centerpiece: it makes replay, redaction, ephemera and the seam
into **byte comparisons**, and it proves context is separable from transport.
If your architecture cannot offer it cheaply, your context is not actually
separate from your transport, which is the finding the exercise exists to
surface.

`render` takes **no flags**. Vendor target, model id and every other request
parameter come from the environment (`LLM_VENDOR=anthropic|openai|gemini`),
exactly as in grader mode. Two renders are compared byte for byte, so the
moment rendering accepts `--model`, byte-identity becomes a property of how you
invoked the command rather than of the log.

**Build them in this order: Anthropic, then OpenAI, then Gemini.** The reason is
in §2.6: the order is what makes the chapter's prediction a real test rather
than a flattering one. Note how long each takes you. That number is the
chapter's actual lesson, and it is yours rather than ours.

### Log serialization

JSON-lines, one event per line, ascending `Seq`. Each line carries at minimum
`seq`, `type`, and the event's own fields; dialogue events carry `actor`. The
format must round-trip: `dump` → `render` must work in a fresh process with no
other state.

**The event-type vocabulary is frozen for this exercise.** Three checks assert
on the *contents* of your dumped log, which is impossible unless we agree on
names. Use the §2.3 set. Spell them as you like: the grader compares type names
lowercased with punctuation stripped, so `ToolCalled`, `tool_called` and
`TOOL-CALLED` are the same event, and it does the same for field names. What it
cannot do is guess that you called it `Halted`.

### No network calls for two of the three vendors

The grader serves **fake endpoints for all three vendors**, so a full seam can
be built and graded with a single API key — or none. Students with one vendor
account are not second-class citizens, and nobody pays three subscriptions to
finish Chapter 2.

### Checks

| check | pts | property |
|---|---|---|
| `session` | 0 | stdio protocol honored; directives acknowledged; request census |
| `ch1parity` | 25 | all seven Chapter 1 checks still pass, unchanged |
| `logdump` | 5 | log round-trips: `dump` → `render` in a fresh process |
| `replay` | 10 | two renders of one log are byte-identical |
| `redaction` | 10 | a `Redacted` event names its target; content absent from later renders |
| `ephemera` | 10 | delivered in exactly one request, and absent from every later one |
| `usage` | 10 | all four token categories normalized from all three vendors into one **disjoint** set — cache reads and writes separated from plain input, summing to the billable total |
| `seam-render` | 15 | one log renders correctly to all three vendor request shapes, and per-call replay material survives a round trip back to the model that issued it |
| `seam-parse` | 15 | three vendor responses produce contexts agreeing on **everything the model said** — actors, text, tool-call names, canonicalized arguments — while `Provenance`, vendor-issued tool-call ids, and model-bound opaque material legitimately differ |

**Sum: 100.**

Notes on the weighting:

- **The parse side still outweighs the render side, 25 to 15**: it is just
  itemized now. `usage` is parsing work: normalizing four token categories
  across three vendors that disagree about whether their own categories
  overlap. Pulling it out of `seam-parse` and naming it separately means a
  student who gets the message shapes right but the accounting wrong is told
  *which* half failed, instead of losing a large undifferentiated block.
  Parsing is where vendor shape hides, and where the author's own seam failed.
- **`seam-render` absorbed the per-call replay property instead of becoming a
  tenth check.** The itemization argument above is about *diagnosis*, and
  distinct failure messages satisfy it; dividing points is for when a student
  can plausibly have one of two skills and not the other. Rendering the right
  request shape and handing a model back its own opaque material are the same
  skill on the same wire format, so the property folds in and the 15 stands.
  What it costs is a second *fixture*, not a second assertion. The property
  needs a log whose provenance matches the render target, and the chapter's
  main exhibit is Anthropic-authored by design: render that one to Gemini and a
  correct implementation withholds the signature, so the assertion passes
  without testing anything.
- **`ch1parity` stays at 25**, honoring the standing guard from Chapter 1's
  review. Below that, a rewrite that silently breaks Chapter 1's contract
  starts to look survivable.
- **`session` is worth zero and can still sink a submission.** Without it, one
  unacknowledged directive fails four checks at once and the student gets four
  mysteries instead of one cause. *(Ratified 2026-09-12 — it stays. It has
  already earned the slot: the `ch1-protocol-broken` mutation fails `ch1parity`
  and `session` together, and `session` is the one that names the cause.)*
- **`ephemera` grades the observable property and takes no position on
  storage** — but state the intended reading in the prose anyway, because the
  wrong one is expensive. An ephemeral part **is recorded in the log** and
  **never enters the dialogue**: delivered in exactly one request, absent from
  every later one. Read instead as "never reaches the log", it breaks §2.4a's
  central promise that `Context = replay(Log)`, since a pending ephemeral would
  then need a second, unlogged path into the context — and §2.6 allows exactly
  one. The mechanism earns a final sentence because it makes §2.3's best rule
  concrete a second time, for free: an ephemeral part arrives as an ordinary
  `MessageReceived` with `Actor: System`, and **the reducer** is what decides it
  is pending rather than dialogue. The capture site does not know, and cannot.

### Two omissions that look alike

Chapter 2 leaves one redaction level ungraded, and it left one struct field
ungraded until the code was built. They look like the same omission. Only one
of them was.

`RedactSummary` is ungraded **on purpose**. §2.4a builds the whole redaction
family and exercises exactly one level of it; the rest is shape ahead of
capability, cashed in by the context-engineering chapter. An ungraded level
there is the plan working. It still had a real bug, the fold described in
§2.4a, which is the distinction to hold onto: a shape can be wrong as well as
untested, and that one had to be found by hand.

`ToolCallPart.Opaque` was a hole. §2.4a lists it as shipped and needed now, and
§2.6 prints it as the entire price of the seam bet: the one field standing
between the student and a 400 the first time a tool call goes back to Gemini.
When the grader was built, deleting its only use scored **100/100**. A field the
book advertises as load-bearing must not be omissible for full marks.

What hid the hole is worth more than the hole. Grep the grader for `opaque` and
you get hits, all of them for the standalone thinking block, which is
thoroughly graded. The distinction the field exists for, a signature bound to
one *call* rather than to the *turn*, was exactly the distinction the tests did
not draw. A field-by-field diff of code against spec reports no drift here and
is wrong.

The question that finds these is not "does the code implement the spec?" It is
**"what would still pass if I deleted this?"** That is mutation testing pointed
at the spec rather than at the code, and it is the first pass to run against
any grader in this book, including the ones already written. A check that
cannot fail is a green dashboard with a schema around it.

### What you are not building

No tool loop — Chapter 3. No mailbox, hints, or interrupts — Chapter 5. No
streaming, no retries, no skills, no sub-agents. **You are building one context
and three ways in and out of it.**

---

## §2.9 Drive it yourself

Ungraded. Do it anyway.

Chapter 1 ended by telling you to chat with the thing you built, because talking
to your own chatbot is a better argument for the architecture than any diagram.
This chapter's payoff is quieter and, once you see it, larger: **the same
conversation, through three different vendors, from one log.**

**Against the fake, free and keyless:**

    go run ./cmd/fakevendor -ch 2 chat
    go run ./cmd/fakevendor -ch 2 -vendor gemini chat
    go run ./cmd/fakevendor -ch 2 -vendor openai chat

The fake will tell you outright that it is scripted and did not read what you
said. Believe it. A fake proves your plumbing, not your prompting.

**Live, against all three:**

    scripts/live.sh 2 anthropic
    scripts/live.sh 2 gemini
    scripts/live.sh 2 openai

That scripted session is doing more work than it looks. It asks the model to
invent a codename, then asks an unrelated question, then asks for the codename
back. The recall proves the entire history is being re-sent and re-rendered on
every request. Run it against all three vendors and watch three different wire
formats produce the same remembered word.

Then read the usage line it prints:

    {"usage":{"input":737,"cache_write":0,"cache_read":0,"output":350}}

Four counters, disjoint by construction. That shape is not the shape any of the
three vendors reports, and §2.6 is the argument for why we record counts and
never money.

For a free-form session, add `chat`:

    scripts/live.sh 2 anthropic chat

Every command in this section was run before it was printed.

**Things worth trying:**

- Record one session, then render it as two different vendors without touching
  the network:

      CH02_LOG=/tmp/s.log go run ./cmd/fakevendor -ch 2 chat
      LLM_VENDOR=anthropic ./ch02 render /tmp/s.log
      LLM_VENDOR=gemini    ./ch02 render /tmp/s.log

  Nine lines of log produced 896 bytes of Anthropic JSON and 851 bytes of
  Gemini JSON on the run that wrote this paragraph. One opens with `system` and
  `messages`, the other with `systemInstruction` and `contents`. Nothing is
  shared but the conversation. That is the chapter, in two commands, and it is
  the exercise `seam-render` grades.

- Ask a question whose answer depends on something you said three turns earlier,
  and watch the request grow as the history is re-sent.

- Record against one vendor and render as another. A conversation that happened
  in Anthropic's format becomes a well-formed Gemini request. Nothing about that
  should work, and it does, because the log is not anybody's wire format.

**Commit and tag the passing state:**

    git commit -am "ch2: one log, three vendors, grader 100"
    git tag ch02-pass

---

## Open questions for Bill (author scaffolding)

1. ~~**Three vendors, or two required plus one as payoff?**~~ **RULED
   (2026-09-12): three.** Anthropic, OpenAI, Gemini — all three graded, in that
   order. Three matches the war story (the disaster was three clients) and is
   the only count that makes the "third is nearly free" prediction testable at
   all. The order is load-bearing and is fixed in §2.6: the alien vendor goes
   **last**, so the prediction is tested where it can actually fail.
   *Retreat position if it proves too heavy when the code is built:* grade two,
   ship Gemini as an ungraded exercise with the prediction attached. Retreat on
   evidence, not in advance.
2. ~~**Which two, if two?**~~ **MOOT** — resolved by the ruling above. Exhibit A
   needs all three to land in any case, since its whole point is that three
   vendors cannot agree on who authored a tool result.
3. ~~**Does `chat` have to work against all three vendors live?**~~ **RULED
   (2026-09-12): no live testing is required to score 100.** The grader's fakes
   are the arbiter: a solution that works against them is accepted. For readers
   who want to run live, a cheap proxy is offered so nobody has to sign up for
   three vendor accounts — bring your own key if you prefer, identical either
   way. But if you want it to work live, you have to test it live; passing
   against a fake is not a claim about production.

   **Say that last part in the prose, because this book has already been caught
   by it.** Draft 3 of this chapter asserted a vendor capability that our own
   fakes happily accepted and that turned out to be false the moment it was run
   against the real API. A fake is a *model* of a vendor, and a model is wrong
   in exactly the places you did not think to model. That is not an argument
   against fakes — they make this exercise affordable and they catch the bugs
   that matter here — it is an argument for knowing what a green grader does
   and does not prove.
4. ~~**Is `ch1parity` at 25 still right?**~~ **RULED (2026-09-12): yes, 25
   stays.** A quarter of the grade for "you did not break what you already
   had" is defensible even with the seam at 35, and it honors the standing
   guard from Chapter 1's review.
5. ~~**Should Chapter 2 state the roadmap?**~~ **RULED (2026-09-12): yes —
   as a contract, not a table of contents.** Drafted in §2.0, "The contract,
   stated once."
6. ~~**Does the goal stack belong in Chapter 2's `Context`?**~~ **RULED
   (2026-09-12): no goal stack.** It arrives with context engineering, which
   is also where the policy that needs it is defined. Deferral is cheap here
   for a structural reason worth remembering: adding a *new* field later is
   additive, while reshaping an existing one is not — which is exactly why
   `RedactData` had to be fixed now and this does not. The `RedactDialogue`
   comment no longer forward-references it.
7. **Where does the context-engineering chapter go? — TBD.** It depends on
   tool results existing (Ch3) and on the system prompt existing (Ch6 skills),
   because one of its central claims is that memory belongs in the message
   history rather than the system prompt. That puts it at Ch7 or later. Its
   *hook* — the `Redacted` event — is established here, so placement is
   genuinely flexible and need not be settled now. Ideas captured in
   `book/chapter-context-engineering-notes.md`.
8. ~~**Is `usage` still worth only 5 points?**~~ **RULED (2026-09-12): raised
   to 10, taking 5 from `seam-parse`.** It was priced when it meant "record two
   numbers." It now means normalizing four categories across three vendors that
   disagree about whether their own categories overlap — a silent bug, wrong in
   opposite directions depending on the vendor. Total seam weight is unchanged
   (the parse side is still 25 against the render side's 15); the hard part is
   simply named now, so a failing student learns *which* half broke.

---

## Appendix — material relocated from Draft 3

`book/chapter-05-actors-parking.md` holds, verbatim, the hint and interrupt
sections and review findings M1, M2, M4, E4, E7. Of particular value when
Chapter 5 is outlined:

- **M1**: a pending hint is not in the dialogue; `RequestSent` is what moves it
  there. The one place a careful student still fails.
- **E4**: render *before* recording `RequestSent`, or the hint is delivered one
  round late and gets misdiagnosed as vendor lag.
- **M9's replacement**: hint responsiveness is an *engine* property, not a
  vendor property — measured at an 11.4s mailbox blackout with a slow tool,
  identical across carriages.
