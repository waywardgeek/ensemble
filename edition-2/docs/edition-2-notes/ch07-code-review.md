# Chapter 7 comparative code review

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Reviewer: `review_restart_ch01`, October 9, 2026. **Accepted with documented
exceptions after two correction rounds (maximum three).** Both initial
implementation findings and the subsequent approved-model Gemini evidence gap
are resolved. The original grader remains **0/100 FAIL**, including its round-1
rerun. Its fixture conflict and inherited exceptions are justified below; this
acceptance does not change that result. Initial findings and correction history
are retained.

Read the entire `.agents/skills/ensemble-coding/SKILL.md`, carryover, workflow,
original Chapter 7 and [completed student review](ch07.md) before deciding.
Chapter SHA-256 is `3f84dd77125f885ee7e49cfc6d7e8caf47726f5da26979af0719f5f2b5f5209f`.
Compared current `solutions/edition-2/main` with accepted frozen `edition-2/ch06`
at `8dc6dd8`, and original `solutions/ch07` with `solutions/ch06`.

## Initial required corrections (resolved in round 1)

1. **Do not identify hint acknowledgements by assistant text.** In `chat.go`,
   `strings.HasPrefix(v.Chunk, "hint:")` routes any matching chunk to stderr and
   skips ordinary reply delivery. Now that actual model text streams through
   this path, `hint:` is also valid assistant content. I built the submitted CLI
   and drove a local SSE peer through its real HTTP/parser/actor/display path:
   one text part arrived as `hint:early` followed by `late`. The process exited
   successfully with stdout **`late\n`**, while **`hint:early\n`** went to stderr.
   The complete reply is `hint:earlylate`. This violates §7.9's reply channel and
   makes visible content depend on arbitrary chunk boundaries. Distinguish an
   actual human-hint acknowledgement from model content without reserving a
   prefix in the assistant's prose. Preserve live hint acknowledgement and add
   a focused real-CLI regression that catches this routing error, including the
   corresponding targeted mutation. No new display framework is needed.

2. **Complete the exported-symbol documentation rule for four new tests.**
   `TestStreamingDeliveryAndContent`, `TestUnknownModelAndStreamFailures`,
   `TestAPIFrameRedaction` in `ensemble/stream_test.go`, and `TestSSEFraming` in
   `internal/engine/sse_test.go` have no declaration documentation comments.
   Bill's rule applies to exported Go symbols, including these functions.
   Add brief purpose/contract comments; the useful comment inside the redaction
   test does not document its declaration. An independent AST inventory found
   these four omissions, with existing inline field comments accounted for.
   No broad comment rewrite or padding is requested.

## Design, scope and comparison

The implementation preserves one parsing seam for both delivery modes. Vendor
accumulators reconstruct complete provider messages and reuse the existing
semantic decoders, keeping provenance, usage normalization and canonical tool
arguments in one place. The shared SSE reader handles framing, while vendor
meaning stays in Engine. Plain responses emit one delta per displayable part;
streamed responses emit actual arriving fragments. Gemini tool objects remain
one complete chunk rather than fabricated incremental arguments.

IDs are chosen by the parser and returned with finals. OpenAI tool-first delivery
cannot collide with later text; Anthropic keeps block indices; Gemini infers
continuation locally and preserves trailing signatures. Actor publication follows
history insertion and uses the stored parts, including assigned missing call IDs.
The inherited nonzero-ID assertion was properly removed: zero is a valid
response-scoped ID, and the new grouped-content assertions protect the relevant
identity contract. The later GUI response-boundary problem remains explicitly
deferred, without resurrecting Chapter 6's withdrawn grouped-wait requirement.

HTTP workers queue deltas with their turn identity; the actor rejects stale
progress after interruption. Tools still execute only complete finalized calls.
Observers and private request replies keep their separate purposes. The display's
small queue lets the terminal drain without blocking the actor or dropping reply
bytes; no terminal redraw or other Chapter 8 work was added. `AskWatching` uses
the existing actor path rather than introducing competing orchestration.

Ownership remains concrete: parser methods reach their Engine's Agent; Engine
reaches the root trace/logger through its actual parent chain. The root owns the
optional trace writer and serializes records; callbacks carry notifications rather
than sibling-service dependencies. Common holds the shared enum/interfaces/data
and JSON dispatch, while framing, lookup policy and vendor behavior stay in the
Engine spoke. The separate public consumers and GUI stub remain outside core.
No mocks, mutable application globals, early sub-agent features or new generic
frameworks were introduced.

The original Chapter 7 core Go source is identical to its Chapter 6 snapshot:
streaming had already landed there. Its separate streaming accumulators duplicate
more final-part construction, whereas the student reuses its existing semantic
decoders. The original also retains common-package behavioral helpers and the
previously reviewed competing actor/synchronous paths. Those inherited differences
must not be counted as new Chapter 7 improvements. The student's compact SSE reader
and small standalone shape consumer are proportionate to this chapter; the
larger integration tests protect behavior absent from the original core tests.

New comments explain tool-first IDs, frame logging, signature preservation,
finalization order and immediate terminal delivery. Core production comments are
19.4% of nonblank lines, near Bill's approximate target; original core is 24.6%.
The important reasons remain beside their mechanisms. The four documentation
omissions above need correction independently of those percentages.

## Counts and chapter growth

Each physical Go line is blank, comment-only (leading `//`), or active, including
imports, braces and literal data; no block comments complicate these trees.
Mixed code/comment lines count as active. Tests, core, GUI and examples are
separate. Counts exclude module manifests/sums, README prose, evidence, logs and
binaries; there is no generated Go or new non-Go implementation. `x/term` and its
indirect `x/sys` dependency support actual terminal detection. Module directive
normalization is support work, not another feature.

| Go category | Active | Comments | Physical | Delta from own Chapter 6: active / comments / physical |
| --- | ---: | ---: | ---: | ---: |
| Original core production | 4,883 | 1,595 | 7,153 | 0 / 0 / 0 |
| Original ch05 example | 99 | 15 | 125 | 0 / 0 / 0 |
| Original ch06 example | 261 | 31 | 328 | 0 / 0 / 0 |
| Original ch07 example | 249 | 45 | 325 | +249 / +45 / +325 |
| Student core production | 3,918 | 944 | 5,158 | +713 / +103 / +855 |
| Student GUI production | 8 | 3 | 13 | 0 / 0 / 0 |
| Student ch05 example | 74 | 13 | 92 | 0 / 0 / 0 |
| Student ch06 example | 267 | 20 | 295 | 0 / 0 / 0 |
| Student ch07 example | 74 | 4 | 82 | +74 / +4 / +82 |
| Original core tests | 548 | 125 | 728 | 0 / 0 / 0 |
| Student core tests, including CLI | 2,106 | 159 | 2,335 | +505 / +14 / +534 |
| Student GUI tests | 29 | 2 | 34 | 0 / 0 / 0 |
| Student ch05 tests | 72 | 12 | 87 | 0 / 0 / 0 |
| Student ch06 tests | 132 | 13 | 147 | 0 / 0 / 0 |

Neither ch07 example has its own automated tests. Combined student production is
**4,341 active / 984 comments / 5,640 physical**, versus original
**5,492 / 1,686 / 7,931**. Combined student tests are **2,339 / 186 / 2,603**,
versus original **548 / 125 / 728**. Student production growth is
**+787 / +107 / +937**; test growth is **+505 / +14 / +534**. The original's only
Go growth is its new example. These totals reflect different implementation timing;
they are not evidence that the original implemented streaming in zero lines or
that the student's required streaming work is bloat.

## Verification, live evidence and exceptions

Independently ran core formatting, vet and uncached race tests: PASS. Separately
ran vet and race tests in GUI, ch05, ch06 and ch07: PASS; ch07 reports no tests and
is compile/vet coverage only. Core tests exercise real translation, actor, history,
tools, jobs and CLI subprocesses against small external HTTP fakes. Gates require
an actual observer or stdout chunk before the server can finish, so incremental
delivery is tested rather than inferred from the final answer. Streamed/plain
facts, usage, replay bytes, channel identity and tool finalization are compared.

Read the complete [mutation output](../../solutions/evidence/ch07/mutations.txt):
26 distinct defects, 27 executions including the corrected reasoning-channel
survivor. The original survivor is honestly retained; its rerun fails the added
semantic channel assertion. Intended liveness failures have working baselines.
The partial-execution mutant first fails the intended argument assertion and then
the related unmatched-result check. No compilation failures, race reports or
unrelated cleanup failures are represented as successful kills. The independent
CLI reproduction above exposes a remaining case the existing tests miss.

Inspected retained API frames, durable events and timestamped CLI reads alongside
the student summary. Anthropic shows actual incremental story/reasoning delivery,
two human turns and one hint; the student's correction of its mtime-based capture
labels is supported by the durable log. OpenAI shows three turns and correctly
reports/recalls the shell command's exit 127 instead of inventing success. Gemini
shows three turns, reasoning summaries, a complete tool-call chunk and successful
output recall. Their traces retain streamed frames and measured usage. The public
shape exercise's live pair returns identical numbered text and parts with **40
versus 1 text deltas**, and first-delta times of 1059ms versus 1149ms. No paid calls
were repeated during this review.

I independently agree with the student's requested fixture exception. Original
§7.7 explicitly names `fake-model` among unknown models and requires successful
plain fallback. Inspected `internal/grade/ch07_harness.go:driveCh7`: both runs set
that exact model while the checks demand streaming. The implementation correctly
does not add a hidden grader-only capability row. Its declared-model tests and
live evidence demonstrate the actual streaming behavior, and its unknown-model
test demonstrates fallback. Adding `fake-model` to the table would contradict the
chapter rather than fix this fixture. In the later author–coder phase, use a
declared streaming model for those five checks and retain a separate unknown-model
fallback check; preserve existing valid per-part/channel/content coverage.

The two actual student grades remain **0/100 FAIL**, not passing results. The
coordinator's frozen-reference baseline was **90/100 FAIL**, with the disabled
run still streaming. That baseline is not evidence that the student's opt-out is
broken: the student independently exercises both `DisableStreaming` modes and
supports the supplied CLI environment convention. I did not repeat the unchanged
full original grader; its complete final output, launch source and independent
behavior checks establish this assessment without another redundant run.

Chapter 6 parity still fails only the accepted logger AST-shape and text-only
unknown-model checks. Those exceptions persist, with actual failures visible.
URI/handle attachments remain deferred; required local-path media and metadata
retention are not waived. Canonical final tool JSON versus verbatim transient
argument whitespace preserves the inherited data contract and the chapter's
authoritative-final/display distinction; it is not permission to lose arguments.

The grader exceptions do not excuse the reply-routing defect or documentation
omissions. Correct those bounded findings and rerun affected checks/mutations;
no new paid matrix, grader feature, observer framework or future capability is
required. Review scratch binaries and local peers were cleaned. No production,
original solution, grader, frozen export or tag was changed by this reviewer.

## Correction round 1

Re-read the full mandatory `.agents/skills/ensemble-coding/SKILL.md`, carryover
and updated student review before reassessment. Both initial findings are
resolved. `PartDelta.Hint` records the actor's actual acknowledgement origin;
only the `hint_received` publication sets it. The CLI branches on that bit,
leaving ordinary model text untouched. This adds one documented boolean to the
existing observation, without another event framework or guessed text convention.
The four requested test declaration comments now explain the behaviors protected.

The compiled-CLI regression sends `hint:early`, withholds the rest of the HTTP
response until stdout is observed, then requires exactly `hint:earlylate\n`.
The existing HTTP-hint and interrupt tests also inspect the actual acknowledgement's
marker while preserving their delivery/lifetime assertions. Read complete
[round-1 mutation output](../../solutions/evidence/ch07/round1-mutations.txt):
restoring prefix routing fails the intended incremental stdout check; removing
the actor marker fails the direct `hint acknowledgement missing origin marker`
assertion. The earlier deadline-based version remains honestly recorded. Final
failures contain no incidental race, build or cleanup failure. Independently ran
uncached race tests selecting the compiled CLI, HTTP hint and both interrupt
tests: PASS (root 1.774s, integration 1.103s). The student's full core race/vet/
formatting result and unchanged initial module/live checks are reused; no paid
calls or broad mutation sweep were repeated.

Using the same counting method, final round-1 core production is **3,919 active /
946 comments / 5,161 physical**, a correction delta of **+1 / +2 / +3** and a
Chapter 6 delta of **+714 / +105 / +858**. Core tests are **2,114 / 171 / 2,355**,
correction delta **+8 / +12 / +20**, chapter delta **+513 / +26 / +554**. Other
categories are unchanged. Combined production is **4,342 / 986 / 5,643**, chapter
growth **+788 / +109 / +940**; combined tests are **2,347 / 198 / 2,623**. The
initial comparative assessment and original-solution counts remain applicable.

Read the complete [round-1 original grader result](../../solutions/evidence/ch07/grade-round1.txt):
**0/100 FAIL**, the same five unknown-`fake-model` streaming failures and two
inherited Chapter 6 parity failures. The scoped exceptions above remain justified;
neither the fixes nor their acceptance turns that result into a grader pass.

**Sole remaining correction: use Bill's approved Gemini model for live evidence.**
The coordinator identified an overlooked standing instruction: Gemini must be
3.0 Flash or newer, subsequently specifying 3.8 Flash for working hints. The
retained Chapter 7 summary, CLI capture and event/API evidence use
`gemini-2.5-flash`, with three human turns and no busy hint. They establish the
reported old-model behavior but do not satisfy Bill's selection instruction.
This was missed by both the coordinator's handoff and my initial review; it is
not solely a student error. The Chapter 6 3.8 run predates this chapter's new
streaming implementation and cannot replace its live evidence.

Request only a bounded own-CLI **Gemini 3.8 Flash streaming and busy-hint
demonstration**, preserving the old run and recording the model-selection error.
The current capability table already declares `gemini-3.8-flash` with streaming
text/thinking support; no unrelated model-table migration, other-provider repeat,
new feature or full verification matrix is requested. Review the actual frames,
visible delivery and acknowledgement/next-request hint behavior before acceptance.
This is the sole remaining correction for round 2. No reviewer scratch or
processes remain; no production source or original artifacts were modified.

## Correction round 2 — accepted with documented exceptions

Read the full mandatory skill and updated student review, including the new
carryover model-selection clarification. This round changes evidence and notes
only. The three recorded CLI/actor source hashes match the current files; prior
code assessment, tests, mutations and round-1 counts remain applicable.

Independently inspected the [discovery result](../../solutions/evidence/ch07/gemini38-discovery.json),
[raw CLI capture](../../solutions/evidence/ch07/gemini38-cli.json),
[API frames](../../solutions/evidence/ch07/gemini38-api.log) and
[durable events](../../solutions/evidence/ch07/gemini38-events.jsonl).
Discovery reports the requested model available. Both durable request destinations
name `gemini-3.8-flash`, and all eleven response frames identify that same model.
There are two requests, six nonempty reply chunks, two thinking chunks and one
complete tool call. Reassembling reply frames equals captured stdout and the
authoritative final text. Final vendor accounting agrees with retained usage:
1,241 input / 181 output, then 1,486 / 755, including thinking tokens.

The real command creates the waiting marker before the launcher sends its hint.
CLI acknowledgement precedes release of that command. Independently, durable
sequence 4–7 is `tool_called` → `hint_received` → `tool_returned` → `request_sent`.
The second request actually contains both the successful tool result and the
new blue-lighthouse instruction. Its reply starts `HINT_38_APPLIED` and describes
the blue lighthouse. Thus this is a hint during real active work that reaches the
next request and changes the answer, not a new idle turn mislabeled as a hint.
The tool and CLI both exit 0 without timeout.

The three nonempty stdout reads span only 1.1 ms. Accept the student's explicit
limitation: this live run demonstrates streamed frames, separately routed thinking
and correct visible content, not perceptibly paced typing. The unchanged gated
compiled-CLI test independently establishes flushing before provider completion.
No paid repeat is needed. A local value comparison against the configured Gemini
credential found zero matches in all five new artifacts without printing the key.

The approved-model live-evidence correction is resolved. The original 2.5 run and
selection/handoff/review error remain preserved. **Chapter 7 is accepted with the
specific unknown-model streaming-fixture and inherited parity exceptions above.**
Original grader **0/100 FAIL** remains the recorded outcome; URI/handle attachment
deferral remains explicit and does not waive local-path media. No architecture,
scope, implementation or evidence finding remains open. No source, grader,
original answer, frozen export or tag was changed by this reviewer, and no
scratch files or processes were created in this reassessment.
