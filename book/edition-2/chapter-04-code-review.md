# Chapter 4 independent code review

Initial review recorded October 7, 2026. Revision review remains open.
The reviewer authored Chapter 3's human-client integration, but did not build
Chapter 4. Expectations and checker preparation preceded the initial Chapter 4
freeze. Source exposure and hashes are in
[the source ledger](chapter-04-review-source-ledger.json).

The preserved initial checkpoint is `9803b006aa2ad14337655595c057b3016d11363e`;
its production source is `9f76d9e99ad1c52c6af5d5c4496bbe2410e3ae70`. The live
binary SHA256 is `296c6f9bd0f529e9765be9d53a980b21c9e02d60a732b981da8109d75d32b68a`.
All 30 source bindings in the student's initial manifest match that immutable
revision. Independent probes and mutations run in disposable copies; no
student source, frozen export, historical implementation or shared grader was
edited by this reviewer.

## Findings requiring revision

1. **Keep tool wire handling in tools.** Initial `internal/jobs/jobs.go`
   `Supervise` decodes JSON, validates tool-specific fields, switches on tool
   names and constructs consuming-call errors. `internal/jobs/limits.go`
   similarly decodes external arguments. The Registry owns the schemas and
   ordinary decoders, so these operations split one responsibility across two
   spokes. Move argument decoding, tool selection and associated presentation
   into tools. Expose typed settings, lookup, input, kill and report operations
   through common interfaces. Jobs must retain pending settings, cursors,
   process workers and terminal/report synchronization. Preserve the logger
   path in moved helpers. The coder accepted this rationale before revision.

2. **Use Call's existing owner chain.** Initial `internal/llm/turn.go`
   constructs Call with its Engine parent, then passes another Agent into
   `dispatch`. Derive the required owner from `c.Engine().Agent()` instead.
   The private admission adapter already provides the necessary services.
   This removes a redundant route and makes the declared parent operational.
   The coder accepted this rationale before revision.

3. **Keep read selection metadata out of the source artifact.** Initial
   `read_file` with source `ABCDEFGHIJ` and `max_bytes:4` produces a 42-byte
   artifact: `ABCD` followed by its source-cap notice. The artifact must contain
   the four selected source bytes; source truncation belongs in report
   metadata. The coder raised the same concern independently. The author
   clarified §4.5 before the affected fix, including delayed first reports and
   the distinction between source selection and recoverable report omission.
   New independent cases cover ASCII, a partial UTF-8 boundary, a selected
   range, both caps together, and a FIFO whose read completes only after the
   initial running report. They fail on all three initial provider paths for
   the intended artifact/byte-count distinction.

## Initial behavior and fault evidence

The prepared CLI/replay checks pass **50/50** after corrections to the
independent checker. Actual Delve runs through human chat reach a breakpoint,
inspect 42, exit, expose the result in `/history`, redact it and confirm the
next request's projection. These use local scripted providers and are separate
from the student's paid-model terminal receipts.

The external public consumer passes **3/3 provider cases**. It checks shared
workspace handles, foreign wait/input/kill refusal, per-Agent pending settings,
omitted tool selection, ordered Agent-attributed observations during parked
HTTP, isolated Agent close and application-wide close. The consumer compiles
outside the library and imports its public package only.

Ten independent storage/lifecycle tests pass under `go test -race`. They cover
failed pre-dispatch persistence, terminal/result failure after a real file
effect, process cleanup with an unwritable log, failed spool writes waking a
waiter, cursor retention after failed report append, both report/terminal
commit orders, kill before startup, twelve kill/exit races, and a killed local
function that still causes an effect while its result is discarded. Barrier
fixtures establish the required order. Format output is empty and vet passes
in the isolated module.

Six pure checker controls pass. Ten targeted deletion audits have passing
controls and the intended failing checks. Mutations cover occupied-file
truncation, repeated report output, incorrect synthesized IDs, suppressed
background observations, skipped durable dispatch, early cursor advancement,
ignored spool-write failure, killed-as-done, shell-only kill and skipped
Agent shutdown. This is targeted evidence, not a claim that every possible
defect or every promise within each check has been mutated.

The initial audit exposed a weak group-kill fixture: its child could die from
SIGHUP when only its shell was killed. The strengthened child ignores SIGHUP;
the passing control still passes, and shell-only kill now fails because the
child survives. Both audit results remain in the evidence directory.

Other preserved checker corrections are explicit:

- Delve was initially absent from the review process's PATH. The installed
  executable was then supplied through PATH.
- The name `debug.go` made a Delve location ambiguous with Go's runtime source.
  Renaming the independent fixture to `audit_target.go` fixes that ambiguity;
  the failed debugger output remains recorded.
- The allocation-error case initially demanded a recoverable paired result.
  The author confirmed that §4.2 permits an infrastructure error before
  dispatch. The corrected check demands a surfaced error, no successful
  answer/usage, no automatic continuation, no effect and no invented job.
- A spool-failure process check initially treated `kill(pid,0)` alone as proof
  of a surviving process, although the immediate `ps` query found none. It now
  uses the same runnable-process distinction as the CLI checks. A read-only
  spool descriptor isolates write failure from a secondary close error.

Commands and raw results are retained in
[chapter-04-review-evidence](chapter-04-review-evidence). The current checker
has 53 CLI/replay cases after adding the three source-cap cases; those new
cases are failed initial evidence, not part of the earlier 50/50 result.

## Comparison with the first-edition standard

The corresponding standard is `solutions/ch04` at
`ab10324a52daf15c6f0449f4ba8b5824cdbb0045`. Reviewed areas include its jobs,
supervision tools, complete turn/dispatch/shutdown engine and command lifecycle.
The student receives the rationale above, never old implementation snippets.

The new answer improves several concrete boundaries. The old Jobs table owns
its own handle counter and uses `os.Create`; the new application allocator
and exclusive artifact allocation preserve output across Agents and restarts.
The old Job retains an unbounded `bytes.Buffer` alongside the file and ignores
file write errors. The new spool is the output store, and write failure faults
the owner and wakes waiters. The old report advances its cursor before the
result is durable and reads status separately for the logged snapshot. The
new report serializes snapshot, append and cursor update with terminal
publication. The old engine buffers lifecycle facts and saves at turn/shutdown
boundaries; the new worker uses Agent's serialized durable append path even
during HTTP. The old Call contains direct Jobs, Job, limits and event fields;
the new Call's Engine parent supports the corrected ownership model.

The new reader/reaper join, process-group cleanup under persistence failure,
split-CRLF handling and complete UTF-8 head/tail cuts strengthen the old
process implementation. Public Agents, observers, separate clients and an
optional GUI module also replace the old single executable's ownership.
Those changes justify additional synchronization; fewer lines alone would
not justify discarding their invariants.

Retain useful intent comments from both approaches: a wait delay changes when
the model looks, not the job's lifetime; ordinary local tools also need jobs;
the artifact is the recovery address; and a late local result cannot reverse
a killed decision. The new comments explaining report/terminal serialization,
reader drain and exclusive allocation are concise and useful. The old code's
separate supervision handlers reinforce finding 1, but its injected Jobs bag,
mutable global registry and weak storage semantics must not be copied.

The main remaining quality work is responsibility separation and the source
artifact correction, followed by independent review of the revised source,
affected checks and retained live evidence. The student teaching review and
author dispositions exist and were read. Final acceptance is not claimed here.

## Revision review: d25d3fd

Reviewed `ee15a56491c00a2a284d94ebde3264045650b544`, then runtime revision
`d25d3fd4e552cd17c75bf814c9899903878cfbd5`. Findings 1–3 are resolved:
Registry now owns wire decoding, regex validation, supervision dispatch and
consumption-note presentation. Jobs exposes typed operations and retains
mutable settings, cursor and worker state. Call derives its Agent from Engine.
A typed execution result separates retained text from report notes; the source
cap tests pass, including delayed completion and overlapping report limits.

The coordinator's manual data review found additional defects that ordinary
model runs did not reveal. A public nil response panicked before validation,
and invalid parser indices could panic while retaining Agent's lock. The
independent probes reproduced these failures on ee15a56. d25 separates temporary
parser metadata from public durable Response and validates indices before
access. Independent tests now check safe rejection, unchanged history and
log bytes, untouched caller facts, and a subsequent valid append. Removed
public metadata-index cases are explicitly skipped as no longer representable;
three separate typed parser-admission cases cover their internal replacements.

The same revision rejects changes to the lifetime log destination atomically,
retains an omitted destination during other configuration changes, and reports
the actual path on replay load. Independent controls verify the original writer
still receives events and the refused path is absent. Job snapshots copy their
exit-status pointer. Engine owns the accounting mutex and documents its lock
order; it no longer relies on an undocumented Agent lock for its usage map.
Ensemble's interface directly exposes its allocator.

Bill's explicit private-runtime ruling resolves placement: private Engine,
Registry, Jobs and Job implementations may remain in their responsible spokes
when common interfaces expose ownership. Remaining scoped casts are not a
placement violation. Engine checks its optional turn capability before dispatch;
Call dispatch follows that admitted Engine path. Job's concrete Service cast
stays within the owning jobs implementation and follows its constructor invariant.
No new sibling dependency or capability bag was introduced.

Independent revision results are **53/53 CLI/replay cases**, **3/3 external
public-consumer cases**, the original ten storage/lifecycle controls under the
race detector, and added malformed-admission/log-identity controls. All eleven
targeted mutations have passing controls and the intended failure, including
the new mutation that puts the source-selection notice back into the artifact.
Six pure checker controls pass. Formatting output is empty; vet and full tests
pass in the isolated main and external consumer modules. The immutable source,
commands, probe hashes and raw results are recorded in
[the revision manifest](chapter-04-review-evidence/revision-manifest.json).

A separate evidence-only finding remained after runtime repair: the verifier
checked each provider's launch identity inside its replay/write loop. Corrupting
the third provider's source identity caused refusal after the first provider's
derivative had already changed. This was reproduced only in an isolated copy.
The coder moved all launch identity checks ahead of every replay and write.
Independent wrong-third-source, wrong-third-binary and wrong-executable controls
now leave the entire copied evidence tree unchanged; the original positive
replay succeeds. The negative receipt is preserved alongside its corrected
controls. Raw terminal and log receipts were never rewritten by this review.

The initial live prose was checked against all three complete terminal records,
all 58 artifact byte counts and hashes, actual debugger output, accepted-response
usage totals and the recorded before/after PID observations. Its figures agree.
The real public-consumer supplement's six Agents also agree with their raw log
usage, artifact hashes and ordered observation sequences. A first arithmetic
script assumed one spelling for the before-EOF alive field; it was corrected to
handle both recorded spellings without changing the receipts.

Runtime revisions are accepted by this reviewer. Final chapter acceptance still
requires the revised live demonstrations, final student/author feedback exchange,
and proofread manuscript reconciliation. This section does not relabel initial
runs as demonstrations of d25, or claim Bill ran any of the sessions.

## Closing runtime and evidence decision

The final revised demonstrations at `9341117` use runtime d25. All three
complete human PTY transcripts were read, all 34 revised job artifact hashes
and lengths checked, and response usage recomputed. Each source-capped read
retains exactly `alph` (four bytes); its source notice and the two-byte report
omission remain distinct. Actual Delve output shows the breakpoint, 42 and
normal exit on all three paths. Default/exact input, limit consumption and
overrides, slow waiting, middle recovery and EOF cleanup were repeated. The
corrected-runtime real public-consumer supplement also passes.

The coder's initial revised prompt wrongly supplied a callback field directly
to read_file. Two model paths attempted it and received the correct refusal;
the observed follow-up removed that field. The third model omitted it. This
preserved user-input mistake is not a new runtime defect. Revised Gemini
responses supplied IDs, so synthesized-ID timing remains a deterministic
barrier proof rather than a claim about those paid responses.

Final receipt review caught an empty source map in boundary-binding.json.
The verifier silently verified zero source files. That failure is preserved.
Evidence-only revision `e1c64886564e7d1e8205ee08c17f98f32e001912` populates all
51 historical Go/module hashes and independently derives the required complete
set before verification. This reviewer checked every hash and seven isolated
controls: empty map, missing entry, wrong hash, late public source mismatch,
late public binary mismatch, wrong CLI, and the original positive invocation.
Each refusal leaves all 273 copied files unchanged; the positive invocation
succeeds without any byte change. No paid run was repeated for this repair.

The student's completion review and author's dispositions were read. The
coder confirms the published explanations resolve the identified difficulties.
Runtime, comparative quality, revised live use and evidence verification are
accepted. No further implementation change is requested. Manuscript final
reconciliation/proofreading and the coordinator's export/tag remain separate.
