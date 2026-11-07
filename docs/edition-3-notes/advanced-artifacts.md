## Book completion pass

Bill approved placement as Chapter 28 before Crossing Over (now 29), one isolated
student attempt for questions and teaching feedback, then one editorial reviewer.
This is not a reconstruction certification campaign. Student baseline 9ed49875;
completed implementation is withheld. Publication remains pending Bill approval.

# Advanced Artifacts: working chapter notes

Current authoring boundary: Advanced Artifacts is Chapter 28, followed by
Crossing Over as Chapter 29. The implementation includes the registered-tool
approval extension through fe99c2d7. The bounded student and single editorial review are complete; final local
publication candidates are built, with publication pending Bill approval.
Six authorized synthetic figures are captured; human usability review is pending; no independent
chapter reconstruction or grader acceptance is claimed. Specialized approval now
includes actual registered tools across the host's Agents, including MCP tools,
with distinct namespaced keys and explicit exact-version approval. See the final
development entry for verification and page-load discovery limits.
The dated development record below preserves earlier states.


Initial packet status (historical): [chapter packet](../../book/chapter-28.md) drafted with
TL;DR, In plain words and exercises; not an implemented migration or a finished
chapter. Develop the code and teaching together. Chapter numbering is undecided.
Foundations: [Chapter 7, Everything Is an Artifact](../../book/chapter-07.md),
especially §7.6, and [Chapter 26, Artifacts and Actions](../../book/chapter-26.md).
The earlier [Artifacts/Actions notes](artifacts-and-actions.md) retain their own
scope and evidence; this work does not reopen their grading campaign.

## First experiment: shared file-reading ownership and shipped preview

Bill approved a `file-reading` dependency as the shared home for `read_file`,
`list_directory`, and `search_files`. The `ensemble`, `full`, and `local-coding`
primaries depend on it instead of duplicating those grants. Exact effective
permission sets remain unchanged; the first enabling skill is now `file-reading`.

The first renderer lives at
`agent/skills/file-reading/renderers/list-directory.js`. An explicit reviewed
build resource and HTML script reference load it before `app.js`; merely finding
or loading another skill does not execute its JavaScript. GUI and CLI outputs
copy the readable source beside the manual. This is a bounded experiment, not
final renderer-declaration syntax or an approval/update mechanism. Editing the
copied file does not update the embedded, frozen GUI asset snapshot.

The renderer receives the existing bounded display text and returns a detached
DOM element. It labels literal text as directory output without splitting lines
into purported entries: filenames can contain newlines. It uses `textContent`,
not HTML interpretation. The host keeps identity, call/result pairing, raw-detail
controls, timing, selection deferral and placement. This preview slot does not
make a card or Tools pane part of the eventual renderer contract. Missing,
throwing or invalid renderers use the plain preview; errors, calls and unpaired
results do not use the directory renderer. This is trusted in-page JavaScript,
not a security sandbox.

A structured directory payload remains separate work. The current text format
and observation envelope cannot reliably identify entries or distinguish an empty
directory from an observation with no output. Preserve operation evidence through
Jobs, History and replay before claiming richer directory semantics. The existing
GUI's text normalization and observation truncation are not repaired by this hook.

Focused checks cover unchanged primary grants; versioned, mounted asset serving;
desktop/phone literal text and fallback; retained DOM/focus; and existing Actions
state, pairing, selection, scroll and timing behavior. Browser cases inject recorded
facts into the actual page handler, not a live directory call. An initial phone
focus assertion failed because the test had not opened the Tools pane; explicitly
opening it and requiring focus before rerender corrected the test. No production
focus change was necessary. Core377/0 and all six default GUI groups pass;
GUI and CLI builds have zero warnings/errors, with byte-identical manual and
renderer copies in both outputs. Versioned-asset and final browser checks pass,
including invalid/connected-node fallback. Owned fixture hosts exited normally.
Evidence is under `/private/tmp/cadence-advanced-artifacts.7bNped`; it is local
verification, not a portable release artifact. No installation or human visual
acceptance is claimed.

## Renderer trials and acceptance

Bill approved version-specific review and proposed separate approval to try a
renderer live and acceptance to keep it. Reload, not application restart, is the
recovery boundary. A candidate must not overwrite accepted bytes. An explicitly
approved trial belongs only to the current page's memory; neither browser storage
nor host startup state may automatically resume an unaccepted trial. A fresh page
loads the accepted version, or the shipped default when none exists.

Acceptance must name the exact reviewed code and executable dependencies, not a
mutable directory. Changed bytes require a new trial approval. Retaining the prior
accepted version provides rollback after a later-discovered problem. These are
agreed lifecycle requirements, not implemented behavior of the first experiment.

Ordinary exceptions can fall back locally. In-page JavaScript can also hang the
page or change unrelated state; a same-page watchdog cannot guarantee recovery.
Reload discards trial selection, but does not undo arbitrary side effects. This
is a recovery mechanism for reviewed code, not containment of hostile code.

Inspection confirms that `Gui.MapAssets()` freezes its asset bytes at host startup;
the current shipped hook cannot load edited source live. GUI `Preferences` is an
appearance/user-information value record, not an executable-code owner. The next
implementation must separate immutable version storage and durable acceptance
from page-local trial selection, preserving the existing fixed asset boundary.
Exact storage ownership, candidate admission and UI controls remain to implement.

## Version storage implementation slice

`Gui/Common/RendererStore.cs` declares immutable package/selection values and a
GUI-operation-owned store. `Gui/Server/RendererStore.cs` stages canonical bounded
JavaScript packages under SHA-256 identities covering entry, renderer identity,
and every declared helper. Reads verify the hash. Staging does not activate code.
Acceptance checks the expected current selection, validates the target snapshot,
and saves before returning success. A prior accepted identity supports rollback,
including rollback to the shipped default. There is no durable trial field.

Selection mutation uses a cooperating-writer file lock and sibling temporary
write/flush/replace. This is not arbitrary-writer protection or a directory-fsync
crash guarantee. No browser route, candidate import UI, or loader uses this store
yet. Package declarations do not statically prove the absence of undeclared
JavaScript imports or external behavior; trusted-code review remains necessary.

Focused public-API store tests pass with a clean GUI-tests build (zero warnings
or errors). They cover dependency-edit identity, reopen without candidate
activation, acceptance/rollback, stale and locked promotion refusal, tool identity,
traversal/size limits, tampered snapshots and duplicate selection fields. The
initial harness failed because its expected-refusal helper omitted
`InvalidDataException`; the intended hash refusal occurred. Correcting that helper
and rebuilding passed. A pre-encoding source-byte bound was added during the
initial build; only the subsequent sequential build certifies final source.
Evidence: `/private/tmp/cadence-renderer-versions.gOOJ6z`. These checks do not yet
establish browser reload recovery or executable approval controls.

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

## Artifacts do not require a shared card format

Bill clarified that stable Artifact identity and content do not imply a common
visual container. A card is one presentation, not a mandatory Artifact shape.
Likewise, tool-result origin does not require placement in a Tools pane. The
host application determines the available presentation surfaces; Homebrew-VTT
has only a chat pane and includes custom-rendered tool results there.

Concrete case: a dice-rolling tool returns a total that appears as a compact
clickable or hoverable element in chat. Interaction explains how the total was
computed from rolls and modifiers. The explanation should also be reachable by
keyboard and touch rather than relying on hover alone. Display the recorded roll;
revealing its explanation must not reroll or fabricate a calculation.

Preserve identity, provenance, original content and lifecycle independently of
layout. Shared Markdown, expansion and other UI helpers should be available for
reuse without imposing a heading, border, status row, speech button or standard
card layout on every Artifact. The trusted host still owns access and execution
authority, but the exact presentation interface must accommodate inline output
and application-specific interaction. Do not turn today's Ensemble card DOM or
two-pane routing into universal renderer requirements.

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

## Agreed first security boundary: trusted in-page renderers

Bill accepted trusted in-page renderers as the first implementation. Ensemble is
for his supervised workflow first: he listens to the engineering explanation,
reviews the work with Cadence and explicitly accepts this trust model. Do not
require an iframe for renderers we develop and approve together. Supervision is
review and detection, not a claim of technical containment or perfect security.

Distinguish Agent execution confinement from browser execution authority. A
sandboxed Agent can produce JavaScript that would gain the main page's authority
if the host later executes it there. The Agent's process restrictions do not
follow the code into the browser. Likewise, autonomous authorship is not itself
a trust classification: review and approval matter, including for downloaded
skills and their dependencies.

Trusted renderers can access the main page and potentially other displayed data,
controls and authenticated application endpoints. This is an explicit trusted
extension boundary. Do not automatically execute downloaded or otherwise
unapproved renderer JavaScript; use the safe default presentation until approval.
The exact approval/configuration mechanism remains to be designed. Approval of
one renderer does not imply approval of arbitrary future replacement bytes.

Tool-result content remains untrusted data even when its renderer is trusted.
Preserve safe Markdown and text handling: a filename, fetched document or MCP
result must not become executable markup merely because a trusted renderer
handles it. Renderer access to application operations remains an explicit design
decision, not an automatic consequence of formatting a result.

### Deferred option: isolated renderers

Iframe isolation is not required for the first useful Artifact implementation.
Keep the renderer interface based on supplied Artifact data so an isolated
implementation can be explored later, without prematurely building a second mode.
Do not advertise safe execution of arbitrary Internet-sourced renderers now.

A future isolated mode would require a deliberate sandbox, restricted resource
and network policy, and a narrow validated messaging interface. A bare iframe
is not a complete security boundary; in particular it does not by itself prevent
network exfiltration. A generic tool-execution bridge would undermine isolation.
An opaque-origin script frame would avoid granting `allow-same-origin`; exact
sandbox/CSP/host policies need implementation and testing before security claims.
Iframe isolation also does not establish that an executable/MCP server in the
same skill is safe, or guarantee freedom from resource exhaustion or deceptive UI.

A bounded iframe prototype is feasible, but a polished general-purpose mode
adds sizing, focus, keyboard/touch accessibility, theme, clipboard, streaming,
lifecycle and message-protocol work. Inline dice explanations expose a concrete
tradeoff: a tooltip inside a frame cannot freely extend outside its bounds;
resizing or host-rendered overlays need additional integration. These are reasons
to defer the mode, not a claim that isolation is impossible.

Lesson: trust belongs to code at the place where it executes. The first chapter
implementation should clearly describe trusted extensions rather than imply
that Agent sandboxing also contains browser renderers.

## Connection to workflow support

Bill's intended definition: workflows are executables that use reverse MCP to
orchestrate sub-agents of the Agent that called the workflow. They can coordinate
writing, reviewing and editing, as in Bill's existing book-production workflows.
The shared skill/library layout supports these executable harnesses as well as
renderers and ordinary tools. A separate workflow scheduler or unrelated Agent
ownership system is not implied.

Inspect the actual reverse-MCP interfaces and caller/child ownership before
choosing a harness API. This records the direction, not a claim that an arbitrary
workflow executable is already supported or tested. Start from one of Bill's real workflows rather than speculative architecture.
Bill expects a small upgrade and considers it the remaining book work after
Advanced Artifacts; it may belong in the Skills chapter instead of a separate
chapter. Decide placement from the actual implementation.

## Tool-formatting direction from Bill

Start with the original six tools, with `wait_for_job` treated as a continuation
of its underlying tool rather than a separate generic output format.

- **run_command:** ANSI-formatted output resembling Bill's dark-theme terminal.
  Investigate an existing JavaScript library. Compare a display-oriented ANSI
  renderer with a terminal emulator; do not assume interactive terminal input is
  required. Current GUI stripping of escape sequences is not ANSI rendering.
  Handle control sequences deliberately; do not interpret arbitrary escape output
  as permission for clipboard, navigation or application actions.
- **edit_file:** a readable visual diff, using a suitable existing library if
  available. Distinguish the actual applied replacement from a whole-file diff;
  correct original/resulting positions require recorded operation metadata.
- **read_file and write_file:** nicely formatted code/text with line numbers and
  language highlighting selected from file suffix where supported. Explore
  sharing highlighting with the diff renderer rather than duplicating libraries.
  Read ranges must retain actual file line numbers, not restart at one.
- **write_file append:** append already exists in `Tools/FilesAndShell.cs` and
  uses `FileMode.Append`. Display resulting file line numbers for the appended
  content. An unterminated final line may be continued rather than starting a new
  line. The current plain success string does not supply that starting position;
  establish it from the actual operation rather than guessing in the renderer.
- **wait_for_job:** Bill wants both model-facing results and GUI output to retain
  the original tool's format. It puts the same Job back in the foreground; it
  should use the original Artifact where possible, especially for long commands.
  Preserve distinct recorded observations and call correlation while designing
  a continuing visual presentation. Do not turn another foreground wait into a
  second execution, silently equate an observation with Job completion, or replay
  already consumed output. Jobs retain their originating `ToolCallPart`; inspect
  how that provenance can reach rendering rather than dispatching only on the
  observer control's name.
- **list_directory:** Bill selected it as the first implementation experiment.
  Keep the listing compact, preserve returned entries and distinguish directories.
- **search_files:** Bill accepted grouping matches by file, with actual line
  numbers, highlighted matches and surrounding context. Preserve returned content
  and make truncation explicit.

No new rendering libraries have been researched, selected or downloaded for this
work. Prefer discovering/reusing existing components over writing a new ANSI,
diff or code-highlighting engine. Compare API, licensing, size, accessibility,
safety and integration only to the extent needed for the chosen experiment.

## Continuing Job Artifacts and the display window

Bill approved keeping the **most recent 64 KiB** of command/Job output in a
scrolling display. This is a GUI display budget, separate from model-facing
observation limits and from retained output on disk. Clearly indicate when older
content has rolled out of the display; do not imply the visible window is the
complete output. Bound by bytes and preserve valid text boundaries rather than
silently interpreting 64 KiB as 65,536 characters.

Bill wants `wait_for_job` and `send_input` to foreground the original ongoing
Artifact and move it to the bottom of the Tools pane. A debugger such as gdb
should read as one continuing session rather than many disconnected snapshots.
Move the existing presentation on explicit foregrounding, not on every background
output update. Preserve focus, selection and internal scroll behavior; following
the tail should not prevent the human from reading older visible output.

Use actual Agent/Job identity to associate the continuing view. Keep underlying
calls and observations as separate recorded facts, and reconstruct the combined
view on replay. Showing submitted input must account for terminal echo so it does
not accidentally duplicate commands. A completed observation is not necessarily
a completed Job. Earlier output remains in the retained Job output, subject to
its existing ownership and retrieval rules.

A GUI window cannot recover bytes already omitted from model-facing observation
text. Investigate a bounded output feed from retained Job output without consuming
the model's observation cursor or silently raising its context budget. A feed is
not implemented yet. Likewise, `write_file` content comes from call arguments;
distinguish the proposed/streaming content from a confirmed successful write.

With ANSI output, retain enough parser state for a rolling window to display
correctly when an escape sequence or styling state crosses an update/window
boundary. Exact terminal-control handling and retained-output presentation still
need a bounded library experiment; the approved 64 KiB choice is not a claim
that every terminal protocol can be handled by slicing raw strings.

## Open design decisions and next step

- Choose explicit renderer declarations, dispatch identity and library references.
- Separate original content and identity/lifecycle from renderer-owned layout
  and host-selected placement. Shared UI helpers are optional, not a mandatory
  card shell. Preserve applicable streaming, replay, copying, speech, selection,
  focus, scroll position and accessibility behavior.
- Define authority/isolation for renderer JavaScript and tool executables. A
  presentational control must not silently gain tool execution or file authority.
- Define packaging, version/update behavior, unavailable-library and renderer
  failure fallback without introducing a second authoritative conversation store.
- Verify the existing skill packaging and asset-serving paths, then choose one
  built-in tool and shared Markdown helper for a bounded first experiment with Bill.

These decisions should be earned through that experiment. Do not turn the notes
into an expanding acceptance campaign or implement workflows merely because the
layout will support them.

## Initial chapter packet and writing checks

The reader stake is understanding and steering the Agent's work without decoding
raw tool dumps or mentally reassembling a continuing session. This is new teaching
built on Chapters 7 and 26, not a rewrite of an earlier Advanced Artifacts chapter.
The discussion's directory customization, Homebrew-VTT dice display, append line
numbering and debugger continuity provide concrete examples. No successful
implementation story or benchmark is invented.

Read `book/README.md`, the current-method portions of the local chapter procedure,
the full Teach Your AI creation procedure, the full book voice guide and the
external generic voice guide before drafting. Bill's current direction is personal
interactive writing and implementation; this draft starts no student campaign.
The packet is deliberately unnumbered and excluded from `chapter-NN.md` assembly.
Renderer/library declarations, approval/update policy and output-feed interfaces
remain explicit design decisions for the first experiment. Six exercises capture
observable agreed behavior, not a claim that an independent build contract or
external graders are already complete. The remaining chapter body is unwritten.


## Executed renderer trial and recovery slice

The GUI host now explicitly supplies `RendererDirectory` under the application's
shared data directory. Embedded hosts leave it null to retain shipped rendering
only. `ServeAsync` owns the store and maps its HTTP endpoints into the existing
origin-guarded, joined mount lifetime. This is not Agent settings or a new server.
The bounded first API supports `list_directory`; general skill declarations and
other renderer kinds remain to be implemented.

A candidate upload is JSON with `entry` and `files`, a map of relative JavaScript
names to source. `main.js` may import a packaged `./helper.js`; all declared files
participate in the snapshot hash. Staging stores bytes without executing them.
Routes refuse unknown/duplicate fields, oversized input, unsafe names, invalid
snapshots, and cross-origin mutations. Mutations require Origin and the renderer
header; the header is not a defense against trusted same-page JavaScript.
Module requests resolve only snapshot members, never caller filesystem paths.

Speech & Display contains candidate upload, literal source review, Try live,
Keep, reload-to-revert, and prior-accepted rollback. A renderer entry exports a
default function `(document, text)` returning a detached HTMLElement. This remains
the bounded directory-preview interface, not a final all-Artifact lifecycle API.
The browser stores trial state only in lexical page memory. Keep compares the
accepted hash observed by that page; stale or unconfirmed writes are not silently
retried. Reload without Keep loads the accepted version, never the candidate.
A `renderers=shipped` query bypasses accepted custom imports for recovery after an
accepted defect. It changes no durable selection; rollback is still available.

Reload is an escape from page-local trial selection, not a promise to undo every
side effect of trusted JavaScript. An import can throw after side effects, and a
renderer can hang the page. No in-page watchdog or sandbox is claimed. Review
must include undeclared imports and network behavior; snapshot membership alone
does not constrain JavaScript's authority. Accepted rendering is shared policy;
already-open pages retain their current code until reload or explicit trial.

Verification: clean GUI consumer and normal-host builds, zero warnings/errors;
public store tests; actual HTTP staging/module/origin/header/strictness/bounds/
stale/rollback/tamper checks; actual browser controls at 1280 and 390 widths with
helper import, no execution on staging, trial discard on reload, Keep surviving
reload, failed import refusing Keep, shipped-only recovery, and rollback. The
existing visible result switched renderers while raw recorded output stayed
unchanged. Six default GUI regression groups passed. All fixture hosts joined;
no provider inference, installed application changes, or restart. Evidence is in
`/private/tmp/cadence-renderer-versions.gOOJ6z`.

One edit script guessed the rendering scheduler name and stopped after writing
Settings markup but before modifying `app.js`. Inspection found `queueCardRender`;
the final event handler uses it to preserve existing deferred selection handling.
This was an editing assumption failure, not a runtime failure. The initial
storage negative test also needed an explicit InvalidDataException catch.

Next: explicit skill association and readable package discovery, then the other
agreed renderer behaviors and continuing Job output. These tests establish this
slice only; they do not complete Advanced Artifacts or its chapter body.

## Executed slice: declared directory renderer and readable source

`SkillDefinition` now carries immutable `SkillRenderer` metadata. The bounded
front-matter subset accepts an explicit mapping:

```yaml
renderers:
  list_directory: renderers/list-directory.js
```

Entries are relative JavaScript paths, bounded to 64 declarations and 256 path
characters; duplicate names, traversal, empty segments and inconsistent indentation
refuse. Metadata alone grants no tool or browser execution. Existing skills remain
valid without mappings. The file-reading dependency owns the shipped declaration.

The shipped directory source now uses the same default-export ES-module interface
as trial candidates. The GUI controller imports it relative to its own frozen,
fingerprinted asset URL before selecting accepted custom code. Import failure
leaves the ordinary text preview available. Lesson: shipped source must be a usable
candidate starting point, not require conversion from a separate registration API.

The read-only `renderers/shipped` route derives its entry from the embedded skill
manual and returns the explicitly embedded build source in inert candidate-package
JSON. It neither stages nor selects code. It does not represent optional disk
asset overrides or arbitrary downloaded skill packages. Settings offers an explicit
source download. The installed skill has a short renderer API note; GUI and CLI
copy that note alongside readable implementation. Browser loading does not place
source in model context. General helper-library discovery remains unfinished.

Verification: clean build and six skill groups including valid/invalid mappings
and no accidental tool grants; clean GUI build, HTTP source comparison against the
actual file plus unchanged selection, desktop/phone trial/reload/acceptance/recovery
checks, and all six default GUI groups. Both browser fixture hosts joined. GUI and
CLI packaged API-note bytes match source. The shared core regression completed with 377 passed and zero failed. Evidence: `renderer-discovery-*`, `renderer-skill-*`, and
`discovery-*` logs under the existing renderer-versions evidence directory. No
installation, provider inference or application restart occurred.

## Executed slice: bounded directory operation evidence

`DirectoryListing` is immutable presentation evidence alongside `ToolResultPart`
and successful returned/observed History events, not an extra model-visible Part.
The listing operation captures individual names in its existing enumeration loop;
newlines in names remain inside one entry. Confirmed directories carry that kind.
Other entries currently remain unknown: `Directory.Exists` returning false does
not prove a regular file exists. No later filesystem reread reconstructs evidence.

The display snapshot retains a prefix of at most 256 entries and 32 KiB of UTF-8
names, plus the actual total count. It reports omitted entries explicitly. The
existing text output, retained output file and model observation clipping remain
unchanged. Jobs retain this value separately from the structured-result branch,
which would otherwise bypass ordinary text clipping. The first successful terminal
observation delivers it, including an empty directory; later observations do not
repeat it. Constructor/codec bounds refuse invalid evidence. Historical records
without this optional field remain plain text. History copies and serialization
preserve evidence; provider request normalization and result redaction omit it.

The GUI projects the optional evidence as `directory_listing`. The preview API's
optional third argument supplies it without breaking two-argument candidates.
The shipped module renders literal names as separate list items, directory labels,
and shown/total/omitted counts. Legacy results keep their text preview. Metadata
identity participates in preview invalidation; raw details are unchanged.

Verification so far: clean core/GUI builds; focused real directory dispatch with
302 entries, a newline filename, a 64-byte observation cap, display truncation,
empty directory, once-only observation, recorded-event and result codec round trips,
and malformed metadata refusals. Actual-page desktop/phone checks use synthetic
frames to establish separate names, literal HTML-looking names, omission labels,
raw details, focus stability and renderer fallbacks. This is not a single end-to-end
live directory-to-browser proof. Desktop/phone trial/reload/Keep/recovery/rollback
and six default GUI groups pass; the shared core suite completed with 378 passed
and zero failed. All owned fixture hosts joined.
The zero-delay scheduling/later-wait branch is not deterministically forced by this
slice's tests; do not claim scheduler luck proves it.

Failures: the new negative test initially used nonexistent EventKind.UserAdded;
compilation refused before testing, corrected to MessageReceived. The first browser
wrapper incorrectly invoked the independently hosted trial script with only an
ambient URL; its missing CLI arguments refused before starting its fixture. The
wrapper's own host joined. Correct explicit SDK/assembly invocation passed. Failed
log retained as directory-browser-invocation-failure.log in the renderer evidence
directory. No installed application, provider inference, or live service changed.


### Written ranges belong in model results too

Bill rejected preserving write-only append at the cost of unknown line positions.
Counting the existing lines is authorized; the successful model-facing result must
report the written range, not reserve that information for the GUI. The initial
unknown-position fallback proposal was not adopted.

`write_file` now uses one read/write OpenOrCreate handle for append, counts lines
with a bounded character buffer, seeks to the end, and writes through that handle.
Ordinary creation still uses CreateNew; overwrite still uses Create. The writer
closes before success is returned. Positions are one-based, columns count UTF-16
code units, and LF advances the line (matching read_file). A final newline belongs
to the line it terminates; empty content claims no written range. This avoids a
second path lookup but does not lock out unrelated concurrent writers. Typed GUI
file evidence and its renderer are still pending.

Verification: isolated build passed with zero warnings/errors; the focused real
file/shell tool group passed, including ordinary write, overwrite, both append
positions, empty content, CRLF/Unicode, new-file append and refusal. All four
sandbox groups passed, including rooted tool append and the existing symlink-swap
check. This is not a concurrent-writer guarantee or a GUI-rendering check. Evidence:
`write-position-build.log`, `write-position-tests.log`, and
`write-position-sandbox.log` in the renderer verification workspace.

### Bounded file-operation evidence (implementation in progress)

Read, write, append and exact edit now capture optional `FileEvidence` beside the
normal text result. The immutable value records operation kind, supplied path,
one-based start line and UTF-16 column, and bounded content. An edit also records
removed text at the unique anchor: this is an applied replacement, not a whole-file
diff. Each excerpt retains at most 32 KiB of UTF-8 on a scalar boundary and its
original byte count, so omitted content is explicit. Reads respect their requested
line range and byte cap; the display never recovers content excluded by that cap.

Evidence is captured from operation-local text, not a later path lookup. Mutation
evidence is validated before writing and attached only after the writer closes
successfully. Written-range text remains visible to the model. Jobs retain display
evidence separately from structured results, preserve ordinary observation clipping,
and deliver it once on the first successful terminal observation. Result and Event
codecs, History copies and GUI projection preserve it; it remains outside Parts.
No renderer consumes this new evidence yet.

The focused file-evidence, existing tool and directory-evidence groups passed on
an initial clean build. They cover UTF-8 bounds, invalid metadata, codec and event
numbering, real append offsets, capped reads, applied edit evidence, error exclusion,
History after the underlying file changes, and once-only delivery. The first test
fixture incorrectly attached model-response display IDs to ToolReturned; production
validation refused it. Removing that invalid setup made the intended checks run.
The shared suite completed with 379 passed and zero failed on the initial compiled
snapshot. A later review moved evidence validation before file mutation; final
focused verification passed: clean core and GUI builds (zero warnings/errors),
file evidence, ordinary tools, all four sandbox groups, directory evidence and all
six default GUI groups. The 379-test full-suite result is from before that final
ordering edit, not a second full-suite run.
No installation or live-browser behavior is established by these checks.

### Shipped file previews and shared lexical helper

The read/write/edit preview entries now consume the fourth optional renderer
argument, recorded file evidence. Directory evidence remains the third argument.
The host includes file evidence in its preview identity key and preserves raw result
text and controls. Missing evidence, errors and failed rendering use the existing
plain-text fallback. Shared `skills/shared/file-excerpt.js` shows actual start lines
and UTF-16 columns, explicit truncated-prefix notices and empty segments. Edits have
textually labeled Removed and Inserted sections, not a claimed whole-file diff.
Known source suffixes receive lexical highlighting; unknown suffixes remain plain.

The unchanged small lexer moved from app.js to skills/shared/source-code.js, loaded
before Markdown and app.js. Both Markdown and file excerpts use it. Renderer modules
are explicitly embedded and copied with GUI/CLI skill source; arbitrary skill files
are still not executed. The source-note contract documents this host helper dependency.
The trial UI and source-package endpoint remain directory-only; generalizing them
is still unfinished. This is not a complete custom-file-renderer lifecycle.

Initial actual-page synthetic-fact checks passed at 1280 and 390 widths: numbered
positions, literal HTML-like source, suffix highlighting, explicit omissions, edit
labels, empty content, fallback, unchanged raw result and stable focus/nodes. Real
operation capture is separately covered by the preceding file-evidence tests.

Final slice checks: clean GUI-tests, GUI-host and CLI builds, six skill groups,
exact GUI/CLI packaged source bytes, six default GUI groups, desktop/phone
directory and file presentation plus trial/reload checks all pass. Markdown
stream/selection/safety and code-copy browser regressions also pass after lexer
extraction. Their origin check was corrected to use the supplied random-port
fixture rather than a hard-coded port. All owned fixture hosts joined. No install,
restart, provider inference or full shared-core rerun in this presentation slice.

## Grouped search evidence and shipped renderer

SearchFiles now captures separate path, actual one-based line number, line text
and exact .NET UTF-16 match spans during the operation. The existing grep text and
merged context windows are unchanged. SearchEvidence travels through bounded Job
observation, durable ToolReturned/ToolObserved records, codecs and GUI projection,
not provider Parts. Failed observations cannot carry it. The browser receives an
optional fifth argument, preserving directory/file renderer argument contracts.

Display evidence retains a prefix of at most 256 full returned lines, 32 KiB of
path-plus-text bytes and 256 spans per line. It stops at the first line that does
not fit and retains the total returned line count for explicit omission reporting.
It does not truncate match spans into an apparently complete line. Recursive
search still skips inaccessible entries under its existing best-effort contract;
complete display does not imply complete filesystem coverage. No new regex timeout
or search-output memory guarantee is claimed by these display bounds.

The shipped skill renderer groups by literal path, shows actual line numbers and
context gaps, highlights recorded spans, labels match/context without color-only
meaning, and marks zero-width matches explicitly. It keeps raw output unchanged
and falls back for older History/errors. General trial selection remains unfinished.

Initial test compilation failed on guessed JsonCodec/root-cleanup APIs; corrected
to JsonDocument/JsonElement and existing synchronous root Destroy (joins ownership).
No tests ran in that failed build. Corrected clean build and focused real-tool test
pass: Unicode spans, path separators/newlines, merged context, capped observation,
no matches, invalid regex, zero-width, display line/byte/span bounds and immutable
replay after source mutation. Actual-page synthetic-fact checks pass at desktop
and phone widths for grouping, spans, gaps, safe text, omissions, raw output,
fallback and focus; existing file/directory checks also pass. Fixture joined.
Final verification: shared core380/0, ordinary tool checks, six default GUI groups,
six skill groups, clean GUI-tests/host/CLI builds and exact packaged search-module
and API-note bytes all pass. Browser fixture joined; no installation or inference.
Bill requested chapter-body images of the actual Artifacts; the draft and private
queue now record purposeful synthetic-data figures, alt text, captions and no
color-only meanings. Capture approval is still pending; no screenshots taken.

### Continuing Job foundation: origin identity and bounded output window

JobInfo now records the original admitted call ID and tool name, sourced from
Job's retained ToolCallPart. Wait/input keep their own model correlation IDs;
the origin is a separate relationship scoped by the owning Agent. Tool facts and
JobKilled serialization preserve it; older records leave origin unknown.

The GUI adds a distinct same-origin, custom-header-protected /jobs/window route.
It resolves Agent membership and the actual owned Job, never a browser path,
and reads at most the last 64 KiB through OpenRegularRead. Its end offset is the
Job's flushed byte-count snapshot; later writes are excluded. It returns exact
base64 bytes, offsets, origin, status, and an explicit rolled-out flag. This is
not a model observation and does not consume the model cursor. The existing
/jobs/output route still requires completion and retains its 16 MiB download
limit. Neither route authenticates bytes against arbitrary local file replacement.

This establishes transport, not terminal rendering. UTF-8/ANSI window boundaries,
page polling lifetime, one continuing Artifact, explicit foreground movement,
and historical reconstruction remain to implement. The byte window can start
inside a multibyte character or escape sequence; the renderer must handle that
rather than mistake a bounded tail for a complete terminal state.

The initial HTTP fixture omitted starting its borrowed host after ServeAsync
mounted routes. Its deadline fired and cleanup joined; no endpoint check ran.
Corrected explicit StartAsync/StopAsync produced a passing HTTP check. A later
core test compile failed nullable analysis on the synthetic termination record;
no core tests ran in that batch. Corrected before final verification. The cursor
test was strengthened to publish an unobserved marker before the HTTP read and
require that marker in the next model wait, not merely check later output.
The synthetic termination codec check then correctly refused sequence zero;
the fixture was corrected to use a positive durable sequence. This was not a
runtime origin failure. Final builds are clean with zero warnings/errors; all
three process/PTY groups, the real bounded-window HTTP test, and six default GUI
groups pass. The HTTP fixture joins its mount and stops its owned host. No browser
aggregation or ANSI rendering is claimed by this foundation. No installation,
provider inference, screenshots, or push.


### Continuing command view

The browser now owns one continuing node per Agent, Job handle, and recorded
origin call. Only run_command origins participate; old records without origin
remain ordinary observations. Call/result cards and their pairing stay immutable.
Wait and input observations foreground the existing session node; background
window reads do not move it. Replay, selection changes, and disconnect reset the
controller, abort requests, clear timers, and reject stale responses. The view
reads only the protected bounded window endpoint, never a filesystem path or the
model observation cursor. Visible running views poll at one-second intervals;
hidden panes/pages skip reads. Completed views can be refreshed explicitly.

The shared skill helper is a terminal-log decoder, not a terminal emulator.
It supports basic SGR colors/emphasis, suppresses complete terminal commands and
OSC hyperlinks, and emits literal text runs. A rolling boundary resets styles,
drops leading UTF-8 continuation bytes and defers incomplete trailing codepoints.
A missing escape-sequence prefix cannot be reconstructed: the rollover notice
explicitly permits literal control fragments at that boundary. Carriage-return
progress becomes log lines. Extended colors and cursor editing are not emulated.
Earlier output remains on disk; completed download limits remain unchanged.

Unchanged windows preserve their spans. Changed output is deferred during text
selection, including a final snapshot; selection-clear applies that bounded
pending value. The outer node and output scroller are stable. The modern DOM
moveBefore operation preserves focus during foreground movement; older browsers
fall back to append and have not been certified for focus retention.

Thirteen pure decoder checks pass. Actual-page synthetic facts and intercepted
window responses pass at desktop/phone widths: one node across run/wait,
foreground-only movement, raw observation preservation, literal hostile text,
SGR styles, focus/selection and unchanged-span retention, and held-response
rejection after reset. Existing directory/file/search browser tests and all six
default GUI groups pass on final implementation bytes; the fixture host joined.
This is separate from the preceding real HTTP/model-cursor proof. No installed
GUI, live command-to-browser integration, phone performance, or full terminal
fidelity is claimed. No screenshots or provider calls were made.

Lesson: stable outer DOM identity does not protect reading if unchanged inner
spans are replaced. Selection can also outlast the final response, so deferral
must have a selection-clear path, not depend on another polling response.


### Shared approval for the five shipped tool previews

The HTTP boundary and Speech & Display panel now select list_directory, read_file,
search_files, write_file or edit_file. The existing RendererStore already owns
independent selections by renderer name; no second registry or persistence format
was needed. Old requests default to directory; unknown or duplicate renderer
queries refuse. Package review and promotion check tool identity. Immutable module
URLs identify the whole package because relative imports do not inherit queries.

Shipped file packages preserve relative directories and include file-excerpt.js
and source-code.js. Inspection found that the excerpt helper depended on a hidden
page-global lexer. It now imports the exact packaged lexer; chat dynamically loads
the same ES module with literal-text fallback and queued repaint. This makes the
reviewed executable dependency set explicit instead of reviewing only a wrapper.
The panel disables changing tools during a trial. Acceptance and rollback remain
independent per tool; startup loads all accepted previews, while recovery mode
bypasses all custom imports. In-page code remains trusted, not contained.

Bounded verification: clean isolated build; HTTP exact embedded-package bytes,
per-tool acceptance/rollback, cross-tool refusal and invalid-query checks; visible
controls at desktop/phone widths exercise all five tools, inert stage, imported
helpers, Try, Keep/reload, safe mode and rollback. Six default GUI groups pass.
Final-byte Markdown/copy and directory/file/search browser checks pass, including
selection, focus and literal hostile data. Hosts joined. No full core rerun,
installation, inference or human acceptance claimed. A whitespace check caught an
extra trailing blank line; review restored strict-mode directive placement after
adding the module loader. General Artifact lifecycle, arbitrary MCP discovery,
command-session customization and remaining built-in/chat/default relocation are
not certified by these preview tests.

### Shipped prose relocation and configurable default

Moved the existing in-place Markdown renderer into `skills/shared/markdown.js`,
retaining script order, literal startup fallback, code-copy controls, selection
deferral, expansion and scroll behavior. This makes shipped source inspectable;
it does not yet make prose an independently approvable package. The page-loaded
pinned parser remains an explicit dependency to address in that next slice.

Extracted `tool_default` into shared skill source and the existing exact-version
approval lifecycle. Specialized preview refusal falls through the configurable
default, then the GUI literal fallback. Identity, raw details, pairing and status
remain GUI-owned. Desktop/phone tests cover unknown tools, live renderer changes,
throwing specialized/default functions and connected-node refusal. HTTP tests
cover the new one-file package and independent acceptance/rollback. No installed
or human-use acceptance is claimed.

### Exact-version prose packages

The `prose` selection now shares the existing Try/Keep/reload/rollback mechanism
for chat and supplied readable thinking. Its default export updates a supplied
body in place: `render(parent, text, allowCopy)`. It does not adopt the detached
node tool-preview API. GUI ownership of identity, original text, speech, update
scheduling and selection deferral remains unchanged. Missing or throwing code
gets a literal-text GUI fallback; that is not containment or rollback of arbitrary
in-page code effects.

The hashed package includes the Markdown entry, lexical helper and parser adapter.
The adapter encloses the byte-identical pinned markdown-it 14.1.0 UMD payload in
module-local CommonJS bindings, with its complete MIT license, rather than using
a page-global parser. No dependency download or library upgrade occurred. Reading
all executable dependencies is necessary for meaningful version approval.

Two old browser tests initially failed because their instrumentation called
classic-script globals (`markdown`, then `closedCodeFence`/`messageMarkdown`).
The first now counts calls through the prose registry slot; the second retains
all 18 fence inputs and checks actual Copy controls through the renderer. No
obsolete runtime globals were restored to satisfy tests. The page/focus/selection
and copy regressions subsequently passed, including a throwing-renderer literal
fallback with recorded text and speech controls retained. Final approval and
packaging verification is recorded after its terminal result below.

Final verification: clean GUI fixture build; HTTP per-renderer isolation and exact
three-file prose download; desktop/phone inert staging, parser/lexer import,
Try/Keep/reload, shipped-only recovery and rollback; Markdown/copy regressions;
all six default GUI groups. Fixture hosts joined. CLI and GUI distribution builds
have zero warnings/errors, and copied prose dependencies/API note match source
bytes. No full core rerun, installation, live provider or human acceptance claimed.


### Continuing command presentation approval

Extracted span presentation from `job-sessions.js` into skill source
`shared/command-output.js`; its exact decoder is part of the approved package.
The existing shared approval lifecycle now includes `command`, with no new store
or automatic execution policy. GUI retains validated Agent/Job/origin identity,
polling/cancellation, placement, rolling notice, selection and scroll ownership.
Its latest bounded snapshot supports renderer changes after Job completion,
without fetching again or changing model observations. Unchanged windows retain
nodes; selection defers a changed renderer. Missing/throwing custom code uses a
literal GUI fallback, not a promise to reverse arbitrary in-page side effects.

The first build failed because the new route block reused `files`, conflicting
with an enclosing local declaration. Renamed it `commandFiles`; no tests ran on
the failed build. Desktop/phone session checks subsequently passed including
selection-held changes, stable session node, safe fallback, raw history, and stale
response reset. Full approval and packaging results follow after verification.

Approval verification passed: clean fixture build, HTTP per-renderer isolation
and exact two-file command package, desktop/phone Try/Keep/reload/recovery/rollback,
and all six default GUI groups. Both normal distribution builds passed and copied
skill source/API bytes matched. Final review then separated the custom renderer's
mutable byte buffer from fallback input; a clean rebuilt fixture passed the
mutation-then-throw negative at both widths with joined host. Earlier broad GUI
and distribution checks precede that one-line controller correction; no full core,
installed acceptance or deployment is claimed.

## Inline recorded-dice example

Added a standalone two-file example beside the shared skill helpers, explicitly
copied into CLI and GUI distributions, not loaded by the normal application.
The synthetic chat host retains the record ID and immutable recorded facts;
`inline-dice.js` returns native details/summary styled inline. Click, Enter/Space
and touch reveal the additive terms without rerolling. Literal labels, bounded
safe-integer terms and agreement with the recorded total are checked. Invalid
facts refuse to the host's raw-record fallback; no inferred dice grammar or game
rules. Native disclosure avoids a bespoke keyboard/event lifecycle and demonstrates
that Artifact identity does not require a common card template or Tools pane.

The browser regression runs from a disposable copy with only two files served by
interception, no listener/providers or screenshots. Desktop1280 and touch390 pass
open/close, keyboard, unchanged node/host ID, 44px target, no horizontal overflow,
hostile literal labels, negative modifier, bounds/overflow/mismatch refusal and
unchanged frozen input. Randomness is replaced with a throwing function before
interaction. This is a synthetic recorded-facts fixture, not a real dice-tool or
Homebrew-VTT end-to-end test, and not a screen-reader usability certification.
Both isolated CLI and ExampleGui distribution builds pass with zero warnings or
errors; copied JS/HTML/index bytes match source. The same desktop/touch browser
checks also pass against the built GUI distribution. Nothing installed or deployed.

## Full-body authoring pass

Read the full book voice/procedure and Teach Your AI creation/voice guidance;
inspected renderer APIs, approval store/routes, command controller/endpoint and
helper sources. Personal implementation/authoring follows Bill's current direction;
no new student or independent-review campaign. Through-line: the human needs to
understand and customize output without losing the facts being supervised. Kept
Bill's recorded dice example and operation-time append-position lesson, with no
invented quotation or live demonstration. The chapter is new, not an edition
rewrite of an earlier Advanced Artifacts chapter.

Replaced aspirational TL;DR with current contracts, added causal explanations,
implementation reference, reproducible user steps, explicit evidence boundaries
and five planned figure captions/alt texts. Six exercises retained and reconciled
with terminal-log reset behavior and exact-version approval. No images captured.
README now distinguishes full draft from published or reconstruction-accepted work.
No runtime bytes changed and no build/test campaign repeated merely for freshness.
Prose hard checks pass; advisories for negative requirements and a long technical
person-gap remain intentional: explicit boundaries teach safety, and no invented
human incident was inserted to satisfy the counter. Final link/assembly checks
are recorded after their actual results.

Final authoring checks: chapter relative links and internal reference anchor,
six exercises, balanced fences, exclusion from the assembler's numbered-chapter
glob, whitespace and hard prose lint pass. No publication build performed.
Queue check initially used a case-sensitive heading that did not match the saved
entry; corrected to inspect its actual wording. The first shell did not fail fast
and its last command masked that assertion's status; subsequent checks used
fail-fast execution. Auto-recall backlog content remains intact and unimplemented.

## Authorized figure capture

Bill authorized isolated browser captures. Six actual component images now sit
beside their explanations: directory, search, replacement, command, approval and
distributed inline dice. Inputs are synthetic and captions distinguish illustration
from tool execution or test receipts. No private page, host screenshot, inference,
installation or publication. Capture source and image/fixture hashes are retained
under book/figures/advanced-artifacts/README.md. Existing compiled fixture reused;
all hosts joined and browser cleanup completed.

Pixel review caught clipped search/edit previews and the approval dialog's bounded
scroll container. Shortened synthetic search text, used the normal 690px Tools
layout, and captured approval with its source disclosure collapsed. No renderer
CSS changed for photography. Review also corrected the synthetic rollover reply
from a short tail to an actual 65536-byte window. Final command image retains its
normal scrolling edge rather than retouching it. Lesson: capture success alone
proves neither readable pixels nor faithful fixture data. Source syntax, all six image links/alt texts, PNG dimensions and hard chapter
prose/whitespace checks pass. Existing negation/person-gap style advisories remain.


## Registered MCP tool renderers

Extended the existing approval route/controller rather than add a second store.
Discovery reads root.Agents()/Agent.Tools() snapshots; no skill activation, MCP
connection, filesystem serving or execution. Deduplicated exact names map to
`tool-` plus UTF-8 SHA256, independently of the eight existing presentation keys.
A separate browser tool map prevents a tool named command/prose from using those
host renderers. Missing/inapplicable code uses default then literal fallback.
Same names share choices; differing server schemas are not asserted compatible.

Discovery occurs at page load. Absent tools are not imported; their durable choice
is retained. Mutations recheck live registration. Module reads remain verified
immutable snapshots and declared files; helper imports do not inherit queries or
require a live tool. Existing explicit trust, recovery and stale-save semantics
remain. This is not auto-loading downloaded skill code or an extension installer.

Clean isolated GuiTests build, HTTP registration/key/cross-tool refusal checks,
and desktop/phone trials passed. Fixture uses real McpClient handshake and owned
Tool registration through an in-memory fake transport on two Agents: shared name
deduplication and peer-only discovery, command-name collision, stage inertness,
Try/Keep/reload/recovery/rollback. Browser output facts remain synthetic, not a live
remote tools/call acceptance claim. Existing directory/file/search browser checks
and all six default GUI groups passed, fixture hosts joined (Job485 terminal0).
No provider/install/restart. Full core remains historical380, not rerun here.

Review caught the C# source-variable shadow before compilation and the old direct
registry lookup's collision with presentation names. Corrected both before tests.
Lesson: exact tool-name sharing needs a separate presentation-key namespace.
Both isolated normal GUI/CLI distribution builds passed with zero warnings/errors;
packaged API-note bytes matched source. Hard prose, chapter links, six exercises/
figures, draft assembler exclusion and whitespace checks passed. Style advisories
remain. Initial chapter-count checker assumed `### Exercise` rather than actual
numbered headings; it failed before notes/queue updates. Corrected the checker
and reran fail-fast, without changing chapter headings to fit the test.

## Bounded student feedback during book completion

One isolated student at baseline 9ed49875 implemented the directory slice, not
the complete chapter. Its saved feedback is in
`/private/tmp/cadence-artifacts-student.5wZD6F/assignment/student-feedback.md`.
The student reported clean offline builds, two focused native groups, and an
isolated intercepted-request browser fixture using approved installed Chrome.
No live asset-delivery or full reconstruction claim follows. The child submitted
and was shut down; experimental code remains isolated and was not imported.

Concrete teaching revisions explain constructing/migrating file-reading while
preserving enabled tools, name a directory-only checkpoint, separate declaration
validation from execution, and explain classic app.js versus module loading.
The missing browser and seq=0 test record were fixture/setup failures, not runtime
defects. Browser setup guidance now forbids assuming a bundled executable.
One editorial reviewer follows these revisions; no student repair campaign.

## One editorial review and author revision

Reviewer feedback saved at `/private/tmp/cadence-artifacts-reviewer-feedback.md`.
One read-only chapter/API pass found three small teaching gaps: state the default
export convention, explain a shipped-renderer/fixture demonstration before the
trial panel exists, and use Removed/Inserted terminology rather than imply a
whole-file diff. All three accepted and corrected. Reviewer submitted and was
shut down. No additional review/student campaign. The opening now distinguishes
the completed directory-only attempt from full reconstruction, and figures from
installed usability. Final artifacts follow this manuscript snapshot.

## Final local book candidate

Thirty chapters assembled, Advanced Artifacts28 before Crossing Over29. Final
EPUB has36 navigation entries and seven embedded images (six figures plus cover).
Final print has559 interior pages at7x10,560 covered reading pages,33 verified
contents references and141 preserved reading links. Actual body10.5pt and minimum
code10pt passed; fourteen publication regressions passed. Six PDF figure objects
verified; revised file/search and approval pages visually inspected. Source
Crossing Over body preserved byte-for-byte except heading number.

The first final print attempt refused shrunken type (10.11/9.62pt): suppressing
duplicate figure captions had made full-width images inline in indented paragraphs.
Browser measurement identified21px image overflow; block display fixed the cause.
The regression now checks block image markup, descriptive alt and no duplicate
figcaption. An image-page count assertion was corrected to count image objects:
two figures legitimately share a page. Failed logs retained in the build workspace.

Verified artifacts copied locally into book/; byte/hash receipt at
`/private/tmp/cadence-book-completion.yr6v6W/final-artifacts.json`. No publication,
push, install or app restart. Print cover wrap must use559 pages and chosen paper;
KDP preview and Bill's publication approval remain separate.

Final chapter prose hard checks passed; advisory negation/person-gap counts remain.
An initial lint invocation omitted its required filename and ran no checks; the
corrected explicit chapter invocation passed. Crossover body, six exercises and
local links verified. Four local release copies match their candidate hashes.


## Human Dev feedback: compact result presentation (2026-11-07)

Bill successfully opened the isolated updated Dev installation and requested a
single Job header, icon controls beside speech, scrolling latest-64-KiB output,
and one local-time/duration footer. Implemented shared-card changes and removed
the redundant directory fallback label. Typed Job state remains authoritative;
running observations retain their historical qualification. Full date and timing
meaning remain accessible. Download and explicit call pairing keep their handlers.

Ordinary result windows use the existing owned endpoint once, never advance model
cursors, validate identity/range/encoding, discard a split UTF-8 prefix, cancel on
replay/disconnect/switch, and reject stale card responses. Complete structured
facts render without a fetch. Partial structured facts retain a labeled text tail.
Original observation stays immutable and supplies speech/audit. Running commands
retain the separate continuing view, not a new polling loop per observation.

Actual-page desktop/phone checks cover 64-KiB tail versus smaller model text,
compact controls, same-line timing, stable focus/pairing, partial evidence and
late-response rejection. Existing pairing/state/directory/file/search groups and
six default GUI groups pass; clean fixture/distribution builds. Test corrections:
await painted DOM rather than assigned fetch state; target visible preview versus
hidden audit payload explicitly; replace obsolete short-preview marker and summary
focus expectations with scrolling-body checks. Every fixture joined. Evidence:
/private/tmp/cadence-renderer-versions.gOOJ6z/compact-*.log. No live provider test.
Dev still needs the new build and Bill's second look. Main untouched; book release
artifacts/figures intentionally await the end of this human feedback phase.


## Numbered reads at the construction point (2026-11-07)

Bill accepted the directory presentation after the compact controls and explicit
`list_directory result` title were installed in Dev. During the next file-read
pass he identified missing model-facing line numbers. GUI source evidence alone
is not a substitute for that tool contract. Chapter3, not Chapter4, constructs
`read_file`; the requirement belongs there.

Model text now prefixes each captured source line with its actual one-based
number and `: `. Raw operation-time evidence remains unnumbered for the source
renderer. Preserve indentation, CRLF, blank lines and unterminated final content;
empty selections do not invent lines. Source UTF-8 clipping precedes prefixes,
and the truncation marker remains unnumbered. Exact edits still use unique text
anchors without presentation prefixes, not line addresses.

Inspection also found that the schema and Chapter3 promised a 32768-byte default
but the implementation used an unlimited source cap. Omitted and zero now honor
the documented default; explicit larger reads remain supported. The independent
Job observation cap still applies to the numbered report. This is a tool-output
correction, not a new syntax-highlighting implementation or deployment.

Verification: focused actual-tool/evidence checks passed; the outside-assembly
consumer passed all10 checks. Its independently built run against the older Dev
assembly failed the intended actual-range assertion, not setup. First full core
run was379/1 because the confined-file test expected unnumbered text; changed
only that expected line, preserving escape/refusal checks. Final full core380/0,
actual-page compact/pairing/state/directory/file/search tests at desktop/phone
widths, six GUI groups, and clean installable GUI build passed. Fixtures joined.
Evidence is retained under `/private/tmp/cadence-renderer-versions.gOOJ6z/numbered-read-*`.
Chapter3 hard prose lint still fails three pre-existing em dashes, byte-identical
to HEAD; two advisories remain. The new diff passes whitespace checks. No frozen
endpoint, historical grader score, release artifact or running installation was
changed. Source highlighting and file-view usability are the next discussion.

### Real-use correction: retained evidence must survive display reconstruction

Bill's Dev read of `cmd/main.go` displayed plain text without a line gutter or
highlighted strings. Local Dev HTTP served all source-renderer modules as
JavaScript, but that alone did not prove browser execution. Inspection found a
concrete data-loss boundary: `Gui.DisplayFacts` reconstructed tool-return events
from retained Context entries without directory, file, or search evidence.
Downstream serialization therefore correctly emitted null evidence. Preserve
all three existing immutable fields when constructing the display event.

A new outside-assembly regression dispatches real read, directory, and search
tools against disposable files, then places their actual results in retained
conversation fixtures. It checks complete evidence equality through ordinary
and snapshot-only GUI projection. Before the repair it built successfully and
failed specifically at missing `file_evidence`; after the repair all three
checks and the default GUI suite passed. Existing synthetic-frame browser
regressions passed at desktop/phone widths, their host joined, and the GUI
distribution built with zero warnings/errors. Logs: `evidence-projection-*` in
the existing renderer verification workspace. These are complementary checks,
not a claim of live-provider or installed-Dev acceptance. No new core-suite run
was needed for this GUI-only field-copy repair.

Lesson: real tool evidence tests plus synthetic browser rendering tests left
the reconstruction boundary uncovered. Readable fallback concealed the loss;
sharing a highlighter did not prove that the source renderer received its data.
The candidate also contains the earlier numbered-read correction. Installation
and Bill's next usability check remain separate gates. No highlighting library,
publication artifact, running installation, or private browser data changed.
