# Chapter 11 contract review

October 8, 2026. The coordinator accepts the draft at `a5fc9ea`, with wording
corrections `302d77e` and `39bc526`, for independent checker preparation.
An accepted Chapter 10 implementation and a published independent checker remain
prerequisites to a fresh student handoff. This is neither runtime acceptance nor
Bill's editorial approval.

## Sources and voice

The coordinator read the complete new outline/evidence and draft, then the final
revised transport, bound, cancellation and watch passages. Read the entire old
Chapter 12, current voice and chapter procedure, architecture, and the relevant
Chapter 9/10 contracts. Current primary MCP base, discovery, transport, tools and
cancellation documentation was checked; the author's evidence record carries
the exact links and distinguishes current protocol from historical examples.
No first-edition MCP implementation was copied or used to prescribe new code.

The old chapter's concrete motivation survives: adding an external tool should
not require another Agent rebuild. The new opening follows its consequence
through shared Agents, cancellation and a lost receipt after an external effect.
That gives the ownership rules a reason. The historical parsed-but-unconnected
configuration motivates explicit preparation; the chapter does not invent a
successful demonstration. Dense tables and literal fixtures belong to the
exercise contract, with the reader's useful notebook tool returning in the
planned spin. Final prose reconciliation still needs the actual student's runs.

## Accepted boundaries

- Bill's transport requirement is explicit from the first MCP implementation:
  complete messages, stdio and memory adapters, and an external public custom
  adapter. Actual WebSocket tunneling and GUI observation/control remain required
  next-chapter work in the optional GUI module.
- Ensemble owns the service, which owns Connections and their child transports.
  Agent bindings and grants stay separate. Parent interfaces provide diagnostics;
  one Agent's cancellation does not close a healthy connection shared by another.
- Discovery and frozen aliases precede Agent publication. Peer descriptions,
  reverse requests and results cannot widen installed or active authority.
- The stated protocol/schema profile is deliberately bounded. Local validation,
  exact correlation, whole-result acceptance and truthful effect uncertainty are
  independently observable requirements, not inferred from discovery success.
- Version-1 plain sessions remain unchanged; MCP bindings use an explicit
  version-2 identity. Offline reconstruction creates no connection. Live resume
  validates stored state before new preparation and never retries old effects.
- Safe binding presentation goes through the Agent's actor/watch boundary.
  Logical MCP endpoint lifetime remains distinct from a future GUI socket owner.

## Review findings and resolutions

The author resolved the grouped findings before freezing the draft: transport
abandonment receives protocol bytes from Connection; issued IDs use bounded
watermark/pending bookkeeping; schema and descriptor byte counts name their exact
canonical representation; cancellation delivery retains its shared operation
permit until settled; error codes are checked losslessly before bounded printing.
These are published requirements, not private implementation instructions.

The final wording pass distinguishes wire whitespace from JSON value nodes and
defines cycle detection across schema containment plus reference expansion.
The recursive-child example is valid JSON. These clarifications resolve the
reported ambiguities without adding another feature.

No local runtime, deletion, legacy regression or paid provider result exists for
this chapter yet. The historical grader remains a separately identified diagnostic;
its discovered-tool flags and source-string checks cannot prove the new contract.
Future checks must establish actual calls/effects, alternative transport behavior,
bounded cancellation and shutdown, authority, persistence and public usability.

The next outline must settle browser bootstrap explicitly: a frozen Agent binding
cannot depend on covert late discovery merely because a GUI usually creates its
Agent before opening a browser connection. The author has that question for
Chapter 12; no unpublished solution is imposed on a student here.

## Narrow handoff-gate clarification review: 5d6255f

October 8 PDT / October 9 UTC 2026. The independent reviewer who reviewed Chapters
17–18 accepts the two-file clarification at
`5d6255feca0e8201e3d0f2102df3d13ecb6074aa`. This is a narrow review of the complete
grouped diff, surrounding TL;DR/§11.10 and validation gate, not a replacement
whole-chapter review or a new runtime assessment. Prior historical/grader exposure
remains disclosed. Current voice/procedure/architecture remain loaded.

The printed `--self-test --receipt PATH` invocation matches the frozen foundation
CLI's parser and reports fixture/oracle controls with runtime_acceptance false.
It cannot grade the student client. The text says that plainly, permits only an
ownership/public-API plan after accepted Chapter 10, and requires the actual
client-runtime invocation before implementation integration or acceptance.
This resolves the circular dependency on a not-yet-documented public seam without
waiving public/custom transport, lifecycle, identity or inherited checks.

Byte comparison confirms the entire §11.10 coverage table and following acceptance
clauses unchanged. The historical grader remains diagnostic. The coordinator still
owns release; no Chapter 11 implementation or live gate is closed here. Existing
prose lint passes hard rules. Scoped review whitespace verification passes. No
checker execution, compiler, provider call or source edit occurred in this review.
