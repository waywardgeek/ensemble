# Chapter 1 independent review

Date: 2026-10-07. Reviewer phase is independent of author and student.
Status: architecture, live prose, and post-run code-quality revision reviewed;
material code-review findings resolved at `75542c1c73388fa1ab618dbb8b1e252e3d80816a`.
Coordinator acceptance/audit remains a separate gate. This is not Bill's approval.

Comparison snapshots: first-edition Chapter 1 at course commit
`ce550e2a35b9a8cd07244c1caebdf01e49318735` (last change to its standard
`solutions/ch01/main.go`: `7eb3e24`); second-edition Chapter 1 at its own
repository commit `459e4ce`. The new student tree was clean when these
identifiers were recorded. Its initial attempt remains preserved in history.

## Reading and evidence

Freshly read in full: `book/voice.md`, `book/chapter-writing-procedure.md`,
the second-edition architecture record, coding skill, and saved global-review
map. The earlier full canonical Chapters 1–22 reading remains recorded in
that map; this phase does not claim all chapters are still resident in
context. Relevant canonical Chapter 2 sections were reread in the earlier
data-model review.

Read all production Go files and tests in the new Chapter 1 solution, then
the added transport test and separate consumer module. For the user-required
post-run comparison, read all of `solutions/ch01/main.go`. No legacy code
was copied or edited. Student access to the old solution was not required.

The whole new Chapter 1 manuscript was read in the initial proofreading
phase; subsequent changes, including the complete final live section, were
read against the actual sanitized `/tmp/ensemble-ed2-live-evidence.json`
and the saved consumer source. The initial Chapter 2 draft was incomplete;
its expanded contract was subsequently read in full and reviewed through
the signed-text provenance amendment. Its status is recorded below.

No paid calls were made for this review. A targeted identifier-rename control
ran the renamed copy's tests; routine behavioral suites were not duplicated.
Root's reported 23 acceptance passes and deletion results are
separate evidence, not checks independently rerun by the reviewer.

## Comparison with the first-edition standard

| Concern | First edition | Second edition and assessment |
| --- | --- | --- |
| Ownership and reuse | One executable package; caller carries conversation separately from Client. | Public library, Ensemble/Agent/Engine ownership, common parent interfaces, isolated Agent state. Required architecture is stronger and must stay. |
| Request/parse placement | Readable named private wire structs in the executable. | Request/parse free functions in `internal/llm`; common contains declarations only. Anonymous positional request construction could use named fields for safer edits. |
| Failed turns | User message appended before send; interactive mode continues after failure. | Valid pair and usage committed only after response validation. A public regression exercises retry after failure without contaminated history. |
| Usage validation | Missing fields become zeros; accounting occurs before answer validation. | Required pointer fields distinguish absence; reject negative/fractional/invalid usage; no failed-response accounting. |
| Transport | Client uses implicit global default transport and normal redirects. | Explicit per-Engine transport and redirect rejection. Review caught the initial nil Transport; student fixed it and added a regression. |
| Diagnostic safety | Raw HTTP errors and provider response bodies can appear in diagnostics. | Safe fixed messages avoid those disclosures, but currently discard useful timeout/cancellation distinctions; see requested revision. |
| History visibility | Direct mutable conversation slice passed around. | History accessor returns a copy and request assembly copies before append. |
| User interface | JSON-lines protocol plus interactive chat. | JSON-lines CLI and separate external consumer. Interactive chat is optional in the new contract, so its omission is not a regression. |

Inspected structural properties: no behavior moved into common; no sibling
implementation imports; CLI imports only the public library; no mutable
application globals; Engine stores common.Agent, Agent stores common.Ensemble;
parser failures follow Engine → Agent → Ensemble → logger. The redirect
callback is standard HTTP policy, not a closure compensating for a missing
parent. The logger and state owners match the chapter.

## Requested revision

At the time of comparison, `internal/llm/engine.go` converts all
`client.Do` failures into `model transport failed`. `parseResponse` converts
all first-decode errors into `malformed model response JSON`. A canceled or
timed-out response body can therefore be described as malformed JSON, while
an external caller cannot distinguish a canceled request through the original
context error identity.

The student was asked to classify cancellation and timeouts without exposing
raw URLs, authorization, provider bodies, or arbitrary error strings. Preserve
safe caller cancellation/deadline identity where practical. Keep the existing
parent logger path, no-retry rule, and success/failure accounting boundary.
Use targeted local tests; another paid successful conversation is unnecessary
for a diagnostic-only change.

Also requested: a concise WHY comment at the validated pair/accounting commit
boundary. Named wire-construction fields were suggested as a readability
improvement, not a new architecture requirement. The author was told the
diagnostic distinction should be taught without restoring unsafe raw errors.

Revision inspected and accepted: `requestFailure` classifies safe context
sentinels and network timeouts at the HTTP call and both JSON decoder sites.
It wraps only canonical cancellation/deadline values, never remote error
text. Named request fields and the validated-commit WHY comment were added.
The author teaches safe diagnostics and recognizable causes in §1.6.

The coder's six public-API cases check cancellation/deadline handling before
headers, in the body, and while waiting for EOF after the JSON object. They
also check error identity, sanitized logger output, and no failed-state
commit. The reviewer identified a potential false-pass risk if cancellation
happened before the decoder stage. Two disposable-copy mutations resolved
the immediate evidence question: removing the first decoder's classifier
failed `body/deadline`; removing the second failed `after-json/deadline`.
Both controls built and failed their intended stage, while the unmutated
suite passed. This proves sensitivity in the recorded runs, not exhaustive
schedule coverage. The initial fixture shutdown hang was corrected with a
test-owned release channel; production behavior was not weakened to fix it.

Durable details are in the new solution's `evidence/DIAGNOSTIC-REVIEW.md`
and `evidence/diagnostic-mutations.json`. Student-reported formatting,
vet/tests in both modules, and inherited grading passed after the revision.
The reviewer inspected the patch and retained evidence rather than repeating
the suite. No paid calls were made. Original live hashes remain bound to
`459e4ce`; `reviewed-source-sha256.json` identifies the revised source. The
local diagnostic revision is not misrepresented as a new live run.

## Prose and live receipts

Initial substantive proofreading findings are resolved: model discovery no
longer places the key in process arguments; the TL;DR names finite timeout
and configurable logger output; the multi-Agent live exercise is a separate
executable public consumer. The historic first-block grader hole matches
commit `835946f`.

The final live section matches retained observations: October 7 at 17:06 UTC,
three CLI answers `Silent Harbor`, `Silent Harbor`, `robraH tneliS`, then
261 input / 209 output, exit 0 and empty stderr. At 17:08 UTC the consumer
kept CORAL-271 and HERON-839 separate, with totals 144/118 and 144/28.
Its third Agent used deliberate local fault injection; the chapter correctly
does not describe that as a provider outage. Reproduction paths point to the
saved external consumer. Source hashes and final checkpoint binding remain
the coordinator's evidence task.

The reviewer reran prose lint after receipt insertion: no hard failures,
only length (3,701 words) and person-gap warnings. No padding or invented
anecdotes were requested. The subsequent safe-diagnostics teaching was
inspected against the accepted code revision.

## Source gate and identifier independence

The reviewer inspected the package checker and requested two corrections:
validate standard-library method signatures rather than names alone, and
resolve helper references rather than treating a shadowed local identifier
as evidence that a common helper is used. The coordinator fixed both using
`go/types` and added wrong-signature, shadowing, and valid-alias controls.
The reviewer inspected the revision and its ten fixtures. Their passing run
is coordinator evidence, not an independently repeated run.

The gate accurately limits its claim to same-module imports and common
behavior placement. It does not prove ownership, absence of mutable global
state, or the semantic purpose of helpers reachable from permitted methods.
Nested modules need separate checks; optional GUI dependencies also require
review across module boundaries. The manual ownership review above remains
necessary.

An actual positive rename control used a disposable copy of `75542c1`:
AST identifiers changed Ensemble/Agent/Engine to Coordinator/Session/Runner,
NewAgent to NewSession, Logf to WriteDiagnostic, History to
ConversationSnapshot, Commit to RecordPair, and parent to creator. Comments
and string literals were unchanged. The package checker passed, the renamed
module's tests passed (public package 0.319s; llm 0.005s), and its external
consumer built. The original solution was untouched. This shows that the
examined structural gate and public consumer do not require those spellings.

## Final motivation pass and Chapter 2 handoff

The revised Chapter 1 opening establishes the practical cost of adding
another Agent, changing configuration, and diagnosing a failed request.
Section 1.1 shows the parser's missing route to the logger before prescribing
the owner chain. The complete skill-read instruction appears before the
first code block and explicitly says loading text does not enforce a design.
This addresses the motivation and instruction-visibility review findings
without weakening the architecture or introducing a sacrificial flat build.

Chapter 2's expanded contract resolves the earlier specification gaps:
failed prompts remain in the log but leave request context; Engine remains
the single accounting authority; recorded request events consume ephemera;
tool-only answers, redaction, transport seams, and all-feature live evidence
have explicit contracts. Subsequent review resolved rendering failure before
request capture, result-reference wire mapping, illegal event transitions,
and a reachable external-consumer result/redaction demonstration. Text-bound
Gemini signatures are retained on their original part with exact provenance.
The final fixture pointer was clarified to name the history fixture and its
result-reference variant. The technical contract is ready for a student
handoff; no Chapter 2 implementation or live success is claimed.
