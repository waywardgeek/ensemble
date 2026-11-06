# Local Models implementation

## Scope and live evidence

Bill approved personal implementation of Chapter27 and its graders. No student
campaign, deployment, host restart, new download or paid-provider campaign is
included. Cadence remains on Astra. MiMo belongs in the ordinary model dropdown;
initial integrated acceptance will use a small test Agent, not a code-enforced
child-only restriction. Existing artifact grading remains70/100 and is unrelated.

Bill's Windows output establishes native Ollama0.40.15, an RTX4060 with8188MiB
VRAM, and `mimo-v2.6:9b` Q4_K_M. The loaded model reported5.6GB,100%GPU and4096
context tokens. Its advertised262144 maximum is not loaded serving capacity.
Buffered localhost Chat Completions returned a structured `add_numbers` call for
17 and25; a matching tool-result message produced42 and `finish_reason: stop`.
The Mac subsequently received version0.40.15 from the laptop's API. These are
human-run endpoint probes, not integrated Engine or streaming acceptance.

The earlier Qwen2.5-Coder trial produced tool-shaped JSON in ordinary content,
not a structured call. That must remain text. The earlier WSL/RTX4070 assumptions
were superseded by actual inventory. The restricted Windows firewall rule was
inspected individually; this is not a universal firewall audit or authenticated
public-service claim. No router forwarding was needed.

## First source slice: ownership and authentication graph

Common declares Ensemble-owned `ModelVendor` and `ModelAuthentication` objects.
Vendors own endpoints; endpoints own model offerings and weakly reference their
authentication binding. Offerings separate local ID, wire name, capability maximum
and configured serving context. `Vendor` remains the existing wire-codec enum;
Ollama's configured vendor identity does not become OpenAI merely because its
endpoint speaks Chat Completions. Constructors validate before publishing both
relationship ends. IDs are unique within their owning collection. Root shutdown
closes admission and removes these objects after Agent work.

Authentication is explicitly none, an owned API key, a reference to the existing
public OAuth owner/account, or a reference to the separate Codex owner. Renewable
bearers are not copied into models. Credential-owner removal detaches weak source
links; replacement owners do not silently reconnect them. Authentication removal
invalidates endpoint bindings without destroying the endpoint/model graph. Newly
owned API keys participate in existing diagnostic redaction. Endpoint addresses
reject userinfo/query/fragment. Credential-bearing destinations require HTTPS or
literal loopback, with the existing additional OAuth/Codex endpoint restrictions.

This is a foundation, not a parallel runtime. Agent selection, request rendering,
credential acquisition, child restoration, GUI configuration and public graders
are still pending. None of the new domain objects is used to send inference yet.
No installed binary, launcher, private credential or frozen endpoint was changed.
The separately queued accessible composer-label edit remains untouched.

### Verification

Evidence: `/private/tmp/cadence-local-models.co2hSe`.

- `domain-build.log` and `domain-tests.log`: clean offline build and initial three
  ownership/refusal/detachment groups pass.
- `auth-domain-build.log` and `auth-domain-tests.log`: clean build, four focused
  groups pass after authentication binding integration. Tests cover both-ended
  relationships, invalid and duplicate refusal, wrong-root rejection, separate
  offerings with identical wire names, capacity distinctions, recursive cleanup,
  authentication invalidation, replacement-source refusal and diagnostic redaction.
- `domain-regression-auth.log`:20 existing OAuth groups pass.
- `domain-regression-codex.log`:6 existing Codex groups pass.
- `domain-regression-model.log`:3 existing model-selection groups pass.

The last three groups protect existing paths; they do not certify new Model-object
selection. Tests use synthetic credentials and owned fixtures, not live accounts
or ET. No full shared regression or independent local-model grade is claimed for
this foundation. Two assertion descriptions were narrowed after execution without
changing their conditions: absent files do not prove every possible read, and graph
cleanup alone does not test Agent joining. A foreign-binding nested-lock possibility
was corrected during source review before testing. No failed behavioral test is
being relabeled as a pass.

## Next integration slice

Give Agent a checked selected-Model relationship and use an immutable request-time
view of that selection. Preserve between-turn admission, existing Engine ownership,
credential-source lifetimes, safe error narration and no-fallback behavior. Reconcile
model identity in child/peer persistence and working-context policies before exposing
choices in the GUI. A changed label must never retain the previous cloud endpoint
or payment authority. Then test real buffered/streaming request bodies and headers
with owned endpoints, develop the chapter grader and update the original teaching.

## Transport prerequisite: no implicit redirect or cookie authority

Engine previously constructed its HTTP handler with redirects/cookies enabled
for initial API-key funding and disabled for initial plan/Codex funding. That
launch-time choice is unsafe to retain once selection can change funding. The
handler now disables both for every inference route. Plan bearer resolution uses
the already-frozen configuration account, and non-success parsing uses that same
configuration rather than rereading the Agent. Existing bounded same-source401
renewal remains distinct from automatic redirect or generic retry.

New real-loopback tests send an actual307 to another owned endpoint and set a
cookie on an actual response before a second inference request. Pre-fix clean build
`transport-before-build.log` and execution `transport-before.log` reject both old
behaviors at their named assertions:0 passed/2 failed, executable exit1. The outer
collector exit0 only preserved that expected failure; it is not a test pass.
Post-fix `transport-after-build.log` is clean and `transport-after.log` reports
2 passed/0 failed. Final fixture cleanup uses Task.WhenAll so one task fault cannot
skip joining another. All endpoints/credentials are synthetic; no ET/cloud call.
The source diff and full final fixture were reviewed. Shared-suite job75 completed
with358 passed/0 failed and exit0 in `transport-shared.log`. This covers the final
runtime/test source; subsequent changes were prose only. No new configured-model
selection, GUI integration or independent grader points are claimed.

## Qualified offering identity and replay boundary

`ModelIdentity` is an immutable credential-free value of vendor, endpoint and
model configuration IDs. `Model.Identity()` derives it through actual parents.
The provider-facing model name stays separate. Provenance now optionally carries
this identity; matching requires both wire identity and offering equality. Legacy
three-field provenance retains its encoding and matches only legacy provenance.
A present invalid offering refuses rather than silently dropping the boundary.
This value does not resolve a route, grant credentials or restore a live object.
Request capture, selected-Model membership and restart resolution remain pending.

In the same evidence directory, `identity-build.log` is clean with zero warnings
and errors. `identity-tests.log` reports four passing domain groups, including
qualified endpoint-local IDs, value equality, cross-offering and legacy mismatch,
public codec round trip, and malformed identity refusal. `identity-shared.log`
reports358 passed/0 failed, exit0, against the final executable changes. The full
source diff and new constructor were reviewed; subsequent edits were prose only.
This is value/codec and existing regression evidence, not an integrated renderer
route-switch test, independent grader score or installed-application acceptance.


## Internal configured request projection (selection still pending)

Configuration now freezes capability facts, optional offering identity and serving
allocation. An internal Model projection follows actual parents under the graph
lock, validates the bound source, keeps Agent instructions/security boundaries,
and replaces route-specific settings. It contains no copied API key. Explicit
BillingMode.None omits credential headers and is identified as unauthenticated;
GUI account status labels it none and does not offer cloud-account controls.
Engine resolves configured keys/plan/Codex owners only through that binding. Legacy
configurations keep their original resolution path. Parsers capture offering
identity; streaming/media/default context policy use frozen capabilities. Recall,
compressor and security copies preserve the binding; unconfigured name overrides
refuse. No public selection API is exposed by this intermediate projection.

`projection-final-build.log` and `projection-gui-build.log` are clean. The three
focused groups in `projection-final-tests.log` pass. The new real-loopback group
checks no-auth versus a separate synthetic API key, absence of the old provider's
headers, configured stream request flag for an uncatalogued model, exact wire
name, response provenance, auxiliary copy identity, serving-context defaults,
detached vendor/name-override refusal and removed-binding refusal. Responses are
buffered JSON: this does not certify SSE conversations or configured OAuth/Codex
renewal. Those paths, selection membership, restart and per-offering persisted
context policies remain open. `projection-shared.log` reports359 passed/0 failed,
exit0 against the final executable source. GUI verification here is a clean build,
not an actual-page account-status test. Full projection, Engine diff and new test
source were reviewed. No ET/cloud requests, installation or new grader points.

## Configured child identity and restoration

The root now resolves `ModelIdentity` through actual vendor, endpoint and model
membership. Same wire names do not alias; removal at any ownership level makes
lookup refuse. Lookup does not grant credentials or substitute another offering.

Child registration preserves the legacy nine-field version1 format. Configured
children write version2 with required `Offering` identity fields, not secrets.
The save operation accepts a complete candidate Configuration so a later selector
can persist its identity before publishing it. Restore validates the immediate
set's identities and wire names before child construction, resolves each offering
again for request projection, and retains ordinary rollback. A legacy record under
a configured parent refuses rather than guessing its former route.

Focused tests use internally projected configurations; the new public model
selector and weak Agent membership are still pending. They exercise actual child
files, destruction and fresh-root restoration under a different legacy parent
route; pause and conversation survive. Missing configuration, wrong wire name,
null/missing/duplicate identity fields and ambiguous legacy restoration refuse
without a partial child or rewritten registration. Existing legacy child and goal
restart tests still pass. This is not peer restoration, configured credential
renewal, live inference, GUI acceptance or independent chapter grading.

Evidence remains `/private/tmp/cadence-local-models.co2hSe`. Initial lookup build
and four domain groups pass in `resolve-build.log` and `resolve-tests.log`.
`configured-child-build.log` and `configured-child-tests.log` record the clean
initial implementation build and three restart groups; four domain groups pass
in `configured-child-domain.log`. After the candidate-save signature change,
`configured-child-final-build.log` records a clean build, `configured-child-final-tests.log` records three passing
restart groups, and `configured-child-final-domain.log` records four passing
domain groups. `configured-child-shared.log` records360 passed, zero failed,
exit0 against the final executable source. No installed application or live
target was changed.

## Live configured selection relationship

Configured Agent construction now resolves and attaches its actual Model before
creating Engine. Model maintains reverse weak Agent membership; Agent teardown
detaches after owned work, and Model destruction clears selections without
cascading into Agent destruction. Reads validate the selected root, offering,
endpoint, binding and capabilities. Engine admission takes ingress then the graph
gate, validates before incrementing its send guard, and checks credential
availability separately. Graph destruction never calls into Agent ingress.

The focused membership test covers both-ended construction, same-ID foreign-root
refusal with rollback, sibling survival, detach/reattach, non-stranded send guards,
Model loss without Agent destruction, refused rendering/inference, and a real
History save after loss. Documentary persistence uses the last immutable
configuration through a private renderer path; it does not resolve a substitute
Model or grant inference. Configured child and transport tests remain exercised.

Two intermediate runs correctly caught shutdown regressions: configuration reads
first rejected the root's closed-admission flag, then authentication admission
validation rejected the same shutdown. Both happened during History saving, not
child restoration. The correction separates relationship validation, new-request
admission and documentary persistence rather than weakening the restart tests.
Logs remain under `/private/tmp/cadence-local-models.co2hSe` with the
`model-membership-` prefix. The final pre-review focused build passed five domain,
three restart and three transport groups. A subsequent atomic-admission correction
has a clean final build. The final shared run reports360 passed and one failed,
exit1: the existing rooted symlink-swap fixture throws DirectoryNotFoundException
in its recovery rename after checking directory existence. Its unchanged focused
Sandbox group subsequently passes4/0, exit0. This is non-reproduction, not a fixed
race or a full shared pass. No Sandbox code or assertions were changed.
`model-membership-shared.log` and `model-membership-sandbox-recheck.log` retain both
results. No public Model setter, GUI choice, peer/working-policy migration or
independent grader points are claimed.

## Per-offering working-context policies

Working-context values, status and saved rows now carry an optional qualified
offering identity. Legacy rows retain their four-field encoding; configured rows
require a strict identity when that field is present. Duplicate checks, replacement
and effective-policy lookup all include the identity. No legacy wire-name policy
is silently migrated onto a configured offering. A shared public value codec now
serves provenance and these settings, with absent versus malformed kept distinct.

Focused tests exercise real Agent mailbox updates, same-wire offerings on different
endpoints, selected status, targeted replacement, legacy-policy isolation and real
settings restart followed by sibling memory/recall saves. Codec cases reject null,
incomplete and duplicate-field identities and repeated qualified rows. The three
working-context groups and five model-domain groups pass; core and GUI builds are
clean. The GUI frame/form/rollback carry qualified identity, but actual-page
configured-offering switching is not yet tested. No dropdown or public Model
selector is exposed by this change. Final shared regression passes361/0, exit0.
The previous membership run's Sandbox failure remains recorded above; this later
passing run does not establish that the fixture race was repaired.

Evidence is under `/private/tmp/cadence-local-models.co2hSe`, with the
`offering-context-` prefix. No live provider, installation or restart occurred.

## Qualified peer launch identity

PeerLaunch now carries an optional ModelIdentity. Legacy registry rows retain
exactly their original four fields; configured rows add the shared strict offering
encoding. Creation inherits the configured offering, and restoration resolves all
launch identities and checks wire names before constructing any peer. It resolves
again for construction rather than treating the preflight as lifetime authority.
Legacy records under a configured host refuse route ambiguity. The registry still
records launch choices, not subsequent mutable Model selections or credentials.

The new real-file test constructs and restores a configured peer under a legacy
host with different routing and funding, checks actual Model membership and retained
conversation, and verifies no registry rewrite. Missing configuration, wrong wire
name, null/incomplete/duplicate identities, and ambiguous legacy records refuse.
A valid first row followed by a missing route uses two new peer names: neither
Agent directory may be created. This distinguishes prevalidation from rollback
that merely removes live membership after startup effects.

Initial and final strengthened-test builds are clean with zero warnings/errors;
all six peer groups pass. The final shared regression passes362/0, exit0 against
the final executable bytes. Evidence is under
`/private/tmp/cadence-local-models.co2hSe`, with the `configured-peer-` prefix.
No inference is requested by the new test. No new dropdown, independent grader
points, installation or live-target acceptance is claimed.

## Public configured Model selection

The existing selection mailbox now accepts a Model object. Between-turn and direct
Engine-send guards remain shared with the legacy string selector. A foreign graph
refuses before lock acquisition. Under ingress then the root graph gate, the new
path validates the candidate, saves its whole restart configuration, and publishes
both Model membership ends and the immutable configuration. An explicit valid
selection can recover from an unavailable old Model without any automatic fallback.
The string path retains its route-preserving contract.

Focused tests use two owned HTTP endpoints with the same wire model name and
separate no-auth/API-key bindings. Real requests check endpoint, exact headers,
retained conversation, busy-turn and direct-send refusals, preserved Engine,
context reset/default, foreign-root refusal, repeated selection, and explicit
recovery after Model destruction. Destroyed targets and removed authentication
bindings must leave the current membership intact. A real child-file destination
failure verifies save-before-publication, unchanged configuration/membership and
preserved saved bytes; repaired storage accepts the complete candidate identity.
No real account or live model is used.

The first command used the wrong focused flag (`--model-selection-tests`); this
harness ran124 unconditional baseline groups and skipped the named selection suite.
That exit0 is not selection coverage. The corrected `--model-tests` run has a clean
build and five passing groups. The final strengthened-test build is clean, all
five selection groups pass again, and the final shared suite passes364/0, exit0.
Evidence uses the `configured-selection-` prefix under
`/private/tmp/cadence-local-models.co2hSe`. No dropdown, actual-page switch,
configured subscription-renewal coverage or independent grader points are claimed.

## Qualified GUI choices and unavailable-selection recovery

Common now declares credential-free ModelSelectionStatus/ModelSelectionChoice
values. Agent snapshots the last selected identity, local configuration availability,
legacy compatible names, and configured choices under ingress then the root graph
gate. It reuses request-projection validation without credential-file reads, renewal
or network probes. Availability is neither server health nor account entitlement.
Live Configuration and Engine admission retain their existing refusal rules.

The usage frame carries qualified choices separately from wire names, with funding
and serving-context descriptions. The normal picker submits a strict offering
object to the existing set_model route; legacy string commands remain supported.
An unavailable current selection has its own disabled placeholder. Recreating an
object with the same IDs does not falsely reconnect the Agent. Explicit selection
still resolves the actual root graph and uses save-before-publication.

Focused domain checks pass5/0 with a clean build. Real socket checks cover two
same-name offerings, malformed-identity refusal, status after Model destruction,
reconnect and explicit recovery, without inference. Final actual-page checks use
headless Chrome, the visible picker and real websocket commands. The fixture
removes and recreates the selected identity; reload preserves unavailable status,
and a deliberate replacement succeeds. Exactly four selection commands occur;
reload does not submit another. Browser and host both exit0 after joined cleanup.
Final page evidence is /private/tmp/cadence-model-picker-final.FVwuz7. Earlier
page evidence /private/tmp/cadence-model-picker-page.1MgBbR predates the same-ID
replacement strengthening. Neither is installed-application or live MiMo evidence.

Initial builds caught a mistaken nullable Surface access, Engine type shadowing,
and an omitted fixture constructor argument; those runs executed no new behavioral
checks. Codec review also caught incorrect browser identity field spellings before
the first actual-page run. A tool-batch refusal executed neither its proposed edit
nor build; the subsequent explicit retry corrected the fixture. These were setup
and implementation corrections, not weakened assertions.

Core/GUI logs use model-status- and model-picker- prefixes under
/private/tmp/cadence-local-models.co2hSe. Final core and GUI builds are clean;
the socket test and five existing GUI regression groups pass. Final shared core
regression passes364/0, exit0. The final page run uses that GUI build; only prose
changes follow execution.
No host model-configuration interface, configured subscription-renewal acceptance,
independent grader points, installation, restart or live inference is claimed.

## Host configuration decision and strict loading foundation

Bill approved a host-owned strict configuration file loaded before Agent/fleet
restoration. GUI editing of endpoints/authentication is not part of that decision.
Reminders were explicitly re-enabled and confirmed; existing launch policy and
Astra funding must remain unchanged unless explicitly selected otherwise.

New Common immutable document definitions and Ensemble.LoadModelConfiguration
provide the bounded library boundary. Strict UTF-8 JSON has a 1 MiB limit, version 1,
exact required/optional fields, duplicate refusal and bounded collections. Endpoint
authentication fields resolve actual existing root-owned bindings; no credential
value, token path, store creation, renewal or network action is accepted from the
document. Capabilities are explicit; unknown optional capacity remains unknown.
The loader holds the root graph gate across construction and rollback. It never
replaces existing vendors or changes an Agent selection. Late failures destroy
only the new subtrees and detach their borrowed authentication memberships.

The final clean build and six domain groups pass, exit0. New tests exercise actual
relationships, separate 4096 serving/262144 advertised capacities, untouched launch
configuration, repeated identity refusal, malformed/duplicate/unknown fields,
invalid UTF-8 and byte limits. Valid-first/invalid-later documents test unknown
bindings, credential-bearing URLs, excessive serving capacity and duplicate models
or vendors. Existing objects, Agent selection and reverse binding counts survive.
No inference is requested. During review, malformed cases were changed to use fresh
vendor IDs so duplicate-existing-vendor refusal could not mask a parser defect.

Evidence uses model-config-{build,tests,final-build,final-tests}.log under
/private/tmp/cadence-local-models.co2hSe. The final strengthened tests are job290;
job288 is the earlier version. Full parser/loader/Common review is job291.
No fresh shared regression is claimed for this isolated new loader; prior364/0
belongs to the completed GUI slice. Startup file I/O, host binding construction,
public grader and real-target acceptance remain pending. No installed settings,
credentials, application restart or live provider changed.


## Host file startup integration

The GUI executable now accepts --model-config / EN_MODEL_CONFIG with command-line
precedence, a one-handle limit-plus-one bounded read, and graph loading before Agent
construction or fleet restoration. Only explicit host bindings are supplied: none,
and launch backed by the already selected payer. Missing API funding leaves launch
unavailable; --billing none ignores ambient API keys. The legacy initial
Configuration is not reconstructed or implicitly selected into the graph.

Evidence: /private/tmp/cadence-local-models.co2hSe/model-host-build.log is a clean
GUI-host build. model-host-verified/result.json records ten actual process/socket
cases: omitted option, none binding, API launch binding, no-auth environment option,
flag precedence, unknown binding, missing launch key, invalid UTF-8, oversize and
missing file. Valid hosts exit0, refused hosts exit1, all are joined without forced
kill. Refusals leave no Agent directory/event log. Valid logs contain no inference
request. Status retains the launch wire name, legacy identity and exact funding;
configured choices expose 4096 serving versus 262144 advertised context.

The initial run reached the socket but failed a test's invented funding enum-label
expectation. Actual funding is descriptive prose; the oracle was corrected and raw
frames retained before assertions. The next nine-case run passed, followed by the
final ten-case run adding the omitted-option control and initial funding check.
No product change was made for that fixture failure. Tests use only the actual host
and public websocket, not a host substitute or private API. No shared-core rerun
is claimed for this composition-root-only change; prior364/0 remains prior evidence.

Limits: no exact return-to-launch choice yet, no configured OAuth/Codex renewal
acceptance, no restored-fleet process test, no live ET call or installed changes.
The bounded read checks growth in code; these tests cover an initially oversized
file, not a concurrent-growth race or hostile special-file behavior. Next address
the explicit launch-return policy without making a second mutable route registry.


## Per-offering request policy

ModelRequestPolicy is an immutable credential-free Common value owned by Model.
Construction normalizes reasoning defaults and feature-required explanations before
membership publication. Configured projection now takes request options from that
Model, while Agent instructions, tool-round limits and security remain unchanged.
The host document accepts a strict optional request_policy object; invalid late
policies roll back newly constructed subtrees without touching existing selections.

Focused evidence in /private/tmp/cadence-local-models.co2hSe: model-policy-tests.log
and model-policy-document-tests.log each record seven passing domain groups with
clean builds. model-policy-model-domain.log, model-policy-model-transport.log and
model-policy-model.log record seven, three and five passing groups after projection
integration. New tests distinguish explicit high/off reasoning, local reset, return
projection, immutable normalization, malformed documents and late policy rollback.
These are library/projection checks, not a host-created launch return offering or
actual-page round-trip acceptance. Final shared regression passes 366/0 in
model-policy-shared.log; the GUI host build is clean in model-policy-gui-build.log.
Full policy/runtime/test and changed-prose review completed. Only prose changed
after the final executable build; no standalone pre-execution hash receipt is claimed.
No installed settings, provider request to ET, credential change or restart occurred.


## Explicit host launch return

The host reserves host-launch/launch/launch before loading the document, using the
actual selected binding and normalized ModelRequestPolicy. Initial legacy selection
is unchanged. A reserved-vendor collision refuses before Agent construction, and
keyless API launches expose no launch binding or substitute return offering.
No backup Configuration becomes a second routing authority. Explicit return has a
qualified identity; legacy opaque provenance is not automatically migrated.

Evidence under /private/tmp/cadence-local-models.co2hSe:
model-launch-build.log is the clean host build; model-launch-startup/result.json
records twelve actual startup/socket cases, all with expected exits and joined
cleanup, including keyless local-only startup and reserved-vendor collision.
model-launch-wire-final/result.json records seven assertions over three actual
public-websocket prompt turns. The owned provider observes launch Responses with
synthetic bearer, high reasoning, summary omitted and launch cache key; local Chat
with streaming requested, no credential and no old reasoning/cache key; and return
Responses with the original route/authentication/options restored. Earlier human
prompts and assistant answers remain in later provider requests. All turns complete,
exactly three requests occur, host exits0, provider thread joins, no forced cleanup.

The initial wire run passed. Review then replaced ineffective HTTPServer.timeout
with accepted-socket timeout and closed failed websocket handshakes; the final run
passes with those corrections. The full helper and host/test diffs were reviewed.
A post-execution source receipt records final bytes, not a claimed pre-run receipt.
No fresh shared-core run is needed for this host-only slice; policy shared366/0
remains prior evidence. Replies are buffered JSON even when streaming is requested:
this is not SSE proof. Messages-only deferred/block-cache options, configured
OAuth/Codex renewal, real ET usefulness and installed configuration remain open.
No live service, credentials, installation, restart or publication was changed.

## Configured Codex credential-owner transport

A new configured-model case uses the existing synthetic Codex grant-file writer
and HTTP-handler fixture. It selects through the public Model overload, sends real
Engine requests through the Responses SSE parser, and observes paired grant A then
B without changing body/session policy. Removing the bound owner and constructing
a replacement leaves the old binding unavailable: admission refuses before a third
transport request. Explicitly selecting a new binding then sends paired grant C.
The Engine disposes its owned handler. No live Codex endpoint or private store is read.

configured-codex-final-build.log and configured-codex-tests.log under the existing
local-models evidence root record a clean build and all seven Codex groups passing.
The first build failed because the test guessed a two-argument LookupModel overload;
the actual out-parameter API was used before any behavioral run. Full test and
registration diff reviewed. Production code is unchanged. This is configured grant
rereading/source-loss/recovery coverage with an injected HTTP handler, not public
OAuth renewal, real-network Codex acceptance, or a new full shared regression.


## Configured public OAuth inference recovery

The existing real-loopback issuer/inference fixture now runs its six scenarios
through explicit Model authentication bindings as well as the legacy route:
success, repeated401, forbidden403, API-key401, terminal refresh failure and
transient refresh failure. Stores use fresh synthetic signed identities outside
any live credential directory. Selection uses the public Model overload and checks
actual membership/qualified identity before send. Existing oracles remain intact:
one byte-identical inference retry at most, bounded refresh count, durable rotated
bearer, no API-key fallback, completion/error accounting once and no Engine-owned
History writes. Fixture disposal releases barriers and joins its server task.

configured-plan-final-build.log and configured-plan-tests.log record a clean build
and three passing focused groups, including both six-scenario recovery variants and
the existing delayed-401 check. Full changed-method/registration review completed;
production code is unchanged. This does not prove live entitlement, configured
owner loss during an in-flight refresh, or Chat-Completions tool streaming. No new
shared-suite pass is claimed; prior policy366/0 remains the shared evidence.


## Configured Chat Completions SSE boundaries

ConfiguredStreamingTests uses an explicitly selected, uncatalogued Model with
streaming capability, explicit no-auth binding and a 4096-token serving limit.
The legacy starting configuration has a synthetic key and disables streaming;
actual loopback requests prove neither setting leaks into the selected offering.
A provider barrier holds incomplete tool arguments. No ToolCalled, ResponseEnded
or effect exists until completion. The successful path executes one real custom
Tool, records its result before the second request, and sends the exact call ID
and result text. Both canonical responses retain offering provenance. Final usage
reports 31 input tokens including seven cached tokens; working-context status is
31, not 38. Exactly two requests and one effect are required.

The same fixture exercises immediate interruption and an explicit provider error
while the first call is incomplete. Neither permits effects or a final response.
Provider failure yields one controlled canonical FormatException stream error.
Both cases reuse the actual Agent for a fresh SSE turn, with no orphan tool call
or result in its request. Cleanup releases the barrier, stops the listener and
awaits both server and turn tasks before final root destruction.

Evidence in /private/tmp/cadence-local-models.co2hSe:
configured-stream-verified-build.log and configured-stream-verified-tests.log show
a clean build and all five model-transport groups passing against final source.
Earlier successful runs are retained; final review strengthened exact error and
orphan-replay oracles before the verified run. Full new fixture reviewed, including
failure cleanup. No production code changed and no new shared-suite result is
claimed. These are owned HTTP/SSE tests, not ET inference, coding usefulness,
independent public grader points, disconnect/malformed-stream coverage in every
shape, or cancellation of an already executing tool.


## Independent Local Models consumer foundation

The new testdata/local-models-consumer project references only Ensemble.dll, not
candidate tests or source projects. It drives actual Agent turns against owned
HTTP with randomized names, call IDs and payloads. Buffered and streamed structured
tools require a real callback, correlated result and canonical recording before
the next request. Ordinary tool-shaped content must remain text with zero effects.
Route, no-auth headers, configured stream policy, serving capacity, provenance and
cached input are checked independently of the candidate's test results.

public-model-final-build.log and public-model-result.json under the local evidence
root record a clean build and 43 named checks passing. An isolated archived core
changes only the no-auth transport branch to emit a synthetic bearer. It builds
cleanly and fails exactly buffered-no-borrowed-credentials-0 after three setup
checks; the loaded DLL hash matches the mutated build. Its path is recorded in
public-model-mutant-path. Good and negative executables exit0/1 respectively after
owned cleanup; the collector exits0 only because the intended negative was verified.

Full consumer source reviewed in bounded reads. Receipt collection is locked across
provider and driver tasks. The README documents the exact offline build/run inputs
and cooperative deadlines. This is a public component, not a scored Chapter27 CLI:
outer process timeout/containment, remaining lifecycle/configuration coverage and
report integration are next. No GUI/live Ollama/student/full-chapter acceptance or
new artifact score is claimed. Main runtime and installed application are unchanged.


## Named public runner and containment boundary

The public `--scope local-models -executable-only` command now builds the embedded
outside-assembly consumer against an explicit candidate directory. It requires an
existing .NET 10 SDK, clears inherited credential environment variables, uses fresh
home/work directories and offline NuGet sources, and validates all 43 expected
receipt names exactly once. Duplicate fields, missing failure/null fields,
unknown checks and trailing JSON refuse. No artifact score or chapter points
change. Current success remains ungraded/exit 3; behavior failure is exit 1 and
setup/API-build failure is exit 2.

The runner owns a process group, uses a two-minute setup and 75-second execution
bound, waits for the direct child, caps each output stream at 64 KiB and limits
inherited-pipe waits. Tests exercise a termination-ignoring child, a parent exiting
with an open descendant pipe, cleared environment, receipt omissions and invalid
CLI flags before any tool launch. This is not detached-descendant containment or
a network sandbox. Temporary files are removed after cleanup; report capture is
caller-owned.

Evidence in the existing local evidence directory: public-runner-final-unit.log
records both Go packages passing. public-runner-external.log records the actual
candidate integration pass. public-runner-cli.json/exit show probe_passed true and
ungraded/exit 3. public-runner-mutant.json/exit reuse the previously compiled header
leak candidate and fail exactly buffered-no-borrowed-credentials-0 with exit 1.
No mutant rebuild or new production change was needed. Full runner, tests and CLI
diff reviewed. Remaining independent lifecycle, configuration/restoration,
unknown-price and live usefulness checks remain open; no whole-chapter acceptance.


## Independent configuration, restoration and unknown cost

The public consumer now requires 56 named checks. Ten new checks use the strict
public document loader, actual binding memberships, public Model selection and
real child registration/History files. A later unknown binding rolls back an
earlier valid vendor without changing launch selection or existing membership.
A paused configured child is restored under a differently routed legacy parent,
using a fresh root and exact offering identity. Its conversation remains and no
request is replayed. Destroying the selected Model preserves Agent/conversation
while refusing live Configuration. A same-wire decoy endpoint cannot substitute
for the absent saved route; refusal leaves registration bytes unchanged.

Three additional checks inspect actual post-conversation Engine meters: response
identity, explicit no-auth billing and 37 input tokens are known, while dollar
cost remains null. No zero-cost estimate or guessed price is accepted. These are
fresh-root, same-process restoration checks, not process restart or crash recovery.
No private APIs, candidate test assemblies or inference in restoration phases.

The initial public runner integration passed all 56 strict receipts. An isolated
archived core then replaced exact endpoint lookup with first-endpoint selection.
Its build succeeded and the public CLI failed exactly at
child-missing-route-no-wire-name-fallback after positive restoration checks passed.
The working runtime was unchanged. Evidence is under public-restore-mutant-path
in the local evidence directory; the referenced directory retains build.log,
result.json and exit 1. Full new helper and runner/consumer diff reviewed.
Final public-configuration-final.log records both Go packages passing with the
real consumer integration enabled. public-configuration-cli.json/exit records
56 checks passing, ungraded status and exit 3. Only documentation changed after
that run. No new shared C# runtime result or installed acceptance is claimed.
