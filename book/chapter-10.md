# Chapter 10: Skills

Every chapter so far has added capability by adding code. That stops
scaling. By chapter 9 the system prompt is a monolith, the tool list
is a wall of forty declarations, and every edit trashes the prompt
cache. The agent tries to call tools it has no business calling
because it can see all of them. Skills fix this. A skill is a named
unit of capability (instructions, tools, dependencies) in a single
markdown file. Loading one changes what the agent can do. Unloading
one marks it for removal at the next compaction. The system prompt is
rendered once and never mutated.

This is also where variable substitution enters. A skill body can
contain `$TOOLS` or `$SKILLS` or any application-specific variable,
and the renderer replaces them at the point of use. The initial
system prompt gets creation-time values. A dynamically loaded skill
gets load-time values. No retroactive updates, no invalidation logic.

## TL;DR

A SKILL.md file lives in a named directory under the skills root:

```
skills/
  base/
    SKILL.md
  code-tools/
    SKILL.md
  search-tools/
    SKILL.md
```

Each SKILL.md has YAML frontmatter between `---` markers and a
markdown body:

```yaml
---
name: code-tools
description: File editing and search tools
type: loadable
tools: edit_file write_file search_files
depends: read-tools
loadable-skills: refactor-tools
---

You now have access to file editing tools.

Available tools: $TOOLS
Available skills to load: $SKILLS
```

**Frontmatter fields:**
- `name`: skill identifier
- `description`: one-line summary
- `type`: `primary` (agent identity, loaded at creation), `loadable`
  (offered to the agent, loaded on demand), or `dependency` (pulled
  in automatically, hidden)
- `tools`: tools this skill enables (FlexibleList: space-separated
  or YAML list)
- `depends`: skills to auto-load when this skill loads
- `loadable-skills`: skills this skill makes available for the agent
  to discover

**Exercise contract:** `EN_SKILLS_DIR=path/to/skills EN_PRIMARY_SKILL=base ./ensemble prompt "load the code-tools skill"`

The grader verifies:
- Initial tool set matches primary skill's declared tools plus
  `load_skill` and `unload_skill` (15 pts)
- `load_skill("code-tools")` adds that skill's tools to
  declarations (15 pts)
- Loading a skill with `loadable-skills` reveals those skills in
  `$SKILLS` (15 pts)
- Dependencies auto-load and contribute their tools (15 pts)
- `$VAR` placeholders in skill body are substituted (10 pts)
- Non-loadable skills are rejected with an error (10 pts)
- All chapter 9 checks still pass (20 pts)

## The Format

The parser splits the file on `---` with a limit of 3, protecting
against horizontal rules in the body. Frontmatter is key-value,
one field per line. Multi-value fields use FlexibleList. Both
formats parse to `[]string`:

```yaml
# Space-separated on one line
tools: edit_file write_file search_files

# YAML list
tools:
  - edit_file
  - write_file
  - search_files
```

Three skill types control visibility. A `primary` skill loads at
agent creation and defines the agent's identity. A `loadable` skill
appears in the system prompt's available-skills list. The agent
calls `load_skill` to activate it. A `dependency` skill is invisible
to the agent; it loads automatically when a skill that depends on it
loads.

## The Registry

`SkillRegistry` discovers skills from a directory (one subdirectory
per skill) and tracks their state through a lifecycle:

```
Discovered → LoadInitial (at creation)
           → LoadDynamic (by load_skill tool)
           → PendingUnload (by unload_skill, removed at compaction)
```

Loading resolves dependencies recursively. If `search-tools` depends
on `search-helpers`, both load and both contribute their tools. Cycle
detection uses a visiting set. A depends B depends A produces an
error, not infinite recursion.

The registry answers two questions the agent asks constantly: "what
tools can I use?" (`IsToolEnabled`) and "what skills can I load
next?" (`LoadableSkills`). The loadable set is the union of all
loaded skills' `loadable-skills` lists, minus anything already loaded.

## Progressive Disclosure

This is the payoff. Loading a skill reveals more skills. The agent's
`$SKILLS` variable updates to show what it can load next. After
loading `code-tools`, it might see `refactor-tools`. After loading
`refactor-tools`, it might see `architecture-tools`. The tree unfolds
as the agent explores.

The agent never sees the full tree. It sees one level ahead: the skills
the skills it has loaded make available. This keeps the context
focused and the tool list manageable.

## Variable Substitution

Skill bodies contain `$VAR` placeholders. A `VarRegistry` maps names
to renderer functions:

```go
vars := common.NewVarRegistry()
vars.RegisterRenderer("TOOLS", func() string {
    return renderToolList(reg.Declarations())
})
vars.RegisterRenderer("SKILLS", func() string {
    return strings.Join(sr.LoadableSkills(), ", ")
})
// Application-specific
vars.RegisterRenderer("CUSTOM_VAR", func() string {
    return os.Getenv("EN_CUSTOM_VAR")
})
```

Built-in renderers handle `$TOOLS` and `$SKILLS`. The API exposes
`RegisterVar` for application-specific variables: settings,
environment, user context, anything the skill body needs to reference
without hardcoding.

Variables are rendered at point of use. The initial system prompt gets
creation-time tool and skill lists. A skill loaded mid-conversation
gets the current lists at load time. No retroactive updates.

## Tool Provenance

Every tool in the registry tracks its source:

```go
type ToolMeta struct {
    Source   ToolSource // SourceInitial or SourceDynamic
    EventSeq int       // 0 for initial, event seq for dynamic
}
```

Most renderers ignore this. But Anthropic's prompt cache hits on
prefix match. Mutating the tools array causes a miss. An
Anthropic-aware renderer can split initial tools (cached prefix) from
dynamic tools (appended without cache miss). The provenance tracking
makes this possible without the renderer needing to know about skills.

## The Constitution

The system prompt is rendered once at agent creation and never
mutated. When `load_skill` runs, the skill's instructions arrive as
the tool result. They flow through the message history, not the
system prompt. The agent's capabilities grow but the constitution is
stable.

This has a consequence for memory. Pre-loaded context (memories,
user preferences, project notes) belongs in the message history as
data, not in the system prompt as instructions. `save_memory` produces
a data message. The system prompt stays clean and cacheable. Chapter
12 will build the memory cascade on this foundation.
