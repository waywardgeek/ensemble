# Edition 2 Chapter 14 / original 15 independent review

## October 11: bounded editorial and supported-model correction accepted

Independent reviewer `reviewer_astra` accepts author correction `6cfa0862` and
model correction `215ae6ac`, with report-only baseline-SHA repair `586b07bd`.
Chapter14 manuscript SHA-256:
`f786497ecda5dc700d8a2456fc86cea0239e2b88585b9f14876df4e574f7f802`.
The author-note SHA is
`9e170186a186fc4969194ac280da80b3e077222ecbf15d8499650eb6745941a5`.
This closes one bounded post-acceptance editorial correction. It does not
rewrite the initial enhancement's zero-round record or the original student's
one implementation correction round. Chapter15 remains paused pending the
coordinator's release; no further chapter work is part of this disposition.

Before finalizing after compaction, reloaded all four complete documents:
workflow, carryover, voice and chapter-writing procedure, plus the entire
[`.agents/skills/ensemble-coding/SKILL.md`](../../.agents/skills/ensemble-coding/SKILL.md).
All four remain active together. Reviewed updated authority headers, the complete
corrected manuscript and author/coder reports, original-relative changes, active
model/test diff and raw validation. The historical acceptance below is preserved,
but its unqualified editorial judgment is superseded by this correction.

### What the audit changed

Bill's concern was justified. My earlier review checked source traceability and
technical accuracy more effectively than it checked unnecessary rewriting and
repetition. I accepted prose that repeatedly interrupted the explanation with
review qualifications. Full-document reading had occurred; reading compliance
did not establish good editing. No invented runtime design was found, but that
did not excuse the editorial drift.

The correction preserves the breakfast opener, original section progression,
manual-loss example, narrated edit, destructive-summary story and stale-address
lesson. It restores the TL;DR's useful numbered explanations instead of merely
retaining their requirements. Necessary corrections still teach actual ownership,
serial batch completion, immediate unload, event validation and the course-only
feature matrix. Repeated certification language and review bookkeeping move to
notes; the relevant cache limitation is explained at the law it qualifies. The
runnable exercise remains because it lets readers apply the mechanism.

The result is 6,569 prose words, with hard lint passing. Counts describe the
change, not its target: the TL;DR returns to about1,100 raw words versus the
original1,115, while the whole chapter is7,936 versus original7,104. Its additional
material teaches verified corrections and an executable exercise. Both narrow
review findings are resolved: tool bytes now explicitly gain a removal policy,
and choosing cuts is described as a tradeoff involving useful work and rereads,
not a categorical claim that rereading achieves nothing. No remaining editorial
or scope blocker was found.

The historical45/100 fixture failure stays separate from the independently found
inline placement defect. Historical Opus5.5/Haiku5.5 identities remain historical;
Bill's restriction and the observed Haiku failure do not become a claim that
supported Sonnet5 failed or that the old failure's cause was experimentally proved.

### Supported models and actual checks

The [coder report](supported-model-correction.md) and
[source bindings](../../solutions/evidence/supported-model-correction/source.json)
cover16 changed source/test files; I independently verified every hash. Chapter5
uses its accepted snapshot's existing Sonnet4.6 row. The legacy snapshot declares
Sonnet5 instead, so its explicit public model-map override is necessary. No
shared compatible Anthropic row was invented and no frozen table was edited.

Active main removes5.5 capability/pricing declarations and uses existing older
rows in current tests and the speech harness. Production shrinks by two lines,
with no new abstraction, export or model-name rejection policy. Unknown text
model behavior remains unchanged. Distinct-model controls still distinguish a
real switch; Sonnet5-to-unpriced-Opus5 checks retain prior measured cost until
unpriced work makes total cost unknown. Removing the second priced Anthropic
model reduces current mixed-known-rate coverage; I accept the documented limit,
not a claim that the old check remains equivalent. Existing OpenAI tier arithmetic
and historical frozen evidence remain. No unsupported prices were fabricated.

Core and GUI Go checks, focused pricing/status/media race checks, revised
Chapter5 boundary checks, vet and JavaScript syntax checks pass. Actual focused
media grading is accepted10/10 and legacy0/10: the latter reaches rendering with
declared models and fails actual blob support. This is not a fresh full Chapter5
grade. The browser selection reports3 passes and2 failures. Both failures recur
on isolated unchanged baseline `5d83c609ddf864f03ab7eedb800cb39e4e1bf034`:
four replies violate earlier one-reply cache/speech expectations. The new known
price exposes the cache discrepancy earlier; it does not create the extra turns.
I accept leaving these reproduced pre-existing failures outside this narrow
correction. Their precise runtime cause is not established by this review.
Listener pending-sleep/EOF controls pass. No blanket browser-pass claim is made.

Only fence22 of24 differs from the previous accepted Chapter14: the launch now
explicitly selects Anthropic/Sonnet5. The exact final manuscript's real GUI and
external-provider-fake run completes7 turns and16 Sonnet5 requests, with six
correlated successful file reads, six cuts, handoff/unload and exit0. I checked
the captured model fields, correlated results and final declarations. This proves
the runnable selection and mechanics, not live provider compatibility or model
quality. Unchanged excerpts and Chapter14 grader retain their earlier evidence;
no redundant paid or broad grading run was performed. The initial stale-manuscript
hash refusal remains a setup failure in the receipt.

Frozen exports have no diff from the recorded baseline. Historical5.5 reports,
raw captured SSE bytes, original chapters/graders and immutable checkpoints remain
preserved. The paused Chapter15 draft is unchanged at SHA
`583da1f152e5f355d2ae8222af30880b60c9cc12db4e984679b48a467387f510`.

## Historical initial acceptance

Independent reviewer `reviewer_astra`, October 11, 2026. **Accepted at initial
combined prose/grader review, zero formal correction rounds.** The preparation
below is retained; the final disposition follows it. Only this chapter is released.

After compaction, reloaded the complete workflow (1,886 lines), carryover (871),
voice (632), chapter procedure (599), AGENTS and the entire mandatory
[`.agents/skills/ensemble-coding/SKILL.md`](../../.agents/skills/ensemble-coding/SKILL.md).
Read all 865 original chapter lines, all 310 student-note lines, all 197
comparative-review lines, and all three original grader files. Original source
revision is `994bf218`, SHA-256
`6c6d953e754f73b8bafd45a2f6009cab5dce9d7557c008e79fcc4084f27c2e79`.
Verified accepted `ch14` tree `0c751dbd3985dee1a80284461e31346471ccc984`
and legacy `ch15` tree `86cb03d46dd900f382266343229908a38290aad9`.

## Scope preflight

The proposed separately versioned grader preserves all eight original checks
and their 100-point total. Public GUI `update_settings` followed by the effective
settings snapshot is a justified replacement for writing launch-cwd settings:
the accepted GUI resolves settings through the owning Agent's DataDir and
applies the target to that Agent. A declared `list_directory` replaces the
undefined `think` fixture without adding a product capability. Keep/stub controls
must retain the original default target, so the independent ladder does not
invalidate the result used to distinguish per-round stubbing.

The original handoff scenario reads a file and calls handoff on separate model
responses, despite claiming same-batch coverage. A genuine same-batch read and
handoff plus the three captured records tests the stated obligation. Serial
dispatch still needs this protection: the later call in the batch may not have
started when the survivor arrives.

Prefix equality and a declaration name found anywhere in messages are insufficient
evidence of valid inline operation. The proposed actual MCP load/invocation/unload,
correlated result checks, fresh automatic ephemera and strict declaration-block
placement directly address the real prior defect. Existing deterministic source,
browser and failure controls remain separate evidence for disk ordering,
atomic malformed-event skipping and UI details. No new paid run is requested.

## Teaching and historical boundaries to retain

The manual disappearing while its tools remain, visible narration preserving
decisions, destructive summaries, and stale job-output addresses give the
mechanism concrete consequences. Preserve their explanations and sourced
attribution rather than replacing the chapter with a contract ledger.

The explicit course feature matrix governs over the contradictory real-Opus
narrative. Retain original **45/100 FAIL**, the initially permissive fake, and
the genuine renderer correction accepted after one implementation round. Gemini's
successful live task and the Anthropic/OpenAI failures have different outcomes;
zero client exit is not task success. No live inline/stubbing or reasoning-quality
improvement is established by those runs.

T controls tool bands, not the whole conversation or a universal request maximum.
An oversized fresh result must still be shown whole once; the original `3T/8`
comment is not an unconditional bound on every outgoing request. Process-crash
recovery does not promise power-loss durability. Final review will separately
assess complete prose, exact examples, revised source, accepted/legacy results,
mutations and their actual failure causes before a combined decision.

## Final combined disposition

Accept author `21c09849` and grader `4918a9b1`. Final manuscript SHA-256 is
`6e29699fc23fa7a3b2d4a549524625b125ac7c69600baeb686754e86ad570117`.
Read the complete final author and coder reports, whole manuscript, original
chapter and historical review, original and revised graders, relevant accepted
and legacy implementations and their own-predecessor comparison. After the
latest compaction, fully reloaded all four governing documents and the entire
`.agents/skills/ensemble-coding/SKILL.md` before continuing. All remain active.
Frozen accepted and legacy tree identities above are unchanged.

The chapter earns its length. The breakfast opener, three questions about bytes,
manual disappearing while tools remain, narrated edit, destructive summary and
stale output address preserve the original teaching and personality. The added
working example and actual failed recollection make the mechanism useful to a
reader. The prose distinguishes policy, reduction and rendering through their
owners instead of becoming an implementation checklist. Hard lint passes at
6,774 words; neither the word count nor lint establishes that judgment.

Resolved draft-stage findings narrow the narration-only and universally-unmeasured
claims, distinguish valid JSON from invalid provider placement, and separate
numeric skip/continuation coverage from atomic validation. Final source review
also corrected the inherited global Seq-order claim: held survivors retain their
recorded identities but append after later companion results. The diagram and
§14.4–14.5 now agree. These corrections add no behavior. No remaining readability,
scope, source-fidelity or technical blocker was found.

### Grader and comparative assessment

All eight original checks and weights remain, totaling100 without parity.
Effective settings through the actual GUI, explicit subscription, declared
`list_directory`, and observed turn completion replace the identified path,
name and timing assumptions. Keep/no-stub retains its separate large target.
The same-batch handoff really executes its companion and checks the recorded
call/result/note order. The MCP scenario checks typed declarations/removals,
successful correlated argument-dependent output, fresh automatic content and
placement. It does not merely search call arguments or accept a connection as
execution. Existing atomic reduction, disk-failure ordering, browser editing and
other unchanged source proofs remain separate from numeric coverage.

Independent review found the first-green placement predicate rejected legal
consecutive system-message groups. The [provider placement contract](https://platform.claude.com/docs/en/build-with-claude/mid-conversation-system-messages)
constrains the group, rather than requiring renderer coalescing. An exact
`35fe32fe` driver overlay reproduces both false rejections; the single-message
control and invalid post-group-user control behave correctly. The unchanged
four-case probe on final driver `b971a8cb…` passes in0.010s. The first attempted
probe had already picked up the coder's correction and is preserved as such,
not relabeled a baseline failure. The grader also passes the actual split-renderer
control. This fixes representation bias without weakening placement.

Physical/nonblank-noncomment/comment-only counts, using stripped `//` lines:
original three grader files1034/858/109; revised two files1243/1025/153; focused
tests63/56/4, counted separately. Embedded fixtures are included in these counts,
so noncomment lines are not claimed as parsed Go statements. The167-line active
increase retains the original scenarios while adding public settings/completion,
actual MCP and stronger delivery observations. This is proportionate work, not a
new framework. The roughly13% comment share is below the teaching target but the
comments explain ownership, scenario isolation, correlation and completion.
The new exported Chapter14 entry and visible fields have useful documentation;
the misleading diff-helper comment was corrected. No comment padding is needed.

The prior runtime comparison remains applicable: policy belongs to Agent,
reduction/journaling to History and request translation to Engine, rather than
the reference's common-package behavior and callback access. Its recorded
own-predecessor production/test comparisons are preserved in
[the original comparative review](ch15-code-review.md); no frozen implementation
was changed or recounted as enhancement work. The revised grader drives those
real applications with external provider/MCP fakes, without mocking their
actors, jobs, reducers, transport or filesystem effects.

### Actual validation and evidence limits

Final complete runs are **accepted100/100 in0.755s and legacy100/100 in0.769s**,
both exit0. These are actual full reports, not composed totals. The earlier
35fe32fe runs remain separately recorded. Focused grader race checks pass in1.895s;
vet/formatting pass. Race instrumentation covers the test process, not the
separately built applications. I verified every final grader hash and both
frozen tree IDs; the [independent receipt](../../solutions/evidence/ch15/enhancement14/reviewer-final-check.json)
records those checks and the narrow probe. No redundant full grader or paid run
was performed by this reviewer.

All fifteen distinct intended defect claims are supported by their patches and
actual failure predicates. Coupled failures after wrong target, premature handoff
and nonzero exit are not additional independent proofs. The initial partial-read
case scored85 while incorrectly passing full delivery: its unrelated restart
failure did not validate that assertion. Requiring all planted bytes now catches
the same defect at total70. Broadly retained pending ephemera instead loops for
1,416 requests and times out before freshness inspection; it is properly excluded.
Narrow duplicate-wire and real cached-snapshot defects hit the intended content
checks. The redundant marker perturbation is not counted twice. Failure stderr
is empty; no compilation or cleanup failure is credited as a behavioral kill.
Split-system and650ms-delayed-completion controls both pass100/100; restored
controls pass. Reused unchanged mutations retain their narrower original claims.

Exact examples were executed at manuscript `55e3a4a…`. The actual sample requests
contain six successful first file results,16 total requests and the final manual
and handoff once each. The script measures7 turns,6 recorded cuts,1 handoff and
normal exit. The vocabulary fence compiles; inserted survivor/replay fences run
with actual Context tests. Of24 fences, the sole subsequent change is the layout
diagram; executable samples remain byte-identical. Preserve the actual tested
hash rather than calling this a new execution. The initial last-resent-result
validator error remains recorded separately.

Original **45/100 FAIL**, the permissive fake and the real first-round wire
correction remain historical. The chapter properly attributes reference byte and
Opus/cache reports, and separates those from the accepted solution's course-only
feature matrix. I independently decoded the repeat-Claude capture: request3
contains the successful CEDAR-17 result before later redacted copies and the
quoted denial of ever reading it. Only Gemini completed the retained task;
zero exits do not turn the other sessions into successes. No new live inline,
cache, reasoning-quality, listening, power-loss or universal-size claim follows.

No enhancement exception or formal correction round is needed. The coordinator
may checkpoint Chapter14 and release Chapter15 under the existing sequence;
this review itself neither drafts nor implements the next chapter.
