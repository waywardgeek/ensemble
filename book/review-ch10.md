# Chapter 10 Review — Skills

## Decisions Record

### System prompt is a constitution
Rendered once at agent creation, never mutated. Dynamic skill loads flow as
events in message history. Anthropic prompt cache prefix is never invalidated.

### Tool provenance: Initial vs Dynamic
Each tool entry carries Source (Initial/Dynamic) and EventSeq. Most renderers
ignore this. The Anthropic renderer uses it to avoid re-sending initial tool
declarations when new tools are added dynamically.

### Variable substitution via registered callbacks
`RegisterVar("NAME", func() string)` — called at render time. Built-in:
`$TOOLS` (current tool list), `$SKILLS` (loadable skills). Application-specific
renderers registered before `NewAgent()`.

### Unload is lazy
`unload_skill` marks the skill `pending-unload`. No cache miss. The skill's
tools remain callable until context compaction removes them.

### Ensemble skill declares all 12 core tools
The `ensemble` primary skill is the production configuration — all tools
visible from agent creation. The `base` skill is the teaching configuration
for progressive disclosure exercises.

## Grader: 9 checks, 100 points

| Check | Points | Tests |
|---|---|---|
| initial-tools | 10 | Only base skill's tools + load/unload in first request |
| system-prompt | 10 | System prompt contains primary skill's rendered body |
| ensemble-tools | 10 | With ensemble skill, all 12 core tools visible |
| load-skill | 15 | load_skill("code-tools") adds edit_file + write_file |
| progressive-disclosure | 15 | Loading code-tools makes search-tools loadable |
| depends-autoload | 10 | search-tools auto-loads search-helpers dependency |
| var-substitution | 10 | $CUSTOM_VAR rendered in skill body |
| blocked-skill | 5 | Non-loadable skill rejected with error |
| ch9-parity | 15 | All ch9 checks still pass |

## P9 Deletion Audit: 4 mutants, all caught

| Mutation | Failing checks |
|---|---|
| IsToolEnabled always true | {initial-tools, load-skill} |
| Delete dependency resolution | {depends-autoload} |
| Render returns text unchanged | {var-substitution} |
| IsLoadable always true | {blocked-skill} |

## Bill's procedure innovation (this chapter)
Bill never reports bugs directly. Instead, he suggests TL;DR improvements.
The grader is upgraded to catch the concern, and a student-coder sub-agent
verifies the chapter text is sufficient. Bug fixes are local; specification
improvements propagate through regeneration.

## Open for Bill
- Chapter title: currently "Skills" — final?
- Ch11: MCP tools (the LLM gets eyes on the GUI)
