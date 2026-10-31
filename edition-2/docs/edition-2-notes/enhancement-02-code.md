# Edition 2 Chapter 2: grader enhancement

Contributor: `enhancement_coder_resumed`, October 10, 2026, recovering
`enhancement_coder`'s draft. Submitted for independent review; no formal
correction round has occurred. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`, read in full with AGENTS, carryover,
enhancement plan and workflow. No implementation, original grader or frozen
export was changed. No mocks, paid calls, commits or Chapter 3 work.

## Scope and preserved coverage

`grade-edition2 -ch 2` reuses the original harness and checks. It removes only
the premature four-point URI-media render check and replaces the original
four-point Ref-redaction check with an eight-point observation of in-scope
behavior. The total remains 100. The other twelve original checks remain; seam
render/parse gain the documented missing cases and Chapter 1 parity incorporates
the already accepted exact-version correction. There is no new CLI or public Go
API requirement. Log-field normalization retains the original grader's naming
flexibility; vendor wire names remain exact.

The added observations are Gemini result names, captured response/call model
identity and call-bound signatures, Anthropic results arriving on both sides of
human text, and nonempty redaction stubs with preserved original log facts.
The unresolved-routing probe deliberately differs from the fake's returned
model, so a prewritten replay fixture cannot stand in for parser capture.

Redaction checks all three Ref kinds. It observes nonempty replacement text,
absence of the removed secret on the wire, and the original exact Kind/Locator,
secret and redaction event in `dump`. This does **not** claim byte-for-byte log
identity or expose the Ref inside the reduced Context. Exact reduced Ref and
append-only preservation remain covered by accepted source tests
`TestRefCarryforwardAndMissingCallIDs` and
`TestReducerStateAndRedactionReplay`, plus their independently reviewed mutations
in [the original record](../../solutions/evidence/ch02/mutations.txt).
No locator rendering or context-dump command was invented to expose internal state.

Source: original Chapter 2 revision
`e1605916114230570b348ab5961c193fa1cebf80`; unchanged accepted export
`solutions/edition-2/ch02/` at `edition-2-ch02-r3`; legacy `solutions/ch02/`.
Latest revision touching reused Chapter 2 grader/fake files is
`c871073c2f5a3117b5969119345a95e02e617746`. Historical results and the media
exception remain in [ch02.md](ch02.md) and [ch02-code-review.md](ch02-code-review.md).

## Preparation corrections

The predecessor's first supplemental fixture failed because it omitted call
provenance. This was grader setup failure, not a student defect. The final
interleaved fixture includes timestamps, request target, response usage,
response/call provenance, dispatch names/arguments and both results. It follows
the published contract rather than relying on permissive zero values.

The reviewer identified that a nonempty response object could still contain an
empty stub. The final check examines nonempty textual payload, accepting the
existing alternative result-wrapper spellings. The empty-stub mutant below
fails this corrected observation. New exported decode fields and the exported
test function now have purpose/contract comments. These address the three
preparation findings in [enhancement-02-review.md](enhancement-02-review.md).

## Actual baseline and compatibility results

Built the original and revised commands once, then ran each sequentially with
`-ch 2 -json` against each directory. Equivalent source invocations are:

```sh
go run ./cmd/grade -ch 2 -json ./solutions/edition-2/ch02
go run ./cmd/grade-edition2 -ch 2 -json ./solutions/edition-2/ch02
go run ./cmd/grade -ch 2 -json ./solutions/ch02
go run ./cmd/grade-edition2 -ch 2 -json ./solutions/ch02
go test -count=1 ./internal/gradeedition2 ./cmd/grade-edition2
go vet ./internal/gradeedition2 ./cmd/grade-edition2
gofmt -l internal/gradeedition2 cmd/grade-edition2
```

| Source | Original result | Revised result |
| --- | --- | --- |
| Accepted Chapter 2 | **92/100 FAIL**, `ref-render` and unredacted media control in `ref-redaction` | **100/100 PASS** |
| Legacy Chapter 2 | **100/100 PASS** | **100/100 PASS** |

Focused tests and vet pass; formatting and scoped whitespace checks are clean.
The regression builds and executes the actual accepted program through its
public CLI and the existing external provider fakes, including the new probes.

A broader root `go test -count=1 ./...` and `go vet ./...` exposed existing
`scripts/edition2` fixture-package setup errors: mixed `gui`/`ensemble` packages
and an unavailable `example.com/ensemble/internal/common` import. Reproduced both
diagnostics from unchanged HEAD copies of those three script files and module
metadata in an isolated temporary directory; this is not a passing whole-module
check. No script or unrelated dependency was changed. At the coordinator's
direction, terminated that specific broad test process and its still-running
`internal/grade` test child after roughly 90 seconds; unrelated grader suites
were outside the affected checks. The broad run is **incomplete**, not a pass.
Focused affected-package results above remain the completed validation.

## Targeted source mutations

Each mutation changed one behavior in a separate scratch copy of the accepted
source. Both actual graders ran against that source; all nine valid mutants
compiled and reached their intended check. Each original result stayed at
92/100 with only its two existing media failures, demonstrating the original
coverage gaps. The unmodified revised baseline was 100/100.

| Defect restored | Revised result and intended failure |
| --- | --- |
| Append `-wrong` to Gemini `functionResponse.name` | **77/100**: `seam-render` reports `read_file-wrong`; Ref stub check also rejects the incorrect result name. |
| Use routing target instead of returned Gemini `modelVersion` | **85/100**: `seam-parse` reports response and call model `gemini-routing-alias`, expected `gemini-3.5-flash-fake`. |
| Use routing target for call provenance only | **85/100**: `seam-parse` reports the call-model mismatch while response identity remains correct. |
| Drop parsed Gemini call signature | **85/100**: `seam-parse` reports lost exact signatures for `fc_fake_1` and `capture-call`. |
| Retain Anthropic arrival order instead of final results-first partition | **85/100**: `seam-render` rejects interleaved result/text/result ordering. |
| Omit redacted stub text from result flattening | **92/100**: `ref-redaction` rejects empty result stubs for kinds 1, 2 and 3. |
| Change original tool-result Ref locator during dump | **92/100**: `ref-redaction` reports lost original facts for all three kinds. |
| Clear original tool-result text during dump | **92/100**: `ref-redaction` reports lost original facts for all three kinds. |
| Omit redaction events during dump | **92/100**: `ref-redaction` reports lost original facts for all three kinds. |

Two setup mistakes were corrected before counting evidence: an overly broad
name-replacement pattern matched both call and result constructors, so its
assertion stopped before mutation; the first dump-locator mutant treated Ref
as a pointer and failed to compile. The final locator mutation checks its
nonzero Kind and produces the intended behavioral failure. Neither setup
failure is a killed mutant; no final valid mutant survived. Scratch source
copies were removed after their checks. Small raw JSON reports were retained
temporarily for reviewer inspection; this table is the durable evidence.

Unchanged checks reuse the original Chapter 2 mutation record and Chapter 1
exact-version mutation evidence in [enhancement-01-code.md](enhancement-01-code.md).
The source-level Ref mutants and baseline tests establish internal preservation
separately; the generic black-box grader cannot observe that private field.

## Exact published example

Extracted the sole `jsonl` fenced block from `book/edition-2/chapter-02.md`
without modification. Fixture SHA-256:
`06fc7b51af868ce7386008bdee4faf7daf2987fac11901b720b26c5848f9bff9`.
It contains the header plus all nine events. Built the frozen accepted solution
and ran `dump` and `render sample.jsonl` with no API key, target `example-model`
and unreachable `http://127.0.0.1:1`, using each target vendor in turn.

`dump` retained the original first-file secret and all nine events. Each render
omitted that secret, retained a nonempty stub, second-file output and the human
instruction, and was byte-identical on a second invocation. Anthropic's final
user message was precisely `tool_result(one), tool_result(two), text`.
Output sizes including trailing newline: Anthropic **718**, OpenAI **685**,
Gemini **756** bytes. The original input file remained unchanged. These are
new offline sample observations, distinct from historical live byte counts.

No runtime behavior changed, so the already reviewed three-provider human-chat,
ephemera and nine-way live-log replay evidence remains applicable. Its provenance,
initial OpenAI 404 and ephemera self-correction limitation remain in
[the student record](ch02.md) and [live evidence](../../solutions/evidence/ch02/live.txt).
No new live-model or media capability is claimed.

## Size and simplicity

Counting physical Go lines as blank, leading-`//` comment-only, or active
(including imports/braces), new Chapter 2 grader support is **371 active /
50 comment-only / 20 blank / 441 physical**. Its integration test is **19
active / 3 comment-only / 3 blank / 25 physical**. Command dispatch is **56
active / 3 comment-only / 4 blank / 63 physical**, with a source delta of
13 added and 5 removed lines. These are grader/support counts, not new Ensemble
production or student-test growth. The wrapper preserves the original harness
instead of copying it; no general mutation runner or evidence framework is added.
Comments explain scope, observation limits, fixture ordering and capture identity.
Frozen implementation counts/documentation remain historical review facts.
