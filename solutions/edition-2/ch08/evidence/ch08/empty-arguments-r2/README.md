# Chapter 8 zero-byte argument maintenance

This is disclosed maintenance by the independent Chapter 6–9 reviewer, with later
source/checker exposure, not a new cold student trial. Root accepted the isolated
Chapter 6 correction before authorizing exact propagation to accepted Chapter 8.
No later Chapter 9 implementation or test file was copied backward.

The branch `edition-2/ch08-empty-args-r2` starts at accepted Chapter 8 r1 source
`446d7f27fcef052d6bb780901c786842dc5286c1`. Corrected runtime/test source is `c4d7aac6d1d72c1578fccf50d6dd7c7afbc05090`.
Only `internal/llm/stream.go` changes among the 3,308 original files;
all 3,195 original evidence files remain byte-identical.
The five-line production delta, independently derived regression and original
provider SSE fixture match reviewed Chapter 6 commit `788c5e9` byte-for-byte.
The fixture is 2,668 bytes, SHA-256
`e6f7b7446b3dab2fe5ef6466111a6790c31ac52d45eccc8296a416ea9a4971ef`.
Its final blank line is the recorded SSE delimiter and intentionally produces
a Git whitespace warning; the exact bytes must not be normalized away.

`before.txt` reproduces exactly three baseline failures: recorded, one-empty,
and two-empty. Ten other regression controls already passed. `local.json`
records all 13 passing after correction, empty changed-Go formatting output,
core vet and full core tests. Empty argument fragments now retain the complete
start object. Actual replacement bytes still replace rather than merge;
malformed/non-object/whitespace replacements, invalid starts/types and incomplete
streams still fail without authorized effects.

## Scoped validation and retained runner failure

`checks.json` binds 115 non-evidence source/asset files,
all checker inputs and the archived CLI. Its initial 26-command affected
run completed with 24 passes and two receipt-writer failures. Both arose after
tests when `Path.relative_to` mixed macOS `/var` and `/private/var` paths for the
disposable checker bundle. They establish no credited result for those two rows.
`validate-initial.py` preserves the exact original runner and its hash.

The correction canonicalizes only that temporary directory. It also adds an
explicit `--only` selection to rerun the two affected commands without repeating
completed checks. `retained-rerun.json` records both passing on identical runtime
source: retained Chapter 5 assertions and actual CLI overflow/recovery with its
intended deletion controls. The underlying fixtures, assertions and mutants are
unchanged. Existing Chapter 8 fixture adaptations are recorded in each receipt;
only a disposable bundle is adapted. The initial failed receipt is unchanged.

Taken together, the 26 affected command groups pass on this one runtime: 14 local
CLI argument/effect/terminal cases, six early public/PTY delivery cases, 61 wire
cases, all 14 retained streaming contract groups, seven public watch/pause groups,
retained prior assertions and CLI recovery/deletions, core race, all eight nested
modules' vet/tests, and complete delivered-tree discovery of 3,310
files in nine modules. Core vet/tests are in `local.json`. The exact original SSE
and held-message-stop no-effects barrier are separately visible in
`cli-boundaries.json`. No failed or truncated operation produces an accepted
response, final part or file effect in those negative controls.

The corrected full affected-run command, from this main directory, is:

```sh
python3 evidence/ch08/empty-arguments-r2/validate.py . c4d7aac6d1d72c1578fccf50d6dd7c7afbc05090 /Users/bill/projects/ensemble/scripts/edition2 /Users/bill/projects/ensemble-edition-2-revisions/executables/ch08-empty-args-r2-cli evidence/ch08/empty-arguments-r2/checks-new.json --stage 8
```

For the narrow receipt repair, use the same command with another receipt name
and `--only retained-prior-assertions --only retained-cli-overflow-deletions`.
The retained narrow run is the actual evidence; a fresh full rerun of the corrected
runner is not claimed. Both runs verify source identities before derived writes
and afterward. Executable SHA-256:
`eb079ff28db11beb0b513b66c0f7dbe53c0ae95f2558652a8130d1ea7c223741`.

This maintenance does not claim a fresh 52-group Chapter 8 gate, new native/browser
speech use, or fresh provider generation. Unchanged settings, browser assets and
prior full deterministic/live results retain their accepted r1 source identities.
No paid call or credential read occurred. Current main, frozen exports and old
tags remain untouched. Root's independent review and final checkpoint/export
choices remain separate. `review.json` binds the scoped results and all receipts.
