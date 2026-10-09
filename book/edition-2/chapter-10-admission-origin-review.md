# Chapter 10 bounded admission and origin fault review

Independent reviewer controls against immutable runtime
`57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4`; all 159 source/API/asset
identities are checked before and after execution. No student runtime file was
edited. These are four focused test groups covering three remaining promises,
not a complete Chapter 10 acceptance claim.

## Results

| Promise | Genuine positive and refusal | Targeted mutation |
|---|---|---|
| Bounded record reads | An actual fresh session initializer padded to exactly 64 MiB passes the real streaming reducer. 64 MiB + 1 and a logical 8 GiB record refuse before reduction. The latter generates only 67,112,960 bytes. Measured incremental allocation is 134,221,808 and 134,221,664 bytes respectively. | Replacing bounded `ReadSlice` with `ReadBytes` reaches the protective read fence and fails the intended assertion. |
| Bounded write preparation | A small actual session event prepares successfully. Caller-owned 64 MiB and 256 MiB bodies exceed the encoded record cap; actual preparation refuses both with session_limit, allocating 4,768 additional bytes in each case. | Replacing bounded encoding with json.Marshal accepts the first oversized record and allocates 805,387,008 bytes, then fails the intended assertion. |
| Skills refusal versus terminal storage | A valid 256-definition catalog produces a transition within decoded material limits whose complete escaped event is oversized. Public LoadSkill returns skill_too_large without state/log/counter changes; the next small load succeeds with activation 2/revision 1. A subsequent oversized ordinary public append returns session_limit, retains prior state/log/checkpoint and ends further admission. | Removing controlled error translation and suppressing terminal storage faulting each fail their separate intended assertions. |
| Immutable origin creation | A real exported checkpoint passes public Inspect/Import with the real exclusive descriptor. Short write, write error, reported sync error and reported close error each return session_io without exposing an Agent or writing later log/checkpoint files. Actual partial origin bytes remain; later open/inspect refuses unchanged, while the same SessionID imports elsewhere in the same root. | Four separate omissions of the short-write, write-error, sync and close checks each expose a live Agent and fail the intended assertion. |

Allocation is measured with Go TotalAlloc, after caller input construction and
GC. The read comparison allows 1 MiB of runtime noise between the short overflow
and long unread tail; the write comparison uses the same allowance. These are
component measurements, not a universal heap/RSS ceiling or a proof about every
snapshot collection. The 8 GiB tail is generated lazily, not stored as a file.
A reader fence protects the review machine from the deliberately unbounded
mutation. Existing physical-file checks retain their separate scope.

The Skills fixture uses 128 offers, 126 dependencies, a primary and a loadable
root. Its 127 newly activated bodies are each 64 KiB after binding expansion;
repeated offer descriptions and escaping enlarge the complete event. Installed
management handlers are real. The terminal branch is the explicitly permitted
ordinary-record storage overflow, not fabricated log/count exhaustion. No model
call or attempted-call limit-consumption claim is added here.

Origin instrumentation replaces only the origin descriptor constructor in an
overlay. The wrapper retains its Store parent and actual os.File, records
open/write/sync/close operations, and scopes injected faults to owned test paths.
Short writes physically write half the export. Sync/close faults are reported
after the real operation, so this proves handling of reported failures, not
hardware failure or power-loss durability. The passing import runs before all
fault cases. Mutations affect only WriteOrigin, not checkpoint writing.

## Receipts and retained failures

- [First run](checkpoint-evidence/ch10-admission-origin-first-57d4aac.json):
  missing management handlers in the fixture prevented Skills construction.
  Other positive groups ran; no mutations were credited. The exact initial
  [Skills fixture](checkpoint-evidence/ch10-admission-skills-initial.go.fixture)
  is retained.
- [Handler correction](checkpoint-evidence/ch10-admission-origin-handlers-57d4aac.json):
  original three groups and seven mutations pass. The exact earlier
  [driver](checkpoint-evidence/ch10-admission-before-write-probe.py.fixture) and
  [read-only allocation fixture](checkpoint-evidence/ch10-admission-read-only.go.fixture)
  preserve the identities preceding the additional write probe.
- [Final read/write audit](checkpoint-evidence/ch10-admission-origin-read-write-57d4aac.json):
  four positive groups, all eight intended mutation failures, empty gofmt output,
  overlay module vet and unoverlaid full-module tests pass. Each command records
  stdout/stderr, time, completion, disk capacity and the expected mutation assertion.
  No compiler or source refusal earns mutation credit.
- [First focused race/checker run](checkpoint-evidence/ch10-admission-origin-final-checks-57d4aac.json):
  the unresolved /var alias failed to overlay actual /private/var package paths;
  Go ran no matching tests. Its zero exit is explicitly marked failed coverage.
  Three Python receipt controls passed (identity refusal preserves evidence,
  interruption retains command status, failed positives prevent mutations).
- [Resolved-path focused race](checkpoint-evidence/ch10-admission-origin-race-resolved-57d4aac.json)
  separately requires both Skills and origin test names to pass; allocation
  measurements deliberately remain outside race instrumentation.

All large inputs here are generated in memory; owned temporary workspaces are
removed after each run. No provider, credential, retained binary or user file was
used or deleted. Bill subsequently freed disk space; the earlier physical 1 GiB
log barrier is historical, and that separately authorized stage is tracked in
its own evidence. Real-model demonstrations and the independent historical
comparison remain outside these deterministic controls.
