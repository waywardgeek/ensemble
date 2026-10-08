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
