# Edition 3 student inputs

Read this with the shared [execution procedure](edition-3-execution.md), the complete mandatory `.agents/skills/ensemble-coding/SKILL.md`, and the assigned chapter before starting and after compaction. These current paths/phases replace historical edition instructions. Architecture, KISS, testing, documentation and review obligations remain.

## Current boundary

Chapters 1–19 are closed; Chapter 19 is commit `0bac5a47018e533e3d97d6e2820d67d3cc01036d`. Chapter 20 entry is accepted. Fresh `e3-restart-ch20-student` is running toward an early compiling HOLD. Its accepted predecessor is `solutions/edition-3/ch19`; the working tree is `solutions/edition-3/main`. The current public chapter requires reusable GUI assets/public imports and a terminal fleet with application-supplied construction. Reuse the existing root library, `Config`, public Agent, GUI module, Jobs and mailbox owners; avoid a second parallel runtime or configuration hierarchy.

Root aligned the historical grader to the separate core/CLI/GUI layout. The current route is `grade-edition3 -ch 20 -core CORE_DIR -gui GUI_TARGET [-json] CLI_TARGET`, with flags before the positional CLI. The CLI entry is `./cmd/ch01`, GUI entry `./cmd` in the GUI module. There is no inherited Chapter 20 grade or operational result. The fresh student is authorized only through an early compiling HOLD, focused local tests and affected-module compilation. Full suites, race/vet, grades, browser/native/provider work and commit remain root-owned later releases.

Bill authorizes live tests for the remaining book without routine permission requests. Root still uses finite inspected plans; account consent, installation/restart, devices and publication remain separate boundaries. The Chapter 19 correction-limit exception was chapter-specific; Chapter 20 starts at zero of the usual ten production correction rounds. Do not replay completed chapters. Earlier invocation sections below are historical/public guidance, not permission to reopen them.

## Permitted inputs

- The assigned `book/edition-3/chapter-NN.md`, Chapter 00 for orientation, and earlier chapters as needed.
- Your own working tree in `/Users/bill/projects/ensemble/solutions/edition-3/main` and the specific accepted predecessor snapshot supplied by the orchestrator under `solutions/edition-3/chNN`. Chapter 1 starts empty; other snapshots are not automatically permitted inputs.
- The current role, shared procedure, this public brief, and narrow public-contract clarifications explicitly supplied by the orchestrator. Mandatory governing instructions may contain incidental historical path/revision references; reading those instructions is allowed, but following those pointers into old solutions or repair evidence is not.
- Public dependency/provider documentation needed for the assigned exercise.

Apart from your assigned workspace and permitted predecessor, do not read `agent/`, other `solutions/` trees, `edition-1/`, future chapters, grader implementation, historical answer/repair notes, or their Git blobs to solve the exercise. Do not read the coordinator's readiness/carryover/evidence records unless a specific student-safe excerpt is supplied. The historical skill's direction to read Edition 2 carryover is replaced for the fresh Edition 3 student by this brief; authors/reviewers still read their complete governing documents. Do not follow answer-revealing links from permitted documents. These are input-separation rules, not a claim of OS sandboxing.

## Chapter boundary

Chapter 1 starts from an empty source tree. Later chapters extend your exact accepted predecessor; do not copy the full reference. The orchestrator supplies the assigned chapter revision, accepted predecessor and grader revision. The orchestrator owns snapshots and repository commits; do not create a nested repository or edit other main-repository files. Submit one concise author-facing chapter review, not an evidence bundle. No later capabilities early. Implement required behavior with the simplest adequate design, with useful explanations of non-obvious choices.

Run the assigned chapter grader with an explicit positional target. Chapter 1 accepts a supported solution directory or a built CLI executable. If your public root is a library with a nested CLI, follow the chapter's “Grade yourself” build instructions rather than grading the library directory. For the executable, from `/Users/bill/projects/ensemble` (the repository-root Go module, not `agent/`):

```sh
go run ./cmd/grade-edition3 -ch 1 -json /absolute/path/to/your/student-cli
```

Substitute the executable you actually built; its package layout is your implementation choice.

For Chapter 2, follow its revised exercise and run both existing graders against
that same built executable from the book repository root:

```sh
go run ./cmd/grade-edition2 -ch 2 -json /absolute/path/to/your/student-cli
go run ./cmd/grade-edition3 -ch 1 -json /absolute/path/to/your/student-cli
```

Both must pass and exit successfully. The second command preserves inherited
Chapter 1 obligations; it does not replace the Chapter 2 pass.

For Chapter 3, build the actual CLI package and use the same executable for:

```sh
go run ./cmd/grade-edition2 -ch 3 -json /absolute/path/to/your/student-cli
go run ./cmd/grade-edition3 -ch 1 -json /absolute/path/to/your/student-cli
```

Both must pass. The Chapter 3 grader includes Chapter 2 parity; the second
command retains the inherited lifetime checks. Follow the chapter’s public-API
tests and source review for the selected-model explanation policy: grader
scores alone do not establish it. The chapter leaves the policy-selection
interface to the implementation; no hidden activation flag is required.

For Chapter 4, the Edition 3 command exposes the chapter's existing grading
checks with separate executable, core-source and external-consumer identities.
Build your actual CLI; do not turn a public library root into `main` to suit a
builder. From the book repository root run:

```sh
go run ./cmd/grade-edition3 -ch 4 -json -core /absolute/path/to/your/core-module -consumer /absolute/path/to/your/consumer-module /absolute/path/to/your/student-cli
go run ./cmd/grade-edition3 -ch 1 -json /absolute/path/to/your/student-cli
```

For a consumer other than the default `shout` example, add
`-consumer-case /absolute/path/to/case.json` before the executable, using the
chapter's public case schema. The first report retains the chapter's 160-point
jobs/inherited/consumer checks; the second retains the lifetime obligation on
that same executable. Both must pass. The core path must be the actual module
resolved by the consumer, not a similarly named copy. Required focused tests,
source ownership review, job lifecycle/cursor/limit checks and inherited
selected-model policy evidence remain necessary beyond those reports.

For Chapter 5, build the actual CLI and keep the core, external consumer and
three-Agent workflow as explicit module paths. Run from the book repository root:

```sh
go run ./cmd/grade-edition3 -ch 5 -json -core /absolute/path/to/your/core-module -consumer /absolute/path/to/your/consumer-module -workflow /absolute/path/to/your/workflow-module /absolute/path/to/your/student-cli
go run ./cmd/grade-edition3 -ch 1 -json /absolute/path/to/your/student-cli
```

Put optional `-consumer-case /absolute/case.json` and
`-media-models /absolute/media-map.json` before the executable. The media mapping
is a JSON object with exactly the nonempty `anthropic`, `openai`, and `gemini`
model IDs selecting your explicit capability rows; it selects grader inputs,
not live-provider capabilities. The case uses the chapter's public schema.
Both reports must pass on that same executable. The first retains Chapter 5's
seven checks and 100-point total, including Chapter 4 parity; the second
retains inherited lifetime checks. Public actor/reply/observer contracts,
interruption and job behavior, round-budget semantics, media loading, actual
workflow tool effects and human-interface exercises still require their own
focused tests and independent acceptance beyond those scores.

For Chapter 6, build the actual CLI and identify the core, inherited consumer,
workflow, and separate public-API streaming exercise modules explicitly. From
the book repository root run:

```sh
go run ./cmd/grade-edition3 -ch 6 -json -core /absolute/path/to/your/core-module -consumer /absolute/path/to/your/consumer-module -workflow /absolute/path/to/your/workflow-module -stream-exercise /absolute/path/to/your/stream-exercise-module -stream-model YOUR_DECLARED_STREAMING_MODEL /absolute/path/to/your/student-cli
go run ./cmd/grade-edition3 -ch 1 -json /absolute/path/to/your/student-cli
```

The streaming model is required and must select an explicit capability row in
your core; the adapter does not invent one. Put optional `-consumer-case` and
`-media-models` JSON paths before the executable as above. The first report
retains the existing Chapter 6 checks and Chapter 5 parity; the second retains
the lifetime obligation on the same binary. Both must pass and exit zero.
Actual arrival order, early human CLI delivery, SSE framing, source-bound replay,
and inherited job wake-up behavior still require the chapter's public-interface
tests, source review and independent acceptance beyond those reports.

For Chapter 7, keep the accepted core CLI and inherited modules separate from the new GUI module. Build the actual CLI target, then run from the book repository root:

```sh
go run ./cmd/grade-edition3 -ch 7 -json \
  -core /absolute/path/to/core-module \
  -consumer /absolute/path/to/consumer-module \
  -workflow /absolute/path/to/workflow-module \
  -stream-exercise /absolute/path/to/stream-exercise-module \
  -stream-model YOUR_DECLARED_STREAMING_MODEL \
  -gui /absolute/path/to/gui-module \
  -gui-model YOUR_DECLARED_GUI_MODEL \
  /absolute/path/to/student-cli
go run ./cmd/grade-edition3 -ch 1 -json /absolute/path/to/SAME/student-cli
```

All five module roots and both model identifiers are explicit; flags precede the executable. Optional `-consumer-case` and `-media-models` paths are forwarded unchanged to the inherited checks. Use models declared by your implementation, not historical defaults or invented feature rows. A passing WebSocket grade does not substitute for the chapter's actual browser, native-speech, replay, and pause exercises. Run isolated servers and browser contexts without disrupting the orchestrator's live GUI. The same frozen core CLI must pass Chapter 7 and Chapter 1, including the real delayed-response phases.

For Chapter 8, use the same separate executable and five module identities. After building the actual CLI target, run from the book repository root:

```sh
go run ./cmd/grade-edition3 -ch 8 -json \
  -core /absolute/path/to/core-module \
  -consumer /absolute/path/to/consumer-module \
  -workflow /absolute/path/to/workflow-module \
  -stream-exercise /absolute/path/to/stream-exercise-module \
  -stream-model YOUR_DECLARED_STREAMING_MODEL \
  -gui /absolute/path/to/gui-module \
  -gui-model YOUR_DECLARED_GUI_MODEL \
  /absolute/path/to/student-cli
go run ./cmd/grade-edition3 -ch 1 -json /absolute/path/to/SAME/student-cli
```

Put optional `-consumer-case` and `-media-models` paths before the executable as above. Both reports must pass and exit zero on the same frozen CLI, including the real delayed-header/body phases. Settings snapshots, restart persistence, and Chapter 7 parity do not replace Chapter 8's public source/integration obligations: effective next-turn budget use, validation/save ordering, and actual browser controls and speech-rate propagation still need evidence. No particular budget-setting API is prescribed.

For Chapter 9, use the accepted Chapter 8 predecessor and the same separate executable and module identities. After building your actual CLI target, run from the book repository root:

```sh
go run ./cmd/grade-edition3 -ch 9 -json \
  -core /absolute/path/to/core-module \
  -consumer /absolute/path/to/consumer-module \
  -workflow /absolute/path/to/workflow-module \
  -stream-exercise /absolute/path/to/stream-exercise-module \
  -stream-model YOUR_DECLARED_STREAMING_MODEL \
  -gui /absolute/path/to/gui-module \
  -gui-model YOUR_DECLARED_GUI_MODEL \
  /absolute/path/to/student-cli
go run ./cmd/grade-edition3 -ch 1 -json /absolute/path/to/SAME/student-cli
```

Optional `-consumer-case` and `-media-models` paths precede the executable. Declare models supported by your implementation rather than adding feature rows to satisfy historical defaults. Both reports must pass and exit zero on the same frozen CLI, including the real delayed-header/body phases. The skills and inherited settings checks do not establish every public obligation: full-graph validation before activation, failed-load atomicity, Agent isolation, variable expansion outside the Agent mutex, and durable enabling-event provenance also require source and integration evidence. `unload_skill` records pending removal only in this chapter; it does not implement future compaction. Preserve the chapter's real GUI/manual exercise without expanding it into a provider campaign.

For Chapter 10, extend the accepted Chapter 9 predecessor. Build your actual CLI target, then use the same explicit artifact identities from the book repository root:

```sh
go run ./cmd/grade-edition3 -ch 10 -json \
  -core /absolute/path/to/core-module \
  -consumer /absolute/path/to/consumer-module \
  -workflow /absolute/path/to/workflow-module \
  -stream-exercise /absolute/path/to/stream-exercise-module \
  -stream-model YOUR_DECLARED_STREAMING_MODEL \
  -gui /absolute/path/to/gui-module \
  -gui-model YOUR_DECLARED_GUI_MODEL \
  /absolute/path/to/student-cli
go run ./cmd/grade-edition3 -ch 1 -json /absolute/path/to/SAME/student-cli
```

Optional `-consumer-case` and `-media-models` paths go before the executable. Both reports must pass and exit zero on the same frozen CLI, allowing the actual delayed-header and delayed-body lifetime phases to finish. The persistence checks and inherited Chapter 9 parity do not establish every public obligation: rich replay through all three vendor renderers, atomic load/refusal, effective-running configuration boundaries, and stable GUI addresses across restart—including already-colliding legacy archives—also require source and integration evidence. Preserve provider payloads and parser-local IDs; no new public identity API or future compaction implementation is prescribed. Use owned save paths for probes and the chapter's real CLI/GUI restart exercise, without importing private histories or expanding it into a multi-provider paid campaign.

For Chapter 11, extend accepted Chapter 10 and use the same command and separate artifact/model flags above with `-ch 11`, followed by `-ch 1` on the **same frozen CLI**. Optional case/media flags still precede the executable. Require both parsed pass and exit zero, preserving the genuine delayed-header/body waits. The MCP checks and inherited persistence parity are finite: review every taught transport, request correlation, declared-skill startup and tool filtering, lifecycle/cancellation, and fresh generic automatic context (including valid empty refresh). `gui_snapshot` and `tts_queue` remain on-demand tools. The browser endpoint must be enabled by a declared skill, not an empty-skill fallback. Complete the chapter’s bounded real-browser exercise in an isolated owned session, without a provider campaign or live-GUI disruption.

For Chapter 12, extend accepted Chapter 11 and use the same explicit artifact/model flags above with `-ch 12`, followed by `-ch 1` on the **same frozen CLI**. The Chapter 12 route checks startup skill selection, explicitly requested snapshots and speech queues, real click/input effects, unload refusal/peer teardown, and Chapter 11 parity. Browser observations must not be injected automatically; legitimate tool-result history may remain after unload. The fake peer does not replace the bounded real-browser exercise. Preserve Chapter 9's dynamic SYSTEM instructions and provider-specific disclosure rules. Follow the phase authorization: a compiling hold and source review precede full gates and operational work.

The Edition 3 command requires an explicit target and currently supports Chapters 1, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16 and 18. Its Chapter 1 mode preserves the inherited conversation checks and adds delayed-header/body lifetime gates. Check the report's `passed` field and exit status, not score alone: required zero-point checks can fail a 100/100 score. Those real delayed probes take about two minutes, so a 90-second suite timeout is invalid. Omitting the path from the older Edition 2 command silently targets an old solution and is not a valid student result. A grader result does not replace required live exercises, test-sensitivity evidence or independent acceptance. Credentials remain in the authorized provider loader/test process, never prompts, command arguments, evidence or commits.

Submit one concise chapter review identifying the chapter/code/grader revisions, what you built, actual tests and live observations, failures, ambiguity, assistance and useful teaching clarifications. You may explain a suspected grader defect; you may not silently redefine success or edit the grader. The orchestrator supplies any missing public contract without old answer code. Follow the shared bounded review loop.


Chapter 13 uses the same explicit core/consumer/workflow/stream/GUI identities and flags-first ordering as Chapter 12, with `-ch 13`. For the accepted nested layout, build the CLI from `./cmd/ch01`; use `-workflow /absolute/student/ch06` and `-stream-exercise /absolute/student/ch07`, not the reverse. Supply `-core /absolute/student`, `-consumer /absolute/student/examples/ch04-consumer`, `-gui /absolute/student/gui`, and explicit `-stream-model` and `-gui-model` before the executable. The speech harness is discovered under the core directory (prefer `scripts/tts-harness.sh` or `tts-harness.sh`) and must honor the chapter's environment and JSONL contract while driving the real GUI. Run from the book repository with `go run ./cmd/grade-edition3 -ch 13 -json ... /absolute/student-cli`, then Chapter 1 on the SAME frozen CLI after Chapter 13 passes. Both require `passed: true` and process exit zero; allow the actual 62-second header and body checks to finish. The grader's browser/synth boundary is controlled evidence, not native hearing or the blind listener exercise. Operational execution requires the root's bounded release.


Chapter 14 directly grades the **GUI executable or GUI module directory**, not the CLI or core library. Build and freeze the GUI and CLI from the same held source, then run from the book repository:

```sh
go run ./cmd/grade-edition3 -ch 14 -json /absolute/owned-output/ch14-gui
go run ./cmd/grade-edition3 -ch 1 -json /absolute/owned-output/ch14-cli
```

For the accepted layout, build `ch14-cli` from `main` with `go build -o /absolute/owned-output/ch14-cli ./cmd/ch01`, and `ch14-gui` from `main/gui` with `go build -o /absolute/owned-output/ch14-gui ./cmd`. Chapter 14 rejects unrelated `-core`, consumer, workflow, stream, media and GUI/model flags; its fixture owns its models and assets. Require parsed pass and process exit zero for both reports, allowing the genuine Chapter 1 62-second header and body phases. Follow phase authorization: a compiling HOLD and architectural review come before these gates. The finite grade does not replace writer-ordering, failed-save/diagnostic preservation, backup-tail overlap, total-reducer, provider-disclosure and scoped operational evidence required by the public chapter. No numeric canonical-event retention cap replaces the validated-BACKUP boundary.

Chapter 15 also grades the **GUI executable or GUI module** directly. Its cache
artifact paths default to `cr/cache-current.txt` and `cr/cache-prior.txt` in the
owned launch directory; override with `-cache-current`/`-cache-prior` only to
match the implementation's actual path. Supply four independently checked
per-million prices with `-cache-prices INPUT,WRITE,READ,OUTPUT`; the adapter
requires them rather than deriving expected costs from student code. Other
chapters reject these flags. For the dated local fixture tariff, from the book
root after building held GUI/CLI artifacts in their own modules:

```sh
go run ./cmd/grade-edition3 -ch 15 -json -cache-prices 3,3.75,0.3,15 /absolute/output/ch15-gui
go run ./cmd/grade-edition3 -ch 1 -json /absolute/output/ch15-cli
```

Preserve all Chapter 14 instruction, recovery, and writer-boundary behavior.
The Chapter 15 grade is finite fake-provider evidence, not a real cache grant,
browser inspection, price verification, or architecture acceptance. Follow the
current phase authorization before running it or paid exercises.

Chapter 16 grades the **GUI executable or GUI module** directly; it takes no
core/parity/model/media/cache flags. From the book repository root, after the
held source is reviewed and the validation phase is released:

```sh
go run ./cmd/grade-edition3 -ch 16 -json /absolute/output/ch16-gui
go run ./cmd/grade-edition3 -ch 1 -json /absolute/output/ch16-cli
```

Build the CLI from accepted `main` with `go build -o /absolute/output/ch16-cli
./cmd/ch01`, and the separate GUI from `main/gui` with `go build -o
/absolute/output/ch16-gui ./cmd`. Both command entrypoints delegate auth commands
to the same owner; the offline GUI-direct exercise invokes `auth login
--print-url` on its supplied executable. The public chapter declares the
synthetic capability row and local override names. Require parsed pass and
exit zero for both reports, with at least 360 seconds allowed for Chapter 1's
real 62-second header and body checks. Keep capability routing conservative;
provider labels alone do not prove Responses, summary or cache support.
The fake OAuth/Responses grade is not personal-account authorization, proof
of live entitlement, native speech, or complete security coverage.


## Chapter 18 restart invocation

Chapter 17 is a field report, so extend accepted Chapter 16 in `main`;
`../ch16` is the permitted predecessor. Build the actual nested CLI and
separate GUI, without changing the public library layout:

```sh
(cd solutions/edition-3/main && go build -o /private/tmp/ensemble-ch18 ./cmd/ch01)
(cd solutions/edition-3/main/gui && go build -o /private/tmp/ensemble-ch18-gui ./cmd)
go run ./cmd/grade-edition3 -ch 18 -json \
  -core "$PWD/solutions/edition-3/main" -gui /private/tmp/ensemble-ch18-gui \
  -status-prices 3,3.75,0.3,15 /private/tmp/ensemble-ch18
go run ./cmd/grade-edition3 -ch 1 -json /private/tmp/ensemble-ch18
```

All flags precede the explicit CLI target. Supply the four independent
input/write/read/output rates per million tokens. The chapter documents
these dated Sonnet 4.6 fixture rates and the optional Opus 4.6 tariff
`-status-second-prices 5,6.25,0.5,25`; omitting the latter selects the
unpriced middle-model control. Both reports must have `passed:true` and
exit zero. Allow at least 360 seconds for Chapter 1. The coordinator
releases these gates only after the coherent compiling HOLD and review.
Source review still follows real owners and every shipped composition
root; numerical grading alone is insufficient. Existing `ToolCall`,
`Parent`, and `Status` seams are valid without aliases for reference names.
Live account consent, provider requests, and hosted web operations require
their own explicit authorization; historical examples do not supply it.


## Restarted Chapter 19 public entry brief

Start from accepted Chapter 18 in `solutions/edition-3/main` after root release.
Read the complete current Chapter 19 and required public prerequisites. Extend
existing Config/common.Config with policy and Child/Clamp; preserve the root
library and nested CLI/GUI owners. Sandbox retains its required Agent parent.
Read actual file, process, job-output, state, media, transport and render owners
rather than add a second security or credential framework around them.

Implement public confinement, safe mode, the web-skill gate, environment and
redaction behavior with positive useful-work cases as well as refusals. Share
CLI/GUI policy composition. Public commands use `./cmd/ch01` and flags before
`chat`. Account consent and host/provider demonstrations need separate release;
historical prose is not new operational evidence.

First stop at an early compiling HOLD, not acceptance. Authorized work upon
root release is implementation, focused pure/local tests and affected-module
compilation, sequential/offline with owned HOME/TMP and existing caches. No
full suites, race/vet, grades, Seatbelt/host scenarios, browser/native/provider
or live OAuth work yet. Save actual changed owners, commands/exits, setup
failures and unfinished work concisely. Root and the retained reviewer inspect
held source before broader gates. Initial implementation is round zero; up to
ten production correction rounds are available. Originals, private graders and
reviews, runtime and root context remain outside student inputs.


## Restarted Chapter 20: early HOLD and correction round 1

Entry is accepted. The same student has submitted a stable early compiling HOLD;
that is not implementation acceptance. Preserve its saved failure evidence and
extend the held source, not a fresh Chapter 19 copy. The same independent reviewer
is retained. The root has read and joined the early-HOLD review, independently
inspected its ownership findings, and releases this consolidated round 1. Do not
repeat entry or predecessor checks.

Current public Chapter 20 governs over restored instructions. Its constructor is
`gui.New(root, agent, log io.Writer) (*Server, error)`; `Assets`,
`Server.Handler(directory)` and `Server.ServeMCP(net.Listener)` are the reusable
API. Only delegated Agents owe a terminal submit and the three corrective wakes;
normal root/peer chat completion remains valid. Explicit root submit is available
under the same validation/admission rules. Earlier root shorthand saying
`Hub.ServeMCP` and restricting submit to delegated Agents was withdrawn.

Keep the existing Ensemble registry, Agent actor and Server/Hub owners. The review
does not ask for a replacement orchestration or GUI framework. Complete one
coherent bounded revision:

- Separate selected-Agent embedding from application-wide GUI policy. Constructing
  a selected view must not rewrite that Agent's policy, affect other Agents, or
  expose their traffic/command targets. Closing the view does not stop the Agent.
  Preserve legitimate application fleet management. Exercise two distinct Agents
  and policies, not only a single-Agent rendering test.
- Fix actual application startup with the default empty GUI-directory override:
  serve embedded assets without requiring index/app/style files in the working
  directory. Keep explicit disk overrides working. Complete the promised reusable
  rendering modules and their actual application imports/consumers, not dummy
  assets added only to make an existence check pass.
- Make factory admission and failure cleanup truthful. Check returned identity,
  parent relationship and contract, recheck cancellation/parent eligibility before
  task activation, and retire any newly owned construction left behind on refusal.
  Never shut down a pre-existing or foreign Agent returned by a mistaken factory.
  Preserve useful terminal records and the existing safe construction boundary;
  do not solve this with a new transaction framework or a lock held across actor
  shutdown. Preserve an independent expected schema when invoking the factory;
  mutable factory arguments must not redefine the contract being checked. Admit
  identity/recovery-path collisions before loading or writing shared recovery
  state or initiating capability effects. Test factory failure/mismatch and
  parent-stop/cancellation boundaries. Clarification during round 1: this is not
  universal rollback of arbitrary application-factory effects. The factory owns
  its intended construction until admission and cleans it on post-construction
  error, or explicitly returns it with the error for ownership-checked cleanup.
  Extra resources/Agents remain the factory's responsibility. Never infer ownership
  from a global registry before/after diff. Prefer the requested-child admission
  boundary and exact identity; use a small attempt handle only if needed, not a
  variadic-context constructor or callback serialization solely to promise universal
  rollback. Concurrent unrelated construction must remain untouched.
- Close future actor admission as part of successful terminal submit, before
  publishing its reply. Ordinary finish currently returns to Idle, allowing
  queued work to start before the fleet supervisor observes the result and stops
  the Agent. A queued prompt must never start another provider request after
  accepted submit. Preserve ordinary root chat and truthful failure handling;
  durable payload/save and final outcome ordering remain essential.
- When child model selection changes, clear/re-resolve inherited model identity
  and dependent policy through existing normalization owners. Verify actual child
  wire identity and relevant capability/accounting identity; changing only the
  displayed Config.Model is insufficient. Do not invent provider aliases or prices.
- Finish the declared remaining public contract: bounded complete schema admission
  and validation (including exact numbers and distinct default/custom payload
  handling); durable submit.md and fallback before successful terminal publication;
  failed-child resolution; exactly three delegated corrective wakes; root explicit
  submit; correlated pending-question visibility and yielding joins; the parentless
  human question sink. Keep waits observational and jobs/lifetimes in existing
  owners. Tests must distinguish idle from submitted/failed/stopped and preserve
  useful failure outcomes. No extra schema language or speculative scheduling.

Upon explicit root release, this is production correction round 1 of at most 10.
Authorized checks remain focused pure/local-loopback regressions and affected
module compilation, sequential/offline with owned HOME/TMP, existing caches and
dummy credentials. No full suites/race/vet/grader campaign, native sandbox or
provider/browser/OAuth operations, extra children, snapshot, prose, commit,
installation, restart or effects on the running GUI. Submit a new stable HOLD with
actual checks, failure evidence and any remaining limitations. Root and the same
reviewer decide when broader validation is warranted.

### Chapter 20 correction round 2/10 — released after R1 HOLD

R1 is collected/joined and independently reviewed, not accepted. Keep the existing
Ensemble registry, Agent actor, factory admission record, exact-arithmetic validator,
and Server/Hub projection. R1's focused and compile checks remain valid for what they
actually exercised; no full campaign or operations have run. Correct these groups
coherently, then return frozen source and a saved HOLD to the same reviewer:

1. **Terminal failure admission.** Successful submit now gates queued work, but a
   delegated error or exhaustion of three corrective wakes still replies/returns
   Idle before its watcher shuts down the actor. Close future admission and settle
   queued/follower requests through the existing actor terminal owner before a
   delegated lifetime failure is exposed. Preserve ordinary root/peer chat errors,
   joining cleanup, truthful failed outcomes and final persistence. Add controlled
   queued-input checks for failure/exhaustion, not just successful submit.
2. **Managed children versus retained outcomes.** Explicit shutdown must remove a
   child from managed enumeration/status while preserving saved work and outcome
   retrieval. The existing resolved check already permits parent submit after
   explicitly resolving a failed child; preserve it. Use the existing entry owner,
   not a second registry. Check recursive shutdown and retained results.
3. **Public fleet and reply interfaces.** Advertise/accept public spawn `task`
   (internally mapping to Prompt is enough). Complete addressed parent/child/human
   delivery and correlated reply envelopes, including a usable sender/return address
   and configurable human address. Do not restrict the public send operation to
   direct children only. Implement the chapter's CLI `/reply TEXT` when exactly one
   human question is pending and `/reply MESSAGE_ID TEXT` when disambiguation is
   required; ambiguous/stale replies must not become ordinary prompts or answer the
   wrong question. Existing `/answer` may remain as an optional compatibility path,
   not the sole supported route. Provide a public terminal string convenience that
   normalizes the default result and enters the same actor-owned durable completion
   path; do not repurpose inherited chat Submit or create a second completion owner.
   Check real tool metadata/dispatch, human queue correlation, and convenience
   completion, including failed admission that leaves work active.
4. **Selected GUI scope.** Construction-time policy mutation is fixed, but gui.New
   still observes/replays every Agent and accepts arbitrary registered Agent targets.
   Scope initial replay, live observations, commands and relevant browser MCP routing
   to the attached Agent. Keep deliberate whole-fleet application behavior available
   through the existing projection machinery without duplicating it. Check existing
   and subsequently created unrelated Agents, histories, control attempts and close.
5. **Reusable artifact component.** Current createArtifactScroll is scroll-follow
   behavior, not the required independent ArtifactScroll.receive component. Finish
   the public artifact-scroll.js and renderers.js interface over the real extracted
   rendering functions; main app must consume the same component, not duplicate it
   or use empty compatibility files. Preserve streaming/final updates, history/tool
   artifacts, fold state, speech controls and scroll behavior. Test a small standalone
   DOM consumer without app.js, WebSocket or global application controls, alongside
   the affected main-app JavaScript behavior. Browser operations are not yet released.
6. **Exact public admission boundaries.** Schema/payload depth is maximum64 with
   root at zero; schema is not limited to32. Numeric literals max128 bytes, exponent
   magnitude max1000. Retain published individual incoming/canonical byte and node
   limits. The combined refusal boundary is canonical schema bytes × canonical
   payload bytes <=67108864, not nodes × schemaNodes <=1000000. Remove the substituted
   undocumented node-product refusal. Check at/over boundaries with valid schema
   and payload data, exact numbers, and no coercion/truncation. Keep the existing
   real recursive validator; no library/framework replacement is requested.

Non-findings: named Join IDs already reject unknown/unrelated children. No null-keyword
counterexample was established. ParentID navigation through the existing Ensemble
owner is adequate; do not replace inherited Parent() merely to match an illustrative
signature. No new relay defect was demonstrated. Factory rollback remains limited to
positively owned intended construction per R1 clarification, never registry snapshots.

Scope: student production only in solutions/edition-3/main plus its existing public
note. Focused unit/actual-Agent loopback/JavaScript checks and affected compilation
only, using the established offline controlled environment. No full/race/vet/grader,
native/provider/OAuth/browser campaign, other children, snapshot/prose/commit or
installation/restart. Preserve failures and distinguish setup failures from defects.
Record actual commands, results and remaining limitations; stop production at HOLD.

### Chapter 20 correction round 3/10 — public browser module closure

R2 HOLD collected/joined and independently reviewed. Its five non-renderer correction
groups are addressed in source/receipt review, not yet full acceptance. The real
ArtifactScroll.receive component and renderer registry are implemented and consumed
by app.js; preserve them. Sole verified blocker: public Chapter20 TL;DR and
component-only recipe mount ONLY artifact-scroll.js and renderers.js from Assets.
Current artifact-scroll imports part-renderer, which imports tool-result, while
renderers.js is absent. That documented consumer cannot load.

Expose the actual single renderer implementation in renderers.js and make the two
advertised files a closed usable module graph. Move/rename/consolidate existing small
helpers as appropriate; no duplicate renderer, empty alias, bundler or framework.
Do not change the public recipe to require extra files. Preserve existing rendering,
custom overrides, streaming/replay/job/tool behavior and main-app consumption.

Verify with an independent minimal consumer that has ONLY those two source modules
available, then the focused affected component/main JS checks and GUI compile check.
Preserve old receipts; carry unchanged core/CLI/schema results explicitly rather than
rerunning them. Same offline controlled environment and same source ownership.
No browser/native/provider/OAuth, full/race/vet/grader campaign, child delegation,
snapshot/prose/commit/install/restart. Save actual checks and limitations in the
existing note and return stable compiling HOLD, then stop production. Same reviewer.

No additional selected-human-question leak exists: broadcast already filters
Question.Sender. Do not implement a fix for that withdrawn reviewer claim.

### After R3: inherited fixture alignment only (not production round 4)

Root core race gate stopped at /private/tmp/ch20-r3-validation.0IWZvx/core-race.log
(exit1); later GUI/vet/modules/grader unrun. Root+retained reviewer inspected actual
failures and current contract. Production remains HELD. Authorize only test edits:

- Ch10R1DefaultAgentPaths: retain two unnamed Agents, omitted SavePath, distinct
  derived-save and actual writable-file assertions. Give each an independent
  caller-owned DataDir. New actual DataDir/submit.md cannot be shared; collision
  admission remains unchanged. Do not automatically repartition production paths.
- Ch3 LoopOrderedBatchProviders, GeminiParameterProjection, Round1Registry: update
  independently enumerated expected default names/count from14 to current21 including
  seven fleet/root-submit tools. Preserve duplicate/name/schema/description, exact
  all-provider call/results/order, empty tools=0, selected tools, per-Agent isolation,
  Gemini exact projection/container identity, no common-schema mutation and other
  vendor bytes. No >= tests or expectations copied blindly from the registry.
- Ch4 SetupFailureCorrelation and AgentWireAndKilledObservation clone live Config
  with managed ID/DataDir. Construct fixture once after local URL exists, or reset ID
  and supply independent DataDir with distinct/default SavePath. Preserve task BaseDir
  and file inputs. Failure obstructions must follow actual effective DataDir/io after
  construction. Preserve exact IDs/results/pairs, no-side-effect assertions, admission
  shutdown injection, six-request sequence, PTY/send_input output, killed-job reason,
  no invented exit metadata and output-byte checks. No skip or production relaxation.

Run only these focused tests under race using existing isolated dummy-key wrapper;
retain failed output and new receipts. Stop/ask if a behavioral/product failure remains.
Save note/return test-only HOLD. No production changes, broad suite, GUI/browser,
grade/provider/native/OAuth, commits, children or installs. Root will inspect diff and
retained reviewer will check alignment before resuming the stopped broader gate.

### After R3: GUI fixture alignment only (still not production round 4)

Root core-race-v2 passed at /private/tmp/ch20-r3-validation.0IWZvx. GUI full race
stopped (gui-race.log/.exit1); later gates unrun. Root+retained reviewer diagnosed
three fixture issues. Production remains held. Authorize only these test repairs:

1. client_round1_test.cjs evaluates now-modular app.js as classic script. Load ACTUAL
   current app/artifact-scroll/renderers source in the existing controlled DOM realm.
   Keep fixture API and ALL inherited assertions; no fake ArtifactScroll, old copied
   renderer, general bundler or broad async-suite rewrite. A small known-graph source
   adapter is adequate; assert recognized import/export forms so unexpected future
   dependencies fail, not silently vanish. Preserve separate real two-file module
   checks (these already pass); this fixture adaptation is not browser evidence.
2. ch13_submit_test.go reads exec stderr bytes.Buffer.String before Wait joins copier;
   all three reported races share this test-owned path. Reuse existing safeBuffer in
   gui/round1_test.go (synchronized Write/String). Preserve decode/EOF and process-exit
   failures, useful stderr, input close and Wait. Do not hide underlying import errors.
3. Ch10 SnapshotPrefixTailWS context_only: initial dialogue-only INPUT is intentional;
   full-fidelity OUTPUT was previously obtained through GUI's unwanted silent policy
   override. Keep initial loaded boundary/mode/no-tail assertions; assert GUI attach
   does NOT change that mode. Then explicitly SetPersistencePolicy(LogFull, current
   diagnostic limit) BEFORE appending live tail/job and BEFORE Ask/autosaves. This
   chooses full OUTPUT; don't restore production GUI mutation. Keep every exact
   call/result/job count/association/display-ID/reconnect/reopen and canonical count
   assertion unchanged. No weakened jobs, accepted-ID-changes, or archive tail cheats.

Run focused affected GUI Go tests under race (including all actual failed names) and
shared JS controlled suite using existing isolated dummy-key environment. No production
edits. If any behavioral failure remains, stop/report for root diagnosis. Preserve old
logs, save new note/receipts and return fixture-only HOLD. No full gate, browser/native/
provider/OAuth, graders, child delegation, commits/prose/install/restart. This test-only
repair consumes no production round. Reviewer/root inspect before resuming GUI gate.

### GUI fixture continuation: explicit Continue versus new human typing

The test-only repair stopped correctly at the unchanged standalone assertion
`cleared typing must not latch subsequent queue`; receipt
`/private/tmp/e3-ch20-student.2X1OuS/after-r3-gui-shared-js-01.txt` remains FAILED.
Root source review confirms the speech/typing owner is unchanged from accepted
Chapter19. This old sequence supplies a state acknowledgement (not explicit
Continue), then new isTrusted:true input, yet expects speech idle to release it.
Current Chapter13 requires retaining genuine human typing and forbids delayed
state acknowledgements from clearing newer input. No production fix is authorized.

Continue the same test-only release by correcting this contradictory sequence:
- Drive real local Continue through its control (or accurate separate external
  continued authority), then its state acknowledgement. Preserve stale callback
  and old-epoch claim refusal: no revived audio or new pause from stale work.
- With NO new human input, check the next speech queue's own pause/release edges
  using counts captured AFTER Continue, not some(unpause) satisfied by Continue.
- Separately test new trusted typing after Continue and a delayed acknowledgement:
  it must acquire/retain pause through stale callbacks and speech idle, until the
  human clears it or explicitly continues. Do not relabel trusted input as stale.
This explicitly revises the old contradictory assertion/sequence; it is NOT a
claim that all old assertions remain unchanged. Preserve their valid safety intent.

Keep the reviewed real-source loader, existing safeBuffer, and explicit full-output
persistence fixture changes. Run the bounded shared JS checks and previously named
focused GUI Go race failures from the preceding release. No production changes,
full gate, browser/native/provider/OAuth, graders, child delegation, commits or
installation. If another behavioral failure appears, stop and report it. Preserve
all failed receipts, save actual diff/results and return stable test-only HOLD.
Production remains R3/10; no production round consumed.

### Final shared-fixture content correction

The second stopped receipt `after-r3-gui-shared-js-02.txt` remains FAILED at
`full readable content must reach native callbacks`. Revised explicit-Continue and
new-trusted-typing cases pass. Root and retained reviewer verified unchanged
speechChunks: it splits this single-space input at word boundaries, trimming the
separator between utterances. Empty-string joining therefore merges three pairs
of otherwise intact words. All450 words, emoji and final suffix remain present.
For THIS single-paragraph single-space fixture, replace only that join('') with
join(' '), retaining exact equality against the independently declared original
text. No general whitespace normalization, substring/token-count-only assertion,
or implementation-derived expectation. Keep callback drain, exactly one done,
queue ownership and all following tests unchanged. Remaining shared cases were
source-reviewed; no other demonstrated drift was found, but they remain unexecuted.

Continue only the previously scoped shared JS (including chapter12 mode) and
focused GUI Go race checks. Preserve both failures and stopped notes. Same stop
rule for any new unexplained behavior; production remains held R3, NOT R4.

## Chapter20 correction round4: application tools versus fleet completion

Root sequential local gate passed full core-race-v2, full gui-race-v2, both vets,
ch06/ch07 and outside ch04 consumer checks. It stopped in examples/ch13-speech:
TestListenerPeerEOFDuringAsk fails before HTTP with `tool "submit" already registered`.
This is a production integration defect, not a fixture correction. Current example
and accepted predecessor explicitly use Tools: []string{} then register exactly
hear/type/submit/wait/sleep. Its submit means send the real GUI composer, not finish
a fleet task. Preserve that example/API, the genuine restricted Agent and EOF test.

Fix the existing ownership boundary coherently:
1. An explicitly application-only composition (non-nil empty Tools, no skill
   selection, as in the listener) must not reserve the new optional fleet catalogue
   in a way that breaks those application tools. Resolve this at existing catalogue
   composition, not by broadly permitting shadowing disabled/compiled/MCP tools.
   Preserve default and skill-selected fleet composition, exact listener declarations
   and guessed-call refusal, no accidental shell/file/fleet grants. No second registry,
   new blanket override, or speculative configuration framework.
2. Fleet completion must follow the resolved fleet operation's ownership, not merely
   the tool string "submit". Currently dispatch validates/routes by name, actor report
   terminalizes by name, and recovery assumes any successful named submit needs a
   fleet submission row. An application tool must use its own schema and handler,
   remain nonterminal, and save/reopen normally without submit.md. Use the existing
   selected entry's ownership and a small explicit accepted-completion fact where
   needed across report/canonical recovery. Preserve legitimate Agent.Complete,
   exact fleet schema validation, durable-before-success, terminal admission and
   refusal of a genuinely accepted fleet completion with missing durable row.
   Do not weaken these by removing recovery validation or using current registration
   names alone to reinterpret historical application actions.
3. Document the narrow composition rule in the existing public API comment/README
   as appropriate. Keep ordinary root chat and all prior accepted behavior intact.

Focused evidence before HOLD: listener constructs with exactly its five tools and
its existing peer-EOF/HTTP test passes; actual-Agent application submit executes its
handler/schema, continues a subsequent turn and saves/reopens without fleet outcome;
default genuine fleet submit and Complete stay terminal/durable, missing-row recovery
still refuses; duplicate/installed compiled-name protection and no unintended grants
remain. Reuse existing tests/helpers, affected compilation and focused local/race
checks only. No full gate, graders, browser/native/provider/OAuth, new children,
snapshot, manuscript, commit, installation or restart. Preserve failures and changes,
stop for consequential uncertainty, save results in current student note and return
stable HOLD. Same retained reviewer; this is production round4/10.


### Chapter 20 — production correction round 5 of 10

R4 local gates passed, but public-contract reconciliation found three remaining
production gaps. Correct these together, retain existing owners, then HOLD.
1. Empty/default assets on the real GUI launcher still pass through unconditional
   disk validation in NewServerWithSettings and exit before embedded serving.
   Default unrelated/empty-cwd startup must use embedded assets; explicit disk
   override remains validated and wins. Test actual current command readiness,
   HTTP assets and normal EOF, not only the convenience constructor. No live8088.
2. §20.8 requires complete submitted JSON in status as well as join/retrieval.
   AgentStatus/Children/Status currently return only identities/state. Expose a
   defensive projection of the existing accepted outcome, preserving exact nested
   JSON and normalized default object, failed/running distinctions, and ownership.
   No separate result store or lifecycle owner; runtime field spelling is flexible.
3. Current TLDR105–113 explicitly requires custom durable rows containing
   version:1, submit_schema, payload. Existing v/custom/payload rows omit the
   immutable schema. Correct authoritative writer AND reader, preserving atomic
   append and whole-file validation, exact schema/payload numbers/types/owner,
   default normalization, refusal of unknown versions/malformed later rows, and
   terminal recovery ordering. Do not bless invalid rows using current config
   alone, drop schema/value, or weaken validated final-save-before-publication.

These are implementation/review misses, not grounds for a new architecture.
Root owns separate private grading protocol adapters; no student grader access.
Use focused actual Agent/status/recovery and real owned launcher checks, affected
race/compilation. Preserve R4 application-submit/fleet-submit identity and listener
contracts. No full campaigns, grader, browser/provider/native, children, snapshot,
prose, install/restart or commits. Record failures and limits; stop for a concrete
scope ambiguity. Same coder/reviewer; return stable saved note and HOLD.

### Chapter 20 — isolated synthetic reuse exercise after R5 (no production edits)

Root source review, retained review, fresh local suites, and current required grades are now clear. Production remains held at R5/10. Complete the public reuse/component exercise in an owned `/private/tmp` directory with an outside Go module importing the held main and GUI via replace directives. This is operational evidence, not another production correction. Do not read private graders, originals, runtime source, other chapter controllers, or credentials.

One finite local attempt, maximum 15 minutes and 16 synthetic model requests total. Existing Playwright may be imported from `/Users/bill/projects/ensemble/agent/gui/node_modules/playwright` strictly as an installed dependency; do not inspect runtime project source. Chrome is `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`. Use a new headless profile, owned loopback ports/listeners/processes only, TTS disabled, no external network, no OS focus/volume/microphone, no live 8088. Dummy keys and controlled synthetic provider only, no paid/provider/OAuth/native sandbox operations. No install/restart/commit or broad suites/grades. Preserve all failed attempts; setup-only repair allowed within this finite attempt, but stop and report any product defect or behavioral failure instead of editing production/retrying it.

1. Construct outside application using actual public Ensemble/Agent, gui.New, Handler, Assets and ServeMCP. Serve full shell from embedded assets with clean working directory; verify explicit owned disk override without replacing embedded default globally. Use immutable held sources and record binary identity.
2. Actual headless Chrome full-GUI interaction: conversation visible, turn delivered, real Agent/Engine request and response, browser MCP tool discovery/request/correlated reply end-to-end through real relay (not a fake socket standing in for browser); disconnect releases only its session. Use a safe DOM-observation tool, not speech/device/focus. Change selected model and, where supported, reasoning settings through the real GUI; next controlled request must show the selected model/reasoning, not only a stored field. Do not invent provider capability aliases. Keep explicit synthetic-versus-live limits.
3. Separate minimal HTML page serves only artifact-scroll.js/renderers.js from exported FS, imports ArtifactScroll, uses actual nested/flat event shapes observed from held source, renders text plus inert image/audio artifacts without adopting full shell. Assert safe DOM, readable mounted content, no extra source imports or missing requests. Check sticky latest behavior with real scrolling and explicit latest control if public component supplies one. Confirm the real rendered page via browser observations; screenshots saved but unseen do not prove visual beauty. No external media requests or audible playback.
4. Normal EOF / owned shutdown cleanup; record request counts, assertions, exact observed browser/relay/next-wire behavior and limitations in existing ch20-restart-student.md with owned receipts. Do not claim a live model exercise. Return stable HOLD with source still frozen; root/reviewer inspect before live fleet work.

### Chapter 20 — production correction round 6 of 10

R5 source/gates remain recorded; actual browser reuse exposed a real integration defect. Public served index loads `app.js` as classic/defer although that file imports ArtifactScroll. Chrome reports “Cannot use import statement outside a module”; Conversation replay never executes. Preserve `/private/tmp/ch20-r5-reuse.eSQXBY` failure and zero-request/normalEOF limits.

Make only the minimal correct module entry in the actual page. Keep existing app, component and renderer ownership. Add focused regression through public Handler-served HTML/assets so this cannot pass merely by transforming JS into classic-script test code. Retain all existing module behavior and tests. No production changes to speech/fleet/model/relay or unrelated owners.

Run focused affected Go/JS/servedasset checks and affected compilation; save exact diff/receipts and stable HOLD. Real-browser continuation after root+retainedreviewer frozen-source review, not silently within this release. No providers/live credentials, native operations, fullgrader/lifetime replay, installation/restart/8088, snapshot/prose/commit ornewworkers. Productionround6; root missinggrant/completion-parser mistakes are unrelated controllers, not student defects.

### Chapter 20 — post-R6 bounded synthetic browser continuation

R6 source has root+retainedreviewer PASS. Production stays frozen. Repeat the existing public reuse exercise in a NEW owned attempt directory, reusing valid consumer/controller source but rebuilding GUI-bearing artifacts so corrected embedded HTML is served; record identities. Preserve R5 failed attempt unchanged. Same finite15minute/16synthetic-request ceiling, no paidrequests/realkeys/native/device/8088 effects, TTSoff andownedloopbackprofile cleanup. No productionedit andstopatfirst unexplainedfailure; setuprepair only after diagnosis. No loader substitutions: actual browser modules and two-filepubliccomponent must load normally. Reach actualConversation/replay, publicbrowserTCP relay discovery/correlatedreply/close, selectedmodel+reasoning next-wire proof, andtwo-fileArtifactScroll actualreceive/scroll. Headless consumer+dependency and unchangedcomponent/source evidence may carry explicitly. No broadgate/graders/lifetimes repeat. Save actualevidence/limits andHOLD; root handles separatepaidfleetattempt. KeepR6 count6.

### Chapter 20 — final focused operational check, no production release

Browser R6 HOLD collected; preserve its actual evidence and limits. Live arithmetic task succeeded with19 realprovider requests, exact3children, real correlatedquestion/reply, actualterminaljoin, validchildsubmissions, independent18arithmetic assertions andnormalEOF. However firstproposedjoin was skipped beforehumanhint, so live did NOT prove hintduringexecutingjoin. Existing addressed-question test isnotthatproof.

Add ONE bounded local actual-Agent/realtool lifecycle regression using existing test/helpers (no newframework). Use syntheticHTTP and controlledchildresponse to establish join_agents actually dispatched and jobrunning BEFORE injecting ordinaryhumanhint. Show hint returns/wakes observation withoutcancellingchild orjob; then releasechild to genuine submit and collect the SAME joinhandle with correct terminaloutcome/payload. No sleeps assoleorderingproof, no directfabricatedsuccess, no claimliveproviderbehavior. Keep finite testdeadline and focusedrace run plusaffectedcompile only. Production frozen; if actualdefect emerges, STOP/report beforeeditingproduction. Save testdiff, orderingproof, exactreceipt andHOLD; preserve earlier browser/livefailures. This is testonly evidence completion, notR7. No paid/native/browser repetition, broadergate/grade/lifetimereplay, newworkers, snapshot/prose/commit/install/restart.

### Chapter 20 — closed release boundary

Frozen R6 code and required operations are accepted with the separate live and local timing qualifications recorded in the execution and final reviewer notes. The 282-file source/support snapshot is `solutions/edition-3/ch20`. Minimal prose completed one formal REQUIRED_REVISION and one bounded response; root directly verified both corrections. All Chapter20 worker contexts are closed. No further production or operational release remains for this chapter. Chapter21 requires its own current-contract entry assessment and public release.
