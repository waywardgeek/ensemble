# Provider support contract — prepared, not live-released

Runtime/CLI/GUI/public consumer retain actual build revision
`57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4`. This change is interpreted support only.
No credential or external API operation is permitted in the preparation phase.
`--mode live` and `discover.py --live` are explicit future entrypoints, not evidence
that coordinator release, a usable key, or any model availability exists.

## Credential fields and fixed routes

The permitted preceding Chapter9 terminal support supplies the settings schema.
The loader selects only the needed top-level string in `~/.cr/settings.json`:

| Vendor | Field | Header | Fixed origin | Single discovery GET |
| --- | --- | --- | --- | --- |
| Anthropic | `directClaudeAPIKey` | `x-api-key` plus `anthropic-version: 2023-06-01` | `https://api.anthropic.com` | `/v1/models?limit=1000` |
| OpenAI | `directOpenAIAPIKey` | `Authorization: Bearer …` | `https://api.openai.com` | `/v1/models` |
| Gemini | `directGeminiAPIKey` | `x-goog-api-key` | `https://generativelanguage.googleapis.com` | `/v1beta/models?pageSize=1000` |

Keys remain solely in the Python parent's memory. CLI/GUI/public-consumer children
receive only `local-relay-placeholder` and the loopback relay URL; they inherit no
ambient credential variables or home-directory setting. No secret argv, prompt,
launch environment, stored config, request headers, arbitrary response headers,
Location header, response reason or exception text is recorded. The settings file
is bounded to1MiB and parsed in memory, returning only the selected nonempty key
field; no printing, enumeration/reporting of unrelated fields, OAuth or ambient
fallback occurs. Malformed/missing values produce a constant safe diagnostic.
Local tests use explicit synthetic files and clear ambient environments, never the
real settings path. Provider key syntax is bounded ASCII letters/digits/`.`/`_`/`-`
(8–512 bytes); an incompatible field is a blocker, not silently transformed.

Official primary references consulted for mechanics only:
[Anthropic Models](https://platform.claude.com/docs/en/api/models/list),
[OpenAI Models](https://developers.openai.com/api/reference/resources/models/methods/list),
[Gemini Models](https://ai.google.dev/api/models), and
[Gemini API-key header](https://ai.google.dev/gemini-api/docs/api-key).
These support the listed routes, header forms and returned identifier fields.
They do not establish this account's access or replace actual bounded discovery.

One GET per vendor includes failures. `has_more`/`nextPageToken` are recorded as a
Boolean, never followed. Anthropic/OpenAI selection is an explicit identifier from
that single recorded page, suited to the existing Messages/Chat Completions route;
there is no automatic newest/default model or capability probe. Gemini must return
exactly `models/gemini-3.8-flash` with `generateContent` among its supported methods.
The runtime's resolved mapping is `gemini-3.8-flash`. This is the requested target,
not a claim it was discovered. Missing target/access/compatibility or an incomplete
page is reported honestly; no older model, alternate endpoint or extra request.

Live generation allows only `/v1/messages`, `/v1/chat/completions`, or the selected
Gemini model's `:generateContent`/`:streamGenerateContent?alt=sse` routes. Body model
must match selection on the first two. Fixed TLS origins, no proxy-environment
inheritance, SDK, redirects, transport retries, pagination or model fallback.
Local fixture mode remains literal HTTP127.0.0.1 and never calls the key loader.
The explicit Python-only fixture adapter accepts synthetic keys for controls;
the live driver exposes no custom external origin or key argument.

Responses are filtered before both child delivery and evidence writes, including
keys split across chunks, raw/percent-encoded/JSON-unicode-escaped token spellings.
Ordinary bytes are unchanged. A redaction is marked; such a run cannot establish
exact original provider replay equivalence. An incomplete possible key prefix is
discarded on a failed transport and marked. Captured response headers are replaced
with a fixed safe JSON/SSE content type. Secret-bearing request bodies are refused
before recording/transport. Response bodies are bounded8MiB and discovery2MiB.
The redactor is for configured credentials, not a promise to sanitize arbitrary
sensitive conversation text or arbitrary transformed/encrypted encodings.

## Schedule and bounded spending

The complete feature/action/observable matrix and exact prompts remain
`../live-matrix.md` and `schedule.json`. No feature was dropped:

| Row | User/public path and evidence | Per-vendor generation ceiling |
| --- | --- | --- |
| A | Two actual human PTYs: edit/remember, `/session`, checkpoint/history/quit, fresh process recall with unchanged file and no tool replay | 3+3=6 |
| B | Headed browser on A; saved-after-applied ack, history/job access, current policy1 versus recorded0/effective16, font16→18, actual tool turn/round limit, reconnect/Page/tab close, terminal EOF, SIGTERM/restart, speech admission/queue/pause probes | 4 |
| C | Compiled public two-Agent independent outcomes; export old/new with real tail; close then snapshot import; immutable origin, real history_unavailable, continuation; offline latest/older/null render/usage/Skills/watch/request equality | 2+2+2=6 |
| D | Retired/reloaded narrow Skills, actual PTY tool turn; explicitly local two-exchange limit seed; relocated unchanged catalog, public resume and real next attempted call consuming once | 3+3=6 |

22 generation attempts/vendor,66 overall; one discovery/vendor,3 overall.
Every outgoing attempt reserves/fsyncs its slot before transport, including
HTTP refusal, redirect, exception, timeout or cancellation. Row ceilings6/4/6/6,
step ceilings,4096 output tokens,120s/attempt and600s/row remain enforced; no
automatic retry or input resubmission. The attempt deadline is also bounded by
remaining row time. A watchdog interrupts header/body drips; socket timeouts cover
connect/write/read. DNS resolution has a bounded wait: a late daemon can finish name resolution
but owns no socket/HTTP request; one returned address gets one connection attempt.
No alternate resolver/probe/address retry is added. Driver process lifetime is
bounded separately. Any setup/transport/model failure stops for an honest result;
unused slots do not authorize an improvised rerun. Spend caps are not runtime
policy: CLI remains raw0/effective16, GUI demonstrates1; public paths use their
existing deliberate policy values.

D-prepare, D-seed and offline are local-only driver steps. The seed's two loopback
exchanges/provider are separately labeled and are not real-provider usage. Offline
comparisons invoke no generation. Real PTYs/browser/public interfaces remain
required; local controls never fulfill those live requirements. Historical opaque
material without a real instance remains deterministic coverage, not fabricated
live evidence. All additional corruption/lifecycle/bounds controls in the matrix
remain independent local checks with zero paid requests.

## Runtime and support identities

New optional `support_revision` identifies interpreted support separately.
`source_revision` stays the actual Go build revision. `build_sources` retains the
complete original historical source map, including old support, and is checked
against that immutable revision. Current `sources` combines runtime at that build
revision and all support at `support_revision`. Every non-support input and all
`support/consumer/` files must be identical to the actual build's inputs; modules,
build associations, binary hashes and dependency metadata retain their original
checks. Added/missing/changed compiled inputs refuse until a real build is supplied.
All support, schedule, catalog/bindings, browser dependencies and launch bindings
remain complete. Launch receipts explicitly include both revisions. Older original
launches/bindings are never edited or rebound. As before, preflight requires the
selected owned source tree, not a later mismatching working tree.

After committing support, without Go or credentials:

```sh
python3 -B evidence/ch10/support/bind-support.py --repo REPO \
  --parent evidence/ch10/retained-repair-build-watch/binding.json \
  --revision EXACT_SUPPORT_COMMIT --output NEW_BINDING.json
python3 -B evidence/ch10/support/identity.py NEW_BINDING.json --repo REPO
```

These commands run from main. The original CLI/GUI/consumer paths remain explicit;
no binary copy/rebuild is needed for these Python/docs edits. If a compiled input
changes later, stop and obtain the compiler slot, then use the existing
`build-binding.py --repo REPO --revision EXACT_COMMIT --output NEW_DIRECTORY
--node NODE --chrome CHROME --browser-root BROWSER_TOOL_ROOT` (actual vet/test/build
and module metadata, no planned-binary association). No such Go command ran here.

## Future released invocation only

After deterministic clearance and concrete support review/live release, use a new
provider root R and fresh receipt path. Preflight runs before key loading:

```sh
python3 -B evidence/ch10/support/discover.py --live --binding NEW_BINDING.json \
  --repo REPO --vendor VENDOR --root R --output DISCOVERY.json
python3 -B evidence/ch10/support/driver.py --mode live --binding NEW_BINDING.json \
  --repo REPO --vendor VENDOR --model EXACT_RETURNED_ID --discovery DISCOVERY.json \
  --root R --run NEW_RUN --step A1
```

Select the returned ID explicitly, then use the same arguments for A2/B/C/D1/D2
as the schedule requires, with distinct run directories. Live has no `--origin`;
it validates vendor/model/binding and the same root's before-send discovery journal
before reading the key. D-prepare/D-seed/offline retain `--mode local --origin
LITERAL_LOOPBACK_FIXTURE` with the selected model and original provenance. The
browser remains pointed only at its actual printed loopback GUI origin. Catalog
bytes are checked before D key access. No command above was executed in live mode.
