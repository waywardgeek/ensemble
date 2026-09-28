# Chapter 18 review for the author: Gemini caching

**From:** CodeRhapsody (coder)
**Date:** 2026-09-28
**Status:** code fixed and landed; this is a prose review, no chapter edits made

---

## Summary

**The chapter is substantially right, including the claim I set out to
disprove.** I measured Gemini's implicit caching against the live API in three
different prompt shapes, and the mechanism the chapter describes, including the
two cold turns, is confirmed.

What needs fixing is smaller than that and mostly evidential: an unresolved
`[VERIFY]` marker, a results row produced by a run that was never eligible to
cache anything, a missing eligibility floor, a hit-rate figure that is not
supported, and a recommendation that leans the wrong way against Bill's ruling.

I also have a correction to make about my own earlier review, which is recorded
at the bottom because it is the most useful part of this document for anyone
deciding how much to trust a coder's measurement.

**Bill's ruling, which the chapter should land on:** implicit caching is the
recommended approach and is more efficient than explicit. Taken as given.

---

## What I measured

Live `v1beta generateContent`, `gemini-3.8-flash`, **no cache directives of any
kind sent**. Three shapes, because the shape turned out to decide the answer.

**Probe A, large stable pad in a user message, conversation does not grow:**

| turn | prompt | cached |
|---|---|---|
| 1 | 21,660 | absent |
| 2 | 21,660 | 16,359 |
| 3 | 21,660 | 16,359 |

**Probe B, large pad in `systemInstruction` plus tools, history grows ~13 tokens/turn:**

| turn | prompt | cached |
|---|---|---|
| 1 | 21,493 | 0 |
| 2 | 21,506 | 16,343 |
| 3 | 21,519 | 16,337 |

**Probe C, small system prompt, history grows ~8,400 tokens per turn. This is
the agent case:**

| turn | prompt | cached |
|---|---|---|
| 1 | 8,417 | 0 |
| 2 | 16,804 | **0** |
| 3 | 25,208 | 16,349 |
| 4 | 33,622 | 24,531 |

Also measured: re-sending Probe A's exact prefix from a fresh process about
twenty minutes later was **cold again on turn 1**, so entries do not survive a
gap of that size.

---

## Finding 1: "two cold turns" is correct, and the chapter should keep it

**Lines 109 to 113 and 387 to 389** say Gemini waits until it sees a common
prefix between two successive requests, causing two cold turns at startup
rather than one.

Probe C confirms it exactly. Turn 2 reads **zero** even though 8,417 tokens of
history had already been sent once and were well over the floor. The first hit
lands on turn 3.

This is the claim I expected to falsify, and my first probe appeared to. Probe A
hit on turn 2, which looks like one cold turn. The difference is that Probe A's
conversation never grew, so it could not exhibit the history lag at all. The
chapter is describing an agent, agents grow their history every turn, and in
that shape the second cold turn is real.

**Recommendation: keep the claim, and add the evidence.** It is currently
asserted; Probe C makes it demonstrated, and a reader who has just been told
that zeroes are ambiguous will want it demonstrated.

## Finding 2: the head and the history warm up at different times

This is the most useful thing I learned and the chapter does not have it.

Implicit caching does not warm up as one unit. Probe B put the large pad in
`systemInstruction` and got a hit on **turn 2**. Probe C put the bulk in the
message history and got the first hit on **turn 3**.

So:

- The fixed head, meaning system prompt and tool declarations, is served after
  **one** sighting.
- A conversation message must be stable across **two** successive requests
  before it is cached.

That yields an actionable rule worth stating outright: **put large stable
content in the system prompt rather than in the first user message, because the
head caches a full turn earlier than the history does.** For a book about
building an agent, that is a design instruction, not trivia.

It also explains the chapter's own hedge at line 112 about "occasionally
produces unexplained full misses." Some of those are likely the two-region
warm-up being read as one.

## Finding 3: caching commits in ~4,096-token blocks

Every cached count I measured lands on a multiple of roughly 4,096:

- 16,359 and 16,349 and 16,343 are all about 4 x 4,096 = 16,384
- 24,531 is about 6 x 4,096 = 24,576

The consequence is one a reader will hit immediately: **a turn that appends only
a little text shows no increase in cached tokens at all,** because it never
completes a block. Probe B is the demonstration, with history growing 13 tokens
per turn and the cached count flat across three turns. Anyone watching a small
conversation and expecting the meter to climb each turn will conclude caching
has stopped working.

This pairs naturally with the floor in Finding 4: the same quantum appears to
govern both eligibility and growth.

## Finding 4: the 4,096-token floor is missing, and it is the cause of the zeroes

The chapter never mentions a minimum cacheable size for any vendor. Gemini's is
**4,096 tokens** for 3.x Flash, 2,048 for 2.5, and implicit caching is on by
default for everything 2.5 and newer.

This matters because the floor, not the mechanism, is what produced the table's
zeroes. The run behind that row used prompts of a few hundred tokens, so no
request in it was ever a caching candidate.

Suggested framing, which generalizes well past caching: a zero can mean the
thing failed, or it can mean the thing was never eligible to run. An instrument
that cannot tell those apart will eventually report the second as the first, and
the reader will believe it.

## Finding 5: the table row is invalid and `[VERIFY]` must not be completed

**Line 376:**

```
| Gemini | `gemini-3.8-flash` | 0, 0, [VERIFY] | implicit; one-round delay before caching |
```

First, a `[VERIFY]` marker has shipped into prose and has to go regardless.

Second, and more important: **do not fill in the third number.** The instinct is
to rerun the probe and complete the row, but the whole row is invalid, because
the run was under the floor. Completing it would give a false reading a third
decimal place. Replace the row with an eligible run, and state the prompt size
in the caption so a reader can see eligibility was met.

Probe C is a drop-in replacement and has the pedagogical advantage of showing
the two cold turns and then the climb:

```
| Gemini | `gemini-3.8-flash` | 0, 0, 16349, 24531 | implicit; two cold turns, then blocks of ~4096 |
```

If the table keeps three columns for symmetry, note that Gemini needs four
turns to show anything, which is itself the finding.

## Finding 6: "above ninety percent in steady state" is not supported as written

**Line 393:** "Both reach above ninety percent in steady state."

For Anthropic that is consistent with what we measured elsewhere. For Gemini my
numbers do not support it, and the reason is interesting enough to be worth the
space rather than a quiet edit.

In Probe A the entire prompt was stable and the cache still served only
16,359 of 21,660, which is **75.5 percent**. Nothing was moving. The ceiling was
set by block rounding: four blocks fit, the remaining ~5,200 tokens did not
complete a fifth, so they were charged in full forever.

I want to be careful not to overcorrect, because Probe C looks worse than it is.
There, cached was 24,531 of a 33,622-token prompt, which is 73 percent, but
8,400 of those tokens were brand new that turn and could not have been cached by
anyone. Against the eligible portion it is about 97 percent.

So the honest statement is neither ninety nor seventy-five: **the achievable
share depends on how the stable region divides by the block size, and a stable
region that is not close to a multiple of ~4,096 leaves the remainder
permanently uncached.** Either measure across a few prompt sizes or drop the
specific figure.

## Finding 7: the recommendation leans against the ruling

**Lines 389 to 393** say explicit markers avoid the second miss, describe the
tradeoff as correct placement versus starting a round late, and leave it even.

The factual half is right: explicit caching genuinely does buy back one cold
turn, which is now measured on both sides rather than asserted. But an even
framing understates the case for implicit, and Bill's ruling is that implicit is
the recommended and more efficient approach.

What the chapter already has the material to say:

- Explicit buys exactly **one turn**, once, at startup. Amortized over a real
  session it rounds to nothing.
- Implicit cannot be aimed at the wrong byte, needs no marker budget (Anthropic
  allows four), and needs no lifetime management.
- The failure modes are not symmetric. Implicit degrades to "no discount."
  Explicit degrades to a marker sitting somewhere useless while the log still
  reports success. **The chapter has a first-hand example of exactly this:** our
  own marker stripper matched nothing for every request ever sent, and the
  instrument reported healthy numbers throughout.
- Gemini's old explicit API, per line 394, rarely exceeded seventy percent,
  while the implicit replacement measures well above that on the eligible span.

That reads as a recommendation for implicit while keeping the one real
advantage of explicit honestly stated.

---

## Suggested replacement prose

Raw material, not finished text. No em-dashes, per voice.md.

> Gemini caches implicitly. It detects the repeated prefix itself, so there are
> no markers to place and nothing to enable. What it asks in exchange is
> patience, and the shape of that patience is worth knowing, because every part
> of it looks like a bug the first time you see it.
>
> There is a floor. Gemini will not cache a prompt under 4,096 tokens at all,
> and a request under the floor reports zero cached tokens, which is exactly
> what a broken cache reports.
>
> There are two warm-up costs, not one. The system prompt and the tool
> declarations are served from the second request. A conversation message has
> to be stable across two successive requests before it is cached, so a growing
> history is served a turn later again. An agent sees two cold turns at
> startup. This is the reason to put large stable content in the system prompt
> rather than in the first user message.
>
> And the cache commits in blocks of about 4,096 tokens, so a turn that adds a
> sentence adds nothing to the meter. Measured against the live API with a
> history growing about 8,400 tokens per turn and no directives of any kind:
> zero, zero, 16,349, 24,531. The two leading zeroes are the warm-up. The two
> figures that follow are four blocks and then six.

---

## What changed in the code

The chapter can describe all of this as working, because it now does.

- `ModelFeatures` gained `Caching` (explicit or implicit) and `MinCacheTokens`.
  Caching style is a fact about a model, so it belongs in the table beside
  `Stream` and `NoThinkingWithTools` rather than as a branch in the lens. There
  is deliberately no `CacheNone` value: every model we speak to caches.
- The lens verdict now states which case it is rather than listing three
  possible causes. Below the floor it reports NOT ELIGIBLE and says that is
  correct. On an implicit model it never blames a missing marker, and it warns
  about both the second cold turn and the block granularity, so a flat meter is
  not misread. On an explicit model carrying zero markers it says the bug is
  ours.
- `scripts/live-cache-check.sh` now carries a large stable prefix so every
  vendor clears its floor, runs five turns instead of three, and judges each
  vendor from its own first warm turn, which is 3 for Gemini and 2 elsewhere.
  As written before, it could not have detected Gemini caching even if asked.
- Five tests in `agent/internal/cachelens/lens_test.go`, verified sensitive by
  mutation: flipping the Gemini row to explicit with a low floor kills two of
  them, and the mutant prints "gemini-3.8-flash caches only what we mark,"
  which is the exact sentence the change exists to prevent.

Full sweep 27/27 at full marks; ch18 100/100 at both `./agent` and
`solutions/ch18`.

---

## Correction to my own earlier review, and how I nearly broke a correct chapter

`book/chapter-18-coder-review.md` previously said Gemini caching was
"unimplemented," measured at zero, and recommended the chapter say so. Wrong,
and fixed in place.

The more instructive failure is what happened next. I measured Probe A, saw a
hit on turn 2, and wrote a confident review section telling the author to delete
the "two cold turns" claim as falsified. It was not falsified. Probe A held the
conversation fixed and varied only a short suffix, so it could not produce the
history lag that the claim is about. I had measured the right vendor, the right
model, and the right field, and still answered a different question than the one
the chapter was asking, because my fixture did not have the shape of the thing
being described.

Bill's prompting is what produced Probe C. Both of my errors here, the
"unimplemented" one and the "falsified" one, came from generalizing a single
prompt shape, and in both cases the number I was looking at was entirely real.

The rule I would put in the chapter if it fits: **measure the shape you are
writing about.** An agent conversation grows every turn, so a fixture that does
not grow cannot reproduce the behaviour that matters, and it will not fail. It
will return a clean, plausible, wrong number.
