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

Four checks strengthened, then M3, M5, M6, M8 re-run.

(results below)
