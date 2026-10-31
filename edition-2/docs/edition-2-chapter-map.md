# Edition 2 chapter mapping

## October 14: background memory follows the fleet

Bill approved this dependency reorder. Current manuscript filenames, section
numbers, chapter index and revised grader entry points follow this table.
Chapters0–14 and24–26 retain their numbers. The older dated mappings below
remain the record of their original checkpoints.

| Previous Edition 2 | Current chapter | Current exercise command |
| --- | --- | --- |
| 17 | [15: The Invisible Invoice](../book/chapter-15.md) | `grade-edition2 -ch 15` |
| 18 | [16: Leaving Anthropic for OpenAI](../book/chapter-16.md) | `grade-edition2 -ch 16` |
| 19 | [17: The Crossover](../book/chapter-17.md) | Ungraded field report |
| 20 | [18: Giving the Agent the Web, and a Status Line](../book/chapter-18.md) | `grade-edition2 -ch 18` |
| 21 | [19: The Sandbox](../book/chapter-19.md) | `grade-edition2 -ch 19` |
| 22 | [20: The Ensemble](../book/chapter-20.md) | `grade-edition2 -ch 20` |
| 23 | [21: Managing the Swarm](../book/chapter-21.md) | `grade-edition2 -ch 21` |
| 15 | [22: Forgetting on Purpose](../book/chapter-22.md) | `grade-edition2 -ch 22` |
| 16 | [23: Auto-Recall](../book/chapter-23.md) | `grade-edition2 -ch 23` |

Invoice now extends Chapter14. Memory22 extends fleet management21, then
Auto-Recall23 adds its stateless relevance judge. Auxiliary accounting and
cache-lens exclusions enter with those later features; the earlier accounting
chapters do not require a compressor or judge before they exist.

Current grader defaults use `solutions/chNN` under the new numbering. The
existing cumulative exports and manifests are still the old-order historical
source at this checkpoint; they are not made valid by renaming. Use an explicit
submission path while separately reviewed reconstructed exports are prepared.
The previous accepted trees and immutable tags remain historical evidence.
See the [renumbering checks](edition-2-notes/codex-compaction-renumber.md).

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

## October 12: Gateway enhancement accepted; publication preparation released

Independent combined review `63709086` accepts author `35930f10` and coder
`cf86ea9e`, at initial review with zero formal correction rounds. Checkpoint:
`edition-2-enhanced-ch26-r1`. Final actual seven-row integration runs score
100/100 on accepted and legacy source; earlier CLI runs retain their own source
versions. The reviewer found a child-registration refusal that still granted
access. Its exact initially surviving defect now loses only C7, scoring80/100.
All14 executed fences and four final grader hashes are verified; original0/100
and the actual11.739-second interview retain their original provenance.
See [the independent review](edition-2-notes/enhancement-26-review.md).

Implementation and manuscript enhancement now cover the first edition through
its final Gateway exercise. Optional24 is complete; crossover25 remains reserved.
Proceed with Bill's authorized archive/promotion and publication preparation in
[the relocation record](edition-2-publication.md). The frontmatter draft is ready;
its review, the new coordinator epilogue and final artifact checks remain.
Publication and a completed crossover are not claimed by this checkpoint.

## October 12: Gateway accepted as Chapter 26

Original27 implementation is independently accepted at `0afc45e8`, source
`b5c35a39`, with the narrow documented grader exception. Direct accepted
`solutions/edition-2/ch26` and its manifest contain the exact197-file source;
checkpoint `edition-2-ch27-r1`. Chapter26 prose/grader enhancement now follows.
Implementation catch-up through original27 is complete. Optional24 is done and
crossover25 remains a placeholder, without an accepted solution export.
Earlier pending statements below retain their historical dates.

## October 11: resume the book; reserve crossover 25; finish Gateway 26

Chapter 24 is the completed optional Codex workaround. Chapter 25 is reserved
for Bill's live crossover, with a brief placeholder only. The book team resumes
at Chapter 16 and completes enhancement through 23, then implements and teaches
original Chapter 27, Gateway, as **Edition 2 Chapter 26**. Edition 1 ends at 27.
Earlier pause and deferred-placement statements below record the prior decision.
Preserve the concurrent crossover changes; no accepted ch25/ch26 export exists yet.

## Edition 1 endpoint and new Edition 2 material, October 11

Bill freezes Edition 1 at Chapter 27. On October 11 he placed the new optional
Codex-endpoint workaround next, as **Edition 2 Chapter 24**, immediately after
the accepted fleet chapter. The coordinator's earlier reservation of 24 for
Gateway and 25 for this workaround was superseded. Original 27 (Gateway) has
not been implemented in Edition 2; its placement is deferred, not silently
claimed complete or folded into this chapter.

The new chapter is short if that serves the material, optional, and ungraded.
Readers whose public OAuth route caches satisfactorily may skip it. Bill assigns
the coordinator authorship, implementation and self-review for this chapter;
the ordinary book team stays paused after Chapter 16. Code is developed in
`solutions/edition-2/main/` and exported as `solutions/edition-2/ch24/` after
validation. Subsequent new chapters are Edition 2 only. Preserve Edition 1.

Bill authorized the coordinator on October 11, 2026 to choose both remaining
merges and proceed autonomously. Exactly three first-edition standalone chapters
are absorbed: 5, 22 and 24. This supersedes the earlier placement stops.

| First edition | Second edition | Direct accepted solution | Placement |
| --- | --- | --- | --- |
| 0–3 | 0–3 | `ch01`–`ch03` (no Chapter 0 implementation) | Unchanged |
| 4 + 5 | 4 | `solutions/edition-2/ch04` | Jobs and public framework consumer |
| 6–20 | 5–19 | `solutions/edition-2/ch05`–`ch19` | Existing mapping |
| 21 + 22 | 20 | `solutions/edition-2/ch20` | Web access, status and accounting |
| 23 | 21 | `solutions/edition-2/ch21` | The Sandbox |
| 24 + 25 | 22 | `solutions/edition-2/ch22` | The Ensemble, including reusable GUI boundary |
| 26 | 23 | `solutions/edition-2/ch23` | Fleet management and GUI; accepted |
| New in Edition 2 | 24 | `solutions/edition-2/ch24` | Optional Codex endpoint; self-reviewed under Bill's instruction |
| New in Edition 2 | 25 | Pending; no accepted export | Codex-to-Ensemble crossover; placeholder only |
| 27 | 26 | `solutions/edition-2/ch26` | Gateway implementation accepted; manuscript enhancement underway |

With first-edition Chapters 1–26 complete, Edition 2 has Chapters 1–23. Chapter
0 remains introductory in both. The optional new Chapter 24 has no first-edition counterpart; future placements
follow their actual assignments rather than an automatic offset. Do not create standalone Edition 2 refactor
chapters.

## Why these merges

Original Chapter 22 completes original Chapter 21 before sandboxing: the agent
gains status access through its existing owners, accurate per-model accounting
and a practical model-switch exercise. Most architectural repair lessons already
govern the student from the beginning; do not recreate their old defects.
Preserve the status feature and structural/runtime verification, including real
executable composition and observer-disconnect behavior. Keep the sandbox
chapter focused on confinement rather than inserting an unrelated architecture
detour.

Original Chapter 24 belongs with original Chapter 25 because a reusable frontend
boundary and public event vocabulary support application assembly around the
agent fleet. Preserve embedded deployment with disk override, component-only
reuse, public MCP relay, external consumers, transitive dependency controls and
unchanged wire behavior. The GUI is already a separate optional module; do not
copy the source's callback bundle or reintroduce coupling just to remove it.
Original Chapter 25 remains the terminal fleet exercise; multi-agent GUI
features wait for original Chapter 26. The later prose should use coherent
sections within the combined chapter, with the right teaching length rather than
a target word reduction.

## Execution and preservation

Current checkpoint: original26 is independently accepted at `0d9e823` after
one correction round, source `450f7a9`, main tree
`27d7fcf9b8b31275068fa705d6969cfe3eafedc1`. Its exact direct `ch23` export is
checkpointed as `edition-2-ch26-r1`. Direct chapters1–23 are accepted and match
the table. Original26 was complete at original25 review, satisfying the condition.
Original24's intermediate tag, the combined24+25 `ch22`, and prior tags remain.
Original results and accepted exceptions are preserved; later prose/grader
enhancement remains separate from these completed implementations.

Use original exercises and unchanged graders in original order: 22, 23, 24, 25.
Each original exercise still receives its own student review, actual grader/live
results, independent comparative review and bounded correction loop. Notes,
evidence and original-numbered tags keep original numbers. Only the book-facing
chapter/solution numbering follows the table above.

Extend `solutions/edition-2/main` throughout. After original Chapter 22
acceptance, replace the direct `ch20` export with the newly reviewed cumulative
original 21 + 22 source and a new manifest/tag. The prior original-21-only
export and manifest remain exactly preserved in immutable tag
`edition-2-ch21-r1`; no existing tag is moved and no new history directory or
symlink is created. This specific replacement is authorized by Bill's requested
merge. Original Chapter 23 acceptance creates direct `ch21`. Review and
checkpoint original Chapter 24 as an intermediate stage in Git, then original
Chapter 25 acceptance creates the combined direct `ch22`. Never publish an
incomplete combined exercise as an accepted final chapter. Original Chapter 26
acceptance, if reached, creates `ch23`.

At original Chapter 25 review completion, inspect original Chapter 26's actual
manuscript, grader, reference snapshot and completion records. A snapshot alone
does not establish that the chapter is complete. If completed, read the full
source, update carryover coverage and proceed autonomously with original Chapter
26→Edition 2 Chapter 23. If still ongoing, checkpoint through `ch22` and report
that condition without inventing requirements. Preserve CodeRhapsody's commits,
index entries and working files at every step.

GPT-6 Astra/high remains mandatory for coding and code review. The complete
implementation precedes the later author–coder manuscript/grader enhancement
phase; this map directs that phase without starting an author ahead of the
coder. All inherited architecture, no-mocks, evidence and safe shared-commit
rules remain.
