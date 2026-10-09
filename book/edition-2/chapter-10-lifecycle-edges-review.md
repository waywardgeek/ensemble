# Chapter 10 additional lifecycle edge controls

The five scoped lifecycle groups passed on immutable
`57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4` under the Go race detector. All four
targeted mutations reached their intended assertions. This is additional
deterministic coverage, not full Chapter 10 acceptance.

The [final audit receipt](checkpoint-evidence/ch10-lifecycle-edges-audit-57d4aac.json)
binds all 159 runtime/API files and exact checker bytes, preserving each command's
start/end, free disk, output and result. Formatting printed nothing; overlay
`go vet ./...` and the unchanged module's `go test ./... -count=1` passed.
The [verification receipt](checkpoint-evidence/ch10-lifecycle-edges-verification-57d4aac.json)
records six Python orchestration controls and confirms that all eight helper/child
PIDs observed across both attempts are absent after cleanup.

The [first audit](checkpoint-evidence/ch10-lifecycle-edges-first-57d4aac.json)
also passed the five runtime positives, but its overly broad floor-lowering
mutation reset a fresh zero floor before the real job101 parent could be created.
That setup failure received no mutation credit. The corrected mutation leaves
fresh zero alone and wrongly replaces an existing root floor with a nonzero
restored value. It reaches the intended observation: got102 rather than201.
The other intended failures report got1 rather than102, a child-held store lock
after owner death, and a wrapped/reused root handle. Both receipts remain intact.

The public high-ID fixture passed both inspection and import before its actual
maximum-handle job and collision/exhaustion attempts. The lock mutation reached
a changed child heartbeat after owner death, then failed reopening the same lock;
it did not fail at construction or child startup. No real provider, credentials,
physical large-file fixture, student source edit or first-edition comparison was
used. The compiler slot was released after these scoped checks.

## Preserved preparation and scope

October 8, 2026. Independent, new-only preparation after `538650a`; no historical
persistence comparison and no student runtime edit. The complete mandatory
coding skill was reloaded before this stage. Current Chapter 10, its published
API/codec and relevant Chapter 4 ordering rules remain the authority.

At initial preparation these checks were **uncompiled and unrun** while the
student owned the compiler. The eventual command uses the following interface:

```sh
python3 scripts/edition2/accept_ch10_lifecycle_edges.py SOURCE_DIRECTORY \
  --source-commit SOURCE_COMMIT --receipt RECEIPT.json --audit
```

| Group | Genuine parent and required observation |
|---|---|
| Root job allocator floor | Burn100 handles through public allocation, run actual job101, export, then import into roots at0 and200. Their next handles must be102 and201. No private root counter assignment is used. |
| Consumed limits then append failure | Run a real local setter turn, checkpoint pending limits, then attempt an unknown tool. The real log wrapper fails only the ensuing tool_called write, after the consumption fact was persisted. State and offline replay must remain consumed; the older checkpoint stays unchanged, no continuation is sent, the complete log prefix is inspectable but unfinished, and live resume refuses. A no-fault control exercises the identical path first. |
| Partial-store topology | Accepted full-origin and imported stores are the parents of checkpoint-without-log, origin-without-log, anchor-without-origin and full-origin-with-extra-origin cases. Live/offline refusal preserves every source byte. Restoring only the fixture change permits reopening, proving reservation release. This is topology evidence, not injected origin-write/sync/close failure evidence. |
| Surviving tool child and store lock | An independent public-consumer process runs an actual shell tool, pairs its initial report, finishes the turn and checkpoints. The shell ignores HUP and updates an owned heartbeat. Killing only the consumer must leave a changing child heartbeat while a new consumer acquires the same lock inode. |
| Job uint64 and final artifact collision | A genuine completed job101 export is rebased consistently to synthetic historical MaxUint64-1, including every typed job reference, locator, report text and both exact watermarks. Public inspection and import must accept it before use. A real job then receives MaxUint64; further allocation cannot wrap. A separately occupied final artifact forces collision skipping into exhaustion without truncation or execution. |

The high-ID fixture is explicitly synthetic history derived from a real valid
parent. It never claims that the earlier command actually ran at the rebased
handle. Unlike an Agent-local activation cursor, the root job allocator can burn
IDs belonging to other Agents; a represented high historical job is compatible
with a small session. A watermark-only alteration is not used as a valid parent.

Chapter 4 §4.2 distinguishes ordinary tool validation refusals from job creation
failures. Its latter rule says to terminate the current request without starting
the handler or continuing to another model request. The job exhaustion fixture
applies this pre-admission creation-failure rule, not a new rule that every tool
error is terminal. It does not impose atomicity on a complete turn or multi-call
batch. Chapter 9 defines an atomic activation group; Chapter 10's generic overflow
sentence does not create an additional event/job batch transaction.

The descriptor fixture binds the shell PID to a unique command token through
`ps`, records holder/child PIDs, token, heartbeat changes and lock inode, and only
cleans a still-matching owned child. It also handles failures before the holder
ready barrier, so an early test failure does not leave an ignored-HUP child
running. A PID-exists probe alone cannot pass as evidence of a surviving child:
a zombie could satisfy that probe without retaining any descriptor.

The prospective mutations remove root-floor advancement, wrongly lower an
already higher floor, permit root handle wrap, and explicitly clear close-on-exec
on the actual lock descriptor. Merely deleting the explicit CloseOnExec call is
not an adequate inheritance mutation because Go's original file open already
uses that flag. The last mutation must pass child startup and positive ordering,
then fail specifically because the surviving child retains the lock. Compiler,
constructor, cleanup or source-binding failures receive no mutation credit.

The runner also receives a fixture-owned process ledger so a Go-test timeout
cannot bypass the shell cleanup. Before signaling, it independently checks the
recorded PID's current command against the unique child token or exact helper
executable/invocation. Unrelated PID reuse is never signaled. Two mocked cleanup
controls verify both the positive matching case and the unrelated-process case.

The runner requires each intended mutation's specific assertion text, as well
as its named failing test. A constructor/child-startup failure, timeout or
compiler failure does not receive deletion credit. A sixth mocked orchestration
control exercises these distinctions before any runtime check.

The runner preserves all frozen runtime identities, adds only disposable Go
overlays, saves started/completed commands atomically and refuses audits before
the full positive subset passes. Six receipt/cleanup-only orchestration tests and
formatting checks were the only results at preparation. The later execution
results above have their own source identity; earlier preparation receipts remain
bound to `8882a18` and are not relabeled as tests of the final runtime.

Exact canonical-state/checkpoint/origin limits, physical log-file proof,
allocation instrumentation, controlled Skills preflight versus storage admission,
and origin-creation fault injection remain separate work. Event uint64 exhaustion
also requires a genuinely allowed boundary: newly initialized contiguous sessions
reach the one-million event-count cap long before uint64 exhaustion. No sparse
live origin or invented grouped atomicity is introduced to manufacture that test.
