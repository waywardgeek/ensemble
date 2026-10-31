# Stage B: structured child submissions

October 14, 2026. Maintenance coder, GPT-6 Astra/high. Design accepted by coordinator and independent runtime reviewer before implementation.
Formal implementation review remains pending.
Baseline `7a362ec2`, current chapter numbers retained. Mandatory reading includes
`.agents/skills/ensemble-coding/SKILL.md`, AGENTS, the full workflow/carryover,
Chapters14/15/22/23 and all three compaction audit/progress notes. This is
maintenance with prior-answer exposure, not a blind student exercise.

## Implemented boundary

- Add `AgentSpec.SubmitSchema json.RawMessage` and `SpawnWithSchema(id, task,
  model string, schema json.RawMessage)`. Keep `Spawn` as the absent-schema
  wrapper. A child's default derivation explicitly clears its parent's schema,
  including the public `AgentSpec.Child` helper.
- Keep `Submit(string)`; it encodes `{"result": text}` and calls the same
  `SubmitJSON(json.RawMessage)` boundary as the model tool. Custom schemas govern
  the entire direct argument object. All accepted data remains JSON, never JSON
  embedded in a string. `ChildStatus.Result` stays the default text field;
  `ChildStatus.Payload` carries custom JSON only. `SubmissionData(id)` returns a typed `Submission` containing default `Result`
  or custom `Payload`. Legacy `Submission(id)` returns default text and reports
  a clear use-SubmissionData error for custom results.
- The Fleet implementation owns a compiled immutable submission contract. The
  composition root compiles it before constructor side effects and passes it to
  `fleet.New`; public `CreateChild` validates/copies before invoking any custom
  factory, covering both direct callers and Spawn.
  Spec ingress/egress, declaration and result snapshots own their byte slices.
  The factory result must retain the requested schema; a custom factory cannot
  silently discard or replace that contract. No new callback/service layer.
- `submit` obtains its declaration from that compiled contract. Runtime validation
  happens before durable write, state change or join notification. Rejection leaves
  the child unsubmitted, allowing its existing bounded lifecycle to correct it.
  This stage does not change submission-versus-quiescence ordering (Stage C).

## Validator and bounded dialect

Use `github.com/santhosh-tekuri/jsonschema/v6` pinned to `v6.0.3`, with Draft
2020-12 selected explicitly and a loader that refuses all external resources.
The library handles metaschema checking and instance semantics. A small recursive
admission check rejects unsupported keywords; it is not a second validator.
The root must explicitly be `type: object`. Nested boolean schemas are allowed.

Supported keywords: `$schema` (only the 2020-12 URI), `type`, `properties`,
`required`, `additionalProperties`, `items`, `enum`, `const`, `minimum`,
`maximum`, `exclusiveMinimum`, `exclusiveMaximum`, `multipleOf`, `minLength`,
`maxLength`, `pattern`, `minItems`, `maxItems`, `uniqueItems`, `minProperties`,
`maxProperties`, and annotations `title`, `description`, `default`, `examples`.
Nested objects/arrays and nullable type arrays use normal draft semantics;
integer accepts mathematically integral numbers, not merely integer spelling.
All other keywords, references (including local), IDs, dynamic references,
combinators, formats/content and custom vocabularies are explicitly rejected.
This avoids recursive/reference expansion and silently unenforced assertions.
Go's standard RE2-style regexp engine remains the library default.

Admission limits: schema64 KiB, submission1 MiB, JSON nesting64, schema nodes2048,
instance nodes65536; numeric lexemes128 bytes and exponent magnitude1000. These
are admission ceilings against recursion/allocation and enormous rational-number
construction, not memory summarization targets. Rejection is explicit and never
truncates data. They leave ordinary nested results and the memory schema ample
room. No arbitrary model deadline or token cap is added.

Both raw and canonical JSON must fit their byte caps: JSON escaping can enlarge
accepted input, and the stored representation must remain admissible on reload.
Before instance validation, canonical schema bytes × canonical argument bytes
must also fit64 MiB (division avoids overflow). This conservative combined budget
can reject otherwise individually admissible pairs, including large annotations.
It bounds repeated validation/diagnostic expansion for this reference-free subset;
it is not a CPU-time guarantee or a second schema evaluator.

Dependency research: [upstream library](https://github.com/santhosh-tekuri/jsonschema)
and [versioned API](https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6@v6.0.3),
plus downloaded v6.0.3 source. `DefaultDraft`, `UseLoader`, `AddResource` and
`Validate` provide the boundary; `UnmarshalJSON` retains exact `json.Number`
values. Upstream documents standard-library regexp as the default. The core and GUI module files now pin the dependency; `x/text v0.14.0` is
transitive. No provider adapter or supported-model change is involved.

## Storage and presentation

Default submissions keep existing `submit.md` JSONL rows `{"result":"..."}`.
Custom rows use an explicit envelope `{"version":1,"submit_schema":{...},
"payload":{...}}`. Each row is self-describing for durable retrieval even when
a same-ID child is respawned with a different contract. Read legacy and custom
rows without promoting an old result into completion of the new task; historical
custom rows are checked against their recorded contract when read. Reject unknown
envelope versions and malformed rows. Ownership lookup precedes every disk read.

Join/status use Payload as raw JSON; `get_submission` returns `agent_id` plus
legacy `result` or custom `payload`. Existing tool-result text carries a JSON
document, as before, but its inner payload is an object. The GUI already renders
tool arguments and result text safely; verify actual selected conversation/replay
and nested JSON rather than introduce a new result browser. Agent startup retains
its caller-supplied immutable schema; saved history preserves old declaration and
tool arguments, and `submit.md` supplies historical result retrieval. Ephemeral
children still are not automatically resurrected.

## Verification

Use real Agent/Engine/Fleet/public constructors with small HTTP provider fakes,
including invalid-then-correct submit, pending Join, direct string API rejection,
default child/grandchild reset, foreign/missing retrieval and durable reload.
Exercise exact memory declarations across Anthropic, Chat Completions, Responses
and Gemini, plus nested values/arrays/null/integer/enum/additional-properties,
invalid schemas/keywords/refs and admission limits. Mutate caller and returned
schema/result buffers and verify retained declarations/results. Check a same-ID
new task cannot become submitted from an old file. Test GUI selection/replay with
a custom result. Run focused races, full core/GUI short/race and vet once coherent;
coordinator owns advertised grader execution/build repair. No live inference.
Targeted scratch mutations will cover the meaningful validation, ownership,
copying, inheritance and persistence claims. Record actual failures/limits here.

Accepted clarifications: default submissions continue accepting extra keys but
normalize to the legacy result string/row; only custom contracts preserve the
full object. Admission checks descend only into schema-bearing positions, never
object values inside enum/const/default/examples. Existing Ask/ask_agent and
TurnEnded text display use compact JSON for custom results, while typed status
and SubmissionData retain raw JSON data. No empty success or new result browser.


## Implementation evidence and initial-review corrections

The public API carries custom objects through declaration, submit dispatch,
status, join, typed retrieval and compact-JSON text completion. Legacy extra keys
remain ignored. Default rows stay unchanged; custom rows carry their exact schema
and payload, with ownership checked before fallback. Invalid submission cannot
write a row, complete Join or turn historical data into the new task's success.
A failed durable write leaves the task unsubmitted. No GUI production machinery
was added: its existing tool argument and result surfaces carry the new data.

During initial review, the independent runtime reviewer identified four source
issues and one assertion gap. Canonical escaping could exceed the reload cap;
independent schema/instance caps could still multiply validation work; and the
moved factory ownership-refusal branch could shut down an unrelated Agent (or
its parent while holding the parent Fleet mutex). The first two now have the
explicit bounds above. Wrong-ownership refusal no longer calls Shutdown on that
Agent; correctly-owned wrong-contract children are still stopped. Real parent
and foreign Agents survive rejection, and refused children never enter the
parent's child list. Public CreateChild now also rejects nil/foreign parents
before invoking the factory, matching NewChildAgent's required-parent rule.
This does not repair the separate root-registry cleanup or
submission-versus-quiescence issues reserved for Stage C.

The GUI check now subscribes before inference, checks actual live tool frames,
then saves and loads a fresh Agent and checks replay. It parses the text result
following `submitted: ` and compares the full nested object, including array,
boolean, null and number, rather than searching frame text. Reload makes no new
provider request. Four fake provider surfaces independently capture the exact
memory schema: Anthropic, Chat Completions, Responses and Gemini. These are real
adapters/Agent/Engine/Fleet paths with small HTTP peers, not mocks or live use.

Actual checks (test children remove only `LLM_`, `OPENAI_`, `ANTHROPIC_`, `GEMINI_`,
`EN_`, `CODEX_` environment prefixes):

- Focused schema/structured-submission/fleet/management race tests pass after the
  canonical/budget and public CreateChild changes (fleet1.109s, ensemble4.597s).
- Full GUI `go test -race ./...` passes (11.856s) with exact live/replay checks.
  The first GUI fixture timed out because it omitted the existing subscribe
  message; that test setup was corrected before the passing run.
- Core and GUI `go vet ./...` pass after the production refinements.
- The earlier full core `go test -short -race ./...` failed only the independently
  reproduced baseline CLI human-sink case; every other package passed, including
  ensemble89.427s. The later full run again has that baseline failure and exposed
  a new factory-survival fixture missing Anthropic usage (ensemble91.132s).
  Correcting the fake's usage fields makes the narrow race regression pass
  (ensemble1.040s, fleet1.047s), then the full ensemble package race rerun passes
  (91.778s). The final required-parent guard was added after that full run began;
  the final focused Stage B race run passes (fleet1.080s, ensemble1.422s), covering
  that boundary and all new submission behavior.
- Intermediate focused/core and GUI reruns failed linking with ENOSPC. The
  coordinator cleared only the regenerable Go build cache, restoring11 GiB;
  these are setup failures, not behavioral results. Subsequent runs above passed.
- Coordinator reran the advertised old22 and old23 compatibility graders against
  the frozen current `./agent`: 200/200 and100/100 respectively, both exit0.
  Its source manifest and logs are under `/tmp/ensemble-compaction-baseline/`.
- An extra `node --test *.test.mjs` invocation was a coder mistake: these browser
  files are not a uniform standalone suite (handoff §9). Components and fleet
  require ASSET_URL/GUI_BINARY from the already-passing Go harness. The independent
  speech script exposed a missing virtual-user submit declaration; an exact
  `a8eec84a` scratch archive reproduces the same assertion (3.590s, exit1), so it is
  inherited. The temporary archive was removed; `speech-baseline.txt` retains the
  result with the mutation logs. Other standalone checks passed in the glob run,
  but that mixed invocation is not claimed as a passing suite. No broad rerun.
- This maintenance stage has performed no live inference or deployment.

Targeted mutations ran one defect at a time in temporary copies, removed after
the runs. Initial15 cases detected skipped runtime validation, unknown-keyword
admission, raw byte and numeric limits, Spec/status/declaration copying,
grandchild schema inheritance, raw input aliasing, durable payload loss,
malformed-row fallback, spawn-tool schema loss, incorrect provider declarations,
string-API bypass and retrieval double encoding. Eight refinement checks cover
canonical schema expansion, canonical payload expansion, combined budget,
updated grandchild reset and raw-input-copy locations, factory survival, pre-factory parent validation and GUI double encoding. All exited1 at the intended check, without build failures or
timeouts. The factory check was repeated after its faithful HTTP fixture passed;
restored Shutdown fails specifically with `agent stopped`. The final grandchild
mutant reaches the real child and exhausts its tool rounds because its inherited
custom contract refuses default submission. The declaration-alias mutant also
causes cleanup serialization errors; those coupled errors are not separate proofs.
Several schema-semantic cases overlap (e.g. short text also violates pattern);
these demonstrate the actual validator boundary, not independent mutation proof
for every upstream keyword. Small boundary mutants are used, not large stress
runs. Logs are in `/tmp/ensemble-structured-submissions-mutations/`; temporary
sources have been removed. There are21 distinct defect claims across24 executions,
including the repaired-factory recheck; two updated location checks repeat prior
claims. The full records remain historical rather than being relabeled final.

Size comparison against `a8eec84a` (same pre-B application as `7a362ec2`): for13
changed/new production Go files, physical/active/comment-only lines move from
2212/1701/332 to2596/2007/388, net384/306/56. Counts exclude blank and `//`-only
lines from active code; moved submission behavior is counted as deletion plus
addition, not wholly new behavior. Three new test files total764 physical,
712 active and33 comment-only lines. Production growth is the bounded schema
contract, typed/durable result path and existing API wiring; tests exercise real
runtime, persistence, copies and provider/GUI boundaries. Comments explain
ownership, admission versus semantics, validation-before-completion and reload
identity. No additional scheduler, generic schema framework or result browser.

Mandatory `.agents/skills/ensemble-coding/SKILL.md` and the full carryover were
reloaded after compaction before these edits. The earlier full workflow/chapter
reading remains recorded above. Formal independent acceptance remains pending;
Stage C, chapter renumbering, exports and live verification are not claimed.
