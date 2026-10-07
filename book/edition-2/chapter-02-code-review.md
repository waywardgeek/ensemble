# Chapter 2 independent post-run code comparison

Date: 2026-10-07. Status: the bounded code revisions are accepted on inspection
and preserved at `cc1bec45c3327c87728a4040f762155d8e860a0b` above the initial
checkpoint. This is separate from the earlier
owner-path audit, root acceptance, live-proofreading, and Bill's approval.

## Compared artifacts and reading

The initial new solution is frozen at
`39a92ca27a418712832ac0dcbbfbe4e32b3bca35` in its own repository, following
the three-provider CLI and public-consumer runs. The first-edition Chapter 2
standard was read from course commit
`24867ca45f5588c73e7caa7c17fa2d64f1ab86de`; its last solution change is
`d87e8d571995af5eed84deee163c8964c422a9a1`. Chapter numbers match here.

The reviewer reread the full mandatory coding skill, current architecture,
voice, and procedure. The Chapter 2 contract was read in full during contract
review; later deferred-input, schema, and diagnostic-related sections were
checked against the implementation. Read all first-edition production files,
its focused OpenAI-content test, and the new production paths, tests, GUI
client, and feature-evidence map across this and the earlier owner audit.
The old CLI test file was not part of this reading. No old solution code was
given to the student or copied into the new implementation.

The student's feature map names actual CLI and external-consumer receipts
for all three providers. Those runs were not repeated by the reviewer.
The initial inherited score of 95 reflects a fixture rendering unresolved
calls under the amended contract; the coordinator owns that fixture review.
Neither that harness issue nor a future passing score substitutes for this
quality comparison.

## Concrete gains and tradeoffs

| Concern | First-edition standard | New answer and review |
| --- | --- | --- |
| Architecture | Flat executable; Engine holds log, context, configuration, HTTP client, and persistence path. | Public library with Ensemble/Agent/Engine owners, event-log and llm spokes, common interfaces, and optional GUI module. Keep the corrected owner tree. |
| Persistence | Events mutate memory first; saving rewrites the file, normally after a completed question. | Exclusive log creation and append-before-application/observation; partial writes fault the Agent. Failed attempts remain explainable without poisoning the next request. |
| Attribution | One Context usage total; response provenance can contain an absent model identifier. | Engine totals by producing provenance, separate requested/reported model identities, explicit fallback flag, and raw usage observation. |
| Parsing | Missing usage becomes zero; invalid cache subtraction is clamped; some empty text is discarded. | Required nonnegative integral counts, invalid subtraction rejected, present empty text retained, missing-call IDs synthesized deterministically. |
| Replay | Typed parts are grouped before rendering; some original order is lost. Signed calls can lose incompatible metadata rather than refuse replay. | Ordered parts, exact provenance compatibility, call-bound refusal, text-bound signatures retained on their part, and explicit deferred-input projection. |
| References/redaction | Result redaction can retain one selected reference; Gemini may fetch a surviving locator again. | Every replaced child retains its own reference; stubs are descriptive and do not refetch removed content. The narrower explicit mapping is preferable. |
| Data ownership | Shared slices and configuration follow the flat implementation's assumptions. | Public inputs, history/context snapshots, configuration, and observer notifications have copy boundaries. Earlier review fixed Load's schema alias before this checkpoint. |
| Diagnostics | Several useful field-specific explanations, but raw transport/provider error text can escape. | Sanitized errors and safe cancellation identity. Event-validation messages became too generic; restore actionable safe reasons, not raw payload interpolation. |
| Comments | Strong explanations of render-before-request ordering and reducer responsibilities, mixed with stale or incorrect claims (for example, JSON map serialization). | More concise code and fewer stale claims; a few essential WHY comments should accompany the subtle ordering/copy boundaries. |
| Economy | Per-vendor private wire structs are explicit but add substantial repetitive machinery. | Small map-based renderers and shared parser helpers are a reasonable scoped choice. JSON copies enforce public ownership, but unnecessary full snapshots on the prompt path need trimming. |

The old interactive CLI/help and additional speculative redaction machinery
are outside the new required feature subset. Their omission is not itself a
regression. The new system has more public ownership and persistence work to
do; line counts would not establish which design is better.

## Requested revisions

1. **Avoid copying all history for an internal predicate or scalar.** In
   `ensemble.go`, `Prompt` calls `Snapshot` for eligibility and again solely
   to obtain `LastSeq + 1`. `Snapshot` serializes/deserializes the entire
   growing context. Use lock-protected internal access for those operations;
   retain the owned snapshot for rendering and every public copy guarantee.
   No new benchmark or claimed speedup is required to remove this visible
   unnecessary work. Preserve synthesized-ID and event-order tests.
2. **Give safe errors a useful reason.** Distinguish a bounded set of field
   and transition failures, such as an invalid reference and an unsolicited
   response, and retain the source line when loading. Use controlled reason
   text, never raw record contents, arguments, keys, or provider bodies. Add
   public regressions distinguishing reasons and rejecting private-marker
   disclosure. The old answer's useful specificity does not justify its
   unsafe interpolation.
3. **Preserve the essential WHY comments.** Explain why rendering precedes
   recording the request (that event consumes ephemera), and why publication
   uses per-recipient copies only after persistence. Use concise original
   explanations rather than copying the older answer's commentary.

These requests were sent to the coder with the mandatory full skill-read
instruction. The author received the corresponding teaching findings.
Preserve the frozen first attempt and report revisions separately. These
internal/diagnostic changes need relevant local checks; they do not justify
repeating unchanged paid successful conversations.

## Revision review and exposure limits

The reviewer freshly read the full coding skill, architecture, voice, and
updated procedure, then inspected the complete production diff and the new
public diagnostic tests. All three requests are resolved:

- `canPrompt` and `nextSeq` read borrowed state while holding the Agent lock.
  Prompt retains one owned context snapshot for rendering; public snapshot
  and observer copy guarantees remain intact. The enclosing operation lock
  still serializes the live turn.
- Private `validationError` values carry controlled categories. Load includes
  those categories with the source line and uses a generic reason for other
  error types. Envelope parsing likewise uses static reasons. The new public
  tests distinguish reference, unsolicited-response, sequence, and payload
  errors and check returned/logged diagnostics for private-marker disclosure.
- Original WHY comments explain render-before-request consumption and
  persistence-before-per-recipient publication.

The saved `reviewed-local-checks.json` records passing format, vet, and tests
in all three modules and corrected inherited grading at 100/100. The reviewer
read that receipt without duplicating the runs or making paid requests. Root
reports revised independent acceptance at 44/44, a control plus ten targeted
defects detected, and full legacy regression passing. Those retained receipts
remain separate evidence from this source review.

The reviewer also read the shared grader's `ch2RoundTripFixture` correction
and its controls. It preserves every dumped byte and appends explicitly
labeled fixture results only for outstanding calls; it leaves completed
histories unchanged and rejects malformed parsed dumps. No scoped coverage
weakening was identified. The root's old-reference baseline and mutation runs
are the behavior evidence, not this inspection alone.

Chapter 2's current live section was checked against its feature ledger: it
distinguishes real CLI and controlled-result consumer runs from offline
request reconstruction, local fault fixtures, and the GUI stub. Totals and
model identities agree with that ledger; the coordinator independently checked
the underlying receipts. No material prose blocker remained. The author was
asked to reconcile the final pending-status sentence with the accepted commit.

No old source snippets or answer text were sent to the student. Feedback
contained the three design-level requests above and their rationale, informed
by the comparison. That is post-run review feedback, not a blind continuation.
The first two student chapters inherited coordinator conversation/history;
the coder reports no direct old-chapter, old-solution, or grader-source reads,
but their context exposure prevents a strictly blind-evaluation claim. The
updated procedure requires a fresh, new-material-only student context for each
subsequent chapter. This limits the experiment's claim, not the observed code
behavior or the utility of the comparison.
