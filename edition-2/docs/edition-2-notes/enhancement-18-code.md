# Edition 2 Chapter 18 grader: final submission

Coder: GPT-6 Astra/high. Original Chapter 19, accepted direct `ch18`, legacy
`solutions/ch19`. Read the complete original, student/comparative records,
mandatory `.agents/skills/ensemble-coding/SKILL.md`, carryover and workflow.
Original graders, runtime exports and historical results remain unchanged.

The separate command is `go run ./cmd/grade-edition2 -ch 18 -json`; the legacy
comparison adds `-gui ./solutions/ch19` and positional `./solutions/ch19`.
It builds the actual optional GUI command, uses a typed primary and isolated
profile/workspace, and preserves all seven original weights: 15/15/15/15/10/15/15.
The small explicitly Edition 2 adapter reuses the original signed OAuth and
Responses HTTP peers. New behavioral predicates and launch code live separately
in `internal/gradeedition2/ch18*.go`.

## Results and fixture corrections

Evidence is under `solutions/edition-2/evidence/ch19/enhancement18/`.
Actual final full results `accepted-final.json` and `legacy-final.json` are
both **100/100**, exit zero, in **2.246 s** and **1.570 s**. These are actual
full runs, not composed totals. `final-source.json` binds the five owned source
files and unchanged frozen trees. Final focused real-command race testing and
vet pass (`final-race.txt`, `final-vet-run.json`). The race test's package timing
is in its raw output; command elapsed time includes build/startup.

Earlier actual full results `accepted-green.json` and `legacy-green.json` were
100/100 in 2.676 s and 1.794 s, bound by `stage-source.json`. Green checkpoint
`d1de19ee` additionally publishes the fixture constructor through its existing
mutex. Its subsequent focused race result passes in 2.557 s. Final changes only
classify abnormal process exits; no runtime implementation changed. Independent
combined acceptance is pending.

Retained initial results are accepted 100/100 and legacy 55/100. The legacy
correctly refused missing-grant and quota scenarios but deliberately exited 1;
that completed refusal is now distinguished from successful sessions, which must
exit zero. Actual token issuance/request admission, observed turn error and no
metered credential switch remain required. A signal, early exit or arbitrary
nonzero status alone cannot establish refusal.

The original fake's terminal envelope omitted `model`. Legacy provenance-aware
replay correctly excluded opaque material with unknown model provenance. Actual
retained terminal responses identify their model: `model-fixture-provenance.json`
records sanitized fields from original live log lines 14 and 191. The forwarding
peer fills only an absent terminal model from the actual request; exact encrypted
replay remains required. The original fake is preserved. A failed initial
extraction used SSE framing against per-event JSON logging and found no events;
the corrected extraction reads the actual event/data representation.

The first corrected accepted attempt scored 30/100 because its callback server
closed after receiving the code but before finishing the browser response. The
source confirms this ordering. Login now waits for the bounded actual command
and checks real token exchanges instead of treating browser EOF as failed login.
That attempt remains `accepted-stage.json`, between the initial and green source
bindings; no separately measured whole-source hash was saved for that intermediate
attempt. The missing browser acknowledgment
is a real usability limitation: the browser may see EOF despite successful
authentication. It is not a Responses parser defect or evidence that token
exchange failed. No frozen runtime repair is included.

The first race check exposed the original OAuth peer assigning its server pointer
after starting its handler goroutine. The adapter publishes that completed
assignment through the existing mutex that handlers acquire before reading it,
before releasing the URL to the child. Both original warning output and corrected
race result remain. This is fixture synchronization, not a runtime change.

## Observed claims and boundaries

Positive API-key and subscription requests use their actual configured/minted
bearers. Secret canaries are absent from observed output, request bodies and
non-profile files. Actual PKCE/issued-client exchange, forged-signature and nonce
rejection, inference-grant refusal, and two proactive rotating renewals are
observed. Each renewed request is checked against its corresponding minted token,
not the final token retrospectively. Profile security, concurrent/cross-process
single flight, atomic replacement, logout revocation and transient recovery retain
accepted source tests and the earlier 39-mutation evidence; this numeric fixture
is not a general OAuth/security proof.

Responses checks include actual successful `/etc/hosts` bytes in the correlated
function result and an encrypted reasoning item in the next request. The plan
peer retains absent Content-Type and empty terminal output. The summary fixture
withholds the remaining response until the GUI receives the first summary text;
valid chunk splitting is allowed. Complete separate summary/text deltas include
tool-call-adjacent content, not a summary conditioned on completed tool output.
Native speech and listening retain their original separate evidence.

API-key caching requires the actual explicit marker on developer constitution
content plus explicit mode and the four-marker budget. Plan requests omit the
paired mechanism under the accepted exception; historical provider rejection
established only the options field, not an independent marker rejection. Both
admission and midstream quota scenarios require real authenticated requests and
reported refusal without switching to the available metered key.

Original grading remains 15/100 on `06cbf20`, including its vacuous billing award.
The real plan task completed in 14.887 s, but its summary-dependent launcher
killed the idle CLI at 150.03 s, without graceful shutdown/save. No public summary
or cache grant was observed. Those historical results, source versions and
limitations are not replaced by these local protocol-peer runs. No paid request
or production sign-in occurred during enhancement.

## Targeted sensitivity and exact examples

`mutations.json` records **21 detected runtime defects** and one valid
split-summary control on `d1de19ee`. `exit-mutations.json` adds **two detected
abnormal-exit defects** on final bytes. All 23 defects fail an intended row;
these overlapping observations are not 23 independent proofs. The no-encrypted-
include and dropped-opaque cases both report missing replay content: this grader
retains the later diagnostic within a row, not every independent failing predicate. Every mutation
has its exact patch and raw report. Both scratch-copy restored controls pass.
The final exit change rejects signals, deadlines and reported Go runtime panics
rather than allowing them to masquerade as expected refusal. Ordinary nonzero
refusal still requires the scenario's actual issuer/request/error observations.

| Changed claim | Targeted defects and observed boundary |
| --- | --- |
| Actual credentials and secret exclusion | Wrong bearer fails credential selection; logged real fake bearer fails non-profile artifact/output scan. |
| Real authorization and permission | Plain PKCE, unchecked signature, unchecked nonce and allowed missing grant each fail OAuth flow. |
| Correct proactive renewed bearer | Expiry-only renewal and retaining the old access token after rotation fail the refresh row. |
| Completed tool and replay content | `store:true`, omitted encrypted inclusion, dropped opaque item, wrong result ID and empty successful read fail Responses format. |
| Paired caching and actual legal marker | Restored plan controls, absent API marker and API marker misplaced on the message object fail caching. |
| Full separate incremental summary | Summary as answer text, summary delayed until final, and trimmed delta whitespace fail summary delivery/content. The delayed-final defect reaches the gate timeout with the working incremental baseline. |
| Actual refusal observation | Dropped `turn_ended` error fails both grant refusal and billing; these are coupled, not independent errors. |
| Successful/failed process lifetime | Exit 17 after successful GUI work fails the success rows. Panics after bad nonce or after failed turn fail the corresponding identity/billing row; earlier printed errors cannot rescue a crash. |

The unchanged original peer still enforces the other valid wire categories.
Existing accepted evidence supplies the wider identity, state, host, resource,
concurrent rotation, terminal/transient recovery, catalog, permissions, namespace,
foreign replay, usage and browser-speech claims: original-numbered `mutations.json`
(28), `supplemental-mutations.json` (7), `plan-cache-mutation.json` (1), and
`round1-mutations.json` (3). Those 39 historical cases retain their source and
scope; they are not newly run here. The preserved captured-wire cases cover
absent-header SSE recognition and completed-item fallback. No general security,
all-error-path, native-audio or provider-compatibility certification is inferred.

`samples.json` binds all nine fences to manuscript SHA-256
`e057e264f91a20d4c8da1d2852f64e244283afcc3e455e8e4f67113e50d6a730`.
The conceptual credential vocabulary compiles separately. The exact core build,
scratch setup, typed skill, scheduling file, login/status/models, supported
`gpt-6.1-sol` high/detailed chat and logout commands execute. The first request
contains the published prompt; the second contains the successful correlated
complete scheduling file. The peer then supplies the known optimal answer.
The client exits zero, creates its conversation save, and logout reaches actual
revocation plus local signed-out state. `samples-run.json` preserves commands,
outputs and exits; `samples-requests.json` preserves actual model/effort/summary
and result bytes. The two grader-command fences are covered by the full command
runs, rather than executing duplicate grades as part of the sample runner.

Only the external issuer, browser authorization visit and inference provider are
local peers. No actual account sign-in or paid request occurred. The reused fake
catalog lists GPT-5.6 rows while the exact sample selects declared GPT-6.1 Sol;
this establishes command mechanics, not catalog-to-selection eligibility. The
historical real account catalog supplies its dated availability evidence. No
summary or cache grant is emitted by the sample peer. Its known answer tests
propagation, not scheduling intelligence. The sample uses the core CLI and
therefore neither connects to nor changes the user's GUI on port 8088.

## Source and preservation

`source-counts.json` separates the new grader, driver, focused test and adapter
from the shared command's whole-file count. Useful comments explain ownership,
real exchange prerequisites, callback/exit distinctions, fixture publication,
stream gating, request-local model metadata and protected-store exclusion.
All newly exported grader/adapter symbols have documentation. The small adapter
avoids copying 2,429 physical lines of existing protocol peers. It exposes only
the options, observations, constructors and fake API canary needed here; it is
not a new product API. The forwarding helper serves these fixed SSE fixtures,
not a general HTTP rewriting relay.

`final-source.json` records exact accepted and legacy tree identities; their
working paths, original grader files and original Chapter 19 are unchanged.
Concurrent main GUI/crossover and first-edition work is excluded. Owned commits
use an isolated index with identical base/parent, configured hooks and expected-old
reference publication. Final disposition belongs to the independent reviewer;
Chapter 19 / original 20 is not started by this submission.
