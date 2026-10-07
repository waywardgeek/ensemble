# Chapter 5 source research

Status: initial contract and architecture reviews are complete. Implementation,
live attempts and subsequent teaching revisions are recorded in the
[current gate record](chapter-05-validation.md); this source-research account
does not itself establish live acceptance. New Chapter 5 corresponds to
first-edition Chapter 6; old architecture-repair
Chapter 5 remains absorbed into first introductions.

Read the full current voice and writing procedure, updated architecture,
first-edition Chapter 6, Chapter 2's current reducer/client contract, Chapter
4's job contract, the global map's actor findings, original Chapter 6 grader
checks and its exercise/refusal/build harness, and the Chapter 6 review in
`astra-chapter-reviews.md`. Read Chapter 21 coder review §9 and the relevant
git records below. These sources are author/reviewer evidence, excluded from
the fresh student's context. Some broad combined search output was truncated;
specific chapter/harness/review passages used for decisions were reread.

## Source conflicts and history

- `4023cff42556d56c99dec37da20f7a03c5095b12` retires the duplicate synchronous
  engine loop, adds request-specific actor completion/cancellation, and wires
  the later tool-batch limit. Its commit description distinguishes the
  actor's owner lifetime from turn lifetime and managed process shutdown.
- `2b2fedbc8b50bc688c31ab823712553e589f8a3b` updates original Chapter 6's Ask
  and lifecycle prose, but its earlier TLDR/pseudocode still stops the actor
  on interrupt and discards late results. New teaching must reconcile all
  surfaces instead of appending another contradictory correction.
- `14961aed8c08da70ba02e2048f73d9d6ff36fd81` records that the old harness
  built the wrong executable; the exercise directory had also been swallowed
  by an ignore pattern. Historical score 45→100 is that commit's account,
  not a newly measured result. New checks must identify the actual program.
- Current `ch6ReplayIsLive` checks stage strings and presence of observations,
  not byte equality of reconstructed requests. Current `ch6WakeOnce` counts
  distinct Agents with turn-ended observations, not coalesced wakeups. These
  are directly read checker limitations, not deletion-audit results.
- The later Chapter 21 review describes a stale frozen baseline that still
  used a different CLI loop. Preserve corrected ownership in every public
  client now. Its instruction to use the old live tree is historical and
  does not supersede this edition's fresh new-snapshot history requirement.

## Forward lessons

Global reviewer recommends reliable request-scoped replies, one mutation
owner, immutable worker messages, queued cancellation and shutdown admission,
and preserving late job facts without reopening a completed request. Chapter
4 already permits durable completion during HTTP; the actor cannot regress
to holding a lock throughout a wait. Keep actual append-time generated call
IDs and usage provenance attached to the request that produced each fact.

The inherited chapter's GUI-separation and reference-type introduction happen
earlier in this edition. Retain those mechanisms rather than falsely
introducing them here. Distinguish model content capability, known capability,
and provider availability. No new live media support or model capability
table is inferred from historical identifiers.

## Next action

The draft supplies turn/hint payloads, a literal replay fixture, separate
queued/active cancellation, late-fact treatment, reliable completion collection,
observer overflow and actual human-chat controls. It preserves Chapter 4's
explicit nonkillable-job shutdown policy and sixteen-request bound rather than
copying the later old200-batch setting. First independent read requested an
explicit stopping state, exact new protocol output/error records, and delayed
report-cursor consumption until acceptance. Final review also separated
intentional interrupt/cancel outcomes from fatal model/round-limit CLI errors.
All amendments are published and independently accepted. The coordinator's
architecture review accepted Agent-owned Actor in llm/common, Engine remaining
an Agent child, Jobs retaining state and report cursors, actor-serialized history,
and Ensemble-owned reliable completion collections. Report reservation/acceptance
and separation of progress from completion remain mandatory. Predecessor gates
remain before student release. Scoped prose lint has no hard failures.
No new live transcript is written before actual student demonstrations.

## Gemini hint clarification and revised live scope

After the initial attempts, Bill specified Gemini 3.0 Flash and newer as the
new validation scope and Gemini 3.8 Flash for hint behavior. The student's
`main/evidence/ch05/gemini38-discovery/receipt.json` records successful model
discovery; `model.json` lists GenerateContent support for
`models/gemini-3.8-flash`. Availability does not certify live hint behavior.

The coordinator added a detailed §5.3 request suffix and a Chapter 2 cross-link.
The ordering follows the existing new-edition mapping: completed tool results,
pending human prompt if any, then literal pending hints, with adjacent Gemini
user entries merged. The example is illustrative and preserves the distinction
between neutral request metadata and provider wire fields. It does not claim
that splitting contents fixes the earlier empty responses: both diagnostic
variants succeeded, so that causal claim was unsupported.

The [GenerateContent thought-signature guide](https://ai.google.dev/gemini-api/docs/generate-content/thought-signatures),
read October 7, 2026, requires preserving a returned signature on its original
part. Its current function-call turn rules also distinguish ordinary user text
from a function-response-only continuation. The book's actor turn and hint
consumption are local semantics, not claims about the provider's turn labels.
The general function-calling documentation now presents Interactions examples;
those request shapes must not silently replace the GenerateContent adapter.

All four Gemini modes are assigned to 3.8 Flash under the revised scope. Older
model receipts remain dated evidence and are not relabeled. Revised prose,
student confirmation and new live results require their own review.
