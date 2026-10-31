# Chapter 1 — one conversation loop in C#

This is the independent Chapter 1 endpoint, not a reduced copy of Chapter 2.
It uses C# 14/.NET 10 and the standard library only.

## Read the code in this order

1. `Common/Values.cs` defines immutable values, exempt from holder tracking.
2. `Common/Owners.cs` declares the actual objects, persistent state and operations.
3. `Application/Ensemble.cs` owns Agents and the diagnostic destination.
4. `Agent/Agent.cs` retains completed pairs and owns one Engine.
5. `Engine/Engine.cs` performs one HTTP exchange and counts measured usage.
6. `Cli/Program.cs` constructs the application and owns the input loop.

Parent add/remove methods finish both sides of a relationship. `Destroy` cascades
through strong children and asks parents to detach it. Configuration lives with
Agent; usage lives with Engine. HTTP does not hold a copied configuration or
logger. Request messages are temporary snapshots, not a second conversation.
Only successful exchanges publish new conversation pairs. Measured usage from
an unusable answer still counts. Calls are serial in this chapter.

## Build and run

With .NET 10 on PATH, from this directory:

```sh
dotnet run --project Tests/Tests.csproj
dotnet publish Cli/Ch01.csproj -c Release -o /your/owned/output
/your/owned/output/ch01 chat
```

For the private SDK, use `/Users/bill/.local/share/ensemble-dotnet10/dotnet` and
set `DOTNET_ROOT=/Users/bill/.local/share/ensemble-dotnet10` when launching the
apphost. Use an owned `DOTNET_CLI_HOME`; suppress development-certificate creation
with `DOTNET_GENERATE_ASPNET_CERTIFICATE=false` for SDK commands. No global PATH,
package installation, application installation or restart is needed.

Set `ANTHROPIC_MODEL` and `ANTHROPIC_API_KEY`. `ANTHROPIC_BASE_URL` defaults to
Anthropic; tests must explicitly set it to their owned local fake. Dummy keys
alone are not a network boundary. Do not run a real-provider chat without the
appropriate account and spending permission.

No argument selects JSON lines: `{"user":"question"}` in, one
`{"assistant":"answer"}` out, then cumulative `usage` at normal EOF. `chat`
selects plain questions and answers, with prompts and usage on stderr. Errors
exit nonzero without a false normal-EOF record. Long lines are not truncated.
Generation has no arbitrary timeout; library callers can cancel the request.

## Verification boundary

The four real-component test groups passed, including direct relationship
postconditions, staged history, multi-block responses, measured failed-answer
usage, safe errors, no retries, and caller cancellation after server admission.
The published executable passed Edition 3 Chapter 1: 100/100, `passed:true`,
exit 0, including separate real 62-second header and body delays.

An actual local-fake CLI exercise covered three-turn chat, normal EOF, a
100,000-character line, an unterminated last line, malformed input and a real
stdin read error. This does **not** establish real-provider integration; the
paid-provider exercise remains incomplete because its allowance is exhausted.
The single project progress record links the saved results. Comments were
improved after the behavioral checks without changing executable behavior.
