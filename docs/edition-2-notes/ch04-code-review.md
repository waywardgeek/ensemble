# Chapter 4 comparative code review

Reviewer: `review_restart_ch01`, October 9, 2026. **Current verdict after round
2: accepted with the documented inherited media exception.** Two of three
correction rounds used. Both lifecycle sensitivity findings and the round 1
test race are resolved. Original grading remains **90/100 FAIL**, with all eight
new checks passing. The initial findings and correction history are retained.

Read the entire `.agents/skills/ensemble-coding/SKILL.md`, AGENTS, carryover,
original Chapter 4 and [completed student review](ch04.md), including the
coordinator's subsequent cwd clarification. Chapter SHA-256 remains
`44e03445c850c939e331cb25c174e98df692063f824f9cedb748efa3ea1fee84`.
Compared the working student tree with frozen `solutions/edition-2/ch03/`, and
original `solutions/ch04/` with its own Chapter 3 predecessor.

## R1 — exercise the claimed lifecycle interleavings

Two current-chapter claims can pass without exercising their protected behavior:

1. In `ensemble/jobs_test.go:142`, the waiter goroutine is started immediately
   before Kill. It can enter Wait after the job is already killed, so no wakeup
   is necessary. Removing only `j.notify()` from `job.Kill` passed this focused
   test **20/20 executions** in an isolated copy. A separate reviewer probe
   observed the real Wait goroutine blocked in its select before killing:
   unmodified production passed, while the omitted notification failed with
   `kill did not wake waiter`. The original grader's 30-second wait also begins
   after kill; it is not evidence for waking an existing waiter.
2. In `ensemble/jobs_test.go:245`, releasing the FIFO does not establish that
   the read handler has reached Finish. The immediate status assertion can run
   first. Removing only the killed-state early return from `job.complete`
   passed **19/20 executions** of this focused test; one execution detected the
   intended `late file result changed killed state` failure.

Make the first check exercise a waiter already blocked at the kill transition,
and the second assert after an actual late completion has been attempted.
Demonstrate that the two isolated defects reliably fail their intended checks
with working baselines. Preserve the real implementation and fake boundaries;
no mocks, new product API or general synchronization framework is requested.
Update the student review/mutation record and rerun affected checks. The existing
delayed-output notification mutant does not establish either distinct claim.

## Design, scope and teaching

The implementation otherwise makes material improvements. Agent retains one
dispatch funnel; every ordinary tool gets a job/file/running record before its
handler runs. Agent owns Jobs, Jobs owns each Job, and both retain actual parent
interfaces. The call context reaches its real Agent and Engine. Import inspection
confirms star dependencies; common holds records/interfaces rather than jobs
behavior. The separate GUI still consumes the public Agent and passes its tests.

One job mutex coordinates output, status, cursor and notification. Process launch
and kill share that lock, closing the launch-before-attachment race. The PTY
reader owns completion; returning from the tool does not close active output.
Unlike the original, retained output is not duplicated in an accumulating memory
buffer. Reports read bounded spans, while regex matching deliberately reads the
unseen span. The serial collection uses a small ordered slice; handles belong to
Ensemble and output directories to Agent configuration. No future scheduler,
mailbox, watchdog, MCP connection or permissions framework was introduced.

Comments explain the important ownership, wait/process lifetime, cursor,
launch/kill race and CRLF-boundary decisions. They are useful teaching comments,
although the combined production ratio is 14.1% of nonblank lines versus 25.9%
in the original, below Bill's approximate 20% target. No padding request follows;
the concrete reasons remain legible beside the code.

Counts classify each physical Go line as blank, leading-`//` comment-only, or
active, including braces/imports and literal schemas; there are no block comments.
Tests are separate. No files moved between categories. Original compiled binaries
are excluded; the student's module manifests/sums are support metadata for the
single PTY dependency, not another implementation language or generated source.

Initial-submission counts follow; final test counts are recorded under round 2.

| Go category | Active | Comments | Physical | Chapter delta: active / comments / physical |
| --- | ---: | ---: | ---: | ---: |
| Original production | 3,096 | 1,083 | 4,565 | +723 / +235 / +1,029 |
| Student core production | 2,373 | 388 | 2,890 | +558 / +88 / +667 |
| Student GUI production | 8 | 3 | 13 | 0 / 0 / 0 |
| Original tests | 343 | 103 | 482 | +109 / +18 / +137 |
| Student core tests | 966 | 24 | 1,022 | +273 / +4 / +283 |
| Student GUI tests | 29 | 2 | 34 | 0 / 0 / 0 |

The student's active production addition is 23% smaller. Different schema
formatting and the original's Chapter 4 addition of search context also affect
the comparison; line count alone is not the improvement. The extra tests exercise
real disk output, PTYs, subprocess groups, FIFO blocking and actual ownership,
using the small HTTP provider fake. Their breadth is justified, subject to R1.

## Results and evidence

Reviewer reruns: empty gofmt output, core vet and uncached race tests pass;
separate GUI vet and uncached tests pass. Original Chapter 4 grading reproduces
**90/100 FAIL**, all eight new checks passing. A focused Chapter 3 grading run
confirms nested parity fails only the inherited `ref-render`/`ref-redaction`
controls. The coordinator's original Chapter 4 reference baseline is 100/100.

Read the [mutation evidence](../../solutions/edition-2/evidence/ch04/mutations.txt):
26 recorded defects ultimately killed, two initial survivors repaired, one
compile failure correctly excluded before repair. Independently confirmed the
repaired exact truncation-marker assertion catches a wrong total. The additional
R1 survivors above remain outstanding; they do not retroactively invalidate the
26 specific recorded results. All reviewer scratch copies were removed.

Audited all five live event logs. The three providers contain real delayed-job,
PTY input/echo and kill results. OpenAI's output report names 564 total/308 omitted
bytes at a 256-byte payload cap. Both debugger logs contain breakpoint/continue,
`p answer` yielding 42, and quit/exit zero. The shutdown repeat records the stated
PID and `job_killed` with reason shutdown; before/after PID disappearance and disk
inspection remain the student's recorded observations, not a new reviewer live run.
Usage sums match the student's table; OpenAI additionally records 6,272 cache-read
tokens separately from its 7,215 ordinary input tokens. No paid calls repeated.

The original Gemini debugger log contradicts the initial student explanation of
an overescaped regex and ten-second wakes: decoded patterns have one backslash
before each parenthesis, and called-to-returned times are
0.270/0.005/0.064/0.002 seconds. The escape-free repeat independently shows
0.431/0.007/0.070/0.003 seconds. The student corrected its explanation before
this verdict, preserving the original artifact and mistaken interpretation;
the repeat was unnecessary. Both logs support prompt matching. The initial 65/100 cwd issue
also includes an incomplete coordinator erratum, not solely student error.

The previously reassessed **media exception remains justified**: §4.3 introduces
RefHandle job metadata and §4.5 expressly keeps it out of rendered parts. Job
output recovery, redaction and truncation are current requirements and are not
waived. Preserve the actual failed parity score until a chapter introduces the
attachment capability. Initial Chapter 4 acceptance awaited R1; no other
production, scope or architecture correction was requested.

## Correction round 1

Reread the entire mandatory skill and the revised student review before this
reassessment. The changes address both original interleavings: `synctest.Wait`
establishes that the real waiter is blocked before Kill; a synchronous call to
the real `job.Finish` establishes an actual late completion attempt before the
killed-state assertion. The latter appropriately does not claim to join the
original FIFO reader. The real PTY/process-group and FIFO paths remain tested;
there are no mocks or production hooks.

Independent focused race runs reproduce the student results: unmodified tests
pass 20/20, omitted Kill notification fails the intended assertion 20/20, and
removed complete guard fails its intended assertion 20/20. Thus the original
R1 sensitivity defects are resolved. Production/dependency files and the prior
design/live assessment are unchanged. The student's original-grader rerun is
still **90/100 FAIL**, all eight new checks passing; no reviewer paid calls,
whole grader rerun or broad mutation sweep was needed.

**Remaining correction:** the new waiter test shares `result` and `waitErr`
between its goroutine and caller. On the missing-notification path, reading
`result` at `ensemble/jobs_test.go:157` is not synchronized with its later write
at line 145 when virtual time releases the waiter. The student's
`mutations-r1.txt` and an independent reviewer run both contain `WARNING: DATA
RACE` naming these accesses. This is a test race, not a production race. The
20 intended failures remain valid evidence, but the regression should fail
cleanly for that behavior without also racing during cleanup.

Synchronize completion observation on both successful and defective paths,
retaining the already-blocked waiter condition and clean waiter exit. Record
this failure and rerun the focused baseline and the two exact defects under
the race detector. No production changes, new API, larger test framework,
paid calls or original-grader rerun are required for this test-only correction.

Round 1 adds 23 active test lines and 10 explanatory comments. Core tests now
total 989 active / 34 comments / 1,055 physical; their Chapter 4 delta is
+296 / +14 / +316. Combined core/GUI tests are 1,018 / 36 / 1,089.
Production counts and the justified inherited media exception remain unchanged.

## Correction round 2 — accepted

Reread the entire mandatory `.agents/skills/ensemble-coding/SKILL.md`, carryover
and updated student review, including its correction of the mistaken claim that
round 1's failures were clean. The original warning remains in the retained
round 1 evidence. The only revised mechanism is the blocked-waiter test's result
delivery: a buffered channel carries text and error together, and both the normal
and broken-notification cleanup paths receive before inspecting them. The
pre-kill observation also uses that channel. This removes the shared-variable
race while preserving the real, already-blocked Wait and its explicit wakeup
assertion. The synchronous late-Finish check and its honest non-join limitation
remain sound. No production hook, mock or larger test framework was introduced.

Independent verification reproduces the [round 2 evidence](../../solutions/edition-2/evidence/ch04/mutations-r2.txt):
both focused tests pass 20/20 under `-race`; removing only Kill's notification
fails its intended wake assertion 20/20; removing only complete's non-running
guard fails its intended state assertion 20/20. Neither defective run reports a
data race, panic, build failure, deadlock or timeout. An initial reviewer text
replacement for the latter did not match because an intervening comment was
omitted from the search; it ran no tests and was not counted. The corrected,
confirmed replacement produced the stated result. All scratch copies were removed.
Formatting is clean and focused package vet passes.

Independently recounted final core tests: **1,000 active / 35 comments / 32 blank
= 1,067 physical**, up 11 active and one comment from round 1. Their Chapter 4
delta against the frozen predecessor is **+307 active / +15 comments / +328
physical**. Combined core/GUI tests total **1,029 active / 37 comments / 1,101
physical**. Production remains **2,373 active / 388 comments / 2,890 physical**
in core, plus unchanged GUI **8 / 3 / 13**. The small test increase now protects
the intended interleavings with clean failures and explanatory comments.

The unchanged production, comparative design, full validation and live evidence
above remain applicable; no paid call, full grader or broad mutation sweep was
repeated for this test-only repair. Accept the chapter with only the previously
justified positive URI-media parity exception. The last original result remains
**90/100 FAIL**, all eight Chapter 4 checks passing. Preserve that inherited
failure visibly in later parity results and reassess when an original chapter
actually introduces attachments; do not turn it into an earlier-feature demand.
There are no remaining current-chapter correction requests.
