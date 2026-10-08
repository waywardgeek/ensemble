# Chapter 8 independent live-evidence review

Accepted on 2026-10-08 for the scopes below. This closes the independent live
receipt audit, not final manuscript proofreading or the coordinator's export/tag.
The reviewer engineered the independent checks and reviewed the implementation;
the reviewer did not author the student runtime, drive these paid sessions, or
repeat provider calls. Source exposure and the historical comparison remain in
[the code review](chapter-08-code-review.md).

## Identity and reproducibility

The audited student evidence freeze is `7f517d8`; the initial student/run/review
freeze `5f9684b` is preserved. Actual launches retain their `bd5c05a`, `cd9de3e`,
`a06d4f3` and support-only `52c9561` identities. Reviewed final runtime is
`a06d4f3c848685d81306166279df7c977d71bddd`; no earlier receipt is relabeled as a
run of that binary. Later coordinator README/CHECKPOINT edits at `40d389b` are
documentation only and were read separately.

Reproduce the local audit from the repository root:

```sh
python3 scripts/edition2/ch08-review-evidence.py
```

[The bound result](checkpoint-evidence/ch08-live-independent.json) records the
script hash, four binding hashes, all 1,067 committed Chapter 8 evidence files,
22 launch records, per-run outcomes/usage, audio and provider totals. All 209
initial-freeze files beneath those run directories remain byte-for-byte equal
to their original Git blobs. All frozen evidence is checked again after the
audit. The exact-key exclusion scan additionally covers 1,097 current evidence
files, including untracked/generated support, with zero matches; only the three
needed credential values were read in memory, and no value was printed.

Before any derived replay write, the audit validates every historical source
set/hash, each executable, exact historical and loaded support, complete browser
dependency identities, and all launches. Historical support is loaded from its
own revision into a temporary directory with an explicit repository root.
Executables are read from their recorded durable external paths, and all hashes
match. Root's separate complete delivered-tree check covers 3,308 files and nine
modules at `7f517d8`; see
[ch08-coordinator-delivered-tree.json](checkpoint-evidence/ch08-coordinator-delivered-tree.json).
This supplements the earlier runtime-focused gate rather than replacing it.

The audit has two positive controls and 24 intended negative controls. A valid
copied-path fixture passes before one identity is changed at a time. Refusals
cover source hash/set/empty maps, the executable set and all ten executable
roles, support hash/set, browser dependency hash/set, the last launch's source,
binary/support/browser/role identities, and a last-run request-body mismatch.
Each produces its specific refusal before creating derived output. Raw originals
remain unchanged. An initial audit implementation treated a headless run as if
it had browser records and raised `StopIteration`; the audit's empty-browser
handling was corrected. No student evidence or behavior changed for that fix.

## Actual provider scope

| Provider and selected model | Admitted prompts | Forwarded HTTP | Scope |
|---|---:|---:|---|
| Messages, `claude-sonnet-4-6` | 11 | 18 | Human CLI, two-Agent headless/public browser, settings and corrected active interruption |
| Chat Completions, `gpt-4.1-mini-2025-04-14` | 9 | 14 | Same public/CLI/browser scope; interruption succeeds in the embedded second Agent |
| generateContent, `models/gemini-3.8-flash` | 10 | 15 | Same scope; bounded correction interrupts a held tool proposal |

All **47 raw request JSON objects** reconstruct exactly through the recorded
CLI and event prefix. Nine single-Agent request-bearing runs also match in
chronological order; multi-Agent sessions use exact multiset comparison because
cross-Agent HTTP order is not a conversation order. The phases contain 21
requests in eight initial launches, 25 in seven native-repair launches, and one
in four final-runtime production launches. Three additional launches replay
retained events without contacting a provider. The 22 launches include empty
or failed demonstrations and settings-only restarts; this is not a claim of
22 successful model sessions.

The original limit was ten prompts and 24 HTTP requests per provider. The
reviewer accepted one additional Anthropic text-only prompt, at most one HTTP
request, after two public-consumer prompts incorrectly requested unconfigured
tools. Its fresh, separately bound launch observes nonempty provisional text
before an accepted interrupt. The final 11/18 total respects that amendment.
Gemini's corrective tenth prompt stays within its original cap. No unbounded
retry or extra provider call was made for numeric settings, replay verification
or this independent audit.

The raw human terminals were read in full. Each shows an initial file read and
continuation, `/usage` and `/history`, then a browser-applied limit of one,
accepted final tool report and `round_limit` cleanup with exit 1. Each public
headless consumer reports separate policies of one and two, one versus two
requests, and fresh-log restarts restoring both policy files. Browser timelines
show policy two applied while the first turn remains governed by one; the next
turn captures two. Restoring zero captures the effective default 16 and permits
continuation. The independent receipt checks these actual event boundaries.

Each final native browser task delivers its hint once, reaches a real file read
and write, and reconnects during work and after completion. Keyboard/pointer
width controls, theme/font settings and viewport changes have retained browser
receipts. Anthropic initially writes the correction including `port=9090`, then
issues another write removing that line. The final file contains only the marker.
OpenAI and Gemini retain the corrected line. This is a documented model outcome,
not evidence that the hint was lost or permission to claim all edits succeeded.

## Native output, replay and settings-only demonstrations

Five WAVs are independently hashed and associated with the recorded Chrome PID:
three from the real-provider native browser sessions, two from the final
OpenAI/Gemini retained-event supplements. The audit decodes each with ffmpeg to
mono float PCM, measures a silent first 1.5 seconds and nonzero output after
2.5 seconds. Each capture lasts over eight seconds. These measurements establish
native output, not intelligibility, a transcript or a human listening claim.
Their mono RMS measurements need not equal the student's per-channel analysis.

The initial tab timeout interference remains recorded. A control in one shared
browser context confirms the defect; the earlier separate-context observation
alone was insufficient. The repaired production speech service's actual-native
control holds A beyond the former timeout while canceling waiting B without
interrupting A. The deterministic browser suite separately covers exact lease,
callback, unsupported-coordination and close races. Coordination remains scoped
to cooperating same-origin/storage-bucket documents.

The Anthropic native session overlaps shared autoplay disable with existing
queued speech. In OpenAI/Gemini's initial native sessions, speech completed too
soon to establish that overlap. Their endpoint-disabled supplements use retained
real-provider cards with explicit local Speak actions: both Pages hold speaking
causes, revision 2/rate 1.2 entries survive revision 4/rate 1.6 plus autoplay off,
Cancel A leaves B speaking and typing, and Cancel B leaves typing. A later manual
utterance uses rate 1.6 and produces fresh native audio. These are actual browser
and native-service operations over replayed text, not fresh model generations
or evidence that replay automatically speaks historical cards.

All three replay launches bind their retained source log and set the provider
endpoint to `127.0.0.1:1`; no new raw request is present. Public `Agent.Append`
assigns a new top-level admission time. The independent audit compares headers,
event count/order, sequence and every other field exactly, allowing only that
specified timestamp change. The original strict equality failure remains in the
student record. Support-only `52c9561` corrects the helper assertion and has its
own successful launcher control; it does not revise provider evidence.

Three zero-request fresh production launches demonstrate persistence: positive
light/autoplay/limit-two values, an explicit false/zero edit, and a later process
restoring false plus stored zero/effective 16. Their first wire snapshots equal
the preceding complete files. Retained two-tab preference conflicts and explicit
fresh-base retries supplement the deterministic writer/failure checks.

## Limits retained

- The original Anthropic tool requests in the public example are refused because
  that example exposes no tools. Its two waits time out; the separately bound
  text-only correction establishes interruption.
- Gemini's first embedded interruption is too late after its 512-token response.
  Those `MAX_TOKENS` outcomes and failed wait remain. Its later paused-proposal
  control establishes an active interrupted turn without claiming a long answer.
- An interrupted transport need not have a final provider usage record. The
  receipt totals accepted `response_ended` usage only; empty usage is not zero
  provider billing. Captured requests and partial raw response bytes are retained.
- Exact disk failure, publication races, numeric boundaries, unsupported parser
  behavior and long request-limit cases remain local distinguishing controls,
  not invented live provider failures.

The implementation/comparative gate and this live audit are accepted with these
explicit scopes. Final manuscript/feedback reconciliation and immutable export
remain separate coordinator gates.


## Final teaching append and reproducible recheck

Student confirmation `446d7f2` appends to `student-review.md` after the audited
freeze. The original `c84f46b` result remains unchanged. The
[final recheck](checkpoint-evidence/ch08-live-final-recheck.json) resolves the
original teaching bytes from `7f517d8` and accepts only that explicitly identified,
committed append whose prefix is byte-for-byte identical. Every other evidence
file still has to match its frozen Git blob. Reproduce this final state with:

```sh
python3 scripts/edition2/ch08-review-evidence.py --teaching-append 446d7f2
```

The recheck repeats all source/launch/request/audio/feature assertions and adds
a passing append control plus intended refusals for an unbound edit, a rewritten
prefix and a changed raw terminal. All 30 controls pass; all 1,067 frozen file
hashes are unchanged, with the separately bound teaching append recorded outside
that map. No broad exemption for modified metadata or receipts was introduced.
Final manuscript and student confirmation are accepted in the
[proofread receipt](checkpoint-evidence/ch08-final-proofread.json); export/tag
remain coordinator work.
