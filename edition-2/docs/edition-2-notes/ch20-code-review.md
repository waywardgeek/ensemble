# Original20 / Edition2 Chapter19 — independent review

## Initial formal disposition: accepted, October 11

Reviewer `reviewer_astra`, GPT-6 Astra/high, read the complete governing documents,
`.agents/skills/ensemble-coding/SKILL.md`, original20, preparation and final
student review. **Accepted at initial review, zero correction rounds.** Final
submission is `bade50cef5cdc53490d4263fa57f5e37abbb93ec`, main tree
`bd2998e68f976dbaf437b642b131d350e086242f`. Actual live source was `fab0315`,
tree `59010a4c2b333c2f74777b6d76079e8f2b3af90d`; only README changed afterward.
The reviewer requested that narrow documentation clarification before submission;
it is attributed assistance, not a formal correction round or runtime change.

Original20 is intentionally prose-only. There is **no original grader or score**,
and no new exception is needed. I accept the student's scope explanation: Bill's
specific crossover goal requires real coding with the approved model, not a
wholesale implementation of the narrative's historical repairs. Prior19's
accepted grader/cache exceptions and actual failed original score remain history.
No implementation or scope finding remains open.

## Source, comparison and teaching quality

The entire product change moves Astra into the existing Responses feature row,
extends the existing model-boundary check, and documents the actual settings.
It introduces no configuration path, runtime abstraction, exported symbol or
mutable global. The actual CLI constructs the public Ensemble, configures auth,
creates its Agent, and uses the same actor/tool/job machinery as before. Model
features still belong in Engine behavior; Agent owns reasoning configuration.
The existing Responses renderer consumes that configuration, and the existing
credential path retains explicit subscription billing and cache compatibility.

[Independent counts](../../solutions/evidence/ch20/reviewer-check.json)
compare the final source against its own accepted predecessor `534112a`. Physical
lines include blanks; active lines exclude blank/comment-only lines, retaining
inline comments as active. No files moved. Production Go changes by **+2 physical,
0 active, +2 comment-only** lines; tests by **+6 physical, +6 active, 0 comments**.
The changed production file is 109 physical/95 active/10 comment-only lines;
the test file is 275/249/19. README is separate prose; the Python scratch task and
launcher are evidence, not production or a new framework. Other languages,
examples and production files are unchanged.

There is no first-edition `solutions/ch20` snapshot or chapter20 implementation
delta to measure. I inspected the relevant first-edition model/transport context
in `solutions/ch19` and the next available `solutions/ch21`: Astra similarly
changes from its default Chat surface to explicit Responses within one existing
model row, with no net active-line growth for that row. The later table's shared
explanations cover model-specific routing and historical live measurements.
Its complete chapter21 growth includes a different exercise and is **not** a
chapter20 comparison. Neither its additional features nor historical model
prices become requirements for this submission.

The student's two new WHY lines explain the tool-calling transport choice and
reuse of Agent settings; useful earlier comments remain intact. No new exported
documentation is missing. The changed file's comment share is about9.5% of
nonblank lines, below the rough20% teaching target, but this tiny row change is
clearly explained beside its cause. Padding the table would not improve teaching.
The size reduction relative to a new transport implementation comes from reuse,
not compressed syntax or deleted explanations. KISS is satisfied.

## Tests and actual crossover evidence

Read the student's core/GUI test and vet results and both targeted mutation
outputs. Restoring Astra's Chat selection fails the intended transport assertion;
replacing requested high with low fails the actual rendered-effort assertion.
Both fail behaviorally, without compilation or setup failures. The correct
baseline passes. My isolated exact-source model-boundary check also passes with
race detection (1.027s). No mock replaces the runtime; the check constructs the
real root/Agent and renders its request. Existing auth/streaming/runtime evidence
remains inherited, not a new whole-book test claim.

The [independent receipt](../../solutions/evidence/ch20/reviewer-check.json)
verifies five captured request/completed-response pairs naming `gpt-6-astra`,
high effort and detailed summaries. Every request streams with `store:false`,
omits paired explicit plan caching, and every terminal response is completed.
The five actual calls are two `read_file`, one failing `run_command`, one
`edit_file`, and the passing `run_command`. Their call IDs match the real outputs
in subsequent requests; the final answer accurately reports the task outcome.
The existing human CLI, runtime and tools performed the work, not the launcher.

I reran the preserved task artifacts in isolated directories: before gives
**3 failures and 3 errors**, after **6/6 passing**. The test file has identical
SHA-256 `d0a5eaab886c93918f853cde4e19819f0ab2fc8b03bf8aae504f5c8152081ea9`.
The small actual repair uses standard CSV parsing with newline preservation and
Decimal arithmetic; unchanged tests exercise real parsing, grouping, quoting,
blank rows and the actual sorted CLI. They are fixed task acceptance examples,
not a general claim of exhaustive CSV correctness or a new grading framework.
The failing baseline supplies a real defect control independently of the model's
description. No additional paid run was needed.

The launch record reports27.46seconds, CLI exit0 and no timeout. Retained save
and final event history support normal completion after EOF, unlike original19's
earlier killed launcher. Explicit `chatgpt` selection, absent configured API key,
the runtime billing announcement and inspected selection source support the
recorded funding route. I did not inspect the profile or make provider calls.
The endpoint is derived from explicit base URL plus exact transport source;
payload capture does not independently record HTTP URL or authentication headers.

There are161 ordinary text deltas and zero public summary deltas. All five
responses report zero cache grants and reasoning-token counts. High is both
requested and echoed; these counters do not measure hidden model behavior.
Progress arrived before the edit. This run establishes a bounded real coding
crossover; it does not establish interactive steering, GUI use, general project
reliability, cache savings, or that Edition3 has been authored. I accept the
student's explicit limits and readiness/authorship distinction.

Coordinator may export exact direct Edition2 `ch19` and checkpoint, then release
original21 / Edition2 ch20. Stop before original22 and24 as required; authorship
remains stopped. Preserve CodeRhapsody's concurrent work and existing snapshots
and tags. Reviewer writes only this record and `reviewer-check.json`; isolated
scratch checks are removed. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.
