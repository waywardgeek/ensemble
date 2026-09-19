# Ch11 Review — Persistence

## Chapter thesis

The event log is the truth; the context is what the truth means right now.
Save and load make the reducer's correctness claim falsifiable.

## Decisions

1. SaveFile = {Context, Log, SaveConfig} — three fields, all JSON-marshaled.
2. SaveConfig captures vendor, model, system prompt, tools — not API keys.
   Portable across machines.
3. Rebuild(events) replays the full log from scratch. The verify command
   tests byte-identical comparison: `./agent verify save.json`.
4. Partial replay: save at event N, continue to event M, verify
   Context(0..M) == Context(0..N) + Apply(N..M).
5. Context sufficiency: agent continues from loaded context with empty log.
   The LLM never reads the log.
6. --save and --load flags on the CLI. No special modes, no subcommands
   (except verify).

## What the grader tests (7 checks, 100 points)

- deterministic-context 20: verify command proves Rebuild == saved context
- save-config 10: save file includes model, vendor, system prompt, tools
- save-roundtrip 10: JSON survives marshal/unmarshal cycle
- resume-continues 15: --load resumes conversation normally
- log-not-needed 10: empty log, loaded context, agent still works
- partial-replay 20: checkpoint + remaining events == full replay
- ch10-parity 15: all ch10 checks still pass

## Deletion audit (4 mutations, all caught)

See brief-ch11-grader.md.

## Voice

Dense, code-forward. 751 prose words. Three invariants stated as testable
properties, not aspirations. Motivational intro deferred (Bill's hint:
short chapters are fine for utility chapters).

## Open

- Interview/resurrection mode (mock tools, exact context replay) — future chapter.
- Memory cascade depends on persistence working correctly — next chapter.
