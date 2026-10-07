# Chapter 4 source research

Status: initial complete contract independently reviewed and accepted;
the later human-chat propagation is under review. No new Chapter 4
implementation, grading, or live run.

Read the complete first-edition Chapter 4, derivation sections A–D and its
closing design observations, current grader job/lifecycle/output fixtures
and checks, global-review lifecycle findings, and the commit records below.
The derived fact sheet contains stale descriptions of voice.md; the current
full voice guide governs instead. Its absent `review-ch04-code.md` reference
was checked and is not treated as a read source.

## History that changes the teaching

- `d08a92b`: killed-as-done mutation originally passed because a kill note
  satisfied a text search. Corrected structural status plus negative normal-
  completion assertions exposed it. The recorded live debugger result is
  historical evidence only.
- `9f00cb3`: per-call cwd, missing-directory refusal and recorded effective
  override. A substring check once mistook an echoed argument for a real
  working-directory change; test the process's output as a separate line.
- `9465464`: one-shot consumption notice must appear on both job and inline
  supervision paths; no-note negative control prevents always-printing from
  passing. Another setter consumes the previous pending value.
- `fcf9a2b`: concurrent mutation subprocesses produced timing flakes in
  unrelated checks. Serial execution and an echo fixture with a deliberate
  response delay distinguished stale-output pattern matches from genuine
  new output. The commit's elapsed times are historical measurements, not
  current performance promises.
- `0253737`: removal of obsolete coder scaffolding. Do not carry old claimed
  editorial tasks forward as if they are still unresolved.

## Concrete source tensions

The original dispatch snippet always finishes a job after tool.Run returns,
while its debugger section correctly says the subprocess reader owns eventual
completion. The new chapter must select one completion authority before code.
Old Call fields inject Job/Jobs/Limits/Events; do not reproduce a sibling bag
under the new actual-owner parent chain.

The first-edition large-read fixture explicitly passes `max_bytes:2000000`.
It does not prove that default reads ignore the source tool's selection limit.
Separate full produced result on disk from full source-file content, and
separate both from the dispatch report cap. PTY merged output is a deliberate
change from the new Chapter 3 separate-stream contract, requiring explicit
instruction and checks instead of a blanket unchanged-parity claim.

Current nine-check weights still total 100. This author has read their
definitions but has not run the Chapter 4 grader or deletion suite. No claim
is made about current pass status, platform coverage, artifact sizes, or live
model capability. Retired corpus percentages, old model defaults, and private
screenshot-story details are not fresh public facts.

## Review and next action

Global reviewer asked for forward lifecycle lessons when the Chapter 2/3
reviews permit. The outline's owner/handle/artifact policy is a proposal,
not settled user instruction or a coder handoff. Full contract drafting
waits for Chapter 3 review; no implementation is authorized by this research.

Coordinator subsequently authorized the full draft and accepted Agent-owned
Jobs plus Ensemble-owned application-wide handle allocation as working choices.
The draft now states those owners, common interfaces, and one completion
authority. Reviewer forward lessons are incorporated: monotone kill/completion,
late output, per-Agent one-shot state, consumed-but-recoverable omitted bytes,
new-output pattern boundaries, process/goroutine limits, and serial timed audits.

Root identified a new asynchronous dependency: predicted response sequences
can become stale when a job event is appended during HTTP. §4.3 now finalizes
missing call IDs on an owned copy under the actual append sequence, retaining
parser facts and supplied IDs. The acceptance table names that positive case.
Report snapshots/cursors and terminal events also require a coherent append
order so a delayed running snapshot cannot follow a done event as a reversal.

Current primary references checked October 7, 2026: creack/pty package reference
for terminal/session setup and its sample-code limitations; Go os/exec reference;
official Delve installation guide. No current dependency version or platform
coverage is invented. The draft links the primary documentation and requires
recording actual installed versions. Scoped prose lint passed all hard rules;
short length/person-gap warnings are advisory and no padding was added.

Independent review caught missing required times and call provenance in the
literal job fixture; both are now present. Clarified successful supervision
of a failed job: the original execution failure stays on its result/job fact,
while a valid later wait reports that failure with a successful observation.
Killed-report restrictions apply to generated metadata, not arbitrary retained
child text containing words such as done or exit_code. Added that adversarial
output as a required positive control instead of relying on whole-text regexes.
Carried the newly taught valid-UTF-8 cap rule into head/tail truncation.

Coordinator's direct terminal/Delve environment probe is capability evidence
for this workspace only, not a new Chapter 4 implementation or live-model run.
The future student must still demonstrate the real multi-command debugger.
The closing comparison instructions now explicitly reserve old-source reads
for the independent reviewer; the student receives findings and new teaching.

A further direct read of the inherited sendinput/debugger fixture and old
supervision argument schema exposed `append_newline` default true. The initial
draft's unmotivated exact-bytes-only rule would strand those ordinary debugger
commands without Enter. Corrected the new contract before implementation:
retain default Enter, allow exact bytes with append_newline:false, and state
empty-input behavior. The source is author evidence only, not student input.
CLI selection of all ten tools and library selection remaining explicit are
also now stated rather than left for inference.

Independent final contract review is accepted in `chapter-04-review.md`.
The reviewer checked the complete revised chapter, including newline defaults,
fixture timestamps/provenance, successful observation of a failed job, and
generated killed metadata versus arbitrary retained child output. This is
contract readiness, not student validation or Bill's editorial approval.
