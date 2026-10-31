# Edition 2 Chapter 13 grader: The Channel Nobody Tested

Grader coder: GPT-6 Astra/high. Original Chapter14 is the source; accepted
implementation is direct `solutions/edition-2/ch13`, legacy is `solutions/ch14`.
Mandatory `.agents/skills/ensemble-coding/SKILL.md`, full carryover and workflow
were reloaded after compaction. Independent final acceptance is pending.

## Scope and preserved history

The driver retains the actual original weights:20 discrimination,20 filters,
10 buffering,15 provider nonstream,10 gate,10 errors,5 boundaries,5 identifiers,
plus5 corrected Edition2 Chapter12 parity. The original printed table disagreed.
It executes the already published tts-harness input/output contract, using the
existing provider fake and actual submitted executable, GUI, renderer and speech
pipeline. No implementation-name detector, new persona protocol, product feature,
original grader edit or frozen-source edit is introduced.

Original Chapter14 remains revision `994bf218`, SHA256
`d39aa61438f05cfe1ff9b496b42952e23d16a67f9ea33aa5f5a51c35e5ee14b2`.
The student’s original95/100 FAIL remains the earlier build’s measurement;
its95 speech points preceded the completed blind client and EOF correction.
The complete student/comparative records are [ch14.md](ch14.md) and
[ch14-code-review.md](ch14-code-review.md). No fresh paid demonstration was run.

## Corrections and measurements

All new receipts are in
[`evidence/ch14/enhancement13`](../../solutions/evidence/ch14/enhancement13).
`probe.go.txt` collects the actual focused check array: its exit0 means collection
succeeded, not that every check passed. The committed regression test asserts
all checks instead.

| Receipt | Actual result | Interpretation |
| --- | --- | --- |
| `accepted-initial.txt` |0/95 focused | Playwright dependency absent; no speech behavior observed |
| `accepted-dependencies.txt` |75/95 focused | Actual relative read failed; silence had no successful result to exclude |
| `accepted-absolute.txt` |95/95 focused | Existing absolute-path read succeeds; original single-bullet fixture |
| `legacy-initial.txt` |95/95 focused | Same single-bullet coverage limitation |
| `accepted-adjacent-bullets.txt` |95/95 focused | Corrected ordinary two-item list |
| `legacy-adjacent-bullets.txt` |75/95 focused | Second dash actually remains in speech in both delivery modes |
| `accepted-full.json` |100/100 full PASS,29.634s | Actual complete command at `1913a0d7` |
| `legacy-undeclared-model-interrupted-run.json` |unscored,42.005s,SIGTERM | Coder supplied an undeclared inherited model; stopped invalid fixture rather than repeating its timeout chain |
| `legacy-full.json` |75/100 full FAIL,41.824s | Declared default Sonnet5 fixture;75/95 speech and0/5 parity |
| `accepted-final-race.txt` |all final95 speech points pass under race | Final boundary assertions; package13.700s, command14.776s |
| `legacy-final-focused.txt` |75/95 focused,26.625s | Final boundary assertions; only the known filter check fails |
| `*-final-composed.json` |accepted100/100, legacy75/100 composed | Final speech plus unchanged earlier full parity; not another complete run |

Installing locked optional GUI dependencies with `npm ci --ignore-scripts`
changed no tracked source. The file fixture error was different: the supplied
harness launches its GUI from a scratch cwd while `WORKSPACE` supplies an Agent
configuration value. The existing reader uses `os.ReadFile(args.Path)`; only
command execution uses `owner.Config().Workspace` as its cwd. Its actual correlated
result reported `open notes.txt: no such file or directory`. The corrected fixture
supplies the planted file’s absolute public path. Root and reviewer independently
verified this disposition. It is not evidence of a speech-filter runtime defect.

The revised discrimination check requires the successful correlated read result
containing ZEBRAFISH, then requires it absent from speech while the distinct
assistant PELICAN occurs once. A successful-answer sentinel cannot substitute for
proof the file bytes arrived. The original silence check lacked that positive
control.

Root then noticed that one isolated bullet avoided the documented legacy defect.
Two adjacent lines now retain their words while forbidding both bullet markers.
The actual legacy transcript is `Bullet words - Following bullet.`. Both earlier
95/95 receipts remain, rather than being silently relabeled. The historical
`book/review-ch14.md` diagnosis is supported by this newly observed behavior;
its obsolete module-loader proposal was not reinstated.

The source audit subsequently found that every exact boundary phrase had a blank
line after it. Removing punctuation splitting could therefore escape those
assertions. Final code also requires the two already present adjacent sentences
in one paragraph as separate utterances. The earlier actual full reports retain
`1913a0d7` provenance; final affected speech checks and unchanged parity are reported
separately rather than claiming another complete run.

The legacy parity diagnostics include actual missing startup instructions,
missing stdio startup, skill and null-save failures. They also include absent
separate consumers, consumer permission-denied, different GUI module layout,
unsupported inherited media-model fixtures and masked downstream observations.
Not every nested message independently diagnoses a runtime defect. Valid observed
failures already justify zero parity; no recursive rerun was spent repairing every
historically incompatible fixture. The initial wrongly selected Haiku fixture is
preserved as our setup error, not a legacy failure.

## What the numeric checks establish

Complete recognizable utterances replace a loose utterance-count bound. A single
wrap remains inside one utterance; blank lines, actual adjacent sentences and an
unpunctuated final tail exercise distinct boundaries. Known phrases occur once;
ordinary capitalization stays intact while camel, snake and acronym boundaries
expand. Markup checks require preserved content as well as absent syntax. A
fence announcement must appear between unique surrounding sentences; its wording
is free, and unrelated introductory text cannot satisfy that location.

Provider streaming/plain mode is observed on the actual requests. Plain output
must be spoken once, but this does **not** establish the renderer’s final-only
branch: the accepted Engine synthesizes display deltas for provider nonstream
responses. Failure speech must mention a failure rather than merely exist.
The executable must exit successfully; a readable partial TTS_LOG cannot hide a
nonzero exit as the old harness allowed. Entire harness process groups are bounded
on failure while the harness retains normal child ownership and cleanup.

The gate’s public log proves ordered typing pause/release and final release.
It cannot identify actual field edits or both independent causes: initial prompt
typing can supply that edge even when mid-turn whitespace changes no edge.
Its title now describes tool-dispatch pause accurately. Whitespace, two-cause
ordering and Agent-wide other-tab unpause retain their separate browser proof.

## Exact published samples

`samples.json` binds manuscript SHA256
`23362a10028368a918a64e28a5da8823408c90f5255848b2f600723b3706e64b`.
Both exact application builds succeeded. The exact primary YAML loads through the
shown GUI command; the listener runs the shown task and relative peer path with
the documented existing-Chrome option. Its real public consuming Agent advertises
exactly hear/type/submit/wait/sleep. One target and six listener requests drive all
five tools. The correlated wait result contains actual filtered `queue Chunk`,
`max tool rounds`, `HTTP Server` and mango; later hear sees the advanced cursor.
The public-client probe passed in1.265s. An initial undefined helper in that temporary
probe failed compilation and was corrected before execution; it is not a runtime
failure. The driver, actual event history and decoded requests are retained.

The three current JavaScript excerpts bind to accepted source after removing
whitespace/comments. The historical advertising comment is explicitly historical.
The exact `--describe` output equals the published JSON. Playwright was already
installed for grading; the documented existing-browser route was exercised rather
than downloading another browser. The published default grader command is the
actual accepted full run above. These are executable/sample observations with
external provider fakes, not new paid task compliance or native listening claims.

## Separate retained obligations

The accepted `gui/speech.test.mjs` and retained `mutation-corrections.json` supply
real final-only WebSocket delivery, duplicate-final and no-audio/transcript
controls. The initial whole-final mutant survived the nonstream-provider test;
the corrected direct delivery check fails it. That history is preserved.
`boundary-mutations.json` retains explicit tool/error flush and paragraph controls;
`persona-mutations.json` retains actual type/submit/wait/cursor failures.

Peer tool-list refusal alone is insufficient. The delivered
`gui/cmd/virtual-user` creates a real public Agent without builtins or skill
lifecycle, connects the peer and checks its declarations. Existing source tests
exercise the actual consuming binary, not only the peer’s whitelist. The sample
run independently observes that real five-tool declaration and successful calls.
No new general blind-driver interface is imposed by the numeric grader.

Pending-sleep EOF is a lifetime obligation, not an ordinary sleep limit. Retained
`round1-checks.json`, `round1-mutation.json` and independent
`reviewer-round1-peer-eof.json` establish normal live sleep, actual client shutdown
with an active sleep job, and corrected peer exit127ms versus earlier3865ms.
Restoring browser-only cleanup fails both pending-peer and pending-client bounds.
The corrected shared process shutdown remains unchanged.

`native-speech.json` retains actual Chrome start/end controls, human whitespace,
programmatic cause separation, actual submit and explicit other-tab unpause.
These do not provide subjective intelligibility judgments or a general claim
about simultaneous talking tabs. Historical Anthropic and OpenAI target-content
failures, Gemini’s extra thinking entries and all live source versions remain as
recorded. Successful listener exit does not mean the target followed its task.

## Simplicity, documentation and final validation

The new driver has364 physical Go lines:309 active,42 comment-only,13 blank.
Its focused real-command test has20 physical,16 active,2 comment-only,2 blank.
The three original grader files total765 physical,579 active and116 comment-only;
the new driver reuses their rubric and the existing external provider fake while
replacing their launch/assertion logic. These are grader counts, not a runtime
implementation-size comparison. The command only adds Chapter13 dispatch/help.
`source-counts.json` records the exact method. All newly exported symbols and
members have useful comments: Chapter13, seven fields and the regression test.
About12% of nonblank driver lines are explanatory comments; the consequential
fixture, evidence and lifecycle boundaries are explained without padding.

Final focused race, vet, command tests and formatting pass; the final legacy
focused score remains75/95. `source-binding.json` records exact grader/manuscript
hashes and both unchanged frozen tree identities. No full parity rerun followed
the sentence-assertion correction. Final totals are explicitly composed100/100
accepted and75/100 legacy, alongside the actual earlier full reports.

`mutations.json` records19 targeted cases at the first-stage grader;
`additional-mutations.json` records punctuation segmentation and ignored provider
nonstream mode at final code. All21 fail their intended checks. Full patches and
raw failed arrays remain; compiled probe exit0 is never treated as mutant success.
The original punctuation-removal control survives the earlier assertions95/95;
`sentence-boundary-initial-survivor.txt` preserves that measured gap. The final
same defect fails the two adjacent-sentence assertions. The order was a source
audit, assertion correction, and before/after probe comparison; it was not an
initially failing baseline.

Mutations are not21 independent general proofs. Fragment flushing and duplicate
finals cause coupled failures. Disabled gate release also times out the actual
read phase, while its dedicated gate phase directly reports typing=true,
resumed=false,paused=true; the timeout is not substituted for that assertion.
The empty-read replacement also touches an unexecuted edit-file reader; no edit
call occurs, and only the missing actual read-result assertion fails. Restored
accepted checks pass. Three final positive controls score95/95: a different valid
fence announcement and actual OpenAI/Gemini harness configurations. These verify
wording freedom and the alternate provider result branches rather than merely
unit-testing our decoding assumption.

The nonzero-exit mutation runs the real harness and then exits17. A separate
single-phase control additionally parses the produced JSON log and confirms19
records before deliberately exiting17: the grader returns no accepted entries
and preserves the failure despite that valid artifact. This is a targeted runner
control, not another mock or product protocol.

Independent final review remains pending. This submission does not release
Edition2 Chapter14.
