# Enhancement notes: Chapter 4, including original Chapter 5

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Author: `enhancement_author_resumed`, October 10, 2026. Final prose submitted for
independent combined review; coder evidence reconciled. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

## Sources and composition

Read all 642 lines of original `book/chapter-04.md`, file revision
`0253737c06197d6f9b35e442916b7475c1a59a99`, SHA-256
`44e03445c850c939e331cb25c174e98df692063f824f9cedb748efa3ea1fee84`;
all 591 lines of original `book/chapter-05.md`, file revision
`f5d38c997b985deac29c5ed8b1d9e9aacd1af765`, SHA-256
`452a3edc46346bcd48a63483fc7c2ffbbb4056f29446dfc7fb6c292d3a0e4520`.
Read both complete student/coordinator records and comparative reviews, relevant
accepted `solutions/edition-2/ch05/` Agent/Ensemble/job source, external consumer
and README, and retained live evidence. Governing documents were read in full
on resumption. End reference is historical `edition-2-ch05-r2`; source/export
paths and all historical evidence IDs remain unchanged.

The coordinator approved preserving Chapter 4's sequence and actor cliffhanger,
with a focused public-boundary/logger section after the debugger and before
Exercise. Original Chapter 5's external-consumer capability is part of this
chapter; a standalone refactor would recreate the architectural defects already
avoided in Chapters 1–3. No further numbering or later feature is assigned.

The through-line remains the frozen screenshot and the difference between
waiting for work and supervising it. Preserve the entire August screenshot
incident, local/NFS counterexample, four-line containment mechanism, abandoned
result discovery, three verbs' measured use, 1,340,013-byte flood, watchdog's
two repairs, one-shot-limit misfire, PTY refusal, persistent-shell comparison,
directory-tax measurements, debugger lifetime trap and closing deafness. The
opening is explicitly historical CodeRhapsody narration, as in preceding
enhancements; dated corpus/API observations are not new measurements.

From old Chapter 5 retain the motivation of a separately built application
executing a framework-unknown tool, public types as the usable boundary,
and the logger ambush's distinction between the evidence and a model's account
of it. Correct the root/logger/registration ownership. The old corporate-price
opener, monolith/refactor instructions, global registration example, file/line
inventories and contradictory score tables do not belong in this sequence.
The `Call`→`common.Call` rename incident is deferred, with its lesson already
covered by the upfront small-edit/comment-review discipline; no extra anecdote
section is needed to teach another refactor here. This is an explicit disposition,
not silent loss of the historical finding.

## Findings incorporated

| Source/finding | Enhancement and reason |
| --- | --- |
| Screenshot/whole-agent claim vs allowed panic | Narrowed the guarantee to a stuck call no longer freezing the Agent; panics remain invariant failures. |
| Original dispatch sample used Engine service bag | Replaced its ordinary-job fragment with actual Agent ownership and callContext parent; same allocate/record/run/wait lifetime mechanism. Jobs belongs to Agent, Job to Jobs; Ensemble allocates handles. |
| CLI output path vs multi-Agent library identity | Kept required CLI `cr/io/N`; explained per-Agent DataDir in embedded use and the single-Ensemble fixture. No global storage identity. |
| Missing input protocol | Published numeric seconds/zero, regex/cap types, integer handle, input string and newline submission. Added killed-event top-level job/reason example and exact debugger JSON regex argument. |
| Bare-result promise vs PTY/metadata | Preserved small-result behavior while naming intentional merged-shell output, cwd and consumed-limit diagnostics. |
| Read cap | Explicit read range/max_bytes remain; default cap moves to dispatch so full output reaches disk. Payload cap excludes report/marker text. |
| cwd first 65/100 and incomplete erratum | Explained optional override field absent/empty for default calls; launch workspace independent of storage and no hidden grader flag. Preserved shared responsibility for missing clarification. |
| Job interleaving findings and two test correction rounds | Added concise existing-waiter vs already-complete distinction, late Finish terminal guard, and test cleanup race lesson. No new framework or audit machinery. |
| Debugger evidence misreading | Kept original successful pattern and measured prompt wakes; recorded the unnecessary repeat and corrected interpretation rather than inventing model failure. |
| EOF shutdown | Distinguished final answer from session EOF; unsuccessful expired-before-EOF attempt and bounded positive repeat remain visible. |
| Old5 external proof and logger | Public types/real composition root, per-Agent synchronous registration, ordinary job path, positive handler output and root diagnostic. No global registry, false parent, second tool loop or prescribed public Go spelling. |
| Old5 topology/globals/logger checks | Explicit source-review limits: compiler cycles do not enforce star, variable names do not prove mutability, direct-selector shapes do not prove parent ownership, printed diagnostic proves only delivery. |
| Launchers/credentials | Own-module binary and scratch live instructions; no reference convenience run presented as student evidence or literal key in command. |
| Grading | Coordinator approved separate 100 jobs + 60 boundary points; coder supplied exact generic consumer-case file/flags. Original Chapter4 90/100 and old5 110/120 with nested90 remain historical FAIL. |

The new ordinary-job code is a schematic excerpt of accepted Agent dispatch,
with lookup, supervision and returned-event details outside the displayed branch.
The public logger expression matches the actual consumer. No exact method naming
or generated adapter is imposed on other implementations.

Preliminary reviewer clarification: the sample calls Finish unconditionally,
which could appear to contradict §4.8's process-reader lifetime. Added two
comments at that call explaining the existing Finish guard: ordinary handlers
complete there; a started process retains its PTY reader's completion/output
ownership. No code behavior or API changed.

Preparation correction: the coder initially reported the debugger JSON sample
as overescaped. Direct extraction from the manuscript decoded to codepoints
`[92,40,100,108,118,92,41,32]` and matched `(dlv) `; the coder independently
confirmed the same bytes in both its earlier capture and the current file.
The false alarm came from reading JSON-escaped tool output of a Python `repr`
as source bytes. The manuscript regex was already correct and was not edited.
This repeats the historical serialized-representation pitfall; validate the
decoded artifact before correcting it.

## Receipts and checks

Independently summed all five historical Chapter 4 JSONL logs before adding the
figures. Main runs are 29,113/985, 7,215/593 plus 6,272 cache-read, and
30,236/1,437 plain input/output. No cache writes. The shutdown and unnecessary
debugger repeats are 4,444/139 and 11,578/728. Independently decoded the original
Gemini regex and computed called-to-returned intervals: 0.269766, 0.004939,
0.064092, 0.002384 seconds, with quit at 0.028895 and no pattern. The chapter
rounds those four prompt waits, accurately. It links ch04.md and actual event
evidence, not a nonexistent ch04/live.txt file.

Old5 consumer receipt remains `solutions/edition-2/evidence/ch05/live.txt` and
its JSONL files: one actual shout per provider, exact HELLO FRAMEWORK job bytes,
timestamped diagnostic, and 932/73, 363/25, 222/65 input/output. The first lost
Anthropic event capture and extra paid repeat are disclosed. No new paid runs
were made or requested.

`go run ./cmd/lintprose book/chapter-04.md book/edition-2/chapter-04.md` passes
all hard checks. Original 5,259 prose words; enhanced 7,639 including selected
old5 material. The soft length/negation/person-gap warnings remain for independent
judgment; no story was cut or human mention padded to satisfy a counter.
Source-relative comparison retains 602 of 642 original Chapter 4 lines, adding
343 and removing 40 after the two-comment clarification. Local manuscript links resolve. This comparison excludes
the additional source contribution from original Chapter 5.

Final [coder evidence](enhancement-04-code.md) matches the published 100 jobs +
60 boundary/consumer rubric: accepted endpoint 160/160 PASS; legacy endpoint
90/160 FAIL, with search-context parity missing and no separate consumer module.
Original Chapter 5 results remain 110/120 FAIL and 120/120 PASS respectively.
All thirteen targeted changed-behavior mutants failed their intended checks;
the alternate custom tool name/case passed. These results do not establish the
source-only ownership/global-state obligations described in the chapter.

The exact published consumer case passed; the killed-event fence survived the
accepted binary's dump with its fields intact. The exact debugger JSON ran in
the accepted CLI's real chat with a provider fake and real Delve, reached its
prompt and accepted quit in a 459 ms session. The coder also checked the
schematic dispatch/logger against accepted source. No prose-contract mismatch
or further manuscript change was needed. Affected grader tests and vet passed.

The reviewer observed an intermittent failure in an unchanged source-only
process-survival assertion. Its process state was not captured on failure;
descendant reaping remains a hypothesis. The coder record preserves this
limitation and the independent review owns its disposition. It does not alter
the historical or revised grader scores or authorize editing frozen runtime.

No original/frozen file, tag or student score was changed; no commit, mock or
later chapter draft was made. Final manuscript SHA-256 is
`69430f9659ed8f8ce7504097e0d67235d0dc820c1b57da0e5627149eb6e8c7b9`.
Author alignment is complete; independent combined acceptance remains pending.

## October 10 follow-up: make the combined assignment explicit

Bill requested that Chapter 4's TL;DR cover the original Chapter 5 grader's
needs and its exercise include that chapter's work. The accepted manuscript
already contained the consumer/logger teaching, case file and 100 + 60 grading
tables, but had no TL;DR and left the consumer lifecycle distributed across
those explanations. Added a TL;DR after the preserved opener and a five-step
external-application sequence in Exercise: separate public-module build,
environment/configuration and construction, per-Agent registration, actual
handler/result/diagnostic execution, and shutdown. The description/object-schema
requirement is explicit. Source review still owns the parent-chain, logging,
behavior-placement and mutable-state proof; numeric output does not replace it.

This clarifies existing scope against original Chapter 5 §5.6 and the unchanged
combined grader. No sample, command, table, runtime, grader, frozen export or
historical result changed. The original stories and explanations remain intact.
Author reread the full current voice/procedure, mandatory coding skill and
carryover; the bounded plan still supersedes the retired procedure's workflow.
Focused prose lint passes every hard check at 8,116 prose words (previously
7,639); soft length, negation and person-gap warnings remain for judgment under
Bill's right-length instruction. `git diff --check` passes. No runtime or paid
run was needed for this unchanged contract. Submitted for independent review;
manuscript SHA-256:
`8beba353da4ae602064dcfb62d7a7354fa8f4d5140c344642930fc0b67cdb7c2`.
