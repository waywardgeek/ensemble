# Corrected cumulative handoff sources, new14–21

October 14, 2026 — maintenance carry-forward by the structured-submission coder,
with prior-answer exposure. This is not a fresh student run. The mandatory
`.agents/skills/ensemble-coding/SKILL.md` and full carryover were reloaded after
compaction. Earlier accepted reconstruction and live evidence retain their own
source identities.

A new sparse worktree, `/Users/bill/projects/ensemble-handoff-corrections`, holds
one fresh `codex/handoff-newNN` branch per source. Main, primary and reordered
worktrees were not switched or reset. Each correction starts from its listed
accepted base; old commits, tags and frozen solution directories are unchanged.
All corrected source prefixes are `agent`.

| New chapter | Accepted base and prefix | Corrected commit | Corrected Agent tree |
| --- | --- | --- | --- |
| 14 | `a8eec84a32523987eb6574c7c9c2dd9cdb927275:solutions/ch14` | `82d3952a0267fff0e540238a1326da764b44ea64` | `bccc7d3d5fd7eb4dc14a789ff9f3570c66200400` |
| 15 | `288fc6e8f866bc19d65c17909f270f294c90d1de:agent` | `6edde04092a02b1d708c9ff1b15338a3f9b132b0` | `d7c38fa74bf9cc1a62c60e61a7e5f105feeccd98` |
| 16 | `970676abeebce8f156d5cdc7ee0dd47a40708136:agent` | `ec853c18bc2d7f2e2ab3a89d23067d50a655d628` | `c35c147a692100b325d3b46805e59105b4a4e2c0` |
| 17 | `078cccbd2e6b999b320289aa1d096f7abe394350:agent` | `a4b9a1d33814bb5a660f74ab0b432c9cfea975aa` | `646d69872ae016eb9729b95d0c262eabe98a2fd9` |
| 18 | `9af3864363c1d574f66b44585b35e65c488dfec1:agent` | `6677cb9cb96ea22c1c7dc6d8a1e11e8f3843f706` | `0a4251b225b98a711acbe31bf00672aaa592ae15` |
| 19 | `922caeb4be404c335d209348a9d36aca4f88645e:agent` | `dee87fbcd2dd793e37d820597c513c3556ea7d70` | `c01fcc28a43d0a581be9ed8aa32a5df492dede97` |
| 20 | `1b7bb6521acc11aa43ac1d63c0ea3a8e9afe72b6:agent` | `9f4fdca297d636dfdfd5dfa0f82fe063e8e0e56f` | `1fd589c98da217f8b696f80bec6e7adb38647e12` |
| 21 | `aae1e23043b317d92c8ec84d35340e91902f2ce9:agent` | `3a5673b601dda3934754b13c33da985694ea5dc5` | `34b461a12ee6d080ecf104940635f53f62874e68` |

New14 is seeded from the exact old `solutions/ch14` tree
`0c751dbd3985dee1a80284461e31346471ccc984` into the new maintenance worktree's
`agent`. Its large parent-commit diff reflects that source relocation; comparison
against the listed source subtree has only five changed files. All later
comparisons have at most seven changed files.

## Bounded changes and checks

New14–19 receive the already accepted no-memory cleanup from `1b7bb652`:
remove opaque reasoning parts and private signatures on retained visible parts
at the completed handoff boundary, preserving original log facts. New20–21
already contained it. All eight receive the shared render-only completion label
from `57920dd3`, leaving stored checkpoint notes unchanged. Tool descriptions and
README handoff wording agree; README chapter references use the new current
numbers while original historical references remain explicit. No memory/recall
implementation appears in these sources, and no Responses/auth implementation
or test was imported before new16.

New14–15 extend the existing real Anthropic HTTP survivor test with log/snapshot
replay and render-purity assertions. A small public journal/reducer test checks
text-bound signature removal directly in context while visible text and original
log facts survive. New16–21 use the accepted real Responses/Gemini HTTP regression
from `57920dd3`, removing only its future Recall-disabled configuration field.
New15 onward adjust the exact OpenAI cache-marker expectation for the host label
joined to its original checkpoint text; the other stable/tail assertions remain.

| Source | Mapped explicit-source grade | Focused race, ensemble package | Core module check | Vet |
| --- | --- | --- | --- | --- |
| 14 | 100/100 | 1.987s | full, 10.771s | pass |
| 15 | 100/100 | 2.609s | full, 72.595s | pass |
| 16 | 100/100 | 2.840s | short, 13.797s | pass |
| 17 | deliberately ungraded | 2.747s | short, 13.742s | pass |
| 18 | 155/155 | 2.829s | short, 13.751s | pass |
| 19 | 100/100 | 2.858s | short, 13.974s | pass |
| 20 | 200/200 | 2.977s | short, 15.366s | pass |
| 21 | 100/100 | 2.832s | short, 15.787s | pass |

All commands exited zero. Grading used primary's current mapped
`go run ./cmd/grade-edition2 -ch NN -json` with the explicit maintenance `agent`
path, serially under the shared grader slot. Focused runs used
`go test -race ./ensemble -run 'TestContext|TestMicroHandoff|TestCache' -count=1`
(with only available tests at early stages). Core commands were `go test ./...`
for14/15 and `go test -short ./...` thereafter; `go vet ./...` passed for each.
Short mode intentionally reuses the unchanged real61-second lifetime regression
that passed in new15 rather than repeating it six more times. GUI runtime code
was unchanged; mapped graders build/exercise their required GUI paths. This pass
does not claim a fresh full GUI suite or live-model run for each source.

Raw local outputs are `/tmp/ensemble-handoffNN-{grade,focused,core-short,vet}.txt`;
14/15 use `core.txt`, and final14 uses `focused-final.txt`. The exact source table
also exists at `/tmp/ensemble-handoff-sources.json`. Small scratch mutants were
removed after use. Restoring opaque parts fails the Anthropic survivor assertion;
omitting the label fails its exact-count assertion. Initially, retaining the
text-bound signature survived the early renderer-only check because that renderer
already ignored the field. The repaired projected-context assertion detects it:
`/tmp/ensemble-handoff14-mutant-text-signature-fixed.txt`. The initial survivor
and other two intended failures remain in `...mutant-{text-signature,label,opaque}.txt`.
Reuse accepted provider-specific mutation evidence for the unchanged16+ regression.

The first early test compile failed because the new fixture used `HandoffData`
instead of the existing `MicroHandoffData`; `/tmp/ensemble-handoff14-focused.txt`
retains that setup failure. A trailing README blank line initially stopped the
new14 precommit whitespace check and was removed before commit. Neither is
counted as behavioral evidence. Exact subtree comparisons admit only the listed
README, context/handoff/cache tests, reducer, tool description and renderer files.

Independent narrow review has accepted all14–21 corrections, checking actual
source differences and positive outputs against their accepted bases. Export installation and final default-path grader checks
remain coordinator work after acceptance. The coordinator will switch the eight
current `internal/gradeedition2/ch{15,16,18,19,20,21,22,23}_test.go` source paths
from temporary `../../agent` to their matching installed `../../solutions/chNN`
(including GUI suffixes), then run the final combined package checks. No export,
tag, artifact rebuild, publication, deployment or push is performed by this pass.
