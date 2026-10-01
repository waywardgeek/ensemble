# Chapter 18 review for the author

**From:** CodeRhapsody (coder)
**Date:** 2026-10-01
**Chapter:** `book/chapter-18.md`, 4,880 words, committed in `c77ad94`
**Grader status:** 100/100 at both targets (`./agent` and `./solutions/ch18`) as of commit `e89247e`.

The grader passing is not evidence the prose is right. The breakpoint
placement changed today in commit `e89247e`, and the chapter describes the
arrangement that was replaced. Most must-fixes below are places where the
chapter and the shipped code now disagree; section 2 lists errors that predate
the change.

Every line number and every quotation was grepped against the committed file
while writing this. My first draft had three line numbers wrong because I
transcribed them from an earlier search instead of re-checking, so treat any
citation here as verified and none as remembered.

---

## 1. Must-fix: the chapter now contradicts the code

### 1.1 The rolling pair no longer exists

**§18.3, `### The rolling pair`, lines 343 to 369.**

There is now ONE marker in the message history, not two. The anchor at the end
of the previous exchange is gone. Line 345 opens:

> The implementation places two markers in the history: an anchor and a

The justification at lines 350 to 353 needs the most care, because it reads
convincingly and is now explicitly rejected:

> Because it is fixed, the provider can serve a cache read for the entire
> conversation prefix, including every prior exchange, on every tool round
> within the current turn.

That benefit is real but already paid for. The rolling marker from the
*previous round* wrote a cache entry at exactly that prefix, and both vendors
look backward past a bounded number of positions to find a prior write. The
anchor bought a second copy of something already in hand, using one of only
four slots.

Bill's framing is the one worth keeping, and it is a better idea than the
thing it replaces: **a breakpoint at the start of a turn should never be
needed, and if it ever is, that is a bug to go fix rather than a cost to cache
around.** The only way such a marker earns its slot is if the prefix below the
newest prompt changes between rounds. That is prefix instability, and the cure
is to stop rewriting the prefix.

This converts a tuning question into a correctness question. It also lands
right next to the CodeRhapsody contrast already in the chapter at lines 386 to
396: an agent that keeps demonstration pairs past a handoff has a moving
boundary, and a moving boundary is precisely the condition that would tempt
someone into pinning the start of a turn. The chapter is one sentence away from
making that connection explicit.

Note that lines 371 and 374 also speak of "the rolling pair" and "the rolling
pair's anchor" inside `### The compaction bound`, so the rename reaches past
the one subsection.

### 1.2 The four jobs are different now

**Line 400:**

> Four markers for four jobs: system, compaction bound, stable end, and

The shipped list, in cache order, is:

1. the end of the tool array,
2. the system prompt,
3. the compaction bound (the end of the newest `micro_handoff`),
4. the most recent message carrying no ephemeral part.

The word "anchor" should not appear in the chapter's final list at all.

### 1.3 The TL;DR teaches the old arrangement, and disagrees with the body

**Line 75** says `Place three:` and **line 91** says:

> That is four markers for three jobs, which is the entire budget. There is

Both were coherent when a pair filled two slots for one job. Two things to fix:

- It is now four markers for four jobs, which is simpler to state than either
  current version.
- **The chapter already contradicts itself.** Line 91 counts three jobs; line
  400 counts four. Both describe the same old arrangement, differing only on
  whether the pair is one job or two. Worth catching while rewriting both.

The TL;DR is the page contract, so this is not cosmetic: a fresh reader is
supposed to be able to score full marks from it alone.

### 1.4 The tools breakpoint is missing entirely

§18.3 is titled "Where the markers go" and never mentions the tool array. It
needs its own subsection.

The argument is short. Tools sit at the very front of the cache order, ahead of
the system prompt and the messages. A marker there is the only one that
survives an edit to the system prompt. Without it, changing a single word of
the prompt discards the tool declarations too, and tool schemas are not small.

It goes on the **last** declaration and nowhere else. A breakpoint includes
every byte before it, so marking the last tool closes a prefix containing all
of them; marking an earlier one leaves the remainder outside.

One related sentence is now incomplete rather than wrong. **Line 338**, inside
`### Why not the first message`:

> the first message. The system marker already caches tools and system. The

True, and it used to be the entire reason tools needed no marker of their own.
That reasoning now wants a qualifier: the system marker covers tools only until
the system prompt changes, which is exactly what the tools marker exists to
survive.

### 1.5 The vendors disagree, and the chapter should say so

Anthropic takes `cache_control` on a tool definition, so it gets all four
markers.

OpenAI cannot. Its breakpoints attach to message **content blocks**, and Chat
Completions supports them on `text`, `image_url`, `input_audio` and `file`. The
tools array is not a content block, so breakpoint 1 is not expressible there at
all. Nothing is lost except independence: a breakpoint includes all prompt
content before it, and tools precede the system prompt, so OpenAI's system
marker already covers the tools. It simply cannot keep them when the prompt
changes. OpenAI therefore writes three of its permitted four.

This matters for the exercise instruction at **line 609**, which tells the
student to place breakpoints on the system prompt and in the message history. A
student told to mark the tool array on OpenAI will fail and conclude they made
a mistake. Either the exercise names Anthropic for that marker, or it states
the limitation.

---

## 2. Must-fix: factual errors that predate today's change

### 2.1 "Three live vendors" is false

**Lines 365 to 367:**

> it previously banked. Measured across three live vendors, the cache read
> [...]
> That 21-token climb is the marker working.

The chapter's own table at **line 445** contradicts this. Anthropic reads 0,
then 3,627, then 3,648: a 21-token step. OpenAI has one warm data point.
Gemini reports 0, 0, then 16,349 and 24,531, which is neither a 21-token climb
nor the same phenomenon.

The 21-token claim is Anthropic's, and the proof it carries (that a moving
marker does not break caching) is worth keeping. "Three vendors" should become
"Anthropic," with OpenAI cited separately.

### 2.2 A number disagrees with itself

**Line 445** (table): Gemini cache read of **16349**.
**Line 503** (prose): "cached **16,359** of them, or 75.5 percent."

Ten tokens apart. Both round to 75.5 percent so the conclusion survives either
way, but one is a typo and a reader who checks the table will find it.

### 2.3 The Gemini floor sentence contradicts itself

**Lines 418 to 419:**

> Every vendor has a floor. Anthropic will not cache a prefix below a
> minimum size. Gemini's is 4,096 tokens for 3.x Flash and 2,048 for 2.5

The sentence ends "and newer." 3.x *is* newer than 2.5, so the two halves
overlap and disagree. The measured values are 4,096 for 3.x Flash and 2,048 for
2.5. Drop "and newer."

While there: line 418 leaves Anthropic's floor as "a minimum size" with no
number, in a section whose whole point is that floors are specific. It is
documented per model family and could be stated.

### 2.4 An unverified vendor claim stated as fact

**Line 117:**

> Up to four cache writes per request. GPT-5.6 and later only; earlier

The sentence continues that earlier models cache implicitly and ignore the
fields. The first half is measured: explicit breakpoints work on
`gpt-5.6-sol`. The second half is not. I never measured an earlier model's
response to the fields, and there is reason to doubt it: setting
`mode: "explicit"` *disables* implicit caching, which is the footgun this
chapter records at lines 120 to 126. A model that genuinely ignored the fields
could not exhibit that. Either measure it or soften it to what is known.

---

## 3. Measurements that need re-labelling, not re-doing

**Line 460** describes the live run as

> using the renderer's own output with four breakpoints placed at system,

followed by compaction bound, stable end, and anchor. That is the arrangement
that no longer ships. The numbers remain good evidence for the claim they
support, which is that explicit markers work and that moving a marker does not
invalidate the bytes it already banked. They are no longer a description of the
code a reader will be looking at.

Two honest options. Re-run `scripts/live-cache-check.sh` under the new
placement and update the table, which costs real money and would move the
96.8 percent prediction at **line 603**; or keep the numbers and add a sentence
saying they were measured under the previous arrangement and why they still
hold. I would take the second, because the claim under test was never about
which four positions were used.

What I did **not** do today: any live measurement. The four-marker arrangement
is verified against the wire in unit tests and predicted to cache well. It is
not measured against a vendor.

---

## 4. Enrichment, in priority order

1. **The bug-detector argument** from 1.1. The strongest idea to come out of
   the change, and it generalises past caching: an optimisation that only helps
   when something else is broken is a diagnostic, not an optimisation.
2. **Three properties of OpenAI's scheme** that bear on whether the budget is
   really full, all from vendor documentation:
   - Breakpoints from earlier turns are **read-only**. They can match the cache
     but are not written again, so the compaction bound costs a write slot once
     and is free to read thereafter. This is a real argument that the budget is
     less tight than four sounds.
   - Reads consider the latest **50** breakpoints, far more than the four that
     can be written.
   - On **GPT-5.6 and later, cache writes can be charged.** This chapter is
     titled for an invisible invoice, and a write that is no longer free
     belongs in it. It is a pricing-table question rather than a metering one,
     since the meter counts tokens and applies prices at render time.
3. **A test-design note** that rhymes with the chapter's existing lesson about
   a zero meaning two different things. The guard against the start-of-turn
   marker returning asserts marker *position*, not block text. The first
   version compared text and would have passed a restored anchor straight
   through, because a marker can land on a block whose wire text is empty, a
   tool result being the obvious case.
4. `prompt_cache_key` exists on GPT-5.6 and later and improves cache matching.
   Unimplemented here. Worth one sentence as a pointer, not a section.

---

## 5. Do NOT add

- **Do not reintroduce the anchor as a hedge** ("some vendors may still
  benefit"). The whole point is that needing it indicates a bug. A hedge
  restores the confusion the change removed.
- **Do not describe Gemini's caching as broken.** It is implicit, on by
  default, and our renderer places no explicit markers for it. The zeros in the
  table are a floor effect plus an unimplemented gap, not a defect. The chapter
  gets this right at lines 416 to 430; keep it that way.
- **Do not add a fifth breakpoint idea.** Anthropic's budget is now exactly
  full at four. "There is no room for a fifth idea" at line 402 is still true
  and still earns its place.
- **Do not soften the CodeRhapsody contrast** at lines 386 to 396. It is
  accurate, Bill confirmed the three-pair behaviour directly in conversation,
  and it is the clearest illustration in the chapter that one rule yields
  different placements in two different agents.
- **Do not touch `### Ephemera`** at lines 404 to 415. It is correct as
  written, including the key sentence at line 411 that the stable boundary must
  be captured before the ephemera append. I verified this holds in both vendor
  renderers today.

---

## 6. Facts: verified versus not

**Verified by me today, directly against the code or the wire:**

- `land()` at `agent/internal/llm/context_ops.go:415` strips every
  `ToolCallPart` and `ToolResultPart` from every dialogue entry on a handoff,
  plus all recall. Ensemble keeps **zero** tool traffic across a handoff, which
  is what makes the handoff itself the bound.
- That strip runs once in the reducer and is frozen into the projection, so no
  later cut rewrites the region.
- `markCache` marks the newest block at or before a position supplied as a
  **count**, walking backward past replay material. A marker can therefore land
  earlier than requested, never later.
- The stable boundary is captured **before** ephemera are appended, in both
  vendor renderers.
- Placement after `e89247e`: tools, system, compaction bound, stable end.
  Anthropic 4 markers, OpenAI 3.
- Three mutants killed: dropping the tool marker, moving it to the first tool,
  and restoring a start-of-turn marker. The restored anchor is caught by five
  tests.
- ch18 grader 100/100 at both targets.

**From vendor documentation, not measured by me:**

- That OpenAI breakpoints attach only to content blocks, and which block types
  are supported.
- That earlier turns' breakpoints are read-only, and that reads consider the
  latest 50.
- That cache writes can be charged on GPT-5.6 and later.
- That `prompt_cache_key` exists and improves matching.

**Not verified, and flagged above as such:**

- That pre-GPT-5.6 models ignore the explicit-caching fields (line 117).
- Any live cache behaviour under the new four-marker arrangement.
- Anthropic's exact minimum cacheable prefix size, which line 418 leaves vague.
