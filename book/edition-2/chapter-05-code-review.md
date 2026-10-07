# Chapter 5 comparative code review

Review started 2026-10-07 after the initial runtime and real-provider attempts
were frozen. New runtime: `8aa40c3`; initial receipts: `5f2576c`; final bounded
Gemini failure and completed initial student review: `525aa10`.

Reviewer exposure: this worker previously served as the independent Chapter 5
grader engineer, reading the new public APIs and implementation to adapt probes.
It authored no student implementation and did not supply historical source to
the coder. The coordinator reused this independent role after a fresh reviewer
spawn hit the thread limit. The complete coding skill was reloaded before this
review; voice and chapter-writing procedure are review authorities.

The independent checker is already 100/100 with eighteen compiling deletion
controls, plus the final same-Ensemble collection controls. That result does not
close the comparative quality or real-provider gates. No paid calls are
permitted in this review. In particular, Gemini's final control attempt failed;
receipt inspection must distinguish request validity, parser behavior, provider
behavior and any unproven causal explanation.

## Review plan and evidence boundaries

Before consulting historical implementation, the semantic review targets are:

- Actual creator identities, data owners and logger paths across Actor, Engine,
  Jobs, Registry, reports, handles, observations and clients; every package and
  executable remains inside the taught boundaries.
- Ownership across concurrent admission, durable apply, job/report transactions,
  cancellation and close; lock order, worker lifetime, cleanup and persistent
  failure must agree with claimed semantics.
- Reliable per-request completion and collection drain semantics independently
  of display subscriptions; ownership of nested values and retained objects.
- Concrete clarity, duplicate mechanisms, comments explaining invariants, and
  the cost of later extension, compared with first-edition Chapter 6 and relevant
  later lifecycle fixes. Historical shortcuts cannot override new architecture.
- Gemini request configuration, tool-call/result/hint grouping, raw response and
  safe failure handling against the published teaching and official provider
  documentation. The successful original/split diagnostic alone establishes no
  causal grouping defect.

Reuse validated independent probes where they establish a property. Add only
narrow checks needed for an unresolved finding. Broad expensive suites and live
calls are not repeated merely to produce another passing ledger.

## Initial decision and grouped findings

The initial runtime is a substantial improvement over the historical Chapter 6
snapshot, but requires two local repairs before acceptance. The 100/100 checker
did not exercise a full process input buffer or subscription storage after
draining. The new probes preserve those uncovered failures; the score was never
a claim that the comparative review had passed.

**R1, blocking: a full PTY input buffer parks the actor.**
`internal/llm/actor.go` dispatch calls `Registry.BeginSupervision`, which calls
`Jobs.Send` synchronously for `send_input`. `internal/jobs/jobs.go` releases
its mutex before `terminal.Write`, but releasing that mutex does not release
the actor executing the write. Valid input to a process that never reads can
block hints, interruption and close indefinitely.

The independent probe starts a raw-mode PTY process with a four-second bounded
sleep, waits for its setup file, and sends one MiB. A `tool_called` observation
establishes dispatch before the hint probe. The initial run failed the 400 ms
hint acknowledgement bound; it completed only after the process exited.
[Initial receipt](checkpoint-evidence/ch05-review-send-initial.json) and
[race-enabled grouped failures](checkpoint-evidence/ch05-review-initial.json)
preserve that result. A small-input control on the same process passed, including
interrupt, live-job preservation and close; see
[positive control](checkpoint-evidence/ch05-review-send-small-positive.json).

Repair requires owned cancelable process-input I/O outside the actor. Keep
Jobs authoritative for the prewrite output position and actual accepted bytes.
An interrupt may acknowledge promptly, but must settle the writer before
reporting that call complete or dispatching another effect. Already accepted
bytes are not reversible. Keep the managed process alive on turn interruption;
close still kills and drains it. A size cap cannot fix an already full buffer.
The coordinator published this clarification in §5.4 before affected code.

**R2, required lifecycle repair: closed subscriptions retain clients and accept
new workers after application shutdown.** `Ensemble.Subscribe` checks agent
existence but omits the owner's closed state. `Unsubscribe` and overflow close
the queue; after the delivery goroutine drains, its subscription record still
retains the observer/client and allocated queue through `Ensemble.observers`.
The separate closed-owner and retained-reference probes both fail on the initial
source, as recorded in the grouped receipt above.

Reject subscription creation after Ensemble close. Once an existing delivery
worker exits, release its callback and queue references and retain the small
queryable reason record. This does not promise forced termination of arbitrary
callback code that violates the existing nonblocking callback contract. The
coordinator added that distinction to §5.6 before repair. The retention probe
examines stored values after unsubscribe while preserving the status query;
it does not prescribe a particular tombstone representation.

**R3, teaching gap resolved in prose: Gemini hint placement required combining
two chapters mentally.** The initial §5.3 explained neutral hints and Chapter 2
explained Gemini user-role merging, but there was no explicit Gemini request
fragment putting them together. The coordinator's new subsection correctly
shows the model function-call part and its original signature, matching result
inside `functionResponse.response.result`, then a distinct literal hint text
part in the merged user content. It distinguishes durable `hints:[seq]` metadata
from provider JSON, acknowledgement from delivery, and one-request consumption
from retained log history. This is a teaching improvement, not evidence of a
renderer defect or an explanation of the earlier provider failures.

## Ownership, concurrency and comparison

The review inspected the new actor, Ensemble facade, collection, common
interfaces, Engine transport/parser/renderer, Jobs process/report paths and
Registry supervision. Existing independent architecture probes discover every
module and spoke, inspect parent identities and exercise runtime logger paths;
the review checks the responsibilities behind those results.

| Responsibility | New implementation and assessment | Historical comparison |
|---|---|---|
| Application ownership | Ensemble owns Agents, logger, subscription delivery and collection construction. Optional GUI modules consume the public library. No extra framework coordinator duplicates the root. | The frozen Chapter 6 Framework duplicates actor/observation coordination and exposes another merged observation queue. |
| Turn decisions | One Agent-created Actor owns activation, request identity, call dispatch, cancellation and durable acceptance; Engine remains Agent-owned transport and usage. | Frozen Chapter 6 runs HTTP synchronously inside the actor loop and allows independent run entry. Later `4023cff` corrected duplicate synchronous orchestration and lifecycle ownership. |
| Replies | Each request closes its own Done signal over an immutable completion; Wait returns a copy. Collections retain their own returned-member state and release their mutex before parking. | Frozen Ask consumes a shared observation stream, so overlapping callers can receive another request's terminal event. Later `4023cff` introduced request replies and save-before-success. |
| Worker facts | Model/report workers post owned results with operation identity; the actor rejects superseded work. Durable job facts still arrive while HTTP runs. | Frozen tool completion counting and requeueing admit stale completion errors; `469fd4a` fixes identity matching and idle requeueing. The new operation identity and independent stale-worker probe cover that rationale. |
| Reports | Jobs prepares an owned interval; actor persists acceptance before Jobs commits its cursor. Interrupted or terminal-overtaken reports are refreshed without consuming output twice. | A simple worker-completed counter does not establish which job/report owns the completion or preserve unread output across interruption. |
| Shutdown | Admission closes before queued activation; actor continues accepting owned cleanup facts; Jobs handles processes while actor joins model/report workers. Managed process output drains before terminal publication. | Later historical shutdown fixes establish single ownership and save ordering. The new contract deliberately avoids waiting forever for an arbitrary noncooperative Go tool, so the historical unbounded join is not copied. |
| Display | Typed part/state observations are copied per subscriber; overflow is visible and completion remains independent. R2 repairs storage lifetime. | Frozen observation paths silently drop full-buffer notifications and call observers synchronously. Historical `4023cff` also explains why tool output must not masquerade as model speech; the new typed events preserve that separation. |

The import star is actual: common declares shared records and interfaces;
llm, jobs, tools and eventlog implement their responsibilities without sibling
imports. Composition creates private adapters exposing the appropriate owning
Agent capability; these adapters are not bags of separately injected services.
Actor reaches Engine and Jobs through Agent. Job reaches Jobs then Agent.
Engine owns usage, Agent owns configuration, and Ensemble owns logging.

The principal lock paths are bounded admission under Actor.mu, Agent's
serialized append/apply, and short Jobs state operations. Jobs posts facts
without waiting for actor acknowledgement while holding its own mutex. The
R1 write is a useful counterexample: avoiding a mutex across I/O is insufficient
when the actor itself remains the synchronous caller. The repair must remove
that wait from the owner, not merely add another lock.

Useful comments already explain Close/activation linearization, nonblocking job
fact posting, prepared-report cursor ownership, PTY drain-before-close, and
immutable completion sharing. Keep those explanations. The new implementation
does not need a second mailbox package or another coordinator merely to match
the older file layout. String-tagged private worker messages are locally
constructed and adequately scoped here; a broad type-system refactor is not
required for this chapter. No unrelated legacy rewrite is proposed.

The compared snapshots are new Chapter 5 at `8aa40c3e840af575724a6b895cf59c060633a9b1`
and the historical Chapter 6 export recorded at
`14961aed8c08da70ba02e2048f73d9d6ff36fd81`.
Historical sources read: `solutions/ch06/agent/internal/llm/actor.go` (blob
`1aba52c70b8e9c7975d288db3e856f08653556cf`), corresponding Framework code and
`book/chapter-06.md`; later fixes `469fd4a817d94a2375ec7e6c0e7ebf9b26687077`
and `4023cff42556d56c99dec37da20f7a03c5095b12`, including actor runtime and
their lifecycle rationale. The historical manuscript includes later corrections
alongside older pseudocode, so neither its stop-on-interrupt pseudocode nor
parallel batch dispatch overrides the new contract. Findings supplied to the
student contain rationale and observed behavior, never historical code.

## Gemini evidence and provider documentation

The final older-model control attempt is
`main/evidence/ch05/controls-gemini-final-attempt`. Its second request has
`maxOutputTokens:4096`, the original call ID and unchanged thought signature,
a successful command result, and the exact pending hint as a following text
part in that user content. The response is HTTP-successful with STOP, a present
empty text part, prompt/total counts of 1912 and no `candidatesTokenCount`.
The parser rejects the absent required usage field. Chapter 2 explicitly
permits a present empty text part; empty text alone is not the parser fault.
No parser relaxation is justified to manufacture a successful demonstration.

The [GenerateContent reference](https://ai.google.dev/api/generate-content)
defines ordered Content parts, function responses and usage fields. Its field
description alone does not establish that an omitted candidate count is a
contractually valid measured zero. The
[GenerateContent signature guide](https://ai.google.dev/gemini-api/docs/generate-content/thought-signatures)
requires returning signatures on their original parts; it also distinguishes
provider turns using fresh user content. Thus a text hint can change the
provider's turn boundary while remaining inside one Ensemble actor turn.
The adapter should preserve signatures regardless. These sources support the
new teaching, but do not prove that merged or split hint grouping caused the
older empty response. Both authorized diagnostic variants succeeded.

The [GenerateContent function-calling guide](https://ai.google.dev/gemini-api/docs/generate-content/function-calling?hl=en)
is the relevant transport guide. The newer general function-calling pages use
Interactions examples; their input/result encoding must not be substituted
into this adapter. Documentation was inspected on October 7, 2026.

Bill's subsequent Gemini scope is 3.0 Flash or newer, with Gemini 3.8 Flash
selected for hints. Discovery proves model availability, not hint behavior.
The initial older-model receipts stay preserved; new live controls, EOF,
workflow and collection receipts remain required under the revised scope.
This review performs no paid calls and claims no Gemini 3.8 success yet.

## Teaching and remaining validation

The outline's story-preservation choice is sound: it retains the reader's need
to correct work while it runs and the explanation that two loops create two
owners. It omits unsupported exact refactoring durations and the old implication
that a synchronous API necessarily serializes independent Agents. The new R1
incident makes that same point concrete: a running child process does not make
the actor responsive when the actor is blocked feeding it input. Detailed
checkpoint accounting belongs in evidence links rather than the chapter's
opening. No invented human dialogue or successful live result is needed.

Reproduction command for the added local probes:

```sh
python3 scripts/edition2/ch05-review-probes.py solutions/edition-2/main
python3 scripts/edition2/ch05-review-probes.py solutions/edition-2/main --small-input --run '^TestReviewSend'
```

The probe runner copies source into a temporary directory, records source and
fixture hashes, and installs tests there. The main tree remains untouched.
The strengthened write probe checks hint and interrupt acknowledgement,
interrupted completion, a surviving process, and bounded Close. Review the
repair's partial-input reporting and deadline-hook lifetime directly as well.
After the student repairs R1/R2, rerun these affected probes and relevant
contract checks, review the revisions, then evaluate the new live receipts.
Full acceptance remains open until those results are recorded.

The final stronger probe version was also run against an archive of frozen
`8aa40c3`, independently of the changing worktree. Its
[negative baseline](checkpoint-evidence/ch05-review-frozen-strong.json)
fails the intended control, completion, close and subscription checks;
the [same-source small-input control](checkpoint-evidence/ch05-review-frozen-small-positive.json)
passes. Both receipts record source and fixture hashes. Python compilation
passed and `gofmt -l` printed nothing for the two new Go fixtures. Existing
broad grader/vet/module receipts are reused; this review adds only targeted
checks justified by the newly found defects.
