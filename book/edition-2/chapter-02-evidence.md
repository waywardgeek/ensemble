# Chapter 2 evidence and source reconciliation

Status: executable contract reviewed and ready for student implementation;
no Chapter 2 implementation tested or live result claimed.

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
