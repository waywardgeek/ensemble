# Chapter 10 Coder Brief: Skills

## TL;DR

Build a skills engine. A skill is a directory containing a SKILL.md file with
YAML frontmatter and a markdown body. Loading a skill adds its instructions to
the agent's context and enables its declared tools. The system prompt is
rendered ONCE at agent creation and never mutated — dynamic skill loads flow
as events through the message history so the prompt-cache prefix is never
invalidated.

## What the student builds

### 1. Skill parsing

Parse SKILL.md: split on `---` (limit 3 — the body may contain horizontal
rules), YAML-decode the frontmatter, keep the body as a string.

Required frontmatter fields:
- `name` (string, kebab-case, required)
- `description` (string, required)

Optional frontmatter fields:
- `type` (string: `"primary"`, `"loadable"`, `"dependency"`; default `"loadable"`)
- `tools` (FlexibleList: tools this skill enables)
- `depends` (FlexibleList: skills auto-loaded when this skill loads)
- `loadable-skills` (FlexibleList: skills this skill makes available for `load_skill`)

**FlexibleList**: every list-valued field accepts BOTH formats:
```yaml
# space-separated on one line
depends: tool-files tool-search tool-context

# YAML list (indented one per line)
depends:
  - tool-files
  - tool-search
  - tool-context
```

Store internally as a space-joined string. Parse with `strings.Fields()` when
you need the slice.

### 2. Skill registry

A `SkillRegistry` that:
- Discovers skills from a directory tree (each subdirectory containing a
  SKILL.md is a skill)
- Stores parsed `SkillProperties` by name
- Tracks load state: `initial` (loaded at creation) vs `dynamic` (loaded
  via `load_skill` event) vs `pending-unload` (marked by `unload_skill`,
  removed at next context compaction)
- Resolves the `depends:` chain with circular-dependency detection (panic
  on a cycle — it's an authoring bug, not a runtime condition)
- Computes the loadable-skills closure: if skill A declares
  `loadable-skills: B` and skill B declares `loadable-skills: C`, then
  loading A makes both B and C available

### 3. Skill types

Three types control visibility:

| Type | Set at | Visible in prompt? | Loadable via tool? |
|---|---|---|---|
| `primary` | agent creation | yes (body in system prompt) | no |
| `loadable` | default | listed by name+description | yes |
| `dependency` | via `depends:` | no | no |

The primary skill's body is part of the system prompt, rendered once at
creation. Loadable skills are listed in the system prompt as available
(name + description only). Dependencies are invisible — auto-loaded when
their parent loads.

### 4. Variable substitution

Skill bodies (and the system prompt) may contain `$VAR_NAME` tokens. These
are rendered by a **VarRenderer** — a `func() string` keyed by variable name.

The top-level API accepts custom renderers:
```go
type VarRenderer func() string

agent.RegisterVar("TOOLS", builtinToolsRenderer)
agent.RegisterVar("SKILLS", builtinSkillsRenderer)
agent.RegisterVar("CUSTOM", myAppRenderer)
```

Built-in renderers (provided by the framework):
- `$TOOLS` — renders the currently-enabled tool declarations as a formatted
  list (name + description)
- `$SKILLS` — renders the currently-loadable skills as a list
  (name + description), so the agent knows what it can load

Application-specific renderers are registered before `NewAgent()`. They are
called at render time — once at creation for the initial system prompt, and
again at load time for dynamically loaded skill bodies.

Substitution: scan for `$IDENTIFIER` (uppercase, underscores, digits).
Unrecognized variables are left as-is (not an error — a skill written for
one agent may be loaded in another that lacks that renderer).

### 5. System prompt assembly (the cache-preservation architecture)

**The system prompt is rendered ONCE at agent creation and never mutated.**

At creation:
1. Render the primary skill's body with variable substitution
2. Append the loadable-skills listing (name + description for each)
3. Set `Config.SystemPrompt` to the result
4. Populate `Config.Tools` with declarations for all initially-enabled tools
5. These are sent as the `system` block and `tools` array in the first
   request and are cache-stable for the agent's lifetime

When `load_skill` is called at runtime:
1. The skill's body (with variable substitution) becomes a **data message**
   in the conversation history — NOT a system prompt mutation
2. The skill's tools are added to the registry with `source: "dynamic"`
3. For Anthropic: new tool declarations are appended to the `tools` array
   (the existing prefix stays cached)
4. For other vendors: the renderer MAY re-render everything (cache miss,
   functionally correct)

When `unload_skill` is called:
1. The skill is marked `pending-unload`
2. Its tools remain callable (no cache miss)
3. At the next context compaction (micro_handoff), the skill's instructions
   and tool declarations are removed from the context

### 6. Tool provenance

Every entry in the tool registry carries:
```go
type ToolEntry struct {
    Decl     ToolDecl
    Handler  ToolHandler
    Source   ToolSource  // Initial or Dynamic
    EventSeq int        // 0 for initial; event sequence number for dynamic
}
```

Most renderers ignore Source/EventSeq. The Anthropic renderer uses them to
avoid re-sending initial declarations when new tools are added dynamically.

### 7. Built-in tools: `load_skill` and `unload_skill`

`load_skill`:
- Input: `{"name": "skill-name"}`
- Validates the skill exists and is loadable (type=loadable, or made available
  by a loaded skill's `loadable-skills`)
- Resolves and auto-loads `depends:` chain
- Enables the skill's declared tools
- Returns the skill's rendered body (instructions) as the tool result
- Records a `SkillLoaded` event in the event log
- Updates the loadable-skills list (the newly loaded skill may declare more)

`unload_skill`:
- Input: `{"name": "skill-name"}`
- Marks the skill `pending-unload`
- Returns confirmation
- The skill's tools remain callable until context compaction

### 8. The exercise

Create test fixture skills in a `skills/` directory:

```
skills/
├── base/
│   └── SKILL.md        # type: primary, tools: read_file think
│                        # loadable-skills: code-tools
├── ensemble/
│   └── SKILL.md        # type: primary, tools: all 12 core tools
│                        # loadable-skills: code-tools search-tools
├── code-tools/
│   └── SKILL.md        # type: loadable, tools: edit_file write_file
│                        # loadable-skills: search-tools
├── search-tools/
│   └── SKILL.md        # type: loadable, tools: search_files
│                        # depends: search-helpers
├── search-helpers/
│   └── SKILL.md        # type: dependency (auto-loaded, invisible)
└── blocked/
    └── SKILL.md        # type: loadable (but NOT in any loadable-skills chain)
```

The exercise tests TWO primary skills:
- `base`: minimal agent (read_file, think, load_skill, unload_skill).
  Tests progressive disclosure — the agent starts small and loads more.
- `ensemble`: full agent with all core tools. Tests that a production-
  grade agent gets everything it needs from its primary skill.

The system prompt sent to the vendor MUST contain the primary skill's
body text with variables substituted. The grader inspects the `system`
field of the first vendor request to verify this.

### 9. Grader checks (~100 points)

| Check | Points | What it tests |
|---|---|---|
| `initial-tools` | 10 | Agent starts with only base skill's tools + load/unload |
| `system-prompt` | 10 | System prompt contains primary skill's rendered body text |
| `ensemble-tools` | 10 | With EN_PRIMARY_SKILL=ensemble, all 12 core tools visible |
| `load-skill` | 15 | `load_skill("code-tools")` adds new tools to declarations |
| `progressive-disclosure` | 15 | Loading code-tools makes search-tools loadable |
| `depends-autoload` | 10 | Loading search-tools auto-loads search-helpers (list_directory appears) |
| `var-substitution` | 10 | A skill body containing `$CUSTOM_VAR` renders the registered value |
| `blocked-skill` | 5 | Non-loadable skills rejected with error |
| `ch9-parity` | 15 | All ch9 checks pass |

### 10. Key design rules

1. **System prompt is a constitution** — written once, cached, immutable.
   Dynamic content flows as data messages in the conversation history.

2. **Tools are declared, not discovered** — the agent knows its initial
   tools from Config.Tools. Dynamic tools are added via events, not by
   re-rendering the system prompt.

3. **Provenance is per-entry, not per-batch** — each tool and skill knows
   whether it was initial or dynamic. The Anthropic renderer reads this;
   others may ignore it.

4. **Variable renderers are called at render time** — once at creation for
   the initial system prompt, again at load time for dynamic skill bodies.
   Values are frozen at render time, never retroactively updated.

5. **Unload is lazy** — marking a skill for removal costs nothing (no cache
   miss). Actual removal happens at context compaction.

6. **Cycles are authoring bugs** — panic on circular `depends:`, don't
   silently break the chain.

7. **Unknown variables are left as-is** — `$MISSING_VAR` stays in the
   text. A skill from another agent that references variables you don't
   have should not crash your agent.

## Ruling requests for the author

1. Should the `SkillLoaded` event appear in the event log (GUI sees it)?
   I assumed yes — it's an observable state change.

2. Should the rendered skill body be visible in the GUI's chat pane (like
   a system message), or only in the API log? I assumed chat pane — the
   human should see what instructions the agent just received.

3. Do we need a `skills/` directory in the exercise tree, or should the
   grader create fixture skills programmatically in a temp dir?
   I lean toward fixture dir committed to the repo — students can read
   the SKILL.md files and understand what they're building toward.
