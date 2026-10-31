# Optional Edition 2 Chapter 24: the temporary Codex endpoint

## Current assignment and outline, October 11

Bill explicitly assigns root author, coder and reviewer roles for this chapter.
This is self-review, not an independent student evaluation. All four full docs
and `.agents/skills/ensemble-coding/SKILL.md` were read before drafting/coding.
TL;DR and plain-language sections were written before application changes.
The chapter is optional, ungraded, and has no length target. The earlier tentative
Gateway24/workaround25 placement below is superseded: this is Chapter24;
Gateway remains unimplemented and deferred. The Chapter16 team remains paused.

Through-line stake: Bill needs affordable repeated context to keep working in
Ensemble; a measured, explicitly selected route lets him continue while retaining
a path back to public OAuth. His supplied LinkedIn post supplies the opener,
case16440622, reported API-key15/15 versus OAuth5/15 experiment and frustration
at the support reply. There is no original chapter to enhance. Preserve the
specific incident and credit for opening subscription access; do not invent
internal executive reactions, traction measurements or a provider diagnosis.
The fresh4998-token paired probe is separate evidence. Correct the post's cost
illustration to the verified named model's rates, without weakening its point.
Bill explicitly requested criticism naming OpenAI for this chapter.

Teaching order: brief opener; optional TL;DR; plain-language cache economics and
route/credential pairing; practical launch/return instructions and actual results.
The striking fact is that every completed paired request answers OK while the
cache counts differ. No full authentication framework or separate grader is needed.

Ownership plan: Agent config selects codex and a file path; the composition
root constructs an auth-spoke reader owned by Ensemble and retained through its
Agents' credential providers. The reader has no mutable token cache and reads
account/bearer atomically from one file read per request. Engine handles route,
headers, redirect refusal and existing SSE/usage. Codex owns refresh. No new
worker, global, sibling import or credential persistence is introduced.

Live checklist before launch: actual CLI PTY with GPT-6 Astra/high; create and
test a small file through real tools; observe Responses completion and reported
cache usage; verify generated file independently. Local tests cover expired/
invalid credentials, redirects, unsupported models, account/token rereading,
no fallback, trace secrecy and usage. Existing Responses tests cover incomplete
streams and model-bound reasoning. GUI shares config/core; run its module checks.
No new paid API-key experiment is needed for the already recorded diagnostic.

## Implementation and self-review result

Completed by root under Bill's explicit combined-role instruction. No delegate,
independent reviewer, cold student evaluation or chapter grader is claimed.
The initial contract required the body cache key but missed Codex's session-id
header. Both initial terminal runs successfully wrote/tested/submitted code but
reported zero cached input (short2029/2080/2134; long8241/8293/8347). Inspecting
actual Codex source supplied the missing header; a fake-only test could not have
established that provider requirement. The short launcher's completion detector
missed the successful submit and waited until90s before EOF; the task itself
finished in about7s. The longer launcher observed actual final submission.

Source: local Codex revision `a9abdeaff1773ec6e867253ae59b19128077676f`,
`codex-rs/core/src/client.rs` `responses_session_id` explicitly describes cache
affinity, and `codex-rs/codex-api/src/requests/headers.rs` spells the HTTP header
`session-id`. An initial local edit used underscore spelling; it was corrected
against that source before the paid rerun. The final renderer and transport use
one stable Agent/model identity. No new sticky-routing state machine was needed.

The corrected actual human CLI used GPT-6 Astra/high, created triple.py, ran its
assertion and submitted LONG_OK. All3 Responses completed; input/cache totals
were8241/0,8297/8064,8351/8192. Exit0 in7.844s; independent execution verifies
triple(14)==42 and the session save exists. [Sanitized live receipts](codex-endpoint-live.json)
retain all three runs, request hashes, actual usage/output, binary identities,
and final production-source hashes. Final binary SHA256:
`e14cde3acbd9958fa437c81724fe280ef5c076b315369d57f8091e5137cfd407`.
This verifies a bounded cache benefit, not universal cache reliability or billed
savings. GUI live use, OS-keychain credential access and Codex token refresh are
not claimed. Codex remains the renewal owner; the GUI uses the same config/core.

Checks:
- Core `go test ./...` and `go vet ./...`: pass; final full core test30.864s.
- Optional GUI `go test ./...` and `go vet ./...`: pass; final GUI test7.184s.
- Focused `go test -race ./ensemble ./internal/engine -run
  'TestCodex|TestResponses' -count=1`: pass, including final auxiliary checks.
- Separate ch06 and ch07 example module tests/vet: pass. ch05 vet passes.
- Separate ch05 example `TestExternalToolLoop` fails an inherited logger-suffix
  assertion: it requires shout to be the final log line, but subsequent cache
  diagnostics follow it. The exact failure reproduces on unchanged predecessor
  `3bb747c30d168a0ce04fec8c857f317aa6c2f0a8` in an isolated archive. The tool's
  real output and timestamped root log exist. Accept this pre-existing example
  fixture failure for the optional chapter; preserve the file and failed result.
- Thirteen focused defects fail their intended assertions; see
  [mutation receipts](codex-endpoint-mutations.json). They cover route/account,
  instructions, explicit-cache omission, key/header, rereading, redirects,
  model guard, auxiliary instructions, explicit path, missing account and endpoint.
  Existing Responses tests supply unchanged completion, replay and normalized
  usage checks. These mutations do not prove the backend's cache policy.
- One mistaken `go vet ./...` invocation at the repository root encountered
  historical mixed-package probe files in scripts/edition2. It was outside the
  application module; the correct main-module vet is recorded above.
- Chapter hard prose lint passes at1384 prose words. Its length warning is
  deliberately accepted under Bill's short-chapter instruction. The source
  incident, separate experiments, prices and exact run counts were checked.

KISS/architecture review: this chapter adds148 active production Go lines,
194 physical lines and33 comment-only lines relative to its predecessor. Counts
exclude blank and //-only lines from active code; all changed Go files and four
new production/test files are counted by role (new production files are auth/
codex.go and engine/codex.go; new test files are ensemble/codex_test.go and
engine/codex_test.go). Tests add288 physical lines (263 active,
15 comment-only), including auxiliary coverage. No first-edition
counterpart exists to compare. The design reuses the current credential seam,
renderer, parser and actor. Auth behavior stays in auth; wire behavior stays in
engine; shared interfaces stay in common. The optional route and explanatory
comments are proportionate to the task. No SDK, mock, new worker or extra grader.

The chapter includes Bill's real complaint and its economic stake, the required
TL;DR/plain-language teaching, runnable launch/return commands and the actual
integration discovery. The diagnostic Sol prices are named separately from the
Astra live run. Self-review accepts this optional chapter with the inherited
ch05 example-test limitation above; editorial approval remains Bill's.

## Source checkpoint and reader snapshot

Application/manuscript source: `471c7ae5f8f745c1ebdda9c388254f165a68177c`.
Exact reader export: `solutions/edition-2/ch24/`, 184 ordinary files;
its sibling manifest binds every file to source tree `ada66c1572f1e5d22efb6ac59f23ab2b61dd03ac`.
All export hashes and package discovery in its five Go modules were verified.
The exported core tests/vet pass (core31.069s). Its GUI initially failed because
Playwright had not been installed in the new directory; the documented npm ci
installation resolved setup, and exported GUI tests/vet pass (6.940s). Generated
node_modules was removed afterward; all184 source hashes still match.
Initial snapshot staging flagged inherited blank-at-EOF bytes in two SSE fixtures
and gui/web/gui/renderers.js. Those three copied source files remain byte-identical;
only that whitespace warning was excluded for the export commit.
Checkpoint tag: `edition-2-optional-ch24-r1` (new; the historical original24 tag
is unchanged). No prior export, first-edition source or paused Chapter16 work
was changed. The optional chapter is self-reviewed, with the inherited ch05
example-test failure retained; this does not claim independent review or Bill's
editorial approval.

## Earlier brief (historical, superseded assignments above)


Coordinator brief, October 11, 2026. This records Bill's assignment and evidence;
it is not a finished chapter, reviewed design or implemented feature.

## Bill's direction

Edition 1 ends at Chapter 27. New chapters belong only to Edition 2. Apply this
change only to `solutions/edition-2/main/`; do not edit `agent/`, first-edition
solutions or frozen Edition 2 exports. Under the existing three merges, Gateway
is planned as Edition 2 Chapter 24; this first wholly new chapter follows as 25.
Gateway's Edition 2 implementation and review remain outstanding.

The cold open should explain why the transition to OpenAI ran into an economic
blocker: our public OAuth route answered successfully but reported no reuse of
the context we kept resending. Bill reports filing a bug with OpenAI. Using
Codex's endpoint is intended as a temporary way to continue while caching on the
official OAuth route is broken in this workflow. The intended destination stays
the official route once a matched probe establishes that it works.

## Evidence for the opening

Use the [probe record](../codex-endpoint-probe.md) and its raw sanitized receipts.
The original three-call Codex test reached 4,864 cached tokens out of 4,998.
The fresh comparison used identical request bodies across the two credential/
endpoint pairs. Both completed public OAuth calls reported zero cached tokens;
both completed Codex controls reported 4,864. The third request on each route
hit overload, and a bounded public retry also hit overload. Keep those failures
separate from completed cache measurements. No claim of a provider-wide diagnosis
or a third successful public request is supported.

For GPT-5.6 Sol, the checked ordinary/cache-read rates are $4/$0.40 per million
input tokens. That tenfold rate difference motivates the chapter. It is not a
measured tenfold total invoice: output and cold input still cost separately,
and the experiment did not observe subscription billing. Bill's report of the
financial impact should remain attributed to him.

The story's concrete sequence is available without inventing an incident:
successful answers concealed zero reported caching; an existing Codex sign-in
reached a different Responses route; the matched request reused most input
there; the public retest still failed to show reuse. Explain why a correct
answer was insufficient and why the workaround must remain replaceable.

## Implementation and authorship boundaries

No implementation or new chapter manuscript is complete. The existing book
team is paused after the Chapter 16 draft. Before later authorship, the author
and reviewer must hold all four governing documents in active context and reload
all four after every compaction. The new chapter must follow the implemented,
independently reviewed change, not invent a replacement authentication design.

Retain separate public OAuth and Codex credential/route identities, current
explicit billing/fallback rules and existing ownership. The successful probe
establishes inference and cache reuse only; production token renewal, tool and
reasoning replay, usage reporting and return to the public route remain work
to implement or verify. Do not turn these items into a speculative framework.
All coding work uses the selected GPT-6 Astra/high; the matched diagnostic model
remains GPT-5.6 Sol because it is the model in Bill's proposed test.
