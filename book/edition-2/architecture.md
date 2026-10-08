# Second-edition architecture decisions

Recorded from Bill's clarification on 2026-10-07. This is the design target
for the rewrite, not a description of what the first-edition code already does.

## Purpose of the edition

Absorb lessons into the chapters where the relevant structures first appear.
Do not build the old architecture and append its repair later. Choose data
structures and ownership early enough that subsequent chapters extend them.
The first-edition chronology is evidence, not a required construction order.

Bill explicitly requires Chapter 1 to explain the methodology in plain English
before any code. The star dependency pattern, directory responsibilities, core
structures and interfaces in `internal/common`, behavior in its owning spoke,
and interface back-pointers apply to the first implementation. Free functions
are the deliberate answer to Go's same-package method restriction. The coder
must read `skills/ensemble-coding/SKILL.md` before every coding task. Neither
a small Chapter 1 exercise nor an older grader excuses a flat implementation.

Starting in Chapter 2, establish clean core data structures and two clients
for Hub/Ensemble: CLI and browser GUI over WebSocket. The GUI may be stubbed
in Chapter 2; the client separation must already exist. The GUI transport
still belongs in the separate optional GUI module, not in the agent library.

## Settled boundaries

- One application-level owner manages potentially many Agents, their
  communication, and communication with the GUI or an external gateway.
  Bill calls this Hub/Ensemble; the final name remains open. Use **Ensemble**
  as the working name to distinguish it from the old GUI `ws.Hub`.
- Agent holds an interface back-pointer to its owning Hub/Ensemble, as child
  objects hold back-pointers to their owners. Observer is how Hub/Ensemble
  receives streaming events and anything that should appear on the GUI in
  real time. It is not the exclusive Agent-to-Hub interface. Requests such as
  creating sub-agents or sending messages can use explicit parent-interface
  methods. Distinguish published observations from requests and shared-service
  access; none requires exposing concrete coordinator internals. The exact
  method set is still to be designed.
- GUI code lives in a **separate Go module**, optionally imported by a user's
  application. Moving it to another package within the agent module does not
  satisfy this boundary.
- The WebSocket implementation is a detail of the GUI server. GUI code does
  not belong under `agent/internal`.
- A user may build Ensemble with its GUI, build an application without that
  GUI, or reuse GUI components such as WebSocket, Artifact, and Connector in
  their own interface. Preserve those composition choices.
- Code likely to need debug logging must have access to the logger through
  the ownership chain. Statelessness alone is not an exemption.
- Tool visibility is per-agent. Registry storage may be per-agent or shared
  on the application owner; Bill explicitly allows either design.

The application root and the shared vocabulary package are different things.
The term "hub" in the star-import rule means a package of shared types and
interfaces. It must not be confused with the runtime Hub/Ensemble that owns
agents and services.

During the manual data-structure review on 2026-10-07, Bill explicitly allowed
private runtime structs in their implementation packages when common
interfaces expose the ownership chain. Engine, Registry, Jobs and Job may
have private concrete implementations. Shared core values and interfaces stay
in `internal/common`; behavior over common data stays in the responsible
spoke, using free functions where necessary. This settles the placement
boundary without requiring all runtime implementations to move into common.

## Consequences for the module contract

The agent library must work without importing the optional GUI module. The
GUI consumes public, transport-independent agent interfaces rather than
reaching into `agent/internal`. The user's application selects the modules
and connects the surfaces. Core agent behavior must not require WebSocket
types or a running GUI server.

When the chapter introduces module boundaries, prove them with both a headless
consumer and a GUI consumer, and exercise reusable GUI components through
their public API. Check imports and module dependencies as well as behavior;
a GUI moved to a sibling directory but still required by the agent is not
optional. This is a proposed verification method for the settled boundary.

Distinguish GUI components from transport-independent data they display.
The mention of an Artifact GUI component does not by itself settle the home
of every artifact data type. Specify that public seam before implementation.

## MCP transport and the later GUI tunnel

Bill explicitly requires transport-independent MCP on October 8, 2026. A later
chapter must tunnel MCP over WebSocket so the Agent can inspect and control its
GUI. This is a required capability, not an optional optimization. The MCP chapter
must teach the transport seam before its first implementation; stdio cannot be
hard-coded into protocol parsing, discovery, request correlation or cancellation.

Declare the shared transport and parent interfaces in common and expose the
public construction seam needed by an embedding application. Keep message
transport separate from stdio line framing and WebSocket frames. Concrete GUI
tunneling belongs in the optional GUI module. Specify actual adapter ownership,
logger access, close/unblock behavior and the distinction between a logical MCP
channel and a GUI-owned physical socket. Closing one channel must not dispose
unrelated GUI work. The chapter's reviewed owner plan will settle concrete APIs.

The MCP chapter must exercise an alternative message transport and a public
custom-transport consumer. The later GUI chapter implements and demonstrates
the real WebSocket tunnel and GUI observation/control. These checks establish
the seam early while preserving the later chapter's actual feature work.

## Decisions still open

- Final name: Ensemble is the coordinator's recommendation, not yet Bill's
  final naming decision.
- Logger implementation decision: coordinator selects the application
  Ensemble as owner, following Bill's preference. Keep headless operation
  possible. This is an explicit working design choice, not a claimed new
  user ruling; interface spelling remains the student's design.
- Registry ownership: Chapter 3 selects Agent-owned storage as the current
  implementation choice, with the same visible set governing declarations and
  dispatch. Bill also permits shared storage on Ensemble; the chapter's choice
  does not turn either arrangement into a universal architecture rule.
- Exact public interfaces and module paths for GUI component reuse.
- Exact request and shared-service methods on the parent interface. Bill has
  resolved that Observer is not exclusive; logging need not be forced into
  the GUI event stream. Logger and registry ownership remain separate choices.

Agent configuration belongs on Agent, per Bill's guidance. The
existing methodology keeps usage on Engine, which knows the producing model.
Reachability does not justify moving a fact to a more convenient ancestor.

Chapter 4's reviewed contract selects Agent-owned Jobs, Jobs-owned live Job,
and an Ensemble-owned application-wide handle allocator. These are explicit
working design choices. Unique handles do not authorize one Agent to supervise
another Agent's jobs. Background completion uses Agent's serialized durable
event path and Observer, including while a model request is in flight.

Chapter 5's reviewed contract selects one Agent-owned Actor, implemented in
`internal/llm` over common declarations, to serialize turn decisions and durable
conversation changes. Engine remains Agent-owned and keeps transport/usage;
Jobs retains job state, output and transactional report cursors. Actor reaches
these owners through Agent rather than injected sibling services. Model and
report workers return owned facts without parking the actor. Ensemble owns
reliable completion collections; transient display observations do not deliver
or consume request completion. The coordinator accepted these working choices
on 2026-10-07; implementation still requires the validated Chapter 4 predecessor.

Chapter 7's comparative review exposed another shared resource: native browser
speech belongs to the document, even when multiple Agent panels have separate
logical queues. The coordinator's accepted revision plan gives an explicit
browser application root ownership of the native speech service and its Page
children. Pages keep their own input, queues and pause causes; they reach the
shared service through that parent. The service admits ready utterances in
FIFO order, owns at most one native utterance, and cancels only the requesting
Page's work. An idle Page cannot cancel another Page's speech. Page/component
close removes owned DOM listeners and fences pending callbacks before a
replacement can reuse the layout. This is optional GUI ownership, with no new
browser dependency in the agent library. See Chapter 7 and its validation
record for implementation/review status; this paragraph records the design
decision, not a completed revision claim.

Chapter 8's actual native-speech test found that application-local ownership
alone did not coordinate ordinary tabs: a waiting tab's playback timeout could
cancel another tab's utterance. The accepted correction keeps SpeechService
under BrowserApplication and adds a same-origin/storage-bucket Web Locks lease
before native speech and its start timer. A canceled waiter aborts only its
wait; a leaseholder cleans up native work before releasing ownership. Page
queues, captured rates and pause causes remain local. This coordinates
cooperating documents, not unrelated sites or profiles. Missing coordination
is a visible unavailable-speech state with no retained speaking pause. See the
Chapter 8 validation and comparison records for source-bound checks and live
limitations. No browser facility becomes a core Agent dependency.

## Source reconciliation

`docs/ensemble-topology-decisions.md` already distinguishes the application
Ensemble from GUIServer and external Gateway. Its dated implementation counts
are historical. Bill's current instruction adds the explicit separate-module
requirement and requires teaching the corrected organization from the start.

The current tree still imports `agent/internal/ws` from `agent/agent.go`, and
`agent/go.mod` requires the WebSocket library. This is evidence of work the
second edition must eliminate, not a precedent the new solution must follow.

Do not start implementation while an unresolved ownership decision affects
that implementation. The coding methodology above is settled and mandatory;
remaining owner choices do not make it optional. Chapter 1's complete exercise
contract and student snapshot are validated; see `chapter-01-validation.md`.
Apply each later chapter's published contract before extending that snapshot.

Every "Taking it for a spin" section requires actual coder-run user-facing
checks with a real model, initially via CLI, covering all chapter features.
Fake grading does not substitute for live usability. See the coding skill for
the evidence checklist and safe use of Bill's authorized local credentials.
