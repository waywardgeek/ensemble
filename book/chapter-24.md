# Chapter 24: Freeing the GUI

Two chapters ago we gutted the architecture. Closure staples, dead
settings, a constructor that took six dependencies because nobody had
drawn the ownership arrow. Chapter 22 repaired real decay, and it
felt thorough. Then I wanted the Artifact scroll for Homebrew-VTT.

The scroll is a self-contained JavaScript component that renders a
streaming event log. It has no framework dependency. It works in any
browser. And it lives behind `internal/`, where Go forbids any other
module from importing it. The most reusable code in the repository
is the code nobody else is allowed to touch.

That flaw survived chapter 22. It survived the chapter *about*
architectural decay. Nobody noticed because nobody tried to use the
GUI from outside, and a flaw that goes unexercised passes every test.
Decay is an event you fix once only in retrospect. In practice it is
a gradient you keep checking.

---

## TL;DR

The framework must not know the GUI exists. Right now it does: the
root package imports the websocket server, so every consumer of the
agent library links a GUI it never started and `gorilla/websocket`
along with it. The GUI also sits under `internal/`, where Go forbids
any outside module from importing it. This chapter fixes both, and
the proof is one command: `go list -deps` on a headless consumer
names neither the GUI package nor gorilla.

1. **Delete the dead exports.** `WSHub`, `NewWSHub`, `ServeHTTP`,
   `NewMCPWSTransport` have zero consumers anywhere in the repo.
   Remove them, and the root package drops its `internal/ws` import
   entirely.
2. **Move the GUI to a public package.** `internal/ws` becomes
   `gui`. `ws.Hub` becomes `gui.Server`. The wire protocol stays
   identical; a chapter 22 behavioral session passes without changes.
3. **Bundle the agent dependencies.** The server currently takes six
   constructor parameters plus a closure set after construction. Pack
   them into a single `AgentHooks` struct defined in `gui`, passed at
   construction time. The post-construction `Model` assignment and
   its apology comment go away.
4. **Embed the assets.** Use `go:embed` so the binary serves the GUI
   from any working directory. Add a flag that overrides with a disk
   path for live JS development. Export the `embed.FS` so a consumer
   can mount individual components (the Artifact scroll) without
   taking the full application shell.
5. **Cut the second gorilla path.** The MCP client-WS dial transport
   in `internal/mcp` also imports gorilla. Move it to a public
   `mcpws` package. After this, the framework root links no
   websocket code at all.
6. **Export the event vocabulary.** Root-alias `Event`,
   `MessageData`, `PartList`, `TextPart`, and the actor constants so
   an outside consumer can build a log the scroll knows how to
   replay. You cannot construct what you cannot name. Keep the MCP
   relay (`ServeMCP`) in `gui`: an agent that can see this GUI
   matters as much as a human who can.
7. **Nothing else changes.** Full chapter sweep green. The chapter 5
   package-var rule gains one carve-out for `embed.FS` variables,
   because `go:embed` has no var-free form.

The reference implementation uses `gui`, `mcpws`, `Server`, and
`AgentHooks` as names. You choose the exact field spelling, the
override flag name, and where the asset files live inside the
package.

Build and grade: `make grade24`.

| # | Behavior | Check | Points |
|---|---|---|---|
| 1 | Framework links no GUI, no gorilla | `go list -deps` on a root-only consumer compiled OUTSIDE the module | 20 |
| 2 | GUI publicly reusable | External consumer with fake hooks: planted event replays to a ws client, a prompt reaches the Send hook, MCP relay answers JSON-RPC with the request's own id | 15 |
| 3 | Wire unchanged | Chapter 22 scripted session against the real binary; recorded vendor requests prove the model switch landed | 15 |
| 4 | Assets embedded | Binary started from an unrelated working directory serves the GUI | 15 |
| 5 | Components without shell | Consumer mounts the Artifact scroll from the exported FS; `index.html` returns 404 | 10 |
| 6 | MCP without gorilla | Headless MCP consumer links no gorilla; virtual-user still dials through `mcpws` | 15 |
| 7 | Root stays clean | No websocket-typed symbols exported, no `gui` or gorilla imports in root | 10 |

## The physics

Go links what you import. A binary that imports a package pays for
every line of that package and every package it in turn reaches,
whether the binary calls any of it or uses it at all. If package A
imports package B and B imports package C, then A links C, and C's
dependencies, and theirs, recursively. The compiler has no mechanism
to say "I import B but I do not want C." Either the import is there,
or it is absent, and everything reachable from it comes along.

`internal/` is the reverse rule. Code inside an `internal/` directory
can only be imported by code rooted at its parent. A module outside
the tree cannot reach it even if it knows the import path.
`go build` refuses outright.

The GUI had both problems at the same time. It sat inside
`internal/`, so the most complete browser frontend in the
repository was invisible to every application outside the
module. And the framework root imported it, so every application
inside the module carried it: a headless batch tool, a CLI that
processes files without ever showing a browser, a test harness that
exercises the model loop and has no use for websockets. All of them
linked the server, the browser-RPC correlation layer, and
`gorilla/websocket`. The dashboard was welded to the engine.

The fix is two inversions. Move the GUI out of `internal/` so other
applications can import it. Remove the import from the framework root
so applications that do not want the GUI stop paying for it. Both
changes are mechanical, and the proof that they worked is a single
command: `go list -deps`.

## 24.1 The archaeology

A refactor begins with a census. Grep before moving.

The GUI server lives in `agent/internal/ws`: 2,603 lines across five
files. `handler.go` (1,041 lines) runs the websocket hub, client
lifecycle, event replay on connect, and inbound message routing.
`gui_call.go` (187 lines) handles browser-RPC correlation for the
`view_gui` tool. `mcp_port.go` (136 lines) tunnels MCP JSON-RPC
through the websocket. Two small files handle TTS and GUI logging.
The total is modest, and the package depends only on
`internal/common` plus `gorilla/websocket`. It has no dependency on
the agent, the engine, or the actor. The GUI never needed the
framework; it only needed to stop living inside it.

The dependency graph before this chapter looks like this:

```
  agent (root)
   ├── internal/common
   ├── internal/llm
   ├── internal/mcp ─── gorilla/websocket  ← path 2
   ├── internal/ws ──── gorilla/websocket  ← path 1
   └── ...
```

Every package in this tree is pulled into any binary that imports the
`agent` root. A tool that only wants the model loop gets the
websocket server free of charge.

The JavaScript side sits in `agent/web/gui`: 2,762 lines. Most of it
is two monoliths, `gui.js` (669 lines) and `mcp.js` (632 lines),
that a future chapter might componentize. One file is already a clean
component: `artifact-scroll.js` at 264 lines, the scroll that
renders a streaming log and the component Homebrew-VTT wants. It has
no dependency on anything else in the repository. Extracting it
requires extracting the package it lives in.

The framework root, `agent/agent.go`, re-exports four websocket
symbols to the public API:

```go
type WSHub = ws.Hub
func NewWSHub(...) *WSHub { ... }
func (a *Agent) ServeHTTP(...) { ... }
func NewMCPWSTransport(hub *WSHub) ... { ... }
```

A grep across every Go file in the repository finds zero consumers
for any of them. `cmd/main.go` calls `ws.NewHub` directly. Nothing
uses the type alias. Nothing calls `ServeHTTP`. Chapter 8 published
these exports as the framework's GUI API, the command binary never
routed through them, and no external consumer existed to notice. The
API was born dead and stayed dead for sixteen chapters. This is the
shape of decay when it meets clean tests: nothing broke, nothing
warned, and the dead code accreted quietly because it never ran.

There is a fifth re-export, `NewMCPClientWSTransport`, and it has
one real consumer: `cmd/virtual-user/main.go`. This one matters and
this one stays, but it does not belong in the framework root, and
the reason takes another section to explain.

## 24.2 Two paths to gorilla

The obvious fix is obvious. Delete the dead exports, move
`internal/ws` to a public package, and the framework root drops its
websocket import. A headless consumer no longer links the GUI.

Run `go list -deps` on a headless consumer after that fix, and
gorilla is still there.

The dependency enters the framework through two doors. The first is
`internal/ws`, which imports `gorilla/websocket` for the hub's client
connections. The second is a single file in a different package:
`internal/mcp/client_ws_transport.go`, which imports gorilla for the
MCP client's dial transport. The root package imports `internal/mcp`
for its MCP tool aliases, so every framework consumer links gorilla
through `mcp` even after the `ws` door closes.

Only `cmd/virtual-user` uses the client dial transport. No headless
framework consumer needs it. The fix is the same pattern: move
`client_ws_transport.go` to a new public `mcpws` package, let
`virtual-user` import it directly, and `internal/mcp` drops its
gorilla import.

The lesson is older than Go modules. `go list -deps` shows every
transitive dependency a package reaches. Run it before the refactor
to know where you stand, and run it after to know whether you
actually moved. Cutting the import you know about and calling it done
is the dependency-graph equivalent of fixing the bug you can see and
skipping the regression test. The tool that verifies the work is the
same tool that should have scoped it.

Gorilla itself is blameless in this story. It is a well-maintained,
minimal websocket library that does exactly what its documentation
says. The defect is placement: two files in the framework's internal
tree import it, and those files sit on import paths that every
consumer must traverse. Removing gorilla from the headless closure
does not mean removing gorilla from the project. It means removing
it from the path where it does not belong. The GUI still imports it,
`virtual-user` still dials through it, and the dependency exists
precisely where its consumers are.

The mechanical move itself has one footnote worth preserving. Bulk
renaming `ws.` to `gui.` across the repository requires word-boundary
matching. BSD `sed` on macOS silently treats `\b` as a literal
backslash followed by the letter b, producing no matches and no error.
`perl -pi -e 's/\bws\./gui./g'` works. The failure mode that should
be noisy and is instead quiet is the one that survives longest.

## 24.3 The split

The Hub carries two jobs in one struct. The transport half manages
websocket clients, replays the event log on connect, broadcasts
JSON-RPC frames, correlates browser-RPC calls, tunnels MCP, and
serves static assets. The agent-session half reaches into the agent
through six constructor parameters and one post-construction closure.

The constructor signature tells the story:

```go
func NewHub(gate *common.PauseGate, send func(common.Inbound),
    guiLogPath string, eventLog *common.Log,
    settings common.SettingsSource, usage common.UsageSource) *Hub
```

Six parameters, five of them reaching past the GUI into the agent's
internals. And a seventh dependency set after construction:

```go
// handler.go:39
Model func() string

// cmd/main.go:582
hub.Model = func() string { return eng.Cfg.Model }
```

A closure assigned after the constructor returns, with a comment at
the assignment site that amounts to an apology. The reason the
closure exists is real: the operator can switch models mid-session,
and a string captured at construction would price every later turn at
the old rate while looking entirely correct. But a field set after
construction is a promise the compiler cannot check. If the command
binary forgets the assignment, the hub silently reports an empty
model for every session, and every usage display in the GUI shows the
wrong price. The bug would be invisible until someone reads a billing
statement and wonders why the numbers disagree.

The pattern is recognizable in any codebase with enough history: a
struct that needs a dependency it cannot receive at construction, a
field assignment somewhere else in the program, and a comment
explaining why this was necessary. The comment is the diagnostic.
Code that needs a comment to justify its wiring has wiring that the
type system did not enforce, and the next contributor will read the
code before reading the comment.

The fix is a construction-time bundle:

```go
type AgentHooks struct {
    Gate     *common.PauseGate
    Send     func(common.Inbound)
    EventLog *common.Log
    Settings SettingsSource
    Usage    UsageSource
    Model    func() string
}

func New(hooks AgentHooks, guiLogPath string) *Server
```

The `Model` closure survives inside the bundle because the reason it
exists has not changed. What changes is when it arrives: at
construction, alongside the five other dependencies that were already
there. The compiler enforces the presence of every field in the
struct literal.

`AgentHooks` is defined in `gui`, where its consumer lives. The
framework's `common` package gains nothing from this type: no other
package constructs it, and carrying shapes for one consumer is how
hub packages grow into the problem that chapter 22 diagnosed.

Bill saw the dotted line before the refactor began. His framing:
the Hub could become Ensemble, the class that manages multiple agents
when sub-agents arrive. The transport half should stay a GUI server;
Homebrew-VTT wants it without any multi-agent machinery. The
session-bundle half is what grows: the thing that owns N agents,
routes inbound messages by agent ID, and fans source-tagged
observations out to however many frontends are attached. The observer
seam already supports multiple observers; the attach point exists.

This chapter does not build Ensemble. It carves the seam. One
`AgentHooks` value represents one agent's control surface handed to
one frontend. The scroll replays one agent's log, the send hook
delivers prompts to one agent's inbox, the model getter prices one
agent's turns. The struct makes the question visible: what happens
when there are two agents? What holds the map from agent ID to hooks?
Chapter 25 answers that question, and the answer is a new type with
a new name, because a rename would carry four old connotations from
a struct that is about to stop existing.

## 24.4 Embedding the assets

The GUI assets live on disk at `web/gui/` relative to the module
root. The binary serves them through `http.FileServer` pointed at
that disk path. Start the binary from any other directory and the GUI
returns 404 on every request.

This is a concrete failure that real deployments hit. The chapter 22
grader harness works around it by setting `cmd.Dir` to the repository
root before launching the binary. Every deployment, every test
harness, every script that starts the agent from a CI runner or a
cron job must know where the source tree lives. The binary carries
its own intelligence and its own tools but cannot find its own
interface. A user who moves the binary to `/usr/local/bin` and runs
it from their home directory gets a working agent with no GUI, and
the error message is not "assets not found" but a silent 404 on every
HTTP request, because `http.FileServer` on a missing directory serves
an empty listing rather than refusing to start.

`go:embed` makes it self-contained:

```go
//go:embed web
var Assets embed.FS
```

Two lines. The compiler bakes the entire `web/` directory into the
binary at build time. `http.FS(Assets)` serves them from memory, from
any working directory, with no runtime dependency on the source tree.

A `--gui-dir` flag overrides the embedded assets with a disk path
for live development: edit a CSS file, reload the browser, see the
change without rebuilding the Go binary. The override is the
development path. The embedded default is the production path.
Neither depends on the working directory.

`go:embed` requires a package-level variable. There is no alternative
syntax and no function form. The `var` declaration is the embed
directive's anchor. Chapter 5's grader checks for package-level
variables because mutable globals are a structural defect in
concurrent code, and `^var` in a non-test file is the simplest
detector that catches the pattern.

An `embed.FS` is immutable in practice. It is initialized at compile
time, and the `embed` package provides no method to modify its
contents after the binary starts. The chapter 5 rule gains a
carve-out for `embed.FS` variables, stated as a type rather than a
package path. Rules that are right deserve amendment when a legitimate
case appears. Rules that get gamed instead of amended accumulate
exceptions until the exceptions are the rule and the rule is a
memory.

## 24.5 The proof

The chapter's thesis is mechanical: the framework must not know the
GUI exists. A mechanical thesis deserves a mechanical proof, and the
tool is `go list -deps`, which prints every package in the transitive
closure of a target's imports.

```
$ go list -deps ./headless-consumer/ | grep -E 'gorilla|gui'
$
```

No output. The headless consumer imports the framework root, uses the
agent, and links neither the GUI package nor gorilla. The same
command on a consumer that imports `gui`:

```
$ go list -deps ./gui-consumer/ | grep -E 'gorilla|gui'
github.com/gorilla/websocket
agent/gui
```

Both paths appear, because `gui` depends on gorilla and Go links
what you import. The headless consumer does not import `gui`, so
neither appears. The proof is in the absence, and the positive
control is in the presence: the same detector, the same `grep`,
applied to a different consumer, shows that the tool fires when its
target is there. A test that only checks the negative can pass by
accident; a test that also checks the positive proves the detector
works before trusting its silence.

The grader compiles both probes outside the module, in a temporary
directory with a `replace` directive pointing at the submission and a
copy of the module's `go.sum`. The probes run with `GOPROXY=off` so
the test does not depend on network access to a module proxy.
An in-tree test cannot prove that `internal/` stopped blocking
external consumers, because in-tree code is already inside the module
boundary. The only honest test of a public API is a consumer that
stands outside and tries to use it.

## 24.6 Reuse for real

The headless proof shows what the framework excludes. A reuse test
shows what the GUI includes.

A consumer outside the module imports `gui`, constructs a
`gui.Server` with fake hooks, and starts serving. No ensemble
binary, no agent, no model. A test websocket client connects, and the
server replays the event log, including a sentinel event the consumer
planted before the client connected. The consumer's `Send` hook
receives a prompt the client typed. Three interactions, three proofs:
the scroll works, the input path works, and no ensemble machinery
was required to make them work.

Building the test fixture surfaced a gap the design doc missed. The
event vocabulary was internal. A consumer could construct a
`gui.Server`, but it could not construct an `Event` to put in the
log, because `Event`, `MessageData`, `TextPart`, and the part list
types lived in `internal/llm`. A GUI with an empty log renders an
empty scroll. You cannot construct what you cannot name.

The fix is root aliases:

```go
// agent.go
type Event = llm.Event
type MessageData = llm.MessageData
type PartList = llm.PartList
type TextPart = llm.TextPart

const MessageReceived = llm.MessageReceived
const ActorHuman = llm.ActorHuman
const ActorAgent = llm.ActorAgent
```

One collision required a decision. The `Actor` type itself is already
taken at root by the actor runtime from chapter 6. The event actor
constants carry the type (`llm.Actor`), which is all a consumer needs
to build an event. The type stays unaliased, and a consumer that
somehow needs to spell it out can import `llm` directly. A flat
namespace forces exactly this kind of choice, and the right answer is
to alias the things consumers construct and leave the things they
receive.

Bill's ruling mid-implementation added a fourth leg to the reuse
test. `ServeMCP`, the relay that gives an agent visual access to the
GUI through tools like `gui_snapshot` and `gui_click`, moved with the
package and was already public: a method on `gui.Server` taking a
bare `net.Listener`, with no hidden dependencies. An agent that can
see the GUI is as much the point as a human who can, and the grader
locks both paths into the reuse surface.

The relay has one subtlety worth knowing. When no browser is
connected, `ServeMCP` answers immediately with an error carrying the
JSON-RPC request's own id. But a connected websocket client that
ignores MCP frames parks the relay forever, waiting on a browser
reply that will never come. The grader probe dials the MCP relay
before any websocket client connects, testing the clean error path
rather than the timeout path. Read the dispatch condition in the
code, not just the error constant's name.

The exported `embed.FS` completes the reuse surface. A consumer can
mount `artifact-scroll.js` and `renderers.js` from the embedded
filesystem without taking `index.html` or the application shell.
Homebrew-VTT gets the scroll component and its renderer, served from
its own binary, with no ensemble dependency at runtime.

## 24.7 Taking it for a spin

Build the binary. Start it from `/tmp`. The GUI loads in a browser
at the configured port, served from the embedded assets, with no
`web/gui/` directory anywhere nearby. The chapter 22 behavioral
session runs unmodified against the moved package: subscribe, drive a
turn with a model switch, and the recorded vendor requests prove the
wire protocol has not changed. The model switch lands at a turn
boundary, so the recorded sequence is `[opus opus sonnet sonnet]`
rather than an instantaneous flip, and the grader asserts that shape:
the first request went to the expensive model and the last to the
cheap one.

The external consumer compiles from outside the module, stands up a
`gui.Server` with fake hooks, and serves the Artifact scroll on its
own port. No ensemble binary is running. A websocket client connects,
and the planted sentinel appears in the replayed log. A prompt typed
into the client arrives at the `Send` hook. The MCP relay answers a
JSON-RPC request with the correct id. Three interactions that each
would have failed before this chapter: the first because the package
was internal, the second because the event vocabulary was internal,
and the third because nobody had tested `ServeMCP` from outside the
framework.

One grader war story belongs in the record, because it illustrates a
failure mode that returns in any project that wraps binaries. The
first grader run appeared to hang for ten minutes, and the report was
already finished. Killing a `go run` wrapper orphans the grandchild
process, because `go run` is a build tool that does not forward
signals. The grandchild inherits stderr and holds the output pipe
open: `tail -f` on the grader's log never sees EOF because the
orphan is still writing to the same file descriptor. The fix is to
build the probe to a real binary and exec it directly, the same
lesson as chapter 22's `Setpgid` and kill-by-process-group. When a
test hangs, check `ps` before theorizing.

The more revealing artifact of this chapter is the dependency graph
itself. Before the work, `go list -deps` on a minimal framework
consumer included `gorilla/websocket`, `agent/internal/ws`, and every
package they transitively required. After, the headless consumer's
closure is smaller. The GUI consumer's closure is the same size,
which is the point: nothing was removed from the GUI, only from the
framework path the GUI was sitting on.

The deletion count is modest. Four dead exports, one package rename,
one struct consolidation, and two package moves. The diff is smaller
than most feature chapters. What changed is the address of the code
in the dependency graph, and that change is the one the compiler can
verify.

## 24.8 Checks

The grader (`make grade24`) runs seven checks. Every absence
assertion is paired with a positive control that proves the detector
fires.

| # | Behavior | Check | Points |
|---|---|---|---|
| 1 | Framework links no GUI, no gorilla | `go list -deps` on a root-only consumer compiled OUTSIDE the module | 20 |
| 2 | GUI publicly reusable | external consumer + fake hooks: planted event replays to a ws client, a prompt reaches the Send hook, and the MCP relay answers JSON-RPC with the request's own id | 15 |
| 3 | Wire unchanged | ch22-style scripted session against the real binary; recorded vendor requests prove the GUI-socket model switch landed | 15 |
| 4 | Assets embedded | binary started from an unrelated cwd serves the GUI | 15 |
| 5 | Components without shell | consumer mounts the Artifact scroll from the exported FS; `index.html` 404s | 10 |
| 6 | MCP without gorilla | headless MCP consumer links no gorilla; virtual-user still dials through `mcpws` | 15 |
| 7 | Root stays clean | structural: root exports no websocket-typed symbols, imports neither `gui` nor gorilla | 10 |

The mutation audit (`make grade24-audit`) runs seven mutants. Each is
proved to land and compile before it runs, and each kills exactly its
expected check set. Mutant 4 found a grader weakness before the
chapter shipped: a reply-count assertion passed a server that
silently dropped `update_settings` frames. The recorded model
sequence from the wire is the observable that caught it. The audit
improving the grader rather than merely validating it is the
page-contract system working as intended.

One lesson from the sweep: a refactor's blast radius includes every
harness that spells the old path. The chapter 14 speech harness
loaded `renderers.js` from `web/gui` by disk path. The chapter 12
grader hunted `internal/ws/handler.go`. Neither was a chapter 24
regression; both were existing paths that the rename surfaced. The
grader-side fix tries the new path and falls back to the old, so it
grades both the current tree and frozen solutions from before the
move. The full sweep found them because it tests every chapter
against every tree, and a per-chapter grader tests only one chapter
against one tree.

The broader lesson is about what a refactor chapter teaches. A
feature chapter adds behavior; the grader checks that the behavior
works. A refactor chapter changes addresses; the grader checks that
every address the system knew about still resolves. The second check
is harder to write and easier to miss, because the symptoms appear
in code that was not touched. A rename in one package is a broken
import in every harness that spells the old path, and the harnesses
are not the refactor's diff. They are its blast radius, and only a
sweep that crosses chapter boundaries can measure it.
