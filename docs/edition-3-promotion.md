# Edition 3 promoted to the repository root

Bill authorized archiving V2 in the same style as V1 and making V3 canonical.
The source boundary before relocation was `7c0a3000`.

## Canonical paths

- `book/`: Edition 3, Chapters 00–26, companion and publishing inputs.
- `agent/`: active C#/.NET 10 application (formerly `solutions/edition-3/main/`).
- `solutions/chNN/`: preserved Edition 3 endpoints, retaining historical numbers.
- `solutions/evidence/`: preserved Edition 3 evidence.
- `edition-2/`: former root Go source, book, endpoints and archival support.
- `edition-1/`: unchanged previous archive.

There are no compatibility symlinks or duplicate active source trees. Root Go
`cmd/` and `internal/` remain shared grading tools, not the application runtime.
The companion's common publishing helper now lives at
`scripts/companion-publication.py`, so current publishing does not import an archive.

Use the root README and Makefile for current commands. `make` builds the C# GUI
and creates the ignored root `ensemble` launcher. Supply `DOTNET=/path/to/dotnet`
when the SDK is not on PATH. The launcher preserves that selection while allowing
an explicit `DOTNET_HOST_PATH` override. The previous ignored Go executable was
retained at `edition-2/ensemble`; it was not launched or installed.

Historical progress entries, manifests and endpoint READMEs retain their original
paths and chapter numbers. They are receipts, not current launch instructions.
The active book, publishing scripts, source tests, policy and grader defaults were
updated for the promoted layout. The Edition 2 root grader now defaults to
`edition-2/solutions/chNN` rather than accidentally selecting C# endpoints.

## Preservation and verification

The pre-move manifest records tracked file SHA-256 hashes and modes. Verification
matched 5,880 relocated V2 files and 685 copied archive support files, plus all
4,103 promoted V3 endpoint files and 21 evidence files. The unrelated `.swp` and
ignored local state were not staged. No credentials were read, host restarted,
application installed, or push performed.

Evidence is under `/private/tmp/cadence-v3-promotion/`: `baseline.json`,
`preservation.json`, build/test logs, publication checks and the final grader report.
The final results are summarized at the top of `edition-3-csharp-progress.md`.

`make test` separates current V3 acceptance from `make test-legacy-go`. The broader
historical Go suite is retained but is not green: some checks assume obsolete
Go layouts (`cmd/ch01`, `cmd/ch02`) or historical pricing behavior that does not
match the preserved endpoints. Those sources and assertions were not rewritten
to manufacture a passing archive. This is not a new certification of V1/V2.

The assembled current manuscript and EPUB are regenerated. Existing V3 PDFs remain the
previously built publications: a fresh print rendering was attempted in a private
output directory but Chrome returned nonzero during teardown after writing an intermediate PDF.
Its diagnostic includes display warnings and a teardown-watchdog expiry; those
messages do not establish a unique cause, and the two-pass build did not complete.
The build now reports captured renderer diagnostics instead of hiding them.
No failed output replaced a retained PDF. Publication remains Bill's decision.

## Final verification

The promoted CLI passed Chapter 5's public route at 65/65. Root `make` and the
current Go tests pass; the C# runner reports 313 passed, 0 failed, and all three
GUI groups pass. Book and companion tests pass (12 and 3 respectively).
The staged path set was checked against the complete relocation manifest. On
macOS, path-based staging omitted some entries and introduced case-only aliases;
explicit verified blob/mode entries and alias removal restored exact intended
spellings, without changing filesystem contents. The final index has no extra
or missing paths. Whitespace findings inherited unchanged from the baseline
were preserved rather than rewriting the archives.
