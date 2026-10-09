# Chapter 10 proposed live-plan review

October 8, 2026. Independent review by `/root/coder_ch08` of the proposed
`main/evidence/ch10/live-matrix.md` in source freeze
`122b04a57e7c6ea158901d00bfc0f1a3a8c75329`, with status-only handback
`1e041f5`. The reviewer read the complete matrix and implementation-status record
against the complete current Chapter 10 contract and loaded live-evidence skill
requirements. Earlier reviewer/grader/Chapter 8 coder exposure remains disclosed
in chapter-10-review.md. This is plan review, not historical comparison or
verification that the proposed runs occurred. No provider call, credential read,
runtime run or mutable implementation read was performed.

## Disposition and current blocker

The four paths cover the right interfaces and most required feature families.
An upper bound of 22 generation requests per provider, 66 total, plus one discovery
request per provider is a plausible starting ceiling. It is not yet a sufficiently
concrete execution/reproduction plan, and this review does not authorize spending.
No extra calls or larger ceiling are requested by the recommendations below.

The coordinator reports a current independent fault finding: a partial public
Append faults the Agent, but Close returns nil and creates/replaces a checkpoint.
Repair and affected independent clearance must precede the final live-source
freeze. The isolated deterministic gate on 122b04a and remaining coverage are
separate; earlier local successes cannot clear this later finding. Bind any
corrected live launch to its actual source, rather than relabeling 122b04a.

## Concrete corrections before live release

1. **Turn the request allowance into an executable schedule.** Give each row its
   exact human/public prompt, expected tool/task shape, process/store ownership,
   and bounded continuation allowance. State finite output-token and operation
   deadlines and small scratch-file/input sizes. Keep the current A/B/C/D ceilings
   6/4/6/6 as hard maxima unless the coordinator reviews a redistribution. Count
   every outgoing generation attempt, including transport failure, cancellation,
   parse refusal and startup/shutdown surprises, before it is sent. Do not count
   only responses that arrive. Disable automatic retry and unauthorized redirects
   or secondary probes; preserve refusal/cap exhaustion as an outcome. Discovery
   has its own one-request/provider counter, including any pagination. Name the
   intended Gemini 3.8 Flash target and preserve a discovery/access limitation
   instead of silently substituting an older model.

2. **Make B's independent current settings observable.** Name the old recorded
   turn cap and the changed current policy, the real prompt that exercises the
   new capture, and the exact result to observe. For example, a one-request turn
   with an actually returned tool call can show the batch finishing before the
   next model request is refused; a plain answer alone does not demonstrate
   that boundary. Record model noncompliance honestly without automatic retry.
   Change one real GUI preference and show that current policy/preferences survive
   reopening independently of older session checkpoint values. Merely displaying
   both settings is weaker than the promised independence. Use actual keyboard
   or pointer actions and record which. Bind screenshot/DOM/socket observations
   to the same launch and acknowledgement; a screenshot alone cannot prove order.

3. **Specify C's branching comparison without buying duplicate work.** Identify
   the real-model export, the genuinely older checkpoint, later accepted tail,
   and empty import destination. Use independent copies/Ensemble lifetimes to
   avoid mounting a copied SessionID twice in one root. Preserve identical pending
   prompt and current render inputs when comparing old-checkpoint/tail, null
   rebuild and prefix-free import paths. A paid continuation can supply the actual
   body; offline rendering/reconstruction with endpoints disabled can compare
   equivalent branches without extra paid generations. State which branch proves
   immutable origin after later save/reopen and which actual earlier request
   returns history_unavailable. Compare usage/Skills/retained watch facts as
   applicable, excluding only declared transient identities. Two-Agent turns
   must both be attempted and report partial result/usage independently if one
   fails; do not silently abandon the other Agent to make the run look complete.

4. **Make D's seeded setting and real effect distinguishable.** Name the public
   accepted-call seam and exact local fixture used to leave pending tool_limits,
   its provenance and the saved field values. Public candidate-authority checks
   cannot be bypassed by inventing a plausible limits event. Label that preparation
   as controlled local evidence, not a provider-generated setting. The subsequent
   real model must actually issue the next attempted call; show the consumed
   setting and resulting effect/refusal once, then the restored default on an
   appropriate later path or retained deterministic control. Include the exact
   catalog/bindings, retired/reloaded activation identities, moved directory and
   manual/grant request markers. Preserve an unexpected extra call consuming the
   setting as the actual result. A provider that returns no replay-bearing opaque
   material cannot establish that live compatibility case; retain independent
   opaque controls instead of inventing a signature.

5. **Finish immutable support preflight before credentials or discovery.** Bind
   the final complete source/module/dependency map, CLI and GUI binaries, compiled
   public consumer binary plus source, catalog/bindings, format document, browser
   driver/proxy/recorder/verifier support, and exact sanitized launch manifest.
   Record actual build association; a source hash beside an unrelated executable
   hash proves neither. Exercise changed support with local fixtures first. The
   verifier must check complete valid-path identities for every provider before
   any replay/derived write, with passing parents and one-at-a-time intended
   source/binary/support/dependency/launch refusals. Reuse permitted preceding
   student machinery when suitable. Preserve original terminal/socket/provider
   receipts and before/after store hashes separately from derived reconstructions;
   a later correction gets a new binding, never an overwritten original launch.

## Scope and limitations to preserve

Row A is a genuine human CLI proposal: two real PTYs/processes, a verified
scratch edit, stable SessionID, restart recall and zero tool reexecution. Keep
the marker out of the recall question and inspect the file after both processes;
an answer containing the marker without a corresponding actual request is not
enough. Local /session, /checkpoint, /history and shutdown actions consume no
generation allowance but still need their original acknowledgements and outputs.

Row B appropriately uses the closed CLI store, so the GUI mount tests resume
instead of competing with an existing writer. A second Page is a second client,
not a second live store owner. Distinguish Page close, socket reconnect, terminal
detach and complete server restart. Show historical no-live-owner status and
session usability afterward. Zero omitted events is an honest small live example;
large-window boundaries can stay deterministic. No-restored-speech claims should
state whether evidence observes speech API admission, pause state, actual native
audio or merely text. Silence on a muted/blocked browser is not proof that no
speech was requested. Do not replay historical text as a live answer to simulate
this property.

The plan correctly separates malformed stores, checked I/O faults, large limits,
signals, lock exclusion, stale handles and forged authority into zero-provider
deterministic controls. Retain applicable source-bound independent results;
those rows need not be repeated through paid generations. Conversely, a deterministic
control cannot replace the required actual all-provider CLI/browser/public use,
real-model snapshot-only import/continuation, or observed file/tool effects.

The plan has no completed run, model access result or successful audio claim.
Some expected outcomes depend on model behavior; the eventual result matrix must
show missing observations and partials explicitly. A revised support/feature
matrix is needed before release, not another user-permission ceremony. Initial
source, real-interface evidence and student teaching experience still freeze
before any historical persistence comparison.

Source bindings at review:

- Proposed live-matrix.md SHA-256:
  `bd2f8332781e4b77fa8ec8bce3102ba4f4d8abaa62258729ece8ed688c271b4a`.
- Current chapter-10.md SHA-256:
  `56f52e6ac8d251f6c775d226c4b6c48fec1311542d15450f3b666617584cbbaa`.

Disposition: useful bounded proposal, pending the five concrete plan corrections,
known runtime-fault repair, remaining local clearance and immutable preflight.
No provider release or complete runtime acceptance is implied.

## Revised plan review: 8882a18

October 8, 2026. Read the complete revised matrix frozen in
`8882a18cf98e9a4b70afccfbe980f6344630aae6`, against the five corrections above.
Its SHA-256 is `d7c273ab81cc81fbb396904a342ce3540aee55dc2e0ae8757d9caf470bcbb4ee`.
This is another proposal review, not a live receipt. Full mandatory skill and
Chapter 10 contract remain loaded. The coordinator reports the append/Close
repair and affected fault/lifecycle clearance on this source; the independently
run storage subset has its own bound receipt. Neither replaces the remaining
local gates or support freeze. No credentials or provider traffic occurred here.

The revised matrix now supplies exact prompts, small scratch inputs, process and
store ownership, 4,096 output tokens, 120-second attempt/ten-minute row deadlines,
and hard before-send accounting for all generation outcomes. The proposed ceiling
remains 6/4/6/6 per provider, 22/provider and 66 overall, plus one discovery/provider.
There is no retry/pagination/older-Gemini escape from those caps.

C now names the genuine older checkpoint and later real-model tail, three inert
comparison branches, the snapshot-only import, unavailable pre-origin send, and
save/reopen check of immutable origin. Identical render inputs and endpoint-disabled
comparison avoid duplicate paid work. Its two Agents must both be attempted and
their partial completions/usage retained independently; the support implementation
must preserve that explicit exception to a blanket stop-on-peer-failure rule.

D names a legitimate loopback model-response fixture through the public Config
and Ask path, so Actor/Tools/Jobs actually create the pending setting. Exact
empty-pattern and 17-byte values, the later real attempted call, moved catalog,
retired/reloaded identities and missing-opaque limitation are explicit. Local
fixture usage/provenance is separated from real-provider outcomes. It is not a
forged limits event or a claim that the provider chose that setter.

The support-freeze section correctly binds compiled consumers, all runtime and
GUI assets, nested modules/dependency selections, catalogs, recorder/proxy/verifier
and sanitized launch identities. It requires a valid positive parent before
each intended identity refusal and before-write verification. These controls are
still prerequisites to execute, not evidence created by revising the prose.

Two small but concrete setup corrections remain before calling the plan complete:

1. **Use the actual CLI policy authority.** A promises recorded turn policy 3
   from a bare CLI launch, and B later compares that history with current 1.
   Chapter 8 supplies the policy path option to GUI/public construction, not the
   standalone CLI. A targeted read of frozen `cmd/main.go`, `cli/main.go` and
   `cli/selection.go` confirms that this CLI has no policy selection and uses
   raw zero/effective 16. It does select write_file and 4,096 output tokens.
   Keep the CLI's actual raw-zero/effective-16 capture, enforce three attempts
   with the separate live proxy, and compare GUI current 1 with that actual
   historical capture. D's first PTY ceiling 3 likewise means proxy spend bound,
   not an unavailable CLI policy setting. Public C/resumed D may deliberately
   update their own policy through the existing public seam. Do not add a new
   runtime feature merely to make the proposed transcript true.
2. **Name the no-restored-speech observation.** B's top row requires this feature,
   while its detailed schedule says no native hearing/speech claim. Keep the
   honest no-hearing limitation, but specify actual native-speech API admission
   and Page pause/queue observations for historical load/reconnect. An absence of
   audible output or displayed text alone cannot prove no speech was requested.
   This requires local browser evidence support, not another generation prompt.

Both findings were sent promptly to root for student-safe correction. The frozen
CLI reads above were for feasibility only; this pass did not inspect mutable
implementation, change runtime, rerun a broad gate or begin historical comparison.
Most original plan corrections are now satisfied at the proposal level. Full
plan closure awaits those two wording/setup fixes; paid launch still separately
requires remaining local clearance and the actual immutable support preflight.

## Frozen provider-support review: 9822b2b

October 8, 2026. Independent replacement reviewer `/root/reviewer_ch17` read
the complete coding skill, architecture, Chapter 10 contract, current live matrix,
retained-repair handback and both reviews above. Full voice and writing procedure
were reloaded. Prior grader, retained-adapter and source exposure remains disclosed;
this reviewer is independent of the student runtime and provider-support work.
The review used immutable Git objects, beginning with support
`065a6a824ec46a70affd85bd5b95cdc19439495d`, then the complete deadline correction
`9822b2b63257724001d7af195b377483baad3531` and final evidence
`78f4f6b14f9a5a89497470c1075262f39db93ce2`. Mutable preparation was not treated as
a handback. No credentials, provider requests, clients or Go commands were run.

Disposition: the concrete support is ready for the coordinator's bounded live
release. No unresolved support blocker was found. Root separately reports final
deterministic clearance in `213b56b`; this review does not replace that record.
The recommendation covers the existing A/B/C/D plan and its unchanged ceiling of
66 generation attempts and three discovery attempts. It is neither evidence that
a provider is accessible nor Chapter 10 acceptance.

The two remaining setup findings above are closed. A records the CLI's actual
raw-zero/effective-16 policy; its three-attempt process allowance is enforced by
the relay. B captures current policy one and requires an actual returned call,
effect/result and round-limit completion. Its browser support measures native and
service speech admissions and Page-owned queues/pauses at resume/reconnect and
replacement boundaries. The plan still requires real keyboard/pointer actions,
separate terminal detach, tab close and server restart, and retained acknowledgements.
Those observations are not inferred from silence or an identity-verifier pass.

The full provider, discovery, relay, driver, identity, binding and verifier sources,
their controls, and retained browser/public consumer were read. Provider keys stay
in the parent adapter; launched runtime children receive a loopback placeholder
and an explicit environment without ambient credential variables. Only the selected
settings field is returned. Fixed TLS origins and route/model checks precede
generation. No SDK retry, redirect follow, second discovery page, capability probe
or model fallback exists. Responses pass configured-key redaction before either
the child or receipt sink; a changed response is marked and cannot prove original
byte equality. Exceptions and arbitrary response headers are omitted. Synthetic
controls cover success, refusal, redirect, split/encoded key echoes, inbound headers,
request-body refusal and child-environment isolation. They do not claim to sanitize
arbitrary user secrets or every possible transformed credential encoding.

Durable locked admission occurs before transport, including failed, canceled and
timed-out attempts. The schedule retains row ceilings 6/4/6/6, step caps, 22 per
provider, one discovery per provider, 4,096 output tokens, 120 seconds per attempt
and 600 seconds per row. The final correction deducts TCP-connect time before TLS
uses its remaining timeout. DNS timeout leaves no worker capable of a later HTTP
send. The D seed remains two explicitly local exchanges; C records both Agent
outcomes independently before deciding whether continuation is possible. Deliberate
operator actions still inspect each result and stop on missing observations; unused
allowance does not authorize a rerun. No new spending or feature condition is added.

Current official documentation was opened to check discovery/header mechanics:
[OpenAI models](https://developers.openai.com/api/reference/resources/models/methods/list),
[Messages models](https://platform.claude.com/docs/en/api/models/list),
[Gemini models](https://ai.google.dev/api/models) and
[Gemini key carriage](https://ai.google.dev/gemini-api/docs/api-key).
OpenAI Docs was used for the OpenAI check. Model selection remains an actual
single-page discovery result; the exact requested Gemini target must be returned
with the required generation method. Documentation establishes no account access.

The retained outputs report 35 local controls at `065a6a8` and 13 affected provider
controls after the deadline correction. These overlapping sets were inspected,
not presented as 48 distinct tests or rerun here. The 22 actual-build identity
refusals retain their `065a6a8` association, with valid parents and specific
before-write refusals. Two final controls bind changed support and launch support
to `9822b2b`. Their launches are explicitly synthetic. The unchanged identity code
and narrow provider/test diff justify retaining the earlier scope.

The reviewer independently executed the frozen final identity preflight read-only
against the committed final binding and actual retained files. It passed: 196
current inputs, 191 historical build inputs, 24 module files and 24 support files.
The canonical binding SHA-256 is
`9f877c13fe2fd564779dc6305240928c3dc710071042ffb5181e99b6587b14a5`.
Runtime/build revision remains `57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4`;
interpreted support is `9822b2b63257724001d7af195b377483baad3531`. Actual hashes:

| Executable | SHA-256 |
|---|---|
| CLI | d43e5b309facf02416a14a1ab390dc44a48eac1749eddc565f5f0196619e62e5 |
| GUI | 57257823725494b87f19ba1ad3de824c671bcea73a2da37dc527af9d1eb369be |
| Public consumer | 7487d1e857310a968658b3da49c4a0c1ee3cab26b1c71b8c5268c9d88ab701a2 |

See the frozen [handback](../../solutions/edition-2/main/evidence/ch10/provider-preparation-handback.md),
[final binding](../../solutions/edition-2/main/evidence/ch10/provider-prep-binding-final.json),
[22 controls](../../solutions/edition-2/main/evidence/ch10/provider-prep-identity/results.json)
and [final controls](../../solutions/edition-2/main/evidence/ch10/provider-prep-final-binding-controls.json).
No old launch was rebound. A future compiled-input change requires its own build;
these interpreted changes require none. Live receipts must still demonstrate every
required A/B/C/D observation, retain failures and missing provider capabilities,
and precede historical comparison and final chapter acceptance.

Scoped whitespace verification passed. The existing external prose-lint executable
passed all hard checks on this review; its soft length, negation-density and
person-gap warnings reflect the accumulated technical review record. No linter
was compiled and no unrelated prose was changed.
