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
