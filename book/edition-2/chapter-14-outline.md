# Chapter 14 outline: Keep the evidence, shorten the request

Through-line stake: a reader needs a long-running Agent to retain the task and
its decisions while ceasing to resend obsolete tool output, without losing the
evidence needed to explain what it did.

Preparation, October 8, 2026. New Chapter 14 maps to first-edition Chapter 15
under the current global review. This is an outline for coordinator review,
not a student contract or implementation claim. Chapter 13 has contract
acceptance only. The actual Chapter 9 matrix is still being reconciled; the
author will return to that work at the next receipt boundary.

## Story and teaching order

The first-edition account records Bill restarting confused chats and describes
an earlier compress_context tool that let a model summarize away the beginning
of a conversation, including its goal. Preserve that consequence with explicit
historical attribution. The reported eighty-percent removal is Bill's account,
not a new measurement or a claim about today's Gemini model. The new mechanism
narrows a model's discretion to named tool material while retaining dialogue.

Carry the loaded-manual failure into the mechanism: erasing a load_skill result
used to erase the instructions while leaving tools enabled. Chapter 9 already
separates those lifetimes; Chapter 14 demonstrates why that preparation mattered.
The before-breakfast design story is optional, short and attributed. Do not
retain the old superlative title, an unqualified bounded-context claim, or the
old cache percentage as proof of this implementation's performance.

Proposed sequence:

1. Separate original evidence, current conversation projection and this request.
   A shorter request does not mean a smaller log or a recovered deleted file.
2. Show what survives: dialogue, enduring instructions, skill activations and
   a new working note. Distinguish purpose from executable authority.
3. Give maintenance one actor commit path, typed candidates and exact targets.
   A pure renderer neither chooses cuts nor changes counters.
4. Define completed batch eligibility, paired removal and keep's exact duration.
5. Explain the two tool-traffic bands and their byte accounting, with a small
   literal sequence that crosses both boundaries.
6. Introduce micro_handoff as a durable note plus a recorded projection change,
   without starting another Agent or pretending the note is hidden reasoning.
7. Carry the selected state through strict replay, snapshot/import and restart.
8. Let CLI, public consumers and browser show the same policy and actual cuts.
9. Demonstrate real tasks, retained output and exact historical reconstruction;
   measure request bytes separately from provider usage and answer quality.

The voice pass should return to the reader's original goal after the detailed
selection rules. Prefer a concrete kept note, a removed result and a recoverable
original over a page of general claims about memory. No new spin is written yet.

## Scope and inherited boundaries

Propose deterministic tool-result stubbing, paired call/result removal,
keep_tool_results and micro_handoff. The initial default preserves existing
context; automatic cutting requires explicit application/user opt-in recorded
with policy. This is the coordinator's October 8 working direction, not a new
Bill ruling or an inferred model-name capability table.

Do not add a summarizer model, memory files, recall, hidden-reasoning extraction,
provider-native context editing, inline tool declarations, log truncation,
automatic session migration or a second journal. Later chapters can extend the
recorded operations without changing their meaning. Dialogue and surviving
material can still grow; tool bands are a policy target, not a whole-request or
model-token ceiling. A newly eligible large batch can exceed a band target.

Preserve these predecessor contracts explicitly:

- Chapter 2's existing redact_result operation, literal stub/reference behavior
  and strict event validation. Extend the vocabulary deliberately for paired
  removal; do not reinterpret an old redaction as a different cut.
- Chapters 4–6's tool admission, independent Jobs lifetime, completed-response
  boundary, exact once-only usage and truthful interrupted/failed attempts.
  Removing visible calls never kills jobs, cancels remote effects or makes a
  removed call ID reusable.
- Chapter 8's Agent-owned execution policy and separately owned GUI settings,
  exact revisions, reject-rather-than-clamp input validation and atomic updates.
- Chapter 9's frozen primary/catalog/bindings, active grant union, immediate
  revocation, retained activation identity and completed-batch hint/manual
  anchors. A cut cannot unload a skill, revive an activation, remove a primary
  or grant a tool. Retired manuals stay unless a separately taught verb removes
  them; this outline proposes no such verb.
- Chapter 10's one append path, settled checkpoint boundary, immutable origin,
  no absent-history fabrication and fail-closed live resume. The log remains
  bounded by its existing storage limit, with visible refusal at that limit.
- Chapter 11's prepared connections and frozen aliases. No dynamic discovery,
  transport replacement or new connection starts while rendering/replaying.
- Chapter 12's captured observations stay in their request_sent receipt;
  historical reconstruction selects them once at the request tail. They are
  neither dialogue to compact nor a pending sample to reuse.
- Chapter 13's diagnostic speech journal remains GUI-owned. Cuts, reconstructed
  requests and replay are silent; removing a context part cannot erase audio
  evidence or turn a native end into a claim that someone heard it.

MCP remains replaceable through its public complete-message transport seam.
The actual GUI tunnel remains optional-module WebSocket code. Context policy
must work in a headless consumer and cannot depend on GUI tool names or sockets.

## Owner, data and lifetime plan

| Fact or operation | Owner and route |
|---|---|
| Current maintenance policy and its revision | Agent, through the existing policy interface; no Engine or GUI copy becomes authoritative |
| Ordered accepted context, pending batch intents, eligibility and kept-batch facts | Existing Actor/reducer state in llm over shared common declarations |
| Cut selection and handoff preparation | llm free functions with the responsible owner interface; candidate cannot append/apply itself |
| Durable commit and public append validation | Actor through Agent's existing append/SessionStore path |
| Tool schemas and JSON decoding | Registry/tools, reaching typed actor operations through the actual parent |
| Provider projection and compatibility | Existing llm renderer through its owner; Engine keeps transport and producing-model usage |
| Original jobs/output and pending one-shot limits | Jobs through Agent, unchanged by projection maintenance |
| Snapshot encoding/writes | SessionStore through Agent; owned capture of existing authorities, no second Context |
| Browser controls and safe display | Optional GUI consumes the public policy/context state; no core DOM or WebSocket import |

Retain immediate-parent interfaces and logger access in every helper. A selection
worker, if needed for a large retained context, belongs to the actor that creates
it and returns an owned candidate tagged with its base sequence/policy revision.
At most one may exist per Agent. Controls, job facts and interrupts stay usable;
a stale candidate is discarded before append and recomputed for a later attempt.
No worker writes state or holds the actor's lock during disk/network operations.
Prefer bounded indexes updated by the reducer over repeatedly rebuilding whole
historical prefixes or keeping a duplicate archive in memory.

The committed projection remains the only current conversation authority. Typed
common values should identify batches, selected call/result parts, action, reason,
policy revision and note bytes. Historical original bodies remain in the log;
live reduced state need not keep another copy merely to generate a stub.

## Candidate behavior for full-contract review

**Policy admission.** Propose disabled, ladder, and ladder-with-keep modes, with
disabled the inherited default. The modes are explicit choices, not claims that
a provider or model is competent to curate its evidence. A target in bytes uses
the historical T/100 threshold and T/8, T/16 band proportions only after exact
accounting is published. Recommend a named default T=400,000 and minimum 20,000;
reject invalid values instead of reviving the old silent clamp. The maximum,
policy-file format and effective-boundary rules remain decisions below.

**Eligibility.** Propose automatic removal only after the full paired batch was
included in an admitted request and a valid final response to that request was
accepted. A preview, request_sent followed by HTTP failure, provisional delta or
interrupted response does not establish that opportunity. Call this accepted
response eligibility, never proof of what a model read internally. Capture the
batch association as durable reducer state. A later replay derives it without
running policy again.

**Keep.** The keep tool targets the eligible batch in the request that produced
its accepted call, not other calls alongside it. It suppresses that batch's
one-round automatic stub decision; it does not grant permanent ladder immunity.
Its own small acknowledgement is excluded from immediate self-stubbing but
remains ordinary paired traffic for a later ladder or explicit handoff. A
duplicate keep has a documented no-op result. An unavailable/invalid keep must
not alter the batch. The full chapter must settle the exact return shape and
pending tool_limits consumption, preserving the literal next-attempt rule.

**Ladder.** Select complete closed batches, oldest first, with newest eligible
tool material retained. Crossing twice a band budget triggers one recorded cut
back toward its budget. Never split a call/result pair or alter surrounding
answer text to hit a numeric target. Publish whether the indivisible boundary
keeps or removes an oversize batch and measure that exception explicitly.
Use one exact neutral byte representation, not provider token estimates or
lifetime usage. The proposed measure is canonical JSON of the retained typed
tool parts, including wrappers, arguments, stubs and references. Print a literal
Unicode/escaping fixture and distinguish this policy measure from HTTP bytes.

**Commit.** Compute exact targets, material hashes/identities and candidate
projection. Validate the whole transition and supported provider rendering before
persisting it. A malformed public fact, nonexistent target, stale base, orphan
pair or unsupported local projection refuses with no mutation. A write failure
keeps existing terminal persistence semantics; no imaginary rollback or implicit
retry. Each committed fact describes its cut, so changing T tomorrow changes
future decisions only. Validation is strict during offline replay too.

**Provider limitation.** Preserve all inherited opaque provenance and byte
rules. Do not strip signed/thinking parts, introduce a beta flag or retry a
different route just to make a cut acceptable. A local compatibility failure
refuses before the cut; an HTTP rejection after a valid committed cut remains
an actual failed request with retained originals. The full contract must name
the visible safe refusal and user-directed recovery path. No local validation
can promise that an opaque provider deployment will accept the next request.

**Handoff.** The note holds bounded explicit working facts: current objective,
decisions, source/evidence locations, failed approaches and next action. It is
ordinary attributed assistant material with a dedicated retention purpose;
it is not a new system instruction. Propose one atomic durable fact containing
the exact paired-removal selection and note, after all results of the accepting
batch are paired. No note enters as an extra tool-result body. This explicit
operation may clear its own batch before another model request sees that batch;
the exception must be taught separately from automatic eligibility. Keep all
human/assistant dialogue, manuals, request receipts, call-ID history and usage.

The full contract must settle acknowledgement timing, multiple handoffs in one
batch and interrupt/write-failure behavior. A tool must not wait synchronously
on a batch whose completion requires that tool's own result. A staged intent and
later actor commit can solve this, but a staged acknowledgement cannot claim the
cut has already committed. Public control behavior should use the same candidate
validation, with explicit immediate-busy behavior where no safe boundary exists.

## Consequential choices before the full draft

| Decision | Options and proposed direction |
|---|---|
| Persisting policy | Extend the existing Agent execution-policy domain with a versioned complete shape, or add an independently owned domain. Prefer the existing owner and a taught version extension, with explicit reading/upgrading of old v1 rather than accepting extra unknown fields. Capture effective maintenance policy at the request decision boundary; record exact cut facts for replay. Do not make current policy session creation identity. |
| Session semantic format | A new common v4 could represent prior v1/v2/v3 identity variants plus context state; alternatively add a separately versioned optional context extension explicitly accepted by each outer version. Prefer one strict v4 profile for context-enabled sessions, retain exact old formats when not enabled, and require deliberate conversion/fresh selection before new facts in an old mounted store. Review migration cost before adding a conversion feature; new session selection may be sufficient here. No automatic format change. |
| Atomic cut representation | Extend redacted with a new typed paired level and explicit targets plus a separate note event, or use a single context_changed/handoff fact naming both the cut and note. Prefer the single commit for handoff, retaining inherited redacted unchanged. Exact targets, bounds and correspondence to current prefix are required, not merely a replacement snapshot. |
| Multiple handoffs / interrupted batch | Permit ordered notes and repeated cuts, or accept only the first valid intent per batch and return a normal refusal for later ones. Prefer first-valid only. A paired batch may complete after interruption; decide explicitly whether its accepted intent commits then or is canceled with a durable outcome. No hidden pending work survives a resumable checkpoint. |
| Safe refusal versus fallback | On a cut incompatible with known opaque/provider rules, either stop the affected attempt visibly or continue unchanged with an explicit reported deferral. Prefer safe refusal for explicit handoff; automatic maintenance may visibly defer while preserving old projection, subject to a bounded no-repeat rule. No silent opaque deletion or undocumented beta behavior. |
| Policy limits / indivisible batches | Need an exact upper T, bounded note size, selected-target count and raw fact size consistent with inherited 64 MiB records and 256 MiB semantic state. Propose a 64 KiB UTF-8 note and at most one million selected identities within those encoded caps; encode ranges or batch IDs to avoid duplicating every body. Decide whether a batch crossing a target is removed whole or retained with an honest overshoot. |

Coordinator review should settle these before student code or checker assertions.
They are design proposals, not hidden requirements. The strict session profile
choice especially must remain understandable to someone using ordinary default
CLI persistence; an opt-in that cannot be enabled on their chosen store needs an
explicit human path and a truthful refusal, not an accidental migration.

## Interfaces and acceptance plan

Provide typed public policy/maintenance operations and an owned safe state view:
effective policy/revision, retained tool bytes, last committed cut identity,
eligible/kept batch identities and any current maintenance status. Expose the
same bounded facts in human CLI and browser. A note preview is not an accepted
note; acknowledgement identifies a committed operation or honestly staged work.
Define any new slash commands before implementation; Chapter 9 /skills remains
read-only. Browser settings use their existing revision/conflict and draft rules.

Required local controls include exact thresholds and plus-one cases, UTF-8 and
large escaped bodies, both ladder stages, newest unseen/failed-attempt survival,
keep-plus-other-tool batches, failed results, duplicate handoffs, mixed text/calls,
hint/manual anchors, running jobs, redaction references, unchanged skill grants,
all session versions, stale candidate/public append refusal, snapshot-only origin
and changed-policy replay. Cut facts with invalid identities must fail before
state mutation, leave the owner usable where refusal is nonterminal, and never
be silently skipped. Test history reconstruction before and after cuts with
HTTP/MCP/browser endpoints disabled. Preserve same-target opaque data rules.

Use deletion controls with genuine positives: long arguments must force the
calls band, not merely a result stub; changed settings must leave past request
reconstruction identical; remove either half of a pair and demand failure.
Historical CH=15 checks remain a diagnostic, with incompatible old assumptions
identified in the evidence record. Publish an independent new command before
student release; do not grade unprinted private state field names.

The future live matrix covers every new feature through all three real provider
paths, human CLI PTYs, actual optional browser and a headless public multi-Agent
consumer. A bounded task compares retained original file/output with the next
request after actual stubbing, then proves keep and explicit handoff and resumes
from a settled checkpoint. One Agent's cut must not affect its peer. Exercise
an integration's recorded result and request observation separately, preserving
the logical channel and remote authority. Native speech is unchanged; any audio
claim still needs its own capture and scoped receipt.

Record exact source/binary/policy/session identities, request bodies, cut facts,
usage and outcome, including provider refusals or extra reads. Deterministic
fixtures cover rare faults and boundaries; they do not replace actual use.
No tokens-saved, cache-hit, quality or timing promise follows from a byte count.
