# Measure X document chat example

A separate, small website using Ensemble's public Agent API: one Bootstrap chat
pane, the airplane Send button, suggested questions, and visible application
instructions. Default model: **gpt-6-astra**, through the Responses API with an
explicit funding choice: API key for local fake-provider fixtures, or Codex subscription funding for the authorized launch. No funding fallback.

## Lifetime and privacy contract

The host reads a nonempty UTF-8 reference file once (maximum 2 MiB). Its immutable
text is shared; each accepted WebSocket connection gets a new Agent, seeded with
the full text as a human/data message before any model request. Loading the page
alone does not invoke the provider. The displayed instructions come from the same
string configured on the Agent, and document text is not put in those instructions.

All six memory settings, including Conversation, are disabled. Recall is disabled.
There are no installed tools, saved History, transcript files, visitor-selected
Agent IDs, cookies, or session-resumption endpoints. The Responses request uses
`store:false`; this is not a promise about the provider's other retention policies.
The host disables normal logging providers and never relays provider exceptions
to visitors. It does not import an existing Agent's conversation.

**Connection lifetime is the session boundary.** Refresh, reconnect, New conversation,
expiry, or disconnect disposes the old session; nothing automatically resends a
question. This deliberately also loses a conversation on a mobile connection drop.
If reconnect-resumption is wanted later, design an authenticated, expiring session
capability rather than accepting arbitrary Agent IDs.

The document is never summarized or dropped to make room. The host checks a
conservative serialized-request byte allowance, including a 65,536-byte reserve,
and closes a full session rather than compacting it. This is NOT token counting
or proof that a particular document fits a model's context limit. Provider usage
and actual model limits still need verification for the real document.

## Run locally

Use the repository's .NET 10 SDK and ordinary build/dependency policy:

```sh
dotnet build agent/Examples/DocumentChat/DocumentChat.csproj
dotnet run --no-build --project agent/Examples/DocumentChat/DocumentChat.csproj -- \
  --document /absolute/path/measure-x-reading-copy.txt
```

Supply `DOCUMENT_CHAT_API_KEY` through the process environment or deployment secret
configuration. Never put it into the page, a URL, source control, or a command-line
argument. The example refuses startup without it; do not use real keys for tests.
Open `http://127.0.0.1:18326/`, matching the default allowed WebSocket Origin.

For explicit subscription funding, use `--billing codex --codex-auth-file /absolute/path/to/auth.json`. The existing root-owned Codex reader reads credentials at send time; this mode ignores API-key environment variables, pins its endpoint, and enables streaming. No credential is copied into the page or installation.

Options are unique option/value pairs:

- `--billing`: `api-key` (default) or `codex`.
- `--codex-auth-file`: required only for Codex; never served to visitors.

- `--document`: required UTF-8 reading copy. This app does not extract or OCR PDFs.
- `--model`: defaults to `gpt-6-astra`.
- `--base-url`: optional operator-selected provider endpoint; local fake for tests.
- `--port`: loopback port; default 18326, range 1024–65535.
- `--origin`: exact external HTTPS origin, or HTTP loopback origin. No wildcard.
- `--session-seconds`: absolute connection lifetime, default 1200; range 2–3600.
- `--context-bytes`: host request allowance, default 1,200,000; range 32,768–8,000,000.

The server binds loopback. Initial limits are eight concurrent connections,
20 questions per session, 4,096 UTF-8 bytes per question, one queued question,
and a 90-second observation deadline per answer. Session teardown destroys the
Agent and joins actual work; cancelling observation alone is insufficient.

**Not yet public-deployment acceptance.** Origin checking is not authentication or
a complete abuse/cost control. Before exposing this through a website, settle TLS,
proxy access policy, global rate/spending limits, provider terms, operator monitoring,
and the document's real token budget. Do not assume the existing Gateway or live
Ensemble instances are configured for this app. Nothing is deployed by building it.

The Measure X reading copy identified in Proposition X has all 266 page markers,
but 20 pages have recovery-OCR warnings. Coverage is not transcription accuracy;
verify those pages against the official PDF before advertising reliable public use.
The example's tests use synthetic text, not that real document or its old dialogue.

## Verification

With an existing Playwright/Chrome installation and a built DLL:

```sh
DOTNET=/path/to/dotnet PLAYWRIGHT_MODULE=/path/to/playwright \
  node agent/Examples/DocumentChat/test-browser.cjs /path/to/DocumentChat.dll
```

The test owns a loopback fake Responses provider and temporary host/workspace. It
checks five real Agent requests: separate visitors, same-session follow-up, reset
and refresh isolation, full reference retention after deleting its source file,
absence of synthetic personal-memory markers/tools, `store:false`, exact visible
instructions, HTML escaping, and phone-width geometry. Browser, provider, and host
are joined. It makes no live inference and installs no browser or dependencies.

The lifecycle fixture below covers adversarial expiry/disconnect/shutdown, origin,
capacity, malformed/oversized input and accumulated-context boundaries. The browser
fixture also verifies that sanitized provider-failure text survives connection close
while the discarded chat is cleared (five successful requests plus one failed request).

Pending: real-document/OCR and actual model token/headroom/cost verification, plus
public-deployment controls and review. A direct client can supply the expected Origin
and repeatedly open fresh sessions: per-session limits are not authentication, a
per-person quota, or a global spending cap. No explicit provider output-token cap is
configured here. Simultaneous cleanup failures can mask an earlier cleanup exception;
the nested teardown attempts later cleanup but is not exhaustive fault-injection
coverage. Do not treat these bounded tests as a security audit.

### Additional lifecycle fixture

Run `test-limits.cjs` with the same `DOTNET`, `PLAYWRIGHT_MODULE`, and built DLL
argument as the browser fixture. It uses only a local fake provider and synthetic
reference. It checks origin/input/slot/question/context refusal, queue overload,
held-inference cancellation on disconnect/expiry/shutdown, slot recovery and fresh
visitor isolation. It also grows conversation with large unique answers and JSON-
escaped questions, verifies the complete reference and every prior answer remain in
admitted requests, checks refusal without an extra provider call, then checks fresh
session recovery. The current fixture makes 28 fake requests and requires clean host
exits (25 for lifecycle cases, three for growth/recovery).
These checks do not establish public deployment security or actual token/cost limits.

### Known reference-fidelity limitations — authorized early release

The existing Measure X reading copy covers 266 pages but is **not a verified transcription**. Bill authorized an early release with explicit warnings rather than waiting for full reconstruction.
Pixel comparison found lost amendment markup and merged numeric table cells even
on page 4, which has no OCR recovery warning; page 17 has similar table problems.
A future faithful, reviewed transcription must preserve additions, deletions and table
relationships. The sixteen reviewed drafts remain separate from the full OCR reference. Full-context retention cannot correct a damaged source. See the
[review notebook](../../../docs/edition-3-notes/document-chat.md#document-fidelity-review--public-release-blocker)
for evidence and limits. No corrected operative legal values are claimed.

## Launch packaging

The launch package includes the original public source PDF as `wwwroot/measure-x.pdf` for verification. Do not place credentials or the OCR source under wwwroot. Only the operator-authorized Codex path is used in the live launcher. The operator accepts subscription quota exhaustion; no separate global quota cap is implemented. The public Codex-endpoint use has not been independently certified as contractually equivalent to the public OAuth route.
