# Chapter 13 validation

Speech-channel chapter preparation, October 8, 2026. New Chapter 13 maps to
first-edition Chapter 14. Its listener exercise must establish which channel
supplied an answer; transcript delivery and native audio remain separate facts.
Bill's editorial approval is separate from technical validation.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Research and outline | Author `/root/coder_ch04` | Outline/evidence `bb6e04c` and full draft `e4ef9a9`; historical listener, fence and missing-event wiring lessons retained with limits | Reconcile future student experience without inventing outcomes |
| Contract and design review | Coordinator, independent proofreader `/root/coder_ch08` | Grouped revision `f3a50d3` accepted for checker preparation; independent closure `88de827` resolves both findings, with full coordinator reading | Derive checks from exact grammar, public boundaries and failure semantics |
| Independent checks | Grader engineer, unassigned | Not started | Derive controls from the published contract and actual public seams |
| Student implementation | Fresh student, unassigned | Not started; accepted Chapter 12 implementation required | Freeze permitted new teaching and predecessor source, then review owner plan |
| Initial live use | Student, coordinator | Not started | Review bounded all-provider CLI/browser/public/listener matrix before calls |
| Historical comparison | Independent code reviewer | Not started | Preserve initial source, runs and teaching review before comparison |
| Manuscript and feedback | Author, student, proofreader | Full contract draft independently proofread; no successful demonstration claimed | Reconcile actual evidence and resolve future reviewer/student findings |
| Export and checkpoint | Coordinator | Not started | Complete required gates before export and immutable tag |

## Working directions

These are coordinator choices under the authorized architecture, not additional
decisions attributed to Bill. Revision `f3a50d3` now publishes the reviewed exact
limits, literal fixtures and failure behavior.

- Keep Page normalization and queue ownership, application-owned SpeechService,
  native arbitration, captured preferences and inherited pause behavior. Bound
  pending ambiguous normalization text. Overflow visibly refuses that part's
  unqueued remainder without manufacturing words, dumping code or canceling
  unrelated admitted speech. Specify recovery and chunk-invariant behavior.
- Let SpeechService own a bounded journal with scoped sequence cursors and
  explicit eviction gaps. Closed-Page records survive until ordinary eviction.
  Prefer explicit bounded export through the listener/public seam, with the
  external harness owning its diagnostic file. Saving failure invalidates that
  evidence run without silently disabling native speech or growing memory.
- Construct a separate restricted listener with frozen rights and explicit
  target scope. Preserve Chapter 12's exact five-tool endpoint. The fast
  exercise consumes recorder-completion text; native admission, start, end,
  cancellation, failure and captured audio remain separately identified.
  Denial controls must prevent a planted answer arriving through another route.
- Add only a narrow, autoplay-gated announcement for a safe terminal turn or
  transport failure, deduplicated by owned request/generation. Ordinary tool
  results remain silent, including failed results. Replay, reconnect and
  intentional cancellation do not invent failure speech; synthesis failure
  cannot recursively announce itself. Publish exact safe text and selection.

Historical accounts motivate these decisions. They are not new native-audio
receipts or claims that a transcript-only model perceived sound. No new general
accessibility framework, sub-agent runtime or speech-engine portability layer
is required by these directions.

## Contract review closure

Independent review [chapter-13-review.md](chapter-13-review.md) closes the
journal-exhaustion and demonstration-ownership findings. Recording faults latch
bounded out-of-band status, preserve retained records and native work, and
explicitly fail recorder delivery and listener/export completeness. A fault
does not require another sequence number or a fabricated terminal record.
The human launcher and public listener example create distinct applications
and retain separate launch/service identities.

The coordinator read the complete revised contract. Three JSON fixtures, the
maximum uint64 read shape, seven unchanged normalization pairs and the `Ready.`
digest are independently checked. Initial prose lint remains bound to the
original draft; the grouped revision has manual voice and exact-literal review.
These establish a reviewed teaching contract, not runtime or live acceptance.
Chapter 12's accepted implementation and a runnable new checker remain required
before the fresh Chapter 13 student starts.
