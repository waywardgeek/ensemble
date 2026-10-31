# Optional GUI library — Chapter 20

`Ensemble.Gui.csproj` references the core; the core does not reference this
assembly or ASP.NET. The existing GUI vocabulary remains in `Common`, with
functional partial implementations in `Server`. Static HTML, CSS and JavaScript
are embedded assembly resources, not required deployment-side source files.
The example under `Examples/Gui` supplies CLI/environment policy and owns its
Ensemble and Agent. No new package dependencies are required.

## Private loopback host

An outside .NET 10 ASP.NET consumer can reference this project and call:

```csharp
await Gui.RunAsync(agent, new GuiOptions(
    Port: 8097,
    SettingsPath: "settings.json",
    TracePath: "gui.log"), cancellation);
```

The operation borrows the Agent. It listens on loopback and joins hosting,
observation, sockets and admitted requests when cancelled. It does **not** shut
down or destroy the Agent. Its caller retains that duty. Relative state paths
resolve through the Agent's Sandbox. User-home defaults, process environment
choices and argument parsing belong to the executable. `AssetsDirectory` is an
explicit trusted disk override; otherwise resources come from the assembly.
A missing override asset returns 404, not embedded fallback.

## Caller-owned application

`Gui.ServeAsync(WebApplication, Agent, GuiOptions, CancellationToken)` mounts the
**whole interactive GUI** in a caller-owned ASP.NET application: embedded assets,
canonical replay, controls, observation, browser MCP relay, and account routes.
Call it before starting the application; registration occurs before its first
await. The returned task is the GUI lifetime, not a registration acknowledgment.
Surface an already-completed task before starting the server, and monitor it
alongside the application's lifetime so initialization failures are observed.
The caller must bind loopback on `options.Port` and keep the Agent alive until
that task finishes. One GUI mount per Agent/application is supported.

Set `BasePath: "/assistant"` to mount the page and controls there. Literal
slash-separated names are supported, not route templates. The bare prefix
redirects to its trailing-slash form; assets, account requests, and WebSockets
resolve relative to the page. This uses the explicit `BasePath` option rather
than a reverse proxy or `UsePathBase` deployment. The existing local-origin
checks are not a remote authentication scheme.

Authentication deliberately has one exception: `/auth/callback` stays at the
application root, even for a prefixed GUI. Credentials requires that fixed
registration path. The caller must reserve it; other account operations stay
under the prefix. No credential-validator escape hatch was added.

Cancelling `ServeAsync` closes GUI admission, stops observation, aborts its
sockets, and joins admitted requests and browser-client teardown. Its routes
then return 503; unrelated caller routes keep running. It does not stop the
application, destroy the Agent, or cancel shared credential attempts. The caller
owns pending-login cleanup when removing its callback endpoint. `RunAsync`
retains standalone hosting's pending-login cleanup and stops only the private
host it created. Neither mode changes the running Agent's funding selection.

`Gui.MapAssets(IEndpointRouteBuilder, directory?)` remains available for asset-only
mounts of the known shipped names, without exposing arbitrary disk paths. It does
not install controls, account routes, or a browser relay.

## Outside-consumer verification

The separate `GuiTests` executable consumes public APIs only. Publish it into an
owned temporary directory, then execute it from an owned empty working directory:

```sh
dotnet publish GuiTests/GuiTests.csproj -o "$OWNED_TEST_BIN"
(cd "$OWNED_TEST_WORK" && dotnet "$OWNED_TEST_BIN/GuiTests.dll")
```

The consumer creates and reports its own unique temporary working directory,
restores the caller directory on exit, and retains its files for inspection. It
runs without deployed `wwwroot`, checks embedded and overridden
assets, and runs two standalone hosting lifetimes with one borrowed Agent. Its
caller-owned application mounts the complete GUI beside an independent health
route. It verifies canonical replay, WebSocket controls, a real browser MCP
initialize/list/call exchange through ordinary Agent dispatch, and termination
of a pending RPC on cancellation. It checks account guards, local-only PKCE
preparation and wrong-state refusal at the fixed callback, and preservation of
the caller's pending attempt after GUI shutdown. No authorization URL is opened,
no token exchange or model request is made, and the test cancels its own attempt.
It verifies relay/client detachment, stopped-route refusal, caller-route
survival, and Agent usability before destruction.

This is not live-provider, OAuth, browser-rendering, audible-speech or
human-accessibility evidence. The separately recorded isolated browser exercise
covers mounted page rendering and controls only; see the top progress record.
The child workflows and scoped executable fleet grade are separately verified;
they do not add GUI credit to that grade. The reviewed source-only `ch20` endpoint
preserves this library alongside the core and consumers. See
`docs/edition-3-csharp-progress.md` at the repository root for the current boundary
and retained evidence.


## Chapter 26 crossover candidate (not replacement acceptance)

The candidate uses pinned local Bootstrap 5.3.8 (graphite/blue, dark default),
without another component library. `wwwroot/vendor/README.md` records upstream
URLs, exact hashes and the bundled license. Bootstrap owns tab behavior; the
small `layout.js` owns only pane sizing/navigation. The browser still uses the
existing socket protocol and durable settings acknowledgments. Selection changes
on replay; relay identity remains a separate server fact. Chat shows narrative,
reasoning and final prose; Tools/Media shows actual tool calls/results and media.
No browser Agent tree or transcript copy replaces the server-owned model.

### Build and foreground start

From `agent`, build the application with .NET 10. Use the
ordinary executable `./start` afterward. It resolves its assembly relative to
itself, keeps the caller's working directory for relative workspace arguments,
and `exec`s the host in the foreground. It never installs, rebuilds, kills a
listener, starts a timer, backgrounds a server or obtains credentials.

For the isolated teaching environment (substitute your own private paths):

```sh
mkdir -p .crossover/{home,tmp,cli,workspace,data}
env -i PATH=/usr/bin:/bin HOME="$PWD/.crossover/home" \
  TMPDIR="$PWD/.crossover/tmp" DOTNET_CLI_HOME="$PWD/.crossover/cli" \
  DOTNET_ROOT=/Users/bill/.local/share/ensemble-dotnet10 \
  DOTNET_CLI_TELEMETRY_OPTOUT=1 DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1 \
  DOTNET_GENERATE_ASPNET_CERTIFICATE=false \
  /Users/bill/.local/share/ensemble-dotnet10/dotnet build Examples/Gui/Gui.csproj
# Use the same sterile environment prefix for the foreground command:
# ./start --port 18200 --workspace "$PWD/.crossover/workspace" \
#   --data-dir "$PWD/.crossover/data" --vendor anthropic --model MODEL
```

Supply only an explicitly authorized provider configuration when actually making
requests. These instructions do not authorize live inference or credential
migration. A second start on an occupied port fails safely and leaves the first
listener alone. Invalid configuration/bind diagnostics are classified without
printing raw nested exception text. Cleanup failures cannot replace the primary
failure; they are reported separately. Partial-startup cleanup and normal SIGINT
were exercised; a second deliberately injected cleanup exception was not.

### Complete Job output

`GET jobs/output?agent=ADDRESS&handle=HANDLE` is mounted under the same BasePath
and request-lifetime guard as the GUI. It requires `X-Ensemble-Output: 1`, checks
origin/fetch-site, and rejects unknown query keys, duplicate identity fields,
invalid handles, running Jobs and unknown/destroyed owners. The browser uses a
same-origin guarded fetch and checks declared and received length against a
16 MiB limit. It refuses oversized output rather than silently clipping it,
then creates a local Blob download. Ordinary anchors do not access this route.

The endpoint resolves a *current* Job through its owning Agent and opens only
that Job's stored output path through the owner's `Sandbox.OpenRegularRead`.
It does not accept a browser path, create a second Sandbox, advance the normal
Job observation cursor, or trust a path shown in a tool preview. It uses a
bounded copy and links mounted shutdown with request cancellation for the entire
response. Returned streams belong to that request and are disposed there.

The limit bounds browser allocation (temporary chunks and Blob storage can add
overhead); it is not a limit on the actual Job output file. Use the local owned
workspace file for larger output. Rooted Sandboxes retain their existing root-relative open
boundary and admissible inward-link resolution. Unconfined Sandboxes retain their deliberately broader symlink policy.
This is ownership-based access through the existing Sandbox policy, **not**
authentication that bytes were never replaced by a hostile local filesystem
actor. FIFOs/nonseekable files refuse without blocking; missing/unreadable/rooted
escape cases refuse without returning filesystem paths or canary contents.

### Reproducible local checks

Build `GuiTests/GuiTests.csproj` using the sterile SDK environment above. The
existing test executable's default run and `--goal-tests` and
`--memory-settings-tests` switches (model controls are in the default run) exercise real mounted
consumer/control paths. See `GuiTests/Program.cs` for exact supported switches.

For browser checks, start `GuiTests/bin/Debug/net10.0/GuiTests.dll
--crossover-host` from `main`. It owns loopback 18202/18203, creates generated
fixture credentials and real Agents/Jobs, and implements a local streaming
provider boundary. It mounts the packaged assets at `/assistant/`.
Run `node GuiTests/crossover-browser.cjs` with the supplied read-only Playwright
and Chrome paths. It writes evidence under `.crossover/evidence` and closes its
private browser context. Run `node GuiTests/output-cancellation.cjs` last to
exercise aborted/slow output readers and stop the fixture, or touch
`.crossover/runtime/stop` and wait for the fixture process to exit. Only after
that clean stop may its private `.crossover/runtime` be removed for a fresh run.
Do not delete a live fixture workspace. The fixture is macOS/Unix-specific.

`python3 GuiTests/startup-test.py` checks the real launcher from a private
working directory (ports 18230/18231). `python3 Trial/lifecycle-test.py` tests
one-shot recovery against real disposable servers on 18220–18228. See
`Trial/README.md` for the independent controller, explicit arming/switch signals,
launchd recipe, original graceful-only shutdown and restricted candidate cleanup.
No working server was switched, and the launchd recipe was not installed/tested.

The browser tests use DOM/focus/geometry, real sockets, clipboard/download bytes,
network observations and screenshots. Screenshots were captured, not visually
inspected as pixels. Native assistive technology, aesthetic acceptance, audible
speech and live provider/OAuth/cache behavior remain outside these local checks.
Bill determines beauty, accessibility and replacement acceptance; this remains
an uncommitted candidate for independent review.

### Compact conversation and flexible panes

The composer starts at one visible line beside the supplied Send icon and grows
with newlines or wrapped text up to five lines; longer input scrolls inside it.
It shrinks when text is removed or sent and refits when pane width or font size
changes. Its placeholder and accessible name follow the selected Agent.
Send stays available during a turn. The browser posts an ordinary Message; the
Agent decides how to handle it when dequeued. There is no GUI Hint mode.

Spacing is reduced without changing font metrics. Pane widths are transient,
not Agent policy: keyboard Home/End or pointer dragging can give either side most
of the workspace, retaining 160px minimum panes and recalculating on window/font
changes. The running host must be rebuilt/replaced to receive embedded asset
changes; editing these source files does not update an already running host.

### User names

Settings → User stores a name and description in the existing GUI settings file.
Save is acknowledged only after settings-file replacement succeeds; the browser uses the
acknowledged name above human messages and in accessible labels. Agent labels
come from the actual Agent tree, not the model or an additional display identity.

At startup the GUI example passes these saved fields as `$HUMAN_NAME` and
`$HUMAN_DESCRIPTION` to skills. `$AGENT_NAME` is reserved and derived from each
Agent. Unconfigured human names display as “Human.” Substitution is one pass:
user text containing `$TOOLS` is data, not another substitution instruction.
Already-rendered skills and restored conversation History are not rewritten by
a User edit. Restart loads the saved values for subsequent skill activation;
it does not retroactively change restored instructions.
