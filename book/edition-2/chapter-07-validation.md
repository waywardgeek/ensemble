# Chapter 7 validation

Scoped streaming maintenance is accepted for `edition-2-ch07-r3`. Runtime
`55ee774` and evidence/export source `662c09c` carry the independently reviewed
Chapter 6 zero-byte argument correction into accepted Chapter 7 r2. The exact
export has 2,239 files. Original acceptance and live receipts below retain their
r2 identities; this is later-informed maintenance, not another cold student.
See [the independent maintenance review](empty-arguments-backport-review.md).

Chapter 7 is accepted for corrected checkpoint `edition-2-ch07-r2`. Fresh new-only student
`/root/coder_ch07` completed the exercise. Its accepted predecessor is
`edition-2-ch06-r1` at `1c6b1f065d11bd3a532c94bc53305394e17e7cc2`. The student's
ownership plan was accepted before affected implementation. Bill's editorial
approval is separate.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract | Author, coordinator | Accepted at `74928f1`; incremental projection lesson `2f105ce`; published initial checker command `7b116b7` | Preserve the new teaching in the fresh student handoff |
| Structure and implementation | `/root/coder_ch07`, coordinator | Initial runtime `da162e8`, support `ba902b7`, initial experience/evidence `8cc87f2`; grouped revised runtime frozen at `9ba7855b31a5eb134819602b35eb9acb277f6342` | Preserve both identities through final evidence |
| Independent checks | Grader engineer, coordinator | Revised full 33/33 gate and grouped comparative 12/12 suite pass against `9ba7855`; final clearance `17aa661`; initial failures and unchanged historical suite retained | Bind receipts into final manifest |
| Initial live use | Student, coordinator | Nine real sessions on all three providers; 44 requests exactly reconstructed; root audit `950cead` passes 24 controls, binds 386 original files and verifies four audio captures | Preserve initial source scope; revised browser behavior needs separate evidence |
| Historical comparison | Independent reviewer, student | Four grouped repairs accepted at `691c2a4` / `17aa661`; no remaining material code finding | Preserve comparison and original attempts in checkpoint |
| Revisions and final evidence | Student, coordinator | Revised evidence `341f15d`; six exact request replays, six PID-bound WAVs, 145 unchanged session files and ten independent controls pass; see `chapter-07-live-review.md` | Bind accepted original and revised evidence with their separate scopes |
| Manuscript and feedback | Temporary author `/root/coder_ch04`, student, independent proofreader | Final actual-use reconciliation `00cf641`, neutral opener `d7e097e`, feedback correction `2112af1`, student confirmation `4897642`; independent final proofread `121b65d` accepted | Preserve resolved teaching exchange and evidence limits |
| Export and checkpoint | Coordinator | Corrected exact 2,229-file export from `9ec94ef`; manifest `solutions/edition-2/manifests/ch07-r2.json`; actual delivered-tree checks pass | Immutable `edition-2-ch07-r2`; r1 remains historical |

The student receives a fresh context, new-edition teaching and the accepted
Chapter 6 source in `solutions/edition-2/main/`. It must read the entire
`book/edition-2/skills/ensemble-coding/SKILL.md` before code and after compaction.
First-edition chapters/solutions, future solutions and author/reviewer research
are outside the initial context. Historical links in the skill do not override
that boundary. Actual reads and help belong in the student's retained review.

The GUI remains a separate optional module. Public watch and pause authority
belong to Agent; browser connections and components use public owner interfaces.
The combined application reuses the CLI and the same Ensemble/Agent. Real
browser synthesis requires observable evidence of produced audio; mocked queue
tests and available voice names alone cannot close that gate.

The grouped revision also corrects shutdown of the browser application itself:
closing all views must stop speech admission before closing individual Pages,
so cancellation cannot start a peer's queued utterance during root teardown.
The independent speech control and its intended deletion protect this case.
This is follow-through on shared-speech finding R3, not an extra feature.

All eight source modules pass vet/tests through the immutable full gate. The
coordinator independently matched its 101 non-evidence files to the export
source: only README and the historical checkpoint document changed after the
runtime freeze. Later student feedback and live evidence retain their own
commits. No runtime or embedded asset changed after final deterministic/live
validation. Export verification compares every file byte and mode to its
committed source, so no additional runtime test is inferred from copying it.

### Packaging correction after the first export

The coordinator and fresh Chapter 8 student independently found that the full
r1 tree's `go list ./...` / `go vet ./...` includes a standalone evidence helper
that imports the optional GUI module. The runtime gate had excluded evidence
from its test copy. Its passing results remain valid for that copy, but did
not establish ordinary testability of the complete delivered tree. The first
checkpoint at `ed5667b` was premature on that packaging property.

Correction `9ec94ef` adds a conventional `//go:build ignore` constraint and
explanatory comment to that one helper. Explicit-file compilation from the GUI
module still succeeds; no executable body, runtime, live support binding or raw
receipt changes. All eight actual main modules pass package discovery, and the
actual core passes vet and full tests. The independent reviewer checked the
exact four added lines and all 13 recorded outcomes, including the preserved
r1 failure, before r2 export. See `checkpoint-evidence/ch07-packaging-r2-checks.json`.
The procedure now requires full delivered-tree package discovery before tagging.

The original `edition-2-ch07-r1` tag is unchanged. Revision r2 replaces the
tracked export through a new exact-source export, with its own manifest and
immutable tag. Chapter 8 starts from this corrected baseline. Original paid
receipts still describe the same runtime; this packaging correction requires
no repeated model calls.

Accepted live scope comprises the original nine sessions/44 requests and the
three revised two-Agent sessions/six requests. Initial and revised receipt
identities remain separate. Gemini's revised short MAX_TOKENS answers, the
unsuccessful first Anthropic queue timing, the reviewer capture-timing correction
and all historical failed attempts remain visible. Audio establishes produced
signal from the selected process, without claiming transcription or Bill use.

Author commit `49ae919` captured the student's concurrently staged revision
paths through the shared Git index. Both agents disclosed the race. The student
reviewed only its owned files and committed the remaining teardown repair at
`9ba7855`; no future chapter text was supplied to the student. History remains
intact, and validation names the actual frozen revision. The procedure now
requires explicit owned pathspecs on commits as well as staging.

The initial checker invocation is:

```sh
python3 scripts/edition2/accept_ch07.py GUI_BINARY
```

Its scope is partial. The complete required coverage remains §7.9 and
the [grader plan](chapter-07-grader-review.md). Retained Chapter 6 assertions
are also required, with fixture placement adapted to the public CLI package.
Local results above do not imply completed live use, successful chapter
acceptance, push or publication.

## Handoff and scheduling

Root supplied committed-only copies of new Chapters 1–7 at
`/Users/bill/projects/ensemble-edition-2-revisions/ch07-student-inputs/`, bound
to the Chapter 6 checkpoint by their manifest. Concurrent external edits to
earlier manuscript files are preserved and excluded from this handoff. The
mandatory repository skill and architecture remain direct required reads.

The original author is preparing Chapter 8 while the fresh student prepares its
Chapter 7 plan. Resuming `/root/grader_ch05` for full Chapter 7 engineering hit
the active thread limit. Its initial checker/coverage plan remain available;
resume that independent role when the author's next checkpoint frees a slot.
Root accepted the new student plan before implementation. After the author
checkpointed Chapter 8 at `cfb0d87`, the independent grader resumed Chapter 7
engineering. No Chapter 7 implementation acceptance is implied by its initial
partial checker.

At the Chapter 7 source freeze, the original author still could not resume at
the thread limit. The coordinator reused the available completed Chapter 4
student thread as a temporary author. It has Chapter 4 implementation exposure,
has not authored Chapter 7 code, and must read the complete current voice and
chapter procedure. Its scope is Chapter 7 prose/outline/evidence/feedback; the
independent grader retains later code comparison and proofreading. The role
change avoids blocking prose preparation on the original thread's availability.

## Early implementation invariants

During the new-contract ownership check, root accepted the public reusable CLI
client outside the internal spokes, importing only the public core and standard
library. The grader will adapt inherited fixture placement while preserving its
assertions; an old `cmd` directory layout is not an architectural requirement.

Root also asked the student to ensure idle watch close releases its actor-owned
recipient registration without requiring a later publication, and that an
overflowed partial projection cannot produce a falsely complete new snapshot.
These are work-in-progress lifetime/completeness findings, before historical
comparison or initial source acceptance; outcomes belong in the student review.

The grader clarified that pause publication precedes the public update return,
but the contract does not require a same-socket observation frame before its
acknowledgement. Wire checks require the matching applied revision/counts and
ordered watch delivery, without inventing a stronger interleaving requirement.

The source-bound live plan is `main/evidence/ch07/live-plan.md`. It covers all
three providers through browser, actual human PTY, shared-Agent CLI, plain
delivery and public two-Agent component reuse. Exact timing/failure boundaries
retain deterministic controls. Native speech start/end callbacks were observed
locally, and macOS screen/audio capture preflight returned true without a
permission change. The subsequent isolated recording selects the launched
Chrome process by PID. Its first 1.8 seconds are silent; the utterance's
time-aligned interval contains non-silent stereo audio. The raw receipt and WAV
are retained with `main/evidence/ch07/audio-isolated/audit.json`. This is a
local fixture feasibility result, without a transcription or human-listening
claim; each required paid browser demonstration still needs its own receipts.

The coordinator's actual Connector probes use disposable source copies and
real sockets. They distinguish 256 accepted outgoing items from the 257th,
exact 128 MiB encoded capacity from aggregate/single-item excess, and verify
drain accounting, selected-sender teardown, independent peer usability and
pause release. A deliberately blocked transport honors the production write
deadline; the writer closes and joins within its five-second bound. A peer
that never dispatches pongs closes within the 30-second contract. All six
positive groups and seven intended deletion failures pass against `da162e8`;
98 non-evidence source files match its Git blobs. The receipt is
`checkpoint-evidence/ch07-connector-mutations.json`; the injected fixture also
passes vet. A nonunique mutation anchor interrupted the first audit and is
retained separately as a reviewer setup failure. The complete combined gate
remains pending.

## Revision 3: propagated Messages correction

Only the original parser changes. The five-line fix and added regression/SSE
fixture exactly match reviewed Chapter 6 source `788c5e9`; 2,128 existing evidence
files stay unchanged. Root review `8812461` verifies 103 source hashes, seven
receipt hashes, 147 checker inputs and the archived executable. All 24 affected
command groups and four local commands pass, including core race and eight-module
vet/tests, early streaming and complete message barriers, watch/pause behavior,
prior assertions and CLI overflow deletions. The original three regression
failures are retained alongside 13 corrected passes. No new full-gate or paid
browser run is claimed; unaffected accepted evidence retains its source identity.

The coordinator verifies all 2,239 export hashes and complete eight-module test
package discovery. The manifest `solutions/edition-2/manifests/ch07-r3.json`
binds these receipts and original r2 acceptance. This immutable tag belongs to
the isolated Chapter 7 branch with matching canonical main and ch07 trees;
integrating its snapshot must not replace current Chapter 9 main. The original
SSE frame delimiter is retained despite its blank-line whitespace warning.
Chapter 8 propagation remains pending. Bill's editorial approval and publication
remain separate.
