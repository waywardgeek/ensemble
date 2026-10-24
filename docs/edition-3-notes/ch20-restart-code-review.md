# Chapter 20 restart — retained independent code review

## Scope and baseline

Accepted predecessor: Chapter 19 HEAD `0bac5a47018e533e3d97d6e2820d67d3cc01036d`. No Chapter 20 student has been launched. This is the bounded **entry** review of root's public instructions and grading alignment, not a student implementation or product-acceptance verdict.

Reviewed the current governing instructions and chapter; the accepted public core/CLI/separate-GUI owners; HEAD Chapter 20 graders and called helpers; and the stable root entry diff in `cmd/grade-edition3/main.go`, `ch20_test.go`, `internal/gradeedition2/ch20*.go`, `book/edition-3/chapter-20.md`, and `docs/edition-3-student.md`. The student owns production changes; root owns public book/fixture alignment and coordinator work. Production was unchanged at entry. This reviewer edited only this note. No builds, tests, graders, native/provider/browser runs, installs, restarts, children, or commits were performed by this reviewer. Student isolation was preserved.

## Correction of withdrawn preliminary findings

This note replaces the prior version and all preliminary summaries. The following earlier claims were unsupported or wrong and must not drive edits:

- There is no `ch20PublicConsumer`, `ch20BuildGUI`, or `publicConsumer` owner in the reviewed HEAD grader. The actual reuse owners include `reuseChecks`, `reuseApp`, `reuseHeadless`, `reuseDeploy`, and `reuseWire`.
- Current public `New` takes `*log.Logger`, not `io.Writer`. `api.New(nil)` is valid. Current `Agent.RecordFact` is the public event-ingestion boundary used by the updated consumer.
- `Part` is a tagged **struct**, not an interface; `PartList` is `[]Part`. The current `api.PartList{{Type: "text", Text: ...}}` literal is correct. No `TextPart` rewrite is needed (`internal/common/history.go`).
- `Agent.MCP().Connect` and `NewMCPTransport` exist and are compatible. Keep them.
- `ConfigFromEnv` supports `LLM_*` and `EN_SKILLS_DIR` / `EN_PRIMARY_SKILL`; the earlier proposed `ENSEMBLE_*` skill-variable rename was wrong.
- CLI plain-text `chat` input and the existing `user`/`assistant` envelope are supported. The Chapter 20 GUI helpers use `prompt`, not `question`.
- `grade.Build` already builds the GUI module's `./cmd`. It cannot treat the core library root as its CLI command; that command is `cmd/ch01`.
- The outside GUI reuse probe expects **one** real provider request containing its planted public event and prompt. The model-switch probe expects **six** provider requests across three two-request turns. The earlier claim of exactly three outside-consumer calls was wrong.
- The earlier statement that `reuseChecks` used `reflect.DeepEqual` to compare extracted base prompts was wrong. No new acceptance condition is inferred from that claim.

## Verified current owners and entry alignment

The accepted core module is `ensemble`, exported at its root. The separate GUI module is `ensemble/gui`, with its command in `gui/cmd`. The core CLI command is `cmd/ch01`. Root `config.go:ConfigFromEnv` owns shared environment configuration.

The accepted GUI's early `gui.New(a) http.Handler` and current `NewServer*` server are not yet the Chapter 20 reusable-hub/assets/relay implementation. The chapter and generated outside consumer deliberately retain the published `gui.New(root, a, writer)` / handler / assets / relay contract as Chapter 20 work. This entry review does not require those new APIs to work in the unchanged Chapter 19 predecessor and does not approve a scripted substitute for the real Agent.

The reviewed patch is appropriately bounded:

1. `Chapter20Targets` separates explicit CLI and GUI artifacts from the core and GUI module roots. Edition 3 dispatch uses existing `-core`, `-gui`, and positional CLI inputs, with early invalid-input rejection. The historical `Chapter20` route remains available. Fourteen checks / 200 points are preserved.
2. The current outside consumer imports the core root, calls `api.New(nil)`, and records a public event via `RecordFact` with the accepted `PartList`. Existing independent core/GUI module replacements are retained. The real history replay, actual prompt/provider request, headless MCP handshake, dependency graphs, browser relay, and assets checks remain in place.
3. `reuseWire` uses the existing `statusLaunchModels` live-control branch, with supported Sonnet / Opus / Sonnet selections, and still checks all six actual provider request models. It does not put model selection back into GUI settings or replace evidence with acknowledgements.
4. `fleetDriver` centralizes current versus historical launch differences. The current route uses the existing `webStatusEnv`, supported default model, owned HOME/TMPDIR, fake provider endpoint/keys, and a fixture skill. Explicit native confinement is removed for this local loopback/provider fixture; it is not presented as a native-confinement test. The GUI driver checks inert static-asset setup errors. The schema/submission assertions themselves were not weakened.
5. The book now describes the actual root module/API and separate build targets, and says the conceptual Config sketch does not require a second public type. The public brief points library-root users to explicit CLI builds. The separately identified headless recipe was subsequently corrected from `ensemble/ensemble` / `New(io.Discard)` to `ensemble` / `New(nil)`.

### Simplicity comparison

Compared with HEAD, this patch keeps the original assertions and shared helpers rather than duplicating the Chapter 20 suite or inventing another process/protocol abstraction. The small `fleetDriver`, current-route booleans, and entry wrapper are proportionate to the two layouts. The generated consumer still exercises public code rather than injecting a fake Agent hook. Production is unchanged, so no historical production rewrite comparison is warranted for this entry decision.

## Reviewed blocker and resolution

**Resolved:** the first stable revision of `internal/gradeedition2/ch20_reuse.go:reuseChecksWith` leaked the outside-consumer build error into the independent GUI checks on the current route.

In that revision, a failed `reuseGo(work, "build", "-o", "bin/", "./...")` left `err` nonnil. The code correctly failed consumer checks 0, 1, 4, 5, and 6, but the current route skipped `grade.Build(gui)` and its subsequent error test also failed checks 2 and 3, suppressing `reuseWire` / `reuseDeploy`. The historical route overwrote the variable with the GUI build result.

Root fixed this with a separate `guiBuildErr`, preserving the supplied current GUI binary and keeping consumer failures attached to their intended checks. Inspected that exact correction. The route test now verifies that a consumer compile failure still reaches the independent deployed-GUI probe. This was a grading-control-flow defect, not a student implementation defect.

## Evidence and limits

Inspected root's saved evidence directory named by `/private/tmp/ch20-entry-current`: `/private/tmp/ch20-entry.m1m1EM`.

- `focused.log` / `focused.exit`: targeted entry/helper tests pass, exit 0.
- `compile-v5.log` / `compile-v5.exit`: repository compile-only check passes, exit 0. This is not a product behavior run.
- `core-api.log` / `core-api.exit`: public core construction/event/MCP snippet compiles, exit 0. Read the actual snippet as well; it does not replace compiling every book example.
- `route-v2.log` / `route-v2.exit`: strengthened route test passes, exit 0, after the error-scope fix.
- `lint-final-v2.log` / `lint-final-v2.exit`: corrected chapter lint passes, exit 0, with the recorded person-gap warning. This reviewer did not rerun lint.
- Reviewer `git --no-pager diff --check` on the entry files reports no whitespace errors.

Reused these results rather than rerunning tests without a new runtime question. Earlier root wrapper-signature compile failures were setup failures, not observed product defects. No full Chapter 20 grade, native test, or live-provider run has been performed for this entry review. The separate unchanged-source identity check remains root-owned.

## Decision

**ENTRY PASS.** The separate-module/artifact routing, current environment and model-control adapters, generated public-core consumer, and public recipe are aligned sufficiently to begin the bounded Chapter 20 student implementation. Both concrete issues found in the stable entry review—the stale headless recipe and independent-check error-flow regression—are corrected and inspected. No verdict of Chapter 20 student/product acceptance is possible yet. Reviewer remains retained for production review; no broader campaign is requested.


---

## Early compiling HOLD — retained production review

**Verdict: EARLY HOLD, not ready for full acceptance.** Entry remains accepted and was not rerun. Student source was reviewed frozen at the first compiling hold; zero correction rounds had been used. This section supersedes no entry history. The student owns production; root owns public alignment/coordinator work. Reviewer changed only this note and ran no builds/tests/graders/native/provider/browser activity.

Current public Chapter 20 governs: `New(root *ensemble.Ensemble, agent ensemble.Agent, log io.Writer) (*Server,error)`, `Assets`, `Server.Handler`, and **Server** `ServeMCP` are required. Explicit root submit is allowed under the same admission rules; only delegated Agents owe submit and receive the three corrective wakes. Ordinary root chat is valid. The resolved root handoff ambiguities are not student defects.

### Architectural assessment

The basic direction is useful: `Ensemble` holds one fleet registry, its entries retain actual Agents, tools mostly translate arguments into that owner, and the existing Agent actor still executes requests. There is no need to introduce another fleet actor or duplicate tool-side orchestration. The small CLI/GUI factory closures own application defaults; duplication alone is not grounds for another abstraction. `Config.Child` copies relevant collections and `Clamp` enforces narrowing; low-level `NewAgent` also clamps and rechecks the parent before publication. Those safeguards should be retained.

A bounded read of the original Chapter 20 factory/fleet boundary was used only as a responsibility comparison. It supports separating application construction policy from one library owner of fleet facts, not copying its structure or treating its behavior as an answer key. No reference implementation details are needed in the student correction brief.

### Prioritized concrete findings

Paths below are relative to `solutions/edition-3/main`.

1. **Factory failure/admission is not transactional.** `fleet.go:60–74` invokes the application factory; `Spawn` then returns directly on factory error or wrong registered child/contract (`111–123`). A factory that calls `NewAgent` and subsequently errors, or returns a newly registered child with the wrong immutable schema, leaves that live registration behind despite spawn reporting failure. There is also no caller-context or parent-liveness recheck between factory return and creating the background task context (`125–129`). `NewAgent` does perform its own pre-publication parent check (`ensemble.go:243–247`), so the gap is specifically the later factory-return/task-launch phase. Keep one construction/admission owner: track the child acquired by this attempt, clean up only that acquired child on failure, and recheck cancellation/parent availability before launching. Preserve an independent expected contract when calling the factory; do not let mutable schema input redefine the contract being verified.

   Related pre-activation boundary: recovery-path collision checks occur after `a.Load`, initial skill/MCP setup, and initial `RecordFact` (`ensemble.go:187–223`, then `230–260`). Move identity/path collision admission before those side effects where possible. Do not solve this by adding a second fleet registry or stopping an arbitrary mismatched factory return.

2. **Accepted submit is not yet terminal at the actor scheduling boundary.** `internal/agent/actor.go:801–810` copies the accepted result then calls the ordinary `finish`; `finish` transitions to Idle and replies (`868–887`). The actor loop subsequently calls `startNext`, which accepts pending work while `!a.stopping` (`430–435`). The separate `Spawn` goroutine only calls `a.Shutdown()` after receiving that reply (`fleet.go:129–145`). Thus queued work can start in the gap after accepted submit. This is more than the already-declared missing durable file: successful submission must close future work admission in the actor before replying, and final fleet success must follow durable payload/save completion. Keep ordinary chat turn completion distinct from delegated lifetime completion. Complete corrective wakes and terminal outcomes in the existing owner, not a competing lifecycle loop.

3. **Child model override can carry stale resolved identity/capabilities.** `internal/common/policy.go:12–34` copies the parent config. Both application factories then change only `Model`/reasoning override (`internal/cli/fleet.go:25–33`, `gui/cmd/fleet.go:25–33`). `NewAgent` prefers retained `ResolvedModel` for stream capability (`ensemble.go:138–145`); `NormalizeOpenAI` fills resolved identity only when empty (`internal/common/auth.go:109–113`), and existing accounting/capability readers prefer it. A new valid child model can therefore retain the parent's resolved identity and derived behavior. The existing live model setter already clears `ResolvedModel` before normalization (`internal/agent/model.go:52–58`). Reuse that policy, clearing/recomputing model-derived state at construction when an override changes the selected model; retain current-parent inheritance when the override is empty. Do not broaden the child provider or sandbox policy to solve this.

4. **The selected reusable GUI still installs fleet-wide application policy.** `gui/server.go:398–408` accepts a selected Agent but calls `newHub(root, settings, configForSelectedAgent)`. `gui/hub.go:80–90` views every existing/future Agent, and `112–120` applies the hub's security settings and `speaking=true` to each. Constructing a selected embedded view can change other Agents' policy; future child views can also overwrite their separately narrowed configuration. Multiple selected views can fight over these writes. Keep one Hub implementation, but make the selected reusable path observe/command the selected Agent without applying application policy to unrelated Agents. Whole-fleet application views, if retained, must not widen child restrictions. Application settings belong at application construction, not as a side effect of attaching a library view.

5. **Default GUI launch cannot reach the embedded-assets handler.** `gui/cmd/main.go:32` now defaults `-gui-dir` to empty, but `NewServerFromConfig` calls `serverAssets` (`gui/server.go:46–49`); `serverAssets` resolves the empty string to cwd and requires disk `index.html`, `app.js`, and `style.css` (`95–104`). Startup therefore fails outside a frontend directory before `Handler("")` can serve embedded assets. Route empty-directory construction consistently to embedded assets and retain disk validation only for an explicit override. This is distinct from the already-known renderer extraction failure.

### Already-declared unfinished work, not a discovery campaign

- Finish the specified bounded schema dialect, exact decimal arithmetic, resource budgets, immutable normalized contracts, and default `additionalProperties:false` normalization (`internal/common/submission.go`).
- Finish durable `submit.md` and retrieval fallback, successful/failed child resolution, three delegated corrective wakes, explicit lifetime outcomes, and cleanup/save ordering (`fleet.go`, Agent terminal path). The present first-task-result watcher immediately fails a delegated Agent without submit; that is an incomplete lifecycle, not evidence that idle is a valid successful outcome.
- Make join yield pending questions and provide the human/root sink (`fleet.go:203–247`, `337–389`). The existing per-question reply channels can be retained; do not add another actor solely for these facts.
- Add explicit root submit under current public admission rules. The current delegated-only check in `fleet.go:251–257` is a known unfinished mismatch, not blame for root's withdrawn shorthand.
- Extract and reuse the required renderer/artifact modules. Do not merely add empty files to satisfy the retained file check; the public requirement is actual shared rendering behavior.

**Smallest coherent completion brief:** repair the construction/model and selected-GUI/startup boundaries above; then finish submission and delegated lifetime as one coherent path in the existing Agent + Ensemble owners (admit, validate, durable commit, stop future work, publish terminal outcome, resolve/join), including root/chat distinction and questions. Finish the renderer extraction without a GUI rewrite. Keep the tools thin. Return at a bounded compiling/focused hold for inspection of these paths rather than launching a full acceptance campaign while declared work remains incomplete.

### Relay review and evidence limits

`gui/relay.go:22–72,130–153` gives each TCP relay a private routing key, binds it to one browser, checks the answering browser, and forwards the original JSON-RPC payload. `runRelay` closes both directions on disconnect; `closeRelays` closes listeners/sessions and waits (`74–128,156–169`). This is a reasonable reuse of existing framing/transport, not grounds for another RPC orchestration layer. No demonstrated new relay-correlation defect is asserted here. Multi-browser correlation and unanswered-call cancellation were source-reviewed but not runtime-proved by this review.

Read `docs/edition-3-notes/ch20-restart-student.md` and actual receipts under `/private/tmp/e3-ch20-student.2X1OuS/`:

- `core-compile-final.txt` and `gui-compile-final.txt`: compile-only passes; not full test execution.
- `core-focused-03.txt`: records a passing focused core-package run (`ok ensemble 0.113s`), consistent with the student's declared local-loopback check; it does not establish the unfinished schema/lifecycle contract.
- `gui-focused-02.txt`: `TestCh20PublicEmbeddingAndRelay` passes; `TestCh20ReusableRendererModules` fails with `open artifact-scroll.js: file does not exist`. This retained FAIL agrees with the declared renderer gap. The no-browser relay coverage does not demonstrate a live multi-browser relay.

No new test campaign, native/provider/browser activity, production edits, or acceptance verdict was performed. Reviewer remains retained for the next bounded correction review.

### R1 narrow architecture clarification — factory provenance

No moving-source review or tests performed for this clarification. Read the public factory/construction requirements, particularly Chapter 20:380–435.

**Narrow the obligation; do not require universal rollback of arbitrary factory effects.** The factory is application policy, not an adversarial sandbox. It constructs the one requested child; it owns cleanup of extra resources or extra Agents it creates outside that request. The earlier HOLD finding concerns the intended child left live after failed admission, not every registration occurring while the callback runs. Snapshot/diff of the global registry is unsafe under concurrent construction and is not recommended.

The simplest adequate ownership path is explicit transfer: the factory owns its newly constructed intended child until the library admits it. On a post-construction error it must either retire its own child or return that child with the error so the admission owner can retire it. The library must verify that an object being retired belongs to this construction attempt—not merely that it was returned by the callback—and must never shut down a foreign/preexisting object. Use the existing requested-ID admission/reservation boundary and exact Agent identity to establish that ownership; if that boundary cannot distinguish attempts, a small per-attempt token/handle is justified. Keep cleanup centralized and limited to that attempt's intended child, with cancellation and parent availability checked before task launch.

**The proposed `NewAgent(Config, ...context.Context)` extension is not required by the public contract and should not be introduced solely to promise cleanup of every misbehaving-factory effect.** A context-carried provenance token can be a bounded implementation choice if the actual admission design needs it, but optional propagation cannot guarantee cleanup of a factory that omits or substitutes that context. It therefore cannot support the stronger promise either. Prefer a single explicit child-construction/admission path over widening the ordinary root constructor with variadic contexts. No transaction framework or parallel registry is warranted. This is architecture guidance, not approval of any uninspected R1 implementation; reviewer remains retained.


---

## R1/10 stable HOLD — retained independent review

**Verdict: HOLD for bounded correction; not accepted.** Reviewed the frozen R1 source, the full current public Chapter 20, the student note, and existing `r1-*` focused/compile receipts in `/private/tmp/e3-ch20-student.2X1OuS/`. No tests/builds/graders/native/provider/browser runs or production edits by this reviewer. Entry remains accepted and was not reopened. Findings below supersede early-HOLD findings where R1 has changed the implementation; prior entry/hold history is preserved.

### What improved; ownership and simplicity

R1 makes substantial real progress: the schema code implements exact rational comparisons and explicit admission budgets; the submit actor closes admission on accepted submit, rejects queued requests, and publishes success after cleanup/final save/durable row; reconstruction recognizes durable completion without another provider call. Corrective wakes, failed-child resolution, question-yielding joins, a human-question sink, root submit, earlier path collision checks, and model-derived identity reset are present. GUI view creation no longer rewrites every Agent's security/speech settings; settings updates target their owner and clamp child permissions. The empty-directory launch path now supports embedded assets. Extraction into browser modules is real, though the required public component remains incomplete below.

The bounded factory admission record in `fleet.go:22–34,96–183` is consistent with the saved provenance clarification: one intended child is tracked, the expected schema is privately copied, and rollback is limited to that owned attempt. No variadic-context constructor, global snapshot/diff, or general transaction framework was introduced. Keep the single Ensemble registry and existing Agent actor; no second orchestrator is warranted.

**Parent ownership clarification:** inherited `Agent.Parent() common.Ensemble` is a valid service-owner pointer. The application factories resolve `Config.ParentID` through the actual Ensemble registry and obtain the parent's current configuration. This review identifies no concrete missing parent behavior requiring replacement of `Parent()` or a second mandatory Agent back-pointer. The chapter's illustrative signature is not a naming/structure mandate.

### Concrete remaining blockers and smallest fixes

All source paths below are relative to `solutions/edition-3/main`.

1. **Delegated failure still has a terminal-admission race.** `internal/agent/actor.go:912–948` exhausts the three wakes or processes an error, then clears the active turn, transitions to Idle, and replies without closing admission. The existing loop's `startNext` can start queued work before the separate watcher in `fleet.go:188–204` receives that reply and calls `Shutdown`. Accepted submit has a terminal gate, but delegated failure does not. Close admission and dispose of queued/follower work in the existing actor terminal owner before reporting delegated lifetime failure; finish cleanup/save before publishing the final outcome. Do not turn every ordinary human-chat turn error into application shutdown. The successful-root-submit queue test is not evidence for this failure branch.

2. **Explicit shutdown resolves outcomes but does not remove managed-child membership.** `fleet.go:220–229` returns every retained child; `ShutdownAgent` at `416–452` only marks resolved/stopping/terminal, and `internal/tool/fleet.go:70–88` deliberately lists retained terminal entries. The public shutdown contract removes the child from the managed fleet. Preserve durable/terminal outcome retrieval, but distinguish that retained record from active managed membership and remove explicitly shut-down children from fleet enumeration and ownership-facing operations. This does not require throwing away results or adding another registry. **Named join-ID validation is correct** at `fleet.go:290–295`; the reviewer's interim claim that unknown named IDs were silently ignored is withdrawn.

3. **The fleet tool/message surface is still narrower than the current public contract.**
   - `internal/tool/fleet.go:43–63` advertises/requires `prompt` and rejects the chapter's `task` argument. Accept and advertise the public task input (mapping it to the existing internal Prompt field is enough).
   - `owned` at `24–36`, used by `send_message`, restricts its target to a direct child. `fleet.go:381–392` uses bare parent/direct-child IDs, not the public addressed child/parent/human exchange. Questions have sender fields, but `fleetQuestion.reply` is `chan string` (`455–462`) and responses at `529–530,610–611` discard the responder envelope. Complete one addressed delivery/reply path so a child, parent, or human reply identifies its sender using a returnable address, including the specified human address. Reuse the existing mailbox and correlation table; no separate messaging actor is needed.
   - The public terminal string convenience route is not present: `agent.Submit`/`SubmitLive` (`internal/agent/actor.go:156–161`) still submit user input, while exported `AcceptSubmission` only reserves fleet admission. Keep inherited chat submission behavior, but expose a completion convenience that reaches the same validation/durable/terminal owner and normalizes a default string result to `{ "result": ... }`. Do not implement a second completion path that bypasses the actor.

4. **Reusable selected GUI is not actually selected for observation or commands.** `gui/server.go:398–410` supplies the selected Agent only as default/settings owner; `gui/hub.go:73–89` still seeds every root Agent and subscribes to all observations (`239–244`), and replay builds all Agent snapshots (`345–380`). `gui/server.go:294–301` accepts any `Root.Agent(id)` for commands. A selected embed therefore exposes unrelated histories and can control unrelated Agents. Construction-time policy mutation was fixed; this distinct isolation requirement remains. Add selection to the existing Hub/Server path so replay, live observation, commands, and relevant browser MCP selection stay within the attached Agent. Retain any intentionally whole-fleet application path separately without duplicating the projection machinery.

5. **Required browser component surface is missing despite useful extraction.** `gui/web/gui/artifact-scroll.js` exports `bindArtifactScroll`, a scroll-behavior helper; it does not provide `ArtifactScroll(container, renderers).receive(message)` that renders and appends chat/MCP/history artifacts. Rendering was extracted to `part-renderer.js` and `tool-result.js`, while the public `renderers.js` export is absent. The current chapter explicitly requires this reusable component and the renderer bundle; it is not a private fixture typo. Build the small component over the extracted renderers and use it in the main GUI, preserving chunk/final handling, tool cards, fold state, speech controls, and auto-scroll. Do not satisfy this with empty filename shims or duplicate rendering code.

6. **Schema admission limits differ from the declared public subset.** `internal/common/submission.go:48–49` allows 256-byte number literals instead of 128; `72` starts depth at one rather than root zero; `91` applies schema depth 32 instead of the public 64. Correct the boundary constants/counting while retaining the exact arithmetic, owned bytes, node/byte limits, and combined budget. These are source-proven acceptance/refusal errors, not a reason to replace the validator. The reviewer verified `const:null` handling and null rejection for the typed keywords inspected; no unverified blanket 'null keywords accepted' finding is included.

### Existing evidence and limits

The saved final R1 core/GUI compile-only receipts pass. The focused R1 receipts report green coverage for the new lifecycle/schema/admission/recovery and GUI embedding/relay cases after earlier retained failures were corrected. Those results are useful, not full chapter acceptance. The renderer test now checks the student's extracted module names; it does not establish the different public `ArtifactScroll.receive`/`renderers.js` surface. The successful-submit queue check does not establish failure-side admission closure; selected-view settings isolation does not establish selected replay/command isolation. No runtime failure is claimed for unexecuted concurrency scenarios beyond the directly visible ordering defects described above.

Relay framing/correlation/cancellation remains on the existing bounded relay/session owner. This review does not assert a new relay defect merely because broad browser/native evidence is absent. No new campaign should be launched to compensate for the concrete source mismatches first.

**Next bounded brief:** close delegated failure admission in the existing terminal path; distinguish stopped membership from retained outcomes; align the public task/addressed-message/completion convenience boundaries; restrict selected GUI observation/control; finish the real reusable artifact component; correct the explicit schema limits. Preserve the current factory ownership contract, exact-arithmetic validator, actor/library responsibility split, inherited chat APIs, and useful durable recovery work. Return after those coherent corrections with focused evidence; root retains acceptance authority. Reviewer remains retained.

### Root reconciliation after collecting/joining R1 review

Root independently read the held owners and focused receipts. Two source-proven
omissions from the review summary are included in the public R2 brief: CLI only
handles `/answer AGENT REQUEST RESPONSE`, not required `/reply TEXT` and `/reply
MESSAGE_ID TEXT`; `submission.go:422–431` imposes schema-node × payload-node <=1000000
rather than the public canonical-byte product <=67108864. Preserve individual byte/
node limits but replace that substituted combined refusal. The actual scroll helper
is named `createArtifactScroll`, not `bindArtifactScroll`; this naming correction
changes neither the missing receive/render interface nor the reuse finding.
Root briefly suspected failed-child resolution was ignored by AcceptSubmission, then
reread its actual `resolved` check and withdrew the suspicion before dispatch. It is
not a production defect. Review/receipt inspection did not execute new product tests.


---

## R2/10 stable HOLD — consolidated retained review

**Verdict: HOLD for one narrow public browser-module correction; not full product acceptance.** Reviewed frozen R2 production, current Chapter 20, the complete R2 brief at the end of `docs/edition-3-student.md`, the student record, root's R1 reconciliation, and actual saved `r2-*` receipts under `/private/tmp/e3-ch20-student.2X1OuS/`. Entry was not reopened. This reviewer ran no tests/builds/graders/browser/native/provider activity and edited no production/public contract.

### Six-group disposition

1. **Delegated failure admission: addressed.** `internal/agent/actor.go:974+` centralizes terminal admission closure; `finish` at `995+` sends failed delegated lifetimes through terminal cleanup rather than replying Idle and letting queued work race the fleet watcher. Ordinary root chat retains its separate behavior. The focused exhausted-wake, provider-error, and tool-budget cases exercise the relevant failure classes.
2. **Managed membership versus outcomes: addressed.** `fleet.go:224+` omits retired children from managed enumeration; explicit shutdown (`489+`) retires them, while retained records still support terminal outcome/submission retrieval. Failure resolution still permits parent submission. No additional registry/framework was introduced.
3. **Task/completion/addressed communication: addressed in the reviewed paths.** The spawn adapter accepts the public `task` while retaining the compatibility alias. `Agent.Complete` (`internal/agent/actor.go:96–103`) submits the default string result through the existing actor/validation/terminal path, including active-turn handling; inherited chat `Submit` is not repurposed. `fleet.SendMessage` (`390+`) and the existing question table return correlated addressed envelopes, including the configured human sink. CLI `/reply [MESSAGE_ID] TEXT` is handled locally (`internal/cli/chat.go:273–295`) with unique-pending, explicit-ID, stale, and ambiguous cases; it does not become provider input. Root's prior `/reply` reconciliation is incorporated.
4. **Selected/global GUI: addressed.** The reusable constructor selects the existing scoped Hub; observation, snapshots/replay, commands, and MCP attachment are filtered, while the intentional whole-fleet composition remains available. The existing Hub remains the owner; no duplicate projection layer was added. **Withdrawal:** an interim reviewer message alleged that `PresentHumanQuestion` could bypass selected filtering. This was incorrect: `Hub.broadcast` explicitly filters `w.Question.Sender` at `gui/hub.go:318–320`. There is no saved human-question selection blocker.
5. **Artifact component functionality: implemented; public module entry remains incomplete below.** The class has `receive`, frame identity/update/removal/replay behavior, default/overridable renderers, and is consumed by `app.js`. It builds on the extracted rendering functions rather than a second implementation. The original helper was `createArtifactScroll`; root's R1 correction is preserved. The remaining issue is the exact mountable module graph, not absence of a component or registry.
6. **Schema limits/product: addressed.** `internal/common/submission.go` now uses 128-byte numeric literals, root depth zero, schema/payload depth 64, and the exact canonical UTF-8 byte-product bound (`PrepareSubmission:413–439`, product check `430`). Default extra fields remain subject to admission budgets before normalization. Exact rational comparison and the bounded validator remain intact. The carried focused schema results cover the changed boundaries. Root's byte-product reconciliation replaces the earlier incomplete reviewer description.

### Sole concrete blocker: documented two-module embed cannot load

Current public Chapter 20 lines `39`, `316`, and the component-only recipe at `885–900` require mounting **`artifact-scroll.js` and `renderers.js` only** from `gui.Assets()`.

Actual source:

- `gui/web/gui/artifact-scroll.js:1` imports `./part-renderer.js`.
- `gui/web/gui/part-renderer.js:1` imports `./tool-result.js`.
- The renderer registry exists in `part-renderer.js:43–50`, but there is no exported asset named `renderers.js`.

An application following the public two-handler recipe cannot resolve that module graph. This is a source-proven packaging/mount contract defect, not a browser-test assertion. A `renderers.js` compatibility re-export **alone** would not fix the documented example if it still imports unmounted `part-renderer.js`/`tool-result.js`.

**Smallest coherent correction:** preserve the working renderer/component ownership and publish the required browser entry with a dependency graph satisfied by the advertised mount. A self-contained public renderer bundle and `artifact-scroll.js` importing it is one adequate path. If root instead changes the public mounting recipe to explicitly include helper dependencies, make the exported `renderers.js` surface and the actual recipe agree; that public alignment remains root-owned. Do not duplicate the rendering implementation or add an application/runtime framework. Recheck only this public asset/import boundary and affected component behavior after the bounded change, as authorized by root.

### Receipt interpretation — no blanket green claim

- `r2-core-focused-03.txt` is an overall failing run because its new tool-budget fixture reused a provider call ID. Its unchanged R1 core regressions, schema boundary tests, factory-model test, and CLI cases passed and remain useful carried evidence.
- `r2-core-focused-04.txt` fixes and passes the intended core failure-budget case, along with terminal-failure, task/Complete, and addressed-routing cases. The **overall run is still failure** because its later CLI parser test failed (`unknown pending human question`).
- `r2-cli-focused-01.txt` separately passes that CLI test after the parser correction. This is a narrow fix/rerun, not evidence that `04` itself passed.
- `r2-core-compile-final.txt` and `r2-gui-compile-final.txt` are compile-only passes. `r2-gui-focused-01.txt` passes selected-scope/policy, embedding, renderer-module, and relay cases. `r2-js-components-01.txt` reports the component/default/custom renderer/frame checks passing; `r2-js-syntax-01.txt` is also retained. These component checks load the student's present module graph; they do not establish the public two-route example above.
- No full grader, broad browser/native/provider campaign, installed behavior, or acceptance is claimed. No new relay or factory-provenance defect was demonstrated by this review.

### Simplicity and next scope

The existing Ensemble fleet/question owner, Agent actor terminal owner, and GUI Hub/Server composition remain coherent. R2 resolves the earlier behavioral boundaries with localized changes; there is no reason to reopen the inherited `Parent()` service pointer, demand another parent actor, or add transaction/messaging/projection frameworks. Preserve prior withdrawals (named join IDs, mandatory duplicate parent pointer, unverified blanket null-keyword defect) and the narrow factory provenance clarification.

Return the single browser-module mount correction for a bounded follow-up, then let root decide the next authorized verification stage. A focused source/receipt review with this one blocker removed would still not itself be full Chapter 20 acceptance. Reviewer retained.

Root directly verified R2 addendum and receipts after join. Factual wording correction:
spawn_sub_agent accepts `task` with additionalProperties:false and has no `prompt`
compatibility alias (internal/tool/fleet.go43–44). The review phrase 'retaining the
compatibility alias' is inaccurate, not a production defect or requested change.
Root also directly verified the downstream Question.Sender filter and two-file
public recipe. Sole correction remains the actual public module/import closure.


---

## R3/10 stable HOLD — narrow source correction accepted

**Verdict: SOURCE PASS for the released two-file module correction.** The sole R2 source blocker is closed. This is not full Chapter 20/product acceptance; further validation and acceptance remain root-owned. Reviewer remains retained.

Read the actual current `gui/web/gui/artifact-scroll.js`, `renderers.js`, application import/use sites, embedded-asset test, student note, and saved R3 receipts in `/private/tmp/e3-ch20-student.2X1OuS/`. No tests/builds/browser/native/provider operations or production/public edits were performed by this reviewer. Other R2 source/evidence is carried unchanged rather than reopened.

- `artifact-scroll.js:1` imports only `./renderers.js`. The actual `renderers.js` has **no module imports** and contains the tool-result helper, part rendering implementation, and exported render/update/remove registry. Thus the public recipe's two mounted files close the implementation module graph.
- `part-renderer.js` and `tool-result.js` are removed from the actual asset directory; no runtime imports of those former helpers remain. This is consolidation of the single existing renderer implementation, not a forwarding shim, duplicated renderer, or new build framework.
- `app.js:1,103,130` still imports/constructs/feeds `ArtifactScroll`. The saved `r3-final-inspection.txt` records exact moved-function/test-copy comparisons and byte-identical application source; the actual functions and consumer sites agree with the reviewed R2 ownership.
- `gui/ch20_embedding_test.go:89–101` now requires nonempty embedded `artifact-scroll.js` and `renderers.js`. `r3-gui-assets-01.txt` passes that focused test. `r3-js-two-file-consumer-01.txt` reports successful imports and component/default/custom renderer checks from the directory containing only those two implementation modules; its saved checks include full tool text, fold state, and speech callbacks. `r3-gui-compile-final.txt` is a compile-only pass; the app syntax receipt is silent success as recorded in the student note. These are existing receipts, not reviewer reruns.

**Limits/prose follow-up:** the component consumer uses a DOM stub, not an observed browser. No layout/audio, full grader, native lifecycle, installed behavior, or release claim follows. `gui/README.md:35–39` still names removed helper files; this disclosed stale documentation belongs in the later authorized prose pass, not another production framework correction. The asset test's old “remaining ... HOLD gap” comment is also stale prose, not a remaining module defect.

**Preserved corrections:** the R2 task tool has **no `prompt` compatibility alias**; root's annotation above supersedes the erroneous reviewer phrase. The human-question isolation concern remains withdrawn because `Hub.broadcast` checks `Question.Sender`. All earlier named-join/parent-pointer/null-keyword withdrawals and the bounded factory-provenance clarification remain intact. R2 receipt qualifications, including the overall failing combined run and separate successful CLI correction, are unchanged.


---

## After-R3 inherited fixture diagnosis and test-only review

**Verdict: PASS for the six-file test-only alignment.** No production correction round is assigned by this review. Root may resume its stopped sequential gate; neither this fixture PASS nor the focused receipt establishes full Chapter 20 acceptance. Reviewer retained.

### First root gate failure — diagnosis now recorded

The fresh root gate stopped in its first core race phase: `/private/tmp/ch20-r3-validation.0IWZvx/core-race.log` and `.exit` (`1`). Later GUI/vet/module/grader phases were unrun. The inspected log showed the reported assertion/setup failures, with no `WARNING: DATA RACE` or panic; that is not proof of race freedom or a successful broad gate.

- **Ch10 default-path setup:** accepted predecessor construction allowed two unnamed Agents given the same explicit DataDir and derived separate `DataDir/agent-N/save.json` paths. Current save derivation remains. The new collision is specifically **`DataDir/submit.md`**, included by `fleet.go:571–578` and used by submission persistence—not a guard that merely claims the directory string. Removing that guard would permit durable submission/recovery interference. Current public ownership requires independent Agent state and does not require automatic repartitioning of a shared explicit DataDir. The minimal adaptation is independent caller-owned DataDirs, preserving unnamed/default-save checks. My provisional automatic-derivation suggestion from diagnosis was superseded; no production change is warranted solely for that old setup.
- **Ch3 declarations:** inherited exact fourteen defaults predate the seven required fleet/root-submit tools. Adapt the independent expected name/count contract; do not remove production tools, accept an arbitrary lower bound, or derive expectations from the implementation under test.
- **Ch4 setup:** cloning a live `a.Config()` retained managed ID and state paths. Duplicate-ID rejection is correct; resetting ID alone would still collide at shared `submit.md`. Construct the intended fixture Agent once after the loopback endpoint exists, preserving task inputs and failure injection at actual job-output—not recovery—storage.

The diagnosis turn performed no tests or edits; this addendum records it now under the subsequent authorization.

### Actual six-file patch review

Read the current diff/source for `ch10_round1_test.go`, `chapter03_loop_test.go`, `chapter03_schema_test.go`, `chapter03_tools_test.go`, `chapter04_round1_test.go`, and `chapter04_round2_test.go`, plus the active student note and `after-r3-fixtures-inspection.txt` / `after-r3-fixtures-race-01.txt`.

- **Exact 21 defaults:** the loop and registry fixtures explicitly enumerate the prior fourteen plus `spawn_sub_agent`, `join_agents`, `agents_status`, `shutdown_agent`, `send_message_to_parent`, `respond_to_parent_message`, and `submit`. Exact cardinality, missing/duplicate-name checks, descriptions/schemas, all-provider declaration shape, ordered calls/results, file effects, empty/selected tools, and per-Agent registry isolation are retained. Expectations are not a production-derived list or `>=` allowance.
- **Explicit default submit projection:** `chapter03_schema_test.go:85–99` independently asserts exactly the object schema with string `result` and required `result` for `submit`. The other twenty schemas still require `additionalProperties:false`. Gemini still must remove only its unsupported keyword, preserve identity/order/all other schema fields, leave common definitions immutable, and leave other-provider rendered bytes unchanged. This is a specific current default contract assertion, not a blanket relaxation.
- **Ch4 single construction:** each affected fixture now creates the actual unnamed Agent once after its loopback URL exists. R1 obtains `ioDir` from `a.Config().DataDir`, then applies the mkdir/open obstruction after construction. Admission shutdown remains. The two-request failure correlation, handler/file no-effects, consumed limits, exact call IDs, one called/returned pair per call, and failed-job metadata/output checks remain. R2 retains all six requests, PTY/send-input/wait/kill/error behavior, exact output/bytes, killed reason/no invented exit code, and event pairing. No assertion was bypassed to pass setup.
- **Ch10:** both Agents remain unnamed, with SavePath omitted; each receives its own caller-owned DataDir. Distinct derived paths, successful Save, and actual file existence remain asserted.

### Evidence and unchanged-production qualification

The saved focused command runs these six tests under `-race -vet=off -count=1`; `after-r3-fixtures-race-01.txt` ends `PASS`, `ok ensemble`, `exit_code=0`. All three provider subcases and all mkdir/open/admission subcases pass. The inspection receipt contains the narrow diff, successful scoped diff check, and isolated offline launcher. This reviewer did **not** rerun tests/builds or perform any browser/native/provider operations.

The reviewed release patch changes only the six test files, and the active student record states production/GUI/browser code remains at R3. Its changes do not alter or relax production admission, state-path ownership, tools, or renderer behavior. **Qualification:** the repository-wide diff against accepted HEAD still contains the preceding Chapter 20 implementation; it is not an independent before/after-R3 byte-identity manifest. No broader unchanged-byte claim is inferred from that aggregate diff or the focused test output. Previously reviewed R3 production/evidence is carried, not freshly revalidated.

Full-gate outcome remains pending root's resumed sequence. All prior withdrawals, the no-prompt-alias correction, R2 receipt limitations, and R3 module-source PASS remain intact.


---

## Stopped GUI fixture repair — source diagnosis, not acceptance

**HOLD for a further bounded test-only clarification. No R3 production defect demonstrated by this stop; no R4 assigned.** Reviewer performed source/receipt review only: no tests, builds, browser/native/provider operations, or production/test edits. This authorized private-note append is the only edit. Focused Go GUI checks remain **unrun**.

Read the actual current three-file fixture diff (`gui/client_round1_test.cjs`, `gui/ch13_submit_test.go`, `gui/ch10_round1_test.go`), the saved `after-r3-gui-shared-js-01.txt`, current speech/pause/Continue owners, and accepted HEAD counterparts. The saved standalone Node run exits **1** at the unchanged line 68 assertion, `cleared typing must not latch subsequent queue`. This is not a passing GUI fixture release or passing GUI race gate.

### Reviewed repairs

- **Actual-source loader:** the shared fixture reads current `renderers.js`, `artifact-scroll.js`, and `app.js`; it checks the known import/export graph, adapts only those module boundaries, and evaluates the actual code in one controlled DOM realm with separate renderer/component scopes. Rendering/speech bodies are not duplicated or stubbed. DOM additions model `parentElement`, node movement/removal, and the explicitly absent optional tool pane. No source evidence links those changes to the stopped speech assertion. This is a bounded fixture adapter, not proof of browser-native module/layout/audio behavior; R3's separate module evidence remains distinct.
- **stderr ownership:** `ch13_submit_test.go` now uses the existing synchronized `safeBuffer`, not a new buffering framework. It retains stderr diagnostics, EOF/subprocess failure reporting and process cleanup. This addresses the test-owned concurrent String/exec-copy race in source; no new race result is claimed before focused Go execution. The original module error must not be hidden.
- **Recovery/output premise:** `ch10_round1_test.go` records/asserts the input LogMode and DiagnosticLimit before/after GUI construction and attachment, preserves context-only initial boundary/no-tail assertions, then explicitly selects `LogFull` before live-tail insertion and Ask/autosaves. Exact display/provenance/job/canonical/reopen checks remain. This replaces the predecessor hub's hidden policy mutation with explicit fixture-owned output policy; it does not loosen retention production or replay assertions.

### Exact speech stop: stale fixture sequence, not an extraction regression

`client_round1_test.cjs:64–69` begins with trusted typing and speech, supplies only a `state_changed` frame at a new epoch, invokes an old callback/claim, then edits the draft to `draft more` with **`isTrusted:true`**. It calls this new human input “stale typing” and subsequently requires an unpause after speech idle.

That premise conflicts with the current accepted protocol:

- `app.js:131–140` deliberately does **not** clear typing on a state acknowledgement, which might arrive after newer input. Explicit Continue authority is separate.
- `app.js:226`, `:280` clear typing for the local actual Continue action; `:139–140` handles another client's explicit `continued` authority. The server publishes that authority in `server.go:338–349`.
- `app.js:294–301` treats the fixture's new `isTrusted:true` edit as new human intent. `updatePause` (`:217–224`) therefore keeps the pause after speech ownership becomes idle.
- Those speech/typing/Continue functions are unchanged from the accepted Chapter 19 source. Chapter 13's current typing-cause and delayed-ack requirements (TL;DR 33–41; prose 115–121) support this behavior.

Thus the observed false result is consistent with a correctly retained human typing cause. The old assertion also uses `some(unpause)`, which would become vacuous if a newly added actual Continue already emitted an earlier unpause.

### Smallest justified next fixture scope

Authorize only a focused revision of this shared controlled sequence, then its previously bounded verification—not a production speech change or a new campaign:

1. Drive **actual explicit Continue** through the existing local control (or accurately model the separate external-client `continued` authority), with the matching state acknowledgement. Keep the old callback/old-epoch claim and assert they cannot revive audio, add a pause, or cancel Continue.
2. In the **no-new-human-input** branch, drive the next fresh queue and its real callback plus matching speech-idle notification. Capture pause/unpause counts after Continue and assert the new queue's own pause/release edges; do not credit an earlier Continue release to the final assertion.
3. Separately supply **new trusted input** after Continue, including a delayed state acknowledgement. Assert it acquires/retains typing pause; stale callbacks and speech idle must not release it. Clearing the trusted draft (or a later explicit Continue) must release the cause exactly as intended.
4. Keep synthetic/stale events distinct from genuine trusted input; do not relabel a new trusted event as stale merely to retain the old expected outcome.

This **does change the old contradictory sequence/assertion**. Earlier “preserve unchanged assertions” scope cannot honestly describe the next authorized repair. Preserve and strengthen the original behavioral intents—stale callback/claim refusal and eventual release absent real typing—while also preserving the newer genuine-human-typing contract. No evidence supports weakening those guarantees or restoring state-ack clearing in production.

Prior human-question false finding and other explicit withdrawals remain withdrawn. No unsupported inference that this Node failure proves a product regression, a loader regression, or a GUI race fix is retained. Student remains stopped pending root's bounded scope decision; reviewer retained.


---

## Second stopped shared-JS repair — exact chunk-boundary diagnosis

**HOLD for one further narrow fixture-content assertion correction; no demonstrated R3 production defect and no new production correction round.** Reviewed held `gui/client_round1_test.cjs`, actual `after-r3-gui-shared-js-02.txt`, current `speechChunks`/speech queue code, accepted HEAD counterpart, and current public Chapter 13. No app/test/build/browser/native/provider execution or production/test edits. The only data processing was reading/counting the two quoted values already in the saved failure receipt; this is not a new behavioral test. This note is the only file edit. SafeBuffer/full-output-policy changes are unchanged from the preceding source review; focused Go checks remain unrun.

### Revised Continue cases

`client_round1_test.cjs:64–83` now drives the actual local Continue action and distinguishes two scenarios. With no new human input, stale callbacks/old claims are inert and the fresh queue must acquire/release its own edges after Continue. With genuinely new trusted input before the delayed acknowledgement, that input acquires a new typing pause, the acknowledgement/stale work cannot erase it, speech idle cannot release it, and a real draft clear releases it. The revised counters avoid crediting an earlier Continue unpause to the later queue. These source assertions preserve the actual contract rather than the old contradictory event sequence.

The saved synchronous run reaches the later content assertion at **line 93** and exits **1**; the preceding revised assertions executed without failure in that receipt. This is not a full standalone-fixture PASS or GUI gate PASS.

### Exact content accounting

Current `app.js:33–46` `speechChunks` is the same implementation as accepted Chapter 19 HEAD: paragraph edges are trimmed; long text cuts at a word separator when possible; the next chunk begins with `trimStart()`. That whitespace is an utterance boundary, not spoken content. Chapter 13 requires readable content, meaningful segmentation, native queue progression and complete trailing delivery, not reconstructing the original string by concatenating utterance bytes with zero separators.

The test at `client_round1_test.cjs:93` supplies exactly `'complete '.repeat(450)+'🙂 final'`, drives every generated native `onend`, then incorrectly uses `utterances.map(u=>u.text).join('')` for its whole-string comparison. Inspection of the actual saved values establishes:

- expected: 450 occurrences of `complete`, one `🙂`, exact final suffix `🙂 final`, 4,057 code points;
- actual: the same 450 occurrences, one emoji and exact suffix, 4,054 code points, with three `completecomplete` joins;
- actual equals exactly four source-derived whole-word chunks containing **133, 133, 133, and 51** `complete` tokens, the last retaining `🙂 final`, joined with no separator;
- those same exact chunks joined with **one space per boundary** equal the entire original expected string byte-for-byte/code-point-for-code-point.

This is concrete fixture reconstruction drift, not evidence of truncation, lost/duplicated words, emoji corruption, missing final delivery, or an R3 speech regression.

### Smallest correction and remaining-case review

For this deliberately single-spaced, single-paragraph input, retain an **exact entire-string equality**, changing only reconstruction to `utterances.map(u => u.text).join(' ')`. Do not trim/normalize both strings, strip all whitespace, compare only word counts/substrings, or regenerate the expected value from the implementation. Exact equality with a single inserted boundary space still detects lost or duplicated words, mid-word breaks, extra internal whitespace, punctuation changes, emoji corruption, and a missing final tail. Do not bake the implementation's 133-token chunk sizes into the durable fixture; they explain this receipt, not a new public chunk-size requirement.

Keep the same callback-draining loop and all surrounding assertions: exactly one `speech_done` for the long task, no premature release between server tasks, the follow-on utterance, one owner queue pause edge, and final server-idle release (`:93–94`). The expected original text remains independently declared.

Reviewed the rest of this shared fixture in this same bounded pass. The later replay/missed-Continue case (`:96–98`), incremental artifact/final fencing (`:99–104`), and synthetic versus trusted-whitespace causes (`:108–118`) match their current source owners; no additional **demonstrated** fixture drift was found. Earlier other-tab speech-cancel and preserved human-typing cases (`:85–90`) also precede the recorded stop. This is source reconciliation, **not a claim that unexecuted post-stop assertions pass**. No unrelated files or new runtime campaign are requested.

Recommend root authorize the exact content assertion correction and the already-bounded shared/focused verification sequence. Production stays at R3; student remains stopped until authorized. All earlier withdrawals and acceptance limits remain intact; reviewer retained.


---

## Final after-R3 GUI test-only fixture review

**PASS for the three-file fixture repair; resuming root's full GUI gate is justified.** This is not a full GUI-gate or product acceptance verdict. Production remains held at R3, not R4. Reviewer performed no tests/builds/campaign/browser/native/provider operations and no production/test edits; only this private-note append. Reviewer retained for subsequent validation findings.

Read the actual current diffs/source for `gui/client_round1_test.cjs`, `gui/ch13_submit_test.go`, and `gui/ch10_round1_test.go`, the active student note, and these actual receipts in `/private/tmp/e3-ch20-student.2X1OuS/`:

- `after-r3-gui-shared-js-03.txt`: default shared controlled-DOM/callback suite passes, exit 0.
- `after-r3-gui-shared-js-chapter12-01.txt`: Chapter 12 shared mode exits 0.
- `after-r3-gui-focused-race-01.txt`: preserved **test compilation failure**, exit 1, from nonexistent `Config.DiagnosticLimit`; no behavior ran. This was fixture setup error, not a production failure. The earlier source-only approval missed this field-name defect and is qualified accordingly; it did not establish compilability.
- `after-r3-gui-focused-race-02.txt`: corrected six-test GUI race selection passes, exit 0: all three Ch10 snapshot modes; Ch13 prompt survives own release (`unpause -> input`, request/claim epoch 2, `speech_claimed`); Ch14 target drafts; Ch15 usage wire/consumer; both Ch12 websocket/stdio skill-frame cases; Ch18 model projection.

### Final source findings

- Actual-source loader retains strict mode, separate renderer/component scopes and one controlled DOM realm. Exact known imports/exports are checked, with unsupported module syntax rejected; real renderer/component/app code is evaluated rather than copied or stubbed. Small DOM parent/move/removal support and the optional tool-pane fallback are appropriate for this fixture. This is not a general module system or proof of browser execution.
- Continue scenarios now explicitly cover actual Continue authority, old callback/claim refusal, later queue-specific pause/release edges, and genuinely newer trusted input surviving delayed acknowledgements/speech idle until human clear. These are intentionally revised/stronger protocol assertions, not a claim that the old contradictory assertions stayed unchanged.
- Long-content correction is exactly `join(' ')` for the independently declared single-spaced input, still checked against the entire original string. No trimming, broad normalization, token-only/substr check or implementation-derived expected value. Native callback draining, one completion, no early release, subsequent queue ownership and final idle checks remain. The prior exact receipt accounting explains the boundary spaces; it does not replace the assertion.
- Ch13 uses existing synchronized `safeBuffer`, preserving decode/EOF/error diagnostics, stdin close, process Wait and nonzero-exit reporting. No underlying Node error is suppressed.
- Ch10 now reads the real normalized `DiagnosticMaxRecords` pointer, checks non-nil, and captures its **integer value** before GUI construction; policy checks cannot pass through a shared-pointer alias. Initial input boundary/mode/canonical/no-tail checks remain, followed by unchanged-policy checks after construction and connection. The fixture deliberately selects full OUTPUT before live facts/Ask/autosaves, retaining exact display IDs, tool/job association, reconnect/reopen and canonical assertions. Earlier Handler signature migrations in the aggregate HEAD diff predate this release.

No behavioral weakening found in this patch. The saved initial Node failures and the compile failure remain visible; the final receipts establish the bounded repair, not retroactive success of those runs. Controlled DOM/synthetic callbacks are not browser/native-audio evidence. Root's already-reported core-race-v2 PASS is carried, not rerun. Production-unmodified status is the held-release scope and student handoff, not a new byte-identity manifest manufactured from aggregate HEAD diff. The stopped full GUI gate and later sequential phases still require root's actual results. All prior withdrawals and acceptance limits remain intact.


---

## Listener `submit` collision — bounded source diagnosis

**Confirmed production integration defect; not another test-fixture adaptation. HOLD pending a bounded production correction authorized by root.** No tests/builds/browser/native/provider operations or production/test edits by reviewer; only this authorized note. Root reports GUI-race-v2, both vets and ch06/ch07/ch04 consumer phases passed. Read actual `/private/tmp/ch20-r3-validation.0IWZvx/examples-ch13-speech-test.log`: `TestListenerPeerEOFDuringAsk` exits before HTTP because `tool "submit" already registered`. The later sequential gate remains incomplete.

### Owners and contract

- `examples/ch13-speech/main.go:470–495` deliberately constructs with `Tools: []string{}` and registers exactly `hear/type/submit/wait/sleep` through `RegisterExternalTool`. Its `submit` sends the real GUI composer; it is not a fleet completion. Accepted HEAD has the same application composition. Do not rename it, bypass the real Agent, or weaken its EOF/HTTP assertions.
- `internal/tool/tool.go:50–74` now unconditionally appends the fleet catalog; `fleetTools` tests only whether its parent implements Fleet (`internal/tool/fleet.go:11–15`). Every ordinary Ensemble does. `internal/skills/skills.go:30–38,69–94` installs all catalog definitions before grants and correctly rejects a duplicate name, including disabled definitions. `Agent.RegisterExternalTool` (`internal/agent/agent.go:464–484`) reaches this registry. Empty grants therefore do not prevent the reservation.
- Public Chapter 3 permits an ordinary library embedding with no enabled default tools and makes one Agent's composition responsible for its registry. Chapter 11 protects installed compiled capabilities against external shadowing. Chapter 13 expressly requires exactly the listener's five names and guessed-name denial. Chapter 20 preserves non-fleet embedding and ordinary root chat while requiring default fleet submit and delegated terminal behavior. A blanket collision override violates these combined contracts.

### Registration alone is not a fix

`internal/tool/dispatch.go:189–197` selects an enabled entry, but then treats its name `submit` as fleet control regardless of that entry's ownership; `control` at `:301+` calls fleet admission. `internal/agent/actor.go:906–922` treats any successful return named `submit` as terminal. `ensemble.go:250–254` rejects recovery of any successful historical `submit` without a durable fleet row. Thus simply permitting the external name would still misvalidate listener `{}` arguments, bypass its ordinary handler, terminalize the listener, or reject its subsequent save/reopen.

The current schema binding is in the construction of the fleet-owned entry (`internal/tool/fleet.go:151–156`); ordinary snapshot projection copies the registered definition (`internal/skills/skills.go:570–581`). **No independent blanket schema-projection override was demonstrated.** Earlier informal “schema replacement loop” wording is corrected here: this is construction of the fleet definition, not a general rewrite loop. Keep application schema untouched and bound to its actual selected entry.

### Smallest coherent next scope

1. Resolve **installation/composition**, not collision permission. An explicitly application-only embedding (existing non-nil empty `Tools` and no skills startup selection, as used by the listener) should not install the newly optional fleet-control catalog. Keep normal/default and skill-selected fleet composition intact and preserve installed builtin/MCP/manual duplicate-name refusal. Do not infer “disabled means replaceable” globally, auto-grant disabled tools, install a second registry, or add an invented `BuiltinTools` switch. A narrowly documented decision at the existing catalog/composition boundary is adequate; root/student should make its exact public semantics explicit.
2. Bind fleet submission effects to the **resolved fleet operation**, not the string name. Existing selected-entry `Source` distinguishes builtin/application ownership; carry an explicit terminal-submission fact from legitimate fleet admission through the actor report/canonical event as needed. Use that same fact for missing-durable-row recovery. This can be one small internal provenance fact, not a routing framework. Preserve the public `Agent.Complete` path as another legitimate terminal admission. Ordinary application `submit` remains an ordinary registered job with its own schema/result and may continue chatting and save/reopen without `submit.md`.
3. Preserve exact listener five declarations and guessed-name denial, no hidden skill/shell/file grants, default fleet submit schema/terminal/durable behavior, normal root chat, and installed-name collision rejection. Do not modify the example merely to evade the shared registry defect.

After authorization, narrow regressions should demonstrate both sides: listener construction/declarations plus its existing EOF/HTTP path; ordinary application `submit` execution and subsequent turn/save-reopen; default fleet submit still terminal with correct schema/durable row, and missing-row recovery still rejected for a genuine accepted fleet completion; duplicate/compiled-name protection and zero accidental grants. This is proposed correction evidence, **not runs performed or a new campaign requested here**.

No product acceptance possible at this stop. Production remains held at R3 until root authorizes a correction round. Reviewer retained; prior fixture-only diagnoses and withdrawals are unaffected.


---

## R4 frozen source and focused-receipt review

**PASS for the bounded R4 listener-composition/submission-ownership correction.** No new blocker found in this scope. Root may resume the affected sequential validation gate; this is **not** Chapter 20 product acceptance, a full-suite result, installation, or observed running/browser/provider behavior. Reviewer ran no tests/builds/live operations and made no production/test edits. Read actual current production/test source, active student note and saved `r4-*` receipts in `/private/tmp/e3-ch20-student.2X1OuS/`. Earlier withdrawals remain in force.

### Ownership and simplicity

- `internal/tool/tool.go:51–85` has one narrow installation rule: non-nil empty `Tools`, nil `Skills`, and no `StartupSkills` omits only the optional fleet catalog. Existing core definitions remain installed/reserved; normal, named-subset and skill-selected compositions retain fleet definitions. No default grants, blanket disabled-name overrides, duplicate registry, new public switch, or second lifecycle owner. Existing Config comment and README document the distinction.
- `internal/tool/fleet.go:11,41,151–156` tags actual fleet specs `builtin:fleet` and constructs the fleet entry with its own submit schema. `dispatch.go:193–203` uses the resolved spec's source **and** action to select fleet admission. Application/compiled `submit` takes the ordinary job/handler route and its own arguments. The existing custom-tool predicate is compiled-source-specific, not `Source != builtin`; the new tag does not itself turn all fleet entries into custom handlers. Normal snapshot/schema publication remains selected-definition based.
- Registry collision checks remain unchanged. The actual example implementation remains unchanged: exactly `hear/type/submit/wait/sleep`, ordinary external registration, real pending HTTP/EOF path. The example regression adds exact declarations and the listener's empty-object submit-schema assertion without removing the HTTP/cancellation checks.

### Terminal fact, persistence, and trust

- `dispatch.go:129–130` sets `ToolData.FleetCompletion` only after successful genuine fleet admission. `internal/agent/actor.go:112–160` sets it on the accepted `Complete` result after schema/lifetime admission. The actor's report path (`:907+`) now uses that fact plus success, rather than a tool-name match, to close admission/skip remaining calls. Ordinary application `submit` neither receives fleet validation nor becomes terminal merely by its name or text output.
- `internal/common/history.go:200–201,268–270` adds the accepted fact to the existing canonical result and projected Context. `internal/history/history.go:525–528` projects it in the existing reducer. JSON clone/copy paths preserve it. This small monotone fact is justified: dialogue-only retention deliberately omits the event prefix, so scanning only retained events loses the accepted-completion guard.
- `ensemble.go:245–252` still validates the durable row with the existing reader, then refuses a missing row when recovered Context records accepted fleet completion. Full-prefix import still reduces events and checks Context equality (`internal/history/persistence.go:262–289`); the new bit participates in existing JSON/context equality rather than being ignored. Dialogue-only retains its existing authoritative-snapshot trust model. This is provenance within the library, not cryptographic authentication of arbitrary host-edited archives.
- Model/tool text cannot set this typed event: application test output literally includes `fleet_completion:true` and remains ordinary. `RegisterExternalTool` supplies application ownership; compiled registration is already a trusted in-process host API, not an untrusted model-controlled Source field. No new hostile-host or arbitrary archive-signing requirement is invented. Existing valid durable rows remain authoritative, and unmarked historical application calls are not reclassified based on a later default catalog.

### Actual assertions and receipts

`ch20_application_tools_test.go` checks real Agent turns: application and compiled submit use a deliberately incompatible-with-fleet schema, execute once, do not skip the batch tail, continue into provider and later ordinary rounds, have ordinary job results/no accepted fact/no row, and reopen without replay under application and later default catalogs. Application dialogue-only recovery is included. Boundary tests independently require exact five names, unregistered/guessed refusal, application duplicate refusal, disabled core/compiled/MCP protection, and default/subset/skill fleet protection.

Recovery tests exercise both fleet tool and `Complete`, full and dialogue-only: accepted fact, valid-row terminal restore, refusal of later work, then **missing-row rejection**. Existing child/root terminal, custom-schema correction, wakes, row failure and final-save failure tests remain checked.

- `r4-listener-before-01.txt`: preserved real duplicate-registration failure before HTTP, exit 1.
- `r4-dialogue-recovery-before-01.txt`: preserved real first-implementation defect: both dialogue-only accepted paths reopened after row removal, exit 1. This was product behavior, not setup failure; it explains the Context projection fix.
- `r4-core-focused-01/02.txt`: earlier focused race passes, including registry/Gemini projection and strengthened ownership boundaries. These do not substitute for final-source retention checks.
- `r4-core-focused-03.txt`: final-source application/dialogue ownership and recovery, catalog boundaries, full/dialogue fleet-tool and Complete recovery, actual child submit, root queue/recovery and task/Complete cases all pass, exit 0.
- `r4-core-focused-04.txt`: final-source wake/failure-resolution and custom-submit/persistence-failure cases pass, exit 0.
- `r4-listener-after-01/02.txt`: actual listener EOF/HTTP regression passes; `-02` is final source with exact declarations/schema, exit 0.
- `r4-core-compile-final.txt`, `r4-gui-compile-final.txt`, `r4-listener-compile-final.txt`: compile-only passes, exit 0, explicitly no tests run.

Focused passes address the bounded correction; they do not cover the entire history/import, core, GUI or nested-module suites. In particular R3 core/GUI/vet gate results predate R4's dispatcher/actor/canonical-reducer changes and are not a full validation receipt for this source. No new reviewer campaign requested or performed; the next gate remains root-owned. Reviewer retained for subsequent validation findings.


---

## R4 current Chapter 20 grader: bounded private diagnosis

**HOLD; mixed fixture-contract defects and one independently confirmed production startup defect.** This is root/reviewer-only fixture analysis, not student instructions. No tests, builds, browser/native/provider runs, fixture/production edits, or student contact by reviewer. Only this private note edited. Root's R4 full local gate passes in `/private/tmp/ch20-r4-validation.OaAEIt` are carried; they do not override the failures below. `grade20-v2.json` is the actual corrected run: **95/200, exit 1**. Grade 1 remains UNRUN. Initial `grade20.stderr` uses invalid `-chapter`; that was root invocation setup, not product behavior; the corrected run uses `-ch`.

### Current-only fleet fixture correction is justified

Public Chapter 20 tool examples/signatures (around 568 and 646) specify `tool_limits` for callback/wait budgets; `join_agents` and `send_message_to_parent` do not take inline `ai_callback_delay`. Actual private fixture supplies it in `ch20_fleet.go:208,284,293,299,304,415` and `ch20_submission.go:52,74`. Strict unknown-argument rejection is correct current production behavior, not a reason to widen those public schemas.

Smallest root-owned correction: at the **current driver** boundary, translate the affected scripted replies into a preceding `tool_limits` with the same delay followed by the same original operation/arguments minus that delay. Preserve the original tool-call ID, targets, payload, budget (5 seconds), order and every grading assertion. `fakevendor.Reply.Tools` already supports ordered multi-call replies; no new orchestration library is needed. Keep historical driver/call construction byte/behavior unchanged. Do not globally sanitize unknown arguments or change intentional invalid-schema/submission probes; do not alter already-valid `wait_for_job`/ordinary command callback paths. Confirm the selected batch execution respects one-shot order before release; a bounded driver regression can assert current conversion versus historical identity, but none was run here.

**Causal limits/races:** `fleetHas` recognizes a correlated result regardless of success. The scripted parent therefore advances after the rejected join as if the barrier had completed. Source and saved failure details account for the observed cascades:

- `join-both` rejects, so no joined payload can be asserted (`spawn-reports-back`, `join-fans-in`).
- Structured parent and grandchild joins reject, so invalid-submit sequence/status/parent exit can race actual child cleanup. The observed reviewer `submitting` state and absent durable row are not independent proof of a submission-lifecycle race after a successful join.
- Three-generation parent exits after rejected join; CLI reports `Agent stopped` while its descendants still need to ask/answer. The leaf's own `send_message_to_parent` also contains the invalid callback argument, independently requiring the same fixture correction.
- Idle join rejects, so its returned text cannot contain actual exhausted-child outcome. Do not relax the exact three-wake count, unresolved-failure blocking, release/visibility/shutdown assertions.

Keep exact nested payload/schema equality, separate real saves, correlated fan-in, addressed sender/content, live/removal visibility, premature-submit refusal and bounded waits. These observations do **not** prove all downstream production paths correct. There is no separate race diagnosis established by the listed failures after invalid barriers; further product/race claims must use valid synchronized calls and retained evidence. Current `fleetTemp` removes its work tree, so the report alone is not a substitute for a surviving full child history on this run.

### Reuse replay: matcher incompatibility shown; actual timeout causality still to capture

`ch20_reuse.go:277–287` sends subscribe and looks only at `guiPart(m).text`. `ch07.go:122–153` accepts a **top-level** part/flattened content, not a replay batch. Current Hub (`gui/hub.go:351–356,394–428`) sends `{type:"replay",data:{events:[actual Wire parts],partials:[...]}}`. The marker was planted before connecting, so its normal delivery is inside that replay envelope. The current probe does not inspect it there.

This is a concrete mismatch in the current fixture reader, consistent with the socket timeout. Root's completed source trace confirms that `Wire.MarshalJSON` is plain except usage: **historical** replay is `type:"replay",data.events:[]Wire`, while ordinary **live** parts remain flat. Any blanket claim that all wire frames/parts are nested is withdrawn. The planted human event is valid and RecordFact appends the real event. No reviewer runtime frame capture/run was performed.

The root-authorized current-route helper should inspect **only real rendered part content in replay `data.events` frames** for this historical marker, using `guiPart` on each actual event. Keep flat live-part/prompt assertions unchanged; do not recursively search arbitrary JSON, partials, tool arguments or unrelated metadata, or accept an empty replay. Retain real prompt delivery, no-browser correlated refusal, two-request browser relay correlation/cancellation and deadlines. Historical route remains unchanged. No causal connection to the fleet callback mismatch is asserted.

### Deployed startup: independently demonstrated production defect, still held

Root's retained probe pointer `/private/tmp/ch20-r4-validation.OaAEIt/deploy-probe-current` names `deploy-probe.r1khzd`. Read actual `startup.stderr` and `startup.exit`: exit **1**, `GUI asset missing: .../deploy-probe.r1khzd/index.html` from an unrelated working directory with default assets unset.

Source matches: `gui/server.go:110–115` calls `serverAssets(assets)` unconditionally from `NewServerWithSettings`; `serverAssets` (:95–106) resolves empty string to the working directory and requires disk `index.html/app.js/style.css`. `Handler("")` (:158–170) correctly serves embedded assets, but startup rejects before reaching it. This is not a fixture/environment/deadline defect. Do not bypass it with disk fixture assets, broaden readiness timeout, or award embedded deployment from component-only success. The deployment probe's early-exit detail should include retained stderr for future diagnosis, without replacing the real binary/HTTP/EOF checks.

Any earlier blanket source statement that empty-directory *launcher startup* was resolved is superseded: a correct embedded Handler is not sufficient evidence that the constructor reaches it. No production correction or release is authorized by this diagnosis; root owns the next bounded decision. Reviewer retained.


---

## Frozen root current-grader patch review

**PASS for this private, root-owned fixture correction only.** Read actual `ch20_fleet.go`, `ch20_reuse.go`, `ch20_contract_test.go`, current execution owners and saved final receipt. No reviewer tests/builds/grade runs, fixture/production edits or student disclosure. Production remains held R4 with the separately confirmed empty-assets startup defect unresolved.

- `chapter20ToolLimits` is a small copy-on-adapt helper, called only after the scripted route under `driver.current`. Historical routes/wrappers bypass it. It translates only singular `join_agents` / `send_message_to_parent` calls that actually contain `ai_callback_delay`. Nil, unrelated, already-multiple replies, missing-delay and malformed argument cases are left alone. It preserves the complete Reply's other fields, does not mutate the original, clears the alternative singular-call representation, and emits ordered `tool_limits` then the original operation. The original ID, action, remaining argument values and exact delay are retained; the auxiliary ID is the original plus `-limits`. No generalized argument sanitizer, policy relaxation or changed assertion.
- Actual ordering is adequate: fakevendor preserves the ordered Tools slice and independently correlates both calls. Agent dispatch prepares only `t.calls[t.index]`, waits for its report, then advances. The real setter writes `r.next`; preparation of the literal next call consumes and clears it. Thus the join/send gets the original one-shot budget; this is not a batch-preparation race. Existing run/wait callback paths remain untouched. This source trace is not a new runtime grader result.
- `reuseReplayText` leaves historical decoding unchanged. Current mode examines only `data.events` for a replay, or the actual flat frame otherwise, and requires `part_final` with rendered `text` content through the existing part decoder. It does not treat partials, raw request arguments, unrelated metadata/error strings or an empty replay as success. Existing flat live prompt/result and MCP relay/correlation/cancellation checks remain unchanged.
- Deployment only adds stderr to the same early-exit failure. It does not inject assets, relax readiness/EOF, suppress the known exit, or award embedded deployment from another check. This read occurs on the completed-process branch; the fixture still waits on its process-completion owner.
- New controls independently cover exact transformation/order/IDs/budget/remaining args, original-reply nonmutation, unchanged ordinary/tool_limits/run/wait/no-delay replies, visible flat/historical text, unchanged legacy decoding, and rejection of envelope text, tool arguments, partials and error text. The patch retains all substantive fleet/submission/reuse assertions.

**Evidence:** `/private/tmp/ch20-r4-validation.OaAEIt/grader-fixture-race-02.log` reports the focused package race run passing; sibling `.exit` is **0**. Credit only **-02** for the frozen narrowed patch. **-01** tested a broader draft while its job was still uncollected and is not final-patch evidence. No production rebuild or corrected grade rerun has occurred in this review. Root may proceed with its authorized next step; no Chapter 20 product acceptance or resolution of the deployment defect is implied. Reviewer retained.


---

## Corrected current grade20-v3: consolidated diagnosis for root decision

**HOLD.** Actual `/private/tmp/ch20-r4-validation.OaAEIt/grade20-v3.json`: **125/200, exit 1**. Source remains held R4; no production release, reviewer tests/builds/grade retries, fixture/production edits or student disclosure. Only this private note updated. Do not rerun the unchanged full grader as diagnosis. Findings below separate required production corrections from root-owned current-only fixture alignment; keep historical routing intact.

### Three demonstrated production corrections

1. **Embedded command startup:** prior exact diagnosis remains. Empty assets enters `NewServerWithSettings` -> unconditional `serverAssets("")` -> cwd `index.html` validation before embedded Handler. V3 again reports the actual early exit. Fix constructor/command composition, not timeout/assets injection; retain explicit disk-override validation and real unrelated-cwd/HTTP/EOF assertions.
2. **Completed status loses the submitted object:** public §20.8, around line 627, explicitly requires structured joins, **status**, and retrieval to preserve the complete JSON object. `internal/common/fleet.go:17–21` defines AgentStatus with only IDs/state; `fleet.go:224–244` returns exactly those fields from Children and Status. Actual `structured-status` output shows submitted reviewer/ordinary with no value. This is a genuine contract omission regardless of the valid choice of field name. Small fix: expose the retained canonical submission from the existing library fleet outcome in status as well, preserving exact nested JSON/default normalized object, ownership and terminal distinctions. No second result store/actor; no reducing the grader to state-only.
3. **Custom durable row violates an explicit public format:** current TL;DR lines 105–113 expressly requires custom rows containing `version:1`, `submit_schema`, and `payload`. This is **not** a discretionary historical field-name assumption. Actual `internal/common/submission_file.go:15–23` writes `v/timestamp/agent_id/parent_id/custom/payload` (or `result`); SaveSubmission (:91+) does not store the custom schema. ReadSubmission requires `v==1`, IDs/timestamp and validates payload only against configured schema. The fixture's row assertion must remain strict, including association with the actual child save. Correct writer and authoritative reader together to retain/validate required immutable schema, version and exact payload; preserve atomic append, all-row validation, malformed/unknown-version refusal, owner identity, default normalization and terminal recovery ordering.

**Correction/withdrawal:** my provisional claim that the custom row keys were merely historical was wrong; public TL;DR is explicit. Also the provisional `schema/data` description is not current source—the actual row is `v/.../payload`. Earlier source PASS on recovery does not excuse this missed public serialization requirement. No fixture relaxation to bless current wrong rows.

### Root-owned fixture alignment, not required production shape changes

**Join envelope and one-shot notice.** Current `JoinResult` (`internal/common/fleet.go:30–33`) has `outcomes` and optional `questions`; outcomes carry `submission` raw JSON. Public behavior does not prescribe a bare status array or `payload`/legacy `result` field names for these runtime structures. `ch20_submission.go:133–155` unconditionally decodes both join and status as historical arrays, while a current join result also has the standard `[tool_limits consumed ...]` line before its JSON. V3's `invalid character 'o' in literal true` is that decoder failure, not corrupted production JSON.

Use a narrow current-route result decoder: remove only the recognized standard consumption notice, decode the actual JSON envelope, then compare actual outcome/status submission values. Preserve exact reviewer nested equality, ordinary normalized `{"result":"DEFAULT_SIBLING_22"}`, both child identities/states, no refused-spawn visibility and original call/result correlation. Historical result decoding stays unchanged. Do not recursively search marker strings or permit missing status values. This fixture alignment will still correctly fail the production status omission until corrected.

**Addressed three-generation fan-in.** The exact V3 root `join-mid` result is successful, yielding `questions:[{agent_id:"leaf",recipient:"middle",request_id:"middle-question-1",text:"QUESTION_22"}]` with no terminal outcomes. That is the public join early-yield behavior. Yet `ch20_fleet.go:309–316` finishes the root as soon as any correlated `join-mid` result exists. It consequently ends the CLI session while the middle and leaf are still awaiting/replying; later HTTP/cancellation/Agent-stopped text is not by itself a product messaging defect.

Middle's scripted sequence is otherwise the intended exact chain: spawn leaf; question-join; `respond_to_parent_message{agent_id:"leaf",response:"ANSWER_22"}` (sole-question omitted ID is public); leaf-join; middle submit. Leaf sends QUESTION_22 and awaits the addressed response before submit. Root must continue its bounded fan-in until the actual middle terminal outcome, not end on the first descendant question. Inspect actual correlated latest result/outcomes; do not answer the middle's question from root or merely substring-search its whole prompt. Preserve original deadlines and all end-to-end checks for question content/leaf identity, actual middle sender (excluding incidental correlation-ID substrings), leaf completion and middle completion. Notice stripping is needed before structural addressed-envelope checking too. No additional production messaging defect is established by this run; valid completed sequencing is needed before one can be alleged.

**Public reuse relay.** Two independent current-fixture assumptions are shown by source:

- Probe opens TCP before browser connection, verifies no-browser refusal, then reuses that **same** TCP session after WebSocket connection (`ch20_reuse.go:256+`, then :333+). `gui/relay.go:53–63,85–101` chooses a browser at accept time; a no-browser session continues producing correlated no-browser errors. The public contract does not require late browser rebinding. Preserve the refusal check, close that session, connect/ready the browser, then open a fresh relay session for successful correlation.
- `gui/mcp.go:33–35` emits flat `{type:"jsonrpc",agent_id:sessionKey,payload:...}`. `gui/relay.go:130+` resolves browser responses by `cmd.AgentID` and checks the exact browser connection. Current fixture answers with historical `source:m["source"]` (`ch20_reuse.go:342,380`), not current `agent_id`. Adapt **only current** handshake and result envelopes to echo the actual agent/session key while retaining payload RPC ID/result. Historical source-based branch unchanged. Do not relax cross-browser/session ownership.

`Server.ServeMCP` is a raw framed relay, not the Agent MCP manager's tools/list normalizer: it forwards bytes between accepted stream and browser transport. The marker on tools/list is not yet shown to be stripped by this path; no production method-policy change is justified. Preserve actual request/response content, caller ID, session correlation, no-browser refusal, concurrent/cancel/EOF checks. No new reviewer capture or runtime retry occurred, so these are source-backed fixture causes, not a new successful relay observation.

### Evidence/setup reconciliation

Root found and corrected stale wrapper HOME/TMP and CH16 paths pointing to R3. Explicit grader CLI/GUI arguments were R4 and remain so; do not call the grader an R3-binary run merely because a temp path contains R3. The copied-wrapper issue is root setup, not a product defect. Read actual `ch16-command-r4-path-02.log` and sibling `.exit`: targeted actual `TestCh16R1HeadlessCommands` run passes, exit **0**. `-01.log` explicitly says `[no tests to run]` from the wrong regex and receives **no behavioral credit**. Full local core/GUI race, vet and module passes remain useful bounded evidence, not proof against these newly demonstrated Chapter 20 contract gaps. Grade 1 remains UNRUN.

Root should consolidate the three production findings into one justified next correction scope, with the separate fixture changes retained privately/root-owned. No piecemeal student release or full-grader retry is authorized by this review. Reviewer retained.


---

## Frozen root v4 fixture review

**Narrow fixture HOLD: one current addressed-reply assertion hole.** Otherwise source-approved for the intended current-only alignment. Read frozen `ch20_submission.go`, `ch20_fleet.go`, `ch20_reuse.go` and new contract controls. No tests/builds/grade runs or fixture/production edits by reviewer; no moving R5 production inspected as final and no student disclosure.

### One correction before fixture PASS

`ch20_fleet.go`, addressed `ask` assertion around lines 364–375: the `sender == "middle"` check and removal of `request_id` happen only when JSON decoding succeeds. On invalid JSON/plain text, the branch falls through to the old substring checks. Thus malformed text containing `ANSWER_22` and `middle` can still pass without a valid addressed reply, including attribution appearing only in malformed correlation-ID text. For **current** route, require successful addressed-object decoding before applying exact sender/content checks; reject parse failure. Preserve the historical fallback unchanged. This is a fixture assertion hole, not evidence of a product defect. Reported to root before saving.

### Source-approved portions

- `currentFleetJSON` strips only the anchored standard join/send one-shot notice and verifies its duration with `time.ParseDuration`; arbitrary prefixes/invalid durations are left to fail JSON decoding. No generic first-brace search/notice sanitizer.
- `fleetStatuses` adapts only current joins to `outcomes` and current status/outcome values to `submission`. Historical arrays/field paths remain. Structured nested equality, real null/boolean/array preservation, ordinary normalized object, terminal states, both identities and absence of refused launches remain strict. Durable custom row `version:1`/`submit_schema`/`payload` assertions remain unchanged.
- Current root fan-in uses a single `rootJoin`/count, unique sequential operation IDs, and at most 32 observations. It reads the result correlated to that current ID and only finishes on a real middle submitted outcome with non-null submission; a yielded question does not count. Final content assertion uses the **final** correlated rootJoin ID. Middle/leaf action sequence and all original question/answer/leaf/middle content checks remain. These small controller variables are confined to the root request branch; no new polling framework or second orchestration owner. Historical route still finishes by its original sequence. Focused controls reject questions, wrong Agent, failed state, missing payload and malformed prefixes as terminal evidence. No live three-generation success is claimed before the authorized corrected grade.
- Relay preserves initial TCP no-browser correlated refusal, then **current only** closes/redials after actual WebSocket/replay/prompt work. The fresh session retains deadline, request/result ID and exact marker check. `reuseBrowserReply` uses `agent_id` only for current and `source` for legacy, preserving RPC payload and avoiding both ownership fields at once; used for actual handshake and relay results. EOF/process close checks remain. No production rebinding/framing or MCP policy change.
- Pure controls cover current/legacy decode distinction, typed structured values, exact terminal outcome boundary and browser-owner envelope preservation. Small helpers are adequate and remain fixture-local.

### Evidence and limits

Existing `grader-fixture-v4-race-01.log` completed during review and reports focused package race PASS; sibling `.exit` is **0**. No duplicate run started. This receipt supports the frozen patch before the requested assertion-hole correction, not a correction not yet read/run. The initial partial root edit/assertion attempt has no test credit; corrected exact-anchor/fail-fast edit is the reviewed source. No full grader retry before student R5 stable HOLD/review. Production's three previously consolidated defects remain outside this fixture review. Reviewer retained.

### v4 attribution correction verified — final fixture PASS

Root corrected the sole review hole with `fleetAddressedReply` in `ch20_fleet.go`. Read actual helper and call site: current malformed/plain JSON returns an error; current sender must equal `middle`; `message_id`, `id` and current `request_id` are removed before the existing full content checks. A parse/attribution error fails the check and cannot fall through. Historical plain-text fallback remains unchanged. No production edits or new controller framework.

Read `TestChapter20AddressedReplyAttribution`: malformed/plain strings, a request-ID-only middle, and wrong sender are refused; genuine middle/leaf/text envelope survives with request ID removed; legacy plain string is preserved. Read final `/private/tmp/ch20-r4-validation.OaAEIt/grader-fixture-v4-race-02.log` and `.exit`: focused package race PASS, **0**. Credit **-02** for the final corrected fixture; -01 remains valid only for the prior snapshot. No reviewer test started.

**Final verdict: PASS for root v4 fixture alignment.** Prior source-approved notice/result decoder, bounded correlated fan-in, strict typed/status/durable assertions and current-only relay session/envelope fixes stand. The one fixture HOLD is resolved. No production R5 source reviewed as final, no full grader retry, no product acceptance. Root retains the agreed full-grade-after-stable-R5-HOLD/review boundary. Reviewer retained.


---

## Frozen R5 production review — three released corrections

**PASS for the bounded R5 correction scope; not full Chapter 20 product acceptance.** Read actual frozen source, the full R5 public brief at the end of `docs/edition-3-student.md`, active student note, focused test bodies and saved R5 receipts. No reviewer tests/builds/graders/live activity or production edits; no fixture-v4 replay. No new architecture or extra correction round is justified by this review.

### 1. Real launcher / asset owner

`gui/server.go:95+` now preserves the empty embedded sentinel before any cwd/path validation; nonempty overrides still use the existing explicit required-file validation. `gui/cmd/main.go:174–200` binds the HTTP listener before starting terminal/readiness, advertises its actual bound address, and uses the existing terminal drain / context / HTTP shutdown ownership. Constructor still does not implicitly listen. This fixes the demonstrated constructor-before-Handler failure rather than bypassing the deploy scenario.

`gui/cmd/ch20_launcher_test.go` invokes actual current `main` in an owned race-instrumented child from a fresh cwd, serves actual expected embedded bytes (including both component modules), covers explicit empty/default/valid override/invalid override, rejects premature readiness on failure and `:0` banners, checks normal EOF/save, and asserts zero startup provider turns. Output reads are synchronized and completion establishes exit ownership. The existing admitted-turn EOF test provides distinct loopback-provider drain/save evidence, not confused with startup-only EOF.

### 2. Complete defensive status from the retained owner

`internal/common/fleet.go:17+` adds optional `Outcome`; `fleet.go:224+` Status/Children call `statusOf` (:694+) under the existing registry lock. That helper exposes a copied outcome only for terminal entries and uses existing `copyOutcome` to copy raw submission bytes. No parallel result store, new lifecycle or mutable alias. Running/starting remain without terminal outcome; failures retain error/failed state.

`ch20_status_rows_test.go` exercises real running/custom/default/failed children, very large integer/precise decimal/booleans/null/arrays, Status/Children/actual `agents_status` dispatch/named Join, and mutates returned outcomes/bytes before re-reading the retained result. Exact-number decoding avoids float-rounding the expected payload. The **`outcome` wrapper is a legitimate public spelling**, expressly allowed by the R5 brief; assess its complete nested submission, not a private fixture's previous direct-field assumption. Root retains any necessary current-only fixture shape ownership; no production alias or extra state is warranted.

### 3. Authoritative custom row and recovery

`internal/common/submission_file.go:15+` now writes explicit `version:1`, `submit_schema`, `payload` plus owner/time metadata for custom rows. Default rows retain normalized `result` without ignored extras. Read/append share `parseSubmissions` over the exact bytes read: unknown fields/trailing JSON, blank or malformed rows, invalid versions/IDs/time, missing/invalid/custom-null schema, mixed/default/custom mismatch, and invalid payload fail. Every row is validated before any last result is returned. Each stored schema is canonicalized, its payload is validated against that schema, then the immutable configured contract is compared; a looser current configuration cannot bless invalid old content. No legacy-shape auto-migration or fallback to an earlier valid row.

Writer validates/normalizes its new value and the whole existing file before temp/write/sync/close/rename atomic replacement. It remains behind existing final conversation save and terminal completion ordering; append failure retains the invalid file as evidence and produces failed outcome without published payload. No separate persistence protocol or schema engine was introduced.

Independent tests declare the required row keys rather than round-tripping only the production struct. Actual-Agent recovery tests retain exact custom schema/value, identity, terminal refusal/no duplicate row, recovered status, later-row rejection, and refusal when only the caller's looser schema would accept corrupt content. Existing genuine fleet-submit/Complete full and dialogue-only missing-row refusal remains tested. R4 application-vs-fleet ownership is unchanged in actual dispatch/actor/history/recovery: accepted fact is not inferred from tool name/output, and ordinary app `submit` does not become terminal. R5 does not alter application-only catalog composition. Existing unchanged R4 ordinary-app/listener behavior evidence carries; current listener result below is compilation only, not a fresh behavioral rerun.

### Receipts actually read and their limits

- `r5-launcher-before-01.txt`: real pre-fix default launcher **FAIL**, missing cwd index before readiness. Preserved product evidence.
- `r5-launcher-after-01.txt`: all four actual-command modes race **PASS**, exit 0; `r5-launcher-eof-01.txt`: actual admitted-turn EOF race **PASS**, exit 0.
- `r5-core-focused-01.txt`: overall **exit 1** from new test setup using nonexistent timestamp fields / PartList as string. No Agent behavioral credit for that failed compile. The independently executed authoritative-row matrix passed; source/inputs unchanged and that useful result is retained.
- `r5-core-focused-02.txt`: real status and custom durable row/recovery race **PASS**, exit 0.
- `r5-core-focused-03.txt`: actual accepted-fact full/dialogue-only/Complete recovery, root submit queue/recovery, custom correction/persistence-failure paths, failed-delegate admission closure, and invalid-file append refusal **PASS**, exit 0. The extra `TestCh20JoinYieldsQuestionsAndHumanRoot` selector matched no test: **no fresh question/human rerun credit**. Named Join projection is separately demonstrated by the new status test.
- `r5-core-compile-final.txt`, `r5-gui-compile-final.txt`, `r5-listener-compile-final.txt`: all **exit 0**, compile-only `[no tests to run]`, not suites/listener behavior. Final inspection records scoped diff-check, not a full gate.

R4's earlier full gates are not a full validation of changed R5 source. Root may resume its authorized sequential validation and corrected public-contract grading after this stable review; no full grade, installation, native/browser behavior or product acceptance is implied. Reviewer retained.

### R5 current status-wrapper fixture follow-up — PASS

Read frozen `fleetStatuses` in `internal/gradeedition2/ch20_submission.go` and `TestChapter20StatusOutcomeWrapper` in `ch20_contract_test.go`. The small current-status-only adapter accepts a non-null `outcome` only as an object, checks exact equality of `agent_id`, `parent_id`, and `state` with the enclosing status, then exposes its typed `submission` to the **existing unchanged** full-value assertions. It does not fabricate a missing payload, accept conflicting identity/state, stringify data or add a production alias. Current JoinResult/outcomes returns by its existing branch, current direct status without a wrapper remains unchanged, and historical arrays return before normalization.

Controls positively compare nested object/boolean/null values and reject malformed wrapper plus identity/parent/state disagreement; legacy wrapper input is not normalized. This is an adequate bounded accommodation of the public-valid R5 Outcome representation, not a relaxed result check or new decoder framework.

Read `/private/tmp/ch20-r5-validation.vTRrtJ/grader-fixture-race-02.log` and `.exit`: focused package race PASS, **0**. Earlier preserved `grader-fixture-race.log` reports `undefined: reflect`; that was root test setup/compile failure and receives no behavioral credit. Actual test import is now present. Reviewer did not rerun tests, inspect/change production, or replay a grader. **PASS for this narrow fixture delta only**; R5 production HOLD/source verdict and pending full validation remain as previously recorded. Reviewer retained.

### R5 final two adapter mismatches — diagnosis and fixture PASS

Read actual `/private/tmp/ch20-r5-validation.vTRrtJ/grade20.json`: **170/200, exit 1**, exactly `child-isolated` and `addressed-messaging` fail. Grade 1 remains UNRUN. These two details identify fixture mismatches, not new production defects; no production edits or reviewer tests/grades/live activity occurred.

- The durable row failure attempts `.../children/reviewer/save.json`. Actual current `ensemble.go:166–177` derives a default save as `DataDir/agentID/save.json`, while row ownership is `DataDir/submit.md`. New `fleetSubmissionSavePath` (`ch20_submission.go:277+`) first requires current `row.agent_id == "reviewer"`, then uses the fixed `rowdir/reviewer/save.json`; it never trusts an arbitrary row ID as a path. Legacy remains `rowdir/save.json`. Caller retains **unchanged exact custom version/schema/typed-payload checks**, JSON decode, and exact childTask dialogue assertion binding the row to the actual child's save. This repairs the owner path; it does not search arbitrary saves, bypass isolation or relax the row contract.
- Actual `common.Question` (`internal/common/actor.go:21`) has exported `Text, Sender, Recipient, RequestID, TurnID` fields without lowercase JSON tags; the failed reply actually contained `Sender:middle`, `Text:ANSWER_22`. Final current `fleetAddressedReply` accepts exact `sender` or `Sender`, requires every present alias to be a string, refuses conflicting aliases, and still requires exact `middle`. Malformed/missing/wrong attribution fails. It removes actual upper/lower request/turn IDs and existing ID/message-ID variants before the unchanged content assertions, so correlation alone cannot supply the answer/attribution. Legacy plain/lowercase handling remains unchanged. No production aliases required.

Read focused controls for uppercase typed reply, wrong/missing/conflicting/non-string (object/array/null) senders, correlation-only answer marker, current/legacy save paths and wrong owner; prior lowercase/malformed controls remain. Recommended two small guardrails during review—TurnID removal and type checking before alias comparison—are present in the final source. No broad normalization, ID-based fallback, new schema or correlation framework.

Read final `grader-fixture-race-04.log` and `.exit`: focused package race PASS, **0**. Credit -04 for the final guarded source; no new reviewer run and no credit to earlier snapshots as final evidence. **PASS for these two root-only fixture corrections.** This does not turn the recorded 170/200 run into a pass, does not claim a new full grade, and does not reopen unchanged R5 production. Full local gates were reported PASS by root; this bounded turn did not rerun/audit them. Reviewer retained.

### R5 185/200 join-tail diagnosis and two-response fixture sequencing — PASS

Read actual `grade20-v2.json`: **185/200, false**, only addressed messaging fails. The exact paired `question-join` result has `is_error:true` and `interrupted, not executed: live input interrupted this pending batch`, after the standard consumed-limits notice. That is not a successful join missing its payload; the fixture's setter+operation batch exposed its dependent tail to legitimate Ch5/6 live-input interruption. Production remained held/unchanged. Earlier approval of ordered same-batch preparation did not establish atomic execution against input; that sequencing assumption is superseded, not a new production race finding.

Read final `chapter20ToolLimits(body, reply)` (`ch20_fleet.go:105+`), current-only route invocation, existing exact `anthropicToolText` pairing owner, and updated `TestChapter20CurrentToolLimits`:

- Applicable only to current authored join/send-to-parent requests with the callback argument. First response contains **only** original-ID-plus-`-limits` `tool_limits`, preserving the exact requested delay. No dependent batch tail.
- Only a successful `tool_result` correlated to that exact limits ID permits the later response containing **only the original operation ID/name**, with the callback removed and other arguments preserved. A different ID cannot authorize it. This uses existing transcript/result ownership, not a scheduler, timer or new retry state machine.
- Exact failed setter produces an explicit controller failure/stop response and **no operation or retry**. Missing original operation/question result still fails existing assertions; no skipped join is relabeled success. This is deliberately fail-fast rather than repeatedly issuing a failing setter with the same ID.
- Authored reply is copied rather than mutated. Legacy never enters this adapter; explicit tool_limits, wait/run, non-callback join, multi-call/unrelated/text/nil replies are untouched. Existing question-join success/error check, exact question/leaf content, genuine final fan-in and addressed-answer checks remain intact.

Controls independently cover separate responses, setter delay, original operation ID/name/remaining argument, original immutability, wrong result ID, setter failure no authorization/retry, and unrelated calls. Read final `/private/tmp/ch20-r5-validation.vTRrtJ/grader-fixture-race-05.log` and `.exit`: focused package race PASS, **0**. No reviewer tests, grades, production edits or broader campaign.

**PASS for the current-only fixture sequencing fix.** The recorded full grade remains 185/200 until a new authorized corrected run; no new full-grade result or product acceptance is claimed. The actual join-question success assertion remains required. Reviewer retained.

### R5 legal predispatch interruption — bounded fresh join observation PASS

Read actual `grade20-v3.json`: still **185/200, false**, addressed messaging alone fails on exact paired `question-join` error `interrupted, not executed: live input arrived before dispatch` (preceded by valid standard consumed-limits notice). The previously source-approved two-response setter/operation split did **not** eliminate legal predispatch interruption. Preserve that failed result; do not claim the earlier fix solved the scenario or treat this as a demonstrated production defect.

Read final current-only correction in `fleetMessagingWith` (`ch20_fleet.go:349–410`) and `fleetUndispatchedJoin` (:582+):

- One local `questionJoin` ID/counter, maximum **four** observations. Only exact successful extraction of that ID's paired result, with **is_error true**, plus either of the two literal runtime undispatched-error texts after the already strict standard-notice decoder, permits a fresh uniquely identified join observation. Arbitrary prefixes, unrelated failures and error=false do not qualify.
- This re-observes only a read/wait fan-in operation that explicitly never dispatched. It does not retry spawn, submit, reply or other side effects, does not generally retry failed operations and adds no scheduler/backoff framework. Current guard only; historical question-join sequence unchanged.
- Bound exhaustion explicitly fails the check. Other errors are not re-observed and remain failures in final validation. Final original assertions now correlate to **final questionJoin ID** and still require success plus QUESTION_22 and leaf. They also retain actual attributed ANSWER_22/middle, terminal LEAF_DONE_22 and final root correlated MIDDLE_DONE_22. An interrupted/skipped join is never accepted as successful question/fan-in evidence.

Read pure classifier controls (positive exact predispatch text and rejection of error=false, unrelated error and prefixed text) and optional `TestChapter20AddressedMessagingCLI`: absent CH20_TEST_CLI skips; with explicitly pinned CLI it runs the actual complete current three-generation check and fails on any check detail. Root reports job2904 ran it once with real pinned R5 CLI alongside controls. Read `grader-fixture-race-06.log` and `.exit`: focused package race PASS, **0**. Log is package-summary/nonverbose; the pinned-CLI invocation fact is root-provided, not inferred from a generic green log or an unset optional test. No reviewer test or grade started.

**PASS for this narrow fixture correction**, supported by source, focused controls and root's pinned actual-scenario run. No production edit/release, broad campaign or new full-grade result. Recorded full v3 remains 185/200 until a later authorized corrected grade. Reviewer retained.

### Bounded pre-live control review — preparation HOLD (no live activity)

Read `/private/tmp/ch20-live.MILYP8/{plan.md,run.mjs,primary.md,prompt.txt,preparation.json,check.json}` and only relevant current CLI/tool owners. No reviewer harness/check execution, key/settings read, process launch, network/provider/browser/native activity or production edit.

Two concrete pre-live corrections:
1. Prompt asks for nonexistent `reply_to_child`. Pinned current tool is **respond_to_parent_message(agent_id,message_id,response)** (`internal/tool/fleet.go:106`). Correct actual tool/argument instruction, then regenerate/check the pinned preparation before any live use.
2. `rootRequest` (`run.mjs:24`) currently authorizes **any** user message containing the full prompt as a substring. That is weaker than the claimed root-request-only provenance: a child task copying the root prompt could authorize a remaining permitted spawn. Small adequate correction: exact initial/root user-task comparison, plus refusal to delegate a child task containing that root identity/signature. Keep current allowed-three-ID/duplicate admission and atomic shared Set; no new authentication/framework machinery requested.

Plan wording also needs to agree with the deliberately held independent-execution boundary: paragraph2 still says root independently runs assertions after exit, whereas controller correctly only writes `independent-pending.txt`. State prior source inspection and a separately root-controlled execution step; do not count independent Python as performed.

Otherwise source shows bounded atomic 32 POST reservations, 512KiB/request and8MiB total request cap, bounded responses, one exact-model availability GET, TLS verification and no automatic redirect/retry/body rewrite, original admitted bytes forwarded, per-attempt marker plus binary/controller/primary/prompt/plan pins, owned process/listener/upstream cleanup, explicit owned HOME/TMP and dummy child key. Real key is read only in live relay memory and suppressed from saved request/response/CLI output if detected. Current shared CLI flags support sandbox/no-web-search/save/chat; DataDir=cr and child paths match `work/cr/children/<id>/submit.md`.

Hint disclosure is correctly limited: a CLI join proposal followed by150ms while no root completion observed is **not dispatch/wait proof**. Require paired results and causal saved-history review; skipped join cannot prove waiting. Existing check receipt says live=false, keyReads=GETs=POSTs=0, no launches/spawns. It proves preparation only, not actual OS confinement, live causality, successful durable payloads or independent arithmetic. No live PASS yet; bounded HOLD pending the above control/text corrections. Reviewer retained.

### Corrected frozen live preparation — PASS (supersedes pre-correction HOLD)

Re-read corrected six files under `/private/tmp/ch20-live.MILYP8`. No reviewer harness/check/live execution, key read, process/provider/network effect or production change.

Prior findings resolved in actual source:
- Prompt now uses current `respond_to_parent_message(agent_id, message_id, response)`.
- `rootRequest` examines only the **first user message**: first text block must exactly equal root task; string representation must start with that task to allow documented concatenated request-local ephemera. It no longer accepts a later message or prefixed quotation. Admission additionally requires a string child task and rejects any child task containing the root prompt, so delegated first-task content cannot carry that identifying prefix. Exact three allowed IDs and duplicate/global admission bounds remain. Read adversarial controls for later-user spoof, prefixed quotation and copied child task, alongside existing no-root/wrong-ID/duplicate controls.
- Plan now explicitly puts independent Python execution after root's source inspection; controller still only writes `independent-pending.txt` and performs no such execution itself.

Read fresh pinned preparation (controller `2aa92db69468ca9c4e6214c34bc53c7ba552f818348b37cf6e0bc58b407649b3`, corrected prompt/plan pins, unchanged R5 CLI SHA `79fc1b2a072d0494694455d5a99664d5019df61088e2fda2f41de19fbcd6ee6c`) and check receipt: ok=true/live=false, zero keyReads/GETs/POSTs/requestBytes, empty launches/spawns/responses. Controller refuses identity drift before credential/network/process activation. Reviewed earlier finite budgets, request/credential handling, actual CLI flags/state paths, owned cleanup and honest hint disclosure remain intact; no new runtime/framework machinery.

**PASS for this corrected bounded preparation; earlier pre-correction HOLD is superseded.** No remaining concrete pre-live blocker found in this scope. This supports root proceeding only within Bill's already-authorized one finite attempt; reviewer performed no activation. It is not evidence of successful live completion, actual confinement, successful join waiting, delivered hint causality, valid submissions or independent arithmetic. Those still require the owned live receipts/canonical causal review and prior-source-inspection boundary already stated. Reviewer retained.

### Actual browser stop after R5 — production module-entry defect; narrow R6 diagnosis

Read `/private/tmp/ch20-r5-reuse.eSQXBY/browser-evidence.json`, current student stopped-HOLD note, actual `gui/web/gui/index.html:3` and `app.js:1`, and zero-request/normal-exit receipts. No reviewer browser/test/build/production action; no moving live-controller review.

**Demonstrated production defect:** actual embedded shell uses `<script src="app.js" defer></script>` (classic script), while the actual app starts with an ES-module import of ArtifactScroll. Real Chrome fetched the owned HTML/CSS/app, reported `Cannot use import statement outside a module`, then timed out waiting for the seeded Conversation replay. This is not a classic-VM fixture that needs another adapter: the real public page itself has the wrong entry semantics.

Smallest coherent R6 scope:
1. Mark the actual app script as `type="module"`; preserve real app/component/renderer owners and module graph. Module scripts already defer evaluation by default. No classic conversion, injected test loader or alternate app.
2. Add a focused regression against HTML returned by the **public embedded Handler**, asserting the app entry really is a module. File/module existence and transformed VM unit execution are insufficient to protect this boundary.
3. After reviewed correction and rebuilt embedded consumer/artifacts, resume only the root-authorized bounded actual-browser exercise on served production files. Require visible seeded replay before crediting later stages; retain this failed attempt. Do not manufacture browser success by evaluating/rewriting source in the harness.

Evidence limits: `requests.json` is null/empty, `consumer.stdout.log` reports `OWNED_SHUTDOWN requests=0`, evidence shows consumer exit0/no signal and empty relay transcript. Independent override HTTP stage completed, but no successful Conversation, model/engine browser turn, relay roundtrip, model-next-wire verification or component-page interaction was reached. Earlier module-closure/source PASS and green transformed-unit/gate results did **not** establish real shell execution; this failure supersedes any broader inference. No full product acceptance.

Separate root live-control observations are kept separate, not diagnosed as production defects here: root reports MILYP8 exhausted32 POSTs with no child/hint/normal EOF because the custom primary lacked actual task-tool grants despite a deferred catalogue; later key-free grant preflight had actual grants but its controller missed `you> [2 turns...]` and root stopped only its owned preview CLI. Prior preparation PASS was not runtime grant or completion-parser evidence. No controller source review/release or paid follow-up in this turn; live remains held until R6 per root. Reviewer retained.

### Corrected-grants controller and key-free wire preflight — control PASS; activation held for R6

Read only `/private/tmp/ch20-live-grants.bBUDYB` plan/primary/prompt/controller/preparation/check/wire-check, actual `wire-request.json` and preserved failed-prefix receipt. No reviewer controller/test/process/key/provider/network effect; no moving R6 production review.

- Primary now explicitly grants the finite task tools. Actual pinned-CLI wire request has exact Sonnet4.6, max_tokens1024, nonstreaming representation (stream omitted), and required spawn/join/parent question+reply/submit/file/command/wait/limit tools **present and non-deferred**. This is actual grant evidence, not inference from a deferred catalogue. Mandatory management definitions also exist; provider-response allowlist deliberately refuses calls outside the finite exercise set (including speculative skill loading) without rewriting original admitted request/response bytes.
- Every actual route request passes the grant/model/byte/count guard **before paid reservation and authenticated upstream request**. One wire-check request is answered locally; a second is refused. Live mode alone can read the approved key and perform the bounded availability GET; dummy child credentials/owned environment remain unchanged. The existing 32POST/1GET/1024token/response/request/time bounds are not expanded; no retry/redirect/substitution introduced.
- Explicit mutually exclusive mode, binary/full preparation checks, and successful matching controller+primary one-request normal-EOF wire preflight gate live mode. Live one-attempt marker and owned cleanup remain. First-user identity/child copied-signature refusal, exact allowed child IDs/duplicates, and prior-source-inspection/independent-pending boundaries are intact.
- Completion regex now admits actual optional `you> ` prefix. Preview gets30s for root completion, live840s, with separate bounded EOF/outer allowances. Hint remains proposal-only and requires later actual paired/canonical causal review.

Evidence: corrected controller/preparation/wire/check pins match `b053a0a78c5f3955606401907472ba936cece1e163c4f29b55c0feccf57c5b95`; primary `6919cdd8...`; same R5 CLI SHA `79fc1b2a...`. Fresh check has zero effects. Wire-check has synthetic=1, keyReads/GETs/paidPOSTs=0, normalExit=true, CLI exit0/no signal, no children/hint, owned preview02. Preserved old failed-prefix receipt remains false, synthetic1/zero keys/upstream, no normalEOF, owned SIGTERM cleanup; it is not recast as success. Earlier paid MILYP8 exhaustion remains a separate failed attempt, not automatically retried.

**PASS for this corrected root control and saved key-free preflight, not activation or live completion. Activation remains HOLD until R6 production review per root.** If R6 is GUI-only and this exact CLI/source identity stays unchanged, this checksum/wire evidence can carry; no pointless CLI rebuild/preflight demanded. Live grant/EOF readiness does not prove future children, waiting/hint causality, valid submissions or independent arithmetic. No review/acceptance of moving R6 source here. Reviewer retained.

### Frozen R6 module-entry correction — bounded source PASS

Read stable student handoff, actual one-line `gui/web/gui/index.html:3` diff, `gui/ch20_embedding_test.go`, and all named r6 receipts under `/private/tmp/e3-ch20-student.2X1OuS`. No reviewer tests/builds/browser/provider/process or production edit.

Only production change is classic/defer → `<script type="module" src="app.js"></script>`. This is the correct browser boundary for the existing imports; modules defer by default. No loader, bundle, renderer duplication or app rewrite added. Saved inspection records app/artifact-scroll/renderers byte-identical to prior held inputs; core/CLI and other owners remain unchanged in this release.

Public external-package regression (`ch20_embedding_test.go:38–63`) obtains HTML through actual `gui.New(...).Handler("")`, requires exactly one module app entry, then retrieves all three actual modules through Handler and checks bytes against exported Assets and valid JavaScript MIME. Existing override, relay-correlation, close and Agent-lifetime checks remain. It does not use the classic VM adapter to manufacture browser entry semantics.

Evidence read: `r6-served-before-01.txt` fails on exact served classic tag, exit1; `r6-served-after-01.txt` public embedding/relay and reusable-module tests PASS under race, exit0; `r6-esm-syntax-01.txt` all three actual modules parse, exit0; `r6-gui-compile-01.txt` GUI and command compile-only pass, no behavior-test credit. Final inspection preserves old Chrome failure, unchanged JS evidence and clean scoped diff. No actual browser continuation yet.

**PASS for the frozen minimal R6 correction.** No source blocker found to root's bounded rebuilt-external-consumer actual-browser continuation and already reviewed fresh corrected-grants live plan. Reviewer performs neither activation. Rebuild the GUI-bearing consumer to include corrected embedded HTML; unchanged pinned core/CLI and valid earlier evidence may carry—no pointless CLI rebuild/full grades/lifetimes requested. Existing R5 browser failure remains valid for that artifact, not erased. New real-browser replay/relay/component behavior and live question/join/hint/submission/arithmetic causality still require their actual operational evidence; no full product acceptance implied. Reviewer retained.

### Completed bounded live arithmetic fleet — qualified success; one exact local wait/hint check remains

Read actual `/private/tmp/ch20-live-grants.bBUDYB/receipt.json`, canonical root/child full saves, all three submit rows, all five generated Python sources, derived causal summary, saved provider request bodies and independent assertion log/exit. Receipt-only/data inspection; no generated-code execution, rerun, credential/process/provider/native/browser effect or production change by reviewer.

Verified completed live work: 19 POSTs/all200, one discovery GET, one key read, exactly add/multiply/subtract spawns, normal EOF/exit0/no signal. All three immutable default rows have correct child/parent identity and result text. Generated arithmetic modules/combined file and15-case root test are ordinary arithmetic code. Root command output actually contains `All assertions passed.` and explicit `EXIT:0`; `; echo` alone could mask Python failure, so do not rely solely on its shell exit. Separately root-operated `independent-assertions.log` reports18 cases including negatives/zero with independent exit0, after source inspection. Those are separate evidence, not silently attributed to the model. Native `/private/var/select/sh` warning is preserved, followed by actual successful assertions; no access broadening performed to suppress it.

Canonical causality (not inferred from CLI prose):
- Root seq20 receives multiply's addressed question `agent-1-question-1`.
- First join seq25/26 is paired `is_error=true`, explicitly undispatched, **before** terminal hint seq28. Therefore the later human hint did not interrupt that executing join; it never executed.
- First reply seq31/32 is also paired undispatched/error. Later reply really dispatches seq37 and returns `Answer delivered.` atseq39.
- Multiply's actual wait_for_job job6 returns Text yes/support negatives, Sender=agent-1, Recipient=multiply, same RequestID. This establishes real addressed delivery, not merely an attempted reply.
- Actual join result seq49 succeeds with all three state=submitted complete payloads.
- Exact human hint reaches provider requests10,12,16–19. All24 causal-summary event owner/seq/type/time references match canonical saves; inspected critical bodies directly. Delivery is established. Negative tests are real, but the original task already requested negatives, so do not invent stronger causal attribution to the hint.

**Qualification required:** public §20.13 says send the terminal hint while join is waiting and follow the same running handle to eventual fan-in. This run does not establish that exact observation. A proposal, tool_called, or skipped join is not a running wait. Existing `TestCh20JoinQuestionsAndHumanSink` covers addressed-question yielding; `TestOpenAIBusyHintToolAdjacencyHTTP` uses another held tool. Neither is silently renamed exact active-fleet-join/human-hint evidence. Root confirms no existing exact proof located.

Recommendation agreed with root: retain **qualified live success**, no additional paid attempt and no acceptance campaign. After the student’s current browser exercise reaches stable HOLD (not concurrently), release **one bounded local loopback actual-Agent check** of real join job/wait already running → terminal hint wakes observation without killing child/job → same original handle later collects submitted result. Use actual dispatched/wait state and correlated results, not timing a proposal; controlled held child/release and finite bounds suffice. This specific public behavior remains pending evidence, not a demonstrated production defect. Reviewer runs/writes no test now.

Do not erase MILYP8's failed32POST missing-grants attempt or the key-free prefix-parser failure. Successful19POST follow-up is separate. GUI reuse is still pending independently. No full chapter/product acceptance in this assessment; reviewer retained for the exact remaining evidence review.

### Post-R6 actual browser/public reuse — bounded synthetic operational PASS

Read stable student handoff, `/private/tmp/ch20-r6-reuse.cBgKMg` consumer/browser/demo sources, browser evidence, actual raw requests/relay/stage records, final inspection, exit/stdout logs and two archived setup receipts. No reviewer run/build/browser/provider/process/production action. No screenshots visually inspected.

Observed evidence supports the stated bounded exercise:
- Rebuilt outside consumer `ec8035d8...` uses actual public Ensemble/Agent, gui.New/Handler/Assets/ServeMCP with owned listeners, workdir/state and EOF cleanup. Actual Chrome loads ordinary R6 modules, not a VM/source substitution. Seeded PUBLIC_REPLAY_SYNTHETIC is visible, page errors/console empty, and two real composer turns render their synthetic Agent/Engine responses. Separate override returns its marker without replacing embedded default.
- Two actual TCP peers initialize and send initialized notifications, then concurrently reuse id77 for different methods. Raw transcript shows peerone id77 returns tools; peertwo id77 returns real gui_snapshot content with selected-agent/header/connected-DOM information. Omitted conversation content is not claimed as snapshot evidence; replay/turn visibility has its own actual DOM assertions.
- Disconnect of peerone with observation88 pending ends its local wait `caller closed`; peertwo subsequently succeeds on id89 and browser/Agent then complete both model turns. This establishes caller/session isolation and survival; do not inflate it into proof that every browser-side synchronous action was internally aborted. Transcript has13 messages and no cross-session id substitution.
- Real GUI model control changes gpt-6.1-sol to gpt-5.6-sol, and the next actual raw request uses gpt-5.6-sol. Both wires preserve configured medium/concise reasoning and nonstreaming. There is no reasoning-effort UI change claim: inspected control inventory contains no such separate control.
- Independent component page actually imports ArtifactScroll and consumes receive frames; browser graph is precisely demo plus artifact-scroll.js/renderers.js. app/old helpers return404. Three initial cards render literal unsafe-looking text and inert image/audio locators, with no img/audio/video/link elements or external media/playback. Actual 220px viewport follows content (top1288/height1508), stays at0 while reading after append, shows Latest and moves downward on click. Minimal page exercises nested final frames; full GUI separately observes its streaming/frame flow.

Bounds/cleanup: original total583.66s <900; total synthetic requests2/16 across0+0+2 segments; final browser-run03 exit0, consumer OWNED_SHUTDOWN requests=2/exit0/no signal. Owned profiles' singleton lock/socket/cookie entries absent; no reported non-owned browser request. Old headless import evidence may carry for unchanged core; no new broad gate inferred from this exercise.

Preserved zero-request setup failures are correctly classified: missing MCP initialized notification produced real -32002 not-initialized; later snapshot assertion wrongly demanded text omitted by its bounded format. Current driver adds required notification and checks actual bounded DOM snapshot; independent replay assertions remain. Each failed segment retained its artifacts/normal owned cleanup. Earlier R5 classic-module production failure also remains preserved, now addressed by R6 plus actual browser evidence.

**PASS for this bounded actual local-browser/public-reuse exercise, with declared limits.** Synthetic endpoint, not live model/deployment/full gate; screenshots saved but no beauty/pixel-layout verdict; no reasoning-UI implementation claim. No new campaign recommended. Separate qualified19POST live arithmetic assessment stands; commissioned actual-Agent active-join/human-hint/same-handle check is still pending and is not silently covered here. No overall chapter acceptance yet. Reviewer retained.

### Final dispatched-join test and code/operations closure recommendation

Read full `solutions/edition-3/main/ch20_dispatched_join_hint_test.go`, stable student handoff, `hint-join-race-01.txt` and `hint-join-final-inspection.txt` under `/private/tmp/e3-ch20-student.2X1OuS`. Reviewer made no test/build/grade/provider/browser/process or production change. This is test-only; frozen production remains R6, no R7 correction.

**Focused test PASS; no source blocker.** Test reuses actual Agent actors/engines, actual join handler/job owner, ordinary Hint and real child submit. Channel-held child/provider response is a controlled barrier, not a sleep or production hook. Before hint it requires ToolDispatched and canonical dispatched release/owned running job, rejects any prior returned join, and correlates turn/call/handle. It then requires exact human HintObserved and canonical release<hint<non-error running report; next real model request contains hint, parent resumes, same join job and child still running. Only then does it release the child. Genuine validated submission must have exact typed payload; Get(handle) must return the same job object, and Complete(ctx) must return done/non-error/exact submitted child/parent/payload with no pending questions. Exactly2parent/1child loopback requests, bounded context/cleanup, no new scheduler/framework.

Saved race run passes in0.31s (package1.340s), exit0: release10/turn1/handle1 → hint11 → running non-error report12 → samehandle done/full payload. Observation limit60s versus test context20s excludes ordinary observation-timeout explanation for that wake. Affected compile-only receipt passes without behavior-test credit. Scoped inspection/hash and preserved browser evidence recorded. This proves one precise **local actual-Agent** ordering, not every scheduler interleaving and not that exact timing in the earlier paid run.

Reconciliation of accepted evidence:
- Frozen R6 source review accepted minimal module entry and actual served-HTML/assets regression; all earlier ownership/schema/lifecycle/GUI fixes and explicit withdrawals stand. Production unchanged by final browser exercise and this test.
- Read saved R5 core/GUI race, both vet, ch06/ch07/ch04consumer/ch13speech exits: all0. Valid unchanged-owner results carry; no claim these broad runs were rerun after R6's HTML-only edit. R6 focused served bytes/syntax/GUI compile and actual browser cover its affected entry path.
- Parsed `/private/tmp/ch20-r5-validation.vTRrtJ/grade20-v4.json` =200/200, passedtrue, exit0; `grade1.json` =100/100, passedtrue, exit0. Root identifies the corrected grader as buildv8; do not confuse grader build version with grade20 result-file suffix. Prior fixture/setup failures remain preserved, not replaced by fabricated fresh receipts.
- Actual rebuilt external consumer/Chrome exercise: real modules/replay/turns, same-ID relay correlation and disconnect isolation, next-wire model, actual two-file component scrolling/inert rendering,2synthetic requests/583.66s/normalEOF; declared reasoning-UI and unseen-screenshot limits remain.
- Paid bounded arithmetic follow-up:19POST/all200,3valid submissions, real addressed reply and successful final fan-in, normalEOF, separate source-inspected independent18casePASS. Earlier32POST missing-grants failure and key-free prefix-parser failure remain. The paid first join was skipped before the hint; that is **not** rewritten as activejoin interruption. Live CLI hint ingress/provider delivery plus this separate controlled local mechanism proof close the remaining requirement honestly, not by claiming one trace showed both.

**Independent recommendation: ACCEPT the frozen R6 Chapter20 code and completed bounded operations, with the stated evidence limits.** No remaining code/operations blocker identified; no further validation campaign, paid retry, broad rerun or new framework recommended. Root retains final acceptance/snapshot authority. No claim of installation/deployment, native/device breadth, visual beauty, reasoning-UI change, exhaustive race freedom or exact paid-run join timing. Minimal fresh prose reconciliation is the next separate phase, not an unperformed code operation or reason to reopen completed work. Reviewer code context can close; retained result and all prior qualifications/withdrawals are saved here.
