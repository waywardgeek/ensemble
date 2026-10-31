# Edition 2 Chapter 26 / original 27: Gateway grader and examples

Grader coder: GPT-6 Astra/high. The entire
`.agents/skills/ensemble-coding/SKILL.md`, workflow and carryover were reloaded
before coding and after compaction. Final source and examples are ready for
independent review; this note does not itself grant acceptance.

## Scope and public contract

The seven original rows retain **15/10/15/15/15/10/20 points, total 100**.
`go run ./cmd/grade-edition2 -ch 26 solutions/edition-2/ch26` builds the actual
submitted `cmd/gateway` and an outside Go consumer. It discovers the module path
and uses the now-published `WelcomePayload.QueueDepth int` (`queue_depth`).
Historical compatibility is explicit: add `-gateway-welcome QueuedCount` for
`solutions/ch27`. This selects an existing public field, without changing either
implementation or inventing a product alias.

The original student score remains **0/100 FAIL**: its undocumented `QueuedCount`
spelling prevented probe compilation. The initially suspected module alias did
compile. Original grader, results, runtime and frozen exports remain unchanged.

The revised consumer checks actual bidirectional Noise and wrong keys; welcome
identity, version and counts; authenticated Source/header and complete body,
conversation and response mode; permitted/refused routing; real child delivery;
and narrow registration. A sender-authored owner header stays body text rather
than becoming authenticated provenance. Channel labels may name the authorizing
group or authenticated agent, the two valid conventions used by these snapshots.
A completed refusal RPC precedes a permitted send on another socket; recipient
order detects refused traffic admitted before that barrier. This is not a
same-socket barrier, as the initial comment incorrectly stated.

For receipt timing, OnDeliver stays held for a bounded 150 ms. A replacement
connection must still find one queued message and receive it intact. Handler
return then permits LastAckedSeq; a subsequent RPC on that client establishes
server processing of its receipt. Offline messages retain order and an explicit
resume cut. A clean process stop and a new actual process restore the exact
queued message from the private data directory. This is clean restart evidence,
not power-loss durability or exactly-once application processing.

The public dependency check requires the actual resolved Gateway package and
rejects that module's private runtime imports. It does not establish ownership
by counting names. Fresh selected ports, launched-child liveness and the fixture's
pinned server key establish the exercised server; all successful scenarios require
normal process exit. The user's listener on 8088 is untouched.

## Actual results and source versions

All artifacts below are in
`solutions/edition-2/evidence/ch27/enhancement26/`.

- Initial actual CLI runs: accepted **100/100**, 1.062 seconds; legacy **100/100**,
  0.943 seconds, both exit zero. `stage-source.json` distinguishes those
  pre-formatting measurements from the formatted first-green stage `6d53e013`.
- Later actual CLI runs, `accepted-final.json` and `legacy-final.json`: **100/100**
  each, 1.299 and 0.916 seconds, both exit zero. Despite their filenames, these
  precede the final C7 admission correction; their receipts remain unchanged.
- Final-source actual `TestGateway26` runs cover all seven rows: accepted baseline,
  restored accepted and legacy each **100/100**, exit zero. Whole-command times
  are 2.171, 1.028 and 1.036 seconds in `refusal-correction.json`. These are full
  chapter observations through the integration test, not a composed score or a
  claimed rerun of the earlier CLI command.
- Final affected race test and vet pass. The raw race output is
  `accepted-final-race.txt`; source hashes and both frozen tree identities are
  in `final-source.json`. Scratch mutation copies were removed.

Welcome-count observations initially debited the delivery row alone when used
there. They now also debit C3, whose title promises the counts. This preserves
individual row attribution as well as overall rejection.

## Sensitivity and retained limits

Seven new isolated source defects in `sensitivity.json` exercise the pre-C7-fix
source: zero queue counts, partial body delivery, forged authenticated sender,
receipt before handler return, omitted queue persistence, a real transitive private
runtime import, and exit 17 after otherwise successful shutdown. Each fails the
intended observation; baseline and restored controls pass. Receipt and persistence
failures also affect welcome counts; body/provenance assertions overlap. The
nonzero process exit is a lifecycle failure, not seven independently demonstrated
runtime defects. These measurements were not repeated merely for freshness.

Independent review then found a genuine survivor: register the forbidden child
key and groups immediately before returning the unchanged widening error. The
old check awarded **100/100**, including C7's 20 points. Preserve the exact
`reviewer-register-before-refusal` patch, output and source binding. The corrected
check attempts authentication using the refused key. For asynchronous clients it
observes welcome or actual disconnect; silence is not credited as refusal. The
existing permitted child still connects and routes a complete message.

That exact defect now scores **80/100**, failing only C7 with “refused wider
registration nevertheless admitted its identity”; whole-command time 0.953 seconds.
Its baseline/defect/restored and legacy receipts are `refusal-*.txt`. Thus eight
new source defects are detected across their recorded grader versions, with one
initial survivor explicitly retained. This is targeted sensitivity, not an
exhaustive security claim or eight independent proofs of all requirements.

Unchanged stronger evidence remains in `ch27.md`, `ch27-code-review.md` and their
original-numbered raw artifacts: actual owner/root diagnostics, dynamic groups,
capabilities, membership/contactability and deletion, corrupted state refusal,
handler/client lifetimes, actual actor submission and the three-process interview.
Relevant source tests include `TestChildrenAndPersistedPolicy`,
`TestCorruptStateRefused`, `TestConnectionDiagnosticReachesRoot` and
`cmd/gateway-agent`'s `TestGatewayUsesActualActor`. The accepted 25-defect record
retains its initial compile-invalid, coupled and mislabeled attempts; it is reused,
not relabeled as 25 new enhancement runs. Numeric import observations cannot
replace the mandatory source review of composition and required parent access.

## Exact published examples

`samples.json` binds all 14 fences to manuscript SHA-256
`55f7c95a2391c85e45efd6e8e2651d270e96fe1220c3ccd3c24b4731c09e7ab7`.
The actual three published commands were built and run in fresh scratch space.
The documented public-key/directory substitutions were filled, and free port
55800 replaced example 8093. The newly launched router's PID owned that listener
before clients started. All three processes exited zero, with four actual model
requests, four recorded replies and the resulting interview artifact in 2.071
seconds. The API vocabulary and construction excerpts compile in their actual
outside owning context. An additional direct Handler/NopHandler conformance
compile is recorded separately in `samples-api-final.json`; no three-process rerun
is implied by that final signature check.

This is a faithful scripted external HTTP peer using synthetic Codex credentials
and a loopback endpoint. It exercises the real parser, actor, public Gateway and
interviewer, not autonomous model judgment, account eligibility or a new paid
interview. No personal credential profile was read. The original actual Astra
interview remains four questions in **11.739 seconds** on `ab6d0516`, with its
separate launch/source provenance for effort/route and zero cache grants. Grader
fences reuse the actual recorded CLI invocations. The short policy fragment has
an explicitly illustrative key; the full runnable configuration receives actual
generated public keys. No new runtime protocol or feature was added for examples.

## Size, teaching and preservation

Line counts classify blank lines and `//`-only comments, with remaining lines
active (including embedded probe text). New driver/probe: **552 active, 23 comment,
590 physical**; opt-in integration test: 34 active, 2 comment, 39 physical.
Original checks/harness/probe total **803 active, 69 comment, 958 physical**.
These are grader/support counts, not application deltas. Reusing existing report,
build and module helpers avoids another framework. Comments explain receipt
ordering, bounded observation, public label freedom and asynchronous refusal.
Their proportion is below the teaching target; no filler was added. Chapter26
and the probe's exported callbacks/test have purpose comments; the reviewer still
assesses explanatory sufficiency independently of these counts.

The original grader and both frozen source trees remain intact. Later publication
migration must preserve required shared `internal/grade` support, including the
original weights, build/report types and module discovery. Current paths remain
unchanged during this gate; migration and new publication are not claimed here.
