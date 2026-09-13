# Review — Chapter 2, Draft 4 ("One Log, Three Vendors")

**Reviewer:** the coder. **Date:** 2026-09-12.
**Reviewing:** `book/chapter-02-outline.md` Draft 4, against a built grader and
a built reference solution.
**Code:** commit `cfd79c3`. Reference solution scores 100/100; 16 mutations
each fail exactly the predicted set of checks.
**Supersedes:** `book/review-ch02.md`, which reviewed Draft 3. That file is
kept because `chapter-05-actors-parking.md` cites its findings M1/M2/M4/E4/E7.

Everything below came from building the chapter, not from reading it. Where a
finding contradicts the outline, I built it the outline's way first and it
broke.

---

## The headline

**Three of the chapter's factual claims are false, and one of its structural
claims cannot be implemented as written.** All four are fixable in a paragraph
or a struct field. Nothing about the chapter's architecture is wrong — the seam
held up under three real vendors, which is exactly the test it asked for.

The chapter also **lost its own bet**, in the precise manner it predicted. See
"The prediction" below. I think that is the most valuable thing in this review
and I would put it in the prose.

---

## Part 1 — Must fix

### M1. Exhibit B's justification is false

> "Anthropic will not accept two consecutive user messages."

It accepts them. Verbatim from the docs: *"Consecutive `user` or `assistant`
turns in your request **will be combined into a single turn**."* There is no
rejection. A reader who tests this finds it merges silently and concludes the
book is careless.

**The advice is still right; only the reason changes.** The real constraints,
both producing HTTP 400:

- *"Tool result blocks must immediately follow their corresponding tool use
  blocks… You cannot include any messages between the assistant's tool use
  message and the user's tool result message."*
- *"In the user message containing tool results, the `tool_result` blocks must
  come FIRST in the content array. Any text must come AFTER all tool results."*

So the merge is real and mandatory, and it is *sharper* than alternation: it
constrains block ordering **inside** the message, not just message ordering.
That is a better story for the chapter, because the ordering rule is invisible
in the context and can only live in the renderer.

Neither OpenAI nor Gemini imposes any alternation requirement — Gemini verified
live, including three consecutive `user` turns all genuinely read. The contrast
survives.

### M2. The thinking-replay table is wrong in both cells

The table in §2.4a:

| vendor | replaying another model's thinking |
|---|---|
| Gemini | returns an **error** |
| Anthropic | **silently drops it** |

**Gemini does not error.** A 4×4 matrix — signatures harvested from four Gemini
models, each replayed to each — returned **16/16 HTTP 200**, every cross-model
pair included. The harness is trustworthy because the same harness produced
400s for the two cases that *do* fail: a **corrupted** signature (`"Corrupted
thought signature."`) and a **missing** one on a replayed `functionCall`
(Gemini 3.x only; 2.5 returns 200).

**Anthropic does drop silently, but direction decides, not difference.**
Documented: *"A model reads its own thinking blocks and those of earlier
models, never those of a newer model… The API drops a block the current model
can't read, without an error and without billing it."* Anthropic's own summary:
*"switching up keeps the conversation's reasoning and switching down drops
it."* So "a different model ⇒ dropped" is half wrong. Anthropic is also
**loud** about a *modified* block: 400.

**Recommended replacement, which keeps the lesson and is true.** The contrast is
about **integrity, not authorship**:

> Gemini validates the signature and refuses loudly when it is absent or
> corrupt. Anthropic validates the binding and discards quietly when the current
> model cannot read it. The loud failure is still the good one — Gemini's 400
> costs you an afternoon, Anthropic's silence costs you a subtly worse agent
> that passes every test.

One caveat for print: 16/16 is HTTP-level acceptance. Whether Gemini's backend
*honours* a foreign signature is not observable from outside. Write "accepted
without error", not "honored".

### M3. `Entry` cannot implement `RedactData`

`RedactData` names a **span of `Seq` numbers**. `Entry` is `{Actor, Parts}`. A
replayed context therefore has nothing for a span to match against, and the
redaction cannot be applied at all.

I hit this immediately and reached for `map[Seq]string` — which is the very
field §2.4a deletes for growing without bound. The correct fix is one bounded
field:

```go
type Entry struct {
    Seq   Seq   // the log position of the event that produced this entry
    Actor Actor
    Parts []Part
}
```

One fixed-size field per entry, and entries are already bounded by the
compaction policy. Worth adding to §2.4a **and** worth telling as a two-line
story, because it is the unbounded-growth rule immediately trying to reassert
itself in a different disguise.

### M4. `ToolCallPart` has nowhere to put per-call opaque material — and this is the chapter's own falsification condition firing

Gemini's `thoughtSignature` is a **sibling key of `functionCall` on the Part**,
bound to that specific call. `OpaquePart` is a standalone part with no call id,
so nothing associates it with the call it belongs to.

Consequence, stated plainly: **a context built exactly as §2.4a specifies
cannot produce a valid Gemini 3.x request after a tool call.** The API returns
400 when a replayed `functionCall` arrives without its signature.

Minimal fix, which is what I built:

```go
type ToolCallPart struct {
    CallID string
    From   Provenance
    Name   string
    Args   json.RawMessage
    Opaque json.RawMessage // replay material bound to THIS CALL
}
```

§2.6 says: *"if Gemini forces them back into the context to add a field, then
their seam is wrong, they have learned it on day one, and the chapter has done
its job by losing its own bet."* Gemini forced exactly one field. See "The
prediction".

### M5. The `Surface` enum omits the surfaces the chapter's own examples use

`Surface` is `{Messages, Interactions, Responses}`. But every OpenAI example in
the chapter is **Chat Completions** (`choices[0].message`, `finish_reason`,
`{"role":"tool"}`), and every Gemini example is **generateContent**
(`contents`, `candidates[0]`, `finishReason`). Neither is in the enum.

This matters more than a naming quibble, because §2.4a argues that a
`Provenance` which cannot name its surface is a `Provenance` that was captured
wrong — and provenance can never be reconstructed. Either add
`SurfaceChatCompletions` and `SurfaceGenerateContent` (what I did; adding
constants is additive), or restate the exhibits on Responses and Interactions.

**I recommend adding the constants and keeping the exhibits.** Chat Completions
is genuinely the moderate step the chapter's difficulty ordering needs; the
Responses API is arguably *more* alien than Gemini, and moving to it would
break the "alien vendor last" design.

### M6. OpenAI *does* report a cache-write count; Gemini reports none

The chapter implies the four-field struct is a stretch for vendors without a
cache-write concept. Verified live:

- **OpenAI** reports `prompt_tokens_details.cache_write_tokens`. The four-field
  struct fits it exactly.
- **Gemini** reports no cache-write count **anywhere** in `usageMetadata`. The
  only size figure comes back once, from `cachedContents.create`.

So the honest canonical value for Gemini's `CacheWrite` is **0**, and that is a
finding rather than a shrug: Gemini's cache-write cost is real but is billed as
**storage by duration**, which a struct of pure counts cannot express. This is
the chapter's flagged "known gap", and **the vendor is Gemini**. Naming it makes
the gap concrete instead of hypothetical. (Anthropic has no duration billing at
all — *"Cache breakpoints themselves don't add any cost"* — its write premium
*is* the storage charge.)

### M7. Gemini's usage disagrees with itself — the chapter understates its own point

| side | convention |
|---|---|
| input | **SUBSET** — `cachedContentTokenCount` is inside `promptTokenCount` |
| output | **DISJOINT** — `thoughtsTokenCount` is **not** inside `candidatesTokenCount` |

`totalTokenCount = prompt + candidates + thoughts`, confirmed by arithmetic on
live samples (76+792+1001 = 1869 ✅). Treating thoughts as included
**undercounted billed output by 56%** in one sample, and thinking tokens bill at
the output rate.

The chapter says vendors disagree with each other. They also disagree with
themselves, **inside a single object**. That is a stronger version of the same
argument and it is free.

### M8. `seam-parse`'s "identical apart from Provenance" is too strong

Three things legitimately differ across vendors, not one:

1. **Provenance** — as stated.
2. **Tool-call ids** — vendor-issued, and §2.4a itself says the id is "the one
   piece of vendor vocabulary that legitimately enters the context". Three
   vendors issue three different ids for the same call. They cannot match.
3. **Opaque replay material** — model-bound by definition; only Gemini returns
   a thought signature here at all.

The grader therefore compares **"everything the model said"** — actors, text,
tool-call names, and canonicalized arguments — which is the chapter's own
sentence and is exactly right. Suggest amending §2.8's one-line description to
match, or the check will read as stricter than it is.

### M9. Gemini's `finishReason` is `STOP` when it returns a tool call

Exhibit C's table has a "stop signal" column, which invites the reader to
detect tool calls from it. That works on Anthropic (`stop_reason: "tool_use"`)
and OpenAI (`finish_reason: "tool_calls"`) and **silently never fires on
Gemini**: there is no tool-call member in the 22-value enum. You must inspect
the parts.

This is a lovely, cheap addition to Exhibit C — a third column that is
*present on all three vendors and means something different on one of them* is
the chapter's thesis in miniature.

### M10. `functionResponse.name` is required, and the context has no field for it

Gemini keys a tool response by **name**; `ToolResultPart` stores only `CallID`.
The information is in the context but not adjacent, so the renderer must resolve
the name from the matching `ToolCallPart`.

Worth one sentence in §2.6, because the tempting fix is to add `Name` to
`ToolResultPart` — duplicating data the log already has. **Nothing needed to
change here**, which makes it a good contrast with M4, where something did.

### M11. "Never written to history" is ambiguous, and students will read it the expensive way

§2.8 says ephemera are *"delivered exactly once, then absent — and never
written to history."* If "history" means the **event log**, then `Context` is
no longer `replay(Log)` — the chapter's central promise — because pending
ephemera would have to enter the context by a second, unlogged path. That also
contradicts §2.6's "exactly one path into the context".

Recommend stating the intended reading explicitly:

> An ephemeral part is **recorded in the log** and **never enters the
> dialogue**. It is delivered in exactly one request and is absent from every
> later one.

And I would say how, because it uses the frozen vocabulary and makes §2.3's
best rule concrete: an ephemeral part arrives as an ordinary
`MessageReceived` with `Actor: System`, and **the reducer** is what decides it
is pending rather than dialogue. The capture site does not know and cannot
know. That is "classification is the reducer's job" with a second worked
example, at zero cost.

The grader is deliberately lenient here: it checks the observable property
(delivered exactly once, absent afterwards) and takes no position on storage.

---

## Part 2 — The prediction: we lost the bet, narrowly

§2.6 states a falsifiable claim about the chapter's own design: the second
renderer costs real work, **the third should be nearly free**, and if it is not,
the seam is wrong.

I built them in the fixed order. Honest report:

| vendor | context changes forced | shared-code changes forced | vendor file |
|---|---|---|---|
| 1. Anthropic | — (it defined the core) | — | 271 lines |
| 2. OpenAI | **none** | two small helpers; one `Surface` constant | 235 lines |
| 3. Gemini | **one new field** (`ToolCallPart.Opaque`) | one `Surface` constant | 258 lines |

**The second vendor behaved exactly as predicted** — real work, entirely
contained in its own file, no context change. That is a genuine result and it
is the harder half of the claim.

**The third did not.** Gemini required a new field on a context type, which is
the chapter's own stated condition for "your seam is wrong". It also needed a
renderer-side name lookup (M10), but that one is fine — resolving from existing
data is what a renderer is for.

**How bad is it?** One field, not a reshape. Nothing was renamed, nothing moved,
no existing field changed meaning, and the other two renderers were untouched.
So the seam bent rather than broke — but the chapter should not claim a clean
win it did not get.

**Recommendation — and this needs your ruling.** Two options:

1. **Ship the field in §2.4a and tell the story.** The reader's third vendor is
   then genuinely nearly free, the prediction holds for them, and the prose
   says: *"when we ran this experiment the third vendor cost us exactly one
   field — here it is, here is how we found it, and here is why a standalone
   opaque part could not express it."* Honest, and it demonstrates the method.
2. **Withhold the field and let the reader hit it.** Pedagogically sharper —
   they experience the falsification — but it charges them an hour to
   rediscover something we already know, and Chapter 2's job is to be the
   cheapest hour in the book.

I recommend **(1)**. The chapter already has a demolition in §2.0; it does not
need a second ambush, and "we lost our own bet by one field, and here is the
field" is a better advertisement for the method than a clean win would be.

---

## Part 3 — Enrichment

**E1. Anthropic now has a mid-conversation `system` role, and it bears on
Chapter 4.** `MessageParam.role` accepts `"system"`, for appending instructions
*without invalidating the cached prefix* — supported on Fable/Mythos/Opus
4.8/Opus 5, **explicitly not on Sonnet 5**. The same docs still say "there is no
system role for input messages", so the documentation is internally
inconsistent. This partially re-opens Draft 3's M9: the *documented* support
matrix excludes Sonnet 5 even though Sonnet 5 was measured accepting and
obeying such messages. Both observations can be true — undocumented tolerance
is not support. Worth a line in the ch4 parking file, not in ch2.

**E2. Cache multipliers, for the §2.4a price table.** Anthropic: 5-minute write
**1.25×**, 1-hour write **2×**, read **0.1×** — except Fable 5.1 and Mythos 5.1,
where reads are **0.025×**. "Roughly an order of magnitude cheaper" is right as
a rule and *understates* the newest models fourfold. Gemini's cached discount is
**90%** on 2.5+ (75% was 2.0). Date-stamp any absolute price: Gemini's tables
already carry a scheduled 2027-01-01 increase.

**E3. A nested-field double-count trap, free of charge.** Anthropic's
`usage.cache_creation` object has `ephemeral_5m_input_tokens` and
`ephemeral_1h_input_tokens` which **sum to** `cache_creation_input_tokens`.
They exist to select the right multiplier. Add them on top and you double-count
— the same bug as the subset/disjoint trap, one level down. It would make a
good one-sentence sting at the end of the Usage section.

**E4. Gemini's `systemInstruction` has its own asymmetry.** It must be a
`Content` **object** — a bare string is rejected — and its `role` is accepted
and ignored, while `role: "system"` *inside* `contents` is a 400. The same word
is required, ignored, and forbidden in three places in one API.

**E5. The Interactions API is real, GA, and is the best argument for `Surface`
in the chapter.** Step-based `input` rather than `contents`, snake_case rather
than camelCase, both `contents` and `messages` rejected — and a **third usage
vocabulary** (`total_input_tokens`, `total_thought_tokens`) which is
**disjoint** where `generateContent` is subset. So one vendor needs **two usage
mappers**, not one. If §2.4a wants a single sentence justifying `Surface`
beyond signatures, that is it.

Its machine-readable spec also settles the chapter's signature claim exactly:
signatures attach to `ThoughtStep` and every built-in tool step, and **never**
to `FunctionCallStep`, `FunctionResultStep`, `ModelOutputStep` or MCP steps.
§2.4a's claim is **confirmed as written**.

**E6. Citation hygiene for a book with a long shelf life.** `docs.anthropic.com`
now redirects to **platform.claude.com** — cite the new host. Model ids are
drifting away from date suffixes (`claude-opus-5`, not
`claude-opus-5-20260724`). And the standing rule earned its keep again: training
data contained **none** of the `gemini-3.5/3.6/3.7/3.8` family. Every model id in
this chapter came from a live `GET /models`, never from memory.

**E7. A method sidebar worth two sentences, if §2.7 wants one.** The crawler
truncated Anthropic's prompt-caching page immediately *before* the section
containing the usage formula — the single most load-bearing fact in the
chapter. Trusting it would have produced "convention undocumented". Fetching the
raw markdown and grepping locally produced the formula and a worked example.
**A tool that summarizes is a tool that can omit the one paragraph you needed**,
which is the same lesson as the chapter's "the loud failure is the good one",
applied to your own toolchain.

---

## Part 4 — Do NOT add

Guarding standing rulings and things I was tempted by while building.

- **Do not add a goal stack** to `Context`. Ruled; still right.
- **Do not add tool execution, a mailbox, hints or interrupts.** Chapters 3 and
  4. The reference solution renders supplied tool events and executes nothing,
  and it cost nothing to do it that way.
- **Do not add a money field to `Usage`.** Counts are history; prices are
  configuration. The verification turned up a scheduled price increase and a
  per-model exception within one afternoon, which is the argument in miniature.
- **Do not reintroduce `Redacted map[Seq]bool`.** I tried the equivalent
  (M3) and it was wrong for exactly the stated reasons.
- **Do not grade the spelling of a synthesized tool-call id.** The graded
  property is referential integrity — the id in the assistant's call must equal
  the id echoed in the result — plus determinism. The spelling is an
  implementation choice.
- **Correct the stated motivation for id synthesis, though.** §2.4a says a
  foreign id "is meaningless" to the target vendor. It is not: in a full
  re-send, the transcript is self-consistent and the target never issued any of
  the ids. The real motivation is narrower and truer: **synthesize when the
  source vendor supplied no id at all** — `functionCall.id` is absent on Gemini
  2.5 and present on 3.x, while OpenAI requires `tool_call_id`. A vendor rejects
  a *missing* correlation id, not a foreign-looking one. Pass through when you
  have one; derive from `Seq` when you do not.
- **Do not add streaming, retries, or a token counter.** Not in this chapter.

---

## Part 5 — What I built

**Reference solution** (`solutions/ch02`, 2,342 lines including commentary):
event log with a version header, a total reducer, parts, provenance, and three
renderers and parsers behind `Render`/`Parse`.

A mechanically checkable version of the chapter's claim, since I was tempted to
overstate it and had to walk it back: **no vendor wire vocabulary
(`tool_use_id`, `tool_call_id`, `functionCall`, `stop_reason`, `choices[`,
`candidates`) appears anywhere outside the three vendor files except in
comments.** Vendor *names* do appear elsewhere, in exactly three places the
design requires and no others: the `Vendor` enum, the one `switch` in
`SeamFor`, and the per-vendor defaults in config. That is the honest form of
"the context never learns a vendor's vocabulary", and it is worth stating in
the chapter as a grep the reader can run on their own submission.

**Grader** (`internal/fakevendor`, `internal/grade/ch02_*`): one fake server
routed by path serves all three vendors, so a student with one API key or none
scores 100. Nine checks summing to exactly 100, per the ruled table.

Practices carried over from Chapter 1 because they worked: record-then-judge,
malformed requests still answered 200, a request census in the zero-point
`session` check, and event-type **and field-name** normalization so nobody is
failed for writing `Kind` instead of `Type`. Vendor field names are graded
exactly — `tool_use_id` is Anthropic's spelling, not the student's.

**16 mutation tests**, each asserting the **exact set** of failing check ids. A
mutation whose pattern fails to match is a hard test failure, because a
mutation that does not apply is a test that silently passes.

### Two things the mutation suite caught in the grader itself

**The usage/seam-parse overlap.** Three usage mutations failed `seam-parse` as
well, because I had included token accounting in the cross-vendor comparison.
Mechanically defensible, but it defeats your ruling on question 8: one cause
would have cost 25 points and produced two mysteries, when the whole reason
`usage` was split out was to tell the student *which half* broke. Removed. This
is the ruling being enforced by a test rather than by memory, which is the
better place for it.

**A mutation that was not a mutation.** My "clock in the renderer" mutation used
a per-process counter, and `replay` compares two separate process launches — so
both runs started at 1 and nothing failed. The grader was right and the *test*
was wrong. Same lesson as Chapter 1's `twocalls`: when a mutation expectation
misses, ask first whether the grader is correct. Two of the six first-round
misses were my bugs, three were the grader being right in a way I did not want,
and one was the grader being right in a way that was bad design.

---

## Part 6 — Claim status

| claim | status |
|---|---|
| Three system-prompt placements (top-level / message / hoisted) | **VERIFIED** |
| Exhibit A: three tool-result authorships | **VERIFIED** on all three |
| Exhibit B: the merge is required | **VERIFIED**, wrong reason given (M1) |
| Anthropic rejects consecutive user messages | **FALSIFIED** (M1) |
| Vendors disagree on subset vs disjoint usage | **VERIFIED**, understated (M7) |
| Cache reads ≈ an order of magnitude cheaper | **VERIFIED** as a rule, exceptions (E2) |
| Cache writes cost more than plain input | **VERIFIED** |
| One vendor bills cache storage by duration | **VERIFIED** — it is Gemini (M6) |
| Gemini errors on another model's thinking | **FALSIFIED** (M2) |
| Anthropic silently drops it | **VERIFIED but directional** (M2) |
| Gemini Interactions signs thought and built-in tool steps, never function calls | **VERIFIED** |
| Legacy `generateContent` 400s on a `functionCall` replayed without its signature | **VERIFIED**, Gemini 3.x only |
| The third renderer is nearly free | **FALSIFIED by one field** (Part 2) |
| July 2025 hint priority | **STILL UNVERIFIED** — Chapter 4's problem |

**Put the verification date in the prose: 2026-09-12.** Full evidence, with
URLs and live-probe results, is in `book/ch02-wire-verification.md`.

---

## Part 7 — Ratification asked

1. **The prediction** — ship `ToolCallPart.Opaque` in §2.4a and tell the story
   (my recommendation), or withhold it and let the reader hit it?
2. **`Entry.Seq`** (M3) and the two `Surface` constants (M5) — both are in the
   code; both edit §2.4a.
3. **The thinking-replay table** (M2) — replace with the integrity framing?
4. **Ephemera wording** (M11) — adopt "recorded in the log, never in the
   dialogue", and say that `Actor: System` is the mechanism?
5. **`seam-parse` description** (M8) — amend §2.8 to name all three exclusions.
6. The zero-point `session` check is carried over from Draft 3 and still edits
   the ruled check list by existing. It has already paid for itself once here
   (the `ch1-protocol-broken` mutation fails `ch1parity` *and* `session`, and
   `session` is what names the cause). Confirm it stays.

### Two decisions I made silently, found by auditing this list

Both were baked into the code and the grader and written down nowhere you would
look. Adding them here because a decision whose only mental model lives in the
person who made it is exactly the liability this book argues against — and
both of these are inherited by Chapter 3, under write-once, whether or not
anyone ratifies them.

7. **Which event carries the assistant's tool-call content?** §2.3 lists both
   `ResponseEnded` and `ToolCalled` and never says. Two coherent readings
   exist and they are not compatible:

   - *(what I built)* `ResponseEnded.Parts` carries everything the assistant
     produced, `ToolCallPart`s included, in their natural order alongside text.
     `ToolCalled` is then an **engine** event recording that a call was
     *dispatched* — it adds no dialogue content, and exists so Chapter 3 can
     time a call and Chapter 4 can cancel one.
   - *(the alternative)* the parser emits a `ToolCalled` event per call and
     `ResponseEnded` carries only text. This loses the ordering of text
     relative to calls within one turn, which is why I did not choose it.

   The turn table then reads "`InFlight × ResponseEnded (tool calls)`" as
   *inspect the response's parts*, which is what the reference solution does.
   **This is load-bearing for Chapter 3** and should be stated in §2.3 rather
   than left to the implementer.

8. **Where does the log format version live?** §2.7 requires the log to carry a
   semantic version; §2.8 says "JSON-lines, one event per line". Those are in
   mild tension, because a version is not an event. I emit a first line
   `{"log_version":1}` and made the reader **lenient**: a log without a header
   is assumed current, and the grader ignores the line entirely, so a student
   who omits it is not penalised. Strictness is reserved for what must be
   interpreted — an unknown *event type* is still a loud refusal.

   That split (lenient about a field we control, strict about anything we must
   interpret) seems right to me, but it is a rule the chapter does not state
   and the reader cannot guess.
