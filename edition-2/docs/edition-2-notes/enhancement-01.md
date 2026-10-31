# Enhancement notes: Chapters 0–1

Author: `enhancement_author`, October 10, 2026. Initial draft for independent
review; not accepted. Handoff skill: `.agents/skills/ensemble-coding/SKILL.md`.

## Source and preservation

Started from the complete current originals, never the deleted Edition 2 draft.
Chapter 0 file revision: `ec41c6e8cf0800e266f7a39ebadf255d94379082`, SHA-256
`1f7ba69fcca58eb1f41423fcb81928854c014cdbf72712fe7b768e80ac1a7244`.
Chapter 1 file revision: `e1605916114230570b348ab5961c193fa1cebf80`, SHA-256
`d22555ddf1cdb40b0ce97733e6550641a63b465ece044d5eb65b6b6d2f2dbce8`.
Author read AGENTS, enhancement plan, full current coding skill, voice, chapter
procedure, carryover, workflow, ch01 student/comparative reviews, accepted ch01
runtime and live evidence, and original §5.9 logger ambush. First-edition prose,
graders and student exports were not edited.

Chapter 0's stake remains Bill's failed attempt to transfer his design judgment
by supplying source alone, resolving in teaching that preserves reasons. Kept
its opening, comparison, steering motivation, specification/receipt argument,
self-wielding explanation and invitation to fork. Product figures are explicitly
historical first-edition claims, not new measurements or independently verified
comparisons. Narrowly qualified promises of automatic 100s and guaranteed
improvement because actual student results contradict them.

Chapter 1 retains the full Windsurf/StackAgent two-week bet and restart, hint
discovery sidebar, model-discovery failure, deterministic usage curve,
first-edition live conversation, Waywardgeek prompt story, price discussion and
seven-check exercise. Its stake remains getting from that bet to a conversation
the reader built. Added the original logger ambush's concrete 5,047-line detail
as motivation for owners before code; the old per-Agent logger location is
superseded by Bill's Ensemble ownership ruling. No standalone refactor restored.

## Findings incorporated and deferred

| Evidence | Disposition |
| --- | --- |
| Carryover: ownership, star imports, common vocabulary versus behavior, required/optional parents, no globals, comment quality | New plain-words subsection before the first code block, using only Ensemble, Agent and Engine. Observer/GUI named as future boundaries without early implementation. |
| Workflow: attempt before reference exposure; real UI and grader are different; targeted mutations; meaningful comparison; bounded review; durable reasons | Chapter 0 execution section, with short cycle and failure/exception distinctions. No internal assignment ledger or expanded acceptance framework copied into teaching. |
| Retired attempt: author expanded exercises and review reinforced it | One sourced Chapter 0 paragraph explains why this pass enhances already exercised teaching. |
| ch01 student: original loop sufficient but architecture arrived outside chapter | Chapter 1 now teaches reasons for the existing owner graph; exercise feature scope unchanged. |
| ch01 student: strict alternation is narrower than full provider contract | Wording and Message comment now identify it as this exercise's rule. |
| ch01 student/reviewer: staging failed turn, actual usage, secret-free diagnostics | Agent-owned Ask snippet stages the pair; short explanation of accepted Engine checks. No retries, recovery subsystem or new scored matrix. |
| ch01 student/reviewer: version-header mutant survived original grader | Wire contract explicitly states exact version; separate corrected grader command agreed with coder. Historical score retained. |
| ch01 student: missing CH variable; launcher consumed PTY input | Correct original command with CH=1; new Edition 2 invocation; short live note shows why clean exit did not prove conversation. |
| ch01 live evidence | Preserved old transcript labelled as first-edition run; added actual October 9 model/totals and Copper Finch observation linked to retained receipt. No new paid run claimed. |
| Carryover feature-specific lessons | Deferred to their respective chapters. No jobs, tools, actor, persistence or future orchestration added here. |
| Crossover destination | Chapter 0 states Bill's GPT-5 Astra/high-reasoning and Edition 3 goal as future verification, not achieved outcome or API identifier. |

## Technical changes and checks

Changed only the Ask example's ownership/staging, usage-owner names, one Message
comment and grader invocation/code comments. Request/response JSON, token table,
historical chat transcript and system prompts are preserved. The parsing excerpt
remains a shortened explanation; the surrounding prose now identifies extra
accepted checks rather than implying it is the whole Engine.

Coder confirmed new command from repository root:
`go run ./cmd/grade-edition2 -ch 1 path/to/your/solution`, optional `-json` before
path. Seven original checks/100 points retained, exact version under `wire`.
Coder owns grader validation; this author does not claim its pending results.

Initial source comparison: Chapter 0 81 added/7 removed lines; Chapter 1 147
added/34 removed (the large addition is architecture before code). Original
text remains the bulk of both chapters. Existing titles and section order stay.

Prose check: `go run ./cmd/lintprose` on the two enhanced files, compared with the
same command on the originals. Chapter 1 has no hard failures. Chapter 0 retains
exactly the original hard findings: four pointer closers (ceiling two) and one
em-dash counted from its table. These are inherited and preserved under the
enhancement constraint, not newly introduced or a claimed lint pass. Enhanced
prose words: 2,609 (ch0), 4,513 (ch1, after logger detail); originals: 1,937/3,475.
Soft person-gap/short-closer warnings remain reading-review matters: the new
execution and ownership passages address the reader directly, and have not been
padded with invented anecdotes or Bill mentions to satisfy the counter.

The original first-person Chapter 0 and its broad historical claims were not
rewritten to satisfy the later voice guide. This is a deliberate preservation
decision under Bill's explicit enhancement instruction. Reviewer should judge
whether the targeted additions remain useful and readable alongside that voice.

## Review correction round 1

The independent reviewer found no blocking preservation/teaching issue, then
identified an inherited command-status ambiguity: `go run` does not propagate
the grader binary's exit 2 as its own shell status. Chapter 1 now distinguishes
the binary's 0/1/2 statuses from the wrapper's status and printed diagnostic.
No exercise, example payload or runtime behavior changed.
