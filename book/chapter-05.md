# Chapter 5: The Big Refactor

Every major AI coding agent, from Claude Code to Cursor, was built as
a coding agent first and asked to do other things later. When SpaceX
valued Cursor at $60 billion, the price was not for a code editor. It
was for the harness: the tool loop, the vendor seam, the job
supervision, the machinery that makes an LLM do real work in the
world. A coding agent is the most valuable thing that harness can do
today. It is not the only thing.

Your agent is 5,047 lines of Go in one directory. It reads files,
writes code, drives a debugger. It is also a monolith that nobody
else can use. This chapter turns it into a general-purpose agent
framework by reorganizing the code into packages with clean import
direction, at a cost of 123 lines and zero new features. At the end,
a twenty-line program imports your framework, registers a custom tool
the framework has never seen, and runs an agent that calls it. Your
agent is no longer just a coding agent. It is a platform.

**What you build.** The same agent, reorganized into packages with
clean import direction. Two rules govern the split:

1. Constructors take interfaces to their parents.
2. Those interfaces live in a shared vocabulary package that every
   other package imports.

The result is a star topology: a hub of shared types at the center, and
implementation packages that each import the hub and nothing else.
No implementation package imports another. The root of the module is
the public API, and a separate program can import it, register a custom
tool, and run an agent without touching any internal code.

```
agent/
  go.mod                       module root, own go.mod
  agent.go                     package agent: the public API

  internal/
    common/                    shared types + interfaces (the hub)
      event.go                   Event, EventType, TextPart, ToolCallPart, ...
      part.go                    Part, BlobPart, Ref, ...
      context.go                 Context, Entry, Usage, ...
      provenance.go              Provenance, Actor, ...
      config.go                  Config, ToolDecl, Vendor, Surface, ...
      interfaces.go              JobHandle, JobManager, ToolRegistry, ...
      event_log.go               Log (event log reader)

    llm/                       engine + vendor implementations
      engine.go                  Engine, NewEngine, Execute, Shutdown
      seam.go                    SeamFor, vendor routing, render/parse helpers
      claude.go                  Anthropic renderer and parser
      gemini.go                  Gemini renderer and parser
      openai.go                  OpenAI renderer and parser

    jobs/                      job lifecycle
      jobs.go                    Job, Jobs (implements JobHandle, JobManager)

    tools/                     tool implementations + registry
      tools.go                   Registry, builtin tools, Declarations, Lookup
      jobtools.go                wait_for_job, send_input, kill_job, tool_limits

  cmd/
    main.go                    package main: the CLI (chat, render, dump)
```

The directory structure above is the reference solution's. Yours will
differ. What the grader checks is the topology, not the names: no
implementation package imports a sibling, and every implementation
package reaches shared types through a common hub.

The interfaces that break the dependency loops:

```go
// JobHandle is what a tool sees of its own job.
type JobHandle interface {
    io.Writer
    Attach(proc *os.Process, stdin interface{ Write([]byte) (int, error) })
    SetExit(code int)
    SetCwd(dir string)
    Status() JobStatus
    Data() *JobData
    HasProcess() bool
    Wait(l Limits) WakeReason
    Report(reason WakeReason, l Limits) string
    Kill(reason string) bool
    SendInput(text string) error
    Bytes() int
    Err() error
    Finish(result string, err error)
}

// JobManager manages the set of running jobs.
type JobManager interface {
    Start(tool, callID string) (JobHandle, error)
    Get(h int) (JobHandle, bool)
    Handles() []int
    Running() []JobHandle
    SetNext(l Limits)
    Take(args json.RawMessage) (Limits, bool, error)
}

// ToolRegistry is what the engine uses to dispatch tool calls.
type ToolRegistry interface {
    Lookup(name string) (Tool, error)
    Declarations() []ToolDecl
}
```

These are the reference solution's interfaces. Yours will be shaped by
where your code draws its package boundaries. What they share is the
principle: each interface declares exactly the methods one package needs
from another, and they live in the hub where every package can see them.

The public API that external programs see:

```go
package agent

type Config    = common.Config
type ToolDecl  = common.ToolDecl
type Vendor    = common.Vendor
type Usage     = common.Usage

func RegisterTool(name, desc string, schema json.RawMessage,
    handler func(json.RawMessage) (string, error))
func NewAgent(cfg Config, logPath string) *Agent
func ConfigFromEnv() (Config, error)
func (a *Agent) Ask(prompt string) (string, error)
func (a *Agent) Shutdown() error
func (a *Agent) Usage() Usage
```

Type aliases re-export the internal types so external callers never
import `internal/` directly. `ConfigFromEnv` reads the standard
environment variables (`LLM_VENDOR`, `LLM_MODEL`, `LLM_API_KEY`,
`LLM_BASE_URL`). `RegisterTool` adds a custom tool to the global
registry before `NewAgent` wires everything together.

**Rules.** Five checks. The grader builds both binaries, inspects
the dependency graph, runs the exercise against the fake vendor, and
reruns the Chapter 4 grader against the refactored agent.

1. **The agent builds from `cmd/`.** `go build -o bin ./cmd/` in the
   `agent/` directory produces the CLI binary. (`agent-builds`, 10)
2. **The exercise builds separately.** A program outside `agent/` that
   imports the framework compiles without error. (`exercise-builds`, 10)
3. **Star topology.** Every implementation package under `internal/`
   imports only the shared hub and the standard library. No
   implementation package imports a sibling. (`star-topology`, 25)
4. **A custom tool is called.** The exercise registers a tool, sends a
   prompt that triggers it, and the tool's output appears in the
   response sent back to the vendor. (`custom-tool-called`, 25)
5. **Chapter 4 still passes.** Every Chapter 4 check runs against the
   refactored agent binary and scores 100/100. The refactoring changed
   nothing observable. (`ch4-parity`, 30)

Five checks, sum 100: `agent-builds` 10, `exercise-builds` 10,
`star-topology` 25, `custom-tool-called` 25, `ch4-parity` 30.

**Yours.** The package names, the file layout, the exact set of
interfaces. The grader does not check names. It checks that the
dependency graph is a star, that a separate program can import and
extend the framework, and that the agent still does everything it did
in Chapter 4.

**Exercise.** Build a separate program in `ch05/` that imports your
framework, registers one custom tool, and uses it to answer a prompt.
The tool can do anything: arithmetic, string manipulation, a lookup.
The grader checks that the tool was called and its output reached the
vendor.

```sh
make grade-dir CH=5 DIR=path/to/yours
make grade5
```

## 5.1 The idea in plain words

A flat package is a room with no walls. Every function can call every
other function, every type can reference every other type, and the
compiler cannot tell you when something reaches across a boundary it
should not cross, because there are no boundaries to cross.

That is fine when the room is small. Chapter 4's agent is not small.
It is 5,047 lines of Go in fourteen files, and the engine, the vendor
parsers, the job manager, and the tool implementations all live in the
same package namespace. A change to the job manager's internal
bookkeeping can accidentally call a tool function, and the compiler
will not object.

Two rules fix this without adding machinery.

**Rule 1: Constructors take interfaces to their parents.** When
you create an engine, hand it an interface to the job manager and an
interface to the tool registry, not the concrete types. The engine
sees only the methods it needs. The job manager sees only the methods
it needs. Neither can reach into the other's internals, because the
interface is the boundary, and the interface declares only what the
caller is entitled to use.

**Rule 2: Those interfaces live in a shared package.** If the engine's
interface to the job manager lives in the engine package, the job
manager has to import the engine to implement it, and the engine has to
import the job manager to use it, and that is a cycle. Move the
interface to a package both can import and the cycle breaks. That
package is the hub. In the reference solution it is called
`internal/common`. Everything in it is vocabulary: types, constants,
interfaces. No logic, no state, no goroutines.

The result is a star. Common is the hub. Every spoke (llm, jobs, tools)
imports only the hub. The root package wires the spokes together. A
spoke that needs something from another spoke gets it through an
interface in the hub, passed to its constructor. If you draw the
dependency arrows, every arrow points inward, toward common. None
points sideways.

This is not a Go idiom. It is a dependency management principle that
Go happens to enforce at compile time, and every language with a module
system can express it. The star topology makes the next five chapters
possible: each one adds a spoke, and the hub grows by an interface or
two, and no existing spoke changes.

## 5.2 What moves where

The mechanical process is straightforward once the packages are named.
Each section below is a package, and the content is what the reference
solution put there. Yours may split differently. The property that
matters is the star: every package imports only common, and common
imports nothing inside the module.

**internal/common** is the hub: 1,455 lines, seven files. Every type
that appears in more than one package lives here. Events, parts,
context, provenance, config, the event log reader, and the three
interfaces that break the dependency loops. This is the largest package
because the vocabulary is large, and that is correct. The vocabulary
is what every package agrees on. It should be large.

**internal/llm** is the engine and the three vendor implementations:
1,516 lines, five files. `Engine` takes a `JobManager` and a
`ToolRegistry` through its constructor, so it never names the jobs or
tools package. The vendor parsers and renderers share helpers in the
same package (classify, SeamFor, the render-and-parse cycle). This
package has the most complex logic and the most files, because the
vendor seam from Chapter 2 lives here alongside the dispatch loop
from Chapters 3 and 4.

**internal/jobs** is the job lifecycle: 424 lines, one file. `Job`
and `Jobs` implement the `JobHandle` and `JobManager` interfaces
from common. The PTY setup, the output buffer, the wait loop, the
kill signal: all here, and nothing else.

**internal/tools** is the tool implementations and the registry: 882
lines, two files. The six tools from Chapter 3 and the four from
Chapter 4 live in `tools.go`. The supervision verbs (wait, send, kill,
tool_limits) live in `jobtools.go`. The split is natural: tools that
become jobs and tools that act on jobs. The registry maps names to
tools and provides `Lookup` and `Declarations` for the engine.

**agent.go** at the module root is the public API: 134 lines, the
thinnest file in the module. Type aliases re-export the internal types,
`NewAgent` wires the concrete implementations together, and
`RegisterTool` adds entries to the global registry. This is the only
file an external program needs to know about.

**cmd/main.go** is the CLI: 277 lines. It imports every internal
package to wire them together, which is the one place that breaks the
star. The CLI is the composer, not a spoke. It creates the jobs, the
registry, and the engine, and hands each the interfaces the others
declared.

The sum is 4,688 non-test lines. Chapter 4 was 4,565. The difference
is 123 lines: the interface definitions in common and the public API
wrapper in agent.go. A refactoring that grew the codebase by less
than three percent is a refactoring that added structure, not weight.

## 5.3 The star, enforced

The topology is a property you can check mechanically. `go list -deps`
on any package prints every package it transitively imports, and the
rule is that no implementation package appears in another
implementation package's list. A script that does this for every
directory under `internal/` takes four lines and catches a sideways
import the moment it happens, which is before the module compiles
because Go enforces the rule that a cycle is a build error.

Go's `internal/` convention adds a second enforcement. A package under
`internal/` can be imported only by code rooted at the parent of
`internal/`. An external program that tries to import
`agent/internal/llm` gets a compile error naming the restriction. This
is the wall, and it has two consequences.

The first is protection: the engine, the vendor parsers, the job
lifecycle, and the tool implementations are not part of the public API.
An external program sees only what `agent.go` re-exports. The type
aliases (`type Config = common.Config`) are the windows in the wall,
deliberately placed and deliberately few.

The second is discovery. When you draw the wall for the first time
around code that was never behind one, encapsulation violations that
compiled without complaint become errors. A tool function that reached
into the engine's internal state, a test file that constructed a Job
directly instead of going through the manager, a helper that imported
a package it had no business knowing about: all of these compiled in
the flat package and break the moment the wall goes up. The compiler
is doing the code review.

This is the argument for doing the refactoring as a separate chapter
rather than folding it into whatever comes next. The wall finds
problems that exist in the code you already have. Adding a feature at
the same time as drawing the wall means you cannot tell whether a
compile error is a boundary violation you should fix or a consequence
of the feature you are adding, and the temptation is to fix both by
removing the wall.

## 5.4 The public API

An agent framework that can only be used through its own CLI is a
toy. The point of the refactoring is that a separate program can
import `agent`, register a tool the framework has never seen, and
run a session. The public API is small enough to list completely:

```go
agent.RegisterTool(name, desc, schema, handler)   // before NewAgent
agent.NewAgent(cfg, logPath)                       // wires everything
agent.ConfigFromEnv()                              // reads LLM_* vars
agent.Ask(prompt)                                  // one turn of the loop
agent.Shutdown()                                   // kills jobs, saves log
agent.Usage()                                      // token counts
```

`ConfigFromEnv` exists because an external program should not have to
know the vendor environment variable names. The framework reads
`LLM_VENDOR`, `LLM_MODEL`, `LLM_API_KEY`, and `LLM_BASE_URL`, picks
defaults per vendor, and hands back a `Config`. A program that needs
different defaults builds its own `Config` and skips `ConfigFromEnv`
entirely.

`RegisterTool` adds a tool to the global registry before `NewAgent`
assembles the pieces. The handler signature is simpler than the
internal `ToolFunc` because external tools do not need the `Call`
context. They receive JSON arguments and return text:

```go
agent.RegisterTool("calculate", "Evaluate an arithmetic expression",
    json.RawMessage(`{
        "type": "object",
        "properties": {
            "expression": {"type": "string", "description": "e.g. 6 * 7"}
        },
        "required": ["expression"]
    }`),
    func(args json.RawMessage) (string, error) {
        var p struct{ Expression string `json:"expression"` }
        if err := json.Unmarshal(args, &p); err != nil {
            return "", err
        }
        return fmt.Sprintf("Result: %s = 42", p.Expression), nil
    },
)
```

That tool is the exercise's proof of concept, and it is the simplest
tool that proves the point. The framework calls it, the vendor sees
the result, and nothing inside `internal/` needed to change.

## 5.5 A rename's blast radius

A lesson from the reference solution, confessed because the cost was
a failing test and a twenty-minute debugging session.

The flat package had a type called `Call`. When it moved to
`internal/common`, every reference became `common.Call`. A
search-and-replace across the package was the obvious move, and
the obvious move corrupted a string that had nothing to do with
the type:

```
Before:  "Call the tool you meant to call"
After:   "common.Call the tool you meant to call"
```

`Call` is both a Go identifier and an English word. The compiler
sees no difference between a type reference in code and a word in
a string literal. Neither does `sed`, which is what ran the
replacement. The test that caught it was a string comparison in
a grader fixture, and it failed with a message about a tool description
containing `common.Call`, which is not a phrase any model should
see.

The rule from the experience: after every rename, grep for the
identifier in all quoted strings and all comments. The compiler
verifies code references. It does not verify prose. A rename's
blast radius is where the word appears, not where the identifier
is used, and the two sets overlap in a language that names things
with English words.

This is not a hypothetical. The reference solution's test caught
it because the grader compares exact strings. A student whose
tests do not compare strings will ship the corrupted description
to the model, the model will try to parse `common.Call` as a
function name, and the failure will look like a model hallucination
rather than a botched rename.

## 5.6 The exercise, graded

The exercise is a separate program that imports the framework, and
the grader checks that it works end to end.

Build a program in `ch05/` that:

1. Calls `agent.RegisterTool` with one custom tool. The tool can
   do anything. The reference solution's tool evaluates arithmetic
   expressions and returns a string like `"Result: 6 * 7 = 42"`.
2. Calls `agent.ConfigFromEnv` to read the vendor configuration.
3. Calls `agent.NewAgent` with the config.
4. Calls `agent.Ask` with a prompt designed to trigger the tool.
5. Calls `agent.Shutdown`.

The grader runs this program against the fake vendor. The fake serves
a tool call that names the custom tool, waits for the tool result in
the next request, and serves a final text reply. The check passes when
the tool's output string appears in the body of the second request.

```sh
make grade-dir CH=5 DIR=path/to/yours
make grade5
```

| check | points | passes when |
|---|---|---|
| `agent-builds` | 10 | `go build -o bin ./cmd/` succeeds in `agent/` |
| `exercise-builds` | 10 | `go build -o bin .` succeeds in `ch05/` |
| `star-topology` | 25 | no internal/* package imports a sibling |
| `custom-tool-called` | 25 | custom tool result appears in the second request |
| `ch4-parity` | 30 | all Chapter 4 checks score 100/100 |

Weights sum to 100. `ch4-parity` is 30 because the refactoring must
not break anything, and "anything" means every check from every
previous chapter, weighted by how much the reorganization could
plausibly disturb it. `star-topology` is 25 because that is the
chapter's thesis: the dependency graph is a star, enforced by the
compiler. `custom-tool-called` is 25 because a framework that cannot
be extended by external code is not a framework.

## 5.7 What this chapter does not do

The agent is still deaf while running a tool. The engine still
blocks on every tool call. There is no mailbox, no inbound queue,
no way for a prompt to arrive while the loop is busy. The agent
cannot supervise multiple tasks, and it cannot be interrupted.

All of that is buildable now because the star topology means adding
an actor loop is adding a spoke. The engine
gains an inbound queue. The framework gains an observer. The hub grows
by two interfaces. No existing spoke changes, because no existing
spoke knows about actors, and the interfaces it uses did not move.

That is what the refactoring bought: the ability to add machinery
without disturbing machinery that already works. A flat package
cannot make that promise. Every addition to a flat package can touch
everything, and "can" becomes "does" the moment a deadline arrives.

## 5.8 Drive it yourself

Build the exercise and run it:

```sh
cd ch05
go build -o ch05agent .
LLM_VENDOR=anthropic LLM_API_KEY=sk-... ./ch05agent
```

The custom tool registers, the agent runs, and the model calls the
tool you declared. The output on the terminal is the same conversational
loop from Chapters 3 and 4, and the code that produced it is a
twenty-line program that imported a framework.

Then try what the exercise actually proves. Go to a project you care
about, write a small Go program that imports your agent framework,
and give it a tool that knows something about your project. A tool
that queries your database. A tool that checks your CI status. A tool
that reads your monitoring dashboard. The framework does not know about
any of these, and it does not need to.

That is the payoff of a refactoring chapter. The code does the same
thing. The codebase does not.


## 5.9 The ambush

How do you debug your agent?

The engine calls tools, the tools call the shell, the shell runs
builds that fail and tests that flap. When something goes wrong, the
only evidence is the model's next response, which is the model's
interpretation of the evidence, not the evidence itself. There is no
logger. There has never been a logger. 5,047 lines of Go, and not one
of them writes a debug message anywhere.

This is not an accident. It is a set-up.

A logger is a facility that every piece of code in the system needs
access to. In a flat package, the solution is a global variable. In a
system that will eventually run multiple agents in one process, a
global variable is a collision waiting to happen. One agent's debug
output interleaved with another's is worse than no debug output at all.

The logger belongs on the top-level object: the Agent. Each agent
gets its own logger. Internal code reaches it through the parent chain
you just built. If the refactoring in §5.2 produced a star where every
constructor takes an interface to its parent, adding the logger is
four changes:

1. Define a `Host` interface in `internal/common` with one method:
   `Logf(format string, args ...any)`.
2. Embed `Host` in the `Call` struct. Every tool function already
   receives a `*Call`, so every tool can now log.
3. Create a `Logger` type in the root package: a mutex-guarded
   writer with timestamps. Put it on the Agent. Make `Agent.Logf`
   delegate to the logger.
4. Thread the `Host` into every constructor that needs it: the engine,
   the job manager, the tool registry. Each one stores the host and
   implements `Logf` by calling `host.Logf`.

That is 72 lines. The parent chain carries the logger from the top of
the tree to every leaf, and no leaf needs to know how the logger
works. It only needs to know that `Logf` exists on its parent interface.

## 5.10 Kill the globals

While you are threading the Host, audit your package-level `var`
declarations. The only globals that should survive are immutable
lookup tables: the maps that convert an enum to a string. Everything
else moves to a struct.

The tool registry is the most important target. In the ch04 solution,
the registry was a package-level map. In the refactored code, it
becomes a `Reg` struct created by `NewRegistry()`, held by the Agent,
and passed to the engine as a `ToolRegistry` interface. This is not
cosmetic. When sub-agents arrive, each one will load skills that
declare tools. A per-agent registry means one agent's skill tools do
not leak into another agent's capability surface. A global map cannot
make that guarantee.

The grader now checks for both:

| Check | Points | What it verifies |
|-------|-------:|------------------|
| `logger-accessible` | 10 | `Logf` declared in common, embedded in `Call` |
| `no-mutable-globals` | 10 | zero mutable `var` declarations outside tests |

If the star topology holds and every constructor takes a parent
interface, both checks pass on the first try. If the refactoring
skipped the interfaces, if constructors still reach up into package-
level variables for their dependencies, the logger has nowhere to
live and the globals have nowhere to go.

That is the ambush. The refactoring was never about moving files into
directories. It was about building the parent chain that makes
everything after this chapter possible.
