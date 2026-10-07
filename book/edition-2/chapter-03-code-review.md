# Chapter 3 independent post-run comparison

Date: 2026-10-07. Status: bounded revisions accepted on source inspection and
independent checks; final commit binding remains with the coordinator.
Initial student production is `590c4f4`, with the complete initial
run/evidence checkpoint `540fb4a` retaining identical production. Historical
Chapter 3 was read from course commit
`b8275b0f8436bf43bb139587d67183bf60af3435`; its last solution change is
`ab10324a52daf15c6f0449f4ba8b5824cdbb0045`.

## Reading and separation

The reviewer read the full new Chapter 3 contract, current architecture, and
mandatory skill before constructing the independent CLI acceptance checker.
That checker was written without inspecting the new student source. Its
initial black-box run found only the empty-file `end_line:0` refusal across
the three surfaces: 33 of 36 scenario groups passed. The clarified contract
and this failure remain distinct from the frozen initial attempt.

After that phase, read the old tool implementation, engine loop, and focused
tool tests, plus the new initial tools/turn production, Chapter 3 changes to
composition/common/rendering/CLI, focused tools and turn tests, and external
tools consumer. Chapter 2's unchanged implementation was already reviewed;
this is a scoped Chapter 3 comparison, not a claim to have reread every file
of both repositories. No old source or answer snippets were given to the
student. Feedback below supplies rationale only.

## Concrete comparison

| Concern | Historical answer | New initial answer |
| --- | --- | --- |
| Ownership | Mutable package Registry; flat Engine and direct tool functions. | Agent-owned Registry, common parent interfaces, Engine-owned turn, public clients. One concrete capture parent violated the interface rule; coordinator correction is reviewed below. |
| Capability authority | One global map supplies declarations/dispatch. | Per-Agent selections govern both; live caller declarations must agree, while explicit offline declarations remain usable. |
| Arguments | A separate argument-summary table drifted from schema; unknown fields retry permissively. | Shared field metadata drives schema/defaults/type checks and actionable diagnostics, with strict unknown-field rejection. |
| Effects and persistence | Dispatch/result facts mutate memory; Save normally occurs at turn end. | Each dispatch is durable before execution; result-write failure faults the Agent and admits that the effect may exist. |
| Loop bound | The seventeenth response can be accepted before stopping, leaving its calls unanswered and returning prior narration. | Sixteen requests including the first; complete the final batch, then report a round-limit error. Only a completed final answer enters CLI stdout. |
| Read/search/list | Partial range semantics, unbounded listing, no search windows, silently skipped read failures. | Explicit empty/range rules, count and byte bounds, merged search windows, binary/symlink policy, and visible failures/omissions. |
| Writes/edits | Stat-then-write overwrite race; useful exact-anchor refusal policy. | Exclusive creation closes that race; exact matches, deletion and no-op are explicit and observable on disk. |
| Commands | Separate streams retained without bounds. | Separate capped writers continue consuming whole writes, allowing the process to finish. The comment explains why returning the retained byte count would break draining. |
| Diagnostics/comments | Strong rationale for refusal and typed dispatch, mixed with verbose or stale assertions. | Concise reasons at durable batch and ephemera boundaries; shared declaration metadata avoids a second argument glossary. |

The new answer performs more ownership and persistence work and has additional
promised behavior. Line counts are not a quality measure. Keep the new
architecture rather than importing the old answer's shortcuts.

## Requested and guided revisions

1. **Owner interface for command capture.** The coordinator found
   `capture.parent *Registry`. Its replacement with `common.Registry` retains
   the actual creator and the Registry-to-Agent logger path. The reviewer
   inspected that guided diff and accepted its design. This was an initial
   implementation violation of an already taught rule, not a new requirement.
2. **Empty-file EOF spelling.** Initial decoding treated supplied `end_line:0`
   as an explicit line target. The author clarified that it means through EOF
   even on an empty file; only an explicit start or positive end requests a
   line there. The guided diff and positive/negative tests resolve that scope.
3. **Avoid whole-configuration copies for workspace lookup.** Each tool path
   and command used `Agent.Config()`, which deep-copies credentials, settings,
   and all schemas merely to obtain the workspace string. Request a locked
   scalar query through the owning Agent interface; preserve public Config
   copies and do not duplicate workspace storage in Registry. No benchmark or
   numerical speedup is claimed. The revised common.Agent interface exposes
   Workspace, the root Agent reads its scalar under its mutex, and tool paths
   and command cwd use it. Accepted without weakening public Config copies.
4. **Do not corrupt valid UTF-8 at byte caps.** An independent frozen-binary
   probe read `éX` with budget 1 and ran `printf 'éX'` with stream budget 1.
   Both results contained U+FFFD: JSON repaired the sliced invalid byte into a
   three-byte replacement character. That is neither a retained source prefix
   nor within the one-byte content budget. The historical solution shares
   this defect; reproducing it is no improvement. The author has now taught
   the longest complete UTF-8 prefix at/below budget for valid UTF-8 input.
   Request complete-rune positive controls, split-rune negatives, and capture
   chunks that divide a rune before finalization. Arbitrary invalid-byte
   transcoding is not a newly inferred chapter feature.

The last two requests are post-run review feedback. They preserve the initial
attempt and need affected local checks, not repeated paid conversations.

## Revision and sensitivity evidence

The reviewed binary is SHA256
`0cc92c69d1f6dd2316af2a0d3f59f5f336f72384ae2c2e65f7d009fe8ff81e9d`;
`evidence/ch03/reviewed-source-hashes.json` binds the source. The reviewer
inspected the complete guided/review diff and new UTF-8 tests. A shared tools
helper returns a complete prefix; capture buffers raw retained bytes until
finalization, so an OS write boundary does not discard a rune completed by
the next write. Tests cover budgets 1/2/3 after JSON encoding, both command
streams, listing/search prefixes, and split capture writes. The helper receives
the common Registry interface, preserving the owner path.

The expanded independent checker passes **39/39** scenario groups on all
three fake provider surfaces. The frozen initial binary passes **33/39**;
its six failures are exactly the empty end-zero and split-UTF-8 cases on each
surface. Eight independent checker control tests pass. The Unicode scenario
was added after the review finding and author clarification, not represented
as a check available to the initial student.

`scripts/edition2/audit_ch03.py` freezes a disposable source copy and runs a
passing control plus five serial mutations. Every mutant builds and fails
exactly its stated check set: read byte slicing, command byte slicing,
stopping a batch after ordinary error, reducing the request bound to fifteen,
and stopping before dispatching the sixteenth batch. The original source hashes
remain unchanged. The audit binds checker, audit, source, and binary hashes
and retains the applied substitutions and complete per-check receipts.
This establishes sensitivity for those five guarantees, not exhaustive
coverage of every chapter promise.

Raw coordinator handoff receipts are `/tmp/ch03-independent-initial.json`,
`/tmp/ch03-independent-initial-expanded.json`,
`/tmp/ch03-independent-reviewed.json`, and
`/tmp/ch03-independent-mutations.json`; preserve them in the chapter evidence
before final checkpointing. No paid calls were made by the reviewer.

## Grader compatibility review

The reviewer inspected the coordinator's shared grader correction and controls:
Gemini accepts either the older Schema field or its JSON Schema field, rejects
conflicting fields, and still requires an object schema. Fake live sessions
receive the fake provider's known resolved model while retaining the requested
alias; offline fixtures explicitly clear inherited resolved identity. This
matches the exact-provenance contract rather than teaching alias guesses.
No scoped coverage weakening was identified. Root owns the preserved legacy
baseline and mutation receipts; the same frozen student binary moved from
65 to 100 through the fixture correction, not a hidden student revision.

Independent acceptance currently checks CLI outcomes, wire facts, persisted
facts and actual disk effects. Public ownership, capability isolation,
injected persistence faults, and live demonstrations are separate gates.
Command labels/markers do not alone establish correct label-to-stream
association; that limited check is explicitly disclosed. Checker controls
and later source mutations are separate from code review.
