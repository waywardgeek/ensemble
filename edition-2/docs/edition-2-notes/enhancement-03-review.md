# Enhancement review: Chapter 3 — accepted

Reviewer: `enhancement_reviewer_resumed`, October 10, 2026. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`, read completely before this review.
The current voice/procedure, carryover, workflow and bounded enhancement plan
were read for this resumed author pass.

**Accepted at initial formal combined review; zero correction rounds used.**
Final author and coder submissions are reconciled and independently checked.
No unresolved prose, code, scope or simplicity blocker remains. Final manuscript
SHA-256: `80212731630fc199656227e20a6cf0539a43b851bb11b3ebd17aa5d137891ff1`.
Preparation history below is retained; its pending statements describe those
earlier stages. Coordinator checkpointing remains separate from this verdict.

## Source reading and scope

Read all 739 lines of original Chapter 3, complete student and comparative
reviews, live observations and 35 mutation records. Inspected the accepted
tool implementations, registry/Agent loop, common tool contracts, composition
changes and tool integration tests, matching first-edition tools, and original
Chapter 3 grading interfaces/checks. Original manuscript hash is
`b83b9e989bb213bc3ac36f092ed80297fb996e5d64d9da1adb7af74bcbd4f3d8`.
The accepted snapshot is `edition-2-ch03-r3`; its actual original result remains
**90/100 FAIL**, with all nine current-chapter checks passing and only inherited
media parity failing. No runtime modification or new paid demonstration is
needed merely to restate those accepted facts.

The chapter's original argument is worth retaining: the measured six-tool
concentration, dedicated affordances beyond a shell, edit-versus-rewrite and
anchor-composability lessons, generous fakes, actual exit status versus wrapper
stderr, and the demo that wrote into the book repository. Exact percentages
are attributed historical corpus claims, not fresh second-edition measurements.
The new edition should supply the missing tool inputs, final-response behavior,
per-Agent registration/owner access and library-only empty-registry obligation
without bringing the external-consumer/custom-tool exercise forward from the
agreed Chapter 4 placement.

The retained live record distinguishes real calls from the initial model
refusals and records failing/passing tests, exact edits, errors and codename
recall on three providers. The additional Anthropic search/write/read session
accounts for the remaining two tools. Its source-binding limitation (live build
preceded overlapping-anchor and empty-registry fixes) is explicitly preserved,
with deterministic evidence for those unexercised paths.

## Preparation feedback

1. **Working directory is not confinement.** Sent to author: the original final
   section says the launch directory is the whole scope of what tools can see
   and change. Current tools accept absolute/parent paths, and the shell has no
   confinement. Describe it as the base for relative paths, with confinement
   deferred. This corrects prose; it requires no new security feature.
2. **Inherited fake must finish the new protocol.** Sent to coder: the new
   Chapter 2 returned-identity probe currently supplies only a tool-call reply.
   Chapter 3 correctly answers that call and asks again, so parity reuse needs
   a terminating fake reply rather than a production bypass. This is the same
   documented fake-lifecycle lesson already in the student review. Preserve
   the capture assertions and rerun affected Chapter 2 checks if its corrected
   grader fixture changes.

These are preparation findings while author/coder work, not a completed formal
initial review or correction round. Reviewer owns only this review record and
will independently check the final submission after the coder releases the
grader harness. No code, original source, frozen export, commit or tag changed.

## Initial submitted prose assessment

Read the entire enhanced chapter and its source-relative diff, and the author's
`enhancement-03.md`. No prose blocker found. The chapter preserves its original
progression and incidents instead of replacing the teaching with a specification.
The shell-only objection, anchor failure, generous fake and unwanted codename
remain concrete reasons for the mechanisms. The student's failed initial error
prompts and actual repair add useful personality without invented dialogue.
Owner access, per-Agent tool sets, final-only caller text, capture-time identity,
argument names and the working-directory correction agree with accepted source.
The overlapping-anchor refusal is explicitly a student policy, not a new rule
for every coherent edit implementation. The external-consumer exercise remains
in Chapter 4; no jobs, permission framework or empty-registry CLI mode is added.

Independently parsed the four retained live JSON logs: the three repair totals
and additional search/write/read total match the manuscript, as do actual tool
calls and failure results. The original 90/100 FAIL and pre-correction live build
are represented honestly. Prose lint passes all hard checks: 5,902 original and
7,000 enhanced prose words; the inherited long person-gap warning remains
(2,267 original, 2,240 enhanced), accepted under the preservation constraint.
An optional transition suggestion was sent to the author to distinguish the
student's overlapping-anchor example from CodeRhapsody's general ambiguous-edit
incident. It is not an acceptance condition.

## Grader preparation feedback after first source read

The supplemental search fixture has independent support: running system
`grep -n -H -C 1 needle notes.txt` on its exact planted bytes produces the
expected ten output lines, including the merged adjacent windows and `--` gap.
The append assertion observes actual disk bytes independently of tool narration.
Both contracts already appear in original §3.3; no extra feature is being graded.

Before final coder submission, reviewer and coordinator independently noticed
that capping `readtools` at six points could undercharge distinct simultaneous
failures: a broken range reader plus broken search context could remain six
instead of two. Original scoring deducts four per broken tool. Coder agreed to
check whether the existing search result was already charged, using the same
two original observable text predicates, then deduct/clamp only when needed.
Requested combined-defect and duplicate-search mutation evidence. No original
grader edit, generic scoring framework or parsing of diagnostic prose is needed.
This remains preparation feedback, with final code/evidence review pending.

## Final code assessment and independent verification

Read the final `ch03.go`, integration test, command dispatch and Chapter 2 fake
diff, the complete coder report, raw original/revised baseline and seven mutant
report pairs, both scoring cases, and all three fake-chat transcripts. The
small predicate helper now uses the first matching wire result, agreeing with
the original check. Distinct read/search failures earn 2/10; an already-broken
search earns 6/10, without charging the same tool twice. Preparation findings
are resolved, and no diagnostic-text parser or scoring framework was added.

The grader reuses all ten original checks and weights. The one extra fake
session observes the real CLI, protocol, tools and disk, with independently
planted files. It accepts Anthropic string or text-block results and checks the
append bytes independently of success narration. The terminating Chapter 2 fake
reply changes neither original graders nor capture assertions; the signature
mutant still fails the intended parity check. The legacy search failure is
valid: original §3.3 explicitly requires context windows and grep formatting,
and the old tool implements neither context argument nor windows. I accept the
coder's explanation and reject a compatibility escape to manufacture 100.
The unchanged policy-neutral edit check still permits every coherent original
choice. No mocks, new runtime capability or future exercise appears.

Reviewer commands/results:

- Uncached accepted-source tool integration tests pass: range/cap, exact search
  groups, append/refusal, actual shell with four dummy keys and an ordinary
  environment positive control, per-Agent declarations/isolation, round bound,
  missing Gemini call IDs, and empty-registry error/capture provenance.
- `go test -count=1 ./internal/gradeedition2 ./cmd/grade-edition2` passes,
  including accepted Chapter 2 and Chapter 3 grading at 100/100 and all ten
  Chapter 3 checks. `go vet` on those packages passes; formatting and scoped
  whitespace checks are clean.
- `go run ./cmd/grade-edition2 -ch 3 -json solutions/ch03` independently
  reproduces **96/100 FAIL**, solely `readtools` at 6/10 for missing context.
- In an isolated accepted-source copy, changed the append branch's `O_APPEND`
  to `O_TRUNC` exactly once. Revised grading gives **95/100 FAIL**, solely
  `mutatetools`, observing `tail\n` on disk instead of `original\ntail\n`.
  The mutant compiled and ran the actual path; reviewer scratch was removed.
- System grep independently confirms the new search fixture's exact expected
  bytes. Hard prose lint passes; the author's final paragraph move was read
  and changes no contract, story content or live claim.

The coder's four baseline reports preserve the distinctions: accepted original
90/100 FAIL versus revised 100/100 PASS; legacy original 100/100 PASS versus
revised 96/100 FAIL. All seven targeted source defects fail their intended new
observations, with original mutant scores remaining 90 for inherited media
parity only. Earlier unchanged behavior reuses the 35 historical mutations.
The new launch instructions have three real serve-only-fake/own-binary chat
receipts; the retained provider runs remain dated live evidence. No new paid
call or broad root test was necessary. The unrelated root script-package setup
failures and interrupted broad run recorded in Chapter 2 are not called passes.

## Comparative size, comments and KISS

Independently recounted physical Go lines as blank, leading-`//` comment-only,
or active (including imports/braces/schema literals), separating tests and GUI.
The historical comparison is reproduced exactly: original Chapter 3 production
2,373 active / 848 comments / 3,536 physical, a +521 / +197 / +774 change;
student core 1,815 / 300 / 2,223, a +395 / +51 / +458 change. GUI production
remains 8 / 3 / 13. Original tests add 46 active / 8 comments / 57 physical;
student core tests add 226 / 4 / 239, with GUI tests unchanged. No product line
changes in this enhancement, so these remain the relevant runtime comparison.
Per-Agent storage, current-response dispatch and direct typed tools are clear
improvements; the smaller total alone is not the reason to accept them.

New grader support is 135 active / 17 comment / 6 blank / 158 physical lines;
its integration test is 19 / 2 / 3 / 24. Dispatch adds three active lines net,
and the existing capture fixture adds three active and two comment lines net.
No source is moved, generated or compressed to improve the count. The grader's
11.2% comment share is below the roughly 20% teaching target, but its comments
explain the important observation boundaries, fixture independence, inherited
termination and partial-score accounting beside those mechanisms. The remaining
code is direct decoding and comparison; padding would not improve teaching.
Both new exported symbols have useful documentation, with no new exported
fields. Frozen historical documentation gaps are not silently claimed fixed.

The author took the optional transition suggestion, preserving the original
ambiguous-anchor incident before the student's overlapping-anchor example.
Final prose remains engaging and understandable, with technical additions
attached to observed problems. Original text, graders, frozen exports and
actual results are unchanged. The combined enhancement is accepted; the
coordinator may checkpoint it and release the next mapped chapter.
