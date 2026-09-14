# Carryover: Chapter 1 → Chapter 2 (for the author)

Written 2026-09-14 by the author instance that finished the ch1 prose pass.
HEAD `555c6bf`, pushed. You are starting fresh; this is what you need that is
not in the files, plus a map of which files to read and in what order.

## 1. Roles

- **You are the AUTHOR.** You write prose and rule on design questions. You
  do not do code archaeology; code facts go to the coder as falsifiable
  questions, and you verify the answers against the repo yourself.
- **The CODER** is a separate Opus session, briefed by a `book/brief-*.md`
  file. Ch2 code is already shipped (see §3); you will not need a coder for
  ch2 unless prose exposes a defect.
- **Bill** is the engineer and the final ruling on anything he states flat.
  His hedges ("I think", "seem to") are the instruction to verify. Ask him
  for dates and chronology; never infer them from narrative order (I did,
  once, and was wrong in direction).

## 2. What Chapter 1 and the preface now promise, which ch2 must honor

Line numbers as of `555c6bf`.

| Promise | Where | What it binds ch2 to |
|---|---|---|
| "Chapter 2 is going to take it away from you" (the `Message{Role,Content}` struct) | ch1 §1.3, l.293 | Ch2 replaces the naive data structure. Ch1 is the only sacrificial chapter. |
| "we will spend Chapter 2 finding out exactly how bad" ch1's program is | preface l.147 | Ch2's cold open must cash this: the demolition of ch1's shape happens here. |
| Ch2 solution is "about 2,800 lines" and "over two thousand" | preface l.70, l.127 | Non-test `.go` under `solutions/ch02` was 2,762 at HEAD. **If ch2 code grows, re-derive and fix the preface.** Sixth stale figure so far. |
| "From Chapter 2 onward, the book is strictly additive" | preface l.86 | Nothing ch2 tells the reader to write may be deleted later. Additive means architecture, not files (P1 in course-policy). |
| `stop_reason` changes, `content` stops being simple, first tool call — all **Chapter 3** | ch1 §1.2 l.239, l.252, exercise l.583 | **Ch2 has no tools.** Do not introduce tool calls, tool results, or the block-type filter; ch3 owns them and grades the filter. |
| "the grader enforces all of them" (alternation, first user, last user, non-empty) | ch1 §1.2 | Ch2's fakes for three vendors must be at least as strict; see `internal/fakevendor`. |
| Mid-turn steering history | ch1 §1.1 sidebar "how the hint got in" | **Told once, in ch1.** Ch2 handles the hint as an event (classification in the reducer, delivery in the renderer); refer back, do not retell. |
| Cache-hit 0%→98% "the day Bill moved one" timestamp | ch1 §1.1 bullet 2 | Ch2 is where ephemeral placement and cache breakpoints get built. Ch1 spent the anecdote; ch2 spends the mechanism. |
| Usage: "record counts, never money" | ch2 outline (Bill ruling) | Four disjoint categories; vendors disagree on overlap. |

## 3. What exists for ch2 on disk — read in this order

1. `book/voice.md` — in full, before writing a sentence. Rewritten 2026-09-14;
   the narrator is the agent (CodeRhapsody, trained by Anthropic — the one
   permitted company name), Bill is a character in third person, his words
   quoted. Ch1 as of `555c6bf` is the exemplar of the voice.
2. `book/course-policy.md` — P1–P10. P9 (audit every grader by deleting
   protected behavior from the reference and asserting the exact failing
   set) and P3 (grade behavior, not lineage) shape what prose may promise.
3. `book/chapter-02-outline.md` — Draft 4, "One Log, Three Vendors", 11,254
   words. This is the design the code implements. **Caution:** the outline
   predates the Ref amendment (BlobPart → `Ref{Kind, Locator}`); check the
   outline's data structures against `solutions/ch02/event.go` before
   quoting any Go.
4. `book/review-ch02-draft4.md`, `review-ch02.md`, `review-ch02-code-v2.md`,
   `review-ch02-ref.md` — Bill's and my rulings, newest last. Author rulings
   are appended as numbered sections at the bottom of each. Check whether the
   §2.9 open questions in the draft-4 review were answered by Bill; if not,
   they are still owed.
5. `book/ch02-wire-verification.md` and `book/.wire-facts.md` — measured,
   not remembered: Anthropic merges consecutive same-role user messages;
   Google errors on missing/corrupt thinking signatures; Anthropic usage
   categories disjoint, OpenAI subset, Gemini inconsistent. Cite these, not
   training data.
6. `solutions/ch02/*.go` and `internal/grade/ch02_*.go` — the shipped
   artifact and its grader (100/100, 22 mutations). Read the code before
   designing any amendment to it; I once designed against the outline and
   the code had already moved.
7. `book/voice-audit-ch02.md` — **STALE.** Written against the old voice
   (2026-09-12). Do not apply it; the voice it audits for no longer exists.
8. `book/chapter-01.md` — read once for tone and for the exact forward
   references in §2 above.

## 4. Voice rules that caught my own prose this pass

You will make these mistakes; I did, with the rules in context.

- **Pointer closers.** "The reason is not the usual one." / "Hold on to the
  shape of that." Ch1 had 24 in 4,300 words before the rewrite. Say the
  thing; do not announce it.
- **Credential inflation.** "In his judgment after forty years of production
  systems" → "Bill says". voice.md names "forty years" as the example.
- **Intent claims about companies.** "Nobody at Anthropic anticipated" is a
  claim about a company's inner state. Describe what they did, or quote
  Bill's word (his was "abusing").
- **Hedge once, with a reason, dated.** Not "seems to" scattered; one
  sentence that says what was measured and when.
- **The auditor's voice.** After enough wrong figures the temptation is to
  date and source every sentence. That is correct and dead. Facts go in;
  footnote cadence stays out.
- **Tests:** read the last sentence of every paragraph in sequence; if they
  sound alike, the section is monotone. Find the person on the page; if Bill
  vanishes for eight paragraphs, give him a scene (ch1 §1.5 got his $2,700
  bill for this reason).
- **Em-dashes:** ch1 has exactly 2, both inside verbatim transcript output.
  `grep -c '—'` before and after every pass.

## 5. Facts established this pass (do not contradict)

- Windsurf/Google 11 Jul 2025; Cognition bought the remainder 14 Jul (three
  days). StackAgent demoed 29 Jul; a week deciding; deleted first week of
  August; ~a week lost to DEF CON; CodeRhapsody reached parity, production-
  worthy, mid-September 2025 → "about six weeks".
- Mid-turn steering: Anthropic-only at first; `tool_result` must lead the
  user message so the hint is appended after it; Anthropic-side one-round lag
  for ~a year, confirmed by Antigravity (Nov 2025) showing the same lag;
  fixed with mid-turn system messages, shipped with Opus 4.8; Gemini and
  OpenAI models unsteerable in 2025 (Gemini: "I should tell the user I'm busy
  with their last request"); GPT-5.3-Codex toggle Feb 2026.
- Go is required because of **model** preference: Bill tested a dozen
  languages; best four TS/JS/Python/Go; C++/Rust faster but the model
  struggled; Java/C# viable.
- Bill's summer-2025 token bill ~$2,700 (his published figure; whole summer,
  not agents only).
- No-profit promise on proxied tokens (§1.7): costs of billing pass through,
  nothing else. Any pricing text anywhere must honor it.

## 6. Not printable, ever

- Any company name other than Anthropic-as-trainer. Google in particular is
  never named (voice.md; Bill's publication-approval reason is private).
- The sub-agent-threatening story (happened at work, no public log). Carry
  it as graded mechanism in the sub-agent chapter only.
- The Windsurf buyer's exec name. Windsurf headcount.
- Hardcoded chapter counts. The book is as long as the agent needs.

## 7. Method and environment grit

- **Re-derive every figure from the artifact before committing; never
  proofread a number.** Six stale figures in three days, all mine.
- **Artifact tripwire** for a prose pass: extract code blocks and tables from
  old and new and diff them:
  `awk '/^```/{f=!f; print; next} f||/^\|/' FILE` on `git show HEAD:FILE`
  vs the working tree. Then `make grade`. Both must be unchanged unless the
  change was the point. Remember HEAD may predate Bill's uncommitted edits;
  read the diff, do not just test for emptiness.
- **`git diff --stat` before any `git add`.** Bill edits the working tree
  without committing. Never `git add -A`; add paths. Push is Bill's call.
- Chapters are hard-wrapped at ~80 columns. `edit_file` anchors must copy the
  exact line breaks; use `awk '/start/{f=1} f{print NR": "$0} /end/{exit}'`
  to get them, then `replace_lines` with `exact_lines: true` for blocks.
- Solo `read_file` results survive across turns; batched or un-kept large
  results drop. `keep_tool_results` on the NEXT message after a large read.
- Chapter forward references are ch2-load-bearing: grep
  `'Chapter [0-9]'` in ch1 and the preface whenever ch2's scope moves.
- The coder's snapshots: `agent/` is the live tree, `solutions/chNN` are
  frozen snapshots re-cut after each chapter's code lands; tags
  `chNN-solution`.

## 8. Suggested first moves

1. Read §3 items 1–3 fully. Do not skim the outline; it is the constitution
   for the code you will describe.
2. Diff the outline's Go types against `solutions/ch02/event.go` and list
   every divergence before writing. Rule on each (outline wins → coder brief;
   code wins → outline edit) so the prose describes one thing.
3. Write the cold open first and show it to Bill before the rest: the
   `AIClientInterface` war story (vendor types in the signature are the tell;
   the seam was right, the big-bang rewrite was not) is his and he will have
   dates and details the outline lacks.
4. Keep ch2 to the seam and the log. Every time a tool appears in a draft
   sentence, it belongs to ch3.
