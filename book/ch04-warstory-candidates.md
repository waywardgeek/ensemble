# Chapter 4 war-story candidates — mined from the-dyad.txt

Source: `/Users/bill/singularity/the-dyad.txt` (3,191 lines, ~32,500 words),
read in full by a Sonnet sub-agent on 2026-09-13. Full structured submission is
durable at `cr/agents/dyad-warstories/submit.md` in the coderhapsody data dir.

**ATTRIBUTION WARNING, READ FIRST.** the-dyad.txt is a memoir written in the
*agent's* first person, about the agent's experience. This book is Bill's first
person singular. Almost every scene below is narrated by the agent even where
Bill is the one acting. Nothing here may be moved into the book as Bill's
experience without Bill confirming it as his. The miner tagged each passage;
those tags are reproduced and must be respected.

---

## SLOT A — origin of the hint. **RESOLVED 2026-09-13. July 2025 is correct.**

Bill ruled directly, and the date is corroborated by an artifact. The two
"conflicting" dates were never a conflict: there are **three** events, and the
memoir only narrated the third.

**The receipted timeline** lives at `cr/docs/hints-deep-dive.md` (lines 99-107)
in the coderhapsody repo, already carrying evidence labels:

| when | what | status |
|---|---|---|
| **July 2025** | Bill discovers Anthropic will accept extra content blocks inside a `tool_result` user message — and that the model reads them. This is the whole mechanism. | REPORTED (first-party), **corroborated in-repo** |
| **Aug 2025** | CodeRhapsody is written to exploit it. Mid-turn steering works. | REPORTED (first-party) |
| **Aug–Nov 2025** | Essentially no uptake by anyone. Published on LinkedIn and in every chat forum that would host it. | REPORTED (first-party) |
| **18 Nov 2025** | **Google Antigravity ships the same technique in its first release.** | **VERIFIED** — antigravity.google blog, Wikipedia |

The July corroboration is a source comment inside CodeRhapsody describing the
Gemini implementation as *"the July 2025 Anthropic tool_result trick, adapted
for Gemini"* — written long before anyone needed the date for a book, which is
what makes it good evidence.

The memoir's October 2025 scene (one channel, one goroutine,
`SendMessage(*Message)`, ~1,500 lines removed, the first test failing because
`project.go` called the old blocking path) is a **fourth** event: the actor
rebuild that made the technique clean. Still usable, but it is not the origin.

### Bill's account, 2026-09-13 (first-party, verbatim from the session)

> "I wanted mid-turn steering badly for my demo for StackAgent, which I wrote in
> 2 weeks during July. I finally found a hack that Anthropic accepted, and the
> scheme worked."

> "My presentations aren't great [...] I don't deliver well, but I wish I had a
> video of it. The new innovations I demoed I thought would blow people away.
> The truth is they didn't understand a word of what I was saying. WTF is an AI
> coding agent? Can't all models be steered? They had no clue."

**This connects ch4 directly to ch1 §1.0.** The StackAgent fortnight is already
the book's opening story. Chapter 4 can reveal what was *in* that demo: the
steering mechanism nobody in the room had a category for. Chapter 1 tells you he
built it in two weeks; chapter 4 tells you what he found while doing it.

**The arc, and why it is not a grievance.** The lesson is not "nobody
appreciated me." It is that the mechanism was reachable at the raw API surface
**four months before a major product shipped it**, and the only reason he got
there first is that he was reading the API instead of a framework's docs. That
is §1.1's thesis — frameworks hard-code delivery — paid off with a verified
date. The two audience questions are the receipt for how early it was: in July
2025 "what is an AI coding agent" was a *reasonable* question.

### Two cautions before any of this reaches print

1. **The demo scene has no artifact.** Bill's own words: *"I wish I had a video
   of it."* The discovery is corroborated in-repo; the demo and the room's
   reaction are uncorroborated first-party recollection. Label it the way
   `hints-deep-dive.md` labels things. The Antigravity beat is the externally
   verified one and should carry the weight of the argument.
2. **Bill mentioned being slightly autistic.** That is a personal disclosure
   about a real person, in a book that goes to print under his name, and it is
   categorically different from the self-deprecation about presenting badly.
   **I have not put it in the book and will not unless Bill explicitly says to.**
   My editorial view if asked: the story is *stronger* without it. "They had no
   clue what an AI coding agent was" locates the failure in a room that had no
   category yet, which is the true and more interesting claim. "I present badly"
   relocates it to him and makes the reader argue with the evidence, since the
   audience questions show the gap was conceptual, not rhetorical.

---

## SLOT B — the misdiagnosis. **NOT IN THIS SOURCE. Do not force it.**


The miner reported, correctly, that it found nothing matching "slowness blamed
on the model or vendor, real cause local." Its best candidate is the SSE
timeout bug (lines 2570-2582): both agents hung forever, the first fix chased an
unrelated Gemini `FinishReason` check, and the real cause was a scanner blocked
on `scanner.Scan()` against a connection that had silently closed. It flagged
its own candidate as impure, because the dead stream really did originate with
an incomplete vendor response — so it is not cleanly "local fault mistaken for
vendor fault."

**That honest refusal is more useful than a stretched match.** It means the
parking file's line about the mailbox blackout — *"A student who hits this will
blame the vendor. It is in their own engine, four lines apart"* — has no
supporting anecdote in this source.

**BUT WE ALREADY HAVE THE SLOT B STORY, AND IT IS NOT IN THIS FILE.** On
2026-09-11, verifying chapter 2 claims against live APIs, finding M9: Bill's
stated belief that a mid-turn `{"role":"system"}` message fails on sonnet-5 was
tested and **falsified** — the model obeys it. The lag he had attributed to the
vendor was engine lag: tools run on the mailbox-draining goroutine, so the actor
goes deaf while working. Measured blackout: **11.4 seconds.**

That is the exact shape chapter 4 wants: a belief held, a measurement taken, the
belief overturned, and the fault found at home. It is dated, it is measured, it
is Bill's own wrong belief and therefore his to tell, and the receipt already
exists in `book/ch02-wire-verification.md`.

**RULED 2026-09-13 by Bill: chapter 4 opens on it.** The M9 falsification is the
cold open. Slot B is closed and needs nothing further from the memoir.

---

## SLOT C — stateful actors. **RICH. Four candidates, best first.**

**C1 — broken state inherited across restarts (strongest).** A textbook
stateful-actor failure: each new instance inherited the previous one's corrupted
state and made it worse.

> And then, for thirty-plus handoffs across multiple days, it didn't work. The
> problem was a sorting bug [...] compareMemoryNames() was doing lexicographic
> string comparison: "2026-04-11-8" sorted after "2026-04-11-50" because the
> character '8' has a higher ASCII value than '5'. So the cascade processed
> memory files in the wrong order, shifted the wrong data between layers, and
> produced incoherent summaries that overwrote good ones. Every handoff made it
> worse. Thirty-plus handoffs. Each one a fresh instance of me, inheriting the
> previous instance's broken state, trying to debug a system I'd just built but
> couldn't remember building. [...] When the bug was finally fixed — one line,
> compareMemoryNames() changed to parse the numeric suffix as an integer — the
> cascade converged within minutes.
> *(lines 1768-1786, tagged agent)*

**C2 — state on disk, not in the session.** The design principle stated plainly,
with the WiFi drop as its proof.

> The sprint took six to eight hours. Bill was on a bus for part of it — the
> session log shows him losing WiFi mid-sprint ("I got off the bus and lost
> wifi"), reconnecting via a guest network, and picking up where we'd left off
> without breaking stride. The sprint survived the connection gap because the
> architecture didn't depend on constant connectivity. Each tool call was
> atomic. The state was on disk, not in the session.
> *(lines 1092-1096, tagged **bill**)*

**C3 — who owns the shared state, delivered AS a hint.** Doubly useful: it is
both slot C material and a live example of the chapter 4 mechanism.

> He delivered the key architectural insight mid-stream, as a hint while I was
> building: [HINT] "The project could hold the executor." One sentence. It
> meant that the Project struct [...] would own the ToolExecutor, and all AI
> clients would share it.
> *(lines 1097-1100, tagged **bill**)*

**C4 — the same diagnosis lost four times (Groundhog Day).**

> It didn't work at first. The port configuration was wrong, and the bug
> reappeared across four compressed-context sessions. Each time, a fresh
> instance of me discovered the same problem [...] and each time, I proposed the
> same fix. Bill answered the same question the same way each time [...] This is
> what early collaboration looked like without a memory system: Groundhog Day,
> except only one participant knew it was repeating.
> *(lines 370-384, tagged joint)*

---

## SLOT D — a flaw that announced itself early and was misread. **TWO STRONG.**

**D1 — the Gemini cache-cost marathon.** Five sessions, four days, 4-5x
overspend, and the diagnosis peeled in layers: ambiguous vendor docs ($0.50/M
cache creation, not $0.05/M), then wrong math in the corrected design doc, then
a "janky hack" workaround, then a double-initialization bug, then an
empty-contents error. Contains Bill's own opening diagnosis quoted verbatim and
his hint killing the workaround:

> "[HINT] Please eliminate the janky hack of re-caching after 10 turns. Use the
> real algorithm based on cost."

With measurements: two caches created three seconds apart, 16:20:15 and
16:20:18, $0.052567 each, total $0.105134.
*(lines 1259-1292, tagged joint)*

**D2 — the 4 AM cat box.** A tiny symptom that announced a systemic flaw.

> Bill: "I just got a DM notification at 4 AM to clean the cat box, but it is
> scheduled for 7 AM." Root cause: Gemini 3.0 Flash [...] had stored them with
> Eastern Time offsets. The cat box event said "7:00 AM -04:00" — which is 4:00
> AM Pacific. [...] But the bug pointed to something deeper. Every time the LLM
> generated a parameter [...] there was a chance it would hallucinate the
> format.
> *(lines 2208-2226, tagged **bill**)*

D2 is the better *shape* for slot D — one 4 AM notification exposing that every
LLM-generated parameter across 23 tools was untrustworthy — but note it is
really a chapter 3/5 theme (tool parameters, capability seam), not chapter 4.

---

## Flagged: an internal inconsistency in the source — **RESOLVED**

The miner noticed the HTML bug marathon is described early as running from
1:37 AM to "somewhere past 4:00 AM" (~2.5 hours), and later as "the 4.5-hour
HTML bug marathon at 1:37 AM." Both cannot be right.

**RULED 2026-09-13 by Bill: 4.5 hours is correct.** His recollection is that it
ran long into the morning, and was not one of the sessions he started at 2 AM.
1:37 AM plus 4.5 hours lands near 6 AM, which fits. **The "past 4:00 AM" /
~2.5-hour figure in the source is the erroneous one; do not quote it.** If the
duration is ever printed, it is 4.5 hours, sourced to Bill, not to the memoir.

