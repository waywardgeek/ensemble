# Chapter 20: Crossover

*Outline for review. This is the second-to-last chapter of edition 1.*

---

## Through-Line Stake

Bill replaces the agent that wrote eighteen chapters with the agent those
chapters describe. Over two weeks, every architectural decision either holds
or breaks against real work — and the bugs that never announce themselves turn
out to be the ones that matter most.

**Resolution.** The architecture holds. The bugs are boundary bugs, not design
bugs. But the instrument needed to catch the invisible ones — the ones that
succeed while failing — is harder to build than any feature in the book. When
the agent can see its own failures, it starts fixing itself, and the crossover
stops being an event and becomes a loop.

---

## Motivational Opener (Bill's voice, §1.2)

The morning of the switch. Two agents on the screen: CodeRhapsody, the one
that wrote this book, and Ensemble, the one the book describes. Same hardware,
same model, same prompt. One of them built by reading eighteen chapters of
instructions. The other built by writing them.

The bet: if the agent that came out of this book cannot replace the one that
went into it, then eighteen chapters taught a toy. If it can, the last chapter
writes itself.

*[2–3 paragraphs. End at the switch. The section after it is third person.]*

---

## TL;DR

Crossover is the moment the agent built from this book's designs becomes the
daily driver. The graders scored it 100/100. The real exam is two weeks of work
with a bill, a deadline, and an operator who has to trust the output with no
second opinion. The bugs that surface sort into three kinds: the ones that
crash on day one (easy), the ones that succeed while failing (expensive), and
the ones that violate rules this book already taught (embarrassing). Each kind
needs a different instrument, and the instrument that catches invisible failures
is harder to build than the failures themselves.

---

## 20.1 — What Breaks on Day One

**Source entries:** 1–3, 7–9, 13

**Wild fact:** Every assistant response rendered inside the first one's bubble,
because part IDs reset to 1 on each API response. The GUI keyed artifacts by
ID, and every turn's first text block was `part_id: 1`.

**Point.** The first bugs are loud. Ctrl-C kills the process without saving.
The GUI opens blank on a restored conversation. Settings look like they take
effect and never persist — `sendPatch` was called on every change event and was
never defined. These are demo gaps: things that work when the session lasts one
turn and break when it lasts a day.

They are easy to fix, and they teach nothing the previous chapters did not
already know. They exist in the chapter to establish a baseline: the agent
works, if the definition of "works" is narrow enough. The next section raises
the bar.

**Subsections:**
- Part IDs are identities, not indices (entry 2)
- Ctrl-C and the exit path that deferred saves make reachable (entry 8)
- The blank screen: a hub that believed its own log was empty (entry 9)
- The settings dialog where every control worked and nothing saved (entry 13)

**Estimated words:** 800

---

## 20.2 — The Bugs That Succeed

**Source entries:** 20, 28–29, 32, "the vendor is a property of the model"

**Wild fact:** The model picker displayed "GPT-6 Astra" while every request
went to Claude Sonnet. The session meter confirmed the selection. Both lied —
the meter read from *intent*, not from the model actually dialed.

**Point.** The second category of crossover bug is the invisible failure: a bug
where every request is answered, the GUI looks right, the tests pass, and the
only thing that disagrees is the invoice — or the model — or the credential —
thirty days later.

These share a shape: the *visual* half of the operation works while the
*functional* half never runs. The model selector that selected nothing (entry
20), confirmed by a meter that read from the wrong source. The credential
provider that was not carried across a vendor switch, so flipping the picker
from Claude to GPT rendered the request correctly, sent it correctly, and
billed it to the metered API key instead of the plan (entry 29). The startup
resolver that settled the vendor first and the model second, so a model picked
in the GUI was sent to whichever vendor happened to be the startup default.

The fix for the meter is the lesson for the section. A readout fed from
*intent* always agrees with the operator. Only a readout fed from the *fact*
can disagree, and disagreeing is the entire job.

**Subsections:**
- The model selector that selected nothing (entry 20)
- The instrument that confirmed the lie (entry 14's meter, repointed)
- The credential that did not cross a switch (entry 29)
- The vendor that was not a property of the model (unnumbered entry)
- Logging which credential is in use, never its value (entry 32)

**Estimated words:** 1,200

---

## 20.3 — Two Paths, One Screen

**Source entries:** 17–18, 27, 34

**Wild fact:** The GUI classified prompts and hints itself, violating the rule
stated in chapter 2 — the rule taught specifically so it would not be violated
here. The fix was a deletion plus a move: remove the guess, let the actor
classify by its own position in the drain loop.

**Point.** The live streaming path and the replay path are different code paths
that must produce equivalent output. When they disagree, the result is a bug
that is invisible during development (when only live streaming runs) and total
after a restart (when only replay runs).

The assistant text that vanished on reconnect: `partToWireMsg` produced finals
with no `part_id`, and the GUI routed every replayed response to a single
invisible DOM element (entry 17). The reset that cleared the screen for
everyone except the operator: `Engine.Record` appended to the log but did not
notify the observer, so the live channel never learned about it while the
replay channel self-healed on refresh (entry 18). The client that decided
whether a message was a hint or a prompt by branching on its own copy of the
turn state — racing the server, incorrectly, in exactly the way chapter 2 said
not to (entry 34).

The two-renderer architecture — CLI and GUI observing the same actor — was
built to diagnose these bugs (entry 27). The diagnosis is mechanical: a bug
that reproduces in both is an agent bug; a bug that reproduces in one is a
front-end bug. Writing a second renderer against the same seam is the cheapest
test of whether a seam is honest.

**Subsections:**
- The assistant text that vanished on reconnect (entry 17)
- The reset nobody saw (entry 18)
- The terminal as a diagnostic tool (entry 27)
- The client that guessed: chapter 2's rule, violated and corrected (entry 34)

**Estimated words:** 1,200

---

## 20.4 — What You Can't See Can't Be Fixed

**Source entries:** 11, 14–15, 21, "cache lens that cried wolf," "cache that
ignores small conversations," "plan route does not cache"

**Wild fact:** The cache alarm had a 100% false positive rate — it fired on
every single request — which made it indistinguishable from an absent alarm.
The one genuine cache break in a morning's session looked identical to the
hundred harmless appends around it.

**Point.** Caching is the one failure mode that produces a correct answer.
Nothing errors. Nothing retries. The session looks normal. The only artifact
that disagrees is the bill.

Chapter 18 teaches the mechanism: breakpoints, placement, the four-slot budget.
This section tells the story of what happens when that mechanism meets a
production session and the operator burns 46% of a weekly allowance in a
morning. The cache lens was built to catch exactly this, and it had been running
the whole time, and every line it printed was a false alarm — because the
OpenAI Responses API names its dialogue `"input"` and the canonicalizer only
knew `"messages"`. An instrument with a 100% false positive rate carries zero
information: the single genuine break is indistinguishable from the hundred
harmless appends.

The fix restored the distinction. The next finding was immediate: the plan
route has a cache-write floor of roughly 22,000 tokens, far above the
documented 1,024. Below it, a perfectly stable prefix gets nothing. This
inverts the usual instinct: past the floor, a longer conversation is cheaper
per turn than a short one.

And past both the floor and the lens fix, 137 requests still went uncached on
requests of 55–73K tokens. That led to filing a bug with the provider: the plan
route with OAuth tokens does not cache at the level the metered route does. The
fix is expected within weeks. Until then, the crossover is real and the bill is
not yet affordable.

**Subsections:**
- The meter: making the money visible (entry 14)
- The lens: measurement before markers (entry 11)
- The alarm that always fires (cache lens that cried wolf)
- Breakpoint placement: four slots, one rule (entries 15, 21)
- The floor that inverts the instinct (cache floor entry)
- Filing the bug: what the agent cannot fix (plan route confirmation)

**Estimated words:** 1,500

---

## 20.5 — The Agent Looks in the Mirror

**Source entries:** 22–23, 33, "engineering standards," "stale actor
completion"

**Wild fact:** The agent could not see its own GUI. The MCP server was built
and documented in chapter 13, and nothing connected to it. The grader never
noticed, because it tested the MCP client against a stdio fake, not the path
from agent to browser.

**Point.** Three capabilities close the self-improvement loop.

First, `view_gui` lets the agent look at its own screen through the MCP server
the book already built (entry 22). Second, `RequiresVisibleReasoning` enforces
that a model narrate before acting, because a silent run cannot be reviewed
and the only safe response is `git reset --hard` — which, when it happened,
left three untracked files behind that did not compile (entry 33). Third,
engineering standards written into SKILL.md after watching a different model
make the same three preventable mistakes every new contributor makes: deleting
load-bearing comments, committing unformatted code, editing without asserting
uniqueness. The standards are mechanisms, not instructions, each stating the
reason it exists, because a terse rule gets optimised against.

Together these make the agent improvable *by* the agent. That is the property
that turns crossover from a one-time event into a continuous one.

**Subsections:**
- The door that was built and never opened: view_gui (entries 22–23)
- The model that worked in silence (entry 33)
- Engineering standards as mechanisms, not instructions (standards entry)
- A stale completion that the graders could never trigger (stale entry)

**Estimated words:** 1,000

---

## 20.6 — The Book That Rewrites Itself

**No specific source entries. This is the synthesis.**

**Point.** The eighteen chapters before this one taught the mechanism: the
loop, the log, the seams, the tools, the skills, streaming, persistence,
context, memory, recall, caching. This chapter is the exam. The grade is not a
number but a question: is the agent, running on itself, good enough to edit
this page?

The answer is yes, with one asterisk: the bill. When the caching bug is fixed,
the asterisk goes away. The crossover taught one thing the graders could not:
the bugs that matter most are the ones that succeed. An agent that can see its
own failures — measure them, log them, refuse to act without explaining — is an
agent that improves from use rather than from instruction. That is the
difference between a tool and a partner.

The next chapter says what comes after.

**Estimated words:** 500

---

## Budget

| Section | Words (est.) |
|---------|-------------|
| Opener + TL;DR | 500 |
| 20.1 What Breaks on Day One | 800 |
| 20.2 The Bugs That Succeed | 1,200 |
| 20.3 Two Paths, One Screen | 1,200 |
| 20.4 What You Can't See Can't Be Fixed | 1,500 |
| 20.5 The Agent Looks in the Mirror | 1,000 |
| 20.6 The Book That Rewrites Itself | 500 |
| **Total** | **~6,700** |

Within the 4,000–7,500 soft ceiling. Room for code blocks (the part-ID
bubble, the `sendPatch` ReferenceError, the cache-lens false alarm, the
hint-classification deletion).

---

## Entries Deferred (in the changelog, not in the chapter body)

The following are real crossover work but are too detailed for the chapter's
narrative arc or are better told as mechanism in their own chapters:

- **Entries 4–7** (settings tabs, model catalog, sectioned file, font size) —
  UI polish. Contribute to the "demo gap" point in 20.1 but are not individual
  stories.
- **Entry 10** (composer grows with the message) — UI detail.
- **Entry 12** (sectioned settings, reverted) — interesting shape (the struct
  is the published contract) but covered by 20.2's theme.
- **Entry 16** (reset as an event, not an edit) — a mechanism point. Told in
  the chapter as part of 20.3 but not narrated at entry length.
- **Entry 19** (view limit vs. retention limit) — mechanism detail, better as
  a reference.
- **Entries 24–26** (plan route specifics, delta/final ID reuse) — ch19
  follow-up material.
- **Entries 30–31** (Responses migration, snapshot drift) — ch19 consequences.
- **Lost-result placeholder, replay duplication, single-loop retirement** —
  detailed engineering fixes, worth having in the changelog for reference but
  not chapter stories.

---

## Voice Notes

- Third person throughout the body (§1.1). Bill's voice in the opener only.
- The confessions budget (3 per chapter) will be spent on: the cache lens that
  cried wolf, the hint-classification violation of chapter 2's own rule, and
  the meter that confirmed the lie. Each is a real mistake with a real fix,
  not a narrative device.
- The wild facts are all receipts: part ID 1, the model picker's value, the
  100% false-positive rate, the 22K floor. Each is verifiable against an
  artifact.
- Chapter 2 callback in 20.3 is the strongest structural beat: the book's own
  rule, taught in the second chapter, violated in the shipping client,
  corrected by a deletion. That is the one the reader will remember.
- No company names except where ch19's exception already applies (the OpenAI
  bug filing). The chapter refers to "the provider" and "the plan route."

---

## Relationship to Adjacent Chapters

- **Chapter 18 (Caching)** teaches the mechanism. 20.4 tells the story of
  applying it and finding that the instrument was broken. No mechanism is
  retaught; the narrative assumes the reader built the breakpoints.
- **Chapter 19 (Leaving Anthropic for OpenAI)** teaches the vendor switch.
  20.2 and 20.4 tell what happened when that switch was exercised daily.
  The entries about the plan route's quirks (24–26) stay in ch19; the
  billing consequences land here.
- **Chapter 21 (Forward-Looking)** picks up from 20.6. The crossover proves
  the architecture works; chapter 21 says what gets built next.

---

## Open Questions for Bill

1. **Title.** "Crossover" is loaded and simple. Alternatives: "Crossover, or
   the Morning Everything Worked" / "The Agent That Replaced Its Teacher."
   The single word might be strongest.
2. **Code in the chapter.** The changelog entries carry code blocks. The
   chapter should carry fewer — perhaps the part-ID allocator, the
   hint-classification deletion, and the cache-lens false alarm. The rest
   are better as prose descriptions of what changed. How much code does the
   reader want to see in this chapter vs. the previous eighteen?
3. **The caching story's weight.** 20.4 is the longest section because the
   investigation has the most narrative tension. If it draws too much
   attention from the crossover thesis (the invisible-failure class), it
   could be compressed — but the cache lens that cried wolf is the chapter's
   best single story.
4. **Grading.** Should this chapter have a grader and frozen solution like
   the others, or is it a narrative-only chapter? The crossover is not a
   feature the student builds; it is a story about building one. If graded,
   the check would be: the student's agent passes all 19 previous chapter
   graders on a fresh install. That is the crossover test.
