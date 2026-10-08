# Chapter 14: Keep the evidence, shorten the request

A long-running Agent needs the goal longer than it needs the third copy of a
file. Removing both is an efficient way to lose the work.

The first-edition account describes Bill restarting chats after they became
confused. An earlier compress_context tool offered another exit: let the model
summarize and discard the beginning. His account says it removed eighty percent
of the conversation, including the original goal, with a poor summary in its
place. That is a documented historical report, not a measurement of today's
Gemini API. Its useful question survives: which material can disappear without
taking the task with it?

This chapter gives tool traffic a shorter lifetime than dialogue. The Agent
keeps the user's instructions, its recorded decisions and loaded manuals. It
replaces old results with stubs that preserve existing references, later removes
complete call/result pairs, and can leave an explicit working note. Every change
is recorded. A reader can
inspect what was removed instead of trusting a summary to be an adequate receipt.

This is the first full contract draft. Implementation, independent checks and
actual use remain pending in the [validation record](chapter-14-validation.md).
The design choices below are coordinator decisions, not new Bill rulings.

## TL;DR

Extend the accepted Chapter 13 source only when released. Read the architecture
ledger and the entire `book/edition-2/skills/ensemble-coding/SKILL.md` before code.
Keep the GUI in its optional module and the MCP message transport replaceable.

1. Add a creation-only context-capable profile. Sessions explicitly selected for
   it use strict version 4; old versions retain their exact formats and compatible
   catalog/handler selections. Maintenance starts disabled. No store conversion,
   automatic handler installation on old sessions or implicit catalog update.
2. Extend Agent policy with the strict complete version-2 shape below. Preserve
   version-1 reads without writing or advancing revision. Context policy is
   sampled at a request decision; the turn's model-request limit stays captured
   at turn start. A policy file cannot turn an old store into a new profile.
3. Keep one Actor/reducer authority and one durable append path. Shared values
   and parent interfaces belong in common, policy/selection/rendering behavior
   in the responsible spokes. Pure rendering chooses no cuts and does no I/O.
4. Automatically stub eligible large results, then reduce the two measured tool
   bands in whole batches. Eligibility requires a valid accepted response to a
   request that included the batch. Keep the newest eligible batch from ladder
   removal and report unavoidable overshoot. Byte targets are not model capacity.
5. Implement keep_tool_results and micro_handoff through typed Actor operations.
   Keep names the preceding represented batch. Handoff stages an intent without
   waiting for its own result, then atomically commits paired removal and a note
   after the batch closes, or records refusal/cancellation.
6. Validate exact targets, provenance, current authority and all encoded bounds
   before publication. Automatic local incompatibility visibly defers; an explicit
   handoff refuses. Neither path strips opaque bytes, retries effects or changes
   the provider route to obtain acceptance.
7. Preserve skill/manual/hint anchors, historical requests, usage, jobs and
   integration identity. Full-log reconstruction can recover earlier requests;
   snapshot-only import cannot invent prefixes its semantic state does not retain.
8. Expose owned current state and controls through public embedding, human CLI
   and browser. Exercise all three real provider paths, independent Agents,
   restart and actual GUI use; distinguish local fault fixtures from paid results.

From main, build `go build -o /tmp/ensemble-ch14-cli ./cmd`. From main/gui, build
`go build -o /tmp/ensemble-ch14-gui ./cmd/ensemble-gui`. The inherited diagnostic
`make grade-dir CH=15 DIR=solutions/edition-2/main` tests an older contract and
cannot establish acceptance here. The independent command is pending publication
before student release; §14.9 defines its required matrix. Public Go names and
private semantic-codec field spelling remain student choices. Printed wire
shapes, owner boundaries, selection rules and errors do not.

## 14.1 Three views of the same work

Suppose the user asks for a configuration repair. A tool returns 30 KiB of file
text. The Agent identifies one faulty value and explains its decision. Five
requests later, the explanation may still matter while those original file bytes
contribute little to the next action. A stub preserves the locator; a later
paired cut preserves the explanation and drops the call and result together.

The original log, the current conversation projection and a serialized request
serve different purposes. The log records what happened. Reduced state determines
what currently belongs in a request. A request receipt records exactly what one
attempt used, including consumed hints and sampled observations. Shortening the
second does not shrink the first or rewrite the third.

Chapter 9 already solved a related mistake. A manual stored inside load_skill's
result disappears when that result is replaced. Its tool could remain enabled
without the instructions that explained its use. Immutable skill material now
has its own activation and position. Context maintenance leaves that material
alone, including retired manuals. It does not load or unload skills.

Keep dialogue, enduring instructions, primary bytes, skill entries and working
notes. Only typed tool_call parts and their matched tool_result parts participate
in this chapter's removal. Preserve surrounding text and standalone opaque parts
in their original order; drop a provider message only if no parts remain after
ordinary rendering. The inherited provenance checks still apply to what survives.
Removing a call includes its attached call-bound opaque value; it never authorizes
detaching that value and replaying it beside another call.

There is no new summarizer, memory file, recall operation or hidden archive here.
Dialogue and manuals can keep growing. Nor is a referenced artifact guaranteed
to remain available: moving a session does not move its relative files, and a
recorded job handle is not a live process after restart. A reference explains
where bytes came from; it cannot promise that rerunning an effect recovers them.

## 14.2 Select capability before opening the session

An old store fixes its installed handler ceiling and whole frozen skill catalog.
Adding two tools or editing an inactive manual changes that identity even if
maintenance is disabled. A decoder that recognizes the old version cannot repair
an incompatible caller configuration.

Add a creation-only public context-profile option, false by default. CLI and GUI
spell its true selection `--context-profile`. It installs keep_tool_results and
micro_handoff in the application's selected ceiling and marks the Agent capable
of this chapter's facts. Normal selection leaves the preceding ceiling and
catalog bytes unchanged. The flag neither enables maintenance nor grants a skill
permission to call either tool. Plain Agents expose their selected handlers;
skill Agents require the ordinary grant union. Model calls still pass the same
Registry permission check as every other tool.

Keep the preceding shipped catalog available unchanged. If supplying a catalog
that offers these tools, place it in a separate `skills-context` directory and
document its exact primary/loadable selection. Do not overwrite `skills` to make
the new demonstration convenient. Public callers may supply their own frozen
catalogs. No implicit removal of configured handlers makes an old resume pass.

| Selection | Store and capability |
|---|---|
| No profile, no MCP/integrations | Existing strict v1 |
| No profile, MCP only | Existing strict v2 |
| No profile, nonempty integration policy | Existing strict v3 |
| Explicit profile, new session directory | Strict v4, maintenance initially disabled unless an explicitly selected compatible current policy enables it |
| Explicit profile, existing v1/v2/v3 store | session_incompatible before file changes; choose a new directory or the exact old configuration |
| No profile, existing v4 store | session_incompatible; caller must deliberately select the recorded capability |
| Fresh standalone constructor with profile | Exclusive standalone log; context facts allowed under current creation configuration, no automatic session import |

Before experimenting, `session inspect .ensemble/session` explains an existing
store without a provider key. Continue it with its compatible catalog, ceiling
and disabled maintenance policy, or select a genuinely new directory:

```sh
/tmp/ensemble-ch14-cli chat --context-profile --session-dir .ensemble/context-demo
```

The inherited CH02_LOG/LogPath selector table still applies. A simultaneous
standalone log and explicit session selector refuses. SetConfig cannot change
profile, DataDir, base identity or ceiling after construction. The two opt-in
tool input schemas are exactly:

```json
{"type":"object","properties":{},"additionalProperties":false}
```

```json
{"type":"object","properties":{"note":{"type":"string"}},"required":["note"],"additionalProperties":false}
```

They belong respectively to keep_tool_results and micro_handoff; byte bounds
are enforced after ordinary decoding, not through a Unicode-character maxLength.
Use stable descriptions explaining previous-batch protection and staged handoff.
A current enabled policy against a non-capable Agent refuses
`context_profile_required`, including startup, before changing session or policy
files. Never silently reinterpret an enabled applied setting.

This chapter adds `--policy PATH` to human CLI chat; Chapter 8 supplied that flag
only to the GUI. For the CLI, omission keeps the in-memory default policy and
reports persistent:false, without looking for the GUI's file. An explicitly
blank path refuses. Resolve a supplied path once against the launch workspace,
without changing cwd, and pass it to the existing Agent policy owner. The path
is creation-only, remains outside session identity, and cannot change through
SetConfig or a browser command. It may accompany either the session selector or
the explicit fresh standalone-log route. GUI selection retains its existing
default `.ensemble/agent-policy.json`; both clients use the same file lifecycle,
directory creation, same-owner/path refusal and strict policy validation.

For an old session requiring a separate disabled policy, create a new file
`disabled-policy.json` containing the complete v2 defaults printed in §14.3,
then run:

```sh
/tmp/ensemble-ch14-cli chat --session-dir .ensemble/session --policy ./disabled-policy.json
```

Keep that session's exact old catalog, primary, bindings and handler selection;
omit --context-profile. This selects another policy authority deliberately,
without rewriting the previously enabled file. A missing selected policy file
also supplies disabled defaults and need not be written until an actual change.

## 14.3 One policy owner, one conversation authority

Agent owns current policy and creation configuration. Actor in llm owns serialized
conversation decisions, pending handoff intents and derived batch indexes. Engine
still owns transport and producing-model usage; Jobs still owns workers and pending
one-shot limits; Skills still owns immutable material and current grants. Context
selection reaches them through Actor's actual Agent parent.

Tool schemas and JSON decoding belong in Tools. Tools asks typed owner operations
to keep or stage; it does not mutate conversation slices. Store writes follow the
existing Actor→Agent→SessionStore route. A bounded selection worker, if needed,
has its creating Actor as parent, returns owned candidates and can reach logging.
At most one selection worker exists per Agent. Fence stale base/policy candidates;
never hold a control lock across disk, provider or MCP I/O. The GUI displays the
public state, rather than owning a second context or calling internal reducers.

Extend the existing Agent policy domain with this complete version-2 file:

```json
{"version":2,"revision":0,"policy":{"max_model_requests":0,"context":{"enabled":false,"target_bytes":400000}}}
```

All fields are required. Context contains exactly boolean enabled and integer
target_bytes, inclusive 20,000 through 16,000,000. The default is 400,000; zero
is invalid here. This upper bound limits selection work and policy arithmetic;
it says nothing about a model's context window. Retain the 64 KiB complete-file
limit, exact uint64 revision, strict fields, checked replacement, conflict and
busy behavior from Chapter 8. A missing file supplies these defaults. Reading a
strict v1 file supplies disabled/default context in memory, preserves its revision,
and performs no write. The first actual change writes a complete v2 file. An
unchanged update writes nothing, including at maximum revision.

Extend the existing patch domain with optional context. When present it must
contain at least one of enabled and target_bytes, with no other keys; merge
into one owned complete policy candidate. Validate profile compatibility before
starting its writer. A nested no-op is a whole-policy no-op. Revision exhaustion
refuses changes before replacement. Preferences remain separately owned.

The turn captures max_model_requests as before. Maintenance instead samples the
latest applied context setting at each request decision, after prior accepted
calls have finished pairing and before collecting this attempt's observations.
Recheck a selection candidate if that applied revision changes. An enabled update
does not cut immediately; a disabled update cancels no committed change and
cannot restore removed parts. It cancels an uncommitted handoff intent at its
normal disposition boundary. Changing target alone affects future automatic
selection; an admitted handoff retains its captured setting and revision. Once
a disabling update has canceled an intent, re-enabling before its disposition
cannot revive it.

For capable Agents, request_sent gains required context_policy containing exactly
revision, enabled and target_bytes, with the same domains. It is the applied
selection captured for that attempt; live append must match the Actor's candidate.
It does not replace turn.policy or restore today's settings on replay. Standalone
inspection accepts the optional field strictly; old session versions refuse it.
Every v4 request carries it, including while disabled. No request receipt is
created merely by previewing or inspecting context.

## 14.4 Know which batch the model had a chance to use

A batch is all tool calls in one accepted response_ended, identified by that
event's positive sequence. Its call order is the response's part order. It closes
only after every call has its one normal paired result, including refused or
failed attempts. Interleaved Jobs events do not split a batch. Removing parts
never makes a sequence or call ID reusable.

A represented batch still has its complete call/result pairs in the current
projection. Results may already be stubs. It becomes eligible for automatic
maintenance only when a valid final response is accepted to a request that
contained all those pairs. Derive this relation from the recorded pre-request
projection and response association, and retain the necessary typed indexes in
semantic state. request_sent alone, provisional deltas, HTTP failure, rejected
usage or an interrupted unfinished response confer no eligibility.

After an accepted response, admit and pair its calls before choosing maintenance
for the next attempt. This gives keep_tool_results in that response an opportunity
to protect what the preceding request carried. A paused, not-yet-attempted keep
does not allow an automatic cut to race past its batch. Interrupt follows the
existing pairing rules before it settles context intents.

keep_tool_results takes exactly `{}`. Its target is the newest represented,
eligible, completed batch in the request that produced the keep call, by batch
sequence. It never targets the batch containing that call. No target returns
`context_no_target`, without a keep fact. A valid call with disabled maintenance
returns `context_disabled`; malformed, ungranted or unsupported calls retain
their ordinary refusal and next-attempt limit consumption.

A successful keep protects the whole target from immediate stubbing and both
ladder stages until the next valid accepted response to a later request containing
that batch. HTTP failures do not expire it. That accepting response can renew
protection with another keep before the next maintenance decision. Multiple keeps
in the same producing batch for the same target share one protection fact; later
ones return the same kept acknowledgement as no-ops. An idle typed public keep
uses the newest currently represented eligible batch and the same next-accepted-
response expiry. Busy public controls refuse immediately. The successful tool
acknowledgement is compact canonical JSON plus LF, for example:

```json
{"batch":7,"status":"kept"}
```

Its batch must be the actual selected identity, never the keep call's own batch.

The newest eligible represented batch is also protected from ladder stubbing and
paired removal, even without keep. This does not exempt its large results from
the immediate rule below. A valid keep supplies that additional exemption. The
keep acknowledgement itself is exempt from immediate stubbing, but its batch
can later enter either ladder stage or an explicit handoff.

## 14.5 Measure tool bytes, then choose exact cuts

Use a neutral policy measure, not token estimates or the provider's HTTP encoding.
For each batch form exactly an object with calls and results arrays. Calls contain
the complete retained typed tool_call parts in original call order; results contain
the complete typed tool_result parts in that same order, including call_id, all
children, error fields where the inherited type has them, references and retained
opaque values. Do not count event/time wrappers, surrounding text, manuals, notes
or request observations. Sum each batch object's canonical UTF-8 byte length;
there is no outer array charge. Use Chapter 10's lossless JSON canonicalization,
including its exact decimal number rules, without changing original payload bytes.

A result's immediate-stub size is the canonical length of that complete
tool_result part. This small control fixture illustrates the string/number rule,
independently of an implementation's surrounding typed part fields:

```json
{"text":"é\n<","n":10.0}
```

Its canonical bytes are `{"n":1e1,"text":"é\n<"}`, exactly 24 UTF-8 bytes,
with a literal two-byte é, a backslash and n, and unescaped `<`. A binary64
round trip that merges distinct large argument integers is invalid accounting.

Let T be target_bytes, S=floor(T/100), R=floor(T/8), C=floor(T/16). At the
default these are 4,000, 50,000 and 25,000 bytes. Use checked integer arithmetic.
Equality does not cross a threshold. At one admitted decision, simulate these
steps in order against an owned candidate:

1. Immediately stub each eligible, unkept result whose current canonical length
   is greater than S, except keep_tool_results acknowledgements. Already fully
   stubbed results are unchanged. Preserve the corresponding call.
2. A batch is in the results band while any result child remains unredacted; a
   batch whose results are all stubs is in the calls band. Empty result children
   count as already stubbed. Sum complete batch sizes for each band. If results
   bytes exceed 2R, stub whole oldest eligible unprotected results-band batches
   until its size is at most R or no eligible unprotected batch remains. Reclassify
   them using their actual retained stub sizes.
3. If calls-band bytes now exceed 2C, remove whole oldest eligible unprotected
   calls-band batches until its size is at most C or no such batch remains.

The immediate rule can move a batch into the calls band before step 2. A partially
stubbed batch remains in the results band. Band totals include ineligible and
protected material; those bytes may force honest overshoot. Remove a selected
batch whole even if doing so goes below the target. Never split pairs to hit an
exact budget. The newest eligible protection and unaccepted material can leave
either band above its target or trigger, which the state view reports plainly.

Stub children exactly as Chapter 2 redact_result does: each selected child becomes
`{"type":"redacted","stub":"[redacted]"}`, retaining its ref when present.
Keep the result wrapper's identity and failure status. An automatic selected
result is fully stubbed. Ordinary explicit Chapter 2 redaction stays available
with its old semantics; it neither requires this policy nor acquires a paired-
removal meaning. A stub that grows a tiny child still counts its actual new bytes.
Repeated already-stubbed selection is not a change.

Here is an illustrative selection for the configuration-repair task, using
stipulated neutral sizes after canonical accounting, rather than a measured
provider run. Set T=20,000, so S=200, R=2,500 and C=1,250. All four batches are
completed and eligible; batch 19 is newest and has an unexpired keep:

| Batch | Configuration evidence before this decision | Whole batch bytes | Bytes after stubbing its results |
|---|---|---:|---:|
| 7 | An older inspection, already all stubs | 2,600 | 2,600 |
| 11 | A long comparison argument and a small result no larger than S | 3,100 | 3,000 |
| 15 | A file read whose complete result is 210 bytes | 500 | 378 |
| 19 | The kept current configuration read | 2,300 | Not selected |

First, result 15 exceeds S and becomes a stub. The results band is now
3,100+2,300=5,400 bytes, above 2R=5,000. The ladder stubs batch 11, leaving
2,300 results-band bytes. The calls band now holds 2,600+3,000+378=5,978 bytes,
above 2C=2,500. It removes batches 7 and 11 whole, oldest first, leaving 378
bytes. Batch 15 survives as a call and stub; batch 19 retains its kept evidence.
The committed fact records only the surviving stub target and removals 7/11,
without a redundant stub target for the removed batch 11.

For the reader checking the repair, the decision text “Use port 9090” and the
loaded editing manual remain where they were. A reference on result 15 survives
its stub; absence of a reference stays absence. The old comparisons are still
in the full log. If the protected newest batch alone exceeds R or C, the same
rules retain it and report overshoot instead of pretending the target was met.

If no effective change is selected, append no maintenance event. Otherwise one
exact context_changed fact commits the final candidate. Each stub target is a
call ID and each removal target a batch sequence, sorted in historical order;
omit stubs whose whole batch is removed by the same candidate. Validate by
recomputing this bounded selection from the prior state and captured policy,
rather than trusting a supplied list of plausible IDs. Replay applies the
recorded decision under that captured policy, never today's settings.

## 14.6 A working note needs a truthful acknowledgement

micro_handoff takes exactly one nonempty UTF-8 string note, at most 65,536 bytes.
It records useful working facts: objective, decisions, evidence locations, failed
approaches and next action. It does not create a new Agent or become a system
instruction. Treat its text as attributed assistant material, not verified truth.

The first valid admitted handoff in a batch occupies its one intent slot.
Validity here means syntax, grant, capability, enabled policy and note bound;
final projection compatibility can still fail after other calls finish. Invalid
calls do not occupy the slot. Later otherwise valid handoffs in the same batch
return `context_duplicate_intent`, even if the first later fails. Every attempted
call follows inherited Jobs limit consumption; the stage acknowledgement is a
normal paired result with intact control fields, like a skill-management receipt.

Commit context_intent before returning that result. Its intent ID is its own
event sequence, so no separate allocator can wrap. The exact payload key intent
has batch, call_id, note and policy, where policy is the complete captured
context_policy shape. The batch/call must be the accepted unanswered handoff.
The immediate result text is compact canonical JSON plus LF:

```json
{"intent":42,"status":"staged"}
```

It says nothing has been removed yet. Waiting synchronously for the enclosing
batch would deadlock on this tool's own missing result. Once normal results have
closed the batch, Actor either commits the removal and note together or records
one terminal disposition. Interrupt or close cancels an uncommitted intent after
normal pairing; disabling policy does likewise. An append failure retains the
inherited terminal fault and may leave no writable disposition. Never claim a
cancellation event was stored when the writer failed.

A successful handoff removes every currently represented completed tool batch,
including its own and kept/ineligible batches. This is an explicit exception to
automatic accepted-response eligibility. It preserves all other parts and every
logical placement anchor. Its note enters once at the handoff's post-batch
boundary, after the batch's sequence-ordered hints/manual material. If that
material becomes available while results are unresolved, finish its inherited
placement before inserting the note. No later policy change moves the note.

For example, let assistant text A surround batch B, hint H precede manual S in
event sequence while B is unresolved, and prompt P arrive after that turn. Before
removal the next request is ordered A→B calls/results→H→S→P. With handoff note N,
it is A→H→S→N→P. Once H is consumed the next projection is A→S→N→P. Removing B
does not turn S into request-tail material or allow H to be consumed twice.
Captured observations follow this entire inherited suffix once per request.
The note renders as ordinary assistant text, exactly `[working note]`, LF, note,
LF, `[/working note]`, with no final LF. For note `Port is 9090.`, the three
provider entries are:

```json
{"role":"assistant","content":[{"type":"text","text":"[working note]\nPort is 9090.\n[/working note]"}]}
```

```json
{"role":"assistant","content":"[working note]\nPort is 9090.\n[/working note]"}
```

```json
{"role":"model","parts":[{"text":"[working note]\nPort is 9090.\n[/working note]"}]}
```

Ordinary adjacent-role merging follows each inherited renderer. These literal
entries do not create an additional provider role or resurrect removed calls.
The note is absent from the micro_handoff tool-result body and appears once in
later requests. Context changes and restored notes are silent in the browser;
they do not replay assistant speech.

At an idle settled boundary a public handoff validates and commits directly,
using null intent, with no invented model call. It returns the committed sequence
or a refusal; otherwise return context_busy immediately. Public controls do not
consume pending tool_limits. No automatic handoff is triggered by size alone.

## 14.7 Commit a transition, not a replacement snapshot

The following new payloads use the existing strict event envelope and sequence.
All objects reject missing, duplicate and unknown structural fields. Positive
identities are full uint64; no browser rounding, allocation wrap or reset is
allowed. References to earlier sequences must resolve in represented state.

| Event / payload key | Exact required payload members |
|---|---|
| context_kept / context | batch, producer, policy; producer is the accepted response sequence containing the tool, or null for idle public control |
| context_intent / intent | batch, call_id, note, policy, as above |
| context_changed / context | action, intent, base, policy, stub_calls, remove_batches, note |
| context_disposed / context | intent, outcome, code |

context_changed action is auto or handoff. Base equals the preceding event
sequence. Policy has revision/enabled/target_bytes and must be enabled. Auto
requires null intent, empty note and at least one effective target. Handoff uses
a pending intent ID or null public intent, empty stub_calls and its exact note;
an empty removal list is permitted because the note itself is new material.
stub_calls is the unique historically ordered call-ID list. remove_batches is
the unique increasing sequence list. Handoff's list must equal every represented
completed batch, rather than an arbitrary subset. The successful changed fact
itself settles a referenced intent; do not append a second committed marker.

For a valid prefix whose pending intent 42 belongs to batch 31 and whose complete
represented batches are 7 and 31, the handoff payload at event sequence 50 is:

```json
{"action":"handoff","intent":42,"base":49,"policy":{"revision":2,"enabled":true,"target_bytes":400000},"stub_calls":[],"remove_batches":[7,31],"note":"Port is 9090."}
```

This is a payload fixture, not a complete standalone log. Its referenced prefix
and matching admitted intent are required; changing 31 to an unrelated batch
must refuse before mutation even though the JSON remains valid.

context_disposed outcome is refused or canceled. Code is respectively one safe
context error or context_interrupted, context_closed, context_disabled. Intent
must be pending, fully paired and undisposed. It installs no note or cut. The
public state reports committed at the context_changed sequence, or the recorded
refused/canceled disposition. A normal model continuation sees its staged tool
receipt and the resulting note on success; a refusal/cancellation is visible in
public state and clients, without inventing an additional tool result or model
message. The embedding application can inspect that disposition through public
state; no extra model-facing inspection tool is introduced here.

context_kept requires a present target and the selection in §14.4. A non-null
producer must contain a valid admitted keep call with no prior protection fact
for that target/producer; the fact precedes its successful paired result. Record the
target's protection boundary in reduced state; accepted response relationships
derive its expiry. An already protected target under the same producer is a
no-op. A new producer can renew it. Live append of every new fact must match the
Actor's admitted typed candidate; an external caller cannot fabricate an intent,
keep, policy revision or cut merely by submitting correctly shaped JSON.

Offline validation checks targets, pairing, identity, expiry, note placement,
captured policy ranges and state transitions without consulting today's settings
or executing effects. It can establish a recorded policy's consistency and use,
not authenticate a settings file absent from the log. Never skip a malformed cut
and continue with a different conversation. Reject without mutating the accepted
prefix, and leave ordinary invalid live controls usable for later valid work.
Safe new error codes are context_profile_required, context_disabled, context_busy,
context_no_target, context_duplicate_intent, context_invalid, context_limit and
context_incompatible, plus the three cancellation codes named above. Invalid
arguments and ordinary permission failures keep their inherited tool error path.
A refused disposition uses context_invalid, context_limit or context_incompatible;
its safe code cannot be arbitrary model/provider text.

Before any durable change, enforce the whole encoded record limit of 64 MiB
including framing LF, the inherited million-element collection limit and the
256 MiB canonical semantic-state limit. A note's decoded 64 KiB allowance is
independent of those whole-fact bounds. Bound raw reads before allocation exceeds
the limit; validate original input bytes rather than a different remarshal.
Candidate writes are measured as the exact bytes to be written. Exhaustion or
oversize refuses context_limit before event/state/counter changes. Ordinary
enclosing call/result/limit facts still describe that attempted management call.
Earlier human-input and ordinary GUI-message caps also remain in force; a public
note at the semantic bound need not fit inside a smaller client's command envelope.
The inherited log-file/event-count exhaustion remains a terminal storage limit,
rather than a reason to discard originals or make room by truncating history.

Local compatibility means the resulting context passes all existing structural,
pairing, reference and opaque-provenance render checks for the selected target.
Do not infer that a changed prefix is accepted by every remote deployment.
Preserve signed text and standalone opaque bytes; no beta header, alternative
route or silent stripping is introduced to obtain acceptance.

If an automatic candidate is locally incompatible, publish safe runtime status
deferred with context_incompatible and proceed once with the unchanged projection,
provided that projection is renderable. Keep only the latest deferral key:
effective selection revision, applied policy revision and target provenance.
The selection revision covers represented parts and anchors, batch completion,
eligibility, newest-eligible selection and keep protection, including renewal
and expiry. It is semantic state, not a hash of rendered bytes or the latest
event sequence. An accepted empty response can make a batch eligible or expire
a keep without changing those bytes; either transition requires reconsideration.
Derive this invalidation from existing reducer transitions; it needs no new
durable event solely to advance a runtime deferral key.

Job facts, request bookkeeping and repeated failed HTTP attempts alone do not
change the selection key. An unchanged blocked candidate gets no retry within
that key; after a relevant semantic, policy or target change, reconsider once
under the new key. Never spin or accumulate deferred work. This status
changes no durable conversation. A failed automatic size candidate similarly
defers with context_limit. Explicit handoff instead records refused for its
pending intent, or returns a public refusal, with old projection unchanged.

A remote rejection after a committed cut is a failed request. It does not roll
back history. The reader can inspect originals, disable future maintenance or
select a fresh session deliberately; disabling cannot restore removed parts.
Current model/API compatibility, cache hits and task quality require their own
source-bound observations. Smaller neutral byte counts prove none of them.

## 14.8 Persist the projection without inventing its past

V4 uses the Chapter 10 checkpoint outer fields with version and state_version
both 4. Identity has exactly mode, system, skills, handlers, mcp_bindings,
integrations and context_profile. The first four keep Chapter 10 semantics;
mcp_bindings and integrations keep Chapters 11/12's normalized definitions but
may now be empty arrays. context_profile is exactly true. Nonempty integrations
still require skill mode and all inherited fixed-alias/grant correspondence.
Policy enabled/target/revision are mutable and are not identity fields.

The v4 initializer's session payload has exactly version, session_id and identity,
with version 4. The v4 anchor has exactly version, session_id, origin_as_of,
origin_sha256 and high_watermarks, also version 4. Keep the old event header and
construction-only sequence rules. Initializer/anchor, checkpoint, semantic state
version and immutable origin agree. An explicit empty binding array is not
silently inserted into v1/v2/v3. Their strict shapes and request-field refusals
remain unchanged. No public append can select the new profile for an old Agent.

The documented strict v4 semantic codec adds current represented batch parts,
stable anchors even after removal, eligibility/protection relationships, pending
intent/disposition identities, working notes and historical captured context
policies to the existing reduced state. Store enough transition provenance to
validate exact cuts, not just a final list with deleted parts missing. Retain
used call IDs, original part identities, usage and allocator watermarks. No
separate journal or duplicate mutable Context is introduced.

Removed-batch summaries retain batch/call identities, measured before/after sizes,
eligibility/protection boundaries and the change sequence. These are bounded
transition metadata, not a second copy of discarded bodies. Full-origin replay
checks them against original parts. A snapshot-only importer checks their internal
consistency and correspondence with represented transitions; it cannot independently
prove the bytes of a discarded original. Its hash and identity do not authenticate
facts supplied by an operator who can rewrite the snapshot.

Full-origin verification still reduces the available prefix and compares semantic
state, then applies the tail once. Changing today's T cannot choose different
historical cuts. A settled checkpoint has no unanswered calls, unresolved batch
or handoff intent. A pending protected batch waiting for its next accepted
response is settled state and may survive restart. No live job, collector,
transport or native speech lifetime is restored.

Snapshot-only import retains only the history represented by its semantic codec.
For the configuration repair, a saved working note can still name port 9090
after the old comparison body has gone. A reader asking what that earlier file
actually contained needs the original log, not a plausible expansion of the note.
If pre-cut original bodies are absent, a request needing them returns
history_unavailable; neither rendering nor import fetches an artifact or contacts
a model to guess them. The public inspector exposes the exact reconstructible
request IDs, not a misleading complete-history claim. Current projection and
post-import requests remain reconstructible under the ordinary captured-input
rules. Test that distinction with full-log and snapshot-only inputs separately.

MCP preparation and integration leases follow the existing resume order: validate
stored state and current caller identity before remote preparation, then bind
the intended logical scope. Removing an MCP result does not close its Connection,
change a skill lease or replay the call. Core maintenance operates over typed
parts for stdio, public custom transports and the actual optional-GUI WebSocket
tunnel alike. A cut never disposes the GUI's physical socket or another channel.

Expose an owned typed context-state getter through the public Agent interface,
captured at one Actor boundary. Its safe wire object has exactly capable, policy,
results_bytes, calls_bytes, overshoot, newest_eligible, kept, last_change and
status. Policy is the complete current context_policy. Byte totals and identities
are exact nonnegative uint64; newest_eligible/last_change are positive or null;
kept is a sorted batch-ID array; overshoot is true if either band exceeds its
target. Status contains state and code: state is idle, staged, committed,
refused, canceled or deferred, and code is empty or one safe error. Status also
contains intent, null unless it refers to a model intent. It carries no original
body, credential, arbitrary provider error or hidden reasoning.

Add this owned value as context_state to capable Agents' safe watch snapshot and
publish context_state_changed when its content changes, under the inherited
ordered watch envelope. That observation has exactly kind, agent_id and
context_state, with the owned value described above. Uncapable Agents expose
the typed getter with capable false and empty batch state, without changing old
session facts. Browser counters,
card keys and controls preserve exact newly introduced integer tokens. Settings
updates retain the existing revision/conflict draft behavior and disabled-control
lifetime rules; rendering old context facts cannot speak or submit a prompt.

Human CLI adds `/context` for that state, `/context on`, `/context off`, and
`/context target N` through current-revision policy updates. `/keep` and
`/handoff TEXT` are the idle typed controls. Empty text or trailing malformed
numeric syntax refuses. These commands make no model request. The browser adds
the same applied context policy controls and safe status, plus explicit idle
keep/handoff actions; a busy action reports refusal without losing a human draft.
Both actions use the subscribed connection's selected Agent and these exact
ordinary command shapes:

```json
{"type":"context_keep","id":"k1"}
```

```json
{"type":"context_handoff","id":"h1","note":"Port is 9090."}
```

Require exactly the shown fields, rejecting unknown, missing, duplicate or
wrong-type members. Note uses §14.6's validation; an empty note is an
invalid_command. Chapter 7's assembled ordinary text-message limit of 65,536
bytes, usable-ID correlation, subscription and
4096-command connection cap remain unchanged, including on an MCP-enabled socket.
Malformed transport or unusable ID closes; a correctable shape error with a
usable ID returns invalid_command and leaves the connection usable. Reused IDs
retain the inherited refusal, rather than repeating an effect.

These commands perform the idle public operations, not a prompt or staged model
call. A busy owner immediately returns context_busy, for example:

```json
{"type":"error","id":"h1","code":"context_busy","message":"context control requires an idle settled Agent"}
```

Other semantic refusals use the same error shape and §14.7's safe code. Preserve
the user's composer and handoff draft on failure. On success, publish the owned
context-state change through the ordinary watch before this exact acknowledgement:

```json
{"type":"context_ack","id":"k1","action":"keep","changed":true,"seq":50,"batch":7,"watch_revision":90}
```

```json
{"type":"context_ack","id":"h1","action":"handoff","changed":true,"seq":51,"batch":null,"watch_revision":91}
```

Action is keep or handoff. Seq is the committed fact's positive sequence; batch
is the selected positive batch for keep and null for handoff. An unchanged keep
acknowledges changed:false, seq:null, the same selected batch and current watch
revision, without another fact or broadcast. Handoff always commits a new note
on success. Preserve every numeric token exactly. The acknowledgement follows
the Actor's owned result; an uncommitted operation cannot claim success because
a socket write succeeded. Clear only the acknowledged handoff draft if it has
not changed since submission. Reconnect shows committed state; it does not replay
the command or retry an acknowledgement lost in transit.

Context policy edits remain policy_update under Chapter 8's revision/conflict
rules, with the added nested patch, for example:

```json
{"type":"policy_update","id":"p3","base_revision":2,"patch":{"context":{"enabled":true,"target_bytes":20000}}}
```

Extend the complete execution_policy snapshot with required context containing
enabled and target_bytes. This is the same Agent-owned value shown with its
revision in context_state.policy. A policy change still publishes policy_changed
and returns policy_ack under the inherited ordering; publish any resulting
context_state_changed before that acknowledgement too. A policy update alone
does not run selection or contact a model.

Do not add a context command to the inherited machine JSON protocol silently;
public consumers already provide the transport-independent control seam.

## 14.9 Taking it for a spin

Actual runs are pending. These steps define the demonstration, rather than report
a successful cut. Use a fresh explicit profile directory and applied disabled
policy first. Ask a real model to read a bounded fixture with a final marker and
report it. Inspect `/context` and the request receipt: disabled policy preserves
the result. Enable maintenance, repeat with a new fixture, and inspect the next
accepted request after the result becomes eligible. Compare original artifact,
deterministic stub and retained explanation. Wait for each answer before following
up; a captured outgoing body does not prove a valid response arrived.

Ask for keep_tool_results alongside another bounded read. Record the selected
previous batch and show its complete result in the next request, then its normal
expiry after a later accepted response. Exercise both ladder stages with enough
bounded tool traffic, including long arguments so the calls band actually crosses
its trigger. Record actual neutral sizes and exact targets; local fixtures prove
boundary arithmetic without buying a huge provider context just for a threshold.

Ask micro_handoff to preserve the task's marker, decision and evidence locator.
Observe staged then committed status and the next request's single note. Repeat
with another simultaneous tool call under the local batch fixture; the actual
provider run must still demonstrate real handoff use, but cannot be assumed to
emit a particular parallel shape. Verify dialogue/manual survival and unchanged
grants. A model refusing the task, a local compatibility deferral or a remote
request rejection remains an observed outcome, with the affected feature gate
left open until its required behavior is demonstrated.

From the actual browser, inspect the same Agent, change applied settings and
perform an idle public handoff. Keep the GUI tunnel active and observe its chosen
scope through an explicitly authorized independent consumer when target typing
or speech pause prevents a same-Agent call. Preserve Chapter 12's distinction
between the pause gate and an admitted GUI action. Capture real screenshots with
descriptions; a DOM stub cannot establish actual GUI control.

A headless public example creates two independent bounded Agents and exercises
typed keep/handoff and model tool paths. Retain each partial result, usage and
terminal outcome even if its peer fails. Cut one Agent while showing the other's
bytes, policy, skills and MCP leases unchanged. Checkpoint settled state, close,
resume with matching v4 identity, and verify exact next projection without
reviving jobs. Reconstruct pre-cut requests from the complete log with every
endpoint disabled; separately import a snapshot lacking those bodies and observe
the honest history_unavailable boundary.

Run the required feature matrix on all three supported provider paths. Retain
source/binary identities, policy/store identity, sanitized PTY/browser/public
actions, raw request/response bodies, artifacts, exact usage and partial outcomes.
Measure serialized bytes separately from normalized usage, cache counters and
task quality. No repeated paid call is justified merely to turn an awkward
model answer into a cleaner story.

| Required check | Distinguishing positive and negative controls |
|---|---|
| Selection | Disabled unchanged; exact S/2R/2C versus plus one; Unicode/escaping/lossless numbers; long arguments force calls band |
| Batches | Parallel calls, mixed text/opaque, failed results, whole-pair cuts, newest/ineligible/kept overshoot |
| Eligibility/keep | Accepted response versus failed/refused/interrupted attempt; previous-batch target; renew/no-op/no-target and failed-HTTP nonexpiry |
| Handoff | Staged/result/batch order, first-valid slot, duplicate, interrupt/disable/close disposition, exact once-only note and no circular wait |
| Safety | Stale candidate, wrong targets, forged policy/intent, append failure, encoded record/state/counter bounds, safe deferral progress |
| Placement | H/S/P before and after removed anchor, retired manuals, surrounding text, captured observations once at tail |
| Persistence | Strict v1–v4, old compatible ceiling/catalog, v1 policy read/no-op, incompatible enabled policy, snapshot-only absent history, unchanged historical request bytes |
| Clients | Exact numeric public/watch state, busy controls, two Agents, real optional browser, replaceable MCP transport and independent logical channels |

Retain every passing inherited check and add deletion controls that fail for the
intended reason. The historical grader's tolerant reducer and frozen inline-tool
prefix are different contracts; they cannot excuse skipping corruption here or
changing today's skill declarations. Independent acceptance must observe public
behavior and the printed format, without inventing private payload field names.

The reader should finish with a shorter working request and an explanation of
every removed byte. The original goal stays where the user put it. The log still
answers what happened when the working note turns out to be incomplete.
