# Edition 2 Chapter 15 grader — final submission

The revised grader preserves original Chapter 16's eleven weights: **97 memory
points plus 3 for corrected Edition 2 Chapter 14 parity**. Final focused accepted
checks pass **97/97**; adding the unchanged measured parity gives a **composed
100/100**, not a new full run. The first actual full runs remain accepted
**100/100** and legacy **3/100**. Independent combined acceptance is pending.

This is grader enhancement, not a new memory implementation. Original Chapter 16
at `994bf218` has SHA-256
`4e8bb1ce58f73aa64b71b5fd57e531460e64d80ca6b2f1e934a1e222deb77791`.
Accepted `solutions/edition-2/ch15` remains tree
`4a253c57dd62b4a6ad086618574ef46af7c2d724`; legacy `solutions/ch16` remains
`13591b9185ec9d53641ada874cb82c0681f3f8db`. Original graders and frozen sources
are unchanged. The historical **20/100 FAIL** belongs to intermediate `ccd6350`,
not the final accepted runtime. Its eight awarded abandonment points did not
establish recovery. Read the full mandatory
[.agents/skills/ensemble-coding/SKILL.md](../../.agents/skills/ensemble-coding/SKILL.md),
workflow and carryover after compaction; all three were reloaded before this
final handoff.

## Results and provenance

All new artifacts are in
[original-numbered evidence/ch16/enhancement15](../../solutions/evidence/ch16/enhancement15).
`final-source.json` binds the final command, checks, driver and test bytes.
First green was committed at `bf0d733c` before sensitivity work. Earlier reports
retain their own source hashes; later corrected predicates do not retroactively
change those measured results.

| Measurement | Actual result | Interpretation |
| --- | --- | --- |
| `accepted-first-full.json` and its run receipt | 100/100, exit 0, 2.695 s wall | Actual full first-source run, including parity. |
| `legacy-first-full.json` and its run receipt | 3/100, exit 1, 1.136 s wall | Actual full run; memory settings never apply, so all ten memory scenarios stop at setup. Only corrected parity passes. |
| `accepted-final-focused.json`, `accepted-final-restored.txt` and run receipt | 97/97, race PASS 2.261 s; 3.423 s wall | Final restored memory checks. With unchanged first-run parity 3, composed 100/100. |
| `legacy-final-focused.json` and run receipt | 35/97, exit 1; test 0.753 s, wall 1.171 s | Supplemental disk-configured legacy diagnosis; not the normal grader and not a full 38/100 result. |
| `final-vet.txt` | PASS | Affected grader/command vet. |

The ordinary command is `go run ./cmd/grade-edition2 -ch 15 -json`. The exact
legacy full command was `go run ./cmd/grade-edition2 -ch 15 -json -gui
solutions/ch16 solutions/ch16/agent`. For this chapter the entry point builds the
explicit GUI directory; that historical trailing core argument was not the built
module. The actual optional GUI executes the Agent and public settings path.
Focused receipts retain the exact Go overlay command and small probe source;
the overlay exposes the existing checks, without changing the runtime.

The first unused-import compile failure remains `initial-build.txt`. Earlier
`accepted-final-race.txt` (2.286 s) and other intermediate focused artifacts remain
separate from the final restored receipt. No original-grader, paid-provider or
unchanged broad parity run was repeated merely for freshness.

## What the corrected checks observe

| Original check and weight | Revised observation |
| --- | --- |
| `micro-handoff-compresses`, 15 | A real checkpoint invokes the submit-only compressor with both ends of selected work; the replaced raw work leaves later message text. |
| `planted-fact-survives`, 10 | The planted fact reaches the compressor, its fake preserves the fact only when supplied, and the exact returned memory reaches later context. |
| `graduation-fires-oldest-first`, 12 | A distinguishing populated band folds oldest files while leaving the newest out, then delivers the fold and retained newer memory. Folding every file cannot satisfy this selection proof. |
| `disable-enable-idempotent`, 12 | Actual observed order and unique identities survive different re-enable orders; an actual selected file is mandatorily edited, and ordinary enabled startup reads its changed bytes. Disabled bands disappear and do not compress. |
| `disabled-neighbor-refused`, 8 | An overfull session band cannot fold into its disabled neighbor; exact submitted memories remain once each. |
| `forced-handoff`, 12 | Low/near/high completed MAIN requests exercise new warning-text delivery, the restricted declaration set and release after a later small MAIN request. A separate high-usage auxiliary call must not force main work. |
| `replay-needs-no-llm`, 10 | Actual startup makes zero HTTP requests; snapshot and complete-log restoration give equal next requests containing the returned memories. Startup also rescans disk, so this is not a pure reducer/no-filesystem proof. |
| `abandon-on-restart`, 8 | After successful session populations, one graduation actually returns HTTP 500 or stays pending while its process is killed. Restart retains all source once, does not retry the compressor, and a later checkpoint completes one fold. |
| `memory-is-data`, 5 | Actual returned first-person fixture text appears in conversation, absent from top-level and inline system instructions. Compressor instruction semantics remain source review. |
| `fresh-start-populates`, 5 | A new directory with only actual memory files, no save/log, includes memory in its first request without auxiliary inference. |
| `ch14-parity`, 3 | The separately corrected preceding chapter still passes. |

Public settings patches are followed by the effective full snapshot before any
prompt. Actual `turn_ended` observations and main-request counts replace fixed
post-request sleeps. Normal EOF must exit successfully; a readable partial
artifact cannot excuse failure. The pending-fold branch observes the real held
HTTP call before killing and joining the process, then releases the external
fake. Completed earlier machinery remains in the exercised path. There are no
mocks or new runtime protocols.

Message readers support strings and text blocks, excluding echoed tool inputs
and results. Declaration/result echoes cannot substitute for retained memory.
Current tests select already-declared `claude-sonnet-5`; the forcing fixture uses
the original declared `claude-ch16-course` 4096-token row. No 5.5 row, model or
pricing was added. These fake-boundary selections make no live support claim.

The forced-handoff arithmetic numerically exercises **Anthropic** disjoint raw
input/cache-write/cache-read fields: cache buckets count once, output does not,
and lifetime or compressor totals do not replace the latest completed MAIN
request. It does not separately establish every provider's normalization.
The original exercise's minimum whole segments/files, low-watermark retirement,
warning logs, pure recorded reducer replay, all-provider placement/arithmetic,
parent ownership and GUI lock ordering retain their accepted source/integration
proof in [ch16-code-review.md](ch16-code-review.md),
`solutions/edition-2/ch15/ensemble/memory_test.go`, and the original-numbered
reviewer/runtime evidence. The generic numeric suite does not claim those extra
predicates merely because related checks pass.

## Legacy compatibility and corrected fixture assumptions

Legacy `internal/common/settings.go` has no memory branch in `ApplyRaw`.
Consequently the initial public-patch run observes ignored settings before
inference, not ten independent memory failures. A supplemental, saved overlay
preseeds its actual cwd `settings.json`; the ordinary patch is still sent and
other supplied snapshot fields are checked. Its omitted `disabled:false` means
default enabled in that public type. This is a diagnostic fixture, not a new
production grader option or credit for missing patch support. See
`legacy-final-fixture.patch`, `legacy-disk-control.py` and `legacy-final-run.json`.

The supplemental final **35/97** passes compression (15), disabled neighbor (8)
and forcing (12). The seven failed rows must not be read as seven independent
implementation defects. Legacy `internal/llm/compact.go`'s graduation path folds
at two files without checking the high watermark. The retained raw captures show
premature 8x and then 64x folds even with a 5000-byte budget. Later missing markers
and fold-result routing therefore include cascades outside this bounded fake's
session/8x scenario. `legacy-observed-*.json` preserves those actual requests.

The initial disk run was **23/97**. Its forcing failure exposed a grader mistake:
requiring normal tools immediately after a handoff whose own response still
reported 3950 input tokens. Bill's rule uses the latest completed request. The
corrected check allows either state at that intermediate point and requires
normal tools after the subsequent 100-input response. Legacy then passes that
row. Initial results remain; neither default-setting differences nor a flawed
intermediate assertion is relabeled a runtime defect.

## Sensitivity and reviewer corrections

`sensitivity-final.json` indexes **27 meaningful numeric defects eventually
detected**, with exact patches and complete failure sets. It combines targeted
reruns with unchanged earlier evidence; these are not 27 fresh runs on final
bytes and coupled failures are not independent proofs. `mutations.json` and
`mutations.py` retain the initial historical sweep, including invalid and
surviving attempts.

The checks detect omitted compression, truncated source/returned memory,
retained raw work, newest-first selection, disabled-band/neighbor violations,
ignored human edits, wrong observed ordering, duplicate identities, cache/output
or lifetime occupancy mistakes, sticky forcing, failed/pending source retirement,
startup recompression/retry, missing fresh-disk population, instruction placement
and nonzero exit after completed work. Important qualifications are explicit:

- The initial inline-system-memory defect survived at 97/97. Inspecting both
  system representations repairs the gap; `system-final` now fails that intended
  check. A valid real-runtime string/block representation control passes 97/97.
- The initially named `no-near-limit-warning` mutation merely changed warning
  wording. Requiring literal `90%` wrongly rejected valid behavior. The final
  predicate removes each exact scripted prompt/reply once and observes added
  residual text across the threshold. Alternate wording now passes 97/97; actual
  omitted warning append fails 85/97. This proves delivery, not semantic quality.
  `wording-usage-final` also rechecks the two usage mutants affected by this
  predicate; both fail the intended boundary on final bytes.
- Requiring literal `first person` or `first-person` likewise exceeded the public
  prompt contract. That predicate is removed. An alternate instruction using
  “I, from my own perspective” passes 97/97. The initial missing-instruction
  failure is retained but excluded from the final numeric kill count; instruction
  semantics belong to the separate source review, not a phrase matcher.
- Injecting an auxiliary response event survived because the selected later
  occupancy path overwrote it before observation. It remains an inexact surviving
  attempt. Using Engine totals instead of last-main usage separately fails at the
  high auxiliary phase; it establishes that particular coupled defect.
- The original premature-retirement mutation failed first at the HTTP-error
  branch. A separate pending-only selection now also passes restored source and
  fails with `pendingtrue source memory1 lost/duplicated after restart` while a
  real compressor is held and the real process is killed. Those raw receipts
  establish crash sensitivity without claiming the earlier failure did so.
- Retaining graduated source causes repeated folds and a completion timeout.
  That is a demonstrated liveness failure, not a reached final retained-text
  assertion. The ambiguous disk-edit mutation setup failure and first compilation
  error are not counted as killed defects.

Final restored focused checks pass under race detection after the corrections;
no frozen runtime was patched. Existing preceding-chapter mutation evidence is
reused for unchanged settings/completion/journal and parity behavior.

## Exact examples and retained live limits

`samples.json`, `samples-requests.json`, `samples-memories.json`, `samples-go.txt`
and `samples.py` bind execution to manuscript SHA-256
`bd2acd5cef67062ee2d9700e5349bc62fdea60ea7d4236ffc2094f00c24a78da`.
The exact vocabulary fence compiles; the published optional GUI build, complete
primary skill, small disk settings, Sonnet-5 launch and three historical prompt
texts run through the real implementation. Five MAIN plus three compressor
requests produce two session memories and one graduation; returned memories
appear in the final request and EOF exits 0. This is a local external-provider
fake, not paid inference, subjective recall quality or forcing: the frozen
Sonnet-5 row has window zero.

Final manuscript SHA-256
`933a67a7fae83951b47c64324e8be76eb8120c9cb1c8d7cd3ad3053534698b72`
changes only prose/rubric attribution from the executed version. An exact
comparison verifies all **11 fenced blocks**, including prompt text, unchanged.
`final-source.json` preserves both hashes without implying another sample run.

Historical five live trials retain their original model identities and failures:
missing reasons, invented uncertainty, wrong recall of facts actually delivered,
and extra tools despite the prompt. Their measured reductions of roughly
1.36–3.68 do not prove an 8x semantic compression ratio. No historical Haiku 5.5
result is relabeled or recommended. These quality/compliance limits remain in
the student and comparative reviews, separate from mechanical grader success.

## Size and review boundary

`counts.json` counts physical lines, blank lines and comment-only `//` lines,
with embedded fixture strings counted as support code. The new check/driver
production pair is 890 physical lines: 787 other, 74 comment-only and 29 blank.
The test is 26 lines: 21 other, 2 comments, 3 blank. Production comment share is
about 8.6% of nonblank lines, below the teaching target; comments explain the
causal settings/turn barriers, fixture boundaries, usage timing and proof limits
rather than padding a quota. The exported `Chapter15` and test entry have useful
documentation. The original memory grader comprised 1321 physical lines; neither
that size comparison nor passing scores alone establishes better design.

Only the separately versioned command/grader, this note and exact enhancement
artifacts are owned here. Original/frozen preservation checks are clean;
CodeRhapsody's concurrent changes remain untouched. Independent review decides
acceptance. No Chapter 16 enhancement has been started.
