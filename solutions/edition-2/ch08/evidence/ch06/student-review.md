# Chapter 6 student review

## Initial source and read ledger

Fresh student; predecessor `edition-2-ch05-r1` (`7a6ef036322e1cf362894bd30b073fc8399a3c30`), accepted source `185ba767545af567d3d2d3917e807222ff0b2771`. Read the full coding skill, architecture ledger, Chapter 6, Chapter 1 architecture rules, Chapter 5 actor contract and workflow. Source reads so far: preceding main common types/actor, llm engine/parser/actor/renderer, CLI chat/session; remaining relevant preceding files will be recorded below. No historical answers, graders, later chapters, or reviewer research read.

## Structure plan — requested new-contract review

- Ensemble retains logger and subscriber queues; Agent retains configuration/history/Actor/Engine/Jobs/Registry. Actor alone publishes model observations and accepts durable responses. Engine owns model operations, HTTP and accepted usage.
- A private llm operation stores its common Engine interface back-pointer. Parsers and framing helpers receive that operation context and reach logging through operation → Engine → Agent → Ensemble. No callback service bag or sibling imports.
- Common declares operation interface, decoded fragment facts, expanded observation fields and parsed response local part mapping. Engine creates each operation with immutable operation/request identity and captured config. Actor retains the current operation interface and invokes its exchange worker; operation reaches actor readiness posting through Engine → Agent's interface, rather than injected callbacks.
- Operation owns parser assembly and positive local part IDs. Its synchronized pending fragment queue is at most 1 MiB; a condition/wake channel allows cancellable capacity/drain waits. One outstanding readiness notice is protected by this mutex. Actor drains at most 64 KiB, publishes owned observations, and appends one notice at the mailbox tail if work remains. No owner/mailbox lock spans waits or HTTP. Final result waits cancellably for queue drain before posting.
- Actor model identity remains distinct from durable sequence and report operation identity. Begin follows persisted request; end occurs exactly once for accepted/rejected current operation. Interrupt/close cancel producer, discard its pending fragments and reject end; stale worker facts cannot settle another handle. Finals map local IDs to actual append sequence and part position, after append and accounting.
- Request delivery is a creation-time snapshot (`stream` default, `plain` disabled), persisted separately on request_sent; absent historical delivery reconstructs plain. Stop reason is retained on response/completion for generation-limit display. Replay has no streaming work.
- Provider decoders assemble bounded ordinary-envelope equivalents and feed existing shared Parse validation, extending that normalization for Gemini text runs/refusal/stop reason. SSE reader owns bounded frame state; UTF-8 is validated on completed field content, and byte accounting includes physical terminators.
- CLI keeps one output owner in its select loop. A subscribed progress bridge delivers bounded queued observations; completion waits only for local display ordering or explicit subscriber overflow (never library completion). Chat flushes fragments, labels thinking/proposals, suppresses normal duplicated completion text and recovers authoritative final text after overflow. Protocol observation serialization uses exact new wire records while default protocol stays unchanged.

## Teaching experience (initial, ongoing)

The distinction among parser completion, response acceptance and request completion is useful. Exact physical frame limits and the readiness queue contract remove otherwise ambiguous resource bounds. No architectural ambiguity encountered yet; implementation and live experience pending.

Root accepted the structure plan against new Chapter 6, architecture and skill before ownership implementation. Baseline `go test ./... -count=1` passed. Additional permitted reads: full preceding ensemble.go, events validation sections, renderer, CLI main and preceding Chapter 5 live-plan (evidence mechanics only).


### Initial ambiguity: Gemini text plus unknown opaque fields

Chapter 6 §6.4 says an opaque field ends a text run. The Chapter 5 parser only retains text and known signatures from a text-bearing part; merely preventing merging would still lose an unknown extra field. Asked root whether such a combined part is wholly opaque or represented as text with replay payload. Proposed whole-part opaque retention without inventing a wire schema. Affected unknown-combined normalization paused while author resolves; known text/thought/signature paths continue.

### Initial deterministic results

Baseline main tests passed. The inherited Chapter 7 grader reported 0/100: its observation protocol and structural assumptions differ from the new contract, including old common concrete ownership/state_changed/model-table assumptions. Root notified; independent new-contract checks are required. First independent CLI checker passed all six early-delivery barriers (human PTY and protocol observe on each API). Pre-streaming test fixtures explicitly opt into plain delivery; their assertions remain intact. New local SSE tests cover every split of UTF-8/framing fixture, LF/CRLF/CR, BOM, ignored material, exact 1 MiB and overflow. Semantic paired fixtures cover text, exposed thinking, signatures, refusal, final usage and Gemini coalescing. Missing-terminal, unknown-delta and pending-store cancellation controls pass. These are initial development results, not the completed validation gate.

Author clarification `f085b95` read in Chapter 6 §§6.4–6.5: validates known field types, retains unknown text combinations as one opaque replay part, suppresses their deltas, treats thought:false like absence. This resolves the ambiguity without duplicating text or losing unknown bytes; implementing paired and negative fixtures.

## Initial pre-live checkpoint

Root accepted the full live feature/provider/public-action plan before any paid request. All seven main/public/optional modules pass `go vet ./...` and `go test ./... -count=1`; main also passes `go test -race ./... -count=1`. Formatting was applied only to changed/new Go files and `gofmt -l` prints nothing. New public barrier checks prove early observation with no accepted response/usage before server release; complete-looking proposed write truncation and interruption produce no accepted response, usage or file effect. The first vet pass found a test-only readiness method copying a mutex-bearing fixture by value; changed it to pointer receiver and repeated vet/tests/race. The unknown-Gemini fixture initially compared raw usage JSON field order; changed the equivalence check to semantic JSON comparison while preserving exact opaque payload assertions.

Initial inherited grader invocation began before implementation but overlapped working-tree edits, so its result is retained as an unbound development observation rather than an immutable baseline receipt. A bound rerun follows this source checkpoint. Its 0/100 output is not evidence that the new contract fails all features, nor is independent barrier success full acceptance.

## Initial student experience — frozen before comparative feedback

Implementation source: `aa5f86a4782c7479ea61b0e6abcbd4163968b574`. This checkpoint and its initial source remain immutable even if later review changes the code. The author clarification arrived before this checkpoint; no historical answer comparison or code feedback was used in this initial build.

Additional permitted source reads: preceding `internal/common/json.go`, `internal/llm/engine_test.go`, selected preceding test configuration helpers and CLI tests, and Chapter 5's student-authored `evidence.py` / `terminal-run.py` machinery. The evidence helper and recorder were adapted rather than importing reviewer-only code. The new independent client checker was executed black-box, never read. Root reported 57 passing independent wire cases, whose implementation was not inspected. No first-edition source, `agent/`, later chapter, grader implementation or author research was read.

What helped: §6.1's three completion boundaries mapped directly to the existing actor. The ownership path and actual-sequence example prevented predictive call IDs. Exact SSE byte accounting, CRLF treatment, required usage, and Gemini signed-empty-part example were directly testable. The pending-store readiness contract made the transport/display boundary implementable without growing the actor mailbox per token. The public observation table let the machine client expose records while leaving default protocol unchanged.

What needed teaching repair: the Gemini text-plus-unknown-field classification recorded above. Author's `f085b95` resolves it explicitly with whole-object retention and a distinguishing fixture. No other unresolved teaching contradiction was encountered during implementation. The hardest code was bounding and draining provisional fragments while preserving cancellation/final ordering, followed by coordinating the asynchronous client subscription with reliable completion without blocking the library actor.

Local implementation mistakes: a test readiness receiver initially copied its mutex; vet caught it. One semantic-equivalence assertion compared raw usage object key order rather than decoded semantics; changed the assertion while keeping exact part/signature assertions. A shell command used the repository-relative test pathname from the module directory and failed before writing; repeated from the correct repository root. No failing check was weakened. Existing pre-streaming fixtures now explicitly select plain delivery, retaining their behavioral assertions. Scope includes all seven Go modules; the module count is seven (main, GUI, consumer, jobs-consumer, tools-consumer, workflow, stream-consumer).

Help received: root reviewed ownership/concurrency and live coverage against the new teaching before implementation/paid calls; author supplied the opaque-field clarification; independent reviewers supplied executable black-box checks and results without answer-key code. Source/executable/launch mismatch controls started from a valid fixture and refused for their intended identity reasons before derived writes. The incremental relay passed a local first-fragment-before-upstream-release handshake before any paid use.

## Actual live experience

The Codex student drove actual human `chat` through execution-tool PTYs plus macOS `script`, on October 7 local time / October 8 UTC 2026. Bill did not run these sessions. Every run is bound to `aa5f86a`, 68 historical Go/module files and executable/support hashes in `initial-binding.json`. `launch.json`, original `terminal.txt`, raw request bodies, retained response bodies, event logs and actual `notes.txt` remain separate from derived replay files.

Nine sessions completed: streaming human chat, plain human chat, and the two-Agent public consumer on each provider. Discovery confirmed `claude-haiku-4-5-20251001`, `gpt-4.1-mini`, and exact `models/gemini-3.8-flash`; returned identities remain in logs. All 33 captured requests reconstruct semantically from the exact request prefix/configuration, including stream flags/usage option and historical plain delivery. Gemini endpoint selection is in the actual launcher/request path; replay output itself is the body. Details and counters are in `live-summary.json`; derived comparisons are in `verified-initial/receipts.json` and `verified-plain-gemini/receipts.json`.

Each stream session exposed text, acknowledged an active hint, displayed a proposed `read_file` call, retained its accepted call/result, and continued with the actual file marker/port. Each retained an interrupted operation with incomplete display and no accepted response/usage for that operation, then returned `RECOVERED-SIX` on a new turn. The three plain sessions returned `PLAIN-SIX` and the actual marker/port once, without provisional deltas. Scratch file bytes and actual tool results were inspected; no write was requested or inferred from model narration.

Timing difficulties are preserved rather than erased. OpenAI's first response completed before the next observation tool call; a longer bounded second response allowed an active hint. Gemini's first hint was correctly refused while idle. A 150-sentence prompt spent much of its output budget before a short visible answer and finished with `MAX_TOKENS`; the human UI visibly reported its generation limit. A literal integer-stream prompt made early output easy to observe and an active hint succeeded. One interrupt then lost the completion race and correctly reported `interrupted=false`. For the final bounded retry, the execution tool inspected each actual PTY read (blank human input kept reads short) and sent `/interrupt` immediately after the first `r5` fragment; it produced `interrupted=true`, incomplete display and recovery on r6. These extra requests each had a concrete reason and remain in the initial receipts.

Hint receipt, consumption and compliance remain separate: stream Anthropic received at seq 4 and consumed by its next request; OpenAI at seq 9; Gemini at seq 14. The request reconstruction checks establish wire inclusion. The model's subsequent concise tool answer is an observed result, not a general compliance claim.

Public consumers returned both independent Agent completions before the deliberately stalled subscription was released. Full identities distinguish each Agent's `r1/m1/part 1`; separate finals-only observations match completion typed parts. Small live streams did not overflow the slow subscription. Gemini's public 512-token responses hit `MAX_TOKENS` and included signed empty text parts; their accepted partial output is not described as completing the explanatory task. No selected model exposed a supported thinking delta in these runs; opaque signatures were retained, never labeled visible reasoning. Deterministic fixtures cover thinking, signature normalization and guaranteed overflow separately. The optional GUI remains an untested stub.

No production runtime change was needed during these live runs. A credential scan compared all Chapter 6 evidence against the three authorized key values in memory and found no match, without emitting credentials. No paid request was repeated merely for prose or evidence repairs.

## Pending independent gates

Initial runtime/tests/live evidence are ready for independent comparative review. The inherited grader remains 0/100 because its old observation/structural contract differs; its source-bound run is retained in `inherited-grade.txt`, not relabeled a pass. Independent client barriers pass. The broader independent public/concurrency/lifecycle/architecture checks and mutation coverage, historical comparison, review revisions, final author reconciliation and chapter acceptance remain pending. The initial learning record above precedes those findings.


## Review repair R1 — recognized Chat Completions refusal replay

Reloaded the entire coding skill, architecture ledger and current Chapter 6 before the repair. Read the expanded author dispositions in `chapter-06-student-feedback.md` and the actual-run reconciliation: these accurately retain the initial timing misses, MAX_TOKENS outcomes, thinking/overflow absence and separate validation gates. I confirm the broader reconciliation resolves the recorded teaching feedback; it does not close independent code review.

The reviewer found a real implementation omission: both parsers retained known refusal data as an opaque part, but matching-provenance Chat Completions rendering rejected it. Four public-path regressions (plain/stream × next-turn/tool-continuation) failed first with `unsupported opaque Chat Completions material`. The renderer now restores only the recognized single-string refusal shape on its original assistant message, alongside its text/calls. Foreign opaque material remains omitted; unsupported shapes, malformed refusal fields and duplicate/incorrect-role refusal material still fail safely. This is renderer behavior within llm, with the existing renderer→Engine→Agent→Ensemble logger path; no ownership, locks or lifecycle change is introduced.

The live initial runs contained no refusal fields, so they remain truthful and source-bound. This synthetic retained-material repair requires deterministic replay/continuation checks; repeating unrelated paid prompts would not exercise it. Consolidated reviewer findings and final validation remain pending.

## R2 structure amendment — requested before data-structure edits

Read the assembly-efficiency teaching added in `0c10608` (§6.3) and its author disposition. The original implementation repeatedly copied accumulated strings and, for Messages/Gemini, re-decoded/re-encoded growing content. The published correction explains a missing performance property rather than changing ownership or weakening the 16 MiB bound.

The Engine-owned model operation remains the lifetime owner; its worker's private parser state remains in llm and reaches the logger through operation→Engine→Agent→Ensemble. Replace mutable concatenated strings with per-part append-oriented builders. Messages blocks retain fixed raw metadata plus builders for text/thinking/signature/arguments; Chat Completions retains text/refusal builders and per-call name/argument builders. Final materialization occurs once before shared response normalization. These private records are parser-owned data, not new services, and remain single-worker-only; no new locks or sharing with Actor.

Gemini assembly will classify/decode each incoming part once into a private part record. Compatible unsigned text runs append into the last record's builder using cached kind/mergeability, without reading or rewriting earlier text. Signed/opaque/call records preserve exact boundaries and original raw payload. Each append returns only its incremental retained-content cost; size enforcement does not scan prior records. Plain normalization uses the same append-oriented run accumulator, then materializes once, so parity does not preserve a second quadratic path. Local part IDs stay attached to logical records and pending fragment transfer remains unchanged.

Existing 1 MiB pending/64 KiB drain synchronization, actor-only acceptance, cancellation waits and final ordering remain intact. A fixed-128-byte-fragment benchmark at 64/128 KiB on all three adapters will retain before/after allocations and time; it has no machine-specific timing threshold. Existing semantic/framing/identity/effect controls run after the repair.

### Repair reload and R4 ownership amendment

After context reload I reread the entire mandatory coding skill, architecture,
Chapter 6 and its feedback dispositions, including 5df3480. The refusal mapping
makes the retained-content replay obligation explicit. The deadline explanation
correctly identifies my missed channel lifetime; the consumer ownership finding
is an implementation mistake under the existing methodology, not a new rule.

For R4, introduce one public-example application client owner holding the actual
public Ensemble and constructing its normal, finals-only and stalled observers.
Each observer stores an interface parent exposing `Ensemble() *ensemble.Ensemble`,
so callback diagnostics reach the actual logger through that creator. The client
owns release lifetime and observer collections; observer mutexes continue to
protect callback snapshots and terminal notification uses the existing once.
No shared internal interface or implementation import is added: these are private
example client types using only the public library. Subscribe failures return
through the normal example error path, with release closed before Ensemble.Close.

### Bundled R1–R4 repair results

The coordinator accepted the R2 owned-buffer plan, R3 operation-deadline lifetime
and R4 actual public-client-owner plan before affected edits. R1's recognized
refusal survives plain/stream next turns and tool continuations; unsupported
opaque objects still refuse and foreign material is omitted. R2 stores growing
strings in private builders and classifies Gemini input once per arrival, with
cached run costs and final materialization. Input replacement removes the old
Messages placeholder's charged bytes before adding the first partial. Accepted
parts are still created by shared normalization and actor acceptance.

Fixed 128-byte fragment benchmarks (three local samples per size) are retained
in assembly-before.txt and assembly-after.txt. At 64→128 KiB, allocations changed
from approximately 57→221 MB to 2.75→5.35 MB for Messages; 20.9→76.1 MB to
3.12→6.10 MB for Chat; and 207→803 MB to 4.85→9.52 MB for Gemini. The revised
measurements show approximately linear allocation and time growth, without a
machine-specific gate. Existing typed/signature/opaque fixtures pass, and new
exact 16 MiB text-run controls pass with intended one-byte overflow refusal on
all three adapters.

R3 snapshots the Engine timeout at operation creation and derives the whole
Exchange deadline, including HTTP and pending capacity/final-drain waits.
A real HTTP regression observes the full 1 MiB store before requiring timeout,
separately tests final drain, and preserves an earlier caller deadline. Deleting
that operation deadline causes the intended capacity-wait failure; restoring it
passes. Safe timeout reasons remain distinct from caller cancellation.

R4's normal/finals/stalled observers are constructed by an actual example client
and retain its interface. They can reach the public Ensemble/logger; the stalled
callback also reaches its creator's release channel through that interface.
Every Subscribe error propagates. Release precedes Ensemble.Close on all exits.
A first module check caught a completion variable shadowing the client receiver;
that failed attempt is retained, then the corrected module passed vet/test/build.

Required local checks: all seven modules passed go vet and go test -count=1;
main passed go test -race ./...; gofmt -l on changed files prints nothing.
Independent black-box client checks passed 6/6 and wire checks 61/61 using
/tmp/ensemble-ch06-revised. The required inherited CH7 grader was rerun and
remains 0/100 under its older contract; its new receipt is retained separately
from the initial failure. This is not relabeled chapter validation. Reviewer
reports no further established findings beyond R1–R4 before this repair freeze.

### Escaped opaque bound correction after intermediate freeze

Independent review found that 3ccaed6 still counted Messages thinking deltas as
decoded strings although the accepted part retained JSON. Forty-eight 64 KiB
NUL fragments yielded an opaque payload above 16 MiB. This is my accounting
defect; the intermediate revision and its passing narrower checks are retained.
I read the additional explanation in 3792a19: retained representation, including
six-byte JSON escapes, is the correct cost. That clarification resolves the
representation distinction without changing the ownership plan or size limit.

The revised counters charge canonical initial opaque objects, the encoded cost
of each arriving thinking/signature/refusal fragment, and fixed structure when
a field is first inserted. They never rescan the accumulated prefix. Incoming
Gemini opaque/signature payloads charge their serialized representation; a single
completion-boundary check measures normalized retained parts and catches raw
argument serialization expansion. Messages replacement subtracts the same
canonical input cost it originally charged.

New meaningful fixtures cover small positives, exactly 16 MiB combined visible
plus opaque data, and one-byte overflow for thinking, newly added signatures,
and Chat refusals. Each frame stays below 1 MiB. All nine pass with the fix;
the 3cca implementation fails the thinking overflow case for the intended
reason (acceptance with nil error), retained in repair-opaque-negative.txt.
Main vet/test/race pass; race including large-bound fixtures takes about52s.
The repeated doubling benchmark remains approximately linear; retained in
assembly-opaque-after.txt (Messages2.75→5.36MB, Chat3.12→6.10MB,
Gemini5.01→9.83MB). No paid calls occurred on the intermediate repair.

### Revised real-provider demonstration at 75bd14d

After root accepted the bounded rerun plan and independent review confirmed the
escaped opaque boundary and R1–R4 source repairs, I drove all seven approved
sessions in actual PTYs on October 7 local / October 8 UTC. Runtime stayed at
75bd14d throughout. Revision1 evidence helpers and four executables were bound
before calls; nine local controls included a valid batch followed by an intended
late-launch identity failure before any derived write. Original helpers/binding,
initial 33 requests and intermediate 3cca remain unchanged.

The reruns used exactly the minimal 20 requests: four streamed human requests per
API (file tool+continuation, actual interrupted integer stream, recovery), two
requests per API in the public two-Agent consumer, and two requests for Gemini's
plain file tool+continuation. All three interruptions acknowledged true after
visible text, each interrupted turn has no accepted response/usage, and each
next turn answered RECOVERED-SIX-REVISION. All four file operations retained the
exact marker/8080 artifact. Gemini plain displayed its accepted answer once.
No extra retry was needed; the raw interrupted responses remain partial evidence.

Every public ordinary/finals-only typed part equals its reliable completion;
both agents complete before the deliberately stalled callbacks are released.
No live overflow or thinking delta occurred. Gemini's public 512-token requests
again reached MAX_TOKENS with signed empty text parts and short partial answers;
these are valid accepted partial responses, not completed 80-word explanations.
The public receipt preserves that outcome rather than spending another request.

Revision1 identity-first verification reconstructs all 20 captured requests from
logs. summarize.py checks chronology, exact file artifact, interrupt accounting,
terminal usage, public correlation and typed-final equality; live-summary.json
records results. receipts-manifest.json binds 75 raw files, including all 10 logs
and four tool artifacts. Credentials were scanned programmatically in memory;
none occur in revised evidence. The GUI remains a stub. No new teaching
ambiguity arose during these reruns; final independent acceptance and receipt
review remain the coordinator's open gates.
