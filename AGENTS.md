# Ensemble student restart

Bill's explicit crossover goal: the coder's new Ensemble must run GPT-5 Astra
at high reasoning, and Edition 3 of the book is to be built using that Ensemble
agent rather than Codex. Keep this destination in view without implementing
later features early. See the workflow's crossover goal and verification notes.

Bill cannot use Codex for continuous real-time collaboration. Do not rely on him
monitoring progress or interrupting a mistaken direction. Work autonomously
within agreed scope. For a critical question, stop the work, pause active
delegates, ask plainly with the relevant context, and wait for Bill's answer.
Do not ask asynchronously and continue working, or treat silence as agreement.
Close real-time collaboration in Ensemble is part of the crossover goal.

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

Start from empty in `solutions/edition-2/main/` when coding resumes. The student
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
