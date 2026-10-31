# Chapter reorder: mechanical adoption

Stage E follows accepted runtime `7feb910a`, combined prose/grader review
`8935f486`, and reconstructed fleet management review `d085bd63`. This is
maintenance with prior-answer exposure. The complete mandatory
`.agents/skills/ensemble-coding/SKILL.md`, workflow, carryover, voice and chapter
procedure were reloaded after compaction before editing.

## Mapping and preservation

| Previous Edition 2 | Current | Exercise |
| --- | --- | --- |
| 17 | 15 | The Invisible Invoice |
| 18 | 16 | Leaving Anthropic for OpenAI |
| 19 | 17 | The Crossover (ungraded) |
| 20 | 18 | Giving the Agent the Web, and a Status Line |
| 21 | 19 | The Sandbox |
| 22 | 20 | The Ensemble |
| 23 | 21 | Managing the Swarm |
| 15 | 22 | Forgetting on Purpose |
| 16 | 23 | Auto-Recall |

Chapters0–14 and24–26 keep their identities. Current chapter files, heading and
section numbers, current references, index, grader filenames/functions/report
names, CLI help/dispatch, Make targets and application README follow this map.
The CLI still defaults to the matching numbered `solutions/chNN` directory.
Its build selection follows the exercise: GUI14–16/18/22/23, core19–21, Gateway26.
Memory retains corrected Chapter14 parity. Existing default-path integration tests
for Invoice/auth/web/Sandbox now use explicit current `agent/`, as the other
current grader tests already do; this avoids treating the still-old-order frozen
exports as reconstructed sources.

Invoice extends14, and its memory/judge accounting, cache-lens exclusion and
surviving-memory breakpoint descriptions now explicitly point forward to22/23.
Those obligations remain in the later memory/recall contracts. Memory extends
managed fleet21. Historical first-edition constructor and bisect stories retain
their original numbers, with explicit provenance. Historical model IDs, grader
rubric IDs, `grade.ChNN` APIs, evidence paths, measurements, dates and temporary
example program names remain unchanged. The handoff description in Invoice
matches the separately accepted opaque/signature removal correction.

No frozen solution, manifest, historical note, preface, assembled Markdown,
PDF or EPUB is changed. The publication record distinguishes current sorted
manuscript inputs from the earlier reading-copy artifacts. Cumulative exports
remain a separate reviewed operation; use explicit submission paths meanwhile.

## Validation

All advertised runs below use the renamed command with explicit `./agent`,
sequentially, and strip only `LLM_`, `OPENAI_`, `ANTHROPIC_`, `GEMINI_`, `EN_`
and `CODEX_` environment prefixes in test subprocesses. Source hashes are in
`/tmp/ensemble-stage-e-grade-source.json`; raw reports and elapsed measurements
are `/tmp/ensemble-stage-e-gradeN.txt` and `/tmp/ensemble-stage-e-grades.json`.
These are new mapping checks, not relabeled old-numbered results. Runtime bytes
include the coordinator's separately reviewed handoff-label fix checkpointed as
`57920dd3`; this coder made no runtime implementation change.

| New chapter | Actual result | Rows | Seconds |
| --- | --- | --- | --- |
| 14 | 100/100, exit0 | 8 | 1.092 |
| 22 | 100/100, exit0 | 11 | 2.970 |
| 15 | 100/100, exit0 | 7 | 0.480 |
| 16 | 100/100, exit0 | 7 | 1.228 |
| 18 | 155/155, exit0 | 11 | 0.766 |
| 19 | 100/100, exit0 | 7 | 0.844 |
| 20 | 200/200, exit0 | 14 | 1.623 |
| 21 | 100/100, exit0 | 7 | 1.175 |
| 23 | 100/100, exit0 | 10 | 16.387 |

The compile-only package check passed before these runs; it is not described as
executed behavioral tests. Existing Stage D mutation evidence remains unchanged:
this move adds no behavioral grader predicate. Package, link, dispatch and prose
checks follow below. Independent Stage E review remains pending.

`go test ./internal/gradeedition2 ./internal/grade ./cmd/grade-edition2`
passes (35.388s wall; gradeedition2 35.046s, inherited grade package cached).
Vet passes. Hard prose lint passes on all affected chapters and epilogue, with
existing soft length/person/paragraph warnings retained; the two final wrapping
edits have a separate passing lint. Raw outputs are
`/tmp/ensemble-stage-e-{tests,vet,lint,final-lint}.txt`.

A built command's help and thirteen empty-scratch default-path/admission cases
pass: the missing submission errors name the correct core/GUI/Gateway paths,
and17/24/25 refuse as ungraded before an application launch. This tests dispatch
without pretending that the old frozen defaults already contain reconstructed
source. Raw: `/tmp/ensemble-stage-e-dispatch.json`.

`/tmp/ensemble-stage-e-preservation.json` records27 ordered numbered inputs,
matching actual index titles and section numbers,136 valid local links, unchanged
preface, frozen exports and publication outputs. Legacy grade APIs and fixture
identifiers compare unchanged across the permutation. Every fenced example is
unchanged except its current solution path or grade-edition2 chapter argument.
No artifact was assembled or rendered. The initial broad preservation query
included the coordinator's two newly committed live receipts and returned false;
the corrected query names frozen exports/publication outputs specifically.

The post-run source comparison finds only the coordinator/reconstruction owner's
`agent/ensemble/cache_test.go` OpenAI marker expectation update. It changes no
production or grader bytes and is outside this checkpoint. The original run
manifest remains untouched; no new source identity is assigned to those runs.
Moved-file diffs are retained at `/tmp/ensemble-stage-e-{chapter,grader}-diff.txt`
for review; unchanged grade predicates reuse Stage D sensitivity evidence.
