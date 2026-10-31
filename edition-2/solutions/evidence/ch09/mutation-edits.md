# Exact Chapter 9 mutation edits

Student evidence clarification requested during initial review. These replacements
are transcribed from the retained tool-call scripts used for the original runs,
not inferred from the failure messages. Each was applied separately to an isolated
copy, then the original file was restored. The temporary driver scripts/copies
were deleted; the tool-call text and output remain the source of this mapping.
Paths below are relative to `solutions/edition-2/main`. Replacements used the
first exact occurrence unless noted. Original results are in `mutations.txt` and
`browser-mutations.json`. This mapping does not claim a new run.

## Go mutations

All but `negative_public_budget` ran `go test -count=1 -timeout=35s -run <test>
./...` from the scratch GUI module. Settings cases used
`TestSettingsWireAndDisk`; budget cases used `TestSavedBudgetEntry`.

| Recorded name | File | Exact old expression → exact replacement |
| --- | --- | --- |
| entry_validation | `gui/settings.go` | `next, err := settingsPatch(s.snapshotSettings(), raw)` → `next := s.snapshotSettings(); err := json.Unmarshal(raw, &next)` |
| load_validation | `gui/settings.go` | `next, err := settingsPatch(s.snapshotSettings(), data)` → `next := s.snapshotSettings(); err = json.Unmarshal(data, &next)` |
| false_snapshot | `gui/settings.go` | JSON tag `json:"tts_enabled"` → `json:"tts_enabled,omitempty"` |
| zero_snapshot | `gui/settings.go` | JSON tag `json:"max_tool_rounds"` → `json:"max_tool_rounds,omitempty"` |
| patch_absence | `gui/settings.go` | `next, err := settingsPatch(s.snapshotSettings(), raw)` → `next, err := settingsPatch(Settings{}, raw)` |
| skip_restore | `gui/settings.go` | `if os.IsNotExist(err) {` → `if true {` |
| broadcast | `gui/gui.go` | `s.broadcast(map[string]any{"type": "settings_changed", "settings": settings})` → `data, _ := json.Marshal(map[string]any{"type":"settings_changed", "settings":settings}); queue <- [][]byte{data}` |
| failed_write_publishes | `gui/settings.go` | `path := s.settingsPath()` → `s.settings = next` followed by `path := s.settingsPath()` |
| skip_budget_apply | `gui/settings.go` | `s.agent.SetMaxToolRounds(next.MaxToolRounds) // Already clamped by settingsPatch.` → `// Skip runtime update.` |
| skip_budget_restore | `gui/settings.go` | `return s.agent.SetMaxToolRounds(next.MaxToolRounds)` → `return nil` |
| limit_16 | `internal/agent/actor.go` | `limit := r.owner.Config().MaxToolRounds` → `limit := min(16, r.owner.Config().MaxToolRounds)` |
| default_16 | `internal/agent/actor.go` | `limit = common.DefaultMaxToolRounds` → `limit = 16` |
| midturn_budget | `internal/agent/actor.go` | `m.turn.batches >= m.turn.limit` → `m.turn.batches >= r.owner.Config().MaxToolRounds` |
| load_ui_validation | `gui/settings.go` | Before the load path's `s.settings = next` and `return s.agent.SetMaxToolRounds(next.MaxToolRounds)`, insert `next.TTSSpeed = 9; next.FontSize = 0; next.Theme = "wrong"`. This forces wrong loaded UI values after validation. |
| corrupt_load_ignored | `gui/settings.go` | In the block immediately after `next, err := settingsPatch(s.snapshotSettings(), data)`, replace `if err != nil { return err }` with `if err != nil { return nil }` (original source has normal Go newlines/tabs). |
| missing_current_snapshot | `gui/gui.go` | `add(map[string]any{"type": "current_settings", "settings": s.snapshotSettings()})` → `// omitted current snapshot` |
| negative_public_budget | `internal/agent/agent.go` | `if limit < 0 {` → `if false {`. Run `go test -count=1 -run TestBudgetUpdateRejectsNegative ./ensemble` from scratch core. |

The broad `load_validation` mutant bypasses the clamp, so startup rejects the
negative disk budget through the real public setter. Its output is that intended
load failure, not a compile/setup defect. It does **not** separately demonstrate
that the test catches each unvalidated UI field: the budget rejection occurs
first. The additional `load_ui_validation` mutant forces wrong UI values after
load validation and fails the earlier valid-save restart assertion. That is
sensitivity to corrupt restored UI values, not an independently isolated mutation
of each clamp branch. The correct baseline does assert all invalid loaded fields.
This is a limitation in the granularity of the mutation evidence, not a claim
that the additional mutant reached the later invalid-disk assertion.

The `midturn_budget` mutant reads the changed public configuration at each batch
boundary. In the executable test the small turn began with a saved limit of 2,
then received an update to 18 while its first provider response was gated. The
mutant executes 18 rather than 2 and fails that specific active-turn assertion.
It does not merely alter startup or the next turn.

## Browser mutations

All static paths are under `gui/web/gui/`. Each ran the retained browser script
against the real unchanged GUI executable and an isolated static directory.
For these runs only, `context.setDefaultTimeout(6000)` was added after browser
context creation in the temporary launcher; explicit native callback timeout
remained unchanged. `CH09_WEB` selected the mutated directory, `CH09_EVIDENCE`
selected a temporary result file. These launcher changes do not mutate behavior.

| Recorded name | File | Exact old expression → exact replacement |
| --- | --- | --- |
| tab_state | `app.js` | `other.setAttribute("aria-selected", selected);` → `other.setAttribute("aria-selected", false);` |
| theme_apply | `app.js` | `document.body.dataset.theme = theme.value === "system" ?` → `document.body.dataset.theme = true ? "dark" : theme.value === "system" ?` |
| font_apply | `app.js` | ``document.body.style.fontSize = `${settings.font_size}px`;`` → `document.body.style.fontSize = "16px";` |
| system_change | `app.js` | `scheme.addEventListener("change", applyTheme);` → `// no theme listener` |
| drag_bar | `app.js` | ``pane.style.width = `${next}px`;`` → `// no width update` |
| state_indicator | `style.css` | `[data-state="in_flight"] .agent-indicator, [data-state="tools_pending"] .agent-indicator` → `[data-state="thinking"] .agent-indicator, [data-state="tool_use"] .agent-indicator` |
| pane_routing | `app.js` | `(tool ? actionsScroll : chatScroll).receive(message);` → `chatScroll.receive(message);` |
| speech_speed | `app.js` | `utterance.rate = Number(document.querySelector("#tts-speed").value);` → `utterance.rate = 1;` |

The speech-speed failure retains the actual native queued rate 1 plus start/end
callbacks. Other browser outputs retain the checks completed before the intended
missing state/value or direct assertion failure. No browser mutant failed because
of script syntax, unavailable Chrome, or an unavailable native speech service.
