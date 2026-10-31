# Voice

*Version 7, 2026-10-18. A specification, written plainly on purpose. An
earlier version written in its own prescribed voice taught the chapter to
copy its tics. This guide describes editorial decisions, not a style to mimic.*

*This revision prioritizes reader understanding, coherent explanations, and
integration of new requirements. It replaces the sentence-level cut test,
relaxes mechanical quotas, and adds sustained technical exemplars. Evidence,
authorship, privacy, and publication guardrails remain in force.*

*Sources: the book's craft specification and Laura Zavelson's brand guide
(`book/suggested-voice.md`, based on Bill's written answers, September 2026).
They serve different surfaces: chapter bodies and Bill's personal register.
Bill's later story-preservation request and codebook ambitions are retained
in §4.1 and §10.2.*

## 0. How to use this

Read §1, §6, and §7 before drafting. When instructions compete, use this order:

1. Truth, technical accuracy, and the guardrails in §9 and §13.
2. The reader's understanding and reason to care.
3. Coherent explanation and useful organization.
4. Voice, rhythm, and stylistic preferences.
5. Lint checks as diagnostics, not the definition of good prose.

A stylistic preference never justifies altering evidence, obscuring a
requirement, or deleting a useful explanation. Run `make lint-prose` before
review and inspect its findings (§5). Record deliberate style exceptions in
the review notes; add durable changes to the guide in §13 with a date.

## 1. Who is speaking

### 1.1 The default is third person

Chapter bodies are third person. No "I" or "we" standing in for an author,
and no invented agent narrator. This is the register chapters 11, 12 and 13 are
written in, and it is the one to match.

The reason is not modesty. An earlier version of this book had the agent
narrate itself, and a reader found it unsettling rather than charming. The
prose should carry personality through concrete choices, candid engineering
judgment, rhythm, and dry humor. Technical density alone does not make a page
worth reading. Keep the people and consequences visible as the mechanism
becomes more detailed.

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
story we do not have is a marked gap, never an invention. Mention him when
his action, judgment, or experience matters to the explanation. Do not ration
his name inside a useful story or insert him merely to satisfy a quota.

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
respect people who have built real things. Skip definitions a senior engineer
already knows, but do not confuse orientation with hand-holding. A reader can
know concurrency and still need a clear account of this system's lifecycle.
See `icp.md`.

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
repeat a summary that adds nothing. Names the reader's objection before the reader
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

When a long stretch of mechanism loses sight of its purpose, reconnect it to
the problem and its consequences. A documented person may carry that stake;
a concrete reader problem can carry it too. A person-gap warning is a prompt to
check engagement, not an instruction to insert a person.

### 4.1 Preserve the human story when rewriting

For every edition rewrite, the author must examine the corresponding
previous-edition material for a documented incident, revealing exchange, reader
frustration, or memorable explanation worth retaining. Record the choice and
its source in the outline, or explain why that material no longer serves the
chapter. The reviewer checks the choice against both versions. This applies
to existing drafts as well as future chapters.

Keep enough of an incident to show what someone wanted, what went wrong,
what they tried, and what the outcome taught. Compressing Bill's two-week bet
into a biographical clause can preserve the fact while losing its reason to
be on the page. Preserve the stakes when they explain why the work matters.
Use Bill's first-person voice in a short motivational opener under §1.2, or
tell the incident in third person in the body under §1.3.

The book teaches the rules before implementation. An earlier
failure can motivate the right design without asking the student to recreate
the flawed architecture. Connect the incident to the decision it explains,
and return to its consequence when the new mechanism resolves the problem.

Personality also belongs in the explanation: name the reader's likely
objection, make a defensible engineering judgment, or let an observed absurdity
speak for itself. Preserve a sharp, accurate sentence when it teaches. Avoid
adding jokes, stock catchphrases, or a personal anecdote to satisfy a quota.
A concrete reader problem can carry a chapter when no historical story fits.

Use actual implementation experience too. A model claiming an edit succeeded
while the file remains unchanged, followed by an observed corrective turn,
can teach more than another instruction to verify results. Retain the failure
and its resolution; distinguish model behavior from an implementation defect.

Historical prose is a lead to evidence, not proof that every detail is true.
Verify quotations, dates, figures, and claimed outcomes. Never invent dialogue,
feelings, scenes, or a successful result. Use the story-slot rule in §9 when
needed. Keep the essential receipt and reproduction steps in the chapter;
link detailed audit ledgers and checkpoint history so they do not overwhelm
the reader's journey. Required exercise contracts remain explicit and easy to
find, separated from the explanation where appropriate (§7.1).

## 5. Style diagnostics and lint

Repeated moves become tics. Counts can reveal them, but cannot decide whether
a particular sentence helps. Read a flagged passage in context before changing
it. A chapter that satisfies every count can still be exhausting to read.

Chapter length has no minimum, maximum, or lint warning. A chapter may be one
page or a hundred; it needs to be the right length for its material and reader.
Do not pad a short chapter or compress a long one to meet a numerical target.

Use these style thresholds as review prompts, not drafting quotas:

| pattern | review prompt |
|---|---|
| confessions or accounts of an earlier draft | more than three: keep the histories that teach; integrate the rest |
| inventory openers such as "Two rules." | more than three: check whether the sections are becoming checklists |
| short final sentences | more than one paragraph in three: read for repetitive emphasis |
| pointer closers such as "That is" or "That number" | more than two: check whether they explain or merely underline |
| negation forms | more than one per 500 words: check whether the correction needs its negative half |
| repeated figures or phrases | more than three uses: check whether the callback adds understanding |

The default register still excludes authorial first person outside the surfaces
allowed by §1. Avoid the stock language in §8 and em-dashes outside code or
quoted artifacts. Do not alter evidence to satisfy a style preference.

`make lint-prose` may enforce older numeric ceilings, including a Bill-mention
limit that this guide no longer endorses. This revision changes the editorial
policy, not the linter implementation. Review every failure: fix real defects
and record justified exceptions for the reviewer rather than distorting prose
to obtain a pass. Report the actual lint result; an editorial exception does
not turn a failed command into a successful one. Any build-policy change is a
separate task.

Passing lint establishes neither readability nor completeness. The explanation
and integration passes in §11 remain necessary.

## 6. Paragraph shape

Let sentence and paragraph length follow the thought. Review for monotony,
but do not manufacture variety. Short sentences can emphasize a consequence;
ordinary connective sentences carry the reader between emphases. A paragraph
may need more than four sentences to complete one useful step of an argument.

**Openings and endings serve the explanation.** Open with the actor, artifact,
question, or claim the reader needs next. End when that unit of thought is
complete. Not every paragraph needs a punch line, a summary, or a surprise.
If a page repeatedly ends with the same emphatic move, revise the passage
rather than mechanically swapping its last sentences.

**Headings help readers find their way.** Name the question, mechanism, or task.
Vary the form when repetition becomes distracting, not merely to make adjacent
headings look different.

**Lists and prose do different jobs.** Use lists for steps, parallel choices,
requirements, and checks that readers need to locate independently. Use prose
to explain causes, dependencies, and consequences. A list of individually
clear rules is not necessarily a clear account of how a system works.

**Prefer identifiable actors.** Say which person, program, request, or event
causes a change. An abstract term can be useful once the reader knows what it
means; avoid abstract subjects that conceal who does what.

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

3. **Each section answers a worthwhile question.** Identify what the reader
   needs to understand, decide, or do. A surprising fact is welcome when the
   material supplies one. Ordinary but consequential explanation needs no
   manufactured surprise.

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

8. **Correct our own wrong claims honestly.** Keep a failure and its result
   when the history teaches the reader something the final rule cannot.
   Otherwise integrate the correction into the current explanation; preserve
   detailed revision history in linked evidence. Never leave a false claim
   uncorrected or imply a failed approach succeeded.

9. **No profanity.** Translate heat into precision.

10. **Never joke inside a correction, a security warning, or a cost figure.**

11. **Diagnose before prescribing.** Say what the common approach does, show
    how it breaks, then give the fix. A fix offered before the failure is
    visible reads as taste. Explain the ordinary path before branching into
    exceptions: walk through a concrete sequence in time order, naming what
    each participant knows, what changes, and why the next step follows.
    Then explain the edge cases against that shared example.

12. **Frustration is allowed, and it points at mechanisms.** Name what is
    broken and why, then show the better way. It never points at a named
    person, and by default it does not point at a named employer. See §9.

### 7.1 Explanation, contract, and evidence

Separate three reading tasks. The main explanation teaches the problem,
mechanism, and consequence. The implementation contract states exact required
behavior and edge cases. The evidence establishes what was observed and how
it was checked. Keep all three accessible, but do not interleave them sentence
by sentence.

Introduce an exception in the main explanation when it changes the reader's
mental model. Put exhaustive cases, grading commands, and regression checks
in clearly labeled exercise or reference sections. Link them at the point of
need. The TL;DR should orient the reader to the problem, design, and deliverable;
it should not reproduce the full grading recipe. Relocation must not weaken,
hide, or contradict a required contract.

### 7.2 Integrate revisions

When a new requirement changes an explanation, revise the surrounding passage
so it teaches the current design as a coherent whole. Do not leave the reader
to assemble that design from an original account followed by successive
corrections. Preserve the history of a mistake when it teaches something the
final rule cannot; otherwise lead with the corrected account.

After an insertion, reread the whole section. Does it explain the design to a
new reader, or does it record the order in which reviewers discovered defects?

### 7.3 State the evidence boundary where it matters

Attribute reports, distinguish observations from inference, and qualify
uncertainty that affects the conclusion. Do the verification work without
narrating every precaution. If an incidental detail requires more qualification
than explanation, omit the detail rather than surrounding it with disclaimers.
Never remove uncertainty merely to make a claim sound stronger.

## 8. What to avoid

These are common symptoms of generic or inflated prose. Review their use in
context; replacing flagged words alone does not improve an explanation.

**Borrowed authority.** "Experts say", "studies show", "research suggests",
"it's well known that". If a claim needs support, cite the artifact.

**Hedging.** No pre-apologizing, no "this might not be for everyone", no
softening a position the book actually holds. Take positions cleanly. Necessary
uncertainty about evidence is accuracy, not a stylistic defect (§7.3).

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

**Business clichés.** Synergy, seamless, holistic, cutting-edge, game-changer,
thought leadership. Name the actual capability or effect instead.

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
- Bill reports that the first edition was published alongside functional
  Ensemble in three weeks of nights and weekends: approximately 80 hours of
  his time and $3,000 in Anthropic API usage. Supplied October 9, 2026.
  Preserve the approximations and the distinction between his time and API
  spend; these are his reported figures, not an independently audited total
  or a comparison with another development process.

`[VERIFY]` before print: the exact months in the Cursor-to-CodeRhapsody
timeline, and any figure quoted from a vendor.

**Not printable:** that he feels undervalued at work, or any framing of the
giveaway as a response to not being listened to at his employer. The existing
ruling against naming the employer covers the company; this covers the
grievance. If the motivation must appear, it appears as a technical one.

### 10.2 Conviction and the codebook's ambition

Bill's own line, supplied and authorized for this document on October 9, 2026:

> I'm not opinionated, I just know the truth.

Keep the confidence and the joke together. Use the line where it reveals Bill
or an engineering disagreement; do not turn it into a repeated catchphrase or
an excuse to dismiss evidence.

Bill believes agentic codebooks can change how complex software is built.
Experts record engineering judgment, reasons and constraints in a book that an
agent can follow. A fresh implementation tests whether that knowledge transfers.
Student difficulties and independent reviews inform the next edition; changed
requirements can be incorporated where they belong, and new chapters can add
new capabilities. His ambition is to make rapid, improved regeneration of a
system practical, with the resulting agent helping build the following edition.
He sees this as a step toward the singularity.

Present that ambition plainly as Bill's conviction. Preserve the reasons behind
it and show the actual experiment: what the student built, what failed, what
review found and what improved. Claims about speed, quality or a completed
self-improvement cycle still need their own evidence. Confidence belongs in
the voice; whether regeneration succeeds is something the work must establish.

## 11. Process

Work from structure to sentences. Lint cannot repair an explanation whose
parts arrive in the wrong order.

1. Outline the through-line stake (§4) and the question each section answers
   (§7, item 3). For an edition rewrite, record the story-preservation decision
   under §4.1.
2. Draft the explanation, then make its contract and evidence easy to find
   (§7.1). Use a concrete sequence before enumerating interacting edge cases.
3. Integration pass. Read each changed section as a new reader would. Rewrite
   around new requirements rather than appending corrections (§7.2). Check
   that relocation preserves every required behavior and its discoverability.
4. Usefulness pass. Ask what each passage helps the reader understand, decide,
   or do. Preserve orientation, causal explanation, and transitions when they
   reduce the reader's effort. A sentence does not earn its place merely by
   containing a unique fact or instruction. Remove, combine, or relocate
   details that interrupt the explanation, even when they are correct. There
   is no target reduction percentage (§13).
5. Read for rhythm and continuity. Scan paragraph endings for repeated emphasis,
   then read the full passages. Let variation follow the thought; do not add
   punch lines or people to satisfy a quota.
6. Verify every changed quotation, figure, and claimed outcome against its
   source. Distinguish observation, attribution, and inference (§7.3).
7. Extract code blocks and tables from the old and new versions and diff them;
   a prose-only pass changes no artifact. Check that reordered requirements
   retain their meaning. Then run `make grade`.
8. Run `make lint-prose`. Inspect findings and report actual results. Fix
   defects; document justified style exceptions under §5.
9. Run the §14 voice check and send for review with the verification results.

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
the reason is not obvious. The defect may be hard to see in an isolated sentence;
it becomes clear in the sentence's similarity to the nine before it.

### 12.5 From requirements to a causal explanation

The next two pairs are constructed editorial demonstrations based on the
edition-3 Chapter 05 wakeup design and Chapter 01 timeout discussion. They are
not quotations, observed runs, or complete implementation contracts. Their
purpose is to show how to organize true technical material.

**Before: requirements carry the whole explanation.**

> On a hint, wake the foreground observation. Do not cancel the job. Resolve
> the outstanding call under its original id with the existing handle,
> bounded unseen output, and truthful state. Record the result and hint before
> the next model request. Do not resend the job or end the turn. Preserve
> completed results and pair undispatched calls with not-executed errors.
> Retire the observation timer. A later wait uses the same job handle.

**After: establish the sequence, then state the contract.**

> The model is waiting for a test when the human asks it to check the first
> failure. The test can keep running while the model answers. What needs to
> end is the current wait for a report, so the human's hint can reach the next
> model request.
>
> The runtime returns the existing job handle, any bounded unseen output, and
> the job's observed state as the result of that wait. It records that result
> and the hint before sending another request to the model. The model can
> respond to the hint and later resume observation of the same job. It does
> not need to launch the test again or wait for the test to finish.
>
> **Required pairing and lifecycle behavior**
>
> - Resolve the wait exactly once under its original call id, using the
>   observed state if completion raced with the hint. Retire its timer.
> - Preserve completed calls' results. Pair remaining undispatched calls
>   with errors saying they were not executed; do not dispatch that tail.
> - Do not cancel the job or end the turn. A later wait refers to the same
>   job handle, whether it is still running or has finished.

The causal sequence now carries the explanation. The exact pairing rules
remain visible in a separate contract, where a reader can check them during
implementation. The rewrite adds orientation rather than optimizing for the
fewest words. A real chapter must also retain its other wakeup requirements,
including signature preservation, one-shot limits, and output-cursor rules.

### 12.6 Integrating a newly discovered failure

**Before: an original instruction followed by a patch.**

> Set a 60-second client timeout so a stalled request cannot hang the agent.
>
> Some valid model responses take longer than 60 seconds, so this timeout can
> kill useful work. Do not use that blanket deadline. Keep connection and TLS
> protections, and make sure cancellation still works.

**After: teach the corrected distinction.**

> A timeout that protects connection setup answers a different question from
> a deadline on the whole model request. The first bounds how long the client
> waits to establish a connection. The second also limits how long a valid
> generation may take.
>
> A blanket 60-second client timeout can therefore cancel useful work even
> when the connection is healthy. Keep bounded connection and TLS setup,
> but do not impose that blanket deadline on generation. Give the request a
> cancellation path so the user can stop waiting without mistaking elapsed
> time for evidence that the model has failed.

The revision replaces the obsolete instruction instead of leaving both
versions for the reader to reconcile. It preserves the connection safeguards
and cancellation requirement, and explains why they serve different purposes.
If an actual failed run is useful evidence, link it or tell its verified story;
do not invent one to dramatize the distinction.

## 13. Rulings that bind the prose

- **Explanation before stylistic quotas:** approved for this revision,
  2026-10-18. The priority order in §0, usefulness test in §11.4, and integration
  pass supersede the old sentence-level deletion test and mechanical quotas
  for Bill mentions, human reappearances, wild facts, and paragraph length.
  Style counts are diagnostics under §5. Factual, technical, privacy, and
  publication guardrails remain binding.
- **Second-edition personality:** preserve documented human story and restore
  warmth and personality during the rewrite. Requested by Bill, 2026-10-07.
  §4.1 is a required author and reviewer pass. This clarifies §1.1; third-person
  bodies, evidence discipline, and architecture taught from the start remain.
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
- **Printable:** "Forty years" as biography (§7, item 7). Bill is 62; Berkeley 1986.
- **Not printable, any form:** the vendor model Bill watched threaten its
  sub-agents at work. No public log exists. Carry as mechanism in the
  sub-agent chapter; cite a public receipt if one appears.
- **Not printable:** the name of any executive at the Windsurf buyer; the name
  of Bill's employer or its internal frameworks; the undervalued-at-work
  framing (§10.1).
- **Print promise:** no profit on proxied tokens; billing costs pass through.
  Any pricing text honours this wording.
- **Length is not the test** (Bill, 2026-09-14, during the Chapter 2 pass;
  clarified 2026-10-18): there is no chapter-length target or linter check.
  A chapter can be one page or a hundred; it needs to be the right length.
  Apply §11.4's passage-level usefulness test. Orientation and transitions
  can earn their space; a unique fact may belong elsewhere. Neither cutting
  to a number nor retaining every fact is a substitute for editorial judgment.
- **Chapter 19 names vendors.** §7, item 6 ("name the API and the model family,
  never the company") is suspended for chapter 19. Anthropic is named for
  their policy decision (freezing the context window); OpenAI is named for
  their approach (OAuth for third-party agents, signature binding as the
  narrower fix). Every criticism carries a receipt (§7, item 4): the arXiv paper,
  the OpenAI blog post, the dated policy change. This is a one-chapter
  exception, not a general license. If Anthropic reverses course, the chapter
  is revised. Ruled 2026-10-02.

## 14. Quick voice check


Run this before sending a chapter for review.

1. Does the opening give the reader a reason to care and a clear direction?
2. Can the reader explain the basic mechanism before meeting its edge cases?
3. Do interacting rules follow a concrete sequence, with actors and causes
   visible, or does a list of prohibitions stand in for an explanation?
4. Are explanation, contract, and evidence easy to find without constantly
   interrupting one another? Are all required behaviors still explicit?
5. Does each revised section teach the current design coherently, rather than
   record the order in which reviewers found defects?
6. Have useful transitions and orientation survived the cut pass? Could any
   technically correct detail be combined, relocated, or omitted to help?
7. Would a senior engineer feel talked down to, or be forced to reconstruct an
   unstated causal step? Fix either problem.
8. Is the rhythm natural across full passages, without manufactured variation,
   mandatory surprises, or a punch line at every paragraph end?
9. Does every criticism carry a receipt? Are quotations, figures, and outcomes
   verified, with consequential uncertainty preserved?
10. Does the prose state relevant evidence boundaries without narrating every
    verification precaution?
11. Are person, privacy, publication, and evidence guardrails respected? Have
    lint findings been reviewed and any style exceptions documented?
12. What human stake, memorable explanation, or documented incident survived
    the rewrite? If one was removed, does the outline explain why? Does the
    mechanism resolve the problem that made the reader care?
