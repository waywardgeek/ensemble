# Chapter 8 initial student review

## Source and exposure (before implementation)

The initial student read root AGENTS.md, the entire mandatory
`book/edition-2/skills/ensemble-coding/SKILL.md`, architecture.md, the full
committed Chapter 8 teaching copy and manifest from
`/Users/bill/projects/ensemble-edition-2-revisions/ch08-student-inputs/`.
Relevant preceding teaching reads cover Chapter 1 §§1.1–1.5 and TL;DR;
Chapter 2 TL;DR and §2.7; Chapter 3 §§3.2–3.4; Chapter 4 §§4.1–4.2;
Chapter 5 §§5.1–5.3 and §§5.4–5.5; Chapter 6 TL;DR and streaming ownership;
Chapter 7 §§7.1–7.7 and §7.9. Some large reads were truncated by the tool;
ownership and wire sections were subsequently read in smaller chunks.
Only accepted preceding main source has been inspected. No first-edition
chapter, old implementation, future chapter, grader implementation, reviewer
script or answer history has been read. Repository status exposed unrelated
untracked filenames only; those files were not opened.

Accepted predecessor: edition-2-ch07-r1, commit
ed5667b38362875c578e88609f4a0fce61946c53. Its main tree is
8fb543cc78eccdd9737bd3a690ee9601193fecba, matching the starting tracked source.
Manifest names accepted source 48976424ab75924f98f4ad01f75f8e046ec26872.
Initial black-box checks against the unmodified Chapter 7 GUI binary failed
at startup because Chapter 8 flags/features are absent. Raw results are retained
in baseline-wire.txt and baseline-validation.txt; neither made a model call.

## Structure, ownership, concurrency and lifetime plan (contract check pending)

- Ensemble remains application/logger owner. Add a synchronized application-local
  resolved settings-path claim registry, solely to prevent two live owners from
  using the same path. Each owning service releases its claim after close joins
  its writer. No global registry, credentials, or Config in settings state.
- Agent owns an execution-policy service, constructed by the public root, with
  shared values/interfaces in core internal/common and private implementation
  in internal/policy. The service stores an Agent interface back-pointer.
  Actor reaches it through Agent. Public getter returns an owned value; public
  update enters the actor mailbox. PolicyPath is creation-only configuration.
- Policy admission/validation and base revision checks run at the actor boundary.
  A service-owned candidate writer does file work outside the actor; its result
  returns as an owned mailbox fact. Applied state remains old until actor accepts
  success, publishes policy_changed, then resolves the update acknowledgement.
  One candidate per domain; concurrent writes get busy, stale bases conflict.
  Close stops admission, joins a started writer and applies its real commit result
  before shutting down the actor. A replacement cannot be reported rolled back.
- Each turn captures policy revision, stored value and effective limit at activation;
  all model-request decisions use this immutable capture. Recorded new turn starts
  carry the capture; strict event validation permits historical absence only.
  Watch snapshots expose current policy and nullable active captured limit.
- Optional GUI Server owns its preferences service and path. GUI common declares
  values/interfaces; a GUI preferences package implements parsing/persistence and
  subscriptions through a Server parent interface, reaching application logging.
  Its short mutex serializes admission, applied state and subscription cuts; disk
  I/O is an owned worker outside this lock. Close refuses updates and joins it.
- A preference subscription atomically captures snapshot plus bounded tail. The
  connection sends that snapshot before the Agent snapshot group, then drains
  newer changes only after snapshot_end. Update replies wait until the originating
  connection has queued their changed revision. Settings workers never park the
  transport reader; connection lifetime cancels delivery, not committed writes.
- Browser Page owns Connector, two reusable ArtifactScrolls, input, draft controls,
  speech queue and pause causes. Applied preferences are shared snapshots; queued
  utterances capture revision/rate. BrowserApplication still owns native speech.
  Changes style existing DOM and preserve card identities/focus/expansion. Pane
  routing is per typed part/event. Divider changes persist at completed gestures.
  Socket-instance fencing protects settings before the Agent generation exists.

No architectural ambiguity found in the published ownership contract. This plan
asks the coordinator to verify lifecycle and path-claim choices before affected code.

## Initial teaching experience

The save/apply/announce ordering, zero-versus-omission distinction, explicit
turn activation boundary and separate shared/local speech semantics are useful
and concrete. No missing requirement has yet blocked implementation. Failures,
implementation difficulties and clarifications will be appended when observed.

## Baseline build issue and coordinator assistance

Before implementation, root-module `go vet ./... && go test ./... -count=1`
failed because accepted Chapter 7 `evidence/ch07/revision-1/replay-server.go`
imports the optional GUI module outside a support module. GUI vet/tests pass.
The coordinator independently reproduced this and assigned build isolation to
its Chapter 7 owner; Chapter 8 implementation is held until immutable r2 release.
The coordinator accepted the ownership plan, including actor-ordered application,
joined writers, application-local path claims and Page-owned transient state.
No Chapter 8 source has been changed during this hold.

## Proposed live feature/action matrix (no paid runs yet)

Use the exact frozen source and independently hashed CLI, GUI and consumer
binaries. Preflight includes complete immutable source/support/dependency maps;
all verifier identity refusals begin from valid passing paths. Keep raw terminal,
GUI trace, browser observations, HTTP/usage receipts and audio originals distinct
from derived reports. Keys stay in child environment/memory. Every run uses a
fresh log/workspace. No automatic provider retries or substituted Gemini model.

For each Anthropic, OpenAI and Gemini backend:

| Surface / action | Required observed result |
|---|---|
| Real GUI, two tabs: inspect initial applied state; change theme/font, keyboard and pointer dividers; shrink/enlarge viewport | Complete shared controls and rendered CSS agree; keyboard moves 10 px, desired widths persist without viewport rewrites; central input remains reachable |
| Two tabs: concurrent same-base settings, invalid draft, deliberate current-base retry | Busy/conflict is useful and correlated, authoritative values remain clear; retry preserves competing applied fields; no model call |
| Close/restart same preference/policy files with fresh log | Nondefault positive values and revision return from server; explicit false and zero also survive a second restart; history is fresh |
| Actual browser synthesis with autoplay on and later rate change | Audible/captured synthesized output, future utterance rate changes; explicit card action works with autoplay off |
| Two pages speaking, B typing; shared autoplay off then A Cancel | Existing queued speech retains captured rate/revision, future automatic text stays silent; A cancel leaves B work/typing pause intact. Exact callbacks also covered locally |
| Browser policy=1, ask for exactly one scratch-file read then report | Accepted model tool batch finishes, round_limit before HTTP2; capture and displayed active policy match |
| Browser policy=0, repeat read/report | Stored zero displayed Default(16), continuation actually occurs; no vendor policy field |
| Browser bounded scratch-file task with text plus calls; expand retained result and inspect actions/chat | Tool proposals/results/jobs in actions, human/answer/thinking in chat; same safe reusable cards, full expansion and stable focus/scroll |
| Browser correction while streaming, pause before next admission, interrupt; reconnect during and after work | Hint acknowledgement, truthful pause boundary and interruption; settings/conversation converge; no prompt replay/history speech |
| Human CLI in actual PTY attached with --terminal: set policy=1 from browser, type read/report task and inspect response | Same Agent enforces policy; round-limit CLI error/cleanup retained. Observe output before follow-up/local command; retain original failure receipt |
| Public headless two-Agent consumer: separate policy files/limits 1 and 2, submit one read/report task to each, mutate returned snapshot, restart | A stops at1, B continues at2, copied snapshot cannot mutate policy; separate persisted values reload without GUI or HTTP at startup |
| Public GUI embedding reusing Page/Connector/ArtifactScroll: two Agents and remount | Shared browser speech service, independent transient controls/queues; custom layout works without copying application |

Real-model caps per provider: at most 10 human prompts and 24 model HTTP
requests total across these paths, plus at most 3 read-only discovery/access
requests. Individual browser/public turns use limits at most4 except the zero
restoration demonstration (observed cap enforced by evidence proxy); one bounded
corrective prompt is permitted if the model declines a required tool, retaining
that attempt. No live >16 proof: deterministic local control covers limit17.
Initial selected models will be checked for access: claude-sonnet-4-6,
gpt-4.1-mini-2025-04-14, models/gemini-3.8-flash. No paid execution before local
failures are fixed and coordinator checks this matrix/frozen launch identity.

## Released predecessor and first implementation results

Coordinator released edition-2-ch07-r2 at
5b82971c3e667e08acfb2eca132307b9cdebbc51; source
9ec94ef551b08de6fbd714bf95efee901db49400, main tree
944b187319224bc2d17b833ae1c72bfd93643c2e. Teaching copies are unchanged;
predecessor-r2.json records only the support helper's build-isolation correction.
The student observed that helper commit's subject while waiting, but read no
historical answer code or diff. Root accepted the structure/lifetime plan.

First backend implementation passes inherited core and GUI tests and all three
independent black-box Chapter 8 commands: settings wire, strict validation and
actual local model policy effects (1, 17, default16, normal completion and active/
queued capture). Raw outputs are retained as initial-*.txt. The browser was not
implemented at that binary boundary, so these are not browser/live claims.
Grader noticed a newly added test fixture method copied a mutex by value; changed
its receiver to a pointer. This was implementation setup, not a teaching gap.
Two shell editing attempts initially used a repository-relative Python path while
already in the module cwd; they failed before writing, and were repeated with
absolute paths. Earlier failed compile output was observed, not a passing gate.

Additional live action required by coordinator: with applied policy1, type a
scratch-file read/report prompt, hold a typing pause before the accepted call's
admission, and apply policy2 while the active turn is held. Observe next-turn2
and active1 simultaneously. Clear only that tab's draft/pause; the accepted read
batch finishes at round_limit with exactly one HTTP request. Submit the follow-up
read/report turn, observe captured2 and its continuation. The zero/default16
restoration remains a separate action. This uses the existing total caps and
will be performed through real browser controls for each provider.

## Browser implementation difficulty and initial local closure

The first deterministic browser run passed theme/font/keyboard/pointer sizing,
viewport behavior and invalid-draft recovery, then failed resetting policy2 to0.
Observed cause: while an earlier save was still pending, an arriving applied
snapshot replaced a newly typed zero draft with2 before the next click. Chapter8
§8.2 explicitly distinguishes editor drafts from applied state, so this was an
implementation mistake, not missing teaching. SettingsPanel now tracks local
drafts and updates the applied summary independently; acknowledgment clears only
the submitted draft. The failed initial-browser-local.json is retained unchanged.
The later timestamped browser receipt passes all seven groups, including actual
DOM rendering, malicious tool-result expansion, restart with nondefaults and
controlled speech revision/rate/cancel behavior. These are fake-model and synthetic
callback tests, not real-model or audible evidence.

The independent public headless checker passes its race tests and replay checks.
Core and GUI vet/test pass; GUI race tests pass. The changed public browser-consumer
module and new headless policy-consumer module each pass vet/test. Formatting lists
no changed/new Go files. Ownership inspection corrected a GUI subscription's
concrete service back-pointer to the declared common interface before freezing.

Required historical diagnostic `make grade-dir CH=9 DIR=solutions/edition-2/main`
returns0/100 because its old CLI/GUI/protocol and structural names do not match
this second-edition contract; full original output is historical-diagnostic.txt.
No assertion was removed or weakened. The three new backend checks pass again:
6 wire groups,90 strict-validation groups,6 actual policy-effect groups. Original
and later binary-bound receipts are kept separately. Live demonstrations,
independent persistence/concurrency/browser checks and comparative review remain
open; no Chapter8 acceptance is claimed.

## Preflight self-review correction

Before any paid run, source inspection found that fast pause/control replies
could enter the outbound queue between initial snapshot records. The chapter
requires a contiguous snapshot group. Connector now defers nonsnapshot records
until snapshot_end under the same combined queue bounds; if deferred work prevents
snapshot progress, it closes for resync rather than retaining an unbounded queue
or blocking controls. Added an actual-browser raw-socket positive sending subscribe
and pause back-to-back, requiring preferences_snapshot first and snapshot_end
before the acknowledgement. The original implementation and preliminary binding
remain committed. This is an implementation correction, not a teaching ambiguity.

Independent mutation testing exhausted generated Go build-cache space. After both
coder and grader confirmed no compiler remained, the grader cleared only that
cache and recovered3.4GiB. No source, support or raw evidence was removed.

## Independent browser lifecycle finding

The grader's actual-Chrome pending-settings-disposal/remount case found a disabled
Theme control remained disabled after a new Page reused the DOM. The old callback
fence worked, but no new owner restored the transient pending-control state.
Policy-save and divider controls share that risk. The new SettingsPanel constructor
now restores these owned controls; stale callbacks still cannot alter replacement
state. The independent narrow check passes on corrected preferences.js
08be359ce21aca87c04d7b874f3e8e6633dfb4c169a8ec2af349c1fa9f233ac5.
This follows the taught Page lifetime requirement and is an implementation issue.

The first full evidence-utility preflight suite passed18 controls on immutable
858f0df, including valid-path replay, binary/source/support/dependency/launch
identity refusals before writes, real browser launch, and HTTP/prompt ceilings.
Its source binding and report are preserved under that revision's name. Because
the remount fix changes browser assets, GUI and embedding binaries will be rebuilt
and rebound before any paid run. No paid demonstration has started.

All eight independent actual-Chrome groups now pass, including both pending
preference and pending policy-save remounts. The original reviewer receipt is
book/edition-2/checkpoint-evidence/ch08-browser-remount-failed.json; the student
did not read checker internals. GUI and browser-consumer vet/test and rebuilt
executables pass after the remount correction. The live adapter now reserves
two prompts for a terminal-attached GUI, supports real pointer drags and retained
card expansion, and tests the pointer path before paid launches.

## First actual native-speech finding (before comparison)

Chapter8 §8.7 says “only A's queue and speaking cause clear” and B remains
paused until its own queue finishes or B cancels. I interpreted the inherited
BrowserApplication-owned native SpeechService as enough because each browser
tab owns its own application. The first real Anthropic two-tab autoplay run
showed a5s no-start timeout in tab0 immediately followed by tab1's native
`interrupted` callback. Chrome appears to serialize speech across those tabs;
calling native cancel from a waiting tab can affect the sounding tab. Controlled
multiple-Page callbacks did not reveal this platform interaction. Original
source bd5c05a and all raw anthropic-browser speech timestamps are preserved.
Root/reviewer notified immediately; affected speech coding and demonstrations
paused pending contract/ownership review of a same-origin Web Locks protocol.
The proposed lock belongs to BrowserApplication's native service, leaves Page
queues local, starts the no-start timer only after native ownership is granted,
and never lets a cancelled waiter call native cancel.

Separate driver error: the active-limit demonstration waited for `Tool: read_file`
when its paused card was correctly `Proposed tool: read_file`. The60s timeout
remains in raw evidence. The following actions still showed active1/next2, then
released the other tab's typing pause and completed round_limit1. No model
request was retried for that label error. Another documentation append initially
used the wrong working-directory-relative path and failed before writing.

Before changing runtime speech, I found the live adapter's inherited
`browser.newPage()` starts separate Playwright browser contexts. Such pages
do not share a storage bucket, unlike ordinary same-profile tabs. This is a
demonstration setup fault; the first observation is not yet proof of the
same-profile isolation contract failing. A fresh local native control uses
one shared browser context and existing real Gemini history, without another
model request. The reviewer conditionally accepted the Web Locks owner/lifetime
plan, while explicitly requiring this same-context evidence first.
The Web Locks specification describes same-storage-bucket cooperation and
release when the callback promise settles: https://www.w3.org/TR/web-locks/ .
That API does not coordinate unrelated origins or separate browser profiles.

The same-context native control confirms the implementation defect in
native-context-1791464657900.json: page0 starts1791464648860, page1 waits,
page0 reports interrupted1791464654899 and page1 reports no-start1791464654900.
The first local control omitted capture:true for nonbubbling diagnostic events
and produced an empty event list; it remains retained as setup failure.

Reviewer-approved correction plan, published before runtime edits: BrowserApplication
continues owning SpeechService and the browser lock manager. SpeechService owns
its pending AbortController, lease and native request. A fixed name shared across
cooperating same-origin/storage-bucket tabs (not Agent-specific) grants exclusive
native ownership. Native speak and its5s no-start timer begin only after grant.
Canceling a lock waiter only aborts its wait. A granted owner fences native
callbacks, cancels its own native request when needed, and resolves its lease
after cleanup. Stale grants/callbacks cannot advance a replaced page. Unsupported
coordination shows unavailable speech and retains no speaking pause. This provides
cooperating same-origin isolation, not control over unrelated sites/profiles.
The live driver will create one browser context and ordinary pages inside it.

Initial live checkpoint bd5c05a: all three providers were discovered with the
requested exact models available (one discovery per provider). Actual human
PTY conversations each completed read/report, followed by /usage and /history,
then a fresh read/report under persistent limit1 returning round_limit after
the final tool batch. Anthropic/OpenAI/Gemini public headless2-Agent consumers
each demonstrate request counts1/2, independent files, owned getter copies,
and reopen without another model call. Anthropic browser capture1 then next2
and settings restart pass; native speech remains a found defect, not a pass.
Exact request reconstruction matches all21 retained requests across eight
launches, including the zero-request OpenAI browser session closed during the
speech hold. Counts before correction: Anthropic6 prompts/9HTTP; OpenAI4/6;
Gemini4/6. Remaining live work stays within original10/24 caps per provider.

Repair validation caught a rapid-input fixture issue: after B received font
revision2, the fixture sent A's divider update while A still held revision1.
Captured raw frames prove an expected revision_conflict; this was not a styling
regression. The fixture now waits for A's control acknowledgment before starting
its next independent mutation, preserving all assertions and original failures.
The GUI build command also initially named nonexistent cmd/gui, corrected to
cmd/ensemble-gui after vet/tests had passed. A local native replay launcher tried
to reuse an existing log through the fresh-session CLI and was correctly refused
with cannot create fresh event log. That setup failure made no model call.

The reviewer withdrew the pre-code manuscript-publication hold as an unnecessary
workflow interpretation: the existing contract and approved plan were sufficient.
The Web Locks repair is implemented, keeping BrowserApplication/SpeechService
ownership and Page queues. Independent eight browser groups pass; the new
independent shared-context Web Locks control passes waiter cancellation, lease
release/next grant, stale callback fencing and missing-coordination pause cleanup.
Actual native component control native-context-1791465187711.json shows A still
speaking after8seconds, B waiting without starting, and B cancellation causing
no A interruption. No synthesis was replaced in that control. The final full
student browser fixture passes all8 groups after explicit initiating-client
acknowledgment waits; both intermediate fixture failures remain retained.

## Corrected actual browser phase

Frozen cd9de3e passes18 evidence controls and all50 independent deterministic
groups after reviewer fixture corrections (no runtime change from that gate).
All three browser/provider runs preserve real settings, task, policy, hint,
reconnect and native speech/audio receipts. Three8-second native captures have
zero silent-prefix RMS and speech RMS about0.092, with1.6 rate events. No native
speech error occurred after the lease repair. Anthropic old revision6/rate1.2
utterances continued after revision8 disabled autoplay; cancellation of A left
B's speech and typing intact. OpenAI/Gemini queues had finished before their
disable action, so that specific overlap remains open for a retained-real-text
replay/native supplement without more model work.

Anthropic honored the one-request hint on its next model request by writing
port=9090, but a later request, after the hint was consumed, rewrote the file to
the original marker-only instruction. Raw bodies show delivery exactly once.
This is observed provider behavior, not proof that the final correction persisted.
OpenAI and Gemini retained the requested marker plus port9090 in their scratch
files. The notes input remained unchanged.

The inherited public embedding example deliberately opts into no tools. I failed
to account for that in two Anthropic embedding prompts; both providers' requested
capabilities were correctly absent and Anthropic declined. The two-Page model
delivery/remount worked, but the hoped-for tool interruption did not occur.
Reviewer approved exactly one additional Anthropic text-only prompt/HTTP, revising
its cap to11 prompts while retaining24HTTP. Original cap10 and both refusals
remain preserved. A fresh source-bound consumer launch avoids rewriting the
original launch identity. No consumer tool configuration changes are justified.
OpenAI's nonempty provisional embedded answer was successfully interrupted.
Gemini's512-token embedded response had already ended when clicked; that too-late
outcome is retained. Its existing tenth-prompt allowance then held a real read
proposal behind another tab's typing pause and interrupted the active turn before
execution, with an error tool report and interrupted completion. Exactly one
model request was used for that corrective path.

The cd9de3e verifier exactly reconstructs24 requests from six completed browser/
embedding launches plus one from Gemini's bounded interruption correction.
Current counters before the final Anthropic correction: Anthropic10/17, OpenAI9/14,
Gemini10/15 (prompts/modelHTTP). One discovery request per provider remains separate.

## Historical comparison: lossless browser revisions

Reviewer found that the taught uint64 disk/wire revision contract exceeds JS
Number's exact integer range. A valid file seeded9007199254740993 loads, but
JSON.parse previously rounded the browser base to9007199254740992, making both
settings domains permanently conflict. A9007199254740992 transition reproduces
the loss after a successful commit too. This is a teaching representation gap
and an implementation defect, not a reason to silently narrow persisted input.
Before affected code, reviewer approved preserving numeric uint64 disk/wire,
retaining unsafe revisions as BigInt internally, exact successors/comparisons,
and a dedicated unquoted base_revision encoder. Only known protocol-counter
positions are normalized; user strings and tool argument objects are untouched.
JSON.parse reviver source-context supplies the original lexeme; a browser without
that support fails visibly before settings-ready or sending a rounded base.
The actual ECMAScript specification was consulted (no historical answer code):
https://tc39.es/ecma262/multipage/structured-data.html#sec-json.parse .
This repair groups with final evidence adapter work before another launch.

The independent maximum-revision browser check then exposed a second loss on
server projection: typed observation revision18446744073709551615 was marshaled
and decoded through float64 before wire encoding, producing18446744073709552000.
The browser correctly refused that out-of-range token. All three typed projection
conversions now retain json.Number tokens while preserving opaque sanitization
and typed-child traversal. This is an implementation correction within the same
approved representation plan, not a new persisted range restriction.

## Final repaired runs and evidence reconciliation

At a06d4f3 the core and GUI vet/tests, both public consumer module vet/tests,
lossless local controls and actual-browser eight-group compatibility pass.
Independent actual-Chrome numeric scenarios pass seed7, safe-to-unsafe transition,
unsafe9007199254740993, max-1 to max and max exhaustion, including consecutive edits,
live applied display, reconnect and exact conflict-current retry. The independent
immutable affected subset is pending reviewer completion; the earlier full50-group
cd9de3e result retains its original source identity.

The expanded20 local evidence controls passed before the final live launch,
including native browser adapter startup, complete source/binary/support/dependency
identity refusals, actual nonempty SSE provisional text then interruption, and
public replay with concurrent settings controls. Full binding covers106 runtime
files,10 executables and8 support files. An initial retained-log lookup correctly
refused before creating a run because Git ignored .log files. The unchanged originals
were still present and had already passed request reconstruction; explicitly
committing them at0aad851 corrected the evidence omission. No original was replaced.

The one approved Anthropic correction used a fresh embedded consumer, ordinary
text-only prompt in its second Page, observed provisional text "The", then accepted
interruption and an incomplete answer with Outcome: interrupted. The idle first
Page closed/remounted without affecting the second. Exactly1 real HTTP was sent
and reconstructed. Final totals are Anthropic11prompts/18HTTP, OpenAI9/14,
Gemini10/15; discovery1 each remains separate. No paid repeats for numeric repair.
The two original Anthropic tool refusals and original cap10 exhaustion remain.

OpenAI and Gemini supplemental speech checks replay their retained real-provider
neutral events through a public consumer whose provider endpoint is disabled.
They are not fresh model generations. Both actual Chrome runs show applied
autoplay-off revision4/rate1.6 while two Pages still hold speech from revision2/
rate1.2. Canceling PageA leaves PageB speech plus typing; canceling PageB leaves
only its typing; clearing its draft clears that last cause. Native start events
confirm both original-rate utterances; subsequent manual playback uses1.6.
Actual native PCM captures have zero RMS during the1.8-second silent prefix and
speech RMS approximately0.047, with separate timestamps/hashes in audio-analysis-
a06d4f3.json. Original paid-run queues had finished before off-toggle; only these
explicitly labelled supplements establish that overlap for those providers.

The replay launcher initially demanded exact top-level event times, but public
Agent.Append intentionally assigns a new admission time. The two final assertions
failed and remain recorded. A separate identity-first comparison proves exact
count/order/seq and every other field, while recording all changed top-level times.
Both original and re-admitted logs remain. This is neutral-event replay with
rewritten admission times, not byte-identical original logs. The helper is being
corrected and locally tested; no runtime repair or provider call is warranted.

Actual two-tab settings controls produced correlated settings_busy errors. A
deliberate fresh edit applied actions width421 while preserving the competing
font19 commit. Three fresh production GUI launches with one shared persistent
workspace then established positive reload (light theme, autoplay true, policy2),
explicit false/zero update, and second reload retaining autoplayfalse/policy0,
with exact applied revisions and fresh empty history. All three launches sent0
model HTTP requests. Persisted files are copied per stage, so later intentional
workspace changes do not erase the earlier positive evidence.

Independent historical comparison otherwise supports the existing owner structure,
checked application after replacement, immutable turn capture and component reuse.
It found no benefit in introducing a common persistence abstraction solely to
remove modest duplication. Added comments explain numeric precision and why
strict-parser service parameters intentionally preserve actual-owner diagnostics.
No historical answer code, reviewer implementation or future teaching was read.
Root still owns author reconciliation, final ledger/export/tag; these student
receipts do not claim those coordinator acceptance steps are complete.

The support-only correction52c9561 now compares neutral facts with only the
documented top-level admission-time exception. All21 local evidence controls pass,
including five exact refusals for changed count/order/seq/payload/missing time.
A fresh bound endpoint-disabled replay-helper-final-control run exits0 and retains
38 exact non-time events with38 rewritten admission timestamps. Runtime source
hashes are identical to a06d4f3; old paid/native receipts retain their a06d4f3
binding in binding-a06d4f3.json. No further provider calls occurred. All six final
browser runs have zero action_failed/pageerror records; final-live-summary.json
records action counts and the actual per-provider counters.
