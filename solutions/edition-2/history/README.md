# Preserved student history

On 2026-10-07 the Edition 2 source moved into the outer Ensemble repository.
`../main/` is the working source. `../ch01/`, `../ch02/`, and `../ch03/`
were exported byte for byte, with original modes and symlinks, from the
commits in `migration.json`. Chapter 3's export covers its earlier tools
contract; its human-chat integration was unfinished at migration.

The manifest binds every exported file and the initial main tree to SHA-256.
Main begins with the tracked working files of Chapter 3's staged merge of
Chapter 2's accepted human client. `ch03-staged-merge.patch` preserves that
index change; the unstaged patch was empty. This is a migration checkpoint,
not a declaration that Chapter 3's new gate passed.

All original repositories, including `.git`, the merge index and MERGE_HEAD,
ignored binaries, and untracked files, were moved intact to the external
archive named in the manifest. Complete before/after inventories matched.
That local archive is not required to obtain the committed source. The three
bundles here preserve all original Git refs and committed objects, and were
verified with `git bundle verify`. Uncommitted/ignored files remain in the
external archive, not in the bundles.

Historical receipts deliberately retain their original paths and binary
hashes. A rebuild at a new path can produce a different executable hash.
Some historical verifiers use `git show ORIGINAL_COMMIT:path`; restore their
objects into a fresh outer clone before running those verifiers:

```sh
for chapter in ch01 ch02 ch03; do
  git fetch --no-tags "solutions/edition-2/history/$chapter.bundle" \
    "+refs/*:refs/edition-2-history/$chapter/*"
done
```

These archive refs are already present in the migration workspace. They are
not active branches or validation tags. Bundles make this restoration possible
without publishing special refs. Alternatively, `git clone chNN.bundle DEST`
reconstructs a standalone historical student repository. The Chapter 3 bundle
contains committed history only; apply its staged patch to its recorded base
commit to reconstruct the tracked merge result.

Future chapter snapshots identify a source commit in the outer repository.
A dedicated chapter commit and immutable annotated revision tag bind the
manuscript, exported source, and evidence. Earlier initial/reviewed attempt IDs
remain historical IDs rather than being relabeled as new validation tags.
