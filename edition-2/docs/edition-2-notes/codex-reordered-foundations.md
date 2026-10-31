# Reordered foundations reconstruction

October 14, 2026. Coder: `reordered_foundations_coder`, GPT-6 Astra/high.
Status: reconstructed accounting, authentication, historical crossover and web/status
and sandbox stages accepted; combined GUI reuse/fleet source plan awaits review.

This is maintenance reconstruction with explicit access to reviewed answers,
not a blind student run. Bill approved current17–23 → new15–21 and current15/16
→ new22/23, staged auxiliary accounting, and independently validated replacement
exports. This preparation does not accept an export or replace the full runtime.

## Source and ownership

Work only in `/Users/bill/projects/ensemble-reordered`, branch
`codex/reordered-foundations`, baseline
`a8eec84a32523987eb6574c7c9c2dd9cdb927275`. Initial status is clean.
The separately developed full application in `ensemble-compaction`, the main
checkout, frozen exports/manifests, grader IDs, chapters and publication assets
are outside this stage. Bill's preface remains untouched.

The starting cumulative source is read-only `solutions/ch14`, tree
`0c751dbd3985dee1a80284461e31346471ccc984` at the baseline. Stage that exact
tree under this branch's `agent/`, then add the scoped changes below. Its old
checkpoint is `edition-2-ch15-r1` (original numbering).

Accounting source is accepted `solutions/ch17`, tree
`2ee06bb6189898809d80de97df2b73c695cfb0d5`, checkpoint
`edition-2-ch18-r1`; identify its feature delta against accepted `solutions/ch16`,
tree `83e2f4cc85eb91ef9f62a54f638baa4d538cacf8`. This delta is a source guide,
not permission to copy the cumulative17 tree. The inspection already found
memory/recall references in marker fallback and mixed accounting tests.

Full preparation read: AGENTS.md, `.agents/skills/ensemble-coding/SKILL.md`,
workflow, carryover, current Chapters14/15/17/18/19, the compaction handoff,
progress record and dependency audit. No chapter prose is authored here.
Reload the mandatory coding skill and governing documents after compaction.

## Bounded implementation selection

1. Retain Chapter14's actor, History, journal, tool ladder, survivor ordering,
   settings and separate GUI module. No Memory/Recaller construction, bands,
   compactor, judge, forcing row, fleet, auth or Responses transport enters new15.
2. Port Engine-owned cache lens, serialized comparison/artifact rotation,
   request-local prediction attribution, counts by model/tier, derived meter,
   and reviewed pricing rows. Preserve the current-process/latest-main
   distinction without adding unused auxiliary entry points.
3. Port Anthropic tool/system/retained-handoff/rolling markers and supported
   OpenAI Chat Completions markers/explicit opt-in. Preserve stable-end exclusion
   of ephemera and actual Chapter14 declaration placement. `compactionBound`
   uses the newest surviving handoff only; no memory-kind fallback yet.
4. Add the public `UsageMeter` alias, `Pricing`/`UsageMeter` vocabulary,
   `ModelFeatures` pricing/explicit-cache fields and Engine `Meter`,
   `AttachCacheLens`, `CacheLens` methods. Keep common declarative and parent
   links real: Ensemble constructs Engine and its lens; GUI reads Engine.
5. Port GUI usage frames on response/end/replay, the four-number header and
   session tooltip. Do not import the memory settings changes from15/16.
6. Adapt only tests actually affected by the new Anthropic block-array system
   encoding and added markers. Reuse real runtime tests and external HTTP fakes.
   Retain the overlap/race/attribution probes and actual renderer fixtures.

Two scoped correctness carry-forwards were approved with explicit provenance:
the supported-model correction `215ae6ac` (remove unsupported Anthropic 5.5
rows and replace active fixture IDs with supported older rows, including
`-5-5` spellings), and transport correction `09111e0b` (no total 60-second
response deadline; preserve explicit cancellation and connection safeguards).
Neither brings later feature machinery forward. Carry its focused lifetime
tests if compatible with Chapter14's public API; report any adaptation.
Preserve historical model identifiers and measured evidence unchanged.

## Obligations deferred, not removed

| Introduction | Required accounting integration |
| --- | --- |
| New22, memory | Surviving-memory marker fallback when a handoff was retired; actual child compressor model/tier usage included in owner session totals exactly once; main latest-request counts/forcing untouched; child calls excluded from main lens; replay creates no spend. |
| New23, recall | Stateless judge model/tier usage included in owner totals; known/unknown auxiliary-price behavior; judge excluded from main lens and latest-main counts; recorded recall/handoff interactions retained. |

The mixed old17 `cache_test.go` sections exercising `Judge`, `Compress`,
`KindSession`, `RecallSettings` and the `recallPeer` helper will not be copied
as hidden prerequisites. Main-send/re-render checks get a small ordinary
provider peer, with no fake internal Engine or Memory. The explicit deferred
cases above must return at their introduction stages. The current17 numerical
grader has no handoff/ephemeral fixture and cannot alone establish these claims.

## Validation and review gate

Run every command in a child environment stripped of inherited `LLM_`,
`OPENAI_`, `ANTHROPIC_`, `GEMINI_`, `EN_` and `CODEX_` variables. Use explicit
`./agent` for graders; retain old chapter IDs until Stage E. No live calls,
deployment changes, secret access, new export/tag or authentication work now.

- Establish accepted14 core/GUI tests and vet baseline after staging, before
  porting accounting; retain any setup or behavioral failures honestly.
- Run gofmt, core/GUI tests and vet, affected race tests, old14 context grader
  and old17 invoice grader through `go run ./cmd/grade-edition2 -ch N -json ./agent`.
  Install only ignored locked GUI dependencies if its actual harness needs them.
  Run affected separate ch05/ch06 example modules as well; the root module does
  not include their tests. Coordinate grader timing with the coordinator.
- Exercise main lens rotation and real-byte marker stripping, handoff and
  pre-ephemera positions, all-provider usage normalization, known/unknown
  pricing/tier arithmetic, process reset, GUI meter and interrupted response
  attribution. Reuse unchanged historical mutation evidence where applicable;
  perform focused scratch mutations for adapted boundaries and actual gaps.
- Compare staged source against accepted14, separating production, comments,
  tests, examples and languages. Record retained source identities and removed
  future prerequisites. A large diff against baseline's full `agent/` is expected
  in this preparation branch and is not a patch to merge into the full runtime.

After a coherent tested stage, checkpoint only owned `agent/` and this note in
this branch and submit it to `compaction_dependency_audit` for independent
scope/architecture/comparative review. Coordinator/reviewer acceptance is required
before this plan's implementation and again before advancing current18/new16.
Stage E export/renumber/publication work remains a later gate after full-runtime
behavior acceptance. The final full application cannot be replaced silently by
this deliberately partial cumulative source.

Independent source-plan review: `compaction_dependency_audit` accepted the plan
before implementation, verifying all three source trees and the narrow 09111e0b
transport change. The added separate-module validation above incorporates its
one clarification. Coordinator accepted the concrete plan and released this
bounded implementation.

## Implementation and initial setup failure

The isolated `agent/` now starts from the exact accepted14 tree. Selected
16→17 changes provide the lens, marker rendering, usage/pricing types and GUI
meter. Engine keeps its single Chapter14 send path; it observes the actual sanitized
send and carries that request's prediction to its own response. There is no
unused auxiliary send API. Handoff markers use surviving Chapter14 entries;
memory fallback, recall and auxiliary tests remain deferred as listed above.
The three ordinary provider replies replacing `recallPeer` are external HTTP
fixtures; actor, transport, parsing, ownership and History still run.

The approved 215ae6ac correction removes 5.5 catalog rows and changes active
test selections to Sonnet-5. The original captured
`ensemble/testdata/anthropic-empty-tool-args.sse` stays byte-identical, including
its historical model identity. The 09111e0b timeout correction is carried forward;
its real 61-second and cancellation tests use this stage's existing OpenAI Chat
Completions wire instead of introducing Responses, OAuth or recall settings.

Before accounting edits, exact14 baseline results in the stripped child
environment were: core `go test -short ./...` PASS (11.467s), GUI
`go test ./...` PASS (1.422s), core/GUI `go vet ./...` PASS. After the port,
the first core test run hit `no space left on device` during test-binary and
temporary-directory creation; GUI also reported failed observations amid
temporary-directory failures, and `npm ci` failed. The host had 142 MiB free
and an 11 GiB shared Go build cache. Those runs cannot establish runtime
regressions or successful validation. At that point root coordinated regenerable
cache cleanup before grading began. Raw output is retained under
`/tmp/ensemble-reordered-checks/` with `baseline-*`, `accounting-*` and
`gui-npm` filenames. Clean rerun results follow below.

## Submitted validation and comparison

Root cleared only the regenerable shared Go build cache after coordinating with
the other workers. The clean reruns below used the same stripped child
environment; the earlier disk-full outputs remain recorded.

| Command, relative working directory | Actual result |
| --- | --- |
| `agent`: `go test -short ./...` | PASS, 17.762s including cold build |
| `agent/gui`: `go test ./...` | PASS, 7.064s including cold build |
| Core and GUI `go vet ./...` | PASS |
| `agent`: `go test -race ./ensemble -run 'TestCache\|TestBlockedModelResponseCancellation' -count=1` | PASS, 9.442s including race build |
| `agent/gui`: `go test -race ./...` | PASS, 6.188s |
| Root: `go run ./cmd/grade-edition2 -ch 14 -json ./agent` | PASS 100/100, 1.670s |
| Root: `go run ./cmd/grade-edition2 -ch 17 -json ./agent` | PASS 100/100, 0.602s; ran after 14, no overlapping grader |
| `agent/ch05`, `agent/ch06`: `go test ./...`, `go vet ./...` | PASS after the ch05 fixture correction described below |
| `agent/ch07`: `go test ./...` | Builds; reports no test files, not behavioral proof |
| `agent/gui`: `node --test cache.test.mjs` | PASS, 1.553s; actual browser known/unknown price display and tooltip |
| `agent`: `go test -race ./ensemble -run TestModelResponseBeyondOneMinute -count=1` | PASS, 70.046s including build; both actual 61-second replies complete |
| `agent`: final `go test ./ensemble -run '^TestCache' -count=1` | PASS, 0.520s after scratch mutations were removed |

The supported-model browser fixture run also passed speech and actual client
shutdown with a sleep job still active. Its separate peer idle-EOF case exceeded
its tight bound while other checks ran. An exact14 scratch control then passed
(6.797s); the staged test run alone passed (6.696s). The peer implementation is
unchanged from14. This is an observed intermittent timing failure with an
unproven cause, not a reproduced baseline defect or a waived runtime guarantee.
Both outputs remain; no timing threshold or runtime code was changed to pass.

The standalone ch05 example initially failed its assertion that the root log
ends with `shout called`: accounting correctly emits later request diagnostics.
The test now selects the actual tool record and validates that record's timestamp.
This preserves ownership/logging coverage rather than requiring silence from
new diagnostics. Production example code is unchanged.

Five focused scratch mutations failed the intended behavioral checks:

- Select dialogue rather than the handoff as cache boundary: both supported
  explicit-cache renderers lose the settled handoff marker.
- Observe every `Render`: all three vendors fail because rendering without a
  send changes the current lens artifact.
- Charge Sonnet ordinary input at 1 instead of 2: the independently specified
  usage/cost comparison reports 0.100213 instead of 0.200214.
- Omit the actual external tool's logging call: the corrected ch05 test fails
  despite surrounding cache diagnostics, proving they cannot supply its credit.
- Restore the real 60-second HTTP client deadline: both 61-second lifetime cases
  fail at 60 seconds with the intended deadline errors. Correct production had
  passed both; scratch was removed without changing the submitted source.

The first omission-mutation launcher used the wrong local variable name and
stopped before applying that mutation; it is setup failure, not a detected defect.
The corrected omission run above executed the real example. Other unchanged
lens/marker/meter behavior retains the prior targeted evidence and limitations in
[the original18 record](ch18.md) and
[its accepted comparative review](ch18-code-review.md), including overlap lock,
request prediction attribution and rotation defects. This stage reran the actual
overlap tests under race, but does not relabel historical mutations as new runs.
The deferred auxiliary and surviving-memory claims receive no current credit.

Comparison against each accounting stage's own predecessor:

| Source category | Accepted14 → staged accounting | Accepted16 → old17 accounting |
| --- | ---: | ---: |
| Production Go, physical / active / comment-only delta | +502 / +388 / +85 | +516 / +405 / +84 |
| Go tests, same counts | +685 / +631 / +32 | +598 / +547 / +35 |
| Separate Go examples including tests | +8 / +6 / +2 | 0 / 0 / 0 |
| Production JavaScript | +8 / +8 / 0 | +8 / +8 / 0 |
| JavaScript tests/harnesses | +45 / +42 / +2 | +45 / +42 / +2 |
| HTML | +1 / +1 / 0 | +1 / +1 / 0 |

Counting method: tracked source plus new nonignored source; nonblank physical
lines not beginning with `//` are active, `//`-only lines are comments. Example
modules are separate from production and primary Go tests. CSS, docs, captures,
dependencies and generated files are outside these code counts. No code was
compressed to reduce them. Current production totals are 9,129 physical,
7,168 active and 1,452 comment-only lines; the teaching balance stays close to14's
roughly one comment per five active lines. The smaller production delta omits
future memory fallback and unnecessary auxiliary plumbing. The larger test delta
includes the 151-line transport lifetime/cancellation regression; old17's mixed
auxiliary tests were deferred. The fixture-only ch05 correction adds eight lines.

Reviewed source scope is 118 files versus Chapter14's 108. Changes from14 are confined
to 31 paths: accounting production/integration, encoding-sensitive existing tests,
supported-model fixture selections, and the timeout correction/regression.
The preserved raw capture is byte-identical. Static inspection finds no Memory,
Recaller, fleet, memory-kind, forcing-window or auxiliary-call API. No snapshots,
manifests, grader source, chapter prose, historical evidence or tags changed.
No live inference, pricing re-verification, deployment or publication was performed.

Submit this stage for independent review before any authentication work. This
branch's deliberate removal of later `agent/` features is reconstruction staging,
not a mergeable replacement patch for the complete application.


## Accounting acceptance and authentication source plan

October 14. Independent reviewer `compaction_dependency_audit` accepted
`288fc6e8f866bc19d65c17909f270f294c90d1de`, Agent tree
`2b671b45dc5f7b760f2faaff2a77c83e6c7c09c8`, at initial review with zero
correction rounds. Review commit `7245ff935803664035a1703035a924261d60238b`
on the primary branch records the comparison and independent race check.
The coordinator read that review and accepted this bounded stage. It remains
maintenance reconstruction, with no frozen export, publication or full-runtime
replacement implied. Do not merge the primary branch's complete Agent tree.

The next bounded stage is current Chapter18 → proposed Chapter16, authentication
and Responses. Its predecessor is exactly the accepted reconstructed tree above.
Use accepted17 `2ee06bb6189898809d80de97df2b73c695cfb0d5` → accepted18
`c72d29294763f23f51071d7484b0d4283ee70c33` as the reviewed delta guide.
Accepted18 is source `534112a2d6f9222577db94159c8bd3dbb8368376`, preserved by
`edition-2-ch19-r1` under its original numbering. The historical implementation
was accepted after one correction round; this reconstruction does not relabel
that history as a fresh student run.

Preparation read the complete current Chapter18 (880 lines), original-numbered
`ch19.md` (545), `ch19-code-review.md` (424), enhanced Chapter18 review (225),
the accepted accounting review, the 32-path source delta, auth implementation,
Responses renderer/parser and relevant tests/public clients. The mandatory
`.agents/skills/ensemble-coding/SKILL.md`, full workflow and carryover remain
loaded after the latest compaction. This is source planning, not manuscript
review or a new claim about current provider support. No external account or
provider documentation refresh is needed to reconstruct the accepted behavior;
a future live checklist must separately verify current contracts.

### Selected changes and public contracts

- Port the reviewed `internal/auth` service, `internal/common/auth.go`, public
  `ensemble/auth.go` and shared public-client `authcli` package. Ensemble owns
  AuthStore; rendering stays independent of credential acquisition. Preserve
  PKCE/state/nonce, issued registration before exchange, signed identity and
  granted-scope checks, serialized cross-process rotation, atomic private
  replacement, terminal versus transient errors, revocation and safe status.
- Add Config's request-only CredentialProvider, BillingMode and reasoning
  effort/summary fields, the ModelFeatures Responses flag, Responses provenance,
  Part.Namespace, DeltaReasoningSummary and BillingSelected. Preserve documented
  public aliases and ConfigureAuth/Auth. No new generalized credential framework.
- Port auth commands into both existing executables, explicit saved OpenAI
  billing choice, API-key behavior for other vendors, pre-inference billing
  announcement and one same-source renewal after a 401. Only `api` and `chatgpt`
  are supported; automatic metered fallback remains explicitly deferred.
- Port Responses rendering/parsing, capability data, safe provider-error mapping,
  model-specific route selection, History's valid-surface bound, GUI/terminal
  summary and billing handling, and the lens's actual Responses `input` support.
  Only the reviewed `gpt-6.1-sol` and course model rows move to Responses now;
  older rows retain Chat Completions. Prices remain as recorded or unknown.
- Adapt the transport additions into this branch's existing single Send path.
  Preserve sanitized actual-send lens observation, request-local prediction,
  latest-main usage and per-model accounting; the 401 retry is the same logical
  request and successful usage is accounted once. Do not recreate the auxiliary
  `send(cfg)` seam from the old cumulative source before there is a caller.
- Preserve terminal-completion and usage requirements, incremental framing
  detection, completed-item recovery when terminal output is empty, actual
  correlated tool results and model/surface-bound opaque replay. API-key explicit
  cache markers use the surviving handoff boundary; the accepted plan capability
  omits options and markers together. Do not introduce a memory fallback.

The old delta's only production judge change is a four-line Responses request
branch in `internal/engine/judge.go`; defer it to new23. Four new test Config
literals contain `RecallSettings{Disabled:true}`; omit those fields because the
service is absent, retaining every actual OAuth/Responses assertion. The auth
README addition may be appended to this branch's earlier README without copying
its old memory/recall sections. No memory, recall, fleet, sandbox, web, Codex
backend route or new model-switching controls enter this stage.

The accepted supported-model and no-total-model-response-deadline corrections
remain intact. The historical auth service's bounded 30-second exchange client
is separate from the unbounded model-response lifetime and remains unchanged.
No additional maintenance correction is proposed. The recorded callback browser
EOF race is an accepted usability limitation, not a reason to claim authentication
failed or to silently repair frozen history. If it affects a test, retain actual
exchange/process evidence and assess that specific failure.

### Focused validation and deferred integration

Use the existing stripped child environment and a fresh temporary EN_HOME for
all launched command tests so even status/configuration cannot inspect the user's
profile. External OAuth/JWKS/Responses fakes run the real library, command and
browser paths; no mocks, real sign-in or inference. Preserve the historical
`ch19-plan-tool.sse` fixture bytes and the earlier Anthropic raw capture identity.

Run gofmt, core/GUI tests and vet; focused OAuth/Responses/cache/cancellation race
checks; and the existing real 61-second lifetime regressions because Send changes.
Run the affected separate examples (ch05/ch06 tests and vet, ch07 build). Run the
Responses browser summary test and existing accounting browser test sequentially;
expand only for an observed failure or directly affected behavior. Coordinate
root grading: old18 with explicit `./agent`, then old17 and old14 regressions,
serialized and retaining their current identities until Stage E. No grader edits.

Retain the accepted signed-JWT, scope/renewal/error, actual tool, provenance,
terminal-authority and browser summary assertions. Strengthen the existing
401-retry fixture only as needed to check successful usage exactly once and its
latest-main meter, with lens rotation still tied to the logical request. Reuse
unchanged historical mutation evidence in ch19 notes/review, preserving its
39 recorded defects and limitations as historical. Focus new scratch mutations
on the adapted credential/401-accounting boundary and the Responses input lens;
rerun inherited relevant regressions rather than rebuilding a whole mutation
matrix. Any new behavioral claim needs its own intended-failure evidence.

New22 child compression must inherit the selected funding/route correctly and
charge its actual model/tier to the owner exactly once while preserving main
usage/forcing and excluding child traffic from the main lens. New23 must restore
the deferred stateless judge Responses branch, shared credential safety and
same-source failure behavior, alongside the already recorded auxiliary accounting
obligations. These receive no acceptance credit from this stage.

After implementation, compare against the accepted reconstructed accounting
predecessor and historical17→18 separately, including comments, tests and actual
public-client wiring. Submit one coherent tested checkpoint for independent
review before historical crossover/current19. **Only this source plan is released
now; coordinator and independent reviewer must approve it before Agent edits.**
No export, renumbering, live use or full-application replacement is authorized here.


## Authentication implementation submitted for review

The independent reviewer accepted plan `185a1e31969eb94c2cbec54a9437455988ce1fba`;
the coordinator then read the complete plan and released implementation. No
formal implementation review or correction round has occurred yet. This stage
extends accepted accounting `288fc6e8` in the same isolated worktree.

The selected 17→18 source delta is ported in 31 paths. The four obsolete recall
settings were omitted, and the judge change remains deferred. Ensemble constructs
the real auth service; both executable clients use its public API. The existing
single Send resolves transport credentials before sanitizing request diagnostics
and lens artifacts. Its request-local prediction survives a same-source 401
renewal; only the eventual successful response contributes usage. No auxiliary
entry point or new model-switching mechanism was introduced.

The existing 401 integration test now asserts latest-main and total usage are
each exactly one successful response, and that no `cache-prior.txt` exists after
that first logical request. A separate small real-Engine test supplies rendered
fixture bytes containing fake configured and resolved bearers and checks that
both are redacted in the actual lens artifact. It uses the real signed issuer,
AuthStore, Engine and HTTP peer, not a mocked internal credential provider.
Existing actor/tool-loop tests remain intact. These additions protect the
adaptation that old18's separate main/auxiliary Send wrapper did not need here.

All checks used stripped provider/application variables plus a fresh temporary
EN_HOME per launcher. The host retained 11 GiB free at the start; there was no
new disk-full failure. Raw logs are `/tmp/ensemble-reordered-checks/auth-*.txt`.

| Command, relative working directory | Actual result |
| --- | --- |
| `agent`: `go test -short ./...` | PASS, 15.836s |
| `agent/gui`: `go test ./...` | PASS, 2.152s |
| Core/GUI `go vet ./...` | PASS |
| `agent`: `go test -race ./ensemble -run 'OAuth\|Responses\|TestCache\|TestBlockedModelResponseCancellation' -count=1` | PASS, 4.698s initially; final focused repeat PASS, 4.907s (3.815s package) |
| `agent/gui`: `go test -race ./...` | PASS, 5.531s |
| `agent/ch05`, `agent/ch06`: tests and vet | PASS; ch07 builds with no test files |
| Root: old18, then old17, then old14 with `go run ./cmd/grade-edition2 -ch N -json ./agent` | Each PASS 100/100; 1.262s / 0.453s / 0.768s; serialized |
| `agent/gui`: `node --test responses.test.mjs` | PASS, 1.101s; actual browser summary delivery before completion, without duplicate final |
| `agent/gui`: `node --test cache.test.mjs` | PASS, 1.385s; run after the Responses browser test |
| `agent`: `go test -race ./ensemble -run TestModelResponseBeyondOneMinute -count=1` | PASS, 62.474s; 62.018s package, both real 61-second responses |
| Three adapted boundary tests after scratch cleanup | PASS, 0.802s |

Final source cleanup removed a stale judge-only comment and two redundant local
model lookups, reusing the already computed request features. The final focused
race repeat covers the resulting source. Earlier full/grader/lifetime timings
retain their before-cleanup provenance; the cleanup does not change behavior.
There was no observed callback EOF or browser failure in these runs. The
historical callback acknowledgment limitation and prior-stage intermittent peer
EOF observation remain limitations; neither is declared fixed.

Four new isolated scratch mutations fail their intended assertions:

1. Account successful usage twice: the 401 test reports incorrect meter totals.
2. Observe the request again on 401 retry: the same test finds a spurious prior
   artifact and rejects duplicate logical rotation.
3. Sanitize the lens body using only the configured API key: the resolved-bearer
   test rejects the leaked fake OAuth bearer in the actual readable artifact.
4. Omit Responses `input` from the lens sections: the retained input-lens test
   rejects the missing conversation prediction and append observation.

`auth-mutations.json` records the four actual failures; all scratch sources were
removed. No compile/setup failure is counted. The unchanged historical 39 auth/
Responses mutations and their limits remain at their original source identities
in ch19 notes/review; they are not new reconstruction results. A first line-count
script used an invalid compound Git tree path and stopped before producing
counts; resolving the tree object first corrected this reporting-only setup
error. No runtime test or grader failed during this stage.

| Source category | Reconstructed accounting → auth | Historical17 →18 |
| --- | ---: | ---: |
| Production Go, physical / active / comment-only delta | +1631 / +1378 / +189 | +1635 / +1383 / +188 |
| Go tests, same counts | +799 / +741 / +40 | +755 / +702 / +36 |
| Separate Go examples including tests | 0 / 0 / 0 | 0 / 0 / 0 |
| Production JavaScript | +5 / +5 / 0 | +5 / +5 / 0 |
| JavaScript tests/harnesses | +53 / +50 / +2 | +53 / +50 / +2 |
| HTML | 0 / 0 / 0 | 0 / 0 / 0 |

Counting uses the same physical/nonblank/non-`//` method as the accounting stage;
docs, captures, generated files and dependencies are excluded. Production totals
are 10,760 physical, 8,546 active and 1,641 comment-only Go lines. The auth delta's
comment share is below the rough teaching target, as in the accepted source,
but preserves explanations of credential ownership, private state, cross-process
rotation, terminal authority, public summaries and safe request-local transport.
The small production difference excludes the judge branch and avoids unused
auxiliary plumbing; test growth is the 44 added lines for the adapted boundaries.
No framework or dependency was added. Public aliases, fields and methods retain
the accepted contract comments.

The preserved `ch19-plan-tool.sse` hash is
`11600e076cc924f5606e5aec2427994c209f2aef939d37c516c3ce87a4b19b39`;
the earlier Anthropic capture remains
`fddbc8bc038ec13e5710ea6c1e4d3439661f319e71a0133b2c0e27c0b1e74f3c`.
Both match accepted18 bytes. The staged whitespace check flags only the captured
SSE's trailing blank event separator at line 39; its exact historical bytes are
preserved rather than altered to silence that check. Static scope inspection
finds no memory, recall, forcing, fleet or sandbox machinery; only the preserved
raw Anthropic capture contains a historical 5.5 identity.

Independent implementation review remains required before current19/new17.
No real provider access, credential-profile inspection, export, grader change,
renumbering, publication or full-runtime replacement occurred. Required handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

## Authentication acceptance and historical crossover source plan

October 14. Independent reviewer `compaction_dependency_audit` and the coordinator
accept authentication commit `970676abeebce8f156d5cdc7ee0dd47a40708136`, Agent tree
`21b15bc6cd1c24a2ee3e03873a87a9c080744402`, at initial review with zero correction
rounds. The primary dependency-audit note records its independent OAuth/Responses
race result (1.236s package time), complete scope comparison and preserved limits.
The earlier submission section above records what was pending at submission;
this acceptance closes that gate without changing its historical test results.

The next bounded stage is current19 → proposed17, the historical crossover.
Extend that exact reconstructed auth predecessor. The source guide is accepted18
tree `c72d29294763f23f51071d7484b0d4283ee70c33` → accepted19 tree
`bd2998e68f976dbaf437b642b131d350e086242f`, final historical source
`bade50cef5cdc53490d4263fa57f5e37abbb93ec`, preserved by `edition-2-ch20-r1`.
That delta has exactly three paths, 19 inserted and eight removed lines. Do not
copy the cumulative19 source, which still contains deferred memory and recall.

Preparation read the full current Chapter19 (745 lines), original-numbered
`ch20.md` and `ch20-code-review.md`, and complete `enhancement-19.md`,
`enhancement-19-code.md` and `enhancement-19-review.md`. Inspected the complete
three-path delta and retained historical reconciliation/mutation receipts. After
compaction, reloaded the entire `.agents/skills/ensemble-coding/SKILL.md`, workflow
and carryover; recovered truncated ranges. This is maintenance source planning
with reviewed-answer exposure, not chapter authorship or a fresh student attempt.

### Selected changes

1. In `agent/internal/engine/media.go`, move only `gpt-6-astra` from its existing
   Chat Completions row to the existing Sol/course Responses row. Retain the
   accepted rationale comment: the existing per-Agent reasoning configuration
   then reaches the renderer without another configuration path. Preserve the
   supported-model correction and absence of future forcing-window fields.
2. In `agent/ensemble/responses_test.go`, extend `TestResponsesModelBoundary`
   with Astra, configure high effort, and assert the actual rendered model and
   reasoning effort for Responses rows while retaining the earlier Chat controls.
   This uses the actual Engine renderer; it adds no internal fake or auxiliary
   service. No recall-disabled fixture field is needed in this reconstruction.
3. Adapt only the historical README paragraph to this branch's shorter README:
   identify the two real Responses rows and explicit Astra/high/subscription
   configuration. Retain existing billing safety, summary distinction and
   authentication contracts. No new public symbol, configuration field,
   dependency, API surface or transport implementation is introduced.

The accepted single Send, credential redaction, once-only successful retry usage,
logical-request lens observation and no-total-response-deadline correction remain
unchanged. No new maintenance fix is proposed for this small stage. Existing
web/status, sandbox, fleet, memory/recall, auxiliary send/judge and Codex-route
capabilities remain at their separately authorized introduction stages. New22/23
still owe every accounting and route-inheritance obligation recorded above.

### Historical evidence and validation boundary

The retained actual coding task belongs to source
`fab031529cda55a76d530478f09de0cd08e7db36`, tree
`59010a4c2b333c2f74777b6d76079e8f2b3af90d`; only README changed before the
accepted19 source. Its 27.46-second, five-request Astra/high CLI run performed
real reads, a failing command, an exact edit and six unchanged passing tests,
then exited normally with a save. It reported zero cache grants and no public
summary deltas. Endpoint selection retains configuration/source provenance.
These are historical observations, not a new execution or proof of general
reliability, current account eligibility, GUI steering or Edition3 authorship.

The enhancement's 1.003-second local scripted replay is separate historical
evidence. Its extra recall-judge exchange and two fixture corrections do not
create an early recall prerequisite. Do not carry that launcher into this stage
or count its scripted decisions as autonomous coding. Current19 remains ungraded;
there is no new grader, numeric score or assignment to reproduce its repair list.

After source-plan approval, run gofmt, core `go test -short ./...` and vet, GUI
tests and vet, plus focused `TestResponsesModelBoundary` under race. No separate
example source changes are planned; the current examples do not select Astra.
Reuse unchanged auth/accounting, browser and long-response evidence. If a new
failure identifies an affected boundary, investigate that boundary specifically.
Use the same stripped child environment and temporary EN_HOME; no paid inference,
account access or fresh local task replay is required. No grader run is proposed
for this ungraded stage.

Repeat the two small historical boundary mutations in isolated scratch source:
restore Astra's Chat route and replace rendered high effort with low. Require
the intended transport/model-effort assertions to fail, then remove scratch and
rerun the focused correct test. Existing mutation evidence remains historical,
including its independently accepted provenance. Compare this stage against its
own accepted auth predecessor and the three-path historical delta, separating
production, comments, tests and README. No larger verification machinery is needed.

### Later maintenance integration remains required

The old accepted19 tree predates `df38abe0`'s visible-action-explanation correction;
neither this stage nor its historical task proves that enforcement is present.
Carry it explicitly to the later coordinated integration gate before the complete
reconstruction is accepted or exported: per-model capability/default-disabled
policy, ordinary same-response explanation before tools, actual streamed ordering,
whole-batch refusal without side effects, paired results, retained opaque history,
bounded retries and after-tool interruption. Prompt inheritance must preserve base
instructions without leaking a parent's model policy into unflagged children.
The synthetic course model remains a deliberate protocol fixture; preserved
historical tool-only captures must retain their actual identity and bytes.

The later integration must also retain the required main-chat/child-submission
distinction and associated prompt behavior from the complete application's
reviewed maintenance sources; do not assume old snapshots include those fixes or
import fleet plumbing now. The coordinator will place and release that bounded
integration with the relevant later capabilities. This ledger is an outstanding
obligation, not current implementation or acceptance credit. The already carried
`09111e0b` response-lifetime correction must survive those later adaptations.

Only this plan is submitted now. Independent source-plan review and coordinator
approval precede Agent edits; a coherent tested checkpoint then requires separate
implementation review. No exports, renumbering, book changes, tag movement or
whole-tree replacement of the complete application is authorized.

## Historical crossover implementation submitted for review

October 14. Independent plan review accepted `484c7234` without findings, recorded
in primary review checkpoint `fa29b7ee`; the coordinator then released this exact
bounded implementation. The three source changes follow the historical delta:
Astra selects the existing Responses row, the real renderer test checks selected
model/high effort and retained Chat controls, and README documents the explicit
configuration. No new public API or abstraction was added. The accepted auth,
accounting, supported-model and response-lifetime behavior remains unchanged.

Validation used stripped provider/application variables and a temporary EN_HOME.
Disk had 8.8 GiB free before checks. Raw outputs and the two small comparison/
mutation receipts are `/tmp/ensemble-reordered-checks/crossover-*`.

| Command, relative working directory | Actual result |
| --- | --- |
| `agent`: `go test -short ./...` | PASS, 14.589s launcher; ensemble package 13.356s |
| `agent/gui`: `go test ./...` | PASS, 1.928s launcher; GUI package 1.118s |
| Core/GUI `go vet ./...` | PASS, 0.263s / 0.247s launcher |
| `agent`: `go test -race ./ensemble -run '^TestResponsesModelBoundary$' -count=1` | PASS, 2.396s launcher; package 1.020s |
| Same focused test after scratch cleanup, without race | PASS, 0.241s launcher; package 0.008s |

Both planned scratch mutations failed their intended assertions: restoring
Astra's Chat route failed `model transport`; forcing low effort failed
`configured model/reasoning lost` on all three Responses rows. Both compiled
and executed the actual renderer. Scratch directories were removed; correct
production remained unchanged throughout. No setup or unrelated test failure is
credited as sensitivity. One initial source-reading command guessed a nonexistent
renderer filename; `rg` located `responses.go` before any mutation was applied.
That read error is not a runtime/test failure.

Against accepted reconstructed auth `970676ab`, production Go adds two physical
lines, zero active lines and two explanatory comment lines; Go tests add six
physical/active lines and no comment-only lines. README grows three lines.
These deltas exactly match accepted18→19 using the same nonblank/non-`//`
counting method. No other language, example, capture or dependency changed.
The comments retain the route rationale; no exported symbol was introduced.
The complete source diff is exactly three paths, 19 insertions and eight removals.
Formatting and whitespace checks pass. No mutation source remains in the tree.

Current19/new17 remains ungraded. No paid inference, new coding replay, profile
access, grader, book, export, manifest, tag or deployment changed. Historical
coding evidence and all deferred auxiliary/visible-explanation/main-chat-child
submission obligations above retain their distinct provenance and stage gates.
Submit this coherent stage for independent implementation review before preparing
the current20/new18 source plan. Required handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

## Crossover acceptance and combined web/status source plan

October 14. Independent implementation review `003ae75d` and the coordinator
accept crossover `078cccbd2e6b999b320289aa1d096f7abe394350`, Agent tree
`f94f414deb977d9f106b8d749a57faa934987049`, at initial review with zero correction
rounds. That ungraded historical stage has no export. Extend this exact accepted
predecessor for current20 → proposed18; the complete application remains separate.

The historical guide has two accepted steps, preserved separately:

| Step | Source / resulting Agent tree | Immutable checkpoint |
| --- | --- | --- |
| Accepted19 → original21 web | `f35b6ae8ce43c91a52b9972361093db6258e3e20` / `f351d66a7bc456650e05ac6656a7db59bbbd725a` | `edition-2-ch21-r1` |
| Original21 → original22 status, completing combined20 | `1da0891a7ea0f121b05225c62be3b2190e4f3918` / `a801628b1dd631e5ea20ad03ea7d255584728d31` | `edition-2-ch22-r1` |

The first step starts at accepted19 `bd2998e68f976dbaf437b642b131d350e086242f`.
The combined delta touches 23 paths: 1,124 insertions and 13 removals. Read full
current Chapter20 (1,163 lines), both original-numbered student/comparative pairs
`ch21*` and `ch22*`, and full enhanced20 coder/reviewer reports. Inspected every
changed production/test path, the actual prior GUI locking and Engine meter,
and the two narrowly relevant maintenance diffs below. The complete mandatory
`.agents/skills/ensemble-coding/SKILL.md`, workflow and carryover remain active
from the latest full reload; no intervening compaction. This is reviewed-answer
maintenance, not a blind student exercise or new provider/service verification.

### Selected capability and public contracts

- Port the MCP-owned HTTP transport, existing client context-send extension,
  URL/auth-env declaration/parser fields and manager connection case. Preserve
  JSON/SSE/202 handling, session echo, real job/connection cancellation, blocked
  receiver wakeup, redirect refusal, complete diagnostics and configured-token
  redaction in both errors and reflected RPC content. Common holds vocabulary;
  HTTP reaches diagnostics through its actual MCP → Agent → Ensemble parents.
- Port the loadable `web-search` skill and primary's explicit loadable grant,
  selected search/scrape tools and existing `wait_for_job`. Preserve declaration
  and actual dispatch gates, coherent primary-MCP startup and immediate remote
  unload. Retrieved bytes stay correlated tool-result evidence; no prompt-
  injection-immunity claim or new registry/protocol framework is added.
- Port `agent_status` as a formatter of `ToolContext.Engine().Meter()` and the
  owner's current model. Keep all four last-main/session buckets, cache writes
  in total input, known estimates and honest unknown cost. Existing per-model/
  tier ownership supplies accounting; no new counter or Engine API is needed.
  Add its shipped permission and update the existing builtin-count assertion.
- Port actor-serialized `Agent.SetModel`, retaining its public interface comment,
  empty/busy rejection and old resolved-identity clearing. Preserve provider,
  endpoint, credentials, billing, history and counts. CLI `/model NAME` and GUI
  `Settings.Model` use that same operation; GUI model patches remain separate
  from persisted preferences and broadcasts report effective state. No provider
  migration, discovery, automatic fallback or saved model preference is added.
- Adapt only those blocks of shared actor/common/GUI/README files. The old
  actor's adjacent `memorySettings` case, common's memory methods, GUI Memory
  field/validation/persistence and JavaScript memory controls stay absent.
  No production memory/judge change exists in the selected 19→20 delta itself;
  old cumulative context is not permission to import those prerequisites.

### Necessary predecessor adaptations

**GUI acknowledgment ordering:** accepted14 calls `updateSettings` while holding
the observation mutex. The new SetModel waits for the actor, which can already
be delivering an observation needing that mutex. Historical combined20 avoids
this cycle because it inherited reviewed correction
`05b43a0b46fa154201a8b3e97b22f96ae1047cb9` from old15. Select only that correction's
settings-order mutex and short snapshot/publication locking in `gui/gui.go` and
`gui/settings.go`. Keep patches/subscriptions ordered; never hold the observer
mutex across the actor acknowledgment. This adds `gui.go` to the 23-path guide.
Do not port Memory settings, validation, actor handling or the memory test.

Add one focused real-GUI regression to the new model test: gate actual streamed
delivery with an additional public observer, observe the model setter waiting
for the actor, then release delivery. A busy selection must return its error,
progress and shutdown must complete, and an idle positive control must still
apply/broadcast. The historical memory test's saved-file barrier is inapplicable
to a deliberately nonpersisted model selection; use the actual blocked setter
state, not a sleep-based guess. Restore only the lock cycle in scratch to verify
the regression's intended bounded liveness failure. No internal mock is needed.

**Supported fixtures:** apply the already accepted `215ae6ac` corrections to the
new `status_test.go`, CLI test and GUI model test. Its real status peer uses
Sonnet-5 then unpriced Opus-5, with independently calculated initial $0.000379
and tool-loop $0.000199, then unknown accumulated cost. Preserve model changes,
prior counts and resolved identity checks. The current20 grader independently
uses Sonnet-5 → unpriced Opus4.6 → Sonnet-5 and verifies unknown cost remains.
This does not freshly prove summing two priced Anthropic models; that claim
retains its historical/source evidence. Do not add a price row or use 5.5 merely
to restore the old fixture. Raw historical identities remain untouched.

The accepted `09111e0b` model-response lifetime and resolved-credential sanitizing
corrections remain unchanged. The new HTTP MCP resource also retains its reviewed
job/connection-controlled lifetime, without a blanket response deadline.

### Verification and remaining obligations

Run gofmt, core/GUI tests and vet; focused HTTP, status, model-selection and
existing settings checks under race, including the actual built CLI regression.
Use real external HTTP peers and WebSockets, plus one small actual-browser smoke
for the new model field and captured next-request model. Reuse the existing
browser launcher pattern in a temporary check, without another GUI framework.
Run affected separate ch05/ch06 modules and ch07 build because shared Agent/MCP
interfaces and builtin inventory changed. Preserve existing settings persistence,
subscription ordering and safe observer teardown; investigate actual failures.

Coordinate the single current20 combined grader run with root:
`go run ./cmd/grade-edition2 -ch 20 -json ./agent`. Its expected total is 155,
not a relabeled original21/22 score. No grader renumber/edit, original timeout
repeat, paid search, sign-in or new scripted chapter-task replay is proposed.
Use temporary EN_HOME and the existing stripped provider/application environment,
also remove inherited `FIRECRAWL_` credentials before launching checks. Only
synthetic test-named MCP credentials belong in the local peer processes.

Reuse unchanged HTTP/dispatch and model-control mutation evidence with its
historical limits. New scratch checks target the adapted GUI lock cycle,
repricing old work at the current selection, and silently dropping unknown-price
work. Require intended assertions with passing controls, then remove scratch.
The combined grader's two surviving manager-only mutations do not establish
connection teardown or per-server isolation: retain the real HTTP cancellation,
parked-Recv and selected-manager tests as separate evidence. Compare this stage
against its reconstructed predecessor and both historical steps, separating
production, comments, tests, examples and languages and identifying the small
locking/fixture differences explicitly.

New22 child compression and new23 recall must later add their measured usage to
the status session total exactly once, preserve latest-main usage/forcing and
exclude auxiliary requests from the main lens. Restore the deferred judge
Responses path and child credential/route inheritance at those stages; no judge
or memory request is needed to run current web/status. The enhancement's three
empty-judge sample exchanges remain historical and are not replay prerequisites.
The 23.35-second actual web run and 16.34-second browser status run retain their
original source and limits; no fresh autonomy, service availability or invoice
claim follows from this reconstruction.

The prior ledger for required visible-explanation and main-chat/child-submission
maintenance remains outstanding for its coordinated later integration. Sandbox,
fleet, memory/recall and Codex backend work are not released here. Only this
concrete source plan is submitted: both independent and coordinator approval
must precede Agent edits, followed by independent implementation review before
advancement. No export, renumbering, manuscript or whole-tree merge is authorized.


## Combined web/status implementation submission

October 14. Implemented approved plan `68793cad` after independent source-plan
acceptance `076048cd` and explicit coordinator release. This extends accepted
crossover `078cccbd`; 24 Agent paths change, including the planned GUI lock-order
prerequisite. No other Agent behavior is imported from the cumulative old tree.
This is an initial implementation submission, not acceptance or an export.

The six new HTTP transport/test, model, status and shipped web-skill files match
the selected accepted20 source byte-for-byte. Shared files contain only their
scoped additions. The status fixture matches accepted supported-model correction
`215ae6ac` exactly; CLI and GUI fixtures likewise use Opus-5 rather than 5.5.
The sole additional regression gates real streamed observation and inspects the
blocked public SetModel call before releasing progress. It exercises the actual
actor, GUI and two WebSockets, with an external provider fake. No internal mock
or new test-only production interface was introduced.

The selected `05b43a0` locking correction orders subscriptions and patches using
settingsMu while holding the observation mutex only for snapshots/publication.
The model operation remains session-only, rejects mixed saved-preference patches
and preserves billing, credentials, history and prior usage. HTTP MCP retains
its actual owner chain and job/connection lifetime; agent_status reads the
owning Engine and Agent instead of receiving copied accounting state. No new
price row or auxiliary request API was needed. Accepted credential sanitization
and model-response deadline corrections remain unchanged.

### Actual validation

All commands used temporary EN_HOME and stripped inherited provider/application
and FIRECRAWL credential variables. Raw local outputs are retained under
`/tmp/ensemble-reordered-checks/web-*.txt`; these are fake-boundary checks, not
live service verification. The temporary browser probe and count/mutation JSON
are in the same directory. No paid call or account-profile read occurred.

| Check | Actual result |
| --- | --- |
| Core `go test -short ./...` | PASS; ensemble integration package 13.511s |
| GUI `go test ./...` | PASS, 1.274s |
| Core and GUI `go vet ./...` | PASS |
| Focused HTTP/status/model/CLI race checks | PASS; core packages 1.217s / 1.021s / 1.256s |
| GUI model/settings race checks | PASS, 1.043s |
| Separate ch05/ch06 tests and vet | PASS; tests 0.008s / 0.246s |
| Separate ch07 module compilation | PASS; reports no test files |
| `go run ./cmd/grade-edition2 -ch 20 -json ./agent` | PASS, 155/155 across all 11 rows |
| Actual Chrome Settings model field, two tabs | PASS; captured Opus-5 then Sonnet-5 requests retain prior conversation; GUI exits 0 |
| Formatting and whitespace | PASS; gofmt lists no files; diff check clean |

The combined grader ran once in the coordinator's exclusive slot; completion
was reported and the slot released. It is the unchanged current20 identity,
not a revised original21/22 result. Browser evidence uses actual field input and
change handling, both tabs' selected values and captured provider requests;
settings acknowledgments alone are not the evidence of effective switching.
The temporary probe is `web-model-browser.mjs`, output `web-browser-model.txt`.

Three isolated scratch mutations detect the planned adapted claims:

- Restoring the observation-mutex hold only across SetModel passes the idle
  two-tab control, then parks the actual setter behind gated observation.
  Releasing that gate produces the intended five-second wire-read failure;
  cleanup is blocked by the same actor/GUI cycle and the 12-second test alarm
  terminates it. Both SetModel and ObserveAgent blocked stacks are retained.
  This is bounded evidence of the actual liveness defect, not a clean ordinary
  assertion exit. The correct baseline and final restored race control pass
  (final package 1.026s), including busy rejection, progress, later idle selection
  and shutdown. Raw file: `web-mutant-gui-lock-cycle.txt`.
- Pricing stored buckets with the currently selected model fails
  “selection changed historical counts or cost,” exit1, 0.602s launcher time.
- Ignoring an unpriced bucket fails the required “cost: unknown” observation
  and incorrectly retains $0.000379, exit1, 0.598s launcher time.

All three mutation outcomes and intended predicates are in `web-mutations.json`;
no compilation/setup failure is credited. Scratch trees were removed and the
restored status test passes (0.014s). Unchanged HTTP/dispatch/control mutation
claims retain their sourced historical evidence and explicitly recorded limits.
No unexpected test failure occurred in this bounded implementation stage.

### Comparative scope and remaining gate

Counts use physical / nonblank non-comment / comment-only lines, with `//`
comment-only detection; examples, tests and languages are separate, excluding
dependencies, captures and documentation. Raw counts: `web-line-counts.json`.

| Delta against its own accepted predecessor | Production Go | Go tests |
| --- | --- | --- |
| Historical web step | +230 / +192 / +25 | +405 / +369 / +27 |
| Historical status step | +118 / +95 / +15 | +314 / +283 / +23 |
| Historical combined19→20 | +348 / +287 / +40 | +719 / +652 / +50 |
| Reconstructed crossover→web/status | +361 / +298 / +42 | +834 / +757 / +57 |

Both combined deltas add three production JavaScript lines and one HTML line;
examples and JavaScript tests are unchanged. The extra production 13 physical /
11 active / two comment lines are the narrow GUI ordering prerequisite. The
extra 115 physical / 105 active / seven comment test lines provide its causal
regression. New production comments explain owner access, request lifetime and
ordering; about 12.4% of added nonblank production lines are comments, below the
rough teaching target without padding. Existing useful explanations and exported
contracts are preserved. Full reconstructed production Go is 11,123 physical,
8,844 active and 1,685 comment-only lines.

Supported fixtures establish priced Sonnet-5 work, unchanged historical cost on
selection and honest unknown cost after unpriced work. They do not newly prove
summing two priced Anthropic models. Historical live web/status results, original
failed grades and enhanced20's auxiliary judge exchanges retain their original
identities; none is relabeled as this run. Existing EOF/callback limitations are
neither repaired nor freshly revalidated here. The unchanged model response path
did not warrant repeating its 61-second deadline control.

New22/23 auxiliary accounting, latest-main/forcing/lens separation, child
credential inheritance and judge Responses adaptation remain deferred exactly
as above. Required visible-explanation and main-chat/child-submission corrections
also remain for later coordinated runtime integration. No sandbox, fleet,
memory/recall, Codex backend, prose, grader, manifest, export or tag was changed.
Submit this checkpoint to the independent dependency reviewer and coordinator;
await acceptance before advancing. Required handoff:
`.agents/skills/ensemble-coding/SKILL.md`.


## Web/status acceptance and sandbox source plan

October 14. Independent review `387b79b2316294d61af198dc49195194547916ad`
and the coordinator accept web/status `9af3864363c1d574f66b44585b35e65c488dfec1`,
Agent tree `6cb2eb95e7a7ad5d5fe9afc8ca2d4f5a7ad3e99d`, at initial review,
zero correction rounds. Prepare current21 → proposed19 from that exact accepted
predecessor. This is a source plan only, with no Agent edits, live calls or exports.

Read complete current Chapter21 (894 lines; file revision `dd96a1b0`), original23
student and independent reports, and enhanced21 coder/reviewer reports. Inspected
the complete historical delta, all new sandbox files/tests and the complete
accepted escape/cwd correction. Full mandatory
`.agents/skills/ensemble-coding/SKILL.md`, workflow and carryover remain active
from the latest full reload, without intervening compaction. The chapter is read
for implementation scope; no manuscript authorship or prose review is performed.

### Historical source and selected changes

Historical accepted combined20 has Agent tree
`a801628b1dd631e5ea20ad03ea7d255584728d31` at `edition-2-ch22-r1`.
Accepted sandbox source `64e5d0c17be8db116fa4c1da52cd290c45b4e2ef` has tree
`9dc3e9e36a854ede4cc2cbef22dc389850cedc0c`, independently accepted by `3df1300d`.
Evidence checkpoint `320859a0` leaves that tree unchanged. Export checkpoint
`574fbab5` / `edition-2-ch23-r1` and present frozen `solutions/ch21` have that
same tree. The first-green/live source `fcb98a94` precedes two real fixes and is
not the reconstruction source. This remains maintenance with reviewed-answer
exposure, not a cold student implementation.

The historical combined20→21 delta has 22 paths, 1,124 insertions and 76 removals.
Select 20 paths; defer `internal/memory/memory.go` and `internal/recall/recall.go`
entirely. For shared files apply only the sandbox delta to this predecessor;
do not replace them with cumulative old files. The selected groups are:

- Add `internal/common/sandbox.go`, `internal/sandbox/files.go` and `command.go`:
  required Agent parent, canonical root/candidate including missing ancestors,
  directory containment and actual Go1.25 rooted operations; caller-spelled
  diagnostics; macOS descendant profile, locale allowlist, rooted HOME/temp and
  local Git identity, with unsupported-platform command refusal. No new module
  dependency or alternate-platform implementation is needed.
- Adapt `common/skills.go`, `common/types.go`, `agent/agent.go` and
  `ensemble/ensemble.go`: Agent owns construction-time spec and attached Sandbox;
  composition resolves/validates state paths and installs the child before its
  consumers. Preserve default full access only when the spec is omitted;
  explicitly supplied zero EnableWebSearch means disabled. Root diagnostics and
  API trace scrub registered configured keys under the existing writer lock.
- Add public `ensemble/spec.go`. AgentSpec becomes a defined public type so
  Child and Clamp behavior lives outside common. Child inherits restrictions and
  clears parent log/save paths; Clamp returns narrowed policy independently of
  its widening diagnostic. These operations introduce no spawned children.
- Route existing file tools, commands, media, journal/snapshot operations,
  cache-artifact rotation and job-output creation/rereads through Sandbox in
  `tools/tools.go`, `engine/media.go`, `history/journal.go`, `cachelens/lens.go`,
  `jobs/jobs.go` and `agent/actor.go`. Preserve the existing persistence protocol,
  actual tool/jobs lifetime and accepted accounting/request behavior.
- Enforce safe-mode declaration, registration and dispatch refusal for the five
  accepted execution/supervision tools, while file tools work. Adapt
  `skills/skills.go` to withhold the named web-search skill and therefore its
  startup/connection graph; this is not a general ban on arbitrary custom MCP.
- Adapt `ensemble/config.go`, `main.go` and README for existing environment/CLI
  sandbox inputs, explicit web default and confined trace/state placement.
  Import the eight real-runtime tests in `ensemble/sandbox_test.go`; update the
  three explicit web-enabled specs in `ensemble/mcp_http_test.go`. Other source
  files, GUI controls and public model/auth operations remain the predecessor.

The selected public contract is Agent.Spec/Sandbox plus the documented Sandbox
operations, AgentSpec identity/policy fields and Child/Clamp methods. Runtime
behavior stays in its existing responsible spokes. The trusted host still owns
provider credentials and HTTP; no secret-request tool, policy framework or
fleet manager is introduced.

### Accepted corrections required at this boundary

Use final `64e5d0c`, including these corrections after first green:

1. The xcrun allowance matches only immediate physical-temp `xcrun_db*` files;
   a subpath-plus-basename rule admitted unrelated nested lookalikes. Preserve
   real nested-file read/write refusals alongside successful Python and Git.
2. Pass the caller's cwd spelling into Sandbox. Resolve a relative cwd against
   Workspace only for full access; confined refusal must retain the requested
   relative/symlink string without exposing the canonical root. Preserve the
   legitimate relative-cwd positive control.
3. SetAPILog uses the same lock as writers. Retain the strengthened readable-log
   credential test and direct OpenFile, snapshot/job-output and absolute-in-root
   media controls; these closed historical test gaps without new capabilities.

Enhanced21 subsequently found legacy readable-log leakage and a child-root
symlink escape, not new defects in this accepted source. The selected root
sanitizer and full-candidate canonicalization already fix those paths; retain
positive existing/missing subroots and actual readable-log checks in the current21
grader. Do not import legacy convenience APIs or repeat its broken original probe.
Both sandbox implementation blobs in the complete pre-reconstruction baseline
`a8eec84a` equal accepted21 byte-for-byte, and their intervening history shows
relocation only. No additional later sandbox implementation patch is assumed.
Existing supported-model restrictions, resolved-credential sanitization and the
no-short-total-model-response-timeout correction remain unchanged. The imported
sandbox fixtures use offline synthetic models or existing gpt-4o-mini media;
no unsupported Anthropic fixture or new price row is required.

### Verification and deferred obligations

After plan approval, run formatting, core `go test -short ./...`, GUI tests and
both vets; focused `TestSandbox` under race, plus affected HTTP/web, accounting
and OAuth retry/redaction controls. Check separate ch05/ch06 tests/vet and ch07
compilation because the public Agent/AgentSpec contracts and earlier consumers
are affected. Confirm Darwin/sandbox-exec availability before interpreting
process results; inability to execute Seatbelt would be setup evidence, not a
confinement pass. Use the real runtime, jobs, processes and small external HTTP
peers, with temporary EN_HOME and inherited credential/application variables
stripped as in the preceding stage. Only synthetic credential markers are used.

Coordinate one unchanged current21 grader run with root:
`go run ./cmd/grade-edition2 -ch 21 -json ./agent` (seven rows, 100 points).
This drives the real built CLI/public consumer and preserves allowed file,
Python/Git, network-host and safe-file controls beside escape refusals. No GUI
browser expansion is proposed because no GUI flow changes. Original0/100,
historical48.394s/9.847s live runs and the enhancement's scripted auxiliary judge
requests remain historical; do not repeat or relabel them as this reconstruction.

Reuse unchanged historical targeted-mutation evidence with its limits, including
the initial OpenFile survivor and media test's wrong-boundary failure. Freshly
restore the two accepted pre-submission defects in isolated scratch: broad
nested-cache allowance and prefixed cwd disclosure. Require the intended escape
or caller-path assertion after allowed controls; retain output, remove scratch
and rerun the restored focused tests. The real copied state/media/direct-open
controls run in the new tree even though their unchanged mutation evidence is
reused. Compare selected delta against this predecessor and historical20→21,
separating production, comments, tests/examples and languages; no code-count quota.

At new22, route memory Scan, writes, temporary installation and retired-file
renames through the existing Sandbox. At new23, restore confined memory/handoff
corpus walk/read, skip nonregular entries and disabled web-skill recall, and keep
explicitly selected external skills in their own rooted trusted corpus. These
are the exact two deferred source paths, not permission to import those services
now. Preserve all previously deferred auxiliary usage, latest-main/forcing/lens,
judge Responses/credential-route and visible-explanation/main-chat-child-submit
obligations. Fleet enforcement of derived child specs belongs to the later fleet
stage; Child/Clamp alone does not claim actual spawned-child containment.

The intended result remains macOS confinement with stated system/device/cache
allowances, not arbitrary native-code, mount/hardlink or general security proof.
Submit this concrete plan for independent and coordinator approval before Agent
edits. No live access, export, tag, numbering, manuscript, grader or whole-tree
merge is authorized. Required handoff: `.agents/skills/ensemble-coding/SKILL.md`.


## Sandbox implementation submission

October 14. Implemented plan `7ab4b60a` after independent acceptance `f36e8f74`
and explicit coordinator release. Exactly 20 Agent paths change from accepted
web/status `9af38643`; all five new files match the selected accepted21 source
byte-for-byte, including the eight sandbox tests. Shared files receive only the
reviewed sandbox delta. No memory/recall services or filesystem hooks enter yet.
This is the initial implementation submission, not an accepted export.

The root constructs and attaches the owner-backed Sandbox before its consumers.
Actual tools, media, persistence, cache artifacts and retained job output use
its confined operations. The public AgentSpec methods and CLI inputs preserve
omitted-spec defaults and explicit restrictions. Safe-mode execution gates and
web-skill withholding remain separate from physical confinement. The accepted
xcrun-cache scope, caller-cwd spelling and trace-writer synchronization corrections
are present. No model catalog, request deadline, credential route, GUI behavior
or earlier accounting logic changes.

All validation uses the actual runtime and real macOS processes, with external
HTTP fakes only, stripped credential/application environment and temporary EN_HOME.
No live provider/service or private profile was accessed. Raw outputs remain in
`/tmp/ensemble-reordered-checks/sandbox-*.txt`; counts and mutation receipts are
`sandbox-line-counts.json` and `sandbox-mutations.json` in that directory.

| Check | Actual result |
| --- | --- |
| Core `go test -short ./...` | PASS; integration package 14.007s, launcher 16.815s |
| GUI `go test ./...` | PASS, package 1.044s |
| Core/GUI vet | PASS |
| Sandbox, HTTP/web, status and OAuth focused race | PASS, package 3.460s |
| Separate ch05/ch06 tests and vet | PASS; packages 0.008s / 0.246s |
| Separate ch07 `go build ./...` | PASS; generated local binary removed |
| Unchanged current21 grader with explicit `./agent` | PASS, all seven rows, 100/100, launcher 0.977s |
| Formatting and whitespace | PASS; gofmt lists nothing, diff check clean |

The grader ran once after root explicitly granted the exclusive slot; completion
and release were reported promptly. Current21 numbering is unchanged. The focused
race command first received misspelled `-timeout60s` and exited1 at argument
parsing, before tests. Preserve that setup failure as `sandbox-race.txt`; the
corrected `-timeout 60s` run is `sandbox-race-corrected.txt`. No runtime test failure
is inferred from the rejected flag. The initial three-way port helper also
stopped on a multi-conflict exit count; explicit conflict resolution retained only
Sandbox fields/methods/import and the safe-mode check, without memory/recall or
forced-memory handoff. Those were source adaptation, not a test or review round.

The eight real tests preserve allowed file/edit/search/list and canonical-alias
controls; actual Python/Git and relative cwd; outside file/descendant/write and
network denial with the same reachable host peer; restricted environment and
readable redacted trace; unconditional Child/Clamp plus its independent alarm;
safe-mode file success with tool absence; disabled web peer plus enabled bridge;
state-file/job-output symlinks and absolute-in-root media before outside refusal.
No additional permanent test harness or production interface was needed.

Fresh isolated restored-defect checks both fail for the intended reason:

- Replacing the anchored immediate-cache regex with the old subpath/basename
  allowance passes legitimate command controls, then actually reads the planted
  nested outside marker with exit0. `TestSandboxCommands` fails “kernel escape,”
  test package 0.173s, launcher 1.354s, exit1.
- Prefixing the caller's cwd with the host Workspace before sandbox resolution
  yields two “cwd refusal leaked root” assertions, with the canonical host path
  visible. The rest of the command test completes, including allowed cwd and
  network controls; package 0.248s, launcher 0.908s, exit1.

Both scratch copies were deleted. The correct source's restored command test
passes under race, package 1.250s / launcher 1.835s. These are intended behavioral
failures, not compilation errors or timeouts. Unchanged historical mutations and
their earlier survivor/wrong-boundary limits remain separately attributed; no
claim is made to have repeated the original 27 attempts or certified all escapes.

Physical / active / comment-only counts use the same nonblank/non-`//` method as
prior stages, separating examples, tests and languages and excluding generated
outputs, dependencies and documentation:

| Own-predecessor delta | Production Go | Go tests |
| --- | --- | --- |
| Historical combined20→21 | +620 / +515 / +73 | +408 / +376 / +20 |
| Reconstructed web/status→sandbox | +587 / +486 / +69 | +408 / +376 / +20 |

The 33 physical / 29 active / four comment-line difference is the deferred recall
integration. Deferred memory changes have no net line growth. No example,
JavaScript or HTML code changes. Full reconstructed production Go is 11,710
physical, 9,330 active and 1,754 comment-only lines. Useful comments explain owner
policy, rooted access, missing ancestors, safe diagnostics, cache location,
Git/environment allowances and unconditional narrowing. Added production comments
are about12.4% of nonblank growth; the lower ratio is reported without padding.
The new exported interfaces and members retain their accepted useful contracts.

Original0/100 and the historical live runs keep their original source identities.
The current100/100 is a new fake-boundary/real-process validation, not a paid run.
The accepted platform and system/device/cache allowances remain; no new Linux,
mount/hardlink, arbitrary hostile-native-code or blanket security claim is made.
Child/Clamp still supplies policy, not a fleet. The exact new22/23 filesystem and
auxiliary-accounting obligations in the plan remain outstanding, as do required
later visible-explanation/main-chat-child-submit corrections. No grader, prose,
export, manifest, tag, renumbering or full-runtime replacement changed. Submit
this source for independent implementation review before further scope.
Required handoff: `.agents/skills/ensemble-coding/SKILL.md`.


## Combined GUI reuse/fleet source plan — current22 → new20

October 14. Sandbox source `922caeb4be404c335d209348a9d36aca4f88645e`,
Agent tree `06c7edc9489eff9cf1bf071a8b6cccba303fcc8b`, is independently accepted
at `403721188e48863438e0b38cc0aaaabceac93868`, initial review, zero correction
rounds. Coordinator read and accepted that disposition. This is the predecessor
for the next bounded reconstruction, not a replacement for the full runtime.
No Agent edits or validation runs are part of this source-plan checkpoint.

Reloaded full AGENTS, `.agents/skills/ensemble-coding/SKILL.md`, workflow and
carryover after compaction. Source preparation includes all current Chapter22
(1,004 lines), original24/25 student and comparative records, enhancement22
coder/reviewer records, selected historical production/test deltas, the accepted
Stage B contract and implementation disposition, and the applicable crossover
maintenance source. Reviewed-answer exposure continues; no cold-run claim.

### Selected sources and boundaries

| Source | Use in this reconstruction |
| --- | --- |
| Original23 accepted `64e5d0c17be8db116fa4c1da52cd290c45b4e2ef`, tree `9dc3e9e36a854ede4cc2cbef22dc389850cedc0c` | Historical predecessor only; reconstructed sandbox already carries its scoped corrections. |
| Original24 accepted `0430df4340fffac5234a94aea843f9a7c9d02282`, tree `8555e3cb9ab5200f0a1c52691007a7fc4a1c6797` | Select the 14-path GUI reuse delta from accepted23. Its live source `35df549` remains separate. |
| Original25 accepted `e00d5bcaf1412c50787428ef1fa08bed59b99e11`, tree `30fb05d154b9808eb88622aaff9c1424cd7e8af1` | Select the 23-path fleet delta from accepted24. This is accepted combined current22; corrected/live `891d513` differs only in final README/test diagnostics. |
| Main-chat maintenance `3aad2b7ea4709f0d686ffcb9d6bcf97bc9c8fad4` | Only the five fleet/role paths: root idle exemption, child/grandchild instructions after primary loading, skill wording, and corresponding fleet tests. Speech calibration, GUI MCP observations and optional Codex changes are not dependencies of this stage. |
| Visible-explanation maintenance `df38abe07427feed6fb8ca1b200c78b829b3d2ca` | Capability flag, supported OpenAI rows, dynamic prompt composition, whole-batch/stream-order enforcement, focused tests and affected existing fixtures. Exclude optional Codex test changes and old memory-dependent catalog fields. |
| Accepted structured-submission Stage B `a0acfb5e706cdec285245a935163a3eb14d19afb` | Generic schema admission, construction ownership, raw JSON submission/declaration/result preservation, durable records/readers and focused tests. Independent and coordinator initial acceptance, zero rounds, is recorded in the primary runtime audit/progress; no Stage C source is selected. |
| Accepted submit-state correction `450f7a9599295dafb260c1650b9b5553010e2a54` | The four-line dispatch correction records existing `turn_ended` after the submitted batch drains. It is a prerequisite of Stage B's actual completed/reloaded GUI state, not introduction of conversation selection or a new lifecycle system. |

The two historical deltas total 36 unique paths (README is shared). With generic
Stage B, visible-explanation files and existing-file maintenance, the anticipated
scope is 55 Agent paths. In addition, only necessary `go.mod`/`go.sum` updates in
the three existing separate ch05/ch06/ch07 example modules are permitted for the
new pinned validator's transitive dependency; no example feature expansion.
Record actual final paths/counts rather than forcing this estimate if a dependency
needs correction. No entire accepted cumulative answer tree is copied.

### GUI reuse and fleet construction

Preserve the already separate optional GUI module and public owner vocabulary.
Select `gui/assets.go`, `relay.go`, component extraction into `artifact-scroll.js`
and `renderers.js`, plus the narrow existing GUI/MCP/command changes and external
consumer/browser tests. Embedded assets are private with an `fs.FS` accessor;
explicit disk override remains available. The component needs no application
shell, WebSocket or speech object. Public TCP JSONL MCP relay retains per-request
caller IDs across clients, handles no-browser refusal, cancels on client/browser/
server shutdown, and consumes its late private IDs. No new core GUI dependency.
Keep the accepted settings lock-order fix and actual model-switch behavior.

Introduce the seven original terminal tools and their owning Fleet, not next
chapter's management panel. Public APIs add AgentFactory, SetAgentFactory,
SetUserSink/UserID/DeliverUser, CreateChild/NewChildAgent, Agent.ParentAgent/Fleet,
FleetMessage, ChildStatus and SubAgentManager. Common contains their contracts;
Fleet contains child/question behavior; Agent's existing actor and Jobs own work.
The CLI supplies its actual recursive factory and human sink, and reads `/reply`
while inference/tool waits continue. GUI remains a single-Agent consumer here.

Trace construction rather than trusting IDs: Spawn derives the parent's spec,
uses `children/<id>` for state/log/save, retains the shared task workspace and
calls Ensemble.CreateChild with the actual spawning Agent. The CLI factory uses
parent.Config (including provider/credential/billing choices), changes only an
explicit requested model, and calls NewChildAgent. NewChildAgent validates the
root relationship and applies the accepted unconditional Clamp before newAgent
constructs the child. The child stores both Ensemble and spawning-Agent pointers;
Fleet publishes it only after valid construction/load/save, before starting work.
The new child's Sandbox reads its own clamped spec, so actual spawned file/tool
operations use inherited restrictions. Successful confined child work and rejected
widening/escape will be observed, not inferred merely from the earlier Child helper.

Retain receiver-owned question IDs, including same-name parent/child controls;
questions yield fan-in without becoming submissions. Omitted reply ID is accepted
only for an unambiguous recipient. Incoming input wakes the current wait, leaves
the job alive and permits later collection; it must not leave a stale wake flag.
Recursive shutdown saves children and frees parent-local IDs for restored tasks.
No-factory earlier consumers keep their single-Agent behavior. Delegated children
retain exactly three corrective wakes and truthful failed status; failed requests
are not retried. The accepted later main-chat rule overrides the historical root
wake expectation: a human-facing root completes in ordinary chat. Child instructions
are appended after primary loading, exactly once for grandchildren, using authored
BaseSystemPrompt rather than storing current-model policy in the inherited role.

### Structured submission at its introduction

Port Stage B's immutable Fleet-owned compiled contract and the exact pinned
`github.com/santhosh-tekuri/jsonschema/v6 v6.0.3` validator plus required x/text
metadata. Do not copy unrelated Noise/Gorilla dependencies from later core source.
AgentSpec gains SubmitSchema; SpawnWithSchema accepts the optional whole-argument
object schema, and ordinary Spawn/Child derivation clears inherited contracts.
Construction validates/copies the schema before filesystem/auth/skill/factory
side effects. CreateChild rejects nil/foreign parents and bad factory ownership;
refusing an unrelated live Agent never grants ownership to shut it down. A validly
owned child with a substituted contract is shut down and refused.

Keep the accepted bounded Draft2020-12 subset and admission limits unchanged:
64 KiB schema, 1 MiB arguments, depth64, 2,048/65,536 nodes, 128-byte numbers,
exponent magnitude1,000, canonical byte checks and combined64 MiB size product.
Unknown keywords, references, combinators and external resolution are refused.
These are Stage B's reviewed limits, not a newly invented validator framework or
a wall-clock guarantee. Schema semantics stay in the pinned dependency.

Public additions are Submission, SpawnWithSchema, SubmitJSON, SubmitSchema,
SubmissionData and the compatible Submission string reader. Submit(string) wraps
legacy result text through the same validator; custom results retain the entire
raw object. Declarations/spec/status/results return independent bytes. Validate
before durable write or completion notification. Retain old exact result-string
JSONL rows and custom version1 schema/payload envelopes in confined `submit.md`.
Historical rows validate against their stored contract, not a new task's contract;
malformed/latest unsupported rows fail, and retrieval never completes a new task.
Ask/TurnEnded stays text (compact JSON for custom results); typed retrieval/status/
join keep raw JSON. GUI tool output stays text containing JSON, with live and
fresh-construction/load replay checking its actual nested value.

The Stage B source reader calls later Fleet.Find. Adapt that one call to a small
private lookup under the existing child-map mutex, with identical owned-child/
unknown-target semantics. Do not add public management Find/Children, ask_agent,
get_submission, monitoring/cursors, named persistent peers, EnableFleet convenience
or conversation selection solely to compile its tests. Adapt tests to the original
factory's retained child references at synchronized completion boundaries and
explicit SetAgentFactory/SetUserSink. Defer the get_submission tool assertion to
new21. The GUI test constructs the single selected Agent directly and uses the
existing inbox helper; it does not need selection machinery. Drop only the absent
Recall.Disabled fixture settings. Test payload keys named `memory` remain inert
schema examples; no Memory service, worker or memory state is introduced.

Port the small accepted `turn_ended` correction before Stage B's text-result
adaptation. Test real completion/reload state and retain a held-response running
control, so simply labeling all states idle cannot pass. This corrects the state
fact at submit's introduction; it does not make submit a worker-quiescence signal.
No Terminal/Quiesce/worker outcome or automatic child-cleanup APIs enter now.

### Visible explanations and fixture adaptations

Carry the previously explicit df38abe0 obligation now: six existing supported
OpenAI model rows opt in; synthetic course and unknown rows stay disabled. Engine
renders one model-policy instruction after the authored role. Actor checks ordinary
nonblank text before the first tool in the same response and actual stream order;
private reasoning/summary is insufficient. Reject the whole unexplained batch with
paired error results, no job or execution side effect, within the existing round
budget, honoring after-tool interruption. This checks presence/order, not quality
or a word-count threshold. Preserve opaque replay and existing request accounting.

Retain the full focused prompt/actor tests with absent recall fixture settings
removed. Adapt the captured Responses test exactly as accepted maintenance did:
unchanged historical bytes execute under the synthetic capability and are refused
under the flagged real model. The streaming test keeps authentic tool-first SSE
and asserts the paired refusal; the empty-primary cache test uses the synthetic
row. GUI budget fixtures add explanations before the same calls. Do not rewrite
historical captures or reintroduce unsupported Anthropic 5.5 rows. The inherited
no-total-response-deadline and resolved-credential sanitization remain intact.

The historical CLI fleet test's idle-root branch is known to fail after the
accepted main-chat change. Correct that obsolete expectation here: observe one
ordinary root completion without escalation; preserve its separate real human
question/reply check and delegated-child three-wake test. This is a fixture
maintenance adaptation, not a waiver of unreachable human delivery or failed
child behavior. Preserve the original failing evidence and its source identity.

### Focused verification and remaining ledger

After both plan approvals, run formatting/diff checks, core `go test -short ./...`
and vet, GUI tests/vet including its real outside-module race/Chrome consumer,
and focused core race tests for fleet, schema, visible explanations, affected
Responses/cache/streaming and actual child Sandbox inheritance. Run separate
ch05/ch06 tests/vet and ch07 build; dependency updates affect their real external
consumer boundary. GUI's extracted components and existing affected speech/
Responses browser tests use their required harness, not a broad environment-free
glob. No unrelated whole-book sweep or new paid demonstration is proposed.

Run one unchanged current22 grader with explicit staged source only after root
releases the serialized slot: `go run ./cmd/grade-edition2 -ch 22 -json ./agent`.
Its 14 rows/200 points retain their original identity. Source review separately
checks both executable composition paths, owner links, transitive headless imports
with positive GUI controls, credentials and actual child narrowing. Original24/25
0/100 and historical9.649s/48.198s live runs remain historical, including received
but unfulfilled steering and earlier setup/sensitivity failures.

Use real Agents, actors, histories, validators, jobs, processes and GUI; only
external peers are small fakes. Strip credential/application environment and use
temporary EN_HOME as in prior stages. Preserve failures before diagnosis; no real
provider/private profile access. Reuse verified unchanged historical mutations,
Stage B's 21-defect/24-execution evidence and df38abe0's gate/order/interrupt
controls with their original limits. Fresh focused mutants cover the adaptation
boundaries: remove NewChildAgent Clamp (widening refusal beside allowed child work), restore
root correction (ordinary-root completion), preserve parent schema in ordinary
Spawn (grandchild/default control), substitute private lookup's first child for
an unknown target (owned-retrieval control), convert custom submitted output to
legacy text (real GUI nested-value/replay), and omit submit's turn-ended fact
(completed/reloaded state beside held-running control). Confirm each intended
assertion after a passing baseline, retain output and remove scratch promptly;
restore and rerun affected checks. These are focused sensitivity checks, not a
new general mutation/evidence system. Additional defects require actual evidence,
not an unbounded matrix.

Compare this chapter delta against its accepted sandbox predecessor and the
historical combined21→22 delta, separating generic Stage B/maintenance growth,
production/tests/examples/languages and physical/active/comment-only lines.
Explain useful exported contracts and ordering/lifetime comments; no line quota.
Independent implementation review follows a coherent tested source checkpoint.

Outstanding new21 obligations now explicitly include management tools (including
typed get_submission adaptation), public lookup/child traversal, GUI fleet and
conversation selection; later worker lifetime cleanup remains new22. New22 also
owes memory filesystem hooks, structured memory-child policy, reliable completion/
save/cleanup boundaries, actual child-model/tier usage exactly once, latest-main
forcing isolation, main-lens exclusion and memory compaction marker fallback.
New23 owes confined memory/handoff recall, nonregular/disabled-web exclusions,
trusted explicitly selected external skills, judge Responses/credential routing
and auxiliary accounting with latest-main/lens isolation. Generic schema support
does not satisfy those obligations. Optional24 Codex transport and remaining
crossover GUI/speech maintenance stay separate and must not disappear at eventual
full-runtime integration. No early memory/recall, full-runtime merge, export,
manifest/tag, renumbering, grader/manuscript edit or publication is authorized.
Required handoff: `.agents/skills/ensemble-coding/SKILL.md`.

### New20 implementation submission — October 14

Implemented the independently accepted plan `53958bcc` (review `ab511273`) from
accepted sandbox `922caeb4`, with the scoped corrections below. This is reviewed
source reconstruction, not an independent student attempt. Full mandatory skill,
AGENTS, workflow and carryover were reloaded after compaction before continuing.
Current22 remains the grader identity; no frozen export or chapter number changed.
The source checkpoint containing this entry is the submitted revision, not an
acceptance or permission to replace the full application.

The actual change is 60 Agent paths: 54 implementation/test/documentation paths
plus six example-module dependency files. The plan's 55-path estimate included
`internal/tools/tools.go`; the final identity correction restores that file exactly
to this stage's predecessor, so it is not a changed path. The core/GUI/examples
add only the pinned validator and required x/text metadata. No Noise, core
WebSocket dependency, memory/recall, management, named peers or optional Codex
transport was imported. Existing supported models, credential sanitization,
main-only lens accounting and no total response deadline remain intact.

**Corrections and source-boundary findings.**

- The proposed df38abe0 after-tool branch and its one
  `TestVisibleReasoningRejectionHonorsAfterToolInterrupt` test require actual
  `AfterTool` state/API first introduced by original26/new21 source
  `55cbef230e675ed0d42547b7ad21e53c5b8a554f`. Root and reviewer approved deferring
  only those four production lines/test to new21; review `3257a817` records this
  missed plan dependency. Immediate interruption, explanation ordering, whole-batch
  paired refusals and round-limit coverage remain here. **Restore the branch and
  test with AfterTool at new21**, not by silently adding the future API now.
- The first actual core run failed in CLI/fleet/provider submission and recursive
  completion: dispatch appended successful payload-free `turn_ended`, but the
  earlier validator still required one payload. The selected `450f7a9` dispatch
  prerequisite was incomplete without that same commit's exact validation guard
  and explanatory comment. Root/reviewer approved that guard; all other matching
  payload/count refusal checks remain. Review `924ba4fe` records the plan-review
  omission and actual failure. Corrected core passes; this was a real integration
  failure, not merely compilation/setup.
- The absent `claude-ch16-course` known-model assertion belongs to the original
  memory delta `ccd63509`; review `924ba4fe` accepts deferring that row assertion
  to new22. All present supported/synthetic/unknown capability checks remain.
- The new child Sandbox test initially looked for the tool name on a returned
  fact, while this runtime retains it on the called fact. The final check finds
  the actual outside path on `tool_called`, then correlates its subsequent return
  by CallID (consuming that pending identity because IDs can repeat by response).
  It verifies an actual error alongside a successful allowed child read, actual
  inherited owner/policy, safe-mode dispatch refusal and attempted widening.
- The first GUI run retains the same pre-correction completion validation error
  and a separate 120-second cleanup timeout in the new held-provider control.
  The fake waited only for request cancellation. Reading its request and supplying
  an explicit cleanup release fixes the peer lifetime; completed/reloaded and
  genuinely held-running state now both pass. No production timeout was added.

**Tool identity correction, also owed by later full-runtime integration.**
The unchanged browser speech test failed because the selected original25
`allowedTool` gate hid the blind client's MCP `submit` solely by name. An exact
accepted922caeb4 Agent archive passes that same test; this is a new20 cumulative
regression, distinct from ordinary root chat completion. Stage B's name-only
schema replacement would also misidentify external/custom `submit`. Root and
independent `compaction_runtime_audit` approved the following bounded correction:

1. `tools.Builtins` contains only its non-fleet definitions. The composition root
   adds FleetTools only when enabled, binding the immutable contract to the actual
   compiled submit definition at installation. Even enabled/no-file-builtins
   construction installs that real fleet handler.
2. Agent declarations uniformly copy their stored schemas. Tool visibility no
   longer classifies fleet ownership by name; unregistered fleet calls are unknown.
   SafeMode and compiled-over-MCP shadow protection remain unchanged.
3. `TestSubmissionToolIdentity` exercises real MCP discovery and invocation,
   application registration, enabled compiled precedence with/without builtins,
   disabled registration/dispatch and a factory enabled after custom/MCP Agent
   construction. The late-factory control is why a FleetEnabled-only declaration
   check was rejected. No new origin API or metadata framework was necessary.

The narrow implementation review `e9fd0676` accepts exact SHA-256 files:
`ensemble/ensemble.go` 933e134218b30459a42d80552e008d99556a24535d2288a6eed171b1b8d8e383;
`internal/agent/agent.go` 9e306096c58ebf11b5c3b0748a8ed4d1c159c5d2585627765db3f2b3e2f54e94;
`internal/tools/tools.go` 3c66f33f2cc95a5aac52d5d92d635e267f2f3f6400ac163f0a74dfe324d9ae44;
`ensemble/submission_test.go` 3b2b60a0fc9e5a23afaeae97c15b711ce19556da547cf13ae6fec66900ba7a8f.
This accepts that correction only; full new20 acceptance remains independent.
Carry this three-production-region correction into the full runtime through its
own reviewed integration, preserving later memory worker filtering. Never merge
this earlier cumulative Agent tree wholesale.

**Validation.** All runs strip provider/application environment, use temporary
EN_HOME and local fake peers. No paid run or private profile was accessed. Raw
outputs are `/tmp/ensemble-reordered-checks/fleet-*.txt`; the prior stage's evidence
and all historical captures are unchanged. Durations below are launcher wall time.

| Check | Actual result |
| --- | --- |
| Initial core `go test -short -count=1 -timeout 120s ./...` | FAIL 16.838s, completion validation and fixture findings retained in `fleet-core.txt`. |
| Corrected core, then final core after identity correction | PASS 16.133s, then PASS 16.254s (`fleet-core-final.txt`). |
| Initial GUI, corrected GUI, final GUI | FAIL 120.401s; PASS 4.514s; final PASS 4.739s. Includes real outside-module race/Chrome component consumer and transitive headless/GUI positive controls. |
| Core/GUI vet after composition correction | PASS 0.225s / 0.142s. |
| Final focused core race: Fleet, Structured, Submission, Visible, Responses, Cache, Streaming | PASS 4.711s; ensemble3.874s and fleet1.081s. |
| Restored actual CLI human-reply/root completion | PASS 0.487s. |
| Final structured GUI completed/reloaded plus held-running race | PASS 1.413s (package1.065s). |
| Separate ch05/ch06 tests and vet | PASS; final affected tests 0.309s / 0.569s. |
| Separate ch07 build | PASS 0.330s; temporary binary removed. |
| Browser Responses incremental speech | PASS 1.382s. |
| Browser speech on candidate / exact922caeb4 control / correction | FAIL 3.154s / PASS 3.535s / PASS 3.856s. |
| Unchanged current22 explicit-source grader | Actual PASS200/200, 14 rows, exit0, 1.244s; serialized slot released immediately. |
| Formatting, diff check and forbidden future-feature/model scan | PASS; no unformatted Go or unsupported5.5 rows. |

The actual grader ran **before** the identity correction, on the166-file source
capture `fleet-grade-source.sha256`, aggregate SHA-256
`fa7009f5c5c8c2f4af3d170911fdda97118b1841b805383fddc1d85841300681`.
Do not relabel it a final-source grader run. Subsequent production changes are
exactly the three identity regions above; tests gained the identity regression
and one Sandbox field comment, and core go.mod formatting was normalized. Staged whitespace checking also removed
one trailing blank line from the newly ported renderer (unstaged checking alone
had not inspected that untracked file).
Final affected real-runtime checks cover those changes. The final separately
revised Stage D grader on the exact new20 export remains mandatory at Stage E;
this old grader result does not discharge that gate. Original24/25 scores remain
0/100, and their9.649s/48.198s paid observations remain historical only.

**Targeted sensitivity.** All six planned defects failed their intended checks:
removing actual child Clamp accepted attempted widening; restoring root correction
produced four requests and escalation; preserving a parent's schema prevented the
ordinary grandchild from completing under its default contract; resolving an
unknown ID to an owned child violated the required unknown-target refusal;
converting custom output to legacy text lost nested GUI result values; omitting
turn_ended left both current and reloaded state input_pending while the running
control remained valid. Raw `fleet-mutant-{clamp,root-idle,schema-inherit,ownership,
custom-text,turn-ended}.txt` retain the exact errors. The ownership mutant reads
the test's deliberately malformed owned row and therefore fails the error-boundary
assertion; it is not a new demonstration of foreign-file exfiltration.

Two additional correction mutants fail: restoring the submit name gate removes
MCP declarations/refuses custom admission; restoring name-only schema substitution
replaces the real question schema with the unrelated memory-key object, including
late-factory cases. The first schema-mutant setup matched two source occurrences
and stopped before executing a test; the narrowed Declarations anchor produced
the intended failure. That setup error is not counted as a kill. Eight executed
behavioral defects are detected, with coupled failures and the setup attempt
kept separate. Scratch copies/binaries are removed; final affected race/CLI/GUI
checks pass after restoration. Historical GUI/fleet, Stage B and explanation
mutation evidence is reused with its original source limits, not claimed rerun.

**Comparative review.** `fleet-line-counts.json` counts physical / active /
comment-only lines; active excludes blanks and lines beginning `//`. Go examples
include their tests; Go tests elsewhere include the outside GUI consumer. No
node_modules/generated evidence or Markdown enters these language counts.

| Category | New20 delta from922caeb4 | Historical combined21→22 delta |
| --- | ---: | ---: |
| Go production | 1,579 / 1,270 / 226 | 1,022 / 836 / 132 |
| Go tests | 2,781 / 2,597 / 123 | 1,087 / 1,015 / 50 |
| JS production, including moved renderer code | 7 / 4 / 3 | 8 / 4 / 3 |
| JS tests/harnesses | 38 / 35 / 2 | 38 / 35 / 2 |
| Go examples and HTML | 0 / 0 / 0 | 0 / 0 / 0 |

The additional434 active production lines over the historical delta implement
reviewed Stage B contracts/durable values, visible-explanation policy and bounded
maintenance, rather than later memory/management. The additional1,582 active test
lines cover those behaviors and actual earlier-stage adaptations. New private
lookup reuses the owned child map; no public Find or parallel registry was added.
Code comments retain parent ownership, construction-before-side-effect ordering,
wait-versus-job lifetime, durable row validation and schema/caller-copy reasons.
Production delta comment share is15.1% (historical13.6%); cumulative production is
13,289 physical /10,600 active /1,980 comment-only, about15.7% comments among
nonblank lines. This remains below the approximate20% teaching target; explanation
quality and exported contract documentation matter more than padding the ratio.
The public API additions and imported source identities are those in the plan.

No new21/22/23 obligations are discharged by this stage. In particular, restore
AfterTool policy/test at new21; restore the memory-only course row assertion with
its real model row at new22. All previously listed management, memory/recall,
auxiliary accounting, completion/save/cleanup and optional Codex/crossover work
remains explicit. Submit this coherent source for full independent implementation
review, then await coordinator release. Required handoff:
`.agents/skills/ensemble-coding/SKILL.md`.


### New20 independent review correction round 1

Initial full submission `f3009a05eb5875d029028c07c047e63adf33579f`, Agent tree
`a709f71f0fb57fcfd541738d8bb578f172d06fd5`, received one documentation finding:
the newly ported README still instructed the root to submit and described
corrective wakes without limiting them to delegated children. The independent
dependency reviewer found no production or test issue after inspecting the full
source and retained evidence. Round 1 changes only that README paragraph to
match the implemented root-chat/delegated-submit distinction, plus this note.
No tests were rerun for this documentation-only correction, as the reviewer
requested. Earlier runtime checks and the pre-identity current22 grader retain
their exact original source provenance; no score is relabeled.

Updated166-file source capture: `/tmp/ensemble-reordered-checks/fleet-round1-source.sha256`,
aggregate SHA-256 `b218df3d36173b91ea9163afdc9f02f48a8e0e57427850130e5fe7754be740e8`.
The previous final capture and f3009a05 remain intact. This is a review submission,
not an accepted export or renumbering. Mandatory handoff skill remains
`.agents/skills/ensemble-coding/SKILL.md`.


## Fleet management and console source plan — current23 → new21

October 14. Extend accepted `b4708839e8bb1a523069be02b9188145ae50dac4`, Agent
`39449051049e6d7325b4c368e627739f97eab28f`. Independent full new20 review
`2a9e7ed983134e604cfc27009a7f6e569e3cef30` accepts after one README-only
correction round; coordinator also accepted. The earlier f3009a05 submission,
166-file captures and pre-identity200/200 grader retain their provenance.
Separately, primary full-runtime carry-forward62c1e781 was accepted by root after
source and all eight actual output checks. That source is not a cumulative base
for this branch. This checkpoint edits only this plan; no Agent edits or new
validation runs are included. It remains maintenance with reviewed-answer exposure.

Full governing reload completed in this active context: AGENTS, mandatory
`.agents/skills/ensemble-coding/SKILL.md`, workflow, carryover, voice and chapter
procedure. Preparation reads include all current23 (583 lines), original26
(466), complete ch26 student/source and comparative records, enhancement23 review,
and the entire selected historical source delta (recovering the truncated core
range separately). No historical paid run is newly reproduced.

### Exact source slices

Historical guide: accepted original25 `e00d5bcaf1412c50787428ef1fa08bed59b99e11`,
main tree `30fb05d154b9808eb88622aaff9c1424cd7e8af1`, to accepted original26
`450f7a9599295dafb260c1650b9b5553010e2a54`, main tree
`27d7fcf9b8b31275068fa705d6969cfe3eafedc1`. Their31-path delta supplies the
management/GUI introduction, rather than copying cumulative old23. Its stages:

| Source | Selected behavior |
| --- | --- |
| `55cbef230e675ed0d42547b7ad21e53c5b8a554f` | Activity/progress observations, eight owned management tools, finalized cursors, distinct graceful/kill lifetimes, actual AfterTool inbound and actor/report state. |
| `33c4e3315b2210bff946aa9d94bb27f7dda0c340` | Public Start/EnableFleet, named registry and real executable composition; corrected per-sender message revisions for selected-set waits. |
| `974f67be89272ea60fa17b460cefde0e478a137b` | Per-Agent GUI views/per-connection selection, recursive tree and visible name/model/skill create/delete; atomic child-admission closure at deletion. |
| `9e2b72aeaddc85cbe1afadec41909b4d66ad3b34` | Complete replay batch admission or connection teardown, real slow-socket regression, two-tab browser isolation and additional monitoring controls. |
| `450f7a9599295dafb260c1650b9b5553010e2a54` | Final accepted comments and completed/restarted/pending selection tests. Its dispatch turn_ended and validator corrections already exist in new20 and must not be duplicated or reverted. |
| `a0acfb5e706cdec285245a935163a3eb14d19afb` | Only remaining generic Stage B integration: typed get_submission tool, no duplicate old string-only Submission in management.go, existing typed status/output preservation and previously deferred real get_submission assertion. |
| `df38abe07427feed6fb8ca1b200c78b829b3d2ca` | Restore exactly visible_reasoning.go's afterTool interruption branch (historical52–55) and TestVisibleReasoningRejectionHonorsAfterToolInterrupt (326–368) with its real newly introduced API. |

Preserve already integrated3aad2b7e root-chat/delegated-submit policy and new20's
reviewed composition-bound tool identity. In particular, do not import the
historical agent.go name-only fleet gate, string-only Submit/Submission storage,
or default schema substitution in Declarations. FleetTools gains the eight
management definitions; composition remains the single point selecting actual
compiled fleet handlers and binding submit schema. No tools.go rewrite is needed.

### Bounded files and adaptations

Of the31 historical paths, omit `internal/history/history.go` (the necessary
zero-payload guard is already present) and `gui/entry_test.go` (its old final
submit batch existed to satisfy the now-superseded root-submit policy). Preserve
the current budget test's18 real shell batches, ordinary root completion and
visible-explanation fixture. Expected scope is33 Agent paths: the remaining29
historical paths plus `internal/fleet/submission.go`,
`ensemble/submission_test.go`, `internal/agent/visible_reasoning.go` and
`ensemble/visible_reasoning_test.go`. A newly demonstrated prerequisite outside
these paths returns for narrow review rather than importing more current runtime.

- Core activity/contracts: `internal/agent/{activity,actor,agent,dispatch}.go`,
  `internal/common/{activity,actor,fleet,types}.go`, `ensemble/fleet.go`.
  Port only the activity/AfterTool/management hunks onto the earlier actor;
  retain existing visible-explanation checks, typed turn completion, supported
  models, credential redaction and no-total-response-deadline correction.
- Fleet behavior: `internal/fleet/{fleet,management,submission}.go` and
  `internal/tools/{fleet,management}.go`. The new public owned `Find` replaces
  the new20 private findChild helper; SubmissionData calls that same lookup.
  Keep the existing immutable contract, raw payload copies and durable typed
  record reader. Do not reinstall the historical string-only writer/reader.
  Merge Activity/Usage into ChildStatus while preserving Result/Payload and
  caller-owned bytes. End records actual failure; human interruption does not
  kill the child. Preserve the root exemption and three delegated corrective wakes.
- Composition: `ensemble/{ensemble,registry}.go`, `main.go`, `gui/cmd/main.go`.
  Start installs recursive construction/human routing before restore. Registry
  state stays with Ensemble; no callback bag or global model-visible map.
  Named peers use their own data/save paths; task children remain ephemeral and
  continue through existing NewChildAgent and Sandbox.Child/Clamp. Retain
  schema validation and actual-parent checks. No Stage C memory-worker factory,
  purpose, dismissal or lifetime API is selected.
- GUI: `gui/{gui,settings,tree}.go`, `gui/web/gui/{app.js,index.html,style.css}`.
  Move current settings/history/partials into views; apply only receiver/owner
  changes to current settings.go, with no old Memory fields or setters. Keep
  existing settingsMu ordering and actor acknowledgment outside the observer
  mutex. Coalesce tree traversal outside synchronous actor observation. Keep
  batch replay admission, selection publication order and truthful retained state.
- Tests/docs: `ensemble/{management_test,registry_test,submission_test}.go`,
  `gui/{selection_test.go,fleet.test.mjs}`, the visible-reasoning test above,
  `README.md`, and `skills/ensemble/SKILL.md`. Add management names while retaining
  role-correct skill/README wording. Remove exactly four Recall.Disabled fixture
  settings (two registry, two selection); no recall service is needed. Real
  provider fakes continue to supply valid visible explanations where required.
  Registry/browser ordinary root and named-peer replies use chat completion;
  delegated children still submit. The explicit named-peer submit in the retained
  completion-state regression tests permitted tool behavior, not a requirement
  that named peers submit. Preserve the historical test's completed/reloaded and
  truly in-flight cases. The budget fixture needs no submit adaptation.

### Public contracts and ownership

Add `Activity`, `ActivityChanged`, `FleetChanged`, actor Activity/ProgressTime,
`Interrupt.AfterTool`, ChildStatus.Activity/Usage, `ChildOutput` and the new
manager methods Find/Children/WaitChange/Output/GracefulShutdown/End/PrepareDelete.
Existing typed SubmissionData and SubmitSchema contracts remain unchanged.
Add Ensemble Start/EnableFleet/CreateAgent/DeleteAgent/Close, NamedAgent,
AgentNode/AgentKey/TreeSnapshot/FindAgent. Model-callable tools resolve only
through the caller's direct children, including before durable-result access.
Trusted browser selection may traverse actual ownership paths; identical local
names in different branches remain distinct.

Actor owns turn/activity state; activityMu protects cross-actor snapshots and
progress time. Fleet mutex protects children, questions, per-sender progress,
submission and admission. Ensemble registryMu protects named identity and startup;
GUI owns views and connections, with settingsMu ordering user patches. Tree
reads occur outside the synchronous observer callback. No new general event,
registry or origin framework is proposed. Graceful shutdown waits current/queued
turns; kill/application teardown cancels lifetime. Both interrupt modes preserve
independent jobs, their eventual real results and the Agent's next task.

Stage B now reaches all newly introduced outputs: get_submission returns
`{agent_id,payload:<object>}` for custom submissions and `{agent_id,result:<text>}`
for default ones. Status, change-wait and output status retain custom payloads;
Ask and ordinary tool output stay text. Restore the deferred get_submission
assertion and add small checks of actual status/change/output boundaries to that
existing structured regression. The new management source uses the already-owned
child map; remove the duplicate private helper instead of maintaining two lookups.

### Focused verification after release

Use temporary EN_HOME, stripped provider/application environment and local fake
peers throughout. No live/profile access, historical evidence rewrite or grader
code change is authorized. Check actual CLI/GUI construction and all new imports;
use the existing public outside-GUI consumer, not a new harness framework.

1. Run gofmt, core `go test -short -count=1 -timeout 120s ./...`, core vet;
   separate GUI tests/vet and focused race for Fleet/Management/Named/Structured/
   Submission/Visible, GUI selection/structured replay/settings locking. Run
   affected separate ch05/ch06 tests/vet and ch07 build after public interfaces
   change; remove the generated build output.
2. Run the built real GUI fleet browser test (nested repeated names, both tabs,
   create/delete with active-child refusal, continuing activity, selection and
   subsequent output). Also rerun speech/Responses browser checks because the
   shared selection/router changes. Capture source and exact actual results.
   Preserve existing role controls and add a named-peer ordinary-chat check in
   registry/browser fixtures so startup cannot silently reinstate root correction.
3. Keep real selected/unselected-question and no-progress controls, finalized
   cursor/usage/durable-result checks, graceful versus kill, both interrupt modes
   and a later question to the same child. Completed selection/restart must stay
   idle while a held provider request stays in_flight; unchanged structured GUI
   replay tests exercise the new per-Agent view path too.
4. Focused scratch mutations: remove the restored AfterTool rejection branch;
   stringify custom get_submission output; restore manager-wide question
   attribution; return immediately from a change wait while its selected child
   is held; restore blocking replay admission and require the bounded real-socket
   shutdown failure. Pair each with its working baseline and intended assertion.
   The no-progress check needs actual held work plus bounded observation, not
   waiting activity alone. Reuse original26's other accepted mutation evidence
   and new20's unchanged schema/ownership/role/identity evidence, preserving
   their historical limits. Remove scratch after each mutation.
5. Request exclusive grader slot from root/structured coder before one unchanged
   current23 run with explicit `./agent` source; retain raw output and source
   capture. Original26's0/100 and historical revised100/70 stay historical.
   A current23 baseline result does not replace the final revised Stage D grader
   on the exact new21 export required at Stage E.
6. Compare actual chapter deltas against b4708839 and historical e00d5bc→450f7a9,
   separating production/comments/tests/examples/languages and moved view code.
   Inspect useful exported comments and ownership rather than padding counts.
   Check no memory/recall, unsupported5.5 model or future Gateway/Codex feature
   entered the staged tree. Checkpoint exact owned source/note, then independent
   implementation review before advancing.

### Remaining stage obligations

This stage discharges management/GUI, typed management output and the deferred
AfterTool restoration only after implementation/review. New22 still introduces
all memory bands and restricted background compaction, exact source/policy
acceptance, reliable completion/save/automatic cleanup, crash reconciliation,
pressure blocking and once-only auxiliary accounting. Its memory-kind cache
boundary, filesystem hooks and claude-ch16-course model/assertion remain deferred.
New23 still introduces recall corpus/judge and its separate auxiliary accounting,
filesystem and lens exclusion. No generic schema example using a memory key is
permission to create a memory service. Optional24, crossover25 and Gateway26
remain separate later integration obligations. No export, manifest, tag, chapter
renumbering or replacement of the full runtime is authorized by this plan.

Submit this concrete plan for independent dependency review and coordinator
approval before Agent edits. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.


## New21 management implementation submission

October 14. Plan7a4ba6fc received independent approval c80755f7 and coordinator
release. Implementation extends `1b7bb6521acc11aa43ac1d63c0ea3a8e9afe72b6`, which
carries the separately reviewed micro_handoff reasoning correction on accepted
new20. Its four affected Agent paths remain byte-identical. Full mandatory skill
and governing documents were reloaded after compaction. This is maintenance
assembly from the reviewed sources above, with answer exposure, not a cold run.

The33 planned Agent paths now carry accepted historical management/registry/GUI
behavior, the Stage B typed management boundaries, and the deferred AfterTool
rejection handling. The owned public Find replaces findChild; immutable compiled
submit schema/identity, raw payload copies, ordinary root/named-peer chat,
delegated corrective wakes and typed turn completion remain intact. Registry and
browser fixtures now demonstrate ordinary named/root chat while actual delegated
children submit. Exactly four obsolete Recall.Disabled fixture fields were
omitted. Existing settings lock order, supported model rows, resolved credential
redaction and unlimited total response duration remain unchanged. No memory,
recall, Gateway or Codex-route implementation entered this stage.

The coordinator narrowed the original plan's mutation step before implementation:
reuse accepted evidence for unchanged historical behavior; mutate only the newly
adapted AfterTool and typed-result claims. Consequently no new message-attribution,
change-wait or replay-admission mutant was run. Their existing real held-request,
selected/unselected-message, stuck/graceful-versus-kill and slow-socket regressions
were carried forward and passed, with original26 review and enhancement23 evidence
retaining their original source identities and limits. This supersedes plan step4's
proposed five-mutant sweep without weakening the source behavior.

All local commands used `/tmp/reordered-run.py` with temporary EN_HOME and stripped
provider/application environment. Raw outputs are in
`/tmp/ensemble-reordered-checks/`; no live profile or paid call was used.

| Actual check / raw basename | Result |
| --- | --- |
| management-core-corrected.txt: core short tests, count1, timeout120s | PASS; ensemble package15.926s |
| management-gui.txt: separate GUI short tests, count1, timeout120s, including actual outside consumer | PASS7.685s |
| management-vet.txt and management-gui-vet.txt | PASS |
| management-race.txt: Fleet/Management/Named/Structured/Submission/Visible | PASS4.879s |
| management-gui-race.txt: Selection/Structured/Settings | PASS4.249s |
| management-ch05-corrected.txt and management-ch06-corrected.txt | PASS0.011s /0.267s |
| management-ch05-vet.txt, management-ch06-vet.txt, management-ch07.txt | PASS; separate example vet/build |
| management-browser-build.txt and management-browser-fleet.txt | PASS; actual built GUI,10 fake-provider requests,111 tree frames; thinking/idle/tool_call/waiting; nested same names, selected replay and two-tab isolation, active-child delete refusal |
| management-browser-speech.txt: existing speech and Responses browser checks | Both PASS; incremental summary before completion and speech-only capability path retained |
| management-mutant-after-tool-rejection.txt | Intended FAIL: removed rejection branch causes `want interrupt, got <nil>` |
| management-mutant-typed-retrieval.txt | Intended FAIL: actual get_submission payload becomes a JSON string instead of an object |
| management-restored.txt: both adaptation checks on correct source under race | PASS1.155s |
| management-current23.txt: `go run ./cmd/grade-edition2 -ch 23 ./agent` |100/100 PASS, exit0,1.071s wall; exclusive slot released |

The first core check, management-core.txt, failed compilation because the new test
used a nonexistent public ChildOutput alias. The wire assertion now decodes its
Status field locally; no extra public API was invented. Two initial example
commands used nonexistent agent/examples/ch05 and ch06 working directories;
Python raised FileNotFoundError before executing Go, leaving their output files
empty. The corrected actual agent/ch05 and agent/ch06 checks above passed. These
are test/setup errors, not passing checks or runtime failures. The final two-line
AfterTool test comment was added after runtime checks; the final source capture
and current23 run include it. No production changed after the green checks.

Source capture `management-source-final.json` contains177 Agent paths and exact
Git blobs, aggregate path/SHA-256
`dbbeba92145d679637577a96bb2fa177a75d12e477b7b92692e8a3342f29bcc5`.
The fleet browser's actual requests/frames are in `management-browser-frames.json`. That file
is this new21 run's capture, not historical live evidence. Scratch mutants were
removed after each run. Formatting and diff checks pass. Historical original26
0/100, revised100/70 and paid task results remain unchanged; this100/100 is the
current23 baseline, not the final revised Stage D/exact-export acceptance.

Comparative check against the actual predecessor and historical e00d5bc→450f7a9:
`management-line-counts.json` counts physical / nonblank non-// / comment-only
lines, with tests/examples separated (same lexical convention as prior stages).

| Chapter delta | Reconstructed from1b7bb652 | Historical accepted management delta |
| --- | --- | --- |
| Production Go | +973 / +827 / +100 | +1,027 / +870 / +109 |
| Test Go | +713 / +660 / +31 | +636 / +585 / +30 |
| Production JavaScript | +34 / +32 / +2 | +34 / +32 / +2 |
| JavaScript tests | +77 / +74 / +1 | +75 / +72 / +1 |
| HTML | +7 / +7 /0 | +7 / +7 /0 |
| Separate Go examples |0 |0 |

CSS adds the same nine physical lines in both deltas. GUI settings/capture/replay
move from Server to its owned per-Agent views; those moves are not additional
services. Production is smaller because memory hooks, duplicated string submission
and already-present turn completion stay out. Added test lines protect the actual
typed management outputs and restored AfterTool boundary. Useful inherited comments
explain owner-scoped lookup, notification-before-read, actor activity ownership,
registry deletion admission and nonblocking whole replay. The delta has100 Go
comment lines against827 active lines; total production has2,083 comment lines
against11,442 active. This remains below the approximate20% teaching target; no
padding or broad rewrite was added to unchanged accepted code.

No source changes outside the reviewed33 Agent paths are needed. Independent
implementation review remains pending; new22 memory and new23 recall obligations
in the preceding ledger are still deferred. No frozen export, tag, renumbering or
primary-runtime replacement is included. Mandatory handoff skill:
`.agents/skills/ensemble-coding/SKILL.md`.

## New22 memory/background-compaction assembly, October14

New21 `aae1e230` / Agent `fd057880` was independently accepted at initial review,
zero correction rounds; coordinator released cumulative memory/recall assembly.
This is maintenance with reviewed-answer exposure, not another blind student run.
Full mandatory `.agents/skills/ensemble-coding/SKILL.md`, AGENTS and all four
workflow/carryover/voice/procedure documents were reloaded after compaction.

Sources: accepted historical memory `6aa7fda`→`05b43a0` under the then-current
`solutions/edition-2/main`, accepted background StageC `a0acfb5e`→`7feb910a`, and
selected final accepted `8935f486` memory Sandbox operations, memory cache-boundary
fallback and typed auxiliary usage. Existing fleet, management, structured submit
identity, supported models, no-total-response-deadline transport and micro_handoff
reasoning cleanup remain. Root additionally released the exact three Agent hunks
from independently reviewed `57920dd3`: render-only completed-handoff label and
its two existing regressions. Stored notes/context/fingerprints remain unchanged.

The actual assembled memory owner, restricted worker, three-request bound,
completion/save-before-cleanup, once-only auxiliary transfer, source/settings
validation, staged commit/recovery, headroom and GUI removal/selection behavior
reuse accepted StageC. No new worker design or general evidence machinery enters.
The original memory switches and byte budgets enter now; actual Sandbox operations
replace the old pre-Sandbox direct filesystem calls. The existing course model
row and its deferred visible-explanation assertion enter with memory.

One transport adaptation preserves the accepted new16 credential/lens correction:
private `send` accepts a request-local config plus main/auxiliary distinction.
Public Send passes main=true; the retained deprecated direct Compress API passes
false. All measured usage contributes to its Engine, but only a main request
updates latest usage and cache observations, after credentials have been resolved
and sanitized. Successful401 retry remains one logical lens/usage observation.
Background workers keep their own Engine and contribute drained usage buckets.
The three-provider cache regression now uses distinct auxiliary counts and checks
actual capture bytes, latest-main usage and paid-but-invalid auxiliary spend.

Deferred precisely to new23: recall package/interfaces/config/events/startup and
actor hooks; StageC recall recovery guard; TestCompactionRecallStartup, the recall
recovery-hold assertion and worker no-Recaller assertion; unrelated Recall.Disabled
fixture fields; judge routing/usage and two differently priced judge controls.
No Recall/Recaller/KindRecall, Codex credential or Gateway Go symbols are present.
Codex24, Gateway26 and unrelated application-configuration/crossover UI remain out.

All runs used stripped provider/application environment and temporary EN_HOME via
`/tmp/reordered-run.py`. Raw basenames below live in `/tmp/ensemble-reordered-checks/`.
No paid calls or live profile/process changes occurred. Historical live evidence
and accepted unchanged StageC mutations retain their original sources and limits.

| Actual check / raw basename | Result |
| --- | --- |
| memory-final-core.txt: core `go test -short -count=1 -timeout120s ./...` | PASS; ensemble16.959s |
| memory-gui.txt: separate GUI tests, including outside consumer | PASS8.380s |
| memory-vet.txt, memory-final-vet.txt, memory-gui-vet.txt | PASS |
| memory-race.txt: Compaction/Memory/Structured/Submission/Cache/OAuth | PASS7.997s |
| memory-gui-race.txt: Compaction/Selection/Settings/Structured | PASS4.456s |
| memory-ch05.txt, memory-ch06.txt; matching -vet logs; memory-ch07.txt | PASS separate example tests/vet/build |
| memory-auxiliary-baseline.txt | PASS1.060s under race |
| memory-mutant-auxiliary-last.txt | Intended FAIL on all three providers: auxiliary77/9 replaces main1/1 or10/2 |
| memory-mutant-auxiliary-lens.txt | Intended FAIL on all three providers: actual main capture rotated |
| memory-handoff-restored.txt | PASS1.441s under race: handoff/replay/context/marker/auxiliary checks |
| memory-restored.txt | PASS1.057s, correct auxiliary path after scratch mutants removed |
| memory-grader-new22.txt | Final revised new22 **100/100 PASS**, exit0,3.152s wall; includes unchanged14 parity |

The grader command ran from primary compaction: `go run ./cmd/grade-edition2 -ch 22
-json /Users/bill/projects/ensemble-reordered/agent`. Its numbering edits were
uncommitted at primary HEAD57920dd3; `memory-grader-source.json` binds dispatch,
new22 and parity14 files. Four proxy context-canceled diagnostics remain in raw
passing output. The final explanatory composition comments and equivalent operand
order were reconciled to accepted primary after this run; no behavior changed.

Actual failures remain: memory-first-core.txt failed compilation on a leftover
cfg.Recall reference; memory-core-second.txt on an unused bytes import after
removing a recall-only test. Corrected memory-core-third.txt passed. These are
boundary assembly errors. After57920dd3, memory-handoff-race.txt failed the exact
OpenAI cache-marker fixture: its text now includes the completion label. The
corrected assertion expects exactly `micro_handoff completed. Your checkpoint
note:settled boundary` for OpenAI, retaining exact Anthropic note text, actual
marker positions/counts and volatile-tail exclusions. This is a fixture correction;
no runtime behavior was weakened. Root received it for the primary test too.

Comparative counts (`memory-line-counts.json`) use physical / nonblank non-// /
comment-only lines, separating tests/examples. Against accepted new21, production
Go grows2225/1797/326 and test Go4137/3895/148. Historical accepted memory alone
adds915/745/124 production and835/782/35 test lines; accepted StageC adds
1308/1055/197 production and3276/3092/108 test lines. These source deltas are guides,
not additive identities: existing GUI locking and fleet prerequisites are reused,
recall hooks excluded, Sandbox/cache obligations restored, and the later handoff
label and auxiliary boundary checks included. JS adds28/27/1; HTML1/1/0;
examples and JS tests unchanged. Total production Go has2409 comment-only against
13239 active lines (about18%); useful existing ownership/receipt/recovery comments
remain, with no padding to reach the approximate20% target.

`memory-source-final.json` binds198 Agent files/blobs; aggregate path/SHA-256 is
`f0bf38db46701a2550c6b012de436ff1a7df541bf9c2887653fc2112b85b50e1`.
Formatting/diff checks pass. New22 independent implementation acceptance is pending.
No frozen export, tag, renumbering or whole-primary-runtime replacement is included.

## New23 recall assembly, October14

New22 `c67173c1` / Agent `23e0aae5` received independent and coordinator acceptance
at initial review, zero corrections. This next stage extends that exact source;
it carries accepted recall forward, with no new design or separate plan gate.
The complete current recall chapter (new23,755lines) and selected source delta
were read. Mandatory `.agents/skills/ensemble-coding/SKILL.md` and all four
full governing/craft documents remain active from the preceding reload.

The guide is historical accepted `05b43a0`→`a2f0a4f` (20paths). Selected final
`8935f486` recall spokes and tests additionally preserve Sandbox corpus access,
trusted external skill-directory confinement, supported Sonnet fixtures, Responses
judge routing, lazy indexing and the StageC committed-recovery hold. Existing
actor begin/mailbox hooks, recorded attachment hashes and checkpoint removal now
enter at their actual dependency boundary. Recall owns its prompt/parser and
reaches the stateless judge through Agent→Engine. Skills exposes its existing
constructor-selected directory; discovery still grants no capability.

The only judge transport adaptation is `send(..., cfg, false)`, preserving new22's
shared auxiliary path and resolved-credential sanitization. Codex's two request
fields remain deferred to24. Actual judge usage keeps its request-local model
and tier, without changing main usage/history/lens. The three-provider auxiliary
regression now executes both direct Compress and Judge with distinctive77/9
measurements, checking both paid counts and unchanged main artifacts/usage.
The accepted unknown-price and supported-but-unpriced judge tests return now.

Restored deferred obligations include StageC's recall startup/permission control,
committed recovery refusal, no-Recaller worker assertion, and existing disabled-
recall fixture settings in compaction, protocol and fleet tests. Typed-management,
OAuth once-only401/redaction, submit-identity and handoff-label regressions remain;
no historical whole-file copy removes their reconstruction-specific assertions.
No Codex/Gateway implementation or unrelated application-configuration UI enters.

All raw logs below are `/tmp/ensemble-reordered-checks/recall-*.txt`, using the same
isolated environment/temporary EN_HOME and real runtime with external peers.

| Actual check / suffix | Result |
| --- | --- |
| core: core short tests count1 timeout120s | PASS; ensemble32.128s, recall0.006s |
| gui: separate GUI tests including outside consumer | PASS8.087s |
| agent-vet and agent-gui-vet | PASS |
| race: Recall/Cache/recovery-hold/construction/handoff/OAuth | PASS19.191s; included actual15s judge timeout |
| auxiliary-baseline | PASS1.074s under race |
| mutant-judge-main | Intended FAIL on all three providers: wrong main=true replaces main occupancy with77/9 |
| restored: correct auxiliary path plus all existing BM25/chunking/quota/parser/cap tests | PASS1.078s ensemble and1.017s recall under race |
| ch05 and ch06 | PASS0.011s /0.273s, external consumers recompiled |
| grader-new23 | Final revised **100/100 PASS**, exit0,16.417s wall |

The first race selector did not select the internal/recall unit names and printed
`[no tests to run]` for that package. Its subsequent restored check selected and
ran those actual tests; no empty test selection is counted as unit verification.
Core tests passed on their first actual run. The final cache fixture addition was
made after that core launch and is covered by the subsequent focused race and
restored checks. No paid/provider profile access, new browser sweep or unchanged
historical mutation sweep occurred. Accepted original recall/Sandbox/StageC evidence
retains its earlier source identities and semantic/live limits.

The grader ran from primary with `go run ./cmd/grade-edition2 -ch 23 -json
/Users/bill/projects/ensemble-reordered/agent`; `recall-grader-source.json` binds
its working dispatch/new23 files while numbering was being checkpointed separately.
`recall-source-final.json` binds206 Agent paths/blobs, aggregate path/SHA-256
`cd2d2711dca4e7b83a4e1c913d312ae8c4ed17679bc6ea64e7efb7b071fc8243`.

Comparative counts in `recall-line-counts.json`, using the existing physical /
nonblank non-// / comment-only convention: new production Go +851/669/144 versus
historical recall +825/643/143; tests +706/635/46 versus +612/550/39. Extra production
is accepted confinement/recovery/Responses integration, partly offset by the
already-present auxiliary transport. Extra tests restore deferred worker/recovery
and judge accounting controls. JS/HTML/examples unchanged. Total Go production
has2553 comment-only lines against13908 active (about18%); comments retain the
existing ownership, lazy-scan, cancellation, format-owner and deduplication reasons.
No added abstraction or speculative feature is justified by the line counts.

Formatting/diff checks pass; scratch mutant removed. Independent new23 acceptance
is pending. Later24/26 carry-forwards and exact-export StageE validation remain.
No frozen exports, tags or primary-runtime replacement were made.

## Optional24 carry-forward, October14

New23 `8faca265` / Agent `4518af32` is independently accepted at `754bc40f`,
initial review with zero corrections. Optional24 extends it with the accepted
16-path `471c7ae5` delta; this is maintenance reuse, not a new student attempt.
The complete chapter, original implementation/self-review record and selected
source were read; `.agents/skills/ensemble-coding/SKILL.md`, AGENTS and all four
full governing/craft documents were reloaded after compaction.

The same read-only Codex credential reader, explicit selection, paired account/
bearer, guarded endpoint, redirect refusal, cache affinity and Responses route
are restored. No external credential file was accessed. Existing engine transport
keeps the private `main` flag, resolved-credential redaction before logical lens
observation, once-only successful401 accounting and caller cancellation without a
short total deadline. The Judge now receives its deferred Codex instruction/key;
its call remains auxiliary `send(..., false)`. Memory, recall, typed submission
identity and both micro_handoff corrections remain unchanged. Merge conflicts in
README ordering and the transport/judge insertion points were resolved before
compilation; no failed runtime check was hidden as a merge setup issue.

The accepted current `8935f486` Codex integration test includes `df38abe0`'s
visible explanation fixture and `3aad2b7e`'s detailed-summary environment check.
No application-configuration feature from that later commit is imported.
Historical13 mutation results and actual paid runs retain their original source
identities in `codex-endpoint-chapter.md`; unchanged behavior is not remutated.
The current tests exercise real transport/tool/replay/credential boundaries using
external HTTP fakes, including auxiliary calls and existing OAuth/cache controls.

Raw `/tmp/ensemble-reordered-checks/codex-*.txt`: core short/count1 tests PASS
(ensemble32.489s), GUI tests PASS8.261s, core/GUI vet PASS, and focused
Codex/OAuth/Responses/main-auxiliary-boundary race tests PASS3.473s /1.017s.
This optional chapter has no grader. No live/provider-profile access or new paid
run occurred. No unrelated browser sweep or unchanged example-module rerun.
Formatting and diff checks pass.

`codex-source-final.json` binds210 Agent files/blobs, aggregate
`c9edbfbc9e3a41e85e6035a71b9c13c4ffc9c2c4d886ceabceb43c75da39381d`.
`codex-line-counts.json` uses the existing physical/active/comment-only method:
production +193/148/33 versus historical +194/148/33 (one preexisting blank line);
tests +341/313/17 versus +288/263/15, due to the accepted explanation and summary
fixtures. Examples and other languages unchanged. Existing ownership, explicit
route and credential lifetime comments remain; no new abstraction is introduced.
Independent carry-forward review is pending. Gateway26 follows; no export, tag,
renumbering, main replacement or live process/configuration change is included.

## Gateway26 carry-forward, October14

Optional24 `f18737f6` / Agent `6fdf4ccc` received independent initial acceptance,
zero corrections. Gateway extends it using the complete accepted15-path delta
`ab6d0516^`→`b5c35a39`. The full current717-line chapter, original student/reviewer
records and all selected source files were read. The mandatory skill and all
four governing documents remain active from this turn's complete reload.

All13 added source/test/README files match accepted `b5c35a39` byte-for-byte and
are unchanged in accepted primary `8935f486`. Existing StageB schema dependencies
are retained alongside Gateway dependencies; the resulting go.mod/go.sum match
`8935f486`. No core actor, transport, GUI or application-configuration behavior
changes. The server owns routing/queues, the independent public client imports
no private runtime, and the separate application uses actual Agent.Submit. Final
accepted parent diagnostics and receipt-written ordering fixes are included.

Raw `/tmp/ensemble-reordered-checks/gateway-*.txt`: affected public-client/server/
actual-actor tests under race PASS1.045s/1.253s/1.045s; core vet PASS; unchanged
GUI recompilation PASS (explicit `-run '^$'`, no GUI test-execution claim).
The current26 grader, run from primary with the explicit reconstructed Agent path,
returns **100/100 PASS**, exit0,0.743s wall. Its seven rows exercise the actual
binary, outside consumer, authentication/authorization and receipt/restart paths.
`gateway-grader-source.json` binds dispatch/checks at primary `b4746ac9`.
The grading slot was released immediately afterward.

No runtime adaptation or fresh mutation is needed: the original25 targeted
Gateway defects, follow-up diagnostics/welcome checks and actual interview remain
historical evidence at their recorded versions in `ch27-code-review.md`/`ch27.md`.
No paid run, external profile read or live process/configuration change occurred.
The first source-identity audit found only Go-sum line ordering from the union;
exact set comparison passed and restored accepted ordering. This was an audit
assertion failure after the passing checks, not a runtime failure. No checksums
or dependencies changed after grading. Formatting/diff checks pass.

`gateway-source-final.json` binds223 Agent paths/blobs, aggregate
`71b1a664bb1a104afaabd39d57982bd1d1e34d684e0f4c8bfcb72ababe884cbc`.
`gateway-line-counts.json`: physical/active/comment-only production delta
+1871/1475/306, test delta+592/543/29, exactly the historical accepted delta.
Other languages/examples unchanged. Existing ownership, callback/receipt and
snapshot-limit explanations are retained. No new abstractions or repeated
historical test sweep were added. Independent carry-forward review is pending;
root owns later exact-export/default-path validation and publication decisions.

## Final cumulative source identities, October14

The earlier pending statements above describe their submission dates. Independent
reviews have accepted all stages in this table. Gateway26 `59e25a25` received
initial independent acceptance with zero corrections after exact-source and actual
output review; optional24 likewise passed initial review without corrections. This table identifies source for export, not a claim that export installation
or final default-path checks have already completed. All entries use the `agent/`
subtree. Chapters0–13 remain outside this reconstruction;25 stays reserved.

| New chapter | Accepted cumulative source commit | Agent tree |
| --- | --- | --- |
| 14 | `82d3952a0267fff0e540238a1326da764b44ea64` | `bccc7d3d5fd7eb4dc14a789ff9f3570c66200400` |
| 15 | `6edde04092a02b1d708c9ff1b15338a3f9b132b0` | `d7c38fa74bf9cc1a62c60e61a7e5f105feeccd98` |
| 16 | `ec853c18bc2d7f2e2ab3a89d23067d50a655d628` | `c35c147a692100b325d3b46805e59105b4a4e2c0` |
| 17 | `a4b9a1d33814bb5a660f74ab0b432c9cfea975aa` | `646d69872ae016eb9729b95d0c262eabe98a2fd9` |
| 18 | `6677cb9cb96ea22c1c7dc6d8a1e11e8f3843f706` | `0a4251b225b98a711acbe31bf00672aaa592ae15` |
| 19 | `dee87fbcd2dd793e37d820597c513c3556ea7d70` | `c01fcc28a43d0a581be9ed8aa32a5df492dede97` |
| 20 | `9f4fdca297d636dfdfd5dfa0f82fe063e8e0e56f` | `1fd589c98da217f8b696f80bec6e7adb38647e12` |
| 21 | `3a5673b601dda3934754b13c33da985694ea5dc5` | `34b461a12ee6d080ecf104940635f53f62874e68` |
| 22 | `c67173c18ba09dc6a1d53c607d69da969293562f` | `23e0aae52df171420e0b00b9e05d3ac447660ea6` |
| 23 | `8faca2652d6bacd41014632450d5d6798862c852` | `4518af327241e99543cd027e62ef9e9596e18f08` |
| 24 | `f18737f62e0f9aaf553a8d8a32334ac2191bb0a9` | `6fdf4cccf54989f45829bbf04ca77ddbd062ab6c` |
| 26 | `59e25a25dab3c2e7d41c32654013694d7de1bce7` | `6982337a032962f36626cb23f68e6488046cccdc` |

The14–21 maintenance replacements are owned by the separate handoff coder and
reviewed in `codex-handoff-carryforward.md` at primary `87b1a616`. They preserve
this note's historical stage commits while carrying the accepted cleanup and
render-only completion label into each proper boundary. New20's corrected source
starts from root's `1b7bb652` carry-forward of accepted `b4708839`; it does not
replace the documented one-round README correction or initial source evidence.
New22–26 already retain those runtime corrections. Their complete implementation
and review provenance remains in the sections above and the primary runtime audit.

The new15 deferred auxiliary accounting/memory cache fallback is restored in22;
judge/recall accounting and corpus hooks are restored in23; AfterTool and management
enter21; Codex request fields enter24; Gateway enters26. Typed submission identity,
root/named-peer ordinary chat and delegated submit obligations remain preserved.
No early memory/recall imports or later application-configuration feature were
used to fill an earlier dependency. This is reviewed maintenance reconstruction
with answer exposure, never a fresh blind student evaluation.

Root owns installation of exact accepted exports, final advertised/default-path
grading, package-path verification and preservation of historical tags/results.
The original paid receipts remain historical; no new paid run or live configuration
change was made by this reconstruction. The full current runtime is maintained
separately and must not be replaced by this cumulative Agent wholesale.
