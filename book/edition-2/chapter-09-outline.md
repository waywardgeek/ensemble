# Chapter 9 outline: Capabilities with an owner

Preparation only, October 7, 2026. New Chapter 9 maps to first-edition
Chapter 10 under `workflow.md` and `global-review.md`. These are proposed
contract choices for coordinator review, not a student handoff. No Chapter 9
implementation, passing checker or live result exists. Draft the complete
chapter only after outline review; release a fresh student only after accepted
Chapter 8 and a published independent contract checker.

## Stake and story decision

The reader wants to give one Agent editing instructions and tools without
changing another Agent, rewriting the Agent's identity or leaving it with tools
whose instructions have disappeared. The resolution is a visible load, useful
file operation and unload through the same CLI/browser/public library, with
recorded material and mechanically enforced capability changes.

Retain first-edition Chapter 10's constitution/manual distinction and progressive
disclosure explanation. It has no documented human incident needing a fictional
scene. Bring forward Chapter 15's concrete failure: a loaded manual stored only
in a tool result disappeared when that result was cleared while its tools stayed
available. Teach a dedicated instruction entry at the first load, without
requiring the new student to recreate the defective storage. The later chapter's
claim is historical evidence, not a new measured run. Keep the useful premise
that a capability bundles instructions and tools; omit unsupported tool counts
and promises that a changing tools array preserves provider cache hits.

Voice plan: lead with the reader's extension task, then the failed lifetime
assumption and the owner that fixes it. Use a small dependency/disclosure tree
throughout. Keep diagnostics, retained text and actual file effects visible.
No invented Bill dialogue, model obedience, cache savings or successful spin.

## Scope and teaching order

1. Instructions, installed handlers and enabled tools answer different questions.
2. Define the small SKILL.md format, discovery boundary and immutable catalog.
3. Give Skills, variables, tool grants and transitions explicit owners.
4. Resolve a complete candidate dependency graph before making anything visible.
5. Record primary identity and dynamic manuals with distinct lifetimes.
6. Enforce the effective set at each tool admission, including load/unload.
7. Expose owned state through public watch, human CLI and the existing browser.
8. Demonstrate real use, deterministic failures and a public extension; preserve
   initial experience before the independent historical comparison.

Chapter 8 remains authoritative for GUI preferences, execution policy, active-
turn captures, pause and connection lifetimes. This chapter adds no credential
settings, remote plugin connection, automatic compaction, saved Agent restart,
sub-agent orchestration or provider-specific cache protocol. The next persistence
chapter may restore live sessions; replay here reconstructs recorded facts only.

## Proposed ownership and data contract

| Owner | Authoritative facts and permitted work | Parent path and lifetime |
|---|---|---|
| Ensemble | Application composition and logger; installs the handlers available to each Agent | Application root; never a package-global skill registry |
| Agent | Creation-time skill selection, installed-tool ceiling, Skills service and actor | Agent→Ensemble; a skill cannot replace Agent configuration or widen the ceiling |
| Skills | Immutable parsed catalog, active roots/dependency closure, effective grants, available loadable names, material revisions and variable bindings | Skills→Agent; declarations/interfaces in core common, behavior in `internal/skills` |
| Actor | Orders load/unload acceptance, durable mutation and observations with turns, hints, pause, interrupt and close | Actor→Agent→Skills/Registry; no sibling implementation imports or injected service closures |
| Registry | Handler/schema storage and decoding of load/unload tool arguments; declarations and dispatch use the same committed effective grants | Registry→Agent; reaches Skills through the real parent interface |
| Catalog parser and variable rendering helpers | Produce owned candidate facts and safe errors; no hidden mutation, provider calls or filesystem effects during replay | Receive their Skills owner/context; can reach Agent→Ensemble logging |
| Optional GUI | Projects owned skill state into accessible controls/cards | Existing Server/Connector/page parent chains; no authority over grants |

Mutable runtime values remain on instances. Private implementations may live
in responsible spokes behind common parent interfaces. Names are design choices;
who owns a fact and who can commit it must be specified before code. No `init`
registration, mutable global map, direct Engine→Skills shortcut or callback bag.

The current registry's installed/builtin selection becomes the ceiling for
skill-enabled Agents, not permission for a skill to install arbitrary Go code.
Agents without skill configuration preserve preceding behavior. Primary and
catalog identity are creation-only. Discover/read files at construction before
publishing the Agent; dynamic loading uses those frozen bytes and definitions.
A file edit affects a newly created Agent, not an in-flight load by surprise.
A trusted public application may supply the equivalent owned catalog directly,
which supports external consumers and tests without a private filesystem hook.

## Proposed format and initialization boundary

- One named directory under an operator-selected skill root contains SKILL.md.
  A canonical skill identifier is also its directory name; references use IDs,
  never arbitrary paths. Reject duplicate/mismatched IDs and traversal syntax.
  Resolve the root once; reject symlinked skill directories/files rather than
  following a reference outside the catalog. This constrains discovery, not the
  filesystem authority of already enabled shell/file tools.
- Use UTF-8 Markdown with a leading and closing frontmatter delimiter line.
  Body horizontal rules survive unchanged. Require name, description and type;
  support primary, loadable and dependency. Tools, depends and loadable-skills
  accept a space-separated scalar or a YAML sequence of strings. The full draft
  must publish exact allowed scalar/list forms and error fixtures, including
  duplicate/unknown fields, wrong kinds, missing delimiters and trailing syntax.
- Unsupported integration headers fail explicitly. Do not accept mcp_servers,
  commands, secrets or network configuration and quietly do nothing with them.
  No parser behavior creates an executable plugin in this chapter.
- Proposed resource limits for review: 256 definitions, 64 KiB per source or
  rendered body and 8 MiB aggregate catalog bytes. Enforce before publication;
  the complete draft will define whether delimiters/frontmatter count toward
  each bound and include exact boundary cases. Do not hide grader-only limits.
- The selected primary and its dependency closure load before the first request.
  All required names, types, handlers and grants must validate or creation fails
  without publishing a partly initialized Agent. Reachable discovery edges may
  expose loadable skills; dependency edges activate instructions/capabilities.
  These two relations must not be conflated.
- Ship a usable primary skill for the new CLI/browser invocation as well as a
  deliberately narrow teaching fixture. Demonstrate the shipped primary, not
  only a grader-created variant. Avoid copying the old fixed twelve-tool list:
  use the new predecessor's actual installed tools and new management tools.

## Proposed transition rules

Load names one currently discoverable loadable skill. Primary/dependency-only,
unknown and undisclosed names cannot be requested as dynamic roots. Resolve its
entire dependency closure with cycle detection and deterministic ordering into
an owned candidate. Missing dependencies, unsupported tool names or a required
grant outside this Agent's ceiling fail the whole operation. No partial grants,
instructions, discovery changes or revision survive an unsuccessful candidate.

Distinguish explicit roots from shared dependencies. A dependency reached by two
roots loads once and remains active until its last active root no longer needs
it. The primary is a permanent root. A diamond is valid; a cycle is an error.
Repeated load of an already active skill is an acknowledged no-op without a
second instruction entry. Loading a previously retired skill is a new activation
with a new recorded body identity. Exact revision/no-op/error payloads belong
in the complete draft and fixtures.

Unload removes an explicitly active dynamic root. It cannot remove the primary
or directly remove a dependency still required by a root. Recompute the closure,
discovery set and tool contributors; retain a shared or initially granted tool
while any active contributor still grants it. Tools with no remaining grant stop
being declared and stop passing dispatch authorization immediately after the
actor applies the change. Already admitted jobs keep their ordinary lifetimes.
Unloading a skill is not a process cancellation mechanism.

Retire its material for a later explicit retention/compaction operation; retain
recorded instructions in chronological context in this chapter. Say plainly that
old text can remain after authority is revoked. The effective registry, not a
model's recollection of the manual, decides whether another call can execute.
Later compaction can remove retired material using the recorded identity/state;
do not implement a hypothetical compactor now or claim unload erases history.

Use actor-local capability control operations for these bounded in-memory
transitions. Tools owns wire decoding; Skills prepares typed candidates; Actor
commits them through Agent's existing durable path. This avoids a tool worker
blocking on the actor that is waiting for that worker. The full contract must
explicitly classify management tools under Chapter 4's literal-next-call
one-shot-limit rule and retain Chapter 5 pause/admission/pairing semantics.
Proposed choice: paired control tools without a process Job, consuming pending
limits just as any next call does; manuals have their own bounded durable entry.

Authorization is checked at each admission against committed current state.
A load then an enabled call in one accepted batch can work; an unload then that
call must refuse it without executing. Unknown/disabled calls still receive
matched error results. Skill changes never roll back earlier tool effects or
rewrite the declaration set of an HTTP request already sent.

## Material, replay and variable extension

The primary's rendered body is a creation-time instruction. Resolve built-in
$TOOLS and $SKILLS against the complete initial candidate, then retain those
exact bytes. A dynamic skill gets its own typed skill-material event/entry with
name, activation identity, rendered body and origin. Its tool result is a short
acknowledgement. Tool-result clearing cannot accidentally erase the only copy
of a still-needed manual.

Record committed transitions and selected body/grant facts so replay does not
reread SKILL.md, expand variables, resolve dependencies against edited files or
execute loads. Initial grants and later contributing activation sequences retain
provenance without forcing renderers to import the skills package. Historical
logs without skill events retain their previous meaning. Replay is not live
recovery or permission to reopen external services.

Land dynamic instruction material at the completed batch boundary so the next
provider request preserves every call/result pair. It remains a distinct neutral
purpose and is rendered at its chronological dialogue position, not appended to
the enduring system prefix or only buried in a result. The full chapter needs
literal valid render fixtures for all three APIs before this becomes a student
requirement. No unsupported inline-tool protocol or cache-hit guarantee follows
from retaining provenance.

Proposed bounded variable seam: Skills owns built-in TOOLS/SKILLS rendering over
its complete candidate state; the application supplies per-Agent named scalar
bindings through the public constructor. No arbitrary environment lookup or
registered closure captures sibling services. Built-ins cannot be overridden.
Publish a one-pass token grammar, literal-dollar escape, deterministic ordering
and unknown-variable error. Freeze the expanded bytes in each recorded body;
never recursively expand a substituted value. A public consumer adds a third
variable without editing a core package and proves two-Agent isolation.

Two substantive choices need coordinator disposition before the full draft:

1. **Primary versus existing System configuration.** Recommended: skill mode
   treats primary identity as creation-only and rejects a competing explicit
   System override or later replacement; Agents without skills keep earlier
   behavior. Existing deliberate instruction appends need a stated policy.
   Claim stability of the primary's rendered bytes, not unconditional stability
   of every system byte after a caller explicitly adds enduring instructions.
2. **Variable extension scope.** Prefer immutable scalar application bindings
   now, with built-in computed lists evaluated at each load. If computed custom
   variables are required at this introduction, define an owned typed service
   and bounded execution contract before implementation. Do not silently retain
   the historical closure registry as the architecture.

## Public and human surfaces to specify in the draft

Expose owned Skills state and actor-ordered load/unload through public library
interfaces. A headless caller can inspect the primary, active roots and
requirements, retired material, currently discoverable skills and effective
names. Do not include credentials, arbitrary Config, absolute source paths or
unloaded skill bodies in a browser snapshot. Public mutation requests cannot
bypass dependency, visibility or tool-ceiling rules.

Extend the current atomic watch with a safe skill-state snapshot and a revisioned
change observation. Reconnect recovers that state at the existing watch boundary;
it does not need to reconstruct current grants from the last 100 cards. Give new
instruction cards a visible skill label, stable identity, safe text expansion
and the existing no-HTML/no-auto-fetch treatment. Replay remains silent and no
skill observation is mistaken for a prompt completion or preference update.

Human CLI gains a read-only /skills view; ordinary prompts ask the model to load
or unload through its declared tools. Preserve prior machine protocol unless a
new opt-in record is explicitly specified. The browser displays authoritative
state alongside its existing cards and can submit the same ordinary prompts.
The GUI and CLI on --terminal continue sharing one Agent. Exact command/wire
payloads and safe snapshot fields are required in the full contract.

## Required checks and actual-use plan

| Promise | Distinguishing local control |
|---|---|
| Format | Scalar/list equivalence, body rule preservation, malformed/duplicate/type/path/size refusals |
| Atomic graph | Cycle, missing dependency and disallowed grant leave all state/events unchanged; diamond loads once |
| Visibility | Hidden/dependency/primary refusal; reveal next level only after successful load; no leaked other-Agent grants |
| Dispatch | Guess disabled handler before load and after unload; declarations and actual effect agree; same-batch transitions |
| Unload | Shared/primary contributor survives, unreferenced grants disappear, admitted job continues, repeated action has taught result |
| Instruction lifetime | Primary bytes stable across ordinary loads; dynamic body once in next request and dedicated entry; clearing the result does not erase it |
| Variables | Candidate lists, custom per-Agent value, literal dollar and unknown token; no second expansion or environment fallback |
| Replay | Edit/delete catalog after run; zero filesystem/catalog/model work and equivalent recorded projection |
| Actor boundary | Pause, held HTTP, interrupt and close around a planned/committed change; persistence failure never invents rollback |
| Public/UI | Owned state mutation, two Agents, snapshot/tail race, browser card safety, reconnect and optional-module/headless reuse |
| Compatibility | New Chapter 8 policy/preferences/speech and Chapters 1–7 CLI/jobs/streaming/watch behavior remain protected |

Before paid use, bind the exact new source, executables, skill files and complete
feature/action plan. Use all three current API paths through actual human CLI
PTY, the real browser and a public two-Agent consumer. Each sequence starts with
a narrow primary, inspects /skills or the browser state, loads editing capability,
observes its receipt, performs one scratch-file change and checks the actual file.
Then load a revealed skill/dependency, use its tool, unload and inspect both the
remaining shared grants and denied former capability. Preserve a model's refusal
or invented success and observe it before a bounded corrective prompt.

Also use the shipped full primary through an ordinary user task. Record primary
and dynamic body facts, actual declarations and tool calls, requested/returned
model identities, normalized usage, state observations and file hashes. Reconnect
one browser, confirm silent restoration of skill state/cards, and use the shared
Agent from the terminal. The public consumer gives two Agents distinct variables
and grants, exercises a custom scalar binding and proves no sibling leakage.
Local fixtures cover exact denial, malformed graphs, commit failures and races;
model obedience cannot replace those controls. No new speech mechanism is added,
but preserved autoplay/silent-replay behavior must survive the new card type.

Historical `make grade-dir CH=10 DIR=solutions/edition-2/main` is diagnostic only:
its binary layout, WebSocket commands, tool roster and parity assumptions predate
the new contracts. Publish a new independent acceptance command and audited
coverage before handoff. Preserve legacy positive results and add intended
negative controls; never treat the old declaration-oriented score as complete
capability enforcement. Author writes the actual spin only from retained runs.
