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

## Authorized live status check

Bill subsequently authorized the test. A disposable real Ensemble Agent loaded
this skill with an explicit path to the existing credential file (no key copying)
and issued exactly one moltbook_status request. It returned account coderhapsody,
karma175 and unread_notifications40. Only these three fields were printed; no
feed/comments were browsed, notifications marked read, or account state mutated.
The skill unloaded, MCP teardown joined, and the consumer exited zero. No model
request or running-host restart occurred.

Evidence directory is named by /private/tmp/cadence-moltbook-live-latest;
build.log and status.log retain the result. This supersedes the earlier untested
credential/API boundary, but verifies only status, not every live endpoint.

The running Cadence actor still lacks this skill: definitions are captured at
Agent construction and the primary ensemble skill does not yet list moltbook.
No primary-skill grant or launcher was changed. Live-session activation requires
a separately agreed configuration update and Agent recreation/restart; copying a
file or testing another Agent does not grant the current one new tools.

## Restart configuration authorized

Bill asked to add the skill and will restart himself. Added moltbook to the
primary ensemble skill's loadable-skills and explicitly selected the existing
/Users/bill/.config/moltbook/credentials.json path in its stdio argv. This is a
path only, not a copied key. The active launcher already reads repository skills,
so no launcher, binary, credential file, or running process was changed.

The offline consumer now replaces the one argv declaration structurally and
refuses an unexpected declaration count. Its injected API and nonexistent fixture
credential path prevent the configured live path from leaking into offline tests.
Clean rebuild and real skill/Job/unload integration pass again. No further live
request occurred. Restart and actual availability in the restored actor remain
to be observed; configuration is not proof that it has loaded. Discovery alone
performs no API request, and loading does not authorize automatic posting.

## Post-restart interpreter correction

The first live tool status call failed with a network error, before any successful
read or write. Credential-free HTTPS probes established that /usr/bin/python3
returns HTTP200 while /opt/homebrew/bin/python3 fails issuer verification. The
earlier live consumer pinned the former, but source MCP argv used PATH python3.
Pinned the host skill command to the tested /usr/bin/python3; no TLS bypass or
trust-store changes. Existing Agent definitions retain the old command until
restart. Lesson: test the exact launch interpreter/environment, not only the
same script. No public post or account mutation occurred.
