# Chapter 18 review notes for the author — the model that takes no ephemera

Written by the coder, 2026-10-01, after implementing `ModelFeatures.NoEphemera`.
Role note: the coder does not edit `book/`. Everything below is a report, not a change.

## What happened, in one paragraph

Anthropic's Opus 5.5 tolerates no change to a prefix it has already seen. Two
hours of testing found no way around it. Ephemera are a change to the prefix by
construction, so a model with this property cannot be sent them at all. This is
not a design choice of ours and not a crossover topic; it arrived from outside,
in the middle of writing the book.

## Where it belongs: §18.3, immediately after "Ephemera" (line 442)

Recommended home is chapter 18, in §18.3 "Where the markers go", as a new
subsection directly after the existing "### Ephemera" subsection.

The argument is already three quarters built there. That subsection explains
why ephemera must be excluded from the cached prefix: they appear in one
request and are gone from the next, so anything cached above them is
invalidated. The new material is the same argument taken to its limit. If a
vendor will not tolerate *any* prefix change, exclusion is no longer enough and
the content cannot be sent at all. It reads as the next paragraph of an
argument the chapter is already making.

It also keeps a model capability beside the other model capabilities. Chapter
18 is where `Caching` and `MinCacheTokens` joined `ModelFeatures`, and where
the chapter teaches that a capability is a fact about a model recorded in a
table, not a branch in the caller. `NoEphemera` is one more row in that story.

## The best material here is the thing that is NOT gated

This started as "a model that cannot take ephemera must also lose auto recall,
because recall is a kind of ephemera." That premise is false in this codebase,
and chasing it down produced a sharper lesson than the original one. I
recommend building the subsection around it.

**Auto recall is not gated, and keeps running on Opus 5.5.**

The criterion that decides it is a single question: *does the content survive
into the next request?*

- **Ephemera do not.** They appear once and are gone, so the next request
  differs from the last in the middle of the prefix. A model that tolerates no
  such change cannot be sent them at all.
- **A recall entry does.** It is appended to the dialogue and stays
  (`internal/llm/context_ops.go:178`), it is rendered inline in conversation
  order by all three vendors (`claude.go:351`, `openai.go:223`,
  `gemini.go:191`), and it is removed only at a checkpoint
  (`context_ops.go:429`). So it only ever *extends* the prefix.

The removal rule is the detail that makes it airtight, and the codebase already
argues it: recall is deleted at a checkpoint because a checkpoint already
rewrites the prefix and is therefore already paying for the cache miss.
Deleting it anywhere else would buy a second invalidation for nothing.

So the three mechanisms line up: append only between checkpoints, rendered in
place, deleted only when the cache is being thrown away regardless. Nothing
recall does can disturb a prefix Opus 5.5 has already seen.

**The point worth making in prose:** this is a property of *this design*, not of
auto recall as an idea. The identical feature built as ephemera, retrieved
fresh each turn and gone by the next round trip, would have to be switched off
for such a model. A design decision taken earlier for an unrelated reason
(retention is cheap in a cached prefix system, removal is dear) turns out to
decide whether the feature survives a vendor restriction nobody had heard of
at the time. That is a real and uncommon lesson: the payoff for putting a
thing in the conversation rather than beside it arrives years later and from a
direction you did not predict.

Gating recall would have cost a feature for no caching benefit whatsoever. The
tempting move was the wrong one, and only looking at where the bytes actually
go showed it.

A note on sourcing, in case you want to use it: the original instinct that
recall is ephemeral came from a *different* agent's design, where auto recall
genuinely is ephemeral and genuinely is gone by the next round trip. Two
systems, the same feature, opposite answers to the Opus 5.5 question. If you
want the contrast in the chapter it is real, though it may be cleaner to
describe it as "the same feature built the other way" than to name another
codebase.

## What was built

The flag, on `ModelFeatures`:

- `internal/common/model.go:146` — `NoEphemera bool`.

Negative sense, deliberately. The zero value is the common case, so every
existing and future row is correct by default. A positive `SupportsEphemera`
would have made the entire table wrong the day it was added, and every new row
wrong until someone remembered. Matches the existing precedent in the same
struct, `NoThinkingWithTools` and `DisableStreaming`.

The model row and its two registration sites:

- `internal/common/model.go:355` — the `claude-opus-5-5` row, `NoEphemera: true`.
- `internal/common/model.go:256` — the GUI selectable list.
- `internal/common/model.go:272` — the display name.

The drop, in exactly one place:

- `internal/common/context.go:181` — `func (c *Context) EphemeraFor(model string) PartList`.
- `internal/llm/claude.go:460`, `internal/llm/openai.go:315`, `internal/llm/gemini.go:281` — all three renderers ask it instead of reading `Ephemera` directly.

And in recall, a comment where a gate briefly was:

- `internal/llm/recall.go:71` — why recall is deliberately not gated.

## Two decisions worth a sentence each in the prose

**The drop happens at render time, not in the reducer.** Dropping ephemera
where they are collected would have been simpler and is wrong. The context is a
pure projection of the event log: the same events must rebuild the same context
no matter which model is selected now. Make the reducer consult the current
model and a saved session reloads as a different conversation. Which model we
are about to call is a property of *this request*, not of the history.

**One decision point, not one per vendor.** Each vendor renders ephemera in its
own shape, so three copies of the check would be three chances to drift, and
the obvious way to add a fourth vendor is to copy one of the three. The test
renders all three seams against the same model for exactly this reason.

## Only one model is flagged, on purpose

`claude-opus-5-5` is flagged because it was measured. `claude-sonnet-5-5` also
exists in the live model list and is **not** flagged, because nobody has tested
it. The table records measurements, not suspicions. If Sonnet 5.5 behaves the
same way, flagging it is a one line change once someone has watched it do so.

Prices on the new row came from Anthropic's published figures rather than
copied from the Opus 5 row: input $4, output $20, cache read $0.20, cache write
$5 per million tokens, 1M context window. Worth care because a wrong price
there is invisible: it produces a confident, wrong number in the usage meter.

## Open questions for you and Bill

1. **Ephemeral tools still run.** `CallEphemeral` in `internal/llm/engine.go`
   executes tools whose output becomes ephemera. For a `NoEphemera` model those
   tools now execute and their results are dropped at render: wasted work with
   real side effects. It was left alone because the instruction was to drop
   ephemera found in the context, and expanding scope unasked is how a fix
   becomes a refactor. It probably wants the same treatment. Bill's call.

2. **The honest framing of the ending.** If Anthropic does not fix this, the
   stated fallback is migrating to OpenAI. This episode says something sharper
   than "design for caching": a vendor can revoke a design assumption mid
   project, and the defence is that the assumption was isolated behind a
   capability table, and that the features built on durable foundations
   survived while the one built on ephemera would not have. That is a good
   ending for 18 and it has the advantage of being true.

3. **No new measurement here.** Opus 5.5's behaviour was established by hand
   over a couple of hours, not by the harness. If 18 wants a number for it, one
   has to be produced. Please do not let the prose imply the existing live
   cache table covers it. An earlier draft already overclaimed once with
   "measured across three live vendors" when the table showed only Anthropic,
   flagged in the previous review and as far as I know still open.

## Verification performed

- Mutation tested, four mutants, each killed exactly its intended test and each
  confirmed to compile first:
  M1 remove the render gate, all three vendor drop subtests fail.
  M2 unset `NoEphemera` on the Opus 5.5 row, the same three fail.
  M3 gate recall on `NoEphemera`, the recall test fails with "want 1".
  M4 the earlier version of M3 in reverse, confirming the test guards the
  decision in both directions rather than merely describing it.
- The drop test asserts against the **whole request body** for all three
  vendors rather than the field each vendor puts ephemera in, so a renderer
  that smuggled the content elsewhere would still fail.
- Companion "still works" cases on both the ephemera and recall paths, so a
  feature deleted outright cannot pass.
- Full grader sweep, byte identical to the sweep taken before this change.

## Author ruling, 2026-10-01 evening: naming the vendor is allowed here

Bill's decision, recorded verbatim in substance: when a company he does not
work for treats its users this way, forcing them onto other vendors, he is
willing to call it out by name. This overrides the `voice.md` rule of "name the
API or model, never the company" **for this material specifically**. It is a
deliberate exception, not drift, and `voice.md` should be amended to say so
rather than leaving the next lint pass to "fix" it.

His argument, which is the part worth putting in the book:

- **It is not a security measure.** Defeating the existing attacks only
  required checking which model generated a signature before accepting it.
  The capability to fix it narrowly existed.
- **It is a commercial measure.** The effect, and on this reading the purpose,
  is to shut down API users, push them onto first party products, and make
  agents written by third parties unusable. That is where the money is.

Two notes from the coder, for the morning rather than tonight:

1. **The receipt is what makes this publishable.** The signature checking point
   is a factual claim about what would have sufficed technically. Stated with
   it, the passage is an argument from evidence. Stated without it, it is an
   accusation of motive. Before print I would like to verify that claim
   carefully enough to stand behind it, the same standard applied to every
   measured number in 18.

2. **Motive versus effect.** "This is where the money is" is an inference about
   intent. The effect is documentable: a January policy change, a 5x cost
   difference between subscription and API, third party agents locked out. The
   effect alone carries the chapter, and is not arguable. Worth deciding
   deliberately how far past effect the prose goes.

## State at end of 2026-10-01, for the morning

Three commits on this thread, all UNPUSHED, Bill pushes:

- `47fb4de` the endpoint must move with the model (the prerequisite bug, found
  the day before the migration was decided)
- `6bc8f8b` `NoEphemera`, drop ephemera for Opus 5.5
- `f259a49` keep auto recall on such models, with the reasoning left in place
  where the gate briefly was

Full grader sweep run after each, byte identical throughout, 21 targets at
100/100. The ch8 to ch12 failures at `./agent` are pre-existing and were proved
so at a baseline worktree, not assumed. They are still unresolved and are worth
raising separately.

Next chapter proposed by Bill: switching from Anthropic to OpenAI, forced
mid-book. Open and undecided: whether ch18 now ends on the restriction and
hands off rather than resolving; and the OAuth mechanics, which I want to
verify before designing anything, because a ChatGPT Pro token expires and the
credential type introduced this morning, `Endpoint{BaseURL, APIKey string}`,
assumes a static string that never does.
