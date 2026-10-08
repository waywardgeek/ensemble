# Chapter 6 validation

Accepted for checkpoint `edition-2-ch06-r1`. Final independent code, teaching,
live-evidence and manuscript acceptance is recorded at `37a37f7`. Bill's
editorial approval remains separate.
The accepted predecessor is `edition-2-ch05-r1`, an annotated immutable tag at
`7a6ef036322e1cf362894bd30b073fc8399a3c30`. Its main source matches the export
from `185ba767545af567d3d2d3917e807222ff0b2771`.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract and structure | Author, coordinator, student | Student's initial and repair ownership plans accepted against new teaching before affected implementation; implementation reviewed | Preserve the stated owners, bounds and lifetime |
| Implementation and local checks | `/root/coder_ch06` | Initial runtime `aa5f86a` retained; repaired runtime `75bd14d` passes all seven modules vet/test and main race; inherited incompatible grader result 0/100 retained | Preserve both source identities |
| Independent acceptance and mutations | Reviewer, coordinator | Full immutable gate passes at `5a95578`: 23 command groups, 14 contract groups, six CLI barriers, 61 wire cases, retained Chapter 5 assertions, 16 implementation mutants and two CLI recovery mutants; three supplementary mutants pass at `340c678` | Bind exact receipts in checkpoint manifest |
| Initial live use | Student, coordinator | Nine sessions / 33 requests frozen at `f73b01e`, `d12a0cb`, `3417575`; exact live executables archived at `bb74fed`. Independent audit passes 16 controls, chronological single-Agent replay, artifacts/usage and public finals; see `chapter-06-live-review.md` | Preserve initial identities and limitations; plan only affected revised demonstrations |
| Historical comparison and revisions | Independent reviewer, original student | R1–R4 repairs accepted at `75bd14d`, after retained intermediate `3ccaed6` and escaped-opaque accounting correction; final comparative record at `5a95578`. Teaching precedes each affected correction | Preserve comparison rationale and original attempt |
| Final live evidence | Student, coordinator | Accepted: seven sessions / 20 requests frozen at `8c73f9b`, bound to `75bd14d`; 16 independent controls, exact replay, artifacts, usage and public finals pass. All 75 original files unchanged; binaries archived | Bind revised audit receipt; earlier unaffected demonstrations retain initial source |
| Manuscript and teaching feedback | Author, student, proofreader | Accepted at `37a37f7`, with manuscript through `815a84f`; full reading, source-bound figures/excerpts and hard prose checks pass; student teaching review and author dispositions retained | Preserve initial experience and review lessons |
| Export and checkpoint | Coordinator | Exact 1,550-file export from `c3fa758`; all hashes verified, all 76 runtime/testdata files match tested `75bd14d`, credentials absent | Extend this accepted source in a fresh Chapter 7 student context |

The student reads the complete coding skill, architecture and new Chapter 6,
with earlier second-edition contracts and source as needed. First-edition
chapters, historical implementations, later teaching, grader internals and
author/reviewer research are excluded from its initial context. Actual reads
belong in `solutions/edition-2/main/evidence/ch06/`; an instruction boundary
in a shared filesystem is not a claim of operating-system isolation.

The independent grader role previously wrote Chapter 5 checks and review
probes, not that chapter's student implementation. Its historical comparison
begins only after this chapter's initial student attempt and runs are frozen.
Root and the resumed author handle contract questions; the author also prepares
Chapter 7 while the student implements Chapter 6.
The initial independent checker and remaining coverage are documented in
[chapter-06-grader-review.md](chapter-06-grader-review.md).

New Gemini live validation uses the discovered `models/gemini-3.8-flash`
target selected by Bill. All three API adapters require actual human-mode
streaming, tools, interruption and plain-mode demonstrations, with the public
consumer paths in §6.8. Strong deterministic barriers and fault injections
complement those runs. Credentials remain in memory or subprocess environments;
the optional GUI remains an explicitly identified stub.

Local commits and tags are authorized. No push, publication or Bill editorial
approval is implied by this validated checkpoint.

## Checkpoint identities and limits

Initial runtime: `aa5f86a4782c7479ea61b0e6abcbd4163968b574`; reviewed runtime:
`75bd14d5d2424778ebf45cb9025e9dbf314f956a`. Revised live evidence:
`8c73f9b77d9f162079d11ce305b03528c023f3e9`. Export source:
`c3fa7583c5e4c3dcf026b3d5a0a93da000dd8307`; only README and historical checkpoint
labels changed after the evidence freeze. The snapshot is
`solutions/edition-2/ch06/`, with manifest `solutions/edition-2/manifests/ch06-r1.json`.

The final immutable gate at `5a95578` tested all seven modules. The export's
exact runtime/testdata hashes match that gate; no redundant execution on the
identical export is claimed. The [export audit](checkpoint-evidence/ch06-export-checks.json),
[source audit](checkpoint-evidence/ch06-coordinator-source-audit.json),
[final proofread](checkpoint-evidence/ch06-review-final-proofread.json),
[code comparison](chapter-06-code-review.md) and
[live evidence review](chapter-06-live-review.md) bind the remaining checks.

Compared with the historical streaming implementation, the accepted solution
separates framing from API completion, parser facts from actor acceptance, and
provisional observations from reliable completion. Review also fixed recognized
refusal replay, quadratic assembly, operation-wide deadlines and public-client
ownership. Exact escaped-byte bounds survived a separately retained failed
repair. These lessons now appear before the corresponding implementation.

The optional GUI remains a stub. Live evidence is macOS and does not claim
Linux use, exposed thinking or subscriber overflow. Deterministic controls
cover those specified stream/overflow behaviors. Gemini token-limited answers
remain accepted partial output, with no claim that the requested explanation
finished. Earlier unaffected paths retain their initial source identity.
Scoped tests and mutations are evidence, not an exhaustive proof.
