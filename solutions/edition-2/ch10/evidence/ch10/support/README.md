# Chapter 10 evidence support — preparation checkpoint

This is support source, not a live receipt or an accepted chapter. Runtime/build
revision remains57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4. The preceding local phase
built the public consumer and exercised CLI/browser/public paths; source-specific
results and limitations are in ../retained-repair-handback.md. This preparation
adds an explicit provider adapter but executes Python/JS localhost controls only.
No credential read, discovery or real-provider call is released. The independent
grader owns Go compilation. See [provider-support.md](provider-support.md) for
separate source bindings, exact key fields, routes and later-release commands.

## Components and ownership

- `schedule.json`: exact prompts, row/step ceilings and current versus historical
  policy. A1/A2/D1 CLI policy is raw0/effective16. Proxy spending caps remain3.
- `relay.py`: one provider-owned durable admission journal, file lock across
  restarts/processes, step/row/provider ceilings, discovery counter and row deadline.
  Reserve/fsync occurs before upstream transport. No retries, redirects or paging.
  Per-request deadline120 seconds; input2MiB/output8MiB are support ceilings. It
  retains request/response bodies and transport outcomes, with header-secret
  redaction explicitly marked if necessary. It never reads a credential file.
- `driver.py`: a deliberately driven `/usr/bin/script` PTY for CLI/GUI or the public
  consumer. No prompt auto-send/retry. It accepts literal
  loopback upstreams in default local mode. Explicit live mode requires fixed provider
  selection plus bound discovery. Process600-second deadline and relay spending caps differ
  from the runtime's policy. It excludes ambient provider settings from children.
- `browser.mjs`: actual browser keyboard/pointer actions, DOM/screenshots and raw
  socket frames with connection identity. A recorded bootstrap suffix exposes the
  actual BrowserApplication, measures service/native speech admissions, reads Page
  queue/pause state, and records Page-applied session changes before saved acks.
  No speech call is manufactured and no hearing is claimed. Driver failure stops
  rather than resubmitting. Second browser tab close, logical Page close/reopen,
  socket reconnect, terminal EOF and process restart are distinct operations.
- `consumer/`: separate public module, no internal imports. C opens two Agents,
  records independent completion outcomes, exports old/new boundaries, closes
  owners before importing, checks genuine pre-origin refusal, captures real origin
  identity and once-restored usage. `offline` uses small separate branches with
  the full log plus latest/older/null-state checkpoint, compares render/context/
  usage/Skills/history/request bytes and normalized public watches, without a
  Submit/Ask. D prepares retired/reloaded skills, seeds pending limits through
  two actual loopback fixture exchanges (same selected wire/model, explicit
  provenance), and records subsequent once-only consumption before the first call.
  Existing binaries and local integration are retained at their actual revisions.
- `identity.py`: complete immutable own source/support/module/catalog identities,
  actual executable identities/build associations and full browser dependency tree.
  Working runtime and interpreted support must match their separate revisions;
  extra/missing source refuses. Historical compiled inputs may not change.
  An explicit repository root is required. Alternative binaries must hash-match.
- `provider.py`/`discover.py`: fixed API-key routes, one discovery attempt, no
  automatic model selection or credential propagation to child processes.
- `bind-support.py`: binds new interpreted support to retained actual builds;
  any compiled input/module change requires a separately released real build.
- `build-binding.py`: later serialized real build/vet/test commands, module graph,
  `go version -m` and actual hashes. Each command receipt is flushed immediately.
  It produces no binding unless real commands and final preflight pass. Planned
  binaries cannot be labeled built. Retained older binaries are not reused as
  repaired-source evidence.
- `verify.py`: validates identities and original-file map before derived writes;
  seals original receipts separately and reports limited observations, never a
  full acceptance verdict. `local-fixture.py` is a bounded synthetic three-wire
  backend for later integration; no outbound transport exists in it.

All mutable support counters/locks belong to the run/provider, not Ensemble.
No application runtime ownership, settings authority, writer or Context is added.
D's bindings are `{}` because the actual CLI exposes no scalar-binding selection;
markers occur literally in prompts. `catalog/` is small and independently bound.
To exercise relocation, copy only its two unchanged manuals into a new provider
scratch directory, close the old owner, and pass that directory for D2.

## Checks already possible without Go

From the outer repository root:

```
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s solutions/edition-2/main/evidence/ch10/support -p test_support.py -v
node solutions/edition-2/main/evidence/ch10/support/test-speech.mjs
node --check solutions/edition-2/main/evidence/ch10/support/browser.mjs
python3 -B -m unittest discover -s solutions/edition-2/main/evidence/ch10/support -p test_provider.py -v
```

Identity tests start from complete valid synthetic parents, then mutate source,
working source, individual binaries, support, module/dependency map, catalog,
build association or launch. They require specific refusal before a derived write.
No synthetic binary/build identity is a real build association. The real committed
source/binary/browser-dependency preflight and negative controls must also run
once actual binaries exist. Speech tests prove the probe's distinguishing behavior,
not the runtime/browser's absence of restored speech.

## Next serialized stage — commands to run only after build release

Use a new output directory outside support; preserve all existing binaries and
receipts. Substitute actual absolute repo/browser-tool/executable paths and the
committed support revision; do not use a moving HEAD as recorded identity.

```
PYTHONDONTWRITEBYTECODE=1 python3 evidence/ch10/support/build-binding.py \
  --repo REPO --revision SUPPORT_REVISION --output NEW_BUILD_DIRECTORY \
  --node NODE --chrome CHROME --browser-root BROWSER_TOOL_ROOT
```

The command above is relative to `solutions/edition-2/main/`. After it passes,
run the local fixture and the deliberately driven sequence below, all with
`--binding NEW_BUILD_DIRECTORY/binding.json --repo REPO --vendor PROVIDER
--model fixture --origin PRINTED_LOOPBACK_ORIGIN --root R --run NEW_RUN`:

1. `driver.py --step A1`, enter printed prompt once, observe response/file,
   `/session`, `/checkpoint`, `/history`, `/quit`. Retain closed A store originals.
2. `--step A2`, enter recall once, inspect commands and exit. Check unchanged file.
3. `--step B`, attach actual browser using `--seal-gui RUN` after its URL appears.
   This writes separate `browser-launch.json`; start browser.mjs with binding,
   repo, run and that exact origin. Each JSON line is a deliberate action:
   `checkpoint`, `reconnect`, `page-close-reopen`, `open`, `close-tab` (page1),
   `settings`, `font`, `policy`, `prompt`, `wait`, `inspect`. Inspect each result.
   Send EOF to attached terminal separately; observe continued browser usability.
   SIGTERM and restart server in a NEW_RUN against R/A; inspect persisted policy1
   and font18, zero speech admissions/queues/pause at the new snapshot. No second
   B prompt: the provider-root `B-prompt-admitted.json` guard survives that restart.
4. `--step C` records both public Agents and import continuation. `--step offline`
   uses the prior C run's `public/` as its input root and a separate output run.
5. `--step D-prepare`, then `--step D1` through actual PTY, then `--step D-seed`.
   Copy only the frozen small catalog to a new location, close owners, use
   `--step D2 --catalog RELOCATED_CATALOG`. Check original manuals, retired/new
   activation IDs and consumption evidence. Seed responses are synthetic; their
   usage stays explicitly labeled separately from real-provider evidence.

For each finished run, retain session/origin/checkpoint and scratch file hashes
alongside original terminal/socket/HTTP receipts before invoking verify.py --seal;
then invoke it without --seal for a new `derived/verified.json`. Never overwrite
originals with rendered/reconstructed requests. The local-fixture TTL is600 seconds
and request ceiling24; restarting that *local backend* does not reset the enforcing
provider budget. The proxy counts failed/canceled upstream attempts and denies N+1.

## Remaining gates

Existing actual builds, local paths and identity controls are retained in the
prior handback. The new provider adapter and split binding require coordinator
review and fresh local controls before live release; neither old binaries nor
planned binaries are relabeled. No Go command runs during this preparation.
Use bind-support.py for unchanged compiled inputs, as documented in
provider-support.md. Only a changed compiled input requires the serialized build
command above. Exact returned Gemini3.8Flash availability remains unverified.
No external endpoint, credential or actual discovery was accessed in this phase.
