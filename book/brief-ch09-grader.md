# Ch9 Grader — P9 Deletion Audit

## Checks (5 checks, 100 points)

| ID                  | Points | What it guards                                    |
|---------------------|--------|---------------------------------------------------|
| settings-roundtrip  | 25     | update_settings → settings_changed echo           |
| settings-on-connect | 20     | current_settings sent on subscribe                |
| settings-broadcast  | 20     | second client receives settings_changed           |
| settings-persist    | 15     | settings survive server restart (file persistence)|
| ch8-parity          | 20     | all ch8 checks still pass                         |

## Mutations (4 mutants, all caught)

### Mutant 1: no-update-handler (handler.go)
Delete `case "update_settings":` block.

**Result:** 20/100 — fails `{settings-roundtrip, settings-on-connect, settings-broadcast, settings-persist}`
**Analysis:** Broad kill — without the handler, no settings are ever applied in memory or on disk. on-connect fails because the second test phase expects settings from the first phase's updates. **Deletes exactly one behavior** (the handler), kills 4 checks by cascade.

### Mutant 2: no-current-on-subscribe (handler.go)
Delete `if h.settings != nil { ... }` block at end of `subscribe()`.

**Result:** 65/100 — fails `{settings-on-connect, settings-persist}`
**Analysis:** settings-on-connect fails directly (new connection gets no current_settings). settings-persist cascades: the persist test restarts the server and checks current_settings on the new connection — without the subscribe delivery, the test can't observe the persisted values.

### Mutant 3: no-broadcast (handler.go)
Delete `h.broadcastSettings(updated)` call in update_settings handler (keep ApplyRaw).

**Result:** 55/100 — fails `{settings-roundtrip, settings-broadcast}`
**Analysis:** settings-roundtrip fails because the sender relies on receiving settings_changed via broadcast (broadcast goes to ALL clients, including sender). settings-broadcast fails directly (second client never gets the message). Persist PASSES because ApplyRaw still persists to disk.

### Mutant 4: no-persist (settings.go ApplyRaw)
Delete `s.persist()` call in `ApplyRaw()` method.

**Result:** 65/100 — fails `{settings-on-connect, settings-persist}`
**Analysis:** settings-persist fails directly (no file written, second server starts with empty settings). settings-on-connect cascades: the second server has empty settings, so current_settings is empty. Same cascade shape as mutant 2.

**GOTCHA:** `Apply()` and `ApplyRaw()` both contain `s.persist()` / `return s.data`. `edit_file` matches the FIRST occurrence. Must use `replace_lines` with exact line number to mutate the right method.

## Coverage Matrix

| Check               | M1-handler | M2-subscribe | M3-broadcast | M4-persist |
|---------------------|-----------|--------------|--------------|------------|
| settings-roundtrip  | FAIL      | pass         | FAIL         | pass       |
| settings-on-connect | FAIL      | FAIL         | pass         | FAIL       |
| settings-broadcast  | FAIL      | pass         | FAIL         | pass       |
| settings-persist    | FAIL      | FAIL         | pass         | FAIL       |
| ch8-parity          | pass      | pass         | pass         | pass       |

Every non-parity check is killed by at least one mutation that doesn't kill ALL checks — the matrix shows independent failure modes. ✓

## Cascade relationships
- settings-on-connect and settings-persist share the subscribe delivery path (both need current_settings on subscribe)
- settings-roundtrip depends on broadcast (sender receives its own echo via broadcast)
- ch8-parity is immune to all settings mutations (tested independently)
