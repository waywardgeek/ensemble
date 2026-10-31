# Edition 2 Chapter 19 technical validation

Status: technical validation complete; independent combined review pending. Coder: `edition2_grader_coder`, October 11, 2026.

This is original Chapter 20, **The Crossover**, a prose-only chapter. There is
no original or revised grader and no numeric score. No runtime repair list,
new implementation, command dispatch or frozen export is part of this pass.
The mandatory `.agents/skills/ensemble-coding/SKILL.md`, full carryover and
workflow, current enhancement and map were reloaded after compaction. The full
original chapter and complete student/comparative records were read.

## Reused evidence, checked against the actual artifacts

Original source is `2161d2e9f26b6acecfb1457e7053065063944fb4`, chapter SHA-256
`9a6d41326806d229d1dc7913297d7b5f2ef04ab861e1068f67d393625e9700ac`.
The accepted direct `solutions/edition-2/ch19` tree remains
`bd2998e68f976dbaf437b642b131d350e086242f` (144 files). Its original independent
acceptance is `04e450da2f38f880e9db9f2300e49ad19c831872`, zero correction rounds.
The paid experiment used source `fab031529cda55a76d530478f09de0cd08e7db36`,
tree `59010a4c2b333c2f74777b6d76079e8f2b3af90d`; only README changed afterward.

[The read-only reconciliation](../../solutions/evidence/ch20/enhancement19/historical-audit.json)
was produced by `python3 solutions/edition-2/evidence/ch20/enhancement19/audit.py`.
It verifies retained hashes and decodes the actual wire, event log and save:

- Five requests and five completed responses select `gpt-6-astra` at high
  reasoning, detailed summary, streaming and `store:false`.
- Two actual reads are followed by failing `run_command`, exact `edit_file`, and
  successful `run_command`. Correlated results contain three failures and three
  errors before the edit, then six passing tests. Applying the recorded unique
  replacement to the original file produces the retained final file exactly.
- The six-test file is byte-identical before and after, SHA-256
  `d0a5eaab886c93918f853cde4e19819f0ab2fc8b03bf8aae504f5c8152081ea9`.
  The existing independent reviewer reran both versions; this pass reuses those
  results rather than claiming another test or model run.
- The saved log matches the complete recorded event log. The original launcher
  records normal exit 0, no timeout, 27.46 seconds and stdin closed after the
  final answer. It did not wait for optional public summaries.
- The trace contains 161 ordinary text deltas and zero public summary deltas.
  All five reported cache-read/write grants and reasoning-token counts are zero.
  The latter is not a measurement of private reasoning. Plan requests omit the
  previously accepted paired explicit-cache controls.

The historical endpoint derives from explicit configuration and transport source;
the retained trace does not independently capture URL or authorization headers.
The signed-in profile was retained, not newly created or logged out by that run.
One bounded CLI repair proves neither interactive steering, GUI use, arbitrary
project reliability, cache savings nor Edition 3 authorship. Chapter 25's actual
ongoing crossover remains separate.

## Existing implementation and sensitivity

The accepted chapter change moves the already selected Astra row onto the
existing Responses renderer. It adds two rationale comments and six active test
lines, with no new runtime abstraction or configuration path. The source test
renders actual requests and checks both transport and configured high effort.
Its existing passing baseline and two intended failures are retained in
`evidence/ch20/baseline.txt`, `astra-chat-route.txt`,
`drop-reasoning-effort.txt` and `mutations.json`: restoring the Chat route fails
`model transport`; replacing high with low fails `configured model/reasoning lost`.
No fresh mutation run or broader inference is claimed. The independent original
model-boundary race result remains 1.027 seconds on its recorded source.

No legacy `solutions/ch20` snapshot exists, so no invented reference delta or
legacy grade is supplied. The original historical repair names and code excerpts
are narrative evidence, not current public API or new Chapter 19 requirements.

## Exact examples

Manuscript SHA-256
`1058dc2c2e68e73c2ae0c14c7205a5bf88717bd8f1c4fec614cca3182a442b09`:
all five fences are bound in
[samples.json](../../solutions/evidence/ch20/enhancement19/samples.json).
One is the retained historical cache diagnostic, not executable code. The other
four are the build/scratch/typed-primary setup, auth status/catalog and chat
launch, exact original prompt, and independent unittest/cmp postcheck.

The exact setup builds `solutions/edition-2/ch19` to `/tmp/ensemble-ch19`, copies
the original broken program and unchanged tests to a new temporary directory,
and loads the published primary. Unlike the historical live run's default tools,
this demonstration explicitly grants `read_file`, `edit_file` and `run_command`;
existing lifecycle/context tools remain part of the runtime. The shell variables
span the published blocks, as they would in the reader's terminal.

A temporary Go overlay starts the existing signed external OAuth peer, then
`enhancement19/samples.py` runs the exact shell blocks. A fresh local profile and
loopback provider replace real credentials and network endpoints. The provider
replays the five retained actual Responses SSE exchanges; it does not implement
or replace any Agent, parser, file tool or command. A small catalog fixture uses
the existing provider shape and the actual Astra identifier/display pair retained
in `evidence/ch19/live-plan-models.json`. It establishes local catalog display,
not current account eligibility. No private profile or GUI port is accessed.

The final overlay command was:

```sh
go test -overlay /tmp/ch19-sample-overlay.json ./internal/gradeedition2   -run '^TestChapter19Samples$' -count=1 -v
```

Its temporary overlay maps `internal/gradeedition2/ch19_samples_test.go` to the
retained `samples-overlay.go.txt`; no permanent test or grader was added.
Actual final result: **PASS, 1.003 seconds**. Five main requests select Astra/high and
carry the expected actual correlated tool outputs. One separate recall judge
gets a valid empty selection. The real CLI reads both complete files, reproduces
three failures/three errors, edits the source to the exact historical result,
and runs six passing tests. The explicit postcheck prints six named passing tests and `OK`; the shell block
then exits 0 from `cmp`, confirming unchanged tests. Shell status alone is not
used to infer the unittest result. Stdin stays open until the final completed text event, then EOF
yields exit 0 and a save. The known external replies supply the edit: this is
sample/transport/tool validation, not a new autonomous coding result or proof
that any fresh model would make the same repair.

The initial sample failed in **1.579 seconds** because the replay peer treated
its preliminary tool-free recall-judge request as a main coding request. It
rejected the missing reasoning field and then consumed replay batches out of
position. The preserved initial output and source hash, plus
`sample-peer-correction.patch`, record that setup error. The corrected peer
answers the judge separately with `[]`; no manuscript or runtime changed. A
scripted final success message from the first attempt was not accepted as proof
of actual file reads. Final assertions inspect the real correlated results.

Independent review found that the intermediate 0.934-second passing sample
still sent an incomplete judge response: its missing usage fields triggered
`Responses response missing usage` and fallback. That run establishes the main
tool task, not the intended valid-empty-judge control. Its full receipts and
binding remain under `before-judge-usage-*`. The narrow peer correction adds
ordinary completed-response usage fields. The final 1.003-second sample has one
judge request, a valid `[]` response and no judge failure/fallback diagnostic.
This is an external fixture correction, not a product or manuscript change.

## Preservation and review boundary

No product, original grader, command dispatch, original manuscript, historical
receipt or frozen source changed. The accepted source tree remains exact. New
artifacts consist only of this note, a read-only historical reconciliation and
one chapter-specific example launcher/receipt. There is no new production code
or exported application API to count or document. Existing source tests,
original comparative review and two targeted defects retain their original
provenance; no broad test sweep, new mutation count or paid freshness run is
claimed. The independent reviewer determines the combined chapter disposition.
