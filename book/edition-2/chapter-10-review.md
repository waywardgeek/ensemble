# Chapter 10 contract review

Coordinator review, October 8, 2026, of draft `9181683`. The complete chapter,
outline and preparation evidence were read. This is a contract review, not
implementation acceptance. Chapter 9 remains an unbuilt predecessor; no
Chapter 10 student or live demonstration is released by this record.

The draft keeps fresh construction distinct from deliberate session resume,
assigns SessionStore to Agent, and preserves one durable append path. Its
shutdown, unfinished-work refusal and historical-job distinction bring the
later lessons forward without inventing process recovery. Default human
startup and explicit standalone selection have separate, reviewable behavior.

## Four proposed decisions

These are coordinator choices within the authorized rewrite, not additional
rulings attributed to Bill.

1. **Request allocation:** accept the durable request index and snapshot cursor
   lower-bound rule. A canceled caller's transient handle need not become a new
   durable admission event. Reconstruction must prevent reuse of every recorded
   request identity; it cannot claim to recover invisible canceled admissions.
   Preserve that distinction in prefix-equivalence checks.
2. **Imported origin:** accept immutable `origin.json` as the seed for a
   snapshot-only store. It is a read-only reconstruction dependency, not another
   append writer. Later checkpoint replacement must retain and validate the
   seed. Missing or partial imports refuse rather than inventing recovery.
3. **Plain-mode handlers:** accept exact installed-definition and visible-set
   compatibility for session reopening. Existing fresh Agents retain their
   earlier behavior. Definition equality does not attest implementation bytes
   or create a security boundary; the draft says this explicitly.
4. **Skill provenance:** accept compact ordered transition provenance referring
   to immutable activation material. A snapshot must validate earlier visibility
   and retirement, not infer legitimacy from the final active roots. Live import
   also checks the caller's compatible frozen catalog and candidate bytes;
   historical inspection remains independent of source files.

## Clarifications before checker preparation

- Make the job watermark explicitly session-local: the greatest historical job
  handle represented in that session, including handles outside the GUI window.
  It is not an exact serialization of Ensemble's shared allocator cursor.
  Other Agents and occupied output files can advance that allocator. Resume
  raises the live allocator to at least the restored maximum without lowering
  its existing value; snapshot/replay comparison must not require equality with
  another Agent's allocations. The current source's `Ensemble.AllocateHandle`
  confirms that the counter belongs to the application root.
- Specify canonical-number behavior sufficiently for independent importers.
  In particular, preserving identity integers must not leave handler-schema or
  arbitrary payload numbers subject to a new lossy conversion during hashing.
  Two distinct accepted values cannot gain the same compatibility identity merely
  because both pass through binary64. Publish the normalization rule and examples
  for integral decimals, exponent spellings, large integers and negative zero;
  distinguish a canonical hash representation from stored original payload bytes.

Both clarifications are resolved in `6cf8660`, reviewed by the coordinator.
The job watermark now denotes this session's historical maximum and only raises
the root allocator's floor. Canonical numbers use exact decimal coefficient and
exponent normalization, with examples preserving large adjacent integers;
original payload bytes remain unchanged. The same revision removes a draft
JavaScript-safe identity ceiling that conflicted with Chapter 9's uint64 domain.
Overflow is refused before allocation or mutation, independently of file bounds.

The complete contract is accepted for independent checker preparation. This
does not release a student: accepted Chapter 9 and a published Chapter 10 checker
remain prerequisites. No runtime, live result or historical rewrite is claimed.

## Independent student ownership-plan review

October 8, 2026, after the Chapter 9 checkpoint. `/root/grader_ch05` read all
384 lines of the fresh student's plan at
`7cb84290af3e8bac5359339f8ab5d4e7e941bf74`, recorded in
`solutions/edition-2/main/evidence/ch10/student-review.md`. This was a design
review against the new teaching, architecture and coding skill, before runtime
implementation. No first-edition persistence implementation or future chapter
requirements informed this check. The reviewer's earlier grading, Chapter 9
comparison and isolated streaming-maintenance exposure remain disclosed in the
Chapter 10 grader review; this is not a cold-student evaluation.

The plan is sufficient to release implementation **after the student rereads
and acknowledges the published Q1–Q3 clarifications**. Those concern session
record bounds and inherited skill-refusal precedence; actual imported
anchor/tail history versus the saved display window; and exact replay bytes
versus purpose-specific canonical equality. This acceptance does not resolve
those questions through unpublished fixture assumptions. No additional
architectural blocker was found.

The owner graph preserves Agent authority and the import star. An Agent-created
stateless codec can serve standalone validation without a SessionStore or
injected sibling service. Private validation owners, accepted-event job capture,
off-actor checkpoint work and close ordering fit the contract. The public
open/inspect/export/import/checkpoint seams support independent checks. Their
names and the conservative shared export/save worker gate are permissible
design choices, not newly required spellings.

Retain two implementation safeguards:

- Private validation owners remain under the real application owner but inert
  and unregistered. Validation must not start processes or an Actor, charge
  replay as usage, write policy files or mutate shared allocator authority
  before the candidate has passed validation.
- GUI delivery-before-acknowledgement waiting belongs to the connection
  lifetime. Overflow or disconnect releases that wait without parking Actor
  or the checkpoint worker; losing the connection cannot undo a committed save.

These safeguards explain how to preserve the planned boundaries; they do not
add public APIs or another owner. The complete nested semantic codec grammar
must still be published before semantic fixtures are written. This record
accepts the ownership plan conditionally, not implementation, local tests,
provider runs, historical comparison or final chapter readiness.

## Independent accepted-byte clarification review

October 8, 2026. `/root/coder_ch08` independently reviewed author freeze
`c5ad6c04a67c0e2eeafde77f52d90dfdd7695d8e`. This is a narrow teaching review
of the prepared-event boundary, separate from the preceding coordinator and
ownership-plan reviews. The reviewer previously implemented Chapter 8, prepared
early Chapter 9/11 checks and reviewed later contracts; this reviewer neither
authored the correction nor implemented Chapter 10. Earlier Q1–Q3 and Unicode
proofreading remains recorded in chapter-10-clarification-review.md.

The complete three-file author delta/direct response, surrounding §§10.3/10.8,
and the retained initial-prefix diagnostic were read. Focused predecessor reads
checked Chapter 2's JSONL, typed parts and raw usage, Chapter 6's transient/accepted
response and retained stream JSON distinctions, and Chapter 9's controlled
whole-record preflight. Full current voice/procedure remained loaded from the
immediately preceding review, without intervening compaction. Working author
files match the frozen revision. No old implementation or checker source was read.

The clarification closes the observed representation gap. New session admission
validates first, prepares one bounded final event encoding, derives owned raw
fragments from it, appends exactly those bytes, and only then applies/publishes
the accepted event. The accepted-event ordering does not turn Chapter 6's earlier
transient stream fragments into durable facts. Preparation may change formatting
and escape spelling while preserving number lexemes, decoded string/manual/
signature bytes, object-member and array order, and opaque semantics. Canonical
hashing and floating-point conversion cannot substitute for that preparation.

Once accepted, the recorded fragments define the exact state/watch/snapshot/
replay comparison. Existing imported records already define their own accepted
bytes and receive no read-side normalization; original physical bytes still
determine their read limits. The strict scalar validation and dedicated raw-JSON
wrapper rules remain applicable. Legacy standalone behavior is explicitly outside
this new session-write boundary. Skills candidate and complete-record preflight
still produces controlled skill_too_large before storage admission; bounded
preparation cannot defer that refusal into a terminal writer failure. Separate
session file/count/write failures retain their existing persistence disposition.

The direct feedback and evidence consistently attribute this integration gap to
the coordinator/grader diagnostic, separately from the student's Q1–Q3. The read
receipt shows successful local creation, untouched inspect refusal at record 7,
and a successful separate checkpoint-free rebuild, with the two raw_usage spacing
differences in context and saved watch state. It records one local HTTP request
and unchanged originals. This reviewer inspected that retained receipt rather
than rerunning the binary; neither the receipt nor the author response proves a
repair of current source or any real-provider result.

No new consequential contradiction was found in the affected teaching. The sole
existing fenced JSON fixture is unchanged by the complete diff. The retained
prose executable independently passes all hard rules at 7,681 words; length,
negation and person-gap warnings are soft. The new paragraphs explain an observed
failure and its admission boundary without inventing a successful restart.

Reviewed SHA-256 values:

- chapter-10.md: `41ef91e7fad0308f32bd874c8cb952fddc8900d41f3534c847183adb0e66a26a`.
- chapter-10-evidence.md: `b205861a3d915c7e0f96d1d118704dd7016a1b393be80da7f4c8e39e4d4fe102`.
- chapter-10-student-feedback.md: `0b4ccab128da68c300c9a831693e2037ebc9e5291949e25bde8bc2318a9ca8d2`.
- ch10-initial-prefix-diagnostic.json: `00305cf60c4cb33c7685edf4fe9a68a041e84dc62d53947233ebe31d2c4c0a1a`.

Disposition: accepted-byte clarification proofreading accepted. Implementation,
affected controls and student confirmation remain separate. No runtime edit,
new build, checker edit, provider call or implementation acceptance occurred.
