# Stage D: source-relative chapter enhancement plan

Author: `/root/compaction_chapter_author`, October 14, 2026. Plan for coordinator
and independent review; no chapter prose has been edited. This is maintenance
from accepted implementation, not a blind student run. Runtime source is
`7feb910a`; independent acceptance is `09e68149`. Documentation-only merge
`a322050b` does not change that source. Current chapter numbers remain in use.

**Through-line stake:** the person working with a long-running Agent needs the
next task to proceed while yesterday's reasons become memory, without losing
those reasons to a failed worker, stale result, or interrupted save. The memory
chapter resolves the mechanical part at safe commitment and recovery, then keeps
the existing live failures as the test of whether the resulting memory is useful.

## Reading and scope

Read completely before this plan: the exact mandatory
[`.agents/skills/ensemble-coding/SKILL.md`](../../.agents/skills/ensemble-coding/SKILL.md),
workflow, carryover, voice and chapter-writing procedure; current Chapters
14, 15, 16, 22 and 23; the complete handoff, structured-submission record and
background-compaction record; progress record; dependency audit's mapping,
material dependencies and change map; runtime audit's Stage B, Stage C design
and final Stage C dispositions. Truncated workflow ranges were recovered. The
fresh workflow opening was read after the coordinator's accepted-C status edit.
All four governing/craft documents remain in active context and must be reloaded
in full with the skill after any compaction before further authorship.

Also inspected actual actor compaction, durable submission and staged-memory
boundaries, plus current17's auxiliary accounting/cache-boundary passages and
current20's corresponding status passages. The accepted design and independent
review supply the broader source/evidence map; this author has not rerun runtime
tests, graders, or live sessions.

Keep the existing chapter bodies and useful TL;DR explanations wherever the
accepted behavior still agrees. No length target or general rewrite. No changes
to preface, Chapter25, frozen exports, Edition1 or generated publication. Current14
is the invariant reference; its final forward pointer changes only at Stage E.

## Chapter 15: Forgetting on Purpose

| Current location | Bounded enhancement and reader-facing reason |
| --- | --- |
| Opening and §15.1 | Retain Bill's family-assistant/workplace memory-bug story verbatim. Add the immediate practical distinction: a checkpoint selects completed work; its child works while the owner continues, and source stays available until replacement is accepted. Keep removal versus preservation and the five-band explanation. |
| TL;DR | Grader coder owns final wording, shared records and all twelve requirements. Replace direct synchronous compression/turn-owned interruption deliberately; preserve the eleven weighted obligations. The author's prose must explain each changed behavior before its detailed exercise. |
| §15.2 One Door | Keep the save_memory/handoff_task split and April18 gap. Replace the paragraphs starting “The Agent runs this work” with one owner-lifetime attempt, whole selected segments, new dialogue outside the selected cut, normal finish/interrupt preservation, and a safe idle/pre-render acceptance boundary. No repeated handoff may overlap an active/awaiting-commit attempt. |
| §15.3 The Five Bands | Preserve band order, Seq0, date/number identity and active/retired filenames. Explain briefly that hidden staged bytes are outside the active corpus. Keep MemoryFileID independent of event Seq. |
| §15.4 Two Events | Preserve the paired retirement/population mechanism and why events carry actual text. Extend launch correlation and terminal outcomes; distinguish submitted candidate, committed projection and disk finalization. Introduce a compact ordering explanation before the recovery details in §15.6. |
| §15.5 The Compressor | Replace the direct-call/not-a-sub-agent argument explicitly with a real restricted child. Start with the observed role conflict and blocked main work, then explain the fresh purpose/source-owner voice and exact submit(memory) object. Teach dedicated instructions, source as role-labelled data, ordinary visible explanation before submit where required, declaration plus dispatch enforcement, no tools beyond submit, no recall/memory/skill/MCP recursion. |
| §15.5 execution details | Explain the finite three-Send/three-batch lifetime, prose-only failure, whole-batch refusal, correction opportunities and no automatic retry after failure. Keep request invocation distinct from underlying bounded HTTP/auth retry. Explain output-token allowance versus file-byte ratio; never truncate to manufacture compression. Provider length/incomplete output refuses commitment even with parseable JSON. The retained Compress API is deprecated external compatibility, never the normal actor's fallback. |
| §15.6 The Graduation Cascade | Keep watermark arithmetic and the five-large-files/eight-files deadlock story. Add selected-file and destination validation, same-identity human edits, exact oldest-prefix retirement and serial successful cascade progress. Explain hidden staging, journaled paired commit, promotion/archive and terminal cleanup in a small table; source remains until the pair commits. Distinguish abandoned pre-commit work from committed unfinished finalization. Recovery precedes startup scans, re-enable, recall and rendering; conflicting files preserve evidence and return an explicit error until repaired. Process crash remains the durability scope. |
| §15.7 Forcing a Handoff | Keep Bill's latest-main correction, provider bucket table and historical recall/prefix experiment limits. Add the narrow95% exception: with an outstanding attempt, defer the next main render and show waiting while settings/interrupt remain usable; failure ends that waiting turn explicitly. Warn/restrict at the existing90/95 thresholds; unknown windows remain unknown. Add once-only final child usage, actual model/tier attribution, preserved main Last and no replayed spending. Explain main cache-lens exclusion and surviving-memory breakpoint here or immediately following the band-layout discussion, recovering current17's deferred obligations. |
| §15.8 Switching Bands Off and On | Preserve the decorative-disk/edit bug and GUI/actor settings deadlock story. Extend relevant-policy generations: off/on cannot revive old work, source/destination settings invalidate/cancel the attempt, unrelated UI settings do not. Re-enable must resolve committed disk recovery before scanning. |
| §15.9 Memory Is Data | Retain the provenance/laundering argument and frozen instruction distinction. Explain that restricted children inherit provider/safety policy, not the owner's primary role or conversation; their output remains owner memory data. Add visible child lifecycle and automatic dismissal after reliable Result/drain/save, retained trace/outcome, and selected browser fallback to the owner. These managed-worker details belong here in the memory chapter, which will follow management after the move. |
| §15.10 Measure It | Preserve the first-edition table and its exact structural/fake-summary limitation. Do not present its synchronous measurement as a new background benchmark. |
| §15.11 The Exercise, Graded | Grader coder owns final rows and observable acceptance wording; author checks readability and agreement with narrative. Keep original11weights/100 and distinguish revised measurements from historical20/100. |
| §15.12 Taking It for a Spin | Preserve case inputs and all five October11 results, including lost reasons, invented uncertainty, unwanted tools and failed recall despite correctly delivered owner/date. Revise reproduction instructions for observable worker outcomes rather than turn-ended/next-request assumptions. State the existing demonstrations exercised the earlier direct compressor. Leave new live claims pending Stage F; never substitute deterministic propagation for semantic quality. |

The source-relative new factual material comes from the accepted B/C record:
the observed no-submit role conflict (handoff §3), independent-turn progress and
95% exception, immutable selected-source policy, staged crash-window repairs,
reliable completion before dismissal and once-only owner accounting. Keep test
counts, manifest hashes, intermediate fixture failures and review history here
or in the linked engineering records rather than repeating them in the chapter.
If a brief new historical observation uses an exact time/token figure, bind it
to that record and say what was actually measured. No invented transcript.

## Chapter 22: The Ensemble

Preserve the steering/delegation opener, Homebrew-VTT boundary story, questions
versus finishing explanation, wiring failures and existing live results. Their
stakes remain: a usable library must be assembled into the actual application,
and a parent must receive work rather than mistake an idle child for completion.

- §22.6 Spawning: explain optional creation-only submit_schema alongside the
  existing factory/parent construction path. Freeze validated copies before
  construction effects; a child's schema is not its grandchildren's default.
  Distinguish available child construction from public fleet-tool enablement.
- §22.8 Asking is not finishing: retain the seven tools, extend spawn arguments
  and the submit row, then show a small non-memory structured result whose schema
  describes the entire direct argument object. Missing/wrong/extra nested fields
  refuse completion; corrected valid data can complete. Default result-text
  submission stays compatible; explicit string APIs cannot bypass custom schemas.
  Explain supported Draft2020-12 subset and refusal of unsupported keywords/refs;
  coder supplies the exact shared public contract/ceilings without duplicating
  validator implementation in prose. JSON validation establishes shape, not truth.
- §22.10: distinguish successful durable submission from actor-tail/save completion
  when explaining lifecycle. Preserve ordinary fleet's three corrective wakes;
  do not accidentally replace that policy with memory-worker's finite Send budget.
- §22.13: extend the existing exercise with a custom structured child and invalid
  then corrected submission, retaining existing default-text fleet demonstration.
  Exact runnable setup is reconciled with the grader coder before publication.
- TL;DR and §22.15 checks: coder owns final contract and scoring. Author retains
  narrative around them. No early memory behavior or managed auto-dismissal here.

## Chapter 23: Managing the Swarm

Keep the sourced opener and traffic-cop/radar explanation, named-peer distinction,
selection race and successful coding/bad-status-label story. Bounded changes:

- §23.3 and §23.4: typed custom payload through owned retrieval, Join/status and
  get_submission; versioned submit.md rows retain the original contract alongside
  payload. Default rows remain result strings; malformed/version-unknown data
  fails. Historical fallback does not complete a new task. Text displays may
  format JSON, but the inner result stays JSON data rather than an encoded string.
- §23.4 and §23.8: keep turn interruption, shutdown and kill distinct; explain why
  submit observed in a manager is earlier than the reliable turn/save boundary.
  Do not claim ordinary task children acquire managed compaction dismissal.
- §23.5/§23.7: existing selected-conversation/replay surfaces display structured
  results; add a precise observation alongside the exercise if needed, not a new
  result browser or product feature. Preserve actual earlier browser receipts.
- Coder owns TL;DR and §23.9 final coverage wording.

## Chapter 16: Auto-Recall

Retain the filing-cabinet/goldfish diary story, BM25 teaching, quota ruling,
judge-format ownership, capped-snippet failure and actual archive demonstration.
Only clarify the accepted interactions:

- TL;DR corpus row and §16.2: active and retired memories remain searchable;
  .compaction/pending.json is neither memory nor recall input. Index construction
  waits until loaded history's recovery barrier permits scanning.
- §16.4: the relevance judge remains one stateless call. Its actual provider/model
  usage enters session totals once while main Last and cache-lens comparisons
  remain main-only. This restores the accounting obligation deferred from17/20.
- §16.5: preserve both actual turn-entry clients, current-context exclusion,
  recorded attachments, capped-source identity and turn cancellation. Explain
  recovery before indexing/attachment; establish committed bytes and refuse
  conflicting disk state before scanning. Retired originals remain searchable.
  Checkpoint removal of
  recall stays the existing behavior, including its cache-boundary consequence.

## Cross-role agreement and deferred mapping

Proposed ownership: author edits narrative after TL;DR, exercises/demonstrations
except numeric graded sections; coder owns complete TL;DRs and §15.11, §22.15,
§23.9 plus any16contract changes. Exchange exact meaningful wording before either
role changes shared blocks. Both plans receive coordinator/independent check
before edits. Independent prose/grader review follows, with full governing context.
Chapter tests/examples and formal results remain coder/coordinator-owned; author
runs only applicable prose/diff checks after drafting, within assigned scope.

Stage E later applies the approved mapping: current17–23→new15–21,
current15/16→new22/23, others unchanged. It repairs chapter links, prerequisite
wording, commands/default paths, headings, section references and publication
inputs together with reviewed reconstruction. Stage D may temporarily describe
fleet concepts at old15 while old22/23 still teach them; this is an intermediate
editorial state, not a ready-to-publish build order. Do not silently rename an
old cumulative export or claim its old score verifies the new sequence.

Stage F still owes bounded live memory-quality/progress/dismissal evidence and
final review. Production migration, installation/restart and publishing are
outside this author's assignment. No new Bill permission ceremony is requested
by this plan; coordinator checking is the already assigned next gate.

## Coordinator plan clarification before release

October14, Astra/root. Current22's TL;DR rule4 still requires a factory-enabled
root to submit and receive three corrective wakes. Accepted runtime
`agent/internal/fleet/fleet.go` exempts Agents without a ParentAgent, following
the approved handoff's existing `df38abe0` main-chat/delegated-submit policy.
Correct that stale distinction in22/23's contract, narrative and affected
examples: the human-facing main Agent finishes in chat; delegated Agents owe
submissions and retain their three ordinary corrective wakes. A named peer is
not a delegated child solely because the GUI displays it. Keep historical
root-submit transcripts attributed to their actual earlier source. This is a
bounded reconciliation with existing authorized behavior, not an additional
runtime design or a replacement of delegated workers' lifecycle. The grader
coder owns corresponding final contract/check wording, with author review.

## October 14 author narrative checkpoint

Author: `compaction_chapter_author`. The released plan at `ac7c67b3`, checked
independently at `215bcd7d`, now has bounded narrative edits in current15/16/22/23.
This is an intermediate author checkpoint against accepted runtime `7feb910a`;
the coder has not yet edited any chapter TL;DR or graded section. Those existing
blocks still describe parts of the earlier contract and await reconciliation.
No final chapter acceptance, renumbering or new live result is claimed.

The source-relative diff preserves the original openers, memory/recall stories,
five-versus-eight explanation, settings deadlock, fleet wiring failures, browser
state failure, historical tables and transcripts. It replaces the synchronous
compressor explanation with the approved restricted child, lifetime, acceptance,
recovery, headroom, accounting and dismissal behavior. Current16 retains its
stateless judge and active/retired corpus. Current22 introduces the agreed direct
schema example; current23 explains typed retrieval and retained JSON. The current
reproduction prompts now distinguish root/named-peer chat from delegated submit,
while the earlier live results retain their actual completion policy.

Coordinator interim feedback corrected two explanations: the hidden staging path
is explicitly this implementation's example, and recovery establishes committed
bytes/refuses conflicts rather than excluding retired originals from recall.
The latter remain searchable. The grader coder agreed the meaningful schema/API
wording before editing and confirmed that this checkpoint contains no unfinished
coder-owned chapter ranges.

Full skill and all four governing/craft documents were reloaded after compaction
before the final corrections. Author checks on these four files:

- `git diff --check` passes; source-relative prose and paragraph endings reviewed.
- `go run ./cmd/lintprose book/chapter-15.md book/chapter-16.md book/chapter-22.md book/chapter-23.md`
  passes all hard checks. Soft counts are 7,713 / 5,243 / 7,131 / 4,853 prose
  words; person-gap warnings persist, and16 retains its negation-density warning.
  No trimming target was applied; the added mechanisms retain their reasons.
- Fence/table comparison finds only the intended memory schema, structured
  launch/result examples, role-policy prompt edits, recovery table and tool
  argument rows. Historical measurements and transcripts are unchanged. All
  three newly added JSON examples parse. An initial blanket parser rejected
  current23's unchanged multi-object WebSocket example as a single JSON document;
  this was a checking-script assumption, not a chapter defect. The corrected
  check parses new JSON examples and inspects unchanged fences separately.

Runtime grading, executable example checks, final TL;DR reconciliation and
independent full prose/grader review remain assigned to the existing later gates.
Stage F still owes the new bounded live observations. Stage E owns the approved
chapter move and coherent cross-reference/sequence changes.

### Interim narrative clarification after review

Author response to `972d7087`: §23.5 now distinguishes same-process selection or
reconnection to a still-registered child from inspecting a retrieved result in
the restored owner's conversation after restart. Ephemeral task children are not
automatically recreated. The existing typed GUI regression explicitly constructs
a schema-bearing root with `NewAgent` and calls `Load`; it does not establish
automatic task-child restoration. This correction adds no product behavior or
new live claim. The full required documents remain in active context without
intervening compaction. Only this paragraph and this note changed; final combined
Stage D acceptance still awaits coder contracts, examples and independent review.
Chapter23 hard prose lint and `git diff --check` pass; the existing soft person-gap
warning remains. No code fence, historical artifact or executable example changed.
