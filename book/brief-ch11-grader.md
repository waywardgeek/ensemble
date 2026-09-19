# Ch11 Grader — Deletion Audit

## Checks (7 × 100 points)

| Check | Points | What It Protects |
|---|---|---|
| deterministic-context | 20 | Rebuild produces byte-identical context from event log |
| save-config | 10 | Save file includes system prompt, model, vendor, tools |
| save-roundtrip | 10 | Save file survives JSON marshal/unmarshal |
| resume-continues | 15 | --load resumes conversation from saved context |
| log-not-needed | 10 | Agent continues from context alone (no event log) |
| partial-replay | 20 | Partial log replay produces same context as full replay |
| ch10-parity | 15 | All ch10 checks still pass |

## Mutations (4 tested, 4 caught)

### Mutation 1: Delete Save() body (return nil)
- **Expected failures**: save-config, save-roundtrip, resume-continues, log-not-needed, partial-replay
- **Actual**: 15/100 — all 5 save-related checks fail ✅
- **Cascade**: ch10-parity passes (Save is independent of skills)

### Mutation 2: Break Rebuild (return empty context)
- **Expected failures**: deterministic-context
- **Actual**: 80/100 — deterministic-context fails ✅
- **Mechanism**: verify command compares rebuilt context to saved, detects MISMATCH
- **Note**: Required strengthening the grader — original check only validated structural properties, not Rebuild correctness. Added `verify` CLI command + grader calls it.

### Mutation 3: Delete persist in SettingsStore.ApplyRaw
- **Expected failures**: settings-persist (from ch10 parity)
- **Actual**: ch10-parity catches it ✅

### Mutation 4: Delete IsLoadable check in SkillRegistry
- **Expected failures**: blocked-skill (from ch10 parity)
- **Actual**: ch10-parity catches it ✅

## Coverage Analysis

- **save-config** and **save-roundtrip** share the Save path — Mutation 1 kills both
- **resume-continues** and **log-not-needed** share the Load path — both need --load to work
- **deterministic-context** is the only check that tests Rebuild via the verify command
- **partial-replay** tests the combination: save mid-conversation, replay remaining events
- **ch10-parity** inherits ch10's 4-mutation coverage

## Conclusion

All 4 mutations caught. The deterministic-context check was originally decorative for Rebuild
(tested structural properties only). Strengthened by adding a `verify` CLI command that
the grader invokes to confirm `Rebuild(log) == saved_context`.
