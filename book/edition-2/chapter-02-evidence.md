# Chapter 2 evidence and source reconciliation

Status: human-client validation reopened; prior contract validated at reviewed student checkpoint
`cc1bec45c3327c87728a4040f762155d8e860a0b`. Initial implementation and
all-three-provider CLI/public-consumer live receipts remain bound to
`39a92ca27a418712832ac0dcbbfbe4e32b3bca35` and its per-run source ledger.
Independent comparison, revisions, and acceptance are complete. Bill's
editorial approval is separate; strict blind-context certification is not
claimed for this pilot (see the qualification below).

## Human client correction

Bill identified that JSON-lines input is a machine interface rather than
human chat. Added the new contract in §2.7 before implementation: explicit
chat/protocol, terminal auto-selection, flushed prompt/readable answer,
local commands, history-derived redaction targets, exact line ceiling and
error/EOF behavior. Procedure now requires actual PTY interaction with all
three real APIs from Chapter 2. Existing API/protocol receipts remain valid
for their measured scope.

Human-mode initial checkpoint `56dacfad01f71f2a1b20d39bca15846edef41ddd`
now has real PTY receipts in `solutions/edition-2/ch02/evidence/ch02/human-chat`.
Author read FEATURES.txt, receipts.json, all three main terminal transcripts,
and each launch record with binary/source hashes. Each had three paid
requests, exact code-name recall/reversal, literal slash input and exit zero.
Input/write/read/output totals were 418/0/0/168 on Messages, 263/0/0/160 on
Chat Completions, and 212/0/0/894 on generateContent. Local EOF/redaction
controls are separately labeled; existing public-consumer receipts were
not rerun. The spin now begins with ordinary-text reproduction and an actual
abridged terminal transcript, followed by earlier machine evidence.

Coder reports independent review requested useful static redaction errors
and owner access for terminal helpers, with bounded revisions now awaiting
re-review. No final acceptance is inferred from the initial terminal success.
Runtime PTY evidence is macOS; automatic terminal detection is implemented
for macOS/Linux, with explicit chat required elsewhere. No cross-platform
runtime success or Bill participation is claimed. Scoped prose lint passes
all hard checks after receipt reconciliation.

## Reading

Read first-edition chapter 2's design/contract through §2.8 and exercise,
relevant first-edition grader fixture definitions and harness/checks,
the chapter-2 derived fact sheet's core declarations, selected git
history, the complete global review map, current architecture decisions,
the full current voice guide and rewritten chapter procedure. First-edition
implementation is evidence for authors only, not a student baseline.

## Historical receipts

- `5a7dfca`: deleting call-bound opaque replay initially scored 100. The
  old fixture lacked opaque call material and had foreign provenance, so
  correct withholding also made the test vacuous. The independent
  same-model Gemini fixture closed the gap without changing points.
- `8a85d5b` and `f2894bf`: Ref serialization/refusal/render/redaction checks
  and deletion audit. A loader test must not be satisfied by a later
  renderer rejecting data the loader incorrectly accepted.
- `ef155b1`: live use found collisions in default log paths, unclear CLI
  behavior, and demonstrations contaminating their own fixtures. New
  evidence must run in an isolated working directory with explicit logs.
- `29cd739`: corrected claim about Go map serialization. Standard JSON
  encoding sorts map keys; the nondeterminism risk includes iterating a
  map to construct ordered slices before encoding.
- Current grader constants allocate parity10 and Ref3/2/2/4/4. The old
  printed parity25 table and surrounding points commentary are stale.

## Later lessons absorbed now

Shared structures/interfaces and parent access follow new chapter 1.
Separate log facts, current context, request configuration, and credential
material. Record provenance at capture and retain per-model usage before
model switching can misattribute it. Keep actor identity distinct from
entry purpose so skills, redaction, and recall do not require changing the
meaning of an existing field. Recorded bytes cannot alias parser buffers.

The optional GUI is a separate module/public client stub in this chapter.
The reviewer recommends no live browser requirement for an explicitly
unimplemented transport, matching Bill's permission to stub the GUI.
If actual WebSocket behavior is implemented or advertised, it becomes a
feature requiring its own live user-path demonstration.

## Current official documentation, read 2026-10-07

Used the OpenAI Docs skill for OpenAI-specific verification, searched
official domains, and opened the relevant pages. These are contract
references, not proof the new implementation works.

- [Chat Completions request/response](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create):
  `max_completion_tokens` bounds visible and reasoning output;
  `max_tokens` is deprecated and incompatible with o-series. Current
  usage documents `cached_tokens` and `cache_write_tokens` inside prompt
  details. Do not assume the latter is a fictional legacy-fixture field.
- [OpenAI prompt caching](https://developers.openai.com/api/docs/guides/prompt-caching):
  current cache observability is available, but this chapter does not
  perform the later OAuth/caching investigation or make a cache-hit claim.
- [Gemini generateContent](https://ai.google.dev/api/generate-content):
  prompt counts include cached content; thoughts and candidate counts
  contribute separately to reported total generation tokens. No current
  deprecation claim is copied from the first edition.
- [Messages caching usage](https://platform.claude.com/docs/en/build-with-claude/prompt-caching):
  ordinary input, cache creation, and cache read are distinct reported
  input categories. No cached-input price ratio is asserted here.
- [generateContent thought signatures](https://ai.google.dev/gemini-api/docs/generate-content/thought-signatures):
  re-opened October 7 during author self-audit. Signatures can accompany
  visible text as well as calls and belong on their original part. The
  contract now retains text-bound metadata and supplies an independent
  matching/foreign fixture, while keeping thought-marked text out of the
  ordinary CLI answer. No performance claim or actual signature is invented.

No credential file has been read by the author. No provider call, timing,
current model selection, or live feature pass is claimed.

## Draft checks and next gate

Ran the prose linter on the initial chapter-2 contract: exit 0, no hard
failures. Soft warnings concern length and negation density in an
847-word working draft; no filler added.

The completed draft now prints the schemas, fixtures, commands, formulas,
adapter subsets, client boundary, and acceptance mapping. Requested route
and returned model identity are separate recorded facts. Usage keys the
response provenance. Opaque matching is exact against the configured or
explicitly resolved target; no learned alias mapping silently relaxes it.
Call-bound incompatibility fails instead of dropping a required signature.

The global reviewer is checking the full contract, especially failed-prompt
projection, single-owner replay accounting, request-time ephemera consumption,
and tool-only completion. Those were concrete gaps in the preliminary draft
and are now taught explicitly. The author reloaded the full voice guide,
procedure, workflow, and architecture after interruption, then the changed
procedure's mandatory first-edition comparison rule. Final chapter validation
requires that comparison and revisions after the initial student run.

Ran `go run ./cmd/lintprose book/edition-2/chapter-01.md
book/edition-2/chapter-02.md` after the complete contract draft: exit 0,
no hard failures. Soft warnings concern length, the long mechanism passage,
and Chapter 2 negation density; no invented person or padding was added.
Root's initial read caught the need to distinguish a pre-HTTP render failure
from an attempted request. §2.3 now states that input/configuration and
unresolved-call validation precede capture; a later render failure records
an error and discards pending input while leaving ephemera unconsumed.

The first full reviewer pass found three further contract gaps, now fixed:
the exact rendering of a URI inside a tool result, legal event transitions
and duplicate IDs, and the real user path for applying result redaction.
Result references now render descriptive MIME/locator text; a separate
human-attachment fixture exercises Gemini `fileData`. The reducer rejects
duplicate/unsolicited transitions while retaining legitimate assistant
continuation after tool results. The external executable consumer obtains
a real call, supplies a controlled result through the normal public append
path, redacts it, and takes another live turn. It uses fresh owned log
writers; no unstated writable-resume feature is required.

These are contract revisions, not claims that those demonstrations have
run. Reviewer resolution and the student's implementation/evidence remain
pending.

After Bill relayed CodeRhapsody's prose critique, reloaded the complete
voice guide and revised the chapter's entry: explain the reader's debugging
problem with a concrete recorded result before deriving log/context/request.
Added brief motivating mechanisms for failure projection and cache-token
normalization. The formal schemas/fixtures stay intact. Kept the settled
Chapter 2 GUI stub and edition map rather than importing suggestions that
contradicted Bill's explicit requirements. No invented live model identifier,
story, or acceptance result was introduced.

The reviewer re-read the completed contract and found its four material
issues resolved, including the signed-text amendment. Corrected the final
fixture pointer to name `history.log` and its result-reference variant.
The student's cold read additionally prompted explicit rejection of
agent/tool `message_received` records and a permanently faulted writer
after a failed log append. These clarify existing invariants without
adding a log-resumption implementation. Contract handoff is ready; actual
Chapter 2 validation remains pending the student build and evidence.

The reviewer also accepted the revised motivation and signed-text contract
after rereading, and the student confirmed its cold-read questions were
resolved. Root released Chapter 2 implementation. A final copyedit corrected
`scored100` to `scored 100`; it changes no fixture or requirement.

During Chapter 3 derivation, its inherited ordering fixture exposed a real
future dependency: a deferred human event occurs before an outstanding tool
result. Coordinator approved a narrow clarification, reviewed by the global
reviewer. The SAME append validator/reducer permits one deferred human event;
synchronous Submit still rejects that input before recording. Renderer waits
for all results, projects them before the deferred input, and leaves log
sequences unchanged. §2.8 now publishes literal bytes plus negative controls
for a second pending human and a premature response. This preserves the
inherited ordering check without a hidden loader-only rule or runtime mailbox.

The student's real Gemini declaration run exposed a field mismatch:
`parameters` rejected `additionalProperties:false`. Author independently
checked the official FunctionDeclaration reference on October 7, 2026;
`parametersJsonSchema` accepts an object JSON schema and is mutually exclusive
with `parameters`. §2.5 now names that field explicitly, preserving neutral
schema constraints. Student reports the corrected live run succeeded; final
receipt reconciliation remains pending. This is a live-discovered adapter
correction, not evidence that the earlier fake had validated the real surface.

## Initial live reconciliation

Author reread full voice/procedure, the durable feature ledger, all six
successful receipt summaries, raw provider diagnostic receipts, observer/
redaction consumer source, and requested/returned identities in actual logs.
§2.10 now records the actual CLI transcript, all-three usage totals,
ephemeral reconstruction limits, real declared calls with controlled result
ingestion, redaction, independent Agents, observer close, and model changes.
No tool execution or live browser transport is claimed. Source binding is
the initial student checkpoint above plus the per-run hashes/sequence in
`solutions/edition-2/ch02/evidence/ch02/FEATURES.md`; the corrected Gemini
path alone was live-rerun after the wire-field change.

Preserve the safely failed tool-mode request and dated capability limit as
an observed request failure, not a universal statement about a model family.
Local rejected-event diagnostics and fake GUI integration are separately
labeled. The inherited dump/render harness conflict is under coordinator
repair: the correctly refused unanswered-call log is not a reason to weaken
the new reducer/renderer rule. No final 100-point or validation claim is made.

Coordinator subsequently corrected that harness precondition: retain the full
dump check, then append labeled supplied results before replay. Reported
corrected student score is 100, with legacy reference and relevant mutation
checks passing; the broader regression and independent revision gate remain
pending. §2.9 records the 95-to-100 fixture correction without weakening the
unanswered-call rule or claiming final validation.

Post-run comparison teaching findings are incorporated: safe static
field/transition reasons in diagnostics; owned per-observer copies at public
boundaries without full-history copying to recover an internal scalar or
predicate; purpose comments for render-before-consumption and persistence-
before-observation. These are implementation-independent requirements,
not private method-name hints. Scoped chapter lint still has no hard failures.

Author read the corrected output in
`checkpoint-evidence/ch02-initial-grade-after-fixture.txt` (100/100), the
targeted legacy result in `ch02-legacy-targeted-after-fixture.txt`, and
`ch02-initial-check-manifest.json`, which binds checks and the original
student snapshot. These verify the newly printed grader-correction result;
they do not finish the still-pending comparison/revision gate.

## Evaluation-context qualification

The student reports no direct first-edition chapter/solution/grader-source
reads. However, inherited historical summaries prevent certifying Chapters
1–2 as strictly blind context evaluations. Preserve that limitation in the
cross-edition comparison. Per the reloaded procedure, subsequent students
start in fresh context without coordinator history and read only new chapters,
the new skill/architecture, and the preceding new solution. Historical sources
linked from the skill and author/reviewer research notes are excluded.

## Final reconciliation

The reviewer accepted the safe diagnostic reasons, efficient owner queries,
and explanatory comments, then accepted the live prose against actual receipts.
Coordinator/coder report the final clean student commit above, passing main/
GUI/consumer module checks, inherited 100, 44 independent acceptance cases,
the positive control plus ten detected defects, and the complete root legacy
suite. This closes the pending stages recorded chronologically above.
No paid rerun was made for the internal diagnostic/query/comment revision;
retain the original source chronology. The student's `SOURCE-EXPOSURE.md`
records the inherited-summary limitation alongside the preserved first answer.
