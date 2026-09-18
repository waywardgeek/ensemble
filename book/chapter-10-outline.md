# Chapter 10 Outline: Skills

## Through-line stake
The system prompt is architecture, not text. Skills are the unit of
capability — each one declares what tools it brings, what it depends
on, and what it makes available next. The agent that earns new
capabilities by loading skills is an agent that can grow without
rewriting its own scaffolding.

## §10.0 Motivational intro
Skills solve two problems that got worse as the tool count grew: the
LLM sees tools it cannot use yet, and the system prompt is a
monolith that breaks cache on every edit. A skill is a named unit of
capability — instructions, tools, dependencies, and variables — in a
single markdown file. Loading one changes what the agent can do
without trashing the system prompt. This is where the agent starts
to configure itself.

## §10.1 TL;DR
- SKILL.md frontmatter: name, description, type (primary/loadable/
  dependency), tools (FlexibleList), depends (FlexibleList),
  loadable-skills (FlexibleList)
- SkillRegistry: discovery, loading, dependency resolution with
  cycle detection, progressive disclosure
- VarRegistry: $VAR substitution in skill body, built-in renderers
  ($TOOLS, $SKILLS), custom renderers via API
- Tool provenance: Initial vs Dynamic, per-entry tracking
- load_skill/unload_skill built-in tools
- System prompt rendered once at creation (cache-stable). Dynamic
  loads flow as events, never mutate the prompt.

Check table from grader.

## §10.2 The why
Every production agent framework hits the same wall: 40 tools in one
prompt burns context and confuses the model. The agent tries to call
tools it has no business calling. The fix is progressive disclosure:
start small, reveal more when the agent asks for it. Skills are the
container that makes disclosure work — each one is a named scope of
capability with explicit boundaries.

The cache story matters. Anthropic's prompt cache hits on prefix match.
Every system prompt mutation causes a miss. Render once, append
dynamically, never mutate.

## §10.3 SKILL.md format
A directory with a SKILL.md file. Frontmatter between `---` markers.
Body is instructions appended to system prompt when loaded.

```yaml
---
name: code-tools
description: File editing and search tools
type: loadable
tools: edit_file write_file search_files
depends: read-tools
loadable-skills: refactor-tools
---
```

FlexibleList: space-separated on one line or YAML indented list. Both
parse to []string. The 3-limit on strings.SplitN("---") protects
against horizontal rules in the body.

Three types:
- primary: agent identity, loaded at creation
- loadable: offered in system prompt, loaded on demand
- dependency: pulled in automatically, hidden from the agent

## §10.4 Parsing
Walk the SKILL.md file. Split on `---` (limit 3). Parse frontmatter
as key-value. The FlexibleList parser handles both formats.

## §10.5 SkillRegistry
Discover skills from a directory (one subdir per skill). Track state:
Discovered → LoadInitial/LoadDynamic → PendingUnload. Load resolves
dependencies recursively with cycle detection via a visiting set.

LoadableSkills() returns what the agent can load next — union of all
loaded skills' loadable-skills lists, minus already-loaded.

## §10.6 Dependency resolution
Loading "search-tools" which depends on "search-helpers": both load,
both contribute tools. Cycle detection: A depends B depends A →
error, not infinite recursion.

## §10.7 Progressive disclosure
The key insight: loading a skill reveals MORE skills. The agent's
$SKILLS variable shows what it can load. After loading code-tools,
it might see refactor-tools. After loading refactor-tools, it might
see architecture-tools. The tree unfolds as the agent explores.

## §10.8 Variable substitution
Skill bodies contain $VAR placeholders. VarRegistry maps names to
renderer functions. Built-in: $TOOLS (current tool list), $SKILLS
(loadable skills). Custom: application-specific, registered via API.
Rendered at point of use — initial prompt gets creation-time values,
dynamic loads get load-time values.

## §10.9 Tool provenance
Every tool entry tracks whether it was present at creation (Initial)
or added by a dynamic skill load (Dynamic), plus the event sequence
number. Most renderers ignore this. The Anthropic renderer uses it to
split initial tools (in the cached prefix) from dynamic tools
(appended without cache miss).

## §10.10 System prompt as constitution
Rendered once. Never mutated. Dynamic skill instructions arrive as
tool results from load_skill — they flow through the message history,
not the system prompt. The agent's capabilities grow but the
constitution is stable.

This also means memories are data, not instructions. Pre-loaded
context goes in the message history as user-role data. save_memory
produces data messages. The system prompt stays clean.

## §10.11 What Chapter 11 does with this
MCP servers. A skill declares `mcp_servers:` in its frontmatter.
Loading the skill starts the server. The server's tools join the
registry as Dynamic entries. The GUI gets an MCP server that lets the
LLM see what you see.
