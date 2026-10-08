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

## Complete first-draft review

Reviewed manuscript and outline `349aa64565ebc50fdd3daa1076dc8b68872cbb91`,
with evidence-only correction `2076cc1c53e05819604f7ec2b375c93d47e16fab`.
The working files match those frozen versions. This continues the same independent
role disclosed above; it is full contract/prose review, not a student attempt,
implementation review or live acceptance.

The complete current voice and writing procedure were reloaded, including after
compaction, followed by the complete manuscript, outline, evidence and advisory.
Architecture was reread in full. Focused predecessor rereads checked Chapter 8
policy paths/persistence/wire semantics and Chapter 10 selection/creation identity;
the earlier explicitly scoped Chapter 2/9/11/12 contract reads remain applicable.
For the required story comparison, this pass additionally read complete
first-edition Chapter 15 prose. Embedded code examples in that chapter are
historical prose exposure; no historical implementation or grader file was opened.
No source under main was edited, no runtime or checker was built, and no provider
call occurred. A short interruption closed Chapter 9's separately assigned factual
proofread; it did not change this review's frozen Chapter 14 input.

### A1–A6 disposition

| Advisory | Full-draft assessment |
| --- | --- |
| A1: capability and strict compatibility | §§14.2/14.8 supply an explicit fresh v4 profile, preserved old ceiling/catalog, no hidden filtering/conversion, exact version/identity agreement and a separate standalone route. The old-session CLI policy selection still needs R3 below. |
| A2: policy authority | §14.3 keeps one Agent policy owner, strict complete v2, read-only v1 defaults, no-op at maximum, pre-write capability refusal and independent turn-limit capture. Disable cancels an uncommitted intent without resurrecting it; target changes preserve its admitted policy. No competing snapshot authority is introduced. |
| A3: intent and atomic completion | §§14.6–14.7 define first-valid admission, intact staged acknowledgement, whole-batch pairing, atomic cut/note, terminal refusal/cancellation, append-fault limits and immediate-busy public controls. The tool never waits for its own missing result. |
| A4: target and placement | §14.4 identifies the preceding newest represented eligible batch and accepted-response expiry, including renew/no-op/no-target. §14.6 preserves logical anchors with the H/S/N/P example and literal provider note shapes. Manual/primary/grant lifetimes and request observations remain distinct. |
| A5: bounds and represented history | §§14.5/14.7–14.8 give exact canonical accounting, checked fractions, strict thresholds, whole-batch overshoot, literal stub/ref behavior, encoded record/state limits and no counter reuse. Full-log reconstruction is explicitly stronger than snapshot-only reconstruction; no absent body is fabricated. |
| A6: compatibility and progress | §14.7 distinguishes local render validity from remote acceptance, preserves opaque provenance, records committed cuts, and offers truthful recovery. R1 below must clarify which semantic transitions invalidate a deferred selection. |

The owner plan is coherent with the architecture: Agent owns policy, Actor owns
ordered conversation decisions and its bounded candidate worker, Engine owns
transport/usage, Jobs owns jobs/limits, and SessionStore remains the single durable
path. A handoff is not a new Agent, a second journal or a skill transition. Optional
GUI and all MCP transports consume the same public seam; removing a result cannot
dispose a Connection, physical socket, channel or unrelated Agent. The Chapter 12
pause/independent-observer distinction remains explicit in the planned browser run.

The planned demonstration covers disabled preservation, actual eligibility/stub,
keep/expiry, both bands, handoff, public/two-Agent isolation, real browser controls,
restart and full-log versus snapshot-only history on all three provider paths.
Boundary faults are correctly separate from paid behavior. This is a feature
plan, not an approved concrete spending/launch matrix or evidence of implementation.
The runnable independent acceptance command is expressly pending publication
before student release, and remains a release prerequisite.

### Consolidated findings for one author revision

**R1 — Include semantic selection changes in the deferral key.** Location:
§14.7, “effective projection revision” (lines 495–501 of the frozen manuscript),
with §14.4 eligibility/protection. Root independently raised this same point.
A byte-identical projection can acquire newly eligible batches after an accepted
empty response, or lose keep protection after a later accepted response. Those
transitions change the lawful selection even if no rendered text changes. Define
the effective revision to include eligibility and protection/expiry changes, not
just body bytes or the event sequence. Retain the exclusion for bookkeeping,
Jobs facts and failed HTTP alone. Require reconsideration after a relevant
semantic transition and no repeat for an unchanged blocked candidate. This
prevents an implementation from treating a legitimate later cut as permanently
deferred or turning every request receipt into an automatic retry trigger.

**R2 — Publish the browser control seam or explicitly delegate its spelling.**
Location: §14.8's browser keep/handoff paragraph (lines 586–590), against the
strict command/ack/error protocol inherited from Chapters 7/8. Root independently
raised this same point. The draft prints context state but supplies no command
envelope for these new actions. Either print exact command, acknowledgement and
correlated refusal shapes, or expressly leave their spellings to the student
under a documented public equivalent. In either case retain usable-ID correlation,
strict unknown/missing/duplicate/type/size validation, exact integer identities,
owned committed acknowledgement/state ordering, no model request, immediate busy
refusal and draft preservation. The checker must adapt only where names are
actually delegated. It cannot invent a private protocol after the student starts.

**R3 — Make the CLI's disabled-policy recovery route runnable.** Location:
§14.2 lines 158–162. Chapter 8 introduces `--policy PATH` for the GUI command
and an optional public Agent policy path; it does not print a human CLI selector.
Chapter 10 adds session selectors but no CLI policy flag. An old-session CLI
reader encountering an enabled policy is therefore sent to an option the cited
teaching has not provided. Print the CLI spelling and a complete invocation with
an explicitly selected separate disabled policy file. If this adds CLI support,
teach it as this chapter's extension, including creation-only path/default/presence
semantics, rather than calling it an inherited CLI option. No new policy owner
or storage format is needed.

**R4 — Carry the concrete task through the selection rules.** Locations:
§§14.3–14.8; the outline's proposed small literal sequence crossing both bands.
The opening configuration repair makes the stake clear, but the middle becomes
more than four thousand words of rules without returning to that reader's work.
The H/S/N/P fixture makes placement inspectable; it does not show the practical
effect of the three-step selection. Add a compact explicitly illustrative worked
case tied to the configuration task: what result is stubbed, what complete batch
is removed, why the newest/kept evidence survives, and which decision/manual stays.
Use measured or transparently stipulated neutral sizes and label it as a contract
example, not a successful spin. A short return to the task at the handoff or
snapshot boundary would explain why the reader needs these distinctions. This
is a teaching/voice finding, not a demand to satisfy a linter quota or invent
another human incident. Also change the opener's “replaces old results with
references” to deterministic stubs that preserve existing references: a result
without a ref does not acquire one under the printed rule.

R1–R3 require resolution before this contract's student release. R4 should be
included in the same prose revision before complete manuscript review closes.
No further architectural redesign is requested.

### Historical story, voice and scoped checks

The new account preserves the old story's consequential failure: trying to
continue a confused chat by discarding its beginning can remove its goal. It
accurately labels eighty percent as the historical report; the old §15.9 contains
the attributed account, not a recovered raw interaction. The new draft does not
claim a current Gemini measurement. The loaded-manual explanation preserves the
old §15.4 lesson while recognizing that new Chapter 9 already corrected it.
Dropping the old superlative, breakfast timing, repeated cache figure, silent
clamp and unlimited-growth promise is justified by evidence and the revised
contract. No invented scene, heard-audio claim or new provider result was found.

The retained prose executable independently passes every hard rule at 5,488
counted words. It reports twelve negation forms and a 4,108-word person gap after
line 172. The negations largely state consequential compatibility/refusal rules;
cutting them mechanically would weaken teaching. R4 addresses the substantive
loss of a concrete task through the dense middle, rather than the count itself.
The author's ten parsed JSON fixtures and 24-byte canonical control remain
author-reported local checks; this reviewer read their printed forms but did not
run a new fixture parser or infer runtime conformance from them.

Frozen source hashes:

- chapter-14.md: `e66837e3d015e1f9ce863e0cb3b8458bb3a2249722fd00a401ad8b458e368fd3`.
- chapter-14-outline.md: `57b4d8ba1e50d14bd3167677885c145383d80e459f9567ea618eee1fc2f8eef6`.
- chapter-14-evidence.md: `0e27439bb711c2df02831707f02365bce8b38363aefc1f408305c8fee23d21d4`.
- First-edition chapter-15.md: `6c6d953e754f73b8bafd45a2f6009cab5dce9d7557c008e79fcc4084f27c2e79`.

Disposition: complete independent first-draft read, with R1–R4 sent together for
author correction and subsequent closure. No implementation, successful live spin
or final chapter validation is claimed.

## Grouped correction closure

Independently reviewed the complete author correction at
`b9515241160d41e4c126b441c1c03b1f1fe152f1` against the fully read draft above.
The full voice and writing procedure remained loaded, with no intervening
compaction or rule change. The three changed author files match this freeze.
Focused predecessor rereads additionally checked Chapter 7's ordinary message
limit, usable-ID/correlation/connection cap, Chapter 8's snapshot/policy/watch
ordering and Chapter 9's analogous safe-state observations.

R1 is resolved. The effective selection revision explicitly covers represented
parts/anchors, completion, eligibility, newest-eligible choice and protection
renewal/expiry. An accepted empty response can invalidate the deferral without
changing rendered bytes. Mere Jobs facts, request bookkeeping and repeated failed
HTTP do not. The key stays bounded and derived, with one reconsideration after
a relevant transition and no new durable event just to advance a runtime key.

R2 is resolved. Browser keep/handoff now have exact command shapes, correlated
safe errors and committed acknowledgements. Strict field/type/ID/size validation,
subscription, command reuse and connection bounds remain inherited. Busy actions
refuse immediately without losing drafts. Changed state is published through
the owned watch before acknowledgement; a no-op keep has null seq and no redundant
fact/broadcast. Handoff acknowledges a committed sequence, not a staged model
intent or successful socket write. Policy updates retain their own revision and
conflict protocol, and both relevant state observations precede policy_ack.
The new integer fields remain exact, and reconnect does not repeat an effect.

R3 is resolved. Human CLI `--policy PATH` is explicitly introduced here.
Omission preserves in-memory defaults without reading the GUI file; a present
blank value refuses. A supplied path resolves once at creation under the existing
Agent owner and file lifecycle, outside session identity. GUI defaults are
unchanged. The old-session invocation deliberately selects a separate disabled
v2 policy, omits the context profile and requires the original catalog/ceiling.
It neither overwrites an enabled policy nor converts an old session.

R4 is resolved. The configuration example now follows immediate stubbing,
results-band stubbing and calls-band removal while showing which decision,
manual and evidence survive. Its arithmetic is consistent on independent manual
inspection: T=20,000 gives S=200/R=2,500/C=1,250; 5,400 exceeds 5,000, leaving
2,300 after batch 11 moves; 5,978 exceeds 2,500, leaving 378 after removing
7 and 11. Omitting a redundant stub for the removed batch agrees with the exact
fact rule. The author records how its stipulated sizes were constructed; this
review does not present those padded values as real tool work. The later
snapshot paragraph returns to the reader needing original file evidence, and
the opener now promises stubs preserving existing references rather than
inventing locators.

A1–A6's accepted direction remains intact. No new competing policy authority,
automatic format migration, hidden archive, opaque stripping, MCP lifecycle
change or GUI dependency entered the correction. The complete diff leaves the
original ten JSON fixture blocks unchanged and adds six visibly valid command,
acknowledgement and error examples. Their automated parsing remains the author's
reported check; no new checker was written for this prose closure.

The retained prose executable independently passes all hard rules at 6,324
words. Its soft negation and person-gap counters still warn; the manual voice
assessment now finds the concrete repair task and its evidence present through
the selection and snapshot explanations. That teaching improvement matters more
than adding an invented incident or mechanically changing the counter.

Final reviewed hashes:

- chapter-14.md: `d50333c5a9af83507b09dd6a177a91088315d5e6e7ea52c8312b7579ef534039`.
- chapter-14-outline.md: `86ba3e497e36c6d2403ddd5caf69b3763ac4c25bf931e14610652d705a29ce15`.
- chapter-14-evidence.md: `f0c76ad9cd0d4e9d2ae5c8d0a66d8abc6bc220ccefca6422613510dea29ad6a8`.

Contract and prose review is accepted; R1–R4 are closed with no remaining
finding in this round. The independent runnable acceptance command and concrete
reviewed live matrix remain prerequisites for their respective student/paid
releases. Implementation, deletion controls, actual demonstrations, comparative
code review, student feedback and final checkpoint remain pending. This closure
performs no runtime, provider, checker-code or build work and validates none of
those future outcomes.
