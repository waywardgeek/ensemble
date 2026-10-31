# Chapter 9 — Skills

## TL;DR

A skill is a small manual that grants a named set of tools. Each Agent owns a
concrete `Skills` manager. The manager validates a dependency closure, renders
its manuals, and records the entire activation in History before making it
visible. Request rendering and tool admission read the same recorded grants.
A tool that is not granted cannot execute, even if the model guesses its name.

The primary manual is rendered once at creation. Later manuals enter provider
**instruction context**, not ordinary tool-result prose. Loading never rewrites
the primary. Unloading in this chapter records a pending marker: instructions
and tools remain usable until a later chapter implements compaction-time removal.

The runnable endpoint is `solutions/ch09`. Its library, CLI, optional
GUI, and separate public-API example use C# 14/.NET 10 without downloaded packages.

## 9.1 In plain language

### A constitution and a shelf of manuals

The primary skill is the Agent's constitution. It says how to work and lists
which additional manuals the Agent may request. A loaded manual supplies both
instructions and the tools needed to follow them. Dependencies are supporting
manuals loaded automatically before the requested manual.

There are two different questions:

1. **Can this tool run?** Its name must be granted by an active skill.
2. **Has the model been told how to use it?** The corresponding rendered manual
   must be present in the request's instruction context.

Satisfying only one question is not enough. A prompt saying “you may write files”
does not enable a handler. Adding a handler does not teach the model the manual.
The recorded activation ties both facts together.

### Load once, record what was taught

Suppose a manual contains `$CUSTOM_VAR`. The manager substitutes the current
registered value when loading that manual, then keeps the resulting text. A
later variable update does not rewrite what the Agent was already taught.
The primary receives the same treatment at creation.

History stores these rendered snapshots. Replay therefore does not depend on
whether a Markdown file was subsequently edited, moved, or deleted. The files
are the catalog; History records this Agent's actual activations.

### Pending does not mean gone

Immediate revocation would leave a conversation describing tools that suddenly
cannot run. This chapter does not solve compaction. `unload_skill` therefore
marks a non-primary skill pending removal and says so honestly. The manual and
its tools remain active. Test this with a tool granted by the loaded skill, not
one that was already granted by the primary.

## 9.2 Ownership and the authoritative facts

Keep the existing hub-and-spoke organization:

- `Common/Skills.cs` declares `Skills`, immutable skill values, parent links,
  fields, and callable partial members.
- `Skills/Parser.cs` implements the documented front-matter subset and value
  construction. `Skills/Skills.cs` implements catalog resolution, variable
  rendering, activation, provenance, and pending removal.
- `Agent/Skills.cs` implements complete two-way attachment and removal. The
  manager reaches History and the handler catalog through its actual Agent
  parent, not copied service references.
- History owns activation and pending-removal events. Engine projects their
  immutable snapshots into provider requests. Jobs performs admission through
  the Agent before starting effects.
- CLI and GUI composition roots choose the skill directory, primary name, and
  environment-derived variables. The core library never reads the environment.

The relevant public vocabulary includes:

```csharp
// Value-like configuration: it does not own an Agent or need holder back-links.
public sealed record SkillSpec(
    string Directory,
    string Primary,
    IReadOnlyDictionary<string, string>? Variables = null);

// Value-like provenance: the earliest activation granting a tool.
public readonly record struct ToolSource(string Skill, long Sequence);
```

`SkillPart` is also value-like. It contains a skill's name, rendered body,
granted tool names, and whether it is primary. Its immutable arrays are not
another mutable capability database. There is **no parallel active-tool set**:
active manuals and grants are derived from the skill parts in History's Context.
Pending markers remain events; they do not remove those parts.

The manager retains only its catalog definitions, registered variables, and
lifecycle state. The Agent's ordinary `Tools()` collection remains the handler
catalog. In a skill-configured Agent, registration alone is not permission to
run. In a library Agent without skills, the earlier explicit tool-installation
behavior remains available.

### Complete relationship operations

`new Skills(agent, specification)` validates its catalog, attaches through
`agent.AddSkills`, installs the two control definitions, and activates the
primary closure. A failed activation removes the added controls and detaches
the manager. No half-created manager is returned.

`RemoveSkills` validates membership before changing either end and clears both
links before returning. `Skills.Destroy` is idempotent and removes that link;
it does not destroy its parent. Agent destruction first drains jobs while
History is available, then destroys its manager, tools, Engine, and History
through the existing protected cleanup path. A retained loader called without
a manager returns an explicit refusal, not a null-reference fault.

## 9.3 The skill file format

A host supplies one explicit directory. The manager reads `SKILL.md` files
recursively in sorted path order and rejects duplicate names. There is no
implicit home-directory fallback or precedence between competing roots.

```markdown
---
name: ensemble
description: The default careful coding assistant.
type: primary
tools: read_file list_directory search_files
loadable-skills: code-tools
---
Explain what you will do and why before using tools.
Inspect actual files and report failures honestly.

Enabled tools:
$TOOLS

Additional manuals:
$SKILLS
```

The parser trims front-matter lines, but deliberately does not implement general
YAML nesting. The supported fields are:

| Field | Meaning |
| --- | --- |
| `name` | Required unique identifier: letters, digits, `_`, or `-`. |
| `description` | Optional plain single-line text for discovery. |
| `type` | Required `primary`, `loadable`, or `dependency`. |
| `tools` | Tool names granted by this manual. |
| `depends` | Manuals activated before this one. |
| `loadable-skills` | Loadable manuals this manual offers explicitly. |

The three name-list fields accept either whitespace-separated names or a
YAML-style dash list:

```yaml
tools:
  - write_file
  - edit_file
  - run_command
```

This is a small documented subset, not a general YAML implementation. Unknown
fields, duplicate fields, malformed identifiers, and missing boundaries are
errors. Quoted scalars, flow arrays, and multiline YAML values are not part of
the contract. The first closing `---` ends front matter; later separators remain
literal body text. Empty lines and comment lines in front matter are ignored.

Catalog syntax is checked at construction. Activation then validates the chosen
closure: missing dependencies, cycles, wrong primary placement, invalid offered
names/types, or unknown handlers refuse the operation. A dependency is not
independently discoverable unless an active manual explicitly offers it as a
loadable skill; `dependency`-typed entries cannot be offered as loadable entries.
A loadable manual may offer further manuals, making discovery progressive.

Both `load_skill` and `unload_skill` are always enabled controls for a configured
primary, even if it offers no additional manuals. They are not ordinary jobs.

## 9.4 Variables and transactional activation

Built-in substitutions are:

| Variable | Snapshot value |
| --- | --- |
| `$TOOLS` | Sorted enabled tool names after the proposed activation. |
| `$SKILLS` | Currently offered, not-yet-active skill names and descriptions. |
| `$CUSTOM_VAR`, or another registered uppercase name | Host-supplied value at activation. |

Unknown variables remain visible. Substitution makes one pass: text supplied as
a variable value is not recursively interpreted. Variable names must match
`[A-Z][A-Z0-9_]*`; hosts cannot replace the built-ins `TOOLS` and `SKILLS`.

`RegisterVariable` updates the manager's explicit variable map under its lock.
It does not mutate existing skill bodies. This is configuration for the next
activation, not a new conversation event.

A successful load performs these steps on the Agent's actor:

1. Require that an active skill currently offers the requested name. An already
   active name is an idempotent success, including while pending unload.
2. Resolve dependencies first, checking the entire closure before publication.
3. Calculate the prospective active names, grants, and discoverable manuals.
4. Render all new bodies against that prospective snapshot.
5. Append **one** `SkillLoaded` event containing all new `SkillPart` values.
6. Return a concise successful tool result naming the loaded manuals.

History's append is durable before publication. A validation failure grants
nothing. A persistence failure propagates as a real failure through the actor;
it is not converted to a successful load or a harmless validation message.

The initial event contains the primary and its dependencies. Primary grants also
include the two controls. Later events contain only newly activated manuals.
`Source(tool)` walks these entries in activation order and reports the earliest
granting skill and sequence. A tool shared by two manuals has one effective
permission, not two mutable copies.

`unload_skill` refuses the primary, marks another active skill pending once,
and returns text explaining that its instructions and tools remain active.
This chapter adds no replacement primary, file watcher, reload command, or
compaction framework.

## 9.5 Instructions, dialogue, and dispatch

Loading a file is not enough. A successful activation must reach the outgoing
request in the correct instruction location:

| Adapter | Instruction location |
| --- | --- |
| Anthropic | Top-level `system`. |
| OpenAI Chat Completions | The initial `system` message. |
| Gemini | `systemInstruction.parts`. |

The primary body is emitted first, followed by the other rendered bodies.
Each occurs once. Skill entries are excluded from ordinary dialogue projection,
so a loaded body does not also appear as tool-result text. Ordinary human,
assistant, and tool exchanges retain their existing roles and correlation IDs.
Signed or opaque provider content is not rewritten to carry instructions.

The **primary body** stays fixed; the **complete instruction block** grows as
manuals are activated. Do not assert that the entire block remains byte-identical
after loading. The loader's result reports success or failure; it is not the
only carrier of the instructions. Clearing ordinary tool-result text must not
silently remove a manual or revoke a grant.

The actor renders from a captured immutable Context before starting HTTP.
Declarations use that snapshot's grants rather than consulting a second live
enabled set halfway through rendering. Tool dispatch checks the recorded active
grants before reservation or effects. A guessed disabled tool call receives a
paired error with its real call ID and cannot create a job or file.

The skill controls run synchronously at the actor's admission point. They still
consume one-shot observation options, validate arguments, and record call/result
facts, but allocate no independent job. The next model request sees the new
manual and permissions. All other scheduling, hint, interrupt, and reliable reply
behavior remains owned by the existing actor.

### Optional Anthropic deferred declarations

Portable mode sends full declarations for enabled tools. That is the default.
For a selected Anthropic model known to support tool search, a host may explicitly
set `Configuration(..., deferredTools: true)`. The executable equivalent is
`EN_DEFERRED_TOOLS=1`; other vendors refuse this option. The Boolean is the host's
capability assertion, not a model-name guess or proof of provider support.

In this mode:

- Initial tools, including the loader, remain eager.
- Tools reachable through the primary's catalog but granted only later are sent
  with `defer_loading: true`. A catalog declaration does not grant dispatch.
- A successful load records newly granted tool names as result metadata.
- Anthropic projects those names into the correctly paired loader result:

```json
{
  "type": "tool_result",
  "tool_use_id": "load-1",
  "content": [
    { "type": "tool_reference", "tool_name": "write_file" }
  ]
}
```

The declarations stay deferred after activation; references reveal their schemas
to the model. Other adapters and ordinary mode retain textual load results.
Result metadata survives event serialization, numbering, replay, and ordinary
result redaction. It is not reconstructed by searching response text.

Anthropic requires the deferred schemas in **every request**. Deferral saves model
context, not outbound wire bytes. The [official tool-search documentation](https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-search-tool)
describes this client-side custom-search path and its reference format. The local
tests verify our projection against that contract, not a real provider exchange.
With no live manager/catalog during offline rendering, use the portable enabled
declarations; do not invent a deferred catalog from missing files.

## 9.6 Constructing and running the C# applications

A library host supplies values directly:

```csharp
var specification = new SkillSpec(
    "/absolute/path/to/owned/skills",
    "ensemble",
    new Dictionary<string, string> { ["CUSTOM_VAR"] = "example" });

var configuration = new Configuration(
    Vendor.Anthropic,
    "selected-model",
    "https://api.anthropic.com",
    apiKey);

var agent = new Agent(app, configuration, skillSpec: specification);
```

The constructor installs the ordinary handler catalog before activating skills.
A host with custom handlers can instead construct an Agent, register its tools,
and then construct `Skills` before accepting input. If logging is required,
open the log **before** constructing the manager: its initial activation is a
real event and must be recorded too. The CLI and GUI use that order.

For example, from a chapter endpoint:

```sh
dotnet build Cli/Ch04.csproj
dotnet build Examples/Gui/Gui.csproj
dotnet run --project Examples/Skills/Skills.csproj -- /new/owned/exercise-directory
```

The separate `Examples/Skills` assembly exercises public construction, disabled
writes, load-time variables, provenance, pending removal, and a real write. It
uses no reflection, internal API grant, or model provider. It leaves its output
file for inspection and explicitly destroys the Agent graph.

The applications copy the shipped manuals from `Skills/*/SKILL.md` beside their
executables. They default to that output `skills` directory and primary
`ensemble`. The primary permits read/inspect tools and offers `code-tools` for
editing and job control. `full` is an explicit full-capability primary for trusted
local exercises, not a reason to grant all tools by default.

Both composition roots accept:

| Setting | Environment | Explicit flag |
| --- | --- | --- |
| Skill root | `EN_SKILLS_DIR` | `--skills-dir PATH` |
| Primary name | `EN_PRIMARY_SKILL` | `--primary-skill NAME` |
| Example variable | `EN_CUSTOM_VAR` | None; register other variables in a library host. |
| Deferred Anthropic mode | `EN_DEFERRED_TOOLS=1` | None. |

Flags override environment values. Existing `LLM_*` connection settings and
streaming options remain unchanged. Credentials never belong in a skill file,
variable value intended for model instructions, or committed request capture.

## 9.7 Exercises and evidence

Use an owned scratch directory and a working local provider fake first. Give a
small primary `read_file` and an offered writer skill `write_file`, with a
`list_directory` dependency. Set `$CUSTOM_VAR` to a harmless marker. Then:

1. Request a write before loading. Check both the paired error and absence of
   the file; merely hiding a declaration is not enforcement.
2. Load the writer. Inspect the next actual request for the primary's unchanged
   body, the newly rendered manuals in instruction context, and enabled tools.
3. Write a file, mark the writer pending, then write a **second** file with the
   dynamically granted writer. Verify bytes, not just the final answer.
4. Read the result and let the CLI reach normal EOF. Inspect its exit and usage.
5. Render the log offline with the original skill directory unavailable. Loaded
   instruction snapshots and control declarations must survive without HTTP.
6. Repeat through the GUI. Expand the real tool-result cards, check the pending
   explanation, reload the page, and verify stable replay without duplicates.

Also test cycles, missing dependencies, unknown handlers, malformed metadata,
idempotent loading, primary unload refusal, variable snapshot timing, two-Agent
isolation, and both directions of manager removal. A held-turn or fake-provider
test must use the real Agent and its actor, not invoke private handlers and call
that an integration test.

The Chapter 9 executable-only grader reuses the existing eight skill checks for
**85 points**. It excludes the inherited Chapter 8/Go topology campaign; it is
not a full Go grade. Component tests, the separate C# consumer, and actual
CLI/browser exercises cover different boundaries and must be reported separately.

Real-provider runs require their own authorization and evidence. Historical Go
sessions and demonstrations do not verify this C# implementation, deferred-tool
support on an account, audible speech, or human accessibility. Keep blocked
integration work explicitly incomplete rather than borrowing an earlier pass.

The next chapter gives recorded conversations a life beyond a single process.
This chapter provides the prerequisite: the actual manuals and grants are facts
in History, not an accidental consequence of whichever files happen to exist
when replay begins.
