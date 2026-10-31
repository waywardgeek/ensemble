# Chapter 20: Managing the Swarm

Chapter 19 gave an agent the ability to spawn children, wait for their submissions, and answer their questions. A parent could send a task downward and receive a result upward. That is necessary, but it is not yet a usable fleet.

A long-running parent may supervise several children while a human watches another conversation. The human needs to see what each agent is doing, select a conversation without taking ownership of its work, and create named agents that survive an application restart. The parent needs finer controls: inspect a child, send steering, queue a question, interrupt a turn, retrieve a saved result, or remove a child deliberately.

These capabilities must compose. A GUI that displays another agent's conversation but sends input to the old one is broken. A tool that can inspect any globally named agent has escaped its authority. A stored result that revives an old task's completion after restart is not reliable persistence.

This chapter builds the C# management layer on the owners from Chapter 19. It does not introduce a fleet manager, Agent factory, callback table, or second live registry. `Ensemble` owns its Agents; a delegated Agent also has its actual parent Agent. The Common declarations expose those relationships and the functional partials implement their operations.

## 20.1 Activity Is Not Completion

"Busy" is useful but incomplete. Is the agent waiting for a model, executing a tool, or observing other work? Conversely, "idle" does not prove that it submitted a result.

Keep the two questions separate:

```csharp
public enum AgentActivity { Idle, Busy, Stopping, Stopped }
public enum WorkPhase { Idle, Thinking, ToolCall, Waiting }

public sealed record WorkActivity(
    WorkPhase Phase, string Tool = "", string Detail = "");

public sealed record AgentWorkStatus(
    AgentActivity Activity, bool Paused, bool Submitted, string? Failure,
    WorkActivity? Work = null);
```

These are value-like observations, not owners and not another scheduler. `WorkStatusAsync` asks the Agent's actor for current scheduling state. It includes direct Engine sends as busy work. Closed lifetimes use synchronized shutdown state rather than racing actor-owned fields.

An Agent derives its work phase from the operation actually in progress:

| Phase | Meaning |
|---|---|
| `Idle` | No current work phase to display. Inspect lifecycle and submission separately. |
| `Thinking` | A model request is in progress. |
| `ToolCall` | A tool is executing. The label is its operation name, not its arguments. |
| `Waiting` | The operation is observing a Job, child, question, or graceful drain. |

Foreground work takes display priority over simultaneous control work. Retain only the last published activity projection to suppress duplicates; do not maintain a second mutable busy flag. Short fixed descriptions are enough. Copying tool arguments or provider text into activity labels would create another disclosure path.

Activity notifications are transient and lossy, like streaming progress. They do not enter canonical History or the next model request. Reliable status comes from the actor. Shutdown does not announce idle merely because cancellation began: owned cleanup must reach its actual boundary first.

A child may have submitted and still be draining its final tool result or save. It may be idle with a retained failure and no submission. It may be paused with queued work and therefore still busy. The UI and every management tool must preserve those distinctions.

## 20.2 One Ensemble, Real Owners

The earlier Go implementation introduced a manager and supplied a factory to avoid package import cycles. Those are historical implementation choices, not the C# design. C# partial declarations and implementations compile into the same assembly and module. Common declares the vocabulary; functional directories supply the bodies. There is no reason to reproduce a factory indirection just because an earlier language needed one.

The Ensemble's existing `Agents()` collection remains its sole live membership collection. A delegated child also appears in its parent's `Children()` relationship. Constructors establish both sides; destruction joins strongly owned work and removes both sides through their declared operations.

The Ensemble can capture an ownership tree:

```csharp
public sealed record AgentTreeNode(
    Agent Agent,
    ImmutableArray<AgentTreeNode> Children,
    string? Address = null);
```

The actual Agent reference is identity. `Children` captures membership; it is not a cached scheduling state. Capture each collection under its own lock and release that lock before descent. Query status and Engine meters afterward. A tree snapshot is not globally atomic: a captured Agent may detach before a later status query, and a new child may appear on the next refresh.

Do not hold ownership locks while awaiting an actor, sending a WebSocket frame, or reading a file. The GUI is a consumer of the ownership model, not its owner or its scheduler.

## 20.3 Homes Are Not Save Paths

There are three different locations to distinguish:

1. The Ensemble's data directory, which contains shared peer launch configuration.
2. Each Agent's own data home, which contains its children and artifacts.
3. A History save path, which may be explicitly placed somewhere else.

I require every true top-level actor, including the initial host actor, to have a sibling home beneath the Ensemble directory. The first displayed actor is not the parent of the other top-level actors.

```text
ensemble-data/
  ensemble.json
  settings.json
  agents/
    root/
      save.json
      save.json.journal.jsonl
      jobs/
      children/
        researcher/
          save.json
          submit.md
          children/
            specialist/
    writer/
      save.json
      jobs/
```

The journal is shown because Chapter 14's persistence still applies; a snapshot alone is not the entire persistence story. File existence depends on which operations have actually run.

`Common/Storage.cs` declares Ensemble and Agent homes independently of History. `StartPersistence` does not assign an Agent home. An explicit History path such as `sessions/review/save.json` must not silently move its owner's children or `submit.md`.

The CLI and standalone GUI accept `--data-dir` / `EN_DATA_DIR` and `--agent-name` / `EN_AGENT_NAME`. The actor name defaults to `root`. Explicit flags win. The container defaults to the configured sandbox root, otherwise the working directory; a relative explicit container follows that base. The host creates its initial named Agent under `agents/<name>/`. Default save, event/diagnostic logs, and Job output follow that home; explicit paths retain their meanings and remain subject to sandbox policy. The standalone GUI's shared settings belong in the Ensemble container.

Persistent names are nonblank local components, not paths: reject `.`, `..`, separators, and invalid filename characters. Do not invent an ASCII-only naming restriction. Unnamed transient library Agents remain supported, but they do not acquire a fabricated global identity. Library callers may give those Agents an explicit home. They may not override the assigned home of a named Agent in a configured Ensemble.

This is a construction contract, not a migration command. Do not move a running actor's files to make a new directory diagram look correct.

## 20.4 Persistent Named Peers

A named peer is an Ensemble-owned top-level Agent. It is not a delegated child, so ordinary chat does not trigger the child's mandatory-submission correction loop.

The initial host actor remains host-configured. It has the same sibling-home placement as other top-level actors, but is not duplicated in `ensemble.json`. The registry contains peers created through the persistent peer operation.

Persist immutable launch choices rather than the running Agent's mutable configuration:

```json
[
  {
    "name": "writer",
    "model": "chosen-model-id",
    "skill": "ensemble",
    "created": "2026-10-28T00:00:00+00:00"
  }
]
```

This is the C# format: an array with exact `name`, `model`, `skill`, and `created` fields. A later human model switch must not rewrite a peer's original launch model when somebody creates another peer. Do not persist credentials, current tool grants, a second copy of History, or a snapshot of all runtime services.

The public operations take an explicit current top-level source Agent:

```csharp
// root is already constructed by this host and belongs to this Ensemble.
var ensemble = root.Parent();
Agent writer = await ensemble.CreatePeerAsync(
    root, "writer", model: "chosen-model-id", skill: "ensemble",
    cancellation: cancellation);

// At a later startup, after constructing the host's initial Agent:
// await ensemble.RestorePeersAsync(root, cancellation);
```

Null or empty model/skill overrides inherit from that source. Snapshot its current immutable runtime configuration and parsed skill templates, then create independent mutable Skills and MCP relationships for the peer. The Ensemble retains no defaults-Agent pointer. Current launch policy and credentials come through current runtime ownership, not stale saved authority.

Serialize registry operations with one gate. Validate the complete registry before restoring any peer: missing files mean no peers, but malformed, inaccessible, duplicate, or damaged entries must not silently become an empty registry. Build peers through their ordinary owners, restore their own History, initialize their selected skills, and save initial History. Only atomic replacement of `ensemble.json` commits a new peer's successful creation. Failure joins and destroys newly constructed live peers without deleting prior conversations.

Restoration creates registry peers, not their saved delegated descendants. Old child directories remain data until an explicit operation uses them. They are not instructions to respawn a tree.

Ordinary application shutdown preserves registry and files. Explicit `DeletePeerAsync` refuses retained unfinished children before closing the target's admission, joins teardown, removes the registry entry, and detaches the live Agent. Saved files remain. If registry replacement fails, retain the stopped registered peer for a retry. Cancellation can stop a request before irreversible teardown; after shutdown begins, join to a definite outcome rather than abandon half a deletion.

This deliberately changes v2 Chapter 21's deletion policy, which also removed the peer's directory. Deleting a peer now removes it from the running and restorable fleet without erasing saved work that may still need inspection. Retained files are historical data, not permission to recreate an Agent: restoration follows the registry, so a deleted peer does not return on restart.

## 20.5 Addresses and Authority Are Different

A local child ID is meaningful only to its owning parent. Two branches may both contain a child named `tests`:

```text
root/build/tests
writer/review/tests
```

Trusted GUI lookup needs an Ensemble-relative address. Derive it from actual parents, with canonical percent-escaping for each local component. Never reconstruct identity by searching for a matching leaf name or taking a filesystem basename. An address is neither a file path nor a guarantee that the same object will still occupy that address after recreation.

`Agent.Address()` returns null for unnamed transient roots and their descendants, and for detached Agents. `Ensemble.FindAgentAddress` follows existing roots and child collections. It does not create an Agent, grant ownership, or keep one alive. The GUI may give its borrowed initial Agent a connection-local empty key; that is not a global ID for every unnamed root.

Model-facing tools have a narrower rule: resolve a live requested child with the calling Agent's `FindChild`. Do not use the trusted global resolver as a fallback. This governs status, messages, output waits, interrupts, shutdown, kill, and selected-child change waits. Historical submission retrieval uses the same parent's owned namespace: the live child's actual home when present, otherwise that parent's declared `children/<id>` directory. It does not require a live child and never searches another tree.

A valid but unavailable ID and an elsewhere-owned ID receive the same ordinary refusal. A positive owned-child control is essential: an implementation that always says "unknown" has not passed an ownership test.

These are process-internal authority checks, not OS isolation. They prevent one tool handler from reaching across the ownership tree. File and process confinement still come from Chapter 18. Arbitrary trusted code sharing the same process has not been made into an adversarial security principal by adding a lookup rule.

## 20.6 The Management Tools

Chapter 19's seven tools remain. The shipped `ensemble` skill now also grants:

| Tool | Contract |
|---|---|
| `get_sub_agent_status` | Current direct-child lifecycle, phase, submission/failure, and Engine usage snapshot. |
| `send_message` | Live steering when busy; a new ordinary turn when idle. Return finalized output since an optional cursor without waiting for that turn to finish. |
| `ask_agent` | Queue a separate question and await that accepted turn's own reliable text reply. |
| `wait_for_agent` | Wait for finalized output after a child-local History cursor, or return when the child is no longer busy. |
| `wait_for_agent_change` | Observe a fixed selected set for progress, questions, removal, timeout, or apparent stalling. |
| `get_submission` | Read a direct child's historical artifact freshly, without constructing or completing a child. |
| `interrupt_agent` | Deliver `now` or `after_tool` through the child's existing actor controls. |
| `kill_agent` | Skip graceful draining; cancel and join teardown, then destroy the selected subtree. |

These remain grant-controlled tools. Registering a handler does not grant it. The actual shipped manual must enable them; passing tests with a custom skill is not enough.

### Status and usage

`get_sub_agent_status` combines one actor status snapshot with one immutable Engine meter snapshot. It reports selected model separately from measured response models, tiers, and billing identities. Unknown counts or prices stay unknown. No response has completed yet is different from a completed response whose cost is unknown: an empty accumulated meter may total zero, while a row for an unpriced model makes the accumulated estimate unknown.

Do not copy counters into a fleet manager or reprice old rows from the currently selected model. Status is not a globally atomic accounting/lifecycle barrier; each actual owner supplies one observation.

### Interrupt, graceful shutdown, and kill

`interrupt_agent` accepts `now` or `after_tool`. Its acknowledgment means the control was delivered, not that stopping or cleanup has finished. The child remains owned and reusable, and independent Jobs retain their existing lifetime semantics. Cancelling the observing tool cannot retract an already admitted interrupt.

The existing library `ShutdownAsync` is forceful. Do not rename the same cancellation path "graceful" in a tool description.

The tool `shutdown_agent` first awaits the existing foreground `FlushAsync` barrier, checks cancellation, and then performs joined forceful teardown and destruction. `kill_agent` skips that barrier. This matches the useful distinction in the earlier implementation without adding another scheduler or a special reserved-closing state.

Flush waits for current and queued foreground work. It does not resume pause, wait for independent Jobs, or itself save History. A paused child with queued work can therefore keep a graceful waiter pending. Cancelling that waiter leaves the child alive. Once final teardown starts, join it without abandoning cleanup; it cancels independent Jobs and descendants and preserves saved files. Cleanup errors are still errors, even when every relationship was detached correctly.

## 20.7 Finalized Output and Reliable Replies

Output polling must not consume the GUI's shared lossy observation queue. Use canonical History snapshots and its durable sequence numbers instead.

`ReadOutputAsync` projects finalized assistant text from `ResponseEnded`, plus text from `ToolReturned` and `ToolObserved`. It excludes private reasoning, human/child hints, and streaming deltas. Its cursor advances across all durable events, including events with no public text. The cursor is a child-local History sequence, not a byte offset or a fleet-global counter. Negative or ahead-of-History cursors refuse.

```json
{
  "agent_id": "researcher",
  "cursor": 42,
  "output": "The finalized text since the previous cursor.\n",
  "activity": "idle",
  "paused": false,
  "submitted": true,
  "failure": null,
  "phase": "idle"
}
```

This illustrates the output shape, not a promise that idle implies submitted. They are independent fields.

`send_message` admits `Message` through the actor, so it becomes steering during active work or a new turn when idle. An incoming message wakes current Job observations without killing the Jobs and without creating a sticky interrupt for future observers, as established in Chapter 19.

`ask_agent` instead uses a separate `SubmitAsync` admission and waits for that turn's reliable reply. It does not consume an older assistant final from a shared cursor. It has no cursor parameter and does not automatically fetch the structured submission payload. Use `get_submission` for historical typed data and `join_agents` for current selected-child completion.

`wait_for_agent` waits for new finalized text or a nonbusy status. It is not a submission wait and not a quiescence guarantee. Two independent output observers can read the same committed events. Observer cancellation leaves the child's work alive. If destruction clears the live History, an already-running observer retains its last immutable snapshot long enough to report the observed output and closed status; that local readback is not another persistent transcript.

## 20.8 Wait for Change, Not for a Timer Alone

A supervisor may need to react before a child completes. `wait_for_agent_change` is a cancellable snapshot loop, not another scheduler or event subscription framework.

At admission, capture the requested direct children. Omitted or empty `agent_ids` means all current direct children. Reject duplicates and unavailable IDs. Later siblings do not silently enter this wait, and a newly created child with the same local name does not replace the captured object.

Capture each child's current History sequence and transient-progress counter before the first status await. Otherwise progress during that await could disappear into the baseline. Then compare current status, committed History, transient progress, and the caller's selected-child question inbox. The implementation uses short, cancellable 100 ms waits between snapshots.

The returned reason can be `changed`, `pattern`, `question`, `stopped`, `removed`, `stuck`, `timeout`, or `empty`. Questions are observed, not consumed. A pattern is a bounded .NET regex over new finalized output, not over private reasoning or arbitrary stream fragments. Defaults are 300 seconds overall and 15 seconds without observed progress; zero disables the respective deadline. Invalid or nonfinite durations refuse.

One Agent counter advances only when an accepted `PartDelta` is published. This avoids calling an actively streaming child stuck while its finalized History is unchanged. Status polling itself must not increment progress. The counter stores no content, does not consume observations, and does not duplicate History.

"Stuck" means busy with no observed status, History, or delta progress for the requested interval. It is a supervisory hint, not proof of deadlock or provider failure. A paused child with queued work can meet that definition. Timeouts return an ordinary observation result; cancelling the Job detaches the observer and does not kill the selected children.

### Historical lesson: a refused Job could hang its observer

An earlier Go implementation registered a Job with a tool's shared cursor before starting the worker. If the concurrency limit then refused the Job, it closed one completion channel but left the shared cursor unfinished. `Observe` waited for an event that would never arrive, and `join_agents` hung until the grader timed out.

The lesson is still current: every admitted observer needs a completion, cancellation, or refusal path. "Refused before execution" is a real lifecycle outcome. In C#, keep that responsibility in the existing Job owner; do not add a fleet-level workaround that merely hides a stranded Job.

## 20.9 Durable Submission Is Historical Data

A parent's context may be compacted. A child may be stopped and removed. The useful result should remain retrievable without reconstructing the child, starting skills, or granting it new authority.

C# keeps one immutable accepted artifact at the child's `submit.md`. Despite the filename, its body is a strict JSON envelope. Chapter 19's schema, exact-number validation, and generation rules still apply. Chapter 20 writes version 2, adding the owning parent's nullable qualified address; version 1 remains readable without provenance it never recorded.

Publication validates the payload, checks current generation and lifetime, atomically replaces the artifact through the actual Sandbox, and only then exposes live completion. The envelope has a 2 MiB bound in addition to the existing stricter schema/payload limits. It is not an append log and not a second mutable output cache.

`RetrieveSubmission` and the `get_submission` tool always read freshly. With a live child, use its actual home, including an explicit library-configured child home. Without one, use `children/<id>/submit.md` beneath this parent's declared home. There is no directory registry for rediscovering a removed child's former custom location.

Validate the exact envelope, local child ID, schema, payload, and available parent provenance. A named parent's version-2 artifact must match that parent address; a copied artifact from another tree is not accepted merely because its leaf ID matches. Version-1 data remains explicitly unprovenanced. Address provenance is relative to this Ensemble, not a cryptographic signature or globally unique identity across unrelated installations.

The tool reports `historical: true`. Reading old bytes does not set `Submission()`, clear a failure, advance a generation, or satisfy a join. Accepting new work invalidates old live completion even if the previous file remains on disk. Restoring History likewise does not revive old submission as new success.

This deliberately differs from a memory-first fallback. Fresh reads make external file replacement visible and avoid a second authoritative cache. Filesystem failures return bounded diagnostics rather than exposing an absolute private path.

## 20.10 Browser Selection Does Not Transfer a Relay

A single-Agent GUI could bind every socket to one conversation. A fleet GUI needs two identities per connection:

- **Display selection:** the Agent whose replay, status, partials, and user controls this tab currently shows.
- **Fixed relay Agent:** the Agent whose browser MCP connection this socket supplies.

Changing the display must not transfer the relay or invalidate an in-flight browser tool call. The connection frame communicates the relay identity independently of replay. A dedicated peer-relay tab opens a new socket for that peer. Simply selecting the peer in the original tab does not give it a browser.

The frontend keeps its current display selection only after the server acknowledges it with replay. It rejects stale conversation frames by field presence and exact ID, including the connection-local empty ID used by an unnamed borrowed initial Agent. Other unnamed roots remain nonselectable rows rather than sharing that empty key.

### Per-Agent projection, per-connection selection

Keep one observation reader and one transient partial buffer per viewed Agent. Otherwise several tabs would compete for the same lossy observation queue or leak one Agent's partial into another conversation. Each connection has its own selected Agent and durable replay cursor.

A coalesced tree refresh captures actual ownership, then queries status outside ownership locks. It includes the mounted root and nested descendants. The tree drives the picker and current work display; it is not another live registry. A captured row may disappear before the next command, so admission rechecks the selected target.

Prompts and actor controls target the display selection. MCP replies remain associated with the fixed relay. Shared GUI preferences remain host-scoped; selected-Agent tool and context limits are read from that Agent. Account and peer-management routes retain their explicitly documented host ownership.

If a displayed peer is deleted, the connection receives replay of the mounted Agent before its display selection changes. Its relay does not transfer. A stale tree snapshot must not reset a newly selected, still-live Agent; recheck current ownership before treating an absent row as deletion.

### Replay admission is the selection boundary

Replay must be bounded and atomic with respect to connection selection. The C# host queues replay as one serialized frame using the existing bounded, nonblocking outbound queue. Build the prospective replay, attempt admission, and only then publish the new selected Agent and cursor. If admission fails, close or reject the connection rather than leave a partial transcript labeled as a successful switch.

The queue is bounded by frame count; do not describe that as a separately enforced byte-size cap. One large replay and many small live updates are different risks. The implementation must preserve its actual bound and truthful failure behavior, not invent a stronger claim because the demo looked responsive.

The browser resets transcript cards and stops old-view narration on switch. Replay is silent. Speech associated with the old view must not send a stale resume command to the newly selected Agent. Structured snapshots and muted browser checks establish routing and rendering, not audible quality or human accessibility.

## 20.11 Browser-Dependent Startup and Trusted Peer Controls

The standalone GUI starts serving while restoring peers. A browser-dependent peer can then appear in the tree while waiting for its own relay, allowing the operator to open the explicit relay tab and complete startup.

Waiting is opt-in: peer creation/restoration defaults to prompt missing-browser refusal. A headless CLI must not hang on a registry it cannot service. Hosts that can accept independent browser connections pass `waitForBrowser: true`. The MCP owner waits on its actual browser relationship; it does not infer availability from display selection or retain another mutable "browser connected" flag.

`EnsureMcp` serializes lazy ownership so skill startup and an arriving socket cannot construct two managers. Waiters hold neither Agent ingress nor the MCP gate while awaiting attachment. Cancellation detaches the waiter; Ensemble destruction cancels and joins pending peer operations. A browser supplied by one Agent cannot release another Agent's wait.

Peer management uses guarded mounted HTTP POST requests, not a long-running operation inside the socket receive loop:

- Create accepts a name and optional model and primary skill; blank overrides inherit from the mounted source Agent.
- The request may remain pending while the tree and relay socket remain usable.
- Cancel aborts the request and joins startup rollback. A pending peer must not be announced as durably created.
- Delete uses the existing registered-peer operation, refuses unfinished children, and retains files.
- The mounted Agent is protected because the caller promised its lifetime for the mount.

These routes use exact-origin/loopback checks, the dedicated custom header, bounded exact JSON fields, no CORS, and noncacheable responses. Raw storage and startup exceptions do not become page text. A guessed route from an untrusted website is not authorized merely because it targets localhost.

Borrowed `ServeAsync` still does not own its caller's Agent or ASP.NET application. It neither restores the fleet nor stops unrelated routes. Cancelling the mount joins its own requests and relay teardown before returning. The caller must keep the Agent alive until that join finishes.

## 20.12 Failure Lessons and Review Order

The original Go chapter uncovered useful failures. A reused settings filename replaced the named-agent registry with UI preferences. A generic fallback instantiated an unknown selected skill and reported success for the wrong Agent. A string-based test harness truncated structured object submissions. A replay queue exposed a switch before it had admitted the new transcript. Preserve those lessons, but do not present them as failures observed in this C# implementation.

The C# adaptation added its own concrete lessons:

- Serving concurrently was insufficient for browser-dependent startup because ordinary MCP claims deliberately refuse a missing browser. The correction was an explicit owner-based wait for startup, not weakening ordinary activation.
- Headless restoration needed to retain prompt refusal, rather than inherit the GUI's deliberate wait.
- Display switching could not destroy and recreate the relay: that would invalidate active MCP work. Display and relay identity are separate.
- A saved result needed nullable parent provenance without inventing IDs for unnamed roots, and retrieval had to remain separate from live completion.
- A background-tab browser fixture could wait on animation-frame polling even when both pages were responsive. Foregrounding the fixture page corrected the test; it was not evidence of a product rendering defect.

The executable grader also needed correction. Before a response completes, a zero aggregate meter is legitimate; unknown pricing applies once an unpriced response exists. A nested fixture that blocked inside the fake provider's globally serialized route prevented the parent's releasing request from running. Moving the gate to a real child question tested the intended lifetime instead of the fake's lock. Neither correction justified changing the product.

Review complete bounded source, not guessed filenames or clipped output. Read declarations and actual call paths, not just the code that a test happens to reach. Passing tests do not establish that Common owns all declarations, that every relationship has both ends, or that destruction attempts every owned cleanup operation.

## 20.13 Verification and Scoring Scope

The original seven-row Go grade remains available on its original route. The C# executable-only adapter retains these original rows and weights:

| CLI-observable row | Points |
|---|---:|
| Activity observed | 15 |
| Ownership enforced | 15 |
| Selected-child change wait wakes | 15 |
| Working composition root and nested workflow | 15 |
| Bottom-up submission and idle correction regressions | 10 |
| **C# executable scope** | **70** |

A 70/70 result passes this executable subset of the original rubric; it is not whole-feature acceptance. The named-peer persistence row (15) and GUI tree row (15) are excluded from this score, not silently passed by source review or unrelated unit tests. C# persistence, mounted hosting, independent browser selection, peer controls, and startup have their own real-component, actual-host, and isolated-browser evidence. They remain necessary review and exercise obligations even though this CLI adapter does not award their points.

From the repository root, build the actual CLI and the grader:

```sh
dotnet build agent/Cli/Ch04.csproj
go build -o /tmp/grade-edition2 ./cmd/grade-edition2
```

Pass the executable host, not the DLL alone. If this shell's installed runtime cannot run the apphost, supply an executable wrapper that invokes the approved SDK's `dotnet` with the built lowercase `ch04.dll` and forwards `"$@"`.

```sh
/tmp/grade-edition2 -ch 21 -executable-only \
  -fleet-skills "$PWD/agent/Skills" \
  "$PWD/agent/Cli/bin/Debug/net10.0/ch04"
```

The adapter supplies a loopback provider, owned working directory, shipped skill activation, correlated JSON control messages, and joined EOF shutdown. It holds model work to inspect activity, tests positive owned-child access and negative cross-tree access, observes a running change wait before releasing progress, checks fresh cursors and queued replies, and exercises recursive questions and bottom-up submission. Retained wire records distinguish finalized output from submission and current completion from historical files.

These local checks do not verify live-provider entitlement, OAuth, billing, native-platform behavior, audible speech, or human accessibility. The earlier failures and the scope of each passing result remain in the progress record.

## 20.14 Exercise

Implement and inspect the management layer through the existing owners:

1. Declare lifecycle/work-phase values, ownership-tree snapshots, homes, peer operations, and child-output contracts in Common. Keep persistent state there and functional bodies in their responsible directories.
2. Derive current activity from the actor, and accounting from Engine. Verify thinking, tool, waiting, paused, submitted-but-draining, idle failure, and shutdown boundaries without a duplicate scheduler flag.
3. Give all top-level actors sibling homes. Separate those homes from History save paths. Persist strict immutable peer launch choices and restore peers without respawning descendants.
4. Derive canonical addresses through actual parents. Keep global trusted GUI lookup separate from direct-child tool authority, including repeated local IDs and unnamed transient roots.
5. Add the management tools and the real shipped grants. Test an owned success as well as cross-tree refusal. Demonstrate that cancelled observation leaves underlying work alive.
6. Verify graceful shutdown versus kill with held real work. Do not assume a prompt acknowledgment proves completed teardown, and do not resume paused work as a side effect of waiting.
7. Read finalized output with History cursors. Test multiple observers, no duplicate output, delta-only progress, a queued question's own reliable answer, and fixed selected membership.
8. Publish bounded versioned submissions before exposing completion. Replace the saved file and demonstrate a fresh read; destroy the child and retrieve without reconstruction; admit new work and prove old bytes cannot satisfy its join.
9. Run two GUI connections with different selections. Switch one while a browser RPC is pending; finish the RPC through its original relay. Open a dedicated peer relay and verify selecting a conversation alone did not supply it.
10. Create and delete peers through the guarded mounted routes. Cancel pending browser startup, refuse unfinished-child deletion, retain files, and leave the mounted caller's host and Agent alive after mount teardown.
11. Preserve the previous endpoint, inspect the assembled source and prose, run the applicable grade, and preserve a source-only runnable endpoint only after that review. A grade is evidence, not permission to skip the architecture.

### Human testing

Use a fresh owned data directory and browser profile. Keep an existing working Ensemble and its browser untouched. Watch a child move through thinking, tool execution, and waiting; then switch between two conversations and confirm that messages and controls go to the selected one. Check that late old-view fragments do not contaminate the new transcript and that replay does not narrate old content.

Create a named peer, choose its model and skill, close the application cleanly, and restore from the same directory. Confirm its own conversation survives, the initial actor remains separately host-configured, and delegated children were not silently recreated. Delete the peer and verify that it no longer restores while its saved data remains available for deliberate inspection.

If a selected skill needs browser tools, open the explicit independent relay. Also try cancellation before the relay arrives and a headless launch against the same requirement. The first must join rollback; the second must refuse promptly by default.

Then assess legibility, speech, and practical supervision with real work. Protocol checks cannot tell you whether the fleet is comfortable to manage.

### Engineering takeaway

A useful fleet is not a larger collection of callbacks. It is a set of understandable owners with precise operations: admit work, observe progress, preserve results, and destroy relationships completely. Keep facts where they belong—activity on the Agent, accounting on Engine, transcript on History, process lifetime on Job, launch configuration on Ensemble—and make every management operation travel through those actual relationships.
