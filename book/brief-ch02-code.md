# Brief for the coder — Chapter 2 code rework

**From:** the author. **Date:** 2026-09-12.
**Spec:** `book/chapter-02-outline.md` (Draft 4).

---

## Read this before `review-ch02.md`

Your Chapter 2 review was excellent and **it reviews a chapter that no longer
exists.** Draft 3 was "the real data structures," Anthropic-only, with hints,
interrupts, a mailbox, concurrency, and the `agent_status` tool. Draft 4 is the
**LLM seam** chapter: three vendors, both directions, no tools, no concurrency.

Your findings were not discarded:

- **M3** (freeze the event vocabulary), **M6** (zero-point `session` check),
  **M7** (`render` contract: stdout only, exit 0, no flags), **M8** (reducer
  totality via "every pair not listed is identity"), **E1** (the mid-turn
  classification sidebar), **E3** (four sources of render non-determinism),
  **E6** (`Apply` is information flow, not value semantics) — all **applied and
  still live**.
- **M1, M2, M4, E4, E7** — hint and interrupt findings, **moved verbatim** to
  `book/chapter-05-actors-parking.md`. Still correct, not yet due.
- **M9** — your falsification of the model-support table stands and is recorded
  in Draft 4's claim-status block.
- **M5** (declare the tool in the request) — **obsolete**. Chapter 2 no longer
  executes tools.

## What Chapter 2 now is

One context. Three renderers, three parsers. Anthropic, OpenAI, Gemini.

- **Bidirectional.** `Render(*Context, Config) (*http.Request, error)` and
  `Parse(status int, body []byte) ([]Event, error)`. Parsing is the harder half
  and scores higher.
- **`Parse` returns events**, not a context. One path in: append events, run the
  reducer. Do not let the parser mutate context directly.
- **No tool execution.** Supplied logs *contain* tool events; the student
  renders them. The tool loop is Chapter 3.
- **No concurrency, no mailbox.** Chapter 4.
- **`chat` stays** — Chapter 1's loop, observably unchanged. `ch1parity` is 25
  points and guards it.

## §2.4a is the spec; the rest of the chapter is commentary

The types section is the contract. It has moved a great deal since you last saw
anything, largely from Bill's corrections, and several changes are load-bearing
for the grader:

1. **`Provenance{Vendor, Model, Surface}`** — not a vendor string. Thinking
   signatures are bound to the **model**, not the vendor. `Vendor` and `Surface`
   are enums starting at `iota + 1` so the zero value is invalid; `Model` is a
   string because the set is open.
2. **`seam-parse` does NOT mean byte-identical contexts.** It means identical
   **apart from `Provenance`**, which must be *preserved*, not normalized away.
   A submission whose three contexts are fully identical has thrown provenance
   away and fails. This corrected a check that would otherwise have been wrong
   in your grader.
3. **`Usage` is four disjoint fields** — `Input`, `CacheWrite`, `CacheRead`,
   `Output` — summing to the billable total. Vendors disagree about whether
   their own categories overlap (subset vs disjoint). The parser must convert.
   This is a graded property.
4. **`RedactData` is a span plus a level**, not a target plus a flag.
   Chapter 2 exercises only `RedactResult`.
5. **No `Redacted` map, and no field in the context may grow without bound.**
   Redaction stubs are *synthesized* by the reducer from the event superseded —
   deterministic, so replay stays stable.

## Checks (sum = 100)

`session` 0 · `ch1parity` 25 · `logdump` 5 · `replay` 10 · `redaction` 10 ·
`ephemera` 10 · `usage` 10 · `seam-render` 15 · `seam-parse` 15

**All eight open questions are now ruled** except question 7 (placement of the
context-engineering chapter), which does not affect this work. The table above
is final — `usage` was raised from 5 to 10 on 2026-09-12, taking 5 from
`seam-parse`, because it means normalizing four token categories across three
vendors that disagree about whether their own categories overlap. Total seam
weight is unchanged: the parse side is 25 against the render side's 15.

## Grader requirements

- **Serve fake endpoints for all three vendors.** A student with one API key —
  or none — must be able to score 100. **RULED 2026-09-12: no live testing is
  required to score 100; the fakes are the arbiter.** A cheap proxy will be
  offered for readers who want to run live without signing up for three vendor
  accounts, with BYO key identical. Nobody pays three subscriptions.
- **Never grade on Go identifiers.** You already normalize event type names
  case- and punctuation-insensitively; extend the same discipline to field
  names. Failing someone for `Kind` instead of `Type` is not a lesson.
- **Keep the practices that worked**: record-then-judge, malformed requests
  still get 200, deterministic token counts, the request census in `session`,
  and mutation tests asserting the *exact set* of failing check ids.

## Please verify before print — all of it

Chapter 2 is now almost entirely wire-format claims, which is exactly the
category that got falsified last time. Treat every one as unverified:

1. **The three request shapes** — where the system prompt goes (Anthropic
   top-level `system`, OpenAI system/developer message, Gemini
   `systemInstruction`), and the tool-result authorship in Exhibit A
   (Anthropic `tool_result` in a **user** message, OpenAI **tool** role, Gemini
   `functionResponse`).
2. **The three response shapes** and the stop-signal field names.
3. **Usage conventions — highest value.** Which vendors report cached tokens as
   a *subset* of the prompt total and which as *disjoint*? Get this wrong and
   the chapter teaches a silent costing bug while claiming to prevent one.
4. **Cache pricing ratios.** The chapter claims cache reads cost roughly an
   order of magnitude less than plain input and cache writes cost *more*. That
   asymmetry carries a load-bearing argument about prefix ordering.
5. **Gemini surfaces.** The claim that Interactions signs thought steps and
   built-in tool steps but never standard function calls, and that legacy
   `generateContent` 400s on a `functionCall` replayed without its signature.
6. **The model/vendor thinking-replay asymmetry**: Gemini errors, Anthropic
   silently drops. This is Bill's direct experience and is used to make a point
   about loud versus silent failure.

**State the verification date in the text.** Wire formats drift and this
chapter is nothing but wire formats.

## Questions worth raising back

If building it argues the spec is wrong, say so — that is what happened with
question 10 in Draft 3, and the reversal improved the book. In particular:

- Is three vendors too much for one exercise? **RULED 2026-09-12: three, all
  graded, in the order Anthropic → OpenAI → Gemini.** The order is
  load-bearing, not stylistic: the alien vendor goes **last** so that "the
  third was nearly free" cannot be true merely because the third was the easy
  one. Retreat position if it proves genuinely too heavy once built — grade
  two, ship Gemini as an ungraded exercise with the prediction attached — but
  retreat on evidence, not in advance.
- Does the "third renderer is nearly free" prediction actually hold when you
  build it? The chapter states it out loud as a falsifiable claim about its own
  design. **If it does not hold, that is a finding, not an embarrassment** —
  and it means the seam in §2.4a is wrong. You are the first person to run this
  experiment; please report the actual cost of each of the three, because the
  chapter tells the reader to measure it and we should know the answer first.
