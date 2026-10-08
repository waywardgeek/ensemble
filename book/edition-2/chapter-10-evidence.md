# Chapter 10 preparation evidence

Author research, October 8, 2026. New Chapter 10 maps to first-edition Chapter
11 under the current procedure, workflow and global-review map. This record
supports [the outline](chapter-10-outline.md), not a completed implementation,
student handoff or actual-use claim. Chapter 9's contract is accepted for
independent checker preparation at `0bca41c`; its implementation is not yet an
accepted predecessor. Exact Chapter 10 contract choices remain under review.

## Role and source boundary

The replacement author on `/root/coder_ch04` previously implemented second-edition
Chapter 4 as its student, then authored Chapter 7 reconciliation and the Chapter
9 contract. This thread has no Chapter 7–10 student implementation authorship.
Historical chapter and grader exposure below is deliberate author research;
this context is unsuitable for a cold Chapter 10 student. No first-edition
solution implementation was copied or edited. No runtime or grader was changed,
and no persistence grader, reference solution or paid session was run during
this preparation.

The Chapter 7 feedback count correction is separately committed at `2112af1`:
“all recorded items” replaces the incorrect count after the initial-live table.
Its independent proofreader was notified. It changes no Chapter 7 receipt.

## Actual read ledger

Full current `book/voice.md`, `book/chapter-writing-procedure.md` and second-edition
`architecture.md` were read and reloaded after compaction. The mapping/workflow
opening and global review's architecture and forward-lesson sections were read;
the complete forward-lesson table was reread for this task. That is not a claim
of rereading every global-review section.

The author retained the complete accepted Chapter 9 contract from drafting and
review, including `0bca41c`'s transition-validation correction. Chapter 8's complete
contract was read during the preceding Chapter 9 preparation; policy persistence
and historical turn-policy rules were reread here. Chapter 2 event sequencing,
immutable log destination, partial-write refusal, CLI/log configuration and
rendering rules were consulted. Chapter 5 shutdown/ownership passages were
located for the existing lifecycle boundary; this preparation does not claim a
new complete Chapter 5 reading.

Historical reads were:

- Complete first-edition Chapter 11, including its eight rules, seven-check table,
  default-load exchange and demonstration instructions.
- Complete `internal/grade/ch11_checks.go` and `ch11_harness.go`, in chunks. These
  are fixture-source observations, not executed results.
- Old Chapter 15's persistence contract, semantic-error recovery text, and later
  artifact-handle/journal discussion around lines 650–704.
- Old Chapter 16's recorded compactor launch/recovery and edited-memory identity
  discussion around lines 368–393 and 482–505.
- Old Chapter 19's credential-provider ownership seam around lines 85–102.
- Old Chapter 20's SIGINT/save and replay/part-identity accounts around lines
  94–110 and 241–294.
- Path-specific Git history for old Chapter 11 and its grader, and relevant old
  Chapter 15/20 history. The full commit messages listed below were read.

Searches across later chapters located these passages; they do not constitute
complete readings of those chapters. Historical credential discussion supplies
an ownership lesson only. This preparation makes no current OAuth, endpoint,
pricing or provider-policy claim requiring fresh external verification.

## Historical contract and changes

| Receipt | Finding retained for teaching |
|---|---|
| Old Chapter 11 §11.6 | The chapter attributes “Let's load by default without a flag.” to Bill. The default-startup lesson and unreadable-file overwrite danger are preserved; no new dialogue or private experience is invented. |
| `f71d71d274634aeaa060c9fd2d3f457aaa5de76f` | Initial persistence introduced save/load/rebuild and sequence reset. Its behavior in common and early load-flag design are historical choices, superseded by current ownership and automatic resume teaching. |
| `5a6cfbc53e30b82b90e3fe8bc3193c61875a138d` | Added a verification command. A self-reported MATCH was later removed as acceptance evidence. |
| `c24b8e84da112ef66f729c0cdc60413ee6fb547d` | Reworked checks around externally observed next requests, a real old-snapshot/new-log tail, default load and isolated workspaces. Its historical scores are commit-message claims, not rerun results here. |
| `1d2acd64c9c5d203f61ef7f9de6a1e982f37a531` | Made the complete-log precondition explicit and removed an unmeasured timing claim. Snapshot-only imports cannot promise a full rebuild from absent history. |
| `459dca538098e0a5f7632907e0d3e1bcb7fb44c5` | Reconciled the body with the anchored contract and the diagnostic-only role of verify. |

The old contract writes one JSON save with config, semantic context, anchor and
log. It loads by default, treats missing as fresh and malformed as fatal, applies
only events after the anchor, and supports a snapshot with no log. Its config
records model/vendor/system/tools without restoring credentials. A complete log
must produce the same next request whether loaded with its snapshot or rebuilt.

Those are useful behavioral starting points. The new edition already owns an
append log and Actor, immutable live log identity, explicit request/call state,
producing-model usage, historical skill material and separate settings domains.
It cannot adopt the old “whole agent on disk” comment literally. Nor can “current
config wins” silently replace a primary skill or resurrect tool authority.

## Later lessons moved forward

| Historical source | Earlier safeguard proposed here |
|---|---|
| Old Chapter 15 artifact-handle collision account | Preserve Chapter 4 exclusive file creation and occupied-handle skipping; a resumed allocator never overwrites old output or turns old handles into process rights. |
| Old Chapter 15 journal/anchor discussion | Preserve one durable append path. A checkpoint names its boundary, and later durable events remain a tail. Do not add a competing journal or claim fsync/power-loss properties from a process-crash test. |
| Old Chapter 15 parsed-event skip policy | Do not import it. New event/skill transition validators refuse impossible state; corruption remains visible and original bytes remain untouched. |
| Old Chapter 16 recorded work launch | A recorded launch is history. Replay cannot start background work or make a model call. The first persistence chapter refuses unfinished live continuation instead of adding a compaction recovery protocol. |
| Old Chapter 16 same-file identity after content edit | Path identity alone cannot establish immutable skill-catalog compatibility. Specify content/graph/grant/binding inputs; historical rendering uses recorded material. |
| Old Chapter 19 credential-provider seam | Credentials and transport remain current application inputs, outside persisted semantic context and renderers. No credential migration or billing fallback is added. |
| Old Chapter 20 SIGINT/save account | Signals must enter coordinated application shutdown; save code after an abandoned input loop is insufficient. Combined GUI terminal detachment retains its distinct lifecycle. |
| Old Chapter 20 missing replay part identity | Snapshot correctness includes public card/request/part identities, not only visible words. Test independent reconstruction and reconnect. |

These are historical prose/code-history findings, not newly reproduced bugs in
the accepted second-edition source. The outline omits the old event-count and
timing figures because no corresponding raw run was independently checked here.
It also declines to copy the old hint/prompt classification into today's explicit
actor requests.

## Inherited grader strengths and gaps

The published first-edition table contains seven checks totaling 100 points:
save-shape 15, default-load 20, replay-equals-snapshot 20, tail-applied-once 15,
log-not-needed 10, bad-save-refused 5, and Chapter 10 parity 15. The six persistence
fixtures and result/check mapping were read; parity's broader orchestration was
not independently audited in this preparation.

Strong controls worth retaining include whole-request byte comparison between
snapshot and full replay, an independently spliced nonempty tail, two process
starts in one case's workspace, and a malformed file whose bytes must survive
startup refusal. The harness isolates independent cases in fresh workspaces,
which matters once default startup reads persistent state.

Limits are concrete:

- Save-shape checks selected fields and increasing sequences. It does not establish
  a complete semantic snapshot, strict versioning, secret exclusion or authority.
- Snapshot-with-empty-log checks continuation above its anchor and retained prompt
  content; it does not prove all identity allocators or producing-model usage.
- Bad-save-refused uses invalid JSON. Add structurally valid corruption, impossible
  transitions, mismatched identity/anchor, duplicate fields and explicit zero-call,
  zero-write assertions.
- The fake backend covers Messages. Add exact projection comparisons on all three
  supported API paths, including mixed parts and Chapter 9 material.
- Old stdin protocol/startup options and EOF-driven GUI exit do not establish the
  current human PTY or optional GUI --terminal detachment contracts.
- Atomic write faults, public multi-Agent isolation, two live writers, signal
  shutdown, stale job handles, pending-next-call settings, catalog compatibility,
  independent Chapter 8 policy/preferences and credential canaries need their own
  distinguishing controls.

`make grade-dir CH=11 DIR=solutions/edition-2/main` is a historical diagnostic,
not a complete new acceptance command. No shared grader change is made here.
The independent engineer must establish the legacy baseline before any repair,
retain passing controls and show intended failures under targeted mutations.
The full new contract must precede that checker work and student handoff.

## Coordinator direction and remaining choices

The coordinator accepts Agent-owned SessionStore/DataDir with an actual parent,
one durable append path and preserved explicit LogPath semantics as the outline's
working direction. The first persistence chapter should refuse live resume at an
unfinished turn/call boundary while permitting honest offline inspection. Even
a completed turn can contain a formerly running job: its saved status confers
no live process ownership or usable stale handle.

Skill-mode resume must compare newly caller-selected catalog, primary, scalar
bindings and installed grants with recorded identity. Same-content, missing and
changed-source cases need explicit behavior; historical rendering remains file
independent. Current route/credentials and Chapter 8 policy stay separate.
These are coordinator working choices, not new rulings attributed to Bill.

The outline lists six decisions for the complete review: concrete store/flag and
LogPath compatibility; settled capture/resume boundaries; snapshot-only validation
versus available-prefix equivalence; pending tool_limits restoration; canonical
skill identity and no-skills system precedence; and cross-process writer/resource
limits. It recommends retained Chapter 2 truncated-record refusal. No silent
salvage or broad crash-work recovery is assumed.

## Demonstration and review boundary

The proposed live matrix requires all three supported providers through actual
human CLI restart, optional browser restart/reconnect and public multi-Agent
consumers. It must show retained dialogue and compatible skill state without
repeating a prior effect, preserve policy/preferences ownership, and bind both
processes to source/binary/store/catalog identities. The student observes each
answer before follow-up and retains failures rather than fabricating successful
recall. The live budget and exact feature/action matrix remain to be written
against the accepted contract and implementation.

Corruption, storage failures, stale-handle attempts and crash boundaries require
separately labeled deterministic controls. This file contains no actual-spin
transcript, proposed text masquerading as output, or claim that Bill tested it.
Preparation ends at outline review; predecessor acceptance, independent checks,
fresh student use and later comparison/proofreading remain ahead.

The scoped author check was
`go run ./cmd/lintprose book/edition-2/chapter-10-outline.md book/edition-2/chapter-10-evidence.md`.
It passed all hard rules. Soft warnings concern the chapter-length floor and
negation density: these are preparation records, and the qualifications
distinguish historical evidence, proposed rules and work still ahead. A cut and
paragraph-ending pass retained those distinctions and checked the human stake
against old §11.6. The existing chapter's literal fixtures were not edited.

## Source fingerprints at preparation

SHA-256 values identify the bytes consulted, including files only read in part.
They do not claim complete readings beyond the ledger above.

```text
17883353cf9653c0df46081c216b5b1623dfa640619e06fe89bff46fc29dfad1 book/voice.md
52a2083fb708a112c47d4bf34a2d7215957f80d11e9559818a66c26e82e37649 book/chapter-writing-procedure.md
4d40b6ce0d223b286b4b5000413fb5a2f7ad40c7ff8e6848e81740bcbf6c33de book/edition-2/architecture.md
b44ee30f1cb5d43a971b06473468e3a175b0890600d54e0a902a7e59d8903384 book/edition-2/workflow.md
15984b674642879880282695a926dee7636c8396793e6ad69d2a43c24fda9b9e book/edition-2/global-review.md
698533c7881f89c5238b2cfa1a72f64b7e3869d1c4fc5f6c646da072cb982c37 book/edition-2/chapter-02.md
6b705930bff324c8930446e1993e2a750cd30801b38799c124443d1da8ba9347 book/edition-2/chapter-08.md
23ec0b75a0a948beaa8a1a1604969f2b2b040cfa01250c998e70ad48c79d4072 book/edition-2/chapter-09.md
ed6482dd1e5247880122382a3b4199276d8fbda13ea8863676eda5c6e49e9df0 book/chapter-11.md
6c6d953e754f73b8bafd45a2f6009cab5dce9d7557c008e79fcc4084f27c2e79 book/chapter-15.md
4e8bb1ce58f73aa64b71b5fd57e531460e64d80ca6b2f1e934a1e222deb77791 book/chapter-16.md
41fe6b803bd837c0ec3f4ed818cf02187c588c63b1ebee4713484d9fa404161c book/chapter-19.md
9a6d41326806d229d1dc7913297d7b5f2ef04ab861e1068f67d393625e9700ac book/chapter-20.md
57681a1dc938f86cf89526073a493fa92d1acd9f9054cc59760f827c9accc079 internal/grade/ch11_checks.go
b834e66359c579340fb505a48b71f7ac99422ef94494a1034fa93fb9fc3fbe40 internal/grade/ch11_harness.go
```

## Full contract draft after outline review

The coordinator read the complete `4744f41` outline/evidence and authorized the
full draft on October 8. The initial preparation and its open-question chronology
above remain intact. The following are coordinator working choices, explicitly
separate from Bill's rulings:

- Preserve fresh NewAgent and offline Load; add explicit public session opening,
  used by human CLI/GUI by default, with workspace-local `.ensemble/session` and
  a shared session-directory selector. An explicit legacy log is never silently
  redirected or migrated.
- Capture at idle/settled actor boundaries with immediate busy refusal and one
  off-actor worker. Refuse unfinished live resume and a partial final record.
- Support semantic snapshot-only import with allocator state. Validate equality
  against the complete available prefix, with replay cost stated honestly.
- Preserve Jobs-owned pending tool_limits with durable set/consume facts through
  the existing writer, including failed literal-next-attempt consumption.
- Compare exact semantic catalog, bindings and installed ceiling. A new path with
  identical content can resume; current credentials/route/model and Chapter 8
  policy remain separate. Plain-mode base System is session identity: omission
  adopts it and a competing explicit value refuses.
- Hold a nonblocking OS lock for live store ownership and publish resource bounds.
  Persist SessionID separately from the current Ensemble's runtime Agent.ID.

A follow-up design question was answered before drafting the affected format:
the outer checkpoint format and semantic invariants must be exact, while private
state-payload names may follow the student's data structures. The student must
publish a strict versioned codec. Independent acceptance uses its public
snapshot/import surface, documented format and externally visible behavior,
rather than an unpublished private-field assumption. Shared values/interfaces
remain in common; the snapshot is no second mutable conversation authority.

The coordinator also accepted initialization/anchor facts as a reasonable proposed
boundary, requiring exact placement, counter accounting, snapshot binding and
construction-only admission. `chapter-10.md` supplies those details and remains
under full contract review. It is not a student-ready contract yet.

### Additional draft choices needing coordinator review

The author surfaced these details while writing, instead of presenting them as
already settled user decisions:

1. **Transient request ordinals.** Chapter 5 queued cancellations have no durable
   turn event. The draft adds request_index only to a session's actual turn-start
   fact; snapshot cursor can retain higher burned ordinals. Prefix comparison
   checks its lower bound and recorded-ID uniqueness instead of requiring exact
   equality to invisible dead handles. This preserves nonblocking admission
   without introducing an enqueue journal. A mount nonce was presented as an
   alternative; the current draft selects explicit ordinal/no-reuse rules.
2. **Imported origin lifetime.** Later checkpoint replacement cannot erase the
   seed named by the first anchor. The draft preserves immutable origin.json,
   binds the first session_anchor to its exact bytes and all high-watermarks, and
   rejects incomplete imports. It remains one events.log writer. Replay from
   origin makes no claim to reconstruct absent pre-origin history.
3. **Plain-mode tools.** The draft also binds installed/visible handler definitions
   in plain mode. A changed set refuses compatibility rather than adding a
   permission-migration/fork operation. Skill mode already requires exact ceiling
   compatibility under the coordinator decision.
4. **Skill transition proof.** Active roots alone cannot prove earlier load
   visibility after an advertiser retires. Snapshot state must retain compact
   ordered transition provenance referencing immutable activation records once,
   so Chapter 9 action/visibility/retirement validation survives snapshot-only
   import without embedding the complete event log or duplicating primary text.

These are concrete review points in a complete draft. No runtime implementation
was instructed to follow them before that review. The accepted predecessor and
independent published checker still gate the fresh student.

### Scope, cost and source follow-up

The draft deliberately excludes automatic crash-work recovery, log truncation,
fork/migration commands, cross-domain settings transactions and distributed locks.
It states the extra prefix replay/capture cost, requires indexed validation and
streaming bounded reads, and distinguishes durable semantic equality from runtime
Agent/watch/worker state. Its filename table refuses missing or inconsistent
origins rather than assuming another valid-looking file wins.

The full current Chapter 9 skills transitions/material and Chapter 5 request,
turn and shutdown rules were reread in targeted passages. Chapter 4 exact limit
consumption, error-before-admission and occupied-artifact behavior were consulted;
Chapter 7 safe watch/wire placement and Chapter 8 independent file lifetimes were
checked. No runtime source was mined to decide the new contract. A targeted read
of `cmd/lintprose/main.go` explained a soft warning: its person detector matches
Bill/I/my/me, so literal reader/user stakes still require the manual reading pass.
No linter source was edited.

For the new local-lock statement, primary documentation was checked on October 8:
[Linux flock(2)](https://man7.org/linux/man-pages/man2/flock.2.html) and
[Apple's archived flock(2)](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/flock.2.html).
The draft confines support to macOS/Linux local filesystems, requires a
noninherited descriptor and retains the lock file across release. It makes no
claim about hostile replacements, arbitrary network filesystem semantics or
power-loss survival. These are documentation checks, not executed lock tests.

Bill is concurrently writing first-edition sandboxing material with CodeRhapsody.
That work is untouched and recorded only as a pending future lesson source.
This preparation neither claims to have read a completed source nor assumes a
new chapter count or a security result from it.

The chapter's demonstration section remains visibly a proposed reproduction
plan. No CLI output, browser screenshot, provider success, save timing or measured
usage has been invented. Independent review, implementation and actual use will
supply those receipts later.

Snapshot-only live compatibility has one further explicit distinction in §10.7:
recorded material is installed unchanged, while validation compares its grants,
graph and expected expansion at the original candidate boundaries against the
caller's matching frozen catalog. Offline inspection/rendering still requires no
catalog or expansion. A mismatch is refused rather than replaced. This prevents
a forged active snapshot from borrowing a valid catalog digest while supplying
different grants or primary bytes; it is included in the coordinator's detailed
contract review, not claimed as an implemented protection.

Additional consulted-source fingerprints:

```text
b18b22663ceb4a9bbe74663a830423b0762afaee783e5f92b706c483d5289a46 book/edition-2/chapter-04.md
1635746aed55332bb82a26df93158f5c5b2a1a6f7bf684fd25fcb0a64b656fb8 book/edition-2/chapter-05.md
b3a567cdec86a334a44d248f405dc316c8f3ce1a97b6b9d01dbf9dcb5d275473 book/edition-2/chapter-07.md
794d5283a3b89b37a40a0a14c26a46a534dda92f4d9457781330ae63543f6954 cmd/lintprose/main.go
```

The complete chapter received a cut/last-sentence/figures pass and scoped prose
lint. All hard rules pass; the remaining chapter soft warning is the narrow
person-name heuristic described above. The manual pass retains the returning
reader, catalog relocation and useful refusal stakes in the later mechanism.
The evidence file's soft length/qualification warnings remain appropriate to an
audit record. No old chapter, literal prior fixture or grader was edited, and
no implementation tests or paid repeats were represented as run for this prose.

## Coordinator review and numeric/allocator clarification

The coordinator's [complete draft review](chapter-10-review.md) accepts all
four proposed decisions from `9181683` as working choices: durable request
indices with a lower-bound cursor comparison; immutable snapshot-only origin;
plain-mode handler definition compatibility; and compact chronological skill
transition provenance. They are coordinator choices, not new rulings attributed
to Bill. Chapter 9 remains an unbuilt predecessor, and no Chapter 10 student or
live spin is released.

The requested narrow followups now distinguish the session's greatest historical
job handle from Ensemble's shared allocator cursor. Other Agents and occupied
artifact names burn values outside this session. Resume raises the live floor
without lowering it; prefix equivalence compares the session's reduced durable
maximum, including events beyond the GUI window.

The first draft's binary64 canonicalization for noninteger values could collapse
distinct handler definitions even though identity integers were preserved.
Section 10.3 now defines an exact decimal coefficient/exponent normalization
for every accepted JSON number, with literal integral-decimal, exponent, unsafe
integer and negative-zero examples. It never expands large exponents. Canonical
hash bytes are distinct from original stored/log/manual/opaque bytes; unchanged
files are not rewritten merely to compute compatibility.

The first draft also imposed a JavaScript-safe maximum on identity integers.
The coordinator explicitly withdrew that restriction after the Chapter 8
precision lesson and Chapter 9 uint64 activation contract. Section 10.8 now
preserves valid predecessor uint64 values with positive/nonnegative distinctions,
exact public/browser representation and pre-mutation overflow refusal. Resource
size bounds remain unchanged. This is a compatibility clarification before
checker work, not an undocumented rejection of valid predecessor state. The CLI
build now names its output to avoid collision with the existing cmd directory.

## Framing-byte consistency correction, October 8

Chapter 9's integration review exposed a missing independent encoded skill-record
bound. Its new 67,108,864-byte cap includes a framing LF when present. During
that decision the coordinator described Chapter 10 as already using the same
boundary; the author checked §10.8 and found its table explicitly excluded LF.
The coordinator then chose to correct Chapter 10 to the inclusive boundary.
This is a new consistency correction, not a claim about the earlier draft.
Whitespace, escaping and all payload bytes count; any otherwise permitted record
without LF counts its actual bytes. Session parsing keeps its existing refusal
of a partial final record. No student implementation, grader or old receipt was
changed, and the current Chapter 9 student was not exposed to this later chapter.

## Initial Chapter 10 student questions

The author read the complete 384-line initial plan at `7cb84290af3e8bac5359339f8ab5d4e7e941bf74`
and its exact Q1–Q3 before affected implementation. Current voice/procedure and
architecture remained loaded from the author pass; relevant Chapters 9/10
record-size, origin and canonical-equality passages were reread. The coordinator
accepted the student's proposed distinctions, including a bounded first-record
probe for generic offline readers. No owner/API plan approval is implied by
these author responses.

The teaching now makes session event-size precedence explicit in §10.8, raw
history versus imported semantic state explicit in §10.4, and original replay
bytes versus purpose-specific canonical comparison explicit in §10.3. The new
direct feedback file answers each question without exposing future chapters,
historical implementation or reviewer internals. It asks the student to confirm
resolution of the substantive questions; no runtime or receipt is invented.

The bounded probe cannot turn an invalid/interior initializer into a mode switch.
Skills retains controlled preflight failure before session storage admission.
The genuine saved watch window is not fabricated raw history. Dedicated validated
raw-JSON strings are allowed as a private codec choice, not a required field name.
The chapter's later checker paragraph also now points to the already-published
initial command while retaining the complete required matrix.

The plan bytes were matched to `7cb84290`. The retained prose checker passes
hard rules at 7,289 chapter words; the existing dense-contract person-gap warning
was reviewed. The original JSON fixture is byte-unchanged and parses. Scoped diff
validation and manual reading of the amended passages pass. No runtime, grader,
model call or accepted historical receipt changed.

## Unicode clarification before affected grading

October 8, 2026. The coordinator/grader identified a lexical-validation gap in
session admission: raw ASCII JSON containing a lone surrogate escape passes a
UTF-8 byte check but Go's decoder replaces the invalid escape with U+FFFD.
The author read the complete local primary documentation using
`go doc encoding/json.Unmarshal`, including its explicit replacement behavior.
This is a coordinator/grader finding, not a new student question or runtime
reproduction. The coordinator chose strict Unicode-scalar validation for session
strings/keys and canonicalized JSON before replacement. Chapter 10 §10.3/§10.8
now print lone-high/lone-low refusals, paired/U+FFFD acceptance and the valid
escaped-backslash distinction; replay-bearing raw fields preserve original valid
bytes after validation. Legacy standalone decoding retains its earlier contract.
A separate direct-feedback addendum preserves that attribution. No code, grader,
provider, snapshot or legacy edit occurred.

## Accepted-byte boundary from initial prefix diagnostic

October 8, 2026. The author read the full retained diagnostic in
checkpoint-evidence/ch10-initial-prefix-diagnostic.json and its grader-review
explanation at `63dea39`. This is the reviewer's reproduction of the preserved
initial binary, not an author execution or a claim about evolving current source.
One local HTTP request created the store; unchanged inspection refused at record
7. A checkpoint-free separate copy rebuilt. The two differing fields were live
context and saved-window raw_usage, with spaced provider JSON versus compact
recorded JSON. Original files remained unchanged; no real provider was contacted.

Focused teaching reads checked Chapter 2's physical JSONL/typed Part/provenance
contract, Chapter 6's retained JSON/merged final usage and framing distinctions,
and Chapter 9's exact material and original-read/emitted-write bounds. They did
not specify a pre-acceptance representation boundary capable of reconciling legal
raw formatting LF with a one-line event. The author proposed bounded final-record
preparation; the coordinator selected it before affected implementation.

Chapter 10 §10.3 now makes the final prepared event encoding authoritative for
new accepted session raw fragments, with validation first, number lexemes and
decoded values/order/opaque semantics preserved, and append before state/observe.
Imported fragments remain exact and are never normalized on load. Controlled
Skills preflight still precedes session storage admission. This resolves the
observed integration gap without relaxing all byte comparisons or adding a second
raw archive. Direct feedback identifies the coordinator/grader finding separately
from original student Q1–Q3; no earlier chapter, frozen snapshot, runtime, grader,
legacy implementation or policy file was edited.

Scoped hard prose lint passes at 7,681 words. The soft length warning reflects
the added admission contract; the paragraph resolves a concrete observed failure
rather than padding. Existing JSON fixture bytes are unchanged and parse; scoped
diff validation is clean. These checks validate prose/literals only, not a runtime
repair. Independent narrow clarification review remains required.
