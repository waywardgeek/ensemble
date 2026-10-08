# Chapter 8: Preferences that do something

The settings panel says speech is off. The next answer starts talking.
Somewhere, a Boolean was saved successfully. The person listening has a more
useful definition of successfully.

The first edition found a related failure after fixing input validation.
The server corrected a value to zero, then omitted that zero from its state
broadcast. The browser kept displaying the rejected value. The same encoding
could not represent speech being turned off: false disappeared too. The
[historical account](chapter-08-evidence.md) records both failures. Saving,
applying and displaying a preference are different jobs. A settings panel can
get two of them right and still be wrong where the reader notices.

This chapter makes the interface worth keeping open. Chat and actions get
separate panes; preferences survive a restart; every tab learns what changed.
An execution setting reaches the actor that actually enforces it. The screen
must show the applied value, and the running program must use that value at
the promised boundary.

**Reviewed contract:** ownership, persistence and speech semantics are
coordinator-accepted working choices. The student handoff still requires an
accepted Chapter 7 and the published independent checker. No Chapter 8 implementation,
passing checker, browser session or audible result is claimed.

## TL;DR

Extend `solutions/edition-2/main/` after the coordinator releases the accepted
Chapter 7 baseline. Read this chapter, the architecture ledger and the entire
`book/edition-2/skills/ensemble-coding/SKILL.md` before editing. Do not consult
the first-edition answer or author research. A validated `ch08/` is a frozen
export, never a second development tree.

1. Reuse the optional GUI module's Connector and ArtifactScroll in a sidebar,
   chat pane and actions pane. Preserve all Chapter 7 safe rendering, identity,
   expansion, reconnect, pause, streaming and public embedding behavior.
   Expose controls through semantic HTML and keyboard interaction.
2. GUI Server owns a persistent display-preference service: theme, font size,
   pane widths, automatic speech enablement and speech rate. Each page still
   owns its input, current/queued speech, cancellation generation and pause
   registration. Core Agent code imports no GUI preferences or transport.
3. Agent owns a separate execution-policy service and its creation-only
   persistence path. This chapter exposes only the maximum model requests per
   human turn: zero selects the existing default of 16; positive 1–256 selects that
   limit. Actor-ordered updates affect turns activated after application;
   an active turn retains its captured policy through every continuation.
4. Define sparse patches separately from complete snapshots. Omission preserves
   a value; explicit false/zero writes it. Reject unknown fields, null, wrong
   types, duplicate keys and out-of-range values before any mutation. Validate
   persisted input with the same schema. Never serialize credentials or Config.
5. Persist a complete versioned candidate before applying and acknowledging it.
   Each owner has its own revision and update serialization. A stale revision,
   busy writer or failed write returns a safe correlated error. No transaction
   across GUI preferences and Agent policy is promised. Publish complete applied
   snapshots, including false and zero.
6. Add the wire commands and snapshot fields below. A reconnect receives a
   current snapshot followed by newer changes in each domain, without a missed
   handoff. Retain Chapter 7 connection bounds and default CLI protocol.
   A settings acknowledgement is never a prompt completion or a pause update.
7. Shared “Autoplay new answers” controls future automatic enqueue. Capture the
   locally applied preference revision and rate when each utterance is queued.
   Existing queues retain their entries and rates. Cancel speech is local to
   the page; a preference patch cannot clear another tab's pause causes.
8. Verify behavior as well as JSON: two-tab changes, false/zero, failure atomicity,
   restart, actual speech, keyboard resizing and active-turn policy isolation.
   Then use the real browser, human CLI and public consumers on all three APIs,
   preserving initial attempts before independent comparison and review.

Build the CLI from main with `go build ./cmd` and the GUI command from
`main/gui` with `go build ./cmd/ensemble-gui`. The historical diagnostic is
`make grade-dir CH=9 DIR=solutions/edition-2/main`; its old layout and wire
assumptions do not grade this new contract. The coordinator must publish the
independent Chapter 8 checker invocation before student grading. Section 8.9
specifies required coverage; a partial checker does not waive it. Public method
names, storage implementation and visual styling remain student choices.

## 8.1 Give each setting an owner and a consumer

A font size has no reason to enter Engine. A model-request limit has no reason
to be read from a browser's settings store. Putting both in one convenient JSON
object makes the first screen easy and the second client difficult.

Server owns a display-preference service inside the optional GUI module. It
holds the applied preference value, its revision and its file identity. Shared
GUI declarations may live in that module's common package; implementation
belongs in the responsible GUI package. The service and its persistence helpers
retain owner access through Server and its public application parent to the
logger. No callback bundle supplies missing reachability.

Agent owns execution policy. Its shared values and owner interfaces belong in
core `internal/common`; policy persistence belongs in a responsible core spoke.
The actor reads and changes the Agent's authoritative policy through that
owner chain. A private service implementation may live in its spoke with an
Agent interface back-pointer. It cannot import llm or the GUI, and Engine does
not acquire a closure that secretly reads the GUI's copy.

The public policy interface works in a headless application. Server forwards a
policy request through that interface and presents the result. It cannot change
Agent fields itself. The existing general configuration path keeps its earlier
active-turn restrictions; the new policy-update operation is the one explicitly
specified live transition here. Model selection, credentials, workspace, log
identity and unrelated configuration do not become editable browser settings.

A page owns what is happening on that page: unfinished input, speech work,
selected tab, scroll position and the live pause registration. The persisted
preferences supply shared presentation defaults and future enqueue policy.
They do not become owners of those transient activities. This split is the
coordinator's working design under the existing architecture rules, not a new
claim that every application must share preferences in exactly this way.

## 8.2 A patch needs to say whether a field exists

The display-preference snapshot has exactly these fields:

| Field | Default | Accepted values |
|---|---|---|
| `theme` | `dark` | `dark`, `light`, `system` |
| `font_size` | 16 | integer 12–28 CSS pixels |
| `sidebar_width` | 260 | integer 180–480 CSS pixels |
| `actions_width` | 380 | integer 200–640 CSS pixels |
| `autoplay` | false | Boolean |
| `speech_rate` | 1 | finite number 0.5–2 |

The stored execution-policy value contains exactly
`max_model_requests`, an integer 0–256. Zero selects the named default of 16;
its stored value remains zero. The UI displays “Default (16)” for zero,
rather than implying the Agent may send no requests. The safe snapshot also
reports the effective value 16 or the positive configured value.

A patch contains any nonempty subset of its domain's fields. An omitted field
is unchanged. A present field is a proposed value even when it is false or
zero. Do not decode a patch into a zero-filled snapshot and then guess which
zero the sender meant. Presence can be represented with dedicated optional
fields or validated raw members; the choice is left to the student.

Reject duplicate JSON object keys, unknown fields, null values, strings used as
numbers, fractional integer fields and values outside the table. Boolean is
not a number. Reject the whole patch if one member is invalid. The same valid
ranges govern a complete file loaded at startup. This chapter uses rejection,
not silent clamping: the control keeps showing the applied value and displays
a useful reason for refusing the proposed one.

The earlier zero-broadcast defect still matters under rejection. Turning
`autoplay` from true to false and resetting a positive request limit to zero
are valid writes. Every complete snapshot includes all fields, without an
omit-empty encoding. An invalid submission may leave the user's draft text in
the editor, but the page must distinguish it from the applied value. Do not
label unsaved text as the current setting.

No file path, API key, endpoint, authorization header, entire Agent Config or
provider payload appears in these snapshots. Diagnostic messages identify a
static field or failure reason without printing the invalid file or message.

## 8.3 Save, apply, then announce

Each domain has a nonnegative integer revision, initially 0. A successful change
increments it once. A valid patch producing the same complete value acknowledges
the current revision without a write or change broadcast. Revisions are local
to their domain: preference revision 4 and policy revision 4 are unrelated.
Persisted revisions survive a restart; they are not Chapter 7 watch revisions.

The GUI command adds `--preferences PATH` and `--policy PATH`, defaulting to
`.ensemble/gui-preferences.json` and `.ensemble/agent-policy.json` beneath its
launch workspace. Resolve these paths once without changing process cwd.
Create absent parent directories for these operator-selected paths. Reject
using the same resolved path for both domains or giving two live owners the
same path in one application. The demo is a single-process writer; concurrent
independent applications sharing the files are outside this chapter's promise.
Paths are creation-only and never accepted from a browser command.

Core Agent construction accepts an optional policy path through its public
configuration. Without a path, a headless Agent starts at the default policy
and supports the same updates in memory; its snapshot explicitly says
`persistent:false`. With a path it uses the file lifecycle below and reports
`persistent:true`. Server preferences always have the configured file path.
Neither snapshot reveals the path. A GUI restart may reuse these settings
files while starting a fresh conversation log; settings persistence does not
promise conversation or job resumption.

The complete on-disk forms are:

```json
{"version":1,"revision":0,"preferences":{"theme":"dark","font_size":16,"sidebar_width":260,"actions_width":380,"autoplay":false,"speech_rate":1}}
```

```json
{"version":1,"revision":0,"policy":{"max_model_requests":0}}
```

A missing file means these defaults. It need not be created until a change.
An existing file must be one complete UTF-8 JSON object, no larger than 64 KiB,
with all required fields, a supported version and a nonnegative revision.
Reject unknown/duplicate fields, trailing data, wrong types and invalid values.
A malformed, unsupported or unreadable file fails startup with a safe reason;
do not overwrite it with defaults. The complete-file parser does not treat
omitted fields as a patch. No migration format is invented for first-edition
settings files.

For an update, validate and merge into an owned candidate using the supplied
base revision. Only one write per domain may be in flight. A second change
while that writer is busy gets `settings_busy`; a stale base gets
`revision_conflict`. Do not queue unbounded preference writes behind a slow
filesystem or silently rebase stale values. The independent domains can work
concurrently because they own different state and files.

Write the candidate to a new private temporary file beside the destination,
check the complete write, sync and close it, then atomically replace the
destination. Treat successful replacement as the persistence commit. Failures
before it leave the applied snapshot and revision unchanged, publish no change,
and do not truncate the old file. Remove an uncommitted temporary file on the
ordinary failure path. This promises whole-file replacement and checked writes;
it does not promise recovery from every filesystem or power-loss failure.

Disk work runs outside the actor and transport read loops. Its owned worker
returns success or a safe error to the authoritative service/actor. Only then
apply the candidate, publish its full snapshot and acknowledge the request.
Until application, reads and new turns use the old applied state. An update
acknowledgement defines the point after which a later read/turn must see the
new value. No lock needed by controls, job facts or another client spans disk I/O.

Close refuses new writes and joins an already started writer, resolving its
actual result. It cannot turn a completed replacement into a claimed rollback.
If the connection disappears or the process stops between replacement and
acknowledgement, the sender may not know whether its edit committed. Reconnect
or restart reads the authoritative snapshot; do not resend the old patch
automatically. An observer delivery failure does not undo a committed setting.

These rules apply separately to each owner. The UI sends separate commands for
preferences and policy. There is no “Save everything” operation that pretends
two files and two authorities commit atomically.

## 8.4 Put the applied values on the wire

Extend Chapter 7's subscribed connection with two commands. Command IDs retain
that chapter's connection-local correlation rules:

```json
{"type":"preferences_update","id":"p1","base_revision":0,"patch":{"autoplay":true,"speech_rate":1.5}}
{"type":"policy_update","id":"p2","base_revision":0,"patch":{"max_model_requests":2}}
```

Both require a nonnegative integer `base_revision` and nonempty object `patch`.
They remain subject to the existing message limit, origin check, strict command
validation and subscribe requirement. A preference/policy update cannot submit
a prompt, set another client's pause, or select a different Agent.

Register a preference subscription and capture its owned current snapshot at
one serialized service boundary. Deliver that snapshot before any newer
preference changes on that connection. The initial record is:

```json
{"type":"preferences_snapshot","revision":0,"preferences":{"theme":"dark","font_size":16,"sidebar_width":260,"actions_width":380,"autoplay":false,"speech_rate":1}}
```

Send it before Chapter 7's Agent snapshot sequence. Defer subsequent preference
changes until after `snapshot_end`, keeping that group contiguous; the bounded
queue and overflow/resync rule still apply. The two snapshots describe separate
owner boundaries, not one atomic transaction across domains. The Connector marks
settings ready only after both initial domains are known. Later changed records use
`type:preferences_changed` with the same complete payload and increasing
preference revisions. The sender gets the change too, followed by:

```json
{"type":"preferences_ack","id":"p1","revision":1}
```

A no-change acknowledgement needs no duplicate broadcast. These messages do
not consume Agent watch revisions. Preference subscription and socket queues
inherit Chapter 7's bounds and close/resync behavior; they cannot silently
lose an applied update while the page claims to be current.

Preference messages belong to the current Connector connection even though
they have their own revisions and arrive before the Agent snapshot generation.
Bind their handlers to that connection's lifetime. A late callback from a
replaced socket cannot overwrite the new connection's settings. The new initial
preference snapshot establishes its applied state; do not mix a partial handoff
with messages from the abandoned connection.

Agent policy uses the existing atomic watch. Add `execution_policy` to safe
snapshot state, containing `revision`, `persistent`, `max_model_requests` and
`effective_max_model_requests`. Add `active_max_model_requests`, null while
idle or the active turn's captured effective limit. A public headless getter
returns the same owned policy value without creating a browser or watch.
After application the actor publishes the next watch revision with:

```json
{"kind":"policy_changed","agent_id":"a1","execution_policy":{"revision":1,"persistent":true,"max_model_requests":2,"effective_max_model_requests":2}}
```

The existing observation envelope carries generation and watch revision.
Publish the change before returning the policy update acknowledgement:

```json
{"type":"policy_ack","id":"p2","revision":1,"watch_revision":43}
```

A no-change policy acknowledgement uses the current policy revision and latest
watch revision. A watch joining before or after application sees the change
in its tail or snapshot, respectively. Never emit a policy change by bypassing
the actor or by directly calling browser connections.

Preserve Chapter 7's transport boundary: invalid JSON/UTF-8, binary, oversized
or non-object messages, and unusable command IDs close the connection. A JSON
object with a usable ID receives a correlated error for a correctable command
or settings failure and leaves the connection usable. Codes are `invalid_command` for
invalid command shape, `invalid_preferences` or `invalid_policy` for invalid
values, `revision_conflict`, `settings_busy`, `settings_persist_failed` and
`settings_closed`. Messages give a useful static reason such as
`speech_rate must be between 0.5 and 2`. A conflict also supplies `domain`
(`preferences` or `policy`) and `current`, the complete safe snapshot for that
domain with its revision: `{revision,preferences}` for preferences, or the
`execution_policy` object shown above for policy. A page can show the applied
state and ask the reader
to retry; it must not quietly overwrite a competing update.

Two settings do not conflict merely because they use different fields. They
conflict when their base revision is stale. If tabs A and B both send at
revision 3, only the first applied candidate advances to 4; the second gets busy
while it is pending, or a conflict afterward. A deliberate retry against 4
merges the new patch into 4. This prevents an old whole form from erasing a
newer change its user never saw.

Do not change the default CLI machine protocol or inject these GUI records into
its output. The existing `--terminal` client continues using the same Agent.
Headless consumers obtain policy acknowledgements through the public API.
Settings commands are controls, not model requests, and incur no model call.

## 8.5 A budget the actor actually reads

The old interface once saved 200 while its execution paths still stopped at 16.
The settings store and broadcast worked. The setting did not. The new budget
must be read at the decision that could send the next request.

Preserve the existing counting unit: model requests per human turn, including
automatic continuations. It is not the number of tools, response fragments or
jobs. The default remains 16. On activation, the actor copies the then-applied
policy into that turn's immutable work state. A request admitted to the queue
earlier but activated after an update uses the new policy. A turn already
active keeps its old value, even while paused or waiting for HTTP or tool output.

An update accepted while a turn is active does not cancel it, create another
model request or change any already-sent body. The page shows both the next-turn
policy and active limit while they differ. Other model/configuration updates
retain their earlier restrictions; this chapter grants no general permission
to mutate in-flight configuration.

Before model request N+1, enforce the captured limit N. If response N contains
calls, finish that entire accepted batch under the existing pause, pairing and
interruption rules, then end with `round_limit` before another HTTP request.
If response N finishes the turn without calls, complete normally. A policy
update cannot turn already accepted effects into a rollback. Independent jobs
retain their Chapter 4/5 lifetimes.

Record the capture in each new `turn_started.turn.policy`:

```json
{"revision":3,"max_model_requests":0,"effective_max_model_requests":16}
```

Validate the three fields, their ranges and their relationship on append and
load. Historical turn-start records without policy keep the earlier effective
limit 16 and unspecified revision; do not invent a historical update. A present
malformed policy is an error. This record explains which rule governed the
turn; replay does not execute turns or change today's policy file. It introduces
no vendor request field and does not change reconstruction of historical HTTP
bodies. The same capture reaches the browser through active watch state.

A persisted policy belongs to the configured Agent. Two Agents can use separate
files and different limits; one Agent's update cannot affect the other. A public
snapshot is an owned value. Mutating it cannot change the stored policy or a
turn already running. Prove this through a public consumer, not a package-private
setter that skips the actor.

## 8.6 Reuse the screen instead of forking it

The wide layout has a sidebar, a chat pane and an actions pane. Chat contains
human input, hints, answer and exposed thinking cards. Actions contains tool
proposals, accepted calls, results and job updates. Lifecycle and pause status
remain visible regardless of the selected sidebar tab.

Use the reusable ArtifactScroll twice. Route individual typed parts and events;
a response with text and a call must not disappear into whichever pane received
the envelope first. Keep Chapter 7's full identities for replacement and
correlation, missing-context labels, bounded previews and full-text expanders.
Both panes use the same safe rendering machinery. A result containing markup
does not become safer because it moved to the right side of the screen.

Give the two dividers pointer and keyboard operation, visible focus and an
accessible label and current value. Arrow keys move the associated requested
width by 10 CSS pixels within its range. A completed pointer drag or keyboard
change sends the preference patch; do not write a file for every mousemove.
A pending control can be disabled until its acknowledgement, provided the
reader can still type, interrupt and cancel speech.

Stored widths are desired widths. Constrain actual layout to the available
viewport so the central input remains usable; a smaller viewport may stack
panes instead of forcing horizontal scrolling. Viewport clamping does not
silently rewrite saved values. After enlarging the window, the saved preference
again supplies the desired width. The exact breakpoint and proportions are
styling choices, but all controls and full retained content remain reachable.

Theme uses shared CSS custom properties for dark and light. `system` follows
the browser's current color preference and updates while that preference
changes. Font size affects readable content and controls in both panes.
Do not rebuild cards, replay speech or reset the conversation just to restyle
it. A remote preference update applies the same theme/font/width values to
other tabs once received.

The sidebar exposes Agents and Settings as labeled controls with selected state.
Show the current Agent name and authoritative lifecycle; keep a tree-shaped
presentation suitable for later children without inventing sub-agents now.
The Settings panel separates Appearance and Speech from Agent policy. Use
actual input values and checked/selected/expanded state, not only CSS classes.
An automated observer and a person using assistive technology need to inspect
the same truth.

The reader should be able to keep an earlier tool result open while a new
answer arrives. Preserve scroll position and focus as Chapter 7 requires.
A two-pane layout that steals focus twice per token has reused the wrong part
of the old interface.

## 8.7 Shared speech defaults, locally owned speech

Label the shared control “Autoplay new answers”. Its help text states the
existing automatic scope: new visible answer text, exposed thinking and concise
tool summaries; tool results and replay remain silent. An explicit card speaker
action still works while autoplay is off.

At each automatic enqueue decision, use the newest preference snapshot already
applied by that page's serialized controller. Capture its revision and speech
rate with the queued utterance. A queued or currently speaking utterance keeps
those values. A later rate change affects future enqueues, not a sentence
already waiting or sounding. Network delay can make two tabs observe a new
revision at different times; this rule describes the exact local boundary
instead of promising simultaneous changes across sockets.

Turning autoplay off prevents subsequent automatic enqueues in every tab after
that tab applies the update. Drop any not-yet-enqueued automatic text buffer
at that point and advance its consumed position so turning autoplay on later
does not recite the backlog. Already queued/current utterances remain owned by
their page and finish normally. Automatic text received while off is marked
seen without being queued; a later final cannot speak it a second time.

Cancel speech is a separate local action. It clears that page's work, advances
its cancellation generation and reconciles its own pause causes exactly as
Chapter 7 teaches. A preferences broadcast never sends a pause request on
behalf of another tab. Turning off autoplay does not mean an occupied queue
became empty, and a rate change cannot release a typing cause.

For example, A and B each have an utterance queued under preference revision 4.
A disables autoplay, committing revision 5. Each page applies 5 and stops adding
new automatic speech, while its existing revision 4 utterance can continue.
A then presses Cancel speech: only A's queue and speaking cause clear. B stays
paused until its own queue ends or B cancels. If B is also typing, ending its
speech still leaves the aggregate pause true. The UI shows current speech
separately from the shared autoplay checkbox, so “off” is not presented as
“this page is now silent”.

A newly connected page loads the persisted default but never replays history
into speech. If synthesis requires a local user activation or is unavailable,
show that status and let the reader enable that page's playback explicitly.
Do not queue work and retain a speaking pause indefinitely while playback is
blocked. The actual speech and failure receipts remain distinct from a browser
report that an API exists.

## 8.8 Taking it for a spin

**Actual Chapter 8 evidence is pending.** The following is the required exercise,
not an invented successful session. Use the actual browser and human CLI with
real models, scratch workspaces and environment-held credentials. The coder may
drive them; do not claim that Bill personally ran them.

Start the GUI with explicit preference/policy paths and a fresh conversation
log. Open two tabs. Change theme, font size and a pane width in one; observe
the controls and rendered result in the other. Use the keyboard divider and
inspect its exposed value. Close and restart with the same settings paths but
a new log, then confirm that applied values return. Do not count a browser's
cached input as server persistence.

Enable autoplay and actually listen to or capture synthesized output. Change
its rate for a later utterance. Use both pages to exercise the example in §8.7:
existing speech survives a shared disable, new automatic speech does not start,
and Cancel on one page leaves the other's queue and typing cause intact.
Use deterministic speech callbacks for exact revision/cancellation races, and
label them separately from audible results. If the environment cannot produce
actual speech, leave that feature's live gate open.

Give the Agent a bounded scratch-file task through the browser, then inspect
its proposed calls, artifacts and final answer in their respective panes.
Try a correction while streaming, inspect pause acknowledgement before the next
tool admission, and interrupt an active turn. Reconnect once during work and
once after completion; preferences, policy and conversation must converge
without replaying a prompt or speaking historical text. Reuse the public GUI
components in the embedding example rather than copying the application.

Set the model-request policy to 1, ask explicitly for one file read followed by
a report, and inspect what the model actually requests. When it requests the
read, the accepted tool batch completes, the turn stops at `round_limit`, and
there is no continuation HTTP request. Restore zero/default of 16, repeat the task
and observe the continuation. If a model declines the requested tool or answers
without it, retain that outcome and use a bounded corrective prompt; do not
claim it exercised the limit. The longer-than-16 proof is a local scripted
control, not a reason to buy many unnecessary live requests.

Exercise the same current Agent through `--terminal`; a limit reached there
retains the existing CLI failure/cleanup behavior. A public headless consumer
uses two Agents with different policy files and demonstrates isolation and
restart. All three APIs need actual browser task/policy and public behavior
as introduced, with the full live plan checked before paid runs. Record exact
source, binaries, selected/returned identities, inputs, visible outputs and
usage. Preserve unexpected model behavior and the corrective follow-up.

## 8.9 Checks that can distinguish a working preference

| Contract | Required distinguishing check |
|---|---|
| Presence and snapshot | True→false and positive→zero survive acknowledgement, broadcast and fresh process load; omitted field stays unchanged |
| Validation | Unknown/duplicate keys, null, wrong kinds, fractional integers, ranges, oversize file, trailing input and unsupported file version fail before mutation |
| Persistence | Nondefault positive survives restart; failure before replace preserves original bytes/revision/state; no-change patch performs no write |
| Concurrency | Two same-base patches produce one change and busy/conflict; deliberate fresh-base retry retains the first change; controls work while persistence is held |
| Subscribe handoff | Hold a change at the snapshot boundary; each client gets either the old snapshot plus newer change or the new snapshot, with no gap or regression; old-socket callbacks cannot overwrite the new connection's preferences |
| Runtime effect | Limit 1 pairs its final tool batch and prevents HTTP 2; limit 17 permits HTTP 17; zero restores 16; settings-consumer deletion fails |
| Turn timing | Change policy during held HTTP and while a turn is queued; active turn retains its capture, queued turn uses policy at activation; two Agents remain independent |
| Recorded policy | New capture validates/replays, malformed capture fails, old absent capture retains historical 16; request reconstruction gets no invented vendor field |
| Speech scope | Enqueue revision/rate captured; off discards only unqueued automatic buffer; existing queues remain; one page's cancel cannot clear another page or a typing cause |
| Browser behavior | Theme/font/width visibly change; dividers work by pointer and keyboard; applied versus unsaved values and selected controls are inspectable |
| Card reuse | Mixed text/call response reaches both panes, finals replace correct partials, malicious content stays inert in preview and expanded views |
| Compatibility | Chapter 7 reconnect/watch/pause/queue bounds survive; CLI protocol stays exact; optional GUI/public/headless consumers keep their module boundary |

Use a valid passing fixture before every mutation. A restart check that writes
the default theme proves little; seed a distinct value, restart and compare the
new process's applied snapshot and actual rendering. A server echo cannot prove
a browser control or an execution policy has a consumer.

Retain the student's first implementation, actual user runs and teaching review
before the independent reviewer reads the historical standard. Give the student
rationale from that comparison, not old answer code. The author updates missing
teaching, the student checks the response, and the reviewer verifies revisions.
A pleasing screenshot and a green settings-roundtrip check are both useful.
Neither closes the other obligations in this chapter.
