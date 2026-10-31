# Edition 2 Chapter 11 grader submission

Coder: `edition2_grader_coder`, GPT-6 Astra/high, October 11, 2026.
Initial submission; independent review pending. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

Read the entire skill, AGENTS, carryover, workflow, enhancement/map, original
Chapter12, complete student/comparative reviews, original grader and relevant
accepted/legacy implementation. Reloaded full skill/carryover/workflow after
compaction. Original chapter revision `ec41c6e8cf0800e266f7a39ebadf255d94379082`,
SHA-256 `a754bd50ca7f3c7a2da868ff5acbdb98b4d5eddaa05349e7890c4a3dbf6cd4b9`.
Accepted source is direct `solutions/edition-2/ch11`, original12 checkpoint
`edition-2-ch12-r1`; legacy is `solutions/ch12`. Frozen/runtime/original grader
files remain unchanged. No mocks or paid inference added.

## Scope and real boundaries

Preserve the six original weights 15/15/15/20/15/10, plus 10 for all corrected
Chapter10 checks. Command: `go run ./cmd/grade-edition2 -ch 11 [core-directory]`,
default accepted `ch11`. Existing separate GUI/inherited fixture flags remain.

The original pipe harness performed initialize/discovery and an unknown reverse
request, then awarded 35 tool/ephemera points without a model conversation. Its
WebSocket check searched fixed source filenames. The revised driver separates
three actual public paths, without adding a product protocol:

- Dedicated `--mcp-pipe`: newline JSON-RPC initialize, initialized notification,
  distinct discovery ID, then a correlated enabled file read whose returned
  marker never appears in arguments. Unknown and hidden automatic names must
  produce actual protocol/tool errors, not successful diagnostic text.
- Human `chat`: a valid primary skill starts a real Python stdio peer. Two turns
  and five actual provider requests establish discovered names/description/schema,
  ordinary call arguments, both returned text blocks, correlated success/error,
  hidden automatic declarations and guessed-call refusal. Each request contains
  exactly its fresh round marker; turn markers appear only on requests 1 and 5.
  A failed automatic call is omitted while both turns complete. Actual subprocess
  RPC records establish five round calls, two turn calls and both ordinary calls.
- Actual separate GUI executable and `/ws`: subscription, JSON-RPC envelope
  handshake and a correlated reverse file result cross both framing directions.
  This establishes tunneling, not browser DOM effects or model attachment.

All paths keep actual Agent, actor, jobs, skill admission, reducer and filesystem.
Only external JSON-RPC peers and provider HTTP are small faithful fakes. The
historical wire version is 2024-11-05; course reverse/ephemeral/cancellation
extensions are not presented as current-standard claims. There are no required
source filenames, method names, constructor shapes or new runtime features.
Chapter10 retains its direct persistence checks and narrow inherited independent
session launcher; MCP introduces no wrapper around those checks.

The CLI completion barrier observes unique final answer text with streaming
explicitly disabled before sending the next prompt. Otherwise the actual actor
may legitimately interpret buffered second input as a hint. Ephemera assertions
inspect model message content, and tool assertions inspect correlated result
blocks, rather than searching declarations or earlier call arguments.

## Fixture discoveries and actual results

Evidence: [enhancement11](../../solutions/evidence/ch12/enhancement11/).
`initial-baseline.txt` retains my invalid comma-plus-space tools header. The
published parser uses whitespace-separated names; commas became part of the
first five names, leaving only final `auto_broken` matched. The corrected valid
header passes 90/90 in `corrected-baseline.txt`. This was a fixture failure.

`legacy-mcp.txt` initially scores 15/90 because the first reverse read arrives
before the legacy command registers its handler. Source confirms registration
follows the tools/list response, while the codec reader is already active.
There is no final ready acknowledgment in the published protocol. Coordinator
and reviewer independently accepted a bounded real-read readiness probe: retry
only correctly correlated error replies to a side-effect-free existing-file read
for at most one second, with distinct IDs. Wrong/empty successful data, malformed
or miscorrelated replies, EOF and response deadlines fail without retry. Unknown
and hidden-name refusals run once after that enabled control. The ordinary read
response deadline remains five seconds; a silent peer does not get endless retries.
`legacy-ready-mcp.txt` then scores 30/90. Final raw reports retain readiness attempt
counts: accepted 1, legacy 2. No implementation-specific error-string adapter or
invented acknowledgment is used.

| Measurement | Actual result |
| --- | --- |
| Historical original full student run | Interrupted/unscored, exit 143 |
| Historical unchanged focused student checks on interim source | 80/90; parity unrun, tool/ephemera credits vacuous |
| Final revised full accepted solution | **100/100 PASS**, exit 0, 16.843s |
| Final revised full legacy solution | **30/100 FAIL**, exit 1, 15.074s |

`accepted-full.json` and `legacy-full.json` are actual full runs, not composed
scores. Their `*-run.json` files preserve exact command/time/exit. The accepted
run includes corrected Chapter10 parity. Legacy passes raw handshake and reverse
execution/refusal. Its shipped chat ignores the supplied skill-declared stdio
startup and exposes its ordinary builtins; no external peer starts, remote names
are unknown and no automatic markers appear. Its shipped GUI never supplies an
observed RPC handshake before the bounded deadline. These are observed command
integration gaps, not evidence that every absent bridge/automatic/tunnel operation
was individually executed and broken. Library helper presence cannot establish
the actual command wiring. No runtime patch or adapter hides those gaps.

Legacy parity remains zero: actual null-save overwrite and skill failures alone
are sufficient. The full raw nested report also retains known inherited setup
limitations, including absent separate consumer/workflow modules, a consumer launch denied
as non-executable, a different GUI module layout and default Anthropic/OpenAI
media rows missing from the legacy table. Those are not newly
established behavior failures. Current MCP chat uses its supported plain fake-model
path; GUI uses declared claude-sonnet-5; current-config persistence uses declared
gpt-5-course. No corrected full legacy media/parity result is claimed, and the
unchanged recursive cascade was not repeated for freshness. No new full original
grader result or historical legacy 100/100 result is inferred.

## Sensitivity and coverage limits

`mutations.json` records 22 isolated compiling defects, exact patches and complete
failed-check arrays. Each defect fails its intended new predicate; no survivor or
invalid compilation attempt occurred. `mutation-baseline.txt` and
`mutation-restored.txt` pass90/90. The bridge process exits 0 when it collected a
report, including failed mutants; that is not a passing mutant grade.

| Defect | Intended observation |
| --- | --- |
| Wrong initialized notification | Invalid handshake sequence |
| Wrong reverse ID; successful empty read; success flag on refusal | Correlation, actual file bytes, real refusal |
| Ordinary dispatch permits automatic tool | Hidden guessed call executes and extra peer invocation is observed |
| Missing discovered tools; visible automatic tools | Actual provider declaration set |
| Changed description; changed schema | Preserved discovered metadata |
| Dropped arguments; empty result; lost second block; cleared isError | Actual correlated result/error and peer arguments |
| Missing round; missing turn; turn on every round | Invocation cadence and fresh context |
| Retained consumed ephemera | Stale round/turn content on later requests |
| Injected automatic failure text; abandoned collector on error | Failure text leaks, or the turn cannot continue |
| Skipped stdio startup | Missing peer and its declarations |
| Wrong outbound envelope; discarded inbound payload | Actual GUI framing in each direction |

The fatal-automatic-error mutation changes the error branch from continue to
return, abandoning collection before any request can complete. The five-second
output deadline catches that lost continuation; passing controls complete both
turns. This is distinct from rejecting failure text as model context.
The incoming-tunnel mutation reaches the real GUI and prevents handshake replies
from reaching its client, so its bounded deadline is an intended routing failure,
not a compile/setup kill. Coupled additional failures are not counted as distinct
proofs. These cases establish selected sensitivity, not exhaustive protocol or
concurrency correctness. Independent `reviewer-readiness.json` records a persistently refusing real reverse
handler failing after the retry bound, with baseline/restored controls passing.
That reviewer-owned evidence is separate from the 22 coder mutations.

Ownership, out-of-order pending requests, per-server grants, compiled-name
shadowing, job cancellation, ordinary callback versus actual automatic completion,
interruption preserving jobs/results, subprocess environment and unload/reaping
remain source/integration obligations. They are not guarantees from the 100-point
number. Existing original12 tests and mutation records retain these unchanged
claims. For exact bridge validation, all six accepted `TestMCP*` tests pass under
race in an isolated copy with the published bridge excerpt substituted into its
actual method; this includes the 3.4-second round/turn checks and interrupted late
result control. Numeric automatic fixtures are intentionally fast and do not
replace those timing proofs.

## Exact samples and historical live reuse

Manuscript SHA-256:
`37ece94fb7c0cf0a3706dad8cd877d5238cb0e4c23b76455995787ed25a6e453`.
`samples.json` binds every extracted fence and each actual command. Both Go
fences run in their stated contexts: the transport interface in an outside
consumer using actual public raw/pipe constructors, and the private bridge
excerpt inside manager.Tools in an isolated accepted source copy.

The exact full browser-tools skill loads after a real pipe peer attaches, admits
its ordinary tools, hides automatic names and executes a correlated gui_click
RPC. This validates skill/server declarations, not DOM clicking. The code-search
fragment is wrapped with normal metadata and exercised using a separately supplied
faithful `scripts/search_server.py` subprocess. The manuscript explicitly requires
that companion; neither a published search implementation nor generic search
behavior is claimed.

The exact Python peer and base primary run through the published `/tmp/ensemble-ch11`
build and human `chat` path. Two captured local-provider requests contain a real
marker-file result and refreshed snapshot; actual counts are read 1 / round 2 / turn 1.
The exact dedicated `/tmp/ensemble-ch11 --mcp-pipe` command also completes its
handshake and correlated file read, then exits 0 on EOF. Coordinator flagged the
previous `./ensemble` ambiguity beside the accepted core's `ensemble/` directory;
author corrected the executable path before this validation. No protocol changed.

The retained original12 three-provider marker sessions and fifteen real Chrome
checks remain their historical source evidence. The initial Claude zero-exit/
failed-argument stream and corrected bounded repeat remain distinct. Browser
font evidence proves applied/effective state, not disk persistence; native queue
observations prove neither subjective audio quality nor paid-provider browser
use. No successful old live/browser claim was repeated for freshness here.

## Validation and handoff

`validation.json` records affected grader race test, command build/test, vet and
formatting success. Command package has no tests; its successful build is not
called a test run. `sample-bridge-and-lifetime.txt` preserves all six actual
source test names/results. Script and temporary bridge sources are plain-text
receipts; scratch trees are removed. No new generic grading framework is added.

Counts: new grader 476 active /55 comment-only /545 physical Go lines; regression
22/2/27. Original preserved harness/checks total 389/55/491. Active means nonblank
and non-`//`-only; inline comments remain active, and the small embedded Python
peer counts within its containing Go file. Added behavior replaces false
observations with three actual boundary sessions; line counts alone establish
neither quality nor coverage. Comments are 10.4% of nonblank grader lines, below
the roughly 20% teaching target; they explain isolation, ownership of stdout,
readiness, completion barriers, result-only assertions and proof limits. The only
new exported function, Chapter11, and the test function have contract comments;
no exported types, fields or interface methods were introduced.

Tested behavior checkpoint `0b9a9abe` precedes explanatory-comment-only additions.
`final-source.json` binds those final files and both immutable solution tree IDs.
Commits use an isolated index with matching base/parent, configured hook and
atomic expected-old reference update; only owned entries are reconciled. Original
source, frozen exports, shared book/chapters.md and all CodeRhapsody work remain
untouched. Stop at this submission until combined Chapter11 review/release.
