# Edition 2 Chapter 25: Codex-to-Ensemble crossover

This is the evolving source record for the crossover chapter, not finished prose
or chapter acceptance. Later entries supersede earlier deployment status. The
active implementation is now `agent/`; historical paths and pause instructions
below are retained as history, not current operating instructions.

## First-use account and the chapter's central story

After the file-configuration migration, Bill supplied essential context for the
chapter: he reports that the book had been published, Codex had shown him how to
launch Ensemble only a couple of hours earlier, and he had never tested Astra
inside Ensemble before this crossover. Preserve this as Bill's first-use account;
it is not an independently verified publication/upload event or a measured elapsed
time for every entry in this cumulative record.

Bill's formulation identifies the accomplishment: **“I gave Astra a book. Astra
gave me an AI coding agent.”** The book included graders. Astra implemented the
agent while working in Codex; Bill then began working with Astra inside the
resulting Ensemble implementation. Do not flatten this into the generic claim
that an AI helped write some software. Conversely, do not inflate it into a
verified historical first, unaided autonomous creation, or proof that the running
session inherited Codex's entire memory. The human-authored instruction, graders,
expert direction, testing and review are central to the account.

The crossover chapter now has concrete material, not just a launch recipe:

- **Moving into the result.** The handoff became a real working session. Bill's
  first requests came from actually using the GUI, including the nonlinear
  Eloquence speech-speed control. The initial staged-binary confusion is part of
  the lesson: building a replacement, installing it, restarting the server and
  refreshing the browser are distinct operations.
- **Making the partnership observable.** Bill repeatedly asked for explanations
  before meaningful actions, not silence followed by a large retrospective.
  He insisted that the Actions pane remain visible so he could inspect tool
  calls and returns. A GUI redesign must preserve that oversight, not optimize
  it away as clutter. The general visual-polish pass is not finished.
- **Separating roles and lifecycles.** Main-agent `submit` instructions and the
  runtime's corresponding wake-up rule belonged to delegated workers, not the
  human conversation. Real sub-agents appeared in the tree, reported results,
  and were joined/closed. Failed work must not be confused with successful idle
  completion; retained data and actual submission matter.
- **Choosing when to observe the browser.** Repeated spoken tool names exposed
  the automatic snapshot/queue collection. Replacement of ephemeral context was
  already correct. Bill chose ordinary on-demand tools, not a broad rewrite of
  ephemeral context. Live verification after a hard refresh confirmed both the
  absence of automatic attachments and successful explicit calls.
- **Following an error to the correct boundary.** The three-second callback was
  a check-in, not a deadline. The actual failure was a 60-second HTTP client
  deadline covering the entire model response. A response lasting 61 seconds
  passed with the fix and failed at 60 seconds with the original code; explicit
  interruption still worked. This is a particularly teachable contrast between
  observing a job and canceling it.
- **Turning hidden launch state into inspectable configuration.** The immediate
  question about reasoning exposed `high` effort with `auto`, not `detailed`,
  summaries. That led to the larger settings-file request. The delivered design
  makes the file authoritative, keeps credentials external, distinguishes active
  from saved restart-only settings, and preserves grader-facing entry points.
  Independent review caught a credential-as-flag-default leak before deployment;
  the regression test rejects the unsafe behavior. This was engineering with
  review and negative controls, not acceptance based on a successful demo alone.

The configuration implementation and crossover evidence are committed at
`ce6edb0e`; earlier runtime checkpoints include `3aad2b7e` and `09111e0b`. The
configuration deployment is installed but awaits Bill's restart, browser refresh
and live verification. Do not narrate that final verification as already done.
The persona speech-test failure, general visual work, memory investigation and
CodeRhapsody skill compatibility are not silently completed by this migration.

Authoring lesson supplied by Bill: an expert guides the build, the solution is
checked and corrected, and the chapter is extracted from that demonstrated work.
Then a fresh agent can use the codebook and graders to build again. These are two
connected directions of knowledge transfer, not merely a specification written
first and an implementation declared successful afterward. This crossover is
source material for the next chapter of that process.

Formal chapter drafting still requires the full governing author/reviewer
context and the normal review gates. The current runtime has retired earlier
full-document tool output; this entry records source facts and editorial intent,
not a substitute for that preparation. No publication artifacts were regenerated.

## Initial crossover checkpoint (historical)

Bill assigns Chapter 25 to the live crossover from Codex to Ensemble, Edition 2.
This is his first look at this GUI. Capture fixes discovered during actual use;
this note is not a chapter draft or an acceptance claim. Chapter 16 remains paused.
All implementation changes here belong in `solutions/edition-2/main/`; preserve
frozen chapter exports and Edition 1.

## First fix: approximate 5x speech speed

Request: raise the slider beyond 2x, accounting for Eloquence's nonlinear native
rate. Bill reports rate 2.0 is about 2x, 2.6 about 3.5x, and 2.9 about 5x; rate
3.0 is already about 6x. These are user observations, not new timing measurements.

Changes in `main/gui`:
- Slider and saved-settings clamp now allow 0.5–5; label says approximate speed.
- Rates through 2 are unchanged. Above 2, linearly interpolate the reported
  (displayed multiplier, native rate) anchors (2,2), (3.5,2.6), (5,2.9).
- Calibration applies to the default browser voice; it is tailored to this
  Eloquence setup, not a universal speech-engine timing guarantee.
- Real settings-wire/disk tests now cover a high fractional setting (4.2),
  multi-tab broadcast, partial patch preservation, restart, and a 5x clamp.
- Browser speech test includes a named subtest exercising the actual slider,
  settings broadcast, reload, and native utterance rates at seven settings.
  Only the native speech boundary is recorded; no audible timing is inferred.

Validation:
- `go test -race -run '^TestSettingsWireAndDisk$' -count=3 .`: PASS.
- Browser subtest `5x slider, Eloquence calibration and reload`: PASS.
- Deliberately replacing calibration with direct native rate makes that subtest
  fail at 3x (expected 2.4, observed 3); correct implementation restored.
- Full GUI Go suite fails `TestSavedBudgetEntry: application did not start`;
  also observed on the run started before source edits. Not repaired here.
- Full browser speech suite fails later at `whole final is spoken`, recording
  four identical artifacts instead of one. Reproduced with the original HEAD
  speech test against the current implementation; not caused by inserting the
  new speed subtest. Its underlying cause remains uninvestigated.
- Independent focused review (`review-speed`): no actionable findings. Reviewer
  did not rerun main actor's suites; explicitly limited timing claims to the
  recorded native-rate boundary.

Deployment: do not interrupt the active crossover session. Stage a replacement
GUI as `/Users/bill/ensemble-edition2-codex/bin/ensemble-gui.next` for a deliberate
restart; the currently running process and browser still use the old version.

## Live GUI fix: browser observations become on-demand tools

The active checkout now uses `agent/` (the earlier `solutions/edition-2/main/`
paths above describe the earlier checkpoint). This entry records a live runtime
change, not a manuscript draft, grader result, or chapter acceptance claim.

Bill noticed hearing "GUI snapshot" and "tts queue" repeatedly while Astra
worked. Inspection showed that `gui-debug` registered both browser tools with
`ephemeral: "round"`. Before every model request the runtime invoked them,
recorded tool events, and attached their latest output as ephemeral context.
Those ordinary tool events also reached the GUI's speech announcement path.
Astra saw one current snapshot and one current queue payload, not a growing
stack of snapshots: replacement of ephemeral context was working as intended.
The unwanted behavior was automatic collection and its audible/Actions traffic.

Bill chose on-demand inspection rather than merely silencing automatic calls:
- `agent/gui/web/gui/mcp.js`: remove the two ephemeral registration fields;
  preserve the `gui_snapshot` and `tts_queue` implementations and schemas.
- `agent/skills/gui-debug/SKILL.md`: instruct explicit calls for inspection and
  fresh snapshots after actions, without claiming current state from old output.
- `agent/gui/README.md`: document on-demand behavior.
- Leave the shared MCP/ephemeral mechanism unchanged for other tools.

Regression coverage uses the shipped embedded browser module in a real headless
browser (`components.test.mjs`, served by the external-consumer reuse test).
It checks all four browser tools remain discoverable, none requests ephemeral
collection, snapshot calls see changed visible text, and one explicit queue
call reads and returns the queue once. Discovery itself does not read the queue.
Initial `go test -race -run '^TestExternalReuse$' -count=1 -v .` in `agent/gui`
passed. Mutation and wider validation results will be added below.

Deployment status at this entry: source and skill updated; no GUI binary built,
installed, or restarted for this change yet. The running browser/server still
advertise the old automatic registrations until the updated assets are loaded
and the MCP connection is recreated. GUI visual polish remains unfinished and
is not part of this observation-tool fix.

Validation/deployment update for the on-demand fix:
- Deliberately restoring `ephemeral: "round"` on `gui_snapshot` makes the real
  embedded browser test fail with `gui_snapshot must be on-demand`. Restored
  the correct source in a `finally` block; the external reuse/race test passed
  again afterwards (including public events, reusable assets and relay).
- Core `go test -race -run '^TestMCP' -count=1 ./ensemble`: PASS; shared ephemeral
  collection and normal MCP calls remain covered by the existing tests.
- Full GUI `go test -race -count=1 ./...`: NOT GREEN. It failed only the reported
  `TestSavedBudgetEntry: application did not start` check in this run, matching
  the earlier recorded startup failure. Its cause was not investigated here.
- Tests above unset inherited provider selection/billing/reasoning environment
  variables to avoid treating live Codex settings as fake-provider credentials.
- Built `agent/gui/cmd` successfully and installed the replacement at the actual
  launcher target `/Users/bill/ensemble-edition2-codex/bin/ensemble-gui`.
  Prior executable retained as `bin/ensemble-gui.before-on-demand` in that same
  external runtime directory. Nothing is left merely staged as `.next`.
- Did not restart the running server or reload the human's tab. User must restart
  Ensemble with the usual launcher, then reload the browser (hard refresh if
  needed). Live confirmation that auto-attachments cease is still pending.

Live verification after deployment: restarting the server alone left the open
browser tab advertising the old ephemeral registrations. HTTP inspection showed
that the restarted server's `/mcp.js` matched the updated source exactly. After
Bill hard-refreshed the tab, the next user message and subsequent tool round no
longer included automatic snapshot/queue attachments. Both tools became directly
callable; explicit `gui_snapshot` returned the current three-pane GUI (with its
normal large-history truncation), and explicit `tts_queue` returned pending
speech. Explicit calls still generate ordinary tool-name speech announcements,
as intended. No further build or server restart was required after refreshing.

## Live configuration: request detailed reasoning summaries

Bill reported seeing only thinking titles. Inspection of the current runtime's
`session/api.log` showed `reasoning: {effort: "high", summary: "auto"}` in the
main Codex requests. Readable reasoning summaries returned by the endpoint were
short, single-line bold headings (roughly 19–32 characters in inspected parts),
or absent. The GUI renders returned readable summary text inside its Thinking
expander; no detailed paragraphs were present in those inspected responses.
Encrypted reasoning alongside summaries is opaque replay state, not display text.

Changed the external launcher `/Users/bill/ensemble-edition2-codex/start` to export
`OPENAI_REASONING_SUMMARY=detailed`, keeping `OPENAI_REASONING_EFFORT=high`.
Existing ConfigFromEnv and Responses request rendering already support this;
no production Go change or binary rebuild is necessary. Shell syntax check passed.

Added `TestCodexDetailedSummaryFromEnvironment` in
`agent/ensemble/codex_test.go`: a fake Codex endpoint asserts that real request
rendering carries both environment-selected reasoning settings. The initial test
setup had an invalid Config field; corrected it before validation. The test passed
three times under the race detector. The broader `^Test(Codex|Responses)` tests
also passed under the race detector with inherited provider selection and
reasoning/billing environment variables cleared.

Live verification is pending: the running process retains its previous environment.
Restart with the usual launcher, then inspect a new request in api.log for
`summary: "detailed"` and separately check returned readable summary lengths/content.
This change requests provider-designated detailed summaries; it neither guarantees
provider verbosity nor exposes encrypted/private reasoning.

## Live runtime fix: model response lifetime is not a tool callback deadline

During the configuration migration discussion, a delegated agent and later main
failed with `context deadline exceeded (Client.Timeout or context cancellation
while reading body)`. The model engine created every HTTP client with a hard
60-second total timeout. Main's debug log recorded a request at 20:49:50 and its
failure at 20:50:50. This is distinct from tool callback intervals: callback
checks must not cancel tool jobs, whereas the HTTP client was aborting the model
response and ending the current turn. Agent objects remained available for later
turns. The settings migration was paused before any implementation files existed;
its failed child left no partial code and was shut down.

Change: `agent/internal/engine/engine.go` now uses an HTTP client without a total
request timeout. Go's DefaultTransport still bounds connection setup and TLS
handshake. There is deliberately no blanket deadline on time to first response
headers or on response-body/model-generation time. Existing request contexts,
interrupts and shutdown remain cancellation mechanisms. No automatic retries or
new inactivity policy were introduced. Tool polling semantics are unchanged.

Regression evidence in `agent/ensemble/response_lifetime_test.go`:
- Real public Agent + local fake provider; both an open streaming body and a
  buffered response withheld before headers complete after 61 seconds. Both
  passed under the race detector (61.01s each, running in parallel). These real
  minute-long checks are skipped under `go test -short`, not silently shortened.
- Interrupt and shutdown each cancel the provider request, before headers and
  during an open stream. After interrupt, the same Agent handles a new request.
  All four cancellation cases passed under the race detector.
- Existing callback and managed-job interruption tests passed: returning a
  callback or interrupting a turn still does not kill its background jobs.
- A Go build overlay restored the committed 60-second client for a control run,
  without modifying the working tree. The new streaming regression failed at
  60.01s with the exact original Client.Timeout error. The control was successful
  precisely because the old implementation failed the new acceptance test.
- `go test -short -race -count=1 ./...` in `agent/`: the ensemble package and
  internal packages passed; the whole command is NOT GREEN. Root CLI tests
  `TestChatFlushesChunksAndSeparatesChannels` (`open : no such file or directory`)
  and `TestFleetCLIHumanReplyAndEscalation/true` (human sink unreachable) failed.
  Both failures reproduced in the committed-engine control under the same test
  environment. They remain separate unresolved issues, not fixed by this patch.
- `go vet ./...` in `agent/`: PASS. GUI `TestExternalReuse` under the race detector:
  PASS. GUI executable build: PASS. Tests cleared inherited provider selection,
  billing mode and reasoning environment variables; no live provider was used.

Deployment: install at the existing external launcher target
`/Users/bill/ensemble-edition2-codex/bin/ensemble-gui`, with the previous executable
saved as `bin/ensemble-gui.before-response-timeout` in that runtime directory.
The running process is not restarted by the coding agent; it retains the old
60-second client until Bill restarts Ensemble. No launcher flags or environment
compatibility interfaces changed. File-based settings remain the next pending
work item, with existing test/grader entry points to be respected.

## File-based application settings crossover

Bill requested visible file-based configuration instead of launcher-exported global environment variables, while preserving any interfaces used by existing tests and graders. Implemented a strict, explicit `--config` path alongside the unchanged no-config environment/flag entry point. File mode does not call the environment config builders: inherited provider/model/endpoint/reasoning/skill/auth-store variables cannot override the selected file. Only explicit `--port`, `--gui-dir`, and `--gui-debug` launch overrides are accepted and reported; competing provider/agent flags are rejected.

Implementation:
- `agent/gui/application_config.go` provides a versioned strict JSON loader, deterministic defaults, validation, config-relative path resolution, private atomic saves and runtime conversion. Snapshots are published only after successful writes. Unknown fields, embedded API-key fields, transient GUI metadata/model fields and malformed/trailing JSON are rejected. Credentials stay external: plaintext API-key file, existing Codex auth JSON, or the separate native ChatGPT auth store, according to billing mode. API-key contents are read only for runtime configuration and never serialized.
- `agent/gui/cmd/main.go` and `cmd/config_flag.go` select file mode before any environment parsing and use explicit workspace, skills, storage and auth settings. The legacy startup contract remains for graders/callers. Public `gui.New` remains intact; `gui.NewConfigured` is additive.
- The main agent's immediate GUI preferences and restart-only provider configuration share the application file. Child agents retain local preferences and inherit the running parent's provider configuration. File mode ignores the old main-agent `cr/settings.json`, but migration does not delete it.
- Settings now displays the configuration path, active provider values and saved next-launch provider/model/billing/reasoning values. Provider edits require an explicit Save and restart; speech/theme/font/memory/budget changes remain immediate. Advanced endpoint/credential-path inputs contain no credential contents. Save request IDs distinguish acknowledgments across tabs and from unrelated preferences, preserve newer in-flight edits, and handle server-normalized relative paths.
- README and `settings.example.json` document operation, credential handling, reload requirements for manual file edits, restart-only behavior and legacy launch compatibility.

Review and verification:
- Separate loader and frontend implementation agents completed; an independent read-only review found a real secret-handling bug: loading an API key before flag registration could expose it as `--api-key`'s default in help/error output. Fixed with an independently empty override flag; regression tests cover help and invalid flags. A temporary Go build overlay restored the unsafe default and the new test correctly failed (`credential leaked into flag output`), without modifying working files.
- Complete GUI Go suite passed under the race detector in a clean provider environment (`go test -race -count=1 ./...`), including the original `TestSavedBudgetEntry` and the real-executable config test. The latter poisons provider/skill/auth environment variables, launches from a different cwd, checks file-sourced credentials and actual high/detailed Responses fields, and verifies save placement. Config and WebSocket tests cover invalid input, private/atomic writes, failed-write rollback, stale legacy preference isolation, preservation of provider settings across GUI saves, and active-versus-restart state. `go vet ./...` and JavaScript syntax checks passed.
- The dedicated application-settings browser test passed, including legacy preference autosave, explicit provider Save, restart indicators, secret-free payloads, unrelated errors/broadcasts, normalized relative credential paths and edits made while a save is in flight. Cache/Responses/virtual-user EOF browser checks passed. Components and fleet browser scripts passed when invoked with their required ASSET_URL/GUI_BINARY fixtures. A naive glob invocation was not a valid aggregate harness; a first inherited-environment run stalled and was explicitly stopped without touching the live GUI.
- The speech pipeline's virtual-user capability assertion still fails because it expects `submit` among five persona tools and receives four. Reproduced identically from an archived committed GUI against unchanged core code: not introduced by this migration. The Eloquence slider subtest passes. This existing persona-tool collision remains separate work; do not claim an entirely green browser suite.
- All three delegated agents submitted, were joined and then shut down. No unfinished delegated work remains.

Deployment:
- Built and installed `/Users/bill/ensemble-edition2-codex/bin/ensemble-gui`.
- Created and validated `/Users/bill/ensemble-edition2-codex/session/settings.json` (0600), preserving OpenAI/gpt-6-astra/Codex routing, high/detailed reasoning, workspace, skills, save/data/log paths and all existing GUI/memory preferences (including speech speed). References `/Users/bill/.codex/auth.json`; no OAuth token was copied. Preserved the prior native auth-store location as an explicit file value too.
- Replaced `/Users/bill/ensemble-edition2-codex/start` with an export-free `exec ... --config ... "$@"` launcher. Retained `start.before-file-config`, `bin/ensemble-gui.before-file-config`, and the original `session/cr/settings.json`.
- The final replacement executable validated the real candidate with `--config ... --help`; launcher shell syntax and settings permissions passed. No automatic restart or live account/provider request was performed. Bill must restart normally and hard-refresh the browser, then verify the live configuration path, settings panel, restart markers and outgoing API request. Live verification is pending, not implied by fixture tests.

Unrelated book/publication and Edition 1 changes were not included in this work. General visual-polish work and the earlier CLI/submission-related followups remain distinct from this configuration migration.

## Visible action explanations as a model capability

Bill required a durable, ordinary-text account of what the model intends to do and why before tool execution. Private reasoning and sparse provider summaries are not that account. He explicitly requested no length check and a model flag, `RequireVisibleReasoning`, defaulting to disabled and enabled for OpenAI models.

Implementation:

- `internal/common/actor.go` adds the boolean model capability. The six supported real OpenAI entries in `internal/engine/media.go` enable it; other/unknown models retain false. The deliberately synthetic `gpt-ch19-course` stays false to preserve historical grader/protocol-fixture behavior. This is capability lookup, not a blanket provider-name guess or a user-editable GUI preference.
- The engine adds a same-response instruction to explain the proposed work and its purpose in ordinary visible text before calling tools. There is no mandatory heading format, minimum word/character count, or claim to assess explanation quality. The mechanical check is non-whitespace text, with correct order; it cannot certify that an explanation is substantively good.
- `internal/agent/actor.go` and `visible_reasoning.go` reject a whole unexplained batch before dispatch. Every proposed call receives a paired `tool_returned` error, but no tool job or `tool_called` execution event is created. The original response and valid opaque replay material remain intact. The next request receives corrective feedback and can retry with an explanation. Rejected batches count toward the existing round limit; rejection also honors an already-pending after-tool interrupt without issuing another request.
- Checking only parser-normalized final parts was insufficient: a Chat Completions stream can deliver a tool-call fragment before text and later normalize its final parts into text-first order. The actor now also tracks actual per-request `DeltaText`/`DeltaToolCall` order. Private summaries do not qualify. Stream-order rejection is explicitly identified in the saved error, so the durable record explains why normalized text-first parts were still rejected.
- A child previously inherited its parent's effective, decorated prompt, which could retain the rule after switching to an unflagged model. Internal `BaseSystemPrompt` separates stable instructions from dynamic capability policy at child construction, preserving user-authored text and child submission instructions. Tests cover children, grandchildren, explicit/empty/default primary prompts, deduplication and model changes. Busy model changes are already rejected, so an in-flight response cannot switch its policy mid-request.

Verification and review:

- The focused actor/prompt/stream regressions passed three times under the race detector. They use real public actors and HTTP parsers, replacing only the provider boundary. Marker-file checks prove rejected batches have no side effects; tests cover successful explained retry, saved/restored history, distinct opaque replay data, fresh explanation on each response, one-character visible text (no length threshold), default-disabled behavior, JSON and SSE paths, round limits, and after-tool interruption.
- Initial test errors were corrected rather than weakening enforcement: public persistence requires Load/Save; a fake key `fixture` accidentally matched fake ciphertext `opaque-fixture` and was legitimately redacted; history dumps are newline-delimited JSON, not a single array. The missing stream-handler update was caught by both regression tests and independent review, then fixed.
- Existing fixtures were updated narrowly. The Codex credential-rotation/tool-loop fake now includes visible explanatory text. The GUI saved-budget fake supplies explanations without changing its execution-budget assertions. A historical captured tool-only Responses stream remains unmodified and is tested both for successful synthetic-course replay and real-model rejection. The stream-delivery test preserves its authentic tool-first order and finalized-content checks, but now expects rejection for flagged OpenAI SSE rather than execution. The empty-primary cache test uses the synthetic capability because real OpenAI models now necessarily have a runtime policy instruction.
- Core `go test -short -race -count=1 ./...` passed the agent, engine, gateway, MCP and other internal packages. The overall command remains red only for `TestFleetCLIHumanReplyAndEscalation/true` (`human sink was unreachable`). A temporary archive of committed HEAD reproduced the same failure; it was not changed here. The earlier 90-second child-suite timeout occurred while running an unflagged fake-model test and was not evidence of an infinite refusal loop; the round-limit test subsequently completed in an isolated run.
- Build overlays independently disabled the gate, stream-order check, and after-tool interrupt handling. Each corresponding regression test failed for the expected behavioral reason. An initial wrong subtest selector was rejected as a failed control, then corrected; a no-tests run was not counted as evidence.
- Independent dispatch review accepted the final stream-order and interrupt fixes. Both workers submitted, were joined, and were closed. `go vet` passed for core and GUI. GUI Go race tests passed, with a final rerun against the completed source recorded separately below.

Deployment uses the existing live launcher/binary path, with a `bin/ensemble-gui.before-visible-reasoning` backup. A restart is required for the running main actor and future children to use the new capability enforcement. No automatic restart or live-enforcement verification is part of this checkpoint.

Separate continuity issue: Bill reports that `micro_handoff` hangs and used Interrupt. No successful handoff return was confirmed; do not claim it saved context or that this policy change repairs it. The checkpoint notes and scoped Git commit are the durable work record in the meantime. The PDF discussion was explicitly withdrawn as a misdirected message; no publication edits were made for it.

Final GUI validation: `go test -race -count=1 ./...` passed for the GUI module against the completed source (GUI package 12.777s; command package 1.010s). Installation is for the next launch only; live enforcement remains to be verified after an explicit user restart.


## Handoff isolated from memory compaction; restart verification

This follow-up supersedes the earlier uncertainty about whether a checkpoint was recorded. It does not establish a fix for the original auxiliary-compressor stall.

Bill disabled all six controls in the live GUI: Conversation, Session, 8×, 64×, Memory, and Soul. The saved application settings confirmed `disabled: true` for all six before the test and after restart. Disabling only the four upper bands had previously left Conversation → Session compression eligible. “Conversation” controls conversation compression, not the ability to chat; “Session” denotes compressed session memory, not the application's lifetime. These labels caused understandable confusion.

The all-disabled test succeeded. Reconciliation of `session/save.json` with `session/save.json.journal`, deduplicated by sequence number, established:

- Pre-test baseline: one `micro_handoff`, one `compactor_launched`, through sequence 3137.
- New call 3146 at 2026-10-14T15:00:37.098325Z; tool return 3147 at 15:00:37.149202Z (about 51 milliseconds).
- Checkpoint 3148 at 15:00:37.186165Z; the next ordinary main-model request began at 15:00:37.363740Z. Its response completed at 15:01:12.936876Z (about 35.6 seconds). This was not an auxiliary compressor request.
- Post-test totals through sequence 3198: two handoffs, still only one compactor launch. No new compactor launched after the test baseline.
- Bill then stopped Ensemble with Ctrl-C and restarted. The new snapshot was written at 15:09:43Z, with `as_of: 3184`; the journal resumed at sequence 3185. The specific “LIVE HANDOFF TEST WITH ALL MEMORY BANDS DISABLED” checkpoint was verified both in the saved handoff event and the saved dialogue. The six disabled settings survived restart.

Astra initially checked the wrong live-evidence source: the exported `event.log` had stopped updating at sequence 3137. The authoritative recovery state was the snapshot plus its active `.journal`. Do not infer that an operation failed merely because an exported trace is stale. The original snapshot was retained as `.bak` during graceful shutdown.

From the model's post-checkpoint context, old tool-result placeholders were no longer visible; new diagnostic results were visible normally. Bill observed only a modest token reduction. This is consistent with clearing already-stubbed tool traffic while retaining conversation and adding the checkpoint note; no exact token-savings claim was measured.

The earlier incident remains distinct: its checkpoint tool returned quickly, but an auxiliary conversation-to-session compactor was launched and had not completed when Bill interrupted about 93 seconds later. That interval was an unfinished wait, not measured successful compression time. All-disabled success isolates the subsequent compaction path for further investigation; it does not prove why that earlier request stalled. No runtime code or timeout policy was changed during this experiment, and memory bands remain disabled until Bill chooses otherwise.


## Instrumented conversation-to-session test: completed prose rejected as missing submission

Bill authorized a live test of the existing auxiliary compressor, not the planned regular background child. Conversation was enabled, Session was temporarily enabled through the GUI, and Soul/Memory/8x/64x remained disabled. Numeric values stayed at their zero/default settings. The GUI snapshot omitted the lower rows; saved settings, rather than an unseen checkbox state, were used to verify the changes.

A separate, metadata-only watcher (job 47) followed the live snapshot/journal and API log for five minutes without cancelling model work. Its initial baseline was handoff count 2, compactor-launch count 1, through seq 3324. Evidence is retained outside the repository in `session/diagnostics/compaction-live-3ywxg9hp/` under the live deployment. `baseline.json`, `observations.jsonl`, and `result.json` contain no source transcript, generated summary, or credentials.

### Observed sequence and outcome

- The main model first prepared the handoff: request seq 3326 at 17:00:15.646324Z, response seq 3327 at 17:02:11.457789Z (2026-10-14). That wait was not compressor execution.
- The checkpoint was recorded as seq 3330 at 17:02:11.621753Z, followed by Session-targeting `compactor_launched` seq 3331 at 17:02:11.707530Z.
- A structurally identified sole-`submit(memory)` request reached the provider. It used `gpt-6-astra`, high reasoning effort, detailed summaries, and about 223,827 serialized input characters. The request contained the main-agent no-submit instruction. No `tool_choice`, `max_output_tokens`, or `max_tokens` field was present.
- The provider promptly opened a response stream and continued sending ordinary text deltas. This test did not stall before HTTP dispatch or remain silent at the provider. The watcher ended its observation window at 17:05:15 without cancelling anything; subsequent inspection captured completion.
- The provider completed with 43,594 input tokens, 7,317 output tokens (130 reasoning tokens), 35,476 characters of ordinary assistant text, and **no function call**. These are metadata counts, not a disclosure of private reasoning or generated memory contents.
- The runtime rejected the completed response with `compressor must call submit exactly once`. No compressed Session population/replacement was recorded; the source was retained. The next main request, seq 3332 at 17:05:24.332683Z, followed approximately 192.625 seconds after the compactor launch. This was a completed-but-unacceptable response, not successful memory compaction.

### What the evidence does and does not establish

The output contract failed: a helper requiring `submit(memory)` used a main-agent prompt that forbids main submission, then received ordinary prose rather than the required tool call. The conflicting instructions are verified, but the trace does not prove the model's internal reason for choosing prose. This is stronger evidence than the earlier interrupted wait; it must not be represented as proof that the original 93-second incident had the identical cause. A background worker alone would not fix this output-contract defect. The planned restricted child needs its own role, an unambiguous result contract, bounded summary sizing, observable completion, and parent-controlled commit/cleanup.

After the test, Session was restored to its pre-test disabled state through the GUI and verified in saved settings. Conversation remains enabled; the four upper bands remain disabled. This prevents another Conversation-to-Session request while the defect is unresolved. No runtime code was changed, rebuilt, or deployed, and no live restart occurred. The background-compaction-child refactor remains unimplemented.

## Restored artifact IDs route new streaming text into old cards (2026-10-14)

Status: **reproduced; not fixed or deployed**. Bill reported that streaming text sometimes appeared in the wrong artifact and recalled an Edition 1 failure. He asked whether every artifact needed a unique ID.

### Diagnosis and reproduction

The current implementation already maps provider/parser-local part IDs to a counter owned by the actor runtime. That handles repeated parser IDs within one run, but a new runtime starts its counter at zero. Saved response parts retain their old `ArtifactID`, and GUI replay restores cards under those IDs. New deltas can therefore address an unrelated, restored card. The renderer keys by `String(part_id)`; a delta appends to that card, and a final replaces its text.

A structured scan of the live snapshot and journal found 470 distinct finalized IDs, with 385 reused on multiple finalized parts, sometimes across different part kinds. A later scan found no collisions within a single response. These are point-in-time observations, not a claim that every visual anomaly has the same cause.

An isolated probe used the real public agent, streaming parser, persistence, and two successive runtimes, with only the HTTP provider replaced by a local fixture. The first reply, `FIRST`, received delta/final ID1. After save/reload, the second reply, `SECOND`, also received delta/final ID1. Both replies remained separate in saved history.

The real artifact renderer was then exercised in an isolated browser page with those frames. Actual DOM assertions confirmed one card throughout: `FIRST` after replay, `FIRSTSECOND` during the second reply, and only `SECOND` after its final. The renderer served by the running GUI was byte-for-byte equal to the inspected source. This was not a test against the live browser, and no real provider inference was used by the diagnostic probe.

### Edition 1 connection and lesson

The Edition 1 crossover changelog already records Bill's rule: anything talking to an artifact needs a unique ID. Its earlier fix introduced agent-lifetime allocation and required deltas and finals to share the same mapping. Related Edition 1 tests cover replay, restart, and final/delta correlation. The current gap is uniqueness across the logical actor's restored history, not merely one process lifetime.

The repair must provide stable producer-owned identities, keep one identity throughout an artifact's stream and finalization, and handle old history that already contains collisions. Seeding a future counter alone would not disambiguate those older replayed parts. Clearing the scroll every turn or generating unrelated DOM IDs would hide or relocate the problem, not establish the identity contract.

The complete source references, private diagnostic location, and proposed regression coverage are in [the artifact identity investigation](artifact-identity-investigation.md), committed as `bd4d907c`. The read-only history reviewer submitted, was joined, and was closed. The actor allocator overlaps Codex's concurrent lifecycle work, so no application patch was made without coordination. No corrected live behavior or full post-fix regression run is claimed.
