# Chapter 05 author carryover

## Delivered

- Full new manuscript: `book/companion/chapter-05.md`, “Let the Loop Listen.”
- Only that manuscript and this note were written. No Git operations, code inspection, builds, graders, live calls, or linked archive reads.

## Approved context fully read

- `book/companion/voice.md` in full; followed the current approved-context restriction rather than its links to older process documents.
- `book/edition-3/chapter-05.md` in full, including exercises, deferred work, demonstrations, and final evidence/review discussion. Initial clipped output was recovered with consecutive bounded reads: lines 1–220, 221–440, 441–590, 591–740, 741–890, 891–1040, 1041–1190, and 1191 through EOF. Extra reads beyond EOF were empty.
- Accepted `book/companion/chapter-04.md` in full; recovered clipping with contiguous reads covering 1–60, 61–115, 116–155, 156–220, and 221 through EOF.
- `docs/companion-notes/chapter-04-author.md` in full. This and the preceding manuscript supplied continuity only, not authority for Chapter 05's exact semantics.

## Editorial decisions and continuity

- Opens at Chapter 04's remaining foreground-deaf wait: the tool already runs on its own goroutine, while the conversation loop parks awaiting the job report and cannot drain inbound events. The initial unsupported narrator anecdote was removed after review.
- Main causal spine: one Agent-owned conversation loop; workers return facts through one reliable mailbox; hints must wake observation and reach the next model request; observers report committed facts without becoming execution owners.
- Gives substantive space to batch closure, exact call pairing, undispatched tails, one-shot limit consumption, timer retirement, and the PTY echo distinction. These are correctness consequences of wakeup, not a recap of grading fixtures.
- Separates hint wakeup, turn interrupt, job kill, and shutdown; distinguishes request-scoped completion from lossy progress and synchronous API from duplicate orchestration.
- Weaves essential exercise knowledge into the peer author/editor/reviewer application: application tools and actual stored artifacts, synchronized shared draft, reliable sequential completion, separate histories, interruption/EOF behavior, and the limit of hardcoded Go coordination. Omits harness spellings and setup commands from prose.
- Includes snapshot aliasing, post-reduction notifications, part identity for delta/finalizer correlation, non-streaming whole-part deltas, observer backpressure, media refusal and actual byte delivery. Does not claim streaming, URI/handle attachment rendering, spawning, or grouped completion exists here.
- Historical evidence is selectively attributed to the source: Bill's stuck-parent report is not treated as a reconstructed wire trace; CLI and workflow observations are historical; media live versus deterministic coverage stays distinct. Omits scores, token budgets, and nonessential exact timings/model names.
- Closing boundary: the loop can listen and applications can compose peer Agents; model-directed collaboration remains later work.

## Checks and limitations

- Verified the entire approved source by recovering clipped reads before drafting. Reread the entire saved manuscript in three bounded reads (lines 1–70, 71–140, and 141 through EOF), checking ownership boundaries, hint/interruption distinctions, batch and timer semantics, historical attribution, and deferred scope. The initial author read missed the three source-fidelity issues corrected after editor/reviewer feedback: the unsupported narrator anecdote, a sequence field incorrectly attributed to `StateChanged`, and a persisted-history part-ID seeding requirement not established in Chapter 05. Rechecked the source's TL;DR and §§5.3–5.4 plus targeted part-ID references; removed those claims rather than recasting them as requirements.
- Reread the corrected opening, observation section, and updated note after saving. Correction scope was limited to the three verified findings above; the other historical stories and implementation rules were left intact.
- No lint invoked; no current execution, testing, or provider behavior claimed. Historical claims depend on the approved source, not independently opened evidence archives.
- Editor and independent reviewer should read the entire chapter. Author remains available for material corrections.
