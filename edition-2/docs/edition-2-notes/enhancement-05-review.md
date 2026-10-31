# Enhancement review: Chapter 5 — accepted

Reviewer: `enhancement_reviewer_resumed`, October 10, 2026. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`, read entirely before this review.
Current carryover, voice/procedure and bounded enhancement plan govern. This
chapter enhances original Chapter 6 only, after Chapter 4 checkpoint `7b8f46a`.
Accepted at initial formal combined review, zero correction rounds after the
initial submission. The preparation history below remains intact. No new
runtime exception, feature or frozen-source change is required.

## Reading and teaching baseline

Read the entire 939-line original Chapter 6, complete `ch06.md` and
`ch06-code-review.md`, original grader harness/checks, accepted actor/Agent/
dispatch/mailbox, public root, external three-role application, media renderer
and actor/media tests. Compared the legacy actor, common mailbox/model and
external application. Original chapter SHA-256 remains
`fc82efd7fc0caaa9805a7bfa7f2b8b457cfb096b341f0c8f92d60b672dff2495`.

The GUI/framework import incident and the deaf dispatcher explain the two seams.
Preserve those motivations and the concrete reliable-reply/lifecycle reasoning.
The accepted implementation has one Agent-owned actor; Engine performs one
exchange. Workers return facts to the mailbox, while jobs own execution state.
Blocking Submit/Ask requests have private replies; bounded progress is not their
completion channel. Post during an active turn supplies a hint, consumed by the
next request, not by HTTP already in flight. Final observations include actual
tool calls/results, not merely text. Save precedes success acknowledgement.

Bill's ruling removes the stopped-actor interrupt and grouped-wake requirements.
Interrupt ends the active turn, preserves ongoing jobs and allows a new turn;
later real output remains recorded without a duplicate vendor result. Kill and
shutdown differ. Three peer Agents execute application-defined roles sequentially;
no spawning tool, grouped completion API or early streaming/GUI is authorized.
Local-path media is required here; URI/handle attachments remain deferred.

Independently checked retained live artifacts: Anthropic hint at
22:45:22.121672Z, interrupt at 22:45:22.172780Z, new COMET answer at
22:45:23.081200Z and real job report at 22:45:34.088930Z. The early erroneous
`job_finished` with running status stays visible as a later-corrected metadata
bug. Gemini actually wrote 59 words, counted them, added `night`, rewrote and
counted 60 before the editor/reviewer stages; the reviewer received edited text.
Role input/output totals are 5,695/1,157, 1,310/591 and 663/324, all cache zero.
CLI totals are 4,431/177. OpenAI's local image answer is `red` at 63/4, with a
separate explicit audio refusal. No new paid run is implied.

## Grader-plan review

Original grading remains accepted source 80/100 FAIL, legacy 100/100 PASS. Its
replay check only tests application markers/observation presence; its refusal
probe sends text and no media; its wake label measures distinct identities and
completion records. The no-dot stdlib heuristic also mistakes the accepted
`ensemble/internal/...` module paths for standard packages. Preserve useful
checks, not these overstated claims or withdrawn features.

Supported the coder's seven-check/100-point plan: full corrected Chapter 4
parity; a separately built workflow with scratch logs and no core fallback;
actual state/final observations and correlated hint/report ordering; correctly
named three-role observations; captured-request/prefix replay; real local image
and unsupported/unknown-media probes; Go metadata for clean imports. Optional
grader-only workflow and three-vendor model inputs avoid prescribing public Go
API names or a fixed model table. These are finite inputs, not a new protocol.

Preparation feedback: compare replay immediately before `request_sent`, with
matching renderer configuration/tools; state whether hints are part of that
probe or retain their stronger source-test evidence. A three-second `think`
and default three-second callback can race; distinguish its correlated report
from terminal execution completion. The simple think/text workflow fixture
cannot prove actual manuscript tools, stored-draft handoff or reliable private
replies. Keep the existing source/integration proof and limitations explicit
rather than infer those facts from role markers. No generalized adapter or new
feature is requested. Await final source/evidence before shared harness runs.

## Draft preparation feedback

Read the entire 1,093-line enhanced chapter and author note. The original
GUI-coupling incident, mailbox motivation, concurrency warning and deaf/listening
ending retain their purpose. Added ownership, request completion and shutdown
explanations connect directly to the motivating failure. The actual 59-to-60-word
correction gives the three-role exercise a concrete outcome. No prose blocker;
final evidence/contract reconciliation remains pending. Suggested two small
precision edits: identity tagging does not mean the entire single-Agent event
log is unchanged, and `Post` avoids execution/receiver waits but can briefly
contend on its mutex. These are preparation comments, not a formal correction
round.

Initial new grader source inspection found three narrow preparation concerns:
the workflow module was not yet checked for resolution to the submitted core;
the final-content predicate accepted any nonempty part map; and exact hint-text
equality was narrower than the published `hint:` acknowledgment prefix. Sent
these to coder/coordinator before submission, requesting existing metadata and
observable payload checks rather than a new framework. No reviewer harness
was run while the coder owned shared validation.

The author applied both precision edits; reviewed manuscript SHA-256 is now
`19a847585e17f658e930cce3512c22d0cd7b1e684511b7c8ae162756801a172b`.
Independent prose lint passed hard checks (before those two small changes,
6,601 words); negation density and the long technical stretch remain soft
warnings. The preserved motivating incidents and newly evidenced correction
carry the chapter; removing sound explanations to hit a density number would
not improve this bounded enhancement.

Source reinspection confirms the three grader concerns are addressed: reuse
the existing public-dependency check for the actual workflow build, recognize
scripted final text rather than an arbitrary map, and trim the acknowledged
hint suffix. The coordinator's additional fixture correction uses a valid
one-sample PCM WAV rather than labeling PNG bytes as audio. A provided media
map now explicitly replaces all three defaults. Final validation is moving to
an isolated committed-base checkout because concurrent first-edition Chapter
25 edits temporarily fail compilation; that setup failure is unrelated to
this pass and must not be described as a successful shared-tree test.

## Initial formal review and decision

Read the complete final grader implementation, shared launcher/dependency
extraction and CLI diff, regression, coder submission and reconciled author
note. All preparation concerns are resolved. The final observer check also
requires actual nonempty differing transition states and a text-typed final;
the payload/type mutations verify those requirements. The original seven
weights remain 100 points. Useful former marker/log checks now live under
the accurately named three-role check, while actual captured requests and
local attachments replace the false replay/media claims.

The final prose preserves the original argument and documented incidents.
Its added detail explains why the actor must remain responsive during HTTP,
why a job report differs from execution completion, why an interrupt cannot
discard real late output, and why request replies cannot use droppable progress.
The three-peer application stays sequential and external. No grouped-wake API,
spawning, early streaming or URI/handle rendering has crept into the exercise.
The original error and reviewer agreement over grouped wake are retained with
Bill's correction rather than rewritten into successful work. Historical live
figures match independently inspected artifacts; no new paid run is implied.

Final manuscript SHA-256 remains
`19a847585e17f658e930cce3512c22d0cd7b1e684511b7c8ae162756801a172b`.
Independent final prose lint passes all hard checks at 6,607 words; the stated
negation/person-gap warnings remain. All 17 JSON records across eight published
fences parse. Coder's actual exact workflow-fence run captures four HTTP requests,
hint index 5 before the correlated result at 8, and three ordered completions.
The initial sample-analysis error matched the tool call rather than the result;
the corrected analysis, not a manuscript change, resolves it.

### Actual scores and proof limits

| Reference | Original Chapter 6 | Full revised run before routing correction | Final composed revised result |
| --- | --- | --- | --- |
| Accepted frozen ch06 | 80/100 FAIL | 100/100 PASS | 100/100 PASS |
| Legacy original ch06 | 100/100 PASS | 25/100 FAIL | 40/100 FAIL |

Read the actual full/composed reports and focused results. The final totals
retain six checks from each full run and replace only replay with its actual
final focused execution. Initial legacy replay never reached HTTP because
`claude-sonnet-5-fake` was absent from its table. Using the original exercise's
`fake-model` consistently for capture/render gives both references three exact
request-prefix matches. This is an identified fixture correction, not a claim
that another whole revised suite ran.

Legacy's remaining failures are corrected parity, no correlated tool-result
final in `not-deaf`, incomplete final-part observations, and local media.
The separate mapped-media run uses its own declared model identifiers and still
finds Anthropic/OpenAI blob rendering unimplemented and Gemini requiring a URI.
That distinguishes real missing local-path behavior from a fixture-name failure.
The original 80/100 exceptions remain accurately documented; a revised pass
does not change historical results.

The simple workflow fake proves tagged sequential observations, including real
state and final payloads, not every application tool or private completion.
Reviewed accepted source tests and retained evidence separately establish
manuscript tools/stored-edited-text handoff, reliable request completion under
overflow, interrupt/new-turn/ongoing-job lifetime, and hints during HTTP replay.
The coder reran those affected actor race and external integration tests
successfully. Ownership, actual root use, parent chains, common behavior and
no mutable application globals remain source-review obligations. Neither an
emitted label nor the numerical grade waives them.

### Independent validation

Verified every one of the eight hashes in `validation-overlay.json` against
both the shared submitted source and isolated `validation/` checkout based on
`c78d217`. Independent commands in that checkout passed:

```sh
go test ./internal/gradeedition2 ./cmd/grade-edition2 -count=1
go vet ./internal/gradeedition2 ./cmd/grade-edition2
```

The package test command exercised the real accepted boundaries and inherited
Chapter 2/3/4 regressions affected by the shared helpers. `gofmt -l` returned no
files, and scoped whitespace checks passed. The unrelated shared Chapter 25
compile failure remains a setup failure, not a graded defect or a passing run.
No broad root sweep or unchanged frozen-source matrix was repeated.

Independently copied only the actual external workflow to scratch, retained its
public import of the accepted core, and removed `text` only from text-final
observation payloads. The real three-role HTTP/job run still passed `not-deaf`
(25) and `multi-agent-observations` (15); only `observer-fires` failed (0/15),
with `state_changed=true, complete part_final=false`. A temporary Go overlay
accessed the existing grader helper without editing shared source. The first
overlay used `/var` while Go resolved `/private/var`, so it reported **no tests
to run**; that is a recorded setup failure. Canonicalizing the path ran the
named test and produced the observed intended failure. Reviewer scratch source
and overlay were removed afterward.

Read all seventeen coder mutation outcomes: actual omissions/corruptions fail
their intended content, identity, replay, media, module-resolution or import
checks. Unrelated compile failures and the first replay pointer-type setup error
are preserved and excluded from kills. The final replay defect was rerun after
the routing correction. Existing actor/media and Chapter 4 controls are reused
with their original provenance, not relabeled as new runs.

### Simplicity, comments and comparison

The new grader production files measure **450 active / 43 comment-only / 510
physical lines**; its regression is **24 / 2 / 29**. Independently counted
nonblank lines excluding `//` comment-only lines; no block comments occur in
these files. The original harness/checks remain 523/131/742, not deleted code.
Shared changes chiefly add CLI selection, an explicit model argument to the
existing launcher and extraction of the existing public-dependency check.
There is no generic runner/AST framework or product-only grading API. The new
code is justified by executing and inspecting boundaries the old heuristics
did not actually test.

Comments are about 8.7% of nonblank new grader production, below the approximate
20% teaching target. This is recorded rather than rounded into compliance.
The explanations cover the non-obvious cutoff/configuration, report timing,
payload correlation, real positive controls, module identity and proof limits;
the remaining substantial code is direct fixture construction/decoding. All
new exported functions and metadata fields have useful contract documentation.
No thin ownership/lifetime explanation or missing exported documentation remains;
padding fixture code would not improve the teaching.

The frozen implementation comparison is unchanged from the complete historical
review read for this gate: accepted core production 3,205 active/841 comments/
4,303 physical versus legacy 4,883/1,595/7,153, with chapter growth respectively
809/444/1,377 versus 1,559/595/2,393. Accepted core tests are 1,601/145/1,801 versus
548/125/728 legacy; examples and GUI stub remain separately counted in
`ch06-code-review.md`. These are retained source-comparison measurements, not
new grader line counts. The meaningful improvement is one owned actor and
reliable replies instead of duplicate orchestration, with tests of real worker,
HTTP, media and public application behavior. The unchanged source need not be
recounted or reimplemented for this prose/grader pass.

**Decision:** accepted at the initial formal review. No unresolved blocker;
coordinator may checkpoint Chapter 5 before releasing the next chapter.
