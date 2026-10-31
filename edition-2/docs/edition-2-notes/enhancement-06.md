# Enhancement notes: Chapter 6 (original Chapter 7)

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Author: `enhancement_author_resumed`, October 10, 2026. Final prose submitted;
grader/sample evidence reconciled, independent combined review pending.
Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## Source and preservation

Read all 642 lines of original `book/chapter-07.md`, file revision
`994bf218b31033b7dbf2740dd0187d509d2927dc`, SHA-256
`3f84dd77125f885ee7e49cfc6d7e8caf47726f5da26979af0719f5f2b5f5209f`.
Read complete ch07 student/coordinator notes and comparative review, accepted
common/Engine/parser/actor/CLI/public-shape source, and retained live evidence.
Reference endpoint remains `solutions/edition-2/ch07`, `edition-2-ch07-r4`.
Original/frozen paths, source hashes, tags and original results remain historical.
The full mandatory skill and governing carryover/voice/procedure/enhancement
plan have been read; the bounded pass supersedes the retired author-first rules.

The source's through-line stays intact: the user has a blank terminal while
a model already has something useful to show. Preserve its three-colour mess,
SSE framing pitfalls, duplication/drift argument, fresh-ID-per-chunk mistake,
OpenAI tool-first identity receipt, unknown-model fallback principle, negative
boolean explanation, disappearing trace risk, buffered-writer failure and
first-edition mutant predictions. Preserve original order and GUI motivation.
Historical 111-line/14-test reader measurements are attributed, not requirements.
The prior chapter already emits whole-part deltas, so the opening now names the
actual missing behavior: delivery while the response is being read.

## Findings incorporated

| Source/finding | Enhancement and reason |
| --- | --- |
| Callback/service architecture | Retain actual StreamCallbacks notifications; distinguish them from configuration/service bags. Engine's actual Agent parent leads to Ensemble logging. Actor queues deltas and records response facts before publishing stored finals. No redesigned sink API or second execution owner. |
| Incorrect Render/config sketch | Use accepted Render(Context) bytes signature; Agent owns configuration, transport constructs URL/headers. Exact names illustrate accepted source rather than impose another public API. |
| Response ID versus persistent display identity | Parser chooses IDs, including zero. Anthropic indices, OpenAI text 0/tool i+1 and Gemini continuation explain why IDs need not be literal vendor indices. Keep response-local scope and defer persistent GUI identity. |
| PartDelta agent label | Retain the existing optional client label; actual Ensemble envelope carries source identity and ownership. Do not claim a nonexistent removal or a second parent. |
| One Parse overclaim | One entry point plus shared semantic decoding avoids drift; a signature alone does not make duplicate policy impossible. Preserve canonical args/provenance/accounting in one finalization path. |
| SSE framing | Correct CRLF invented failure rate to separator normalization. Preserve multiline data/comments/terminator/EOF cases and existing shared reader. |
| Gemini delivery/signatures | New content arrives in response-shaped envelopes; text is incremental, tool objects can be whole. Preserve thought and trailing text signatures. Do not call frames cumulative complete answers. |
| Capability table | Historical implemented rows are not current availability or crossover proof. Unknown models fall back to plain text, separately from unsupported-media refusal. New wire shapes still require parser work. |
| Streamed tool JSON | Display actual fragments; complete authoritative JSON may be canonicalized. Compare decoded argument value where whitespace differs; execution waits for complete parts. |
| Frame trace claims | Describe ordered sanitized data payloads, not byte-complete HTTP/framing/credentials. Root owns trace; configured-key reflection is redacted before logging/forwarding. |
| Human flush and hint: prefix bug | Preserve buffered flush story, explain actual unbuffered CLI writes/display drain. Add legitimate hint:early text lost to stderr, actual origin marker repair, and gated CLI test establishing before-completion delivery. |
| Original mutation lessons | Preserve unintended globals-check failure and thinking/text check separation as first-edition evidence, not new revised-grader outcomes. |
| Missing exercise protocol | Publish core observation envelope/prompt and distinguish separate aggregate shape application, two disable variables and exact grader flags. No new runtime command or per-delta requirement for the shape application. |
| Live provenance | Retain 40-versus-one shape pair, failed OpenAI print command/accurate recall, Anthropic mtime misclassification, old Gemini 2.5 selection error and approved 3.8 busy-hint correction. No new paid matrix. |

No GUI implementation, redraw, persistent stream ID protocol, partial execution,
backpressure framework or sub-agent capability is assigned. Source/integration
proof of ownership, stale-turn rejection, finalization ordering and terminal
flush is distinct from observed shape or a numeric grade.

Reviewer preliminary precision suggestion applied: IDs may restart in another
response; globally unique IDs are also valid. Response-local uniqueness does
not mandate reuse. This narrows the claim without changing behavior.
Coordinator's boundary clarification also applied: existing `in_flight`
observations scope live grader responses; they do not by themselves provide
persistent response identity for a reconnecting GUI.

## Agreed grader contract

Coder confirmed six original weights, total 100, with inherited parity renamed
`ch5-parity` (20). `-stream-model` selects a declared Anthropic streaming fixture
row (historical default `claude-sonnet-5`), while a deliberately unlisted control
checks plain fallback. It does not add a fake-only capability row or claim live
availability. `-stream-exercise DIR` selects the public-library module, default
`ch07` inside/beside core; inherited consumer/workflow/media flags remain.

The core receives the exact configuration-question JSON prompt, emits individual
observations and respects `EN_DISABLE_STREAMING`. The separate application
provides an aggregate report/answer, uses `CH07_NO_STREAM`, and is not required
to print each delta. The grader checks all three existing delta counts and
answer delivery, but only the presence/nonnegative value of first-delta timing.
Timing accuracy and finalized-part counting remain source-review obligations;
the chapter explicitly retains the gated CLI test as separate human-flush
evidence. Tool arguments compare decoded JSON, preserving the
distinction between wire fragment whitespace and authoritative canonical form.
Final coder evidence is in [enhancement-06-code.md](enhancement-06-code.md).
Revised results are composed: accepted 100/100 PASS and legacy 70/100 FAIL.
The final five streaming checks ran again (80/80 and 70/80), combined with
unchanged corrected-parity results (20/20 and 0/20) from the initial full
revised runs. Those full runs also scored 100/100 and 70/100; the final fixture
adds thinking to response two so valid text-ID reuse is exercised. These are
not claimed as new full runs. Original 0/100 and 90/100 failures remain intact.

The coder caught eighteen changed-behavior defects and retained an attempted
missing-thinking mutation that survived because the parser supplied a valid
fallback; removing the actual notification then failed correlation. No compile
or setup failure is counted as a killed mutant. Source checks separately cover
framing, opaque retention, trace redaction and real before-completion CLI flush.

## Retained receipts

Independently summed response usage from four actual JSONL logs: initial
Anthropic 5,991/261, OpenAI 3,419/160, old Gemini 2.5 5,826/248, corrected
Gemini 3.8 2,727/936 input/output. All cache counters zero. Read the shape
reports directly: 40 versus one text delta, one finalized text part each,
1,059/1,149 ms first delta and 1,350/1,150 ms total, same numbered text,
733 input/40 output each. These two live responses happened to match;
deterministic parity does not promise identical independent generations.

The initial summary's 62/61/4 stdout reads are OS reads, not tokens. Anthropic's
raw third-input label was corrected from mtime guess to two turns/one hint.
The OpenAI command failed with exit 127 and the model recalled that failure.
Initial Gemini 2.5 was a real successful run but missed Bill's explicit model
restriction, overlooked in handoff and initial review as well as selection.

The actual Gemini 3.8 correction used waiting/release files and a received hint
acknowledgment to establish the interleaving. Two requests, eleven frames,
six text chunks, two thinking chunks, one tool frame, successful next-request
hint and HINT_38_APPLIED blue-lighthouse reply remain linked. Three stdout reads
span only 1.1 ms: not evidence of perceptibly paced typing. First response
1,241/181 and second 1,486/755 include 130/649 thinking tokens respectively.
No source change or paid provider matrix repeat accompanied this evidence-only
correction. Preserve original student 0/100 FAIL and legacy 90/100 FAIL.

## Validation status

Enhanced prose lint passes all hard checks after contract reconciliation:
5,199 words versus original 3,540. Negation/person-gap warnings remain for
independent judgment; original length/short-closer warnings are recorded.
Source-relative comparison retains 517 of 642 original lines, adds 336 and
removes 125. All local links resolve and JSON fences parse. Submitted SHA-256:
`754129c9911367845e2e3add7904dadef98cbf400ea6171a846532c58cd44f6b`.
The coder verified this exact manuscript hash and both JSON fences; the exact
core JSON prompt runs in the grader. Extracted human lighthouse and shape
1-through-20 prompts ran in binaries built from their own module roots, each
with streaming enabled and its respective disable variable. All four sample
checks passed actual answer-byte and captured request-flag assertions against
faithful fakes. Only published `/tmp` output paths moved into private scratch;
no manuscript correction was needed. Schematic Go examples were checked against
accepted source, not claimed as standalone programs. Detailed sample receipts
and composed grading evidence are linked from the coder report.
Independent combined-review disposition remains pending.
No commits, original/frozen edits, mocks, paid calls or next-chapter work.
