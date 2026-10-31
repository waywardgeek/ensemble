# Chapter 20: The Ensemble

I have been steering a single agent through this book. I listen
to its reasoning at five times speaking rate and redirect between tool
calls. Now give that model three sub-agents. It can decide that
delegating was the job. Three children finish. The parent is not listening.

There is another kind of isolation that nobody asked for. After the
first edition's architecture repair, I wanted the Artifact scroll for
Homebrew-VTT. The component was reusable. Its Go package lived behind
`internal/`, where another application could not import it. The chapter
about architectural decay had missed a boundary nobody had tried to cross.

This chapter connects both ends: an interface another application can use,
and agents that can delegate without abandoning the conversation.

---

## TL;DR

The framework must not know the GUI exists. The GUI must still be useful
outside the command that originally shipped it. A fleet adds another
construction boundary: the library manages children, while the application
supplies the factory that creates them. A passing library fixture does not
establish that either boundary is connected in the actual executable.

### A reusable GUI

Keep the existing optional GUI module and public core API. A headless consumer's
`go list -deps` must name neither the GUI nor `gorilla/websocket`, including
through MCP. An outside GUI consumer must construct real public values, replay
an event and deliver a prompt through the Agent. Preserve the wire protocol
and verify that a requested model switch changes actual subsequent provider
requests.

Embed the browser assets so a built binary serves its interface from an
unrelated working directory. Provide an explicit disk override for development.
Expose a read-only asset filesystem so another application can mount
`artifact-scroll.js` and `renderers.js` without serving `index.html` or starting
the chat shell. The public browser MCP relay accepts a listener supplied by the
application and returns each JSON-RPC response with its caller's request ID.
No browser means a correlated error. Closing the client, browser or server must
release the corresponding pending work.

The accepted API uses `gui.New(root, agent, log)`, `gui.Assets()`,
`server.Handler(directory)` and `server.ServeMCP(listener)`. Its root and Agent
already exist; no callback bundle or new ownership layer is needed.

### A terminal fleet

Build the fleet in four stages:

1. **A manager for each Agent.** It spawns, tracks and shuts down children.
   A child is a full Agent with its own actor, tool visibility, save file and
   data directory under its parent's directory. The application installs an
   `AgentFactory` on Ensemble; descendants reach it through their actual owner
   chain. Derive each child specification from its parent and enforce Chapter
   21's narrow-only constraints. Recursive shutdown preserves the saved work.
   These separate state directories do not give children private copies of the
   task workspace. Optional `submit_schema` describes this child's whole submit
   argument object. Validate and defensively copy it before construction, preserve
   it through the factory, and return independent schema/result bytes. Ordinary
   descendants reset to the default unless their own launch supplies a schema.

   Each Agent also has its own ID and its parent's routing ID. The user is an
   address, configurable at the application boundary, but never a managed child.
   Addressing a person does not grant permission to spawn or shut that person down.

2. **Seven workflow tools.** Add `spawn_sub_agent`, `join_agents`,
   `agents_status`, `shutdown_agent`, `send_message_to_parent`,
   `respond_to_parent_message` and `submit`. Fan-in waits for every selected
   child's submission, but yields a pending question so the parent can answer
   it. A reply carries its sender. An omitted message ID answers the sole
   matching pending question; ambiguity is an error.

   Asking and finishing are different verbs. `submit` is the explicit completion
   signal; prose saying “done” is not one. The framework observes submission.
   The parent still judges whether the submitted work is correct. Default submit
   accepts `{"result":"text"}`. A custom submit takes its direct JSON object, with
   no wrapper or JSON string encoding. Runtime validation precedes persistence
   and completion; invalid arguments return a tool error and leave it unsubmitted.
   Preserve nested values through joins, status and durable storage. String
   convenience calls use the same validator; they cannot bypass a custom contract.

3. **Interruptible waits.** An incoming message wakes an Agent waiting on a
   tool while the job keeps running. Preserve its handle and eventual result.
   Message wake does not depend on the callback timer. Existing `tool_limits`
   supplies a one-shot override for the next tool call; it is not a new
   per-tool-name configuration map. Do not leave a sticky interrupt for a future
   waiter when no waiter is present.

4. **Bounded delegated correction.** An ordinary delegated Agent whose successful
   turn ends without `submit` receives exactly three consecutive framework wakes
   before failure is reported to its parent. Reset the count on submission. Active
   or unresolved failed children prevent early submit; finish or shut them down
   first. Failed turns do not enter this retry loop. Main and named peers have no
   Agent parent and may finish normally in chat, even with a factory installed.
   Explicit root submit remains available under the same admission rules.

The terminal installs the factory and human sink. A non-fleet embedder keeps its
ordinary turn behavior. Submission can wake a join before the child's tool-result
tail and reliable completion/save finish; wait for that completion and drain before
disposing of it. The GUI remains a single-Agent client until Chapter 21.

### Submission contract

Absent `submit_schema` retains the default object schema with required string
`result`. Extra keys remain accepted and ignored on that legacy path; store the
normalized row `{"result":"text"}`. Explicit custom schemas preserve the complete
object, even when their shape happens to resemble the default. Custom JSONL rows
in the child's `submit.md` contain `version: 1`, `submit_schema` and `payload`.
Unknown versions or malformed rows are errors, not permission to return an older
valid result silently.

Use the following supported subset of JSON Schema Draft 2020-12. The root must
declare `type: "object"`; nested schemas can be objects or booleans. An optional
`$schema` must name `https://json-schema.org/draft/2020-12/schema`.

| Supported keywords | Meaning |
| --- | --- |
| `type`, `enum`, `const` | JSON types (including type arrays), permitted values and exact values |
| `properties`, `required`, `additionalProperties`, `minProperties`, `maxProperties` | Object fields and bounds, recursively validated |
| `items`, `minItems`, `maxItems`, `uniqueItems` | Array item schema, bounds and uniqueness |
| `minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum`, `multipleOf` | Exact numeric constraints |
| `minLength`, `maxLength`, `pattern` | String bounds and Go regular-expression syntax |
| `title`, `description`, `default`, `examples` | Annotations; no default insertion or data coercion |

Refuse all references, combinators, formats/content and other unsupported keywords
at launch; never fetch schema resources. Validate the schema itself, not merely
its keyword names. Annotation and enum objects are data, not nested schemas.
Preserve JSON numbers without rounding through `float64`.

Admission limits are refusal boundaries, never truncation: schema 65,536 bytes
and 2,048 JSON value nodes; payload 1,048,576 bytes and 65,536 nodes; maximum
nesting depth 64 with root at zero; number literals at most 128 bytes and exponent
magnitude at most 1,000. Byte caps apply to incoming and canonical owned JSON.
Also require canonical schema bytes × canonical payload bytes ≤ 67,108,864 before
validation. This conservative combined budget can reject an otherwise individually
admissible pair; it does not promise a CPU deadline. Use a real schema validator
for the admitted semantics rather than treating a declaration as enforcement.

Exact type and field spelling is yours where the public behavior is equivalent.
The seven fleet tool names are the contract; extra fleet tools are unnecessary.
Build and grade from the repository root:

```sh
go run ./cmd/grade-edition2 -ch 20 solutions/ch20
```

The combined rubric retains the seven GUI checks and seven fleet checks, 200
points in all. The exercise and check table at the end cover both halves.

## The physics

Go's import graph is transitive. If A imports B and B imports C, then C is a
build dependency of A. `go list -deps` exposes that reachability. It does not
measure how many machine-code bytes survive linking; unused code can be removed.
The boundary being checked here is what the headless application depends on.

`internal/` is the reverse rule. Code inside an `internal/` directory can only
be imported from the tree rooted at its parent. An outside application cannot
work around a missing public type by spelling its internal import path.
[Go's package rules](https://pkg.go.dev/cmd/go#hdr-Internal_Directories) enforce
the boundary at build time.

The first-edition GUI had both problems. Outside applications could not import
it, while the framework root imported it for everyone: batch tools, headless
clients and model-loop tests. The dashboard was welded to the engine. Moving
the GUI out of `internal/` and removing the core's import address different
problems; doing only one leaves the other intact.

## 20.1 Two paths to the same dependency

A refactor begins with a census. Grep before moving.

The original server package held client lifecycle, replay, inbound routing,
browser-RPC correlation, MCP tunneling and logging. Its JavaScript included a
self-contained Artifact scroll. That scroll was the part Homebrew-VTT wanted,
but reaching it meant reaching the package that owned the rest.

The historical graph made the first repair look easier than it was:

```text
  agent (original root)
   ├── internal/common
   ├── internal/llm
   ├── internal/mcp ─── gorilla/websocket  ← path 2
   ├── internal/ws ──── gorilla/websocket  ← path 1
   └── ...
```

The root re-exported four WebSocket symbols: `WSHub`, `NewWSHub`, `ServeHTTP`
and `NewMCPWSTransport`. The original census found no consumers. The command
called the internal constructor directly. The published API could sit unused
while the product worked, because no external application tried to use it.

Delete those exports and remove the obvious GUI import, and Gorilla still
appears in `go list -deps`. The MCP client's WebSocket dial transport brings
it in through the second door. Only the browser client needs that transport;
headless MCP still needs its protocol and other transports.

The lesson is older than Go modules. Run the transitive dependency check before
the move to know where you stand, and after it to know whether you moved.
Cutting the import you know about and calling it done is the dependency-graph
equivalent of fixing the bug you can see and skipping the regression test.
Gorilla remains useful where the browser actually needs it.

The second edition already keeps both WebSocket paths in the optional GUI
module. This exercise demonstrates that boundary through outside consumers.
It does not ask you to weld the dashboard back on so you can remove it again.

## 20.2 Construction without a bag of callbacks

The first-edition Hub had transport work and an Agent control surface mixed
in one constructor. Six arguments supplied a pause gate, send callback, log,
settings and usage. A seventh dependency arrived afterward: a closure returning
the current model.

The reason for reading the current model was real. An operator can switch models
mid-session; a string copied at construction would describe the old model while
looking entirely plausible. But a field assigned after construction could be
forgotten by another caller. Moving that closure into a keyed `AgentHooks`
literal did not make all its fields mandatory: omitted Go struct fields receive
zero values. The [language specification](https://go.dev/ref/spec#Composite_literals)
is explicit about that distinction.

The existing owner chain removes the need for the bundle. The optional server
uses the supplied public Agent to reach its configuration, Engine, history and
pause state. Ensemble supplies observations. The constructor checks the supplied
root, Agent and logger, including that the Agent belongs to that root:

```go
func New(root *ensemble.Ensemble, agent ensemble.Agent, log io.Writer) (*Server, error)
```

This is the GUI package's public signature. The server owns browser connections
and frontend settings; it does not become the owner of the Agent merely because
it displays one. The application controls both lifetimes. Public event values
let the application construct a history entry, and the same Agent handles a
prompt whether it came from a terminal or a WebSocket.

Bill saw the ownership distinction while the original GUI was being extracted.
The frontend should remain useful to Homebrew-VTT; the object managing Agents
should be able to serve more than one frontend. That object is already Ensemble
here. A reusable view and a fleet need different responsibilities, even when
the same application eventually presents both on one screen.

## 20.3 The binary carries its interface

The original GUI served `web/gui/` relative to the source tree. Start the binary
from another directory and the interface disappears into HTTP 404s. A harness
that always sets the working directory to the repository hides the deployment
problem. The binary carries its model loop and tools but cannot find its own
interface.

`go:embed` makes the files part of the binary. The current GUI keeps its embed
anchor private and exports an accessor rooted at the browser assets:

```go
//go:embed web/gui
var assets embed.FS

func Assets() fs.FS {
    files, err := fs.Sub(assets, "web/gui")
    if err != nil {
        panic(err)
    }
    return files
}
```

`server.Handler("")` serves that filesystem. A nonempty directory selects disk
assets instead; the standalone GUI command exposes it as `--gui-dir`. Edit a
CSS file, reload, and see the change without rebuilding Go. An absolute override
path works independently of the launch directory. A relative override remains
relative, exactly as its spelling suggests.

The embed directive requires a package-level variable. Its files are read-only,
but an exported Go variable could still be assigned another filesystem value.
Keeping the anchor private and returning `fs.FS` exposes the resource without a
public mutable binding. [The embed documentation](https://pkg.go.dev/embed)
describes the read-only value and package-scope directive; this is not a general
exception for mutable application globals.

A deployment test must inspect what came back. In the student run, an initial
asset check accepted HTTP 200 from a directory listing that contained no GUI.
Requiring the application shell caught the same defect. A page status answers
whether the server returned a page. It does not identify the page.

## 20.4 Reuse for real

The headless proof shows what the framework excludes. A reuse test shows what
the GUI includes. Build both consumers outside the source module, so Go cannot
accidentally grant them access to internal types. The accepted layout uses two
module replacements: `ensemble` for core and `ensemble-gui-stub` for the GUI.
The latter is the optional module's historical name, not an indication that its
browser interface is still a stub.

The outside GUI consumer constructs Ensemble and an Agent, appends a public
`Event` containing `MessageData` and `Part` values, then constructs the server.
A WebSocket subscription must replay that event. A prompt must reach the real
Agent and appear in the provider request. An HTTP provider fake can make that
last exchange deterministic while leaving the actual runtime in the path.
A callback that merely records a string would miss the integration being tested.

You cannot construct what you cannot name. This was the second gap in the
original reuse attempt: the server was public, but its event vocabulary lived
under `internal/`. Public aliases now expose the types consumers actually use.
An in-tree test cannot establish that this works outside the import boundary.

The Artifact scroll is a smaller consumer still. Mount only
`artifact-scroll.js` and `renderers.js` from `gui.Assets()`. Instantiate
`ArtifactScroll` with a supplied DOM element and feed it display messages through
`receive`. It needs neither `app.js` nor application controls nor a WebSocket.
Its optional speech dependency can stay absent. Homebrew-VTT gets a scroll in
its own page instead of another copy of the Ensemble application.

The browser MCP relay is public for the same reason. `ServeMCP` accepts a
`net.Listener` chosen by the embedding application. It relays JSONL requests
through a connected browser and restores the original caller's JSON-RPC ID in
the response. Two clients may reuse an ID while browser replies arrive in reverse
order; the relay must still deliver each response to the right client.

Without a browser, the error is immediate and correlated. With a connected but
silent browser, the call remains pending until a relevant connection or the
server closes. A no-browser error test does not establish that latter lifetime.
Keep reading the relay socket while requests wait so a disconnected caller can
release its own work. Closing the server must close its listeners and join its
connection workers; a timed-out test is not permission to leave them behind.

## 20.5 Two kinds of silence

A single Agent operates in a loop: reason, call a tool, receive the result,
reason again. The human can send hints while that work proceeds. Delegation
makes two existing risks particularly expensive.

The parent calls `join_agents` and waits for its children. If the terminal's
input reader is blocked on that same turn, a prompt typed by the human cannot
reach the mechanism meant to wake it. The job supervisor can be correct while
the product still takes the wheel away.

The second failure is quieter. A model spawns three children and calls no more
tools. Its turn ended normally. The status says idle, just as it would after
finishing a task correctly. The model decided that delegating was the job, and
delegating is complete. Three children can finish and submit while the parent
is no longer attending to them.

These failures can combine. A delegated parent can receive a framework wake after
stopping early. The parent remembers to join, enters a blocking wait, and the disconnected input
path prevents the human from steering it. Recovering attention does not ensure
that the model uses it well. The fleet needs child ownership, message delivery
through waits, and a bounded response when a delegated model stops without submitting.
The human-facing root still has to attend to its children and verify their work;
its ordinary chat completion is not proof that delegation succeeded.

## 20.6 Spawning

Give each agent a `SubAgentManager`. The manager holds the children
map and exposes three operations: spawn, shut down, and list. A child
is a full agent with its own actor, its own tool visibility, its own
save file, and its own data directory, created under the parent's
directory so the filesystem mirrors the ownership tree. A child whose
task fails can be shut down and respawned without replacing the parent's
conversation. These Agents share a process; this is not process-crash isolation.

The manager does not know how to build an agent. That job belongs to an
`AgentFactory`:

```go
type AgentFactory func(parent Agent, spec AgentSpec) (Agent, error)
```

The composition root supplies the factory at startup. It receives the actual parent and
child specification and returns a constructed Agent. The manager loads its
save, registers it and starts the task. The application selects the child
configuration through that factory, retaining the existing provider and
credential-resolution path instead of inventing a second one inside the manager. The library says what it needs. The binary decides how to
provide it. The split is the same one that `http.Handler` draws between
"handle this request" and "listen on this port."

The factory seam exists because the same library runs in different
contexts. The full Ensemble binary supplies a factory that creates
agents with real models, real API keys, and real file systems. A test
harness can use the same real Agent against a scripted external provider
and temporary directories. A headless CLI might supply a factory with
no GUI hooks at all. The library code that calls `Spawn` does not
change. What changes is the factory registered at startup, and the
factory is the only code that knows which kind of agent this binary is
meant to run.

The terminal installs the factory on its existing root. This is the actual
construction seam, with configuration inherited from the owning parent:

```go
app.SetAgentFactory(func(parent ensemble.Agent, child ensemble.AgentSpec) (ensemble.Agent, error) {
    config := parent.Config()
    if child.Model != "" {
        config.Model = child.Model
    }
    return app.NewChildAgent(parent, config, child)
})
```

The manager derives IDs and paths from the parent; `NewChildAgent` checks the
real owner relationship and clamps the specification. A sandboxed parent cannot
spawn an unsandboxed child, and a parent without web access cannot grant it.
Credentials remain configuration owned by the application and Agent; they are
not routing fields copied into a model-visible specification.

The shared root makes the factory available at every depth. A child reaches
Ensemble through its actual parent relationships, rather than receiving a second
copy of a callback whose propagation can be forgotten.

The terminal installs this application-supplied factory before constructing
its Agents. An embedding application that leaves the factory unset keeps its
ordinary tool surface. Installing a factory does not turn the human-facing root
into a delegated worker that must finish with `submit`: completion policy follows
the actual Agent parent, rather than the availability of child construction.

A launch can also supply `submit_schema`, an immutable contract for that child's
answer. Validate and copy it before construction effects, and require the factory's
returned child to retain it. Otherwise a caller could change the live contract by
editing a byte slice after launch, or a factory could quietly drop the promised
validation. Ordinary descendants start with the default text contract unless
their own launch explicitly supplies a schema; the parent's answer format is
not automatically appropriate for a different task.

Shutting down a parent cascades to its children, recursively. The
fleet's lifetime is bounded by its root. A child that outlives its
parent is a resource leak with a model attached, and the cascade
prevents it.

The proof of isolation is on disk. Spawn a child, and a directory
appears under the parent's data tree. Inside it: a save file that
belongs to the child alone, with its own event log and its own tool
state. Shut down the child, and the directory survives, because the work
inside it might matter to whoever reads the results. Restart the same
child with the same id, and its saved conversation is loaded before the new
task starts. The parent's conversation and save are separate. The task workspace is
still shared: a child's file edits can affect the parent's work. State isolation
is not a promise that children cannot damage shared files; Chapter 19 supplies
the optional confinement boundary.

## 20.7 The parent and the user

Every agent knows two things about its place in the fleet: its own id
and its parent's.

```go
type AgentSpec struct {
    ID       string // e.g. "researcher-1"
    ParentID string // parent's ID, or UserID for root agents
    // ...
}
```

The root agent defaults to `"main"`. A child's parent id is set by the
spawn call. And the user gets a reserved address, so the same messaging tool can address an Agent parent or the human sink.
The routing ID does not replace the runtime back-pointer: a root Agent still
belongs to Ensemble, and a child retains its actual Agent parent.

The user is addressable but never a member of the fleet. No agent can
spawn the user, shut down the user, or query the user's status. This is
a design boundary. Addressing is uniform:
`send_message_to_parent` works the same way whether the parent is
another agent or the human at the terminal. Lifecycle is not: the
operations that manage agents do not apply to a person sitting in a
chair. Putting the user in the children map would force an "unless it
is the user" guard into every loop that walks the fleet, and that guard
would exist because the abstraction promised something it could not
deliver.

The user's address is configurable. The default is the role, but the
binary can set it to a name instead. A reply addressed to a name is
a reply the model can follow up without consulting an address book.

A reply is a full message with a sender address, not a bare string.
The reason is practical: the first time a child receives an answer and
needs to ask a follow-up, an answer with no sender cannot be continued.
The child does not know where the answer came from. Every reply carries
the sender's address, so the conversation can continue without the
child maintaining its own routing table.

A question also needs an identity in the place that will answer it. During the
student implementation, a parent and child with the same local name generated
the same first question ID. The child question overwrote a pending question to
the human. Making IDs local to each sender had not made them unique at the
recipient. The recipient's manager now allocates question IDs under the lock
that already protects its pending questions. That is the scope in which replies
must be unambiguous.


## 20.8 Asking is not finishing

The first edition initially gave the parent link only
at submit time. A child could do its work and deliver a result, and the
result would reach the parent. Bill caught the problem immediately: a
child that needs to ask a question before it finishes would have to
call `submit` to talk. Asking and finishing would be the same verb, and
a child cannot ask for help without lying about its state.

The fix separates the two. `send_message_to_parent` is the general
verb: a child calls it at any point during its work, the child blocks,
and the parent receives the message. `submit` is the narrow verb: the
child is finished, here is the result. The framework observes that a
child called `submit` with a conforming payload. Whether the work that
payload describes is correct is a judgment the parent must make, not a
fact the framework can verify.

Seven tools form the fleet API:

- `spawn_sub_agent` creates a child from a specification.
- `join_agents` blocks until every listed child has submitted, returning
  each child's result.
- `agents_status` lists live children and their current state.
- `shutdown_agent` stops a child and removes it from the fleet.
- `send_message_to_parent` blocks the child until the parent replies.
- `respond_to_parent_message` delivers the parent's reply and unblocks
  the child.
- `submit` records the Agent's completed result.

These seven names are the exercise. Additional fleet tools are outside this
chapter. An omitted reply ID resolves the sole matching pending question; if
there are several, require the ID rather than guessing which child to resume.

`join_agents` is the fan-in, and it has one behavior beyond blocking
that the rest of the fleet depends on: it returns early when a child
sends a message. Consider the deadlock it prevents. A child calls
`send_message_to_parent` and blocks, waiting for an answer. The parent
calls `join_agents` and blocks, waiting for the child's `submit`.
Neither yields. Both wait forever. The early return breaks the cycle:
the parent wakes when the child's message arrives, answers the
question, the child unblocks and resumes work, and the parent goes back
to waiting for `submit`, which now arrives because the child can
proceed.

The tool arguments are small:

| Tool | Arguments |
| --- | --- |
| `spawn_sub_agent` | `agent_id`, `task`, optional `model` and `submit_schema` |
| `join_agents` | optional `agent_ids`; omitted means all current children |
| `agents_status` | none |
| `shutdown_agent` | `agent_id` |
| `send_message_to_parent` | `content` |
| `respond_to_parent_message` | `agent_id`, `response`, optional `message_id` |
| `submit` | `result` by default; the complete custom argument object when launched with `submit_schema` |

A parent reviewing files may need findings it can inspect individually instead
of a paragraph it has to parse. The launch can make that expectation explicit:

```json
{
  "agent_id": "reviewer",
  "task": "Review the change; submit findings.",
  "submit_schema": {
    "type": "object",
    "properties": {
      "summary": {"type": "string"},
      "findings": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {"path": {"type": "string"}, "fixed": {"type": "boolean"}},
          "required": ["path", "fixed"],
          "additionalProperties": false
        }
      }
    },
    "required": ["summary", "findings"],
    "additionalProperties": false
  }
}
```

The child then calls `submit` with the direct object:

```json
{"summary":"Checked the change.","findings":[{"path":"main.go","fixed":false}]}
```

There is no outer `result` field and no quoted JSON document inside a string.
The same compiled contract supplies the declaration and validates the actual
arguments before any durable submission or completion notification. Omitting
`fixed`, supplying `"false"` as a string, or adding a forbidden nested field
returns a tool error and leaves the child unsubmitted. It can correct the
arguments within its existing bounded lifecycle. `false` is a valid boolean;
whether the file is actually fixed still requires inspecting the work.

The supported contract is an explicit subset of JSON Schema Draft 2020-12.
Nested object and array constraints are enforced; unsupported keywords and all
references are refused at launch. A declaration that accepts a keyword while
ignoring its meaning would promise a check that does not exist. Schema and
payload admission limits refuse oversized or excessively nested data instead of
truncating it. The TL;DR specifies that public boundary. No schema resolution
fetches a URL or reads a local file.

Without a custom schema, `submit({"result":"text"})` keeps its existing meaning.
The public string convenience method encodes that same default shape and enters
the same validator, so it cannot bypass a custom contract. Structured joins,
status and retrieval preserve the complete JSON object. Text-valued displays can
format it for a reader without changing its stored type.

Questions and fan-in use the ordinary supervised tool path. If the initial
callback returns a running handle, `wait_for_job` retrieves the later reply or
completed fan-in. Ending one wait does not discard the conversation it was
waiting for. At the terminal, `/reply TEXT` answers the sole human question;
`/reply MESSAGE_ID TEXT` selects one when several are pending.


## 20.9 Waking the blocked Agent

The job supervisor already separates a tool's execution from the Agent's wait
for it. When the callback delay expires, the Agent gets a running notice and a
handle. The process or handler continues. Incoming messages use that same
separation: wake the waiter, deliver the message and leave the job alive.

Message arrival is independent of the timer. The earlier `tool_limits` tool
sets a one-shot limit for the next call, including a third-party tool without
callback fields of its own. It does not make messages deliverable by installing
a per-name timeout. A long callback delay must not turn into a long delay before
a human can be heard.

An interrupt aimed at a job with no current waiter is a no-op. Otherwise a
later wait could consume a stale signal and appear to have been interrupted by
a message that arrived minutes earlier. Preserve the eventual result under the
same handle so the parent can answer the hint and return to its original join.

The input reader matters as much as the supervisor. Keep reading stdin while
the Agent works. A human message typed during a turn becomes a hint; a `/reply`
command answers a pending fleet question without being misclassified as a new
model task. The actor remains the single place that orders conversation changes.

## 20.10 Three wakes, then tell someone

“Done” has a precise meaning for a delegated Agent: it called `submit`, and its
children have finished. A normal model response ending is not sufficient for
that worker. The human-facing main Agent and named peers finish normally in chat;
they have no Agent parent awaiting a submission. The framework can enforce the
delegated protocol without deciding whether the result is good.

First, `submit` rejects while any child is still active. The error tells the
parent which children need attention. Submissions happen bottom-up. A failed
child is unresolved too; shutting it down removes it from the managed set while
preserving its files. Failure does not quietly become a successful child result.

Second, a delegated Agent's successful turn that ends without submission receives
an immediate framework correction. With active children, the correction asks the Agent to
join or shut them down. Without active children, it asks for the missing submit.
No polling timer is needed: the runtime is already handling turn completion.

The correction lives in the library's real completion path and checks the actual
Agent parent. A factory-enabled root remains a chat agent; an ordinary delegated
child owes its parent a result. This keeps construction, public tool availability
and completion policy from becoming three accidental meanings of “has a factory.”

Count consecutive framework wakes per Agent and reset the count when it submits.
After exactly three wakes, stop re-prompting and escalate to the parent. That
parent can explain the failure to the human or resolve the child before continuing.
The bound limits attempted correction; it cannot force a model to cooperate or
guarantee a useful result.

That distinction produced a real student correction. An exhausted child first
remained labelled running. The parent could then wait for work the framework
had stopped trying to complete. The corrected state is failed, with an error
and no fabricated submission. Fan-in can report that failure instead of waiting
forever or presenting an empty result as success.

A failed turn takes a different path. An HTTP rejection or other execution error
must be reported without entering the three-wake loop. Repeating a failed
exchange merely spends the correction budget on a different problem. A model
refusal delivered as a successful response can still be an ordinary unsubmitted
turn; use the actual runtime outcome rather than guessing from its prose.

Framework wakes and escalation notices are recorded in the Agent's event log.
The terminal's human sink exposes questions the root sends to the person. These make
the protocol inspectable: which Agent stopped, how many wakes it received, and
whether it submitted or exhausted the bound.

Submission also precedes quiescence. Publishing the accepted result can wake a
join while the child's submit tool still has a result to append and its actor
still has completion and save work to finish. A caller that needs to stop or
dispose of that worker must wait for its reliable completion and drain, rather
than interrupt it from inside the submit callback.

## 20.11 Five wiring failures
The first edition records five failures when its real executable finally ran
the fleet. Each had been hidden by a fixture supplying what the product omitted.
They are construction lessons, not bugs to recreate in this implementation.

**No agent factory in the composition root.** The binary returned
"no agent factory configured" on the first spawn attempt. Every fleet
test in the library passes, because every fleet test constructs its own
factory during setup. The application startup path never registered one. The library knew
how to use a factory. The product never provided one.

**No user sink in the composition root.** The user was addressable in
principle and unreachable in practice. A top-level agent's parent was
the user, and `send_message_to_parent` dispatched correctly to the
user's address. But nobody had wired a sink to receive the message.
Escalation from a top-level agent had nowhere to arrive. This is the
one place a human most needs to hear from the fleet, because the fleet
has run out of agents to escalate to.

**The factory never propagated to children.** A child's manager held a
nil factory. The parent could spawn, but the child could not spawn its
own children. The fleet was permanently one level deep. The grader
found it because its messaging probe runs three levels deep: a
grandchild sends a message to its parent, which relays to its own
parent. A two-level probe need not exercise this: it can create a child without
asking that child to create the next generation.
"Arbitrarily deep" was the design. One level was the product.

**`respond_to_parent_message` required a `message_id` the model had to
copy from prose.** Every parent message includes the id in its payload,
so the information is available, but requiring the model to extract an
opaque string from natural language and paste it into a tool call is a
reliability tax on every reply. The fix makes the id optional when exactly one matching question is pending.
Multiple questions still require an unambiguous selection.

**The command loop read stdin on the goroutine that waited for the
turn.** Stage 3 built a mechanism to wake a blocked agent when a
message arrives. The terminal could not use it. The goroutine that
reads user input was the same goroutine that blocked on `waitForTool`,
so a human typing a command while a tool ran could not deliver it. The
mechanism existed. The wire from the terminal to the mechanism did not.
A comment in the command loop claimed that slash commands could steer
"a turn that is already running." The comment had been there since the
file was written. It had never been true. The structure forbade what
the comment promised, and the comment sat beside the input path it misdescribed.

The first-edition tests supplied these dependencies themselves. That let the
library paths work while the application forgot to assemble them. A library
fixture can test factory use, message delivery and idle correction; a built-CLI
exercise checks whether the executable actually installs and exposes them.
Both are useful. Neither substitutes for the other.

The same lesson connects the GUI half of this chapter to the fleet half. A
public constructor with no outside caller and a working factory with no product
registration both describe capabilities the user cannot reach. Exercise the
boundary from the side that depends on it.


## 20.12 What the sweep found

Two more defects surfaced in the full cross-chapter sweep, and neither
belonged to the five.

**A package counter cost chapter 5 ten points.** Stage 3 added
`var messageSeq atomic.Uint64` at package scope to generate unique
message ids without a lock. The number worked. The scope was wrong.
Chapter 5's grader checks for package-level variables in non-test
files, to keep mutable application state with its owner.
The fleet scenarios in that suite did not reject the counter's scope.
Source review or a focused ownership check could have found it; the historical
cross-chapter sweep happened to do so. The fix moved the counter to a field on the messaging state
struct, behind the lock that already protected it. The lesson is worth stating plainly: if the id is per-fleet, the counter
must live on the fleet, not on the package.

**A bisect indicted innocent code.** Original Chapter 19 scored 85 against the
live tree and 100 against four consecutive worktrees. The bisect framed
a one-line change from this chapter as the culprit. That change could
not touch event logs. The contradiction was the clue.

Every worktree path is absolute. The sweep runs graders with relative
paths. The original Chapter 19 harness set the working directory to the student
tree and passed a relative path to the event log file. In the live
tree, where the working directory matched the student directory, the
relative path resolved correctly. In a worktree, where the student
directory lived at a different absolute location, the relative path
pointed somewhere else. Same code, same grader, same tree: 100 with
an absolute path, 85 with a relative one.

How a test is invoked is part of the test.

## 20.13 The exercise

Extend the preceding solution in two connected stages. First demonstrate the
existing public boundary with reusable assets and an outside application. Then
add the terminal fleet to the real command. Keep the GUI's single-Agent wire
behavior intact while the core learns to manage children.

For the GUI stage, build the optional executable and start it away from the
source tree. Verify the actual embedded shell and modules, then select a disk
override and verify its changed contents. Build separate outside consumers for
the headless API and GUI. The first must exclude the GUI and WebSocket packages;
the second must replay public event values, deliver a real prompt and expose the
MCP relay. Mount the two Artifact modules independently and render a message
without the application shell. Check the existing GUI model control by inspecting
requests before and after a turn-boundary switch.

For the fleet stage, implement the seven tools and application factory, retaining
actual parent relationships and narrow-only child specifications. Demonstrate
recursive spawning, distinct state files, pending questions and sender-bearing
replies, fan-in, status and recursive shutdown. While the parent waits, deliver
a terminal hint and later collect the same running job. Exercise successful
delegated submission, early-submit refusal, exactly three corrective wakes and truthful
failure after exhaustion. A failed provider turn must not trigger the correction
loop. The GUI fleet panel is next chapter's exercise.

Also launch the structured reviewer from §20.8. Have a controlled provider first
omit `fixed`, then correct it to boolean `false`; only the corrected object may
complete the child or satisfy a join. Inspect the nested result and its durable
row, not merely a marker in printed JSON. Retain a default text child and an
ordinary grandchild as controls. This checks the contract's rejection and data
preservation without asking a live model to make a particular mistake on demand.

### An outside GUI application

This small consumer uses the public API directly. It has two modes: the full
single-Agent GUI and a page that mounts only the Artifact component. It binds
fresh loopback ports and prints their addresses, avoiding another running
Ensemble. Begin at the repository root:

```sh
ensemble_source=$(pwd)/solutions/ch20
reuse_dir=$(mktemp -d /tmp/ensemble-ch22-reuse.XXXXXX)
cd "$reuse_dir"
go mod init example.org/ensemble-reuse
go mod edit -require=ensemble@v0.0.0 -require=ensemble-gui-stub@v0.0.0
go mod edit -replace="ensemble=$ensemble_source"
go mod edit -replace="ensemble-gui-stub=$ensemble_source/gui"
```

Save this as `main.go`. The component mode makes no model request. The full mode
uses your existing provider environment or Chapter 16 subscription profile;
credentials do not belong in this file.

```go
package main

import (
    "context"
    gui "ensemble-gui-stub"
    "ensemble/authcli"
    "ensemble/ensemble"
    "flag"
    "fmt"
    "io/fs"
    "net"
    "net/http"
    "os"
    "os/signal"
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func run() error {
    components := flag.Bool("components", false, "serve only the component demo")
    directory := flag.String("gui-dir", "", "override embedded full-GUI assets")
    flag.Parse()
    mux := http.NewServeMux()
    if *components {
        for _, name := range []string{"artifact-scroll.js", "renderers.js"} {
            mux.HandleFunc("/"+name, func(w http.ResponseWriter, r *http.Request) {
                data, err := fs.ReadFile(gui.Assets(), name)
                if err != nil {
                    http.Error(w, err.Error(), 500)
                    return
                }
                w.Header().Set("Content-Type", "text/javascript")
                w.Write(data)
            })
        }
        mux.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "text/html")
            fmt.Fprint(w, `<!doctype html><title>Artifact reuse</title>
<div id="scroll"></div><script type="module">
import { ArtifactScroll } from './artifact-scroll.js';
const scroll = new ArtifactScroll(document.querySelector('#scroll'));
scroll.receive({type:'part_final', part_id:'demo', replay:true,
  part:{type:'text', text:'An Artifact without the application shell.'}});
</script>`)
        })
    } else {
        cfg, err := ensemble.ConfigFromEnv()
        if err != nil { return err }
        cfg.DataDir, cfg.BuiltinTools = "cr", true
        root := ensemble.New(os.Stderr)
        if err := authcli.Configure(root); err != nil { return err }
        agent, err := root.NewAgent(cfg, ensemble.AgentSpecFromEnv())
        if err != nil { return err }
        defer agent.Shutdown()
        if err := agent.History().Append(ensemble.Event{
            Type: "message_received",
            Message: &ensemble.MessageData{Actor: ensemble.Human,
                Parts: []ensemble.Part{{Type: "text", Text: "PUBLIC_REPLAY"}}},
        }); err != nil { return err }
        server, err := gui.New(root, agent, os.Stderr)
        if err != nil { return err }
        defer server.Close()
        relay, err := net.Listen("tcp", "127.0.0.1:0")
        if err != nil { return err }
        fmt.Println("MCP", relay.Addr())
        go server.ServeMCP(relay)
        mux.Handle("/", server.Handler(*directory))
    }
    listener, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { return err }
    fmt.Println("HTTP http://" + listener.Addr().String())
    server := &http.Server{Handler: mux}
    defer server.Close()
    stopped, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()
    served := make(chan error, 1)
    go func() { served <- server.Serve(listener) }()
    select {
    case <-stopped.Done(): return nil
    case err := <-served: return err
    }
}
```

Build before launching so the supervised process is the application itself:

```sh
go mod tidy
go build -o /tmp/ensemble-ch22-reuse .
/tmp/ensemble-ch22-reuse --components
```

Open `/demo` at the printed HTTP address. It displays one Artifact; `/index.html`
must return 404. Stop with Control-C, then run the full mode:

```sh
LLM_VENDOR=openai LLM_MODEL=gpt-6-astra \
OPENAI_REASONING_EFFORT=high OPENAI_REASONING_SUMMARY=detailed \
/tmp/ensemble-ch22-reuse
```

The full page should replay `PUBLIC_REPLAY`. Send a short prompt and inspect its
actual request and response. Between completed turns, select another supported
model in Settings and send another prompt. A settings acknowledgment proves the
control was received; the next request's model field proves it took effect.
This small embedding intentionally starts a fresh conversation; the terminal
fleet below supplies the persistent task demonstration.

To check the override, create a directory containing a distinctive `index.html`
and run full mode with `--gui-dir /absolute/path/to/that/directory`. Inspect the
body, then return to the embedded default. To check the relay's no-browser path,
connect to its printed MCP address before opening the page and send a JSONL
request such as:

```json
{"jsonrpc":"2.0","id":"outside-1","method":"tools/list"}
```

The error must retain `outside-1`. Repeat with the real browser connected to
exercise forwarding and a correlated reply. Disconnect a caller while its
request waits, then close the application and verify that no relay connection
or listener survives. Component rendering, a no-browser error and a forwarded
browser response establish different parts of the reuse surface.

Check dependency reachability from this same outside module. A small headless
consumer gives the negative control a real public-library import:

```sh
mkdir headless
cat > headless/main.go <<'GO'
package main
import (
    "ensemble/ensemble"
    "io"
)
func main() { _ = ensemble.New(io.Discard) }
GO
go build -o /tmp/ensemble-ch22-headless ./headless
go list -deps ./headless > headless-deps.txt
go list -deps . > gui-deps.txt
rg 'ensemble-gui-stub|gorilla/websocket' headless-deps.txt
rg 'ensemble-gui-stub|gorilla/websocket' gui-deps.txt
```

The first search should find nothing and return status 1. The second should
name both the GUI module and Gorilla. The positive control matters: a detector
that never matches could make a coupled core look clean. The same comparison
belongs around a headless MCP consumer and its optional WebSocket client; the
protocol alone must not pull in the browser transport.

An in-tree test can hide an internal import, and a library dependency check can
miss a different command's wiring. Inspect the executable being delivered.
`go list -deps` on this outside consumer establishes an import boundary; it does
not establish that all of its functions execute or that a browser rendered the
response. Those are the reasons for replay, prompt, relay and component tests
alongside the graph.

The old supervised probe once finished its report but left the harness waiting
for EOF. Its `go run` wrapper had been killed while a child still held the
output pipe. The first-edition fix built the probe and executed the binary
directly. When a test appears hung after its final assertion, inspect the process
that still owns the pipe before theorizing about model latency.

### A fleet with work to hand back

Build the core from the accepted chapter directory, then create a scratch task
and an explicit primary. This setup gives both root and children the tools they
need; it is not a claim that a tool name in the prompt enables that tool.

```sh
cd "$ensemble_source"
go build -o /tmp/ensemble-ch22 .
fleet_dir=$(mktemp -d /tmp/ensemble-ch22-fleet.XXXXXX)
mkdir -p "$fleet_dir/skills/fleet"
cat > "$fleet_dir/skills/fleet/SKILL.md" <<'SKILL'
---
name: fleet
type: primary
description: A small coding fleet with explicit questions and submissions
tools: read_file write_file edit_file list_directory search_files run_command wait_for_job tool_limits spawn_sub_agent join_agents agents_status shutdown_agent send_message_to_parent respond_to_parent_message submit
---
Use tools to inspect facts and test code. Ask a parent question with
send_message_to_parent and answer child questions with respond_to_parent_message.
Join or shut down unfinished children. Delegated agents finish with submit;
the human-facing main Agent finishes with a normal chat answer. Ordinary text
does not submit a child's result. Keep task files in the shared working directory.
SKILL
cd "$fleet_dir"
LLM_VENDOR=openai LLM_MODEL=gpt-6-astra \
OPENAI_REASONING_EFFORT=high OPENAI_REASONING_SUMMARY=detailed \
EN_SKILLS_DIR="$fleet_dir/skills" EN_PRIMARY_SKILL=fleet \
/tmp/ensemble-ch22 chat
```

Use your already selected billing mode and authenticated profile, or the intended
API-key environment. Keep the terminal open while work runs. Give the parent
this task:

```text
Spawn exactly three children named add, multiply and subtract. Each should write
its named Python module with a two-argument integer function of the same name,
run assertions, then submit its result. Tell add to ask you whether negative
integers are required before implementing; answer yes. Tell each child to run
sleep 4 && python3 followed by its test command, so there is time to steer while
you join. You must collect all three submissions, run combined assertions, and
answer in chat with a concise result containing FLEET_DONE. Do not implement their functions
yourself.
```

While a join is waiting, type:

```text
In your final answer include STEER_COBALT, and confirm that negative
integers were tested.
```

A delivered hint should wake the wait without killing the children. Follow the
running handle back to the eventual fan-in result. After the main Agent has
collected the child results and answered, exit normally with Control-D. Inspect
`cr/children/add/save.json` and the other child saves, and run the combined assertions independently. If a model reports
success without a file or test result, keep that discrepancy in the record.

## 20.14 Taking it for a spin

The second-edition GUI deployment ran from a fresh directory with no source
asset tree nearby. Chrome loaded the embedded shell and its six supporting
assets. GPT-6 Astra at high reasoning wrote `deployment-demo.txt`, read it back
and reported the actual 42-byte content. The run took 9.649 seconds and exited
normally. Component-only rendering and relay lifetime have separate external
consumer checks; that file task alone does not demonstrate them.

The terminal fleet then ran a bounded coding task under the earlier root-submit
policy, with three children, add, multiply and subtract. Each wrote a Python
function and tested it. Add asked
whether negative integers mattered; the parent answered yes. One child's first
shell command was malformed and exited 1. The model corrected the command and
its test passed. This was model recovery within the task, not an automatic retry
of a failed turn by the fleet manager.

The parent joined the children. A terminal hint arrived while join job 12 was
running and woke its wait; the job remained alive. A later wait on that same
handle returned all three submissions. The parent ran combined assertions and
submitted `FLEET_DONE`. Independent execution of the files also passed. The
48.198-second run exited normally after twenty Astra/high requests.

There was a useful failure inside that success. The hint asked for
`STEER_COBALT` in the final submission. The exact hint appeared in a later
provider request, but the final submission omitted the marker. Delivery worked;
compliance did not. A fleet can preserve the human's opportunity to steer without
guaranteeing that the model follows the instruction. Check the input path and
the resulting work separately.

The reproduction above uses the current main-chat/delegated-submit distinction.
The recorded parent's submission and omitted marker remain observations of that
earlier run, rather than a new demonstration under the changed policy.

The child conversations and results remained in their own saves. The terminal
showed the parent's view; it did not become a monitor for every child's reasoning.
The browser still showed one Agent, one stream and one event log. The next
chapter makes the fleet visible through that existing frontend boundary.

## 20.15 Checks

The revised combined grader keeps both original seven-row rubrics. The original
student runs each scored 0/100 before their valid behaviors could be assessed:
the GUI harness assumed different public import and executable paths, and the
fleet harness omitted a required model. Those historical scores are not revised
passes. The current command supplies the actual public launch and library
contracts, with source review for ownership and retained browser evidence for
rendering.

| Behavior | Check | Points |
| --- | --- | --- |
| Headless framework boundary | Outside core consumer excludes GUI/WebSocket dependencies, with a positive GUI control | 20 |
| Public GUI reuse | Outside consumer constructs real public inputs, replays an event, delivers a prompt and exercises a correlated relay reply | 15 |
| Preserved wire behavior | Actual GUI session and provider requests show the requested model switch | 15 |
| Embedded deployment | Built binary serves the real shell from an unrelated directory; explicit disk selection still works | 15 |
| Components without shell | Selected Artifact modules are served while the application shell is absent | 10 |
| Headless MCP boundary | Core MCP consumer excludes WebSocket implementation while the optional client retains it | 15 |
| Public root boundary | Resolved imports and constructible public API preserve the core/GUI split | 10 |
| Spawn and result | Default text and custom-schema children; launch/schema refusals, invalid argument correction and exact nested payload delivery | 15 |
| Child state and factory | Actual construction and distinct paths; structured durable row; default contract on an ordinary grandchild | 15 |
| Fleet visibility and shutdown | Status reports a managed child; shutdown removes it | 10 |
| Fan-in | Selected children all submit; correlated results preserve arrays, booleans and nulls without string encoding | 15 |
| Questions and replies | Three-generation question/reply routing; fan-in yields so the parent can answer | 15 |
| Wake a running wait | Incoming input wakes the wait and its background job remains collectible | 15 |
| Bounded correction | Early submit fails; delegated missing submission receives three wakes; main chat finishes without a wake | 15 |

The structured fixture uses a non-memory reviewer contract. It checks missing,
wrong-type, nested and extra fields, then a valid correction; invalid, unsupported
and referenced schemas must refuse launch. Parse actual correlated tool results
and saved rows rather than searching a frame for a marker. Defensive copies,
factory-contract ownership and Go string-call validation retain separate public-API
regressions; those source tests award no numeric credit to another implementation.

An import check does not inspect the runtime parent chain. HTTP delivery of a
JavaScript module does not show a browser rendering it. Keep the actual factory,
owner, question-identity, human-reply and shutdown checks beside the public-client
and browser checks rather than stretching one score to stand for all of them.

The original fleet audit changed a child's data directory to its parent's and
lost six checks at once. That establishes a serious shared-state failure; it
does not prove six independent protections or that a narrower defect is
impossible. Verify the intended assertion, preserve survivors, and keep a working
control. For the reusable GUI, the analogous trap is a model-switch test that
counts replies while the setting never reaches a provider request.

Package moves also affect applications and harnesses that still spell the old
path. The first edition's speech harness read a renderer from disk and an MCP
check searched an internal server filename. Follow those affected consumers
when changing the public boundary. A clean build in the moved package says
nothing about a caller that can no longer find it.
