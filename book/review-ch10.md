# Ch10 Review — Skills

## What was built

### Engine
- `internal/common/skill.go`: SkillProps, ParseSkillMD, FlexibleList (space-sep + YAML)
- `internal/common/skill_registry.go`: SkillRegistry (discover, load, dependency chain, cycle detect, progressive disclosure)
- `internal/common/vars.go`: VarRegistry ($VAR substitution, built-in $TOOLS/$SKILLS, custom renderers via API)
- `internal/tools/tools.go`: ToolMeta (provenance: Initial/Dynamic), WireSkills (load_skill/unload_skill tools), onToolsChanged callback, InitialDeclarations/DynamicDeclarations
- `internal/common/event.go`: SkillLoaded event type + SkillData
- `internal/common/event_log.go`: Len() method

### Public API
- `agent.go`: RegisterVar, DiscoverSkills, LoadSkill, WireSkillTools

### Fixtures
- `agent/skills/`: base (primary), code-tools (loadable), search-tools (loadable, depends search-helpers), search-helpers (dependency), blocked (type: dependency, not loadable)

### Grader
7 checks, 100/100. P9 audit: 4 mutants, all caught.

## Key decisions

1. **System prompt is a constitution**: rendered once at creation, never mutated. Dynamic skill loads flow as tool results in message history.
2. **Tool provenance**: Initial vs Dynamic tracking enables cache-aware Anthropic rendering without the renderer knowing about skills.
3. **FlexibleList**: both space-separated and YAML list formats. Three-limit `SplitN("---")` protects against horizontal rules in body.
4. **Unload is lazy**: `PendingUnload` state, removed at next compaction. No cache miss.
5. **VarRegistry callbacks via API**: `RegisterVar` exposes application-specific renderers. Built-in $TOOLS and $SKILLS handle the common case.
6. **onToolsChanged callback**: the bridge between tool handlers and cfg.Tools. Without it, `load_skill` runs but the next request sees stale tool declarations.

## Bugs found during build

1. **onToolsChanged never wired in cmd/main.go**: The Agent wrapper calls `SetOnToolsChanged` but `cmd/main.go` uses internal API directly. Load_skill ran but tools never updated. Fixed by adding callback in main.go.
2. **Value vs pointer map entries**: `map[string]SkillEntry` vs `map[string]*SkillEntry` determines whether map mutations propagate. We use `*SkillEntry` (pointer) correctly.

## Gotchas

- Go build cache across worktrees with same module path serves stale binaries. Always `go clean -cache` for mutation testing.
- `edit_file` matches first occurrence. When two methods have identical lines (e.g. `s.persist()` in both Apply and ApplyRaw), use `replace_lines` with exact line numbers.
