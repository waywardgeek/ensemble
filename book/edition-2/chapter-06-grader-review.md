# Chapter 6 independent acceptance coverage

Contract-first plan, 2026-10-07. The grader engineer loaded the complete coding
skill, architecture, Chapter 6, voice and chapter-writing procedure before
implementation. Contract SHA-256 at planning was
`2b5960dfcd8c77ee4a0194d97b3ee06a23e128f45e4d4813903c6a1613796e54`.
This role previously graded and reviewed Chapter 5, including its historical
comparison. It has not inspected the Chapter 6 student implementation or the
first-edition Chapter 7 answer at this milestone. The later comparative review
must wait for preserved initial source, actual runs and student review.

## Coverage plan

These groups derive from §6.9 and the detailed requirements it summarizes.
They are a coverage plan, not a report that the complete checks exist or pass.

| Group | Independent distinguishing evidence | Required deletion or negative control |
|---|---|---|
| Early output | Hold terminal frames until a public observer and, separately, an actual chat PTY receive the first fragment; repeat all three APIs | Buffer until response completion; suppress CLI flush |
| Correlation | Text plus two interleaved calls, another operation, then two Agents; collect by full identity and verify each accepted final against its durable response sequence and index | Merge argument buffers; reuse operation identity; map a final to another part |
| Typed equivalence | Paired stream/plain fixtures compare ordered neutral parts, exact argument objects, signatures, requested/returned model identities and usage; include Messages thinking, Gemini signed empty text and signed calls | Lose opaque content; move a signature; mix thinking with visible text; normalize stream and plain differently |
| Framing and bounds | Every byte split; UTF-8/CRLF/CR/BOM/comments/multiline data; exact 1 MiB physical frame and one-byte overflow including ignored frames; 16 MiB assembled content; unfinished EOF and non-SSE success | Count decoded data instead of wire bytes; dispatch EOF fragment; reset bound incorrectly; accept invalid UTF-8; ignore content type |
| Terminals and accounting | API-specific lifecycle, usage-only final, cumulative replacement, required fields, consistent returned model, each accepted usage increment once; valid limit completion and malformed-argument rejection | Omit terminal or required usage; add snapshots; accept conflicting model; discard valid limit response |
| Effects | Hold after complete-looking tool arguments, then truncate, error or interrupt; no new response, tool effect or usage; retain an earlier accepted round | Dispatch a proposed call; persist partial response; account failed operation |
| Job concurrency | Job terminal consumes a sequence while HTTP remains open; missing Gemini call IDs and part finals use the actual later append sequence | Predict the future sequence before job completion |
| Lifecycle and capacity | Full producer capacity with cancel/interrupt/close/timeout, bounded fragment notices, 64 KiB actor fairness, next-turn usability and stale-fact rejection | Uncancelable capacity wait; fragment flood monopolizes actor; late success survives interruption |
| Observer/display recovery | Slow subscriber overflows independently; another subscriber and reliable completion continue; CLI labels incomplete display and recovers final text, with gap before completion | Block actor on observer; consume completion; silently lose display or duplicate normal answer |
| Compatibility | Default protocol exact prior output; opt-in ordering; plain no deltas; unknown flags/env fail before HTTP; historical absent delivery means plain; replay restores delivery and sends no HTTP | Always emit observations; ignore recorded delivery; accept invalid disable setting |
| Ownership | Discover all spokes/modules/executables; inspect actual operation→Engine→Agent→Ensemble chain, parser diagnostic access, actor-only persistence and lifetime/lock boundaries | Sibling import; injected callback/service bag; bypassed parent; worker persistence |

Every mutation starts from a passing positive fixture, changes one intended
behavior, and must fail at its intended assertion. Setup/path failure does not
prove a later semantic guard. Score and complete promise coverage cannot be
inferred from a few check-level mutations. Existing Chapter 1–5 and historical
grader code remains unchanged. If shared infrastructure later needs edits,
record its legacy baseline first and retain reference, mutation and required
regression passes.

## Initial executable milestone

Run against a built CLI:

```sh
python3 scripts/edition2/accept_ch06_clients.py /absolute/path/to/ensemble
python3 -m unittest discover -s scripts/edition2 -p test_accept_ch06_clients.py -v
```

The checker currently covers six CLI cases: actual PTY and opt-in protocol for
each API. Each local server withholds terminal frames until the first fragment
is observed. It checks streaming request options, basic operation/part identity,
text accumulation, final ordering and mapping, one completion, normal display
without duplicate text, and opaque-signature suppression in the human view.
Its output explicitly labels the scope partial. It awards no chapter score.
The public library observer barrier and the other groups remain to implement.

Three harness test methods passed: six positive executable-fixture cases,
buffer-until-terminal mutants for both client modes, and an operation-identity
mutant. The buffered mutants fail specifically at the held-open first-fragment
barrier; the identity mutant gets through that barrier before its intended
refusal. These tiny executable fixtures validate checker mechanics, not the
student implementation, provider parsers or real-model usability.

The accepted Chapter 5 executable is a deliberately pre-streaming negative
baseline, identified by its archived SHA-256 and runtime revision in
[the baseline receipt](checkpoint-evidence/ch06-grader-baseline.json).
Its six failures establish that this new check does not silently accept the
predecessor. They are expected Chapter 6 feature absences, not retroactive
Chapter 5 failures. No shared historical checker was changed, so this milestone
does not claim a newly rerun historical regression suite.

## Next handoff

### Coordinator wire-check milestone

The coordinator added `accept_ch06_wire.py` and its receipt-assertion controls
while the author occupied the available worker slot. This is independent of
the Chapter 6 student implementation. The coordinator had inspected preceding
second-edition public data declarations and the new ownership plan, but did not
read the historical Chapter 7 answer to derive these checks.

```sh
python3 scripts/edition2/accept_ch06_wire.py /absolute/path/to/ensemble
python3 -m unittest discover -s scripts/edition2 -p 'test_accept_ch06_*.py' -v
```

All 61 wire cases passed against the CLI built from initial student source
`aa5f86a4782c7479ea61b0e6abcbd4163968b574`. The coordinator verified all 68
historical Go/module file identities. The receipt
`checkpoint-evidence/ch06-wire-initial.json` binds source, binary and both
checker files. The earlier 57-case development result remains separately
recorded; it predates the additional cumulative-usage and model-conflict cases.

These checks exercise paired full plain/stream typed responses and accounting,
known expected parts, CRLF/CR/BOM/comments/multiline framing, exact data/comment
frame sizes and overflow, repeated frame-counter reset, one-byte server writes,
required usage and terminals, invalid UTF-8, non-SSE success, cumulative snapshot
replacement and conflicting returned identities. A complete-looking proposed
write without the required API terminal must leave no file, accepted response,
final observation or `tool_called` event. Its filesystem witness is independent
of the model's or CLI's description.

One-byte server writes do not prove every possible client read split. Receipt
assertion controls prove their intended refusal with valid positives; they are
not substitutes for deleting the corresponding student behavior. Full public,
concurrency, capacity, replay, ownership and implementation-mutation coverage
remain pending. Neither this command nor the initial client checker awards a
chapter score. Historical graders and shared infrastructure remain unchanged.

Resume this independent role after the author has capacity. Reload the entire
`book/edition-2/skills/ensemble-coding/SKILL.md` before edits, and the current
chapter before extending fixtures. Implement public consumer checks through
the student's chosen public API spelling without imposing new requirements;
then complete framing, semantic, lifecycle and architecture groups and their
mutations. A separate `accept_ch06.py` orchestrator can combine these checks
with preserved prior coverage once the groups exist. Do not label the initial
CLI command full chapter acceptance in the manuscript. No paid call is needed
for any work in this plan. Actual live receipts and the later historical
comparison remain separate mandatory gates.

### Coordinator CLI overflow/recovery probe

`ch06-review-cli-recovery.py SOURCE_COMMIT` exports committed Go/module files
to disposable directories and injects `ch06-review-cli-recovery_test.go.txt`.
It runs the real human and observed-protocol loops, real HTTP parser, Agent and
Observer, with an output writer blocked at the first visible fragment. Each
following wire fragment waits for acknowledgment that the real client bridge
received the previous one. Three hundred deliveries therefore overflow its
bounded queue without relying on a scheduling race or a fast producer. The
server completes and the observer sees durable `turn_ended` while output is
still blocked. Releasing output must recover the complete reliable answer.

The four cases cover two-fragment controls and 300-fragment overflow for both
clients. Protocol must report exactly one correctly identified gap before its
single complete successful completion. Human chat must report the gap before
completion and label the recovered whole answer; ordinary small streams must
show neither a gap nor duplicated final text. The writer seam makes ordering
deterministic; this fixture complements actual PTY delivery tests and is not a
claim of live-provider or PTY overflow.

On initial `aa5f86a`, all four pass under race detection. Suppressing the gap
callback fails exactly the two overflow cases; suppressing human recovered
text fails exactly that human overflow case. Both small-stream controls still
pass. The retained receipt is
`checkpoint-evidence/ch06-cli-recovery-initial.json`, binding all 68 historical
Go/module sources, both probe files, commands and exact failed leaf-test sets.
No production source or historical grader is changed. Full Chapter 6 acceptance
remains the reviewer's broader gate.

The first attempt to broaden this probe to the whole main module omitted
committed `testdata/history.log` from its disposable export. The targeted probe
passed, while four existing replay-fixture tests correctly failed at missing
input. That failed receipt remains as `ch06-cli-recovery-fixture-omission.json`.
The corrected runner includes and hashes committed testdata; whole-module vet
and tests then pass. An intermediate directory-entry path error stopped the
runner before tests and was corrected by filtering non-files first.

### Data-frame overflow fixture clarity

The reviewer questioned whether adding a byte before a field could accidentally
test malformed content instead of the frame bound. Inspection established that
this check's exact frame began with a padding comment, so the original extra
`x` changed `:` into the valid ignored field `x:` while retaining the data line.
It did not remove the model payload. The fixture is now easier to assess: add
one JSON whitespace byte inside the existing data field. A new fixture control
proves exact and overflowing frames decode to the same JSON with byte counts
1,048,576 and 1,048,577. All 61 wire cases pass against the archived initial CLI;
all seven client/wire checker test methods pass. The source/binary/checker-bound
receipt is `checkpoint-evidence/ch06-wire-fixture-clarification.json`.

The original receipt remains historical with its original checker hash. This
change strengthens fixture clarity without removing an assertion or changing
student/historical runtime code. Separate implementation deletion evidence in
the reviewer's frame group proves that removing the bound is detected.

## Expanded independent implementation milestone

The independent reviewer completed contract-derived public and internal groups
against frozen initial source `aa5f86a`, after the initial live/student-review
freeze permitted comparative source inspection. Eleven of thirteen groups
passed; recognized Chat refusal replay and whole-operation timeout exposed
material runtime defects. The consolidated findings and repair history are in
[the code review](chapter-06-code-review.md). Eleven initial implementation
mutations had passing positives and their intended behavioral failures. A
size-doubling benchmark independently demonstrated quadratic initial assembly.

The first repair `3ccaed6` passes those thirteen original groups and the retained
Chapter 5 assertions (100/100) via a separate delivery/interface adapter. The
adapter changes only the old fixtures' delivery setting and transport double,
and encodes workflow fixture responses as SSE for that default-streaming public
consumer. Assertions remain unchanged. The inherited historical grader and
its incompatible 0/100 result remain unchanged too.

A new escaped-opaque boundary check then caught a remaining R2 defect: decoded
thinking length was charged while a larger escaped JSON payload was retained.
Its exact-limit positive and one-byte-over control include control characters,
quotes, backslash, HTML and Unicode escaping. The intermediate repair is
preserved and is not a final accepted source.

The complete deterministic command is:

```sh
python3 scripts/edition2/accept_ch06_gate.py SOURCE_COMMIT
```

It creates an immutable disposable source archive, builds the CLI, runs the
six early-delivery client checks, 61 wire cases, expanded public/internal
groups, retained prior assertions, implementation mutations, CLI overflow
controls, an allocation-scaling observation, and all discovered module vet/tests.
The final receipt must pass; the command's existence is not a claim that the
chapter is validated. Live evidence, manual ownership/comparative assessment
and manuscript proofreading remain separate gates. The initial and first-repair
receipts retain their exact source maps and captured output even as the fixtures
are strengthened for the final gate.

One genuine checker defect was found by deletion: the reviewer's first draft
of the isolated data-frame overflow fixture prepended a byte to the field name,
which could fail for missing content instead of the size limit. Inserting that
byte into the valid data payload gives the intended frame-bound refusal. A
separate signature deletion initially failed earlier shared validation; it was
replaced with valid but changed signature content to isolate retention. Neither
setup/earlier failure is credited as proving the later assertion. These are
reviewer fixture corrections, not student runtime defects.
