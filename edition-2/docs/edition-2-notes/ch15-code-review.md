# Original Chapter 15 / Edition 2 Chapter 14 code review

Independent reviewer `enhancement_reviewer_resumed`, October 10, 2026.
Initial formal review of submitted `4cd143c`, extending accepted original14
`9d4b977`. Read the complete mandatory
[ensemble-coding skill](../../.agents/skills/ensemble-coding/SKILL.md), governing
carryover/workflow, original Chapter15 (revision/hash in [student notes](ch15.md)),
the completed student review, changed implementation/tests and matching original
`solutions/ch15` against its own `solutions/ch14` predecessor. No old answer
code was supplied to the student. Authoring and subsequent chapters remain stopped.

**Decision: needs correction. One confirmed blocker; correction round 1 is
released, of at most three after this initial review.** The two requested grader
exceptions are justified, but do not waive valid provider wire behavior.

## Initial finding

**Inline tool changes plus automatic ephemera produce invalid Anthropic message
placement.** In `internal/engine/anthropic.go:43–66`, a declaration entry becomes
a contentful `system` message. The later ephemeral entry becomes a new `user`
message. A real Agent, actual MCP transport/registry and automatic-context path,
with only the external peer and HTTP provider faked, captured these roles:

```
user, assistant, user, system, user
```

The system entry contains the actual newly connected MCP tool declaration.
Anthropic's [current placement contract](https://platform.claude.com/docs/en/build-with-claude/mid-conversation-system-messages)
requires a contentful mid-conversation system message to follow a user turn and
either end the messages or precede an assistant turn. This request violates that
contract; the documented result is HTTP400. This is required inline support on
the course fixture, not a request to enable the feature on real model rows.

The first supplemental probe passed prefix equality, MCP declaration presence
and replay equality because its HTTP fake accepted arbitrary request JSON.
Checking the actual placement exposed the defect: `inline system at 3 followed
by user`. Preserve this correction to the evidence claim. Fix the renderer's
valid ordering while retaining automatic context, tool-result correlation,
frozen tools/system bytes and recorded history. Add a faithful boundary check
and demonstrate a restored-defect failure; no paid retry is needed.

[Probe receipts](../../solutions/evidence/ch15/reviewer-probes.json),
[exact supplemental test](../../solutions/evidence/ch15/reviewer-inline-probe.go.txt).
All reviewer executions ended and the scratch copy was removed. Browser/grader
ownership returns to the coder for this bounded correction.

## Actual grader result and exception assessment

The unchanged original grader completed **45/100 FAIL** at `165ee52`, exit1
(make exit2). Handoff15, keep/stub10, crash recovery15 and total reducer5 passed.
The final browser edit guard and test/comment changes postdate that binary; no
final-source original grade is claimed.

- **Settings path:** accept the proposed exception for the three affected checks.
  The harness writes launch-cwd `settings.json`; the existing GUI reads its
  Agent's `DataDir/settings.json`, default `cr/settings.json`. Preserving actual
  Agent ownership is correct. Skill-survival stops when the first result remains;
  restart stops on that same original content; snapshot/log comparison stops
  before comparison because the session contains no redaction. Those failures
  do not independently establish faulty replay. Nor do they prove all later
  predicates passed.
- **Undefined `think`:** accept the missing-contract exception, consistent with
  the earlier chapters. The fixture grants the name without defining its schema
  or behavior. Prefix bytes already pass its comparisons; it then fails on
  `think` before testing the MCP declaration. A dummy compatibility tool is not
  justified. The independent MCP probe checks the masked valid obligation and
  exposed the separate placement finding above.

The real target20,000 integration exercises both bands, initial full presentation,
old paired-traffic removal and surviving skill instructions. I extended that
same real runtime in scratch through a handoff and compared snapshot and log-only
rendering after target900,000; the race run passed. Thus its context includes
skill, cut and handoff events together, unlike treating the original check's
name as evidence. The recorded-cut policy remains outside rendering.

An independent real GUI WebSocket patch test follows `context_target` through
the public Agent into captured subsequent HTTP requests: target20,000 redacts
old results, while zero/default preserves them. It passes alongside the actual
settings wire/disk test under race. This complements the coder's real-browser
save/clamp/zero/two-tab controls rather than mistaking an echoed number for policy
execution. See [GUI probe](../../solutions/evidence/ch15/reviewer-gui-policy-probe.go.txt).

## Comparison, ownership and teaching

Counts use tracked files, physical lines and nonblank/non-comment-only active
lines; comment-only lines use leading Go/JS/block/HTML comment markers. Totals
and method are in [reviewer counts](../../solutions/evidence/ch15/reviewer-counts.json).
These are each edition's own chapter deltas, not a comparison of unrelated totals.
No production move or split changes the counts in this chapter.

| Category | Student delta physical / active / comments | Original delta physical / active / comments |
| --- | --- | --- |
| Production Go | +619 / +532 / +60 | +1162 / +784 / +309 |
| Go tests | +342 / +317 / +17 | 0 / 0 / 0 |
| Production JavaScript | +13 / +11 / +2 | +4 / +4 / 0 |
| HTML | +7 / +7 / 0 | +5 / +5 / 0 |

Student examples, browser tests, CSS and package-lock/support files are unchanged;
original deletion of a generated log is excluded. Evidence launchers are support,
not product code. The new tests exercise real actor/jobs/history/tool paths;
the original snapshot adds no source tests, relying on its external grader.

The student's smaller policy uses complete batches and recorded sequence cuts;
its history spoke owns reduction and disk ordering. Engine accesses its Agent,
and Agent its Ensemble, without the original's target callback or common-package
runtime behavior. GUI settings stay in the optional public consumer and update
Agent configuration. No mock, future memory/compressor, generic framework or
parallel dispatch was introduced. The additions are proportional to this exercise.

New exported fields/methods explain their contracts, including empty startup sets,
held survivors, captured references and checkpoint order. New Go comment share is
about10% of net active-plus-comment growth, versus about28% in the original;
the full student Go production share is about17%. This is below the approximate
20% teaching target, but the substantive ownership, ordering and lifetime reasons
are present beside the non-obvious code. Routine branches do not need padding.
The original offers more extended explanation but also places behavior in common
and uses wiring patterns Bill forbids. The placement defect must be corrected;
there is no separate bloat or documentation blocker.

## Validation and honest limits

Coder module test/vet and core/GUI race receipts are complete. The new context
checks, journal failure control and independent supplemental scenarios passed
under race in scratch before the stricter placement assertion exposed its failure.
The GUI policy/settings race checks also pass. Existing test changes follow this
chapter's explicit manual-survivor, immediate executable-unload, known-event skip
and address-free stub rules; they are not blanket deletion of failing coverage.

Reviewed the twenty-two distinct intended defect checks, including the separate
GUI edit-race mutation. Compile/setup failures remain failed setups; the broad
redaction mutant that failed an earlier assertion remains unsuitable evidence
for the narrower removal claim. Its corrected removal-only probe reaches the
intended assertion. Mutations establish those sensitivities, not completeness:
the permissive request fake missed the placement defect found here.

Journal append precedes live publication; checkpoint installs its snapshot before
journal truncation and retains the previous snapshot. Actual original SIGKILL
coverage and the real filesystem failure control support the specified process
crash behavior. No power-loss/fsync guarantee or exhaustive crash-point coverage
is claimed. Snapshot/log-only comparisons measure vendor bytes, not a mandated
canonical disk format.

I checked the retained raw live events and actual provider tool-result blocks,
not markers in call arguments. Gemini3.8 completed the task with five result cuts,
two tool cuts and one handoff. Both corrected Anthropic/OpenAI runs delivered all
six codes in real result content before six result cuts and one tool cut.
Anthropic subsequently denied having seen old redacted results and refused the
handoff; OpenAI repeated handoffs until the configured bound and produced no
successful final answer. The earlier poor-path fixture and OpenAI attempt remain
failures. All clients exiting zero is not task success. No more paid runs are
required to erase those model-quality limitations.

The explicit exercise's feature-row matrix governs the contradictory later
real-Opus narrative. Existing real rows remain false; course-fake inline/keep
proof is distinct from live ladder/handoff evidence. The current documented
inline definition/header shape is supported, but neither the documentation nor
the fake establishes successful live inline use. Acceptance remains pending the
single renderer correction above; actual45/100 and these live limits will remain.

## Correction round 1 — accepted with recorded grader exceptions

Reviewed frozen `80b3e09`, the full appended student response and
`round1-checks.json`, `round1-mutation.json`, `round1-source.json`.
Independently verified all108 source hashes. **Accepted after one correction
round**, with the two specific original-grader exceptions above; there are no
remaining confirmed findings. Original **45/100 FAIL** and all recorded live
outcomes remain unchanged.

The renderer now retains declaration-block order while placing those blocks
after the complete user/tool-result/ephemeral turn, before the next assistant
or at the request end. This is a small rendering correction: event chronology,
reducer behavior, Agent policy and automatic-context lifetime are unchanged.
The added comments explain the provider restriction and correct the older
blanket statement about system instructions. No new runtime abstraction or
feature was introduced; the production change remains localized.

The **exact independent strict probe** that failed initially now captures
`user, assistant, user, system` and passes under race. The coder's real Agent/MCP
regression also passes independently under race, checking both sides of the
system boundary, actual echo results, connect/disconnect deltas, one-request
round/turn ephemera, frozen prefix and byte-identical replay of each captured
request before its consumption marker. The existing skill/handoff prefix test
passes alongside it. Final focused command: three tests, exit0,
`ok ensemble/ensemble 1.214s`.

The retained restored-renderer mutant fails the intended placement assertion and
HTTP400 path, not compilation or unrelated cleanup. All twelve recorded final
module test/vet/race commands exit0. No paid, browser or full original grader
repeat was needed for this renderer-only fix. The initial permissive probe's
false confidence remains part of the record.

[Independent round1 receipt](../../solutions/evidence/ch15/reviewer-round1.json)
records source verification, the small metadata-helper setup mistake and cleanup.
All reviewer processes ended; scratch was removed. Coordinator owns the accepted
checkpoint and direct Edition2 `ch14/` export. No subsequent chapter is released
by this review itself.
