# Chapter 9 independent live-evidence review

Status: preliminary closed-run audit; final student matrix and initial-attempt
freeze pending. This is not chapter acceptance or historical comparison.

## Reviewer boundary

The reviewer previously implemented Chapter 8, prepared the initial Chapter 9
checker before handing it off, prepared Chapter 11 checkers, and reviewed
Chapter 13 prose. The reviewer did not implement the Chapter 9 runtime. For
this audit the reviewer read the complete mandatory coding skill, architecture,
current Chapter 9 contract and student live plan; relevant student-review
entries; the independent public-attempt and stream-repair reports; closed
launches, terminal/event/browser originals and bindings; and the student
recovery/verification/keyboard support identified below. No first-edition
implementation or comparative review was opened. No credentials, provider
calls, runtime edits or builds were needed.

## Closed originals and source identities

`checkpoint-evidence/ch09-live-audit-closed-inventory.json` records 16 closed
launches and 499 original-file hashes, all matching. Fresh discovery identities
and selected model IDs match each launch. The initial inventory had a stale
hardcoded label saying Gemini G was unclosed even though its scan included the
closed run; `ch09-live-audit-closed-inventory-initial.json` preserves that
reviewer metadata error. No student receipt was rewritten.

| Provider | Recorded model | N | F | G | P | Total |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Anthropic | `claude-sonnet-4-6` | 14 | 5 | 9 | 4 | 32 |
| OpenAI | requested `gpt-4.1`; accepted responses `gpt-4.1-2025-04-14` | 11 | 5 | 8 | 5 | 29 |
| Gemini | requested `models/gemini-3.8-flash`; resolved/reported `gemini-3.8-flash` | 16 | 5 | 7 | 4 | 32 |

These are 93 generation HTTP attempts against the reviewed 108 ceiling, with
each provider below 36. Three separately recorded discovery requests bring
HTTP including discovery to 96. Failures count against the budget.

Original Anthropic N/F/G/P use `c0e3171`. The actual empty-argument correction
and subsequent launches use `75a7255`. OpenAI P recovery uses the separately
frozen support revision `3da764b`. These are distinct identities, not one
retrospectively relabeled run. Independent reports at `7dc9cc1` and `7da5052`
already establish the partial Anthropic public attempt and old/new complete
source/executable bindings plus 14 stream controls; this audit retains their
scope rather than rerunning those established checks.

The recovery binding independently passes its complete 134-source,
nine-executable, six-support, six-catalog and 74-browser-dependency preflight.
Its runtime source map and all executable identities equal the stream-repair
binding. Review of the copied support finds only its new location/root,
OpenAI-specific F5/P5 limits, and safe transport diagnostics changed; the other
four original support files remain byte-identical. The originals remain at
their earlier paths. Diagnostics select exception/reason type names and an
integer-or-null errno, without exception text, URLs, headers or keys.

Independent rerun of `support-recovery/test-recovery.py` reproduces the frozen
local controls: exactly four remaining OpenAI P forwards, then refusal;
exhausted provider/P refusal; F5 refusal; unchanged Gemini P4; and safe mocked
transport classification. No real origin or credential is used. The 25
identity-control receipt entries were reviewed with their passing-parent
fixture implementation, not independently rerun here. See
`checkpoint-evidence/ch09-live-audit-support-browser.json`.

## Feature evidence and limits

| Feature | Closed evidence and disposition |
| --- | --- |
| Human CLI and `/skills` | Actual `script` PTYs retain ordinary prompts, follow-ups, local skill inspection, responses and exit. N/F exist on all three providers; these are distinct from the public consumer's structured output. |
| Load, dependency sharing, dynamic offers and unload | All three N/G sequences load edit, review and search, then unload edit. Revision 4 retains read and explicit review/search roots while removing write_file. Actual scratch writes have retained exact output files. |
| Revoked writing | Anthropic/OpenAI answer that writing is unavailable and leave the created file unchanged. Gemini removes write_file from state and the next offered schema, but its follow-up request fails in the relay before an accepted answer. That failure is neither a verbal refusal nor a runtime forced-call rejection. Forced admission denial belongs to deterministic controls. |
| No-op and unavailable name | Anthropic/OpenAI N make the actual unchanged search and absent calls. Gemini's final management run makes load review, unchanged review, then absent, with revision 1 preserved after the no-op/refusal. Earlier failed Gemini attempts remain separate. |
| Full primary and next-call limits | F on all providers starts with the full primary tool set, applies max_output_bytes=1, consumes it on inactive unload, then reads the complete notes and writes a summary. The intact management acknowledgement and subsequent full read establish the intended consumption behavior. |
| Public two-Agent configuration | All providers have real model requests from two Agents with distinct ceilings, material/bindings and notes. Beta's edit load is refused; inspect succeeds; Alpha can load edit. Contributors and owned material appear in public inspections. Anthropic P remains partial as described below. OpenAI recovery completes four requests under the revised budget. |
| Retired material and replay | Retired edit cards survive unload and reconnect. The separately bound public result-redaction demonstration and request reconstruction are derived evidence, not new model work. Final all-attempt replay reconciliation is pending. |
| Browser surfaces | All three G originals show two browser views of the same Agent, management cards in Actions, skill material in Chat, current sidebar state and reconnect. Original expansion actions are mouse clicks. A separate real Enter action on the running Gemini GUI is recorded; its pre-close launch identity still needs reconciliation below. |
| Manual and automatic speech | All three browser originals record manual submission of the full 2,455-character retained body. Automatic starts use ordinary answer/tool-summary keys, with no automatic manual-skill start; reconnect adds no speech start. Real Chrome-only WAVs retain sound energy after a quiet control. They do not prove human hearing, intelligibility or complete audible delivery of every character. |

The keyboard screenshot was independently viewed. It shows the expanded
retained text in Chat beside revision 4, review/search roots, read retained as
a dependency, and no write_file. The raw keyboard receipt records focus plus
Enter and only subscribe/pause frames, with zero prompts. Its helper, binding
and screenshot hashes match; its `launch_sha256` does not match the finalized
launch because the action was taken before closure. Preserve the original
receipt and reconcile that earlier launch snapshot before final linkage is
claimed. No additional paid run is indicated.

The audio audit independently reads float32 stereo 48 kHz WAV samples. All four
captures have zero RMS in the first 1.5 seconds and nonzero RMS after 2.5 seconds
(approximately 0.088–0.091). The first Anthropic capture's quiet span is shorter
than the nominal two seconds; the second controlled capture remains distinct.
The reviewer did not listen to or transcribe these recordings. Full-text
submission, native callback timing and captured sound are separate facts.

## Failures retained

Anthropic P exits 1 after Beta's extra management calls exhaust its two-request
turn limit. Its notes read and required Skills features were independently
established at `7dc9cc1`; a final Beta answer and consumer success were not.
Do not turn that partial task into a successful scenario.

The original Anthropic G empty-object stream is a real runtime failure. Its
explicit-path recovery succeeded on the old binary; that does not prove the
repair. The later three-request N supplement on `75a7255` actually accepts the
complete start `{}` plus zero-byte delta and executes `list_directory {}`.

OpenAI P request 001 and Gemini N request 008 encounter local relay HTTP 502,
with no upstream response retained. The old relay omitted exception details;
the cause remains unknown, and the approximately 20-second OpenAI failure is
not evidence of its configured 60-second timeout. OpenAI's later four-request
recovery completes without erasing that first spent request.

Gemini N recovery receives a final empty-text STOP with no required
`candidatesTokenCount` and refuses it. The subsequent four-request management
run supplies the missing no-op/unavailable demonstration; the malformed-usage
attempt stays exit 1. No inferred-zero runtime change is justified here.

OpenAI G's browser wait times out looking for a local composer status for a
terminal-originated turn. The terminal and actual state/cards show success;
the wrong selector and timeout remain in the browser original, and no repeat
prompt is credited.

## Remaining closure work

Receive the frozen final student matrix and failed-attempt replay report;
reconcile them against these closed originals and each actual source identity.
Resolve the keyboard receipt's pre-close launch hash without rewriting it.
Carry forward the Gemini revoked-write and native-audio limits explicitly.
Historical comparison remains prohibited until root declares the student's
initial source/live/teaching freeze complete. No complete acceptance is claimed
by this preliminary review.
