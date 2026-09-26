# ch16 mutation audit — plan and results

P9: each mutant deletes exactly ONE behavior, is reverted after the run, and
the EXACT set of failing check IDs is recorded. A mutant that kills nothing
is a hole in the grader.

Run one mutant:
```
cd ~/projects/ensemble
<apply patch>
go run ./cmd/grade -ch 16 ./agent 2>&1 | grep -E '^\[(PASS|FAIL)|/100'
git checkout -- <file>
```
A full grader run is ~7 minutes.

## The nine required mutants (brief §3.13)

| # | Mutant | File | Patch | Expected to kill |
|---|--------|------|-------|------------------|
| M1 | Compressor writes third person | agent/internal/llm/compress.go | "first person" -> "third person" in compressorPrompt | memory-is-data |
| M2 | Conversation never measured after a checkpoint | agent/internal/llm/compact.go | early `return` at top of compactConversation | micro-handoff-compresses, planted-fact-survives, memory-is-data, graduation-fires-oldest-first, disable-enable-idempotent, disabled-neighbor-refused, replay-needs-no-llm, abandon-on-restart, fresh-start-populates |
| M3 | Graduation folds newest instead of oldest | agent/internal/llm/compact.go | `src := files[:n]` -> `src := files[len(files)-n:]` | graduation-fires-oldest-first |
| M4 | Force only warns, never removes tools | agent/internal/llm/forcing.go | early `return` at top of restrict | forced-handoff |
| M5 | Disable does not suppress the cascade | agent/internal/llm/compact.go | drop the `cfg.Session.Disabled` guard | disable-enable-idempotent |
| M6 | Re-enable replays a stored copy instead of re-reading disk | agent/internal/common/context.go | BandHas compares ID only, ignoring text | disable-enable-idempotent |
| M7 | Replay re-runs the compressor | agent/internal/llm/bands.go | on populate, re-compress the file instead of loading it | replay-needs-no-llm |
| M8 | Abandoned launch retried immediately | agent/internal/llm/compact.go | drop the abandonedSince guard | abandon-on-restart |
| M9 | Graduation into a disabled band drops content | agent/internal/llm/compact.go | on disabled neighbour, retire sources without populating | disabled-neighbor-refused |

## Results

(filled in as each runs)

| # | Killed by | Verdict |
|---|-----------|---------|
| M1 | | |
| M2 | | |
| M3 | | |
| M4 | | |
| M5 | | |
| M6 | | |
| M7 | | |
| M8 | | |
| M9 | | |

## Notes

- M6 was found to be a REAL BUG in the reference, not just a mutant: replay
  restored the old BandPopulated text, SyncBands matched on file ID alone and
  skipped the file, so an edit on disk never reached the context. Fixed by
  making BandHas compare text as well as identity. The mutant now re-introduces
  the original bug, which is exactly what a good mutant should do.
