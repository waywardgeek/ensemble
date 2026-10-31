# Enhancement notes: Chapter 3

Author: `enhancement_author_resumed`, October 10, 2026. Initial prose review found
no blockers; grader validation is complete and final independent review pending.
Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

## Sources and story preservation

Read all 739 lines of original `book/chapter-03.md`, file revision
`e1605916114230570b348ab5961c193fa1cebf80`, SHA-256
`b83b9e989bb213bc3ac36f092ed80297fb996e5d64d9da1adb7af74bcbd4f3d8`.
Read the complete student/coordinator and comparative-review records, retained
live observations, accepted Agent/Ensemble/tools source and relevant tests from
`solutions/edition-2/ch03/` (`edition-2-ch03-r3`). Read the fake command's launch
behavior to separate the reference runner from an own-binary demonstration.
Full governing skill, carryover, workflow, enhancement plan, voice and chapter
procedure were read on resumption. No original or frozen source was changed.

The through-line remains the reader discovering that a small set of tools makes
the conversation useful, while apparently successful operations can hide the
most expensive mistakes. Preserve the full corpus-count opening and its explicit
private-data limitations; the fifty-one-row curve and seven-to-one editing ratio;
Bill's shell-only question; the quoting and portability failures; CodeRhapsody's
ambiguous-anchor bug; the thirty-to-one anchor/line contrast; fake generosity;
the wrapper exit-code trap; the unwanted codename file in the book repository;
and the synchronous freeze that motivates jobs. First-person CodeRhapsody prose
is explicitly attributed to the historical first-edition account, as in Chapter
2, under Bill's enhancement constraint. Historical counts and API anecdotes are
not new measurements or current model recommendations.

## Dispositions and changed examples

| Finding | Enhancement and reason |
| --- | --- |
| Missing argument keys stopped the student; prose requirements exceeded harness calls | Added the six-tool argument table after §3.3, including optional range/cap, append and search context. Small defaults are labeled accepted implementation choices, not extra exercise constraints. |
| Flat callback example lacked owner access | Extended the existing ToolFunc fragment with a narrow common ToolContext exposing Engine. Explained actual Agent owner, per-Agent declaration/dispatch set, simple ordered registry and behavior in tools. No copied service bag or invented parent. |
| Empty-registry omission confused with a CLI mode | Clarified ordinary library configuration and direct tests; no hidden command/switch. Retained unexpected-call error result even for an empty registry. |
| Student first score 55 from intermediate narration in the final answer | Explained complete ResponseEnded history versus final-only returned text. The later original 90/100 and inherited media failure remain separate. |
| Missing provider call IDs | Explained assigning deterministic distinct identities at capture, before dispatch, using the already accepted representation. No new ID protocol. |
| Overlapping edit anchors | Added `strings.Count("aaa", "aa")` example and actual correction, without requiring the student's refusal policy of every implementation. The original choice discussion and story remain. |
| Shell credential inheritance | Added the four supported key variables to remove from unrelated command environments; ordinary settings remain. Dummy-value testing and no confinement claim; no security framework. |
| First-edition sample commands ran the reference | Changed fake instructions to serve-only plus own built binary; live instructions use direct chat in scratch projects. Removed literal key-in-command example and model defaults presented as current advice. |
| “Where you launch” falsely implied confinement | Narrowed to relative-path starting directory; absolute/parent paths and shell access remain possible. Reviewer independently flagged the same inherited issue. |
| New live evidence | Added actual three-provider repair diff, recall and measured totals; additional search/write/read covers remaining tools. Preserved failed initial error requests and live-build timing before two narrow fixes. |
| Grader gaps | Agreed with coder: preserve all ten checks/weights, corrected Chapter 2 parity, add existing context-window/append coverage. Empty-registry and chosen edit-policy internals stay library/source evidence. |
| Mapping | Actor/mailbox remain new Chapter 5; beyond-pass compaction reference names the feature. External-consumer/custom-tool proof remains next Chapter 4 with old Chapter 5. |

The new live section is grounded in
`solutions/edition-2/evidence/ch03/live.txt` and the adjacent actual JSON logs:
16,271/863, 7,375/343, 8,467/563 input/output for the repair sessions, and
5,975/476 for the extra search/write/read session, all without cache tokens.
The exact typo diff and `Found Greeting\n` bytes are inspected artifacts, not
invented transcript text. Model names identify those dated runs only.

The changed Go fragment is shared vocabulary illustrating the actual owner
interface; it does not prescribe a larger registration API. The
[coder's completed validation](enhancement-03-code.md) confirms its agreement
with the accepted interface and the argument/default table. The changed
serve-only fake instructions were exercised with the independently built
student CLI in disposable projects: all three dialects exited zero after three
requests, with successful directory/read results and the planted module text.
These are deterministic launch checks, distinct from the retained live results.

Revised grading is accepted source **100/100 PASS**, legacy **96/100 FAIL**;
original grading remains accepted **90/100 FAIL**, legacy **100/100 PASS**.
The new search check exposes the legacy implementation's missing context
lines, an obligation already in the original prose. The author reconciled all
ten check names/weights and both added observations with the coder's final
report; no contract mismatch or runtime change was needed. Empty-registry and
chosen edit-policy checks remain source-level evidence. The capture fake's
zero-usage terminator applies the same inherited-fake lesson taught in §3.2.

## Checks and limits

`go run ./cmd/lintprose book/chapter-03.md book/edition-2/chapter-03.md` passes
all hard checks. Original prose count 5,902; enhanced 7,000. The inherited soft
person-gap warning remains (2,267 words original, 2,240 enhanced). A new pointer
closer exceeded the hard budget in the first draft; that sentence was revised
before submission. Original stories were not cut or padded to satisfy a counter.

Initial source-relative diff is approximately 215 added / 77 removed lines;
over 660 original lines remain. Most removals are misleading reference-launch
instructions, unsupported current model advice and their replacement with actual
student use. The title, section order, argument and original human incidents
remain. Independent review must judge teaching/enjoyment as well as lint.

No original chapter/grader, frozen student export or actual student result was
edited. No Chapter 4 draft, commit, mock or paid call was made. Await
independent review before claiming acceptance. At the reviewer's optional
editorial suggestion, moved the new overlapping-anchor paragraph after the
historical ambiguous-anchor incident; its original “That failure” antecedent
now remains adjacent and the two bugs are clearly distinct.
