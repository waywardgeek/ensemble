# Chapter 10 student feedback and author response

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
