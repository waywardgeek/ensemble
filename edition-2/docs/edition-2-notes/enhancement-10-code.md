# Edition 2 Chapter 10 grader submission

Coder: `edition2_grader_coder`, GPT-6 Astra/high, October 11, 2026.
Initial submission; independent review pending. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

Read the full skill, AGENTS, carryover, workflow, enhancement/map, original
Chapter11, complete student/comparative reviews, original persistence harness
and checks, and accepted/legacy implementation paths. Reloaded the full skill,
carryover and workflow after compaction. Original Chapter11 revision is
`994bf218b31033b7dbf2740dd0187d509d2927dc`, SHA-256
`ed6482dd1e5247880122382a3b4199276d8fbda13ea8863676eda5c6e49e9df0`.
Accepted source is direct `solutions/edition-2/ch10`, original11 checkpoint
`edition-2-ch11-r1`; legacy is `solutions/ch11`. No frozen source, original
grader or original prose changed. No mocks or paid calls were added.

## Preserved scope and corrected fixtures

The six original persistence weights remain15/20/20/15/10/5, plus15 for corrected
Edition2 Chapter9 parity:100 total. Command:
`go run ./cmd/grade-edition2 -ch 10 [core-directory]`; default is accepted `ch10`.
Existing GUI and inherited fixture flags retain their meanings.

Both actual commands receive existing stdin prompt envelopes and EOF. Core and
GUI build separately; GUI uses `--port 0` and its public asset argument. Real
startup, public Agent, actor, history, reducer, filesystem and shutdown all run;
only HTTP is the existing course fake. Private context/events stay opaque.
The grader checks the four public fields, event sequence identities, actual next
request bytes and actual saved file effects. It does not require particular Go
objects, packages, constructor shapes or canonical disk encoding.

The new phases preserve and strengthen the original valid obligations:

- Save contains config/as_of/context/log, a non-null normal snapshot, positive
  increasing event identities and an anchor equal to the last retained fact.
  A restart appends without changing or dropping its loaded log prefix.
- Saved model/vendor/effective system prompt/declarations match the actual HTTP
  request. No primary is supplied initially, exposing the historical raw-empty
  versus effective-default prompt defect. Fake key and endpoint must be absent.
  Undefined `think` is removed; existing declarations are compared to the wire.
- Default restart retains each user prompt exactly once. `--save` loads and
  saves the selected path without also creating the default file.
- Current model/vendor/primary/tools win over recorded config while retaining
  conversation. The declared OpenAI course row is `gpt-5-course`.
- Full snapshot, null-context complete log, older snapshot plus newer tail,
  and snapshot-only inputs yield byte-identical next requests with identical
  current settings. Fixtures are independently derived from the complete save.
- First new sequence is exactly one beyond the larger retained identity/anchor,
  including a snapshot-only input whose removed log cannot be recreated.
- Truncated JSON and JSON `null` both refuse with nonzero exit, name the save,
  make no HTTP request and leave its bytes unchanged. Timeout is not refusal.

The command deliberately chooses cwd/save.json; library AgentDataDir ownership
remains separate source proof. No journal, crash-recovery framework, hashes,
canonical disk format, persisted skill runtime or old retired machinery is added.

Three fixture corrections are preserved honestly. The first draft buffered two
prompts together; the accepted actor correctly classified the second as a live
hint. `initial-baseline.txt` retains that failure. One prompt per process and EOF
now provide a real completion/restart boundary. No sleeps or actor edits were
needed. The first current-model name and my subsequent `gpt-5` assumption were
both absent from the legacy model table, so each failed before HTTP. Their
artifacts remain; the final row comes from the actual table. Those failures do
not demonstrate configuration restoration defects.

The first full accepted run also exposed a newly relevant inherited fixture:
original Chapter2 runs independent Anthropic/OpenAI/Gemini sessions in one cwd.
Persistence correctly loaded each previous conversation. Actual dumped contexts
contained one, two and three sessions, causing normalization/provenance failures.
The Chapter10-only parity launcher assigns a fresh explicit save path to each
no-argument process, preserving cwd and fixture files. Offline render/dump
arguments pass unchanged. It executes the actual submitted binary and its normal
CLI; its own temporary directory owns every selected save and is removed later.
All six persistence checks run the unwrapped command before that launcher exists.
Thus isolation of old independent sessions cannot suppress current persistence
failures. Original graders remain untouched. Focused Chapter2 changes85→100;
`parity-ch02-initial.json` is a text report despite its historical filename.

## Actual scores and provenance

Evidence directory: [enhancement10](../../solutions/evidence/ch11/enhancement10/).

| Measurement | Result |
| --- | --- |
| Historical original full accepted-student run | Interrupted/unscored, exit2 |
| Historical unchanged focused student persistence | 70/85 on initial source; undefined think stops shape before later predicates |
| Historical original legacy baseline | 100/100 PASS, retained rather than repeated |
| Initial revised accepted full | 85/100 FAIL: persistence85/85; independent-session parity fixture fails |
| Initial revised legacy full | 60/100 FAIL: undeclared-model setup, null-save refusal, inherited parity |
| Final actual accepted full | **100/100 PASS** |
| Final actual legacy persistence | **80/85**; only null-save refusal fails |
| Final legacy composed total | **80/100 FAIL**, not a fresh full run |

`accepted-final10.json` is the actual complete final run. Its one repeat was
necessary because the new parity launcher affects inherited invocations.
`legacy-composed-final.json` combines the final affected persistence measurement
with the unchanged failing skill checks from the prior full run. Those actual
Chapter9 failures alone keep all-previous-checks parity at zero; the legacy
Chapter2 isolation correction was not rerun and is not called passing.
`legacy-revised10.json` retains all earlier nested details. Setup/model failures
are not carried forward as final behavior failures.

Final legacy null input produces exit0, one provider request and changed bytes
in both client launches. Truncated JSON refusal passes. Correctly accepting a
snapshot without a log is unrelated to accepting no SaveFile at all.
`legacy-final-persistence.txt` retains the intermediate undeclared `gpt-5` run;
`legacy-declared-model-persistence.txt` is the final affected result.

## Mutation sensitivity and limitations

`mutations.json` records21 isolated compiling defects and a passing baseline.
Each `.patch` and `.txt` preserves the exact change and actual check arrays.
`affected-mutations.json` adds three distinct shape defects and reruns the four
current-config defects under the final declared-model fixture. All24 distinct
defects fail their intended behavior; no survivor or compilation failure occurred.
Baseline/bridge exit0 means collection succeeded, not that a mutant passed.

| Defect group | Intended failed claim |
| --- | --- |
| Raw configured prompt; missing system field; missing top-level anchor; missing snapshot; lost declarations | Shape and effective wire metadata |
| Stale anchor; dropped loaded log; sequence gap | Anchor, retained prefix, exact next identity |
| Skipped tail; replayed already folded events | Tail applied exactly once |
| Respected anchor during null-context rebuild | Complete-log replay |
| Discarded snapshot; lost trimmed anchor | Snapshot-only completeness and next identity |
| Fake key or endpoint appended to metadata | Explicit transport-data absence |
| Ignored --save; skipped startup Load | Symmetric path and actual restart |
| Restored old model, vendor, prompt or ignored current tools | Current runtime configuration wins |
| Skipped exit Save | Actual save existence |
| Ignored bad Load; saved after failed Load | Refusal before HTTP and unchanged damaged file |

Missing top-level field and absent save fail at the first observable artifact;
subsequent setup propagation is not counted as additional independent kills.
Four final current-config repeats are repeated observations, not four new defects.
The parity before/after result demonstrates the newly corrected isolation issue;
original fixtures and genuine persistence mutations remain separately tested.

The small temporary test bridge invokes `persistenceClient` on the built scratch
copy. It was compiled with `go test -c`, then removed; scripts and bridge source
are retained as plain-text receipts. Scratch mutants were removed. No submitted
source or runtime test was changed. Mutations demonstrate these chosen predicates,
not exhaustive malformed-input validation or all concurrency interleavings.

Affected grader/command tests pass (10.439s), vet passes, and formatting passes.
Accepted `TestSaveSnapshotReplayAndTail`, `TestSaveOwnershipAndRuntimeConfig`,
`TestBadSaveLeavesHistory` and `TestSaveIntentionallyEmptyPrimary` pass under race.
These actual source tests establish rich neutral three-vendor replay, owner paths,
empty-primary semantics and staged refusal beyond the numeric fixtures. Historical
18 source mutations plus the two prompt-correction mutations retain those
unchanged obligations; no repeat full original timeout cascade or paid freshness
run was needed. Atomic replacement failure/durability boundaries retain source
review and original tests, not a new numeric guarantee.

## Exact public samples and live reuse

Tested manuscript SHA-256:
`32723da4cf2531b5786330502708d58c62db613fd71a7a73acd855c2e83b6b22`.
Both Go fences were extracted verbatim: vocabulary types compiled with the real
common types; private restoreContext compiled in an isolated real history package
and exercised actual validate/apply for complete-log, older-tail, snapshot-only
and invalid-event inputs. This tests the explicitly scoped excerpt, not missing
surrounding configuration/order validation.

The exact published `/tmp/ensemble-ch10` build and human `chat` invocation ran.
The exact `--save chosen.json chat` ordering also ran twice, retaining the first
prompt without creating default save.json. Four local fake-boundary HTTP requests
were captured. Public GUI prompt/EOF is exercised by both-client numeric grading.
`published-samples.json` records hashes, commands and outputs. No sample correction
was needed. This is runnable sample validation, not new live-provider evidence.

The retained October10 Claude/OpenAI/Gemini3.8 restart receipts remain historical:
eight successful response processes including independent Claude snapshot/log-only
variants, then a ninth corrupt-file refusal with zero HTTP and unchanged bytes.
The initial-source provenance and later effective-prompt correction stay distinct.
No claim of general semantic quality follows from recalling cobalt-juniper.

## Size and handoff

Counts use nonblank, non-`//`-only lines as active Go; inline comments remain active.
New grader:467 active /56 comment /543 physical lines; integration test22/2/27.
Original preserved harness/checks:693 active /121 comment /890 physical lines.
Command changes add only Chapter10 dispatch/default/range support.

New grader comments are10.7% of nonblank lines, below the roughly20% teaching
target. They explain lifecycle boundaries, independent-session isolation, opaque
format handling, actual-wire comparison, retained log identity and evidence
limits. The only new exported function, Chapter10, has its contract comment;
there are no new exported fields or interface methods. Counts do not replace
independent KISS/coverage assessment. No general grading framework was added.

Exact final source hashes and commands are in `final-validation.json`. Isolated
index uses the same base/parent and atomic expected-old HEAD publication; only
owned files are committed and reconciled. Configured credential hook remains
active. Raw mutation patches retain literal diff whitespace and are excluded
only from whitespace lint. Preserve all CodeRhapsody and unrelated work.
Stop after this submission until independent Chapter10 acceptance/release.

Receipt correction after independent review: the initial composed JSON summary
incorrectly included `load-skill`, which passed, and abbreviated other IDs.
The corrected receipt names the actual four skill failures: `initial-tools`,
`progressive-disclosure`, `depends-autoload`, `var-substitution`. Raw results and
the80/100 composed score are unchanged; no rerun or source change was needed.
