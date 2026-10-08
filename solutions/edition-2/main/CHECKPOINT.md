# Source checkpoint records

The current implementation is Chapter 7. Reviewed runtime/support identity:
`9ba7855b31a5eb134819602b35eb9acb277f6342`. The independent immutable gate
passed 33/33 and comparative repair suite passed 12/12, including distinguishing
deletion controls. Initial runtime `da162e8`, support binding `ba902b7` and
initial live/student freeze `8cc87f2` remain preserved. The reviewed batch fixes
killed-job snapshot replay, DOM disposal/remount, shared native speech ownership
and explanatory refusal of larger/fragmented oversized messages.

Revised launch binding and receipts live in `evidence/ch07/revision-1/`:
three public two-Agent browser sessions, two prompts and two HTTP requests per
provider, six exact request reconstructions, and two captured audio samples per
provider. No retries. Gemini returned short `MAX_TOKENS` answers under the
example's 512-token budget; retained audio and UI checks use that actual text.
Initial full-feature browser/CLI/embedding runs retain their 44 requests and
original source identity. This documentation update changes no runtime.

The [Chapter 7 validation record](../../../book/edition-2/chapter-07-validation.md)
tracks final coordinator acceptance, export and tag. No final tag is asserted
by this student checkpoint. No historical answer or grader implementation was
read for the Chapter 7 revision; reviewer rationale and new teaching guided it.

## Historical Chapter 6 checkpoint

At this earlier checkpoint, the implementation was Chapter 6. Its repaired runtime is frozen at
`75bd14d5d2424778ebf45cb9025e9dbf314f956a`; revised launch bindings are in
`8946020` and revised live/student receipts in `8c73f9b`. Changes to this file
and README are documentation only. Current acceptance, export and immutable tag
status belong in the
[Chapter 6 validation record](../../../book/edition-2/chapter-06-validation.md).
This record does not assert that a final Chapter 6 tag already exists.

Chapter 6 streams all three API adapters by default, exposes plain delivery
through `EN_DISABLE_STREAMING=1` / `Config.DisableStreaming`, and adds optional
`protocol --observe` records while preserving reliable request completions.
The new public `examples/stream-consumer` exercises two Agents with ordinary,
finals-only and stalled subscribers. The root, optional GUI and five public
example modules are independently checked; see README for commands. The GUI
remains a transport stub.

The initial runtime `aa5f86a` and its nine live sessions / 33 requests remain
preserved with receipts `f73b01e`, `d12a0cb` and `3417575`. Intermediate repair
`3ccaed6` is retained before the escaped-opaque size correction. The repaired
runtime's seven affected-path sessions / 20 requests preserve actual tool reads,
interruption/recovery, plain Gemini delivery and public finals. Gemini public
MAX_TOKENS responses remain accepted partial answers. Each receipt retains its
own immutable source and executable identity under `evidence/ch06/`.

## Historical Chapter 5 checkpoint

Chapter 5 runtime was reviewed at `959c663` and live evidence frozen at
`469730f`. Its accepted Chapter 5 tag is `edition-2-ch05-r1` at `7a6ef03`;
that checkpoint's main source matches the export from `185ba767`.
The [Chapter 5 validation record](../../../book/edition-2/chapter-05-validation.md)
and `evidence/ch05/` retain its acceptance and live history. The dated Chapter 2
record below is also historical; its next-action instructions apply to that
earlier checkpoint.

## Historical Chapter 2 reviewed student checkpoint — 2026-10-07

Code checkpoint validated: reviewer accepted the comparison revisions; all module
checks, inherited grader and independent offline acceptance/audit pass. Manuscript
and Bill's editorial approval are separate. The initial student implementation
`39a92ca27a418712832ac0dcbbfbe4e32b3bca35` remains preserved, derived from Chapter 1
`75542c1`.

All three actual CLI backends and all three external-consumer demonstrations
succeeded. The public consumer exercises actual typed calls, supplied results,
redaction/reference replay, observations, unsubscription, independent Agents and
per-model usage after changing configuration. No tool execution is claimed. The
GUI remains an optional separate-module stub with a public integration test;
there is no browser/WebSocket transport claim.

The post-run comparison led to two improvements: internal admission/sequence
checks no longer clone growing history; malformed logs report useful static
categories and line numbers without echoing private data. Public ownership
boundaries remain intact. WHY comments explain consumption and publication order.
No paid calls were repeated for these internal/diagnostic changes. Earlier owner
helper omissions, loaded-config aliasing, and Gemini schema mapping were fixed
before the initial checkpoint; their discovery and receipts remain recorded.

Validation: gofmt printed no paths; vet and tests passed in the main, GUI and
external-consumer modules. Corrected inherited grader100/100. Independent revised
offline CLI acceptance44/44; passing mutation control plus ten exact detected
defects. Audit source hashes match the committed revision files. Coordinator full
root regression passed exit0 (internal/grade513.863s); root vet/formatting were
reported clean. Exact reports and source hashes are under evidence/ch02/.

The original inherited95/100 result remains preserved. Its round-trip fixture
tried to render unresolved calls; root corrected the fixture by explicitly
completing only unanswered calls in a separate input while retaining dumped data.
The student did not weaken the required refusal, read grader implementation, or
copy a first-edition answer.

SOURCE-EXPOSURE.md records a material process limit: there are no known direct
old-answer reads, but inherited historical summaries and compaction mean this
run cannot certify strictly blind regeneration. Implementation validation and
that context-isolation claim are distinct. No historical skill links were followed.

No credentials are stored here; all working-tree files passed an in-memory scan
against the actual three keys. No running student processes or pending approvals.

Next: root/author complete manuscript review. Do not begin Chapter 3 in this
context. The next cold chapter uses a fresh coder with no inherited turns and
explicit new-edition inputs. Reload the full mandatory skill before any new fix.
