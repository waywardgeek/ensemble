# Edition 2 source

The working implementation is in [main](main/). It is a separate Go module
inside the Ensemble repository; its optional GUI and public examples have
their own modules. The historical first-edition implementation remains in
`agent/` and the original chapter solutions remain untouched.

| Path | Purpose |
|---|---|
| `main/` | Authoritative development tree, extended by each fresh student |
| `chNN/` | Frozen source export for that chapter; never a parallel development tree |
| `history/` | Original student Git bundles, migration manifest and restoration instructions |

Follow the [current book](../../book/edition-2/) and its
[progress record](../../book/edition-2/progress.md) for the active chapter and
validation status. A working source commit is not automatically a validated
chapter. At the initial consolidation, Chapter 2's human client was accepted;
Chapter 3's human integration was unfinished. The migration preserves both
states without relabeling either one's receipts.

The original snapshots retain their original README/checkpoint files as
historical source. Their old nested-repository instructions describe those
earlier builds. Current development uses the outer repository and `main/`.
See [history/README.md](history/README.md) to restore the original Git objects
for historical source-bound evidence checks.

After validation, the coordinator commits the accepted source and exports that
exact committed tree with `scripts/edition2/export_snapshot.py`. The separate
manifest records its source revision, tree, blob IDs and file hashes. A
dedicated chapter commit binds the manuscript, export and evidence, followed
by an immutable annotated `edition-2-chNN-rN` tag. Corrections receive a new
revision; earlier tagged trees remain available. Neither export generation
nor a tag replaces the chapter's tests, live demonstrations and review.

All chapter Git operations run in the outer repository. Do not initialize
another repository in these directories or commit credentials or runtime
scratch files. Read the full mandatory coding skill before editing source.
