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

Keep the core headless and the GUI optional. Another application borrows its
existing Agent, mounts the public GUI, and joins view teardown before destroying
the Agent. Children are actual Agents with both Ensemble and Agent parents,
independent state, explicit submission and joined destruction. Asking, submitting,
finishing a turn and shutting down are four different operations.

The C# edition uses Common declarations and functional partial implementations,
not a manager/factory layer copied from the historical Go solution. Passing
library tests still does not establish that the shipped host has wired the tools,
skills and human input. Exercise those boundaries from outside the library.

## The physics

Dependencies are transitive. A public-looking API is not reusable if its required
types are inaccessible to its caller, and a headless library is not independent
if importing it pulls in the GUI. Go enforces an additional `internal/` import
boundary; C# enforces assembly accessibility and cannot combine partial type
halves across assemblies. Moving a directory is not enough in either language.
Use real consumers to test both a headless dependency graph and the public GUI
construction path. Neither check measures rendered pixels or guarantees that
all request and teardown paths still work.

## 20.1 The specification

This chapter has two boundaries to make real: another application can borrow the
GUI, and an Agent can delegate work without losing ownership of it. I require
both boundaries to use the existing objects, not a parallel framework that only
resembles them.

### A reusable GUI

The headless `Ensemble.csproj` remains independent of ASP.NET and the GUI.
`Gui/Ensemble.Gui.csproj` is an optional library containing the GUI's Common
declarations, functional partial implementations and embedded assets. The
example in `Examples/Gui` is its composition root, not the library API.

`Gui.RunAsync(agent, options, cancellation)` owns a private loopback host while
borrowing an existing Agent. `Gui.ServeAsync(host, agent, options, cancellation)`
mounts the same behavior in a caller-owned ASP.NET application. Mount before
starting that application, bind the loopback port declared by `GuiOptions`, and
await the serving task before destroying the Agent. Cancelling this view joins
its admitted requests, sockets and observation loop. It neither stops unrelated
routes nor destroys the borrowed Agent. `RunAsync` additionally stops its own
host. No callback bundle, service copy or second conversation owner is needed.

`Gui.MapAssets(routes, directory)` mounts just the shipped page and assets;
null selects embedded resources and a supplied directory selects an explicit
host-authorized override. The C# frontend keeps its existing modules; this
chapter does not invent the historical Go Artifact component API.

A literal `BasePath`, such as `/assistant`, prefixes page, assets, socket and
account controls. The validated OAuth callback remains `/auth/callback` at the
application root and must be reserved by the caller. Borrowed hosting does not
cancel shared caller-owned login attempts. Existing loopback, Host, origin and
account-request defenses remain in force. Keep the single-Agent wire behavior,
canonical replay, model control and public browser MCP relay intact.

### A terminal fleet

Each delegated Agent has two actual strong parents: Ensemble and its Agent
parent. Both directions are maintained by declared add/remove operations. A
child has its own actor, Skills, clients, History and Jobs. It does not acquire
private copies of the task workspace. State isolation is not process isolation
or protection against another child editing a shared file. Chapter 19 supplies
the optional narrow-only confinement boundary.

The shipped `ensemble` skill grants these seven tools:

| Tool | Contract |
| --- | --- |
| `spawn_sub_agent` | Validate the local ID and optional schema/model, create or restore a ready child, then admit its task through the existing actor. Return the child ID promptly while the existing Job observes the reliable task reply. |
| `join_agents` | Capture selected direct children, or all current children, and collect their current explicit submissions. Yield pending questions promptly while retaining the running Job handle. |
| `agents_status` | Query each current direct child through its actor. Report activity, pause, submission and failure separately; idle is not success. |
| `send_message_to_parent` | Ask the actual Agent parent, or the root's human inbox. Wait for a sender-bearing reply without completing the task. |
| `respond_to_parent_message` | Answer a child's pending question. Omit its message ID only when exactly one matching question is pending. |
| `submit` | Validate and durably publish the child's answer. Refuse unresolved retained children, invalid payloads and obsolete work generations. Ordinary prose is not submission. |
| `shutdown_agent` | Join and destroy the selected owned subtree, detach both parent relationships and preserve saved files. |

`CreateChildAsync` is construction, not admission. It uses the parent's current
immutable configuration and parsed skill templates, restores only the child's
History, and finishes skill initialization before returning. No factory or
SubAgentManager is added. Cancelling a spawn/join/question observer does not
cancel the independently owned child or withdraw an admitted question.

Incoming human input and child notices wake current Job observations without
killing their work. The pulse is non-sticky: no current observation means no
future interruption. The CLI keeps stdin live and accepts `/reply TEXT` or
`/reply MESSAGE_ID TEXT`; JSON mode accepts `kind: "reply"`. Human replies do not
admit a model turn. Child notices are labeled tool data, never system authority.

Delegated turns owe a submission. After a successful unsubmitted turn, issue an
immediate correction on the same reliable task reply, at most three consecutive
corrections per child. Durable submission resets the count. Execution failures
fail directly rather than consuming this correction budget. A standalone root
finishes in ordinary chat; the C# `submit` operation requires an Agent parent.
Submission can precede final tool-result recording and task drain. It is not a
shutdown barrier.

### Submission contract

A missing `submit_schema` selects a legacy object with required string `result`.
Extra keys are accepted and ignored only on this default path; the published
payload is normalized to `{"result":"text"}`. An explicit schema preserves the
whole validated argument object, even when it resembles the default schema.

The C# artifact is one JSON object in `submit.md`, atomically replaced rather
than appended as a JSONL ledger. Its fields are `version: 1`, `agent_id`,
`submit_schema` and typed `payload`. Validation precedes writing; publication
follows successful flush and replacement. The retained immutable contract is
copied before child construction. Accepted replacement work invalidates visible
completion; an older turn cannot submit for that newer work. Restoration loads
History, not a previous artifact as fresh completion. The artifact remains
inspectable on disk after shutdown, but cannot satisfy a new join by itself.

The supported JSON Schema Draft 2020-12 subset has an object root with
`type: "object"`; nested schemas may also be booleans. An optional `$schema`
must be `https://json-schema.org/draft/2020-12/schema`.

| Keywords | Meaning |
| --- | --- |
| `type`, `enum`, `const` | JSON types and exact value comparison |
| `properties`, `required`, `additionalProperties`, `minProperties`, `maxProperties` | Object fields and bounds |
| `items`, `minItems`, `maxItems`, `uniqueItems` | Array item schema, bounds and uniqueness |
| `minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum`, `multipleOf` | Exact numeric constraints |
| `minLength`, `maxLength`, `pattern` | Unicode code-point bounds and culture-invariant .NET regular expressions |
| `title`, `description`, `default`, `examples` | Annotations, without coercion or default insertion |

References, combinators, formats/content and other unsupported keywords refuse
at launch; validation never fetches schema resources. Validate keyword values,
not just names. Annotation and enum objects are data, not nested schemas.
Patterns have a 100-millisecond match timeout, which refuses the submission;
this is neither Go regex compatibility nor a whole-validation CPU deadline.
Exact finite-decimal arithmetic preserves integers beyond `double` precision.

Limits apply before recursive validation: 65,536 schema bytes, 1,048,576 payload
bytes, depth 64 (root at zero), 2,048 schema nodes and 65,536 payload nodes. Node counting visits
container values as well as primitives. Duplicate keys refuse. Number literals
are at most 128 characters with absolute exponent at most 1,000. Canonical
schema bytes multiplied by canonical payload bytes must not exceed 67,108,864.
The contract bounds hostile work; it is not an unrestricted schema service.

### Build and grade

From the repository root, using C# 14 and .NET 10:

```sh
dotnet build solutions/edition-3/main/Cli/Ch04.csproj
dotnet build solutions/edition-3/main/Examples/Gui/Gui.csproj
go build -o /tmp/ensemble-grade-edition2 ./cmd/grade-edition2
```

The C# executable route is `-ch 20 -executable-only -fleet-skills ABSOLUTE_DIR`
with a positional executable wrapper that starts the built `ch04.dll`. It scores
the seven fleet rows, 100 points. The separate 100-point Go GUI/reuse consumer
is excluded, not credited from C# tests. The historical Go route is unchanged.
Section 20.13 gives the wrapper and separate GUI consumer exercise. A score does
not inspect ownership, prove live-provider operation or establish accessibility.

## 20.2 Why the public boundary matters

The earlier GUI already projected one Agent's events and sent controls back to
that Agent. A reusable library should expose that work, not duplicate it. An
outside application supplies the existing Agent and decides where to host the
view. The GUI borrows the relationship; it does not become an owner that can
silently destroy the application's conversation.

In C#, a partial type spans source files in one assembly, not arbitrary DLLs.
The GUI's own Common declarations and implementations therefore move together
into `Ensemble.Gui`. The headless core keeps its own assembly and public API.
The example retains environment parsing, defaults and application construction.
This is a real optional dependency, not a directory renamed while its imports
still pull the web host into every consumer.

The new boundary must preserve replay, transient output, controls and browser
RPC. Moving code is not evidence that those paths still work. Use an outside
consumer, a built host and an isolated browser for their different observations.

## 20.3 The import test

The historical Go boundary exposed a useful failure: importing the core dragged
Gorilla WebSocket into a headless binary, and core types under `internal/` could
not be constructed by an outside caller. Moving files did not resolve either
problem. The test had to resolve actual imports and compile real public values.

The corresponding C# question is assembly reachability. `Ensemble.csproj`
explicitly compiles its core Common and functional directories and has no GUI
project reference or ASP.NET framework dependency. `Gui/Ensemble.Gui.csproj`
references that core and the ASP.NET shared framework. A plain external console
consumer must build with only the core reference; an ASP.NET consumer provides
the positive control by referencing the GUI library and calling its API.

Do not copy private declarations into a test to make it compile. Compilation of
the supplied consumer is the evidence for public accessibility. It still says
nothing about request routing, browser rendering or shutdown; test those paths
from the outside too.

## 20.4 The library must have a consumer

`GuiTests` is a separate-assembly public consumer, not another functional module
of the core. It creates its own empty working directory before creating assets
or skills. That detail matters: an early launch from the source directory
allowed a fixture's lowercase `skills/browser` path to collide with the shipped
`Skills` tree on a case-insensitive filesystem. The fixture was corrected to
own its workspace. Launch-directory discipline must not be the only protection
against a test overwriting product files.

The consumer borrows an Agent twice, mounts a view at `/assistant` beside
`/health`, and tests embedded assets and explicit overrides. It sends real
WebSocket controls, exercises ordinary browser MCP discovery/calls and pending
RPC cancellation, and verifies that stopping the GUI detaches its relay while
leaving the caller's Agent and unrelated route usable. Local account preparation
checks do not contact an issuer or establish successful OAuth.

The existing browser relay is `Agent.Mcp().Browser()`, not a new dispatcher.
Discovery begins when an activated skill needs it; merely opening a page without
a granting skill must not be mistaken for a promised discovery handshake.
Disposal aborts sockets, so a fixture cannot invent a graceful close-handshake
requirement and call the existing lifetime behavior a product defect.

Serve the page without a trailing slash as well as with one. Absolute URLs from
the old standalone host can otherwise escape a mount prefix while the initial
HTML still appears to work. Account controls follow the prefix; the deliberately
fixed root callback is documented separately, not quietly weakened to fit a test.

## 20.5 How a test is invoked is part of the test

An outside consumer needs explicit project references and an owned working
directory. Build before launching so the supervised process is the application,
not a build wrapper with a surviving child holding stdout open. Retain real exit
codes and distinguish compilation, HTTP delivery, browser behavior and joined
shutdown. None is a substitute for the others.

The historical Go harness initially used the wrong executable path and omitted
a required model. Those were setup failures, not failures of fleet behavior. The
same distinction applies to C#: `Cli/Ch04.csproj` produces lowercase `ch04.dll`,
and `Examples/Gui/Gui.csproj` is the actual web host, not `GuiStub`.

The Chapter 20 executable adapter keeps its own local-provider wire records and
uses explicit shipped skill selection. Its early attempts assumed shell tools
were granted by the `ensemble` skill, compared observation fields as payload
schema, and expected all Job output in the final cursor read. Those were grader
mistakes. Fixing them does not earn a product correction; identifying a missing
sender field in question replies did. Keep that distinction visible in the
record rather than presenting every failed test as a software defect.

## 20.6 Spawning

A child is a full Agent with both its Ensemble parent and Agent parent supplied
at construction. Ensemble owns the process-wide collection; the Agent owns its
direct children and their lifetime. Local child IDs are unique among siblings,
not a second global name registry. The declared parent operations validate
before mutation, establish both links before returning, and remove both links
during destruction. No manager, copied service bag or callback factory is needed.

`CreateChildAsync(id, model, submitSchema, cancellation)` constructs a ready
child, not a running task. It derives Chapter 19's narrow-only policy, captures
parsed skill templates and current values, creates independent Skills and MCP
clients, configures `children/<id>/save.json` beneath the parent's persistence
directory, restores that child's History, and awaits skill initialization. It
then returns the actual Agent. The ordinary `SubmitAsync` operation admits work.
The model-facing spawn tool composes these two operations and observes the
reliable reply through its existing Job.

Take configuration/template snapshots before acquiring locks that their owners
might need in reverse order. Sandbox policy derivation must not hold its gate
while reaching back into Agent ingress. Recheck admission after preparation;
shutdown racing with creation must not attach a half-live child.

A new optional model keeps the parent's provider and billing route. A schema is
validated and copied before construction effects. Descendants without an
explicit schema get the default contract, not their parent's answer format.
Configuration and templates may be immutable values; live clients, activation
relationships and mutable variable collections must not be shared.

Creation failure joins cleanup and preserves files, permitting a same-ID retry
after detachment. Successful admission leaves an independently owned child even
if the spawn observer stops waiting. Recursive shutdown joins descendants and
attempts all cleanup, preserving errors instead of allowing the last `finally`
to hide an earlier failure. Explicit destruction detaches both parents. Saved
work survives; process and relationship lifetime do not.

## 20.7 The parent and the user

A child reaches its parent through `AgentParent()`, not by searching an ID map.
The root's human inbox belongs to that root. There is no fabricated human Agent
to spawn, destroy or include in a child status list. Uniform asking does not
require pretending that a person has the lifecycle of an Agent.

A pending question is an edge strongly owned by its asking and receiving Agents.
For a human question those owners are the same root. Ensemble's question lock
changes both links atomically. Code holding it must not acquire Agent ingress;
a labeled child notice crosses to the actor only after that lock is released.
Answering or destroying the edge removes both links before completing its
waiter, and never destroys either Agent.

Question IDs are allocated at the recipient, where answers must be unambiguous.
Omitting the ID is convenient only when exactly one matching question remains.
Wrong IDs and ambiguity refuse without consuming anything. Caller cancellation
stops only that wait; the admitted question still needs an answer or owner
shutdown. Shutdown of either endpoint cancels it and detaches both links.

A reply is `QuestionReply(Sender, Response)`, an immutable value captured before
detachment. Agent answers use the actual recipient's child ID or name (the
unnamed root uses `root`). Human answers default to `user`; a library host can
supply a nonblank human address to `RespondToQuestion`. That argument cannot
forge an Agent sender. The tool returns both `sender` and `response`.

This explicit contract caught a C# omission during grader reconciliation: the
first implementation returned text alone. The correction changed the existing
question edge's completion value, not its owner or routing mechanism. A focused
failing assertion established the missing field before the correction.

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
{"summary":"Checked the change.","findings":[{"path":"Program.cs","fixed":false}]}
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
truncating it. Section 20.1 specifies that public boundary. No schema resolution
fetches a URL or reads a local file.

Without a custom schema, `submit({"result":"text"})` keeps its existing meaning.
The public string convenience method encodes that same default shape and enters
the same validator, so it cannot bypass a custom contract. Structured joins, submission readback and the durable artifact preserve the
complete JSON object. Status reports whether submission occurred, separately from
activity and failure; it does not duplicate the payload. Text-valued displays can
format a payload for a reader without changing its stored type.

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


In C#, each Job owns a replaceable observation pulse. An observer captures it
once, after acquiring the observation semaphore. A message replaces and completes
the old pulse; a later observer sees only the fresh one. `Jobs.WakeObservers`
walks existing Jobs rather than maintaining another registry. Execution tokens,
shared output cursors, attention output and real interrupt/shutdown semantics
remain separate. A wake returns the same running handle, not a cancelled task.

The actor records human hints before pulsing. Child notices use `HintReceived`
with `Actor.Tool` and labeled question data, never system/developer instructions.
An idle parent records that notice without starting inference; root-human
questions stay in the human inbox rather than waking the root to answer itself.
Chat input uses actor-decided turn-or-hint admission. JSON `prompt` and `hint`
retain their separate meanings. Displaying a human question observes the inbox;
it does not consume it or create another pending-question store.

## 20.10 Three wakes, then tell someone

“Done” has a precise meaning for a delegated Agent: it called `submit`, and its
children have finished. A normal model response ending is not sufficient for
that worker. The human-facing main Agent and named peers finish normally in chat;
they have no Agent parent awaiting a submission. The framework can enforce the
delegated protocol without deciding whether the result is good.

First, `submit` rejects while any retained child lacks a current submission. The error tells the
parent which children need attention. Submissions happen bottom-up. A failed
child is unresolved too; shutting it down removes it from the managed set while
preserving its files. Failure does not quietly become a successful child result.

Second, a delegated Agent's successful turn that ends without submission receives
an immediate framework correction. With active children, the correction asks the Agent to
join or shut them down. Without active children, it asks for the missing submit.
No polling timer is needed: the runtime is already handling turn completion.

The correction lives in the library's real completion path and checks the actual
Agent parent. A standalone root remains a chat agent; an ordinary delegated
child owes its parent a result. Construction, tool grants and delegated completion remain distinct decisions.

Count consecutive framework wakes per delegated Agent and reset the count only
after a durable submission. New admitted work clears old failure, not this count.
After exactly three wakes, stop re-prompting and escalate to the parent. That
parent can explain the failure to the human or resolve the child before continuing.
The bound limits attempted correction; it cannot force a model to cooperate or
guarantee a useful result.

In the historical Go exercise, that distinction produced a student correction. An exhausted child first
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

## 20.13 The C# exercise

Extend the preceding C# endpoint. Demonstrate public GUI hosting and terminal
delegation separately, then check their shared ownership paths. Keep the GUI's
single-Agent wire behavior intact; a fleet panel belongs to the next chapter.
Use owned directories and loopback ports, never an existing user's browser or
running server. The following commands assume .NET 10 is on your chosen PATH;
they do not install it or select a private credential store.

### An outside GUI application

From the repository root, create an external ASP.NET project with a public
project reference. Use `main` while implementing and the preserved `ch20`
endpoint after finalization:

```sh
ensemble_source=$(pwd)/solutions/edition-3/main
reuse_dir=$(mktemp -d /tmp/ensemble-ch20-reuse.XXXXXX)
cat > "$reuse_dir/Reuse.csproj" <<EOF
<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <TargetFramework>net10.0</TargetFramework>
    <LangVersion>14</LangVersion>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>enable</Nullable>
  </PropertyGroup>
  <ItemGroup>
    <ProjectReference Include="$ensemble_source/Gui/Ensemble.Gui.csproj" />
  </ItemGroup>
</Project>
EOF
cd "$reuse_dir"
```

Save this as `Program.cs`. Supply an unused loopback port when launching. This
example uses no credential and makes no inference request; it demonstrates a
mounted view, replay, and explicit lifetime ownership.

```csharp
using System.Net;
using Ensemble;
using EnsembleGui;

int port = int.Parse(args[0]);
var builder = WebApplication.CreateBuilder(Array.Empty<string>());
builder.Logging.ClearProviders();
builder.WebHost.ConfigureKestrel(server => server.Listen(IPAddress.Loopback, port));
await using var host = builder.Build();
host.MapGet("/health", () => "caller-owned");
var root = new Ensemble.Ensemble(Console.Error);
using var stop = new CancellationTokenSource();
Task? serving = null;
try
{
    var agent = new Agent(root, new Configuration(Vendor.OpenAI, "offline",
        baseUrl: "http://127.0.0.1:1"), name: "demo");
    agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
        parts: [new TextPart("PUBLIC_REPLAY from the caller's Agent")]));
    serving = Gui.ServeAsync(host, agent,
        new GuiOptions(port, BasePath: "/assistant"), stop.Token);
    if (serving.IsCompleted)
    {
        await serving;
    }
    await host.StartAsync();
    Console.WriteLine($"http://127.0.0.1:{port}/assistant");
    Console.WriteLine("Press Enter to stop. This example does not send a prompt.");
    // Observe hosting failure while the caller waits for its stop request.
    Task<string?> input = Console.In.ReadLineAsync();
    await Task.WhenAny(input, serving);
    if (serving.IsCompleted)
    {
        await serving;
    }
    else
    {
        await input;
    }
}
finally
{
    stop.Cancel();
    try
    {
        if (serving is not null)
        {
            await serving;
        }
    }
    finally
    {
        try
        {
            await host.StopAsync();
        }
        finally
        {
            root.Destroy();
        }
    }
}
```

Build `Reuse.csproj`, then run its `bin/Debug/net10.0/Reuse.dll` with your selected
port. Inspect `/health` and `/assistant`, and use an isolated browser to observe
replay and controls. Press Enter for joined shutdown. To test an override, pass
an explicit `AssetsDirectory` in `GuiOptions`; do not derive that path from model
input. To test assets alone, mount `Gui.MapAssets(host.MapGroup("/assets"))`
without `ServeAsync`. That serves assets, not a second live Agent session.

Build a separate console consumer with only `Ensemble.csproj` referenced and
construct an Ensemble and Agent using the same public types. Inspect its
resolved dependencies alongside the GUI consumer's positive control. Then use
`GuiTests` for the broader public-API regression: repeated borrowed hosting,
mounted routes, browser RPC, account guards and cancellation with work pending.
Neither compilation nor HTTP delivery proves pixel rendering, audible speech or
human accessibility.

### A fleet with work to hand back

Build the actual CLI and create a small wrapper. Use an absolute DLL path so the
process can run from its own task workspace:

```sh
cd "$ensemble_source"
dotnet build Cli/Ch04.csproj
fleet_dir=$(mktemp -d /tmp/ensemble-ch20-fleet.XXXXXX)
dotnet_path=$(command -v dotnet)
cat > "$fleet_dir/cli" <<EOF
#!/bin/sh
exec "$dotnet_path" "$ensemble_source/Cli/bin/Debug/net10.0/ch04.dll" "\$@"
EOF
chmod +x "$fleet_dir/cli"
cd "$fleet_dir"
EN_SKILLS_DIR="$ensemble_source/Skills" EN_PRIMARY_SKILL=ensemble \
EN_STARTUP_SKILLS=code-tools "$fleet_dir/cli" chat
```

Select the intended provider/model and billing route explicitly using the
previous chapters' host options and an already authorized credential source.
The command does not create credentials. The `ensemble` primary grants fleet
tools; the separately activated `code-tools` skill grants file and shell work.
A prompt naming a tool grants nothing.

Ask for three children, `add`, `multiply` and `subtract`. Each writes and tests a
small Python function and explicitly submits its result. Have `add` ask whether
negative integers are required before implementing. Require the parent to answer,
join all three typed results, run combined assertions and finish in ordinary chat.
While a join waits, type a hint asking for `STEER_COBALT` in the final answer.
Collect the same running Job after the wake, then inspect the child saves and
`submit.md` artifacts and run the assertions independently. Delivery of a hint
and model compliance are separate observations.

Also have the root ask the human a question. Answer with `/reply TEXT` when one
is pending, or `/reply MESSAGE_ID TEXT` otherwise. In JSON mode, use:

```json
{"kind":"reply","message_id":1,"text":"Include negative integers."}
```

`message_id` is optional only for one pending human question. A wrong ID must
refuse without consuming it. Quit while a question is pending too: closing the
question should become an ordinary tool outcome, not an unhandled Job failure.

Finally launch the structured reviewer from §20.8. A controlled local provider
should omit `fixed`, then correct it to boolean `false`; only the corrected
object may satisfy a join. Check nested values and the durable artifact, default
text and ordinary-grandchild controls, early parent-submit refusal, exhaustion
after three corrections, and a failed provider turn with no correction loop.
These controlled failures need not be entrusted to a live model to reproduce.

From the repository root, the scoped executable check is:

```sh
/tmp/ensemble-grade-edition2 -ch 20 -executable-only \
  -fleet-skills "$ensemble_source/Skills" -json "$fleet_dir/cli"
```

This driver uses its own local provider and state directories. Keep its reports
and actual process exits; do not transfer its 100 points to the excluded Go
consumer or to live integration.

## 20.14 Taking it for a spin

The C# implementation was exercised locally with controlled HTTP peers, actual
CLI processes, an outside GUI consumer and an isolated muted browser. JSON and
chat CLI runs drove a root, child and grandchild, routed parent and human
questions, retained typed submissions and removed the subtree after joined
shutdown. The chat run also accepted a busy hint and refused a wrong reply ID
without consuming the question. A separate pending-question `/quit` exercise
initially exposed an unhandled cancellation; the corrected question-tool
boundary reports independently cancelled questions as ordinary tool errors.

Sender attribution then received a focused failing test and correction. The
final shared regression was 191 passing groups, with clean builds of both actual
hosts. The scoped executable grade was 100/100 across five clean host sessions
and 59 local requests. Earlier unsuccessful grader runs included fixture errors;
the current progress record retains them rather than treating every score as a
product regression. GUI mounting, replay and browser relay evidence have their
own scope; the executable fleet score does not include a GUI consumer score.

These are local results, not live-model task quality, account entitlement,
provider billing, audible speech or human accessibility. Live OAuth remains a
separate deferred exercise. The following second-edition account is preserved
history, not another measurement of this C# implementation.


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

The historical combined rubric below has seven GUI/reuse rows and seven fleet
rows, 200 points. C# executable-only grading uses only the final seven rows,
100 points, including structured-submission assertions within those same rows.
It excludes the first seven Go consumer rows. The original Go route remains
available; neither its historical failures nor its consumer successes are C#
evidence. Public C# consumers and browser exercises retain separate results.

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
| Child state and construction | Actual construction and distinct paths; structured durable row; default contract on an ordinary grandchild | 15 |
| Fleet visibility and shutdown | Status reports a managed child; shutdown removes it | 10 |
| Fan-in | Selected children all submit; correlated results preserve arrays, booleans and nulls without string encoding | 15 |
| Questions and replies | Three-generation question/reply routing; fan-in yields so the parent can answer | 15 |
| Wake a running wait | Incoming input wakes the wait and its background job remains collectible | 15 |
| Bounded correction | Early submit fails; delegated missing submission receives three wakes; main chat finishes without a wake | 15 |

The structured fixture uses a non-memory reviewer contract. It checks missing,
wrong-type, nested and extra fields, then a valid correction; invalid, unsupported
and referenced schemas must refuse launch. Parse actual correlated tool results
and saved rows rather than searching a frame for a marker. Defensive copies,
contract ownership and string convenience validation retain separate public-API
regressions; those source tests award no numeric credit to another implementation.

An import check does not inspect the runtime parent chain. HTTP delivery of a
JavaScript module does not show a browser rendering it. Keep the actual construction,
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
