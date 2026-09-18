# Ch9 Code Review: Build Your Dream GUI

## What was built

### Server-side
- `internal/common/settings.go`: `Settings` struct (theme, tts_enabled, tts_speed, font_size), `SettingsStore` with `Apply`, `ApplyRaw`, thread-safe with mutex, JSON persistence.
- `internal/ws/handler.go`: `update_settings` handler → `ApplyRaw` + `broadcastSettings`; `current_settings` sent on subscribe; `NewHub` takes `*common.SettingsStore`.
- `agent.go`: Type aliases for `Settings`, `SettingsStore`, `NewSettingsStore`.
- `cmd/main.go`: Creates `SettingsStore` at `./settings.json`, passes to `NewHub`.

### Client-side
- `web/gui/index.html`: Three-pane layout (sidebar, center, actions), drag bars, sidebar tabs (Agents/Settings).
- `web/gui/styles.css`: CSS custom properties for theming (~15 vars), dark/light/system themes, three-pane flexbox, drag bars, agent tree, settings panel, sidebar tabs.
- `web/gui/gui.js`: Message routing (chat vs actions), drag bar logic, settings panel, theme application, agent tree, sidebar tab switching.
- `web/gui/tts.js`: Updated with `enabled`/`rate` settings, `cancel()` method.
- `web/gui/artifact-scroll.js`: Added `error` message type handling.

### Grader
- `internal/grade/ch09_harness.go`: `Ch9Run` with settings roundtrip, on-connect, broadcast, persist, parity.
- `internal/grade/ch09_checks.go`: 5 checks (100 points total).
- `cmd/grade/main.go`: Case 9 dispatch.
- `Makefile`: `grade9` target.

### Exercise, solutions, snapshot
- `solutions/ch09/`: Complete snapshot.

## Grader checks
| check | pts | mechanism |
|---|---|---|
| settings-roundtrip | 25 | Send update_settings{theme:"dark"}, wait for settings_changed, verify theme matches |
| settings-on-connect | 20 | Connect new client, verify current_settings received with correct theme |
| settings-broadcast | 20 | Send update from client A, verify client B gets settings_changed |
| settings-persist | 15 | Set theme, kill server, restart with same data dir, verify current_settings has theme |
| ch8-parity | 20 | All ch8 checks pass |

## P9 deletion audit
4 mutations, all caught. Coverage in `book/brief-ch09-grader.md`.

## Questions for Bill
1. Chapter prose is ~1065 words (dense, code-heavy). Consistent with ch8 direction. OK?
2. Agent tree currently single node — sub-agents will appear in a later chapter. Is the stub sufficient?
3. Settings struct uses fixed fields (theme, tts_enabled, tts_speed, font_size). Should it support arbitrary key-value pairs instead?
