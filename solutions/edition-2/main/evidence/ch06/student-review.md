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
