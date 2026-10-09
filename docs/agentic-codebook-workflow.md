# Agentic codebook: workflow and learning record

Bill's goal is to establish whether an agent can follow the book as a student,
build an excellent AI coding agent framework from scratch, pass the original
graders, and demonstrate the features through a usable interface with real
models. Autonomous completion and an improved implementation are outcomes to
demonstrate, not claims established by the first attempt.

This is a living record. Bill has asked us to document everything we learn in
this process. Record substantive discoveries, failed approaches, corrections,
unresolved questions and evidence as they arise. Keep chapter-specific details
in the student's chapter review; bring lessons affecting the whole process here.
Do not wait for a successful chapter to record a failure or a scope problem.

## Sources of authority

- Bill defines the goal, required behavior and architectural constraints.
- The original first-edition chapter defines the current student exercise.
- [AGENTS.md](../AGENTS.md) and the mandatory
  [coding skill](../.agents/skills/ensemble-coding/SKILL.md) govern execution.
- The retained [second-edition Chapter 1](../book/edition-2/chapter-01.md)
  teaches the architecture improvements. It does not authorize extra features.
- The [first-attempt history](edition-2-attempt-1.md) records the retired work;
  its expanded contracts are not requirements for the restart.

Required behavior and architecture are binding. Suggested designs and
aspirational directions are provisional: the coder may reject them for a simpler
working approach and explain why. When actual requirements conflict or their
meaning changes user-visible behavior, ask Bill before implementing the affected
part. Neither the coordinator nor a reviewer may silently make that decision.

## How the agents work together

| Role | Responsibility and boundary |
| --- | --- |
| Bill | Resolve product and consequential architecture questions; supply direction and corrections. |
| Coordinator | Keep scope, handoffs and records coherent; route questions to Bill; stop drift and unnecessary work. Do not invent requirements. |
| Student coder | Read the original chapter and mandatory skill, implement from scratch or extend its accepted predecessor, run the original grader and live exercise, and leave a review for the future author. Do not consult old answers. |
| Code reviewer | After implementation, compare against the matching first-edition solution and read the coder's review. Enforce KISS, architecture and actual scope; return findings and reasons, not old solution code. |
| Future author | Begin only after the complete implementation succeeds and comparative reviews establish its improvements. Revise chapters from the demonstrated results and student experience. |

There is no author or grader agent working ahead of the student. Suspected
grader defects are recorded for a future maintainer; the student does not change
the grader to pass. Agents may raise questions through the coordinator, who must
relay consequential questions to Bill rather than answer on his behalf. Autonomy
does not remove the need for collaboration when the assignment is ambiguous.

For each chapter:

1. Give the coder the original chapter and exact mandatory skill path. Keep old
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

The first-attempt history links these observations to preserved material. Bill
reported substantial subscription usage; exact cost attribution was not
measured. The repository evidence establishes scope expansion, not a diagnosis
of which training data or internal model mechanism caused it.

## How to add the next lesson

For each substantive discovery, record the chapter/revision, observation,
supporting command or artifact, interpretation, decision and remaining question.
Label a hypothesis as a hypothesis. When later evidence changes a conclusion,
record the correction and its reason rather than silently erasing the lesson.
Do not claim an experiment ran or a feature worked without the result.

Update the coding skill when Bill accepts a general operational rule; this
document supplies its motivation and history. Put local teaching feedback in the
chapter review for the eventual author. At that later author phase, preserve the
human story and explain why the rules exist, using actual experiences rather
than invented battle scars.

At this restart checkpoint, the previous agents are stopped and the discarded
solutions are deleted. Creating this record does not restart them. The new
workflow's effectiveness remains to be tested.
