# Chapter 12 outline: The Agent sees its interface

**Stake:** the reader needs the Agent to report what its interface actually did,
including what it cannot observe, instead of diagnosing the intended interface
from source code or inventing a result from an incomplete snapshot.

Preparation, October 8, 2026. New Chapter 12 maps to first-edition Chapter 13,
“The Agent Sees Itself.” The corrected global map and Chapter 11 review retain
that numbering. This is an outline for architectural review, not a complete
student contract. Chapter 11's accepted contract direction is at 0c82feb, with
final wording at 39bc526; predecessor implementation/checkpoint and a new checker
remain prerequisites. No Chapter 12 implementation or actual spin is claimed.

## Story and teaching order

Keep the old chapter's concrete distinction between reading intended behavior
in source and observing a running screen. Preserve two documented mechanisms:
programmatic input once tripped a typing pause that submission did not clear;
a snapshot that omitted selected/expanded state led an observer to report the
wrong panel. The corresponding September 20 commits support the repairs.
Do not copy the old global flag or whole-document selectors into the new design.

The October 2 correction adds a harder lesson: the shipped gui-debug integration
had not actually connected the browser, while its grader supplied a fake stdio
skill. The new chapter must show the real WebSocket tunnel working from the
shipped entry point. The old chapter's ninety-eight-second debugger result,
blanket claims about unsupervised findings and speech-off story need stronger
receipts or narrower wording; do not repeat them as newly verified facts.

Proposed sections and their consequential facts:

1. A snapshot can describe a control yet omit its selected state. Explain the
   reader's diagnostic problem before the transport.
2. Boot the browser endpoint before freezing the Agent's aliases. A parsed
   debug configuration must become an actual usable connection.
3. One physical GUI socket can carry independent logical MCP channels. Closing
   one tool endpoint must leave ordinary observation and another Page alive.
4. Scope observation and controls to one explicit view, with observable state
   and honest omission/retrieval information.
5. Automated input uses the owned Page action path without claiming a human is
   composing forever or clearing another client's pause.
6. A skill enables configured context and leases; it cannot discover new grants.
7. Refresh observation at a request boundary and record the accepted sample so
   replay does not ask today's browser what yesterday's screen looked like.
8. Demonstrate the ordinary CLI/browser/public paths on all three providers.
   A separate blinded usability study belongs to the following chapter.

Voice: third person, reader/task present through the mechanisms, no invented
Bill dialogue. Use the old phantom-panel episode as an attributed historical
finding, then return to the exact state fields that prevent that blind spot.
Keep historical chronology in evidence rather than turning the chapter into a
commit tour. New actual failures and their corrections will supply the spin.

## Authority and scope

Bill explicitly requires MCP tunneled over WebSocket so the Agent can see and
control its GUI. Chapter 11 already requires the transport-independent public
seam. The implementation belongs in the optional GUI module; core neither
imports WebSocket nor inspects DOM. Use MCP 2026-07-28 with Chapter 11's exact
metadata, correlation, bounded schema/results and cancellation semantics.

The coordinator accepts the following working directions for outline development:

- Bootstrap an unbound browser view/endpoint before creating the Agent whose
  remote aliases must be frozen. Ordinary no-debug startup remains unchanged.
- Use typed, application-selected skill-to-connection/context policy in Agent
  creation configuration. Chapter 9 frontmatter stays unchanged. No callback bag
  or GUI type enters core common.
- A skill activation enables only policy inside that fixed ceiling. Agent leases
  distinguish its use from shared root connection ownership. Unload releases its
  lease and prevents later automatic-context admission; unrelated Agents retain
  their work. Scope logical close only to resources actually owned by that lease.
- Publish the persistence/version impact before implementation. No new fields
  appear silently inside Chapter 11's strict identity.

Required now: actual browser MCP endpoint and tunnel, useful snapshot and speech
queue observation, click/input/submit controls, shipped gui-debug activation,
configured automatic context, lease lifecycle, safe browser state, replay and
restart compatibility, and public composition. DOM-free microphone/listener
personas and a separate virtual-user evaluation harness remain the following
chapter's work. No arbitrary JavaScript evaluation, cross-origin frame driving,
credential editor, generic browser automation product or hosted MCP service is
added to obtain this narrower capability.

## Concrete boot proposal

The startup dependency is genuine: Chapter 11 discovers before Agent creation,
while the old GUI assumed an Agent already existed before its socket opened.
Give the browser surface an earlier lifetime instead of changing the grant rule.

1. The optional application creates Ensemble and GUIServer, reads typed caller
   configuration and validates any existing session through Chapter 11's local
   resume phase. Corrupt/incompatible stored state stops before MCP preparation.
2. In explicit gui-debug mode, serve a BrowserApplication with an unbound view
   scaffold. Its scoped DOM root and endpoint exist, but its prompt controls
   cannot submit to a nonexistent Agent. Show a waiting state and startup URL.
3. Browser connects to the same-origin GUI server and explicitly advertises the
   selected view endpoint. The server attaches it to the caller-selected logical
   connection key. No broadcast-to-all-tabs/first-answer selection is permitted.
   Duplicate claims for one live view are refused until the prior owner closes.
4. The public GUI transport constructor supplies a logical channel through the
   actual Chapter 11 Connection parent. Discover the browser's fixed tool
   definitions and validate the selected aliases. Preparation uses the normal
   protocol service, without an alternate GUI-only MCP parser in core.
5. Construct/resume the Agent with those frozen bindings and the selected
   integration policy. For the convenience flag, perform the ordinary typed
   gui-debug skill load before accepting a user prompt; failure is visible.
6. Mount that Agent's Page into the original scoped view and attach its watch,
   pause registration and optional human CLI. Then enable prompt submission.
   Endpoint identity does not change merely because its initially empty view
   acquired an Agent.

The full draft must give bootstrap a bounded, cancelable wait distinct from the
MCP preparation timer, and document what the operator sees without a browser.
Recommendation: 120 seconds from serving the URL, with clean root shutdown on
expiry; protocol preparation keeps Chapter 11's 30-second bound after attachment.
No paid request or silent headless fallback occurs while waiting.

The same boot path must work for a public embedding application and a resumed
session. A browser reconnect opens a new logical generation only after explicit
view rebinding and definition compatibility. A caller cannot attach gui-debug
to a previously constructed Agent that lacks those installed aliases. Use a
new explicitly configured Agent/session in that case.

## Owner and data plan

| Owner | Authoritative data/lifetime | Parent and boundary |
|---|---|---|
| Optional application/GUIServer | HTTP server, physical GUI connections, explicit view-to-endpoint selection | Existing public Ensemble owner interface; logger through actual parent |
| GUI physical Connection | Single socket sender/reader, bounded multiplexing queues and endpoint registrations | GUIServer |
| Core MCP Connection | Protocol state, IDs, pending operations and transport generation | MCP service → Ensemble |
| GUI MCP transport | One logical endpoint lease and its message delivery/close state | Actual MCP Connection; constructor receives it |
| BrowserApplication | Browser endpoint, view scaffolds, Pages and shared native speech service | Explicit document application owner |
| Browser endpoint/view | Its scoped root/control map, command correlation and mount generation | BrowserApplication |
| Page | Human/programmatic draft state, typed actions, watch, local queue and pause causes | BrowserApplication; no module-global programmatic-input flag |
| Agent | Frozen integration policy and binding identity | Ensemble |
| Agent context collector, proposed | Per-attempt sample work, owned leases and bounded results | Agent; reaches MCP through Agent → Ensemble |
| Skills / Actor | Existing authority / serialized admission, capture acceptance and persistence | Agent; no independent grants in the collector |

The transport's logical endpoint is a leased resource supplied through the
public constructor seam allowed by Chapter 11. It is not a sibling service bag
or an alternate parent. The physical connection retains its socket; closing the
lease detaches only the logical route and wakes its blocked operations. The
full constructor plan must expose both actual owner chains and logger reachability.
A fake parent whose only useful field is a closure back to a GUI singleton fails.

Core common contains transport-independent policy/sample values and owner
interfaces. GUI-specific wire envelopes/control descriptions live in the optional
module's shared vocabulary. Browser objects keep private state on instances.
Snapshot and control helpers reach their Page/view owner, including code that
otherwise appears stateless. Inspect actual route and parent behavior in the
external embedding example, not only package names.

## Tunnel and endpoint contract candidates

Use a dedicated GUI envelope identifying logical channel and generation, with
one complete MCP JSON message as the payload. It coexists with the existing
command/watch protocol. Both directions validate outer shape, channel ownership,
size and generation before dispatch. Chapter 11 parses the inner MCP envelope.
Publish exact outer names/bytes and bounds in the full draft; no newline framing
or raw JSON-RPC sniffing decides which handler receives a browser frame.

Route to one selected browser endpoint. A second tab must not receive a call
intended for the first, and multiple replies must not produce repeated effects.
A disconnected endpoint fails promptly rather than queueing a click for the next
person who opens the page. Never replay a sent click/input/submit after transport
loss. A stale generation cannot settle a new request or mutate a replacement view.

Keep one owned physical sender and bounded per-channel delivery. A channel that
closes or overflows fails its own operations; it cannot close the shared socket
while that socket is healthy. A physical failure closes every child route with
an explicit transport outcome. Preserve Chapter 7 resubscription and Chapter 11's
64 operation permits, including staged cancellation. Flood tests must show that
MCP traffic cannot create an unbounded queue behind ordinary watch traffic.

Cancellation is advisory. Before browser action admission, a canceled operation
has no effect. Once an action has synchronously changed the UI, a later cancel
cannot undo it. Async Page operations need their own completion and stale-result
fencing. Do not add an unbounded browser canceled-ID set; specify a bounded
request lifecycle under the same monotone-issued identity rule.

## Observe a specific view truthfully

Propose gui_snapshot, tts_queue, gui_click, gui_input and gui_submit as the
installed aliases. Snapshot and speech queue remain callable through ordinary
Jobs and may additionally be selected for automatic context by trusted policy.
The names themselves confer no read-only property; automatic eligibility is
explicit caller configuration validated against the fixed supported operation.
Peer annotations cannot opt themselves in.

Snapshot reads only the selected view's visible semantic interface: control
identity/label/type, enabled/disabled, selected, expanded, checked, current
applied settings, connection/pause state and bounded artifact previews. Include
state rather than inferring it solely from CSS labels. Exclude unrelated Pages,
hidden credentials, scripts and raw HTML. Unsubmitted human drafts need an
explicit policy; recommendation is presence/length only, while a tool-owned
draft can be read back for verification until a human edits it.

Selectors should be stable control selectors emitted by this view's snapshot
and resolved in its owned map. They do not grant arbitrary document-wide CSS
queries, JavaScript execution or navigation. Refuse missing, ambiguous, stale,
out-of-scope, disabled and unsupported controls. A remount invalidates old target
identities. The full draft must choose exact argument/error shapes and define
when a control changed between snapshot and action.

Proposed bounded default: at most 256 controls and 20 artifact previews, each
preview at most 1,024 Unicode scalar values, with a 64 KiB encoded snapshot cap.
Every cap reports exact omitted counts/bytes in structured metadata; truncating
the final string must not remove its own warning. Add a bounded targeted-read
form of gui_snapshot for retained artifact text, so a caller can recover what
was omitted without granting file access. Define cursor/version behavior before
checker work; a retrieval must not silently splice two changing artifacts.

Tts_queue reads the actual Page queue/service ownership chain, including captured
rate and queued/waiting/active state. It exposes this view's text only. Another
Page using native speech can be reported as contention without exposing that
Page's utterance. This is application telemetry, not evidence that audible speech
was heard or understood. Preserve native coordination and failures from Chapters
7–8; the next chapter separately tests the listener's actual channel.

## Control without trapping the pause gate

The old global programmatic-input flag is unacceptable with multiple Pages.
Use the same owned Page actions as the human controls, with an explicit action
origin and operation lifetime. A gui_input draft does not create a human typing
cause that would prevent the Agent from issuing its next submit/click. Human
keystrokes still create their ordinary registration cause, including a space.
Never clear another Page's typing or speaking cause to make automation proceed.

Proposal: refuse to overwrite a nonempty human-owned draft, including one created
after RPC admission; a human edit transfers the draft to human ownership. Submit
uses the Page's ordinary prompt/hint choice and acknowledgement path, not a
synthetic key event assumed to act like a browser gesture. GUI controls that
require trusted user activation can refuse visibly; do not claim synthetic
input has that privilege. Success distinguishes an action dispatched from an
acknowledged server change. A returned confirmation alone is insufficient for
an applied policy/settings claim.

A self-directed submit can enqueue later work on the same Agent under inherited
FIFO rules. Bound queued work as before and avoid waiting synchronously for that
future turn inside the current tool call. Closing/remounting Page removes its
listeners, pending actions and transient disabled state; late callbacks cannot
alter the replacement. The ordinary root/Agent/terminal EOF distinctions survive.

## Skill policy, automatic context and lease lifetime

Coordinator direction: keep Chapter 9's exact frontmatter unchanged. Add typed
Agent creation policy selecting skill ID, logical connection, fixed aliases and
ordered automatic sources. Validate all references against the frozen catalog
and installed bindings. A shipped gui-debug skill uses normal tools/dependencies;
application configuration supplies the connection/context policy. The default
example must actually prepare the browser, rather than relying on a fixture-only
mcp_servers declaration.

A skill transition derives this Agent's active integration leases from accepted
Skills state. It cannot change a handler definition or widen the installed
ceiling. Lease counts are runtime ownership bookkeeping, not a second permission
map; source admission checks current Skills through Agent. Acquiring an already
prepared connection lease is local work and cannot park the Actor on browser I/O.

Recommendation for unload: prevent new automatic sampling immediately, cancel
this Agent's in-flight automatic sample, discard its late response, and release
its lease after owned work settles. Already admitted model-issued Jobs retain
Chapter 11's independent lifetime. Other Agents keep their leases/calls and the
root keeps the shared physical connection. An exclusively owned logical endpoint
may close only when its last actual lease/work owner releases it; no callback
removes unrelated Registry entries or destroys the GUI socket.

Before each model attempt with active configured sources, gather a fresh snapshot
and TTS queue in policy order through the normal MCP service. Use one bounded
Agent-owned collector worker; the Actor still handles hints, interrupts, policy
changes and close. Proposed bounds: two seconds per source, five seconds total,
128 KiB accepted combined material. These are review candidates, not measured
latencies. Unsupported/disconnected source becomes explicit unavailable data;
never reuse an old screen as though it were fresh.

The exact durable capture protocol remains a decision before full drafting.
Recommendation: a typed observation-sample fact carries copied source/result
bytes and producing view/generation; request_sent names the exact samples it
used. Project them as one-request external data after complete call/results,
separately from enduring skill manuals. Repeated rendering is pure; replay uses
recorded samples with no browser. A canceled attempt must have an explicit
retirement/discard rule so its pending sample cannot leak into a later attempt.
Do not quietly overload Chapter 2's one-request ephemera with incompatible
consumption semantics or use a returned snapshot as system authority.

## Session and restart identity impact

The typed integration policy changes authority and automatic request material.
It belongs in creation identity, along with stable view scope and logical
connection selection. Physical socket IDs, tab instances, outstanding controls,
leases, native speech and DOM objects never become durable executable state.

Propose a strict version-3 identity/snapshot extension only for sessions using
the new policy. Plain Chapter 10 v1 and MCP-only Chapter 11 v2 retain their exact
formats and cannot acquire this policy during resume. The full draft must define
all outer/init/anchor version agreement, required policy fields, semantic hashes
and recorded sample state before a checker encodes them. Same logical scope,
policy, skill catalog and definitions at a newly selected physical browser may
resume after ordinary local validation and fresh compatible discovery.

Restart does not restore pending clicks, old DOM targets, old speech, pause
registrations or leases from history. Restore permitted current policy from the
matching creation configuration and derive its active leases from restored
Skills state. A corrupt/unfinished session refuses before remote preparation.
Historical rendering remains possible without any browser or policy source file.

## Required demonstration and independent checks

Every row remains a plan. Prepare provider-specific action/receipt bounds before
paid work; no result, screenshot or personal participation is invented here.

| Surface/action | Evidence required |
|---|---|
| Human CLI on all three APIs attached to the real GUI Agent | Model sees its own selected Page, describes actual tab/settings state, changes a permitted control and observes the result through the real tunnel |
| Browser model task | Actual snapshot/tool cards, applied control changes, responsive hint/interrupt, no programmatic-input deadlock; screenshot and accessible description |
| Skill load/unload/reload | Frozen declarations versus effective grant changes, source samples appear/disappear at the taught attempt boundary; other Agent's calls remain usable |
| Public optional-GUI embedding | Real bootstrap/discovery before construction, custom view reusing components, logical close preserving socket/watch and another Page |
| Two tabs/views | Explicit routing, no broadcast duplicate action, conflicting target refusal, stale generation fencing, unaffected speech/pause owner |
| Persistence | Captured provider bodies replay with endpoint disabled; compatible restart reacquires current endpoint without replaying historical UI actions |
| Usability limits | Omitted controls/artifacts are reported and retained text is retrievable; unavailable audio/telemetry is labeled honestly |

Local controls cover malformed outer/inner frames, size/work bounds, no browser,
failed preparation, unsafe state counters, paused human ownership, wrong/stale
selectors, deliberate script-looking labels, slow senders, cancel races, close
and remount. Tests must use the actual delivered JavaScript and real browser
socket for the tunnel boundary; a fake stdio server remains a protocol fixture.
Retain passing controls and intended deletion failures, plus predecessor gates.

Historical `make grade-dir CH=13 DIR=solutions/edition-2/main` remains diagnostic.
Its fake-server markers and source-based cleanup/parity do not prove a live GUI,
read-only versus mutating authority, lease isolation or owner-safe shutdown.
Publish a new independent command before student release. No grader/runtime
changes are authorized by this outline.

## Decisions for coordinator review before the full draft

1. Accept the bootstrap/view/lease proposal in concrete form, including explicit
   view selection, bounded wait and no silent Agent-first lazy registration.
2. Settle automatic sample admission, atomic durable capture/consumption/discard,
   unload race and ordering with inherited hints/skill material. These affect
   replay and request bytes, not merely API names.
3. Settle the strict policy identity/version-3 extension and what a compatible
   replacement browser means for stable view scope.
4. Settle the bounded snapshot targeted-read contract, stable selector identity
   and human/programmatic draft ownership. Avoid a generic automation API while
   making the required controls usable and honestly observable.

Bill's new first-edition sandboxing work remains an untouched pending lesson
source. No new chapter count or scope claim is derived from that unfinished work.
