# Chapter10 provider support preparation — LOCAL ONLY

Runtime and actual CLI/GUI/public-consumer builds remain
`57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4`. Initial adapter/split-binding support is
`065a6a824ec46a70affd85bd5b95cdc19439495d`; source-specific later validation is
recorded below. Final support is `9822b2b63257724001d7af195b377483baad3531`,
which tightens the remaining TCP→TLS deadline and adds its local control.
No credentials, real discovery/authentication/provider requests,
Go build/test/vet, cache cleanup, runtime changes, push, snapshot or tag occurred.
No compiler slot was used. The final inbox now reports it released after the
independent deterministic stage; this support needs no rebuild. This is not chapter acceptance.

Changed support paths under `evidence/ch10/support/`: `provider.py`, `discover.py`,
`bind-support.py`, `test_provider.py`, `provider-support.md`, `driver.py`, `relay.py`,
`identity.py`, `verify.py`, `README.md`. Owned evidence changes are live-matrix,
student-review, implementation-status, provider-prep command/control machinery,
command outputs and this handback. Exact commit path lists identify all retained
receipts; no whole-tree staging or binary relabeling was used.

[Provider support contract](support/provider-support.md) contains the concrete
schema, fixed origins/routes, future commands and complete A/B/C/D action/observable
summary. [Live matrix](live-matrix.md) retains the full prompts/schedule and local
fault/lifecycle controls. Key fields (schema only):

| Vendor | Only selected settings field | Origin / one GET |
| --- | --- | --- |
| Anthropic | `directClaudeAPIKey` | `https://api.anthropic.com/v1/models?limit=1000` |
| OpenAI | `directOpenAIAPIKey` | `https://api.openai.com/v1/models` |
| Gemini | `directGeminiAPIKey` | `https://generativelanguage.googleapis.com/v1beta/models?pageSize=1000` |

Keys stay in the parent relay's memory. Children get a dummy loopback token and
no ambient credentials. Local tests use synthetic files/keys only. Fixed header
mapping and safe exception handling reuse permitted preceding Chapter9 support;
current official primary docs were checked for discovery/header mechanics (links
in the contract). Those public documentation reads were not API discovery.

Generation stays Messages/Chat Completions/generateContent, selected model only,
max output4096. Discovery is one request/vendor including failure; no additional
page/probe. Anthropic/OpenAI IDs must be explicitly selected from that actual
receipt. Gemini must actually return `models/gemini-3.8-flash` supporting
`generateContent`; no older substitute. None has been discovered or accessed here.

Budget remains A6/B4/C6/D6 =22/vendor,66 maximum; discovery3 maximum total.
Before-send journal accounting includes refusal/redirect/exception/timeout/cancel.
120s/attempt,600s/row, no redirects/retries/fallback. Literal raw/request bytes are
preserved unless credential redaction is needed, which is marked and invalidates
any exact-original-byte claim. Keys cannot pass to client output before sanitizing.
HTTP headers/Location/exception text are omitted or fixed safe values. Local D
seed exchanges remain separately labeled; proxy spending is separate from actual
CLI raw0/effective16 and GUI1 policy. Actual human PTYs, headed browser/speech
instrumentation and public two-Agent consumer remain live requirements.

35 local Python checks passed at the initial support freeze. They cover original
budget/relay/identity controls plus all-route synthetic-key success/HTTP refusal/
redirect, split-key byte and encoded-form redaction, captured/inbound headers,
credential-bearing URL/body refusal, transport exceptions, fixed model/routes,
one-page discovery and bound selection, header drips/DNS-wait deadline, child-env
isolation and split-source/build/launch refusal. Python compilation writes no
bytecode. [Command ledger](provider-prep-commands.jsonl) retains all commands and
outputs. First run failed on a closed-socket timeout after complete discovery;
fixed completion handling and explicit outcome assertions, retained original.
Intermediate controls passed38/33/34 cases as coverage evolved; final initial set
has35 without duplicate imported base-test execution.

Final deadline refinement: all13 affected provider checks pass, including the new
remaining-time handshake control. The prior35-check suite and22 actual-build
identity refusals retain their065a6a8 association. No unaffected Go or broad suite
was repeated. All22 identity mutations began with a valid actual build parent and
refused before derived writes; the launch was explicitly synthetic and never
executed. Results: [identity controls](provider-prep-identity/results.json).

Binding strategy: preserve complete historical build_sources and each actual
Go build association at57d4aac, verify current runtime and compiled consumer are
byte-identical, and bind interpreted support separately. Changed/added/removed
compiled inputs or module selections require a new real build; no map omissions
or current-source substitution is permitted. New launches include both revisions.
Old receipts/bindings remain unchanged. No build is needed for these interpreted
support changes. Future compiled edits require the coordinator's serialized slot
and the actual `build-binding.py` vet/test/build command documented in the contract.

The current [final binding](provider-prep-binding-final.json) records runtime/build
57d4aac and support9822b2b, with complete preflight against actual retained binary
and dependency hashes. The earlier provider-prep-binding.json remains bound to
065a6a8; it is not rewritten. Actual source/build metadata and binary/dependency
bytes were read for identity verification only. Small alternate support inputs
for the byte-mutation control remain preserved; copied Go/module inputs were
renamed with `.fixture` suffixes after testing to avoid extra apparent modules.
The final9822b2b binding passed complete actual-build preflight and two focused
support-source/launch-support refusal controls before any derived write:
[final controls](provider-prep-final-binding-controls.json). These supplement the
22 controls of the unchanged split-identity implementation at065a6a8.

Read/exposure/assistance ledger is appended in student-review.md. Coordinator
reports narrowly adapted management138/138 and exact replay passing on57d4aac;
the original135/138 receipt remains preserved. Latest mapping diagnosis was unused
Part.Parts nilness, with no runtime repair requested. Final inbox reports deterministic
checks passing: narrowly adapted14/14 inherited response groups and19/19 intended
deletions, with joined70/70 source-bound result being finalized and original68/70
preserved. These are attributed inbox reports, not student checker/reviewer
inspection or a student rerun. Concrete support review and separate live release
remain required. No model/key capability is assumed. Stop after this preparation
handback; no question to Bill is needed.
