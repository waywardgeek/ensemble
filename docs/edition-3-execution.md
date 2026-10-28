> **Historical reset record — not the current execution procedure.**
> Bill has authorized the C# run under
> [the new chapter procedure](edition-3-csharp-procedure.md).
> The state and next-work directions below describe the reset checkpoint,
> not the current working tree, skill loading, or continuation authority.

# Edition 3 — restart baseline

Bill authorized removal of the previous generated Edition 3 work and a restart on 2026-10-25. He explicitly requested retaining the old manuscript for selective reuse and starting the new manuscript from Edition 2.

## Current state

- Previous student closed; no new student, reviewer, or author assigned.
- `solutions/edition-3/main` is empty. Previous generated implementations and chapter snapshots were deleted; their Git history was not rewritten.
- Old manuscript moved intact to `/Users/bill/projects/cadence-backups/edition-3-manuscript-before-restart-20261025T042452`, with bytes and modes verified.
- New `book/edition-3` contains 47 source/support files copied byte-for-byte from Edition 2's top-level `book` files, including Chapters 00–26. Generated publications, the assembled manuscript, rendered print interior, caches, and the separate companion book were not copied. No old Edition 3 revisions were merged.
- Old implementation backups, 112 identified Edition 3 temporary artifacts/workspaces, and 568 old run-note/evidence entries were removed. A follow-up cleanup removed `docs/edition-3-evidence` and 154 additional identifiable temporary entries created by our work, including older chapter workspaces and completed runtime/website receipts. Uncertain, system, and Laura-related temporary files were left alone. Bill said the old evidence system will not be used; do not recreate its directory/receipt bureaucracy by default. Separate runtime records `ch27-design.md` and `runtime-speech-pause-observation.md` remain. Owned exercise servers and leftover headless test browsers were stopped; their isolated browser profiles were removed.
- Earlier editions, shared grader and skill source, running application source, unrelated work, credentials, and shared caches were not reset. Shared graders have not been validated against this new baseline.
- No commit, installation, publication, or new implementation/test/grading campaign was performed.

## Next work

Begin with Chapter 1 only after Bill and Cadence establish the corrected method and actual runtime skill loading. Skills must be loaded, not treated as activated by file reads. The last attempt to load `ensemble-coding` returned unknown skill; that remains unresolved. Do not silently substitute another file read or launch workers with that gap.

The old run's acceptance does not carry into this restart. No old implementation is a seed. The preserved manuscript is available for deliberate, selective reuse, not automatic reinstatement.

Automatic continuation reminders remain disabled during this human-led method work. A new run does not renew provider, credential, installation, restart, or publication authority.
