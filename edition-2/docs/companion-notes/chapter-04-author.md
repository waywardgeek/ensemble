# Chapter 04 author carryover

## Reading and scope

Initially read the complete approved set: `book/companion/voice.md`; all 1,074 lines of `book/edition-3/chapter-04.md`; accepted `book/companion/chapter-03.md`; and `docs/companion-notes/chapter-03-author.md`. Recovered clipped Chapter 03 output with bounded reads. After the editor read the full draft and note, reread the source at lines 220–487, 530–610, and 625–655 for the targeted revision. Followed no inherited links into process documents, evidence records, or implementation. Wrote only this note and `book/companion/chapter-04.md`; no Git operations.

## Manuscript and source coverage

The chapter follows the source narrator’s framebuffer screenshot hang and Bill’s process-tree recovery into observation versus lifetime, dispatch-wide admission and handles before effects, Agent-owned jobs with Ensemble-allocated handles, shutdown admission races, and the four control tools’ non-job bypass. It explains defaults/pending/call precedence, very-next-call consumption including control calls and failures, replacement setters starting from defaults, mandatory visible consumption diagnostics, patterns versus arbitrary-output wakes, terminal completion, destructive shared cursors, unseen-byte matching, finished-job waits, initial full-result compatibility versus later output wakes, and preview caps versus full retained output.

The middle covers PTY versus pipes, input echo versus processed replies, the historical Delve demonstration, shell launch returning before job completion, process-group cleanup without sandbox claims, killed state versus late finish, physical cancellation limits for blocked handlers, waiter notification, and shutdown evidence. Per-call cwd, launch workspace versus log storage, optional override serialization, and explicit persistent-shell jobs explain the deliberate absence of sticky `set_cwd`.

The shell-only/file-tool negative controls, cancellation-before-start lifecycle reasoning, historical synchronization-test defects, misleading killed/done score, dependency/setup false detections, cold-build timing assumption, and measured macOS/assumed Linux distinction are causal stories rather than grading instructions. External consumption proves a separate-module boundary using the existing loop and jobs. Correlated execution, saved output and metadata are distinguished from a plausible final answer. Logger ownership follows actual parents to Ensemble; runtime output alone cannot prove that source path.

## Editor-requested revision

Targeted rereading identified errors in the initial draft; no saved comparison establishes an intervening source change. Removed the unsupported persistent Agent/Ensemble `tool_limits` setter scopes and the incorrect claim that delay and pattern conflict. The revised passage permits a pattern within a bounded observation, explains dispatch-order consumption, consuming `kill_job` or another setter, replacement fields reverting to defaults, and loud consumption/misfire diagnostics. Added the watchdog-map/declaration story’s causal turn from a tool’s allowed blocking duration to the caller’s desired observation interval.

Added the omitted cwd distinction and no-sticky-setting rationale. Corrected an adjacent ownership error in the initial draft identified during the targeted reread: each Agent owns jobs/output while Ensemble allocates handles. Replaced the vague startup-grace account with the source’s initial-observation versus later-wait distinction. Added the explicit shell handoff: launch returns; the job/reader owns subsequent process life, output and exit; dispatch must not finish the job and close output merely because the handler returned.

Final reviewer corrections were checked against source lines 1–45, 625–670, 860–900, and 895–1074. Removed the invented localhost/Playwright opening; the source narrator requests the framebuffer screenshot, and Bill’s recovery kills the whole agent tree. Replaced the unsupported Delve quit-echo story with the documented serialized-escaping misread, measured fast returns, unnecessary repeat, and result `42` with exit code 0. Recast both shell-only perfect-score references as the negative-control counterfactual, not original chronology. Removed unsupported historical attributions for cancellation-before-start and short-timeout teardown, retaining the general lifecycle reasoning and documented waiter/finish ordering tests. The consumer capture paragraph remains: targeted rereading confirmed its log-header/scratch account, and the editor explicitly withdrew the request to change it. These corrections concern draft errors, not an established source-version change.

## Non-obvious choices and handoff

- Historical observations remain attributed to Bill/the source/edition records, not presented as new verification.
- Predicate precedence describes terminal completion, pattern replacing arbitrary-output wake, and delay returning control without cancellation; no invented simultaneous-event tie-break.
- One-shot limits remain literal next-call settings, not a target matcher or speculative policy chain. Explicit shell persistence has a handle; implicit sticky cwd is rejected.
- Additive tool registration and the historical demo’s deliberately disabled builtins serve different purposes and are distinguished.
- Caps bound presentation, not retention or implicit wait pagination. Shared cursors are not a broadcast subscription.
- Exercise knowledge is explanatory prose, with no assignments, contract tables, code dumps, or student pipeline.

Jobs survive a wait, not process exit. Handles do not make foreground human input interruptible. The ending hands prompts, interruptions, and completions to Chapter 05’s nonblocking inbound-loop problem without implementing its mailbox or turn states. External tools motivate the boundary, but no external tool server is claimed.

## Actual checks and limitations

Before the editor revision, a read-only Python check confirmed the heading, final newline, existing relative source-link target, and absence of code fences/tables. That was a structural check, not prose lint or behavioral verification. Revision used targeted source reads and exact-anchor replacements (with unique-match checks); no broad audits, implementation inspection, builds, graders, live calls, or `cmd/lintprose` run. Editor and independent-review feedback has been applied; the revised manuscript is ready for their final recheck. Available for material revisions.
