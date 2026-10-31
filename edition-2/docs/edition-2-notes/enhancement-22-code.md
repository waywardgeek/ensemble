# Edition 2 Chapter 22: reusable GUI and terminal fleet

The recorded full revised runs at `a02dc4ce` score **200/200 on both accepted
and legacy source**: accepted 2.286 seconds, legacy 1.615 seconds, both exit 0.
A later row-attribution correction has a focused accepted GUI baseline and
restored control (100/100); with unchanged fleet 100/100 its final accepted total
is composed 200/200. The earlier actual full runs and legacy result retain their
source identity; no later full or legacy run is claimed. Independent combined
review remains pending.
The grader preserves the fourteen original rows: original 24 GUI reuse 100 plus
original 25 terminal fleet 100. Both original student grades remain **0/100**;
their import/launch and omitted-model failures were not confinement or fleet
results. No original grader, frozen runtime or current main/crossover file changed.

Sources are direct `solutions/edition-2/ch22` tree
`30fb05d154b9808eb88622aaff9c1424cd7e8af1` and legacy `solutions/ch25` tree
`43b66f047c6ba6885b4a26371df5af2ac6177b31`. Evidence lives in
[`enhancement22`](../../solutions/evidence/ch25/enhancement22/).
`final-source.json` binds all six command/grader/template/test files, both frozen
trees and the measured manuscript. `accepted-final22.json` and
`legacy-final22.json`, their stderr and `*-run.json` files retain exact commands,
exit status and wall time. The focused real-client race check passes in 2.219
seconds; `final-vet.txt` is clean. These checks exercise submitted binaries and
outside applications against HTTP/browser/TCP peers; they do not mock Agents,
actors, histories, fleet managers or jobs.

Commands from the repository root:

```sh
go run ./cmd/grade-edition2 -ch 22 -json solutions/edition-2/ch22
go run ./cmd/grade-edition2 -ch 22 -reuse-legacy -json solutions/ch25
```

The existing `-gui` flag selects a separate GUI module. `-reuse-legacy` selects
that export's existing public constructor and module layout. Its old `AgentHooks`
constructor receives the actual Agent actor's Send method and actual history;
there is no fabricated Send implementation. The accepted consumer uses the
owner-based public root and Agent directly. This historical compatibility does
not teach a callback bundle as the new design.

## What the revised observations establish

The outside program constructs a public history event, sees its actual replay,
delivers a prompt to the real Agent and observes the final response Artifact.
It serves selected embedded component modules without the application shell,
selects a distinctive disk override, and exercises both a correlated no-browser
error and a successful browser relay response preserving the caller ID. A
separate built GUI serves its embedded shell from an unrelated working directory.
Fresh listener addresses and actual child lifetime replace port 8088 assumptions.
An HTTP 200 directory listing cannot substitute for a shell containing scripts.

Resolved dependency sets have positive controls: the real public root and
outside headless MCP consumer exclude the optional GUI/WebSocket implementation,
while the external GUI consumer and actual GUI binary include it. The headless
consumer completes an actual initialize/discovery exchange. Actual core and GUI
binary graphs are inspected too. Legacy's shipped combined command correctly
appears as GUI-dependent; this does not make its public headless library coupled.
The reused real model-switch driver observes three completed turns and six
provider requests selecting Sonnet-5, Opus 4.6 and Sonnet-5. It awards no duplicate
status/accounting points.

These numeric observations do **not** prove runtime ownership, factory propagation,
absence of every callback design, JavaScript rendering or general relay lifetime.
The public-root row is titled accordingly. Existing mandatory source review and
accepted tests remain the evidence for those separate obligations. In particular,
the current optional blind listener uses a real Node peer rather than requiring
a named Go WebSocket adapter; source names are not pass conditions.

Fleet checks run the actual CLI with a typed primary and supported Sonnet-5
external-provider fixture. They inspect correlated submission results, two-child
fan-in, distinct valid parent/child saved conversations, a genuinely held live
child, shutdown removal, and three-generation questions/replies. Correlation IDs
are excluded from the reply-sender observation so an ID containing a parent's name
cannot substitute for the actual sender. A real command reaches a file barrier;
ordinary terminal input wakes it before its independent 60-second callback timer,
and waiting on that same handle retrieves its later arithmetic output. Active-child
submit refusal and exactly three corrective wakes are observed separately.

The child-state check accepts either a snapshot or valid complete JSONL journal,
as the original contract allows. It looks at actual stored text values, not the
parent's echoed spawn arguments or empty files. Human-address delivery/reply,
same-name question isolation, failed-turn nonretry, restoration and stronger
recursive lifetime/ownership guarantees retain their separate accepted real-CLI
and source-test proof. This numeric three-generation case is not a human-reply test.

## Preserved fixture failures

First focused accepted fleet output failed three fixture assumptions: raw save
searches matched task names in the parent's spawn arguments, running notices could
say `job N` rather than `handle N`, and exhausted-child joins may correctly return
an error. `initial-fleet-fixtures.txt` is explicitly a transcription of tool output,
not a newly captured raw run. Its corrected focused control passes in 0.190 seconds.
The first complete accepted run scored 130/200 because the external fixture called
an absent `ServeWS` method; its actual public API exposes `Handler`. Correcting
that fixture produced the first green 200/200 stage at `273b4fbb`; `stage-source.json`
binds that earlier source.

Initial legacy 60/200 was masked by two fixture mistakes: its application must
create the selected data directory, and a successful submit may receive a final
tool-free assistant request. Treating every tool-free request as a recall judge
returned `[]` instead of the terminal sentinel despite actual successful fleet
work. After those corrections, the full run scored 185/200; the remaining snapshot-only
assumption rejected its valid child journals. `legacy-journal-fleet.txt` then passes
focused fleet checks in 0.246 seconds. These outcomes remain beside the final
200/200; none is relabeled as an implementation repair or seven independent bugs.
`accepted-refined22` and `legacy-refined22` retain the intermediate 200/200 controls
before the final sender/Artifact predicates.

## Sensitivity and inherited proof

Twelve distinct source defects were applied in disposable copies. Nine initially
failed their intended observations. Two initially **passed**: reply sender removal
still matched the parent's name inside `message_id`, and omission of final Artifacts
still matched `turn_ended.text`. Both exposed grader gaps. The final predicates
exclude correlation IDs and require actual `part_final` content; both restored
defects now fail the intended assertions. A third attempt, hiding live children,
initially failed only through a coupled premature-submit timeout. Inspecting the
already captured status before returning that session error now directly records
`status omitted the held live child: []`, alongside the unchanged coupled timeout.

The remaining defects cover shared child saves, truncated fan-in, four corrective
wakes, lost late command output, omitted message wake, directory-listing deployment,
ignored disk override, lost relay caller ID and a transitive headless WebSocket
import. `sensitivity.json`, all patches and raw outputs preserve the original
attempts; `sensitivity-corrected.json` preserves the three focused follow-ups.
Shared saves cause coupled fleet failures as well as the intended missing-isolation
observation. Message-wake omission is a causal liveness failure: the command is
still held and the ordinary timer cannot explain the successful control. These
are twelve defects with overlapping observations, not twelve independent proofs.
The nine unchanged cases were not rerun solely for final-source freshness; their
receipts precede the three narrow predicate changes. The first batch lacks a
contemporaneous full grader hash manifest, so it is attributed to that recorded
pre-correction source state, not falsely bound to final bytes. Both final restored
focused controls pass (fleet 0.181 seconds, reuse 0.664 seconds).

`inherited-source-bindings.json` verifies twelve relevant production/test files
are byte-identical to their accepted original-24 (`0430df4`) or original-25
(`891d513`) sources. Reuse the actual evidence and its limits in
[ch24-code-review](ch24-code-review.md), [ch25-code-review](ch25-code-review.md),
and the corresponding original-numbered evidence directories. In particular,
external component rendering, reverse-order same-ID relay routing, disconnect/close
lifetimes, factory/parent ownership, human reply, question collision, restore and
failure-nonretry proof remain mandatory. The old directory-listing survivor and
relay cleanup timeout remain historical failures; no new paid run or broad
unchanged product sweep was spent to restate them.

Initial review found one further attribution gap: the headless WebSocket defect
already failed the framework/root rows but still earned the 15-point headless MCP
row. That row now receives the existing `./headless` dependency and actual-root
positive-control failures too; no successful predicate or runtime behavior changes.
The same isolated defect now explicitly fails `mcp-without-gorilla`; its prior
receipt remains intact. Focused accepted GUI baseline/restored controls pass in
1.029/0.665 seconds, with the defect failing in 0.846 seconds. See
`sensitivity-row-corrected.json` and `post-review-source.json`; the earlier
`final-source.json` remains the binding for the actual full runs. This is an
attribution correction to the same twelve-defect evidence, not a thirteenth defect.

## Exact published examples

`samples.json` binds manuscript SHA-256
`d2f514380905245e14720fbf3d9d9a1a6555e813077487da3385ab5505bfdf3e`
and all sixteen fence identities. The full outside Go example compiles; real
Chrome renders its component-only `/demo` and the shell URL returns 404. Full mode
replays `PUBLIC_REPLAY`, delivers two actual prompts, switches the next provider
request from Astra to Sol, returns a correlated browser relay result and serves
its selected disk override. The headless and GUI dependency commands give their
expected negative/positive controls. The signature, asset accessor, factory type
and actual factory callback match their owning source; the explicitly partial
AgentSpec vocabulary has both stated fields (`samples-go-bindings.json`).

The exact fleet setup and prompts run through the actual CLI with a faithful
scripted Responses peer. Line wraps in the prose prompt are folded into one
terminal input. Three actual children write add/multiply/subtract modules, run
their delayed assertions, ask/answer the negative-integer question and submit.
The input hint wakes join job 13; its same-handle later result contains all three
submissions. Independent combined assertions pass and normal EOF retains all
child saves. There are 22 main requests (two GUI plus twenty fleet) and four
successful auxiliary judge requests. The selected synthetic API-key environment
is one of the published launch choices; no real credential/profile or paid
provider is used. Scripted decisions prove client/tool mechanics, not a second
autonomous coding result or model quality. The historical 9.649-second GUI task
and 48.198-second actual Astra/high fleet task retain their original provenance
and hint-compliance limits.

Initial sample attempts are retained: the harness first mistook a public human
replay for `part_final`, then waited for the wrong settings-envelope name, then
incorrectly required auxiliary judge requests to carry main-request reasoning
settings. The last attempt had already completed the task but forced valid judges
into fallback; it is not a valid empty-judge control. The corrected peer supplies
complete successful responses and all four `[]` controls pass. None required a
runtime or manuscript repair.

## Size and review boundary

`source-counts.json` separates whole-line comments, blank and active physical
lines, counting inline comments with code. New grader drivers/dispatch have
789 active lines and 31 comment-only lines at `a02dc4ce`; the row-attribution
follow-up adds six active and two explanatory comment lines (`post-review-count.json`).
The outside-consumer template adds
35/3 and the opt-in tests 42/4. Original two harnesses/check tables total 1,198
active and 261 comment-only lines (1,585 physical). This is a bounded external
consumer plus four fleet scenarios reusing existing peers, GUI model-switch and
CLI helpers, not a new framework. The command gains one chapter dispatch and one
explicit historical-API option. Comments explain external boundaries, post-submit
classification, stored-text isolation, correlation and source-proof limits;
the ratio is below Bill's teaching target and requires reviewer judgment rather
than padding. The one new exported production API, `Chapter22`, is documented.

Mandatory `.agents/skills/ensemble-coding/SKILL.md`, complete carryover/workflow
and current enhancement/map were reloaded after compaction before this work.
Independent acceptance is the remaining gate; no Chapter 23 work is started.
