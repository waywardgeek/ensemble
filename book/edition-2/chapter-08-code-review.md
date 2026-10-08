# Chapter 8 independent code comparison

Review date: 2026-10-08. This is the independent grader engineer's subsequent
code-review role, not a second independently spawned reviewer. The reviewer
helped interpret the new contract, wrote acceptance checks and reviewed the
student's structure plan, but authored no student runtime. The historical
answer was first opened only after the student's implementation, initial user
runs and teaching review were preserved at
`5f9684b07fc02081ca91fa48cceeb676f27d8051`.

## Sources and current disposition

The equivalent historical scope is first-edition Chapter 9, snapshot commit
`fd93089b8f53bdde42f9846c9aa36d5d0e6a9051`, tree
`c47e315b814e814f5ffccfd08470543a71fea3b0`. The new initial runtime is
`bd5c05a62069266c6bc5bd3b1a859bc76c85f4f3`; the independently tested native-speech
repair is `cd9de3e4ec6be4a560144bf46caf9ddd54302dc2`.
[Source identities](checkpoint-evidence/ch08-review-source-identities.json)
bind the concrete files compared. Later historical commits `c2b24f6` and
`6c99a19` provide the full-state encoding correction and the audit of settings
that had no execution consumer. They are evidence, not authority to restore
obsolete architecture or change the new request-counting contract.

The [50-group deterministic reconciliation](checkpoint-evidence/ch08-full-gate-reconciled.json)
remains valid for its stated coverage. Comparative review subsequently found
a numeric representation boundary not protected by those fixtures. The new
revision repair and author reconciliation below remain open. Neither a green
gate nor this comparison accepts the still-finishing live evidence package.

## Consolidated findings

### R1: Preserve exact settings revisions in the browser

**Open; repair plan accepted.** `internal/policy/policy.go` and
`gui/internal/preferences/preferences.go` accept persisted nonnegative uint64
revisions. `gui/web/gui/connector.js` originally parsed every frame with ordinary
`JSON.parse`; `preferences.js` then copied the rounded number into
`base_revision`. This is a representation gap in the teaching as well as an
implementation defect. The contract does not limit revisions to JavaScript's
exact Number range.

The [actual-Chrome reproduction](checkpoint-evidence/ch08-review-revision-precision.json)
uses the frozen production GUI executable and a local endpoint that records
and refuses model traffic. No model request occurred. Revision 7 is the passing
positive: both domains persist revision 8. Starting at 9007199254740992 commits
9007199254740993 on disk while the browser still represents the old number.
Starting at 9007199254740993 makes both normal UI edits send base 9007199254740992;
both get `revision_conflict` and neither changes the file. Retrying the displayed
state cannot repair the lost integer.

The reviewer accepted a representation-preserving repair: retain the numeric
wire/on-disk schema, preserve unsafe revision lexemes losslessly inside the
browser, compare consecutive revisions exactly and encode only the top-level
settings base as its exact numeric token. Conflict snapshots need the same
handling. Unrelated tool arguments and strings must not be rewritten. A browser
without the required lossless parsing capability must refuse visibly before
claiming readiness or emitting a rounded command. No arbitrary new persisted
ceiling, string-number acceptance or weakened file validation was authorized.

The first browser-only repair exposed a second conversion at the same boundary.
`gui/projection.go` marshals typed observations and decodes them into
`map[string]any`, turning numbers into float64 before sending them. An actual
policy update from uint64 maximum minus one produced a precise acknowledgement
but a changed-record revision of 18446744073709552000. The repaired browser
correctly refused that out-of-range frame and reconnected. The
[provisional revision failure](checkpoint-evidence/ch08-review-revision-projection-failed.json)
binds the tested interim executable and inspected source separately; it is not a
run of frozen cd9de3e. The correction must preserve numeric tokens through server
projection while retaining typed opaque sanitization. Lower unsafe values also
need an assertion on the live applied display before reconnect can hide a mismatch.

Required narrow acceptance covers safe and unsafe loaded values, consecutive
updates, reconnect, stale/conflict-current/retry, uint64 exhaustion, and absence
of corruption to unrelated content. A deletion control must restore the original
failure from a valid positive. Normal provider work is unaffected; this repair
alone does not justify repeating paid generation.

### T1: Teach the actual native speech boundary

**Runtime repair independently accepted; author reconciliation pending.** The
student's actual native test exposed a resource shared beyond a document. A
Page-local queue and application-local native owner did not prevent one tab's
no-start timeout from cancelling another tab's speech. The initial separate
Playwright contexts did not establish ordinary same-profile behavior; the
subsequent shared-context native control did. Those observations and the repair
plan predate historical comparison and remain in the student review.

The repaired application-owned SpeechService holds an abortable Web Locks lease
before native speech and before its no-start timer. A waiter can cancel without
native effects; a holder cleans up before releasing. Pages retain their own
queues, captured preference revisions/rates and pause causes. This preserves the
ownership model instead of introducing a mutable global registry. The accepted
scope is cooperating documents in one origin/storage bucket, not unrelated sites
or profiles. Unsupported coordination remains visible and holds no speech pause.
The [independent actual-Web-Locks checks and deletions](checkpoint-evidence/ch08-speech-tabs-deletions.json)
use real Chrome leases with controlled native callbacks; they do not claim audio.
The student's native/audio receipts have separate identities and review scope.

Add this observed boundary and compatibility limit to §8.7. Explain why waiting
for another owner is different from playback that failed to start. The reviewer
withdrew an unnecessary publication-before-code hold once the existing contract
and pre-code owner plan already established authorization; do not describe that
hold as a skill requirement.

### T2: Narrow the historical claim about turning speech off

**Open author correction.** The opener currently claims that dropped false made
speech impossible to turn off in the first-edition reference. The frozen Chapter
9 client unconditionally converts an absent `tts_enabled` member to false on
application. Its particular workaround makes that claimed outcome false for
this snapshot, although the encoding still loses presence and remains a bad
complete-snapshot contract.

The [controlled historical-client check](checkpoint-evidence/ch08-review-historical-settings.json)
executes the whole frozen `gui.js` with a controlled DOM/socket. A true snapshot
first enables speech. A later snapshot with the omitted false disables both
speech and its checkbox. By contrast, omitted temperature zero leaves a draft
-5 unchanged, reproducing the incomplete-snapshot ambiguity in that field.
This is a local code control, not an invented historical user run.

Keep the real missing-field/zero lesson and its human consequence. Either source
a specific later client or live receipt for the stronger speech claim, or narrow
that sentence. Commit `c2b24f6`'s narrative is not sufficient proof when the
corresponding frozen client demonstrates otherwise. Do not restore the historical
claim that a saved temperature reached Engine: the later dead-settings audit
explicitly found no such consumer.

### C1: Explain the non-obvious boundaries in comments

**Suggested with the R1 repair, without broad refactoring.** Keep a short comment
beside the lossless decoder/encoder explaining why only protocol counters are
normalized and why whole-JSON replacement is unsafe. The existing new native
service comments already explain lease lifetime and callback fencing accurately.
The strict parser helpers' owner argument is currently unused but deliberately
keeps diagnostic access through the service; documenting that intent prevents a
future cleanup from removing the required ownership path. No execution change
or paid rerun is needed for these comments.

## Comparative assessment

| Area | Historical implementation | New implementation and review judgment |
|---|---|---|
| Responsibility | `common.SettingsStore` contains mutex, merge and file I/O; WebSocket Hub receives direct gate/log/settings services and callbacks. | Common declares values and interfaces. Agent owns policy in its spoke; optional GUI Server owns preferences. Constructors and actual parent interfaces reach Ensemble logging. Preserve this separation. |
| Execution effect | Both historical orchestration paths compare a constant 16; stored `MaxToolRounds` has no consumer despite an affirmative comment. | Actor captures applied policy at activation and checks the captured request budget, including automatic continuations. The independent limit1/17/default16 and queued/active tests exercise actual work. This is a concrete improvement, with deliberately different counting semantics from later historical corrections. |
| Persistence | State changes before unchecked `os.WriteFile`; mutex spans I/O; malformed loads and write failures are ignored. | Owned candidate, bounded strict load, checked write/sync/close/replacement, publication after successful replacement, busy/conflict errors and joined close. Fault controls distinguish pre-commit failure from committed state. The additional state machine is warranted by these promises. |
| Wire state | Omitempty erases meaningful values, permissive patches partially accept input, and broadcasts can silently drop. | Sparse presence is separate from complete values; strict domain/envelope validation, full snapshots, bounded subscriptions and contiguous initial handoff. Revision precision is the remaining R1 boundary. |
| Browser reuse | One IIFE reaches global DOM, a global mutable TTS object, mouse-only unpersisted widths, and immediate input coercion. | Application/Page/component owners, two reused ArtifactScroll instances, semantic controls, desired persisted widths, keyboard/pointer operation, applied/draft distinction and remount cleanup. Mixed-part routing and safe expansion are protected by actual-Chrome controls. |
| Speech | Queue entries contain text; playback reads the then-current global rate; native cancellation is global and callbacks are not generation-fenced. | Entries capture local preference revision/rate; shared defaults do not own local work; application FIFO plus browser lease isolates cooperating tabs. This is more machinery, but each piece protects an observed lifetime or resource boundary. |
| Comments | Settings' comment claims nonexistent execution wiring; the later correction records why that claim was harmful. | Policy's package/Close comments describe actor application and joined persistence. Native comments identify the lease scope. Preserve explanatory invariants and add the precision note rather than narrating each statement. |

The two small strict parsing/replacement implementations have similar mechanics
but belong to different optional-module owners and enforce different schemas.
Moving behavior into common merely to remove this duplication would violate the
new architecture. No shared abstraction is warranted solely by the line count.
Likewise, the public GUI consumer intentionally exposes no tools unless configured;
its correct refusal during the student's mistaken live prompt is not a runtime
regression or reason to widen tool visibility.

Historical reuse worth retaining is the two ArtifactScrolls and CSS-property
restyling without conversation reconstruction. The new implementation preserves
that economy while adding the actual accessibility, lifetime and settings
consumers that the historical four round-trip tests could not prove. The new
checks run the delivered browser and complete module/package tree; historical
graders remain unchanged, with their incompatible result retained as diagnostic.

## Validation and next disposition

No historical runtime or student source was edited by this reviewer. The new
contract, skill, architecture, current voice and chapter procedure were fully
read before this comparison. Historical code and prose were read only after the
cold checkpoint, and only rationale was sent to the student.

The revision diagnostic has a valid low-revision positive and exact intended
conflicts in both domains. Its runner refuses a different executable hash before
launch. Browser-script syntax and Python compilation checks pass. The historical
client control has its enabled positive and checks both omitted-false behavior
and the omitted-zero counterexample. These supplement the full gate; they do
not relabel the gate's original fixture failures or replace final live review.

Next: independently review and test R1's frozen repair, obtain the author's T1/T2
responses and student confirmation, then proofread the complete reconciled chapter
against the final live receipts. Final chapter acceptance remains pending those
specific gates.
