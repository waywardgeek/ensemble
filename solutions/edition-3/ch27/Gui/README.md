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
