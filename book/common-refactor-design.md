# internal/common: diagnosis and refactor

Status as of commit `cc44df4`. All measurements taken from the tree, not from memory.

## The complaint

`internal/common` looked like a dumping ground: the place coder agents put a file
when they could not decide which spoke owned it.

## The measurement

Non-test lines, before any change:

| package | non-test lines | files |
|---|---|---|
| llm | 5,092 | 19 |
| **common** | **4,917** | **24** |
| tools | 3,166 | 9 |
| ws | 1,268 | 5 |
| mcp | 1,146 | 6 |
| recall | 604 | 3 |
| jobs | 412 | 3 |

Two numbers carry the diagnosis:

1. **51% of the hub was function bodies** (2,512 of 4,917 lines). A vocabulary
   package should be mostly declarations.
2. **11 of 25 hub files had exactly one spoke consumer**, totalling 2,145 lines
   (44% of the hub). A type used by one spoke is not shared vocabulary.

## What was NOT wrong

The import graph was already correct and still is. Every spoke imports only
`common`; no spoke imports a spoke; `common` imports no first-party package.
The star topology was never violated in the formal sense that the ch6
`star-topology` and `hub-clean` checks measure.

The problem was subtler than a bad edge: **the shape was right, but the hub was
doing work.**

## The mechanism (the real finding)

Go forbids defining a method on a type declared in another package. So the
moment a type lives in the hub, every operation on it must also live in the hub,
as a method, or be rewritten as a free function.

`common/segment.go` is the pure case: 203 lines of compaction behavior used by
nothing but `llm`, pinned to the hub because they are methods on `common.Context`
(`Segments`, `SelectForCompression`, `splitSegment`, `endsWithCheckpoint`).
Attempting the move produces:

```
cannot define new methods on non-local type common.Context
```

This is why hubs in Go accrete. It is not laziness. It is that placing the
central data type in the hub silently makes the hub the only legal home for
every operation on it. Escaping requires converting methods to free functions
and updating call sites, which is a design change, not a file move.

## What was done: skills becomes a spoke

The skill system was the clearest offender: 696 lines of frontmatter parser,
dependency resolver with cycle detection, and `$VAR` renderer, sitting in the
hub and consumed by exactly one spoke.

The seam turned out to be tiny. `internal/tools` called exactly five methods
(`IsToolEnabled`, `Get`, `LoadDynamic`, `MarkUnload`, `LoadableSkills`) and named
three types. 696 lines of hub code behind a five-method seam.

Evidence the previous author felt the pull and patched around it: `vars.go`
already declared an inline anonymous interface,
`toolReg interface{ Declarations() []ToolDecl }`, to avoid depending on the tool
registry. The seam wanted to exist; it just was not declared.

The split follows the rule *data in the hub, behavior in the spoke*:

- **Moved to `internal/skills`**: `ParseSkillMD`, `ParseSkillType`, `FlexibleList`,
  `SkillRegistry` and its methods, `VarRegistry` and the builtin renderers.
- **Stayed in `common/skill_types.go`**: `SkillType`, `LoadState`,
  `MCPServerConfig`, `SkillProperties`, `SkillEntry`, `SkillSummary`.

The data types had to stay. The new `Skills` and `Vars` interfaces mention them,
and a hub that named a spoke's type would have to import the spoke, which is the
one edge the topology forbids.

`internal/tools` now depends on `common.Skills` and `common.Vars`. The
composition roots (`agent.go`, `cmd/`) import the spoke directly, which is their
job.

Result: hub 4,917 to 4,310 lines, behavior 51% to 45%, new spoke 652 lines.

## What was done: settings becomes a spoke

Same shape, smaller. `SettingsStore` is a mutex, `settings.json` file I/O and
clamping: behavior, consumed by exactly one spoke (`ws`) through two methods,
`Get` and `ApplyRaw`.

- **Moved to `internal/settings`**: `SettingsStore`, `NewSettingsStore`,
  `LoadFromDisk`, `ApplyRaw`, persistence, and `clamp`. The tests moved with
  the code they exercise.
- **Stayed in `common`**: the `Settings` struct and its bound constants
  (`MinTemperature`, `MaxTokensCeiling` and the rest). That struct is the wire
  contract the GUI speaks, and the new `SettingsSource` interface mentions it.

This one illustrates the escape from method pinning. `clamp` was a method on
`Settings`, so it could not follow the store while `Settings` stayed behind.
It became a free function `clamp(s *common.Settings)` in the spoke. That
conversion is exactly what `segment.go` would need, at larger scale.

Result after both extractions: hub **4,917 to 4,071 lines**, behavior
**51% to 43%**.


## Verified

- `go build ./...`, `go vet ./...`, `go test ./...` all clean.
- Star topology re-checked with eight spokes: no spoke imports a spoke, hub
  imports no first-party package.
- Grader sweep, run serially, one at a time, after each extraction:
  ch5 120/120, ch7 through ch17 all 100/100.

## Two incidental findings

**The `star-topology` check has a stale spoke list.** `internal/grade/ch06_checks.go`
hardcodes `llm`, `tools`, `jobs`. The tree now has seven spokes; `mcp`, `recall`,
`ws` and the new `skills` are ungoverned by it. Also, the check only ever runs
against `solutions/ch06/agent`, so the live `./agent` tree is never checked for
star violations by the canonical sweep.

**`no-mutable-globals` has a false positive.** It flags
`var _ common.Skills = (*SkillRegistry)(nil)`, the standard compile-time
interface assertion, which declares no mutable state. The assertions were
removed rather than weaken the check, but the check is over-broad.

## Round two: give up method call syntax

The first round treated Go's method rule as a constraint to respect. That was
wrong, and Bill overruled it: a syntax restriction is not an architectural
argument. `Apply(c, e)` instead of `c.Apply(e)` costs nothing but sugar and
buys a hub that is actually vocabulary.

Applying that consistently found the real problem. `context.go` was a God
object: **863 lines, 581 of them function bodies, 20 methods on `Context`** -
and every exported one was called by exactly one spoke, `llm`. It was in the
hub purely because `Context` is declared there.

Moved to `internal/llm` as free functions over `*common.Context`:

- `context.go` behavior (the reducer `Apply` and 23 helpers) to `context_ops.go`
- `event_log.go` behavior (append, replay, on-disk form); the `Log` struct stays
- `journal.go` and `save.go` wholesale
- `segment.go`, the case round one declared blocked
- `band_test.go`, which left `internal/common` with no test files at all

`common/context.go` went from 863 lines to 166.

## The refined rule

Not all methods can become free functions. The exception is **stdlib interface
satisfaction**: `MarshalJSON`, `UnmarshalJSON`, `String` and `Error` are reached
by `encoding/json` and `fmt` through an interface, so they must be declared as
methods on the type. They stay in the hub with the type they serialize.

That is most of the behavior the hub still has: `band.go` has 4 such methods,
`part.go` 3, `event.go` 9, `interfaces.go` 4. It is a real floor, not laziness.

So the rule has three cases, in order:

1. Behavior used by one spoke: move it, as free functions over the hub type.
2. Behavior satisfying a stdlib interface: it must stay a method on the type.
3. Behavior used by several spokes: in a strict star it has nowhere else to go,
   unless the type becomes an interface in the hub with the implementation in a
   new spoke, wired by the composition root. That is what was done for `skills`
   and `settings`.

One consequence worth noting: when behavior moves to another package it loses
access to unexported fields. `Log.next` and `Log.clock` had to be exported.
For a pure data type in the hub that is consistent, but it is a genuine cost of
the split and should be a deliberate choice rather than a surprise.

## Result

| | before | after |
|---|---|---|
| non-test lines | 4,917 | **2,763** |
| files | 24 | **19** |
| function-body lines | 51% | **30%** |
| own test files | 1 | **0** |

Verified after each step, serially: build, vet and tests clean; star topology
intact across eight spokes; hub still imports no first-party package;
ch5 120/120 and ch7 through ch17 all 100/100.

## What chapter 5 says, and what it leaves out

Chapter 5 states the rule correctly:

> Everything in it is vocabulary: types, constants, interfaces. It has no
> behavior of its own.

But it never says what to do when behavior operates on a hub type. A coder hits
`cannot define new methods on non-local type`, and the shortest path is to put
the function in `common` - silently breaking the rule the chapter just stated.
That is exactly how this hub grew.

Two claims in that passage have also rotted. It describes the hub as "1,455
lines, seven files" and predicts it will grow "by an interface or two". And one
sentence actively licenses the drift:

> This is the largest package because the vocabulary is large, and that is
> correct.

Recommendation for the chapter: keep the rule, add the mechanism. Say that Go
will not let a spoke declare a method on a hub type, that the answer is a free
function in the spoke rather than a method in the hub, and name the stdlib
interface exception. Replace the absolute line counts with the invariant, since
counts rot and invariants do not.

## Remaining candidates

| file(s) | body lines | note |
|---|---|---|
| `band.go` | 151 | `Before`, `For`, `High`, `Low`, `Up`, `Zero` are llm-only and could move; the 4 JSON and String methods cannot. |
| `part.go`, `event.go` | 207 | Almost entirely stdlib interface methods. Floor. |
| `interfaces.go`, `pause.go`, `tools_decl.go` | 166 | Small; `pause.go` and `tools_decl.go` have no stdlib methods and are worth a look. |
| `thinking.go`, `delta.go`, `model.go` | 398 | The hub holds policy functions over the model table. Whether a hub should hold policy at all is the open question. |
| `save.go` `Save` | - | Dead code: zero consumers. Flagged, not deleted. |

## Method

Worth repeating, because it was faster than analysis: move the file, change the
package clause, and let the compiler enumerate the seam. Every conclusion above
that contradicted my static grep was produced by the compiler in seconds.

Two greps that lied, for the record:

- Counting `common.X` references with `grep -c` across multiple files and summing
  `$1` under `-F:` sums the *filename*, not the count. Every "zero root usage"
  figure it produced was false.
- Word-boundary grep for symbol names matches field names. `Text`, `File`,
  `Name`, `From`, `To` and `Model` produced a cloud of false consumers.
