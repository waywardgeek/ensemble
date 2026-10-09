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

## Executable schedule refinement after coordinator review c9ba32c

Still a proposal: no discovery, credentials or real requests are released. Use one
scratch root R per provider P. Record the exact absolute R, executable paths,
workspace, arguments and sanitized configuration in the launch manifest. Its
layout is R/A (plain session shared sequentially by A/B/C), R/C-fresh (different
SessionID), R/C-import (initially absent), R/D (skill session), R/catalog-original,
R/catalog-relocated, R/checkpoints and R/offline-branches. Never mount two copies
of one SessionID in a single Ensemble. Close each owner before the next process.
Markers are literal CH10-A-P, CH10-C-P and CH10-D-P after substituting the recorded
provider label only. Scratch content is ASCII, each file at most1KiB; prompt plus
fixture manual material at most8KiB per row. Bind actual files, not just this layout.

Use max output4096 tokens (the actual CLI/GUI setting and the public consumer's
explicit setting), one120-second deadline per generation attempt and a10-minute
row deadline. Set per-turn policy limits as specified below, count each outgoing
attempt BEFORE transport, and keep a root-wide22-attempt budget per provider plus
row ceilings6/4/6/6. Failed, canceled or timed-out attempts spend their slot. No
transport retry, model retry, alternative endpoint probe or automatic prompt
resubmission. Stop a row on refusal, missing required observation, limit or deadline;
unused slots are not an invitation to repeat it. The recorder/proxy must be locally
proven before live use. Native engine request timeout may expire earlier. CLI/GUI
wire output bounds remain4096; a proxy must not silently rewrite the request body.

Discovery: a single request per provider, including any pagination (therefore do
not follow another page). Select only a returned permitted model. Gemini target is
Gemini3.8Flash as Bill requires; record unavailability, do not substitute an older
model or spend a probe. No remembered exact provider identifier is treated as
verified. Credential carriage will be memory/environment only after release.

A: launch actual PTY process1 with `CLI --session-dir R/A chat`. The CLI has no
policy selector: historical turn policy is raw0/effective16. The enforcing proxy
independently stops each process at3 outgoing attempts; this is a spending bound,
not a captured runtime policy value.
Prompt exactly: "Remember CH10-A-P. Use write_file once to create marker.txt with
exactly CH10-A-P followed by one newline. Then tell me why you made that edit."
Observe response and actual file, then type /session, /checkpoint, /history, /quit,
observing each command before the next. Preserve checkpoint export A1.json at its
acknowledged boundary plus original store files after close. Launch PTY process2
with the same selection and type: "What exact marker did you save in marker.txt,
and why did you write it? Answer from this conversation; do not call any tool or
change a file." Observe, /session, /quit. Require file unchanged and no new tool
call on recall. At most3 attempts per process,6 total. No expected output is inserted
into the transcript. A missing edit/recall is a failed observation, not replaced
with a local fixture.

B: after A exits, launch `GUI --session-dir R/A --terminal --port 0 --policy
R/current-policy.json --preferences R/preferences.json` in a PTY. Browser navigates
to the actual printed loopback origin. Record exact keyboard/pointer actions,
launch-bound WebSocket frame IDs/revisions, screenshot and accessible DOM text.
Click Checkpoint and require applied session_changed before saved ack. Disconnect
and reconnect that Page's socket; close a second Page separately; enter EOF in the
attached terminal separately. After each, verify server/session still usable.
Using actual Settings controls, change current max_model_requests from the recorded
A raw0/effective16 to raw1/effective1; change the actual Font size input from its initial16 to18
(`font_size`, supported range12–28), record the applied revision and acknowledgment.
Prompt: "Use write_file once to create policy-proof.txt containing CH10-B-P and a
newline, then explain the result." A real returned call plus paired effect/result
and round-limit completion with no continuation is required to demonstrate the
current cap; a plain answer alone is explicitly missing evidence. Preserve old
turn.policy raw0/effective16 and new turn.policy raw1/effective1, with their actual
revisions. No CLI policy flag or new runtime setting is introduced. Stop the server
via SIGTERM and restart it, reopen Settings and verify both the current policy and
changed preference independently persisted. Server restart is distinct from socket
reconnect, Page close and terminal detach. Historical cards must show no live owner.
No native hearing/speech claim. At most4 attempts overall, with cap1 on this turn;
unused slots stay unused. The driver must check the actual initial preference16
and record a failed precondition if the prepared launch differs.

C: compiled headless public consumer opens R/A and R/C-fresh in one Ensemble with
current per-turn cap2. Agent1 prompt: "Remember CH10-C-P in this conversation and
repeat it once. Do not call tools." Agent2 prompt: "Reply ONLY FRESH-P. Do not call
tools." Collect each completion independently; retain partial success if its peer
fails. Before these calls, export R/A to checkpoints/C-old.json. After both finish,
export checkpoints/C-new.json and retain the complete events.log containing the
newer tail. Close both. For differential controls, use three inert offline branches:
(1) full log with C-new; (2) same full log with C-old; (3) same full log with C-old's
state/hash explicitly nulled. Endpoints/credentials are disabled; public inspection
and Render use identical copied current inputs, and ReconstructRequest uses the
same genuine recorded send. Compare byte-identical request bodies, per-model usage,
Skills and normalized watch facts. No paid duplicate request is needed for equality.

Import C-new through ImportSession into the initially empty R/C-import after the
old owner is closed. Save SHA256 of exact origin.json. Check actual Events/Dump
contain anchor/tail only and a genuine pre-origin request sequence refuses with
history_unavailable. Make one actual prompt: "What marker beginning CH10-C is in
this conversation? Answer without tools." Save, close, reopen, compare unchanged
origin bytes/hash, once-restored usage and retained coordinates/window. Mount this
identity only once per root at every step. This row permits at most2+2+2 attempts,
6 total. Checkpoint export bytes and null-state modifications are local branch
artifacts, kept separately from untouched original records.

D: use a frozen primary plus narrow scratch capability, with explicit empty
bindings `{}` and exact catalog/handler hashes in the manifest. The actual CLI
configuration reader has no scalar-binding input, so the marker is supplied in
the literal prompt. Do not claim an unsupported CLI binding selection. Primary
has tool_limits available and offers scratch; scratch grants only write_file plus
required Chapter9 management tools. Public LoadSkill/UnloadSkill/LoadSkill records
retired and new activation identities; record actual returned IDs/revisions rather
than assuming contiguous numeric literals in the evidence. Prompt in the actual
PTY: "Use write_file once to create skill-proof.txt containing CH10-D-P and one
newline. Explain what capability allowed it." Observe file and response. The CLI
records raw0/effective16; proxy D1 spending cap3 is separate. Public D2 deliberately
sets runtime policy3 via UpdatePolicy and the proxy also enforces its cap3.

To leave pending limits without trusting model compliance, use the permitted public
local-HTTP fixture pattern already tested by TestSessionOneShotSettingSurvivesAndConsumesUnknownCall:
while idle, retain current Config, replace only BaseURL/APIKey with a loopback fixture
and dummy key via SetConfig, keep the selected wire API/model, and Ask exactly
"LOCAL FIXTURE: seed one-shot tool_limits; this is not a provider demonstration."
The local fixture emits one accepted tool_limits call with
`{"ai_callback_pattern":"","max_output_bytes":17}`, then a final text response,
using unique fixture call IDs. Actor/Tools/Jobs perform the actual set transition;
the harness does not append a limits fact or edit state. Record both fixture HTTP
requests, nominal selected vendor/model provenance, and explicitly mark their
source as loopback fixture, never actual provider output. Restore the prior Config
before saving. These two bounded local responses spend zero paid slots and are
excluded from claimed real-model usage/response evidence (their known fixture
usage is separately identified in the cumulative session ledger).

Copy only the small unchanged catalog to catalog-relocated, then close/reopen R/D
through public session opening with the relocated catalog. Inspect exact retained
manual bytes, retired/new identities, pending17-byte setting and empty-pattern
presence. Actual resumed prompt: "Use write_file once to create after-resume.txt
containing CH10-D-P and one newline, then confirm completion." Require a real next
attempted call, one consumed fact with exact copied overrides before called/result,
and the observed effect or refusal. If it produces no call, record missing evidence
and stop; do not manufacture a live consumption claim. Cap3 on each real turn,6
paid attempts total. If no opaque signature appears in actual output, keep opaque
restrictions covered deterministically and explicitly report no live instance.

## Support freeze prerequisite (pending; no claim it already exists)

Before credentials, produce a reviewed immutable manifest covering all core/GUI/
consumer Go source, every nested go.mod/go.sum and dependency selections, GUI assets,
CLI/GUI/compiled-public-consumer SHA256 and exact build association, format/API docs,
catalog/bindings, driver, enforcing proxy, recorder, verifier and sanitized launches.
The public consumer/driver must implement this schedule and budget as reviewed;
these prose edits do not claim a compiled consumer or an enforced live counter.
Original PTY/socket/provider bytes and timestamps remain separate from reconstructed
requests and comparisons. The proxy must omit authorization/credential data from
receipts and decline attempt N+1 before forwarding it; test cancellation/timeout
accounting and no automatic retries with a local server.

From one complete valid local parent manifest/run, independently change source,
CLI binary, GUI binary, consumer binary, support script, dependency/module mapping,
catalog/binding, and sanitized launch identity. Each intended mismatch must refuse
before any derived write; retain target-directory before/after hashes and controls
for missing/incomplete map. A bad path rejected before the intended hash check is
not that negative control. Only then freeze support for coordinator review. Support
build/testing follows the serialized disk schedule; no competing broad build is
started while independent graders hold it. This remains a separate prerequisite
and does not expand the66/3 ceiling or authorize live access.


## Support preparation corrections and executable seam

Coordinator's latest two corrections are incorporated above: CLI history is
raw0/effective16, and no restored speech requires measured admission/owner state.
The earlier conceptual row table's “policy” refers to actual captured values;
request spending ceilings are independently enforced by support/schedule.json.
The support CLI configuration inspection also required empty D bindings, as
explained above. No runtime feature is added for the demonstration.

Browser driver support/browser.mjs installs native speech `speak` counting before
Page startup, wraps the actual application-owned SpeechService.submit to count
logical admissions, and reads each actual Page's queue/current work and pause
causes plus service pending/active ownership. It exposes the real application via
a recorded app.js instrumentation suffix; runtime source is unchanged. At initial
resume, after connection generation replacement, after Page.close, and after
replacement Page creation, require zero new native/logical admission, no queued or
current work, and no speaking pause. Socket frames retain connection identity;
applied session_changed and saved acknowledgment are recorded at the actual Page
methods. These are explicit local instrumentation observations, not inference from
silence, screenshot, or a claim of hearing. Ordinary tab close and server restart
remain separately recorded operations. No extra model turn is needed.

The support tree contains the bounded schedule, source/build identity preflight,
proxy/attempt journal, deliberate PTY/browser recorders, public C/D consumer,
original-receipt sealer/verifier and small Python/JS controls. New Go consumer is
formatted but uncompiled pending the coordinator's compiler slot. Its source is
support only. No runtime files have changed and no actual binary binding exists.
The binding builder must execute real commands and collect module/build identity
evidence before creating one. Local synthetic identity controls are explicitly
fixtures, not evidence of a Go build or actual browser/provider behavior.

The original loopback-only support has now been exercised locally; source-specific
CLI/browser/public results and failures are preserved in retained-repair-handback.md.
Runtime/build57d4aac remains the compiled authority. The next preparation adds an
explicit, separately bound provider adapter and discovery entrypoint; these are
unreleased operations, not actual provider results. Default fixture mode remains
literal loopback with no key access. See support/provider-support.md for the exact
settings fields, fixed origins, single-page discovery, selected-model constraints,
sanitization and unchanged66-generation/3-discovery ceiling. No source/binary is
relabeled, and a split support binding retains original compiled-input identities.

The full A/B/C/D feature matrix, prompts, interface requirements, independent
local controls, D seed provenance and policy/spending distinction above remain
unchanged. Ready support does not imply usable credentials, discovered models,
deterministic clearance or permission to execute live. Anthropic/OpenAI identifiers
must be selected from their actual one-page discovery; Gemini must actually return
models/gemini-3.8-flash with generateContent support. No page/fallback/probe or model
substitution is added. Coordinator support review and separate live release are
required before any settings read or API call. Preparation uses Python/JS local
controls only while the grader owns the Go compiler.
