# Original Chapter 11 independent code review

Reviewer: `enhancement_reviewer_resumed`, October 10, 2026.
Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## Initial review — correction required

Read the full original Chapter 11 and student submission in [ch11.md](ch11.md),
mandatory skill and current carryover/workflow. Compared the submitted runtime
with accepted original Chapter 10 in `solutions/edition-2/history/ch10`, then
read the matching first-edition Chapter 11 persistence implementation and its
changes from `solutions/ch10`. Source chapter remains revision `994bf218b`,
SHA-256 `ed6482dd1e5247880122382a3b4199276d8fbda13ea8863676eda5c6e49e9df0`.
Implementation was checkpointed as `4e6cc3f`; the subsequent anonymous-field
documentation additions are included in this source assessment. No original
chapter, grader or answer was edited.

**One blocking finding:** `internal/history/save.go` records the raw
`Config.SystemPrompt`. Without a skill or explicit instruction this is empty,
while `internal/engine/engine.go` selects its nonempty default instruction for
the vendor request. Thus an ordinary human CLI session saves a false record of
what shaped the wire, contrary to Chapter 11 §11.2 and the new `SaveConfig`
documentation. The original check's supplied primary skill hides this case.

An independent scratch control extended the actual CLI lifecycle test to compare
the saved instruction with its captured HTTP request. Both default and overridden
save paths failed with saved `""` versus the full default instruction. The
[receipt](../../solutions/evidence/ch11/reviewer-system-record-control.txt)
also preserves the control's first printf-type/vet setup error, which is not
counted as a behavioral failure. Correct the effective metadata through existing
ownership, without duplicated literals, behavior in common, restored historical
configuration, or loss of intentionally empty primary instructions. Add a narrow
wire-based regression and sensitivity evidence; no paid replay is needed.

No other implementation blocker was found. This completes the initial review;
the coder may now use **correction round 1 of at most 3**. Final acceptance waits
for the correction, affected verification and reconciled student record.

## Actual grading and independent execution

The original full invocation is **interrupted / unscored**, exit 2. Its launch
preceded the GUI stdin-envelope correction; its partial output must not be
treated as a final-source pass. The coordinator bounded the previously established
recursive GUI/headless mismatch and the coder released the harness after cleanup.

The reviewer copied unchanged grader dependencies to scratch and invoked the
existing exported `Ch11Run` and `Ch11Checks`. Recursive Chapter 10 parity is
explicitly unrun. The [bridge](../../solutions/evidence/ch11/reviewer-original-bridge.go.txt),
[exact source manifest](../../solutions/evidence/ch11/reviewer-initial-source.json),
[raw output](../../solutions/evidence/ch11/reviewer-initial-original.txt)
and [result](../../solutions/evidence/ch11/reviewer-initial-original.json)
are retained. The wrapper completed in 16.49 seconds and exited 0 as an evidence
collector, not as an all-checks-pass assertion.

Current-chapter checks earned **70/85**. Default load, snapshot/replay equality,
older snapshot with tail, snapshot-only restoration/continued numbering and
corrupt-save refusal pass **70/70**. Save-shape fails **0/15**, reporting missing
`think` in saved tool declarations. This is the same undefined capability accepted
as a narrowly documented exception in Chapter 10, not a requirement to invent or
falsely advertise a handler. Its provenance remains applicable here.

However, save-shape returns at that missing name **before** its later log-order
and anchor assertions. Those later predicates did not pass in this run. Source
tests and their targeted mutations separately exercise numbering, retained log
and anchor; final review will retain the distinction. Neither the skipped parity
marker nor 70/85 may be reported as a completed original 70/100 grade. The earlier
first-edition 100/100 baseline is retained historical evidence, not a new rerun.

## Code, comments and tests

The four-field save remains small. Agent owns the selected path and public
load/save lifecycle; History owns snapshot/anchor/log restoration and uses its
existing reducer. Common adds only typed vocabulary. Restore stages its state
before installation, keeps a snapshot-only anchor, validates ordered facts and
does not restore historical configuration. Replacement writes close a complete
temporary file before rename. Both application clients install saving only after
successful load, join jobs before saving and use the public root/Agent. Explicit
application-selected cwd paths reconcile the exercise with per-Agent library
defaults; there is no global save owner. No journal, canonical disk format,
compaction, memory or skill-runtime restoration was added.

Independent counts exclude blanks and `//`-only lines from active Go; inline
comments remain active code, and physical counts include everything. No source
moved or was generated. Counts below include the final initial-submission field
comments and exclude unchanged examples.

| Go category | Own Chapter 10 active / comments / physical | Submitted Chapter 11 active / comments / physical |
| --- | --- | --- |
| Student production, including optional GUI | 4,975 / 1,135 / 6,482 | 5,158 / 1,174 / 6,714 |
| Student tests | 3,180 / 237 / 3,518 | 3,446 / 252 / 3,807 |
| First-edition production | 6,573 / 1,999 / 9,475 | 6,725 / 2,087 / 9,729 |
| First-edition tests | 769 / 160 / 1,012 | 769 / 160 / 1,012 |

The student adds 183 active production lines versus the original's 152; student
tests add 266 versus zero. The student's three external examples remain
415 active / 37 comments / 469 physical lines. JavaScript, CSS, HTML and the
original shell helper are unchanged within each edition. Evidence launchers are
support, not production savings. The modest growth provides separate CLI/GUI
entry wiring, owned library operations and staged restore, rather than the
original's command/common-package behavior and optional `verify` command. No
bloat finding accompanies the metadata correction.

The addition has 39 comment lines, about 17.6% of nonblank growth. Comments
explain anchors, ownership, transactional restore, replacement order and EOF
lifetime. Newly exported fields/methods, including the self-corrected anonymous
input/test fields, have useful documentation. The original contains more
explanation but also a stale flag-only-save comment. No comment-padding quota
or arbitrary source-size target is imposed.

The new tests exercise the real reducer/renderers, library ownership and compiled
CLI through a small external HTTP fake. No mocks or substituted core machinery
were introduced. Read all 18 intended mutation results and working controls:
they catch the recorded snapshot/tail/numbering/config/path/refusal defects for
the stated reasons. These do not cover the newly found effective-prompt case.
All five module test/vet passes are reported by the coder; independent affected
execution will follow the correction instead of repeating an unchanged suite now.

Read the three actual live artifacts and launcher: all three vendors recall the
word in a second human-chat process; Anthropic also recalls from independent
snapshot-only and log-only variants. The corrupt variant exits 1, makes no request
and preserves bytes. Eight successful model-response processes plus that refusal
are the actual nine-process result. Models were reused from same-day discovery,
including Gemini 3.8 Flash. Secrets remain in the approved child environment and
are redacted before artifact retention. No new live run is needed to repair
historical metadata that is deliberately not restored into runtime configuration.

## Correction round 1 — accepted with recorded exceptions

The initial blocker is resolved. Read the entire revised student record and
correction diff saved in `3f021a0`; independently verified all thirteen source
hashes against aggregate
`a8ad2cec304e81b68a8a71e553c84ac0f33cb487580efca8aa263b4cf2490ed9`.
Engine now exposes its existing effective-instruction computation through the
documented common interface. History reaches it through its Agent parent. The
three renderers use that same computation; no duplicate default or second policy
was added. Runtime configuration still wins on restore, and an explicitly empty
primary body still suppresses the legacy default.

The new real-CLI assertion compares the saved instruction with captured HTTP
bytes, while the separate empty-primary test checks actual rendering and saved
metadata. Both restored-defect mutations fail their intended assertions, with
passing controls before and after. The original eighteen mutation records remain
distinct from these two correction checks.

Independent final execution in a byte-matched scratch copy: persistence and
empty-primary library tests under `-race` PASS (1.057s), actual CLI lifecycle
with additional save-shape inspection PASS (0.237s), and core vet PASS. Retained
[check results](../../solutions/evidence/ch11/reviewer-round1-checks.json),
[source hashes](../../solutions/evidence/ch11/reviewer-round1-source.json),
and actual first-session saves for
[default](../../solutions/evidence/ch11/reviewer-default-save.json) and
[overridden](../../solutions/evidence/ch11/reviewer-override-save.json)
paths show exactly the four fields, nonempty effective default instruction,
real tool declarations, three strictly ascending positive event sequences and
anchor 3 equal to the final event. Snapshot/tail tests separately check retained
loaded events and next-save anchor after appending. These observations cover
the valid shape obligations masked by `think`; they do not turn the original
save-shape check into a pass. Coder core/GUI test/vet reruns also pass.

Final student deltas are +184 active / +41 comment-only / +235 physical
production lines and +304 active / +21 comment-only / +334 physical test lines.
The initial comparison is preserved above. The small accessor and specific
regressions resolve the bug without expanding the exercise; comments remain
useful and all new exports documented. No further implementation finding remains.

**Accepted with exceptions after one correction round.** The missing-`think`
expectation remains an independently justified continuation of Chapter 10's
undefined-tool exception; advertising a fake handler is not a persistence fix.
The recursive GUI/headless launch mismatch remains the bounded inherited-harness
exception. Neither exception waives actual persistence behavior. The full
original command remains interrupted/unscored, and the measured focused result
remains 70/85 on the initial submission, with its early-return limitation stated
above. Round 1 changed effective metadata for the no-skill case; the supplied
primary used by that focused run already rendered correctly. Its unchanged
behavioral evidence is reused with the independent final correction checks,
not presented as another original run or an invented aggregate score.

All three vendors' actual human persistence receipts remain applicable; no paid
repeat or recursive grader sweep was necessary. Scratch copied source, test
bridge and mutation-control trees were removed after retaining these small
receipts. The coordinator owns the accepted checkpoint and next assignment.
