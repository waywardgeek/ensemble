# Chapter 20: Managing the Swarm

A delegated child and a colleague are different relationships. A child has a parent, an assigned task, and a required result. A colleague may continue independently while the human moves between conversations. Making both look like tabs does not make their lifetimes interchangeable.

Chapter 19 supplied delegation. This chapter adds named top-level Agents sharing one Ensemble, with independent workspaces, conversations, provider settings, and runtime state. The browser can select which Agent receives a message without making that selection the ownership structure.

The implementation also corrects the title's most tempting implication. It does not introduce a global society in which every model discovers and commands every other model. Ordinary model-facing fleet operations remain scoped to the caller's direct children. The human and embedding application can create and select peers through a different boundary.

## Names identify relationships

Agent names are immutable, with hierarchical identity derived from the actual parent path. A delegated child's name is local to its parent. This is not a mutable display label attached to a separately invented permanent UUID. The implementation keeps direct object relationships and resolves control targets through their owners.

The initial host Agent is host-configured and has the same sibling-home placement as other top-level Agents. It is not duplicated in the persistent peer registry. Other top-level peers have distinct names and workspaces; duplicate names and aliases of an already owned workspace are refused. A friendly label must not become permission for two Agents to write the same save and event files.

Creation is transactional. Configuration and persistence are prepared, current startup skills become ready, and only then does the root collection publish the new peer. Failure unwinds the provisional Agent and its registrations rather than leaving a selectable object that never finished construction.

Creation takes an explicit current source Agent. Missing model and skill overrides inherit from that source. The new peer receives its immutable runtime configuration snapshot and parsed skill templates, then constructs independent mutable Skills and MCP relationships. Ensemble keeps no defaults-Agent pointer. Sharing a template does not share a live grant, connection, or conversation.

## Persist launch intent, not the process

The root's `ensemble.json` records immutable launch choices: each peer's name, model, primary skill, and creation time. Individual conversation saves live in their corresponding workspaces. These records have different jobs: one says how to reconstruct a peer, the other says what it remembers. Creating another peer must not overwrite an older peer's launch model with a later interactive model choice.

Named peers are reconstructed at launch. Current startup configuration determines their runtime authority; saved history does not resurrect credentials, old tool handlers, or unfinished processes. Restoring the root collection likewise does not respawn a previous process's descendant tree. The new process starts with real new objects and explicit readiness.

Writes preserve this distinction. Saving a conversation should not require rediscovering the roster by scanning arbitrary directories. Updating launch records should not rewrite every peer's transcript. Workspace paths and identities are validated before registration, including collisions that would otherwise appear only after restoration.

Deletion removes the live peer and its launch record, after the application confirms the required work boundary. It retains the workspace, conversation save, and event files. Deleting a selectable peer is not permission to erase its history from disk.

The deletion operation refuses while unfinished descendant work still depends on the peer. It is not a shortcut around the fleet's bottom-up completion rules. Once deletion is permitted, teardown is joined and the registry is replaced. If replacement fails, the stopped registered peer remains available for a retry. A browser selection change cannot substitute for cleanup or repair a half-completed deletion.

## Give the browser a tree, not another registry

The Agent tree represents Ensemble's real relationships. Its snapshot exposes canonical addresses derived from parent paths, names, and current status. An unnamed or detached Agent has no such address. An address is a routing description, not a permanent identity: deleting and recreating the same named path can produce a different object. Already admitted work retains the actual object it concerns. Live updates let a viewer observe work elsewhere without switching conversations. The tree is a projection, not a second place that owns Agents.

Selecting another Agent changes the target for input, hints, pause, interrupt, settings, and transcript viewing. It does not pause the old selection or cancel its background jobs. Each control carries its actual target; a delayed click must not act on whichever Agent happens to be selected when the server receives it.

The browser bridge has a separate connection boundary. Its existing MCP relay stays with the Agent that established it; selecting another conversation does not transfer or renegotiate that connection. Peer selection and browser capability ownership must not be conflated.

Recorded transcript state and current runtime observation remain different. Reconnection rebuilds the selected view from authoritative state and resumes updates; it does not infer completion from the absence of recent chunks. A quiet peer may be waiting for a provider, and a peer with visible final text may still have obligations to children.

Identity is also essential for delayed observations. A result from an old selection still belongs to that Agent. Reusing one mutable “current Agent” pointer when displaying callbacks would make the interface fast and wrong.

## Delegation keeps its existing boundary

The fleet tools still act on children owned by their caller. Spawning creates that parent-child relationship. Waiting, messaging, inspecting status, and shutting down operate on the corresponding child handles. Knowing a top-level peer's ID does not grant a model the right to control it.

Child questions remain request/response operations. The parent can answer the exact question while ordinary child output continues independently. A running observation is not submission, and a child that has produced a reply is not necessarily finished with its task. Submission remains a validated, durable event rather than a guess based on conversational tone.

The observation tools answer different questions. Sending a message steers active work or admits a turn when idle. Asking a child instead admits a separate turn and waits on that exact reliable reply; an old assistant final cannot satisfy it. Waiting for finalized text or an idle status is not waiting for typed submission. Reading a saved submission is not the same operation as joining current selected-child completion.

A change wait captures the current child objects and their observation baselines at admission. It notices status, committed output, questions, and transient streaming progress without taking messages away from another observer. A later sibling does not silently join the wait, and a replacement with the same local name does not become the old captured object. Lack of observed progress is a useful supervisory hint, not proof of deadlock. An observation timeout leaves the work alive.

The parent and child lifecycles also remain visible. Parent shutdown must cancel and join owned work before destroying resources it uses. A display card disappearing does not establish that the work stopped. Conversely, retaining a saved child directory does not keep the corresponding live Agent alive.

This boundary is intentionally narrower than arbitrary peer-to-peer messaging. A global broker would need its own authority, discovery, and lifetime rules. The current application has no reason to acquire those mechanisms merely because multiple Agents appear in one window.

## Why immutable configuration matters here

The pressure to share mutable settings increases as soon as several Agents exist. A host already has a convenient configuration object; passing it to each new Agent seems cheaper than resolving a snapshot.

That shortcut makes later behavior difficult to explain. A model change in one pane might alter another Agent's next request. A parent's updated policy might silently affect a child midway through work. Restoration might mix the old peer's saved intent with the currently selected peer's provider route.

Resolved snapshots remove that ambiguity. Deliberate settings changes still go through the owning Agent and apply at their documented boundary. They do not travel accidentally through an alias to somebody else's settings.

The same discipline applies to observers. The browser can receive a merged view, but each underlying event still identifies its owner. Aggregation is useful precisely because it does not erase the facts that make the individual streams intelligible.

## Parallel work is only useful if its cost is legible

A second Agent can investigate a different failure, review a draft, or work in a separate workspace while the first continues. It can also repeat the first Agent's work, wait forever for an answer already sent elsewhere, or spend tokens solving an unassigned problem.

The application cannot make delegation useful by multiplying model calls. It can make the assignment, ownership, status, and completion visible enough for a human or parent Agent to manage them. Explicit submission, questions, cost observations, and retained traces are the practical control surface.

This chapter therefore adds a modest thing with substantial consequences: several independent roots inside one usable application, and an honest view of the children they own. The next problem is temporal rather than spatial. Even a correctly identified Agent eventually runs out of room to remember why its work matters.

[Implementation chapter](../chapter-20.md)
