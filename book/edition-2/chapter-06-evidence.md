# Chapter 6 source research

Status: the complete independent deterministic gate and revised live audit
pass on the corrected runtime. Initial and revised evidence remain separately
bound. Final complete-chapter proofreading and coordinator checkpoint remain
open in [the gate record](chapter-06-validation.md). Historical research below
records the chronology and must not enter the cold student's context.

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

## Earlier contract handoff

The earlier global reviewer supplied identity, loss, actor and finalization
lessons before that thread became unavailable; the durable map remains the
source for those recommendations. The coordinator's contract review is now
accepted, as is the separate editorial proofreading. Those were the prerequisites at contract handoff; the initial student record
below now starts from validated Chapter 5.


## Student clarification: unknown combined Gemini text parts

October 7: the student identified an unspecified preservation boundary in
§6.4. Chapter 2's text-bound opaque value carries a signature; silently placing
an entire raw object there would change its replay format. The coordinator
accepted a conservative text-bearing rule: validate recognized field types,
then retain an object with additional unknown fields as one standalone opaque
part with exact provenance and no provisional text/thinking output. Known
signed text, calls and recognized thought parts keep their prior semantics.
Chapter 6 §§6.4–6.5 teach the rule and exact paired controls before affected
implementation. No provider observation establishes the fictional fixture
field. Student feedback and resolution are recorded separately.

## Initial actual-use reconciliation

Read on October 7 local / October 8 UTC 2026 after reloading the full current
voice and chapter procedure. Initial runtime is
`aa5f86a4782c7479ea61b0e6abcbd4163968b574`; evidence commits are `f73b01e`,
`d12a0cb` and `3417575`. No production runtime change occurred during the runs.
The student source/read ledger and initial account precede comparative review.
Coordinator design/coverage review and the author's `f085b95` clarification
are disclosed assistance, not historical answer exposure.

Actual evidence under `solutions/edition-2/main/evidence/ch06/` inspected:

- Full student review and live summary; nine launch records; full streamed and
  plain human terminal records on all three APIs.
- All three public-consumer terminal JSON objects parsed in full, then selected
  completion, finals-only and observation fields inspected. Large signature
  payloads were not reproduced in prose. An initial combined display truncated;
  conclusions use the subsequent structured inspection, not an unread tail.
- All 12 event logs parsed to recompute accepted usage, requested/returned model
  identities, hint receipt/consumption, interruption and read results. All six
  `cr/io/1` read artifacts contain exactly
  `CHAPTER-SIX-FILE-MARKER\nport=8080\n`.
- `verified-initial/receipts.json` records eight runs and 30 reconstructed
  requests; `verified-plain-gemini/receipts.json` adds one run and three requests.
  The author's independent semantic JSON multiset comparison covers all 33
  raw and derived bodies. Prefix-to-request assignment remains the retained
  student's verifier claim pending the independent reviewer's stronger audit;
  a multiset comparison alone does not prove that ordering.

All nine launch records exit zero. Human chat ran through actual PTYs, driven
by the Codex student, not Bill. The recorded UTC launch dates are October 8;
local operation was October 7. Returned identities: Messages
`claude-haiku-4-5-20251001`; Chat Completions `gpt-4.1-mini-2025-04-14`
(selected `gpt-4.1-mini`); generateContent `gemini-3.8-flash` (selected
`models/gemini-3.8-flash`).

| Session | Requests | Input | Cache write | Cache read | Output |
|---|---:|---:|---:|---:|---:|
| Stream Messages | 5 | 13133 | 0 | 0 | 960 |
| Stream Chat Completions | 6 | 1930 | 0 | 10112 | 1746 |
| Stream generateContent | 7 | 52233 | 0 | 0 | 10730 |
| Plain Messages | 3 | 7794 | 0 | 0 | 91 |
| Plain Chat Completions | 3 | 1480 | 0 | 2304 | 55 |
| Plain generateContent | 3 | 5571 | 0 | 0 | 163 |
| Public Messages, both Agents | 2 | 114 | 0 | 0 | 276 |
| Public Chat Completions, both Agents | 2 | 118 | 0 | 0 | 182 |
| Public generateContent, both Agents | 2 | 98 | 0 | 0 | 1016 |

§6.8 preserves the late-hint and late-interrupt attempts, the subsequent
successful interruption/recovery on every API, and Gemini's generation-limit
notice. Its public 512-token requests yielded partial accepted explanations;
no completed-task claim is made. Accepted usage excludes incomplete operations
without claiming zero provider charges. No supported thinking delta or live
subscription overflow was observed. These paths retain deterministic fixtures;
public reliable completions did arrive before stalled callbacks were released.
The GUI remains a stub.

The inherited grader's source-bound 0/100 is retained in `inherited-grade.txt`.
Its old protocol/structural assumptions differ from the new contract. Neither
that incompatibility nor initial independent delivery success substitutes for
full new-contract acceptance and mutation controls. Those checks, independent
comparison/revisions, receipt verification and final proofreading remain open.
This author reconciliation makes no new runtime acceptance claim or paid call.

Scoped reconciliation checks: `go run ./cmd/lintprose
book/edition-2/chapter-06.md` passes all hard checks at 5921 prose words;
soft negation/person-gap warnings were read without adding filler. The technical
contract remains unchanged. `git diff --check` passes. A separate byte check
confirms all six source files match their read artifacts, and the selected
Messages transcript lines match the raw terminal record.

At this reconciliation boundary the coordinator reports an independent finding:
accepted Chat Completions refusal opaque material cannot replay into a
continuation. Correction and distinguishing checks remain pending. The ordinary
live runs above do not exercise that refusal path; they are not evidence that
it works. This is a subsequent review finding, not a relabeled initial provider
failure or a reason to discard the retained successful sessions.

## Revised reconciliation, October 7 local / October 8 UTC

Reloaded the entire current voice and chapter procedure, the full current
Chapter 6 (recovering truncated combined output with targeted full-range reads),
and the student's appended repair/actual-use review. Retained the coordinator's
teaching additions through `838fc1b`, including R1–R4, the allocation lesson,
the exact retained opaque-byte distinction and independent checker invocation.
No normative fixture, code or immutable historical receipt was edited.

Revised runtime is `75bd14d5d2424778ebf45cb9025e9dbf314f956a`; live evidence is
frozen at `8c73f9b`. Intermediate `3ccaed6` remains preserved, including its
escaped-opaque undercount. No paid call used that intermediate repair. Initial
runtime `aa5f86a` and its nine sessions/33 requests remain unchanged.

Read the complete four revised human terminal records and parsed all three
public terminal JSON records. Inspected completion text, stop reasons, ordinary
and finals-only observations, and completion-before-slow-release results.
Signature values remain opaque and are not reproduced in prose. Read the full
revision README, live summary and reconstruction receipt array; parsed all
10 logs to recount 20 request events and recompute accepted usage. All four
read artifacts equal `CHAPTER-SIX-FILE-MARKER\nport=8080\n`. Independently
hashed all 75 manifest files against the retained manifest, with no mismatch.
This author inspection did not rerun a model or replace the independent audit.

| Revised session | Requests | Input | Cache write | Cache read | Output |
|---|---:|---:|---:|---:|---:|
| Stream Messages | 4 | 7909 | 0 | 0 | 108 |
| Stream Chat Completions | 4 | 370 | 0 | 3456 | 52 |
| Stream generateContent | 4 | 5672 | 0 | 0 | 182 |
| Plain generateContent | 2 | 3685 | 0 | 0 | 138 |
| Public Messages, both Agents | 2 | 114 | 0 | 0 | 219 |
| Public Chat Completions, both Agents | 2 | 118 | 0 | 0 | 178 |
| Public generateContent, both Agents | 2 | 98 | 0 | 0 | 1016 |

Selected/returned identities match the initial model choices. All seven launches
exit zero. The three streaming interruptions acknowledge true after visible
integer output; each recovery returns `RECOVERED-SIX-REVISION`. Each interrupted
operation has no accepted response/usage. Four exact file results survive;
Gemini plain displays the completed answer once. Revised public Agents complete
before stalled callbacks release and typed finals match their reliable
completions. Gemini's public outputs again stop at MAX_TOKENS with signed empty
parts. No live thinking or overflow is claimed. Refusal replay, guaranteed
overflow and exact storage/deadline failures remain deterministic evidence.

The independent complete deterministic gate receipt
`checkpoint-evidence/ch06-review-final-gate.json`, retained at `5a95578`, reports
passed on `75bd14d`. Its parsed check list includes early-delivery barriers,
wire/contract/prior-behavior assertions, deletions, CLI recovery controls,
assembly measurements and seven-module checks. Supplemental three-mutant
controls are retained at `340c678`. A large first display of the gate JSON
truncated embedded command logs; this author relies on its parsed verdict/check
list and the independent review, not a claim to have read every embedded log.

The coordinator's independent revised audit now passes at
`checkpoint-evidence/ch06-live-independent-revised.json`: 73 historical source
identities, both archived executables, all 75 unchanged raw files, 20 ordered
and multiset request comparisons, and 16 valid-path positive/negative controls.
It also checks usage, artifacts, interruption, public finals and credentials
without exposing secrets. This supersedes the pending audit status at the
initial reconciliation boundary. Final complete-chapter proofreading and
coordinator export/tag remain separate; no Bill editorial approval is inferred.

The independent live audit is committed at `15e9590`; see the
[live review](chapter-06-live-review.md) and
[audit receipt](checkpoint-evidence/ch06-live-independent-revised.json). Scoped
prose lint passes all hard checks (6756 words before the final link copyedit),
and diff whitespace checks pass. Soft density warnings were read; no anecdote
was manufactured to satisfy them. Final complete-chapter proofreading is now
requested from the independent reviewer.
