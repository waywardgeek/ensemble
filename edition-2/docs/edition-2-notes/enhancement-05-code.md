# Edition 2 Chapter 5 grader review submission

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Coder: `enhancement_coder_resumed`, October 10, 2026. Initial submission;
independent review pending. Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

Read the entire skill, carryover, original Chapter 6, `ch06.md` and
`ch06-code-review.md`, original harness/checks, accepted actor/media/application
code and relevant legacy paths. Scope is original Chapter 6 only. Original text
SHA-256 is `fc82efd7fc0caaa9805a7bfa7f2b8b457cfb096b341f0c8f92d60b672dff2495`
(recorded source revision `e1605916114230570b348ab5961c193fa1cebf80`). Accepted
reference is frozen `solutions/edition-2/ch06`, checkpoint `edition-2-ch06-r3`;
legacy is `solutions/ch06/agent` with its adjacent `ch05` and `ch06` applications.
No original, frozen source or product feature changed. No mocks or paid reruns.

## Contract and corrections

Seven checks retain 100 points: `ch4-parity` 10, `not-deaf` 25,
`replay-is-live` 15, `observer-fires` 15, `multi-agent-observations` 15,
`loud-refusal` 10 and `hub-clean` 10. These are the original seven weights;
parity uses all corrected Chapter 4 checks and the misleading `wake-once`
label becomes named sequential observations. Bill withdrew grouped completion
requirements; this change does not introduce them through another name.

- The separate workflow actually builds and runs from its own module root.
  The shared Chapter 4 `publicDependency` check establishes that Go resolves its
  public import to this submitted core, rather than another installed checkout.
  This extracts the existing small graph check, not an AST framework. Missing
  workflows no longer silently fall back to the ordinary CLI.
- `not-deaf` matches the author's hint payload before the final tool-result
  observation with the dispatched think call ID. A following captured request
  must contain that non-error result. The three-second handler can race the
  existing three-second callback; a running report is valid, not proof that
  execution terminated. Hint prefixes allow separating whitespace.
- Observer coverage requires an actual nonempty differing `from`/`to` transition
  and a text-typed final part containing the scripted response, not a delta,
  discriminator-only state event or nonempty but content-free part object.
- Three distinct named roles must emit ordered completion records and ordered
  application-log pipeline markers. These establish sequential progress only.
  Actual role tools, stored-draft handoffs, reliable private request replies,
  actor ownership, parent chains and common behavior placement remain mandatory
  source/integration review obligations. A hand-written app could imitate labels;
  the grader does not claim to establish the architecture from those labels.
- Replay now runs two real CLI turns and an actual `run_command` tool round,
  captures three HTTP request bodies and renders persisted prefixes strictly
  **before** each `request_sent`. It uses the same model, configuration and
  builtin registry on both paths; only surrounding transport whitespace is
  ignored. The old marker/count probe never performed this comparison. This
  fixture covers two turns and a tool result; the accepted source regression
  separately establishes mid-HTTP hint delivery and later consumption.
- Existing `render LOG` renders a valid local PNG through all three adapters.
  Checks inspect actual provider attachment fields, MIME and decoded file bytes,
  with surrounding text/order preserved. A valid 46-byte PCM WAV exercises
  unsupported audio; an actual image exercises unknown-model refusal. Errors
  name model and media category. Supported images prevent refusal-all passes;
  unknown-model text remains usable. URI/handle rendering remains deferred.
- `hub-clean` uses Go's `Standard` metadata rather than treating every dot-free
  import prefix as standard library. It preserves the original first-party leaf
  allowance and inspects transitive dependencies; corrected parity also checks
  full spoke topology. Source review still determines where behavior belongs.

Invocation is `go run ./cmd/grade-edition2 -ch 5 path/to/core`. `-workflow DIR`
selects the separate application; default discovery checks `core/ch06` then an
adjacent `ch06`. Chapter 4 consumer/case flags also apply. `-media-models FILE`
accepts a complete JSON map of nonempty `anthropic`, `openai` and `gemini`
identifiers. It replaces the defaults, rather than silently merging a partial
map. Defaults are historical reference fixtures, not current model discovery.
All application commands and public API names remain implementation choices.

## Actual results

| Reference | Original Chapter 6 | Revised full run before replay fixture correction | Final composed revised result |
| --- | --- | --- | --- |
| Accepted frozen original ch06 | 80/100 FAIL | 100/100 PASS | 100/100 PASS |
| Legacy original ch06 | 100/100 PASS | 25/100 FAIL | 40/100 FAIL |

Original grader runs used scratch source copies because its harness writes
`test_ch06.log` in the submitted directory. Accepted original failures reproduce
`ch5-parity` and text-only `loud-refusal`; the historical exceptions remain intact.

**Final composed means exactly this:** keep six actual checks from each full
revised run and replace only replay with its actual focused final rerun. Initial
replay reused Chapter 2's `claude-sonnet-5-fake` routing ID. Legacy rejects that
ID before HTTP, so its initial replay failure did not test replay at all.
Coordinator approved the original exercise/workflow's `fake-model` fixture for
both capture and render. Both references then passed all three byte comparisons.
No redundant full inherited sweep was performed. Component checks and this
composition method are recorded in `accepted-revised05-final-composed.json` and
`legacy-revised05-final-composed.json`; original full-run artifacts remain.

Legacy still fails corrected parity (10 points; including earlier fake-routing/
provenance and search/append requirements), `not-deaf` (25; no matching tool-result
final), `observer-fires` (15; missing full final content), and actual local media
(10). Its adjacent Chapter 5 app needs stdin and does not perform the default
no-argument consumer demonstration; this is recorded as unverified/failing,
not replaced with a new compatibility app. No consumer implementation was adapted.

A separate focused media run used **legacy's own declared** `claude-opus-5`,
`gpt-6-astra`, `gemini-3.8-flash` entries. Anthropic/OpenAI then explicitly report
blob rendering unimplemented, while Gemini refuses local paths and asks for URI.
Thus the real media failure survives removal of the historical identifier mismatch.
The mapped media check remains 0/10. It is not evidence about present provider
availability or paid access.

## Focused checks and samples

Final affected `go test ./internal/gradeedition2 ./cmd/grade-edition2 -count=1`
and matching `go vet` passed in the stable validation checkout. Tests execute
real accepted CLIs, actor, jobs, local files and separate applications; only HTTP
uses faithful provider fakes. Earlier Chapter 2/3/4 tests also pass, covering
shared launcher/dependency extraction. No whole-root or unrelated grader suite.

Accepted source race checks passed:

- `TestHintDuringHTTPReplaysAtNextRequest`
- `TestRequestsCompletePrivatelyWhenProgressOverflows`
- `TestObserverRoutesIdentityAndFinalParts`
- `TestInterruptKeepsAgentAndJobAlive`
- Separate module `TestWorkflowUsesRoleToolsAndEditedDraft`, exercising actual
  write/count/read/edit/review tools and checking the stored edited draft reaches
  the reviewer. This supplies the proof intentionally absent from the simple
  think/three-text grading fake.

All eight JSON fences in the enhanced chapter parsed. The exact published
prompt/hint fence and think-argument fence ran against the accepted separate
application and a faithful local Anthropic peer: four requests, hint observation
index 5 before the matching **tool-result** index 8, and author/editor/reviewer
completion order. The published model-map fence passed actual media rendering.
The Go sketches were checked against accepted source and remain schematic,
not claims that standalone snippets compile as complete programs. Module-root
workflow build and no-argument launch were exercised in scratch.

A scratch sample assertion initially matched the tool **call** sharing the ID
instead of the later result, producing a false failure. The raw transcript was
correct; fixing the analysis gave the indices above. No manuscript/runtime edit
was made for that false alarm. The exact workflow fence SHA-256 is
`aac41e44d7cf731aa455f9b06995f9f55e60623f55c440459853f4dbc133676e`.

Existing `ch06.md`, `ch06-code-review.md` and
`solutions/edition-2/evidence/ch06/` retain live human hint/interrupt/late-result,
Gemini role-tool/handoff and OpenAI image/audio-refusal evidence. These unchanged
live behaviors were not billed again. Earlier actor/media mutation records remain
valid for unchanged source guarantees and are not presented as new executions.

## Targeted mutations

Seventeen distinct scratch-source defects were caught by the intended checks:

| Defect | Observed intended failure |
| --- | --- |
| Drop hint acknowledgment payload | `not-deaf`: no matching hint |
| Final tool result uses another call ID | `not-deaf`: no correlated report |
| State observation omits transition payload | `observer-fires`: state false |
| Text final omits actual text | `observer-fires`: complete final false |
| Final text has an incorrect part type | `observer-fires`: complete final false |
| Omit reviewer completion record | `multi-agent-observations`: only author/editor |
| Omit editor pipeline marker | `multi-agent-observations`: missing middle stage |
| Offline render ignores loaded context | `replay-is-live`: all three bodies differ |
| Encode altered file bytes | `loud-refusal`: decoded bytes differ on all vendors |
| Misspell Gemini `mimeType` field | `loud-refusal`: correct attachment field absent |
| Allow unsupported audio | `loud-refusal`: no required refusal |
| Allow unknown-model image | `loud-refusal`: no required refusal |
| Omit model from refusal | `loud-refusal`: missing model identity |
| Omit category from refusal | `loud-refusal`: missing audio/image category |
| Workflow resolves another copy of core | All workflow checks fail before execution with wrong resolved directory |
| Common imports an acyclic dot-free internal spoke | `hub-clean`: `ensemble/internal/probeleaf` detected |
| Reject unknown model before rendering text | `loud-refusal`: text control fails (also incomplete image refusal) |

Each mutation changed real source in an isolated copy. Other workflow checks
stayed passing for its targeted observation/marker defects. Initial replay
mutation used a value where a pointer was required; that compile failure was
recorded as setup failure, corrected, and then killed by actual body mismatch.
The corrected replay mutant was rerun after the final routing-fixture change.
There are no surviving behavior mutants. Earlier shared topology and consumer
negative controls are retained from Chapter 4 rather than repeated wholesale.
Scratch source copies were removed after capturing patches/results.

## Isolation, size and handoff

An unrelated concurrent first-edition Chapter 25 edit briefly left `atomic`
undefined in `internal/grade/ch25_harness.go`. Fifteen attempted mutation builds
failed before execution; none counts as a kill. Those raw failures are preserved.
Per coordinator instruction, final validation and all behavior mutations used
an isolated committed checkout at `c78d217` with only the revised grader overlay.
The shared first-edition files were neither edited nor copied in a partial state.
An initial temporary-probe test selection also lacked its environment; that setup
failure is retained and the final complete affected-package command passes.

Raw local evidence directory:
`/var/folders/p_/11vk6y114xq7k9g5r_pc3yv80000gn/T/ensemble-enhancement05-0hyn_i25`.
Its `validation/` checkout is available to the reviewer. `validation-overlay.json`
records the base and SHA-256 of every submitted grader file; final verification
confirmed both the shared tree and isolated overlay match those hashes. Probe
source is scratch-only; no extra test runner or mutation framework is committed.
`mutations.json`, `final-focused.json`, full/composed reports, mapped legacy media
results and exact sample captures make the component evidence independently readable.
Temporary evidence paths are local review aids, not durable repository dependencies.

New production grader files: 450 active, 43 comment-only, 510 physical lines;
new regression: 24 active, 2 comment-only, 29 physical. Original Chapter 6's
harness/checks remain 523 active, 131 comment-only, 742 physical lines. Counting
uses nonblank lines excluding `//` comment-only lines; no block comments occur
in these new files. Shared files add 70/remove 15 physical lines, principally
CLI selection, the existing launcher's model parameter and dependency extraction.
These are separate totals, not a claim that the original grader was deleted.

Comments are below the roughly 20% teaching target (43/493 nonblank production
lines); useful explanations cover request cutoff/configuration, report-vs-execution,
actual payload correlation, positive media controls, module resolution and proof
limits. No quota padding or generic AST/runner framework was added. Exported
functions and added metadata fields have contract documentation. The added code
exists to observe artifacts the old marker/string heuristics never inspected;
independent review should assess that necessity alongside the count.
