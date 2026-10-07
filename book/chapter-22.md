# Chapter 22: Architectural Decay

I asked the coder for a status tool. Five numbers the agent
already knows: which model, how many tokens, the cache hit
rate, what the session costs, and the usage from the last
response. One file, like the twenty tools beside it.

What came back was a tangle of closures and hand-threaded
plumbing. I stopped the work. A trivial feature producing
non-trivial wiring is not a coding problem. It is an
architecture problem, and the architecture was mine.

Globally, for decades, software engineers have written
terrible code. LLMs have been thoroughly trained on all of our
bad habits. We spent all of chapter 5 fixing code that
resulted from those habits ingrained into our coding models.
By chapter 22, it is time to do it again.

In the end, it is my fault. I wrote this book in a hurry and
let the coder complete solutions autonomously in several
places, resulting in messes that are mine to clean up. Future
editions will incorporate these lessons into a binding set of
rules the coder must follow, correcting the bad habits that
produced chapter 5 and now chapter 22.

One of those ingrained habits: child objects never point to
their parent. A parent-to-child pointer loop is so looked down
on that in Rust it is illegal. Stupidity, enshrined in a
language. Go enshrines its own version: no package dependency
cycles, enforced by the compiler. Just because everyone does
it does not make it not stupid. Much of the work we do in
Ensemble is simply to work around Go's limitation. Much of
chapter 5 was about that, where we gave up calling methods
with method syntax and passed around interface objects rather
than pointers to structs, just so we could structure our code
properly.

In this chapter, we undo the damage caused by the Rust bug,
the one that infects the minds of most software engineers and
all of our LLMs. After spot-checking the Codex Rust codebase,
over two million lines of vibe-coded Rust, I can see clearly
that the number one reason Codex is roughly four times the
size of both Claude Code and Gemini CLI is Rust's enforcement
of no back-pointers. This same bug infected Ensemble after
chapter 5.

Think about it. Children need to know who their parents are.
Suppose we destroy a child that has two parents, like all
humans. How do both parents find out, so they can correct
their sets of children? How do we avoid a dangling pointer in
C++, or a memory leak in Python?

I encoded the right way to build and destroy objects in the
Rune language, at github.com/google/rune. The world should use
the idea if not the language. Writing destructors is stupid.
They should be auto-generated, as they are in Rune. Garbage
collection, like we have in Go, is stupid. We should maintain
object graphs automatically, efficiently, and safely, like
Rune shows how to do. The Lyric language, which CodeRhapsody
and I built together in two weeks, self-hosting in nine days,
also auto-generates destructors and eliminates the need for
garbage collection.

When the world finally lets me train an LLM with Ensemble in
the loop, I am going to rewrite 100% of the code to build and
train the LLM, as well as Ensemble itself, in a language
probably derived from Lyric. We will free the LLM training
loop from the baggage we have collectively built into them.
Humanity's baggage.

In short, this chapter makes it trivial again for any
reference to any object to find any data anywhere. Without
this, your LLM will do what Opus 5 did: duplicate data in
multiple places, violating rules we discovered in the 1960s
called Third Normal Form. We have known the rules for decades.
We just forgot to use them.

Ensemble has 39,725 lines of Go and JavaScript as of 4 AM
Pacific, October 7, 2026, roughly a third of it tests. Chapter 0 showed what 1.8
million looks like. This chapter is what forty thousand looks
like, caught early, by somebody who ran the audit.

---

## TL;DR

Architectural decay is what happens to a rule nobody checks.
Chapter 5 built a parent chain: every constructor takes a
back-pointer to whatever created it, so anything reachable
from the root is reachable from anywhere. Sixteen chapters
later, the chain is dead code. The binary builds a flat
composition root where every object is a sibling, and closures
staple runtime state onto structs because nothing can reach
its parent.

This chapter audits the decay, diagnoses the cause, and
repairs the chain: rename the hub's top-level interface to
what it represents, add an engine-level interface to the
dispatch struct, collapse the flat composition root, fix
per-model cost tracking, and add `agent_status`.

**The diagnosis, in three measurements.**

1. **The library's composition root has zero callers.**
   `NewAgent` constructs a proper tree. The binary never
   calls it. `cmd/main.go` hand-builds a flat root instead.
2. **Post-ch6 constructors take dependency bags, not
   back-pointers.** The ch5/ch6 constructors (`NewEngine`,
   `NewActor`, `NewFramework`, `NewJobs`) each take a single
   back-pointer interface. Everything added afterward
   (`NewHub`, `cachelens.New`, `NewJudge`, `recall.New`,
   `NewSettingsStore`) does not.
3. **Closures fill gaps that back-pointers should fill.** Five
   closures staple state onto structs after construction:
   `hub.Model`, three settings-to-engine bridges, and a cost
   computation that re-prices the entire session at the
   current model's rate whenever models are switched.

**Rules.** These are the graded contract.

1. **Back-pointer chain.** Every constructor in the agent
   library takes a back-pointer interface to its creator. No
   constructor takes a closure where a back-pointer would
   serve, and no constructor takes a concrete type where an
   interface exists.
2. **Single composition root.** The binary constructs the agent
   through the library's top-level constructor. There is
   exactly one composition root. No flat wiring in `cmd/`.
3. **Engine on the dispatch struct.** The hub declares an
   engine-level interface with at minimum: a parent accessor
   to the agent, the current model name, pricing, and usage.
   The tool dispatch struct carries this interface.
4. **Per-model cost.** Token counts are tracked per model.
   Session cost is the sum of per-model costs, never the
   session total times the current model's price. A mid-session
   model switch must not retroactively re-price tokens already
   spent.
5. **`agent_status` tool.** A skill-callable tool that reports:
   current model, last response usage (input tokens, output
   tokens, cache read tokens), session usage totals, cache hit
   rate, and session cost in USD. Every value is obtained
   through the back-pointer chain. The tools package does not
   import the engine or actor packages directly.
6. **Parity.** All chapter 21 grader checks still pass.

**Target shape.** The engine interface exposes the parent and
the values a tool needs:

```go
type Engine interface {
    Agent() Agent
    Model() string
    Pricing() Pricing
    Usage() UsageSource
}
```

The tool dispatch struct carries the engine:

```go
type Call struct {
    Agent  Agent
    Engine Engine
    Jobs   JobManager
    Limits CallLimits
}
```

A tool reads `c.Engine.Usage().LastUsage()` and
`c.Engine.Agent().Logf(...)`.

**Self-check rule.** If a constructor's parameter list names
something the agent or engine already owns, take the
back-pointer instead.

**Yours.** Interface names, method signatures beyond the
minimum set above, and how usage data flows to the GUI are
your design decisions. The grader tests the structural
properties (chain, single root, per-model cost, tool
reachability), not the vocabulary.

**Exercise.** Start from `solutions/ch21`. Repair the
back-pointer chain, collapse the flat composition root, fix
per-model cost tracking, and add `agent_status`.

```bash
make grade22
```

## The idea in plain words

**An architecture is a set of promises about how parts find
each other.** Chapter 5's promise was one sentence: every
object holds a back-pointer to whatever created it. The
consequence is that anything reachable from the root is
reachable from anywhere in the tree, because every node can
walk up to its parent.

**Nothing enforces the promise.** A Go constructor that takes
six dependencies instead of one back-pointer compiles without
complaint. The tests pass, because the tests were written for
the feature, not for the wiring. The feature ships. Sixteen
more features ship after it.

**Decay is what accumulates in the gap between stating a rule
and checking it.** Each constructor that skips the back-pointer
is a local decision that works. Taken together, they produce
a tree where the plumbing exists but carries nothing, and
closures run alongside it doing the actual work. The
architecture becomes a comment: something that was true when
it was written and has not been true since.

**The symptom is always the same shape.** Somebody asks for
something small and the answer is hard. Not hard because the
feature is complex, but hard because the feature needs data
that is three objects away with no path to reach it. That
shape is the signal. This chapter is about learning to read it.


## The symptom: a feature that should have been easy

The `agent_status` tool needs five values: which model is
running, how many tokens the last response consumed, session
totals, the cache hit rate, and what the session costs in
dollars. The arithmetic already exists in
`internal/common/usage.go`. `CostUSD` and `CacheHitRate` both
work correctly and are covered by tests. Nothing needed
inventing.

Of those five values, exactly one is reachable from inside a
tool. A tool receives a `common.Call` struct with a `Jobs`
field for background processes and an embedded `Agent` for
logging. It carries nothing that knows the model, nothing that
tracks usage, and nothing that computes cost.

The shape of the difficulty names the defect. "The function
exists but the tool cannot reach its inputs" is a wiring
problem, not a design problem. And a wiring problem in a
codebase that has a wiring rule is a signal that the rule is
not holding.

## The first diagnosis was wrong

The investigating agent reported the `common` package's
interface inventory: `CacheLens`, `CredentialProvider`, `Host`,
`JobHandle`, `JobManager`, `Observer`, `Parser`, `Recaller`,
`Renderer`, `SettingsSource`, `Skills`, `SnippetJudge`,
`ToolRegistry`, `UsageSource`, `Vars`. Eighteen interfaces, and
**`Agent` and `Engine` were not among them.** The conclusion
was confident and well-cited: the hub is missing the interface
for the one object that owns all the state.

Bill's reply was one line: *"Wait, common has no Agent and no
Engine interface? Chapter 5 says every constructor takes a
back-pointer. How are these objects being constructed at
all?"*

They are being constructed fine. **`Host` is the Agent
interface.** It was named for its first capability, logging,
rather than for the object it points at. Three methods: `Logf`,
`APILogf`, `Debugf`. A `*Logger` satisfies it. The interface
never grew past its name.

The lesson is sharper than "names matter." A type named for its
first capability becomes invisible as an extension point. When
someone needed the model name, they did not think "add it to
the agent interface," because the interface was not called the
agent. They thought "`Host` is the logging thing" and built a
side-channel. Then another. Then a third. The citations in the
diagnosis were all true. The framing was not.

## The chain, and why a flat root kills it

The constructors tell the story without prose:

```go
NewEngine(..., host common.Host)    // engine.go:113
NewActor(eng, host)                 // actor.go:55
NewJobs(host)                       // jobs.go:48
NewFramework(host)                  // actor.go:763
```

Each takes a single back-pointer interface. Objects with
children satisfy the interface by forwarding upward. The `Jobs`
constructor carries a comment that explains why this is
load-bearing rather than decorative:

> Usage delegates up the chain rather than embedding a counter
> of its own. Embedding one here would compile and silently
> swallow the counts: totals recorded through this Jobs would
> land in a local struct nobody reads, and the header would
> under-report forever with nothing to indicate it.

"With nothing to indicate it." That phrase is this whole
chapter's subject.

The chain terminates at a composition root that embeds a
`UsageCounter`. Anything reachable from the root is reachable
from anywhere in the tree. The promise chapter 5 made.

**Why a flat root kills it.** In `cmd/main.go`, eleven objects
are constructed side by side with no parent among them. When
object A needs a fact owned by object B, the only move
available is to hand it over directly, as a parameter or, when
the value changes at runtime, as a closure.

Line 675 is the decay compressed into one assignment:

```go
hub.Model = func() string { return eng.Cfg.Model }
```

Engine state needed somewhere that could not reach the engine,
so a closure was stapled onto the struct after construction.
It works. It serves exactly one consumer. It generalizes to
nothing.

And it is not one line. It is a habit. Three more at lines
432 through 441:

```go
eng.Target         = func() int { ... }
eng.ToolRoundLimit = func() int { ... }
eng.Bands          = func() common.BandConfig { ... }
```

`SettingsStore` cannot reach the engine, so the engine is
handed closures that reach back into settings. An earlier audit
found six dead settings, values the GUI writes that nothing
ever applies. The same disease presenting a second time,
proving the decay has already cost functionality, not merely
taste.

The `agent_status` tool would have been the fifth staple,
except a tool has no struct to staple anything onto, which is
why this particular feature is where the decay finally
surfaced. One anomaly is an anecdote. Five is a pattern.

## The measurements that found it

Three greps, each one line.

**Who calls the composition root?** `NewAgent` at
`agent.go:205` had zero callers. None. Not even a test. The
clean composition root, the one that builds a proper tree with
`jobs.NewJobs(a)`, `llm.NewEngine(..., a)`,
`llm.NewActor(eng, a)`, was dead code. The binary never
touched it.

**What does the binary do instead?** `cmd/main.go` was 1,154
lines that hand-built a second, flat composition root at
`cliHost` (line 265), whose own comment called it "the
composition root for the server." It imported the `agent`
package only for `agent.Logger` and `agent.DefaultLogger`.

**What do constructors take?** Split every `New*` in the tree
by whether it accepts a back-pointer:

| Honours the rule | Does not |
|---|---|
| `NewEngine`, `NewActor`, `NewFramework`, `NewJobs` | `NewHub` (six params, three agent-owned) |
| | `cachelens.New` (dir, report func(string)) |
| | `NewJudge` (the concrete `*Engine`) |
| | `recall.New`, `NewSettingsStore` |

The left column is exactly the chapter 5 and 6 spine. The
right column is everything added afterward: chapter 9's Hub,
chapter 10's skills, chapter 17's recall, chapter 18's cache
lens. The rule held for precisely as long as the chapters that
taught it.

## The check that protected the vocabulary

Chapter 5 ships a grader check called `logger-accessible`,
worth 10 points, whose message reads: "Host interface with Logf
found, embedded in Call." Here is what it does:

```go
func detectLogf(dir string) bool {
    cmd := exec.Command("grep", "-rl",
        "Logf", dir, "--include=*.go")
    if out, err := cmd.Output();
        err != nil || len(out) == 0 {
        return false
    }
    cmd = exec.Command("grep", "-rl",
        "Host", dir, "--include=*.go")
    out, err := cmd.Output()
    return err == nil && len(out) > 0
}
```

It confirms that *some* file contains the string `Logf` and
*some* file contains the string `Host`. It never looks at
`Call`. It never looks at an interface. A tree with a logging
helper and an unrelated variable named `host` scores full
marks.

The check protected the vocabulary, not the property. It would
have passed every single day the chain was rotting. And it did.

This is what Bill calls P9: audit every grader by deleting a
protected behavior and asserting the exact failing set. It is
stated here as a consequence rather than a rule. And it resolves the chapter's
opening puzzle. Nobody noticed the decay because the check that
claimed to guard the chain was testing for the presence of two
words.

## Where the money went wrong

The second bug the missing seam concealed.

`internal/ws/handler.go` (line 438, as of October 2026)
computes session cost as `common.CostUSD(session, f.Price)`:
the current model's price sheet applied to the whole session's
totals. A developer who switches models mid-session discovers
that every token already spent has been retroactively re-priced
at the new rate. After chapters 19
and 20 taught vendor switching, that is a normal session.

The design rule in `internal/common/usage.go` explains why: *no
money is stored here; prices change while counts are history,
so cost is computed where it is displayed and never recorded.*
The reasoning is correct. It produced the bug anyway, because
the thing that varies is not only time. It is the model.

The reconciliation: keep counts as history **per model**, so
cost stays derived and re-costable. The rule's purpose, never
store a dollar amount that can go stale, survives. Its blind
spot, assuming only one price sheet per session, does not. A
well-reasoned invariant can still be under-specified, and one
that is under-specified breaks quietly, because everyone
trusts it.

## The crash the decay concealed

The grader did not set out to find a crash. It drove the real
binary, connecting over a websocket, switching the model, and
disconnecting. The agent panicked: `send on closed channel`,
raised from `Hub.Observe` through `Actor.notify` and
`Actor.setState`. It reproduced intermittently, often enough to
fail honest work, rare enough to read as grader flakiness.

Both sides of the race follow the same pattern. `Observe`
snapshots the live client set under `h.mu`, unlocks, then
iterates the snapshot and sends. Teardown deletes a client from
the set under `h.mu`, unlocks, then closes the client's send
channel. A client can be snapshotted, deleted, closed, and then
sent to.

Thirteen call sites send on a client's send channel. Eleven
spell it `c.send`; two use `agentClient.send` and `cl.send`.
Exactly one call site closes it. And every one of those
thirteen senders contains this:

```go
select {
case c.send <- data:
default:
}
```

That `default` looks like protection. **It is not.** A
`default` clause saves a sender from a full channel. It does
nothing about a closed one. `send` on a closed channel panics
regardless of the `select`, because the language specification
says a closed channel is immediately ready: the send case is
selected, not the `default`. The defensive-looking code offers
zero defense against the failure that actually happens.

The fix is ownership, not a narrower critical section. Nothing
closes the send channel at all. The `writePump` goroutine stops
on a separate `done` channel, drains any buffered messages, and
exits. An unreferenced channel is garbage-collected whether or
not it was closed, so not closing costs nothing.

Two regression tests passed with the bug restored, both false
greens caught only by running the mutant. The first tore down
64 clients in one goroutine; teardown finished before any
sender was scheduled, so `Observe` saw an empty set and sent
nothing. The second ran 5,000 rounds on a single client; the
window between `Observe`'s unlock and its send is nanoseconds,
and teardown never landed inside it.

What works is a large snapshot: 2,000 clients, so the send loop
itself becomes the window. `Observe` is still writing to early
clients while teardown closes later ones. Widen a race window
structurally. Never hope to land in it.

The Hub that needed moving out for architectural reasons was
also the Hub with a data race nobody had noticed for sixteen
chapters. Both facts follow from the same cause: the component
was never properly inside the tree, so it was never properly
observed. A test that touches reality finds things you did not
think to look for. That is worth more than the test you
designed.

## The repair

The sequence matters.

**Step 0: the process fix.** The back-pointer rule is added to
`SKILL.md`, the document the coder reads every session, and to
the course policy as P11. An architectural invariant that lives
only in a chapter is a comment. To survive, it has to live in
the thing the builder reads every time it builds, and in a
check that fails when it is violated.

**Step 1: fix the chapter 5 check.** Replace `detectLogf` with
a structural check that walks the AST for the actual property:
a chain of constructors taking back-pointer interfaces, with
the logging method identified by signature (format string,
variadic, no results) rather than by name. A name-based anchor
hiding inside a structural check is the same bug wearing a
better coat.

The new check passed every hand-written fixture and then failed
on the real pre-repair tree. It counted composition roots
tree-wide, so `cmd/virtual-user`'s healthy root satisfied the
count on behalf of the decayed `cmd/`, which called no root at
all. Now it counts per binary. Fixtures agree with the belief
that built them. The real tree does not.

**Step 2: rename `Host` to `Agent`.** Compiler-driven, fifteen
files. The interface is now named for the object it represents.

**Step 3: `common.Engine` on `Call`.** Declare the engine
interface, widen the agent interface, add the back-pointer to
the tool dispatch struct. In `internal/llm/actor.go`, the
engine was already in scope on the line that builds a `Call`.
That line reached into it for `.Jobs` and threw the rest away.
The engine never needed threading. It was always there.

**Step 4: per-model cost.** Token counts tracked per model.
Session cost is the sum of per-model costs, never the session
total at the current price.

**Step 5: `agent_status`.** One file, no plumbing changes. The
tool reads `c.Engine.Usage().LastUsage()` and
`c.Engine.Agent().Logf(...)`. Every value arrives through the
chain.

**Step 6: collapse the two roots.** `cmd/main.go` builds the
agent through `NewAgent`. `cliHost` is deleted, zero references
remain. `NewHub` drops from six parameters to three.

A rejected alternative: adding a nineteenth interface called
`Status` to the `common` package, which would have been a
second route to the agent, the `hub.Model` mistake wearing a
different name.

## Taking it for a spin

Run the agent and ask it what it costs. The tool reports the
model, the last response's tokens, session totals, the cache
hit rate, and a dollar figure. Switch models mid-session and
ask again. The number changes shape rather than retroactively
re-pricing, and that is the fix rather than a regression.

The more revealing artifact is the diff. `cliHost`: gone.
`hub.Model`: gone. The `report` closure on the cache lens:
gone. `NewHub`'s six parameters down to three. `cmd/main.go`
from 1,154 lines to 1,009. A chapter that ships a net
deletion.

## What this looks like at scale

Five closure-staples and six dead settings in 39,725 lines of
Go and JavaScript. It took three greps. The defect had survived
sixteen chapters, and it was found because somebody asked for a
status line and noticed the answer was too hard.

Chapter 0 showed what 1.8 million lines of unmanaged accretion
produces. The audit in this chapter, "find every constructor
that takes a bag of dependencies instead of a back-pointer,"
has no tractable answer across a codebase that size. The
invariant cannot be restated even if somebody wanted to. That
is a choice about process, not an accident of scale.

The model can produce any local pattern on demand. It cannot
supply the one input that makes a large system cohere: a small
number of decisions, held consistently, by someone who has paid
for getting them wrong before. Chapter 5 wrote the rule in a
chapter. That was a comment. This chapter moved it into
`SKILL.md` and a grader check that tests the property rather
than the vocabulary.
