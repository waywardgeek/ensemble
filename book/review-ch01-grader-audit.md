# Ch1 grader audit — by deletion (course-policy P9)

**Role:** coder. **Method:** P9's retroactive clause — delete each behaviour the
chapter promises from a COPY of the reference solution, score it, and record the
failing check set. **A row reading 100 → 100 is the finding.**

Rig: `scripts/grader-audit/ch01_promises.py` (promise deletions) and
`ch01_reachability.py` (reachability probes), runnable from anywhere in the repo.
Every mutation asserts its anchor matched exactly once; a silently
unapplied mutation would score 100 and manufacture a fake finding, which is the
failure mode this audit exists to catch.

---

## Headline

**The chapter's loudest warning was graded by nothing.**

§1.2 line 116 says the response content is "always a list of typed blocks. Walk
it and concatenate the `text` blocks", and calls reading `content[0].text` *the
classic day-one stumble* — the asymmetry that catches everyone once. Deleting the
walk from the reference solution and reading `content[0].text` scored **100/100**.

The cause is the one ch2 already taught us: **the fixture could not exercise the
property.** `fake.go` served exactly one block:

```go
"content": []map[string]string{{"type": "text", "text": reply}},
```

With one block in the list, walking it and indexing it are the same program. The
chapter warns loudest about the mistake the grader could not see.

---

## Table 1 — promises deleted from the reference solution

`before` is the score prior to this session's fixes; `after` is current.

| # | Promise (chapter) | Deletion | before | after | failing now |
|---|---|---|---|---|---|
| 1 | L116 walk + concatenate text blocks | read `content[0].text` | **100** | 65 | memory, replies |
| 2 | L116 concatenate the **`text`** blocks | drop the `type == "text"` filter | **100** | **100** | — *(open, see F3)* |
| 3 | L296 read all three env vars, hardcode none | hardcode `ANTHROPIC_API_KEY` | **100** | 85 | wire |
| 4 | L296 (same) — `ANTHROPIC_MODEL` | hardcode the model | **100** | **100** | — *(open, documented at L305)* |
| 5 | L296 (same) — `ANTHROPIC_BASE_URL` | hardcode a URL | 0 | 0 | all seven |
| 6 | the model's replies are appended to the conversation | drop the assistant append | 60 | 60 | memory, wire |
| 7 | cumulative usage from the first request | accumulate last-only | 90 | 90 | usage |
| 8 | on EOF print the usage line | suppress it | 90 | 90 | usage |
| 9 | send `anthropic-version` | drop the header | 85 | 85 | wire |
| 10 | send `content-type: application/json` | drop the header | 85 | 85 | wire |
| 11 | send `x-api-key` | drop the header | 85 | 85 | wire |
| 12 | `max_tokens` is required | send 0 | 85 | 85 | wire |
| 13 | stdout carries the protocol only | add a `Println` | 85 | 85 | protocol |
| 14 | after the usage line, exit 0 | exit 3 | 85 | 85 | protocol |
| — | control: unmodified reference | none | 100 | 100 | — |

Row 5 is worth a note: hardcoding the base URL is **self-enforcing**. The program
never reaches the fake, so every check fails at once. That promise needs no rule.

---

## Table 2 — reachability probes

Table 1 asks "is each promise graded?". This asks P9's two vacuity questions: can
each rule inside a check actually FIRE, and does any assertion pass vacuously?
`wire` is a bundle of rules living in the fake, so each needed its own probe.

| Probe | score | failing | verdict |
|---|---|---|---|
| `system` field dropped | **100** | — | **open, see F4** |
| `stream: true` | 85 | wire | rule reachable |
| empty model | 85 | wire | rule reachable |
| empty message content | 60 | memory, wire | rule reachable |
| first message not `user` | 85 | wire | rule reachable |
| last message not `user` | 45 | growth, memory, wire | rule reachable |
| sliding window (last 2 messages) | 45 | growth, memory, wire | `memory` does its job |
| answer mangled (uppercased) | 90 | replies | rule reachable |
| two API calls per round | 80 | calls, replies | rule reachable |

**No rule inside `wire` is decoration** — every one of them kills a mutant. All
seven checks are killed both by a purpose-built mutant student *and* by a deletion
from the reference solution.

**Negative controls.** The two checks that assert *absence* have both directions:
`protocol` asserts no extra stdout lines (control passes clean; `stdout-noise` and
the pre-existing `chatty` mutant fire), and `memory` asserts the planted fact is
*present* (`history-window-2` and `content-empty` fire).

---

## What I changed

No check's `Points` value was touched and no points were re-divided — per the
brief, that is the author's call.

1. **`internal/fakeanthropic/fake.go` — `splitIntoBlocks`.** The fake now serves
   each reply as **two** text blocks that concatenate to exactly the original
   string. Expected answers and the derived output-token counts are unchanged;
   the walk is now load-bearing. This closes F1 through the **existing** `replies`
   check — no new check, no new points.

2. **`internal/fakeanthropic/fake.go` — `ExpectedAPIKey`.** `x-api-key` was
   checked for non-emptiness only. It now must equal the value the harness put in
   `ANTHROPIC_API_KEY`. `harness.go` sets the env from the same exported constant,
   so there is one source of truth.

3. **`testdata/students/mutant/main.go` + `grader_test.go`** — two new permanent
   mutants, `firstblock` and `hardkey`, asserting the exact failing sets
   `{memory, replies}` and `{wire}`. Suite is now 12 mutants, all detected.

Gates: `gofmt` clean, `go vet` clean, `go test ./internal/grade/` passes, ch1
reference **100/100**, ch2 reference **100/100** (the fake is shared — checked for
regression).

Incidental finding: `firstblock` fails `memory` as well as `replies`, because a
truncated assistant turn can lose the planted fact. Two checks catch it; only one
was designed to.

---

## Open — author decisions (F3, F4, and the model)

I did not close these. Each needs an outline change or a ruling, and the brief
says only require what the chapter already promises.

**F3 — the type filter may not be closable with an honest fixture.**
Dropping `if block.Type == "text"` still scores 100. To catch it the fixture must
serve a non-text block **carrying a `text` field** — and no real Anthropic block
does that. A `thinking` block has a `thinking` field, `tool_use` has `input`; a
student who concatenates `.text` from every block gets byte-identical output on
real traffic. So the filter is good defensive practice that is **unobservable by
construction**. Options: (a) accept it and say so in §1.2 as a deliberate gap, or
(b) drop "the `text` blocks" to advice rather than a graded instruction. I lean
(a) — and note that inventing a fake block type to make it gradeable would be a
fixture that teaches something untrue about the wire.

**F4 — `system` is promised in the anatomy but not in the graded contract.**
L107 lists `system` as part of the request; deleting it scores 100. But the wire
contract at L330-340 — the list the grader actually implements — does not mention
it. So this may be deliberate. If you want it graded it is a one-line rule in the
fake; if not, it deserves a sentence saying so.

**The model hardcode (row 4) is documented, and the documentation is now
asymmetric.** L305 already admits "a student who hardcodes the model passes the
fake (which accepts any non-empty model)". That was *also* true of the API key
until this session, and is no longer. The harness sets `ANTHROPIC_MODEL`, so the
same one-line equality would close it — but that contradicts L305, so it is your
call. Either close it and edit L305, or keep it and narrow L305 to say the model
is the *only* one of the three that is not enforced.

---

## Correction to the record

An earlier memory of mine claimed ch1's grader "has never been audited by
deletion". That was wrong, and I checked before relying on it: ch1 already
shipped 10 mutants with exact-failing-set assertions, every check id killed by at
least one. The real gap was narrower and is the one P9's retroactive clause now
names: **check-level coverage is not property-level coverage.** Ch1's suite proved
the grader detects ten specific breakages. It did not prove every behaviour the
reference solution implements is *required* — and F1 sat inside `replies`, a check
that already had mutants.
