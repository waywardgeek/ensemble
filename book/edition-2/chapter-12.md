# Chapter 12: The Agent sees its interface

A model can read the code that opens a settings panel and still report the wrong
panel. The source describes what should happen. The person at the keyboard needs
to know what happened in the running page.

The first edition ran into that distinction hard. Programmatic input left the
Agent paused as though a person were still composing, and nobody noticed until
a human sat watching the frozen prompt. A panel-observation failure later exposed
missing semantic state (§12.4). Then an integration review turned up something
worse: the shipped debug skill had never actually connected the browser. The
grader's fake stdio server answered its calls, so the tests passed, but the
real GUI had never been observed. The
[historical evidence](chapter-12-evidence.md#historical-findings-and-limits)
distinguishes the source repairs from unrecovered model transcripts.

This chapter carries MCP through the actual GUI connection. The wiring is
mostly mechanical and occupies the first half. The second half reports what
happened when the capability pointed at a GUI nobody had audited and the human
supervising it agreed to stay quiet. The Agent can inspect one selected view,
operate its supported controls, and check the result. A missing browser produces
an unavailable observation. It cannot produce an imaginary screen that happens
to agree with the code.

*The [validation record](chapter-12-validation.md) tracks prerequisite and
acceptance gates.*

## TL;DR

Read the complete [coding skill](skills/ensemble-coding/SKILL.md) and
[architecture](architecture.md). Extend the accepted Chapter 11 source.

1. Implement the real MCP WebSocket tunnel in the optional GUI module through
   Chapter 11's public complete-message transport seam. Protocol discovery,
   correlation and cancellation remain in MCP. A logical endpoint's close cannot
   close another channel or its healthy physical GUI socket.
2. In explicit GUI-debug startup, create an unbound browser view first, discover
   its endpoint, freeze aliases and policy, then construct/resume the Agent and
   mount its Page. Wait at most 120 seconds for explicit view attachment. Ordinary
   no-debug startup retains its earlier behavior.
3. Expose gui_snapshot, tts_queue, gui_click, gui_input and gui_submit, scoped to
   the selected view and current mount. Return actual semantic state, bounded
   previews and truthful omissions. Targeted reads use artifact identity/version;
   controls refuse stale, disabled or out-of-scope targets.
4. Use Page-owned actions. Programmatic input does not create a human typing
   pause. Human edits retain their rights: refuse to overwrite a nonempty human
   draft, preserve other clients' pause causes and await the normal server
   acknowledgement for applied settings.
5. Add immutable typed skill-to-integration policy without changing Chapter 9
   frontmatter. Skills governs admission; Agent leases govern its work. Unload
   cancels automatic collection and releases only that Agent's lease, preserving
   admitted model Jobs and shared connections.
6. Collect automatic observations off the Actor through normal MCP. Revalidate
   the attempt and generation, then capture the exact ordered observations inside
   request_sent before HTTP. Pure rendering takes explicit owned samples;
   historical reconstruction uses that request's samples with no browser I/O.
7. Use strict session version 3 only with nonempty integration policy. Retain
   v1/v2 formats. Validate stored state before remote preparation; resume requires
   the same logical authority but binds a newly selected physical view explicitly.
8. Prove actual human CLI, browser and public composition on all three APIs,
   plus local fault controls. Retain screenshots, applied effects, original
   failures and exact request reconstruction. A fake MCP server is insufficient.

Public method names and private implementation types remain student choices.
Observable wire shapes, ownership, bounds and lifecycle below are requirements.
From `solutions/edition-2/main/`:

```sh
go build -o /tmp/ensemble-ch12-cli ./cmd
go vet ./...
go test ./... -count=1
```

Run vet/tests in every affected module and require empty gofmt output. Build the
GUI from `solutions/edition-2/main/gui/` with
`go build -o /tmp/ensemble-ch12-gui ./cmd/ensemble-gui`. From the repository root,
retain
`make grade-dir CH=13 DIR=solutions/edition-2/main` as a historical diagnostic.
Publish the new independent checker command before student handoff; §12.10 is
its required behavioral matrix.

## 12.1 Open the view before freezing its tools

Chapter 11 freezes remote aliases before Agent construction. The GUI normally
starts with an Agent already available. Registering tools lazily when a tab opens
would evade that creation boundary and give timing control over permissions.
Instead, let the browser view exist before it has an Agent.

The optional launcher gains `--gui-debug` and `--gui-view NAME`, with main as the
default view name. A name follows Chapter 11's logical-key grammar. These options
select an explicit integration; they do not change plain CLI or no-debug GUI
startup. The core CLI alone cannot serve the debug browser and refuses those
GUI-only options visibly. The optional launcher's `--terminal` attaches human
chat to the same Agent after preparation.

Startup proceeds in this order:

1. Create Ensemble and the optional GUIServer. Read caller configuration and
   validate any selected saved session locally, including settled state and
   identity. No new transport preparation precedes that validation.
2. Serve a BrowserApplication and an unbound view scaffold. Show the startup URL,
   selected view and waiting state. Disable prompt submission until an Agent
   exists. Start a cancelable 120-second attachment timer when serving begins.
3. Attach exactly the selected browser endpoint. The view owns a scoped DOM root
   and a fresh mount identity. A duplicate live claimant is refused; a caller
   must close/unbind the prior owner before selecting its replacement.
4. Supply the GUI transport constructor to the public MCP service. Discover the
   fixed browser tool definitions, validate aliases and construct/resume with
   immutable bindings and integration policy. Chapter 11's 30-second preparation
   bound starts after attachment, separately from the waiting timer.
5. Mount the Agent's Page in that existing view, register its ordinary watch and
   pause ownership, and activate the configured gui-debug skill through the
   existing typed skill control. Only then enable human prompts and the terminal.

Failure displays a safe stage/error and shuts down the failed startup's owned
resources. It makes no model request and does not silently continue without the
selected integration. A shared connection prepared for another Agent remains
owned by Ensemble. Public embedding exposes this same phased boot path; it must
not depend on private launcher functions or a global pending Agent.

The default example ships a valid gui-debug loadable skill and a primary that
offers it. Its ordinary tools list grants the five selected aliases. The
application installs those aliases before construction and supplies the typed
policy in §12.6. An incompatible caller-selected catalog or alias collision is a
startup refusal. The convenience flag cannot rewrite the catalog, add hidden
grants, or reinterpret an old `mcp_servers` field.

After a disconnect, the caller explicitly rebinds the same logical view scope
to a new mount, then reopens the MCP connection through Chapter 11's compatible
discovery path. Frozen definitions must still match. Never replay a sent control
operation or select whichever tab replies first. A new Agent whose ceiling lacks
these aliases requires deliberately new creation configuration.

## 12.2 A channel is smaller than a socket

Closing a debug skill must not close the conversation it is helping to debug.
The GUI owns the physical socket. MCP owns a logical endpoint carried on it.
Keep those resources separate even when the demonstration uses one of each.

| Object | Owned state | Actual parent |
|---|---|---|
| GUIServer | Physical connections, selected view bindings and application shutdown | Public Ensemble owner |
| GUI Connection | Socket reader/sender, bounded routing queues and channel registrations | GUIServer |
| MCP Connection | Protocol, discovery, pending calls, IDs and transport generation | MCP service → Ensemble |
| GUI MCP transport | Logical route, owned delivery/receive work and close state | Actual MCP Connection |
| BrowserApplication | View scaffolds, endpoints, Pages and native speech service | Explicit browser application root |
| Browser endpoint | Scoped control/artifact indexes and current mount | BrowserApplication |
| Page | Draft origin, action state, watch and local pause causes | BrowserApplication |
| Agent | Immutable integration policy; Actor, Skills and context collector | Ensemble |
| Context collector | Bounded transient collection and this Agent's leases | Agent |

The public transport constructor receives the actual new MCP Connection parent.
It may carry the application-owned logical endpoint resource allowed by Chapter
11; that resource does not replace its parent or become a bag of sibling services.
Helpers retain the appropriate owner route to diagnostics. Core declares shared
transport-independent policy/sample values and interfaces in common; GUI envelopes
and DOM descriptions belong to the optional module's vocabulary.

Closing a logical route unblocks its readers and staged writes and cancels only
its operations. It does not dispose Page, clear unrelated pause, stop another
channel or close the healthy physical socket. Physical socket failure honestly
fails every child route. Root shutdown first ends admission, cancels work and
then joins workers; it cannot wait forever for a stopped browser to read.

BrowserApplication owns endpoint disposal just as it owns Page disposal. Remove
listeners, registrations and pending callbacks before another owner reuses the
DOM root. A late action reply or reconnect callback from the former mount cannot
alter the replacement. Native speech keeps Chapters 7–8's application ownership
and cooperating-tab lease; MCP acquires no new right to cancel another Page.

## 12.3 Carry MCP through the actual connection

Use a distinct, versioned tunnel envelope alongside ordinary GUI command/watch
traffic. Its route identifies one logical channel and generation before the
complete MCP payload reaches Chapter 11. Both ends reject wrong ownership and
stale generations; a browser does not gain the right to call Agent handlers by
sending MCP requests toward the client. The endpoint implements the supported
MCP server direction, with current discovery metadata and frozen tool schemas.

An explicitly tunnel-enabled socket permits physical text frames up to 12 MiB.
Here frame means one complete WebSocket text message, after any wire-level
fragmentation is assembled. Enforce the bound across that whole message, not
separately on each RFC frame. The tunnel adds no application-level fragments.
Every inherited command still has its exact 65,536-byte bound; no-debug sockets
keep the old physical limit too. This increases the bounded memory exposure of
an enabled connection. Read within the physical bound, identify its top-level
type with duplicate-safe bounded parsing, check the corresponding size, then
strictly validate that envelope. A large ordinary command is refused before its
ordinary payload is decoded or admitted. Do not claim an early streaming
discriminator if the chosen library first buffers a bounded complete frame.

The new envelope has exactly type, version, channel, generation and message:

```json
{"type":"mcp_frame","version":1,"channel":"gui_main","generation":1,"message":"e30="}
```

This is only a framing example: e30= decodes to `{}`, which MCP subsequently
rejects. Version is integer 1. Channel uses Chapter 11's logical-key grammar;
generation is that
Connection's positive uint64, and message is canonical padded RFC 4648 base64
of exactly one complete UTF-8 MCP JSON message. The alphabet and zero padding
bits follow [RFC 4648 §§3–4](https://www.rfc-editor.org/rfc/rfc4648.html).
Reject whitespace, noncanonical padding and decoded size above 8 MiB before
allocating beyond that bound. The whole outer frame is at most 12 MiB, including
whitespace and metadata. No
fragments, extra message ordinal or newline delimiter are introduced. Verify
lossless generation before dispatch. Invalid type/shape/encoding has no fallback
to the ordinary command handler; malformed physical envelopes close the socket
with a static protocol reason. A valid frame for a closed/stale route is discarded
with bounded diagnostics, without admitting effects on another route.

Enable and bind a tunnel explicitly through a pre-subscription command with
exactly type `mcp_attach`, version 1, nonempty ordinary id, view and mount. It uses
the
65,536-byte ordinary limit and command-ID budget. The server accepts only the
application-selected view and replies with type `mcp_attached`, version 1, the same id,
view, mount, channel and generation; those values are fixed for that attachment.
Duplicate/unauthorized attachment gets a correlated gui_unavailable error.
Unbinding uses type `mcp_detach`, version 1, id, channel and generation and returns
type `mcp_detached` with version 1 and id; its admission detaches that route alone.
An enabled
physical socket can carry separately selected channels through the public API,
up to Chapter 11's 32 prepared root connections. The launcher selects one.
Neither attachment nor a socket URL grants Agent tools.

During first boot, accept the selected attachment as an owned pending endpoint
resource, then start public MCP preparation. Its constructor receives the actual
Connection and generation; only then send mcp_attached and start discovery.
This ordering breaks the startup cycle without a fabricated Connection parent.
Attachment waiting is covered by the startup timer; disposal wakes both sides.

Retain one physical sender. Bound per-channel work and preserve Chapter 11's
64-operation admission, including cancellation delivery. A queued cancellation
must not allocate an unbounded goroutine after its pending RPC disappears.
Delivery has Chapter 11's one-second total staging-to-completion bound. Logical
queue overflow fails the affected route; an actual physical writer failure can
fail the socket. Give ordinary command/watch traffic service between MCP messages
so a tool burst cannot indefinitely prevent a human interrupt.

Bound queued plus actively writing MCP outer frames to 24 MiB and 64 frames
per logical channel, and 48 MiB per physical connection across its channels.
Account actual owned encoded bytes before enqueueing; capacity refusal closes
only the overflowing logical route. Retain the ordinary watch queue's separate
inherited bound. Reserve the per-channel delivery permit until delivery or close
settles; do not free it merely because the RPC caller stopped waiting. On receive,
allow one bounded physical frame under validation and at most one delivered
complete message per logical route awaiting its consumer. A slow consumer faults
its route instead of blocking ordinary command receipt. Decode/copy workspace is
bounded by these frame limits, not multiplied by the number of pending callers.

In the browser, a successful WebSocket.send call only queues bytes. Account its
outstanding bufferedAmount in the physical budget and do not repeatedly release
application capacity while the native send buffer keeps growing. The
[WebSocket send contract](https://websockets.spec.whatwg.org/#dom-websocket-send)
does not acknowledge peer receipt. A bounded sender tracks progress and faults
stalled delivery under the same deadline; an RPC reply remains the result receipt.

Expose explicit attach, unbind and compatible rebind through the optional public
API. No endpoint means immediate unavailable for calls on a previously failed
route. Do not save a click for the next browser to connect. Bind mount and channel
generation before admitting a control; never change its destination while queued.

The browser is the RPC server here. A new canonical rpc-N request is work to
admit, not an unknown reply to a browser-issued call. Scope admission state to
channel plus generation: at most 64 pending operations, a contiguous seen prefix
starting at zero, and at most 64 sorted disjoint seen-ID ranges above it. Merge
adjacent ranges and advance the prefix whenever possible. A fresh ID is a positive
uint64 in Chapter 11's canonical spelling absent from both the prefix and ranges.
Range comparisons must not wrap at the maximum uint64.
Check capacity, then mark it seen before validating tool arguments or admitting
an effect. Even a request that returns an argument error cannot execute again.

Discard a seen duplicate without another effect or reply, leaving any original
pending operation intact. Unseen requests may arrive out of order: rpc-2 followed
by rpc-1 is valid, and either may finish first. A long run above one missing ID
uses one merged range, not one record per request. If adding a fresh request
would require a 65th disjoint range or pending operation, fault only that logical
route before accepting it. Malformed/noncanonical request identities also fault
the route. Completion removes pending work after owned cleanup/delivery settles;
the compact seen representation remains until generation close.

Keep each request ahead of its own cancellation notice in tunnel delivery;
independent requests and completions need no common order. Cancellation names a
pending request: before action admission it prevents the effect; afterward it
cancels further waiting without undoing the action. Fence the canceled operation's
late callback and suppress a second settlement. A canonical cancellation naming
unseen or settled work is ignored without a tombstone. On rebind, create fresh
state only for the new generation after fencing old work from the new mount.
These server rules complement Chapter 11's client reply classifier; they do not
replace it or require retaining every completed request body.

## 12.4 Show enough state to correct a mistake

The reader asking which panel is open needs its selected state, not a list of
panel labels. The first-edition account asked the observer to click Artifacts.
It reported Settings, although the interface had no Settings tab. Labels without
selected state left room for a plausible answer to pass as observation. The
source repair adding semantic state is verified; the raw model transcript was
not recovered, so this remains an attributed historical account rather than a
newly reproduced result. Build the view from owned components with explicit
selected, expanded and checked state. Do not scrape the whole document or infer
hidden controls by guessing CSS selectors.

The endpoint advertises exactly the five remote names gui_snapshot, tts_queue,
gui_click, gui_input and gui_submit. The default aliases are identical. Public
embedders may choose other local aliases while preserving frozen remote semantics.
Each input schema uses Chapter 11's bounded profile and forbids extra properties.
All argument objects reject wrong types, duplicates and missing required fields.
Output schemas may be absent; every supported result still has the exact shapes
below. Return `content:[]`, structuredContent holding that shape and isError false
on success. Error results use §12.5's shape and isError true.

A mount is 32 lowercase hexadecimal characters from 16 fresh random bytes. It
changes whenever the endpoint view is disposed/remounted, including restart.
Stable logical view scope is a configured name such as main; mount is a runtime
identity. Control IDs are nonempty ASCII `[A-Za-z0-9_.:-]{1,128}` strings unique
within that mount and never reused there. They resolve through the view's owned
control map. Artifact IDs obey the same string grammar and identify one retained
artifact; version is a positive uint64, advancing before each content change.
Refuse exhaustion instead of reusing a version. Preserve numeric tokens exactly
through browser parsing and projection.

The snapshot call has exactly `{"mode":"view"}`. Its structured result has
exactly view, mount, ready, agent_id, state, controls, artifacts, omitted and
truncated. Before Page attachment, ready is false, agent_id is null, state is
null and both arrays are empty. A ready state contains the following safe fields:

| State field | Meaning |
|---|---|
| active_tab | Current application tab's stable ASCII name, at most 128 characters |
| input_owner, draft_scalars | empty/human/tool and current Unicode-scalar length; human draft text is excluded |
| typing, speaking | Exact nonnegative pause counts reported for this Agent |
| theme, autoplay, rate, preferences_revision | Applied Chapter 8 preferences and exact revision |
| default_turn_limit, policy_revision, active_turn_limit | Applied next-turn default, exact policy revision and active limit or null |

Each control has exactly id, label, kind, enabled, selected, expanded, checked
and value. Kind is button, tab, toggle, text, number or select; the three semantic
state fields are boolean or null when inapplicable. Label is at most 256 Unicode
scalars, with explicit truncation accounted below. Value is a bounded displayed
string of at most 1,024 scalars or null; the human-owned prompt's value is always
null. A tool-owned draft may be observed until a human edit transfers ownership.
Never include credential
fields, hidden scripts, raw HTML or another Page's contents.

Return controls in stable registration order, up to 256. Return up to 20 retained
artifacts in display order. Each artifact has exactly id, version, kind, preview,
total_scalars and omitted_scalars. Preview contains at most 1,024 Unicode scalars;
counts are exact when the owned retained text provides them, otherwise null.
Treat script-looking labels/text as data through existing safe rendering.

Bound the complete canonical structured snapshot, including metadata, to 65,536
bytes. Reserve metadata first, then admit whole control/artifact items in order;
do not cut the final JSON string. Values and labels can consume the remaining
budget only as complete scalar prefixes. Omitted has exactly controls, artifacts,
text_scalars and bytes: each is a nonnegative uint64 or null when the bounded
owned traversal cannot establish its total. A null means unknown, never zero.
Truncated is true if any content was omitted or its completeness is unknown.
Untruncated empty arrays have exact zero omissions. Existing owned indexes should
supply counts without rescanning an unlimited document.

Omitted.bytes counts the original UTF-8 bytes of eligible display text omitted
from control labels/values and retained artifact text. Omitted.text_scalars counts
the same text in Unicode scalars. Include a truncated suffix and all eligible
text of a wholly omitted item; count each displayed field once, even if two fields
have equal text. Exclude JSON quotes/escaping, structural metadata and text that
is outside observation scope, such as human drafts or credentials. If any required
text total is unknown within the bounded indexes, the corresponding aggregate
is null. For example, retaining `A` from eligible text `Aé🙂` omits six UTF-8 bytes
and two scalars, regardless of how JSON would escape them.

For example, the complete unbound-view structured value is:

```json
{"view":"main","mount":"0123456789abcdef0123456789abcdef","ready":false,"agent_id":null,"state":null,"controls":[],"artifacts":[],"omitted":{"controls":0,"artifacts":0,"text_scalars":0,"bytes":0},"truncated":false}
```

This is a shape fixture with a fixed example mount, not a live screenshot.

Targeted recovery uses gui_snapshot with exactly mode artifact, mount, artifact,
version, offset and length. Offset is a nonnegative uint64 scalar offset; length
is an integer 1–4,096. Resolve the exact mount/artifact/version before reading.
Return exactly view, mount, artifact, version, offset, text, next and end, where
next is offset plus the returned scalar count and end is true at retained-text
EOF. Offset equal to size returns empty text and end true; offset beyond size
refuses. A changed version refuses; never splice two versions behind a cursor.
The result remains within 65,536 bytes, including worst-case JSON escaping.
Report unavailable if the text is no longer retained. Retrieval grants no file
access and cannot recover bytes the application never owned.

Tts_queue takes exactly `{}` and reads the actual selected Page and its speech
service. Return exactly view, mount, ready, available, reason, items, omitted and
truncated. Items have text, rate and state (queued, waiting or active); include
at most 64 items, each text at most 1,024 scalars, and cap the complete canonical
result at 32,768 bytes using the same truthful omission policy. Omitted is a
nonnegative item count or null. Reason is empty or a safe local speech code.
Unavailable coordination must not be reported as a working empty queue. Another
Page's speech may cause this view's waiting state but does not expose its text.
These are playback telemetry, not proof that anyone heard or understood audio.

## 12.5 Act through the Page that owns the control

gui_click takes exactly mount and control. Resolve the ID in that mount, require
an enabled supported button/tab/toggle, then use its owned action. gui_input
takes exactly mount, control and text; text is at most 4,096 Unicode scalars.
It supports the prompt and explicitly registered ordinary text/number/select controls,
using their existing validation. gui_submit takes exactly mount and mode, with
mode prompt or hint, and operates the current draft through Page's corresponding
ordinary submission path. Hint while idle has the inherited refusal. Explicit
prompt can enqueue a later turn while the Agent is active. No tool accepts
arbitrary CSS, JavaScript, URLs or simulated privileged browser gestures. A
constructed event does not become a trusted user action through dispatch; the
[DOM event rule](https://dom.spec.whatwg.org/#dom-event-istrusted) distinguishes
those origins. Supported Page actions do not promise clipboard, fullscreen or
other privileges that require a real browser gesture.

Programmatic input has an explicit Page-owned origin. It does not assert human
typing and then depend on a later synthetic Enter to clear it. A human edit,
including a space, transfers origin to human and uses the inherited typing pause.
Refuse automation that would overwrite or submit a nonempty human-owned draft.
Check again at the point of effect: a human can type after RPC admission. An
empty human draft can be replaced; clearing one Page cannot clear another's
typing/speaking cause. Disposal resets transient action/disabled state before a
replacement Page acquires the same DOM.

Success has exactly view, mount, control, outcome and request_id. Control is null
for submit, otherwise the operated ID. Outcome is edited, applied or admitted.
Editing a prompt/settings draft returns edited after the owned draft changes;
that is no claim about saved preferences. A local tab change returns applied
after its state changes. A control that applies a preference/policy change waits
for its ordinary successful server acknowledgement before returning applied;
dispatch alone cannot claim it. Submit returns admitted with its accepted
request identity, or the acknowledged hint's active request identity, and does
not wait for that turn's model answer. Other successes have request_id null.
This lets a tool enqueue a later prompt on its own Agent without waiting for
itself to finish. Existing queue bounds and duplicate-admission rules remain.

Failure has exactly error and message, with static safe wording. Stable errors
are gui_unavailable, gui_stale_mount, gui_unknown_control, gui_disabled,
gui_unsupported_action, gui_human_draft, gui_stale_artifact, gui_unavailable_text,
gui_invalid_offset and gui_action_failed. Invalid arguments retain Chapter 11's
schema-refusal behavior and issue no action. Do not echo an invalid selector,
draft or raw exception in diagnostics. A server command refusal maps to
gui_action_failed with a safe explanation; its original accepted history stays.

Once a command has reached the server, a later timeout can leave its result
uncertain. Report that uncertainty rather than claiming rollback or resending
the command. Cancellation after a local click likewise cannot unclick it.
Ordinary model calls retain Chapter 11 Jobs/artifacts, so a human can inspect
the actual result independently of the model's account of it.

## 12.6 A skill enables a fixed integration

A tool description cannot decide that it should run before every model request.
Automatic sampling is an application decision with real effects on data exposure
and latency. Put that choice in the Agent's immutable creation configuration.

Declare a typed public integration policy, owned by Agent, with these serialized
fields. The same representation becomes session identity in §12.8. The policy is
an array, sorted by skill then connection then scope; duplicate triples refuse.
Each item has exactly skill, connection, scope, aliases and sources. Skill names
an existing frozen catalog definition, connection names a prepared logical MCP
connection, scope is a stable logical view name, and aliases is a sorted unique
array of bound aliases on that connection. Sources is an ordered array of
objects with exactly id, alias and arguments. IDs use the logical-key grammar,
are unique across the entire policy, and identify request observation sources.
Arguments is a copied JSON object validated against the frozen input schema.
Those arguments become durable request/identity data; keep credentials out of
them and in the caller-selected connection configuration instead.

Require skill mode, at most 32 policy items, 16 total sources, 64 KiB canonical
arguments per source and 1 MiB canonical policy bytes. Empty sources is allowed
for a control-only integration. Every source alias occurs in its item's aliases;
every alias belongs to the installed ceiling. Validate references before Agent
exposure. A policy changes neither the catalog nor the active grant union.

The shipped GUI policy is:

```json
[{"skill":"gui-debug","connection":"gui_main","scope":"main","aliases":["gui_click","gui_input","gui_snapshot","gui_submit","tts_queue"],"sources":[{"id":"screen","alias":"gui_snapshot","arguments":{"mode":"view"}},{"id":"speech","alias":"tts_queue","arguments":{}}]}]
```

The optional application validates these default observational operations.
Core does not hardcode the names or attempt to prove that an arbitrary handler
has no side effects. General embedders are responsible for selecting observational
aliases and arguments. Peer annotations confer no sampling rights. Even explicit
policy must pass current Skills visibility before each source call; an absent
grant produces observation_denied with no RPC. An inactive policy skill produces
no samples for that policy item.

Derive active leases from accepted Skills state. A lease associates this Agent's
integration work with its existing connection; it neither owns the whole service
nor supplies a second grant map. Local lease acquisition performs no network I/O
on Actor. Unload immediately prevents new sampling, cancels this Agent's affected
collection and fences its late replies, then releases its lease after that work
settles. Already admitted model-issued Jobs keep Chapter 11's independent
lifetime. Other Agents keep their leases and calls. Root retains the shared
physical connection; an exclusively owned logical endpoint closes only after its
last actual lease/work owner releases it.

Agent owns one context collector, whose worker reaches MCP through Agent and
Ensemble. Before each model attempt with active sources, Actor admits one ordered
collection only after the inherited pause and turn/request-limit gates allow
continuation. The worker performs normal MCP calls in policy/source order, with
two seconds per source and five seconds for the whole collection from admission.
These are absolute upper bounds, including delivery; no timer resets on progress.
Expired remaining sources receive observation_timeout without another call.

Automatic calls are not model-issued Jobs: they create no tool_called or
tool_result, consume no job handle or one-shot tool_limits, and require no fake
assistant call. They still share MCP's schema, message, operation-capacity and
cancellation machinery. An active model Job and a source collection cannot evade
the same connection's 64-operation/byte limits by using different entry points.

Actor remains responsive to hints, settings, interrupts and close while the
worker waits. A pause raised during collection cancels and discards it; after
resume a later collection starts fresh. Unload, interrupt or close does the same.
If skill or effective request-policy/configuration state changes, discard the
whole candidate before capture and collect again under the current generation.
Do not join an earlier screen to a later queue. A newly received hint need not
invalidate observations; it is captured in the eventual inherited receipt order.
Repeated concurrent edits may postpone admission, but at most one collection
worker and no accumulating retry queue may exist.

Expose an owned public integration-state getter and add required integrations
to WatchState, empty without policy. Its items have exactly skill, connection,
scope, active, collecting and sources in policy order. Sources contains only
id and alias in configured order. Active means the controlling skill is active;
it does not promise a healthy connection or grant every alias. Collecting is
true only while this Agent has an admitted collection for that item. Arguments,
observed text, source paths and diagnostics are excluded. Actor publishes
integration_changed with agent_id and the complete safe integrations array
when that projection changes, using the inherited watch revision boundary.
Offline state shows active from recorded Skills and collecting false. The
browser labels integration state alongside existing MCP/skill status; human
`/integrations` displays the same safe selection and collection state.

## 12.7 Record what this request observed

A screenshot taken after the answer cannot establish what the model saw before
it answered. The request must keep its own observation. That does not require
another conversation log or a durable queue waiting to be cleaned up.

Collector candidates are transient owned values. Before HTTP, Actor rechecks the
attempt, active source order, permissions, configuration generation and pause.
Render using those exact samples and current inherited material. Append the
ordinary request_sent, now with required observations for a policy-enabled Agent,
atomically with its existing hint/ephemera consumption. Only a successful append
admits the HTTP handoff. An append failure uses existing terminal storage rules.
No request event means no durable automatic sample and no later reuse. A failed
HTTP attempt retains the samples it tried to send, while the next attempt collects
anew. Observations never enter enduring dialogue or the primary system prefix.

Each observations item has exactly source, connection, alias, arguments, scope,
generation, result and error. The first five identify the configured source and
its exact fixed arguments. Generation is the producing Connection's positive
uint64, or null if no current generation exists. Result is Chapter 11's normalized
accepted object (content, isError and optional structuredContent), or null for
a local failure. Error is empty for an accepted result, including isError true;
otherwise use observation_denied, observation_timeout, observation_limit or a
Chapter 11 safe local error code. No arbitrary diagnostic text is stored here.
For GUI results the producing mount is inside structuredContent. Core does not
interpret it as a DOM capability.

Bound the complete canonical observations array to 131,072 bytes, including
identity, arguments, separators and result wrappers. Reserve a failure-form item
for every active source before collecting. Reserve the largest permitted safe
error spelling and uint64 generation representation. Creation refuses a policy
whose full source set cannot fit those failure forms. As ordered results arrive,
admit a
whole result only if the array still fits with all later reserved failures;
otherwise retain that source with result null and observation_limit. Do not
truncate an accepted structured value into invalid or misleading data. A 64 KiB
snapshot fits with its small default source metadata and the 32 KiB speech cap;
the exact whole-array admission test still governs custom callers.

Live public append validates observations against the actor-owned pending
candidate and current policy, in exact active source order with no duplicates,
missing sources or extra members. It cannot supply fabricated sampled data or
change source arguments. The internal capture path and public append share this
validation. Offline replay validates shapes, bounds and, when a v3 identity is
present, policy/source/grant correspondence at that recorded prefix. Offline
history cannot prove that a remote program told the truth; it never contacts that
program to try. Corrupt sample fields fail before reducer mutation.

The field is also accepted in standalone version-1 event logs for inspection:
absence means no automatic observations, and presence requires the strict sample
shape, limits and unique source IDs. Such a log has no session policy identity
against which to establish provenance. It grants no live integration authority.
New live standalone Agents with policy capture the same field using their own
creation configuration. Session v1/v2 retain their existing request shape and
refuse this field; policy-enabled sessions use v3. Existing standalone request,
ephemera and hint compatibility otherwise remains unchanged.

Pure public rendering accepts optional explicit owned observations for a
hypothetical request. It performs no collection or consumption; a bare-prefix
render has none. Exact historical reconstruction instead selects a request_sent
and its pre-event prefix, using that record's captured configuration, consumed
hint/ephemera selections and observations. Repeated reconstruction must produce
identical provider bytes with every browser endpoint disabled. Expose both
operations clearly; a preview is not a receipt for an HTTP request that occurred.

Render each item as one ordinary user text entry after all inherited request
material, including complete call/results, anchored skill/hint material, the
pending human prompt and remaining one-request hints/ephemera. Preserve those
earlier placements unchanged. Use `[observation ID]`, LF, the complete canonical
item JSON, LF, and `[/observation]`, with no trailing LF. Keep source order. This
labels external data without treating it as an instruction or hiding it inside
a tool result. There is no promised stable provider prefix or cache outcome.

For a compact public-embedding fixture, source screen calls an observational
status tool on connection panel with `{}` and returns one text block. This is a
generic source fixture, not the larger GUI snapshot schema. The neutral item is:

```json
{"source":"screen","connection":"panel","alias":"status","arguments":{},"scope":"main","generation":1,"result":{"content":[{"type":"text","text":"Settings selected."}],"isError":false},"error":""}
```

After an ordinary prompt `Inspect.`, the exact Messages user suffix is:

```json
{"role":"user","content":[{"type":"text","text":"Inspect."},{"type":"text","text":"[observation screen]\n{\"alias\":\"status\",\"arguments\":{},\"connection\":\"panel\",\"error\":\"\",\"generation\":1,\"result\":{\"content\":[{\"text\":\"Settings selected.\",\"type\":\"text\"}],\"isError\":false},\"scope\":\"main\",\"source\":\"screen\"}\n[/observation]"}]}
```

Chat Completions preserves the consecutive user entries:

```json
[{"role":"user","content":"Inspect."},{"role":"user","content":"[observation screen]\n{\"alias\":\"status\",\"arguments\":{},\"connection\":\"panel\",\"error\":\"\",\"generation\":1,\"result\":{\"content\":[{\"text\":\"Settings selected.\",\"type\":\"text\"}],\"isError\":false},\"scope\":\"main\",\"source\":\"screen\"}\n[/observation]"}]
```

generateContent merges them into ordered user parts:

```json
{"role":"user","parts":[{"text":"Inspect."},{"text":"[observation screen]\n{\"alias\":\"status\",\"arguments\":{},\"connection\":\"panel\",\"error\":\"\",\"generation\":1,\"result\":{\"content\":[{\"text\":\"Settings selected.\",\"type\":\"text\"}],\"isError\":false},\"scope\":\"main\",\"source\":\"screen\"}\n[/observation]"}]}
```

For Chapter 9's H/S/P held-batch fixture, retain its exact results→H→S→P
suffix, then append this observation text. The later request has results→S→P
and only its newly captured observations; H stays consumed. A source failure
uses the same envelope with result null and its safe error, visibly recording
what was unavailable instead of borrowing yesterday's successful screen.

## 12.8 Resume the authority, reconnect the view

A stable logical scope identifies the selected integration. A transient tab or
mount identifies one running instance. Confusing the two either prevents every
restart or lets history replay a click onto an unrelated replacement.

Use the following strict extension; the event log header remains version 1:

| Session selection | Checkpoint version / state_version | Creation identity |
|---|---|---|
| No MCP bindings or integration policy | 1 / 1 | Exact Chapter 10 identity |
| MCP bindings, empty integration policy | 2 / 2 | Exact Chapter 11 identity |
| Nonempty integration policy | 3 / 3 | Chapter 11 identity plus required nonempty integrations |

For v3, session_initialized.session has exactly version, session_id and identity,
with version 3. Identity has exactly mode, system, skills, handlers, mcp_bindings
and integrations. Integrations is §12.6's normalized policy and mode must be
`skills`. All Chapter 10/11 catalog, binding and installed-ceiling equality
checks remain. Compare every configured integration, including inactive skills,
by lossless semantic JSON equality; source order is meaningful. Changing a fixed
argument, source ID, scope, alias, connection or controlling skill refuses
session_incompatible. No old store acquires policy by automatic migration.

A v3 session_anchor.session has exactly version, session_id, origin_as_of,
origin_sha256 and high_watermarks, with version 3. Checkpoint/origin outer fields
otherwise retain Chapter 10's meanings, hashes and limits. Initializer or anchor,
checkpoint version, state_version and immutable origin all agree. Metadata
remains construction-only; public append cannot adopt a session or widen policy.
Explicit v1 payload rules remain as Chapter 11 teaches, including its absence
convention for the old initializer/anchor version field.

The strict documented semantic codec retains policy identity and each historical
request's captured samples wherever it retains that request's reconstructible
prefix/configuration. They are historical inputs, never a pending collector queue.
Snapshot-only import retains its existing absent-history limitation: it can
reconstruct represented requests only, and cannot manufacture discarded prefixes.
No saved state includes live leases, socket objects, callbacks, pause registrations,
pending GUI actions or native utterances. An unfinished attempt still refuses
live resume; interrupted transient collection adds no special recovery protocol.

Validate stored state and caller policy before serving/preparing a new endpoint
for resume. Then explicitly bind a browser with the same logical scope and
compatible discovered definitions; its physical URL, mount and generation can
change. Publish the Agent only after these phases succeed. Derive active leases
from restored Skills state; never replay past actions or samples. Offline
inspection requires no GUI, MCP server, catalog files or current credentials.
Chapter 8 preference/policy files remain separate settings authorities; they
cannot silently rewrite session integration identity.

## 12.9 Taking it for a spin

Actual implementation and runs are pending. These reproducible steps define the
exercise; they are not a transcript. Replace the planned outcomes with retained
source-bound observations after the student uses the interface.

Build the core and optional GUI commands above. In a fresh scratch workspace,
launch `/tmp/ensemble-ch12-gui --gui-debug --gui-view main --terminal`, using the
inherited provider configuration and explicit session-directory selector when
needed. Open the printed loopback URL before the wait expires. The page should
show waiting, then the bound Agent, ready MCP aliases and the gui-debug skill.
Record these actual states; a configuration file existing on disk proves none.

In the human PTY, ask the model to inspect which tab is selected, select another
supported tab, and inspect again. Wait for each answer and compare it with the
browser and accepted tool result before asking the next question. Exercise input
and submit with a harmless subsequent prompt, checking that the programmatic
draft does not strand typing pause and that its admission completes without
waiting for the future turn. A human-owned draft asserts typing pause, so a new
same-Agent model tool waits at the inherited gate. Record that gate separately.
To exercise the browser's gui_human_draft refusal, type a draft in the target
Page and use an explicitly scoped independent observer Agent or public action
consumer to attempt the overwrite. Preserve issuer/target identities, the refusal
and unchanged draft without clearing the target pause. A controlled action
already admitted before the human edit can separately prove the race-time recheck.

Use snapshot to read applied preferences and a bounded artifact preview, then
retrieve omitted text through its versioned targeted read. Make a real bounded
preference change and observe its server acknowledgement and subsequent snapshot.
Read the actual TTS queue while speech is enabled, distinguishing queued/active
telemetry from captured sound. Preserve Chapters 7–8's native audio evidence
requirements for changed speech behavior; do not call a text queue a listening
test. The following chapter adds that separate user's perspective.

Keep the pause gates during this exercise. The same Agent may be unable to begin
a new queue-reading call while its own speaking cause is active. Where active
telemetry needs concurrent observation, configure an independent observer Agent
or public consumer explicitly for that target view's endpoint. Its selected
scope may be another Page, but that is caller-granted access, never default
cross-Agent visibility. Record which Agent issued the call and which Page was
observed; leave the target's pause untouched.

Use the public consumer's Chapter 9 actor-ordered typed unload operation for
gui-debug, observe that automatic samples stop, and invoke its typed load operation
to restore it. Alternatively, ask the model to call unload_skill and load_skill
with `{"name":"gui-debug"}`, retaining the actual tool results. Human `/skills`
only inspects state; this exercise introduces no load/unload slash command.
Compare the actual outgoing request bodies and replay them exactly with the
endpoint disabled. A model can ignore delivered
context; delivery and successful task completion need separate evidence. Preserve
unavailable or timed-out samples and any corrective prompt instead of scripting
a successful account after the fact.

An external embedding application must reuse public BrowserApplication/view,
Page and server components with two Agents and selected views. Run one real model
task through the second view; verify that unloading/closing the first Agent's
integration leaves the second usable and ordinary watch traffic alive. Exercise
explicit rebind after a disconnect, stale-target refusal and compatible session
restart without replaying an old action. Repeat the chapter's model-driven
features on Messages, Chat Completions and generateContent, within an approved
bounded checklist. Local fault fixtures supplement these observations.

Retain source/binary identities, sanitized configuration, terminal exchanges,
browser actions, raw tool replies, applied file/UI outcomes, exact request bodies
or neutral reconstruction and usage. Capture actual screenshots with accessible
descriptions. Do not publish credentials, unsubmitted human drafts or invented
Bill participation. Preserve the initial student review and actual attempt before
historical comparison; later fixes keep their own evidence identities.

## 12.10 Checks that distinguish a browser from a fixture

The inherited Chapter 13 grader supplies fake stdio tools and searches request
text. It can establish parts of a protocol conversation while the shipped browser
integration remains disconnected. Keep its historical diagnostic, then test the
boundary it never reaches: the delivered browser code using the real WebSocket.

| Required property | Independent distinguishing control |
|---|---|
| Bootstrap | Endpoint before Agent; timeout/no browser and incompatible resume produce zero model calls; aliases cannot arrive late |
| Ownership | Public custom-view consumer; logical close preserves physical watch and peer channel; root joins a stopped peer |
| Framing | Real browser round trip above 65,536 bytes through bounded tunnel; an equally large ordinary command still refuses; malformed base64/generation cannot fall through |
| Bounds | Actual encoded-byte queue caps, stopped sender/consumer, cancellation burst, no pending-work growth after repeated canceled calls |
| Request identity | Out-of-order fresh requests and completions pass; duplicate pending/settled requests have one effect; unseen/settled cancel leaves no tombstone; 65th disjoint range faults only its route |
| Scope | Two views, no broadcast, stale mount/control refusal, hidden credentials/drafts excluded, script-looking text remains text |
| Human draft | Same-Agent typing gate demonstrated separately; explicitly scoped peer/public action refuses overwrite; already-admitted action rechecks after human edit |
| Effects | Selected/expanded/checked state agrees with actual DOM; programmatic input/submit works without clearing human/peer pause; policy success requires server ack |
| Speech observation | Same-Agent pause gates retained; explicitly scoped independent observer reads the target's actual active queue without access to other Pages or canceling its speech |
| Recovery | Exact omission or explicit unknown; targeted Unicode slices and changed-version refusal; complete metadata at the byte cap |
| Collection | Current grants, one worker, pause/unload/config races discard whole candidate, no Job/limit consumption, no stale reuse |
| Capture | Append failure means no HTTP, failed HTTP retains samples, all three literal suffixes, hints/manuals preserve order |
| Replay | Duplicate/unknown/reordered/mismatched samples refuse safely; exact reconstruction with endpoint disabled; standalone limitations disclosed |
| Persistence | Strict v1/v2/v3 agreement, inactive-policy mismatch refused before preparation, compatible new mount, no historical effects |
| Compatibility | Earlier CLI, Skills, Jobs, MCP, watch, settings and native-speech gates retain their coverage |

For each failure assertion, retain a passing control and delete the protected
behavior to require the intended failure. Test default shipping configuration
as well as public embedding. Review every new constructor, helper and executable
for real parents and logger access. Final acceptance requires the actual spin,
independent historical comparison, resolved teaching feedback and complete prose
review, with the initial attempts preserved.
