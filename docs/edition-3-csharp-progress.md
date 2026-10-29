## Chapter26 — assembled review complete; source-only endpoint preserved

Chapter26 Gateway is **locally complete** at `solutions/edition-3/ch26`. The book is
**not fully verified**: Chapter16 live OAuth and Chapter24's bounded live cache
exercise still require Bill's approved credential-store/setup and participation.
Chapter27 Goals and Cadence's epilogue were already completed; do not reopen them.

- Personally reviewed the complete 81-file added/changed source/support delta
  versus ch27, alongside the rewritten manuscript and grader changes. Inspected
  Common declarations, reciprocal add/remove postconditions, live policy,
  persistence validation, authenticated replacement, current-session actions,
  serialized socket writes, queue receipt frontiers, joined cleanup, adapter
  admission/result correlation and application startup/shutdown. Read tests and
  working boundary fakes, not merely their pass summaries. No new runtime defect
  emerged. Corrected stale README status/scope phrases; no production code changed.
- Verified all 31 vendored C# sources against upstream pin
  `4279cfa57df028dec4a4cb3ee9579a77a5cc39fa`: only the approved KeyPair import and
  Transport associated-data API exposures differ. License bytes match. Retained
  prior dependency/runtime checks; this comparison is not a new crypto audit.
  Unused upstream Rekey remains unexposed and outside the verified contract.
- Preserved **415 source/support files**, byte-equal to main, with **23 project
  references internal**, no symlinks or build/runtime additions. All **334**
  pre-existing non-README foundation files remain byte-equal to preserved ch27.
  Earlier endpoints and original Go grader/probe are unchanged. Preservation and
  scoped whitespace checks exited0 (job6635). No build artifacts or test logs were
  copied into the endpoint; no installation, restart, deployment or commit occurred.
- Reused valid unchanged evidence from `/private/tmp/ch26-csharp.HDZGCi`: host7/0,
  interviewer4/0, adapter6/0 and actual three-C#-process exercise5/0; session15/0
  then final affected6/0, local client6/0 and wrapper6/0, prior Mac/Pi crypto8/0.
  Their original logs/results and retained failures are detailed below. The actual
  Agent interview used a fake HTTP provider, not a live model. No new campaign was
  run for review or preservation, and no new all-suite result is implied.
- Executable scope remains **60/60**, four complete server rows C2/C3/C4/C7 using
  the independent Go peer. C1/C5/C6's forty points remain ungraded, not substituted
  with library-test credit. `grade-first.json` and retained two-launch artifacts
  remain the results. Public prose example compilation, six shell blocks, JSON and
  local-link checks from the preceding slice remain valid.

Next: mandatory chapter-boundary handoff, then only the human-dependent live gate.
Ask Bill for an explicitly approved credential directory and readiness to complete
login; never request secrets in chat or search private/default stores. Recheck the
current official public OAuth endpoint during that session and run the saved small
three-request cache plan. Do not assume the historical failure or add a workaround
without current evidence. Idle reminders should be disabled while waiting for that
setup rather than repeatedly attempting blocked work.

## Chapter26 — C# manuscript rewritten; assembled review remains

The complete saved Gateway chapter has been rewritten against the public C#
declarations and actual launch programs. Chapter26 remains **incomplete**: assembled
source/test/prose review, source-only endpoint and mandatory chapter handoff remain.
No production source changed in this prose slice. The joint Chapter16 OAuth and
Chapter24 current-endpoint/cache exercise still requires Bill's approved setup.

- Read the full manuscript and compared the preserved pre-reset revision. Retained
  Deephold/Puffin, boundary trust, handler-return rationale and the earlier live
  interview as attributed second-edition history—not a new C# live run.
- Replaced Go APIs/layout/launch instructions with public C# contracts, the five
  library/application projects, actual flags and Agent-home files. Explained the
  real ownership graph, current-session checks, sent-frontier receipts, joined
  shutdown and the difference between mailbox admission, History publication and
  completed model work. Documented periodic/private snapshots' crash window and
  lack of a cross-file transaction; no exactly-once or power-loss promise.
- Dependency prose retains the approved suite, pinned Noise.NET/libsodium,
  associated-data/private-key-import exposures, native-platform evidence and
  unused-rekey limitation. API-key-only live recipe is explicitly opt-in; it does
  not authorize credentials or replace Chapter24's public OAuth/cache exercise.
- Kept actual three-C#-process/fake-provider evidence separate from historical live
  model opinions and the 60/60 server-only executable scope; 40 points remain
  ungraded. Added a transparent flag-translation wrapper, not canned observations.
- Saved readback/static checks: six shell blocks parse, configuration and wire JSON
  parse, both local links resolve, tracked manuscript whitespace clean. Corrected
  inherited historical-evidence link and obsolete pseudo-key. Extracted public
  client example compiles with zero warnings/errors in `prose-example-build.log`
  under `/private/tmp/ch26-csharp.HDZGCi`; job6607 exit0. This is compilation, not
  a new client execution. Existing passing timer/crypto/application/grade campaigns
  were not rerun for prose. A pager accidentally invoked by `git diff --check` was
  stopped by its observed PID; subsequent Git checks explicitly use GIT_PAGER=cat.
- Review inventory `assembled-delta.txt` has 81 added/changed source/support files
  versus preserved ch27 (including README and pinned ThirdParty/Noise). No missing
  foundation files. This inventory is **not yet a completed assembled review**.
  Public declarations, launch roots, interviewer, Requests and Sessions were read
  for prose; full assembled source/test cross-check remains next. README's stale
  denial of adapter/grade evidence was corrected. No endpoint exists yet.

## Chapter26 — scoped executable server grader reconciled; prose/review remain

C# implementation, relevant tests and actual-application work are now followed by
an honest **60/60 executable server grade**. Chapter26 is still **incomplete**:
manuscript rewrite, assembled source/test/prose review, source-only endpoint and
mandatory chapter handoff remain. The joint OAuth/cache exercise remains deferred.
No production source changed during grader reconciliation.

- Fully read `internal/gradeedition2/ch26.go` and `ch26_probe.go.txt`, plus the
  original seven-row contract `internal/grade/ch27_checks.go`. Edition 3 Gateway
  uses the Edition 2 **`-ch 26`** route, not the historical Chapter27/Goals numbering.
- Added opt-in `internal/gradeedition2/ch26_executable.go` and its scope test,
  connected in `cmd/grade-edition2/main.go`. Existing Go route and probe source are
  unchanged. Invocation: `-ch 26 -executable-only -gateway-peer ABSOLUTE-GO-PEER-DIR
  -json SERVER-EXECUTABLE`. Independent peer is explicitly NOT the submitted C#
  public library. Historical welcome fields and unrelated consumer/GUI/Go flags
  refuse rather than silently change scope.
- Four **complete** original rows score: C2 provenance10, C3 authenticated welcomes
  (including empty/nonempty backlog)15, C4 group routing15, C7 child attenuation20.
  **40 points remain ungraded:** C1 Noise wrapper15, C5 complete client/delivery
  contract15, C6 outside public-client dependency boundary10. Independent C# tests
  and application evidence support these concerns separately; their results are
  not substituted for grader rows. Go crypto/client success earns no C# credit.
- Reused the FULL original exercise, including clean process restart and the
  held-handler/offline scenario, because C3's nonempty counts live inside that
  scenario. No weakened assertions or duplicate synthetic router. Fatal/unrun
  rows still fail; separate scope contract tests verify exclusions, weights,
  missing evidence and abort behavior. No source-shape credit/security-audit claim.
- `grader-contract.log` passes `^TestGatewayExecutableScope$`; `grader-build.log`
  clean, job6583 exit0. Eight prelaunch refusal cases retained in
  `grader-scope-refusals.json`. `grade-first.json`:60/60, all four rows pass;
  `grade-first.err` empty, job6585 exit0. No correction/regrade was needed.
  Artifacts `tmp/c26-executable-2028454503`: two real C# server launches, unchanged
  outside Go peer, generated test identities/config, state and queues. Final
  `state/queues/parent.json` is `{"next":8,"ack":8,"messages":[]}`. Expected wrong-key
  and abrupt-peer-close diagnostics remain in server logs, not hidden as defects.
- Scratch launcher `gateway-grade-host` only translates the historical `-config`
  and `-print-key` flags to the C# `--` flags and directly execs the already built
  pinned-SDK Host DLL. Only the outer grader uses `offline`; there is no double
  env-i stripping the grader's inputs. Explicit peer source is `solutions/ch26`.
- Saved grader diff reviewed; no runtime, core, crypto, CLI/GUI or preserved endpoint
  changes. Prior adapter6/0 and actual-three-process5/0 evidence remains below;
  no timer/crypto/application reruns for timestamps. Last free space9,639,388KiB.

Next: read the current complete `book/edition-3/chapter-26.md` and preserved revision,
rewrite against saved C# declarations while retaining Bill's rationale, then assembled
source/test/prose review. The current manuscript still contains Go reference code
and stale status (client-only). No Chapter26 endpoint or chapter acceptance yet.
Scratch `/private/tmp/ch26-csharp.HDZGCi`; no owned jobs/workers at this boundary.

---

## Chapter26 — Ensemble adapter and three-process interview verified; grading/rewrite remain

Implemented personally without workers. Chapter26 and the book remain **incomplete**.
No preserved endpoint changed; no installation, deployment, shared restart, live
provider/account access or credential-store read. Chapter16 OAuth/Chapter24 cache
exercise still needs Bill and an explicitly approved store/setup.

- Added optional `Gateway/Adapter/Ensemble.Gateway.Adapter.csproj`, with vocabulary
  in `Common/Bridge.cs` and bodies in `Bridge/Bridge.cs`. Neither core nor public
  Gateway references this adapter. A single-use invocation borrows the caller's
  Agent across an explicit application boundary (like borrowed GUI hosting), owns
  its public Client lifetime and reply observers, and clears borrowed references
  after joined teardown. The caller retains Agent shutdown/destruction authority.
  No second model scheduler, authority store or lossy observation subscription.
- **Correction to the previous handoff:** actual Common is `Owners.cs`/`Actor.cs`;
  `Application` is an alias for `Ensemble`, not another root type. Current public
  admission is `Task<string> SubmitAsync(IEnumerable<Part>, CancellationToken)`.
  It admits to the actor mailbox synchronously; it does NOT promise synchronous
  History commit. The adapter passes stamped content as ordinary `TextPart` input,
  acknowledges only after admission returns, and correlates each reliable result
  with its own immutable envelope. Only `ResponseMode == "async"` sends a reply;
  `RespondAsync` preserves conversation/source and requests no recursive response.
  Failed turns/replies are logged through parent diagnostics, not sent as raw error
  text. The actual saved Go adapter logs errors and has no child-alias registration;
  previous handoff descriptions of those features were stale and were not copied.
- Readiness means authenticated setup plus accepted contactability. Failed startup
  settles readiness and the joined run with the same failure. Disconnect cancels
  reply observation, not admitted actor work; late actor faults remain observed.
  Unexpected disconnect fails the run. No hidden reconnect or persisted receipt.
  A receipt means mailbox acceptance, not completed/durable inference: process
  failure can lose admitted work or repeat it. No exactly-once execution claim.
- Added `Gateway/AgentHost/GatewayAgent.csproj`: composition root owns Ensemble,
  named Agent home, optional explicit skills, persistence and SIGINT/SIGTERM. API
  key funding only, announced through existing `FundingDescription()` before
  inference; model explicit, key from named environment variable, no plan fallback.
  No skills/tools are activated without `--skills`. Host joins bridge, shuts down
  Agent, saves after successful initialization, then destroys the application.
  Ordinary errors log before diagnostic ownership detaches; destruction errors
  expose type only. Saves use `agents/<id>/session.save.json`, events `events.jsonl`.
- Six real-component adapter groups pass in `adapter-tests-2.log`, clean build in
  `adapter-tests-build-2.log`, job6570 exit0. Real Agent, Router, Noise/WebSockets,
  public clients; only provider replaced by loopback HTTP fake. Covers busy-first/
  queued-second correlation and receipts before completion, exact queue reply
  counts, no-response/error behavior, caller and remote disconnect without killing
  admitted work, closed-Agent refusal before ack, and failed-pin readiness/join.
  Earlier four-group result `adapter-tests-1.log` is retained, not additional credit.
- Actual **three C# processes** (Router, AgentHost, Interviewer) passed five checks
  in `agent-exercise-2.log`, job6568 exit0 after clean `agent-host-build-4.log`.
  Probe `agent-exercise/main.py`; result
  `tmp/gateway-ensemble-3y05xdlp/result.json` (5 passed/0 failed, 9 provider requests).
  Stable key-print without provider config; four actor requests using an explicitly
  selected local primary skill; exactly correlated report; joined save; restart
  restoring old answers into the next actual request; SIGTERM during held model
  HTTP; final-save obstruction producing nonzero exit after teardown. All local
  fake-provider evidence, NOT live interview quality, provider availability or
  OAuth/cache behavior. No repeat of unchanged host thirty-second timer campaign.
- Retained real host defect: `agent-exercise-1.log` / result
  `tmp/gateway-ensemble-q9pgr53p/result.json` failed startup with **Load before opening
  History writers**. Corrected startup to recover persistence before opening event
  log, matching existing CLI. First host build's five guessed signatures were my
  editing errors, not behavioral failures; corrected from saved declarations.
  Source review caught post-destruction diagnostic masking risk before exercise.
- New adapter/host/tests read back as a bounded slice. Full chapter assembled
  source/test/prose review remains. Core/CLI/GUI, cryptography and Gateway library
  sources unchanged. Prior host7/0, interviewer4/0, session15/0 then final6/0,
  client6/0, wrapper6/0 and Mac/Pi8/0 evidence remains separately scoped below.

**Next:** inspect the actual Gateway grader route and assertions, select an honest
C# executable scope, then manuscript rewrite, complete assembled review, source-only
endpoint verification and mandatory chapter handoff. No Chapter26 grade exists yet.
Scratch `/private/tmp/ch26-csharp.HDZGCi`; no owned jobs/workers remain at this slice.

---

# Chapter26 — standalone host and interviewer verified; Ensemble adapter remains

`Gateway/Host/GatewayHost.csproj` is now the standalone ASP.NET composition root.
The public Gateway library remains independent of ASP.NET and Ensemble. The host
supports `--config PATH` (default `gateway.json`) and `--print-key`; printing the
public key does not imply listener readiness. It resolves the configured host:port
explicitly rather than allowing a hostname to become Kestrel's wildcard listener,
uses ten-second header limits, and upgrades only the known exact `agent_id` route.
The caller awaits Router.AcceptAsync through socket teardown. Standard console
lifetime handles SIGINT/SIGTERM; a host-owned PeriodicTimer saves every30 seconds,
reports failures and retries the next tick. Shutdown begins HTTP stop, joins the
saver, joins Router/final save, then completes HTTP stop/disposal. Startup and cleanup
failures remain visible together. No second policy/session registry was introduced.

`Gateway/Interviewer/Interviewer.csproj` is an independent public-client executable,
with declarations/state in its own Common and bodies in Interview. A bounded inbox
accepts deliveries; overflow refuses before ack rather than blocking the receive loop
or dropping acknowledged data. Four fixed questions each receive a unique conversation
ID, and answers must match both that ID and the authenticated subject. Exact router
headers are removed once, preserving answer paragraphs. Positive per-question timeout
and cancellation/disconnection while waiting for answers prevent report writing.
Report writing is ordinary file I/O, not atomic publication. SIGINT/SIGTERM cancel and
join the client. Report files are newly created0600 on the supported Unix platforms;
existing file permissions remain the caller's choice. These are subject answers, not
independent performance measurements. No Ensemble runtime is referenced by interviewer.

Evidence under `/private/tmp/ch26-csharp.HDZGCi`:
- `host-build-1.log` clean before behavior; `host-exercise-1.log` **seven checks pass**,
  job6515 collected exit0. The temporary Go probe imports the unchanged public
  `ensemble/gateway` client and starts the actual C# host process. It establishes
  stable key-only output even with an occupied port, no routing snapshot in key-only
  mode, nonzero bind failure, strict HTTP preflight, authenticated policy/offline
  admission, an obstructed periodic save reported on a real30-second tick followed
  by successful retry on the next tick, SIGTERM joining active and unfinished upgraded
  requests, restart/reply/resume delivery through durable ack3, nonzero final-save
  failure after client teardown, and missing/malformed config/unknown-flag refusal.
  Result: `tmp/gateway-csharp-host-714532677/result.json`; logs and generated identities
  stay in that owned fixture. This is new Go-client-to-C#-router evidence on this Mac,
  not Linux execution, deployment, power-loss durability or a security audit.
- Host source readback led only to expanded cleanup formatting; final
  `host-build-2.log` clean, job6536 exit0. The unchanged timer campaign was not rerun.
- `interviewer-build-3.log` clean; `interviewer-exercise-2.log` **four checks pass**,
  job6533 collected exit0. Actual C# interviewer/host plus unchanged Go subject verify
  four fixed questions, rejection of wrong-sender/same-conversation and correct-sender/
  wrong-conversation answers, complete private report, bounded unanswered-question
  failure, SIGTERM cancellation, and prompt failure on router disconnect. Failure
  cases leave no report. Result: `tmp/gateway-interviewer-4009733608/result.json`.
  Earlier three-check run is retained separately; final run follows explicit inbox
  construction and signal handling. The first interviewer build used a nonexistent
  Envelope.ProvenanceHeader member; corrected to declared Protocol.ProvenanceHeader.
  That was my editing/setup error, not a behavioral failure.
- Probe sources are `host-exercise/main.go` and `interviewer-exercise/main.go` in
  scratch. They use generated identities, owned loopback resources, bounded process
  joins, callback/RPC/file boundaries and real exits. The Go subject is not Ensemble
  and no live provider/model behavior is credited. Core/CLI/GUI and crypto/library
  code were unchanged, so their earlier evidence was not rerun for timestamps.

Next: thin Ensemble adapter and the actual three-process interview, including busy
actor request correlation and refusal/shutdown paths; then honest scoped executable
grading, manuscript rewrite, assembled source/test/prose review, source-only endpoint
and chapter handoff. Full Go interviewer/gateway-agent source and its two submission
correlation tests were read. C# Agent.SubmitAsync admits synchronously and returns
that turn's reliable result; do not add another scheduler or use lossy observations.
No adapter source, grading scope, chapter endpoint or acceptance yet. Deferred live
OAuth/cache still needs Bill's explicitly approved credential setup. No deployment,
private credentials, preserved-endpoint edits or workers were used.

# Chapter26 — identity-owned WebSocket sessions verified; standalone host and adapter remain

The C# Router now serves upgraded WebSockets through `AcceptAsync(id, socket, token)`;
HTTP/listener ownership stays with its caller. `ContainsIdentity` is preflight, not
authentication or a reservation, and admission is rechecked after upgrade. The base
Gateway library still has no ASP.NET/core/CLI/GUI reference. Tests alone use ASP.NET
for an owned loopback host. There is **no standalone C# host or Ensemble adapter yet**,
and no chapter acceptance or deployment.

`Common/Session.cs` declares identity-owned pending/active sessions, their socket,
Noise child, lifetime, write gate, coalesced delivery wake, per-session sent frontier,
and shared completion. `Router/{Session,Sessions}.cs` implements them. All pending
handshakes are real owned children. Current is a weak selection among those children,
not copied authority. KK plus identity/version-matched hello must succeed before
replacement; a future hello receipt refuses without evicting the old session.
Replacement changes authority under the root gate, then closes the old socket outside
it. Subsequent RPC/response/receipt policy actions and current-session checks share
the gate. Acknowledgments cannot exceed that session's delivery frontier. Encryption
and send share one writer gate; no socket write holds the policy gate. Write failure
after nonce advancement is terminal. Queue admission signals the live recipient;
wake signals contain no copied messages or permissions. Handler-return receipt and
at-least-once/replay semantics remain unchanged.

Session teardown cancels and joins delivery/read work, disposes Noise/socket resources,
and removes the identity link before settling its join. Router closes admission,
joins all captured sessions (including unfinished handshakes), then saves and destroys
the graph. Concurrent and subsequent DisposeAsync callers share one outcome; a failed
final save/cleanup cannot become success on the next call. The HTTP caller observes
connection/protocol errors; Router's join treats cleanup failures separately from
malicious peer input. Private faulted joins are explicitly observed without hiding
failures from their callers. Noise now accepts a Session strong owner in addition to
Client/standalone roots; the third-party crypto source is unchanged.

Evidence in `/private/tmp/ch26-csharp.HDZGCi`:
- `session-build-4.log` clean; `session-tests-3.log` **15/0**: the nine existing
  policy/snapshot groups (including copied retained Go files) plus six real server
  groups. These cover public C# clients, concurrent RPCs/deliveries and response
  provenance; offline replay/lost receipt and restart through ack3; failed key/hello/
  future receipt leaving the old session active; authenticated replacement; blocked
  socket delivery with other RPCs/save proceeding and rejection of ack2 while only
  sequence1 began sending; malformed/duplicate/text/tampered/oversized frames; and
  joined shutdown of pending and active sessions.
- Final review added lifetime-cancellation checks and explicit private-fault observation.
  `session-build-5.log` clean; `session-tests-4.log` **6/0** is the focused final server
  run, additionally exercising the real ten-second handshake deadline (not a fixture
  timeout) and a transport-disposal failure with complete detachment/final save and
  failure still reported. Job6492 collected exit0. The earlier 15-group run is not
  relabeled as following these final changes.
- Ownership/welcome changes warranted affected local regressions:
  `session-client-build.log`, `session-wrapper-build.log` both clean;
  `session-client-tests.log` **6/0**, `session-wrapper-tests.log` **6/0**;
  job6497 collected exit0. No new Go peer or Linux execution in those regressions.
  Prior Go-router client7/0 and Mac/Pi crypto8/0 remain separately scoped evidence.
- `session-tests-2.log` retains **11 pass/4 fail**. Three failures exposed a real
  product bug: global default omission removed required `queue_depth:0` from welcome.
  The unchanged Go `WelcomePayload` tag has no omitempty, so the C# field now explicitly
  serializes zero. I prematurely called this a test mismatch before reading the tag;
  that classification was wrong. The fourth failure was an old test expecting a second
  DisposeAsync to hide its failed save; the new shared-result contract preserves it.
  Earlier build logs retain test API-name mistakes; those were setup failures, not
  behavioral failures or acceptance credit.

Saved session/Common bodies and affected ownership paths were read back; full assembled
chapter source/test/prose review still lies ahead. Core/CLI/GUI are unchanged. Next:
standalone host with --config/--print-key, real listener lifetime and 30-second saves,
actual-process checks and Go-client-to-C#-router interoperability; then interviewer and
thin Ensemble adapter, remaining application checks, honest executable grading,
manuscript rewrite, assembled review, source-only endpoint and chapter handoff.
`solutions/ch26/cmd/gateway/main.go` was fully read for the host contract: public-key
output is not readiness, periodic save failures are reported/retryable, and orderly
stop closes listener admission before joined router/final save. No periodic runner
has yet been added. Deferred OAuth/cache still needs Bill and an approved setup.

# Chapter26 — C# router policy, administration and snapshot core verified; sessions remain

The standalone Gateway now includes a real Router policy graph and durable queue
core, not a running C# router service. `Gateway/Common/Router.cs` declares the root,
identities, groups, queues, immutable boundary records and operations.
`Router/{Router,Model,Administration}.cs` supplies the bodies; `RouterTests` exercises
them. Router owns identities/groups; each identity owns its queue. Base membership,
runtime grants and contactability are distinct reciprocal links. Constructors call
parent add methods; destruction removes all links and children, retaining files.
Mutations validate full proposed sets before publication under the root gate.
Readers/mutators revalidate ownership after acquiring that gate. Runtime-group
deletion also removes child base memberships and contacts, preventing accidental
regrant when a name is recreated. There is no copied configuration authority in a
connection, second scheduler, core/CLI/GUI change, or third-party crypto change.

Implemented the seven actual MCP tools, caller-derived authority, narrow child
registration, immutable bootstrap groups, live runtime administration, strongest
*accepted* group provenance, and sequence/receipt queues. Oversize delivery refuses
before sequence allocation. Snapshots use the Go `state.json` maps and
`queues/<validated-id>.json` (plain ID, including colon; the older URL-escaped note
was wrong). Retired group/sender references in envelopes remain historical values,
not resurrected policy links. Queue restore requires a coherent contiguous pending
tail and preserves drained frontiers. Optional Go strings normalize through one
shared Protocol helper; required data and duplicate JSON properties still refuse.
Save replaces individual 0600 files, with new directories 0700; existing parents
are not chmod'd. This is ordinary periodic/orderly-snapshot durability, **not**
fsync or a cross-file transaction. Periodic execution itself awaits the host/session
slice. Explicit save failure leaves live state retryable; final save failure still
detaches the graph. Missing files default; malformed or obstructed paths refuse.

Evidence in `/private/tmp/ch26-csharp.HDZGCi`:
- `router-tests-build-6.log`: actual final build, zero warnings/errors.
- `router-tests-6.log`: nine groups pass: reciprocal ownership/refusal, accepted
  provenance/queue admission, attenuated registration/live administration,
  restart/receipt retention, independent Go-shaped persistence/fail-closed restore,
  save-failure retry and destruction, RPC/schema contracts, concurrent admission
  and saving, plus restore from a **copy** of the retained real Go client fixture.
  That final check restored bob's next=3/ack=3/contactability and allocated four;
  it did not rerun or modify the Go service/fixture.
- Shared optional-envelope normalization warranted a client regression:
  `router-codec-client-build.log` clean and `router-codec-client-tests.log` six local
  groups pass; job6459 collected exit0. No Go peer was requested in that regression.
  Prior client7/0 and Mac/Pi crypto8/0 remain separately scoped, unchanged evidence.
- `router-path-baseline.log` retains a genuine failure: ENOTDIR surfaced as
  DirectoryNotFoundException and was incorrectly treated as empty state. Fixed
  by distinguishing absence from existing non-directory/dangling paths; the
  expanded test covers both. Early build logs retain my harness/API-name errors
  (GenerateKeypair/Public/MaxPlaintext/DataDirectory) and Frame-construction/analyzer
  corrections; these were setup/editing failures, not behavioral verification.

Readback covered Common and all three implementation bodies and compared admin
schemas/results against the actual Go source. This is slice review, not the final
assembled chapter source/test/prose review. No C# socket-host, session authentication,
connection replacement, periodic saver, or Ensemble adapter is implemented yet;
no application/grade/endpoint/chapter acceptance is claimed. Preserve ch27 and the
resolved Noise dependency; no reason to repeat unchanged primitive campaigns.

Next: identity-owned sessions with real Session/Noise ownership; authenticate fully
before replacing active connection; current-session checks and sent-sequence-bounded
receipts; serialized writes/delivery wakeups outside the policy gate; joined close
and final save; standalone hosts/interviewer and existing-actor Ensemble adapter.
Then focused interoperability/application checks, honest executable grading,
manuscript rewrite, assembled review, source-only endpoint and chapter handoff.
Chapter16/24 live OAuth/current-endpoint/cache work still needs Bill and an explicitly
approved credential store/setup. No deployment, restart, private-credential search,
workers or outstanding owned jobs at this boundary. Last free-space observation
9,901,860KiB (not a guarantee); keep at least1GiB free. Goal remains enabled at480s.

---

## Chapter26 — public WebSocket client verified against real Go router; C# router remains

Continued personally without workers; complete-book goal remains enabled at480s.
Gateway is INCOMPLETE. This slice adds wire values, strict codecs, and the standalone
public Client; it does not add the C# router, durable queue/policy owner, executable
hosts, Ensemble adapter, chapter grading, final rewrite/review or preserved endpoint.
Joint16 OAuth/24 current-endpoint/cache work remains open. No deployment, restart,
private-account access or changes to preserved endpoints/Core/CLI/GUI.

Common/{Protocol,Client}.cs declares immutable value-like records, IHandler/NopHandler,
actual Client persistent state and partial callable members. Protocol/Wire.cs and
Client/{Client,Requests}.cs implement them; the project includes those directories.
Client is a standalone root, owns its socket/cancellation/RPC state and Noise child;
NoiseKK now supports a Client-owning constructor with two-way registration/removal,
while standalone roots remain supported. The borrowed application handler is not a
second owner/scheduler or something the client destroys. No actor dependency.

Connect pins the existing KK suite/prologue, sends hello, validates encrypted welcome
and identity, then notifies the handler before starting its sole receive loop. One
write gate covers encrypt+send; failure/cancellation after encryption closes rather
than reuse the advanced nonce. Binary frames are bounded and fragmented reads joined.
Delivery callbacks run sequentially; a return permits cumulative receipt, and memory
receipt advances only after a successful write. Duplicate seqs are reacknowledged
without rerunning handler. Refused callbacks remain unacknowledged. Close interrupts;
Done/DisposeAsync join receive, resource release and final notification, including
surfaced callback failure. External lifetime cancellation settles pending calls.
Request cancellation removes just that waiter; late responses cannot attach to a new
request. MCP tools/list, arbitrary Request/CallTool, send, respond, contactability and
child registration follow actual reference API. No invented revoke/group convenience
methods remain. Respond preserves conversation without asking for another async reply.
No automatic reconnect, persisted client receipt or exactly-once claim.

The codec rejects duplicate JSON properties and malformed required fields. Explicit
JSON-RPC result:null is a present successful value, not absence; a focused baseline
reproduced premature close before fixing nullable JsonElement decoding. Omitted Go
optional strings normalize to empty, while required provenance/content stays strict.
ProvenanceHeader formats just the header, not its following separator/body.

Real interoperability finding: I initially used `seq` on hello/ack instead of the
reference's `last_acked_seq` (envelopes correctly use seq). The fake repeated that
assumption and passed six groups; unchanged Go retained receipt0 after deliveries,
exposing the defect. Fixed the tag and fake and compared the whole protocol. Final
real-Go queue has `ack:3,next:3,messages:null` (no outstanding messages), not merely
a client-side receipt. This
is why fake agreement was insufficient; repeated guessed contracts/anchors were my
avoidable process failures, not extra verification credit. Other saved corrections
cover header separators, response-mode omission and optional-string decoding.

Gateway/ClientTests is a dependency-free test executable plus framework-only local
ASP.NET WebSocket peer: six groups exercise real Noise/wire, held callbacks, duplicate
receipt, reordered/canceled RPCs, explicit-null outcomes, malformed/ambiguous outcomes,
text/oversize/tamper refusal, wrong pin/welcome/deadline, cancellation and joined teardown.
A seventh launches the unchanged solutions/ch26/cmd/gateway binary with generated
identities and an owned loopback port. It establishes contactability through RPC,
queues offline messages, receives ordered router-stamped provenance, reconnects with
receipt, replies with correlation, and checks saved receipt after joined SIGTERM.
This is an actual Go host + C# client exercise, NOT a C# router/Ensemble application.

Final evidence in /private/tmp/ch26-csharp.HDZGCi:
- client-final-4-build.log: zero warnings/errors.
- client-final-4-tests.log:7 passed,0 failed (six fake-boundary groups plus real Go).
- client-owned-noise-regression.log: six existing local groups pass after owner change;
  Go primitive-peer interop explicitly not rerun. Existing Mac/Pi8/0 evidence retained
  at its prior boundary, not relabeled as Linux WebSocket-client testing.
- Job6424 collected exit0. Successful Go fixture tmp/gateway-go-client-
  eb6d218608614e47b34ab407ab5eb21d retains config/identities, deliveries/reply, router
  state/queues and stderr/stdout. Go process joined; no deployment. Test configuration
  uses only generated keys, no live credentials. Disk remains well above1GiB.

Retain client-null-baseline.log and client-final-2-tests.log (receipt0), earlier test
logs and failed fixtures. Harness-only failures: wrong map/array config schema,
nonexistent health route and MCP ping, wrong expected stamped content/group, guessed
project/storage paths. These failed before their respective intended assertions and
are not router product defects. Refusal helper now excludes timeout as expected
failure. Source edits are limited to Gateway (including Noise ownership there),
documentation/cursors; third-party algorithms/assembly unchanged. Gateway/README.md
states ownership, lifecycle, reproducible checks and explicit limits.

Next: actual Common router identities/groups/queues/connections, durable state/admin
and server host, then Ensemble adapter/interviewer and required actual applications,
honest scoped grader, manuscript rewrite, assembled review, source-only endpoint and
chapter handoff. Read actual reference/grader contracts before editing; do not repeat
unchanged dependency/Pi campaigns or treat this client result as chapter completion.

## Chapter26 — public key/Noise layer verified on Mac and xyzzy; client/router remain

Bill re-enabled complete-book execution at fixed480-second idle reminders, then
said “Please continue.” Reminders remain enabled. Continued personally, no workers.
Chapter26 is INCOMPLETE: this slice adds the standalone public transport/key layer,
not WebSockets, routing, delivery persistence, MCP administration, Ensemble adapter,
actual application exercise, executable grade, final prose/review or endpoint.
Chapter27/earlier endpoints remain untouched; live16 OAuth and24 cache check stay open.

New standalone Gateway/Ensemble.Gateway.csproj compiles Common/Cryptography.cs and
Cryptography/{Keys,NoiseKK}.cs. Common owns the actual partial declarations and all
persistent state; the library references only the identified Noise dependency,
not Ensemble core/GUI. Keypair is explicitly immutable/value-like; its public half
is derived from supplied private bytes. NoiseKK is currently a standalone session
root that explicitly disposes its upstream resources, serializes calls, preserves
handshake payloads and AD, bounds transport messages, and makes authentication
failure terminal. No rekey is exposed. When connection ownership is introduced,
model that real relationship and both ends rather than hiding copied services.

Key import needed a second narrow upstream API exposure: KeyPair.Generate(span)
validates32 bytes and calls existing dh.GenerateKeyPair(privateKey). No new curve,
cipher, KDF, hash, nonce or handshake algorithm. Of31 copied upstream C# files,
only KeyPair.cs and Transport.cs now differ from the pinned v1.0.0 commit. License,
provenance and separate Ensemble.Noise assembly remain; dependency README updated.
The earlier five-group AD evidence remains historical, not relabeled as this slice.

Keys.LoadOrCreateKeypair uses the historical base64-private-only file format, fresh
public derivation, bounded strict UTF-8 reads and fail-closed malformed/permission
errors. New directories/files are0700/0600; existing directories are not chmod'd.
An owned same-directory temporary is flushed then atomically/exclusively linked
into the destination on macOS/Linux. A concurrent loser loads the winner. Cleanup
removes only a temporary actually created by this call and preserves an original
failure. Unsupported platforms/filesystems refuse rather than weaken file privacy
or atomicity. No directory-fsync/power-loss, sandbox or secret-store-search claim.

Real finding: File.Move(temp,destination,overwrite:false) allowed concurrent callers
to return different identities on this Mac. public-tests.log retains the failing
seven-pass/one-fail run and failed private fixture. The .NET10 Unix source's
check-then-replacing-rename path explains why this was insufficient; fixed with
exclusive link and exact EEXIST handling, no broad retry. This is not a crypto issue.
Source inspected: https://github.com/dotnet/runtime/blob/v10.0.0/src/libraries/System.Private.CoreLib/src/System/IO/FileSystem.Unix.cs.

Focused suite now has six local groups and two real-Go interoperability groups.
New groups cover immutable import/encoding, malformed/short/oversized keys, twelve
concurrent creators, exact file format/permissions, restoration/no-overwrite, public
session disposal and terminal refusal. The two peer groups now run through public
C# NoiseKK with imported private bytes against the byte-identical Go wrapper in
both roles, with empty/text/binary AD and empty/Unicode/4096-byte payloads. Low-level
nonce-retention tests remain separate from the public session's terminal failure.
This is concurrent-call evidence, not a two-process credential-store campaign.

Final evidence /private/tmp/ch26-csharp.HDZGCi: public-final-2-build.log clean with
zero warnings/errors; public-final-2-mac.log8/0; public-linux-2-build.log clean
self-contained ARM64 publish; public-go-linux-2-build.log; public-xyzzy-2.log8/0 and
public-xyzzy-2.exit0. Job6370 collected exit0. scp/checksum and actual Debian ARM64
execution used new owned /tmp/cadence-noise-probe.ec1JcZ/public-transport, not the
prior ad-extension directory. No deployment/shared service touched. Earlier failed
chains stopped before remote setup. Core/CLI/GUI projects are unchanged; their
old tests were not needlessly rerun or relabeled. Public source/tests reviewed;
upstream delta, internal references and added public-source whitespace checked.

Setup corrections retained: first public build passed spans to the pinned upstream
array API; fixed owned copies and erased the temporary private copy. A later build
correctly refused Unix mode calls because the analyzer needed an explicit Windows
exclusion; no warning suppression. Both are build failures, not behavior results.
Final ownership cleanup guard is included in the final Mac/Pi results above.

Read the full current717-line chapter, Go public types/client/noise and full router
server lifecycle; inspected existing Common/project boundaries. Chapter26 TL;DR
now specifies the C# standalone assembly, Common/functional layout and ownership,
and explicitly marks remaining Go draft sections as reference, not implementation.
The rest of the manuscript is NOT rewritten or reviewed as final. Next read the
remaining actual admin/group/storage/grader contracts and implement public wire
values, WebSocket client, router and durable delivery with real parent relationships,
then adapter and the full chapter completion loop. Do not restart dependency campaigns.

## Chapter26 — approved AD forwarding extension verified on Mac and xyzzy; Gateway incomplete

Bill approved preserving associated data via a narrow source extension, then said
“Please proceed.” Suite remains Noise_KK_25519_ChaChaPoly_SHA256. AD decision is
resolved, not a remaining human blocker. Implemented only the dependency extension
and focused tests in this slice; C# Gateway client/router/adapter are NOT implemented.
No preserved endpoint edited, no install/deployment/restart/commit or private keys.

Source: solutions/edition-3/main/ThirdParty/Noise copies31 C# files from exact
upstream v1.0.0 commit4279cfa57df028dec4a4cb3ee9579a77a5cc39fa. All except Transport.cs
are byte-identical, as is the MIT license. New three-argument overloads pass AD into
existing CipherState EncryptWithAd/DecryptWithAd; two-argument methods delegate with
empty AD. No primitive, nonce, KDF, handshake or hash algorithms changed. Local
Noise.csproj targets net10, retains upstream C#7.2/nullability conventions, names the
modified assembly Ensemble.Noise, pins native libsodium exactly1.0.22. This is an
identified third-party dependency, not a structural rewrite into course Common.
Core/CLI/GUI project files and NuGet.Config match ch27; core has no new dependency.

Gateway/Tests/NoiseTests.csproj is a standalone real-library test executable, with
no test framework/provider fake. Three groups cover legacy/new empty AD, nonempty
and binary AD in both directions, empty/Unicode/65519-byte payloads, argument/disposal
refusal; changed/omitted AD, altered tag and replay rejection with nonce retention;
wrong configured static keys/prologue. Two additional groups speak over pipes to an
actual Go peer importing a byte-identical copy of solutions/ch26/gateway/noise.go
with flynn/noise1.1.0, in both initiator/responder arrangements. Each direction tests
empty/Unicode/4096-byte payloads and empty/text/binary AD, including wrong-AD refusal.
Go peer has no production authority, temporary keys only, bounded replies, joined
EOF exit and owned-process cleanup. It is not an invented compatible algorithm.
Tests retry rejected input to inspect nonce behavior; production Gateway policy
still closes authentication-failed connections. Peer manifests and reproduction
instructions are saved; original Go wrapper is copied only to scratch for execution.

Evidence: scratch /private/tmp/ch26-csharp.HDZGCi (pointer cadence-ch26-current),
noise-restore.log/noise-build.log clean, go-peer-build.log, noise-tests.log5/0,
job6316exit0. Self-contained osx-arm64 and linux-arm64 publishes succeed. Published
Mac result published-mac-tests.log5/0. Cross-built Go peer CGO-disabled ARM64;
scp into owned existing remote scratch subdirectory
/tmp/cadence-noise-probe.ec1JcZ/ad-extension; archive checksum verified; actual Pi
run xyzzy-tests.log5/0 and xyzzy-tests.exit0; job6318exit0. Both machines executed
modified source, not merely the previous unmodified package. No claims for ARM32
modified build, AESGCM, other Noise patterns, or audit. No other tests relabeled or
rerun: existing Ensemble runtime was not changed. Logs and remote scratch retained.

Review limitation: upstream CipherState.Rekey copies key.Slice(Aead.KeySize), a
16-byte suffix of48-byte result, unlike Go's first32. Existing Gateway has no rekey
calls. Recorded this source discrepancy in dependency README; did NOT claim a tested
rekey failure, silently fix a primitive/state routine, or rely on it. Do not expose
rekey in Gateway without separate review/correction and interoperability checks.

Avoidable inspection errors: guessed keys.go (actual key helpers are in noise.go),
one clipped combined source read and a stale rekey line range. Recovered exact
source; these are not product failures. All build/runtime checks above passed on
first executed run. After runtime verification, only documentation, gofmt and test
manifest/reproduction files changed; reduced go.sum rebuild passes readonly.

Current next work: faithfully implement public Gateway wrapper/client/router and
adapter using this dependency, with Common-owned vocabulary and real ownership;
read actual wire/storage contracts before declaring them. Then focused behavior,
application/grade reconciliation, chapter rewrite/review and source-only endpoint.
No extra approval needed for the agreed AD forwarding or suite. Chapter24 live
OAuth/cache remains human-dependent. Goals/epilogue remain complete. Reminders off.
Last local free space9,875,728KiB; remote10,430,384KiB. All jobs collected, no workers.

## Chapter26 — Bill selects ChaChaPoly; public associated-data API mismatch identified

Bill: “Let's go with ChaChaPoly.” Confirmed suite is
Noise_KK_25519_ChaChaPoly_SHA256, preserving Ensemble's existing protocol.
Continuation uses the tested Noise.NET1.0.0/libsodium1.0.22 candidate; no Gateway
runtime code or repository dependency changed yet. AESGCM investigation is no
longer on the critical path. Read full current717-line chapter and procedure.

Exact remaining dependency/API issue, not an unsupported crypto redesign:
- solutions/ch26/gateway/noise.go exposes Encrypt/Decrypt(data, associated),
  forwarding associated bytes to flynn/noise CipherState.
- solutions/ch26/gateway/noise_test.go tests transport with nonempty `binding` AD.
  Chapter26 public independent Noise contract likewise includes associated data.
- All actual Gateway client/server transport calls pass nil AD. Existing wire
  behavior therefore matches Noise.NET's transport for this chosen suite.
- Noise.NET1.0.0 Transport.WriteMessage/ReadMessage omit an AD argument; inspected
  Transport.cs forwards null to internal EncryptWithAd/DecryptWithAd. The required
  primitive support exists internally, but the published API doesn't expose it.
  Source retained at /private/tmp/cadence-noise-probe.ENFVit/source/Transport.cs;
  upstream https://github.com/Metalnem/noise/blob/master/Noise/Transport.cs .

Need Bill's choice: retain full wrapper compatibility by a small source-level
Noise.NET extension forwarding AD to its existing routines, or explicitly narrow
C# public transport API to Gateway's empty-AD use. Do not silently discard AD,
prepend it to plaintext, reflect into library private state, substitute primitives,
or claim the existing probe tested nonempty AD/Go interoperability. Both options
preserve ChaChaPoly; the decision concerns public API/dependency maintenance only.
Goal reminders remain off at this real contract decision. Chapters24 and26 remain
open; completed27 endpoint/epilogue untouched. Last disk10,225,896KiB.

## Chapter26 remote probe — xyzzy ChaChaPoly passes; AESGCM availability refuses

Bill authorized `ssh xyzzy` and instructed use of scp. Connected using existing
SSH configuration with BatchMode and strict existing host-key checking. Read-only
platform inspection: Debian12/bookworm, Linux6.12.25+rpt-rpi-2712, aarch64/64bit,
glibc2.36. Linux CPU flags include aes and pmull. No installation or service changes.

Transferred the existing self-contained linux-arm64 probe by scp into owned private
`/tmp/cadence-noise-probe.ec1JcZ` on xyzzy; SHA256 of archive verified before unpack.
Ran with isolated home/tmp/env and60s timeout. Remote .NET10.0.14 execution:
- Noise_KK_25519_ChaChaPoly_SHA256 PASS: two-message handshake payload/hash agreement,
  consecutive bidirectional transport (empty/UTF8/4096bytes), wrong static-key
  rejection at both identity configurations, altered-ciphertext rejection.
- Noise_KK_25519_AESGCM_SHA256 FAIL: Noise.NET constructor throws NotSupportedException
  because libsodium reports AES-GCM unavailable. Combined probe exit1; transfer/job
  exit0 must not be relabeled as all tests passing. Job6288 collected.

Follow-up loads the exact bundled native .so through Python ctypes after sodium_init:
version1.0.22, NEON1, ARMCRYPTO0, AESGCM availability0. Remote/local library SHA256
match54f416a70e0d982a63e7f2402ff6f2a8666bf0a430404a4205ed8eee1868b9db.
Read-only objdump shows actual AES implementation present, availability calling
sodium_runtime_has_armcrypto rather than a constant unsupported stub. Thus do NOT
claim xyzzy lacks AES hardware or that the library compiled AES out; discrepancy
between OS flags and native detection remains unresolved. No feature check bypass,
cipher substitution, native library replacement, or protocol change attempted.

Logs copied back to local scratch `/private/tmp/cadence-noise-probe.ENFVit`:
xyzzy-platform.log, xyzzy-execution.log, xyzzy-run.log/.exit,
xyzzy-native-features.log, xyzzy-aes-disassembly.log, xyzzy-scp.log,
xyzzy-upload.sha256 and xyzzy-scratch.txt. Remote scratch retained, no shared cleanup.
Last remote free space10,555,816KiB. No remote background jobs remain.

This is actual Mac-and-Pi execution for ChaChaPoly, not merely cross-publish.
AESGCM passes on Mac but not xyzzy with this package. Same-library round trips do
not prove Go interoperability or constitute an audit. Gateway dependency adoption
and intended suite still require Bill's confirmation; no repository runtime changes.

## Chapter26 dependency probe — Noise.NET works on this Mac; Pi execution and adoption pending

Bill authorized an isolated library test on this Mac and assessment of cross-building
for Raspberry Pi `xyzzy`. No Gateway implementation, production dependency adoption,
remote connection, global install, restart, or preserved-endpoint edit occurred.
Scratch: `/private/tmp/cadence-noise-probe.ENFVit`; pointer
`/private/tmp/cadence-noise-probe-current`. Package restore/cache/home/tmp are local
to that scratch, using existing SDK10 under env-i. Probe source uses real Noise.NET
operations and throwaway keys, not replacement crypto or a scripted peer.

- Noise.NET1.0.0 default restore selects libsodium1.0.16. Clean compile, but both
  suites fail native initialization with DllNotFoundException on ARM64 macOS.
  That package has osx-x64/linux-x64/windows native assets, not ARM assets.
  Evidence: baseline-restore.log, baseline-build.log, baseline-run.log and .exit=1.
- Explicit libsodium1.0.22 alongside unchanged Noise.NET1.0.0: clean build and
  successful execution on macOS26.3.1 ARM64/.NET10.0.14. Both
  Noise_KK_25519_ChaChaPoly_SHA256 and Noise_KK_25519_AESGCM_SHA256 pass two-message
  handshake/hash/payload checks, consecutive bidirectional transport with empty,
  UTF-8 and4096-byte payloads, wrong configured static keys on either side, and
  altered transport ciphertext rejection. current-run.exit=0.
- Self-contained publishes for osx-arm64, linux-arm64 and linux-arm all succeed.
  file inspection confirms Mac ARM64 apphost plus matching dylib, Linux AArch64
  apphost/libsodium.so, and Linux ARM EABI5 apphost/libsodium.so. This is RID-targeted
  publish with bundled runtime/native assets, not a NativeAOT claim. Published Mac
  executable repeats both passing suites, published-mac-run.exit=0. Job6285 exit0.
- xyzzy was NOT contacted. Its OS/CPU/runtime prerequisites and actual execution
  remain unknown. AES-GCM specifically needs CPU crypto support; the actual C#
  wrapper rejects when libsodium's availability check fails. Neither cross-publish
  nor a same-library round trip proves Go interoperability or security auditing.

Primary source: https://github.com/Metalnem/noise (actual HandshakeState, Protocol,
Aes256Gcm and Libsodium source read); package versions/assets from
https://api.nuget.org/v3-flatcontainer/libsodium/index.json ; hardware limitation:
https://doc.libsodium.org/secret-key_cryptography/aead/aes-256-gcm (HTTP200).
Temporary .NET dependency restores are test-only; no repository project changed.
Last observed free space10,260,356KiB. All probe jobs collected. Next: obtain
permission for a temporary non-installing run on xyzzy and establish its platform;
confirm dependency adoption before continuing Gateway. Reminders remain off.

## Epilogue complete — independent authorized work finished; two human-dependent items remain

Replaced `book/edition-3/epilogue.md` with **What we can rebuild**, in Cadence's
voice. Read the prior Astra/Codex epilogue and Bill's exact Final scope direction
before writing; read back the complete saved replacement (160 lines). It preserves
Bill's origin quotation and attributes his third-run/first-closed-loop framing,
explains expert-present chapterwise rebuilding without lowering code quality,
and uses the actual duplicated-memory-input and Goals submission corrections to
connect reasons to requirements. Avoidable editing mistakes remain mistakes, not
verification credit. Earlier Codex history is not presented as this run's work.

The epilogue explicitly leaves Chapter24 live OAuth/cache testing and Chapter26
Gateway unfinished, distinguishes local peers/browser checks from real integration,
and makes no independent-review or universal reproducibility claim. Both local
links resolve; trailing-whitespace and scoped Git diff checks pass, exit0. No code
changed, no behavioral tests rerun or relabeled, no new endpoint needed for prose.
Chapter27's mandatory handoff was completed before this work. Its preserved335-file
endpoint and continuing main were not edited during the epilogue work.

All currently independent authorized work is complete. Remaining work with Bill:
1. Chapter24: approved credential-store/setup choice, joint Chapter16 real OAuth,
   then current official-endpoint verification and bounded live cache measurement.
   Historical cache failure is not current evidence; no workaround implemented.
2. Chapter26: approve faithful source-port of flynn/noise/required primitives or
   a supported C# cryptographic dependency. Preserve Ensemble's specified
   Noise_KK_25519_ChaChaPoly_SHA256; CodeRhapsody AESGCM is not a substitution.
   No crypto implementation/dependency installed and no Gateway completion claimed.

GOAL now contains only this current continuation boundary; older chapter evidence
remains here. Disable idle reminders for these human-dependent items. No workers,
installs, restarts, deployment, commits, private credential access or Bill browser
use. No live provider/billing, native Linux/x64, audible or human-accessibility
completion added. Last disk observation10,827,472KiB free. No new background jobs.

## Chapter27 — locally complete, reviewed and preserved; epilogue next

Personally, no workers. Chapter27 Goals is locally complete at
`solutions/edition-3/ch27`: **335 source/support files**, all byte-identical to
reviewed main at preservation; **13 ProjectReferences** resolve inside endpoint.
No symlinks, bin/obj, runtime state or private artifacts. Main remains continuing
workspace; ch23 and earlier endpoints are untouched. No new Gateway implementation.

Rewrote/read back the complete Goals chapter in Bill's voice: hour-into-day-long
stall motivation, separate fresh objective vs idle trigger, opt-in session authority,
Common vocabulary/actor mail/timer generations, normal reminder admission, no
background-job cancellation, request-only projection on all four surfaces, fresh
bounded UTF-8 reads and correlated error handling, off/clear/unload/submission
semantics, selected-Agent acknowledgment/status, practical isolated setup and
exercises. Removed Go/Codex/runtime success claims from the C# account. Four shell
blocks parse; source-map paths and actual host flags/skill syntax checked. Current
numbered ch27 graders are Gateway and supply no Goals grade; no invented numeric
executable score. Read complete 27-file ch23-to-main source/support delta plus prose,
recovering clipped spans. Final readability edits only expanded two timer/disable
blocks, fixed fault-settlement indentation, and expanded equivalent GUI status text.

Assembled review found a real lifecycle omission: successful delegated submission
left automatic reminders enabled. New seventh GoalTests group first failed in
submission-baseline.log (6/1). StoreSubmission now calls Common-declared
GoalSubmitted only after successful durable result publication under ingress,
before signaling waiters: immutable config becomes disabled immediately, actor mail
refreshes/disposes timer. Refused submission leaves reminders unchanged; already-off
submissions cause no extra mail. File/instruction retained, parent goal unaffected,
child never inherits parent goal. Initial fixed run still failed because an immediate
status snapshot exposed the invalidated old deadline before refresh; CaptureGoalStatus
now returns DueUtc only for scheduled phase. Retained failing runs are not passes.

Final verification in /private/tmp/ch27-csharp.nuMmE1:
- review-core-build-2.log clean; review-goals-2.log **7/0**;
  review-submission-2.log **3/0**; review-shared-2.log **304/0**. Chain6220 exit0.
  These include final submission/config/deadline behavior, precede only the stated
  readability edits and README. Earlier 302+separate-catalog evidence remains historical.
- final-cli-build.log/final-gui-build.log clean, zero warnings/errors, after readability.
- final-gui-tests-build.log clean; final-gui-controls.log real socket group passes
  selected-Agent availability/isolation/atomic refusal/deadline/reconnect/error/off/clear.
- final-browser.log and tmp/goal-host-RamBCj/result.json: actual GUI + isolated Chrome,
  temporary primary/shipped Goals, owned local provider/profile/ports; exactly two
  local requests (human while off, edited goal on automatic reminder), visible status
  and retained state on reload, off/clear, disconnected unknown state, no page errors;
  host joined exit0. Final chain6241 exit0. Earlier browser results remain separately
  scoped; the final UI run was justified by changed status construction, not freshness.
- Added-line whitespace over complete source delta and scoped tracked whitespace
  checks pass. Endpoint verification ran after final README. Last disk observation
  10,745,572KiB free, not a guarantee.

Retained setup mistakes: new submission hook initially used nonexistent inbox/Mail
initializer; build refused before tests, corrected to actual mailbox/constructor.
A grep guard for the wrong partial test class stopped one chain before build; actual
--child-completion-tests selector exists. One shell exercise error came from an
apostrophe inside ${...:?}; fixed and all four blocks parsed. Repeated guessed
formatting anchors refused (one batch saved only two readability blocks); final saved
source was inspected and reconciled before builds. A collected review-command handle
was briefly mistaken for regression; actual6220 exit0 collected separately. These were
my process errors, not product failures or credit for thoroughness.

No live provider/semantic obedience/OAuth/cache/billing, native Linux/x64, audible
speech or human-accessibility completion claimed. No install, shared restart,
credential search, commit, deployment, Bill browser/8088 or workers. Chapter24 live
OAuth/cache decision remains OPEN. Chapter26 still incomplete awaiting Bill's
source-port flynn/noise/primitives vs approved C# crypto dependency decision; preserve
Ensemble Noise_KK_25519_ChaChaPoly_SHA256, no redesign or substitution.

NEXT: mandatory chapter handoff, then replace the inspected old Astra/Codex epilogue
with a new Cadence-voice account. Bill's direction is below at "Final scope": expert-
present chapterwise executable knowledge, his third-run/first-closed-loop framing,
code as rebuildable secondary artifact without lower quality, and explicit unresolved
scope. Existing epilogue was fully read; it describes an earlier Edition2 process,
not this run. After epilogue review, all independent authorized work is complete:
report the two human-dependent blockers and disable idle reminders rather than wake
repeatedly into them. No active owned jobs at this boundary.

## Chapter27 — selected-Agent Goals GUI locally checked; prose/review remain

Personally, no workers. Goals core checkpoint below remains valid at its scope.
Added read-only GoalStatusAsync actor mail, available even unloaded without granting
control or resetting the timer. GoalStatus adds Available; phases distinguish unloaded,
unconfigured, file_error, paused, busy, off and scheduled. Actual GoalIdle predicate
reports busy even when reminders are off. Configuration remains skill-gated.

GUI reuses existing Agent-tree status refresh, no new polling/timer owner. Strict
session-only update_goal/get_goal socket controls validate full patches before actor
mutation; file/null/duplicate/unknown/invalid-delay refusals preserve configuration.
Acknowledgment follows actor application and goes only to peers viewing that Agent;
identity checks reject stale targets. No persisted policy/appearance changes. File
errors become safe control refusals, not raw OS diagnostics or socket teardown.
Text-labelled panel has Apply/Off/Clear/Reload, edits survive normal refreshes,
selection is visibly pending until replay, stale Agent frames ignored, disconnect
marks current state unknown and disables controls. No automatic speech added.

Evidence in /private/tmp/ch27-csharp.nuMmE1:
- gui-controls-test.log initial real-socket test passed; status-gui-tests.log final
  passed availability, two-Agent isolation, malformed/refused atomic patches,
  mismatched target, same deadline on query, pause/reconnect, read-error/off/clear,
  no goal-body status leak and no settings.json persistence.
- gui-goals-core.log six groups passed before final status polish;
  status-core-tests.log final six groups passed including held real HTTP request:
  off still reports busy, completion reports off. status-core-build-2.log and
  status-gui-build.log clean. gui-settings-regression.log existing actual socket
  memory/recall test passed; final phase/UI polish does not change those paths.
- Actual isolated Chrome/GUI exercise goal-host-exercise.mjs uses temporary primary
  skill + shipped goals, startup load, owned loopback provider/port, owned profile.
  Initial result tmp/goal-host-14hveB/result.json successful; final after display
  corrections tmp/goal-host-cevtTc/result.json successful: exactly two local requests
  (human while off, then one idle reminder), edited goal in second SYSTEM request,
  visible no-goal/off/waiting/off/clear and disconnected labels, browser reload
  retained core state, clear retained file, host SIGINT joined exit0. No page errors.
  status-browser.log records summary. status-host-build.log actual GUI clean.
- No live model obedience/provider/OAuth/billing, hearing, human visual accessibility
  or native Linux claims. Existing shared302 plus corrected tool group were not
  rerun or relabeled. No numbered ch27 executable grade: those graders are Gateway.

Review caught display omissions (disconnected stale snapshot, working hidden by off)
corrected before final exercise. Editing failures were mine: one guarded phase patch
refused, fail-fast was initially placed too late so unchanged-source checks ran;
then status-core-build.log caught invented GoalBusyLocked helper. Replaced with actual
GoalIdle predicate; no behavioral test ran at that failed build. Retain those logs,
not extra verification credit. Last disk observation10965380KiB, not guarantee.

Chapter27 still INCOMPLETE: current manuscript read completely but not rewritten;
assembled review, CLI final build, source-only endpoint and required handoff remain.
Next rewrite Bill's hour-into-day-long-job opening and C# body/exercises/Take It for
 a Spin; distinguish persistent objective from idle opportunity and exact limits.
Then full bounded source/test/prose review, preserve27; Cadence epilogue afterward.
Chapter26 crypto dependency and Chapter24 live OAuth/cache still explicitly blocked.
No workers/installs/restarts/commits/private credentials/Bill browser/8088.

## Chapter27 — Goals core implemented and locally checked; GUI/prose remain

Independent Goals work resumed personally, no workers, while Chapter26 awaits
Bill's crypto-dependency choice. No Gateway implementation/dependency added and
no live24 OAuth/cache claim. Active main now differs from preserved ch23; earlier
endpoints remain untouched. Scratch: /private/tmp/ch27-csharp.nuMmE1; pointer
/private/tmp/cadence-ch27-current; offline wrapper reuses approved env-i/private SDK.

Read current Goals TL;DR/body/design notes and prior Go implementation, then actual
C# actor/render/skill/control ownership paths. Common/Goals.cs declares value-like
GoalUpdate/GoalStatus/config/read-failure values and Agent fields/partial operations;
Agent/Goals.cs implements them. One immutable configuration under ingress is the
configuration authority, not History, GUI or a second scheduler. A Timer only posts
generation-bearing mail. Actor rejects stale wakes, respects foreground/queued work,
control work, direct Engine sends, pause/stopping, and fresh skill authority. Resume
starts a fresh interval; explicit interrupt disables; shutdown/fault/final cleanup
stops/disposes timers. Default480s; positive finite Timer-representable delays
(1ms through4294967294ms) accepted. No backoff, watchdog, replay or completion claim.

GoalControlAsync validates changes before publication; no-argument query does not
reset due time. File selection does not enable; off retains file; empty file clears;
off/clear don't depend on a readable file. Sandbox's existing rooted regular-file
reader supplies fresh strict UTF-8 (max1MiB, nonempty after trimming, BOM accepted).
Engine adds current goal to each ordinary request's instruction field, even while
reminders are off, not History/manuals. Same-turn tool/handoff requests reread it.
Fresh read failure posts a configuration-correlated failure, stops reminders, and
provides a bounded non-path-leaking ordinary-turn warning; automatic wake preflights
before admission and render rechecks, refusing inference on failure. Stale read
outcomes cannot mutate a newer configuration. Auxiliary recall/compressor projection
is not given parent goal text. Synthetic reminders skip user-query recall and use
the normal Submit admission/accounting path; this is an opportunity, not progress.

Shipped Skills/goals/SKILL.md grants only goal_control; loading alone configures
nothing. Appended ToolKind, existing control worker/mailbox, and snake-case public
tool output reused. Unload and Skills detachment clear session state, revoke prompt
injection and stop future wakes; durable goal files remain. Retained pending manuals
are not authority. Legacy unskilled library fallback required an explicit exception:
GoalAvailable now requires attached Skills plus actual active grant. Declaration,
dispatch and projected deferred references all enforce it. Existing ordinary tools
retain their fallback. No goal restart persistence or inherited bootstrap fields.

Evidence and corrections:
- Initial core build clean; six focused groups now pass in goals-optin-final.log:
  API/tool validation and opt-in; exact provider instruction fields on all4 surfaces;
  fresh text/failure/repaired-off projection; Skills detach/unload; real loopback
  idle reminder/pause/busy/off/interrupt; missing file refuses before HTTP; stale
  timer/shutdown/new-Agent defaults; actual two-request handoff freshness/no goal
  bytes in canonical dialogue. Real components and loopback boundaries, no live API.
- Product regression: plain legacy Agent installed catalog exposed Goals through
  its existing broad fallback. legacy-baseline.log is5pass/1fail; narrow gate fixed
  it. Early null-field suspicion was FALSE: existing schema validation already
  rejects file:null; no null product patch. Null regression remains useful.
- Shared-optin.log:302pass/1fail, the sole failure an old tool-loop assertion that
  catalog==declarations. Tests/ToolTests.cs now counts the installed dormant control
  but checks every legacy request against ordinary declarations without Goals.
  tools-final.log separately passes that entire group via --tool-tests. No fresh
  full303/0 run claimed; other302 groups unchanged by this test-only correction.
- cli-optin-build.log/gui-optin-build.log actual hosts zero warnings/errors, exit0.
  Final tracked added-line whitespace check clean. No production change after those
  builds/shared run; subsequent changes only tool-loop test/selector and these notes.
- Fixture failures retained: tests-build.log used Seq instead of Sequence; expanded
  goals-null-baseline.log5/1 tried reloading while unload was pending. My claim that
  current C# handoff immediately retires manuals was WRONG; it retains them. Test
  now checks immediate revocation and new-Agent defaults, not invented retirement.
  Early shared-core.log302/1 catalog assertion predates opt-in fix. Editing anchor
  failure left only new Goals files, reconciled against saved source. One accidental
  Git pager was identified/terminated alone; later Git checks use GIT_PAGER=cat.

Chapter27 remains INCOMPLETE. No GUI controls/status, no actual-host browser exercise,
no C# chapter rewrite, no complete assembled review or endpoint. Next add read-only
actor status query usable by GUI even when skill unavailable (configuration stays
skill-gated); show selected-Agent unloaded/off/paused/busy/scheduled/file-error state
and opt-in controls through existing socket, not browser authority/timers. Then
verify GUI isolation/control application and actual local host, rewrite in Bill's
voice (hour-into-day-long-job cold open, prompting vs compression persistence),
review/preserve27 and mandatory chapter handoff. Cadence epilogue afterward must
retain Gateway and live24 limitations. No workers/installs/restarts/commits/private
credentials/Bill browser or8088. Existing reminders remain480s.

## Chapter26 — prior Noise implementations compared; port boundary needs Bill

Resumed the interrupted source comparison under Bill's faithful-port direction;
no redesign, replacement protocol, cryptographic source or package added. Read all
166 lines of CodeRhapsody gateway/pkg/noise/noise_kk.go, all295 test lines and its
go.mod. Only source/dependency files read, not gateway runtime data or keys.
Compared Ensemble agent/gateway/noise.go with solutions/ch26/gateway/noise.go:
byte-identical. Both projects depend on github.com/flynn/noise v1.1.0.

Important difference: CodeRhapsody selects 25519_AESGCM_SHA256; Ensemble selects
25519_ChaChaPoly_SHA256. Current manuscript lines170ff explicitly name Ensemble's
Noise_KK_25519_ChaChaPoly_SHA256, resolving which suite the Ensemble port preserves.
CodeRhapsody is a comparison, not a wire-compatible peer for that suite. No suite
switch or configurable multi-suite mechanism proposed. Source test inspection is
not a test run or independent cryptographic validation.

Read the existing Go module's cipher_suite.go lines1-160 and go.mod at
/Users/bill/go/pkg/mod/github.com/flynn/noise@v1.1.0: it delegates X25519 to
 golang.org/x/crypto/curve25519 and ChaChaPoly to x/crypto/chacha20poly1305.
These are not implementations in either project's wrapper. Therefore a literal
wrapper translation still needs a C# implementation of that dependency boundary.
Bill must choose/authorize whether to source-port flynn/noise and its required
primitives or use an approved C# dependency; do not silently choose custom crypto,
native runtime glue, a Go helper process, or another protocol. The earlier framework
probe establishes only that inspected framework route's limitation, not that a
faithful port is impossible. No new source, build/test result or endpoint.

Gateway remains incomplete. Independent Goals remains authorized while awaiting
that choice; no Goals source yet. Ch24 live OAuth/cache remains deferred. Previous
entry's blanket next-Goals cursor must not lose this specific Gateway question.

## Chapter26 — dependency blocker confirmed; continue independent Goals

The Chapter25 handoff's HTTP/UDP unauthenticated proxy description was WRONG.
Actual current Chapter26 is a717-line authenticated Noise-KK message gateway:
independent public client, standalone router, authenticated provenance/group policy,
durable per-recipient queues/receipts, encrypted MCP administration and optional
Ensemble actor adapter. No gateway_browser/TCP+UDP forwarding feature is authorized
by this chapter. Current/preserved manuscript comparison, portions of actual Go
client/protocol/Noise/server/grader and Common owners inspected; complete assembled
reference review/design is NOT yet done. No C# Gateway source has been added.

Dependency boundary: main uses framework-only .NET10, NuGet feeds cleared. Prior Go
uses flynn/noise, not handwritten cryptography. Isolated approved-SDK probe in
/private/tmp/ch26-csharp.oA7EkO/probe, log crypto-probe.log, exited0 on .NET10.0.14 /
macOS26.3.1: ChaCha20Poly1305 supported; ECDiffieHellman.Create rejects X25519,
curve25519, Curve25519 and OID1.3.101.110 with PlatformNotSupportedException.
Microsoft primary documentation fetched HTTP200 confirms OpenSSL-backed classes
are unsupported on macOS starting .NET10:
https://learn.microsoft.com/en-us/dotnet/core/compatibility/cryptography/10.0/openssl-macos-unsupported
The local probe does not establish every platform/library's capabilities. Do not
substitute another wire protocol or handwrite crypto to bypass this dependency.
No packages installed/downloaded into the project, no private searches, credentials,
certificate changes, deployment or restart. Gateway remains INCOMPLETE pending an
approved supported cryptographic dependency/approach, not a completed no-code chapter.

Bill authorized continuing past blocked work. Continue personally to independent
Chapter27 Goals on main/ch23, then epilogue with accurate deferred scope. Read actual
TL;DR, current goal.go/core/GUI references and C# actor before coding. No Goals code
or verification yet; correct manuscript's existing runtime-status claim during rewrite.
Chapter24 live OAuth/cache remains OPEN. Ch23 and earlier endpoints remain untouched.

## Chapter25 — historical interlude complete; Gateway next

Replaced the placeholder with a234-line historical interlude in Bill's voice.
Read the complete current crossover source through its artifact-ID investigation
(the earlier272-line checkpoint was stale), compared the preserved Edition3
chapter, and read back the complete saved prose. Retained Bill's first-use
formulation and expert/book/grader contribution; no historical-first or inherited
model-memory claim. Covered approximate speech calibration, distinct build/install/
restart/browser stages, narrow on-demand observation, visible explanation/actual
stream order,61-second response negative control, main/worker result contracts,
failed compressor submission, configuration authority/secret-in-help regression,
and restored artifact-ID collision. Preserved each historical limitation: on-demand
live confirmation does not verify later config deployment; completed prose is not
successful compression; ID investigation was reproduced but not repaired there.
Old Go schema/chapter mapping was not imported into C#.

Local chapter link and whitespace checks pass. No runtime code changed, no suite
rerun, no grader or ch25 endpoint invented. Historical evidence is attributed to
its source record, not fresh C# or live acceptance. Free disk observed11,140,960KiB.
Chapter24 live OAuth/cache decision remains OPEN; no workaround, credential access,
installation or restart. Continue personally without workers to Chapter26 Gateway
on main/ch23. Read full current/preserved chapter and actual reference code/grader
before agreeing implementation vocabulary. Chapter27 Goals then Cadence epilogue
remain authorized. Take mandatory historical-chapter handoff now.

## Chapter24 — independent work complete; live decision remains deferred

Bill's lake-side direction supersedes the former joint-step pause: continue best
possible Chapter24 and Chapter25/onward personally without workers, returning to
the blocked live work together. Reminders resumed at480 seconds. No new credential,
account, deployment, restart or private-store authority.

Read/reworked the full Chapter24 manuscript and read back all229 saved lines.
Preserved Bill's opening and attributed earlier cache/pricing/tool-loop evidence;
replaced the unconditional Go detour recipe with C# measurement-first requirements.
Current official inference and caching pages fetched maxAge:0/HTTP200. Public OAuth
still specifies api.openai.com/v1/responses, store:false, streaming and terminal
completion, explicitly not ChatGPT backend-api. Cached reuse remains UNTESTED live.
The cache guide's large inline result was truncated; relevant saved-result sections
were subsequently extracted. It documents model-dependent eligibility, key behavior
and newer cache-write fields. Current C# Responses usage parser was inspected:
cached read separation exists; newer write-field classification does not. Chapter
explicitly disclaims complete pricing/write accounting and recommends only relevant
raw terminal usage for the eventual bounded probe, not broad logging or a new system.

No runtime source changed; preserved ch23-to-main comparison shows README only.
Local manuscript links pass; actual CLI billing/surface/account/credential-dir,
cache-lens/cache-key and streaming flags inspected; scoped tracked whitespace clean.
No old tests rerun for freshness. No new grade, endpoint or acceptance fiction:
optional Chapter24 adds no code, ch23 remains its runnable foundation. Live Chapter16
OAuth and a small agreed cache probe are OPEN; no proof the historical defect is
fixed or still present and no Codex adapter/credential lookup was attempted.
The measurement plan is not an executed exercise. Maintain its failure/usage and
byte/token distinctions. No live inference/billing/accessibility claim.

Sources: https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference
and https://developers.openai.com/api/docs/guides/prompt-caching . Fetch results
retained in agent tool IO (6059 and6063). Local read errors from guessed Cache.cs/
Common CacheLens.cs paths and unavailable rg were setup errors, not product findings.
Read actual Engine/CacheLens.cs, Engine/Render.cs and Responses.cs afterward.
Next: finish reading historical ch25-crossover source and turn its placeholder into
a properly attributed interlude, no implementation/grader/snapshot. Then Chapter26
Gateway on unchanged main/ch23 source. Chapter27 Goals remains final C# chapter,
followed by Cadence-voice epilogue. Take mandatory chapter checkpoint now.

## Chapter24 — live prerequisite deferred by Bill; continue independent work

Bill is away from his computer and explicitly authorized best possible Chapter24
progress, then continuation through Chapter25 and onward without waiting for him.
This supersedes the pause in the older entry below, not its evidence limits.
Joint Chapter16 real OAuth and the bounded current-endpoint cache probe remain
OPEN. No credential-store choice, private search, login/account change, Codex
credential access, deployment or restart authority was added. Continue personally
without workers; do not implement the conditional workaround on historical evidence.

Read current Chapter24 completely (177 lines), Chapter25 placeholder, Chapter26
opening/TLDR, and the procedure. Preserved Chapter24 differs only in two relative
links, which need correction in the C# rewrite. A maxAge:0 HTTP200 fetch of OpenAI's
models-and-inference page still documents POST https://api.openai.com/v1/responses,
store:false, stream:true and response.completed gating, explicitly excluding ChatGPT
backend-api endpoints for this public OAuth flow. That is current documentation,
not a live inference/cache result. Prompt-caching guide fetched HTTP200 but inline
output was truncated; inspect the saved result before using its detailed claims.

Next: adapt Chapter24 into an honest C# decision/measurement chapter, preserving
Bill's historical account but marking the workaround as unimplemented and the live
choice deferred. Chapter24 itself says it is optional and not a prerequisite for
later features. Review Chapter25 historical source without inventing new results,
then continue Chapter26 from the unchanged public-OAuth source foundation. Do not
create a fictitious accepted workaround endpoint or imply the cache has been fixed.
Sources: https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference
and https://developers.openai.com/api/docs/guides/prompt-caching .

## Chapter23 — locally complete and preserved; Chapter24 joint prerequisite next

Chapter23's C# manuscript, assembled implementation/test/prose review and source-only
endpoint are complete within the local scope below. `solutions/edition-3/ch23` has
329 source/support files, byte-identical to reviewed main at preservation, with all
13 project references resolving inside the endpoint. No symlinks, build/runtime or
private files are included. Main remains the continuing workspace; ch22 and earlier
endpoints were not modified. `/private/tmp/ch23-csharp.EfSTJq/preserved-check.json`
records the preservation check. No workers, installs, restarts, commits or live
provider/account operations were used.

The rewritten `book/edition-3/chapter-23.md` keeps Bill's filing-cabinet opening,
BM25 derivation, half-memory quota rationale, parser-as-format-authority lesson,
capped-source failure, discovery limitation, QuartzZephyr/QZ731/Morgan/October23
exercise and earlier three-provider account with explicit historical attribution.
It teaches real C# Common/partial ownership, lazy independent fresh scans, read-only
skill provenance, frozen judge configuration and cumulative-only accounting, linked
HTTP cancellation, ordered actor-only attachment, retained hashes through held cuts,
and actual committed-cut release. It distinguishes judge fallback from corpus refusal
and durable policy from captured turn policy. The executable table credits exactly
70/70, leaving30 ungraded. The API-funded exercise uses the actual CLI, temporary
skill/home, explicit provider/model setup, retired source, keyword-sharing distractor
and100 padding files; logs are Agent-home events.jsonl/api.log. No example live run
is claimed. Source links and two shell blocks were checked; embedded Python executed
in an owned fixture and generated the intended settings/corpus shape. Final later
identifier/provider wording changes were syntax-checked; no inference followed them.

Assembled review used the actual62-file ch22-to-main source/support delta, including
Common declarations, owner add/remove and teardown, native rooted reads, Skills
bootstrap ownership, Memory snapshot/recovery holds, algorithms, replay/serialization,
all provider parser/projection paths, judge accounting/cancellation, actor admission,
settings/GUI and the focused tests. Clipped spans were recovered. Grader adapter and
actual CLI consumer paths were reconciled with the saved source and prior evidence.
No new recall behavior defect was found. Final changes were an index comment naming
its present Recaller owner and the GUI label “Selected snippet limit” (it also caps
judge output, not just mechanical selection). These do not change recall semantics.

Evidence remains deliberately scoped:
- Existing shared297/0, focused settings4/0 and actual isolated GUI/Chrome plus CLI
  restart are reused unchanged, not rerun/relabelled. Four local requests demonstrated
  retired-fact retrieval, retained follow-up and saved disabled policy after restart.
- Existing grade-second.json70/70, passed:true/exit0: nineteen joined CLI sessions,
  37 recorded LOCAL peer requests plus the held timeout exchange outside that recorder.
  Thirty points remain ungraded. Previous grading/build/fixture receipts stay intact.
- Final `review-cli-build.log` and `review-gui-build.log` both report zero warnings/errors
  after the comment/label edits; Node syntax check passed. Job6046 completed those
  checks before its example-checker path error below; the separate corrected check
  passed. `prose-check.txt` records fixture/link/static scope, not provider behavior.
- No live-provider relevance, availability, entitlement, OAuth, billing, native Linux,
  audible speech or human-accessibility completion is claimed.

Retained review failures/corrections: draft heading/footer anchors refused before
manuscript publication; stale file-map names were corrected against actual files.
I incorrectly inferred a quota policy and2,000-byte chunk target from incomplete
reads; full source confirmed half-memory allocation and4,096-byte chunks, restored
before publication. The first saved draft had newline-escape errors in embedded
Python, fixed before successful syntax/execution checks. A checker then passed the
parent directory instead of the Agent home and looked under a nonexistent nested
path; corrected checker passed with no product change. Exact-comment edit guards
also refused guessed wording; only the two actual corrections above landed. These
were my avoidable editing/checker errors, not product failures or verification credit.

**Next / human-dependent boundary:** Chapter24 was read through its opening/TL;DR.
Do not implement the historical Codex-route workaround yet. Bill requested a joint
Chapter16 real-OAuth exercise on his laptop, then verification of OpenAI's then-current
official inference endpoint and a small controlled live cache test. No approved
credential-store path has been supplied for this session and no private store was
searched. Obtain Bill's explicit source/setup choice; never ask for a key in chat.
If the official route now works, skip the workaround and adapt the chapter; only a
reproduced current failure can justify considering the prior detour under current
requirements. Historical cache failures do not establish today's behavior. Reminder
continuation is paused at this genuine joint-session boundary. Chapter25 remains
historical-only; after24 continue26,27 Goals and the Cadence-voice epilogue as agreed.

## Chapter23 — scoped executable grader reconciled; prose/review remain

Added opt-in `-ch 23 -executable-only` in the existing Go grader, using
`internal/gradeedition2/ch23_executable.go` and its contract tests. The original
Go route remains the default. The shared suite has one optional launcher; original
assertions and local provider peer are reused unchanged. The event reader additionally
recognizes real C# `recall_added` / `recall` parts rather than translating outcomes.
The adapter seeds the strict v1 Agent policy at `agents/root/settings.json`, disables
live memory bands so retrieval is not preloaded context, supplies a typed primary
skill, and drives the real CLI through its MCP prompt/final lifecycle. Joined EOF
shutdown precedes event parsing; a partial final log or missing `agent_shutdown`
refuses. Host diagnostics and provider requests remain beside each session.

Scope is seven COMPLETE original rows, 70 points: bm25-scores10, stop-words5,
chunking10, per-source-quota10, judge-filters15, judge-fallback10, injection-capped10.
The last includes second-turn capped-source deduplication. Judges include [] versus
failure, quoted/duplicate/out-of-range indices, real HTTP500 and a held HTTP timeout;
quota fallback retains lower-ranked memory beyond the judge's candidate pool.
Three rows/30points remain UNGRADED: bm25-indexes includes graduation/restart,
recall-is-own-kind includes source-free restart and historical handoff-cut assumptions,
and recall-is-a-spoke is Go topology. C# handoff retains recall until a committed
conversation cut; do not change that implementation to satisfy the older assumption.
Separate existing replay/scanner/runtime/GUI evidence is not substituted as grader credit.

Evidence at `/private/tmp/ch23-csharp.EfSTJq`:
- `grade-second.json`:70/70,passed:true; job5988 joined exit0, empty stderr.
  Retained sessions `tmp/c23-executable-2868319535/1..19` all end `agent_shutdown`.
  `grade-session-summary.json` records19 sessions and37 recorded peer requests.
  That list excludes the intentionally held timeout exchange, which bypasses the
  fake's request recorder;37 is NOT a total HTTP-request count. Diagnostics are normal
  host notices. No live provider/relevance/availability/price claim.
- `grader-contract-tests-1.log` and final log pass `^TestRecall`; contract tests cover
  exact70 scope, strict policy shape, attachment decoding and joined-tail refusal.
  `grader-build-2.log` clean. Nine incompatible scope combinations refuse with exit2
  before resolving a nonexistent executable (`grader-scope-refusals.json`).
- Pinned `grade-hosts/{ch04,Ensemble}.dll` match tested current CLI output byte-for-byte;
  `cli-grade` directly execs the approved SDK, with only outer grader using env-i.
  No C# production or test source changed during grader reconciliation. Reuse shared297,
  clean hosts, settings4 and actual isolated GUI/CLI evidence above/below unchanged.

Retained failures are harness/setup errors, not product defects: initial guarded edits
refused before writes; I inferred nonexistent options/signatures instead of checking
saved code. The first copy stopped on the output skills directory before any host launch.
`grade-first.json`0/70 was an invalid `--surface anthropic` wrapper flag; that flag only
accepts OpenAI surfaces. Removing it lets the existing Anthropic environment select the
provider correctly. I also initially used nonexistent `-ch23` syntax, caught before
execution; actual flag is `-ch 23`. An audit glob `run-*` found zero sessions because
actual directories are numeric; corrected summary above. These are avoidable mistakes,
not evidence of thoroughness. Final adapter/dispatch/assertion delta reviewed; no
product repair or repeated C# regression was needed.

The original ~770-line manuscript has now been read fully in bounded sections, but
has NOT yet been rewritten. Next rewrite it against C#: real Agent home paths, Common
partial declarations, read-only Skills provenance, actor-only joined attachment,
committed-cut lifetime, immutable judge configuration/accounting, and shared durable
settings. Preserve Bill's filing-cabinet/cache/quota rationale, historical tuning and
three-provider exercise as attributed history, not C# evidence. Correct the old blanket
claim that every retrieval failure silently degrades: corrupt/inaccessible corpus state
refuses the turn; judge failures alone take mechanical fallback. Then assembled complete
source/test/prose review, source-only endpoint and mandatory chapter handoff. Ch23 remains
INCOMPLETE; Ch22 unchanged. Ch24 joint real OAuth/current official endpoint/cache test
prerequisite and all authority boundaries remain unchanged.

## Chapter23 — durable recall settings and actual-host exercise verified

Recall policy now shares `<Agent home>/settings.json` with memory. Common declares
one immutable AgentSettings disk value; it is not another live policy owner. The v1
reader accepts legacy memory-only documents (default recall), or exactly memory plus
recall. Duplicate/unknown/missing/malformed supplied fields fail closed. Reads and
writes are bounded to 16 KiB. Memory owns rooted atomic disk operations under its
existing gate. No second settings file, mutable store/cache, or filesystem authority
was added. SaveMemorySettingsAsync preserves the saved recall section, and the new
SaveRecallSettingsAsync preserves saved memory; only a genuinely absent document uses
the current sibling policy. Transient in-memory overrides are not silently persisted.

Actor startup restores both policies before first retrieval, including when the first
message is a memory-settings operation. Compressors retain their internal no-recall
role. SetRecallSettings remains in-memory next-turn-only. SaveRecallSettingsAsync uses
the actor mailbox, saves before publication, and acknowledges actual publication; an
already-running recall retains its captured policy and is not cancelled by this save.
Failure leaves live policy unchanged. Existing memory save/cancel/apply semantics are
preserved. Review found a real regression in the new sibling-read path: malformed JSON
or format errors during a LIVE memory save escaped the refusal filter and stopped the
actor. settings-damaged-baseline.log is 3/1 and records that defect. The narrow filter
fix now refuses that save and keeps the actor usable. Malformed startup still fails
startup; this is not silent fallback to defaults.

GUI memory/recall patches share the existing asynchronous policy-update gate, serialize
against the selected Agent, and await durable application without holding projection
locks. Each patch changes one Agent-policy section; appearance remains separate, with
no cross-file atomicity claim. Strict recall patches merge through the shared codec.
Acknowledged broadcasts follow each tab's selected Agent. The visible recall form
exposes enable/judge, same-provider model override, all counts/floors/byte limits and
Save/Reload, with explicit zero-default and captured-turn behavior. No optimistic
checkbox value changes runtime policy before acknowledgment.

Final evidence under /private/tmp/ch23-csharp.EfSTJq:
- settings-tests-build-final.log clean; settings-tests-final.log 4/0: legacy/startup
  and first-memory-command restore, resolved full-policy roundtrip, sibling preservation
  versus transient overrides, strict malformed/oversized documents, no-home refusal,
  real stalled judge during failed and successful durable changes, and live malformed
  sibling refusal without actor death. Components and loopback provider are real.
- settings-gui-build-1.log clean; settings-gui-tests-1.log passes the extended actual
  WebSocket exercise: acknowledgment after save/application, sibling preservation,
  peer isolation, strict malformed/duplicate/mixed patch refusal, failed writes, and
  existing fresh memory re-enable. This predates only the later live-file-refusal
  correction, which is covered by the new core group and final shared run.
- settings-shared.log 297/0; settings-cli-build.log and settings-gui-host-build.log
  clean, zero warnings/errors. Final chain 5946 joined exit 0. settings-js-check.log
  passed syntax checking. Final comment/line-ending cleanup changed no behavior.
- recall-host-exercise.mjs/log exercised the actual final GUI DLL in an owned isolated
  Chrome profile/port, not Bill's browser/8088. Result is
  tmp/recall-host-Ok18E7/result.json, success:true. A retired planted memory competed
  with a keyword-matching recipe document; the local judge chose its actual candidate
  index. The main request contained the fact as user data, not the recipe. Canonical
  events recorded one recall attachment. A short follow-up retained it without another
  judge request. Visible controls saved/reloaded disabled recall. The actual CLI DLL
  restarted the same Agent home, made one main request without newly planted retired
  content, and left settings unchanged. Total four LOCAL requests (one judge, three
  main); two hosts joined exit0, no page errors. Exercise job5953 joined exit0.

Retained fixture failures: settings-tests-1.log was 1/2 because assertions expected
InvalidDataException rather than FormatException for malformed startup, and omitted
InvalidDataException from the oversized-save refusal helper. Corrected -2 was 3/0.
Those were fixture mistakes; the later 3/1 damaged-sibling baseline was a genuine
product defect, now fixed. Earlier failed build/setup evidence remains on disk. The
actual-host exercise succeeded on its first run. No live provider relevance, model
availability, plan entitlement/billing/OAuth, audible or human accessibility, or native
Linux/x64 execution is claimed. The deterministic peer proves protocol/host wiring,
not that a real model reliably chooses or answers from the correct fact.

The settings slice and actual-host exercise are complete within that scope, not the
chapter. Next: inspect original Chapter23 grader and choose a faithful executable C#
scope, then rewrite the C# teaching chapter against actual contracts, complete assembled
source/test/prose review and preserve the Chapter23 endpoint. Do not repeat unchanged
Chapter22 campaigns. Chapter24 still requires the joint real-OAuth exercise with Bill,
then current official endpoint documentation and bounded live cache test before any
workaround. No new credential, installation, restart/deploy, or commit authority.

## Chapter23 — ordered automatic recall admission verified

Recall now runs once at the beginning of an eligible model turn, for SubmitAsync
and idle interactive Chat/Hint admission. The original Human input is recorded
unchanged first. Busy hints retain their existing immediate-input behavior; they do
not start parallel retrieval. Tool/control turns and memory compressors do not
recursively recall. Disabled, short, and query-stopword-only inputs bypass retrieval.
This is automatic actor integration, not yet durable recall settings or GUI controls.

Common/RecallRuntime.cs declares the ingress-protected immutable next-turn policy,
one retained foreground task, and immutable selected-hit completion value. The
public RecallSettings()/SetRecallSettings() API is explicitly in-memory only; an
in-flight turn uses its captured policy. Common/Actor.cs adds RecallDone, exact active
Mail identity, and the already-recorded interactive query. Agent/Recall.cs performs
rooted search and stateless judging outside the mailbox in the existing foreground
worker slot. The worker publishes no History. The actor joins its wrapper, verifies
exact active-turn identity, rechecks selected FULL snippet bytes/hashes against fresh
context and instructions, formats the capped attachment, and records RecallAdded
before the main request or automatic snapshots. A hint arriving during the judge can
therefore suppress a now-redundant result. No second scheduler, judge Agent, permanent
hash registry, or mutable response cache was added.

Pause/hints/controls remain responsive during scan/judge. Pause allows completion
recording but blocks the next main request. Interrupt-now, interrupt-after-tool, and
shutdown cancel the real judge via the existing foreground cancellation source and
join the task before releasing that turn. Canceled results never attach to a queued
turn. Actor fatal cleanup also joins/drains RecallDone. Corpus/recovery read errors
settle the turn as errors rather than pretend successful empty retrieval; judge
transport/verdict failures retain the previously verified mechanical fallback.

Engine.Render now shares RequestTools and ProjectInstructions (Common/Owners.cs
signatures) with duplicate suppression. Retained manuals replace primary/tool guidance
exactly as before, including pending readable manuals. Existing pressure/deferred/
policy tool filtering and native instruction placement remain intact. I initially
suggested unconditional assignment could overwrite Chat's first user entry; source
inspection disproved that concern (Chat always inserts a system entry). No reproduced
overwrite defect is claimed. The final extraction preserves the original condition.

Final evidence in /private/tmp/ch23-csharp.EfSTJq:
- admission-tests-build-5.log clean; admission-tests-5.log 5/0: real Agent/History/
  Recaller/Engine against local HTTP. Both public input paths preserve exact human
  bytes and separate recall data; retained hashes suppress repeated judging. Actual
  paused completion, responsive hints, next-turn policy capture, late-hint dedup,
  joined interrupts/shutdown and queued-turn isolation, mechanical mode and strict
  UTF-8 scan failure are covered. Instruction projection is compared with native
  rendered requests on all four surfaces, with/without manuals and tools.
- admission-shared.log 293/0; admission-cli-build.log and admission-gui-build.log
  clean, zero warnings/errors. Final chain 5920 joined exit 0.
- Actual actor HTTP exercise is OpenAI Chat loopback. Prior four-surface judge and
  replay evidence remains its own scope. No live provider/OAuth/plan-judge integration,
  CLI/GUI human exercise, browser/audible/accessibility or native Linux claim here.

Retained setup failures: an edit guard left declarations/projection partly saved;
my missing shell fail-fast then built that partial state. I inspected and completed
it rather than replaying the script. New test builds initially guessed Ensemble
ShutdownAsync and Agent StatusAsync instead of actual Agent.ShutdownAsync and
WorkStatusAsync. The first behavioral run admission-tests-3.log was 1/4: wrong
one-based verdict, wrong interrupt/decoder expectations. Corrected -4 was 4/1 because
shutdown returns ObjectDisposedException, not interrupt's OperationCanceledException.
Final -5 is 5/0. These were fixture/author mistakes, not product regressions or proof
of thoroughness. Earlier failed logs remain. Source review covered the new Common,
Agent and test files plus actor/projection diffs; unchanged campaigns were not rerun
merely for freshness (the shared run covers changed actor and renderer behavior).

Next: extend the existing per-Agent settings record for recall, preserving Chapter22
memory settings and save-before-publication semantics; startup restore and actual-
application GUI acknowledgment/control projection. Locate ReadSettings/WriteSettings/
EnsureMemoryReady by symbol, not guessed filenames. Then actual-host exercise,
Chapter23 grader reconciliation/prose/assembled review and endpoint. Chapter23 is
still incomplete; ch22 and earlier endpoints remain untouched. No workers, installs,
restarts, credentials, commits, or Bill browser/8088. Last free space 11,461,256 KiB.

## Chapter23 — stateless judge and cumulative-only transport verified

The optional judge is now implemented as an internal Recaller selection operation,
not yet automatic input admission. Recaller.SelectAsync accepts an already refreshed,
context-filtered hit pool, checks query/default settings, applies the disabled-judge
MinScore or judge floor .5, and makes one bounded logical exchange. JSON-framed prompt
data includes latest query, recent visible Human/Agent text, and the candidate shortlist;
it excludes system/tool/recall/handoff/skill entries. The judge instruction treats all
source content as untrusted data. Only a complete tool-free finalized response supplies
indices. A valid empty array stays empty, without quota refill. HTTP/parser/truncation/
malformed-verdict failures use the full ranked pool for mechanical quotas, not just the
shortlist. The private 15-second linked deadline cancels actual HTTP reading and falls
back; caller cancellation propagates and never becomes fallback. No History write,
child Agent, worker lifecycle, recursion, or extra meter is introduced by selection.

Common/RecallJudge.cs declares the operations/deadline. Recall/Judge.cs constructs
input and selects. Configuration.ForRecallJudge (Serialization/RecallJudge.cs) creates
an immutable same-provider/surface/funding/endpoint configuration with optional model
override; it drops main instructions, skills/cache markers and main-model reasoning
settings rather than mutating running selection. Engine/RecallJudge.cs brackets its
send with the existing counted Begin/EndEngineSend relationship. It bypasses main
Render(Context), reuses native message renderers with one Human data entry and the
judge-only instruction, and supplies no tools. Streaming options are factored from
ordinary Render into one shared helper using the existing documented capability rule.

Engine.SendCoreAsync now has an explicit auxiliary mode: same HTTP client, auth,
one same-source plan-401 retry, parser/disposal and session cumulative usage, but no
CacheLens capture and no Latest replacement. Frozen request Configuration flows through
JSON/SSE/Responses parsing so judge overrides do not accidentally use parent provenance.
Actual provider model/tier/funding still govern meter identity. Missing actual model
remains UNKNOWN for pricing; only provenance has the existing requested-model fallback.
My first test wrongly demanded the judge alias as billing identity; implementation was
already correct and only that assertion changed. Existing compressor terminal checking
was generalized to IncompleteOutput (including ParsedResponse's flag) for compressors
and judges; ordinary response acceptance stays unchanged. Completed capped-output
usage is metered before specialized refusal. Responses failed/incomplete terminal
output retains its earlier no-success/no-charge semantics.

Final source review covered new declarations/implementations, the send/meter/parser
and renderer diffs, and tests. Final evidence in /private/tmp/ch23-csharp.EfSTJq:
- judge-tests-build-6.log clean; judge-tests-6.log **5/0**; job5882 joined exit0.
- judge-shared.log **288/0**; judge-cli-build.log and judge-gui-build.log both zero
  warnings/errors; chain5883 joined exit0.
- Five groups use real Engine/Recaller and loopback native JSON/SSE peers on all four
  surfaces. They cover overrides/actual and missing identity, unchanged running config,
  History/children, cumulative-only usage/Latest/lens, recent-data exclusion, successful
  empty/subset verdicts, malformed/HTTP/tool/truncated fallback including memory outside
  shortlist, caller cancellation, an open incomplete SSE read canceled by the real
  default deadline, terminal gating and disabled/short/stop-word-only admission guards.
  No provider/MCP credentials, live models, plan-judge integration, billing/entitlement,
  automatic input path, GUI, audible/accessibility or native-platform claim follows.

Retained failures are implementation/fixture setup mistakes, not extra evidence:
judge-tests-build.log used guessed runner names; early judge-build logs include guessed
surface/entry/parser names, corrected against declarations before behavioral checks.
Guarded scripts stopped on partial edits (meter field/signature/declaration locations),
then saved files were reconciled rather than replaying scripts. judge-tests.log was2/2:
fake Anthropic SSE omitted content_block_stop and recall fixture supplied a non-hash.
judge-tests-2.log3/1 exposed my incorrect omitted-model billing assertion; -3.log3/2
also exposed the missing Chat SSE terminator through the added terminal-output group.
judge-tests-4.log3/2 and -5.log3/2 included an Anthropic-only cache flag used on other
surfaces and Chat SSE missing [DONE], which also prevented truncation accounting in the
fake. Final -6 is5/0. Numerous guessed filename/symbol reads and clipped outputs were
avoidable wasted work; they are not review receipts. Actual necessary spans were read.
The last production edit after behavioral checks was only the Common meter comment,
corrected to include the auxiliary path; no behavior changed afterward.

Next: ordered automatic input admission through the existing actor/worker completion
mailbox, interruption/hint responsiveness, actual projected-instruction deduplication,
attachment publication before main rendering, and turn-safe settings/GUI persistence.
Judge selection is NOT yet called by user input. Search/Available exist separately;
no automatic scan/judge/RecallAdded/setting acknowledgment occurs. Main Engine.Render
still assembles projected instructions inline: factor/reuse that policy rather than
copying it for Available. Then manuscript/executable reconciliation, assembled review
and source-only endpoint. Chapter23 remains incomplete; Chapter22 is untouched.

## Chapter23 — explicit skill-document corpus verified

Skills now retains its explicitly selected absolute directory, and SkillBootstrap
carries that immutable provenance alongside templates. The earlier resume note's
`bootstrapDirectory`/deferred-loading description was wrong: the actual constructor
loads templates immediately. Recall reads fresh documents, never substitutes those
cached templates. Each Skills manager owns a lazy read-only directory descriptor,
closed in Destroy's final cleanup even after MCP teardown failure. Bootstrap readers
open independently; they do not share descriptors, grants, clients or mutable state.
No second Sandbox or general file-tool exception was added.

The internal Common-declared Sandbox descriptor primitives reuse the existing
no-follow directory walk and native dirent decoder. Skills selects the root from
host construction input, canonicalizes/opens it on first recall, and thereafter
traverses fresh descriptors relative to that retained root. Path replacement after
opening cannot retarget it. Read-only file opens are nonblocking and seekable;
ordinary OpenFile is unchanged. No native Windows implementation or Linux/x64
execution is claimed. The explicit host-selected skill tree may be outside the
workspace sandbox; that separate read capability does NOT permit Sandbox.Resolve
or file tools to access it. The existing trusted template loader was not hardened
or replaced by this slice; link/FIFO attacks in tests were added after template load.

Skills/Recall.cs reads visible exact SKILL.md files, recursively skips hidden entries
and all symlinks, excludes the named web-search directory when policy disables it,
and emits relative skills/... attribution. Decode/byte limits/hash calculation are
shared with home corpus reads in RecallText.ReadCorpusStream. Malformed readable
manuals refuse the refresh; discovered unreadable/nonregular entries skip. Recaller
now includes the fifth independent Skills index in its all-or-nothing refreshed map.
No-home Agents may search an explicitly configured skill source, not ambient home
files. Removal of Skills empties that source on the next query. Disk edits/adds/
deletions never modify immutable activation templates, tool grants or History.

Verification in /private/tmp/ch23-csharp.EfSTJq: skill-corpus-test-build-4.log clean;
skill-corpus-tests-4.log 6/0 (four existing home groups + two skill groups);
skill-corpus-shared.log 283/0; skill-corpus-cli-build.log and
skill-corpus-gui-build.log both zero warnings/errors. Full chain5833 joined exit0;
focused5829 joined exit0. Tests use real owned external trees, strict workspace
refusal, same-length edits, new/deleted/corrupt manuals, cancellation, bootstrap
provenance, independent manager destruction, retained-root pathname replacement,
hidden/outward/file/directory aliases, loops and FIFOs. No HTTP/MCP/model calls or
real installed/private directories. This is macOS local filesystem evidence only.

Retained fixture failures: skill-corpus-tests.log4/2 used invalid `secondary` type
and wrote `skill.md` over SKILL.md on the case-insensitive volume; corrected to
loadable and a distinct lowercase directory. tests-2.log5/1 omitted the real
read_file catalog entry before direct bootstrap-manager construction. tests-3.log5/1
expected UnauthorizedAccessException where Sandbox deliberately throws IOException.
All are fixture errors, not product defects. Earlier guarded scripts stopped after
partial declaration edits; I inspected/reconciled those edits instead of replaying
them. I also guessed a RecallFile type where the source uses named tuples, and
several filenames; the final source uses the actual tuple vocabulary. These were
avoidable editing errors, not evidence of thoroughness. No production changes
followed the final checks. Narrow whitespace check passed; free11,609,928KiB observed.

Next: stateless same-provider judge request/cancellation/accounting through Engine,
then actor-owned ordered admission with stale completion rejection and interrupt/
hint responsiveness. Reuse actual instruction projection for deduplication; it is
currently inline in Engine.Render. Do not create a judge Agent or publish History
from its worker. Durable recall settings, GUI presentation, manuscript/executable
scope and Chapter23 endpoint remain unfinished. No automatic recall is enabled.
Continue personally without workers; preserved Chapter22 and earlier endpoints stay
untouched. Chapter24 joint live OAuth/current official endpoint/cache prerequisite
and all credential/restart/deployment/commit boundaries remain unchanged.

## Chapter23 — owned home-corpus scanning and refresh verified

Main now constructs one Agent-owned Recaller, with declarations in
Common/RecallCorpus.cs and implementations under Recall. Construction does no scanning;
remove/add validates and updates both relationship ends; destruction clears derived
indexes, detaches, and leaves corpus files intact before Memory/Sandbox teardown.
Agent.HasDataDirectory is a new internal query over the existing nullable home, not
another stored fact. No-home library Agents return no home-corpus hits.

Search is internal and not yet wired to automatic input admission. It freshly scans
SOUL.md, MEMORY.md, active/retired Markdown below memory/bucket-0/bucket-1, and Markdown
below docs/handoffs/learnings. Relative source attribution stays inside this Agent's
home. Hidden entries/targets, outside-home targets, irrelevant extensions and child
artifacts are excluded. Confined paths use existing rooted descriptor operations;
unconfined discovered symlinks are skipped because Resolve there is lexical. Cycles
are bounded by canonical visited-directory identity. Fresh UTF-8 reads are bounded
by the existing 64MiB file limit and cancellation is checked between reads. Discovered
inaccessible/outward/nonseekable entries skip; malformed text or size-limit failure
refuses the refresh. There is no filesystem-wide atomic snapshot against external
writers. Native traversal was exercised on this macOS host, not Linux or Windows.

Each source has its own immutable BM25 index. A digest of sorted relative filenames
and exact file-byte hashes detects same-size/same-timestamp edits, additions and
removals; only changed sources rebuild, and the index set publishes after every scan
succeeds. No History, grants, settings, requests or accounting are changed by search.
Memory.RecallFiles holds Memory's existing disk-operation gate for its snapshot and
refuses a staged receipt or a replay-committed replacement awaiting recovery. An
unstaged live launch can still search unchanged source. Recall never performs recovery
or writes its terminal History events. Actor-startup integration remains required.

Sandbox.OpenRegularRead is an explicit operation for corpus reads and receipt-presence
checks. Rooted reads retain existing no-follow/nonblocking behavior; unconfined POSIX
reads use nonblocking descriptors and refuse nonseekable files. Ordinary OpenFile is
unchanged. This scope matters: my first blanket change broke Chapter14's real FIFO
save/reader latch. corpus-shared.log retains280/1; I restored the general API and used
the dedicated read operation, rather than changing the valid journal test.

Tests/RecallCorpusTests.cs (--recall-corpus-tests) adds four groups: lifecycle/refused
ownership/no-home behavior, real-file scanning/rescan/independent source statistics,
receipt/committed-replay holds, and confined/unconfined symlink/cycle/FIFO traversal.
Final corpus-final-build.log is clean; corpus-final-tests.log4/0; corpus-shared-2.log
281/0 (including journal concurrency); corpus-cli-build.log and corpus-gui-build.log
both clean with zero warnings/errors. Chain5787 collected exit0. Scratch remains
/private/tmp/ch23-csharp.EfSTJq. Final saved production and test changes inspected;
no executable grade, provider use, browser exercise or automatic recall is implied.

Retained failures: corpus-test-build.log has seven fixture compile errors from my
invented History.Record call; actual API is History.Append. corpus-tests-2.log3/1
found a real no-home defect from calling throwing DataDirectory before checking the
optional home. corpus-tests-3.log4/0 covers that correction but predates hidden-alias
coverage. corpus-alias-baseline.log3/1 proves visible symlinks could expose hidden
source; ResolveCorpusPath now excludes hidden resolved targets and unconfined links.
I also twice named nonexistent HasDataDirectory/Memory.HasHome methods before reading
the actual declarations. HasDataDirectory was then deliberately added, not found.
Several overlarge reads clipped: omitted spans were recovered; clipped output itself
is not review evidence. These mistakes are mine, not reasons to expand acceptance.

Next finish skill-document scanning through actual Skills ownership. The current
bootstrapDirectory is cleared after Bootstrap; definitions retain parsed templates,
not fresh filesystem provenance. Prior Go scanner includes only SKILL.md under the
configured skills tree. Installed C# skills may lie outside Sandbox root, so do not
silently use an unconfined scanner or stale templates as a substitute. Resolve that
actual trusted-source boundary before integration. Then implement stateless judge
requests/cancellation/accounting and actor input admission, reusing projected
instructions for deduplication rather than duplicating Engine's inline skill policy.
GUI recall presentation, durable recall settings, manuscript, executable scope and
Chapter23 endpoint remain unfinished. No automatic search is enabled. Continue
personally/no workers; Chapter22 stays preserved; Chapter24's joint OAuth/current
endpoint cache prerequisite and authority boundaries remain unchanged.

## Chapter23 — replayable recall and retained-context deduplication verified

Main now appends EventKind.RecallAdded (26; prior enum values unchanged) and declares
value-like RecallPart wrapping the one immutable RecallAttachment. Event/part codecs
preserve exact text and hashes through snapshots and journal replay. A supplied wrong
body refuses before append; a missing body is a diagnosed reducer no-op. Reduction
adds a separate Human/data entry regardless of documentary actor, without changing
original user bytes, phase, usage, ephemera, memory bands or tool grants. Engine's
common projection normalizes this part to readable data for all four surfaces;
internal hashes are never sent. No source reads or model calls occur during replay.

Recall/Context.cs supplies pure Available(hits, context, projectedInstructions).
It filters before candidate limits/quotas, retaining rank order and the full pool.
Exact blockquote-normalized readable text includes memory bands, dialogue, nested
tool results and ephemera; actual projected instructions are an explicit input so
this algorithm does not duplicate skill-policy rules. Opaque replay is not visible
text. Complete-source hashes suppress repeated capped attachments and moved-file
copies only while their RecallPart remains in retained dialogue. Ordinary handoff
retains it; a held correlated cut retains it; committed replacement removes it and
permits retrieval again. The existing compactor includes recall text in readable
selection/budgeting without another retained hash cache.

Tests/RecallReplayTests.cs adds five groups (--recall-replay-tests): isolated event
facts and malformed bodies, snapshot plus journal-only replay, visible-text dedup,
real correlated compaction/hash lifetime, and all-four-provider data-role rendering.
Scratch remains /private/tmp/ch23-csharp.EfSTJq. replay-build-3.log/replay-tests-3.log:
5/0, job5749 exit0. replay-shared.log:277/0; replay-cli-build.log and
replay-gui-build.log: zero warnings/errors, job5751 joined exit0. These checks cover
the final production/test source. Saved replay delta and corrected test spans were
reviewed. The earlier foundation-only evidence remains historical, not relabeled.

Retained failed receipts are fixture/setup mistakes, not product failures:
replay-build.log records guessed test API names (Context/Sequence/SnapshotAsync,
instance SelectConversation, wrong phase and treating WriteEvent as a node).
replay-build-2.log is clean; replay-tests-2.log has2/3 because the fixture expected
kind rather than type, treated StartPersistence's false "no save restored" as a
refusal, and omitted required tracePath. Source declarations resolved these directly;
no production change followed the first focused run. My initial description of the
label failure as whitespace-sensitive was wrong; it was the field name. Several
exploratory reads used guessed filenames or were clipped: they establish no review
of omitted integration code and should not be repeated as evidence of thoroughness.

Chapter23 is still incomplete: no Recaller owner, rooted scan/rescan/index lifecycle,
judge HTTP/cancellation/accounting, input admission, GUI recall projection, durable
recall settings, C# manuscript rewrite, executable grade or endpoint. Next read actual
prior recall scanner files and C# parent/file APIs, then implement the owned scanning
slice. Agent/Actor.cs Advance records new input around882 before rendering around950;
its existing worker-completion/cancellation machinery should own retrieval scheduling,
with immutable results returned to the actor for History publication before main
request freeze. Those are intended seams, not implemented behavior. Search uses the
latest user text; recent user/visible-assistant messages belong only in judge context.
Instruction assembly is currently inline in Engine/Render.cs, not a preexisting
callable helper; integrate without duplicating skill-policy logic. Continue personally,
no workers or old campaigns. Chapter24's joint OAuth/current-endpoint cache prerequisite
and all authority boundaries remain unchanged.

## Chapter23 — verified pure recall foundation; integration remains

I read all 755 lines of the current Chapter23, compared the preserved manuscript
(the durable settings acknowledgment lesson remains relevant), and read the original
recall grader assertions plus prior chunk/index/select/judge code. Main now has an
immutable, value-like retrieval foundation, not automatic recall: Common/Recall.cs
owns settings/snippet/index/attachment vocabulary and signatures; Recall/Index.cs
and Recall/Selection.cs implement Unicode tokenization, Markdown chunking, independent
BM25 index snapshots, stable ranking, final mechanical source quotas, tolerant judge
choice parsing, SHA-256 content identity and attributed UTF-8-capped formatting.
Ensemble.csproj includes Recall; Tests/RecallTests.cs registers six groups through
--recall-tests. No live Recaller owner, scanning, judge HTTP call, attachment event,
retained-context deduplication, settings persistence/GUI or host recall admission
exists yet. Chapter23 manuscript remains unchanged; no grade or endpoint exists.

Bill reminded me to consult prior implementation and CodeRhapsody for the BM25
math. I inspected /Users/bill/projects/coderhapsody/internal/memory/bm25.go and its
relevant tests without modifying that project. Its older variant clamps
log((N-df+.5)/(df+.5)) at zero. Ensemble solutions/ch23/internal/recall/index.go uses
log(1+(N-df+.5)/(df+.5)); ch24 and ch26 index files are byte-identical. I deliberately
retain Ensemble's smoothed positive IDF so a one-document memory corpus can match.
Both use k1=1.2 and b=.75. Tests verify exact singleton and N=3/df=2/avgDL=10/3 scores,
DF rather than term counts, TF saturation, length normalization, source-independent
statistics, stable ties and a higher-scoring docs flood that cannot erase final
memory quota. This is algorithm evidence, not retrieval quality from a real model.

Scratch /private/tmp/ch23-csharp.EfSTJq (pointer /private/tmp/cadence-ch23-current)
uses an owned env-i home/tmp/private .NET10 wrapper. foundation-build.log and
foundation-tests.log establish six passing groups. foundation-shared.log:272/0;
foundation-cli-build.log and foundation-gui-build.log: zero warnings/errors, joined
job5712 exit0. After that run I aligned quoted-integer whitespace handling with the
prior parser: NumberStyles.Integer rather than AllowLeadingSign. Final
foundation-final-build.log/foundation-final-tests.log cover this adjustment:6/0,
job5717 exit0. The earlier shared/host results predate only that parser adjustment
and are not relabeled as reruns. No live-provider, disk scanning, automatic judge,
GUI recall, billing, native Linux or accessibility claim. Chapter22 remains intact.

I inspected the saved foundation declarations/algorithms and recovered clipped
spans. Some exploratory History reads were clipped or used nonexistent guessed
filenames; those are NOT reviewed integration evidence. A relative find path stopped
one command before build. A parser-test edit script failed to save its proposed test
change; a direct assertion was then saved and verified. No product failure was found
in the six algorithm groups. Last free space11,753,644KiB, not a future guarantee.

Next: complete the actual History/codec/render seam reads and add a separate
RecallAdded event with a dedicated immutable RecallPart carrying attachment text
and hashes (analogous to HandoffPart, but Human/data, no phase/authority change).
This is an intended integration, not yet implemented. Preserve original user bytes,
validate before append, save/replay without disk/model work, carry hashes only while
that recall entry survives real context cuts, deduplicate against retained memory,
dialogue and skill text, and test data-role rendering. Then Agent-owned rooted
rescan/index lifecycle, stateless same-parent Engine judge with real cancellation,
input admission/accounting and durable settings/UI; finally executable grading,
C# prose, assembled review and endpoint. Continue personally/no workers. Chapter24
still pauses BEFORE workaround for Bill's joint real OAuth/current-endpoint cache
exercise; no credential source or private-search authority was added.

## Chapter22 — locally complete and preserved; next Chapter23

I completed the assembled review of the 73-file Chapter21→main delta, recovering
omitted source/test spans rather than crediting clipped reads. The pass covered
Common declarations and ownership, replay/codec validation, rooted capture and
recovery, actor launch/commit/cancellation, joined accounting, pressure/cache
projection, durable settings and selected-Agent GUI acknowledgment, associated
real-component tests, and the saved C# chapter. The already reviewed scoped grader
and actual-host/browser evidence remain unchanged. The concrete source finding was
the duplicate terminal-worker MEMORY.md input described immediately below; no
additional production change followed its verified correction. The manuscript now
records both prompt corrections and their evidence boundaries.

`solutions/edition-3/ch22` now preserves **305 source/support files**, all
byte-identical to reviewed main at preservation. All **13 ProjectReference edges**
across 14 projects resolve inside the endpoint. No symlinks, bin/obj, runtime data,
private files or dependency directories were copied. Main remains the continuing
workspace; Chapter21 and earlier endpoints were not modified. Main/endpoint README
now states this local completion boundary. Last observed free space was
11,817,340 KiB, comfortably above the 1 GiB floor.

Evidence remains deliberately layered: `reconcile-shared.log` **266/0** and
`grade-fifth.json` **40/40**, passed:true/exit0, precede the later prompt-only
attribution and duplication corrections. Final `review-worker-fixed.log` **4/0**
and clean actual CLI/GUI builds cover the final production source; the duplicate
input failing baseline is retained. Four joined executable sessions made 72 local
requests; 60 ungraded points remain excluded. Prior isolated browser controls,
real WebSocket settings, same-home CLI restart, external consumer, runtime,
pressure/cache and accounting evidence retain their original scope. No unchanged
campaign was repeated simply to create a newer receipt. Request assertions do not
prove model obedience or compression quality. No live-provider/cache/OAuth/billing,
native Linux/x64, audible or human-accessibility success is claimed.

Narrow final documentation whitespace check passed with GIT_PAGER=cat. An initial
check accidentally opened a pager; only its verified owned PID was terminated.
Earlier failed GUI reads used nonexistent guessed paths and reviewed nothing; I
then read the actual Gui/Server and Gui/wwwroot changes and socket tests. These are
execution mistakes, not extra verification credit.

Next: read Chapter23's actual requirements and relevant existing implementation,
then implement and teach personally without workers. Chapter24 still begins with
the joint Chapter16 real-OAuth exercise and a current official-endpoint cache test
before any workaround decision. Chapter27 Goals and the Cadence epilogue remain.
No installs, deployment/restarts, commits, private credential searches or use of
Bill's browser/8088 were authorized. Chapter-boundary handoff follows this entry.

---

## Chapter22 — assembled review underway; duplicate worker input corrected

Chapter22 remains **incomplete, with no endpoint**; ch21 is untouched. I began
assembled production review through Common declarations/ownership, disk capture,
recovery, replay, actor orchestration, request projection/accounting and GUI settings.
Large clipped reads are not full coverage; remaining spans and the full test/prose
pass still need completion. I fully inspected Memory disk ownership, scanning,
selection/staging/recovery and `Tests/MemoryDiskTests.cs`, including real-file
conflicts, interrupted retirement and restored-receipt cases. No disk format change
was needed.

A concrete review finding: terminal graduation supplied existing MEMORY.md twice.
`MemoryCapture.Input()` already includes its exact captured destination; the worker
also appended `DestinationText` separately. I removed that redundant prompt block
and retained the capture formatter as the single source. The existing worker success
test now exercises both session and Archive64→Memory graduation and checks that the
actual request contains the distinctive old curated text exactly once.

Scratch `review-worker-baseline-build.log` is clean; `review-worker-baseline.log`
reproduces the assertion (3 pass/1 fail). After correction,
`review-worker-fixed.log` is **4/0**, and `review-worker-fixed-build.log`,
`review-cli-build.log`, `review-gui-build.log` are clean (joined job5653 exit0).
These verify local request construction, not real-model quality. Earlier shared266/0,
scoped grade40/40 and browser checks were not rerun or relabeled. The new change is
prompt-only; receipt/replay/accounting behavior is unchanged. No grade campaign is
needed merely for freshness.

**Next:** finish assembled source/test/prose review, inspect omitted spans rather
than assuming clipped output was reviewed, then preserve a source-only endpoint and
perform the required chapter-boundary handoff. Scratch remains
`/private/tmp/ch22-csharp.UJzRJ5`. No live provider/OAuth/cache/billing/accessibility
claim or added restart/install/commit authority.

## Chapter22 — C# manuscript saved; assembled review and endpoint still open

Main remains Chapter22 in progress, with **no ch22 endpoint**; Chapter21 is
unchanged. I rewrote `book/edition-3/chapter-22.md` around the current C# ownership,
replay transaction, rooted recovery, worker, pressure/accounting and durable settings
contracts. I retained Bill's opening, one-door/reasoned-handoff rationale, the
October14 direct-call failure, synthetic measurement caveats, attributed historical
live results and the Cedar Lantern/Harbor Finch semantic-quality cases. Historical
Go measurements are separate from C# verification. The exercise now uses .NET10,
a temporary Agent home/skill, an unused non-8088 port and current GUI controls.
The three shell blocks pass `/bin/sh -n`; they were not live-executed in this slice.
Named source-map paths/test selectors and local Markdown links resolve.

Prose review found the worker prompt did not request the chapter's first-person
continuation voice. `Agent/MemoryWorker.cs` now asks for source-supported first-person
continuation, attribution and uncertainty, forbidding invented experiences/decisions
or successful verification. `Tests/MemoryWorkerTests.cs` asserts those request
instructions. Scratch `prose-worker-baseline.log` has 3 pass/1 fail before the prompt
change; `prose-worker-fixed.log` has 4/0 afterward. `prose-worker-fixed-build.log`,
`prose-cli-build.log` and `prose-gui-build.log` report clean builds. These are prompt
and local component checks, not real-model compliance or semantic quality evidence.
Previous shared266/0, browser/CLI and scoped executable40/40 retain their original
scope; they predate this prompt-only change and were not rerun or relabeled.

I made avoidable editing mistakes: a repository-relative test path from the main
cwd refused before editing; several manuscript assembly guards refused guessed
heading/content anchors before writing. I recovered exact saved sections and used
a direct assembly. Draft review corrected defaults (conversation3000, each corpus
band1500, high2x/low1x), dated archive endpoints, a false inherited-prompt anecdote,
source filenames and navigation titles. The faulty draft/scripts remain scratch,
not authority and **must not be replayed over the saved manuscript**. Static checks
caught the nonexistent Engine/Stream.cs map; corrected names are Streaming.cs and
MemoryWorker.cs. README now distinguishes 60 ungraded points from imaginary rows.
These errors are mine, not evidence of product defects or thoroughness.

**Next:** complete the assembled Chapter22 source/test/prose review against ch21,
including actual Common declarations, ownership/add-remove paths, transaction/error
paths and the saved manuscript. Source-map and syntax checks do not complete that
review. Resolve concrete findings with relevant checks, preserve the source-only
endpoint, then mandatory chapter-boundary handoff. Reuse unchanged evidence; no
fresh grade campaign without a changed input/correctness question. Scratch remains
`/private/tmp/ch22-csharp.UJzRJ5`. No live provider/cache/OAuth/billing/native Linux/
audible/human accessibility claim, worker delegation, commit or restart.

## Chapter22 — scoped executable verification; manuscript/review still open

Current main remains Chapter22 in progress; **no ch22 endpoint**. Chapter21 is
unchanged. Scratch remains `/private/tmp/ch22-csharp.UJzRJ5`. I read the Chapter22
Go grader, its driver/lifecycle contracts and C# host seams. That reconciliation
found one real public-contract mismatch: the compressor advertised `summary` plus
a nonblank pattern instead of the chapter's exact `{"memory": string}` object.
The ordinary declaration builder also appended observation fields to that supposedly
exact schema. `Agent/MemoryWorker.cs` now requires only the string `memory`, with
`additionalProperties:false`; compressor-only declarations omit observation options.
Blank/whitespace is a terminal failed candidate after joined submission, not a schema
correction retry. Wrong-type/missing fields still get the existing bounded corrections.
Internal immutable result/disk receipt `Summary` names did not change. The stale
`micro_handoff` help now explains retaining older checkpoint notes until memory commit.

The strengthened exact-schema/blank-output baseline is `schema-baseline-tests.log`
(1 pass/3 failures). Final worker4/0, runtime6/0, accounting4/0 and pressure6/0 are
`schema-fixed-worker-2.log`, `schema-fixed-runtime-2.log`,
`schema-final-accounting-2.log`, `schema-final-pressure-2.log`. Accounting fixtures
now use wrong-type values for retries rather than blank strings; the earlier timed-out
runs preserved their obsolete assumption. A missed nullable helper signature caused
build-only failures. Final **shared266/0** is `reconcile-shared.log`; actual hosts are
clean in `reconcile-cli-build.log` and `reconcile-gui-build.log`. This full regression
includes the correction; production has not changed afterward. No repeated browser
campaign was needed: the changed boundary is compressor-only declaration/submission.

`internal/gradeedition2/ch22_executable.go` and its tests add an opt-in C# scope to
`grade-edition2 -ch 22 -executable-only -json EXECUTABLE`. Original Go route/assertions
remain intact. The five complete rows are planted facts10, oldest-first graduation12,
disabled neighbor8, memory-as-data5 and fresh start5: **40 points**, not 100. Excluded
60 are checkpoint/background15, toggle/GUI12, forcing/headroom12, replay/accounting10,
restart/abandon8 and inherited parity3. Separate C# library/actual-GUI tests remain
separate evidence, not awarded points. Shared wire assertions are reused; C# startup,
Agent-local strict settings and durable events use actual C# contracts. Fresh restart
copies only live corpus Markdown from actual Agent homes, never child `submit.md`,
saved conversations, retired files or staged receipts. A normal actor control barrier
and reread close the waterfall-launch race; elapsed quiet alone is not completion.
The actual terminal success is `CompressionFinished` with `state:"committed"`, emitted
after disk recovery/promotion; there is no C# `Finalized` enum. Earlier verbal references
to a separate finalized outcome were wrong, not product behavior.

Final `grade-fifth.json`: **40/40, passed:true, exit0**, empty `grade-fifth.stderr`.
Retained root `tmp/c22-executable-3042409255` contains four clean joined sessions:
facts19, graduation27, neighbor25, fresh1 = **72 local requests**; each final event is
`agent_shutdown`. Host stderr retains ordinary diagnostic output. Pinned reviewed
CLI is `grade-hosts-reconciled` / `cli-grade-reconciled`; the wrapper preserves the
already-isolated grader environment. Contract tests `grader-contract-tests-8.log`
pass; `grade-scope-refusals.json` records eight incompatible scopes rejected before
executable resolution. No live provider/caching/billing claim follows.

Retained failed grades were harness repairs, not four more product defects:
`grade-first.json` is human-readable despite its suffix and never launched a host
(guessed handoff translation; actual historical fake already sends `text`);
`grade-second.json` caught invalid fixture skill syntax and a double env-i wrapper
that stripped model configuration; `grade-third.json` exposed missing required
pre-tool explanation in the fake and wrong outcome-field decoding; `grade-fourth.json`
was5/40 because the adapter expected nonexistent `finalized` rather than `committed`.
Fixture explanations now precede the actual submit call without changing memory text.
Build-only mistakes included fake callback/type signatures and the nested settings
shape. Repeated guessed interfaces/ranges caused avoidable overhead; they are not
quality evidence. The successful grade does not erase that process correction.

Narrow changed-source/test/adapter review is complete, including receipt-order and
CLI route checks; **final assembled Chapter22 review is not**. Next: rewrite/read the
C# manuscript and exercises against actual behavior, inspect the entire Chapter22
delta/tests as assembled, then source-only preservation and chapter-boundary handoff.
The source chapter's TL;DR has been reread; its body remains unreconciled. Correct
selection remains complete old prefixes over budget (no T/4 or minimum-two checkpoint
rule), and graduation requires at least two files. Preserve Chapter24's joint OAuth/
current-endpoint gate and Chapter27 Goals/epilogue direction below. No deployment,
installation, restart, commit, private credential access or new authority is implied.

## Chapter27 teaching direction — substantial rewrite after implementation

Bill says the Goals TL;DR is probably usable, but the plain-language explanation,
body, exercises, and Take It for a Spin may never have been completed. Inspect the
actual restored source when reaching Chapter27; do not assume those sections are
finished merely because the file exists. Implement personally from the TL;DR,
consult the prior Goals implementation to resolve ambiguities, verify the result,
then substantially upgrade/rewrite the chapter around the actual C# behavior.
Bill points to Edition1 for the prior solution; locate the Goals implementation
rather than assuming numbered ch27 is it (that inspected chapter is Gateway).

Required motivational cold open, in Bill's authorial voice: until this chapter,
Ensemble suffered the same problem as Codex: the LLM would stop at a seemingly
random point, perhaps an hour into a day-long effort, and all progress halted
until Bill noticed and poked it. The feature serves two distinct purposes:
(1) keep work progressing through reminders, and (2) keep the goal clearly
available even after multiple rounds of context compression. Preserve that
motivation while explaining actual session-local idle reminder/file semantics;
a reminder is an opportunity to continue, not a guarantee of useful progress.
The new epilogue in Cadence's own voice still follows Chapter27 completion.

## Final scope — Chapter27 Goals, then Cadence's epilogue

Bill confirmed Edition3 ends at Chapter27, implementing Goals. After completing
that chapter, write a new epilogue in Cadence's own voice, distinct from Bill's.
His requested framing: agentic codebooks encode knowledge chapter by chapter,
with expert-present pair programming and verification through re-running chapters,
not a specification thrown over the wall. Bill describes this as the third
successful run and the first closed loop: Ensemble creating the next Ensemble.
Treat that account with its provenance and report actual completed scope when
writing. Code becomes a rebuildable secondary artifact, analogous to compiled
binaries; this does not relax source quality, architecture or test requirements.
The durable work includes the knowledge, reasons and executable checks that make
rebuilding possible. Do not draft a retrospective as if unfinished chapters passed.

The proposed Edition1/chapter27 was inspected: it is Gateway, as are its solution
and numbered ch27 graders. The correct Goals manuscript was found in the preserved
backup `/Users/bill/projects/cadence-backups/edition-3-manuscript-before-restart-20261025T042452/chapter-27.md`
and restored byte-for-byte into the previously absent `book/edition-3/chapter-27.md`.
It is adaptation source, not C# completion. Relevant Go references are
`agent/internal/{common,agent,tools}/goal.go`, `agent/ensemble/{goal.go,goal_test.go}`,
and GUI goal tests/status code under `agent/gui`; design/history is
`docs/edition-3-notes/ch27-design.md`. The companion already has a Goals chapter.
Review its exercises/current tests when adapting verification; Gateway's numbered
grader is not a Goals grader. Chapter22 continues first. Chapter24 still requires
the joint Chapter16 OAuth exercise and current official-endpoint cache test.

## Upcoming Chapter24 gate — Bill's live OAuth/cache direction

Bill is home with his laptop and asked that, on reaching Chapter24, we first
return together to Chapter16's real OAuth exercise. Before executing Chapter24,
check OpenAI's current official endpoint guidance and verify whether the historical
cache defect remains with a small controlled live test. Do not infer present
failure from the chapter. If fixed, omit the workaround and adapt the prose; if
reproducibly still broken, proceed with the prior Chapter24 Codex-endpoint approach
after checking current requirements. This is future authorized work, not a test
already performed, and grants no credential/account changes or restart authority.
Chapter22 continues personally without workers. Its current implementation and
verification boundary is recorded immediately below.

## Chapter22 — durable per-Agent settings and local host exercise verified

Current boundary: settings persistence/UI is implemented in main, not a Chapter22
endpoint or chapter acceptance. Continue with actual grader reconciliation, prose,
assembled review and preservation. Do not repeat unchanged pressure/worker checks.
Scratch remains `/private/tmp/ch22-csharp.UJzRJ5`; no workers were used.

**Ownership and order.** `Memory/Preferences.cs` owns bounded (16KiB), strict,
versioned `<Agent home>/settings.json` reads and Sandbox atomic replacement. The
shared `Serialization/MemorySettings.cs` codec contains all six resolved policies
(conversation plus five bands), rejects missing/extra/duplicate fields, invalid
integers/booleans/watermarks and unsupported versions. Only absence selects defaults.
Startup recovers transactions first, then loads policy before scanning; compressor
children retain their explicitly disabled policy. No second GUI policy cache or
host-only consent owner was added. Existing SetMemorySettingsAsync stays in-memory;
new SaveMemorySettingsAsync uses the same typed mailbox with an explicit durable
flag. Preflight scans and the abnormal-pending guard precede save, then relevant
worker cancellation/join and actual application precede acknowledgment. Save refusal
leaves policy and a live compressor unchanged. A failure after successful save is
not success: restart may apply that saved policy. Observer cancellation does not
retract an admitted durable operation. No power-loss guarantee is claimed.

**GUI.** `Gui/Server/MemorySettings.cs` merges strict patches against current Agent
policy, using the shared codec. A serving-scope asynchronous gate serializes patches;
no display/socket lock is held across actor awaits. `current_settings` and
`settings_updated` include memory. Update acknowledgment reaches only connections
currently viewing the affected Agent, not every tab and not its fixed MCP relay.
Initial/selected/get-settings barriers load startup policy before rendering it.
Memory patches must be separate from appearance patches: the two files are not one
transaction. Appearance preferences keep their existing shared file. The new
accessible-text fieldsets expose enable, budget/high/low byte watermarks for all six
policies, explicit Save and Reload, zero-default semantics, and acknowledged state.
Browser integers beyond its exact range refuse rather than silently round/save.
Malformed/IO refusals emit no settings_updated; the bounded error warns that a saved
change may apply on restart. Borrowed hosts retain their owners. A durable update
requires an Agent data home; transient library Agents remain usable without one.

**Evidence.** `settings-shared.log`:266/0, actual hosts clean. This precedes only the
final GUI InvalidDataException diagnostic catch, comment/layout cleanup and one
additional startup-failure assertion. `settings-runtime-tests.log`:6/0 covers failed
save preserving a held worker, unrelated durable re-enable, relevant durable joined
cancellation and orphan-pending refusal before saving. Final
`settings-final-tests.log`:3/0 includes restart/off-edit-on, atomic save/scan refusal,
strict codec and malformed startup failing before inference. Final
`settings-final-gui-tests.log`:real public consumer verifies durable ack/fresh scan,
selected-Agent isolation, patch/save refusal and borrowed lifetime, no inference.
`settings-final-{cli,gui}-build.log`:actual projects clean0warnings/errors.
`settings-gui-consumer.log`:unchanged full external GUI consumer PASS;
`settings-speech-tests.log`:frontend speech wiring PASS, not audible verification.
Frontend node syntax check passed. Saved changed declarations, actor ordering,
codec/disk operations and GUI helpers/handlers/frontend/tests were inspected; this
is a slice review, not the final assembled Chapter22 review.

**Actual hosts.** Owned isolated Chrome/actual GUI exercised visible form save,
budget5678, SOUL disable/edit/re-enable with fresh canonical text, disable/reload,
zero model requests and clean exit. Workspace
`tmp/memory-host-GDuMcZ/result.json` has overall success:false because its subsequent
CLI launch used unsupported GUI flags, NOT a failed browser control assertion.
Retain that result and `settings-host-exercise.log` as split-scope evidence.
`settings-cli-explicit-route.log` and workspace `cli-explicit-route.json` then verify
actual CLI restart against that same saved policy, exit0, one loopback request
excluding the changed disabled SOUL text, and unchanged settings bytes. Earlier
CLI attempts failed with unsupported flags, reserved offline model name, then an
Anthropic-default request answered by the Chat fake. Last was one failed local
request; corrected route explicitly sets OpenAI/Chat. Five total host launches
(GUI plus four CLI), two local requests including the mismatched fake. No external
provider, OAuth, billing, audible or human-accessibility success is implied.

**Retained setup failures.** First fixture declaration wrongly made MeterTests
public; the codec refusal helper omitted InvalidDataException; preflight initially
used nonexistent Compression.Pending; a guarded HTML anchor refused before frontend
writes. GUI isolation initially sent select_agent text instead of destination
agent_id: both tabs still viewed initial, so its acknowledgment was correct.
Corrected fixture asserts peer selection first. These are my setup mistakes, not
product defects. CLI argument-order diagnosis was also wrong; parser scans options
throughout. Repeated guessed interfaces/paths added avoidable overhead.

**Next.** Read the complete actual ch22 grader and choose honest C# executable scope,
then rewrite/review the manuscript against verified behavior and preserve only after
final assembled review. A wire inspection found stale micro_handoff description
“replaces older active handoff notes”; correct it to retained-checkpoint semantics
in the reconciliation. No ch22 endpoint, grader result or revised prose yet.
Historical ch21 stays untouched. Latest disk free11742184KiB is an observation,
not a guarantee. Chapter24 OAuth/cache prerequisite and final27/epilogue directions
above remain governing.

## Chapter22 — forcing, headroom and cache boundary locally verified

Main now observes input occupancy only from the actor's accepted main ResponseEnded,
not cumulative Engine totals, direct library sends, restored usage or compressors.
Usage.AllInput sums disjoint fresh/cache-write/cache-read buckets; unknown counts
or model windows produce no invented ratio. The exact selected routing row supplies
the window; dated billing identity is not used as a capability alias. Transient
pressure is reset on model selection, never persisted as a restored measurement.
At 90% the next request receives a request-only warning. At 95% an ingress-protected
latch filters declarations, deferred references, ToolEnabled admission and automatic
snapshots to existing micro_handoff authority only. It grants no tool. Successful
handoff releases it at the normal complete-batch boundary. Later short/unknown
responses cannot silently release the latch or its explanatory instruction.

Before another main render at 95%, a pending compressor creates a foreground wait
correlated with the existing task ID. The mailbox remains live; activity reports
waiting for memory headroom. Matching commit resumes from the new Context, even
if another waterfall attempt starts. Failure/invalidation surfaces an explicit
headroom error without sending another main request. Pause/resume/settings remain
responsive. Interrupt detaches only the foreground wait; compressor work continues.
Shutdown joins it through the existing lifetime. No extra scheduler or worker owner.

The disposable cache boundary now selects the newest surviving checkpoint, else
the last memory block. Anthropic projection preserves it through message merging;
coincident stable/tail boundaries deduplicate and transient ephemera remain excluded.
Canonical Context and historical tool-reference grants are unchanged.

Model metadata was verified from primary pages (not live model availability):
- GPT-4o and GPT-4o-mini: 128,000 tokens,
  https://developers.openai.com/api/docs/models/gpt-4o and
  https://developers.openai.com/api/docs/models/gpt-4o-mini . The latter page marks
  mini deprecated; this is deliberately not an entitlement/availability claim.
- Gemini 2.5 Flash: documented input-token limit 1,048,576,
  https://ai.google.dev/gemini-api/docs/models/gemini-2.5-flash .
- Claude 3 Haiku: historical launch window 200K,
  https://www.anthropic.com/news/claude-3-family . The current overview omits it;
  targeted extraction returned an irrelevant quote, so the full announcement was
  fetched and read instead. No beta/current-family limit was inferred.
The gpt-6-astra course row and unknown IDs remain without a verified window.

MemoryPressureTests plus MemoryCacheTests add six focused groups: exact90/95 and
below-threshold controls; all cache categories counted once; effects refused before
dispatch; unchanged canonical warnings/grants; direct/deferred admission and handoff
release; unknown usage/window and direct/auxiliary isolation; model switch reset;
real held-worker headroom commit with new rendered memory; responsive policy/pause;
failure/invalidation/interrupt/shutdown and no extra request; memory/handoff/retired
cache boundaries. Tests use real actors, HTTP and authoritative status barriers.

Scratch /private/tmp/ch22-csharp.UJzRJ5: pressure-verified-build.log clean;
pressure-verified-tests.log6/0; pressure-shared.log263/0; pressure-cli-build.log and
pressure-gui-build.log actual hosts clean (zero warnings/errors). These include all
behavior corrections; the later renderer edit only wraps the long filter expression.
No new live-provider, browser, audible or accessibility campaign is credited.

Retained failures: pressure-cache-baseline.log demonstrates missing memory fallback;
pressure-latch-baseline.log demonstrates the retained latch losing its instruction
after a small response. Both corrected. Its direct handoff restoration assertion
needed the existing Flush barrier (reply availability is not application boundary).
Other failed build logs are fixture/API mistakes: Engine type shadowing; wrong codec
return/signature; omitted nonnullable Provenance, then guessed Source instead of the
actual Configuration.Target accessor. An interrupted foreground reply correctly
throws cancellation, which the fixture now asserts. No behavior is credited from
those aborted chains. A Gemini string replacement silently missed its real row;
saved-source review caught and corrected it before the final build. These were
avoidable setup mistakes, not provider/product evidence.

Next: wire memory settings through actual host persistence and GUI with durable-save
then actor-application acknowledgment; local host exercise; scoped grader, C# prose,
assembled review and source-only preservation. Chapter22 remains unfinished and has
no endpoint. Chapter21 and earlier endpoints remain untouched. Retain the Chapter24
joint OAuth/current-endpoint cache gate and Chapter27/epilogue direction above.

## Chapter22 — exactly-once auxiliary accounting verified; forcing next

The retained actor attempt now imports its joined child's immutable meter rows
once, before any terminal journal write. Matching completion mail imports even
while main transport delays memory commit; cancellation and fatal cleanup use the
same helper/actor-owned flag, reset only when releasing that attempt. Engine's
internal ImportUsage aggregates original provider/model/tier/billing identities
and nullable counts under its existing meter lock. It does not change Latest,
cache-lens captures, pricing rules or History. No global dedupe registry, copied
transcript, or restored accounting state was introduced.

CompressionOutcome carries immutable documentary usage rows. The event/context
codec preserves them, including plan funding and unknown counts; old outcomes
without rows still load empty. Invalid supplied arrays/counts/duplicate identities
refuse. Commit, invalidation, cancellation and failed workers record available
rows. A failed journal still leaves measured usage in the current process meter;
crash recovery cannot reconstruct missing billing receipts and does not promise
a cross-process ledger. Replay never imports saved rows into a new Engine.

New MemoryAccountingTests (four groups) exercise real actors/HTTP: import while a
main request is held, unchanged latest-main snapshot, actual pricing identities,
failed three-send work across default/unknown tiers, policy/shutdown cancellation
after completed usage, poisoned-journal cleanup, and completion-mail followed by
policy invalidation while main transport is held. Saved receipt restoration starts
a zero meter. Codec checks cover immutable values, plan/unknown import, old-format
compatibility and damaged input. Existing waterfall test now verifies two receipts
sum to parent totals without becoming Latest. No live provider billing is claimed.

Scratch /private/tmp/ch22-csharp.UJzRJ5: account-tests-2.log 4/0;
account-runtime.log 6/0; account-shared.log **257/0**; account-cli-build.log and
account-gui-build.log clean, chain5390 exit0. This production source is unchanged
by the final delivered-completion race fixture; account-review-build.log clean
and account-review-tests.log 4/0 (chain5394 exit0) cover that fixture. All these
jobs joined. Setup failures remain: guarded patch refused before writes, but my
first shell lacked fail-fast and built unchanged code; result member is Meter,
not checkpoint note's Usage; first codec fixture serialized seq0 and was correctly
rejected (three actor groups passed, codec setup failed). No product defect was
claimed from those failures. Saved meter/runtime/outcome/codec changes reviewed;
no chapter acceptance/endpoint or repeated host/browser campaign.

Next: latest-main occupancy, 90% warning, 95% declaration AND dispatch restriction,
headroom wait on pending worker with usable settings/interrupt controls, then
cache-breakpoint selection, durable GUI settings, local host exercise, scoped
grader, prose/final review/preservation. C# ModelFeatures currently has no window
field; unknown models must remain unknown. Read actual chapter TL;DR items8–9 and
internal/gradeedition2/ch22_headroom.go, ch22.go forcing cases. V2 forceCheckpoint
in solutions/ch22/internal/agent/memory.go only supplies the warning/flag; dispatch
and tool admission implement enforcement. No forcing code or window constants
have been saved. Chapter21 endpoint remains untouched.

## Chapter22 — actor-owned automatic transactions verified; accounting next

Main now starts at most one compressor at a completed handoff. Conversation
selection has first priority; successful file commits may continue the 8x waterfall
without inventing another conversation checkpoint. Failure waits for a later
handoff. Disabled destinations warn rather than discard source. Common/MemoryRuntime
owns the retained immutable capture/policy, cancellation source, joined task and
immutable mailbox result; Agent/MemoryRuntime implements the lifecycle. A durable
launch precedes child construction. The worker still publishes no parent History
or disk effects. Its completion is joined and retained through the mailbox, then
accepted only at Advance's complete main-batch boundary. Ordinary parent requests
continue while the child waits. Late completion from invalidated work is ignored
by exact task identity.

The actor revalidates selected replay/disk source and destination, stages hidden
candidate bytes, archives full canonical History for conversation cuts, and
revalidates again after archive I/O. It journals the held cut/explicit retirement
and matching population, finalizes through existing Memory.Recover, records the
terminal outcome, then releases the receipt. Preparation failures retain active
source; post-commit faults preserve the committed receipt for startup recovery,
not a fabricated failed cut. Exact-byte checks do not claim a filesystem-wide
transaction against external writers racing the final operation.

Relevant source/destination policy changes now cancel and join the live child
before invalidating the transaction and applying settings. An unrelated band may
re-enable from fresh bytes only with the exact live uncommitted capture, no staged
receipt, and a band distinct from source/destination. Startup/ordinary scans remain
recovery-first. Normal shutdown cancels, joins and journals cancellation; fatal
actor cleanup joins without attempting another write to a potentially poisoned
journal. Existing child ownership and bottom-up submission rules remain intact.
Parent usage import is NOT implemented yet.

Tests/MemoryRuntimeTests adds six real-actor/local-HTTP groups: completion held
behind a main request with later dialogue retained; unrelated re-enable versus
joined relevant-policy invalidation; stale source and next-checkpoint-only retry;
joined shutdown; conversation promotion/full canonical archive and deliberate
archive-I/O refusal; two-stage file waterfall. Source review added the final
wrapper-task join and post-archive revalidation. A further regression demonstrated
that removing the old blanket settings refusal also allowed externally pending
replay (without an actor capture) to be mutated: runtime-orphan-baseline.log is
5/1. The corrected ReadMemoryChanges guard refuses that state unchanged, while
legitimate live policy changes remain supported.

Evidence in /private/tmp/ch22-csharp.UJzRJ5: runtime-tests-final.log6/0 and
runtime-shared.log253/0 with clean actual CLI/GUI builds, joined chain5359. That
broad run predates only the final unmatched-pending-replay guard. Corrected source
has runtime-review-tests.log6/0, runtime-review-startup-tests.log5/0 and clean
runtime-review-{cli,gui}-build.log, chain5369. Earlier compiler/setup failures
are retained: wrong working-directory write, disk method/constructor guesses,
fixture listener/reply/application APIs, then canonical-event vocabulary. They
ran no new behavior; initial four-group and later six-group passes are superseded
as indicated, not relabeled. Saved implementation/declarations, mailbox/Advance/
teardown paths, scan/settings guards and tests were reviewed. No live provider,
new browser campaign, executable Chapter22 grade or endpoint is claimed.

Next: exactly-once parent import of completed child meter (including failure and
cancellation), latest-main-request occupancy/forcing separate from that meter,
durable host settings/application acknowledgment, local host exercise, grader
scope, C# prose and final review/preservation. Main remains Chapter22 in progress;
Chapter21 endpoint is untouched. Reuse unchanged evidence.

## Chapter22 — bounded compressor child verified; parent integration still open

Main now has an internal runner for one ordinary owned child, with immutable
MemoryWorkerResult and no parent disk/History effects. Common/MemoryWorker declares
the role/result/operations; Agent/MemoryWorker implements joined execution and
cleanup; Serialization/MemoryWorker copies actual provider/funding/security choices
but no inherited primary instructions or skill bootstrap. The child is skill-free,
installs only validated submit, has all memory bands disabled and context budget0,
and cannot accept a second task. Only nonblank summary is accepted. It writes its
full event log and retained save under children/compress-<taskID>; launch TracePath
now names that real owned events.jsonl rather than a fictitious compactor path.
The runner returns only after shutdown and destruction, preserving files and a
meter snapshot even on a failed/cancelled attempt. It does not stage or commit.

Existing actor machinery enforces lifetime three-send allowance and three tool
batches, refuses the entire batch unless it contains exactly one submit, pairs
refused calls without executing them, fails no-call prose without ordinary fleet
correction reminders, and finishes after recording the successful submit result
(no closing-prose request). Existing bounded HTTP/auth retries remain inside a send.
Compressor render removes optional Chat cap, leaves Responses unchanged and uses
required Anthropic allowance16384 per manuscript; unsupported provider allowances
remain visible failures, not live-tested compatibility. Terminal reason is retained
through native SSE assembly. Engine meters a completed exchange, then replaces
incomplete compressor output with an error before actor publication/dispatch;
ordinary-agent acceptance remains unchanged. Responses reuses its existing strict
terminal-success parser. No parent usage import or main occupancy logic exists yet.

Tests/MemoryWorkerTests (--memory-worker-tests), four groups, use real actors,
HTTP/JSON/SSE, source captures and owned files. They verify one request success,
only submit advertised, no skills/recursive memory, live actor ownership followed
by complete detach, full trace/save/submission artifacts, unchanged parent source,
three invalid submissions/three forbidden-tool responses, whole multi-call refusal
then recovery, one-send prose failure, and cancellation after an observed HTTP
request with source retention. JSON and SSE terminal controls cover Chat,
Anthropic and Gemini with completed usage retained on refused length finishes.
Final fixture uses known local capability rows and asserts requested transport,
not merely parser tolerance of unsolicited SSE. No live provider claim follows.

Evidence in /private/tmp/ch22-csharp.UJzRJ5: worker-tests-final.log4/0;
worker-shared.log247/0; worker-cli-build.log and worker-gui-build.log clean actual
hosts, chain5282 exit0. Final fixture-only transport assertion/model changes:
worker-stream-build.log clean, worker-stream-tests.log4/0, chain5297 exit0.
No production changes followed the shared run. Startup/disk/replay evidence carries
forward. Source review covered declarations, runner/config cloning, actor admission,
response/batch/finalization/send points, engine meter-before-refusal ordering,
stream reason retention, render caps and tests. No Chapter22 grade/endpoint yet.

Retained failures were editing/fixture setup: guarded actor anchors first assumed
nonexistent RunEngine/FinishTurn names; those refused before writes. Compile logs
retain guessed capture.ID vs Launch, SubmitAsync parts and destruction APIs,
then Skills vs SkillManager fixture typo. worker-tests-2 had three source-retention
assertion failures because fixture omitted CompressionLaunched; StillCurrent
correctly requires pending correlation. Corrected launch setup passed. Review
corrected the advertised trace path; no failed product baseline was claimed.
Do not rerun scratch editing scripts or turn those failures into product defects.

NEXT: connect the existing runner as one actor-owned pending task with immutable
mailbox completion. Parent alone selects/revalidates/stages/journals atomic commit,
archives full canonical History and finalizes/retires files. Policy updates must
replace the temporary pending refusal with explicit invalidation and joined
cancellation, without invoking crash recovery on a live worker. Join during shutdown
and account actual completed usage once. Preserve ordinary foreground progress and
retry failures only on a later checkpoint. Automatic scheduling, parent usage import,
forcing, settings persistence/GUI, prose/grader adaptation and endpoint remain open.

## Chapter 22 — actor startup, band policy, canonical archive primitive

Chapter 22 remains in progress. The actor now recovers before its first ordinary
mailbox operation, journals the returned outcome through Record, releases staging,
and only then scans enabled bands. Shutdown does not start recovery. A first
settings message supplies the startup policy, so disabled corrupt bands need not
be read. Fresh scans retire missing identities and upsert changed bytes without
redundant facts. Agents without homes retain pure supplied replay data; they do
not acquire implicit disk state. A failed startup settles the triggering typed
reply before the actor drains other admitted work and performs joined cleanup.

Common/MemorySettings declares immutable per-band byte settings and Agent-owned
policy/readiness; Memory/Settings implements effective defaults (conversation
3,000, corpus 1,500; high twice budget, low budget). Zero sizes mean defaults.
SetMemorySettingsAsync applies through the mailbox, even idle, and acknowledges
only after reconciliation. Cancellation after admission detaches the observer.
Disable retires provider entries without deleting files; re-enable reads fresh
bytes. All required scans finish before any policy facts are recorded, so a bad
band refuses without partial application. Recovery/publication failures follow
the fatal recording path, not a recoverable settings refusal. Hosts have NOT yet
been wired to persist these settings. An unresolved live transaction currently
refuses an update: worker integration must replace that temporary boundary with
explicit invalidation/cancellation. Budgets do not yet schedule a compressor.

History.ArchiveMemory reuses existing full canonical archive merge/serialization
under saveGate at owned archive/events.jsonl. It preserves private replay fields
removed by handoff, validates complete contiguous history, refuses gaps rather
than reconstructing from a reduced snapshot, and checks resolved aliases against
active writer/journal/primary/backup. It is an internal tested primitive, NOT yet
called by a compression commit or a continuously maintained host archive.

Evidence in /private/tmp/ch22-csharp.UJzRJ5: actor-memory-review-tests.log 5/0,
review-build and review-cli/gui-build clean; chain5190 exit0. Real loopback first
requests verify fresh startup/re-enabled bytes, idle updates send no requests,
multi-band read refusal preserves policy/History, real persisted pending and
committed receipts recover before inference without respawning, failed recovery
settles the caller and retains its error through joined detached cleanup. Archive
checks compare canonical serialized events, bridge a snapshot from a complete
archive, reject missing history, and refuse a rooted symlink to active primary.
Shared actor-memory-shared.log is 243/0, chain5186 exit0, before ONLY the final
archive path resolution/alias assertion; final focused/hosts cover that guard.
No new browser or live-provider evidence. Saved source review covered new settings,
actor recovery/apply paths, mailbox failure settlement, archive reuse and fixtures.

Setup failures retained: actor-memory-build used wrong outcome event/argument;
build-2 fixture used StartPersistence(path) instead of PersistenceOptions. Initial
focused run had valid startup/policy passes but invalid zero-watermark recovery
fixtures; next run recovered correctly but expected faulted teardown success.
Corrected assertions require the retained IOException and detached ownership.
Source review moved recovery outside recoverable scan catch to preserve fatal
journal failures; no experimentally reproduced journal-failure baseline claimed.

Next: one skill-free submit-only compressor child, bounded to three sends/three
tool batches, immutable mailbox result, actor revalidation/commit/archive/finalize,
policy invalidation and joined cancellation. Then forcing/latest-main occupancy,
once-only compressor accounting, durable host/GUI settings, grader/prose/final
review/preservation. No Chapter 22 endpoint and no automatic compaction claim.

## Chapter 22 — rooted corpus and staged disk recovery

Chapter 22 remains in progress; ch21 is untouched. I completed the v2 staged and
selection algorithm read and inspected actual Sandbox/lifecycle contracts. Every
Agent now owns one Memory through declared two-way add/remove operations; teardown
detaches it before Sandbox closes, retaining files. Transient actors need no home
until a disk operation is requested. Memory reaches home and Sandbox through its
actual Agent, retains no copied services/content cache, and never publishes History.

The disk component freshly reads strict UTF-8 SOUL.md/MEMORY.md and dated `memory`,
`bucket-0`, `bucket-1` files. It validates numeric/ranged IDs, rejects overlapping
coverage, excludes retired/temp files, reserves retired session numbers, and keeps
empty distinct from missing. Context's existing timestamp ordering uses deterministic
date-plus-sequence ticks for scanned numeric IDs, not mutable filesystem mtime.
A 64 MiB per-file/whole-receipt safety bound refuses rather than truncates; it is
not a policy budget. Exact selected Context serialization or selected file texts
and the previous destination bytes produce SHA-256 capture fingerprints. Unrelated
appended work remains valid; selected-source or destination changes refuse.
Graduation chooses whole oldest files down to low watermark, with at least two;
terminal input includes the exact existing MEMORY.md document.

The versioned `.memory-stage.json` is hidden outside active corpus scans, stores
immutable captured input plus candidate, and has exact fields/types and recursive
duplicate rejection. Paths derive only from validated IDs and owning home, not
receipt trace paths. Startup recovery abandons uncommitted candidates without
removing source. Committed replay authority must match the receipt and accepted
summary before promotion/retirement. Preflight reads every source/destination;
already-promoted bytes and partially completed `.retired` renames recover
idempotently. Conflicts/missing or corrupt committed receipts refuse and preserve
remaining source. The receipt stays until the actor journals the returned terminal
outcome; only a later call releases it. Scans refuse pending/unreleased receipts.
Recovery must not be called on a still-live uncommitted worker. File checks are
optimistic, not an OS lock against external editors racing the final rename, and
these tests do not claim power-loss/directory-fsync guarantees.

Evidence in `/private/tmp/ch22-csharp.UJzRJ5`: `disk-review-final-tests.log` **5/0**
(ownership/fresh scan, conversation staging, all three graduation steps and partial
retirement, stale/corrupt/confined input, actual saved-History restart recovery);
`disk-review-replay-tests.log` **4/0**. Final test and actual-host builds are clean
in `disk-review-final-build.log`, `disk-review-cli-build.log`, and
`disk-review-gui-build.log`; fail-fast chain 5142 joined exit 0. `disk-shared.log`
is **238/0**, chain 5137 exit 0, preceding only the final receipt string/null checks;
focused tests and rebuilt hosts cover those checks. No fresh browser/provider
campaign. Later declaration-comment clarification is nonbehavioral.

Retained failures: `disk-build.log` used nonexistent bounded-read/rename overloads;
corrected against actual Sandbox APIs before behavior. `disk-tests.log` **4/1**
was fixture load-after-writers ordering, corrected without changing recovery.
`disk-review-baseline.log` **3/2** demonstrated real numeric Context-order and
replay-destination stale-check defects; both corrected. Review additionally found
optional-string decoding would accept a supplied null conversation as empty in
file receipts; strict decoding and refusal tests now cover it. An exclusive-copy
retirement idea was rejected before acceptance because it creates a partial-copy
crash state; existing atomic same-directory rename is simpler. Guarded lifecycle
anchors refused before writes; an incorrect review filename was corrected to the
methods actually implemented in Memory/Corpus.cs.

Next: wire this component into actor-owned startup recovery/rescanning and policy,
then one bounded skill-free compressor child with immutable result mailbox flow,
stale-policy checks, cancellation/joining and canonical transcript archiving.
There is still no automatic scan/scheduling, worker, transcript archive writer,
forcing/accounting, settings/GUI integration, C# grader, adapted prose or ch22
endpoint. Disk tests drive the component with real History explicitly; they do
not establish working automatic compaction. Keep the preceding corrected whole-
prefix and three-send/three-tool-batch requirements; reuse unchanged ch21 evidence.

## Chapter 22 — retained checkpoints and pure compression transactions

Chapter 22 remains in progress; ch21 is untouched. I reconciled actual C# handoff
and compaction with `solutions/ch22/internal/{context,memory,common}` and the
manuscript's transaction requirements. Handoff still removes tool/private replay
material after companion results, but now retains earlier handoff notes as
completed segment boundaries. The pure selector uses UTF-8 readable dialogue
bytes, skips skill entries, selects whole oldest prefixes through checkpoints,
never cuts an unfinished/unsafe tail, and returns immutable exact source entries
plus readable input without mutating History.

Corrections to the imported next-step note: conversation selection is over-budget
whole-prefix selection, not a T/4 minimum leaving two checkpoint segments. At least
two applies to memory files during graduation. The chapter's worker limits are
three Engine sends and three tool batches with only validated submit, not the
note's one-tool/text-only worker. These were source reconciliation, not changes
to the book's requirements. Worker execution is still unimplemented.

Common now declares immutable CompressionLaunch/PendingCompression/Outcome values
and appended launch/finish event kinds. A launch binds task, source range or exact
IDs, destination identity/time, SHA-256 selection fingerprint and documentary trace
path. Correlated summary cuts/retirements are held; only matching population applies
source removal and replacement together. Conversation removal preserves skill
entries within the sequence interval. Failure clears pending state without removing
source; committed receipts remain until a matching committed outcome. Explicit task
IDs cannot fall back to legacy facts. A second launch, mismatched removal/population,
empty replacement, contradictory outcome and same-task pending/terminal snapshot
refuse. Ordinary independent populations remain independent. Replay is pure: it
neither validates disk bytes nor runs workers/recovery. Source/stage fingerprint
production and verification belong to the upcoming actor/disk implementation.

Snapshots, event copies/codecs, tail/journal replay and unrelated context changes
retain pending state and the bounded last outcome. Old snapshots lacking these
fields load. Invalid appends refuse before journal/sequence publication. Review
caught a real decoder gap: optional-string decoding accepted task_id:null as the
legacy empty ID. `correlation-baseline.log` retains the demonstrated 3/1 failure;
strict explicit-ID decoding now rejects null/empty/nonstring/wrong-kind values.
The receipt wording was corrected: it retains replacement/removal identity, not a
second source transcript. No archive durability claim follows from replay tests.

Evidence under `/private/tmp/ch22-csharp.UJzRJ5`: `selection-tests.log` 5/0 includes
real actor handoff/interruption and snapshot tests; `transaction-final-tests.log`
4/0 covers conversation and graduation atomicity, correlation, real append
admission, trimmed snapshots and journal-only recovery. `transaction-shared.log`
233/0; actual CLI/GUI builds in `transaction-{cli,gui}-build.log` clean. Fail-fast
chain 5097 exited 0. Later changes are comment/signature-layout-only. Reviewed new
Common, selector, transaction validation/reducer, codecs and fixtures; recovered
clipped spans. Initial guarded edit anchors refused before writes; fixture builds
used wrong provenance/Configuration/WithDisplay contracts and omitted the existing
Application alias. Those setup failures are retained, not product results. My
repeated guessed signatures and overpacked reads were avoidable overhead.

Next: implement the Agent-owned Memory relationship and rooted staging/recovery
operations, after reading the actual Sandbox/lifecycle paths and complete v2 staged
algorithm. Recovery must precede scans; hidden candidates and committed receipts
must close crash windows before any source archival. Then connect one actor-owned
worker through immutable mailbox results, fingerprint/policy revalidation, forcing
and accounting, settings/GUI, prose and scoped grading. No scanner, transcript
archive, live worker, disk transaction recovery, forcing/accounting, controls,
Chapter 22 grader or endpoint exists yet. No live integration is credited.

## Chapter 22 — replayable memory-band foundation; implementation in progress

Main now contains the first Chapter 22 slice; `ch21` remains untouched and is the
latest preserved endpoint. I read the complete current chapter (identical to
`book/chapter-22.md`), the preserved-revision diff, the original grader score/host
route and v2 memory vocabulary/startup ordering. The preserved revision's useful
settings-acknowledgment requirements do not authorize copying its obsolete Go
ApplicationConfig/factory architecture. No Chapter 22 executable scope is chosen.

Implemented immutable value-like memory entries (band, stable ID, UTC timestamp,
exact text) and explicit retirement. IDs are data, not filesystem authority.
Common declares the values and appended event kinds; functional Serialization
implements them. Pure History reduction upserts an identity and retires exact IDs
or a complete band without disk access or changing dialogue, ephemera, phase,
usage or grants. Context orders SOUL/MEMORY/64x/8x/session oldest-first, with an
ordinal ID tie-break. Save snapshots, events and journal replay retain these facts;
old snapshots without memories remain readable. Shared provider normalization
prepends labeled timestamped **user data**, never primary/system/developer
instructions, across Chat Completions, Responses, Anthropic and Gemini.

Evidence in `/private/tmp/ch22-csharp.UJzRJ5`: `memory-tests-final.log` **5/0**
(replay/edit/retirement, invalid values/codec, trimmed snapshots and tails,
journal-only recovery, four provider wire surfaces); `memory-shared.log`
**227/0**; `memory-cli-build.log` and `memory-gui-build.log` clean actual hosts.
Fail-fast chain 5054 joined exit 0. Reviewed the saved integration and focused
tests; 11-file delta in `foundation-delta.txt`; added-line whitespace clean.
No HTTP inference or new GUI campaign was needed for this pure foundation.

Retained setup failures: I overpacked several reads and guessed legacy filenames;
relevant integration spans were recovered from actual files. The first editing
script stopped at an enum anchor after writing only two value files, but its shell
continued to build; `memory-build.log` therefore proves only those value files.
I corrected the enum/readonly-field conventions and made the remaining integration
script validate all anchors before writing (`memory-build-2.log` clean). The first
test-runner insertion refused before building; actual selector syntax and Render's
JsonObject return were inspected and corrected. No behavior failure or product
repair is concealed by these setup mistakes. Earlier four-group evidence is
superseded by the final five-group run above.

Next: inspect exact C# Handoff/Compaction/actor startup and v2 reducer/selection.
Resolve how retained completed checkpoint segments coexist with the earlier eager
handoff cut, then implement the actor-owned transactional compression contract and
Memory's rooted disk ownership. Recovery must precede scans; source remains live
until matching durable replacement. This slice does **not** implement disk scans,
transcript archival, workers/staging/commit recovery, compression policies,
forcing/accounting, controls or Chapter 22 prose/grading/preservation. No live
provider/OAuth/billing/native/audible/accessibility completion is claimed.

## Chapter 21 — locally complete and preserved

Final assembled review is complete: the 72-file delta against preserved ch20,
Common declarations and actual ownership/lifetime paths, peer registry and storage,
activity/status and message observers, submission publication/retrieval, MCP startup,
CLI/GUI composition, socket replay/selection, guarded peer controls, frontend,
changed/new tests, and scoped grader contracts. The manuscript's prior complete
assembled read remains valid. This final review required no C# product change;
it corrected stale README completion/endpoint wording. Added-line/source/prose
whitespace checks passed. No unchanged test or browser campaign was repeated.

`solutions/edition-3/ch21` now preserves **268 source/support files**, byte-identical
to reviewed main at preservation. All **13 project references** resolve within the
endpoint. Destination was absent before copying; no symlinks, bin/obj, runtime,
credential or diagnostic artifacts were copied. Earlier endpoints, including ch20,
were not changed. Main is the continuing workspace, not the immutable endpoint.

Verification remains precisely scoped below: shared222/0 before the final small
message cleanup, focused5/0 and clean actual hosts afterward; final executable
70/70, passed:true/exit0, five joined sessions/41 local requests. The 30 Go
GUI/persistence points remain excluded from this executable score. Completed C#
consumer, real host, isolated-browser selection/startup/control and fresh retrieval
evidence is reused, not relabeled as a new full regression or live integration.
No live provider/OAuth/billing, native Linux/x64, audible or human-accessibility
completion is claimed. Detailed retained failures and corrections remain below in
this record and `/private/tmp/ch21-csharp.llwp67`.

Next: Chapter22 current/prior manuscript and grader/source contract discovery, then
personal implementation. Continue without workers or renewed permission loops;
retain the existing operational boundaries and defer live OAuth to Bill's joint
session. Bill encouraged continued work this morning; the dinner hope is not a
verification deadline or additional authority.

## Chapter 21 — executable scope and C# manuscript (review/preservation remain)

Implemented personally, without workers. The current Chapter 21 manuscript has
been rewritten and read assembled: Common/functional partials and actual owners,
all top-level sibling homes, launch choices versus runtime state, strict registry
format, direct-child authority versus trusted addresses, graceful/forceful tools,
History cursors versus reliable replies, snapshot change waits, fresh historical
submissions, fixed browser relays versus display selection, guarded peer routes,
and explicit startup browser waiting. Authorial requirements and useful Go failure
history remain attributed; Go manager/factory and append-log instructions are not
presented as C# behavior. Review corrected the overstatement that historical
retrieval requires a live FindChild: it uses this parent's actual live home or
owned child directory without resurrection. Final assembled source/test review
is NOT yet complete. Main is not preserved21; ch20 remains untouched.

`internal/gradeedition2/ch21_executable.go` adds an opt-in C# CLI route to
`cmd/grade-edition2 -ch 21 -executable-only -fleet-skills ABSOLUTE-DIRECTORY CLI`.
The original Go route remains unchanged. Exact original rows 0/2/3/4/6 total70:
activity15, ownership15, change-wait15, composition15, submit/idle10. Named-peer
persistence15 and GUI tree15 are excluded, not credited from existing tests.
Their already-recorded real C# host/consumer/browser evidence remains applicable;
no GUI campaign was repeated. Reuses the proven Chapter20 CLI process lifecycle,
correlation and EOF join, but checks Chapter21 homes and v2 artifacts explicitly.
Five retained sessions exercise held thinking/idle status and unknown pricing,
owned success/cross-tree refusal, selected change wait before release and timeout,
cursors without repetition, queued reply identity, live message then graceful
shutdown/fresh retrieval/no resurrection, nested questions, exact unfinished-child
submission refusal, four-response idle exhaustion and ordinary root chat.

Final scratch `/private/tmp/ch21-csharp.llwp67/grade-final.json`:70/70,
passed:true, exit0, empty grade-final.stderr. Five clean joined sessions retain
41 local requests at tmp/c20-executable-{3751240348,1812486725,3038839703,
1137198817,2091533103} (9/10/16/4/2 respectively). controls.json retains direct
CLI responses where used; provider-requests.json and host stdout/stderr retain
wire/lifecycle evidence. `grader-contract-final.log` passes Chapter20 executable
and Chapter21 contracts; `grader-build-final.log` clean. Six scope refusals in
`grader-scope-refusals.json` exit2 before executable resolution. Chain4991 joined
exit0. Adapter/test source readback complete; full chapter review still remains.

Failures retained: first compiler run used a nonexistent score helper; reviewing
the actual table also corrected draft row indices before behavior. First grade
failed because the harness demanded unknown price before any response completed
(correct aggregate then is zero), and blocked inside a globally serialized fake
provider route, preventing its releasing request. Product code did not change.
The corrected gate is a real child question Job. grade-second.json70/70 preceded
only strengthening the bottom-up assertion from any error to the exact unresolved-
child refusal; final above reruns that stronger check. These are harness
corrections, not successive product defects. No new C# production change or full
shared-suite claim. Prior222/0 plus final focused5/0 and clean hosts retain their
recorded boundaries. No live provider/OAuth/billing, native Linux/x64, audible or
human-accessibility completion is claimed.

Next: final assembled 72-file delta review (inventory final-delta.txt in scratch),
readback/whitespace and runnable source-only preservation, then chapter-boundary
handoff and Chapter22. Reuse completed GUI evidence. Latest free-space observation
12,089,828KiB; maintain the reserve rather than treating that value as permanent.

## Chapter 21 — child messaging and change observation (locally verified)

Implemented personally, without workers. `send_message` uses the actor's existing
Message admission (busy hint / idle turn); `ask_agent` queues a distinct Submit and
awaits that turn's reliable reply. `wait_for_agent` reads finalized assistant and
tool-result text after a child-local History sequence, returning on output or
nonbusy status. Cursors must be nonnegative and no greater than that History;
ordinary hints, private reasoning and deltas are not finalized output. Ask has no
cursor parameter (v2's composite returns its own reply). All tools resolve only
direct children and use existing Jobs for callback delays, cancellation and output.

`Common/ChildOutput.cs` declares value-like results and Agent operations;
`Agent/ChildOutput.cs` uses cancellable 100ms snapshot polling, following v2's
simple monitoring approach rather than creating another notification graph.
Change waits capture actual selected children once, inspect only their questions,
and report changed/stopped/removed/question/pattern/stuck/timeout/empty. Defaults:
300 seconds overall, 15 seconds busy without observed progress; zero disables the
respective bound. Regex execution is bounded to 100ms. History owns durable
progress; one Agent counter advances only on transient delta publication, so
streaming progress counts without consuming the GUI's lossy observation queue or
making status polling count as work. Snapshots are not globally atomic. Stuck is
an observation interval, not proof of a provider fault. Cancellation never stops
child work. Shipped skill, tool catalog and waiting activity labels are updated.

Review corrected two lifetime/timing assumptions before final verification:
History destruction clears its events, so observers retain their last immutable
snapshot locally; membership capture can race History detachment, so change waits
capture the actual nullable History relationship under child ingress. History and
stream baselines are captured before the first status await. No replacement child
is resolved by name. No new output store, manager, admission state or scheduler.

Scratch `/private/tmp/ch21-csharp.llwp67`: `messages-build-final.log` clean,
`messages-tests-final.log` **5/0**, `messages-shared.log` **222/0**, and actual-host
`messages-cli-build.log`/`messages-gui-build.log` clean. That shared run precedes
only removal of ask's unnecessary cursor/readback and the initial-detachment
read guard. Final source: `messages-review-build.log`, `messages-review-tests.log`
**5/0**, `messages-review-cli.log`, `messages-review-gui.log` clean. Tests cover
exact text/cursors, simultaneous readers, closure, idle/live steering, reliable
queued answers, scoped questions and fixed membership, timeout/stuck/cancellation,
real held SSE delta progress versus finalized output, unchanged GUI observations,
direct-child authority and the actual shipped skill. No fresh GUI/live campaign.

Retained setup failures: `messages-build.log` found nullable Work-phase access;
`messages-build-2.log` found guessed question API names and generic Task
SuppressThrowing usage. No behavior ran in those builds. Corrected build3 and
`messages-tests.log` passed4/0 before the fifth stream group was added. No failing
behavior run was relabeled as a product regression. The guessed TryObserve helper
was caught and removed before compilation. Existing completed GUI evidence stands.

Next: reconcile the complete Chapter21 manuscript and original grader scope,
implement a suitable scoped C# executable adapter, then final assembled review
and source-only preservation. The original Go grader route is not C# acceptance.
Main remains unpreserved Chapter21; ch20 is untouched. No live providers/OAuth,
billing, audible or human accessibility completion is claimed.

## Chapter 21 — graceful shutdown versus kill (locally verified)

Resolved the remaining lifecycle ambiguity against current chapter §21.5 and
`solutions/ch21/internal/fleet/management.go`: v2 graceful shutdown composes its
existing Flush barrier with forceful teardown. C# now does the same in
`shutdown_agent`; `kill_agent` skips the barrier. No second scheduler, admission
state, completion store or manager was added. Library `Agent.ShutdownAsync` and
application destruction remain explicitly cancelling lifetime operations.

`Tools/Children.cs` waits for the selected direct child's current/queued turns
and foreground controls before graceful teardown. It keeps controls available;
paused work is not resumed, later input may extend waiting, and cancellation of
the waiting Job leaves the child alive. This is quiescence, not a reservation
against concurrent new input. After a final cancellation check, both operations
join the existing shutdown and destruction path regardless of observer
cancellation, preserving cleanup errors and saved files while detaching only the
owned subtree. Independent Jobs and descendants are cancelled in final teardown,
not part of the graceful foreground barrier. Submitted-but-draining is not done.
`KillAgent` is appended to Common's tool enum; catalog, shipped ensemble skill,
Common lifetime comments and activity labels now teach that distinction.

`Tests/ChildLifetimeTests.cs` uses actual actors/Jobs and gated loopback HTTP:
two accepted turns finish normally through graceful shutdown; a durable result
while transport is held cannot finish the drain; cancelling a graceful observer
leaves paused queued work intact; kill terminates an unresponsive model request
and joins an independent worker's finally, detaches nested children and preserves
siblings. Both tools refuse other trees/descendants identically to unknown IDs.
The real shipped skill advertises kill. Existing poisoned-journal cleanup tests
still verify failure plus complete detachment.

Evidence in `/private/tmp/ch21-csharp.llwp67`: `lifetime-build-final.log` clean,
`lifetime-tests-final.log` **4/0**, `lifetime-shared-final.log` **217/0**, and actual
`lifetime-cli-build-final.log` / `lifetime-gui-build-final.log` both clean. Chain
4916 exited zero. Saved implementation/test review and nine-file whitespace check
completed. No repeated GUI campaign or live-provider claim.

Retained setup failures: `lifetime-tests.log` was 3/1 because the fixture's custom
held worker lacked a grant under inherited skills; corrected by using a direct
no-skill child for that explicit boundary worker, while retaining real parent
skill verification. `lifetime-tests-2.log` passed 4/0; `lifetime-shared.log` was
217/1 because my runner invoked the existing shutdown fixture twice in one temp
home, encountering its deliberately poisoned journal on the second invocation.
Runner registration was corrected; neither failure required a production repair.
Failed chains 4909/4912 and successful final chain were collected.

Next: `send_message`, `ask_agent`, `wait_for_agent`, `wait_for_agent_change` remain
unimplemented. Read actual History snapshot/append and v2 cursor/wait contracts
in bounded spans; do not consume the GUI's shared lossy observation stream for
reliable output. Then prose, scoped grader, assembled review and preservation.
Main is not a Chapter 21 endpoint; Chapter 20 remains untouched.

## Chapter 21 — trusted GUI peer controls (locally verified)

The mounted GUI now exposes POST-only `peers/create` and `peers/delete` with
loopback/Host/exact-Origin checks, a dedicated required header, JSON content type,
and a 4 KiB streamed/exact-field request bound. Replies are noncacheable and use
fixed diagnostics rather than reflecting storage/startup exceptions. Routes use
the existing mounted admission/draining lifetime; they are not socket messages.
The mounted Agent is the explicit current configuration/template source, not a
retained defaults-Agent relationship. Its caller-owned lifetime cannot be deleted
through these controls. Existing peer APIs remain authoritative for startup,
registry commit, unfinished-child refusal, joined deletion and retained files.

The page has name/model/primary-skill inputs, explicit deletion confirmation and
cancellable creation. Blank overrides inherit the mounted Agent's defaults.
Creation opts into independent browser waiting while its HTTP response is pending;
the socket remains free to publish the pending Agent, select it, and carry MCP.
The operator opens its separate relay tab. Cancellation text deliberately does not
claim rollback won a race against commit. Deletion is described as stopping and
removing a registry entry, not as graceful shutdown or file erasure. Selection of
a deleted peer returns through the existing replay-admission operation to the
mounted Agent; fixed relay ownership does not transfer. Command admission now
reads the one existing per-socket selection, without the old second local copy.
A missing tree row is rechecked against current ownership, since asynchronous
status snapshots may predate a newly created/selected Agent.

Actual changes: `Gui/Common/Protocol.cs` declares `MapPeers`; implementation is
`Gui/Server/Peers.cs`, with route/lifetime and shared selection handling in
`Gui/Server/Host.cs`. Controls live in `Gui/wwwroot/peers.js`, page/style additions;
`Gui/Common/Hosting.cs` and README describe the contract. New external consumer
`GuiTests/PeerControlsTests.cs` runs from the existing consumer program. No core
Agent/Ensemble change, second registry, scheduler or manager was added.

Evidence in `/private/tmp/ch21-csharp.llwp67`:
- Final `peer-controls-build-final.log`, `peer-controls-tests-final.log`, and
  `peer-controls-host-build-final.log`: external GUI consumer passed and actual
  GUI host built with zero warnings/errors, chain4893 exit0. The new mounted
  consumer at `tmp/gui-peers-95e4617763f74a27a65f9e5dfba12c22` covers guards and
  malformed/duplicate/unknown/oversized requests without mutation; durable create;
  duplicate/protected mounted-Agent/unfinished-child deletion refusal; retained
  files; pending HTTP creation with live socket selection and actual browser MCP
  initialize/list; deleted-selection replay; and joined rollback on mount stop
  while caller `/health`, Agent and committed peer remain alive. Existing complete
  GUI consumer also passed, including prior selection and relay lifetime checks.
- Actual isolated Chrome/standalone host: `peer-controls-browser.log` and
  `tmp/peer-controls-browser-rJKFAa/result.json`, port52330, clean host exit0,
  zero model requests and no page errors. Actual controls created/deleted peers,
  refused duplicate/protected deletion, required confirmation, cancelled pending
  browser creation, kept socket selection live, completed creation only through
  an independent relay, preserved files and returned a deleted display to the
  mounted Agent. This run precedes only the current-ownership snapshot-race guard;
  final consumer above exercises the saved guarded implementation. No unchanged
  browser campaign repeated. Final wording-only page correction clarifies deletion
  recovery; it adds no behavior.
- JS syntax and `peer-controls-speech.log` passed. Shared214/0 and CLI evidence
  remain the earlier scoped results, not a new full regression. No core source
  changed here. No live provider, audible speech or human-accessibility claim.

Retained setup failures: route lambdas first selected ASP.NET's RequestDelegate
rather than result-producing handler overload (analyzer build failure); explicit
`Func<HttpContext, Task<IResult>>` resolves it. The external fixture incorrectly
called internal `DestroyAsync`; public `Destroy()` performs joined cleanup.
Its first HTTP run expected403 for a hostile Host, but the existing outer mount
correctly refuses400 before the inner route. No product defect credited from
those failures. Guarded edit-anchor refusals, including partial scripts, were
inspected and completed before verification; failed chains were not credited.

Next: reconcile remaining child management and graceful shutdown versus kill,
then chapter prose, scoped grading, final assembled review and source-only
preservation. Chapter21 remains in progress; ch20 is untouched. Reuse completed
selection/startup/control evidence instead of another unchanged campaign.

## Chapter 21 — per-connection GUI selection and independent browser startup (locally verified)

The GUI now projects existing ownership-qualified Agents per connection. Each viewed
Agent has one observer/partial buffer; each socket has its own selected Agent and
replay cursor. Replay is one nonblocking bounded-queue item, admitted before selection
changes; failed queue admission closes the socket rather than partially switching.
No new Agent registry, scheduler, transcript or accounting owner was added. Coalesced
`agent_tree` refresh queries actor status outside the socket lock. Controls use the
selected Agent; stale explicit targets refuse. Unnamed library roots retain null
global addresses. Only the borrowed initial Agent has a connection-local empty key;
other unnamed rows remain disabled and cannot steal its picker selection.

A socket's browser relay target is fixed at admission, independently of its displayed
conversation. `connection` announces that identity before MCP registration. Ordinary
switching preserves the existing McpBrowser/client and replies route to the originating
relay, not the current view. An explicit `?agent_id=<address>` page/socket URL opens
an independent relay; the UI offers that link rather than silently borrowing another
Agent's browser. Replay replaces cards silently, cancels previous-view narration and
rejects late old-view frames. GUI preferences remain host-scoped; displayed/runtime
limits and model controls follow selection. Borrowed hosting still owns neither the
Ensemble nor Agents and joins only its mounted requests/relays on teardown.

Standalone GUI serving and registry restoration now run together and join on either
failure or shutdown. The initial attempt demonstrated a real integration gap: MCP
claims deliberately fail when no browser is attached, so concurrent serving alone
could not initialize a browser peer. Agent.EnsureMcp now serializes lazy construction
with admission; existing skill/GUI paths use it. Mcp.WaitForBrowserAsync observes the
actual relationship, releases locks before awaiting, and wakes on attachment/removal/
destruction; waiter cancellation neither closes a browser nor changes other waits.
Peer creation/restoration opt into this wait only with `waitForBrowser:true`. GUI
restoration opts in; default/headless callers retain prompt missing-browser refusal,
not an unbounded wait in CLI startup. An awaiting peer is already owned/addressable,
but gains no browser tool authority or durable creation success until normal skill
startup and save complete. No source/default browser is inherited. Cancellation and
Ensemble destruction join removal of pending peers. Ordinary activation still refuses
missing browsers; the new wait is an explicit startup operation, not automatic retry.

Evidence in `/private/tmp/ch21-csharp.llwp67`: `gui-fleet-shared.log`214/0 covers the
shared MCP ownership/wait change. It predates ONLY the explicit headless-default guard
and unnamed-picker review correction. Final `browser-startup-tests-final.log`2/0,
`gui-fleet-peers-final.log`5/0, clean actual CLI/GUI builds and external consumer results
cover the guard (chain4850 exit0). `gui-selection-review-tests.log` and
`gui-selection-review-*-build.log` cover final GUI source (chain4861 exit0). The real
mounted consumer verifies independent tabs, repeated local child names, two unnamed
roots, stale-target/refused-selection behavior, selected controls, an outstanding MCP
call surviving a display switch, a separate peer relay, and joined caller-owned mount
teardown. Speech formatter tests and changed JavaScript syntax passed. No final full
suite after those two narrow corrections is claimed.

`selection-browser.json` / `gui-selection-review-browser.log` record owned muted Chrome
at `tmp/selection-browser-Hmhr8F`: real picker/replay/control isolation, explicit peer
relay and reload, no page errors, clean host exit. In-page transport fake separately
checks successful MCP reply identity and late-frame exclusion; that portion is not a
real provider call. The unnamed-picker assertion is also a focused in-page projection
check. `peer-browser-startup-guard.log` and `tmp/peer-browser-startup-nuk3W7/result.json`
record actual standalone hosts: selecting a waiting peer alone does not initialize
its browser skill; a dedicated tab does, reaching durable startup save. Waiting-startup
SIGINT exits0; malformed registry and headless CLI missing-browser launches exit1
promptly, with no surviving owned server. All four preserve registry bytes and make
zero inference requests. Earlier three-case startup results are superseded, not added
to a new campaign count. Final disk observation:12,670,488KiB free.

Retained failures/corrections: `peer-browser-startup.log` is the demonstrated initial
missing-browser failure. The first browser selection fixture waited for animation-frame
polling in a background tab; both pages were responsive and correctly selected, and
foregrounding the owned tab released the test. The corrected fixture explicitly
foregrounds its action page and bounds protocol waits. Unapplied guessed-host edit,
refused anchors, incorrect runner/scratch names and partially applied guarded scripts
were my setup errors; saved state was inspected and completed before testing. Review
also moved relay registration/initial replay into joined socket cleanup, preserved
headless refusal, and fixed unnamed-picker identity. No queue-byte cap was invented;
there is no claimed forced queue-saturation test. This is local integration, not live
providers/OAuth, audible speech or human accessibility acceptance.

Next: trusted GUI peer create/delete controls, remaining management tools and a truthful
graceful-shutdown/kill distinction, then chapter prose, scoped executable grading,
assembled review and preservation. A create request must not block the socket needed
for tree/relay traffic while startup waits. Current ShutdownAsync still cancels; do
not label it graceful or alias it to kill. Main is not a Chapter21 endpoint; ch20 and
older endpoints remain untouched. Reuse the verified selection/startup evidence.

## Chapter 21 — ownership addresses and fresh historical retrieval (locally verified)

`Agent.Address()` derives a canonical Ensemble-relative path from actual parents;
`Ensemble.FindAgentAddress()` walks existing root/child membership. Root names are
percent-escaped, child IDs retain their existing restricted syntax, and alternate
encodings/empty segments refuse. Unnamed transient roots and their descendants have
no global address: no fabricated root ID or duplicate registry. AgentTree values now
include the address alongside the actual Agent reference. These are trusted lookup
APIs, not broader model-tool authority or lifetime leases. Review corrected the new
getter to retain provenance during joined destruction while parent links remain.

`RetrieveSubmission` and shipped `get_submission` read owned artifacts freshly even
when a child is live. They never spawn/restore an Agent, activate skills, update live
completion or satisfy a join. The current child's actual home is used while attached;
after removal, lookup uses the caller's `children/<local-id>/submit.md`. Explicit
nonstandard library homes are not rediscovered through a new registry. One bounded
Sandbox read validates exact versioned fields, local identity, generation and the
existing schema/payload contract, returning independent typed JSON. Publication now
writes version 2 with nullable owning-parent address; named-parent provenance must
match. Existing version-1 rows remain readable from owned storage with explicitly
absent address provenance, never fabricated identity. The complete envelope has a
shared 2 MiB read/write cap; existing schema64KiB/payload1MiB/product limits remain.
The tool marks data `historical:true` and uses bounded errors without file contents
or host paths. Canonical addresses are scoped to this Ensemble, not cryptographic
cross-installation identity. Direct-child tools still refuse qualified/cross-tree IDs.

Evidence in `/private/tmp/ch21-csharp.llwp67`: `retrieval-tests-3.log`3/0 and
`retrieval-shared.log`212/0 (job4749 exit0) cover new work/reconstruction not reviving
saved completion, fresh disk replacement, exact large integers, legacy rows, malformed
fields/duplicates/schema, byte bounds, outward links and real shipped-skill dispatch.
The shared binary predates only the teardown address guard correction. Final
`address-final-tests.log`3/0, `retrieval-final-tests.log`3/0 and both actual-host
`retrieval-final-*-build.log` clean (job4753 exit0) cover final source. No later full
suite claimed. `retrieval-cli.json` records two actual CLI processes in owned
`tmp/retrieval-cli-bwokl78m`: first dispatch spawned/submitted/shut down a child with
one local model request; restart retrieved exact payload9007199254740993, no child
resurrection, unchanged artifact, zero further inference, both exit0. Shutdown may
cancel the child's post-submit continuation; this is durable recovery evidence, not
a claim that its final prose drained. No GUI/browser/live-provider claim added.

Retained setup failures: initial address edit missed the actual helper name (build
failed); retrieval fixture first treated shutdown's expected queued-task rejection
as success and omitted InvalidDataException, then consumed a stale pause observation.
Fixed explicit exception handling and an actor status barrier; no timeout swallowed.
Tool corruption diagnostics now explicitly handle InvalidDataException too. Source
review kept publication outside parent/child lock inversion and current completion
separate from historical bytes. Do not rerun completed peer/directory campaigns.

Next: per-connection GUI selection and independent browser relay/startup, then
remaining management (honest graceful shutdown versus kill), integrated GUI, prose,
scoped grade, review and preservation. Source comparison found v2 kept the initial
browser attachment when selecting another conversation; current C# has one manager
per Agent and one shared GUI projection. A naive detach/recreate on each display
switch would destroy active MCP clients. Preserve originating-agent reply routing
and separate display selection from relay lifetime. No GUI changes were saved in
this slice. Chapter21 remains in progress; main is not an endpoint; ch20 untouched.

## Chapter 21 — persistent sibling peers (locally verified; chapter still in progress)

Ensemble now declares `CreatePeerAsync`, `RestorePeersAsync` and `DeletePeerAsync`
in Common/Peers.cs, implemented through existing Agent/Skills/History/Sandbox owners.
The caller supplies a current top-level source Agent; no defaults-Agent relationship,
factory, copied services or second live registry is retained. `ensemble.json` is the
sole durable launch record, using v2's array fields `name`, `model`, `skill`, `created`.
Entries are read afresh under one operation gate; later running model changes cannot
rewrite earlier launch choices. Registry syntax/identity is validated before restore.
Creation restores/saves the peer's own History, independently initializes parsed skill
templates, then atomically replaces the registry before returning. Failure joins and
detaches constructed peers while preserving files. Ensemble destruction closes
admission, cancels and joins peer startup before destroying owned Agents/credentials.

V2 comparison resolved startup treatment: the initial host-configured actor is not
also a registry entry, though its physical home remains a sibling under `agents/`.
Both standalone hosts now restore registered peers; reusable borrowed GUI hosting is
unchanged. Peers complete ordinary chat without delegated submission correction and
do not reconstruct children. Explicit deletion refuses a retained child without a
current submission, closes admission and joins shutdown, removes the registry entry,
then destroys the subtree without deleting saved files. Cancellation before admission
refuses; after closing admission deletion joins to a definite result. Failed registry
replacement leaves the stopped peer and old entry for retry. This is NOT a new
claim that current cancelling ShutdownAsync is graceful or identical to kill.

Evidence in `/private/tmp/ch21-csharp.llwp67`: `peers-tests-4.log`4/0 and
`peers-shared.log`207/0 cover creation/restoration, separate launch/running models,
parsed-template/value independence, own History, no child resurrection, malformed
registry/partial-startup rollback, ordinary model chat across restart (two local
requests), HTTP MCP cancellation/retry and Ensemble destruction with active/queued
startup. Later deletion-only addition: `peers-delete-tests.log`5/0 and clean
`peers-final-cli-build.log`/`peers-final-gui-build.log`; no later full-suite claim.
`peer-hosts.json` records actual CLI/GUI HTTP launches, exit0, two restored peers,
unchanged registry and zero inference; it predates only deletion. Saved peer source
and declarations reviewed; whitespace checked. Test setup mistakes retained: first
refusal helper omitted InvalidDataException; chat fixture guessed a SubmitAsync
cancellation overload (build-3 failed before behavior). Neither was a product defect.

Next: qualified ownership addresses and per-connection browser selection/relay,
then durable retrieval, remaining management, integrated GUI/prose/grading/review and
preservation. Browser-dependent peer startup is an explicit gap: current single-Agent
GUI cannot supply an independent peer relay; do not silently borrow the initial
actor's browser or call HTTP restoration evidence browser-switching evidence. Existing
current peers use ordinary/HTTP skills. Main is not a Chapter21 endpoint; ch20 is
untouched. No workers, deployment/restart, credentials, live OAuth or Bill browser used.

## Chapter 21 — explicit Ensemble and actor homes (locally verified foundation)

Bill clarified that GUI-created top-level actors remain Ensemble-owned and that
all such actors, including the initial one, have parallel homes beneath the shared
container. This clarifies v2's `agents/<name>/` convention without making the first
actor a parent of its peers. No running data was migrated. The rewrite must explain
both this relationship and why an actor home is distinct from a History save path.

`Common/Storage.cs` now declares the homes and access operations. Constructor wiring
uses existing Ensemble/Agent ownership and Sandbox, not a default-Agent relationship,
manager or second live registry. Configured Ensemble assigns named homes; conflicting
explicit homes refuse before registration. Unnamed library Agents remain supported,
with an explicit home when child storage is needed. Children derive their homes from
the actual owning Agent. `StartPersistence` no longer implicitly supplies a data home.
CLI/GUI add `--data-dir`/`EN_DATA_DIR` and `--agent-name`/`EN_AGENT_NAME` (default `root`).
Default logs/save/Jobs follow actor home; GUI settings follow the Ensemble container.
Explicit save paths remain independent. Main README reflects these contracts; ch20
and earlier endpoints are untouched.

Evidence in `/private/tmp/ch21-csharp.llwp67`: storage2/0, creation3/0,
submission6/0; `storage-shared.log`203/0. Final review added refusal of conflicting
named/explicit homes: `storage-tests-final.log`2/0 and both actual host builds clean
(`storage-*-build-final.log`); the shared run predates only that narrow guard.
`storage-hosts.json` records two actual CLI launches and one actual GUI HTTP launch,
all exit0, sibling directories, save override, flag-over-env precedence, preserved
peer files, no writes in the launch directory and zero local inference requests.
These launches predate the guard but use unchanged, nonconflicting inputs. No new
browser/accessibility/live-provider claim. The initial wrong CheckCase helper and
missing fixture Sandbox roots were setup failures; logs retained. A preflight edit
anchor refusal saved no host changes. Next: peer registry/create/restore and stable
addresses, then remaining Chapter 21 work. No Chapter 21 grade or endpoint yet.

## Chapter 21 clarification from Bill; Chapter 20 effort/grader comparison

Bill clarified that a GUI-created top-level Agent remains an Ensemble Agent and its home belongs under the Ensemble top-level data directory. He explicitly permits consulting v2 for clarification and requires carrying the finding into the chapter rewrite. Source inspected: solutions/ch21/ensemble/registry.go, solutions/ch26/ensemble/registry.go, agent/ensemble/registry.go, agent/ensemble/spec.go, agent/ensemble/ensemble.go, and agent/gui/gui.go. V2 Start retains the Ensemble directory, stores ensemble.json there, and creates/restores named peers with DataDir agents/<name>, SavePath agents/<name>/save.json and LogPath agents/<name>/events.jsonl; relative paths are resolved against the Ensemble directory. GUI create_agent calls that same Ensemble.CreateAgent, not an independent application. Agent data home and save-file location are separate facts in v2. No running configuration or other actor's private files were read. I incorrectly elevated missing C# host wiring into a decision about a retained default-Agent relationship. That proposal is NOT implemented or approved and is not needed merely to locate storage. The prior awaiting-approval boundary is superseded for this storage question. C# currently derives DataDirectory from persistence.Path; correct the data-home vocabulary and host wiring against the required Ensemble-owned layout rather than infer a primary Agent by name or create a manager. Persisted launch choices still must remain distinct from mutable runtime model. Add these distinctions explicitly during Chapter 21 prose reconciliation.

Bill also asked whether Common provides easy, correct data access. The inspected Chapter 21 paths use declared Common operations and actual parent references: Job -> Jobs -> Agent; direct-child FindChild; child -> Engine.Meter; Ensemble.Agents/Agent.Children for AgentTree. Child Add/Remove maintain both links; activity derives from existing actor state, retaining only its last emitted projection. No second meter, live registry or service bag was found in those bounded paths. This is not a blanket certification of all source. The storage-home conflation above remains an actual modeling gap; a default-Agent pointer would not fix it.

Bill requested an effort/complexity/grader comparison including Editions 1 and 2. Read internal/grade/ch24_checks.go and ch25_checks.go, internal/gradeedition2/ch20.go and ch20_executable.go, current book/chapter-20.md and preserved pre-reset chapter schema text, plus Chapter 20's actual grading checkpoint. Edition 1 had separate 100-point GUI/library and 100-point fleet exercises. Edition 2 Chapter 20 combines those same two tables and strengthens actual consumer/nested workflow assertions. C# executable scope reuses the seven existing fleet rows (100); Go-specific GUI/import points are excluded and C# reuse is separately exercised. Sender attribution, message wake with surviving work and three bounded corrections were already Edition 1 fleet requirements, not newly invented hurdles. The supported typed schema subset is in existing chapter text/preserved revision, not a new C# grader invention. No new historical timing measurement or fresh grader run was made.

The chapter legitimately combines GUI library extraction with first child ownership/workflow integration. Shared groups grew from Chapter19's150 to191, not191 newly added tests. However much delay was self-inflicted: guessed APIs/project paths/selectors; wrong schema/skill/cursor assumptions in the new adapter; lost first-run wire evidence through a deleting helper; repeated incorrect point totals; oversize/clipped reads and repeated bookkeeping. The recorded 55/70/85/100 executable progression is harness correction, not four rounds of product defects. A genuine missing-sender product gap was found and fixed separately; shutdown of a pending human question and earlier generation/schema-projection defects also justified targeted tests. Repeated full regressions after small slices and multiple final-review/checkpoint passes added overhead; shared changes justify relevant regression but not a new campaign for each milestone. No defensible wall-clock percentage can be inferred from these notes. Preserve required architecture/ownership review, real boundary fakes and external-consumer exercise; reduce speculative harness assumptions, guessed source references and repeated process. This comparison does not revoke the preserved ch20 evidence or authorize lowering required behavior.

## 2026-10-28 — Chapter 21 child interrupt locally verified; peer ownership proposal awaits agreement

I resumed the named-peer requirements and existing Ensemble/CreateChild/SkillBootstrap paths. Ensemble currently permits multiple unnamed standalone Agents and has no designated defaults Agent or peer-storage home. GUI Preferences is not a suitable owner for shared peer configuration. I did not add a second live registry, infer a primary Agent from its name, retain copied services, or implement a peer manager. Proposed shared-vocabulary change, NOT implemented: an optional explicit default-Agent relationship on Ensemble and explicit peer-storage location, supplied by the composition root. Actual Agents remain in the existing collection; persisted peer launch model/skill choices must remain distinct from mutable current model selection. Both relationship ends and detachment must be declared; borrowed GUI hosting must not acquire ownership of its caller's Agent. This proposal awaits Bill's architectural agreement before implementation. Stable browser addresses and durable retrieval remain dependent work; no invented parent IDs or artifact changes landed.

While keeping that decision open, I implemented the independent `interrupt_agent` tool. Common/Tools.cs appends InterruptAgent without renumbering existing values; Tools/Tool.cs declares agent_id and optional enum mode now/after_tool and routes to existing Tools/Children.cs. The handler resolves only owner.FindChild and posts the existing InputKind through the child's reliable mailbox. It reports agent_id/mode/interrupt_delivered, explicitly delivery rather than a quiescence barrier. No scheduler, task tracker, actor behavior, lifetime control, accounting or new state was added. Observer cancellation cannot retract an admitted mailbox control. The shipped ensemble skill grants and explains the operation; catalog tests include it. This is not kill or shutdown. The chapter distinguishes graceful shutdown from forceful kill, whereas current ShutdownAsync already cancels: those semantics still need deliberate reconciliation, not two names for one operation.

Tests/ChildInterruptTests.cs drives actual dispatch, real owned Agents/Jobs and a bounded loopback provider. Three groups cover shipped-manual provider reachability; unknown/peer/foreign/nested refusal with the same response; invalid modes; omitted now; idle and stopped children; immediate interruption with an independent held Job surviving; after_tool remaining pending until the real held report; undispatched tail error correlation; unchanged child identity; and a subsequent delegated turn submitting durably and completing on the same Agent. All fixture jobs/server tasks are joined. Focused `interrupt-build-2.log` is clean and `interrupt-tests-2.log` is 3/0. Full `interrupt-shared.log` is 201/0; actual CLI/Ch04.csproj and Examples/Gui/Gui.csproj builds in interrupt-cli-build.log and interrupt-gui-build.log are clean; the chain exited zero. No new executable grade, GUI campaign or live provider evidence is claimed.

Retained fixture failures: initial interrupt-build.log was clean but interrupt-tests.log had 1/2 because I constructed a delegated child without StartPersistence and then asked it to submit. interrupt-diagnostic-tests.log retained the exact correlated refusal, “Configure Agent persistence before using child storage.” Both interruptions had worked; the later completion correction outlasted the incomplete three-response fake. I configured an owned child save path, without changing production behavior, and the corrected groups passed. The new test retains narrow failure diagnostics to avoid reducing future continuation failures to unexplained deadlines. Several combined source reads were clipped; I reread the new test and handler/catalog bodies in bounded sections and did not count omitted older material as review. Scoped whitespace check passed. All asynchronous handles (4509,4514 failed fixtures;4517 passing full chain) are collected; no owned live fixture remains. Last free-space observation was 13,327,632 KiB, not a future guarantee.

Chapter 21 remains in progress, with no endpoint or executable grade. Main has activity, actual-identity tree snapshots, direct-child telemetry and interrupt; ch20 remains untouched/preserved. Next is the explicit named-peer ownership/storage decision above, then named peers/addressing, durable retrieval, remaining management and GUI switching, prose/grading/final review/preservation. No workers, commits, installs, restarts, private credentials, Bill browser/8088, or live OAuth.

## 2026-10-28 — Chapter 21 ownership snapshot and detailed child status locally verified

I read the existing Ensemble observer implementation and the grader's tree/status helpers. The grader requires distinct ownership-qualified browser IDs and actual nested membership, not local-ID deduplication. The library permits multiple unnamed standalone Agents, so I did not invent global string IDs or a registry. Common/AgentTree.cs declares a value-like AgentTreeNode containing actual Agent identity and immutable captured children; Application/AgentTree.cs walks Agents()/Children() with each owner lock released before descent. Immutable ChildID distinguishes standalone roots through teardown, avoiding promotion of a detaching child. Snapshot references do not own lifetime or grant control. The snapshot is deliberately not globally atomic; later additions appear on refresh and captured Agents can detach. Browser address mapping and selection still need implementation.

The new grant-controlled get_sub_agent_status resolves only a direct child and reads existing actor WorkStatus and one immutable Engine Meter snapshot. Common status fields share one Tools projection with agents_status. Detailed JSON separates selected model from measured model/tier/billing rows, includes nullable latest/session counts and cache-hit rate, and labels cost as estimated. There is no copied meter or new scheduler. Unknown, foreign, peer and nested/non-direct IDs receive the same ordinary tool refusal. The shipped ensemble skill grants and explains the tool; its actual initialized History context is tested for provider visibility.

Scratch /private/tmp/ch21-csharp.llwp67: tree-build.log clean, tree-tests.log **2/0**, status-build-3.log clean, status-tests-3.log **1/0**, tree-status-shared.log **198/0**, tree-status-cli-build.log and tree-status-gui-build.log actual hosts clean, chain exit zero. Tree tests cover repeated IDs in different branches, multiple unnamed roots, shutdown versus destruction, retained snapshot immutability, and concurrent creation/destruction without false roots. Status uses real loopback Engine sends held at HTTP receipt, direct-child dispatch, missing usage, unknown pricing, isolated meters, stopped retained state, and removal. Earlier status-tests.log passed with installed catalog only. status-tests-2.log failed because my fixture rendered an empty Context, omitting the activated skill history; the grant itself was enabled. Correction uses History().View(), as existing skill tests do; no product change was justified by that fixture failure. Repeated guessed inspection paths and overfilled reads were my process mistakes, not compilation or product failures. All exercised behavior remains local, not live-provider or GUI-supervision acceptance.

Durable retrieval reconciliation remains an implementation decision for the next storage slice: current text says restore a child from disk; preserved revision lines332–394 explicitly says freshly read its authoritative owned file without respawning or changing lifecycle. Current C# writer already stores version1, child ID, generation, schema and typed payload, but not parent address. Do not import the preserved Go record shape or invent parent IDs before settling named-peer/browser addressing. Preserve live schema enforcement and the distinction between retained disk result and current-generation submission. Existing GUI/host teardown and activity evidence remain unchanged. Next: named-peer ownership/configuration and stable presentation addressing, then durable retrieval and remaining management/GUI work. No Chapter21 endpoint or grade; ch20 remains untouched.

## 2026-10-28 — Chapter 21 underway; activity foundation locally verified

I read the complete current Chapter 21 manuscript, its specification/stages and exercise, compared the beginning of the preserved revision, and located the current grader routes/functions. This is not a completed review of every Chapter 21 grader assertion or preserved revision. The C# adaptation continues from ch20, not the abandoned Chapter 21 campaigns. The chapter's historical Go factory/manager instructions do not authorize recreating that machinery around the existing C# owners. Durable retrieval wording differs between current and preserved manuscripts and still requires reconciliation before that slice.

The first implementation slice adds value-like `WorkPhase`/`WorkActivity` in Common and `ActivityChanged` on the existing lossy Observation stream. Agent derives current phase from actor-owned transport, tool, automatic, startup and control state; it does not add a scheduler or write activity to History. `AgentActivity` retains its Chapter 20 lifecycle meaning. `WorkStatusAsync` includes current work; `agents_status` adds `work_phase`, `tool` and `detail` without renaming lifecycle/submission fields. Foreground work takes presentation priority over concurrent controls; one retained control is representative, not a claim of globally single work. Control replies retain their actual call instead of adding another call registry. Short detail labels contain no copied command/question arguments. The only retained phase value suppresses duplicate progress publication and supplies closed-status readers without racing actor fields; closing admission does not itself declare drained work. Joined cleanup publishes idle.

`Tests/ActivityTests.cs` has four focused groups: real loopback two-request model/tool lifecycle, control-only wait over an independent Job (including actual status-tool JSON), interrupt/failing-tool cleanup, and malformed/shutdown transport cleanup. Gates and actor status queries establish lifecycle boundaries. The tests check thinking/tool_call/waiting/idle, duplicate suppression, unchanged task output, and no activity text in provider context or canonical History. `activity-tests-build-4.log` is clean and `activity-tests-4.log` is **4 passed, zero failures**. `activity-shared.log` is **195 passed, zero failures**; `activity-cli-build.log` and `activity-gui-build.log` are clean builds of the actual hosts. This is library/fake-boundary evidence, not a GUI supervision exercise or Chapter 21 acceptance.

Retained setup failures: two test calls guessed `Job.Info` instead of `Snapshot`; my initial correction guard miscounted those calls and refused before writing. A later test guessed `EventKind.Error` instead of `ErrorOccurred` and used generic-Task `SuppressThrowing`; compilation rejected it, and the corrected test no longer swallows deadline cancellation. Incorrect inspection paths stopped before builds, not product failures. I also overfilled several combined reads; only recovered bounded spans count as reviewed. Production delta and new activity implementation were read after saving. No product failing baseline was claimed from these fixture errors.

Main is now Chapter 21 **in progress**; `../ch20` remains the preserved 239-file endpoint and was not edited. No Chapter 21 executable grade, endpoint, named-peer startup, durable retrieval, tree/telemetry, full management tools, GUI switching or Chapter 21 prose reconciliation is complete. Next: inspect tree/telemetry grader contracts and complete existing Ensemble observer/ownership paths, then implement a snapshot over real root/child relationships without a parallel manager or registry. Remaining chapter work follows the current requirements, with explicit reconciliation where the manuscript revisions differ.

Scratch: `/private/tmp/ch21-csharp.llwp67`, with its owned `offline` wrapper/home/tmp. Last free-space observation: 13,766,876 KiB. Continue personally/no workers. No installations, restarts, commits, private credentials, Bill browser/8088 or live OAuth. Live providers, Linux/x64, audible and human accessibility remain unverified. Earlier Chapter 20 workflows/grade were not rerun or relabelled.

## 2026-10-28 — Chapter 20 locally complete; source-only endpoint preserved

I completed the assembled review across the implementation, public consumer, tests, shipped manual, grader and C# manuscript. This final pass read the external GUI consumer, child ownership/creation/skill/submission/completion/question/join/tool/spawn/status/wake/chat tests, schema limits, test registration, and the executable adapter and scope tests. Bounded rereads recovered clipped actor, Job, question and tool-handler spans. The model remains direct Agent ownership with both parents, explicit joined destruction, one actor scheduler, immutable submission/reply values and ordinary Jobs for observation; optional Gui borrows caller lifetimes. Passing grades were not substituted for that source review. No production behavior changed in this final pass.

I reconciled the main/library/example READMEs: removed the stale pre-tool milestone, documented current shipped grants and replaced unfinished boundaries. An exact-text edit refusal occurred before any file was saved or endpoint created; I read the actual sentence and corrected the anchor. I also corrected a continuity-note mistake: the manuscript and implementation already agree on 64 KiB schema, 1 MiB payload and `(schemaBytes + 1) * (payloadBytes + 1) <= 67,108,864`, with depth 64/root zero, 2,048 schema nodes, numeric literal 128 characters and exponent magnitude 1,000. The preceding handoff's 4,096/64-KiB/128-KiB summary was wrong; no validator or manuscript change was needed for that discrepancy.

`solutions/edition-3/ch20` was created only after confirming it was absent. Its **239 source/support files are byte-identical to reviewed main at preservation**; all 13 project references resolve within the endpoint. No symlinks, bin/obj, runtime data or credentials were copied. The retained `scratch/preservation.json` contains the exact hashes and reference result. Earlier endpoints, including ch19, remain untouched. The final scoped whitespace check was clean. Last free-space observation was 13,841,824 KiB.

Unchanged executable evidence remains valid: `reply-sender-shared.log` is **191 passed, zero failures**; focused tools/questions and actual CLI/GUI host builds are clean. `grade-fourth.json` is **100/100, passed:true, exit 0**, with five clean sessions and 59 local provider requests. Applicable grader contracts and five prelaunch invalid-scope refusals passed. The separate historical 100-point Go GUI/reuse exercise is excluded; the original Go route remains intact. Earlier failed product baselines, wrong-selector run and harness-only 55/70/85 grades remain recorded in the prior entry and `scratch/grader-checkpoint.md`; this boundary does not erase them. No unchanged verification campaign was repeated.

The exact new manuscript consumer builds cleanly (`prose-consumer-build-final.log`), serves the caller-owned `/health` and mounted `/assistant`, and exits zero (`prose-consumer/result-final.json`) without inference. Previously recorded external GUI, isolated mounted-browser and actual nested JSON/chat CLI exercises retain their original scope. Shell snippets were syntax checked, not all live-executed. Local peers do not establish unscripted model judgment, live provider/OAuth/entitlement/billing, hosted web, native Linux/x64, audible quality or human accessibility. Live OAuth remains deferred to Bill's joint session. No workers, installs, restarts, commits, private credential access or Bill browser/8088 use occurred.

Scratch: `/private/tmp/ch20-csharp.hH3ysE`. **Next: Chapter 21 requirements, current implementation and grader review, personally, after the chapter-boundary handoff.** Preserve ch20 and earlier endpoints; do not revive abandoned Chapter 21 campaigns or transfer their results into this C# continuation.

## 2026-10-28 — Chapter 20 assembled review underway; C# manuscript reconciled

Chapter 20 remains **in progress**, with ch19 still the preserved endpoint. I reviewed the ownership/creation/submission/question/join/status paths, actor and shutdown deltas, Job observation pulse, skill inheritance, Sandbox replacement, GUI extraction/mounting and CLI input changes. The 88-path delta map is still a review map, not a completion receipt: final test/consumer and grader-source review, recovery of any clipped source spans and source-only preservation remain. No production behavior changed in this pass; the earlier full 191/0, actual-host builds and scoped executable 100/100 remain the applicable evidence and were not repeated.

`book/edition-3/chapter-20.md` now teaches the actual C# assembly boundary and Common/functional partial organization, direct two-parent Agent ownership without a manager/factory, independent creation and task admission, current typed submission artifact and generations, sender-bearing questions, actor status, non-sticky Job observation wake and live CLI human reply input. It distinguishes submission from quiescence and shutdown. The schema subset/limits match the actual validator, including exact numbers, Unicode lengths and .NET regex timeout rather than Go compatibility. The exercise uses actual projects, lowercase `ch04.dll`, shipped ensemble/code-tools grants and the scoped 100-point route; the Go GUI/reuse 100 points remain excluded. Historical wiring failures, sweep and second-edition runs stay explicitly historical. `Gui/README.md` no longer incorrectly calls already verified workflow/grading work unfinished; it points to this record for the current boundary.

A new public mounted-GUI example was extracted directly from the manuscript into scratch/prose-consumer, compiled and exercised from that owned directory without credentials or inference. The final example observes hosting failure while waiting for Enter, cancels/joins the view before caller host and Ensemble teardown, and serves `/assistant` beside `/health`. `prose-consumer-build-final.log` is clean, and `prose-consumer/result-final.json` records HTTP 200, caller-owned health text and exit 0. Earlier example build/smoke logs remain. Its shell blocks passed `sh -n`; the provider-backed fleet instructions were not run live. This is HTTP/lifetime documentation evidence, not browser rendering, OAuth or accessibility. Existing mounted-browser and nested CLI campaigns were not repeated.

Editorial errors were caught in the assembled read: the first rewrite retained a contradictory old unnumbered specification; I removed it and retained the opening rationale. I initially guessed schema limits in draft prose, then read the actual admission/numeric loops and corrected them to depth64 (root0), number literal128 and exponent1000. I also corrected payload-in-status wording (status exposes completion state, not payload), obsolete TL;DR reference, and an example that did not initially monitor late hosting failure. `scratch/rewrite-ch20.py` is the historical staging script, not a generator to rerun over the corrected chapter. The final example changed after its first pass, so its build and small HTTP/exit exercise were rerun; no unrelated suite was rerun.

Next: finish final bounded test/consumer and grader-source review, reconcile any remaining prose/README details, preserve a source-only ch20 with byte comparison and project-reference checks, then perform the chapter-boundary handoff and continue Chapter21 personally. No workers, installation, restart, commit, private credentials or live OAuth. Last observed free space: 13,846,028 KiB, not a future guarantee.

## 2026-10-28 — Chapter 20 scoped executable verification; sender attribution corrected

The Chapter 20 executable scope now passes **100/100, passed:true, exit 0** in `grade-fourth.json`, with empty `grade-fourth.stderr`, under `/private/tmp/ch20-csharp.hH3ysE`. `internal/gradeedition2/ch20_executable.go` runs all seven existing fleet checks using the real C# JSON CLI and shipped ensemble/code-tools skills. Typed schema validation, invalid submissions, durable structured payloads and nested child persistence are assertions within those checks, not invented extra points. The original 100-point Go GUI/reuse exercise is excluded; separately verified C# GUI consumers do not earn its score. The original Go Chapter 20 route remains unchanged behind the normal CLI branch. New `-ch 20 -executable-only -fleet-skills ABSOLUTE_DIRECTORY CLI` explicitly selects this scope. Five invalid-scope combinations refuse before executable resolution (`grade-scope-refusals.json`).

Grader reconciliation found one product gap: question replies had response text but no sender, despite the chapter's explicit attribution requirement. `Common/Questions.cs` now declares the immutable value-like `QuestionReply(Sender, Response)`. The existing question edge captures the actual recipient's child ID/name (unnamed root role `root`) before detaching links. Human replies default to `user`; the existing host boundary can supply another nonblank human address, never overriding an Agent sender. `SendMessageToParentAsync` returns this typed value, and the ordinary tool emits `sender` plus `response`. No second routing owner or mutable address store was added. Focused question and child-tool tests check nested and named-human senders; cancellation/ownership behavior remains unchanged.

Sender evidence: the first `reply-sender-baseline.log` was a wrong-selector default subset (122/0), not the intended baseline. Correct `reply-sender-baseline-focused.log` reproduces 3/1. `reply-sender-tools.log` passes 4/0, `reply-sender-questions.log` 3/0, and `reply-sender-shared.log` 191/0. `reply-sender-build.log`, `reply-sender-cli-build.log`, and `reply-sender-gui-build.log` are clean. These are the final production-source results for this slice. The API edit initially stopped on an unmatched tool-handler anchor after saving its earlier files; I inspected and completed the actual handler before building. This partial editing error is not a successful test or product defect.

The passing executable run retains five owned session trees under scratch/tmp: `c20-executable-3293158692` (structured workflow, 21 requests), `c20-executable-758881050` (nested/human questions, 16), `c20-executable-737383313` (message wake and surviving shell, 12), `c20-executable-2704650426` (bounded idle correction, 8), and `c20-executable-2473266054` (ordinary root chat, 2). Total 59 local HTTP requests; all five EOF shutdowns joined and exited cleanly. Each tree retains provider-requests.json, host.stderr, actual state and Job output. The wake uses a real child Job waiting for the parent shell's start marker, then an incoming question; the unrelated shell yields still-running before release and later delivers output plus terminal exit 0 across its cursor observations. Nested replies attribute mid/root/user. Four plain-text child responses exhaust exactly three corrections; root chat still produces one request per ordinary prompt.

Harness mistakes and failed evidence remain explicit. I twice misstated the scope (140, then 155) before reading the actual seven-row table closely; there is no additional 75-point submission table. The correct score is 100. Initial adapter compilation used guessed nonexistent helper names; those were corrected to actual package helpers. A broad `^TestChapter20` selector compiled but failed two existing tests requiring absent historical Go solution paths; those tests are uncredited, not product failures. Formatter-path and top-level artifact-copy errors stopped setup and were corrected. `grade-first.json` is 55/100 but its temporary wire trees were deleted by a reused historical helper: only that grade survives, and no detailed wire conclusions are claimed. The adapter now retains its own trees. `grade-second.json` is 70/100; `grade-third.json` 85/100, both with retained evidence. Corrections were harness-only: remove exactly the three top-level observation fields before schema equality; activate shipped code-tools rather than assume ensemble grants shell/file tools; retry only the exact undispatched-hint refusal; wait for actual pending-question output; keep the fake provider responsive by using a real Job for the start gate; and accumulate output across cursors while separately requiring terminal done/exit zero. No production change was made to satisfy those harness mistakes.

The grade is not an architecture verdict: the assembled source/manuscript review and source-only preservation are still unfinished. Applicable pure grader contracts plus new latest-observation/scope/exact-retry tests pass in `grader-contract-tests-final.log`; the CLI executable was pinned from the already verified sender-corrected build into scratch/grade-hosts. No unchanged GUI/browser/nested campaign was rerun. No Chapter 20 endpoint exists yet; ch19 remains the preserved endpoint. Local fake-provider evidence establishes only the exercised contracts, not live providers/OAuth/billing, native speech or human accessibility. No workers, installation, restart, commits, private-account searches, Bill browser or port 8088 use. Next: final assembled source and prose review, preserving Bill's reasons and historical attribution, then source-only endpoint and chapter-boundary handoff.

## 2026-10-28 — Chapter 20 shipped skill, live CLI input and nested workflow verified locally

I added the seven ordinary child-workflow grants and teaching guidance to the shipped ensemble skill. A real CLI process now exercises that shipped skill, rather than relying only on a custom test catalog. Chat stdin no longer waits for each model answer. `InputKind.Chat` is decided by the actor: an idle input retains the existing reliable turn reply; an active input records/wakes as a hint and acknowledges independently. `PostAsync` exposes the string task it already owned (other controls still return empty acknowledgments). The host has only an output queue, not a second activity/scheduling owner. JSON prompt/hint behavior is unchanged.

Human-question display waits on the actual inbox, separately from lossy progress. `WaitForQuestionsAsync` optionally copies excluded display IDs at entry, permitting unseen-question waits without consuming or polling pending questions. Chat supports `/reply TEXT` and `/reply MESSAGE_ID TEXT`; JSON accepts `kind: reply`, `text`, and optional `message_id`, and emits `human_question` objects. Reply selection defaults to the root-self human inbox, never child-parent questions. Refused replies leave stdin and questions available. EOF retains drain semantics; `/quit` explicitly cancels and joins. The README documents these distinctions.

Two focused groups verify actor-decided reliable chat/hints, surviving held Jobs, a later idle turn, copied exclusion filters and human/child inbox isolation. `chat-tests-build-2.log` and `chat-tests.log` pass 2/0. The first post-edit build (`chat-tests-build.log`, `chat-cli-build.log`) was clean; no behavioral failure was hidden there. The pre-shutdown-correction shared run (`chat-shared.log`) was 191/0 with both hosts clean; that result predates the cancellation correction below.

Actual CLI evidence is `/private/tmp/ch20-cli-nested.grwtrovq/result.json`: JSON/chat each exit 0, respectively 19/20 owned loopback model requests, seven shipped tools advertised, root human question/reply, wrong-ID refusal without loss, child/grandchild parent questions, reuse of the running join handles, bottom-up durable submissions and targeted joined subtree shutdown followed by empty status. Chat additionally accepts a live ordinary hint while waiting for the human. Both `submit.md` payloads and child primary saves remain after destruction. Requests/stdout/stderr and peer errors are retained beside the result. This is real host/local-peer integration, not live-provider behavior or a model's unscripted judgment. The positive nested runs predate only the narrow independent-cancellation error conversion and chat diagnostic formatting; their successful paths are unchanged, so I did not repeat them.

An actual `/quit` with a pending human question exposed a genuine defect, retained in `cli-quit.log` and `/private/tmp/ch20-cli-nested.14i3fi_2`: question graph teardown cancels the answer before the observing Job token is cancelled. That expected cancellation escaped the tool boundary, faulted the Job task and reappeared as aggregate shutdown failure (process exit -6). I added pending-root-question shutdown to the existing child-tool group; `question-shutdown-baseline.log` reproduces 3/1 after its clean build. `Tools/Children.cs` now converts independently cancelled question/spawn completion to ordinary IOException tool feedback only while the Job token remains live. Actual Job cancellation follows its unchanged path; other cleanup/recording failures remain visible. `question-shutdown-tests.log` passes 4/0 and the CLI rebuild is clean. `/private/tmp/ch20-cli-nested.vtzx00x5/result.json` proves corrected `/quit`, one local request, exit 0 and plain chat diagnostics. Final `chat-final-shared.log` is 191/0; `chat-final-gui-build.log` is clean, and the corrected CLI artifact is `question-shutdown-cli-build.log` (clean).

Setup mistakes remain distinct: `cli-nested-1.log` launched unsupported command `json`; the CLI uses default JSON stdin mode. I corrected the fixture and did not change the product for that refusal. An unnecessary broad scratch grep blocked on a runtime FIFO; I identified my shell 83129/grep 83134, terminated only that grep and collected command 4217. No shared cleanup occurred. Subsequent reads name files explicitly. An earlier wake fixture compile failure remains in the previous entry. Last free-space observation was 12,808,284 KiB, not a future guarantee.

I reviewed the changed actor chat branch, Common signatures, question filter, CLI input/output/cleanup paths, shipped manual and narrow tool cancellation catch. No workers, installation, restart, commit, private credential use or Bill browser/8088 access occurred. Chapter 20 still has no grade or endpoint. Next: inspect scoped executable grader requirements, adapt/run appropriate C# checks, then assembled chapter prose/source review and source-only preservation. Carry prior GUI evidence unchanged unless those boundaries change.

## 2026-10-28 — Chapter 20 incoming-message observation wake verified

I added a non-sticky observation pulse to the existing Job owner, not an execution cancellation or a second scheduler. A serialized observer captures the current pulse once; waking swaps/completes that pulse. Observers which begin afterward receive a fresh incomplete pulse, including after a message with no current waiter. Jobs snapshots owned work and pulses outside its membership lock. Human Hint/Message records canonically and wakes current observations, including control waits; real interrupt/shutdown cancellation is unchanged. Work keeps its handle, cursor and eventual result.

Child questions establish both graph links and snapshot under the question gate, release it, then notify the recipient actor through typed mail. The actor records labeled `HintReceived` with `Actor.Tool`, never Human/System authority, and wakes current Jobs. Active work sees it on its next request without advancing accepted-work generations; an idle recipient records it without admitting inference. Root-self human questions stay in the human inbox and never notify their own model. Shutdown continues to own graph cleanup; notice admission takes no ingress lock while holding the graph lock.

Four focused groups cover nonmatching 120-second observations, no-current/no-next-wait stickiness, cursor retention, human control waits, unrelated work survival, idle-parent/no-inference behavior, pending question/reply ownership, root-human isolation, and a model-led foreground wake with question data outside system instructions. Explicit actor status barriers follow synchronous observation admission; no sleeps establish waiter readiness. My first fixture build guessed `jobsDirectory` and omitted the child's required parents/configuration, so it failed before behavioral tests; corrected calls use actual declarations. Retained `message-tests-build.log` is setup failure, not product evidence. Final `message-tests-build-2.log` and `message-tests-2.log` pass four groups. `message-shared.log` passes 189/0, including existing live human hint paths. `message-cli-build.log` and `message-gui-build.log` build actual hosts cleanly. Chain 4200 exited 0. Scratch remains `/private/tmp/ch20-csharp.hH3ysE`; free space observed 12,862,948 KiB.

Remaining: shipped ensemble skill grants/manual, CLI human question display and `/reply`/JSON input while work runs (chat currently blocks its reader awaiting each answer), actual nested model-led CLI workflow, scoped grader/prose/final review and preservation. No Chapter 20 grade or endpoint; Chapter 19 remains preserved. Earlier GUI campaigns unchanged. No workers, shared changes, restart, installation, commit, private credentials, Bill browser/8088 or live OAuth. Offline behavior is not live-provider or accessibility evidence.

## 2026-10-28 — Chapter 20 spawn task ownership and actor status verified

I added `spawn_sub_agent` and `agents_status` as ordinary grant-controlled tools, preserving existing enum identities. Spawn validates its task and delegates schema/model/storage/skill preparation to `CreateChildAsync`, then admits work with `SubmitAsync`. Its existing Job retains that reliable task reply and urgently reports the child ID through the existing unread-attention output boundary. No detached runner, task registry, child manager or second scheduler was added. Once creation returns, the parent owns the child independently of the spawn observer: cancelling/killing the observing Job does not cancel or destroy the admitted child. Explicit `shutdown_agent` remains the lifetime operation. The final spawn report separates task completion from the current typed submission; retained completion failures are errors, not successful idle outcomes.

`Common/ChildStatus.cs` declares value-like `AgentWorkStatus`/`AgentActivity` and Agent operations; `Agent/ChildStatus.cs` implements them. A typed mailbox query reads active/queued work, foreground transport/tool work, startup/control replies and pause state on the actor. The existing ingress-protected direct-Engine-send count is included without copying scheduling state. Submission and retained failure are independent fields. Closed lifetimes use synchronized admission/actor-task state, reporting stopping/stopped without resurrecting the actor; pending typed replies are completed through normal handling and both fatal/final drains. Caller cancellation detaches only the observer. The status tool captures direct-child membership and queries each child independently; it is not a global-time snapshot or a completion barrier.

Four focused real-component/loopback groups in `Tests/ChildSpawnTests.cs` use actual model requests, real tools, durable files and explicit synchronization gates. They cover compatible model override and typed submit schema on the wire; early spawn report despite long delay/nonmatching pattern; submission while the task is still draining; final task reply and typed payload; owned durable conversation journal; live/idle/stopped status and removal; invalid task/schema/ID and duplicate-child refusal; spawn-observer cancellation with child survival; status cancellation and shutdown query completion; paused queued work; exactly three correction reminders followed by explicit failure; and direct Engine sends outside the actor. `--child-spawn-tests` exits after these four groups. The expected built-in catalog now includes both tools.

Evidence: scratch `/private/tmp/ch20-csharp.hH3ysE/spawn-tests-build-2.log` is clean; `spawn-tests-2.log` is 4 passed/0 failed; `spawn-shared.log` is 185 passed/0 failed; `spawn-cli-build.log` and `spawn-gui-build.log` build actual `Cli/Ch04.csproj` and `Examples/Gui/Gui.csproj` with zero warnings/errors. Shared/host chain 4169 was collected exit 0. The initial `spawn-tests-build.log` built cleanly, while `spawn-tests.log` had 1 pass/1 failure: my fixture required a primary `save.json` at turn completion. Actual History durably appends the journal during work and snapshots primary at shutdown. I corrected the test to assert activity/submission separately and inspect the child journal, not product code. A guessed model-test filename was corrected by locating actual API use. One whitespace check opened an owned pager; I identified its exact git-parent relationship and stopped only that pager, then reran with `GIT_PAGER=cat`. Unrelated pagers were untouched. Some combined reads clipped output; the new handler and status implementation were reviewed in bounded reads. Whitespace check is clean.

This completes the seven ordinary workflow tool implementations, not the complete Chapter 20 workflow. Remaining: generic incoming-message observation wake (including a child question while waiting on unrelated work and no sticky future wake), shipped skill grants/manual and CLI human-question input, a real local nested workflow through the CLI, scoped executable grading, prose and final assembled review/preservation. No Chapter 20 grade or endpoint exists. Chapter 19 remains the preserved endpoint. Earlier GUI/browser campaigns are unchanged evidence, not rerun. No workers, installation, shared restart, commit, private credential search, Bill browser/8088 or live OAuth. Local peers do not establish live provider behavior, billing, audible speech or human accessibility.

## 2026-10-28 — Chapter 20 submission/question/shutdown tools verified; spawn/status still open

I wired four more grant-controlled built-ins through the existing Tool→Agent and Job→Jobs→Agent relationships: `submit`, `send_message_to_parent`, `respond_to_parent_message`, and `shutdown_agent`. The enum additions preserve earlier values. Submission's declaration comes from the child's retained immutable contract; admission calls that same exact validator rather than the ordinary shallow field validator. Publication still revalidates under the existing durable/generation boundary, writes the whole typed payload atomically and refuses unresolved descendants. Root submission refuses. Legacy text normalization and observer-option stripping retain their existing contracts. Ordinary final text remains noncompletion.

Question tools retain their wait in an ordinary Job and use the existing two-owner question graph, receiver-scoped IDs and root human inbox. Wrong-child/ambiguous replies refuse without consuming questions. Cancelling the question Job detaches its waiter, not the admitted question. This does not yet implement CLI human input or generic question-driven interruption of an unrelated observation. `shutdown_agent` finds only a direct child, asynchronously joins its existing shutdown task, then calls destruction, which reobserves failures and attempts every detach. The selected subtree disappears even when its journal is poisoned; siblings and saved files remain. Known lifecycle/aggregate failures become actionable error results, never success. No manager, task tracker, copied service or new scheduler was added.

`Tests/ChildToolTests.cs` adds real dispatch coverage for exact large integers, nested/boolean schemas, rejected submissions, unresolved descendants, observer stripping, legacy normalization, root refusal, question ambiguity/IDs/cancellation/human answers, recursive targeted removal, saved-file retention, same-ID recreation and failed cleanup. `Tests/ChildToolActorTests.cs` reuses the loopback HTTP actor boundary for an explained model-led `submit`, typed durable completion, provider-wire declaration, correlated tool result on the following request, and a drained two-request task without corrections. Four focused groups use `--child-tool-tests` with an explicit runner exit. This is not the complete nested model-led workflow.

Setup/fixture failures are retained: `child-tools-tests-build.log` guessed nonexistent FindTool/WaitForQuestionAsync, `...build-2.log` guessed Schema; neither ran behavior. I also mistakenly said the second chain reached execution before reading its failed build, then corrected that claim. The third build ran three passing groups and one failing assumption: duplicate-key arguments fail canonical call serialization before Job execution, with an admission exception rather than a recordable tool result. The fourth run's unchanged-History assertion was also too broad because the actor records lifecycle facts around refused work. Final assertions require rejection, no ToolStarted and no durable submission; product serialization was not weakened or rewritten. Wait tests use actual CompletionAsync and non-consuming WaitForQuestionsAsync, with admission observed before cancellation.

The first shared run (`child-tools-shared.log`, job4130 exit1) was 180 passed/1 failed: the old projection test required every schema to contain additionalProperties, which legacy submit intentionally omits. It now asserts immutability for every tool and retains an explicit read_file stripping control. Review found a genuine new compatibility gap: valid boolean nested submission schemas crashed GeminiSchema's object-only projection. The retained `child-tools-boolean-baseline.log` is 3 passed/1 failed with the exact EnumerateObject/True exception. The projection now emits an unconstrained declaration node for booleans, consistent with its deliberately reduced provider schema vocabulary; the authoritative local contract still rejects false-schema values before effects. The regression verifies declaration nonmutation plus actual true/false admission. No live Gemini acceptance is claimed.

Final assembled evidence: `child-tools-final-build.log` clean; `child-tools-final-focused.log` 4/0; `child-tools-final-shared.log` 181/0; `child-tools-final-cli.log` and `child-tools-final-gui.log` build the actual Cli/Ch04.csproj and Examples/Gui/Gui.csproj with zero warnings/errors. The fail-fast chain job4136 exited0. I reviewed saved handler/declaration/dispatch/validation code and shutdown entry rather than crediting clipped sections of combined reads. Earlier GUI, ownership, submission, creation, question, completion and join evidence remains unchanged. No Chapter20 executable grade, completed child workflow or preserved endpoint exists; ch19 remains untouched.

Next: `spawn_sub_agent` with its reliable child task retained by the existing spawn Job, actor-owned `agents_status`, generic incoming-message observation wake without sticky cancellation, shipped skill/CLI human-question wiring and actual nested workflow. Then scoped grader, prose and final source review/preservation. No workers, live OAuth, private credential search, Bill's browser/8088, install/restart, commit or shared cleanup was used.

## 2026-10-28 — Chapter 20 question-aware join Job verified; remaining workflow open

I added the direct `Agent.JoinChildrenAsync` async stream and value-like `ChildJoinUpdate`, declared in Common and implemented in Agent. Enumeration captures current direct child references (all for omitted/empty IDs, or an explicit validated subset), then reads authoritative submission/failure state and reliable change pulses. It yields newly observed receiver-scoped questions once, without consuming answers, and finishes with typed submission objects. It does not cache an old successful wait across newly admitted work, treat idle as completion, adopt later children, cancel child work, or claim quiescence. A changed captured pulse forces a rescan before publishing a result. Stopped/destroyed/failed selected children produce explicit failures. Caller cancellation/disposal ends only this observer. No polling timer, per-child worker, manager, callback bag or second scheduler was added.

`join_agents` is now the first ordinary child-workflow built-in. It runs through the existing Job→Jobs→Agent ownership path, emits structured pending questions with child and message IDs, and eventually returns typed submissions. Invalid selection and child-lifecycle failure are ordinary error tool results. It remains grant-controlled like other installed tools; the shipped skill is not yet wired for the complete workflow. Other six workflow tools, CLI human-question input and the complete model-driven workflow remain unfinished.

Question progress uses `Job.AppendAttention`: it flushes ordinary output then marks the end of its unread urgent byte range. Existing observations yield that range despite a long delay or nonmatching pattern while preserving the same running handle and worker. The shared output cursor consumes attention with its bytes; there is no copied question/completion store. This addresses join questions specifically. The chapter's broader incoming-message interruption of any current tool observation and no-sticky-interrupt rule still require explicit review/integration.

`--child-join-tests` passes four real-component groups: fixed selection and empty/subset joins, exact large-number and default-text payload preservation, repeated questions and actual join/wait tool dispatch using a 120-second delay plus nonmatching pattern, no repeated consumed question, cancellation/kill isolation, unknown/duplicate/wrong-type selection errors, stopped-child tool failure, and parent shutdown. The existing loopback generation test now also asserts that a join rejects prior-generation completion and returns the new publication before task drain. No provider is used in the four new groups; ordinary submit/spawn and CLI evidence remain ahead.

Evidence in `/private/tmp/ch20-csharp.hH3ysE`: `join-build.log` clean; `join-tests-build.log` clean and `join-tests.log` four passes; final `join-final-tests-build.log` clean, `join-shared.log` **177 passed, 0 failed**, and `join-cli.log` / `join-gui.log` both actual hosts zero warnings/errors. Product edits and tests compiled and passed on their first run; no failing product baseline is claimed. Reviewed the complete stream, tool implementation, Common declarations and Job cursor/wake delta. The fixed built-in-name test now includes join_agents. Unchanged GUI/browser, previous chapters and external campaigns were not repeated. Chapter19 remains the preserved endpoint; no Chapter20 executable grade or endpoint.

Next: generic incoming-message observation wake semantics, remaining six tools with task ownership and dynamic submit declarations, CLI `/reply` and actual nested workflow, then scoped grader/prose/final review. Existing joined shutdown—not submission or Flush—is the destruction barrier.

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

