# Chapter 9 — Skills

A tool needs both permission and an explanation. Registering a file writer gives the runtime a handler; it does not tell the model when to use it. A paragraph saying “write the file” teaches an intention; it does not grant execution. A skill joins these two facts: a rendered manual and a named set of capabilities, activated together and recorded as such.

The primary manual establishes the Agent’s working rules and tells it which additional manuals it may request. Supporting dependencies load before the manual that needs them. This keeps the initial instruction set manageable without reducing later capabilities to unexplained function signatures.

## The record says what was taught

Each Agent owns a concrete Skills manager. Common declares its relationships, persistent fields, immutable skill values, and required partial operations. The functional Skills files implement catalog parsing, resolution, variable substitution, and activation; Agent implements the attachment relationship. No Go-style package interface or runtime adapter is needed to connect these parts.

The manager keeps catalog definitions and registered variables. History owns the activation facts. Active manuals and grants are derived from skill parts in reduced Context rather than duplicated in a mutable enabled-tools set. Engine renders from those same facts, and Jobs checks them before reserving work or causing effects.

That common source prevents two unpleasant states: a model being shown a tool that dispatch considers disabled, and dispatch allowing a tool whose manual never entered the conversation. Registered handlers remain the Agent’s catalog, not automatic permission in a skill-configured Agent. Library callers that do not configure skills retain the earlier explicit-installation behavior.

A tool granted by several manuals still has one effective permission. Its provenance is the earliest recorded activation that grants it, including that activation’s sequence. The answer to “why can this run?” is therefore a conversation fact, not whichever dictionary happened to be inspected last.

## A small catalog, a complete activation

The host supplies one skill directory. The manager recursively reads `SKILL.md` files in sorted order and rejects duplicate names; there is no hidden home-directory fallback or precedence contest between roots. The front matter deliberately supports a small documented subset rather than general YAML. Predictable errors are more valuable here than accepting an ambiguous configuration language by accident.

Discovery is narrower than “every file found.” Only active manuals’ explicit offers make loadable skills available. A dependency is not independently requestable merely because it sits in the same directory. Loading resolves the entire dependency graph, detects cycles, checks types and handler names, and refuses the operation if any required piece is invalid.

Validation completes before mutation. Newly needed dependencies and the requested manual are rendered, then recorded in one dependency-first activation event. Publication succeeds as a whole or not at all. A repeated load adds no duplicate manual. A broken dependency must not leave half its closure active.

Construction applies the same discipline to the primary. If activation fails after the manager attaches and installs its controls, those additions are removed and the relationship is detached. Destruction clears both sides of the relationship without destroying the parent. These are ordinary ownership rules doing useful work, not a special plugin framework.

## Render once means remember once

Variables are substituted when a manual activates. The resulting text is the remembered manual. Updating a registered variable later affects future loads, not what the primary or an already loaded manual said. Editing or deleting the source Markdown likewise cannot rewrite a recorded activation during replay.

The primary is rendered first and stays fixed. Later manuals enlarge the complete instruction block without mutating that first body. This matters for built-in substitutions as well as custom values: a rendered tool list is a snapshot, not a live getter hidden inside prose.

Variable values come from the host. Core does not inspect the environment. Substitution is deliberately simple: names match case-sensitively, unknown references remain literal, and inserted text is not scanned again as another template. Neither recursive expansion nor shell syntax is smuggled into an instructional Markdown file. The host still has to choose supplied values carefully; a value intentionally placed in a manual will reach the model.

## A manual belongs in instruction context

A successful load result reports that loading succeeded. It is not the delivery mechanism for the manual itself. Anthropic receives manuals in its top-level system field, OpenAI Chat Completions in the initial system message, and Gemini in `systemInstruction.parts`. The primary comes first, followed by other active rendered bodies, each once.

Skill entries are excluded from ordinary dialogue projection. Burying the manual only inside tool-result prose would give it the wrong role and make its survival depend on ordinary result retention. Clearing a verbose load result must not erase the manual or revoke its grant. Nor should signed provider content be rewritten to carry new instructions.

The actor captures an immutable Context before rendering a request. Declarations use that snapshot’s grants, rather than consulting a live set halfway through serialization. Dispatch checks recorded grants again before execution. A model that guesses a disabled tool name receives a paired error using the actual call ID; it does not create a job or file.

The local skill controls run synchronously at the actor’s admission point, recording their call and result without an independent Job. The next model request sees the new instructions and capabilities. Skill loading has not invented another scheduler.

## Unloading is intentionally unfinished here

This chapter marks a non-primary manual pending removal. Its instructions and tools remain active. The primary cannot be unloaded, and repeated pending requests do not add redundant state.

That behavior avoids claiming to solve a harder context problem early: abruptly removing tools while the retained conversation still describes their availability. Compaction-time instruction removal belongs to later work. The result text must say what actually happened, rather than presenting a pending marker as immediate revocation. Later MCP support will separately close live remote connections on unload; retaining instructions and retaining executable resources are not the same lifetime.

## Deferred schemas do not defer authority

Portable rendering sends full declarations for enabled tools. An explicit Anthropic-only option supports tool search for a model the host knows can handle it. The host makes that capability assertion; a model-name heuristic cannot prove provider support.

Initial tools stay eager. Tools reachable through the primary’s catalog but granted only later can have deferred declarations. A successful load records the newly granted names as result metadata, which Anthropic renders as tool references in the correctly paired loader result. Searching prose for names would lose that structure during replay or redaction.

The deferred schemas still travel in every request. This saves model context, not outbound bytes. More importantly, appearing in the catalog never grants dispatch. Deferral changes how the model discovers a schema; History still establishes whether the Agent may use it. Without a live catalog during offline rendering, the safe representation is the ordinary enabled declarations, not an invented reconstruction of missing manuals.

Source: [Chapter 9](../chapter-09.md).
