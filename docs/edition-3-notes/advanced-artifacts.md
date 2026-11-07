# Advanced Artifacts: working chapter notes

Status: initial [chapter packet](../../book/advanced-artifacts.md) drafted with
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
