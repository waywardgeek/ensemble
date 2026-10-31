# Original Chapter 12 independent review

## Interim original-grader probe — October 10, 2026

Contributor: `enhancement_reviewer_resumed`. This is a requested interim probe,
not initial formal source review or chapter acceptance. The coder is still
validating and refining its implementation. No correction round has begun.
Read the full mandatory `.agents/skills/ensemble-coding/SKILL.md`, carryover,
workflow, original Chapter 12 and its coordinator launch erratum.

Original chapter: file revision `ec41c6e8cf0800e266f7a39ebadf255d94379082`,
SHA-256 `a754bd50ca7f3c7a2da868ff5acbdb98b4d5eddaa05349e7890c4a3dbf6cd4b9`.
The original full grader was stopped during the inherited GUI/headless parity
chain; the coder reports exit 143 and verified its owned processes were gone.
Its partial output remains **interrupted/unscored**, recorded in
`../../solutions/edition-2/evidence/ch12/grader-interrupted.json`.

After that harness release, I copied the original grader and interim student
tree into an isolated scratch directory. Repository HEAD was
`7636c43765b45641aaac43761eae34c229686ab6`; this is the base revision, not a
commit containing the uncommitted student implementation. All 89 selected source
and module-file hashes matched before copying, afterward, and in the copy.
The retained manifest identifies those actual bytes and the original grader
files. This result must not be attributed to later corrected source.

The scratch-only bridge calls unchanged `Ch12Run` and `Ch12Checks`. It explicitly
marks recursive Chapter 11 parity skipped. Command:
`go test ./internal/grade -run '^TestReviewerCh12Probe$' -count=1 -v -timeout 1m`,
with `REVIEW_SUBMISSION` pointing to the copied student and `REVIEW_REPORT` to
the scratch result. The collector exited 0 in 1.311 seconds; that means collection
succeeded, not that all grading checks passed.

| Current chapter check | Actual result |
| --- | --- |
| mcp-handshake | 15/15 |
| tool-discovery | 15/15 |
| mcp-tool-call | 15/15 |
| ephemeral-round | 20/20 |
| reverse-call | 15/15 |
| ws-tunnel | 0/10: could not find `gui/handler.go` or `internal/ws/handler.go` |

The measured current-chapter result is **80/90**. Recursive parity's remaining
10 points were **unrun**, not a measured failure. The raw check array retains
its explicit skipped marker; do not report this as a full 80/100 grade.

Independent inspection of the complete unchanged harness and check table finds
important limits:

- Handshake credit observes the initial request and ID; the harness treats the
  chapter's initialized notification as optional. Discovery credit observes a
  list request, not the resulting registry or model declarations.
- Tool-call and ephemeral credit is assigned after the protocol exchange,
  without an engine turn or actual tool invocation. The claimed separate static
  ephemeral check is absent here. These 35 points do not establish either
  behavior, delivery timing or exclusion from model-visible tools.
- Reverse credit observes a correlated response to an unknown tool; it does
  not establish successful enabled-tool execution or skill/trust enforcement.
- WebSocket grading inspects fixed file names and substrings, not a running
  browser tunnel. Its failure on the student's separate GUI layout is not proof
  of a broken tunnel; passing that source check would not prove one either.
  No dummy file, relocation or core GUI dependency should be added to appease it.

The final review must independently assess the real job/cancellation path,
round/turn ephemera, visibility and reverse dispatch, skill server lifecycle and
browser behavior through the submitted source and relevant actual evidence.
These are existing chapter obligations, not a new protocol or grader framework.
No exception or acceptance of the still-changing implementation is granted here.

Small retained artifacts are under `../../solutions/edition-2/evidence/ch12/`:
`reviewer-interim-source.json`, `reviewer-interim-bridge.go.txt`,
`reviewer-interim-result.json`, `reviewer-interim-stdout.txt` and
`reviewer-interim-exit.json`. Manifest SHA-256:
`526e412e397b7eb00e3e15a5e0624ee7717fcf7ea544b26923c11b405ae1c16c`.
The harness matches file revision `f9028379fadf8555232534d412bb2208e7c7682a`
and SHA-256 `a7c91fe21dc09412fd0bdb31427986f138ebfe45f8c3964697657aeaa6f3573d`;
the checks SHA-256 is
`9e3e74c172fb479376133a5a11b4e44cd3eb36e043502d74972fae95abeeb630`.
No original grader, chapter or student implementation was edited. No additional
paid call or broad test sweep was run. Scratch copies were removed after capture.

## Initial formal review — correction required

Reviewed the complete final student submission in `ch12.md`, production changes,
new integration tests and retained evidence at source commit
`2c459d945affd958ee547a7b6c58a2396221ce0a`. Independently verified all 91 file
hashes in `student-source.json`. Compared the student's change with accepted
original Chapter 11 (`f50837f`, exact `history/ch11`), and read the first-edition
Chapter 12 MCP/browser code and its integration changes against `solutions/ch11`.
This is the initial formal review; **one correction round is requested**.

**Blocking finding: automatic context must wait for actual completion.**
`internal/agent/mcp.go:116` gives automatic invocations the ordinary job callback
wait; `ephemera` then treats its returned status text as completed tool content.
The default wait is three seconds. A real pipe peer returning turn context after
3.4 seconds produced a provider request at three seconds containing
`job 1 running; 0 bytes total ... Use wait_for_job` instead of the actual context.
The once-per-turn flag was already set, so its eventual result is never attached.
The same peer with an immediate reply passes the identical assertion.

The chapter explicitly combines automatic results into Ephemera before
RequestSent. A model-facing job handle is not that result, particularly for a
hidden tool the model cannot call. Preserve ordinary tool callback/wait behavior,
real job supervision, turn interruption and shutdown while making automatic
collection use completed content or an actual logged failure. Add a focused
regression and restored-defect sensitivity check; no paid repeat or grader rewrite
is needed. `reviewer-slow-ephemera.json` retains both observations and
`reviewer-slow-ephemera.go.txt` the small external-peer probe. It exercises the
real Agent, jobs, codec, reducer, renderer and HTTP boundary, not mocked machinery.

One small evidence correction also belongs in the response: the browser check
labeled “persisted and effective setting” reads the DOM value and computed font
size. Those prove applied/effective state; that assertion does not read the saved
file. Describe it accurately and cite the unchanged persistence evidence
separately rather than adding a redundant new persistence suite.

No other blocking source, architecture, scope or comment finding was identified.
The proposed old-path WebSocket exception is justified: actual framing belongs
in the optional GUI module and uses the public MCP transport interface. A named
core file or substring would not establish routing. The 80/90 result remains
the **interim-source** probe, not a final-source grade. Recursive parity remains
unrun for the independently established optional-GUI/headless launch mismatch;
neither exception waives actual inherited behavior or this chapter's new tests.
Acceptance waits for the automatic-context correction.

### Comparison and teaching quality

Counts below are independent physical-line counts: exclude blanks and whole-line
`//`/`/*` comments to get active lines, retaining physical totals for context.
Go tests and JavaScript are separate. Nested student `ch05`/`ch06`/`ch07` example
Go production/tests are separate and unchanged; logs, captured SSE, Markdown,
HTML/CSS and evidence scripts are excluded. No production move needs adjustment.
Detailed totals and method are in `reviewer-counts.json`.

| Category | Student own-chapter delta: active / comments / physical | First-edition own-chapter delta: active / comments / physical |
| --- | --- | --- |
| Production Go | +932 / +116 / +1115 | +906 / +178 / +1203 |
| Go tests | +385 / +17 / +418 | +104 / +8 / +125 |
| Production JavaScript | +94 / +7 / +105 | +320 / +34 / +392 |

Student production Go totals are 6091 active / 1292 comment / 7832 physical;
legacy totals are 7631 / 2265 / 10932. Student production Go comments are about
17.5% of nonblank lines overall and 11.1% of the delta; legacy is about 22.9%
overall and 16.4% of its delta. The lighter new share is not a quota failure:
the code explains correlation ownership, why reverse execution leaves the reader
free, admission versus discovery, shutdown ordering, job versus wait lifetime,
and GUI queue reliability. Exported contracts and members have useful comments.
The newly documented inherited test fields are sensible local cleanup.

The slightly larger active Go delta implements real skill startup/unload,
per-server grants, Agent ownership and job cancellation that the old answer's
new declarations/bridge do not themselves establish. A single correlation client
instead of separate codec/client layers is a reasonable simplification. Runtime
behavior stays in implementation packages; common adds data and interfaces.
Manager→Agent→Ensemble and client→manager relationships reach admission and
diagnostics without a service bag. Raw streams are external resources, not
invented domain parents. The GUI transport keeps its actual Server parent and
uses the public core vocabulary; source and import controls preserve the optional
boundary. Browser code avoids replacing the global WebSocket constructor and
implements current control-state/truncation obligations in fewer lines. No future
resources/prompts protocol, journal, sandbox or generalized YAML framework appears.

The new tests protect observable engine requests, out-of-order correlation,
reverse file reads, job kill versus callback timeout, disconnect visibility,
per-server admission and actual subprocess reaping. The old added test chiefly
exercises handshake/discovery over a pipe. Extra tests here protect real behavior;
they do not mock away the existing job/history/runtime. The browser's direct DOM
checks do not claim to prove speech: its separate full app path observes native
speech. No mocked core behavior or mock framework was found.

### Evidence assessed independently

- Existing MCP integration tests independently pass under race detection:
  `go test -race ./ensemble -run '^TestMCP' -count=1 -timeout 20s`, 1.079 seconds.
  `reviewer-focused-checks.json` records this. Their passing result coexists with
  the separate slow-peer failure; a fast peer cannot establish the missing case.
- Read the five-module test/vet and core/GUI race outputs in `checks.json`, and
  the headless negative/GUI positive dependency controls. No broad repeat needed.
- Inspected the 16 distinct core mutation claims and six browser mutations.
  Preserved initial argument/admission survivors, the correlation mutation that
  only broke handshake, and the compile-failed close mutation are not valid kills.
  Corrected artifacts show intended failures: actual rendered result omission,
  per-server grant escape, mismatched concurrent response, live subprocess after
  unload, and hidden automatic dispatch. Restored controls pass. These demonstrate
  sensitivity for their inputs, not completeness; the slow-peer finding exposes
  a different valid timing case.
- Browser source and raw artifact show 15 checks in actual Chrome 154.0.8037.99,
  the real Go tunnel/Agent and an external provider fake. Three requests exercise
  click/input and fresh hidden snapshots; a fourth observes native speech. Native
  start was observed, the capped snapshot is 4027 UTF-8 bytes, state/truncation
  disclosure is present, and missing selectors produce tool errors. This is not
  a subjective audibility assessment. The persistence-label limit is noted above.
- All three human CLI receipts show a marker-bearing final answer, two round
  snapshots, one turn snapshot and one real local-server marker read: Anthropic
  `claude-haiku-5-5`, OpenAI `gpt-4o-mini`, Gemini `gemini-3.8-flash`. These are
  actual provider runs with a local MCP service, not live-browser/provider proof.
  Credentials are loaded only in the launcher and passed to the intended child;
  subprocess MCP peers receive a narrow ordinary environment.
- The initial Anthropic attempt has exit zero but no answer and a parser error;
  it remains a failure. The successful repeat captured empty `partial_json`
  beside valid start-block `{}`. The narrow parser fix preserves that valid
  input, and restoring the old condition kills its captured-wire regression.
  Initial frames were not retained; the student correctly labels that capture
  limitation instead of claiming inferred bytes were observed.

Correction round 1 should resolve the single behavior finding and the evidence
wording, then return source, affected checks and targeted sensitivity evidence.
The original grading history and all initial failures remain unchanged.

## Correction round 1 — accepted with recorded grader exceptions

Read the full revised student response and its affected-check/mutation records.
Verified all 92 final hashes in `student-source-round1.json` against source
commit `e5377f86be883e921c3876ea83f3cf65841c73eb`. Only three source paths changed:
the automatic collector, its existing peer helper and the new completion tests.
The initial submission and failed probe remain above and in their original
artifacts. **Chapter 12 is accepted after one correction round**, with the
fixed-path WebSocket and inherited recursive-parity exceptions described above.

The production correction adds a buffered handler-completion result for automatic
collection. It follows real job completion and rejects a killed job; ordinary
callers retain `Job.Wait` and callback semantics. The existing invocation context
lets a turn abandon collection without killing its job, while the eventual
result still reaches retained output/history. This resolves the finding without
enlarging a timeout, copying job behavior or adding a new interface/framework.
The added comments explain precisely that lifetime distinction.

I reran the **same independent faithful-peer probe** under race detection against
the corrected source: both the immediate peer and the 3.4-second peer pass.
The delayed provider request now includes `ACTUAL_SLOW_TURN_CONTEXT`. Independently
reran the interrupt and ordinary correlation/kill/callback tests under race
detection; both pass. Raw commands, outputs and source revision are in
`reviewer-round1-checks.json`. The temporary source copy was removed.

The coder's gated regressions additionally check both round and turn tools,
no provider request before completion, a new turn completing while the old
automatic call is still held, retained late output and no stale context leaking
into the newer conversation. The two restored-defect mutations fail for the
intended reasons: premature provider requests/running notices, and a cancellation
notification caused by incorrectly killing the old job on turn interrupt.
Their outputs contain no compile failure or race warning. Core tests/vet and
the complete MCP race suite pass on the restored source; no broader repeat is
needed for this correction.

The browser label now says “applied and effective.” Independently compared its
artifact with the initial commit: only that check label and the explanatory
`label_correction` changed. Captured requests, trace and observations are intact.
The revised note accurately separates inherited persistence evidence from this
DOM/effective-state assertion.

Relative to the initial review, production grows only 8 active / 5 comment / 13
physical Go lines; tests add 123 active / 6 comment / 133 physical lines and two
same-line helper replacements. The completed chapter production delta is thus
940 active / 121 comment / 1128 physical Go lines; its Go-test delta is 508 /
23 / 551. JavaScript and examples are unchanged from the initial comparison.
The additional tests cover the demonstrated lifetime defect and its relevant
interruption constraint; their size is justified. KISS, useful documentation,
actual ownership, optional GUI separation and no-mocks findings remain clear.

Acceptance does **not** revise grading history: the full original run remains
interrupted/unscored (reported exit 143), the unchanged current-chapter probe is
80/90 on interim source, and recursive parity remains unrun. The old-path
WebSocket checker cannot evaluate this required optional GUI shape; actual
browser/tunnel evidence establishes that behavior separately. Passing protocol
credits do not establish the grader's unexercised engine/tool/ephemera claims;
the inspected integration, mutation and live evidence supplies those proofs
within its stated limits. No Chapter 12 behavior or architecture requirement is
waived. The three-provider live receipts and 15 browser checks are reused with
their original source provenance; no paid, browser or full-grader rerun was
performed or implied. No unresolved finding remains.

Checkpoint manifest clarification: the coordinator's export audit found that
the initial 91-path and round-1 92-path student manifests include the ignored
generated `gui/test_ch06.log`. Their measured hashes remain preserved, but those
counts are not counts of tracked source. The accepted export contains **91
tracked files**, exactly matching the working-tree Git tree at
`edition-2-ch12-r1` (`cab2281`). Excluding that generated log changes neither
the reviewed implementation nor test results or acceptance. Earlier references
above to “92 source hashes” mean all manifest entries were checked, including
that support artifact; they do not authorize exporting generated logs.

