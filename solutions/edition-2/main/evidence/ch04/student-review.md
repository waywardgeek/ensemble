# Chapter 4 student teaching review

## Initial implementation experience (2026-10-07, before comparative feedback)

Source: accepted Chapter 3 commit `1a61e1f2487cdc94ee65bd0e1593065cc1e95f45`; teaching hashes are in `teaching-source-versions.json`. Only the new teaching, required skill, and preceding main source/tests were read. No first-edition or grader implementation was read. This account is being recorded during implementation; live use remains pending.

The explicit ownership in §4.1, commit-time response identity in §4.3, and separation of wait timing from cancellation in §§4.4–4.7 were useful. They directly determined the new Jobs parent chain and independent event-append serialization. The contract has not required architectural clarification so far.

§4.6 explicitly changes run_command from two captured streams to a merged PTY. Existing Chapter 3 command unit tests asserted the incompatible old representation. I removed only the pipe-specific assertions from the tools tests and am replacing their exit-status, drain and UTF-8 guarantees with managed-job checks. Existing fixed event-index assertions also needed event-kind lookup after job_ended was introduced. These are expected implementation/test adaptations, not defects in the chapter. The coordinator confirmed that interpretation. The old frozen sources remain untouched.

First compile exposed the removed capture type still referenced by a UTF-8 test; first integration run exposed the shifted event index. These were my implementation/test migration errors. The inherited main suite now passes.

First black-box Chapter 4 grade: 90/100. All job-specific checks, including actual grader-driven Delve, passed; ch3parity reports inherited ch2parity failure without detail. This is not yet classified as implementation or fixture fault. I am running the lower-level black-box grade to identify the observable discrepancy. No live-model or human-interface proof is claimed by that grader result.

Concrete suggestion at this point: mention that predecessor tests using absolute event-array indexes should select by event kind/call identity once asynchronous terminal events can interleave. This preserves their original behavioral intent. No chapter ambiguity currently blocks code.

## Difficulty surfaced during implementation: retained artifacts across launches

§4.2 says application handles start at one and output is created exclusively, while §4.5 preserves artifacts. I interpreted an existing `cr/io/1` in a restarted application's workspace as a failed allocation which refuses execution. A new public-library regression reproduces that behavior: the first application reads a file, the second shares the workspace and refuses its first read while preserving the artifact. The inherited grader runs provider cases in one workspace; the same refusal explains its OpenAI/Gemini missing continuation and usage. This is a teaching/contract usability gap as well as a fixture interaction, not grounds to truncate artifacts or weaken tests.

The coordinator confirmed the shared fixture and asked the author to specify skipping occupied exclusive-create handle candidates while retaining other allocation failures. Affected allocation changes are waiting for published wording. Other lifecycle and race checks continue.

## Author clarification applied

Read the published §§4.1–4.2 clarification: consumed candidates skip occupied exclusive-create leaves (including directories and symlinks), other errors still refuse execution, and only empty workspaces promise consecutive successful handles. This resolves the restart ambiguity while preserving artifacts and Agent permissions. Allocation now follows that wording; the public regression retains the original preservation check and adds successful restart at handle 3 without importing prior job state. The new §4.3 guidance also addresses the event-index suggestion. No old implementation was consulted.
