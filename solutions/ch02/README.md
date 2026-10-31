# Chapter 2 — One Log, Three Vendors (C#)

C# 14 / .NET 10, standard library only. This is a bounded Chapter 2 solution,
not a migration or installation of the working Go Ensemble application.

## Progression from Chapter 1

Keep the owners and parent paths from `../ch01`; replace Agent's completed-pair
list with an Agent-owned History. The log records admitted input, attempts,
replies and failures. Unlike Chapter 1, a failed request does not erase its input.
Engine normalizes usage into response events; History derives cumulative usage
from them. There is no second counter or transcript.

Read Common's vocabulary and contracts first, then the owner constructors and
Destroy methods. Follow one appended event through History.Reduce before reading
the three renderers and parsers. Finish with the CLI, which connects these
operations without reimplementing them. The tests exercise these real components
and replace only the external HTTP service with a working loopback fake.

## Read the structure first

- `Common/Owners.cs`: owner state, parent references and callable operations.
- `Common/Values.cs`: immutable, explicitly value-like events, parts and views.
- `Common/JsonCodec.cs`: vendor-neutral JSON operation declarations.
- `Application/`, `Agent/`, `History/`, `Engine/`, `Serialization/`: implementations
  of those partial declarations, compiled together in `Ensemble.csproj`.
- `Cli/`, `GuiStub/`, `Tests/`: independent consumers of the same core library.

Ensemble owns Agents; each Agent owns one History and one Engine. Constructors
call parent add methods, which establish both ends. Parent remove methods detach
both ends before returning; the destructor does not finish their work. Refused
changes leave both ends untouched. Destroy cascades through strong children,
including when another child's cleanup fails. This chapter has no multiply owned
model object or weak child relationship requiring extra machinery. Values have
no holder back-pointers. Diagnostics is a borrowed framework boundary; destroying
Ensemble does not close the caller's console.

Engine reaches configuration, History and diagnostics through its parents.
History alone owns the canonical events and their derived Context. Render changes
neither. Opaque data needs exact provenance; JSON tool numbers stay exact.
Nullable usage counts distinguish unknown from zero. Logs use ordinary
flush-before-publication writes, not a promised crash-recovery or backup system.
There is no tool execution, streaming, turn scheduler, media loader or GUI here.

## Build and run

With a normal SDK installation, use dotnet. For Bill's approved private SDK:

```sh
export DOTNET_ROOT="$HOME/.local/share/ensemble-dotnet10"
DOTNET="$DOTNET_ROOT/dotnet"
"$DOTNET" run --project Tests/Tests.csproj -c Release
"$DOTNET" run --project GuiStub/GuiStub.csproj -c Release
"$DOTNET" publish Cli/Ch02.csproj -c Release -o /your/owned/output
```

Run these from this directory. DOTNET_ROOT is per-shell configuration, not a
global PATH change. Invoke /your/owned/output/ch02 chat, render LOG, or dump;
omit the command for JSON-lines input. Only {"user":"..."} and
{"ephemeral":"..."} are accepted. See Chapter 2 for the environment and schema.
Never put credentials, raw logs or build output in Git.

The existing Go grader accepts the published executable directly, not the C#
source directory. Local fakes validate protocol behavior, not provider access.
Real-provider exercises remain unauthorized under the exhausted prior allowance.

## Verification scope

- 15 real-component test groups passed, including direct parent add/remove
  postconditions, refused-operation preservation, replay, all state/event pairs,
  exact JSON numbers, redaction, opaque provenance and local HTTP behavior.
- Existing Edition 2 Chapter 2 grader: 100/100, passed:true, exit zero on the
  corrected ownership build. The first 59/100 result exposed codec compatibility
  omissions; they were corrected in the solution, not hidden by grader changes.
- Existing Edition 3 Chapter 1 grader: 100/100, passed:true, exit zero on that same
  build, including actual separate 62-second delayed-header and delayed-body
  phases. The earlier Edition 2 Chapter 1 report was ordinary parity only.
- Actual CLI chat/EOF/dump/render and ephemeral replacement/clearing passed
  against owned local fakes for all three vendor protocols. Three-turn requests
  carried 1/3/5 messages; offline render was deterministic and required no key.

The teaching pass expanded control flow, JSON construction, and comments without
adding a new service layer or feature. The current progress record in
`docs/edition-3-csharp-progress.md` records the affected checks and carried evidence.
These results do not establish real-provider integration or deployed behavior.
