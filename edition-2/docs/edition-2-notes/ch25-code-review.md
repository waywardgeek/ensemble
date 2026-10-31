# Original25 / Edition2ch22 completion: independent code review

Reviewer `reviewer_astra`, GPT-6 Astra/high, October 11, 2026.
**Accepted with two documented exceptions on initial formal review; zero correction rounds.**
Final source `e00d5bcaf1412c50787428ef1fa08bed59b99e11`, main tree
`30fb05d154b9808eb88622aaff9c1424cd7e8af1` (170 tracked source files).
The preparation and pre-submission history below remain preserved.
Mandatory `.agents/skills/ensemble-coding/SKILL.md` applies. Current full-source
reading is recorded in [ch22 review](ch22-code-review.md) and [ch25 notes](ch25.md).
No original answer, provider or grader execution was used for this preflight.

## Missing public-input facts

Source `internal/grade/ch25_harness.go`, revision
`009d74c70b108681b2107b4c3a08685508567cee`, SHA-256
`1896697aef507f31c7955c359594501dd644f70c16fcd388f9f955a6e4bb9bc6`.
The shared builder selects `./cmd/` when present. Sessions launch without
arguments in fresh working directories, with `LLM_BASE_URL` and dummy
`OPENAI_API_KEY`/`ANTHROPIC_API_KEY`; no model or skill override is supplied.
Input is JSONL `{"kind":"prompt","text":"..."}`. A JSON `assistant` string
ends the observed turn, including an empty string. The harness waits90 seconds,
closes stdin and allows20 seconds for shutdown/save before killing the process.
Legacy terminal-envelope differences must remain distinct from fleet behavior.

The seven tools receive these public argument shapes:

| Tool | Arguments |
| --- | --- |
| `spawn_sub_agent` | `agent_id`, `task` strings |
| `join_agents` | `agent_ids` array of strings |
| `agents_status` | empty object |
| `shutdown_agent` | `agent_id` string |
| `send_message_to_parent` | `content` string |
| `respond_to_parent_message` | `agent_id`, `response` strings; `message_id` omitted for the single pending message |
| `submit` | `result` string |

These facts were sent to the coordinator for a narrow student-safe erratum,
not as answer code or permission to expose grader internals to the student.

## The two recorded wording conflicts

Original4 explicitly specifies one-shot next-call `tool_limits`, with no target
argument. Source `book/chapter-04.md`, revision
`0253737c06197d6f9b35e442916b7475c1a59a99`, SHA-256
`44e03445c850c939e331cb25c174e98df692063f824f9cedb748efa3ea1fee84`.
Original25 calls it a per-tool-name setting and says that a missing timer prevents
message wakeups, despite depicting independent timer/message select cases.

The actual25 harness calls `tool_limits` with only `ai_callback_delay:60`, then
`run_command`; no per-name target or persistent configuration is provided. It
delivers another ordinary prompt during that command, before the timer expires,
and expects a provider request with the message while the command remains
unfinished. Later actual tool output must still arrive. This supports the
existing next-call interface and an independent message wake trigger.

Recommendation to the coordinator: preserve accepted next-call semantics and
record a narrow prose correction rather than adding a per-tool policy. Timer
expiry and incoming messages are separate causes of ending a wait; neither
means killing the job. This is sourced contract preparation, not an implemented
result or advance acceptance. Current original22 review takes priority, and
original26 remains conditional at the25 review boundary.

## First-green comparative preflight, October 11

Frozen source `7175d33ebfd14c4d2807580a17320786a131a507`, main tree
`9528cdd0dbfda11a7c3b3b77bf33223a4dade259`. This is pre-submission feedback,
not a formal disposition; zero correction rounds. Read the student's stage
review, current implementation, and matching original fleet, messaging, idle,
spawn, workflow-tool code and tests only after this implementation existed.
The original grader actually scored **0/100 FAIL** in0.610s: all seven checks
stopped before inference with `provider model is not set`. No fleet behavior
ran; latent terminal-envelope differences are not observed failures. Final
exception assessment awaits the complete student review and runtime evidence.

Two independent real-runtime probes found implementation defects and were sent
through the coordinator without old-answer code:

- A child with its parent's accepted ID (`main`) generates the same first
  question ID as the root. Its question overwrites the root's pending human
  question; after answering the child, answering the root fails with `no matching
  pending question`. The same test with distinct child/root names passes.
  [Probe](../../solutions/evidence/ch25/reviewer-question-collision_test.go.txt),
  [failure and control](../../solutions/evidence/ch25/reviewer-question-collision.txt).
- A worker that exhausts its original request plus three wakes escalates and
  stops, but still appears as `running` in the parent's fleet status. The probe
  waits until the child's escalation also causes the root's bounded four
  requests and human escalation: eight actual HTTP requests, then the stale
  child status. This is current lifecycle truthfulness (§25.2/§25.4), not a
  request for Chapter26 activity states.
  [Probe](../../solutions/evidence/ch25/reviewer-idle-status_test.go.txt),
  [failure](../../solutions/evidence/ch25/reviewer-idle-status.txt).

Both probes use real Ensemble, Agents, managers, actors and persistence; only
external HTTP inference is faked. The coder owns corrections and affected
regressions. No production file was edited by this reviewer. Final source,
mutation sensitivity, live evidence, full counts and disposition remain pending.

## Corrected-green independent checks (not final acceptance)

Source `891d513bf142013d6888760cfae4fcaab1edd680`, tree
`6cbf18933f82ee450d91750442161b176e469f8e`, fixes both findings. The exact
original probes plus a new omitted-ID human/child reply-selection control pass
under the race detector (1.229s). Root and same-named child now receive distinct
recipient-owned IDs; the human response releases only the root and retains
`From: Bill`, while the child's response retains its actual parent sender.
Exhausted child status is `failed`, with the explicit three-wake reason and
empty `Result`; it is not silently promoted to submitted completion. The
corrected join returns that failure. [Check output](../../solutions/evidence/ch25/reviewer-corrected-probes.txt),
[selection probe](../../solutions/evidence/ch25/reviewer-reply-selection_test.go.txt).

Preliminary own-predecessor counts at this freeze, in order physical / active /
comment-only / blank: student production Go adds **802 / 645 / 112 / 45** and
Go tests **711 / 659 / 36 / 16**; original25 adds **1698 / 950 / 593 / 155**
production Go and **1754 / 1290 / 275 / 189** test lines over original24.
Other counted implementation languages are unchanged. The lexical method and
whole-source totals are in [counts](../../solutions/evidence/ch25/reviewer-counts.json);
tracked source only, inline code remains active, tests/support separate, net
changes count moves once. An initial inventory excluded implementation folders
named memory/data; corrected before recording these totals, whose predecessor
matches the accepted original24 review.

The reduced production size comes from reusing the existing actor, jobs,
persistence and parent interfaces rather than a separate child execution path.
The new fleet manager keeps immediate children and pending questions with their
owner; the application factory decides provider configuration. Comments explain
that ordering, recipient-owned IDs, cancellation, bounded correction and saving.
Its added-comment share is about15% of nonblank production, below the roughly20%
teaching target, but useful explanations and exported contracts are present;
this is not a padding quota. Original explanations are more extensive and
sometimes repeat the same rationale across adjacent types and methods.
The student tests retain real actors and jobs behind small external HTTP fakes,
and add actual CLI construction/human input; fewer lines alone are not acceptance.

Factory installation selects the fleet protocol for the entire Ensemble, so
shared-actor correction applies to roots and every recursively spawned child.
The executable installs the factory. Existing non-fleet embedded/GUI clients
retain their earlier protocol. Final review must explicitly assess this bounded
compatibility departure against the manuscript's universal wording; it is not
an unrecorded exception or permission for a fleet child to bypass correction.
Final student review, live provenance and mutation classification remain pending.

## Final disposition — accepted with exceptions

Read the **entire final student review** in [ch25.md](ch25.md), final source and
README, original25 implementation/tests against original24, and the actual
original grading, final checks, live artifacts and all mutation outputs. Final
source `e00d5bc` differs from independently checked/live `891d513` only by README
prose and two correlation-test diagnostics; production bytes are identical.
The affected final correlation test independently passes (0.018s). Repeating the
paid run or broad unchanged checks would not strengthen this delta's evidence.

I accept these two narrowly bounded explanations under Bill's review exception
rule:

1. **Missing-model grader fixture.** Actual original25 result remains **0/100
   FAIL**, make2/grader1,0.610s on exact first-green `7175d33`. The unchanged
   harness provides a fake endpoint and dummy keys but no model; every scenario
   exits before inference. Keeping explicit model selection is consistent with
   the accepted application and avoids inventing a remembered provider default.
   This waives that launch fixture, not the seven behavioral obligations. No
   terminal-envelope or other unobserved failure is silently waived.
2. **Factory-selected fleet protocol.** The chapter's universal “any Agent”
   wording is narrowed for existing applications that never enable fleet mode.
   Installing the factory selects the protocol for the entire Ensemble; the
   shared actor enforces correction for root and recursive descendants, and the
   actual CLI installs both factory and human sink. No-factory clients hide
   fleet tools at declaration and dispatch and preserve their earlier completion
   behavior. This is a small compatibility boundary for existing non-fleet
   consumers, including the optional GUI, not application-managed correction
   or a bypass for fleet children. README and student review disclose it.
   Future fleet-enabled clients must install this composition explicitly.

The valid masked obligations are covered by real runtime checks: recursive
three-generation construction with both actual parents; isolated child files,
restore and cascading shutdown; current child status/removal; all-selected fan-in;
question yield and addressed replies; incoming-message wake while the same real
job continues and later delivers bytes; and literal three-wake correction,
escalation, active-child rejection, successful-submit reset and no failed-turn
retry. Built-CLI checks exercise human reply during a blocked turn and root
escalation. Core tests/vet, GUI tests/vet and focused fleet race results actually
pass; no empty test selection is credited. Earlier sandbox/accounting behavior
reuses accepted coverage. The two independent pre-submission regressions above
now pass, including same-name routing and truthful failed-not-submitted status.
There are no remaining blocking implementation findings.

### Mutation assessment

Read all 21 attempt records and raw outputs: **19 intended assertion failures,
one initial eight-second liveness failure, and one initial collision diagnostic
requiring inspection**. The initial generic early-join timeout is causally a
failed liveness check with a working baseline, not the originally predicted
output assertion. Its narrower parked-parent test detects missing question yield
at 500ms without generic wait wakeup masking it. The collision's initial context
failure is preserved; the final diagnostic reports one retained question after
one child request, and the restored-collision defect fails that check. Both
replacement runs are included in 19, not additional independent proofs. The
factory bypass actually fails depth/ownership rather than its initially expected
broad label; that revised interpretation is supported by the raw output.
The missing-human-sink mutant reaches its intended six-second UI liveness checks
and bounded child-process cancellation, not a package-wide test timeout.
No compile/setup failures or unresolved survivors are counted as kills.

### Live result and limits

Independently verified the exact live binary SHA, all 20 actual Astra/high request
records,20 Astra completions and matching created/completed response-ID sets.
All 16 non-submit tool-return payloads match actual provider inputs by call ID.
Root and three child logs show real question/reply, file creation, command tests,
explicit submission and collected results. Actual artifact bytes match saved
copies; independent Python assertions pass. Join 12 starts 55.751, receives stdin
steering 55.885, returns running 55.887, then the same handle returns all three
submissions at18:31:10.340. The full steering text occurs in actual provider
input 15. The CLI exits 0 after 48.198s.

The final submit omits `STEER_COBALT`. I accept the student's explicit limit:
transport delivery and job continuity worked, while the model did not fully obey
that instruction. This is not evidence of perfect steering compliance. Preserve
also the add child's malformed shell command (exit1) and self-correction (exit0),
which is ordinary tool-error recovery, not automatic failed-turn retry. Shared
wire records do not justify ordinal request/response pairing. Raw wire bytes are
preserved in `live-api.txt`; [reviewer receipt](../../solutions/evidence/ch25/reviewer-live-check.json)
records its verified hash and these limits. No fresh provider call, credential
access, billing-charge claim or endpoint interception was performed by review.

### Comparative conclusion and release

The final counted delta remains the 802/645/112/45 production-Go and
711/659/36/16 test-Go figures above. README adds 22 prose lines, kept outside
implementation counts. The original 950 active-line delta includes three unrelated
model-catalog additions; excluding those still leaves 947, versus student's 645.
No source moves conceal growth. Exported additions have useful contracts and
ownership/lifetime explanations; the useful comments are thinner than the
original but not replaced by compressed syntax. Tests protect real earlier
machinery and actual CLI wiring without mock substitution or a new framework.
Required capabilities, not line count alone, support the simpler implementation.
No optional fleet tools, GUI fleet controls or speculative registries were added.

**Accept original25, completing combined original24+25 / Edition2 Chapter22.**
Zero formal correction rounds used; the two real defects were corrected with
attributed pre-submission assistance. Coordinator may now export the exact
accepted main tree and perform Bill's conditional original26 readiness check.
This acceptance neither authors Edition3 nor preaccepts original26.
Mandatory `.agents/skills/ensemble-coding/SKILL.md` governs the next handoff.
