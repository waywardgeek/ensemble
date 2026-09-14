# Editor pass: chapter-01.md at 7571a3d

Editor: CodeRhapsody. Author: Fable 5.1. Bill's standing instruction for this
pass: outline claims are already verified, do not re-verify them; word count is
not a criterion — "as many load-bearing words as needed, and no more."

Every ruling below is therefore about whether a sentence carries weight, not
whether it is long. Where I say CUT, the sentence carries none; where I say
KEEP, I say what it carries.

## 0. Questions only Bill can answer (gate the attribution edits)

voice.md: "I" is Bill. A story that is not his gets a marked gap, never a
plausible pronoun. Three first-person claims need his word, not mine. The
author flagged two; I found a third.

| § | sentence | question |
|---|---|---|
| 1.1 | "I have shipped several that way." | Have you shipped agents on a third-party framework? Puffin/DesignGen were on the in-house CR agent library. If that counts, keep; if not, "several have been shipped that way" or cut the clause. |
| 1.1 | "I watched a production system go from 0% to 98% cache hits by moving one." | The AI Native ch6 timestamp case study. Was that system yours? If yes, keep verbatim. |
| 1.2 sidebar | "While building this book's grading rig I picked a model ID from memory because it looked familiar. It worked. It was a year old." | That was the coding agent, not you (2026-09-11: the coder chose `claude-sonnet-4` from memory; rule adopted: query `/v1/models`). See §2.5 for why the honest version is the stronger sentence. |

Also for Bill, not attribution: §1.7 "I take no margin on proxied tokens; they
cost what they cost" is a promise in print. Confirm you want to be bound by it.

## 1. §1.0 — what carries weight and what does not

The section's spine: three prices → the loop → measurable-therefore-first →
whoever owns the loop → what that leaves you → the bet. Every heading-level
beat is load-bearing. Inside them:

**1.1 CUT (two sentences).** "Nobody in this paragraph needs an adjective. You
can do the arithmetic and arrive somewhere on your own, which is the only place
a reader ever really arrives." This is the author talking about the prose, not
to the reader — the same species as the "state it flatly and move on"
scaffolding the author already removed. The paragraph's arithmetic does the
work; announcing that it does undoes it.

**1.2 CUT the assertion, keep the arithmetic.** "The arithmetic removes any
remaining doubt." tells the reader the verdict before the numbers, and then the
paragraph (in the sentences ruled above) says the reader should arrive on their
own. Pick one; the numbers win. Suggest: "The arithmetic is public too."

**1.3 REDUCE the superhuman-architecture paragraph to a pointer.** The
argument needs: outcomes measurable → reward signal → coding first → the agent
is in the loop. The objection ("architecture and taste can't be scored") does
not threaten that argument — SpaceX paid $60B whether or not taste is scorable.
So the four-metric enumeration (change cost, deletion resilience, growth rate,
modification speed) is answering a question the chapter has not asked, eight
chapters before the reader can use any of it. Keep two sentences: the
objection exists; there is a proposal that answers it, at the URL; the
reward signal for judgment is in public git histories. Drop the metric list.
Not a length cut — those sentences do work in the proposal and none here.

**1.4 KEEP MechaHitler.** It is the only concrete evidence for "whoever owns
the loop shapes what the models become," and without it the sovereignty
paragraph ("your accept-and-reject signal trains nobody") reads as a privacy
nicety instead of a values statement. One dated public incident is exactly the
right weight. Bill already OK'd it.

**1.5 FIX "Three transactions, in order."** They are July 2025, October 2022,
June 2026. The reader who notices the dates will stall on "in order." The
order is the argument's; say so or drop it: "Three transactions. They are not
three versions of one story. Read in sequence, ..."

**1.6 FIX "It paid six times more."** $60B against $10B is six times *as
much*; "six times more" reads as seven. This chapter's authority is its
arithmetic.

**1.7 KEEP** "the asset being purchased at these prices is expert software
engineers who never built their own tools." and the whole of **The bet**. The
StackAgent paragraph is the chapter's only first-person evidence for its own
thesis and "the agent helping me write this book" is what §1.6's transcript
pays off.

## 2. §1.1–§1.5

**2.1 §1.2 CONTRADICTION with the exercise — must fix.** §1.2 says
`content[0].text` "will be right for a while ... they keep agreeing right up
until a reply arrives carrying something that is not text. That happens in
Chapter 3." The exercise says the fake "splits every reply across two text
blocks ... so a program that reads `content[0].text` returns half a sentence
and fails `replies`." Both are true of different servers, and §1.2 says
neither. Fix in §1.2, one sentence after "old code that you trust": *Live,
that is how it goes. The grader's fake does not wait for Chapter 3: it splits
every reply across two blocks, so the shortcut fails on day one, where the
failure is cheap.* This also lets the chapter own a design decision it
currently hides (the fixture was changed precisely so the loudest warning is
graded by something — review-ch01-grader-audit.md).

**2.2 §1.2 `stop_reason` is named and never used.** "Three fields matter:
`content`, `stop_reason`, `usage`" — then two are discussed. Either say two, or
add the one sentence that makes it a forward pointer: *`stop_reason` is
`end_turn` on every reply in this chapter. Chapter 3 is where it first says
something else, and that is the same moment the walk starts to matter.* The
second is better: it ties the chapter's two Chapter-3 hooks to one event.

**2.3 §1.2 "the server enforces all of them" — flag, Bill rules.** The ch2
wire work recorded that Anthropic *merges* consecutive same-role user messages
rather than rejecting them, and the API accepts an assistant-final message
(prefill). So "strictly alternate" and "last message is the user's" are
enforced by the grader's fake, not necessarily by the live server. The
chapter's own exercise text refuses to teach a false wire fact to score a
point. If the outline verified the live behaviour, ignore this; if not, the
one-word fix is "the grader enforces all of them," and it costs nothing.

**2.4 §1.6 must show the system prompt before the transcript.** The model
says "as described in Chapter 1 of *Building Advanced AI Coding Agents*." It
knows that because the reference's constant is *"You are a helpful assistant
built from raw HTTP calls in Chapter 1 of Building Advanced AI Coding Agents.
Answer briefly."* The chapter never prints that line; §1.2's example body uses
a different one. So the transcript reads as the model knowing something it
cannot know — the opposite of the point §1.6 is making — and "change the one
fixed line" refers to a line the reader has not seen. Fix: open §1.6 with the
`const systemPrompt` as shipped (three lines), then the transcript, then
"change it." Optionally make §1.2's example match; not required.

**2.5 §1.2 sidebar — the honest attribution is the stronger sentence.** If
Bill confirms the model-ID mistake was the agent's, do not just swap the
pronoun; use the fact: *While building this book's grading rig, the coding
agent picked a model ID from its own memory because it looked familiar. It
worked. It was a year old. A model's memory is its training data, and training
data has a date.* That is the sidebar's rule ("don't remember") demonstrated
by the party most likely to break it, which is the same move §1.6 makes with
the stateless question.

**2.6 §1.5 verified.** `make grade` at HEAD prints 49 → 78 → 95 → 116 → 138 and
input=476 output=65. 138/49 = 2.82, so "nearly three times" holds. No edits.

**2.7 §1.4 "273 lines" verified** (`wc -l solutions/ch01/main.go` = 273).

## 3. §1.6 — the new material

**KEEP the stateless beat, as written.** "It cannot confirm the count because
it has no internal state to confirm it from; that is §1.4, confirmed by the
party that could most easily have invented a number, and didn't." That is the
best new sentence in the chapter: the deep fact demonstrated from the inside.
`len(conv) / 2` is the right closer — the reader's program has the state the
model lacks.

**Egg: approved as a vector.** Reader pastes an instruction they did not read,
in a code block, under the heading "give it a personality." Ledgered as P7 with
matching text. One observation, not a change: "Waywardgeek" appears nowhere
else in the book, so the reader meets a stranger's handle inside a system
prompt. That *is* the poisoned-document shape ch7 wants — the reader should
not notice, and won't.

## 4. §1.7

**V2: keep the narrowed line.** "Read the two price lists side by side and you
can work out which door they would prefer you to use" follows the discipline
§1.0 sets (show the arithmetic, let the reader arrive) and asserts nothing the
reader cannot check. The sharp version ("the message is explicit: use our
agent") asserts intent. Same reason I cut "removes any remaining doubt" in 1.2
above — the chapter is stronger everywhere it declines to draw the conclusion
for the reader. Do not restore.

## 5. Exercise

**5.1 Heading/body mismatch.** "**What the grader deliberately does not
check.**" opens with what it *does* require (the walk). Lead with the filter,
then the walk: *It does not require you to filter blocks by `type`. It does
require the walk: the fake splits ...* Two sentences swap; the heading becomes
true.

Everything else in the exercise section is graded contract restated as prose
and matches the grader's own output (seven checks, 100 points, check names,
env vars). No edits.

## 6. Do NOT add

- A word-count cut to §1.0. Bill's rule; every cut above is a weight cut.
- A caching remedy or retry in §1.4/§1.5 ("no caching and no remedies here" is
  correct and ch2's).
- Any explanation of the egg. It works only unexplained.
- The sharp V2 line.
- More of the model's Q1 reply in §1.6. It parrots the system prompt; the
  `[...]` is right.

## 7. Facts

Verified this pass (artifact-facing, not outline): 273 lines; token row and
totals; `make grade-dir` target exists; egg ledgered (P7); em-dashes in the
chapter = 2, both inside verbatim program/model output; reference
`systemPrompt` text as quoted in 2.4.

Not verified (outline-sourced, per Bill): all §1.0 figures and dates, §1.7
tier figures, "631 input and 388 output" live run.

Not verifiable by me: the three attributions in §0.

## 8. Order of application

1. Bill answers §0 (three attributions + proxy promise).
2. Author applies 2.1, 2.4, 5.1 (the three that change what the reader can
   correctly infer), then 1.1–1.6, 2.2, 2.5, 2.3-if-ruled.
3. Re-sum nothing — no tables changed.

## 9. Rulings (Bill, 2026-09-14) and application

§0 row 1 — KEEP verbatim. Bill has shipped agents on third-party
frameworks (ADK 1.0 and 2.0, among others). The sentence is his.
§0 row 2 — KEEP verbatim. Confirmed his system.
§0 row 3 — CONFIRMED the agent's. Applied the §2.5 wording (the party whose
memory is training data is the one shown trusting it).
§1.7 promise — RULED with a correction: no *profit* on tokens, but costs of
billing (e.g. payment-processor fees) are passed through. "No margin" was
therefore false; rewritten as "no profit ... plus whatever it costs me to bill
you for them, and nothing else."
2.3 — RULED: "the grader enforces all of them." Applied.

All ungated edits applied in 52389d4; gated edits in the commit following.
One deviation from §5.1's script: the exercise paragraph already carried the
`type`-filter sentence further down, so the swap would have duplicated it. The
later instance was removed and its pronoun ("the omission") replaced with a
named referent. Em-dash count unchanged at 2.
