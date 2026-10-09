# Chapter 10 student feedback and author response

Current status: the final section responds to the initial real-use experience
frozen at `44d7627`. Earlier dated clarification responses remain in order;
their pending implementation statements describe those earlier boundaries.
See the [validation record](chapter-10-validation.md) for current chapter gates.

October 8, 2026. This responds to the initial plan's Q1–Q3, before affected
implementation. The student read the Chapter 10 contract pinned at the accepted
Chapter 9 release and preserved its questions in main/evidence/ch10/student-review.md.
The author read the complete plan. The coordinator supplied the following
decisions; these are published teaching clarifications, not new Bill rulings
or an assertion that the implementation/owner plan has passed review.

## Q1: physical record bounds and refusal precedence

Section 10.8 now explicitly selects 64 MiB for every physical session event,
including its actual framing LF when present. Standalone ordinary/skill/header
bounds remain inherited, and headers keep their previous bound in either mode.
The session reader selects its mode before event reads. Generic offline loading
can use a bounded first-record probe to recognize valid session initialization;
no unbounded ReadBytes or invalid/interior initializer can bypass the limits.

The student's proposed precedence is accepted. Skills validates its complete
candidate first and retains controlled skill_too_large without a skill transition
or Agent fault. Subsequent session log/count admission is a separate boundary,
with terminal session_limit. Ordinary session-record overflow also follows that
terminal rule. This resolves a teaching omission about two applicable layers;
it does not relax the decoded skill bounds or exclusive append contract.

## Q2: imported raw history versus saved state

Section 10.4 now names the public distinction. Events and Dump expose actual
anchor/tail records only; inspection and /history label origin_as_of and incomplete
raw history. The genuine saved watch window remains a bounded display witness,
without manufacturing absent events in raw history or events.log.

Pre-origin ReconstructRequest refuses the stable code history_unavailable.
Available-tail reconstruction starts from validated origin state and applies the
preceding actual tail, then that send's captured inputs. A full-origin log retains
its complete history. The student identified a public-behavior gap; its proposed
distinction is accepted and now printed before implementation.

## Q3: byte preservation versus semantic equality

Section 10.3 now requires exact recorded opaque/raw replay bytes and original
JSON number lexemes across the semantic snapshot. Canonical hash normalization
does not rewrite them. Exact raw/text fields compare byte-for-byte; specified
schema, identity and argument-value comparisons use parsed canonical semantics.
Thus a semantic comparison may treat 1.0 and 1 equally while reconstruction
still retains whichever spelling was recorded.

Dedicated validated raw-JSON string fields are an acceptable codec design, with
private spelling left to the student. A decoded semantic value alone cannot
satisfy a byte-preserving replay field. Validate nested raw text under its actual
schema and limits; encoding JSON as a string cannot hide invalid payloads.
This confirms the student's conservative design and clarifies the comparison's
purpose, rather than adding a second context or prescribing a private codec.

The selector/origin/canonicalization explanations were useful according to the
initial review. The difficult remaining work is validating sufficient reduced
state without disguising a whole event archive as a semantic snapshot. The
coordinator/reviewer is checking that owner/API plan separately. No failed run,
successful restart or live demonstration is claimed by this response. The student
should record whether Q1–Q3 now resolve the original questions.

## Coordinator/grader addendum: escaped Unicode validation

October 8, 2026. This is a coordinator/grader finding during source review,
separate from the student's original Q1–Q3. Go's documented JSON decoder replaces
invalid UTF-16 surrogate escapes with U+FFFD; checking raw UTF-8 alone does not
detect an ASCII escape spelling such as `"\ud800"`.

Section 10.3 now explicitly requires Unicode scalar values in decoded session
JSON strings and keys, including replay-bearing raw JSON. Reject lone high/low
surrogates before replacement; accept valid pairs and genuine U+FFFD. An escaped
backslash followed by literal u/digits remains ordinary valid text. Preserve
valid original raw bytes. Section 10.8 includes that validation in the strict
session boundary. Legacy standalone decoding keeps its prior contract. This is
published before affected grading; it neither reports a student-discovered
question nor claims that the new validation has already been implemented.

## Coordinator/grader addendum: accepted raw-record boundary

October 8, 2026. An independent local diagnostic of the preserved initial binary
created a session with legal spaced raw usage. Its untouched checkpoint retained
that spacing in context and the saved window, while replay read the writer's
compact fragment and refused the mismatch. Rebuilding from the complete log on
a separate copy succeeded. This exposed an unprinted acceptance boundary; it is
not a new Bill ruling or a reason to blame the student for that missing rule.

The coordinator selected one preparation boundary, now taught in §10.3. Validate
the candidate, prepare a bounded final one-line encoding once, and derive its
owned raw fragments from that exact encoding. Append those bytes before applying
or observing the accepted event. Formatting/escape normalization at preparation
preserves number lexemes, decoded text/manual/signature values, member/array order
and opaque semantics. It is neither canonical hashing nor floating-point decoding.
After acceptance preserve exact recorded fragments across state, watch, snapshots
and replay. This covers call arguments/opaque fields as well as raw usage.

Imported records already establish their accepted bytes; preserve them and their
original physical read-size accounting. Existing standalone behavior remains
unchanged. Keep Skills' controlled preflight before session storage admission
and preserve append-before-apply/observe. A second raw archive or blanket semantic
comparison of snapshot/prefix data is unnecessary and would weaken the contract.
No repaired runtime, successful rerun or paid call is claimed by this response.

## Q4 acknowledgment and Q5: exact durable identity maxima

October 8, 2026. The student confirms that the prepared accepted-record boundary
resolves Q4 before its affected repair, and retains its own whitespace-only
reproduction alongside the independent diagnosis. This acknowledgment does not
claim that the runtime repair or its validation is complete.

Q5 identifies contradictory wording: §10.3 previously allowed every non-event
watermark to be at least its represented maximum, while §10.4 required exact
activation/job maxima. The coordinator resolves that contradiction in favor of
exact durable maxima in both full-origin and snapshot-only cases. This is a
teaching correction, not a new Bill ruling or a fault in the student's question.

Sections 10.3/10.4/10.8 now require activation and job watermarks to equal the
maximum durable identity in complete validated semantic state, or zero when
none exists. Retired activations and historical jobs outside the watch window
count. A snapshot-only origin can represent an earlier identity whose raw event
is unavailable; the origin's semantic facts establish that maximum. It cannot
supply an arbitrary higher watermark without such a fact. Reject unsupported
high or low values with session_corrupt before mutation.

Request remains different: queued admissions may burn ordinals without a turn
event, so its captured cursor may exceed the recorded request_index maximum.
The job watermark is this session's historical maximum, not the shared Ensemble
allocator position. On resume raise that live allocator's floor to at least the
session maximum, preserve an already higher position, and keep occupied-artifact
skipping. Other Agents and skipped candidates create no historical job in this
session. Chapter 9 already makes failed/unchanged activation candidates consume
no ID and retains retired activation records, so it needs no separate exception.

The student's proposed snapshot-only lower-bound relaxation is therefore not
adopted. Validate the complete origin's represented maxima, then the available
tail; do not require absent original raw events or weaken semantic-state checks.
Please confirm that this resolves the affected above-maximum import question.

## Coordinator/reviewer addendum: controlled argument errors and exact replay

October 8, 2026. The coordinator reports two regressions in retained `8882a18`:
a duplicate-name skill-management call in standalone CH02_LOG ends in terminal
session validation instead of a paired error, and four Chat Completions
continuations reconstruct compacted argument-string whitespace. This author read
the relevant Chapters 2, 9 and 10 teaching, not those runtime artifacts; the
observations remain attributed to the coordinator/reviewer. Neither defect is
reported as repaired here.

The standalone controlled-error rule and exact replay rule were already explicit.
Chapter 9 §9.6 requires duplicate management arguments to receive
invalid_skill_arguments, consuming pending limits once without a Skills change.
Chapter 10 explicitly preserves standalone behavior and decoded string bytes.
The session extension was ambiguous: strict canonical/snapshot JSON forbids
duplicate structural keys, while argument data retains its inherited schema.
The coordinator selected the narrow clarification now printed in §10.3; it is
not a new Bill ruling or permission to weaken structural validation.

Designated raw argument data remains a bounded syntactically valid object.
Duplicate members cause the controlled tool-schema error in either mode; broken
JSON, non-object arguments, invalid scalar escapes under the applicable mode,
and malformed provider envelopes retain their existing refusal. A semantic codec
preserves ambiguous accepted text through a designated string/byte wrapper so its
host structure remains strict and its hash never collapses duplicate keys.
Correspondence falls back to exact accepted text when canonical comparison is
undefined. A first/last-key-wins map and a single-member substitute are invalid.

The literal repeated-name fixture uses name empty, because repeated name members
provide no unique valid name under Chapter 9's existing error shape. Checks must
cover both modes, once-only limit consumption, paired call/error, unchanged Skills,
and the next valid continuation. Session checkpoint/rebuild/snapshot-tail must
preserve this outcome and exact text. Structural duplicates and changed argument
substitutes remain negative controls. A separate valid spaced argument-string
fixture checks every later replay; canonical semantic equality cannot excuse
removing whitespace from inside that decoded string.

Disposition: teaching published before session-mode repair; independent review
and student confirmation pending. No runtime, grader, provider or authentication
work occurred. The full voice/procedure remain loaded from this author task;
scoped whitespace and manual voice/fixture checks pass. No existing prose-lint
executable was found in the checked temporary paths, and no compiler was invoked;
automated prose lint remains pending with the coordinator/reviewer.

## Initial actual-use response after freeze 44d7627

October 8 PDT / October 9 UTC, 2026. The author read the complete live-handback.md,
the initial real-use student-review section and current complete Chapter 10,
alongside the actual excerpts and selected original artifacts used in the prose.
The coordinator released this comparison/teaching phase only after freezing the
initial experience. Runtime `57d4aac` and support `9822b2b` retain their own identities;
this prose work neither changes them nor validates later runtime corrections.

The student's revised-local section confirms Q1–Q5 and escaped-Unicode teaching
resolved its questions. Its grouped dd1111e acknowledgment explicitly confirms
the duplicate-argument boundary and exact string replay distinction. Preserve
those confirmations alongside the original questions, including the withdrawn
arbitrary-watermark proposal. Initial real use reports no new ownership or
persistence-format conflict.

| Finding | Author response | Remaining confirmation/gate |
|---|---|---|
| Comparative Q3: the goldfish opener overstates what Chapter 9 lacks | Narrowed the opener to safe live resume; existing durable logs/offline rendering remain acknowledged. Preserved the next-morning reader stake and original-author default-load incident. This Q3 is a comparative label, separate from the student's earlier canonical-equality Q3. | Student confirms `0e754b9`; independent prose review agrees. Design-review wording corrected at `700434d`, closed at `b130e88`. |
| Actual clients made history versus live ownership and current policy clear | §10.9 now uses exact restart status, historical-job labels, current policy 1 versus old raw 0/effective 16, and the actual screenshot/DOM. | Student affirmatively confirms the complete walkthrough; source/evidence acceptance stays independent. |
| Gemini extra empty Answer/Accepted cards suggested missing text | Preserved initial cards/answers and replaced the tentative cause with the established two ordinary/two signed empty text parts. §10.9 separately attributes the `70d86f7` compact presentation and captured-data checks; raw parts/order remain unchanged. | Student confirms diagnosis and repair scope. Final corrected-source/evidence review remains independent. |
| Invalid activation-like skill names and report/file byte confusion | Retained both controlled management refusals and successful write. The resumed answer's 36-byte explanation is checked against the actual 17-byte file. §10.9 asks readers to inspect effects rather than trust narration. | No provider claim erased; no broader tool authority or relaxed validation requested |
| Combined GUI shutdown order was awkward | Spin now says terminal EOF first, then server termination; checkpoint-after-detach remains part of the actual observed behavior. First SIGTERM-before-EOF attempt and eventual exit 0 stay in the account. | Clarifies reproduction without changing lifecycle contract |
| Initial summary script KeyError and exclusive-create evidence-label collision | Kept in linked handback/evidence chronology. These were evidence-tool failures; originals and runtime were unchanged. No fictional application failure or passing first attempt is substituted. | Independent evidence audit |
| Synthetic limits seeding could be mistaken for paid usage | Spin labels six local exchanges, two per API, separately and excludes their nominal counters from real-provider totals. Public resumed real call consumption remains distinct. | Independent usage/receipt verification |

No new student code is requested by these author dispositions. The local/public
duplicate-argument repair already has its own coordinator/reviewer results; this
response does not reopen it. The student read and affirmatively confirmed the
complete `0e754b9` chapter and direct feedback. Final proofreading of the quality
addition and independent corrected-source/evidence review remain separate. The
source-bound initial evidence and later correction retain separate attribution.

## Grouped post-run quality response

Read the complete quality-handback.md and the student's grouped quality
acknowledgment/result. Q1's committed activation scalar read preserves the existing
owner/confinement and full capture snapshot. Q2 clarifies exact represented
activation/job maxima without removing the separate burned-request exception.
Neither requires another live generation to substantiate unchanged captured values;
independent local review must establish that equivalence.

Q4 changes the optional GUI presentation of present exactly-empty response text,
including both ordinary and signed parts. The chapter now describes the actual
source `70d86f7`, evidence freeze `265fe34`, separate screenshot/DOM and 43
captured-data checks without
claiming native hearing or a new real-provider run. Original `44d7627` receipts on
runtime `57d4aac` / support `9822b2b` stay untouched. This is a post-run quality
choice, not a retroactive requirement blamed on the initial student.

The student's final nuance is adopted: actual Chat Completions continuation
arguments are compact. The fifteen byte-exact real-request comparisons are not
evidence that a provider emitted the deliberately spaced local fixture. §10.9
now makes that distinction explicit. The final addition awaits narrow proofreading;
no new student implementation or paid repeat is requested by this response.
