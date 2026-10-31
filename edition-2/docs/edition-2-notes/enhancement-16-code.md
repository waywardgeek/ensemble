# Edition 2 Chapter 16 / original 17 grader submission

Coder: GPT-6 Astra/high. Mandatory `.agents/skills/ensemble-coding/SKILL.md`,
full carryover/workflow and current enhancement/map reread after resumption.
First-green checkpoint is `e9cf3f18`; this final submission awaits independent
acceptance. No original grader, frozen runtime, active application, historical
model identity or concurrent crossover work was changed.

## Scope and public commands

The separate grader preserves original17's ten weights, total100, with **no
parity row**. It builds the actual optional GUI and inspects the submitted core's
resolved dependency graph. The typed primary grants real `read_file` and
`micro_handoff`, distinguishing a tool-free judge from a main request. Existing
settings subscription establishes readiness; actual turn completion and successful
EOF replace request-count sleeps and ignored exit status. No settings patch is
needed for these default-recall fixtures.

```
go run ./cmd/grade-edition2 -ch 16 -json
go run ./cmd/grade-edition2 -ch 16 -json -gui solutions/ch17 -recall-data . solutions/ch17
```

`-recall-data` selects only the grader's corpus root under scratch cwd: accepted
`cr`, legacy `.`. It is not a new application flag or duplicate compatibility
corpus. Current provider fixtures use declared Sonnet-5. The faithful external
HTTP peer leaves the actual Agent, recall, history, renderer and GUI in the path.

The original fake holds its capture mutex while calling Route. Holding a judge
there also blocks the main response and request inspection. This small peer
captures under a short lock, then waits outside it. The full timeout scenario
observes a completed main turn while its judge handler remains held until
cancellation or cleanup. No shared-fake or runtime protocol was modified.

| Check | Points | Revised observation |
| --- | ---: | --- |
| bm25-indexes | 10 | Actual docs/handoffs/learnings/available-skill content, older/newest archive files and same-process disk edit reach recall. |
| bm25-scores | 10 | Competing hits distinguish rarity, saturation and length; no exact arbitrary score or tied-order requirement. |
| bm25-stop-words | 5 | Query stop-word trap stays absent; document stop words still affect length. |
| chunking-splits | 10 | Correct section/subsection/breadcrumb, with an excluded adjacent paragraph; no hard cut through indivisible prose required. |
| per-source-quota | 10 | Final mechanical memory reservation survives candidate crowding, successful document-only judge selection stands, and unused slots flow to documents. |
| judge-filters | 15 | Actual judge receives competing content; integer, quoted/wrapped, mixed/duplicate and empty choices affect main context. |
| judge-fallback | 10 | Malformed/HTTP-failed/held judge yields fallback; empty corpus and short query continue without recall. |
| recall-is-a-spoke | 10 | Built dependency graph resolves submitted recall, imports only hub/stdlib, and has no reverse implementation-spoke edge. |
| recall-is-own-kind | 10 | Recorded recall bytes persist once across turns and restart without the file or a startup HTTP call; actual checkpoint removes them. |
| injection-capped | 10 | Complete formatted attachment stays within6144 bytes and its capped source is not reattached next turn. |

Recorded recall identity is read from the existing save Log. Actual provider
content accepts string and text-block representations, excluding tool arguments
and results. A renderer may coalesce user data without changing stored entry
identity. The lifetime check exercises ordinary snapshot restoration; pure full-log
replay and parent semantics retain separate accepted source-test/review evidence.

## Actual results and fixture corrections

All new artifacts are under `solutions/edition-2/evidence/ch17/enhancement16/`.
`final-source.json` binds final grader, test, manuscript and unchanged source trees.

- `accepted-first-green.json` and `accepted-first-green-run.json`: actual full
  **100/100**, exit0,16.563s, at the first-green grader hashes.
- `legacy-first-full.json` and `legacy-first-full-run.json`: actual full
  **60/100**, exit1,16.327s, at those same hashes.
- `accepted-corrected-focused.txt`: final nine runtime rows pass,0.74s,
  omitting only the unchanged15-second timeout. Final race check, which also
  checks the dependency graph, passes2.229s; vet passes.
- `legacy-corrected-focused.txt`: final runtime rows earn50/90,0.64s.
  Unchanged measured topology adds10. The final totals are explicitly
  **composed100/100 accepted and60/100 legacy**, using the unchanged timeout
  observation and topology where needed. These are not new full runs.

Initial `accepted-initial.json` is actual10/100: the driver assumed `CH02_LOG`,
while the GUI sets its own event-log filename. Nine runtime checks stopped at
that missing observation file. Reading the public save Log corrected the fixture;
those failures were not nine demonstrated recall defects.

`accepted-save-log.json` is actual95/100. Its supposed stop-word-only trap used
heading “Archived note”, while the query legitimately included “archived”. The
accepted index includes heading text. A neutral “Record” heading removes that
accidental hit. The failed run and earlier hashes remain.

The first legacy quota scenario used dated session files, which its memory
loader also populated into the live band. Its earlier graduation behavior then
made an auxiliary request, masking the recall assertion. Ordinary archive names
outside that date-number pattern now isolate recall in both implementations.
The corrected legacy scenario actually attaches three documents, failing final
memory reservation. Same-process refresh and duplicate ordinary/capped recall
also fail. The shared result map retains the last diagnostic within a weighted
row, so the refresh line does not enumerate every source-discovery failure;
legacy's newest-file exclusion has separately inspected source support. Do not
interpret the initial auxiliary request as a recall defect.

A resumed focused invocation accidentally reused an unbound pre-existing temporary
binary. Its failures are preserved in `unbound-temp-binary-focused.txt` and are
not attributed to frozen accepted source. Rebuilding the actual frozen GUI yields
the passing control above; `focused-source.json` records the binary hashes.
The sample launcher's first HTTP handler also assumed block content and rejected
a valid judge string. `samples-initial-error.txt` records the setup failure;
corrected extraction handles both forms. Neither required a product change.

## Sensitivity and restored controls

`mutations.json` contains28 applied attempts; `mutations-extra.json` contains
three follow-ups. All compile and fail the intended relevant check, with exact
patches/raw output retained. They are **31 attempts, not31 independent proofs**.
In particular, the first document-stopword patch called query tokenization, which
also deduplicates tokens. Its coupled ranking failure is preserved; the narrow
`document-stopwords-only` patch removes only the shared stop words and fails the
intended document-length comparison.

The targeted defects cover source omission/newest exclusion/stale index;
raw frequency, missing IDF/saturation/length and stop-word handling; paragraph
and breadcrumb loss; candidate-only quotas, missing memory reservation/spare
slots and overriding a successful judge; empty/mixed/duplicate verdict handling;
fallback, short-query gate and timeout; ordinary/capped dedup, stored kind and
checkpoint removal;6144 cap, both forbidden import directions and nonzero exit.
The wrong stored kind is caught at actual checkpoint retention, not claimed as
an independent initial-kind inspection. The no-timeout defect is caught when
the held judge prevents completion at25.16s; baseline completion is the measured
15-second fallback. A process exiting17 after its ordinary run cannot receive
credit from its completed turns/save artifacts.

Reviewer identified that first-green paragraph assertions could miss omitted
paragraph splitting. The final fixture adds an excluded adjacent paragraph
marker. A new explicit spare-slot case likewise makes redistribution observable.
`first-green-gap-controls.json` reruns these two defects against exact
`e9cf3f18` grader bytes via a Go overlay: **both pass that older stage**, then
fail final assertions. These retrospective controls were measured after the
fixes; they are not relabeled as earlier-in-time executions. No arbitrary4096
hard paragraph limit or new runtime feature was introduced.

Scratch mutants were removed; restored accepted runtime/race and vet controls
pass. The focused test overlay, small runners and patches are retained for
reproduction, not installed as a new product/test framework. Earlier accepted
all-provider, interrupt, pure replay, owner and capped-identity propagation
proof in `ch17-code-review.md` and original-numbered evidence remains separate.
No new numeric semantic-relevance or all-provider claim is made.

## Exact examples and review limits

`samples.json` binds all nine fences to manuscript
`3b80e0917cc19de98904d54a9ee197ee59a7310ae99d275169cfb6d97c52e7c5`.
The Go vocabulary compiles. The actual published build, scratch-directory/corpus,
complete primary and Sonnet-5 chat launch execute unchanged with the two exact
questions. There is one tool-free judge request, two main requests, one recorded
recall and normal exit0. `samples-requests.json`, `samples-events.json` and stdout
retain the real boundary artifacts. The local fake chooses the actual numbered
candidate containing the planted fact. Its scripted answers establish command,
retrieval and retained-delivery plumbing, **not semantic judging or model quality**.
Historical original0/100 and all three successful paid CLI runs retain their
original earlier-source/model identities; no paid freshness run occurred.

New grader/driver total743 physical lines,677 active and39 comment-only;
test30 physical,24 active and3 comment-only. Counts exclude blank/comment-only
lines from active; mixed lines count active. Original grader files total1281
physical lines, with their existing check definitions reused. This is not an
application-size comparison. Comments explain fixture roots, completed-turn and
capture lifetimes, stored/wire identity and topology boundaries; roughly5.4%
of production nonblank lines are comment-only, below the teaching target rather
than padded to it. The exported entrypoint and capture method are documented.

Final independent review still governs acceptance. Chapter17 and Gateway work
are outside this submission. Required handoff:
`.agents/skills/ensemble-coding/SKILL.md`.
