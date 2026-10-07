# Chapter 4: author response to student feedback

Recorded October 7, 2026 during the initial student implementation, before
first-edition comparative feedback. The coder's retained account lives in
[the frozen student review](../../solutions/edition-2/ch04/evidence/ch04/student-review.md). It identifies
predecessor `1a61e1f2487cdc94ee65bd0e1593065cc1e95f45` and the exact teaching
hash ledger. No initial experience is rewritten by this response.

| Finding | Classification and disposition | Resolution check |
|---|---|---|
| Chapter 3 assertions assumed separate stdout/stderr, and an old capture type remained in a UTF-8 test. | Expected contract transition plus a student migration error. §4.6 already teaches merged PTY output. Retain exit-status, output-drain and valid UTF-8 guarantees in new managed-job checks; no relaxation is requested. | Student reports inherited main suite passing after adaptation; the subsequent independent revision review accepts the retained guarantees. |
| Fixed event-array indexes fail after job terminal events interleave. | Accepted teaching suggestion. §4.3 now explicitly tells readers to select event kind and call identity while preserving their original assertions. | Coder confirmed the guidance explains the migration on October 7. |
| Initial inherited grade was 90/100 with job checks passing and inherited parity failing. A fresh application found occupied artifact names in the reused workspace. | Genuine usability gap clarified before the affected fix. The coordinator independently confirmed workspace reuse. §4.1, TLDR5 and §4.2 now consume allocation candidates monotonically, skip exclusive-create collisions, preserve occupied files/directories/symlinks, and refuse other creation failures. Empty workspaces still start at 1. Existing artifacts survive CLI restart; old jobs are not resumed. No grader weakening is requested. | Coder confirmed resolution and appended its original review. Coordinator reports revised inherited grade 100/100, all four modules' vet/tests and main race checks passing. Original 90 retained. Independent collision/fault checks, comparison and live gates remain pending. |

## Initial actual-use feedback

The student froze its initial experience at `9803b006aa2ad14337655595c057b3016d11363e`
after the real all-three-API human sessions and before comparative feedback.
The author read the entire report and raw terminal records.

- **Model accounts can misdescribe correct tool behavior.** Accepted. §4.9
  now checks the 33-byte slow-command artifact against its exit-status line,
  and the zero-delay read against running/done events and all42 final bytes.
  The explanation teaches status and artifact inspection, without demanding
  that implementation compensate for inaccurate model narration.
- **Exact input needed a human correction.** Retained as actual behavior in
  §4.9: Chat Completions first sent `beta`, then a fresh job received `beta\n`
  with append_newline:false and returned REPLY:beta. Both attempts survive;
  no new input-API defect is claimed.
- **The launcher missed Delve on its terminal PATH.** Retained as an environment
  prerequisite failure before any model call. The chapter already requires
  installed dlv on PATH; the subsequent real model-driven debugger is separately
  shown and linked.

The coder confirmed these prose dispositions on October 7. Supplemental
real-provider public-consumer coverage and all five modules' final commands
are separately recorded in `supplemental-public-receipts.json` and
`supplemental-checks.json`, committed at `8a4794d`; they do not change the
first four-module check record. The author inspected those receipts and each
supplemental Agent log/artifact. Comparative review has since requested two
implementation architecture corrections under already-taught rules; no
weakening of the contract is proposed, and revised acceptance remains open.

The student's report found ownership, append-time response identity and
wait-versus-cancel teaching useful. Preserve those mechanisms. This feedback
record was retained as an interim exchange before final revised acceptance.
The completion reconciliation below records the later result separately.

## Comparative clarification and boundary checks

- **Source truncation leaked presentation text into read_file artifacts.**
  The reviewer reproduced the issue and the student raised the boundary.
  §4.5 now gives exact bytes: source ABCDEFGHIJ with max_bytes4 stores ABCD,
  counts four produced bytes, and retains the source-cap notice separately
  for initial or later reports. Report truncation and source truncation have
  different recovery promises. The reviewer confirmed the teaching and added
  ordinary, UTF-8, line-range and delayed-read controls; they fail on the
  initial implementation as expected. Corrected runtime d25 passes these
  independent controls, and all three revised human runs preserve exact source
  bytes. The student confirms that the published distinction resolved it.
- **Artifact creation failure was not promised as a recoverable tool error.**
  §4.2 explicitly permits a terminal infrastructure error before admission,
  without execution, invented job or fabricated paired result. Preserve the
  accepted history, stop continuation, and close the Agent. This clarifies
  the published refusal rule rather than imposing the checker's unpublished
  recovery assumption. Revised independent checks pass and the student confirms
  the terminal-error/history policy is clear.
- **Malformed public append input could panic before validation.** The
  coordinator reproduced missing response and invalid missing-call index
  cases. §4.3 now explicitly requires structural validation before transient
  ID finalization, unchanged history on refusal, and lock release on errors.
  This is a safe-boundary implementation defect made explicit in the teaching.
  Independent public and typed-admission controls now pass on d25, and the
  student confirms the clarification resolves the boundary.
- **Changing reported log path did not change the actual writer.** The
  coordinator reproduced this configuration inconsistency. Chapter 2 §2.3
  now distinguishes lifetime-immutable log identity from mutable model
  configuration and requires atomic refusal of a changed path. No live log
  migration is introduced. The coder was notified before the affected repair;
  independent atomic-refusal and actual-writer controls pass on d25, and the
  student confirms resolution.

## Completion reconciliation

The final student review retains the original account and confirms all
published clarifications. The independent reviewer accepts runtime
`d25d3fd4e552cd17c75bf814c9899903878cfbd5` and its revised all-three-API human
and public-consumer demonstrations frozen at `9341117`. §4.9 now describes
those runs separately from the initial source, including the coder's erroneous
direct callback argument to read_file, the correct refusals and observed
follow-up. This is a prompt error, not grounds to broaden the tool schema.

The reviewer suggested explaining why apparently local file reads need jobs
and spelling out the supervision wire/state boundary. Both are accepted in
§4.1: unavailable mounted storage can block a read; tools decodes and presents
tool-facing metadata while typed Jobs operations own lifecycle and settings.
These explain existing obligations without prescribing new method names.

The student's evidence-tool errors are preserved too: late launch checks and
an empty historical source map. §4.9 links their independent review and records
the repaired preflight rule. Evidence-only `e1c6488` passes the full 51-file
historical binding check and seven independent verifier controls, with no
runtime change or paid rerun. The original bad manifest is not relabeled.
The independent reviewer accepted the final manuscript and these dispositions.
Validated checkpoint `edition-2-ch04-r1` binds the dedicated commit and
immutable tag. The linked review is retained in the matching frozen
`ch04/` export; Bill's editorial approval remains separate from the accepted
code/live/evidence and proofreading results.
