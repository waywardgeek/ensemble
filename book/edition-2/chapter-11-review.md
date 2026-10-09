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

## Partial runtime preparation and teaching closure: a36aa8a / b43fc3f

October 8 PDT / October 9 UTC 2026. The independent reviewer accepts the
published partial CLI runtime preparation at `a36aa8a` and the two-file author
clarification at `b43fc3f76f2631e70cd5d24aadb06a2927c4fdb3`. The separate ownership
plan disposition remains [4d9c015](chapter-11-plan-review.md). This review closes
the finite preparation/teaching handoff, without declaring a Chapter 11 runtime
positive or complete acceptance. Prior historical and grader exposure remains
disclosed; the reviewer is not the student.

The read ledger includes the complete reloaded coding skill, voice, writing
procedure and architecture; current Chapter 11 contract and complete grouped
author diff; all ten preparation files at a36aa8a; the imported Chapter 9 model
fixture and relevant retained Chapter 11 oracle helpers; and the exact original,
final, identity-control and retained-foundation receipts. No student implementation
or mutable draft supplied an expected answer. Protocol requirements here are
checked against the published contract; no new external provider claim is made.

The checker starts with a real CLI/stdio successful-call parent before attempting
its three refusal cases. That parent requires literal discovery/list/call traffic,
one actual notebook append, exact canonical Job artifact, alias declaration,
recorded model continuation, selected environment/path behavior and real metadata
whose exclusion can be observed. It does not award exclusion for an absent peer.
Invalid arguments must retain a controlled error with zero remote call; the mixed
result must preserve the already performed external effect while rejecting all
accepted-result bytes. Missing configuration must fail before process/model
launch. The remaining coverage table expressly retains full schema, authority,
lifecycle, persistence, alternative/public transport and inherited obligations.
Neither an exit code nor the partial receipt replaces those gates.

Independent local verification reproduced these results:

- `PYTHONDONTWRITEBYTECODE=1 python3 scripts/edition2/test_ch11_runtime.py`:
  all four methods pass, including real fixture-peer notebook writes. Canned
  captures test predicates only; none is a student runtime positive.
- The final baseline's complete 164-entry map, immutable source revision
  `70d86f7419c82fcf7cb8a394d54e472feccd2eed`, retained executable and build association
  pass the actual binding function unchanged. Independently copied valid receipts
  then produce the precise binary-hash, valid-path source-hash and missing-entry
  refusals. The copies are removed; no CLI is launched by this binding check.
- The original baseline remains 1 pass / 5 fail / 3 blocked under its original
  checker hash. The corrected baseline is 0 pass / 6 fail / 3 blocked: accepted
  Chapter 10 rejects the new option before any peer or HTTP request. This is an
  expected missing feature, not a Chapter 10 regression. All final checker hashes
  agree with a36aa8a; the earlier vacuous predicate is not relabeled as corrected.
- The retained foundation and deletion receipts match their unchanged script
  hashes and contain all 90 passing oracle rows and nine intended deletion
  failures. These stored results were inspected, not represented as newly run
  foundation or student mutation tests.

The complete source/binary binding runs before launch and again before publication;
checker/import identities are also bracketed. The three identity refusals start
from a successful identity path, independently of the Chapter 10 runtime's
expected failure. The public provenance text correctly requires observed build
receipts: an association alone cannot prove compilation. Future runtime integration
still needs an actual passing Chapter 11 parent, public adapters, complete checks
and source-bound mutation results. There is no requirement to finish future runtime
acceptance before the coordinator releases implementation.

The author's clarification resolves Q1/Q2/Q3 and the root-parent reachability
issue exactly within the reviewed scope. Standalone offline logs report mcp:[]
without inventing identity; live and version-2 offline status retain their distinct
meanings. Shared bounded syntax belongs behind a reachable Ensemble-owned common
interface in a responsible neutral spoke; MCP and SessionCodec keep policy and
raw replay preservation. Valid owner-local exhaustion fixtures require the last
successful allocation plus overflow refusal before mutation or handoff, without
production setters. The student's acknowledgment and concrete amended parent
path can occur in the released task before affected implementation.

Byte comparison confirms the original printed JSON fixtures and §11.10 coverage
table unchanged. The printed runtime invocation exactly matches the published
command. Direct feedback preserves the original plan questions and pending student
confirmation. Existing external prose lint and scoped whitespace checks pass.
No runtime/source edits, builds, credentials or provider calls occurred. No further
preparation blocker was found; full public/runtime coverage remains pending.
