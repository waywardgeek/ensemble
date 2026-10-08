# Chapter 14 advisory contract review

October 8, 2026. Reviewed outline/evidence freeze `9758e89`; this is design
advice before the full manuscript, independent checker or student implementation.
It is not contract acceptance. No runtime, provider or build work occurred.

## Role and reading boundary

This reviewer previously coded Chapter 8, prepared initial Chapter 9 and Chapter
11 checks, reviewed Chapter 13 prose and audited Chapter 9 live evidence. The
reviewer did not author Chapter 14 or implement its runtime. Reads for this
advisory: complete voice, writing procedure, architecture, Chapter 14 outline
and evidence; global-review authority, forward map, data recommendations and
contradictions; relevant new Chapter 2 opaque/redaction/render rules, Chapter 8
policy persistence, Chapter 9 material anchors, Chapter 10 selection/semantic
snapshot/commit/resume/bounds, Chapter 11 version extension, and Chapter 12
integration/request-observation/version rules. Truncated combined reads were
followed by focused reads. No first-edition chapter, implementation or grader
was opened for this advisory; the author's historical evidence account is not
independently verified here. Full story comparison belongs to manuscript review.

## Direction supported

Keep the existing Actor, Agent policy owner and SessionStore append path. A
recorded cut belongs to conversation state; policy chooses future cuts and
cannot rewrite earlier receipts. The renderer stays pure. Original log evidence,
current request projection, retained material and model usage remain distinct.
The outline correctly rejects a second journal, automatic migration, a hidden
summarizer, cache promises and opaque stripping to obtain a successful request.

The simplest boundary is an explicitly selected fresh v4 context-capable
session, with maintenance disabled initially. Keep v1/v2/v3 shapes strict and
unchanged. Context capability is creation identity; enabled mode/thresholds are
mutable Agent policy. A v4 session stays v4 when maintenance is disabled. An old
session cannot become v4 through a settings update. No conversion feature is
needed in this chapter.

The following issues need explicit text before student release. They are
consequences of the published predecessors, not requests for new Bill approval.

## A1. Old-format decoding does not by itself preserve usable resume

Locations: outline “Scope and inherited boundaries,” “Session semantic format,”
and public interfaces; Chapter 10 §10.7 and Chapters 11/12 session tables.

An old store binds the exact installed handler ceiling and complete frozen
catalog, including inactive definitions. Automatically registering keep/handoff
or changing a shipped skill body/grant can therefore break old-store resume even
when maintenance is disabled. Retaining its decoder alone does not meet the
compatibility promise. Root agrees this trap needs an explicit teaching response.

Teach fresh-v4 selection together with opt-in context-tool registration/profile
and a deliberate old-compatible catalog/ceiling selection for existing stores.
Do not silently remove handlers from caller configuration to make identity match.
Print the human CLI/GUI route for a reader whose default `.ensemble/session`
already exists: inspect it, continue with its compatible old configuration, or
choose a genuinely new directory and v4 profile. Selecting v4 on an existing
old directory must refuse without conversion or file changes. Specify the
corresponding public constructor/open options and standalone-log behavior.

Publish the exact v4 initializer/anchor, identity, checkpoint/state-version and
origin agreement rules. A unified v4 identity with explicit empty bindings and
integrations is teachable, provided nonempty integrations retain Chapter 12's
skill-mode restrictions. It must not add empty fields to old versions or turn
their previously refused request-observation field into an accepted extension.

## A2. Version the existing policy without making two settings authorities

Locations: policy-admission proposal and “Persisting policy”; Chapter 8 §8.3 and
Chapter 10 §10.7.

Prefer a strict complete policy v2 under the existing Agent service. An old v1
file supplies the inherited max-model-requests value/revision and default-disabled
maintenance. Reading it should not write a file or advance a revision. The first
actual accepted change can persist a complete v2 candidate under the existing
replace/apply/acknowledge rules. A no-op remains a no-op, including at max revision.
Unknown fields still refuse according to their actual version.

Separate the turn-captured request limit from the maintenance decision boundary;
do not accidentally make max-model-requests change mid-turn merely because new
maintenance policy is sampled before an attempt. Record the policy value/revision
used by a cut or intent. Specify whether a later disabling update cancels an
already staged handoff or applies only to new admissions.

Define the incompatible-policy case explicitly: enabling maintenance against an
old-format mounted store should refuse before the policy write. Reopening an old
store with an already-enabled policy file also needs a named refusal and an
explicit disabled-policy selection path. Silently treating an applied enabled
policy as disabled, or rewriting that file during session opening, would make
the displayed authority untrue. Session snapshots must not restore a competing
current policy copy.

## A3. Handoff needs an intent and a terminal disposition, not a circular wait

Locations: handoff/commit proposals and duplicate/interruption decision.

Use the first valid admitted intent in a batch. “Valid” must identify the stage:
arguments, grant and capability can be accepted before other calls finish;
final target/projection compatibility can still fail at commit. A malformed or
disabled call should not consume the batch's first-valid slot. Once an intent
occupies it, later handoffs in that batch receive a normal paired refusal even
if the original intent later fails. Every attempted call still consumes inherited
literal-next-call limits as required; none may wait for its own missing result.

The tool's immediate paired result should say staged and identify the intent,
not claim committed removal. After the whole batch is paired, one successful
durable fact installs both exact cut targets and the note and constitutes the
committed outcome. A nonterminal refusal/cancellation needs a durable disposition
so a later reader can distinguish it from still-pending work. No new tool result
or fabricated model message is needed to carry that disposition.

The least surprising interruption rule is to cancel an uncommitted intent after
normal pairing/cleanup, leaving earlier committed cuts intact. If root chooses
commit-after-interrupt instead, the chapter must say so before the exercise.
Checkpoints remain busy while an intent lacks a disposition; crash history with
an unresolved intent stays unfinished under Chapter 10. An append failure faults
the Agent under inherited rules, rather than promising a cancellation fact that
cannot be written. Public handoff should use the same validation at a settled
boundary and return busy immediately elsewhere.

## A4. Name the keep target and preserve logical placement after removal

Locations: eligibility, keep and ladder paragraphs; Chapter 9 §9.7.

“The eligible batch in the request” is ambiguous when a request contains many
eligible batches. Name its stable identity and exact selection rule, such as
the newest represented completed batch, and state whether keep with no such
batch is a no-op or refusal. Define when its one-round protection expires,
including failed HTTP attempts and an accepted keep alongside other calls.
Automatic eligibility must remain tied to accepted final responses, not merely
request_sent. The outline already makes that distinction well.

Removing a paired batch must not erase its logical placement boundary. Hints
and immutable manuals anchored there still need their original sequence order
without invented call/result placeholders or relocation to the request tail.
Print a before/after fixture combining surrounding assistant text, the removed
pair, H/S/P material and the new attributed handoff note. Specify the note's
position relative to material accepted during the held batch. Keep request
observations at their inherited tail, and never turn them into persistent dialogue.

## A5. Give byte targets operational bounds and a truthful snapshot promise

Locations: ladder, bounds and snapshot/replay acceptance paragraphs.

Publish one canonical neutral measurement, integer arithmetic/rounding for every
T fraction, exact threshold comparison, default/minimum/maximum T and exact stub
bytes. State whether result children, wrappers, refs and opaque fields count.
Use Chapter 10's lossless number canonicalization where that codec is selected;
a generic float64 JSON round trip cannot provide the promised measurement.
Require a Unicode/control-escaping fixture and long arguments that independently
force the second band. HTTP size, normalized tokens and billed cost remain
different observations.

Select whole batches. Prefer an explicitly protected newest batch and truthful
overshoot when indivisible/protected material prevents the target. Never split
a pair to satisfy a numeric promise. The note's proposed 64 KiB UTF-8 bound does
not replace complete encoded-fact sizing: escaped note bytes, selected identities
and wrappers all count toward the inherited 64 MiB record including LF. The
one-million collection cap and 256 MiB canonical semantic-state cap also apply
before publication. Counter allocation must be prechecked without wrapping.

Clarify reconstruction after snapshot-only import. A full original log can
reconstruct a request before its cut; reduced state that discarded those bodies
cannot invent that prefix. Retain whatever original facts the documented semantic
codec requires for requests it still promises to reconstruct, or expose that
such an earlier request is outside the imported snapshot's represented history.
Do not secretly keep a second mutable archive merely to claim both body removal
and unrestricted historical reconstruction. Test full-log past requests and
snapshot-only post-cut continuation separately.

## A6. Compatibility deferral must make progress without concealing damage

Locations: provider-limitation and safe-refusal proposals.

Publish the exact local structural/provenance checks that can refuse a cut.
Do not equate a locally renderable candidate with proof that a remote deployment
accepts its opaque context. Preserve same-target opaque bytes and call-bound
identity, including signed text/calls; do not introduce a new blanket omission
rule or silently change models, flags or endpoints. Existing foreign-provenance
render rules remain their own explicit contract.

For automatic maintenance, a visible deferral should permit the unchanged
projection to proceed if it is otherwise renderable. Define a bounded attempt
identity, for example base sequence plus policy revision and target provenance,
so an unchanged blocked candidate cannot spin inside one request decision.
Retry only after the relevant state/explicit policy selection changes, without
an accumulating queue. Explicit handoff reports a refusal and retains its old
projection; a remote rejection after a committed cut remains a real failed
request and does not roll history back.

Teach recovery that the interfaces can actually perform: inspect originals,
disable future maintenance where appropriate, or explicitly select a fresh
session. Disabling future policy cannot resurrect already removed projection
parts. Do not suggest “retry without cutting” for a cut already committed unless
a separate restoration operation is taught, which this outline does not propose.

## Reader path and next review

The through-line is sound: keep the objective and decisions while ceasing to
resend obsolete tool output. Retain a small literal example throughout the
chapter so the reader sees what survives before meeting counters and codecs.
The author's attributed historical anecdote and explicit refusal to recycle old
timing/cache claims are appropriate preparation; their primary story verification
remains open for full prose review.

Before code, settle A1–A6 into one coherent printed contract and a runnable
fresh-session demonstration. A bounded live matrix should distinguish actual
cut/stub/keep/handoff effects from deterministic malformed/interrupt/overflow
controls and remote refusals. It must not need a repeated paid prompt to discover
that its default session cannot select v4. This advisory introduces no checker
assertions, implementation requirement by private naming, or runtime acceptance.
