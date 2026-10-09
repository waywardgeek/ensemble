# Agentic codebook: edition workflow and learning record

**Draft for Bill's review, October 9, 2026.** This is the edition-level procedure
for building Edition 2. Review it together with the
[carryover document](edition-2-carryover.md) before starting the new coder.
The current work is preparation and manuscript review only. Bill and Codex will
review both documents together; do not start Edition 2 coding before that review
is complete and Bill says to proceed.

These are the two review documents. The carryover answers “What did the previous
edition teach us that the student should know?” This workflow answers “How do
we produce the next implementation and, afterward, its book?” The coding skill
supplies the working instructions; chapter notes preserve discoveries.

Keep both [voice.md](../book/voice.md) and the
[chapter writing procedure](../book/chapter-writing-procedure.md) for the author
and prose reviewer to use when editing the book after the coder is done.
The chapter procedure still contains instructions from the retired author-first
attempt, including renumbering and paths that no longer apply. Reconcile those
parts at the later author phase. This edition workflow and Bill's current
instructions govern the overall sequence; using the chapter procedure does not
restart authorship ahead of implementation.

Bill's goal is to establish whether an agent can follow the book as a student,
build an excellent AI coding agent framework from scratch, pass the original
graders, and demonstrate the features through a usable interface with real
models. Autonomous completion and an improved implementation are outcomes to
demonstrate, not claims established by the first attempt.

Bill has designated Codex as the primary creator of Edition 2, responsible for
coordinating the student implementation, reviews, accumulated lessons and the
eventual evidence-based manuscript. The author phase still follows successful
implementation; this responsibility does not restore writing ahead of the coder.

Edition 1 is a living source. Bill writes it with CodeRhapsody and publishes
successive versions on Amazon as he adds chapters. On October 9 he reports that
CodeRhapsody is writing Chapter 24, the GUI refactor that moves the GUI out of
`internal`. Its text was not available in the preparation snapshot. The sub-agent
chapter is also forthcoming. Use those chapters when available; do not infer
their details from their titles. Do not treat the current chapter
count as final or invent a replacement assignment. Record the source revision
used for each exercise; when new or revised chapters arrive, read them and update
the carryover and affected chapter notes before applying their lessons. Existing
checkpoints remain evidence of the versions actually followed.

## A textbook humans enjoy

Bill explicitly wants an enjoyable textbook for human students as well as
instructions an agent can execute. He relies on Codex and Claude Opus 4.6 for
the writing. Technical correctness, student success and reading pleasure are
distinct goals; none can stand in for the others.

During the later author phase, the author and prose reviewer both read
[voice.md](../book/voice.md) and the chapter writing procedure. Follow the voice
guidance on motivation, explanation and preserving the human story. It already
calls for personality, candid engineering judgment and documented experience.
Keep those qualities as the
book becomes more precise. Explain why a rule matters before demanding compliance,
give mechanisms concrete examples, and preserve the discoveries and frustrations
that make the work meaningful. Never invent anecdotes or results for effect.

The prose reviewer considers the reader's experience as well as accuracy:
motivation, clarity, pacing, useful examples and whether the chapter sustains
interest. Prose lint cannot make that judgment. Keep detailed workflow records
outside the teaching narrative unless they help the reader understand something.
Agents may capture memorable moments in chapter notes now; manuscript writing
still waits for the implementation and its lessons.

The implementation itself teaches. Bill's target is roughly 20% comments,
about one comment line per four code lines excluding blanks, with explanations
of why the design works, ownership and non-obvious choices. The coder writes
these explanations alongside the implementation; the code reviewer assesses
their usefulness as well as their quantity. Do not postpone them to authorship,
remove them to reduce code counts, or add filler to hit a percentage.

## Explicit crossover goal

Bill requires the coder's new Ensemble to run **GPT-5 Astra at high reasoning**.
The third edition of the book is to be built using that new Ensemble agent,
rather than Codex. The framework must become the working agent for the next
iteration; completing chapter exercises alone does not establish that crossover.
Bill expects the forthcoming sub-agent work to help Ensemble become a capable
coder for Edition 3. That is an intended outcome to verify through use, not a
claim that Ensemble already matches Codex's capabilities.

Preserve Bill's requested model and reasoning setting. When the relevant model
integration and crossover work arrives, verify the supported provider model
identifier, access method and reasoning parameter against current documentation
and live behavior. Do not infer an API identifier from the display name or
silently substitute another model. Record any availability question for Bill.

Demonstrate the configured model doing real coding work through Ensemble's own
runtime and human interface, recording what actually worked and what remains.
Distinguish readiness for Edition 3 from the later fact of producing Edition 3
with Ensemble. This destination does not expand early chapter exercises or
authorize a new orchestration framework; develop capabilities when their
chapters call for them and document gaps in the chapter notes.

This is a living record. Bill has asked us to document everything we learn in
this process. Record substantive discoveries, failed approaches, corrections,
unresolved questions and evidence as they arise. Keep chapter-specific details
in shared chapter notes and the student's review; bring lessons affecting the
whole process here.
Do not wait for a successful chapter to record a failure or a scope problem.

## Sources of authority

- Bill defines the goal, required behavior and architectural constraints.
- The original first-edition chapter defines the current student exercise.
- [AGENTS.md](../AGENTS.md) and the mandatory
  [coding skill](../.agents/skills/ensemble-coding/SKILL.md) govern execution.
- The [carryover document](edition-2-carryover.md) gives the student sourced
  lessons to apply from the start or when the relevant feature arrives. The
  original Chapter 1 stays unchanged. Bill has deleted the last second-edition
  Chapter 1 draft as well; no second-edition manuscript is a student prerequisite.
- The [first-attempt history](edition-2-attempt-1.md) records the retired work;
  its expanded contracts are not requirements for the restart.

Required behavior and architecture are binding. Suggested designs and
aspirational directions are provisional: the coder may reject them for a simpler
working approach and explain why. When actual requirements conflict or their
meaning changes user-visible behavior, ask Bill before implementing the affected
part. Neither the coordinator nor a reviewer may silently make that decision.
Bill has delegated one explicit exception: the reviewer may accept an
unreasonable exercise requirement as described below and pass the coder onward.

## Prepare the carryover before starting an edition

Read the entire available current edition before launching its successor's
student run. A reviewer is now assigned to this preparation; it is lesson
extraction, not an author drafting new chapters. Extract the lessons learned
during that edition into a compact carryover document, giving
each lesson a source, reason and point of application. Record the source revision
and reading coverage, including uncommitted source changes when applicable;
a partial review must not be described as complete. Cover lessons throughout
the edition, not just its architecture-repair chapters. Separate principles
that apply immediately from lessons to apply when the relevant feature arrives.

Present the complete available-source carryover and this workflow to Bill for
review together. Identify unresolved contradictions, recommendations that need
his judgment and forthcoming chapters explicitly. Incorporate his corrections
before the coder starts. New chapters and materially revised source need a
corresponding carryover update; do not silently claim they were already reviewed.

Distinguish corrections and Bill's explicit decisions from tentative advice.
Keep future features in their own exercises, exclude old solution code, and
resolve consequential contradictions with Bill. This preparation teaches known
lessons without rewriting chapters or designing the next implementation ahead
of experience. Students read the carryover and skill with the original chapters;
reviewers check that the carryover has not become a new source of feature creep.

Bill notes that CodeRhapsody keeps the `short` summaries from `learnings.json`
in context alongside its memory.
Use that continuity principle in student handoffs: retain the accepted carryover,
mandatory skill, current assignment and a brief account of relevant decisions,
rejected approaches and unresolved questions. After compaction, reload the
governing documents; a summary does not replace them. Date summaries and link
their evidence, distinguishing accepted instructions from proposals and historical
observations. Keep old answers out of the student's context. Maintain this in
the existing handoff and chapter notes; no additional reporting system is needed.

## Learn the process before automating orchestration

Bill uses “dynamic workflow” to mean code that orchestrates sub-agents. The
coordinator initially confused this with interactively assigning and messaging
agents; those are different mechanisms.

For now, coordinate interactively while learning which agents are useful, what
their skills need to say and how their handoffs should work. Record adjustments
and outcomes here. Do not build an orchestration system during this discovery
phase. Once the process works reliably, consider encoding its demonstrated
handoffs and decisions in a small automated workflow. Automation is a possible
later step, not a current deliverable or proof that the process is sound.

## How the agents work together

| Role | Responsibility and boundary |
| --- | --- |
| Bill | Resolve product and consequential architecture questions; supply direction and corrections. |
| Coordinator | Keep scope, handoffs and records coherent; route questions to Bill; stop drift and unnecessary work. Do not invent requirements. |
| Carryover reviewer | Read the current edition before coding, extract sourced lessons and surface contradictions for Bill. Do not write new exercises or implementation plans. |
| Student coder | Read the carryover, original chapter and mandatory skill, implement from scratch or extend its accepted predecessor, run the original grader and live exercise, and leave a review for the future author. Do not consult old answers. |
| Code reviewer | After implementation, compare against the matching first-edition solution and read the coder's review. Enforce KISS, architecture and actual scope; return findings and reasons, not old solution code. |
| Future author | Begin only after the complete implementation succeeds and comparative reviews establish its improvements. Revise chapters from the demonstrated results and student experience. |
| Prose reviewer | During the author phase, assess accuracy, evidence, voice, readability and preservation of the human story; return findings to the author. |

There is no author or grader agent working ahead of the student. Suspected
grader defects are recorded for a future maintainer; the student does not change
the grader to pass. Agents may raise questions through the coordinator, who must
relay consequential questions to Bill rather than answer on his behalf. Autonomy
does not remove the need for collaboration when the assignment is ambiguous.

CodeRhapsody's `SOUL.md` offers a useful collaboration rule: make it easy to
admit uncertainty and correct mistakes. Give concise intentions, findings and
limitations while working so Bill can steer in real time. Treat his corrections
as instructions for the ongoing work. Challenge findings and designs without
blame or threats; help a struggling agent identify the missing fact or mistaken
assumption. A confident report is still a claim to check against evidence.

For each chapter:

1. Give the coder the carryover, original chapter and exact mandatory skill path.
   Start with a fresh context containing permitted materials, not a fork of an
   orchestrator or reviewer conversation that has seen old answers. Keep old
   solutions and discarded requirements out of the coder's handoff and follow-ups.
2. Implement the exercise using the agreed architecture and the simplest clear
   design. Record material difficulties and decisions while they are fresh.
3. Run the original grader and appropriate checks. Actually use the human
   interface with a real model to exercise the chapter's features. Record failures
   honestly; fake-server success cannot stand in for a live demonstration.
4. Leave the student review. The code reviewer then compares active production
   code, comments, tests, design and scope with the original solution, using the
   counting and review rules in the skill.
5. Address findings, rerun affected checks and have the revisions reviewed.
   Checkpoint the accepted source and exact snapshot with a new immutable tag.
   Preserve failed attempts and earlier revisions in history.

Keep one authoritative working source tree at `solutions/edition-2/main/` when
coding resumes, tracked by this outer repository without a nested Git repository.
Chapter snapshots are exports, not competing working trees. Use new revision
tags; never reuse or move an earlier attempt's tag.
Carry corrections forward and revalidate affected behavior. Use concise reviews
and sanitized test observations rather than building a new evidence system.
Bound live runs and retries, clean up large generated test data, and never put
credentials in the records. Do not repeat unchanged successful checks without a
reason; investigate persistent failures instead of spending indefinitely.
CodeRhapsody reports interference between concurrent graders; run the original
graders sequentially unless their isolation is established. Check what their
commands actually executed. Inspect the staged diff and commit only owned paths;
a shared Git index can already contain Bill's work before an agent stages its own.

Architecture review checks actual imports, constructor calls and paths to owned
data, rather than finding identifiers such as `Host` or `Agent`. Also inspect
where function bodies live: a hub can obey the import rules while accumulating
behavior that belongs in the spokes. CodeRhapsody's short learnings call out this
failure explicitly; import edges alone cannot establish good organization.
Useful small checks, when those capabilities are present, include calling the public library
from an external consumer, giving two Agents distinct conversations and checking
their histories and usage remain separate, and confirming a request failure
reaches the application's logger through the owner chain without exposing keys.
The deleted Chapter 1 (§§1.9–1.10) made these examples concrete. Use them to test
the existing architecture; they do not require another acceptance framework or
a new paid multi-agent demonstration for every chapter.

### Reviewer acceptance of an unreasonable exercise

The reviewer reads the coder's notes, including objections to the exercise.
If the reviewer independently agrees that a requirement is unreasonable, Bill
authorizes it to pass the chapter with an exception and move the coder to the
next chapter. No further approval is required for that decision. Record the
requirement, evidence and reasoning, reviewer agreement, accepted departure and
any implications for later exercises in the chapter notes for the future author.
Keep actual grader and live results alongside the acceptance decision: a failed
or unrun check remains failed or unrun. Do not edit the original grader or chapter
to manufacture a pass. Bill's explicit architecture rules still apply.

This lets the student and reviewer learn that an assignment needs correction
without spending indefinitely to satisfy it. An objection deserves substantive
review; neither its mere presence nor the time spent establishes that it is right.

## Turn the completed implementation into the next edition

Only after the coder has completed the new Ensemble and its comparative reviews
does the author begin. Account for forthcoming first-edition chapters and any
accepted exceptions when deciding what “complete” covers; do not equate reaching
the currently available last chapter with achieving the full crossover goal.

Give the author the original edition, Bill-reviewed carryover, this workflow,
all agents' chapter notes, the coder's reviews, comparative findings, accepted
exceptions and the actual implementation and live results. The author reads
both `book/voice.md` and `book/chapter-writing-procedure.md`; the prose reviewer
does too. Use them to guide each chapter's editing and review, with this workflow
governing the edition's sequence. The author incorporates the lessons
directly into the new book where readers need them: execution guidance near the
beginning, architecture before the first code, and feature-specific lessons in
the relevant chapters. The final book must teach these lessons rather than
requiring readers to reconstruct them from private workflow notes.

Revise chapter by chapter from what worked and what the student actually
learned. Preserve worthwhile stories and explanations. If evidence is missing
or a claim is unresolved, say so and resolve it rather than inventing an outcome.
Keep original chapter numbering during the student run; decide the new book's
organization from the completed experience rather than imposing the discarded
attempt's chapter count or mapping. Include Bill's requested final comparison
of the editions, using the collected code, comment, test and live-use evidence.

The prose reviewer checks the new chapters and returns findings to the author.
Editing the book must not silently add implementation obligations or revive an
author-first loop. Grader improvement proposals remain recorded for a future
maintainer and separately authorized work; they do not retroactively change the
student's results. Preserve the first edition and its historical evidence.

When the second-edition manuscript is ready for publication, Codex writes a
new epilogue in its own first-person voice, as Bill requested. This is an
explicit exception to the ordinary chapter voice, not permission to impersonate
Bill. Write it from the completed experiment and its actual outcomes, including
the failed first attempt and what was learned. Leave the first-edition epilogue
intact. The epilogue is the final writing step, not a prediction drafted now.

Publication and pushing remain Bill's decisions. Local checkpoints document
the work; they do not imply editorial approval or a published release.

## Lessons established so far

1. **Writing ahead changed the assignment.** The first attempt replaced small
   original exercises with elaborate contracts. The persistence chapter grew
   from a first-edition delta of 254 physical production Go lines to 4,585.
   These are chapter deltas, excluding tests, not executable-statement counts.
   The future book must follow working code and the student's experience.
2. **Review can reinforce a mistake.** A reviewer checking an invented contract
   can approve unnecessary complexity. Review must trace behavior to the original
   exercise or Bill's instructions and challenge scope independently of passing
   tests. The coordinator approved expanded requirements and shares responsibility.
3. **Product choices cannot be inferred from architecture.** Per-tab pause
   ownership entered the author/coordinator contract before coding. Bill intended
   one Agent-wide pause state: explicitly unpausing in any tab unpauses it for
   everyone. Ownership boundaries did not justify a new user-visible veto system.
4. **More code and more tests do not establish improvement.** Compare chapter
   deltas against each edition's own predecessor; separate production, tests,
   examples and languages. Examine useful comments and protected behaviors.
   Dense code, removed explanations and tests for invented features can distort
   superficial comparisons. Better tests remain valuable when they protect real
   requirements and failure modes.
5. **Architecture can be taught before failure.** Retain star dependencies,
   common interfaces, behavior in its responsible package and parent-interface
   ownership chains from the start. Bill permits private runtime implementations
   behind those interfaces. These rules do not require speculative subsystems.
6. **The student experiment needs a clear information boundary.** A coder who
   reads the old answer is no longer demonstrating that the chapter is sufficient.
   The reviewer can inspect the answer afterward. Record assistance and teaching
   gaps so eventual success is described accurately.
7. **Human usability requires human-style testing.** A grader protocol alone
   missed the intended interactive CLI experience. Taking a chapter for a spin
   means using the interface a reader would use, with a real model behind it.
8. **History is part of the result.** Bill requested explicit deletion, not a Git
   reset, so the unsuccessful attempt remains inspectable. Preserve the first
   edition as a historical artifact and distinguish checkpoints from endorsement.
9. **Carry lessons separately from the exercises.** Bill clarified that an
   upfront carryover document removes the need to revise Chapter 1 before the
   student starts. Read the prior edition for its lessons; give the student the
   distilled guidance while preserving the original exercise and its grader.

The first-attempt history links these observations to preserved material. Bill
reported substantial subscription usage; exact cost attribution was not
measured. The repository evidence establishes scope expansion, not a diagnosis
of which training data or internal model mechanism caused it.

## Chapter notes belong to every agent

Any agent may contribute a lesson: the coder, code reviewer, orchestrator or
another participating agent. Add chapter-specific notes to
`docs/edition-2-notes/chNN.md`, using the original chapter number and creating
files only when there is something to record. The coder's required review can
live there or be linked from it; do not duplicate reports merely to fill a format.

A useful entry identifies its contributor and role, the chapter/source revision,
what was observed, the supporting evidence and the proposed teaching improvement
or open question. Distinguish an observation from a hypothesis or Bill's decision.
Agents can add notes as they learn, including during difficulty or failure;
they need not wait for chapter completion or permission to contribute. Preserve
other contributors' findings and disagreements, adding corrections with reasons.

Notes collect evidence for later authorship. They do not amend the current
assignment, add acceptance criteria or make a suggestion binding. A reviewer's
explicit acceptance of an exercise exception is a separate, recorded decision
under Bill's authorization above. Keep old solution code and answer-revealing
comparisons out of notes supplied to the
student; the reviewer returns findings and rationale through the existing review
process. The future author considers all contributors' notes alongside the
tested implementation and reviews, rather than treating any agent's proposal as
an instruction to rewrite the chapter.

The purpose is deliberate revision between editions: gather lessons freely while
the student works against stable exercises, then change the book using evidence.

CodeRhapsody's review categories are useful here: distinguish must-fix findings,
optional enrichment, **do not add**, and facts verified versus still uncertain.
Use them when they clarify a review; no extra report is required. Optional
enrichment must not silently become a requirement. For the author, check quoted
words, numbers, durations and commit identifiers against their actual sources,
not remembered summaries. Condense notes into teaching rather than exporting
private memory wholesale. Keep reasons and documented experience; old role
assignments, platform workarounds and historical completion claims in
CodeRhapsody's memory are not authority for this edition.

## How to add the next lesson

For each substantive discovery, record the chapter/revision, observation,
supporting command or artifact, interpretation, decision and remaining question.
Label a hypothesis as a hypothesis. When later evidence changes a conclusion,
record the correction and its reason rather than silently erasing the lesson.
Do not claim an experiment ran or a feature worked without the result.

Update the coding skill when Bill accepts a general operational rule; this
document supplies its motivation and history. Put local teaching feedback in the
shared chapter notes for the eventual author. At that later author phase,
preserve the human story and explain why the rules exist, using actual
experiences rather than invented battle scars.

On October 9, Bill generalized the back-pointer rule: an object with two actual
parents must retain both back-pointers, as in a family-tree node. The carryover
and coding skill now describe all actual parent relationships; a single-parent
chain is Ensemble's current example, not a restriction on the rule.

At the October 9 preparation checkpoint, the previous implementation agents
remain stopped, all discarded second-edition solutions and manuscript drafts
have been removed, and the carryover review team has read Chapters 0–23 and
the epilogue in full. Chapter 24 and the sub-agent chapter remain forthcoming.
The two drafts are ready for review with Bill. No new coder has started; the
new workflow's effectiveness remains to be tested.

Codex's own review on October 9 covered both drafts in full, their consistency
with Bill's instructions, source support for consequential carryover lessons,
links and scope. It clarified the coder's fresh-context boundary, small
save/load scope and the teaching-comment target. The carryover's remaining
source contradictions are explicitly left for discussion with Bill; neither
draft is marked approved.
