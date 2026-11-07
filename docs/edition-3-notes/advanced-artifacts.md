# Advanced Artifacts: working chapter notes

Status: design discussion with Bill; not an implemented migration or a finished
chapter. Develop the code and teaching together. Chapter numbering is undecided.
Foundations: [Chapter 7, Everything Is an Artifact](../../book/chapter-07.md),
especially §7.6, and [Chapter 26, Artifacts and Actions](../../book/chapter-26.md).
The earlier [Artifacts/Actions notes](artifacts-and-actions.md) retain their own
scope and evidence; this work does not reopen their grading campaign.

## Reader and user purpose

Users should be able to describe in words how they want output to look. Their
Agent should discover the relevant presentation code, understand its interface,
and adapt it. Bill wants to learn the design and implementation as we develop
this chapter, not receive a completed opaque redesign.

Artifact presentation is shared across Agents for now. It belongs to the tool or
output type, not the particular Agent that invoked it. Include tools supplied
through MCP connections in skills, not just hard-coded built-in tool names.

## Agreed direction: colocate renderers with skills

Move all JavaScript Artifact renderers into the skills directory structure.
Renderers for tools belong alongside the skill that declares/provides those
tools, including built-ins. Chat and readable thinking are Artifacts too; their
renderer and the default renderer need suitable skill-package homes, still to be
chosen. Do not interpret “all Artifact renderers” as a decision to distribute
GUI connection management or authoritative conversation state across skills.

A missing tool-specific renderer uses the existing default Artifact presentation.
The default itself should be readable and customizable. Preserve raw received
content and its provenance independently of presentation, with an accessible
fallback when a custom renderer cannot display a result.

Ship readable sources and usage documentation with Ensemble so customization
works without a repository checkout. This establishes a distribution requirement;
**building an installer is not a chapter in this work**.

## Shared libraries across skill packages

Provide a shared library structure in or adjacent to the skills directory.
Its exact name, location and import/package conventions are not yet decided.
Shared code is not limited to browser JavaScript: Python, Go, and other systems
used to provide tools should be reusable by multiple skills too. Sharing an
installation layout does not require one cross-language loader or universal
execution interface.

The Markdown renderer is the first concrete example. A custom renderer should
reuse the shared safe Markdown helper, not bundle another parser or reproduce
its policy. Keep library implementation, concise API documentation, and a small
usage example available as distinct resources.

### Progressive discovery for the model

Use the existing skill-discovery pattern rather than putting every library's
manual and implementation into every model request. A short artifact-authoring
manual or library index can identify available helpers and where to learn more.
The Agent reads a helper's API/example when needed, and its implementation when
changing or debugging that helper. Exact authoring-skill/index structure remains
to be designed.

Distinguish three operations:
- Discover a library and its purpose.
- Read instructions/source into model context when relevant.
- Load executable code in the browser or tool process.

A browser can use a shared Markdown library without sending that library's code
to the model. Browser lazy loading is a separate choice from model-side context
management. React is not required for this separation; no React migration or new
extension framework has been selected.

## Current code observations

Inspected in the current source, not newly live-tested for these notes:

- `agent/gui/wwwroot/app.js`: `cardFor(id)` retains stable identity, original text,
  kind, role and status. `render(card)` handles presentation and trusted controls.
- Text, thinking and reasoning-summary cards all call the same `markdown()` helper.
  This concerns provider-supplied readable text, not opaque reasoning metadata.
- `agent/gui/wwwroot/markdown.js` constructs one shared markdown-it parser. It
  disables raw HTML, restricts clickable links to HTTP(S), and renders Markdown
  images as escaped descriptions rather than fetching remote images.
- `queueCardRender()` coalesces streaming updates at 50 milliseconds; selection
  can defer repaint. The helper reparses the changed card's accumulated text and
  rebuilds its body while preserving selected controls/state where possible.
  This is not a fresh parser per Artifact, nor a guarantee of cheap rendering for
  unbounded output on every device.
- Tool results currently use expandable raw details, bounded previews and a few
  hard-coded cases such as applied edits. There is no skill-provided renderer
  interface established by these observations.
- The local boolean `artifact` inside `render()` currently means Tools-pane
  content. Its name is narrower than the book's “everything is an Artifact”
  model; do not propagate that confusion into the new interface.

## Connection to the Workflows chapter

Bill's intended definition: workflows are executables that use reverse MCP to
orchestrate sub-agents of the Agent that called the workflow. They can coordinate
writing, reviewing and editing, as in Bill's existing book-production workflows.
The shared skill/library layout supports these executable harnesses as well as
renderers and ordinary tools. A separate workflow scheduler or unrelated Agent
ownership system is not implied.

Inspect the actual reverse-MCP interfaces and caller/child ownership before
choosing a harness API. This records the direction, not a claim that an arbitrary
workflow executable is already supported or tested. Start the Workflows chapter
from one of Bill's real workflows rather than speculative architecture.

## Open design decisions and next step

- Choose explicit renderer declarations, dispatch identity and library references.
- Separate original content, the shared card/control structure, and replaceable
  content rendering. Preserve streaming, replay, copying, speech, selection,
  focus, scroll position and accessibility.
- Define authority/isolation for renderer JavaScript and tool executables. A
  presentational control must not silently gain tool execution or file authority.
- Define packaging, version/update behavior, unavailable-library and renderer
  failure fallback without introducing a second authoritative conversation store.
- Verify the existing skill packaging and asset-serving paths, then choose one
  built-in tool and shared Markdown helper for a bounded first experiment with Bill.

These decisions should be earned through that experiment. Do not turn the notes
into an expanding acceptance campaign or implement workflows merely because the
layout will support them.
