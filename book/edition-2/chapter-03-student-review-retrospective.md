# Chapter 3 student teaching review, retrospective

Recorded October 7, 2026, after Chapter 3 acceptance, at the coordinator's
request. The writer is now preparing independent Chapter 4 checks and has
read that chapter, the global review map, and checker sources. This report
is retrospective feedback from the preserved Chapter 3 human-client work;
it is not a pre-comparison cold-student review of the complete Chapter 3
implementation. The writer did not build the original six-tool implementation.

The integration began with Chapter 3 checkpoint `7cbbd8e2` and the reviewed
Chapter 2 human client at `ad0d80e3`. The entire new Chapters 2 and 3, Chapter 1
architecture, architecture decisions, and mandatory skill were read before
implementation. Source reads are recorded in
[the human-client ledger](../../solutions/edition-2/ch03/evidence/ch03/human-chat/source-read-ledger.json).
The initial integration and real terminal runs were preserved at `a347ce3`;
the subsequent evidence-verifier correction is `339a2a6`.

The teaching made the implementation boundary clear. Chapter 2 §2.7 specified
terminal selection, ordinary text, local commands, input limits and failures.
Chapter 3 §§3.3 and 3.9 made a completed tool turn the unit returned through
the same public submission service. That combination let the human client
reuse the existing loop without introducing another conversation or HTTP
client. Chapter 3 §3.10's `/history` then `/redact` sequence gave the person
a concrete way to find and change a real result. No architectural ambiguity
blocked this integration.

The first merged test run failed because an inherited Chapter 2 fixture
repeated a tool-only response with the same call ID. Chapter 3 correctly
continued it and rejected the repeated identity. This was a fixture migration
issue, not unclear job ownership or a provider failure. The revised test
returns a final response after an ordinary tool error and checks paired
continuation, cumulative usage and final-only display. The inherited notice
that tools were unavailable and the stale Chapter 2 README also needed
updating. A short transition note could remind readers that Chapter 2's
tool-only notice and fixtures stop applying when automatic execution arrives;
the existing Chapter 3 lifecycle already supplied enough information to infer
that change.

The live runs exposed a useful distinction between a model's explanation and
its actions. In the OpenAI session, the model omitted `overwrite:true` and
then described an incomplete recovery. Inspection of the actual calls and
file bytes identified the omission; a plain-text follow-up completed the
write. Its later summary also called retained `é` an incomplete sequence,
although the tool bytes held the complete character and omitted `X`. These
were model responses, not tool defects. Keep the chapter's independent disk
checks and consider preserving one short corrective follow-up in the reader
walkthrough. The [feature ledger](../../solutions/edition-2/ch03/evidence/ch03/human-chat/FEATURES.txt)
records both observations without rewriting the original transcripts.

Review found a separate defect in the student's evidence verifier: it trusted
a temporary executable path without enforcing the recorded identities. The
repair requires an explicit binary and immutable source revision, validates
both before running, and writes reconstructed requests separately. Mismatches
leave evidence unchanged. This was an evidence-tool defect; the subsequently
strengthened skill now states the missing operational safeguard explicitly.
No production fix or paid rerun was needed for that revision.

The coordinator also migrated the unfinished merge into the canonical outer
repository while preserving it. That changed the working location, not the
chapter's runtime contract. These process interventions and the later review
findings are part of the account, not evidence of an unaided initial attempt.
The concrete suggestions above concern inherited notices/fixtures, observable
model recovery, and evidence identity. No additional blocking teaching defect
is claimed from this integration.

Resolution check, October 7, 2026: the author recorded all three dispositions
in [the feedback response](chapter-03-student-feedback.md). The new §3.3
transition explicitly replaces the obsolete tool-only notice and fixture.
Section 3.10 already preserves the exact corrective request, replacement
bytes and complete-character observation. Chapter 0 and the mandatory skill
now require immutable source and executable checks before evidence replay.
Those responses resolve this report's suggestions. This confirmation is
post-acceptance; it changes neither the original tag nor the live receipts.
