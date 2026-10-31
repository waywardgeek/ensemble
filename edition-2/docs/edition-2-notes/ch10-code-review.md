# Original Chapter 10 independent code review

Reviewer: `enhancement_reviewer_resumed`, October 10, 2026.
Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

**Initial review: accepted with recorded exceptions.** No implementation
correction is required. The evidence-only correction below was incorporated
before acceptance; this does not consume one of the three correction rounds.
This accepts original Chapter 10, not a full original-grader pass or a later
chapter. The coordinator owns the checkpoint and subsequent assignment.

## Scope and source

Read the full skill, current carryover/workflow, original Chapter 10, complete
student submission in [ch10.md](ch10.md), and inherited
[Chapter 9 review](ch09-code-review.md). The chapter remains revision
`e1605916114230570b348ab5961c193fa1cebf80`, SHA-256
`b04d6facb19ae61b78a8246028ac505955bb8c139a710b9aaaaf29f5c4fb3476`.
After submission, read the new skills implementation and all changed runtime
paths/tests, matching first-edition Chapter 10 skills implementation and its
changes from original Chapter 9. Student comparison uses its own accepted
original Chapter 9 bytes in `solutions/edition-2/history/ch09`.

Reviewed runtime is the implementation saved in progress commit `1395ec2`.
The reviewer independently hashed all 57 Go files before scratch validation and
verified they still matched the shared submission afterward. Exact hashes and
commands/results are in
[reviewer-checks.json](../../solutions/evidence/ch10/reviewer-checks.json).
The three generated GUI logs accidentally staged by the coordinator are excluded
from source counts; their subsequent cleanup changes no implementation.

## Actual grading and exceptions

The original full command, `make grade-dir CH=10
DIR=solutions/edition-2/main/gui`, was **interrupted and unscored**, exit 2.
Its [partial output](../../solutions/evidence/ch10/original-grade.txt)
and [termination record](../../solutions/evidence/ch10/grader-interrupted.json)
are retained. The coordinator explicitly bounded its known inherited
GUI/headless timeout cascade; the coder stopped only its owned processes and
verified they were gone before reviewer execution.

Inspection found an important ordering detail: `Ch10Run` invokes recursive
Chapter 9 parity **before** the full-primary skills phase. Stopping the cascade
alone would leave that new phase untested. With coordinator authorization, the
reviewer copied the unchanged grader and dependencies to scratch, added one
test bridge, and invoked all four original skills phase functions plus the
original system-prompt predicate and `Ch10Evaluate`. Only recursive parity was
explicitly skipped. No grader expectation, original file or student source was
changed. Retained [bridge](../../solutions/evidence/ch10/reviewer-original-bridge.go.txt),
[source manifest](../../solutions/evidence/ch10/reviewer-grader-source.json),
[raw output](../../solutions/evidence/ch10/reviewer-original-skills.txt)
and [results](../../solutions/evidence/ch10/reviewer-original-skills.json)
make that distinction inspectable.

The focused original checks earned **65/85**:

| Original check | Actual focused result |
| --- | --- |
| Initial tools | 0/10, only `think` missing |
| Primary system body | 10/10 |
| Full-primary tools | 0/10, only `think` missing; other eleven present |
| Dynamic tools | 15/15 |
| Progressive disclosure | 15/15 |
| Dependencies | 10/10 |
| Variable substitution | 10/10 |
| Unreachable skill refusal | 5/5 |

The six other skills checks pass **65/65**. The wrapper exits 0 because it
successfully collects evidence; that is not a passing grade. Its explicit
skipped-parity entry is not a measured 0/15 result. Do not report this as a
completed 65/100 original run. The recorded first-edition baseline remains
100/100; it was not rerun for this review.

I accept the student's narrow missing-`think` exception. Independent search of
public Chapters 0–10 finds the name in Chapter 10's example but no tool behavior
or argument/result contract. The accepted predecessor has no such handler.
Current-chapter grader phases demand the name without exercising its behavior;
inherited probes do invoke a tool of that name, so the narrower claim must not
be generalized to the whole recursive harness. Advertising an unimplemented
capability would violate declaration/dispatch agreement; inventing an unrelated
tool solely to satisfy the name would not teach the supplied exercise. The two
failed name checks stay failed. This does not waive any installed tool's
visibility or execution requirement, and `tool_limits` remains available when
enabled by a skill. A future exercise that actually defines `think` requires
fresh implementation/review, not perpetual waiver.

I also accept the bounded inherited-harness exception. Source inspection confirms
the recursive chain keeps passing the GUI submission into earlier binary builds
and headless drivers. It does not select this implementation's separate CLI and
external exercise modules. The original Chapter 9 review already established
that mismatch, including its actual 80/100 result and the unknown-model delivery
conflict. Nothing in this skills change modifies provider fallback, settings,
GUI protocol, jobs or the retained external examples. Those required behaviors
remain binding: the coder reran all five real modules' tests/vet, including
settings/budget and public exercise regressions. Prior mutation/browser evidence
is reused for unchanged behavior. The present interrupted run does not establish
which score every inherited subcheck would eventually receive; no such claim or
blanket feature waiver is made.

## Implementation and comparative assessment

The new private registry is a real child of Agent. Construction validates the
required parent, attachment checks both sides, and handlers reach it through
their owning call/Agent. `common` contains values and interfaces, not registry
behavior. Both executables still construct the public Ensemble root; environment
translation stays at their application boundary. Variable render functions are
the chapter's point-of-use extension, not a bundle of sibling services.

Declaration filtering and dispatch use the same enabled-tool predicate. The
whole dependency graph is validated before granting capability, including cycle
and missing-dependency failures. Dependencies contribute instructions as well as
tools; renderers run after activation and outside the registry lock because
`$TOOLS` re-enters declaration lookup. The initial body is frozen; dynamic bodies
travel through actual tool results. Shared initial tools retain their provenance,
and dynamic metadata records the just-appended `tool_called` sequence through
the real actor path. Pending unload keeps instructions/tools until the later
compaction lifecycle; this is consistent with this chapter. There is no premature
compaction, persistence, caching policy, memory or sub-agent implementation.

Compared with the first-edition addition, this avoids behavior in `common`,
parallel command/library assembly, copied declaration updates and extra public
registry machinery. It also enforces dispatch visibility and retains capability
during pending unload. The original registry's partial dependency mutation and
panic-on-cycle path are replaced by one small validation pass with ordinary
errors. These are correctness and ownership improvements, not merely fewer lines.
The flat parser, one renderer map and existing tool/job/history path are an
appropriate KISS implementation. No unresolved scope or bloat finding remains.

Independent counts use physical lines and active lines excluding blanks and
`//`-only comments; inline-comment code stays active. No Go source moved or was
generated. Examples are separate below, rather than credited as core reduction.

| Go category | Own Chapter 9 active / comments / physical | Chapter 10 active / comments / physical | Active delta |
| --- | --- | --- | --- |
| Student production, including optional GUI, excluding three examples | 4,573 / 1,070 / 5,975 | 4,975 / 1,135 / 6,482 | +402 |
| Student examples `ch05`, `ch06`, `ch07` | 415 / 37 / 469 | 415 / 37 / 469 | 0 |
| Student tests across modules | 2,961 / 231 / 3,284 | 3,180 / 237 / 3,518 | +219 |
| First-edition production | 5,862 / 1,872 / 8,535 | 6,573 / 1,999 / 9,475 | +711 |
| First-edition tests | 769 / 160 / 1,012 | 769 / 160 / 1,012 | 0 |

JavaScript/CSS/HTML are byte-unchanged within both editions: student physical
totals 461/165/53; original totals 589/435/108. The original 68-line shell live
helper is unchanged. Evidence launchers and fixture Markdown are support
artifacts, not runtime savings. The student's combined Go production totals
agree with its reported 5,390 active, 1,172 comment-only and 6,951 physical lines.

New production comments are about 14% of nonblank added lines; whole production
is about 18%, compared with the original addition's roughly 15% and whole-tree
23%. Every new exported Go symbol/member has useful documentation. The important
comments explain ownership, frozen text, deferred removal, atomic loads and the
re-entrant rendering lock boundary. This is sufficient teaching explanation for
the small addition; padding to an exact 20% would add no value. The lower line
count does not come from deleting inherited explanations or compressing control
flow.

## Validation and evidence correction

The first edition adds no chapter-specific test code here. The student's two
public-library tests exercise real Agent, registry, jobs, filesystem effects,
history and vendor rendering behind a small external HTTP fake; its CLI test
builds and runs the actual command. These are faithful boundary fakes, not mocks.
They protect the new behavior rather than asserting a substituted internal call
sequence. Existing tests remain intact.

Independently reran the two skills tests with `-race` (PASS, 1.055s), actual
prompt-command test (PASS, 0.214s) and core `go vet ./...` (PASS) in a byte-matched
scratch copy. Inspected all twenty claimed behavioral mutation records and their
intended failure outputs. The earlier dependency mutation's compilation failure
is correctly excluded and preserved, followed by its compiling replacement.
As an additional independent control, changed the engine in scratch to append
current declarations to the system text: the existing test failed twice at
`system prompt changed after load`. Restored the file and reran the focused
baseline successfully (0.012s). The
[mutation receipt](../../solutions/evidence/ch10/reviewer-frozen-prompt-mutation.txt)
is distinct from the coder's twenty mutations. Scratch copied source and bridge
trees were removed after retaining these small records.

Read the three discovery receipts, bounded live launcher and actual captured
chat artifacts. Anthropic `claude-haiku-5-5`, OpenAI `gpt-4o-mini`, and Gemini
`gemini-3.8-flash` each exited 0, loaded instructions, performed marker write/read,
marked pending unload and returned the verified marker. Captured system content
is identical across their respective 5, 4 and 5 requests. No new paid runs were
needed. The launcher keeps credentials in memory/child environment and redacts
before retaining output.

One initial evidence claim needed correction: OpenAI's read occurred after
unload, but `read_file` belongs to the initial base skill. That observation alone
cannot prove the dynamic `write_file` remained enabled. The student has retained
the correction and narrowed the claim in `ch10.md`. The declaration check and
targeted early-removal mutant separately establish dynamic-tool retention;
declaration/dispatch share the same admission predicate. No runtime change or
extra paid demonstration was required. The revised explanation is accepted.
