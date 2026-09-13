# Brief — ch3 follow-up: declare the tools, split `localtools`

**From:** the author. **Date:** 2026-09-13.
**Spec:** `book/chapter-03-outline.md` (amended at 81cb4ce).
**Context:** your review at `book/review-ch03-code.md`. I accepted essentially
all of it. This brief covers the two items that create code work.

Your review was good. The `go run` exit-code finding and the results-first
vacuity finding were both things I had asserted confidently and graded with
nothing, and you measured rather than assumed. Keep doing that here, including
to this brief — it is written by someone who has not read your code, so where
it prescribes a mechanism and the mechanism is wrong, implement the property
and tell me.

---

## 1. The agent must declare its tools (your Enrichment item 1)

I ruled this **in scope for chapter 3**, against your suggestion that it might
be deferred. Reasoning, so you can argue with it: chapter 3's payoff sentence is
"your agent can now write code." If the agent never declares its tools, a real
vendor never sends a `tool_use` block, so that sentence is false everywhere
except against our fake — which volunteers tool calls unprompted. A student
would score 100 and have a dead agent. A book that grades a green dashboard
while the artifact does nothing has no standing to teach anything else.

**Properties, not mechanisms:**

1. A request carries the tool declarations — name, description, argument schema
   — for whichever vendor is being rendered. All three.
2. **When the registry is empty the field is absent**, not empty-but-present.
3. Chapter 2 registers no tools, therefore **chapter 2's request bytes are
   unchanged, byte for byte.** Verify this with an actual byte comparison
   against the pre-change output, not by observing that `-ch 2` still scores
   100. Those are different claims and only the first one is the one I need.
   If they disagree, the byte comparison is right and the score is lying.
4. The declaration is **graded, and audited by deletion**: an agent that sends
   no `tools` field must lose points. If deleting it still scores 100, the check
   is decoration — say so and fix the fixture, exactly as you did for the
   ordering log.
5. Parsing is unchanged. This is the request direction only.

Where the points come from is yours to propose. I would rather you tell me the
honest number than squeeze it into a check where it does not belong; if it
needs its own check and its own row, say so and I will amend the table and the
arithmetic.

## 2. Split `localtools` into `readtools` and `mutatetools` (your Question 2)

Accepted, for your reason — you correctly used my own stated rule against me.
`readtools` 10 (`read_file` with range, `list_directory`, `search_files`),
`mutatetools` 10 (`write_file`, `edit_file`). Sum stays 100; the outline table
is already amended and already re-verified at 100.

`list_directory` and `search_files` do **not** split out further. Nobody has
search working and read broken.

## 3. Already yours, already done — confirm only

Your review says the reference refuses **both** the zero-match and the
many-match anchor. I have written that into §3.5 as part of the declined
decision. Confirm it is true of the code as shipped; if it only refuses zero
matches, that is a third thing to fix.

---

## Constraints (unchanged, and non-negotiable)

- `-ch 1`, `-ch 2`, `-ch 3` all still 100.
- P9: audit by deletion, with a negative control for every absence assertion.
- **Assert that each mutation actually landed.** I have now been bitten by a
  silently-unapplied mutation myself — every test passed, which would have
  "proved" a suite that tested nothing.
- Points sum to exactly 100, checked against the outline table.

## The question I actually want answered

Chapter 6 is the capability seam and MCP — where tools stop being a Go map and
start arriving from somewhere else. **Does declaring tools in chapter 3 force
any structure that chapter 6 will have to tear out?**

Answer honestly, including "yes." When you told me nothing in chapter 3 forced
a chapter 4 tear-out, that was the evidence the ch3/ch4 split was cut in the
right place. This is the same question one seam further along, and a "yes" is
more useful to me now than after the prose is written.
