# Chapter 7 zero-byte argument maintenance

This is disclosed maintenance by the Chapter 6–9 independent reviewer, with
later source/checker exposure. It is not a fresh cold student implementation.
Root reviewed the isolated Chapter 6 five-line correction and its independently
derived regression before authorizing this exact propagation. No later Chapter 9
implementation or tests were copied backward.

The isolated branch `edition-2/ch07-empty-args-r3` starts at accepted Chapter 7 r2
source `9ec94ef551b08de6fbd714bf95efee901db49400`. Runtime/test source is `55ee774bcaeb7b182c56e41b32ab19eca3c49e6b`.
Only `internal/llm/stream.go` changes among the 2,229 original files;
all 2,128 original evidence files remain byte-identical.
The added regression and original SSE fixture, and corrected parser, match the
reviewed Chapter 6 commit `788c5e9` byte-for-byte. The fixture is 2,668 bytes with
SHA-256 `e6f7b7446b3dab2fe5ef6466111a6790c31ac52d45eccc8296a416ea9a4971ef`.
Its EOF blank line is the original SSE frame delimiter; the corresponding Git
whitespace warning is intentional, not a reason to change the recorded bytes.

`before.txt` records exactly three failures (recorded, one-empty, two-empty)
against the original production source with the added test. All ten other
regression controls already passed. `local.json` records the same 13 cases
passing after correction, empty changed-Go formatting output, core vet and
complete core tests. The corrected parser retains the start object until actual
argument bytes arrive; whitespace, malformed/non-object replacement and invalid
start/type/terminal cases remain refusals.

`checks.json` binds 103 non-evidence source/asset files,
all checker inputs, the archived executable and 24 passing affected
command groups. These include 14 local CLI argument/effect/terminal controls,
six early public/PTY delivery checks, 61 wire cases, the 14 streaming contract
groups, seven public watch/pause groups, retained Chapter 5 assertions, actual
CLI overflow/recovery and intended deletions, core race, all seven nested modules'
vet/tests, and complete delivered-tree discovery of 2,231 files/eight modules.
Core vet/tests are in `local.json`, so every module is covered without repeating
the initial core checks. `cli-boundaries.json` keeps the exact provider-byte
positive and held-message-stop no-effects barrier separately.

To reproduce the affected checks from this main directory:

```sh
python3 evidence/ch07/empty-arguments-r3/validate.py . 55ee774bcaeb7b182c56e41b32ab19eca3c49e6b /Users/bill/projects/ensemble/scripts/edition2 /Users/bill/projects/ensemble-edition-2-revisions/executables/ch07-empty-args-r3-cli evidence/ch07/empty-arguments-r3/checks.json --stage 7
```

The runner verifies source identities before writing derived receipts and again
after checking. The executable SHA-256 is `0fb64d370e6b48553faa448709f5e33bfab4ba97e00fa160af01390ca44fd307`.
This affected run is not a fresh full 33-group Chapter 7 gate or a paid/browser
speech demonstration. Unchanged GUI/assets, original full-gate/mutation controls,
and live demonstrations retain their accepted r2 evidence and original identities.
No provider call or credential read occurred. No current main, frozen export or
tag changed. Root's independent review and final export/checkpoint decision are
separate. `review.json` binds the scoped results and evidence files.
