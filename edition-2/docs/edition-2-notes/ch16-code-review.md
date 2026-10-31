# Original Chapter 16 / Edition 2 Chapter 15 — independent review

## October 11 preparation: original forcing and provider usage

Reviewer `enhancement_reviewer_resumed`. This is Bill-authorized, narrow
preimplementation inspection of the original answer, **not student review or
chapter acceptance**. Bill resolved the manuscript's cumulative-usage wording:
forcing uses total input of the prior request, including correct provider cache
accounting. The coordinator records the student-safe ruling and current official
provider sources in [ch16.md](ch16.md). Keep the original implementation details
below out of the student's handoff.

Read the entire mandatory `.agents/skills/ensemble-coding/SKILL.md`, carryover and
workflow. Inspected Anthropic and OpenAI first, then Gemini, as Bill requested.
No product, original source or grader changes; no tests, grader or paid calls.

**Bill's recollection is correct.** Original `internal/llm/forcing.go:85–100`
explicitly distinguishes the last request from running totals. It scans backward
to the latest `ResponseEnded` carrying a response and uses that response's
`Input + CacheRead + CacheWrite`, excluding output. This is provider-reported
input from the latest completed response, not a newly estimated context size.
The separate running total is accumulated by `internal/common/context.go:200–205`
and is not the forcing numerator. `forcing.go:38–51` divides the selected count
by the current model's context window; absent window or zero selected usage
returns without forcing. Missing usage is therefore not independently recognized
as unknown in this implementation.

The original test `forcing_test.go:101–117` records 3,932 input tokens and then
512, asserting that tools return after the latter. Their sum still exceeds the
4,096-token window, so the test's expectation supports the last-request reading.
This is source inspection of that test, not a newly executed result.

### Exact parser accounting

All paths ultimately emit canonical usage in `ResponseEnded`. The common usage
type documents disjoint input/cache-write/cache-read/output buckets
(`internal/common/event.go:315–332`).

| Provider | Original plain-response normalization and resulting forcing input | Streaming behavior |
| --- | --- | --- |
| Anthropic | `claude.go:375–380,448–454`: Input=`input_tokens`, CacheWrite=`cache_creation_input_tokens`, CacheRead=`cache_read_input_tokens`. Forcing adds those three disjoint fields. Neither output nor any nested cache breakdown is added. | `claude.go:587–595` takes input/cache counts from `message_start`; `640–644` changes only output from `message_delta`; `666–670` emits the final usage. Input/cache updates appearing later would not be incorporated by this parser. |
| OpenAI Chat Completions | `openai.go:281–288,344–356`: Input=`max(prompt_tokens−cached_tokens−cache_write_tokens,0)`; the latter two fields come from `prompt_tokens_details` and become the two cache buckets. With consistent subset counts, forcing's sum is exactly `prompt_tokens`, not prompt plus cached tokens. An absent optional cache-write field decodes to zero. | `openai.go:431–438` replaces local usage whenever a usage object arrives, including a no-choices usage chunk; `498–500` emits it. It does not add usage chunks together. |
| Gemini | `gemini.go:321–326,391–402`: Input=`max(promptTokenCount−cachedContentTokenCount,0)`, CacheRead=`cachedContentTokenCount`, CacheWrite=0. With consistent subset counts, forcing's sum is exactly `promptTokenCount`. Output is candidates plus thoughts, but is excluded from forcing. | `gemini.go:479–483` replaces local usage with each nonzero `usageMetadata`; `566–568` emits it. The comment's “cumulative” means within this response's stream, not lifetime accumulation across requests. |

The clamps in OpenAI/Gemini mean contradictory subset counts exceeding the
reported prompt count could make their canonical sum exceed that prompt count.
That is an algebraic limitation, not evidence that a live provider supplied such
data. Likewise, original comments about historical live cache fields are source
claims, not fresh verification of today's schema. The coordinator's official
documentation check must distinguish those claims from currently documented
fields, especially OpenAI `cache_write_tokens` and Anthropic stream updates.

The narrow contract conclusion is: use the previous completed request's total
input; add Anthropic's disjoint cache buckets, but do not add already-included
cached subsets again to OpenAI/Gemini prompt totals. A cached token still occupies
context. This confirms Bill's ruling without supplying an implementation recipe.

### Provenance

Inspected from repository HEAD `1878e64` with no modifications under original
`solutions/ch16`. Its latest directory revision is
`37697fe3f133cb3f85d86e7475375a654c23cb3d`. Individual last-changing revisions and
SHA-256 values below identify the exact files; paths are relative to that directory.

| File | Last-changing revision | SHA-256 |
| --- | --- | --- |
| `internal/llm/forcing.go` | `050f9c14d0c7328cff3b898a1d66c322edfc6589` | `fc72060d2bd4244264300f29d537f509cbb9ccaf67cdc3927d7ac3c32acc8e5d` |
| `internal/llm/forcing_test.go` | `050f9c14d0c7328cff3b898a1d66c322edfc6589` | `2a1fe886bb8ca70801e21134b092e32d6de33be9a6a418bc246c6507edf4910b` |
| `internal/llm/engine.go` | `37697fe3f133cb3f85d86e7475375a654c23cb3d` | `358c6e448f7832cd7798889cdedcc3e435916153d23c6bc44a6ac904ff5993d3` |
| `internal/common/context.go` | `dbe4a56a445d5724ff2fc9bd3281b57686fde5cd` | `2c2663c0b244cd603b83ec11351467841e4f92d7c52cf1917fbde1c2b49d9af0` |
| `internal/common/event.go` | `050f9c14d0c7328cff3b898a1d66c322edfc6589` | `38b24f31c39ce85f48441985c2638e848dadb8b1c0b106706c707cbf5202ad20` |
| `internal/llm/claude.go` | `050f9c14d0c7328cff3b898a1d66c322edfc6589` | `25da4f9940e9fb39e48a14ac0b8a5ff5c7fb2d73b4b250200b91f72f82a640af` |
| `internal/llm/openai.go` | `050f9c14d0c7328cff3b898a1d66c322edfc6589` | `92d68fed485f740dfa5335982bf839d2cd4825f0c116e00a8362c38da2e0900a` |
| `internal/llm/gemini.go` | `050f9c14d0c7328cff3b898a1d66c322edfc6589` | `532dca26aab706f7beef3f6abf270dfc2cd791b8c1b3df37a7153651739c0de9` |

A narrowly scoped later-file comparison found `forcing.go` byte-identical in
original Chapters 17, 18 and 22 (same SHA-256 above). The distinction was already
correct in Chapter 16; it was not introduced by those later snapshots. This is
not a review of their other implementation or provider behavior. No formal
student review round has begun.

## October 11 grader preflight — intermediate source, no acceptance

Reviewer reread the mandatory skill and the current student checkpoint in
[ch16.md](ch16.md). This assessment uses committed student source
`ccd63509c3b2e08ff5a231b0409423b236953022` (tree
`58dd0d51c9eb34e1096a1134fa2d045c97d07c97`), not the coder's changing working tree.
Final provider, mutation and live evidence remains pending. No formal review
round or chapter acceptance begins here; no grader/live call was repeated.

The original grader completed **20/100 FAIL**, grader exit1 / make exit2, on
that intermediate source. The committed [provenance](../../solutions/evidence/ch16/original-grade-provenance.json)
and [output](../../solutions/evidence/ch16/original-grade.txt) are from
`bd4213bf0d52adafc5d9bc6b5112416201da7aea`; output SHA-256 is
`6be427fd976941c65c215f09d8332ec0eba2fe83ef03a0a57132d422ce06ca5d`.

**The settings-path exception is independently justified, narrowly.** Original
`ch16_harness.go:50–72` writes `cwd/settings.json`, launches the GUI in that cwd
and supplies no Agent-directory override. The student's GUI command at
`gui/cmd/main.go:77` selects `DataDir="cr"`; its existing owner-based settings
path is `gui/settings.go:91–92`, and the loader applies memory settings to the
public Agent at lines85–88. The grader's small conversation/band budgets and
disabled flags therefore never reach the Agent. Its normal conversation budget
is 3,000 bytes (`internal/memory/memory.go:16–39`), rather than the fixture's 300.
The output consistently reports zero compressors/memories. This is the same
pre-existing ownership mismatch as original15, not grounds to relocate settings
to global cwd or waive compression behavior. The chapter expressly leaves disk
layout to the student; Bill's Agent ownership remains binding.

The proposed exception covers the **fixture location**, and hence the following
eight failed check outcomes (77 points), not their valid behavioral obligations:

| Check (actual earned/possible) | Still requires independent supplemental proof on final source |
| --- | --- |
| `micro-handoff-compresses` 0/15 | A real checkpoint crossing an effective budget compresses oldest complete segments and removes replaced raw work. |
| `planted-fact-survives` 0/10 | The actual compressor receives the selected work; its returned memory reaches later provider context. |
| `graduation-fires-oldest-first` 0/12 | Crossing effective watermarks folds oldest files, publishes their replacement, removes those originals from active context, and leaves newer work. |
| `disable-enable-idempotent` 0/12 | Disabled bands leave context and stop receiving compression; re-enable rescans edited disk files, preserves canonical order, and ordinary restart respects disk edits. |
| `disabled-neighbor-refused` 0/8 | An actually full source refuses a disabled destination without losing old content, with the required warning. |
| `replay-needs-no-llm` 0/10 | Actual recorded compactions rebuild memory without another compressor call. |
| `memory-is-data` 0/5 | First-person compressor instruction and actual conversation memory, separate from the frozen system prefix. |
| `fresh-start-populates` 0/5 | Real disk memories populate before the first request with no prior save/log needed. |

Each fails before its later meaningful assertions. In particular, “no memory”
does not establish a replay, ordering or data-placement defect; neither does a
fixture exception establish those behaviors passed. Final review must assess
the student's real-runtime tests, captured effects, targeted mutations and live
quality, without treating names of passing tests as proof.

**The two awarded checks need different interpretations.** `forced-handoff`
12/12 actually presents high reported input and observes only `micro_handoff`
on the later request (`ch16_run.go:740–803`). Its fixture does not establish
Bill's latest-versus-lifetime distinction, provider cache arithmetic, 90% warning
or eventual restoration; those remain separate obligations. `abandon-on-restart`
8/8 (`863–917`) does not require a compressor to run or return its scripted500.
With no compression, old work remains and zero copies satisfy its at-most-one
memory test. It also does not launch a SIGKILL case. Thus the actual awarded
eight points are retained, but this run proves neither failed/abandoned-fold
recovery nor crash safety. Supplemental proof needs an actually launched fold,
its failure/interruption, retained source information and nonduplicating recovery.

**Inherited parity remains 0/3**, with exactly the four reported original15
failures: `skill-survives-the-ladder`, `ladder-is-recorded`,
`replay-equals-snapshot`, and `frozen-prefix`. Reinspection confirms the first
three retain the cwd-settings mismatch (`ch15_harness.go:366–369`) and stop on
missing cuts/uncut content before their later comparisons. The fourth passes
the preceding system/tools prefix comparisons, then stops at undefined `think`
before the MCP declaration predicate (`ch15_run.go:500–549`). The independently
accepted [original15 exceptions](ch15-code-review.md) continue to explain those
specific failures. They do not waive memory's interactions with recorded cuts,
skills, declarations, journal recovery or request replay; reuse unchanged valid
evidence and check affected behavior at the final gate.

Source provenance: original `ch16_harness.go` and `ch16_run.go` still match the
revisions/hashes recorded in the student-safe launch erratum. `ch16_checks.go`
is revision `174eb9778243914db251aa99f4ef9d61889ce5cd`, SHA-256
`4e99b4427dc115ced53c92747e73b3299de5f1995122a51fb226949ec9497521`.
Original15 harness/run are respectively `107c59d45ee2977d82910f1c1e734eabb73d538a`
and `374916a10263c4b32cc3dc22999043f07202f95c` (unchanged original files).
No source correction is requested by this preflight. Final comparative review,
complete student response and authentic provider evidence are still required.

## October 11 initial formal review — correction required

Reviewer `enhancement_reviewer_resumed`; frozen submission
`e1c89721d8ba266fd7521333ec2ba46f5258baa3`, extending accepted original15
`6aa7fda`. Read the complete final student response above its evidence links in
[ch16.md](ch16.md), original Chapter16, both editions' chapter deltas and relevant
active code/tests. Reread the mandatory skill, carryover and workflow.
**Not accepted: one confirmed current-scope blocker.** This is the initial
formal review, with zero correction rounds used; the following finding releases
round1 of the maximum three. Preparatory assistance remains separately attributed.

### Blocking finding: GUI settings and actor observation deadlock

`gui/gui.go:283` holds `Server.mu` across `updateSettings`; the new synchronous
`SetMemorySettings` waits for the actor's acknowledgment (`agent.go:254–270`).
During a response that actor can be inside synchronous Ensemble observation,
blocked acquiring the same GUI mutex (`gui.go:64`). Neither can advance. This
also prevents Agent shutdown. The settings implementation's intention to order
apply/broadcast with subscription is valid, as is the new requirement that an
idle settings acknowledgment follow its effective content change; neither
justifies blocking the actor's progress behind this cycle.

The independent [bounded real-runtime probe](../../solutions/evidence/ch16/reviewer-gui-lock_test.go.txt)
ran in an isolated `git archive` of the exact frozen submission. It uses the
existing public-root/Agent/GUI/WebSocket helpers and a small HTTP SSE provider
fake; no internal behavior was mocked or production source changed. An idle
patch was acknowledged. After a real `part_delta`, the ninth active-response
patch timed out; cleanup then blocked and the whole test ended at its 12-second
bound. The [complete stack](../../solutions/evidence/ch16/reviewer-gui-lock.txt)
shows the WebSocket handler waiting in `SetMemorySettings` while holding the
GUI lock and the actor waiting in `ObserveAgent` for that lock. This is observed
liveness failure, not an inferred race or a timeout excused as flaky.

Correction must preserve effective idle settings, ordered persistence/snapshots
and the existing nonblocking-progress/Agent lifetime contract while breaking
this wait cycle. Add a bounded real-runtime regression with an idle positive
control and settings during an active response, including eventual completion
and shutdown. Restore the defect to demonstrate that the check detects this
liveness failure. No new product protocol, paid repeat or broad test sweep is
required. All reviewer harnesses are released.

### Scope, architecture and comparison

The Memory child has its actual required Agent parent; common contains shared
data/interfaces rather than memory operations. File selection/persistence live
in the memory spoke, the actor commits history and worker results, and Engine
performs a finite compressor call. The optional GUI uses public interfaces.
No new global service, callback bundle, mock, sub-agent or recall machinery was
introduced. Existing methods are extended instead of a second orchestration or
transaction framework. Pending cuts, recorded replacement bytes and retained
retired files provide a comprehensible failure boundary for this chapter.
The GUI cycle above is the one implementation finding, not evidence that the
overall ownership design needs replacement.

Independent counts use tracked files at the recorded commits: physical lines,
active lines excluding blank/comment-only lines, and comment-only lines. Inline
comments count as active; the simple lexical method and full categories are in
[reviewer-counts.json](../../solutions/evidence/ch16/reviewer-counts.json).
Each delta is against that edition's own predecessor, not mismatched tree totals.

| Chapter delta | Physical | Active | Comment-only |
| --- | ---: | ---: | ---: |
| Student production Go | +900 | +732 | +122 |
| Original production Go | +1,925 | +1,156 | +615 |
| Student Go tests | +702 | +662 | +25 |
| Original Go tests | +1,101 | +860 | +133 |
| Student browser JS | +28 | +27 | +1 |
| Student HTML | +1 | +1 | 0 |

The student's three separate examples are unchanged (469 physical production
Go lines and 234 test lines); browser test/peer MJS is unchanged. README adds
15 lines. JSON, dependency locks and external evidence/launch scripts are
support, excluded from production/test totals; no production move inflates this
delta. Original browser JS/HTML is unchanged. The student's production-Go
comment share is about 16.5% of active-plus-comment lines, and 14.3% for this
delta, below the approximate teaching target. New exported types, fields,
methods and aliases have useful contract documentation. Ownership, byte/token
distinction, frozen work data, commit ordering, replay and interruption receive
local explanations. Routine selection and file loops do not need commentary
padding. The original has more explanatory prose but also larger common-package
behavior and Engine-side memory orchestration. The student's smaller, separated
implementation is clearer; line count is supporting evidence, not the verdict.

**The `summarizeSpan` departure is accepted.** Keeping that inherited function
literally untouched would retain an empty placeholder and the already-compressed
handoff. Removing the selected completed dialogue plus its included handoff,
while preserving independent skills/declarations, fulfills replacement rather
than inventing a new clearing policy. Whole-segment and replay checks exercise
that distinction; the change remains tied to a recorded successful replacement.
The minimum-two-file waterfall, canonical front placement and verified/unknown
model-window choices are also reasonable readings of the stated exercise and
Bill's corrected forcing contract. None introduces a later capability.

### Verification and exceptions assessed

The independent final-source memory race run passes:
`go test -race ./ensemble -run '^TestMemory' -count=1 -timeout=45s`,
`ensemble/ensemble` 1.773s. [Probe provenance](../../solutions/evidence/ch16/reviewer-initial-probes.json)
records commands, source tree and artifact hashes. This run includes the new
ordinary enabled-band restart regression: offline restoration retains recorded
bytes, then actual startup rereads the human-edited file without duplicating its
identity. It also exercises compression, graduation, failure/interruption,
forcing, disk controls and replay through the real runtime. Its success does
not contradict the separate GUI deadlock it never interleaved.

Read the coder's actual validation/mutation records and intended failure outputs.
The 27 distinct eventual mutation claims cover meaningful observable defects;
the fresh-disk panic, wrong-module commands, initially indirect settings failure
and interruption assertion repair remain failed/limited attempts, not clean
kills. Tests use real earlier Agent/jobs/history/renderers with provider-boundary
fakes. Compared with the original's more separated reducer/Engine fixtures,
the student's integration tests protect the ownership and completion path as
well as content. They still missed the new GUI/actor lock interaction.

The eight masked memory obligations in the preflight table have supplemental
evidence: selected source reaches the compressor and later requests; whole
oldest segments leave only after success; minimum-two oldest files graduate
through watermarks while retaining newer work; disabled neighbors retain source
and warn; disk toggles and ordinary restart reread edits; pure snapshot/log
replay makes no HTTP calls; all-provider memory is data outside instructions;
fresh startup reads per-Agent files. Failed/malformed compression and an actual
interrupted HTTP compressor retain source, while pending-cut restoration never
launches another compressor. These do not become original-grader passes.

I independently agree with the narrow settings-location and inherited
undefined-tool/parity exceptions already detailed above. The original result
remains **20/100 FAIL on intermediate `ccd6350`**, not the final source: 12 forcing
points are actual observations; the awarded eight abandonment points are
vacuous for recovery. Bill's prior-request/cache ruling is tested separately
from that grader, including a lower later request despite growing session totals.
The invalid fixture does not waive the GUI blocker or any valid memory behavior.
Inherited journal/cut evidence is reused; no new claim of power-loss durability
or arbitrary crash-interleaving exhaustiveness is made.

I compared retained live stdout with actual graduated files, including initial
Anthropic's missing rejection reasons, the corrected Anthropic's invented open
question, and corrected OpenAI's wrong owner/date despite correct stored bytes.
Gemini recalls the requested facts but violates the no-extra-tools constraint.
All five runs show useful mechanics, not five fully successful tasks; three
compressor calls and exit0 do not erase those quality/compliance failures.
Measured reductions are roughly 1.36–3.68, not proof of an 8x ratio. The bounded
repeats and honest limits satisfy the live exercise requirement; no further paid
run is requested for this deterministic lock correction.

**Initial disposition: needs correction for the single GUI/actor deadlock.**
Other assessed departures/exceptions are justified within the scope above; final
acceptance waits for round1 source, affected checks and a sensitive real-runtime
regression. No next chapter or export is released by this review.

## October 11 correction round1 — accepted with recorded exceptions

Reassessed frozen `05b43a0b46fa154201a8b3e97b22f96ae1047cb9`, main tree
`4a253c57dd62b4a6ad086618574ef46af7c2d724`, after reading the complete appended
student response and causal mutation/validation records. **Accepted with the
recorded grader exceptions after one correction round.** The sole initial
blocker is resolved; no remaining implementation, architecture or scope finding.

A separate settings-order mutex serializes patches and subscription. The GUI
observation mutex now protects only the brief snapshot/publication operations,
not disk I/O or the actor acknowledgment. Lock order is settings-order then
observation where both are needed; observation never takes the settings-order
lock. The actor remains free to publish progress before processing the settings
message. A subscription still waits for an in-flight patch before its snapshot,
and subsequent patches cannot overtake broadcasts. The correction also reports
actor-setting errors rather than silently announcing success. Core behavior,
including effective idle memory disabling, is unchanged.

The first paced restored-lock mutant **survived**, as honestly retained in the
student evidence. It does not establish regression sensitivity. The final test
uses a real additional display subscriber to gate the first delta before GUI
observation, then observes the actual settings file before releasing that delta.
With the submitted defective GUI restored, that file write proves the socket has
already entered the lock-held transaction; its subsequent actor wait and the
released GUI observer form the original cycle. I inspected the raw stack: idle
control passes, active delivery times out, and settings/observer/shutdown waits
match the original defect. This is a valid bounded liveness kill, not a setup
error. The public subscriber adds scheduling control without replacing the
Agent, its observer or any earlier runtime behavior; the provider fake remains
at the HTTP boundary. Release/stream cleanup is explicit, and the overall test
timeout bounds the defective case whose shutdown intentionally cannot complete.

Independent execution from an isolated archive of the final revision reran the
**unchanged original reviewer reproduction**, the new gated regression and
existing settings wire/disk test together under race detection. All three pass:
original probe 0.61s, settings wire/disk 0.01s, gated regression 0.03s; package
1.661s, exit0. Active updates reach both clients with effective Agent values,
fresh subscription sees the final value, streaming ends and Agent shutdown
returns. [Raw output](../../solutions/evidence/ch16/reviewer-round1.txt)
and [source/command provenance](../../solutions/evidence/ch16/reviewer-round1.json)
preserve these observations. The coder's full GUI tests, vet and GUI race pass
are separate affected-module evidence. No paid or original-grader repeat was
needed; no shared product source was modified by the reviewer.

The correction adds 15 physical/13 active/2 comment-only production Go lines and
133 physical/120 active/10 comment-only test lines. Thus final chapter production
delta is +915/+745/+124 and Go-test delta +835/+782/+35 under the initial review's
method; examples and other languages are unchanged by this round. The extra
mutex and local ordering comments directly address the demonstrated lifetime
problem. The larger test exercises a real cross-component interleaving, two-tab
publication and shutdown; it is not a generic concurrency framework. The new
exported test and observer method have useful purpose/lifetime documentation.

Acceptance does **not** revise the original **20/100 FAIL on intermediate
`ccd6350`** or turn its vacuous eight recovery points into proof. The narrow
cwd-versus-Agent settings fixture and inherited undefined-tool/parity exceptions
remain independently justified as recorded above, with valid masked obligations
supported by the real-runtime and replay evidence. The `summarizeSpan` departure
remains accepted for actual recorded replacement. Bill's latest-request input
ruling remains distinct from lifetime accounting. All five live runs and their
reason-loss, invented-question, recall-error and extra-tool limitations remain
historical outcomes; acceptance claims working mechanisms and the documented
quality limits, not perfect summarization or universal task success.

Reviewer harnesses are idle and source review is complete. Coordinator owns the
checkpoint and direct Edition2 `ch15` export. Original Chapter22/24 decision
gates and the stopped author phase are unchanged. Mandatory next handoff:
`.agents/skills/ensemble-coding/SKILL.md`.
