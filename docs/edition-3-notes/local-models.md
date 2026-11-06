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
