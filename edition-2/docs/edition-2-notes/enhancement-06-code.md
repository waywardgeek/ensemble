# Edition 2 Chapter 6 grader review submission

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Coder: `enhancement_coder_resumed`, October 10, 2026. Initial formal submission;
independent review pending. Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

This bounded enhancement covers original Chapter 7 streaming only. I read the
full skill, carryover, original chapter, student notes/comparative review, original
grader and accepted/legacy clients. The coordinator and reviewer approved the
narrow plan before implementation. No original grader, frozen solution, runtime,
public API or future feature was changed. No mocks or paid reruns were used.

## Sources and retained history

Original Chapter 7 source revision is
`994bf218b31033b7dbf2740dd0187d509d2927dc`, SHA-256
`3f84dd77125f885ee7e49cfc6d7e8caf47726f5da26979af0719f5f2b5f5209f`.
The accepted endpoint is `solutions/edition-2/ch07` at `edition-2-ch07-r4`;
legacy comparison is `solutions/ch07/agent` with its adjacent applications.
Student history and comparative review remain in [ch07.md](ch07.md) and
[ch07-code-review.md](ch07-code-review.md).

The original accepted **0/100 FAIL** is preserved: its harness selects
`fake-model`, while the chapter explicitly requires unknown models to fall back
to plain responses. The legacy table actually declares that name streamable;
its original **90/100 FAIL** instead includes ineffective core disable handling.
Neither historical result becomes a passing original grade. The student's
initial Gemini 2.5 selection violated Bill's instruction; the recorded correction
used approved Gemini 3.8. Both that error and the later live evidence are retained.

## Revised contract and scope

`cmd/grade-edition2` now accepts `-ch 6`, with the accepted Chapter 7 endpoint as
its default. `-stream-model` defaults to `claude-sonnet-5`, verified in both local
capability tables. This is a historical fake-server fixture identifier, not a
current model-availability claim or a mandated table spelling. The unknown control
is `unknown-stream-model-edition2`, verified absent from both tables.
`-stream-exercise DIR` selects the independently built application, defaulting to
`ch07` under or beside the core. Earlier consumer/workflow/media flags carry
forward to corrected Chapter 5 parity.

The original six check weights and valid minimums remain:

| Check | Points | Preserved or corrected proof |
| --- | ---: | --- |
| stream-deltas | 20 | Existing text chunk and part minimums. |
| deltas-match-final | 20 | Reassembly for each text part, scoped to its response and final ID. |
| thinking-streamed | 15 | Existing count/kind separation plus each thinking group's content and matching finalized identity. |
| tool-params-streamed | 15 | Existing count/name check plus matching finalized ID, name and decoded JSON arguments. |
| delivery-not-content | 10 | Existing reply/kind/count comparison plus observable final content, usage, every request flag, real tool result, unknown fallback and external aggregate exercise. |
| ch5-parity | 20 | All corrected Chapter 5 checks must pass; no old logger-shape or partial-score exception. |

The two replies exercise a real `run_command` call with
`{"command":"printf STREAM_TOOL_RESULT"}`. Request two must carry the nonerror
result for `call_ch7_1` with that marker; a scripted final answer cannot conceal
failed execution. Both replies contain thinking then text, positively exercising
legitimate reuse of text ID 1 in different responses. Existing `state_changed`
to `in_flight` starts the comparison scope in both implementations. A missing
boundary fails; the grader does not guess a new response from a repeated ID.
Internal comparison aliases allow reuse of the original ID-keyed checks without
changing runtime IDs or inventing a GUI identity protocol. Job-result finals do
not masquerade as provider parts.

Text content compares exactly. Tool arguments compare as decoded JSON so harmless
whitespace/key order is free. Thinking checks its scripted content and finalized
identity; complete opaque replay remains source/integration proof where the
observation envelope does not expose it. Correlation checks run in both directions:
a group needs its final, and a final needs its group. Duplicate finalized IDs in
one response fail. Expected measured totals are input 310, output 88 and zero
cache counts in both modes. Every known-mode captured request must carry the
selected streaming flag; the absent-model control must complete with all requests
plain. Core `EN_DISABLE_STREAMING` and application `CH07_NO_STREAM` retain their
distinct existing contracts.

The separate application is built from its own module root, and the existing
public-dependency check verifies resolution to the submitted core. Its existing
`stream_stats` aggregate must report the actual answer, all three delta kinds
(stream minimums text 8/thinking 4/tool 3; plain counts 2/2/1), and nonnegative
numeric `first_delta_ms`. This checks reported shape and presence, not timing
measurement accuracy or finalized-part count accuracy. Those, genuine public
composition/owner access, early CLI flushing, SSE framing and full opaque replay
remain explicit source/integration obligations. No per-delta application output
protocol was imposed on the accepted aggregate-only application.

## Actual baseline and final results

| Implementation | Original Chapter 7, actual full run | Initial revised full run | Final revised composed result |
| --- | --- | --- | --- |
| Accepted frozen Chapter 7 | 0/100 FAIL | 100/100 PASS | 100/100 PASS |
| Legacy Chapter 7 | 90/100 FAIL | 70/100 FAIL | 70/100 FAIL |

The final results are explicitly **composed**, not additional full runs: the final
five streaming checks were executed again (accepted 80/80, legacy 70/80), then
combined with unchanged corrected-parity results from the full revised run
(accepted 20/20, legacy 0/20). The initial fixture had thinking only in reply one;
the final fixture adds thinking to reply two to exercise valid ID reuse. Final
component artifacts and both full initial runs are preserved separately.

Accepted delivery produces 36 text, 54 thinking and 14 tool deltas (104 total)
versus five plain deltas, five provider finals and the same 55-byte final reply.
Actual captured requests demonstrate the mode flags and successful tool result.
Legacy still sends streaming requests with core `EN_DISABLE_STREAMING`, produces
104 deltas in both core modes, and rejects the genuinely absent model. Its separate
application does honor `CH07_NO_STREAM`, but reports two plain tool deltas
(name plus arguments), rather than the one whole-part notification. These are
observed failures, not model-name incompatibility guesses. Corrected inherited
parity failures remain in the full result; no legacy compatibility shim was added.

## Validation and exact published examples

Final isolated affected checks passed:

- `go test ./internal/gradeedition2 ./cmd/grade-edition2 -count=1` (6.757 seconds;
  the command package has no tests).
- `go vet ./internal/gradeedition2 ./cmd/grade-edition2`.
- `gofmt` and whitespace checks; final submitted grader bytes match the overlay
  hash manifest. Subsequent edits before this report were comments only.

Focused accepted source checks also passed with `-race -count=1`:
`TestStreamingDeliveryAndContent`, `TestUnknownModelAndStreamFailures` and
`TestAPIFrameRedaction` in `./ensemble`; `TestChatFlushesChunksAndSeparatesChannels`
and `TestSSEFraming` in the root/`./internal/engine` packages. These exercise real
parsers/actors/CLI with faithful provider fakes. They preserve source-level proof
of vendor handling, opaque retention, framing, early flush and the corrected
legitimate assistant text beginning `hint:`. Existing 26 student mutation records,
the thinking-channel correction and two CLI hint-origin controls support unchanged
claims; the recorded live vendor exercises, including approved Gemini 3.8, are reused.

The final manuscript SHA-256 is
`754129c9911367845e2e3add7904dadef98cbf400ea6171a846532c58cd44f6b`.
Both JSON fences parse; the exact core JSON prompt is used by the grader. A
scratch sample check extracted the actual human CLI prompt (`write a short story
about a lighthouse keeper who collects fog`) and the actual shape application
argument (`Return only the numbers 1 through 20 separated by spaces.`). It built
each module and ran both enabled and respective disabled modes against faithful
fakes, checking actual answer bytes and captured request flags. All four passed;
`published-samples.json` preserves stdout/stderr/request evidence and
`published-samples.txt` records the 0.464-second check. Published `/tmp` binary
locations were redirected into private scratch paths; module-root builds,
arguments and mode variables were otherwise retained. Schematic Go examples were
checked against source, not presented as standalone compilable programs.

## Changed-behavior mutation evidence

Eighteen actual defects were caught; each has a patch, raw output and entry in
`mutations-final.json`. Unrelated compilation errors do not count as kills.

| Defect | Intended observed failure |
| --- | --- |
| Shared text/thinking delta ID shifted | Text and thinking group/final correlation fail. |
| Tool delta ID shifted | Tool group lacks matching final. |
| Thinking content corrupted | Thinking content differs from provider fixture. |
| Tool argument fragments corrupted | Decoded streamed arguments differ from final. |
| Final tool arguments changed | Final arguments differ from streamed content. |
| Plain finalized content changed | Delivery content differs despite unchanged final reply/kinds. |
| Duplicate final ID within response | Provider identity collision rejected. |
| Request-start state omitted | Delta before known response boundary rejected. |
| Reported usage zeroed | Measured input/output totals fail. |
| Unknown model guessed streamable | Unknown control requests streaming and emits many chunks. |
| Actual tool result corrupted | Captured correlated result lacks the real marker. |
| External first-delta time omitted | Required nonnegative numeric observation absent. |
| External module resolves another core | Submitted-core resolution check fails. |
| External thinking count omitted | Aggregate thinking shape fails. |
| External tool count omitted | Aggregate tool shape fails. |
| Second thinking notification group removed | Final lacks matching group; external count also fails. |
| Only request two switched to plain | Only delivery fails its per-request mode assertion. |
| Cache usage invented | Zero-cache measured totals fail despite correct input/output. |

The second-request control is particularly useful: text/thinking aggregate
minimums and all final-content checks still pass, so the new per-request assertion
is what detects the defect. The valid baseline itself proves IDs may repeat
across responses without collision.

One attempted missing-thinking mutation **survived**: suppressing the parser's
delta notification left accumulated content intact and triggered its legitimate
whole-part fallback, so the group still existed.
`second-thinking-fallback-survivor.patch/.txt` retain that result. Moving the defect
to the notification callback removed the actual group while leaving parsing
intact; the intended correlation check then failed. The survivor is not counted
as a kill. Original chunk minimums remain aggregate minimums, not an invented
minimum for each individual part.

Other setup corrections are retained honestly: an initial compile used the wrong
fake request type (`Request` instead of `Recorded`); an absent-model mutation
pattern did not match the actual zero-value lookup and stopped before editing;
and a scratch legacy probe initially looked under the core rather than beside it
for the external module. Each setup was corrected before counting results. The
text-ID mutant changes the shared text/thinking callback, so both failed checks
are reported rather than calling it a text-only control.

## Size, comments and review handoff

Counting nonblank production lines excluding comment-only lines, new
`ch06.go`/`ch06_stream.go` contain **333 active + 37 comment-only, 383 physical**.
The new real integration test has **18 active + 2 comment-only, 22 physical**.
Original harness/checks remain **533 active + 79 comment-only, 677 physical**,
unchanged and reused for their valid predicates. Command dispatch changes add
16/delete 9 physical lines. Counts separate tests; mixed code/comment lines count
as active. This is a narrow versioned driver and comparison layer, not a runtime
rewrite or generic grading framework.

The roughly 10% comment share is below Bill's approximate 20% teaching target.
Useful comments explain request-boundary ordering, comparison aliases, job-final
exclusion, JSON semantics, two-response fixture reuse, measured counts and limits
of aggregate proof. Exported symbols have documentation. Reviewer assessment of
clarity remains required; padding fixture code to reach a quota would not explain
more behavior.

Raw local evidence is
`/var/folders/p_/11vk6y114xq7k9g5r_pc3yv80000gn/T/ensemble-enhancement06-hmxgeevc`
(pointer `/tmp/ensemble-enhancement06-current`). Its `validation/` worktree uses
committed **e7f40fa** plus only the submitted command/Chapter 6 grader overlay;
`validation-overlay.json` records all four final SHA-256 hashes. This avoids
unrelated concurrent original-edition edits. `accepted-revised06-final-composed.json`
and `legacy-revised06-final-composed.json` contain reviewable final component
checks; `accepted-final-components.txt`/`legacy-final-components.txt`, captured
requests, full original/revised JSON, patches and sample receipts remain beside them.
Disposable mutant and original-grading source copies were removed. The stable
validation worktree remains for independent review until coordinator checkpoint.
Coder harnesses are idle; the reviewer owns the harness now. No Chapter 7 work
is authorized by this submission.
