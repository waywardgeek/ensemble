# Enhancement review: Chapter 4 — accepted with a recorded test exception

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Reviewer: `enhancement_reviewer_resumed`, October 10, 2026. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`, read entirely before review. Current
voice/procedure, carryover and bounded plan remain governing. This combines
original Chapter 4 with the relevant public-consumer and logging lessons of
original Chapter 5, using accepted `edition-2-ch05-r2` as the end reference.
Final combined decision: accepted at initial formal review, with the inherited
immediate-PID assertion exception described below. Zero correction rounds after
initial review. Preparation history remains below; it is not the current gate.

## Reading and retained teaching

Read all 642 lines of original Chapter 4 and all 591 lines of original Chapter
5, complete student and comparative reviews for both, original job predicates
and external-consumer harness/checks, accepted jobs/process/Agent/root/consumer
implementation and consumer test, and matching legacy jobs. Chapter hashes:

- Original 4: `44e03445c850c939e331cb25c174e98df692063f824f9cedb748efa3ea1fee84`.
- Original 5: `452a3edc46346bcd48a63483fc7c2ffbbb4056f29446dfc7fb6c292d3a0e4520`.

The screenshot freeze motivates uniform supervision; the buffered-channel
containment and retained-result explanation distinguish waiting from execution.
The `send_secret`/hand-maintained timeout-list failure motivates per-call limits;
the debugger's terminal refusal motivates a PTY and process/output ownership.
The shell-directory statistics and killed-as-done mutant supply concrete reasons
for visible per-call state and structural result assertions. Preserve these
incidents and their historical attribution. Old Chapter 5's logger ambush and
actual outside consumer are useful here without recreating a monolith/refactor.

## Preparation feedback and review boundaries

Sent author the sourced missing contracts: numeric fractional seconds, integer
handle/string input with default line submission, `job_killed`'s top-level job
and nested reason, optional cwd override absent for a default call, and dispatch
default truncation versus explicit read limits. The old dispatch fragment's
service bundle must not override actual-parent architecture. Logging belongs to
Ensemble; the custom tool uses the existing job/dispatch path and per-Agent set.

Historical student original grades remain Chapter 4 **90/100 FAIL** and Chapter
5 **110/120 FAIL**, with nested Chapter 4 still 90. Chapter 4 adds job-output
references and explicitly defers URI attachments; the inherited media exception
therefore remains justified. Revised grading must not relabel the old results.

The original Chapter 5 tool predicate accepts any result marker, including an
unknown `calculate` error, and never separately builds the external consumer.
The logger AST predicate demands a direct stored-interface logging call and a
dispatch struct in common, rejecting the accepted valid multi-hop/private
implementation. Its parity threshold accepts 80%, contrary to full parity.
These are demonstrated grader defects, not reasons to impose the old API or
add logger forwarders, a core calculator or a generic AST-analysis framework.
New positive consumer proof must execute application-defined behavior through
the public framework; source review still assesses ownership and independence.

The original Chapter 4 kill scenario waits after killing. It proves that a
terminal job is observable immediately, not that killing notifies a waiter
already blocked. Accepted source tests and their corrected mutation evidence
cover that distinct interleaving and late completion. Keep the original valid
check, but describe its proof accurately. Do not repeat the historical test race
or infer a waiter started merely because its goroutine was launched.

Live evidence reconciliation must preserve the original Gemini debugger's
successful decoded regex and subsecond wakes; the later repeat followed an
evidence-reading mistake, not a provider failure. OpenAI's 6,272 cache-read
tokens are separate from its 7,215 ordinary input. The first live shutdown
attempt and lost first external-consumer capture also remain historical failed
evidence attempts, with their successful repeats distinguished.

No original/frozen file, runtime or grader changed by this reviewer. Await the
coder's narrow plan and final submission before running grader validation or
deciding acceptance. No Chapter 5 enhancement work is released by this baseline.

## Draft review, before final evidence submission

Read the complete enhanced manuscript and new Chapter 4 grader source. The
agreed composition preserves nine original job checks/weights at 100 and adds
60 for separate consumer execution, resolved imports and diagnostic output.
The descriptor supplies an arbitrary tool call/result; it does not prescribe a
public Go API or new product protocol. `go list` binds the external dependency
to the submitted directory and checks actual imports rather than a guessed hub.

Independently assessed original Chapter 5's global-variable detector. It greps
package declarations and exempts lines by spelling (`names`, `name =`,
`embed.FS`), rather than checking mutable application state. A grouped `var (`
line is seen but its contents/uses are not understood. Retaining its automatic
PASS would preserve neither a sound absence claim nor legitimate immutable
declarations. Accepted the coordinator's explicit source-review replacement,
with the architecture still binding and the limitation stated in prose/report.
Read the accepted root, public consumer, common vocabulary, dispatch, jobs,
process ownership and registration paths: the consumer constructs the public
root and calls its real Agent; the handler follows parent interfaces to root
logging; the registry is Agent-owned and no package application state is used.
Historical registration/global-state mutation evidence remains applicable.

The new prose keeps the original sequence, engineering consequences and actor
cliffhanger. The public-consumer/logger addition earns its space despite the
soft length warning. No invented incident or new paid demonstration is claimed.
Independently summed all eight retained job/consumer JSONL sessions; published
usage, decoded debugger arguments, prompt-wake intervals and shutdown reasons
match. The failed capture and unnecessary repeat remain visible.

One pre-submission teaching clarification sent to the author: the new dispatch
snippet's unconditional `Finish` call needs its existing process-transfer guard
explained beside it. Otherwise it appears to perform the premature output close
that §4.8 correctly warns about. No runtime/API change is requested. Final
grader evidence, affected checks and combined acceptance remain pending; this
is preparation, not a completed formal review/correction round.

## Final independent review and decision

Read the final author and coder submissions and all changed grader files. Final
manuscript SHA-256:
`69430f9659ed8f8ce7504097e0d67235d0dc820c1b57da0e5627149eb6e8c7b9`.
The `Finish` explanation is corrected beside the example, with no runtime or API
change. The later debugger-escaping alarm was a preparation diagnosis error:
independent extraction and JSON decoding gives codepoints
`[92,40,100,108,118,92,41,32]`, matching `(dlv) `. No source regex repair occurred.
Read the actual fake-backed debugger transcript (458.91825 ms), consumer-case
result and killed-event dump. They support the published examples without new
paid calls or a claim that fake success replaces the retained live evidence.

The combined teaching is accepted. The old Chapter 4 sequence and incidents
remain substantial, rather than being reduced to a specification. The selected
old Chapter 5 consumer/logger lesson fits after the debugger: it asks whether
another program can use the machinery the reader has just built. Dropping its
standalone monolith/refactor and contradictory logger/global examples follows
the bounded plan. The author records those dispositions. Independent prose lint
passes hard checks; the 7,639-word, negation and person-gap warnings are real.
The original is 5,259 words with existing negation/person-gap warnings. The added
capability, missing contracts and evidence earn their space; neither cutting
retained stories nor adding token human mentions would improve the chapter.

All nine original job checks/weights remain. The corrected Chapter 3 parity is
strict, and the new consumer cannot earn success merely from an unknown-tool
result or an unrelated final sentence. The import check includes common's
outgoing edges and resolves the consumer to this source tree. The final consumer
build correctly targets the same module-root program that `go list` inspects.
The descriptor remains small and permits an arbitrary name/arguments/result.
Actual root use, parent interfaces, behavior placement and no mutable application
globals were independently source-reviewed; the numeric score does not certify
those properties. No mock, generic AST framework, public-API adapter or extra
product feature was introduced.

### Independent validation

- `go test -count=1 ./internal/gradeedition2 ./cmd/grade-edition2` passed in
  3.098 seconds; focused vet passed and gofmt listed no changed files.
- In frozen `solutions/edition-2/ch05/ch05`,
  `go test -race ./... -run '^TestExternalToolLoop$' -count=1` passed. In the
  core, `go test -race ./ensemble -run
  'Test(BlockedFileToolCanBeSupervised|RegistrationIsAgentOwned)$' -count=1`
  passed, covering late completion and Agent-owned registration.
- Independently copied the grader and accepted source into scratch. A passing
  real import-graph baseline preceded adding common's import of a new independent
  implementation package. Both core and consumer compiled. The actual topology
  check then failed specifically on that edge and lost its 25 points. This was
  behavioral rejection, not a compiler failure. All reviewer scratch copies and
  temporary test drivers were removed; none was placed in the shared repository.
- Read all thirteen raw changed-behavior mutation results and the corresponding
  coder report: declaration/description/schema, exact result, error flag, call
  identity, unknown tool, disconnected diagnostic, timestamp, sibling/common
  imports, wrong resolved checkout and direct private consumer import. Each
  fails the intended check(s). The different-name/input positive case confirms
  the default `shout` spelling is optional. Unchanged jobs/registration mutation
  evidence is retained rather than rerun as a whole-book sweep.
- Independently read the final full reports: accepted **160/160 PASS** (jobs
  100, boundary 60); legacy **90/160 FAIL** (jobs 90, boundary 0). Legacy's
  corrected search-context parity genuinely fails, and it has no separate
  consumer module. Original Chapter 5 results remain **110/120 FAIL** and
  **120/120 PASS**, respectively; accepted nested original jobs remain 90/100.
  No full root sweep or new paid demonstration was needed or claimed.

### Inherited immediate-PID assertion exception

The first independent source command also included
`TestKillWakesWaiterAndKillsGroup`. It failed at `ensemble/jobs_test.go:207`,
`process-group child survived`. A focused `-count=5` rerun reported one such
failure. The failing assertion calls `syscall.Kill(pid, 0)` immediately after
Kill has signalled the group and reaped the shell. The already-blocked-waiter
subtest passed; no race-detector warning occurred. These two failures remain
failures and are not covered by an all-source-tests-pass claim.

To inspect the failure, a scratch copy changed only the failure diagnostic to
report the PID and `ps` process state. Twenty focused iterations and a final
bounded hundred all passed. No failing PID state was captured; in particular,
an unreaped zombie was **not observed**. No further statistical sweep was made.

Source and independent evidence support the required group kill: the fixture
starts its child before reporting its PID; the implementation sends SIGKILL to
the negative shell PID, drains the PTY reader and waits for the shell. The child
inherits those terminal descriptors. Retained real group-kill/live evidence and
the original/revised process checks remain separate positive evidence. Immediate
PID nonexistence is stronger than signalling the entire group and reaping the
process this Agent owns; descendant reaping can belong to the OS. A reaping race
is a plausible inference, not an established cause from this investigation.

Accept this unchanged source assertion as a recorded test exception for the
enhancement gate. The evidence does not establish a surviving running child or
a production regression, and does not justify editing frozen code, widening a
sleep or inventing new process machinery. Group kill remains binding. A future
test correction should distinguish execution from retained process identity and
preserve the positive group-kill proof; it must not silently turn this failed
assertion into a pass. The coordinator agrees with this scoped disposition.

### Simplicity, comments and comparison

New production grader files contain 319 physical / 256 active / 50 comment-only /
13 blank lines; the integration test is 19 / 15 / 2 / 2. The command grows from
66 to 86 physical lines, 59 to 79 active. Counts exclude whitespace-only and
trimmed `//` lines from active code; no new block comments need special handling.
The roughly 16% comment share is below the approximate 20% teaching target but
the comments explain the consequential choices: ownership-proof limits,
independent result evidence, direct versus transitive imports, canonical source
resolution and matching build/inspection targets. Exported new symbols and
fields have useful documentation. No padding is needed. Original job predicates
are reused rather than copied, and the focused fake/integration tests exercise
real code. The size is proportionate to a separately built external application
and a real import graph, not another grading framework.

The unchanged runtime comparison in the accepted Chapter 4/5 reviews still
supports the design. Across original 3→5, legacy production rises from 2,373 to
3,324 active lines (+951), comments from 848 to 1,000 (+152), physical from
3,536 to 4,760 (+1,224). Accepted core rises from 1,815 to 2,396 active (+581),
300 to 397 comments (+97), 2,223 to 2,926 physical (+703); its separate consumer
adds 74 active / 12 comment / 91 physical lines and GUI stub remains separate.
Moved environment parsing is counted once. The accepted design avoids duplicate
output buffering and a second dispatch/logging path. Core tests add 358 active /
16 comment / 383 physical lines; the consumer adds 70 / 2 / 75. Their real
lifetime and external-use coverage earns the growth. Frozen historical exported
documentation gaps and lower comment density remain recorded shortcomings,
not newly repaired by this grader pass.

No unresolved prose, architecture, scope or grader finding blocks acceptance.
The only exception is the explicitly retained inherited process-existence test
limitation above. This pass did not change original graders, first-edition text,
frozen exports or runtime; unrelated shared-tree first-edition work is outside
that claim. Coordinator owns the commit/tag and the next chapter gate.

## October 10 follow-up: combined assignment clarified

Accepted the focused prose clarification requested by Bill. Independently read
the new TL;DR and expanded Exercise against original Chapter 5 §5.6, its logger
and globals sections, and the existing combined grader. The five steps now make
the external task explicit: independently build a public-library consumer, read
provider configuration, construct its Ensemble and Agent, register and invoke
an application-defined tool, and shut down. Instance registration and actual
parent access preserve Bill's architecture without imposing the original API
spellings or restoring its global registration example.

The TL;DR covers the existing tool declaration, successful correlated result,
timestamped diagnostic, import topology and ownership obligations. The unchanged
rubric remains jobs 100, consumer 25, star 25 and diagnostic 10. Source review
still establishes real composition, parent relationships, behavior placement
and absence of mutable application globals; diagnostic output alone does not.
No new feature, protocol, runtime change or grader requirement was introduced.

The diff adds 53 lines and removes none; the original opener and stories remain.
The summary and explicit task sequence earn their space by making the combined
assignment findable. Independent focused prose lint passes all hard checks at
8,116 words; its length, negation and person-gap warnings remain warnings, not
reasons to remove useful teaching. Scoped whitespace checks pass. Verified final
manuscript SHA-256:
`8beba353da4ae602064dcfb62d7a7354fa8f4d5140c344642930fc0b67cdb7c2`.

No runtime suite or paid run was repeated for this prose-only change. Prior
scores, validation and the recorded immediate-PID assertion exception remain
unchanged. This follow-up is accepted at initial review with no correction
round; the coordinator owns its checkpoint. Mandatory skill, carryover and full
voice/procedure were reread under the bounded enhancement plan.
