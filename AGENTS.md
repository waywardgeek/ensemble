# Ensemble student restart

Bill retired the author-driven second-edition attempt. Its history is preserved
at `edition-2-attempt-1-stopped`; do not resume its agents or workflow. See
`docs/edition-2-attempt-1.md` for the decision.

Before coding or reviewing code, and after compaction, read the entire skill:
`.agents/skills/ensemble-coding/SKILL.md`. Every coder/reviewer handoff must include
that exact path. It replaces the deleted second-edition workflow skill.

The student follows first-edition `book/chapter-NN.md`, in order from Chapter 1,
with Bill's architecture rules applied from the beginning. The retained
`book/edition-2/chapter-01.md` supplies architecture teaching, not authority to
expand the original exercise. There is no author agent. Use the original grader;
record suspected grader defects and teaching gaps for future work rather than
changing chapters or graders during the student attempt.

Only after the complete new Ensemble implementation succeeds and its comparative
code reviews establish the improvements does the author phase begin. The author
then works chapter by chapter from the original book, the coder's reviews and
the reviewed implementation. Incorporate demonstrated lessons into the second
edition; preserve the first edition as history. Do not draft ahead of the coder
or turn editorial work into new feature requirements.

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
