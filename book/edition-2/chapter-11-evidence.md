# Chapter 11 preparation evidence

Author-only research, October 8, 2026. New Chapter 11 corresponds to old
Chapter 12, MCP. This author previously implemented new Chapter 4 and authored
new Chapters 7–10, with no Chapter 11 implementation authorship. Historical
chapters, old grader source and review notes below are author evidence and
outside the future cold student's reading set. No old solution was copied and
no new runtime/grader was written or run in this preparation.

## Authority and map

Read the complete current voice and chapter-writing procedure in the immediately
preceding author phase and retained them for this task; reloaded current
architecture, working procedure and the relevant global review. The workflow
maps old Chapters 6–21 to new 5–20. Actual old `book/chapter-12.md` is “MCP: The
Extension Protocol”; the global-review table's “12: voice” label is stale. Old
Chapter 13 is “The Agent Sees Itself” and remains the next concrete GUI consumer.
No new chapter-count claim or map adjustment follows from that stale label.

The coordinator accepted current MCP 2026-07-28 as the versioned subset direction
for the outline, after independently checking the official versioning page.
This replaces neither historical artifacts nor a claim of legacy compatibility.

Bill then explicitly required MCP over WebSocket for future GUI observation and
control: “The MCP transport should not be hard-coded.” The coordinator records
that instruction at `ad2b2e8`. The outline makes a transport-independent complete-
message boundary and an independent public alternative adapter **required now**.
Actual WebSocket framing/multiplexing belongs in the optional GUI module in
Chapter 12. The core cannot grow a GUI dependency to satisfy it.

The Ensemble-owned connection service/per-Agent binding direction is proposed,
with the coordinator requesting exact creator, parent and close paths. The
outline's open decisions remain real review questions, not silently accepted
student requirements. Chapter 10's contract was accepted at `ae21bf4`, with the
subsequent explicit absent-exponent clarification at `b4e3fb7` noted by the
coordinator. Its runtime and checkpoint remain future prerequisites.

Bill/CodeRhapsody's new first-edition sandboxing files were not opened, edited,
or treated as a completed lesson source. They remain pending future evidence.

## Actual local source reads

- Complete old `book/chapter-12.md`: protocol/transport sketch, ad hoc ephemeral
  fields, reverse calls, browser forward sketch and old seven-check exercise.
- Old Chapter 13 opening/TL;DR and lifecycle sections through its exercise:
  the actual skill-to-MCP-to-browser integration and its callback-based wiring.
  This pass did not reread that whole chapter's later user-study narrative.
- Old Chapter 21 MCP integration and security passages: parsed but unconnected
  primary configuration; actual server surface versus per-skill allowlist;
  returned content cannot grant rights. Historical prices/tool counts are not
  adopted as current figures.
- Old Chapter 22 owner-reachability diagnosis and constructor discussion. The
  chapter's old example shortcuts do not override current architecture.
- New Chapter 2 request ephemera/capture and rendering rules; new Chapter 5
  owned workers, serial admission, pause/interruption and durable request facts;
  new Chapters 9–10 configuration, frozen ceiling, strict skill format, transition
  acceptance, session ownership/identity and replay rules. Their exact shared
  authorities constrain MCP integration rather than being replaced by it.
- Complete `internal/grade/ch12_checks.go` and `ch12_harness.go`, and complete
  `scripts/ch12_audit.sh`. The latter mutates `agent/` and was **not executed**.
- Git history for old Chapter 12 and its graders, and selected exact diffs below.
  No current or frozen first-edition MCP implementation body was read for this
  preparation; directory names and history statistics were inspected.

Prior source teaching makes two apparent old shortcuts unsuitable immediately:
Chapter 9 forbids new handler grants outside the installed ceiling and permits
only its published restricted frontmatter keys; Chapter 10 fixes handler
identity for live resume. A runtime `mcp_servers` parser plus a callback that
registers arbitrary discovered tools would violate both. The outline therefore
proposes explicit external preparation and alias installation, leaving a reviewed
next-chapter connection-policy extension rather than silently accepting old YAML.

## Git chronology checked

| Revision | Evidence and limit |
|---|---|
| `8c9d6c8cbb6802ebabc64499b05a36fd622fe2d5` | Initial old chapter/brief/review. Read its full historical `book/review-ch12.md` with git show because that file is absent at current HEAD. Its “nothing must-fix” verdict is historical, not this author's conclusion. |
| `c1b113901680cbd277f7111fbfd2e7719bdc1428` | Grader introduction and Done/EOF history; commit message/statistics read. Claimed old scores remain claims of that run, not freshly measured results. |
| `d83575f4f41a541b44b8e665ef038a3e067182c7` | First old Chapter 12 solution snapshot/audit identity located through history; implementation not copied or read in this pass. |
| `3d7b7d19ea14c774b0f2480868b3d43af1ac523e` | Exact chapter diff adds honest GUI-snapshot elision and visible control-state obligations after virtual-user findings. Carry those into the concrete Chapter 12 consumer. |
| `c24b8e84da112ef66f729c0cdc60413ee6fb547d` | Exact old harness diff gives MCP runs their own cwd after default persistence caused shared save files. Preserve isolated workspace fixtures from the outset. |
| `297cb7e9c6f1256c6031e28e1e71a5fd6df22953` | Later old ch12–14 snapshot history carries anchored persistence forward; located as comparative-review candidate, not opened as an answer. |

The historical review notices that tool-call and ephemeral flags are assigned
true, then asserts other checks independently protect those behaviors. Current
`ch12_checks.go` only reads those flags. That source comparison does not support
its assurance. The old audit script's existence likewise proves no passing
mutation or intended failure set, particularly where textual mutations might
not compile. No mutation was executed during this author research.

## Current protocol research

Official sources were checked on October 8, 2026. The unversioned versioning URL
redirects to the 2026-07-28 documentation and explicitly names that revision
current. A search result describing its release candidate is earlier evidence;
use the current versioning/specification pages rather than treating the older
blog title as current status.

| Primary source | Specific finding used |
|---|---|
| [Versioning](https://modelcontextprotocol.io/docs/2026-07-28/learn/versioning) | Current revision 2026-07-28; dates identify incompatible revisions. |
| [Versioning and compatibility](https://modelcontextprotocol.io/specification/2026-07-28/basic/versioning) | Requests declare version; initialize-based 2025-11-25 and earlier are a separate legacy family. A modern-only client may fail old servers visibly instead of implementing silent fallback. |
| [Base protocol](https://modelcontextprotocol.io/specification/2026-07-28/basic) | Required per-request metadata; exact string/integer IDs; resultType distinction; self-reported server identity is not verified identity. Schemas default to 2020-12 and external references must not be fetched automatically. |
| [Discovery](https://modelcontextprotocol.io/specification/2026-07-28/server/discover) | server/discover returns versions/capabilities/self-reported identity. Client use is optional in the protocol; making it the chapter's initial compatibility probe is an application choice. |
| [Message patterns](https://modelcontextprotocol.io/specification/2026-07-28/basic/patterns) | Server-initiated JSON-RPC requests are forbidden. Input-required and subscriptions are separate explicit patterns, not a general reverse-tool right. |
| [Transport overview](https://modelcontextprotocol.io/specification/2026-07-28/basic/transports) | Framing and delivery are bindings; custom transports preserve protocol semantics and request metadata. WebSocket tunnel is custom, not a standard transport named by this version. |
| [Stdio binding](https://modelcontextprotocol.io/specification/2026-07-28/basic/transports/stdio) | Newline-delimited UTF-8 messages; stdout is protocol-only, stderr is separate. Close input, wait, then terminate as needed. A connection/process is not a conversation or skill activation. |
| [Cancellation](https://modelcontextprotocol.io/specification/2026-07-28/basic/patterns/cancellation) | Stdio cancellation names an outstanding request via notifications/cancelled. HTTP uses request-stream closure. Races and uncancelable completed effects prevent a rollback guarantee. |
| [Tools](https://modelcontextprotocol.io/specification/2026-07-28/server/tools) | Paginated discovery and remote-name scope; complete versus input-required results; protocol errors differ from isError. Text and structured values are distinct from other content kinds. |
| [Legacy lifecycle](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle) and [legacy cancellation](https://modelcontextprotocol.io/specification/2025-11-25/basic/utilities/cancellation) | Contrast only: initialize/initialized belongs to that version; even that version uses notifications/cancelled rather than old chapter's $/cancelRequest. |

Some guessed newer URLs for message-patterns/utilities failed; following the
current base page's links resolved the actual `/basic/patterns` locations above.
No unavailable page or generic search snippet supplies the contract. No remote
service was contacted as an application client, and no paid request was made.

The protocol's suggestion to restart an unexpectedly exited stdio process does
not establish that retrying an external effect is safe. The proposed course
contract retains explicit reopen and no automatic effect retry, with a truthful
unknown external outcome after transport loss. Discovery/setup retries and an
already sent tool call are different decisions; the full draft must name them.

## Legacy grader gaps, from inspected source

Seven old checks allocate 100 points, but source evidence supports a narrower
claim than their names:

- Handshake waits for initialize and an ID, then answers with 2024-11-05. It
  treats initialized notification as optional and does not validate the new
  version/metadata semantics.
- Discovery observes tools/list, then returns a canned pair. It does not prove
  that the discovered tools entered the right Agent's authorized set.
- ToolCallOK and EphemeralOK become true immediately after that exchange. Their
  15+20 points require neither a remote tool call/result nor a model request
  with refreshed ephemeral material.
- Reverse-call sends nonexistent_tool and accepts a response with either result
  or error. It proves neither authorized execution nor a denied real effect.
- WebSocket routing checks source strings `jsonrpc` and `WSTransport` at old core
  GUI paths. It neither opens a browser socket nor proves an interchangeable
  transport; those paths conflict with the new optional-module boundary.
- Parity depends on the surrounding runner. This preparation did not run it and
  does not claim a fresh passing old baseline.

Keep the historical diagnostic identity CH=12. New checks must protect actual
protocol behavior, external effects, alternative transports, ownership and
shutdown. Any shared grader enhancement requires its own old positive baseline
and retained mutation controls; an author outline does not authorize weakening
coverage or editing first-edition sources.

## Open design questions sent to the coordinator

The outline proposes Ensemble-owned service/connections and Agent-scoped frozen
aliases, with complete-message transport children receiving their actual
Connection parent. It separates generic protocol logic from stdio process/framing
and the next chapter's GUI physical socket. Public custom-transport embedding and
same-suite in-memory behavior are required, reflecting Bill's instruction.

Four questions remain before full draft: exact shared-connection lifetime and
constructor contract; external preparation versus skill-triggered lifecycle and
automatic context scope; persisted logical remote-binding identity/versioning
under Chapter 10; and the bounded schema/result subset with explicit unsupported
content and size/time errors. These are real behavior/ownership choices. The
outline records proposals instead of inventing accepted answers.

No literal run output, screenshot, provider model selection, tool result,
latency/cost claim or successful grader score has been created for Chapter 11.
The eventual all-provider human CLI/browser/public plan appears as a plan.

## Local source fingerprints at preparation checkpoint

```text
17883353cf9653c0df46081c216b5b1623dfa640619e06fe89bff46fc29dfad1 book/voice.md
1131abf07ea86d246421122b20ee368cce30fb2cd684a6ca3ab279c220ec7d3c book/chapter-writing-procedure.md
9dada72e65c36fb649659f9076b4ca2bdfc8379fef919a3b575652dca0e49327 book/edition-2/architecture.md
b44ee30f1cb5d43a971b06473468e3a175b0890600d54e0a902a7e59d8903384 book/edition-2/workflow.md
15984b674642879880282695a926dee7636c8396793e6ad69d2a43c24fda9b9e book/edition-2/global-review.md
a754bd50ca7f3c7a2da868ff5acbdb98b4d5eddaa05349e7890c4a3dbf6cd4b9 book/chapter-12.md
9e3e74c172fb479376133a5a11b4e44cd3eb36e043502d74972fae95abeeb630 internal/grade/ch12_checks.go
2cfb3f2278514022ec704a3c0c61c82e70e47c9929ec9e1eb5501c25d052c560 internal/grade/ch12_harness.go
b0392a9f47f508549d68c28bc3d970ac06fc3f6282a53a9a57cf2c4ea5c23536 scripts/ch12_audit.sh
bc0d4a6f162e9dde53fb351c85512050ca0c2cf25d7500125e805b6eefea0b33 book/edition-2/chapter-09.md
8a4deaffecf6ff8b91e2774755a5c39a3ec220e27b9ad5c0bd707951a1a1de07 book/edition-2/chapter-10.md
```

Author checks: scoped prose lint passes every hard rule for the outline and
research record. Short length and negation-density warnings are appropriate to
preparation documents with explicit scope/refusal candidates; no full chapter
length is claimed. The owner table and proposed contract were read for duplicate
authorities, blocking Actor waits and transport-specific assumptions. The cut
pass retained each proposed bound and each real unresolved decision; no runtime
acceptance result is inferred from these prose checks. `git diff --check` is clean.

## Full contract draft, October 8, 2026

Reloaded the complete current voice and writing procedure after compaction,
plus architecture. A combined read was truncated; the remaining voice and both
procedure ranges were reread separately before prose. Rechecked Chapter 10's
exact outer/initializer/anchor/state/identity and numeric contracts, Chapter 4's
job statuses/errors, and Chapter 7's safe watch/observation boundary. No old
implementation body, runtime edit, grader change or paid call was involved.

The coordinator accepted all four outline directions before drafting, and
accepted the concrete cancellation-delivery and v1/v2 compatibility direction
during the draft. These are coordinator working choices. Bill's explicit
transport-independent MCP requirement remains separately attributed to ad2b2e8.
The initial proposal paragraphs and fingerprints above retain their earlier
pending status. Root corrected the old12/old13 map labels at 0cb54a4 without
changing the chapter numbering.

The draft now specifies:

- A public actual-parent constructor and shared service lifetime, with stdio,
  memory and external public custom-transport proofs. Chapter 12 still owns the
  required real GUI WebSocket tunnel and automatic-context/lifecycle consumer.
- An exact current-version discovery/list/call/cancel sequence. Connection owns
  correlation and constructs cancellation; adapters only deliver the supplied
  message or abandon their own request stream. Canonical rpc-N IDs are allocated
  at attempted issue, with no unsent reserved holes; a watermark and pending map
  distinguish stale replies from unknown IDs without accumulating tombstones.
- A published bounded schema vocabulary, local-reference resolution and
  validation work limits, with explicit refusal outside the profile. Current
  [JSON Schema validation](https://json-schema.org/draft/2020-12/json-schema-validation)
  and [core](https://json-schema.org/draft/2020-12/json-schema-core) sources were
  checked for keyword semantics. This is deliberately not full schema compliance.
- A whole-result text/structured-JSON subset, canonical artifact bytes with LF,
  truthful isError and safe local failure records. Rechecked official current
  MCP tools, cancellation and stdio pages for the relevant protocol distinctions.
- Strict checkpoint/state version 2 and explicitly versioned initializer/anchor
  payloads for nonempty MCP bindings. Zero-binding stores remain exact version 1.
  Both phases of live resume and the zero-effect historical path are explicit.
- CLI/launcher configuration, an Agent-safe mcp watch array, transient actor-
  published mcp_changed observations, exact generation handling and an ordinary
  browser display. No credentials, paths or stderr enter that projection.

Root read the on-disk full draft and identified three useful bounded-profile
precisions before freeze. The final draft names each schema's canonical bytes
and sums normalized retained descriptor-object bytes without array punctuation;
ignored metadata still counts against whole-wire/node limits. The 64 operation
permits now remain owned through staged cancellation delivery even after pending
correlation is removed, with one writer and a one-second staging-through-delivery
bound. This prevents canceled calls from creating an unbounded notification queue.
Remote JSON-RPC error codes use an explicit losslessly checked int32 domain,
avoiding huge decimal expansion in a supposedly safe diagnostic.

The human stake remains attaching a useful external tool without rebuilding the
Agent, then preserving control when it stalls or disappears. The planned utility
has actual scratch-file behavior, and the planned browser/public demonstrations
have their own required receipts. Offline inspector restart with endpoints
disabled is distinguished from live compatible resume, which may discover but
must never repeat a historical tools/call. No planned success is presented as
an observed outcome.

A short priority interruption resolved the fresh Chapter 9 student's three
pre-code teaching questions at 8f24360, including a narrow Chapter 10 inherited
state clarification. That was prose work only and does not supply a Chapter 11
implementation. The Chapter 11 contract remains subject to coordinator and
independent checker review; predecessor acceptance and a published new checker
command are still required before student handoff.

Final author checks for this draft: scoped lint passes every hard rule. The
chapter is 5,963 counted prose words before any later review edits; the TL;DR is
within the 700-word target. Ten JSON fixtures and two canonical artifact/error
fixtures parse. `git diff --check` is clean. The full chapter and paragraph
closers were read for repeated conclusions and cross-section contradictions.
The soft person-gap warning was examined against the actual reader/task passages
in §§11.4, 11.6 and 11.7; no personal incident was invented to satisfy a detector.
The remaining negation warning reflects explicit compatibility/refusal boundaries.
Short outline/research lengths are not presented as full chapters.

The voice, procedure and architecture fingerprints still match the earlier
ledger. The relevant new contracts after the priority clarification are:

```text
fb86fa4890cb516d0ce4c51a536d0fcb5f09a5362f937945cc6d700643805209 book/edition-2/chapter-09.md
31272ff9621f932d92f252be06b67dfec04ee8826ac3952ecbc14fdd18696982 book/edition-2/chapter-10.md
```

No build, race, legacy grader or live acceptance result is inferred from these
prose/fixture checks. The independent full-contract review remains the next gate.
