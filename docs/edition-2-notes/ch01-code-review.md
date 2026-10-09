# Chapter 1 comparative code review

Reviewer: `review_restart_ch01`, October 9, 2026. **Accepted** on initial
review; zero revision rounds used. No requirement exception is needed.
Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

Read AGENTS.md, the entire skill, carryover, workflow, original Chapter 1,
`solutions/ch01/main.go`, the completed [student review](ch01.md), all submitted
student source/tests and both evidence files. Chapter source SHA-256 matches
the student's recorded `d22555ddf1cdb40b0ce97733e6550641a63b465ece044d5eb65b6b6d2f2dbce8`;
its latest source commit is `e1605916114230570b348ab5961c193fa1cebf80`.
The comparison is against the unchanged first-edition solution, not the retired
second-edition attempt.

## Size, comments and simplicity

Counts classify physical Go lines as blank, comment-only (leading `//`), or
active, including declarations, imports and braces. Neither tree has block
comments. Tests are separate. Chapter 1 starts from empty in each edition, so
these totals are also each edition's Chapter 1 additions; there are no moved
files to discount.

| Category | First edition | Student | Change |
| --- | ---: | ---: | ---: |
| Production active Go | 199 | 281 | +82 |
| Production comment-only | 48 | 59 | +11 |
| Production blank | 31 | 29 | -2 |
| Production physical | 278 | 369 | +91 |
| Test active Go | 0 | 123 | +123 |
| Test comment-only / blank / physical | 0 / 0 / 0 | 6 / 4 / 133 | +6 / +4 / +133 |

The student additionally has a three-line `go.mod`; the reference uses the
repository module. Two short text evidence files are documentation, not
production or test machinery. There are no generated sources, examples,
retained mutants, additional languages or new testing frameworks in this submission.

Active production grows 41.2%. This is justified by Bill's required ownership
and public composition API, with small error-handling improvements. The student
has 84 active lines in the executable, 35 in the public root, 30 in common, 40
in Agent and 92 in Engine. The reference's one-file client is smaller, but does
not provide this required owner graph. The new interfaces express actual current
relationships; there is no speculative subsystem or abstraction framework to
remove. Straightforward getters are short without compressing control flow.

Comments are 17.4% of nonblank production lines versus 19.4% in the reference.
The new comments explain ownership, synchronous lifetime, staging failed turns,
provider accounting and secret-free diagnostics beside their implementations.
They retain the chapter's reasons for replaying history and walking reply blocks.
Their usefulness meets the roughly 20% teaching target; adding filler would not
improve the code. No comment accuracy issue blocks acceptance.

## Architecture and behavioral review

The executable really constructs the public Ensemble and calls its Agent.
Ensemble retains Agents and owns the logger; Agent retains configuration,
conversation and Engine; Engine owns its HTTP client and cumulative usage.
Constructors store required parent interfaces, attachment checks the Engine's
actual Agent, and accessors return those owners. Diagnostics follow Engine →
Agent → Ensemble. The root alone imports both implementation spokes. Common
contains vocabulary and interfaces, with no implementation imports or displaced
behavior. No mutable application globals or callback service bags are present.

Both CLI modes use that same path. The scope remains Chapter 1: one raw HTTP
request per turn, exact retained dialogue, response-block concatenation and
cumulative measured usage. No future provider, streaming, tool, GUI or actor
work has entered the submission.

I accept the student's small departures: staging a turn until success prevents
an invalid next conversation; missing usage and empty replies fail explicitly;
bounded exchanges and restrained diagnostics are proportionate. Counting
provider usage before rejecting an unusable text response preserves actual
accounting. These are inexpensive current safeguards, not recovery machinery.
The student explicitly limits concurrency and timeout-test claims. No failed
grader check has been concealed or waived.

## Tests and evidence

The original grader remains the shared principal coverage for both editions.
The added tests exercise the real public root, Agent, transport and parser
against small external HTTP fakes. They cover separate Agent state and owners,
failed-response handling, retry absence, successful use after failure, root
logging, credential omission and the precise API-version header. There are no
mocks or interfaces introduced solely to replace internal behavior in tests.
The tests add meaningful boundary coverage without duplicating the entire grader.

Independent checks passed: `gofmt -l .` produced no names; `go vet ./...` passed;
uncached `go test -count=1 ./...` passed; the original
`make grade-dir CH=1 DIR=solutions/edition-2/main` scored **100/100**, five requests,
**476 input / 65 output**. The coordinator also recorded the original reference
baseline at 100/100.

Read all [mutation evidence](../../solutions/edition-2/evidence/ch01/mutations.txt):
30 distinct mutations, 31 student executions, with the initial version-header
survivor honestly preserved. Independently reproduced that exact mutant in a
temporary copy: the original grader still scores 100/100, while the added test
fails specifically with `wrong Anthropic API version`. Inspection confirms the
original fake checks only header presence. This is a genuine grader gap for the
future author, not grounds to alter the grader or claim an exception now.

During this initial review, the student clarified an omitted second replacement
in the streaming mutation record and its pre-final-gofmt spelling. The completed
record explains how it compiled and reached the intended wire check; no rerun
was claimed. The Engine-parent mutation fails at the real attachment guard,
which is a valid behavioral rejection rather than a compile/setup failure.
Timeout/cancellation sensitivity and future nontext handling remain explicitly
unproven, as the student states; they are not silently included in the evidence.

The [live transcript](../../solutions/edition-2/evidence/ch01/live.txt) documents
model discovery followed by three actual PTY `chat` turns, retained “Copper
Finch,” visible cumulative usage ending at **350/33**, and clean Ctrl-D exit.
It also preserves the initial launcher-EOF mistake. I accept this student-run
human-interface evidence; I did not repeat a paid live session. The model's
raw-HTTP answer alone is not implementation proof; source inspection supplies
that proof independently.

No blocking findings remain. Preserve the grader-version gap, invocation
correction and launcher lesson for the later author–coder phase. Do not expand
Chapter 1 to address future features.
