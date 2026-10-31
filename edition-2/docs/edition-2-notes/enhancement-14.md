# Edition 2 Chapter 14 enhancement: The World's Best Context Engineering, Before Breakfast

Author: `edition2_author`, October 11, 2026.

## October 11 editorial audit and bounded correction

**Current status: correction validated and submitted for independent review.
Chapter 15 remains paused.** This supersedes the dated
pending/acceptance states below without changing their evidence.

Bill questioned the chapter's growth and preservation of his voice. The audit
compared original Chapter 15 (7,104 raw words; 6,106 lint prose words) with the
accepted Chapter 14 (8,141 raw; 6,774 prose). The original TL;DR had 1,115 raw
words; the accepted version had 710. The new runnable section accounted for
698 raw words. Root and reviewer found no invented runtime design, but the
author agrees that repeated qualifications and review bookkeeping diluted the
voice, and some sentence rewrites were unnecessary. Reviewer withdrew editorial
approval pending this bounded correction. Earlier review `f3d6e7f0`, checkpoint
`5f07a062` and tag `edition-2-enhanced-ch14-r1` remain historical.

Writing stopped for the audit. Chapter 15 remains an untracked partial draft,
unchanged SHA `583da1f152e5f355d2ae8222af30880b60c9cc12db4e984679b48a467387f510`;
no Chapter 15 note, lint, validation or acceptance is claimed. A compaction
occurred after that pause and was disclosed before any writing resumed.

Before this correction, the author reloaded all four full documents, with every
range returned completely: workflow 1–300, 301–650, 651–1000, 1001–1350,
1351–1700, 1701–1937; carryover 1–300, 301–600, 601–886; voice 1–330,
331–632; procedure 1–300, 301–599. No range was truncated. Also reread the
entire mandatory `.agents/skills/ensemble-coding/SKILL.md` (1–373) and the
new AGENTS/enhancement headers. All four remain active together; no subsequent
compaction has occurred. Full original Chapter 15 and the audited Chapter 14
were reread side by side in bounded chunks before editing.

The correction restores the original explanatory TL;DR progression, including
why survivors need kinds, the three handoff records, previous-batch keep,
recorded sequence bounds and checkpoint ordering. It keeps the specific
corrections for current ownership, serial dispatch, immediate executable
unload, provider placement, known-event validation and course-only flags.
The small tested vocabulary fence stays; the original obsolete numeric-event
ellipsis and flat implementation snippets are not reinstated. The original
opener, three byte questions, manual-loss explanation, two cache laws,
narrated edit, checkpoint lineage and stale-address story remain. No new
mechanism or requirement was introduced.

Prose once again explains the mechanism directly instead of repeatedly saying
what the review did or did not prove. Historical accounts are attributed at their
entry; source hashes, mutation status, paid-run bookkeeping and detailed fixture
provenance stay in this note and the linked records. Necessary limits remain
where useful: prefix reuse versus billing in §14.3, tool budgets versus a total
window cap in §14.7, process crash versus power loss and retention in §14.12.
The runnable exercise and actual Haiku denied-read lesson remain.

Bill's model restriction is stated separately from the transcript: use supported
older Anthropic models such as Sonnet-5; no 5.5 recommendations or future tests.
The historical Opus 5.5 and `claude-haiku-5-5` identities remain unchanged.
The Haiku failure is not offered as a result for Sonnet-5 or as causal proof
of Bill's reported incompatibility. No successful Sonnet live outcome is invented.
The GUI launch now explicitly sets `LLM_VENDOR=anthropic` and
`LLM_MODEL=claude-sonnet-5`, already present in the frozen model table.

Exact artifact comparison against `5f07a062` finds 24 fenced blocks and only
one changed block: the GUI launch's two model environment values. All Go,
Python, YAML, JSON and prompt blocks remain byte-identical. The original
sample binding `55e3a4a4…` remains the earlier execution, not a rerun of this
revision. The existing numeric grader is unchanged; final full 100/100 accepted
and legacy results retain `4918a9b1` provenance. No paid run or frozen edit is
needed for this editorial correction.

### Correction validation and final handoff

The independent prose reviewer caught two draft-stage wording errors during this
correction: the TL;DR's opening could imply tool bytes stop without a policy,
and a new sentence incorrectly said repeated rereading gains nothing from a
smaller request. The first now says tool bytes gain a removal policy; the second
judges usefulness by the work and rereads required. The historical reread tradeoff
remains. The original 45/100 fixture failures are also explicitly separate from
the independent review's inline-placement finding. These corrections introduce
no runtime or grader work.

Final manuscript SHA-256: `f786497ecda5dc700d8a2456fc86cea0239e2b88585b9f14876df4e574f7f802`.
Direct `go run ./cmd/lintprose book/edition-2/chapter-14.md` passes all hard
checks at 6,569 prose words. Its one soft person-gap warning was read; no story
was invented to satisfy it. Raw whitespace counts are 7,936 for the whole
chapter and 1,100 for the TL;DR, versus 7,104/1,115 in the original and
8,141/710 in the audited enhancement. These measure the edit, not its objective.
The original-relative reading and complete correction diff were inspected;
`git diff --check` passes. The chapter retains its original section sequence,
with the previously authorized runnable section after the historical live account.

The coder's [new sample receipt](../../solutions/evidence/supported-model-correction/samples.json)
binds that exact manuscript. The changed Sonnet-5 launch passes through the real
GUI with the existing external HTTP fake: seven completed turns, sixteen actual
requests, six successful correlated reads, six recorded redactions, one handoff,
load/unload and normal exit zero. The author inspected the runner, receipt and
[captured requests](../../solutions/evidence/supported-model-correction/samples-requests.json):
all sixteen select `claude-sonnet-5`; first successful result deliveries occur
at requests 3/5/7/9/11/13. The final request has the manual and checkpoint once,
and only the four lifecycle/context tools remain declared. This verifies model
selection and runtime mechanics with an external fake, not live Sonnet judgment,
ephemeral-context compatibility or provider cache behavior. Unchanged Go sample
checks retain their earlier execution.

The first launch-validation attempt stopped at its manuscript hash guard before
launch because the author had changed a score-attribution sentence after supplying
the earlier hash. That failed attempt remains recorded; it is not a runtime
failure. The final run waited for the frozen hash above.

Only this manuscript and author note are in the author commit. No first-edition
file, frozen solution, grader, Chapter 5 or paused Chapter 15 file changed.
Independent reviewer owns the correction decision; root owns its checkpoint and
any later release. All four full governing documents remain active after the
complete reload documented above, with no subsequent compaction. Mandatory
technical handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## Earlier author record (retained)

The following sections describe the initial enhancement and its final handoff.
Their pending statuses predate the acceptance and subsequent editorial audit.


## Reading and outline

The author read the full original chapter (865 lines), complete student notes
(310 lines) and comparative review (197 lines), plus the historical
`book/review-ch15.md` (182 lines). Original teaching, reviewed implementation and
actual student results supply the revision; later main is not its source.

A compaction occurred during this draft. Before any subsequent authorship, the
author reloaded all four required documents completely: workflow 1–1886,
carryover 1–871, voice 1–632 and chapter procedure 1–599. All four remain active
together. The mandatory `.agents/skills/ensemble-coding/SKILL.md` was reread in
full, 1–366; an outer-output truncation was recovered with a separate complete
read. Summaries did not substitute. Repeat all four complete reads immediately
after every subsequent compaction. Current AGENTS, enhancement assignment and
chapter map govern scope and preservation.

Read accepted common/context, Agent policy, context/skill tools, reducer,
save/journal, Anthropic renderer and GUI settings in full, with relevant actor,
History restoration, Engine request boundary, skill registry, model-table,
application configuration and GUI startup paths. Read the first-edition skill
payload declaration to distinguish its older captured body from the student's
new survivor event. The historical design record was read selectively (opening,
A.11 cache/stale-handle notes and honesty section), not claimed as a full read.
Git history confirms the first design commit's 07:42 time; the dictation/shower
account remains Bill's original testimony. Enhanced Chapter 13 supplies continuity.

Through-line: a person wants the agent to continue useful work after removing
old tool traffic, without losing its manual or mistaking a smaller window for
accurate recollection. Preserve Bill's breakfast opener, three byte questions,
manual-lost-before-tools example, two cache laws, narrated versus silent edit,
compress_context and handoff stories, stale output address and historical live
rereads. Resolve the stake with recorded cuts, distinct survivors, replay and the
actual Claude recollection failure. No invented story or length-reduction target.

Teaching order remains the original: motivation and contract; working set versus
storage; reasons for bytes; prefix laws; survivor kinds and same-batch holding;
layout and narration; ladder and keep policy; checkpoint lineage; recorded replay;
stale-address recovery; process-crash journal; settings and historical byte study;
exercise; preserved reference live account; actual accepted implementation exercise
and five candid student outcomes. The added final section supplies reproducible
public inputs without replacing the original live story with an incompatible one.

## Source identities and historical evidence

Original `book/chapter-15.md`, file revision
`994bf218b31033b7dbf2740dd0187d509d2927dc`, SHA-256
`6c6d953e754f73b8bafd45a2f6009cab5dce9d7557c008e79fcc4084f27c2e79`.
Direct accepted `solutions/edition-2/ch14` tree:
`0c751dbd3985dee1a80284461e31346471ccc984`, the 108-file export checkpointed at
`edition-2-ch15-r1`. Corrected source is `80b3e09`, one formal student correction
round. [Student notes](ch15.md) and [independent comparison](ch15-code-review.md)
retain the prior attempts and accepted exceptions. Both source hash and export
tree were rechecked during this author pass.

Original grading remains **45/100 FAIL**, make exit 2/grader exit 1, on initial
`165ee52424634d4e848c257a6a63745fed4d1cb4`, not the corrected accepted source.
The author inspected the raw score/failure block and
[provenance](../../solutions/evidence/ch15/original-grade-provenance.json).
Three failures follow a fixture's wrong settings path; the intended small target
never reaches the Agent. A fourth asks for undefined `think`. Handoff 15, keep 10,
crash 15 and reducer 5 produce the actual 45. Those awards do not independently
prove every clause, and masked assertions are not passing obligations.

Historical `book/review-ch15.md` reports the reference synthetic comparison:
64 requests in each run; total request bytes 17,470,874 without ladder,
6,172,202 with ladder and 2,334,066 with stubs; largest ladder request 146,378;
2,991-byte stable prefix. The chapter attributes these measurements to that
review. Its reported 100/100 and mutation totals are not newly run student evidence.
The source's September 22 CodeRhapsody cache account reports 77% cumulative after
a cold request over 100,000 tokens. Neither that account nor synthetic byte counts
measures cache grants on the accepted student solution.

The original chapter's September 23 real-Opus narrative is retained as its reported
reference evidence: 11 requests, 10 calls, nine stubs, seven source files totaling
60,865 bytes, two rereads and zero reported cache reads/writes. The current pass
did not locate/reproduce a separate raw live capture for that historical table.
Its feature configuration differs from the explicit course-only matrix. Quoted
model commentary that a retained Ref makes dropped bytes fetchable is explicitly
not adopted as a durable-address guarantee. Preserve the observation without
silently validating the model's explanation.

## Teaching dispositions

- Preserve the ambition/title and Bill's original opener; state the design
  record's own distinction between a best-known design and a demonstrated best
  system. The reader's working-set problem carries the later mechanism.
- Correct blanket bounded-context and never-lose-bytes language. T governs tool
  bands, while dialogue and survivors can grow. A newly delivered oversized
  result must be sent whole once; completed batches constrain cuts. The sum 3T/8
  is not an absolute request maximum. Retention can discard older raw history.
- Explain actual owners: Agent chooses policy using its configuration; History
  validates, journals and projects; Engine renders. Common holds vocabulary.
  No callback-bag design, new framework or later memory feature is imported.
- Keep the frozen startup instruction and, on inline-capable rows, startup tool
  array. Non-inline rows fold recorded changes into effective top-level tools.
  Stable request bytes are not cache grants or billing measurements.
- Label the beta header and nested addition/removal blocks as historical course
  wire contracts. The course Opus row alone enables both new features; Sonnet
  course and all existing real rows leave them false. No current API or live
  feature support claim is inferred from the dated source.
- Keep survivor kinds separate from tool parts. The student's predecessor had
  its manual in load results; this chapter records it separately. First-edition
  SkillData history is attributed as that implementation's history.
- Correct parallel-dispatch prose to the actual serial batch. A handoff can run
  before a companion has started, so defer destructive effects until the whole
  model reply's calls have results. Revised grading supplies a genuine same-batch
  companion; the original scenario was sequential despite its description.
- Separate immediate executable-grant/server removal from retained instruction
  text. Shared active grants survive; guessed removed tools remain refused.
  Instruction-removal compaction is a later exercise.
- Preserve narration/checkpoint reasons while qualifying their quality. Handoff
  is available across rows; per-model automatic stubbing is a separate policy.
  Prompting for intent cannot guarantee truthful narration or successful recall.
- Preserve replay's numeric recorded cuts, sufficient-log precondition and
  validate-before-mutate skip boundary. Unknown vocabulary or broken save JSON
  still refuses. Numeric reducer credit covers observed skip/continuation;
  atomicity has separate source/integration proof.
- Explain selected-save journal, previous snapshot and checkpoint order for
  process crashes. No fsync/power-loss guarantee, speculative partial-write
  repair or claim that retained snapshots recover discarded older history.
- Publish the actual Agent data path (`cr/settings.json` for this application),
  public settings update and protected unsent number edits. The old adjacent
  settings fixture does not authorize new workspace semantics.
- Retain the first-edition byte-study/live account with attribution; remove the
  contradictory statement that no live session existed. Add the student's
  distinct results, including both failed providers and Gemini's success.
- Preserve the eight original weights, totaling 100 without a parity row, while
  correcting fixtures to effective settings and real `list_directory`. Separate
  coverage titles from stronger claims requiring supplemental proof.

## Direct evidence checks and limits

The author parsed all five raw live artifacts' provider requests and actual
result blocks, not just whole-request substring matches. Request counts are
9 initial Claude, 15 initial OpenAI, 10 Gemini, 15 repeat Claude, 18 repeat
OpenAI. Actual result content contains all six inspection codes in every run
except initial Claude. Event logs independently yield result/tool cut counts
0/0, 0/0, 5/2, 6/1 and 6/1, and handoff counts 0, 11, 1, 0 and 4. All exits are
zero; only Gemini completes the requested task. All real feature flags are false.
These are retained October 10 local-time runs, not new paid tests.

The raw repeat-Claude stdout says “That file was redacted, so I never read it.”
Its earlier actual provider-bound `tool_result`, call ID
`toolu_01HqHzGapgj6ABgk6ZfJ6cik`, contains CEDAR-17 and verified status, followed by
the planted filler. This supports the manuscript's specific contradiction rather
than a universal claim about model memory. Repeated rendered stubs are copies
across requests, not a count of distinct redaction events.

Read [initial independent probes](../../solutions/evidence/ch15/reviewer-probes.json),
[round-one verification](../../solutions/evidence/ch15/reviewer-round1.json)
and [restored-defect result](../../solutions/evidence/ch15/round1-mutation.json).
Initial prefix/MCP/replay checks passed a permissive fake; adding faithful
placement rejected roles user/assistant/user/system/user. Corrected source stages
declarations after ephemera, producing user/assistant/user/system. Restoring the
old renderer yields the intended placement assertion and HTTP 400. This was
valid JSON rejected for provider placement, not malformed JSON or a settings bug.

The corrected browser receipt preserves target 20,000, retention 3, zero/default
values and an unrelated cwd settings file. The restored GUI defect fails with
“broadcast erased unsent context edit.” Source/browser evidence establishes that
pending human edit, beyond an echoed settings value. No browser rerun is claimed.

## Initial stage checks and next gate

Stable initial manuscript SHA-256:
`55e3a4a4900f8cfbd10b65aa58ce000435d7f404129937b1c02394c8095cf816`.
Direct `go run ./cmd/lintprose book/edition-2/chapter-14.md` passes hard checks,
6,753 words. Soft negation/person-gap warnings were read; they do not justify
removing necessary distinctions or adding invented people. The original teaching
progression and stories earn the current length.

Coder has the exact current Go excerpts, course inline JSON, complete two-skill
fixture, Python preparation and optional GUI build/launch. The new final prompt
asks for actual failed approaches instead of prescribing the historical false
“no approaches failed” note. It is explicitly a reproducible variant, not a claim
of identical historical bytes. Exact sample execution and final revised grading
remain pending; no prospective score is printed as fact.

Only this note and `book/edition-2/chapter-14.md` are author-owned. Original
chapters/graders, accepted snapshots and CodeRhapsody work remain untouched.
Independent reviewer owns combined acceptance; root owns checkpoint and next
release. All four full documents remain active after the reload recorded above;
no further compaction occurred before this handoff. Every technical handoff
requires `.agents/skills/ensemble-coding/SKILL.md`.

The final cross-reference pass maps the inherited wrong-data observation to
Edition 2 Chapter 12. Local links resolve; sample bytes are unchanged.

Draft-stage reviewer feedback narrowed two inherited claims: readable narration
survives tool clearing alongside retained opaque material, and the retained
experiments did not measure reasoning-quality effects. The prose no longer
claims narration is the only survivor or that nobody has measured the question.
These are prose corrections; exact sample bytes remain unchanged.

## Exact sample reconciliation

The coder's [sample receipt](../../solutions/evidence/ch15/enhancement14/samples.json)
binds unchanged manuscript SHA `55e3a4a4900f8cfbd10b65aa58ce000435d7f404129937b1c02394c8095cf816`.
The exact published optional-GUI build/launch, both complete skills, Python
records/settings and seven human prompts run through the actual application
with a small external Anthropic HTTP fake and a declared non-stubbing course
model. The process exits zero after 16 requests, six successful reads, six
recorded cuts and one handoff. This is executable sample evidence, not a paid
run, model-quality finding or native-browser replay.

The author independently inspected [captured requests](../../solutions/evidence/ch15/enhancement14/samples-requests.json),
correlating calls and results: requests 3/5/7/9/11/13 first carry the six successful
record results with their exact path and code. The last request retains the
manual once and the checkpoint note, while its declaration array contains only
micro_handoff, keep_tool_results, load_skill and unload_skill. The sample has
actually removed the file reader rather than merely saying it did.

The [Go receipt](../../solutions/evidence/ch15/enhancement14/samples-go.txt)
shows the vocabulary snippet compiling and the exact survivor/replay excerpts
running in their actual owning packages with Context tests passing. Historical
fragments remain labeled as historical; no old API is revived. No manuscript
sample correction was required. Final grader source and mutation reconciliation
still await the coder's frozen handoff.

## Final grader and author reconciliation

Read the complete [final coder report](enhancement-14-code.md) at `4918a9b1`,
its final-source binding and both raw final command/results. All four grader
source hashes match the binding; accepted and legacy frozen tree identities
match their retained source. Actual final complete grades are **100/100 accepted**
in 0.755 seconds and **100/100 legacy** in 0.769 seconds, both exit zero and all
eight checks passing. These are full runs, not composed totals. Earlier
`35fe32fe` full 100/100 runs in 0.814/0.842 seconds remain separate provenance.
Original student 45/100 remains unchanged.

The final grader observes public settings before traffic, waits for actual
completed turns before EOF/SIGKILL, checks complete correlated file bytes and
same-batch handoff chronology, and follows real external MCP additions, calls,
automatic content and removals. No new runtime mode, fixture-only tool or
settings-path behavior was added to frozen solutions. Revised placement is
validated from observed requests; the historical strict fake's HTTP 400 remains
a separate source regression, not a claimed feature of this revised HTTP peer.

Fifteen distinct targeted defects fail their intended predicates, with two
passing 100/100 positive controls: legal split system-message groups and delayed
650 ms completion observations. The delay control's full run is 10.248 seconds.
Coupled downstream failures after wrong target, early handoff or nonzero exit
are not independent proofs. The first partial-read mutation incorrectly passed
the intended whole-delivery check at total 85/100; a separate restart failure
could not establish that claim. Requiring all planted bytes catches that exact
defect. Its final total 70/100 retains both failures.

The broad pending-ephemera mutant produced 1,416 requests and timed out before
the intended content predicate. It remains an explicitly limited runaway result,
excluded from the fifteen intended proofs. Narrow duplicate-wire and stale-cache
mutations subsequently reach their intended assertions. An extra marker
perturbation is also not counted as an independent defect. The initial missing
subscribe setup, sample hash guard and mistaken latest-resent-result validator
are retained as setup/verification failures, not product defects.

The independent review's valid consecutive-system-group control found a grader
representation error: separate system messages can form a legal group. The
corrected check preserves rejection of user content after the group without
requiring the implementation to coalesce it. The initial probe already used
corrected source; the reviewer subsequently reproduced the old failure using
an overlay of exact first-green driver source. Neither attempt is relabeled.

Focused race checks and affected vet/formatting pass. Race instrumentation covers
the grader test process, not its separately built application binaries. Existing
source/integration/browser evidence retains its specific jobs, atomic reduction,
journal/checkpoint, feature-table, settings and historical strict-boundary claims.
No new paid inference, audible-quality result, provider cache saving, power-loss
guarantee or universal context bound follows from these grades.

A final reviewer source check corrected §14.5's inherited sorted-Seq claim.
Held survivors retain their event identities but land after the completed batch,
possibly after results with later sequence numbers. Both diagram and explanation
now agree with §14.4 and the actual reducer. This is a prose/diagram correction;
all executable examples remain byte-identical to the tested sample manuscript.
The coder preserves the actual sample SHA rather than calling it a rerun.

Final manuscript SHA-256:
`6e29699fc23fa7a3b2d4a549524625b125ac7c69600baeb686754e86ad570117`.
Direct hard prose lint passes at 6,774 words. Sample execution remains bound to
`55e3a4a4900f8cfbd10b65aa58ce000435d7f404129937b1c02394c8095cf816` with the
later diagram/wording difference recorded. Historical stories, model/task failures,
source versions and course-only feature flags remain intact. All four complete
governing documents remain active after the recorded reload; no further
compaction has occurred. Independent combined acceptance and root checkpoint
remain pending, and Chapter 15 is not released.
