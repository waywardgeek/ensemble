# Original Chapter17 / Edition2 Chapter16 — independent review

## October11 public-contract preparation — no acceptance

Reviewer `enhancement_reviewer_resumed` read all 552 lines of original
`book/chapter-17.md`, file revision `c4770e9ba9475fcb6622df7349d0214c88ba3af9`,
SHA-256 `9b77e32240d10743da3abf0dfd22e585cebcc382990f0517f68d9d006a02f893`,
and current [student-safe preparation](ch17.md). Mandatory skill, carryover
and workflow have been read; subsequent Bill rulings are recorded at `7646941`.
No original answer, grader implementation or changing student source was read
for this preparation. No code/test/live execution or formal review round.

Bill's two decisions resolve the consequential source contradictions: final
mechanical quotas apply when the judge is absent or fails, without overriding
successful judge choices; search all memory files while skipping content already
in actual context. The historical newest-two/system-prompt premise is superseded.
The future review must distinguish a valid empty judge selection from failure,
and recorded recall bytes from a new search during replay. These are existing
chapter obligations, not additional features. Actual public turn entrypoints,
checkpoint lifetime, changed-file indexing and source attribution remain binding.

The chapter describes algorithms and defaults but supplies no concrete executable
grading protocol. Coordinator should recover only missing invocation/fixture
facts if needed: launch target/flags and input envelope; exact configuration
keys for source paths and judge settings; corpus record envelope, especially
the `long` field in learnings; any public inspection/output envelope required
by the named checks. Array judge replies, defensive parsing, score thresholds,
chunk boundaries and byte cap are already public and need no answer disclosure.
Fixed harness paths or source-shape checks cannot silently supersede Agent-owned
files and the optional GUI boundary. No requested contract recovery authorizes
private CodeRhapsody corpus access, a new diagnostic framework or model features.

The student is implementing from accepted original16 tree
`4a253c57dd62b4a6ad086618574ef46af7c2d724`. Formal comparison waits for frozen
submission, completed original grading, meaningful mutations and bounded live
evidence. Original22/24 stops and the stopped author phase remain. Required
handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## October11 grader preflight — first green source, not acceptance

Read the current student checkpoint note and narrowly inspected frozen source
`fcd091f619f879341d4bee318e5f1bf8fb4bb21b`, main tree
`a2fd5a4f71d8fde7f58cfdefd18a0cfc31206d0f`, against the unchanged original
Chapter17 launch/check code. The coordinator's full original run is still in
progress at this writing; the observed log repeatedly reports
`skill "base" has invalid type ""`. No final score or successful runtime
behavior is inferred here. No grader was modified or repeated by the reviewer.

**The missing skill type is a justified fixture exception.**
`ch17_run.go:325–334` writes a base skill with name and description but no type.
The launch selects it as primary (`ch17_harness.go:71–72`). The inherited student
parser requires one of primary/loadable/dependency (`internal/skills/parse.go:96–99`);
original Chapter10 lists these choices and supplies typed examples without
defining an absent-type default. GUI construction consequently fails at
`root.NewAgent`, before HTTP startup or any judge/main request. This repeats the
independently accepted original13 fixture mismatch; no permissive parser change
or dummy primary is warranted merely to satisfy this fixture.

**Two separate assumptions would remain after fixing that launch fixture.**
The GUI selects Agent DataDir `cr` (`gui/cmd/main.go:77`), while grader fixtures
are written below cwd/docs, cwd/handoffs, cwd/learnings and cwd/memory/memory
(`ch17_run.go:82–206`). Recall's source scan correctly follows its Agent DataDir
and configured skills directory (`internal/recall/recall.go:79–96`). No launch
override selects the fixture's cwd as Agent identity. Skills remain separately
reachable through EN_SKILLS_DIR; that does not make the other corpus roots
reachable. This is a latent owner-path mismatch, not an observed BM25 defect.

The structural check (`ch17_run.go:831–915`) examines `internal/recall` beneath
the submitted GUI module. The required separate optional GUI consumes the core
module, where the actual recall spoke lives. Absence below GUI is not absence
of the implementation. Its import detector also exempts paths not containing
`/ensemble/`, so it cannot establish the student's `ensemble/internal/...`
star graph merely by returning success on a different target. Actual core
imports, common behavior and parent relationships still require review.

Nine behavioral checks totaling90 points are masked by launch failure:

| Check | Valid obligation still requiring final supplemental evidence |
| --- | --- |
| bm25-indexes (10) | Real owner corpora, including docs/handoffs/learnings/skills and all eligible memories, are searchable and refresh after edits. |
| bm25-scores (10) | Smoothed IDF, TF saturation and length normalization produce meaningful ranking on a representative corpus. |
| bm25-stop-words (5) | Query stop words are excluded while document length statistics retain them. |
| chunking-splits (10) | Relevant sections, oversized subsections and paragraph chunks retain attribution without unrelated whole-document injection. |
| per-source-quota (10) | Final mechanical selection reserves source shares and redistributes unused slots, while successful judge choices stand. |
| judge-filters (15) | An actual stateless model call changes selection; valid empty choices remain empty, and defensive reply parsing handles documented forms. |
| judge-fallback (10) | Malformed/error/timed-out judge calls degrade to mechanical selection; empty corpus and short/disabled paths remain usable. |
| recall-is-own-kind (10) | Actual entry/event carries immutable selected bytes separately from human input, persists until checkpoint, deduplicates, and replays without retrieval. |
| injection-capped (10) | The whole formatted attachment meets the chapter's6144-byte default cap, including attribution, rather than only the grader's looser8192 bound. |

The remaining10-point structural obligation is valid but its fixed submitted-path
expectation is not. None of these exceptions establishes the masked behavior
passed. Final review must inspect actual runtime tests, mutations, captured
requests/replay and live use; no weakened grader or fabricated total is needed.

One additional latent own-kind predicate deserves final assessment:
`ch17CheckOwnKind` rejects recalled text sharing an Anthropic wire message with
the prompt and categorically rejects role:user (`ch17_run.go:425–451`).
Chapter17§17.5 instead separates stored conversation identity from renderer
placement and explicitly permits an extra part on an existing turn. The existing
Anthropic renderer coalesces adjacent user data. A separate recorded entry is
therefore not equivalent to a separate vendor message or a universal midstream
system role. The current failed launch has not reached this predicate; preserve
that distinction and independently inspect valid wire placement at final review.
Do not add an unconditionally unsupported system message for this source check.

Source provenance: harness/run revisions and hashes match the student-safe
erratum. `ch17_checks.go` revision
`a91362b61fcdd3316bd15069f28add0151892f37`, SHA-256
`476677c924fd26c24923e82f72c3484c4836659b30e3cde6c82468c708a690a0`.
Unlike earlier chapters, the Chapter17 command directly evaluates its ten checks;
this runner adds no recursive Chapter16 parity stage. Formal comparative review
and final student evidence remain pending; no correction round begins here.

Completion update: read the coordinator's committed `c7a84f9`
[provenance](../../solutions/evidence/ch17/original-grade-provenance.json)
and [full result](../../solutions/evidence/ch17/original-grade.txt).
Actual original score is **0/100 FAIL** on the source/tree above, grader exit1 /
make exit2, completed12:57:45 UTC. All nine runtime checks report startup
failure (fallback stops on its first case); the structural check reports absent
`gui/internal/recall`. Output SHA-256:
`23da91957bf4259cc5376b69932e962d8418a1060397f5424d332c2395dab4f5`.
Thus the missing-type and wrong-module diagnoses are actual failed outcomes;
corpus-path and vendor-message predicates remain unreached limitations. The
0 score is not a measurement of failed BM25/judge behavior. Final acceptance
remains pending complete student submission and independent review.

## October11 source-review probe — capped recall duplicates

During authorized postimplementation comparison, before final submission,
reviewer reproduced one current-scope deduplication defect on the isolated
`fcd091f` source above. The whole-source-content comparison in Recall precedes
formatting; when formatting truncates a snippet, that complete source content
never appears in the conversation. The next turn selects it again and appends
identical capped bytes. This defeats §17.5's reason for checking already-present
snippets and repeatedly spends context on material already delivered.

The real public Agent received two Ask turns through its actor/history/renderers
with only an external HTTP provider fake and mechanical recall enabled. A short
control (435 source bytes,515 formatted) produced one recorded entry and one
recall heading in the second request. The capped case (11,247 source bytes,
6,144 formatted) produced two recorded entries with byte-identical Parts and two
recall headings in the actual second request. Exit1 was the intended assertion,
0.024s total; short control passed. No student source was changed.
[Exact probe](../../solutions/evidence/ch17/reviewer-capped-recall_test.go.txt),
[output](../../solutions/evidence/ch17/reviewer-capped-recall.txt),
[provenance](../../solutions/evidence/ch17/reviewer-capped-recall.json),
and both [short](../../solutions/evidence/ch17/reviewer-recall-short.json)
and [capped](../../solutions/evidence/ch17/reviewer-recall-capped.json)
request/entry captures preserve the result. A subsequent metadata print initially
used `Parts` instead of the JSON key `parts`; correcting that inspection confirmed
identical attachment bytes and did not alter the test or recorded result.

This finding requires correction of already-present capped attachment handling,
with a positive short control and meaningful restored-defect sensitivity. It
does not prescribe the old answer's attribution heuristic, new persistent IDs
or a new retrieval policy. Final source/evidence and initial formal disposition
remain pending; no correction round is consumed by this interim source finding.
The coordinator is separately tracking the student's own pre-submission fixes
for relative Agent/absolute skill paths and request-local judge model headers.

### Comparative preparation and live inspection, awaiting final freeze

Completed source reading of the student's new recall/index/judge code, shared
interfaces, actor worker, Engine request-local judge route, reducer/skill changes
and actual core integration/unit tests. Also read matching original Chapter17
recall package (including its complete tests), common and Engine recall/judge
code, and changed composition/actor/rendering/event paths against Chapter16.
Original generated logs/binaries are support, not production lines; original
removal of the virtual-user example must be separated from recall's own delta.
No original answer code was sent to the student. Final counts/documentation
assessment waits for its pending comment and regression changes.

The student's unified actor begin path reaches actual recall for both public
Ask and Post, without two parallel entrypoint hooks. A required parent pointer
and common interface connect Recall to its owning Agent and Engine; recall owns
the judge's reply format. Its cancellable bounded transport avoids merely
abandoning a still-running judge goroutine. These are substantive improvements
over the original's separate wiring and direct scripted judge test doubles,
not reasons to claim every behavior verified. The current BM25 unit tests prove
query/document filtering, positive smoothed IDF and floor, but do not yet directly
discriminate competing-hit TF saturation/length normalization; final sensitivity
evidence should address those core ranking claims.

Independently parsed the retained actual api.log requests for all three live
runs and read stdout against the fixture facts. Each has one tool-free judge
request and two main requests. Both main requests carry QZ-731 with one recall
heading and no cafeteria distractor; the judge sees both relevant and distractor
material. Each event log contains one recall_attached and two completed response
events. Anthropic/OpenAI/Gemini stdout correctly gives Morgan, October23, the
rejected retry rationale, unresolved ledger placement and nine-day rollback.
These are actual retrieved-request and answer observations, not marker matches
in a judge prompt alone. They precede the final path/override/dedup refinements;
no paid repetition is needed merely because those independent boundaries change.

Coordinator relayed the confirmed cap finding as attributed pre-submission
assistance, while the student was already correcting its own path/override bugs.
It remains initial evidence, not an independently discovered student defect and
not a consumed formal correction round. Preserve the failing probe for a final
source recheck; final acceptance and formal initial disposition remain pending.

The ranking sensitivity concern was subsequently tested with coordinator
authorization in the isolated `fcd091f` copy. Replacing the BM25 contribution
with raw term frequency (retaining the local IDF variable only to compile)
**survived** TestBM25, TestRecallJudge and TestRecallTurnReplayCheckpoint in
both packages. Baseline exits0 (0.005s/0.077s); mutant exits0 (0.006s/0.076s).
This is a test gap for valid masked ranking obligations, not a production-formula
defect or a compiler failure. Correct source was restored.
[Exact mutation](../../solutions/evidence/ch17/reviewer-ranking-mutation.json),
[baseline](../../solutions/evidence/ch17/reviewer-ranking-baseline.txt) and
[survivor](../../solutions/evidence/ch17/reviewer-ranking-mutant.txt)
are retained. Representative competing-hit or independent score evidence must
distinguish real BM25 from plausible raw counts; no old answer fixture or code
was supplied to the student. Coordinator relayed this as further attributed
pre-submission assistance, with final assessment still pending.

## October11 initial formal review — accepted with recorded exceptions

**Accepted with exceptions, zero formal correction rounds.** Reviewed frozen
`a2f0a4f34f41bd6d75ed568a524ed5d055cc62cb`, main tree
`83e2f4cc85eb91ef9f62a54f638baa4d538cacf8`; independently compared all121 tracked
working source files byte-for-byte with that commit. Read the complete final
[student review](ch17.md), final source changes, original chapter and matching
original17/16 solution changes. The cap and ranking findings above were supplied
before final submission and remain attributed reviewer assistance, not independent
student discoveries or erased failures. No unresolved scope, implementation or
teaching-documentation blocker remains. Original22/24 stops remain binding.

The actual original result remains **0/100 FAIL on fcd091f**, not on final source.
I accept the student's missing-primary-type and wrong-module structural exceptions
for the reasons established above. Owner-scoped corpus paths must not become cwd
global data to fit the unreached fixture. Likewise, a separately recorded recall
entry need not become a separate Anthropic wire message: §17.5 permits renderer
placement as a part of the turn, and actual Anthropic requests succeeded with that
placement. These latter two are latent fixture limitations, not additional observed
failed checks. Markdown learnings follow the supplied public corpus; an unspecified
private JSON `long` envelope is not imposed. Bill's final-quota/all-memory rulings
govern over the historical conflicting passages. No valid retrieval obligation,
architecture rule or live-use requirement is waived, and no grader was changed.

### Independent checks and evidence assessment

Replayed the **unchanged** independent capped-recall probe in an isolated archive
of final source, with the real public Agent/actor/history/renderers and only the
external HTTP peer replaced. Short control remains435 source/515 attachment bytes;
capped remains11247/6144. Both now produce exactly one recall entry and one heading
in the second actual request. The previous duplicate is resolved. The full focused
recall/index suite under race detection passes: ensemble16.389s, recall1.023s,
including the real15-second judge cancellation, hint/interrupt control, all-provider
replay/checkpoint, relative skill path, request-local judge header and capped
source edit/replay cases. See [output](../../solutions/evidence/ch17/reviewer-final-checks.txt),
[provenance](../../solutions/evidence/ch17/reviewer-final-validation.json),
and [capped request capture](../../solutions/evidence/ch17/reviewer-final-capped.json).

Applied the **same raw-frequency replacement** that survived earlier to the final
isolated index. All three new competing-document assertions fail for their intended
reasons: common instead of rare, repeated instead of diverse, long instead of
focused. Restored source passes. [Mutant](../../solutions/evidence/ch17/reviewer-final-ranking-mutant.txt)
and [restored control](../../solutions/evidence/ch17/reviewer-final-ranking-restored.txt)
preserve those results; no shared source changed. Student's individual IDF,
saturation and length mutations likewise reach the corresponding ranking failures.

Reviewed the recorded mutation changes and outputs across initial/follow-up,
regression and cap/ranking batches. The short-query survivor, mixed-parser matcher
mistake and wrong-owner-path unintended failure remain honestly distinguished.
Corrected checks fail on the actual missing guard, parser output, missing corpus,
relative path and request header respectively. The hash-propagation mutant fails
the initial missing-identity assertion, so that mutation alone is not an isolated
later replay proof; the real dump/load/second-turn control establishes retention
through replay. Checkpoint removes the whole recall entry, including identity.
The other meaningful mutations cover the valid obligations in the preflight table.
These are representative sensitivity checks, not exhaustive retrieval-quality proof.

The five-module test/vet receipts are consistent with the claimed coverage (the
streaming example has no tests). Independently inspected the live launcher and
retained requests/answers described above: prompts do not supply the target facts;
the target is an archived `.md.retired` file. All three bounded live sessions
demonstrate judged retrieval, correct answers and one retained attachment. Their
source remains fcd091f; final deterministic path/header/cap corrections do not
retroactively change live provenance or justify a paid repeat. No GUI/speech or
broad corpus-quality result is inferred from these CLI sessions.

### Comparative code, comments, tests and KISS

Counts compare each implementation with its own predecessor, counting physical
lines and nonblank lines after excluding comment-only line prefixes. Mixed code/
comment lines count as active; this is a source-size measure, not Go statements.
[Per-file counts](../../solutions/evidence/ch17/reviewer-comparison-counts.json)
record exact source pairs and method inputs. No production moves occurred.

| Chapter delta | Physical | Active | Comment-only |
| --- | ---: | ---: | ---: |
| Student production Go | +825 | +643 | +143 |
| Original production Go | +1876 | +1064 | +684 |
| Student Go tests | +612 | +550 | +39 |
| Original Go tests | +479 | +377 | +68 |
| Original removed virtual-user example, separate | -179 | -128 | -29 |

No JS/CSS/HTML product delta is hidden in these numbers. Student README adds24
lines; five evidence Python launch/mutation scripts total194 physical lines and
are support, not runtime. Original generated logs/binaries are excluded, as are
student captured receipts. Other example modules are unchanged.

The smaller active implementation reflects one turn-entry path, reused provider
transport and straightforward per-source indexing/selection, rather than deleted
explanations or compressed logic. The original has richer prose comments but
duplicated entrypoint wiring and larger selection/judge machinery. Student's
production delta has18.2% comment-only lines among active+comment lines; its six
new files have17.7%. Useful explanations cover ownership, cancellation, query-only
filtering, stable quotas, corpus refresh, recorded-byte ordering and capped identity.
All new exported Go types, fields, methods and constants have useful documentation.
This meets the teaching intent without padding to a numeric target.

Student test growth is justified by real-runtime paths absent from the original's
direct scripted judge tests: actual HTTP transport, Agent/Engine/history, entrypoint
steering, disk edits, recorded replay and all-provider rendering. The small boundary
fake does not substitute for the actor, jobs or retrieval under test. Pure algorithm
tests complement those paths. No unapproved mocks or generic testing framework were
introduced.

Recall has an actual required Agent parent; the composition root attaches both
sides and reaches configuration, skill directory, Engine and logging through owner
interfaces. Common contains vocabulary, not retrieval behavior; implementation
spokes retain the star import boundary. GUI remains an unchanged optional public
consumer. The per-entry hash is proportionate to the demonstrated truncation bug:
it records only represented complete-source identities, changes on edits, replays
with recorded bytes and retires with the entry. It adds neither an external ledger
nor a new retrieval policy. No later capability or grader-compatibility framework
is needed. Reviewer scratch copies are removed after preserving the small evidence;
coordinator owns the direct Edition2 ch16 export/checkpoint. Required handoff:
`.agents/skills/ensemble-coding/SKILL.md`.
