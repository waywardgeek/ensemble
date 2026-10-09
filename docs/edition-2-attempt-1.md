# Second-edition first attempt: preserved history

Bill ended this attempt after reviewing the expansion of the persistence
chapter and tracing invented product requirements back through earlier chapters.
The discarded working trees were cleared by explicit deletion in commit
`751faf6`, not by resetting or rewriting Git history. All second-edition solutions
and all second-edition book material except Chapter 1 were deleted. The first
edition and its solutions remain unchanged.

The annotated tag `edition-2-attempt-1-stopped` preserves the final manuscript,
completed snapshots, and unfinished Chapter 11 source before deletion. Chapter
11 was interrupted, not validated. Earlier chapter tags remain historical
checkpoints, not Bill's editorial approval. The last preceding commit was
`44b17c06e6752652df4d10a700e05d075c9b70eb`.

The archive includes tracked evidence and the unfinished coder's text/source
files. Untracked generated executables and ignored caches are not added to Git;
they are removed with the discarded working tree. External historical student
directories and process transcripts are outside this deletion.

## What went wrong

The author/coordinator workflow replaced the original exercises with expanded
contracts. Reviews checked compliance with those contracts without adequately
challenging their scope. For persistence, the first edition added 254 physical
production Go lines; this attempt added 4,585, excluding tests. Custom canonical
JSON, imported origins, extensive semantic validation and acceptance machinery
turned a small save/load lesson into a much larger subsystem.

Per-tab pause registrations were another product change, already required in
the initial Chapter 7 draft and approved by the coordinator before coding.
Bill clarified that pause is Agent-wide: explicitly unpausing in any tab must
unpause the Agent for everyone. Architecture did not require per-client vetoes.

Bill reported substantial time and subscription usage spent on this attempt.
No exact subscription-cost attribution was measured in this review. The lesson
is scope discipline, not that passing tests proves a better design.

## Restart decision

There is no author agent in the restart. The student follows the original
first-edition chapters and passes their original graders, while applying Bill's
architecture rules from the beginning. It leaves concise teaching/grader reviews
for future authors and grader maintainers. It does not rewrite the assignment,
repair graders, copy old solutions, or implement speculative features.

The current coding instructions are in `AGENTS.md` and
`.agents/skills/ensemble-coding/SKILL.md`. The retained
`book/edition-2/chapter-01.md` preserves the earlier architecture teaching;
the original `book/chapter-01.md` defines the first exercise. Later first-edition
architecture-repair chapters do not require deliberately undoing correct design.

The ongoing [workflow and learning record](agentic-codebook-workflow.md) describes
the agent roles, guard rails and how to preserve subsequent lessons.

To inspect discarded material, use
`git show edition-2-attempt-1-stopped:PATH`. Existing commits and annotated chapter tags
remain intact. Do not resume the stopped author or unfinished Chapter 11 coder.
