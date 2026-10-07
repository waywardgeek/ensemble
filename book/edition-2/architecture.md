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
