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

## Contract review before the student run

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
and separation of progress from completion remain mandatory. That review preceded student release and actual demonstrations. The final
author reconciliation below records the later outcome.

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

All four Gemini modes were assigned to 3.8 Flash under the revised scope. Older
model receipts remain dated evidence and are not relabeled. Their subsequently
completed runs and independent verification are recorded below.


## Final author reconciliation, October 7, 2026

Read the actual terminals for all 14 logical runs in
[`verified-gemini38/receipts.json`](../../solutions/edition-2/ch05/evidence/ch05/verified-gemini38/receipts.json),
the addendum plan, initial/revised bindings and appended student review. Parsed
all run logs to recompute usage and inspect outcome/model identity. Independently
checked all 78 captured-request hashes and 14 terminal hashes against the map;
read the public workflow files and the Gemini hint's three successive wire
bodies. The author did not independently implement a second renderer.

Eight applicable Messages/Chat Completions runs retain source
`8aa40c3e840af575724a6b895cf59c060633a9b1`. Six affected-path runs use
`959c663400b74927578a3609ce58b0a51263e654`: all four Gemini modes at discovered
`models/gemini-3.8-flash`, and scoped Messages/Chat Completions process-input
runs. The revised runtime fixes R1 blocking actor input and R2 subscription
lifetime, accepted independently at `743dca3`. Evidence freeze `469730f`
retains both source bindings; planned export source `185ba76` adds documentation
cleanup without changing runtime or receipts. Do not relabel the eight earlier
runs as executions of the revised binary.

The spin table uses the core-control logs only: Messages 28280/0/0/613,
Chat Completions 5247/0/7040/267, generateContent 45025/0/0/1423, ordered as
input/cache-write/cache-read/output. It is not a total across all runs. Selected
and returned model identities are printed separately. The coder drove these
human PTYs; no Bill participation is claimed. Public workflow and collection
are executable consumer demonstrations; GUI is still a stub.

Preserve the older Gemini HTTP-200/STOP responses with missing required output
usage, the inconclusive merged-versus-split diagnostic, the initial invented
`timeout_ms` prompt field and late idle-hint refusal. The revised Gemini request
includes hint sequence 7 in request sequence 10 after the signed call's result;
the next captured request omits it. The raw wire establishes delivery, while
the answer separately establishes observed compliance. The full-buffer and
subscription faults remain deterministic evidence, not invented vendor events.

The independent supplemental audit is
[`ch05-review-final-receipts.json`](checkpoint-evidence/ch05-review-final-receipts.json),
with observations in
[`ch05-review-live-observations.json`](checkpoint-evidence/ch05-review-live-observations.json).
It reproduces all 78 requests across 14 logical runs and verifies 204 manifest
hashes, both 59-source bindings and archived executable identities. All 247 raw
files remain unchanged. Its passing fixture plus 13 isolated identity mutations
reach the intended refusals before replay or derived writes. The original 13
student negatives instead stopped at an earlier path guard; their labels did
not establish the advertised checks, and that masked attempt remains preserved.

The manuscript's frozen `ch05` evidence links anticipate the coordinator's
verified export. Final proofreading and export/tag status belong in the linked
gate record, separate from Bill's editorial approval. No new paid call or
runtime change was made for this author reconciliation.
