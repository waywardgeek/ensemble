# Chapter 4 validation

Accepted October 7, 2026 for checkpoint `edition-2-ch04-r1`. Bill's editorial
approval is separate. The fresh student used the new curriculum and accepted
Chapter 3 source; first-edition comparison followed the initial implementation
and actual runs. Actual source reads and teaching feedback remain in the export.

Runtime is `d25d3fd4e552cd17c75bf814c9899903878cfbd5`; initial revised live
evidence is `9341117`, with final evidence-only repair
`e1c64886564e7d1e8205ee08c17f98f32e001912`. The exact 780-file export is
`solutions/edition-2/ch04/`. Its manifest at
`solutions/edition-2/manifests/ch04-r1.json` binds source blobs, file hashes,
contract, skill, independent checks and evidence. Earlier attempts retain
their own identities and are not relabeled as corrected runs.

Independent runtime/evidence review was completed at `1b652a5`; final prose
review at `05b170a`. The reviewer did not implement Chapter 4, but did implement
the preceding chapter's human-client integration; that exposure is disclosed.

## Results

- All five Go modules pass vet/tests; main race and formatting checks pass.
- Inherited Chapter 4 grader: 100/100.
- Independent CLI/replay: 53/53; external public-consumer cases: 3/3.
- Ten original storage/lifecycle race controls and added response-admission
  and log-identity controls pass. Eleven targeted mutations are detected.
- Actual macOS human PTYs on all three supported APIs exercised jobs, waiting,
  input, source/report limits, artifact recovery, explicit kill, Delve showing
  42, and EOF cleanup. Real public-consumer runs exercised two Agents per
  provider, distinct handles, owned observations and foreign-handle refusal.
- All 51 corrected-runtime source hashes and seven verifier identity/manifest
  controls were independently checked. Rejected verifications leave the copied
  evidence unchanged. Evidence repairs required no new paid generation.
- Student teaching review, author dispositions, comparative code review and
  final manuscript proofreading are complete.

The coordinator also checks all five modules from the exported location;
exact commands/results are in `checkpoint-evidence/ch04-export-checks.json`.

## Improvements and limitations

The independent comparison records stronger Agent ownership, exclusive artifact
creation, retained output, joined process cleanup, transactional report cursors
and durable job transitions. Review corrected wire parsing in Jobs, redundant
owner access, mixed artifact/report metadata, malformed-response panics,
mutable log identity and snapshot aliasing. See `chapter-04-code-review.md`
and `data-structure-review.md` for rationale and scoped resolutions.

The GUI remains a stub. These runtime demonstrations are macOS evidence, not
Linux live evidence. Killing a local goroutine's job cannot forcibly stop its
function or undo its effects. Scoped tests and mutation audits are not an
exhaustive proof. No shared legacy Go grader was modified for Chapter 4;
earlier legacy regression receipts retain their original scope.

Full human runs, partial attempts, provider/coder mistakes and verifier defects
remain preserved. The final manuscript distinguishes them from the accepted
runtime demonstrations. No session is attributed to Bill unless he ran it.
