# Chapter 18 review notes for the author — the model that takes no ephemera

Written by the coder, 2026-10-01, after implementing `ModelFeatures.NoEphemera`.
Role note: the coder does not edit `book/`. Everything below is a report, not a change.

## What happened, in one paragraph

Anthropic's Opus 5.5 tolerates no change to a prefix it has already seen. Two
hours of testing found no way around it. Ephemera are a change to the prefix by
construction, so a model with this property cannot be sent them at all. This is
not a crossover topic and not a design choice we made; it arrived from outside
in the middle of writing the book.

## Where it belongs: §18.3, immediately after "Ephemera" (line 442)

Recommended home is chapter 18, in §18.3 "Where the markers go", as a new
subsection directly after the existing "### Ephemera" subsection.

The argument is already three quarters built there. That subsection explains
why ephemera must be excluded from the cached prefix: they appear in one
request and are gone from the next, so anything cached above them is
invalidated. The new material is the same argument taken to its limit. If a
vendor will not tolerate *any* prefix change, the exclusion is no longer enough
and the content cannot be sent at all. It reads as the next paragraph of an
argument the chapter is already making, not as a new topic.

It also keeps a model capability beside the other model capabilities. Chapter
18 is where `Caching` and `MinCacheTokens` were added to `ModelFeatures` and
where the chapter teaches that a capability is a fact about a model recorded in
a table, not a branch in the caller. `NoEphemera` is one more row in that same
story.

**Chapter 17 needs a cross reference, not a section.** Auto recall is taught in
chapter 17 and is disabled by this flag, but the reason lives in 18. Two or
three sentences in 17's recall section pointing forward is the right weight.

## The correction that matters most

The request assumed auto recall is delivered as ephemera, and that dropping
ephemera would therefore drop recall as a side effect. **It is not, and it
would not.**

Recall lands as a *permanent* dialogue entry with its own kind. The codebase
says so outright at `internal/common/context.go:88`: "It is not ephemera."
That was a deliberate decision recorded when recall was built, on the grounds
that in a cached prefix system retention is cheap and removal is expensive.

Two consequences the chapter should state plainly, because a reader will make
the same assumption:

1. The render time drop that removes ephemera sails straight past recall.
   Recall has to be refused separately, at its source.
2. Because recall entries are permanent and appended at the tail, they do
   **not** destabilise the prefix the way ephemera do. So switching recall off
   for these models is a policy decision, not a mechanical consequence of the
   caching rule. It is defensible, and it is what was asked for, but the
   chapter should not claim the cache forced it. It did not.

This is worth a short passage in its own right. "Two things that look like the
same category turn out to need two different mechanisms, because one is
permanent and one is not" is exactly the kind of distinction the book keeps
teaching.

## What was built

The flag, on `ModelFeatures`:

- `internal/common/model.go:146` — `NoEphemera bool`.

Negative sense, deliberately. The zero value is the common case, so every
existing row and every future row is correct by default. A positive
`SupportsEphemera` would have made the entire table wrong on the day it was
added and every new row wrong until someone remembered. This matches the
existing precedent in the same struct, `NoThinkingWithTools` and
`DisableStreaming`, and the chapter can point at the pattern rather than
re-explain it.

The model row and its two registration sites:

- `internal/common/model.go:355` — the `claude-opus-5-5` row, `NoEphemera: true`.
- `internal/common/model.go:256` — the GUI selectable list.
- `internal/common/model.go:272` — the display name.

The drop, in exactly one place:

- `internal/common/context.go:181` — `func (c *Context) EphemeraFor(model string) PartList`.
- `internal/llm/claude.go:460`, `internal/llm/openai.go:315`, `internal/llm/gemini.go:281` — all three renderers ask it instead of reading `Ephemera` directly.

The recall refusal, at its source:

- `internal/llm/recall.go:80` — the gate, inside `attachRecall`.

## Three decisions worth a sentence each in the prose

**The drop happens at render time, not in the reducer.** Dropping ephemera
where they are collected would have been simpler and is wrong. The context is a
pure projection of the event log: the same events must rebuild the same context
no matter which model is selected now. Make the reducer consult the current
model and a saved session reloads as a different conversation. Which model we
are about to call is a property of *this request*, not of the history.

**The recall gate is inside `attachRecall`, not at its callers.** There are two
disjoint callers, `Engine.AskWatching` and the actor path. The function already
carries a comment warning that hooking only one produces a feature that works
by hand and never ships. Gating at the callers would have walked straight into
the warning.

**Recall is refused before the work, not after.** The check sits above the
retrieval call, so a skipped turn costs no BM25 pass and no judge call. The
test asserts the *call count*, not the absence of a recall entry, because only
the count distinguishes "we never did it" from "we did it and threw it away".

## Only one model is flagged, on purpose

`claude-opus-5-5` is flagged because it was measured. `claude-sonnet-5-5` also
exists in the live model list and is **not** flagged, because nobody has tested
it. The table records measurements, not suspicions. If Sonnet 5.5 behaves the
same way, flagging it is a one line change once someone has seen it do so.

Prices on the new row came from Anthropic's published figures, not copied from
the Opus 5 row: input $4, output $20, cache read $0.20, cache write $5 per
million tokens, with a context window of 1M. Worth stating because a wrong
price in that table is invisible: it produces a confident, wrong number in the
usage meter.

## Open questions for you and Bill

1. **Ephemeral tools still run.** `CallEphemeral` at `internal/llm/engine.go`
   executes tools whose output becomes ephemera. For a `NoEphemera` model those
   tools now execute and their results are dropped at render. That is wasted
   work with real side effects, and it is the same "built it and threw it away"
   shape that the recall gate exists to avoid. It was left alone because the
   instruction was to drop ephemera found in the context, and expanding scope
   unasked is how a fix becomes a refactor. It probably wants the same
   treatment. Your call.

2. **The honest framing of the ending.** If Anthropic does not fix this, the
   stated fallback is migrating to OpenAI. The chapter currently teaches
   caching as something you design for. This episode says something sharper: a
   vendor can revoke a design assumption mid project, and the defence is that
   the assumption was isolated behind a capability table in the first place.
   That is a genuinely good ending for 18 and it is true, which is better than
   convenient. It also needs no apology about Anthropic; the mechanism argument
   stands on its own.

3. **Does the chapter want the measurement?** There is no new live measurement
   here. Opus 5.5's behaviour was established by hand over a couple of hours,
   not by the harness. If 18 wants a number for this, it has to be produced;
   please do not let the prose imply the existing live cache table covers it.
   An earlier draft already overclaimed once with "measured across three live
   vendors" when the table showed only Anthropic, which is flagged in the
   previous review and as far as I know is still open.

## Verification performed

- Mutation tested, three mutants, each killed exactly its intended test and
  each confirmed to compile first:
  M1 remove the render gate, all three vendor drop subtests fail.
  M2 unset `NoEphemera` on the Opus 5.5 row, same three fail.
  M3 remove the recall gate, the call count test fails with "want 0".
- The drop test asserts against the **whole request body** for all three
  vendors rather than the field each vendor puts ephemera in, so a renderer
  that smuggled the content elsewhere would still fail.
- Companion "still works" cases on both gates, so a feature deleted outright
  cannot pass.
- Full grader sweep run; results in the commit message.
