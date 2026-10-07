# Chapter 4 independent contract and prose review

Date: 2026-10-07. Status: revised contract ready for student handoff, subject
to the coordinator completing Chapter 3 and freezing the predecessor.
No Chapter 4 implementation, tests, or live results are claimed.

## Reading and teaching

Read the full Chapter 4 draft, with the current voice guide and chapter
procedure freshly loaded in this review sequence. Reviewed it against the
mandatory coding skill, architecture decisions, amended Chapter 3 contract,
and durable later-lesson map. Reread original Chapter 4 lifecycle/output/audit
sections while advising the author; that historical material is not student
input. This is not a claim that the entire old book remains in context.

The debugger-versus-exit opening gives the reader an immediate reason for
jobs. The chapter explains why returning a handle cannot finish the process,
why a wake differs from cancellation, and why retained output differs from
the reported interval. The worker/dispatcher distinction is taught before
implementation rather than discovered in a later repair chapter.

## Material findings resolved

1. The literal job-log fixture initially omitted required UTC timestamps and
   the tool call's provenance. Those are now present. The compact imported
   response rule does not excuse missing required part or envelope fields.
2. Waiting successfully on a failed job had an ambiguous error flag. The
   original execution result and job retain their failure; a valid later wait
   is a successful observation, with `tool.is_error:false` and an explanatory
   report. Invalid supervision remains an error. The revised text states this.
3. A killed report must retain child output that may literally contain
   `done` or `exit_code:7`. The restriction now applies to generated status
   metadata, not those retained bytes. The audit section adds that adversarial
   output as a positive control, avoiding a misleading whole-report regex.
4. The Chapter 3 UTF-8 correction now applies at both head and tail cut
   boundaries. The omitted count uses actual retained bytes, so unused budget
   does not become corrupted text or inaccurate accounting.
5. The author's final compatibility pass made `send_input`'s newline policy
   explicit: `append_newline` defaults true, false writes exact supplied bytes,
   and empty input either sends a blank line or only waits according to that
   flag. The reviewer checked those defaults and edge cases in the revised
   table and prose. The CLI explicitly selects all ten tools while an omitted
   public selection remains empty.

The reviewer rechecked the amended passages and fixture. No remaining
material contradiction was identified in this synchronous-dispatch/job scope.

## Implementation and grader obligations

- Agent owns Jobs, Jobs owns live Job, and Ensemble allocates application-wide
  handles. Capability visibility and supervision remain per Agent; shared
  workspace artifacts must not collide. Interface parents reach actual owners.
- Output/process completion belongs to the worker. Pre-dispatch persistence,
  terminal publication, report/cursor commitment, close, and killed-versus-done
  races must agree. Cleanup cannot abandon processes because logging failed.
- Background append must work while a prompt waits or HTTP is in flight.
  Missing call IDs use the actual committed response sequence on an owned
  copy; do not freeze real-time observations merely to reserve a number.
- One-shot limits are consumed by the literal next accepted attempt, including
  invalid/unknown operations and another setter. Pending state belongs to Jobs.
- Complete artifacts and bounded reports are different. Cursor advancement
  includes omitted bytes; pattern matching considers eligible unseen output,
  including split reads. Send-input's eligibility begins before its new input.
- PTY's merged stream deliberately replaces Chapter 3's separate streams.
  Reconcile inherited/new checks explicitly instead of claiming both contracts.
- Imported job records create no live work and require no artifact access.
  Event validation must protect identity, locator, monotone bytes, and terminal
  transitions. A late record cannot reverse a terminal status.
- Model-backed interactive use and the debugger are still required. Pending
  receipts, prerequisite failures, fault controls, and public-client evidence
  remain separate from this prose/contract acceptance.

The author's cited current PTY/Delve documentation was not independently
opened in this phase; this review checked the contract rather than certifying
dependency-version or installation facts. Verify the chosen dependency and
actual environment during implementation. Bill's editorial approval remains
separate from this handoff recommendation.
