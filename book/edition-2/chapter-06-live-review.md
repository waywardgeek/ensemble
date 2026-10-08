# Chapter 6 independent live-evidence review

Initial and revised evidence accepted at their respective source identities.
Coordinator review, October 7 local / October 8 UTC
2026. No paid requests were made during this audit.

The initial runtime is `aa5f86a4782c7479ea61b0e6abcbd4163968b574`. Initial live
evidence was frozen at `f73b01e`, with original durable logs at `d12a0cb` and
read-file artifacts at `3417575`. The archived CLI and public-consumer executable
copies have exactly the hashes in `initial-binding.json`; this audit uses their
archive locations rather than relying on the original temporary paths.

Reproduction:

```sh
python3 scripts/edition2/ch06-review-evidence.py
```

The source-bound result is
[`checkpoint-evidence/ch06-live-independent-initial.json`](checkpoint-evidence/ch06-live-independent-initial.json).
It identifies all 68 historical Go/module files, both archived executables,
the review script, and every original file in the nine run directories. The
student verifier also checks its three support-file identities, the Python
interpreter and terminal recorder. This preserves exact local evidence; it is
not a claim that another operating-system installation contains those binaries.

## What was checked

All nine launches finished with exit code zero. Their 33 captured request bodies
match replay from the exact historical CLI and recorded request configurations.
The student's verifier compares multisets to accommodate concurrent Agents.
The independent audit additionally compares the six single-Agent sessions in
chronological order, so a reordered pair cannot satisfy the human-session check.
The three public sessions retain multiset comparison because their two Agents
can reach the relay in either order.

The coordinator read all six human terminal transcripts in full and inspected
the launch/relay/consumer source. Streaming transcripts show provisional text,
active hint acknowledgments, real `read_file` proposals and continuations,
successful interruption with incomplete display, then `RECOVERED-SIX` on a new
turn. Plain transcripts show `PLAIN-SIX`, a completed read and its actual marker
and port, with no provisional proposal output. All six recorded tool results
match both `notes.txt` and the corresponding `cr/io/1` file bytes. Printed final
usage equals the sum of accepted durable responses; interrupted operations add
neither accepted responses nor tool effects in these retained sessions.

| Streaming session | Hint receipt sequence | Consuming request sequence | Captured wire request |
|---|---:|---:|---:|
| Messages | 4 | 9 | 2 |
| Chat Completions | 9 | 14 | 3 |
| generateContent | 14 | 19 | 4 |

Each hint is absent from earlier captured bodies, appears exactly once as human
guidance on the consuming request, and is absent from later bodies. Gemini's
hint is a separate text part in `role:"user"`, outside `systemInstruction`.
These hints carried over to the next human turn because the active model
operation finished without an automatic continuation. Receipt, wire delivery
and subsequent model compliance remain distinct facts.

All public records contain two distinct Agent identities. Finals-only data
matches complete parts from the corresponding reliable completion and durable
response, including Gemini's signed empty text parts. The consumer source
waits for reliable handles and captures the result before releasing its blocked
observer callbacks. Its small live responses did not overflow subscriptions;
the separate deterministic overflow probes establish that behavior. No selected
model exposed a supported thinking delta. The initial consumer ownership gap
identified as R4 remains a code-review finding despite these behavioral receipts.

## Controls and limitations retained

Sixteen checks include the original positive batch, a positive copied-path
batch, thirteen independently mutated identities, and a late request-body
mismatch. Identity controls cover wrong/incomplete/empty historical source,
each executable, support hash/completeness, and the last launch's source,
executable, support and role. Each fails for its intended reason before the
output directory exists. The last-body mismatch also fails before any derived
write, after preceding valid requests have been reconstructed. Every original
run file is unchanged. Comparing evidence against the three authorized key
values in memory found no credential match; neither keys nor settings were
printed or copied into the review receipt.

The initial timing misses remain visible: one OpenAI prompt ended before an
active hint could be sent; Gemini refused an idle hint, hit `MAX_TOKENS` on the
longer explanation, and completed before one interrupt. Later bounded attempts
demonstrated the intended active controls. Gemini's two public responses also
hit their generation limit and are accepted partial output, not completed
explanatory tasks. No new success is inferred from a prose summary.

This accepts the initial evidence's identity and stated scope. R1–R4 repairs,
full independent acceptance and final revised live use remain separate gates.
Earlier receipts must retain their original source identity; repeat only the
demonstrations affected by the reviewed changes.

## Revised live acceptance

The repaired runtime is `75bd14d5d2424778ebf45cb9025e9dbf314f956a`. Seven revised
sessions, frozen at `8c73f9b77d9f162079d11ce305b03528c023f3e9`, made exactly 20
requests. The CLI and consumer binaries were archived with their recorded
hashes. Reproduce the independent local audit with:

```sh
python3 scripts/edition2/ch06-review-evidence.py --revision1
```

The [revised receipt](checkpoint-evidence/ch06-live-independent-revised.json)
binds 73 historical source files, both executables and all 75 original run
files. All 16 controls pass, including copied-path positives and intended
identity/body refusals before derived writes. All 20 requests replay; the four
single-Agent sessions also match wire order. The same generalized checker was
rerun against the initial batch and retained all 16 passing controls and 33
matching requests. It does not relabel that earlier batch as revised runtime.

The coordinator read all four revised human terminal transcripts. Each of the
three streamed sessions completed a real file read and continuation, accepted
an interruption after visible integer output, labeled the display incomplete,
and answered the next prompt with `RECOVERED-SIX-REVISION`. The Gemini plain
session completed the same read without provisional display. All four read
artifacts match their workspace files and durable tool results. Accepted usage
matches each printed total; interrupted operations contain no accepted
response or tool effect. The revised plan did not repeat hint demonstrations
or the unchanged Messages/Chat Completions plain paths; those retain the
initial source-bound evidence above.

Each revised public session completed two independently identified Agents.
Finals-only observations match reliable completion and durable response parts.
The revised consumer's owner interfaces, checked subscription errors and
deferred release were inspected: completion and output precede release of
the slow callbacks. Neither live overflow nor exposed thinking is claimed.
Both Gemini public responses again ended with `MAX_TOKENS`; accepted partial
output is not evidence that the explanatory task finished.

All original revised files remain unchanged. An in-memory comparison against
the three authorized keys found no credential matches. No model request was
made by this independent audit. Refusal replay, deadline stalls, escaped-byte
limits and guaranteed display overflow retain their separate deterministic
fault coverage in the final gate at `5a95578`.
