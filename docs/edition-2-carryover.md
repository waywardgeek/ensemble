# Edition 2 carryover: lessons for the fresh student

**Student run authorized by Bill, October 9, 2026.** Bill reviewed this carryover
with [the workflow and learning record](agentic-codebook-workflow.md), refined the
rules and explicitly authorized autonomous work. Use it with the original chapter;
it is not a replacement assignment, new chapter or implementation design.
Unresolved source contradictions remain identified below for resolution when relevant.

Read this alongside the mandatory
[`.agents/skills/ensemble-coding/SKILL.md`](../.agents/skills/ensemble-coding/SKILL.md)
and the original chapter being attempted. Start at unchanged first-edition
Chapter 1, from empty `solutions/edition-2/main/`. Bill has deleted the remaining
second-edition Chapter 1 draft; no replacement manuscript is needed. The student
does not consult `agent/`, existing answers, discarded answers, their Git blobs,
or grader implementation. Original chapters and graders remain unchanged.

**Authority and timing.** Bill's explicit decisions govern architecture and
behavior. Source lessons below explain failures and how to avoid them; they do
not bring later features forward. Suggested designs remain provisional. The
student may choose a simpler implementation and explain it. Consequential
unresolved contradictions go to Bill. **The student need not pass a flawed grader
test:** it may explain the failed check and supporting evidence in its chapter
review instead. Bill authorizes a reviewer to accept a documented grader defect
or unreasonable exercise requirement when it independently agrees with the coder:
record the exception, its reasons and actual grader/live
results, then advance. A failed check remains failed; this cannot waive Bill's
architecture rules. See the [exception procedure](agentic-codebook-workflow.md#reviewer-acceptance-of-an-unreasonable-exercise).

Reviewer feedback may lead the student to revise both code and its chapter
review. The [bounded review loop](agentic-codebook-workflow.md#bounded-studentreviewer-loop)
allows three revision rounds after the initial review by default, stopping on
acceptance or, if still unable to pass, bringing Bill the unresolved evidence
before further work. The limit is adjustable by Bill; failed grader tests can
still receive documented reviewer acceptance as described above.

**Destination.** Bill requires the new Ensemble to run **GPT-5 Astra at high
reasoning**, and Edition 3 to be built using that Ensemble rather than Codex.
Verify the supported model identifier, access method, reasoning setting and real
coding use when integration reaches that point. Do not infer an API identifier
from a display name, substitute another model silently, or build future
orchestration early. Readiness and actually producing Edition 3 are different
claims. Codex is the primary creator of Edition 2; authorship begins only after
the complete implementation and comparative reviews establish its improvements.
At that point the author and coder work together to apply the reviews to both
chapters and graders. The coder validates grader improvements and technical
claims with the author; prose and code reviewers check the resulting changes.
Preserve the original edition, grader coverage and student-run evidence while
recording corrections and new results explicitly.

## From the first line: restraint, ownership and evidence

**Work as the student.** Autonomous execution is not proof of expertise in this
methodology. Learn the reasons in the book and carryover before substituting a
familiar design. When evidence challenges an instruction, use the documented
review and exception process. Preserve corrections with their reasons and apply
them in later chapters; writing a note is not enough if the same mistake returns.
This is the operational lesson drawn from Bill's
[self-improvement essay](../book/the-improvement-loop.md), not a new requirement
to implement training or expanded memory during this student run.

**Live-test credentials.** Bill has authorized the provider API keys in
`~/.cr/settings.json`. Parse that file locally in the launcher/application and
pass only the needed key in memory or the intended test child's environment;
never print the file, extract values into tool output or interpolate them into
shell commands. Keep settings, auth headers and keys out of logs, prompts and
committed evidence. The [coding skill](../.agents/skills/ensemble-coding/SKILL.md)
contains the access procedure, including safe schema discovery and logging checks.
This path is Bill's local test setup, not a required Ensemble user configuration.

**Keep reasons, not just mechanisms.** Chapter 0's failed reconstruction and the
[epilogue](../book/epilogue.md) explain why code alone did not transfer Bill's
judgment: a plausible alternative can recreate a failure whose reason was lost.
Keep useful WHY comments and short records of failed approaches, corrections and
human assistance. A comment may become obsolete; prove its underlying condition
is gone before deleting its explanation. KISS means the smallest clear design
meeting the current exercise, not compressed code or removal of useful comments.
Sources: [Chapter 0, “What this book is”](../book/chapter-00.md#what-this-book-is),
[Chapter 20, §20.5](../book/chapter-20.md#205-the-agent-looks-in-the-mirror).

**Astra: this code is for teaching. Comments matter.** Bill found the discarded
attempt far too light on comments. Write for a student who needs to understand
why the code has this shape without reconstructing the design from other files
or old conversations. Explain purpose, ownership, lifetime, ordering constraints,
and non-obvious choices beside the code they govern. For example, explain why a
job continues after a wait returns; merely labeling the wait call teaches little.
Aim for **roughly 20% comments** in the implementation, about one comment line
for four code lines, excluding blanks. This is a teaching target, not a quota
to satisfy with padding or narration of obvious syntax. The reviewer assesses
both the balance and whether the explanations actually help the student. Useful
comments are part of code quality, not bloat. This is Bill's October 9 direction,
in addition to the first edition's lessons.

**Own the request bytes.** Chapter 1 starts with raw HTTP, without an SDK or
framework, because payload placement later determines steering and cache
behavior. That is not permission to implement either feature immediately.
Walk the response blocks as the exercise requires, retain the conversation and
report actual usage; keep protocol output usable and diagnostics separate.
Discover available models during setup/testing instead of treating the book's
historical names as an inventory. Discovery is not an extra API call to add
automatically to the exercise's conversation loop.
Source: [Chapter 1, §§1.1–1.2](../book/chapter-01.md#11-frameworks-and-why-this-book-uses-none).

**Give facts owners and make those owners reachable.** The logger ambush in
Chapter 5 and the status-tool failure in Chapter 22 share a cause: a small
consumer could not reach existing facts without globals or new wiring. Shared
core values and interfaces belong in `internal/common`; implementation spokes
import common, not one another. Behavior belongs in its responsible package,
using free functions over common values where needed. Standard-library dispatch
methods may remain with their types; private runtime structs behind common
interfaces are allowed. Do not create speculative packages to anticipate later
chapters. Sources: [§5.1](../book/chapter-05.md#51-the-idea-in-plain-words),
[§5.9](../book/chapter-05.md#59-the-ambush),
[Chapter 22's diagnosis](../book/chapter-22.md#the-symptom-a-feature-that-should-have-been-easy).

**Bill's clarification is stronger than some examples.** Every child constructor
receives and stores interface back-pointers to all required parents, which cannot
be nil. Optional parents may be absent and add the child later. Adding a child
modifies both objects: the parent records the child, and the child updates its
back-pointer to that parent. Each object exposes its parent relationships through
interface methods. The rule allows multiple parents: a pin must belong to an IC
package but may float with no wire. Connecting it to a wire establishes a second
parent relationship, retaining the package relationship. A family-tree node with
two parents likewise needs both back-pointers. The data model determines which
relationships exist and which are required, not a single-parent restriction.
A bag of sibling services is not a substitute for those relationships. Reach
configuration, services and logging through ownership; no mutable application
globals or callback bags. The root needs no invented parent. One Ensemble owns potentially
many Agents and the logger; Agent owns configuration; Engine owns usage. Tool
visibility is per-Agent, whether registry storage belongs to Agent or Ensemble.
Introduce these objects only as needed. The executable must actually use the
public library's composition root: a correct constructor with zero callers
protects nothing. Source: [Chapter 22, “The measurements that found it”](../book/chapter-22.md#the-measurements-that-found-it);
binding clarification: [coding skill](../.agents/skills/ensemble-coding/SKILL.md#architecture-from-the-first-line).

Make the chain concrete: Engine's parent is Agent; Agent's parent is Ensemble.
Agent owns its conversation and configuration; Engine owns its HTTP transport
and usage. `common` is shared vocabulary, not the runtime Ensemble object.
Stateless parsers and helpers likely to need diagnostics also receive owner
access. Follow the actual parent relationships rather than adding a root shortcut or
downcasting to a concrete parent. When multiple spokes need a service, expose
its interface through common and the ownership chain, keeping its behavior in
the responsible package. These clarify the existing architecture, not new services
to build in advance. They were stated explicitly in the deleted Chapter 1,
§§1.2–1.4 (historical comparison recorded below).

**Name the root and its clients distinctly.** The
[topology decisions](ensemble-topology-decisions.md#the-three-objects) record
Bill's distinction: Ensemble owns Agents and their lifetimes; GUIServer owns
browser connections and WebSocket framing; Gateway authenticates and translates
at a remote boundary and owns no Agents. Ensemble is a **gateway client**, not
the gateway itself. The application constructs Ensemble, including in the
single-Agent headless case; do not leave essential ownership absent in the bare
configuration used by tests. This clarifies the existing root requirement, not
an instruction to build a gateway or sub-agent manager in Chapter 1.

**Agent identity must not come from process-wide accidents.** The topology
document records Bill's rulings that the application supplies the root data
directory, each Agent receives its own directory, and defining skills are
per-Agent constructor inputs when skills arrive. Model and request settings
also belong to the Agent. Keep filesystem and skill identity separate from LLM
request configuration that is rendered onto the wire. As persistence, memory
and skills arrive, resolve their files from the owning Agent's directory instead
of baking in the process working directory or one global skill choice. A shared
`./save.json` accidentally limits the application to one Agent. Establish one
consistent path convention at the application boundary; mixing relative and
already-prefixed paths can duplicate sub-agent directory prefixes. The recorded
default is a directory tree under the application-selected root; confinement
remains the sandboxing chapter's responsibility. Introduce only the identity
fields needed by the current feature, not all future services. Source:
[per-Agent data and skills](ensemble-topology-decisions.md#per-agent-data-directory-and-per-agent-skills-bill-ch22-session).

**Keep clients outside the core.** Bill requires GUI/WebSocket implementation in
an optional separate Go module outside `agent/internal`, using public core
interfaces. From Chapter 2, retain the requested small GUI stub and a usable
live CLI over the same Agent; grow the GUI when its chapter arrives. Streaming
observations reach Ensemble through Observer; explicit parent service methods
are also allowed. MCP transport stays replaceable when introduced, including a
WebSocket tunnel. Chapter 6's GUI-coupled framework and Chapter 20's CLI that
bypassed the actor show why a seam must work in the real clients. Sources:
[§6.2](../book/chapter-06.md#62-the-framework-that-imported-its-own-gui),
[§20.3](../book/chapter-20.md#203-two-paths-one-screen).

An external GUI module must be able to use the public library without importing
its `internal/common` package. Expose the values and interfaces its actual
features need through the public API; moving files alone does not establish
that boundary. Keep MCP service/routing concerns separate from the GUI's
WebSocket framing so a tunnel is one transport, not a headless runtime
dependency. When agent switching arrives, a newly attached observer needs that
Agent's history as well as live events. Keep replay cursors associated with the
right Agent and check the replay/live handoff for gaps or duplicates. This is
future feature guidance, not a multi-Agent GUI requirement for the early stub.
Source: topology decisions,
[known costs and open questions](ensemble-topology-decisions.md#known-costs-and-open-questions).

**Passing a grader and working for a person are separate results.** Use the
original grader, then the actual interface with real models as features arrive.
Keep demonstrations bounded and use scratch work where appropriate: Chapter 3's
live agent edited an unwanted codename into the book repository. Preserve
failures honestly. A fake and an implementation built from the same assumption
can agree while both are wrong; every fixture also hides the configuration it
replaces. Chapter 21's supplied skill passed while the deployed primary skill
could not load web search. Sources: [§3.7](../book/chapter-03.md#37-fakes-first),
[§3.8](../book/chapter-03.md#38-drive-it-yourself),
[Chapter 21's live defect](../book/chapter-21.md#the-defect-the-grader-could-not-see).
This is a testing discipline, not a new audit framework or permission to edit
graders. Comparative review must examine active code, comments, meaningful tests
and scope; more code or more tests alone proves no improvement.

**Verify what actually ran.** CodeRhapsody's memory records false confidence
from excluded tests, an unavailable command and a harness building the wrong
program. Check the actual command output, selected packages/build tags and
executable being exercised. A successful build is not a test run; an empty or
truncated result is not complete evidence. For example, `gofmt -l` can return
success while listing files needing formatting. Diagnose the actual invocation
and observed wire/artifact before assuming a cache problem or enumerating
producers. Prefer a faithful boundary fake over mocks that remove the behavior
under test. **No mocks without Bill's explicit approval.** Use the grader's
provided fakes where suitable or small additional fakes. A perceived need for a
mock is a critical question: stop, explain why fakes cannot meet the test need
and wait for Bill's decision. Reviewer exceptions cannot authorize mocks.
Bill's cross-chapter rationale matters: a Chapter 12 coder may not know what the
Chapter 4 coder learned. Mocking that earlier implementation can erase precisely
the behavior and integration constraints the later chapter needs to exercise.
Use the real earlier code through its interfaces, with fakes at external
boundaries; chapters divide the teaching, not the system's correctness.
Run the required checks and affected regressions; this is not a
request for blanket test expansion or a new verification framework.

CodeRhapsody's `learnings.json` adds concrete checks: a root-module test command
does not cover separate Go modules; a pipeline's final command can hide the test
command's failure; and missing stdout is a reason to inspect stderr before
guessing. Assert the requested observable, such as the actual endpoint rather
than just the vendor name. Derive fixtures from observed artifacts so the test
and implementation do not merely share one wrong assumption. Before calling a
failure pre-existing, reproduce it on the recorded baseline in an isolated
worktree when feasible; otherwise label that explanation unverified.

**Test the tests.** Demonstrate that each distinct behavioral claim in the
chapter's tests detects a targeted defect in that behavior. Confirm the mutation
applied and the intended check fails for the intended reason; a compile error or
unrelated failure does not count. Investigate survivors and record limitations
honestly for the reviewer. Reuse evidence for unchanged behavior and keep the
checks focused, using the [skill's mutation procedure](../.agents/skills/ensemble-coding/SKILL.md).
Mutations test sensitivity; they do not prove that an expected value or protocol
assumption is correct. Those still need independent support. This strengthens
existing verification without requiring a new testing framework.

**Make changes small enough to inspect.** Read the assigned chapter and
carryover before editing. For structural changes, move a coherent piece, use
compiler feedback and inspect the resulting diff, including comments, strings
and prompts that broad text replacements can corrupt. Explain the reason for a
choice and make uncertainties visible. CodeRhapsody's craft guidance supports
simple, well-explained code; it does not authorize rewrites, history resets or
new design requirements beyond Bill's current instructions.

## Apply when the relevant feature arrives

**Chapter 2: keep history, context and requests distinct.** The ordered,
append-only log records facts; the reducer derives context; the renderer makes
a disposable vendor request. Parsers emit facts rather than secretly mutating
context. Preserve provenance, call identities and opaque material that cannot
be reconstructed later; do not pass one model's private replay material as
another's thinking. Keep infrastructure failures distinct from ordinary failed
tool results. Determine current turn state from the relevant end/state, not
from the existence of old tool results somewhere in resent history. These
boundaries enable later replay without adding later policy now. Sources:
[§2.3](../book/chapter-02.md#23-history-context-request),
[§2.4](../book/chapter-02.md#24-the-log),
[§2.6](../book/chapter-02.md#26-the-seam).

Ephemeral delivery is different from erasing the record. Chapter 2 records an
ephemeral arrival in the event log for replay, delivers it in one request and
then clears the pending material; it does not become retained dialogue. As
context assembly grows, keep volatile material at the request tail, after the
stable prefix, instead of rewriting old instructions or accumulating stale copies.
The deleted Chapter 1's shorthand “never in history” must not be mistaken for
“never recorded in the event log.” Sources: Chapter 2's
[grading clarification](../book/chapter-02.md#the-checks) and Chapter 15,
[§15.5](../book/chapter-15.md#155-the-layout).

**Chapters 3–4: make tools recoverable, and waiting distinct from execution.**
Dedicated file/search/edit tools constrain output and avoid shell quoting
hazards; do not collapse the exercise into one shell tool. Return every tool
result with the right call identity, including ordinary command failures. A
supposedly quick tool can hang, so Chapter 4 supervises calls uniformly at the
dispatch point rather than maintaining a list of slow tools. A wake delay returns
control and a handle; it does not prove execution stopped. Preserve full output
outside the bounded model readout. The job owns interactive process/output
lifetime, completed jobs remain waitable, and cancellation must describe what
actually stopped. Avoid sticky hidden shell state. Sources:
[§3.4](../book/chapter-03.md#34-do-we-need-anything-other-than-run_command),
[§4.3](../book/chapter-04.md#43-stop-throwing-away-the-result),
[§4.7](../book/chapter-04.md#47-the-terminal),
[Chapter 13's debugger failure](../book/chapter-13.md#a-debugger-that-would-not-start).
Do not add process/job machinery in Chapter 3 to solve Chapter 4 early.
When implementing waiting, check that every state transition that can satisfy
a waiter also arranges its notification. CodeRhapsody records this as a
job-system lesson; a larger timing margin does not repair a missing
notification.

**Chapter 6: progress is not an acknowledgement.** One actor coordinates
conversation changes; worker results return through its ordered input path.
A slow observer must not stop execution, which makes a lossy progress stream
unsuitable for acknowledging a particular request. Blocking calls and dependent
work need their actual request completion, not a shared idle notification.
Avoid duplicate orchestration paths that produce different lifetimes or results.
Interrupt, explicit job kill and shutdown have different meanings; the source's
remaining contradictions are listed below. Its external multi-agent exercise
does not authorize agent-created sub-agents ahead of their forthcoming chapter.
Sources: [“Ask, rebuilt”](../book/chapter-06.md#ask-rebuilt),
[§6.6](../book/chapter-06.md#66-the-wait-primitive),
[§6.7](../book/chapter-06.md#67-hints-and-interrupts).

**Chapters 7–8: streaming and replay must identify the same content.** A parser
knows which deltas belong to a part; consumers must not guess that mapping from
its eventual position. Preserve identity through finalization and across the
scope where clients retain artifacts. Chapter 7's IDs are response-scoped;
Chapter 20 found cross-response collisions, so response-local IDs alone cannot
key a persistent GUI. Finals are authoritative, deltas transient. Flush the
terminal incrementally: buffered correct deltas can still look like no streaming.
On reconnect, rebuild completed content and current partials; do not silently
drop irreplaceable replay messages as though they were replaceable progress.
Test a second turn and a restart, not only the first streamed reply. Sources:
[§7.4](../book/chapter-07.md#74-part-ids-who-is-allowed-to-name-a-part),
[§7.9](../book/chapter-07.md#79-the-terminal-and-the-flush-that-makes-it-real),
[§8.7](../book/chapter-08.md#87-two-tiers-of-state),
[§20.3](../book/chapter-20.md#203-two-paths-one-screen).

**Chapters 8–9 and 13: visible controls need observable, effective state.**
A setting may appear changed without being saved, or be saved without affecting
execution. Follow it to its consumer and display the effective state. Validate
both patches and loaded settings; preserve the meaning of absent fields versus
zero/false. Expose selected/expanded state semantically so assistive and automated
users can observe it. Render untrusted tool text as text; disclose truncation and
leave omitted material reachable. Chapter 13's source-blind GUI driver found
bugs source inspection missed, but also invented a panel when state was hidden.
Record what was observable and distinguish operational help from supplied
answers. Sources: [Chapter 9, “Settings over the wire”](../book/chapter-09.md#settings-over-the-wire),
[Chapter 13's settings finding](../book/chapter-13.md#the-bug-that-only-appears-after-you-fix-the-bug),
[“What an interface owes an observer”](../book/chapter-13.md#what-an-interface-owes-an-observer),
[§20.2](../book/chapter-20.md#202-the-bugs-that-succeed).

**Chapters 8 and 14: preserve the person's chance to intervene.** Serial
dispatch supplies a boundary where pause can stop the next tool. Pause must
remain interruptible. Speech needs its own evidence: a correct DOM proves
nothing about phrase boundaries, missing errors, a complete response delivered
without deltas, or duplicated streamed finals. Test the transcript and listen;
the grader cannot judge intelligibility. Tool results intentionally remain
unspoken under the Chapter 14 ruling. Bill separately settles ownership:
**pause is Agent-wide; explicitly unpausing in any tab unpauses everyone.**
Local speech/typing causes do not create independent tab vetoes. Sources:
[§8.3](../book/chapter-08.md#83-one-tool-at-a-time),
[§14.4](../book/chapter-14.md#144-the-run-that-succeeded-for-the-wrong-reason),
[§14.6–14.8](../book/chapter-14.md#146-the-bug-nobody-could-hear).

**Chapters 10, 12–13 and 21: capability visibility must agree with dispatch.**
Skills progressively expose instructions and tools; loading is not a license to
publish every tool advertised by an external server. Enforce enabled tools at
dispatch as well as in declarations, including guessed names, and actually remove
them when the applicable lifecycle calls for it. Distinguish declared names from
normalized lookup keys: Chapters 20 and 23 show that mixing them silently hides
or preserves underscored tools. Keep MCP behind its transport seam and existing
job/cancellation ownership; preserve useful transport and tool-error diagnostics.
Test connection, discovery, invocation and unload through the real client path.
Sources: [Chapter 10, “The Registry”](../book/chapter-10.md#the-registry),
[Chapter 12](../book/chapter-12.md),
[Chapter 13, lifecycle](../book/chapter-13.md#skill-based-mcp-lifecycle),
[Chapter 21, security gate](../book/chapter-21.md#the-tools-header-is-a-security-gate),
[Chapter 23, §23.9](../book/chapter-23.md#239-what-does-not-work).

**Chapters 11 and 15: recovery needs an anchor and honest failure semantics.**
A snapshot plus a log needs an `AsOf` boundary so restoration neither duplicates
history nor loses its tail. Compare next vendor requests from snapshot and full
replay, including a deliberately older snapshot with newer events. Runtime
configuration wins over historical configuration; secrets do not belong in the
save. Refuse a broken existing save rather than silently starting fresh and
overwriting it. Chapter 15 adds journaling and ordered snapshot/truncation for
process crashes; it expressly does not promise power-loss durability. Its
known-event diagnostic-and-skip behavior requires validation before mutation;
unknown interpreted event types still fail loudly. Sources:
[§11.3–11.6](../book/chapter-11.md#113-snapshot-plus-tail),
[§15.10](../book/chapter-15.md#1510-compaction-is-described-never-performed),
[§15.12](../book/chapter-15.md#1512-crash-safe-persistence).

Keep Chapter 11's save/load exercise small: its four-field JSON save and restore
loop are sufficient starting points. Byte-identical replayed **vendor requests**
do not require a custom canonical JSON format on disk. Do not bring Chapter 15's
journal forward or add the retired attempt's hash/origin acceptance machinery.
The [first-attempt history](edition-2-attempt-1.md) explains why that expansion
is specifically a failure to avoid.

**Chapters 15–17: preserve meaning when discarding bytes.** A loaded manual
stored only as a tool result disappears when tool results are cleared. Give
surviving instructions, handoffs, memory and recall their own appropriate kinds;
keep calls/results paired and defer destructive changes until the batch is
complete. Removal fits recoverable tool bytes; irreplaceable conversation needs
preservation through the specified memory mechanism. Record the actual cuts and
nondeterministic compressed/recalled text when they occur, so replay needs
neither filesystem guesses nor new model calls. Do not adopt old job-output
addresses as durable facts when handle reuse can make them name different bytes.
Sources: [§15.4](../book/chapter-15.md#154-what-survives-is-never-a-tool-call),
[§15.11](../book/chapter-15.md#1511-keep-the-words-let-the-bytes-go),
[§16.1](../book/chapter-16.md#161-in-plain-words),
[§17.5](../book/chapter-17.md#175-attachment-permanent-until-a-checkpoint).

**Chapters 16–17: use the smallest mechanism and test useful recall.** The
compressor and relevance judge are bounded calls, not prematurely built
sub-agents. One checkpoint trigger avoids divergent clearing paths; failed or
abandoned compaction must leave old information available. Re-enable memory from
disk so human corrections matter. A tiny synthetic BM25 corpus can make valid
scores look broken; test representative scoring and fallback. Let the component
parsing a judge reply own its requested format, or duplicate instructions can
force fallback silently. Hook recall into every actual turn entry path. Fixed
fake summaries prove cascade structure, not real compression quality. Sources:
[§16.2](../book/chapter-16.md#162-one-door),
[§16.5–16.10](../book/chapter-16.md#165-the-compressor),
[§17.2](../book/chapter-17.md#172-bm25-the-thirty-year-old-formula-that-still-works),
[§17.4–17.5](../book/chapter-17.md#174-the-judge-filtering-by-meaning-not-keywords).

**Chapters 15 and 18–20: instrument costs independently of correct answers.**
A cache miss can return a perfect answer. Stable-prefix design therefore needs
observations of real request artifacts and provider grants. Compare actual
serialization, distinguishing append from rewrite and marker metadata from
content; a test and instrument sharing the same invented format can certify a
broken instrument. Change one experimental variable at a time. Distinguish
bytes from tokens and last-response figures from session totals. Preserve counts
rather than computed money, show unknown prices honestly, and attribute counts
to the model that incurred them once switching exists. Chapter 22 corrects
repricing all previous work at the current model's rate. Sources:
[§15.3](../book/chapter-15.md#153-two-laws),
[§18.2](../book/chapter-18.md#182-the-lens),
[§18.5](../book/chapter-18.md#185-the-meter),
[§20.4](../book/chapter-20.md#204-what-you-cannot-see-cannot-be-fixed),
[Chapter 22, accounting](../book/chapter-22.md#where-the-money-went-wrong).

**Chapters 19–20: authentication, effective routing and billing move together.**
A successful login is not evidence of inference permission; an answered request
is not evidence of the intended credential or funding source. Keep credential
renewal out of conversation/rendering concerns, serialize rotating-token refresh,
and distinguish terminal rejection from transient failure. Observe credential
kind without exposing tokens. On a model/vendor switch, check the complete
operation, including credential and billing route. Metered fallback requires its
explicit authorization and announcement before spending. Stream reasoning
summaries distinctly and incrementally so a listener can steer while work is
happening. Historical provider details require verification at implementation
time; the source itself records route differences. Sources:
[Chapter 19, identity](../book/chapter-19.md#identity-is-not-permission),
[refresh](../book/chapter-19.md#tokens-that-expire),
[billing](../book/chapter-19.md#the-billing-state-machine),
[§20.2](../book/chapter-20.md#202-the-bugs-that-succeed).

**Chapters 21–23: absence and enforcement need positive controls.** Retrieved
web text is evidence, not authorization. A skill's tool filter limits capability;
it does not establish that the model will ignore malicious instructions. Keep
credentials outside model-visible content from their first use. When Chapter 23
introduces confinement, in-process file operations and spawned descendants need
their separate enforced boundaries. Canonicalize both roots and candidate paths;
provide confined operations so callers do not accidentally check one path and
open another. Narrow child specifications unconditionally and report attempted
widening separately. Test allowed work beside rejected escapes: a sandbox that
denies everything passes negative-only tests. Verify that removed tools really
cannot execute. Respect the chapter's macOS implementation limit rather than
silently running unconfined elsewhere. These are Chapter 23 mechanisms, not
permission to invent the missing sub-agent exercise. Sources:
[Chapter 21, trust boundary](../book/chapter-21.md#a-web-page-cannot-authorize-a-command),
[§23.3–23.7](../book/chapter-23.md#233-two-boundaries-not-one),
[§23.9–23.10](../book/chapter-23.md#239-what-does-not-work).

## Coverage of the two architecture-repair chapters

Bill intends to absorb **Chapter 5, The Big Refactor**, and **Chapter 22,
Architectural Decay**, into the appropriate earlier teaching in the new edition.
Chapter 20 is **The Crossover**, not the second repair chapter. The student still
follows the original numbering and exercises; this future editorial consolidation
does not delete first-edition history or skip required capabilities.

| Original lesson or capability | Carry forward and eventual placement |
| --- | --- |
| Ch5: star imports, shared interfaces, behavior in the right package, owner access and no mutable globals | Apply from the first code; covered in the ownership guidance above and mandatory skill. |
| Ch5: an independently built application imports the framework, registers a custom tool and receives its result through the real tool loop | Preserve the external-consumer exercise when public tool registration is taught. An import that merely compiles is insufficient. Existing tool/job behavior must still work. See [§5.6](../book/chapter-05.md#56-the-exercise-graded). |
| Ch22: the running binary actually uses the library's composition root; constructor and logging checks test relationships, not names | Apply throughout; the workflow's architecture review and mutation checks guard the real wiring. Do not recreate a flat root merely to refactor it later. |
| Ch22: a tool dispatch context reaches Engine through a common interface and follows its parent accessors | Establish this when tool dispatch arrives. Expose the owning Engine, not copied usage/model values or a special status callback that creates a second route. See [Ch22 TL;DR](../book/chapter-22.md#tldr). |
| Ch22: per-model usage and cost, without repricing old work after a model switch | Apply when accounting/model switching arrives; covered in the cost guidance above. |
| Ch22: `agent_status` reports current model, last response usage, session totals, cache hit rate and cost through the ownership chain | Preserve this feature and its live model-switch exercise. The later author–coder phase assigns its place alongside tool/usage teaching; it is not a Chapter 1 feature. |
| Ch22: disconnect can race with observer sends; superficially defensive code and timing-heavy tests can both hide the defect | Apply when GUI connection lifetimes arrive, as described below. |

Chapter 22's [disconnect failure](../book/chapter-22.md#the-crash-the-decay-concealed)
teaches that a `select` with `default` does not make sending to a closed channel
safe. Coordinate teardown with all senders; the demonstrated solution used a
separate completion signal instead of closing a data channel still reachable
by senders. Race regression tests must expose the relevant interleaving and fail
when the defect is restored, not merely repeat a timing lottery. Preserve this
lesson without requiring the first edition's exact client counts or test layout.

## Contradictions and recommendations for Bill's review

Already settled by Bill: actual-parent ownership supersedes sibling dependency
examples; root logging supersedes Chapter 5's per-Agent logger wording; private
runtime implementations are allowed; common is not a behavior collection; GUI
implementation is external; pause has no per-tab veto. Examples using globals,
flat `cmd/` wiring, common-package services or callback bags do not reopen these
decisions. Later repair chapters need not recreate their motivating defects.

The topology document's `NewAgent(cfg, spec)` sketch omits a parent argument;
Bill's constructor back-pointer rule still governs the student implementation.
Its suggested `AgentSpec` shape and “defaults plus overrides, probably” wording
are design sketches, not mandatory fields or a new configuration subsystem.
Its first-edition Chapter 22 migration scope does not tell the fresh student to
recreate the old layout before correcting it.

The following remain source discrepancies, not newly imposed requirements:

| Source tension | Recommended reading or decision still needed |
| --- | --- |
| Chapter 6 TL;DR/§6.5 stop the actor and discard results on interrupt; §6.7 preserves actor lifetime and real non-killed outcomes. §6.10 uses `Post`/observation `Wait` where §6.6 requires reliable completions. | Confirm the corrected lifecycle/completion paragraphs govern. Keep request completion distinct from progress. Do not silently follow the older pseudocode. |
| Chapter 6's `PartFinal` example is text-only although its contract includes tool calls; its hint language alternates between same-event classification and a distinct hint event. | Preserve observable content and Agent-owned classification; settle any externally consequential event/protocol ambiguity at that exercise. |
| Chapter 8 §8.3 deliberately dispatches serially for pause; Chapter 15 §15.4 says tools run in parallel. | Preserve serial dispatch and the batch-completion invariant; do not add parallel execution to satisfy explanatory prose. |
| Chapter 8 calls replay-window configuration both current and future, and its GUI-log example is not the JSON-lines shape named by its check table. | Recommend a fixed window for that exercise; confirm the actual required log format if the original grader disagrees. Record the result without changing the grader. |
| Chapters 10 and 16 put mutable memory in history as data; Chapters 17–18 still refer to recent memory/system-prompt placement. Chapter 17 excludes recent files on that premise. | Recommend preserving memory-as-data. Resolve the recall corpus exclusion against what context actually contains, rather than copying an obsolete placement assumption. |
| Chapter 17 TL;DR allocates quotas to candidates; §17.3 says final mechanical selection, without overruling a successful judge. | Recommend §17.3's explicit correction; this changes observable recall selection and needs Bill's agreement. |
| Chapter 19 requires explicit plan-route caching; §20.4 reports that route rejected it. Chapter 19 also names quota-exhaustion errors differently in its TL;DR and billing section. | Verify the current supported route and error behavior when reached. Document conflicts with original grading; do not claim historical wire details are current, or trade billing safety for a pass. |
| Chapter 12 requires MCP cancellation but its bridge example supplies `context.Background()`. | Preserve cancellation through the actual job lifetime; the sample cannot override the stated requirement. |

**A useful recommendation from deleted Chapter 1, §§1.7–1.8, for Bill's review:**
keep a failed HTTP exchange from masquerading as a completed conversation pair.
For the initial text conversation, staging the new pair until a valid reply
arrives avoids poisoning the next request after a failure. Missing usage is not
measured zero usage. Use bounded waits and secret-free diagnostics that preserve
useful timeout/cancellation distinctions for library callers. These are proposed
implementation lessons, not restoration of that draft's exact CLI exit rules or
additional pass/fail matrix. Later event logs may properly record failed requests;
recording a failure and claiming successful completion are different things.

Several editorial inconsistencies also deserve later correction, without blocking
unrelated exercises: Chapter 5's check totals are unreconciled; Chapter 9's patch
example omits its stated validation; Chapter 13 contradicts itself about zero
serialization and says a quotation is absent from Chapter 9 when this snapshot
contains it; Chapter 15 says no live session exists before describing one.
Skill-unload passages evolve across Chapters 10, 13 and 15: distinguish prompt
text retention from executable tool removal. Treat model narration about
recoverable stubs as narration, not proof of a durable-address guarantee.

Chapter 23's next-chapter promise and the epilogue's future-sandbox discussion
are stale sequencing. Bill reports CodeRhapsody is writing Chapter 24, a GUI
extraction/refactor; it was **not available in the reviewed snapshot**. The
sub-agent chapter is also forthcoming. Neither has been read or completed, and
neither changes Bill's already-confirmed module boundary. Read new chapters and
record revised coverage before applying their lessons. Provider names, prices,
limits and capability claims throughout this review are historical source
claims; no provider access or external verification was performed here.

## Reading coverage and later use

Codex read [the topology decisions](ensemble-topology-decisions.md) in full on
October 9, at file revision `70acadd09115ec0543dd39cec3a3d7196dd15295`.
The document records October 6 decisions with Bill alongside recommendations
and migration notes. Its ownership and identity lessons are incorporated above;
its line counts, implementation-status statements and type-promotion inventory
are historical observations, not measurements of the new student implementation.

At Bill's request, Codex read CodeRhapsody's `~/.cr/SOUL.md` and
`~/.cr/MEMORY.md` in full on October 9 (815 and 3,159 words respectively).
They were found directly under `~/.cr/`, not its `memory/` subdirectory.
The verification, small-edit and wakeup guidance above draws on these notes;
collaboration and author-facing lessons are in the workflow. These are
CodeRhapsody-authored recollections, not independently verified incident reports
or additional statements from Bill. Only the distilled guidance belongs in the
student handoff. The raw notes include old roles, architecture, personal context
and tool-specific rules that must not silently replace the current assignment.

Codex also read `~/.cr/memory/summary-7d.md`, whose recorded update date is
April 11, 2026. Its decisions, rejected approaches and active-context sections
informed the workflow's continuity guidance. Its old implementation status and
product recommendations are not current facts or new student requirements.
Bill clarified that the summaries he meant are the `short` entries in
`~/.cr/learnings.json`. Codex read all 65 of those entries on October 9, without
reading their expanded descriptions. The verification guidance above and the
workflow's check for misplaced behavior draw on them. Personal details,
unrelated project instructions and historical product claims were not imported.

Codex also read the deleted second-edition Chapter 1 in full from
`0f05359c9240e76213fc0d3937306f81094f5295:book/edition-2/chapter-01.md`
on October 9. This was a historical gap check, separate from the first-edition
reading below. Its useful ownership clarifications, ephemera motivation and
failure-handling recommendation are recorded above; practical review examples
and source-tree guidance are in the workflow. Its old skill path, permission to
copy a reference answer and expanded acceptance matrix are not reinstated.

All 25 available manuscript files were read in full, including code examples:

| Reader | Complete coverage |
| --- | --- |
| `carryover_reviewer/read_early` | Chapters 00–07, 5,399 lines |
| `carryover_reviewer/read_middle` | Chapters 08–15, 3,017 lines |
| `carryover_reviewer` | Chapters 16–23 and epilogue, 5,327 lines; synthesis and direct checks of cited passages from both helpers |

Reviewed October 9, 2026 from an immutable working-tree snapshot based on commit
`bd8e35b4ea9dbff870f82015e4c2b2afc2f00ee6`, **including uncommitted edits**.
All files matched that commit except Chapter 23, whose reviewed SHA-256 is
`3b19960c74cf4d01e225b9c60bc175b199cc95ec01cad754b97d9a3dcf4bd061`.
The aggregate SHA-256 of ordered `book/path + space + file-SHA256 + newline`
records, Chapters 00–23 then epilogue, is
`82d62af93062b0543ae66bf6d50e6586c25fd35f3856ef53bf642cb47b7b43b6`.
All 25 hashes were verified. Repository links above point to the evolving source,
not immutable copies. No implementation answers or grader files were inspected;
no coding, grading or live-model exercise was performed for this review.

Carry substantive discoveries into the workflow and attributed
[chapter notes](agentic-codebook-workflow.md#chapter-notes-belong-to-every-agent),
including disagreement, corrections and actual assistance. Notes are evidence
for the future author, not additions to the student's assignment. The
[retired attempt](edition-2-attempt-1.md) showed how author-written contracts and
agreeable review can inflate small exercises. Keep the author phase after the
completed implementation and comparative reviews.

At that later phase, author and prose reviewer use both
[voice.md](../book/voice.md) and
[chapter-writing-procedure.md](../book/chapter-writing-procedure.md), under the
current workflow's ordering. Preserve real stories, explain why mechanisms
exist, and judge enjoyment and understanding alongside correctness. Codex writes
a new first-person epilogue only when the Edition 2 manuscript is ready,
grounded in actual outcomes; preserve the first-edition epilogue. A successful
regeneration is an experiment to demonstrate, not a result this draft assumes.
