# Ensemble student restart

Bill's explicit crossover goal: the coder's new Ensemble must run GPT-5 Astra
at high reasoning, and Edition 3 of the book is to be built using that Ensemble
agent rather than Codex. Keep this destination in view without implementing
later features early. See the workflow's crossover goal and verification notes.

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
Bill authorizes a specific exception: when the coder documents why an exercise
requirement is unreasonable and the reviewer independently agrees, the reviewer
may pass the chapter with a recorded exception and advance the coder without
asking again. Record the actual grader/live results and what was waived for the
future author; do not describe a failing grader as passing. This does not waive
Bill's explicit architecture rules or authorize extra features.

Before coding or reviewing code, and after compaction, read the entire skill:
`.agents/skills/ensemble-coding/SKILL.md`. Every coder/reviewer handoff must include
that exact path. It replaces the deleted second-edition workflow skill.

Before the student run, the coordinator reads the entire current edition and
completes `docs/edition-2-carryover.md` with sourced lessons and recorded reading
coverage. This is preparation, not chapter authorship or new feature design.
The student reads that carryover alongside the mandatory skill and follows
first-edition `book/chapter-NN.md`, in order from the unchanged Chapter 1.
Bill's architecture rules apply from the beginning. The retained second-edition
Chapter 1 draft is historical material, not required student reading.
There is no author agent. Use the original grader;
record suspected grader defects and teaching gaps for future work rather than
changing chapters or graders during the student attempt.

Only after the complete new Ensemble implementation succeeds and its comparative
code reviews establish the improvements does the author phase begin. The author
then works chapter by chapter from the original book, all agents' chapter notes,
the coder's reviews and the reviewed implementation. Incorporate demonstrated
lessons into the second edition; preserve the first edition as history. Do not
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
reviewer preferences for the original exercise. The coder may improve its answer
and rerun affected checks before the chapter checkpoint.

Feature creep and code bloat violate the task. Ask Bill about consequential
ambiguity; coordinator preferences are not user requirements. In particular,
pause is Agent-wide: an explicit unpause in any tab unpauses it for everyone.
