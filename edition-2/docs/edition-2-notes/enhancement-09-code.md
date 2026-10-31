# Edition 2 Chapter 9 grader submission

Coder: `edition2_grader_coder`, GPT-6 Astra/high, October 11, 2026.
Initial submission; independent review pending. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

Read the full skill, AGENTS, carryover, workflow, enhancement plan/map, original
Chapter 10, complete ch10 student/comparative review, original grader and relevant
accepted/legacy skills paths. Original chapter revision remains `e1605916`, SHA-256
`b04d6facb19ae61b78a8246028ac505955bb8c139a710b9aaaaf29f5c4fb3476`.
Accepted implementation is direct `solutions/edition-2/ch09`, original10 checkpoint
`edition-2-ch10-r2`; legacy is `solutions/ch10`. No runtime, frozen export,
original grader or first-edition prose changed. No mocks or paid calls were added.

## Scope and contract

Root and author agreed the original eight skill weights (10/10/10/15/15/10/10/5)
plus 15 for corrected Edition 2 Chapter 8 parity, total 100. The command is
`go run ./cmd/grade-edition2 -ch 9 [core-directory]`; default core is
`solutions/edition-2/ch09`. Existing GUI/model and inherited consumer/workflow/
media/streaming flags retain their meaning. The reusable GUI launcher gains only
per-child environment arguments; earlier call sites are unchanged.

The new fixture removes undefined `think`, includes inherited `tool_limits` in
the full twelve-tool primary, and supplies descriptions already documented in
original frontmatter. Every skill phase enters through the actual WebSocket
prompt path. This corrects the original full-primary stdin assumption and keeps
headless parity on the separate core executable. No new public protocol or
implementation naming convention is required.

The original predicates were short, but two hardcoded `think`, others accepted
presence/counts or error words, and their private launcher used the wrong client
for recursive grading. The small new driver retains the eight valid behavioral
claims and weights rather than copying those defective predicates or fabricating
results to satisfy them. It reuses the existing provider fake, GUI process/peer,
correlated Anthropic-result reader and report type.

One real seven-request conversation tries a hidden write, loads code-tools,
writes a different file, loads search-tools and its helper, marks code-tools
pending unload, then writes a third file. Independent fresh processes exercise
full-primary startup and an unreachable skill. Checks establish:

- Exact declaration sets (order-independent, including absence/duplicates) and
  a real hidden-write error with no file effect.
- Rendered primary text, preserved text after a Markdown horizontal rule, and
  equal instruction text across dynamic activity.
- Correlated nonerror load instructions, enabled write, pending-unload result,
  retained declarations and another actual dynamic write after unload.
- One-level `$SKILLS` at creation and after each load, excluding loaded names and
  hidden dependencies; dependency tools and instructions in the load result.
- `$TOOLS` matching each expansion's current capabilities and the custom value
  absent before its skill, then substituted in that skill's result.
- Unreachable load returns a correlated nonempty error and grants no declaration.

Fixture text marks the boundaries of variable expansions; it does not require a
renderer to use a particular format inside them. The fake supplies provider
replies only. All submitted parsing, ownership, tools/jobs, file writes, history
and rendering execute normally. File paths differ so a later valid write cannot
conceal an earlier forbidden effect.

## Actual results and fixture correction

Evidence is [enhancement09](../../solutions/evidence/ch10/enhancement09/).

| Run | Actual result |
| --- | --- |
| Historical original full student grader | Interrupted/unscored, exit 2; no aggregate numerical result |
| Historical unchanged focused student skills phases | 65/85; only two undefined-`think` declaration checks fail |
| Historical original legacy baseline | 100/100 PASS, retained rather than rerun |
| Initial revised accepted full run | 100/100 PASS, 15.538s |
| Initial revised legacy full run | 0/100 FAIL, 25.786s; skill phases blocked by my fixture omission |
| Final revised accepted skills | 85/85 |
| Final revised legacy skills | 40/85 |
| Final composed accepted / legacy totals | **100/100 PASS / 40/100 FAIL** |

The first fixture omitted descriptions. Accepted parsing permits their absence,
but the legacy parser requires them and therefore started no skill conversation.
This was a grader setup error, not proof that legacy skills failed. All original
published skill fixtures already have descriptions. Adding them repairs valid
input rather than inventing a compatibility requirement. The initial failed
receipt is `initial-legacy-revised09.json`; it is preserved with stderr/timing.

Final totals are explicitly **composed**, not fresh full runs: actual rerun skill
checks after adding descriptions plus measured, unchanged parity from the initial
complete runs. Reports are `accepted-composed-final.json` and
`legacy-composed-final.json`; raw focused outputs are `*-skills-final.txt`.
No broad parity was repeated for a skill-description-only fixture correction.
The initial accepted 100/100 remains its actual earlier-fixture run.

Final legacy skills expose real differences: the hidden writer executes;
`$SKILLS` is rendered before activation, so the initial offer is missing and
later lists lag; dependency instructions are omitted from the returned body;
`$TOOLS` is stale/incomplete. Frozen system text, full primary, dynamic/pending
writes and unreachable refusal pass. Source inspection confirms rendering before
state assignment, returning only the requested skill body and incomplete tool
expansion. Those findings are not a general security or correctness assessment.

Legacy inherited parity still fails settings false/zero snapshots, replay
identity/content, the optional-module boundary and earlier core checks. Its
missing separate workflow/consumer setup does not prove behavior that never ran;
media fixture/model limitations and actual nested failures remain in the raw
full report. No adapter or forced pass was supplied.

## Targeted defects and checks

`mutations.json` records a passing baseline and seventeen isolated, compiling
accepted-source defects. Each `.patch` and `.txt` retains the actual edit and
check output. All seventeen fail the intended behavior; no survivor or mutation
setup failure occurred. The separate description omission above is still a
failed grader setup, not a mutation kill.

| Defect | Intended check that fails |
| --- | --- |
| Publish all installed declarations | Initial capability set |
| Bypass dispatch admission | Hidden write error and absent-file control |
| Drop primary text | Primary body |
| Append current declarations to primary | Frozen system text |
| Omit `tool_limits` activation | Full-primary declaration set |
| Truncate body at later horizontal rule | Primary body tail |
| Ignore YAML list items | Dynamic tools |
| Drop dynamic instructions | Correlated load body |
| Omit dependency tools | Dependency declarations |
| Omit dependency body | Dependency instructions |
| Freeze `$TOOLS` at initial set | Point-of-use tool expansion |
| Freeze `$SKILLS` at initial offer | Current discovery list |
| Disclose skills from every definition | One-level discovery boundary |
| Keep loaded skills offered | Loaded-name exclusion |
| Leave `$CUSTOM_VAR` literal | Custom substitution |
| Permit unreachable load | Error/no-grant control |
| Remove tools at pending unload | Real post-unload write and declarations |

The mutation runner used the small retained `probe.go.txt` test bridge, built via
`go test -c`, then executed with `SKILL_PROBE_DIR` pointing at each scratch copy's
actual GUI. It invokes the same `skillsChecks` as the command; exit 0 means the
bridge collected results, not that a mutant passed grading. Verdicts come from
its JSON checks. Copies and temporary test bridge were removed. Mutations used
the initial otherwise identical fixture without descriptions; final accepted
and legacy baselines were rerun after that metadata-only correction.

Affected grader/command `go test` passes (10.245s) and vet passes. Final focused
skills test (0.748s command wall time), final vet, formatting and scoped diff
check pass after the fixture correction. The actual accepted
`TestSkillsThroughLoop` and `TestSkillLoadFailures` pass under race detection
(2.145s command wall time). Commands and hashes are in `validation.json` and
`final-validation.json`; raw outputs are adjacent. No broad root suite or known
original GUI/headless timeout cascade was repeated.

Required graph cycles, missing-dependency atomicity, first-enabling provenance,
Agent isolation, wrong-type refusal and custom renderer replacement retain the
real source tests, original twenty targeted mutations and independent frozen-
prompt mutation under original-numbered `evidence/ch10`. These are separate
obligations, not claims established by the numeric grader. Pending state itself
has source proof; the numeric post-unload write proves retained capability.
No compaction/removal, persisted skills or general provider compatibility is
claimed for this chapter.

## Published samples and retained live evidence

Tested manuscript SHA-256:
`804d37e6449f38421c0ad26b09835ee3301fc9caed0bff0b307f08ccc34d1393`.
Both Go fences were extracted verbatim and compiled together in a temporary
outside module. The public construction/RegisterVar excerpt ran a real Agent
HTTP/tool loop: dynamic instructions used `load-time value`, while primary text
stayed fixed. The three exact published live skill files ran through the built
CLI `prompt` command, loading/writing/reading/unloading in five requests and
leaving `note.txt` equal to `SKILLS_OK`. Exact full-primary YAML declares all
twelve tools. The first frontmatter example also loads successfully when its
explicitly required read-tools/refactor-tools companions are supplied. Results:
`published-go.txt`, `published-samples.json`, `published-cli.json`.

These are fake-boundary API/CLI sample checks, not new human chat or live-provider
runs. They reuse the real command contract; scratch output paths replace the
manuscript's illustrative executable path. No manuscript correction was needed.
An interim message mistakenly repeated the reviewer's older chapter hash prefix;
it was promptly corrected against the actual extraction receipt above.

Three historical October10 live chat receipts retain their original provenance:
Claude five requests, OpenAI four, Gemini3.8 five, successful marker write/read,
pending unload and unchanged system text. The original OpenAI post-unload read
only exercised the initially enabled `read_file`; it did not prove retention of
the dynamically added writer. The manuscript preserves that corrected inference.
No new paid repetition was needed; model names are historical, not an inventory.

## Size, comments and handoff

Counts exclude blanks and `//`-only lines from active Go; inline comments remain
active. New grader source: **261 active / 25 comment / 297 physical** lines;
integration test: **16 / 2 / 20**. Original checks/launcher total **828 active /
104 comment / 1,091 physical**, preserved unchanged. Command changes add Chapter9
dispatch and update flag ranges; shared launcher adds one optional environment
argument and its application. No runtime code or frozen test changed.

New production comments are about 8.7% of nonblank lines, below the roughly20%
teaching target. They explain observation limits, real boundaries, independent
process identity, correlated results, file-effect controls, fixture delimiters,
ordering and valid serializer alternatives. The only new exported function has
a contract comment. Sequential fixture/cleanup code stays small rather than
adding wrappers or comments merely to reach a ratio. Reviewer should assess
clarity and retained coverage independently of these counts.

Final source SHA-256 values and exact commands are in `final-validation.json`.
A broad staged whitespace check initially flagged the literal context prefixes in
retained patch files (space followed by source tabs). Source/prose checks pass;
raw patches remain verbatim and are excluded from that whitespace-only check.
The configured credential hook still inspects every staged path.

No later chapter work is released by this submission. Preserve first-edition
work and all existing historical checkpoints.
