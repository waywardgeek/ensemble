# Original Chapter19 / Edition2 Chapter18 — independent review

## October11 public-contract preparation only

Reviewer `enhancement_reviewer_resumed` read all666 lines of original
`book/chapter-19.md`, revision `e1605916114230570b348ab5961c193fa1cebf80`,
SHA-256 `41fe6b803bd837c0ec3f4ed818cf02187c588c63b1ebee4713484d9fa404161c`,
and the complete [coordinator preparation and official contracts](ch19.md).
Mandatory `.agents/skills/ensemble-coding/SKILL.md` was reread in full; current
carryover/workflow apply. No implementation, old answer, grader, private auth
profile or current main source was inspected. No provider request was made.
Original18 remains active and its formal review takes priority.

The chapter's scope is credential separation, actual PKCE/OIDC identity and scope
validation, serialized rotating-token refresh and persistence, model/route-aware
Responses transport, separate incremental reasoning summaries, and visible billing
policy. Existing owner interfaces and the optional GUI boundary still apply.
The paragraph saying a renderer calls GetBearerToken conflicts with the surrounding
no-renderer-auth seam; the governing ownership and credential-isolation rules settle
this without a new architecture question. Credentials must stay out of requests'
model-visible content, snapshots and diagnostics. Passing login is not inference
permission, and HTTP200 is not a completed successful Responses turn.

The coordinator's sourced current contract resolves the short versus full quota
error-code discrepancy. HTTP status alone cannot authorize credential destruction
or metered fallback. The explicit “Yours” permission to defer chatgpt_then_api
remains meaningful despite the general three-mode description and billing rubric.
If the student implements fallback, explicit enablement and announcement before
the first paid request are binding; deferral is not permission for silent fallback.
Multiple-account schema support does not require a multi-account UI. Provider-managed
compaction mentioned in the narrative is not part of this exercise.

The remaining consequential route question is already documented: Chapter19's
unconditional explicit caching and later reported plan restrictions cannot both
be treated as universal facts. Use actual supported-route evidence when reached;
do not silently discard required behavior or call fake cache reads a live grant.
The same applies to model eligibility and streamed summaries: actual nontrivial
reasoning/tool use supplies evidence; one trivial response without summaries does
not settle capability. Current official notes are reused, not independently
re-researched or claimed as an authenticated account result here.

The public source shows login/status/ask examples but lacks the complete grader
invocation, fake authorization/discovery/JWKS/Responses endpoint overrides, isolated
profile selection and observation envelope for summary/billing events. Coordinator
may recover these missing interface facts when required, with exact source and
supplied contract recorded. Do not disclose old implementation choices or infer
new functionality from a private harness. Suggested helper/type names remain
adaptable to the existing codebase.

Live sign-in may require Bill's direct participation when19 arrives. Existing
permission to use selected API keys does not authorize copying or rotating another
running application's OAuth profile. A faithful local OAuth/JWKS fake is valid
protocol evidence but cannot be labeled live authentication. This preparation
does not initiate sign-in, waive live use, or stop unrelated Chapter18 work.

No formal review round or acceptance begins here. Wait for the authorized chapter,
frozen submission and complete student review before comparative assessment.
Original22/24 stops, stopped authorship and eventual direct Edition2 ch18 export
remain unchanged. Required handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## October11 narrow grader-contract preflight (implementation still underway)

Independently read the coordinator's newly supplied external contract, unchanged
checks/runner and relevant launch/fake-provider sections. No old answer or
unfinished student source was inspected; no grader, provider or auth flow ran.
Everything below is a **latent source observation**, not a student failure,
score, acceptance or formal correction round.

The external invocation described in ch19.md matches the harness: login uses
`auth login --print-url`, isolated HOME/EN_HOME and issuer override; the same GUI
binary later receives a private absolute `--save`, provider configuration, JSONL
stdin and WebSocket summary/text/turn observations. Both scenarios have an API
key but only logged-in ones require the minted OAuth bearer. This tests default
credential selection, not permission to override an explicitly selected billing
mode or silently fall back. A course model is an input fixture, not a live model.
The skill fixture indeed omits `type`; established original10 validation can mask
inference startup. Login might execute earlier, so do not predict every scenario
or score from this omission before the actual run.

A consequential historical conflict is now explicit in the unchanged runner:
`ch19CheckCacheBreakpoints` rejects **any** plan-route cache options/markers;
`ch19CheckCacheBreakpointsMetered` requires explicit mode and an actual developer
content breakpoint on the API-key route. That disagrees with original19's
unconditional explicit-cache wording and matches the later historical route
restriction already noted in preparation. The harness comment attributes it to
its own live observation; this is not current account-specific evidence. No
implementation policy or exception follows merely from reading the fake.
Retain the actual supported-route evidence requirement; do not remove metered
caching globally to pass a plan restriction.

Scope precision correction: fake OAuth token issuance at lines953–972 returns
openid/profile/email/offline_access plus chatgpt.tokens.use.direct, without
resource.invoke. I initially described this too strongly as a possible required
login rejection. Coordinator re-opened the current official sign-in page: its
request table includes resource.invoke, but the explicit granted-scope validation
checks chatgpt.tokens.use.direct. Therefore the fake's omission is only a scope-
shape observation, **not a confirmed defect or waiver**. Request and granted-scope
requirements must not be conflated. The coordinator records the independently
verified official wording in student-safe ch19.md; this review did not repeat
provider research or access credentials.

For later review, the seven labels overstate several predicates:

- Credential15 checks minted bearer versus API key and some secret exposure;
  it cannot prove common-interface ownership, all durable/log surfaces or that
  rendering remains independent of credential acquisition.
- OAuth15 checks the positive PKCE/issued-client exchange and a bad-signing-key
  scenario. The latter passes whenever no inference occurs, including unrelated
  login/startup failure. It needs a successful positive control; state/nonce/
  audience/expiry/granted-scope rejection and sign-out need their own valid
  current-scope evidence, not inference from that score.
- Refresh15 uses a90-second token (inside the source's five-minute window) and
  checks refresh occurrence, no observed token reuse and two turn completions.
  It does not prove concurrent serialization, atomic persisted replacement,
  terminal-error behavior or preservation on transient failure.
- Responses15 checks actual route/request fields, requested encrypted reasoning,
  successful turn flags and a correlated call/output pair. Merely requesting
  encrypted output does not prove replaying its received bytes; an error tool
  output can satisfy the pair predicate. Preserve actual parser/tool/replay proof.
- Summary15 counts separate kinds and ordering before the first text delta.
  A buffered batch emitted at completion can preserve that ordering; this alone
  does not prove incremental delivery while inference is still active.
- Billing15's two rejection scenarios use `usage_limit_reached`, not the prepared
  current full quota code. More seriously, a launch fatal is accepted as a stop,
  so zero actual inference can pass. It does not exercise permitted announced
  fallback, persisted billing policy or credential retention. Deferring optional
  fallback remains allowed by the public exercise; if implemented, the existing
  consent/announcement requirements still bind.

These are masked-obligation risks for the later source/test/live review, not
instructions to expand scope or rewrite the grader. Current public security,
transport and billing behaviors remain required even if a fixture prevents them
from running. No extra retry policy, profile feature or test framework is implied.

Exact inspected file provenance (file revision; SHA-256):

- ch19_checks.go: `3bfb49407e060eea45e8665fa1c969818195eb31`;
  `5be31eb750e63ccf2aeab13e8948099b8fb60a10d37a9b2684494801982b80b3`.
- ch19_run.go: `d68fef808f0d6d50daf682afec976cb401e1b586`;
  `f0b667d1669a9002e606caf0204d8d68b1f11d7e0701ba80139c1e376466fdbd`.
- ch19_harness.go: `70e9e3f4d72848c21143c1a3874a165debff32f7`;
  `d5acd0fdb3db968a7242fb96d0a45afbbb96c8756dde8c2a1c3a181674cff37e`.
- ch19_fakeoauth.go: `6012fc717f0cf33785ef5093b1373f17ba0a1e3e`;
  `26291e3641daae5fcf196e981c2738b7143bb5b9879af913f5683b24c183d218`.
- ch19_fakeresponses.go: `d68fef808f0d6d50daf682afec976cb401e1b586`;
  `0de734804694fc31c336dfea663f0c9afe323109476ded6784fc5a2cb130c5e2`.

Keep grader-internal coverage analysis in this reviewer-only record. Only narrow
public-contract corrections go through the coordinator to the student. Mandatory
handoff `.agents/skills/ensemble-coding/SKILL.md`; preserve Edition1 work and the
original22/24 stops. Await frozen submission for formal comparative review.

## Original grader observation — before final submission

October11 independent evidence check. The unchanged original command completed
on `06cbf20000cf3592ac2966a1041ece12b33dbc04`, main tree
`c550fa3c13faec59268cc8d7486a6ed8f590d68f`: **15/100 FAIL**, grader exit1,
make exit2. Read [raw output](../../solutions/evidence/ch19/original-grade.txt)
and [provenance](../../solutions/evidence/ch19/original-grade-provenance.json),
independently matched the recorded source tree and raw SHA-256
`bd940a590b17495f5789b0a6e39fa9c825e7e7997d5151ca83a3c914f92e1a6a`.
Coordinator retained these at d44532b. No repeat was run.

Six failed checks (85points) report the same actual GUI startup rejection:
`skill "base" has invalid type ""`. Five first failures name scenario A and
refresh names B. This confirms the preflight's fixture mismatch, not a discovered
Responses, caching, refresh or reasoning implementation failure. From the
unchanged launch order, reaching those GUI errors means their fake login command
returned successfully; it does **not** establish all OAuth validation claims,
because the later positive predicates were skipped. It is not real plan login.
The first-failure aggregation does not separately expose every later scenario's
observations, so do not infer them from repeated labels.

The awarded billing15/15 supplies no positive billing proof. The unchanged C
predicate explicitly accepts a launch fatal and requires neither a successful
inference request nor a quota refusal. The common untyped-skill launch prevents
inference; the aggregate PASS is therefore consistent with—and explained by—the
preflight's vacuous path. No per-scenario positive C trace appears in this output.
Likewise the bad-key scenario's absence of inference cannot alone separate
correct signature rejection from unrelated startup failure. Preserve the actual
15/100 rather than calling this an85-point implementation failure or a demonstrated
billing success.

Cache-route restrictions, default credential selection, refresh error classes,
summary incrementality, replay and real tool outcomes remain **masked obligations**
for final independent review. No cache rejection or private-save-path problem was
observed by this run. The resource.invoke request/grant distinction remains the
corrected observation above, not an accepted fixture defect. The narrow missing-
type explanation is independently substantiated, but chapter acceptance and any
final exception decision await frozen implementation, full student response and
real supplemental evidence. No old answer or changing product was inspected in
this check; zero formal correction rounds. Required handoff:
`.agents/skills/ensemble-coding/SKILL.md`.


## Initial formal review — correction required, October 11

Read the complete student review through the production-plan attempt, original
chapter and governing documents; inspected the submitted source and matching
first-edition Chapter19 changes against each edition's own predecessor. Initial
source was `7a2d562` / `f2435030`; the independently approved cache-only adaptation
is `dc74e25` / `bb0a9102482209f95720bd2555b562820d7cba43`. Final acceptance is
**withheld for one Responses transport blocker**, comprising two observed wire
conditions below. Correction round1 is authorized; no other blocker is confirmed.
This is not a request for another auth architecture, new model, metered fallback,
full grader sweep or additional feature.

### Confirmed finding and independent reproduction

The compatible plan request passed admission, but the actual CLI still failed:
`invalid Responses response`, no completed task, exit0. Its captured body begins
with SSE `event:`/`data:` records, while the parser selected the whole-JSON path.
The same capture delivers a function call in `response.output_item.done` and then
`response.completed` with completed status, usage831 input/20 output and an empty
output array. The current parser ignores the completed item and obtains final
parts only from that empty terminal array. Correcting delivery detection alone
therefore cannot complete the tool loop.

The [independent probe](../../solutions/evidence/ch19/reviewer-plan-wire-probe.go.txt)
uses an isolated exact `dc74e25` archive, the real public Ensemble/Agent/Engine,
and an external HTTP server replaying those captured bytes. Under a controlled
application/json header it reproduces `invalid Responses response`; changing only
the header to text/event-stream produces `Responses response has no content`.
A positive control copies the already-completed item into the terminal output and
succeeds. These controlled headers are not claims about an unrecorded production
header. Preserve that distinction until the coder recovers exact header evidence.
The [receipt](../../solutions/evidence/ch19/reviewer-initial-probe.json)
records source/body hashes, commands and observations. Its PASS means both defects
and their control were reproduced, not that the submitted parser passes review.

Preserve incremental delivery and part identities, the actual completed tool
item/result, authoritative terminal success and usage, and failure on truncated
streams when correcting these two supported wire conditions. The initial fakes
supplied an SSE Content-Type and full terminal output, so their passing tests did
not cover the live shape. Captured-wire tests must exercise the real parser and
show restored defects failing for their intended reasons. No old solution code
or implementation recipe was supplied to the student.

### Independently accepted narrow exceptions

The original **15/100 FAIL on06cbf20** remains unchanged. I accept the documented
missing-type fixture exception: it prevents six checks from reaching the Agent;
the awarded billing15 is vacuous, not demonstrated routing. The valid masked
obligations remain binding and are assessed from supplemental source/tests/live
evidence. No unconditional waiver of authentication, refresh, tools, summaries or
billing follows from startup failure. Deferring chatgpt_then_api is expressly
allowed by “Yours”; refusing it and retaining selected billing on errors is valid.

I separately accept the live-supported exception to unconditional explicit plan
caching. `live-plan-rejection.json` records HTTP400 invalid_parameter naming
prompt_cache_options on the selected ChatGPT route; the same model's API-key
request was accepted. The renderer sends options and markers as one explicit-mode
mechanism, matching Chapter18's documented opt-in. Omitting the paired mechanism
for ChatGPT is a proportionate compatibility choice, while keeping API-key
explicit caching and both routes' lens. **Only options rejection was independently
observed; marker omission is not a claim of separately observed marker rejection.**
The small route-data delta and restored-plan-cache mutation support this accepted
departure. It does not authorize another billing source or establish a cache grant.

### Comparison, KISS and documentation

[Counts](../../solutions/evidence/ch19/reviewer-counts.json) use tracked
source lines, excluding blanks/comment-only lines from active code; inline
comments remain active. Tests, languages and example commands are separate;
module files, Markdown/design documents, runtime artifacts and support evidence
are excluded. Each edition is measured against its own preceding chapter.
No student source move disguises the delta.

| Chapter delta | Student physical / active / comment-only | Original physical / active / comment-only |
| --- | ---: | ---: |
| Production Go | +1608 / +1362 / +182 | +4724 / +2526 / +1783 |
| Go tests | +687 / +645 / +26 | +4752 / +3714 / +606 |
| Production JavaScript | +5 / +5 / 0 | +54 / +32 / +19 |
| JavaScript tests | +53 / +50 / +2 | none |
| Example Go command | unchanged | +45 / +28 / +15 |

Original HTML adds4 active lines; CSS adds one comment. The student's three
existing nested consumer/workflow/stream examples are unchanged. Original19 also
contains a large design document and unrelated tool/cache/reset/MCP growth; those
are not a target feature inventory for this exercise. The student's smaller
production delta is meaningful because it retains the required protocol and
public clients rather than compressing syntax or deleting explanations. Three
auth implementation files, shared public auth CLI, existing Engine parser/renderer
and a small route-data table are proportionate; no generalized auth framework or
future provider-managed compaction is introduced.

The original has extensive useful WHY prose about registration, rotating tokens
and final/delta identity, but also old structural limitations: its common package
contains credential/route behavior, its engine directly holds sibling services,
and its command/library attachment paths differ. The student retains actual
Ensemble-owned auth, Agent-owned configuration and Engine-owned transport/usage;
GUI and headless clients construct the public root and share the small auth CLI.
Common adds data/interfaces and standard JSON methods, not auth implementation.
The browser-opening callback is a concrete client boundary, not a service bag.
No new mutable application globals or mocked earlier runtime were found.

New production comment share is about12% of net nonblank Go growth, below the
rough20% teaching target. Useful attached explanations cover ownership, refresh
serialization, registration persistence, request-local tokens, safe error prose,
opaque provenance and terminal authority. Exported new symbols/fields/interface
methods and the changed test entrypoints carry purpose comments; no missing
exported documentation finding is open. The count is not a padding quota. The
original's much higher comment share includes long design essays; retaining its
useful reasons does not require copying those essays or its architecture.

### Tests and honest evidence limits

Reviewed all35 recorded mutation outputs and their actual test boundaries, plus
the36th paired-cache control. They fail the intended identity/grant/refresh,
selection/redaction, route, tool/replay, summary/completion/usage, lens and browser
predicates rather than compilation. The summary timeout is causal: the external
provider is gated until the live summary callback/browser pipeline observes it.
Tests use real signed JWTs, local issuer/HTTP peers, the real Agent/tool loop and
built CLI/browser; no mocked job/actor/auth implementation replaces the behavior.
The independent exact-source OAuth/Responses race run passes3.146s. It still does
not test the two captured conditions above. Positive fake results and mutation
sensitivity do not establish completeness of their protocol assumptions.

The API-key live session is one prompt plus hint continuation, three requests,
90 summary deltas, actual read_file and an optimal makespan13 answer. Its zero
cache counts remain zero; the offline lens replay is a byte-prefix diagnostic,
not a provider grant. Production sign-in is real but performed by Bill/coordinator;
I did not inspect the dedicated credential profile. The first plan task failed
at admission; the paired-mode repeat failed in parsing. Neither exit0 establishes
success. Existing Node speech-pause failure and isolated repeat remain an
unresolved intermittent observation, not a proven pre-existing defect or a new
speech feature requirement. Final chapter disposition awaits the corrected
source, complete updated student response and meaningful plan-route evidence.

All reviewer work used source/evidence and isolated scratch only: no production
profile access, paid request, grader rewrite, original-source edit or future
chapter work. Preserve CodeRhapsody's commits/working changes. Required handoff:
`.agents/skills/ensemble-coding/SKILL.md`.


## Correction round1 — independent deterministic reassessment

Source `534112a`, main tree `c72d29294763f23f51071d7484b0d4283ee70c33`.
Read the complete appended student response and exact production/test delta.
The correction resolves the consolidated parser blocker without changing auth,
parent ownership, billing policy or the accepted cache exception. A six-byte
buffered lookahead does not consume or wait for the complete body. Completed
items retain output-index order only when terminal output is absent; terminal
success and usage remain prerequisites for response facts. The live header was
not retained; recovery of that header is **not** an acceptance obligation and
warrants no additional provider request solely for diagnostics.

The unchanged initial diagnostic probe now observes nil errors for all three
original captured-body/control cases. Its old assertions expected those errors,
so that literal invocation exits1; this is preserved, not mislabeled a passing
regression. Replacing only obsolete expectations with successful completion and
adding exact captured call-ID/read_file and usage831/20 assertions passes.
Together with the real-Agent captured-tool/result regression, misleading-header
summary gate and incomplete/failed completion checks, the focused race run passes
**1.125s**. See [the round1 receipt](../../solutions/evidence/ch19/reviewer-round1-probe.json)
and [updated success assertions](../../solutions/evidence/ch19/reviewer-round1-plan-wire-probe.go.txt).
One initial launch used a nonexistent placeholder scratch directory and executed
no test; the corrected path produced the recorded results. Both reviewer scratch
archives were removed afterward.

Read all three restored-defect outputs: absent framing recognition fails invalid
JSON; dropped item.done fails no-content; whole-body sniffing blocks the causal
summary gate until its intended deadline. No compile failure is counted. New
production delta is27physical/21active/6comment-only Go lines; test delta is
68physical/57active/10comment-only lines plus a39-line captured SSE fixture,
separate from production. Final chapter net Go production thus adds
1635physical/1383active/188comment-only lines; Go tests add755/702/36.
Added anonymous test fields have attached useful comments. The raw captured-body
hash independently matches `3d6bf5b09bd59c0260e962a5facf54189d44b62edf63572ed53ea804a3511844`.
The reduced fixture accurately retains the relevant observed events and fields.

No deterministic blocker remains. **Final chapter disposition still awaits the
bounded live plan result and its complete student reconciliation.** Original
15/100 FAIL, initial/repeat task failures, actual source provenance, accepted
exceptions and live cache limits remain as recorded above. This is round1 of the
maximum three, not a new initial review or automatic whole-chapter pass.


## Final disposition — accepted with recorded exceptions after round1

Read the complete final student handoff at `2338b108e550548f9d827db34daa950c0f4d99f8`.
Independently confirmed product source still matches
`534112a2d6f9222577db94159c8bd3dbb8368376`, main tree
`c72d29294763f23f51071d7484b0d4283ee70c33`; later changes are notes, evidence and
launcher completion detection only. **Original19 / Edition2 Chapter18 is accepted
with the missing-type grader and paired plan-cache exceptions stated above,
after one correction round.** No current implementation finding remains open.
Original grading remains **15/100 FAIL on06cbf20**, never a final-source pass.

The [independent final live check](../../solutions/evidence/ch19/reviewer-round1-live-check.json)
verifies two actual completed plan responses, the correlated read_file result in
request2 and the valid optimal makespan13 answer. Both requests select
GPT-6.1-Sol/high/detailed, store:false, stream:true and omit the paired explicit
cache mechanism. Announced ChatGPT billing and the launcher's absent API key
establish the recorded source; I did not access credentials. Observed usage is
831input/20output and957input/337output, with zero cache writes/reads; 166 text
deltas and zero summary deltas. Completion took14.887seconds. No plan cache grant,
summary delivery or hint use is claimed from this session.

The outer launcher subsequently killed the idle CLI at150.03seconds, exit-9,
because its completion path depended on a summary-triggered hint. I accept the
student's distinction: the task succeeded, but normal EOF/shutdown save did not
run and is not demonstrated. Its small launcher-only correction has offline
positive/negative evidence, not a retroactive live pass. There is no reason to
spend another inference merely to induce optional public summary output or
repeat unchanged shutdown behavior. Actual API-key summary/hint evidence and
causal runtime/browser delivery checks remain separate. Production renewal and
terminal-error paths retain external-peer evidence; no forced live token expiry,
revocation or account mutation was required for acceptance.

The parser correction and real task establish the behavior masked by the original
fixture alongside the reviewed OAuth/JWT/rotation, route/billing, result/replay,
usage and incremental tests. All39 recorded targeted failures have appropriate
behavioral reasons, with setup failures and initial missing fixture coverage
retained. Source documentation/ownership and the final small delta remain
acceptable under KISS; no further feature, framework, mock or paid test is needed.

Stop here for coordinator-owned direct Edition2 ch18 export/checkpoint. Do not
advance original20: Bill's exact GPT-5 Astra destination is absent from the observed
catalog, and GPT-6 Astra is not an authorized substitution. Coordinator owns that
separate decision gate. Original22/24 stops and stopped authorship remain binding.
Preserve CodeRhapsody's Edition1 commits and working changes. This review touched
only its own record and sanitized probe receipts, with no reset, tag move or
production modification. Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.
