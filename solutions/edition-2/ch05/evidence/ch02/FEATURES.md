# Chapter 2 feature evidence

Status: first student implementation; independent comparison/revision gate pending.
All dates are 2026-10-07 UTC. Historical files outside this directory are Chapter 1
receipts inherited from `75542c1`, not proof of Chapter 2 behavior.

## Actual user-facing runs

`live/discovery.json` records provider model discovery. The temporary runner reads
only the three authorized key fields into memory, passes the selected key through
the child process environment, and scans saved evidence for those actual values.
Neither keys nor settings were printed, passed in arguments, or stored here.

Each CLI command invokes `/tmp/ensemble-ed2-ch02` with three human prompts and one
intervening ephemeral directive; the exact inputs/outputs are in each receipt.
Each fresh log was also dumped and rendered twice with credentials removed.

| Vendor/model | Receipt directory under `live/` | Answers | Usage I/W/R/O |
|---|---|---|---|
| Anthropic `claude-sonnet-5-5` | `anthropic-rj4hqrm9` | Silent Harbor; Silent Harbor; robraH tneliS | 285/0/0/115 |
| OpenAI `gpt-6-luna` | `openai-jbkw49we` | Velvet Comet; Velvet Comet; temoC tevleV | 239/0/0/130 |
| Gemini `models/gemini-3.8-flash` | `gemini-eursj_gg` | Cobalt Echo; Cobalt Echo; ohcE tlaboC | 184/0/0/562 |

All exited 0, produced empty stderr, and yielded identical repeated render bytes.
`ephemeral-replay.json` in each directory proves the marker is absent/present/absent
in the three requests reconstructed from the actual live log prefixes. These
prefix renders are offline evidence, not an HTTP capture. The recorded request
facts consume ephemera `[]`, `[4]`, `[]`.

The separately built public program `examples/consumer` uses actual model call
IDs, ingests a controlled result (no claimed tool execution), renders before/after
redaction, asks the model again, checks ordered Agent-attributed observations,
unsubscribes and appends another event, then creates a separate Agent and changes
its model. Its deliberately rejected unknown-call event logs through Ensemble.

| Vendor | First model / second model | Receipt | Status |
|---|---|---|---|
| Anthropic | claude-sonnet-5-5 / claude-sonnet-5 | anthropic-7ypqjdg4 | PASS |
| OpenAI | gpt-4.1-mini / gpt-6-luna | openai-bfylkpz1 | PASS |
| Gemini | gemini-3.8-flash / gemini-3.7-flash | gemini-tbijb3lq | PASS |

All three successful consumers observed 8 ordered events before unsubscription, kept
Agents independent, and recalled ORCHID-572 after changing models without altering
the first model's accounting. OpenAI recorded the returned identity
`gpt-4.1-mini-2025-04-14` separately from requested `gpt-4.1-mini`.

## Preserved live failures and teaching findings

- `openai-v8y51fgm`: initial tool request to gpt-6-luna failed safely with HTTP400.
  A bounded separate diagnostic identified this model's default-reasoning/tool
  restriction on Chat Completions. Selecting discovered gpt-4.1-mini for the tool
  phase corrected the demonstration without inventing a new adapter or retry.
- `gemini-k6_8m17i`: initial declaration failed safely with HTTP400. The bounded
  diagnostic identified `additionalProperties` inside the OpenAPI `parameters`
  field. The shared declaration is JSON Schema; Google's official
  [FunctionDeclaration documentation](https://googleapis.github.io/js-genai/release_docs/interfaces/types.FunctionDeclaration.html)
  specifies `parametersJsonSchema` for that representation. The corrected adapter
  preserves schema constraints, passes its focused regression, and completed the
  full real Gemini consumer run. Gemini supplied a signed call; explicit returned
  model identity permitted its replay in the redacted follow-up.
- `*-tool-diagnostic.json` contains sanitized provider diagnostics from these
  explicitly separate, bounded probes. They are not successful user-path runs.
- The coordinator found missing owner context on several free helpers and a
  caller-buffer alias in Load configuration. Those were student implementation
  errors against already-taught rules. The full helper pass and central clone
  correction received a targeted reviewer acceptance; local regressions pass.

## Local checks and feature map

These are deterministic local fixtures, not real-model claims:

| Feature | Evidence |
|---|---|
| Failure history, consumed ephemera, atomic usage | TestEphemeraFailureAndUsageAttribution; retained Chapter 1 atomicity/error tests |
| Exact provenance, signed call compatibility, synthesized Gemini call ID | TestToolResultsOpaqueAndOwnership |
| Supplied result/redaction/reference mapping | Live consumer render files plus TestPublishedReplayFixtures and inherited grader |
| Normalized cache/thinking counts, OpenAI JSON argument decoding | TestToolResultsOpaqueAndOwnership; TestOpenAIParsingAndInstructionRoles |
| Durable schema rejection and all-or-nothing replay | TestMalformedLogs; independent offline CLI acceptance |
| One deferred human, result-before-human projection, unresolved-call refusal | TestDeferredHumanPreservesFactsAndProjectsAfterResults |
| Caller/event/returned-buffer ownership | TestToolResultsOpaqueAndOwnership; TestLoadOwnsConfiguration; GUI public test |
| Partial append permanently faults writer | internal/eventlog TestPartialWritePermanentlyFaultsLog |
| Explicit Engine transport; safe cancel/deadline logging | retained Engine transport and public cancellation regressions |
| Public client boundary, attribution, close/unsubscribe, parent logger | separate GUI module TestPublicClientBoundary |
| Optional headless use | separate consumer module builds without importing GUI |

The GUI has no browser transport. Its fake-backed public test is not a live
WebSocket claim. Imported reference/opaque/invalid-record fixtures are labeled
local; provider responses cannot be expected to produce those negative controls.

Executed checks so far: `go vet ./...` and `go test ./... -count=1` in main, gui,
and examples/consumer modules; all PASS. Inherited `make grade-dir CH=2
DIR=solutions/edition-2/ch02` first run passed 100/100. Coordinator independent
CLI acceptance passed 33/33 on the helper/config/deferred-rule revision; expanded
coordinator checks are separate. Final format/vet/tests passed all three modules
with empty gofmt output (local-checks.json). The final inherited grader scored
95/100: its dump-to-render fixture contains an unanswered call and conflicts with
the newly clarified refusal rule. The initial pre-clarification100/100 and this
failure are both retained; no student behavior was weakened. Coordinator fixture
resolution and independent comparison/revision remain pending.

## Source binding and limits

`initial-source-sha256.json` identifies this initial checkpoint's Go files.
`pre-schema-fix-source-sha256.json` preserves the code before the Gemini field
correction. The Anthropic CLI ran before the helper reachability/Load-copy pass;
OpenAI/Gemini CLI and Anthropic/OpenAI consumers ran after that pass and before
the Gemini-only declaration correction. The final Gemini consumer uses the
corrected schema field. No claim is made that the final binary was rerun against
every vendor; the changed Gemini behavior alone needed its live correction.
All other vendor paths remain locally checked. Paid calls were bounded: no
automatic retry loop was used.

Runner sources are saved here without credentials. Receipt commands name the
actual temporary binaries and fresh directories. Those paths are historical;
rebuilding and using fresh paths is necessary to reproduce the demonstrations.
Provider replay signatures are part of the recorded neutral history, not API
credentials. All new evidence was scanned in memory against the three actual
keys before copying and before commit.

## Post-comparison status

The initial checkpoint39a92ca is preserved. The reviewer accepted the subsequent
internal-access, safe-diagnostic and WHY-comment revision described in
REVIEW-REVISION.md. Current module gates and inherited grader pass100/100.
The original95 result remains; the coordinator corrected its unresolved-call
fixture without weakening the implementation. Initial external acceptance and
mutation receipts are prefixed initial-ch02; revision results use reviewed-.
SOURCE-EXPOSURE.md distinguishes known direct reads from inherited summaries and
the limitation on claiming this was strictly blind regeneration.
