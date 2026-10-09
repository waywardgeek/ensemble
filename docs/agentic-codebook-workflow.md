# Agentic codebook: workflow and learning record

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
successive versions on Amazon as he adds chapters. He reports that the sub-agent
chapter has not yet been written. It will encode his experience with sub-agents
and supply the exercise for that capability. Do not treat the current chapter
count as final or invent a replacement assignment. Record the source revision
used for each exercise; when new or revised chapters arrive, read them and update
the carryover and affected chapter notes before applying their lessons. Existing
checkpoints remain evidence of the versions actually followed.

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
  original Chapter 1 stays unchanged; the retained second-edition draft is history.
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

Read the entire current edition before launching its successor's student run.
The coordinator can do this preparation without an author agent. Extract the
lessons learned during that edition into a compact carryover document, giving
each lesson a source, reason and point of application. Record the source revision
and reading coverage; a partial review must not be described as complete.

Distinguish corrections and Bill's explicit decisions from tentative advice.
Keep future features in their own exercises, exclude old solution code, and
resolve consequential contradictions with Bill. This preparation teaches known
lessons without rewriting chapters or designing the next implementation ahead
of experience. Students read the carryover and skill with the original chapters;
reviewers check that the carryover has not become a new source of feature creep.

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
| Student coder | Read the carryover, original chapter and mandatory skill, implement from scratch or extend its accepted predecessor, run the original grader and live exercise, and leave a review for the future author. Do not consult old answers. |
| Code reviewer | After implementation, compare against the matching first-edition solution and read the coder's review. Enforce KISS, architecture and actual scope; return findings and reasons, not old solution code. |
| Future author | Begin only after the complete implementation succeeds and comparative reviews establish its improvements. Revise chapters from the demonstrated results and student experience. |

There is no author or grader agent working ahead of the student. Suspected
grader defects are recorded for a future maintainer; the student does not change
the grader to pass. Agents may raise questions through the coordinator, who must
relay consequential questions to Bill rather than answer on his behalf. Autonomy
does not remove the need for collaboration when the assignment is ambiguous.

For each chapter:

1. Give the coder the carryover, original chapter and exact mandatory skill path. Keep old
   solutions and discarded requirements out of the coder's handoff.
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
coding resumes. Chapter snapshots are exports, not competing working trees.
Carry corrections forward and revalidate affected behavior. Use concise reviews
and sanitized test observations rather than building a new evidence system.
Bound live runs and retries, clean up large generated test data, and never put
credentials in the records. Do not repeat unchanged successful checks without a
reason; investigate persistent failures instead of spending indefinitely.

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
under Bill's authorization above. Keep old
solution code and answer-revealing comparisons out of notes supplied to the
student; the reviewer returns findings and rationale through the existing review
process. The future author considers all contributors' notes alongside the
tested implementation and reviews, rather than treating any agent's proposal as
an instruction to rewrite the chapter.

The purpose is deliberate revision between editions: gather lessons freely while
the student works against stable exercises, then change the book using evidence.

## How to add the next lesson

For each substantive discovery, record the chapter/revision, observation,
supporting command or artifact, interpretation, decision and remaining question.
Label a hypothesis as a hypothesis. When later evidence changes a conclusion,
record the correction and its reason rather than silently erasing the lesson.
Do not claim an experiment ran or a feature worked without the result.

Update the coding skill when Bill accepts a general operational rule; this
document supplies its motivation and history. Put local teaching feedback in the
shared chapter notes for the eventual author. At that later author phase,
preserve the human story and explain why the rules exist, using actual experiences rather
than invented battle scars.

At this restart checkpoint, the previous agents are stopped and the discarded
solutions are deleted. Creating this record does not restart them. The new
workflow's effectiveness remains to be tested.
