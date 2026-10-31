# Edition 2 Chapter 9 independent enhancement review

Reviewer: `reviewer_astra`, October 11, 2026. Mandatory skill:
`.agents/skills/ensemble-coding/SKILL.md`.

**Accepted at initial combined prose and grader review.** Zero formal correction
rounds used. This accepts Edition 2 Chapter 9, original Chapter 10. The coordinator
owns checkpointing and release of the next chapter; historical results remain
unchanged.

## Reading and prose assessment

Read the full current workflow (1,784 lines), carryover (871), voice (632) and
chapter procedure (599) into active context, plus AGENTS, the complete skill,
enhancement plan and chapter map. Read all original Chapter 10, its complete
student/coordinator notes and comparative review, and enhanced Chapter 8 for
continuity. Full rereading after any compaction remains mandatory.

Reviewed the complete draft and author record at `e533aa35`, against accepted
`solutions/edition-2/ch09` registry, parser, public construction/variables,
tool handlers and real-loop tests. The original section progression and fixed
constitution explanation survive. There is no original personal anecdote to
invent; the reader's expanding capability set supplies the concrete problem.
The added partial-load example, guessed writer, rendering lock explanation and
unload-evidence correction make the consequences understandable. The chapter
enhances the original rather than replacing it with a workflow specification.

The undefined `think` declaration becomes the already implemented `tool_limits`;
the complete twelve-tool list, per-Agent ownership, `name` arguments, environment
boundary and separate CLI/GUI paths agree with accepted source. Dynamic manuals
remain tool results, the initial prompt stays fixed, and unload remains pending.
Provenance motivates later caching without claiming a cache grant. No compactor,
memory service, cache policy or new runtime feature is introduced. Exact public
snippets and revised grading claims received final validation below.

Independently inspected the retained live launcher and all three raw receipts.
The actual skill graph matches the demonstration, with the disclosed added
descriptions and body wrapping. Claude/OpenAI/Gemini have 5/4/5 requests with
nonempty, identical system content within each session, normal exits, the stated
model identities, matching file markers and actual load/write/read/unload calls.
OpenAI unloads before reading, but that reader was initially enabled: the text
correctly preserves the limitation rather than claiming dynamic-tool retention.
The launcher uses chat through pipes; no personal PTY or listening claim appears.
These are historical receipts, not current model availability or new paid runs.

Direct `go run ./cmd/lintprose book/edition-2/chapter-09.md` passes every hard
check: 2,275 prose words, only the soft lower-length warning. Useful explanation
is sufficient for this scope; padding to the warning threshold would not improve
the chapter. Historical full grading remains interrupted/unscored, and the
separate unchanged Skills phases remain 65/85, never a completed 65/100 result.

## Final grader and sample review

Read the complete coder submission [enhancement-09-code.md](enhancement-09-code.md)
at `c4f21b4f`, author reconciliation `31242d48`, all new grader/test code, command
dispatch and shared-launcher changes, original Chapter 10 predicates/fixture,
and matching legacy registry/variable implementation. Manuscript SHA-256 remains
`804d37e6449f38421c0ad26b09835ee3301fc9caed0bff0b307f08ccc34d1393`.
Independently verified all four submitted source hashes against final validation,
the manuscript hash against the sample receipt, and every chapter-local link.

The eight original skill claims and 85 points remain, with 15 for corrected
Chapter 8 parity. Removing undefined `think` and including installed
`tool_limits` corrects the supplied contract without inventing a handler.
Actual GUI prompts exercise skills; earlier headless checks use the core binary.
The shared launcher's only behavior change accepts additional child environment
entries, leaving existing calls unchanged. Neither a source-layout requirement
nor a new product protocol was added.

The new checks compare exact, order-independent declarations, correlate results
by call identity in the next provider request, and inspect actual file contents.
Different hidden/enabled/pending paths prevent a later valid write from masking
an earlier forbidden effect. The frozen prompt and current expansion checks
observe actual instruction text, allowing both Anthropic system representations
and formatting within fixture-authored variable boundaries. Dependency body
delivery and current discovery are observed rather than inferred from a final
answer. Fresh processes supply independent initial identities. The provider fake
does not replace the registry, actor, jobs, filesystem or renderer.

I accept the coder's fixture correction: descriptions are already present in the
original published format and fixture. Omitting them prevented legacy startup;
adding valid metadata fixes that setup mistake without excusing runtime defects.
Final legacy failures are independently supported by source: hidden dispatch
consults installed handlers, skill bodies render before activation, and dynamic
load returns only the requested body rather than its dependency instructions.
Its tool expansion is correspondingly stale/incomplete. Passing pending writes
remain passing even though another registry helper excludes pending state; the
review does not substitute a predicted source outcome for the measured path.

Final revised totals are **composed 100/100 PASS accepted and 40/100 FAIL legacy**:
actual final skill reruns (85/85 and 40/85) plus unchanged measured parity from
the initial full runs. Initial full results were accepted100/100 and legacy0/100;
the latter's skill component was blocked by the description omission. These are
not fresh full final-source runs. Missing legacy application setup and inherited
media/model limits do not prove unexecuted behavior. The report retains those
limits alongside genuine settings/replay/boundary differences. Original legacy
100/100 and the student's interrupted/full versus focused65/85 remain historical.

Both extracted Go fences compile in an outside module; the actual Agent loop
observes renderer replacement while preserving primary text. The exact three
published skill files drive the built CLI through five requests and leave
`SKILLS_OK`; full-primary and first-frontmatter examples also have successful
real-runtime checks. The first example explicitly supplies its companion files.
These receipts are deterministic sample validation, not a new paid or human
session. No source/runtime change was needed to make the examples work.

## Test sensitivity, comparison and disposition

Inspected all seventeen mutation patches and complete failure result arrays;
independently matched every raw JSON array to `mutations.json` and checked output
outside the arrays for compilation/cleanup failures. All seventeen compiling
defects lose the intended check. Broad declaration mutations fail coupled checks;
that is not seventeen independent proofs. The bridge's exit0 means collection,
not mutant acceptance. Its metadata-only earlier fixture provenance is explicit;
the final correct baselines were rerun. No survivor is relabeled, and the legacy
description setup error is not counted as a kill.

An additional independent control ran the unchanged final skills checks against
an isolated, byte-matched accepted core/GUI. Baseline passes (Go test0.513s).
Changing only the pending writer to return success after writing empty bytes
fails **only `load-skill`**, at the actual `pending.txt` content assertion
(0.362s). Restoring the original file passes (0.316s). This proves that successful
results and declarations alone cannot satisfy the new pending-write check.
The retained [receipt](../../solutions/evidence/ch10/enhancement09/reviewer-validation.json),
patch and complete outputs preserve the source hashes and exact bridge. Scratch
copies were removed; no frozen source or shared grader was edited.

Affected grader/command tests, vet, formatting and real skills race checks pass
in the coder evidence; only the focused independent control above was repeated.
Cycles, atomic failed loads, first-enabling provenance, wrong types, Agent
isolation and changed custom renderers retain the actual accepted source tests
and original targeted mutations. The numeric score does not purport to prove
those obligations. The horizontal-rule mutation specifically proves preservation
of the following body text; exact separator retention additionally has the
existing source assertion. No provider rerun or exhaustive audit is warranted.

Independent line counts match the submission: new grader **261 active / 25
comment-only / 297 physical** Go lines; test **16 / 2 / 20**. Original preserved
checks/launcher total **828 / 104 / 1,091**. Count nonblank, non-`//`-only lines as
active, keeping inline comments active; no block comments affect these files.
Command delta is +12/-10 physical; shared launcher +2/-1. This is a separately
versioned grader, not a claimed runtime deletion. Its roughly8.7% production
comment share is below the teaching target but the comments explain correlation,
file effects, lifetime, observation limits and serialization alternatives.
The new exported entry point has a useful contract comment. The direct sequential
driver reuses existing helpers; no framework or padding is justified.

Runtime comparison remains the accepted [original Chapter 10 review](ch10-code-review.md):
against each own predecessor, student production grew402 active/65 comment-only/
507 physical Go lines versus legacy711/127/940, with student tests +219 active
and legacy no new chapter-specific tests. The actual gains are owner-local state,
shared declaration/dispatch admission, atomic graph validation and useful tests;
counts alone establish none of them. Original explanatory comments remain, and
this enhancement changes no runtime, frozen test, first-edition chapter or grader.

No blocking prose, scope, coverage, architecture or simplicity finding remains.
The concrete teaching improvements and corrected grader claims agree with the
accepted solution, including their limits. Chapter 9 is accepted; no later
chapter has been reviewed by this disposition.
