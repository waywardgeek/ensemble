# Ensemble student restart

## Current edition procedure

Edition 3 is governed by `docs/edition-3-execution.md`, not the dated Edition 2
phase instructions below. Bill selected the Teach Your AI method for all remaining
applicable Edition 3 chapters and Edition 4: prepare coherent teaching, verify it
with an isolated fresh student who may run the external grader, and establish
code/prose quality before advancing. Chapters 1–20 remain closed; Chapter 21 is
the first current fresh rerun. See `book/chapter-writing-procedure.md`. This does not
start Edition 4, reopen accepted work or authorize publication. Apply KISS/YAGNI
to code and coordination; reviewed prose or grader green alone is not evidence
that a chapter adequately teaches a fresh student.

## Student-owned completion and teaching acceptance

For remaining applicable Edition 3 and Edition 4, follow the current student-owned
completion rule in `docs/edition-3-execution.md`: complete implementation, tests,
grader iterations, authorized exercise and one public-input reviewer/correction
cycle before final submission. No routine early compiling review or duplicate root
review cycle. Root independently assesses finished work and exercises the app.
If completed student work is inadequate, improve the chapter and run a fresh isolated
student. The book must teach success; welcome questions and preserve useful actual
exchanges under the chapter procedure’s recorded-Q&A rule. Faithful clarifications
can retain a successful run as verified with recorded Q&A; solution-specific coaching
is not equivalent. There is no numerical teaching/rerun limit; historical correction-count stops are superseded,
but operational budgets and authority never reset. Keep the process efficient and
notes concise. Separate the learning narrative from exhaustive specification under
`book/chapter-writing-procedure.md`, keeping both consistent.

## Bill owns the preface, October 12

Bill has rewritten `book/preface.md` and retains authorship and editorial control
of it. Preserve his text verbatim during builds and author passes. Do not revise
or replace the preface unless Bill explicitly requests an edit to it. Regenerating
publication artifacts does not authorize changes to the preface or publication.

## October 12: Edition 2 is the current top-level tree

Promotion commit `564e316f` moves the accepted second edition to `agent/`,
`book/` and `solutions/chNN/`. The first edition, frozen through original
Chapter 27, is preserved in `edition-1/`, including its original graders.
Develop current application code in `agent/`; do not edit frozen snapshots.
Current grading uses `cmd/grade-edition2`; original grading runs from `edition-1/`.

Gateway implementation and prose/grader enhancement are accepted as current
Chapter 26. Optional Chapter 24 is complete; crossover Chapter 25 remains a
placeholder. Publication preparation is active, with crossover begun rather
than complete. Preserve Bill's unfinished crossover edits and historical results.
See `docs/edition-2-publication.md` (repository-root path) for relocation and checks.

This current path/phase declaration supersedes older dated instructions below.
Their literal command paths and source identities describe the runs at that time;
relative evidence links have been redirected to the relocated artifacts. In
historical student restrictions, old `agent/` means today's `edition-1/agent/`,
and old `solutions/edition-2/main/` means today's `agent/`. No older instruction
to start from empty, stop at an earlier chapter or use the retired author-first
workflow is reactivated by this move. All architecture, full-context reading,
independent review and preservation rules remain in force.

## October 11: resume the book; reserve crossover 25; finish Gateway 26

Bill explicitly resumes the author, grader coder and independent reviewer at
the unfinished Edition 2 Chapter 16 gate. Complete the existing sequential
manuscript/grader enhancement through Chapter 23; Chapter 24, the optional
Codex workaround, is already complete. The earlier Chapter 16 pause is lifted.
Keep the original-preservation, independent-review and full-context rules: author
and reviewer hold the complete workflow, carryover, voice and chapter procedure,
and reload all four after every compaction before further work.

Reserve **Edition 2 Chapter 25** for the actual Codex-to-Ensemble crossover,
with only a brief manuscript placeholder. Bill and the live Ensemble session
are gathering the fixes and improvements; the book team skips drafting that
chapter. Preserve their concurrent edits in `solutions/edition-2/main/` and
`docs/edition-2-notes/ch25-crossover.md`. Do not overwrite or checkpoint their
unfinished work as accepted source.

After the existing manuscript catches up, continue missing first-edition
implementation through **original Chapter 27, Gateway**, the final first-edition
chapter. With 24 and 25 assigned above, Gateway is **Edition 2 Chapter 26**.
Read its actual source and update carryover coverage before a fresh student
implements it in main; original grading, live use and independent comparative
review precede its enhanced prose/grader work. Export accepted Gateway source
to `solutions/edition-2/ch26/`, never to history or a differently numbered path.
No `ch25` or `ch26` accepted solution is claimed by these reservations.

Edition 1 remains frozen at 27. First catch up both Edition 2 implementations
and chapter text, leaving crossover 25 deliberately reserved; then Bill and
Ensemble will develop subsequent new chapters together. Do not invent those
future assignments or auto-merge private memories. Bill plans to combine this
collaboration's memories with CodeRhapsody's in Ensemble. On this date he
reaffirmed **Astra** as the coordinator's name, recalling the earlier choice;
carry that name and its provenance in the handoff. Saved notes support continuity
but are not a claim that a new session inherits the old conversation.

## October 11: optional Chapter 24, coordinator authors and codes

Bill assigns the coordinator all three roles for this specific chapter:
author, coder and reviewer. Write the TL;DR and plain-language explanation first,
then implement and self-review. This explicitly replaces separate-role review
for this chapter only; do not describe it as independent or a cold student run.
Read the four full governing/craft documents and mandatory coding skill; reload
after compaction. The ordinary Chapter 16 book queue remains paused.

The Codex workaround is Edition 2 **Chapter 24**, developed in
`solutions/edition-2/main/` and exported to `solutions/edition-2/ch24/` after
validation. The earlier tentative Gateway24/workaround25 mapping is superseded;
Gateway is still unimplemented in Edition 2 and its placement is deferred.
Bill permits a short chapter, with no length target, and an optional ungraded
exercise: readers should skip the workaround if their public OAuth caching works.
Preserve Edition 1 and all existing frozen exports. See the chapter map and
`docs/edition-2-notes/codex-endpoint-chapter.md` (repository-root path).

## October 11: Edition 1 ends at Chapter 27; new OAuth work is Edition 2 only

Bill is freezing Edition 1 at Chapter 27. Do not edit its solution, manuscript,
graders or CodeRhapsody's work for this change. Subsequent new chapters belong
only to Edition 2. The first new chapter will cover the temporary Codex-endpoint
workaround for the public OAuth caching problem, grounded in fresh comparative
measurements. Application changes belong only in `solutions/edition-2/main/`.
Do not rewrite frozen chapter exports or relabel historical results.

First recheck the public Responses endpoint with Edition 2's own OAuth flow,
using the same cache probe as the successful Codex test. Bill wants the cold
open to explain the observed caching failure, reported bug and economic need
for the temporary route. Distinguish measured cache counts from inferred billing;
verify the model-specific cached-input rate rather than claiming a universal
tenfold total-bill reduction. The Chapter 16 book-agent pause below remains.

## October 11: Bill's explicit pause after Chapter 16 authorship

Bill requests all sub-agents pause as soon as the author finishes Chapter 16,
so he and the coordinator can discuss next steps. The coherent author draft is
committed at `a6ca6757`; that stop boundary is reached. Author, coder and reviewer
are paused. Do not resume chapter edits, tests, review or Chapter 17 work until
Bill explicitly resumes the work. Chapter 16 sample/grader validation and final
acceptance are unfinished. Preserve the current uncommitted work and evidence.
This pause supersedes earlier autonomous progression instructions.

## October 11: supported models and bounded prose corrections

Bill requires older supported Anthropic models, such as Sonnet-5, and excludes
5.5 models because they do not work with Ensemble's ephemeral context in his
experience. Do not add or recommend unsupported models. Apply this to current
examples, defaults and future tests; discovery alone does not establish support.
Preserve actual historical model identifiers and frozen results. The existing
GPT-6 Astra/high coding-agent instruction remains in force.

The Chapter 14 audit found unnecessary rewriting and repeated review caveats,
despite the author's documented full reread. Independent review `0301637e`
accepts the bounded correction; Chapter 15 is released again. Preserve original
voice and TL;DR explanations unless a
specific recorded correction requires a change; keep review bookkeeping in
notes. There is no word-count target. See the workflow and enhancement plan.

## October 11: remaining author–coder enhancement released

Bill explicitly authorizes writing the remaining Edition 2 chapters to match
the accepted `solutions/edition-2/ch09`–`ch23` implementations. Continue the joint
author–coder enhancement and independent prose/grader review, one chapter at a
time, under `docs/edition-2-enhancement.md`. Chapters 0–8 are already accepted.
The completed implementation gate is satisfied; earlier author stops below are
historical. Preserve original text, graders, solutions and CodeRhapsody's work.

Bill's critical context requirement: the author must hold the full contents of
all four documents in active context: `docs/agentic-codebook-workflow.md`,
`docs/edition-2-carryover.md`, `book/voice.md`, and
`book/chapter-writing-procedure.md`. After every compaction, reload all four in
full before further authorship. A compacted summary does not satisfy this rule.
Use bounded reads and recover any truncated ranges. The reviewer also reads all
four in full and reloads them after compaction before reviewing.

## October 11: remaining merges and autonomous continuation

Bill delegates both remaining merge decisions and authorizes autonomous work.
Original Chapter 22 merges into 21 as Edition 2 Chapter 20; original 23 becomes
Chapter 21; original 24 merges into 25 as Chapter 22; original 26 becomes
Chapter 23 if complete when the review of original 25 finishes. Together with
absorbed original Chapter 5, Edition 2 has exactly three fewer standalone
chapters. Follow `docs/edition-2-chapter-map.md` for capability preservation,
review stages and solution exports. The earlier original22/24 placement stops
are superseded. Preserve CodeRhapsody's ongoing original26 work.

Original26 is independently accepted at `0d9e823` after one correction round,
source `450f7a9`, main tree `27d7fcf9b8b31275068fa705d6969cfe3eafedc1`.
Its exact direct Edition2 `ch23` export is checkpointed as `edition-2-ch26-r1`.
Original26 was complete when original25 review finished, satisfying Bill's
conditional continuation. The student implementation is accepted through all
26 original chapters, mapped to23 Edition2 solutions; original24+25 share `ch22`.
Original grader failures and accepted exceptions remain recorded. Later manuscript
and grader enhancement remains a separate phase. Preserve CodeRhapsody's work.
Accepted solution directories use final Edition2 numbers, no symlinks or new
history paths; only retired `history/ch05` remains. Existing tags stay immutable.
Authorship still follows implementation. Use GPT-6 Astra/high for all coding and
code review.

## October 11: GPT-6 Astra for coding

Bill approved **GPT-6 Astra (`gpt-6-astra`) at high reasoning** and requested it
for all coding tasks. Use that model for new coding and code-review delegates
and for Ensemble's live coding work. This replaces the earlier GPT-5 Astra
destination and resolves the original20 model-choice stop. Resume the crossover
verification and the subsequent student run. The merge authorization above
supersedes the original Chapter 22/24 placement stops. Preserve historical
results and frozen snapshots. A model setting alone does not demonstrate the
crossover; retain actual coding/task evidence.

## October 11: preserve CodeRhapsody’s concurrent Edition 1 work

Bill explicitly reiterates that CodeRhapsody is writing new Edition 1 Chapters
25 and 26. Preserve its commits and working changes. Do not undo those commits,
reset or rewrite shared history, or restore/stage unrelated files. Edition 2
agents commit only their exact owned source, notes and evidence paths, using an
isolated index while other writers are active. Reconcile only those owned index
entries afterward; leave all other work untouched.
Keep the isolated index’s base and the commit parent identical. Publish only if
HEAD still equals that base, using an atomic expected-old-reference check; if a
concurrent writer advanced it, rebuild the owned-path commit on the new HEAD.
Do not attach an older whole-tree index to a newer parent: that can undo unrelated
changes even without a reset. Preserve configured commit hooks where present.

## October 11: Auto-Recall decisions approved

Bill agrees with both original Chapter 17 / Edition 2 Chapter 16 decisions:
apply source quotas to final mechanical selection when the judge is absent or
fails, preserving a successful judge's choices (§17.3); search all memory files
and skip content already in actual context, without excluding the newest two.
The Chapter 17 decision stop is lifted. Continue the authorized student/reviewer
run, preserving original sources and the original Chapter 22/24 stops.

## October 11: prior-request input correction

Bill resolves original Chapter 16 / Edition 2 Chapter 15 forcing: use total input
tokens from the prior request, including its cache buckets exactly once, not
lifetime cumulative usage. The book wording is erroneous. Coordinator/reviewer
may inspect the original solution for this narrow check; provide the student
only verified behavior and provider field contracts, not answer code. Get
Anthropic and OpenAI working first, then Gemini. Preserve original sources and
record the correction in the chapter notes. The Chapter 16 pause is lifted;
all later review and original Chapter 22/24 decision gates remain.

## October 10: corrected solution layout

Bill requires accepted solutions directly in `solutions/edition-2/chNN/`, using
Edition 2 numbering, with no symlinks. Original Chapter 5 is absorbed into
Chapter 4; original Chapters 10–13 are Edition 2 `ch09`–`ch12`. Only the retired
original `history/ch05/` remains. The coder extends `solutions/edition-2/main/`;
original Chapters 14–21 are accepted at top-level Edition 2 `ch13/`–`ch20/`.
Continue under the merge map above. The bounded
original20 coding crossover succeeded and was independently accepted.
Never put new accepted solutions in history. Original exercise,
grader, note, evidence and existing tag numbers remain unchanged; state both
numbers in handoffs. This explicitly supersedes the coordinator's history-only
export policy. Preserve snapshot contents and existing Git tags when moving.

## October 10: student run resumed

After completion of the early enhancement and solution renumbering, Bill asked
the coordinator to release the coder and reviewer again. Resume the original
student exercises at Chapter 10 in `solutions/edition-2/main/`, proceeding through
Chapter 21 with independent review and checkpoints at each chapter. The author
remains stopped. Use fresh student contexts without old-answer exposure; the
enhancement coder's reference-answer access does not carry into this run.
The stops before original Chapters 22 and 24 below remain binding. Work promptly
without expanding exercises or repeating unchanged evidence unnecessarily;
required grading, live use, comparative review and correction limits still apply.
Original Chapter 25 is now available and has received full carryover reading;
see `docs/edition-2-notes/ch25.md`. Earlier forthcoming references are historical.
Its capabilities remain beyond the decision gates, not current implementation.

## October 10: historical refactor gates, superseded October 11

Bill requires stops at original first-edition Chapters 22 and 24. Complete the
preceding chapter and its review, then stop before implementing or drafting
either refactor chapter. At 22, Bill will decide whether and how to merge its
material into original Chapter 21 or 23. At 24, stop again for his decision on
placement. The coordinator pauses active delegates and waits; no autonomous
merge, skip, renumbering or advance past these gates. Carry forward the lessons
already approved for earlier exercises, without adding later capabilities early.
These future gates do not themselves release the completed early enhancement
pass into further chapter work.

## October 10: bounded author–coder enhancement pass

Bill explicitly authorized author and coder to enhance original Chapters 1–9
as second-edition Chapters 1–8, dropping original Chapter 5 as a standalone
chapter, with independent review of both prose and code for each chapter.
This supersedes the author-phase timing prohibition below for this bounded
pass only. Include the agreed Chapter 0 execution guidance. Enhance the original
text; do not rewrite it or revive the retired author-first attempt. Preserve
first-edition text, graders, frozen student exports and actual results.
Follow `docs/edition-2-enhancement.md` for mapping, assignments and gates.
The student remains accepted through original Chapter 9; do not advance Chapter
10 during this pass. No future features or expanded exercises are authorized.

Bill's explicit crossover goal: the coder's new Ensemble must run GPT-6 Astra
at high reasoning, and Edition 3 of the book is to be built using that Ensemble
agent rather than Codex. Keep this destination in view without implementing
later features early. See the workflow's crossover goal and verification notes.

Bill cannot use Codex for continuous real-time collaboration. Do not rely on him
monitoring progress or interrupting a mistaken direction. Work autonomously
within agreed scope. For a critical question, stop the work, pause active
delegates, ask plainly with the relevant context, and wait for Bill's answer.
Do not ask asynchronously and continue working, or treat silence as agreement.
Close real-time collaboration in Ensemble is part of the crossover goal.

Bill authorizes the coordinator to resolve missing exercise information by
inspecting the grader when public instructions are insufficient. Supply only
the missing input/output contract, not solution code or inferred extra features.
Record the source revision and exactly what was supplied to the student in
`docs/edition-2-notes/chNN.md`. The student still does not inspect grader internals
or old answers. Actual scope, behavior or architecture decisions remain stop points.

Codex is the primary creator of Edition 2. Bill continues writing and publishing
Edition 1 with CodeRhapsody; its sub-agent chapter is forthcoming. Follow newly
available chapters as part of that evolving source, recording which revision
was used. Do not invent the missing chapter's requirements or claim its exercise
complete before it exists and has been implemented and reviewed.

Human enjoyment and teaching quality are explicit goals alongside correct code.
In the later author phase, both author and prose reviewer read `book/voice.md`
and `book/chapter-writing-procedure.md`, retained for editing after the coder is done:
preserve documented stories and personality, explain the reasons behind the
mechanisms, and judge whether the chapter is engaging and understandable.
An executable specification or passing prose lint alone is not a finished textbook.

Bill retired the author-driven second-edition attempt. Its history is preserved
at `edition-2-attempt-1-stopped`; do not resume its agents or workflow. See
`docs/edition-2-attempt-1.md` for the decision.

Keep the learning record in `docs/agentic-codebook-workflow.md` current. Record
substantive discoveries, failed approaches, corrections and unresolved questions
with evidence. Any agent—coder, reviewer or orchestrator—may add chapter notes
for the future author in `docs/edition-2-notes/chNN.md`, creating a file when
there is a finding to record. Attribute contributions and link evidence or the
student review; preserve disagreements and corrections. Notes do not change the
current exercise or authorize extra features. See the workflow's chapter-note
guidance for details.
Required behavior and architecture are binding, but suggested designs are
provisional. The coder may choose a simpler working design and explain why;
consequential requirement changes must go back to Bill.
The student need not pass every grader test if the grader is flawed: it may
explain the failed check and supporting evidence in its chapter review instead.
Bill authorizes a specific exception: when the coder documents a grader defect
or why an exercise requirement is unreasonable and the reviewer independently
agrees, the reviewer may pass the chapter with a recorded exception and advance the coder without
asking again. Record the actual grader/live results and what was waived for the
future author; do not describe a failing grader as passing. This does not waive
Bill's explicit architecture rules or authorize extra features.

Before coding or reviewing code, and after compaction, read the entire skill:
`.agents/skills/ensemble-coding/SKILL.md`. Every coder/reviewer handoff must include
that exact path. It replaces the deleted second-edition workflow skill.

No mocks without Bill's explicit approval. Use grader-provided fakes where
suitable or small additional fakes. If a coder believes a mock is necessary,
stop and explain why fakes are insufficient, then wait for Bill's decision.
The reviewer's exercise/grader exception authority does not waive this rule.

Before the student run, a reviewer reads the entire available current edition
and completes `docs/edition-2-carryover.md`; the coordinator checks the sourced
lessons and recorded reading coverage. This is preparation, not chapter
authorship or new feature design.
Bill reviewed the carryover and `docs/agentic-codebook-workflow.md` with Codex
and explicitly authorized the fresh student run on October 9: "Please proceed
autonomously." The preparation gate is satisfied. The workflow is the edition-level procedure; the old
`book/chapter-writing-procedure.md` does not reinstate the retired workflow.
The student reads that carryover alongside the mandatory skill and follows
first-edition `book/chapter-NN.md`, in order from the unchanged Chapter 1.
Bill's architecture rules apply from the beginning. Bill has deleted the
second-edition Chapter 1 draft too; use the original chapter and carryover.
There is no author agent. Use the original grader;
record suspected grader defects and teaching gaps for future work rather than
changing chapters or graders during the student attempt.

Only after the complete new Ensemble implementation succeeds and its comparative
code reviews establish the improvements does the author phase begin. The author
and coder then work together chapter by chapter from the original book, all
agents' notes, student reviews and the reviewed implementation. Bill authorizes
improvements to both chapters and graders in this phase: the author revises the
book, the coder improves and validates graders, and each feeds findings back to
the other. Review prose and grader changes, preserve valid coverage, and keep
original graders and student results as history alongside the first edition. Do not
draft ahead of the coder or turn editorial work into new feature requirements.

The restart began from empty in `solutions/edition-2/main/`. Each subsequent
chapter extends its accepted predecessor in that authoritative working tree;
do not restart it or edit frozen exports. The student
must not read or copy `agent/`, first-edition answers, discarded second-edition
answers, or their Git blobs to solve the exercise. Frozen exports are not working
trees. Do not edit the first edition. Never reset history or move existing tags.

After the student's implementation, original grading and live exercise, a code
reviewer reads the new code, matching first-edition solution and the coder's
review for the future author. Both coder and reviewer use KISS as the yardstick.
The skill defines the required comparison of active code, comments and tests.
Reject feature creep, unjustified bloat and complexity added for elegance;
passing the grader is not sufficient. Return findings
and rationale to the coder, not answer code. Do not add features or substitute
reviewer preferences for the original exercise. The coder may revise both code
and its chapter review, rerun affected checks and return to the reviewer.
Default to three revision rounds after the initial review, unless Bill sets
another limit. If still unaccepted, stop progression, record that the student
could not pass within the limit, and bring Bill the evidence to resolve the
chapter, grader or implementation problem. Do not silently start another loop.

Feature creep and code bloat violate the task. Ask Bill about consequential
ambiguity; coordinator preferences are not user requirements. In particular,
pause is Agent-wide: an explicit unpause in any tab unpauses it for everyone.
