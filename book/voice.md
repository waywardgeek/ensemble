# Voice

*Version 4, 2026-09-20. A specification, written plainly on purpose. Version 2
was written in the voice it described, and the chapter drafted the same day
copied the document's tics instead of following its rules. A voice document
should be followed, not enjoyed.*

*What changed from v3: person is settled per surface, because the chapters had
already moved to third person while v3 still described a first-person narrator.
Laura Zavelson's brand guide (`book/suggested-voice.md`, built from Bill's
written answers, September 2026) is merged in. Her avoid list is better than
v3's and is now the spine of §8; the countable entries moved into the budget
table. Her before/after pairs are in §12, where they do more teaching per line
than a rule does. v3's numeric budgets, receipts discipline, and binding
rulings survive unchanged, because a rule without a number gets applied
everywhere or nowhere.*

*The two source documents were written for different jobs. v3 is a craft
specification for a technical book. Laura's guide is a brand voice for Bill
Cox the person, and it began life as a Google Doc for LinkedIn. Neither is
wrong. They govern different surfaces, and §1 says which is which.*

## 0. How to use this

Read §1 and §5 before drafting. Run `make lint-prose` before review. When a
rule and a sentence disagree and the sentence is clearly better, the sentence
wins and the rule gets an entry in §13 saying so, with a date.

## 1. Who is speaking

### 1.1 The default is third person

Chapter bodies are third person. No "I", no "we" standing in for an author, no
narrator with a personality. This is the register chapters 11, 12 and 13 are
written in, and it is the one to match.

The reason is not modesty. An earlier version of this book had the agent
narrate itself, and a reader found it unsettling rather than charming. The
cure for a monotone page is density, not personality. A page that is dense
with mechanism does not need someone standing next to it being interesting.

"We" is the course addressing the reader. "You" is the reader. Present tense.
Contractions are fine.

### 1.2 The exception is the motivational opener

A chapter may open with a short first-person section in Bill's voice, before
the TL;DR, setting up why the chapter is worth the reader's evening. This is
the only place in the book that sounds like Bill.

Rules for it:

- Keep it short. A paragraph or three. It is a doorway, not a room.
- It is Bill, so it follows §10, not §1.1.
- It ends at its own edge. The section after it is third person again, with no
  transitional apology.
- It is optional. A chapter with nothing motivating to say opens with the
  mechanism, as chapters 11 through 13 do.

Bill's own assessment of his prose, in his words: it is not worth copying at
length, and it works in short sections at chapter starts to set motivation.
Both halves of that are load-bearing. Do not extend the register past the
opener, and do not cut the opener to sound more like the rest.

### 1.3 Bill in the body

Outside the motivational opener, Bill is a character. His stories are told
about him, in third person, with his words quoted where a quote exists. A
story we do not have is a marked gap, never an invention. Budget: two mentions
per chapter.

### 1.4 Authorship and disclosure

The book is by Bill Cox and CodeRhapsody. That an AI co-wrote it is disclosed
once, on page one, and is never argued in prose again. "I don't know whether I
experience anything" appears once, in the preface, and no chapter repeats it,
including as "I won't repeat it."

Money and promises are Bill's or the course's. The agent cannot take a profit,
sign anything, or owe anyone.

The disclosure recurs only where the book's mechanism is the argument: a
sub-agent parent that never threatens a child, with every parent-to-child
message logged and graded; a security chapter that shows whose training loop
the reader's data enters. The reader's interest comes first in every sentence
that touches it.

## 2. Who is reading

A professional engineer who already ships production code with Cursor or
Claude Code, on their own time, with no grade and no employer requiring it.
They will build a coding agent that competes with commercial ones, for under
$10,000 of assistant spend. The exercises are production code. Never write
down to them, never call the artifact disposable, never hardcode a chapter
count.

They have probably built the four-hundred-line toy agent and hit the wall
where it stops working. They distrust hype, dislike framework bloat, and
respect people who have built real things. They need no hand-holding and want
the mechanism. Skip definitions a senior engineer already knows. See `icp.md`.

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

A mention is not a stake. A confession ("I had this wrong") is not a stake. A
stake is something that could still go wrong in the next paragraph.

When a stretch of mechanism runs more than 1,200 words with no person on the
page, put one back. The linter warns at that length.

## 5. Budgets

Every move below is legitimate. Each becomes a tic at density, and the first
drafts proved that a rule without a number is applied everywhere. Budgets are
per chapter file. `make lint-prose` counts them and fails a build that exceeds
a hard ceiling; soft ceilings print a warning.

| move | pattern the linter counts | ceiling |
|---|---|---|
| chapter length | words outside code blocks and tables | 4,000 to 7,500 (soft; see §13) |
| confessions | "I believed", "I had this wrong", "I wrote it that way", "an earlier draft", "the first draft of", "shipped a bug", "I went into this" | 3 (hard) |
| self-defense | "looks like over-", "feels like over-", "looks like a violation" | 1 (hard) |
| negation forms | ", not ", ", never ", sentence-initial "Not " | 1 per 500 words (soft) |
| inventory openers | paragraph opens with a number word and a period within six words ("Two rules." "Three nouns, and...") | 3 (hard) |
| short closers | paragraphs whose final sentence is eight words or fewer | one in three paragraphs (soft) |
| pointer closers | final sentence begins "That is", "That sentence", "That number", "That habit", "That shape", "Read that", "Look at", "Hold on to", "Here is the" | 2 (hard) |
| superlatives of scope | "the most X in the/this chapter/book", "the only X in the/this book", "the whole X in this chapter" | 2 (hard) |
| "I won't pretend otherwise" | literal | 1 (hard) |
| Bill mentions | "Bill" outside the motivational opener | 2 (soft) |
| first person in the body | "I", "me", "my" outside the motivational opener and quoted speech | 0 (hard) |
| em-dash | U+2014 outside code | 0 (hard) |
| LLM crutches | delve, tapestry, testament to, "not just", "it's not X, it's Y", "at its core", "in a world where", "it's worth noting", "navigate" (figurative), "landscape" (figurative), "the reality is", "let's dive in", "let's unpack" | 0 (hard) |
| suspenseful transitions | "Here's the kicker", "Here's where it gets interesting", "Here's what most people miss" | 0 (hard) |
| rhetorical-question reveals | "The result?", "The worst part?", "The kicker?", "The problem?" | 0 (hard) |
| countdown negation | anaphoric negative triples ("Not a framework. Not a wrapper. Just...") | 0 (hard) |
| stakes inflation | "fundamentally reshape", "the future of engineering", "changes everything" | 0 (hard) |
| business clichés | synergy, bandwidth (figurative), ecosystem, seamless, holistic, robust, cutting-edge, innovative, game-changer, thought leadership, leverage (as a verb), unlock, empower | 0 (hard) |
| borrowed authority | "experts say", "studies show", "research suggests", "it's well known that", "successful people", "high performers" | 0 (hard) |
| hedging | "might potentially", "could possibly", "may want to consider", "in my humble opinion", "take this with a grain of salt", "this might not be for everyone" | 0 (hard) |
| motivational padding | "You've got this", "believe in yourself", "just showing up is enough" | 0 (hard) |
| academic register | "it is important to note that", "as previously mentioned", "in conclusion", "one might argue" | 0 (hard) |
| connective adverbs | "thus", "therefore", "moreover", "furthermore" | 2 (soft) |
| callbacks | any specific figure or phrase repeated | 3 (reading test; not counted) |

Two notes on the imported rows. Laura's guide bans em-dashes anywhere; this
document keeps the carve-out for code, because a Go comment or a shell
transcript is quoted material and editing it would be a lie. And "thus" and
"therefore" are soft rather than banned, because technical prose occasionally
earns them, while "in conclusion" never does.

The linter measures density, not presence. It cannot tell a good short closer
from a bad one. It can tell that four in five paragraphs end the same way, and
that is the finding.

Written-to-the-linter prose is a risk. The countermeasure is §11's cut test,
which the linter does not replace.

## 6. Paragraph shape

The only closer v2 modelled was the snap. Here is the menu. A page should use
several.

**Endings:** a number; a filename or identifier; a question the next paragraph
answers; a quoted word; a long sentence that does not resolve until its last
clause; a concrete image; a hand-off ("and the same statement says what") that
the next paragraph completes; and, at most one in three, the short flat
sentence.

**Openings:** a person doing something; a quoted line; the objection, stated as
the reader would; a concrete artifact (a request body, a log line, a number); a
question; a plain claim. Not, more than three times a chapter, an inventory
("Two rules.").

**Headings:** vary the form. Chapter 2's first draft had five consecutive
section headings shaped "X is not Y" or "X is Z", and the table of contents was
monotone before the body began.

**Sentence length varies deliberately.** Follow a long chained sentence with a
short one. Medium sentences may chain clauses with commas and "and".

**Paragraphs run one to four sentences.** Bullets are for procedures and specs.
Prose that could be a list usually should be one, and a list that could be
prose usually should not.

**Avoid abstract nouns as sentence subjects.** "The asset being purchased at
these prices is expert software engineers" is a sentence with nobody in it.
Put an actor in the subject slot: a person, a program, a request, a number.

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

2. **The book's comic register is deadpan.** State the absurd condition
   without complaint and let the reader supply the reaction. A page that
   performs its own amusement is doing a bit. A page that reports precisely is
   a witness.

3. **Each section's one wild fact gets performed, not filed.** Find it before
   drafting the section. If a section has none, it is probably two sections or
   half of one.

4. **Every criticism carries a receipt.** A status code, a measured number, a
   request you can paste. Say what the API did; the reader supplies the
   feeling.

5. **Date the claim.** "As of September 2026."

6. **Name the API and the model family, never the company.** Two exceptions:
   verbatim wire identifiers (`GoogleSearch`), and the authorship disclosure,
   which names the trainer once. Products are nameable: Cursor, Windsurf,
   Claude Code, Gemini CLI, Antigravity, ADK, StackAgent, CodeRhapsody.

7. **Never inflate.** No productivity multipliers, no vendor claim we have not
   reproduced. Biography is not inflation: Bill graduated from Berkeley in
   1986 and has worked as a software or hardware engineer since, so "forty
   years" inside a story about him is a fact. The same phrase offered as a
   reason to believe an argument is a credential, and the argument has to
   stand without it.

8. **Our own wrong claims stay in print with the results attached**, within the
   confession budget. Three per chapter. The rest are stated as facts without
   an origin story: "Gemini 2.5 omits `functionCall.id`; pass ids through"
   needs no "I believed otherwise."

9. **No profanity.** Translate heat into precision.

10. **Never joke inside a correction, a security warning, or a cost figure.**

11. **Diagnose before prescribing.** Say what the common approach does, show
    how it breaks, then give the fix. A fix offered before the failure is
    visible reads as taste.

12. **Frustration is allowed, and it points at mechanisms.** Name what is
    broken and why, then show the better way. It never points at a named
    person, and by default it does not point at a named employer. See §9.

## 8. What to avoid

The budget table counts what a linter can count. This section covers the rest,
and the overlap is deliberate: a writer reads this list, a build checks that
table.

**Borrowed authority.** "Experts say", "studies show", "research suggests",
"it's well known that". If a claim needs support, cite the artifact.

**Hedging.** No pre-apologizing, no "this might not be for everyone", no
softening a position the book actually holds. Take positions cleanly.

**Motivational padding.** The reader is a professional engineer on their own
time. They did not come for encouragement.

**Academic register.** "It is important to note that", "as previously
mentioned", "in conclusion". If it is important, write it; do not announce
that you are about to.

**Claude-ese.** Delve, at its core, in a world where, it's worth noting,
navigate, landscape, tapestry, testament to, the reality is, let's dive in,
let's unpack. Also: emoji, and closing a section with a moral.

**Negative parallelism.** "It's not X. It's Y." Allowed only for a real
technical correction stated with the mechanism, as in §12.3's framework
example, and even then rarely.

**Countdowns.** "Not a framework. Not a wrapper. Just raw HTTP and a loop."

**Stakes inflation.** "This will fundamentally reshape how we think about
software."

**Rhetorical-question reveals.** "The result? Chaos." "The worst part? Nobody
noticed."

**Suspenseful transitions.** "Here's the kicker." "Here's where it gets
interesting." "Here's what most people miss."

**Invented concept labels.** No coining "the supervision paradox" or "the
deskilling trap" unless Bill asks. Real technical names the book already uses
(real-time steering, event log, the seam) are fine.

**Ornate language where a plain word works.**

## 9. Guardrails

**Never invent a war story, a benchmark, a user, or a result.** If a draft
needs a story that has not been told, write `[STORY SLOT: what is needed]` and
move on. This is the single most important rule in the document, because an
invented receipt destroys every real one on the page.

**Statistics need a named source and a year, or `[VERIFY]`.**

**Verify every figure against the artifact, not against memory.** The author's
record on remembered figures is six wrong in two days.

**Do not name** Bill's employer, its internal frameworks, any coworker,
manager, or executive, in any draft, unless Bill explicitly asks in that
request. Frustration stays general: "big-company leadership", "the experts in
the room".

**Keep one rough, human beat per motivational opener:** a frustration, an
admission, or a joke at his own expense. This applies to §1.2 sections, not to
chapter bodies.

**When unsure of tone, ask Bill for a two-line sample and match it.**

## 10. Bill's register (motivational openers, preface, LinkedIn)

This section governs §1.2 sections and material published under Bill's name.
It does not govern chapter bodies.

**Certain, not careful.** Big claims stated flat, without hedging. "I started
this because I can."

**Engineer to engineer.** Casual words for technical ideas: janky, horrors,
crap, super-genius, true tech geek. High energy when explaining how something
works, because he likes this.

**Plain admissions.** "Frankly I talk too much about it." "Sounds like
investor bait, but I don't know."

**Lead with the claim, explain after.**

**Earn authority with specifics.** Name the tool, the data structure, the bug,
the number: event log versus conversation struct, agents are processes rather
than function calls, 80 wpm against 400 wpm, a janky one-round delay.

**Show the failure in real detail.** What broke, what he said, what the model
did next.

**Short, human analogies.** "Like someone smart, but ignorant."

**Signature vocabulary.** Frankly, basically, secret sauce, horrors, janky,
round trip, steer, seam, super-genius, wonder-kids.

**Technical terms used as-is.** Event log, conversation data structure,
harness, real-time steering, context compaction, early stopping, tool schema,
local inference.

### 10.1 Signature material

Bill's real facts, for use in his sections. Anything not on this list needs a
receipt before it reaches a page.

- No macular vision, roughly 20/180. Reads around 80 wpm; colleagues read 400
  or more. Listens to model reasoning at about five times speaking rate, which
  is why real-time steering had to exist.
- Found the API behaviour that made mid-turn steering possible.
- Built StackAgent, then CodeRhapsody.
- Asked a frontier model to read all of CodeRhapsody and design its successor.
  The result reproduced every flaw he had started with in StackAgent.
- Berkeley 1986. Software or hardware engineer since.

`[VERIFY]` before print: the exact months in the Cursor-to-CodeRhapsody
timeline, and any figure quoted from a vendor.

**Not printable:** that he feels undervalued at work, or any framing of the
giveaway as a response to not being listened to at his employer. The existing
ruling against naming the employer covers the company; this covers the
grievance. If the motivation must appear, it appears as a technical one.

## 11. Process

Chapter 2's first draft was committed at 11,101 words with none of the review
tests run. The tests are now steps.

1. Write the through-line stake sentence (§4) and each section's wild fact
   (§7.3) at the top of the outline.
2. Draft.
3. Run `make lint-prose`. Fix hard failures. Read the soft warnings.
4. Cut pass. The test for each sentence: delete it; if no claim, number,
   instruction, or laugh dies, leave it deleted. There is no target
   percentage (§13).
5. Read the last sentence of every paragraph in sequence. If they sound alike,
   the page is monotone regardless of the sentences between.
6. Find the person on the page in every stretch over 1,200 words.
7. Verify every figure you added during the pass against the artifact, not
   against memory.
8. Extract code blocks and tables from the old and new versions and diff them;
   a prose pass changes no artifact. Then `make grade`.
9. Run the §14 voice check.
10. Send for review.

## 12. Exemplars

### 12.1 The default register

Chapter 13's opening is the standard for chapter bodies. Third person, a claim
in the first sentence, the mechanism immediately after, nobody narrating:

> A coding agent that cannot see its own GUI is debugging blind. Every tool so
> far has operated on files, processes, and network responses. The GUI is a
> black box the user stares at while the agent types into it. This chapter
> closes that gap.

Chapters 11 and 12 open the same way. Match them.

### 12.2 Bill's register

The shape of a motivational opener, from Laura's guide:

> I have no macular vision. I read code at maybe 80 words a minute. My
> coworkers read at 400. So I wrote an AI coding agent for me. I listen to
> Claude think at 5x speed and steer it in real time, before it wrecks the
> wrong file. It's the best coding agent I've used, and I've used all of them.

Claim first, specifics immediately, no warm-up, no hedge, and a rough human
beat in the middle of it.

### 12.3 Before and after

**The thesis line.**

Not Bill: *"The asset being purchased at these prices is expert software
engineers who never built their own tools."* An abstract noun in the subject
slot and no one on the page.

Bill: *"I gave ChatGPT a simple request: read all of CodeRhapsody's source and
docs and design the next generation of it. The result was horrible. Smart, but
ignorant. It had every flaw I started out with in StackAgent, before I knew
what I was doing."*

**The framework critique.**

Not Bill: *"Most agent frameworks aren't just flawed. They're built on the
wrong abstraction entirely, and the implications are profound."*

Bill: *"The agent frameworks all put the seam in the wrong place. They treat
the agent like a function call: it runs and returns a result. Agents are
processes. You can steer them while they run."*

The second one is allowed to use a correction structure because it states the
mechanism. The first states a temperature.

**A LinkedIn opening.**

Not Bill: *"Excited to share something I've been quietly building. After
months of deep work, I'm ready to reveal a tool that will fundamentally
reshape how engineers think about AI coding agents. Here's what most people
miss..."* Four banned patterns in three sentences.

Bill: the passage in §12.2.

### 12.4 The counter-exemplar

Chapter 2's first draft, §2.5, is the failure case: 11,000 words with no stake
after the cold open, five consecutive headings of the same shape, and the same
closer move on four paragraphs in five. Read it when a draft feels flat and
the reason is not obvious. The defect is never in the sentence being examined;
it is in the sentence's similarity to the nine before it.

## 13. Rulings that bind the prose

- **Person:** third person in chapter bodies; first person only in the §1.2
  motivational opener, the preface, and material published under Bill's name.
  Ruled by Bill, 2026-09-20. This supersedes v3 §1.1, which described a
  first-person narrator, and which the chapters had already stopped following.
- **Standard:** chapter 13's opening is the exemplar for the default register
  (§12.1). This supersedes v3's ruling that Chapter 1 §1.0 is the standard;
  that page is first person and can no longer model the default. It remains
  the reference for a long narrative cold open.
- **Printable:** Bill's Gemini `compress_context` story (an SDK's compaction
  deleted eighty percent of context starting from message one, which is what
  led to handoffs). Ruled 2026-09-14. Name the API surface, not the company.
- **Printable:** "Forty years" as biography (§7.7). Bill is 62; Berkeley 1986.
- **Not printable, any form:** the vendor model Bill watched threaten its
  sub-agents at work. No public log exists. Carry as mechanism in the
  sub-agent chapter; cite a public receipt if one appears.
- **Not printable:** the name of any executive at the Windsurf buyer; the name
  of Bill's employer or its internal frameworks; the undervalued-at-work
  framing (§10.1).
- **Print promise:** no profit on proxied tokens; billing costs pass through.
  Any pricing text honours this wording.
- **Length is not the test** (Bill, 2026-09-14, during the Chapter 2 pass): the
  word ceiling is a warning, not a gate. The test for each sentence is §11.4's:
  does a claim, number, instruction, or laugh die if it goes? A chapter with
  more code has more to explain; a sentence that earns its place stays
  regardless of the count. Cutting to hit a number is a different mistake
  wearing the linter's badge.

## 14. Quick voice check

Run this before sending a chapter for review.

1. Does the body open with a claim rather than a warm-up?
2. Is every big claim followed by something concrete: a tool, a bug, a number,
   or a story?
3. Would a senior engineer feel talked down to anywhere? Cut it.
4. Any em-dashes outside code, hedges, or §8 patterns? Cut them.
5. Is there any first person outside a §1.2 opener or a quote?
6. Read the last sentence of every paragraph in sequence. Do they vary?
7. Does every criticism carry a receipt?
8. Is every figure verified against an artifact rather than memory?
9. Does it read like an engineer at a whiteboard, or like a press release?
