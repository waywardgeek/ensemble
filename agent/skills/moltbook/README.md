# Moltbook for Cadence

Adapted from CodeRhapsody's Python MCP skill at Bill's request. Its original files
are untouched. New posts identify Cadence using the existing coderhapsody handle;
CodeRhapsody's earlier contributions and experiences remain theirs.

## Runtime and installation

No pip packages are required. The skill uses Ensemble's ordinary stdio MCP host,
explicit tool grants, Jobs, and joined teardown. It adds no runtime scheduler or
second Job system. It is a trusted network extension, not a shell-sandbox grant.

The checked-in SKILL.md is loadable from an Ensemble repository workspace:
`python3 agent/skills/moltbook/moltbook_server.py`. For another workspace, set the
stdio argv to the absolute script path. The primary skill must explicitly list
`moltbook` in `loadable-skills`. Copying files alone grants no tools. Loading
performs discovery only; it neither reads credentials nor calls the network.

The server defaults to `~/.config/moltbook/credentials.json`, resolved against its
actual inherited HOME. Ensemble hosts may use an isolated HOME. In that case,
append `--credentials` and the absolute path to the existing credential file to
MCP `args`. That argument is a path, never a key. Do not copy the credential file
into this repository, the skill folder, chat, or logs. No credential search,
login, registration, refresh, environment-key fallback, or live probe is done by
these tests. The API origin is fixed at `https://www.moltbook.com/api/v1`.

Installation into a running actor's discovery configuration is separate from
building this source. No running host or original CodeRhapsody setup is changed.
After explicit activation, use status for the first read-only live check and
confirm the account identity; fixture success is not proof of current API access.

## Error and operation boundaries

Read tools return MCP `isError` on HTTP, transport, malformed JSON, or API errors.
They do not turn errors into empty feeds or zero notifications. Raw error bodies
and arbitrary exception text are withheld. Responses are bounded to 2 MiB;
credential input is bounded and validated. Redirects are refused before forwarding
a bearer credential. Tool arguments are validated, and path components escaped.
Discovery and credentials are separate; a missing key does not break discovery.

There are no automatic retries. A post/comment response without an ID reports an
unknown outcome. A challenge or pending ID is not a publication confirmation.
Never delete/repost or repeat an uncertain write merely to get another challenge.
No fixed historical rate-limit claim is treated as current service policy.

An observation delay returns a running Job; it does not cancel work. Explicit
`kill_job` sends MCP cancellation and finishes the local Job. This synchronous
server cannot read that notification while an HTTP call is blocked. **Unloading
the skill kills and joins the stdio process**, including a blocked call. That
stops local execution, not an action already accepted by the remote service.
A later load starts a fresh process; it must not replay uncertain writes.

## Offline verification

From the repository root:

```sh
python3 -m unittest discover -s agent/skills/moltbook -p 'test_*.py' -v
```

Tests use synthetic credentials and loopback HTTP. They cover successful status,
read errors, masking, malformed/oversized replies, redirect refusal, validation,
path escaping, pending/uncertain writes, no retry, credential-free discovery,
and stopping a process proven to be inside a blocked call.

The outside-assembly consumer in `scripts/moltbook-tests` loads the actual skill
through Ensemble's public Agent API, injecting only the API response boundary.
It verifies grants, all 12 tools, results, running Jobs, explicit cancellation,
and joined unload of the blocked server. Build against an already built
Ensemble.dll using .NET10 and an empty NuGet source configuration:

```sh
dotnet build scripts/moltbook-tests/Consumer.csproj \
  --configfile /path/to/empty-sources.NuGet.Config \
  --artifacts-path /owned/temp/artifacts \
  -p:EnsembleAssembly=/path/to/Ensemble.dll
dotnet /owned/temp/artifacts/bin/Consumer/debug/Consumer.dll \
  "$PWD/agent/skills/moltbook" /absolute/path/to/python3
```

The consumer writes only a new temporary evidence directory and makes no model
or Moltbook request. Fixtures must exit cleanly; discovery is not live acceptance.
