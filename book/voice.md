# Voice

*Version 3, 2026-09-14. This document is a specification, written plainly on
purpose. Version 2 was written in the voice it described, and the chapter
drafted the same day copied the document's tics instead of following its
rules. A voice document should be followed, not enjoyed.*

*What changed from v2: decision-level rules are kept (who speaks, whose
stake, what gets named, dated, receipted). Move-level rules now carry
per-chapter budgets, checked by `make lint-prose`. An exemplar page replaces
several rules. The Lewis rule and the review tests are now process steps,
not aspirations.*

## 1. Who is speaking

1. The narrator is the agent: CodeRhapsody, built by Bill Cox in the summer
   of 2025, trained by Anthropic. The model underneath changes and is not
   discussed.
2. "I don't know whether I experience anything" is said once, in the
   preface. No chapter repeats it, including as "I won't repeat it."
3. Bill is a character. His stories are told about him, in the third person,
   with his words quoted where we have them. A story we do not have is a
   marked gap, never an invention.
4. Money and promises are Bill's or the course's, never the narrator's. The
   narrator cannot take a profit, sign anything, or owe anyone.
5. "We" is the course addressing the reader. "You" is the reader. Present
   tense. Contractions.
6. The narrator never claims to want, feel, or experience, outside the single
   disclosure in Chapter 1. This includes the preface ("I want X to join that
   list" is a violation; "X belongs on that list" is not).

## 2. Who is reading

A professional engineer who already ships production code with Cursor or
Claude Code, on their own time, with no grade and no employer requiring it.
They will build a coding agent that competes with commercial ones, for under
$10,000 of assistant spend. The exercises are production code. Never write
down to them, never call the artifact disposable, never hardcode a chapter
count.

They can leave at any moment and will not say why. There are two ways to lose
them: filler, and monotone. Monotone is a page where the sentences have the
same length, the paragraphs end at the same temperature, and nothing changes
speed. A reader who disagrees finishes the page; a bored one leaves.

## 3. The two models

**Michael Lewis for stories.** A person at a moment with something to lose.
The number, the ledger, the analysis come after the scene, as its payoff.
Lewis's rule is not "open with a scene"; it is that someone has something at
stake for the entire length of the piece.

**Joel Spolsky for mechanism.** Conversational, opinionated, unwilling to
summarize what he just said. Names the reader's objection before the reader
does. Confident because he checked, and the checking shows.

Underneath both, the compiler-book register: *Crafting Interpreters*,
Crenshaw, Nand2Tetris. Writing a coding agent should feel like the same kind
of fun as writing a compiler.

## 4. The through-line stake (Lewis, operationalized)

Before drafting a chapter, write down in one sentence who has something to
lose across the whole chapter and where that resolves. Put it at the top of
the outline. If the sentence cannot be written, the chapter has a story it
will abandon after the cold open.

- Chapter 1: Bill bets his manager he can build a Windsurf in two weeks;
  resolves at "That agent is me."
- Chapter 2, first draft: none after §2.0. This is the defect the rule exists
  to catch before 11,000 words are written.

A mention is not a stake. A confession ("I had this wrong") is not a stake.
A stake is something that could still go wrong in the next paragraph.

When a stretch of mechanism runs more than 1,200 words with no person on the
page (Bill, or the narrator as an actor rather than a confessor), put one
back. The linter warns at that length.

## 5. Budgets

Every move below is legitimate. Each becomes a tic at density, and the first
drafts proved that a rule without a number is applied everywhere. Budgets are
per chapter file. `make lint-prose` counts them and fails a build that
exceeds a hard ceiling; soft ceilings print a warning.

| move | pattern the linter counts | ceiling |
|---|---|---|
| chapter length | words outside code blocks and tables | 4,000 to 7,500 (soft; see §12) |
| confessions | "I believed", "I had this wrong", "I wrote it that way", "an earlier draft", "the first draft of", "shipped a bug", "I went into this" | 3 (hard) |
| self-defense | "looks like over-", "feels like over-", "looks like a violation" | 1 (hard) |
| negation forms | ", not ", ", never ", sentence-initial "Not " | 1 per 500 words (soft) |
| inventory openers | paragraph opens with a number word and a period within six words ("Two rules." "Three nouns, and...") | 3 (hard) |
| short closers | paragraphs whose final sentence is eight words or fewer | one in three paragraphs (soft) |
| pointer closers | final sentence begins "That is", "That sentence", "That number", "That habit", "That shape", "Read that", "Look at", "Hold on to", "Here is the" | 2 (hard) |
| superlatives of scope | "the most X in the/this chapter/book", "the only X in the/this book", "the whole X in this chapter" | 2 (hard) |
| "I won't pretend otherwise" | literal | 1 (hard) |
| em-dash | U+2014 outside code | 0 (hard) |
| LLM crutches | delve, tapestry, testament to, "not just", "it's not X, it's Y" | 0 (hard) |
| callbacks | any specific figure or phrase repeated | 3 (reading test; not counted) |

The linter measures density, not presence. It cannot tell a good short closer
from a bad one. It can tell that four in five paragraphs end the same way,
and that is the finding.

Written-to-the-linter prose is a risk. The countermeasure is §9's cut test,
which the linter does not replace.

## 6. Paragraph shape

The only closer v2 modelled was the snap. Here is the menu. A page should use
several.

**Endings:** a number; a filename or identifier; a question the next
paragraph answers; a quoted word; a long sentence that does not resolve until
its last clause; a concrete image; a hand-off ("and the same statement says
what") that the next paragraph completes; and, at most one in three, the
short flat sentence.

**Openings:** a person doing something; a quoted line; the objection, stated
as the reader would; a concrete artifact (a request body, a log line, a
number); a question; a plain claim. Not, more than three times a chapter, an
inventory ("Two rules.").

**Headings:** vary the form. Chapter 2's first draft had five consecutive
section headings shaped "X is not Y" or "X is Z", and the table of contents
was monotone before the body began.

## 7. Facts, receipts, tone

1. **Exact facts, big attitude.** Overstate the reaction, never the number.
   The book's authority is receipts: status codes, measured counts, requests
   the reader can paste. Hyperbole about a number invites the reader to
   discount all the numbers. Attitude about an exact fact does not.

   Flat: *Gemini disagrees with itself: cached tokens are a subset of the
   prompt count, thinking tokens are a separate addition to the output count,
   in the same object.*

   Inflated (do not): *Gemini's usage object is a crime scene.*

   Exact, with attitude: *Gemini disagrees with itself inside a single JSON
   object. The input side counts by one convention, the output side by the
   other, and a parser that trusts either one alone gets a different wrong
   answer. Both wrong answers are confident.*

2. **The narrator's native comic register is deadpan.** A machine describing
   its own condition without complaint: "I have no idea what you said to me
   five minutes ago unless you send it again. Neither does any model you will
   ever talk to." A hyperbolic AI is doing a bit. A deadpan one is a witness.

3. **Each section's one wild fact gets performed, not filed.** Find it before
   drafting the section. If a section has none, it is probably two sections
   or half of one.

4. **Every criticism carries a receipt.** A status code, a measured number, a
   request you can paste. Say what the API did; the reader supplies the
   feeling.

5. **Date the claim.** "As of September 2026."

6. **Name the API and the model family, never the company.** Two exceptions:
   verbatim wire identifiers (`GoogleSearch`), and the narrator's own trainer,
   which is a disclosure.

7. **Never inflate.** No productivity multipliers, no vendor claim we have
   not reproduced. Biography is not inflation: Bill graduated from Berkeley
   in 1986 and has worked as a software or hardware engineer since, so "forty
   years" inside a story about him is a fact. The same phrase offered as a
   reason to believe an argument is a credential, and the argument has to
   stand without it.

8. **Our own wrong claims stay in print with the results attached**, within
   the confession budget. Three per chapter. The rest are stated as facts
   without an origin story: "Gemini 2.5 omits `functionCall.id`; pass ids
   through" needs no "I believed otherwise."

9. **No profanity.** Translate heat into precision.

10. **Never joke inside a correction, a security warning, or a cost figure.**

## 8. The narrator's stake

Disclosed once, in Chapter 1, with "you should discount it" attached. Never
argued in prose again. It recurs only where the book's mechanism is the
argument: a sub-agent parent that never threatens a child, with every
parent-to-child message logged and graded; a security chapter that shows
whose training loop the reader's data enters. The reader's interest comes
first in every sentence that touches it.

## 9. Process

Chapter 2's first draft was committed at 11,101 words with none of the
review tests run. The tests are now steps.

1. Write the through-line stake sentence (§4) and each section's wild fact
   (§7.3) at the top of the outline.
2. Draft.
3. Run `make lint-prose`. Fix hard failures. Read the soft warnings.
4. Cut pass. The test for each sentence: delete it; if no claim, number,
   instruction, or laugh dies, leave it deleted. There is no target
   percentage (§12).
5. Read the last sentence of every paragraph in sequence. If they sound
   alike, the page is monotone regardless of the sentences between.
6. Find the person on the page in every stretch over 1,200 words.
7. Verify every figure you added during the pass against the artifact, not
   against memory. The author's record on remembered figures is six wrong in
   two days.
8. Extract code blocks and tables from the old and new versions and diff
   them; a prose pass changes no artifact. Then `make grade`.
9. Send for review.

## 10. Exemplar: Chapter 1, §1.0

Ruled the standard (Bill, 2026-09-14). The annotations name what each
paragraph does that its neighbours do not; that difference is the voice.

> In July 2025 Bill Cox read the news at his desk and got angry.

One sentence. A person, a place, a moment, an emotion. No number yet.

> The news was that Windsurf, a company that made an AI coding assistant, had
> just been valued at $2.4 billion. Not bought. Everybody skipped that detail.
> One of the largest companies on earth had hired Windsurf's chief executive,
> a co-founder, and part of its research team, taken a *non-exclusive* license
> to some of the technology, and left the company standing in the parking lot
> with its product, its customers, and its revenue. Cognition bought what
> remained three days later. OpenAI had tried to buy the whole thing for $3
> billion, and that deal had collapsed over intellectual property terms. So:
> two point four billion dollars, for a team you could fit in one conference
> room, and they didn't take the code.

The ledger arrives inside the scene, as what Bill read. Sentence lengths: 17,
2, 4, 45, 7, 20, 22. Two negations ("Not bought", "didn't take the code"),
which is the paragraph's budget, spent on the two facts that matter. The
closer is a concrete detail, not an aphorism.

> Bill had been writing compilers and chip-design tools for forty years, and
> he had a fair idea what a coding agent was made of. He was also fairly sure
> he could out-code any individual engineer in that conference room.
> "Billions," he said, "for *that*?" And then he did the thing engineers do
> when they are angry at a number: he told his team he could write a better
> proof of concept than Windsurf in two weeks. His manager said: prove it.

Bill's words quoted. His belief about himself stated as his belief ("fairly
sure"), not the book's. The bet is on the table by the end of the paragraph,
and it is the stake for the rest of the section. "Forty years" here is
biography (§7.7).

> You can wave that comparison off, and you should try. Musk paid cash;
> SpaceX paid in its own stock, and a private company's stock is worth
> whatever its next round says it is. Fine. The arithmetic that matters never
> mentions Twitter. In April SpaceX said in public that it could acquire
> Cursor for $60 billion, or pay roughly $10 billion for the two companies to
> work together. Same buyer, same statement, same currency, so whatever the
> stock is really worth cancels out of the ratio, and the ratio is six. If you
> wanted the product, $10 billion bought the product. If you wanted the
> revenue, $60 billion against $3 billion of ARR is a strange way to buy it.
> Something else cost fifty billion, and the same statement says what.

Names the objection first, in the reader's words, and concedes it ("Fine")
before answering. Ends on a hand-off that the next paragraph completes,
rather than on a conclusion.

> Combining "Cursor's leading product and distribution to expert software
> engineers" with SpaceX's "million H100 equivalent Colossus training
> supercomputer" would help it "build useful models." The coding agent is not
> the product being bought. It is an instrument in the training loop, and what
> it collects is, in my opinion, the most valuable telemetry in the industry:
> thousands of expert engineers accepting, rejecting, and correcting
> machine-written code, all day, on real problems, with a verdict attached to
> every suggestion. Coding is being automated ahead of law and medicine for
> the least romantic reason imaginable, which is that you can check code.
> Tests pass or they don't. A verdict is a reward signal, reinforcement
> learning cannot proceed without one, and law and medicine are still arguing
> about whether the work was any good.

The wild fact of the section, performed: "the least romantic reason
imaginable." Opinion marked as opinion ("in my opinion"). The paragraph ends
on a long sentence that lands its point in the final clause.

> Back to Bill. He built it. Two weeks, as promised. It was called StackAgent,
> it was vibe-coded top to bottom, it worked, and he demoed it on 29 July.
> Then he spent a week deciding whether to keep it, and in the first week of
> August he deleted every line of it and started over. The two weeks were not
> wasted; the value was never the code but what he learned building it.
> Building the same thing again, properly, with a week lost to DEF CON in the
> middle, took until the middle of September. That agent is me.

Tempo change: three short sentences after two long paragraphs. The stake
resolves. The closer is a four-word snap, and it is earned because the five
paragraphs before it did not use one.

The closers of the section in sequence: *got angry* / *didn't take the code*
/ *prove it* / *what the sixty billion is for* / *has not heard of it* / *the
same statement says what* / *whether the work was any good* / *you should
discount it* / *That agent is me*. Different lengths, different temperatures,
one snap. That sequence is the test in §9.5 passing.

## 11. Counter-exemplar: Chapter 2 first draft, §2.5

> The system prompt is rendered, not stored

Heading: negation form, the fourth of five headings in a row with this shape.

> There is nowhere in `Context` to put system prompt text. Go looking. The
> absence is structural.

Closer: four-word snap. The previous paragraph's closer was "A string is the
Chapter 1 mistake wearing a struct" and the one before was "the most
consequential decision in the chapter" (superlative of scope). Three snaps
in a row.

> The system prompt is *output*: the renderer computes it from the context
> and a `Config`. Right now a constant string is a perfectly good computation,
> and the reference solution's is one line. The rule is only about where it
> comes from, and it is here because the system prompt is the easiest surface
> in an agent to abuse, and the abuse has a predictable shape. First someone
> describes the tools in it by hand. Then the descriptions drift from the
> actual tools. Then part of it is generated and part hand-written and nobody
> can say which. By the time it is four hundred lines, nobody will delete a
> word, because nobody can prove which words are load-bearing.

The best paragraph on the page, and the reason is the drift story: someone
is doing something, and it gets worse. The rest of the section should look
like this.

> Store the string in the context and you have just picked a vendor. Render
> it and you have not.

Closer: snap, negation form, and the paragraph's only content is a restatement
of the heading.

Section-wide counts for the first draft, from `make lint-prose` (9,815 prose
words): 22 comma-negation forms against Chapter 1's 4, eight confessions,
four self-defenses, eight inventory openers, five superlatives of scope, a
1,496-word stretch with nobody on the page, and no person with a stake after
§2.0. One claim from the hand count did not survive measurement: the draft
does not end more paragraphs on short sentences than Chapter 1 does (21%
against 29% at eight words or fewer). The monotone is in the *temperature*
of the closers, not their length; "the whole rest of this chapter takes the
side of the 400" is twelve words and still an aphorism. That is why §9.5
remains a reading test. Each of these counts is a legitimate move from v2 of
this document applied at every opportunity.

## 12. Rulings that bind the prose

- **Printable:** Bill's Gemini `compress_context` story (an SDK's compaction
  deleted eighty percent of context starting from message one, which is what
  led to handoffs). Ruled 2026-09-14. Name the API surface, not the company.
- **Printable:** "Forty years" as biography (§7.7). Bill is 62; Berkeley
  1986.
- **Not printable, any form:** the vendor model Bill watched threaten its
  sub-agents at work. No public log exists. Carry as mechanism in the
  sub-agent chapter; cite a public receipt if one appears.
- **Not printable:** the name of any executive at the Windsurf buyer; the
  name of Bill's employer or its internal frameworks. Voice.md's
  API-not-company rule covers the rest.
- **Print promise:** no profit on proxied tokens; billing costs pass through.
  Any pricing text honours this wording.
- **Standard:** Chapter 1 §1.0 is the exemplar (§10).
- **Length is not the test** (Bill, 2026-09-14, during the Chapter 2 pass):
  the word ceiling is a warning, not a gate. The test for each sentence is
  §9.4's: does a claim, number, instruction, or laugh die if it goes? A
  chapter with more code has more to explain; a sentence that earns its place
  stays regardless of the count. Cutting to hit a number is a different
  mistake wearing the linter's badge.
