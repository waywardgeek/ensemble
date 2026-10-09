# Chapter10 grouped post-run quality handback

Source and local browser check: `70d86f7419c82fcf7cb8a394d54e472feccd2eed`.
Initial live record remains `44d762789ea65a1e11fd7982ec9b748d4a8112c1`, on runtime
`57d4aac` / support `9822b2b`. No initial receipt, screenshot, signature, part or
request was changed. This task used no credentials, discovery or real-provider
calls. Independent review of this revision remains pending.

| Request | Disposition and invariant |
|---|---|
| Q1 | `Skills.LastActivation()` reads the committed scalar under the existing Actor/Agent or inert-construction confinement. Common declares the interface; Skills implements it through its existing parent. `watermarks` and the initialization check no longer build inspections. Capture retains one full owned Snapshot; material, transitions, retired identities and represented maxima are unchanged. No new state, cache, lock or sibling dependency. |
| Q2 | The first Watermarks definition now requires exact activation/job maxima, zero absent, including retired and out-of-window history. The separate request lower bound and burned-admission exception remain. No runtime format change. |
| Q4 | Optional Artifact presentation recognizes only `type: text` with an own `text` member exactly equal to the empty string. Its existing key/order becomes a compact “Empty response text” indication. Speech/expansion controls are hidden, disabled and guarded; nonempty replacement restores ordinary controls. Projection, signed-empty preservation, speech protocol, reducers and raw history are unchanged. |

Q1's test covers zero before initialization, an uncommitted candidate, retired
identities, unchanged/failed operations, an allocation-free scalar read, and
returned snapshot material/transition isolation. Existing public capture/import/
Skills tests provide broader behavior coverage. No benchmark framework was added.

The local headed-browser check uses the exact original Gemini B and B-restart
captures. Raw response5/20 part1 are ordinary empties; response10/15 part1 are
signed empties. All four projected positions remain present. It separately feeds
the recorded B live-final observations and restart snapshot, checks two identical
resets with no duplicate cards, preserves provisional-to-durable identity, and
tests nonempty replacement, absent text, nonempty whitespace, opaque placeholders
and empty tool-result lifecycle. Empty automatic/manual speech admits nothing;
nonempty manual and automatic positives admit work through the real SpeechQueue
using a local service probe. This is not a native-hearing or new provider claim.

[Local screenshot](quality-20261009/browser/captured-restart.png),
[DOM](quality-20261009/browser/dom.txt), and
[43-check result](quality-20261009/browser/result.json) are new captured-data
evidence. The harness validates its immutable source revision, the complete served
asset set, original sealed files against44d7627, actual Node/Chrome binaries and
the retained browser dependency map before creating derived output. Its intended
source-mismatch control starts from that passing parent, changes one owned asset,
refuses at the asset hash before output-directory creation, then restores the
exact bytes. Original live support/bindings were not relabeled for this check.

Commands, timestamps, outputs and failures are in
[quality-20261009/commands.jsonl](quality-20261009/commands.jsonl).

| Local command/check | Result |
|---|---|
| Affected `gofmt -l` | empty output |
| Core and optional GUI `go vet ./...` | pass |
| Core and optional GUI `go test ./... -count=1` | pass |
| Core `go test -race ./... -run 'Session\|Skill\|Snapshot\|Checkpoint\|Projection\|LastActivation\|NarrowReads' -count=1` | pass |
| GUI `go test -race ./... -count=1` | pass |
| Published `accept_ch10_public.py MAIN --receipt ...` | all16 top-level public groups pass under race |
| Published `accept_ch09.py CLI --receipt ...` | 51/51 |
| Published `accept_ch10.py CLI --receipt ...` | 93/93 |
| Published `accept_ch10_clients.py CLI GUI --source-directory MAIN --receipt ...` | 9/9 |
| Published `accept_ch10_faults.py MAIN --source-commit 70d86f7… --run '^TestCh10ReviewAppendFailure' --receipt ...` | pass; source unchanged |
| Revision-bound headed captured-data component check | 43 checks pass; zero provider calls |
| Intended source-identity negative | refused before derived writes, as intended |
| Diff against44d7627 initial live directory/handback/binding | empty |

Fresh CLI and GUI binaries remain in `evidence/ch10/quality-build/` and are not
committed. [Build inputs](quality-20261009/build-inputs.json) bind164 files and
discover12 nested modules; [build association](quality-20261009/build-association.json)
records real commands, binary hashes and Go build metadata. This is not a claim
that root tests cover every nested module: core/GUI were directly tested, and the
public black-box runner compiled its separate external consumer. Unaffected
example/support modules were discovered, not given an unnecessary broad rerun.

The only unexpected validation failure was evidence-tool setup: the first vet
command omitted HOME from its isolated environment, so Go could not locate the
module cache. `core-vet.out` preserves that failure. Restoring HOME and explicit
Go cache-path variables fixed it; no cache was removed. The source-identity refusal
is an intentional negative, not an application failure. No grader implementation
was inspected, changed or weakened.

The complete0e754b9 chapter/direct feedback were manifest/Git verified and read.
The walkthrough matches the frozen initial experience; the new opener accurately
acknowledges preceding logs/offline rendering. Q4's precise two-signed/two-ordinary
diagnosis supersedes my tentative metadata explanation. The prose may clarify
that the actual OpenAI argument string inspected here is compact: byte-exact live
replay is demonstrated, while the deliberately spaced-string regression remains
separate local evidence. No new ownership or persistence-contract gap was found.

Reuse scope: Q1 changes capture cost without changing captured values; Q2 changes
only prose. Q4 deliberately supersedes the old empty-card appearance, now covered
locally using those original inputs. No other initial live observation is
invalidated by these changes, and no additional paid run is proposed. The new
binary has not itself run with a real provider. Coordinator/reviewer should judge
that limited evidence reuse; initial live acceptance is not silently transferred
to a later build. Compiler work is complete and its slot is released. No push,
chapter export or tag.
