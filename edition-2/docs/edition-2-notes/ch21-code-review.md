# Original21 / Edition2 Chapter20 — independent review

## October 11 preparation and original-grader preflight

Reviewer `reviewer_astra` read all692 lines of original21, the student-safe
preparation, and reread the complete `.agents/skills/ensemble-coding/SKILL.md`.
The governing carryover/workflow were read in full for the preceding review.
Original20 is accepted; the current coder extends its direct Edition2 `ch19`
predecessor. This is **preparation only**, zero formal correction rounds. No
unfinished source, original answer, credential profile or provider was inspected;
no grader or service request was run.

The original exercise adds HTTP MCP to the existing transport and loadable-skill
seams. Its required behavior includes JSON/SSE POST replies,202 notifications,
session headers, environment authentication with a missing-key diagnostic,
meaningful transport/tool errors, and skill-authorized dispatch/unload. Preserve
the real prior runtime and cancellation/lifetime behavior. The deployed primary
must allow loading the supplied web-search skill; a fixture's primary proves
nothing about that shipped configuration. Fetched content remains tool-result
evidence, not a guarantee of model resistance to prompt injection. No Chapter22
status/accounting features or unrelated HTTP/MCP framework are authorized.

Read all four original21 grader files. The runner builds the supplied directory,
launches `--port`/`--gui-dir`, supplies skill/provider environment, and drives
stdin prompt envelopes while awaiting assistant/error completion. Thus the
optional GUI executable is its intended entry, even though the actual live
exercise may use the CLI. Unlike prior fixtures, its base explicitly declares
`type: primary`; do not reuse the missing-type exception automatically. The
base also names `think`, but these checks neither require that declaration nor
call it, so its name alone is not an observed fixture failure or waiver.

Coverage limits, to assess after the frozen implementation and actual result:

- Before/after declaration checks and real URL calls establish skill loading.
  They do not test guessed unauthorized dispatch, unload, primary no-connect,
  or the deployed primary. The fake advertises exactly the two allowed tools;
  it never presents an extra tool to filter.
- The fake returns single-frame SSE and202 notifications. It does not issue
  session IDs, demand authentication, vary JSON response framing, or exercise
  concurrent close/blocked receive/cancellation. These public obligations need
  their corresponding implementation and focused supplemental evidence.
- Search and nonce checks inspect real calls and actual tool-result bytes.
  The nonce protects against a canned fetch. They do not establish save/replay
  portability or unrestricted live-provider reliability.
- Injected text must appear only inside a tool result, which is a useful
  structural check. Its scripted model cannot demonstrate prompt-injection
  resistance or broader authorization enforcement.
- Both MCP `isError` and HTTP500 diagnostics must be actual error results.
  Matching text alone does not earn those checks. The fake's404 intentionally
  uses `isError:true`, whereas the historical real service could return a404
  page as ordinary content. Do not invent a metadata reinterpretation feature
  solely to hide that distinction.

These are test boundaries, not preemptive findings or a request to expand scope.
No score or reviewer exception is claimed. The source's assertion that skill
gating cannot be mutation-tested does not waive the mandatory targeted-defect
procedure. Meaningful existing inherited evidence can be reused; no whole-book
audit or unnecessary paid repeat follows from these observations.

Source chapter revision is `7159f0ba8708c7582f4c7d64cf09a204d644dd70`, hash
`e1d44feb80a51676e628ed4e5682484325099120d1f5a29aa5c8be060cc359c1`.
The current grader source provenance is recorded below before any student run.

| File | Last file revision | SHA-256 |
| --- | --- | --- |
| `ch21_checks.go` | `a87e6d406f6acd6e24a0885ae48e22a8ed6e51d0` | `e58283417a8e141f5409c29b6917530e574f2dfd86d990ed7bb96206d4c4a171` |
| `ch21_run.go` | `ad3a6dc7ed20d96a785e86d1880c9caac36f7162` | `63979a1d0cea261a95d72c089f814249b63d9fdbd7950f76904e9fb239706f61` |
| `ch21_harness.go` | `a87e6d406f6acd6e24a0885ae48e22a8ed6e51d0` | `f948d70153334739356cbb74ea97514406472d2f5581bca6385e1b88413eb76a` |
| `ch21_fakemcp.go` | `107c59d45ee2977d82910f1c1e734eabb73d538a` | `1fcd849b77794f85a80d23ede777571cd716a8d0bf31c6c3c0ef54b5b9c8293e` |

Wait for frozen source and the complete student review before comparative
assessment. Original22 is the next binding stop after accepted21; original24
and stopped authorship remain. Preserve CodeRhapsody's concurrent work.
Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## First frozen stage: independent probes, before final submission

Inspected source `5a00b98f489bc06b533ac612145e1102e9199a60`, main tree
`02e9bb14883f4dddc0e76c844656369bd91f6973`. No formal disposition yet; final
student review, live results and targeted mutation work remain pending.

The focused real-runtime HTTP tests and shipped-skill check pass independently
under race detection,1.143s, together with a small external-peer primary probe.
That probe confirms a departure from Rule5: constructing a primary with an HTTP
MCP declaration performs initialize/initialized/discovery and exposes its selected
tool. Existing primary loading already starts every newly enabled skill after
the root attaches MCP. This coherent inherited behavior may warrant an explicit
exception to the chapter's historical silent-no-connect trap rather than a
deliberate regression. The coordinator asked the student to explain the departure
in its review; no waiver or implementation change is imposed before that response.

The independent **no-receive-broadcast mutant survives** both new HTTP tests,
exit0 in0.025s. Removing only `t.ready.Broadcast()` preserves their apparent
unload success because `client.close` closes its Done channel before transport
shutdown. That channel is not evidence that the already-parked HTTP `Recv` exits.
The production wakeup is present; this is a sensitivity gap, not an observed
production failure. Return the finding for a focused receive-lifetime check and
restored-defect proof before final submission. No additional runtime capability
or paid run is requested. This is attributed pre-submission review assistance,
zero formal correction rounds. See [probe receipt](../../solutions/evidence/ch21/reviewer-initial-probes.json)
and [primary probe](../../solutions/evidence/ch21/reviewer-primary-probe.go.txt).
Both isolated scratch checks were removed afterward.

The original grader's completion wait also disagrees with the existing GUI
terminal: the harness requires JSON `assistant`/`error` on stdout, while this
command prints raw deltas/newlines and errors on stderr. Visible remote calls
can complete while the harness waits90seconds. The eventual score and exact
failure remain for the final evidence review; this source observation does not
turn visible tool output into proof of every masked rubric assertion.

Initial [comparative counts](../../solutions/evidence/ch21/reviewer-counts.json)
are tied to this frozen stage. Original21 has no original20 snapshot predecessor,
so original19→21 includes the crossover's unrelated model, cache, actor and GUI
repairs. Do not attribute all that growth to the HTTP feature. Relevant original
HTTP transport/tests and integration were read separately. Final counts and
comparative judgment will follow the submitted source, not the moving worktree.
Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## Final disposition: accepted with two narrow exceptions, initial review

Read the complete final student review and frozen submission
`f35b6ae8ce43c91a52b9972361093db6258e3e20`, main tree
`f351d66a7bc456650e05ac6656a7db59bbbd725a`. Independently verified all148 manifest
files. **Original21 / Edition2 Chapter20 is accepted with the two exceptions
below, zero formal correction rounds.** The earlier reviewer assistance and
failed live attempt remain pre-submission history, not erased or relabeled.
No current implementation or scope finding remains open.

### Exceptions and actual original result

The unchanged original grade remains **0/100 FAIL on5a00b98**, make exit2/grader
exit1,181seconds. Its raw SHA-256 independently matches
`83384811db3bb4aad2ed7af333039cc84e932e51e341e8e3d657f0ebaf2b38df`.
I accept the student's terminal-envelope exception: the existing GUI command
accepts prompt input but presents human terminal output, whereas this harness
waits for legacy JSON completion. Both scenarios perform visible remote work
then wait90seconds; the first fatal masks all seven score predicates. This is
not a passing grade, a final-source grade, or proof from visible output alone.
Adding a second terminal protocol solely for that fixture is unnecessary here.
The valid masked behavioral obligations remain binding and are covered below.

I also accept the requested Rule5 exception. Existing primary loading starts
declared MCP peers coherently after the composition root attaches the manager.
The unchanged independent external-peer probe confirms that behavior on the
final source. Deliberately making valid primary configuration silently do
nothing would recreate the defect the chapter itself criticizes. The actual
web skill remains loadable, does not connect before load, grants only its
declared names, and unloads immediately. This exception preserves explicit
primary configuration; it does not authorize automatic web access or expand
remote grants. Both departures should be explained in later prose/grader work,
while preserving this original chapter and score.

### Implementation, comparative review and KISS

The complete source delta uses the existing manager/client/registry/actor/jobs.
URL/auth-name fields enter existing shared data and parsing. The manager owns
the HTTP boundary; its parent interface reaches the real Agent/root logger.
Agent still owns skill/configuration state and Engine retains rendering/usage.
The small optional context-aware send interface is justified: HTTP Send can
block before the client's ordinary reply-select, so job cancellation must reach
the actual POST. Connection close cancels all requests and wakes Recv. No
callback bag, new registry, mutable application global or GUI dependency appears.

Read the corresponding original HTTP implementation, deterministic and opt-in
live tests, skill parsing, host connection/auth wiring and MCP enablement test.
The original uses a whole-body parser,300-character diagnostic clipping,
host-owned callback wiring and a fixed HTTP timeout. The student retains complete
diagnostics except its own credential, supports multiline SSE without a scanner
cap, and connects cancellation to actual job/connection lifetime. Its integration
tests exercise real earlier components through external HTTP peers. The original
enablement test substitutes the connect callback and asserts registry state;
the student also tries the unauthorized call through actual Agent dispatch.

[Counts](../../solutions/evidence/ch21/reviewer-counts.json) use physical
source lines and active lines excluding blanks/comment-only lines; inline
comments remain active. Each edition uses its own predecessor; tests, examples
and languages are separate, with Markdown, manifests, module metadata, scripts
and evidence excluded. No student source moves hide growth.

| Delta | Physical / active / comment-only |
| --- | ---: |
| Student production Go | +230 / +192 / +25 |
| Student Go tests | +405 / +369 / +27 |
| Original19→21 production Go | +1472 / +856 / +513 |
| Original19→21 Go tests | +2973 / +2440 / +333 |
| Original19→21 JavaScript | +48 / +12 / +34 |
| Original19→21 HTML | +6 / +6 / 0 |
| Original19→21 CSS | +74 / +43 / +21 |

There is no original20 snapshot. Original19→21 includes many unrelated crossover
repairs, so those totals cannot measure HTTP alone. A useful narrower comparison
is the new HTTP file: student204/172/20 versus original183/138/31. The original's
direct HTTP tests are245/202/27, plus111/89/11 opt-in live-test lines; its separate
grant test and host/parser changes are also relevant. The student's modestly
larger transport pays for owner access, cancellable I/O, correct frame parsing
and secret handling, not speculative features. The405 test lines include real
runtime integration, actual blocked-job cancellation, deployed-skill regression,
parked receive and redirect boundaries. Test volume alone is not an improvement.

Production comment share in the delta is about11.5%, below the rough20% target.
Useful explanations cover why the queue is unbounded, why HTTP needs job context,
the two cancellation lifetimes, complete frame handling, owner access and
credential boundaries. Existing explanations were preserved. All new exported
types, fields, methods, functions and test entrypoints have useful attached
documentation. Making the internal constructor exported permits the external
boundary test to construct the real transport with its actual owner; it creates
no additional abstraction. No padding or arbitrary line quota is warranted.

### Verification and limits of the tests

Read the final core/GUI test and vet receipts; they pass. Independently reran
the final focused source with race detection: HTTP package1.018s and real-runtime
suite1.148s. That run includes the primary-departure probe. The exact formerly
surviving Broadcast removal now fails **close did not wake receiver** after one
second, with no compile failure or unrelated timeout. The new check observes
an actual receiver parked in Cond.Wait before closing; client Done is no longer
misreported as receive completion. See [final receipt](../../solutions/evidence/ch21/reviewer-final-check.json).

Read all25 applied mutation outputs. Twenty-four have the intended behavioral
failure; repeated grant/unload edits record strengthened assertions, not25 unique
behaviors. The retained storage-detach survivor is correctly scoped: closed
clients remain invisible for the observed capability-removal claim. It does
not prove whole-program equivalence or correct future reconnection. Skipping
unload fails connection-close evidence; independent skill admission still blocks
the attempted call. Missing grant enforcement now fails both declarations and
actual guessed dispatch. Header/session/framing, arguments/result bytes and
kind, error flags/bodies, both credential-redaction boundaries, shipped skill
access, cancellation, redirects and receive wakeup have causal checks. The5second
kill/unload failures exercise actual blocked POSTs with working controls.

The masked grader's search/nonce/result-kind/error obligations are supported by
the real-Agent HTTP integration and live results, not merely the original score.
The marker assertion inspects the actual provider tool-result block. Source
review confirms fetched data does not mutate tool permissions or become system
instructions; actual positive/negative dispatch controls protect the grant.
The tests do not establish general prompt-injection immunity. The multiline
progress notification in the framing fixture is intentionally ignored by the
client, so its presence alone is not a separate delivery guarantee. No mock
replaces the earlier runtime, and no extra paid run or original-grader repeat
was necessary after the final local redaction change.

### Actual live task, including the failed first attempt

The first attempt completed transport work but answered before collecting the
slow scrape. Its primary did not expose the existing wait_for_job capability.
I accept the student's diagnosis and small correction: expose that supervisor
and explain running notices in the web skill. Preserve the initial incomplete
answer and27.04second process success; they are not a successful research task.

The corrected source is `2eac96a`, main tree
`c257777ba841665766b6ea6c5cd54480678e6822`; the later reflected-RPC redaction and
test changes have local evidence, not retroactive live provenance. I independently
checked all five raw request/completed-response pairs: Astra/high throughout,
no remote tools in request1, then exactly search/scrape from the observed28-tool
inventory. Actual correlated calls are load, one search, one scrape with maxAge0,
and one wait_for_job. The scrape initially returns a running notice; the wait's
actual output in the final provider request contains the requested filename and
the explicit1,763,735-byte omission notice. It is not merely an earlier argument
or a model-authored assertion.

The retained full scratch job is1,775,735bytes and independently matches the
excerpt receipt's SHA-256
`73f611c1eafc3a5de053b7f9005a4efb5550cfeea3bff40491551fe3a3d87c55`.
It contains the official Go page, HTTP200 metadata and the filename cited in the
answer. Only the selected12,000 output bytes plus job metadata reached the model;
the final answer accurately discloses truncation. The live record shows23.35seconds,
normal CLI exit0 and save. The final provider response reports4,096 cached input
tokens; no general cache or reliability conclusion follows from that one grant.

The launcher selects explicit ChatGPT billing and the dedicated profile, omits
model API keys and passes the named Firecrawl key. Source and runtime billing
announcement support that route; no fallback is implemented or observed. I did
not read credentials or make provider calls. Endpoint provenance remains explicit
configuration plus source, not independently intercepted URL/auth headers. The
final receipt records one reviewer assertion wording correction: the job says
“bytes omitted,” not the literal word “truncated.” It was not a product defect.

Coordinator may export exact direct Edition2 `ch20` and create the immutable
checkpoint, then **STOP before original22** for Bill's placement decision. Do
not implement or draft22, merge it, skip it or advance to23. Original24 and
stopped authorship remain binding. Preserve CodeRhapsody's concurrent changes,
all original sources and historical results. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.
