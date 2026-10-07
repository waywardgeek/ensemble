---
name: ensemble-coding
description: Mandatory before writing or editing Ensemble code, chapter solutions, or their graders. Applies star dependencies, common data/interfaces, behavior in owning packages, interface parent chains, logger reachability, optional GUI modules, and actual live chapter demonstrations.
---

# Ensemble coding rules

## Mandatory load before code

Read this entire file before every coding task, including fixes, refactors,
tests, and grader changes. Reload it after compaction or a change of task.
Do not substitute recollection or a summary. A continuing edit within the same
task does not require re-reading before every keystroke; the rules must remain
in context throughout. Every delegated coding task must name this file and
require the coder to read it before editing.

These rules apply from **Chapter 1, before the first implementation**. A small
exercise is not permission to start with a flat package and repair it later.
Read the chapter's plain-English architecture rules and current exercise
contract too. Do not begin affected code while a contract contradicts these
rules or leaves required ownership ambiguous; report the gap to the author or
coordinator, who can ask Bill. Do independent, unblocked work meanwhile.

## 1. Organize by responsibility; imports form a star

- Core shared data structures, constants, and interfaces belong in
  `internal/common`. Interfaces for back-pointers belong there too.
- Implementation belongs in the directory responsible for the work:
  `internal/llm` for model requests, rendering, parsing, and conversation
  operations; `internal/jobs` for job lifecycle; `internal/tools` for tools.
  Add another spoke when its responsibility exists, not an empty directory
  for a hypothetical feature. Private implementation details stay in the spoke.
- Within the agent library, implementation spokes import `internal/common`,
  never one another. `common` imports no first-party implementation package.
  Standard-library use does not constitute a sideways dependency.
- The library's composition root can import concrete implementations to build
  the ownership tree. Executables call the public constructor; they must not
  duplicate its internal wiring. A legal, acyclic Go import can still violate
  the star. Do not confuse compiler acceptance with architectural compliance.

The shared vocabulary package is the star's center. It is **not** the runtime
Hub/Ensemble object that manages agents. Keep the two meanings separate.

## 2. Put behavior with its owner, even when that costs method syntax

Go requires a receiver's base type to be defined in the same package as its
methods. Do not let that language rule choose the architecture.

If a core type lives in `common` and its operations belong in `llm`, write free
functions in `llm` accepting that type. Use `Apply(context, event)` rather than
moving the reducer into `common` to retain `context.Apply(event)`. This is valid
Go. The price is function-call syntax, not misplaced behavior. A type alias in
the spoke does not let it attach methods to a type defined in another package.

Three cases from Chapter 5 govern placement:

1. Behavior used by one spoke: free functions in that spoke over the shared
   data. Do not move core data out of `common` merely to keep method syntax.
2. Methods required for standard-library interface dispatch, such as JSON
   encoding/decoding, `String`, and `Error`: retain those methods with the
   type they implement. This narrow exception does not license unrelated
   orchestration, persistence, parsing, or model policy in `common`.
3. Behavior needed by several spokes: declare an interface in `common`, put
   its implementation in an owning spoke, and construct it through the root.

Do not use `common` as a miscellaneous utility package. Shared declaration
does not imply shared implementation. Convenience methods are not a reason
to move responsibility across a boundary.

## 3. Reach data through the interface parent chain

Each child object receives and stores an **interface back-pointer to its
creator**, declared in `common`. Each parent interface exposes its own parent.
The application root has no parent to invent. Follow the chain to the object
that owns the needed data or service, then use its interface. Data stored in
another component must not become unreachable just because its concrete type
lives in another package.

For example, use the path `Call.Engine.Agent().Ensemble()` as those objects
are introduced. Do not add a redundant direct `Call.Agent` or sibling-service
field when that object is already reachable through the parent. Do not pass a
bag of capabilities the parent owns or attach model/logging/settings closures
after construction to compensate for a missing chain. Configuration values
that establish a child's identity are different from injected sibling services.

Capabilities stay on the object that owns the fact. Agent configuration belongs
on Agent; token usage belongs on Engine, which knows which model spent it.
Do not move facts onto an ancestor just because that ancestor is convenient
to reach. Name interfaces for their objects and extend the appropriate owner's
interface when a new capability is needed.

**Every line of code likely to need debug logging must have access to the
logger.** This applies to parsers and stateless helpers too. A free function
that needs access receives the appropriate owning context/interface; changing
from a method to a function must not sever that access. Do not wait until a
failure requires logging to discover that the chain is missing. No package
global logger and no unrelated logger parameter used to bypass ownership.

Mutable counters, registries, configuration, and session state belong to
objects, not package globals. Immutable lookup tables and compile-time
interface assertions are not mutable session state.

## 4. Keep the application owner separate from its clients

One Hub/Ensemble object owns potentially many Agents. Each Agent has an
interface back-pointer to it. **Ensemble** is the working name; final naming
does not alter the ownership rule. One instance per application does not mean
a package-global variable.

Starting in Chapter 2, use clean core data structures and provide two client
interfaces on Hub/Ensemble: a CLI and a browser GUI through WebSocket. The GUI
client may be stubbed in Chapter 2, but its interface and separation must exist.
Both clients use the same agents and application services, not independent
copies of the agent loop.

Hub/Ensemble receives streaming events and real-time display updates through
Observer. Observer is not the exclusive communication channel: explicit parent
interface methods may provide services or request actions such as creating a
sub-agent or sending a message. Do not replace observations with direct GUI
calls, or expose concrete coordinator internals to Agents.

GUI code, including its WebSocket transport, belongs in a **separate, optional
Go module**, never `agent/internal`. Merely moving GUI code to a sibling
package in the agent module does not satisfy this rule. The agent library
works without that module. GUI consumers use public interfaces; the user's
application chooses whether to include the GUI or reuse its components.
A Chapter 2 GUI stub is not permission to put WebSocket code in the core.

For the new implementation, the coordinator selects Hub/Ensemble as logger
owner, following Bill's preference; the chapter contract teaches that choice. Tool
visibility is per-agent. Registry storage may be Agent-owned or shared on
Hub/Ensemble. These remaining choices do not weaken the rules above. Settle
the relevant owner in the chapter contract before implementing it; do not
silently turn either registry option into a universal requirement.

## 5. Prove the rules, not their vocabulary

Before handing code back, check responsibilities, actual imports, constructors,
and the paths to owned services. A search for `Host` or `Logf` proves no chain.
Discover all present spokes and check each executable's use of the library
root; a historical hardcoded package list leaves new code unprotected.

Graders check behavior and structural properties, never authorship or lineage.
Equivalent identifier names should pass. Delete each protected behavior and
verify the expected failing checks and score loss, with passing controls and
negative controls for absence. Check-level coverage alone does not prove all
the promises within a check are protected. Keep fixtures independent enough
that one defect does not falsely accuse an unrelated behavior.

Do not weaken a test merely to reach green. Preserve comments explaining why
an unusual design exists until its underlying condition is demonstrably gone.
Before deleting such a comment, record that proof in the change evidence and
commit message; a refactor alone is not proof. Prefer targeted patches. A
scripted text replacement must assert that its anchor is unique before writing.
Report surprises and what they taught. Check formatting output, not just its
exit status. Record actual test commands/results and never present unrun
acceptance checks as evidence.

Before committing code, run formatting checks on the changed Go files and run
`go vet ./...` and `go test ./... -count=1` in each affected Go module.
`gofmt -l` must print nothing; it can exit zero while listing unformatted files.
Run the chapter's required grader/audit too. Do not format unrelated legacy
files or assume that testing the repository root tests nested Go modules.

Student coders use the chapter, this skill, and the preceding second-edition
baseline. They do not consult later solutions or grader implementation to
guess missing requirements. Surface missing teaching instead.

After the initial implementation and runs, preserve the student checkpoint
and take part in the mandatory independent comparison with the first-edition
standard. The code reviewer reads that standard and returns concrete findings;
revise the new code from their rationale without copying the old answer. A
passing grade is insufficient: seek better design, clearer and tighter code,
useful comments, and better teaching. Send instruction gaps to the author,
rerun affected validation, and obtain review of the revisions before declaring
the chapter complete. See §5 of `book/chapter-writing-procedure.md` for the
comparison evidence and quality gate.

## 6. Actually take every chapter for a spin

Every chapter with a "Taking it for a spin" section requires the coder to run
the actual user-facing program with a **real model backend**, initially through
the CLI. Verify that a real user can exercise **every feature in the chapter**.
Do not substitute direct calls to internal functions, a fake server, or a
fabricated transcript for the user path.

**The CLI must have a human chat mode.** Starting in Chapter 2, the reader
must be able to launch it, enter ordinary text, see understandable answers,
continue the conversation, and exit without constructing JSON. Teach its
invocation and behavior in the chapter before implementing it. Preserve the
machine protocol for automation as a separate interface.

The coder must drive that same human mode in an actual interactive terminal
(a PTY is suitable), observe each response, and then send follow-up input.
Exercise the chapter's features and recovery paths with the supported real
providers, retaining sanitized terminal evidence. Feeding JSON lines to stdin
does not fulfill this human-interface gate, even with a real backend. A direct
shell or debugger run outside Ensemble does not prove Ensemble's interface.
"Taking it for a spin" must guide the reader through this usable mode and
report the coder's actual run. Do not claim Bill or another reader tested it
unless that happened; their separate participation is not an invented receipt.

Before running, map each chapter feature to a concrete user action and an
observable result. Exercise the real interface, inspect the results, and record
the command, date, provider/model actually used, and sanitized evidence for
each feature. Fix inaccessible or broken features through the chapter contract
and appropriate regressions, then repeat the affected live demonstration.
The author writes the demonstration from those receipts, not predictions.

If shipping a replay/evidence verifier, bind it to the recorded executable
hash and immutable source revision before it runs or rewrites derived files.
Allow an explicit executable location; a temporary pathname alone is not an
identity. Resolve historical source against that revision, not today's working
tree. A mismatched executable or source must fail before changing evidence.
Keep original terminal/log receipts distinct from reconstructed requests.

Deterministic fake-server grading and mutation audits remain required where
specified. They complement live demonstrations; neither replaces the other.
A deliberately stubbed GUI must be labeled as such, never reported live-tested.
If credentials, network, provider access, or an unfinished interface prevent
a required live check, report that specific blocker and leave the chapter
unvalidated. Do not quietly downgrade acceptance to fake-only testing.

Bill authorized using API keys in `~/.cr/settings.json` for these live checks.
Read only the needed settings programmatically; never dump that file or print
keys. Keep secrets in memory or a subprocess environment, not command-line
arguments, transcripts, prompts, or repository files. Disable or redact
credential-bearing request logs before running. Evidence and git diffs must
contain no credentials. Never commit the settings file or a copied secret.
Use bounded demonstrations, not unattended retry loops. Exercise all three
initial vendors (Anthropic, OpenAI, Gemini) as their adapters are introduced.
At the caching chapter, re-test Bill's reported OpenAI OAuth caching failure
using current official subscription-access guidance and fresh measurements;
do not treat the historical report as proof of current behavior.

## Workspace and regression boundaries

Write the student rewrite in `solutions/edition-2/main/`, tracked in the outer
Ensemble repository. Its Go modules remain independent of the historical
implementation. Extend the accepted preceding second-edition source. Frozen
`solutions/edition-2/chNN/` directories are exact exports for readers and
verification; do not edit them as parallel working trees or initialize nested
Git repositories. Never edit or copy the existing `agent/` implementation or
first-edition solutions.

Preserve the initial student attempt before comparison. After each chapter
passes its required gates, export the source with a manifest identifying the
source revision, then commit the manuscript, source, snapshot, and evidence in
the outer repository and create an immutable annotated `edition-2-chNN-rN`
tag. Corrections receive new revisions, never moved tags. Carry earlier fixes
forward through affected chapters and rerun their affected checks; freezing a
snapshot does not make a later solution inherit a fix automatically. Historical
nested student repositories and their unfinished work were preserved during
consolidation; see `solutions/edition-2/history/README.md` for restoration.

Grader improvements are allowed, including limitations discovered by the
student, but coverage must not be weakened and legacy tests must still pass.
Record pre-change baselines and investigate regressions. A new grader pass
does not substitute for checking old behavior against the enhanced grader.

## Sources and current authority

Paths below are relative to repository root:

- `book/edition-2/chapter-01.md`: rules taught before any code.
- `book/edition-2/architecture.md`: current decisions and open owner choices.
- `book/edition-2/workflow.md`: mandatory handoffs and author/student loop.
- `book/chapter-05.md`, especially sections 5.1, 5.2, 5.9, and 5.10:
  original architecture lessons, corrected by Bill's current instructions.
- `book/common-refactor-design.md`: free-function rule and its rationale.
- `agent/skills/ensemble/SKILL.md` and course policy P9/P11: parent chain,
  real ownership, and checks of architectural properties.

Bill's 2026-10-07 instructions supersede the first-edition chronology and
examples that postpone architecture, inject siblings, duplicate roots, or
describe mutable global registries. Chapter 1's small conversation model is
not an exemption from the coding methodology.
