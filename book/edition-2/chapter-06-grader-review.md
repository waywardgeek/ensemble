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
