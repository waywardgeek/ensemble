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

## Remaining candidates, with recommendations

| file(s) | lines | sole consumer | recommendation |
|---|---|---|---|
| `settings.go` | 275 | ws | **Done.** Promoted to `internal/settings`; see above. |
| `segment.go` | 203 | llm | **Needs a design decision.** Blocked by method pinning above. Requires converting methods on `common.Context` to free functions in `llm`. |
| `band.go`, `journal.go` | 502 | llm | `journal.go` is persistence behavior. `band.go` is mostly enum and config vocabulary and may legitimately belong in the hub. Note ch16 deliberately folded the memory store into `llm` to avoid a spoke-to-spoke edge; these are the residue of that decision. |
| `thinking.go`, `delta.go`, `model.go` | 398 | llm | **Keep in the hub.** Proven by compiler: `ThinkingEffort` and `DeltaKind` are used by the hub's own `config.go` and `observer.go`, and both consult `LookupModel`. The hub holds policy functions over the model table. Whether those policy functions should be in a hub at all is the open question. |
| `tools_decl.go` | 56 | (three) | **Keep.** `ToolDecl` is now referenced by `skills`, `llm` and `tools`. Genuinely shared. |

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
