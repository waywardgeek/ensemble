# Chapter 9: Skills

Chapter 8 follows a preference from its saved value to the behavior it actually controls. Skills pose a similar question at a larger scale: when an Agent says it has learned a capability, what has changed?

A new paragraph in the conversation is not necessarily a new instruction. A newly advertised tool is not necessarily executable. A hidden tool may still execute if dispatch never checks whether it is enabled. And a successful call after unloading may prove nothing about the capability supposedly being unloaded.

The [source chapter](../edition-3/chapter-09.md) contains that last trap in miniature. A historical OpenAI session loaded `code-tools`, wrote a scratch file, marked the skill for unload, and then successfully read the file. It looked like evidence that pending unload preserved the acquired tools. But `read_file` had belonged to the primary skill all along. The demonstration had carefully tested the lifetime of the wrong tool.

Skills make these distinctions explicit. Each skill packages instructions, tools, and dependencies in a named Markdown file. The Agent can acquire a new manual and the means to act on it without becoming a different Agent. The interesting work is keeping those two halves together, through dependency resolution, prompt rendering, provider protocols, and execution.

## A Small File, Several Different Powers

The workbench now has one Agent worth using. Adding every future capability directly to its system prompt would turn that prompt into a communal cupboard: identity, operating instructions, project facts, tool descriptions, and preferences all squeezed together. Any edit could disturb the supposedly stable prefix, and the Agent would receive material unrelated to its current task.

A skill supplies a smaller unit of composition. Its file lives at `skills/<directory>/SKILL.md`, with frontmatter followed by an ordinary Markdown body. The frontmatter gives it a name, a description, a type, and lists of tools, dependencies, and further skills it offers. The body supplies its instructions. This is packaging for existing capabilities, not a way to manufacture handlers by naming them. The source corrected an earlier example that advertised an undefined `think` tool. A name in a manual cannot implement a function.

The three skill types have different jobs. A `primary` skill defines the Agent’s initial identity and loads at creation. A `loadable` skill can be offered for later acquisition. A `dependency` loads automatically when needed and stays out of the Agent’s discovery list. Both the primary and its startup dependencies contribute instructions and tools. When skills are configured, `load_skill` and `unload_skill` remain available so the Agent can manage acquisition even if the primary declares only a reader.

The file format is deliberately smaller than general YAML. Scalar fields and two list forms suffice: space-separated names on one line or an indented list. The parser requires an opening frontmatter boundary, a name, and a valid type, and reports malformed files with their paths. It splits only at the frontmatter boundaries, leaving later Markdown horizontal rules in the body. Otherwise, an innocent separator halfway through a manual could become a configuration event.

The directory locates a definition; the declared name identifies it in dependency lists and load requests. Duplicate names are errors. Letting directory traversal order choose which definition wins would make the Agent’s identity depend on an accident of discovery.

None of this discovery grants capability. Finding a file means the registry knows a definition exists. Offering it means the Agent may ask to load it. Loading it means its instructions and tools become active. Those are separate states, even when a small example makes them happen close together.

## A Load Must Succeed as a Whole

Consider the source’s small scratch-file example. The primary `base` skill enables `read_file` and offers `code-tools`. Loading `code-tools` adds `write_file` and requires a hidden `helper` dependency, which adds `list_directory` and instructions to stay in the scratch directory. The next model request must carry both manuals and the newly enabled tools. The dependency’s instructions are not optional decoration attached to its function names.

The registry first resolves the entire dependency graph. A visiting set detects a cycle such as A depending on B depending on A. A resolved order lets dependencies activate before the skill that needs them. An unavailable requested name, wrong type, missing dependency, or cycle rejects the load without partially activating the graph.

The order matters more than the recursion. Suppose the first dependency grants a writer and the second is missing. Enabling as traversal proceeds would return an error while leaving the writer available. The caller asked for one capability and received a failed request with a useful side effect. Resolving first and activating afterward keeps failure intelligible.

After activation, the available-skills list is recomputed from the `loadable-skills` lists of all loaded skills. Only discovered definitions of type `loadable` qualify, and already loaded skills disappear from that list. A dependency does not become directly loadable because another file happens to offer its name.

This produces progressive disclosure. `code-tools` can offer `refactor-tools`, which can offer something more specialized. The Agent sees the next choices made available by its current manuals, rather than the whole directory as a menu. That does not require the definitions to form a literal tree: dependencies and shared offers can make a graph. The useful boundary is reachability from what has already loaded.

Progressive disclosure reduces what the Agent must consider, but it is not itself an execution guard. The enabled-tool set must govern both provider declarations and local dispatch. If a hidden `write_file` handler remains callable through the installed-handler table, the model can bypass disclosure simply by producing the name. The ordinary response to that guessed call is a tool error, with no write. After the skill loads, the same installed handler can execute.

This boundary controls admission to the tool. It does not turn the helper’s scratch-directory prose into a filesystem sandbox. Instructions describe how the Agent should use a capability; dispatch determines whether that capability is available at all.

## One Agent Owns the Registry

Several Agents can read the same skills directory without sharing a mutable skill state. Loading a writer into one must not grant it to another. The primary choice, enabled tools, offered skills, rendered bodies, and pending removals belong to the particular Agent.

That ownership does not justify importing the skills implementation directly into the Agent package. The existing star-shaped package structure still applies. Common holds the lifecycle interface and shared types; the skills package owns parsing, traversal, and rendering; the composition root constructs the registry and injects it into the Agent. The registry also receives its required Agent parent through the common interface. Both sides are attached before the primary loads, because rendering can consult the owning Agent’s declarations.

The application supplies construction inputs. The CLI and optional GUI translate `EN_SKILLS_DIR` and `EN_PRIMARY_SKILL` into a per-Agent skill specification. The registry does not discover its identity by consulting process environment whenever it needs an answer. An embedding application can therefore create different Agents in one process without a process-wide primary rewriting their identities.

Application variables follow the same boundary. The command’s `EN_CUSTOM_VAR` is one example of translating configuration into a rendering callback, not a reason to give the registry environmental authority. Omitting the optional skill specification preserves the earlier public tool-registration behavior. Skills are a way to assemble an Agent, not a forced migration for every existing consumer.

## Render Once, at the Relevant Time

The primary and other startup bodies form the initial system prompt. Their text is rendered at creation and remains intact. Later loads render new bodies and add them to that instruction context. The initial identity survives; the collection of manuals grows.

`$TOOLS` expands from the owning Agent’s currently enabled declarations. `$SKILLS` expands from its currently loadable set. Application variables use registered callbacks. Constructor-supplied variables must be installed before primary rendering, or the first and most consequential body would miss them.

A dynamic body renders after activation, so it can describe the tools just acquired and the next skills just revealed. The original primary’s `$SKILLS` expansion does not get rewritten to match. It records the choices at creation; the new manual supplies the later view. This is an intentional sequence of rendered instructions, not a live template silently rewriting every earlier paragraph.

The same distinction applies to application context. If `$CUSTOM_VAR` renders as a creation-time value in the primary, replacing its callback changes future renderings, not that existing text. In the accepted implementation described by the source, unknown placeholders remain verbatim. A visible `$CUSTOM_VAR` is easier to diagnose than an unexplained hole in a sentence.

Rendering callbacks run outside the registry lock. This detail is easy to lose in an otherwise tidy design: `$TOOLS` asks for declarations, and declarations consult the same enabled-tool state. Calling the callback while holding that lock can deadlock the first manual useful enough to mention its own tools.

Substitution also moves data across a privacy boundary. Whatever a callback returns becomes model-visible instruction text. A convenient environment variable is not automatically an appropriate variable to render, particularly when it contains a credential.

Skills do not make every piece of context an instruction. Memories, project notes, and changing user-preference data belong in message history as data, separate from the system’s operating instructions. This chapter adds neither memory storage nor a memory-saving tool. Keeping that distinction now avoids building a constitution out of yesterday’s project notes.

## Instructions and Tool Disclosure Travel Separately

The simplest delivery path puts all startup instructions in the system prompt and startup tools in the provider’s normal tool declarations. A later load adds rendered instructions to the system prompt and updates the ordinary declarations. That is the source chapter’s policy for providers other than Anthropic, and the fallback for Anthropic cases that cannot use incremental disclosure.

It costs prefix stability. Changing SYSTEM can invalidate a cache, and changing tool declarations can do so too. Preserving an old prefix by putting new operating instructions only in a tool result would preserve the wrong thing. A tool result can report that loading succeeded; it is not a substitute for delivering the loaded manual in the provider’s instruction context. The current source explicitly supersedes its earlier frozen-system design on this point.

Anthropic’s accepted path in the source separates tool discovery from instruction delivery more finely. Under the documented [tool-search protocol](https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-search-tool) used there, the full tool catalog still travels in every request. Startup tools use ordinary declarations. Undisclosed tools carry `defer_loading`. A successful model-called `load_skill` reveals newly acquired tools through `tool_reference` blocks in that call’s paired `tool_result`.

The wire result’s shape matters. The source records an accepted Edition 3 exercise in which mixing human-readable load status with `tool_reference` blocks caused HTTP 400. A corrected result containing references only was accepted, and the Agent then used the loaded writer. The helpful sentence was the protocol error.

Canonical history still records the textual load status. The provider renderer projects the successful discovery result into reference-only content for that provider. This preserves the application’s readable history without pretending that its canonical representation is always the wire representation.

The references disclose tools, not instructions. The documented API used by the source supplies no separate mid-turn instruction channel, so loaded manuals still join SYSTEM. Keeping the catalog stable can avoid unnecessary declaration changes; it cannot guarantee a cache hit when the instruction text grows.

There are limits to when this form is legitimate. A direct or manual load has no model call whose result can carry the references. A load without a paired discovery result has the same problem, and an unsupported model cannot use the protocol merely because the application prefers it. These paths use ordinary enabled tool definitions. The renderer must not invent a model call or enabling event to make the incremental representation fit.

These are the provider mechanisms and observed correction recorded by the current source, not a fresh claim about every model’s present support. The distinction is especially useful here because the source also preserves earlier live Claude, OpenAI, and Gemini sessions. Those sessions loaded the example graph, wrote and read their markers, marked a skill pending unload, and exited normally. Their captured system text stayed identical within each session. They establish behavior of the earlier frozen-system design, not verification of the instruction delivery now required.

## Provenance Makes the Projection Possible

The provider renderer needs to know whether a tool belongs in the startup declarations or was enabled by a particular discovery call. It does not need to know how a skills directory is organized.

Each enabled tool therefore records its first enabling source and event sequence. Initial tools have sequence zero. A dynamically acquired tool records the existing `tool_called` sequence of the actual `load_skill` call that enabled it. There is no separate skills counter to reconcile with conversation history.

First means first. If the primary already supplies `read_file`, another skill mentioning it does not reclassify the reader as newly acquired. It retains its initial provenance. Otherwise, the renderer could try to disclose an old tool through a new event, or later lifetime reasoning could mistake a shared capability for one introduced by the latest skill.

This metadata lets the Anthropic renderer pair references with the successful result that actually enabled them, without importing the registry implementation. It also distinguishes a manual load, which has no enabling model call and uses ordinary declarations. A stable event identity connects the application’s capability transition to the provider’s conversation structure.

## Pending Unload Means Still Loaded

Loading has an effective transition now. Unloading does not yet have its removal boundary.

`unload_skill` accepts a loaded name, marks it pending, and reports that state. An unloaded name is an error. Pending unload retains both instructions and tools. It does not rewrite prior manuals or disable their handlers. Actual removal is deferred until the later compaction boundary; this chapter records the intent without implementing that compactor.

That is not merely reluctance to delete a map entry. The model has already received instructions and, potentially, disclosed tool definitions. The application needs a coherent point at which its capability set and the model’s context can agree about removal. Anthropic’s context-editing facility can clear selected tool results, but that does not semantically retract a manual or revoke an application capability. Deleting one piece of conversation content is not a general unlearning operation.

The historical unload demonstration at the start now has a precise diagnosis. Reading after pending unload exercised a startup capability, so it could not reveal premature removal of the dynamic writer. The source’s focused test instead checked that `write_file` remained declared after pending unload. Declaration and dispatch use the same admission predicate, and restoring early removal made that check fail. The revised grader also exercised a dynamic write after marking unload. Those checks concern the capability whose lifetime is actually changing.

The Agent can now expand its working instructions and executable tool set without replacing its initial identity. Its registry has one owner, loads are whole-graph transitions, provider projections have real event anchors, and pending unload says exactly what has not happened yet. None of this saves the conversation across process death. The next chapter takes up that separate lifetime: allowing the conversation to outlive the process that created it.
