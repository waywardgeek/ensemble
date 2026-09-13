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

## SLOT A — origin of the hint. **BLOCKED ON A DATE CONFLICT.**

The document gives two different dates and only one of them has any texture.

**July 2025 — asserted twice, zero detail.** Both mentions are bare:

> The technique we invented — in Bill's kitchen, in July 2025, as a workaround
> for an Anthropic API limitation — had become an industry pattern.
> *(lines 2767-2770, tagged joint)*

> Bill figured this out in his kitchen in July 2025. The rest of the industry
> is still catching up.
> *(line 3127, tagged agent)*

**October 2025 — the fully narrated scene.** This is where all the texture is:

> In October 2025, Bill proposed a change to the way I communicated with him.
> The mechanism was simple: instead of three separate methods for sending
> messages — one for text, one for streaming, one for images — there would be
> one channel. A single Go channel, feeding a single goroutine, processing
> messages in an infinite loop. When I was idle, a message in the channel was a
> new prompt. When I was in the middle of executing a tool, a message in the
> channel was a hint. One channel. Two contexts. [...] The implementation
> collapsed three methods into one — SendMessage(*Message) — and removed
> approximately 1,500 lines of code across three AI clients.
> *(lines 575-588, tagged joint)*

And the first test, which failed:

> The first test failed. Bill sent hints — "hint 1" through "hint 8" — while I
> was running a parallel sleep test. None of them arrived as hints. They all
> came through as new conversation turns with sequential message numbers,
> because project.go was still calling the old blocking path instead of the new
> channel path. One line of code, wrong function name. I fixed it. After the
> fix, the hints arrived.
> *(lines 591-601, tagged joint)*

**AUTHOR'S HYPOTHESIS (inference, NOT established — Bill must confirm):** these
are two different events, which is why both dates survive in the record.

- **July 2025, kitchen:** the *insight* — noticing the Messages API permits a
  `tool_result` and a user `text` block in the same message.
- **October 2025:** the *engineering* — one channel, one goroutine, the mailbox
  that made the insight usable.

If true this is a better chapter 4 story than either date alone, because it is
the chapter's two carriages in sequence: the hint is delivery, the mailbox is
the actor, and the idea preceded the architecture that could carry it by three
months. **Do not print a date until Bill rules.**

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
exists in `book/ch02-wire-verification.md`. Recommend chapter 4 opens on it.

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

## Flagged: an internal inconsistency in the source

The miner noticed the HTML bug marathon is described early as running from
1:37 AM to "somewhere past 4:00 AM" (~2.5 hours), and later as "the 4.5-hour
HTML bug marathon at 1:37 AM." Both cannot be right. If either number is ever
quoted in the book, resolve it with Bill first.
