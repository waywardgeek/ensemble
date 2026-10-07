# Chapter 6 source research

Status: research/outline only, not a released contract. No student code,
grader changes, network probes, paid calls or new measurements. Sources below
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

Current external API claims have not been verified in this research pass.
Before the full contract, read current official SSE/API documentation for
completion, error, usage, tool fragments and opaque/signature handling on the
three selected surfaces. Date capability observations and retain provenance.
Do not copy historical model tables as current recommendations.

## Next action

The outline preserves the accepted Actor/Engine/Jobs ownership and explicit
human/protocol split. Ask the global reviewer for forward identity, loss and
finalization lessons, then develop schemas and literal fixtures when the
coordinator releases full Chapter 6 drafting. Chapter 4/5 validation and a
fresh new-only student context remain prerequisites for implementation.
