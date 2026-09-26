# ch16 mutation audit — plan and results

P9: each mutant deletes exactly ONE behaviour, is reverted after the run, and
the EXACT set of failing check IDs is recorded. A mutant that kills nothing is
a hole in the grader, not a success.

Driver: `/tmp/mutate.sh`. Raw output: `/tmp/mutants/M*.txt`.
Each run is a full grader pass (~7 minutes); the driver reverts with
`git checkout --` between mutants. All nine patches were confirmed to apply and
build before the audit started, so a silently non-applying patch could not fake
a clean result.

## Round 1

| # | Mutant | Score | Checks killed | Verdict |
|---|--------|-------|---------------|---------|
| M1 | compressor writes third person | 95 | memory-is-data | killed |
| M2 | conversation never measured after a checkpoint | 23 | micro-handoff-compresses, compressor-sees-the-work, graduation-fires-oldest-first, disable-enable-idempotent, disabled-neighbour-refused, replay-needs-no-llm, memory-is-data, fresh-start-populates | killed |
| M3 | graduation folds NEWEST instead of oldest | **100** | none | **SURVIVED** |
| M4 | force only warns, never removes tools | 88 | forced-handoff | killed |
| M5 | a disabled session band still compacts | **100** | none | **SURVIVED** |
| M6 | BandHas ignores text, comparing identity only | **100** | none | **SURVIVED** |
| M7 | replay re-runs the compressor | 83 | disable-enable-idempotent, fresh-start-populates | killed, but see note |
| M8 | abandoned launch retried immediately | **100** | none | **SURVIVED** |
| M9 | graduation into a disabled band drops content | 92 | disabled-neighbour-refused | killed |

Five killed, four survived.

### Why each survivor survived

- **M3.** The scenario never produces more session memories than a single fold
  consumes. When `n == len(files)`, `files[:n]` and `files[len(files)-n:]` are
  the same slice, so the mutant is *equivalent* in the only scenario the grader
  runs. To distinguish oldest from newest the band must hold strictly more
  files than one fold takes.
- **M5.** The band is switched off and back on, but no turn in between drives
  the conversation past its high watermark, so the suppressed code never had an
  opportunity to run. The check proves the band is restored, not that a
  disabled band stays out of the way.
- **M6.** The hand-edit check switches the band OFF before editing. That
  removes the band's entries, so `BandHas` finds nothing and returns false
  whatever it compares. The text comparison only matters on the path where the
  entry is still present — a plain restart, where the log replays the file's
  old contents. The check passed the reference for the wrong reason.
- **M8.** Nothing in the scenario takes a turn after the abandoned launch is
  discovered, so an immediate retry has no chance to be observed.

### Note on M7

M7 makes replay call the vendor, and `replay-needs-no-llm` is the check named
for exactly that property — but it was killed by *two other* checks instead.
That check counts vendor requests over a window that does not cover the
startup sync, so it does not yet see the call it exists to forbid.

## Round 2

Three checks strengthened. Two of the four survivors now die, one is an
equivalent mutant, and one remains an open hole with a known recipe.

| Mutant | Behaviour deleted | Round 1 | Round 2 | Now killed by |
|---|---|---|---|---|
| M5 | a disabled band still compacts | 100 survived | **88 killed** | `disable-enable-idempotent` |
| M6 | re-enable replays a stored snapshot instead of re-reading disk | 100 survived | **88 killed** | `disable-enable-idempotent` |
| M3 | graduation folds the newest memories | 100 survived | 100 **equivalent** | — see below |
| M8 | an abandoned graduation is retried on restart | 100 survived | not re-run | — see below |

### What killed M5, and why it took two attempts

The first attempt drove the conversation past its high watermark with the
session band switched off and relied on the harness's request-count
assertion to notice the extra call. It did not notice. The harness counts
*turn* requests, and a compressor is not a turn, so a compaction that should
never have happened was invisible to the very mechanism meant to catch it.

Counting the router's own calls across the phase kills it immediately. The
lesson generalises: a check that infers a background activity from a
foreground count can be blind to the thing it is named for. Count the thing
itself.

### What killed M6

Nothing in the disable-and-restore path could see it, because disabling a
band empties it and the restore then has nothing to compare against. The
check now also edits a memory file and *restarts without disabling
anything*. That is the case where a stored snapshot and a fresh read of the
disk disagree, and it is also the case a user actually hits: edit a memory
by hand, restart, expect the edit to count.

### M3 is equivalent, and the diagnostics say why

Folding from the oldest end and folding from the newest end are the same
operation whenever a fold consumes the whole band, and under this design a
fold always does. Graduation triggers on a byte watermark and then folds
`min(FoldFactor, len(files))`. Instrumenting the reference showed a fold
firing with five files in the band and taking all five; raising the turn
count to 32 produced five folds, each still taking everything it found, and
the first fold request was byte-identical under the mutation.

So there is no observable difference to grade. Making one would mean
changing the product — requiring a full batch before folding — which is
exactly the behaviour removed earlier in this chapter, because a band over
budget that cannot fill a batch would then never fold at all.

The check retains a first-fold assertion anyway. It costs nothing, and it
catches wrong-end folding in any implementation whose fold policy leaves
more behind than it takes.

### M8 remains open

The abandon scenario answers every compressor with a 500, so no memory is
ever written, the session band stays empty, and graduation — the only thing
the abandonment guard protects — is never attempted. The scenario cannot
observe the behaviour it was written for.

Closing it needs a compressor that fails *folds* while letting session
memories succeed. That router flag was written and then reverted: once a
fold fails, the harness's cumulative request counts stop being predictable,
and destabilising a passing grader to chase one mutant was the worse trade.
The recipe is recorded here rather than left implicit.

