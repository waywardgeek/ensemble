# Chapter 6 independent review

Review in progress, 2026-10-07. Initial runtime is
`aa5f86a4782c7479ea61b0e6abcbd4163968b574`; the initial student experience and
live evidence were frozen in `f73b01e`, `d12a0cb` and `3417575` before this
comparative feedback. The reviewer authored independent grader fixtures, not
the student implementation. Earlier exposure includes Chapter 5 grading and
comparison; this is not a fresh cold-student evaluation.

The corresponding historical standard is `solutions/ch07/agent/internal/llm/`
at `515e2d883ef0994db652a031eb13e6974ba7b829`, viewed after the freeze. The new
contract and architecture remain authoritative. Historical callback injection,
model allowlists, EOF dispatch and raw frame logging are not proposed repairs.

## Consolidated findings, first round

### R1: Accepted Chat Completions refusal prevents continuation

Material behavior defect. `internal/llm/parse.go` retains a recognized refusal
as opaque material, as Chapter 6 requires. `internal/llm/render.go` rejects all
matching-provenance Chat Completions opaque parts. A complete response containing
visible text, two valid calls and recognized refusal material is accepted;
`read_file` and `list_directory` execute; the next model request then fails
locally with `unsupported opaque Chat Completions material`.

The independent `TestC6TypedCallsAndOpaque/openai` uses distinct interleaved
argument fragments and an actual file witness. The failure happens after
acceptance, not because a fixture lacks a terminal or usage. Messages and
Gemini paired fixtures pass. Raw JSON object key order is compared semantically;
ordered parts, exact strings, provenance, arguments and signatures remain
protected.

Repair the recognized refusal mapping for matching provenance while preserving
safe rejection of unsupported opaque shapes and omission for foreign targets.
Cover plain and streamed receipt, tool continuation and a subsequent user turn.
Refusal text remains opaque and cannot satisfy the visible-text-or-call rule.
Do not expose it as invented reasoning or lose other typed fields. Root owns
the corresponding explicit teaching clarification.

### R2: Incremental assembly repeatedly copies the growing response

Material design/performance defect. `stream.go` repeatedly concatenates the
entire accumulated text/arguments. Messages also decodes and re-encodes its
growing block, while Gemini repeatedly decodes, measures, merges and re-encodes
the previous text run. The work increases quadratically with accumulated
content and fragment count. A bounded final object does not bound cumulative
allocation or repeated work to a reasonable multiple of its size.

An independent benchmark supplies fixed 128-byte fragments. Doubling final text
from 64 KiB to 128 KiB gives these allocated-byte measurements:

| Adapter | 64 KiB result | 128 KiB result |
|---|---:|---:|
| Messages | 57,408,504 | 221,483,224 |
| Chat Completions | 20,731,680 | 75,767,848 |
| generateContent | 207,420,176 | 803,534,824 |

The Gemini samples took 0.395 and 1.574 seconds on this machine. These are local
measurements, not provider latency or universal timing thresholds. The separate
16 MiB assembly boundary passes without race; the Gemini case took 15.62
seconds. A combined race-instrumented boundary run exceeded its 90-second
test deadline while repeatedly decoding Gemini's preceding run. That timeout
does not establish a semantic size-limit failure.

[The benchmark receipt](checkpoint-evidence/ch06-review-assembly-scaling.txt)
records the observations. The historical adapter used append-oriented
accumulation; its unsafe completion and signature behavior should not return
with that useful idea. The student receives the rationale, not historical code.

Use operation-owned append-oriented storage for incremental strings and argument
bytes, incremental size accounting, and bounded final materialization through
the shared semantic validator. Preserve original opaque objects, signature
positions, thought/text boundaries, stable IDs, UTF-8 fragments, limits and
cancellation. Review allocation scaling after repair without imposing a brittle
wall-clock score threshold. Root owns a short teaching addition explaining why
serializing the whole partial answer on every token defeats streaming.

### R3: HTTP timeout does not cover operation capacity and drain waits

Material lifecycle defect. `Engine.client.Timeout` governs HTTP transport, but
`operation.Emit` and the final pending-fragment drain wait select on the outer
context. The actor creates that context with cancellation only. When the HTTP
deadline expires while the reader is parked in operation-owned transfer work,
closing the transport does not wake that wait.

`TestC6HTTPTimeoutIncludesCapacityWait` supplies a real local SSE response,
sets the Engine timeout to one second, and observes the operation's transfer
store actually reach 1 MiB before waiting past the deadline. The operation
remains parked and only settles after the probe explicitly cancels and discards
it. This is not a missing-network-response or fixture setup failure. The
existing explicit cancellation controls pass.

Give the whole model operation the configured request deadline, including
reading, incremental assembly, waiting for transfer capacity and final drain.
Preserve the configured duration and safe timeout diagnostics. Verify timeout
without an external interrupt, one rejected model end and reliable completion,
no acceptance/usage/effects, and a usable subsequent turn. Chapter 6 §6.3
already requires timeout during capacity waits; an ownership explanation can
make that instruction harder to overlook.

### R4: Public stream-consumer children lose the owner/logger path

Material architecture gap in `examples/stream-consumer/main.go`. Both `observer`
and `slowObserver` retain mutable callback state but have no parent interface
back-pointer. Their callback code cannot reach its actual application owner or
logger. The sample must teach the same diagnostic reachability as the library.

Construct these children through their client/application owner with a public
interface back-pointer that reaches the Ensemble-owned logger. Keep public
imports and the actual ownership relationship; do not inject an unrelated
logging closure. Check subscription errors rather than discarding them. This
is a local structure/consumer correction and does not by itself invalidate the
earlier provider exchanges, whose original identities must remain intact.

### R2 follow-up: escaped opaque payload accounting

The first repair was frozen as `3ccaed61c03356a2b07f4892d0afbb1b28d90b8f`
before this final boundary probe completed. Its Messages thinking counter
charges decoded fragment length, but the retained opaque JSON escapes that
content. A valid visible-plus-thinking response retained 18,874,423 opaque
bytes after 48 separately bounded 64 KiB NUL fragments. A small positive with
the same shape passed. This is a size-accounting defect within R2, not a new
response limit or an incomplete wire fixture.

A stronger control uses the serializer's quote, backslash, control-character,
HTML and Unicode escapes. It constructs exactly 16 MiB of retained opaque plus
visible content and then adds one byte. The exact positive passes; the next
byte is incorrectly accepted by the first repair. The independent receipt is
[ch06-review-repair1-opaque-bound.json](checkpoint-evidence/ch06-review-repair1-opaque-bound.json).
The coder is correcting incremental retained-byte accounting before paid reruns.

## Validation status

Public two-Agent/repeated-operation barriers, interrupt/cancel/close after a
complete-looking proposed write, slow/fast observer isolation, recorded-delivery
replay including absent historical delivery, exact SSE split/size controls,
and bounded UTF-8 transfer/cancellation pass on the initial source. The job
interleave fixture also passes: a real PTY job ends while the next model
operation is open, and a missing Gemini call ID uses the later actual response
sequence. New tests run in disposable copies only.

The frozen initial source passes 11 of 13 independent groups; only R1 typed
continuation and R3 timeout fail. All eleven initial targeted implementation
deletions have passing positive controls and their intended behavioral refusal.
Their source maps each contain the same 68 Go/module files as the initial
revision. The later repair passes all thirteen original groups and the unchanged
Chapter 5 assertions through an explicit plain-delivery/interface adapter
(100/100). Its newly added opaque boundary fails as recorded above. These
receipts predate the final strengthening of fixture assertions; their source
maps and captured test output remain the record of what actually ran.

Root completed a separately source-bound CLI overflow audit (`1964a54`), with
four positive cases and two intended deletion controls, and an initial live
receipt audit (`6f0af86`) covering all 33 requests without changing the 111
original run files. Full gate assembly and revised-source review remain open. No paid calls have been
made by this reviewer. These findings do not erase the initial source, student
review, timing misses or live token-limit responses.

## Comparative design and clarity assessment

The first-edition Chapter 7 adapters provide useful append-oriented assembly,
but their surrounding seams are weaker than the new contract. The historical
SSE reader dispatches pending data at EOF and recognizes `[DONE]` in transport;
the new reader leaves API completion to the API decoder and rejects unfinished
framing. Historical streaming callbacks mix delivery with engine recording.
The new operation returns owned facts, and the actor alone decides whether the
operation is current, appends its response, publishes mapped finals and starts
calls. This is a material improvement in both ownership and failure behavior.

The new implementation also improves on the historical treatment of signatures
and returned identities. Its typed acceptance retains signed empty Gemini text
and call-bound material at their actual positions, isolates exposed thinking,
checks required usage and terminal signals, and rejects conflicting returned
models. No raw-frame production trace or model-name allowlist was reintroduced.
The historical source is evidence for comparison, not authority for those old
shortcuts. This review covers corresponding Chapter 7 streaming behavior; it
does not claim an audit of later features absent from the new chapter.

R2 corrects the main avoidable cost in the initial rewrite. Owned builders and
incremental accounting make the structure explain what changes on each frame.
Materializing an ordinary response envelope once and sharing semantic parsing
costs a bounded final pass, but avoids maintaining two unrelated acceptance
policies. That tradeoff is justified; the remaining escaped-payload correction
must preserve it. Raw retained parts and builders are private assembly values,
not separately running components requiring invented service dependencies.

The ownership review follows constructors and actual calls. Agent constructs
Engine and Actor; Engine constructs operation; streamParser and SSE reader hold
the operation interface; diagnostics traverse operation to Engine, Agent and
Ensemble. Operation's mutex protects only pending fragments, byte count,
readiness state and wakeups. It releases that mutex before notifying Agent or
waiting. The actor drains a bounded amount and schedules remaining work at the
mailbox tail. Response acceptance, usage and tool dispatch remain downstream
of durable append. Existing subscriber workers retain their separate bounded
queue and reliable completion does not depend on callback acknowledgment.
The package checker discovers all present modules and spokes rather than
relying on an old package list.

The first repair's public example now creates callback children through a
client owner. Its interface exposes the actual Ensemble and release lifetime;
callbacks can reach the application-owned logger through that path. It checks
all subscription errors, retains public imports, releases the intentionally
blocked observers before application shutdown and owns their mutable snapshots
under locks. This resolves R4 structurally without changing prior paid receipt
identities.

Comments that explain transfer-notice ownership, why a transport timer cannot
wake a channel wait, conservative unknown-part retention and narrow refusal
recognition are useful and agree with the code. They should remain. There is
no reason to import historical callback machinery or remove explanatory
comments merely to shorten the solution. The appended teaching paragraphs
explain the failure mechanisms before prescribing append-oriented storage and
a whole-operation deadline; the refusal paragraph closes the acceptance-to-
replay gap explicitly. Final manuscript and revised live acceptance remain
separate from this scoped assessment.

## Repair acceptance, runtime `75bd14d`

The independent complete deterministic gate passes all 23 command groups on
`75bd14d5d2424778ebf45cb9025e9dbf314f956a`, bound to 73 Go/module files and three
testdata files. Its nested results include all 14 independent contract groups,
six client barriers, 61 wire cases, retained Chapter 5 100/100, sixteen targeted
implementation mutations, two CLI recovery mutations and all seven module
vet/tests. Public/internal concurrency probes use race detection; the large
exact-size fixtures run without it, and the student's separately retained full
main race run covers its revised source. The checker harness's seven tests also
pass. [The full receipt](checkpoint-evidence/ch06-review-final-gate.json) records
commands, source hashes and checker hashes. No historical grader was changed.

R1's actual tool continuation retains refusal in both delivery modes; finals
preserve the typed content. R3's configured deadline now terminates a producer
whose transfer store is observably full, without external cancellation. R4's
actual public ownership path is accepted as described above. R2's corrected
opaque counter passes the exact escaped-payload boundary and the one-byte-over
rejection, with a final single normalization-bound check as defense for argument
serialization. No per-fragment growing-answer scan was reintroduced.

The independently repeated allocation benchmark now measures:

| Adapter | 64 KiB result | 128 KiB result |
|---|---:|---:|
| Messages | 2,594,608 B | 5,053,736 B |
| Chat Completions | 2,965,352 B | 5,801,216 B |
| generateContent | 5,681,752 B | 11,192,200 B |

Those ratios are approximately linear for fixed 128-byte fragments, unlike the
initial source. The immutable initial rerun is separately bound in
[ch06-review-initial-scaling.json](checkpoint-evidence/ch06-review-initial-scaling.json);
individual allocation and timing samples vary. Neither benchmark establishes a
universal latency target. All R1–R4 code findings are resolved. Required revised
live demonstrations, their independent receipt audit and final manuscript
reconciliation remain open at this milestone.

## Final manuscript and evidence acceptance

Final independent proofreading accepts the complete Chapter 6 manuscript and
its outline, evidence and student-feedback reconciliation at author revision
`815a84f`. The reviewer read the complete chapter, current voice and writing
procedure, support records and student teaching review. This continues the
independent grader/comparative-review role disclosed above; it is not a fresh
student evaluation or an additional implementation author.

Three narrow proofreading findings are resolved: provider rejection of
streaming delivery is distinguished from retained refusal content; Messages
opaque blocks exclude ordinary text and typed tool calls; and the outline's
original check plan is explicitly historical, with the actual independent gate
identified. Author changes `e6c1cd3`, `6d0e8c3` and `815a84f` close these findings
without runtime changes. No material teaching or code finding remains open.

The prose retains the first edition's reader frustration with a silent terminal
and explains why a correctly buffered writer can still hide every fragment.
The revised chapter connects that consequence to an actual PTY barrier. Its
new allocation story is supported by the initial and revised fixed-fragment
measurements and makes the assembly rule worth understanding. Paragraph endings
and pace vary despite the necessary contract detail. Scoped prose lint passes
all hard checks at 6,768 words; the negation and long-person-gap soft warnings
were considered during the complete reading rather than treated as edit quotas.

The final proofread independently recomputed accepted usage from all 22 retained
initial/revised logs, verified all 186 original-file hashes against the accepted
live audits, checked the two abridged Messages excerpts in original order and
checked all 13 local Markdown link targets in the four author records. The
original nine sessions/33 requests remain bound to `aa5f86a`; the revised seven
sessions/20 requests remain bound to `75bd14d` with receipts frozen at `8c73f9b`.
The revised human terminals and public Gemini partial endings agree with the
manuscript. Root's independent replay/source-identity audit remains the evidence
for semantic reconstruction; this proofreading pass did not rerun paid requests
or broad runtime checks.

The [final proofread receipt](checkpoint-evidence/ch06-review-final-proofread.json)
binds exact manuscript, support, policy and evidence hashes, including the
complete deterministic gate and additional mutation controls. It preserves the
limits: no live thinking or overflow claim, no completed-explanation claim for
Gemini's token-limited partials, no GUI implementation claim, and no relabeling
of earlier binaries or failed repairs. The documentation-only README/CHECKPOINT
revision `c3fa758` is consistent with those identities. Required code, teaching,
live and final proofreading review is accepted; export/tag is the coordinator's
next step, and Bill's editorial approval remains separate.
