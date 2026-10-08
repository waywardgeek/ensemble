# Chapter 10 proposed bounded live matrix — NOT RELEASED OR EXECUTED

This is a proposal for coordinator completeness review after local completeness.
No credential reads, provider discovery or provider requests occurred in the local
phase. Bind the eventual approved source commit, both executable SHA256 values,
public consumer source, catalog/bindings identity, provider/model discovery receipt
and the selected state_version1 format before any real run. Current model IDs are
not inferred from old receipts. Stop on any unexpected extra requests or refusal;
retain it and request a revised bounded matrix rather than retry automatically.

Apply rows A–D separately to Messages/Anthropic, Chat Completions/OpenAI and
GenerateContent/Gemini, using separately selected scratch stores and distinct
markers. Real human-mode input must be driven interactively in actual PTYs, with
responses observed before sending follow-ups. Browser actions require actual Page
and socket behavior, not screenshots of injected state. Public consumers use only
exported library APIs. All files edited by a model are disposable scratch files.

| Row | Actual action and feature | Required observation/evidence | Request ceiling per provider |
|---|---|---|---|
| A | PTY1: open fresh explicit session, ask to remember marker and make one narrowly scoped scratch edit; inspect response/file; /session, /checkpoint, /history, /quit. PTY2: reopen same store; ask about retained marker and why file changed; /session and /quit. | Stable SessionID, honest fresh/resumed state, saved anchor, retained calls/raw usage/provenance, unchanged file after recall, no tool reexecution/startup HTTP; two actual process transcripts, wire requests/responses, usage and store hashes. | 6 generation requests total across both processes; local commands use zero. |
| B | Actual browser starts on closed CLI store; view retained cards/session/omission count, checkpoint, reconnect. Select today's independent policy and make one real turn demonstrating it. Open second browser page, then close it; detach combined --terminal via EOF. | Applied session_changed before matching saved ack, same anchor on reconnect; historical job is labeled no live owner and cannot supervise; no restored provisional text/speech/pause registration; policy/preferences persist independently. Closing page/terminal leaves server/session usable. | 4 generation requests; checkpoint/reconnect/detach use zero. |
| C | Public headless consumer: open two independent Agents, one resumed and one fresh, in separate directories; make one turn each. Export settled real-model state, close its owner, import to empty destination, inspect origin boundary, then make one real continuation. Keep an older checkpoint while a later accepted turn is in its tail and reopen under the same current render inputs. | No cross-Agent state/usage/limits; source+binary-bound consumer transcript; exact real next request on each API, usage restored once, owned buffers, origin bytes immutable after save/reopen, Events/Dump only actual anchor/tail, history_unavailable before origin; true nonempty-tail path equals complete rebuild. | 6 generation requests. |
| D | Skill-mode public/PTY consumer with fixed narrow catalog: initialize, load visible capability, retire/reload it through public controls, record a real turn. Leave pending typed limits using a controlled public accepted-call fixture, save/close, move unchanged catalog, resume, make one real next attempted call. | Exact active/retired manuals, grants, transition coordinates and placement survive; semantic catalog relocation accepted; next attempt consumes once before dispatch/refusal; recorded signature/opaque projection restrictions preserved. No invented new grants. Preserve unexpected model behavior; deterministic tests prove the exact refusal cases separately. | 6 generation requests. |

Total generation ceiling:22 per provider,66 overall; no automatic retries. A
ceiling includes every model exchange in a tool loop, including error/refusal
responses. The harness must enforce the ceiling before issuing the next request.
Proposed discovery ceiling:one list/discovery request per provider, only after
separate release; discovery and generation are distinct metered actions. Root
may reduce/split this matrix before authorization. If a provider requires a
separate capability probe, it needs an explicitly reviewed addition first.

Local-only controls accompany each relevant store, using copied disposable
fixtures while preserving original hashes: selector conflicts/fresh legacy mode,
System presence/identity mismatch, changed/missing catalog/binding/handler,
writer exclusion and duplicate SessionID, unfinished/corrupt/partial/missing
origin refusals, exact/one-over bounds, queued/HTTP/report-worker busy controls,
slow checkpoint with accepted job tail, concurrent saves, checked write/sync/
close/replace failures, failed append after consumed limits, repeated close,
CLI EOF/quit/SIGINT/SIGTERM and SIGKILL lock release, child descriptor exclusion,
historical stale-handle refusal, all limit attempt variants, forged public
adoption and returned-buffer mutations. These request zero real-provider calls.
They are deterministic local controls and do not themselves constitute the spin.

Retain actual original PTY/browser/public transcripts separately from reconstructed
requests; bind any derived verifier to source and executable identities before
it writes. Record unexpected failures and missing observables, never substitute a
fake response for a missing live result. Freeze the initial student experience
before independent historical comparison. No runtime acceptance, live acceptance,
author prose acceptance or immutable chapter release is claimed by this plan.
