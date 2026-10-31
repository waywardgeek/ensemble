# Edition 2 Chapter 3: grader enhancement

Contributor: `enhancement_coder_resumed`, October 10, 2026. Initial submission
for independent review; preparation feedback below has not used a formal
correction round. Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.
Read that skill in full, original Chapter 3, the accepted student/reviewer
records and all 35 historical mutation records. Author and coordinator agreed
the narrow plan before implementation. No runtime, original grader, frozen
export or first-edition text changed; no mocks, paid calls, commits or Chapter 4
work.

## Change and source

`grade-edition2 -ch 3` preserves all ten original checks and their weights. It
substitutes corrected Chapter 2 parity, then uses one additional real CLI session
with the existing provider fake to check two explicit §3.3 obligations missing
from the original harness: search context windows and append behavior.

The search fixture checks matching/context delimiters, adjacent-window merging
and a separated group. Expected bytes independently match system
`grep -n -H -C 1 needle notes.txt`; they were not accepted merely because the
student produces them. The append probe plants `original\n`, asks for
`tail\n` with `append:true` and no overwrite flag, then checks both a non-error
result in the actual next model request and exact `original\ntail\n` disk bytes.
The probes use separate files so a write defect cannot conceal a search defect.
Anthropic string and text-block result forms are both accepted. No internal
student package names or public API spelling are required by the grader.

`readtools` retains the original four-point deduction for a broken search tool;
`mutatetools` retains its five-point budget. A narrow helper repeats only the
original search verdict's two text predicates on its first wire result to avoid
charging search twice while preserving deductions for other read tools. Original
checks still evaluate their complete evidence; no human diagnostic is parsed.

The empty-registry invariant remains a library/source-test observation: there
is no new CLI selector. The original policy-neutral `editcontract` remains
unchanged; the student's overlapping-anchor regression documents its chosen
unique-refusal policy, not a universal new grader requirement. Read cap/default
choices and credential filtering retain their existing source tests and reviewed
mutations. The external custom-tool consumer remains in Chapter 4 with old 5.

Original source revision: `e1605916114230570b348ab5961c193fa1cebf80`, Chapter 3
SHA-256 `b83b9e989bb213bc3ac36f092ed80297fb996e5d64d9da1adb7af74bcbd4f3d8`.
Accepted source is frozen `solutions/edition-2/ch03/`, `edition-2-ch03-r3`;
legacy is `solutions/ch03/`. Latest revision touching reused Chapter 3
grader/fake files: `c871073c2f5a3117b5969119345a95e02e617746`.
Historical results remain in [ch03.md](ch03.md), [the comparative review](ch03-code-review.md)
and [original grader output](../../solutions/evidence/ch03/grader.txt).

## Inherited capture-fixture correction

The newly accepted Chapter 2 capture probe supplied one tool-call response.
Chapter 2 records it and returns; Chapter 3 executes it and correctly asks for
another response. The repeating fake therefore hit the sixteen-round limit.
Before correction, the revised Chapter 2 grader on accepted Chapter 3 scored
**85/100 FAIL**, solely `seam-parse`: `tool loop reached 16 rounds`.

Added a terminating text response with zero usage to that new fake. The model,
call and exact-signature assertions remain. This is the already documented
cross-chapter fake-lifetime lesson, not a runtime bypass or an original-grader
edit. Accepted Chapter 2, legacy Chapter 2 and accepted Chapter 3 each now score
**100/100** under revised Chapter 2. The accepted Chapter 2 checkpoint and its
recorded results remain history. The signature mutant below confirms that
termination did not hide missing capture data.

## Actual validation

Built both grader commands, then ran their binaries sequentially with `-ch 3
-json`. Equivalent source invocations are:

```sh
go run ./cmd/grade -ch 3 -json ./solutions/edition-2/ch03
go run ./cmd/grade-edition2 -ch 3 -json ./solutions/edition-2/ch03
go run ./cmd/grade -ch 3 -json ./solutions/ch03
go run ./cmd/grade-edition2 -ch 3 -json ./solutions/ch03
go test -count=1 ./internal/gradeedition2 ./cmd/grade-edition2
go vet ./internal/gradeedition2 ./cmd/grade-edition2
gofmt -l internal/gradeedition2 cmd/grade-edition2
```

| Source | Original grader | Revised grader |
| --- | --- | --- |
| Accepted Chapter 3 | **90/100 FAIL**, inherited media parity only | **100/100 PASS** |
| Legacy Chapter 3 | **100/100 PASS** | **96/100 FAIL**, search context only |

The legacy tool returns only matching lines; it ignores `context_lines`. This
agrees with the prior independent comparison and is an actual missing behavior
under original §3.3, not a new requirement or a reason to add a compatibility
escape. Its append probe and all other revised checks pass.

Affected-package uncached tests and vet pass, including the accepted Chapter 2
and Chapter 3 regressions. Formatting and scoped whitespace checks are clean.
No broad root suite was repeated: its unrelated script-package setup errors and
interrupted run are already recorded in [enhancement-02-code.md](enhancement-02-code.md).

## Targeted mutations and preparation feedback

Each valid behavioral mutation changed an isolated copy of accepted Chapter 3;
replacement occurrence was asserted, both graders ran, the intended check
failed, and scratch source was removed. Original scores stayed **90/100** with
only inherited parity failing. Revised baseline/restored source is **100/100**.

| Defect | Revised observation |
| --- | --- |
| Drop preceding search context (`i-context` becomes `i`) | **96**, `readtools` reports missing context/incorrect groups. |
| Stop merging adjacent windows (`<=` becomes `<`) | **96**, `readtools` reports extra group separator. |
| Omit separated-group `--` | **96**, `readtools` reports missing separator. |
| Use context delimiter for matching lines | **96**, `readtools` reports missing matching-line colons. |
| Replace append-open with truncation-open | **95**, `mutatetools` observes `tail\n` instead of original plus tail. |
| Disable the append branch | **95**, `mutatetools` observes a tool error rather than successful append. |
| Drop parsed Gemini call signature | **90**, `ch2parity` now fails specifically on corrected `seam-parse`, proving capture sensitivity survives the terminating fake. |

Preparation found a score-accounting mistake: an initial whole-group failure
gave the legacy source 90; capping the read score at six then avoided double
charging search but missed simultaneous failures in different tools. Reviewer
feedback led to the small predicate-based accounting above. Two focused source
mutants verify it: ignored read range plus missing search context earns
**2/10 readtools**; an already-broken original search output also failing the
new probe earns **6/10**, with no duplicate charge. The reviewer also caught
first-versus-last replay occurrence; the helper now matches the original's first
result. Prior subgroup summaries are labelled `original probes` when a new
failure is appended, so they do not claim to summarize the strengthened check.
These are grader-preparation corrections, not student correction rounds.

No valid mutant survived and no compile failure counted as evidence. Unchanged
behavior reuses the [35 original mutation results](../../solutions/evidence/ch03/mutations.txt)
and Chapter 2 enhancement evidence. This is targeted sensitivity, not exhaustive
proof. No general mutation framework or duplicate source trees are retained.

## Chapter examples and live evidence

Validated the author's serve-only instructions using the actual `cmd/fakevendor`
binary and an independently built accepted Chapter 3 CLI. Parsed the fake's
printed environment assignments without shell evaluation, launched `chat` in a
disposable project containing `go.mod`, and used its own `CH02_LOG`. Anthropic,
OpenAI and Gemini each exited zero, made **three** requests, returned successful
`list_directory` and `read_file` results, and included the planted module text.
This verifies the changed launch path executes the student's binary rather than
the reference convenience launcher. It is deterministic fake use, not a new
live-provider claim. The owner-interface schematic agrees with accepted
`common.ToolContext`; the argument/default table agrees with the existing tools.

Reused the independently reviewed [live record](../../solutions/evidence/ch03/live.txt):
three-provider failing-test → targeted edit → passing-test work, error recovery,
codename recall, plus the extra search/write/read session exercising all six
tools. Its initial refusals, actual totals and live-build timing limits remain
visible in the student review. No paid run was needed for these grader changes.

## Size and simplicity

Physical Go lines are classified as blank, leading-`//` comment-only, or active
(including imports/braces). New Chapter 3 grader: **135 active / 17 comments /
6 blank / 158 physical**; integration test: **19 / 2 / 3 / 24**. Command dispatch
is **59 / 3 / 4 / 66**, with seven added/four removed lines against Chapter 2.
The capture fake correction adds six lines and removes one. These are grader
support changes; Ensemble production/test counts remain those in the historical
comparison. Comments explain observation boundaries, fixture independence,
cross-chapter termination and score accounting. All new exported symbols have
purpose comments; no new exported fields were introduced.
