# Chapter 6: author response to student feedback

October 7, 2026. The student's initial review lives in
`solutions/edition-2/main/evidence/ch06/student-review.md`; runtime and live
validation remain separate from this teaching response.

## Unknown fields on a Gemini text-bearing part

During implementation the student found that §6.4 prevented coalescing across
unknown content without specifying how the retained neutral part preserved it.
The earlier parser could discard an extra field while keeping the text. The
student proposed preserving the whole object as standalone opaque material,
rather than inventing a second meaning for the existing text signature field.

Accepted after coordinator review. §6.4 now states the recognized text shape,
validation of known types, absent/false thought equivalence, conservative
whole-object preservation with exact provenance, and no visible/thinking delta
for that opaque object. Known signed text and calls retain their existing
rules. Plain and streaming normalization must agree. §6.5 publishes a paired
fixture, matching/foreign replay expectations, and malformed-known-field and
opaque-only negatives. This is a teaching clarification before the affected
fix. The student subsequently confirmed that it resolves the question and
implemented the paired and negative fixtures before initial runtime freeze
`aa5f86a`. The coordinator reports local module/race and initial independent
checks passing. Live evidence and final validation remain separate; this
confirmation closes the teaching question, not those gates.

## Initial teaching experience and actual use

The full initial student account is now frozen with runtime `aa5f86a` and
initial live evidence `f73b01e`, `d12a0cb`, `3417575`. It reports no remaining
teaching contradiction after `f085b95`. The following dispositions preserve
that pre-comparison account; independent code quality and acceptance remain
separate gates.

| Student finding | Author disposition |
|---|---|
| The three completion boundaries, actual-sequence identity and exact SSE/usage examples made the design testable. | Retain §§6.1–6.5. They explain why early display cannot authorize effects or count successful usage. |
| Bounding/draining fragments and ordering display against reliable completion were the hardest implementation work. | Retain the explicit queue, cancellation and overflow contracts; difficulty alone does not justify weaker ownership or completion rules. No additional ambiguity was reported. |
| A fixture receiver copied a mutex; a usage assertion compared JSON key order. | Student implementation/test mistakes, corrected locally. Keep semantic usage comparison and exact opaque-byte assertions distinct; do not alter the contract to hide either mistake. |
| The inherited grader reports 0/100 under old observation and structural assumptions. | Retain the result and disclose the incompatibility in §6.9. Coordinator/new independent acceptance remains necessary; no passing score is claimed. |
| Initial prompts finished before a hint or interrupt could be sent. | §6.8 retains the actual attempts and bounded follow-ups. A received hint, wire inclusion and model compliance remain different observations. A late interrupt correctly reports false. |
| Gemini hit MAX_TOKENS in human and public runs. | §6.8 distinguishes valid accepted partial output from completing the requested explanation, and records the human limit notice. |
| No supported thinking delta or live overflow occurred. | State that absence explicitly. Keep deterministic fixtures separate from the demonstrated stalled-subscriber completion path. |
| Actual use confirmed tool proposals, accepted file reads, interruption/recovery and plain delivery. | Replace the planned spin with the actual human path, exact abridged excerpts and reproducible scratch-file commands; link the detailed ledger. |

The student explicitly confirmed the Gemini clarification before runtime
freeze. Confirmation of this broader author reconciliation remains requested;
no review completion is inferred from silence. Initial all-seven-module checks
and race results belong to the student's retained local receipts. Independent
broader checks, comparison and final receipt/prose review remain open in the
[gate record](chapter-06-validation.md).

## Comparative feedback: incremental assembly cost

Independent review found repeated rebuilding of growing parts in all three
initial stream assemblers. The initial contract specified storage ceilings and
early delivery but did not explicitly teach the cost of repeatedly copying,
decoding and serializing accumulated content. Section 6.3 now explains that
failure, requires append-oriented owned accumulation and incremental size
accounting, and calls for a size-doubling benchmark with fixed-size fragments.
This is a teaching improvement from comparative review, added before the
affected repair. It does not rewrite the initial student's experience or relax
any semantic, identity, cancellation or exact-bound requirement. The coder's
plan and revised measurements remain pending.

## Comparative feedback: replay, deadline and consumer ownership

The initial refusal parser met the retention instruction, but the renderer
could not send that recognized material on the next request. Section 6.4 now
connects retention to the explicit matching-target assistant `refusal` field
and requires both next-turn and tool-continuation checks in plain and stream
modes. Unsupported material still fails safely. The initial live runs did not
contain refusals; their absence cannot establish this behavior.

The existing §6.3 requirement already covered timeouts during capacity waits.
A short explanation now makes its ownership concrete: the deadline covers the
whole model operation, including final fragment drain, because stopping an
HTTP transport cannot wake an unrelated channel wait automatically. This is
an implementation defect with a clearer teaching explanation, not permission
to weaken the configured timeout.

Independent review also found missing ownership back-pointers and ignored
subscription errors in the public streaming example's observers. The existing
skill applies to example clients as well as library runtime objects; that rule
needs enforcement in the revised consumer. No new architectural exception is
introduced. Student confirmation and reviewed repair outcomes remain pending.

The first bundled repair at `3ccaed6` still undercounted escaped thinking data
retained as opaque JSON. Independent review supplied a passing small control
and a valid oversized opaque response. Section 6.3 now spells out that the
existing bound counts the retained representation, including JSON escapes.
This clarification preserves the original limit and the intermediate failed
repair; a decoded text-length counter does not establish an opaque-byte bound.
