# Chapter 6 source research

Status: full draft contract/prose exists; independent review and validated
predecessor remain required before implementation. No student code,
grader changes, live API probes, paid calls or new measurements. Sources below
are author/reviewer evidence and must not enter the cold student's context.

## Reads and source boundaries

Read the full current `book/voice.md`, `book/chapter-writing-procedure.md`,
architecture decision ledger, first-edition Chapter 7, its historical coder
brief and superseded outline. Read the global map's streaming/GUI lessons,
the Chapter 7 Astra review, relevant current Chapter 5 observation semantics,
and the complete inherited `internal/grade/ch07_checks.go` and
`internal/grade/ch07_harness.go`.

`book/review-ch07-code.md` is absent from the working tree. Recovered its
historical content with `git show` at
`76a753d121a9ecc627ec615a988b6a465e949ba5`; its reported measurements bind to
the old source named inside it, not today's second edition. Combined command
output was truncated near that review's ending, so no conclusion depends on
an unread tail. The numbered review findings and corresponding checker source
support the observations below. The superseded outline was reread separately
in full after an earlier combined output truncation.

## History that changes the teaching

- `131690e5c41c5ec6454fb60e4243f373892a81a0` introduced one parser surface,
  shared framing, stable part IDs, delivery capability flags and frame tracing.
  Preserve the semantic goals, not its old ownerless parser or callback wiring.
- `b2b94c4e7162f046ce1db9881cc41ef4f7dcf6a5` added the exercise, checks and
  terminal rendering. Its `AskWatching` compatibility path must not reintroduce
  a second owner after new Chapter 5. Its reported chunk counts are historical
  fixture receipts, not new results.
- `511116d9f1c182235afc68ceb9708e9b227a4b37` added the equivalence check after
  the review found that a chapter thesis had been asserted without a check.
  Teach the actual scope of that check and add full-content fixtures rather
  than interpreting its label as proof of every field.
- `85223628d172e8ea4f0ef81dab97cc8fd53607e5` scoped cross-kind IDs to responses
  but inferred response boundaries by reused finalized IDs. That historical
  workaround is evidence to publish a real public boundary now.
- `76a753d121a9ecc627ec615a988b6a465e949ba5` preserves the author rulings and
  prose. The broad invariant requires one part per ID across every kind;
  actual checks remain narrower. New teaching should distinguish an intended
  rule, implemented checker coverage, and a deletion result.

## Directly observed checker limits

These are source-inspection findings, not newly executed mutation results.

1. `stream-deltas` counts chunks after the process completes. A student could
   buffer a response and emit many fragments afterward without proving useful
   terminal delivery. A synchronization fixture should observe output while
   the server deliberately withholds completion.
2. Harness `TextByPart` and `FinalText` use only the local part ID. Delta records
   contain an inferred response index for other checks, but text accumulation
   does not use it. Repeated local IDs can merge distinct responses.
3. `tool-params-streamed` concatenates all tool-call chunks and searches for
   finalized names. It does not compare each call's reconstructed arguments
   against its own final object. Cross-kind exclusion is not full correlation.
4. `delivery-not-content` compares the last assistant string and ordered part
   kinds/count, plus delta-count differences. It does not compare every typed
   part, opaque bytes, call arguments, provenance or usage.
5. The harness drives only the Anthropic fake and uses an old `think` tool and
   `fake-model`. New fixtures must use taught tools and compatible explicit
   identities; preserve historical reference behavior rather than rewriting
   old answers. Shared grader enhancement belongs to the coordinator.

## Corrections needed before publishing a contract

The old chapter calls IDs vendor-owned while also describing an API without
block indices. Use parser-assigned local identity with an explicit operation
scope. New Chapter 4 already assigns generated call IDs at actual durable
append; a streaming parser must not predict that sequence before background
job events finish.

The old prose equates a reader EOF with successful inference in places,
promises the same story from separate live generations, and calls parsed
frame payloads byte-complete wire logs. Keep framing, successful completion,
deterministic fixture equality, live behavior and retained trace scope separate.
The historical stateless-parser logging exemption contradicts the current
methodology and must not return.

The initial research pass had not verified current external API claims. The
subsequent official-source read below closes that drafting prerequisite; it
does not establish live model capability or validate a student implementation.
Historical model tables are not current recommendations.

## Official wire verification, October 7, 2026

Read the official sources below while drafting. The OpenAI documentation
skill was loaded for Chat Completions research. No model invocation was made.

- [WHATWG SSE](https://html.spec.whatwg.org/multipage/server-sent-events.html):
  UTF-8, LF/CRLF/CR, comment fields, multiline data and blank-line dispatch.
  Pending data at EOF is discarded by the framing specification. The chapter
  turns unfinished data into an adapter failure; it does not implement a
  browser EventSource client's automatic reconnect policy.
- [Messages streaming](https://platform.claude.com/docs/en/build-with-claude/streaming):
  block starts/deltas/stops, message_stop, ping and error events; cumulative
  usage snapshots; distinct thinking and signature fragments. The chapter's
  no-retry policy is its own lifecycle decision, not a claim that the guide
  forbids recovery strategies.
- [Chat Completions streaming events](https://developers.openai.com/api/reference/resources/chat/subresources/completions/streaming-events)
  and [create request](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create):
  indexed choices/tool calls, optional name/argument fragments,
  stream_options.include_usage, final empty-choices usage before [DONE].
  The generic platform streaming URL redirected to an overview and was not
  used as evidence for the event schema.
- [generateContent API](https://ai.google.dev/api/generate-content): SSE
  endpoint, indexed candidates, modelVersion, finishReason and usageMetadata.
  Empty finishReason denotes ongoing generation. The guide labels this API
  legacy; this chapter keeps the already-taught surface rather than silently
  migrating to Interactions. Latest-snapshot handling and orderly-EOF policy
  are explicit chapter assembly rules; no claim is made that arbitrary chunks
  have independent additive token charges.
- [generateContent thought signatures](https://ai.google.dev/gemini-api/docs/generate-content/thought-signatures):
  an empty text part may carry the signature late in a streamed response.
  Preserve that exact signed part instead of attaching it to earlier text.
- [generateContent function calling](https://ai.google.dev/gemini-api/docs/generate-content/function-calling?hl=en):
  complete functionCall name/args/id representation. The newer redirected
  Interactions guide was excluded from this adapter's contract. No partialArgs
  requirement is inferred from a different API surface.

## Draft design and checks

The coordinator accepted default streaming on supported adapters, explicit
disable, safe provider refusal without a paid retry, no frozen model-name
allowlist, and Engine-owned operation contexts. API delivery support remains
separate from model-specific thinking/content support. No new Bill ruling is
claimed. Actor retains durable acceptance; generated call IDs use actual
append sequence despite asynchronous job events.

New request.delivery records effective delivery; absent historical field means
plain, including reconstruction under a later default-on configuration. New
response.stop_reason retains supplied reasons without imposing a finish-reason
allowlist on plain parsers. Token-limit completion and incomplete JSON are
distinct; full semantic validation still precedes all effects.

The draft contains literal framing and three API text/signature streams,
paired plain-response descriptions, interleaved-tool fixture requirements,
operation/part observation shapes, opt-in protocol, human PTY instructions
and distinguishing controls. Every fixture counter and identity is explicitly
synthetic, not a new measurement. Production frame tracing is deferred;
safe owner-reachable diagnostics remain required.

After the cut and consistency pass, scoped prose lint passed all hard checks
with 4687 counted prose words. Soft negation/person-gap warnings were read;
the reader's incomplete-display problem remains explicit through the body,
and no invented story was added to satisfy a counter. `git diff --check`
passed. No code, tests or old grader files were edited by the author.
Independent full-draft review remains pending.

The coordinator's full draft review requested two bounded clarifications.
§6.2 now specifies a 1 MiB operation-owned pending-fragment store, cancellable
producer backpressure, one readiness notice, and at most 64 KiB actor service
per notice before returning remaining work to the mailbox tail. Chapter 5's
growable admission mailbox remains unchanged; accepted finals and reliable
completion cannot be lost to that fragment bound.

§6.3 now counts physical wire bytes including ignored fields/comments, line
terminators and the blank delimiter, excluding only the initial BOM. Reset
occurs at every dispatched or ignored blank-line event boundary. Exact LF,
CRLF and comment fixtures were independently counted with a local byte-length
calculation: each is 1,048,576 bytes. This is fixture arithmetic, not a parser
test. The revised manuscript passes all hard prose checks at 5037 counted
words. The coordinator subsequently read the exact clarification diff and
accepted both changes. The draft remains gated on validated Chapter 5 and
actual implementation evidence. The separate voice-v5 editorial pass preserves
the old terminal-flush explanation. Independent proofreading is now accepted,
with the final status resolution recorded in review commit `3d6effd`.

## Next action

The earlier global reviewer supplied identity, loss, actor and finalization
lessons before that thread became unavailable; the durable map remains the
source for those recommendations. The coordinator's contract review is now
accepted, as is the separate editorial proofreading. Chapter 4/5 validation
and a fresh new-only student context remain prerequisites for implementation.
