# Chapter 9: Skills

Every chapter so far has added capability by adding code. That stops
scaling. The agent already has file tools, process tools and job controls,
and the system prompt can grow with every feature. A configurable primary
skill was deliberately deferred until now. The system prompt is an easily abused
feature in LLM applications: teams stuff instructions, context, tool
descriptions, and user preferences into a monolith whose supposedly stable
prefix changes on every edit. Skills fix this before the problem arrives.

A skill is a named unit of capability: instructions, tools, and
dependencies, in a single markdown file. Loading one changes what the
agent can do; unloading one marks it for removal at the next
compaction. This chapter records that pending state; compaction arrives later.
The system prompt renders once from the primary skill at
agent creation and stays fixed. Capabilities added after creation
reach the agent through the event stream.

Variable substitution lives here too. A skill body can contain
`$TOOLS` or `$SKILLS` or any application-specific variable, and the
renderer replaces them at the point of use. The initial system prompt
gets creation-time values; a dynamically loaded skill gets the values
current at load time.

## TL;DR

Extend Chapter 8 with a per-Agent skills registry. A SKILL.md file lives in
a named directory under the skills root:

```
skills/
  base/
    SKILL.md
  code-tools/
    SKILL.md
  search-tools/
    SKILL.md
  read-tools/
    SKILL.md
  refactor-tools/
    SKILL.md
```

Each SKILL.md has flat YAML-style frontmatter between `---` markers and a
markdown body. This example requires the named dependency and offered skill
to have their own files:

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

The primary skill defines the agent. Its body, together with its initial
dependencies' instructions, becomes the system prompt, rendered once at
creation. Its `tools` list and those dependencies determine startup capability;
`load_skill` and `unload_skill` are always available when skills are configured.
A minimal `base` skill can declare only `read_file`. A full `ensemble` skill
declares the tools built so far:

```yaml
---
name: ensemble
description: Full AI coding agent
type: primary
tools:
  - read_file
  - write_file
  - edit_file
  - list_directory
  - search_files
  - run_command
  - wait_for_job
  - send_input
  - kill_job
  - tool_limits
  - load_skill
  - unload_skill
loadable-skills: code-tools search-tools
---
You are an autonomous AI coding agent.

## Available Tools

$TOOLS

## Available Skills

$SKILLS
```

The original example named `think` without teaching its behavior or arguments.
An unknown name cannot create a handler. This list uses the existing
`tool_limits` capability instead; do not advertise a tool that cannot execute.

The exercise is:

1. Discover the files and validate each skill's name and type. Accept both
   list forms shown below and preserve horizontal rules in the body.
2. Load the selected primary and its dependencies at Agent creation. Render
   the system text once; later skill loads leave it byte-for-byte unchanged.
3. Implement `load_skill` and `unload_skill`, each taking a JSON object such
   as `{"name":"code-tools"}`. A successful load resolves dependencies,
   enables tools, reveals the next available skills and returns rendered
   instructions through the existing tool result. Reject unavailable names,
   wrong types, missing dependencies and cycles without a partial activation.
4. Use the same enabled set for vendor declarations and dispatch. A guessed
   hidden name must not execute. Mark unload pending while retaining its
   instructions and tools; actual removal waits for the later compaction chapter.
5. Support `$TOOLS`, `$SKILLS` and application variables at point of use.
   Track each tool's first enabling source and event sequence. Keep registries
   independent between Agents and preserve earlier behavior without a skill spec.

From the core module, build the headless command. With ordinary provider
settings already configured, the public launch is:

```sh
go build -o /tmp/ensemble-ch09 .
EN_SKILLS_DIR=path/to/skills EN_PRIMARY_SKILL=base /tmp/ensemble-ch09 prompt "load the code-tools skill"
```

Use `chat` in place of `prompt TEXT` for the human interface. Both CLI and
optional GUI translate the same skill environment values into per-Agent
construction inputs. The GUI keeps Chapter 7's separate module and launch.
`EN_CUSTOM_VAR` supplies `$CUSTOM_VAR` for these commands; it is an application
example rather than a reason for the registry to read process environment.

From the repository root, grade the submitted core and its optional GUI:

```sh
go run ./cmd/grade-edition2 -ch 9 path/to/core
```

`-gui DIR` selects the GUI module, defaulting to `gui/` under the core;
`-gui-model` and inherited consumer, workflow, media and streaming fixture
options remain available. The skills phases use the actual WebSocket prompt
path. Headless parity uses the headless command.

| Check | Points | Observable behavior |
| --- | ---: | --- |
| Initial tools | 10 | Primary/dependency tools plus both lifecycle tools |
| Primary body | 10 | Rendered initial instructions stay fixed across loads |
| Full primary | 10 | All twelve installed tools in the example are declared |
| Dynamic tools | 15 | Load adds instructions and executable tools; pending unload retains capability |
| Progressive disclosure | 15 | A load reveals its offered skills in current `$SKILLS` |
| Dependencies | 10 | Dependencies contribute both tools and instructions |
| Variables | 10 | Placeholder values reflect their rendering time |
| Unavailable skill | 5 | An unreachable skill is refused |
| Chapter 8 parity | 15 | Prior checks still pass through their appropriate clients |

The numeric checks do not prove every obligation. Dependency cycles, failed-load
atomicity, provenance and Agent isolation also need focused source tests and
review. Exercise the real registry and tool loop, replacing only the external
provider with a small fake where deterministic replies are needed.

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

The supported format is deliberately small: scalar fields and these two list
forms. There is no need for a general YAML document processor. Require the
opening frontmatter boundary, a name and one of the three types; report a
malformed file with its path. The directory locates the file, while `name`
identifies the skill in `depends` and tool arguments. Reject duplicate names
rather than letting directory order choose an identity.

Three skill types control visibility. A `primary` skill loads at
agent creation and defines the agent's identity. An offered `loadable` skill
appears in the available-skills list. The initial list is rendered into the
system prompt; later lists arrive with loaded instructions. The agent
calls `load_skill` to activate it. A `dependency` skill is invisible
to the agent; it loads automatically when a skill that depends on it
loads.

## The Registry

`SkillRegistry` discovers skills from a directory (one subdirectory
per skill) and tracks their state through a lifecycle. The name describes a
responsibility; it need not be an exported concrete type:

```
Discovered → LoadInitial (at creation)
           → LoadDynamic (by load_skill tool)
           → PendingUnload (by unload_skill; removal is a later chapter)
```

Loading resolves dependencies recursively. If `search-tools` depends
on `search-helpers`, both load and both contribute their tools. Cycle
detection uses a visiting set. A depends B depends A produces an
error, not infinite recursion.

Resolve the whole graph before enabling any of it. If the first dependency
supplies `write_file` and the second is missing, an error accompanied by a newly
enabled writer is a partial success the caller never asked for. A visiting set
and a resolved order let the registry validate first, then activate dependencies
before the requested skill. Their instructions travel with their tools.

Discovery reads definitions; it grants no capability. The registry belongs to
one Agent even when several Agents read the same directory. Its required parent
is that Agent, supplied through the common interface; the composition root
attaches both sides before loading the primary. Keep parsing, traversal and
rendering in the skills package. Shared types and interfaces remain in common.

The command's environment variables are construction convenience. An embedding
application passes its skills directory and primary choice to the Agent it is
creating. Changing a process-wide primary must not change every live Agent's
identity. Omitting the optional skill specification preserves the earlier
public tool-registration path.

The registry answers two questions the agent asks constantly: "what
tools can I use?" and "what skills can I load
next?" (`LoadableSkills`). The loadable set is the union of all
loaded skills' `loadable-skills` lists, restricted to discovered `loadable`
skills and minus anything already loaded. A dependency does not become directly
loadable merely because someone puts its name in that list.

The enabled-tool answer must reach execution as well as the request renderer.
Hiding `write_file` from the declaration array still leaves a working writer if
dispatch consults only the installed-handler table. Test the guessed name before
loading its skill: it should return an ordinary tool error and leave the file
absent. Then load the skill and prove that the same real handler can write it.
This checks both sides of the boundary without replacing the tool with a mock.

`unload_skill` marks a loaded name pending and reports that fact. It does not
rewrite old instructions or disable the skill's tools in this exercise. An
unloaded name returns an error. Compaction later provides the boundary where
removal becomes effective; adding a compactor now would turn a lifecycle marker
into another assignment.

## Progressive Disclosure

Loading a skill reveals more skills. The `$SKILLS` variable updates
to show what the agent can load next: `code-tools` might reveal
`refactor-tools`, which in turn reveals `architecture-tools`. The
tree unfolds as the agent explores.

The agent sees only one level ahead: the skills made available by
the ones already loaded. Context stays focused and the tool list
stays manageable.

The original `$SKILLS` text in the system prompt does not change as this tree
unfolds. The newly loaded body's `$SKILLS` is rendered after activation and
arrives in the tool result, where it can name the next level. This is how the
agent discovers `refactor-tools` without rebuilding its initial instructions.

## Variable Substitution

Skill bodies contain `$VAR` placeholders. A small map of names to renderer
functions is enough. The skills package supplies `$TOOLS` and `$SKILLS` from
the owning Agent's enabled declarations and current loadable set. Application
variables enter through the public API. For example, given the existing
Ensemble root `app` and its provider configuration `cfg`:

```go
agent, err := app.NewAgent(cfg, ensemble.AgentSpec{
    SkillsDir: "skills",
    PrimarySkill: "base",
    Variables: map[string]func() string{
        "CUSTOM_VAR": func() string { return "creation value" },
    },
})
if err != nil {
    return err
}
// Future loads see the new renderer; the primary text is already frozen.
agent.RegisterVar("CUSTOM_VAR", func() string { return "load-time value" })
```

The constructor's variables must be installed before primary rendering.
`RegisterVar` changes future renderings. It cannot reach back into text already
sent to a model. Select non-secret settings or context deliberately; substituting
a value puts it into model-visible instructions, so an environment variable
holding a credential is not a suitable example.

Variables render at point of use: the initial system prompt gets
creation-time lists, and a skill loaded mid-conversation gets
whatever the lists contain at load time.

In the accepted implementation, unknown placeholders remain visible verbatim.
A missing `$CUSTOM_VAR` is easier to diagnose that way than as a silently empty
sentence. Render callbacks outside the registry lock: `$TOOLS` reads declarations,
which consult the same enabled-tool state. Holding the lock while calling it
would deadlock the first useful variable.

## Tool Provenance

Every tool in the registry tracks its source:

```go
// ToolMeta records when this Agent first gained a tool.
type ToolMeta struct {
    // Source distinguishes creation-time and later capabilities.
    Source ToolSource
    // EventSeq is zero for initial tools, otherwise the enabling tool_called.
    EventSeq uint64
}
```

If the primary already enabled `read_file`, loading another skill that also
names it must preserve its initial provenance. A newly enabled `write_file`
records the sequence of the actual `load_skill` call that enabled it. Use the
existing event identity rather than inventing a separate skill counter.

Most renderers ignore this metadata at this stage. It gives a later renderer
enough information to distinguish initial declarations from dynamic additions
without knowing about skills. That matters when arranging a stable request
prefix for caching. The metadata alone does not preserve a cache hit: this
chapter still changes the tool declarations, and provider-specific placement,
cache markers and measured grants belong to the caching work later in the book.

## The Constitution

When `load_skill` runs, the skill's instructions arrive as the tool
result. They flow through the message history, not the system prompt.
Capabilities grow while the system prompt stays frozen at its
creation-time content.

That separation matters for memory. Pre-loaded context (memories,
user preferences, project notes) belongs in the message history as
data rather than the system prompt as instructions. The system prompt stays
separate from those changing facts. Chapter 22 builds the memory cascade on
this foundation; this chapter adds no `save_memory` tool or memory storage.

## Taking it for a spin

Use a scratch directory so a successful tool call changes a file you meant to
change. For a small demonstration, create these three files under its `skills`
directory. They form a separate example from the larger skill tree above.

`base/SKILL.md`:

```markdown
---
name: base
description: Small coding assistant
type: primary
tools: read_file
loadable-skills: code-tools
---
You are a concise coding assistant. Available tools: $TOOLS.
Skills you can load: $SKILLS.
```

`code-tools/SKILL.md`:

```markdown
---
name: code-tools
type: loadable
description: Write and verify a scratch file
tools: write_file
depends: helper
---
Write note.txt with content $CUSTOM_VAR using write_file, then use
read_file to verify it. Next available skills: $SKILLS.
```

`helper/SKILL.md`:

```markdown
---
name: helper
description: Scratch directory guidance
type: dependency
tools: list_directory
---
Work only in the current scratch directory.
```

Launch the built command from that scratch directory, with your chosen model
and provider credentials already configured in its environment:

```sh
EN_SKILLS_DIR=./skills EN_PRIMARY_SKILL=base EN_CUSTOM_VAR=SKILLS_OK /absolute/path/to/ensemble chat
```

Ask it to load `code-tools`, follow its instructions once, mark it for unload,
and answer with the verified file content. Inspect `note.txt` as well as the
answer. The load result should contain the helper's scratch-directory instruction
and the substituted marker; the next request should declare `write_file` and
`list_directory`. The system instructions should still be those rendered before
the first turn. Keep credentials out of any retained request capture.

The October 10, 2026 student sessions used discovered `claude-haiku-5-5`,
`gpt-4o-mini` and `gemini-3.8-flash` models. Each loaded this small skill graph
(the published examples add descriptions and wrap the body text),
wrote and read its own marker, marked `code-tools` pending unload and answered
with the file content. Each process exited normally. Their captured requests
retain identical system text within each session: five requests for the Claude
run, four for the OpenAI run and five for Gemini. These are historical run
identities; discover the models available to your account when repeating the
exercise. The [Claude](../solutions/evidence/ch10/anthropic-live.json),
[OpenAI](../solutions/evidence/ch10/openai-live.json) and
[Gemini](../solutions/evidence/ch10/gemini-live.json) receipts contain
the actual calls, results, file content and request bodies.

One successful run invited the wrong conclusion. The OpenAI session called
`unload_skill` before its final `read_file`, and the read succeeded. That looked
like proof that pending unload had retained the dynamic tools. But `read_file`
belonged to `base` all along. The read would have succeeded even if unload had
incorrectly disabled everything `code-tools` added.

The focused source test instead checks that the dynamically added `write_file`
remains declared after pending unload. Declaration and dispatch share the same
admission predicate, and restoring early removal makes that check fail. The
revised grader also exercises a dynamic write after marking unload. Match the
tool in the assertion to the capability whose lifetime is in question; another
successful paid run of the same read would prove no more.

The historical full grader was interrupted during its inherited GUI/headless
parity cascade and produced no aggregate score. Separate unchanged Skills phases
scored **65/85**, losing only the two declaration checks that required undefined
`think`; this was not a completed 65/100 run. The
[independent review](../docs/edition-2-notes/ch10-code-review.md) preserves
those results and their accepted exceptions. Edition 2's grader uses the
implemented tool set and the proper clients, with its new results recorded
separately in the [grader report](../docs/edition-2-notes/enhancement-09-code.md).

The agent can now acquire a manual and its tools without rewriting its
constitution. The next chapter gives the conversation a life beyond the
process that created it.
