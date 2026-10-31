# Edition 2 Chapter 14 grader — final submission

Coder: GPT-6 Astra/high. Mandatory full reading:
`.agents/skills/ensemble-coding/SKILL.md`, carryover and workflow (reloaded after
compaction), original Chapter15 at `994bf218` (SHA-256
`6c6d953e754f73b8bafd45a2f6009cab5dce9d7557c008e79fcc4084f27c2e79`),
and complete original15 student/comparative records. Accepted direct solution
is `solutions/edition-2/ch14`, tree `0c751dbd3985dee1a80284461e31346471ccc984`;
legacy `solutions/ch15` tree is `86cb03d46dd900f382266343229908a38290aad9`.
Neither source, original grader nor first-edition manuscript changed.

## Contract and retained coverage

Run `go run ./cmd/grade-edition2 -ch 14 -json`. This chapter builds the optional
GUI; `-gui` selects another existing GUI command. Legacy invocation adds
`-gui solutions/ch15 solutions/ch15`. The original eight weights remain
15/15/10/15/10/15/15/5, totaling100 with no parity row. Chapters1–13 and their
shared implementations are unchanged.

The runner applies `context_target` through public `update_settings` and checks
the complete `settings_changed` snapshot before prompting. Ladder uses20000;
the restart experiment explicitly applies20000 versus400000. Keep/no-stub uses
400000, preventing ladder cuts from confounding per-round stubbing. The Agent
chooses its settings file; no grader path convention overrides `AgentDataDir`.
An actual `turn_ended` replaces the original400ms delay before EOF or SIGKILL;
normal EOF now requires successful process exit. The first local runner omitted
`subscribe`; `initial-test.txt` retains its zero-request setup failure. Adding
that existing public message made `subscribed-test.txt` pass in0.971s.

The manual grants existing `list_directory`, replacing undefined `think`.
Original ladder, survivors, pairing, keep/no-stub, restart, snapshot/replay,
crash-tail and known-bad-event checks remain. Added predicates require successful
correlated delivery of **all planted bytes** before later disappearance proves a
cut. A same-batch read follows `micro_handoff`; the next request must clear both,
while the public saved log proves the companion actually succeeded and the
ordinary handoff dispatch/result preceded its separate note.

The prefix scenario loads, invokes and unloads the external MCP peer. It checks
actual declaration/removal block kinds and schemas, fresh output-only automatic
content, correlated argument-dependent results, stable startup bytes and inline
system placement after the whole user turn. It reuses Chapter12's faithful
subprocess peer. Consecutive inline system messages form one legal group; the
independent reviewer caught a first-green predicate requiring coalescing.
The corrected group-boundary check retains rejection of user content after that
group. The retained regression and actual split-renderer positive control cover
this valid representation. This is observed request validation, not a claim that
this revised permissive HTTP fake itself rejects the request at its boundary.
The original strict HTTP400 source regression remains separate evidence.

## Results and exact examples

Evidence directory: `solutions/edition-2/evidence/ch15/enhancement14/`.
First-green checkpoint **35fe32fe** produced actual full accepted100/100 in0.814s
and legacy100/100 in0.842s, exit0. Its source and reports remain preserved.
After full-byte and system-group corrections, actual final full runs are:

| Runtime | Actual result | Receipt |
| --- | --- | --- |
| Accepted |100/100, exit0,0.755s|`accepted-final-full.json`, `accepted-final-run.json`|
| Legacy |100/100, exit0,0.769s|`legacy-final-full.json`, `legacy-final-run.json`|

These are complete runs, not composed totals. `final-source.json` binds grader,
command, test, frozen tree identities and final manuscript bytes. Focused grader
race checks pass in1.895s; affected vet and formatting pass. The race flag applies
to the grader test process, not the separately built runtime binaries.

Exact author samples at manuscript SHA-256
`55e3a4a4900f8cfbd10b65aa58ce000435d7f404129937b1c02394c8095cf816`
pass in an isolated accepted-source copy. The current Go vocabulary compiles;
the exact survivor/replay fences replace their actual owning functions and the
Context tests pass. Historical fragments remain explicitly historical, not
claimed current APIs. The exact GUI build/launch, complete skills and Python
settings/files fixture run with a small external HTTP peer:7human turns,
16requests,6successful correlated reads,6recorded cuts,1handoff, final manual and
note each once, normal exit0. `samples.json`, `samples-go.txt` and
`samples-requests.json` retain that evidence. The declared no-stub course model
supplies deterministic responses; this is no new real-provider or quality claim.
The initial sample guard stopped for concurrent prose-only hash updates. The
first validator then incorrectly selected the latest resent (possibly stubbed)
result; selecting the first correlated delivery corrected that validation error.
Both corrections are recorded in `sample-attempts.json`.

## Sensitivity and limits

Fifteen distinct targeted defects fail the intended behavioral predicate:
wrong effective target, empty successful read, partial first read, premature
same-batch handoff, missing durable handoff call, inline delta before ephemera,
addition disguised as text, removal disguised as text, empty MCP result, wrong
MCP arguments, empty automatic output, duplicated wire ephemera, stale cached
snapshot, rewritten startup declarations, and nonzero exit after completed work.
Raw patches/reports are retained with `mutations.json` and `partial-control.json`.
The nonzero-exit and wrong-target cases cause coupled downstream failures; those
are not independent proofs. The premature handoff fails on an actual orphaned
result before the next request and also prevents dependent replay.

Two meaningful positive controls pass100/100: a real renderer splitting a system
group into separate messages, and actual completion observations delayed650ms
(10.248s run). The latter exceeds the old400ms fixture delay while retaining
completed-turn ordering. Restored runtime controls pass; the final accepted run
covers all final assertions.

The initial partial-read mutation scored85: the new delivery assertion wrongly
passed on its marker, while a separate restart comparison failed. This was a
missed intended predicate, not a successful full-delivery proof. The strengthened
full-byte assertion now fails that exact defect. Separately, broadly retaining
pending ephemera caused1416requests and a completion timeout; it never reached
the intended freshness assertion. It remains a limited runaway result, not one
of the15 intended proofs. Narrow duplicate-wire and actual snapshot-cache
mutations subsequently fail the content-count checks directly. An extra marker
perturbation is retained but not counted as another independent plausible defect.

Unchanged source evidence is reused from `ch15.md`, `ch15-code-review.md`,
`evidence/ch15/mutations*.json`, `mutation-removal.json`, `round1-mutation.json`
and `reviewer-round1.json`: model feature gating, previous-batch keep, unseen
results, both bands, recorded cuts/replay, atomic known-event skips, unknown-event
refusal, journal append/checkpoint/backup/retention, address-free stubs and the
strict inline/ephemera HTTP boundary. Existing GUI browser receipts separately
prove effective editing and unsent-field preservation. Their original setup
failures and wrong-assertion attempts remain recorded. No new numeric guarantee
of every crash interleaving, disk atomicity, provider cache savings, native
speech, reasoning quality or universal context bound is claimed.

The original **45/100 FAIL** and partial live Anthropic/OpenAI task failures
remain historical. Only Gemini completed the retained live task. The explicit
course-only feature matrix governs these inline/stubbing tests; all existing
real rows keep their original false flags. No paid freshness run, frozen edit,
new feature, framework or mock was introduced.

## Size and review boundary

The two grader files total1243physical lines:1025other,153comment-only,65blank;
other includes embedded fixture text, not just Go statements. The focused tests
have63physical lines,56other and4comment-only. Much is retained original scenario
coverage; the additions replace timing/path/name assumptions with public settings,
completion, same-batch and actual MCP observations. Exported function/field
comments and explanations cover consequential boundaries without comment padding.
The original runner/harness/checks total1034physical lines; both versions reuse
existing external vendor machinery. Final source/evidence is submitted for
independent review; no Chapter15 work is released by this report.

Post-review manuscript reconciliation: author `21c09849` corrects §14.5's
projected-order explanation. Held survivors retain their Seq identities and land
after batch completion; the projection is not globally Seq-sorted. Final manuscript
SHA-256 is `6e29699fc23fa7a3b2d4a549524625b125ac7c69600baeb686754e86ad570117`.
All15 executable Go/sh/Python/YAML/JSON fences are byte-identical to the actually
executed55e3a4a manuscript (`manuscript-reconciliation.json`). No rerun is claimed;
original sample/source receipts keep their actual execution-time binding.
Independent combined review accepted the chapter at `f3d6e7f0`, zero formal rounds.
