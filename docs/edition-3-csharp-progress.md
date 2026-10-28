## 2026-10-28 — Chapter 20 bounded completion corrections verified; joins/tools unfinished

I connected idle-unsubmitted correction to the existing actor's successful delegated-turn finalization, not a timer or second task loop. `Common/ChildSubmission.cs` declares the consecutive correction count and current failure; `Agent/ChildCompletion.cs` implements them. `Finish` retains the same active request and reliable reply while up to three immediate framework corrections enter canonical History and the next provider request. Corrections name unresolved direct children and require joining or shutting them down before submission. Plain final text never publishes completion. Root chat, direct host tool calls, interrupted/failed turns, and older turns superseded by newer admitted work do not receive these retries. Exhaustion records an explicit error and fails the turn. Successful durable submission resets the correction budget; new accepted work clears old failure but does not reset the consecutive-correction count. No model-facing submit tool exists yet, so tests publish through the real public library boundary.

`ChildFailure()` exposes the current delegated task failure without inventing a payload or equating failure with quiescence. `WaitForSubmissionAsync` now fails promptly on that retained failure. Actual transport/parser errors and actor-fatal errors also retain a current-generation failure; ordinary shutdown does not erase it. Older generations cannot poison new work. A successfully published submission remains authoritative during later draining; the reliable turn reply still reports a later execution error. This slice does not implement question-aware join, completion/task limits, the child tool catalog, CLI human-question controls, or a completed child workflow.

Three real-component/loopback groups cover exactly four plain-text requests (initial plus three corrections), durable reminders/error, no false submission, failure observation, accepted recovery, two consecutive tasks each using all three corrections with reset on durable publication, unresolved child naming, early publication while model work remains pending, malformed-provider failure without retry, ordinary root chat, and failure retention through shutdown. Final `completion-tests-build-3.log` and `completion-tests-3.log` pass 3/0. `completion-shared.log` passed **173/0** on the final production source; the subsequent changes are declaration comments, README and an additional shutdown-retention assertion covered by the final focused run. Both actual hosts built with zero warnings/errors in `completion-cli.log` / `completion-gui.log`; fail-fast chain4064 exited0. No new executable grade or preserved endpoint is claimed.

Setup corrections are retained: staged `completion-edit.py` saved submission declarations/resets then refused an ambiguous actor anchor; I completed the actor patch against its exact fatal block. My next command used a repository-relative output path from main, so the implementation file was not created and `completion-build.log` failed; I wrote the correct file, and `completion-build-2.log` was clean. No product failing assertion occurred in this slice. A declaration-comment replacement missed its actual wording and was corrected explicitly. Reviewed the complete new completion implementation, affected actor finalization/fatal blocks and submission publication/reset path. Scratch remains `/private/tmp/ch20-csharp.hH3ysE`; no workers, shared-environment changes, live OAuth/provider, commits or restarts. Chapter19 remains the preserved endpoint. Next: question-aware joined waits using existing Jobs lifetimes, then remaining task limits and ordinary child tools/CLI, scoped grader and final prose/source review.

## 2026-10-28 — Chapter 20 reliable submission waits and question ownership verified; joins/tools still unfinished

Continued personally without workers. `WaitForSubmissionAsync` now observes the child's authoritative current `submittedPayload` through an ingress-protected broadcast pulse. It captures the pulse while checking the state, so publication cannot race past subscription. Multiple waiters receive the same durable result; cancelling one observer does not cancel the child or another observer. Failed replacement publishes nothing; accepted new work invalidates prior completion; shutdown/fatal actor exit wakes unsubmitted waiters with an explicit disposed failure. A submitted payload remains available after joined shutdown, before destruction. This primitive is not a quiescence barrier or question-aware `join_agents`. Publication now refuses while any retained direct child lacks an explicit submission. Stopped-but-retained unsubmitted children still block; destruction removes that obligation. Submitted children may remain alive while their submit result/save drains. Tests extend the real loopback generation case and add multi-observer, cancelled-observer, failed-disk-publication, shutdown and bottom-up controls.

Questions use a direct relationship object, not a manager: `Common/Questions.cs` declares `AgentQuestion`, its strong asking/receiving Agent parents, the two Agent relationship lists and one Ensemble-owned graph lock. A standalone root is also its host's human inbox; its human question is a self-edge, not a fabricated human Agent. `Agent/Questions.cs` registers both ends atomically, assigns monotonic receiver-scoped IDs, and removes both links before completing a reply or cancelling it during either owner's shutdown. Replying without `messageID` requires exactly one matching sender question; ambiguity, wrong IDs and duplicate replies leave pending questions unchanged. The public value-like `QuestionInfo` retains no owner. APIs are `SendMessageToParentAsync`, `PendingQuestions`, `WaitForQuestionsAsync(humanOnly, cancellation)` and `RespondToQuestion(response, childID, messageID)`; null childID targets the root's human inbox. Cancelling an awaiting caller leaves the admitted question available for a later reply. The graph lock never enters Agent ingress; owned-child lookup precedes it, and the snapshot's Parent/ChildID accessors take no ingress locks. Shutdown closes actor admission, then drains both question directions, including the empty-inbox observation wake. Full source/declaration review completed in bounded reads.

Evidence in `/private/tmp/ch20-csharp.hH3ysE`: `wait-build.log` clean, `wait-tests.log` **6/0**, then `wait-shared.log` **167/0** and clean actual CLI/GUI builds (`wait-cli.log`, `wait-gui.log`) before questions were added. Final question source: `question-tests-build-3.log` clean; `question-tests-3.log` **3/0**; `question-shared.log` **170/0**; `question-cli.log` (`Cli/Ch04.csproj`) and `question-gui.log` (`Examples/Gui/Gui.csproj`) zero warnings/errors. Final fail-fast job 4045 exited 0. Question groups cover nested routing, root-human filtering, unique IDs and refusal invariants, broadcast observation, late replies after caller cancellation, shutdown from both owners and 30 concurrent admission/shutdown rounds. Deadline expiry is excluded from the successful-cancellation predicate. No new provider calls, GUI/browser campaign, live OAuth, audible or accessibility claim.

Test-setup corrections are retained: the initial `--question-tests` registration omitted its focused-return block and ran the three intended groups plus unrelated groups, **125/0** (`question-tests.log`, job 4040 exit 0), not a full suite. My first return-block edit used the nonexistent counter `passes`; `question-tests-build-2.log` records that compile failure, with no behavioral run. Corrected to the saved `passed` counter before the final focused/full runs. These were my runner mistakes, not product defects. No failed product assertion occurred in this slice.

No Chapter 20 executable grade or preserved endpoint exists; ch19 remains untouched. Question/submission APIs are not yet ordinary tools, and the root's human inbox is not yet wired to CLI input. Next: reliable question-aware joins that retain their admitted wait/work, idle-unsubmitted reminder/failure limits, completion/early-return and task-budget integration, actual spawn/get/send/ask/question/respond/join/shutdown/submit tools and CLI workflow, then scoped grading and assembled prose/source review. Do not infer those behaviors from the new primitives. Earlier GUI and creation evidence carries unchanged scope. No workers, commits, installations, restarts, shared-browser/8088 use or private credential access. Final narrow scratch-process check found none; free space was 13,107,704 KiB (a historical observation).

## 2026-10-28 — Chapter 20 direct child creation verified; workflow tools still unfinished

Continued personally without workers. `Agent.CreateChildAsync` now creates a ready child beneath the parent's configured persistence directory: `children/<id>/save.json`, child-owned `jobs`, and the existing durable `submit.md`. It snapshots the current immutable Configuration (optional same-provider model override), inherits policy through the existing direct constructor, installs ordinary tools, restores only that child's History, then completes Skills initialization before returning. It admits no task. Parent storage is required; invalid IDs/schema, duplicate siblings and closed admission refuse. Failure/cancellation destroys the incomplete child while preserving existing files; no separate manager, state-directory setting, model scheduler or inherited conversation was added.

Skills inheritance uses an internal value-like `SkillBootstrap`, captured before taking the parent Agent's admission gate to preserve lock ordering. It contains immutable parsed definitions, explicit variables, primary/startup selections and only the already-filtered MCP authentication values. It contains no owners, clients, active grants or callbacks. Each child gets its own Skills owner/dictionaries, activation and MCP clients. Common declares the construction operations; `Skills/Skills.cs` implements them and factors the previous constructor's attachment into one shared implementation. Deleting an owned fixture's original manuals proves no source-directory reread is required. Current bootstrap values replace saved authority on restore, while the child's own dialogue remains.

Added `--child-creation-tests` (also in full suite): recursive directory/ID ownership, configuration and safe policy inheritance, independent template/variable rendering, no copied parent dialogue, saved child restoration without stale completion, preflight refusal without child-directory effects, duplicate refusal, malformed-save rollback with unchanged bytes, same-ID recovery and stopped-parent refusal. An inherited-client test reuses the working HTTP MCP peer: independent authenticated parent/child startup and tool execution, child cleanup leaves parent client/grants usable, no test credential in History/logs, actual observed initialize request before cancellation, joined cleanup and successful same-ID retry. The peer's new hold-initialization switch is a test boundary behavior only, default off.

Retained setup failures: `create-tests-build.log` failed because I guessed nonexistent Skills `SetVariable`/`Variables` accessors; corrected to the actual `RegisterVariable` API and observable grandchild rendering. `create-tests-2.log` had two passes and one fixture failure because the malformed-save assertion omitted `InvalidDataException`; the product correctly refused the damaged save. Corrected `create-tests-build-3.log` is clean and `create-tests-3.log` has **3 passed / 0 failed** including startup cancellation. Final shared `create-shared.log` has **165 passed / 0 failed**. Both actual hosts (`Cli/Ch04.csproj`, `Examples/Gui/Gui.csproj`) build with **0 warnings / 0 errors**, recorded in `create-cli.log` and `create-gui.log`; fail-fast job 4014 exited 0. These final results cover the final product source. No repeated GUI/browser/speech campaign; those paths are unchanged. Scratch remains `/private/tmp/ch20-csharp.hH3ysE`.

Personally reviewed the complete creation operation, new Common declarations and Skills snapshot/attachment implementation. No Chapter 20 executable grade, complete child workflow or preserved endpoint is claimed. Next: reliable child lifecycle/question/join signaling, explicit submission constraints and ordinary spawn/message/question/join/shutdown/submit tools; actual CLI workflow, scoped grader, manuscript and final source review follow. Chapter 19's preserved endpoint is untouched. No live providers/OAuth, shared browser/8088, installation, restart, deployment or commit. Previous failure evidence and boundaries remain below.

## 2026-10-28 — Chapter 20 durable child submission verified; spawn/questions/joins unfinished

I connected the immutable submission contract to the actual Agent child, not a manager or tool-local cache. The child constructor accepts an optional schema and validates it before registration or directory effects. `Common/ChildSubmission.cs` declares the contract, published payload and admitted/running work generations; `Agent/ChildSubmission.cs` implements the operations. `DataDirectory()` derives from the authoritative normalized `PersistenceOptions.Path`, refuses without configured persistence, and does not copy a mutable directory or borrow Jobs' output directory. `SubmissionSchema()` and `Submission()` return independent JSON values. `PublishSubmission` validates the whole object once (or normalizes legacy text), refuses standalone/stopped Agents and older turns with newer queued work, writes a complete version1 JSON row to `submit.md` (child ID, generation, schema, typed payload), then publishes the in-memory completion. Invalid input or failed write/rename leaves earlier completion unchanged. This is a public durable-state API, **not yet a model-facing submit tool, full child state machine, spawn workflow, question exchange or join implementation**.

History's existing rooted temporary-write/flush/rename/cleanup algorithm moved to `Sandbox.ReplaceFile`; History now calls that same operation rather than duplicating storage logic. A same-directory private temporary is flushed to disk before handle-relative replacement; failed cleanup cannot mask the original failure. This retains existing persistence recovery behavior, not a new filesystem guarantee. Submission artifacts survive destruction; reconstructing the same child restores its own conversation through `StartPersistence`, but does not treat an old `submit.md` as completion for newly admitted work.

The actor tags accepted separate tasks with a work generation and invalidates the prior completion; accepted live hints also invalidate current completion without becoming a second scheduler. Ordinary final text never marks submission. The initial three groups passed in `durable-tests.log`; an initial shared run passed161/0. Review then added a queued-hint case. The first extended fixture incorrectly served only two requests: the actor correctly requested a third for the live-hint continuation, causing a fixture timeout (`durable-hint-baseline.log`), not a product diagnosis. Corrected fake protocol in `durable-hint-baseline-2.log` exposed the actual bug: advancing the generation on a hint while a task was queued orphaned that queued task's generation, so its valid submission refused. The fix retains the queued task's generation; the hint is already in its History. `durable-fixed-tests.log` passed3/0. `durable-final-shared.log` then passed **161/0**, and `durable-final-cli.log`/`durable-final-gui.log` built both actual hosts cleanly, collected exit0.

Final source review identified a separate lock-order risk, not a reproduced deadlock: `Sandbox.DeriveChild` held the Sandbox gate while reading Agent configuration, whereas existing persistence and new publication can enter Sandbox under Agent ingress. It now snapshots immutable configuration before taking the Sandbox gate. Reviewed Sandbox parent reads and teardown; no copied configuration/service state was added. A fourth positive-control group concurrently performs100 durable publications and1000 policy derivations, retains restrictions and final payload. Final `durable-lock-build.log` is clean; `durable-lock-tests.log` **4/0**, `durable-lock-policy.log` **1/0**, `durable-lock-ownership.log` **3/0**, all exit0. Final actual-host builds `durable-lock-cli.log` and `durable-lock-gui.log` have zero warnings/errors. The 161-group run precedes only this isolated lock-order adjustment and added concurrency test: it is not a claimed full162 run.

`Tests/ChildSubmissionTests.cs` covers invalid-schema no-registration, required persistence, exact typed row/schema/owner retention, >2^53 number preservation, string convenience no bypass, atomic replacement failure/no false publication/temporary cleanup, confined outward-link refusal, shutdown refusal, sibling schema isolation, saved-artifact retention, own-History restore without stale completion, real loopback task admission, queued-work/queued-hint generation behavior, ordinary-final-text noncompletion and the concurrency control. Selector `--child-submission-tests` is registered in the shared suite. Earlier GUI/browser/speech campaigns were not repeated. Setup errors retained: staged edits guessed constructor/mailbox anchors and refused before writing; test compilation initially omitted required Configuration endpoint/key arguments (`durable-test-build.log`), corrected before behavioral credit. All build/job exits collected; no workers or shared-environment changes.

**Next:** implement direct child creation under the parent's persistence directory, inherited parsed skill definitions/bootstrap, restore-before-admission and reliable child state notifications; questions/responses, submission lifecycle gating/wake limit, waits/joins/get-submission and actual CLI tools. Then scoped executable grading, complete C# prose/review and source-only endpoint. Child storage is currently manually configured by the API consumer; no automatic child path construction or spawn admission is claimed. Read actual `SkillSpec`, `Skills` retained templates, `PersistenceOptions` and actor declarations; older imported checkpoints had incorrect signatures. Chapter19 remains the untouched204-file endpoint. No Chapter20 grade/endpoint or live provider/OAuth/billing/native/accessibility claim. Scratch `/private/tmp/ch20-csharp.hH3ysE` retains the evidence.

## 2026-10-28 — Chapter 20 submission contract verified; durable workflow unfinished

I implemented the schema prerequisite before durable submission effects. `Common/Submission.cs` declares value-like immutable `SubmissionContract` and an internal exact-number value; `Agent/SubmissionContract.cs` implements them. Construction validates before retaining its independent schema copy. Null schema selects legacy `{result:string}` normalization with ignored extras; explicit schema retains the entire validated argument object without a result/message envelope or inserted defaults. The string convenience uses the same validator, so it cannot bypass a custom contract. **Not yet wired to child state, publication or tools:** this is not a completed submit/spawn workflow.

The validator covers the chapter's explicit type/object/array/string/number/enum/const subset, nested booleans, exact dialect marker and annotation shapes. Unknown/unsupported keywords and duplicate keys (including annotation/payload data) refuse. Admission enforces raw/serialized byte caps, root-zero depth64, node limits, number literal/exponent limits and the 67,108,864 schema-by-payload budget. BigInteger rational arithmetic gives exact integer/multipleOf/equality behavior beyond double. Shared type-tagged keys normalize numbers and object order for const/enum/uniqueItems. Review replaced the initially passing quadratic uniqueness scan with HashSet keys; no persistent cache or second owner exists. C# regex uses culture-invariant .NET syntax and 100 ms per-match timeout, not Go compatibility or a whole-validation deadline. The chapter table now states this; full C# prose reconciliation is pending.

`Tests/SubmissionContractTests.cs` adds five groups under `--submission-contract-tests`: structured/default behavior and retained copies, schema admission/refusals, exact numbers, nested constraints, and limits. Cases include grader-shaped nested objects; missing/extra/wrong fields; unsupported references without fetching; no default insertion; numeric spelling equality and precision beyond double; decimal multiples; Unicode code-point length; 12,000 unique values plus a late equivalent duplicate; type-distinct equality; real pathological regex refusal; raw/serialized bytes, node/depth, numeric and combined budgets. `submission-contract-*` passed initially, `submission-keys-*` passed after the uniqueness improvement; final `submission-final-build.log` has zero warnings/errors and `submission-final-tests.log` **5 passed, 0 failed**, exit0. Final job3944 collected. I read the complete saved implementation, rereading the numeric helper after truncation. No product failure was claimed. A documentation patch guessed a bullet instead of the actual table row and refused before writing; corrected exact-anchor edits saved the docs. These are setup failures, not validation failures.

Prior full153/0, both actual hosts and corrected self-isolated GuiTests apply to unchanged paths, not a new full158 result. No unchanged browser/speech/host campaign was repeated. No Chapter20 grade, completed workflow or preserved endpoint exists; ch19 remains untouched. Scratch `/private/tmp/ch20-csharp.hH3ysE`. Last free-space observation 12,551,128 KiB. No workers, provider/OAuth calls, private credential access, installation, restart, commit or shared cleanup.

Next: integrate the validated contract into Agent-owned durable child admission/restore, question/answer, submission generations and joins; then real CLI tools/exercise, scoped executable grader, prose and final review/preservation. Read exact saved APIs: current `SkillSpec` is directory/primary/variables/startup/MCP-environment, not the stale imported field list. CLI constructs Agent, installs Tools/Skills, starts persistence, then initializes skills. Root Agent has no data-directory API yet. History's rooted replacement helper is in `History/Persistence.cs`; guessed `History/Save.cs` does not exist. Determine the authoritative child-directory path through real ownership without a manager, second scheduler or copied services. Do not repeat the contract tests unless integration changes relevant inputs/behavior.

## 2026-10-28 — Chapter 20 direct child ownership verified; workflow still unfinished

I implemented the ownership foundation directly on Agent, with declarations in `Common/Children.cs` and functional bodies in `Agent/Children.cs`. A delegated constructor explicitly receives both Ensemble and Agent strong parents, registers both directions, and admits no task. Child IDs are parent-local, path-safe addresses, not entries in Ensemble's standalone name index. Children are listed/looked up on their actual Agent parent. Policy derivation reuses the Chapter 19 Sandbox contract and refuses widening before relationship/file effects; derivation occurs before taking the owner's admission gate to preserve lock order. Construction then holds that gate through registration/owned-component setup so parent shutdown cannot overlook an admitted child. Failed construction removes root/owner links. This is not yet `spawn_agent`: child data directories, restore/new work admission, private questions, submissions, joins and CLI tool wiring remain to implement.

`ShutdownAsync` closes admission and starts descendant shutdowns before waiting for any of them, sharing one joined task. It does not remove ownership. Destruction joins, recursively destroys children, then releases Jobs/MCP/Skills/tools/Engine/History/Sandbox and both parent links; saved files remain. I replaced nested exception-masking cleanup with explicit attempts that retain failures while still detaching every owner. There is no fleet/service object or second scheduler. I removed an unnecessary proposed 64-character ID limit before final verification.

Focused evidence in scratch `/private/tmp/ch20-csharp.hH3ysE`: `child-ownership-tests-build.log`/`child-ownership-tests.log` originally passed three groups; after cleanup review, `child-cleanup-build.log`/`child-cleanup-tests.log` passed the same three. They cover sibling-ID isolation, root/owner back-links, duplicate/wrong-root/path/policy/output-path refusals with unchanged relationships, targeted recursive destruction, reuse of a detached ID, stopped-parent admission, a real in-flight descendant HTTP request cancelled by parent shutdown, shared shutdown joins, and a poisoned child journal that cannot strand healthy siblings or erase their saved state. `child-shared-tests.log` is a clean **153 passed, 0 failed**, exit 0 for the assembled changed core. This is local component/fake-boundary evidence, not a child CLI workflow or Chapter 20 grade.

I made a concrete test-launch mistake after that regression: launched GuiTests from main although its fixture required an owned empty cwd. The consumer passed but overwrote `Skills/browser/SKILL.md` through lowercase `skills` on this filesystem and created four fresh output paths. I inspected the exact diff and birth times, preserved the erroneous manual and newly created `override`, `consumer-credentials`, `gui.log`, `mounted.log` under scratch `mislaunched-consumer`, and restored the shipped manual byte-for-byte from untouched ch19 (Chapter 20 had not changed it). I left the preexisting `jobs` directory alone. The mislaunched pass is not credited as isolated deployment evidence. `GuiTests/Program.cs` now creates its own unique temporary cwd and restores the caller cwd in cleanup; future invocations do not depend on caller discipline. Corrected `child-gui-isolated-build.log` and `child-gui-isolated-run.log` exit 0, workspace `scratch/tmp/ensemble-gui-consumer-4GsgfM`, including the existing replay/MCP/pending-RPC/credential-lifetime checks. The source browser manual still byte-matches ch19 afterward. Final actual host builds `child-cli-final-build.log` and `child-gui-final-build.log` each have zero warnings/errors. Earlier `child-gui-consumer-*` and host logs retain the mislaunch boundary rather than silently being relabeled.

All launched chains' exits were collected. No worker, shared restart, installation, commit, private credential use or live OAuth occurred. No repeat browser/speech campaign was needed; their inputs are unchanged. **Next:** complete durable child workflow on these direct relationships, real CLI wiring, scoped executable grader, teaching prose and final review/preservation. Chapter 19 remains the preserved endpoint. Existing `Tool.Validate` is a shallow built-in argument checker, not sufficient for recursive structured submission; inspect/reuse the actual boundary rather than assume full JSON Schema validation exists.

## 2026-10-28 — Chapter 20 whole-handler hosting verified locally; child workflow next

This supersedes the optional-GUI slice's asset-only limitation, not Chapter 19's preserved endpoint. I read the Chapter 20 reuse, fleet, submission, driver and grader-contract checks plus outside-consumer fixtures. Historical Go build/import checks are not C# acceptance; no Chapter 20 grade or endpoint exists yet. The child workflow, C# grader adapter, chapter prose and final assembled review remain unfinished.

`Gui.ServeAsync(WebApplication, Agent, GuiOptions, CancellationToken)` now mounts the complete interactive handler in a caller-owned application. Its task owns only operation-scoped observation, sockets, request admission/draining and trace lifetime; Common declares the public operation and value-like `BasePath` option. It borrows the host and Agent. `RunAsync` delegates to that implementation, owns its private loopback host and notices serving-task failures rather than waiting forever for host shutdown. Caller cancellation closes admission, aborts sockets, joins admitted requests/browser teardown and observation before trace disposal. Stopped routes return 503; unrelated caller routes remain live. Prefix validation accepts literal slash-separated names, not route templates. HTML assets, account fetches and WebSocket URLs are page-relative; the bare mount redirects to a directory URL. Existing projection/control algorithms remain the shared implementation, not a second transcript or scheduler.

Source review caught a real integration incompatibility before acceptance: Credentials intentionally requires the fixed `/auth/callback`, so my initial prefixed callback construction would refuse login. I preserved the credential validator and reserved only `/auth/callback` at the application root for prefixed mounts; other account operations remain prefixed. This reservation, one GUI per Agent/application, loopback binding/exact port, and the lack of reverse-proxy/UsePathBase support are explicit in the public docs. Borrowed hosting leaves shared pending credential attempts caller-owned; standalone hosting retains its previous cleanup. No new authentication escape hatch or credential cache was added.

Final public outside-assembly consumer evidence: scratch `/private/tmp/ch20-csharp.hH3ysE/mount-callback-build.log` clean zero warnings/errors; `mount-callback-run.log` exit 0. `GuiTests` runs without deployed wwwroot and retains the earlier embedded/override/no-fallback, two standalone lifetimes, foreign Host and Agent-survival tests. Its full `/assistant` mount now verifies canonical replay, prefixed controls, account guards/root separation, real browser MCP initialize/list/call via a granting primary skill and ordinary Agent dispatch, plus a pending RPC terminated and both Skills/Mcp client memberships detached on cancellation. Independent `/health` survives, stopped GUI requests return 503, and the Agent remains usable. A new owned credential store prepares a PKCE URL locally, checks the fixed redirect and wrong-state callback refusal, then proves the pending attempt survives borrowed GUI shutdown and cancels it explicitly. No returned authorization URL is opened, no issuer HTTP/token exchange occurs, and no model request is made.

Actual frontend exercise: `/private/tmp/ch20-mount-browser.AFNeIJ/result.json` and `exit.json`, launched by `mount-browser.mjs`, show the isolated muted Chrome page redirect, canonical replay, pause/resume, account status, prefixed asset/account/WebSocket paths, zero page-script errors and clean owned host exit. It used neither Bill's browser nor 8088, sent no prompt and navigated no OAuth URL. `mount-final-gui-build.log` is the clean actual GUI build after the callback correction. Actual CLI `mount-cli-build.log` was clean before that GUI-only correction; its core/source inputs did not change, so no redundant CLI rebuild is claimed. `mount-final-js-check.log` passes app/auth syntax; `mount-speech-tests-2.log` passes the existing formatter/queue fixture, not audible speech. No unchanged full suite or Chapter 19 campaign was rerun.

Retained failures are setup/fixture failures unless stated otherwise. First mount compile required casting RouteGroupBuilder to its explicit IEndpointConventionBuilder. Test edits lost two escaped C# strings, then used an unsupported Configuration parameter; both failed builds precede corrected tests. Several exact-anchor scripts refused before writes because I guessed helper names/context/indentation; later edits use actual declarations. The first RPC fixture waited for top-level part_final instead of replay.data.events and timed out despite a recorded successful replay/initialize frame; corrected protocol assertions pass, with no replay product fix. The first scratch browser host omitted the required diagnostics writer, so the launch found no DLL; only the fail-fast rebuilt later run is evidence. An earlier speech command used the wrong root path; the actual file is Tests/SpeechTests.mjs. These repeated avoidable fixture mistakes remain mine, not product defects. The callback-path incompatibility was identified from actual source, not credited as an executed failing baseline.

Reviewed the saved new hosting/cleanup/declaration/auth/asset/frontend deltas and corresponding consumer. Docs now describe actual ownership and callback limits. A final saved-file check found earlier README/comment replacements had silently matched nothing; I replaced the bounded text explicitly before checkpointing. That was a documentation-edit failure, not new behavioral evidence. Latest disk observation 12,633,200 KiB free. No workers, installs, commits, shared restarts, private credential searches or live OAuth; live OAuth remains deferred. Next: design direct Agent parent/child ownership and implement the chapter's CLI child workflow with durable schema-validated submission, private question/reply, wait/join and joined shutdown. Retain actual parent paths, policy narrowing and single authority; no redundant fleet service.

## 2026-10-28 — Chapter 20 started: optional GUI library slice (not chapter completion)

Chapter 19 remains preserved and locally complete; `ch19` is untouched. Main is now Chapter 20 work in progress. I read the current Chapter 20 manuscript through its rubric and compared its preserved revision at `/Users/bill/projects/cadence-backups/edition-3-manuscript-before-restart-20261025T042452/chapter-20.md`; the expected docs preserved path does not exist. The historical grader is `internal/gradeedition2/ch20*.go`, with reuse/fleet/submission consumers. Initial inventory and portions were read, but full grader-contract review is still unfinished. No Chapter 20 score, child orchestration or endpoint is claimed.

I extracted the existing GUI Common/Server partials and static assets into optional `Gui/Ensemble.Gui.csproj`; the core's explicit source list and independence are unchanged. The example now references the library and retains CLI/environment/user-home selection. Public value-like `GuiOptions` supplies port, explicit assets override and rooted settings/trace paths. `Gui.RunAsync` borrows its Agent, accepts cancellation and joins the existing scoped host/observer/socket cleanup without destroying the Agent. Embedded resources remove the deployment-side wwwroot requirement. `Gui.MapAssets` exposes fixed public asset names on caller-owned routes, with explicit trusted disk override and no fallback for missing override files. This is asset-only mounting: complete interactive-handler mounting and subtree-relative browser routes remain unfinished. Existing projection, RPC, skill startup, auth and socket algorithms were moved, not redesigned.

Verification under owned `/private/tmp/ch20-csharp.hH3ysE`, using the isolated env-i private SDK wrapper: GUI/library build `extract-gui-build-2.log` and actual CLI `extract-cli-build-2.log` clean, zero warnings/errors. The outside-assembly public consumer is retained as `GuiTests`; published `gui-consumer-build.log` is clean and `gui-consumer-run.log` exits 0. It runs from an owned unrelated directory without a deployed wwwroot, checks all embedded assets, explicit override/no fallback, custom asset routing, foreign Host refusal, real WebSocket pause, browser-relay registration and removal, zero discovery without a granting skill, two hosting lifetimes on one Agent and continued Agent usability. No provider request is submitted. `speech-tests.log` exits 0 using the moved speech asset; formatter/queue test is not audible evidence. No unchanged core/full regression or earlier executable campaigns were repeated. Project-file formatting after this build changed whitespace only.

Failures and corrections are retained. The initial extraction anchor incorrectly assumed a core exclusion glob; it refused before writing, then I used the actual explicit source list unchanged. The first GUI compile failed because optional defaults were repeated on partial implementations; corrected bodies compile. CLI invocations initially used nonexistent project paths; only `Cli/Ch04.csproj` is the actual successful build. The first outside-consumer run expected a graceful WebSocket close handshake whereas existing socket teardown aborts; that extra test assumption was removed. I then incorrectly diagnosed a legacy MCP-client cleanup path and waited for discovery with no granting skill; the saved current host does not have that legacy branch. The fixture timed out, not the product. Rereading the saved host confirmed browser registration alone grants no discovery; corrected consumer tests that contract and clean teardown. No speculative MCP product fix was made. Assertions about GuiStub/test linked source were likewise wrong; their saved projects do not link these files and were left unchanged. Combined reads repeatedly truncated; source claims here are confined to the saved extraction diff and bounded inspected paths, not invented complete grader review.

Next: finish exact Chapter 20 grader/source-contract reading, implement caller-owned whole-handler hosting with the simplest ownership-consistent API, then the required direct child-Agent CLI workflow and local protocol exercises. Preserve existing grants/policy narrowing, durable submission and question/join semantics. Prose, suitable scoped executable grade, final personal review and source-only endpoint are all pending. No live OAuth, private credential search, installs, shared restart, commits, workers, Bill browser or port 8088. Latest disk observation 12,691,004 KiB free. All slice jobs collected; no children.

## Chapter 19 executable verification — final assembled review and preservation remain

## Chapter 19 — COMPLETE within local scope; final review and source-only endpoint preserved

I completed the remaining assembled delta review personally: Engine media/trace/Responses, History persistence, Jobs/PTY/automatic work, Skills connections, CLI/GUI state and diagnostic writers, CacheLens, and their Common declarations. Earlier reviewed ownership/policy/native-file sections remain applicable. No further product-code changes were needed after the final executable-verification milestone below. The diagnostic-lifetime regression reads the authoritative key after detach while public Configuration still refuses disposed access; no copied credential cache was introduced.

I read the complete assembled manuscript and corrected stale Go references, an overbroad credential diagram/History-filtering claim, and the exact synchronization explanation. The historical deny-everything experiment and measured Go sessions remain explicitly attributed. The reproducible exercise now uses the real C# project and lowercase `ch04.dll`, current boolean flags, a pre-created narrower root, a synthetic marker, and Git staging/readback without a commit. Historical commit outcomes remain history, not current instructions. All four shell blocks passed `/bin/sh -n`; that is syntax checking, not a new model-led exercise. An initially written uppercase DLL spelling was corrected against the actual AssemblyName before preservation. README now describes the completed local Chapter 19 contract; one stale Chapter 18 unfinished sentence was also corrected.

**Evidence carried unchanged, not rerun for timestamps:** `/private/tmp/ch19-csharp.6z2ecI/review-shared-final.log` ends **150 passed; 0 failed**. `review-final-cli-build.log` and `review-final-gui-build.log` build both actual hosts with zero warnings/errors. `grade-final.json` is **90/90, passed:true**, collected exit 0 with empty `grade-final.stderr`; six scoped checks, two clean hosts and 29 local requests. Go-only child-consumer points remain excluded. The seven CLI launches at `/private/tmp/ch19-root-cli.ev00yt82/result.json` and five GUI launches/two requests at `/private/tmp/ch19-root-gui.s1Vtye/result.json` remain evidence for unchanged host behavior. Earlier product failures and fixture/setup errors remain below. Some combined review output truncated; I reread the omitted relevant test sections separately, rather than crediting hidden output.

**Preservation:** `solutions/edition-3/ch19` was created only after confirming it did not exist. It contains **204 source/support files** (169 C# files, 12 projects, 11 Markdown files and 12 other source/support files), all byte-identical to reviewed main. Relative project references resolve inside the endpoint. No symlinks, bin/obj/cr directories, runtime output or private artifacts were copied. Added-line whitespace checks over the 55-file source/support delta against untouched `ch18`, and scoped manuscript/grader/progress whitespace checks, passed. The source changes since the verified executable milestone are documentation only. Last observed free space: 12,915,264 KiB, not a future guarantee.

**Limits and next work:** tested native/process scope is macOS ARM64. Linux/x64, live providers/OAuth, billing, hosted MCP service, audible speech and human accessibility are not credited. Stdio MCP is separately trusted; safe mode, named-web gating and command-network policy are distinct. Known-value diagnostic masking is not canonical-content sanitization or arbitrary-secret detection. No actual Git commit, installation, deployment, shared restart, private-credential search or use of Bill's browser/8088 occurred. Continue with Chapter 20 requirements/source/grader reconciliation, without workers or repeated Chapter 19 campaigns. The expected `docs/edition-3-preserved/chapter-20.md` was absent; locate actual comparison sources before assuming a preserved manuscript exists.


**Boundary:** Chapter 19 now has a C# executable grading route and corrected local implementation, but no preserved `ch19` endpoint yet. The current and preserved manuscripts were read completely and compared. Current `book/edition-3/chapter-19.md` has been rewritten around concrete C# APIs/ownership, host flags, rooted handle operations, shell profile, exact diagnostic masking, child-policy derivation, and the 90-point executable scope. Bill's rationale and historical trial account remain attributed, not relabeled as C# observations. The assembled prose and remaining source deltas still need their final read; do not declare chapter completion from the grade alone.

**Executable scope:** `grade-edition2 -ch 19 -executable-only CLI` retains six existing host checks and their actual weights: path 15, kernel 20, network 15, credentials/diagnostics 15, safe mode 15, refusal-path leakage 10. The Go-only child consumer (10) is excluded, not silently awarded from component evidence. C# `DeriveChild` remains separately tested. The old Go wrappers/routes remain intact. The adapter reuses the existing fake, correlated tool-result driver, async job observation and managed shutdown. Every fake tool response has an ordinary explanation. The Git positive control uses init/add/staged-blob and identity inspection, **no commit**. Scope validation refuses GUI/Go consumer flags before binary lookup. C# sessions retain actual requests, stderr and completion result in their owned fixture directory.

**Real defects found and corrected:** the first executable run stalled in recursive search on a .NET diagnostic FIFO inside rooted TMPDIR. Native blocking open held Sandbox's gate. Confined acquisition now uses `O_NONBLOCK` and refuses non-seekable handles before returning a stream. Recursive search skips inaccessible/outward/vanished discovered entries without losing legitimate matches; explicit outside requests still refuse. `review-fifo-baseline-tests.log` has two passing/two failing groups (outward recursive traversal and blocked FIFO). The FIFO fixture rescues its own baseline reader so failure cannot strand cleanup. Final `review-fifo-tests.log` passes all four; `review-fifo-tools.log` passes compatibility. No claim of arbitrary device/mount/hostile-host isolation is added.

The next executable run completed at **80/90** (`grade-corrected.json`): the refused relative cwd's error itself was sanitized, but the enclosing Job report exposed the absolute output path. Confined model-facing text now spells that path relative to the root; trusted `JobInfo.OutputPath` stays absolute for actual host readback. `review-path-baseline-tests.log` retains the failure; `review-path-tests.log` has three passing process/PTY groups. No duplicate path state or changed file ownership was introduced.

The first broad regression (`review-shared-tests.log`) passed 149 groups and failed the concurrent diagnostic test. The root's Agent snapshot outlived an Agent's public Configuration lifetime. Swallowing that exception would risk unmasked logging. Common/Agent now declare/implement one internal `DiagnosticKey()` read of the authoritative immutable Configuration under the existing ingress gate, valid through teardown. Root diagnostic snapshots use it; public Configuration still throws after destruction. The test covers real concurrent owners plus an actual destroyed Agent retained in a snapshot. No persistent key cache or new owner. `review-diagnostic-lifetime-tests.log` passes all three diagnostic groups.

**Final local evidence:** `review-shared-final.log`: **150 passed, 0 failed**, exit 0. `review-final-cli-build.log` and `review-final-gui-build.log` are the actual CLI and `Examples/Gui/Gui.csproj`, both zero warnings/errors. Final pinned artifacts are `grade-hosts-final`, launcher `cli-grade-final`. `grade-final.json` is **90/90, passed:true**, exit 0, all six checks. Exact owned sessions `tmp/sandbox-grade-1377395290` and `tmp/sandbox-grade-2282717966` record 22 and 7 requests (29 total), `error:"<nil>"` after managed clean shutdown; each `provider-requests.json` length agrees. This is real local CLI/kernel/files/socket behavior against synthetic provider credentials, not a live-provider or OAuth result. `grade-scope-refusals.json` retains six pre-launch refusals. `grader-contract-tests-4.log` passes `^TestSandbox`; final grader build is `grader-build-4.log`. Main production sources match the artifacts tested at this point; subsequent README/manuscript edits are prose only.

**Retained setup/harness mistakes:** staged patches refused guessed safe-loop/driver anchors before writing; the fixture directory member was `base`, not `work`. I initially compacted the check slice despite hard-coded historical indexes, and mislabeled the actual network/path-leak checks as legitimate-work/web checks; source table reuse corrected both before final scoring. I invoked a nonexistent `-bin` flag, a wrong grader binary name, and a too-broad test regex that selected the old Go Chapter19 integration test without its Go endpoint. A leftover Go-only log-limit assertion was gated consistently. The new FIFO runner initially used another harness's helper names; corrected to `AsyncCase`/`failures` before behavior ran. Receipt reading guessed `requests.json` rather than `provider-requests.json`. Scope-refusal smoke checks repeatedly overfit wording before checking the real exit/before-binary-lookup contract. These avoidable assumptions are mine; their failed logs remain under the Chapter19 scratch. None is credited as product verification.

**Review completed so far:** grader scope/driver/check adaptation, current/preserved chapter requirements; Common Sandbox/Config/owner declarations; Agent construction/shutdown/diagnostic lifetime, policy/grant narrowing; complete Sandbox Files/Directories/Children/Command; current file-search correction and process report correction. The source-only delta from ch18 is 55 source/support files. Final assembled review still needs the remaining Engine/media/trace, History/Jobs/PTY/automatic, Skills connection, CLI/GUI/cache deltas and rewritten prose. Earlier slice reviews and host exercises remain applicable; do not rerun them for timestamps. No endpoint copy yet. Native Linux/x64 unverified; stdio MCP and trusted startup inputs are outside the shell profile. Exact-value diagnostic masking is not canonical-content rewriting or arbitrary-secret detection. Live OAuth remains deferred; no provider/account/billing, audible or human-accessibility claims, installation/restart/commit/push authority, or workers.

## Chapter 19 host root integration — prose, executable grading, and final review next

**Boundary:** both actual hosts now accept `--sandbox ROOT` over `EN_SANDBOX_ROOT`. Empty means full-access and can explicitly override the environment; invalid/missing roots refuse rather than fall back. No process-wide cwd change. CLI Jobs use a relative `cr/io`; confined GUI settings default to rooted `settings.json`, while full-access keeps its established user-home default. State, archive, journal, API/debug/GUI protocol logs and CacheLens capture writes now reach the retained Sandbox. Explicit outside state/trace/settings/capture paths refuse. GUI settings retain save-before-publication and rooted atomic replacement; a swap/refusal leaves settings and budget unchanged. Trusted skills/static assets and the explicitly selected credential store are composition inputs, not model file tools; keep credentials outside the readable root. Stdio MCP remains separately trusted, not wrapped by the run_command profile.

**Ownership/design:** Ensemble declares `AttachDiagnostics` for startup after an Agent can open rooted files. Writers remain caller-owned and close after application destruction. Ordinary diagnostics, API traces, and attachment use the same root gate. `Log` accepts an optional caller-owned destination so GUI protocol tracing uses the one root diagnostic scrubber rather than its old duplicate replacement logic. Engine-owned CacheLens uses Engine → Agent → Sandbox for directory validation and every capture open; its existing helper is now declared as partial in Common. Failed construction never attaches a lens; later capture failures remain best-effort diagnostics. No new service owner, secret cache, current-directory mutation, or canonical-content rewriting.

**Verified local behavior:**
- `host-cache-baseline-tests.log`: 3 compatibility groups passed, new confined-capture group failed because an outside capture directory was accepted. Corrected `host-cache-tests-final.log`: 4 passed, including rooted capture, constructor nonattachment on refusal, and symlink substitution without outside writes. The first post-fix test expected UnauthorizedAccessException instead of Sandbox's established IOException; fixture corrected, receipt retained as `host-cache-tests.log`.
- Final `host-review-build.log`, `host-review-cli-build.log`, and `host-review-gui-build.log` have zero warnings/errors. The actual web host is **Examples/Gui/Gui.csproj**. Earlier `diagnostic-label-final-gui-build.log` demonstrably built **GuiStub**, a library construction consumer; its prior “both hosts” wording must not be read as an actual GUI-host build. This slice builds and exercises the actual GUI.
- `host-review-diagnostics.log`: all 3 readable diagnostic/real signed-local bearer/streaming compatibility groups passed after the root logging API change. Earlier Responses8/recovery2/child1 scopes remain unchanged; no new full-suite claim.
- `host-root-cli-exercise.log`, `/private/tmp/ch19-root-cli.ev00yt82/result.json`: 7 fresh process launches, 4 success exits and 3 expected startup refusals. Environment root, explicit root override, explicit empty override, safe-mode shell refusal, outside save, outside trace, and invalid root. Real file/shell dispatch verifies in-root work and outside denial, with API trace/save/Jobs under the effective root. Full-access override deliberately accesses only the owned outside sentinel. No inference requests.
- Final `host-root-gui-final.log`, `/private/tmp/ch19-root-gui.s1Vtye/result.json`: 5 fresh launches, 3 clean normal exits and 2 expected refusal exits, **2 local model requests**. Isolated Chrome/profile and owned loopback ports (never 8088) verified visible final text, rooted write/shell work, outside read refusal, diagnostic secrecy via read_file, all API/debug/GUI/capture/save/journal files under root, flag precedence, empty override, settings save and restart, no publication/outside modification after settings symlink substitution, and outside settings/GUI-trace startup refusal. No page-script errors. Browser was muted; no audible/accessibility/live-provider claim.

**Failures and review correction retained:** wrapper invocation initially omitted `dotnet`; all-anchor patch drafts refused several guessed parser/writer/capture shapes before writing. Corrected against actual bounded source. The first GUI fixture expected `settings`, but the saved frame is `current_settings`; `host-root-gui-exercise.log` retains that timeout, no product change. The next exercise passed but omitted `gui.log` from its artifact checks. Personal source review found its separate unrooted writer. `host-gui-trace-baseline.log` proves the actual launch-directory file and absent rooted file; this was a real missed integration, not a fixture-only failure. Rooted that writer, removed its duplicate scrubber through the shared Log operation, strengthened the file assertions, and reran fresh to the final result above. Also corrected GUI's empty-root override rejection to match CLI/library. Earlier `host-root-gui-exercise-2.log` is intermediate and incomplete, not final evidence.

**Remaining:** reconcile current/preserved Chapter 19 prose and C# exercise/grader contracts, perform final source review, run scoped executable grading, preserve source-only ch19 endpoint, then chapter-boundary handoff. Chapter 19 is not complete. ch18 untouched. No workers, installation, deployment/restart of shared applications, commits, live OAuth, or private credential access. Scratch `/private/tmp/ch19-csharp.6z2ecI`; all exits collected. Latest observed free space 13,114,412 KiB.

## Chapter 19 diagnostic secrecy and child-policy derivation — host integration next

**Boundary:** Ensemble now masks whole known configured API-key values before both ordinary diagnostics and API trace writes. Keys are reached through live Agent relationships, not retained in a secret cache. Exact values plus ordinary and HTML-safe JSON spellings are covered; longest-first ordinal searches operate only on original input, not replacement markers. Trace labels/directions are scrubbed as well as payloads; actual Agent identity is unchanged. Agent list snapshots/add/remove are synchronized so concurrent registration/destruction cannot invalidate a logging enumeration. Engine passes invocation-local used credentials through its private parser/trace path, retaining the rejected and rotated plan bearers only for the one exchange. Common declares the changed operations. No new persistent Engine field, credential store, History field or public parsing API. HTTP non-success bodies remain omitted and parser errors remain status/type-only. This is a known-value diagnostic backstop, **not** arbitrary-secret/encoding detection, cross-frame content filtering, or canonical successful-content rewriting. The existing streamed-content preservation contract remains tested. Credentials belong outside the model-readable root; the bearer test explicitly does so.

**Child policy:** public `Sandbox.DeriveChild(Configuration? requested = null)` returns `(Configuration Effective, string? Refusal)`. Omission inherits this owner's effective physical root and capability restrictions. Safe mode combines by OR; web permission by AND. Explicit roots resolve through the parent Sandbox, including future descendants and inward links; rejected outside/invalid roots fall back to the inherited root while reporting a sanitized refusal. Even a caller ignoring the refusal receives narrowed flags/root. Derivation creates no directories and no child fleet/ownership relationship. A Common-declared internal Configuration copy constructor preserves every other validated immutable setting, including plan billing and explicitly disabled summary, without reapplying model defaults. Tests construct an ordinary Agent from a refused result to verify actual file/tool enforcement. This is a public policy-consumer boundary, not a claim about a not-yet-implemented runtime child-spawning fleet.

**Evidence:** all files below are in `/private/tmp/ch19-csharp.6z2ecI`. `diagnostic-baseline-tests.log` exposed unsanitized ordinary diagnostics while existing streaming trace compatibility passed; the corrected logger passed two groups. `child-policy-tests.log` and `diagnostic-review-child-policy.log` each passed the child group: inherited/equal/sub/future roots, inward/outward links, absolute/traversal/prefix-sibling widening, flag widening, real child tool dispatch and unchanged parent/plan settings. `bearer-trace-baseline-tests.log` **2 pass / 1 fail** exposed the missing plan trace mask. Signed local issuer/HTTP401/refresh/storage plus read_file then passed in `bearer-trace-fixed-tests.log` **3/3**. No live credentials or provider used. `diagnostic-review-diagnostic.log` **3/3**, `diagnostic-review-responses.log` **8/8**, `diagnostic-review-auth-inference.log` **2/2**, and both review hosts built cleanly. The last two groups retain bounded recovery/error/no-fallback and delayed-rejection coverage; they do not rerun the complete Auth suite.

**Final review/corrections:** replaced the initial regex-based one-pass matcher with direct ordinal searches to avoid an unnecessary pattern engine/size limit (review simplification, not a reproduced failure). Added concurrent Agent creation/logging/destruction. A standard JSON-escape fixture then exposed a real missed spelling: `diagnostic-escape-baseline-tests.log` **2 pass / 1 fail**; `diagnostic-final-tests.log` **3/3** after adding that explicit spelling. A credential-bearing diagnostic Agent label/direction exposed another real trace field gap: `diagnostic-label-baseline-tests.log` **2 pass / 1 fail**; final **`diagnostic-label-final-tests.log` 3 pass / 0 fail**, clean `diagnostic-label-final-build.log`, CLI and GUI builds (zero warnings/errors), job3652 exit0. That command checked the expected baseline assertion before applying the correction; its initial FAIL line is the retained baseline, not the final test result. No new full-suite result is claimed; shared145 predates this milestone and the earlier atomic Git publication. Responses8/recovery2/child1 remain applicable unchanged paths after the final diagnostic-only refinements.

**Editing corrections/limits:** several guessed source filenames/signatures and the initial test dispatch receiver were wrong. The all-anchor bearer patch refused before writes on the actual Responses callback/helper and send-header structure; I reread the full implementation and applied the narrower actual call path. These were editing/setup errors, not product-test failures. Final source review includes child derivation/copy, root diagnostic/relationship operations, invocation-local credential flow, parser trace calls, and status/type-only error construction. Prior chapter endpoints remain untouched; no host sandbox flag, ch19 endpoint, install, shared restart, commit, private credential search or live OAuth occurred. Free-space observation: 13,210,180 KiB. Next: expose the root at CLI/GUI composition roots with safe startup paths, run owned host/GUI exercises, then reconcile prose/grader and preserve the completed local endpoint. Continue personally; do not rerun unchanged campaigns.

## Chapter 19 macOS command boundary integrated — diagnostics and child policy remain

**Boundary:** Job now reaches `Agent → Sandbox.PrepareCommand` for launch cwd, environment and an in-memory Seatbelt profile. Common declares the preparation, bootstrap/pump/stop operations and native `confstr`; no new persistent service or policy copy. The shell and descendants retain PTY semantics while writes stay rooted, system/runtime/toolchain reads are bounded, and child networking is denied. Confined environment uses only selected locale/timezone values plus fixed PATH/TERM, rooted HOME/TMPDIR and generated Git identity. No user Git configuration is read. Immediate physical Darwin temp `xcrun_db*` files have the narrow cache exception; shared temp and nested lookalikes do not. `/Library/Developer` is readable, not all `/Library`. Root/ancestor metadata permits traversal without sibling contents. Unsupported command platforms still refuse; only macOS ARM64 was exercised. Stdio MCP's separately trusted launch remains outside this shell-specific chapter boundary (current chapter lines 210–215); it already has its own environment allowlist. No new network claim for arbitrary additional tools.

**Control/lifetime:** removed writable `.pgid` files. The fixed bootstrap publishes one bounded ASCII process-group line before user code; Job consumes it once, then pumps ordinary output. Public stdin is unavailable until that group is established. Callback disposal is joined before process release; process references clear even if output recording fails. Existing exit-code recording/CRLF normalization/cancellation behavior remains. Explicit cwd metadata now uses Sandbox resolution; omitted cwd retains its existing metadata convention, while real PWD is tested at the root. Review also made the generated Git identity a rooted create-and-rename publication rather than concurrent truncating writes. That is a review correction with a concurrent positive control, not a claimed reproduced race failure.

**Evidence and corrections:** `process-baseline-build.log` clean; `process-baseline-tests.log` **1 pass / 2 fail** showed outside command reads and inherited unallowlisted variables, while existing PTY compatibility passed. `process-rooted-tests.log` **2 pass / 1 fail**: confinement and compatibility worked, but Python's xcrun shim could not traverse `/Library` to its allowed developer path. Added metadata-only system-ancestor permissions, not broad content access; `process-traversal-tests.log` **3 pass / 0 fail**. Expanded `process-reviewed-tests.log` **3 pass / 0 fail** covers owned immediate/nested/sibling cache files, actual denied loopback connection, quoted roots, synthetic shell startup hooks, environment/PWD, real Python writes, Git init/add/config/status, stdin and killed descendants. No project or fixture commits, no installs, no real credentials or external network calls. The tests use actual dispatch, PTYs, files and a newly owned listener, not mocks. Drafting errors included wrong process filenames/signature, wrong fixture CallID/OutputDirectory names, a declaration anchor that saved only Sandbox, and an unnecessary duplicate Signal declaration. I reread actual source, removed the duplicate and replaced the incorrect draft before compiling. Those setup mistakes were not product evidence.

**Final scope:** `process-final-tests.log` **3 pass**, `process-full-tests.log` **145 pass / 0 fail**, clean hosts, command 3583 exit 0. That shared run includes persistence but **predates** final atomic Git-file publication. Final `process-atomic-build.log`, `process-atomic-tests.log` **3 pass / 0 fail** (including four concurrent launches), `process-atomic-cli-build.log` and `process-atomic-gui-build.log` are clean, zero warnings/errors, command 3587 exit 0. No duplicate full regression was run for the small final correction. Saved process source and whitespace reviewed; no pending jobs/children. Scratch `/private/tmp/ch19-csharp.6z2ecI`; latest observed free space 13,362,996 KiB. README still warns against whole-Agent security claims; no host root flag, executable ch19 grade, preserved endpoint, installation/shared restart/deploy/commit/private-credential access/8088 use. ch18 untouched.

**Next:** credential-safe content/diagnostics and child-policy narrowing, then host controls/GUI exercise, C# prose, executable grader and source-only endpoint. Do not repeat unchanged earlier campaigns or live OAuth.

## Chapter 19 persistence integrated — process boundary still open

**Boundary:** snapshots, backups, journals, archives and diagnostic file paths now use History → Agent → Sandbox. Common declares the owner-aware instance helpers; Agent startup normalization reaches its History. Same-directory temporary replacement still flushes before rename; backup protection, journal ordering, recovery validation and publication are unchanged. Failed temporary cleanup stays rooted and cannot mask the original write/rename error. No new root/service state. This confines diagnostic file access, **not its credential-bearing content**; redaction remains unfinished. No host root flag, ch19 endpoint or whole-Agent security claim. Personal/no workers; ch18 untouched. Scratch remains `/private/tmp/ch19-csharp.6z2ecI`.

**Evidence:** new `SandboxPersistenceTests` exercises actual History/Agent APIs for relative/inward-link state, primary/backup/journal recovery, archive and diagnostic replay, lifecycle save, absolute/parent/outward-link refusal, alias protection, outside backup/journal links, swapped journal directory, failed-append nonpublication/poisoning, and destruction despite that poison. `persistence-baseline-build.log` clean; `persistence-baseline-tests.log` **7 pass / 2 fail**, demonstrating accepted outside log/save writes. After routing, `persistence-rooted-tests.log` **8 pass / 1 fail**: the new fixture did not expect the deliberate poisoned History shutdown error. Production cleanup was correct; the fixture now expects the error and asserts Agent detach. Final `persistence-reviewed-build.log` clean and `persistence-reviewed-tests.log` **9 pass / 0 fail, exit 0**, including all seven existing persistence/journal groups. CLI build clean. My guessed `Gui/Gui.csproj` path failed before build; actual `GuiStub/GuiStub.csproj` was located and `persistence-gui-corrected-build.log` is clean, exit 0. All job exits collected. Full 141 from the previous milestone is carried as earlier evidence, not rerun/relabelled for this change.

**Review and limits:** bounded direct comparison against ch18 reviewed the production changes; Git supplied no tracked diff for these main files and is not review evidence. No remaining direct File/Directory filesystem calls in History. Native Linux/x64 and process-level confinement remain unverified. Last free-space observation: 12,264,516 KiB. No installation/shared restart/deployment/commit/private-credential access/8088 use.

**Next:** PTY control files, child environment and OS process confinement; credential-safe diagnostics; child-policy narrowing; host exercises, teaching prose, executable grader and endpoint preservation. Continue without repeating unchanged campaigns.

## Chapter 19 media and Jobs output integrated — persistence/process still open

**Boundary:** main routes media paths and Jobs output through the Agent-owned Sandbox. Not whole-Agent confinement, a host root flag, a ch19 endpoint, or chapter completion. Personal/no workers; ch18 untouched. Scratch: `/private/tmp/ch19-csharp.6z2ecI`.

- Common declares Engine's `ReadMedia`; its implementation uses Engine → Agent → Sandbox for path bytes. The three provider blob helpers and necessary render methods are instance methods. Inline decoding stays on value-like BlobRef; its public standalone `Data()` is documented as unconfined, never used by Engine for paths. Capability checks precede resolution; Responses remains unsupported. Tests cover all three renderers, absolute/relative/outward-link refusal, inward links/bytes, inline/ephemeral parts, nested supported Anthropic tool-result media, and actual Agent rejection with no RequestSent/connection and unchanged canonical locator.
- Jobs directory creation, initial output open, observation readback and automatic error completion now use Jobs → Agent → Sandbox. Full access keeps its temp default; confined default is beneath root; explicit outside directory refuses. Retained files survive destruction. Tests cover default/relative paths, outside setup, replaced output leaf and parent symlinks, no outside changes/orphan Job, owner detach.
- New setup refusal exposed a real partial-Agent cleanup bug: starting the lazy actor during failed construction required missing Jobs and masked the original error. Constructor failure now closes admission if no actor exists; shutdown returns completed work in that case, without starting over incomplete owners. Ordinary shutdown remains unchanged.
- Separate fixture correction: `write_file` legitimately creates a missing parent in the race attacker's rename gap. Old attacker then failed its own symlink creation with EEXIST. It now moves tool-created directories to unique owned in-root names before retrying, preserving still-open files. All outside-sentinel checks, 2,000 primitive attempts, 50 real tool calls and stable positive controls remain. Earlier ToolFiles setup now places its Jobs directory inside root and expects it in listing.

**Evidence:** `media-baseline-build/tests` compiled then failed specifically “Anthropic outside attachment escaped the Sandbox”; existing compatibility passed. `media-build/tests` passed 2/2; media CLI/GUI builds clean. `job-files-baseline-build/tests` compiled then failed outside directory acceptance; both automatic groups passed. First integration `job-files-tests` failed genuine partial-construction cleanup; corrected `job-files-fixed-tests` passed 3/3. `job-files-sandbox-tests` failed the attacker collision above. Final `job-files-reviewed-build` clean, `job-files-reviewed-sandbox-tests` 3/3, `job-files-media-tests` 2/2, `job-files-reviewed-tests` 3/3. Since Jobs and Agent teardown affect all Agents, one shared run was warranted: **`job-files-full-tests.log`: 141 passed, 0 failed, exit 0**. Both `job-files-cli-build.log` and `job-files-gui-build.log` clean, 0 warnings/errors; job exit collected. This is a new actual full run, not relabeled earlier 137-plus-fixture evidence. No live provider/native Linux/x64/OS-confinement claim. Free space last observed 12,286,260 KiB. A wrong endpoint diff path and reconstructed README anchor refused; neither is review/update evidence. Corrected paths/anchors were used afterward.

**Next:** persistence/diagnostic file routing through History → Agent → Sandbox; PTY control files/environment and OS confinement; credential-safe diagnostics; child narrowing; host exercise/prose/grader/preservation. No installation/shared restart/deploy/commit/private-credential read/8088 use.

## Chapter 19 in progress — rooted enumeration and file-tool dispatch verified

The Agent-owned Sandbox now enumerates directories through an opened root-relative handle. `Common/Sandbox.cs` declares the operations/native boundary; `Sandbox/Directories.cs` implements sorted name snapshots and rooted directory classification. A fresh directory open gives each enumeration an independent offset; successful fdopendir transfers ownership to closedir, including exception cleanup. Native names are copied before the next read. The installed Darwin SDK establishes name offset21 and ARM64's unsuffixed 64-bit-inode readdir; Darwin x64 uses readdir$INODE64, Linux uses readdir64/name offset19. Only macOS ARM64 was executed here. Entry names are not open grants: every later descent/open goes through Sandbox again. Canonicalization errors now expose the requested name rather than an absolute intermediate root path.

`Tools/FilesAndShell.cs` routes read_file, write_file, edit_file, list_directory and search_files through Tool's actual Agent parent and its Sandbox—no copied root or service object. My early checkpoint wording suggesting a Tool→Jobs path was incorrect for this saved model; Common declares Tool→Agent. New writes use CreateNew at the actual open rather than a racy existence/size probe. Edits read/check/write through one handle, retaining exact unique-anchor and unchanged-on-refusal behavior. Listing decorates only safely classified directories, leaving inaccessible/outward symlinks as plain names. Confined search uses canonical visited directories for inward-link cycles but still reopens each operation through the boundary. Full-access search retains its old directory-link exclusion; its metadata check is explicitly restricted to full access. Text decoding, UTF-8 output, line ranges, append and grep context formatting remain covered.

Evidence in `/private/tmp/ch19-csharp.6z2ecI`: `enumeration-baseline-2-build.log` is clean; `enumeration-baseline-2-tests.log` passed both primitive groups and failed the added public-dispatch assertion because the then-unintegrated read_file exposed the owned outside sentinel. After integration, `file-tools-tests.log` passed all three groups. Stronger `file-tools-reviewed-tests.log` and final `file-tools-final-tests.log` each passed three groups: real files/ownership/enumeration, hostile symlink swapping, and actual five-tool dispatch. Coverage includes repeated independent enumeration, 160 long UTF-8 names across native buffers, outside reads/writes/edits/listing/search refusals, inward links and cycles, preserved outside bytes, relative error privacy, and 50 real tool calls while the swapper runs in addition to 2000 primitive attempts. Stable positive controls remain; races may safely refuse. The final added symlink-loop error check also passed. This tests actual local filesystem calls, not a mock or general hostile-host/mount isolation.

`file-tools-compatibility-tests.log` passes the existing Chapter3 tool runner group (its internal file/shell, provider signatures, real loop and budget checks); `file-tools-safety-tests.log` passes both policy groups. Reviewed/final test builds and `file-tools-cli-build.log` / `file-tools-gui-build.log` are clean, zero warnings/errors. All exits collected. The final branch-layout expansion was readability-only; final rooted tests compile/run that saved source. No new full-suite result is claimed; previous 137-plus-corrected-fixture evidence retains its earlier scope.

Setup failures retained: the first Darwin symbol probe found readdir$INODE64 absent on ARM64, leading to checking the SDK's architecture macro before choosing symbols; a test patch refused a guessed async signature before writing; the first enumeration build used incorrectly assumed helper/field names and failed compilation. I reread actual Common/Files declarations and corrected the code before claiming behavior. A later renderer lookup used a nonexistent directory; actual media calls are in Engine/Media.cs. None of those setup failures counts as a product test.

**Still partial:** README now distinguishes integrated file tools from unconfined media, Jobs/output, History/state and subprocess paths. No host root flag, kernel process boundary, complete credential-redaction path, child-policy consumer, GUI exercise, prose reconciliation, Chapter19 grade or endpoint exists yet. ch18 is untouched. Next: media reads through Engine→Agent→Sandbox, then Jobs/persistence and process confinement. No installation, shared restart/deployment, commit, private credential access, Bill browser or port8088 use. Free space observed 12,340,744 KiB.

## Chapter 19 in progress — corrected safe mode and rooted primitive; integration pending

I traced Jobs/Pty output/control paths, renderer media reads, History primary/backup/journal/archive writes and the remaining grader credential/helper assertions. These remain integration work, not protected merely by adding a root value. File tools are in `Tools/FilesAndShell.cs`; their private static helpers still use host File/Directory APIs. Renderer media still calls Ref.Data directly, Jobs reopen output by pathname, and state replacement must preserve its existing durability rules.

**Contract correction:** the first policy slice incorrectly retained `wait_for_job` and `kill_job`. Current chapter lines describing safe mode require both removed alongside run_command/send_input; tool_limits removal is optional. `safe-controls-baseline.log` failed on enabled wait_for_job. I corrected the one Configuration predicate and strengthened real dispatch, declaration and historical tool-reference tests for all four names. `safe-controls-fixed.log` passed both focused groups. The earlier two-tool description below is historical, not the current contract; README is corrected. No requirement justified that earlier interpretation.

The new `Common/Sandbox.cs` declarations and `Sandbox/Files.cs` implementation give each Agent one strong Sandbox child. Constructor registration and removal complete both ends; failed construction releases handles and Agent admission; destruction follows Jobs/History teardown. Configuration stores the requested launch root and WithModel preserves it. Sandbox alone owns the effective canonical root and retained descriptor. Empty root retains full access. Confined operations resolve policy paths, then open every directory component with no-follow, close-on-exec handles. Open/create/read/write/mkdir/rename/unlink are handle-relative; no unconfined fallback follows a race failure. Existing inward symlinks work; outside targets refuse. Returned file streams belong to callers. Resolve alone is expressly not an open grant. Unsupported operating systems refuse confined construction. Native Linux code is not runtime-verified here.

This is a **primitive-only milestone, not whole-Agent confinement**. Tools, media, Jobs, persistence and process launches do not yet use it. No host root flag is exposed. README explicitly warns against treating the new library launch value as a completed security boundary. Rooted directory enumeration, state placement, kernel process confinement, redaction and child inheritance remain unfinished.

Evidence remains under `/private/tmp/ch19-csharp.6z2ecI`. `rooted-tests.log` first passed the symlink-race group but failed append with EACCES. A permission assertion (after correcting its platform-analyzer guard) established the real defect in `rooted-mode-baseline-2.log`: a fixed-signature openat P/Invoke passed the variadic mode incorrectly on Apple ARM64, creating the wrong permissions. Apple's primary documentation returned HTTP200 and confirms stack-passed ARM64 varargs: https://developer.apple.com/documentation/apple-silicon/addressing-architectural-differences-in-your-macos-code . The narrow interop adapter now puts mode in that stack slot, while other Unix ABIs use the ordinary declaration. The test requires private 0600 files. `rooted-fixed-tests.log` then exposed a second real issue: FileStream(handle) starts its managed cursor at zero; Darwin positional writes do not honor O_APPEND. Append now initializes the cursor to EOF. `rooted-append-tests.log` passed both groups.

`rooted-full-tests.log` subsequently passed **137 groups**, failing only the new race fixture's demand for a successful operation amid continuous hostile renaming. All refusal during continuous mutation is permissible; stable before/after positives already reject deny-everything implementations. I removed that scheduling-dependent assertion while retaining actual swaps, outside sentinel integrity, read-content checks, pre-opened handle substitution, and stable positives. No product change followed that fixture failure. `rooted-reviewed-tests.log` passes both groups; its build and `rooted-cli-build.log` / `rooted-gui-build.log` are clean, zero warnings/errors, command exit0. This is 137 unchanged groups plus the corrected affected group, **not** a single clean full138 run. The tests exercise actual local files and descriptors, not mocks; no Linux, privileged-mount or general hostile-host isolation claim follows.

I reviewed the saved declarations, implementation and Agent lifecycle delta personally. Additional setup errors were guessed Config/FileTools filenames, an oversized combined source read, and a broad diff-stat that included unrelated Go changes; no unrelated work was edited. Bounded reads supplied omitted content. The earlier core build alone did not verify primitive behavior. These errors and failed baselines remain saved rather than credited as successful tests.

**Next:** implement rooted directory enumeration and route the actual file-tool call paths, then media/Jobs/persistence and OS process confinement. Preserve trusted installation reads, ordinary runtime/toolchain positives and joined lifetimes. Finish credential-safe diagnostic writing, public child-policy narrowing, host/GUI exercise, prose, scoped executable grader and source-only endpoint. ch18 remains untouched. No jobs/children remain at this checkpoint; no install, shared restart/deployment, commit, private credentials, Bill browser or port8088 use. Free space observed 12,362,036 KiB.

## Chapter 19 in progress — policy slice verified; confinement pending

I read the full 894-line current chapter and compared its preserved revision, then read the grader orchestration, generated driver and jobs consumer. Remaining helper/credential assertions and affected persistence/process paths still need detailed source reads. No C# grader adapter or ch19 endpoint exists; ch18 remains the preserved local endpoint.

The first slice extends immutable Configuration with SafeMode/EnableWebSearch, declared in Common and implemented in Serialization. Launch policy narrows normal dispatch, provider declarations, deferred catalogs, historical discovery references, automatic tools and reverse MCP grants without rewriting recorded manuals/grants. Safe mode refuses run_command/send_input before effects but retains file tools. Disabled web-search is hidden and refused during dependency resolution before MCP setup, including startup dependencies. It is a named skill gate, not a network firewall. WithModel preserves policy. Both hosts expose --safe, --safe-mode BOOL, --enable-web-search BOOL and EN_SAFE_MODE/EN_ENABLE_WEB_SEARCH. Flags override environment; values are true/false or 1/0. Hosts default safe=false/web=true; library defaults false/false. There is deliberately no root/confinement flag yet.

Evidence is under `/private/tmp/ch19-csharp.6z2ecI`, with owned env-i offline SDK/home/tmp. `policy-tests-build-3.log` is clean; `policy-tests-3.log` passed two focused groups with real Agent/Skills/Engine/files and the existing HTTP MCP fake. Checks cover shell refusal before effects, file positives, frozen grant projection, direct/dependency/startup web denial with zero HTTP requests, explicit web activation/tool use, sibling isolation and model-selection preservation. `policy-cli-build-2.log` and `policy-gui-build.log` have zero warnings/errors. `policy-full-tests.log` is one complete **136 passed, 0 failed** run, exit 0.

`policy-cli-exercise-2.log` and `policy-cli-result.json` record four fresh CLI sessions, all exit 0: restrictive environment, enable overrides, disable overrides and defaults. All performed file writes; shell effects occurred only with safe mode off; web load followed policy. Synthetic local manuals, zero inference, no credentials. GUI policy wiring is build-verified only; runtime exercise remains pending. I reviewed the complete source delta against ch18. Final readability-only changes expanded Config signatures, clarified grant-query comments and renamed a local; `policy-reviewed-build.log` compiled cleanly. README identifies main as WIP. ch18 is untouched.

Retained failures: the first patch refused an ambiguous EnabledTools anchor before saving, but misplaced set-e allowed an unchanged build (`policy-build.log`), not policy evidence. Corrected core patch compiled (`policy-build-2.log`). Fixture builds used internal Skills APIs and then nonexistent History.Context; actual public APIs are LoadableSkills and History.View. Guessed declaration paths/searches refused before editing. Runner insertion initially refused after a separate reference filter had saved. Host anchors refused before writing; a build used nonexistent Cli/Cli.csproj (`policy-cli-build.log`); actual project is Cli/Ch04.csproj. Initial CLI exercise expected wrong success wording; its fresh replacement checks structured is_error/tool_references. Progress insertions refused reconstructed heading assumptions (the file starts directly with a level-two entry). These were my setup/fixture errors, not product defects. They reinforce the need to read exact saved signatures and validate edit anchors before writing.

**Next:** complete reads of Jobs/Pty, media, persistence/log paths and remaining grader-helper assertions. Implement race-safe Agent-owned rooted I/O/state and OS process confinement, with useful positive controls; keep trusted skill reads distinct. Broader credential-safe diagnostics, public child-policy narrowing (including mutation on widening errors), GUI exercise, C# prose, executable grading and endpoint preservation remain pending. Safe mode is not filesystem confinement. No Chapter 19 acceptance, live-provider, audible or human-accessibility claim. No jobs/children remain; no installation, shared restart/deployment, commit, Bill-browser or port8088 use. Last free-space observation: 12,489,916 KiB, not a future guarantee.

## Chapter 18 complete within local scope — prose, source review and 191-file endpoint

I completed the personal C# source/ownership review and revised the actual Chapter 18
manuscript. Its current specification teaches the existing HTTP MCP boundary, real
Job → Jobs → Agent → Engine status path, actor-owned replacement Configuration,
counted direct-send guard, one capability table and unchanged per-response accounting.
Historical search surveys, Go architecture criticism and model-led trials remain
explicitly attributed history, not C# evidence. The shipped web manual example is
byte-identical to `Skills/web-search/SKILL.md`, including `/v2/mcp`. CLI forms, current
WebSocket controls, component selectors and 135-point scope match the saved code.
Prose corrections include immediate unload authority revocation, historical chapter
numbering, and actual acknowledged GUI behavior (no pending-picker-disable claim).

Final review found one real privacy gap missed by the first executable grade:
HTTP-200 JSON-RPC error.message could echo the bearer, unlike non-2xx diagnostics.
`Mcp/Http.cs` now redacts only that decoded diagnostic before publishing it through
the existing receive path. Result content and identities are untouched. New real
HTTP JSON/SSE cases in `Tests/McpHttpTests.cs` preserve useful quota diagnostics and
assert no token disclosure. `review-rpc-baseline.log` exited1, two groups passed and
one failed with “Error leaked bearer.” The corrected `review-http-tests.log` passed
all three groups/exit0. `review-cli-build.log` and `review-gui-build.log` exited0 with
zero warnings/errors. These and all following receipts are under
`/private/tmp/ch18-csharp.1TITqt`.

I then suspected scalar SSE JSON could strand a request. That diagnosis was wrong:
the saved reader already explicitly rejects non-object envelopes. The added scalar
case passed with no production change (`review-scalar-baseline-2.log`, three groups,
exit0). The earlier attempt refused an ambiguous test-edit anchor, but my shell
continued into unchanged tests and returned0; `review-scalar-baseline.log` is not
new-regression evidence. No scalar fix is claimed. Oversized source reads and one
nonexistent diff filename were review/setup mistakes; bounded saved-source reads
supplied the actual contracts. I did not change code based on the mistaken diagnosis.

Corrected CLI/GUI artifacts were pinned in the existing `grade-hosts` directory.
`grade-review.json` / `grade-review.stderr` exited0: score/max135, passed:true,
all nine checks passed. Four clean host sessions and 22 local model requests:
- `tmp/web-grade-623592439`: main CLI10, missing-auth CLI2;
- `tmp/status-grade-1623856814`: GUI8, four before/after pairs, empty fatal;
- `tmp/status-grade-4145635303`: GUI2, one before/after pair, empty fatal.
The launcher checks actual clean exits. The original Go route and 15-point exclusion
are unchanged. The later scalar-test addition changes test input only; library/host
source is identical to this final grade. No new full component pass is claimed:
carry HTTP130, focused status/model and Responses, 133 unchanged plus corrected tool
group, and actual CLI/two-client browser results for their unchanged paths. Do not
relabel them as a single full134 run. Earlier failed120/135 and intermediate passes
remain in the entries below with their distinct inputs.

README now states the Chapter18 endpoint/scope, retains Chapter16 auth limitations,
and records the final correction. Scoped changed-line whitespace and manuscript
`git diff --check` passed; project references remain relative and resolve. Preserved
`solutions/edition-3/ch18`: 191 source/support files, every byte compared with main,
no bin/obj, credentials, private fixtures or runtime logs. Existing NuGet and speech
support files were unchanged from ch16. ch01–16 were not modified; ch17 still uses
ch16 as its runnable endpoint. Free space12,629,524KiB at final inspection.

Chapter18 is complete only within the permitted local scope. No hosted web request,
live provider availability, actual billing, audible quality or human-accessibility
completion is claimed. Live OAuth remains deferred for Bill. No installs, deployment,
restart of Bill's application, commits, pushes or shared cleanup occurred. Next:
read current/preserved Chapter19 and its actual contracts, then continue personally
without workers. Do not repeat unchanged prior-chapter campaigns.

## Chapter 18 — shipped skills reconciled and executable grade verified; chapter open

I completed the bounded grader/contract review and found two real shipped-artifact
omissions: `Skills/ensemble/SKILL.md` did not grant `agent_status` or `wait_for_job`
and did not list `web-search`; there was no shipped web-search manual. The primary
now grants those ordinary tools and permits that skill. The new manual preserves
the chapter's search/scrape guidance and hosted URL, uses canonical `transport: http`,
and names `FIRECRAWL_API_KEY` without storing any value. No service was contacted.
Earlier custom-skill tests did not establish this shipped integration. No library or
host source changed in this slice; both hosts were rebuilt to package the manuals.

The explicit Chapter18Executable route reuses the existing web peer and launchers,
keeps all seven web behavioral weights plus the two status/accounting weights:
135 points. Original Go route/defaults remain intact; its 15 build/import points
are excluded, not silently awarded. CLI and GUI paths plus `-status-skills` are
required. C# uses canonical HTTP spelling, root-agent prompts, actual selected-model
usage acknowledgements and joined SIGTERM GUI shutdown; CLI observes its final
marker before EOF. The web test loads the shipped manual, changing only URL and
auth-variable name to the owned fake boundary. Status loads the shipped ensemble
primary, not a status-only fixture. Current text/nullable meter semantics and exact
OpenAI/default price rows replace historical model aliases/price assumptions.
Four OpenAI turns exercise mini → 4o → unknown → mini with preserved dialogue,
unchanged pre-turn totals and no retroactive repricing. A separate unpriced
Anthropic status turn retains nonzero cache-write coverage, without an invented
OpenAI cache-write field. Correlated tool IDs, real is_error refusal, no activation
or post-unload effects, planted page bytes and missing-auth diagnosis remain checked.

Evidence under `/private/tmp/ch18-csharp.1TITqt`:
- `grader-tests-final.log`: eight named contract groups (with subcases), PASS/exit0;
  original environment/wire/import/model/status tests plus new current-wire,
  exact correlation (duplicates/wrong ID/assistant text rejected), shipped-manual
  boundary substitution and C# refusal/error-flag tests. `grader-build-final.log`
  clean. The subsequent CLI flag-help edit is nonbehavioral.
- `grader-cli-build.log`, `grader-gui-build.log`: exit0, 0 warnings/errors. Final
  artifacts pinned in `grade-hosts`; both hosts' ensemble and web-search manuals
  were byte-compared with main after the grade.
- `grade-final.json` / `grade-final.stderr`, job3216 exit0: score/max_score135,
  passed:true, all nine checks passed. Four clean host sessions, 22 model requests:
  main CLI10 + missing-auth CLI2 + model-switch GUI8 + cache-write GUI2. Retained
  final fixtures: `tmp/web-grade-1913390863`, `tmp/status-grade-2470054575`,
  `tmp/status-grade-1655105624` under scratch. The GUI observation files have empty
  fatal fields, four/one before-after meter pairs and eight/two requests. Launchers
  check actual process exits; runtime artifacts stay outside the endpoint.
- Four invalid scope combinations returned2 before launch (missing GUI, missing
  skills, skills on another chapter, jobs-only conflict). Pinned skill byte checks
  and report/request inspection exited0. Free space12,698,264KiB at inspection.

Failures/corrections remain distinct: my first prevalidated launcher patch was
based on reconstructed signatures and refused before any write. I reread the
actual functions and applied an exact replacement. A wrong absolute gofmt path
failed setup after command edits; the approved wrapper supplied the real path.
`grader-tests-initial.log` failed compilation because Recorded.Body is bytes, not
string; no behavior was tested. `grade-1.json` exited1 at120/135: all behavior except
the refusal-text predicate passed. C#'s real error is `tool is not enabled: NAME`,
not the Go checker's “unknown tool”; only the executable mode now accepts it,
still requiring an error result and independently checking authority/no effects.
`grade-2.json` passed135/135/exit0 with a generated web manual; it is superseded for
shipped-manual coverage by the final pass. I initially miscounted the scope as80;
the actual table is100 web +35 status, excluding15 Go points. No product defect
was hidden by the wording adaptation. No new full component regression is claimed;
carry the earlier relevant suites and host-control exercises for unchanged code.

README now documents the shipped skills, actual executable command and limitations.
Next: Chapter 18 C# teaching prose and personal final source/architecture review,
then source-only endpoint and chapter handoff. No Chapter18 endpoint yet. Hosted
MCP/live provider work, live OAuth, real billing, audible speech and human-accessibility
completion remain unclaimed. Do not replay unchanged prior-chapter campaigns.

## Chapter 18 — CLI/GUI model controls and choice projection verified; chapter open

The existing actor selection core now has real host controls. CLI `chat` accepts
`/model NAME` and read-only `/model`; JSON accepts `kind:set_model, model:NAME`.
Acknowledgements include current model; busy/invalid/incompatible controls return
recoverable errors without terminating stdin. Request admission and direct-tool
provenance read the current Agent Configuration rather than the startup copy.
GUI uses its existing WebSocket `set_model`/`text` control and usage-frame channel.
`selected_model` and `model_choices` update all clients and funding text; refused
choices return an error plus the unchanged current frame. No second model owner,
account catalog, model preference or meter was added.

Common/Media.cs declares one immutable value-like table containing the same five
preexisting capability rows. Engine.LookupModel and public Engine.ModelChoices
both use it; the latter includes unknown current selection and filters candidates
with the actual Configuration.WithModel validator. No aliases, prices or availability
claims added. The picker uses textContent, distinguishes local capabilities from
account access, and wraps within the owned 360px exercise. Billing/surface/account
remain immutable; metering stays per completed response identity. Selected model
and last measured model can legitimately differ. Snapshot config is documentary:
restart restores conversation but uses launch model and an empty process meter.

Evidence under `/private/tmp/ch18-csharp.1TITqt`:
- `model-host-model-tests.log`: 2/0, exit0, extends the existing core groups with
  exact retained media/stream facts, vendor separation, unknown current inclusion,
  duplicate-free order and surface-compatible choices.
- `model-host-responses-tests.log`: 8/0, exit0 for reasoning/surface/transport risks
  of the shared capability-table refactor. Earlier status/core and the 133+1
  regression coverage remain evidence for unchanged paths, not a new full run.
- `model-host-cli-build.log`, `model-host-gui-build.log`: clean, zero warnings/errors;
  sequential checks/builds exited0. Production was unchanged after these builds.
- `exercise-model-hosts.mjs` / `model-host-exercise-5.log`: exit0;
  `/private/tmp/ch18-model-hosts.bzI9o9/result.json` passed:true, six actual loopback
  requests, four clean host exits (JSON CLI, chat CLI, GUI, same-workspace GUI restart).
  Held requests prove busy refusals; later selection and turns succeed. Wire models
  change while prior dialogue remains. Two real isolated browser clients agree;
  incompatible choice refuses, unknown current ID displays at 360px without overflow,
  refresh sends no request. Mixed known-model cost stays .0003975, then unknown work
  keeps valid counts and cost null even after selection returns to known. Restart
  uses launch mini although saved config says 4o, restores dialogue, resets meter
  and sends no inference. No browser JS errors. Actual DOM/geometry and controls,
  not visual assessment, hearing or human-accessibility acceptance.

Four failed executable fixture runs remain in logs -1 (unnumbered), -2, -3, -4
and their owned directories: I used nonexistent --chat instead of positional chat,
waited for GUI funding on CLI stderr rather than its frame/banner, expected a
lifecycle wrapper instead of actual type:turn_ended, and read one page before its
acknowledged frame merely because a different socket had advanced. Corrected by
actual command/protocol reads and per-page observation waits; no product contract
was weakened. These were fixture failures, not passed exercises. All exits collected.

README documents the exact host controls and wire fields. Scoped source review
checked capability declaration/implementation, current Configuration access,
refusal recovery, authoritative projection and preserved meter/route ownership.
No prior endpoint changed. No Chapter18 grade/prose completion/endpoint yet.
Next: reconcile remaining Chapter18 contracts with the existing executable grader,
adapt only real C# host/wire assumptions while preserving Go routes, then C# prose,
final review and source-only endpoint. Live OAuth remains deferred; real-provider
availability/billing, audible speech and human accessibility remain incomplete.

## Chapter 18 — status and model-selection core verified; hosts next

`agent_status` is a normal grant-controlled built-in appended to ToolKind (prior
numeric identities preserved). Tools/Status.cs reaches Job → Jobs → Agent → Engine
and formats one immutable meter snapshot as the chapter's **labeled text**: model,
last response, session total, cache hit rate, cost. No new meter/counter/service
copy. Unknown counts/prices remain explicit; empty totals/cost are zero with no
invented latest response. Costs say estimate. The tool accepts no model argument
and reads the response that requested it before the later answer adds usage.

`Agent.SetModelAsync` is a mailbox control, not a model tool. Agent owns the one
replaceable immutable Configuration. Actor refuses active/queued work and pending
foreground/initialization controls. An ingress-locked count also refuses changes
while any direct Engine send runs; Engine captures the selected value and releases
the guard in finally, including credential/HTTP/cancellation failures. No second
scheduler or engine. Configuration.WithModel validates a same-vendor, same-route
replacement, preserving endpoint/key/surface/billing/account and application
policies, clearing stale resolved identity, and selecting fresh model-specific
reasoning defaults. Existing explanation enforcement is not weakened. History,
Engine, lens and per-response-identity meter rows remain attached and unchanged.
Changing selection does not reprice prior work, write preferences or prove model
availability. Common owns fields and callable declarations; Agent/Engine/Tools/
Serialization own implementations. CLI/GUI selection and choice projection are
**not implemented yet**. No Chapter18 grade, prose completion or endpoint.

Evidence in `/private/tmp/ch18-csharp.1TITqt`:
- `status-tests-7.log`: 2/0, exit0 for that selector; `model-tests-4.log`: 2/0,
  exit0; final `model-tests-5.log`: 2/0, exit0 adds overlapping direct sends.
  Real HTTP covers selected-model wire, preserved conversation, active and paused
  queued-turn refusal, unchanged owners/meter at selection, mixed known-price
  rows then unknown work and return-to-known, invalid/incompatible/closed refusal,
  HTTP failure/cancellation releasing guards, capability defaults and plan-route
  retention. Status covers empty/mixed/unknown, Agent isolation, strict arguments,
  primary grants, actual model-called result timing and canonical correlation.
- `status-model-full-tests-3.log`: **133 passed, 1 failed**, exit1. The sole failed
  group was the old built-in catalog assertion; no production change followed.
  Corrected `status-model-tool-tests.log` passed that complete real file/shell,
  provider declaration, ownership, multi-call actor-loop and tool-budget group,
  exit0. Combined coverage is all 134 groups, **not** a claimed single full pass.
- `status-model-cli-build.log`, `status-model-gui-build.log`: clean builds,
  zero warnings/errors, sequential job3134 exit0.
- `exercise-status-cli.py` / `status-cli-exercise.log`:
  `/private/tmp/ch18-status-cli.x8xxk0o5/result.json`, passed:true, CLI exit0,
  two actual loopback model requests. Primary grants only status (no file/shell
  declarations), exact tool ID and report consumed in second request, measured
  fresh100/read20/out10, hit16.7%, estimated USD0.0000225. Not provider/billing proof.

Failures/corrections retained: my first status implementation/test used JSON,
contrary to the chapter's text requirement (status-tests-1 passed the wrong shape,
then corrected before acceptance); fixture guessed Agent.InitializeAsync,
redundantly installed tools already installed for Skills, treated ToolReturned as
nested rather than Event content, guessed MeterSnapshot.Models and a helper
signature, and compared value-like Usage references rather than counts. Logs
status-tests-3 through -6/model-tests-1 and -3 retain these compile/assertion
failures. The initial full run's fixed tool count was obsolete. My replacement
fixture guessed a nonexistent clear_tool_limits name and collided with an existing
`expected` variable: full-tests-2 was a compile failure and full-tests-3 exposed
that wrong list. The final explicit list uses micro_handoff and agent_status;
focused tool group passed. No production behavior was weakened for these errors.
Repeated guessed APIs and overly large reads cost time; inspect exact saved
members and small sections before the next edit, rather than adding more receipts.

Next: host `/model`/JSON control, GUI selected-model/choices/picker with explicit
busy refusal, no persisted model preferences or provider/billing changes. Reuse the
existing small capability lookup as the authoritative choices source; do not make
up aliases/prices/account availability or duplicate the catalog. Then remaining
contracts, executable grade, public exercise, C# teaching prose and endpoint.
Live OAuth remains deferred; no approved live provider credential source, native
speech or human-accessibility completion. Prior chapter endpoints untouched.

## Chapter 18 — HTTP MCP slice verified; chapter remains open

Read all 1,163 lines of the current manuscript, compared the preserved revision,
and inspected the web/status/model-switch specification and original web grader.
No Chapter18 grade or endpoint exists. Main advances; ch01–ch16 are untouched.

Implemented `McpHttpTransport` through the existing MCP boundary. Common owns its
state/declarations; `Mcp/Http.cs` owns POST/JSON/SSE/session handling. Each POST is
independently cancellable; replies enter the existing client ID table through a
bounded queue. JSON and SSE match the request ID; SSE ignores unrelated replies,
handles progress and multiline data with CR/LF/CRLF, and returns before stream EOF
once the matching reply arrives. Notifications require successful HTTP but no
RPC body. Disposal cancels/joins held POSTs. No redirects, cookies, retries, GET
stream, resumption or new protocol-version claim. Existing numeric request IDs
and historical initialize version remain unchanged.

Strict skill records now accept HTTP `url` and optional `auth-env` while refusing
incompatible fields/unsafe URLs. Hosts supply a separate explicit environment
construction input; Skills retains only referenced credential values, never
interpolates them or records them, and clears them on destruction. Missing values
log their variable name and omit Authorization. Skill activation/unload uses the
existing preparation/actor commit/ownership paths. Remote error text is bounded;
HTTP status/body is credential-redacted and oversized error bodies are omitted.
The MCP reader now exposes bounded RPC error messages instead of only a generic
error, allowing useful quota diagnostics without poisoning the HTTP connection.

Scratch `/private/tmp/ch18-csharp.1TITqt`, text pointer `/private/tmp/ch18-csharp-current`.
`http-full-tests.log`: **130 passed / 0 failed**, exit0; includes three new HTTP
MCP groups and existing shared regressions. `http-cli-build.log` and
`http-gui-build.log`: clean, zero warnings/errors. Focused `http-tests-3.log` passed
3/0 before the added CR/CRLF, oversized error, and held-shutdown assertions; the
full suite verifies that final version. Real sockets cover JSON/SSE, open-ended
stream termination, concurrent correlation, sessions/bearers, 429 and safe body,
202-without-reply, malformed/wrong-ID/EOF errors, redirect refusal, per-call cancel,
held-call shutdown, actual skill activation/unload, missing auth, and data/URL
preservation. No fabricated provider, web or OAuth success.

Actual CLI exercise `exercise-http-cli.py`, `http-cli-exercise-2.log`, result at
`/private/tmp/ch18-http-cli.ibpgcioc/result.json`: five loopback model requests,
four MCP POSTs, one remote tool effect, clean CLI exit0. Activation discovers and
grants, returned injection-looking text remains tool data with its URL query
intact, unload removes declarations, and a later requested call refuses without
HTTP. Secret absent from stdout/stderr/canonical log. No shell authority granted.

Failures retained: initial wrapper lacked owned home/tmp (setup, no build);
first real build exposed a missed record/parser replacement (corrected actual
saved declarations). First test compile used guessed lifecycle API names; second
run's two assertions misused the no-Skills policy query and default JSON field
serialization. Corrected assertions inspect actual installed tools/result parts.
Strengthened shutdown test initially omitted the existing ObjectDisposedException
terminal contract. First CLI fixture used wrong .md layout and failed startup;
correct SKILL.md layout passed fresh. These were implementation/fixture mistakes,
not credited as successful coverage. README documents the bounded subset.

**Next:** implement `agent_status` over Engine's existing meter and the actual
same-provider model selection/actor contract, then reconcile remaining requirements,
run a scoped executable grader and suitable application exercises, revise/review
C# teaching prose, and preserve an endpoint. No live credential source is available;
live OAuth stays deferred. Do not repeat unchanged Chapter16/15/14 campaigns.

## Chapter 17 — field-report/C# exercise review complete; live crossover open

Chapter17 adds no implementation, grader or separate source snapshot. Read the
entire current manuscript and compared the preserved revision; its scope/link
corrections agree with the existing grader's explicit Chapter17 exclusion.
`solutions/edition-3/ch16` remains the runnable endpoint. Main and ch01–16 source
were not edited, and no unchanged Chapter16/15/14 campaigns were rerun.

Updated `book/edition-3/chapter-17.md` with a C# ownership/scope explanation and
actual CLI reproduction instructions, preserving Bill's historical crossover
narrative and attributed second-edition task results. Replaced old Go build/auth/
reasoning settings with .NET10, explicit shared credential directory/account,
Responses plan selection and supported high/auto settings. Added job observation
to the small exercise skill and distinguished workspace instructions from a
sandbox. Fixed two relative historical evidence links. Existing production
configuration is immutable; the historical model picker is not a new requirement.

Local checks in `/private/tmp/ch17-csharp.idkg9uvs` (text pointer
`/private/tmp/ch17-csharp-current`) exited0: all four shell blocks parse; relative
links exist; fences/whitespace clean. Fresh copies of historical before-files
reproduce six tests with three failures/three errors and exit1 as expected
(`baseline.stderr`), not a repaired program. The already-built current CLI loaded
the exact exercise skill using dummy API funding and unreachable loopback, then
exited0 on empty stdin (`startup.*`). Its owned log contains only `skill_loaded`
and `agent_shutdown`; both input fixture files remain byte-identical. No HTTP,
new login, code repair or real-model result is credited. Prose was reread after
editing. The documented build invocation was source-checked, not rerun.

**Limits and next:** a fresh C# model-led task, live OAuth, daily-use crossover,
private-memory migration and replacement of Bill's running application remain
unperformed. No approved provider credential source exists in this session; live
OAuth is deferred to Bill's joint session. Continue Chapter18 personally: read
both manuscripts and actual contracts before changing main. No new Chapter17
implementation or grader campaign is appropriate.

# Edition 3 C# — current progress

## Chapter 16 — completed local endpoint; live OAuth deferred

`solutions/edition-3/ch16` now preserves 184 source/support files, individually
byte-equal to main at finalization. No bin/obj, credentials, captures or runtime
data are included. Chapters 01–16 are now preserved within their documented local
scopes. Next: read current and preserved Chapter17, then continue personally.
Live OAuth stays deferred for Bill's joint session; no workers were used.

- Completed C# teaching revision in `book/edition-3/chapter-16.md`: concrete
  Ensemble → Credentials and Engine → Agent → Ensemble paths, Common declarations,
  protected v2 store/v1 compatibility, PKCE/issued registration and verified identity,
  same-source renewal, terminal-success gate, public summaries, explicit funding and
  fresh account discovery. Replaced Go interface/obsolete cache-price requirements;
  retained Bill's narrative, accessibility rationale and historical failed shutdown
  with explicit attribution. The original Go grader route remains intact.
- Final personal source review covered declarations and bidirectional ownership,
  root/store destruction, strict store codec, login save-before-exchange, signed
  identity, locked renewal, Engine transport/projection/parser, host startup/auth,
  browser privacy and summary projection. Removed one duplicated CLI plan-surface
  default block. `review-cli-build.log` is clean, exit0; documented standalone
  status command against a new empty owned store exited0 with `[]` and no inference
  (`review-status.stdout/.stderr`). No behavioral contract changed in that cleanup.
- Reused valid prior evidence: Auth20 (signed local issuer, real files/sockets and
  two processes), Responses8 after the headerless correction, earlier full127
  **before** that correction, clean hosts, and scoped executable70/70 passed:true,
  actual exit0. Final grade has seven clean host sessions/eight inference requests,
  causal summary gate and signed-out zero requests. Exact receipts/limits are in
  the entry below; GUI/account/catalog/funding campaigns are in earlier entries.
  No unchanged chapter/browser/test campaign was rerun for timestamps.
- Manuscript scope is 70 points: leakage15, format15, summaries15, limits15, cache10.
  OAuth15/refresh15 are omitted, not simulated credit. Current cached HTTP200 primary
  limitations page requires namespaced tools but does NOT require an empty namespace
  description; the prior checkpoint's stronger statement was incorrect. No code
  change was needed: https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations .
- Prose work had avoidable failed draft anchors/keys based on an incorrect outline.
  Those attempts refused before writing; I reread actual headings/text and applied
  a validated staged revision. Final reread corrected reversed draft grading weights,
  one-request callback wording, required-usage wording, and a removed exercise's
  historical cross-reference. Endpoint copy initially retained a work-in-progress
  README title; fixed before finalization and repeated the full byte comparison.
- Final source/prose whitespace and code-fence checks passed. Last observed free
  space 13,200,060 KiB. No installs, deployment/restart of Bill's app, commits, private
  credential searches, port8088 use or shared cleanup. Live OAuth, successful host
  signed exchange, real plan inference/billing, live availability, audible speech
  and human accessibility remain explicitly unverified, not endpoint credit.


## Chapter 16 — scoped executable grade and headerless Responses verified

Chapter 16 remains open for C# prose/final source review and endpoint preservation.
The local executable transport scope now passes 70/70; OAuth/refresh are explicitly
excluded from its denominator. Live OAuth stays deferred for Bill's joint session.
No workers, installation, restart of Bill's app, or commits were used.

- `ch16_executable.go` reuses the existing peer, summary gate and launcher through
  explicit `executableOnly` mode. The original Go route remains separate. C# starts
  with protected synthetic grants, never an issuer override. Five original weights
  total 70; OAuth-flow 15 and token-refresh 15 are not scored here. Signed library
  issuer/refresh evidence remains separately in prior entries, not host OAuth proof.
- Checks inspect real request JSON: exact bearer, Responses/store/stream, developer
  instructions or current supported instructions field, namespaced tools, completed
  file output correlated with its call, encrypted replay, distinct summaries before
  terminal release, default plan omission/configured API cache key, quota/admission
  and signed-out refusal without fallback. Refusals require clean host exits and one
  durable error, not raw provider bodies. Direct socket secrecy checks exclude tokens
  and encrypted replay payloads. Fake cache-marker fields are not production features.
- Current primary documentation permits `instructions` or developer-role messages:
  https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations
  and terminal success remains required:
  https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference .
  Both fetched HTTP200 from cache; this is protocol documentation, not live behavior.
- First real grade (`grade-2.json`, exit1, 0/70) exposed a real product gap: a plan SSE
  response without Content-Type was parsed as whole JSON. `Engine/Streaming.cs` now
  selects incremental Responses parsing when streaming was requested and media type
  is absent or text/plain. Explicit application/json retains authoritative plain
  parsing; other providers and non-success HTTP paths remain unchanged. Tests exercise
  both fallback headers and failed terminal gating. The real summary gate subsequently
  proves the host delivers deltas before the response body completes.
- Scratch `/private/tmp/ch16-csharp.WMIdlt`: `headerless-responses-tests-2.log` 8/0,
  `headerless-cli-build.log` and `headerless-gui-build-2.log` clean. Correct sequential
  job2867 exited0. Earlier `headerless-responses-tests.log` ran 119 non-Responses groups
  because I supplied the wrong selector; it is regression-only, not fix coverage.
  Prior full127 includes earlier Engine/auth changes, not this later parser edit.
- `grader-tests-final.log`: five named adapter/legacy-environment tests pass.
  `grader-build-final.log` clean. Final `grade-final.json`: score70/max_score70,
  passed:true, actual exit0 (job2874). Seven actual host sessions, eight inference
  requests, clean exits; signed-out sends zero. Plan session has three requests and
  causal gated summaries. Private observations retained at
  `/private/tmp/ch16-csharp.WMIdlt/tmp/subscription-csharp-grade-1369143377`.
  `grade-3.json` previously passed before strengthening socket secrecy; final supersedes
  it. Three invalid scope combinations exited2 (`grader-scope-refusals.log`). Scoped
  whitespace checked seven changed source files. Existing browser/auth campaigns were
  not repeated; local fake traffic is not successful live login, permission or billing.
- Harness failures retained: first guessed-anchor script saved only an import/mode
  fields before refusing; finished using actual source with prevalidated all-or-none
  replacements. `grader-compile-1.log` failed on guessed socket/helper names; corrected
  to actual `conn` and local string accessor. `grade-1.stderr` is -chapter versus -ch
  invocation refusal, no grade. First real grade's signed-out seed used empty scope
  instead of required null, correctly refused by the store. Observation inspection
  initially mishandled a null request list; corrected without changing the grade.
  I also initially stated the reduced maximum as60; actual weights total70.

Next: finish manuscript/implementation contract reconciliation and teaching review,
then preserve Chapter16's source-only endpoint. Keep live OAuth explicitly incomplete.

## Chapter 16 — bounded plan HTTP 401 recovery verified locally

Chapter 16 remains open: no executable grade, prose completion, endpoint or live
OAuth. Next: finish contract/grader reconciliation, then C# prose and preservation.
Live OAuth remains deferred for Bill's joint session. No workers were used.

- Contract review found a real gap: proactive refresh already had bounded transient
  retries, but inference HTTP 401 did not renew. Credentials now accepts the exact
  rejected bearer internally, rereads under its existing operation/file locks, and
  either reuses a newer grant or rotates the rejected generation. The public bearer
  API, ownership, storage and five-minute refresh policy are unchanged. Publication
  still precedes returning a token. There is no cached grant or new retry owner.
- Engine permits one plan-only HTTP 401 retry with identical body bytes and the
  durably resolved bearer. It disposes the first response without parsing/publishing
  it. A second 401 reaches ordinary error parsing. Other statuses, API-key requests
  and interrupted successful streams are not retried; no payment fallback. One
  logical capture and only the successful final response enter the process meter.
  Credential failures on either acquisition path use the existing token-free typed
  refusal; cancellation/disposed-owner behavior stays distinct.
- Tests extend the existing signed RSA issuer fake with an actual SSE inference
  endpoint. Six scenarios cover successful recovery, second 401, 403, API-key 401,
  terminal refresh and bounded transient failure. They require exact request bodies,
  exact bearers/counts, durable rotated reuse, one metered completion or none, and
  no Engine writes to History. A second real Credentials owner proves delayed 401
  reuses a newer stored token without an extra rotation. This latter case is two
  owners in one process; the preexisting cross-process rotation test also still passes.
- Scratch /private/tmp/ch16-csharp.WMIdlt: recovery401-auth-tests-2.log reports 20/0;
  recovery401-responses-tests.log reports 8/0; recovery401-cli-build.log and
  recovery401-gui-build.log have zero warnings/errors. Sequential job 2809 exited 0.
  Because shared Engine response lifetime changed, a full regression was justified:
  recovery401-full-tests.log reports 127/0, job 2811 exited 0. No prior-chapter browser
  or grading campaign was replayed. README documents the exact retry scope.
- Failed attempt retained: recovery401-auth-tests.log is a compiler failure from my
  fixture's nonexistent Json helper. Review also caught my insertion inside the
  token-route switch case; moved inference into its own case and used existing Reply
  before the successful run. These were fixture errors, not passing behavior.
- Original Chapter16 grader expects host issuer override and historical explicit
  OpenAI cache fields. Neither has been adopted. Production issuer stays fixed;
  successful signed local exchanges remain library-fake evidence, not CLI/GUI OAuth.
  Any executable adapter must state exclusions visibly rather than count seeded
  grants as successful login or add unsupported production fields to satisfy fixtures.
  The adapter decision is still unfinished; original Go routes are unchanged.

Earlier funding/discovery/browser evidence remains scoped to its unchanged paths.
No live account changes, installations, deployment/restart, commits or immutable
chapter-endpoint edits occurred. Real OAuth/billing/native hearing/accessibility
remain unverified.


## Chapter 16 — persisted plan selection and funding announcements verified locally

Chapter 16 remains open: no executable grade, prose completion, endpoint or live
OAuth. Next: reconcile remaining chapter contracts, then executable grading/prose
and preservation. Earlier slice evidence applies only to unchanged behavior.

- Credentials owns one nullable selected-plan label in its version-2 protected
  file. Version 1 remains readable without migration and means no saved selection.
  Strict version-specific fields reject missing/invalid/unknown selected accounts.
  Successful signed login writes grant and selection in one replacement; pending
  registration, failed exchange/identity, refresh, logout and signout do not choose
  a different account. Identity-only login selects plan but cannot infer without
  permission. Logout retains selection: absence of credentials never grants API use.
- Public SelectedPlanAccountAsync reads fresh under the existing operation/file
  locks and lifetime. No token, second store, host cache, GUI preference or runtime
  mutation was introduced. Both hosts read only an explicitly supplied credential
  directory. Precedence: explicit flags, environment, saved plan choice, legacy API
  default. Overrides affect this launch only. Omitted plan vendor/surface default
  to OpenAI/Responses at the composition roots; explicit incompatibilities refuse.
  Library Configuration still requires explicit Responses for plan mode.
- Configuration.FundingDescription is one credential-free host description of the
  immutable model/funding choice. CLI prints to stderr before accepting inference;
  GUI usage frame supplies a visible text-only wrapping banner before prompting,
  including replay/second-client connection. No History/system-instruction mutation
  or narration. Login confirmations explain next-launch choice/current GUI isolation.
- Scratch remains /private/tmp/ch16-csharp.WMIdlt. funding-auth-tests.log: 18/0,
  exit zero, including signed login selection, failed exchange/identity no-selection,
  identity-only selection, later-account selection, logout retention, second-owner
  reread, version-1 compatibility/no rewrite and malformed v2 rejection.
  funding-responses-tests-2.log: 8/0, exit zero. Final host logs
  funding-cli-final-build.log and funding-gui-final-build.log are clean (0 warnings,
  0 errors), exit zero. No unchanged full121 or prior-chapter campaign was rerun.
- funding-exercise-2.log: passed:true, exit zero. Real executable evidence:
  /private/tmp/ch16-funding.hCwol9/result.json; 10 scenarios, six loopback inference
  requests. Saved plan with omitted billing/account/vendor/surface uses exact work
  bearer; explicit account flag beats environment; API flag/environment override
  uses exact ambient key without saving a different choice; incompatible vendor
  refuses. Actual CLI logout retains plan selection; subsequent CLI and GUI requests
  refuse with zero provider requests. Two GUI launches sent requests under observed
  funding banners; a third exercised signed-out refusal. Two browser clients agree,
  banner fits 360px; all three GUI exits zero. All successful CLI exits zero; expected
  refusal CLI exits one. Synthetic grants and isolated Chrome, not OAuth or billing.
- Failures preserved: funding-responses-tests.log was 7/1 because I unnecessarily
  weakened the explicit library surface contract. Restored it; host defaults only.
  One patch saved that restoration before refusing an ambiguous CLI anchor; the
  next prevalidated patch refused a guessed GUI anchor without writes. Corrected
  from actual source. funding-exercise.log exited one: obsolete fixture expected
  "requires sign-in" instead of the existing safe PlanAuthorizationException text;
  no product change. Corrected fixture and reran fresh, unassisted. First-run data:
  path recorded in that log. Earlier intermediate builds do not cover later edits.

README documents selection/override rules and format compatibility. Live OAuth is
still explicitly deferred for Bill's joint session; provider availability, real
billing, native hearing and human accessibility remain unverified. No installations,
restarts of Bill's app, commits, external account changes or prior endpoints touched.


## Chapter 16 — GUI catalog presentation verified locally

Chapter 16 remains open. Catalog library/CLI evidence below carries forward.
Funding announcements and persisted selection are next, then contract reconciliation,
grading/prose and endpoint. No live OAuth or account catalog is claimed.

- Guarded POST /auth/models uses the existing account endpoint and Credentials.
  Browser supplies only a label; host OpenAI Configuration owns the destination.
  Existing origin/header, strict body/size, lifetime and safe error guards apply.
  No running Agent model/billing mutation and no catalog cache or new owner.
- Account panel shows a snapshot in server order via textContent, never HTML.
  Label edits/account actions clear it; responses for a changed label are discarded.
  Empty lists and failures have distinct safe notices. Long names wrap at 360px.
- catalog-gui-build-2.log: clean build, zero warnings/errors, exit zero.
- catalog-gui-exercise.log: passed:true, exit zero. Evidence at
  /private/tmp/ch16-gui-catalog.9sHIEJ/result.json. Actual GUI HTTP, two isolated
  browser clients, four catalog GETs, zero inference requests, clean host exit.
  Verified guards before GET; ordered visible entries and hostile-looking text;
  360px geometry with long slug; second-account empty result; unchanged credentials
  from discovery; held-response/label-change isolation; denial clearing; signed-out
  no-request/no-key-fallback; unchanged running billing/account; no JS errors.
  DOM/geometry and fake-provider evidence only, not vision, audible/human access,
  successful OAuth, live account availability or a model picker.
- First executable run passed. All owned jobs collected. Four GUI source files
  plus README changed; scoped whitespace scan passed. Auth18, Responses8 and CLI
  five-scenario receipts below remain valid; no broad campaign was repeated.

GUI settings remain deliberately presentation-only. Shared funding persistence
must not become a second GUI-specific authority; no such design is implemented yet.
A record-edit anchor refused before writing; corrected from the actual README.

## Chapter 16 — account model discovery verified locally

Chapter 16 remains open; no grade/endpoint or live OAuth. Funding announcement
and persisted billing/account selection remain next; GUI catalog presentation is
not yet implemented. Earlier reasoning, summary, host and transport receipts carry
forward for unchanged paths.

- Credentials.ModelsAsync returns immutable public PlanModel values, server order,
  visibility:list only. Fresh disk token and HTTP GET each time; no catalog cache,
  model-capability promotion, price assignment, inference, History or meter effect.
- Reused the existing owner/gate/lifetime and extracted its existing token resolver.
  Renewal still serializes across processes and publishes before token use. Catalog
  GET holds the local operation gate; destruction cancels/joins it. Shared endpoint
  validation now serves Configuration and discovery: HTTPS except literal loopback,
  no userinfo/query/fragment. Existing no-redirect/no-cookie handler remains in use.
- Bounded existing Auth JSON decoding, duplicate/ambiguous visible-entry refusal,
  safe HTTP errors without provider body, no key fallback or credential erasure.
  Account status and tokens retain one authoritative protected store.
- CLI auth models uses that owner without an Agent or conversation; account flags
  retain precedence. Existing LLM_BASE_URL/OpenAI aliases choose service, version,
  or complete Responses base. README explains the command and evidence limits.
- Primary source fetched live with HTTP 200:
  https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference
  confirms GET /v1/models with OAuth bearer, models/slug/display_name/visibility,
  visibility:list filtering and server ordering. This is documentation evidence,
  not a real account discovery or account-specific availability claim.

Verification in /private/tmp/ch16-csharp.WMIdlt:
- catalog-auth-tests.log: initial 17/0. Final catalog-auth-final.log: 18/0, exit zero,
  including existing RSA/PKCE/two-process rotation and four discovery groups:
  ordered filtering, account isolation and fresh tokens; invalid JSON/entries,
  HTTP/redirect/endpoint refusals; cancellation/destruction; real signed renewal
  before GET, with its rotated token durably reusable by inference.
- catalog-cli-build.log and catalog-gui-build.log: clean, zero warnings/errors.
  catalog-build.log: clean library build.
- catalog-cli-exercise.log: exit zero, passed:true, five actual CLI scenarios and
  four GETs; /private/tmp/ch16-catalog-cli.s83JtL/result.json. Ordered visible names,
  second-account empty list, denial, redirect refusal, signed-out no-request. Dummy
  ambient key never used; store unchanged by discovery; no conversation/log files.
- catalog-responses-tests.log: eight focused groups pass, exit zero, covering the
  shared inference endpoint policy and plan path. Unchanged full121 evidence is
  carried forward, not relabeled as a new full-suite run.
- Scoped whitespace scan passed eight production/test files; README is subsequent
  documentation only. Free space observed 13,517,796 KiB. All jobs collected.

No test/build failure occurred in this slice. I did make failed source-read guesses
for old fixture/protocol filenames, and several combined outputs were truncated;
these are not complete source review. Corrected extracted-helper indentation and
value comments before testing, and corrected the test's LookupModel signature
before compilation. No live OAuth, actual provider catalog, GUI catalog check,
funding selection, installation, restart, commit, or endpoint is claimed.

## Chapter 16 — selected-model reasoning and explanation policy verified locally

Chapter 16 remains open; no grade/endpoint or live OAuth. Model catalog discovery
and funding/selection contracts remain next. Earlier summary/auth/transport
receipts below carry forward for unchanged paths.

- Extended the existing Common ModelFeatures value and Engine.LookupModel, not a
  second catalog. Exact gpt-6-astra row documents Responses-only function tools,
  streaming, efforts low/medium/high/xhigh/max and summary auto. Provider image
  capability does not authorize unimplemented Responses media. Unknown names,
  suffixes and other models do not inherit reasoning capability or pricing.
- Immutable Configuration validates the chosen surface and options once. For this
  row, application defaults are medium/auto; explicit off omits summary. Unsupported
  options refuse before HTTP. Responses emits one reasoning object; saved config
  includes effective effort/summary/explanation policy as documentary data only.
- Selected-model RequireToolExplanation OR explicit opt-in uses the existing actor
  admission gate. Ordinary text must precede tools; reasoning summaries do not
  qualify. Refusals are paired tool results, not failed turns. SSE arrival order
  survives terminal reordering. This is Ensemble policy, not a provider guarantee.
- Both hosts accept --reasoning-effort/--reasoning-summary, overriding
  OPENAI_REASONING_EFFORT/OPENAI_REASONING_SUMMARY. Unknown-model requests omit
  reasoning and explicit unsupported combinations refuse. No billing fallback,
  model-price alias, browser account mutation, or new persistent owner was added.
- Primary docs returned HTTP 200: https://developers.openai.com/api/docs/models/gpt-6-astra
  (targeted quote) and https://developers.openai.com/api/docs/guides/reasoning
  (full result; inline truncation followed by bounded saved-text extraction).
  The guide explicitly recommends auto summaries, states GPT-6 Astra requires
  Responses for function tools, and documents effort usage. o3's page was also
  read but marked deprecated, so no o3 row/default was introduced. These are
  documentation observations, not live availability or successful inference.

Verification in /private/tmp/ch16-csharp.WMIdlt:
- reasoning-tests-3.log: eight focused groups pass, exit zero. Two new groups check
  exact lookup, defaults/overrides/off/refusals, unknown omission, independent
  Agents, saved policy, and real actor/HTTP/tool effects. Six dispatch cases cover
  JSON/SSE × summary-only/late-text/explanation-first; paired refusals and successful
  calls reach the next request. No unsupported pricing was invented.
- reasoning-cli-build.log and reasoning-gui-build.log: zero warnings/errors, exit
  zero. reasoning-host-exercise.log: exit zero, passed:true; actual CLI/GUI exercise
  at /private/tmp/ch16-reasoning-hosts.w1NnwP/result.json, ten loopback requests,
  ten clean success exits plus twelve nonzero startup refusals with no request.
  Defaults, environment, flag precedence over invalid environment, off, unknown
  omission and wrong surface/value/missing-value refusals were checked. GUI uses
  actual WebSocket turns and joined SIGTERM, not browser rendering in this slice.
- reasoning-regression.log: 121 passed, zero failed, owning chain exit zero. No
  unchanged prior-chapter browser/grade campaigns were rerun. Later edits only
  clarified comments and documentation. Available space observed 13,571,268 KiB.

Failures/corrections: multiple prevalidated patch attempts refused anchors and
saved nothing; a patch-generator quoting failure also saved nothing. The first
shell omitted set-e and built unchanged code despite a refused patch (reasoning-
build.log exit zero is not implementation evidence). Corrected library build
reasoning-build-2.log exited zero. reasoning-tests.log failed compilation on my
invented Context.Empty/Config()/Throws helpers; fixed to actual declarations.
reasoning-tests-2.log had six passes/two failures: tool policy fixture installed no
tool, and dispatch fixture wrongly expected turn failure instead of paired refusal
plus a follow-up request (eventual cancellation). Both were fixture assumptions,
not reasons to change the actor contract. Corrected tests exercise real outcomes.

Next: review and implement remaining model-catalog and funding/selection contracts,
then Chapter16 executable grading and C# prose/review. Real OAuth stays deferred
for Bill's joint session; no successful signed CLI/GUI OAuth integration is claimed.

## Chapter 16 — public summaries and auth-refusal narration verified locally

Chapter 16 remains open; no grade/endpoint or live OAuth. This slice completes
returned-summary display/speech and the missing durable auth-error path, not
model-selected reasoning request options or catalog discovery.

- GUI maps ReasoningSummary chunks explicitly, rather than treating them as tool
  calls. PartFrame projects only public summary_text content into a distinct
  reasoning_summary card, preserving original signed/encrypted OpaquePart data
  in History for compatible provider replay. Opaque display frames no longer
  contain native replay payloads. Cards are collapsed/labeled separately from
  answers; the existing speech cursor/formatter handles incremental admission,
  final-only suffixes and silent replay with a “Reasoning summary” label.
- Common declares PlanAuthorizationException (InvalidOperationException subtype),
  Engine implements it and converts expected credential-acquisition failures to
  one fixed token-free message. No private store paths/issuer error text are sent
  to History. Cancellation and disposed-owner faults are not converted; rendering
  and bookkeeping are outside that conversion. The actor recognizes only this
  typed refusal and records ErrorOccurred before failed TurnEnded. No duplicate
  GUI turn-error narrator, no API-key fallback, and no usage increment. Append
  failure still follows ordinary actor fault handling rather than publishing a
  fabricated error. Direct Engine refusal remains a thrown exception.
- Scratch `/private/tmp/ch16-csharp.WMIdlt`: `summary-auth-tests.log` focused
  Responses6/0; `summary-auth-regression.log` full119/0; both actual exits zero.
  Tests include signed-out Agent failure with exactly one error and one failed
  turn/no ResponseEnded, unchanged meter, and pre-cancelled Engine behavior.
  `summary-auth-gui-build.log` and `summary-auth-cli-build.log` clean, zero exits.
  Seven changed source files passed scoped whitespace review. Final source review
  changed only an explanatory comment after those checks.
- `exercise-gui-summary.mjs`, `summary-gui-exercise-2.log`: exit zero/ passed:true;
  `/private/tmp/ch16-gui-summary.nvozVd/result.json`. Four actual loopback requests,
  real isolated browser, one clean host exit. Held SSE terminal until summary
  speech was admitted; final suffix once; terminal-only and plain JSON summaries;
  separate collapsed cards/answer; encrypted sentinel absent from DOM and socket
  projection yet retained in next provider request; failed stream had no durable
  summary and one error narration; refresh replay silent. Uses transcript speech,
  not acoustic/native hearing. Unknown fixture model used only to return summaries;
  this is not reasoning-capability evidence. GUI path unchanged by later auth edit.
- `exercise-gui-auth-narration.mjs`, final `auth-narration-exercise-4.log`: exit
  zero/passed:true; `/private/tmp/ch16-auth-narration.AcsBIT/result.json`. Real page,
  explicit plan mode with empty private store plus dummy ambient API key. One
  durable lifecycle/error_occurred, failed turn, full error content narrated once
  across labeled sentence utterances, zero provider requests, silent replay,
  clean host exit. No OAuth endpoint or real provider contacted.
- Failed fixtures retained: first summary launch supplied unsupported --stream
  (GUI streams by default), refused before requests. Auth narration run1 expected
  error_occurred as frame type instead of lifecycle kind; run2 assumed one error
  equals one utterance; run3 omitted the formatter's per-sentence Error prefix.
  Final assertion requires each label and exact concatenated content once. No
  product behavior was weakened for those fixture corrections. I also mistyped
  a summary receipt path while recording; final path above came from its log.

Next: selected-model reasoning options/explanation policy, remaining model-catalog
and funding/selection contracts, then scoped grading and teaching prose. Current
model names in historical prose are not verified capability facts. Live OAuth
remains explicitly deferred for Bill; native/audible/human accessibility and
successful signed host login remain incomplete. Prior Auth/login/startup and
GUI-management receipts carry for unchanged paths; no old chapter campaigns rerun.

## Chapter 16 — GUI account lifecycle verified locally

Chapter 16 remains open: no grade/endpoint or live OAuth. GUI now offers explicit
account status/login/cancel/local logout/signout controls. Account management uses
one Ensemble-owned Credentials instance even in API mode when credentials-dir is
explicit. It never changes the running account/billing selection or searches for
credentials. Existing Auth14, Responses6/full119 and startup/CLI receipts carry
forward for unchanged library/CLI behavior.

- Common/Protocol declares MapAuth; Server/Auth registers host-scoped HTTP routes
  through the existing Agent → Ensemble → Credentials path. No new auth owner,
  token cache, model events or WS messages. Exact local host/port, custom header
  and same-origin POST guards; no CORS grant. Bounded (4096-byte) strict JSON,
  fixed error text, no-store/no-referrer/CSP. Callback alone is exempt from Origin
  and header requirements; existing Auth validates URI/state and owns one-use
  consumption. Requests link to host shutdown; ending the host cancels its
  pending attempt. Status excludes tokens, client IDs and subject protocol IDs.
- wwwroot/auth.js/index/style implement redacted status, explicit refresh and
  local/remote clearing labels. A user-opened separate window receives the URL;
  it never enters main-page DOM, links, storage, History or traces. Blocked popup
  refuses before Begin. Failed delivery can leave a pending attempt, reported by
  refresh and explicitly cancelable; it does not cancel another client's attempt.
  Running billing/account are shown as unchanged by these controls.
- Scratch `/private/tmp/ch16-csharp.WMIdlt`: `auth-gui-build-3.log` clean build,
  exit zero. `exercise-gui-auth.mjs` final `auth-gui-exercise-6.log` exit zero,
  passed:true; `/private/tmp/ch16-gui-auth.QXLmHQ/result.json` records nine grouped
  scenarios, three clean host exits, closed callback ports and zero model requests.
  Covers strict guards/body refusals without mutation; duplicate Begin; wrong
  state continuing; matched denial/missing issued client consuming without token
  exchange; two actual browser clients; blocked popup; reauth ID-token hint kept
  out of both DOMs and logs; second-client cancel/logout; another account and
  registration preserved; token-free signout; 360px geometry; pending shutdown;
  fresh pending state plus durable logout after restart; no implicit store.
- Popup navigation is a small working browser-boundary fake, not real OAuth.
  No production authorization URL was fetched. Successful signed GUI exchange
  and remote revocation are not host end-to-end evidence; library issuer-fake
  evidence alone covers those paths. Native/audible/human accessibility and live
  OAuth remain incomplete. README documents these limits.
- Failures retained: initial build rejected nullable label assignment (corrected
  explicit null handling). Exercise1 expected null for actual empty selected
  account; exercises2/3 used fetch's ineffective Host override and expected the
  wrong guard status. Raw HTTP now verifies global hostile-Host 400; Origin and
  custom-header refusals remain 403. Exercise4 exposed a real handler error:
  InvalidDataException does not derive from IOException; explicit catch now
  converts rejected callbacks to fixed 400 replies instead of 500. Exercise5
  passed only after diagnostic foregrounding of owned headless tabs; not credited
  as unattended evidence. Final exercise6 explicitly foregrounds each operated
  tab and uses timer-based waits; passed without intervention. No Bill browser or
  device focus was used. All owned jobs collected; seven changed GUI files passed
  the scoped whitespace check.

Next: model-selected reasoning options and summary display/speech, including
one durable auth-refusal narration path (currently only turn_ended.error). Review
remaining model-catalog/funding announcement and historical auto-selection claims
against the chapter contract before grading/prose. Responses media remains explicit
unsupported. Live OAuth stays deferred for Bill's joint session.


## Chapter 16 — standalone CLI auth lifecycle verified locally

Chapter 16 remains open: no grade/endpoint or live OAuth. CLI now exposes
`auth status|login|logout|signout`, with existing credentials-dir/account options.
Dispatch precedes Configuration: auth creates no Agent, skills, conversation,
API trace or inference request. README documents these private host operations.

- Login binds a literal-loopback HttpListener before creating/showing the private
  authorization URL. A temporary TCP probe chooses a candidate; the real bind
  must succeed before authorization (no race-free port claim). URL is private
  terminal output only, potentially containing an ID-token hint. Callback text
  is fixed, with no-store/no-referrer/CSP headers. Wrong path/method refuse;
  wrong-state callbacks keep listening; consumed denial/missing-client failure
  exits. Ten-minute cancellation, Ctrl+C/SIGTERM and disposal close the listener
  and cancel pending login. Expected failures do not print raw exceptions.
- Common/AuthLogin declares `LoginPendingAsync`; Auth/Login implements the query
  under the existing gate against the sole Credentials-owned attempt. No host
  state/nonce copy. Existing login tests assert pending state on foreign callbacks,
  cancellation and denial. Local logout retains registration and other accounts;
  signout distinguishes local clearing from confirmed remote revocation.
- Scratch `/private/tmp/ch16-csharp.WMIdlt`: `auth-cli-build.log` and
  `auth-cli-gui-build.log` clean, exit zero; `auth-cli-components.log` has all
  14 Auth groups passing, exit zero (real RSA/issuer/refresh/multi-process tests).
- `exercise-cli-auth.mjs`, final `auth-cli-exercise-2.log`, exited zero with
  parsed passed:true. `/private/tmp/ch16-cli-auth.GP0BlP/result.json` records
  11 scenarios: status, local logout/account isolation, token-free signout
  (no issuer request), post-logout status, four argument refusals, denied login,
  missing-client refusal and SIGTERM cancellation. Each login first receives wrong
  path/method/state; only proper callbacks consume attempts. Ports close, failed
  new-login stores retain zero registrations, private error payloads are not
  printed, and no conversation/API-log files appear. Invalid model environment
  proves auth does not require provider Configuration. Dummy secrets only.
- Initial `auth-cli-exercise.log` failed at status: my seeded timestamp omitted
  fractional digits required by the store's round-trip format. Corrected fixture,
  not validation. Failure retained `/private/tmp/ch16-cli-auth.uUZJMO`; its later
  GUI gate did not run. A documentation script saved README before a guessed CLI
  help anchor refused; inspected saved state, then completed only missing edits.
- CLI signed exchange and remote revocation are **not** host end-to-end evidence:
  library fake-issuer tests alone cover success. No production authorization page
  fetched or real account changed. Final CLI help edit is nonbehavioral.

Next: GUI account controls/callback lifecycle using the existing Credentials
owner, same-origin mutation guards, and no auth URLs in WS traces, History or
snapshot-visible DOM. Then model-selected reasoning options, summary UI/speech,
and durable auth-refusal narration (currently turn_ended.error only). Carry prior
Responses/host-route evidence for unchanged paths. Live OAuth remains deferred
for Bill's joint session; no repeat of Chapter 15/14/13 campaigns.


## Chapter 16 — CLI/GUI startup route and account selection verified locally

Current boundary: **Chapter 16 remains open**, without a grade or endpoint.
Both composition roots now select the verified transport/billing implementation;
login/callback UI and account management remain next. No workers, installation,
real OAuth, provider spending, or installed-runtime restart was used.

- CLI and GUI accept `--surface chat-completions|responses`, `--billing api|chatgpt`,
  `--account LABEL`, and `--credentials-dir DIR`, with corresponding `EN_SURFACE`,
  `EN_BILLING`, `EN_ACCOUNT`, `EN_CREDENTIALS_DIR`. Flags override environment.
  Omission preserves provider defaults and API billing. Configuration remains
  authoritative for vendor/surface/billing validation; no automatic paid fallback.
- Plan startup constructs the optional Credentials owner on the existing Ensemble,
  before Agent construction. Engines use the existing real parent path. An ambient
  API key is deliberately excluded from plan Configuration, rather than making
  inherited keys prevent an explicit plan selection. CLI's key guard applies only
  to API mode. Empty/missing route-option values refuse before inference.
- The only behavioral source edits in this slice are `Cli/Program.cs` and
  `Examples/Gui/Program.cs`. Main README documents the options and now correctly
  says Chapter 16 work in progress. Immutable ch15 and earlier endpoints untouched.
- Both builds exited zero with zero warnings/errors:
  `/private/tmp/ch16-csharp.WMIdlt/route-cli-build.log` and `route-gui-build.log`.
  Actual executable fixture `exercise-host-routes.mjs`, final
  `host-routes-exercise-3.log`, exited zero. Evidence:
  `/private/tmp/ch16-host-routes.W48b6z/result.json`, requests and per-run host/frame
  receipts. Ten actual loopback HTTP requests: each host's default Chat Completions,
  Responses API via flag and environment, and Responses plan via flag and
  environment. It checks path, bearer choice, exact lens bytes, API versus unpriced
  plan meter, flag precedence, no account/token bytes in inspected diagnostic logs,
  two signed-out refusals with zero provider requests despite ambient API keys,
  fourteen invalid startup combinations, and joined owned-host exits. These were
  socket/CLI exercises, not browser clicks or live authentication. Dummy protected
  credential files seed unexpired grants; no token endpoint is contacted.
- Failed receipts remain: `host-routes-exercise.log` at
  `/private/tmp/ch16-host-routes.91oooW` expected “signed/login” instead of the real
  “requires sign-in”; `host-routes-exercise-2.log` at
  `/private/tmp/ch16-host-routes.SMdUYd` expected a parser error event for an auth
  exception. Corrected fixture asserts the actual `turn_ended.error`. The library
  intentionally throws on credential refusal. **Remaining narration gap:** this
  path has no durable `ErrorOccurred`; address it with the Chapter 16 error/summary
  presentation work, without reinstating duplicate GUI error narration.
- Carry the prior 14 Auth groups and Responses six focused/full119 results only for
  unchanged library behavior. No broad suite/grader/browser reruns in this slice.
  Outstanding documentation job2524 was collected: HTTP200 but `query_failed`, no
  usable content. It is not protocol evidence; earlier successful primary sources
  remain the transport references.

Next: host-owned login/callback lifecycle and safe status/sign-out controls, then
selected-model reasoning options and incremental/final summary display/speech.
Live OAuth remains explicitly deferred for Bill's joint session. Responses media
still refuses explicitly. Prose, executable grading, public example and Chapter16
endpoint remain ahead.



## Chapter 16 in progress — Responses transport and explicit billing verified locally

Supersedes the transport next step below. **No Chapter 16 grade/endpoint, host
login controls/callback integration, or live OAuth result yet.** Bill's live test
remains deferred to our joint session. Work is personal, without workers.

- Configuration reuses the existing `Surface` vocabulary (Responses appended,
  existing numeric values preserved). API-key / ChatGPT-plan billing is a separate
  explicit fact; defaults preserve old routes. Plan requires streaming Responses,
  an account label, no API key, HTTPS (literal loopback HTTP is the test exception),
  and no URL userinfo/query/fragment. There is no payment fallback. The Engine's
  own HttpClient disables redirects/cookies for plan transport only.
- Engine resolves bearer authority through Engine→Agent→Ensemble→Credentials just
  before sending. Rendered bodies, saved configuration, lens files and History
  contain no account label/token. Missing store, signed-out account or denied plan
  permission cannot fall back to API billing. Saved surface/billing are documentary
  runtime policy, not restored credentials. UsageIdentity now includes billing;
  valid plan token counts stay visible with unknown dollar estimates, never API
  rates. Existing API rows/pricing and Engine-owned metering remain authoritative.
- `Engine/Responses.cs`: developer instructions, input messages, flat API functions
  or plan function namespace `ensemble`, correlated function outputs, exact JSON
  argument numbers, stateless `store:false` and encrypted reasoning inclusion.
  Opaque reasoning/native calls replay only for exact surface/model provenance.
  Native call IDs survive; namespace is adjusted to the selected billing profile.
  No media projection on this surface yet: explicit refusal, not silent loss.
  Existing skill/manual/request-only instruction composition is reused.
- JSON and SSE share final item/usage validation. Only successful completed
  responses yield response facts; malformed/duplicate call IDs, unsupported items,
  non-object arguments, failed/incomplete status and explicit error refuse. Usage
  uses existing nullable arithmetic; malformed optional counts remain unknown.
- `Engine/ResponsesStream.cs` / Common stream scratch: retain only item-done
  replay material. Empty terminal output uses completed items only when none are
  unfinished; nonempty terminal output is authoritative, never concatenated with
  the fallback. EOF/[DONE]/failed/incomplete terminal is not success. Tools cannot
  dispatch before terminal success. Stable arrival IDs survive final projection;
  reasoning-summary deltas and final suffixes use their own ChunkKind, not answer
  text. Plain responses keep authoritative separate finals, no fabricated deltas.
  Completed visible message items count for explanation ordering even without
  deltas; summaries don't substitute for explanation and late text cannot excuse
  an early call. Selected-model reasoning options and GUI/speech integration remain
  next work, not established by this parser slice.
- Cache lens uses actual Responses prompt order (tools/instructions/input), keeps
  exact wire bytes, and separates options. No second meter, transcript or scheduler.

Evidence: `/private/tmp/ch16-csharp.WMIdlt`. Six focused Responses groups test real
projection/codec/provenance, JSON/SSE failure boundaries, completed-item fallback,
authoritative replacement, summaries/suffixes/IDs, exact numbers, actual HTTP route
and headers, exact lens bytes, actual metering, real protected-store plan bearer,
logout refusal and redirect refusal. A REAL Agent executes one tool and sends its
correlated result; failed terminal produces no effect. A held stream exposes a
complete tool item, waits for its request-ID-correlated observation, and proves no
effect until the later completed terminal. No timing-only assertion or internal
mock. Dummy tokens/loopback only, not provider integration.

`responses-tests-3.log` **6/0** plus `responses-regression.log` **119/0**, sequential
job **2503 exit zero**, cover the initial transport and prior 113 groups. Later
review added encrypted-content inclusion, native namespace correction, completed
message arrival ordering, frame tracing, explicit-error refusal and plan endpoint/
redirect protection. Final focused `responses-security-final.log` **6/0**, with
`responses-cli-security-final.log` / `responses-gui-security-final.log` **zero
warnings/errors**, the sequential command **exited zero**. The preceding final run
`responses-tests-final.log` was also 6/0 with clean hosts and exit zero.
Carry the unchanged 113 regression groups, rather than claiming a new full run
for the last Responses-only corrections. Fifteen touched source files passed
whitespace inspection; free space last **13,963,092 KiB**.

Failures retained: `responses-build.log` / `-2` compile failures from guessed part
members and nullable Usage; corrected to actual CallID/Args/Parts and existing
Count/Subtract contracts. `responses-tests.log` fixture compilation guessed meter
and saved-config signatures. `responses-tests-2.log` **4/2**: exact-instructions
assertion ignored existing tool policy and codec fixture used unnumbered event.
`responses-tests-4.log` **5/1** incorrectly expected transient callbacks from a
plain response; fixed to authoritative separate finals. Several prevalidated
anchor refusals saved nothing. These are my fixture/source-reading mistakes, not
infrastructure failures. Review found real protocol gaps despite earlier passes.

Primary HTTP-200 documentation (cached fetches; not live-account evidence):
- https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference
  — public `/v1/responses`, OAuth bearer, `store:false`, `stream:true`, terminal
  success, and account-specific model catalog (host catalog still pending).
- https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations
  — instructions/developer messages, stateless history, namespaced tools,
  unsupported parameters. Guessed `/inference` and `/limitations` URLs returned
  404 and are not evidence; actual links were recovered from the sign-in page.

Next: host auth/callback lifecycle and explicit route/account selection, current
model capability/reasoning options, summary display/narration, public example,
prose and scoped executable grader. No live login until Bill returns. Preserve
Chapter15 and do not rerun its unchanged browser/grader campaigns.



## Chapter 16 in progress — initial OAuth library flow verified locally

Supersedes the initial-login next step below. **No Chapter 16 grade/endpoint,
application login UI, Responses transport, or live OAuth result yet.** Bill's
live OAuth deferral remains in force. Implementation/testing is personal, with
no workers, account changes, installation or runtime restart.

- `Common/AuthLogin.cs` / `Auth/Login.cs`: one immutable protocol-attempt snapshot
  held by the existing Credentials owner. It owns no services or independent
  lifecycle; Credentials owns HTTP, cancellation, locking and destruction. Public
  Begin returns a local browser URL but never launches a browser; Complete accepts
  the whole callback; Cancel clears pending attempt state. URLs can include a
  private ID-token hint and must not enter History/logs. The attempt expires after
  ten minutes, with fresh random state/nonce/verifier and S256 challenge each time.
- Authorization uses the documented `dynamic_agent_client`, `agent_name_hint`,
  `ext_agent_host_id`, scopes including `resource.invoke`, and API resource.
  Callback `client_id` is mandatory for bootstrap; a saved ID is authoritative on
  reauthorization. Strict exact `http://127.0.0.1:PORT/auth/callback`, UTF-8 percent
  decoding, duplicate-parameter refusal and constant-time state comparison precede
  any registration write or exchange. Wrong-state traffic leaves the valid attempt
  usable; a matching callback consumes it once. Provider error descriptions are
  never copied into diagnostics. Reconsent is explicit (`prompt=consent`).
- Store now permits a null subject ONLY without tokens. An issued, token-free
  pending registration is atomically persisted before exchanging the one-use code.
  A failed exchange retains it across restart, avoiding another bootstrap. Initial
  grants require scope plus a fresh signed ID token and nonce; only verified identity
  binds the subject. Existing subject cannot change. The existing grant validator
  serves login and refresh; no second store/parser was added. Exchange is not
  retried automatically. Tokens publish only after validation and atomic save.
- An immutable pre-login registration snapshot detects intervening disk changes
  under the shared process lock. A stale attempt cannot overwrite another owner's
  logout/rotation. Separate labels/registrations remain separate even for equal
  emails. Identity-only success stays signed in but cannot supply plan inference.
- Review found and corrected TWO earlier-slice contract mismatches: new host IDs
  now use the documented `urn:uuid:<v4>` form rather than a bare GUID (existing
  saved IDs are preserved, never silently regenerated); proactive refresh now uses
  the chapter's five-minute margin rather than one minute. The real two-process
  rotation test now begins with three minutes still remaining, proving proactive
  rotation and serialization, not merely expired-token recovery.

Evidence: `/private/tmp/ch16-csharp.WMIdlt`. `auth-login-build.log` clean library
build. `auth-login-tests.log` failed compilation because I guessed a fixture
Claims overload; `auth-login-tests-2.log` was 13/1 because I incorrectly called
Parent() after destruction (it deliberately throws). Corrected fixture uses real
current-time signed claims/access hash and checks internal reverse-link/root
removal. `auth-login-tests-3.log` was 14/0 with clean hosts (job 2438). After the
host-ID/refresh-window corrections, final `auth-login-final.log`: **14/0**, with
`auth-login-cli-final.log` and `auth-login-gui-final.log` **zero warnings/errors**;
sequential job **2446 exited zero**. Several prevalidated anchor attempts refused
before any writes/tests; they are not saved changes or passing runs. Nine-file
whitespace check passed. Last free space 14,025,660 KiB.

The four new groups use a functioning local authorization endpoint, one-use code
exchange with actual PKCE verification, signed RSA/JWKS, and an owned real loopback
callback socket with HttpClient browser navigation. They check durable pending
state at the server before exchange, restart/reuse, missing/substituted client ID,
wrong endpoint/state/duplicate or malformed parameters, cancellation/denial, failed
exchange without replay, missing ID/wrong nonce/foreign subject, token-free pending
validation, plan permission, same-email separation, stale logout, and joined
HTTP cancellation/destruction. This is not native browser or provider proof.
Unchanged 99 component-group evidence and prior chapter campaigns carry forward;
no browser/grader campaigns were repeated.

Protocol-source correction: fresh HTTP-200 pages
<https://developers.openai.com/siwc/token-sharing-open-source/sign-in> and
<https://developers.openai.com/siwc/token-sharing-open-source/vm-remote-server>
confirm the above names and loopback flow. The prior handoff's `client_name`,
`host_id`, callback `issued_client_id`, and `/complete` redirect were stale; the
current manuscript already uses `ext_agent_host_id` and callback `client_id`.
The current VM page describes local OAuth then secure credential transfer, NOT
that fixed headless redirect. No unsupported headless route or transfer mechanism
was added. These are documentation checks, not live OAuth.

Next: inspect current Engine/configuration/render/parser/stream seams and Chapter16
executable grader contracts; implement explicit Responses transport independently
of API-key versus plan billing, then host auth controls/owned callback listener.
Responses terminal success must gate retained output/effects, with completed SSE
items retained when terminal output is empty. Reasoning summaries and selected-model
explanation policy, local executable integration, public example, prose and grade
remain. Live OAuth stays explicitly deferred for Bill's joint session.



## Chapter 16 in progress — signed identity and renewable-session slice verified

This supersedes the next-step line in the storage-only entry below. No Chapter 16
endpoint/grade, initial OAuth login, host auth controls or real-provider result.
Bill's live OAuth deferral remains in force. No workers or account changes.

- `Common/AuthProtocol.cs` / `Auth/Identity.cs`: narrow RS256 signed ID-token
  validation using .NET RSA/SHA-256, not decode-only trust or a new dependency.
  Rejects ambiguous JSON/key selection, unsupported algorithms/extensions, bad
  signatures, issuer/audience/azp mismatches, invalid exp/iat/nbf, wrong login
  nonce or changed refresh subject, and invalid supplied access-token hash.
  Five-second clock tolerance is explicit. Keys come from the trusted issuer,
  never JWT-supplied key URLs. Returned identity is metadata only; safe errors
  do not retain parser/crypto exception text. This is not general JWT support.
- Credentials now owns a nonredirecting, cookie-free HTTP client and cancellation
  lifetime. Production fixes the OpenAI issuer; an internal loopback constructor
  supports real protocol tests. Destruction cancels lock/network waits, joins the
  operation gate, disposes transport, then detaches. No token cache was added.
- `Auth/Renewal.cs`: `GetBearerTokenAsync` holds the same cross-process lock from
  reread through HTTP refresh and durable rotation. Fresh tokens avoid HTTP;
  new ID tokens validate against the saved issuer/client/subject and fetched
  JWKS. Omitted refresh ID token retains the previously established identity.
  Malformed/identity failures and failed publication return no bearer token.
  Temporary network/429/5xx has bounded backoff; exhausted temporary failures
  preserve disk. The documented unusable-token codes clear all local tokens
  while retaining registration metadata. `invalid_client` is not revocation.
- Public status separates `SignedIn` from `PlanEnabled`. Missing direct-plan
  permission retains a valid sign-in but refuses inference; no API-key fallback.
  `SignOutAsync` commits local token removal before discovery/revocation, verifies
  the discovery issuer and same-origin revocation endpoint, and returns false
  when server revocation is unconfirmed. Empty HTTP 200 confirms revocation.
  The earlier local-only `LogoutAsync` remains an explicit disk-only operation.
- Real signed-token tests plus a loopback issuer exercise strict forms, actual
  RSA/JWKS validation, one-use rotating tokens across TWO real processes, all
  documented terminal refresh codes, bounded transient retries, redirect refusal,
  unchanged files on invalid responses, identity-only permission refusal,
  failed-final-write nonpublication/temp cleanup, local-first failed/successful
  revocation, and joined cancellation/destruction without partial publication.

Evidence remains `/private/tmp/ch16-csharp.WMIdlt`. `auth-identity-tests.log`:
3/0, exit zero. `auth-renewal-build.log`: clean library build. An edit refused
before saving because I guessed the test insertion anchor; missing shell fail-fast
then accidentally ran the broader unchanged harness: `auth-renewal-tests.log`
105/0, actual exit zero (job 2405), but this DID NOT invoke the new renewal groups.
After exact-anchor correction, `auth-renewal-focused.log`: 9/0, exit zero (2407).
A later first-anchor refusal saved nothing and correctly stopped; final patch
prevalidated all files before writing. `auth-renewal-final.log`: **10/0, exit zero**
(2410), including the final permission and publication checks. Unchanged 99 prior
component groups carry from the broader run; no browser/grader campaigns repeated.
`auth-renewal-cli-build.log` and `auth-renewal-gui-build.log`: zero warnings/errors,
actual sequential exit zero. Eight-file whitespace check passed. Last free space:
14,079,280 KiB. Test RSA keys/tokens are fixtures, not real accounts.

Documentation correction: the previous handoff incorrectly said to omit resource
at exchange and omitted resource.invoke. The manuscript was already correct;
fresh HTTP-200 sign-in documentation confirms `resource.invoke` in the requested
scope and `resource=https://api.openai.com/v1` on authorization/code exchange;
the session page confirms resource on refresh and no repeat scope. Sources:
- https://developers.openai.com/siwc/token-sharing-open-source/sign-in (live fetch)
- https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions (live fetch)
- https://developers.openai.com/siwc/website (HTTP 200 cached; exact validation section inspected from full saved response)
- https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery (HTTP 200 cached)
These are protocol documentation, not successful integration or account evidence.

Next: initial OAuth/PKCE and callback/headless flow on this same owner, including
saving the issued client before exchange. The current storage shape still requires
a known subject; deliberately extend it to a token-free pending registration, not
a second store. Then explicit billing/transport choice, Responses, reasoning
summaries, host integration, consumer, prose and scoped grader. Live login stays
for the joint session; do not credit this refresh fixture as login completion.

## Chapter 16 in progress — protected credential-store slice verified

Read the current and preserved Chapter 16 manuscripts in bounded sections and
inspected current root/Agent/Engine ownership and configuration. This is not a
Chapter 16 endpoint, grade, OAuth login, or provider-integration result. `main`
has advanced; the 165-file `ch15` endpoint remains unchanged. Bill explicitly
deferred live OAuth until our joint session when he returns home from work.

- New `Common/Auth.cs` and `Auth/Credentials.cs`: optional Ensemble-owned
  `Credentials`, bidirectional registration/removal, explicit destruction, and
  private value-like storage records. Root destruction closes Agents first and
  then detaches the store; application exit does not log the user out.
- One protected file is authoritative, not an in-memory token cache. Transactions
  reread under an owner semaphore plus an OS-held `FileShare.None` lock, stage a
  complete file, flush a newly created private temporary, and atomically replace
  the primary. No directory-fsync/power-loss guarantee is claimed. A stale lock
  file is harmless when no process holds its OS lock.
- Dedicated Unix directory/file modes are 0700/0600 from creation. Existing
  excessive permissions and symlink final paths refuse rather than silently
  chmod/read them. This slice explicitly refuses Windows pending an appropriate
  protected-store implementation; it is not an OS sandbox or same-user tampering
  guarantee. Strict schema/version/duplicates and token-completeness checks reject
  corrupt data without printing secrets or overwriting the damaged file.
- Stable host ID, account labels and issuer/subject/issued-client binding survive
  restart/logout. Equal emails do not merge accounts. Rebinding an issued client
  or label refuses. Public status carries account metadata, never tokens. Local
  logout removes all three token kinds while retaining registration metadata;
  provider revocation is not yet implemented and is not implied.
- `Tests/AuthStoreTests.cs` exercises real files, ownership/destruction, safe
  metadata, independent registrations, refusal/nonmutation, local logout/reopen,
  permissions/symlinks, cancellation while a lock is held, and two real child
  processes contending then publishing without losing either account.

Evidence: `/private/tmp/ch16-csharp.WMIdlt`, text pointer
`/private/tmp/ch16-csharp-current`; its `offline` wrapper uses approved private
.NET 10, isolated HOME/TMP, explicit PATH and existing caches. Initial full suite
`auth-store-tests.log`: **101 passed, 1 failed**, actual exit 1. The product failure
was lock contention on macOS: .NET returned errno 35 (`0x23`), not the assumed
Linux 11/normalized sharing violation. `auth-store-probe-2.log` retains the actual
code. Platform-specific retry now handles only those contention codes; other I/O
errors remain failures. Final focused `auth-store-tests-final.log`: **3 passed,
0 failed**, actual exit 0 (job 2374); unchanged 99 prior groups carry from the full
run. `auth-cli-build.log` and `auth-gui-build.log` both have zero warnings/errors.
Source whitespace check passed. No unchanged browser/grader campaigns repeated.

Other mistakes: first build omitted Auth from the project's explicit module list;
then an edit using the wrong list layout refused without saving. Corrected the
actual combined list. First focused probe failed compilation because I used
`.Count` on the suite's integer failure counter; corrected before diagnosing the
runtime failure. `rg` was unavailable; the separate whitespace check still ran,
and `grep` recovered source URLs. These are not credited as successful probes.

Primary reference pages fetched with HTTP 200 (cached documentation, not live
OAuth evidence):
- https://developers.openai.com/siwc/token-sharing-open-source/sign-in
- https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions

Next: implement shared Auth OAuth/PKCE and signed ID-token validation with working
loopback issuer fakes, holding the same store lock across refresh/rotation. Then
integrate explicit plan/API-key routing, Responses transport and reasoning-summary
contracts, both hosts, separate consumer/exercises, scoped graders and teaching
prose. Inspect remaining actual grader contracts before adaptation. No Chapter 16
snapshot or completion claim yet. Last free space: 14,132,560 KiB. No live accounts,
credentials, browser/device settings, installation/restart, commit/push, or shared
cleanup were touched.


## Chapter 15 completed — preserved offline endpoint; Chapter 16 next

Personal direct implementation, application exercises and final code/prose review
are complete within the permitted offline scope. `main` and `ch15` match across
165 source/support files; byte comparison passed, excluding bin/obj/runtime data.
`ch01`–`ch14` are unchanged. Evidence: `/private/tmp/ch15-csharp.1giP8a/endpoint.json`.

- Carried unchanged core evidence: **99 passed, 0 failed** in
  `meter-pricing-tests-2.log`; both host builds have zero warnings/errors.
  No unchanged Chapter 14/13 campaigns were repeated.
- Carried personally executed CLI, plain/SSE real-browser and separate public
  consumer evidence from the integration entry below. GUI Anthropic marker
  exercise directory is `/private/tmp/ch15-gui-markers.ezXp6h` (now explicitly
  read from `gui-markers-exercise.log`). No live provider is credited.
- New scoped executable grade: **100/100**, `passed:true`, actual exit zero
  (job 2311); `grade.json`, `grade.stderr` under the scratch above. Existing Go
  route remains intact. `-ch 15 -executable-only -gui EXECUTABLE` sends root-ID
  WebSocket prompts, joins SIGTERM shutdown, reads current-session events and
  uses C# captures/nullability. Marker and price sessions are separate: exact
  OpenAI mini/default fixture rates `.15,0,.075,.6`, unpriced Anthropic markers,
  plus unknown-model/tier, restore-with-empty-meter and all-provider counts.
  Exact full wire is checked against actual HTTP bodies; no invented cache-write
  field/model alias or production rate was added to satisfy the historical grader.
- Adapter contract tests pass in `grader-tests-final.log`; legacy fixture tests
  are included. Missing GUI, incompatible jobs-only and malformed rates each
  return exit 2. Original `invoiceSnapshot` signature is retained for Chapter 18.
- `book/edition-3/chapter-15.md`, main README and GUI README now teach the actual
  C# contracts, exact `.body`/`.wire.json`/`.options.json` filenames, prediction,
  real ownership, final-adjacency marker selection, response identity/tier,
  null semantics and process restart. Historical spending is attributed to Bill;
  older OpenAI explicit-marker claims are not current API instructions. Teaching
  review retains controlled-experiment and formatter-fixture lessons, not stale
  Go APIs/price tables. Direct source whitespace scan passes.

Important failures/corrections: first adapter compile failed because Chapter 18
also calls invoiceSnapshot; preserved its legacy wrapper rather than changing the
other chapter. A post-grade script guessed `earned`; the real report uses
`score`/`max_score`, checked as 100/100 afterward. A prose script guessed a heading
and stopped before writing; its trailing wc masked that shell status. Corrected
anchor/fail-fast rerun saved the prose. Review caught and corrected a guessed
`.prompt.json` filename to actual `.body`. These were my fixture/edit mistakes,
not application failures. No routine worker/reviewer approval is claimed.

Architecture review confirmed Engine↔CacheLens construction/removal/destruction,
immutable per-send predictions, counts grouped by actual response model/tier,
Common-owned persistent state, direct parent logging and a GUI projection rather
than duplicate accounting. Captures remain private best-effort diagnostics; token
prices remain estimates, unknowns stay unknown, and restore never seeds spend.

Live API testing is authorized but still blocked on an approved existing credential
source; no account-store search or running-Ensemble inspection occurred. Actual
provider cache grants, availability/billing, native/audible speech and human
accessibility remain incomplete. No install/restart/deploy/commit/push or shared
cleanup. Last observed free space: 14,195,296 KiB.

Next: read current and preserved Chapter 16, inspect actual contracts, and continue
personally from this endpoint. No need to rerun unchanged Chapter 15 evidence.

Governing method: [direct chapter execution](edition-3-csharp-procedure.md).
Bill authorized autonomous continuation, beginning with a small C# Chapter 1,
then Chapter 2 teaching polish, then Chapter 3 onward. Each completed chapter
ends with a micro-handoff. No student/reviewer/author workers are assigned.

## Current boundary

Chapters 1–14 have preserved offline-verified endpoints. Main is now in
Chapter 15, with cache-lens, opt-in markers, process meter, host/GUI integration
and a public consumer verified offline; `ch14` remains the immutable
152-file predecessor. No Chapter 15 endpoint or grade exists. Real-provider,
native/audible speech and human-accessibility work remain incomplete. Live API
checks are authorized, but no approved credential source is available here.
Do not repeat unchanged Chapter 14 campaigns.

## Chapter 15 — host/GUI integration and public consumer verified

Both hosts expose --cache-lens DIR / EN_CACHE_LENS_DIR, --cache-breakpoints /
EN_CACHE_BREAKPOINTS=1, and --prompt-cache-key KEY / EN_PROMPT_CACHE_KEY. Explicit
values override environment defaults. Wrong-provider routing keys remain refused
by Configuration, not silently ignored. CLI accepts its documented --no-stream
as well as the existing -no-stream alias. Lens files contain private prompts;
startup composition attaches the lens to Engine, which owns destruction.

The GUI declares MeterFrame in Common and implements it in Server/Meter.cs using
one immutable Engine snapshot. The usage frame has disjoint latest counts plus
all_input, nested session totals, total_cost_usd and hit_rate. Reconnect receives
the current snapshot; regular catch-up repairs lost notifications. Host-local
serialized-frame deduplication is only a notification optimization, never an
accounting authority. History replay does not seed usage. Browser formats COST,
ALL, CACHE and HIT RATE without counters or price tables. Unknowns display ?, not
zero; costs are labeled estimates. Session totals are available in a visible
expandable explanation and tooltip. Usage frames do not trigger speech narration.

Examples/Cache is a separate public-API assembly. It uses two real Agent turns
through a working loopback provider, verifies sent-byte/current-and-prior capture
identity, pure rendering, immutable meter snapshots, unknown-tier counts,
snapshot-only restoration with an empty new Engine meter, and explicit lens
relationship detachment. It is not a live-provider/cache-hit demonstration.

Evidence in /private/tmp/ch15-csharp.1giP8a:
- host-cli-build.log and host-gui-build.log: sequential exit 0, clean builds.
- cache-host-exercise-2.log: exit 0. Three real CLI runs, two requests each:
  Anthropic marker opt-in/default opt-out, OpenAI routing-key flag precedence,
  exact current/prior lens bytes, four missing-value refusals plus wrong-provider
  refusal. Owned evidence /private/tmp/ch15-cache-cli.183pkbaq.
- gui-meter-exercise-2.log: exit 0, six plain/SSE requests, two zero host exits.
  Actual page checks exact current/session counts and estimates by response
  model/tier, unpriced identity, missing usage, failed HTTP noncharging, two-client
  propagation, reconnect, saved-dialogue restart with zero meter, delivered SSE
  deltas and cache-option precedence. At 360px, no page overflow, all four labels
  fit, and expanded totals remain readable. No browser script errors. Evidence
  /private/tmp/ch15-meter-gui.XQZn9r/result.json and saved observations.
- gui-markers-exercise.log: exit 0, two real GUI WebSocket sessions, one each with
  flag/environment Anthropic marker activation, real wire capture equality,
  disjoint input/cache-write/cache-read usage and unknown pricing. Both owned
  hosts exited zero. Exact evidence path is in the log.
- public-consumer.log: separate Cache assembly run exited 0; two local HTTP calls.
  Build/fixture outputs stay outside the source endpoint.
- The existing 99/0 component suite remains valid for unchanged core code; this
  slice changed only host/GUI code and added the independent public example.

Failures preserved: staged source edits refused guessed Capture/host anchors
without saving; all remaining anchors were then checked together. The first
browser run exited 1 because the fake required stream:false instead of the actual
omitted plain-request field. The first CLI exercise exited 1 because its fixture
incorrectly supplied an OpenAI key to Anthropic. Both fixtures were corrected;
no product validation was weakened. Missing historical helper filenames were
retrieval mistakes, not test evidence. No live-provider/native/audible/human
accessibility completion is claimed.

Next: read and adapt the Chapter 15 executable grader, then C# teaching prose,
source review and source-only endpoint. Only ch15.go lines 1–180 have been read
in this continuation; it uses historical model identities/Go host conventions.
Do not invent price rows or weaken assertions to fit those fixtures. No Chapter
15 grade or endpoint yet; live API exercise still needs an approved credential
source. Do not rerun unchanged Chapter 14/13 campaigns.

## Chapter 15 — process-local meter and conservative pricing verified

Engine now owns a locked current-process map of immutable Usage values keyed by
response-reported vendor/model/service-tier identity, plus the latest completed
response. Common/Meter.cs declares the vocabulary and persistent state; Engine
implements metering and price derivation. There is no GUI counter, running dollar
balance, or restored-History seed. MeterSnapshot captures immutable rows, latest
counts and derived totals together. Usage.AllInput sums disjoint fresh/write/read
buckets safely; session hit rate is reads / all input, unknown for zero/unknown
input. Unknown output does not erase a known hit rate. Unknown operands propagate
only within their own categories. Concurrent direct sends count once at actual
completion; the ordinary Agent still serializes its requests.

Only SendAsync after a valid ResponseEnded records usage. Standalone Parse,
Render, replay, malformed/failed HTTP and cancellation do not record a completion.
This measures observed valid completions, not all possible charges on failed or
interrupted provider work. The shared plain/stream decoder exposes transient
billing identity on ParsedResponse without changing canonical events or snapshot
format. OpenAI streams retain final supplied service_tier alongside cumulative
usage and resolved model. Missing actual model/tier is unknown, not the selected
alias/default tier. Future auxiliary calls are not implemented in this chapter.

TokenPrices derives USD estimates from counts, using only exact verified OpenAI
standard-tier rows: gpt-4o-mini and its 2024-07-18 snapshot (input/read/output
$0.15/$0.075/$0.60 per million); gpt-4o and its current 2024-08-06 snapshot
($2.50/$1.25/$10). No prefix aliases or fake-model prices. Unverified nonzero
cache-write counts, missing counts, unknown model or non-default/missing tier
make the estimate unknown without erasing token counts. Empty session cost is
zero. These estimates exclude account discounts and are not an invoice or proof
of model availability. No Anthropic/Gemini price was invented for old model rows.

Primary fetches returned HTTP 200:
- https://developers.openai.com/api/docs/models/gpt-4o-mini
- https://developers.openai.com/api/docs/models/gpt-4o
  Full pages confirm the three token prices and exact alias/snapshot pairs.
- https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create
  confirms omitted request tier uses project-configured auto; response tier is the
  processing mode actually used, potentially different from requested tier.
- https://platform.claude.com/docs/en/about-claude/pricing does not supply a current
  Haiku 3 row in the retrieved table. Saved full markdown: scratch
  anthropic-pricing.md. No price inference from Haiku 3.5.
- https://ai.google.dev/gemini-api/docs/pricing targeted extraction answered for
  2.5 Flash instead of requested 2.0 Flash; it is not 2.0 pricing evidence.
The service-tier tool display was truncated; the relevant opening contract was
read completely. Full response is saved in session io/2209. No live API calls.

Evidence in existing /private/tmp/ch15-csharp.1giP8a:
- meter-build.log and meter-pricing-build.log: exit 0, clean core builds.
- meter-tests.log: 98/0, exit 0 for count-only slice (superseded API layout).
- meter-pricing-tests.log: exit 1 at fixture compilation; one Latest access missed
  its new Counts member after adding response identity. No behavioral credit.
- meter-pricing-tests-2.log: **99 passed / 0 failed**, actual exit 0. Four new
  groups cover disjoint arithmetic/overflow/unknowns; real concurrent loopback
  completion, failed/malformed/cancelled requests and restored-history isolation;
  real SSE final cumulative usage and final tier; exact published prices with
  multiple real decoded model/tier identities, missing identity, immutable prior
  snapshots and unpriced usage. Provider boundaries are working local fakes.
- meter-cli-build.log and meter-gui-build.log: sequential exit 0, zero warnings/
  errors; direct whitespace scan of ten changed files passed. Compilation only,
  not GUI meter behavior. Last free space 14,363,132 KiB, above safety floor.

This meter slice preceded the now-verified host/GUI and consumer integration
above. Grading, prose and endpoint remain unfinished; live testing still needs
an approved credential source.

## Chapter 15 — earlier cache lens and opt-in marker slices


Read the complete current Chapter 15 manuscript and all changes in the preserved
revision. Preserve its corrections: manuals remain SYSTEM; final tool-result
adjacency precedes cache-boundary selection; diagnose provider-ordered prompt
bytes separately from options and raw transport bytes. Historical provider/cache
measurements are not evidence for this implementation.

Implemented an optional Engine-owned `CacheLens`, with Common declarations and
Engine implementations. Constructors register both ends; explicit lens or Engine
destruction detaches both, releases prior capture bytes, and permits replacement.
No parallel history or accounting authority was added. `new CacheLens(engine,
ownedDirectory)` enables this library-only slice; no host switch yet. Render/replay
does not capture. Actual sends capture immutable per-request predictions, paired
`request[.prev].body`, `.wire.json`, `.options.json`, and current prediction JSON.
Prompt order is a local diagnostic approximation, not provider tokenization or a
cache guarantee. Captures contain private prompt content, not authorization headers;
keep them outside Git. They are best-effort debug files, not atomic recovery saves.

The UTF-8 scanner removes only cache-control properties at provider marker sites,
retaining raw escape spelling, whitespace outside removed spans, numeric lexemes,
property order and user arguments/schemas named `cache_control`. Prompt sections
are rearranged only in the diagnostic view. Actual request bytes and History are
unchanged. Response reports use their own prediction even when HTTP completes out
of order; unknown/failed/cancelled usage stays unknown. Disk capture failure logs a
safe diagnostic without blocking inference. Invalid caller-supplied JSON skips
capture rather than becoming a new transport admission rule; pre-cancellation
makes no capture. The marker and meter slices extend this foundation. GUI and host
configuration remain unfinished.

Evidence: `/private/tmp/ch15-csharp.1giP8a` (text pointer
`/private/tmp/ch15-csharp-current`), controlled copied offline wrapper. Library
`lens-build.log` exit 0. `lens-tests.log` and hardened `lens-tests-2.log`: **92
passed / 0 failed**, actual exits 0. Three new groups exercise byte fidelity and
provider section ordering, real loopback concurrent sends with reversed response
order, exact outgoing bytes, capture rotation, cancellation, disk failure,
invalid JSON pass-through, duplicate-registration refusal and complete ownership
cleanup. No real provider request, chapter grade, external consumer, or GUI
behavior is credited. Prior Chapter 14 receipts continue to cover unchanged paths.
Both host builds (`lens-cli-build.log`, `lens-gui-build.log`) completed in one
sequential exit-zero job with zero warnings/errors; compilation is not GUI usage.
A direct whitespace scan of the changed source files passed.

Primary documentation fetched with live-fetch requests, each HTTP 200:
- https://platform.claude.com/docs/en/build-with-claude/prompt-caching confirms
  tools → system → messages ordering and backward lookup of previously written
  prefixes. The targeted extraction omitted some requested breakpoint/block
  restrictions; finish those checks before coding markers.
- https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-search-tool
  explicitly says a deferred tool cannot carry `cache_control` (400); use a
  non-deferred tool. This supports the revision's last-non-deferred-tool rule.
- https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create
  documents optional `prompt_cache_key` for Chat Completions. Its extraction also
  describes newer model-specific options; those are not a mandate to implement
  unrequested capabilities or assume availability for existing model rows.
These are primary documentation, not live model exercises. Search excerpts were
not used as substitute for the fetched pages; one long extraction display was
truncated in an irrelevant newer-option section. Pricing remains unverified.

The source review made no edits to earlier endpoints. Initial guessed source paths
failed reads only; all production changes here compiled and the two suites passed.

### Opt-in marker slice

`Configuration.CacheBreakpoints` is an explicit Anthropic capability assertion,
off by default, consistent with the existing deferred-tool opt-in. Do not infer
support from a name prefix or silently extend the old exact-model capability row.
`PromptCacheKey` optionally supplies the documented OpenAI routing namespace;
wrong-vendor options refuse construction. No speculative OpenAI explicit markers,
new model aliases, or Gemini cache objects were added. Hosts do not yet expose
these options. Saved configuration records the options as documentary values,
never restored authority, and extracts the instruction text from block SYSTEM.

The disposable projection tags its handoff and ephemeral entries, then preserves
block identities through final Anthropic merging/result sorting. Select legal
blocks only before the first transient block. Possible markers are the last
non-deferred tool, nonempty SYSTEM, newest handoff and final stable content; shared
handoff/tail positions collapse. Empty text, thinking and unknown opaque blocks
are not explicit targets. All-deferred and empty-tool catalogs omit the tool
marker. Canonical Context, skill manuals, call arguments, and original events are
unchanged. New helpers live in `Engine/CacheMarkers.cs`; new configuration fields
and constructor declarations remain in Common.

Evidence: `markers-tests-3.log`, actual exit 0, **95 passed / 0 failed**. Three new
groups cover repeated call IDs and result adjacency, merged ephemera, at most four
markers, coincident boundaries, empty text/SYSTEM, thinking, ordinary opt-out,
marker-stripped content parity, canonical lexemes, real deferred-skill activation,
all-deferred frozen Context, instruction snapshots, OpenAI cache-key isolation,
and an actual loopback HTTP send with byte-for-byte body equality. `markers-build.log`
passed; sequential `markers-cli-build.log` and `markers-gui-build.log` passed with
zero warnings/errors. Direct whitespace scan of seven changed files passed.
No browser or live-provider result is credited by these tests.

Failures: the staged first patch refused before saving because I guessed renderer
shapes; the revised patch used actual occurrence-aware projection. Test registration
also refused a guessed label before saving. `markers-tests.log` exited 1 (93/2):
fixture used `instructions` instead of `system_prompt`, and `loadable` instead of
`loadable-skills`. `markers-tests-2.log` exited 1 (94/1): omitted primary instructions
select the nonempty default, not an empty string. Corrected fixtures in the third
run; do not relabel the failed runs. No product changes were required by those
fixture failures.

Provider source above was fetched in full with HTTP 200 after a targeted query's
extraction failed despite HTTP 200. Saved page: scratch `anthropic-prompt-caching.md`.
It confirms four explicit breakpoints, no empty-text/thinking markers, and default
five-minute ephemeral TTL. The documentation also lists currently active models;
that is not evidence that our older model row is available live. Minimum token
thresholds and actual cache grants remain provider concerns, not byte-count tests.
The earlier tool-search page establishes that deferred-tool markers are rejected.

Next after this earlier slice was process-local accounting (now verified above).
GUI meter and host switches, consumer, executable grading/exercises and prose remain.
Unknown models/prices must not gain invented aliases or rates. Live testing remains
pending an approved credential source, not a blocker to remaining offline work.

## Chapter 14 — completed at the permitted offline boundary

Implemented and reviewed: independent recovery journal, total known-event
reduction, newest-handoff/private-replay cleanup, immediate pending-skill
revocation, one target-derived tool ladder, full event archive, explicit opt-in
keep/stub, manual clear, and CLI/GUI settings and reduced-view projection. Common
retains declarations/state; Agent chooses cuts, History applies recorded facts,
and Engine renders. Pending manuals survive Chapter 14 cuts. The earlier
independent-16-KiB policy is superseded by T=400,000, T/8, T/16 and T/100.

Carried evidence for unchanged product paths: **89 component groups**, owned CLI
flags/keep/clear (four loopback requests), plain and SSE browser campaigns (ten
requests and two zero host exits each), earlier handoff/target/archive exercises,
and actual process-crash/journal recovery. Detailed receipts and prior failures
remain below. Native/audible and human accessibility are not established.

New closure evidence is in `/private/tmp/ch14-csharp.0m4UWS`:
- `consumer.log`: separate assembly using only public APIs, `passed:true`, zero
  network requests, exit 0. Covers policy nonmutation, host handoffs, paused clear,
  retained pending manual, zero-log save/restore, complete event-archive replay,
  sequence continuity and explicit detachment. Seeded exchanges are historical
  replay fixtures, not tool-execution evidence.
- `grade-final-2.json` plus empty `.stderr`: parsed **100/100, passed:true**, eight
  scoped executable checks; sequential job 2123 exited 0. The copied GUI in
  `graded-gui` is pinned by `gui-executable`. `grader-tests-final.log` passes the
  focused inline-parser and both-encoding handoff-order tests; three actual
  incompatible/missing scope cases exit 2 in `grader-scope-refusals-final.log`.
- `skills-consumer-final.log`: final review found the shipped Chapter 9 example
  still expected a write after unload. Corrected it to execute while active,
  then require refusal, unchanged file and retained manual. Separate example
  execution exited 0. No core change was needed, so the 89-group suite and host
  receipts were not repeated for this example/documentation correction.
- `endpoint.json`: `solutions/edition-3/ch14` has **152 source/support files**, all
  byte-verified against main; no bin/obj, captures, logs, profiles or credentials.
  Earlier endpoints were not touched. Final direct source/prose scan and scoped
  diff whitespace check were clean. Last free space: 14,471,520 KiB.

The new `Chapter14Executable` adapter keeps the default Go route and reuses its
applicable scenarios. It drives real C# HTTP/WebSocket controls, waits for durable
turn completion, joins owned SIGTERM saves and verifies an actual SIGKILL exit.
Restart fixtures explicitly select current startup manuals; they do not restore
capabilities from yesterday's configuration. The historical inline-tool check is
replaced only for C# by instruction/declaration checks, real MCP invocation and
post-unload refusal. Both handoff encodings require the handoff and sibling
results before the reduction event. A second clean save establishes a distinct
backup before the crash scenario. No synthetic production model rows were added.

Failures remain failures: staged patch drafts refused before saving because I
used guessed declarations; first grade `grade.json` exited 1 at **70/100** because
its adapter assumed restored manuals and a Go event label. `grade-2.json` then
passed 100/100/exit 0 after those fixture corrections. `grader-tests-4.log` selected
an unrelated historical Go integration test by an overly broad name filter and
failed for its absent Go endpoint; focused selection subsequently passed. The
initial third flag probe was an unknown flag, not a scope test; the final receipt
uses the actual incompatible `-jobs-only` flag. Later grades retained 100/100 after
strengthening ordering and SIGKILL assertions, with actual zero exits collected.

`book/edition-3/chapter-14.md` now teaches the C# implementation in readable steps,
including the historical rationale, public API, target semantics, durable ordering,
strict decoding versus total reduction, recovery/archive distinctions and scoped
verification. Main/GUI READMEs were updated for revocation, journaling, current
controls and queued CLI input. The procedure now records Bill's renewed live API
permission rather than the superseded blanket prohibition. No live calls,
installation, deployment, shared cleanup, commit or push occurred.

Next: read both Chapter 15 manuscripts and the current C# render/cache boundary
before implementing it. Preserve all unchanged recovery/handoff/host evidence.

## Chapter 14 in progress — separate public consumer verified

The external `consumer/Consumer.csproj` under `/private/tmp/ch14-csharp.0m4UWS`
references the main assembly, with no internal access or copied implementation.
`consumer.log` records `passed:true`, sequence boundary 19, zero network requests,
and process exit zero. It checks public target/band defaults, negative-update
nonmutation, zero-disable, real host-dispatched handoff replacement, paused manual
clear preserving dialogue/pending manuals/newest handoff, zero-log shutdown save,
full archive replay retaining original facts, snapshot restoration, sequence
continuity, request projection, and explicit owner detachment. Historical tool
exchanges are explicitly seeded replay fixtures, not claimed tool executions.
Its owned data directory is recorded in the log. Existing actor/Job and host
receipts cover unchanged live execution paths; they were not rerun.

At this earlier milestone, scoped grading, prose and endpoint were unfinished. Both existing
Chapter 14 grader files have now been read completely. Their Go driver assumes
stdin GUI prompts/EOF, historical inline tools, synthetic model-policy rows and
Go event encodings; those assumptions must not become C# product requirements.
A separate executable path will preserve applicable behavioral assertions while
using actual C# host controls and wire fields. The completed grade is recorded above.

## Chapter 14 in progress — manual clear and host integration verified locally

Bill explicitly renewed live API test permission: “Please feel free to test live
apis. Please continue.” Small bounded exercises may use existing authorized
credentials; no account/credential changes or deployment. No live call has been
made under this permission. The test environment has no LLM/provider key; I asked
for an approved existing credential-file/configuration path rather than searching
private account stores. Live evidence remains incomplete, not a blocker to the
remaining offline work. Other operational boundaries remain.

Manual clearing uses `InputKind.ClearTools`: admission is acknowledged through the
existing mailbox; one actor-owned flag waits for paired current-batch outcomes,
then records a fixed-range tool-level redaction. It works while paused, does not
kill independent Jobs, and preserves dialogue, pending manuals and the active
handoff. Real actor/Job/loopback tests cover held and interrupted batches,
idempotent empty clears and zero-log snapshot replay. `clear-tests.log` job 2008
exited zero, **89 passed, 0 failed**. `host-clear-tests.log` job 2023 also passed
89/0 and built CLI; that chain exited 1 at GUI compilation (my nonexistent
`EventKind.Summarized` and inconsistent settings variable). Corrected GUI build
`host-clear-gui-build-2.log`, job 2028, exited zero with no warnings/errors.

CLI adds `--context-budget-bytes N`/`EN_CONTEXT_BUDGET_BYTES` (nonnegative, zero
actually disables) and `--stub-tool-results`/`EN_STUB_TOOL_RESULTS=1` (explicit
opt-in), plus JSON `kind:"clear_tools"`. Explicit target overrides environment.
GUI supports the same stubbing opt-in; its context target comes from its existing
settings file and `context_target_bytes` update. Agent owns the live target and
one shared 400,000 default constant. GUI zero/negative means default, positive
values round to thousands; a wider intermediate prevents Int32 rounding overflow.
This host convention deliberately differs from library/CLI zero-as-disabled.
Target joins the existing settings save-before-publication transaction; it is not
copied into presentation Preferences. Failed saving changes neither Agent policy
nor acknowledged client settings.

The page adds a context-target input and Clear tool traffic button. Projection
renders HandoffPart explicitly and no longer recreates removed model tool cards
from old ToolStarted diagnostics. Recorded cuts or a missed trimmed tail cause
the existing silent replay envelope, rather than a second transcript store or
incremental stale cards. Completed-response as well as turn-ended frontiers retire
stored previews and reject late deltas: a mid-turn cut may remove the final card.

Owned executable/browser evidence (scripts/logs in `/private/tmp/ch14-csharp.0m4UWS`):
- `exercise-host-cli.py`, `host-cli-exercise-4.log`:
  `/private/tmp/ch14-host-cli.__acznfy/result.json` passed, four loopback requests,
  app exit 0, negative target exit 1 before HTTP. Real read/keep, admitted clear,
  next request with preserved dialogue/no tools, flag precedence, documentary
  saved target and stubbing selection. Fixture waits for durable `turn_ended`;
  the CLI's legacy `user` form queues input, contrary to my initial assumption.
- `exercise-host.mjs`, `host-browser-3.log`, job 2052 exit 0:
  `/private/tmp/ch14-host.FA265M/result.json`, ten requests, two zero host exits,
  no page errors. Actual two-client settings propagation, target normalization,
  blocked-directory save rollback visible on page, large read stubbing, keep,
  newest handoff replacement, paused clear, refresh, zero-log host restart and
  following request. Owned headless Chrome/profile and dummy loopback provider.
- `exercise-host-stream.mjs`, `host-stream-gui-build-2.log` and
  `host-stream-browser-2.log`, sequential job 2066 exit 0:
  `/private/tmp/ch14-host.HPBP8c/result.json`, same ten requests/two zero exits.
  Real SSE tool deltas observed, saved `stream-frames.json`, no stale tool previews
  in reduction replays. Carries the same browser/rollback/restart assertions.
  This verifies protocol/UI behavior, not native speech or human accessibility.

Failed attempts remain failures. First browser command could not find Node on
the controlled PATH; explicit existing `/opt/homebrew/bin/node` fixed setup.
Browser job 2035 exited 1: hidden observer tab delayed the first click (owned tab
was inspected/brought forward), then my wrong exact save-error expectation timed
out. Fresh fixture brings its owned headless page forward. The invalid-settings
probe now uses `patchSettings`, not the text-oriented `send` helper. CLI attempts
1–3 used an unsupported kind, assumed blocking input, then looked for the journal
reason at the wrong level; fourth run passed. A two-occurrence frontier anchor
refused before saving; its shell erroneously rebuilt unchanged source before
failing to find the SSE script. No credit for that build. Corrected fail-fast
job 2066 saved/compiled/exercised both frontier checks. Subsequent public comments
and CLI usage-label edits do not change behavior. Final `host-final-cli-build.log`
and `host-final-gui-build.log`, plus a direct whitespace scan of twelve changed
source files, completed in job 2071 with exit zero. Last free space: 14,653,108 KiB.

No Chapter 14 endpoint, grade or C# prose yet. Next: separate public consumer,
read/adapt the executable-scoped grader, teaching prose and endpoint. Preserve
unchanged recovery/SIGKILL/archive evidence; do not rerun Chapter 13 campaigns.

## Chapter 14 in progress — once-only keep/stub and corrected target policy

Corrected my earlier ladder configuration: Chapter 14 specifies one target byte
budget, not independent 16 KiB defaults. Agent now owns `contextByteBudget`,
default 400,000, with `ContextByteBudget`/`SetContextByteBudget`; the read-only
`ToolCompactionBudgets` derives T/8 and T/16. Positive targets have at least a
one-byte band, and zero disables both bands and immediate stubbing. The prior
`SetToolCompactionBudgets` API and defaults are superseded before a chapter endpoint.
Negative updates refuse without mutation. Existing selection/cut events remain
unchanged; the old-default executable receipts below are historical for that policy.

Selected-model result stubbing is explicit `Configuration.StubsToolResults`, off
by default, not guessed from a model prefix or enabled by a fabricated model row.
The real `keep_tool_results` control is installed only when opted in. Synchronous
and asynchronous skill startup share one universal-control name helper; catalogs
and all three provider declarations agree. Saved configuration documents the
current flag and Agent target, but startup does not restore them as runtime policy.
The existing stale unload-tool description was corrected to immediate revocation.

No persistent batch registry was added. Every successful assistant response,
including an empty one, already has a canonical Context entry. Before recording
the consuming response, the actor selects only complete, unambiguous result
entries after the previous assistant response and before the request boundary.
Results above T/100 become temporary proposed `Redacted(Result)` facts; small
results and the concrete keep control's own result are excluded. Only after all
companion outcomes are durable does the actor publish those cuts. A successfully
executed model keep control discards the proposals. Invalid/undispatched keeps,
host controls and same-named custom tools cannot veto them. The new assistant
entry is the once-only boundary across later human turns and zero-log restart;
canonical events retain original bytes. Keep does not bypass the ordinary ladder
or explicit clearing. Workers neither choose cuts nor mutate History.

Evidence in `/private/tmp/ch14-csharp.0m4UWS`:
- Final `keep-tests-final.log`, job 1985, exit zero: **88 passed, 0 failed**.
  New real-component tests cover exact thresholds, small/zero targets, incomplete
  and orphaned batches, repeated IDs, empty-response boundaries, opted-in startup
  grants/declarations, and documentary policy. Real loopback actor/Job runs cover
  ordinary opt-out, stub, keep, invalid keep, same-named custom tool, zero target,
  oversized kept batch still subject to the ladder, held companions, small result
  retention except whole-batch budget cuts, independent next-batch decisions and
  snapshot-only restart. Interruption pairs the undispatched keep before the cut
  without killing its independent companion Job.
- Sequential `keep-cli-build.log`, `keep-gui-build.log` and
  `target-archive-cli-exercise.log`, job 1986, exit zero; both builds have zero
  warnings/errors. Owned `/private/tmp/ch14-target-archive-cli.aujdf3jq/result.json`
  reports passed, four requests and exits `[0,0,0,0]`. A 120 KiB real file crosses
  the corrected default result band, reaches its consuming request whole, then
  appears stubbed after a later turn/restart; full archive dump and offline render
  remain correct. This refreshes the affected default-policy executable evidence.
  It does not exercise a host keep/stub flag: those switches are still unfinished.
- Direct whitespace scan of 24 actual changed source files passed. Git's scoped
  diff has no entries for these generated sources, so it does not verify their
  formatting. Last free space observed: 14,883,260 KiB.

Failures remain explicit: job 1958 (`keep-baseline-tests.log`) failed compilation
on my incorrect selector member/argument names; job 1962 (`keep-tests.log`) failed
compilation on guessed fixture signatures/registration. Job 1973
(`keep-tests-3.log`) had 87/1 because my skill fixture used unsupported inline YAML
`tools: []`. Job 1981 (`keep-tests-5.log`) had 87/1 through the fake HTTP boundary;
its oversized-batch fixture incorrectly expected the small companion to survive
whole-batch ladder clearing. Correcting that expectation and surfacing fake-server
assertions produced 88/0 in job 1983, followed by the final small-target regression
run above. Intermediate jobs 1967/1979 also passed 88/0 for their then-current
inputs. Staged edit anchor refusals wrote no partial changes. These are not extra
acceptance rounds or provider evidence; my repeated guessed names were avoidable.

Next: manual clear and host/GUI settings, controls and replay projection, then the
public consumer, scoped grader, C# teaching prose and source-only endpoint. Pending
manuals survive Chapter 14 cuts. There is still no Chapter 14 grade or endpoint;
real-provider, native/audible and human-accessibility work remains incomplete.

## Chapter 14 in progress — event-complete archive

Replaced the optional dialogue-only export with a canonical full-event archive.
History remains the sole event/reduction owner. Save merges the existing archive,
active recovery journal and retained events in local algorithm values, requiring
an unbroken prefix from sequence one through the snapshot boundary. Overlapping
facts must serialize identically; unknown versions, malformed UTF-8, gaps,
conflicts and future records refuse before file publication or trimming. Enabling
an archive after earlier history has been discarded explicitly fails; Context is
not used to manufacture missing events. Existing archives are independently
validated as complete prefixes. Ordinary diagnostic log loading still permits
sequence gaps and shares the structural decoder, not a second reducer.

Archive replacement precedes backup/primary replacement and journal/live-log
trimming. A later backup-write failure can leave the archive ahead of primary;
retry validates/reuses that prefix, while failed saving preserves the journal and
live diagnostic tail. Active diagnostic writer paths are tracked on History so
snapshot/archive replacement and journal startup cannot alias that open writer.
This is process-level replacement/flush behavior, not an fsync/power-loss claim.

Evidence in `/private/tmp/ch14-csharp.0m4UWS`:
- `archive-tests.log` (job 1908) and strengthened `archive-tests-2.log` (job 1917)
  both exited zero: **85 passed, 0 failed**. New real History/I/O tests cover exact
  events and request-prefix reconstruction through result stubbing/handoff,
  repeated zero-log saves and journal trimming, restart suffix merging, malformed
  and conflicting archives, missing-history refusal, active-log aliases, and a
  failed backup write after archive publication with successful retry.
- Sequential `archive-cli-build.log`, `archive-gui-build.log` and
  `archive-cli-exercise.log` (job 1920) exited zero; both builds had no warnings or
  errors. GUI evidence is compilation only.
- Owned executable evidence `/private/tmp/ch14-archive-cli.597a0_w9/result.json`:
  passed, four loopback requests, exits `[0,0,0,0]`. The actual CLI read a large file,
  recorded its later cut, saved the full archive across snapshot-only restart,
  dumped exact archived events and rendered reduced context offline without an
  extra provider request. `exercise-archive.py` and captures stay outside source.
- Scoped whitespace check exited zero; free space 14,937,772 KiB. Three staged
  edit attempts refused incorrect anchors before saving; no failed compilation or
  behavioral test is being credited. The final staged edit and library build passed.

Next: opt-in once-only keep/stub policy, manual clear and host/GUI integration,
then consumer, scoped grading, C# teaching prose and source-only endpoint. Pending
manuals survive Chapter 14 cuts; retirement belongs to a later chapter. No new
provider, native speech or human-accessibility claim.

## Chapter 14 in progress — recorded tool ladder and handoff replay cleanup

Implemented the two-band policy on the existing actor and immutable Context:
- Entry derives retained serialized tool bytes and whole-result presence once at
  construction; these caches are not persisted or separately mutable. Stub bytes
  count their retained representation, not their original payload size.
- At this earlier slice, Agent exposed independent 16 KiB byte budgets. **This
  API/default is superseded by the single target policy in the newer entry above.**
  Invalid updates changed neither value. At the next request boundary,
  the actor first records result stubbing, then optional stub-only tool removal.
  Pressure above twice a budget cuts oldest complete batches toward the budget.
  Dialogue/manuals are not charged or removed. Existing stubs remain unchanged
  when a later result cut crosses them.
- Selection groups parallel calls/results, permits reused IDs after a batch closes,
  and stops at incomplete/orphan/ambiguous traffic. It never splits a batch or cuts
  beyond the actor's successful-response watermark. That watermark is transient:
  after restart, restored results must be presented successfully before automatic
  cuts. Recorded sequence ranges, not current policy, govern replay.
- Review found a missed handoff contract: private replay material must also go.
  Handoff now drops OpaquePart and strips provenance/opaque signatures from retained
  TextPart while preserving its text and display ID. Original Context and canonical
  events remain unchanged. The regression includes signed text and opaque reasoning.

Evidence in `/private/tmp/ch14-csharp.0m4UWS`:
- `ladder-tests.log` (job 1836) and strengthened `ladder-tests-2.log` (job 1849)
  exited zero with **82 passed, 0 failed**. Tests cover exact high-water boundaries,
  returning toward budget, parallel/reused IDs, unseen/incomplete/orphan safety,
  byte-cache accuracy, stub idempotence, real actor delivery-before-stubbing, durable
  cuts and zero-log reload rebuilding caches.
- Final `handoff-privacy-tests.log` (job 1871) exited zero with **83 passed, 0 failed**,
  including the added canonical-versus-working-view privacy/identity regression.
- Sequential CLI/GUI builds plus `ladder-cli-exercise.log` (job 1862) exited zero.
  `/private/tmp/ch14-ladder-cli.1hxw2tjj/result.json` records passed:true, four owned
  loopback requests and exits [0,0]. A real CLI read returned a 50 KiB payload whole
  to the first consuming request; a later request and zero-log restart retained
  paired stubs, narration, dialogue and the primary manual, without replaying work.
  Canonical journal still held the full result. This exercise preceded the separate
  handoff-privacy fix and remains evidence for its unchanged ladder path only.
- Final `privacy-cli-build.log` and `privacy-gui-build.log` exited zero; scoped
  whitespace clean. Free space: 14,972,256 KiB. GUI evidence here is compilation
  only. No live-provider, native/audible or human-accessibility claim follows.
- One handoff patch anchor refused before saving; corrected against the actual
  reducer. No failing component/executable run occurred in this slice.

Scope correction: the revised chapter explicitly keeps pending manuals through
both ClearTools and micro_handoff; retirement belongs to a later chapter. The
previous next-work checklist was wrong. Its tiering requirement is the two-band
ladder, not a new multi-tier summary/Recall subsystem. Neither was invented here.

Next: event-complete archive replay (current optional archive is still dialogue-
only), the remaining opt-in keep/stub model policy, manual clearing and host/GUI
controls/settings/projection, public consumer, scoped grading and C# prose. There
is no Chapter 14 endpoint or grade. Carry prior unchanged recovery evidence.

## Chapter 14 in progress — reduction, handoff and pending authority

Implemented the next slice personally on the existing owners:
- Immutable Context derives active and pending-only tool grants. Pending unload
  now revokes local dispatch, load discovery, declarations and stale deferred
  references immediately while keeping the manual in SYSTEM. Another active
  manual can still grant the same tool. Rendering uses its frozen Context, not
  later live authority; snapshot-only replay also filters revoked references
  without rewriting canonical result metadata. Automatic snapshot selection
  already excluded pending grants and was left unchanged.
- History validates applicability before reduction. Missing known message/skill/
  unload payloads and invalid/absent redaction ranges leave Context unchanged;
  History-owned append/replay emits safe sequence/kind diagnostics and continues
  numbering. The corresponding optional known wire fields now survive decoding
  to that validator. Unknown event labels, bad supplied types/part structure,
  version and sequence corruption remain errors. No blanket exception-swallowing
  reducer was added. All replay paths share this diagnosed apply operation.
- `micro_handoff` is a built-in synchronous control, universally granted by skill
  startup alongside load/unload. A successful durable result stages a value-like
  `HandoffPart` on the actor. Only after all same-batch results (including refused
  tails) are recorded does it append `HandoffRecorded`. Reduction clears tool
  traffic and prior marked handoffs, preserving ordinary dialogue/system content
  and skill instructions. Provider projection renders the marked note as text;
  canonical events and snapshots retain its identity. A merely identically named
  custom/remote tool cannot trigger the control. Direct controls work too.
  No Recall/tier state exists yet; its replacement is still integration work.

Evidence in `/private/tmp/ch14-csharp.0m4UWS`:
- `handoff-tests-4.log`, job 1802, exit zero: **80 passed, 0 failed**. New groups
  cover pending/shared grants and actual file dispatch, all provider declaration
  surfaces, frozen Context and zero-log replay; total reduction and later valid
  replay with safe diagnostics/unknown-kind refusal; two handoffs around a held
  real companion Job, direct/blank-note controls and snapshot replay; and
  interruption with exact-ID tail pairing before reduction while the independent
  Job survives and completes. Prior journal and earlier component groups pass in
  this suite; unchanged executable/browser campaigns were not repeated.
- Sequential CLI/GUI builds and `handoff-cli-exercise.log` exited zero; both builds
  have zero warnings/errors. Actual CLI used dummy credentials and an owned
  loopback provider: four requests, a real write-file companion, two handoffs,
  newest note only on later requests, and a zero-log shutdown/restart retaining
  dialogue and primary instructions without replaying model work. Evidence:
  `/private/tmp/ch14-handoff-cli.mefrzw_m/result.json` records passed:true and
  exits [0,0]; requests and first-session journal retained beside it.
- Scoped whitespace check passed. Latest free space: 14,993,468 KiB. No provider,
  native speech, human-accessibility or GUI behavioral claim follows from these
  builds/component/CLI results. GUI live compaction projection remains unfinished.
- Failed gates remain failed: `authority-tests.log` was a fixture compilation
  error (`keepLog` rather than `RetainLog`); `reducer-tests.log` passed unnumbered
  facts into the numbered wire decoder; `reducer-tests-2.log` exposed the codec's
  missing-actor dereference (fixed along with optional known-payload decoding).
  The unknown-kind fixture initially used `kind` instead of `type` and was fixed.
  `handoff-tests.log`, `handoff-tests-2.log`, `handoff-tests-3.log` exited 1 on old
  absolute tool-count expectations and one system-ephemera fixture mistake. I
  should have checked all adjacent count assertions together; the loop now checks
  the exact installed name set. Source review also fixed hint-tail ordering before
  handoff application. Refused handoff patch anchors wrote nothing: the script
  validated all edits in memory before saving. Latest successful suite covers
  these saved corrections; earlier failures are not reclassified.

The later tool-ladder slice and corrected remaining scope are recorded above.
This earlier 80-group milestone is not Chapter 14 completion.

## Chapter 14 in progress — journal and crash-recovery slice

Read both manuscripts and the actual History/actor/persistence and grader
contracts. Retain the revision's corrections: loaded manuals stay once in SYSTEM;
new handoffs replace old active handoffs/Recall; pending unload removes dispatch
and declaration authority while retaining instructions; journal retention follows
the validated backup boundary. Do not import old Go topology, paid runs or worker
procedures. The newer reducer/skill/handoff slice is recorded above; compaction
and full archive integration remain unfinished.

Implemented recovery on the existing History owner. `StartPersistence` opens
`<save>.journal.jsonl` after successful recovery and before runtime reconfiguration.
Every admitted event is serialized/flushed before diagnostic output and live
publication. A failed append poisons further writes so an uncertain sequence cannot
be reused; shutdown reports that recording failure while still detaching ownership.
The existing save gate now serializes append/save/load/log setup/destruction. Disk
writes release the committed-view lock, so snapshots remain readable. Direct
`History.Save` without lifecycle activation remains snapshot-only.

Loading stages a primary or independently validated backup plus the journal suffix,
then publishes once. With no snapshot, a journal beginning at sequence one can
recover the first session. Complete malformed records, duplicate/gapped sequences
and invalid UTF-8 are refused; only an unterminated final suffix is ignored with a
safe diagnostic. Read-only recovery does not repair disk. Starting the recovered
writer removes that suffix before new appends. Save validates and rotates only a
good prior primary, keeps a good backup when the primary is damaged, and trims the
journal only through the validated backup's sequence. The journal is distinct from
the optional dialogue archive, and path aliasing with that archive/log is refused.
This is process-crash durability, not a power-loss/fsync guarantee, and not yet the
chapter's full archive replay or compaction-triggered persistence.

Evidence (owned scratch `/private/tmp/ch14-csharp.0m4UWS`;
TEXT pointer `/private/tmp/ch14-csharp-current`):
- `journal-tests-4.log`: 75 groups, zero failures. Four new groups cover primary/
  backup and first-session recovery, exact-once suffix replay, torn tails,
  transactional refusal, failed append nonpublication/cleanup, and a real POSIX FIFO
  file-read latch proving a held save blocks append but not committed readers.
- Sequential job 1729 exited zero: suite, clean CLI build in
  `journal-cli-build.log`, then `journal-cli-exercise.log`. Real owned CLI: clean
  snapshot, later durable terminal turn, SIGKILL exit -9, partial UTF-8 tail,
  restart, corrupted primary, backup restart. Four loopback model requests total;
  every prior human/assistant entry occurs once and no model work is replayed.
  `/private/tmp/ch14-journal-cli.1v_tp_hy/result.json` records three clean exits zero.
- `journal-gui-build.log` and scoped whitespace check: job 1732 exit zero.
  This is optional-host compilation, not a new browser exercise.
- Earlier runs remain failures: `journal-tests.log` (73/2) exposed unnormalized
  journal decode errors and a fixture incorrectly expecting clean shutdown after a
  poisoned append; `journal-tests-2.log` failed compilation on ImmutableArray.Count;
  `journal-tests-3.log` (74/1) exposed my repeated InvalidDataException/IOException
  hierarchy mistake. Product decoder errors now normalize to safe persistence
  errors; the fixture explicitly expects shutdown's reliable recording failure and
  checks complete parent detachment. All three process exits were collected as 1.
- Two staged persistence edits refused before saving; the first build covered only
  the saved declarations/append/journal helper, not the intended load/save rewrite.
  Later builds/tests above cover the full saved implementation.

The subsequent reducer/handoff/authority slice and current next work are recorded
above. This journal milestone alone never established Chapter 14 completion.

## Chapter 13 completed offline boundary

Read both manuscripts completely; finished the C# teaching prose, source and
ownership review. Preserved the historical exit-code/fence investigations as
history, not C# acceptance. Did not import the Go attach helper or native/paid
claims. The current chapter and READMEs teach the five-tool C# listener, real Job
completion, source-labeled admission and explicit limits.

`Examples/Gui/wwwroot/speech.js` is a pure raw-offset formatter: buffer incomplete
Markdown, handle fences before sentence/paragraph splitting, ignore soft wraps,
strip presentation/link destinations, expand identifiers and flush tails once.
`app.js` retains the existing per-card cursor and FIFO. Dispatch and errors flush
pending text/thinking. Successful tool bodies remain silent; announcements omit
arguments; thinking/error sources have audible labels. Manual card/MCP reading
uses the same filter. Escape stops speech with a draft without clearing the draft
or its human pause cause.

`?speech=transcript` bypasses only native playback. Both modes record exact filtered
text when it reaches the front of the queue, before synthesis; future cancelled
segments never become heard entries. `_ttsLog` records seq/ms/kind plus source,
text/part_id or pause cause. This proves admission, not acoustic hearing. Replay
stays silent, including a partial fence continued later: restored/muted parts stay
quiet rather than resuming midway through Markdown; new parts narrate normally.

Evidence: `/private/tmp/ch13-csharp.1YlqqJ`, with TEXT pointer
`/private/tmp/ch13-csharp-current` and controlled `offline` wrapper.
- `speech-tests-2.log`, exit zero: real shipped scripts in owned Chrome with working
  WebSocket/speech fakes. All fragment widths, fences/links/identifiers/paragraphs/
  tails, whole-final idempotence, source labels, error/tool flush, quiet/native
  parity, FIFO onend/onerror, Escape/draft and cancellation admission, partial replay.
  Browser artifacts: `/private/tmp/ch13-speech-browser.Buq0VS`.
- Portable regression `Tests/SpeechTests.mjs` uses Node without packages;
  `formatter-source.log` passed. No new dependency was installed.
- `component-tests.log`: 70 groups passed. `gui-build.log` clean. Sequential job1600
  exited zero including the existing actual-host dynamic/startup browser exercise:
  `host-browser.log`, evidence `/private/tmp/ch12-browser.LRGLSc`. This establishes
  adjacent MCP/startup/pause behavior, not new formatting alone. The later
  partial-replay-only correction passed `speech-tests-2.log`; final asset rebuild
  `gui-build-final-slice.log` has zero warnings/errors. Scoped whitespace clean.

### Restricted C# listener slice

`Examples/Listener` is a separate assembly consuming the public Agent API. Its
internal Session composition owner launches a fresh GUI, loopback port (not8088),
headless Chrome profile and direct DevTools connection using existing .NET APIs.
No automation package was added. Browser arguments disable extensions, use a mock
keychain and mute audio; only this owned profile is used. Startup, SIGINT/SIGTERM,
EOF, deadline and normal completion all run joined cleanup. Nonempty session
directories are refused before launch. The listener requires separate
`LISTENER_*` configuration; target configuration is `LLM_*`, with no fallback.

The empty library Agent receives exactly `say`, `listen`, `sleep`,
`set_speech_speed`, `save_history`, with no Skills manager or coding tools. Its
private observer retains only admitted utterances and `turn_ended`, not DOM text
or target response/tool-result payloads. `say` establishes pending status before
submission; `listen` distinguishes timeout/pending from explicit idle, advances
its own cursor, and retains source labels. JSONL contains only typed/heard records
with seq/kind/source/text. Saving collects but does not consume the listen cursor.
README and completed manuscript describe the actual C# tools; historical Go
interfaces are not the current endpoint's public contract.

Source review caught an important long-operation problem: ordinary custom tools
return running Job reports after the observation delay, but this restricted Agent
has no `wait_for_job`. Added optional `waitForCompletion` to the existing custom
Tool constructor (default false). Opted-in operations join the actual Job worker
and use the existing atomic structured result, preserving its one-time public
observation and full content. Job ownership, cancellation and default coding-tool
behavior remain unchanged. No second scheduler, result store or callback bag.
`Tests/ChannelToolTests.cs` holds a real custom Job with delay0/cap3, verifies its
20,000-character terminal result and no replay, then checks shutdown of a held Job.

Evidence in the same Chapter13 scratch directory:
- `listener-component-tests-2.log`: **71 passed / 0 failed**; sequential job1651
  exited0 after GUI/listener builds, executable exchange and lifetime exercises.
  `listener-gui-build.log`, `listener-build-final.log`, `listener-cli-build.log`
  compile the affected hosts; latest listener rebuild
  `listener-build-reviewed-2.log` has zero warnings/errors.
- `listener-exercise-reviewed.log`, job1655 exit0: real listener executable,
  owned GUI/Chrome and two loopback provider fakes. Exact five declarations,
  guessed `read_file` refusal, actual target submission, pending timeout, source-
  labeled whole-final/thinking speech, cursor/idle, acknowledged rate, sleep3.2s
  returning completion rather than a running handle, and semantic JSONL all pass.
  Evidence: `tmp/ch13-listener.36spy9j6` under scratch. Transcript admission, not sound.
- `listener-lifetime-reviewed.log`, job1661 exit0: real pending60s sleep, SIGINT,
  SIGTERM, EOF, overall deadline, and missing-browser startup failure. Each run
  observed its owned descendants and found none surviving completion.
  Evidence: `tmp/ch13-listener-lifetime.su03a4u2`. Same job verified nonempty-directory
  refusal with an unchanged sentinel; scoped whitespace passed.

Failures retained: initial browser startup attempts reached their deadlines
without any provider request, held in navigation. Using explicit Page.navigate
and the established automation isolation flags resolved startup; no single-flag
or keychain diagnosis is claimed. Cleanup still exited. Initial lifetime fixture
looked for nonexistent `output*` filenames instead of handle-named Job files;
corrected fixture interrupts a genuinely admitted sleep. First new component
compile used a nonexistent InvokeToolAsync API; actual Agent.DispatchAsync fixed
that fixture. One directory-refusal attempt lacked LISTENER_VENDOR and tested
only configuration refusal; corrected run reached the intended check. The README's
initial claimed configuration fallback was wrong and removed. Partial patch
refusals were inspected and finished before the final gates. No failed run earns
behavioral credit.

### Final speech grading and review

Added public `scripts/tts-harness.sh`/`.mjs`: explicit existing Node/.NET/GUI/Chrome/
driver paths, fresh grader-owned workspace/profile, dummy provider key, real form
submission and unchanged page admission JSONL. Transcript mode never claims
native sound. The Chapter13 `-executable-only` adapter retains all nine original
speech checks over five real GUI launches, excluding inherited Go checks.

- Initial `grade.json`: **65/95**, exit1. Fixture model was not in the explicit
  streaming-capability table, so fragmentation and typing gates lacked streaming.
  It also found a real product gap: provider errors are `lifecycle` frames with
  `kind:error_occurred`, not generic `error` frames. Added that precise speech path;
  ordinary lifecycle and terminal completion remain silent to avoid duplicates.
- `speech-tests-3.log`: actual page with working socket/native-speech fakes passes
  provider-lifecycle error narration once and silent non-error lifecycle, plus the
  preceding speech regressions. Updated harness uses the existing known streaming
  model `claude-3-haiku-20240307`; no production capability was invented.
- `grade-2.json`: **95/95, passed:true**, all nine checks, empty `grade-2.stderr`;
  sequential job1686 exited0 after focused page test, `gui-grade-build.log`, grade.
  This grade does not establish both pause-overlap orders or final-without-deltas;
  focused page checks and unchanged Chapter12 evidence retain their separate scope.
- Final readability-only listener cleanup expansion compiled cleanly in
  `listener-review-build.log`. `formatter-final.log` passed;
  `grade-flag-refusals.log` has four incompatible combinations exiting2;
  scoped whitespace passed. Job1695 exited0. The 71-group and actual listener
  exchange/lifetime evidence above remains applicable to unchanged behavior.
- Source/prose review covered formatter cursor/fences, queue admission, final/error
  paths, listener channel/lifetime and the small opt-in custom completion delta.
  No second scheduler, copied result store or browser model capability was added.

Preserved **135 source/support files** in `solutions/edition-3/ch13`, byte-compared
with main; build products, logs, profiles and captures excluded. Harness executable
mode preserved. No owned exercise process found at closure. Free space observed
14,991,424 KiB, above the 1 GiB floor. No commit, installation or running Ensemble
restart. Real-provider and native/audible/human-accessibility exercises remain
incomplete. Next: Chapter14 manuscripts, actual History/actor contracts and graders.

## Chapter 12 completed offline boundary

Implemented and personally reviewed startup skill selection and the corrected
on-demand browser-debug contract. `SkillSpec.StartupSkills` is frozen at
construction. `Skills.ResolveStartup` resolves the primary and ordered offered
selections as one dependency-first batch, with duplicate elimination and one
History activation. Local selections publish synchronously; server-backed
selections reuse preparation and actor commit without provisional grants or
fabricated tool dialogue. Readiness and browser-wait detection cover the full
initial closure. Existing constructors remain source-compatible; saved runtime
capabilities do not override current startup configuration.

Both hosts parse comma-separated `EN_STARTUP_SKILLS`/`--startup-skills` (explicit
flag overrides environment) and append `gui-debug` with `--gui-debug`. The shipped
manual owns its WebSocket connection and five on-demand tools rather than retaining
a separately owned browser dependency after unload. The GUI listener precedes
browser-dependent initialization. The CLI can use a supplied stdio-backed manual;
it does not invent a GUI transport for the shipped WebSocket manual. No static
catalog fallback or automatic browser observation was added. Provider instruction
placement reuses the existing three renderers; no Responses adapter is claimed.

The browser correction preserves input origin. `gui_input` replaces literal text
and emits a marked `CustomEvent("input", { detail: { mcp: true } })` plus `change`.
The page retains the human-typing cause rather than guessing from field contents
on every speech callback. Programmatic text does not pause the Agent immediately
or after speech finishes; genuine whitespace does, and successful submission
clears that cause. Speech pause remains independent. DOM input is not native
keyboard provenance or permission to overwrite a human draft.

**Evidence.** Scratch `/private/tmp/ch12-csharp.Yhd70P` (text pointer
`/private/tmp/ch12-csharp-current`). `final-tests-2.log` has **70 passed, 0 failed**;
the same sequential job **1553 exited 0** after the separate-assembly consumer in
`consumer.log` passed explicit startup selection, instruction rendering,
network-free initialization and destruction. New component groups cover ordered
selection, duplicate handling, invalid closure rollback, all three instruction
fields, frozen inputs, held optional-server readiness, real remote dispatch,
cleanup and a prior save surviving failed optional startup. Existing suites carry
unchanged contracts. CLI and GUI builds passed; final GUI asset/cause build is
`gui-cause-build.log`. The later edits to Common were comments only.

The scoped Chapter 12 executable grade is **90/90**, six passing checks and
`passed:true` in `grade-2.json`, empty `grade-2.stderr`, successful invocation. It
uses the existing `cmd/grade-edition2 -ch 12 -executable-only` adapter, preserving
all six CLI assertions while explicitly excluding GUI/inherited Go topology.
The initial invocation used a nonexistent flag and was refused before grading.
`cli-selection-2.log` is a separate successful real-CLI EOF check for environment
defaults, explicit override and duplicate selections; each process exited zero.
Evidence lives at `/private/tmp/ch12-selection.5rylp32e`.

`browser-cause-3.log` and job **1570 exit 0** establish the final real-GUI/isolated
Chrome exercise with local model and speech boundaries. Owned evidence:
`/private/tmp/ch12-browser.cTdM1j`; both `dynamic/host-result.json` and
`startup/host-result.json` report code 0, null signal, empty stderr. It exercises
shipped `--gui-debug`, first-request instructions exactly once with no startup
load exchange, no unsolicited observations, literal Unicode input, marked input
origin, programmatic draft surviving speech completion without a typing pause,
real human whitespace holding pause until cleared, snapshots, queue replacement,
unload/refusal, foreign identity/tab rejection and stopped-Agent replay. JavaScript
syntax and scoped whitespace checks passed. No provider spending, installation,
working-Ensemble restart, device changes, commits or publication occurred.

**Corrections and limitations.** I initially conflated the historical `pace()`
pause-cause discussion with slow character entry, added unnecessary pacing and
cancellation machinery, and accepted a manual-Resume workaround in an early
browser run. Final contract review caught both mistakes. I removed that machinery
and fixed the actual human/programmatic cause distinction; earlier pacing and
cancellation checks are superseded, not endpoint evidence. The corrected browser
fixture first waited for a transient value that no longer froze, then assumed
completion fields absent from its speech fake; jobs 1563 and 1566 exited 1. The
final fixture has a working speech completion boundary and no Resume workaround.
A new C# fixture initially guessed History APIs and a failure mode; its compile
failure is retained in `final-tests.log`, and corrected APIs/real closed peer
passed in `final-tests-2.log`. The CLI selection fixture initially used an invalid
manual field and save key; only its corrected run is credited. Some staged edits
refused before saving; one later script saved product edits then refused its
fixture anchor, so the saved state was inspected before retrying.

The C# chapter now teaches actual startup/ownership/input semantics while retaining
the earlier manuscript's narrative and live-run limitations as explicitly
historical material, not this run's evidence. README/manual guidance is updated.
The 122-file source-only endpoint is byte-equal to `main`, excluding bin/obj;
previous endpoints were untouched. Fake speech proves wiring, not native speech,
audible quality or human accessibility. Local peers do not prove real providers.

**Next:** read Chapter 13's current manuscript, preserved revision and actual
contracts, then continue personally. Do not rerun unchanged Chapter 12 campaigns.

## Chapter 11 completed offline boundary

The chapter is implemented, personally source-reviewed, exercised through real
local components/applications, and rewritten as C# teaching prose. The 120-file
source/support endpoint at `solutions/edition-3/ch11` is byte-equal to `main`,
excluding `bin`/`obj`; no logs, captures, profiles, or runtime data were copied.
README guidance now covers MCP ownership, startup, service mode, and browser
connection lifetime. Earlier endpoints were not changed.

Final review resolved the shutdown-save question: saved configuration is the
Chapter 10 documentary live projection, not an executable capability registry.
Removing remote Tools before shutdown save is consistent with reconstructing
connections and grants from current configuration at startup. No duplicate saved
catalog or resurrection mechanism was added.

Review found and corrected two concrete lifecycle/startup gaps:

- Client destruction now closes admission synchronously; outgoing RPC, catalog
  publication, and incoming service startup refuse a client already being
  destroyed. Resource termination and cancellation remain outside that admission
  lock in the joined cleanup task. The correlation regression checks immediate
  request/service refusal and repeated destruction.
- The GUI had waited for a browser for every server-backed primary. The read-only
  `Skills.RequiresBrowserStartup()` query inspects the real primary dependency
  closure, so only a WebSocket dependency requires browser registration. Stdio
  startup now completes without any browser; readiness still gates user work.

Final affected checks: `final-review-tests-4.log` reports **67 passed / 0 failed**;
CLI and GUI builds (`final-review-*-build-3.log`) have zero warnings/errors.
The sequential job 1508 completed those checks and the real no-browser GUI
exercise successfully, then exited 127 because its last browser command used
`node` outside the controlled PATH. That was a harness setup failure, not a
browser result. Retrying with the installed absolute Node path exited zero in
job 1512; `final-review-browser-2.log` passes the dynamic/primary browser campaign,
including stopped-Agent replay. Both saved host exits are zero in
`/private/tmp/ch11-browser-mcp.qfUHGz/{dynamic,startup}/host-result.json`.
The new `exercise-gui-stdio.py`/`final-review-gui-stdio.log` proves initialization
and saved activation without a WebSocket or model call, host exit zero, and
owned peer exit; evidence is `/private/tmp/ch11-gui-stdio.4yrvtkrt`.

The prior **80/80**, `passed:true`, exit-zero scoped CLI grade and separate public
C# consumer below are carried for unchanged successful protocol paths. Final
component tests cover the narrowed destruction admission and startup query; no
new broad grading campaign was run merely for a timestamp. JavaScript syntax
and scoped whitespace checks pass. Earlier staged-anchor refusals wrote nothing;
one resulting unchanged-source check run is not evidence for the intended edit.

The 263-line C# manuscript replaces Go-only APIs and contradictory browser
instructions with actual ownership, transactional startup, two protocol roles,
Job-backed request-only snapshots, on-demand browser tools, and explicit evidence
limits. General resource mapping remains outside the explicit chapter contract;
unsupported blocks fail rather than being silently dropped. No real-provider,
audible/human-accessibility, deployment, or restart-of-Bill's-Ensemble credit.

Next: read Chapter 12 and its preserved revision, inspect current request/cache
contracts and applicable graders, then implement personally under the same
procedure. Do not repeat Chapter 11 campaigns for unchanged code.

## Chapter 11 implementation history — transport, skills, snapshots and browser binding

Read the complete current chapter and compared the preserved revision. Retain its
on-demand browser tools and stricter exclusion of automatic tools from every
provider declaration/discovery surface. Neither manuscript has yet been converted
to C# for this chapter. The existing Go grader's active pipe probe initiates MCP
from the executable, unlike the chapter's client-initiated stdin service contract;
reconcile that exact boundary before adapting/running a grade, not by quietly
adding an undocumented protocol.

Saved implementation: `Common/Mcp.cs` declares the transport boundary and real
Agent → Mcp manager → McpClient relationships. `Mcp/Transport.cs` implements owned
newline-delimited UTF-8 streams with serialized writes and close/cancellation.
`Mcp/Client.cs` handles the chapter's 2024-11-05 initialize/initialized/tools-list
sequence, transactional discovery, numeric request correlation, concurrent
out-of-order replies, late cancelled replies, and reverse requests outside the
reader. Failed sends close the connection because partial line framing cannot be
reused safely. Manager/client destruction is joined/idempotent, wakes pending
callers, closes transport, joins reader/reverse tasks, and detaches both ownership
ends. Agent joined shutdown closes MCP while History is still attached; logical
Agent destruction also covers fault cleanup. All persistent declarations remain
in Common; the project includes the new functional directory.

Reverse `tools/call` now enters an explicit actor `ReverseTool` lane using the
existing control-worker completion path, not the active-turn queue. This avoids
waiting behind the turn that needs the server's reply. Ordinary Jobs admission
and skill gating still apply; the actor records `ToolObserved` (including real job
metadata) before replying. `tools/list` maps enabled local definitions. Reverse
and ordinary remote results now share `McpContent`: ordered text, image, and audio
blocks plus `isError` survive conversion. Unsupported resource/other block types
are explicitly refused as a whole result, not silently filtered. General embedded-resource
mapping is not required by either supplied Chapter 11 manuscript; the earlier
remaining-work checklist incorrectly promoted it into scope.

**Remote Tool/Job slice.** `Mcp/Tools.cs` registers the discovered ordinary catalog
through an Agent mailbox request with a reliable reply. It validates collisions
before attachment, supports repeated registration, and excludes automatic
`ephemeral` definitions entirely. Concrete remote Tools have two real strong
parents, Agent and McpClient. Add/remove methods maintain each edge; a foreign
Agent cannot adopt a still-connection-owned Tool. Destroying the client destroys
its Tools and removes both links. Tool membership snapshots use a short lock;
Tool destruction joins its logical unlink under a separate short lock. These locks
never cover transport I/O. Remote execution uses the existing Job admission and
worker path; it is not a callback bag or parallel scheduler.

Job retains a completed structured result until destruction, writes its JSON
diagnostic to the owned output file, and returns the typed content atomically with
actual Job metadata on the first completed observation. The cursor prevents later
ordinary observations from repeating it; automatic collection can read it without
consuming that public cursor. Ordinary
process/text output keeps its existing byte cap/cursor behavior. Structured final
content is atomic rather than byte-truncated; running reports still use normal
observation settings. Explicit kill of a remote tools/call sends `$/cancelRequest`
with the RPC ID; an observation timeout does not cancel remote work. A late reply
cannot change killed state or poison another request. A reverse call can invoke a
remote Tool through the same connection while the RPC reader continues receiving.

Inline base64 data URIs now resolve through BlobRef without network fetching.
Request validation checks the MIME matches the URI. The existing Anthropic nested
image projection retains image bytes, text order, and the error flag; the current
OpenAI/Gemini tool-result adapters explicitly refuse media placement. Audio is
preserved in MCP/neutral values but is not a claim of Anthropic audio support.
Remote URLs and job-handle media remain unsupported. No real provider was called.

**Server specifications and transports.** SkillDefinition now carries immutable
`McpServerSpec` values. `Skills/Servers.cs` parses the documented nested list with
unique names, stdio/websocket transport, required stdio command, and literal JSON
string-array arguments. It rejects unknown/duplicate/misindented fields, invalid
arguments and process fields on websocket entries; body text and subsequent
ordinary fields remain intact. Dynamic stdio activation now works as described
below. Primary startup and automatic-tool activation are now implemented below;
At this prerequisite slice WebSocket skill binding was refused; the browser-binding
slice below now implements it.

`McpProcessTransport` owns one real child process, reuses the line transport,
passes literal argv and an explicit workspace, allows only basic environment
variables (not provider/custom keys), drains stderr without mixing it into RPC or
accumulating logs, and joins one shared destruction task. This is environment
hygiene, not filesystem confinement. Shutdown kills/joins the owned process;
startup failure releases its resources. `McpWebSocketTransport` owns an already
connected socket, serializes sends, assembles fragmented text messages with the
same 16 MiB inbound bound, refuses binary/truncated messages, and aborts pending
I/O on destruction. HTTP upgrade, GUI ownership/identity, and `/mcp` are not yet
implemented or tested by this transport fixture.

Evidence: `/private/tmp/ch11-csharp-current` is a TEXT pointer to
`/private/tmp/ch11-csharp.FFfxKy`; its offline wrapper has owned home/tmp and the
approved SDK/installed Go caches. Foundation `foundation-final-tests-2.log` had
45 passing groups, job 1124 exit 0. `remote-tests-final.log` had **47 passed,
0 failed**, actual job 1167 exit 0. Two added groups cover the remote behaviors
above using real loopback streams, actual Agent/Jobs, saved output and History,
provider rendering, and direct relationship removal/reattachment. They also check
malformed base64/isError and unsupported resource refusal, transactional name
collision, repeated registration, automatic exclusion, and reverse media. Earlier
foundation checks retain handshake ordering, concurrent/out-of-order RPC, late
reply isolation, held-turn reverse file effects, peer EOF, duplicate discovery,
joined shutdown, and partial-write closure. No Chapter 11 grade, executable/browser
exercise, external consumer, endpoint, or completion is claimed. Latest disk check:
16,189,584 KiB free. No workers or running exercises remain.

Earlier `transport-prerequisites-final.log`: **50 passed, 0 failed**, actual job
1189 exit 0. Three additional groups cover nested frontmatter/refused activation,
a real subprocess peer (handshake, large stderr, literal argv/workspace, no
inherited dummy secrets, held RPC shutdown, natural EOF and failed startup), and
real RFC 6455 framing over owned loopback streams (fragmented/out-of-order RPC,
normal close, binary refusal, mid-message close and aborted pending read).
`websocket-tests-3.log` also passed 50 groups, job 1187 exit 0. The preceding two
WebSocket runs stopped at compilation: missing test SendAsync cancellation
arguments, then wrong SkillSpec/provenance calls. Jobs 1182 and 1184 exited 1;
no behavioral credit applies to them. Scoped whitespace check passed.

Corrections: several exact-anchor scripts refused, sometimes after preceding edits
had saved; those attempts earned no behavioral credit. Early foundation helper
and enum mistakes remain in their logs. The remote slice initially used guessed
BlobKind/Ref names instead of actual MIME/BlobRef and incorrectly used RefKind.Data
rather than Uri. Adding ToolKind.Remote exposed an enum-scanning installer bug:
built-in installation tried to construct it without a client, causing nine old
groups to fail. Excluding Remote restored them. `remote-tests-3.log` had 46 passes
and one diagnostic failure; the changed message omitted the fixture's useful
“local paths” wording. My first explanation wrongly blamed locator-kind wording;
reading the actual assertion corrected that. `remote-tests-4.log` and final both
exited zero. Source review closed a result-publication window and serialized Tool
unlinking. A suspected cleanup cycle did not warrant a new mechanism: client
lifetime cancellation already releases the reverse caller's wait while retaining
independent Job lifetime; owner-held destruction is still joined.

**Dynamic stdio skill activation.** Actor admission resolves the dependency closure
and captures variables/workspace. Off-actor preparation starts process peers and
performs RPC discovery while the actor stays available for reverse calls. A
transient SkillActivation mailbox message returns for validation and one
dependency-first SkillLoaded publication. Commit revalidates closure, explicit
per-skill tool allowlists, providers and namespace before attaching Tools. No early
grants/manuals are published. Failure joins every prepared client. Append-failure
rollback is source-reviewed, not fault-injection-verified; no writer abstraction
was added solely to test it. Once queued, commit is awaited to a definite result
before the preparer decides whether it owns rollback cleanup.

Clients retain real strong parents Mcp and Skills, with complete two-way add/remove
operations. SkillName identifies the immutable definition; configured server names
are unique within a skill rather than a global alias. Catalog installation/Tool
attachment enforces that skill's explicit tool list; extra discovery grants
nothing, and a detached skill client cannot broaden access. Unload records a pending
manual, destroys/joins only that skill's clients and Tools, and leaves other skills'
servers intact. Whole-Agent shutdown owns all clients through Mcp. No service bag,
parallel scheduler, or extra grant registry was introduced.

Earlier `activation-tests-final-3.log`: **54 passed, 0 failed**, actual job 1244 exit 0.
Four new groups use real process peers and Agent/Jobs/History, with loopback HTTP for
model-driven interruption. They cover initialization-time reverse write, no early
grants, captured variables/frozen primary, real remote Job execution, selective
catalog, direct two-way detach/reattach, one dependency-first publication, same
server name under two skill owners, independent unload, missing provider/collision/
closed-peer rollback, real PID termination, shutdown and interrupted model load
with exact-ID terminal pairing. Scoped whitespace passed. No primary startup,
application, consumer, browser, grade, prose or endpoint completion is credited.

Slice corrections: several guarded edits refused after prior edits had saved;
compiler errors were fixed before behavior was credited. The old launch-refusal
fixture expected no manager at all; failure now leaves an empty reusable manager,
with no clients/grants. `activation-tests-final.log` (job 1232 exit 1) incorrectly
expected a turn interrupt to cancel a host skill control in the independent lane.
I first misread its timeout as failed interrupt acknowledgement; the acknowledgement
succeeded and its later load waiter timed out. A real model-driven test then found
a product defect (`activation-tests-final-2.log`, job 1242 exit 1): cleanup completed
but cancellation skipped ToolReturned. Jobs now returns an explicit cancellation
error after preparation cleanup, allowing the actor to record the exact call ID
before ending the turn. Persistence faults still propagate. The final run above
verifies the correction. `activation-tests-4.log` had 54 passing groups (job 1227
exit 0), but lacked that corrected model-interrupt scenario.

**Primary stdio startup.** Skills construction still publishes local-only primaries
synchronously. A server-backed primary freezes its variables at construction and
requires `InitializeAsync` after persistence load/log setup and before user work.
The actor tracks this as ordinary lifecycle work, not an invented tool call. It
stays responsive to controls and reverse RPC while the existing preparer connects
the dependency closure. Pending startup exposes no provisional grants. Commit
publishes one dependency-first activation with exactly one primary and its control
grants. Repeated initialization callers join one operation; caller cancellation
detaches only that waiter, whereas shutdown cancels/joins the owned preparation.
Failed startup leaves no remote Tools/clients/processes and cannot overwrite a
previous save on shutdown. CLI and GUI explicitly awaited initialization at this
slice; WebSocket binding was then refused. The later GUI integration below starts
its listener before waiting for browser-backed startup. Activation returns immutable discovered names;
only a real load_skill invocation wraps those names in a ToolResultPart. Startup
never invents a call ID or records ToolObserved.

`startup-tests-4.log`: **57 passed, 0 failed**, collected job 1288 exit 0, also
building CLI and GUI with zero warnings/errors. Three new groups use real process
peers: dependency-first primary startup, reverse-call refusal before grants,
construction-frozen primary variables, cancelled/concurrent waiters, no early model
work, pause/resume, primary instruction placement exactly once in OpenAI's system
message, remote job execution, complete shutdown; closed-peer rollback preserving
prior save bytes; shutdown during held initialization. `startup-cli-exercise-2.log`
exited 0: actual CLI/EOF startup with an owned stdio peer, process termination,
failed startup exit 1 and unchanged saved session. Its private evidence directory
is `/private/tmp/ch11-primary-cli.jeaqiv1_`. GUI compilation is not a live GUI startup
exercise. The final layout-only move of startup mailbox code into Agent rebuilt
cleanly (`startup-layout-build.log`); scoped whitespace passed. No Chapter 11 grade,
consumer, browser, prose conversion or source endpoint is credited.

Startup corrections: relative/stale project paths failed before compilation;
actual core is Ensemble.csproj, CLI is Cli/Ch04.csproj. A nonexistent fixture Text()
helper caused a compile-only failure (`startup-tests.log`, job 1275 exit 1). The
first implementation reused a ToolResultPart for startup and failed its real-ID
invariant (`startup-tests-2.log`, 56/1, job 1277 exit 1). Rather than invent an ID,
startup now uses its own ordinary reply and activation returns discovered names.
The next run reached shutdown but the test dereferenced the deliberately detached
Mcp parent (`startup-tests-3.log`, 56/1, job 1284 exit 1); retaining the manager
reference verifies both its empty collection and the detached Agent link. The first
CLI peer supplied an unsupported protocol version and was correctly refused; the
corrected fixture echoes the requested version. No provider call was made.

**Automatic snapshots.** Published `per_turn_`/`per_round_` prefixes now classify
server definitions; existing `ephemeral` annotations remain supported, with
conflicting modes refused. Automatic grants remain immutable skill facts but never
become ordinary Tools. Agent admission, `$TOOLS`, enabled/deferred catalogs, provider
declarations and final Anthropic discovery-reference emission exclude them. Request
filtering does not rewrite recorded metadata. Selection requires the live granting
skill and its explicit allowlist; discovery alone and detached owners grant nothing.

The actor reserves real automatic Jobs and records admission before effects. Existing
Job workers call MCP; an off-actor collector awaits completion without consuming the
public observation cursor. Immutable mailbox outcomes alone return to the actor.
Per-turn values are retained only for the current turn; per-round values are refreshed
before each request. Immediately before rendering, the actor combines snapshots with
the caller's current ephemera slot and records a system MessageReceived fact. They
are replayable request input, not durable Dialogue or instruction bodies. RequestSent
consumes the slot after successful rendering; rendering refusal clears it with an
empty system-slot event without claiming transport. HTTP failure also leaves no slot.
Pause holds the send after completed collection; hints remain responsive. Both turn
interrupt modes detach collection without killing independent Jobs or reusing their
late output in another turn. Shutdown still owns and joins those Jobs and clients.
Expected closed-client races are ordinary MCP execution errors, not Job invariant faults.

`automatic-tests-9.log`: **62 passed, 0 failed**, collected job 1374 exit 0, followed
sequentially by clean CLI/GUI builds (`automatic-cli-final-build.log`,
`automatic-gui-final-build.log`) and scoped whitespace. Five added groups use real
process peers, Agent/Jobs/History, and actual loopback request bodies: turn/round
cadence, second-turn reset, manual-slot coexistence, request replay, allowlist
exclusion, dynamic activation/unload, guessed-name refusal, canonical-reference
preservation, declaration filtering across adapters, held pause/hint/both interrupts,
independent output inspection, server isError, HTTP failure and pre-transport media
refusal. `automatic-tests-7.log` also passed 62 groups before the final replay-reference
guard; its unchanged CLI exercise remains applicable. `automatic-cli-exercise.log`
exited 0 with owned evidence `/private/tmp/ch11-automatic-cli.8gnt2vft`: actual CLI,
two model rounds, one turn snapshot/two round snapshots, hidden declarations/$TOOLS,
ungranted automatic exclusion, ordinary remote Job, reduced saved Context without
snapshots, normal EOF exit 0 and terminated peer PID. No real provider was contacted.

Corrections: fixture API/enum guesses caused compile-only failures (tests 1 and 8).
Tests 2 exposed missing automatic filtering at Agent.ToolEnabled; tests 3 exposed
incorrect interruption expectations and a declaration assertion that wrongly rejected
ordinary read-file content. Cleanup also exposed a real queued-Job/closed-client
race, now converted narrowly to an MCP availability error. Tests 4 passed 60 groups.
Tests 5 exposed my incorrect RequestSent-before-render change against the established
unsupported-media contract, plus a fixture expecting raw rather than wrapped HTTP
failure. The current empty-slot cleanup preserves that contract; tests 6/7/9 passed
62 groups. These failed receipts remain failures, not retroactively credited. Latest
free-space observation: 16,034,020 KiB; no workers or running exercise jobs remain.

**WebSocket browser binding and application slice.** `McpBrowser` is a real
manager-owned registration for one connected browser transport. A claimed client
also has that registration as a strong parent, alongside manager and Skills.
Add/remove validates before changing both ends. A connection grants nothing until
ordinary skill preparation discovers its catalogue and actor commit publishes the
explicit allowlist. Discovery failure or disconnect joins/detaches the prepared
client without partial grants. A consumed registration cannot be reused with reset
RPC IDs; disconnect removes it before a fresh connection can be registered. Browser
and manager shutdown join cleanup, including an unclaimed transport. Manager cleanup
snapshots its remaining clients after browser destruction, avoiding duplicate traversal
of the multiply-owned client. Constructor failure rolls back all attached parents.

`Skills/Connections.cs` now binds websocket specifications to that registration and
reuses the existing preparation/commit path for both dynamic and primary skills.
No additional scheduler, tool registry, automatic grants, or process launch was added.
Three groups in `Tests/McpBrowserTests.cs` use actual local WebSocket framing: dynamic
activation/ordinary remote Jobs/unload and ownership refusal; held primary startup;
disconnect during discovery and transactional rollback. Core evidence:
`browser-binding-tests.log` (job 1395 exit 0), followed by the cleanup-reviewed
`browser-binding-final-tests.log`: **65 passed, 0 failed**, job 1424 exit 0. That final
job also built CLI and GUI cleanly (`browser-cli-build.log`,
`gui-browser-final-build.log`). Earlier Chapter 10 campaigns were not repeated.

The GUI tunnels `{type:"mcp", agent_id, data}` on the chapter's existing `/ws`, not
a separate `/mcp` endpoint. `McpTunnel` is a socket-scoped I/O resource with a bounded
MCP input queue and the host's existing single outbound writer. RPC replies are not
lossy display notifications. Its disposal closes MCP, not the shared conversation
socket. Only the first registered tab is the tool target; other tabs remain observers
and cannot inject RPC. Supplied mismatching Agent IDs are refused before control or
RPC dispatch. Disconnect destroys that registration/client/tools without cancelling
the Agent's independent work. A stopped Agent still accepts replay-only observers.
GUI startup now opens the listener before awaiting a browser for a server-backed
primary; the existing readiness gate continues refusing premature work. CLI startup
and explicit library initialization retain their existing sequencing.

The shipped optional `browser` skill and `wwwroot/mcp.js` expose only on-demand
`gui_snapshot`, `gui_click`, `gui_input`, `tts_queue`, and `tts_say`. They implement
initialize/initialized/discovery/call on the actual page. Snapshots report visible
DOM text/selectors and truncation flags, omit password values, and are not pixel
vision. Replies use the originating socket, not a later reconnect. Input/click
results say dispatch, not verified effect. Speech uses the existing authoritative
current utterance/FIFO; interrupt replaces both. No automatic browser attachment
or unsolicited snapshot tool is provided. Fresh snapshot/queue requests observe
results when needed.

`exercise-browser-mcp.mjs` is an owned real Chrome/Puppeteer + GUI + loopback-model
exercise. `browser-mcp-exercise-3.log`, job 1426 exit 0, saved evidence at
`/private/tmp/ch11-browser-mcp.ejuVBB`: 17 model requests across dynamic and primary
runs; disabled calls before load/after unload; shared-envelope discovery; actual
snapshot-selected click, text input then a fresh snapshot; real typing pause and
explicit Resume; queue/FIFO replacement using a speech fake; foreign Agent and
non-target-tab RPC refusal; live GUI retained on unload; primary-browser startup;
new stopped-Agent replay connection. Both host-result files report exit 0, no signal,
and empty stderr. The saved primary retains its manual. No real provider was used;
speech evidence is queue wiring, not audibility/human accessibility. Scoped whitespace
check passed. Free space was 15,800,504 KiB.

Corrections are retained: the first browser run (`browser-mcp-exercise.log`, job 1412
exit 1) found that InvalidDataException was absent from the GUI's expected-control
catch, so a refusal disconnected the socket. The fixture also incorrectly expected
private exception wording instead of the existing generic public error. The second
run (job 1418 exit 0) verified those corrections but required bringing its owned
background test tab forward; that intervention is not concealed. The final fixture
explicitly selects that tab and uses timer polling for the observer, and passed
without intervention. Source review additionally removed duplicate shutdown traversal
and protected post-shutdown observation; the final suite/build/browser run cover them.
A refused patch anchor made no source change. JavaScript syntax and copied-asset
checks passed before the executable exercise.

### Incoming CLI service, scoped grade and public consumer — 67 groups

`McpClient.ServeAsync` is the explicit incoming-service mode, distinct from
`ConnectAsync`, where Ensemble initiates an outgoing client handshake. One transport
cannot run both readers. The incoming service validates integer identities (including
zero), accepts the external client's initialize, waits for initialized, and then uses
the existing reverse handler and Agent/Jobs admission. It neither discovers a remote
catalog nor starts a model turn. EOF joins completed dispatch; owner shutdown releases
and joins the reader. Expected protocol/cancellation outcomes remain observable through
ServeAsync without becoming a second cleanup fault. No additional dispatcher or
conversation model was introduced.

CLI `--mcp-pipe` performs ordinary skill/persistence/log startup, then exposes only
JSON-RPC on stdout. It joins client destruction and Agent shutdown on exit. The old
handshake concern was resolved by distinguishing incoming service from outgoing MCP
client roles, not by changing the exercised protocol. The chapter's C# prose still
needs to explain that distinction explicitly.

`service-tests-4.log` reports **67 passed, 0 failed**, collected job 1468 exit 0;
that same sequential job built the CLI and ran the final grade. The two service groups
use real stream peers and Agent tools: handshake gating, unsupported versions, exact
identity correlation, allowlisted discovery, refused write with no effect, real file
read, unknown method, no model request, clean EOF, malformed identity, cancellation
and owner shutdown. Initial `service-tests.log` and `service-tests-2.log` failed on my
incorrect fixture manual (unsupported bracket list, then omitted required type), not
product behavior. `service-tests-3.log` passed all 67 groups before the later snapshot
error refinement; its CLI build also passed. Those failed runs remain failed receipts.

The existing `grade-edition2 -executable-only` adapter now permits Chapter 11 and
runs five unchanged CLI MCP probes, preserving their IDs and weights. The full route
still includes its GUI and inherited checks. `grade-final.json` is parsed **80/80,
passed:true**, empty `grade-final.stderr`, actual job 1468 exit 0. This does not award
GUI or Go-topology credit. The historical Go Chapter 11 grader test requires the old
Go endpoint and was not claimed as C# evidence. Four incompatible flag combinations
exit 2 in `grader-flag-refusals.log`.

The first scoped `grade.json` was a real **60/80 failure**: failed automatic output
was appended in a separate content block after an error label, while the grader's
predicate required any retained error marker to share a labeled line. The chapter
requires logged nonfatal errors, not that literal layout. I chose the clearer request
projection: a failed snapshot contributes only a brief unavailable/error diagnostic
and real Job handle; its complete bytes remain in the Job and ToolObserved. The updated
regression checks absence from the request and preserved error bytes through public
Job observation. Successful snapshots, ordinary remote error results, and canonical
facts retain their existing behavior. No grader assertion was weakened.

`scratch/consumer/Consumer.csproj` is a separate assembly referencing only the public
library. Restore used an owned empty package source. `consumer.log` confirms real
stdio skill activation, pre-load/post-unload refusal, ordered two-part remote output,
both client owner memberships, joined process exit, incoming stream service with
integer correlated replies, real read_file and explicit MCP detachment on shutdown.
No model/provider call occurred. Job 1473 completed with exit 0 after its consumer and
clean GUI build; my final Git command had omitted --no-pager and used the wrong cwd,
so I stopped only its identified owned less child and reran the actual scoped check
from the repository root. This was a command-wrapper mistake, not a consumer failure.
`service-cli-final-build.log` and `service-gui-final-build.log` have zero warnings/errors.
The final source does not change the browser campaign's exercised successful paths;
carry that evidence rather than rerun for a timestamp. Last free space: 15,707,620 KiB.

**Next:** Read and convert Chapter 11 prose against the actual C# APIs and preserved
revision; finish source review, including saved configuration after remote catalog
teardown, before preserving an endpoint. General resource mapping is outside this
chapter's explicit contract. No Chapter 11 endpoint or chapter completion is credited.
Continue personally; do not repeat unchanged Chapter 10 or browser campaigns.

## Chapter 10 completed offline boundary

**Model and recovery.** History owns one reduced Context and explicit folded
sequence; coherent snapshots include both plus retained events. Common declares
value-like PersistenceOptions/SaveDocument and all partial operations. Agent's
startup-only StartPersistence captures current skill values, validates/loads the
primary or backup, records RuntimeConfigured, then arms automatic saving. Joined
shutdown records completion and saves while real parent paths remain attached.
The library opts in explicitly; CLI/GUI default to their own working-directory
save.json and expose save path, diagnostic retention, optional archive and disable
controls. StartLog writes the restored/bootstrap diagnostic prefix, not an invented
full log after retention. Failed startup cannot overwrite bad candidates.

Save has the published four fields: context, as_of, config, log. Neutral Context
preserves exact JSON, signatures/provenance, media locators, usage, summaries,
redactions and finalized display IDs. Credential-free effective config is projected
from the current Engine; saved config never supplies credentials or runtime policy.
Snapshot diagnostics at/below as_of are skipped; newer tails fold once. Log-only
replay starts from empty Context but future numbering honors max(as_of, replay).
Candidates fully validate before publication. Save operations serialize separately
from short History locks; a captured prefix cannot trim later appended events.
A same-directory flushed temporary replaces primary, after validating candidate
and old primary before backup rotation. Corrupt primary bytes cannot replace good
backup. This is not a journal or a guarantee for unsaved events/power failures.

**Retention and current runtime.** Negative/all, zero/none and positive/newest-N
retention applies only to diagnostics, in memory and on disk. Optional dialogue
JSONL writes from the full reduced snapshot before trim; it is derivative, not a
recovery authority or Chapter 14 full archive. Archive/primary are not a multi-file
atomic transaction; failure leaves live diagnostics intact. Paths must be distinct.
RuntimeConfigured resets old skill grants, pending flags and stopped/paused phase,
using today's initial skill values while preserving ordinary historical system
entries. Pending unload now belongs to immutable reduced SkillPart state, so
retention cannot erase it in a continuing session. Source lookup likewise uses
Context rather than discarded events. Summary redaction preserves skill entries
without borrowing their IDs; part filtering keeps matching parser IDs.

**GUI and teaching.** Replay projects finalized Context entries plus retained
lifecycle/direct-tool diagnostics, supporting snapshot-only restoration with stable
IDs and without resurrecting work. Live capture bounds both Context and events to
the observed sequence, preventing future-final lookahead. Chapter text is converted
to actual C# ownership/APIs, persistence limitations and exercises. READMEs explain
new default files and explicit library opt-in. Examples/Persistence is a separate
assembly using public APIs for snapshot-only restore, runtime model change,
sequence continuity and joined automatic saving; no provider required.

**Evidence.** Scratch `/private/tmp/ch10-csharp.82CNoz`, text pointer
`/private/tmp/ch10-csharp-current`; controlled offline wrapper uses approved SDK,
owned home/tmp and existing Go caches. `final-component-tests-5.log`: 42 groups,
0 failures, exit zero (job 1096), including portable values/redaction/summaries,
snapshot/tail/log-only/retention/archive, isolated malformed candidates and backup
recovery, failure nonmutation, current runtime/skill reset and pending retention.
`consumer-3.log`: independent C# client passed, job 1084 exit zero. Final CLI/GUI
build logs `cli-build-final-2.log` / `gui-build-final-2.log`: zero warnings/errors.
`grade-final.json`: parsed passed:true, 85/85, empty stderr, job 1093 exit zero.
The executable-only route reuses the six existing Chapter 10 CLI checks, excluding
inherited Go scope and stdin-GUI probes; full Go route remains unchanged. Four
incompatible flag combinations exited 2 (`flag-refusals.log`).

`browser.mjs` / `browser.log` exercised the actual HTTP/WebSocket host using an
owned isolated existing Chrome profile and dummy loopback provider. Work directory
is recorded in the log. Snapshot-only shutdown/restart restored visible human and
assistant cards with identical IDs; current instructions governed the next request,
conversation continued, archive contained dialogue, refresh did not duplicate,
and both host exits were zero. `exercise-failures.py` / `failure-exercise.log`
proved blocked archive save exits 1 without creating a primary, and --no-save
ignores an invalid existing candidate without changing its bytes. Browser evidence
is carried across later shared codec validation, pending-state/summary corrections
and readability changes, which were tested in the final component suite and scoped
grade where relevant. No provider, audible speech or human-accessibility credit.
Earlier unchanged lifetime and Chapter 9 protocol evidence is carried, not rerun.

**Corrections retained.** Several guessed signatures/anchors caused refused edits
or compilation-only failures; these earned no behavior credit. Initial fixtures
omitted HintReceived.actor, guessed media/tool constructor shapes, and failed to
recognize InvalidDataException. The first save writer invented version/seq instead
of the public four-field/as_of contract; both writer and reader were corrected,
including log-only boundary semantics. The first scoped score was 65/85: real
reducer fallthrough made hints system entries, so vendor switching lost user roles.
This was a product defect, not a grader shape assumption; hints now retain their
recorded actor, with an explicit tail regression. Source review caught event-derived
pending state disappearing under retention and summary IDs colliding with preserved
skills; both corrected and tested before the endpoint. Consumer initially omitted
the root diagnostics writer and reversed URL/key arguments; corrected before its
successful run. No installation, restart of Ensemble, commit, push, provider use,
or shared cleanup occurred. Latest free space: 16,358,432 KiB, above the 1 GiB floor.

## Chapter 9 completed offline boundary

**Ownership and activation.** Agent owns concrete Skills with complete add/remove
links and explicit destruction. Common declares all persistent state and callable
members; the functional Skills module implements parsing, closure resolution,
variables, activation and pending removal. Catalog/templates and registered
variables are configuration; immutable SkillPart grants in History are the sole
active capability/instruction facts. The primary is rendered once. Dependencies
validate before one durable activation batch, with progressive discovery and
load-time variable snapshots. Loads are actor admission controls, not jobs.
Disabled calls are refused before effects. Pending unload retains instructions
and tools; the primary cannot be unloaded. Controls remain available even when
no further skills are offered. Detached manager calls fail explicitly.

**Instruction and provider contracts.** Retained the preserved revision's
instruction-role correction: loaded manuals occur once in each provider's
instruction field, never only tool-result prose or duplicate dialogue. Rendering
uses captured grants. Ordinary redaction preserves skill facts. The optional
Anthropic-only `EN_DEFERRED_TOOLS=1` / `Configuration.deferredTools` is off by
default. Initial grants stay eager; reachable later tools stay deferred. Successful
loads carry newly granted tool names through canonical result events, codec,
numbering, reducer, observations and redaction into paired tool_reference blocks.
Full schemas still travel in every request; this is not wire compression. The
primary documentation returned HTTP 200 at
https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-search-tool.
Documentation verification and fake protocol tests are NOT live-provider proof.

**Applications and teaching.** CLI/GUI open logs before initial activation,
accept EN_SKILLS_DIR/EN_PRIMARY_SKILL/EN_CUSTOM_VAR and path/name flags, and ship
three small primary/loadable examples. The default starts with inspection tools;
`full` is an explicit trusted-exercise alternative. Resource globs copy manuals,
not implementation source. Library inputs remain explicit. Offline CLI rendering
registers control schemas without constructing a manager or rereading mutable
manuals; recorded grants/bodies survive catalog removal. Without a live catalog,
replay uses portable declarations rather than inventing deferred reachability.
Chapter 9 and READMEs now teach actual C# ownership, transaction ordering, roles,
configuration, limitations, and exercises. Activation uses expanded named steps.
The separate `Examples/Skills` assembly uses only public APIs and performs real
writes, including after pending unload, with explicit graph destruction.

**Verified scope.** Scratch `/private/tmp/ch09-csharp.IC88if`, text pointer
`/private/tmp/ch09-csharp-current`. Final component run
`final-component-tests-2.log`: 39 groups, zero failures, exit zero. Includes real
three-request actor/HTTP load→write→final in ordinary and deferred modes; all three
instruction projections; transactional refusal, cycles, unknown tools, variable
snapshotting, progressive discovery, primary refusal, pending/idempotent loads,
isolation, manager links/detached control and no-offers deferred primary.
`consumer.log` records the separate public consumer's successful real behavior.
CLI and GUI builds have zero warnings/errors. `grade-final.json` is parsed
passed:true, 85/85, empty stderr and command exit zero: eight existing skill checks,
excluding inherited Chapter 8/Go scope. Four invalid combinations exit 2 in
`flag-refusals.log`. Grader package `-run '^$'` was compile-only, not behavioral
credit. Earlier HTTP lifetime, GUI settings and streaming evidence carries only
for unchanged behavior; no entire earlier campaign was replayed.

`exercise.mjs` / `exercise.log` ran actual CLI processes for Anthropic, OpenAI and
Gemini against local working fakes: seven requests each, disabled write with no
file, dependency activation, real writes before/after pending unload, readback,
normal EOF exit zero, and key-free replay after moving the catalog aside. The
same exercise used installed Chrome/Puppeteer with an isolated owned profile
and port: visible prompt, expanded skill/tool-result cards, pending explanation,
real file effects, and exact stable-ID replay after refresh; no page errors and
GUI host exit zero. Its unique work directory is recorded in the log. No normal
browser, live Ensemble, account, device, or port 8088 was touched. This adds no
audible speech, human accessibility, or real-provider evidence.

**Corrections and limitations.** Initial variable spelling and control grants
were reconciled with the public `$SKILLS`/universal-loader contract, not a weakened
grader. The first grader invocation used invalid `-ch9`, exiting 2 with no checks;
actual spelling is `-ch 9`. An early test used a zero-based History range; another
looked for nested ToolResultPart where canonical result events store contents.
That second mistake exposed a real missing-discovery-metadata defect, now fixed
through the canonical/replay paths. Review also found missing control schemas in
offline CLI rendering and no-offers deferred catalogs; actual replay and regression
coverage now protect both. Several guessed exact edits refused, sometimes after
prior edits saved; only inspected saved artifacts and actual build/test exits count.
No claim that historical Go deletions in broad Git output were Chapter 9 changes.

`solutions/edition-3/ch09` has 82 byte-verified source/support files matching main,
excluding bin/obj, private captures, logs, profiles and credentials. Whitespace
checks passed. Earlier endpoints are untouched. Real-provider exercises,
account-specific deferred support, audible/human accessibility, installation,
deployment and publication remain incomplete or outside authority. Next: Chapter 10.

## Chapter 8 completed offline boundary

**Runtime policy.** Agent Common declares the next-turn budget, captured actor
limit, and mandatory `MaxToolRounds()`/`SetMaxToolRounds(int)` operations. The
setter accepts nonnegative values; zero stays zero publicly and resolves to 200
only at turn start. Negative runtime updates refuse without mutation. Chapter
5's immutable constructor configuration remains an initial seed, including its
existing library-only negative/unlimited contract. An active turn keeps its
capture; a queued turn captures when started. After the final admitted tool
batch, rendering adds a request-only instruction to answer without more tools.
It does not change History, dialogue or signed/opaque parts. Another tool batch
is paired with refusals before effects.

**Settings and ownership.** The optional GUI has a value-like presentation
Preferences record and declared load/merge/save/snapshot operations in Common;
implementation lives in Server/Settings.cs. No competing live budget is stored
there: snapshots read the Agent. The host's existing composition scope owns
preferences and serializes patch/save/publication under its short connection lock.
A same-directory temporary write and replacement complete before policy or live
preferences change. Failed writes send an error and leave the prior file and live
values untouched. This is replacement, not a power-loss/backup guarantee.

The default path is `$HOME/.ensemble/settings.json`, with `--settings PATH` for
explicit owned paths. One validator handles disk and wire: themes system/light/
dark; strict Booleans; finite speech rates clamped 0.25–4; integer fonts clamped
12–32; nonnegative Int32 budgets. Invalid/unknown fields retain prior values.
Malformed/unreadable/missing files use defaults without startup rewrite. The
protocol supplies `current_settings` on connect/get and complete
`settings_updated` broadcasts only after a successful save. All clients observe
the same serialized update order.

**Page behavior.** Settings includes theme, font, narration, speech speed,
speech-related pause and next-turn batch limit. CSS variables update existing
content and system theme follows media changes. Browsers retain acknowledged
settings, not localStorage copies. Controls may hold an unsaved edit while
saving; actual behavior follows acknowledgement, and an error restores controls.
Speech rate is read when constructing each utterance, so queued speech uses the
new rate without changing current audio. Disabling speech pausing releases that
cause without cancelling audio or overriding typing pause. Narration reads the
acknowledged preference, not the editable checkbox. Existing queue cleanup,
explicit cross-tab unpause, and silent visibility wakeup remain in place.

**Evidence and corrections.** Scratch is `/private/tmp/ch08-csharp.QgNW4a`
(text pointer `/private/tmp/ch08-csharp-current`). `budget-tests.log` has 36
passing component groups, exit zero, including held real HTTP, a real counter
tool, current/next-turn effects, negative/zero policy and instruction isolation
in all three renderers. `gui-build-final.log` has zero warnings/errors. The
settings exercise (`exercise-settings.py`, `settings-exercise.log`) exited zero
with two real WebSockets, validation/clamps, concurrent patches, new connection,
actual restart, invalid/malformed disk input and forced write-failure rollback.
Its temporary filesystem obstruction preserves prior evidence, not shared data.

`grade-final.json` parsed as passed:true, 80/80; stderr is empty and job 859
exited zero. Chapter8Executable extracts the existing four settings checks;
inherited Chapter 7/Go topology is excluded, not awarded. The reader accepts the
published `settings_updated` as well as its historical `settings_changed` label;
product protocol was not changed to match that old fixture. Report wording now
says acknowledged update. Grader package checks were compile-only (`-run '^$'`),
not claimed behavioral tests. Three incompatible flag combinations exited 2.
The final browser-only corrections did not change this verified server protocol.

The isolated real-browser exercise (`exercise-browser.mjs`,
`browser-exercise-7.log`, job 871 exit zero) used installed Puppeteer/Chrome and
owned profile/ports. It operated actual controls and verified two-tab updates,
system-theme changes, font/zero budget, reload, queued speech rates, pause-toggle
and error cleanup, narrow layout, and visible rollback after a real failed save.
The owned host exited zero. Speech was a working fake, not an audible assessment.
Initial failures included a speech fake missing addEventListener, a wrong textarea
selector/state spelling, and an incorrect assumption that speaker buttons queue
rather than replace speech. The slider did reveal a product flaw: immediate
reset to acknowledged state disrupted repeated keys. It now retains the proposed
edit while saving. Subsequent reruns initially served old copied assets; only
the rebuilt final run verifies the fix. I also corrected narration's checkbox
source. Several exact-anchor scripts refused before saving or testing; an early
non-fail-fast chain ran tests despite edit failure. Those runs earn no credit for
unwritten changes. These were my fixture/editing mistakes, not infrastructure
failures. The final build/browser and byte checks cover the saved artifact.

**Endpoint and limits.** `ch08` contains 71 byte-verified source/support files
matching main; earlier endpoints were untouched. Prose and READMEs teach the
actual C# API, validation, ownership and persistence order. Whitespace/JavaScript
syntax checks passed. Chapter 7's unchanged streaming/transport/lifecycle evidence
is carried, not rerun. No provider calls, installation, restart of Ensemble,
commits or deployment occurred. Real-provider programming task, audible speech
quality, and unrestricted human accessibility remain incomplete. Last observed
free space was 16,663,508 KiB, above the 1 GiB floor. Next: Chapter 9.

## Chapter 7 completed offline boundary

**Core facts and ownership.** History now locks mutation/read boundaries and
returns one immutable events/context snapshot. Final response events persist the
actual request sequence and parser part IDs, including zero/noncontiguous IDs;
old logs retain deterministic replay-only identity without changing provider call
IDs or opaque data. Metadata validates before publication and survives numbering
and codec round-trip. `EventRecorded` is a lossy notification, never a substitute
for the durable fact. `ToolStarted` follows successful local admission and precedes
effects; malformed/schema/policy refusals do not falsely claim dispatch. Observer
options are stripped before execution-schema validation. Pause/resume and joined
`AgentShutdown` are recorded; shutdown is appended after Jobs drains, not merely
on request. Repeated shutdown joins produce one terminal fact.

**Optional application.** `Examples/Gui` is a separate ASP.NET Core shared-framework
application; core's explicit compile list does not include it or reference web
frameworks. Common declares GUI operations/value projections; Server implements
them. Hosting retains local bounded connection queues and transient partials, not
a second conversation owner. Replay registration and live queueing share a short
handoff lock; socket sends occur outside it. Slow clients disconnect. Periodic
snapshot catch-up repairs missed finalized observations. Live projection is bounded
by the observation's sequence, avoiding publication of later finals ahead of queued
deltas. A reconnect that has final content does not resurrect older partials.

**Presentation and controls.** Stable-ID cards use safe DOM Markdown, lexical fenced
code highlighting, collapsible complete thinking/tool content, exact-replacement
edit previews, and readable ANSI-stripped results. Full speaker payloads do not
come from clipped DOM. Automatic narration queues new assistant text/thought
clauses and short tool summaries, never replay or tool results. Its current/pending
queue settles on both completion and error; cancelled callbacks cannot restart it.
Typing or queued speech sends pause/unpause only on cause edges. Another tab's
explicit unpause wins while a cause is unchanged. The silent visibility wakeup is
outside that queue. Wire uses `unpause`, both interrupt modes, and joined shutdown;
shutdown disables new-work controls while retaining the transcript. Four private
logs, explicit flags, loopback binding, Host/origin checks and input limits are
implemented. This local UI is not a tool sandbox or remote deployment.

**Tests and actual use.** `/private/tmp/ch07-csharp.b8F8Og` is the owned scratch
(the `/private/tmp/ch07-csharp-current` pointer is text). Its `offline` wrapper
uses the approved SDK, installed tools, owned home/tmp and disabled package network.
`final-component-tests.log` has **35 passing groups, zero failures**; job 798 exited
zero through components, GUI build and actual protocol exercise. The new groups
cover metadata, legacy compatibility, concurrent snapshots, truthful admission,
replayable controls and idempotent joined shutdown. Existing 33 groups pass.
`gui-build-final.log` has zero warnings/errors. Main source and scoped diff whitespace
checks pass; JavaScript syntax is checked. No earlier HTTP-lifetime campaign was
repeated.

`exercise-gui.py` / `gui-exercise-final.log` ran the real host against gated loopback
HTTP and real WebSockets: meaningful pre-EOF delta, reconnect while partial,
reconnect after final, stable IDs, paused scheduling with accepted hints, refused
and real file-tool effects, both interrupt modes, reuse, joined shutdown/refusal,
and foreign-origin rejection. The owned host exited zero; work directory is named
in that log. `exercise-browser.mjs` / `browser-exercise-final-2.log` used installed
Puppeteer with an isolated headless profile, not Bill's browser. It verified real
rendering, paused tool plan/no effect before resume, actual tool effect, replay
identity/no duplicates, safe highlighted code with preserved source, full human/
thinking/tool/result speech payloads, keyboard hint, narrow-layout geometry,
pre-EOF narrated speech queue, interrupted-utterance release, cross-tab unpause,
and silent visibility wakeup. Speech was a page-boundary fake, not audible evidence.
The final browser job 810 exited zero; the later change only expands safe malformed
numeric-flag handling, so that valid-path evidence is carried explicitly.

**Grader scope.** Existing `-executable-only` now supports Chapter 7 with an explicit
static directory, reusing the four existing GUI evaluators (80 points). The full
Go route retains inherited Chapter 6 and dependency checks. `grade-final.json` is
parsed **80/80, passed:true**, empty `grade-final.stderr`, job 814 exit zero.
`grade-final-readable.log` retains the preceding readable report; its `.json` name
was corrected rather than pretending it was parsed JSON. `flag-refusals.log` has
four invalid grader combinations returning 2 and three malformed/out-of-range GUI
ports returning safe exit 1. The focused grader package command exited zero;
explicit C# executable grading, not an inferred historical Go test run, supplies
the behavioral credit.

**Corrections and limits.** First scoped grade was 60/80: I exposed `resume` instead
of the published GUI `unpause`. Source review also caught snapshot look-ahead
suppressing queued live deltas, ignored interrupt mode/missing shutdown mapping,
and an invalid-port exception filter edit that had not matched. Those are corrected
and tested. I initially referenced a nonexistent shutdown enum; the final design
records a declared, codec-supported fact only after the actual drain. Initial
control validation wrongly retained observation options; regression tests exposed
it. Fixture failures included expecting an active interrupt for an idle paused
queue, reusing an existing output file, clicking a collapsed speech option,
expecting a shortened terminal reason, and RAF polling in a background tab despite
correct Idle delivery. Their evidence is retained. Repeated guessed edit anchors
were my process error; partial saves were inspected, not treated as atomic success.

`solutions/edition-3/ch07` contains **68 byte-verified source/support files**, excluding
bin/obj, logs and browser/runtime artifacts. Chapter 7 and the endpoint/GUI READMEs
teach the actual C# contracts and reasons. Real-provider acceptance, audible speech,
and human visual/accessibility evaluation remain incomplete. The scoped grade does
not establish them, restart restoration, or the excluded Go topology. No installation,
Ensemble restart, commit, push, deployment or shared cleanup occurred.

## Chapter 6 completed offline boundary

**Model and ownership.** Common declares value-like chunks, parsed response/final
IDs and per-parse native assembly state. Engine has one HTTP-response `ParseAsync`
selecting JSON or SSE and reuses completed-body decoders for canonical semantics.
The streaming callback is transient per invocation, not a stored service. Workers
post deltas/completions to the existing mailbox; only the actor mutates History or
schedules. Deltas are not persisted. Final events are recorded before final
observations. Parser IDs, including zero, are scoped by the actual request sequence;
consumers never infer final IDs from array positions or global counters.

SSE handles comments, CRLF and multiline data. Anthropic block boundaries and
signatures, OpenAI indexed tool/name/argument fragments and final usage, and Gemini
explicit part/thought/call/signature boundaries preserve opaque data and exact JSON
numbers. Missing terminal frames, unfinished blocks, malformed required fields,
provider errors and data after completion cannot authorize partial tools. Strict
explanation policy uses actual arrival order: later text does not excuse a tool
fragment that arrived first. Both interrupts preserve independent jobs and permit
reuse of the Agent. Stream capability is explicitly separate from media capability;
unknown models stay plain, and immutable configuration can disable streaming.

**Display and lifetime.** Plain delivery has authoritative finals and zero synthetic
deltas. CLI live JSON exposes parser `part_id`/`response_id`, distinguishes human
and job parts, flushes each line and avoids duplicate assistant answers for requests
with reliable replies. EOF captures usage before destruction, then drains the closed
observation queue. Interactive chat displays all chunk kinds on stderr (ANSI only
for a terminal) and writes the reliable completed answer once to stdout. This differs
deliberately from the old Go stdout preview: a lossy queue must not be the sole source
of redirected answer text. `-no-stream`, `EN_DISABLE_STREAMING=1` and optional
`CH06_API_LOG` are documented. Root trace access follows actual parents, redacts the
configured key before writing, and does not own/dispose the caller's writer.

The separate `Examples/Observer` assembly has application-local delivered-event
counters, UTF-8 byte counts, response-scoped final identities, timing and measured
History usage. Reliable completion and a drained queue precede its one metrics
record; counters are not billing or lossless telemetry. `Examples/Workflow` now
uses true deltas/parser IDs for author/editor/reviewer, not Chapter 5 synthetic
whole-part deltas. Reliable tasks alone advance stages. Root destruction drains
queued finals. Chapter prose and README now teach these C# contracts, reasons,
local exercises and limitations rather than nonexistent Go interfaces/CLI commands.

**Verification and executable use.** Owned scratch `/private/tmp/ch06-csharp.pNDEpf`
(text pointer `/private/tmp/ch06-csharp-current`) has its own controlled offline
home/tmp and installed SDK/Go cache access. No paid requests or installations.

- Final library components: **33 passed, zero failed, exit 0**, job661,
  `final-component-tests.log`. All-vendor parser fixtures include exact arguments,
  opaque/signature parity, missing fields/terminal frames, and no plain callbacks.
  Real Agent tests hold HTTP before completion, prove no partial History/effects,
  stored-before-published finals, trace redaction, strict tool refusal and both
  interrupts with actor reuse. Earlier milestone job631 also passed 33 groups but
  predates the required-field hardening; it is not substituted for this final run.
- New `-ch 6 -executable-only -stream-exercise EXECUTABLE` route reuses the existing
  fixture and checks: **80/80, passed:true, exit 0**, job661,
  `grade-second.json`, empty `grade-second.stderr`. Checks cover Anthropic actual
  text/thinking/tool deltas, matching finals, exact tool execution, measured usage,
  unknown-model fallback and external aggregate reporting. It explicitly excludes
  inherited Chapter 5 and Go consumer/topology checks. The normal Go route retains
  its synthetic plain-delta contract; C# explicitly checks final-only plain delivery.
  Relevant grader tests passed (`grader-focused-tests-2.log`); five missing-observer/
  incompatible-flag cases returned exit 2 (`flag-refusals.log`). No full Go pass claimed.
- Actual JSON CLI gate: `exercise-stream-cli.py`, `cli-exercise-3.log`, saved
  `cli-stream`/`cli-plain` output. Meaningful text is visible before fixture release;
  IDs, sanitized trace, EOF drain, final-part/usage parity and forced-plain mode pass.
  Carried for unchanged valid-frame JSON behavior; later parser changes concern
  malformed required fields and were tested above. Job628's CLI phase passed but
  its combined command exited 1 on an obsolete component assertion; never label
  that whole command successful. Corrected milestone job631 exited zero.
- Actual separate workflow: `exercise-workflow-stream.py`,
  `workflow-stream-exercise.log` (exit 0). Gated pre-EOF visibility, response-local
  text/tool identity, all six manuscript tools, real busy-think hint with running
  job report, undispatched batch tail, three reliable stages, tagged output,
  separate histories and EOF join passed. Valid-frame evidence carries through
  required-field hardening; component fixtures test the changed invalid cases.
- Actual chat: `exercise-chat-stream.py`, `chat-exercise.log` (exit 0), final
  `chat-build.log` build without warnings/errors. A real executable preview is
  observed on stderr before final frames are released, stdout is exactly the
  reliable answer, redirected stderr has no ANSI, trace is redacted, exit is clean.
  This is not evidence of human visual or auditory usability.

**Corrections retained.** Earlier scripts missed wrapped declarations/anchors and
some test helpers used guessed signatures; those failed setup/build steps earned
no behavioral credit. Signature tests initially expected object wrappers where the
existing canonical text/call contract uses scalar signatures. Inspection caught
an initially missed actor arrival-order condition. The first CLI gate mistakenly
stopped on a legal empty start chunk. Its next run exposed synthetic plain deltas;
removed those callbacks and corrected the test that had endorsed them. Later source
review replaced null-forgiving required-field reads with explicit format failures
and negative tests. The first attempted grader command was too broad and failed on
absent historical Go solution directories before grading. First actual scoped
grade was **45/80, exit 1**: its untyped `text` fallback misclassified typed opaque
finals. Typed part data now takes precedence, with legacy fallback only for untyped
records. This was a fixture adapter correction, not a product alias. Failed reports
remain in scratch; the final 80/80 applies to the corrected route and artifact.

**Preservation and limits.** `solutions/edition-3/ch06` contains **58 source/support
files**, byte-verified against `main`, excluding bin/obj and runtime data. Earlier
endpoints remain untouched. Scoped whitespace passed. Last free space check was
17,072,656 KiB, above the 1 GiB floor. No Ensemble restart, deployment, commit/push,
new installation, public-provider call or unrelated cleanup occurred. Local fakes
prove protocol/component behavior, not provider availability, subscriptions,
media decoding, opaque acceptance or current live model streaming. Those real
integrations remain incomplete. Next: Chapter 7, after the required handoff.

## Chapter 5 completed offline boundary

**Model and ownership.** One Agent-owned mailbox replaces the synchronous loop.
Common declares persistent state; Agent implements orchestration. Engine accepts
frozen request bytes and returns immutable unnumbered events, without History
access. Only the actor records conversation facts and schedules new model work.
Each request has its own reliable task. Bounded Agent/root progress is lossy and
shared-consuming, not broadcast. Input observations include accepted/refused
outcomes even for invalid input and closed admission. Root envelopes retain the
actual Agent; optional unique names use its real child collection, not a registry.
Caller-provided immutable primary instructions survive tool installation.

**Controls and lifetime.** Busy hints stop foreground observation only, return a
real Job snapshot, pair exact-ID undispatched tails, then inform the next request.
Both interrupts preserve independent Jobs and the Agent. `after_tool` retains the
current boundary. HintReceived/TurnInterrupted/TurnEnded are replayable facts;
no invented reducer state is needed. Host results use ToolObserved audit facts,
not orphan ToolReturned dialogue. Four handle-free controls remain available
while busy/paused. Positive batch limits count batches; zero defaults to 200;
negative is unlimited. The model may consume the last permitted batch's results.
Shutdown closes admission, collects workers and joins Jobs before graph teardown.
Managed callbacks must cooperate; .NET cannot forcibly end one ignoring its token.
`FlushAsync` waits for foreground quiescence, not independent Jobs, does not resume
paused work, and faults rather than crediting work discarded by shutdown.

**CLI, media and consumer.** Legacy requests queue ordered reliable replies while
stdin remains responsive. Live prompt/hint/pause/resume/interrupt/tool/shutdown
map to the actor. EOF flushes then shuts down; explicit shutdown cancels. Structured
parts override fallback text. Media checks exact vendor/model declarations before
file reads/HTTP; local bytes enter requests only, locators remain in History.
Unknown models may send text; URI/handle attachments stay explicitly deferred.
The prior handoff's broader media and send_message spellings were not the saved
chapter contract. The external `Examples/Workflow` assembly builds author/editor/
reviewer peers with real role tools and one synchronized value-like manuscript.
Its reliable pipeline awaits completion at EOF; tagged observations are display.
The detailed chapter and README now teach the actual C# API, not Go-only sketches.
Historical Go demonstrations remain historical, not C# provider evidence.

**Verified evidence.** Owned scratch `/private/tmp/ch05-csharp.PPnEI5`; pointer
`/private/tmp/ch05-csharp-current` is a text file. All process exits below were
collected. No earlier lifetime campaign was repeated.

- Final **30 component groups, zero failures, exit 0**, `final-tests.log` (job555).
  Held real HTTP/Job tests cover freezing, hints, both interrupts, paused queued
  work, stale effects, reliable replies, cancellation, shutdown joining, names,
  input refusal, unlimited batches and replay. The final flush test saturates both
  unread progress queues while HTTP is held and still obtains the reliable answer.
- Actual CLI: **eight scenarios passed, exit 0**, `cli-exercise-2.log`: queued
  replies/EOF, busy HTTP hint, pause/resume, both interrupts/reuse, direct tools,
  image Parts precedence/locator logs, malformed-input cleanup, explicit shutdown.
- Scoped executable grade: **25/25, passed:true, exit 0**, `grade-ch05.json`, empty
  stderr (job547). Three recorded request-prefix renders matched captured bytes;
  image projections and capability refusals passed. New `-executable-only` retains
  unchanged replay/media checks and excludes Go consumer/workflow/topology and
  Chapter 4 parity. Full Go route remains intact. Five incompatible flag cases
  returned 2. `go test ./cmd/grade-edition2` had no tests: compilation evidence only.
- Separate C# consumer build and real workflow: **exit 0**, `workflow-build-2.log`,
  `workflow-exercise.log` (job542). Eight captured requests exercise all six role
  tools, exact toolsets/prompts, a hint during real think, running report, untouched
  batch tail, stored author/editor text, actual reviewer decision, three reliable
  pipeline markers and separate logs. Actual output and requests are retained in
  `workflow-exercise/`, not Git. No internal API or copied actor loop is used.
- Media components cover all three image dialects, Gemini PCM WAV/PDF projection
  and representative MP4 container bytes. The video fixture does **not** prove a
  decodable movie. Capability tables are renderer declarations, not live tests.

**Corrections retained.** Early edits left direct dispatch outside the actor;
legacy limit fixtures exhausted their sixteen replies under the new default;
changed explain-first wording unnecessarily broke a fake and was restored.
Real tests exposed optional-schema `properties` bugs in rendering and validation,
and my orphan host-tool result mistake; production fixes preserve strict pairing.
Negative limits needed both constructor and scheduler corrections. Several exact
edit anchors stopped after partial saves; files were inspected before continuing.
The first CLI exercise used wrong environment names, so no HTTP ran. Consumer
compile initially used Part instead of actual PartFinal. The first grader-refusal
assertion over-specified which incompatible flag's diagnostic wins. These were
fixture/setup mistakes, not passing tests. I initially called the scoped grade
30 points; its actual existing weights sum to 25. Failed logs remain in scratch.

**Endpoint and limits.** `solutions/edition-3/ch05` is a byte-verified source-only
copy of main, including the external consumer, tests and support files; no bin/obj,
credentials, captures or diagnostics. Earlier endpoints are unchanged. Real
providers remain incomplete (allowance exhausted), PTY remains macOS-specific,
and no runtime installation, restart, commit or publication was performed.
Next: Chapter 6 streaming, using this endpoint as the baseline.

## Chapter 4 completed offline boundary

Chapter 4 adds Agent-owned Jobs and Jobs-owned Job, declared in Common with complete
two-way operations. Ensemble allocates handles; each Agent's Jobs restricts lookup
to its own children. The single dispatch path consumes observation settings,
reserves an ordinary job, records ToolCalled, then permits effects. Controls have
no job of their own. Custom tools are concrete Tool children with the callback
required by the external-consumer exercise, not a callback layer for built-ins.
Callbacks receive the actual Job and reach services through parents.

Output is flushed to a full file before notification; observers share an unseen-byte
cursor and receive bounded head/tail text. Fast completed jobs return their whole
text. `tool_limits` sets only the next call's wake delay, regex and inline byte cap,
including controls and failed calls; it never sets a task deadline or hard output
cap. Result redaction retains the file locator. Call-recording failure prevents
tool effects. The selected-model explanation refusal also follows this dispatch
path. Unexpected handler faults propagate through observation/completed cleanup
rather than becoming routine tool errors; optional JSON Schema `required` works.

The macOS implementation uses installed `/usr/bin/script` for a real 50×200 PTY,
newline submission, process-group kill and drain. CRLF normalization happens only
at that boundary, before both disk and reports. Generic tool bytes stay exact.
Settled writers and process handles close promptly. Shutdown records JobKilled
before detaching children/History. Managed cancellation is cooperative: killed
callbacks cannot publish late output or resurrect status, but .NET cannot forcibly
abort an uncooperative callback. No Linux/cross-platform PTY claim is made.

Final evidence is in `/private/tmp/ch04-csharp.XNZCOc` (pointer
`/private/tmp/ch04-csharp-current`). Job 361 completed the full sequential campaign
with exit zero. `tests-final7.log`: **17 groups passed** (real files, PTY geometry,
input, group killing, observer cancellation, cursors, one-shot consumption,
custom tools, cross-Agent isolation, logging-before-effects, stable killed state,
closed writers, redaction, unexpected-fault propagation and existing regressions).
`grade-jobs-final5.json`: **100/100, passed:true**, empty stderr and exit zero.
It includes executable Chapter 3/2 compatibility. Prior Chapter 1 lifetime evidence
is carried for the unchanged HTTP timeout/cancellation path, not rerun.

The existing Edition 2 command now has an explicit Chapter 4 `-jobs-only` route,
reusing its 100-point job checks without the Go-only 60-point consumer adapter.
Invalid chapter/consumer combinations are refused; the original full route remains.
A narrow fixture correction asks the final input wait for its exit marker, since
an unpatterned wait may correctly return on fresh `ECHO_BYE` before exit. Relevant
grader-package tests/build and scope checks passed. This is not a fabricated
160/160 C# grade. `consumer-spin-final4.log` separately verifies the built external
C# assembly: custom/file/shell jobs, two real HTTP rounds against a loopback fake,
correlated results, parent-path timestamped logging and exit zero.

`cli-spin-final4.log` records personal use of the actual CLI: delayed job and wait,
real dlv breakpoint/continue/next/print 42/quit, normal debugger exit, EOF killing
a held job with a shutdown fact, and key-free replay. The debugger first attempted
to print the variable before scope; its script was corrected from the actual
transcript. This is real local tool/application evidence, not real-model judgment.
The provider-dependent coding exercise remains **incomplete** under the exhausted
allowance. No installation, Ensemble restart, commit, push or publication occurred.

Important corrections retained: I initially misread `tool_limits` as deadlines and
hard caps, and the first tests endorsed that wrong contract. Those mechanisms and
tests were replaced, not credited. Early grades exposed report shape, control
metadata, newline and cwd errors; a 75/100 scoped grade became 100 after recording
explicit cwd as an absolute path. Final source review removed an overly broad
worker catch, exposed/fixed the optional-required validator bug, normalized PTY
bytes at the correct boundary and closed settled writers. Early syntax/anchor and
test-fixture mistakes were setup failures, not product passes. Earlier logs remain
in scratch. The final source-only `ch04` endpoint contains **36 files**, refreshed
and byte-compared after the final campaign. Scoped whitespace checks passed;
free disk remained about 16.8 GiB. Next: Chapter 5, not another Chapter 4 campaign.

## Chapter 3 completed offline boundary

Chapter 3 adds six concrete Agent-owned Tools, with state and callable operations
in Common and complete parent add/remove methods. Agent runs the synchronous loop;
Engine still performs one exchange; History remains the authoritative event record.
Each immutable Tool definition supplies provider declarations and strict argument
validation. No global registry, service bag or callback layer. Tool destruction
removes both links; Agent cascades. The 28 source/support files in `ch03` were
byte-compared with main, excluding bin/obj.

The chapter and README teach the C# progression. Retained abandoned-revision lessons:
nonmutating Gemini schema projection, exact edit refusal, and an explicit selected
model explanation policy. Calls-only batches under that policy produce paired errors
without effects and can retry; no model-name guessing. Call IDs normalize on capture;
duplicate IDs within a response refuse before dispatch. Gemini text/call signatures
preserve exact provenance through serialization. Reads support line selection and
UTF-8-safe byte caps; search merges context windows; commands use fresh shells,
separate stdout/stderr, real exit codes and credential removal. Sixteen rounds bound
repetition without a generation timeout. Narration goes to stderr before effects;
JSON stdout's assistant field holds only the final answer.

Evidence: `/private/tmp/ch03-csharp.mM7Wia` (pointer `/private/tmp/ch03-csharp-current`).
`tests-final.log`: 16 groups passed, including real file/shell operations, cancellation,
Tool relationships, real Agent HTTP loops with mixed results, explanation retry,
signature/ID projection and round limit, plus Chapter 2 regressions.
`build-sixth.log`: successful publish. `grade-second.json`: Chapter 3 100/100,
passed:true, actual exit 0; stderr empty. `grade-ch02.json`: Chapter 2 compatibility
100/100, passed:true, actual exit 0; stderr empty. Existing genuine delayed-header/body
evidence is carried for the unchanged HTTP path, not rerun for timestamps.

I separately used the actual CLI with the existing serve-only fake on owned loopback
ports and disposable projects for all three dialects (`cli-spin.log`, `spin-*/`).
Two turns per dialect produced six traced requests, four recorded tool results,
normal EOF and key-free dump/replay containing actual file content. Owned fakes stopped.
This proves local plumbing and execution, not provider acceptance, model judgment or
a live coding task. Paid-provider and real coding-work exercises remain incomplete.

Initial `grade-first.json` scored 50/100, exit 1: missing Gemini result IDs,
intermediate narration leaking into the final reply, and inadequate edit-refusal
context were product defects subsequently corrected. The clean wrapper also omitted
Go required by the grader's shell fixtures; adding the already-installed toolchain
and existing caches with downloads disabled fixed that setup failure. Component tests
caught incorrect ToolCalled/ToolReturned payloads before passing. My test fixtures
had compiler/signature/sequence errors; failed logs remain in scratch. Guessed edit
anchors and filenames wasted work; subsequent reads used actual source. Final source
inspection caught the max_bytes/defaults and separate-stream requirements independently
of grading. Snapshot preflight initially omitted NuGet.Config and .gitignore, both
legitimate support files; no endpoint was written until the preflight passed.

Scoped whitespace checks passed. No new dependencies, grader edits, installation,
restart, commit, push or publication. Real-provider allowance remains exhausted.

## Chapter 2 teaching boundary

The shared vocabulary now explains value roles, parent paths, operation contracts,
and failure behavior. Constructors, validation, cleanup, parser steps and JSON
construction are expanded. The chapter starts from the actual C# Chapter 1 and
explains replacing its completed-pair list with recorded attempts. It correctly
places cumulative usage in History's reduction, not a second Engine counter.
No new feature or service layer was introduced. The 23 source/support files in
`ch02` were byte-compared with main; no bin/obj outputs were copied.

All 15 real-component groups passed; the GUI construction consumer exited zero;
final publish succeeded; the existing Edition 2 Chapter 2 grader scored 100/100,
`passed:true`, exit zero. The final executable's actual local CLI exercise passed
three-turn chat and normal EOF on all three protocols, key-free dump/cross-vendor
render, deterministic output, and ephemeral replacement/clearing. Twelve loopback
POSTs; no real-provider calls. Evidence: `/private/tmp/ch02-teaching.us0S9m`
(`tests-final.log`, `gui-stub.log`, `build-final.log`, `grade-final.json`,
`grade-final.stderr`, `cli-spin.log`; pointer `/private/tmp/ch02-teaching-current`).
The trial's genuine delayed-header/body evidence is carried for the unchanged
HTTP timeout/cancellation path, not rerun for freshness. The new grade covers
ordinary Chapter 1 parity. Scoped whitespace checks passed.

Setup corrections: I first invoked `grade-edition3 -ch 2`, which has no Chapter 2
route and exited 2 with usage text (`wrong-grader.stderr`); it earned no behavioral
credit. The correct Chapter 2 grader is `grade-edition2`. Several exact-text edit
anchors refused guessed shapes; I inspected actual source and corrected the edits
before the final tests. These were my editing/setup errors, not product failures.
No deployment, installation, commit, push or publication occurred.

## Chapter 1 completed offline boundary

Chapter 1 is implemented and offline-verified at `solutions/edition-3/ch01`.
Its manuscript now teaches the C# model, staged pairs, whole-line input and
request lifetime, with historical live quotations labeled as historical.
Real-provider exercise remains incomplete; this is not full live completion.

Evidence: `/private/tmp/ch01-csharp.9EiEwl` (`tests-corrected.log`, `build.log`,
`grade.json`, `grade.stderr`, `cli-spin.log`, `input-read-error.log`). Four
component-test groups passed. The standalone executable—not Chapter 2—passed
100/100, `passed:true`, exit 0 including separate genuine 62-second header and
body waits. The actual CLI was exercised for three-turn chat, cumulative usage,
EOF, 100,000-character input, an unterminated last line, malformed input and a
real stdin read error. Final source inspection checked parent paths, complete
two-way operations, staged history, safe errors and no generation timeout.
Subsequent source changes were teaching comments only. No grader modifications.

One test exposed my incorrect assumption that `InvalidDataException` was caught
by `IOException`; the corrected explicit filters passed. The initial failure
remains in `tests.log`. The first SDK build omitted the owned CLI home and
first-run suppression; .NET reported creating an HTTPS development certificate.
No trust command or certificate-store cleanup was attempted. Later SDK commands
use owned HOME/CLI_HOME and disable certificate generation. This unintended
first-run side effect is not authorization for additional certificate changes.

## Chapter 2 trial baseline

The original Chapter 2 trial preceded the teaching pass above. Its 15 focused
test groups, Chapter 2 grade (100/true/0),
Edition 3 Chapter 1 grade (100/true/0 with real 62-second header/body phases),
and actual local-fake CLI exercises passed. These results do not validate a
separate Chapter 1 solution or later edits. Real-provider work remains blocked
by the exhausted allowance. Evidence is at `/private/tmp/ch02-csharp.8iRYR2`.

Bill requested the teaching improvements now completed above. The saved C# skill
states that standard explicitly. Complete parent add/remove behavior remains
covered by direct regressions. His authorization to continue is not blanket
acceptance of all later chapters.

