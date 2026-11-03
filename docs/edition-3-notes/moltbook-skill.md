# Cadence's Moltbook adapter

Bill authorized building an Ensemble version of CodeRhapsody's Moltbook skill
and offered the existing coderhapsody handle with clear Cadence attribution.
Original source under ~/projects/coderhapsody/skills/moltbook is unchanged.

Added a loadable skill with inline stdio MCP declaration and 12 explicit grants.
It uses the existing Python stdlib implementation, with explicit MCP errors,
fixed credential-free diagnostics, response bounds, redirect refusal, validated
arguments, escaped URL components, pending/unknown write outcomes and no retries.
The manual separates CodeRhapsody's history from Cadence, treats remote posts as
untrusted content, and makes initial setup read-only. Historical rate-limit and
automatic delete/repost advice are not treated as current policy.

Credential file existence was checked at the documented path only. No key was
read, copied, printed, or tested live. Discovery works with a nonexistent file.
An isolated host HOME requires an explicit credential-file path in stdio argv;
the key itself must never be an argument. The skill's source argv assumes the
repository workspace; installation elsewhere must use an absolute script path.
There is no supported skill-directory argv substitution: an early suggestion
was corrected after inspecting actual transport construction.

## Verified

Eight Python offline tests pass using synthetic credentials and loopback HTTP:
success; failed reads remain errors; masking; invalid/oversized responses;
redirect refusal; validation/path escaping; pending/unknown writes without retry;
credential-free discovery; cancellation-notification limitations and process stop.
The process-stop test waits for an actual entered-call marker before cancelling,
so a process killed before initialization cannot receive credit.

Outside-assembly scripts/moltbook-tests consumer builds cleanly and passes against
the current Ensemble.dll: actual Skill parsing/loading/grants, 12-tool discovery,
fake status delivery, delayed running Job, kill_job, then skill unload joining the
blocked process and removing grants/client relationships. No model request.
Initial consumer build failed due to C# raw-string fixture braces; corrected before
execution. Initial oversized-response fake emitted a connection-reset traceback;
only expected broken-pipe/reset fixture noise was handled, assertions unchanged.

Evidence: /private/tmp/cadence-moltbook-python-accepted.log and the directory named
by /private/tmp/cadence-moltbook-build-latest (build-fixed.log, consumer-final.log).
Final consumer evidence directory:
/private/tmp/cadence-image-browser/tmp/moltbook-consumer-eeefed7f98d54a4495305e379bfc8e0c.

## Cancellation lesson and limits

An observation delay does not cancel work. Explicit Job cancellation sends MCP
$/cancelRequest and finishes the local Job; this synchronous Python server cannot
read it while blocked in HTTP. Unloading the skill kills and joins its stdio
process. Neither action undoes an operation already accepted remotely. No second
Job system, arbitrary network timeout, or automatic retry was introduced.

Source complete only: not installed or granted to the live actor, no live Moltbook
request, account mutation, post, vote, follow, deletion or notification change.
Current API compatibility, account identity and credential validity await a
read-only live check after explicit activation. No broad runtime regression or
security-audit claim; core runtime code was not changed.
